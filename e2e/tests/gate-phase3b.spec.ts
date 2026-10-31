import { resolve } from 'node:path';
import { expect, test } from '@playwright/test';

// Phase 3b gate evidence: every screenshot the gate cites, taken by a
// committed spec so the evidence is regenerable (brief 04's honest-scope
// rule; the lead's remediation blocker #1). Written to phase-3b/, never
// phase-3/ (that is history).
const OUT = resolve(import.meta.dirname, '../../docs/design/screens/phase-3b');

// The relief is WebGL-heavy and Playwright screenshots stall on
// ReadPixels under load - the Phase 3 lesson.
test.setTimeout(180_000);

async function sceneReady(page: import('@playwright/test').Page) {
  await page.locator('main.lookdev[data-scene-ready="1"]').waitFor({ timeout: 60_000 });
  await page.evaluate(
    () => new Promise<void>((r) => requestAnimationFrame(() => requestAnimationFrame(() => r()))),
  );
  await page.waitForTimeout(400);
}

// The stale close-up's subject: a real stale station from the fixture.
const STALE_STATION = { id: 640, name: 'เมืองสุโขทัย', lat: 17.02127, lon: 99.81973 };

test.describe('phase 3b gate evidence', () => {
  for (const mode of ['poles', 'discs'] as const) {
    test(`gauges (${mode}) at national and station views`, async ({ page }) => {
      const errors: string[] = [];
      page.on('pageerror', (e) => errors.push(String(e)));
      page.on('console', (m) => {
        if (m.type() === 'error') errors.push(m.text());
      });

      await page.goto(`/lookdev?gauges=${mode}`);
      await page.getByLabel('สรุปสถานการณ์น้ำ').waitFor({ timeout: 60_000 });
      await sceneReady(page);
      await page.waitForTimeout(3500); // tiles settle
      await page.screenshot({ path: `${OUT}/gauges-${mode}-national-1440.png` });

      await page.getByRole('button', { name: /ใกล้ฉัน/ }).click();
      await page.waitForFunction(() => !document.querySelector('.lookdev__flying'), null, {
        timeout: 60_000,
      });
      await page.waitForTimeout(2000);
      await page.screenshot({ path: `${OUT}/gauges-${mode}-station-1440.png` });

      expect(errors).toEqual([]);
    });
  }

  // The stale encoding's evidence (remediation #3): the hollow cap
  // stroke must actually draw - before the fix every stale fragment
  // was discarded and stale poles were invisible. The camera is cut to
  // the stale station through the e2e map handle (a cut, not a flight:
  // the subject is the pole, not the journey), then the canvas is
  // sampled for the stale cap colour: uSteel #aebac4 * 0.75, written
  // through the same path every token colour reaches the screen.
  test('the stale cap stroke is visible at close range', async ({ page }) => {
    await page.goto('/lookdev');
    await page.getByLabel('สรุปสถานการณ์น้ำ').waitFor({ timeout: 60_000 });
    await sceneReady(page);
    // The camera cut: a level-3 pole is 17 km tall and its cap band sits
    // 13.3 km up, so a close cut (zoom 12) puts the cap *above* the top
    // of the frustum - the camera is below it looking down. The cut is
    // pulled back and the centre set 18 km behind the station (opposite
    // the facing bearing) so the whole cap cross lands mid-frame.
    await page.evaluate((st) => {
      const map = (window as unknown as { __mahanathiMap?: { jumpTo: (o: object) => void } })
        .__mahanathiMap;
      const offKm = 18;
      const east = -Math.sin((-24 * Math.PI) / 180) * offKm;
      const north = -Math.cos((-24 * Math.PI) / 180) * offKm;
      map?.jumpTo({
        center: [st.lon + east / (111 * Math.cos((st.lat * Math.PI) / 180)), st.lat + north / 111],
        zoom: 10.6,
        pitch: 60,
        bearing: -24,
      });
    }, STALE_STATION);
    await page.waitForTimeout(3000); // tiles settle at the new zoom
    await page.screenshot({ path: `${OUT}/gauges-poles-stale-closeup-1440.png` });

    // The stale cap: steel * 0.75, unmixed - at this camera the station
    // is ~40 km out and the fog starts at ~355 km, so nothing blends.
    const capPx = await page.evaluate(() => {
      const src = document.querySelector('.maplibregl-canvas');
      const c = document.createElement('canvas');
      c.width = src.width;
      c.height = src.height;
      const ctx = c.getContext('2d', { willReadFrequently: true });
      ctx.drawImage(src, 0, 0);
      const d = ctx.getImageData(0, 0, c.width, c.height).data;
      // #aebac4 * 0.75 = (130, 139, 147); generous bands for AA edges.
      let n = 0;
      for (let i = 0; i < d.length; i += 4) {
        if (
          d[i] >= 105 &&
          d[i] <= 155 &&
          d[i + 1] >= 114 &&
          d[i + 1] <= 164 &&
          d[i + 2] >= 122 &&
          d[i + 2] <= 172 &&
          d[i] < d[i + 1] &&
          d[i + 1] < d[i + 2]
        )
          n++;
      }
      return n;
    });
    // The hollow cap is a thin cross on a slim blade: hundreds of
    // pixels at this range, but strictly more than zero (the bug was
    // exactly zero). The strict r < g < b ordering keeps the horizon
    // atmosphere out of the band.
    expect(capPx).toBeGreaterThan(20);
  });

  test('gate shots at three viewports plus variants', async ({ page, browser }) => {
    const errors: string[] = [];
    page.on('pageerror', (e) => errors.push(String(e)));
    await page.goto('/lookdev');
    await page.getByLabel('สรุปสถานการณ์น้ำ').waitFor({ timeout: 60_000 });
    await sceneReady(page);
    await page.waitForTimeout(3500);
    await page.screenshot({ path: `${OUT}/gate-national-desktop.png` });
    await page.getByRole('button', { name: /ใกล้ฉัน/ }).click();
    await page.waitForFunction(() => !document.querySelector('.lookdev__flying'), null, {
      timeout: 60_000,
    });
    await page.waitForTimeout(2000);
    await page.screenshot({ path: `${OUT}/gate-station-desktop.png` });
    expect(errors).toEqual([]);

    // Phone + tablet in fresh contexts (the viewport is the variant).
    for (const [name, width, height] of [
      ['phone', 390, 844],
      ['tablet', 768, 1024],
    ] as const) {
      const ctx = await browser.newContext({
        viewport: { width, height },
        locale: 'th-TH',
        timezoneId: 'Asia/Bangkok',
      });
      const p = await ctx.newPage();
      await p.goto('/lookdev');
      await p.getByLabel('สรุปสถานการณ์น้ำ').waitFor({ timeout: 60_000 });
      await p.locator('main.lookdev[data-scene-ready="1"]').waitFor({ timeout: 60_000 });
      await p.waitForTimeout(3500);
      await p.screenshot({ path: `${OUT}/gate-national-${name}.png` });
      await p.getByRole('button', { name: /ใกล้ฉัน/ }).click();
      await p.waitForFunction(() => !document.querySelector('.lookdev__flying'), null, {
        timeout: 60_000,
      });
      await p.waitForTimeout(2000);
      await p.screenshot({ path: `${OUT}/gate-station-${name}.png` });
      await ctx.close();
    }

    // Reduced motion: the cut, not the flight.
    const rm = await browser.newContext({
      viewport: { width: 1440, height: 900 },
      locale: 'th-TH',
      timezoneId: 'Asia/Bangkok',
      reducedMotion: 'reduce',
    });
    const rp = await rm.newPage();
    await rp.goto('/lookdev');
    await rp.getByLabel('สรุปสถานการณ์น้ำ').waitFor({ timeout: 60_000 });
    await rp.locator('main.lookdev[data-scene-ready="1"]').waitFor({ timeout: 60_000 });
    await rp.waitForTimeout(3000);
    await rp.getByRole('button', { name: /ใกล้ฉัน/ }).click();
    await rp.waitForTimeout(2500);
    await rp.screenshot({ path: `${OUT}/gate-station-reduced-motion.png` });
    await rm.close();
  });
});
