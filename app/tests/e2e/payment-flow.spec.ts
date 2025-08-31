import { test, expect } from '@playwright/test';

test.describe('Payment Flow', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
  });

  test('should navigate through the payment flow', async ({ page }) => {
    // Test homepage loads
    await expect(page).toHaveTitle(/CrossPay/);
    await expect(page.locator('h1')).toContainText('CrossPay Protocol');

    // Navigate to payment form
    await page.click('text=Send Payment');
    await expect(page).toHaveURL('/pay');
    await expect(page.locator('h2')).toContainText('Send Payment');

    // Test form validation
    await page.click('button[type="submit"]');
    await expect(page.locator('.alert-error')).toBeVisible();

    // Fill out payment form with valid data
    await page.selectOption('select[id="chain"]', '4202'); // Lisk Sepolia
    await page.fill('input[id="recipient"]', '0x742d35Cc6634C0532925a3b8D5c9a7f53b3e1234');
    await page.fill('input[id="amount"]', '0.1');
    
    // Verify fee calculation
    await expect(page.locator('text=0.0001 ETH')).toBeVisible(); // 0.1% fee

    // Note: We can't actually submit without wallet connection in E2E
    // This tests the form validation and UI flow
  });

  test('should display receipts page', async ({ page }) => {
    await page.goto('/receipts');
    await expect(page.locator('h1')).toContainText('Payment History');
    
    // Should show connect wallet message when not connected
    await expect(page.locator('.alert-info')).toContainText('Connect your wallet');
  });

  test('should handle individual receipt page', async ({ page }) => {
    await page.goto('/receipt/1');
    await expect(page.locator('h1')).toContainText('Payment Receipt');
    
    // Should show loading or error state without real data
    await page.waitForTimeout(2000); // Wait for loading
    await expect(page.locator('text=Receipt #1')).toBeVisible();
  });

  test('should be responsive on mobile', async ({ page }) => {
    await page.setViewportSize({ width: 375, height: 667 });
    
    await expect(page.locator('nav')).toBeVisible();
    await expect(page.locator('h1')).toContainText('CrossPay Protocol');
    
    // Test mobile navigation
    await page.goto('/pay');
    await expect(page.locator('h2')).toContainText('Send Payment');
    
    // Form should be usable on mobile
    await expect(page.locator('input[id="recipient"]')).toBeVisible();
  });

  test('should validate ethereum addresses', async ({ page }) => {
    await page.goto('/pay');
    
    // Invalid address
    await page.fill('input[id="recipient"]', 'invalid-address');
    await expect(page.locator('text=Please enter a valid address')).toBeVisible();
    
    // Valid address
    await page.fill('input[id="recipient"]', '0x742d35Cc6634C0532925a3b8D5c9a7f53b3e1234');
    await expect(page.locator('text=Please enter a valid address')).not.toBeVisible();
    
    // ENS name detection
    await page.fill('input[id="recipient"]', 'vitalik.eth');
    await expect(page.locator('text=ENS name detected')).toBeVisible();
  });

  test('should handle chain switching', async ({ page }) => {
    await page.goto('/pay');
    
    // Switch to Base Sepolia
    await page.selectOption('select[id="chain"]', '84532');
    await expect(page.locator('option[value="84532"]')).toBeVisible();
    
    // Switch to Lisk Sepolia
    await page.selectOption('select[id="chain"]', '4202');
    await expect(page.locator('option[value="4202"]')).toBeVisible();
  });
});