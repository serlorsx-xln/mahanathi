import { mkdirSync, writeFileSync } from 'node:fs';
import { resolve } from 'node:path';
import { test } from '@playwright/test';

// Measured evidence for the Phase 3 type decision, written as JSON
// beside the screenshots. The implementing session cannot view images
// (docs/DEVLOG.md, 23:25), so the specimen's visual questions are
// answered with numbers: ink extents from canvas TextMetrics (the
// Anuphan clipping question), typeset width of the national sentence
// per face (density), and figure behaviour (tabularity, minus sign).
const OUT_DIR = resolve(import.meta.dirname, '../../docs/design/screens/phase-3');

const NATIONAL = 'ตอนนี้ 18 สถานีน้ำล้นตลิ่ง ส่วนใหญ่อยู่ที่นครปฐม กาญจนบุรี และสุพรรณบุรี';

const UI_FAMILIES = [
  ['Mahanathi Sans (IBM Plex Sans Thai)', "'Mahanathi Sans', sans-serif"],
  ['Anuphan', "'Anuphan', sans-serif"],
  ['Noto Sans Thai', "'Noto Sans Thai', sans-serif"],
] as const;

const SENTENCE_FAMILIES = [
  ['sans 600', null],
  ['Noto Serif Thai', "'Noto Serif Thai', serif"],
  ['Pridi', "'Pridi', serif"],
  ['Trirong', "'Trirong', serif"],
] as const;

// The strings whose stacked marks decide the leading and clipping
// questions (docs/research/typography.md, "Specimen must test").
const INK_STRINGS = ['น้ำล้นตลิ่ง', 'ผู้ใหญ่บ้าน', 'ปิ่นเกล้า', 'ปั๊มน้ำ', 'ญี่ปุ่น'];

test('specimen measurements', async ({ page }) => {
  test.skip(test.info().project.name !== 'desktop', 'measured once, on desktop');

  await page.goto('/specimen');
  await page.getByRole('heading', { name: 'ตัวอย่างตัวอักษร' }).waitFor();
  await page.evaluate(() => document.fonts.ready);

  const measurements = await page.evaluate(
    ({ national, uiFamilies, sentenceFamilies, inkStrings }) => {
      const out = {
        fontsLoaded: {} as Record<string, Record<string, boolean>>,
        ink: {} as Record<
          string,
          Record<string, { ascentPct: number; descentPct: number; lineHeight16Use: number }>
        >,
        sentence: {} as Record<string, { widthEm: number }>,
        figures: {} as Record<string, { digit: number; digitsDiffer: number[] }>,
        dom: [] as {
          label: string | null;
          stackClipped: { scrollOverflowPx: number; clientHeight: number }[];
        }[],
      };

      const canvas = document.createElement('canvas');
      const ctx = canvas.getContext('2d');
      if (!ctx) throw new Error('no 2d context');
      const PX = 100; // measure at 100px so extents read as percents of em

      for (const [name, family] of uiFamilies) {
        out.fontsLoaded[name] = {
          400: document.fonts.check(`16px ${family.split(',')[0]}`, 'ม'),
          600: document.fonts.check(`600 16px ${family.split(',')[0]}`, 'ม'),
        };

        // Ink extents at body size: line-height 1.6 means the line box
        // is 160px per 100px of em. Ascent+descent beyond that clips.
        ctx.font = `400 ${PX}px ${family}`;
        out.ink[name] = {};
        for (const s of inkStrings) {
          const m = ctx.measureText(s);
          out.ink[name][s] = {
            ascentPct: +(m.actualBoundingBoxAscent / PX).toFixed(3),
            descentPct: +(m.actualBoundingBoxDescent / PX).toFixed(3),
            // ink height as a fraction of the 1.6 line box
            lineHeight16Use: +(
              (m.actualBoundingBoxAscent + m.actualBoundingBoxDescent) /
              (PX * 1.6)
            ).toFixed(3),
          };
        }

        // Figure behaviour: tabularity and the minus sign.
        out.figures[name] = {
          digit: ctx.measureText('3').width,
          digitsDiffer: [
            ctx.measureText('0').width,
            ctx.measureText('1').width,
            ctx.measureText('3').width,
            ctx.measureText('7').width,
            ctx.measureText('14:00').width,
            ctx.measureText('3.42').width,
            ctx.measureText('−0.30').width,
            ctx.measureText('+0.39').width,
          ].map((w) => +(w / PX).toFixed(3)),
        };

        // Sentence width per face, in em: the density comparison.
        for (const [sName, sFamily] of sentenceFamilies) {
          ctx.font = sFamily ? `400 ${PX}px ${sFamily}` : `600 ${PX}px ${family}`;
          const key = `${name} / ${sName}`;
          out.sentence[key] = {
            widthEm: +(ctx.measureText(national).width / PX).toFixed(3),
          };
        }
      }
      return out;
    },
    {
      national: NATIONAL,
      uiFamilies: UI_FAMILIES,
      sentenceFamilies: SENTENCE_FAMILIES,
      inkStrings: INK_STRINGS,
    },
  );

  // DOM-level: does any stacked-mark line overflow its hidden box?
  measurements.dom = await page.evaluate(() => {
    const pairs = Array.from(document.querySelectorAll<HTMLElement>('.specimen__pairing'));
    return pairs.map((p) => ({
      label: p.querySelector('h2')?.textContent ?? null,
      stackClipped: Array.from(p.querySelectorAll<HTMLElement>('.specimen__stack')).map((el) => ({
        scrollOverflowPx: el.scrollHeight - el.clientHeight,
        clientHeight: el.clientHeight,
      })),
    }));
  });

  mkdirSync(OUT_DIR, { recursive: true });
  writeFileSync(
    resolve(OUT_DIR, 'specimen-measurements.json'),
    JSON.stringify(measurements, null, 2),
  );
});
