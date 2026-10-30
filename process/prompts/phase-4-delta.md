# Phase 4 delta: owner's mid-phase directions (same session, continue)

The owner added three directions while you were running. The first P0s (view modes, loopless type, copy rule) are committed and stay. Fold these in, in this order, with the hard stop still in force — whatever does not fit is cut and named in the report.

## 1. Rename the view labels — DEFAULT / TERRAIN 3D
The owner clarified the Google-Maps intent: the pair is a plain default map vs the realistic view. Ours is terrain relief, not satellite imagery — so the honest, still-familiar labels are `DEFAULT` / `TERRAIN 3D` (Google Maps itself has a Terrain layer). Rename the strings, keep everything else about the toggle as shipped. If copy already wrote `2D MAP` anywhere (docs, e2e, DEMO.md), update it in the same commit.

## 2. Loading screen + intro animation (the owner's ask)
A branded entry moment: quiet wordmark reveal ("มหานที") over the ground token, then a short fade-through-fog into the globe as tiles land — ~1.2 s total, cut instantly under reduced motion, and it must never delay interactivity (the app is interactive the moment data lands; the intro is a choreographed reveal of the already-ready frame, not a blocker). Reuse the existing skeleton surface and motion tokens. No autoplaying hero video, no heavy asset.

## 3. Tailwind v4 + shadcn/ui for NEW components — foundation now, migration later
The owner wants shadcn + Tailwind for components, and template-first everywhere. Today: install Tailwind v4 (Vite plugin form) and shadcn/ui init WITHOUT disturbing the existing token CSS (the tokens.css mirror and its drift test must keep passing — scope shadcn's CSS variables in their own layer, mapped from tokens.ts where a component needs a colour, so the single-source-of-truth rule holds). Then build the new UI pieces on shadcn components: the view-mode toggle and share button (the share deep-link if not yet committed) become shadcn `Toggle`/`Button` + `Tooltip` on top of our tokens. Do NOT re-skin the existing panel/sheet — those passed the anti-slop gate and the migration is a documented post-interview phase. Record the two-stage decision in the DEVLOG honestly: foundation + new components today; full migration later, deliberately, with the gates re-run.

## 4. Template-first as standing policy
Add it to the design brief (and follow it): before hand-building any UI piece, search for an existing component/template that fits (shadcn/ui first, then broader). Hand-build only what provably does not exist. A branded intro is one of the few honest exceptions — no template carries our world; the choreography stays ours.

Then: full suite, real numbers, report, stop for the lead. Cut anything that does not fit the clock — fewer, better.
