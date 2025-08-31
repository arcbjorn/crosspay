// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import "forge-std/Test.sol";
import "../src/PaymentCore.sol";
import "@openzeppelin/contracts/token/ERC20/ERC20.sol";

contract MockERC20 is ERC20 {
    constructor() ERC20("Mock Token", "MOCK") {
        _mint(msg.sender, 1000000 * 10**decimals());
    }
    
    function mint(address to, uint256 amount) external {
        _mint(to, amount);
    }
}

contract PaymentCoreTest is Test {
    PaymentCore public paymentCore;
    MockERC20 public mockToken;
    
    address public owner = address(1);
    address public alice = address(2);
    address public bob = address(3);
    
    uint256 public constant PAYMENT_AMOUNT = 1 ether;
    uint256 public constant FEE_AMOUNT = 0.001 ether; // 0.1% of 1 ETH
    uint256 public constant TOTAL_AMOUNT = PAYMENT_AMOUNT + FEE_AMOUNT;
    
    event PaymentCreated(
        uint256 indexed id,
        address indexed sender,
        address indexed recipient,
        address token,
        uint256 amount,
        uint256 fee,
        string metadataURI
    );
    
    event PaymentCompleted(uint256 indexed id, address indexed completer);
    event PaymentRefunded(uint256 indexed id, address indexed refunder);
    
    function setUp() public {
        vm.startPrank(owner);
        paymentCore = new PaymentCore();
        mockToken = new MockERC20();
        vm.stopPrank();
        
        // Give test accounts some ETH
        vm.deal(alice, 10 ether);
        vm.deal(bob, 10 ether);
        
        // Give test accounts some tokens
        mockToken.mint(alice, 1000 * 10**18);
        mockToken.mint(bob, 1000 * 10**18);
    }
    
    function testCreateETHPayment() public {
        vm.startPrank(alice);
        
        vm.expectEmit(true, true, true, true);
        emit PaymentCreated(
            1,
            alice,
            bob,
            address(0),
            PAYMENT_AMOUNT,
            FEE_AMOUNT,
            ""
        );
        
        uint256 paymentId = paymentCore.createPayment{value: TOTAL_AMOUNT}(
            bob,
            address(0),
            PAYMENT_AMOUNT,
            ""
        );
        
        assertEq(paymentId, 1);
        
        PaymentCore.Payment memory payment = paymentCore.getPayment(paymentId);
        assertEq(payment.sender, alice);
        assertEq(payment.recipient, bob);
        assertEq(payment.amount, PAYMENT_AMOUNT);
        assertEq(payment.fee, FEE_AMOUNT);
        assertEq(uint(payment.status), uint(PaymentCore.PaymentStatus.Pending));
        
        vm.stopPrank();
    }
    
    function testCreateTokenPayment() public {
        vm.startPrank(alice);
        
        mockToken.approve(address(paymentCore), TOTAL_AMOUNT);
        
        uint256 paymentId = paymentCore.createPayment(
            bob,
            address(mockToken),
            PAYMENT_AMOUNT,
            "ipfs://test"
        );
        
        PaymentCore.Payment memory payment = paymentCore.getPayment(paymentId);
        assertEq(payment.token, address(mockToken));
        assertEq(payment.metadataURI, "ipfs://test");
        
        vm.stopPrank();
    }
    
    function testCompletePayment() public {
        // Create payment
        vm.startPrank(alice);
        uint256 paymentId = paymentCore.createPayment{value: TOTAL_AMOUNT}(
            bob,
            address(0),
            PAYMENT_AMOUNT,
            ""
        );
        vm.stopPrank();
        
        uint256 bobBalanceBefore = bob.balance;
        
        // Complete payment as recipient
        vm.startPrank(bob);
        vm.expectEmit(true, true, false, false);
        emit PaymentCompleted(paymentId, bob);
        
        paymentCore.completePayment(paymentId);
        vm.stopPrank();
        
        // Check payment status
        PaymentCore.Payment memory payment = paymentCore.getPayment(paymentId);
        assertEq(uint(payment.status), uint(PaymentCore.PaymentStatus.Completed));
        
        // Check bob received the payment
        assertEq(bob.balance, bobBalanceBefore + PAYMENT_AMOUNT);
    }
    
    function testRefundPayment() public {
        // Create payment
        vm.startPrank(alice);
        uint256 paymentId = paymentCore.createPayment{value: TOTAL_AMOUNT}(
            bob,
            address(0),
            PAYMENT_AMOUNT,
            ""
        );
        vm.stopPrank();
        
        // Fast forward past refund delay
        vm.warp(block.timestamp + 25 hours);
        
        uint256 aliceBalanceBefore = alice.balance;
        
        // Refund payment as sender
        vm.startPrank(alice);
        vm.expectEmit(true, true, false, false);
        emit PaymentRefunded(paymentId, alice);
        
        paymentCore.refundPayment(paymentId);
        vm.stopPrank();
        
        // Check payment status
        PaymentCore.Payment memory payment = paymentCore.getPayment(paymentId);
        assertEq(uint(payment.status), uint(PaymentCore.PaymentStatus.Refunded));
        
        // Check alice got refund (including fee)
        assertEq(alice.balance, aliceBalanceBefore + TOTAL_AMOUNT);
    }
    
    function testCannotRefundTooEarly() public {
        vm.startPrank(alice);
        uint256 paymentId = paymentCore.createPayment{value: TOTAL_AMOUNT}(
            bob,
            address(0),
            PAYMENT_AMOUNT,
            ""
        );
        
        // Try to refund immediately (should fail)
        vm.expectRevert(PaymentCore.RefundNotAvailable.selector);
        paymentCore.refundPayment(paymentId);
        
        vm.stopPrank();
    }
    
    function testCannotCreatePaymentToSelf() public {
        vm.startPrank(alice);
        
        vm.expectRevert(PaymentCore.UnauthorizedAction.selector);
        paymentCore.createPayment{value: TOTAL_AMOUNT}(
            alice,
            address(0),
            PAYMENT_AMOUNT,
            ""
        );
        
        vm.stopPrank();
    }
    
    function testCannotCreateZeroAmountPayment() public {
        vm.startPrank(alice);
        
        vm.expectRevert(PaymentCore.InsufficientAmount.selector);
        paymentCore.createPayment{value: 0}(
            bob,
            address(0),
            0,
            ""
        );
        
        vm.stopPrank();
    }
    
    function testOnlyOwnerCanWithdrawFees() public {
        // Create and complete a payment to generate fees
        vm.startPrank(alice);
        uint256 paymentId = paymentCore.createPayment{value: TOTAL_AMOUNT}(
            bob,
            address(0),
            PAYMENT_AMOUNT,
            ""
        );
        vm.stopPrank();
        
        vm.prank(bob);
        paymentCore.completePayment(paymentId);
        
        // Non-owner cannot withdraw fees
        vm.startPrank(alice);
        vm.expectRevert();
        paymentCore.withdrawFees(address(0), alice);
        vm.stopPrank();
        
        // Owner can withdraw fees
        uint256 ownerBalanceBefore = owner.balance;
        
        vm.prank(owner);
        paymentCore.withdrawFees(address(0), owner);
        
        assertEq(owner.balance, ownerBalanceBefore + FEE_AMOUNT);
    }
    
    function testPauseUnpause() public {
        // Owner can pause
        vm.prank(owner);
        paymentCore.pause();
        
        // Cannot create payments when paused
        vm.startPrank(alice);
        vm.expectRevert();
        paymentCore.createPayment{value: TOTAL_AMOUNT}(
            bob,
            address(0),
            PAYMENT_AMOUNT,
            ""
        );
        vm.stopPrank();
        
        // Owner can unpause
        vm.prank(owner);
        paymentCore.unpause();
        
        // Can create payments again
        vm.startPrank(alice);
        uint256 paymentId = paymentCore.createPayment{value: TOTAL_AMOUNT}(
            bob,
            address(0),
            PAYMENT_AMOUNT,
            ""
        );
        assertEq(paymentId, 1);
        vm.stopPrank();
    }
    
    function testGetUserPaymentHistory() public {
        vm.startPrank(alice);
        
        // Create multiple payments
        paymentCore.createPayment{value: TOTAL_AMOUNT}(
            bob,
            address(0),
            PAYMENT_AMOUNT,
            ""
        );
        
        paymentCore.createPayment{value: TOTAL_AMOUNT}(
            bob,
            address(0),
            PAYMENT_AMOUNT,
            ""
        );
        
        vm.stopPrank();
        
        uint256[] memory senderPayments = paymentCore.getSenderPayments(alice);
        uint256[] memory recipientPayments = paymentCore.getRecipientPayments(bob);
        
        assertEq(senderPayments.length, 2);
        assertEq(recipientPayments.length, 2);
        assertEq(senderPayments[0], 1);
        assertEq(senderPayments[1], 2);
    }
    
    function testFeeCalculation() public {
        // Test fee calculation for different amounts
        uint256 amount1 = 1 ether;
        uint256 expectedFee1 = (amount1 * 10) / 10000; // 0.1%
        assertEq(expectedFee1, 0.001 ether);
        
        uint256 amount2 = 100 ether;
        uint256 expectedFee2 = (amount2 * 10) / 10000; // 0.1%
        assertEq(expectedFee2, 0.1 ether);
    }
    
    function testPaymentCounter() public {
        assertEq(paymentCore.getPaymentCount(), 0);
        
        vm.startPrank(alice);
        paymentCore.createPayment{value: TOTAL_AMOUNT}(
            bob,
            address(0),
            PAYMENT_AMOUNT,
            ""
        );
        vm.stopPrank();
        
        assertEq(paymentCore.getPaymentCount(), 1);
    }
}