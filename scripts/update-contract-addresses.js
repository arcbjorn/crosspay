#!/usr/bin/env node

/**
 * Updates contract addresses in frontend after deployment
 * Usage: node scripts/update-contract-addresses.js <chainId>
 * Example: node scripts/update-contract-addresses.js 4202
 */

const fs = require('fs');
const path = require('path');

const chainId = process.argv[2];
if (!chainId) {
  console.error('Usage: node scripts/update-contract-addresses.js <chainId>');
  process.exit(1);
}

const deploymentFile = path.join(__dirname, `../deployments/${chainId}.json`);
const contractsFile = path.join(__dirname, '../app/src/lib/contracts/index.ts');

try {
  // Read deployment addresses
  if (!fs.existsSync(deploymentFile)) {
    console.error(`Deployment file not found: ${deploymentFile}`);
    process.exit(1);
  }

  const deployment = JSON.parse(fs.readFileSync(deploymentFile, 'utf8'));
  console.log(`📋 Found deployment for chain ${chainId}:`);
  console.log(`   PaymentCore: ${deployment.PaymentCore}`);
  console.log(`   ReceiptRegistry: ${deployment.ReceiptRegistry}`);

  // Read current contracts file
  let contractsContent = fs.readFileSync(contractsFile, 'utf8');

  // Update addresses for the specific chain
  const chainIdPattern = new RegExp(`${chainId}:\\s*{[^}]*}`, 'g');
  const newChainConfig = `${chainId}: { // ${getChainName(chainId)}
    PaymentCore: '${deployment.PaymentCore}' as Address,
    ReceiptRegistry: '${deployment.ReceiptRegistry}' as Address,
  }`;

  if (contractsContent.includes(`${chainId}:`)) {
    // Update existing chain config
    contractsContent = contractsContent.replace(chainIdPattern, newChainConfig);
    console.log(`✅ Updated existing addresses for chain ${chainId}`);
  } else {
    console.log(`⚠️  Chain ${chainId} not found in contracts file`);
    console.log('Please manually add the following to CONTRACT_ADDRESSES:');
    console.log(newChainConfig);
  }

  // Write back to file
  fs.writeFileSync(contractsFile, contractsContent);
  console.log(`📝 Updated ${contractsFile}`);

  // Show next steps
  console.log('\n🚀 Next steps:');
  console.log('1. Test the frontend with deployed contracts');
  console.log('2. Verify contract functionality on block explorer');
  console.log('3. Run E2E tests against live contracts');

} catch (error) {
  console.error('❌ Error updating contract addresses:', error.message);
  process.exit(1);
}

function getChainName(chainId) {
  const chains = {
    '4202': 'Lisk Sepolia',
    '84532': 'Base Sepolia'
  };
  return chains[chainId] || `Chain ${chainId}`;
}