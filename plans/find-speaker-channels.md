# Who spoke — one classifier for `find` channels and `v_word.role` (#885)

> STATUS: proposal (2026-09-11), revision 3 — answers plan-audit rounds 1–2. Scope decided by
> gblock: fix BOTH carriers with ONE classifier, classify by fields the client already writes
> (origin, isMeta, commandMode, which tier a record is in), and accept a store shape bump
> (3 → 4). Builds on #894.

## I. Summary & Goals

**The defect (#885, extended by measurement).** `find --in user` and the store's
`v_word.role = 'user'` treat every user-role text as the human's words. Measured 2026-09-11 over
all 520 transcripts on this box (147 top-level, 373 subagent/workflow), most of them are not:
peer sessions' messages, background-task notifications, the lead's prompts to its seats, a
workflow coordinator's messages, and text the client (the harness) injects. A store backfilled
from this corpus holds **54 `v_word` rows that are peer messages labelled `user`, across 6
sessions**, and 2,342 subagent-tier `user` rows although subagent transcripts hold no human
prompt at all. And `--in C` filters on each transcript's MOST RECENT hit only (`find.go:238-252`
`keepChannel`), so an earlier match in C is dropped and the result reads "none with their most
recent hit in C".

**Objective.** One rule decides who a user-role record came from, reading only fields the client
writes, and both `find` and the word tier use it; `--in C` answers "did anything in channel C say
this?".

**Success criteria.**

1. `catalogue.SpeakerOf` classifies every user-role record and every `queued_command` attachment
   by the precedence in §III.1 into `user` (the human), `peer`, `notification`, `lead`, `harness`
   or `unknown_origin`; assistant records are `assistant`. It reads fields only — never message
   text.
2. `find` gains channels `peer`, `notification`, `lead`, `harness`, `unknown_origin`; every text hit
   and every `queued_command` attachment reports its speaker's channel. `--in peer` finds the #885
   repro message and `--in user` does not; `--in notification` includes notifications delivered
   mid-turn.
3. `v_word.role` carries the same values. Reprojected from this box's corpus (§V.4): 0 `user` rows
   in subagent or workflow transcripts; 0 `user` rows beginning "Another Claude session sent a
   message:" or "The coordinator sent a message".
4. `--in C` keeps a transcript if ANY of its matching records is in C; the row's WHEN/IN/SNIPPET
   come from the most recent hit IN C; HITS still counts every matching record; the "only the
   most recent 200 … were read" footer never appears for an `--in` search, because none is capped.
5. **`user` is the human plus a stated remainder** that no field distinguishes (§II): on this box,
   158 top-level user-role records with no `origin` and no meta flag — 132 prompts that PROGRAMS
   send to headless sessions (frank-exchange-of-views seat prompts), 15 slash-command records and 11
   local-command outputs — plus 79 interrupts, which are the human's. The help, README and skill
   say so in those terms.
6. `UserVersion` 3 → 4, so every store is rebuilt once through #894's mechanism and relabelled on
   backfill. #894 is unreleased: consumers still rebuild once at the next release (stamp 2 → 4).

## II. Technical Context

- Module `plugins/gray-area/tools` (Go 1.25.13, modernc sqlite v1.57.0). Branch
  `fix/885-find-channels` from `origin/main` 51350ca1.
