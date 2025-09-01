import { describe, it, expect, vi, beforeEach } from 'vitest';
import { OracleService } from '@services/oracle';

// Mock fetch
global.fetch = vi.fn();

describe('OracleService', () => {
	let oracleService: OracleService;

	beforeEach(() => {
		vi.clearAllMocks();
		oracleService = new OracleService('http://localhost:3003');
	});

	describe('getCurrentPrice', () => {
		it('should fetch current price successfully', async () => {
			const mockResponse = {
				symbol: 'ETH/USD',
				price: 2500.0,
				timestamp: Date.now(),
				decimals: 8,
				valid: true
			};

			vi.mocked(fetch).mockResolvedValueOnce({
				ok: true,
				json: () => Promise.resolve(mockResponse)
			} as Response);

			const result = await oracleService.getCurrentPrice('ETH/USD');

			expect(fetch).toHaveBeenCalledWith('http://localhost:3003/ftso/price/ETH/USD');
			expect(result).toEqual(mockResponse);
		});

		it('should handle network errors', async () => {
			vi.mocked(fetch).mockRejectedValueOnce(new Error('Network error'));

			const result = await oracleService.getCurrentPrice('ETH/USD');
			expect(result).toBeNull();
		});

		it('should handle API errors', async () => {
			vi.mocked(fetch).mockResolvedValueOnce({
				ok: false,
				status: 404,
				statusText: 'Not Found'
			} as Response);

			const result = await oracleService.getCurrentPrice('INVALID/USD');
			expect(result).toBeNull();
		});

		it('should handle invalid JSON response', async () => {
			vi.mocked(fetch).mockResolvedValueOnce({
				ok: true,
				json: () => Promise.reject(new Error('Invalid JSON'))
			} as Response);

			const result = await oracleService.getCurrentPrice('ETH/USD');
			expect(result).toBeNull();
		});
	});

	describe('constructor', () => {
		it('should use provided base URL', () => {
			const service = new OracleService('https://api.example.com');
			expect(service['baseUrl']).toBe('https://api.example.com');
		});

		it('should use environment variable when no URL provided', () => {
			vi.stubEnv('VITE_ORACLE_URL', 'https://env.example.com');
			const service = new OracleService();
			expect(service['baseUrl']).toBe('https://env.example.com');
		});

		it('should fallback to localhost when no URL or env var', () => {
			vi.unstubAllEnvs();
			const service = new OracleService();
			expect(service['baseUrl']).toBe('http://localhost:3003');
		});
	});
});
