// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import {Test, console} from "forge-std/Test.sol";
import "../src/ConfidentialPayments.sol";

contract ConfidentialPaymentsTest is Test {
    ConfidentialPayments public confidentialPayments;
    address public alice = address(0x1);
    address public bob = address(0x2);
    address public compliance = address(0x3);
    address public auditor = address(0x4);
    
    function setUp() public {
        confidentialPayments = new ConfidentialPayments();
        
        confidentialPayments.grantRole(confidentialPayments.COMPLIANCE_ROLE(), compliance);
        confidentialPayments.grantRole(confidentialPayments.AUDITOR_ROLE(), auditor);
    }

    function testCreateConfidentialPayment() public {
        vm.deal(alice, 1 ether);
        vm.startPrank(alice);
        
        bytes memory encryptedAmount = hex"0123456789abcdef";
        
        vm.expectRevert();
        uint256 paymentId = confidentialPayments.createConfidentialPayment{value: 0.1 ether}(
            bob,
            address(0),
            encryptedAmount,
            "ipfs://test-metadata",
            true
        );
        
        vm.stopPrank();
    }

    function testRequestDisclosure() public {
        vm.startPrank(compliance);
        
        vm.expectRevert(ConfidentialPayments.InvalidPaymentId.selector);
        confidentialPayments.requestDisclosure(999, "Compliance investigation");
        
        vm.stopPrank();
    }

    function testApproveDisclosure() public {
        vm.startPrank(alice);
        
        vm.expectRevert(ConfidentialPayments.InvalidPaymentId.selector);
        confidentialPayments.approveDisclosure(999);
        
        vm.stopPrank();
    }

    function testRevealPayment() public {
        vm.startPrank(compliance);
        
        vm.expectRevert(ConfidentialPayments.InvalidPaymentId.selector);
        confidentialPayments.revealPayment(999);
        
        vm.stopPrank();
    }

    function testGrantDisclosurePermission() public {
        vm.startPrank(alice);
        
        vm.expectRevert(ConfidentialPayments.InvalidPaymentId.selector);
        confidentialPayments.grantDisclosurePermission(auditor, 999);
        
        vm.stopPrank();
    }

    function testEmergencyDisclosureToggle() public {
        vm.startPrank(compliance);
        
        confidentialPayments.toggleEmergencyDisclosure();
        
        vm.stopPrank();
    }

    function testGetConfidentialPayment() public {
        vm.expectRevert(ConfidentialPayments.InvalidPaymentId.selector);
        confidentialPayments.getConfidentialPayment(999);
    }

    function testGetPaymentCount() public {
        uint256 count = confidentialPayments.getPaymentCount();
        assertEq(count, 0);
    }

    function testPauseUnpause() public {
        confidentialPayments.pause();
        assertTrue(confidentialPayments.paused());
        
        confidentialPayments.unpause();
        assertFalse(confidentialPayments.paused());
    }

    function testAccessControl() public {
        vm.startPrank(alice);
        
        vm.expectRevert();
        confidentialPayments.pause();
        
        vm.expectRevert();
        confidentialPayments.toggleEmergencyDisclosure();
        
        vm.stopPrank();
    }

    function testInvalidRecipient() public {
        vm.deal(alice, 1 ether);
        vm.startPrank(alice);
        
        bytes memory encryptedAmount = hex"0123456789abcdef";
        
        vm.expectRevert(ConfidentialPayments.UnauthorizedAction.selector);
        confidentialPayments.createConfidentialPayment{value: 0.1 ether}(
            address(0),
            address(0),
            encryptedAmount,
            "ipfs://test-metadata",
            true
        );
        
        vm.expectRevert(ConfidentialPayments.UnauthorizedAction.selector);
        confidentialPayments.createConfidentialPayment{value: 0.1 ether}(
            alice,
            address(0),
            encryptedAmount,
            "ipfs://test-metadata",
            true
        );
        
        vm.stopPrank();
    }
}