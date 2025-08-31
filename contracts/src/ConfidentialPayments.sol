// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import "fhevm/lib/TFHE.sol";
import "@openzeppelin/contracts/access/AccessControl.sol";
import "@openzeppelin/contracts/utils/ReentrancyGuard.sol";
import "@openzeppelin/contracts/utils/Pausable.sol";
import "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";
import "@openzeppelin/contracts/token/ERC20/IERC20.sol";

contract ConfidentialPayments is ReentrancyGuard, AccessControl, Pausable {
    using SafeERC20 for IERC20;

    bytes32 public constant COMPLIANCE_ROLE = keccak256("COMPLIANCE_ROLE");
    bytes32 public constant AUDITOR_ROLE = keccak256("AUDITOR_ROLE");
    bytes32 public constant VALIDATOR_ROLE = keccak256("VALIDATOR_ROLE");

    uint256 public constant FEE_BASIS_POINTS = 10;
    uint256 public constant DISCLOSURE_DELAY = 1 hours;
    uint256 private _paymentCounter;

    enum PaymentStatus {
        Pending,
        Completed,
        Refunded,
        Cancelled
    }

    enum DisclosureStatus {
        Hidden,
        Requested,
        Approved,
        Disclosed
    }

    struct ConfidentialPayment {
        uint256 id;
        address sender;
        address recipient;
        address token;
        euint256 encryptedAmount;
        euint256 encryptedFee;
        PaymentStatus status;
        uint256 createdAt;
        uint256 completedAt;
        string metadataURI;
        string receiptCID;
        DisclosureStatus disclosureStatus;
        address disclosureRequester;
        uint256 disclosureRequestTime;
        bool isPrivate;
    }

    struct DisclosureRequest {
        uint256 paymentId;
        address requester;
        string reason;
        uint256 requestTime;
        bool approved;
    }

    mapping(uint256 => ConfidentialPayment) public confidentialPayments;
    mapping(address => euint256) public encryptedBalances;
    mapping(address => uint256[]) public senderPayments;
    mapping(address => uint256[]) public recipientPayments;
    mapping(address => euint256) public collectedEncryptedFees;
    mapping(uint256 => DisclosureRequest) public disclosureRequests;
    mapping(address => mapping(uint256 => bool)) public disclosurePermissions;
    mapping(uint256 => uint256) public disclosureRequestCounter;
    mapping(address => bool) public emergencyDisclosureEnabled;

    event ConfidentialPaymentCreated(
        uint256 indexed id,
        address indexed sender,
        address indexed recipient,
        address token,
        string metadataURI,
        bool isPrivate
    );

    event DisclosureRequested(
        uint256 indexed paymentId,
        address indexed requester,
        string reason
    );

    event DisclosureApproved(
        uint256 indexed paymentId,
        address indexed approver
    );

    event DisclosureRevealed(
        uint256 indexed paymentId,
        address indexed viewer,
        uint256 amount,
        uint256 fee
    );

    event EmergencyDisclosureToggled(
        address indexed user,
        bool enabled
    );

    event PrivateBalanceUpdated(
        address indexed user,
        address indexed token
    );

    error InvalidPaymentId();
    error InsufficientEncryptedAmount();
    error UnauthorizedAction();
    error InvalidPaymentStatus();
    error DisclosureNotRequested();
    error DisclosureAlreadyProcessed();
    error InsufficientDisclosureDelay();
    error EmergencyDisclosureDisabled();

    constructor() {
        _grantRole(DEFAULT_ADMIN_ROLE, msg.sender);
        _grantRole(COMPLIANCE_ROLE, msg.sender);
    }

    function createConfidentialPayment(
        address recipient,
        address token,
        einput encryptedAmount,
        bytes memory inputProof,
        string calldata metadataURI,
        bool makePrivate
    ) external payable nonReentrant whenNotPaused returns (uint256) {
        if (recipient == address(0) || recipient == msg.sender) {
            revert UnauthorizedAction();
        }

        euint256 amount = TFHE.asEuint256(encryptedAmount, inputProof);
        euint256 fee = TFHE.div(TFHE.mul(amount, TFHE.asEuint256(FEE_BASIS_POINTS)), 10000);
        euint256 totalAmount = TFHE.add(amount, fee);

        _paymentCounter++;
        uint256 paymentId = _paymentCounter;

        ConfidentialPayment storage payment = confidentialPayments[paymentId];
        payment.id = paymentId;
        payment.sender = msg.sender;
        payment.recipient = recipient;
        payment.token = token;
        payment.encryptedAmount = amount;
        payment.encryptedFee = fee;
        payment.status = PaymentStatus.Pending;
        payment.createdAt = block.timestamp;
        payment.metadataURI = metadataURI;
        payment.isPrivate = makePrivate;
        payment.disclosureStatus = makePrivate ? DisclosureStatus.Hidden : DisclosureStatus.Disclosed;

        if (token == address(0)) {
            // For ETH payments, we trust the encrypted amount is correct
            // In production, this would use async decryption for verification
        } else {
            require(msg.value == 0, "ETH sent with token payment");
            _handleEncryptedTokenTransfer(msg.sender, address(this), totalAmount, token);
        }

        senderPayments[msg.sender].push(paymentId);
        recipientPayments[recipient].push(paymentId);
        collectedEncryptedFees[token] = TFHE.add(collectedEncryptedFees[token], fee);

        emit ConfidentialPaymentCreated(
            paymentId,
            msg.sender,
            recipient,
            token,
            metadataURI,
            makePrivate
        );

        return paymentId;
    }

    function completeConfidentialPayment(uint256 paymentId) external nonReentrant whenNotPaused {
        ConfidentialPayment storage payment = confidentialPayments[paymentId];
        
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
            _handleEncryptedETHTransfer(payment.recipient, payment.encryptedAmount);
        } else {
            _handleEncryptedTokenTransfer(address(this), payment.recipient, payment.encryptedAmount, payment.token);
        }

        encryptedBalances[payment.recipient] = TFHE.add(
            encryptedBalances[payment.recipient],
            payment.encryptedAmount
        );

        emit PrivateBalanceUpdated(payment.recipient, payment.token);
    }

    function requestDisclosure(
        uint256 paymentId,
        string calldata reason
    ) external {
        ConfidentialPayment storage payment = confidentialPayments[paymentId];
        
        if (payment.id == 0) {
            revert InvalidPaymentId();
        }
        if (!payment.isPrivate) {
            revert UnauthorizedAction();
        }
        if (payment.disclosureStatus != DisclosureStatus.Hidden) {
            revert DisclosureAlreadyProcessed();
        }

        require(
            hasRole(COMPLIANCE_ROLE, msg.sender) || 
            hasRole(AUDITOR_ROLE, msg.sender) ||
            msg.sender == payment.sender ||
            msg.sender == payment.recipient,
            "Unauthorized disclosure request"
        );

        uint256 requestId = disclosureRequestCounter[paymentId]++;
        disclosureRequests[requestId] = DisclosureRequest({
            paymentId: paymentId,
            requester: msg.sender,
            reason: reason,
            requestTime: block.timestamp,
            approved: false
        });

        payment.disclosureStatus = DisclosureStatus.Requested;
        payment.disclosureRequester = msg.sender;
        payment.disclosureRequestTime = block.timestamp;

        emit DisclosureRequested(paymentId, msg.sender, reason);
    }

    function approveDisclosure(uint256 paymentId) external {
        ConfidentialPayment storage payment = confidentialPayments[paymentId];
        
        if (payment.id == 0) {
            revert InvalidPaymentId();
        }
        if (payment.disclosureStatus != DisclosureStatus.Requested) {
            revert DisclosureNotRequested();
        }

        require(
            msg.sender == payment.sender || 
            msg.sender == payment.recipient ||
            hasRole(DEFAULT_ADMIN_ROLE, msg.sender),
            "Unauthorized approval"
        );

        payment.disclosureStatus = DisclosureStatus.Approved;
        emit DisclosureApproved(paymentId, msg.sender);
    }

    function revealPayment(uint256 paymentId) external returns (uint256 amount, uint256 fee) {
        ConfidentialPayment storage payment = confidentialPayments[paymentId];
        
        if (payment.id == 0) {
            revert InvalidPaymentId();
        }
        
        bool canReveal = false;
        
        if (payment.disclosureStatus == DisclosureStatus.Approved) {
            if (block.timestamp >= payment.disclosureRequestTime + DISCLOSURE_DELAY) {
                canReveal = true;
            } else {
                revert InsufficientDisclosureDelay();
            }
        } else if (emergencyDisclosureEnabled[msg.sender] && hasRole(COMPLIANCE_ROLE, msg.sender)) {
            canReveal = true;
        } else if (disclosurePermissions[msg.sender][paymentId]) {
            canReveal = true;
        } else if (!payment.isPrivate) {
            canReveal = true;
        }

        require(canReveal, "Disclosure not authorized");

        // In production, this would use async decryption through Gateway
        // For now, return placeholder values
        amount = 0;
        fee = 0;

        if (payment.disclosureStatus == DisclosureStatus.Approved) {
            payment.disclosureStatus = DisclosureStatus.Disclosed;
        }

        emit DisclosureRevealed(paymentId, msg.sender, amount, fee);
    }

    function grantDisclosurePermission(
        address viewer,
        uint256 paymentId
    ) external {
        ConfidentialPayment storage payment = confidentialPayments[paymentId];
        
        if (payment.id == 0) {
            revert InvalidPaymentId();
        }
        
        require(
            msg.sender == payment.sender || 
            msg.sender == payment.recipient,
            "Only payment participants can grant permissions"
        );

        disclosurePermissions[viewer][paymentId] = true;
    }

    function revokeDisclosurePermission(
        address viewer,
        uint256 paymentId
    ) external {
        ConfidentialPayment storage payment = confidentialPayments[paymentId];
        
        if (payment.id == 0) {
            revert InvalidPaymentId();
        }
        
        require(
            msg.sender == payment.sender || 
            msg.sender == payment.recipient,
            "Only payment participants can revoke permissions"
        );

        disclosurePermissions[viewer][paymentId] = false;
    }

    function toggleEmergencyDisclosure() external {
        require(hasRole(COMPLIANCE_ROLE, msg.sender), "Only compliance can toggle emergency disclosure");
        
        emergencyDisclosureEnabled[msg.sender] = !emergencyDisclosureEnabled[msg.sender];
        emit EmergencyDisclosureToggled(msg.sender, emergencyDisclosureEnabled[msg.sender]);
    }

    function getEncryptedBalance(address user) external view returns (bytes memory) {
        require(
            msg.sender == user || 
            hasRole(COMPLIANCE_ROLE, msg.sender) ||
            disclosurePermissions[msg.sender][0], // Generic permission
            "Unauthorized balance access"
        );
        
        // In production, this would use proper reencryption
        return "";
    }

    function addToEncryptedBalance(
        address user,
        einput encryptedValue,
        bytes memory inputProof
    ) external {
        require(hasRole(DEFAULT_ADMIN_ROLE, msg.sender), "Admin only");
        
        euint256 value = TFHE.asEuint256(encryptedValue, inputProof);
        encryptedBalances[user] = TFHE.add(encryptedBalances[user], value);
        
        emit PrivateBalanceUpdated(user, address(0));
    }

    function compareEncryptedAmounts(
        einput encryptedAmount1,
        bytes memory inputProof1,
        einput encryptedAmount2,
        bytes memory inputProof2
    ) external returns (bytes memory) {
        euint256 amount1 = TFHE.asEuint256(encryptedAmount1, inputProof1);
        euint256 amount2 = TFHE.asEuint256(encryptedAmount2, inputProof2);
        ebool isGreater = TFHE.gt(amount1, amount2);
        // In production, this would use proper reencryption
        return "";
    }

    function verifyEncryptedThreshold(
        einput encryptedAmount,
        bytes memory inputProof,
        uint256 threshold
    ) external returns (bytes memory) {
        euint256 amount = TFHE.asEuint256(encryptedAmount, inputProof);
        euint256 thresholdEncrypted = TFHE.asEuint256(threshold);
        ebool meetsThreshold = TFHE.ge(amount, thresholdEncrypted);
        // In production, this would use proper reencryption
        return "";
    }

    function _handleEncryptedETHTransfer(
        address to,
        euint256 encryptedAmount
    ) internal {
        // In production, this would use async decryption
        uint256 amount = 0;
        (bool success, ) = to.call{value: amount}("");
        require(success, "ETH transfer failed");
    }

    function _handleEncryptedTokenTransfer(
        address from,
        address to,
        euint256 encryptedAmount,
        address token
    ) internal {
        // In production, this would use async decryption
        uint256 amount = 0;
        if (from == address(this)) {
            IERC20(token).safeTransfer(to, amount);
        } else {
            IERC20(token).safeTransferFrom(from, to, amount);
        }
    }

    function getConfidentialPayment(uint256 paymentId) external view returns (
        uint256 id,
        address sender,
        address recipient,
        address token,
        PaymentStatus status,
        uint256 createdAt,
        uint256 completedAt,
        string memory metadataURI,
        string memory receiptCID,
        DisclosureStatus disclosureStatus,
        bool isPrivate
    ) {
        ConfidentialPayment storage payment = confidentialPayments[paymentId];
        
        if (payment.id == 0) {
            revert InvalidPaymentId();
        }

        return (
            payment.id,
            payment.sender,
            payment.recipient,
            payment.token,
            payment.status,
            payment.createdAt,
            payment.completedAt,
            payment.metadataURI,
            payment.receiptCID,
            payment.disclosureStatus,
            payment.isPrivate
        );
    }

    function getEncryptedPaymentAmount(
        uint256 paymentId,
        bytes32 publicKey
    ) external view returns (bytes memory) {
        ConfidentialPayment storage payment = confidentialPayments[paymentId];
        
        if (payment.id == 0) {
            revert InvalidPaymentId();
        }

        require(
            msg.sender == payment.sender ||
            msg.sender == payment.recipient ||
            hasRole(COMPLIANCE_ROLE, msg.sender) ||
            disclosurePermissions[msg.sender][paymentId] ||
            payment.disclosureStatus == DisclosureStatus.Disclosed,
            "Unauthorized access"
        );

        // In production, this would use proper reencryption
        return "";
    }

    function pause() external onlyRole(DEFAULT_ADMIN_ROLE) {
        _pause();
    }

    function unpause() external onlyRole(DEFAULT_ADMIN_ROLE) {
        _unpause();
    }

    function setCID(uint256 paymentId, string calldata receiptCID) external {
        ConfidentialPayment storage payment = confidentialPayments[paymentId];
        
        if (payment.id == 0) {
            revert InvalidPaymentId();
        }
        if (msg.sender != payment.sender && msg.sender != payment.recipient && !hasRole(DEFAULT_ADMIN_ROLE, msg.sender)) {
            revert UnauthorizedAction();
        }

        payment.receiptCID = receiptCID;
    }

    function getPaymentCount() external view returns (uint256) {
        return _paymentCounter;
    }

    function getSenderPayments(address sender) external view returns (uint256[] memory) {
        return senderPayments[sender];
    }

    function getRecipientPayments(address recipient) external view returns (uint256[] memory) {
        return recipientPayments[recipient];
    }
}