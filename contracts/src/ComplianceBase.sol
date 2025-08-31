// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin/contracts/utils/Pausable.sol";

abstract contract ComplianceBase is Ownable, Pausable {
    mapping(address => bool) public kycVerified;
    mapping(address => bool) public blacklisted;
    mapping(bytes32 => bool) public jurisdictionBlocked;
    
    uint256 public amlThreshold = 10000e18; // $10,000 equivalent
    bool public kycRequired = false;

    event KYCStatusUpdated(address indexed user, bool verified);
    event BlacklistStatusUpdated(address indexed user, bool blacklisted);
    event JurisdictionBlocked(bytes32 indexed jurisdictionHash, bool blocked);
    event AMLThresholdUpdated(uint256 oldThreshold, uint256 newThreshold);
    event KYCRequirementUpdated(bool required);

    error UserBlacklisted();
    error KYCRequired();
    error JurisdictionBlockedError();
    error AMLThresholdExceeded();

    constructor() Ownable(msg.sender) {}

    modifier onlyCompliant(address user, uint256 amount, bytes32 jurisdictionHash) {
        _checkCompliance(user, amount, jurisdictionHash);
        _;
    }

    function _checkCompliance(
        address user,
        uint256 amount,
        bytes32 jurisdictionHash
    ) internal view {
        if (blacklisted[user]) {
            revert UserBlacklisted();
        }

        if (kycRequired && !kycVerified[user]) {
            revert KYCRequired();
        }

        if (jurisdictionBlocked[jurisdictionHash]) {
            revert JurisdictionBlockedError();
        }

        if (amount > amlThreshold) {
            if (!kycVerified[user]) {
                revert AMLThresholdExceeded();
            }
        }
    }

    function setKYCStatus(address user, bool verified) external onlyOwner {
        kycVerified[user] = verified;
        emit KYCStatusUpdated(user, verified);
    }

    function setBlacklistStatus(address user, bool _blacklisted) external onlyOwner {
        blacklisted[user] = _blacklisted;
        emit BlacklistStatusUpdated(user, _blacklisted);
    }

    function setJurisdictionBlocked(bytes32 jurisdictionHash, bool blocked) external onlyOwner {
        jurisdictionBlocked[jurisdictionHash] = blocked;
        emit JurisdictionBlocked(jurisdictionHash, blocked);
    }

    function setAMLThreshold(uint256 newThreshold) external onlyOwner {
        uint256 oldThreshold = amlThreshold;
        amlThreshold = newThreshold;
        emit AMLThresholdUpdated(oldThreshold, newThreshold);
    }

    function setKYCRequired(bool required) external onlyOwner {
        kycRequired = required;
        emit KYCRequirementUpdated(required);
    }

    function isCompliant(
        address user,
        uint256 amount,
        bytes32 jurisdictionHash
    ) external view returns (bool) {
        if (blacklisted[user]) return false;
        if (kycRequired && !kycVerified[user]) return false;
        if (jurisdictionBlocked[jurisdictionHash]) return false;
        if (amount > amlThreshold && !kycVerified[user]) return false;
        return true;
    }

    function getUserComplianceStatus(address user) external view returns (
        bool isKYCVerified,
        bool isBlacklisted,
        bool canTransact
    ) {
        isKYCVerified = kycVerified[user];
        isBlacklisted = blacklisted[user];
        canTransact = !isBlacklisted && (!kycRequired || isKYCVerified);
    }
}