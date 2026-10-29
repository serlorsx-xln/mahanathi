# How this project is built

This project is built in phases by an AI coding agent. The owner (ชานน) sets the direction. A lead AI session writes the briefs, reviews every phase and handles outward actions such as GitHub and deploys. An implementing Claude Code session (Opus 5.5, effort xhigh) writes the code. Every phase resumes that same session.

| Path | What it is |
|---|---|
| `docs/brief/` | The lead's briefs: product, design direction, engineering, workflow |
| `process/prompts/phase-N.md` | The exact prompt given to the implementer for each phase |
| `process/lead-notes/phase-N.md` | The lead's review of the previous phase, appended to the next prompt |
| `process/reports/phase-N.md` | The implementer's report at the end of each phase |
| `process/run-phase.sh` | Runs one phase (`process/run-phase.sh 2`) |
| `process/status.sh` | Shows the running phase, its recent steps and recent commits |

Each loop goes: prompt → plan → small commits with tests → verification (with screenshots for UI) → report → lead review → next prompt.
