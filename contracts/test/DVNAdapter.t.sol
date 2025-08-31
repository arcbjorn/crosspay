// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import {Test, console} from "forge-std/Test.sol";
import "../src/adapters/DVNAdapter.sol";
import "../src/PaymentCore.sol";

contract MockLayerZeroDVN {
    function verifyPayload(
        uint32,
        bytes32,
        uint64,
        bytes32
    ) external pure returns (bool verified, bytes32 digest) {
        return (true, keccak256("mock_digest"));
    }
    
    function quote(uint32, uint16, address) external pure returns (uint256 nativeFee) {
        return 0.01 ether;
    }
}

contract MockLayerZeroEndpoint {
    uint64 private _nonce = 1;
    
    function send(
        uint32,
        bytes32,
        bytes calldata,
        address,
        address,
        bytes calldata
    ) external payable returns (bytes32 messageId, uint64 nonce) {
        messageId = keccak256(abi.encodePacked(block.timestamp, _nonce));
        nonce = _nonce++;
        return (messageId, nonce);
    }
    
    function verify(uint32, bytes32, uint64) external pure returns (bool) {
        return true;
    }
}

contract DVNAdapterTest is Test {
    DVNAdapter public dvnAdapter;
    PaymentCore public paymentCore;
    MockLayerZeroDVN public mockDVN;
    MockLayerZeroEndpoint public mockEndpoint;
    
    address public alice = address(0x1);
    address public bob = address(0x2);
    
    function setUp() public {
        mockDVN = new MockLayerZeroDVN();
        mockEndpoint = new MockLayerZeroEndpoint();
        paymentCore = new PaymentCore();
        
        dvnAdapter = new DVNAdapter(
            address(mockDVN),
            address(mockEndpoint),
            address(paymentCore)
        );

        paymentCore.setRelayValidator(address(dvnAdapter));
        paymentCore.setTrustedAdapter(address(dvnAdapter), true);
        
        vm.deal(alice, 10 ether);
        vm.deal(bob, 10 ether);
    }
    
    function testInitiateCrossChainPayment() public {
        vm.startPrank(alice);
        
        // Create a payment first with correct fee calculation
        // Fee = 1 ether * 10 / 10000 = 0.001 ether
        uint256 paymentId = paymentCore.createPayment{value: 1.001 ether}(
            bob,
            address(0), // ETH
            1 ether,
            "test-metadata",
            "alice.eth",
            "bob.eth"
        );
        
        // Initiate cross-chain payment
        bytes32 messageId = dvnAdapter.initiateCrossChainPayment{value: 0.02 ether}(
            paymentId,
            30184, // BASE_EID
            bob,
            ""
        );
        
        assertTrue(messageId != bytes32(0));
        
        DVNAdapter.CrossChainPayment memory ccPayment = dvnAdapter.getCrossChainPayment(paymentId);
        assertEq(ccPayment.localPaymentId, paymentId);
        assertEq(ccPayment.destinationChain, 30184);
        assertFalse(ccPayment.verified);
        
        vm.stopPrank();
    }
    
    function testVerifyDVNProof() public {
        vm.startPrank(alice);
        
        uint256 paymentId = paymentCore.createPayment{value: 1.001 ether}(
            bob,
            address(0),
            1 ether,
            "test-metadata",
            "alice.eth",
            "bob.eth"
        );
        
        dvnAdapter.initiateCrossChainPayment{value: 0.02 ether}(
            paymentId,
            30184,
            bob,
            ""
        );
        
        vm.stopPrank();
        
        // Verify DVN proof
        bytes32 payloadHash = keccak256("test_payload");
        bool verified = dvnAdapter.verifyDVNProof(paymentId, payloadHash);
        
        assertTrue(verified);
        
        DVNAdapter.CrossChainPayment memory ccPayment = dvnAdapter.getCrossChainPayment(paymentId);
        assertTrue(ccPayment.verified);
    }
    
    function testUnsupportedChain() public {
        vm.startPrank(alice);
        
        uint256 paymentId = paymentCore.createPayment{value: 1.001 ether}(
            bob,
            address(0),
            1 ether,
            "test-metadata",
            "alice.eth",
            "bob.eth"
        );
        
        vm.expectRevert(abi.encodeWithSelector(DVNAdapter.UnsupportedChain.selector, 99999));
        dvnAdapter.initiateCrossChainPayment{value: 0.02 ether}(
            paymentId,
            99999, // Unsupported chain
            bob,
            ""
        );
        
        vm.stopPrank();
    }
    
    function testInsufficientGas() public {
        vm.startPrank(alice);
        
        uint256 paymentId = paymentCore.createPayment{value: 1.001 ether}(
            bob,
            address(0),
            1 ether,
            "test-metadata",
            "alice.eth",
            "bob.eth"
        );
        
        vm.expectRevert(DVNAdapter.InsufficientGas.selector);
        dvnAdapter.initiateCrossChainPayment{value: 0.005 ether}( // Insufficient gas
            paymentId,
            30184,
            bob,
            ""
        );
        
        vm.stopPrank();
    }
    
    function testLzReceive() public {
        uint32 srcEid = 30101; // Ethereum
        bytes32 srcAddress = bytes32(uint256(uint160(address(dvnAdapter))));
        uint64 nonce = 1;
        
        bytes memory message = abi.encode(
            123, // sourcePaymentId
            alice, // sender
            bob, // recipient
            address(0), // token (ETH)
            1 ether, // amount
            "cross-chain-metadata", // metadataURI
            block.timestamp // timestamp
        );
        
        // Simulate receiving cross-chain message
        vm.deal(address(paymentCore), 10 ether);
        vm.prank(address(mockEndpoint));
        dvnAdapter.lzReceive(srcEid, srcAddress, nonce, message);
        
        // Verify local payment was created
        uint256 expectedPaymentId = paymentCore.getPaymentCount();
        assertTrue(expectedPaymentId > 0);
    }
    
    function testAddRemoveSupportedChain() public {
        uint32 newChainEid = 12345;
        
        assertFalse(dvnAdapter.supportedChains(newChainEid));
        
        dvnAdapter.addSupportedChain(newChainEid);
        assertTrue(dvnAdapter.supportedChains(newChainEid));
        
        dvnAdapter.removeSupportedChain(newChainEid);
        assertFalse(dvnAdapter.supportedChains(newChainEid));
    }
    
    function testUnauthorizedActions() public {
        vm.startPrank(alice);
        
        vm.expectRevert();
        dvnAdapter.addSupportedChain(12345);
        
        vm.expectRevert();
        dvnAdapter.setDVN(address(0x123));
        
        vm.expectRevert();
        dvnAdapter.pause();
        
        vm.stopPrank();
    }
    
    function testPauseUnpause() public {
        dvnAdapter.pause();
        assertTrue(dvnAdapter.paused());
        
        vm.startPrank(alice);
        
        uint256 paymentId = paymentCore.createPayment{value: 1.001 ether}(
            bob,
            address(0),
            1 ether,
            "test-metadata",
            "alice.eth",
            "bob.eth"
        );
        
        vm.expectRevert("Pausable: paused");
        dvnAdapter.initiateCrossChainPayment{value: 0.02 ether}(
            paymentId,
            30184,
            bob,
            ""
        );
        
        vm.stopPrank();
        
        dvnAdapter.unpause();
        assertFalse(dvnAdapter.paused());
    }
    
    function testIsPaymentVerified() public {
        vm.startPrank(alice);
        
        uint256 paymentId = paymentCore.createPayment{value: 1.001 ether}(
            bob,
            address(0),
            1 ether,
            "test-metadata",
            "alice.eth",
            "bob.eth"
        );
        
        assertFalse(dvnAdapter.isPaymentVerified(paymentId));
        
        dvnAdapter.initiateCrossChainPayment{value: 0.02 ether}(
            paymentId,
            30184,
            bob,
            ""
        );
        
        assertFalse(dvnAdapter.isPaymentVerified(paymentId));
        
        vm.stopPrank();
        
        bytes32 payloadHash = keccak256("test_payload");
        dvnAdapter.verifyDVNProof(paymentId, payloadHash);
        
        assertTrue(dvnAdapter.isPaymentVerified(paymentId));
    }
}
