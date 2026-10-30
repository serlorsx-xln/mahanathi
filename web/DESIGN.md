---
name: มหานที (Mahanathi)
description: A night-time hydrographic instrument — every river gauge in Thailand, live, on a dark relief of the country.
colors:
  ground: "#0e161d"
  sea: "#0d1a23"
  panel: "#15202b"
  panel-veil: "rgba(21, 32, 43, 0.88)"
  elevated: "#1a2733"
  hover: "#1f2e3c"
  border: "#24313d"
  border-strong: "#33434f"
  text-primary: "#e6edf2"
  text-secondary: "#a4b4bf"
  text-faint: "#8494a0"
  accent-water: "#7fb5b0"
  sediment: "#caa46a"
  selection-bg: "#caa46a"
  selection-text: "#10161d"
  caret: "#7fb5b0"
  focus-ring: "#7fb5b0"
  scrollbar: "#33434f"
  scrollbar-track: "#0e161d"
  gauge-steel: "#aebac4"
  gauge-track: "#2a3742"
  relief-albedo-low: "#8f8a74"
  relief-albedo-high: "#a89e8c"
  relief-neighbour: "#4d5860"
  relief-fog: "#0d1a23"
  situation-1: "#798fb4"
  situation-2: "#8b8f80"
  situation-3: "#599c5b"
  situation-4: "#e3a93f"
  situation-5: "#ee5d58"
  situation-stale: "#5b6a75"
typography:
  display:
    fontFamily: "'Mahanathi Sans', 'Noto Sans Thai', sans-serif"
    fontSize: "clamp(2.5rem, 1.5rem + 4vw, 4rem)"
    lineHeight: 1.3
  sentence:
    fontFamily: "'Mahanathi Sans', 'Noto Sans Thai', sans-serif"
    fontWeight: 400
    fontSize: "clamp(1.375rem, 1rem + 2vw, 2.125rem)"
    lineHeight: 1.65
  title:
    fontFamily: "'Mahanathi Sans', 'Noto Sans Thai', sans-serif"
    fontWeight: 600
    fontSize: "1.375rem"
    lineHeight: 1.5
  title-sm:
    fontFamily: "'Mahanathi Sans', 'Noto Sans Thai', sans-serif"
    fontWeight: 600
    fontSize: "1.125rem"
    lineHeight: 1.5
  body:
    fontFamily: "'Mahanathi Sans', 'Noto Sans Thai', sans-serif"
    fontWeight: 400
    fontSize: "1rem"
    lineHeight: 1.6
  small:
    fontFamily: "'Mahanathi Sans', 'Noto Sans Thai', sans-serif"
    fontSize: "0.875rem"
    lineHeight: 1.55
  caption:
    fontFamily: "'Mahanathi Sans', 'Noto Sans Thai', sans-serif"
    fontSize: "0.75rem"
    lineHeight: 1.5
  wordmark:
    fontFamily: "'Mahanathi Sans', 'Noto Sans Thai', sans-serif"
    fontWeight: 400
    fontSize: "1.25rem"
    lineHeight: 1.2
rounded:
  sm: "2px"
  md: "4px"
  lg: "8px"
  sheet: "12px"
  full: "999px"
spacing:
  "1": "0.25rem"
  "2": "0.5rem"
  "3": "0.75rem"
  "4": "1rem"
  "5": "1.5rem"
  "6": "2rem"
  "7": "3rem"
  "8": "4rem"
components:
  panel:
    backgroundColor: "{colors.panel-veil}"
    textColor: "{colors.text-primary}"
    rounded: "{rounded.lg}"
    padding: "1.5rem"
  panel-mobile:
    padding: "1rem"
  button-primary:
    backgroundColor: "{colors.panel}"
    textColor: "{colors.text-primary}"
    typography: "{typography.small}"
    rounded: "{rounded.lg}"
    padding: "0.5rem 1rem"
  button-primary-hover:
    backgroundColor: "{colors.hover}"
  station-sheet:
    backgroundColor: "{colors.panel-veil}"
    textColor: "{colors.text-primary}"
    rounded: "{rounded.sheet}"
    padding: "1rem 1.5rem"
  chip-preset:
    backgroundColor: "{colors.panel}"
    textColor: "{colors.text-secondary}"
    typography: "{typography.caption}"
    rounded: "{rounded.full}"
    padding: "0.25rem 0.75rem"
  chip-preset-active:
    textColor: "{colors.text-primary}"
  error-card:
    backgroundColor: "{colors.panel}"
    textColor: "{colors.text-primary}"
    rounded: "{rounded.lg}"
    padding: "1.5rem"
  station-label:
    backgroundColor: "{colors.panel}"
    textColor: "{colors.text-primary}"
    typography: "{typography.caption}"
    rounded: "{rounded.md}"
    padding: "0.25rem 0.5rem"
