/// <reference types="node" />
import { readFileSync } from 'node:fs';
import { join } from 'node:path';
import { describe, expect, it } from 'vitest';
import { colors, contrast, renderCss, situations, stale } from './tokens';

const BODY_MIN = 4.5; // WCAG AA, body text
const LARGE_MIN = 3; // WCAG AA, large text and UI glyphs

const textSurfaces = [
  ['ground', colors.ground],
  ['panel', colors.panel],
  ['elevated', colors.elevated],
] as const;

const textRoles = [
  ['textPrimary', colors.textPrimary],
  ['textSecondary', colors.textSecondary],
  ['textFaint', colors.textFaint],
  ['accentWater', colors.accentWater],
] as const;

describe('tokens.css is the generated mirror of tokens.ts', () => {
  it('matches renderCss() exactly', () => {
    const onDisk = readFileSync(join(import.meta.dirname, 'tokens.css'), 'utf8');
    expect(onDisk).toBe(renderCss());
  });
});

describe('every text role passes AA on every text surface', () => {
  for (const [role, hex] of textRoles) {
    for (const [surface, bg] of textSurfaces) {
      it(`${role} on ${surface} is at least ${BODY_MIN}:1`, () => {
        expect(contrast(hex, bg)).toBeGreaterThanOrEqual(BODY_MIN);
      });
    }
  }
});

describe('every situation colour is legible as text on every surface', () => {
  // Stale is excluded: it is the stroke of a hollow gauge, a graphic
  // token (3:1 against ground, below), never a text colour. Text about
  // stale data uses textFaint.
  for (const s of situations) {
    for (const [surface, bg] of textSurfaces) {
      it(`level ${s.level} (${s.label}) on ${surface} is at least ${BODY_MIN}:1`, () => {
        expect(contrast(s.hex, bg)).toBeGreaterThanOrEqual(BODY_MIN);
      });
    }
  }
});

describe('non-text UI tokens pass the 3:1 minimum', () => {
  it('focusRing against ground', () => {
    expect(contrast(colors.focusRing, colors.ground)).toBeGreaterThanOrEqual(LARGE_MIN);
  });
  it('gaugeSteel against ground (the bank-full cap must be visible)', () => {
    expect(contrast(colors.gaugeSteel, colors.ground)).toBeGreaterThanOrEqual(LARGE_MIN);
  });
  it('stale stroke against ground', () => {
    expect(contrast(stale.hex, colors.ground)).toBeGreaterThanOrEqual(LARGE_MIN);
  });
  it('selectionText on selectionBg', () => {
    expect(contrast(colors.selectionText, colors.selectionBg)).toBeGreaterThanOrEqual(BODY_MIN);
  });
});

describe('the situation scale is ordered and restrained', () => {
  it('has exactly the five ThaiWater levels in order', () => {
    expect(situations.map((s) => s.level)).toEqual([1, 2, 3, 4, 5]);
    for (const s of situations) {
      expect(s.label).toMatch(/^[ก-๙][ก-๙\s]+$/);
    }
  });

  it('only levels 4 and 5 are strongly saturated', () => {
    const sat = (hex: string) => {
      const h = hex.replace('#', '');
      const parts = [0, 2, 4].map((i) => parseInt(h.slice(i, i + 2), 16));
      const max = Math.max(...parts);
      const min = Math.min(...parts);
      return max === 0 ? 0 : (max - min) / max;
    };
    for (const s of situations) {
      if (s.level <= 3) {
        expect(sat(s.hex)).toBeLessThan(0.45);
      } else {
        expect(sat(s.hex)).toBeGreaterThan(0.5);
      }
    }
  });
});
