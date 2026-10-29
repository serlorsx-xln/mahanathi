# Typography shortlist: Thai + Latin OFL faces

Researched 2026-10-29 by the implementing agent (an AI model), as input to the Phase 3 type specimen (`/specimen`, item 3 of `process/prompts/phase-3.md`). This is a shortlist with a hypothesis ranking. **It does not pick a winner.** The specimen picks, and ADR `0005-typography.md` records the choice.

Every number below comes from running the scripts in `scripts/research/` on the font files pinned in section 8. The inspection script is `scripts/research/inspect_fonts.py`. Renders are HarfBuzz output, not browser output. Raw reports (`.cache/fonts/_report/`), renders (`.cache/fonts/_render/`) and subsets (`.cache/fonts/_subset/`) are gitignored research artefacts.

## Summary

- **UI face shortlist:** the three faces the brief named.
  - Pairing A: IBM Plex Sans Thai
  - Pairing B: Anuphan
  - Pairing C, the control: Noto Sans Thai
- **No fourth face qualified.** I tested every Google Fonts family with a Thai subset (34 families, 21 of them sans) plus LINE Seed Sans TH.
  - Eight of the Google Fonts sans are loopless, and one of those (Pattaya) is a display script.
  - The five loopless faces beyond the named three, plus LINE Seed Sans TH, fail on figures, genre or ownership (section 6).
- **Sentence serif shortlist:** Noto Serif Thai, Pridi, Trirong, with Taviraj ranked last. All four have looped Thai. There is no OFL upright loopless Thai serif on Google Fonts.
- **Figures:** no UI candidate and no serif has a `tnum`, `lnum`, `pnum` or `onum` feature.
  - Default figures are already tabular and lining in all three UI faces, so the brief's `font-variant-numeric: tabular-nums lining-nums` changes nothing in them.
  - Proportional figures for prose are not available in any of them.
- **Plex licence:** only the Plex faces declare a Reserved Font Name ("Plex"). Self-hosted subsets of them must be renamed.
- **Leading:** the faces differ most in vertical extent. Plex Sans Thai needs the most leading and Noto Sans Thai the least. Anuphan's real ink rises above its own hhea ascender, which is a clipping risk.
- **Hypothesis ranking:**
  - UI: Plex Sans Thai > Anuphan > Noto Sans Thai.
  - Sentence: Noto Serif Thai ≈ Pridi > Trirong > Taviraj.
  - A sans-only sentence is also a valid outcome.
  - Mono: probably redundant.

## 1. Constraints recap

From `docs/brief/01-DESIGN-DIRECTION.md`, `docs/brief/00-PRODUCT.md` and `docs/brief/02-ENGINEERING.md`:

- **World.** "A night-time hydrographic instrument". The theme is dark and must survive a bright interview room on a laptop.
- **UI face.** A loopless Thai sans "with real character" and a harmonised Latin, all under OFL. The brief names IBM Plex Sans Thai, Anuphan and Noto Sans Thai.
- **Sentence face (optional).** An upright Thai serif (Noto Serif Thai, Taviraj, Trirong are named) "if the specimen proves it". Never italic.
- **Banned as the voice:** Inter, Geist, Roboto, Arial, the system UI font.
- **Banned type tells:**
  - italic serif display
  - crushed tracking
  - flat hierarchy
  - one family doing everything with no size or weight steps
- **Numbers.**
  - `font-variant-numeric: tabular-nums lining-nums` for measurements.
  - Monospace only for raw readouts, and only if it earns its place.
- **Line height.** Body starts at about 1.6. Verify it with "น้ำล้นตลิ่ง", "ผู้ใหญ่บ้าน", "ปิ่นเกล้า" and "จังหวัดพระนครศรีอยุธยา".
- **Display size.** 6rem at most. There must be a clear scale with obvious steps.
- **Font files.**
  - Self-hosted WOFF2, subset to Thai and Latin.
  - Preload only critical faces.
  - `font-display: swap`.
  - The CSP allows self, plus font origins only if required ("preferably none").
- **No official look.** "never imitate a government site" (product brief). This rules Sarabun out (section 6).
- **Wordmark.** "มหานที" set in type with care.
- **Budget.** LCP under 2.5 s on throttled Fast 4G, mid-tier mobile. The national sentence is the likely LCP element.

## 2. Method and scripts

### Tools

- fontTools 4.60.2 and uharfbuzz 0.51.7 (HarfBuzz 12.1.0), on Python 3.9.6.
- They run in `.cache/venv`, always as `.cache/venv/bin/python -I`.
- The versions are pinned in `scripts/research/requirements-fonts.txt` (brotli 1.2.1 provides WOFF2).
- Font files are passed as paths, and the scripts write only where told.

| Script | What it does | Re-run (from the project root) |
|---|---|---|
| `scripts/research/inspect_fonts.py` | Reports, per file: name table, licence, axes, coverage, GSUB/GPOS features per script, whether numeric features actually change digits, vertical metrics, shaped ink extents and line-height limits, loop test, proportions | `.cache/venv/bin/python -I scripts/research/inspect_fonts.py [--var wght=600] [--json OUT.json] FONT...` |
| `scripts/research/summarize_reports.py` | Compact table from the JSON reports | `.cache/venv/bin/python -I scripts/research/summarize_reports.py .cache/fonts/_report/regular.json` |
| `scripts/research/glyph_advances.py` | Advances of ASCII and Thai digits, the minus sign, punctuation and spaces next to figures | `.cache/venv/bin/python -I scripts/research/glyph_advances.py FONT[@wght=400]...` |
| `scripts/research/compare_outlines.py` | Glyph-by-glyph comparison of two fonts' Latin (advance, outline structure, point deviation, bbox), to detect derived Latin | `.cache/venv/bin/python -I scripts/research/compare_outlines.py A[@wght=400] B --quiet` |
| `scripts/research/render_specimen.py` | HarfBuzz-shaped SVG of test lines with hhea lines and the CSS line box drawn in | `.cache/venv/bin/python -I scripts/research/render_specimen.py --out .cache/fonts/_render/x.svg --lh 1.6 FONT[@wght=400]...` |
| `scripts/research/subset_sizes.sh` | WOFF2 subset size, hinted and unhinted, with optional instancing of variable fonts | `scripts/research/subset_sizes.sh LABEL FONT ["wght=400"]` |

### Definitions used in the tables

**Ink extents.**
- Measured in em (font units ÷ unitsPerEm), after shaping with script Thai and language th, with default features. That means GSUB `ccmp` plus GPOS `mark` and `mkmk`.
- Digits are script Common and take the script of the surrounding run, so numeric features were tested in both a `latn` run and a `thai` run.

**Corpora.**
- *Real words:* the four brief strings, the national sentence "ตอนนี้ 22 สถานีน้ำล้นตลิ่ง ส่วนใหญ่อยู่ลุ่มเจ้าพระยาตอนล่างและภาคใต้ตอนบน", and 26 words with tall or deep stacks. Examples: ปั๊มน้ำ, ฟื้นฟู, ญี่ปุ่น, ฤๅษี, กฎหมาย, ปฏิบัติ, อ.ศรีสัชนาลัย, เขื่อนภูมิพล. The full list is in the script.
- *Stress:* every tall consonant (ป ฝ ฟ ฬ, plus ก น ส) with every upper vowel and tone, and ฎ ฏ ฐ ญ ฤ ฦ ป ก ร ฬ with every below vowel and tone.
- Stress includes non-words such as ป็๋. Plan line-height on the real-word numbers. Stress is the worst case for user-typed search input.

