// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import "forge-std/Script.sol";
import "../src/PaymentCore.sol";
import "../src/ReceiptRegistry.sol";
import "../src/ComplianceBase.sol";

contract GenerateABIs is Script {
    function run() external view {
        // This script generates ABI files for TypeScript consumption
        // The actual ABI generation is handled by the forge build process
        // and post-processed by the generate-abis npm script
        
        console.log("ABI generation script executed");
        console.log("PaymentCore ABI will be extracted from out/PaymentCore.sol/PaymentCore.json");
        console.log("ReceiptRegistry ABI will be extracted from out/ReceiptRegistry.sol/ReceiptRegistry.json");
        console.log("ComplianceBase ABI will be extracted from out/ComplianceBase.sol/ComplianceBase.json");
    }
}