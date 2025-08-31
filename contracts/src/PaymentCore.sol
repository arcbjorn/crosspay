// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import "@openzeppelin/contracts/utils/ReentrancyGuard.sol";
import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin/contracts/utils/Pausable.sol";
import "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";
import "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import "./RelayValidator.sol";

contract PaymentCore is ReentrancyGuard, Ownable, Pausable {
    using SafeERC20 for IERC20;

    uint256 public constant FEE_BASIS_POINTS = 10; // 0.1%
    uint256 public constant REFUND_DELAY = 24 hours;
    uint256 public constant HIGH_VALUE_THRESHOLD = 1000 ether; // Requires validator approval
    uint256 private _paymentCounter;
    
    RelayValidator public relayValidator;

    enum PaymentStatus {
        Pending,
        ValidatorRequired,
        Completed,
        Refunded,
        Cancelled
    }

    struct Payment {
        uint256 id;
        address sender;
        address recipient;
        address token;
        uint256 amount;
        uint256 fee;
        PaymentStatus status;
        uint256 createdAt;
        uint256 completedAt;
        string metadataURI;
        string receiptCID;
        string senderENS;
        string recipientENS;
        string oraclePrice;
        bytes32 randomSeed;
        uint256 validatorRequestId;
        bool requiresValidation;
    }

    mapping(uint256 => Payment) public payments;
    mapping(address => uint256[]) public senderPayments;
    mapping(address => uint256[]) public recipientPayments;
    mapping(address => uint256) public collectedFees;

    event PaymentCreated(
        uint256 indexed id,
        address indexed sender,
        address indexed recipient,
        address token,
        uint256 amount,
        uint256 fee,
        string metadataURI,
        string senderENS,
        string recipientENS
    );

    event ReceiptStored(uint256 indexed paymentId, string receiptCID);
    event OraclePriceSet(uint256 indexed paymentId, string price);
    event RandomSeedSet(uint256 indexed paymentId, bytes32 seed);
    event ValidatorApprovalRequired(uint256 indexed paymentId, uint256 indexed requestId);
    event ValidatorApprovalReceived(uint256 indexed paymentId, uint256 indexed requestId);

    event PaymentCompleted(uint256 indexed id, address indexed completer);
    event PaymentRefunded(uint256 indexed id, address indexed refunder);
    event PaymentCancelled(uint256 indexed id, address indexed canceller);
    event FeesWithdrawn(address indexed token, uint256 amount, address indexed to);

    error InvalidPaymentId();
    error InsufficientAmount();
    error UnauthorizedAction();
    error InvalidPaymentStatus();
    error RefundNotAvailable();
    error TransferFailed();
    error ValidatorApprovalPending();
    error InvalidValidatorSignature();

    constructor() Ownable(msg.sender) {}

    function setRelayValidator(address _relayValidator) external onlyOwner {
        relayValidator = RelayValidator(_relayValidator);
    }

    function createPayment(
        address recipient,
        address token,
        uint256 amount,
        string calldata metadataURI,
        string calldata senderENS,
        string calldata recipientENS
    ) external payable nonReentrant whenNotPaused returns (uint256) {
        if (recipient == address(0) || recipient == msg.sender) {
            revert UnauthorizedAction();
        }
        if (amount == 0) {
            revert InsufficientAmount();
        }

        uint256 fee = (amount * FEE_BASIS_POINTS) / 10000;
        uint256 totalAmount = amount + fee;

        _paymentCounter++;
        uint256 paymentId = _paymentCounter;

        Payment storage payment = payments[paymentId];
        payment.id = paymentId;
        payment.sender = msg.sender;
        payment.recipient = recipient;
        payment.token = token;
        payment.amount = amount;
        payment.fee = fee;
        // Check if high-value payment requires validator approval
        bool requiresValidation = amount >= HIGH_VALUE_THRESHOLD;
        payment.requiresValidation = requiresValidation;
        payment.status = requiresValidation ? PaymentStatus.ValidatorRequired : PaymentStatus.Pending;
        payment.createdAt = block.timestamp;
        payment.metadataURI = metadataURI;
        payment.senderENS = senderENS;
        payment.recipientENS = recipientENS;

        if (token == address(0)) {
            if (msg.value != totalAmount) {
                revert InsufficientAmount();
            }
        } else {
            if (msg.value != 0) {
                revert InsufficientAmount();
            }
            IERC20(token).safeTransferFrom(msg.sender, address(this), totalAmount);
        }

        senderPayments[msg.sender].push(paymentId);
        recipientPayments[recipient].push(paymentId);
        collectedFees[token] += fee;

        // Request validator approval for high-value payments
        if (requiresValidation && address(relayValidator) != address(0)) {
            bytes32 messageHash = keccak256(abi.encodePacked(paymentId, msg.sender, recipient, amount, block.timestamp));
            uint256 requestId = relayValidator.requestValidation(paymentId, messageHash, amount);
            payment.validatorRequestId = requestId;
            emit ValidatorApprovalRequired(paymentId, requestId);
        }

        emit PaymentCreated(
            paymentId,
            msg.sender,
            recipient,
            token,
            amount,
            fee,
            metadataURI,
            senderENS,
            recipientENS
        );

        return paymentId;
    }

    function completePayment(uint256 paymentId) external nonReentrant whenNotPaused {
        Payment storage payment = payments[paymentId];
        
        if (payment.id == 0) {
            revert InvalidPaymentId();
        }
        if (payment.status != PaymentStatus.Pending && payment.status != PaymentStatus.ValidatorRequired) {
            revert InvalidPaymentStatus();
        }
        if (msg.sender != payment.recipient && msg.sender != payment.sender) {
            revert UnauthorizedAction();
        }
        
        // Check validator approval for high-value payments
        if (payment.requiresValidation) {
            if (address(relayValidator) == address(0)) {
                revert ValidatorApprovalPending();
            }
            
            // Verify the validator request is completed and valid
            (,,,, , RelayValidator.ValidationStatus status,,, ) = relayValidator.getValidationRequest(payment.validatorRequestId);
            if (status != RelayValidator.ValidationStatus.Completed) {
                revert ValidatorApprovalPending();
            }
            
            // CRITICAL: Verify the aggregated BLS signature on-chain
            bool signatureValid = relayValidator.verifyAggregatedSignature(payment.validatorRequestId);
            if (!signatureValid) {
                revert InvalidValidatorSignature();
            }
            
            emit ValidatorApprovalReceived(paymentId, payment.validatorRequestId);
        }

        payment.status = PaymentStatus.Completed;
        payment.completedAt = block.timestamp;

        if (payment.token == address(0)) {
            (bool success, ) = payment.recipient.call{value: payment.amount}("");
            if (!success) {
                revert TransferFailed();
            }
        } else {
            IERC20(payment.token).safeTransfer(payment.recipient, payment.amount);
        }

        emit PaymentCompleted(paymentId, msg.sender);
    }

    function refundPayment(uint256 paymentId) external nonReentrant whenNotPaused {
        Payment storage payment = payments[paymentId];
        
        if (payment.id == 0) {
            revert InvalidPaymentId();
        }
        if (payment.status != PaymentStatus.Pending) {
            revert InvalidPaymentStatus();
        }
        if (msg.sender != payment.sender) {
            revert UnauthorizedAction();
        }
        if (block.timestamp < payment.createdAt + REFUND_DELAY) {
            revert RefundNotAvailable();
        }

        payment.status = PaymentStatus.Refunded;
        payment.completedAt = block.timestamp;

        if (collectedFees[payment.token] >= payment.fee) {
            collectedFees[payment.token] -= payment.fee;
        } else {
            collectedFees[payment.token] = 0;
        }

        uint256 refundAmount = payment.amount + payment.fee;
        
        if (payment.token == address(0)) {
            (bool success, ) = payment.sender.call{value: refundAmount}("");
            if (!success) {
                revert TransferFailed();
            }
        } else {
            IERC20(payment.token).safeTransfer(payment.sender, refundAmount);
        }

        emit PaymentRefunded(paymentId, msg.sender);
    }

    function cancelPayment(uint256 paymentId) external nonReentrant whenNotPaused {
        Payment storage payment = payments[paymentId];
        
        if (payment.id == 0) {
            revert InvalidPaymentId();
        }
        if (payment.status != PaymentStatus.Pending) {
            revert InvalidPaymentStatus();
        }
        if (msg.sender != payment.sender && msg.sender != payment.recipient) {
            revert UnauthorizedAction();
        }

        payment.status = PaymentStatus.Cancelled;
        payment.completedAt = block.timestamp;

        if (collectedFees[payment.token] >= payment.fee) {
            collectedFees[payment.token] -= payment.fee;
        } else {
            collectedFees[payment.token] = 0;
        }

        uint256 refundAmount = payment.amount + payment.fee;
        
        if (payment.token == address(0)) {
            (bool success, ) = payment.sender.call{value: refundAmount}("");
            if (!success) {
                revert TransferFailed();
            }
        } else {
            IERC20(payment.token).safeTransfer(payment.sender, refundAmount);
        }

        emit PaymentCancelled(paymentId, msg.sender);
    }

    function getPayment(uint256 paymentId) external view returns (Payment memory) {
        if (payments[paymentId].id == 0) {
            revert InvalidPaymentId();
        }
        return payments[paymentId];
    }

    function getSenderPayments(address sender) external view returns (uint256[] memory) {
        return senderPayments[sender];
    }

    function getRecipientPayments(address recipient) external view returns (uint256[] memory) {
        return recipientPayments[recipient];
    }

    function withdrawFees(address token, address to) external onlyOwner {
        uint256 amount = collectedFees[token];
        if (amount == 0) {
            revert InsufficientAmount();
        }

        collectedFees[token] = 0;

        if (token == address(0)) {
            (bool success, ) = to.call{value: amount}("");
            if (!success) {
                revert TransferFailed();
            }
        } else {
            IERC20(token).safeTransfer(to, amount);
        }

        emit FeesWithdrawn(token, amount, to);
    }

    function pause() external onlyOwner {
        _pause();
    }

    function unpause() external onlyOwner {
        _unpause();
    }

    function setCID(uint256 paymentId, string calldata receiptCID) external {
        Payment storage payment = payments[paymentId];
        
        if (payment.id == 0) {
            revert InvalidPaymentId();
        }
        if (msg.sender != payment.sender && msg.sender != payment.recipient && msg.sender != owner()) {
            revert UnauthorizedAction();
        }

        payment.receiptCID = receiptCID;
        emit ReceiptStored(paymentId, receiptCID);
    }

    function setOraclePrice(uint256 paymentId, string calldata price) external {
        Payment storage payment = payments[paymentId];
        
        if (payment.id == 0) {
            revert InvalidPaymentId();
        }
        if (msg.sender != owner()) {
            revert UnauthorizedAction();
        }

        payment.oraclePrice = price;
        emit OraclePriceSet(paymentId, price);
    }

    function setRandomSeed(uint256 paymentId, bytes32 seed) external {
        Payment storage payment = payments[paymentId];
        
        if (payment.id == 0) {
            revert InvalidPaymentId();
        }
        if (msg.sender != owner()) {
            revert UnauthorizedAction();
        }

        payment.randomSeed = seed;
        emit RandomSeedSet(paymentId, seed);
    }

    function getCID(uint256 paymentId) external view returns (string memory) {
        if (payments[paymentId].id == 0) {
            revert InvalidPaymentId();
        }
        return payments[paymentId].receiptCID;
    }

    function getPaymentCount() external view returns (uint256) {
        return _paymentCounter;
    }

    function getValidationStatus(uint256 paymentId) external view returns (bool requiresValidation, uint256 requestId, RelayValidator.ValidationStatus status) {
        Payment storage payment = payments[paymentId];
        if (payment.id == 0) {
            revert InvalidPaymentId();
        }
        
        requiresValidation = payment.requiresValidation;
        requestId = payment.validatorRequestId;
        
        if (requiresValidation && address(relayValidator) != address(0)) {
            (,,,, , status,,, ) = relayValidator.getValidationRequest(requestId);
        }
    }
}