**Line-height limits.** These follow the CSS half-leading model. A = hhea ascender, D = |hhea descender|, U = unitsPerEm. All values are unitless line-heights.
- **ink box** = max((2·top − A + D)/U, (2·|bottom| − D + A)/U). Below this line-height, shaped ink leaves the line box. That matters for `overflow: hidden`, `line-clamp`, scroll containers and anything stacked tightly under the text.
- **collide** = (top + |bottom|)/U. Below this line-height, the deepest below-mark of one line can touch the tallest stack of the next line. It only happens if the two land at the same x position, so the specimen must check real wraps. This number does not depend on the font's metrics.

**Loop test.**
- For each of 17 letters whose looped forms have a closed head loop (บปนพฟมวหยลสรคถภดต), count contours wound opposite to the outer contour (counters).
- Loopless faces score 0/17.
- Every looped verdict was confirmed in a render. Open loops are missed: Krub scores 3/17 but is visibly looped.
- An earlier assumption in this research, that Bai Jamjuree is loopless, was wrong. This test caught it.

**WOFF2 subset.**
- Character set: Thai block, Basic Latin, U+00A0, ° ± ³ · ×, U+2009, U+200B, U+202F, en dash, curly quotes, ellipsis, U+2212 and U+25CC. These are the `UNICODES` in `subset_sizes.sh`.
- Options: `--layout-features='*'`.
- "Hinted" keeps the TrueType hinting. "Unhinted" adds `--no-hinting --desubroutinize`.
- Variable fonts on Google Fonts ship unhinted, so their two numbers are equal.

**Popularity.**
- This is the `popularity` field of `https://fonts.google.com/metadata/fonts`. Lower means more used.
- Several families share a value (Roboto, Open Sans and Inter are all 2), so it is a coarse rank.
- Its exact definition is unverified.

## 3. Facts that hold across the candidates

1. **No numeric OpenType features.**
   - None of Plex Sans Thai, Anuphan, Noto Sans Thai, Noto Serif Thai, Taviraj, Trirong or Pridi has `tnum`, `lnum`, `pnum` or `onum`. I checked by shaping digits with each feature on, in both a latn run and a thai run.
   - ASCII figures are lining in all of them by default, and tabular in all but Pridi:
     - Plex Sans Thai 600
     - Anuphan 600
     - Noto Sans Thai 572
     - Noto Serif Thai 559
     - Taviraj 581
     - Trirong 587
   - So the brief's `font-variant-numeric` declaration is a no-op on these files. Keep it, because it still applies to any fallback font.
   - **No faces offer proportional figures for prose.** The "22" in the national sentence will be tabular.
2. **Figure width across weights.**
   - Plex Sans Thai: 600 at 400, 500, 600 and 700.
   - Anuphan: 600 from 300 to 700.
   - Noto Sans Thai: 571-572.
   - So a bold and a regular number occupy the same width in these three.
   - Taviraj and Trirong do not: Regular is 581 and 587, Medium is 597 and 602.
3. **Thai digits ๐-๙ are proportional in every face:**

   | Face | Thai digit advances |
   |---|---|
   | Plex Sans Thai | 556-841 |
   | Anuphan | 536-830 |
   | Noto Sans Thai | 504-720 |
   | Noto Serif Thai | 570-762 |
   | Taviraj | 636-811 |
   | Trirong | 644-823 |
   | Pridi | 576-783 |

   Columns of Thai digits will not align.
4. **Spaces next to figures.**
   - No face has U+2009 (thin space) or U+202F (narrow no-break space).
   - U+2007 (figure space) is missing from the sans faces. Taviraj, Trirong and Pridi map it to a zero-advance glyph.
   - U+00A0 is present everywhere, at the width of the normal space.
   - Use U+00A0 between number and unit ("1.40 เมตร", "14:00 น."). Any of the missing spaces would be drawn from a fallback font at an uncontrolled width.
5. **Minus sign U+2212.**
   - Plex Sans Thai and Anuphan: 600, which equals the digit width. The plus sign is also 600, so "−0.35" and "+0.35" align in a column.
   - Noto Sans Thai: 322, the same as its hyphen, against a 572 digit.
   - Noto Serif Thai: 310.
   - Taviraj: 582. Trirong: 588. Pridi: 412.
6. **Licences.** Every shortlisted face is OFL 1.1, according to its `OFL.txt` and name ID 13.
   - Only IBM Plex Sans Thai and IBM Plex Mono declare a Reserved Font Name: `Copyright © 2017 IBM Corp. with Reserved Font Name "Plex"`.
   - OFL condition 3: "No Modified Version of the Font Software may use the Reserved Font Name(s) … This restriction only applies to the primary font name as presented to the users."
   - OFL FAQ 2.6: "Removing any parts of the font when delivering a webfont … is considered modification. This is permitted by the OFL but would not normally allow the use of RFNs."
   - So a self-hosted Plex subset must not carry "Plex" in name IDs 1, 3, 4, 6, 16 or 17. The CSS family name should avoid it too.
7. **Licence text in subsets.**
   - OFL condition 2 requires the copyright notice and licence "in each copy", either as files or in metadata fields.
   - `pyftsubset` keeps only name IDs 0-6 by default, which drops IDs 13 (licence) and 14 (licence URL).
   - Adding `--name-IDs=0,1,2,3,4,5,6,13,14` cost 84 bytes on the Plex Sans Thai Regular unhinted subset (16,892 to 16,976 B) and 64 bytes on the Anuphan VF.
   - Alternatively, ship `OFL.txt` next to the fonts.
8. **Thai shaping support.** Every candidate has GSUB `ccmp` and GPOS `mark` and `mkmk` under the `thai` script. Anuphan has no GPOS `kern` under `thai`; its Latin is kerned.
9. **Vertical metric sets agree within each UI file.**
   - Plex Sans Thai and Noto Sans Thai: hhea = typo = win.
   - Anuphan: win (1119/324) differs, but USE_TYPO_METRICS is set and typo = hhea.
   - How each browser and platform picks a set was not tested here (unverified).
10. **Dotted circle U+25CC.**
    - Missing from Anuphan, Taviraj, Trirong and Pridi.
    - Present in Plex Sans Thai, Noto Sans Thai and Noto Serif Thai.
    - A shaper only shows it for a mark with no base, for example mistyped search input.

## 4. Pairings

All three pairings should be built in the specimen with the sentence face switchable. The options are none, Noto Serif Thai, Pridi and Trirong. That way the UI decision and the serif decision stay independent.

### Pairing A: IBM Plex Sans Thai + Noto Serif Thai (+ IBM Plex Mono, optional)

**Roles.**
- **UI and numbers:** IBM Plex Sans Thai, static.
  - Likely weights: 400 and 600, with 500 as an option.
  - Upstream v1.2 also has a "Text" weight (usWeightClass 450) that Google Fonts does not ship.
- **Sentence:** Noto Serif Thai, a variable instance near wght 400. Pridi Light is the alternate.
- **Raw readouts (optional):** IBM Plex Mono Regular.

**Why it fits.**
- **Harmonised by construction.** One team drew the Thai and Latin as one family (name table: Mike Abbink, Paul van der Laan, Pieter van Rosmalen, Ben Mitchell, Mark Frömberg; manufacturer Bold Monday).
  - Thai body height น is 0.546 em against a Latin x-height of 0.516 (ratio 1.058).
- **Character.** In the renders (`.cache/fonts/_render/sans1.svg.png`), the Thai has squared shoulders and cut terminals. It reads as engineered, which suits an instrument. This is a visual judgement from HarfBuzz renders on white, not the product's dark theme.
- **Coverage.** The most complete of the three: Latin-1 96/96, and every symbol in the test set except U+2009 and U+202F.
- **Figures behave.**
  - Digits are 600 at every measured weight.
  - Minus and plus equal the digit width.
  - `zero` works, and ss03/ss04 give alternate zeros.
  - ss01 and ss02 give alternate a and g, a lever on character.
