#!/usr/bin/env python3
"""Read what `universe.sh resume` needs out of a run's stream, or refuse.

The first Workflow call in the stream holds the script and the args, the event carrying it names the
lead session, and that call's result holds the run id. Each is a field of the stream, so none of them
is retyped by an operator, and a resume passes the SAME args because they are the ones the run used.

Prints the lead session on the first line and the lead's prompt after it. Refuses with exit 1 and a
reason on stderr when the stream has no complete Workflow call, or when the run's record has an
outcome or cannot say whether it has one.
"""
import json
import os
import sqlite3
import sys


def refuse(why: str) -> "None":
    print(f"[universe] REFUSING to resume: {why}", file=sys.stderr)
    raise SystemExit(1)


def blocks(ev: dict) -> list:
    content = (ev.get("message") or {}).get("content")
    if not isinstance(content, list):
        return []
    return [b for b in content if isinstance(b, dict)]


def first_workflow(stream: str) -> tuple[str, dict, str]:
    """The lead session, the input and the run id of the stream's first Workflow call."""
    session, call_id, call = "", "", None
    with open(stream, encoding="utf-8") as f:
        for line in f:
            try:
                ev = json.loads(line)
            except json.JSONDecodeError:
                continue
            if not isinstance(ev, dict):
                continue
            if call is None and ev.get("type") == "assistant":
                for b in blocks(ev):
                    if b.get("type") == "tool_use" and b.get("name") == "Workflow":
                        session, call_id, call = ev.get("session_id") or "", b.get("id"), b.get("input") or {}
                        break
            elif call is not None and ev.get("type") == "user":
                for b in blocks(ev):
                    if b.get("tool_use_id") == call_id:
                        result = ev.get("tool_use_result")
                        run_id = result.get("runId") if isinstance(result, dict) else ""
                        return session, call, run_id or ""
    if call is None:
        refuse(f"{stream} holds no Workflow call, so there is no debate to resume")
    refuse(f"{stream} holds a Workflow call with no result, so it names no run id")
    raise AssertionError


def require_no_outcome(run_dir: str) -> None:
    db = os.path.join(run_dir, "records", "record.db")
    if not os.path.isfile(db):
        refuse(f"no record at {db}, so nothing says the run has no outcome")
    try:
        c = sqlite3.connect("file:" + db + "?mode=ro", uri=True)
        rows = list(c.execute('SELECT "verdict" FROM "outcome"'))
    except sqlite3.Error as e:
        refuse(f"could not read the outcome from {db}: {e}")
    if rows:
        refuse(f"the run already has an outcome ({rows[0][0]}), in {db}")


def main() -> int:
    session, call, run_id = first_workflow(sys.argv[1])
    args = call.get("args")
    # The tool accepts the args as an object or as its JSON text.
    if isinstance(args, str):
        try:
            args = json.loads(args)
        except json.JSONDecodeError as e:
            refuse(f"the Workflow call's args are not JSON: {e}")
    script = call.get("scriptPath")
    if not isinstance(args, dict) or not args.get("runDir"):
        refuse("the Workflow call's args name no runDir")
    if not script:
        refuse("the Workflow call names no scriptPath")
    if not session:
        refuse("the Workflow call's event names no session")
    if not run_id:
        refuse("the Workflow call's result names no run id")
    require_no_outcome(args["runDir"])

    print(session)
    print(
        "Operator instruction. This session's process died before the debate reached an outcome, so "
        "the debate is resumable, and the operator has directed resuming it. Resume it now: invoke the "
        f'Workflow tool with resumeFromRunId "{run_id}", scriptPath "{script}", and the SAME args as '
        f"the original call: {json.dumps(args, ensure_ascii=False)}. If this is a different session "
        "from the run's lead session, first run the research skill's take-over step from "
        f"--from-session {session}. Then carry on with the research skill's remaining steps as "
        "written, through capture once the assembly finishes."
    )
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
