// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

interface IOracle {
    struct PriceData {
        uint256 price;
        uint256 timestamp;
        uint256 decimals;
        bool valid;
    }

    struct RandomData {
        bytes32 seed;
        uint256 timestamp;
        bool fulfilled;
    }

    struct ProofData {
        bytes32 merkleRoot;
        bytes32[] proof;
        bytes data;
        uint256 timestamp;
        bool verified;
    }

    event PriceUpdated(string indexed symbol, uint256 price, uint256 timestamp);
    event RandomRequested(bytes32 indexed requestId, uint256 timestamp);
    event RandomFulfilled(bytes32 indexed requestId, bytes32 seed);
    event ProofSubmitted(bytes32 indexed proofId, bytes32 merkleRoot, uint256 timestamp);
    event ProofVerified(bytes32 indexed proofId, bool valid);

    function getCurrentPrice(string calldata symbol) external view returns (PriceData memory);
    function getPriceAtTimestamp(string calldata symbol, uint256 timestamp) external view returns (PriceData memory);
    function requestRandomNumber() external returns (bytes32 requestId);
    function getRandomData(bytes32 requestId) external view returns (RandomData memory);
    function submitExternalProof(bytes32 proofId, bytes32 merkleRoot, bytes32[] calldata proof, bytes calldata data) external;
    function verifyExternalProof(bytes32 proofId) external view returns (bool);
    function isOracleHealthy() external view returns (bool);
}