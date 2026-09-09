# The catalogue: gray-area as a trajectory miner with a SQL surface

> Status: **under review** — this plan is the decision document; no code lands until it merges.
> Companion to [`gray-area.md`](gray-area.md), whose Phase 3 this implements.

**Client under test:** Claude Code 2.1.260. Every figure below was measured on this box on
2026-09-08 with the command named beside it. Three of them overturned the design that preceded
them — hook frequency, the `ATTACH` cap, and the meaning of an absent `is_error` — and the one
figure that is an estimate says so.

---

## I. Summary & Goals

**Gray Area is a trajectory miner: it answers questions about the behaviour and intent of agents
from their recorded actions, words and thoughts** (gblock, 2026-09-08). Today it records only
*where* a trajectory is, in a per-project manifest, and nothing can read across sessions.

The cost of that is measured, not hypothetical. On 2026-09-06 two sessions argued to a standstill
over a census drawn from **1 of 28 project directories** and reported as a total; both conclusions
were wrong and the artifact could not arbitrate. On 2026-09-08, with a release tag held and six
agents live on this box, **every coordination failure of the day was a query nobody could run** —
three claims relayed from memory that the record would have refuted, a shared blocker board gone
stale because its rows were hand-copied, and four messaging round-trips to ask what was
outstanding.

So the surface is a **switchboard**: *which agents are active, who is working on what, is anyone
else editing this file, what did that agent actually do — and what was it trying to do.*
`SendMessage` asks an agent what it **believes** and needs it alive and willing. This asks the
record what **happened**.

Goals, in priority order:

1. **A queryable surface, not a fixed menu.** Baked verbs for the common questions, and **raw
   read-only SQL** for everything else, because the questions worth asking are not knowable in
   advance. The schema is therefore a **published contract**, not an implementation detail.
2. **Ingest stays inside stated latency budgets: 20 ms per turn, 50 ms at `SessionStart`
   steady-state, 500 ms on the one `SessionStart` per UTC day that runs the 325 ms sweep, and
   500 ms at `SessionEnd`.** The last is bounded from above by the vendor: all `SessionEnd` hooks
   share a 1.5 s budget (per the hooks reference; raisable only by a per-hook `timeout`), and
   gray-area shares it with prosthetic-conscience's `sc-sessionend`. gray-area does **not** raise
   `timeout`; it stays at a third of the shared budget. Four numbers, because each is a different run — an earlier draft stated the middle
   one alone, which its own measurement refutes once a day.
   Not "zero added latency" — an earlier draft said that while also making `SessionStart` do work,
   so the goal could not be traced to any passing check. A budget with a number can be. §II shows
   the budget still rules out the obvious design by an order of magnitude.
3. **Actions, words and thoughts** — all three, because intent is not recoverable from acts alone.
4. **Answers are conclusions, not dumps** ([[context-efficiency]]).

**Non-goals.** Cross-machine aggregation; semantic search; a resident daemon (§II); an MCP surface
(§III names why it is scoped out); replacing `SendMessage` for intent and negotiation.

---

## II. Design

### The measurements that decide it

| measurement | result |
|---|---|
| bare Go process spawn — the floor under any hook | **4.33 ms** |
| tool calls per turn / turns per session | **8.8** / ~131 |
| `PostToolUse` ingest — busiest measured session, 2,187 calls (`5627f39a`) × spawn | **9.5 s**; at the ~1,150-call average (8.8 × 131) **5.0 s** |
| **post-turn (`Stop`) ingest — 131 turns × spawn** | **0.57 s per session** |
| SQLite insert on an open connection | ~4 µs (251k rows/s) |
| indexed query at 1M acts — *who touched this path* | **1.00 ms** |
| sequential open of 30 daily partitions | **14.6 ms** |
| `ATTACH` limit | **10 per connection** (main + 10 = 11), **not per process** — 60 separate connections opened fine |
| all three tiers, 3 weeks of corpus | **11.8 MB = 4.2%** of 280 MB |
| ripgrep over the whole corpus (real binary) | **33–54 ms** |
| `DELETE` one day (33k rows) from a 1M-row table | **325 ms**; file plateaus, no `VACUUM` |
| `is_error` absent on a `tool_result` | **18.4%** of calls — the normal SUCCESS shape for every tool but Bash |
| FTS5 index cost | **1–2 GB, ESTIMATED — not measured** |

### Ingest is a post-turn hook, not a daemon and not per-tool-call

`PostToolUse` fails goal 2 by an order of magnitude no store choice can fix: at the 4.33 ms spawn
floor it is **5.0 s per average session** (8.8 calls × 131 turns ≈ 1,150) and **9.5 s on the
busiest measured** (2,187 calls, session `5627f39a`), before the hook does anything. An earlier
draft quoted "2,000 calls" against the 131-turn average, which do not describe the same session.

Post-**turn** is a different frequency entirely — 8.8 tool calls per turn, ~131 turns per session,
so **0.57 s per session**. The `Stop` hook ingests incrementally from a per-file byte offset, so
steady-state work is proportional to new bytes, not corpus size.

**The final turn's text is NOT recovered from the payload, and that is a ruling, not an
oversight** (gblock, 2026-09-09).

The vendor documents that at `Stop`/`SubagentStop` the transcript may lag, and recommends reading
`last_assistant_message` from the payload instead. A design to do that — a *provisional* `word`
row, keyed and superseded once the transcript caught up — occupied rounds 12 through 21 of this
plan's audit and was rewritten five times, each time correctly: the supersession rule was
undecidable at close, then unexecutable, then lost data on the majority path, then duplicated on
it, before landing on a fire-time comparison with a measured 0.12% residue.

**Building the store settled it.** Ingest is transcript-sourced end to end, and on the real
corpus it stores 24,617 acts, 10,601 words and 248 thoughts — **not one row of which needed the
provisional path.** The machinery was never load-bearing for the tiers; it was a separate feature
recovering a final turn on the **4.1% of fires** that carry no usable payload text, at the cost of
the most defect-dense section of the design.

So it is **not built**, and the residue is stated instead of engineered around:

