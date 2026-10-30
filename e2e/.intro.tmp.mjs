import { chromium } from '@playwright/test';
const b = await chromium.launch();
const page = await b.newPage({ viewport: { width: 1440, height: 900 }, locale: 'th-TH', timezoneId: 'Asia/Bangkok' });
const errors = [];
page.on('pageerror', (e) => errors.push(String(e)));
page.on('console', (m) => { if (m.type() === 'error') errors.push(m.text()); });
await page.goto('http://localhost:4173/lookdev');
const early = await page.evaluate(() => ({
  intro: document.querySelector('.lookdev__intro')?.dataset.intro ?? null,
  wordmark: document.querySelector('.lookdev__intro-wordmark')?.textContent ?? null,
  ready: document.querySelector('main.lookdev')?.dataset.sceneReady ?? null,
}));
await page.locator('main.lookdev[data-scene-ready="1"]').waitFor({ timeout: 60000 });
const atReady = await page.evaluate(() => ({
  intro: document.querySelector('.lookdev__intro')?.dataset.intro ?? null,
}));
await page.waitForFunction(() => !document.querySelector('.lookdev__intro'), null, { timeout: 10000 });
const after = await page.evaluate(() => ({
  introGone: !document.querySelector('.lookdev__intro'),
  pillEnabled: [...document.querySelectorAll('.lookdev__view')].every((x) => !x.disabled),
  nearmeEnabled: !document.querySelector('.lookdev__nearme')?.disabled,
}));
console.log(JSON.stringify({ early, atReady, after, errors }, null, 1));
const ctx = await b.newContext({ viewport: { width: 1440, height: 900 }, locale: 'th-TH', reducedMotion: 'reduce' });
const rp = await ctx.newPage();
await rp.goto('http://localhost:4173/lookdev');
await rp.locator('main.lookdev[data-scene-ready="1"]').waitFor({ timeout: 60000 });
await rp.waitForTimeout(800);
const rm = await rp.evaluate(() => ({ introGone: !document.querySelector('.lookdev__intro') }));
console.log(JSON.stringify({ reducedMotion: rm }, null, 1));
await b.close();
