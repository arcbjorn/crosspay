import { test, expect } from '@playwright/test';

test.describe('Storage and Verification', () => {
	test('should display storage page', async ({ page }) => {
		await page.goto('/storage');

		await expect(page.locator('h1')).toContainText('Filecoin Storage');
		await expect(page.locator('text=Permanent Storage')).toBeVisible();
	});

	test('should handle file verification page', async ({ page }) => {
		await page.goto('/verify');

		await expect(page.locator('h1')).toContainText('Verify Receipt');

		// Should show CID input form
		await expect(page.locator('input[placeholder*="CID"]')).toBeVisible();
		await expect(page.locator('button', { hasText: 'Verify' })).toBeVisible();
	});

	test('should handle verification with CID parameter', async ({ page }) => {
		await page.goto('/verify?cid=bafybeigtest123');

		// Should auto-populate CID field
		await expect(page.locator('input[value="bafybeigtest123"]')).toBeVisible();
	});

	test('should validate CID format', async ({ page }) => {
		await page.goto('/verify');

		// Invalid CID
		await page.fill('input[placeholder*="CID"]', 'invalid-cid');
		await page.click('button[type="submit"]');
		await expect(page.locator('text=Invalid CID format')).toBeVisible();

		// Valid CID format
		await page.fill('input[placeholder*="CID"]', 'bafybeigtest123456789abcdefghijklmnopqrstuvwxyz');
		await expect(page.locator('text=Invalid CID format')).not.toBeVisible();
	});

	test('should handle ENS resolution page', async ({ page }) => {
		await page.goto('/ens');

		await expect(page.locator('h1')).toContainText('ENS Resolution');

		// Should show resolution form
		await expect(page.locator('input[placeholder*="ENS name"]')).toBeVisible();
		await expect(page.locator('button', { hasText: 'Resolve' })).toBeVisible();

		// Should show reverse lookup form
		await expect(page.locator('input[placeholder*="Ethereum address"]')).toBeVisible();
		await expect(page.locator('button', { hasText: 'Reverse Lookup' })).toBeVisible();
	});

	test('should validate ENS name format', async ({ page }) => {
		await page.goto('/ens');

		// Invalid ENS name
		await page.fill('input[placeholder*="ENS name"]', 'invalid-name');
		await expect(page.locator('text=Must end with .eth')).toBeVisible();

		// Valid ENS name
		await page.fill('input[placeholder*="ENS name"]', 'vitalik.eth');
		await expect(page.locator('text=Must end with .eth')).not.toBeVisible();
	});

	test('should validate ethereum address format', async ({ page }) => {
		await page.goto('/ens');

		// Invalid address
		await page.fill('input[placeholder*="Ethereum address"]', 'invalid-address');
		await expect(page.locator('text=Invalid Ethereum address')).toBeVisible();

		// Valid address
		await page.fill(
			'input[placeholder*="Ethereum address"]',
			'0x742d35Cc6634C0532925a3b8D5c9a7f53b3e1234'
		);
		await expect(page.locator('text=Invalid Ethereum address')).not.toBeVisible();
	});

	test('should be responsive on mobile', async ({ page }) => {
		await page.setViewportSize({ width: 375, height: 667 });

		// Test oracle dashboard mobile view
		await page.goto('/oracles');
		await expect(page.locator('h2')).toContainText('Oracle Health Status');
		await expect(page.locator('text=FTSO')).toBeVisible();

		// Test storage page mobile view
		await page.goto('/storage');
		await expect(page.locator('h1')).toContainText('Filecoin Storage');

		// Test verification page mobile view
		await page.goto('/verify');
		await expect(page.locator('h1')).toContainText('Verify Receipt');
		await expect(page.locator('input[placeholder*="CID"]')).toBeVisible();

		// Test ENS page mobile view
		await page.goto('/ens');
		await expect(page.locator('h1')).toContainText('ENS Resolution');
		await expect(page.locator('input[placeholder*="ENS name"]')).toBeVisible();
	});
});