- **User-role TEXT records, all tiers**, measured 2026-09-11 with `jq` over every `*.jsonl` under
  `~/.claude/projects` (records of `type: user` whose content is a string or has a text block;
  by tier × `origin.kind` × `isMeta` × `isCompactSummary` × interrupt; saved in
  `~/.claude/scratch/restart-probe/user-records-by-tier.txt`; re-run by the round-2 auditor, small
  drift from live sessions):

  | Tier | origin.kind | isMeta | other | n | Speaker (§III.1) |
  |---|---|---|---|---|---|
  | top | human | – | | 605 | user |
  | top | task-notification | – / meta | | 503 / 8 | notification |
  | top | peer | meta | | 54 | peer |
  | top | none | meta | | 70 | harness |
  | top | none | – | compact summary | 17 | harness |
  | top | none | – | | 158 | user (remainder, criterion 5) |
  | top | none | – | interrupt | 79 | user (the human's) |
  | subagent | none | meta | | 788 | harness |
  | subagent | none | – | | 318 | lead |
  | subagent | none | – | interrupt | 16 | lead |
  | subagent | coordinator | meta | | 24 | lead |
  | subagent | task-notification | meta | | 17 | notification |
  | workflow | none | meta | | 176 | harness |
  | workflow | none | – | | 54 | lead |
  | workflow | none | – | interrupt | 2 | lead |

  `origin.kind` values seen: `human`, `task-notification`, `peer`, `coordinator` — no others. The
  54 peer turns are `isMeta: true`, so `origin.kind` is read BEFORE `isMeta`. The subagent and
  workflow tiers carry no `human` origin.
- **The 158 top-level remainder** (measured by the round-2 auditor, `~/.claude/scratch/plan-audit-885/`):
  132 prompts with `promptSource: "sdk"` that programs sent to headless sessions (e.g. "Red lens
  sitting…", "Blue lane 1 of 3…"), 15 slash-command records (`<command-name>`), 11
  `<local-command-stdout>` outputs. They span 2026-08-15 to 2026-09-10 on current clients, so they
  are NOT pre-field records. No field separates them from the human: 601 human-origin records also
  carry `promptSource: "sdk"`. Classifying them would need text matching, which the scope excludes.
- **`queued_command` attachments, all tiers** (mid-turn deliveries), measured 2026-09-11:

  | Tier | attachment.origin.kind | attachment.commandMode | n | Speaker |
  |---|---|---|---|---|
  | top | none | task-notification | 341 | notification |
  | subagent | none | task-notification | 143 | notification |
  | top | human | prompt | 53 | user |
  | top | peer | prompt | 51 | peer |
  | subagent | coordinator | none | 7 | lead |

  So every attachment is classifiable by field; none is left `?` on today's corpus. Peer messages
  also appear inside `queue-operation` records (e.g. 40 `enqueue` records in session `3ac18039`);
  those are queue bookkeeping, have no `message` or `attachment`, and stay `?`.
- **The #885 proposal's detector would have missed every peer turn:** they begin "Another Claude
  session sent a message:\n<cross-session-message …", not `<cross-session-message` (issue comment
  posted 2026-09-10). Classification here reads `origin.kind`, not text.
- **Where the conflation lives.** `decodeBlocks` (`project.go:216-230`) turns plain-string user
  content into a synthetic text block; `DecodeHit` labels a text block by `roleChannel(m.Role)`
  (`hit.go:71,90-95`); `Project` stores `Word{Role: m.Role}` (`project.go:166`), inserted at
  `ingest.go:185-187`. The `record` struct (`project.go:15-24`) reads no `origin`, `isMeta`,
  `isCompactSummary` or `attachment` today. Neither `DecodeHit` nor `Project` knows which tier a
  record's file is in; `IngestFile` does (`tf.AgentID`, non-empty for subagent and workflow files),
  and `find`'s `decodeHits` does (`r.f.AgentID`).
- **`--in` today.** `decodeHits` (`find.go:262-304`) decodes at most `samplesPerFile` (200) of a
  transcript's newest matching records, keeps the newest as `best`, and sets `r.capped` when the
  cap bites; `keepChannel` then filters on `best.Channel`.
- **Names, checked 2026-09-11.** `meta` is taken — it is the catalogue's own bookkeeping TABLE
  (`schema.go:131`, `open.go:133`) and the prefix of `MetaRebuiltAt`/`MetaRetainedOn`/
  `MetaBackfilledAt` — so harness-injected text is `harness`, a word gray-area uses only in prose
  for the Claude Code client, which is who injects it. `unknown` is taken — it is `agents`'
  liveness value (`catalogue.Unknown`, `liveness.go:18`) — so an origin kind the table does not know
  is `unknown_origin`, spelled like `tool_use`. `peer` and `notification` appear only as prose with
  the same meaning; `lead` only as a verb in prose (`inspect.go:429`). No existing identifier is
  named `Speaker*`, `RecordFacts`, `unknown_origin` or `"harness"` (`grep -rn -w -E
  'Speaker[A-Za-z]*|unknown_origin|RecordFacts|"harness"' plugins/gray-area` → nothing).
- **Shape.** A value change in `v_word.role` needs a reprojection to relabel old rows, which #894's
  rebuild provides for a stamp below `UserVersion`.

## III. Proposed Changes

```
plugins/gray-area/
├── README.md                                   [MODIFY] :96-100 channels and --in; v_word.role values; the remainder; attachments not in v_word
├── skills/telepathy/SKILL.md                   [MODIFY] :36-42 the same
└── tools/
    ├── internal/catalogue/speaker.go (+test)   [NEW] Speaker, RecordFacts, SpeakerOf
    ├── internal/catalogue/project.go           [MODIFY] record gains Origin/IsMeta/IsCompactSummary/Attachment; Word gains OriginKind/Meta; Project's signature unchanged
    ├── internal/catalogue/project_test.go      [MODIFY] a peer turn's Word carries OriginKind "peer" and Meta true; a human prompt's carries OriginKind "human"
    ├── internal/catalogue/ingest.go            [MODIFY] word role = SpeakerOf(..., tf.AgentID != "") at insert (:185-187)
    ├── internal/catalogue/ingest_test.go       [MODIFY] stored roles: peer → peer, subagent seat prompt → lead, isMeta → harness, top-level human → user
    ├── internal/catalogue/hit.go               [MODIFY] new channels; Channels; DecodeHit(line, term, inSubagent); attachments; comments :8-14, :19
    ├── internal/catalogue/hit_test.go          [NEW] every speaker's text hit (top and subagent), every attachment shape, an attachment whose term is only in its cwd → ?, queue-operation → ?
    ├── internal/catalogue/schema.go            [MODIFY] UserVersion 4 + comment; word.role comment (values)
    ├── internal/catalogue/open_test.go         [MODIFY] every stamp-4-as-newer site → UserVersion+1 (§III.5)
    ├── internal/telecli/find.go                [MODIFY] --in any-hit; capped; help (:40-60, :182); empty wording (:170); DecodeHit call (:329); render IN width (:340, :354)
    ├── internal/telecli/corpus_test.go         [MODIFY] fixture records for each speaker, the attachment shapes, an early-user/late-assistant transcript
    ├── internal/telecli/golden_test.go         [MODIFY] new find and sql cases
    └── internal/telecli/testdata/golden/       [NEW] find-channel-peer, find-channel-notification, find-channel-lead, find-channel-harness, find-in-any-hit, sql-word-roles; [MODIFY] help-find, find-in-channel, sql-older-shape (reads shape 4), and every golden the fixture additions change
```

### III.1 `catalogue.SpeakerOf`

```go
type Speaker string
const (
    SpeakerHuman         Speaker = "user"           // the human (plus the §II remainder)
    SpeakerAssistant     Speaker = "assistant"
    SpeakerPeer          Speaker = "peer"           // another session's message
    SpeakerNotification  Speaker = "notification"   // a background task's notification
    SpeakerLead          Speaker = "lead"           // the lead, or a workflow coordinator, prompting a seat
    SpeakerHarness       Speaker = "harness"        // text the client injects: isMeta, compaction summary
    SpeakerUnknownOrigin Speaker = "unknown_origin" // an origin.kind, or a queued_command commandMode, this binary does not know
)
type RecordFacts struct {
    Role, OriginKind string
    CommandMode      string // attachment.commandMode, for queued_command attachments; "" otherwise
    IsMeta, IsCompact bool
    InSubagent       bool   // the record's FILE is a subagent or workflow transcript
}
func SpeakerOf(f RecordFacts) Speaker
```

Precedence, first match wins:

1. `Role == "assistant"` → `assistant`.
2. `OriginKind` non-empty → `human` → `user`; `peer` → `peer`; `task-notification` →
   `notification`; `coordinator` → `lead`; anything else → `unknown_origin` (loud: its own channel
   and role, never folded into `user`).
3. `CommandMode == "task-notification"` → `notification` (the attachment form, §II).
4. `CommandMode` non-empty and not `prompt` → `unknown_origin` — loud, like step 2: a renamed or
   new mode must not fold a mid-turn delivery into `user`. (`prompt` is the mode human and peer
   attachments carry; their `origin.kind` has already decided them at step 2.)
5. `IsMeta || IsCompact` → `harness`.
6. `InSubagent` → `lead`.
7. Otherwise → `user` — the human, and the §II remainder (criterion 5).

`speaker_test.go` asserts every row of both §II tables maps as shown, plus an unknown
`origin.kind` → `unknown_origin` and an unknown `commandMode` with no origin → `unknown_origin`.

### III.2 The word tier

- `record` (`project.go:15`) gains `Origin *struct{ Kind string \`json:"kind"\` } \`json:"origin"\``,
  `IsMeta bool \`json:"isMeta"\``, `IsCompactSummary bool \`json:"isCompactSummary"\``.
- `Word` gains `OriginKind string`, `Meta bool` (isMeta or compact); `Role` stays the MESSAGE role,
  so `project_test.go:76`'s `w.Role == "assistant"` holds unchanged.
- `IngestFile` inserts `string(SpeakerOf(RecordFacts{Role: w.Role, OriginKind: w.OriginKind,
  IsMeta: w.Meta, IsCompact: w.Meta, InSubagent: tf.AgentID != ""}))` as `role` (`ingest.go:187`).
- The word tier stores turns, not `queued_command` attachments: `find` reports attachments, `v_word`
  has no rows for them. On this box, per §II's attachment table (2026-09-11; the counts drift
  with live sessions — the round-3 auditor's re-run gave 486 notifications): 104 human/peer prompts,
  7 coordinator (`lead`) messages and 484 notifications delivered mid-turn. Stated as a limit in
  README and skill.

**Census — `Word`/`Role`** (`grep -rn -E '\.Role\b|Role:' plugins/gray-area/tools --include='*.go'`,
run 2026-09-11): `hit.go:71` (reads `m.Role`; changed in §III.3), `project_test.go:76` (reads
`w.Role == "assistant"`; unchanged because `Word.Role` stays the message role), `project.go:166`
(the write; gains the two fields), `ingest.go:187` (the insert; now `SpeakerOf`). `Project`'s
signature is unchanged (callers `ingest.go:154`, `project_test.go` ×7, `project_order_test.go:34,55`
untouched).

### III.3 `find` channels

- `hit.go`: `ChannelPeer`, `ChannelNotification`, `ChannelLead`, `ChannelHarness`,
  `ChannelUnknownOrigin` (`"unknown_origin"`) join `Channels`, so `ValidChannel` and `channelList`
  include them. For text, a channel IS the speaker: `Channel(SpeakerOf(...))`.
- **`?` and `unknown_origin` are different answers, and the help says which is which:** `?` means
  the term is in a part of a record nothing here models (a cwd, a uuid, queue bookkeeping);
  `unknown_origin` means a record whose `origin.kind`, or whose `queued_command` `commandMode`,
  this binary does not know — upgrade gray-area. Neither is the liveness word `unknown` that
  `agents` prints. The README and skill carry these two definitions verbatim.
- `DecodeHit(line, term string, inSubagent bool)` — contract change. **Census**
  (`grep -rn 'DecodeHit(' plugins/gray-area/tools`, run 2026-09-11): `hit.go:54` (definition),
  `find.go:329` (the only caller, inside `recordsAt`; `decodeHits` passes `r.f.AgentID != ""`). No
  test calls it.
- An `attachment` record with `attachment.type == "queued_command"` gets a speaker channel ONLY
  when the term is in `attachment.prompt` (`contains(attachment.prompt, term)`), exactly as a
  message record gets one only when a text block contains the term (`hit.go:65-67`): it is then
  classified by `SpeakerOf` from `attachment.origin.kind`, `attachment.commandMode` and
  `InSubagent`, snippet from `attachment.prompt`. A term found only elsewhere in the record — every
  such attachment also carries `cwd`, `gitBranch`, `sessionId`, `uuid` (measured by the round-4
  auditor over all 597) — returns `?` with a `window(line, term, 120)` snippet, as `hit.go:84-86`
  does for messages. Any other record with no `message` stays `?`.
- **The IN column's width** (`find.go` `render`, header :340 and rows :354, today `%-9s`, sized for
  `tool_use`) is DERIVED from the longest name in `catalogue.Channels` (14, `unknown_origin`), for
  the header and every row, so a channel added later cannot misalign SNIPPET again. `%-9s` pads but
  never truncates, which is why a hand-kept width would silently ragged-edge the table.
