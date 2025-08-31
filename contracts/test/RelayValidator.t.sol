// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import {Test, console} from "forge-std/Test.sol";
import "../src/RelayValidator.sol";

contract RelayValidatorTest is Test {
    RelayValidator public relayValidator;
    address public validator1 = address(0x1);
    address public validator2 = address(0x2);
    address public validator3 = address(0x3);
    
    function setUp() public {
        relayValidator = new RelayValidator();
        // Align validator addresses with signing keys used by vm.sign
        validator1 = vm.addr(1);
        validator2 = vm.addr(2);
        validator3 = vm.addr(3);
    }

    function testRegisterValidator() public {
        vm.deal(validator1, 20 ether);
        vm.startPrank(validator1);
        
        uint256[4] memory blsKey = [uint256(1), uint256(2), uint256(3), uint256(4)];
        relayValidator.registerValidator{value: 15 ether}(blsKey);
        
        address[] memory activeValidators = relayValidator.getActiveValidators();
        assertEq(activeValidators.length, 1);
        assertEq(activeValidators[0], validator1);
        
        RelayValidator.Validator memory validatorInfo = relayValidator.getValidatorInfo(validator1);
        assertEq(validatorInfo.stake, 15 ether);
        assertEq(uint(validatorInfo.status), uint(RelayValidator.ValidatorStatus.Active));
        
        vm.stopPrank();
    }

    function testInsufficientStake() public {
        vm.deal(validator1, 5 ether);
        vm.startPrank(validator1);
        
        vm.expectRevert(RelayValidator.InsufficientStake.selector);
        uint256[4] memory blsKey = [uint256(1), uint256(2), uint256(3), uint256(4)];
        relayValidator.registerValidator{value: 5 ether}(blsKey);
        
        vm.stopPrank();
    }

    function testDuplicateRegistration() public {
        vm.deal(validator1, 30 ether);
        vm.startPrank(validator1);
        
        uint256[4] memory blsKey = [uint256(1), uint256(2), uint256(3), uint256(4)];
        relayValidator.registerValidator{value: 15 ether}(blsKey);
        
        vm.expectRevert(RelayValidator.ValidatorAlreadyRegistered.selector);
        uint256[4] memory duplicateKey = [uint256(1), uint256(2), uint256(3), uint256(4)];
        relayValidator.registerValidator{value: 10 ether}(duplicateKey);
        
        vm.stopPrank();
    }

    function testRequestValidation() public {
        _setupValidators();
        
        bytes32 messageHash = keccak256("test payment");
        uint256 requestId = relayValidator.requestValidation(1, messageHash, 1000 ether);
        
        (
            uint256 id,
            uint256 paymentId,
            bytes32 hash,
            uint256 requiredSigs,
            uint256 receivedSigs,
            RelayValidator.ValidationStatus status,
            , // createdAt (unused)
            , // deadline (unused)
            bool isHighValue
        ) = relayValidator.getValidationRequest(requestId);
        
        assertEq(id, requestId);
        assertEq(paymentId, 1);
        assertEq(hash, messageHash);
        assertTrue(requiredSigs > 0);
        assertEq(receivedSigs, 0);
        assertEq(uint(status), uint(RelayValidator.ValidationStatus.Pending));
        assertTrue(isHighValue);
    }

    function testSignValidation() public {
        _setupValidators();
        
        bytes32 messageHash = keccak256("test payment");
        uint256 requestId = relayValidator.requestValidation(1, messageHash, 100 ether);
        
        vm.startPrank(validator1);
        
        bytes32 ethSignedHash = keccak256(abi.encodePacked("\x19Ethereum Signed Message:\n32", messageHash));
        (uint8 v, bytes32 r, bytes32 s) = vm.sign(1, ethSignedHash);
        bytes memory signature = abi.encodePacked(r, s, v);
        
        relayValidator.signValidation(requestId, signature);
        
        vm.stopPrank();
    }

    function testSlashValidator() public {
        _setupValidators();
        
        RelayValidator.Validator memory validatorInfo = relayValidator.getValidatorInfo(validator1);
        uint256 initialStake = validatorInfo.stake;
        
        relayValidator.slashValidator(validator1, "Failed to respond to validation request");
        
        validatorInfo = relayValidator.getValidatorInfo(validator1);
        assertEq(uint(validatorInfo.status), uint(RelayValidator.ValidatorStatus.Slashed));
        assertLt(validatorInfo.stake, initialStake);
        
        address[] memory activeValidators = relayValidator.getActiveValidators();
        assertEq(activeValidators.length, 2); // Should only have 2 remaining
    }

    function testValidatorExit() public {
        _setupValidators();
        
        uint256 initialBalance = validator1.balance;
        
        vm.startPrank(validator1);
        relayValidator.exitValidator();
        vm.stopPrank();
        
        RelayValidator.Validator memory validatorInfo = relayValidator.getValidatorInfo(validator1);
        assertEq(uint(validatorInfo.status), uint(RelayValidator.ValidatorStatus.Exiting));
        assertEq(validatorInfo.stake, 0);
        
        assertGt(validator1.balance, initialBalance);
        
        address[] memory activeValidators = relayValidator.getActiveValidators();
        assertEq(activeValidators.length, 2);
    }

    function testExpireValidation() public {
        _setupValidators();
        
        bytes32 messageHash = keccak256("test payment");
        uint256 requestId = relayValidator.requestValidation(1, messageHash, 100 ether);
        
        vm.warp(block.timestamp + 6 minutes);
        
        relayValidator.expireValidation(requestId);
        
        (,,,,,RelayValidator.ValidationStatus status,,,) = relayValidator.getValidationRequest(requestId);
        assertEq(uint(status), uint(RelayValidator.ValidationStatus.Expired));
    }

    function testInvalidValidationRequest() public {
        vm.expectRevert(RelayValidator.InvalidValidationRequest.selector);
        relayValidator.getValidationRequest(999);
    }

    function testSetHighValueThreshold() public {
        uint256 newThreshold = 5000 ether;
        relayValidator.setHighValueThreshold(newThreshold);
        assertEq(relayValidator.highValueThreshold(), newThreshold);
    }

    function testPauseUnpause() public {
        relayValidator.pause();
        assertTrue(relayValidator.paused());
        
        relayValidator.unpause();
        assertFalse(relayValidator.paused());
    }

    function testUnauthorizedActions() public {
        vm.startPrank(validator1);
        
        vm.expectRevert();
        relayValidator.slashValidator(validator2, "Unauthorized");
        
        vm.expectRevert();
        relayValidator.setHighValueThreshold(1000 ether);
        
        vm.expectRevert();
        relayValidator.pause();
        
        vm.stopPrank();
    }

    function testInsufficientValidators() public {
        bytes32 messageHash = keccak256("test payment");
        
        vm.expectRevert(RelayValidator.InsufficientSignatures.selector);
        relayValidator.requestValidation(1, messageHash, 100 ether);
    }

    function testMessageAlreadyProcessed() public {
        _setupValidators();
        
        bytes32 messageHash = keccak256("test payment");
        relayValidator.requestValidation(1, messageHash, 100 ether);
        
        vm.expectRevert(RelayValidator.MessageAlreadyProcessed.selector);
        relayValidator.requestValidation(2, messageHash, 100 ether);
    }

    function _setupValidators() internal {
        vm.deal(validator1, 20 ether);
        vm.deal(validator2, 20 ether);
        vm.deal(validator3, 20 ether);
        
        vm.prank(validator1);
        uint256[4] memory blsKey1 = [uint256(1), uint256(2), uint256(3), uint256(4)];
        relayValidator.registerValidator{value: 15 ether}(blsKey1);
        
        vm.prank(validator2);
        uint256[4] memory blsKey2 = [uint256(5), uint256(6), uint256(7), uint256(8)];
        relayValidator.registerValidator{value: 15 ether}(blsKey2);
        
        vm.prank(validator3);
        uint256[4] memory blsKey3 = [uint256(9), uint256(10), uint256(11), uint256(12)];
        relayValidator.registerValidator{value: 15 ether}(blsKey3);
        
        address[] memory activeValidators = relayValidator.getActiveValidators();
        assertEq(activeValidators.length, 3);
    }
}
