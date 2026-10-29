# ADR 0005: Typography — IBM Plex Sans Thai + Noto Serif Thai

- **Status:** accepted (Phase 3, look-dev); evidence re-verified 2026-10-30 after the corruption incident below
- **Date:** 2026-10-29 (decision), 2026-10-30 (evidence correction, Phase 3b Step 0)
- **Deciders:** implementing agent, from the specimen (`/specimen`) and the
  measurements in `docs/design/screens/phase-3/specimen-measurements.json`;
  the lead confirms at the review.

## Evidence integrity note (2026-10-30)

The version of `specimen-measurements.json` committed with the original
decision had been silently corrupted: the e2e re-run that produced it
happened **after** the rejected candidates had been trimmed from the
build, so its Anuphan and Noto Sans Thai rows measured the system
fallback font (the tell: two different families with bit-identical
floats, `digit: 55.615234375` for both). `document.fonts.check`
reported them "loaded" anyway, because an unknown family resolves to
the already-loaded default. The ADR's own numbers had been measured
before the trim and were correct, so the ADR and its committed
evidence contradicted each other — the lead's review rightly flagged
this. The re-verification (fonts restored with
`scripts/research/promote_fonts.py`, all families proven loaded by
`document.fonts.load`, then trimmed back) confirms every original
value; the committed JSON is now the verified one, and
`e2e/tests/measure.spec.ts` refuses to write a partial or
fallback-collapsed measurement ever again.

## Context

The brief demands a loopless Thai sans with real character and a
harmonised Latin for the UI, an upright Thai serif for the human
sentence only if the specimen earns it, self-hosted OFL WOFF2 subsets,
tabular figures for measurements, and Thai leading that survives
stacked vowels and tone marks. `docs/research/typography.md` shortlisted
three pairings (A: IBM Plex Sans Thai, B: Anuphan, C: Noto Sans Thai as
the control) and four sentence options (sans-only, Noto Serif Thai,
Pridi, Trirong), with a hypothesis ranking but no winner.

The specimen page (`web/src/routes/Specimen.tsx`) sets every candidate
with real Phase 2 output: the national sentence, three station
sentences, the stacked-mark check line and a tabular readout with
signed values. Screenshots at both gate viewports are committed in
`docs/design/screens/phase-3/`.

A constraint on this decision: the implementing session cannot view
images (DEVLOG 23:25), so the choice rests on the research's measured
facts, the canvas/DOM measurements below, and the research's HarfBuzz
renders — with the lead's visual pass as the confirming step the
workflow already requires.

## Measured evidence

From the re-verified `specimen-measurements.json` (canvas TextMetrics
at 100 px, Chrome, via `e2e/tests/measure.spec.ts`):

| Question | IBM Plex Sans Thai | Anuphan | Noto Sans Thai |
|---|---|---|---|
| Digit width (em) | 0.600 | 0.600 | 0.572 |
| `14:00` (em) | 2.692 | 2.692 | 2.556 |
| `−0.30` vs `+0.39` (em) | 2.672 / 2.672 | 2.672 / 2.672 | 2.306 / 2.556 |
| Tallest ink ascent (em, of น้ำล้นตลิ่ง) | 1.083 | 1.066 | 1.009 |
| Ink above own hhea ascender | no | **yes** (ascender 1.025) | no |
| Ink use of a 1.6 line box (worst string) | 0.844 | 0.790 | 0.794 |

- Signed columns align in Plex and Anuphan (minus equals the digit and
  the plus); **Noto Sans Thai's minus is 0.25 em narrower than its
  plus**, so signed measurement columns would misalign — a data-adjacent
  flaw. Anuphan matches Plex here because its Latin *is* Plex's Latin
  (`docs/research/typography.md`), which is also why A vs B is almost
  purely a Thai-drawing comparison.
- Anuphan's ink rises above its own ascender (research: 41 units above
  hhea 1025), so in any container that clips to the content box (chips,
  ellipsis, `line-clamp`) its tall stacks can be cut. No DOM overflow
  was observed at line-height 1.6 in any face.
- The national sentence sets 31.1 em in Plex 600, 30.9 in Anuphan,
  30.5 in Noto Sans Thai, 31.0 in Noto Serif Thai, 29.2 in Trirong,
  29.0 in Pridi.

From the research: Plex has the most complete coverage (Latin-1 96/96,
dotted circle U+25CC present), one team drew its Thai and Latin
together, and it is the only face needing the RFN rename. Anuphan is
unmaintained since 2024. Noto Sans Thai's Latin is the generic one the
brief's control was there to expose.

## Decision

**UI and numbers: IBM Plex Sans Thai, weights 400 and 600**, self-hosted
as subsets renamed **"Mahanathi Sans"** (OFL condition 3: no Reserved
Font Name on a modified version; verified in the shipped files' name
IDs 1/3/4/6). Hinting kept — small Thai on Windows.

**Sentence: Noto Serif Thai 400**, unhinted (display sizes do not need
it). It gives the human sentence a second, quieter voice against the
engineered UI face, has the lowest collide measurement of the serifs
(1.349), tabular figures, and the smallest 400 subset (16.3 KB). Pridi
sets narrowest and its loopless-terminal Thai is the closest to the UI
face's genre, but it renders dark at Regular, is classified display,
and has proportional figures inside a sentence full of numbers; Trirong
reads formal/official, which the product brief forbids imitating.

**Mono: none.** All chosen figures are already tabular; the research
found Plex Mono adds only equal-width punctuation at 7 KB.

**Leading:** body 1.6 (Plex's worst ink box uses 84 % of it), the
sentence 1.65, display 1.3 — all in the tokens and tested against the
measured ink extents.

## Consequences

- Payload: 23.5 + 24.5 KB (UI, hinted) + 16.3 KB (serif, unhinted) ≈
  64 KB of fonts, self-hosted, `font-display: swap`; preloads for the
  critical UI faces land with the Phase 4 product shell.
- The rename means the CSS family is "Mahanathi Sans"; the specimen
  labels keep the source name for honesty.
- Rejected candidates are deletable from the build (done) but
  restorable with `scripts/research/promote_fonts.py` (no arguments)
  for a live side-by-side; the committed screenshots and measurements
  remain the record of the comparison.
- Every string the product sets is covered by the subset character set
  (`scripts/research/subset_sizes.sh`'s UNICODEES); anything new needs
  a subset rebuild, which the promote script does.
