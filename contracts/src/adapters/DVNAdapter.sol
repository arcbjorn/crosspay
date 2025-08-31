// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin/contracts/utils/Pausable.sol";
import "../PaymentCore.sol";

interface ILayerZeroDVN {
    struct DVNConfig {
        uint256 confirmations;
        uint8 requiredDVNs;
        address[] dvnAddresses;
    }
    
    function verifyPayload(
        uint32 srcEid,
        bytes32 srcAddress,
        uint64 nonce,
        bytes32 payloadHash
    ) external view returns (bool verified, bytes32 digest);
    
    function quote(
        uint32 dstEid,
        uint16 confirmations,
        address sender
    ) external view returns (uint256 nativeFee);
}

interface ILayerZeroEndpoint {
    function send(
        uint32 _dstEid,
        bytes32 _receiver,
        bytes calldata _message,
        address _refundAddress,
        address _zroPaymentAddress,
        bytes calldata _options
    ) external payable returns (bytes32 messageId, uint64 nonce);
    
    function verify(uint32 _srcEid, bytes32 _srcAddress, uint64 _nonce) external view returns (bool);
}

contract DVNAdapter is Ownable, Pausable {
    ILayerZeroDVN public dvn;
    ILayerZeroEndpoint public endpoint;
    PaymentCore public paymentCore;
    
    uint32 public constant ETHEREUM_EID = 30101;
    uint32 public constant BASE_EID = 30184;
    uint32 public constant ARBITRUM_EID = 30110;
    uint16 public constant DEFAULT_CONFIRMATIONS = 15;
    
    mapping(uint32 => bool) public supportedChains;
    mapping(bytes32 => uint256) public crossChainPayments; // messageId => paymentId
    mapping(uint256 => CrossChainPayment) public paymentMappings; // paymentId => CrossChainPayment
    
    struct CrossChainPayment {
        uint256 localPaymentId;
        uint32 sourceChain;
        uint32 destinationChain;
        bytes32 messageId;
        bytes32 sourceAddress;
        uint64 nonce;
        bool verified;
        uint256 timestamp;
    }
    
    event CrossChainPaymentInitiated(
        uint256 indexed paymentId,
        uint32 indexed destinationChain,
        bytes32 messageId,
        address indexed recipient
    );
    
    event CrossChainPaymentVerified(
        uint256 indexed paymentId,
        bytes32 messageId,
        bool verified
    );
    
    event DVNVerificationRequested(
        bytes32 indexed messageId,
        uint32 srcChain,
        bytes32 payloadHash
    );
    
    error UnsupportedChain(uint32 chainId);
    error PaymentNotFound(uint256 paymentId);
    error VerificationFailed(bytes32 messageId);
    error InsufficientGas();
    error InvalidDVNResponse();
    
    constructor(
        address _dvn,
        address _endpoint,
        address _paymentCore
    ) Ownable(msg.sender) {
        dvn = ILayerZeroDVN(_dvn);
        endpoint = ILayerZeroEndpoint(_endpoint);
        paymentCore = PaymentCore(_paymentCore);
        
        // Initialize supported chains
        supportedChains[ETHEREUM_EID] = true;
        supportedChains[BASE_EID] = true;
        supportedChains[ARBITRUM_EID] = true;
    }
    
    function initiateCrossChainPayment(
        uint256 paymentId,
        uint32 destinationChain,
        address destinationRecipient,
        bytes calldata options
    ) external payable returns (bytes32 messageId) {
        require(!paused(), "Pausable: paused");
        if (!supportedChains[destinationChain]) {
            revert UnsupportedChain(destinationChain);
        }
        
        // Verify payment exists and sender is authorized
        PaymentCore.Payment memory payment = paymentCore.getPayment(paymentId);
        require(
            msg.sender == payment.sender || msg.sender == payment.recipient,
            "Unauthorized"
        );
        
        // Encode payment data for cross-chain message
        bytes memory message = abi.encode(
            paymentId,
            payment.sender,
            destinationRecipient,
            payment.token,
            payment.amount,
            payment.metadataURI,
            block.timestamp
        );
        
        // Calculate required gas
        uint256 nativeFee = dvn.quote(destinationChain, DEFAULT_CONFIRMATIONS, msg.sender);
        if (msg.value < nativeFee) {
            revert InsufficientGas();
        }
        
        // Send cross-chain message via LayerZero
        bytes32 destinationAddress = bytes32(uint256(uint160(address(this))));
        uint64 nonce;
        (messageId, nonce) = endpoint.send{value: msg.value}(
            destinationChain,
            destinationAddress,
            message,
            payable(msg.sender), // refund address
            address(0), // no ZRO payment
            options
        );
        
        // Store cross-chain payment mapping
        paymentMappings[paymentId] = CrossChainPayment({
            localPaymentId: paymentId,
            sourceChain: _getChainEid(),
            destinationChain: destinationChain,
            messageId: messageId,
            sourceAddress: bytes32(uint256(uint160(address(this)))),
            nonce: nonce,
            verified: false,
            timestamp: block.timestamp
        });
        
        crossChainPayments[messageId] = paymentId;
        
        emit CrossChainPaymentInitiated(paymentId, destinationChain, messageId, destinationRecipient);
        
        return messageId;
    }
    
    function verifyDVNProof(
        uint256 paymentId,
        bytes32 payloadHash
    ) external returns (bool) {
        CrossChainPayment storage ccPayment = paymentMappings[paymentId];
        if (ccPayment.localPaymentId == 0) {
            revert PaymentNotFound(paymentId);
        }
        
        // Request DVN verification
        (bool verified, bytes32 digest) = dvn.verifyPayload(
            ccPayment.sourceChain,
            ccPayment.sourceAddress,
            ccPayment.nonce,
            payloadHash
        );
        
        if (!verified) {
            revert VerificationFailed(ccPayment.messageId);
        }
        
        ccPayment.verified = true;
        
        emit DVNVerificationRequested(ccPayment.messageId, ccPayment.sourceChain, payloadHash);
        emit CrossChainPaymentVerified(paymentId, ccPayment.messageId, verified);
        
        return verified;
    }
    
    function lzReceive(
        uint32 srcEid,
        bytes32 srcAddress,
        uint64 nonce,
        bytes calldata message
    ) external {
        require(msg.sender == address(endpoint), "Only endpoint");
        
        // Verify the message via DVN
        bool verified = endpoint.verify(srcEid, srcAddress, nonce);
        if (!verified) {
            revert VerificationFailed(bytes32(uint256(nonce)));
        }
        
        // Decode and process the cross-chain payment
        (
            uint256 sourcePaymentId,
            address sender,
            address recipient,
            address token,
            uint256 amount,
            string memory metadataURI,
            uint256 timestamp
        ) = abi.decode(message, (uint256, address, address, address, uint256, string, uint256));
        
        // Create local payment record via trusted adapter path
        uint256 localPaymentId = paymentCore.createPaymentFromAdapter(
            recipient,
            token,
            amount,
            metadataURI
        );
        
        // Mark as cross-chain verified
        CrossChainPayment memory ccPayment = CrossChainPayment({
            localPaymentId: localPaymentId,
            sourceChain: srcEid,
            destinationChain: _getChainEid(),
            messageId: bytes32(uint256(sourcePaymentId)),
            sourceAddress: srcAddress,
            nonce: nonce,
            verified: true,
            timestamp: timestamp
        });
        
        paymentMappings[localPaymentId] = ccPayment;
    }
    
    function getCrossChainPayment(uint256 paymentId) external view returns (CrossChainPayment memory) {
        return paymentMappings[paymentId];
    }
    
    function isPaymentVerified(uint256 paymentId) external view returns (bool) {
        return paymentMappings[paymentId].verified;
    }
    
    function addSupportedChain(uint32 chainEid) external onlyOwner {
        supportedChains[chainEid] = true;
    }
    
    function removeSupportedChain(uint32 chainEid) external onlyOwner {
        supportedChains[chainEid] = false;
    }
    
    function setDVN(address _dvn) external onlyOwner {
        dvn = ILayerZeroDVN(_dvn);
    }
    
    function setEndpoint(address _endpoint) external onlyOwner {
        endpoint = ILayerZeroEndpoint(_endpoint);
    }
    
    function pause() external onlyOwner {
        _pause();
    }
    
    function unpause() external onlyOwner {
        _unpause();
    }
    
    function _getChainEid() internal view returns (uint32) {
        uint256 chainId = block.chainid;
        if (chainId == 1) return ETHEREUM_EID;
        if (chainId == 8453) return BASE_EID;
        if (chainId == 42161) return ARBITRUM_EID;
        return uint32(chainId); // Fallback for other chains
    }
    
    // Emergency functions
    function emergencyWithdraw() external onlyOwner {
        payable(owner()).transfer(address(this).balance);
    }
    
    receive() external payable {}
}
