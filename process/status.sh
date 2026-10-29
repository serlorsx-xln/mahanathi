#!/usr/bin/env bash
# Prints where the build stands: running phase, last tool calls, recent commits.
set -euo pipefail
cd "$(dirname "$0")/.."

latest="$(ls -t process/logs/*.jsonl 2>/dev/null | head -1 || true)"
if [[ -n "$latest" ]]; then
  echo "log: $latest ($(wc -l < "$latest" | tr -d ' ') events, updated $(date -r "$latest" '+%H:%M:%S'))"
  python3 -I -c '
import json, sys
events = []
for line in open(sys.argv[1], encoding="utf-8"):
    try:
        events.append(json.loads(line))
    except ValueError:
        pass
done = next((e for e in reversed(events) if e.get("type") == "result"), None)
if done:
    print("status: finished", "(error)" if done.get("is_error") else "", f"turns={done.get(\"num_turns\")}", f"cost=${done.get(\"total_cost_usd\", 0):.2f}")
else:
    print("status: running")
steps = []
for e in events:
    if e.get("type") != "assistant":
        continue
    for c in e.get("message", {}).get("content", []):
        if c.get("type") == "tool_use":
            inp = c.get("input", {})
            what = inp.get("description") or inp.get("file_path") or inp.get("command") or inp.get("pattern") or ""
            steps.append(f"{c.get(\"name\")}: {str(what)[:110]}")
print("recent steps:")
for s in steps[-12:]:
    print("  -", s)
' "$latest"
fi
echo "commits:"
git log --oneline -15 2>/dev/null | sed 's/^/  /'
