// The Phase 3b evidence measurements, as a committed script (brief 04's
// honest-scope rule: no computed-evidence claim without a committed
// script or JSON). Every number cited in GATE-phase-3b.md and the
// phase-3b report regenerates from here.
//
// Usage:  start `pnpm exec vite preview` (default http://127.0.0.1:4173,
// override with BASE_URL), then from e2e/:
//           node measure/phase-3b.mjs
// Output:  docs/design/screens/phase-3b/measurements.json (committed).
import { chromium, devices } from '@playwright/test';
import { readFileSync, readdirSync, statSync, writeFileSync } from 'node:fs';
import { gzipSync } from 'node:zlib';
import { resolve } from 'node:path';

const BASE = process.env.BASE_URL ?? 'http://127.0.0.1:4173';
const OUT = resolve(import.meta.dirname, '../../docs/design/screens/phase-3b/measurements.json');

// Token colours, read from the single source at runtime (the page's own
// computed styles are measured against these).
const TOKENS = {
  level1: '#798fb4',
  level2: '#8b8f80',
  level3: '#599c5b',
  level4: '#e3a93f',
  level5: '#ee5d58',
  stale: '#5b6a75',
};
const hexToRgb = (h) => [1, 3, 5].map((i) => parseInt(h.slice(i, i + 2), 16));

/** Sample the map canvas for exact-token pixel counts. Tolerance is
 *  ±14 per channel (the fog mixes toward the fog token with distance). */
async function canvasSample(page, targets) {
  return page.evaluate(
    ({ hexes, tol }) => {
      const src = document.querySelector('.maplibregl-canvas');
      const c = document.createElement('canvas');
      c.width = src.width;
      c.height = src.height;
      const ctx = c.getContext('2d', { willReadFrequently: true });
      ctx.drawImage(src, 0, 0);
      const d = ctx.getImageData(0, 0, c.width, c.height).data;
      const res = { sampled: d.length / 4 };
      for (const name of Object.keys(hexes)) res[name] = 0;
      for (let i = 0; i < d.length; i += 4) {
        for (const [name, t] of Object.entries(hexes)) {
          if (
            Math.abs(d[i] - t[0]) <= tol &&
            Math.abs(d[i + 1] - t[1]) <= tol &&
            Math.abs(d[i + 2] - t[2]) <= tol
          ) {
            res[name]++;
          }
        }
      }
      return res;
    },
    {
      hexes: Object.fromEntries(Object.entries(targets).map(([k, v]) => [k, hexToRgb(v)])),
      tol: 14,
    },
  );
}

async function newPage(browser, opts) {
  const ctx = await browser.newContext({
    viewport: { width: opts.width, height: opts.height },
    locale: 'th-TH',
    timezoneId: 'Asia/Bangkok',
    reducedMotion: opts.reduced ? 'reduce' : 'no-preference',
    ...opts.device,
  });
  const page = await ctx.newPage();
  const errors = [];
  page.on('pageerror', (e) => errors.push(String(e)));
  page.on('console', (m) => {
    if (m.type() === 'error') errors.push(m.text());
  });
  await page.goto(`${BASE}/lookdev${opts.query ?? ''}`);
  await page.getByLabel('สรุปสถานการณ์น้ำ').waitFor({ timeout: 60_000 });
  await page.locator('main.lookdev[data-scene-ready="1"]').waitFor({ timeout: 60_000 });
  await page.waitForTimeout(4000);
  return { ctx, page, errors };
}

async function flyToStation(page) {
  await page.getByRole('button', { name: /ใกล้ฉัน/ }).click();
  await page.waitForFunction(() => !document.querySelector('.lookdev__flying'), null, {
    timeout: 60_000,
  });
  await page.waitForTimeout(2000);
}

/** WCAG contrast of a computed colour against a composited background. */
function ratio(fg, bg) {
  const lum = ([r, g, b]) => {
    const f = (v) => {
      const s = v / 255;
      return s <= 0.03928 ? s / 12.92 : ((s + 0.055) / 1.055) ** 2.4;
    };
    return 0.2126 * f(r) + 0.7152 * f(g) + 0.0722 * f(b);
  };
  const [a, b] = [lum(fg), lum(bg)].sort((x, y) => y - x);
  return (a + 0.05) / (b + 0.05);
}

