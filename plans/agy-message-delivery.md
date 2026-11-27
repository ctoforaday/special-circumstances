# Messages reach an Antigravity conversation: route selection and queue visibility

> STATUS: implemented 2026-11-27 in #1951. Tracks #1923. Plan 1 of 2, by the human's ruling of 2026-11-26 ("Split: A+C now, wake plan next"). Plan 2 (`plans/agy-sidecar.md`) replaces this plan's server registry with one gray-area sidecar that delivers every message and notices queued mail at its start. This plan gates the next gray-area release.

## I. Summary & Goals

`telepathy send` to an Antigravity (agy) conversation failed twice on 2026-11-26.

1. **Delivery queued for 13 hours.**
   - `agy-nightly` restarted the agy remote-control daemon at 03:00:07.
   - A send at 03:3x to the idle conversation 31aafaf1 went to its on-disk mailbox, which is read only at the conversation's next turn.
   - The sender was told "queued", but not why, nor for how long.

   Plan 2 owns the wake. This plan owns telling the sender the truth.
2. **A false "second writer" refusal.**
   - Once the human opened the conversation, the current daemon (pid 284285) held its database.
   - Send still refused, with "process pid 284285 already has conversations/<id>.db open — sending live would create a second writer".
   - Pid 284285 was the daemon itself.

The same queue happened again on 11-26 at 18:4x, after a deliberate daemon restart: "no current daemon server is recorded".

Goals:

- **G1 (bug fix).**
  - Send never names, as a second writer, a holder that is the current daemon or the process of any Live server record.
  - A dead server record never shadows a live route.
  - The registry drops records whose process is gone.
- **G2 (sender visibility).** The sender sees:
  - which route delivered a message;
  - or, when it queued, the cause and the two times that bound the wait (the daemon's start, and the conversation's last recorded step);
  - every conversation with a message waiting, independent of whether anything holds it;
  - which of its own messages are still unread, by the client's own record, whether or not the recipient ever runs telepathy.

Value case: the human ruled the scope ("Everything") on 2026-11-26. G1 is a correctness bug. G2 replaces a bare "queued" that hid a 13-hour wait.

| Goal | Steps | Proven by (§V) |
|---|---|---|
| G1 | A1, A2, A3, A4, A5, A6 | V1, V2, V4, V5, V6 |
| G2 | C1, C2 | V1, V2, V3, V6 |

