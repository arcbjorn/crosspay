import { describe, it, expect, vi, beforeEach } from 'vitest';
import { ENSService } from '@services/ens';

// Mock fetch
global.fetch = vi.fn();

describe('ENSService', () => {
	let ensService: ENSService;

	beforeEach(() => {
		vi.clearAllMocks();
		ensService = new ENSService('http://localhost:3002');
	});

	describe('resolveName', () => {
		it('should resolve ENS name to address successfully', async () => {
			const mockResponse = {
				address: '0x1234567890123456789012345678901234567890'
			};

			vi.mocked(fetch).mockResolvedValueOnce({
				ok: true,
				json: () => Promise.resolve(mockResponse)
			} as Response);

			const result = await ensService.resolveName('alice.eth');

			expect(fetch).toHaveBeenCalledWith('http://localhost:3002/resolve/alice.eth');
			expect(result).toBe('0x1234567890123456789012345678901234567890');
		});

		it('should handle unresolved ENS names', async () => {
			const mockResponse = {
				error: 'ENS name not found'
			};

			vi.mocked(fetch).mockResolvedValueOnce({
				ok: true,
				json: () => Promise.resolve(mockResponse)
			} as Response);

			const result = await ensService.resolveName('nonexistent.eth');

			expect(result).toBeNull();
		});

		it('should handle network errors', async () => {
			vi.mocked(fetch).mockRejectedValueOnce(new Error('Network error'));

			const result = await ensService.resolveName('alice.eth');
			expect(result).toBeNull();
		});

		it('should handle API errors', async () => {
			vi.mocked(fetch).mockResolvedValueOnce({
				ok: false,
				status: 500,
				statusText: 'Internal Server Error'
			} as Response);

			const result = await ensService.resolveName('alice.eth');
			expect(result).toBeNull();
		});
	});

	describe('lookupAddress', () => {
		it('should reverse resolve address to ENS name successfully', async () => {
			const mockResponse = {
				name: 'alice.eth',
				address: '0x1234567890123456789012345678901234567890'
			};

			vi.mocked(fetch).mockResolvedValueOnce({
				ok: true,
				json: () => Promise.resolve(mockResponse)
			} as Response);

			const result = await ensService.lookupAddress('0x1234567890123456789012345678901234567890');

			expect(fetch).toHaveBeenCalledWith(
				'http://localhost:3002/reverse/0x1234567890123456789012345678901234567890'
			);
			expect(result).toBe('alice.eth');
		});

		it('should handle addresses without ENS names', async () => {
			const mockResponse = {
				error: 'No ENS name found for this address'
			};

			vi.mocked(fetch).mockResolvedValueOnce({
				ok: true,
				json: () => Promise.resolve(mockResponse)
			} as Response);

			const result = await ensService.lookupAddress('0x9999999999999999999999999999999999999999');

			expect(result).toBeNull();
		});
	});

	describe('constructor', () => {
		it('should use provided base URL', () => {
			const service = new ENSService('https://api.example.com');
			expect(service['baseUrl']).toBe('https://api.example.com');
		});

		it('should use environment variable when no URL provided', () => {
			vi.stubEnv('VITE_ENS_URL', 'https://env.example.com');
			const service = new ENSService();
			expect(service['baseUrl']).toBe('https://env.example.com');
		});

		it('should fallback to localhost when no URL or env var', () => {
			vi.unstubAllEnvs();
			const service = new ENSService();
			expect(service['baseUrl']).toBe('http://localhost:3002');
		});
	});
});
