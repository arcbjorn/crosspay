// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import "forge-std/Script.sol";
import "../src/PaymentCore.sol";
import "../src/FlareOracle.sol";

contract DeployCitrea is Script {
    function run() external {
        uint256 deployerPrivateKey = vm.envUint("PRIVATE_KEY");
        address deployer = vm.addr(deployerPrivateKey);
        
        console.log("Deploying to Citrea testnet...");
        console.log("Deployer address:", deployer);
        
        vm.startBroadcast(deployerPrivateKey);
        
        // Deploy PaymentCore for Citrea (cBTC support)
        PaymentCore paymentCore = new PaymentCore();
        console.log("PaymentCore deployed at:", address(paymentCore));
        
        // Deploy FlareOracle (for price feeds even on Citrea)
        FlareOracle oracle = new FlareOracle();
        console.log("FlareOracle deployed at:", address(oracle));
        
        // Set up initial configuration for cBTC
        console.log("Setting up cBTC configuration...");
        
        // Add supported symbols for Citrea
        oracle.addSupportedSymbol("CBTC/USD");
        oracle.addSupportedSymbol("BTC/USD");
        oracle.addSupportedSymbol("USDC/USD");
        
        // Update initial prices
        oracle.updatePrice("CBTC/USD", 4500000000000, block.timestamp); // $45,000.00 (8 decimals)
        oracle.updatePrice("BTC/USD", 4500000000000, block.timestamp);   // $45,000.00 (8 decimals)
        oracle.updatePrice("USDC/USD", 100000000, block.timestamp);      // $1.00 (8 decimals)
        
        console.log("Initial prices set for Citrea testnet");
        
        vm.stopBroadcast();
        
        // Write deployment addresses to file
        string memory deploymentInfo = string(abi.encodePacked(
            "# Citrea Testnet Deployment\n",
            "PaymentCore: ", vm.toString(address(paymentCore)), "\n",
            "FlareOracle: ", vm.toString(address(oracle)), "\n",
            "Network: Citrea Testnet (Chain ID: 5115)\n",
            "Deployed by: ", vm.toString(deployer), "\n",
            "Deployment time: ", vm.toString(block.timestamp), "\n"
        ));
        
        vm.writeFile("deployments/citrea.txt", deploymentInfo);
        console.log("Deployment info written to deployments/citrea.txt");
    }
}