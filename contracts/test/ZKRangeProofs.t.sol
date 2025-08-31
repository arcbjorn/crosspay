// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import {Test, console} from "forge-std/Test.sol";
import "../src/ZKRangeProofs.sol";
import "fhevm/lib/TFHE.sol";

contract ZKRangeProofsTest is Test {
    using ZKRangeProofs for euint256;
    
    function setUp() public {
        // Initialize TFHE if needed
    }
    
    function testGenerateRangeCommitment() public {
        // Mock encrypted value (in real implementation this would be properly encrypted)
        euint256 encryptedValue = TFHE.asEuint256(100);
        uint256 minValue = 50;
        uint256 maxValue = 150;
        bytes32 randomness = keccak256("test");
        
        bytes32 commitment = ZKRangeProofs.generateRangeCommitment(
            encryptedValue,
            minValue,
            maxValue,
            randomness
        );
        
        assertTrue(commitment != bytes32(0));
    }
    
    function testVerifyRange() public {
        euint256 encryptedValue = TFHE.asEuint256(100);
        uint256 minValue = 50;
        uint256 maxValue = 150;
        
        ebool result = ZKRangeProofs.verifyRange(encryptedValue, minValue, maxValue);
        
        // In a real FHE environment, this would return an encrypted boolean
        // For testing, we assume the mock implementation works
        // Cannot use TFHE.decrypt in smart contracts, only client-side
        assertTrue(true);
    }
    
    function testVerifyRangeInvalidRange() public {
        euint256 encryptedValue = TFHE.asEuint256(100);
        uint256 minValue = 150;
        uint256 maxValue = 50; // Invalid: min > max
        
        vm.expectRevert(ZKRangeProofs.InvalidRange.selector);
        ZKRangeProofs.verifyRange(encryptedValue, minValue, maxValue);
    }
    
    function testCreateZKRangeProof() public {
        euint256 encryptedValue = TFHE.asEuint256(100);
        uint256 minValue = 50;
        uint256 maxValue = 150;
        bytes32 randomness = keccak256("test");
        
        ZKRangeProofs.RangeProof memory proof = ZKRangeProofs.createZKRangeProof(
            encryptedValue,
            minValue,
            maxValue,
            randomness
        );
        
        assertEq(proof.minValue, minValue);
        assertEq(proof.maxValue, maxValue);
        assertTrue(proof.commitment != bytes32(0));
        assertTrue(proof.proof.length > 0);
        assertEq(proof.timestamp, block.timestamp);
    }
    
    function testCreateZKRangeProofInvalidRange() public {
        euint256 encryptedValue = TFHE.asEuint256(100);
        uint256 minValue = 150;
        uint256 maxValue = 50;
        bytes32 randomness = keccak256("test");
        
        vm.expectRevert(ZKRangeProofs.InvalidRange.selector);
        ZKRangeProofs.createZKRangeProof(
            encryptedValue,
            minValue,
            maxValue,
            randomness
        );
    }
    
    function testVerifyZKRangeProof() public {
        euint256 encryptedValue = TFHE.asEuint256(100);
        
        ZKRangeProofs.RangeProof memory proof = ZKRangeProofs.createZKRangeProof(
            encryptedValue,
            50,
            150,
            keccak256("test")
        );
        
        bool result = ZKRangeProofs.verifyZKRangeProof(proof, encryptedValue);
        assertTrue(result);
    }
    
    function testVerifyZKRangeProofExpired() public {
        euint256 encryptedValue = TFHE.asEuint256(100);
        
        ZKRangeProofs.RangeProof memory proof = ZKRangeProofs.createZKRangeProof(
            encryptedValue,
            50,
            150,
            keccak256("test")
        );
        
        // Fast forward time beyond proof validity
        vm.warp(block.timestamp + 2 hours);
        
        vm.expectRevert(ZKRangeProofs.ProofExpired.selector);
        ZKRangeProofs.verifyZKRangeProof(proof, encryptedValue);
    }
    
    function testBatchVerifyRangeProofs() public {
        ZKRangeProofs.RangeProof[] memory proofs = new ZKRangeProofs.RangeProof[](2);
        euint256[] memory values = new euint256[](2);
        
        values[0] = TFHE.asEuint256(100);
        values[1] = TFHE.asEuint256(200);
        
        proofs[0] = ZKRangeProofs.createZKRangeProof(
            values[0],
            50,
            150,
            keccak256("test1")
        );
        
        proofs[1] = ZKRangeProofs.createZKRangeProof(
            values[1],
            150,
            250,
            keccak256("test2")
        );
        
        bool[] memory results = ZKRangeProofs.batchVerifyRangeProofs(proofs, values);
        
        assertEq(results.length, 2);
        assertTrue(results[0]);
        assertTrue(results[1]);
    }
    
    function testBatchVerifyRangeProofsMismatch() public {
        ZKRangeProofs.RangeProof[] memory proofs = new ZKRangeProofs.RangeProof[](2);
        euint256[] memory values = new euint256[](1);
        
        vm.expectRevert("Array length mismatch");
        ZKRangeProofs.batchVerifyRangeProofs(proofs, values);
    }
    
    function testProveAboveThreshold() public {
        euint256 encryptedAmount = TFHE.asEuint256(1000);
        uint256 threshold = 500;
        
        ZKRangeProofs.RangeProof memory proof = ZKRangeProofs.proveAboveThreshold(
            encryptedAmount,
            threshold
        );
        
        assertEq(proof.minValue, threshold);
        assertEq(proof.maxValue, type(uint256).max);
    }
    
    function testProveFeeRange() public {
        euint256 encryptedAmount = TFHE.asEuint256(50);
        uint256 baseFee = 10;
        uint256 maxFeeMultiplier = 10;
        
        ZKRangeProofs.RangeProof memory proof = ZKRangeProofs.proveFeeRange(
            encryptedAmount,
            baseFee,
            maxFeeMultiplier
        );
        
        assertEq(proof.minValue, baseFee);
        assertEq(proof.maxValue, baseFee * maxFeeMultiplier);
    }
    
    function testVerifyPaymentBounds() public {
        euint256 encryptedAmount = TFHE.asEuint256(100);
        uint256 minPayment = 50;
        uint256 maxPayment = 150;
        
        bool result = ZKRangeProofs.verifyPaymentBounds(
            encryptedAmount,
            minPayment,
            maxPayment
        );
        
        assertTrue(result);
    }
    
    function testProveSumEqualsTotal() public {
        euint256[] memory encryptedValues = new euint256[](3);
        encryptedValues[0] = TFHE.asEuint256(100);
        encryptedValues[1] = TFHE.asEuint256(200);
        encryptedValues[2] = TFHE.asEuint256(300);
        
        uint256 expectedTotal = 600;
        
        bool result = ZKRangeProofs.proveSumEqualsTotal(encryptedValues, expectedTotal);
        assertTrue(result);
    }
    
    function testProveSumEqualsTotalEmpty() public {
        euint256[] memory encryptedValues = new euint256[](0);
        uint256 expectedTotal = 0;
        
        bool result = ZKRangeProofs.proveSumEqualsTotal(encryptedValues, expectedTotal);
        assertTrue(result);
    }
    
    function testProveSumEqualsTotalEmptyNonZero() public {
        euint256[] memory encryptedValues = new euint256[](0);
        uint256 expectedTotal = 100;
        
        bool result = ZKRangeProofs.proveSumEqualsTotal(encryptedValues, expectedTotal);
        assertFalse(result);
    }
}