const measure = { generatedAt: new Date().toISOString(), baseUrl: BASE };

const browser = await chromium.launch();

// --- Pixel shares: the landing frame and the national field ---
{
  const { ctx, page, errors } = await newPage(browser, { width: 1440, height: 900 });
  await flyToStation(page);
  const red = await canvasSample(page, { alert: TOKENS.level5 });
  const grey = await page.evaluate(() => {
    const src = document.querySelector('.maplibregl-canvas');
    const c = document.createElement('canvas');
    c.width = src.width;
    c.height = src.height;
    const ctx2 = c.getContext('2d', { willReadFrequently: true });
    ctx2.drawImage(src, 0, 0);
    const d = ctx2.getImageData(0, 0, c.width, c.height).data;
    // Protomaps DARK's own greys (#1f1f1f earth, #3d3d3d roads, #111 buildings).
    let n = 0;
    for (const g of [0x1f, 0x3d, 0x11]) {
      for (let i = 0; i < d.length; i += 4) {
        if (Math.abs(d[i] - g) <= 2 && Math.abs(d[i + 1] - g) <= 2 && Math.abs(d[i + 2] - g) <= 2)
          n++;
      }
    }
    return { protomapsGreyPx: n, total: d.length / 4 };
  });
  measure.landingDesktop = {
    alertRedPx: red.alert,
    alertRedPct: +((red.alert / red.sampled) * 100).toFixed(2),
    protomapsGreyPct: +((grey.protomapsGreyPx / grey.total) * 100).toFixed(2),
    consoleErrors: errors.length,
  };
  await ctx.close();
}
for (const [label, device] of [
  ['nationalDesktop', undefined],
  ['nationalPhone', devices['iPhone 13']],
]) {
  const { ctx, page, errors } = await newPage(browser, {
    width: label === 'nationalDesktop' ? 1440 : 390,
    height: label === 'nationalDesktop' ? 900 : 844,
    device,
  });
  const counts = await canvasSample(page, {
    level1: TOKENS.level1,
    level3: TOKENS.level3,
    level4: TOKENS.level4,
    level5: TOKENS.level5,
  });
  measure[label] = { ...counts, consoleErrors: errors.length };
  await ctx.close();
}

// --- Contrast: the elements the gate cites, computed styles against
// --- their effective (composited) backgrounds.
{
  const { ctx, page } = await newPage(browser, { width: 1440, height: 900 });
  const contrast = await page.evaluate(() => {
    const parse = (s) => {
      const m = /rgba?\((\d+), (\d+), (\d+)(?:, ([\d.]+))?\)/.exec(s);
      return m ? [+m[1], +m[2], +m[3], m[4] === undefined ? 1 : +m[4]] : null;
    };
    const over = (fg, bg) => fg.map((v, i) => (i < 3 ? v * fg[3] + bg[i] * (1 - fg[3]) : 1));
    const els = {
      panelSentence: '.lookdev__sentence',
      panelDt: '.lookdev__numbers dt',
      panelDd: '.lookdev__numbers dd',
      panelMeta: '.lookdev__meta',
      nearme: '.lookdev__nearme',
      nearmeNote: '.lookdev__nearme-note',
      preset: '.lookdev__preset',
      exaggLabel: '.lookdev__exagg',
      nogl: '.lookdev__nogl',
      stationSentence: '.lookdev__sentence--station',
      stationDd: '.lookdev__numbers--station dd',
      linehintP: '.lookdev__linehint p',
    };
    const out = {};
    for (const [name, sel] of Object.entries(els)) {
      const el = document.querySelector(sel);
      if (!el) continue;
      const cs = getComputedStyle(el);
      const fg = parse(cs.color);
      // Walk up until a non-transparent background is found.
      let node = el;
      let bg = [14, 22, 29, 1]; // --color-ground
      while (node && node !== document.documentElement) {
        const b = parse(getComputedStyle(node).backgroundColor);
        if (b && b[3] > 0) {
          bg = b;
          break;
        }
        node = node.parentElement;
      }
      if (fg)
        out[name] = {
          color: cs.color,
          background: getComputedStyle(node).backgroundColor,
          rgb: over(fg, bg),
        };
    }
    return out;
  });
  const ground = hexToRgb('#0e161d');
  const panel = hexToRgb('#15202b');
  measure.contrast = {};
  for (const [name, v] of Object.entries(contrast)) {
    const bg = v.rgb ? null : null;
    void bg;
    const isPanel = name.startsWith('panel') || name.startsWith('station');
    measure.contrast[name] = {
      color: v.color,
      ratioVsPanel: +ratio(v.rgb, panel).toFixed(2),
      ratioVsGround: +ratio(v.rgb, ground).toFixed(2),
    };
  }
  // The ใกล้ฉัน note in its hover state (the gate's one-time failure).
  const hover = await page.evaluate(() => {
    const el = document.querySelector('.lookdev__nearme-note');
    const cs = getComputedStyle(el);
    return cs.color;
  });
  const hoverBg = [31, 46, 60]; // --color-hover #1f2e3c
  measure.contrast.nearmeNoteHover = {
    color: hover,
    ratioVsHover: +ratio(hexToRgb('#8494a0'), hoverBg).toFixed(2),
    ratioAfterFixVsHover: +ratio(hexToRgb('#a4b4bf'), hoverBg).toFixed(2),
  };
  await ctx.close();
}

