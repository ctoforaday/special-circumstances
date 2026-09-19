#!/usr/bin/env python3
"""Render a `claude -p --output-format stream-json` stream as readable progress.

WHY: `--output-format json` returns ONE object when the process exits. A research run is tens of
minutes of subagent dispatch, and for all of it the operator sees nothing — so a run that is
working and a run that is wedged look identical until it ends. That is the same
indistinguishable-zero this engine keeps being bitten by, applied to the operator.

`stream-json` emits an event per message as it happens. This turns that into lines worth reading:
the seats dispatched, the verbs they run, and anything that failed. Everything else is dropped,
because a firehose nobody reads is the same as no output.

Reads the stream on stdin, echoes the raw JSONL to `--raw <path>` so nothing is lost, and prints
the filtered view to stdout.
"""
import argparse
import json
import sys
import time

# The verbs worth seeing. A seat runs many more calls than these; these are the ones that MOVE the
# debate, so the line count stays proportional to progress rather than to effort.
INTERESTING = ("register", "mint", "close", "verdict", "outcome", "halt", "certify",
               "position", "closing", "dispatch", "ingest", "edit", "prove", "cite",
               "log", "regrade", "motion", "carry", "class")


def main() -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("--raw", help="also write every raw stream line here")
    args = ap.parse_args()

    raw = open(args.raw, "a", encoding="utf-8") if args.raw else None
    t0 = time.time()
    seats: set[str] = set()

    def say(msg: str) -> None:
        print(f"[{int(time.time()-t0):>5}s] {msg}", flush=True)

    for line in sys.stdin:
        if raw:
            raw.write(line)
            raw.flush()
        line = line.strip()
        if not line:
            continue
        try:
            ev = json.loads(line)
        except json.JSONDecodeError:
            continue

        kind = ev.get("type")

        # A SUBAGENT STARTING IS THE EVENT THAT MATTERS MOST — it is the one thing the previous
        # harness could never show, because it dispatched top-level processes and had no subagents.
        if kind == "system" and ev.get("subtype"):
            sub = ev["subtype"]
            if sub in ("init", "compact_boundary"):
                say(f"session {sub}")
            continue

        if kind == "assistant":
            for block in (ev.get("message", {}).get("content") or []):
                if block.get("type") != "tool_use":
                    continue
                name = block.get("name", "?")
                if name == "Task":
                    inp = block.get("input") or {}
                    label = inp.get("description") or inp.get("subagent_type") or "?"
                    seats.add(label)
                    say(f"DISPATCH  {label}")
                elif name == "Bash":
                    cmd = ((block.get("input") or {}).get("command") or "")
                    for verb in INTERESTING:
                        if f" {verb}" in cmd and "feov-record" in cmd:
                            say(f"  verb    {verb}")
                            break
            continue

        if kind == "user":
            for block in (ev.get("message", {}).get("content") or []):
                if block.get("type") == "tool_result" and block.get("is_error"):
                    say("  ERROR   " + str(block.get("content"))[:160].replace("\n", " "))
            continue

        if kind == "result":
            cost = ev.get("total_cost_usd") or 0
            say(f"DONE  {ev.get('subtype','?')}  turns={ev.get('num_turns')}  ${cost:.2f}")
            if ev.get("is_error"):
                say("  RESULT IS AN ERROR: " + str(ev.get("result"))[:300])
            break

    if raw:
        raw.close()
    say(f"summary: {len(seats)} dispatch(es) seen")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
