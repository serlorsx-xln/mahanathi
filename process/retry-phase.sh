#!/usr/bin/env bash
# Retries a phase on the SAME implementer session when the owner's router
# kills it with a transient "network error". The full phase prompt is sent
# once; every retry resumes with a short continuation prompt so the
# session's context does not balloon. Usage: process/retry-phase.sh <phase>
# MAXTRIES and BASE_WAIT are env-overridable. Never starts a new session.
set -euo pipefail

cd "$(dirname "$0")/.."
phase="${1:?usage: retry-phase.sh <phase-number>}"
maxtries="${MAXTRIES:-12}"
base_wait="${BASE_WAIT:-60}"
session_file="process/.session-id"
mkdir -p process/logs
log="process/logs/phase-${phase}-retry-$(date +%Y%m%dT%H%M%S).jsonl"

attempt_log() { # $1 = prompt file
  printf '%s' "$(cat "$1")" | claude -p \
    --model glm-5.3 \
    --permission-mode bypassPermissions \
    --output-format stream-json --verbose --name mahanathi \
    $( [[ -s "$session_file" ]] && echo "--resume $(cat "$session_file")" ) \
    >> "$log"
}

last_state() {
  python3 -I -c '
import json, sys
state = "no-result"
for line in open(sys.argv[1], encoding="utf-8"):
    try: e = json.loads(line)
    except ValueError: continue
    if e.get("type") == "result":
        r = str(e.get("result", ""))
        state = "api-error" if (e.get("is_error") and ("网络错误" in r or "network error" in r.lower())) else ("error" if e.get("is_error") else "success")
print(state)
' "$log"
}

cont_prompt="process/.retry-prompt-${phase}.txt"
cat > "$cont_prompt" <<'EOF'
Your previous run was cut off by a transient error in the owner's model router. Nothing is lost; your session and the git history are intact. Continue the current phase exactly where you left off:
1. `git status`, `git log --oneline -8`, then re-read `docs/PROGRESS.md` and the tail of `docs/DEVLOG.md`.
2. Re-read the last phase report if the phase is mid-verification.
3. Continue the remaining work. Do not repeat completed steps, do not rewrite committed work, do not re-read the whole phase prompt.
EOF

echo "retry loop for phase ${phase} started $(date '+%H:%M:%S') → ${log}"
process/run-phase.sh "$phase" >> "$log" 2>&1 || true
state="$(last_state)"
echo "attempt 1: ${state}"

i=1
while [[ "$state" == "api-error" && "$i" -lt "$maxtries" ]]; do
  i=$((i+1))
  wait=$(( base_wait * (i-1) ))
  echo "attempt ${i} in ${wait}s ($(date '+%H:%M:%S'))"
  sleep "$wait"
  attempt_log "$cont_prompt" || true
  state="$(last_state)"
  echo "attempt ${i}: ${state}"
done

# keep the session pointer in sync with any successful continuation
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

echo "retry loop finished $(date '+%H:%M:%S'): ${state} after ${i} attempt(s)"
rm -f "$cont_prompt"
[[ "$state" == "success" ]]
