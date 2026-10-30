# Design direction: มหานที

> This is the most important brief. The judges design websites for a living and have publicly complained that AI-made work "looks the same". The bar is a site that could sit on an Awwwards or Information is Beautiful shortlist and that no one would guess came from a prompt. Premium here comes from **cartographic truth, typographic craft and restraint**, not from effects.

## 1. The world (committed, do not average it away)
**A night-time hydrographic instrument.** Picture the national survey's plaster relief model of Thailand in a dark map room, lit only where water is being measured right now. It is calm, exact and quietly dramatic. Everything visible is either geography, data or the words that explain the data.

**Use scene → dark theme.** The primary user checks at night on a phone, so the theme is dark. The demo happens in a bright interview room on a laptop, so text contrast and gauge legibility must survive a bright room. Do not use ultra-low-contrast greys.

## 2. References: what to take (never copy their look)
| Reference | Take this |
|---|---|
| earth.nullschool.net | The data *is* the visual. Chrome is nearly absent, with one typographic overlay. |
| Mini Tokyo 3D | Live system on real geography. The camera follows the entity you selected. |
| Eduard Imhof's Swiss relief cartography | Hillshade lit from the north-west, aerial perspective, restrained hypsometric tint. Relief that feels carved, not "3D-rendered". |
| Reuters / NYT 3D flood and terrain graphics | Explanatory camera moves and annotations set in good type at the right moment. |
| Apple Weather map layers | Premium comes from motion quality, type and restraint. |

Do **not** look at "3D globe hero" templates, Dribbble dashboards or SaaS landing pages.

## 3. Composition
- **The relief is the page.** There is no landing page, hero section, feature cards, testimonials or marketing footer. The first frame *is* the product: Thailand in relief, the current state of every gauge, and one computed sentence about the country now, for example *"ตอนนี้ 22 สถานีน้ำล้นตลิ่ง ส่วนใหญ่อยู่ลุ่มเจ้าพระยาตอนล่างและภาคใต้ตอนบน"*.
- **One primary action:** **ใกล้ฉัน**. Secondary actions are a search field, the timeline and an about page.
- **Answer panel:** the human sentence is set large, as the typographic centrepiece. Underneath are the measured numbers in tabular figures, the measured time, the source and a small 72-hour chart against bank-full.
  - On a phone it is a bottom sheet with snap points.
  - On desktop it is a left-hand column floating over the relief with generous margins, and the relief shifts right to make room.
- **Thailand is tall.** Portrait phones fit it naturally. On desktop, compose it off-centre.
- **Wordmark:** "มหานที" set in type with care. No icon logo and no gimmick.

## 4. The 3D relief: specification and taste
- **Real elevation:** Thailand plus its neighbours from a DEM. Neighbours are desaturated and recede through aerial perspective rather than being cut out abruptly. The Thai border is a fine, precise line. Rivers come from real geodata, with line width set by river order.
- **Vertical exaggeration:** the central plain where floods happen is very flat, so exaggerate about 6–15×. Choose the value in look-dev by eye and with screenshots, and record it in an ADR.
- **Lighting:** one directional key light from the north-west (cartographic convention) plus soft fill.
  - No bloom by default, and no stock "city" HDRI.
  - No auto-rotate and no idle camera wobble.
- **Material:** matte and slightly warm, like sediment or plaster, so it reads as a carved surface. A subtle hypsometric tint is allowed only if it stays quiet. The sea is dark and calm, not glowing.
- **Gauges (the signature mark):** each station is a slim vertical gauge, not a glowing pillar.
  - A thin cap marks **bank-full height**, and the **water fill** sits inside it, so the encoding is "how full is this river relative to its banks".
  - When water passes bank-full, the fill rises *above the cap* in the alert colour. This should be readable at a glance and correct by construction.
  - Use hierarchy and level of detail. At country zoom, levels 4 and 5 are prominent, levels 1–3 are small and quiet, and stale stations are hollow. As you zoom in, gauges grow and gain labels.
  - 800 identical glowing pillars is the slop version. Do not do that.
- **Camera:**
  - Default is an oblique view.
  - Orbit is clamped: polar angle limits, target bounded to Thailand, and no losing the horizon.
  - Flights use exponential ease-out over 1.2–2.0 s and can be interrupted by the user.
  - `prefers-reduced-motion` cuts instead of flying.
