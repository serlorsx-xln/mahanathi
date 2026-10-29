# ADR 0006: The relief look — vertical exaggeration 12×, tokens at face value

- **Status:** accepted (Phase 3, look-dev)
- **Date:** 2026-10-30
- **Deciders:** implementing agent, from the measured screenshots in
  `docs/design/screens/phase-3/`; the lead confirms at the Phase 3 review
  (the visual pass this session cannot make — DEVLOG 2026-10-29 23:25).

## Context

The relief is a 1,815 × 2,048 heightmap over 92–111° E, 3–24° N laid on a
plane of ~1,880 × 2,320 km. At true scale the vertical dimension is a
rounding error: Doi Inthanon (2,565 m) would be 0.11% of the country's
width — a billiard table. Every hydrographic relief exaggerates; the
question is how much before the land stops reading as land.

Three forces pull against each other:

1. **The central plain must stay a plain.** The product's answer is about
   rivers on flat land; if exaggeration tilts the paddy plain, the
   geometry lies about where floods live.
2. **The ranges must read as structure**, not texture, at the national
   zoom where fog and distance already compress contrast.
3. **The gauges are absolute-size objects** (poles 12–44 km at this
   scale). The relief around a station must not dwarf or vanish under
   them.

## Evidence

At the ลุ่มเจ้าพระยา preset (the plain ringed by ranges), measured from
the committed screenshots (luminance gradients as a relief-legibility
proxy):

| Exaggeration | gradient mean | L p10 | Notes from the screenshots |
|---|---|---|---|
| 6× | 0.0149 | 0.151 | modal terrain colour covers the most area; ranges read as gentle mottling |
| 12× | 0.0168 | 0.151 | ridge structure carries; plain still samples as flat modal colour |
| 30× | 0.0193 | 0.147 | self-shadowing steepens (darkest decile drops), silhouette spikes occlude neighbours |

The exaggeration control stays in the page (slider, 4–40×) so the lead's
visual pass can retune the default without a rebuild.

## Decision

**Default vertical exaggeration 12×** (metres → kilometres: Inthanon
becomes 30.8 km on a 1,880 km-wide country, roughly the ratio of a good
physical globe's relief model). The plain's modal terrain colour and its
flatness are preserved at 12× (the paddy toning keys off measured
flatness, which exaggeration does not change — the shader scales the
normal's vertical term by the same factor, so `flatness` is invariant).

**The colour pipeline is flat by design**: the Canvas runs
`NoToneMapping` with an opaque framebuffer, and both custom fragments
close with three's `tonemapping_fragment`/`colorspace_fragment` chunks,
so a token colour that survives lighting unchanged reaches the screen
exactly (the national screenshot samples back sea `#0d1a23` at 34% and
ground `#0e161d` at 21%). Tone mapping would have rewritten every token
colour opaquely; the brief's "no colour exists outside the tokens" is
enforceable only without it. The night mood comes from an explicit
moonlit Lambert scale in the terrain shader (`0.22 + 0.55·key +
0.16·fill`), not from a curve.

**Aerial fog** `smoothstep(900, 2600)` km toward the sea's colour, 85%
at most: at the national zoom this is what keeps the far north and the
neighbours receding while the centre stays legible (terrain ≈ 3:1 over
the sea at the country's centre); at the plain and station zooms the
camera sits inside the fog-free radius and only the horizon fades.

**Answer-panel legibility over the relief is measured, not assumed:**
the veil (88% panel over the brightest sampled terrain) blends to
`#212a32`, on which textPrimary is 12.3:1, textSecondary 6.8:1 and even
textFaint 4.7:1 — the caption colours stay AA over the brightest thing
the relief can put under the panel.

## Consequences

- The slider stays a dev control on `/lookdev`; if the lead's pass
  moves the default, it is a one-number change in `Lookdev.tsx` and this
  ADR gains the new evidence row.
- Removing tone mapping means the shader's light scale is the only
  brightness control; any new material must include the output chunks or
  it will render dark (the mistake is recorded in DEVLOG 2026-10-30
  ~00:20 and the screenshot-sampling check now guards it).
- 12× makes gauge poles 12–44 km tall against terrain relief of up to
  ~31 km at the ranges: prominent at the national zoom, proportionate at
  the station zoom. If Phase 4's station close-ups show the relief
  swamping the gauge, the pole heights — not the exaggeration — are the
  knob to turn.
