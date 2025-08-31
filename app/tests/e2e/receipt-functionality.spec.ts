import { test, expect } from '@playwright/test';

test.describe('Receipt Functionality', () => {
  test('should display receipt page with all details', async ({ page }) => {
    await page.goto('/receipt/1');
    
    // Should show receipt header
    await expect(page.locator('h1')).toContainText('Payment Receipt');
    await expect(page.locator('text=Receipt #1')).toBeVisible();
    
    // Should show payment details
    await expect(page.locator('text=Payment Details')).toBeVisible();
    await expect(page.locator('text=Amount')).toBeVisible();
    await expect(page.locator('text=Protocol Fee')).toBeVisible();
    
    // Should show participants section
    await expect(page.locator('text=Participants')).toBeVisible();
    await expect(page.locator('text=From (Sender)')).toBeVisible();
    await expect(page.locator('text=To (Recipient)')).toBeVisible();
    
    // Should show transaction info
    await expect(page.locator('text=Transaction Info')).toBeVisible();
    await expect(page.locator('text=Transaction Hash')).toBeVisible();
    await expect(page.locator('text=Block Number')).toBeVisible();
  });

  test('should handle ENS name display', async ({ page }) => {
    await page.goto('/receipt/1');
    
    // Wait for page to load
    await page.waitForTimeout(2000);
    
    // ENS section should be visible if ENS names are present
    const ensSection = page.locator('text=ENS Names');
    if (await ensSection.isVisible()) {
      await expect(page.locator('text=Sender ENS')).toBeVisible();
      await expect(page.locator('text=Recipient ENS')).toBeVisible();
    }
  });

  test('should show receipt storage section', async ({ page }) => {
    await page.goto('/receipt/1');
    
    // Wait for page to load
    await page.waitForTimeout(2000);
    
    // Should show either stored receipt or "not yet stored" message
    const permanentStorageSection = page.locator('text=Permanent Receipt Storage');
    const notStoredSection = page.locator('text=Receipt not yet stored');
    
    await expect(permanentStorageSection.or(notStoredSection)).toBeVisible();
  });

  test('should handle download and verify receipt actions', async ({ page }) => {
    await page.goto('/receipt/1');
    
    // Wait for page to load
    await page.waitForTimeout(2000);
    
    // If receipt is stored, should show download and verify buttons
    const downloadButton = page.locator('text=Download Receipt');
    const verifyButton = page.locator('text=Verify Receipt');
    
    if (await downloadButton.isVisible()) {
      await expect(downloadButton).toBeVisible();
      await expect(verifyButton).toBeVisible();
      
      // Test navigation to verify page
      await verifyButton.click();
      await expect(page).toHaveURL(/\/verify\?cid=/);
    }
  });

  test('should show oracle price data when available', async ({ page }) => {
    await page.goto('/receipt/1');
    
    // Wait for page to load
    await page.waitForTimeout(2000);
    
    // Oracle price section should be visible if price data exists
    const oraclePriceSection = page.locator('text=Price Oracle Data');
    if (await oraclePriceSection.isVisible()) {
      await expect(page.locator('text=Exchange Rate at Payment')).toBeVisible();
      await expect(page.locator('text=Flare FTSO price feed')).toBeVisible();
    }
  });

  test('should handle copy functionality', async ({ page }) => {
    await page.goto('/receipt/1');
    
    // Wait for page to load
    await page.waitForTimeout(2000);
    
    // Copy buttons should be present
    const copyButtons = page.locator('button[title="Copy address"], button[title="Copy hash"], button[title="Copy CID"]');
    await expect(copyButtons.first()).toBeVisible();
  });

  test('should handle share and print actions', async ({ page }) => {
    await page.goto('/receipt/1');
    
    // Share and print buttons should be present
    await expect(page.locator('button', { hasText: 'Share' })).toBeVisible();
    await expect(page.locator('button', { hasText: 'Print' })).toBeVisible();
  });

  test('should show proper navigation breadcrumbs', async ({ page }) => {
    await page.goto('/receipt/1');
    
    // Breadcrumbs should be present
    await expect(page.locator('.breadcrumbs')).toBeVisible();
    await expect(page.locator('text=Home')).toBeVisible();
    await expect(page.locator('text=Receipts')).toBeVisible();
    await expect(page.locator('text=Receipt #1')).toBeVisible();
  });

  test('should be responsive on mobile', async ({ page }) => {
    await page.setViewportSize({ width: 375, height: 667 });
    await page.goto('/receipt/1');
    
    // Should show receipt details in mobile layout
    await expect(page.locator('h1')).toContainText('Payment Receipt');
    await expect(page.locator('text=Payment Details')).toBeVisible();
    
    // Action buttons should be stacked properly on mobile
    await expect(page.locator('button', { hasText: 'Share' })).toBeVisible();
    await expect(page.locator('button', { hasText: 'Print' })).toBeVisible();
  });
});