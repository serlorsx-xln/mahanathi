# Lead notes for phase 5 (README unit)

State sync since your last turn (all lead actions, already committed and
pushed — nothing for you to redo):

- `b908886` — the root entry now renders the look-dev surface itself
  (`web/src/App.tsx`): the judges' QR points at the root, and the root was
  still the Phase-1 wordmark stub. The App test re-pins the entry. Your
  `/lookdev` route stays reachable and is the same surface.
- `352ebd8` — prettier formatting on seven files (map layer, routes, e2e) that
  CI's format step had been rejecting; the web job had been red for a while.
- `8702e4d` — DEVLOG records the find and the fix.
- Production is deployed and browser-verified at the root: canvas, national
  sentence, both view modes, presets. The only 404 is the designed
  `thailand.pmtiles` fallback (terrain-only basemap).

One caution for the README: the interview is 2026-10-31 16:00, so this unit
must land quickly and cleanly — prefer short and true over long and impressive.
If any command or number is uncertain, check the file it lives in rather than
writing from memory, and leave it out if you cannot verify it.
