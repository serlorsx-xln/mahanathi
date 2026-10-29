// The single source of truth for the visual world's numbers. The CSS
// custom properties in tokens.css are GENERATED from this module by
// `pnpm tokens:css`; a test fails if the two drift apart. Shaders and
// canvas code import this module directly, so no colour exists outside
// it (docs/brief/01-DESIGN-DIRECTION.md, "single source of truth").
//
// The world: a night-time hydrographic instrument. Night slate, never
// #000; ochre sediment; muted paddy green; one disciplined alert.

/** Colour roles. Names describe the job, not the hue. */
export const colors = {
  // Ground and surfaces.
  ground: '#0e161d', // the map-room dark, blue-green slate
  sea: '#0d1a23', // calm dark water, slightly bluer than the ground
  panel: '#15202b', // answer panel, solid equivalent (the CSS one is a veil)
  panelVeil: 'rgba(21, 32, 43, 0.88)', // over the relief, with a functional blur
  elevated: '#1a2733', // sheets, menus, the legend expanded
  hover: '#1f2e3c',
  border: '#24313d',
  borderStrong: '#33434f',

  // Words.
  textPrimary: '#e6edf2',
  textSecondary: '#a4b4bf',
  textFaint: '#8494a0', // measured times, source credits; still AA on every surface

  // The two disciplined accents.
  accentWater: '#7fb5b0', // rivers, links, focus; muted teal, never neon
  sediment: '#caa46a', // the ochre of Chao Phraya floodwater; relief tint and selection

  // Browser surfaces.
  selectionBg: '#caa46a',
  selectionText: '#10161d',
  caret: '#7fb5b0',
  focusRing: '#7fb5b0',
  scrollbar: '#33434f',
  scrollbarTrack: '#0e161d',

  // Gauge hardware.
  gaugeSteel: '#aebac4', // the pole and the bank-full cap
  gaugeTrack: '#2a3742', // the quiet slot behind the fill
} as const;

/** The relief material, consumed by the 3D scene, not by CSS. */
export const relief = {
  albedoLow: '#8f8a74', // the central plain: muted paddy green over warm silt
  albedoHigh: '#a89e8c', // the ranges: plaster and rock, lighter
  neighbour: '#4d5860', // neighbours desaturated, receding
  fog: '#0d1a23', // aerial perspective; the sea and the haze share the air
} as const;

/**
 * The situation scale, redesigned (ThaiWater's five colours were neither
 * ordered nor harmonious). Only levels 4 and 5 get strong colour; level 1
 * is a drought concern in a muted slate blue, a different family from
 * the flood arm entirely. Level 2 is near-grey, level 3 a quiet green,
 * so the quiet third of the scale stays quiet. Every entry passes WCAG
 * AA as small text on ground, panel and elevated (tested in
 * tokens.test.ts).
 *
 * The scale was chosen by constrained search (scripts/research/
 * cvd_scale_search, recorded in the Phase 3 look-dev evidence) so that
 * under Machado deuteranopia and protanopia simulation every pair of
 * levels stays at least ~15 delta-E apart: reds and greens both collapse
 * toward olive, so the separation is carried by lightness, by
 * colorfulness (the grey level 2 against the saturated 4 and 5), and by
 * the blue axis, which survives both simulations and carries level 1.
 */
export const situations = [
  { level: 1, key: 'critically-low', label: 'น้ำน้อยวิกฤติ', hex: '#798fb4' },
  { level: 2, key: 'low', label: 'น้ำน้อย', hex: '#8b8f80' },
  { level: 3, key: 'normal', label: 'น้ำปกติ', hex: '#599c5b' },
  { level: 4, key: 'high', label: 'น้ำมาก', hex: '#e3a93f' },
  { level: 5, key: 'overflow', label: 'น้ำล้นตลิ่ง', hex: '#ee5d58' },
] as const;

/**
 * Stale stations are hollow: a quiet stroke, no fill. This is a graphic
 * stroke colour (3:1 against the ground, tested), not a text colour;
 * copy about stale data uses textFaint.
 */
export const stale = { key: 'stale', label: 'ยังไม่อัปเดต', hex: '#5b6a75' } as const;

export type Situation = (typeof situations)[number];

/**
 * Type. The pairing is the specimen's decision (ADR 0005): IBM Plex
 * Sans Thai for the UI, renamed "Mahanathi Sans" in the subsets as the
 * OFL requires, and Noto Serif Thai for the human sentence. Thai needs
 * the leading: body 1.6, the sentence more.
 */
export const type = {
  uiStack: "'Mahanathi Sans', 'Noto Sans Thai', sans-serif",
  sentenceStack: "'Noto Serif Thai', 'Trirong', serif",
  scale: {
    caption: { size: '0.75rem', lineHeight: 1.5 },
    small: { size: '0.875rem', lineHeight: 1.55 },
    body: { size: '1rem', lineHeight: 1.6 },
    titleSm: { size: '1.125rem', lineHeight: 1.5 },
    title: { size: '1.375rem', lineHeight: 1.5 },
    sentence: { size: 'clamp(1.375rem, 1rem + 2vw, 2.125rem)', lineHeight: 1.65 },
    display: { size: 'clamp(2.5rem, 1.5rem + 4vw, 4rem)', lineHeight: 1.3 },
    wordmark: { size: '1.25rem', lineHeight: 1.2 },
  },
  weightRegular: 400,
  weightSemibold: 600,
} as const;

