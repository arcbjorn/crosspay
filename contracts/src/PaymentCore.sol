// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import "@openzeppelin/contracts/utils/ReentrancyGuard.sol";
import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin/contracts/utils/Pausable.sol";
import "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";
import "@openzeppelin/contracts/token/ERC20/IERC20.sol";

contract PaymentCore is ReentrancyGuard, Ownable, Pausable {
    using SafeERC20 for IERC20;

    uint256 public constant FEE_BASIS_POINTS = 10; // 0.1%
    uint256 public constant REFUND_DELAY = 24 hours;
    uint256 private _paymentCounter;

    enum PaymentStatus {
        Pending,
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
        string metadataURI
    );

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

    constructor() Ownable(msg.sender) {}

    function createPayment(
        address recipient,
        address token,
        uint256 amount,
        string calldata metadataURI
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
        payment.status = PaymentStatus.Pending;
        payment.createdAt = block.timestamp;
        payment.metadataURI = metadataURI;

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

        emit PaymentCreated(
            paymentId,
            msg.sender,
            recipient,
            token,
            amount,
            fee,
            metadataURI
        );

        return paymentId;
    }

    function completePayment(uint256 paymentId) external nonReentrant whenNotPaused {
        Payment storage payment = payments[paymentId];
        
        if (payment.id == 0) {
            revert InvalidPaymentId();
        }
        if (payment.status != PaymentStatus.Pending) {
            revert InvalidPaymentStatus();
        }
        if (msg.sender != payment.recipient && msg.sender != payment.sender) {
            revert UnauthorizedAction();
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

        collectedFees[payment.token] -= payment.fee;

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

        collectedFees[payment.token] -= payment.fee;

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

    function getPaymentCount() external view returns (uint256) {
        return _paymentCounter;
    }
}