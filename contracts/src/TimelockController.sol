// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import "@openzeppelin/contracts/governance/TimelockController.sol";

contract CrossPayTimelock is TimelockController {
    uint256 public constant MIN_DELAY = 24 hours;
    uint256 public constant EMERGENCY_DELAY = 1 hours;
    
    mapping(bytes32 => bool) public emergencyOperations;

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

    function scheduleEmergency(
        address target,
        uint256 value,
        bytes calldata data,
        bytes32 predecessor,
        bytes32 salt
    ) public onlyRole(PROPOSER_ROLE) returns (bytes32) {
        require(isEmergencyOperation(data), "Not emergency operation");
        
        bytes32 operationId = hashOperation(target, value, data, predecessor, salt);
        emergencyOperations[operationId] = true;
        
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