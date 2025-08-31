// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin/contracts/utils/ReentrancyGuard.sol";

contract SubnameRegistry is Ownable, ReentrancyGuard {
    
    struct SubnameRecord {
        string subname;
        string domain;
        address owner;
        address resolver;
        uint256 createdAt;
        uint256 expiresAt;
        bool active;
        mapping(string => string) textRecords;
    }
    
    struct DomainInfo {
        address owner;
        bool delegationEnabled;
        uint256 registrationFee;
        uint256 renewalFee;
        uint256 maxSubnames;
        uint256 activeSubnames;
    }
    
    // Storage
    mapping(bytes32 => SubnameRecord) private subnameRecords;
    mapping(string => DomainInfo) private domainInfo;
    mapping(string => bytes32[]) private domainSubnames;
    mapping(address => bytes32[]) private userSubnames;
    
    // Events
    event SubnameRegistered(
        bytes32 indexed subnameHash,
        string subname,
        string domain,
        address indexed owner,
        address indexed resolver,
        uint256 expiresAt
    );
    
    event SubnameRenewed(
        bytes32 indexed subnameHash,
        uint256 newExpiresAt
    );
    
    event SubnameRevoked(
        bytes32 indexed subnameHash,
        string subname
    );
    
    event TextRecordSet(
        bytes32 indexed subnameHash,
        string key,
        string value
    );
    
    event DomainDelegated(
        string domain,
        address indexed owner,
        bool delegationEnabled
    );
    
    event BulkRegistration(
        string domain,
        address indexed owner,
        uint256 count
    );
    
    // Errors
    error SubnameAlreadyExists();
    error SubnameNotFound();
    error UnauthorizedAccess();
    error DomainNotDelegated();
    error SubnameLimitExceeded();
    error InsufficientPayment();
    error SubnameExpired();
    error InvalidSubname();
    
    constructor() Ownable(msg.sender) {}
    
    modifier onlySubnameOwner(bytes32 subnameHash) {
        if (subnameRecords[subnameHash].owner != msg.sender) {
            revert UnauthorizedAccess();
        }
        _;
    }
    
    modifier onlyDomainOwner(string memory domain) {
        if (domainInfo[domain].owner != msg.sender && msg.sender != owner()) {
            revert UnauthorizedAccess();
        }
        _;
    }
    
    function delegateDomain(
        string calldata domain,
        address domainOwner,
        uint256 registrationFee,
        uint256 renewalFee,
        uint256 maxSubnames
    ) external onlyOwner {
        DomainInfo storage info = domainInfo[domain];
        info.owner = domainOwner;
        info.delegationEnabled = true;
        info.registrationFee = registrationFee;
        info.renewalFee = renewalFee;
        info.maxSubnames = maxSubnames;
        
        emit DomainDelegated(domain, domainOwner, true);
    }
    
    function registerSubname(
        string calldata subname,
        string calldata domain,
        address resolver,
        uint256 duration
    ) external payable nonReentrant returns (bytes32) {
        DomainInfo storage info = domainInfo[domain];
        
        if (!info.delegationEnabled) {
            revert DomainNotDelegated();
        }
        
        if (info.activeSubnames >= info.maxSubnames && info.maxSubnames > 0) {
            revert SubnameLimitExceeded();
        }
        
        if (msg.value < info.registrationFee) {
            revert InsufficientPayment();
        }
        
        bytes32 subnameHash = keccak256(abi.encodePacked(subname, domain));
        
        if (subnameRecords[subnameHash].active) {
            revert SubnameAlreadyExists();
        }
        
        if (!_isValidSubname(subname)) {
            revert InvalidSubname();
        }
        
        uint256 expiresAt = block.timestamp + duration;
        
        SubnameRecord storage record = subnameRecords[subnameHash];
        record.subname = subname;
        record.domain = domain;
        record.owner = msg.sender;
        record.resolver = resolver;
        record.createdAt = block.timestamp;
        record.expiresAt = expiresAt;
        record.active = true;
        
        domainSubnames[domain].push(subnameHash);
        userSubnames[msg.sender].push(subnameHash);
        info.activeSubnames++;
        
        // Transfer fee to domain owner
        if (info.registrationFee > 0) {
            (bool success, ) = info.owner.call{value: info.registrationFee}("");
            require(success, "Fee transfer failed");
        }
        
        // Refund excess
        if (msg.value > info.registrationFee) {
            (bool success, ) = msg.sender.call{value: msg.value - info.registrationFee}("");
            require(success, "Refund failed");
        }
        
        emit SubnameRegistered(
            subnameHash,
            subname,
            domain,
            msg.sender,
            resolver,
            expiresAt
        );
        
        return subnameHash;
    }
    
    function bulkRegisterSubnames(
        string[] calldata subnames,
        string calldata domain,
        address resolver,
        uint256 duration
    ) external payable nonReentrant returns (bytes32[] memory) {
        DomainInfo storage info = domainInfo[domain];
        
        if (!info.delegationEnabled) {
            revert DomainNotDelegated();
        }
        
        uint256 count = subnames.length;
        if (count == 0 || count > 100) {
            revert InvalidSubname();
        }
        
        if (info.activeSubnames + count > info.maxSubnames && info.maxSubnames > 0) {
            revert SubnameLimitExceeded();
        }
        
        uint256 totalFee = info.registrationFee * count;
        if (msg.value < totalFee) {
            revert InsufficientPayment();
        }
        
        bytes32[] memory hashes = new bytes32[](count);
        uint256 expiresAt = block.timestamp + duration;
        uint256 successfulRegistrations = 0;
        
        for (uint256 i = 0; i < count; i++) {
            if (!_isValidSubname(subnames[i])) {
                continue;
            }
            
            bytes32 subnameHash = keccak256(abi.encodePacked(subnames[i], domain));
            
            if (subnameRecords[subnameHash].active) {
                continue;
            }
            
            SubnameRecord storage record = subnameRecords[subnameHash];
            record.subname = subnames[i];
            record.domain = domain;
            record.owner = msg.sender;
            record.resolver = resolver;
            record.createdAt = block.timestamp;
            record.expiresAt = expiresAt;
            record.active = true;
            
            domainSubnames[domain].push(subnameHash);
            userSubnames[msg.sender].push(subnameHash);
            hashes[i] = subnameHash;
            successfulRegistrations++;
            
            emit SubnameRegistered(
                subnameHash,
                subnames[i],
                domain,
                msg.sender,
                resolver,
                expiresAt
            );
        }
        
        info.activeSubnames += successfulRegistrations;
        
        // Transfer fees
        uint256 actualFee = info.registrationFee * successfulRegistrations;
        if (actualFee > 0) {
            (bool success, ) = info.owner.call{value: actualFee}("");
            require(success, "Fee transfer failed");
        }
        
        // Refund excess
        if (msg.value > actualFee) {
            (bool success, ) = msg.sender.call{value: msg.value - actualFee}("");
            require(success, "Refund failed");
        }
        
        emit BulkRegistration(domain, msg.sender, successfulRegistrations);
        
        return hashes;
    }
    
    function renewSubname(
        bytes32 subnameHash,
        uint256 duration
    ) external payable onlySubnameOwner(subnameHash) {
        SubnameRecord storage record = subnameRecords[subnameHash];
        
        if (!record.active) {
            revert SubnameNotFound();
        }
        
        DomainInfo storage info = domainInfo[record.domain];
        
        if (msg.value < info.renewalFee) {
            revert InsufficientPayment();
        }
        
        record.expiresAt = block.timestamp + duration;
        
        // Transfer fee
        if (info.renewalFee > 0) {
            (bool success, ) = info.owner.call{value: info.renewalFee}("");
            require(success, "Fee transfer failed");
        }
        
        // Refund excess
        if (msg.value > info.renewalFee) {
            (bool success, ) = msg.sender.call{value: msg.value - info.renewalFee}("");
            require(success, "Refund failed");
        }
        
        emit SubnameRenewed(subnameHash, record.expiresAt);
    }
    
    function revokeSubname(
        bytes32 subnameHash
    ) external onlySubnameOwner(subnameHash) {
        SubnameRecord storage record = subnameRecords[subnameHash];
        
        if (!record.active) {
            revert SubnameNotFound();
        }
        
        record.active = false;
        domainInfo[record.domain].activeSubnames--;
        
        emit SubnameRevoked(subnameHash, record.subname);
    }
    
    function setTextRecord(
        bytes32 subnameHash,
        string calldata key,
        string calldata value
    ) external onlySubnameOwner(subnameHash) {
        SubnameRecord storage record = subnameRecords[subnameHash];
        
        if (!record.active || block.timestamp > record.expiresAt) {
            revert SubnameExpired();
        }
        
        record.textRecords[key] = value;
        
        emit TextRecordSet(subnameHash, key, value);
    }
    
    function getSubnameRecord(bytes32 subnameHash) external view returns (
        string memory subname,
        string memory domain,
        address owner,
        address resolver,
        uint256 createdAt,
        uint256 expiresAt,
        bool active
    ) {
        SubnameRecord storage record = subnameRecords[subnameHash];
        return (
            record.subname,
            record.domain,
            record.owner,
            record.resolver,
            record.createdAt,
            record.expiresAt,
            record.active
        );
    }
    
    function getTextRecord(
        bytes32 subnameHash,
        string calldata key
    ) external view returns (string memory) {
        return subnameRecords[subnameHash].textRecords[key];
    }
    
    function getDomainSubnames(
        string calldata domain
    ) external view returns (bytes32[] memory) {
        return domainSubnames[domain];
    }
    
    function getUserSubnames(
        address user
    ) external view returns (bytes32[] memory) {
        return userSubnames[user];
    }
    
    function getDomainInfo(string calldata domain) external view returns (
        address owner,
        bool delegationEnabled,
        uint256 registrationFee,
        uint256 renewalFee,
        uint256 maxSubnames,
        uint256 activeSubnames
    ) {
        DomainInfo storage info = domainInfo[domain];
        return (
            info.owner,
            info.delegationEnabled,
            info.registrationFee,
            info.renewalFee,
            info.maxSubnames,
            info.activeSubnames
        );
    }
    
    function _isValidSubname(string memory subname) internal pure returns (bool) {
        bytes memory subnameBytes = bytes(subname);
        
        if (subnameBytes.length == 0 || subnameBytes.length > 63) {
            return false;
        }
        
        // Check for valid characters (simplified)
        for (uint256 i = 0; i < subnameBytes.length; i++) {
            bytes1 char = subnameBytes[i];
            if (!(
                (char >= 0x30 && char <= 0x39) || // 0-9
                (char >= 0x61 && char <= 0x7A) || // a-z
                char == 0x2D // hyphen
            )) {
                return false;
            }
        }
        
        // Cannot start or end with hyphen
        if (subnameBytes[0] == 0x2D || subnameBytes[subnameBytes.length - 1] == 0x2D) {
            return false;
        }
        
        return true;
    }
}