/** Spacing on a 4 px base. */
export const space = {
  1: '0.25rem',
  2: '0.5rem',
  3: '0.75rem',
  4: '1rem',
  5: '1.5rem',
  6: '2rem',
  7: '3rem',
  8: '4rem',
} as const;

/** Radii are small: this is a precise instrument, not a soft app. */
export const radius = {
  sm: '2px',
  md: '4px',
  lg: '8px',
  sheet: '12px', // the bottom sheet on phones
  full: '999px', // swatch dots only
} as const;

/**
 * Motion: exponential ease-out everywhere. UI 160-260 ms, camera flights
 * 1.2-2.0 s, interruptible. Reduced motion cuts instead of flying
 * (handled by the scene, not by a token).
 */
export const motion = {
  fast: '160ms',
  base: '200ms',
  slow: '260ms',
  camera: '1600ms',
  easeOut: 'cubic-bezier(0.16, 1, 0.3, 1)',
} as const;

/** Stacking order, ground up. */
export const z = {
  ground: 0,
  annotation: 10, // projected HTML labels over the relief
  panel: 20,
  chrome: 30, // wordmark, search, ใกล้ฉัน, timeline, legend
  overlay: 40,
  toast: 50,
} as const;

/** One soft, offset shadow for Phase 4's elevated surfaces (menus, the
 *  expanded legend). No halos, no zero-blur blocks. Unused as of the
 *  look-dev gate: the /lookdev panel separates by veil and border. */
export const shadow = {
  panel: '0 4px 12px rgba(4, 8, 12, 0.4), 0 16px 48px rgba(4, 8, 12, 0.5)',
} as const;

// ---------------------------------------------------------------------------
// WCAG contrast, computed here so the tests and the look-dev evidence
// use the same numbers.
// ---------------------------------------------------------------------------

function srgbToLinear(c: number): number {
  return c <= 0.04045 ? c / 12.92 : Math.pow((c + 0.055) / 1.055, 2.4);
}

/** Relative luminance of an sRGB hex string. */
export function luminance(hex: string): number {
  const h = hex.replace('#', '');
  const r = srgbToLinear(parseInt(h.slice(0, 2), 16) / 255);
  const g = srgbToLinear(parseInt(h.slice(2, 4), 16) / 255);
  const b = srgbToLinear(parseInt(h.slice(4, 6), 16) / 255);
  return 0.2126 * r + 0.7152 * g + 0.0722 * b;
}

/** WCAG 2.x contrast ratio of two hex colours. */
export function contrast(a: string, b: string): number {
  const la = luminance(a);
  const lb = luminance(b);
  return (Math.max(la, lb) + 0.05) / (Math.min(la, lb) + 0.05);
}

// ---------------------------------------------------------------------------
// The generated CSS mirror. `pnpm tokens:css` writes tokens.css from
// this; tokens.test.ts fails if the file is stale.
// ---------------------------------------------------------------------------

function prop(name: string, value: string | number): string {
  return `  --${name}: ${value};`;
}

export function renderCss(): string {
  const lines: string[] = [
    '/* GENERATED from web/src/design/tokens.ts by `pnpm tokens:css` - do not edit. */',
    '',
    ':root {',
  ];
  for (const [k, v] of Object.entries(colors)) {
    lines.push(prop(`color-${kebab(k)}`, v));
  }
  for (const [k, v] of Object.entries(relief)) {
    lines.push(prop(`relief-${kebab(k)}`, v));
  }
  for (const s of situations) {
    lines.push(prop(`situation-${s.level}`, s.hex));
  }
  lines.push(prop('situation-stale', stale.hex));
  lines.push(prop('font-ui', type.uiStack));
  lines.push(prop('font-sentence', type.sentenceStack));
  for (const [k, v] of Object.entries(type.scale)) {
    lines.push(prop(`type-${kebab(k)}-size`, v.size));
    lines.push(prop(`type-${kebab(k)}-line-height`, String(v.lineHeight)));
  }
  lines.push(prop('type-weight-regular', String(type.weightRegular)));
  lines.push(prop('type-weight-semibold', String(type.weightSemibold)));
  for (const [k, v] of Object.entries(space)) {
    lines.push(prop(`space-${k}`, v));
  }
  for (const [k, v] of Object.entries(radius)) {
    lines.push(prop(`radius-${k}`, v));
  }
  for (const [k, v] of Object.entries(motion)) {
    lines.push(prop(`motion-${kebab(k)}`, v));
  }
  for (const [k, v] of Object.entries(z)) {
    lines.push(prop(`z-${k}`, String(v)));
  }
  for (const [k, v] of Object.entries(shadow)) {
    lines.push(prop(`shadow-${kebab(k)}`, v));
  }
  lines.push('}');
  return lines.join('\n') + '\n';
}

function kebab(s: string): string {
  return s.replace(/([a-z0-9])([A-Z])/g, '$1-$2').toLowerCase();
}
