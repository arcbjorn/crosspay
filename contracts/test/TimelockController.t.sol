// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import {Test, console} from "forge-std/Test.sol";
import "../src/TimelockController.sol";

contract MockTarget {
    uint256 public value;
    
    function setValue(uint256 _value) external {
        value = _value;
    }
    
    function emergencyFunction() external {
        value = 999;
    }
}

contract TimelockControllerTest is Test {
    CrossPayTimelock public timelock;
    MockTarget public target;
    
    address public proposer = address(0x1);
    address public executor = address(0x2);
    address public admin = address(0x3);
    address public user = address(0x4);
    
    bytes32 public constant PROPOSER_ROLE = keccak256("PROPOSER_ROLE");
    bytes32 public constant EXECUTOR_ROLE = keccak256("EXECUTOR_ROLE");
    bytes32 public constant CANCELLER_ROLE = keccak256("CANCELLER_ROLE");
    
    function setUp() public {
        address[] memory proposers = new address[](1);
        address[] memory executors = new address[](1);
        proposers[0] = proposer;
        executors[0] = executor;
        
        timelock = new CrossPayTimelock(
            proposers,
            executors,
            admin
        );
        
        target = new MockTarget();
        
        vm.label(proposer, "Proposer");
        vm.label(executor, "Executor");
        vm.label(admin, "Admin");
        vm.label(user, "User");
    }
    
    function testInitialState() public {
        assertEq(timelock.getMinDelay(), 24 hours);
        assertTrue(timelock.hasRole(PROPOSER_ROLE, proposer));
        assertTrue(timelock.hasRole(EXECUTOR_ROLE, executor));
        assertTrue(timelock.hasRole(timelock.DEFAULT_ADMIN_ROLE(), admin));
    }
    
    function testScheduleOperation() public {
        bytes memory data = abi.encodeWithSignature("setValue(uint256)", 42);
        bytes32 salt = keccak256("test");
        uint256 delay = 24 hours;
        
        vm.prank(proposer);
        timelock.schedule(
            address(target),
            0,
            data,
            bytes32(0),
            salt,
            delay
        );
        
        bytes32 id = timelock.hashOperation(
            address(target),
            0,
            data,
            bytes32(0),
            salt
        );
        
        assertTrue(timelock.isOperationPending(id));
        assertEq(timelock.getTimestamp(id), block.timestamp + delay);
    }
    
    function testExecuteOperation() public {
        bytes memory data = abi.encodeWithSignature("setValue(uint256)", 42);
        bytes32 salt = keccak256("test");
        uint256 delay = 24 hours;
        
        // Schedule operation
        vm.prank(proposer);
        timelock.schedule(
            address(target),
            0,
            data,
            bytes32(0),
            salt,
            delay
        );
        
        // Fast forward time
        vm.warp(block.timestamp + delay + 1);
        
        // Execute operation
        vm.prank(executor);
        timelock.execute(
            address(target),
            0,
            data,
            bytes32(0),
            salt
        );
        
        assertEq(target.value(), 42);
    }
    
    function testEmergencySchedule() public {
        bytes memory data = abi.encodeWithSignature("emergencyFunction()");
        bytes32 salt = keccak256("emergency");
        uint256 emergencyDelay = 1 hours;
        
        vm.prank(proposer);
        timelock.scheduleEmergency(
            address(target),
            0,
            data,
            bytes32(0),
            salt
        );
        
        bytes32 id = timelock.hashOperation(
            address(target),
            0,
            data,
            bytes32(0),
            salt
        );
        
        assertTrue(timelock.isOperationPending(id));
        assertEq(timelock.getTimestamp(id), block.timestamp + emergencyDelay);
    }
    
    function testCannotExecuteBeforeDelay() public {
        bytes memory data = abi.encodeWithSignature("setValue(uint256)", 42);
        bytes32 salt = keccak256("test");
        
        vm.prank(proposer);
        timelock.schedule(
            address(target),
            0,
            data,
            bytes32(0),
            salt,
            24 hours
        );
        
        vm.expectRevert();
        vm.prank(executor);
        timelock.execute(
            address(target),
            0,
            data,
            bytes32(0),
            salt
        );
    }
    
    function testCannotScheduleWithoutRole() public {
        bytes memory data = abi.encodeWithSignature("setValue(uint256)", 42);
        bytes32 salt = keccak256("test");
        
        vm.expectRevert();
        vm.prank(user);
        timelock.schedule(
            address(target),
            0,
            data,
            bytes32(0),
            salt,
            24 hours
        );
    }
    
    function testCannotExecuteWithoutRole() public {
        bytes memory data = abi.encodeWithSignature("setValue(uint256)", 42);
        bytes32 salt = keccak256("test");
        
        vm.prank(proposer);
        timelock.schedule(
            address(target),
            0,
            data,
            bytes32(0),
            salt,
            24 hours
        );
        
        vm.warp(block.timestamp + 24 hours + 1);
        
        vm.expectRevert();
        vm.prank(user);
        timelock.execute(
            address(target),
            0,
            data,
            bytes32(0),
            salt
        );
    }
    
    function testCancelOperation() public {
        bytes memory data = abi.encodeWithSignature("setValue(uint256)", 42);
        bytes32 salt = keccak256("test");
        
        vm.prank(proposer);
        timelock.schedule(
            address(target),
            0,
            data,
            bytes32(0),
            salt,
            24 hours
        );
        
        bytes32 id = timelock.hashOperation(
            address(target),
            0,
            data,
            bytes32(0),
            salt
        );
        
        assertTrue(timelock.isOperationPending(id));
        
        vm.prank(proposer); // Proposer also has CANCELLER_ROLE
        timelock.cancel(id);
        
        assertFalse(timelock.isOperationPending(id));
    }
}