- **Mono match.** Plex Mono's digits are 600 wide and 8 of 10 are within 15 units of Plex Sans Thai's, so an inline readout would not change the figure shapes.

**Risks.**
- **Tallest marks of the three.**
  - Real-word top is 1.083 em at 400, rising to 1.148 at 700.
  - Body at 1.6 keeps ink inside the line box at 400 (ink box 1.584), but not at the upstream Text weight 450 (1.608) or at 500 and above (1.640 to 1.714).
  - For the display sentence at 600, no-touch needs 1.485 or more (collide).
  - The sentence will be set looser than in B or C.
- **Static files.** One file per weight.
- **RFN.** "Plex" must be renamed out of the subsets (section 3, item 6).
- **Corporate identity.**
  - The Google Fonts description opens "IBM Plex® is the corporate typeface for IBM worldwide".
  - The Latin, IBM Plex Sans, is widely used (popularity 8).
  - That may sit against "use our own brand", but the brief names the face as a candidate.

**Payload estimate (Thai + Basic Latin subsets).**
- Plex Sans Thai 400 + 600: 22.9 + 23.9 = 46.8 KB hinted, or 16.5 + 17.4 = 33.9 KB unhinted.
- Plus Noto Serif Thai 400 instance: 16.0 KB.
- Plus Plex Mono 400, if used: 10.3 KB hinted, 6.7 KB unhinted.

**Specimen must test.**
- The sentence at the leading it needs (1.45-1.5): does it still look set, or loose?
- Plex at 400 on the dark background: engineered, or cold?
- The upstream Text (450) weight against 400 for light-on-dark body text.
- Whether Plex Mono earns a place at all next to digits that are already 600 wide.

### Pairing B: Anuphan + Pridi (alternate: Trirong) (+ IBM Plex Mono, optional)

**Roles.**
- **UI and numbers:** Anuphan variable, wght 100-700.
- **Sentence:** Pridi Light or Regular. Trirong Light or Regular is the alternate.
- **Raw readouts (optional):** IBM Plex Mono. Anuphan's Latin is Plex Latin (section 5.3).

**Why it fits.**
- **Isolates the Thai.** The Latin matches the Latin in Plex Sans Thai: same advance on 61 of 62 alphanumerics at wght 400 (u differs by 2 units) and on 62 of 62 at 700, with 15 outlines point-for-point identical. So A against B compares only the Thai drawing.
- **The Thai drawing.**
  - Anuphan's Thai is rounder in the renders, and slightly larger: body 0.570 em against 0.546 for Plex, Thai/x 1.105.
  - It sets the national sentence 2% narrower: 30.864 em against 31.556 for Plex.
  - It was drawn from the Plex Sans Latin, not modified from Plex Thai. In the Google Fonts description: "All drawings and outlines of Thai characters in this project are based solely on the Latin version of IBM Plex Sans".
  - None of 53 Thai letters share an outline or bbox with Plex Sans Thai.
- **One variable file** covers wght 100-700 (46.0 KB), so in-between weights (350, 450) can tune light-on-dark text.
- **Compact marks.** Real-word collide is 1.333 at 400 and 1.387 at 600, so the sentence can be set about 0.1 tighter than in A.
- **Pridi.**
  - Its Thai is looped, but Google Fonts describes it as "designed specifically within loopless terminal concepts", which makes it the closest serif to a loopless UI face.
  - It ships no italics at all.
  - It has the smallest real ink box of the serifs (1.458; collide 1.38).
  - Its figures are proportional, which suits a number inside a sentence.

**Risks.**
- **Anuphan's hhea ascender (1025) is below its own ink.** "น้ำล้นตลิ่ง" reaches 1.066 em, 41 units above. Tall stacks draw outside the content area. At `line-height: normal` (1.30), and in any clipped container, they can be cut.
- **Provenance.**
  - Google Fonts and the Cadson Demak article call it a "self-improvement/educational" project.
  - The upstream's last commit is a8ee844 on 2024-07-12 ("fix v-metrics (Winascents)"), and there are no tagged releases.
  - The OFL permits use regardless, but nobody is maintaining it.
- **Gaps.** No GPOS `kern` under `thai`. No U+25CC, U+200B, U+2009 or U+202F.
- **Pridi.**
  - Google Fonts classifies it DISPLAY as well as SERIF.
  - Its Regular renders dark (`.cache/fonts/_render/serif.svg.png`), so Light is the probable sentence weight. Light's subset size was not measured.
  - Its proportional figures will not match the panel's tabular numbers.

**Payload estimate.**
- Anuphan full VF: 46.0 KB. Two instances instead (400 + 600): 17.2 + 18.7 = 35.9 KB.
- Plus Pridi Regular: 25.6 KB hinted, 14.8 KB unhinted. Or Trirong Regular: 26.5 KB hinted, 16.1 KB unhinted.

**Specimen must test.**
- Clipping of น้ำ, ปั๊ม and ปิ่น under `overflow: hidden`, `line-clamp`, `text-overflow: ellipsis`, `background-clip: text`, and at `line-height: normal`.
- Anuphan at 380-450 on the dark background.
- Whether Pridi's slab Latin and its "loopless terminal" Thai sit well beside Anuphan.
- Trirong as the alternate. It is narrow and tall, "formal" per Google Fonts, and needs about 1.46 for no-touch.

### Pairing C (control): Noto Sans Thai + Noto Serif Thai

**Roles.**
- **UI and numbers:** Noto Sans Thai variable, wght 100-900, wdth 62.5-100.
- **Sentence:** Noto Serif Thai variable.
- No mono.

**Why it is here.**
- **Cheapest and most complete superfamily.**
  - The 400 instance is 13.8 KB, the smallest of all.
  - A weight-only build (wdth pinned at 100) is 44.1 KB.
  - Upstream also ships a weight-only "slim" VF.
- **Condensed widths for dense labels.** At wdth 75 the national sentence is 25.55 em against 30.422 at 100, about 16% narrower.
- **Tightest marks.**
  - Real-word collide is 1.274 at 400, the lowest of all faces.
  - The serif's vertical metrics are set up the same way: hhea = typo = win, with USE_TYPO_METRICS set in both.
- **Control role.** It answers whether A or B is actually better than the default-looking choice.

**Risks.**
- **Latin identical to Noto Sans** (62 of 62 identical).
  - Noto Sans in turn is a close relative of Open Sans: 40 of 62 near-identical within 15/1000 em, 6 the same structure, 16 different. Open Sans has popularity 2.
  - This Latin has the least "real character" of the three.
- **Its own description positions it for display.** Google Fonts says Noto Sans Thai is "mainly suitable for headlines, packaging and advertising".
- **Coverage gaps.** No ±, ³ (as in m³), ↑ or ↓.
- **Signed values do not align:** minus is 322 against a 572 digit.
- **Unverified:** it may also be the Thai fallback on some platforms. If so, it would read as the system UI font, which the brief bans as the voice.

**Payload estimate.**
- 400 + 600 instances (13.8 + 14.0 KB) plus Noto Serif Thai 400 (16.0 KB): 43.8 KB.
- Or the weight-only VF (44.1 KB) plus the serif (16.0 KB): 60.1 KB.

**Specimen must test.**
- Whether the Latin reads as generic next to A and B, in "Gt.11", "MSL", "ThaiWater" and "Mahanathi".
- wdth 75-87.5 for legend and timeline labels.
- The minus sign in signed columns.

### Hypothesis ranking (to be confirmed or overturned by the specimen)

**UI face.**

1. **IBM Plex Sans Thai.**
   - For: Thai and Latin designed together; the most complete coverage; the best figure behaviour (digits at 600 at all weights, minus equal to digit width); the most instrument-like drawing.
   - Against: needs the most leading, and the build needs an RFN rename.
