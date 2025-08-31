// Auto-generated contract types
export interface Payment {
  id: bigint;
  sender: string;
  recipient: string;
  token: string;
  amount: bigint;
  fee: bigint;
  status: PaymentStatus;
  createdAt: bigint;
  completedAt: bigint;
  metadataURI: string;
  receiptCID: string;
  senderENS: string;
  recipientENS: string;
  oraclePrice: bigint;
  randomSeed: string;
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
  FlareOracle: string;
  SubnameRegistry: string;
  ComplianceBase?: string;
  deployer: string;
  chainId: string;
  timestamp: string;
}

export type PaymentStatus = 'pending' | 'completed' | 'refunded' | 'cancelled';

export type Address = `0x${string}`;
