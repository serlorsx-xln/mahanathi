import { resolve } from 'node:path';
import { expect, test } from '@playwright/test';

// The gate's behaviour variants (design brief, section 10): the same
// pages under reduced motion and without WebGL. Both are desktop-only:
// the variant under test is the behaviour, not the viewport.
const OUT = resolve(import.meta.dirname, '../../docs/design/screens/phase-3');

test.describe('gate variants', () => {
  // The relief is WebGL-heavy and Playwright screenshots stall on
  // ReadPixels under load (the Phase 3 lesson, now true for this spec
  // too since /lookdev gained terrain, basemap and gauges).
  test.setTimeout(120_000);

  test('reduced motion cuts the flight instead of flying', async ({ page }) => {
    test.skip(test.info().project.name !== 'desktop', 'behaviour variant, desktop only');
    await page.emulateMedia({ reducedMotion: 'reduce' });
    await page.goto('/lookdev');
    await page.getByLabel('สรุปสถานการณ์น้ำ').waitFor({ timeout: 60_000 });
    await page.locator('main.lookdev[data-scene-ready="1"]').waitFor({ timeout: 60_000 });
    await page.getByRole('button', { name: /ใกล้ฉัน/ }).click();
    // No flight under reduced motion: the station is already there.
    await expect(page.getByLabel('สถานีที่เลือก')).toBeVisible();
    await expect(page.getByLabel('สถานีที่เลือก')).toContainText('สองพี่น้อง');
    await page.waitForTimeout(500);
    await page.screenshot({
      path: `${OUT}/lookdev-reduced-motion-${test.info().project.name}.png`,
    });
  });

  test('without WebGL the shell and the sentence survive', async ({ page }) => {
    test.skip(test.info().project.name !== 'desktop', 'behaviour variant, desktop only');
    // Stub the context away before any app code runs, which exercises
    // exactly the app's own detection path.
    await page.addInitScript(`
      const get = HTMLCanvasElement.prototype.getContext;
      HTMLCanvasElement.prototype.getContext = function (type, ...rest) {
        if (type === 'webgl2' || type === 'webgl') return null;
        return get.call(this, type, ...rest);
      };
    `);
    await page.goto('/lookdev');
    await page.getByLabel('สรุปสถานการณ์น้ำ').waitFor({ timeout: 60_000 });
    await expect(page.locator('.lookdev__nogl')).toContainText('ไม่รองรับกราฟิก 3D');
    // The loading status must resolve: once the data is in, no
    // skeleton may remain announcing "กำลังโหลด" forever.
    await expect(page.locator('.lookdev__skeleton')).toHaveCount(0);
    await page.screenshot({ path: `${OUT}/lookdev-no-webgl-${test.info().project.name}.png` });
  });

  test('a failed fetch names the problem and offers a retry', async ({ page }) => {
    test.skip(test.info().project.name !== 'desktop', 'behaviour variant, desktop only');
    // Brief §9: never a silent skeleton. The fixture request fails;
    // the geodata may still succeed, so fail everything to be sure.
    await page.route('**/fixture/**', (route) => route.abort());
    await page.goto('/lookdev');
    const error = page.getByLabel('โหลดข้อมูลไม่สำเร็จ');
    await expect(error).toBeVisible({ timeout: 30_000 });
    await expect(error).toContainText('โหลดข้อมูลระดับน้ำไม่ได้');
    // The retry clears the way when the network returns.
    await page.unrouteAll();
    await page.getByRole('button', { name: 'ลองใหม่' }).click();
    await page.getByLabel('สรุปสถานการณ์น้ำ').waitFor({ timeout: 60_000 });
  });
});