2. **Anuphan.**
   - For: the same Latin as Plex, a softer and larger Thai, tighter leading, and one variable file.
   - Against: metrics that let ink escape the ascender, and an unmaintained educational project.
3. **Noto Sans Thai.**
   - For: cheapest and the most typographically flexible (wdth axis).
   - Against: the most generic Latin and coverage gaps that touch our data (³, ±, minus width).

The tests that decide it:
- the sentence's look at the leading each face needs
- clipping with Anuphan
- each face at 400 on the dark background, in a bright room

**Sentence face.**

- **Noto Serif Thai ≈ Pridi.**
  - Noto Serif Thai has tabular figures, a variable weight, the lowest serif collide (1.349) and a small 400 instance (16.0 KB). Its Latin provenance is unverified.
  - Pridi has the looped Thai closest to a loopless UI face, no italics to misuse, and compact ink. Its figures are proportional and it is classified display.
- **Trirong.** Narrow and tall, but "formal" with transitional Latin. It needs about 1.46 leading. It risks the official look.
- **Taviraj.** "Well-suited for formal usage", wide, and in the "farangses" genre. It is the most ceremonial and has the largest real ink box of the serifs (1.614).
- **None.** A sans-only sentence with a real size and weight step is a legitimate result. Use it if every serif makes the sentence feel like an announcement.

**Mono.**

Hypothesis: drop it.
- All three UI faces already have tabular figures.
- Plex Mono adds only equal-width punctuation (period and colon 600, against 272 and 292 in Plex Sans Thai). That spreads "14:00" and "3.42" apart.
- Plex Mono covers no Thai except ฿ (U+0E3F).

## 5. Per-face facts

### 5.1 UI candidates

| | IBM Plex Sans Thai | Anuphan | Noto Sans Thai |
|---|---|---|---|
| Designer, foundry (name table) | Mike Abbink, Paul van der Laan, Pieter van Rosmalen, Ben Mitchell, Mark Frömberg; manufacturer Bold Monday; vendor IBM. GF METADATA: "Mike Abbink, Bold Monday" | Cadson Demak; vendor NONE. GF description credits Mint Tantisuwanna; the Cadson Demak article on it is by Kornkanok Tantisuwanna | Monotype Design Team; manufacturer Monotype Imaging; vendor GOOG |
| Licence (from `OFL.txt`) | OFL 1.1, **RFN "Plex"** | OFL 1.1, no RFN | OFL 1.1, no RFN |
| Google Fonts file version | Version 1.1 | Version 3.002 | Version 2.002 |
| Upstream | IBM/plex `@ibm/plex-sans-thai@1.1.0` (2024-11-13): files say Version 1.2, same outlines on the 120 glyphs compared, typo metrics changed to 780/-220/650 with USE_TYPO off, adds Text weight (450) | cadsondemak/Anuphan, last commit a8ee844 (2024-07-12), no releases; repo has TTF, OTF, variable and webfonts | notofonts/thai `NotoSansThai-v2.002` (2023-09-26): full VF (467 glyphs) and a wght-only slim VF (459 glyphs) |
| Format | Static TTF: 100, 200, 300, 400, 500, 600, 700 | Variable: wght 100-700 (default 400) | Variable: wght 100-900, wdth 62.5-100 |
| Glyphs | 520 | 790 | 467 |
| Thai / Basic Latin / Latin-1 | 87/87, 95/95, 96/96 | 87/87, 95/95, 96/96 | 87/87, 95/95, 84/96 (no ¤ ¦ ¬ soft hyphen ± ² ³ µ ¹ ¼ ½ ¾) |
| Missing from symbol test set | U+2009, U+202F | U+2009, U+200B, U+202F, U+25CC | U+2009, U+202F, ±, ³, ↑, ↓ |
| GSUB features | aalt calt ccmp dnom frac liga numr ordn salt sinf ss01-ss05 sups zero | aalt case ccmp dlig frac liga ordn rlig sinf ss01 ss02 subs sups | aalt ccmp liga ordn ss01 |
| GPOS under `thai` | kern mark mkmk | mark mkmk (no kern; `latn` has kern) | dist kern mark mkmk |
| tnum / lnum / pnum / onum | none | none | none |
| case / zero | none / works | case alters only the grave, dieresis, macron and acute accents / none | none / none |
| Stylistic sets | ss01 a, ss02 g, ss03 and ss04 zero forms; ss05 not examined | ss01 alternates for ข ฃ ฆ ฉ ช ซ ญ ฑ ใ (subtle), ss02 ญ | ss01 alternates for ญ ฐ |
| ASCII digits (default) | tabular 600, lining | tabular 600, lining | tabular 572 (472 at wdth 75), lining |
| Minus / plus / hyphen | 600 / 600 / 399 | 600 / 600 / 399 | 322 / 572 / 322 |
| UPM; hhea asc/desc/gap | 1000; 1116 / -534 / 0 | 1000; 1025 / -275 / 0 | 1000; 1061 / -450 / 0 |
| typo; win; USE_TYPO | = hhea; 1116/534; on | = hhea; 1119/324; on | = hhea; 1061/450; on |
| `line-height: normal` (hhea) | 1.65 | 1.30 | 1.511 |
| Latin x-height / cap (measured ink) | 0.516 / 0.698 | 0.516 / 0.698 | 0.536 / 0.714 |
| Thai body น / tall ป / body ÷ x-height | 0.546 / 0.722 / 1.058 | 0.570 / 0.780 / 1.105 | 0.548 / 0.745 / 1.022 |
| National sentence advance at 400 | 31.556 em | 30.864 em | 30.422 em (25.55 at wdth 75) |
| Tallest real-word ink at 400 | 1.083 em, น้ำล้นตลิ่ง | 1.066 em, น้ำล้นตลิ่ง (**41 units above hhea ascender**) | 1.009 em, น้ำล้นตลิ่ง |
| Deepest real-word ink at 400 | -0.350 em, ผู้ใหญ่บ้าน | -0.267 em, ผู้ใหญ่บ้าน | -0.265 em, ผู้ใหญ่บ้าน |
| Real words: ink box / collide | 1.584 / 1.433 | 1.382 / 1.333 | 1.407 / 1.274 |
| Stress: top / bottom | 1.125 ป็๋ / -0.415 ฎู | 1.077 ก็๊ / -0.324 ฎฺ | 1.082 ฬ็่ / -0.432 ฎู |
| Stress: ink box / collide | 1.668 / 1.540 | 1.404 / 1.401 | 1.553 / 1.514 |
| Loop test | 0/17 (loopless) | 0/17 | 0/17 |
| WOFF2 subset, 400 | 22.9 KB hinted / 16.5 KB unhinted | 17.2 KB (instance) | 13.8 KB (instance) |
| WOFF2 subset, 600 | 23.9 / 17.4 KB (SemiBold) | 18.7 KB (instance) | 14.0 KB (instance) |
| WOFF2 subset, all weights | one file per weight; upstream Text 450: 23.6 / 17.3 KB | 46.0 KB (full VF) | 44.1 KB (wght VF, wdth pinned at 100); 77.7 KB (full VF) |
| GF popularity | 95 (IBM Plex Sans: 8) | 121 | 45 (Noto Sans: 5; Open Sans: 2) |
| google/fonts directory, last commit | `ofl/ibmplexsansthai` d935ac9235ee, 2026-02-27 | `ofl/anuphan` 95f4904fc8bc, 2026-03-03 | `ofl/notosansthai` 8b0a1d0f5983, 2026-03-12 |

### 5.2 Ink and leading by weight (real words)

These are ink box / collide, from the per-weight JSON reports in `.cache/fonts/_report/`. Plex Sans Thai 450 is the upstream Text weight. Anuphan and Noto are variable instances, with Noto at wdth 100.

