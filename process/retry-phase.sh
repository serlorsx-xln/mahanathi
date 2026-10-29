#!/usr/bin/env bash
# Runs a phase on the SAME implementer session, retrying on the owner
# router's transient "network error" cuts. Every attempt is a direct
# claude call appended to ONE stream-json log, so state tracking reads
# the last result event. The full phase prompt goes once; retries resume
# with a short continuation prompt. Never starts a new session.
# Usage: process/retry-phase.sh <phase>   (MAXTRIES, BASE_WAIT env-able)
set -euo pipefail

cd "$(dirname "$0")/.."
phase="${1:?usage: retry-phase.sh <phase-number>}"
maxtries="${MAXTRIES:-20}"
base_wait="${BASE_WAIT:-60}"
session_file="process/.session-id"
mkdir -p process/logs
log="process/logs/phase-${phase}-retry-$(date +%Y%m%dT%H%M%S).jsonl"
model="${IMPLEMENTER_MODEL:-glm-5.3}"

claude_attempt() { # $1 = prompt file
  printf '%s' "$(cat "$1")" | claude -p \
    --model "$model" \
    --permission-mode bypassPermissions \
    --output-format stream-json --verbose --name mahanathi \
    --resume "$(cat "$session_file")" \
    >> "$log"
}

state() {
  python3 -I -c '
import json, sys
state = "no-result"
for line in open(sys.argv[1], encoding="utf-8"):
    try: e = json.loads(line)
    except ValueError: continue
    if e.get("type") == "result":
        r = str(e.get("result", ""))
        low = r.lower()
        state = ("api-cut" if ("网络错误" in r or "network error" in low or "overloaded" in low)
                 else "error" if e.get("is_error") else "success")
print(state)
' "$log"
}

# assemble the full phase prompt exactly as run-phase.sh does
full_prompt="process/.phase-prompt-${phase}.txt"
{
  cat "process/prompts/phase-${phase}.md"
  notes="process/lead-notes/phase-${phase}.md"
  [[ -f "$notes" ]] && { printf '\n\n---\n\n## Lead notes from the previous review\n\n'; cat "$notes"; }
} > "$full_prompt"

cont_prompt="process/.retry-prompt-${phase}.txt"
cat > "$cont_prompt" <<'EOF'
Your previous run was cut off by a transient error in the owner's model router. Nothing is lost; this is the same session, and the git history is intact. Continue the current phase exactly where you left off:
1. `git status`, `git log --oneline -8`, then re-read `docs/PROGRESS.md` and the tail of `docs/DEVLOG.md`.
2. If mid-verification, re-read your own last phase report's unfinished sections.
3. Continue the remaining work. Do not repeat completed steps, do not rewrite committed work, and do not re-read the full phase prompt.
EOF

echo "retry loop for phase ${phase} started $(date '+%H:%M:%S') → ${log}"
echo "attempt 1 (full prompt) $(date '+%H:%M:%S')"
claude_attempt "$full_prompt" || true
st="$(state)"; echo "attempt 1: ${st}"

i=1
while [[ "$st" == "api-cut" && "$i" -lt "$maxtries" ]]; do
  i=$((i+1))
  wait=$(( base_wait * (i-1) )); [[ "$wait" -gt 600 ]] && wait=600
  echo "attempt ${i} (continuation) in ${wait}s $(date '+%H:%M:%S')"
  sleep "$wait"
  echo "attempt ${i} $(date '+%H:%M:%S')"
  claude_attempt "$cont_prompt" || true
  st="$(state)"; echo "attempt ${i}: ${st}"
done

# sync the session pointer (resume keeps the same id; read the last event)
python3 -I -c '
import json, sys
sid = None
for line in open(sys.argv[1], encoding="utf-8"):
    try: e = json.loads(line)
    except ValueError: continue
    sid = e.get("session_id") or sid
if sid:
    open(sys.argv[2], "w").write(sid)
' "$log" "$session_file"

echo "retry loop finished $(date '+%H:%M:%S'): ${st} after ${i} attempt(s)"
rm -f "$cont_prompt" "$full_prompt"
[[ "$st" == "success" ]]
