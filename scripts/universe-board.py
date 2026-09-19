#!/usr/bin/env python3
"""Print the live state of a run's BOARD, read-only, from the record.

The stream view shows what the harness is doing — which seats were dispatched, which verbs ran.
This shows what the DEBATE is doing: who has sat, what is on the board, what closed, where the
epochs got to. Those are different questions and only the second one says whether the run is
going well.

READ-ONLY BY CONSTRUCTION. Opened with SQLite's `mode=ro` URI, so watching a live run can never
take a write lock or disturb a seat mid-sitting. The run directory is the database; nothing here
writes to it.
"""
import glob
import os
import sqlite3
import sys


def count(c: sqlite3.Connection, table: str) -> int | None:
    try:
        return list(c.execute(f'SELECT count(*) FROM "{table}"'))[0][0]
    except sqlite3.Error:
        return None


def main() -> int:
    root = sys.argv[1] if len(sys.argv) > 1 else "."
    dbs = sorted(glob.glob(os.path.join(root, "research", "*", "records", "record.db")))
    if not dbs:
        print("(no run record yet under %s)" % root)
        return 0
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

    for table, label in (("gate", "chair verdict"), ("outcome", "run outcome")):
        try:
            for row in c.execute(f'SELECT "verdict" FROM "{table}"'):
                print(f"  {label}: {row[0]}")
        except sqlite3.Error:
            pass
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