| wght | Plex Sans Thai | Anuphan | Noto Sans Thai | Noto Serif Thai |
|---|---|---|---|---|
| 300 | not measured | 1.328 / 1.296 | 1.377 / 1.257 | 1.572 / 1.344 |
| 400 | 1.584 / 1.433 | 1.382 / 1.333 | 1.407 / 1.274 | 1.582 / 1.349 |
| 450 | 1.608 / 1.447 | not measured | not measured | not measured |
| 500 | 1.640 / 1.464 | 1.406 / 1.359 | 1.425 / 1.283 | 1.582 / 1.349 |
| 600 | 1.676 / 1.485 | 1.432 / 1.387 | 1.445 / 1.293 | not measured |
| 700 | 1.714 / 1.506 | 1.496 / 1.433 | 1.471 / 1.306 | not measured |

How to read this:
- At body line-height 1.6, real-word ink stays inside the line box for every UI candidate at 400.
- For Plex at 500 and above, the ink box is exceeded, but collide (1.464-1.506) is still below 1.6. Lines will not touch, but clipped containers will cut marks.
- For a display sentence set at 1.3:
  - Noto Sans Thai is clear from 300 to 600 (collide up to 1.293) and just over at 700 (1.306).
  - Anuphan is clear only at 300 (1.296); at 400 it needs 1.333.
  - Plex is not guaranteed clear at any weight measured (1.433 and up).

### 5.3 Latin harmonisation

| Face | Ships own Latin? | Where the Latin comes from (measured with `compare_outlines.py`) | Note |
|---|---|---|---|
| IBM Plex Sans Thai | Yes | The Plex Sans design, drawn by the same team. Not outline-compared against the standalone IBM Plex Sans files (unverified) | Plex Sans is widely used (popularity 8) and is IBM's corporate face |
| Anuphan | Yes | Matches the Latin in IBM Plex Sans Thai: 15 of 62 alphanumerics point-for-point identical; the others have the same advance and either the same bounding box or the same contour structure, except u at 400 (2 units narrower). 62/62 match at 700 | So A and B share one Latin |
| Noto Sans Thai | Yes | Identical to Noto Sans (62/62). Noto Sans against Open Sans: 40 near (within 15/1000 em), 6 same structure, 16 different | The most familiar web Latin of the three |
| Noto Serif Thai | Yes | Differs entirely from Noto Sans Thai's Latin (0/62 shared). Not compared against Noto Serif (unverified) | |
| Taviraj, Trirong, Pridi | Yes | Drawn with the Thai by Cadson Demak and Katatrad, per name table and Google Fonts descriptions: serif (Taviraj), transitional serif (Trirong), slab serif (Pridi) | Not compared against other families |
| IBM Plex Mono | Latin only | Plex family; digits 600, 8 of 10 within 15 units of Plex Sans Thai's | |

### 5.4 Sentence serifs (all looped Thai, 17/17 on the loop test)

| | Noto Serif Thai | Taviraj | Trirong | Pridi |
|---|---|---|---|---|
| Designer, foundry (name table) | Monotype Design Team; Monotype Imaging | Katatrad Team; CadsonDemak | Katatrad Team; CadsonDemak | Katatrad Team; CadsonDemak |
| Licence | OFL 1.1, no RFN | OFL 1.1, no RFN (© 2015 Cadson Demak) | same | same |
| Google Fonts file version | Version 2.002 | Version 1.001 | Version 1.001 | Version 1.001 |
| Upstream | notofonts/thai `NotoSerifThai-v2.002` | github.com/cadsondemak/taviraj | github.com/cadsondemak/trirong | github.com/cadsondemak/pridi |
| Format | Variable: wght 100-900, wdth 62.5-100; no italic | Static, 9 weights plus italics | Static, 9 weights plus italics | Static, 6 upright weights (200-700), **no italics** |
| Glyphs | 467 | 772 | 772 | 773 |
| Thai / Basic Latin / Latin-1 | 87 / 95 / 84 | 87 / 95 / 95 | 87 / 95 / 95 | 87 / 95 / 95 |
| Missing from symbol set | U+2009, U+202F, ±, ³, ↑, ↓ | U+2009, U+200B, U+202F, U+25CC | same as Taviraj | same as Taviraj |
| Google Fonts description, in its own words | "a modulated (“serif”) design for texts" | "wide structure", "well-suited for formal usage", "farangses" genre | "narrow and tall structure", "Transitional serif Latin works well in formal contexts" | "slab serif Latin", Thai "designed specifically within loopless terminal concepts"; classified SERIF and DISPLAY |
| Numeric features | none; ss01 = ญ ฐ alternates | none | none | none |
| ASCII digits | tabular 559 | tabular 581 (597 at Medium) | tabular 587 (602 at Medium) | **proportional** 406-621 |
| Minus U+2212 | 310 | 582 | 588 | 412 |
| hhea asc/desc | 1064 / -534 | 1172 / -534 | 1200 / -534 | 1100 / -450 |
| typo; USE_TYPO | = hhea; on | 750/-250/200; off | 750/-250/200; off | 750/-250/200; off |
| `line-height: normal` | 1.598 | 1.706 | 1.734 | 1.55 |
| Thai body ÷ x-height | 1.034 | 1.198 | 1.133 | 1.133 |
| National sentence advance at 400 | 32.334 em | 31.467 em | 30.075 em | 29.743 em |
| Real ink top / bottom at 400 | 1.056 / -0.293 | 1.126 / -0.337 | 1.118 / -0.339 | 1.054 / -0.326 |
| Real ink box / collide: Light, 400, Medium | 1.572/1.344, 1.582/1.349, 1.582/1.349 | 1.556/1.431, 1.614/1.463, 1.620/1.468 | 1.554/1.446, 1.570/1.457, 1.586/1.468 | 1.434/1.375, 1.458/1.380, 1.454/1.370 |
| Stress ink box / collide at 400 | 1.746 / 1.610 | 1.744 / 1.680 | 1.720 / 1.680 | 1.518 / 1.495 |
| WOFF2 subset, 400 | 16.0 KB (instance); wght VF 55.3 KB; full VF 97.2 KB | 27.9 KB hinted / 16.6 KB unhinted | 26.5 / 16.1 KB | 25.6 / 14.8 KB |
| GF popularity | 145 | 219 | 144 | 109 |
| google/fonts directory, last commit | 8b0a1d0f5983, 2026-03-12 | 037aad08f644, 2026-04-24 | 3311d9477232, 2026-04-24 | 1334e4632823, 2026-04-24 |

"Light" for Noto Serif Thai means the wght 300 instance. The renders of all serifs are in `.cache/fonts/_render/serif.svg.png`, with the national sentence at line-height 1.4.

### 5.5 IBM Plex Mono (optional, raw readouts only)

| | Google Fonts `ofl/ibmplexmono` | Upstream `@ibm/plex-mono@2.5.0` (2026-06-11) |
|---|---|---|
| Version, glyphs | Version 2.3, 1033 | Version 2.005, 1207 |
| Licence | OFL 1.1, RFN "Plex" | same |
| Format | Static, 100-700 plus italics | Static |
| Thai | U+0E3F ฿ only | U+0E3F only |
| Numeric features | none (`zero`, ss01-ss05, salt present) | `lnum`, `onum` (onum works), plus the same sets |
| Missing from symbol set | U+2009, U+200B, U+202F, U+25CC | U+25CC |
| Every glyph | 600 wide | 600 wide |
| hhea asc/desc; typo; USE_TYPO | 1025 / -275; = hhea; on | 1025 / -275; 780/-220/300; off |
| WOFF2 subset (Latin part only) | 10.3 KB hinted / 6.7 KB unhinted | 11.6 / 7.7 KB |

