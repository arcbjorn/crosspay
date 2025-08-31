// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin/contracts/utils/Pausable.sol";
import "@openzeppelin/contracts/utils/cryptography/ECDSA.sol";
import "@openzeppelin/contracts/utils/cryptography/MessageHashUtils.sol";

contract ReceiptRegistry is Ownable, Pausable {
    using ECDSA for bytes32;
    using MessageHashUtils for bytes32;
    
    struct Receipt {
        uint256 paymentId;
        string metadataCID;
        string receiptCID;
        uint256 timestamp;
        address creator;
        bool isPublic;
        string language;
        string format;
        uint256 version;
        bytes32 contentHash;
        bytes signature;
        bool verified;
        string complianceFields;
        mapping(string => string) customFields;
    }
    
    struct ReceiptMetadata {
        uint256 paymentId;
        string metadataCID;
        string receiptCID;
        uint256 timestamp;
        address creator;
        bool isPublic;
        string language;
        string format;
        uint256 version;
        bytes32 contentHash;
        bool verified;
        string complianceFields;
    }

    mapping(uint256 => Receipt) public receipts;
    mapping(address => uint256[]) public userReceipts;
    mapping(bytes32 => uint256) public contentHashToReceipt;
    mapping(string => bool) public supportedLanguages;
    mapping(string => bool) public supportedFormats;
    
    uint256 private _receiptCounter;
    uint256 public constant CURRENT_VERSION = 1;

    event ReceiptCreated(
        uint256 indexed receiptId,
        uint256 indexed paymentId,
        string metadataCID,
        string receiptCID,
        address indexed creator,
        bool isPublic,
        string language,
        string format
    );

    event ReceiptUpdated(
        uint256 indexed receiptId,
        string newReceiptCID,
        address indexed updater
    );
    
    event ReceiptVerified(
        uint256 indexed receiptId,
        address indexed verifier,
        bytes32 contentHash
    );
    
    event CustomFieldSet(
        uint256 indexed receiptId,
        string key,
        string value
    );
    
    event ReceiptShared(
        uint256 indexed receiptId,
        address indexed sharedWith,
        string shareType
    );

    error InvalidReceiptId();
    error UnauthorizedAccess();
    error EmptyMetadata();
    error UnsupportedLanguage();
    error UnsupportedFormat();
    error InvalidSignature();
    error ReceiptAlreadyVerified();
    error DuplicateContentHash();

    constructor() Ownable(msg.sender) {
        // Initialize supported languages
        supportedLanguages["en"] = true;
        supportedLanguages["es"] = true;
        supportedLanguages["fr"] = true;
        
        // Initialize supported formats
        supportedFormats["json"] = true;
        supportedFormats["pdf"] = true;
    }

    function createReceipt(
        uint256 paymentId,
        string calldata metadataCID,
        string calldata receiptCID,
        bool isPublic,
        string calldata language,
        string calldata format,
        bytes32 contentHash,
        bytes calldata signature,
        string calldata complianceFields
    ) external whenNotPaused returns (uint256) {
        if (bytes(metadataCID).length == 0) {
            revert EmptyMetadata();
        }
        
        if (!supportedLanguages[language]) {
            revert UnsupportedLanguage();
        }
        
        if (!supportedFormats[format]) {
            revert UnsupportedFormat();
        }
        
        if (contentHashToReceipt[contentHash] != 0) {
            revert DuplicateContentHash();
        }

        _receiptCounter++;
        uint256 receiptId = _receiptCounter;

        Receipt storage receipt = receipts[receiptId];
        receipt.paymentId = paymentId;
        receipt.metadataCID = metadataCID;
        receipt.receiptCID = receiptCID;
        receipt.timestamp = block.timestamp;
        receipt.creator = msg.sender;
        receipt.isPublic = isPublic;
        receipt.language = language;
        receipt.format = format;
        receipt.version = CURRENT_VERSION;
        receipt.contentHash = contentHash;
        receipt.signature = signature;
        receipt.verified = false;
        receipt.complianceFields = complianceFields;

        userReceipts[msg.sender].push(receiptId);
        contentHashToReceipt[contentHash] = receiptId;

        emit ReceiptCreated(
            receiptId,
            paymentId,
            metadataCID,
            receiptCID,
            msg.sender,
            isPublic,
            language,
            format
        );

        return receiptId;
    }

    function updateReceiptCID(
        uint256 receiptId,
        string calldata newReceiptCID
    ) external whenNotPaused {
        Receipt storage receipt = receipts[receiptId];
        
        if (receipt.creator == address(0)) {
            revert InvalidReceiptId();
        }
        if (receipt.creator != msg.sender) {
            revert UnauthorizedAccess();
        }

        receipt.receiptCID = newReceiptCID;

        emit ReceiptUpdated(receiptId, newReceiptCID, msg.sender);
    }

    function verifyReceipt(
        uint256 receiptId,
        bytes32 contentHash
    ) external whenNotPaused {
        Receipt storage receipt = receipts[receiptId];
        
        if (receipt.creator == address(0)) {
            revert InvalidReceiptId();
        }
        
        if (receipt.verified) {
            revert ReceiptAlreadyVerified();
        }
        
        if (receipt.contentHash != contentHash) {
            revert InvalidSignature();
        }
        
        // Verify the signature (simplified - in production would verify against creator's signature)
        bytes32 messageHash = keccak256(abi.encodePacked(
            receipt.paymentId,
            receipt.receiptCID,
            contentHash
        ));
        
        address recoveredSigner = messageHash.toEthSignedMessageHash().recover(receipt.signature);
        
        if (recoveredSigner != receipt.creator && recoveredSigner != owner()) {
            revert InvalidSignature();
        }
        
        receipt.verified = true;
        
        emit ReceiptVerified(receiptId, msg.sender, contentHash);
    }
    
    function setCustomField(
        uint256 receiptId,
        string calldata key,
        string calldata value
    ) external whenNotPaused {
        Receipt storage receipt = receipts[receiptId];
        
        if (receipt.creator == address(0)) {
            revert InvalidReceiptId();
        }
        if (receipt.creator != msg.sender) {
            revert UnauthorizedAccess();
        }
        
        receipt.customFields[key] = value;
        
        emit CustomFieldSet(receiptId, key, value);
    }
    
    function getCustomField(
        uint256 receiptId,
        string calldata key
    ) external view returns (string memory) {
        Receipt storage receipt = receipts[receiptId];
        
        if (receipt.creator == address(0)) {
            revert InvalidReceiptId();
        }
        if (!receipt.isPublic && receipt.creator != msg.sender) {
            revert UnauthorizedAccess();
        }
        
        return receipt.customFields[key];
    }
    
    function shareReceipt(
        uint256 receiptId,
        address recipient,
        string calldata shareType
    ) external whenNotPaused {
        Receipt storage receipt = receipts[receiptId];
        
        if (receipt.creator == address(0)) {
            revert InvalidReceiptId();
        }
        if (receipt.creator != msg.sender) {
            revert UnauthorizedAccess();
        }
        
        // Add to recipient's receipts list for viewing access
        userReceipts[recipient].push(receiptId);
        
        emit ReceiptShared(receiptId, recipient, shareType);
    }
    
    function getReceiptMetadata(uint256 receiptId) external view returns (ReceiptMetadata memory) {
        Receipt storage receipt = receipts[receiptId];
        
        if (receipt.creator == address(0)) {
            revert InvalidReceiptId();
        }
        if (!receipt.isPublic && receipt.creator != msg.sender) {
            revert UnauthorizedAccess();
        }

        return ReceiptMetadata({
            paymentId: receipt.paymentId,
            metadataCID: receipt.metadataCID,
            receiptCID: receipt.receiptCID,
            timestamp: receipt.timestamp,
            creator: receipt.creator,
            isPublic: receipt.isPublic,
            language: receipt.language,
            format: receipt.format,
            version: receipt.version,
            contentHash: receipt.contentHash,
            verified: receipt.verified,
            complianceFields: receipt.complianceFields
        });
    }
    
    function findReceiptByContentHash(bytes32 contentHash) external view returns (uint256) {
        return contentHashToReceipt[contentHash];
    }

    function getUserReceipts(address user) external view returns (uint256[] memory) {
        return userReceipts[user];
    }

    function addSupportedLanguage(string calldata language) external onlyOwner {
        supportedLanguages[language] = true;
    }
    
    function removeSupportedLanguage(string calldata language) external onlyOwner {
        supportedLanguages[language] = false;
    }
    
    function addSupportedFormat(string calldata format) external onlyOwner {
        supportedFormats[format] = true;
    }
    
    function removeSupportedFormat(string calldata format) external onlyOwner {
        supportedFormats[format] = false;
    }
    
    function isLanguageSupported(string calldata language) external view returns (bool) {
        return supportedLanguages[language];
    }
    
    function isFormatSupported(string calldata format) external view returns (bool) {
        return supportedFormats[format];
    }
    
    function getReceiptsByPaymentId(uint256 paymentId) external view returns (uint256[] memory) {
        uint256 count = 0;
        
        // Count matching receipts
        for (uint256 i = 1; i <= _receiptCounter; i++) {
            if (receipts[i].paymentId == paymentId && 
                (receipts[i].isPublic || receipts[i].creator == msg.sender)) {
                count++;
            }
        }
        
        // Collect matching receipts
        uint256[] memory result = new uint256[](count);
        uint256 index = 0;
        
        for (uint256 i = 1; i <= _receiptCounter; i++) {
            if (receipts[i].paymentId == paymentId && 
                (receipts[i].isPublic || receipts[i].creator == msg.sender)) {
                result[index] = i;
                index++;
            }
        }
        
        return result;
    }
    
    function getVerifiedReceiptsCount() external view returns (uint256) {
        uint256 count = 0;
        for (uint256 i = 1; i <= _receiptCounter; i++) {
            if (receipts[i].verified && receipts[i].creator != address(0)) {
                count++;
            }
        }
        return count;
    }

    function getReceiptCount() external view returns (uint256) {
        return _receiptCounter;
    }

    function pause() external onlyOwner {
        _pause();
    }

    function unpause() external onlyOwner {
        _unpause();
    }
}