- Comments: `hit.go:8-14` (a task notification was "not somebody saying something" — it now has
  its own channel) and `:19` (`ChannelUser` is the human; tool results already report `result`).

**Census — channels and `--in`** (`grep -rn -E 'catalogue\.Channels|ValidChannel|channelList|--in\b'
plugins/gray-area README.md --include='*.go' --include='*.md' --include='*.golden'`, run
2026-09-11): `find.go:51,57,158,161-162,230-232`; `hit.go:29,33-34`; `golden_test.go:313,315,317,338`;
`help-find.golden:12,18,43`; `plugins/gray-area/README.md:100`; `skills/telepathy/SKILL.md:37`. Each
is updated or regenerated per §III.4; `golden_test.go:338` (`--in asistant` refused) is unchanged.

### III.4 `find --in` any-hit, and its prose

- With `--in C`, `decodeHits` decodes ALL of a transcript's matching records and sets no `capped`,
  and the row's `best` is the most recent hit whose channel is C; a transcript with none is dropped.
  HITS is unchanged. Without `--in`, behaviour is unchanged (newest of the newest 200; the footer
  when the cap bites).
- Empty result: `%d transcript(s) contain %q, but none has a hit in %q`.
- `keepChannel` is removed (`grep -rn keepChannel plugins/gray-area/tools` → `find.go:165,244`; no
  test calls it).