---

# Design System: มหานที (Mahanathi)

## Overview

**Creative North Star: "The Night-Time Hydrographic Instrument"**

Picture the national survey's plaster relief model of Thailand in a dark map room, lit only where water is being measured right now. Everything visible is either geography, data, or the words that explain the data. The relief is the page: there is no landing page, no hero section, no feature cards, no footer — the first frame *is* the product (Thailand in relief, the current state of every gauge, one computed sentence about the country now). Chrome is nearly absent, in the earth.nullschool sense: a wordmark, one primary action (**ใกล้ฉัน**), and quiet panels that float over the terrain.

Restraint is the luxury. The palette comes from the subject — night slate (never `#000`), the ochre of Chao Phraya floodwater, muted paddy green, one disciplined alert family — and every colour lives in one TypeScript module (`web/src/design/tokens.ts`) that generates the CSS custom properties (`tokens.css`, via `pnpm tokens:css`, with a drift test). Shaders import the module directly, so no colour exists outside it. The 3D pipeline is deliberately flat (`NoToneMapping` Canvas, explicit output-colour-space chunks in both custom shaders) so a token colour that survives lighting reaches the screen exactly: the national screenshot samples the sea and ground back at their token hexes (ADR 0006). Type is Thai-first and loopless: Mahanathi Sans (IBM Plex Sans Thai, OFL-renamed, 400/600) for everything — the UI, the numbers and the human sentence (ADR 0005 and its loopless addendum; the serif is retired). Motion is exponential ease-out, 160–260 ms for UI, one 1600 ms interruptible camera flight as the single authored moment; `prefers-reduced-motion` cuts instead of flying.

This system was proved on the dev surfaces `/lookdev`, `/specimen` and `/color` (Phase 3), gated item-by-item against the anti-slop checklist in `docs/brief/01-DESIGN-DIRECTION.md` §10 with screenshot evidence in `docs/design/screens/phase-3/` and the scored pass in `docs/design/GATE-phase-3.md`. Phase 4 extends this world; it must not average it away.

**Key Characteristics:**
- Dark theme only; night slate ground (`#0e161d`), never pure black, never ultra-low-contrast greys (must survive a bright demo room).
- The relief is the page; UI floats over it as veils with 1px borders, not cards with shadows.
- Two disciplined accents: muted water teal and ochre sediment. Strong colour belongs to data levels 4–5 only.
- The situation scale is ordered, CVD-searched (Machado deuteranopia/protanopia, pairwise ΔE ≥ ~15), and geometrically redundant — the gauge encoding survives colour-blind vision and monochrome alike.
- Thai-first typography with generous leading (stacked vowels and tone marks), tabular lining figures on every number, no italic anywhere.
- One functional `backdrop-filter: blur(14px)` pair (answer panel, station sheet) for legibility over terrain; zero `box-shadow` declarations in shipped route CSS.
- Camera moves only on input, presets, or the one flight; the renderer idles (`frameloop="demand"`).

## Colors

The palette is built from the subject — night, sediment, paddy — and locked in `tokens.ts`; the hexes below are that module's values (the frontmatter is normative).

### Primary
- **Muted Water Teal** (`accent-water`, `#7fb5b0`): rivers on the relief, links, the focus ring, the caret, and the range-input accent. The instrument's "water is here" colour — never neon cyan.
- **Sediment Ochre** (`sediment`, `#caa46a`): the colour of Chao Phraya floodwater. Text selection background (with `selection-text` `#10161d` on it) and the relief's warm tint family. A close second accent, used sparingly.