- **Thai text never goes into WebGL text** unless Thai shaping (stacked tone marks and vowels) is verified. Default to HTML/CSS overlays projected from 3D positions.
- **Atmosphere:** subtle depth fog or aerial perspective for scale. No starfield, no particles, no halos, no lens flare.
- **Antialiasing on.** No z-fighting on borders or rivers. Use adaptive DPR. The renderer idles (`frameloop="demand"`) when nothing moves.

## 5. Colour
- **Build the palette from the subject:**
  - night slate (never pure #000)
  - the ochre-brown of Chao Phraya floodwater and sediment
  - muted paddy and forest greens for the land tint
  - one disciplined alert colour
- **The situation scale is redesigned, not inherited.** ThaiWater's five colours (red, blue, green, yellow, dark-red) are not ordered and clash.
  - Design an *ordered* scale where saturation and lightness rise with danger.
  - Only "น้ำล้นตลิ่ง" (level 5) and "น้ำมาก" (level 4) get strong colour.
  - "น้ำน้อยวิกฤติ" (level 1) is a drought concern, so give it a distinct but muted hue, not the same family as flood.
  - It must pass a deuteranopia and protanopia simulation, so provide evidence in the look-dev doc.
- **Single source of truth:** tokens are CSS custom properties *and* a TS module consumed by the shaders. No hard-coded colours in components or shaders.
- **Banned:**
  - purple, indigo or violet gradients
  - neon cyan on dark
  - gradient text
  - radial glow halos and soft spotlights behind content
  - glowing borders
  - glassmorphism as decoration (one functional backdrop blur for the panel over the relief is allowed if it solves legibility)
  - cream or beige "un-slop" defaults
- **Contrast:** WCAG AA for all text, with 4.5:1 for body and 3:1 for large text, checked by computation.

## 6. Typography (Thai-first)
- **UI face:** a loopless Thai sans with real character and a harmonised Latin, all OFL. Candidates to test in a specimen are IBM Plex Sans Thai, Anuphan and Noto Sans Thai.
- **Sentence face:** the human sentence may use an upright Thai serif such as Noto Serif Thai, Taviraj or Trirong if the specimen proves it. It must never be italic, since an italic serif display is a known slop tell.
- **Do not use** Inter, Geist, Roboto, Arial or the system UI font as the voice.
- **Line height:** Thai needs more than Latin because of stacked vowels and tone marks. Start body at about 1.6 and verify with real strings such as "น้ำล้นตลิ่ง", "ผู้ใหญ่บ้าน", "ปิ่นเกล้า" and "จังหวัดพระนครศรีอยุธยา".
- **Numbers:** measurements use `font-variant-numeric: tabular-nums lining-nums`. Monospace is allowed only for raw measurement readouts if it earns its place, never as a "techy" costume.
- **Font files:** self-host WOFF2 subset to Thai and Latin, preload only the critical faces and use `font-display: swap`.
- **Hierarchy:** a clear scale with obvious steps. Use the impeccable typeset guidance and a display size of 6rem at most.

## 7. Motion
- **One authored moment:** the ใกล้ฉัน flight that lands on *your* river while the sentence settles in. Everything else is quiet.
- **Motion encodes data.** Water levels animate when you scrub the timeline.
- **Banned:**
  - fade-up on every element
  - bounce or elastic easing
  - pulsing dots (unless a gauge genuinely changed in the latest ingest, and then once)
  - marquees
  - hover scale or rotate on everything
- **Timing:** exponential ease-out, 160–260 ms for UI and 1.2–2.0 s for camera moves.

## 8. Copy (Thai), a calm knowledgeable friend
- **Thai is the voice, English is the instrument (Phase 4 rule).** Thai carries the sentence, the situation labels and anything the user reads as speech; English carries the measurement labels, units, timestamps, view labels and credits. A caveat the reader must understand (the ม.รทก. datum note) is voice and stays Thai.
- Be plain and factual. Always state the measured time and the source.
- **Never** say ปลอดภัย, จะท่วม, ไม่ท่วม or anything that predicts or reassures. The humanize package enforces this with a banned-phrase test.
- No marketing words: ปฏิวัติ, ยกระดับ, ล้ำสมัย, ครบจบในที่เดียว.
- No em-dash habit and no "ไม่ใช่แค่ X แต่คือ Y" slogans.
- **Errors** name the problem and the way out, for example *"เปิดตำแหน่งไม่ได้ ลองพิมพ์ชื่ออำเภอแทนได้เลย"*.
- **About page** is short and honest: what the data is, who publishes it, how often it updates, what we deliberately don't do, and the source and licence credits.

## 9. Chrome, states and browser surfaces
- **Chrome is minimal:** wordmark, search, ใกล้ฉัน, timeline, an expandable legend and about. No navbar of links and no footer.
- **States are designed, not left to defaults:**
  - Loading: the relief appears first, followed by the last known data. Use skeletons, never spinners on the main path.
  - Geolocation denied
  - Stale data
  - API unreachable: show the last snapshot with its timestamp.
  - No WebGL or low-power devices: a 2D fallback in the same visual language.
  - LINE in-app browser: add `?openExternalBrowser=1` handling.
- **Browser surfaces are themed:** text selection, focus rings, caret, scrollbars, underline offset and tabular numerals. Models skip this most reliably, and it is the cheapest signal that a page was *built*.
- **Icons:** one consistent drawn set or authored SVG. No emoji or Unicode glyphs as icons.

## 10. The anti-slop gate (a hard checklist, run with screenshot evidence)
Sources: the [impeccable slop catalog](https://impeccable.style/slop), [prg.sh, "Why your AI keeps building the same purple gradient website"](https://prg.sh/ramblings/Why-Your-AI-Keeps-Building-the-Same-Purple-Gradient-Website), [925studios slop tells](https://www.925studios.co/blog/ai-slop-design-tells), [capitalandcompute: fix slop tell by tell](https://capitalandcompute.net/blog/fix-ai-slop-design/), and the impeccable plugin's craft floor (installed locally).

**Why slop happens:** models average the most common patterns when the brief is vague. The cure is a committed world, named references, explicit bans, tokens locked in one place, and reviewing the render rather than the code. Fixing tells cosmetically only produces a differently flavoured template, so fix the root cause.

**Page and layout tells (all banned):**
- hero section with pill badge or eyebrow label above the headline
- identical icon-heading-text card grids
- hero-metric template (big number, small label, row of stats)
- section numbers 01/02/03
- nested cards
- monotonous spacing
- everything centred
- decorative grid backgrounds
- modals for tasks that need no interruption

**Surface tells (all banned):**
- gradient text
- glass and blur as decoration
- hairline border plus wide shadow
- thick coloured side-stripes on cards
- extreme border radius
- glow halos and spotlights
- neon accents on dark
- zero-blur offset "brutalist" shadows
- sparklines or progress rings standing in for content

**Type tells (all banned):**
- Inter, Geist or Roboto as the voice
- italic serif display
- crushed tracking
- flat hierarchy
- one family doing everything with no size or weight steps

**Copy tells (all banned):**
- generic marketing claims
- em-dash habit
- "Not X. Y." slogans
- invented numbers or testimonials

**3D tells (authored for this project, all banned):**
- spinning or auto-rotating globe or hero object
- starfield, particles or floating dust
- bloom and glow on everything
- synthwave neon wireframe terrain or grid floor
- arcs flying between points (the GitHub-globe clone)
- hex-bin globe
- hundreds of identical glowing pillars with no hierarchy
- free-spin OrbitControls that lose the horizon
- default grey `MeshStandardMaterial` under a stock HDRI
- stretched low-resolution textures
- jaggies or z-fighting
- 3D that blocks first paint or janks on a phone
- camera wobble or parallax unrelated to data
- broken Thai shaping in 3D text

**Process at each UI gate:**
1. Take Playwright screenshots at 390×844, 768×1024 and 1440×900, plus the reduced-motion and no-WebGL variants, into `docs/design/screens/<phase>/`.
2. Run `/impeccable critique` against the screenshots and this brief.
3. Score every checklist item pass or fail in `docs/design/GATE-<phase>.md`.
4. Fix root causes, then re-shoot.
5. The gate passes only with zero banned items and every requirement findable within seconds.