Inline mixing caveat: Plex Mono's hhea (1025/-275) differs from Plex Sans Thai's (1116/-534). A mono run inside a sans line changes the content area. The specimen must check the baseline and line-box height, or normalise with `ascent-override` and `descent-override` in `@font-face`.

## 6. Rejected faces

**Loopless, but rejected:**
- **LINE Seed Sans TH** (seed.line.me, Version 1.003; Latin by Dalton Maag, Thai by Cadson Demak)
  - Proportional digits (402-656) and no `tnum`.
  - Latin-1 coverage is 18/96.
  - It is the brand face of LY Corporation's LINE.
  - The zip ships no `OFL.txt`; OFL is stated only in name ID 13 and on the site.
- **Athiti** (Cadson Demak, Version 1.033)
  - Proportional digits, no `tnum`.
  - Its Google Fonts description calls it an "informal sans Latin" based on "humanist calligraphy".
- **Kanit**
  - Proportional digits, no `tnum`.
  - Popularity 16, among the most used Thai faces on Google Fonts, which makes it a generic Thai web look.
- **Prompt**
  - Proportional digits, no `tnum`.
  - Popularity 15, for the same reason as Kanit.
- **Mitr**
  - Proportional digits, no `tnum`.
  - Rounded, informal.
- **Pattaya.** A display script face (Lobster extended to Thai).

**Looped sans (the brief requires loopless for the UI):**
- **Sarabun.** Looped, and its Google Fonts description says "It is the 'TH Sarabun New' font" and that it "is used in the Government Gazette of Thailand". It is the official document look the product brief rules out.
- **Google Sans.** Looped Thai.
  - Its `TRADEMARKS.md` restricts use of the names "Google" and "Google Sans".
  - It is Google's own brand face.
  - The variable font is about 5 MB.
  - It is the only Thai face here with working `tnum`, `pnum`, `case` and `zero`.
- **IBM Plex Sans Thai Looped and Noto Sans Thai Looped.** Looped versions of candidates A and C.
- **Bai Jamjuree.** Square loops. I first assumed it was loopless; the loop test proved that wrong.
- **Chakra Petch, K2D, Niramit, Fahkwang, KoHo, Kodchasan, Thasadith:** looped (17/17).
- **Krub.** Looped. The test scores 3/17 because its loops are open, but the render shows them.

**Serif:**
- **Maitree.**
  - Google Fonts says it suits "works that require an antique and historical manner".
  - Proportional digits.
  - Wide, with large loops.

**Out of genre:** Sriracha, Itim, Mali, Charm, Charmonman and Playpen Sans Thai (handwriting), and Chonburi and Srisakdi (display).

**Banned by the brief and not assessed:** Inter, Geist, Roboto, Arial, system UI.

**Coverage of this search:**
- Google Fonts lists 1959 families. 34 have a Thai subset: 21 sans, 5 serif, 6 handwriting, 2 display.
- All 21 sans and all 5 serifs were inspected.
- Loopless sans: IBM Plex Sans Thai, Anuphan, Noto Sans Thai, Kanit, Prompt, Mitr, Athiti and Pattaya. Everything else is looped.

## 7. Phase 3 specimen checklist

Build `/specimen` so that each pairing renders the same blocks, with the sentence face switchable (none, Noto Serif Thai, Pridi, Trirong). Screenshot at 390×844 and 1440×900 into `docs/design/screens/phase-3/`.

**Strings (real, from the briefs):**
- [ ] National sentence, both brief variants:
  - "ตอนนี้ 22 สถานีน้ำล้นตลิ่ง ส่วนใหญ่อยู่ลุ่มเจ้าพระยาตอนล่างและภาคใต้ตอนบน"
  - "ตอนนี้ 22 สถานีน้ำล้นตลิ่ง ส่วนใหญ่อยู่ที่พระนครศรีอยุธยา สุราษฎร์ธานี และชุมพร"
  - Also the counts 0, 1 and 11 as the humanize generator writes them.
- [ ] Station sentences:
  - "แม่น้ำเจ้าพระยาใกล้คุณที่สุด (สถานีบางไทร, 3.2 กม.) ตอนนี้ต่ำกว่าตลิ่ง 1.40 เมตร สูงขึ้น 12 ซม. ใน 6 ชั่วโมงที่ผ่านมา · วัดเมื่อ 14:00 น."
  - "สูงกว่าตลิ่ง 0.35 เมตร (น้ำล้นตลิ่ง)"
  - "ข้อมูลล่าสุดเมื่อ 09:00 น. (ยังไม่อัปเดต)"
  - "สถานีนี้ยังไม่มีข้อมูลระดับตลิ่ง"
- [ ] Long names:
  - จังหวัดพระนครศรีอยุธยา
  - สุราษฎร์ธานี
  - จ.ปราจีนบุรี
  - อ.ศรีสัชนาลัย
  - บ้านช้างแรก (Gt.11)
- [ ] Stack stress: น้ำล้นตลิ่ง, ผู้ใหญ่บ้าน, ปิ่นเกล้า, ปั๊มน้ำ, ฟื้นฟู, ญี่ปุ่น, ฤๅษี, กฎหมาย, ปฏิบัติ.
- [ ] Situation labels: น้ำล้นตลิ่ง, น้ำมาก, น้ำปกติ, น้ำน้อย, น้ำน้อยวิกฤติ.
- [ ] The wordmark "มหานที" at display size, in each UI face and weight.
- [ ] Dates and times: "29 ต.ค. 69", "14:00 น.".

**Vertical rhythm:**
- [ ] National sentence at the phone display size and at the desktop size (6rem at most), at line-height 1.2, 1.3, 1.4 and 1.5.
  - Look for ญ, ู and ุ on one line touching น้ำ, ปั๊ม or ปิ่น on the next.
  - Compare with the collide numbers in 5.2. Use real wraps, not forced ones.
- [ ] Body text at 1.5, 1.6 and 1.7. Confirm 1.6 at each weight in use.
- [ ] Clipping:
  - [ ] `overflow: hidden`
  - [ ] `-webkit-line-clamp`
  - [ ] `text-overflow: ellipsis`
  - [ ] `background-clip: text`
  - [ ] `line-height: normal`
  - Test these with each UI face, Anuphan first: its ink rises above the ascender.

**Numbers:**
- [ ] A column of 0.35, 1.40, 12, 102.35, −0.35 and +0.35 at 400 and 600. Check digit alignment, and the minus sign (U+2212) against the hyphen.
- [ ] Confirm that `font-variant-numeric: tabular-nums lining-nums` changes nothing in the shortlisted faces. Overlay it with the declaration on and off.
- [ ] "11 สถานี" and "1.40" inside the sentence. Tabular 1s look loose in prose, and no candidate offers proportional figures.
- [ ] Number and unit joined with U+00A0. No U+2009 or U+202F anywhere, because no face has them.

**Weight and contrast:**
- [ ] Weights 400, 450, 500 and 600 on the dark background:
  - 450 is the upstream Plex Text weight, or an Anuphan or Noto instance.
  - Check on a laptop at full brightness in a lit room, and on a phone at low brightness.
  - Light text on dark reads heavier. Decide the body weight from this test, not from the light-theme habit.
- [ ] Legibility at 12-14 px on a phone for loopless look-alikes in real names: บ/ป, ถ/ภ/ฤ, ฎ/ฏ, ข/ช.
- [ ] Hierarchy: at least three clear size steps and two weight steps. No crushed tracking (leave `letter-spacing` at 0 for Thai).

**Mixing:**
- [ ] Latin inside Thai lines ("Gt.11", "MSL", "ThaiWater"): x-height and baseline.
- [ ] Plex Mono inline in a Plex Sans Thai or Anuphan line, if mono is still in play: baseline and line-box height.
- [ ] Pairing C: Noto Sans Thai at wdth 75 and 87.5 for legend and timeline labels.
- [ ] Serif: upright only. Set `font-synthesis: none` (or at least `font-synthesis-style: none`) so `<em>` or a missing weight never produces a synthetic oblique or bold.

