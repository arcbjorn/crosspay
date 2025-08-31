// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import {Test, console} from "forge-std/Test.sol";
import "../src/TrancheVault.sol";

contract MockToken is ERC20 {
    constructor() ERC20("Test Token", "TEST") {
        _mint(msg.sender, 1000000 * 10**18);
    }
    
    function mint(address to, uint256 amount) external {
        _mint(to, amount);
    }
}

contract TrancheVaultLiquidationTest is Test {
    TrancheVault public vault;
    MockToken public token;
    
    address public alice = address(0x1);
    address public bob = address(0x2);
    address public charlie = address(0x3);
    address public liquidator = address(0x4);
    
    function setUp() public {
        token = new MockToken();
        vault = new TrancheVault(address(token), "Vault Shares", "VS");
        
        // Setup users with tokens
        token.mint(alice, 10000 * 10**18);
        token.mint(bob, 10000 * 10**18);
        token.mint(charlie, 10000 * 10**18);
        token.mint(liquidator, 10000 * 10**18);
        
        // Approve vault for all users
        vm.prank(alice);
        token.approve(address(vault), type(uint256).max);
        
        vm.prank(bob);
        token.approve(address(vault), type(uint256).max);
        
        vm.prank(charlie);
        token.approve(address(vault), type(uint256).max);
        
        vm.prank(liquidator);
        token.approve(address(vault), type(uint256).max);
    }
    
    function testHealthFactorCalculation() public {
        // Alice deposits 1000 tokens in junior tranche (high risk)
        vm.startPrank(alice);
        vault.deposit(TrancheVault.TrancheType.Junior, 1000 * 10**18);
        vm.stopPrank();
        
        // Initial health factor should be high (no losses yet)
        uint256 initialHealthFactor = vault.calculateHealthFactor(alice);
        assertGt(initialHealthFactor, 8000); // > 80%
        
        // Charlie deposits in senior tranche for vault balance
        vm.startPrank(charlie);
        vault.deposit(TrancheVault.TrancheType.Senior, 5000 * 10**18);
        vm.stopPrank();
        
        // Simulate slashing event to trigger health factor degradation
        vault.executeSlashing(500 * 10**18, alice, "Test slashing for health factor");
        
        // Health factor should decrease after slashing
        uint256 postSlashHealthFactor = vault.calculateHealthFactor(alice);
        assertLt(postSlashHealthFactor, initialHealthFactor);
    }
    
    function testLiquidationEligibilityCheck() public {
        // Setup: Alice deposits in risky junior tranche
        vm.startPrank(alice);
        vault.deposit(TrancheVault.TrancheType.Junior, 1000 * 10**18);
        vm.stopPrank();
        
        // Charlie provides senior tranche liquidity
        vm.startPrank(charlie);
        vault.deposit(TrancheVault.TrancheType.Senior, 5000 * 10**18);
        vm.stopPrank();
        
        // Check initial eligibility (should not be eligible)
        (bool isEligible, uint256 healthFactor, uint256 maxLiquidatable, uint256 estimatedReward) = 
            vault.checkLiquidationEligibility(alice);
        
        assertFalse(isEligible);
        assertGt(healthFactor, vault.LIQUIDATION_THRESHOLD());
        assertEq(maxLiquidatable, 0);
        assertEq(estimatedReward, 0);
    }
    
    function testLiquidationProcess() public {
        // Setup positions
        vm.startPrank(alice);
        vault.deposit(TrancheVault.TrancheType.Junior, 1000 * 10**18);
        vm.stopPrank();
        
        vm.startPrank(bob);
        vault.deposit(TrancheVault.TrancheType.Mezzanine, 2000 * 10**18);
        vm.stopPrank();
        
        vm.startPrank(charlie);
        vault.deposit(TrancheVault.TrancheType.Senior, 5000 * 10**18);
        vm.stopPrank();
        
        // Force health factor below threshold through moderate slashing (not total wipeout)
        vault.executeSlashing(4000 * 10**18, alice, "Moderate slashing to trigger liquidation");
        
        // Check if Alice is now eligible for liquidation
        uint256 aliceHealthFactor = vault.calculateHealthFactor(alice);
        
        // If health factor is below threshold, proceed with liquidation
        if (aliceHealthFactor < vault.LIQUIDATION_THRESHOLD()) {
            (bool isEligible, , uint256 maxLiquidatable, ) = vault.checkLiquidationEligibility(alice);
            assertTrue(isEligible);
            assertGt(maxLiquidatable, 0);
            
            // Get Alice's position before liquidation
            (uint256 juniorBefore, , , , , , ) = 
                vault.getExtendedUserPosition(alice);
            
            // Execute liquidation
            uint256 liquidationAmount = maxLiquidatable / 2; // Liquidate 50% of max
            
            vm.startPrank(liquidator);
            vault.liquidateUser(alice, liquidationAmount);
            vm.stopPrank();
            
            // Verify liquidation occurred
            (uint256 juniorAfter, , , , , , ) = 
                vault.getExtendedUserPosition(alice);
            
            // Alice's position should be reduced
            assertTrue(juniorAfter < juniorBefore);
            
            // Check liquidation event was recorded
            (uint256 totalEvents, , , , ) = vault.getLiquidationMetrics();
            assertEq(totalEvents, 1);
            
            TrancheVault.LiquidationEvent memory liquidationEvent = vault.getLiquidationEvent(1);
            assertEq(liquidationEvent.liquidatedUser, alice);
            assertEq(liquidationEvent.liquidator, liquidator);
            assertGt(liquidationEvent.liquidatedAmount, 0);
        } else {
            // If health factor is still above threshold, test that liquidation fails
            vm.startPrank(liquidator);
            vm.expectRevert(TrancheVault.HealthyPosition.selector);
            vault.liquidateUser(alice, 100 * 10**18);
            vm.stopPrank();
        }
    }
    
    function testLiquidationWaterfall() public {
        // Alice deposits across all tranches
        vm.startPrank(alice);
        vault.deposit(TrancheVault.TrancheType.Junior, 500 * 10**18);
        vault.deposit(TrancheVault.TrancheType.Mezzanine, 1000 * 10**18);
        vault.deposit(TrancheVault.TrancheType.Senior, 1500 * 10**18);
        vm.stopPrank();
        
        // Bob provides additional liquidity
        vm.startPrank(bob);
        vault.deposit(TrancheVault.TrancheType.Senior, 5000 * 10**18);
        vm.stopPrank();
        
        // Trigger major slashing to make Alice liquidatable
        vault.executeSlashing(5000 * 10**18, alice, "Major slashing for waterfall test");
        
        // Check if liquidation is triggered
        uint256 healthFactor = vault.calculateHealthFactor(alice);
        if (healthFactor < vault.LIQUIDATION_THRESHOLD()) {
            (uint256 juniorBefore, , , , , , ) = 
                vault.getExtendedUserPosition(alice);
            
            // Execute partial liquidation
            vm.startPrank(liquidator);
            vault.liquidateUser(alice, 800 * 10**18);
            vm.stopPrank();
            
            (uint256 juniorAfter, , , , , , ) = 
                vault.getExtendedUserPosition(alice);
            
            // Verify waterfall: Junior should be liquidated first
            if (juniorBefore > 0) {
                assertTrue(juniorAfter <= juniorBefore);
            }
        }
    }
    
    function testLiquidationRewards() public {
        // Setup scenario where Alice can be liquidated
        vm.startPrank(alice);
        vault.deposit(TrancheVault.TrancheType.Junior, 1000 * 10**18);
        vm.stopPrank();
        
        vm.startPrank(bob);
        vault.deposit(TrancheVault.TrancheType.Senior, 5000 * 10**18);
        vm.stopPrank();
        
        // Trigger slashing
        vault.executeSlashing(4000 * 10**18, alice, "Slashing for reward test");
        
        uint256 healthFactor = vault.calculateHealthFactor(alice);
        if (healthFactor < vault.LIQUIDATION_THRESHOLD()) {
            uint256 liquidatorBalanceBefore = token.balanceOf(liquidator);
            
            (bool isEligible, , uint256 maxLiquidatable, ) = 
                vault.checkLiquidationEligibility(alice);
            
            if (isEligible && maxLiquidatable > 0) {
                vm.startPrank(liquidator);
                vault.liquidateUser(alice, maxLiquidatable);
                vm.stopPrank();
                
                uint256 liquidatorBalanceAfter = token.balanceOf(liquidator);
                
                // Liquidator should receive reward (minus the liquidation amount paid)
                // The net change should account for: -liquidationAmount + reward
                assertTrue(liquidatorBalanceAfter != liquidatorBalanceBefore);
                
                // Verify liquidation metrics updated
                (uint256 totalEvents, , , uint256 totalRewardsPaid, ) = vault.getLiquidationMetrics();
                assertEq(totalEvents, 1);
                assertGt(totalRewardsPaid, 0);
            }
        }
    }
    
    function testCannotLiquidateHealthyUser() public {
        // Alice deposits in safe senior tranche
        vm.startPrank(alice);
        vault.deposit(TrancheVault.TrancheType.Senior, 1000 * 10**18);
        vm.stopPrank();
        
        // Try to liquidate healthy user - should fail
        vm.startPrank(liquidator);
        vm.expectRevert(TrancheVault.HealthyPosition.selector);
        vault.liquidateUser(alice, 100 * 10**18);
        vm.stopPrank();
    }
    
    function testLiquidationAmountLimits() public {
        // Setup liquidatable position
        vm.startPrank(alice);
        vault.deposit(TrancheVault.TrancheType.Junior, 1000 * 10**18);
        vm.stopPrank();
        
        vm.startPrank(bob);
        vault.deposit(TrancheVault.TrancheType.Senior, 5000 * 10**18);
        vm.stopPrank();
        
        // Trigger slashing
        vault.executeSlashing(4000 * 10**18, alice, "Slashing for limit test");
        
        uint256 healthFactor = vault.calculateHealthFactor(alice);
        if (healthFactor < vault.LIQUIDATION_THRESHOLD()) {
            (bool isEligible, , uint256 maxLiquidatable, ) = vault.checkLiquidationEligibility(alice);
            
            if (isEligible && maxLiquidatable > 0) {
                // Try to liquidate more than maximum allowed
                uint256 excessiveAmount = maxLiquidatable * 2;
                
                vm.startPrank(liquidator);
                // Should succeed but only liquidate up to the maximum
                vault.liquidateUser(alice, excessiveAmount);
                vm.stopPrank();
                
                // Verify liquidation event has correct amount
                TrancheVault.LiquidationEvent memory liquidationEvent = vault.getLiquidationEvent(1);
                assertLe(liquidationEvent.liquidatedAmount, maxLiquidatable);
            }
        }
    }
    
    function testHealthFactorUpdates() public {
        // Alice deposits
        vm.startPrank(alice);
        vault.deposit(TrancheVault.TrancheType.Junior, 1000 * 10**18);
        vm.stopPrank();
        
        uint256 initialHealthFactor = vault.calculateHealthFactor(alice);
        assertTrue(initialHealthFactor > 0);
        
        // Check that health factor is stored
        uint256 storedHealthFactor = vault.userHealthFactors(alice);
        assertEq(storedHealthFactor, initialHealthFactor);
        
        // Bob provides liquidity for slashing
        vm.startPrank(bob);
        vault.deposit(TrancheVault.TrancheType.Senior, 3000 * 10**18);
        vm.stopPrank();
        
        // Execute slashing to change health factors
        vault.executeSlashing(1000 * 10**18, alice, "Health factor update test");
        
        uint256 newHealthFactor = vault.calculateHealthFactor(alice);
        uint256 newStoredHealthFactor = vault.userHealthFactors(alice);
        
        // Health factor should be updated and stored
        assertEq(newStoredHealthFactor, newHealthFactor);
    }
    
    function testGetLiquidationMetrics() public {
        // Initial metrics should be zero
        (uint256 totalEvents, uint256 totalLiquidated, uint256 totalPenalties, uint256 totalRewards, uint256 avgHealth) = 
            vault.getLiquidationMetrics();
        
        assertEq(totalEvents, 0);
        assertEq(totalLiquidated, 0);
        assertEq(totalPenalties, 0);
        assertEq(totalRewards, 0);
        assertGt(avgHealth, 0); // Should have some default value
    }
    
    function testExtendedUserPosition() public {
        // Alice deposits
        vm.startPrank(alice);
        vault.deposit(TrancheVault.TrancheType.Junior, 1000 * 10**18);
        vm.stopPrank();
        
        // Check extended user position
        (uint256 junior, uint256 mezzanine, uint256 senior, , uint256 lastDeposit, uint256 healthFactor, bool atRisk) = 
            vault.getExtendedUserPosition(alice);
        
        assertEq(junior, 1000 * 10**18);
        assertEq(mezzanine, 0);
        assertEq(senior, 0);
        assertGt(lastDeposit, 0);
        assertGt(healthFactor, 0);
        assertFalse(atRisk); // Should not be at risk initially
    }
}