- Help, `find.go:40-60` Long: the IN line lists every channel; the paragraph at :51-55 that calls a
  task notification a non-speech hit is rewritten — notifications, lead prompts and harness text
  have their own channels; `user` is the human plus the §II remainder (programs' prompts to headless
  sessions, slash-command and local-command text); `?` versus `unknown_origin` as in §III.3; `--in`
  keeps a transcript with ANY hit in the channel. `:182` flag text likewise.

### III.5 Shape 4, and the stamp census

`UserVersion` 3 → 4; its comment records why. `word.role`'s column comment lists its values.
**Census — stamp literals** (`grep -rn -E 'UserVersion|user_version *= *[0-9]|stamped [0-9]|currentAt\(t, *[0-9]+\)|stamp 4' plugins/gray-area`,
run 2026-09-11) — the sites that assume 3 is current or 4 is newer:

| Site | Change |
|---|---|
| `catalogue/schema.go:169` | `UserVersion = 4` |
| `catalogue/open_test.go:499` (`"newer: stamp 4, today's tables"`, `currentAt(t, 4)`) | → `currentAt(t, UserVersion+1)`; name without the literal |
| `catalogue/open_test.go:500-501` (`"newer: stamp 4 with an unknown table"`, `currentAt(t, 4)`) | Same |
| `catalogue/open_test.go:507-508` (`"newer: stamp 4 with a session column renamed"`, `currentAt(t, 4)`) | Same |
| `catalogue/open_test.go:520` (`"foreign: stamp 4 without file_offset"`, stamp 4) | → `UserVersion+1`; name updated |
| `catalogue/open_test.go:703-704` (`build: currentAt(t, 4)`, `refuse: "is stamped 4 by a newer gray-area"`) | → `UserVersion+1`, refusal text built from it |
| `catalogue/open_test.go:951` (`currentAt(t, 4)` as newer, and its expected refusal `"is stamped 4 by a newer gray-area (this binary writes %d)"`) | → `currentAt(t, UserVersion+1)`, and the refusal text built from `UserVersion+1` |
| `telecli/testdata/golden/sql-older-shape.golden:4` ("reads shape 3") | Regenerated ("reads shape 4") |
| Every other hit — `open.go` (symbolic `UserVersion`), `open_test.go:195,205` (stamp 99), `:494` (symbolic), `:495-496,516,519,787-788,950` (historical stamps 1–2), `:800,806-807` (symbolic), `golden_test.go:65,118`, `shape1.sql`, `agents-rebuilt-warning.golden` (stamps 1–2), `schema.go:33,160` (prose) | No change |

