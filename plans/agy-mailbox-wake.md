# Messages to Antigravity always queue; a wake message starts the turn that reads them

> STATUS: proposed 2026-11-27 (gblock: "Daemon test, then build"). Supersedes the routing half of `plans/agy-message-delivery.md` (server records, link, route choice). Every fact below is cited to **V** = `~/.claude/scratch/agy-verify/FINDINGS.md` or **L** = `~/.claude/scratch/agy-ls-access/FINDINGS.md`; nothing unmeasured is built on.

## I. Summary & Goals

`telepathy send` to an Antigravity (agy) conversation delivers content live only through a server record a conversation's own shell wrote (`telepathy link`). After a daemon restart no record exists, so sends queue unread for hours; and live content can reach a server that adopts a conversation another process holds (split-brain, V "Holder" 3/3).

Design: **content never travels live.** Send always writes the agy mailbox (`peer.EnqueueAgy`); then, when nothing but the daemon holds the conversation, it sends the daemon a fixed **wake message** ("You have queued messages.") so the conversation takes a turn now, and that turn reads the mailbox. The daemon's token is one the human provisions (`ANTIGRAVITY_SERVER_CSRF_TOKEN`, which agy honours, L).

Goals:
- **G1.** A send to an idle daemon conversation is read within one turn (V: V1 7/7, real-daemon round trip).
- **G2.** No send creates a second writer of a conversation another process is seen to hold (V "Holder"). Residues (§IV): a holder that has not yet opened the database (n=1), a holder in another pid namespace, and the check-then-send window.
- **G3.** A daemon restart wakes every conversation with mail waiting (`agy-wake`, run by systemd).
- **G4.** No test, and no universe run, reads the host token file or reaches the host daemon.

Value case: plan 1's registry, link verb, hook link request, target ordering and refusal/429/UNKNOWN classification exist because content travelled live and could duplicate or split. A wake message asks for nothing and is harmless to repeat or lose (K4, K5), so all of it goes.

