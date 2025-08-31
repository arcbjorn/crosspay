// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import "forge-std/Script.sol";
import "../src/PaymentCore.sol";
import "../src/ReceiptRegistry.sol";
import "../src/ComplianceBase.sol";

contract Deploy is Script {
    function run() external {
        uint256 deployerPrivateKey = vm.envUint("PRIVATE_KEY");
        address deployer = vm.addr(deployerPrivateKey);
        
        console.log("Deploying contracts with the account:", deployer);
        console.log("Account balance:", deployer.balance);
        
        vm.startBroadcast(deployerPrivateKey);
        
        // Deploy contracts
        PaymentCore paymentCore = new PaymentCore();
        ReceiptRegistry receiptRegistry = new ReceiptRegistry();
        
        console.log("PaymentCore deployed to:", address(paymentCore));
        console.log("ReceiptRegistry deployed to:", address(receiptRegistry));
        
        // Save deployment info
        string memory deploymentInfo = string.concat(
            '{\n',
            '  "PaymentCore": "', vm.toString(address(paymentCore)), '",\n',
            '  "ReceiptRegistry": "', vm.toString(address(receiptRegistry)), '",\n',
            '  "deployer": "', vm.toString(deployer), '",\n',
            '  "chainId": "', vm.toString(block.chainid), '",\n',
            '  "timestamp": "', vm.toString(block.timestamp), '"\n',
            '}'
        );
        
        string memory filename = string.concat("deployments/", vm.toString(block.chainid), ".json");
        vm.writeFile(filename, deploymentInfo);
        
        vm.stopBroadcast();
    }
}