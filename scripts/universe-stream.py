#!/usr/bin/env python3
"""Render a `claude -p --output-format stream-json` stream as readable progress.

WHY: `--output-format json` returns ONE object when the process exits. A research run is tens of
minutes of subagent dispatch, and for all of it the operator sees nothing — so a run that is
working and a run that is wedged look identical until it ends. That is the same
indistinguishable-zero this engine keeps being bitten by, applied to the operator.

`stream-json` emits an event per message as it happens. This turns that into lines worth reading:
the seats dispatched, the verbs they run, and anything that failed. Everything else is dropped,
because a firehose nobody reads is the same as no output.

Reads the stream on stdin to its end and prints the filtered view to stdout. The raw stream is not
this script's to keep: claude writes it to a file and this reads a tail of that file, so nothing here
can stop the capture.

A `result` event is the end of a TURN, not of the stream. A resumed session's queued notification
gets one with num_turns 0 before the prompt is read, and only end of input says the process is done.
"""
import argparse
import json
import sys
import time

# The verbs worth seeing. A seat runs many more calls than these; these are the ones that MOVE the
# debate, so the line count stays proportional to progress rather than to effort.
INTERESTING = ("register", "mint", "close", "verdict", "outcome", "halt", "certify",
               "position", "closing", "dispatch", "ingest", "edit", "compute", "cite",
               "log", "regrade", "motion", "carry")


def content_blocks(ev: dict) -> list[dict]:
    """A message's blocks. A user message's content can be plain text, which holds none."""
    content = (ev.get("message") or {}).get("content")
    if not isinstance(content, list):
        return []
    return [b for b in content if isinstance(b, dict)]


def main() -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("--resumed", action="store_true",
                    help="the stream is a resumed lead session, not a fresh one")
    args = ap.parse_args()

    t0 = time.time()
    seats: set[str] = set()

    def say(msg: str) -> None:
        print(f"[{int(time.time()-t0):>5}s] {msg}", flush=True)

    for line in sys.stdin:
        line = line.strip()
        if not line:
            continue
        try:
            ev = json.loads(line)
        except json.JSONDecodeError:
            continue
        if not isinstance(ev, dict):
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
            for block in content_blocks(ev):
                if block.get("type") != "tool_use":
                    continue
                name = block.get("name", "?")
                # THE ENGINE DISPATCHES THROUGH `Workflow`, NOT `Task`, and the first version of
                # this renderer only knew about Task — so the single most important event in a run
                # scrolled past unprinted and the view sat silent for minutes while the debate ran.
                # Measured on the first universe smoke: 24 Bash calls and exactly one Workflow.
                if name == "Workflow":
                    inp = block.get("input") or {}
                    seats.add("workflow")
                    say("WORKFLOW DISPATCHED — the debate runs from here")
                    if inp.get("name"):
                        say(f"  script  {inp['name']}")
                elif name == "Task":
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
            for block in content_blocks(ev):
                if block.get("type") == "tool_result" and block.get("is_error"):
                    say("  ERROR   " + str(block.get("content"))[:160].replace("\n", " "))
            continue

        if kind == "result":
            cost = ev.get("total_cost_usd") or 0
            say(f"RESULT  {ev.get('subtype','?')}  turns={ev.get('num_turns')}  ${cost:.2f}")
            if ev.get("is_error"):
                say("  RESULT IS AN ERROR: " + str(ev.get("result"))[:300])

    say(f"summary: {len(seats)} dispatch(es) seen")

    # A RUN THAT DISPATCHED NOTHING IS A FAILED RUN, and it is the one failure this whole script
    # exists to make visible, so it must not be reported as a zero to be read past. `claude -p`
    # exits 0 whenever the assistant produced an answer — including when the plugins did not load
    # and `/frank-exchange-of-views:research <topic>` reached the model as ordinary prose. That run
    # printed `DONE success turns=1 $0.05` above a `0 dispatch(es)` line, and the wrapper agreed.
    # The engine cannot research anything without dispatching the workflow, so the absence of one
    # is decidable here and is stated as an error rather than left in the summary's arithmetic.
    #
    # A resumed lead has run the engine and has loaded its plugins, so the same absence means something
    # else there: it was told to restart the debate and did not.
    if "workflow" not in seats and args.resumed:
        say("FAILED: the resumed lead dispatched no Workflow — the debate did not restart")
        say("  the stream holds what the lead did instead")
        return 2
    if "workflow" not in seats:
        say("FAILED: no Workflow was dispatched — the engine never ran, whatever the exit code says")
        say("  most likely the plugins did not load; check that the marketplace path in the")
        say("  universe's settings.json still exists, then rebuild")
        return 2
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
