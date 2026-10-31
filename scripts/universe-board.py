#!/usr/bin/env python3
"""Print the live state of a run's BOARD, read-only, from the record.

The stream view shows what the harness is doing — which seats were dispatched, which verbs ran.
This shows what the DEBATE is doing: who has sat, what is on the board, what closed, where the
epochs got to. Those are different questions and only the second one says whether the run is
going well.

READ-ONLY BY CONSTRUCTION. Opened with SQLite's `mode=ro` URI, so watching a live run can never
take a write lock or disturb a seat mid-sitting. The run directory is the database; nothing here
writes to it.

What the record cannot say is whether a process is still writing it, so the caller passes that in:
`universe.sh watch` reads the pid file. Exit 3 means TRUNCATED: the process is dead and the record has
no outcome.
"""
import argparse
import glob
import os
import sqlite3
import sys


def count(c: sqlite3.Connection, table: str) -> int | None:
    try:
        return list(c.execute(f'SELECT count(*) FROM "{table}"'))[0][0]
    except sqlite3.Error:
        return None


def sitting_balance(c: sqlite3.Connection) -> tuple[list[str], list[str]]:
    """The agents with an open sitting, and the closes that close no open sitting, in event order."""
    try:
        rows = list(c.execute(
            'SELECT "agent_id", 1, "event_id" FROM "sitting_open"'
            ' UNION ALL SELECT "agent_id", 0, "event_id" FROM "sitting_close"'
            ' ORDER BY 3'))
    except sqlite3.Error:
        return [], []
    sitting: dict[str, bool] = {}
    extra: list[str] = []
    for agent, is_open, event in rows:
        if is_open:
            sitting[agent] = True
        elif sitting.get(agent):
            sitting[agent] = False
        else:
            extra.append(f"{agent}@ev{event}")
    return sorted(a for a, s in sitting.items() if s), extra


def process_verdict(state: str, detail: str, resume: str, ended: bool) -> int:
    """What the process and the record say together. Silence on the board is not an input."""
    if state == "running":
        print(f"  process: running ({detail})")
        if not ended:
            print("  no run outcome yet")
        return 0
    if state == "dead" and ended:
        print(f"  process: ended ({detail})")
        return 0
    if state == "dead":
        print(f"  !! TRUNCATED: the process is dead ({detail}) and the record has no outcome. To resume: {resume}")
        return 3
    if not ended:
        print(f"  !! no run outcome, and nothing says whether a process is alive: {detail}")
    return 0


def main() -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("root")
    ap.add_argument("--process", required=True, choices=("running", "dead", "unknown"))
    ap.add_argument("--detail", required=True, help="the pid and EXIT line, as the caller read them")
    ap.add_argument("--resume", required=True, help="the command that resumes this run")
    args = ap.parse_args()
    root = args.root
    dbs = sorted(glob.glob(os.path.join(root, "research", "*", "records", "record.db")))
    if not dbs:
        print("(no run record yet under %s)" % root)
        return process_verdict(args.process, args.detail, args.resume, False)
    db = dbs[-1]
    run = os.path.basename(os.path.dirname(os.path.dirname(db)))

    c = sqlite3.connect("file:" + db + "?mode=ro", uri=True)
    print(f"\n=== {run}")

    # Sittings per seat, in dispatch order. This is the line that says whether the engine is
    # moving: a seat list that stops growing is a run that has stalled, and nothing else shows it.
    try:
        # THE SEAT IS ON THE EVENT, NOT THE REGISTER BODY. `register` carries the agent's identity
        # fields; the seat id lives on the event row, so grouping the body table by "seat_id" reads
        # the literal column name back. Caught on a live run printing `seat_id×4`.
        rows = list(c.execute(
            'SELECT e."seat_id", count(*) FROM "events" e'
            ' JOIN "register" r ON r."event_id" = e."id"'
            ' GROUP BY e."seat_id" ORDER BY min(e."id")'))
        if rows:
            print("  seats:  " + "  ".join(f"{s}×{n}" for s, n in rows if s))
    except sqlite3.Error:
        pass

    # THE HOOK-WRITTEN ROWS, called out on their own. These are empty under a harness that spawns
    # top-level processes — the whole reason the universe exists — so seeing them non-zero is the
    # evidence that subagent dispatch is real here.
    spans = count(c, "sitting_open")
    if spans is not None:
        print(f"  hook-written sitting spans: {spans}"
              + ("   <- subagents are dispatching" if spans else "   <- none yet"))

    for label, table in (("minted", "mint"), ("closed", "close"), ("retired", "retire"),
                         ("positions", "position"), ("closings", "closing"), ("log", "log")):
        n = count(c, table)
        if n:
            print(f"  {label:<10} {n}")

    try:
        gaps = list(c.execute('SELECT "gap_id", "class" FROM "mint" ORDER BY "event_id"'))
        closed = {r[0] for r in c.execute('SELECT "gap_id" FROM "close"')}
        for g, cls in gaps:
            print(f"    {g:<5} {cls:<34} {'closed' if g in closed else 'OPEN'}")
    except sqlite3.Error:
        pass

    ended = False
    for table, label in (("gate", "chair verdict"), ("outcome", "run outcome")):
        try:
            for row in c.execute(f'SELECT "verdict" FROM "{table}"'):
                print(f"  {label}: {row[0]}")
                if table == "outcome":
                    ended = True
        except sqlite3.Error:
            pass

    # A TRUNCATED RUN AND A RUNNING ONE RENDER IDENTICALLY, and the board alone cannot tell them apart:
    # it just stops. `claude -p` terminates its background tasks on a ceiling, so a research run is
    # killed mid-debate and leaves exactly this — seats sat, gaps minted, no outcome. Measured on a
    # run whose last event was a lens's `mint` with no `sitting_close` after it. process_verdict is
    # what tells them apart.
    #
    # The unbalanced span is the evidence, and it is decidable here: every sitting the harness opens
    # it also closes, so opens > closes means a seat was cut off mid-sitting rather than a seat
    # being slow — the record cannot tell you WHEN, but it can tell you the arithmetic does not
    # balance, and that never happens on a healthy finished run.
    #
    # BOTH DIRECTIONS, PER AGENT. A sitting is one span with one close, so a close with no opening
    # since the agent's previous close is as wrong as an opening with no close. Totals hide it: m34
    # held 80 closes for 75 opens, and a global opens > closes check said nothing.
    unclosed, extra = sitting_balance(c)
    if unclosed:
        print(f"  !! {len(unclosed)} sitting(s) opened and never closed"
              f" — a seat is either still sitting, or was KILLED mid-sitting: {' '.join(unclosed)}")
    if extra:
        print(f"  !! {len(extra)} sitting close(s) with no open sitting to close"
              f" — one sitting closed twice, or a close with no opening: {' '.join(extra)}")
    return process_verdict(args.process, args.detail, args.resume, ended)


if __name__ == "__main__":
    raise SystemExit(main())