**Wrapping and loading:**
- [ ] Thai line breaks of the sentence at 390 px, with and without `text-wrap: balance` and `text-wrap: pretty`. Check that no break splits "14:00 น.".
- [ ] `font-display: swap`: measure the layout shift of the sentence. Try a metric-matched fallback `@font-face` (`size-adjust`, `ascent-override`, `descent-override`).
- [ ] Preload only the face that sets the first-frame sentence and the UI regular. Confirm the LCP effect with a trace (unverified here).
- [ ] Chrome, Firefox and Safari. Safari shapes with Core Text, not HarfBuzz, so mark positions may differ from these measurements.

**Build (for the chosen pairing):**
- [ ] Subset with the `UNICODES` in `scripts/research/subset_sizes.sh`. Keep `--layout-features='*'` unless a trimmed list is proven safe for Thai.
- [ ] Keep name IDs 0, 13 and 14 (`--name-IDs=0,1,2,3,4,5,6,13,14`), or ship `OFL.txt` beside the fonts.
- [ ] If Plex is chosen:
  - Rename name IDs 1, 3, 4, 6, 16 and 17 before subsetting, so that "Plex" does not appear.
  - Use a CSS family name without "Plex".
  - Do not use "IBM" to promote the product (OFL condition 4).
- [ ] Compare the shipped WOFF2 sizes against the estimates in section 5.

## 8. Download instructions and pinned files

**Pin.** google/fonts at commit `58466038e19d5a14053e0c9f863715e74af0fe0e` (2026-10-29T09:12:06Z). Raw URLs follow this pattern, with `[` and `]` encoded as `%5B` and `%5D`:

`https://raw.githubusercontent.com/google/fonts/58466038e19d5a14053e0c9f863715e74af0fe0e/ofl/<dir>/<file>`

On 2026-10-29 the pinned URL of the Regular or variable file of each shortlisted family, Plex Sans Thai's `OFL.txt`, and every upstream URL below returned HTTP 200. The other files in the commands follow the same pattern and were downloaded the same way during this research. The Anuphan, Plex Sans Thai Regular and Noto Sans Thai files served at the pin match the sha256 values below.

**Google Fonts web download.** `https://fonts.google.com/download?family=…` now returns an HTML page, not a zip. `https://fonts.google.com/download/list?family=Anuphan` returns a JSON manifest of fonts.gstatic.com URLs. The pinned GitHub raw files are the reproducible route.

```sh
# From the project root. One directory per family; downloads are data, never executed.
G=https://raw.githubusercontent.com/google/fonts/58466038e19d5a14053e0c9f863715e74af0fe0e/ofl
mkdir -p .cache/fonts/ibmplexsansthai .cache/fonts/anuphan .cache/fonts/notosansthai \
         .cache/fonts/notoserifthai .cache/fonts/pridi .cache/fonts/trirong .cache/fonts/ibmplexmono
for w in Regular Medium SemiBold; do
  curl -fsSLo ".cache/fonts/ibmplexsansthai/IBMPlexSansThai-$w.ttf" "$G/ibmplexsansthai/IBMPlexSansThai-$w.ttf"
done
curl -fsSLo ".cache/fonts/anuphan/Anuphan[wght].ttf" "$G/anuphan/Anuphan%5Bwght%5D.ttf"
curl -fsSLo ".cache/fonts/notosansthai/NotoSansThai[wdth,wght].ttf" "$G/notosansthai/NotoSansThai%5Bwdth,wght%5D.ttf"
curl -fsSLo ".cache/fonts/notoserifthai/NotoSerifThai[wdth,wght].ttf" "$G/notoserifthai/NotoSerifThai%5Bwdth,wght%5D.ttf"
for w in Light Regular; do
  curl -fsSLo ".cache/fonts/pridi/Pridi-$w.ttf" "$G/pridi/Pridi-$w.ttf"
  curl -fsSLo ".cache/fonts/trirong/Trirong-$w.ttf" "$G/trirong/Trirong-$w.ttf"
done
curl -fsSLo .cache/fonts/ibmplexmono/IBMPlexMono-Regular.ttf "$G/ibmplexmono/IBMPlexMono-Regular.ttf"
for d in ibmplexsansthai anuphan notosansthai notoserifthai pridi trirong ibmplexmono; do
  curl -fsSLo ".cache/fonts/$d/OFL.txt" "$G/$d/OFL.txt"
done
shasum -a 256 .cache/fonts/*/*.ttf
```

**sha256 of the inspected files (google/fonts at the pin):**

| File (`ofl/<dir>/…`) | sha256 |
|---|---|
| ibmplexsansthai/IBMPlexSansThai-Regular.ttf | eee061d1bac39be40f9bb94498898cd1894f5d3d9a2f0ee6cbd48a6bd03d052b |
| ibmplexsansthai/IBMPlexSansThai-Medium.ttf | 46ba1b7365cc2c583914f23914cce46e836a72662e34f6e3029b15aaa10b792e |
| ibmplexsansthai/IBMPlexSansThai-SemiBold.ttf | 9d7526a0c8dfad63c49815d0cce9eea1026f018468b81c579997e953620a67d3 |
| ibmplexsansthai/IBMPlexSansThai-Bold.ttf | f88fbe0a3717c93deddf4b13e50a05c9adcd38f58d675d9d5d0b869961851667 |
| ibmplexsansthai/IBMPlexSansThai-Light.ttf | b8e2da56ec11f1965af59a7119ee0b9b78773a9c9a89ae049b4975c9197b5c8d |
| ibmplexsansthai/IBMPlexSansThai-ExtraLight.ttf | b1ba0c97d959657832c1a0929c274e4d4d8128ca2c52d24c6835cf87fb4a5a04 |
| ibmplexsansthai/IBMPlexSansThai-Thin.ttf | 65f2cd4ed6f23048622bba1a1062c05c2ff80e603d202fdccf303cf3d5d085e5 |
| anuphan/Anuphan[wght].ttf | 7ec47ebcaa9d6459d974691cbb42fae52e1449478a607b071bfaf3ccdb433ce4 |
| notosansthai/NotoSansThai[wdth,wght].ttf | 5a1c559bb539583c8a1fd99d1c5b9491e5e14478c9cd2bd0970d5c3096cc9ef8 |
| notoserifthai/NotoSerifThai[wdth,wght].ttf | 34a7ad11647c845303aabdde639059806c56b84719e5d2ceb28eb038711bdf53 |
| taviraj/Taviraj-Light.ttf | 57670abb418f67f90031eb073e7afb69d98e041b7ef824a0d68981104f13beed |
| taviraj/Taviraj-Regular.ttf | 077e6299711fddca973f9d43455949018459c2d29c6e790666d41017ee1a27d5 |
| taviraj/Taviraj-Medium.ttf | 126a0ed061ff96c499b314bdcdacfa1c2ccd153fcd78e4fbb3575c88e98dae3d |
| trirong/Trirong-Light.ttf | 4302056ccc54744ae2d9b80f34b6377c6858cb95cd534ce2531fb13a4a780793 |
| trirong/Trirong-Regular.ttf | 3cd19b35a24df5fcf12c1f917324cf821d5aa582f7ea36fb0d2a91ba5e8ccd05 |
| trirong/Trirong-Medium.ttf | 7d9baefbf837849e8767f8649f19dd6eb9ef83d5f63baca0268efdc566e2fe03 |
| pridi/Pridi-Light.ttf | 3f328cfdface42003202df36d275b0d3498c7484dd384add3fbb066c5829dab0 |
| pridi/Pridi-Regular.ttf | cc5a2ea9660d78fc64c909c29582ab8bd7195128f81bed825d7cee985369a4da |
| pridi/Pridi-Medium.ttf | d38ca6dd91f8ede51cb793dc05c56c2381a4933a8849ddc09281ecb50943294b |
| ibmplexmono/IBMPlexMono-Regular.ttf | 6a3412f058c7d8dfd9170c41e85ade48e5156ecb89356110ca57a0a27734af46 |
| ibmplexmono/IBMPlexMono-Medium.ttf | a9b4c49bb299e05b5f6c481e7fb5e78943d2793249a0c8874ab574a2d1ea6755 |

