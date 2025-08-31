// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import "@openzeppelin/contracts/governance/TimelockController.sol";

contract CrossPayTimelock is TimelockController {
    uint256 public constant MIN_DELAY = 24 hours;
    uint256 public constant EMERGENCY_DELAY = 1 hours;
    
    mapping(bytes32 => bool) public emergencyOperations;
    mapping(bytes32 => uint256) public emergencyTimestamps;

    event EmergencyOperationScheduled(bytes32 indexed id, bytes32 indexed operation);
    
    error NotEmergencyOperation();
    error InsufficientDelay();

    constructor(
        address[] memory proposers,
        address[] memory executors,
        address admin
    ) TimelockController(MIN_DELAY, proposers, executors, admin) {
        _grantRole(PROPOSER_ROLE, admin);
        _grantRole(EXECUTOR_ROLE, admin);
        _grantRole(CANCELLER_ROLE, admin);
    }

    // Override schedule to allow emergency delays
    function schedule(
        address target,
        uint256 value,
        bytes calldata data,
        bytes32 predecessor,
        bytes32 salt,
        uint256 delay
    ) public virtual override onlyRole(PROPOSER_ROLE) {
        bytes32 operationId = hashOperation(target, value, data, predecessor, salt);
        
        // Allow emergency delay for emergency operations
        if (emergencyOperations[operationId] || isEmergencyOperation(data)) {
            require(delay >= EMERGENCY_DELAY, "Delay too short for emergency");
            // For emergency operations, bypass parent delay checks
            _scheduleWithDelay(operationId, target, value, data, predecessor, delay);
        } else {
            require(delay >= getMinDelay(), "Delay too short");
            super.schedule(target, value, data, predecessor, salt, delay);
        }
    }

    function _scheduleWithDelay(
        bytes32 id,
        address target,
        uint256 value,
        bytes calldata data,
        bytes32 predecessor,
        uint256 delay
    ) private {
        require(emergencyTimestamps[id] == 0, "Emergency operation already exists");
        uint256 timestamp = block.timestamp + delay;
        
        // Store timestamp for emergency operations
        emergencyTimestamps[id] = timestamp;
        
        // Manually emit the event that the parent would emit
        emit CallScheduled(id, 0, target, value, data, predecessor, timestamp);
    }

    // Override getTimestamp to return emergency timestamps when applicable
    function getTimestamp(bytes32 id) public view override returns (uint256) {
        if (emergencyOperations[id] && emergencyTimestamps[id] > 0) {
            return emergencyTimestamps[id];
        }
        return super.getTimestamp(id);
    }

    // Check if emergency operation is pending
    function isEmergencyOperationPending(bytes32 id) public view returns (bool) {
        return emergencyOperations[id] && emergencyTimestamps[id] > block.timestamp;
    }

    // Check if emergency operation is ready
    function isEmergencyOperationReady(bytes32 id) public view returns (bool) {
        return emergencyOperations[id] && emergencyTimestamps[id] > 0 && emergencyTimestamps[id] <= block.timestamp;
    }


    function scheduleEmergency(
        address target,
        uint256 value,
        bytes calldata data,
        bytes32 predecessor,
        bytes32 salt
    ) public onlyRole(PROPOSER_ROLE) returns (bytes32) {
        require(isEmergencyOperation(data), "Not emergency operation");
        
        bytes32 operationId = hashOperation(target, value, data, predecessor, salt);
        
        // Set emergency flag BEFORE calling schedule
        emergencyOperations[operationId] = true;
        
        // Now call schedule with emergency delay
        schedule(target, value, data, predecessor, salt, EMERGENCY_DELAY);
        
        emit EmergencyOperationScheduled(operationId, keccak256(data));
        
        return operationId;
    }

    function executeEmergency(
        address target,
        uint256 value,
        bytes calldata data,
        bytes32 predecessor,
        bytes32 salt
    ) public payable onlyRole(EXECUTOR_ROLE) {
        bytes32 operationId = hashOperation(target, value, data, predecessor, salt);
        
        if (!emergencyOperations[operationId]) {
            revert NotEmergencyOperation();
        }
        
        execute(target, value, data, predecessor, salt);
        
        delete emergencyOperations[operationId];
    }

    function isEmergencyOperation(bytes calldata data) public pure returns (bool) {
        bytes4 selector = bytes4(data[:4]);
        
        return selector == bytes4(keccak256("emergencyPause()")) ||
               selector == bytes4(keccak256("emergencyUnpause()")) ||
               selector == bytes4(keccak256("slashValidator(address,string)")) ||
               selector == bytes4(keccak256("emergencyCancel(uint256)"));
    }
}