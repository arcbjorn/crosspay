// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin/contracts/utils/Pausable.sol";

contract ReceiptRegistry is Ownable, Pausable {
    struct Receipt {
        uint256 paymentId;
        string metadataCID;
        string receiptCID;
        uint256 timestamp;
        address creator;
        bool isPublic;
    }

    mapping(uint256 => Receipt) public receipts;
    mapping(address => uint256[]) public userReceipts;
    
    uint256 private _receiptCounter;

    event ReceiptCreated(
        uint256 indexed receiptId,
        uint256 indexed paymentId,
        string metadataCID,
        string receiptCID,
        address indexed creator,
        bool isPublic
    );

    event ReceiptUpdated(
        uint256 indexed receiptId,
        string newReceiptCID,
        address indexed updater
    );

    error InvalidReceiptId();
    error UnauthorizedAccess();
    error EmptyMetadata();

    constructor() Ownable(msg.sender) {}

    function createReceipt(
        uint256 paymentId,
        string calldata metadataCID,
        string calldata receiptCID,
        bool isPublic
    ) external whenNotPaused returns (uint256) {
        if (bytes(metadataCID).length == 0) {
            revert EmptyMetadata();
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

        userReceipts[msg.sender].push(receiptId);

        emit ReceiptCreated(
            receiptId,
            paymentId,
            metadataCID,
            receiptCID,
            msg.sender,
            isPublic
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

    function getReceipt(uint256 receiptId) external view returns (Receipt memory) {
        Receipt storage receipt = receipts[receiptId];
        
        if (receipt.creator == address(0)) {
            revert InvalidReceiptId();
        }
        if (!receipt.isPublic && receipt.creator != msg.sender) {
            revert UnauthorizedAccess();
        }

        return receipt;
    }

    function getUserReceipts(address user) external view returns (uint256[] memory) {
        return userReceipts[user];
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