import { writeFileSync } from 'node:fs';
import { resolve } from 'node:path';
import { expect, test } from '@playwright/test';

// Phase 3 look-dev evidence: screenshots of the dev routes at the two
// gate viewports, full page, into docs/design/screens/phase-3/. The
// anti-slop gate and the ADRs cite these files. The path is resolved
// from this spec's location because Playwright runs with the e2e/
// directory as cwd.
const OUT = resolve(import.meta.dirname, '../../docs/design/screens/phase-3');

test.describe('look-dev evidence pages', () => {
  test('specimen', async ({ page }) => {
    await page.goto('/specimen');
    await page.getByRole('heading', { name: 'ตัวอย่างตัวอักษร' }).waitFor();
    await page.evaluate(() => document.fonts.ready);
    await page.screenshot({
      path: `${OUT}/specimen-${test.info().project.name}.png`,
      fullPage: true,
    });
  });

  test('colour scale', async ({ page }) => {
    await page.goto('/color');
    await page.getByRole('heading', { name: 'มาตราส่วนสถานการณ์น้ำ' }).waitFor();
    await page.screenshot({ path: `${OUT}/color-${test.info().project.name}.png`, fullPage: true });
  });
});

test.describe('lookdev relief', () => {
  // The scene is heavy on purpose (1.27 MB heightmap, 804 instanced
  // gauges, ~500 draped line strings) and Playwright's WebGL
  // screenshots stall on ReadPixels, so each test needs minutes, not
  // the 30 s default.
  test.setTimeout(180_000);

  // The panel can appear before the lazy scene chunk has painted; only
  // data-scene-ready (set from the Canvas onCreated) plus two rAFs
  // guarantee pixels on screen.
  async function sceneReady(page: import('@playwright/test').Page) {
    await page.locator('main.lookdev[data-scene-ready="1"]').waitFor({ timeout: 60_000 });
    await page.evaluate(
      () => new Promise<void>((r) => requestAnimationFrame(() => requestAnimationFrame(() => r()))),
    );
    await page.waitForTimeout(400); // let the demand loop settle
  }

  test('renders the relief at every preset with no console errors', async ({ page }) => {
    const errors: string[] = [];
    page.on('pageerror', (e) => errors.push(String(e)));
    page.on('console', (m) => {
      if (m.type() === 'error') errors.push(m.text());
    });

    await page.goto('/lookdev');
    // The national sentence is the LCP element; it only appears once
    // the fixture and the heightmap have loaded and the scene is up.
    await page.getByLabel('สรุปสถานการณ์น้ำ').waitFor({ timeout: 60_000 });
    await sceneReady(page);
    await page.screenshot({ path: `${OUT}/lookdev-national-${test.info().project.name}.png` });

    await page.getByRole('button', { name: 'ลุ่มเจ้าพระยา' }).click();
    await page.waitForTimeout(2100); // the preset flight is 1.6 s
    await page.screenshot({ path: `${OUT}/lookdev-plain-${test.info().project.name}.png` });

    await page.getByRole('button', { name: 'สองพี่น้อง (น้ำล้นตลิ่ง)' }).click();
    await page.waitForTimeout(2100);
    await page.screenshot({ path: `${OUT}/lookdev-station-${test.info().project.name}.png` });

    expect(errors).toEqual([]);
  });

  test('the near-me flight lands on the station', async ({ page }) => {
    await page.goto('/lookdev');
    await page.getByLabel('สรุปสถานการณ์น้ำ').waitFor({ timeout: 60_000 });
    await sceneReady(page);
    // page.screenshot stalls seconds on WebGL ReadPixels, longer than
    // the 1.6 s flight, so mid-flight evidence comes from CDP
    // screencast frames instead, which stream as the page composites.
    const cdp = await page.context().newCDPSession(page);
    const frames: Buffer[] = [];
    cdp.on('Page.screencastFrame', (ev) => {
      frames.push(Buffer.from(ev.data, 'base64'));
      void cdp.send('Page.screencastFrameAck', { sessionId: ev.sessionId }).catch(() => undefined);
    });
    await cdp.send('Page.startScreencast', { format: 'png', everyNthFrame: 1 });
    await page.getByRole('button', { name: /ใกล้ฉัน/ }).click();
    await page.waitForTimeout(1000); // inside the 1.6 s flight
    await cdp.send('Page.stopScreencast');
    // Motion is part of the deliverable: the frames during the flight
    // must stream and must differ from each other.
    expect(frames.length).toBeGreaterThan(4);
    const first = frames[0] ?? Buffer.from([]);
    const last = frames[frames.length - 1] ?? Buffer.from([]);
    expect(first.equals(last)).toBe(false);
    const mid = frames[Math.floor(frames.length / 2)] ?? Buffer.from([]);
    writeFileSync(`${OUT}/lookdev-flight-mid-${test.info().project.name}.png`, mid);

    await page.waitForTimeout(1500);
    const sheet = page.getByLabel('สถานีที่เลือก');
    await expect(sheet).toBeVisible();
    await expect(sheet).toContainText('สองพี่น้อง');
    await page.screenshot({ path: `${OUT}/lookdev-flight-end-${test.info().project.name}.png` });
  });

  test('exaggeration extremes for the ADR', async ({ page }) => {
    await page.goto('/lookdev');
    await page.getByLabel('สรุปสถานการณ์น้ำ').waitFor({ timeout: 60_000 });
    await sceneReady(page);
    // At the national zoom the fog and the distance swallow the
    // relief, so the exaggeration is compared where it is legible: the
    // central plain, ringed by ranges. The default (12x) evidence is
    // lookdev-plain, taken by the preset test.
    await page.getByRole('button', { name: 'ลุ่มเจ้าพระยา' }).click();
    await page.waitForTimeout(2100);
    const slider = page.getByLabel(/อัตรายกสูง/);
    await slider.fill('6');
    await page.waitForTimeout(600);
    await page.screenshot({ path: `${OUT}/lookdev-exagg6-${test.info().project.name}.png` });
    await slider.fill('30');
    await page.waitForTimeout(600);
    await page.screenshot({ path: `${OUT}/lookdev-exagg30-${test.info().project.name}.png` });
  });
});
