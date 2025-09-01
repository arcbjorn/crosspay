// Re-export types from packages for cleaner imports
import type {
	Payment as PaymentType,
	PaymentStatus as PaymentStatusType,
	Receipt as ReceiptType,
	ContractAddresses as ContractAddressesType,
	Address as AddressType
} from '@packages/types/contracts';

export type Payment = PaymentType;
export type PaymentStatus = PaymentStatusType;
export type Receipt = ReceiptType;
export type ContractAddresses = ContractAddressesType;
export type Address = AddressType;
