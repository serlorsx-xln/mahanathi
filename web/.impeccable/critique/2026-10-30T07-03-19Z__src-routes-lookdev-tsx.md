---
target: src/routes/Lookdev.tsx
total_score: 31
max_score: 40
na_heuristics: 
p0_count: 0
p1_count: 2
target_identity: "file:/Volumes/UNITEK/marketting/analysis/mahanathi/web/src/routes/Lookdev.tsx"
target_fingerprint: "sha256:88bd988f485bbbe3a65eb062bac5ef3d1696bfbea8e6d8f888c66d08954b8983"
target_path: /Volumes/UNITEK/marketting/analysis/mahanathi/web/src/routes/Lookdev.tsx
timestamp: 2026-10-30T07-03-19Z
slug: src-routes-lookdev-tsx
---
---
run: 2026-10-30T13:45:00+07:00
target: src/routes/Lookdev.tsx
note: >-
  The Phase 3b remediation's second critique iteration (the mandatory
  two), run as two isolated sub-agents against the post-remediation
  build (commits ad6e49d..3c826af plus the copy fixes), synthesized in
  the parent. Iteration 1 scored 27/40 - the 28 previously reported
  was an arithmetic slip, corrected in gate/DEVLOG/report this session.
total_score: 31
max_score: 40
p0_count: 0
p1_count: 2
---

# `/impeccable critique` — Phase 3b remediation, iteration 2

**Method:** dual-agent (A: design review; B: detector + browser evidence), isolated sub-agents, synthesized after both finished. Detector clean (`[]`, exit 0) on all seven changed files. Zero console/page errors in all seven scenarios; zero horizontal overflow; all contrast AA with margin. Neither session can view images; every claim is a computed measurement.

## Heuristic scores (0-4)

1: 3 — sentence lands 71-104 ms after navigation; skeleton + SR status; no signal for basemap/DEM refinement after scene-ready; flight status never shows on phone.
2: 3 — plain Thai, real geography; ใกล้ฉัน flies to a fixed sample station; sheet answers in ม.รทก.
3: 3.5 — presets jump back, sheet closes (verified), flight interruptible, retry works, aria-pressed on presets.
4: 3.5 — stale hollow in all three surfaces; LINE hint and no-WebGL note both overlap the presets row 11.3 px on phones (dev chrome collider).
5: 3.5 — stale cannot read as a level by construction; overflow clamped; misses named.
6: 3 — pole anatomy (cap = bank-full) never explained; ม.รทก. assumes training.
7: 2.5 — no search (J3), no keyboard path, no deep-linking yet (Phase 4 scope).
8: 3.5 — desktop chrome 28.9%; phone national 58.1% with dev chrome wrapping to two rows.
9: 3.5 — fetch, WebGL, map-init and shader-compile failures all degrade to designed floors.
10: 2 — no about, no encoding legend (Phase 4 with the about page).

**Total: 31/40** (iteration 1: 27/40, previously misreported as 28 - corrected).

## Design specificity

Specific at rest, and now specific in motion. At the station zoom, ±4/channel sampling found 0 pixels of Protomaps' own greys and 92.1% ground-family. Level-5 at landing 6.7% of desktop canvas (was 24.45%). Remaining generic: close-zoom structure still borrows a dark vector basemap's rhythm; the primary action's label promises what the system doesn't do yet.

## Priority issues

- [P1] ใกล้ฉัน means "a station we chose": flies to fixture station 736 regardless of location, disclosed only by a 12 px faint note; the landing sheet names no river or province. Fix in Phase 4 (geolocation is J2): disclosure in the sheet head + river/province fields.
- [P1] Thailand's tips are cropped in the default desktop frame: at 1440x900 the northernmost land projects to y=-7 and Betong to y=923, outside the viewport despite fitBounds padding 72 (globe projection interacts with the fit). Phone clean. Fix: reconcile the fit with the globe + an e2e assertion on the four extreme land points.
- [P2] no-WebGL copy overclaimed ("data still reads as normal" while only the national summary is reachable). FIXED this session: เห็นเฉพาะสรุปภาพรวม ไม่สามารถดูรายสถานีได้.
- [P2] source credit misattributed the publisher (said กรมชลประทาน/RID; the platform is HII's, stations from four agencies). FIXED this session: ThaiWater · สถาบันสารสนเทศทรัพยากรน้ำ (HII), matching the server's committed footer.
- [P2] desktop first relief frame 6.5 s in the headless/software-GL run (phone 688 ms - the trustworthy number). Fix: profile tile loading at the desktop fit zoom.

## Persona red flags

- Casey (night, one-handed): sample-station ใกล้ฉัน; overflow at national view 38 red px on phone (the sentence does all the work); LINE escape pill's bottom 11.3 px under the presets row.
- Sam (a11y): canvas is the first tab stop with an English accessible name ("Map"); no keyboard path to any station (Phase 4 with search).
- Jordan (judge from QR): dev chrome wraps to two rows on phone (13.7% of height); the 20-second path works end-to-end (sentence 71 ms -> tap -> 2 s flight -> serif sentence).

## Minor observations

Panel counts total 792 of 804 (12 no-level stations in no count); station sheet has measured time but no source line; discs register far better than poles at every national measurement (level-3: 3062 vs 742 px) - the poles-vs-discs question is genuinely open.

## Strengths

1. The peak moment survives measurement: after ใกล้ฉัน on phone the sheet takes 215 px of 844, the gauge projects on-screen clear of it, ~72% relief visible.
2. One encoding across three surfaces (pole shader, disc expressions, DOM legend).
3. The degradation floors are real and verified (fetch/WebGL/map-init/shader-compile).
