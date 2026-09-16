# gray-area hooks — why each entry in `hooks.json` is wired the way it is

`hooks.json` holds the wiring; this file holds the reasons. They used to sit in the file itself as a
`_comment` key, and the client prints an "unknown key … ignored" warning for every such key at the
start of every session (#913), so the reasons live here. `scripts/validatejson` fails any key in a
plugin `hooks.json` that the hooks reference does not document.

## Every hook

Every command is wrapped in the bootstrap guard prosthetic-conscience established: a fresh
plugin-cache version ships from git WITHOUT binaries, and an unguarded hook crash-storms every tool
call in that window. The guard hands a missing binary to hooks/fetch-bin.sh, which installs the
plugin's binaries from its pinned release in the background (prosthetic-conscience's hooks/README.md
describes the fetcher).

## SessionStart

SessionStart records where the MAIN session's transcript is — SubagentStop only ever hands over a
SEAT's, so without this row a consumer has no non-guessing way to find the session's own trajectory
(plans/gray-area.md §11.2). The event is passed with -event rather than inferred from the payload:
the wiring already knows which hook it registered for.

SessionStart also REGISTERS the session in the catalogue before it closes out any other: a row for
it, reopened if closure had settled it, and the Remote Control cloud id copied from the client's own
session file (`~/.claude/sessions/<pid>.json`, the hook's parent). That file exists at SessionStart
and is gone by SessionEnd, and transcripts do not reliably carry the id, so this is the moment it
can be captured — and it is what `telepathy agents --lost` offers as the reattach after a restart.
The session counts itself live for that closure pass, so it is never closed by its own start.

## Stop and SessionEnd

Stop and SessionEnd feed the CATALOGUE and write no manifest row: Stop is the turn boundary, where a
session's own transcript has just grown, and SessionEnd is the last chance to read it before the
session is gone. Their cost is the read of one file's tail, not a reprojection.

Stop registers the session as SessionStart does, in the same open, before it ingests: a session
resumed after closure is reopened at its first turn end, and a session whose SessionStart ran before
capture was installed still gets its cloud id. SessionEnd only ingests — the session file the
registration reads is already gone by then.

## What happens when the hook fails

Every event here exits 0 whatever happened — a broken store must never cost a session its turn — and
a hook's stderr at exit 0 reaches the client's debug log and nobody (measured: it appears in the
transcript only inside a `hook_success` record whose `content` is empty, and a live session asked
straight afterwards had not seen it). So each failure is also a field on a record,
`~/.local/state/special-circumstances/gray-area/failures.json`, keyed by the STAGE that failed and
the project it failed in, written by `internal/hookfailures` (generated from
`scripts/internal/hookfailures`, shared with the other plugins).

SessionStart and Stop say what is due as one top-level `systemMessage`, each stage at most once every
10 minutes. SubagentStop and SessionEnd record and stay silent: neither displays a message, and a
subagent hook that talks can re-invoke its seat. An entry clears when its stage next works, so NO
FILE means every stage last worked — the one reading that an absent record must never have is
"nothing looked".
