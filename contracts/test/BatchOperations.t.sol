// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import {Test, console} from "forge-std/Test.sol";
import "../src/BatchOperations.sol";
import "../src/ConfidentialPayments.sol";
import "../src/RelayValidator.sol";
import "../src/TrancheVault.sol";
import "@openzeppelin/contracts/token/ERC20/ERC20.sol";

contract MockToken is ERC20 {
    constructor() ERC20("Mock Token", "MOCK") {
        _mint(msg.sender, 1000000 * 10**18);
    }
    
    function mint(address to, uint256 amount) external {
        _mint(to, amount);
    }
}

contract BatchOperationsTest is Test {
    BatchOperations public batchOps;
    ConfidentialPayments public confidentialPayments;
    RelayValidator public relayValidator;
    TrancheVault public trancheVault;
    MockToken public token;
    
    address public user = address(0x1);
    address public recipient1 = address(0x2);
    address public recipient2 = address(0x3);
    
    function setUp() public {
        token = new MockToken();
        
        confidentialPayments = new ConfidentialPayments();
        relayValidator = new RelayValidator();
        trancheVault = new TrancheVault(
            address(token),
            "Test Vault",
            "TV"
        );
        
        batchOps = new BatchOperations(
            address(confidentialPayments),
            address(relayValidator),
            address(trancheVault)
        );
        
        // Fund user
        token.mint(user, 1000 * 10**18);
        vm.deal(user, 10 ether);
        
        vm.label(user, "User");
        vm.label(recipient1, "Recipient1");
        vm.label(recipient2, "Recipient2");
    }
    
    function testBatchCreatePayments() public {
        BatchOperations.BatchPayment[] memory payments = new BatchOperations.BatchPayment[](2);
        
        payments[0] = BatchOperations.BatchPayment({
            recipient: recipient1,
            token: address(0), // ETH payment
            encryptedAmount: abi.encode(1 ether),
            metadataURI: "ipfs://test1",
            makePrivate: false
        });
        
        payments[1] = BatchOperations.BatchPayment({
            recipient: recipient2,
            token: address(0), // ETH payment
            encryptedAmount: abi.encode(2 ether),
            metadataURI: "ipfs://test2",
            makePrivate: false
        });
        
        vm.prank(user);
        uint256[] memory paymentIds = batchOps.batchCreatePayments{value: 1 ether}(payments);
        
        assertEq(paymentIds.length, 2);
        assertTrue(paymentIds[0] > 0);
        assertTrue(paymentIds[1] > 0);
    }
    
    function testBatchCreatePaymentsTooLarge() public {
        BatchOperations.BatchPayment[] memory payments = new BatchOperations.BatchPayment[](51);
        
        vm.expectRevert(BatchOperations.BatchTooLarge.selector);
        vm.prank(user);
        batchOps.batchCreatePayments(payments);
    }
    
    function testBatchCreatePaymentsEmpty() public {
        BatchOperations.BatchPayment[] memory payments = new BatchOperations.BatchPayment[](0);
        
        vm.expectRevert(BatchOperations.BatchTooLarge.selector);
        vm.prank(user);
        batchOps.batchCreatePayments(payments);
    }
    
    function testBatchDeposit() public {
        BatchOperations.BatchDeposit[] memory deposits = new BatchOperations.BatchDeposit[](2);
        
        deposits[0] = BatchOperations.BatchDeposit({
            tranche: TrancheVault.TrancheType.Junior,
            amount: 100 * 10**18
        });
        
        deposits[1] = BatchOperations.BatchDeposit({
            tranche: TrancheVault.TrancheType.Mezzanine,
            amount: 200 * 10**18
        });
        
        vm.startPrank(user);
        token.approve(address(trancheVault), 300 * 10**18);
        batchOps.batchDeposit(deposits);
        vm.stopPrank();
        
        // Verify deposits were made
        assertEq(trancheVault.getUserDeposit(user, TrancheVault.TrancheType.Junior), 100 * 10**18);
        assertEq(trancheVault.getUserDeposit(user, TrancheVault.TrancheType.Mezzanine), 200 * 10**18);
    }
    
    function testBatchDepositTooLarge() public {
        BatchOperations.BatchDeposit[] memory deposits = new BatchOperations.BatchDeposit[](11);
        
        vm.expectRevert(BatchOperations.BatchTooLarge.selector);
        vm.prank(user);
        batchOps.batchDeposit(deposits);
    }
    
    function testBatchGrantDisclosurePermissions() public {
        // First create some payments
        uint256[] memory paymentIds = new uint256[](2);
        paymentIds[0] = 1;
        paymentIds[1] = 2;
        
        address[] memory viewers = new address[](2);
        viewers[0] = recipient1;
        viewers[1] = recipient2;
        
        // This would normally require the caller to be the payment owner
        // For testing, we're just checking the function doesn't revert on proper input
        vm.expectRevert(); // Will revert because payments don't exist
        vm.prank(user);
        batchOps.batchGrantDisclosurePermissions(paymentIds, viewers);
    }
    
    function testBatchGrantDisclosurePermissionsSizeMismatch() public {
        uint256[] memory paymentIds = new uint256[](2);
        address[] memory viewers = new address[](1);
        
        vm.expectRevert(BatchOperations.BatchSizeMismatch.selector);
        vm.prank(user);
        batchOps.batchGrantDisclosurePermissions(paymentIds, viewers);
    }
    
    function testBatchGrantDisclosurePermissionsTooLarge() public {
        uint256[] memory paymentIds = new uint256[](101);
        address[] memory viewers = new address[](101);
        
        vm.expectRevert(BatchOperations.BatchTooLarge.selector);
        vm.prank(user);
        batchOps.batchGrantDisclosurePermissions(paymentIds, viewers);
    }
    
    function testBatchRevokeDisclosurePermissions() public {
        uint256[] memory paymentIds = new uint256[](2);
        paymentIds[0] = 1;
        paymentIds[1] = 2;
        
        address[] memory viewers = new address[](2);
        viewers[0] = recipient1;
        viewers[1] = recipient2;
        
        vm.expectRevert(); // Will revert because payments don't exist
        vm.prank(user);
        batchOps.batchRevokeDisclosurePermissions(paymentIds, viewers);
    }
    
    function testBatchCompletePayments() public {
        uint256[] memory paymentIds = new uint256[](2);
        paymentIds[0] = 1;
        paymentIds[1] = 2;
        
        vm.expectRevert(); // Will revert because payments don't exist
        vm.prank(user);
        batchOps.batchCompletePayments(paymentIds);
    }
    
    function testBatchCompletePaymentsTooLarge() public {
        uint256[] memory paymentIds = new uint256[](51);
        
        vm.expectRevert(BatchOperations.BatchTooLarge.selector);
        vm.prank(user);
        batchOps.batchCompletePayments(paymentIds);
    }
    
    function testEstimateBatchGas() public view {
        BatchOperations.BatchPayment[] memory payments = new BatchOperations.BatchPayment[](3);
        
        uint256 gasEstimate = batchOps.estimateBatchGas(payments);
        
        // Should be: 21000 (base) + (3 * 185000) (per payment) + 5000 (batch overhead) = 581000
        assertEq(gasEstimate, 581000);
    }
    
    function testBatchRequestValidationsOnlyOwner() public {
        BatchOperations.BatchValidation[] memory validations = new BatchOperations.BatchValidation[](1);
        
        validations[0] = BatchOperations.BatchValidation({
            paymentId: 1,
            messageHash: keccak256("test"),
            amount: 100 * 10**18
        });
        
        vm.expectRevert("Only validator owner");
        vm.prank(user);
        batchOps.batchRequestValidations(validations);
    }
}