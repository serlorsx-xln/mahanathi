import { expect, test } from '@playwright/test';

// The root IS the relief (the entry fix): the smoke gate asserts the
// product chrome and a clean console, not the Phase-1 wordmark stub.
test('the page loads as the relief surface with no console errors', async ({ page }) => {
  const errors: string[] = [];
  page.on('console', (msg) => {
    if (msg.type() === 'error') errors.push(msg.text());
  });
  page.on('pageerror', (err) => errors.push(err.message));

  await page.goto('/');

  await expect(page).toHaveTitle('มหานที');
  await expect(page.locator('.lookdev__wordmark')).toHaveText('มหานที');
  await expect(page.getByRole('button', { name: 'TERRAIN 3D' })).toBeVisible();
  await expect(page.getByRole('button', { name: /ใกล้ฉัน/ })).toBeVisible();
  expect(errors).toEqual([]);
});