// --- Tap targets at 390x844: the resting chrome, the LINE hint (its
// --- UA), and the station close (after a flight).
{
  const { ctx, page } = await newPage(browser, {
    width: 390,
    height: 844,
    device: devices['iPhone 13'],
  });
  await flyToStation(page);
  measure.tapTargets = await page.evaluate(() => {
    const sels = {
      nearme: '.lookdev__nearme',
      preset: '.lookdev__preset',
      stationClose: '.lookdev__station-close',
    };
    const out = {};
    for (const [name, sel] of Object.entries(sels)) {
      const el = document.querySelector(sel);
      if (!el) continue;
      const r = el.getBoundingClientRect();
      out[name] = { w: Math.round(r.width), h: Math.round(r.height) };
    }
    return out;
  });
  await ctx.close();
}
{
  const ctx = await browser.newContext({
    ...devices['iPhone 13'],
    locale: 'th-TH',
    userAgent:
      'Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Line/13.20.0 Mobile/15E148 Safari/604.1',
  });
  const page = await ctx.newPage();
  await page.goto(`${BASE}/lookdev`);
  await page.getByLabel('สรุปสถานการณ์น้ำ').waitFor({ timeout: 60_000 });
  await page.waitForTimeout(1500);
  measure.tapTargets.linehintClose = await page.evaluate(() => {
    const el = document.querySelector('.lookdev__linehint-close');
    if (!el) return null;
    const r = el.getBoundingClientRect();
    return { w: Math.round(r.width), h: Math.round(r.height) };
  });
  await ctx.close();
}

await browser.close();

// --- Bundle budget (gzip), from the built dist ---
{
  const dir = resolve(import.meta.dirname, '../../web/dist/assets');
  // exFAT AppleDouble files (`._*`) are junk, not assets.
  const files = readdirSync(dir).filter((f) => /\.(js|mjs)$/.test(f) && !f.startsWith('._'));
  measure.bundleGzipKb = {};
  for (const f of files) {
    const buf = readFileSync(`${dir}/${f}`);
    const gz = gzipSync(buf).length;
    const kind = f.startsWith('index-')
      ? 'initialShell'
      : f.startsWith('LookdevMap-')
        ? 'lazyReliefChunk'
        : f.startsWith('maplibre-gl-worker-')
          ? 'worker'
          : 'other';
    measure.bundleGzipKb[kind] = {
      file: f,
      rawKb: +(statSync(`${dir}/${f}`).size / 1024).toFixed(1),
      gzipKb: +(gz / 1024).toFixed(1),
    };
  }
}

writeFileSync(OUT, JSON.stringify(measure, null, 2) + '\n');
console.log(`wrote ${OUT}`);
console.log(JSON.stringify(measure, null, 1));
