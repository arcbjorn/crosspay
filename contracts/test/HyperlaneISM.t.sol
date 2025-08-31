// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import {Test, console} from "forge-std/Test.sol";
import "../src/adapters/HyperlaneISM.sol";
import "../src/PaymentCore.sol";

contract MockISM {
    function moduleType() external pure returns (uint8) {
        return 3; // LEGACY_MULTISIG
    }
    
    function verify(bytes calldata, bytes calldata) external pure returns (bool) {
        return true;
    }
    
    function validatorsAndThreshold(
        bytes calldata
    ) external pure returns (address[] memory validators, uint8 threshold) {
        validators = new address[](3);
        validators[0] = address(0x100);
        validators[1] = address(0x200);
        validators[2] = address(0x300);
        threshold = 2;
        return (validators, threshold);
    }
}

contract MockHyperlaneMailbox {
    bytes32 private _messageId = keccak256("mock_message");
    mapping(bytes32 => bool) public delivered;
    
    function dispatch(
        uint32,
        bytes32,
        bytes calldata
    ) external payable returns (bytes32 messageId) {
        messageId = keccak256(abi.encodePacked(_messageId, block.timestamp));
        _messageId = messageId;
        return messageId;
    }
    
    function process(bytes calldata, bytes calldata) external {
        // Mock processing
    }
    
    function latestCheckpoint() external pure returns (bytes32 root, uint32 index) {
        return (keccak256("mock_root"), 100);
    }
}