- A session's **final** turn may be absent from the catalogue when the transcript lagged and no
  later `Stop` or closure pass read it. Closure narrows this — it ingests before marking a session
  shut, so the bytes are read whenever the session ends while the box is still running hooks.
- The `word` table keeps `source`, `provisional` and `attribution` columns, always `'transcript'`,
  `0` and `NULL` today. They cost nothing, they are already on the published view, and they leave
  the door open if the residue is ever measured to matter.
- **What would reopen this:** a measurement showing final-turn loss above the residue rate, which
  §V.17 is the check for. Not an argument — a number.

The alternative to a stated residue was a state machine with four closers, a positional key, a
multiset comparison and an unverified-attribution flag, to recover one turn in twenty-five.

### Liveness is observed, not inferred

`~/.claude/sessions/<pid>.json` already holds, for every live session on the box:

```json
{"pid":2716,"sessionId":"67082c17-…","cwd":"…/worktrees/bridge-cse_019wLrgs…",
 "startedAt":1788512788587,"procStart":"5677","version":"2.1.260","kind":"interactive",
 "pidDomain":"linux:885fdabadec87be82908a4d724a5fe6e:pid:[4026531836]"}
```

Measured: six entries, six live sockets in `/run/user/<uid>/cc-socks/`. `sessionId` links to the
transcript and `cwd` names the worktree, so *which agents are active* and *who is working on what*
are answerable from this alone.

**`procStart` is what makes it sound.** A pid alone is unsafe — pids are reused, so a recycled one
would report a dead agent as alive, a wrong answer indistinguishable from a right one. pid plus
process start time is a durable process identity. So a session is live when its file exists, the
pid exists, and the start time matches; otherwise **`ended`** — **but only inside the same
`pidDomain`.** The vendor writes that field as `linux:<machine-id>:<pid namespace>` (verified:
`/etc/machine-id` + `readlink /proc/self/ns/pid`), and it is their own guard that a pid is only
meaningful within one namespace on one machine. An earlier draft compared pid and `procStart`
against local `/proc` unconditionally, so a session file written from a container or another host
sharing `$HOME` would have resolved to `ended` — the confident-wrong answer the third value exists
to refuse. The rule: **`pidDomain` must equal the querier's own before pid or `procStart` are
compared; a mismatch is `unknown`, never `ended`.**

**And that comparison is per-OS code this repository does not have** — `grep` finds no
process-liveness anywhere under `plugins/` or `scripts/`, while the release ships six GOOS/GOARCH
pairs. Rather than hand-roll three probes, **liveness is scoped to Linux** and the other platforms
report **`unknown`**, explicitly, as a third value beside `live` and `ended`. On Linux the start
time comes from field 22 of `/proc/<pid>/stat`, which needs no dependency. `unknown` is a real
answer — a board that renders an unmeasurable agent as `ended` would be the confident-wrong output
this plugin exists to refuse — and extending it is a later change with its own files. That is a
scoped-down **complete** concept, not a truncated one: the full concept additionally needs
`liveness_darwin.go` and `liveness_windows.go`, named here so the omission is tracked rather than
discovered.

The hook payload carries **no** pid (verified across 1,228 real rows: `agent_id`, `agent_type`,
`session_id`, `transcript_path`, `agent_transcript_path`, `cwd`, `effort`, `permission_mode`,
`prompt_id`, `stop_hook_active`, `background_tasks`, `session_crons`, `hook_event_name`,
`last_assistant_message` — nothing process-shaped). Identity comes from the payload; liveness comes
from `sessions/`. They are different questions with different sources.

### The three tiers, and why keeping them is free

| tier | contents | measured |
|---|---|---|
| **`action`** | `(session, seq, ts, tool, target, outcome)` | 23,076 calls, ~2.8 MB |
| **`word`** | assistant text and user prompts | 9,837 texts, 9.0 MB |
| **`thought`** | thinking summaries | 248, 0.06 MB *(capture was off until 2026-09-08)* |

**11.8 MB of extracted text — 4.2% of the 280 MB corpus.** The other 96% is tool *results*: file
contents, command output, base64 images.

**On disk the store is 28.2 MB, 9.3%** — measured after building it, against 4.2% for the text
alone. The gap is SQLite: row overhead, three indexes and the WAL. Still small enough that the
retention argument is unchanged, but the 4.2% figure describes the signal and **not** what holding
it costs, and an earlier draft used it as though it described both. Bulky, reproducible, and not what you mine for behaviour or intent. So the
discipline is not "index, not copy" — it is **keep the signal, drop the bulk**, and the bulk was
always the expensive part.

**`outcome` is an enum, and the absent case is SUCCESS — which an earlier draft got backwards.**
Measured across 20,947 invocations paired by `tool_use_id`, four ways rather than three:

| | share | |
|---|---|---|
| result present, `is_error=false` | 79.7% | `ok` |
| result present, **`is_error` absent entirely** | **18.4%** | `ok` |
| result present, `is_error=true` | 1.9% | `error` |
| **no `tool_result` block at all** | 0.1% (12 calls) | `unresolved` |

The convention is **per tool**, and that is the whole finding (re-measured 2026-09-08; the corpus
is live, so absolute counts drift and only the RATES are stable):

    tool          calls   false   true   absent   no result
    Bash          17,318 16,969    338        0          11
    Read           1,342      0     23    1,319           0
    Edit           1,325      0     21    1,304           0
    SendMessage      107      0      0      107           0

**Bash never omits the flag: 0 absent, on any call that produced a result.** An earlier draft wrote
"17,027 of 17,027", which did not sum — 16,682 + 335 left 10 calls unaccounted, and those 10 were
Bash calls with no `tool_result` at all. The denominator matters and is now named: of Bash calls
that **produced a result**, the explicit-flag rate is **100.00%**; over *all* Bash calls it is
99.94%, and the difference is the `unresolved` population, not a flag that went missing.

So absence of `is_error` is the *normal success shape* for every tool except Bash. A rule mapping
absent → `unresolved` — which a previous draft of §V asserted — would have mislabelled **3,844
successful calls**. `unresolved` means exactly one thing: no `tool_result` was ever recorded,
because the call was killed or interrupted mid-flight. A boolean would force those 12 into `ok` or
`error`, and both are lies.

