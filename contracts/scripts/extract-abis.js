const fs = require('fs');
const path = require('path');

const contracts = ['PaymentCore', 'ReceiptRegistry', 'ComplianceBase'];
const outDir = path.join(__dirname, '../out');
const typesDir = path.join(__dirname, '../../packages/types');

// Ensure types directory exists
if (!fs.existsSync(typesDir)) {
  fs.mkdirSync(typesDir, { recursive: true });
}

console.log('Extracting ABIs from Foundry build artifacts...');

contracts.forEach(contractName => {
  const artifactPath = path.join(outDir, `${contractName}.sol/${contractName}.json`);
  
  if (fs.existsSync(artifactPath)) {
    const artifact = JSON.parse(fs.readFileSync(artifactPath, 'utf8'));
    const abi = artifact.abi;
    
    // Write ABI to types package
    const abiPath = path.join(typesDir, `${contractName}.json`);
    fs.writeFileSync(abiPath, JSON.stringify(abi, null, 2));
    
    console.log(`✅ Extracted ${contractName} ABI to ${abiPath}`);
  } else {
    console.error(`❌ Artifact not found: ${artifactPath}`);
  }
});

// Generate TypeScript type definitions
const typeDefinitions = `// Auto-generated contract types
export interface Payment {
  id: bigint;
  sender: string;
  recipient: string;
  token: string;
  amount: bigint;
  fee: bigint;
  status: 0 | 1 | 2 | 3; // Pending, Completed, Refunded, Cancelled
  createdAt: bigint;
  completedAt: bigint;
  metadataURI: string;
}

export interface Receipt {
  paymentId: bigint;
  metadataCID: string;
  receiptCID: string;
  timestamp: bigint;
  creator: string;
  isPublic: boolean;
}

export interface ContractAddresses {
  PaymentCore: string;
  ReceiptRegistry: string;
  ComplianceBase?: string;
  deployer: string;
  chainId: string;
  timestamp: string;
}

export type PaymentStatus = 'pending' | 'completed' | 'refunded' | 'cancelled';

export type Address = \`0x\${string}\`;
`;

const typesPath = path.join(typesDir, 'contracts.ts');
fs.writeFileSync(typesPath, typeDefinitions);

console.log(`✅ Generated TypeScript types at ${typesPath}`);
console.log('ABI extraction complete!');