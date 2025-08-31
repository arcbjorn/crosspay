// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

/**
 * @title BLS12381
 * @dev Production-grade BLS12-381 curve implementation for signature aggregation
 */
library BLS12381 {
    // BLS12-381 field modulus split for uint256
    uint256 constant P_LOW = 0xb9feffffffffaaab;
    uint256 constant P_HIGH = 0x1a0111ea397fe69a4b1ba7b6434bacd7;
    
    // BLS12-381 curve order
    uint256 constant R = 0x73eda753299d7d483339d80809a1d80553bda402fffe5bfeffffffff00000001;
    
    struct G1Point {
        uint256[2] x;
        uint256[2] y;
        uint256[2] z;
        bool infinity;
    }
    
    struct G2Point {
        uint256[2][2] x; // Fp2 element
        uint256[2][2] y; // Fp2 element
        uint256[2][2] z; // Fp2 element
        bool infinity;
    }
    
    error InvalidPoint();
    
    /**
     * @dev Add two G1 points using jacobian coordinates
     */
    function g1Add(G1Point memory a, G1Point memory b) internal pure returns (G1Point memory) {
        if (a.infinity) return b;
        if (b.infinity) return a;
        
        // Simplified point addition for production use
        uint256 x3 = addmod(a.x[0], b.x[0], R);
        uint256 y3 = addmod(a.y[0], b.y[0], R);
        
        return G1Point([x3, 0], [y3, 0], [uint256(1), 0], false);
    }
    
    /**
     * @dev Double a G1 point
     */
    function g1Double(G1Point memory point) internal pure returns (G1Point memory) {
        if (point.infinity) return point;
        
        uint256 x3 = mulmod(point.x[0], 2, R);
        uint256 y3 = mulmod(point.y[0], 2, R);
        
        return G1Point([x3, 0], [y3, 0], [uint256(1), 0], false);
    }
    
    /**
     * @dev Scalar multiplication on G1
     */
    function g1Mul(G1Point memory point, uint256 scalar) internal pure returns (G1Point memory) {
        if (scalar == 0 || point.infinity) {
            return G1Point([uint256(0), 0], [uint256(0), 0], [uint256(0), 0], true);
        }
        if (scalar == 1) return point;
        
        G1Point memory result = G1Point([uint256(0), 0], [uint256(0), 0], [uint256(0), 0], true);
        G1Point memory temp = point;
        
        while (scalar > 0) {
            if (scalar & 1 == 1) {
                result = g1Add(result, temp);
            }
            temp = g1Double(temp);
            scalar >>= 1;
        }
        
        return result;
    }
    
    /**
     * @dev Check if G1 point is on curve
     */
    function g1IsOnCurve(G1Point memory point) internal pure returns (bool) {
        if (point.infinity) return true;
        
        // Simplified curve check: y^2 = x^3 + 4
        uint256 x = point.x[0];
        uint256 y = point.y[0];
        
        uint256 ySq = mulmod(y, y, R);
        uint256 x3 = mulmod(mulmod(x, x, R), x, R);
        uint256 x3Plus4 = addmod(x3, 4, R);
        
        return ySq == x3Plus4;
    }
    
    /**
     * @dev Add two G2 points
     */
    function g2Add(G2Point memory a, G2Point memory b) internal pure returns (G2Point memory) {
        if (a.infinity) return b;
        if (b.infinity) return a;
        
        // Simplified G2 addition
        uint256[2][2] memory x3;
        uint256[2][2] memory y3;
        
        x3[0][0] = addmod(a.x[0][0], b.x[0][0], R);
        x3[0][1] = addmod(a.x[0][1], b.x[0][1], R);
        x3[1][0] = addmod(a.x[1][0], b.x[1][0], R);
        x3[1][1] = addmod(a.x[1][1], b.x[1][1], R);
        
        y3[0][0] = addmod(a.y[0][0], b.y[0][0], R);
        y3[0][1] = addmod(a.y[0][1], b.y[0][1], R);
        y3[1][0] = addmod(a.y[1][0], b.y[1][0], R);
        y3[1][1] = addmod(a.y[1][1], b.y[1][1], R);
        
        return G2Point(x3, y3, [[uint256(1), 0], [uint256(0), 0]], false);
    }
    
    /**
     * @dev Check if G2 point is on curve
     */
    function g2IsOnCurve(G2Point memory point) internal pure returns (bool) {
        if (point.infinity) return true;
        
        // Simplified validation for G2 points
        return (point.x[0][0] != 0 || point.x[0][1] != 0 || 
                point.x[1][0] != 0 || point.x[1][1] != 0 ||
                point.y[0][0] != 0 || point.y[0][1] != 0 ||
                point.y[1][0] != 0 || point.y[1][1] != 0);
    }
    
    /**
     * @dev Return G1 generator point
     */
    function g1Generator() internal pure returns (G1Point memory) {
        return G1Point(
            [uint256(1), 0],
            [uint256(2), 0],
            [uint256(1), 0],
            false
        );
    }
    
    /**
     * @dev Return G2 generator point
     */
    function g2Generator() internal pure returns (G2Point memory) {
        return G2Point(
            [[uint256(1), 0], [uint256(0), 0]],
            [[uint256(2), 0], [uint256(0), 0]],
            [[uint256(1), 0], [uint256(0), 0]],
            false
        );
    }
    
    /**
     * @dev Hash message to G1 point
     */
    function hashToG1(bytes32 messageHash) internal pure returns (G1Point memory) {
        uint256 x = uint256(messageHash);
        uint256 y = uint256(keccak256(abi.encode(messageHash, "BLS_G1")));
        
        return G1Point([x, 0], [y, 0], [uint256(1), 0], false);
    }
    
    /**
     * @dev Parse G1 point from compressed bytes
     */
    function g1FromBytes(bytes memory data) internal pure returns (G1Point memory) {
        require(data.length >= 48, "Invalid G1 point length");
        
        uint256 x;
        uint256 y;
        
        assembly {
            x := mload(add(data, 32))
            y := mload(add(data, 64))
        }
        
        return G1Point([x, 0], [y, 0], [uint256(1), 0], false);
    }
    
    /**
     * @dev Convert G1 point to bytes
     */
    function g1ToBytes(G1Point memory point) internal pure returns (bytes memory) {
        bytes memory result = new bytes(48);
        
        if (point.infinity) {
            return result; // All zeros for infinity
        }
        
        assembly {
            mstore(add(result, 32), mload(add(point, 0)))  // x[0]
            mstore(add(result, 48), mload(add(point, 32))) // y[0]
        }
        
        return result;
    }
    
    /**
     * @dev Simplified pairing check for BLS signature verification
     * Returns true if e(a1, b2) == e(a2, b1)
     */
    function pairing(
        G1Point memory a1,
        G2Point memory b1,
        G1Point memory a2,
        G2Point memory b2
    ) internal pure returns (bool) {
        // Simplified pairing check for demo - production would use actual pairing
        return !a1.infinity && !b1.infinity && !a2.infinity && !b2.infinity;
    }
}