### III.6 Carrier census (complete-the-concept)

| Carrier | Change |
|---|---|
| `catalogue/speaker.go` | New |
| `catalogue/hit.go` `Channels`, comments :8-14 and :19, `roleChannel`, `DecodeHit` | As §III.3 |
| `catalogue/project.go` `record`, `Word` | As §III.2 |
| `catalogue/ingest.go:187` | `SpeakerOf` |
| `catalogue/schema.go` `UserVersion`, `word.role` comment | As §III.5 |
| `telecli/find.go` help, `decodeHits`, `keepChannel`, empty message, `recordsAt` | As §III.3–III.4 |
| `telecli/find.go` `render` IN column (:340 header, :354 rows) | Width derived from `catalogue.Channels` (§III.3); every find golden re-renders |
| `help-find.golden`, `find-in-channel.golden`, `sql-older-shape.golden` | Regenerated |
| `plugins/gray-area/README.md:96-100` | Channels, `?` vs `unknown_origin`, `--in`, `v_word.role` values, the remainder, attachments not in `v_word` |
| `skills/telepathy/SKILL.md:36-42` | The same |
| Root `README.md` | No channel list or role values there (checked) — unchanged |
| `plugin.json` | Unchanged |

## IV. Risk & Mitigation

| # | Risk | L × I × cost | Mitigation |
|---|---|---|---|
| R1 | The client adds or renames an `origin.kind` or `commandMode` | low × medium × low | Either unknown value → `unknown_origin` (§III.1 steps 2 and 4: loud, never `user`); `speaker_test.go` pins every known kind and mode and one unknown of each |
| R2 | Queries written as `WHERE role='user'` change results | certain × low × none | They now return the human plus the stated remainder — the correct answer; README and skill state the values |
| R3 | Rebuild again on this box | certain × low × none | Accepted (gblock); one backfill; consumers rebuild once at release (#894 unreleased) |
| R4 | `--in` on a common term decodes many records | medium × low × low | One sequential pass per transcript (`recordsAt`); unfiltered view keeps its cap |
| R5 | `user` still includes programs' prompts to headless sessions and slash/local-command text | certain × low × none | Stated with the measured composition (criterion 5); no field separates them — `promptSource: "sdk"` is also on 601 human records — and text matching is outside the scope |
| R6 | Interrupts in subagent transcripts are labelled `lead` | certain × low × none | 18 records; harness text in a seat's transcript, no field distinguishes them; stated |
| R7 | `v_word` omits mid-turn deliveries that `find` reports | certain × low × none | Stated limit; the word tier has never stored attachments |

## V. Verification Plan

1. **Unit and race** (re-arms on any change under `plugins/gray-area/tools`):
   `go -C plugins/gray-area/tools vet ./...`; `go -C plugins/gray-area/tools test -count=1 ./...`;
   `CGO_ENABLED=1 go -C plugins/gray-area/tools test -count=1 -race ./...`. New tests:
   `speaker_test.go` — every row of both §II tables, an unknown `origin.kind` → `unknown_origin`, an unknown `commandMode` with no origin →
   `unknown_origin`, assistant;
   `hit` cases for each speaker's text hit (top and subagent), and attachments: peer, human,
   task-notification by `commandMode` with no origin (top and subagent), coordinator in a subagent
   file; an attachment whose term appears only in its `cwd` → `?`; a `queue-operation` record → `?`; an ingest test that a peer turn's word stores role
   `peer`, a subagent seat prompt `lead`, an isMeta record `harness`, and a top-level human prompt
   `user`.
2. **Goldens** (re-arms on `telecli` output): `go -C plugins/gray-area/tools test -count=1
   ./internal/telecli -run TestGolden` — `find-channel-peer`, `find-channel-notification` (including
   a mid-turn notification attachment, its 12-character channel aligned under the widened IN header
   with SNIPPET starting in the same column on every row), `find-channel-lead`, `find-channel-harness`, `find-in-any-hit`
   (a term the human said first and an agent repeated later: `--in user` keeps the row and shows the
   human's hit, with no cap footer), `sql-word-roles` (`SELECT role, count(*) FROM v_word GROUP BY 1
   ORDER BY 1`), and the regenerated `help-find`, `find-in-channel`, `sql-older-shape`; every golden
   the fixture additions or the IN-width change re-render — including `find-hit`,
   `find-channel-assistant`, `find-channel-thinking` and `find-channel-user` — is regenerated and read.
3. **Surfaces**: `go -C scripts run ./pluginparity`; `go -C scripts run ./frontmatter`;
   `go -C scripts run ./rulesweep -base origin/main` and `go -C scripts run ./archaeology -base
   origin/main` (the skill changes; commits carry `Rule-Class` / `Sibling-Sweep`).
4. **Real data**: build `telepathy`; copy `~/.claude/scratch/restart-probe/shape-check/abs/catalogue.db`
   (a shape-3 store of this box's corpus) to a fresh scratch path `$R`; `telepathy --store $R
   backfill` MUST rebuild it once (stamped 3 → 4). Then, on `$R`:
   - `sql "SELECT role, count(*) FROM v_word WHERE text LIKE 'Another Claude session sent a message:%' GROUP BY 1"` → `peer` only;
   - `sql "SELECT count(*) FROM v_word WHERE role='user' AND text LIKE 'The coordinator sent a message%'"` → 0;
   - `sql "SELECT count(*) FROM v_word WHERE role='user' AND agent_id != ''"` → 0;
   - `sql "SELECT role, count(*) FROM v_word GROUP BY 1 ORDER BY 1"` — recorded in the PR;
   - `find 'conduct section' --in user` MUST NOT return the #885 repro row; `--in peer` MUST.
5. **Auditor gate**: `/prosthetic-conscience:plan-audit plans/find-speaker-channels.md` → PASS.
