# gray-area hooks — why each entry in `hooks.json` is wired the way it is

`hooks.json` holds the wiring; this file holds the reasons. They used to sit in the file itself as a
`_comment` key, and the client prints an "unknown key … ignored" warning for every such key at the
start of every session (#913), so the reasons live here. `scripts/validatejson` fails any key in a
plugin `hooks.json` that the hooks reference does not document.

## Every hook

Every command is wrapped in the bootstrap guard prosthetic-conscience established: a fresh
plugin-cache version ships from git WITHOUT binaries (they arrive via doctor --fix), and an
unguarded hook crash-storms every tool call in that window. The guard degrades to ONE stderr line
pointing at the fix.

## SessionStart

SessionStart records where the MAIN session's transcript is — SubagentStop only ever hands over a
SEAT's, so without this row a consumer has no non-guessing way to find the session's own trajectory
(plans/gray-area.md §11.2). The event is passed with -event rather than inferred from the payload:
the wiring already knows which hook it registered for.

## Stop and SessionEnd

Stop and SessionEnd feed the CATALOGUE and write no manifest row: Stop is the turn boundary, where a
session's own transcript has just grown, and SessionEnd is the last chance to read it before the
session is gone. Their cost is the read of one file's tail, not a reprojection.
