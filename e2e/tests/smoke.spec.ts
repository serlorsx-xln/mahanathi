import { expect, test } from '@playwright/test';

test('the page loads with the wordmark and no console errors', async ({ page }) => {
  const errors: string[] = [];
  page.on('console', (msg) => {
    if (msg.type() === 'error') errors.push(msg.text());
  });
  page.on('pageerror', (err) => errors.push(err.message));

  await page.goto('/');

  await expect(page).toHaveTitle('มหานที');
  await expect(page.getByRole('heading', { level: 1, name: 'มหานที' })).toBeVisible();
  expect(errors).toEqual([]);
});
