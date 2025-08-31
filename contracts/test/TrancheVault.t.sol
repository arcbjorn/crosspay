// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import {Test, console} from "forge-std/Test.sol";
import "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import "../src/TrancheVault.sol";

contract MockToken is ERC20 {
    constructor() ERC20("Mock Token", "MOCK") {
        _mint(msg.sender, 1000000 * 10**18);
    }
    
    function mint(address to, uint256 amount) external {
        _mint(to, amount);
    }
}

contract TrancheVaultTest is Test {
    TrancheVault public vault;
    MockToken public token;
    address public alice = address(0x1);
    address public bob = address(0x2);
    address public charlie = address(0x3);
    
    function setUp() public {
        token = new MockToken();
        vault = new TrancheVault(address(token), "Tranche Vault Shares", "TVS");
        
        token.mint(alice, 1000 * 10**18);
        token.mint(bob, 1000 * 10**18);
        token.mint(charlie, 1000 * 10**18);
    }

    function testDepositJuniorTranche() public {
        uint256 depositAmount = 100 * 10**18;
        
        vm.startPrank(alice);
        token.approve(address(vault), depositAmount);
        
        vault.deposit(TrancheVault.TrancheType.Junior, depositAmount);
        
        (uint256 juniorDeposit,,,, uint256 lastDeposit) = vault.getUserPosition(alice);
        assertEq(juniorDeposit, depositAmount);
        assertEq(lastDeposit, block.timestamp);
        
        vm.stopPrank();
    }

    function testDepositAllTranches() public {
        uint256 depositAmount = 100 * 10**18;
        
        vm.startPrank(alice);
        token.approve(address(vault), depositAmount * 3);
        
        vault.deposit(TrancheVault.TrancheType.Junior, depositAmount);
        vault.deposit(TrancheVault.TrancheType.Mezzanine, depositAmount);
        vault.deposit(TrancheVault.TrancheType.Senior, depositAmount);
        
        (uint256 junior, uint256 mezzanine, uint256 senior,,) = vault.getUserPosition(alice);
        assertEq(junior, depositAmount);
        assertEq(mezzanine, depositAmount);
        assertEq(senior, depositAmount);
        
        vm.stopPrank();
    }

    function testInsufficientDeposit() public {
        vm.startPrank(alice);
        token.approve(address(vault), 1000);
        
        vm.expectRevert(TrancheVault.InsufficientDeposit.selector);
        vault.deposit(TrancheVault.TrancheType.Junior, 1000);
        
        vm.stopPrank();
    }

    function testWithdrawalRequest() public {
        uint256 depositAmount = 100 * 10**18;
        
        vm.startPrank(alice);
        token.approve(address(vault), depositAmount);
        vault.deposit(TrancheVault.TrancheType.Junior, depositAmount);
        
        vault.requestWithdrawal(TrancheVault.TrancheType.Junior, 50 * 10**18);
        
        vm.stopPrank();
    }

    function testWithdrawalAfterDelay() public {
        uint256 depositAmount = 100 * 10**18;
        
        vm.startPrank(alice);
        token.approve(address(vault), depositAmount);
        vault.deposit(TrancheVault.TrancheType.Junior, depositAmount);
        
        vault.requestWithdrawal(TrancheVault.TrancheType.Junior, 50 * 10**18);
        
        vm.warp(block.timestamp + 8 days);
        
        vault.withdraw(TrancheVault.TrancheType.Junior);
        
        (uint256 juniorDeposit,,,, ) = vault.getUserPosition(alice);
        assertEq(juniorDeposit, 50 * 10**18);
        
        vm.stopPrank();
    }

    function testWithdrawalBeforeDelay() public {
        uint256 depositAmount = 100 * 10**18;
        
        vm.startPrank(alice);
        token.approve(address(vault), depositAmount);
        vault.deposit(TrancheVault.TrancheType.Junior, depositAmount);
        
        vault.requestWithdrawal(TrancheVault.TrancheType.Junior, 50 * 10**18);
        
        vm.expectRevert(TrancheVault.WithdrawalDelayNotMet.selector);
        vault.withdraw(TrancheVault.TrancheType.Junior);
        
        vm.stopPrank();
    }

    function testSlashingWaterfall() public {
        uint256 depositAmount = 100 * 10**18;
        
        vm.startPrank(alice);
        token.approve(address(vault), depositAmount);
        vault.deposit(TrancheVault.TrancheType.Junior, depositAmount);
        vm.stopPrank();
        
        vm.startPrank(bob);
        token.approve(address(vault), depositAmount);
        vault.deposit(TrancheVault.TrancheType.Mezzanine, depositAmount);
        vm.stopPrank();
        
        vm.startPrank(charlie);
        token.approve(address(vault), depositAmount);
        vault.deposit(TrancheVault.TrancheType.Senior, depositAmount);
        vm.stopPrank();
        
        uint256 slashAmount = 150 * 10**18;
        vault.executeSlashing(slashAmount, address(0x999), "Test slashing");
        
        TrancheVault.SlashingEvent memory slashEvent = vault.getSlashingEvent(1);
        assertEq(slashEvent.amount, slashAmount);
        assertEq(slashEvent.juniorSlashed, depositAmount); // Junior takes first 100
        assertEq(slashEvent.mezzanineSlashed, 50 * 10**18); // Mezzanine takes remaining 50
        assertEq(slashEvent.seniorSlashed, 0); // Senior protected
    }

    function testYieldCalculation() public {
        uint256 depositAmount = 100 * 10**18;
        
        vm.startPrank(alice);
        token.approve(address(vault), depositAmount);
        vault.deposit(TrancheVault.TrancheType.Junior, depositAmount);
        vm.stopPrank();
        
        vm.warp(block.timestamp + 365 days);
        
        uint256 userYield = vault.calculateUserYield(alice);
        assertTrue(userYield > 0);
        assertTrue(userYield > depositAmount / 10); // Should be more than 10% of deposit
    }

    function testTrancheAPY() public {
        uint256 juniorAPY = vault.getTrancheAPY(TrancheVault.TrancheType.Junior);
        uint256 mezzanineAPY = vault.getTrancheAPY(TrancheVault.TrancheType.Mezzanine);
        uint256 seniorAPY = vault.getTrancheAPY(TrancheVault.TrancheType.Senior);
        
        assertTrue(juniorAPY > mezzanineAPY);
        assertTrue(mezzanineAPY > seniorAPY);
        assertEq(juniorAPY, 1200); // 12%
        assertEq(mezzanineAPY, 800); // 8%
        assertEq(seniorAPY, 500); // 5%
    }

    function testVaultMetrics() public {
        uint256 depositAmount = 100 * 10**18;
        
        vm.startPrank(alice);
        token.approve(address(vault), depositAmount * 3);
        vault.deposit(TrancheVault.TrancheType.Junior, depositAmount);
        vault.deposit(TrancheVault.TrancheType.Mezzanine, depositAmount);
        vault.deposit(TrancheVault.TrancheType.Senior, depositAmount);
        vm.stopPrank();
        
        (
            uint256 totalAssets,
            uint256 juniorTVL,
            uint256 mezzanineTVL,
            uint256 seniorTVL,
            , // insuranceBalance (unused)
            uint256 totalSlashingEvents
        ) = vault.getVaultMetrics();
        
        assertEq(totalAssets, depositAmount * 3);
        assertEq(juniorTVL, depositAmount);
        assertEq(mezzanineTVL, depositAmount);
        assertEq(seniorTVL, depositAmount);
        assertEq(totalSlashingEvents, 0);
    }

    function testRebalancing() public {
        uint256 depositAmount = 100 * 10**18;
        
        vm.startPrank(alice);
        token.approve(address(vault), depositAmount * 3);
        vault.deposit(TrancheVault.TrancheType.Junior, depositAmount);
        vault.deposit(TrancheVault.TrancheType.Mezzanine, depositAmount);
        vault.deposit(TrancheVault.TrancheType.Senior, depositAmount);
        vm.stopPrank();
        
        vault.rebalanceTranches();
        
        uint256 juniorUtil = vault.getTrancheUtilization(TrancheVault.TrancheType.Junior);
        uint256 mezzanineUtil = vault.getTrancheUtilization(TrancheVault.TrancheType.Mezzanine);
        uint256 seniorUtil = vault.getTrancheUtilization(TrancheVault.TrancheType.Senior);
        
        assertEq(juniorUtil, 2000); // 20%
        assertEq(mezzanineUtil, 3000); // 30%
        assertEq(seniorUtil, 5000); // 50%
    }

    function testEmergencyPause() public {
        vault.emergencyPause();
        assertTrue(vault.paused());
        
        vm.startPrank(alice);
        token.approve(address(vault), 100 * 10**18);
        
        vm.expectRevert();
        vault.deposit(TrancheVault.TrancheType.Junior, 100 * 10**18);
        
        vm.stopPrank();
        
        vault.emergencyUnpause();
        assertFalse(vault.paused());
    }

    function testPerformanceFees() public {
        uint256 depositAmount = 100 * 10**18;
        
        vm.startPrank(alice);
        token.approve(address(vault), depositAmount);
        vault.deposit(TrancheVault.TrancheType.Junior, depositAmount);
        vm.stopPrank();
        
        vault.distributeYield();
        
        uint256 ownerBalance = token.balanceOf(address(this));
        vault.withdrawPerformanceFees();
        
        assertTrue(token.balanceOf(address(this)) >= ownerBalance);
    }

    function testUnauthorizedActions() public {
        vm.startPrank(alice);
        
        vm.expectRevert();
        vault.executeSlashing(1000, address(0x999), "Unauthorized");
        
        vm.expectRevert();
        vault.setYieldRates(1000, 800, 600);
        
        vm.expectRevert();
        vault.emergencyPause();
        
        vm.stopPrank();
    }
}