### Tertiary (the situation scale — data, not decoration)
- **Level 1, Critically Low** (`situation-1`, `#798fb4`): a drought concern in a muted slate blue — a different family from the flood arm entirely, carried by the blue axis that survives both CVD simulations.
- **Level 2, Low** (`situation-2`, `#8b8f80`): near-grey. The quiet third of the scale stays quiet.
- **Level 3, Normal** (`situation-3`, `#599c5b`): a quiet green.
- **Level 4, High** (`situation-4`, `#e3a93f`): amber. Strong colour begins here.
- **Level 5, Overflow** (`situation-5`, `#ee5d58`): red. The alert colour; also the gauge fill above the bank-full cap.
- **Stale** (`situation-stale`, `#5b6a75`): a graphic stroke colour (3:1 against ground), never a fill and never a text colour; copy about stale data uses `text-faint`.

Every situation colour passes WCAG AA as small text on ground, panel and elevated (tested in `tokens.test.ts`); the scale was chosen by constrained search (`scripts/research/cvd_scale_search.py`) so every pair stays ≥ ~15 ΔE under Machado deuteranopia and protanopia simulation, replayed live at `/color`.

### Neutral
- **Night Slate Ground** (`ground`, `#0e161d`): the map-room dark; page background and clear colour.
- **Calm Sea** (`sea`, `#0d1a23`): slightly bluer than the ground; the sea and the aerial-perspective fog share the air.
- **Panel** (`panel`, `#15202b`) and **Panel Veil** (`panel-veil`, `rgba(21, 32, 43, 0.88)`): the answer panel's solid equivalent and its over-the-relief veil.
- **Elevated** (`elevated`, `#1a2733`) and **Hover** (`hover`, `#1f2e3c`): sheets, menus, the expanded legend; and the hover step.
- **Borders** (`border`, `#24313d`; `border-strong`, `#33434f`): 1px separations that do the work shadows would do elsewhere; `border-strong` is also the draped border line on the relief.
- **Text** (`text-primary`, `#e6edf2`; `text-secondary`, `#a4b4bf`; `text-faint`, `#8494a0`): primary reading, subordinate data, and measured times / source credits (faint stays AA on every surface, including the panel veil over the brightest terrain: 12.3:1 / 6.8:1 / 4.7:1, ADR 0006).
- **Gauge hardware** (`gauge-steel`, `#aebac4`; `gauge-track`, `#2a3742`): the pole and bank-full cap; the quiet slot behind the fill.
- **Relief material** (`relief-albedo-low`, `#8f8a74`; `relief-albedo-high`, `#a89e8c`; `relief-neighbour`, `#4d5860`; `relief-fog`, `#0d1a23`): the paddy-plain and range albedoes of the hypsometric tint, the desaturated neighbour tone, and the fog. Consumed by the 3D scene, not by CSS.
- **Browser surfaces** (`selection-bg` / `selection-text`, `caret`, `focus-ring`, `scrollbar` / `scrollbar-track`): themed in `base.css` — selection is ochre on near-black, focus ring and caret are the water teal, the scrollbar is thin on our slate, `color-scheme: dark`.

### Named Rules
**The Single Source Rule.** Every colour lives in `web/src/design/tokens.ts`. `tokens.css` is generated from it (`pnpm tokens:css`) and a test fails on drift; shaders import the module. A hard-coded colour in a component, stylesheet or shader does not ship.

**The Flat Pipeline Rule.** The Canvas runs `NoToneMapping` (`flat`) with an opaque framebuffer, and both custom fragments close with three's `tonemapping_fragment` / `colorspace_fragment` chunks, so token colours round-trip to the screen exactly. Any new material must include the output chunks or it renders dark (the mistake is recorded in DEVLOG 2026-10-30 ~00:20 and guarded by a screenshot-sampling check).

**The Quiet Third Rule.** Only levels 4 and 5 get strong colour. Levels 1–3 stay muted — the scale's CVD separation is carried by lightness, colorfulness (grey level 2 against saturated 4–5) and the blue axis, not by making everything loud.

**The Hollow Stale Rule.** Stale is a stroke, never a fill, everywhere: the 3D gauge is a hollow cap stroke, the legend dot is a hollow ring. Stale must never read as a level.

