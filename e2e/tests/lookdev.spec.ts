import { resolve } from 'node:path';
import { test } from '@playwright/test';

// Phase 3 look-dev evidence: screenshots of the dev routes at the two
// gate viewports, full page, into docs/design/screens/phase-3/. The
// anti-slop gate and the ADRs cite these files. The path is resolved
// from this spec's location because Playwright runs with the e2e/
// directory as cwd.
const OUT = resolve(__dirname, '../../docs/design/screens/phase-3');

test.describe('look-dev evidence pages', () => {
  test('specimen', async ({ page }) => {
    await page.goto('/specimen');
    await page.getByRole('heading', { name: 'ตัวอย่างตัวอักษร' }).waitFor();
    await page.evaluate(() => document.fonts.ready);
    await page.screenshot({ path: `${OUT}/specimen-${test.info().project.name}.png`, fullPage: true });
  });

  test('colour scale', async ({ page }) => {
    await page.goto('/color');
    await page.getByRole('heading', { name: 'มาตราส่วนสถานการณ์น้ำ' }).waitFor();
    await page.screenshot({ path: `${OUT}/color-${test.info().project.name}.png`, fullPage: true });
  });
});
