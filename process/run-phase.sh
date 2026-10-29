#!/usr/bin/env bash
# Runs one phase of the build with the implementing Claude session.
# Usage: process/run-phase.sh <phase-number>
#        FRESH=1 process/run-phase.sh <phase-number>   (start a new session)
# The prompt is process/prompts/phase-N.md plus the lead's review notes from
# process/lead-notes/phase-N.md, if that file exists. Every phase after the
# first resumes the same session, so the implementer keeps its context.
# FRESH=1 starts a new session instead; the next phase resumes that one.
set -euo pipefail

cd "$(dirname "$0")/.."
phase="${1:?usage: run-phase.sh <phase-number>}"
prompt_file="process/prompts/phase-${phase}.md"
notes_file="process/lead-notes/phase-${phase}.md"
session_file="process/.session-id"
mkdir -p process/logs
log="process/logs/phase-${phase}-$(date +%Y%m%dT%H%M%S).jsonl"

[[ -f "$prompt_file" ]] || { echo "missing $prompt_file" >&2; exit 1; }

prompt="$(cat "$prompt_file")"
if [[ -f "$notes_file" ]]; then
  prompt+=$'\n\n---\n\n## Lead notes from the previous review\n\n'"$(cat "$notes_file")"
fi

args=(-p
  --model glm-5.3
  --permission-mode bypassPermissions
  --output-format stream-json
  --verbose
  --name mahanathi)
# Model note (lead, 2026-10-29 21:10): Phases 1 and half of 2 ran on
# claude-opus-5-5 --effort xhigh. The owner's Anthropic subscription is
# blocked for past-due payment and the owner's router serves no Claude
# models, so on the owner's decision the build continues on glm-5.3.
# Do not pass --effort here: the router rejects effort-mapped thinking
# settings for this model ("该模型始终思考", it always thinks anyway).
if [[ -s "$session_file" && "${FRESH:-0}" != 1 ]]; then
  args+=(--resume "$(cat "$session_file")")
fi

echo "phase ${phase} started $(date '+%Y-%m-%d %H:%M:%S') → ${log}"
printf '%s' "$prompt" | claude "${args[@]}" > "$log"

# Remember the session so the next phase resumes it.
python3 -I -c '
import json, sys
sid = None
for line in open(sys.argv[1], encoding="utf-8"):
    try:
        ev = json.loads(line)
    except ValueError:
        continue
    sid = ev.get("session_id") or sid
if sid:
    open(sys.argv[2], "w").write(sid)
' "$log" "$session_file"

echo "phase ${phase} finished $(date '+%Y-%m-%d %H:%M:%S')"
