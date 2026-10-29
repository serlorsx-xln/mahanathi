---
version: 1
slug: "src-app-tsx"
primary_target: "src/App.tsx"
related_targets: ["src/design/tokens.css"]
---

# Surface brief: look-dev (Phase 3)

Scope: the dev-only routes `/lookdev`, `/specimen` and the colour-scale strip, plus the tokens under `web/src/design/`. The product surface itself is Phase 4; this surface proves the visual world in isolation. Visitor mode: **Operate** — the visitor (the owner, the lead, the judges at the demo) inspects an instrument at work.

## Audience, job, proof

- Audience: the lead (must approve the look before Phase 4) and the owner (must defend it in the interview).
- Job: see the committed world rendered at full fidelity — relief, gauges, type, colour — with real data, and judge whether it feels like a premium instrument.
- Proof: Playwright screenshots at 390×844 and 1440×900, a colour-blindness simulation, computed contrast numbers, and the anti-slop gate scored item by item.

## Direction contract

THESIS: a night-time hydrographic instrument — the national survey's plaster relief of Thailand in a dark map room, lit only where water is being measured right now. The category-default arrangement refused: no dashboard, no hero-plus-cards, no glowing-pillar globe.

OWN-WORLD: night-slate ground (never #000), matte ochre-sediment relief with a quiet hypsometric tint, one disciplined alert colour, an ordered five-step situation scale where only levels 4–5 get strong colour. Slim gauges with a bank-full cap, not pillars. Thai-first typography with a loopless UI sans and, if the specimen proves it, an upright Thai serif for the sentence.

STORY: the visitor opens the page and the room goes quiet — Thailand in relief at night, every gauge showing real water, one sentence about the country now. They trust it because every number carries its measured time.

FIRST VIEWPORT: at 1440×900 the relief fills the frame, composed off-centre (Thailand is tall); the national sentence sits top-left in the answer panel over the relief; the legend and timeline are quiet chrome at the bottom. At 390×844 the relief fills the portrait screen, the sentence at the top, the panel becomes a bottom sheet. The primary action ใกล้ฉัน sits top-right.

FORM: brief-pinned (docs/brief/01-DESIGN-DIRECTION.md); the roll was not run because the world is committed. seed: none (unavailable — direction pinned by brief).

FINISH: unreviewed and undocumented is unfinished; this build ends with the finish review, the verdict, DESIGN.md, and every shipping raster carrying its provenance.

## Unresolved decisions

- The type pairing (specimen decides: IBM Plex Sans Thai vs Anuphan vs Noto Sans Thai; serif only if earned).
- Vertical exaggeration (ADR 0006, by eye with screenshots).
- The exact alert colour and situation-scale stops (contrast + CVD simulation decide).
