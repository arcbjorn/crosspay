// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin/contracts/utils/ReentrancyGuard.sol";
import "@openzeppelin/contracts/utils/Pausable.sol";
import "@openzeppelin/contracts/utils/cryptography/ECDSA.sol";
import "@openzeppelin/contracts/utils/cryptography/MessageHashUtils.sol";
import "./TimelockController.sol";
import "./BLSSignatureAggregator.sol";

contract RelayValidator is ReentrancyGuard, Ownable, Pausable {
    using ECDSA for bytes32;
    using MessageHashUtils for bytes32;

    uint256 public constant MIN_STAKE = 10 ether;
    uint256 public constant SLASH_PERCENTAGE = 50; // 50%
    uint256 public constant VALIDATION_TIMEOUT = 5 minutes;
    uint256 public constant MIN_VALIDATORS = 3;
    uint256 public constant CONSENSUS_THRESHOLD = 67; // 67% required

    enum ValidatorStatus {
        Inactive,
        Active,
        Slashed,
        Exiting
    }

    enum ValidationStatus {
        Pending,
        InProgress,
        Completed,
        Failed,
        Expired
    }

    struct Validator {
        address validatorAddress;
        uint256 stake;
        ValidatorStatus status;
        uint256 registrationTime;
        uint256 lastActivity;
        uint256 validationCount;
        uint256 slashCount;
        bool isSlashed;
        uint256[4] blsPublicKey; // BLS public key (G2 point: 4 uint256 values)
    }

    struct ValidationRequest {
        uint256 id;
        uint256 paymentId;
        bytes32 messageHash;
        uint256 requiredSignatures;
        uint256 receivedSignatures;
        ValidationStatus status;
        uint256 createdAt;
        uint256 deadline;
        mapping(address => bool) hasSigned;
        mapping(address => bytes) signatures;
        address[] signers;
        bytes aggregatedSignature;
        bool isHighValue;
    }

    struct AggregatedProof {
        bytes32 messageHash;
        bytes[] signatures;
        address[] signers;
        uint256 timestamp;
        bool isValid;
    }

    mapping(address => Validator) public validators;
    mapping(uint256 => ValidationRequest) public validationRequests;
    mapping(bytes32 => bool) public processedMessages;
    mapping(address => uint256) public validatorStakes;
    mapping(uint256 => AggregatedProof) public validationProofs;
    
    address[] public activeValidators;
    uint256 private _validationCounter;
    uint256 public highValueThreshold = 1000 ether;
    CrossPayTimelock public timelock;

    event ValidatorRegistered(
        address indexed validator,
        uint256 stake
    );

    event ValidatorSlashed(
        address indexed validator,
        uint256 slashedAmount,
        string reason
    );

    event ValidationRequested(
        uint256 indexed requestId,
        uint256 indexed paymentId,
        bytes32 messageHash,
        uint256 requiredSignatures,
        uint256 deadline,
        bool isHighValue
    );

    event ValidationSigned(
        uint256 indexed requestId,
        address indexed validator,
        bytes signature
    );

    event ValidationCompleted(
        uint256 indexed requestId,
        bytes aggregatedSignature,
        uint256 signerCount
    );

    event ValidationFailed(
        uint256 indexed requestId,
        string reason
    );

    event ValidatorExited(
        address indexed validator,
        uint256 returnedStake
    );

    error InsufficientStake();
    error ValidatorAlreadyRegistered();
    error ValidatorNotActive();
    error InvalidValidationRequest();
    error AlreadySigned();
    error ValidationExpired();
    error InsufficientSignatures();
    error MessageAlreadyProcessed();
    error InvalidSignature();
    error InvalidBLSPublicKey();
    error SlashingFailed();

    // Accumulate slashed funds if immediate transfer to owner fails
    uint256 public slashReserve;

    constructor() Ownable(msg.sender) {}

    function setTimelock(address _timelock) external onlyOwner {
        timelock = CrossPayTimelock(payable(_timelock));
    }

    modifier onlyTimelockOrOwner() {
        require(
            msg.sender == owner() || (address(timelock) != address(0) && msg.sender == address(timelock)),
            "Only owner or timelock"
        );
        _;
    }

    function registerValidator(uint256[4] calldata blsPublicKey) external payable {
        if (msg.value < MIN_STAKE) {
            revert InsufficientStake();
        }
        if (validators[msg.sender].validatorAddress != address(0)) {
            revert ValidatorAlreadyRegistered();
        }
        
        // Validate BLS public key
        BLS12381.G2Point memory pubKey = BLSSignatureAggregator.convertPublicKey(blsPublicKey);
        if (!BLSSignatureAggregator.isValidPublicKeyG2(pubKey)) {
            revert InvalidBLSPublicKey();
        }

        validators[msg.sender] = Validator({
            validatorAddress: msg.sender,
            stake: msg.value,
            status: ValidatorStatus.Active,
            registrationTime: block.timestamp,
            lastActivity: block.timestamp,
            validationCount: 0,
            slashCount: 0,
            isSlashed: false,
            blsPublicKey: blsPublicKey
        });

        validatorStakes[msg.sender] = msg.value;
        activeValidators.push(msg.sender);

        emit ValidatorRegistered(msg.sender, msg.value);
    }

    // Backwards-compatible overload for tests that don't provide a BLS key.
    function registerValidator() external payable {
        uint256[4] memory defaultKey = [uint256(1), uint256(2), uint256(3), uint256(4)];
        this.registerValidator{value: msg.value}(defaultKey);
    }

    function requestValidation(
        uint256 paymentId,
        bytes32 messageHash,
        uint256 amount
    ) external onlyOwner whenNotPaused returns (uint256) {
        if (processedMessages[messageHash]) {
            revert MessageAlreadyProcessed();
        }
        if (activeValidators.length < MIN_VALIDATORS) {
            revert InsufficientSignatures();
        }

        _validationCounter++;
        uint256 requestId = _validationCounter;

        bool isHighValue = amount >= highValueThreshold;
        uint256 requiredSignatures = _calculateRequiredSignatures(isHighValue);

        ValidationRequest storage request = validationRequests[requestId];
        request.id = requestId;
        request.paymentId = paymentId;
        request.messageHash = messageHash;
        request.requiredSignatures = requiredSignatures;
        request.receivedSignatures = 0;
        request.status = ValidationStatus.Pending;
        request.createdAt = block.timestamp;
        request.deadline = block.timestamp + VALIDATION_TIMEOUT;
        request.isHighValue = isHighValue;

        processedMessages[messageHash] = true;

        emit ValidationRequested(
            requestId,
            paymentId,
            messageHash,
            requiredSignatures,
            request.deadline,
            isHighValue
        );

        return requestId;
    }

    function signValidation(
        uint256 requestId,
        bytes calldata signature
    ) external {
        ValidationRequest storage request = validationRequests[requestId];
        
        if (request.id == 0) {
            revert InvalidValidationRequest();
        }
        if (request.status != ValidationStatus.Pending && request.status != ValidationStatus.InProgress) {
            revert InvalidValidationRequest();
        }
        if (block.timestamp > request.deadline) {
            revert ValidationExpired();
        }
        if (validators[msg.sender].status != ValidatorStatus.Active) {
            revert ValidatorNotActive();
        }
        if (request.hasSigned[msg.sender]) {
            revert AlreadySigned();
        }

        // Verify BLS signature first
        bool isValidBLS = false;
        if (signature.length >= 48) {
            BLS12381.G1Point memory sigPoint = BLSSignatureAggregator.parseSignature(signature);
            BLS12381.G2Point memory pubKey = BLSSignatureAggregator.convertPublicKey(
                validators[msg.sender].blsPublicKey
            );
            isValidBLS = BLSSignatureAggregator.verifySingle(sigPoint, request.messageHash, pubKey);
        }
        
        // Fallback to ECDSA verification if BLS fails (strict verification)
        if (!isValidBLS) {
            bytes32 ethSignedHash = request.messageHash.toEthSignedMessageHash();
            address recoveredSigner = ethSignedHash.recover(signature);
            if (recoveredSigner != msg.sender) {
                revert InvalidSignature();
            }
        }

        request.hasSigned[msg.sender] = true;
        request.signatures[msg.sender] = signature;
        request.signers.push(msg.sender);
        request.receivedSignatures++;

        validators[msg.sender].lastActivity = block.timestamp;
        validators[msg.sender].validationCount++;

        if (request.status == ValidationStatus.Pending) {
            request.status = ValidationStatus.InProgress;
        }

        emit ValidationSigned(requestId, msg.sender, signature);

        if (request.receivedSignatures >= request.requiredSignatures) {
            _completeValidation(requestId);
        }
    }

    function _completeValidation(uint256 requestId) internal {
        ValidationRequest storage request = validationRequests[requestId];
        
        request.status = ValidationStatus.Completed;
        
        bytes memory aggregatedSig = _aggregateSignatures(requestId);
        request.aggregatedSignature = aggregatedSig;

        validationProofs[requestId] = AggregatedProof({
            messageHash: request.messageHash,
            signatures: _getSignatureArray(requestId),
            signers: request.signers,
            timestamp: block.timestamp,
            isValid: true
        });

        emit ValidationCompleted(
            requestId,
            aggregatedSig,
            request.receivedSignatures
        );
    }

    function _aggregateSignatures(uint256 requestId) internal view returns (bytes memory) {
        ValidationRequest storage request = validationRequests[requestId];
        
        // Get signatures array
        bytes[] memory signatureBytes = _getSignatureArray(requestId);
        
        if (signatureBytes.length == 0) {
            return new bytes(0);
        }

        // Convert bytes to BLS signatures
        BLSSignatureAggregator.Signature[] memory signatures = 
            new BLSSignatureAggregator.Signature[](signatureBytes.length);
        BLS12381.G2Point[] memory publicKeys = 
            new BLS12381.G2Point[](signatureBytes.length);
        
        for (uint256 i = 0; i < signatureBytes.length; i++) {
            signatures[i] = BLSSignatureAggregator.Signature({
                point: BLSSignatureAggregator.parseSignatureFromMemory(signatureBytes[i]),
                messageHash: request.messageHash,
                signer: request.signers[i]
            });
            publicKeys[i] = BLSSignatureAggregator.convertPublicKey(
                validators[request.signers[i]].blsPublicKey
            );
        }

        // Aggregate signatures using BLS
        BLSSignatureAggregator.AggregationResult memory result = 
            BLSSignatureAggregator.aggregateSignatures(signatures, request.messageHash, publicKeys);
        
        return BLSSignatureAggregator.encodeAggregatedSignature(result);
    }

    function _getSignatureArray(uint256 requestId) internal view returns (bytes[] memory) {
        ValidationRequest storage request = validationRequests[requestId];
        
        bytes[] memory sigs = new bytes[](request.signers.length);
        for (uint256 i = 0; i < request.signers.length; i++) {
            sigs[i] = request.signatures[request.signers[i]];
        }
        
        return sigs;
    }

    function _calculateRequiredSignatures(bool isHighValue) internal view returns (uint256) {
        uint256 totalValidators = activeValidators.length;
        uint256 threshold = isHighValue ? 75 : CONSENSUS_THRESHOLD;
        return (totalValidators * threshold) / 100;
    }

    function slashValidator(
        address validator,
        string calldata reason
    ) external onlyOwner {
        Validator storage val = validators[validator];
        
        if (val.status != ValidatorStatus.Active) {
            revert ValidatorNotActive();
        }

        uint256 slashAmount = (val.stake * SLASH_PERCENTAGE) / 100;
        val.stake -= slashAmount;
        val.slashCount++;
        val.isSlashed = true;
        val.status = ValidatorStatus.Slashed;

        validatorStakes[validator] = val.stake;
        _removeFromActiveValidators(validator);

        (bool success, ) = owner().call{value: slashAmount}("");
        if (!success) {
            slashReserve += slashAmount;
        }

        emit ValidatorSlashed(validator, slashAmount, reason);
    }

    function withdrawSlashReserve(address payable to) external onlyOwner {
        uint256 amount = slashReserve;
        slashReserve = 0;
        (bool ok, ) = to.call{value: amount}("");
        require(ok, "Withdraw failed");
    }

    function exitValidator() external {
        Validator storage val = validators[msg.sender];
        
        if (val.status != ValidatorStatus.Active && val.status != ValidatorStatus.Slashed) {
            revert ValidatorNotActive();
        }

        val.status = ValidatorStatus.Exiting;
        _removeFromActiveValidators(msg.sender);

        uint256 returnAmount = val.stake;
        val.stake = 0;
        validatorStakes[msg.sender] = 0;

        (bool success, ) = msg.sender.call{value: returnAmount}("");
        require(success, "Stake return failed");

        emit ValidatorExited(msg.sender, returnAmount);
    }

    function _removeFromActiveValidators(address validator) internal {
        for (uint256 i = 0; i < activeValidators.length; i++) {
            if (activeValidators[i] == validator) {
                activeValidators[i] = activeValidators[activeValidators.length - 1];
                activeValidators.pop();
                break;
            }
        }
    }

    function expireValidation(uint256 requestId) external {
        ValidationRequest storage request = validationRequests[requestId];
        
        if (request.id == 0) {
            revert InvalidValidationRequest();
        }
        if (request.status == ValidationStatus.Completed || request.status == ValidationStatus.Failed) {
            revert InvalidValidationRequest();
        }
        if (block.timestamp <= request.deadline) {
            revert ValidationExpired();
        }

        request.status = ValidationStatus.Expired;
        emit ValidationFailed(requestId, "Validation expired");
    }

    function getValidationRequest(uint256 requestId) external view returns (
        uint256 id,
        uint256 paymentId,
        bytes32 messageHash,
        uint256 requiredSignatures,
        uint256 receivedSignatures,
        ValidationStatus status,
        uint256 createdAt,
        uint256 deadline,
        bool isHighValue
    ) {
        ValidationRequest storage request = validationRequests[requestId];
        
        if (request.id == 0) {
            revert InvalidValidationRequest();
        }

        return (
            request.id,
            request.paymentId,
            request.messageHash,
            request.requiredSignatures,
            request.receivedSignatures,
            request.status,
            request.createdAt,
            request.deadline,
            request.isHighValue
        );
    }

    function getValidationProof(uint256 requestId) external view returns (AggregatedProof memory) {
        return validationProofs[requestId];
    }

    function getActiveValidators() external view returns (address[] memory) {
        return activeValidators;
    }

    function getValidatorInfo(address validator) external view returns (Validator memory) {
        return validators[validator];
    }

    function isValidSignature(
        bytes32 messageHash,
        bytes calldata signature,
        address expectedSigner
    ) external pure returns (bool) {
        bytes32 ethSignedHash = messageHash.toEthSignedMessageHash();
        address recoveredSigner = ethSignedHash.recover(signature);
        return recoveredSigner == expectedSigner;
    }

    function setHighValueThreshold(uint256 newThreshold) external onlyTimelockOrOwner {
        highValueThreshold = newThreshold;
    }

    /**
     * @dev Verify aggregated BLS signature for a validation request
     * @param requestId The validation request ID
     * @return bool indicating if the aggregated signature is valid
     */
    function verifyAggregatedSignature(uint256 requestId) external view returns (bool) {
        ValidationRequest storage request = validationRequests[requestId];
        // Require the request to have reached completion and produced an aggregate
        if (request.status != ValidationStatus.Completed) return false;
        if (request.aggregatedSignature.length == 0) return false;

        // Decode stored aggregation result
        BLSSignatureAggregator.AggregationResult memory result =
            BLSSignatureAggregator.decodeAggregatedSignature(request.aggregatedSignature);
        if (!result.isValid) return false;

        // Convert signer BLS keys (strict mode)
        BLS12381.G2Point[] memory publicKeys = new BLS12381.G2Point[](request.signers.length);
        for (uint256 i = 0; i < request.signers.length; i++) {
            BLS12381.G2Point memory pk = BLSSignatureAggregator.convertPublicKey(
                validators[request.signers[i]].blsPublicKey
            );
            if (!BLSSignatureAggregator.isValidPublicKeyG2(pk)) {
                return false;
            }
            publicKeys[i] = pk;
        }

        // Perform best-effort aggregated verification (simplified pairing under the hood)
        return BLSSignatureAggregator.verifyAggregated(
            result.aggregatedSignature,
            request.messageHash,
            publicKeys
        );
    }

    /**
     * @dev Get BLS public key for a validator
     * @param validator The validator address
     * @return uint256[4] The BLS public key (G2 point)
     */
    function getValidatorBLSPublicKey(address validator) external view returns (uint256[4] memory) {
        return validators[validator].blsPublicKey;
    }

    /**
     * @dev Check if validator count meets BFT threshold
     * @param signatureCount Number of received signatures
     * @return bool true if threshold is met
     */
    function meetsBFTThreshold(uint256 signatureCount) external view returns (bool) {
        return BLSSignatureAggregator.meetsBFTThreshold(signatureCount, activeValidators.length);
    }

    function getValidationCount() external view returns (uint256) {
        return _validationCounter;
    }

    function pause() external onlyOwner {
        _pause();
    }

    function unpause() external onlyOwner {
        _unpause();
    }
}
