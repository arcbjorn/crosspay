import { test, expect } from '@playwright/test';

test.describe('Mini App', () => {
	test('should load mini app successfully', async ({ page }) => {
		await page.goto('/mini/');

		// Check that the page loads
		await expect(page).toHaveTitle(/CrossPay Mini/);
		await expect(page.locator('h1')).toContainText('CrossPay Mini');

		// Check key UI elements exist
		await expect(page.locator('text=Quick Pay')).toBeVisible();
		await expect(page.locator('text=Streak')).toBeVisible();
		await expect(page.locator('text=Referred')).toBeVisible();
	});

	test('should display payment form in mini app', async ({ page }) => {
		await page.goto('/mini/');

		// Check payment form elements
		await expect(page.locator('input[placeholder*="0x"]')).toBeVisible();
		await expect(page.locator('input[placeholder*="Amount"]')).toBeVisible();
		await expect(page.locator('button[type="submit"]')).toBeVisible();
	});

	test('should show QR code functionality', async ({ page }) => {
		await page.goto('/mini/');

		// Look for QR code related elements
		await expect(page.locator('text=Scan QR')).toBeVisible();
	});

	test('should display viral mechanics', async ({ page }) => {
		await page.goto('/mini/');

		// Check for streak and referral elements
		await expect(page.locator('text=7 day streak')).toBeVisible();
		await expect(page.locator('text=5 people')).toBeVisible();

		// Check for share functionality
		await expect(page.locator('text=Share CrossPay')).toBeVisible();
	});

	test('should be under 100KB', async ({ page }) => {
		const response = await page.goto('/mini/');
		const buffer = await response?.body();

		if (buffer) {
			const sizeKB = buffer.length / 1024;
			console.log(`Mini app size: ${sizeKB.toFixed(2)} KB`);

			// Should be under 100KB as per requirements
			expect(sizeKB).toBeLessThan(100);
		}
	});

	test('should work on mobile viewport', async ({ page }) => {
		await page.setViewportSize({ width: 375, height: 812 }); // iPhone X
		await page.goto('/mini/');

		await expect(page.locator('h1')).toBeVisible();
		await expect(page.locator('input[placeholder*="0x"]')).toBeVisible();

		// Should be touch-friendly
		const payButton = page.locator('button[type="submit"]');
		await expect(payButton).toBeVisible();

		// Check minimum touch target size (44px is iOS guideline)
		const buttonBox = await payButton.boundingBox();
		expect(buttonBox?.height).toBeGreaterThanOrEqual(44);
	});
});
