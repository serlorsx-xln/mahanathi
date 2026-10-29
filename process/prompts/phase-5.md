# Phase 5: Polish and hardening

Lead notes from the Phase 4 review are at the bottom; address them first.

**Goal:** get from good to undeniable. In the judges' hands, nothing should feel unfinished.

## Tasks
1. **Audit and polish:** run `/impeccable audit` and then `/impeccable polish` on the live product. Fix every P0 and P1 item, and list any deliberately deferred P2 items.
2. **Accessibility:**
   - The keyboard path covers the whole product.
   - Focus is visible and themed.
   - Screen-reader labels on the canvas region give a text alternative listing the national sentence and the selected station.
   - The panel uses proper landmarks.
   - Verify contrast by computation.
3. **Mobile:**
   - Test on 390×844 and 360×800 viewports.
   - Handle the iOS Safari safe areas and the bottom sheet gestures.
   - Make sure there is no scroll-jacking and no horizontal overflow.
   - Touch targets are 44px or larger.
4. **Browser surfaces:** text selection, caret, scrollbars, focus rings, underline offset and numerals are all themed.
5. **Copy pass (Thai):**
   - Read every string aloud and remove anything robotic or marketing.
   - Errors name the problem and the way out.
6. **Meta:**
   - Title and description in Thai.
   - An OG image made from a real render, not a template card.
   - Favicon, `theme-color`, `manifest` and a designed 404.
7. **Hardening:**
   - Verify the CSP and security headers with curl.
   - Test the rate limit.
   - Graceful behaviour when ThaiWater is down: serve the last good data with an honest timestamp.
   - Check log quality.
8. **Performance:** meet every budget. Show the before and after numbers.

## Gate
- Lighthouse mobile: Accessibility 95 or more, Best Practices 95 or more, Performance 85 or more.
- `docs/design/GATE-phase-5.md` is all pass, with fresh screenshots in `docs/design/screens/phase-5/`.

## Report and stop
Write `process/reports/phase-5.md`. Stop.
