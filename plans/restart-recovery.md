# Restart recovery — `lost` sessions in `telepathy agents`, each with its recovery commands

> STATUS: implementation (2026-09-11). Revision 16 — the spec as implemented. Round 5 FAILED on the
> MODE read (a real defect: the newest `user` record is usually a tool result with no mode — fixed
> here) and on test and wording detail. Per gblock ("if you're hitting diminishing returns on the
> audit and feedback is code level you can move to implementation") there is no round 6: every
> round-5 gap is folded in below, and the independent review checks them. Revision 15 — gblock ruled (2026-09-11) that a
> recovered session resumes in EXACTLY the state it began in, which retires the mode pin: the
> reattach carries no mode flag (a server-respawned session keeps its own, §II) and the local
> resume passes the session's own last recorded mode; also answers plan-audit round 4. Revision 14
> answered round 3 and recorded two probes run with gblock on 2026-09-11 (§II): a restarted server
> brings its own sessions back, and a LOCAL resume honours `--permission-mode` while a SERVER
> respawn did not. The pin-and-test ruling below is superseded. Revision 13 answered
> plan-audit round 2. gblock
> ruled (2026-09-11): the reattach command pins `--permission-mode default` too, measured in ONE
> consented Remote-Control-on probe run, and dropped from the rows if the measurement fails; that
> run is allowed. Revision 12 answered plan-audit round 1 of the
> rebased plan: time is TRANSCRIPT time, the mode check reads a source that exists, only `--lost`
> needs a boot time, only `live` sessions are excluded, a verified cwd is never overwritten, and
> the captured column gets a real-data check. Revision 11 was revision 10 rebased onto current main:
> the store-shape rebuild merged (#894), the speaker classifier made the store shape 4 (#902), and
> `find`'s offset attribution merged (#908, no shape change). This plan's column therefore makes
> the store shape **5**, line references are re-measured, and three carriers revision 10 missed are
> listed (§III.5). Revision 10 was rewritten BEST-EFFORT by gblock's ruling after eight plan-audit
> FAILs on revisions 1–9, whose guarantees ("never a mismatched directory", "always listed again")
> each invited a new edge case. Home: gray-area, with `plans/gray-area.md` §4 amended (§III.5).
> Revisions 1–10 were never committed; the measured facts they established are carried in §II.

## I. Summary & Goals

**The problem.** On 2026-09-10 the host shut down for a kernel swap (09:05 → 16:33) and every
Remote Control session vanished from the app. Nothing was deleted, but nothing recorded which
sessions had been live, so recovery was archaeology. Resuming by local id from the right directory
worked for all three sessions that mattered; reattaching by cloud id worked for the one whose
environment had not expired.

**Objective.** After a restart, `telepathy agents` shows the sessions the restart cut off, as
state `lost`, and each such row says exactly what recovery needs: the session id, where to run
from, the cloud id if known, and the commands to run. A skill walks the human through choosing
and running them. **Best-effort by design:** the list is a set of candidates for a human, and the
resume attempt itself is the check — `claude --resume` from the wrong directory, or a reattach to
an expired environment, fails loudly, and the skill falls back and reports.

**Success criteria.**

1. After a restart, `telepathy agents` lists, below the live sessions, every session whose last
   sign of life precedes the boot, lies within `--window` (default 48h) of the newest pre-boot
   activity, was not closed by the catalogue after that activity and before the boot, and is not
   `live` — state `lost`. **Every time here is TRANSCRIPT time**, never read time: a session's
   *last sign of life* is the newest `ts` across its `act`, `word` and `thought` rows; *the newest
   pre-boot activity* is the largest such value below the boot across all sessions. A session
   registered but with no rows yet falls back to `ingested_first` — set once, at the hook that
   registered it, while the session was running — and its row says `(registered, no records)`.
   `ingested_last` is never read (`schema.go:41-45`: it is when the catalogue read the file, which
   `backfill` and the 4 → 5 rebuild move past the boot for every session).
2. `telepathy agents --lost` prints only those, and for each: session id, `RESUME FROM`, `CLOUD`,
   last sign of life, `MODE` — the session's own last recorded permission mode: the
   `permissionMode` of the newest `user` record THAT CARRIES THE FIELD (tool-result `user` records
   never do — round 5 measured 11,141 of them, none with it), read at query time from the
   session's own transcript (its `file_offset.path` with `agent_id = ''`, else
   `project_dir/<session_id>.jsonl` for a session registered with no rows; no new column), else
   `unknown` — and the two commands: `claude remote-control --session-id <cloud>`
   (when a cloud id is known; NO mode flag — a server-respawned session keeps its own mode, §II)
   and `claude --resume <id> --permission-mode <MODE> --remote-control` (a recorded `default` is
   written `manual`, the spelling `claude --help` lists, §II; the flag is omitted when MODE is
   `unknown`, and the row reads exactly `MODE unknown — the resume starts in whatever mode the CLI
   chooses`, naming no mode, because what a flagless resume chooses is measured only in §V.4).
   **A recovered session resumes in exactly the state it began in** (gblock, 2026-09-11): the
   reattach keeps it by construction, and the local resume is given it — measured to override the
   CLI's own default (§II).
3. The skill first has the human open, in the app, the sessions a running Remote Control server
   can bring back itself (§III.4); then backfills, gets the human's choice, runs the commands in
   detached tmux SESSIONS, falls back on failure, and reports per session what happened.
4. Driven on real data (§V.4): a probe session killed with `SIGTERM` shows as `lost` with commands
   that, when run, bring it back.

**Accepted costs.** False positives (clean exits after the last closure sweep; sessions deferred by
closure's cap; sessions that never finished a turn) — the human filters them. Idle-but-live
sessions beyond `--window` are missed; widening it recovers them. `RESUME FROM` may be wrong for
unusual sessions (moved folders, very long paths); the resume attempt reveals it and the skill
asks. Display names are server-side and must be set again by hand. A store rebuild loses captured
cloud ids AND every `closed_at` — and shipping this plan IS one (shape 4 → 5 rebuilds the store
empty on first open). `telepathy backfill` only ingests (`backfill.go:48`); only closure sets
`closed_at` (`sweep.go:99`). So after the upgrade the cloud ids accrue again from the next
`SessionStart`/`Stop` (the fallback command works without them), and every session closed before
the boot is unclosed again and can appear as a `lost` candidate once — more false positives, for
the human to filter, until the next boot. Within a server's bring-back window (§II), a
server-spawned session the daemon would bring back LAZILY reads `lost` until someone opens it in
the app — the skill's first step exists for exactly that. Each reattach leaves its own
single-session "at capacity" entry in the app until that session ends; each local resume registers
one remote session per process. A session that ran in `auto` before the restart resumes in `auto`
(gblock's ruling: exactly the state it began in); the row shows `MODE`, and the skill names such
sessions in the choice it puts to the human (§III.4).

**Out of scope.** An unattended boot-time resumer; any guarantee beyond "the commands were run and
their outcome reported".

## II. Technical Context

- Module `plugins/gray-area/tools`. No new binary, no new hook event, no new verb. Base: current
  main (5915be84), store shape 4 (`schema.go:177`, `UserVersion = 4`).
- **The shape bump uses #894's mechanism unchanged.** `recognise` (`open.go:147-165`) classes a
  stamp below `UserVersion` as `Older` when `file_offset` exists, and `Open` rebuilds an `Older`
  store inside one `BEGIN IMMEDIATE` transaction, marking it `rebuilt_at`/`rebuilt_from` so every
  read verb warns until `telepathy backfill` runs. Recognition keys on `sessionCore`, the five
  `session` columns every shape has carried (`open.go:139`, restated in the `schema.go:30-35`
  comment); `bridge_session_id` is NOT added to `sessionCore`, so neither needs to change.
- **`agents` carries no speaker text.** It renders `SESSION STATE ACTS LAST CWD`, where LAST is the
  last tool and its age (`agents.go:47-57`, from `v_action` in `query.go:23-39`). #902's speaker
  vocabulary (`v_word.role`, `find`'s IN) does not reach it, and the `lost` rows add none.
- **`agents` today lists only advertised sessions.** `catalogue.Agents` (`query.go:23`) starts from
  `~/.claude/sessions/*.json` and joins the store for ACTS and LAST; a session with no session file
  never appears. The `lost` rows therefore come from a separate store query (§III.3), not from
  widening `Agents`.
- **Release dependency.** Installed gray-area here is 0.9.0 (no catalogue hooks). Nothing reaches a
  consumer until a gray-area release — gblock's call.
- **Measured 2026-09-10** (each re-checked by at least one independent plan-audit round):
  - `~/.claude/sessions/<pid>.json` (the only local holder of the cloud id, as
    `bridgeSessionId: session_<suffix>`) exists at `SessionStart` and is gone by `SessionEnd`; the
    reattach flag takes `cse_<suffix>`. Transcripts do not reliably record the cloud id
    (`identity`'s has none) — hence the one captured column. `SessionFile` (`liveness.go`) does not
    read it today.
  - `SessionEnd` fires on `SIGTERM`/`SIGHUP` with reason `other`; `SIGKILL` fires nothing.
  - `closed_at` is set by the `SessionStart` closure sweep and never cleared, so a session resumed
    after closure would otherwise never read as unsettled again (`schema.go:48-51`).
  - A transcript lives under `~/.claude/projects/<key>/`, `<key>` = the directory with every
    character outside `[a-zA-Z0-9]` replaced by `-` (hashed past 200 characters). For 140 of 141
    transcripts measured, the first record's `cwd` encodes to the key; `EnterWorktree` sessions
    move folders.
  - Display names are server-side only (`nameSource: derived` in every local session file).
  - Resumed sessions reached the app here only because user settings enable Remote Control at
    startup — hence `--remote-control` on the fallback.
  - The daemon-spawned `otel` session took a turn by itself in auto mode on resume — the case R4
    accepts under gblock's 2026-09-11 ruling, with every row showing `MODE` and the skill naming
    `auto`/`bypassPermissions` sessions in the choice it puts to the human.
- **Measured 2026-09-11 (plan-audit round 1 of revision 11):**
  - Session files carry NO permission mode. The keys of all 6 live files: `bridgeSessionId, cwd,
    entrypoint, kind, messagingSocketPath, name, nameSince, nameSource, peerFeatures,
    peerProtocol, pid, pidDomain, procStart, sessionId, startedAt, status, statusUpdatedAt, tmux,
    updatedAt, version`. Transcript records carry `permissionMode` (1,340 user records in the
    corpus census), and so does the hook payload (`main.go:172`).
  - `/proc/stat` btime reads 1789058032 = 2026-09-10 16:33:52 UTC, matching the reboot.
  - A SIGKILL leaves the session file behind; `liveness.go` reports such a file `ended` (its
    process is gone). Whether the client later removes stale files is NOT measured, and this plan
    does not depend on it (§III.3 excludes only `live`).
  - Ingest projects only the bytes after the stored offset (`ingest.go:144-154`), so any rule over
    a session's cwds sees one pass at a time.
- **Measured 2026-09-11 — two probes run with gblock** (scratch repos under `~/.claude/scratch/`,
  since removed; Claude Code 2.1.268):
  - **A restarted server brings its own sessions back.** A scratch `claude remote-control
    --spawn=worktree` server was stopped with SIGTERM (what systemd sends at shutdown) and started
    again with the identical command in the same directory. Both its sessions returned under the
    SAME local `sessionId` and cloud `cse_` id, in the same app entry, with no new device entries:
    the pre-created session at once, the on-demand worktree session LAZILY — nothing for 12
    minutes, then back the moment gblock opened it in the app. The docs give "about four hours
    after the server stopped" (https://code.claude.com/docs/en/remote-control, "Resume sessions
    after stopping the server"); the 2026-09-10 outage was 7.5 h, past it — which is why nothing
    came back that day. The real daemon, a real reboot and the four-hour limit were NOT tested.
  - **Respawned children resume from the cloud copy.** Their command line is `--print --sdk-url
    …/cse_X --session-id cse_X --resume=https://api.anthropic.com/v1/code/sessions/cse_X
    --permission-mode <server flag>`, and the local transcript was REWRITTEN from the cloud copy
    (29 → 18 lines, every user and assistant record kept, local-only records dropped). Ingest
    already re-reads a transcript whose head changed or that shrank (`ingest.go:132`).
  - **A server respawn does not apply its `--permission-mode` to a session's own mode.** The
    on-demand session, created from the app, recorded `auto`; after the respawn its child command
    line said `--permission-mode default` and its next prompt still recorded `auto`. The
    pre-created session recorded `default` throughout. So a reattach, which reaches the session
    the same way, resumes it in the mode it had with no flag at all — what gblock's ruling asks
    for (§I criterion 2).
  - **A local resume does apply it.** Interactive sessions, Remote Control off: a fresh `claude`
    here starts in `auto` (status line "⏵⏵ auto mode on"; no settings file sets `defaultMode`);
    `claude --resume <that id> --permission-mode default` and `… --permission-mode manual` both
    switched it to "⏸ manual mode on", recorded `"default"`; `… --permission-mode plan` switched
    it to plan mode.
  - **`default` is spelled `manual` in the main CLI now.** `claude --help` lists `acceptEdits,
    auto, bypassPermissions, manual, dontAsk, plan`; `claude remote-control --help` lists
    `acceptEdits, auto, bypassPermissions, default, dontAsk, plan`; both spellings are accepted by
    `claude --resume`, and the transcript records `"default"`.
  - **A reattach is its own server.** `claude remote-control --session-id` "can't be combined with
    … `--spawn`, `--capacity`", serves only that session, exits when it ends, and shows in the app
    as a separate "at capacity" device entry (seen by gblock on 2026-09-11 for `identity`'s).

## III. Proposed Changes

```
plugins/gray-area/
├── skills/restart-recovery/SKILL.md          [NEW]
├── skills/telepathy/SKILL.md                 [MODIFY] agents' lost state + pointer to restart-recovery
├── README.md, ../../README.md                [MODIFY] agents' lost state; the captured column
├── requirements.json, hooks/README.md        [MODIFY] gray-area-capture purpose; the hooks rationale (hooks.json holds no prose since #915): registration at SessionStart and Stop
└── tools/
    ├── internal/catalogue/schema.go          [MODIFY] session.bridge_session_id; v_session +column (:144-149); ViewColumns["v_session"] (:182); UserVersion 5 (:177); closed_at comment (:48-50)
    ├── internal/catalogue/register.go (+test) [NEW] RegisterSession
    ├── internal/catalogue/ingest.go (+test)  [MODIFY] cwd from the session's own transcript; clear closed_at on new bytes
    ├── internal/catalogue/project.go (+test) [MODIFY] Projection gains CWDs
    ├── internal/catalogue/lost.go (+test)    [NEW] the lost query and its recovery commands
    ├── internal/catalogue/boot_{linux,other}.go (+test) [NEW] boot time
    ├── internal/catalogue/liveness.go        [MODIFY] SessionFile gains BridgeSessionID
    ├── internal/telecli/agents.go            [MODIFY] lost rows; --lost, --window, --boot; help
    ├── internal/telecli/root.go              [MODIFY] Env.Boot (:37-42), defaulting to the /proc/stat reader; the :93 example gains `telepathy agents --lost`
    ├── internal/telecli/session.go           [MODIFY] :18-19 (points at `agents --lost`), :33-34 and help: closed_at is cleared on a sign of life
    ├── internal/telecli/testdata/golden/     [NEW] agents-lost*.golden, agents-all-ended.golden; [MODIFY] agents, agents-rebuilt-warning, help-agents, help-root, help-sql, help-session
    └── cmd/gray-area-capture/
        ├── catalogue.go                      [MODIFY] sweep(in, stderr) (:83) registers + counts itself live; atStop(in, stderr)
        └── main.go                           [MODIFY] Stop → atStop (:526); SessionEnd unchanged; sweep(in, stderr) (:535); stale -event help
plans/gray-area.md                            [MODIFY] §4 amendment (:166)
plans/README.md                               [MODIFY] entry
```

### III.1 Capture: registration at `SessionStart` and `Stop`

`RegisterSession(db, sessionID, projectDir, bridgeID, now)` upserts the session row (setting
`ingested_first`/`ingested_last` on insert), sets `closed_at = NULL`, and records
`bridge_session_id` when non-empty (never overwriting a known id with empty). `bridgeID` comes from
`~/.claude/sessions/<ppid>.json` only when its `sessionId` equals the payload's. At `SessionStart`,
`sweep` registers BEFORE closure and adds the payload's id to closure's `live` map; `Stop` opens
once, registers, then ingests. `SessionEnd` is unchanged (ingest only). Every failure: one stderr
line, exit 0. `project_dir` is `dir(transcript_path)` or `''`, never `"."`.

### III.2 Ingest: cwd and reopen

`Projection` gains `CWDs` (distinct record `cwd`s in order). For the session's own transcript, on
each pass (ingest sees only the new bytes): a cwd MATCHES when its unhashed key equals the file's
folder name. If this pass has a matching cwd, write the latest one. Otherwise write this pass's
first cwd ONLY if the stored `session.cwd` is empty or itself does not match — **a matching value is
never replaced by a non-matching one**, so an `EnterWorktree` move's later passes cannot overwrite
the verified directory. A non-matching value is a hint, labelled `(unverified)` in output. New
bytes (`BytesRead > 0`) set `closed_at = NULL` — the session is reopened for closure; this is not
the "last sign of life", which is transcript time (§I criterion 1).

### III.3 `telepathy agents` gains the `lost` state

- `agents` keeps listing advertised sessions with their MEASURED liveness (`live`/`ended`/
  `unknown`), and appends store rows selected by §I criterion 1 with state `lost`, which is
  INFERRED — the `Liveness` type is not widened; the help says which values are measured and
  which inferred. **Only `live` excludes a session from `lost`.** An advertised session that reads
  `ended` (a SIGKILL or crash left its file behind) and meets criterion 1 is shown ONCE, as a
  `lost` row with its commands, not as an `ended` row — that is the unplanned-restart case. An
  `unknown` session (a foreign pid namespace or unsupported platform) is not ours to judge: it
  stays an `unknown` row and is never `lost`.
- `--lost` prints only lost rows, one block per session: id, `RESUME FROM` (with `(unverified)`
  when the `cwd` did not match its folder key), `CLOUD` (`session_X` → `cse_X`, else `unknown`),
  last sign of life (transcript time, or `(registered, no records)`), `MODE` (§I criterion 2), and
  the commands of §I criterion 2, ready to paste.
- `--window` (default 48h) and `--boot` (a Unix time). The boot defaults to `/proc/stat` btime on
  Linux. **Only `--lost` refuses without one** (*boot time not measurable here — pass --boot*,
  exit 1). Plain `agents` without a measurable boot prints that as ONE line under the table and
  still lists the advertised sessions exactly as today, so `agents` on Windows and macOS is
  unchanged.
- `Env` gains `Boot func() (int64, bool)` (`root.go:37-42`), defaulting to the `/proc/stat`
  reader; tests and goldens inject a fixed boot, so no golden depends on when the host booted or
  on its platform. A NIL `Boot` reads as "not measurable" (never a panic), and `Boot` is called
  only after `openRead` succeeds, so the existing `Env{` literals without it (`root_test.go:51,187`)
  keep their behaviour. `agents --lost` prints the boot it used and its source on its first line
  (`boot 2026-09-10 16:33:52 UTC (from /proc/stat)` or `(from --boot)`), and on its second the
  newest pre-boot activity — the largest last sign of life below the boot, §I criterion 1 — with
  the session it belongs to and the bring-back arithmetic the skill's first step needs — computed
  HERE, not by the agent: `newest pre-boot activity 2026-09-10 09:04:51 UTC (session 3aacb6a6) —
  7h29m ago, past the ~4 h a restarted Remote Control server brings its sessions back in`, or
  `… — 1h02m ago, inside the ~4 h window: open server-spawned sessions in the app first`. With no
  activity before the boot (the empty store after the 4 → 5 rebuild, before `backfill`) it reads
  exactly `newest pre-boot activity: none in the store — run telepathy backfill, then ask again`.
- **Every advertised session reads `ended` and moves to `lost`** (a crash leaves all files
  behind): the advertised table would be empty, and `agents.go:38`'s "no sessions are advertised
  … nothing is running" would be false — the files are there. That case prints `<n> session
  file(s) are advertised and none is live — each was cut off by the restart and is listed below as
  lost` instead of an empty table (`root.go:10-12`: a sentence, never an empty table). The golden fixture's gamma and delta (in the store, not advertised,
  `golden_test.go:43-44`) become `lost` rows under an injected boot after their activity.
- Worded absences: no store / other shape (existing); the boot line above; *the store holds no
  sessions — run `telepathy backfill`, or capture has not run here*; *nothing was cut off by the
  restart at <boot>*. The existing "no sessions are advertised" line (`agents.go:38`) stays true
  only of advertised sessions; with lost rows present it is printed above them, not instead of
  them.

### III.4 The skill — `skills/restart-recovery/SKILL.md`

- FIRST, let a running server do it. The docs' window runs about four hours from when the server
  STOPPED, and a server-spawned session comes back when it is OPENED — that is, now. The server
  stopped at or after the newest pre-boot activity, so `now − newest pre-boot activity` — which the
  second line of `agents --lost` computes and states as inside or past the window — is an UPPER
  bound on the time since it stopped: inside means the server will bring its sessions back when
  opened; past means it may not, and opening is still worth one try. A second line reading `none in
  the store` means `backfill` has not run: run it and ask again. A server
  is running again when `systemctl --user is-active claude-rc@<name>` prints `active` (on this box
  `claude-rc@special-circumstances`), or a `claude rc` / `claude remote-control` process has the
  repository as its cwd. Then tell the human to open, in the app, every `lost` row whose
  `RESUME FROM` is that repository or one of its `.claude/worktrees/bridge-cse_*` worktrees — the
  two shapes a server-spawned session has (§II) — and re-run `telepathy agents --lost`: a session
  that came back reads `live` and drops off; only what is still `lost` gets commands.
- BEFORE listing, run `telepathy backfill` (safe to repeat; catches sessions whose hooks never
  wrote a row, and clears the rebuild warning after an upgrade), then `telepathy agents --lost`.
- BEFORE resuming, YOU MUST get the human's choice — the list has false positives and resuming
  re-activates agents.
- Per chosen session: `tmux new-session -d -s rr-<first 8 of its id> -c <RESUME FROM>
  '<command>'` — one tmux SESSION per resumed session, never a window inside another. Run the
  reattach command if a cloud id is shown; on *"its environment has expired"* stop it at once with
  `tmux kill-window -t =rr-<id8>:` (it created an empty cloud session) and start the local resume
  the same way. On *no conversation found*, the directory was wrong: report and ask. Resume one
  session at a time, waiting for its session file before the next. The session NAME is the plain
  `rr-<id8>`; `=rr-<id8>:` is only ever a `-t` TARGET (§III.4 traps).
- THE SESSION RESUMES IN THE MODE IT HAD (gblock, 2026-09-11: "exactly the state it began in").
  Its last recorded mode is the `permissionMode` of its newest `user` record — the only local
  record of it: session files carry none (§II), and the 89 `permission-mode` records in 6 of 168
  transcripts are all `auto`, all at startup (round 2). The reattach needs no flag: a
  server-respawned session kept its own `auto` under a `--permission-mode default` on its command
  line (§II). The local resume passes `--permission-mode <MODE>`, measured to override the CLI's
  own default (a fresh `claude` here starts in `auto`, §II). BEFORE resuming a session whose MODE is
  `auto` or `bypassPermissions`, the skill says so in the choice it puts to the human — once
  resumed it acts without asking — and the human's choice stands.
- AFTER each: report reattached / resumed locally (same session id, new cloud session) / failed,
  the mode each resumed in, the app URL, empty cloud sessions to archive, that names must be reset,
  and that each reattach holds its own single-session "at capacity" entry in the app until that
  session ends.
- Traps (measured): `claude rc --help` does not exit; the trust prompt defaults to *No, exit*;
  tmux targets are `=<name>:` — exact session, with the colon — for `send-keys`, `capture-pane`
  and `kill-window`: `<name>:` PREFIX-MATCHES another session (round 3 measured `sc-identity:`
  hitting `sc-identity-x`) and `=<name>` without the colon fails; `new-session -s` takes the PLAIN
  name — `-s '=rr-x:'` creates a session literally named `=rr-x_` that no target matches (round 4,
  tmux 3.4); a fresh interactive `claude` here starts in `auto`; project folders begin with `-`.

### III.5 Carriers

`agents` help and goldens (`agents.golden`, and `agents-rebuilt-warning.golden`, which also runs
plain `agents`); `root.go` `Env` (the injectable boot); `help-sql.golden` (new `v_session` column);
`schema.go` `ViewColumns` (`:182`, the declared view contract pinned by
`TestViewColumnsAreTheContract` in `internal/catalogue/open_test.go`, cited at `telecli/sql.go:17-18`
— missed by revision 10);
`help-session.golden` and `session.go:33-34` (closed_at meaning); `schema.go` comments (`:48-50`
closed_at, the hint column); both READMEs; the telepathy skill's verb notes and its "an upgrade can
rebuild the store EMPTY" paragraph, which this plan's own upgrade exercises; `hooks/README.md`
(the hooks rationale, moved out of `hooks.json` by #915) and `requirements.json`'s `gray-area-capture` purpose (`:34`); `main.go` `-event` help;
`plans/gray-area.md` §4 (`:166`, appended: restart recovery observes across sessions —
gray-area's side of §4's line — and adds one metadata column; checkpointing stays in
prosthetic-conscience); `plans/README.md`; `schema.go:8` (its comment still says `schema_test.go`
pins the view columns — corrected to `TestViewColumnsAreTheContract` in `open_test.go`).
Unchanged, checked: `open.go`'s `sessionCore` and the
`schema.go:30-35` recognition comment (the new column is not core); `find` and `v_word` (no speaker
text in `agents`). `frank-exchange-of-views/.../hookcmd.go:24` is already stale and not changed.
No `plugin.json` bump.

### III.6 Consumer census (run 2026-09-11 on 5915be84; raw output `~/.claude/scratch/restart-probe/rr-census.txt`)

Each contract this plan changes, the command that finds its consumers, and every hit assessed.
Paths are relative to the repository root.

**`v_session` gains `bridge_session_id`** — `grep -rn 'v_session' plugins README.md` (26 hits):
- `schema.go:138,144,145,170,182` — the view and its declared contract: MODIFY (§III tree).
- `help-sql.golden:7` — the generated column list: regenerated.
- `plugins/gray-area/README.md:55`, `SKILL.md:71`, `root.go:75`, `help-root.golden:19` — list the
  view NAMES only: unchanged.
- `plugins/gray-area/README.md:127` — explains `first_act`/`last_act` vs `ingested_*`: unchanged,
  and consistent with §I's transcript-time rule.
- `session.go:19`, `help-session.golden:4` — "query v_session for the ones that have ended": see
  the `agents` census below.
- `golden_test.go:251,256` (name their columns), `:479,480` (`count(*)`); `open_test.go:114,142,
  510,547,809,810,940`; `query.go:104` (`project_dir, closed_at`) — select named columns or a
  count: unchanged by an added column.
- `testdata/shape1.sql:89` — the shape-1 fixture: unchanged (it must stay shape 1).

**`agents` gains `lost` rows and `--lost`/`--window`/`--boot`** — `grep -rn -E 'telepathy
agents|newAgentsCmd|catalogue\.Agents\(' plugins README.md` (12 hits):
- `agents.go:12,29`, `help-agents.golden:9` — MODIFY / regenerated.
- `root.go:131` (registration), `cmd/telepathy/main.go:6` (a comment on the binary's name):
  unchanged.
- `session.go:18-19`, `help-session.golden:3-4` — "Run 'telepathy agents' for the ones running
  now, or query v_session for the ones that have ended": MODIFY to add "and, after a restart, the
  ones it cut off (`telepathy agents --lost`)" — missed by revisions 10–12.
- `root.go:93` example, `help-root.golden:25` — MODIFY: add `telepathy agents --lost` as the
  second example line (the recovery entry point).
- `plugins/gray-area/README.md:51`, `SKILL.md:28` — the verb one-liners: MODIFY to "who is
  running, in which worktree, doing what — and after a restart, what it cut off".
- `README.md:117` (root) — MODIFY the row's question to add "and which did the last restart cut
  off?".

**`sweep(stderr)` → `sweep(in, stderr)`** — `grep -rn -E '\bsweep\(' plugins/gray-area/tools`
(2 hits): `catalogue.go:83` (definition), `main.go:535` (the only caller) — both MODIFY.

**`SessionFile` gains `BridgeSessionID`** — `grep -rn 'SessionFile' plugins/gray-area/tools`
(16 hits): `liveness.go:21-65` (the struct and its readers) MODIFY; `query.go:24` and
`cmd/gray-area-capture/catalogue.go:100` read the files and ignore the new field: unchanged;
`sweep_test.go:146,207` and `liveness_test.go:15-56` build files or literals without it: an added
optional field leaves them valid. The capture's `bridgeID` is read from `<ppid>.json` directly
(§III.1), not through `ReadSessionFiles`.

**`Env` gains `Boot`** — `grep -rn -E 'Env\{' plugins/gray-area/tools` (4 hits): `root.go:46`
(the default constructor) sets it to the `/proc/stat` reader; `golden_test.go:45` (the golden
harness) injects a fixed boot; `root_test.go:51,187` leave it nil — unmeasurable, called only after
`openRead` succeeds (§III.3), so both keep their behaviour.

**`IngestFile` now clears `closed_at` on new bytes and writes `session.cwd`** —
`grep -rn 'IngestFile(' plugins/gray-area/tools` (run 2026-09-11 on 5915be84; raw output
`~/.claude/scratch/restart-probe/ingestfile-census.txt`; 21 lines = the definition
`internal/catalogue/ingest.go:106` + 20 call sites):
- Production: `cmd/gray-area-capture/catalogue.go:71` (Stop/SessionEnd ingest),
  `internal/telecli/backfill.go:48`, `internal/catalogue/sweep.go:92` (closure ingests, then closes
  at `:99`, so a session closure closes stays closed after its own read) — all keep working; the
  reopen is what §I criterion 1's "not closed after that activity" relies on.
- Tests, `internal/catalogue/ingest_test.go:96,105,127,135,163,175,198,215,234,282,326,331` — offsets,
  rotation, a vanished file and projection; none asserts `closed_at` (`grep -n closed_at` on the
  file: 0 hits), so the reopen cannot change what they check.
- Tests, `internal/catalogue/sweep_test.go:35,69,88,143,245` — the only `closed_at` assertions in
  either test file are `:57-59` and `:280`. `TestClosureIngestsBeforeMarkingShut` (`:26`) appends a
  turn and runs closure: closure's own ingest reopens, then closure closes, so `closed_at` is set
  afterwards, as asserted. `TestCapBoundsAnInvocationAndTheBacklogDrains` (`:236`) ingests every
  file BEFORE closure, so closure's reads find no new bytes, nothing reopens, and the queue drains
  to zero open sessions, as asserted. Both hold under the new rule; §V.1 adds the case that
  exercises it (a closed session given new bytes reopens).

**`Projection` gains `CWDs`** — `grep -rn -E '\bProjection\b' plugins/gray-area/tools` (5 hits):
`project.go:90-149` MODIFY; `project_test.go:28` (`outcomes`) reads other fields: unchanged.

## IV. Risk & Mitigation

| # | Risk | L × I × cost | Mitigation |
|---|---|---|---|
| R1 | False positives | medium × low × none | Accepted; human selection |
| R2 | A needed session is not listed | low × medium × low | `--window`; the skill says to widen it before concluding a session is gone |
| R3 | Wrong `RESUME FROM` | low × low × none | `(unverified)` label; the resume attempt fails loudly; the skill asks |
| R4 | A resumed agent acts without asking | medium × high × low | ACCEPTED by gblock's ruling (2026-09-11): a session resumes in the mode it had. Every row shows `MODE`, and the skill names each `auto`/`bypassPermissions` session in the choice it puts to the human before anything is resumed (§III.4) |
| R9 | A session the running daemon would bring back lazily is listed `lost` | likely within ~4 h × low × none | The skill's FIRST step: open those sessions in the app, then re-list (§III.4) |
| R10 | The CLI drops a mode spelling (`default` → `manual` is already half done) | low × medium × low | A recorded `default` is written `manual`, the spelling `claude --help` lists; both spellings are measured to work today (§II); a rejected flag fails loudly at start, and the skill reports it |
| R5 | Expired reattach leaves an empty cloud session | certain on expiry × low × low | Stopped at once; reported for archiving |
| R6 | Concurrent resumes re-close each other | low × low × low | One at a time, waiting for each session file |
| R7 | Never reaches this box without a release | certain until release × high × none | Stated; post-release check §V.5 |
| R8 | The shape bump rebuilds the store empty | certain at upgrade × medium × low | #894's rebuild markers make every read verb warn until `backfill`; the skill backfills first; cloud ids accrue again from the next hook |

## V. Verification Plan

1. **Unit and race** — any change under `plugins/gray-area/tools`: `go -C plugins/gray-area/tools
   vet ./...`; `strace -f -e trace=open,openat -o ~/.claude/scratch/restart-probe/rr-trace.txt go -C
   plugins/gray-area/tools test -count=1 ./...` then `grep -c
   "${XDG_STATE_HOME:-$HOME/.local/state}/special-circumstances"` on the trace MUST print 0;
   `CGO_ENABLED=1 … test -count=1 -race ./...`. Cases: registration (insert, reopen, bridge id
   kept, session-id mismatch refused, own id counted live by the same sweep); ingest (cwd chosen by
   key, fallback to first, reopen on new bytes, no reopen without; TWO PASSES: a first pass with a
   matching cwd, a second pass of new bytes whose only cwd does not match → the stored cwd stays the
   matching one); selection (pre-boot + unclosed listed; closed after activity and before boot
   excluded; reopened listed; window honoured; a PRE-BOOT TAIL READ AFTER THE BOOT is still listed —
   its `ingested_last` is past the boot and its transcript time is not; a store rebuilt 4 → 5 and
   then backfilled lists a SUPERSET of the sessions it listed before — the fixture holds one
   session closed before the boot, which the rebuilt store lists and the original does not (the
   rebuild loses `closed_at`, Accepted costs); `live` excluded; an advertised `ended` session
   shown once, as `lost`; `unknown` never `lost`; a registered session with no rows uses
   `ingested_first` and says so); a closed session given new bytes reopens (`closed_at` NULL);
   command rendering, as EXACT strings — with cloud id `cse_X` and MODE `auto`:
   `claude remote-control --session-id cse_X` and `claude --resume <id> --permission-mode auto
   --remote-control`; a recorded `default` renders `--permission-mode manual`; no cloud id → no
   reattach line; MODE `unknown` → the local command without `--permission-mode`, and the row
   reading exactly `MODE unknown — the resume starts in whatever mode the CLI chooses`; the MODE is
   the newest `user` record THAT CARRIES `permissionMode` (a session that went `auto` → `plan` shows
   `plan`; a session whose newest `user` record is a tool result with no mode shows the earlier
   prompt's mode, not `unknown`); the second `--lost` line names the newest pre-boot activity, its
   session and the window verdict (`inside` for an injected 1h, `past` for 7h), and reads exactly
   `newest pre-boot activity: none in the store — run telepathy backfill, then ask again` when
   nothing precedes the boot; a shape-4 store
   is recognised `Older` and rebuilt to 5 with the new column; `agents` with an injected
   unmeasurable boot prints the one line and the advertised rows, exit 0, while `agents --lost`
   refuses, exit 1; a nil `Boot` behaves as unmeasurable; every advertised session `ended` → the
   "none is live" sentence, not an empty table (golden `agents-all-ended`). The `/proc/stat`
   reader (`boot_linux.go`), on fixture text: a `btime` line → that value; no `btime` line, a
   non-numeric one, or an unreadable file → not measurable.
2. **Goldens** — any `telecli` change: `go -C plugins/gray-area/tools test -count=1 ./internal/telecli
   -run 'TestGolden|TestEveryVerbIsDocumented'`, then `go -C scripts run ./golden` (uncached, as CI).
3. **Surfaces** — `go -C scripts run ./pluginparity`, `./frontmatter`, `./archaeology -base
   origin/main`, `./rulesweep -base origin/main` (skill commits: `Rule-Class:
   policy-without-mechanism`, `Sibling-Sweep:` the telepathy skill and both READMEs).
4. **Real data** — build the binaries into `~/.claude/scratch/restart-probe/bin`.
   **Local probe** (Remote Control off; nothing leaves the box): in a scratch git repo at
   `~/.claude/scratch/restart-probe/rr-local/`, start `tmux new-session -d -s rrlocal -c <that dir>
   'claude --settings <no-rc.json> --permission-mode plan'` — `plan`, so its state differs from
   this box's default `auto` — where `<no-rc.json>` carries `"remoteControlAtStartup": false` and
   hooks that write `$PPID` to a scratch file and then `exec` the dev capture binary with
   `XDG_STATE_HOME` in scratch. Accept the trust prompt first — it defaults to *No, exit*: send
   Down, then Enter, to `-t =rrlocal:`. Then type, with `tmux send-keys -t =rrlocal:`, one prompt
   that uses exactly one read-only tool, *"read README.md with the Read tool, then reply with the
   single word OK"*, so that the newest `user` record afterwards is a TOOL RESULT carrying no mode
   (plan mode allows the read). The prompt's `user` record MUST read `plan`, and while it
   runs `~/.claude/sessions/<recorded ppid>.json` MUST exist with the payload's `sessionId` (the
   `bridgeID` source is the session's own file, not a guess). `kill -TERM $(cat <recorded ppid
   file>)`; `~/.claude/scratch/restart-probe/bin/telepathy --store <scratch store> agents --lost
   --boot <now>` MUST list it with `MODE plan` — from the prompt's record, although the newest
   `user` record is the tool result — and `claude --resume <id> --permission-mode plan
   --remote-control`. First, a DISCRIMINATING run: `tmux new-session -d -s rrflagless -c <RESUME
   FROM> 'claude --settings <no-rc.json> --resume <id>'` — NO mode flag — one tool-free prompt
   through `-t =rrflagless:`, its `user` record's mode noted as F, then `/exit`. If F is `plan`, a
   flagless resume already keeps the mode and the final check below cannot tell the flag from its
   absence; the PR says so. Then run the printed command WITHOUT `--remote-control` and WITH the same `--settings`
   file (`~/.claude/settings.json:20` would otherwise open a cloud session) as `tmux new-session -d
   -s rr-<id8> -c <RESUME FROM> '…'`; one more tool-free prompt through `-t =rr-<id8>:` MUST produce
   a `user` record reading `plan` — the state it began in. Then the same binary with `--store <scratch store> agents --lost` and NO `--boot` MUST
   report a boot equal to the `btime` line of `/proc/stat` (the default reader, on real data).
   Afterwards `tmux kill-window -t =rr-<id8>:` (and `=rrlocal:` if still up), and the scratch repo
   and its transcripts are removed.
   **The consented Remote-Control-on run (gblock, 2026-09-11; exactly one) — a SERVER-SPAWNED
   session, because that is what the reattach command brings back.** Server mode refuses global
   flags such as `--settings`, so: a scratch git repo at `~/.claude/scratch/restart-probe/rr-server/`
   whose COMMITTED `.claude/settings.json` carries the same capture hooks, so every worktree the
   server creates carries them; its trust accepted once with an interactive `claude --settings
   <no-rc.json>` (Remote Control off; the trust prompt defaults to *No, exit*); then `tmux
   new-session -d -s rrprobe -c <repo> 'claude remote-control --spawn=worktree --capacity=4
   --remote-control-session-name-prefix=rrprobe'`. gblock creates ONE New session in that server's
   entry in the app and sends it *"reply with the single word OK; use no tools"*, denying any
   permission request; its `user` record's mode is noted as M, whatever the app chose. Asserted: the
   recorded `$PPID` names `~/.claude/sessions/<ppid>.json` holding the payload's `sessionId`. Then
   SIGTERM the server and do NOT restart it; `~/.claude/scratch/restart-probe/bin/telepathy --store
   <scratch store> agents --lost --boot <now>` MUST list the session with `CLOUD cse_X`, where
   `session_X` is the `bridgeSessionId` its session file held, and `MODE M`. Its printed reattach
   command (no mode flag) is run from its `RESUME FROM` as `tmux new-session -d -s rrreattach -c
   <RESUME FROM> '…'`: it MUST start and bring the session back in the app; gblock sends *"reply
   with the single word OK again; use no tools"* from the app (server mode has no local input),
   denying any permission request; the session's own transcript (`<sessionId>.jsonl` in its
   worktree's project folder) MUST gain a `user` record reading M — the state it began in.
   Afterwards `tmux kill-window -t =rrreattach:` (and `=rrprobe:` if still up), the scratch repo and
   its transcripts are removed, and the cloud sessions are reported for gblock to archive.
5. **After release, on this box:** at the next reboot, follow the skill; record the outcome here.
6. **Auditor gate:** `/prosthetic-conscience:plan-audit plans/restart-recovery.md` → PASS (its
   prerequisite, `plans/catalogue-shape-rebuild.md`, merged as #894).