**And that is why R2's detector is the Bash invariant, not the absent case.** `is_error` is a
vendor field; if it were renamed, every Bash result would read as absent-therefore-`ok`, the error
rate would fall silently to zero, and a check asserting "absent ⇒ unresolved" would pass
identically before and after — firing on the healthy path and silent on the failure, the exact
[[facts-are-fields]] clause-3 defect this plan cites elsewhere. What *does* move is the shape:
**of Bash calls that produced a `tool_result`, 100.00% carry an explicit `is_error`.** So the check
asserts that rate — over that denominator, not over all calls — and that Bash's error count over
the window is non-zero. A rename breaks both: every Bash result would become `absent`, dropping the
rate to 0% and the error count to zero.

**Reasoning is mined, not withheld.** Recovering intent is an explicit goal of the plugin. The
guard against summaries becoming evidence lives in **FEOV's citation ledger**, not here (gblock,
2026-09-08; landed as #844) — an anchor requires a `finding_id` and a location, and a summary has
neither, so it cannot form one. That is a property of the schema rather than a rule anyone must
remember.

### Store, retention, cleanup

    ~/.local/state/special-circumstances/catalogue/catalogue.db

**ONE DATABASE, NOT A PARTITION PER DAY — and this overrides an earlier instruction, so it is
flagged rather than swapped quietly.** gblock specified "a db with a timestamp for a name and a
ttl, and the hook as the cleanup". That predates the raw-SQL ruling, and the two are in direct
tension: **SQLite cannot join across separate connections**, and `ATTACH` is capped at **10 per
connection** (measured: refused at the 11th). So a day-partitioned store cannot serve arbitrary
user SQL over a month — a query would silently see one partition, or eleven days, and Goal 1 is
the whole point of the surface. Partitioning loses.

The cost of giving up `unlink` is measured and small. On a 1M-row table:

| | |
|---|---|
| `DELETE` one day (33,333 rows), indexed on `day` | **325 ms** |
| indexed query afterwards | **0.18 ms** |
| ingest a fresh day back | 211 ms |
| file size across a full delete+insert cycle | **69.2 MB → 69.2 MB** |

**No `VACUUM` is needed**: SQLite reuses freed pages, so at steady state the file plateaus rather
than growing.

**The sweep needs its own budget, because 325 ms does not fit in 50 ms.** An earlier draft claimed
both in one sentence, which is a contradiction its own measurement refutes. The sweep is a no-op on
every `SessionStart` but the first of a new UTC day; on that one run it costs 325 ms. So the budget
is stated as two numbers — **50 ms steady-state, 500 ms on the day-rollover run** — and §V.2
asserts both, with the rollover case forced rather than waited for.

Under `~/.local/state/`, **not** `~/.claude/`: writing our store into the vendor's directory is
squatting in a namespace we do not own.

**Archiving is not observable and nothing keys off it.** `ArchiveSession` is an HTTP POST to a
remote API (`[bridge:api] POST … 409 (already archived)`) operating on cloud sessions; it writes no
local state. Whether mined signal should outlive the window at all is **#845**, deferred.

### The SQL surface

Raw read-only SQL, plus verbs (`agents`, `session`, `touched`, `find`, `backfill`) that are ordinary queries
with names.

**Read-only is enforced and measured.** On `mode=ro&_pragma=query_only(1)`: `INSERT`, `UPDATE`,
`DELETE`, `DROP`, `CREATE` all refused. A runaway `act a, act b, act c` cross join was cancelled by
context deadline at **exactly 300 ms**, so a bad query cannot wedge the box.

**Two holes, and the gate that closes them is named rather than gestured at.** `ATTACH` and
`PRAGMA writable_schema` are both permitted on a read-only connection (measured). On this box
neither is an escalation — every caller already has `Read` and `Bash`, so attaching a file it could
simply open grants nothing — but it becomes one the moment the surface is less trusted than the
caller.

