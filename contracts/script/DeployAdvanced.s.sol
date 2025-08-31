// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import {Script, console} from "forge-std/Script.sol";
import "../src/ConfidentialPayments.sol";
import "../src/RelayValidator.sol";
import "../src/TrancheVault.sol";
import "../src/GrantPool.sol";
import "../src/TimelockController.sol";
import "../src/BatchOperations.sol";
import "@openzeppelin/contracts/token/ERC20/ERC20.sol";

contract MockUSDC is ERC20 {
    constructor() ERC20("Mock USDC", "USDC") {
        _mint(msg.sender, 1000000 * 10**6);
    }
    
    function decimals() public pure override returns (uint8) {
        return 6;
    }
    
    function mint(address to, uint256 amount) external {
        _mint(to, amount);
    }
}

contract DeployAdvancedScript is Script {
    ConfidentialPayments public confidentialPayments;
    RelayValidator public relayValidator;
    TrancheVault public trancheVault;
    GrantPool public grantPool;
    TimelockController public timelock;
    BatchOperations public batchOps;
    MockUSDC public usdc;
    
    address public deployer;
    address public multisig = 0x742d35Cc6634C0532925a3b8D34300e8; // Replace with actual multisig
    address public complianceRole = 0x1234567890123456789012345678901234567890; // Replace with compliance address
    
    function run() external {
        uint256 deployerPrivateKey = vm.envUint("PRIVATE_KEY");
        deployer = vm.addr(deployerPrivateKey);
        
        vm.startBroadcast(deployerPrivateKey);
        
        console.log("Deploying advanced security contracts...");
        console.log("Deployer:", deployer);
        console.log("Multisig:", multisig);
        
        // Deploy mock USDC for testing
        if (block.chainid == 31337 || block.chainid == 11155111) { // localhost or sepolia
            usdc = new MockUSDC();
            console.log("Mock USDC deployed at:", address(usdc));
        } else {
            // Use real USDC on mainnet/other networks
            usdc = MockUSDC(0xA0b86a33E6441946e6f2ad68b9C2e915F5d98D35); // Replace with actual USDC
        }
        
        // 1. Deploy ConfidentialPayments
        confidentialPayments = new ConfidentialPayments();
        console.log("ConfidentialPayments deployed at:", address(confidentialPayments));
        
        // 2. Deploy RelayValidator
        relayValidator = new RelayValidator();
        console.log("RelayValidator deployed at:", address(relayValidator));
        
        // 3. Deploy TrancheVault
        trancheVault = new TrancheVault(
            address(usdc),
            "CrossPay Tranche Vault",
            "CPV"
        );
        console.log("TrancheVault deployed at:", address(trancheVault));
        
        // 4. Deploy GrantPool
        grantPool = new GrantPool();
        console.log("GrantPool deployed at:", address(grantPool));
        
        // 5. Deploy TimelockController
        address[] memory proposers = new address[](1);
        address[] memory executors = new address[](1);
        proposers[0] = multisig;
        executors[0] = multisig;
        
        timelock = new TimelockController(
            24 hours, // minDelay
            proposers,
            executors,
            multisig // admin
        );
        console.log("TimelockController deployed at:", address(timelock));
        
        // 6. Deploy BatchOperations
        batchOps = new BatchOperations(
            address(confidentialPayments),
            address(relayValidator),
            address(trancheVault)
        );
        console.log("BatchOperations deployed at:", address(batchOps));
        
        // Setup roles and permissions
        setupRoles();
        
        // Transfer ownership to multisig
        transferOwnership();
        
        vm.stopBroadcast();
        
        // Log deployment summary
        logDeploymentSummary();
        
        console.log("Advanced security deployment completed successfully!");
    }
    
    function setupRoles() internal {
        console.log("Setting up roles and permissions...");
        
        // Grant compliance role
        confidentialPayments.grantRole(
            confidentialPayments.COMPLIANCE_ROLE(),
            complianceRole
        );
        
        // Setup initial validator threshold
        relayValidator.setHighValueThreshold(1000 * 10**6); // 1000 USDC
        
        // Set initial vault parameters
        trancheVault.setYieldRates(1200, 800, 500); // 12%, 8%, 5%
        trancheVault.setPerformanceFeeRate(1000); // 10%
        
        console.log("Roles and permissions configured");
    }
    
    function transferOwnership() internal {
        console.log("Transferring ownership to multisig...");
        
        confidentialPayments.transferOwnership(multisig);
        relayValidator.transferOwnership(multisig);
        trancheVault.transferOwnership(multisig);
        grantPool.transferOwnership(multisig);
        
        console.log("Ownership transferred to:", multisig);
    }
    
    function logDeploymentSummary() internal view {
        console.log("\n=== DEPLOYMENT SUMMARY ===");
        console.log("Network:", getNetworkName());
        console.log("Deployer:", deployer);
        console.log("Multisig:", multisig);
        console.log("Compliance Role:", complianceRole);
        console.log("");
        console.log("Contract Addresses:");
        console.log("- ConfidentialPayments:", address(confidentialPayments));
        console.log("- RelayValidator:", address(relayValidator));
        console.log("- TrancheVault:", address(trancheVault));
        console.log("- GrantPool:", address(grantPool));
        console.log("- TimelockController:", address(timelock));
        console.log("- BatchOperations:", address(batchOps));
        console.log("- USDC Token:", address(usdc));
        console.log("");
        console.log("Next Steps:");
        console.log("1. Verify contracts on block explorer");
        console.log("2. Register initial validators");
        console.log("3. Fund grant pools");
        console.log("4. Configure monitoring systems");
        console.log("5. Update frontend contract addresses");
    }
    
    function getNetworkName() internal view returns (string memory) {
        uint256 chainId = block.chainid;
        
        if (chainId == 1) return "Mainnet";
        if (chainId == 11155111) return "Sepolia";
        if (chainId == 137) return "Polygon";
        if (chainId == 8453) return "Base";
        if (chainId == 42161) return "Arbitrum";
        if (chainId == 31337) return "Localhost";
        
        return "Unknown";
    }
}