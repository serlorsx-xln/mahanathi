# phase-4-intro: state at the disk-full stop (2026-10-30 ~15:10)

## What happened

The internal disk (`/`, which holds `/tmp/claude-501`, Claude Code's
temp directory) filled completely (ENOSPC) mid-verification. The Bash
tool can no longer run ANY command — the harness itself cannot capture
output. Write/Edit still work (they target the UNITEK volume). This
note is the handoff.

## State of the one unit (the branded entry moment)

**Code: complete, statically green, UNVERIFIED in-browser, UNCOMMITTED.**

- `web/src/routes/Lookdev.tsx`: the intro state machine — `revealedAt`
  set once from the map's `onReady` (the external event), `introDone`
  set by one timeout (subscription form; passes the
  `react-hooks/set-state-in-effect` rule), `introPhase` derived in
  render (`wordmark` → `revealing` → `done`; under reduced motion the
  reveal is skipped — the cut). The overlay div is the last child of
  `<main>`, `aria-hidden`, pointer-transparent.
- `web/src/routes/lookdev.css`: `.lookdev__intro` (ground token,
  z-overlay, pointer-events none), the wordmark reveal (600 ms,
  motion-ease-out, opacity + 6 px rise), the fade-through-fog
  (ground → relief-fog → gone, 600 ms, matches `INTRO_FADE_MS`), the
  reduced-motion cut.
- Verified BEFORE the disk filled: `tsc --noEmit` clean,
  `eslint --max-warnings=0` clean, unit **35/35**, `pnpm build` green.
- NOT done (needs Bash): the in-browser verification script
  (`e2e/.intro.tmp.mjs` — written, then the run was blocked; the file
  may or may still exist), the gate screenshot re-shoot, the full e2e,
  the commit, the report addendum.

## What the lead needs to do

1. Free space on `/` (the biggest reclaimable items are likely
   `/tmp/claude-501/playwright-*` artifacts and whatever filled the
   disk), or restart Claude Code with `CLAUDE_CODE_TMPDIR` pointed at
   the UNITEK volume.
2. A vite preview server from this session may still be listening on
   `127.0.0.1:4173` — kill it if it interferes (`lsof -ti :4173`).
3. Then the session (or the next continuation) runs: the in-browser
   intro check, `playwright test tests/gate-phase3b.spec.ts
   --project=desktop`, and commits the unit (code + re-shot
   screenshots + the report addendum) — the commit message should
   credit the phase-4-intro prompt.

## Honesty note

No in-browser claim about the intro has been made anywhere. The
verification numbers above (tsc/eslint/unit/build) are from the runs
that completed before the fill. Anything further is unverified until
run.