## Typography

**UI and Numbers Font:** Mahanathi Sans (IBM Plex Sans Thai 400/600, self-hosted hinted WOFF2 subsets; fallback `'Noto Sans Thai', sans-serif`)
**Sentence Font:** Mahanathi Sans 400 — the sentence's hierarchy is size, leading and the tone colour, never family. Noto Serif Thai carried the sentence until the loopless rule retired it (ADR 0005 addendum).
**Label/Mono Font:** none — all chosen figures are already tabular; mono was considered and rejected (ADR 0005).

**Character:** one engineered loopless Thai sans with a harmonised Latin (one team drew both scripts) carries everything — the instrument's voice and the human sentence's. The sentence once had a second, quieter voice in an upright Thai serif; the owner's all-Thai-loopless rule (no หัว) retired it, and the hierarchy it carried now lives in size, leading and tone colour alone. The wordmark มหานที is the sans at weight 400. One family, no italic anywhere (every `@font-face` is `font-style: normal`).

### Hierarchy
- **Display** (clamp(2.5rem → 4rem), 1.3): the scale's ceiling — display caps at 4rem against the brief's 6rem maximum.
- **Sentence** (Mahanathi Sans 400, clamp(1.375rem, 1rem + 2vw, 2.125rem), 1.65): the human sentence, the typographic centrepiece of the answer panel; `text-wrap: balance`. Leading 1.65 because Thai stacked marks need room.
- **Title** (600, 1.375rem, 1.5) and **Title-sm** (600, 1.125rem, 1.5): station names and section heads.
- **Body** (400, 1rem, 1.6): the default; Plex's worst measured ink box uses 84% of a 1.6 line box (ADR 0005).
- **Small** (0.875rem, 1.55) and **Caption** (0.75rem, 1.5): subordinate readouts, meta lines, measured time and source credits.
- **Wordmark** (serif 400, 1.25rem, 1.2): มหานที, set with care, no icon logo.

### Named Rules
**The Tabular Numbers Rule.** Every measurement, count and timestamp is set with `font-variant-numeric: tabular-nums lining-nums` — signed columns must align (Plex was chosen partly because its minus equals its plus in width, ADR 0005).

**The Thai Stays in the DOM Rule.** Thai text never enters WebGL; labels over the relief are DOM elements projected from 3D positions (drei `Html`), so stacked tone marks and vowels are shaped by the browser.

Fonts are self-hosted WOFF2 subsets (Thai + Latin), `font-display: swap`, ≈64 KB total; subsets renamed "Mahanathi Sans" per the OFL's Reserved Font Name condition. New strings need a subset rebuild (`scripts/research/promote_fonts.py`).

## Layout

The scene surface is a fixed full-viewport canvas (`position: fixed; inset: 0`) — the relief fills the frame and everything else floats over it. Thailand is tall: portrait phones fit it naturally; on desktop the country composes off-centre. Stacking is tokenised (ground 0, annotation 10 for projected labels, panel 20, chrome 30, overlay 40, toast 50). Spacing runs on a 4px base (`space-1` = 0.25rem through `space-8` = 4rem); steps are chosen by role — chrome padding `space-4/5`, panel `space-5`, sheet `space-4/5` — so rhythm steps rather than repeating monotonously.

- **Top chrome:** wordmark left, the primary action ใกล้ฉัน right.
- **Answer panel:** top-left over the relief, max-width 24rem, veil + blur + 1px border; on phones (<768px) it becomes a full-width top sheet so the sentence leads. The sentence first, then a live-count `<dl>` of the situation scale (dots + labels + counts), then the measured time and source line.
- **Station sheet:** bottom sheet on phones (left/right `space-3`, radius `sheet`); at ≥768px a 22rem card anchored bottom-right.
- **Bottom chrome:** dev presets and the exaggeration slider on `/lookdev` (not shipping); the product's timeline and legend live here in Phase 4, with `env(safe-area-inset-bottom)` respected.
- **Dev evidence pages** (`/specimen`, `/color`): one quiet column, max-width 46rem, ruled sections (`border-top`), no cards.

