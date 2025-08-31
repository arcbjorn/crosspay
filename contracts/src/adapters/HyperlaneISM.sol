// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin/contracts/utils/Pausable.sol";
import "../PaymentCore.sol";

interface IInterchainSecurityModule {
    enum Types {
        UNUSED,
        ROUTING,
        AGGREGATION,
        LEGACY_MULTISIG,
        MERKLE_ROOT_MULTISIG,
        MESSAGE_ID_MULTISIG,
        NULL, // used with relayer carrying no metadata
        CCIP_READ
    }
    
    function moduleType() external view returns (uint8);
    
    function verify(
        bytes calldata _metadata,
        bytes calldata _message
    ) external returns (bool);
    
    function validatorsAndThreshold(
        bytes calldata _message
    ) external view returns (address[] memory validators, uint8 threshold);
}

interface IHyperlaneMailbox {
    function dispatch(
        uint32 destinationDomain,
        bytes32 recipientAddress,
        bytes calldata messageBody
    ) external payable returns (bytes32 messageId);
    
    function process(
        bytes calldata metadata,
        bytes calldata message
    ) external;
    
    function delivered(bytes32 messageId) external view returns (bool);
    
    function latestCheckpoint() external view returns (bytes32 root, uint32 index);
}

contract HyperlaneISM is Ownable, Pausable {
    IInterchainSecurityModule public ism;
    IHyperlaneMailbox public mailbox;
    PaymentCore public paymentCore;
    
    // Hyperlane domain identifiers
    uint32 public constant ETHEREUM_DOMAIN = 1;
    uint32 public constant BASE_DOMAIN = 8453;
    uint32 public constant ARBITRUM_DOMAIN = 42161;
    uint32 public constant POLYGON_DOMAIN = 137;
    
    mapping(uint32 => bool) public supportedDomains;
    mapping(bytes32 => uint256) public hyperlanePayments; // messageId => paymentId
    mapping(uint256 => HyperlanePayment) public paymentMappings;
    mapping(bytes32 => bool) public processedMessages;
    
    struct HyperlanePayment {
        uint256 localPaymentId;
        uint32 sourceDomain;
        uint32 destinationDomain;
        bytes32 messageId;
        bytes32 recipientAddress;
        bool delivered;
        uint256 timestamp;
        bytes metadata;
        bytes messageBody;
    }
    
    struct ValidationProof {
        address[] validators;
        uint8 threshold;
        bytes32 merkleRoot;
        uint32 checkpointIndex;
        bytes[] signatures;
        bool verified;
    }
    
    event HyperlanePaymentDispatched(
        uint256 indexed paymentId,
        bytes32 indexed messageId,
        uint32 indexed destinationDomain,
        bytes32 recipientAddress
    );
    
    event HyperlanePaymentProcessed(
        bytes32 indexed messageId,
        uint32 sourceDomain,
        bool verified
    );
    
    event ISMVerificationCompleted(
        bytes32 indexed messageId,
        address[] validators,
        uint8 threshold,
        bool verified
    );
    
    error UnsupportedDomain(uint32 domain);
    error MessageAlreadyProcessed(bytes32 messageId);
    error ISMVerificationFailed(bytes32 messageId);
    error InvalidValidatorsThreshold(uint8 threshold, uint256 validatorCount);
    error InsufficientSignatures(uint256 received, uint256 required);
    
    constructor(
        address _ism,
        address _mailbox,
        address _paymentCore
    ) Ownable(msg.sender) {
        ism = IInterchainSecurityModule(_ism);
        mailbox = IHyperlaneMailbox(_mailbox);
        paymentCore = PaymentCore(_paymentCore);
        
        // Initialize supported domains
        supportedDomains[ETHEREUM_DOMAIN] = true;
        supportedDomains[BASE_DOMAIN] = true;
        supportedDomains[ARBITRUM_DOMAIN] = true;
        supportedDomains[POLYGON_DOMAIN] = true;
    }
    
    function dispatchPayment(
        uint256 paymentId,
        uint32 destinationDomain,
        address destinationRecipient
    ) external payable returns (bytes32 messageId) {
        require(!paused(), "Pausable: paused");
        if (!supportedDomains[destinationDomain]) {
            revert UnsupportedDomain(destinationDomain);
        }
        
        // Verify payment exists and sender is authorized
        PaymentCore.Payment memory payment = paymentCore.getPayment(paymentId);
        require(
            msg.sender == payment.sender || msg.sender == payment.recipient,
            "Unauthorized"
        );
        
        // Encode payment data for Hyperlane message
        bytes memory messageBody = abi.encode(
            paymentId,
            payment.sender,
            destinationRecipient,
            payment.token,
            payment.amount,
            payment.metadataURI,
            block.timestamp
        );
        
        // Dispatch via Hyperlane
        bytes32 recipientAddress = _addressToBytes32(address(this));
        messageId = mailbox.dispatch{value: msg.value}(
            destinationDomain,
            recipientAddress,
            messageBody
        );
        
        // Store Hyperlane payment mapping
        paymentMappings[paymentId] = HyperlanePayment({
            localPaymentId: paymentId,
            sourceDomain: _getCurrentDomain(),
            destinationDomain: destinationDomain,
            messageId: messageId,
            recipientAddress: recipientAddress,
            delivered: false,
            timestamp: block.timestamp,
            metadata: "",
            messageBody: messageBody
        });
        
        hyperlanePayments[messageId] = paymentId;
        
        emit HyperlanePaymentDispatched(
            paymentId,
            messageId,
            destinationDomain,
            recipientAddress
        );
        
        return messageId;
    }
    
    function handle(
        uint32 origin,
        bytes32 sender,
        bytes calldata messageBody
    ) external {
        require(msg.sender == address(mailbox), "Only mailbox");
        
        bytes32 messageId = keccak256(abi.encodePacked(origin, sender, messageBody));
        
        if (processedMessages[messageId]) {
            revert MessageAlreadyProcessed(messageId);
        }
        
        // Decode payment data
        (
            uint256 sourcePaymentId,
            address senderAddr,
            address recipient,
            address token,
            uint256 amount,
            string memory metadataURI,
            uint256 timestamp
        ) = abi.decode(messageBody, (uint256, address, address, address, uint256, string, uint256));
        
        // Avoid unused variable warnings
        sourcePaymentId; // Source payment ID (informational only)
        senderAddr; // Sender address (informational only)
        
        // Create local payment record via trusted adapter path
        uint256 localPaymentId = paymentCore.createPaymentFromAdapter(
            recipient,
            token,
            amount,
            metadataURI
        );
        
        // Mark as Hyperlane processed
        HyperlanePayment memory hlPayment = HyperlanePayment({
            localPaymentId: localPaymentId,
            sourceDomain: origin,
            destinationDomain: _getCurrentDomain(),
            messageId: messageId,
            recipientAddress: sender,
            delivered: true,
            timestamp: timestamp,
            metadata: "",
            messageBody: messageBody
        });
        
        paymentMappings[localPaymentId] = hlPayment;
        processedMessages[messageId] = true;
        
        emit HyperlanePaymentProcessed(messageId, origin, true);
    }
    
    function verifyISMProof(
        uint256 paymentId,
        bytes calldata metadata
    ) external returns (ValidationProof memory proof) {
        HyperlanePayment storage hlPayment = paymentMappings[paymentId];
        require(hlPayment.localPaymentId != 0, "Payment not found");
        
        // Store metadata for ISM verification
        hlPayment.metadata = metadata;
        
        // Get validators and threshold from ISM
        (address[] memory validators, uint8 threshold) = ism.validatorsAndThreshold(
            hlPayment.messageBody
        );
        
        if (threshold > validators.length) {
            revert InvalidValidatorsThreshold(threshold, validators.length);
        }
        
        // Verify using ISM
        bool verified = ism.verify(metadata, hlPayment.messageBody);
        
        if (!verified) {
            revert ISMVerificationFailed(hlPayment.messageId);
        }
        
        // Get latest checkpoint for additional verification
        (bytes32 root, uint32 index) = mailbox.latestCheckpoint();
        
        proof = ValidationProof({
            validators: validators,
            threshold: threshold,
            merkleRoot: root,
            checkpointIndex: index,
            signatures: new bytes[](0), // Would be extracted from metadata
            verified: verified
        });
        
        emit ISMVerificationCompleted(
            hlPayment.messageId,
            validators,
            threshold,
            verified
        );
        
        return proof;
    }
    
    function processWithISM(
        bytes calldata metadata,
        bytes calldata message
    ) external whenNotPaused {
        // Verify message with ISM before processing
        bool verified = ism.verify(metadata, message);
        if (!verified) {
            revert ISMVerificationFailed(keccak256(message));
        }
        
        // Process the verified message
        mailbox.process(metadata, message);
        
        bytes32 messageId = keccak256(message);
        processedMessages[messageId] = true;
        
        emit HyperlanePaymentProcessed(messageId, 0, verified);
    }
    
    function isMessageDelivered(bytes32 messageId) external view returns (bool) {
        return mailbox.delivered(messageId);
    }
    
    function getHyperlanePayment(uint256 paymentId) external view returns (HyperlanePayment memory) {
        return paymentMappings[paymentId];
    }
    
    function getValidationProof(bytes32 messageId) external view returns (address[] memory validators, uint8 threshold) {
        uint256 paymentId = hyperlanePayments[messageId];
        HyperlanePayment memory hlPayment = paymentMappings[paymentId];
        
        if (hlPayment.localPaymentId == 0) {
            return (new address[](0), 0);
        }
        
        return ism.validatorsAndThreshold(hlPayment.messageBody);
    }
    
    function addSupportedDomain(uint32 domain) external onlyOwner {
        supportedDomains[domain] = true;
    }
    
    function removeSupportedDomain(uint32 domain) external onlyOwner {
        supportedDomains[domain] = false;
    }
    
    function setISM(address _ism) external onlyOwner {
        ism = IInterchainSecurityModule(_ism);
    }
    
    function setMailbox(address _mailbox) external onlyOwner {
        mailbox = IHyperlaneMailbox(_mailbox);
    }
    
    function pause() external onlyOwner {
        _pause();
    }
    
    function unpause() external onlyOwner {
        _unpause();
    }
    
    function _getCurrentDomain() internal view returns (uint32) {
        uint256 chainId = block.chainid;
        if (chainId == 1) return ETHEREUM_DOMAIN;
        if (chainId == 8453) return BASE_DOMAIN;
        if (chainId == 42161) return ARBITRUM_DOMAIN;
        if (chainId == 137) return POLYGON_DOMAIN;
        return uint32(chainId);
    }
    
    function _addressToBytes32(address addr) internal pure returns (bytes32) {
        return bytes32(uint256(uint160(addr)));
    }
    
    function _bytes32ToAddress(bytes32 b) internal pure returns (address) {
        return address(uint160(uint256(b)));
    }
    
    // Emergency functions
    function emergencyWithdraw() external onlyOwner {
        payable(owner()).transfer(address(this).balance);
    }
    
    receive() external payable {}
}
