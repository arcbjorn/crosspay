import type { Address } from 'viem';

export interface TokenInfo {
	address: Address;
	symbol: string;
	name: string;
	decimals: number;
	logoUrl?: string;
	isNative?: boolean;
	coingeckoId?: string;
}

export interface ChainTokens {
	[chainId: number]: TokenInfo[];
}

// Native token addresses (0x0 represents native currency)
export const NATIVE_TOKEN_ADDRESS: Address = '0x0000000000000000000000000000000000000000';

// Token configurations per chain
export const SUPPORTED_TOKENS: ChainTokens = {
	// Lisk Sepolia
	4202: [
		{
			address: NATIVE_TOKEN_ADDRESS,
			symbol: 'ETH',
			name: 'Ethereum',
			decimals: 18,
			isNative: true,
			coingeckoId: 'ethereum'
		},
		{
			address: '0x326C977E6efc84E512bB9C30f76E30c160eD06FB' as Address,
			symbol: 'LINK',
			name: 'Chainlink Token',
			decimals: 18,
			coingeckoId: 'chainlink'
		},
		{
			address: '0xf08A50178dfcDe18524640EA6618a1f965821715' as Address,
			symbol: 'USDC',
			name: 'USD Coin',
			decimals: 6,
			coingeckoId: 'usd-coin'
		}
	],

	// Base Sepolia
	84532: [
		{
			address: NATIVE_TOKEN_ADDRESS,
			symbol: 'ETH',
			name: 'Ethereum',
			decimals: 18,
			isNative: true,
			coingeckoId: 'ethereum'
		},
		{
			address: '0x036CbD53842c5426634e7929541eC2318f3dCF7e' as Address,
			symbol: 'USDC',
			name: 'USD Coin',
			decimals: 6,
			coingeckoId: 'usd-coin'
		},
		{
			address: '0xE4aB69C077896252FAFBD49EFD26B5D171A32410' as Address,
			symbol: 'LINK',
			name: 'Chainlink Token',
			decimals: 18,
			coingeckoId: 'chainlink'
		}
	],

	// Citrea Testnet
	5115: [
		{
			address: NATIVE_TOKEN_ADDRESS,
			symbol: 'cBTC',
			name: 'Citrea Bitcoin',
			decimals: 8,
			isNative: true,
			coingeckoId: 'bitcoin'
		}
		// Citrea may have wrapped tokens, but for now just native cBTC
	]
};

/**
 * Get supported tokens for a specific chain
 */
export function getSupportedTokens(chainId: number): TokenInfo[] {
	return SUPPORTED_TOKENS[chainId] || [];
}

/**
 * Get token info by address and chain
 */
export function getTokenInfo(chainId: number, tokenAddress: string): TokenInfo | undefined {
	const tokens = getSupportedTokens(chainId);
	return tokens.find((token) => token.address.toLowerCase() === tokenAddress.toLowerCase());
}

/**
 * Get native token for a chain
 */
export function getNativeToken(chainId: number): TokenInfo | undefined {
	const tokens = getSupportedTokens(chainId);
	return tokens.find((token) => token.isNative);
}

/**
 * Check if a token is native (ETH, cBTC, etc.)
 */
export function isNativeToken(tokenAddress: string): boolean {
	return (
		tokenAddress === NATIVE_TOKEN_ADDRESS ||
		tokenAddress.toLowerCase() === NATIVE_TOKEN_ADDRESS.toLowerCase()
	);
}

/**
 * Format token amount for display
 */
export function formatTokenAmount(
	amount: bigint,
	decimals: number,
	displayDecimals: number = 4
): string {
	const divisor = 10n ** BigInt(decimals);
	const whole = amount / divisor;
	const remainder = amount % divisor;

	if (remainder === 0n) {
		return whole.toString();
	}

	const decimalString = remainder.toString().padStart(decimals, '0');
	const trimmedDecimals = decimalString.slice(0, displayDecimals).replace(/0+$/, '');

	return trimmedDecimals ? `${whole}.${trimmedDecimals}` : whole.toString();
}

/**
 * Parse token amount from string
 */
export function parseTokenAmount(amount: string, decimals: number): bigint {
	if (!amount || isNaN(Number(amount))) {
		return 0n;
	}

	const [whole, decimal = ''] = amount.split('.');
	const paddedDecimal = decimal.padEnd(decimals, '0').slice(0, decimals);

	return BigInt(whole + paddedDecimal);
}

/**
 * Get token logo URL (placeholder implementation)
 */
export function getTokenLogoUrl(symbol: string): string {
	const logoMap: Record<string, string> = {
		ETH: 'https://ethereum.org/static/6b935ac0e6194247347855dc3d328e83/6ed5f/eth-diamond-black.webp',
		USDC: 'https://cryptologos.cc/logos/usd-coin-usdc-logo.svg',
		LINK: 'https://cryptologos.cc/logos/chainlink-link-logo.svg',
		cBTC: 'https://cryptologos.cc/logos/bitcoin-btc-logo.svg'
	};

	return (
		logoMap[symbol] || `https://via.placeholder.com/24x24/4F46E5/FFFFFF?text=${symbol.slice(0, 2)}`
	);
}

/**
 * Common testnet token addresses for development
 */
export const TESTNET_TOKENS = {
	USDC_SEPOLIA: '0xf08A50178dfcDe18524640EA6618a1f965821715' as Address,
	LINK_SEPOLIA: '0x326C977E6efc84E512bB9C30f76E30c160eD06FB' as Address,
	USDC_BASE_SEPOLIA: '0x036CbD53842c5426634e7929541eC2318f3dCF7e' as Address,
	LINK_BASE_SEPOLIA: '0xE4aB69C077896252FAFBD49EFD26B5D171A32410' as Address
} as const;