Gate-measured at 390×844: the national panel takes 34.8% of viewport height, top chrome 8.6%, leaving the relief ~44% clear (`docs/design/screens/phase-3/lookdev-mobile-coverage.json`).

## Elevation & Depth

Depth is tonal, not shadowed. Surfaces rise by colour — ground → sea → panel → elevated → hover — and separate by 1px borders; the panel veil (88% panel over the relief, plus the one functional blur) is measured legible over the brightest terrain the relief can put under it (ADR 0006). Zero `box-shadow` declarations exist in shipped route CSS; one soft, offset shadow token (`shadow.panel`: `0 4px 12px rgba(4, 8, 12, 0.4), 0 16px 48px rgba(4, 8, 12, 0.5)`) is reserved for Phase 4's elevated surfaces (menus, the expanded legend) — no halos, no zero-blur blocks.

In the 3D scene, depth is Imhof-style: one north-west key light (`normalize(-0.52, 0.72, -0.45)`) plus a soft fill, under an explicit moonlit Lambert scale (`0.22 + 0.55·key + 0.16·fill` — nothing black, nothing at full albedo, and the only brightness control since the pipeline is flat); a quiet hypsometric tint graded over 0–2200 m; paddy toning on measured flatness; neighbours desaturated via the provinces mask; and aerial perspective — `smoothstep(900, 2600)` km toward the sea's colour, 85% at most — that keeps the far north and the neighbours receding while the centre stays legible. Vertical exaggeration defaults to 12× (ADR 0006; dev slider 4–40×); the plain's flatness is invariant under it. No bloom, no postprocessing, no starfield, no particles, no glow.

### Named Rules
**The One Blur Rule.** `backdrop-filter: blur(14px)` exists on exactly two surfaces — the answer panel and the station sheet — both for legibility over terrain. Dev chrome sits on solid panel precisely to keep this honest. Glass is never decoration.

## Shapes

A precise instrument, not a soft app: radii are small and few — 2px (sm), 4px (md), 8px (lg), and 12px (sheet) for the phone bottom sheet. `full` (999px) is for the situation swatch dots; the dev-only preset pills borrow it and do not ship. Form language is rectangle-plus-1px-border everywhere; the only circles in the system are data (swatch dots) and the gauges' square-section poles. Lines on the relief are hairlines: the Thai border at width 1.1 in `border-strong` at 90% opacity; rivers in the water teal at 55% opacity with width by river order (Mekong 3.0, Chao Phraya 2.4, down to 1.4; ADR 0004's one-line-style borders).

## Components

### The Relief (signature surface)
A single displaced mesh from a 1,815 × 2,048 heightmap (~1 km/texel) over 92–111°E, 3–24°N, custom `ShaderMaterial` (ADR 0004/0006). The sea is an early-out: dark, calm, never shaded as rock. Vertex shader lifts by squared elevation codes; fragment shader does everything listed under Elevation & Depth. `antialias: true`, `dpr` clamped `[1, 2]`, `frameloop="demand"`, `fov 42`.

### The Gauges (signature component)
804 stations, one `InstancedMesh`, one draw call. Each gauge is a slim vertical pole (2.6 km square section) seated on the sampled terrain, with a geometric bank-full encoding that survives colour-blind vision and monochrome printing: the cap sits at 0.78 of the pole height; the water fill rises inside it in the station's situation colour; when the river is over its banks the fill rises *above* the cap in the level-5 red; stale stations are a hollow cap stroke (steel at 0.75) with everything else discarded. Pole heights encode hierarchy by level — 44 km (level 5), 34 (level 4), 17 (levels 1–3), 12 (no level) — so 800 gauges never read as identical glowing pillars. Zoom-in labels are Phase 4.

### Answer Panel
The veil surface: `panel-veil` background, `blur(14px)`, 1px `border`, radius `lg`, padding `space-5` (`space-4` on phones). Sentence first (serif, `text-wrap: balance`), then the situation-count `<dl>` (caption labels with 0.5rem swatch dots, small tabular counts), then the meta line (caption, faint): measured time and source, always.