Residues accepted (the human's brief: one design, one PR): a conversation on a non-daemon server (standalone `agy`, a TUI, the IDE root) and every host without the provisioned daemon get the mailbox only, read at that conversation's next turn; every daemon start wakes every daemon-root conversation with unread agent mail, whatever its age (7 today, mail from 11-03 to 11-13, by the audit's read-only count); each `agy-provision` rotates the token, so wakes fail (loudly, message still queued) until the human restarts the daemon.

## II. Technical Context

Go 1.27, module `plugins/gray-area/tools`, agy 1.5.18, systemd user unit `antigravity-cli-daemon.service`. KNOWN facts:

- **K1.** A server started with `ANTIGRAVITY_SERVER_CSRF_TOKEN=<t>` accepts `SendAgentMessage` with header `x-codeium-csrf-token: <t>` (200), refuses a wrong token (401) and no token (401); proven on the real daemon through the human's drop-in (L "Real daemon confirmation"; V P1).
- **K2.** The token reaches no shell, background task, hook or MCP server (V P2; L matrix). Only a file the human provisions can carry it to telepathy.
- **K3.** The daemon's ports are LISTEN rows of `/proc/<pid>/net/tcp{,6}` whose inodes are the daemon's own socket fds. Tokenless, the HTTP (Connect) port answers 401 Connect JSON `unauthenticated: missing CSRF token`; the HTTPS port answers 400 "Client sent an HTTP request to an HTTPS server" (L unauth). No log file is read.
- **K4.** A turn woken by a wake message (sender `telepathy-wake`, title `telepathy`, content "You have queued messages.") first injects every undelivered mailbox message as a `SYSTEM_MESSAGE`, clears its marker, adds it to `read.json`; the model answered the queued content 7/7 (one queued message per wake; S9(b) queues two). The mailbox is read at every turn start, also for a loaded conversation (V V1). Real daemon: conversation 3d159adc answered QMARK8S (V "Real daemon round trip").
- **K5.** A wake leaves its own mailbox entry (sender `telepathy-wake`) with a marker until a turn reads it; a server that adopts a conversation also leaves a `system` notice with a marker. A wake whose turn never runs leaves its own entry queued (V U4, 3/3).
- **K6.** `ConversationBusy(root, id, exclude)` sees a holder by presence lock or open database before a send; a server sent a message for a conversation another agy holds adopts it anyway (V "Holder" 3/3).

NOT MEASURED (stated, not built on): `MainPID` and `HOME` inside `ExecStartPost=` (the lead's real-daemon check settles it; `CurrentDaemon` falls back to a `/proc` scan); a TUI holder; a resumed server that has not yet opened the database (invisible to K6, n=1); a wake sent to a conversation mid-turn; whether a turn clears markers of unreadable messages; a wake sent while the daemon runs the conversation mid-turn (worst case equals queue-only: the next turn start reads the mailbox, K4).

## III. Proposed Changes

```
plugins/gray-area/tools/internal/peer/
  agy_token_unix.go, agy_token_other.go [NEW]  DaemonTokenVar, ReadDaemonToken, WriteDaemonToken
  agy_port_linux.go, agy_port_other.go  [NEW]  DaemonAddress(client, pid)
  agy.go     [MODIFY] WakeSender, SendWake; Waiting.Wakes; registry, records, refusal shapes go
  peer.go    [MODIFY] package doc: one route per harness; Route keeps claude-socket, antigravity-mailbox
  agy_proc_*.go, agy_lock_*.go [MODIFY] HostServer, isAgyProcess, renameRecord, procStartTime go
  record_test.go, registry_concurrency_test.go, registry_wait_test.go, agy_dial_linux_test.go [DELETE]
plugins/gray-area/tools/internal/telecli/
  agyprovision.go, agywake.go [NEW]; link.go [DELETE]; send.go, root.go, agents.go [MODIFY]
  send_route_test.go [DELETE]; send_test.go, send_queued_test.go, main_test.go [MODIFY]; goldens
plugins/gray-area/tools/cmd/gray-area-capture/ link.go, link_test.go [DELETE]; main.go [MODIFY]
plugins/gray-area/tools/cmd/telepathy/ live_agy_test.go, live_agy_status_test.go, live_agy_lock_*_test.go [DELETE]
  live_agy_wake_test.go [NEW]; main_test.go, runtimedir_test.go [MODIFY]
shared/hookcore/ asymmetry_table_test.go, events_test.go, live_events_test.go, live_asymmetry_test.go [MODIFY]
docs: plugins/gray-area/{README.md,hooks/README.md,skills/telepathy/SKILL.md}, README.md, AGENTS.md,
  docs/local-development.md, scripts/universe.sh, plans/README.md, plans/agy-message-delivery.md [MODIFY]
```

**S1. `send` to an Antigravity conversation.** `EnqueueAgy` first; a write failure is the send's error. Then the wake, each step ending it with a one-line reason printed under the QUEUED line, in this order: (1) Linux; (2) conversation root is the CLI daemon root; (3) `ReadDaemonToken(env.TokenFile)` — before any network contact, so an environment without the file never probes a daemon; (4) `CurrentDaemon()`; (5) `DaemonAddress(client, daemon.PID)`; (6) `ConversationBusy(root, id, daemon.PID)`: `None` → (7) `SendWake`; any holder → "held by pid N, whose next turn reads the mailbox; no wake sent" (K6); uninspectable → its detail. (6) runs last before the request to narrow the check-then-send window. The QUEUED report keeps plan 1's two bound lines (daemon start, last step) when not woken. Exit 0 whenever the message is queued. The token is never printed.

**S2. Token file** `<os.UserConfigDir()>/gray-area/antigravity-daemon.env` (`Env.TokenFile`): exactly one line `ANTIGRAVITY_SERVER_CSRF_TOKEN=<value>` (one trailing newline), value 1–256 bytes of `[A-Za-z0-9-]` checked byte by byte. Opened with `O_NOFOLLOW` and checked on the descriptor (`fstat`); refused, each with its own error naming the path and never the value: absent; a symlink or not a regular file; any group/other bit; owner ≠ caller. Unix only; elsewhere an error.

**S3. Port** (`DaemonAddress`, Linux): socket inodes from `/proc/<pid>/fd` links; `net/tcp{,6}` rows in state `0A` with those inodes on a loopback address; one tokenless POST `{}` to `.../Heartbeat` per port (2 s); the HTTP port is the one answering 401 with Connect code `unauthenticated` and message prefix `unauthenticated: missing CSRF token` (K3). Zero or several matches → error naming each candidate and its status.

**S4. `telepathy agy-provision`** (Linux; the human runs it, or an agent with the human's agreement; no hook, test or other verb runs it on the host): `WriteDaemonToken` (fresh v4 UUID, 0600, dir 0700, temp + rename — rotating the token); writes `<config>/systemd/user/antigravity-cli-daemon.service.d/telepathy.conf` with `EnvironmentFile=-<the token path written>` and `ExecStartPost=-<home>/.gemini/config/plugins/gray-area/bin/telepathy agy-wake` (agy's unversioned install; Claude Code's cache path is versioned and goes stale), refusing a path with whitespace, quotes or `%`; warns when the binary at that path is absent or does not know `agy-wake` (`<path> agy-wake --help` exits non-zero), which holds until a release carrying the verb reaches agy's install; prints `systemctl --user daemon-reload && systemctl --user restart antigravity-cli-daemon` for the human, saying the restart ends every running turn and the first start wakes every conversation with mail waiting. It runs no `systemctl`.

**S5. `telepathy agy-wake`** (the `ExecStartPost=` act). systemd terminates the whole unit when `ExecStartPost=` outlives `TimeoutStartSec` (90 s by default), whatever its `-` prefix (audit, systemd `service.c`), so the run has ONE total bound, 60 s from its start, and every request inside it carries a timeout of at most 5 s and never past the bound. It reads the token (soft-fail), polls `CurrentDaemon` and `DaemonAddress` every 500 ms for at most 45 s; then, for each row of `ScanMailboxes([DaemonRoot/brain])` with at least one readable message whose sender is neither `system` nor `telepathy-wake` (K4 measured only those clearing; K5 says the others would re-wake at every start), S1 steps (6)–(7). One stderr line per conversation (`woke`, `not woken <why>`, and `not woken: the 60 s bound ran out` for each not reached); every soft failure exits 0 so the daemon's start never fails. No state file.

**S6. `ScanMailboxes`** counts `telepathy-wake` entries in a new `Waiting.Wakes`, never as mail and never in `Oldest`; a mailbox holding only wakes is still a row; `telepathy agents` shows them as `+N wakes` (golden `agents-waiting`, `help-agents`).

**S7. Deletions** (no reader of the old shape). Census at f31e34f02, per-file counts:
- Registry: `git grep -c -E 'RegisterWith|RegistryDir|LinkedDir|HasLinked|markLinked|NotAnAntigravityShell|NoRuntimeDir|AlreadyLinked|LookupServerByKey|LookupDaemonServer|ListServers|LookupConversationServer|NewestLiveListing|RecordLiveness|recordLiveness|\bPrune\b|pruneServers|lockRegistry|registryLockWait|RecordShape|LinkedConversation|linkedAt|renameRecord|antigravity-servers|antigravity-linked|register\.lock|ServerKey|HostServer|EnvAddress|EnvToken' -- plugins/gray-area shared docs README.md scripts AGENTS.md` → gray-area README 1, hooks/README 1, capture link.go 7, link_test 16, main.go 1, live_agy_status_test 1, live_agy_test 8, peer agy.go 102, agy_lock_other 1, agy_lock_unix 1, agy_lock_unix_test 5, agy_lock_windows 2, agy_proc_linux 2, agy_proc_other 2, peer.go 1, peer_test 42, record_test 66, registry_concurrency_test 25, registry_wait_test 10, telecli link.go 9, root.go 16, send.go 8, send_queued_test 4, send_route_test 14, send_test 25, help-link.golden 2, send-queued-records-unreadable.golden 1. All go, with `Server`, `ServerIdentity.Daemon`, `Delivery.NotWoken`; `gofrs/flock` stays (tests).
- Routes: `git grep -c -E 'RouteAgyDaemon|RouteAgyOwnRecord|RouteAgyHolderRecord|antigravity-daemon\b|antigravity-server-own|antigravity-server-holder' -- . ':!plans'` → gray-area README 1, SKILL.md 2, live_agy_status_test 1, live_agy_test 1, agy_dial_linux_test 1, peer.go 6, peer_test 6, send.go 7, send_route_test 6, send_test 2, help-send.golden 3. All go.
- Link: `git grep -c -E 'telepathy link|link --help|linkRequest|newLinkCmd|send`/`link|link\.go' -- . ':!plans'` → AGENTS.md 1, README.md 1, docs/local-development.md 1, gray-area README 2, hooks/README 1, SKILL.md 3, capture link.go 7, link_test 20, main.go 1, telecli link.go 1, root.go 1, send.go 1, send_route_test 1, help-link 1, help-send 1, hookcore asymmetry_table_test 4, events_test 6, live_asymmetry_test 1, live_events_test 3. The capture hook keeps its PreInvocation version check and loses only `AdditionalContext`. hookcore fixtures and failure texts are reworded; asymmetry `RelyingOn` paths move to `gray-area-capture/main.go` and `telecli/agyprovision.go`. `frank-exchange-of-views/.../rawhtml.go` is another sense and stays. `hook_test.go:31`'s `localPidDomain` stub goes with `link.go`. `ProcessStartTime`/`procStartTime` go if only deleted tests call them.
- Prose the regexes miss, rewritten: `scripts/universe.sh:264-266`, `docs/local-development.md:43`, `telecli/main_test.go:13`, `cmd/telepathy/main_test.go:37`, `runtimedir_test.go:16`.
- Refusal shapes: `git grep -c -E 'IsRefusal|RefusedError|RateLimitedError|UnreachedError|retryAfter|SendAgyServer|NewAgyClient|rateLimit' -- plugins/gray-area shared docs` → live_agy_status_test 11, agy.go 23, agy_dial_linux_test 5, peer.go 1, peer_test 18, send.go 15, send_queued_test 1, send_route_test 19. All go but `NewAgyClient`; the missing-token recogniser lives in S3. `Env.RuntimeDir`, `Env.HostServer`, `Env.RecordLiveness` go; `Env.Sleep` stays for `agy-wake`'s port poll.
- Goldens: delete `help-link`, `send-queued-{daemon-not-linked,daemon-record-unknown,daemon-record-unreadable,records-unreadable,holder-took-it,holder-unreached,no-runtime-dir,two-holders,refused,rate-limited,daemon-unreached}`, `send-unknown`; add `help-agy-provision`, `help-agy-wake`, `send-queued-{no-token-file,token-file-refused,port-not-found,wake-failed,woken}`; update `help-root`, `help-send`, `help-agents` and kept `send-queued-*`.

**S8. Isolation.** `telecli` `TestMain` also points `XDG_CONFIG_HOME` at the suite's directory and `TestDefaultsNeverReachTheHostConfig` asserts `Defaults().TokenFile` lies under it (on Windows `TestMain` sets `APPDATA` too). `runtimedir_test.go`'s sweep requires every exec of the built binary to set `XDG_CONFIG_HOME` as well as `XDG_RUNTIME_DIR` (the latter names the host's user bus, which `CurrentDaemon` asks); `TestTheBuiltBinary` sets both. `universe.sh`: the claude harness gets `XDG_CONFIG_HOME=$WORKDIR/xdg-config`, one symlink per entry of the host's config directory except `gray-area` and `systemd`, so gh, git and the rest keep the host's configuration and the token file is out of reach (agy's is already `$WORKDIR/home/.config`).

**S9. Live tests** (`HOOKCORE_LIVE=1`, `shared/testhome` only, stream-json servers without `--dangerously-skip-permissions`, a chosen token in the server's environment, token file in a `t.TempDir()` outside the testhome `HOME`, `send`/`agy-wake` run in-process with `Env.CurrentDaemon` set to the test server): one test, `TestLiveAntigravityMailboxWake`, with four subtests: (a) `ChosenTokenIsHonoured`: `DaemonAddress` finds one port; the chosen token 200, another 401; (b) `SendWakesIdle`: two `send`s to an idle conversation (the first with no daemon injected, so it only queues) then the second queues, wakes, and the transcript shows the queued content injected and the model's reply carrying its marker; (c) `SendHeldElsewhere`: a second server that has run a turn on C holds it → no wake, message queued, the test server's pid has no C database open; (d) `WakeOnlyRealMail`: conversations with a real message, only a `telepathy-wake` entry, only a `system` entry → `agy-wake` wakes only the first.

## IV. Risk & Mitigation

| Risk (likelihood × impact) | Mitigation |
|---|---|
| Wake races a holder that opens C between S1(6) and (7) (low × high) | (6) runs last; residue stated (K6 is a check, agy has no server-side lock) |
| A resumed holder that has not opened the database is invisible (low × high; n=1) | accepted: the wake is the only live request, and the gap is one server's start-up; stated in `send --help` |
| A holder in another pid namespace (container sharing agy's data) is invisible (low × high) | accepted as plan 1 stated it; `send --help` keeps the sentence |
| `agy-wake` outlives `TimeoutStartSec` and systemd terminates the daemon (med × high) | one 60 s total bound, 5 s per request; a test with a stub that never answers exits within it |
| First provisioned start wakes every conversation with mail (certain × med) | S4 prints it; the human chooses when to restart |
| A marker that no turn clears re-wakes at every start (low × med) | S5 wakes on readable agent mail only (K4, K5) |
| A test reaches the host daemon or token (low × high) | S1 order (token first), S8 sweep, S9 token file outside testhome |
| Token leaks to output (low × high) | errors name the path only; a test searches every output for the fixture token |

## V. Verification Plan

1. `unset $(env | grep -o '^FEOV_[A-Z_]*')`; `go -C plugins/gray-area/tools test -count=1 ./...` with `XDG_RUNTIME_DIR`, `XDG_CONFIG_HOME`, `XDG_STATE_HOME` on scratch → ok. Includes `TestSendReadsTheTokenBeforeAnyContact` (port and daemon seams fail the test if called), `TestSendNeverPrintsTheToken`, `TestReadDaemonTokenRefuses`, `TestDaemonAddressFindsTheHTTPPort` (two local listeners answering K3's two shapes), `TestProvisionWritesBothFiles`, `TestAgyWakeOnlyRealMail`, `TestAgyWakeExitsZeroOnSoftFailure`.
2. `GOOS=windows` and `GOOS=darwin` `go vet ./...` in the module → clean.
3. `go -C scripts run ./check` on the committed tree → "0 failed".
4. Delete-the-fix: each new hunk reverted names a red test.
5. Live (a)–(d) in one `HOOKCORE_LIVE=1` run, with a stat listing of `$XDG_RUNTIME_DIR/gray-area` and `~/.config/gray-area` before and after → unchanged.
6. Real data, only with the human's explicit agreement at the time, naming that the restart ends every running turn and that `agy-wake` wakes the daemon-root conversations with unread agent mail (listed by `telepathy agents` beforehand): the human removes the experiment's `telepathy-token.conf` by hand, runs `agy-provision` and its printed restart, then one `send` to a scratch conversation through the branch's binary (V's round trip). G3's observable — `agy-wake`'s per-conversation lines in the unit's journal after a start — needs a release carrying the verb in agy's install; it is a tracked follow-up of the release. Plan audit: one `plan-auditor` pass before code.