**Upstream alternatives:**

| What | URL | sha256 | Notes |
|---|---|---|---|
| IBM Plex Sans Thai 1.1.0 (files say Version 1.2) | https://github.com/IBM/plex/releases/download/%40ibm%2Fplex-sans-thai%401.1.0/ibm-plex-sans-thai.zip | d7203f43c20f9abd40487f845c48db4077d2056ea18632c8959591c6815d7fb9 | downloaded and inspected; Text weight at `ibm-plex-sans-thai/fonts/complete/ttf/IBMPlexSansThai-Text.ttf`; licence at `ibm-plex-sans-thai/LICENSE.txt`; no variable font |
| IBM Plex Mono 2.5.0 (Version 2.005) | https://github.com/IBM/plex/releases/download/%40ibm%2Fplex-mono%402.5.0/ibm-plex-mono.zip | 6d23f01257663d8cc49a0d64c22ced630b79e0e2a0ac08a0da86e9a38bbc481c | downloaded and inspected; adds `lnum` and `onum`, U+2009, U+202F |
| Noto Sans Thai v2.002 | https://github.com/notofonts/thai/releases/download/NotoSansThai-v2.002/NotoSansThai-v2.002.zip | af889cc673fc714060ce5e4e088fbad32aa4c0571a19958efeaff128a22da485 | downloaded and inspected; wght-only VF at `NotoSansThai/full/slim-variable-ttf/NotoSansThai[wght].ttf` (459 glyphs); `OFL.txt` at the zip root |
| Noto Serif Thai v2.002 | https://github.com/notofonts/thai/releases/download/NotoSerifThai-v2.002/NotoSerifThai-v2.002.zip | not downloaded | URL returned 200 |
| Anuphan source | https://github.com/cadsondemak/Anuphan/tree/a8ee8448324cc13fb69d585c4f6da4994f7652b0/Fonts | n/a (repository tree) | the commit Google Fonts built from (METADATA.pb) |

**Subsetting (the command used for the size estimates, plus the licence name IDs):**

```sh
.cache/venv/bin/python -I -m fontTools.subset IN.ttf \
  --unicodes="U+0020-007E,U+00A0,U+00B0,U+00B1,U+00B3,U+00B7,U+00D7,U+0E01-0E3A,U+0E3F-0E5B,U+2009,U+200B,U+202F,U+2013,U+2018,U+2019,U+201C,U+201D,U+2026,U+2212,U+25CC" \
  --layout-features='*' --flavor=woff2 --name-IDs=0,1,2,3,4,5,6,13,14 \
  --output-file=OUT.woff2
# For a variable-font instance first:
.cache/venv/bin/python -I -m fontTools.varLib.instancer IN.ttf wght=400 --output INSTANCE.ttf
```

Use one file per weight covering both Thai and Latin. Every view shows both scripts, so a `unicode-range` split would only add requests. Self-hosting keeps the CSP at `font-src 'self'`.

## 9. Not verified

- **Browser rendering.** All shaping and ink numbers come from HarfBuzz 12.1.0. Chrome and Firefox are expected to match, but this was not tested. Safari uses Core Text and was not tested.
- **Hinting.** Windows rendering of hinted against unhinted subsets.
- **Colour and weight.** On-screen weight on the dark theme. Every render here is black on white.
- **Character judgements.** "Engineered", "rounder", "dark" and similar words come from HarfBuzz renders at one size, on white.
- **Plex Latin identity.** Whether IBM Plex Sans Thai's Latin is outline-identical to the standalone IBM Plex Sans. It was not compared.
- **Noto Serif Thai's Latin provenance.** It was not compared against Noto Serif.
- **IBM Plex Sans Thai ss05** was not examined.
- **LINE Seed Sans TH's Reserved Font Name status.** The zip has no `OFL.txt`.
- **Google Fonts and the Plex RFN.** Whether Google Fonts has a separate agreement with IBM that covers serving Plex subsets under the Plex name.
- **Noto as system fallback.** Whether Noto Sans Thai is the system Thai fallback on current Android or other platforms.
- **Popularity field.** The exact meaning of Google Fonts' `popularity` field.
- **U+200B in Anuphan.** That a U+200B inside Anuphan text is harmless (it is default-ignorable), and how browsers pick hhea, typo or win per platform.
- **Pridi Light size.** Its WOFF2 subset size was not measured; only Regular was.
- **Preloading.** The LCP effect of preloading.

## 10. Sources

All accessed 2026-10-29.

- https://github.com/google/fonts, at commit 58466038e19d5a14053e0c9f863715e74af0fe0e. Directory listings came via https://api.github.com/repos/google/fonts/contents/ofl/… and files via https://raw.githubusercontent.com/google/fonts/58466038e19d5a14053e0c9f863715e74af0fe0e/ofl/…
  - Files read: `METADATA.pb`, `DESCRIPTION.en_us.html`, `OFL.txt`, `upstream_info.md` and the font files.
  - Directories: ibmplexsansthai, ibmplexsansthailooped, anuphan, notosansthai, notosansthailooped, notoserifthai, taviraj, trirong, pridi, maitree, sarabun, athiti, kanit, prompt, mitr, baijamjuree, thasadith, chakrapetch, k2d, krub, niramit, fahkwang, koho, kodchasan, pattaya, googlesans (including `TRADEMARKS.md`), notosans, opensans, ibmplexmono.
  - Accessed 2026-10-29.
- https://api.github.com/repos/google/fonts/commits (last commit per family directory), accessed 2026-10-29
- https://fonts.google.com/metadata/fonts, accessed 2026-10-29
- https://fonts.google.com/download/list?family=Anuphan, accessed 2026-10-29
- https://api.github.com/repos/IBM/plex/releases, accessed 2026-10-29
- https://github.com/IBM/plex/releases/download/%40ibm%2Fplex-sans-thai%401.1.0/ibm-plex-sans-thai.zip, accessed 2026-10-29
- https://github.com/IBM/plex/releases/download/%40ibm%2Fplex-mono%402.5.0/ibm-plex-mono.zip, accessed 2026-10-29
- https://api.github.com/repos/notofonts/thai/releases, accessed 2026-10-29
- https://github.com/notofonts/thai/releases/download/NotoSansThai-v2.002/NotoSansThai-v2.002.zip, accessed 2026-10-29
- https://api.github.com/repos/cadsondemak/Anuphan/commits and the repository tree at a8ee8448324cc13fb69d585c4f6da4994f7652b0, accessed 2026-10-29
- https://seed.line.me/index_en.html, accessed 2026-10-29
- https://seed.line.me/src/images/fonts/LINE_Seed_Sans_TH.zip, accessed 2026-10-29
- https://api.github.com/search/repositories (search for an official LINE Seed source repository; none found, third-party mirrors not used), accessed 2026-10-29
- https://openfontlicense.org/ofl-faq/ (entries 2.6 to 2.8), accessed 2026-10-29
- https://www.cadsondemak.com/medias/read/design-like-a-bilingual-ibm-plex-thai ("Design like a Bilingual - IBM Plex Thai", Kornkanok Tantisuwanna, published 2021-09-27), accessed 2026-10-29
