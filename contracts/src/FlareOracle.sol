// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin/contracts/utils/ReentrancyGuard.sol";
import "./IOracle.sol";

contract FlareOracle is IOracle, Ownable, ReentrancyGuard {
    
    // FTSO (Flare Time Series Oracle) Integration
    mapping(string => PriceData) private currentPrices;
    mapping(string => mapping(uint256 => PriceData)) private historicalPrices;
    mapping(string => bool) private supportedSymbols;
    
    // Secure Random Number Integration
    mapping(bytes32 => RandomData) private randomRequests;
    uint256 private randomRequestCounter;
    
    // FDC (Flare Data Connector) Integration
    mapping(bytes32 => ProofData) private externalProofs;
    mapping(bytes32 => bool) private verifiedProofs;
    
    // Circuit breaker and health monitoring
    bool private oracleHealthy;
    uint256 private lastHealthCheck;
    uint256 private constant HEALTH_CHECK_INTERVAL = 300; // 5 minutes
    uint256 private constant PRICE_STALENESS_THRESHOLD = 600; // 10 minutes
    
    // Oracle configuration
    uint256 private constant PRICE_DECIMALS = 8;
    uint256 private constant RANDOM_FULFILLMENT_DELAY = 60; // 1 minute
    
    constructor() Ownable(msg.sender) {
        oracleHealthy = true;
        lastHealthCheck = block.timestamp;
        
        // Initialize supported trading pairs
        supportedSymbols["ETH/USD"] = true;
        supportedSymbols["BTC/USD"] = true;
        supportedSymbols["FLR/USD"] = true;
        supportedSymbols["USDC/USD"] = true;
    }

    modifier onlyHealthy() {
        require(isOracleHealthy(), "Oracle is unhealthy");
        _;
    }

    modifier validSymbol(string calldata symbol) {
        require(supportedSymbols[symbol], "Unsupported symbol");
        _;
    }

    // FTSO Implementation
    function getCurrentPrice(string calldata symbol) 
        external 
        view 
        override 
        validSymbol(symbol) 
        returns (PriceData memory) 
    {
        PriceData memory priceData = currentPrices[symbol];
        
        // Check if price is stale
        if (block.timestamp - priceData.timestamp > PRICE_STALENESS_THRESHOLD) {
            priceData.valid = false;
        }
        
        return priceData;
    }

    function getPriceAtTimestamp(string calldata symbol, uint256 timestamp) 
        external 
        view 
        override 
        validSymbol(symbol) 
        returns (PriceData memory) 
    {
        return historicalPrices[symbol][timestamp];
    }

    function updatePrice(string calldata symbol, uint256 price, uint256 timestamp) 
        external 
        onlyOwner 
        validSymbol(symbol) 
    {
        PriceData memory newPrice = PriceData({
            price: price,
            timestamp: timestamp,
            decimals: PRICE_DECIMALS,
            valid: true
        });

        currentPrices[symbol] = newPrice;
        historicalPrices[symbol][timestamp] = newPrice;

        emit PriceUpdated(symbol, price, timestamp);
    }

    // Secure Random Number Implementation
    function requestRandomNumber() external override nonReentrant returns (bytes32 requestId) {
        randomRequestCounter++;
        requestId = keccak256(abi.encodePacked(
            block.timestamp,
            block.difficulty,
            randomRequestCounter,
            msg.sender
        ));

        randomRequests[requestId] = RandomData({
            seed: bytes32(0),
            timestamp: block.timestamp,
            fulfilled: false
        });

        emit RandomRequested(requestId, block.timestamp);
        return requestId;
    }

    function fulfillRandomNumber(bytes32 requestId, bytes32 seed) 
        external 
        onlyOwner 
    {
        require(randomRequests[requestId].timestamp > 0, "Request not found");
        require(!randomRequests[requestId].fulfilled, "Already fulfilled");
        require(
            block.timestamp >= randomRequests[requestId].timestamp + RANDOM_FULFILLMENT_DELAY,
            "Fulfillment too early"
        );

        randomRequests[requestId].seed = seed;
        randomRequests[requestId].fulfilled = true;

        emit RandomFulfilled(requestId, seed);
    }

    function getRandomData(bytes32 requestId) 
        external 
        view 
        override 
        returns (RandomData memory) 
    {
        return randomRequests[requestId];
    }

    // FDC (Flare Data Connector) Implementation
    function submitExternalProof(
        bytes32 proofId,
        bytes32 merkleRoot,
        bytes32[] calldata proof,
        bytes calldata data
    ) external override {
        require(externalProofs[proofId].timestamp == 0, "Proof already exists");

        externalProofs[proofId] = ProofData({
            merkleRoot: merkleRoot,
            proof: proof,
            data: data,
            timestamp: block.timestamp,
            verified: false
        });

        emit ProofSubmitted(proofId, merkleRoot, block.timestamp);
    }

    function verifyExternalProof(bytes32 proofId) 
        external 
        view 
        override 
        returns (bool) 
    {
        ProofData memory proofData = externalProofs[proofId];
        if (proofData.timestamp == 0) {
            return false;
        }

        // Mock verification - replace with actual Merkle proof verification
        bool isValid = _verifyMerkleProof(
            proofData.merkleRoot,
            proofData.proof,
            keccak256(proofData.data)
        );

        return isValid;
    }

    function confirmExternalProof(bytes32 proofId) external onlyOwner {
        require(externalProofs[proofId].timestamp > 0, "Proof not found");
        require(!externalProofs[proofId].verified, "Already verified");

        bool isValid = this.verifyExternalProof(proofId);
        externalProofs[proofId].verified = isValid;
        verifiedProofs[proofId] = isValid;

        emit ProofVerified(proofId, isValid);
    }

    // Health monitoring and circuit breaker
    function isOracleHealthy() public view override returns (bool) {
        if (block.timestamp - lastHealthCheck > HEALTH_CHECK_INTERVAL) {
            return false;
        }
        return oracleHealthy;
    }

    function performHealthCheck() external onlyOwner {
        // Check FTSO health
        bool ftsoHealthy = _checkFTSOHealth();
        
        // Check random number service health
        bool randomHealthy = _checkRandomHealth();
        
        // Check FDC health
        bool fdcHealthy = _checkFDCHealth();
        
        oracleHealthy = ftsoHealthy && randomHealthy && fdcHealthy;
        lastHealthCheck = block.timestamp;
    }

    function emergencyPause() external onlyOwner {
        oracleHealthy = false;
    }

    function emergencyResume() external onlyOwner {
        oracleHealthy = true;
        lastHealthCheck = block.timestamp;
    }

    // Administrative functions
    function addSupportedSymbol(string calldata symbol) external onlyOwner {
        supportedSymbols[symbol] = true;
    }

    function removeSupportedSymbol(string calldata symbol) external onlyOwner {
        supportedSymbols[symbol] = false;
    }

    function isSupportedSymbol(string calldata symbol) external view returns (bool) {
        return supportedSymbols[symbol];
    }

    // Internal functions
    function _verifyMerkleProof(
        bytes32 root,
        bytes32[] memory proof,
        bytes32 leaf
    ) internal pure returns (bool) {
        bytes32 computedHash = leaf;

        for (uint256 i = 0; i < proof.length; i++) {
            bytes32 proofElement = proof[i];
            if (computedHash <= proofElement) {
                computedHash = keccak256(abi.encodePacked(computedHash, proofElement));
            } else {
                computedHash = keccak256(abi.encodePacked(proofElement, computedHash));
            }
        }

        return computedHash == root;
    }

    function _checkFTSOHealth() internal view returns (bool) {
        // Check if we have recent price updates for critical pairs
        string[3] memory criticalPairs = ["ETH/USD", "BTC/USD", "USDC/USD"];
        
        for (uint i = 0; i < criticalPairs.length; i++) {
            PriceData memory price = currentPrices[criticalPairs[i]];
            if (price.timestamp == 0 || 
                block.timestamp - price.timestamp > PRICE_STALENESS_THRESHOLD) {
                return false;
            }
        }
        return true;
    }

    function _checkRandomHealth() internal view returns (bool) {
        // Mock health check - would verify Flare RNG service availability
        return true;
    }

    function _checkFDCHealth() internal view returns (bool) {
        // Mock health check - would verify FDC attestation service availability
        return true;
    }
}