### Station Sheet
Same veil recipe, radius `sheet`, bottom sheet on phones / 22rem bottom-right card at ≥768px. Station name (title-sm, 600), the station sentence (serif at title-sm size, 1.6), and a 2-column tabular `<dl>`: ระดับน้ำ, ระดับตลิ่ง, % ของตลิ่ง, วัดเมื่อ — missing values render as `·`, never as invented zeros.

### Buttons
- **Shape:** radius `lg` (8px) for the primary; `md` (4px) for the error retry.
- **Primary (ใกล้ฉัน):** `panel` background, `text-primary`, semibold small type, 1px `border-strong`, padding `space-2 space-4`; hover swaps to `hover` in 200ms ease-out.
- **Focus:** global — 2px solid `focus-ring`, 2px offset, visible only.
- **Ghost/secondary:** the preset chips (dev): `panel` background, `text-secondary`, radius `full`, padding `space-1 space-3`; active state is `text-primary` + `border-strong` border, no fill change.

### States (designed, not defaulted)
- **Loading:** a full-bleed skeleton on `ground` with a barely-there functional sweep (rgba(255,255,255,0.015)) over 1.6s ease-out, plus a screen-reader status; the scene chunk is lazy behind it and the gate waits on `data-scene-ready`.
- **Fetch failure:** a centred panel card (the only centred element) that names the problem and offers a working retry — "โหลดข้อมูลระดับน้ำไม่ได้ ตรวจการเชื่อมต่อแล้วลองใหม่ได้เลย".
- **No WebGL:** the shell, sentence and data survive; a quiet status line says the device shows text only.
- **Reduced motion:** the skeleton sweep stops; the camera snaps instead of flying.

### Camera Rig
Oblique default; `CameraControls` clamped (polar 0.12–1.42 rad, distance 70–2800 km), target bounded to Thailand; preset moves and the one ใกล้ฉัน flight run 1600ms exponential ease-out (`easeOutExpo`, mirrored by the CSS `cubic-bezier(0.16, 1, 0.3, 1)`) and are interrupted by any user input.

## Do's and Don'ts

### Do:
- **Do** take every colour from `tokens.ts` (or its generated CSS variables) — including shaders, which import the module.
- **Do** include `tonemapping_fragment` / `colorspace_fragment` in any new custom fragment, and keep the Canvas `flat`.
- **Do** put measured time and source on every data surface, in caption `text-faint`, with tabular lining figures.
- **Do** keep the answer sentence in Noto Serif Thai with `text-wrap: balance` and 1.65 leading.
- **Do** keep gauge encoding geometric first (cap at 0.78, fill inside, overflow above the cap, hollow stale) so colour is redundant, never load-bearing.
- **Do** use 160–260ms exponential ease-out (`cubic-bezier(0.16, 1, 0.3, 1)`) for UI state, 1.2–2.0s for camera flights, and cut instead of flying under `prefers-reduced-motion`.
- **Do** theme browser surfaces (selection, focus ring, caret, scrollbar, underline offset, `color-scheme: dark`) — they are part of the design.
- **Do** keep every surface that floats over the relief on the veil + 1px border recipe, and verify its text contrast over the brightest terrain.

### Don't:
- **Don't** add a hero section, card grids, section numbers, nested cards, a navbar of links or a marketing footer — the relief is the page.
- **Don't** use purple, indigo or violet gradients, neon cyan on dark, gradient text, glowing borders, radial glow halos or spotlights, or cream/beige "un-slop" defaults.
- **Don't** use glassmorphism beyond the two functional blurs, `box-shadow` in shipped chrome (the one reserved shadow is Phase 4's), or extreme border radii.
- **Don't** use Inter, Geist, Roboto, Arial or the system UI font as the voice, or italic in any face, or crushed tracking (zero `letter-spacing` declarations exist).
- **Don't** auto-rotate, wobble, starfield, particle, arc or bloom the 3D scene; the camera moves only on input, presets, or the one flight.
- **Don't** put Thai text into WebGL, or let an unbounded orbit lose the horizon.
- **Don't** give stale data a fill, or levels 1–3 strong colour, or the scale a sixth stop without re-running the CVD search.
- **Don't** say ปลอดภัย, จะท่วม, ไม่ท่วม or anything that predicts or reassures; never claim safety. (Test-enforced.)