**Both holes are closed, with two calls.** An earlier draft of this plan asserted that neither a
statement classifier nor an authorizer callback existed in `modernc.org/sqlite`, and shipped the
holes on that basis. **That was an unmeasured negative claim and it was wrong** — the plan-auditor
caught it, and re-measuring confirms (on **v1.57.0**; note this is *feov's* pin and does not constrain this module — `go mod tidy` here resolves **v1.58.0**, on which all four mitigations were re-checked and hold):

| mitigation | measured |
|---|---|
| `_defensive=1` in the DSN | `PRAGMA writable_schema=1` becomes a **silent no-op** — reads back `0` with it, `1` without, on the same connection |
| `sqlite.Limit(conn, sqlite3.SQLITE_LIMIT_ATTACHED, 0)` (exported at `sqlite.go:1302`, constant at `lib/sqlite.go:3876`) | `ATTACH` fails: **`too many attached databases - max 0`**. Previous value was 10 |

**But `Limit` binds a CONNECTION, not the pool — and a draft of this section got that wrong too.**
`func Limit(c *sql.Conn, …)` takes a single connection. Measured on v1.57.0: after
`Limit(c1, ATTACHED, 0)`, `c1` refuses `ATTACH`, while **a second connection from the same
`*sql.DB` reads the limit back as 10 and `ATTACH` succeeds** — as does `db.Exec("ATTACH …")`
straight through the pool. `_defensive` is different: it is applied at connection open, so it holds
on every pooled connection (measured: `writable_schema` reads 0 on both).

So the query path **must pin its own connection**: acquire a `*sql.Conn`, apply
`Limit(ATTACHED, 0)` to it, run the user's statement on that connection, release it. Never
`db.Query` through the pool, because the pool may hand out a connection the limit was never
applied to. §II's own measurement row — *"10 per connection … not per process"* — was the evidence
for this all along, and an earlier draft printed it without drawing the conclusion.

So the surface is: `mode=ro` + `query_only(1)` + `_defensive=1` at the DSN, plus a **pinned
connection** carrying `Limit(ATTACHED, 0)` per query. An ordinary `SELECT` works under all four.

Two honest qualifications:

- **`RegisterPreUpdateHook` is NOT a veto** and is not proposed as defence in depth. Its callback is
  `func(SQLitePreUpdateData)` with no error return (`pre_update_hook.go:35`), so it can *observe* a
  write and cannot *refuse* one. Writes are refused by `mode=ro`, which is measured; the hook would
  only tell us afterwards.
- A **text pre-filter** over the SQL — "must start with SELECT" — remains deliberately **not**
  proposed. It is a pattern standing in for a schema, defeated by a leading comment or an
  unanticipated form, and now unnecessary: the two mechanisms above are SQLite's own.

**The schema is a contract, so callers query VIEWS.** If agents write their own SQL, a renamed
column breaks work that is not in this repo and cannot be swept. #819 is the warning — 0 of 7 run
archives are readable by any current binary, because words were retired and an epoch moved. So:
stable views named `action`, `word`, `thought`, `session` and `skip`; base tables free to evolve beneath
them; views evolve **additively only**; and `capture_build` on every row (#818's pattern) so a
reader can tell which binary wrote it.

### Full text stays unindexed

Ripgrep searches the whole corpus in **33–54 ms**, measured through a real binary. An FTS5 index is
**estimated** at 1–2 GB — not measured — and would need invalidating on every append and rebuilding
whenever the client deletes a session, to beat 40 ms. `find` shells to `rg` over the paths the
catalogue names and joins each hit back to who/when/where, passing the search term as **a single
argv element, never through a shell**.

Note for anyone re-measuring: inside Claude Code `rg` is a shell **function** the client injects,
not a binary, so a hand-run measurement is taken through a path no Go program can use. ripgrep is
declared in `requirements.json` as `recommended` (#843) and is **promoted to `required` in the
change that ships `find`**, because from that moment a missing binary returns zero hits that read
exactly like an honest zero.

---

## III. Scope — carriers and consumers, censuses run and adjudicated

**No new command directories.** Ingest is a new `Stop` event on the existing `gray-area-capture`;
queries are new verbs on the existing `gray-area`. So `requirements.json`'s `_hook_binaries` and
`docs/setup-script.md:148`'s hook-binary count (18) are **untouched**, and `scripts/pluginparity`
stays green. This was checked because two new directories would have turned it red.

- [NEW] `plugins/gray-area/tools/internal/catalogue/` — schema, views, incremental projection,
  query path (pinned connection + `Limit`).
- [NEW] `internal/catalogue/liveness_linux.go` and `liveness_other.go` (build-tagged) — the
  `/proc/<pid>/stat` start-time comparison, and the `unknown` answer everywhere else. Named
  because §II scopes liveness to Linux and an unnamed omission is one nobody tracks.
- [NEW] `internal/catalogue/testdata/corpus/` — **synthesized** transcripts (see §V).
- [NEW] `internal/catalogue/backfill.go` — **the component an earlier draft omitted entirely.** A
  `Stop`-hook-only store contains nothing until it has run for a month, so §V's checks over a
  month window would have had nothing to read. Backfill projects the existing corpus once, is
  triggered by the `gray-area backfill` verb — **the fifth new verb, counted in §III's dispatch-table and README carriers** — explicit, never from a hook, and is **exempt from goal 2's
  budgets because it is not on any agent's turn** — **MEASURED, now that the component exists, and the estimate was 7x optimistic.** The
  estimate said "a cold seed should be around a second", from a 545 MB/s `python3` line-scan marked
  as an upper bound. The real backfill — full JSON parse plus SQLite writes — runs the whole
  422-file corpus at **44 MB/s: 304 MB in 6.8 s**. A second pass over unchanged bytes is **14 ms**
  and adds zero rows, so the cost is paid once. The scan rate was measuring the wrong thing: it
  bounded reading, and the work is parsing and writing. It is idempotent by
  `(session, seq)` so a re-run cannot double-count.
- [MODIFY] `cmd/gray-area-capture/main.go` — handles `Stop` (ingest) and sweeps rows past the window
  at `SessionStart`. **`Stop` is an explicit new branch, NOT a fall-through.** Today
  `main.go:522–537` routes every non-`SessionStart` event to `appendRow(buildRow(...))`, i.e. a
  seat/turn-end row; a `Stop` binding that merely adds ingest without leaving that path would write
  ~131 `event-names-no-seat` rows per session into the manifest. So `Stop` **and `SessionEnd`** are each explicit branches that
  ingest and write **zero manifest rows** — an earlier draft gave `Stop` the branch and left
  `SessionEnd` to the same fall-through it had just named. The existing `SessionStart` and
  `SubagentStop` manifest writes are unchanged. The `-event` flag help at `main.go:488`
  (`SubagentStop | SessionStart`) gains both events.
- [MODIFY] `cmd/gray-area/main.go`, `inspect.go` — five verbs (`agents`, `session`, `touched`, `find`, `backfill`) plus `sql`. The existing six are
  untouched.
- [MODIFY] `plugins/gray-area/hooks/hooks.json` — two new bindings, `Stop` and `SessionEnd`.
- [MODIFY] `plugins/gray-area/requirements.json` — **three** changes, the third a carrier an earlier
  draft missed: `:8`'s `_tier_comment` reads *"RECOMMENDED, AND DELIBERATELY AHEAD OF ITS CONSUMER …
  PROMOTE TO `required` in the same change that ships the verb"*. Flipping `tier` without rewriting
  that leaves a carrier instructing the change this plan is making. The other two: The `gray-area-capture` entry's
  `event` field says `SubagentStop` and already omits `SessionStart`; it becomes accurate and
  gains `Stop` and `SessionEnd`. **And
  ripgrep is promoted `recommended` → `required`**, because §II commits to promoting it "in the
  change that ships `find`" and this is that change. On `main` today it is `recommended` (#843,
  merged); an earlier version of this branch carried `required` already, from a superseded commit
  that has been dropped — so the delta is real and is this change's to make. The consumer cost is
  graded as R10.
- [MODIFY] `plugins/gray-area/README.md` — the miner's new surface, the store, the month window.
- [MODIFY] **`README.md`** (repo root) — its verb table at `README.md:116`, **and `README.md:120`**,
  which says *"the manifest is an index of where trajectories are, never a copy of their
  contents."* That sentence is the plugin's consumer-facing surveillance disclosure and it becomes
  false the moment the `word` and `thought` tiers exist. It is rewritten to the new scoping
  statement: the manifest stays an index; the catalogue at
  `~/.local/state/special-circumstances/catalogue/catalogue.db` copies **tool names and targets,
  assistant text, user prompts and thinking summaries** — never tool results — for **one month**,
  on this machine only, never in the repository.
- `plugins/gray-area/README.md:37` — *"`SubagentStop` also carries `last_assistant_message`, and
  that is deliberately not recorded."* **UNCHANGED, and now true of the catalogue as well as the
  manifest**: with the provisional path not built, nothing in this plugin reads that field. An
  earlier draft listed this line as a carrier needing rewriting, which it would have been had the
  machinery shipped.
- [MODIFY] `plugins/gray-area/README.md:124` — *"the manifest is gitignored; nothing leaves the
  box."* Gitignore is no longer the privacy mechanism once the store lives under
  `~/.local/state/`, outside any repository. Rewritten to the same scoping statement: nothing
  leaves the box because the store is outside the tree, not because of a gitignore line.
- [MODIFY] `plugins/gray-area/requirements.json:34` — `_hook_binaries.gray-area-capture.purpose`
  says *"an index, never a copy of conversation content."* Rewritten: the manifest row is an
  index; the same binary now also writes the catalogue, which copies the three tiers named above.
- `plugin.json` / `marketplace.json` descriptions ("Captures each seat's own transcript by path
  at SubagentStop") are **partial, not false** — unchanged, noted so the omission is deliberate.
- [MODIFY] `plugins/gray-area/tools/go.mod` + [NEW] `go.sum` — **not "one dependency".** Measured
  in a scratch module: **pinned at `modernc.org/sqlite v1.57.0`** to match the version §II's mitigations were measured on, rather than whatever `tidy` resolves (currently v1.58.0). The `go` directive is set to a **current patch release** (`go 1.25.13` today) rather than the `go 1.25.0` that `tidy` writes — see the qlty entry below for the measured reason. `go mod tidy` adds **1 direct + 9 indirect** requires (`modernc.org/libc`,
  `mathutil`, `memory`, `golang.org/x/sys`, `go-humanize`, `uuid`, `go-isatty`, `go-strftime`,
  `bigfft`), 25 modules in `go.sum`, **and rewrites the directive `go 1.25` → `go 1.25.0`.**
  **cgo rejected**: the release job cross-compiles six platforms and a cgo SQLite needs a C
  toolchain per target — a cost this repository already pays once for tesseract and should not pay
  twice for a store whose switchboard queries measure 1.00 ms and 2.38 ms at 1M rows. All six pairs
  cross-compile pure-Go today.
- **NO qlty suppression is needed — measured, and this reverses both an earlier draft and a ruling
  made on its framing.** The draft said `go 1.25` → 0 findings and `go 1.25.0` → 39, quoting
  `.qlty/qlty.toml:88-99`, and gblock ruled "add a real suppression" on that basis. Running the
  scanner instead of quoting the comment:

      go 1.25.0    ->  46 findings      (the comment says 39; CVEs have accumulated since)
      go 1.25.13   ->   0 findings
      go 1.25      ->   0 findings, but `go build` and `go test` REFUSE once a dependency
                        exists: "updates to go.mod needed; to update it: go mod tidy"

  So the fix is **not** a suppression, which would permanently hide real findings on that file. It
  is to set the directive to a **current patch release**, which is exactly what
  `frank-exchange-of-views/tools/go.mod` already does (`go 1.25.13`, scanning clean today). This
  module does the same. The cost is that the directive needs bumping as CVEs accumulate against
  whatever patch it names — a maintenance task, not a hidden finding.

  **`.qlty/qlty.toml:101` is a [MODIFY] CARRIER for this change**, which an earlier draft missed by
  treating that file as already-corrected. Its roster line reads *"a module with no dependencies
  keeps the short form (gray-area, prosthetic-conscience); a module WITH dependencies pins a current
  patch release (frank-exchange-of-views, scripts)"* — and this change moves gray-area across that
  line, making the sentence false about the post-change tree. `f6795bac` (the correction) is already
  an ancestor of this branch, so moving gray-area into the pinned group is *this* change's job.
  Separately, **`.qlty/qlty.toml:88-99` was STALE** — it said feov carried 39 accepted findings when
  feov reports clean — and **was corrected in #850, already merged and an ancestor of this branch** —
  so the in-tree comment is right about today and the roster line above is what this change still
  owes. An earlier draft described the correction as pending. All four `go.mod`
  files were scanned before rewriting it: 0 findings each.
- [MODIFY] `.github/workflows/hooks.yml` — `cache-dependency-path` names
  `plugins/gray-area/tools/go.mod` at **:95, :828, :1469**; with a `go.sum` present the key must
  track it, as feov's already does. `:673` (not `:663`) records that Windows `t.TempDir` SQLite tests cost ~73 s for
  `recordsql` — but that block already installs the ImDisk RAM disk which takes it to ~5 s, so the
  cost this plan imports into the gray-area leg is seconds, not a minute. Corrected rather than
  left overstated.

**Census 1 — manifest-format readers.** `grep -rlnE 'trajectories-|SessionRow|Seats\('` with **no
`--include` filter** — an earlier run of this census used one and returned 18, missing two `.mjs`
files. **21 entries, every one adjudicated:**

| # | file | disposition |
|---|---|---|
| 1 | `plans/gray-area-catalogue.md` | this plan |
| 2–3 | `plans/gray-area.md`, `plans/rearm-coverage-experiment.md` | design documents, not consumers |
| 4 | `feov/tests/simulator/debate.test.mjs` | **string sharer** — `citationSeats(`. No change |
| 5 | `feov/tests/simulator/prompts.test.mjs` | **string sharer** — `captureSeats(`. No change |
| 6–11 | `feov/tools/internal/{cli/diagnostics.go, record/planguard/planguard_test.go, record/queries.go, record/writelookups_parity_test.go, seatclass/seatclass.go, seatclass/seatclass_test.go}` | **string sharers** — `KnownSeats()` / `RegisteredSeats(`, FEOV's *debate* seats, a different concept sharing a token ([[facts-are-fields]] clause 4). No change |
| 12 | `plugins/gray-area/README.md` | **carrier** — [MODIFY], in scope |
| 13–14 | `gray-area/tools/cmd/gray-area-capture/{main.go,main_test.go}` | **carriers** — [MODIFY] |
| 15–16 | `gray-area/tools/cmd/gray-area/{inspect.go,main.go}` | **carriers** — [MODIFY] |
| 17–20 | `gray-area/tools/internal/claims/{coverage,manifest}{,_test}.go` | **carriers**, read-side. Unchanged by this plan: the catalogue is built *from* them |
| 21 | `prosthetic-conscience/.../checkpointseal/main_test.go` | **string sharer** — `ConcurrentSeats(`. No change |

**Census 2 — verb-surface consumers.**
`grep -rln -E 'gray-area (tools|checkpoint|rework|stalls|pr|coverage)' .` — **17 entries**, pasted
in full (an earlier draft claimed 18 and enumerated 17, and gave no command at all):

    1  README.md                                                   CHANGES - repo verb table
    2  plans/gray-area-phase-2-build.md                            unchanged - design doc
    3  plans/gray-area-phase-2.md                                  unchanged - design doc
    4  plans/gray-area.md                                          unchanged - design doc
    5  plans/rearm-coverage-experiment.md                          unchanged - design doc
    6  plugins/gray-area/README.md                                 CHANGES - plugin verb table
    7  plugins/gray-area/commands/audit-checkpoint.md              unchanged - names `checkpoint`
    8  plugins/gray-area/commands/audit-pr-body.md                 unchanged - names `pr`
    9  plugins/gray-area/commands/audit-repetition.md              unchanged - names `rework`
    10 plugins/gray-area/commands/audit-seat-coverage.md           unchanged - names `coverage`
    11 plugins/gray-area/tools/cmd/gray-area/main.go               CHANGES - the dispatch table
    12 .../internal/claims/conformance_test.go                     unchanged - names `checkpoint`
    13 .../internal/claims/manifest.go                             unchanged - error text
    14 .../internal/claims/recheck_test.go                         unchanged
    15 .../internal/prbody/prbody.go                               unchanged - names `pr`
    16 .../internal/prbody/prbody_test.go                          unchanged
    17 prosthetic-conscience/.../checkpoint/loopproblems_test.go   unchanged - golden text

**Three change; fourteen do not.** The six existing verbs keep their names and behaviour, so every
document that names only those stays correct. `cmd/gray-area/main.go` changes because it *is* the
dispatch table (the `case` at `main.go:180`).

**`cmd/gray-area/inspect.go` is [MODIFY] in the change list and correctly absent from this census**
— it implements verb bodies but contains no `gray-area <verb>` string, so the census's pattern does
not and should not match it. Noting the reconciliation because a file that is changing while
missing from a census is exactly the shape that should be questioned.

**Scoped out, per [[complete-the-concept]]:** an **MCP surface**. No plugin in this repository
declares an MCP server and no plugin-level `.mcp.json` exists, so the declaration mechanism is
unverified; naming a file that does not exist is what the first audit caught. The full concept
would additionally touch that mechanism, `plugin.json`, and a per-session stdio shim. The CLI plus
raw SQL is a complete concept without it.

**Not in the v2.0.0 boundary**: additive, no shipped schema changes, manifest keeps schema 5.

---

## IV. Risks, graded — each with the §V check that would notice it

| # | risk | grade | check |
|---|---|---|---|
| R1 | **Projection drifts from the vendor's transcript format.** Fails toward under-counting. | likely eventually × medium × contained | §V.4 — a corrupt line yields `unparsed` with a count, never a session with zero acts |
| R2 | **`is_error` is renamed and every call reads `ok`.** The failure rate silently becomes zero. The naive check — "absent field ⇒ `unresolved`" — CANNOT notice this: absence is already the normal success shape for every tool but Bash, so it passes identically before and after. | medium × high × cheap | §V.5 — asserts the invariant that actually moves: **of Bash calls that produced a `tool_result`, 100.00% carry an explicit `is_error`** (0 absent of 17,307 measured), and its error count is non-zero. A rename drives the rate to 0% and the count to zero |
| R3 | **A renamed column breaks callers' SQL** outside this repo. | certain over time × high × cheap | §V.6 — a test asserts the view columns, so renaming one fails here rather than in someone's query |
| R4 | **The month TTL deletes acts that could not be recovered.** An earlier draft graded this *certain*, on a retention claim now measured false: **nothing has been deleted**, so a swept row remains reprojectable from the transcript for as long as the client keeps it. The risk is real but **conditional** — it bites only if and when the client begins expiring transcripts, which it does not do today and which no setting on this box enables. | conditional × medium × **accepted** | **No check — accepted.** A check that the sweep works is not a check that the loss is acceptable, and the loss does not currently occur. #845 owns whether mined signal should outlive the window; the trigger to revisit is the client gaining a retention policy |
| R5 | **Byte offsets desynchronise** (a transcript truncated rather than appended). | low × medium × cheap | §V.8 — offset record stores size and the sha of the first 4 KiB; a mismatch reprojects from zero |
| R6 | **A malicious or clumsy query wedges the box.** | medium × medium × closed | §V.9 — writes refused; cross join cancelled at a deadline; `ATTACH` and `writable_schema` refused on the pinned connection |
| R7 | **Liveness reports a dead agent as live** via pid reuse. | low × high × closed | §V.10 — a stale `procStart` must resolve to `ended`, asserted with a fabricated mismatch |
| R8 | **First dependency in a zero-dependency plugin**, and its consumer-side install cost — `scripts/bootstrap-plugins.sh:83-99` builds hook binaries **on the consumer's machine**, and its own comment names dependency download as the part most likely to overrun the ~5 minute budget. Measured cold against an empty module cache: **3.7 s download + 23.0 s build, 348 MB cache.** Inside the budget, stated rather than assumed. | certain × low × **low to mitigate** — one-time and measured; nothing to do unless that budget tightens | named as a decision at §III's `go.mod` bullet (pin, cgo rejection, 1+9 requires), not absorbed |
| R11 | **A session's final turn can be missing** when the transcript lagged its last `Stop` and no closure pass read it. **ACCEPTED, not engineered around** (gblock, 2026-09-09): the payload recovery this would need was designed across ten audit rounds and is not built, because ingest needs none of it — see §II. Closure narrows it by ingesting before it marks a session shut. | likely × low × **accepted** | §V.3 compares word counts, so the residue's SIZE is measured rather than assumed; §V.17 measures the lag itself once the hook exists. A number reopens this, not an argument |
| R10 | **Promoting ripgrep to `required` BLOCKS consumers** who lack the binary. The mechanism is `toolchain.MergeStrictest` + `doctor.verdict` — a missing `required` tool returns BLOCKED — **not** a manifest key: `_tier_comment` is documentation on the entry, and an earlier draft cited it as though it were the enforcement. Deliberate: from the moment `find` ships, a missing `rg` returns zero hits indistinguishable from an honest zero, and a wrong answer is worse than a blocked install. | certain × medium × accepted | §V.14 — asserts the verb and the tier move in the same change, so neither can ship alone |
| R9 | **The board reads as complete when reasoning capture was off.** | medium × medium × mitigated | §V.11 — a session with zero non-empty thoughts is distinguishable from one never asked |

---

## V. Verification plan

**FIXTURE-BACKED VS LOCAL-ONLY, because CI has no corpus.** `hooks.yml:652` runs the gray-area job
on ubuntu and windows with `go test ./...`; no runner has a `~/.claude` transcript tree. A check
written against "the live corpus" therefore passes for absence of data, or skips and reports green
— the plausible zero this plan cites [[facts-are-fields]] clause 3 against, reproduced inside its
own verification plan.

So every check below is one of two kinds, and says which:

- **[FIXTURE]** runs everywhere, against `internal/catalogue/testdata/corpus/` — a tree of
  **SYNTHESIZED** transcripts written to the shapes measured in §II: one session with a Bash error,
  one with an absent `is_error`, one tool call with no `tool_result`, one corrupt line, one
  truncated file. **Not captured transcripts.** This repository is PUBLIC, and real transcripts
  carry user prompts, assistant text and `cwd` worktree paths — committing them is exactly the
  disclosure #845 was filed about, which an earlier draft of this very section walked into one page
  after citing it. Synthesis is not a compromise here: every shape the checks need is named in §II
  and can be written by hand, and a fixture that reproduces a measured shape is a better test than
  one that happens to contain it.
- **[LOCAL]** needs the real corpus and is **skipped in CI by an explicit guard that prints why**,
  never by silence. A `[LOCAL]` check that finds no corpus reports `NOT MEASURED — no transcript
  tree at ~/.claude/projects` and fails if `GRAY_AREA_REQUIRE_CORPUS=1`, which the pre-merge run
  sets. Absence of data is reported as absence, not as a pass.

Written before implementation. **Re-arms on:** any change under `internal/catalogue/`,
`cmd/gray-area-capture/`, or the view definitions.

1. **[FIXTURE]** `cd plugins/gray-area/tools && go build ./... && go vet ./... && GOOS=windows go vet ./... &&
   go test -count=1 ./... && go test -race ./...` — `-race` because both
   `scripts/check/gates.go:152` and CI run it over this module, and this change is the first to
   give it a connection pool. The Windows vet is in the loop because a hand-rolled path match cost a
   red CI leg on #839.
2. **[LOCAL]** — retagged: a wall-clock budget measured on one Linux box cannot be asserted on
   windows-latest or under `-race`, both of which CI runs (`hooks.yml:837-838`). It is guarded to
   linux, non-race, and reports `NOT MEASURED` elsewhere rather than passing quietly.
   `go test -run TestIngestBudget ./internal/catalogue/` — **execs the BUILT
   `gray-area-capture` binary** rather than benchmarking library work in-process, because the
   budget is stated as spawn floor plus ingest and an in-process benchmark cannot measure a spawn.
   Four cases are forced rather than waited for: the day-rollover (seed a row dated outside the
   window), a **closure with a nonempty unread tail** (the new work goal 2 previously had no case
   that could see — assert against the 50 ms steady-state budget), a
   **backlog exceeding both halves of the cap** — ≥40 non-live file-tails totalling ≥10 MB unread —
   asserting exactly **≤12 tails and ≤4 MB read per invocation**, the invocation inside 50 ms, and the
   remainder still queued rather than dropped; and **closure coinciding with the day-rollover sweep**
   (assert against 500 ms). An earlier draft asserted "the 8-per-run cap" over "20 non-live sessions"
   — the cap this section repudiated, in the unit it says bounds nothing, and a case an
   implementation with no cap at all would pass. —
   **goal 2, measured**: the `Stop` path stays under **20 ms per turn** (4.33 ms spawn floor +
   ingest); `SessionStart` under **50 ms** steady-state and under **500 ms** on the forced
   day-rollover run that performs the 325 ms sweep; **`SessionEnd` under 500 ms**, a third of the
   1.5 s the vendor shares across every `SessionEnd` hook. Four numbers, because each is a
   different run and an earlier draft asserted a single budget its own measurement refuted.
3. **[LOCAL]** `go test -run TestProjectionMatchesIndependentCount ./internal/catalogue/` — reads
   **the hook-populated catalogue** (not a fresh projection) and compares against a transcript-derived
   count taken by a different code path: session counts, per-session **act** counts, and per-session
   **`word`** counts. The arithmetic is **stated, not gestured at** — an earlier
   draft said promoted rows have no transcript counterpart "by construction", which §II refutes:
   byte-offset ingest *does* recover a lagged turn on the next `Stop`, so only a **final** lagged
   turn's text never reaches the file. So:

       FOR sessions closure has actually closed (session.closed_at IS NOT NULL):
         catalogue words  ==  transcript words

   **The equality is now exact rather than approximate**, because the provisional path is not
   built: every `word` row is transcript-sourced, so there is no residue term to carry. An earlier
   draft needed `+ count(source='payload' AND provisional=0)` and three stated exclusions; two of
   those exclusions remain as SCOPE — live sessions have bytes closure has not read, and sessions
   predating capture have transcript words and no catalogue words — which is why the comparison is
   over closed sessions and not the corpus.

   Words are compared, not only acts, because a final turn lost to transcript lag changes the word
   count and leaves the act count untouched — the residue §II states is the thing this check
   measures the size of.
4. **[FIXTURE]** `go test -run TestCorruptLineIsUnparsedNotZero ./internal/catalogue/`
5. **[FIXTURE]** for the model, **[LOCAL]** for the ratio: `go test -run 'TestAbsentIsErrorIsSuccess|TestBashAlwaysStatesTheFlag' ./internal/catalogue/` —
   the first pins the corrected model (absent ⇒ `ok`, only a missing `tool_result` ⇒ `unresolved`);
   the second is R2's real detector: Bash's explicit-flag rate is 100.00% **over calls that produced
   a result** (the denominator an earlier draft left unnamed, making the assertion fail on the live
   corpus at 99.94%), and its error count is non-zero. An earlier draft asserted the opposite of the first and would have mislabelled
   3,844 successful calls.
6. **[FIXTURE]** `go test -run TestViewColumnsAreTheContract ./internal/catalogue/` — asserts the exact column
   set of each view.
7. **[FIXTURE]** `go test -run TestSweepDeletesExactlyExpiredRows ./internal/catalogue/`
8. **[FIXTURE]** `go test -run TestTruncatedTranscriptReprojectsFromZero ./internal/catalogue/`
9. **[FIXTURE]** `go test -run 'TestWritesAreRefused|TestRunawayQueryIsCancelled|TestAttachRefusedOnQueryPath|TestWritableSchemaIsNoOp' ./internal/catalogue/` —
   asserts each mitigation §II claims, **through the connection the `sql` verb actually uses**, not
   a hand-held one: writes refused; a cross join cancelled by deadline; `ATTACH` refused; and
   `PRAGMA writable_schema=1` reading back 0. The `ATTACH` case is the one that would regress
   silently — a pooled `db.Query` instead of a pinned `*sql.Conn` reopens the hole and every other
   test still passes, so this check exercises the verb's own path deliberately. **And it must hold
   a SECOND live connection while it does so** (`SetMaxOpenConns(>1)`, the query issued while
   another conn is checked out): measured, an implementation that pins, sets the limit, releases,
   then queries through the pool is refused single-goroutine — because the pool hands back that
   same connection — and ships the hole under concurrency. A check that passes for that
   implementation is not a check.
10. **[FIXTURE]** (fabricated `sessions/` dir, not the real one) `go test -run 'TestStaleProcStartIsEnded|TestForeignPidDomainIsUnknown' ./internal/catalogue/` —
    the second case writes a session file whose `pidDomain` differs from the querier's and asserts
    the answer is `unknown`, not `ended`, even when a local pid happens to match.
11. **[FIXTURE]** `go test -run TestCaptureOffSessionIsDistinguishable ./internal/catalogue/`
12. **[PROCESS]** Revert-check every new test** against the mutation it exists to catch — one at a time, from a
    **file backup, never `git checkout`**: on #839 that silently failed to restore untracked files,
    four mutations accumulated, and the "KILLED" lines were damage piling up rather than tests
    working. The tell was three mutations reporting the same two failures.
13. **[FIXTURE]** Wiring pinned apart from mechanism — **and the fall-through pinned shut**:
    `go test -run 'TestStopWritesNoManifestRow|TestSessionEndWritesNoManifestRow' ./cmd/gray-area-capture/`
    asserts `-event Stop` **and** `-event SessionEnd` each leave the manifest at zero rows, because a wiring test that only proves the ingest CALL is
    present cannot see a manifest write that should not fire.** The `Stop` ingest call and the sweep call each get a
    test that fails when the CALL is removed, not only when the function breaks — #839's
    `boxWarnings` seam is the pattern.
14. **[FIXTURE]** `go test -run TestFindShipsWithRequiredRipgrep ./internal/catalogue/` — asserts
    the `find` verb and `requirements.json`'s `required` tier move together (R10), so neither can
    ship without the other.
15. **[FIXTURE]** `go test -run 'TestBackfillIsIdempotent|TestBackfillVerbDispatches|TestProvisionalWordIsSupersededOnce' ./internal/catalogue/` —
    the third asserts `backfill` idempotence over BOTH act and `word` counts: a second pass over
    the same corpus adds nothing. Measured on the real corpus at 14 ms and zero new rows. The
    provisional cases an earlier draft enumerated here — two outstanding at once, the `[X, X]`
    tail, the unverified attribution — are gone with the machinery they tested.

16. **[PROCESS]** `qlty check --no-progress --no-fix --filter=osv-scanner
    plugins/gray-area/tools/go.mod` — **expected: 0 findings.** Run on the branch before review,
    because the directive `go mod tidy` writes (`go 1.25.0`) scans at 46 findings and the one this
    change pins scans at 0. A non-zero result means the pinned patch release has accumulated CVEs
    and needs bumping — the maintenance this approach trades for not suppressing anything.
17. **[LOCAL]** `go test -run TestStopLagMeasured ./cmd/gray-area-capture/` — with the `Stop`
    hook bound, at fire time record transcript size, whether the payload's
    `last_assistant_message` is already present in the transcript tail, **and whether the payload's
    `prompt_id` resolves to the turn that just ended rather than the next one** — the untested leg
    of the identity claim in §II. Across a real session.
    **Expected: a number, reported either way** — "0 of N turns lagged" is a result; an absent
    measurement is not. This is the check the plan could not run before the hook existed.
18. **[LOCAL] End-to-end, observed.** Against the real corpus: `gray-area agents` names this session and
    the other five live ones; `gray-area touched <a path this session edited>` returns it;
    `gray-area sql 'SELECT tool, count(*) FROM action GROUP BY tool'` returns a distribution
    matching item 3's independent count **over the whole month window** — asserted against rows
    older than 11 days, because that is the bound a partitioned store would have imposed and the
    single-database design exists to remove.
