# frank-exchange-of-views hooks — why each entry in `hooks.json` is wired the way it is

`hooks.json` holds the wiring; this file holds the reasons. They used to sit in the file itself as
`_comment` keys, and the client prints an "unknown key … ignored" warning for every such key at the
start of every session (#913), so the reasons live here, one section per event.
`scripts/validatejson` fails any key in a plugin `hooks.json` that the hooks reference does not
document.

## Every hook — run and identity injection

RUN AND IDENTITY INJECTION (#281, #510), one dedicated binary. When a Bash call is made inside a live
run (the payload's cwd finds a run marker, and the marker's live row names the payload's `session_id`
as the run's lead), PreToolUse REWRITES the command, prefixing
`export FEOV_RUN='<runDir>'; export FEOV_AGENT_ID='<agent>'; export FEOV_AGENT_TYPE='<type>'; ` — so a
seat stops typing an absolute path it can mistype (measured: one typo cost five manifest receipts
and produced a false bug report), and never loses the identity that binds it to its seat on the
record. UNCONDITIONAL, WITH NO MATCHER on command text: injection was once gated on recognising
`feov-record` in command position, a pattern standing in for a schema whose miss was silent — no
injection, no error, and a seat's identity simply absent, which reads exactly like a main-session
call. The shape was widened three times, each after a run had already paid (a heredoc bail cost one
seat its bibliography, command substitution cost another its identity, and `RB="…/feov-record"; $RB
…` — a seat aliasing the long path, which no command-text matcher can see — cost 21 of 65 registers
across six runs their agent_id). The harm the matcher claimed to prevent cannot occur: injection
PREPENDS, so a documentation write or a log entry carrying the token is byte-identical either
way. A needless `export` on a command that does not use the tool is inert; a missing one destroys a
seat's work. THE TYPE IS A DIFFERENT FACT FROM THE HANDLE and refuses a different thing: agent_id
says WHICH agent, agent_type says which CONFIGURATION it was dispatched as, so `register` can reject
a seat id from the wrong family — the membership half the roster gate states outright that it
cannot reach. Under #709 this binary is the ONLY reader of agent_type left, where it used to be the
lockdown's; it is parsed here and nowhere else. THIS USED TO BE TWO JOBS: it also enforced the
blue-report write-lockdown (a PreToolUse deny of a non-author raw write to blue/report.md, and a
PostToolUse backstop for a dropped finding-marker). Under report-as-record (#709) there is no
blue/report.md file — the report is the record, every change is an appended event, and a raw write
cannot reach it — so the lockdown and its PostToolUse backstop are gone; only this injection remains.
DEDICATED SINGLE-PURPOSE BINARY rather than a verb on feov-record: as `feov-record hook pretooluse`
the invocation carried no --seat-id, the identity-scoped surface refused it, and the refusal's exit
2 denied every mutating tool call in the session. The bootstrap guard mirrors prosthetic-conscience:
a fresh plugin-cache version ships without binaries, and an unguarded hook would crash-storm every
tool call in that window; the guard hands a missing binary to hooks/fetch-bin.sh, which installs it
from the plugin's pinned release in the background. The
decision ALWAYS travels in the stdout JSON with exit 0, enforced at the process boundary by
hookcmd.Run rather than documented on a leaf.

## PreToolUse

MATCHER `Bash|Read|StructuredOutput`, ONE ENTRY: the three tools the binary acts on. A Bash call gets
the run and identity injection above, and is refused when a command naming `feov-record` carries a
backtick the shell would run. A Read of the assembled report (`report.md` or `report.html` under the
run directory) by a seat is refused: the report is read through the tool. `StructuredOutput` goes to
the envelope gate below. Every other tool reaches no decision in the binary, so it is not spawned for
one. Under Antigravity the manifest matches every tool, because its tool names are not these.

The length of a sitting is the client's to bound, not this hook's: every agent definition carries
`maxTurns` (written by `scripts/agentgen`), and a seat that reaches it ends without its envelope —
an ordinary failed sitting.

THE ENVELOPE GATE covers `StructuredOutput`, the call a seat returns its envelope with and that ends
its sitting. The hook asks feov-sitting-write (`-phase envelope`) whether the record refuses it: while
a blocking item on the seat's own work list stands, the call is denied naming each item, and the
refusal is logged under the seat. A check that cannot run refuses the envelope naming the error. The
CHAIR'S ENVELOPE CARRIES NO PLAN: the workflow script's only surface is a seat's envelope, so the hook
replaces the chair's call with the plan the record stored last in its sitting (`updatedInput`, measured
reaching a Workflow seat's result on 2.1.300), `notes` kept, and denies a call that carries a field of
the plan or comes with no plan standing. The
gate admits nothing: an envelope it lets through still faces the harness's schema check, which runs
inside the tool's own call, after every PreToolUse hook. The admission is PostToolUse's.

THE RECORD'S READ goes on the same call. An envelope the record does not refuse is allowed with the
input it carried and `record_read` set to what feov-sitting-write answers (`record.RecordRead`: the
event read at, the hearings owed, whether the bench has halted, the outcome once stamped), overwriting
anything the seat typed there. The engine runs in a sandbox with no tools, and this field is the only
thing it learns of the record, so an envelope that comes back without it, or with a value missing a
required field or naming an event below 1 (what a seat types there), stops the run UNVERIFIED and names
the seat. `capture`'s envelope-read audit recomputes each read at the event it names, which is what
catches a seat that typed a plausible event under a dead hook. Claude Code
2.1.305 is the floor (`requirements.json`), which `sc-doctor` and `setup` both hold.

## PostToolUse

MATCHER `StructuredOutput|Bash`, ONE ENTRY: the binary is spawned for the call that returns a seat's
envelope, and for every Bash call (below). PostToolUse fires on a call that SUCCEEDED, so when it fires the envelope gate refused
nothing and the harness accepted the envelope against its schema. The hook hands the call to
feov-sitting-write (`-phase admission`), which records the envelope as ADMITTED — an
`envelope_admitted` event under `harness`, keyed to the agent, one per sitting. That admission is what
tells a finished sitting from a failed one: a closed sitting of a seat whose work arrives as items,
with no admission, failed (its envelope was refused by the record or rejected by its schema until the
harness gave up, or its agent died), and capture names it. A call the schema rejects fails, and
reaches PostToolUseFailure, which logs it as the seat's `failure` entry. An admission the record does
not take goes on the hook record (stage `envelope-admission`), and the sitting reads as failed. A call
from no seat — the main session, an operator — writes nothing. The bootstrap guard mirrors every other
hook here.

ON A Bash CALL, A LENS IS HANDED ITS NEXT ITEM. A lens's work arrives one item at a
time — SubagentStart hands it the current item in full and an index of the rest — and the next item
has to arrive after the act that discharged the current one, which is a Bash call to the record. So the
binary returns at once for every Bash call that is not a lens's (the payload's `agent_type`, read
without the roster: `runlive.IsLensAgentType`) or whose command does not name `feov-record`; for the
rest it hands the call to feov-sitting-write (`-phase next`), which reads the lens's work list, orders
it into its queue (`record.QueueOf`), and compares it with what the lens was last handed — the current
item's key and digest and the keys queued beside it, one file per lens sitting under the run's
`.work-delivery/`. The current item stays current until it leaves the queue; a moved one is handed
again marked `changed_since_delivered`; another item leaving is one line; nothing changed is nothing.
Two calls returning together hand the next item once: the comparison runs under an O_EXCL marker. What
the writer prints is the lens's, passed on as `additionalContext` — which reaches a Workflow seat's
context and an Agent subagent's alike, up to the harness's 10,000-character cap, past which it is
spilled to a file in the PARENT session with a two-kilobyte preview (measured on 2.1.318); a delivery
is held to 9,000 bytes (`sittingwrite.DeliveryBudget`) by construction: past it, parts give way in a fixed
order (`sittingwrite.fitted`) down to the current gap's own prose, and each cut is said in an `_omitted`
or `_clipped` field. A writer that fails goes on the hook record
(stage `next-item`) and is said to the lens as a statement that nothing was determined, never as
silence. Under Antigravity a PostToolUse response carries nothing back to the agent, so a lens there
reads its next item off its work list.

## SubagentStart

Opens the span. THE SITTING SPAN (#265). Seat durations read ZERO because both ends came from
FILESYSTEM MTIMES: on the fallback path StartedMs == EndedMs for every completed seat so the duration
branch never fired, and a running seat's 'elapsed' measured time since the last WRITE. These two
events put both ends on the record as facts a hook OBSERVED rather than values derived from file
metadata. NO MATCHER — these events take none.

THE BINARY SPEAKS TO THE SEAT, AND THIS IS THE ONE EVENT IN THIS PLUGIN THAT CAN. A
`{"hookSpecificOutput":{"hookEventName":"SubagentStart","additionalContext":"…"}}` document reaches
the DISPATCHED SEAT's context and nothing else — one firing, verified three times
(plans/hook-surface-spike.md §5 and §10, and again 2026-09-25, where the marker landed as a
`hook_additional_context` attachment on the seat's transcript and the seat returned it verbatim). The
nine-firing loop described under SubagentStop is that event's and does not apply here. What this
event injects is the seat's own WORK LIST, read from the record at the moment of dispatch and
rendered by the writer, so a seat that owes nothing can end its turn having run no commands at all.
It is the projection's bytes, not a second author of the seat's duties.

FAILURE IS A DIFFERENT CHANNEL FROM INJECTION. This event displays nothing to a HUMAN, so a failure
to record the span goes on the hook record and is read out by the seat's next Bash, Read or
`StructuredOutput` call (PreToolUse).
A work list that could not be read says so in the injected text instead — "I could not tell" and
"nothing engages you" must never be the same bytes.

ONLY A SEAT IS RECORDED: the sibling event also fires at the MAIN agent's turn end, with a minted
agent id and no agent_type — 19 seats against 50 turn ends in one measured session, separated by
agent_type with zero exceptions either way (§7a). An event with no agent_type is not a sitting and is
dropped. AND ONLY THE RUN'S OWN SESSION'S SEAT: a run's members are the agents of the session that
leads it, and every subagent's payload carries its dispatching session's `session_id` — so an agent of
another conversation, in a worktree nested under the run's checkout, finds the marker and is still
attributed to nothing. A payload with no `session_id`, or one of this plugin's agents from a session
that leads no live run there, goes on the hook record (stage `run-membership`). That record is the user-wide failure record (`failures.json` under `$XDG_STATE_HOME/special-circumstances/frank-exchange-of-views/`, else `~/.local/state/…`), scoped to the marker's project, and the fault clears there when a later payload in that project is attributed to the run. The bootstrap guard
mirrors every other hook here: a fresh plugin-cache version ships
without binaries and an unguarded hook would crash-storm the window; the guard hands the missing
binary to hooks/fetch-bin.sh, which says nothing on this event. THE PAYLOAD CANNOT NAME THE SEAT — it
carries the harness handle and the agent configuration and nothing the workflow supplied (#290,
measured 2026-08-23) — so the seat is resolved from the configuration where that is one-to-one, and
otherwise recovered later by joining agent_id to the register event that names it.

## SubagentStop

Closes the span. THE SITTING SPAN (#265). Seat durations read ZERO because both ends came from
FILESYSTEM MTIMES: on the fallback path StartedMs == EndedMs for every completed seat so the duration
branch never fired, and a running seat's 'elapsed' measured time since the last WRITE. These two
events put both ends on the record as facts a hook OBSERVED rather than values derived from file
metadata. NO MATCHER — these events take none. THE BINARY WRITES NOTHING TO STDOUT, and that is the
contract rather than a minimal implementation: a SubagentStop hook that emits `additionalContext`
RE-INVOKES the seat and fires again — nine firings for one seat in the measured case, the returned
context discarded every time (plans/hook-surface-spike.md §10). A log-only hook fires exactly once.
Anything added here that talks turns one event into nine. ONLY A SEAT IS RECORDED: SubagentStop also
fires at the MAIN agent's turn end, with a minted agent id and no agent_type — 19 seats against 50
turn ends in one measured session, separated by agent_type with zero exceptions either way (§7a). An
event with no agent_type is not a sitting and is dropped. The bootstrap guard mirrors every other
hook here: a fresh plugin-cache version ships without binaries and an unguarded hook would
crash-storm the window; the guard hands the missing binary to hooks/fetch-bin.sh, which says nothing on this event.

NOTHING DOWNSTREAM REFUSES A MISSING CLOSE: a sitting with no close has not failed, so it reads as
neither sat through nor failed, and a run whose every close vanished reads zero failed sittings while
no lens retires. `capture`'s sitting-closes audit reads the record for it: it FAILS a run whose
sittings ended (their seat sat again) with not one close, and names each ended sitting with no close.