contract HyperlaneISMTest is Test {
    HyperlaneISM public hyperlaneISM;
    PaymentCore public paymentCore;
    MockISM public mockISM;
    MockHyperlaneMailbox public mockMailbox;
    
    address public alice = address(0x1);
    address public bob = address(0x2);
    
    function setUp() public {
        mockISM = new MockISM();
        mockMailbox = new MockHyperlaneMailbox();
        paymentCore = new PaymentCore();
        
        hyperlaneISM = new HyperlaneISM(
            address(mockISM),
            address(mockMailbox),
            address(paymentCore)
        );
        paymentCore.setTrustedAdapter(address(hyperlaneISM), true);
        
        vm.deal(alice, 10 ether);
        vm.deal(bob, 10 ether);
    }
    
    function testDispatchPayment() public {
        vm.startPrank(alice);
        
        // Create a payment first
        uint256 paymentId = paymentCore.createPayment{value: 1.01 ether}(
            bob,
            address(0), // ETH
            1 ether,
            "test-metadata",
            "alice.eth",
            "bob.eth"
        );
        
        // Dispatch payment via Hyperlane
        bytes32 messageId = hyperlaneISM.dispatchPayment{value: 0.02 ether}(
            paymentId,
            8453, // BASE_DOMAIN
            bob
        );
        
        assertTrue(messageId != bytes32(0));
        
        HyperlaneISM.HyperlanePayment memory hlPayment = hyperlaneISM.getHyperlanePayment(paymentId);
        assertEq(hlPayment.localPaymentId, paymentId);
        assertEq(hlPayment.destinationDomain, 8453);
        assertFalse(hlPayment.delivered);
        
        vm.stopPrank();
    }
    
    function testVerifyISMProof() public {
        vm.startPrank(alice);
        
        uint256 paymentId = paymentCore.createPayment{value: 1.01 ether}(
            bob,
            address(0),
            1 ether,
            "test-metadata",
            "alice.eth",
            "bob.eth"
        );
        
        hyperlaneISM.dispatchPayment{value: 0.02 ether}(
            paymentId,
            8453,
            bob
        );
        
        vm.stopPrank();
        
        // Verify ISM proof
        bytes memory metadata = abi.encode("mock_metadata");
        HyperlaneISM.ValidationProof memory proof = hyperlaneISM.verifyISMProof(paymentId, metadata);
        
        assertTrue(proof.verified);
        assertEq(proof.validators.length, 3);
        assertEq(proof.threshold, 2);
        assertEq(proof.validators[0], address(0x100));
        assertEq(proof.validators[1], address(0x200));
        assertEq(proof.validators[2], address(0x300));
    }
    
    function testHandle() public {
        uint32 origin = 1; // Ethereum domain
        bytes32 sender = bytes32(uint256(uint160(address(hyperlaneISM))));
        
        bytes memory messageBody = abi.encode(
            456, // sourcePaymentId
            alice, // sender
            bob, // recipient
            address(0), // token (ETH)
            2 ether, // amount
            "hyperlane-metadata", // metadataURI
            block.timestamp // timestamp
        );
        
        // Simulate receiving message via Hyperlane
        vm.deal(address(paymentCore), 10 ether);
        vm.prank(address(mockMailbox));
        hyperlaneISM.handle(origin, sender, messageBody);
        
        // Verify local payment was created
        uint256 expectedPaymentId = paymentCore.getPaymentCount();
        assertTrue(expectedPaymentId > 0);
    }
    
    function testProcessWithISM() public {
        bytes memory metadata = abi.encode("verification_metadata");
        bytes memory message = abi.encode(
            789,
            alice,
            bob,
            address(0),
            1 ether,
            "ism-processed",
            block.timestamp
        );
        
        // Process message with ISM verification
        hyperlaneISM.processWithISM(metadata, message);
        
        bytes32 messageId = keccak256(message);
        // Would check if message was processed in a real implementation
    }
    
    function testUnsupportedDomain() public {
        vm.startPrank(alice);
        
        uint256 paymentId = paymentCore.createPayment{value: 1.01 ether}(
            bob,
            address(0),
            1 ether,
            "test-metadata",
            "alice.eth",
            "bob.eth"
        );
        
        vm.expectRevert(abi.encodeWithSelector(HyperlaneISM.UnsupportedDomain.selector, 99999));
        hyperlaneISM.dispatchPayment{value: 0.02 ether}(
            paymentId,
            99999, // Unsupported domain
            bob
        );
        
        vm.stopPrank();
    }
    
    function testGetValidationProof() public {
        vm.startPrank(alice);
        
        uint256 paymentId = paymentCore.createPayment{value: 1.01 ether}(
            bob,
            address(0),
            1 ether,
            "test-metadata",
            "alice.eth",
            "bob.eth"
        );
        
        bytes32 messageId = hyperlaneISM.dispatchPayment{value: 0.02 ether}(
            paymentId,
            8453,
            bob
        );
        
        vm.stopPrank();
        
        (address[] memory validators, uint8 threshold) = hyperlaneISM.getValidationProof(messageId);
        
        assertEq(validators.length, 3);
        assertEq(threshold, 2);
    }
    
    function testAddRemoveSupportedDomain() public {
        uint32 newDomain = 12345;
        
        assertFalse(hyperlaneISM.supportedDomains(newDomain));
        
        hyperlaneISM.addSupportedDomain(newDomain);
        assertTrue(hyperlaneISM.supportedDomains(newDomain));
        
        hyperlaneISM.removeSupportedDomain(newDomain);
        assertFalse(hyperlaneISM.supportedDomains(newDomain));
    }
    
    function testUnauthorizedActions() public {
        vm.startPrank(alice);
        
        vm.expectRevert();
        hyperlaneISM.addSupportedDomain(12345);
        
        vm.expectRevert();
        hyperlaneISM.setISM(address(0x123));
        
        vm.expectRevert();
        hyperlaneISM.pause();
        
        vm.stopPrank();
    }
    
    function testPauseUnpause() public {
        hyperlaneISM.pause();
        assertTrue(hyperlaneISM.paused());
        
        vm.startPrank(alice);
        
        uint256 paymentId = paymentCore.createPayment{value: 1.01 ether}(
            bob,
            address(0),
            1 ether,
            "test-metadata",
            "alice.eth",
            "bob.eth"
        );
        
        vm.expectRevert("Pausable: paused");
        hyperlaneISM.dispatchPayment{value: 0.02 ether}(
            paymentId,
            8453,
            bob
        );
        
        vm.stopPrank();
        
        hyperlaneISM.unpause();
        assertFalse(hyperlaneISM.paused());
    }
    
    function testMessageAlreadyProcessed() public {
        bytes32 messageId = keccak256("duplicate_message");
        
        uint32 origin = 1;
        bytes32 sender = bytes32(uint256(uint160(address(hyperlaneISM))));
        
        bytes memory messageBody = abi.encode(
            123,
            alice,
            bob,
            address(0),
            1 ether,
            "duplicate-test",
            block.timestamp
        );
        
        // Process message first time
        vm.deal(address(paymentCore), 10 ether);
        vm.prank(address(mockMailbox));
        hyperlaneISM.handle(origin, sender, messageBody);
        
        // Try to process same message again
        vm.prank(address(mockMailbox));
        vm.expectRevert(abi.encodeWithSelector(HyperlaneISM.MessageAlreadyProcessed.selector, keccak256(abi.encodePacked(origin, sender, messageBody))));
        hyperlaneISM.handle(origin, sender, messageBody);
    }
}
