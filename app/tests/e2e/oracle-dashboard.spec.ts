import { test, expect } from '@playwright/test';

test.describe('Oracle Dashboard', () => {
	test.beforeEach(async ({ page }) => {
		await page.goto('/oracles');
	});

	test('should display oracle health status', async ({ page }) => {
		await expect(page.locator('h2')).toContainText('Oracle Health Status');

		// Should show health indicators
		await expect(page.locator('text=Overall')).toBeVisible();
		await expect(page.locator('text=FTSO')).toBeVisible();
		await expect(page.locator('text=RNG')).toBeVisible();
		await expect(page.locator('text=FDC')).toBeVisible();

		// Refresh button should be present
		await expect(page.locator('button', { hasText: 'Refresh' })).toBeVisible();
	});

	test('should display real-time price feeds', async ({ page }) => {
		await expect(page.locator('h2')).toContainText('Real-time Price Feeds');

		// Should show supported trading pairs
		await expect(page.locator('text=ETH/USD')).toBeVisible();
		await expect(page.locator('text=BTC/USD')).toBeVisible();
		await expect(page.locator('text=USDC/USD')).toBeVisible();
	});

	test('should handle random number generation', async ({ page }) => {
		await expect(page.locator('h2')).toContainText('Secure Random Numbers');

		// Request random number button should be present
		const requestButton = page.locator('button', { hasText: 'Request Random Number' });
		await expect(requestButton).toBeVisible();

		// Click to request (will show loading state)
		await requestButton.click();
		await expect(page.locator('text=Requesting...')).toBeVisible();
	});

	test('should handle FDC proof submission', async ({ page }) => {
		await expect(page.locator('h2')).toContainText('External Proof Verification (FDC)');

		// Should show form inputs
		await expect(page.locator('input[placeholder*="proof identifier"]')).toBeVisible();
		await expect(page.locator('input[placeholder*="0x"]')).toBeVisible();
		await expect(page.locator('textarea[placeholder*="comma-separated"]')).toBeVisible();

		// Fill form with test data
		await page.fill('input[placeholder*="proof identifier"]', 'test-proof-123');
		await page.fill('input[placeholder*="0x"]', '0x1234567890abcdef');
		await page.fill('textarea[placeholder*="comma-separated"]', '0xabc123, 0xdef456');

		// Submit button should be enabled
		const submitButton = page.locator('button', { hasText: 'Submit Proof' });
		await expect(submitButton).toBeVisible();
		await expect(submitButton).toBeEnabled();
	});

	test('should handle proof verification', async ({ page }) => {
		// Fill in proof ID for verification
		await page.fill('input[placeholder*="Enter proof ID to verify"]', 'test-proof-123');

		// Verify button should be enabled
		const verifyButton = page.locator('button', { hasText: 'Verify Proof' });
		await expect(verifyButton).toBeVisible();
		await expect(verifyButton).toBeEnabled();

		// Click verify button
		await verifyButton.click();
		await expect(page.locator('text=Verifying...')).toBeVisible();
	});

	test('should be responsive on mobile', async ({ page }) => {
		await page.setViewportSize({ width: 375, height: 667 });

		// All sections should be visible and usable on mobile
		await expect(page.locator('h2')).toContainText('Oracle Health Status');
		await expect(page.locator('text=Real-time Price Feeds')).toBeVisible();
		await expect(page.locator('text=Secure Random Numbers')).toBeVisible();
		await expect(page.locator('text=External Proof Verification')).toBeVisible();
	});
});
