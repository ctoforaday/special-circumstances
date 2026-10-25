# Run membership by launching session

**Ruling (gblock): "By launching session".** A subagent belongs to a run only if its hook payload's
`session_id` is the session that leads the run. Before this, `runlive.InferRunDir` walked up from the
payload's `cwd` to a `.claude/run-live.json`, so another conversation's agent in a worktree nested
under the run's checkout was attributed to the run (sitting spans, seat turns, `FEOV_RUN` injected).

## The key

`RunLiveMarker.LeadSession` (`leadSession` on the marker row). `runlive.Caller` gains `Session`;
`InferRunDir` counts a live row only when `Caller.Leads(row)` (non-empty and equal). New inferences:
`OtherSession` (silent: an unrelated session), `Sessionless` and `SeatOfAnotherSession` (membership
faults, stage `run-membership`). The marker is a file, not the record: no event-schema change.

## Where the session id comes from

- Hooks: the payload's `session_id`. Evidence: every Workflow-seat SubagentStop payload gray-area
  captured across 24 universe runs carries the lead's id (`~/.claude/scratch/universe-*/src/.claude/
  gray-area/trajectories-<lead>.jsonl`, 0 mismatches); plans/hook-surface-spike.md §7 and §9d.
- Verbs (setup, take-over, CLI inference): `CLAUDE_CODE_SESSION_ID`, which the client sets in every
  Bash call — a subagent's included — to the same id (`seatenv.Session`).

## Every reader that resolves a run

Membership (session-keyed): `hookcmd.Pre` (injection, report-read deny), `enforceLimit`,
`refuseEnvelope`, `PostFailure` (all through `seatRunDir`), `sittinghook.handoff`
(SubagentStart/Stop), `cli/root.go dispatchedSeat`, `cli/seat.Of`.
Liveness only (zero Caller, no membership, unchanged): `PruneRunLive` (setup, capture),
`setup.refuseADecidedRun`, `dashboard --serve` and its watcher, `bench outcome`'s mark.
Not membership, deliberately unchanged: prosthetic-conscience's `pushfreezeguard` and
`toolchainnudge` — project-wide commitments while any run is live. `seatRunDir`'s `FEOV_RUN` /
wrapper fallback is the process's own environment, not inference, and stays.

## Setup

Records `CLAUDE_CODE_SESSION_ID` as the lead and refuses to build a run without one. The
one-live-run-per-project refusal existed only because the hooks could not tell two runs' seats apart;
the session key removes that for runs led by different sessions, so it narrows to ONE LIVE RUN PER
LEAD SESSION (two rows one session leads are still ambiguous). The decided-but-uncaptured refusal is
a different reason and stays. A re-setup of a live run another session leads is refused (in setup
and again in `WriteRunLiveMarker`).

## Resume re-binding

A resume (`resumeFromRunId`) may run in a new session. `feov-record take-over <runDir>
--from-session <lead>` moves the lead, refusing unless `--from-session` is the row's current lead, so
no session takes a run without naming the one it replaces. Until then the new session's plugin agents
are `SeatOfAnotherSession`: unattributed and recorded, naming the take-over. The research skill (step
3, step 4's resume) and `engine-stopped --help` say so.

## No session_id

Attributed to nothing; `Sessionless` goes on the failure record (`run-membership`, project-scoped),
cleared only by a payload that IS attributed — an unrelated session's calls never clear it.

## The post-outcome bench exception (#1535)

`StateOf`'s decided-run exceptions (the recording agent, any `lead-judge`) now also require
`Caller.Leads(row)`. Through the hooks membership already gates it; `StateOf` keeps its own check.

## Tests

Hook entry points with real payload JSON and the real writer: `sittinghook/membership_test.go`,
`hookcmd/membership_test.go` cover (a) another session's Explore in a nested worktree, (b) the lead's
seat, (c) resume + take-over and no stealing, (d) the bench exception, (e) no `session_id`, plus
PostToolUseFailure. Units: `runlive/membership_test.go`, `setup/membership_test.go`, the CLI
end-to-end `cli/takeover_test.go`. Each guard was deleted in turn and the suite run.