Out of scope:
- the wake itself (plan 2);
- reading the daemon's token from its process, environment or files (#1741);
- live delivery off Linux (#1905).

## II. Technical Context

The census is on origin/main 7d16b9e70. Paths are relative to `plugins/gray-area/tools/`. KNOWN means measured or read in code.

**Routes.** `internal/telecli/send.go` `deliver` (215–307) tries three routes in order:
- **Route 1, the conversation's own record.**
  - `peer.LookupConversationServer` (`internal/peer/agy.go:252`) picks the record with the newest `registered_at` that lists the conversation.
  - It is gated by `Env.ConversationBusy(root, id, own.PID)` (`internal/telecli/root.go:61`; `peer.ConversationBusy`, `agy_lock_unix.go:26`). That function returns `(bool, string)`, and the string is prose.
- **Route 2, the current daemon's record.**
  - `Env.CurrentDaemon()` (`root.go:60`; `agy_proc_linux.go:49`, systemd MainPID) supplies the key.
  - It is gated by `ConversationBusy` excluding the daemon PID.
  - Any server record's token delivers to any conversation that server hosts or adopts (KNOWN: `cmd/telepathy/live_agy_test.go:458`).
- **Route 3, the mailbox.**
  - The message is written to `brain/<id>/.system_generated/messages/<uuid>.json` (fields `id, recipient, sender, priority, timestamp, renderDetails, content`), with an empty `messages/undelivered/<uuid>` marker.
  - The client reads the message at the conversation's next turn, sets `read.json[<uuid>] = true` (an id → `true` object with no time; KNOWN on the box 11-26), and removes the marker.
- **Today's queue causes** (`send.go` 222–305):
  - no runtime directory;
  - own-record busy;
  - conversation root not the daemon root;
  - daemon lookup error;
  - daemon busy;
  - an HTTP error from a record's server;
  - an own-record lookup error;
  - no daemon record since the restart.

**Records** live in `$XDG_RUNTIME_DIR/gray-area/antigravity-servers/`, under `lockRegistry` (a flock on `.register.lock`).
- `peer.RegisterWith` (`agy.go:120-175`) writes them. It is called from `PersistentPreRun` (`root.go:162-170`, every verb except `link`) and from `telepathy link` (`link.go:44`).
- An already-listed conversation returns `AlreadyLinked` with nothing written (`agy.go:146-150`).
- `registered_at` is the first-record time.

**Liveness, as it exists.**
- `catalogue.Liveness` (`internal/catalogue/liveness.go:13`) has the values `Live`, `Ended` and `Unknown`.
- `catalogue.Of` (`:72`) returns `Unknown` for a pid from a foreign pid domain (`linux:<machine-id>:<pid ns>`, `LocalPidDomain()` in `liveness_linux.go:11`), and `Ended` only for a local pid that is gone or reused.
- `peer` does not import `catalogue` today. `catalogue` imports only `shared/trajectory`, so `peer` may import it without a cycle.

**The hook.** `cmd/gray-area-capture/link.go:34-65` asks the agent to run `telepathy link` on a turn's first invocation, when neither the host server's record nor the conversation's record answers a TCP dial.
- It uses the stale wording ("…points at a language server that is no longer running…") only when a record for the conversation still exists.
- Its comment at :57-59 records a measurement: a conversation whose history showed `linked` ignored the plain request.

**Why failure 2 happened** (KNOWN, from the census and the registry):
- At about 16:05, route 1's newest record for 31aafaf1 was the previous daemon's (pid 266530, gone).
- Excluding 266530, `ConversationBusy` found 284285 holding the database, and route 1 refused.
- `send.go:246` skipped route 2 because route 1 had refused as busy.
- 284285 had no record until 16:07:22. So the correct outcome at 16:05 was QUEUED, with a cause naming the unlinked daemon. From 16:07:22, the correct outcome is delivery by route 2.

**Queued mail on the box** (KNOWN 11-26, from the plan audit):
- 10 conversations have `undelivered` markers, the oldest from 11-03.
- `telepathy agents` shows none of them: it lists agy conversations only while a process holds their `.db` (`catalogue.ActiveAntigravitySessions`).

**Install skew.** telepathy ships in two installs (Claude Code's plugin cache, and agy's `~/.gemini/config/plugins/gray-area/bin/`). They update independently and share one registry.

## III. Proposed Changes

```
plugins/gray-area/tools/
├── internal/peer/
│   ├── agy.go               [MODIFY] record shape (A3), RecordLiveness (A1), Prune (A4), mailbox scan (C2)
│   ├── peer.go              [MODIFY] package doc states A2's routing model
│   ├── agy_lock_unix.go     [MODIFY] ConversationBusy → Holder from /proc/locks, fail-closed rules (A2)
│   ├── agy_lock_other.go    [MODIFY] same signature
│   ├── agy_lock_windows.go  [MODIFY] same signature
│   ├── agy_db_unix_other.go [MODIFY] otherDBHolder/flockHeldOnlyBy return the holder pid (unix, !linux)
│   ├── agy_proc_linux.go    [MODIFY] readProcStat splits "no such process" from "unreadable"; ticks → time
│   ├── agy_proc_other.go    [MODIFY] the same split
│   └── *_test.go            [MODIFY] A1–A4 and C2 unit tests
├── internal/catalogue/liveness.go [MODIFY] Of's pid-domain rule exported for server records (no new type)
├── internal/telecli/
│   ├── root.go              [MODIFY] Env.ConversationBusy signature; Env.RecordLiveness seam; PersistentPreRun prunes
│   ├── send.go              [MODIFY] route selection (A2), cause wording (C1)
│   ├── link.go              [MODIFY] linked_at refresh, prune, help text
│   ├── agents.go            [MODIFY] "waiting for Antigravity" section (C2)
│   ├── send_test.go         [MODIFY] incident and holder fixtures, per-cause queued text
│   ├── agents_test.go       [MODIFY] C2 fixtures
│   └── testdata/golden/     [MODIFY] help-link, help-send, help-agents, help-root, agents.golden;
│                            [NEW] send-queued-*.golden (one per cause), agents-waiting.golden
├── cmd/gray-area-capture/
│   ├── link.go              [MODIFY] liveness via RecordLiveness; two openings chosen by the linked marker; no ask without a pid domain (A5)
│   └── link_test.go         [MODIFY] pid-keyed fixtures; the A5 pruned-record case
├── cmd/telepathy/main_test.go      [MODIFY] TestTheBuiltBinary env; TestBuiltBinaryNeverSeesHostRuntimeDir (A6)
└── cmd/telepathy/live_agy_test.go [MODIFY] wording at :178, :211, :464; ConversationBusy calls at :401, :406, :420;
                                      ProcessStartTime at :366; the presence-lock 'second writer' expectation at :441
                                      stays a 'second writer' cause (the holder is the test process)
plugins/gray-area/skills/telepathy/SKILL.md, plugins/gray-area/hooks/README.md,
scripts/universe.sh [MODIFY] exports XDG_RUNTIME_DIR under the universe (A6)
plugins/gray-area/README.md, README.md, docs/local-development.md   [MODIFY] routing wording;
SKILL.md :29, :45, :193 also point a sender at `telepathy agents`' waiting section
```

### A. Route selection and record hygiene `[MODIFY]`

- **A1. Server-record liveness uses the catalogue's liveness concept.**
  - Records gain `pid_domain` (`catalogue.LocalPidDomain()` at write).
  - `peer.RecordLiveness(s Server) catalogue.Liveness` returns the catalogue's `Live | Ended | Unknown` and uses its pid-domain guard. It differs from `catalogue.Of` in two places:
    - an unreadable `/proc` gives `Unknown`, not `Ended`;
    - an empty domain gives `Unknown`. Stated residue: on a host without `/etc/machine-id` (common in containers) every record is `Unknown`, and live delivery to agy there queues with the cause said.

    Session liveness through `Of` is unchanged. Its rule:
    - a foreign or empty `pid_domain` → `Unknown`;
    - no pid (`addr-` key) → `Unknown`;
    - `/proc/<pid>` unreadable for any reason other than "no such process" → `Unknown`;
    - no such process, or a start time different from the record's → `Ended`;
    - otherwise → `Live`.
  - `peer.readProcStat` (`agy_proc_linux.go`) returns an error that separates "no such process" (`fs.ErrNotExist` on `/proc/<pid>/stat`) from other read failures, so `Ended` is a positive finding. `ProcessStartTime` keeps its signature and is built on it. `catalogue.ProcStart` is unchanged.
  - It is injected as `Env.RecordLiveness`, beside `Env.CurrentDaemon`, and passed to `Prune`.
  - Route selection, prune and the hook all call it.
  - `Unknown` records are never a delivery target: send has no dial today and gains none. A conversation whose only record is `Unknown` queues, and the cause says so (stated residue).
- **A2. Route selection: one holder discovery, then one target.**
  - `ConversationBusy` returns `Holder{PID int; Kind HolderKind; Detail string}`, where `Kind` is `None | PresenceLock | DBOpen | Uninspectable`.
    - Callers branch on `PID` and `Kind`; `Detail` is only printed.
    - `Env.ConversationBusy` takes the new signature, and every stub follows.
    - For `PresenceLock`, `PID` is the lock holder's pid, read from `/proc/locks`; `flockHeldOnlyBy` already parses that file and drops the pid.
    - Only `Uninspectable` has `PID == 0`, and no output ever prints "pid 0".
    - Holders are read from `/proc/locks` and the database's open files; the probe never takes agy's presence lock.
    - The probe fails closed: it returns `Uninspectable` when `/proc/locks` names a locker that has exited, and when it names no holder while the lock is on a network filesystem (NFS, SMB, Ceph, AFS, 9P, GFS2, OCFS2, Lustre, Coda, GPFS). FUSE is a stated blind spot. Stated limit: the probe sees only the processes of telepathy's own pid namespace, so a holder outside it is invisible, and a container that shares Antigravity's data with the host, or a lock taken on another host through a FUSE client of a distributed filesystem, can create a second writer (the human's ruling of 2026-11-27, "Drop it, state limit").
  - **Discovery.** Send probes once, excluding nothing: `H := ConversationBusy(root, id, 0)`.
  - **When `H.Kind == None`** (no process holds the conversation), the targets are tried in order:
    1. the `Live` record listing the conversation with the greatest `linked_at` (A3);
    2. the current daemon's record (route 2), which adopts the conversation.

    **Every SendAgentMessage is immediately preceded by `ConversationBusy(root, id, target.PID)` returning `None`.**
    - Send moves to target 2 only after a dial failure on target 1: the request never arrived. It re-probes first. The client bounds the dial at 2 s (a dial that never completes is a dial failure), uses no proxy, and follows no redirect.
    - If target 1 now holds the conversation, send queues, and the daemon is never called.
    - Two kinds of non-200 end the send without UNKNOWN; neither tries another server:
      - a refusal before any handling, in the shapes measured on agy 1.5.18 (`TestLiveAntigravitySendAgentMessageRefusals`): a 401 whose body is the CSRF interceptor's Connect error (code `unauthenticated`, message starting `unauthenticated: missing CSRF token` or `unauthenticated: invalid CSRF token`), or a 404 or 405 whose body is not a Connect error. Any other shape of those statuses is UNKNOWN, so a reworded or handler-level error never queues a message that may have arrived. A refusal queues with "the language server recorded at <addr> refused the message before acting on it, HTTP <n>: <body>";
      - a 429 is asked again on the same server, up to 3 requests, waiting its Retry-After (1 s when absent, at most 5 s) and re-probing after each wait, then queues with "the language server recorded at <addr> answered 429 (rate limited) 3 times" (the human's ruling of 2026-11-27). Each retry opens a new connection, and a retry that cannot reach the server queues: once a target has answered, the send stays with it.
    - Every other failure from a target (a timeout, a reset, a redirect, any other status, 400 included) ends the send. No other server is tried and nothing is queued, because the message may have arrived; a second server would adopt the conversation target 1 is about to run. Send prints "delivery UNKNOWN: <cause> — the message may have arrived, so no other server was tried and nothing was queued; check the conversation before sending it again" and exits non-zero (the human's ruling of 2026-11-26, "Report unknown"). A 400 stays UNKNOWN because agy answers it both for an undecodable body and for a method's own validation.
    - The cause printed is the one that stopped the last target tried.
  - **When `H.PID` belongs to a record X whose `RecordLiveness` is `Live`** (any record, daemon or not):
    - The sole-holder gate runs `ConversationBusy(root, id, H.PID)`.
    - If it returns `None`, send delivers through X. X holding its own presence lock is excluded by the gate, so it reads `None`.
    - If the gate returns `Uninspectable`, the message queues with the uninspectable cause.
    - If the gate returns a holder with pid N ≠ 0 and N ≠ `H.PID`, two distinct holders were observed, and the message queues: "held by the language server recorded at <X.address> and also by pid N".
    - When X cannot be reached (a dial failure), send queues with "the language server recorded at <X.address> cannot be reached". A refusal (401, 404, 405) or an exhausted 429 from X queues as above. Any other failure from X is UNKNOWN. None falls back to another target, because X holds the conversation.
  - **When `H.PID == CurrentDaemon().PID` and the daemon has no record**, the message queues: "the daemon (started <T>) has not been linked since it started".
  - **When `H.Kind == Uninspectable`**, the message queues: "cannot inspect whether another process holds the conversation: <detail>".
  - **Any other `H`** queues: "held by another process, pid N; sending live would create a second writer".
  - The gate only ever excludes the pid of a `Live` record, so a pid from a foreign namespace never excludes a local process.
  - `send.go:246`'s skip of route 2 is removed; the structure above replaces it.
  - **Where the route is recorded.** The existing `peer.Delivery.Route` (`peer.go:45-56`: `claude-socket`, `antigravity-server`, `antigravity-mailbox`) is extended. `antigravity-server` is replaced by three values:
    - `antigravity-daemon`, when the target is the current daemon's record;
    - `antigravity-server-own`, when the record lists the conversation;
    - `antigravity-server-holder`, otherwise.

    `help-send`, `SKILL.md:45` and the delivery line follow. The qualifier lives in the field, never composed into `Via`.
- **A3. The record carries `linked_at` per conversation.**
  - The shape becomes `{shape: 2, key, pid, start_time, pid_domain, address, token, daemon, registered_at, conversations: [{id, linked_at}]}`.
  - Every `RegisterWith` sets the caller's `linked_at` to now, even when the conversation is already listed. `AlreadyLinked` still names the state.
  - The reader accepts `shape == 2` only. Any other shape is not a record, and A4 deletes it.
- **A4. Prune: `peer.Prune(runtimeDir, liveness func(Server) catalogue.Liveness)`.**
  - It runs under `lockRegistry`, from `telepathy link` and from `PersistentPreRun` after registration. That covers every caller with a runtime directory whose registry directory exists, inside an agy shell or not.
  - When the registry directory does not exist, it does nothing: it neither creates the directory nor prints.
  - Every take of `.register.lock` (prune and `RegisterWith`) waits at most 2 s, so a wedged holder never blocks a verb. When the lock cannot be opened or taken in that time, it skips the prune and prints one line, "telepathy: server records not pruned: <error>", the way registration failures print today.
  - It reads only `*.json` files and deletes only:
    - a file that parses as a JSON object whose `shape` is not 2 (the 13 legacy files have none);
    - a shape-2 record whose liveness is `Ended`.
  - It keeps unparseable files, files it cannot stat, `Unknown` records, and every non-`.json` file. `.register.lock` is therefore never touched.
- **A5. The link request no longer depends on a kept record.**
  - Prune deletes the record the hook's stale wording used to read. `RegisterWith` therefore also writes an empty marker, `$XDG_RUNTIME_DIR/gray-area/antigravity-linked/<conversation-id>` (directory 0700), inside the same `lockRegistry` critical section, after the record's rename, so a write that failed leaves no marker. Prune never reads that directory.
  - The hook keeps main's two openings, chosen by the marker:
    - no marker: "Other agents cannot yet wake this conversation with `telepathy send`.";
    - marker present: "This conversation's earlier `telepathy link` points at a language server that is no longer running, so other agents cannot wake it."

    Both continue "Before your task, run this shell command once: … link — it prints `linked` and needs nothing further."
  - Stated residue: the marker is on tmpfs, so after a reboot a conversation that had linked is asked with the plain opening.
  - The hook never asks on a host with no pid domain (every record would read `Unknown`, so a link would not help). Nor does it ask when it finds no Antigravity process above it and the conversation's record names no server pid: asking would rewrite the same unusable record, in the stale wording. `link` says when the record it wrote cannot be used by send, and why (no pid, no pid domain, or the server's liveness).
  - `shared/hookcore` tests match only "run telepathy link" and are unaffected.
  - Stated residue, until plan 2:
    - the hook does not ask when the conversation's own record is `Live`, even if the host server is a different, unrecorded one;
    - for example, a standalone `agy` still running while the daemon hosts the conversation;
    - while that standalone runs, sends queue with "daemon not linked since it started".
    - The dial-based hook has the same hole today.
- **A6. No test, mutation run or dev harness runs a checkout-built verb against the host's registry.**
  - A4 makes every verb prune. KNOWN:
    - `scripts/check/gates.go:178` runs gray-area's `./...` with the host environment;
    - `cmd/telepathy/main_test.go` `TestTheBuiltBinary` runs its binary with no `cmd.Env`, and its `--store <absent> touched x.go` case reaches `PersistentPreRun` before failing;
    - no `internal/telecli` test executes a root built from `Defaults()`, because `Env.RuntimeDir` is injected.
  - `TestTheBuiltBinary` gives every exec of the built binary an environment with `XDG_RUNTIME_DIR` set to `t.TempDir()`.
  - A structural test, `TestBuiltBinaryNeverSeesHostRuntimeDir` in `cmd/telepathy`, reads every non-live `*_test.go` in the module that execs the built telepathy, and fails if any `exec.Command` of it lacks an `XDG_RUNTIME_DIR` entry in its `Env`. It has the shape of `TestLiveTestsUseTesthome`. `HOME` is not redirected, so live tests keep the host credentials that `testhome` lends.
  - `scripts/universe.sh` `universe_env` exports `XDG_RUNTIME_DIR="$WORKDIR/run"` (mode 0700) for **both** harnesses, before the agy-only branch. The host-token renewal subshell (`universe.sh:379`) sets the host's runtime directory explicitly, as `renewHostAgyToken` (`liveauth.go:294-312`) requires.
  - §V check 2 runs every mutation with `XDG_RUNTIME_DIR` set to a scratch registry seeded from a copy, so a mutation that removes A6 deletes nothing of the human's.
  - A host-listing comparison after the suite is not built: it would detect only after the files are gone, and it reddens on another session's legitimate prune. The structural test is the pin.
- **A future record shape is deleted by this binary** (stated residue): prune deletes any record whose `shape` is not 2, so a shape-3 record written by a later release is deleted by any shape-2 binary on every verb until both installs update. The release note says so.
- **Every record drops once, at upgrade** (stated residue).
  - No record on the box has a `shape` field (KNOWN, 42 of 42 on 11-26). The first verb the new binary runs therefore deletes every record, including those of running servers.
  - Until each conversation's next turn, when the hook asks it to link, sends to it queue with the cause stated.
  - The no-compatibility rule forbids keeping a reader for the old shape. Plan 2 (`plans/agy-sidecar.md`) deletes the registry.
  - The release note tells the human this happens once.
- **Install skew** (stated residue):
  - Until both installs update, the old binary cannot read shape-2 records and writes its own shape, and the new one prunes those.
  - Routes can flip, and a message may queue instead of delivering, with the cause said.
  - Both installs update from the same release, and `agy-nightly` refreshes agy's plugins nightly, so the on-disk installs converge within a day. A running session keeps the binary it loaded, though, so the skew lasts until the longest-running old session ends.
  - The release note says so.
- **"Nothing is written" carriers end.** A3 rewrites the record on every link, so every text saying a relink writes nothing changes:
  - `internal/telecli/link.go:54` (printed by `link`);
  - `internal/telecli/link.go:28` and `help-link.golden:13` (help);
  - `internal/peer/agy.go:90` (the `AlreadyLinked` doc);
  - `internal/peer/peer_test.go:124-126` (test comment).

  Enumerated by `git grep -nE 'nothing written|writes nothing|nothing was written|must not be rewritten' -- plugins/gray-area`, which after the change returns only `internal/catalogue/ingest.go:216` and `internal/catalogue/open.go:279`. Both mean something else and are not edited.
- **Tests** (each fails when its hunk is reverted, §V check 2):
  - `TestSendIncidentAt1605Queues`:
    - fixture: own record `Ended` (previous daemon), no current-daemon record, holder = current daemon;
    - expect QUEUED, with "not linked since it started" and no "second writer".
  - `TestSendIncidentAfterLinkUsesRoute2`: the same, plus a current-daemon record → `antigravity-server (daemon)`.
  - `TestSendHolderIsAnotherLiveRecord`:
    - the newest-`linked_at` record is S1 (Live);
    - the sole holder is S2's pid (Live; S2's record does not list the conversation);
    - expect delivery through S2 with `Route == HolderRecord`.
  - `TestSendSoleHolderWithNoListingRecord`:
    - no record lists the conversation (as after the upgrade drop);
    - the sole holder is S2 (Live, not the daemon);
    - expect delivery through S2;
    - the printed text never contains "also by".
  - `TestSendTwoHoldersQueues`:
    - S1 and S2 (both Live) both hold the conversation;
    - expect QUEUED, the cause naming both;
    - no line calls either pid a second writer;
    - removing the sole-holder gate turns this red.
  - `TestSendForeignDomainPidNeverExcluded`:
    - a foreign-`pid_domain` record's pid number equals the local DB holder's pid;
    - expect QUEUED, naming that pid as the holder.
  - `TestSendUnreachedCauseIsLastTarget`:
    - target 1 cannot be dialled;
    - the daemon has no record;
    - the printed cause is the daemon cause, not target 1's.
  - `TestSendUnrelatedHolderQueues`: the holder pid belongs to no record and is not the daemon; expect QUEUED, naming that pid.
  - `TestSendPresenceLockHolderNamed`: a presence-lock holder is named by its pid from `/proc/locks` (fixture).
  - `TestSendUninspectableQueues`: never prints "pid 0".
  - `TestSendReprobesBeforeDaemonFallback`:
    - the target-1 stub cannot be dialled, then reports itself holding the database;
    - expect QUEUED, with the daemon stub receiving zero requests.
  - `TestSendGateUninspectableNoAlso`: the gate returns `Uninspectable`; expect the uninspectable cause, and never "also by pid 0".
  - `TestSendHolderUnreachedQueues`: X holds the conversation and cannot be dialled; a daemon record exists; expect QUEUED, with the daemon not called.
  - `TestSendHolderErrorStatusIsUnknown`: X holds the conversation and answers 500; expect UNKNOWN, non-zero exit, no mailbox file, the daemon not called.
  - `TestSendTimeoutAfterSendIsUnknown`: target 1 accepts the request and never answers; expect exactly one request at target 1, none at the daemon, the daemon never probed, no mailbox file, UNKNOWN.
  - `TestSendErrorStatusIsUnknown`: target 1 answers 500; expect UNKNOWN (golden `send-unknown`).
  - `TestSendDialFailureFallsBack`: target 1 cannot be dialled; expect delivery through the daemon.
  - `TestSendAgyServerSaysWhetherTheMessageCanHaveArrived`: only a dial-stage error is unreached; a reset after the request (OpError `read`), a redirect, no answer and an error status are not.
  - `TestSendAgyServerDialTimeoutIsUnreached`, `TestSendAgyServerRefusedShapes` (4 confirming, 9 disconfirming shapes), `TestSendAgyServerRateLimited`, `TestRetryAfter`.
  - `TestSendRateLimitedThenUnreachedQueues`, `TestSendRateLimitRetryOnANewConnection`, `TestSendHolderRateLimitedQueues`, `TestSendDaemonRefusedQueues`, `TestSendFollowsNoRedirect`.
  - `TestLinkRequestPidlessGuardNeedsBothHalves`.
  - `TestSendDaemonErrorStatusIsUnknown`, `TestSendDaemonTimeoutAfterFallbackIsUnknown`.
  - `TestSendRateLimitedThenDelivers`, `TestSendRateLimitedQueues`, `TestSendRateLimitWaitIsCapped`, `TestSendRateLimitedReprobes`, `TestSendRefusedQueues`.
  - `TestFailedRegisterLeavesNoMarker`.
  - Live: `TestLiveAntigravitySendAgentMessageRefusals` — each refusal (no token, wrong token, unknown method, wrong HTTP method) leaves the target transcript unchanged, and the same request made valid answers 200 and arrives; `peer.IsRefusal` agrees with every measured row.
  - `TestConversationBusyFindsTheDBOpen` and `TestConversationBusyChecksTheDBPastTheExcludedLockHolder`: the database-open check, without stubs.
  - `TestLookupSkipsEndedNewerRecord`.
  - `TestLinkedAtRefreshesOnRelink`.
  - `TestRecordLiveness`, covering live; reused pid → `Ended`; missing pid → `Ended`; unreadable `/proc` (injected) → `Unknown`; foreign `pid_domain` → `Unknown`; `addr-` → `Unknown`.
  - `TestPruneDeletesOnlyEndedAndOldShape`: removes an `Ended` record and a legacy file; keeps a `Live` record, an `Unknown` record, an unparseable file and `.register.lock`.
  - `TestPruneNoRegistryDirIsSilent`.
  - `TestPruneUnderLockWithWaiter`: a `RegisterWith` blocked on the lock while prune runs completes afterwards, and its record is present.
  - `TestReaderRejectsOldShape`.
  - `link_test.go`: with a pid-keyed record pruned away, the hook still asks with the A5 wording.

### C. The sender sees what happened `[MODIFY]`

- **C1. Every QUEUED line states its cause and its bounds.**
  - Each cause in §II's list, and A2's rules 3–5, prints one fixed sentence. One golden per cause: `send-queued-<cause>.golden`.
  - Causes not yet in §II's list, each with its own golden:
    - "no daemon is running" (`send.go:275`);
    - "live delivery is Linux-only" (`send.go:281`);
    - "the daemon's language server cannot be reached" (golden `send-queued-daemon-unreached`).
  - Two time lines are added:
    - "daemon started <T>", or "no daemon is running", or "daemon start not measured" when `Boot` or `/proc` cannot be read or the host is not Linux: `ProcessStartTime(CurrentDaemon().PID)` in clock ticks, converted with the existing `Env.Boot` / `catalogue.BootTime` (`internal/catalogue/boot_linux.go:17`) and USER_HZ = 100.
      - USER_HZ is the fixed Linux userspace clock-tick rate.
      - It is stated as a constant because no cgo-free `sysconf` exists.
      - Goldens inject both values.
    - "last step recorded at <T>": the `created_at` of the last line of `brain/<id>/.system_generated/logs/transcript.jsonl`, read by seeking from the end. Every step line carries `created_at` (KNOWN). If the transcript is missing or empty, the line reads "no step recorded".
- **C2. `telepathy agents` gains a "waiting for Antigravity" section.**
  - It lists every agy conversation with at least one `messages/undelivered/*` marker, independent of whether any process holds it and of its age.
  - Population: `brain/*/.system_generated/messages/undelivered/`, a directory listing per conversation.
  - Each row shows:
    - the conversation;
    - the number waiting;
    - the oldest message's `timestamp`;
    - which of the waiting messages the caller sent, matched on the mailbox file's `sender` field and marked "yours".
      - The caller's id is its session id.
      - From a plain shell, whose sender is `telepathy`, messages are marked "from a shell", never "yours".
  - Messages with `sender: system` are agy's own notices, not mail between agents. They are excluded and counted on the row as "+N agy notices".
  - The section prints in the default `telepathy agents` output, ignores `--window`, and is not printed under `--lost`.
  - A message disappears from the section when the client reads it, which is the client's own record: the marker is removed and `read.json` has its id. The sender sees its message as unread for exactly as long as the client has not read it, whether or not the recipient runs telepathy.
  - No read time is shown, because the client records none. The plan does not depend on that: if agy adds one, C2 still holds and is merely less informative.
  - `read.json` is not needed to compute the section, so its absence (6 of the 10 queued conversations have none) or a parse failure cannot produce a wrong row.
  - **Every marker counts toward its conversation's row.**
    - A marker whose `<uuid>.json` is missing or does not parse shows as "N unreadable" on that row, and is never dropped.
    - A conversation whose only markers are `sender: system` gets a row like any other, showing only "+N agy notices".
    - A brain directory that cannot be listed prints "waiting messages not measured: <error>", never an empty section.
  - Goldens: `agents-waiting.golden` (new); `agents.golden`, `agents-all-ended.golden`, `agents-rebuilt-warning.golden` and `help-agents.golden` (updated).
  - Tests: `TestAgentsWaitingUnreadableMessage` (missing file, unparseable file) and `TestAgentsWaitingScanError`.
- **Tests:**
  - `TestSendQueuedCauseGoldens` (one per cause);
  - `TestQueuedLastStepTime` (3-line, 1-line, 0-line and missing transcript);
  - `TestTicksToTime` (fixture `btime` and ticks);
  - `TestAgentsWaitingListsIdleConversations`: a conversation with markers and no process holding it is listed; a conversation with no markers is not; "yours" is matched on `sender`.

### Census of carriers

Enumerated by `git grep -nE 'antigravity-servers|LookupConversationServer|CurrentDaemon|ConversationBusy|ProcessStartTime|RegisterWith|registered_at|AlreadyLinked|undelivered|second writer|its language server at|no longer running|nothing written|writes nothing|nothing was written|must not be rewritten' -- plugins/gray-area docs README.md`. Re-run at implementation time. Every hit is in §III's tree.

Added by hand: `internal/telecli/agents.go` (C2) and `internal/catalogue/liveness.go` (A1); neither matches the command.

## IV. Risk & Mitigation

| Risk | Likelihood | Impact | Cost to mitigate | Mitigation |
|---|---|---|---|---|
| A2 sends into a real second writer | low | high (corrupts a conversation) | low | The sole-holder gate: every delivery path probes excluding only the target's pid and delivers only on `Kind == None`. Two holders can never pass. Only a dial failure moves to another server; a 401/404/405 (measured) or an exhausted same-server 429 queues; any other later failure is UNKNOWN. `TestSendTwoHoldersQueues`, `TestSendUnrelatedHolderQueues` and `TestSendTimeoutAfterSendIsUnknown` pin it. |
| A4 deletes something it should not | low | medium (one queued message until the next link) | low | Only shape≠2 JSON objects and `Ended` records are deleted, and `Ended` needs positive evidence (A1). Waiter and no-directory tests pin it. |
| A1 misreads a pid from a sandbox's pid namespace as gone | low | medium | low | A foreign or empty `pid_domain` reads `Unknown`, which is never pruned (the catalogue's existing guard). |
| A5's marker is lost at reboot | certain after a reboot | low (the plain opening is still true and was measured to work) | — | Stated residue. |
| An UNKNOWN delivery is resent and arrives twice | low | low (one duplicate message) | — | The UNKNOWN line tells the sender to check the conversation first. |
| A local test, `check` or universe run prunes the host registry with a checkout build | high without A6 | high (every route on the box drops, repeatedly) | low | A6: `TestTheBuiltBinary`'s env; `TestBuiltBinaryNeverSeesHostRuntimeDir` (every exec shape); `internal/telecli`'s `TestMain`; `universe.sh`'s runtime directory. |
| Install skew flips routes after the release | medium | low (queued, cause said) | high (compatibility code is forbidden) | Stated residue; ends when both installs update. |
| C2's scan is slow | low | low | low | It lists one directory per conversation and reads only files that have markers. Timed in V6. |

## V. Verification Plan

Before any test: `unset $(env | grep -o '^FEOV_[A-Z_]*')`.

1. `go -C plugins/gray-area/tools test -count=1 ./internal/peer/ ./internal/catalogue/ ./internal/telecli/ ./cmd/gray-area-capture/ ./cmd/telepathy/` → `ok` for each, 0 `FAIL`.
   - Re-armed by any file in §III's tree.
2. Delete-the-fix, one hunk at a time, each run with check 1. Record the red test's name in the PR.
   - Hunks:
     - A1 liveness gate in route 1;
     - A1 `Ended` versus `Unknown` split;
     - A2 holder-record delivery;
     - A2 daemon fallback only after a dial failure;
     - A2 an UNKNOWN failure ends the send;
     - the database-open holder check;
     - the probe's fail-closed rules;
     - the 2 s lock bound;
     - A2 refusal keyed on status and body;
     - A2 429 same-server retry, each on a new connection, after the wait's re-probe; a retry's dial failure queues;
     - the dial bound and no-redirect client;
     - A2 Live-only exclusion;
     - A2 re-probe before each send;
     - A2 gate-result rendering;
     - the `pid_domain` write in `RegisterWith` (without it, every record reads `Unknown`);
     - `Route` set on each delivery;
     - each A2 queue cause;
     - A2 removal of the `send.go:246` skip;
     - A3 `linked_at` order;
     - A3 relink refresh;
     - A3 shape check;
     - A4 `Ended` deletion;
     - A4 `.json`-only rule;
     - A4 no-directory rule;
     - A6 `TestTheBuiltBinary` env (`TestBuiltBinaryNeverSeesHostRuntimeDir` goes red);
     - the call sites: Prune's call in `PersistentPreRun`, Prune's call in `link`, and the hook's call to `RecordLiveness` (delete each call and run the suite);
     - A2 sole-holder gate;
     - A5 marker and the two wordings;
     - C1 per-cause text;
     - C1 time lines;
     - C2 population;
     - C2 "yours".
   - A revert that stays green blocks merge.
3. `UPDATE_GOLDENS=1 go -C plugins/gray-area/tools test -count=1 ./internal/telecli/`, then read every changed line in:
   - `help-link`, `help-send`, `help-agents`, `help-root`;
   - `agents.golden`, `agents-waiting.golden`, `send-queued-*.golden`.
4. `go -C scripts run ./check` → "0 failed".
5. Live, through `shared/testhome` only, on a branch containing #1935: `HOOKCORE_LIVE=1 go -C plugins/gray-area/tools test -count=1 -run 'TestLiveAntigravityLinkAndSend|TestLiveAntigravityDaemonRestartWake|TestLiveAntigravityRealDaemonDiscovery|TestLiveAntigravitySendAgentMessageRefusals' -v ./cmd/telepathy` with `XDG_RUNTIME_DIR` on scratch → all PASS, and a listing of the host registry identical before and after.
   - Phase 3 names the host daemon's start time, because `CurrentDaemon` is the host's.
   - Re-armed by `hooks.json`, `cmd/gray-area-capture/link.go`, `internal/peer`, and telecli `send` and `link`.
6. On the box, with the built binary (`go -C plugins/gray-area/tools build -o ~/.claude/scratch/agy-delivery/telepathy ./cmd/telepathy`). Every step runs against a **copy** of the registry; the human's live registry is never pruned by an unreleased binary. Residue: the released binary is the first to prune the real registry, as stated in A's upgrade residue.
   - **Setup.** `C=$XDG_RUNTIME_DIR/gray-area-v6/gray-area; mkdir -p $C; cp -a $XDG_RUNTIME_DIR/gray-area/antigravity-servers $C/`. The copy stays on tmpfs, under the same 0700 runtime directory.
   - **Prune.**
     - Run `XDG_RUNTIME_DIR=$XDG_RUNTIME_DIR/gray-area-v6 ~/.claude/scratch/agy-delivery/telepathy agents`, whose `PersistentPreRun` prunes the copy.
     - For each file deleted from the copy (`comm -23` of before and after listings), check that its `pid` is not a running process, or that the file has no `shape`.
     - Print keys and field names only, never values.
     - `.register.lock` is still present.
   - **Waiting section.** `time` the same command. It lists every conversation shown by `ls -d ~/.gemini/antigravity-cli/brain/*/.system_generated/messages/undelivered/* | cut -d/ -f7 | sort -u` in under 2 s.
   - **Route named.**
     - Create a scratch conversation with `agy -p`.
     - In a second `agy -p --conversation <scratch>` turn, run `~/.claude/scratch/agy-delivery/telepathy link` with `XDG_RUNTIME_DIR` set to the copy.
     - Then run `~/.claude/scratch/agy-delivery/telepathy send <scratch> -`, with `XDG_RUNTIME_DIR` set to the copy, while the `agy -p --conversation` turn is still running (it is told to wait 60 s after linking). Expect `Route == OwnRecord` on the delivery line.
     - After that one-shot server exits, run the copy's `agents` once more. The shape-2 record the link wrote is deleted as `Ended`, which exercises `Ended` deletion on real data. Each deleted file's pid is read from its `pid-<pid>-<start>` key, against a `ps` snapshot taken just before.
     - Afterwards, delete the scratch conversation's mailbox file and its marker, so no test message stays in the waiting section. The scratch conversation itself stays in the brain directory, titled by its prompt ("gray-area check 6 scratch").
   - Teardown: `rm -r $XDG_RUNTIME_DIR/gray-area-v6`.
   - No message is sent to any conversation the human uses.
7. Release: gray-area is bumped once this lands. The tag is the human's call. After it, `/plugin update` and agy's plugin refresh, then re-run check 6 on the installed build.
