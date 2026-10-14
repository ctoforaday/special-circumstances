# An avenue becomes an inquiry: the noun, its question, its fates and the act that records them

> STATUS: **PROPOSED 2026-10-13; every fork ruled (§III.1, gblock, 2026-10-13).** It awaits the plan-auditor gate.
> No code changes with this file.

`F` is `plugins/frank-exchange-of-views`, `T` is `F/tools`. Every census is a command run from the repository
root after the four assignments under S1; a command that begins `go -C` names its own directory. Line numbers and
counts are at `e8f6ec9c` (`origin/main`, 2026-10-13, event-schema epoch 25); the implementer re-runs each census
on the pull request's base. `L/F` is lines / files. Archive figures come from
`python3 -I ~/.claude/scratch/inquiry-rename/plan/final.py <dir>` over the 20 tarballs of `run-archive/` extracted
to `<dir>` (its first draft is `~/.claude/scratch/inquiry-rename/scan.py`). It is not committed: the PR pastes the
script and its whole output — per archive the act and final-state counts, the totals, the `supersedes_status`
values, the proposals whose `reason` equals their `line`, and the rows ending `concluded` whose reason holds
`refut|falsif`.

## I. Summary & Goals

**The problem.** Blue records each approach the research could take, and what became of it. Five names on that
concept say something other than what the thing is.

| Today | What it fails to say | Evidence |
|---|---|---|
| `avenue` (the noun) | It holds a question, a hypothesis and a finding, and names none. The report labels its text `**Question:**` and its id letter is Q | `T/internal/flags/shapes.go:174`; `plans/rulebook-audit.md:109–111` |
| `line`, filled by `--reason` | That it is a question. `--reason` names two different texts on the two acts of one group, and `propose` writes its question into the field `reason` as well | 275 of 275 archived proposals hold `reason == line` (`final.py`) |
| `concluded` | Whether the answer was yes | on universe-m22 two seats read it two ways and the dispute held the gate (`T/internal/record/recordpb/record.proto:432–435`) |
| `pursued` | That it is still open: the one running state reads as a finished past tense | a prompt has to explain it (`F/skills/research-protocol/scripts/debate.js:218`) |
| `avenue move` | What re-recording the same state is: `move --as pursued` on a pursued avenue moves nothing | `T/internal/cli/seat/help/move.md:7` |

**The objective.** One noun, **inquiry**; its text is its **question**; it is **underway** while open, and
**confirmed** or **refuted** when it ran and answered; the act that says what became of it is **`inquiry record`**.
After one pull request the live tree accepts and prints only these words, at epoch 26, with one translation in
`migrate` per changed word.

### Success criteria

| # | Criterion | Command | Today | Target |
|---|---|---|---|---|
| S1 | No old word in the live tree | under this table | noun 2,247 lines / 226 files; fate words 366 / 116 | noun: no output; fate words: the English list, byte for byte |
| S2 | Inside `migrate`, an old word is a translation key | under this table | 36 lines | no output |
| S3 | Every archive migrates | `go -C T test -count=1 -run 'TestEveryArchivedRunRenders\|TestEveryArchivedRunMigrates' ./internal/record/migrate/` | 20 runs, 0 refusals at 25 | 20 runs, 0 refusals at 26 |
| S4 | Assembled archive documents differ only in the renamed words | §V.3 | — | the mapped diff prints nothing |
| S5 | A record at epoch 25 is refused, then migrates | §V.4 | — | the refusal names 25, 26 and `migrate`; 0 refusals after |
| S6 | No translation is unexercised | §V.2 step 3 | — | every deleted registry row fails a named test |
| S7 | Gates | `go -C scripts run ./check`; the release fuzz (§V.5) | — | fail count 0; `fuzzed 40 debate runs · 0 failed` |
| S8 | The engine runs | §V.6, before merge | — | a VERIFIED or CEILING outcome whose report set holds `inquiries.md` |

S1 — over every tracked file outside §III.2's not-the-live-tree list:

```
F=plugins/frank-exchange-of-views; T=$F/tools
NOT='^(plans|law|research|run-archive)/|^feov-memory/[a-z]+-scorecard\.md$'
NOT="$NOT|^$T/(research|internal/record/migrate|internal/tessocr/testdata/corpus)/"
GEN="^$F/agents/|\.golden$|\.pb\.go$|/testdata/|^$F/docs/vocabulary\.md$|^$F/skills/(research-protocol|adversarial-audit)/SKILL\.md$"
git grep -lIi 'avenue' | grep -vE "$NOT"                           # the noun: no output
git grep -IiE 'conclud|pursu' | grep -vE "$NOT" | sort             # the fate words and their inflections
```

The noun's grep returns nothing: a plain-English use is reworded too (§III.6). The fate pattern is a substring, so
it takes the enum identifier, the quoted word, the bare word in a sentence and the verb's other forms alike: 366
lines today — 134 generated, 131 in tests, 101 authored — of which 111 hold a fate as an identifier, a quoted word
or a value after `--as`, and the rest are prose. **Every one is read (N14).** A line that names the fate is
renamed; plain English ("the lens concluded that…") stays exactly as it is. What the second command prints after
the rename is the English list: each line marked English by its reader, pasted in the PR, never committed. S1
holds when `diff` of the command's output against the list in the PR prints nothing. The output carries no line
numbers, so a merge that only moves a line leaves it equal; a line the diff prints arrived or left with that
merge, and is read before it is renamed or joins the list. The excluded set — `grep -E "$NOT"` in place of
`grep -vE` — holds those path prefixes and nothing else.

S2 — inside `migrate` an old word is a string literal: a key a translation maps from, or in a test the archived
input and the planted row that feed one. It is never an identifier, a file name or a comment:

```
git grep -nIiE 'avenue|conclud|pursu' -- $T/internal/record/migrate | grep -vE '"[^"]*(avenue|conclud|pursu)[^"]*"'
```

S1 and S2 are run by hand, at landing and on every merge attempt (§III.9). Neither is committed as a gate.

### The value case

- **What it buys.** The five rows above stop being taught. Two have a measured cost: a gate held on universe-m22 by
  two readings of one fate, and one flag meaning two texts in one group. The report already prints "Question /
  Hypothesis / Fate"; the record, the verbs, the flag and the prompts then say the same words.
- **What it costs.** The noun is on 2,247 lines in 226 files of the live tree: 675 regenerate and about 1,570 are
  edited by an identifier rename or by hand, 913 of them in 91 test files outside `migrate`; one review of a diff
  near 4,000 lines each way. Reading no tool shortens: 366 fate-word lines, and 1,026 lines holding `line` and 589
  holding `move` in the files that name the concept (§III.3). The rationale comments that narrate a run or an
  incident in the old words are deleted from 37 files (§III.2): that reasoning then lives in the commit history
  and nowhere in the tree. One epoch: every live run and scratch record is refused until migrated. The landing:
  two or three merge attempts and one or two smoke runs of about 45 minutes, two to four hours from the finished
  tree (§III.9). The enum-rendered lists (§III.6) add about 30 lines of Go and one generated constant, and remove
  seven hand-kept lists.

## II. Technical Context

Go 1.27 (`GOTOOLCHAIN=local`), one module at `T`; Node for the engine script and simulator. The record is SQLite,
append-only.

- **Schema and epoch.** `T/internal/record/recordpb/record.proto` is the source: message and field names become
  table and column names, enum values become the words stored and typed, `(means)` becomes help text and the
  `enum_*` tables. The stored views are hand-written SQL in `T/internal/record/recordsql/views.go`. The epoch
  (`F/requirements.json`, `eventSchema: 25`) is compared at `setup` and at every open of a database the binary did
  not create; a mismatch either way is refused with a pointer to `migrate`.
- **Migrate** replays an old record through the current write path. Its registry is keyed by the old word, one hop
  to the current word (`Entries()`, `T/internal/record/migrate/entries.go:15`; `eraEntries()`, `eraentries.go`);
  `renamedBodyTables` (`sqlitesource.go:228`), `valueRenames` (`valuerenames.go:19`, keyed by the current enum's
  full name) and `avenueArm` (`eraentries.go:211`) carry the other renames. A migrated id is
  `sha256(source hash ‖ archived id)` under the kind's letter (`remap.go:173–182`), independent of the kind's name.
  Migrate reads tables, never views.
- **Generators.** Each is re-run, never hand-resolved.

  | Command | Source → output | Check |
  |---|---|---|
  | `go -C scripts run ./protogen` | `record.proto` → `record.pb.go`, `record.proto.sha256` | `git diff --exit-code` after a re-run |
  | `go -C scripts run ./schemagen` | `F/requirements.json` → `recordpb.EventSchema` | `schemagen -check` |
  | `go -C T run ./devcmd/enginegen` | roster, enums, `record.Plan` → `debate.js:106–125`, `:357–411` | gate `enginegen` |
  | `go -C scripts run ./agentgen` | `scripts/agentgen/src/**`, the tool's `manual` → `F/agents/*.md` (11), two `SKILL.md` | `agentgen -check` |
  | `go -C scripts run ./vocabdoc` | `T/internal/terms/terms.json` → `F/docs/vocabulary.md` | gate `vocabdoc` |
  | `UPDATE_GOLDENS=1 go -C T test -count=1 <pkg>` | the binary → `difftest/testdata/*.golden` (22), `recordsql/testdata/schema.sql`, `recordpb/testdata/payload-keys.txt`, `dashboard/testdata/*.golden` | gate `golden`; `go -C scripts run ./golden -review` |
  | `UPDATE_GOLDENS=1 node --test plugins/frank-exchange-of-views/tests/simulator/*.test.mjs` | `debate.js` → `F/tests/simulator/testdata/prompt-*.golden` (14) | the same without `UPDATE_GOLDENS` → `fail 0` |

- **Gates that bear on the rename** (under `T`). `TestEveryEnumValueNamedInAPromptIsAccepted`
  (`integration/surface/promptverbs_test.go:926`) fails a `--<flag> <value>` no enum bound to that flag holds.
  `TestVocabularyProse` (`integration/surface/vocabulary_prose_test.go:47`) runs the terms registry over
  `vocabularySources(t)`; it fails a banned variant and a mask or allow that did no work, and requires no entry
  for any word. `TestNoSeatPromptGrowsPastItsCeiling` (`internal/seatprobe/promptsize_test.go:92`) counts bytes for
  four seats. `archaeology` reads ADDED lines of agent-facing files, so a renamed line re-arms it on wording the
  line already held; `rulesweep` needs its two commit trailers; `decisions` reads
  `F/docs/seat-command-triggers.md`; `deadcode` fails a function nothing reaches.
- **CI.** The `hooks` workflow on a pull request takes a median 5.5 minutes from push to every check done, maximum
  6.1 (22 green runs, 2026-10-12T12:06Z to 2026-10-13T06:50Z: `gh run list --event pull_request --limit 60 --json
  createdAt,updatedAt,conclusion`). Its longest job runs `go test ./...` whole on `T`, the race leg over the tests
  `scripts/racescope` selects and one raced fuzz test (`gh run view 38353666439 --json jobs`); Windows runs
  `-short`. The release fuzz is not in it.
- **Constraints.** CLAUDE.md, Change discipline: the live code accepts and prints only the new form; `migrate` is
  the only upgrader. Versions do not move in this PR. No network, secret or privilege is involved. `gopls` is not
  installed on the development box, and installing it writes outside the working tree: no step uses it.

## III. Proposed Changes

### III.1 Rulings (gblock, 2026-10-13)

| # | RULED |
|---|---|
| N1 | The noun `avenue` → **inquiry**: `show inquiries`, `inquiries.md`, tables, events, keys and the chair's acts follow |
| N2 | The proposal's text `line` → **question**: flag `--question`, field `question` |
| N3 | The act `avenue move` → **`inquiry record`**: one verb, every state, `--as <fate> --reason …` |
| N4 | The fates `concluded` → **confirmed**, `pursued` → **underway**; `refuted`, `abandoned`, `declined`, `deferred` unchanged |
| N5 | Kept: the verb `propose` and the state `proposed`; the fate's prose flag `--reason`; the id letter `Q-`; `--hypothesis`, `--method`; the three report headings; `## Open questions`; the title's `**Question:**` line; the grade-motion ruling `rejected`; the `inquest` verb |
| N6 | ONE pull request, epoch 25 → 26, one translation per changed word in `migrate`, each tested on an archived run |
| N7 | Every occurrence of an old word in the live tree is renamed. A passage whose only job is to say what something was called, or to narrate a run or incident in the old words, is DELETED; its why goes in the commit. What is not the live tree is §III.2's list and nothing else |
| N8 | The terms registry defines `inquiry` and holds no ban, no note about an earlier name and no new collision note. No word is added to any roster of retired names |
| N9 | The forward check is two things and no more. A test fails an `--as` value that is not a fate of its verb, read from the enum, per verb. Where agent-facing text ENUMERATES the fates, the list is rendered from the enum. No check judges a backticked word in prose (§III.6) |
| N10 | A proposal holds a question and NO reason: `propose` stops filling `reason`, migrate drops the copied text (M15), the report's guard on the copy is deleted, the thin-reason detector skips `proposed` |
| N11 | The scorecard metrics are `inquiries` and `thin_inquiry_reasons` |
| N12 | The archive before/after harness is committed as `T/devcmd/archdump`, in the PR's first commit |
| N13 | No freeze. The rename lands by merge-main-in, regenerate, re-verify, each time main moves (§III.9). Merging on green CI is the owner's standing instruction for this workstream |
| N14 | READ EVERY ONE. The sweep takes every occurrence of each old fate word in the live tree; each is read and classified. A fate is renamed; plain English stays as written and is not reworded. The classification is pasted in the PR and never kept in the tree (S1) |
| N15 | THE SMOKE RUNS BEFORE MERGE. A universe is built from the finished branch and a smoke reaches an outcome before the PR merges (§V.6, §III.9) |
| N16 | `F/tests/simulator/testdata/m16-chair6-relay.json`, a captured chair relay, is a TEST INPUT in the live tree and not a record: its text is rewritten in the new words like everything else |

**The naming census the rulings rest on.** `inquiry`:
`git grep -nIi inquir -- $T $F/skills $F/agents scripts/agentgen/src` → 35 lines in 11 files — `migrate` keys
(identity after §III.5), one narrated incident (deleted, N7) and two registry bans (deleted, N8). `question`:
`git grep -nIE '"question"|--question|flags\.Question' -- "$T/*.go" ':!*_test.go'` prints nothing — no flag,
field, JSON key or enum value; the run's brief line and `## Open questions` are kept by N5. `underway`: `git grep -ci underway -- $F
scripts` → 0. `confirmed`: plain English only. `record` as a verb: `git grep -n 'Use: *"record' -- $T` → 0; the
noun is the tool and the store (R5). `inquest`: a verb group on the bench and the chair, 67 lines in 18 Go files
(`git grep -nIi inquest -- "$T/*.go"`; R5).

### III.2 Carrier census (the consumer census for every contract below)

Per row: `git grep -niI avenue -- <path> | wc -l`, and `-l` for files. **Complete when S1's noun grep prints
nothing and the noun census at the base surfaces no path this table omits.**

| Carrier | Path | L/F | How it changes |
|---|---|---|---|
| Schema, generated code | `T/internal/record/recordpb/record.proto`; `record.pb.go` | 50/1; 205/1 | by hand from §III.4; `protogen` |
| Stored views, schema goldens | `recordsql/views.go`; `recordsql/testdata/schema.sql`, `recordpb/testdata/payload-keys.txt` | 20/1; 46/2 | by hand (2 views); regenerate, read |
| Migrate | `T/internal/record/migrate/*.go` | 28/6 non-test | by hand, §III.5; old words stay only as literals (S2) |
| Record folds | `T/internal/record/*.go` non-test | 210/22 | identifier rename, then literals and comments |
| CLI and help pages | `T/internal/cli/**` non-test `.go`; `cli/seat/help/{propose,move,review,dispatch,mint}.md` | 105/17; 15/5 | by hand; `move.md` → `record.md` |
| Report, view, scorecard, consistency, seatprobe, flags | `T/internal/{report,view,scorecard,consistency,seatprobe,flags}` non-test | 145/15 | rename, then literals and comments |
| Other Go | `anchor`, `graph`, `diagnostics`, `refusalcheck`, `reportvoice`, `enumhelp`, `fetchcache` | 1–2 each | by hand |
| Go tests; of them the release fuzz | `*_test.go`; `T/releasegate/fuzz/` | 939/95; 167/3 | compiler, then literals and comments; the fuzz by hand (§III.8 step 8) |
| Engine prompts | `F/skills/research-protocol/scripts/debate.js` | 20/1 | by hand; `:399` and the sections sentence by `enginegen` |
| Constitutions and skills | `scripts/agentgen/src/**` → `F/agents/*.md`, `F/skills/**/SKILL.md`; `references/{report_template,run-directory}.md` | 15/6 → 235/11, 19/4 | sources and references by hand; `agentgen` |
| Goldens, simulator | `T/internal/difftest/testdata/`; `F/tests/simulator/` | 133/22; 69/11 | regenerate (§III.8 step 7); `debate.test.mjs` and the relay fixture (N16; 4 old words) by hand |
| Terms | `T/internal/terms/terms.json`; `F/docs/vocabulary.md` | 7/1; 8/1 | by hand (§III.6); `vocabdoc` |
| Docs, class registry | `F/README.md`; `F/docs/{seat-command-triggers,seat-surface-naming,duty-docket-preregistration,why-a-seat-stops}.md`; `feov-memory/protocol-class-registry.md:110` | 1; 17/4; 1 | by hand |
| Names | 13 files (`git ls-files \| grep -i avenue \| grep -vc '^plans/'`); 30 test functions (`git grep -nIE 'func Test[A-Za-z0-9_]*Avenue'`) | — | `git mv`; by hand |

**Not carriers**: `git grep -ciIE 'avenue|inquir' -- plugins ':!plugins/frank-exchange-of-views'` → 0; the same
over `$F/hooks $T/internal/capture $T/cmd $T/devcmd` → 0.

**Not the live tree** — the whole list; S1 excludes exactly these prefixes.

| Path | Why |
|---|---|
| `T/internal/record/migrate/` — the string keys a translation maps from, and the tests that read archived runs | that is what `migrate` is; S2 holds it to literals |
| `run-archive/` | the gzipped record of each captured run |
| `plans/` (this file included) | dated design records: `plans/README.md` keeps a plan as the record of what changed |
| `law/` | holdings and proposed classes, as ruled |
| `research/`, `T/research/` | the research corpus: reports and records as written |
| `feov-memory/*-scorecard.md` | scorecard history, appended run over run at capture |
| `T/internal/tessocr/testdata/corpus/` | a third party's page as OCR read it (`FIFTH AVENUE BANK`, one line): not this concept, not this repository's text |

**Passages deleted under N7.** The finder —

```
git grep -nIiE 'avenue|inquir|conclud|pursu' -- plugins scripts feov-memory ':!*.golden' ':!*.pb.go' ':!*testdata*' \
    ":!$F/agents" ":!$T/research" ":!$T/internal/record/migrate" ':!feov-memory/*-scorecard.md' \
  | grep -E 'MEASURED|[Mm]easured|20[0-9]{2}-[0-9]{2}-[0-9]{2}|universe-m[0-9]+|#[0-9]{3,}|\b[0-9]+ of [0-9]+\b|verbatim|retired|CUT|used to|became'
```

— prints 59 lines in 37 files. Each is a sentence in the code's present voice, which is renamed, or a narration,
which is deleted whole and never renamed in place. Known members: the incident above `whereItLives`
(`T/internal/cli/root.go:288–297`) and its copy (`refusalteaches_test.go:201–205`; the test keeps its assertion and
takes a current verb as input); the stories in `record.proto:238,432–435`, `T/integration/surface/promptverbs_test.go:20–23,95–97`,
`T/internal/cli/blue/avenue.go:49–53`, `T/internal/record/{avenue.go:17,available.go:65–67}` and
`T/internal/fetchcache/backends.go:863–869`; the measurement tables and the retired or CUT rows of the four `F/docs`
files. An instruction a deleted passage carried is restated in the present tense where it stood (R7).

### III.3 The words grep cannot settle: the fate words, `line` and `move`

A fate word in a sentence is a fate or English; `\blines?\b` is mostly text lines; `\bmov(e|ed|es|ing)\b` is
mostly gaps and grades. Each sweep's file set is closed by a command that selects on the word it sweeps for, and
the three classification tables are pasted in the PR.

1. **The fate words: S1's second command**, 366 lines, classified **fate** (renamed) or **English** (left).
2. **`line` and `move`, each in every authored file that names the concept:**
   `for w in '\blines?\b' '\bmov(e|ed|es|ing)\b'; do git grep -lIi avenue | grep -vE "$NOT|$GEN" | xargs git grep -nIiE "$w" --; done`
   → `line`: 1,026 lines in 132 files (524 in 60 outside tests); `move`: 589 lines in 113 files (291 in 59),
   among them the seven that hold `move` beside the noun and no `line` (`record/{dispatch,refs,required,id,
   convergence_refusal}.go`, `cli/lens/mint.go`, `report/motions.go`). Each `line` is classified **inquiry** (the
   thing: "a declined line"), **question** (its text: "the line you propose") or **unrelated** ("one line on
   why"); each `move` as **record** ("its fate is recorded") or unrelated. A file that says `move` of this concept
   and never names it is outside the set: passes 3 and 4 are what reach it.
3. **The compiler lists the readers.** Renaming `Avenue.Line` and `GetLine()` breaks every site that reads the
   field. At each, the comment above and the literal beside it are read.
4. **The rendered text is read, not its source.** After §III.8 step 6 the eleven regenerated `F/agents/*.md`, the
   two generated `SKILL.md`, the fourteen simulator prompt goldens and the relay fixture are searched for all three
   patterns (`line` 209 lines, `move` 199 today) and every hit classified again: a sentence assembled from two
   constants shows its seam only here.

### III.4 The name table

Numbers, positions and every unlisted name are unchanged. **The rule:** a name holding `avenue`, `Avenue` or
`AVENUE` takes `inquiry` in the same case and position (plural `inquiries`) — the Go names (127 distinct, 54 of
them outside tests:
`git grep -hoIE '\b[A-Za-z0-9_]*(Avenue|AVENUE)[A-Za-z0-9_]*\b|\bavenue[A-Z][A-Za-z0-9_]*\b' -- "$T/*.go" ':!*.pb.go' | sort -u`),
the 13 files, the 30 test functions, and every row below that states no exception.

**[MODIFY] `record.proto` (then `protogen`) and the SQL derived from it.**

| Kind | Old | New |
|---|---|---|
| by the rule: event types, enums, messages, arms, id fields | `EVENT_TYPE_AVENUE = 3`, `EVENT_TYPE_AVENUE_REVIEW = 32`; `AvenueStatus`, `AvenueRuling` and their value prefixes; `MOTION_SUBJECT_AVENUE = 3`, `ABOUT_KIND_AVENUE = 2`; `Avenue avenue = 44`, `AvenueReview avenue_review = 51`, `AvenueMotion avenue = 12`, `MotionRule`'s `AvenueRuling avenue = 12`; `avenue_id = 1` (twice) | the same with `INQUIRY`, `Inquiry`, `inquiry`; stored words `inquiry`, `inquiry_review` |
| by the rule: tables and columns | `avenue`, `avenue_review`, `motion_avenue`, `enum_avenue_status`, `enum_avenue_ruling`; `avenue_id`, `motion_rule.avenue`, `motion_answers.avenue` (`views.go:756`) | the same with `inquiry` |
| enum value | `AVENUE_STATUS_PURSUED = 2` (`pursued`) | `INQUIRY_STATUS_UNDERWAY = 2` (`underway`) |
| enum value | `AVENUE_STATUS_CONCLUDED = 6` (`concluded`) | `INQUIRY_STATUS_CONFIRMED = 6` (`confirmed`) |
| field, column | `Avenue.line = 2`, `(sql).flag: "reason"` | `Inquiry.question = 2`, `(sql).flag: "question"` |
| unchanged | status words `proposed = 1`, `deferred = 3`, `declined = 4`, `abandoned = 5`, `refuted = 7`; the ruling words; fields `hypothesis = 3`, `method = 4`, `status = 5`, `supersedes_status = 6`, `reason = 7` | — |
| view (`views.go:793`) | `avenue_state`, columns `avenue_id`, `line`, `avenue_ruling` | `inquiry_state`, columns `inquiry_id`, `question`, `inquiry_ruling` |
| `(means)` text | 11 annotations: the seven status meanings, the two event types', the motion subject's, the about-kind's | rewritten under §III.3; each is seat-facing help |

**[MODIFY] CLI surface** — each row is a contract; its consumers are the lines of
`git grep -nE 'avenue move|avenue propose|avenue review|motion avenue|show avenues'` — 59 authored
(`-- plugins scripts ':!*.golden' ':!*_test.go' ':!*.mjs' ":!$F/agents" ':!*testdata*'`), 85 in tests
(`-- '*_test.go' '*.mjs'`) and 80 generated — and every one changes.

| Old | New |
|---|---|
| `blue avenue propose --reason <line> [--hypothesis] [--method]` | `blue inquiry propose --question <question> [--hypothesis] [--method]` |
| `blue avenue move --id <Q-…> --as <fate> --reason <account>` | `blue inquiry record --id <Q-…> --as <fate> --reason <account>` |
| `chair avenue review --reason …`; `motion avenue rule --id <Q-…> --as endorsed\|out_of_scope\|too_thin` | `chair inquiry review …`; `motion inquiry rule …` |
| `show avenues` (and `--json`: `{"avenues":[{"id","line",…}]}`) | `show inquiries`; `{"inquiries":[{"id","question",…}]}` |
| `--about-kind section\|avenue\|gap`; `--as pursued\|concluded` | `--about-kind section\|inquiry\|gap`; `--as underway\|confirmed` |
| result keys `avenue_id`, `line`, `moved` (`blue/avenue.go:193–196`; `moved` is read at `:212`) | `inquiry_id`, `question`, `recorded` |
| flag shape `avenue-id` (`flags/shapes.go:177`), id kind `avenue` (`anchor/anchor.go:70`) | `inquiry-id`, `inquiry`; the prefix stays `Q-` |
| blocker kind `avenue_review` (`record/passblockers.go:39`; `debate.js:399`, generated) | `inquiry_review` |
| scorecard metrics `avenues`, `thin_avenue_reasons` (`scorecard/scorecard.go:451–473`) | `inquiries`, `thin_inquiry_reasons` |
| consistency check `avenue-md` (`consistency/consistency.go:506`) | `inquiry-md` |
| run file `avenues.md`, nav `Avenues`, title `avenues` (`report/docs.go:54,136`); its headings `## pursued (n)`, `## concluded (n)` | `inquiries.md`, `Inquiries`, `inquiries`; `## underway (n)`, `## confirmed (n)` (from the enum) |
| help template value `{{.AvenueSectionsRule}}`; `debate.js` constants `AVENUE_SECTIONS`, `avenueClause` | `{{.InquirySectionsRule}}`; `INQUIRY_SECTIONS` (generated, §III.6), `inquiryClause` |

**Identifiers that are not the rule:** `.Line`, `GetLine` → `.Question`, `GetQuestion`; `newAvenueMove` →
`newInquiryRecord`; `avenueResult.Moved` → `inquiryResult.Recorded`; the fuzz's `avenueMoves` → `inquiryRecords`;
`ExistingProposalByText`'s parameter `line` → `question`; `TestRedCannotProposeAAvenue` →
`TestRedCannotProposeAnInquiry`. `report.ran()` reads `underway`, `confirmed`, `refuted`.

**[NEW] `flags.Question = "question"`**, registered on `propose` with `flags.Text` (as `--hypothesis` is, which
attaches the quoting rule) and `seat.Require`. `propose` stops being a `seat.Prose` verb: its `ReasonIs` row is
[DELETE]d (`seat/seat.go:941`); `satisfiedByAnyOf["line"]` (`seat.go:537`) and the payload-key row `"line": Reason`
(`flags/names.go:430`) become `"question": {flags.Question}` and `"question": Question`; `record.go:1307`'s refusal
names `--question` through `flags.ForPayloadKey`. The 12 call sites passing `--reason` to `propose`
(`git grep -nIE 'avenue.{0,3}propose.{0,80}--reason|"propose".{0,60}"--reason"' -- plugins scripts`) change with it.

**[MODIFY] A proposal holds no reason (N10).** Consumer census —
`git grep -nE '\b(a|q)\.Reason\b|\ba\.GetReason\(\)' -- $T/internal/{report,view,scorecard} $T/internal/record/avenue.go ':!*_test.go'`
→ 8 lines, with the writer:

| Site | Changes |
|---|---|
| `cli/blue/avenue.go:116–117` (the writer) | yes: `propose` writes `question` and leaves `reason` unset |
| `report/assemble.go:684` (`a.Reason != "" && a.Reason != a.Line`) | yes: the second operand is [DELETE]d |
| `scorecard/scorecard.go:434` (`st != "pursued" && len(reason) < 20`) | yes: the detector skips `proposed` and `underway` |
| `report/assemble.go:689`; `view/view.go:775–776`; `record/avenue.go:158,329` | no: they print or carry a reason when one exists, so `show inquiries` prints a proposed inquiry's question once |
| `record/avenue.go:94` (`StruckAvenueTexts`: the struck act's `line`, else its reason) | no: a proposal always holds a question, and the act that holds none is a recorded fate, whose reason is its text |

```
F/requirements.json                               [MODIFY] eventSchema 26
F/{README.md,docs,agents,skills,tests/simulator}  [MODIFY] by hand or generated, per §III.2
T/devcmd/archdump/main.go                         [NEW] (N12)
T/devcmd/enginegen                                [MODIFY] writes INQUIRY_SECTIONS (§III.6)
T/integration/surface/fatenames_test.go           [NEW] (§III.6)
T/internal/cli/seat/help/record.md                [NEW]; move.md [DELETE]
T/internal/{cli/blue,cli/chair,record}/inquiry.go [git mv]; and 10 `inquiry*_test.go` beside their packages
T/internal/record/recordpb/, recordsql/views.go, migrate/*.go, T/internal/terms/terms.json   [MODIFY]
scripts/agentgen/src/** (6), feov-memory/protocol-class-registry.md (1 line)                 [MODIFY]
```

### III.5 Migration: one translation per changed word

Old-word keys map straight to the current word, so a key already named `inquiry` becomes identity and is deleted,
and a key named `avenue` becomes a new translation. No word passes through two hops. Counts: `final.py`.

| # | Old word | Where, with counts | Maps to | Mechanism |
|---|---|---|---|---|
| M1 | body table `avenue` | 14 SQLite archives (2026-09-02 … 10-12), 540 acts | `inquiry` | [NEW] `renamedBodyTables["avenue"]`, entry `"avenue": mapped("inquiry", ren{"avenue_id": "inquiry_id", "line": "question"}, …)` |
| M2 | event `line-of-inquiry`, registry key `line_of_inquiry` | 6 JSONL runs (2026-08-22/23), 405 events | `inquiry` | [MODIFY] `eraentries.go:58`: `mapped("inquiry", ren{"line": "question"}, …)`; its `inquiry_id` rename is identity, deleted |
| M3 | field `line` | 275 proposal acts in all 20 archives | `question` | the `ren` of M1 and M2 |
| M4 | status `pursued` | 508 acts in 16 archives | `underway` | [NEW] `valueRenames["feov.record.v1.InquiryStatus"]` |
| M5 | status `concluded` | 33 acts in 4 archives (wave-a, m19, m20, wave-b) | `confirmed` | same row |
| M6 | body table `avenue_review` | 4 archives, 28 events | `inquiry_review` | [NEW] `renamedBodyTables` row and entry |
| M7 | body table `inquiry_review` | 10 archives, 40 events | identity | [DELETE] `entries.go:25` and `sqlitesource.go:230` |
| M8 | event `inquiry-support`, registry key `inquiry_support` | 6 JSONL runs, 277 events | `inquiry_review` | [MODIFY] `eraentries.go:63` target; drops unchanged |
| M9 | motion subject `direction` | 5 rulings (quadratic-formula) | `inquiry` | [MODIFY] `valueRenames` `MotionSubject: {"direction": "inquiry", "avenue": "inquiry"}` |
| M10 | motion subject `avenue` | 14 rulings (m20 10, wave-b 4) | `inquiry` | same row |
| M11 | era motion subject `inquiry` | 5 JSONL rulings (4 + 1) | `inquiry` | [MODIFY] `eraentries.go:193–199`: the arm and enum renamed |
| M12 | ruling column `motion_rule.direction`, `motion_rule.avenue`; arm table `motion_direction`, `motion_avenue` (`avenue_id`) | columns: the 19 rulings of M9, M10; arm tables: 0 rows in every archive | arm `inquiry`, field `inquiry_id` | [MODIFY] `inquiryArm`, renaming both old arms and the arm's id field |
| M13 | `about_kind = inquiry` | 5 findings, 3 mints (b, b5, b6, b7) | identity | [DELETE] `valuerenames.go:24` |
| M14 | `about_kind = avenue` | 33 findings, 4 mints (wave-a, m19, m20, wave-b) | `inquiry` | [NEW] `AboutKind: {"avenue": "inquiry"}` |
| M15 | a proposal's `reason` equal to its `line` | 275 of 275 proposal acts | dropped (N10) | in the M1 and M2 entries: a row with no `supersedes_status` whose `reason` equals its `line` carries no reason forward |

Not translated: `supersedes_status` holds `"1"` or nothing in all 945 acts, never a status word; the `enum_*` tables
and views are not read by migrate; ids keep their value (§II). `remap.go`'s lists change name only.

**Residue.** `migrate` translates words one for one and re-reads no prose: the 508 acts recorded `pursued` arrive
`underway` and the 33 recorded `concluded` arrive `confirmed`, and of the 25 archived inquiries that end
`concluded`, 3 state a refutation in their reason and arrive `confirmed`. Accepted.

**Tests, each on an archived run** ([MODIFY] `inquiryfate_test.go`, `inquiryrename_test.go`, `shardarchive_test.go`):

| Translation | Archive | Asserts |
|---|---|---|
| M1, M3 | `2026-09-11_is-91-prime-b9` | 26 acts arrive as `inquiry`; every proposal's `question` equals the archived `line` byte for byte |
| M2, M8 | `2026-08-22_sqlite-schema-source` | 57 `line-of-inquiry` and 18 `inquiry-support` events arrive as `inquiry` and `inquiry_review` |
| M4 | b9 | the 11 `pursued` acts arrive `underway`; no migrated act holds `pursued` |
| M5 | `2026-10-11_is-91-prime-m20` | the 7 `concluded` acts arrive `confirmed`; `abandoned` and `declined` are untouched; none arrives `refuted` |
| M6, M7 | m20; b9 | 5 and 6 reviews arrive as `inquiry_review` |
| M9, M12 | `2026-09-02_quadratic-formula` | its 5 `direction` rulings arrive under subject `inquiry` with the ruling in the `inquiry` arm |
| M10; M11 | m20; `2026-08-22_is-7-prime-run2` | its 10 rulings likewise; its 4 era rulings likewise |
| M12 (arm table) | b9 with a planted `motion_direction` row, as `avenuerename_test.go:40–48` plants a ruling today | the arm and its `inquiry_id` arrive; no archive holds such a row, and the test says so |
| M13, M14 | `2026-09-09_is-91-prime-b`; `2026-10-12_is-91-prime-wave-b` | 2 and 7 findings (and wave-b's 3 mints) arrive `about_kind = inquiry` |
| M15 | b9 | no migrated proposal holds a reason; every later act's reason is byte-equal to the archive's |
| all | `TestEveryArchivedRunRenders`, `everyarchive_test.go` | 20 runs, 0 refusals; the three `migrate/testdata/archived_*.golden` files are expected unchanged — a changed line is read, not regenerated past |

### III.6 The terms registry, the fate lists and the one check on fate values

**[MODIFY] `terms.json`, then `vocabdoc`.** The entry at `:924` becomes **inquiry**: "An inquiry is one question the
research could take up — proposed by blue with the hypothesis that would be true if it paid off, recorded with what
became of it, ruled on by the chair, and reviewed against the report once per epoch." Its three bans (`:933–947`)
and two collisions (`:949–957`) are [DELETE]d and nothing is added: `terms.Load` requires of an entry a term, a
seat and a one-sentence definition, and `TestVocabularyProse` requires no entry at all (§II), so no collision note
for `question`, `record` or `inquest` is owed. The collision at `:838` reads "an inquiry's state keeps the word".
`retiredSurfaces` (`T/integration/surface/retiredsurfaces_test.go`) gains no row. Plain English that holds the
noun is reworded because S1's noun grep must print nothing: "BOTH AVENUES" in
`scripts/agentgen/src/blue-researcher.md:53,58` becomes "BOTH ROUTES".

**Every place the fates are enumerated by hand.** The census — lines outside generated files and tests that hold
three or more of the seven status words:

```
W='proposed|pursued|deferred|declined|abandoned|concluded|refuted'
git grep -lIiE "$W" | grep -vE "$NOT|$GEN|_test\.go$|\.mjs$" | xargs git grep -nIiE "$W" -- \
  | awk '{l=tolower($0); n=0; split("proposed pursued deferred declined abandoned concluded refuted",w," ");
          for(i in w) n+=(index(l,w[i])>0); if (n>=3) print}'
```

→ 23 lines in 13 files. Seven are a list of the set, and each stops being hand-kept:

| Site | What it is | It becomes |
|---|---|---|
| `T/internal/cli/seat/verbs.go:281` | the view table's two descriptions of the inquiries view, rendered into `show --help` and all 11 agent definitions | [MODIFY] joined from `record.InquiryStatusNames()` — the pattern of `view/view.go:760` |
| `debate.js:189` (`AVENUE_SECTIONS`) | which fate prints under which heading: a hand copy of `report.AvenueSections`, which `report/layout.go:253` already writes from the layout rows | [MODIFY] `enginegen` writes `INQUIRY_SECTIONS` into its first generated region from `report.InquirySections`; the hand copy asserted at `debate.test.mjs:692` is [DELETE]d — `report/assemble_test.go:440` holds the rendered prompt to the Go sentence |
| `record.proto:244` (the event type's `(means)`) | an annotation, which no generator can write | reworded to need no list: "an inquiry: its question, its hypothesis and each fate recorded on it" |
| `terms.json:925` → `F/docs/vocabulary.md:529` | the definition | the definition above, which lists none |
| `F/docs/seat-command-triggers.md:89` | a docs row | "what became of an inquiry you already proposed"; the verb's page prints the fates |
| `references/report_template.md:74` | a template comment restating the sections rule | names the section's content without the list; the rule is `INQUIRY_SECTIONS` |
| `T/internal/report/docs.go:137` | the run index blurb | "every inquiry on the record: the path each one took…" |

The other sixteen are not lists of the set. One is code: `report/assemble.go:636` (`ran()`), pinned per fate by
`TestEveryAvenueFateIsSaidOnItsRow`. Seven are comments, which reach no seat: `cli/blue/avenue.go:49` (deleted,
§III.2), `cli/chair/avenue.go:19`, `record/avenue.go:154–155`, `record.proto:1335`, `report/assemble.go:612,679`.
Eight are prose that names particular fates to say what each means: `cli/seat/help/move.md:7`, `debate.js:208`
and `:218`, `record.proto:431,436,444,981`, `view/view.go:807`. With them go the sentences under the census's
threshold, which S1's 366 lines hold — among them `cli/blue/avenue.go:141`, `seatprobe/boards.go:367`,
`scorecard/scorecard.go:453`. All of these are renamed by the sweep and held by review.

**[NEW] `TestEveryFateNamedToASeatIsAFate`** (`T/integration/surface/fatenames_test.go`).

- **Reads.** For every leaf of the cobra tree that takes `--as`, the seat-filable names of the enum that verb
  binds to it: `record.Names(record.SeatFilable(…))` over `record.EnumFields` for the verb's event type, and
  `record.MotionVerdicts` for a motion's subject. No word is written in the test.
- **Scans.** `vocabularySources(t)` — the file set of `TestVocabularyProse`: constitutions, the protocol skill with
  its references and templates, help pages, the engine's prompt text, seat-facing Go string literals, the proto's
  annotation text, both READMEs — plus the fourteen rendered prompts (`F/tests/simulator/testdata/prompt-*.golden`).
- **Fails.** An `--as <value>` pair (the pair pattern of `promptverbs_test.go`, which a backticked flag never
  matches) whose value — each side of a `|` — is not a fate of ITS verb. The verb is the command path nearest
  before the pair in the same sentence or code line, matched as whole words, longest path first, so `inquiries`
  is not `inquiry` and `motion inquiry rule` is not `inquiry record`. A pair with no command path before it is
  left to `TestEveryEnumValueNamedInAPromptIsAccepted`, which passes a value legal for any enum bound to `--as`.
  The test prints the pairs it attributed to each verb and fails when the scan yields no source.
- **What it reads today.** No fate:
  `git grep -nIoE -- '--as [A-Za-z][A-Za-z_|-]*' -- $F/agents $F/skills $T/internal/{cli,flags,view,record} README.md $F/README.md $F/tests/simulator/testdata ':!*_test.go' ':!*.pb.go' ":!$T/internal/record/migrate"`
  finds 95 pairs and one holds a fate — a comment (`cli/blue/avenue.go:49`), which the scan does not read. Seats
  learn the fates from the verb's page, which cobra prints from the enum; the check holds the next sentence that
  tells a seat to type one.
- **Mutation rows**, in the style of `TestVocabularyProseMutations`, each at a path the scan reads — one a prompt
  golden — and a row whose path it does not read fails. Fail: `inquiry record --id Q-1 --as zzfate`;
  `inquiry record --as confirmed|zzfate`; `inquiry record --as endorsed` (legal for `motion inquiry rule`, so the
  any-enum gate passes it); "read `show inquiries`, then for each of your inquiries run `inquiry record --as
  granted`" (a plural and a view name take no part in the attribution); `motion inquiry rule --as confirmed`.
  Pass: `inquiry record --as confirmed`; `motion inquiry rule --as endorsed`. By hand, once: delete the test's
  call to the scanner and it fails on the empty scan; delete `INQUIRY_STATUS_REFUTED` in a scratch copy and
  regenerate — the two rendered lists lose the word with no edit.
- **The blind spot, stated.** A fate word that is neither an `--as` value nor inside a rendered list — the eight
  prose sites above, and any sentence like them — is held by the sweep at landing and by nothing afterwards.

### III.7 The byte budget for prompts

`go -C T test -count=1 -run TestNoSeatPromptGrowsPastItsCeiling -v ./internal/seatprobe/` at `e8f6ec9c`. `inquiry`
is 1 byte longer than `avenue`, `inquiries` 2 longer than `avenues`, `underway` 1 longer than `pursued`,
`confirmed` equal to `concluded`, `question` 4 longer than `line`.

| Seat (gated) | Now / ceiling | Concept words in the prompt | Delta | After |
|---|---|---|---|---|
| `red-chair` | 10,131 / 10,200 | `debate.js:1030`: "VOTE EVERY AVENUE", "answer every line", "once per line", "BLUE'S AVENUES", "becomes of the avenue" | +1 +3 +3 +2 +1 = **+10** word for word; "answer each against that pass" in place of "answer every line…" makes it +1 | ≤ 10,141 |
| `red-lens-evidence` | 12,289 / 12,300 | none: it takes no steelman clause (`debate.js:1011`), and its prompt goldens hold 0 concept words | 0 | 12,289 |
| `blue-respond` | 11,170 / 15,200 | `:1035`: 2 × avenue, 1 × AVENUES, 2 × line, 1 × pursued | ≤ +12 | ≤ 11,182 |
| `judge` | 7,056 / 7,900 | none | 0 | 7,056 |

Prompts no ceiling measures: the logic and dark-side lens prompts gain about 23 bytes from the steelman clause; the
frontier and lane prompts shrink, because "ITS LINE IS THE QUESTION the avenue asks" loses its first half. No
ceiling is raised. The implementer pastes the `-v` table before and after.

### III.8 Order of work, and the check that closes each step

By kind, never a repository-wide text replace (R1). No step begins until the one before it passes its check. One
commit per step. The PR merges by merge commit, as every PR here does, so an intermediate commit is not required
to pass: step 1 closes on a tree that does not compile, and its commit message says so.

| # | Step | The check that closes it |
|---|---|---|
| 0 | Baseline at the base: commit `T/devcmd/archdump` (N12), touching no other path; dump every archive's assembled set to `~/.claude/scratch/inquiry-rename/before/`; build `…/bin/feov-record-25`; save each `--help` page of the group | 20 directories, 0 `REFUSALS` lines, 0 `ASSEMBLE_ERROR` |
| 1 | Schema: `record.proto` from §III.4; `protogen`; `requirements.json` to 26 and `schemagen`; the two views by hand | `go -C T build ./internal/record/recordpb/` passes and nothing that imports it does, which is step 2's worklist |
| 2 | Go identifiers, led by the compiler. First the names under the rule, generated and hand-written alike: for each name §III.4's census prints, `gofmt -l -w -r '<Old> -> <New>'` over `git ls-files "$T/*.go" ':!*.pb.go'` — it rewrites whole identifiers, declarations and uses, and no comment or literal; a name that holds the noun belongs to this concept wherever it stands, which is what makes a rename without types safe. Then the names outside the rule (`.Line`, `GetLine`, `Moved`, `avenueMoves`, a parameter): each declaration by hand, and `go -C T vet ./... 2>&1 \| grep -E '\.go:[0-9]+'` — which type-checks tests as well — names every reader, a package at a time as its imports come to compile; repeat until it prints nothing | `go -C T build ./... && go -C T vet ./...`; `UPDATE_GOLDENS=1 go -C T test -count=1 ./internal/record/recordsql/ ./internal/record/recordpb/`, then every changed line of `schema.sql` read: `SELECT name FROM sqlite_master GROUP BY name HAVING count(*) > 1` is empty, and `inquiry_state` reads FROM `"inquiry"` |
| 3 | Migrate: §III.5, tests first | the tests of §III.5's table but its last row, by `-run` on their names; S6; S2 prints nothing. The package whole — whose render tests print words step 4 moves — closes at step 4 |
| 4 | CLI: verbs, flags, JSON keys, refusals, N10, literals in the 17 CLI files and the record folds | `go -C T test -count=1 ./internal/cli/... ./internal/record/...`, migrate whole among them; each of `inquiry propose`, `inquiry record`, `inquiry review`, `motion inquiry rule`, `show inquiries` rendered with `--help` and read against step 0's page for a lost instruction |
| 5 | Terms, the rendered lists and the fate check: §III.6; then `enginegen`, `agentgen`, `vocabdoc` | `go -C T test -count=1 -run 'TestVocabularyProse\|TestEveryFateNamedToASeatIsAFate' ./integration/surface/`: both pass, the first with no stale mask; the mutation rows fail and pass as written; `agentgen -check` and gate `enginegen` are clean |
| 6 | Seat-facing prose: help pages, `(means)`, the four prompt constants, constitutions, skill sources, references — sentence by sentence (§III.3), inside the byte budget (§III.7); then `enginegen`, `agentgen`, `vocabdoc` | `TestEveryFateNamedToASeatIsAFate`, `TestNoSeatPromptGrowsPastItsCeiling -v`, `TestTheSeatPromptsNameNoVerb`, `TestNoPromptGrowsItsCommandCatalogue`, `TestDetailsDoNotRestateTheFlagsCobraPrints` pass; `agentgen -check` and gate `enginegen` are clean |
| 7 | Goldens: for each, write the PREDICTED file in scratch — the old golden with §III.4's word map applied by a script — regenerate (both commands of §II's table), and `diff` predicted against regenerated | every line the two disagree on is read and explained in the PR; lines that agree are the rename and nothing else |
| 8 | The release fuzz, by hand: the three files, the drive (`fuzz_test.go:1938–1988`, `:5121–5163`), the coverage key per recording verb, the oracle's messages | §V.5; and by mutation: take `refuted`, then `underway`, out of `nextInquiryStatus`'s rotation — the coverage gate turns red each time |
| 9 | Simulator, docs, comments, tests: `debate.test.mjs`, the relay fixture (N16), the prompt goldens, `F/README.md`, `F/docs/*.md`, the class registry line, §III.2's deletions | `node --test plugins/frank-exchange-of-views/tests/simulator/*.test.mjs`; gates `decisions`, `mjsparity`, `archaeology`; **S1 holds** |
| 10 | §V.1 to §V.5 in full, then the smoke (§V.6) | S2–S8 |

### III.9 Landing without a freeze (N13), with the smoke inside it (N15)

**Rate, measured** (2026-10-13, the 72 hours before `1bd4332f`):
`git log --first-parent --since='72 hours ago' --format=%ct origin/main -- $F/skills/research-protocol/scripts/debate.js $F/agents $T/internal/difftest/testdata $T/internal/record/recordpb $F/tests/simulator/testdata $T/internal/terms scripts/agentgen/src $T/internal/cli/seat/help`
prints 38 of main's 59 commits — the ones touching a carrier this rename regenerates. The gap between two of them
has median 80 minutes and is at least 30 minutes in 31 of 37 cases, 45 in 29, 60 in 23, 75 in 20, 90 in 17, 120
in 13. Of the 38, 26 added a line holding the noun or a fate word to an authored seat-facing source — a sample
taken while this concept was being worked on, so an upper figure.

**The attempt window W, measured at the base** (16 cores, warm build cache, load 0.4 rising to 7.6 under the run
itself; `FEOV_*` unset; `TMPDIR`, `GOTMPDIR` under `~/.claude/scratch/inquiry-rename/tmp`): step 3 takes 50 seconds
(the Go goldens 43), step 5 takes 187 (the test subset 107, the gates 78) and CI 5.5 minutes (§II) — 10 minutes of
machine. The rest is reading: conflicts, what main added, the golden diff. It is known only once an attempt
exists, so the implementer times attempt 1 and writes W in the PR; the estimate below takes W = 30.

**One attempt**, begun at the first push and again each time main moves a carrier:

1. `gh pr view <n> --json state` → `OPEN`. Note the merge base (`git merge-base HEAD origin/main`), then
   `git fetch origin main && git merge origin/main` — a merge, never a rebase.
2. Conflicts. In an AUTHORED source (`record.proto`, `views.go`, Go, `debate.js` outside its generated regions,
   `scripts/agentgen/src`, `terms.json`, help pages, docs): take main's hunk and re-apply §III.4's names to it. In a
   GENERATED file: resolve nothing by hand — take either side; step 3 overwrites it.
3. Regenerate, in this order: `protogen`; `schemagen`; `go -C T build ./...` (the compiler names every identifier
   main added under an old name); `enginegen`; `agentgen`; `vocabdoc`; both golden commands of §II's table.
4. Read what main ADDED:
   `git diff <the noted merge base>..origin/main | grep -E '^\+[^+]' | grep -iE 'avenue|conclud|pursu|\blines?\b|\bmov(e|ed|es|ing)\b'`
   — every line it prints is a new sentence that may be in the old words, and is classified as §III.3 classifies.
   Then S1 holds and S2 prints nothing; `git diff HEAD -- '*.golden' '*.sql'` is read — every changed line is one
   main changed.
5. The fast subset: `go -C T vet ./...`;
   `go -C T test -count=1 ./internal/record/... ./internal/cli/... ./internal/report/ ./internal/view/ ./internal/seatprobe/ ./internal/difftest/ ./integration/surface/`;
   `node --test plugins/frank-exchange-of-views/tests/simulator/*.test.mjs`;
   `go -C scripts run ./check -only archaeology,frontmatter,decisions,rulesweep,validatejson,vocabdoc,enginegen,agentgen,golden`
   → fail count 0 in each.
6. Commit the merge, check `gh pr view <n> --json state` again, push, and read CI by fail count. **Merge when CI is
   green and the smoke stands** (N13, N15). If main moves a carrier first, that is the next attempt; a main commit
   that touches no carrier needs only steps 1, 4 and 5.

**What gates a later attempt, and what runs once.** On every attempt the gate is step 5 and CI's suite, which runs
the whole of `T`'s tests (§II). Run once, on the first complete tree, and again only when a merged commit touches
the re-arm surface: §V.2 steps 1 to 3; §V.3 and §V.4 on `T/internal/{record,report,view}`; §V.5 on a record, flag
or enum change; §V.6 as below. **`before` at a moved base** (§V.3): `archdump` exists only on the branch, and step
0's commit touches no other path and is written against the base's names, so
`git worktree add --detach ~/.claude/scratch/inquiry-rename/base origin/main`, `git -C …/base cherry-pick <step 0's commit>`,
and the dump runs from there; it fails to compile, loudly, if main changed what it calls.

**The smoke, once.** It runs on the tree of the first attempt, beside that attempt's CI. A later attempt re-arms
it only when it changed a seat-facing carrier BY HAND — a conflict resolved in step 2, or a line step 4 renamed,
in a file `vocabularySources(t)` reads (§III.6), a cobra command or flag declaration, or `record.proto`. A clean
merge and a regeneration change nothing the rename wrote, and re-arm nothing.

**Estimate.** An attempt lands when no carrier moves on main within its window. A plain attempt, W = 30, lands 84%
of the time. The first attempt carries the smoke — the universe's build, about 45 minutes of run, the reading of
its outcome — so its window is 75 to 90 minutes and it lands about half the time (54% to 46%): the smoke roughly
triples the window. When it does not land, the next attempt re-arms the smoke in at most 26 of 38 cases. Expected:
about 1.5 smokes and two or three attempts, two to four hours from the finished tree. Past that, the attempt is
begun at a quiet hour; nothing in §V is cut.

## IV. Risk & Mitigation

Graded likelihood × impact × cost to mitigate (L/M/H).

| # | Risk | L × I × cost | Mitigation, and the step or gate that carries it |
|---|---|---|---|
| R1 | A mechanical rename damages what it cannot read: a view lands on a table's name, a phrase is split across literals, an article is left wrong, a check passes for the wrong reason | H × H × M | Step 2's `sqlite_master` check; step 2 renames whole identifiers and touches no literal; §III.3 pass 4 reads rendered text; S6 deletes each translation; `git grep -nIE '\ba inquir'` → 0 at step 9. No step uses a repository-wide text replace |
| R2 | Main moves under the branch, and adds a sentence in the old words | H × M × M | §III.9: generated files are never resolved by hand; step 4 reads what main added on every attempt and S1 is re-run; `gh pr view --json state` before every push |
| R3 | After the merge, new prose names a fate or a surface that does not exist | M × L × L | An identifier fails to compile; an `--as` value fails `TestEveryFateNamedToASeatIsAFate`; a list of the fates is rendered and cannot be stale; a verb or view a prompt names fails `TestEveryVerbNamedInAPromptExists` or `TestEveryViewNamedInAPromptExists`. A fate word in running prose is held by review alone (§III.6, the blind spot) |
| R4 | The epoch bump refuses every live run, scratch record and universe at epoch ≤ 25 | H × M × L | That is the rule working. The PR names `feov-record migrate --from <old> --to <new>`; a running session keeps the binary it loaded. §V.4 proves the refusal's wording and the migration on a real epoch-25 record |
| R5 | `inquiry record` reads as a noun phrase beside "the record" and `feov-record`; `inquest` is read for `inquiry` on the chair's root help | M × L × L | The group's Short and the page's menu line open with the act ("say what became of an inquiry you proposed"); the Shorts are read side by side in step 4; the seat-probe board (`seatprobe/boards.go:367`) and the smoke (§V.6), which runs before merge, show whether a seat finds the verb |
| R6 | A translation is wrong and silent: `question` filled from the wrong column, or M15 dropping a real reason | M × H × L | M1/M3's test compares `question` with the archived `line` byte for byte; M15's asserts every non-proposal reason is byte-equal; S4 diffs every assembled document of 20 archives |
| R7 | A deleted passage carried the only statement of a live instruction, or a deleted docs row is one `decisions` reads | M × M × L | Each deletion is read for an instruction first, and the instruction restated in the present tense (§III.2); gate `decisions` at step 9; the PR lists every deleted passage by file and line range |
| R8 | The fate check reads nothing, or attributes a value to the wrong verb | M × M × L | It reads no fate value in today's text and the plan says so (§III.6); its mutation rows are what prove it fires — a path the scan does not read fails, a value legal for another verb fails, the empty scan fails |
| R9 | The chair prompt passes its ceiling (59 bytes spare after +10); a lens prompt no ceiling measures grows unseen | L × M × L | §III.7's wording; gate `TestNoSeatPromptGrowsPastItsCeiling`; the fourteen simulator prompt goldens show every byte in the diff |
| R10 | `archaeology` fails a renamed line for wording it already carried; `deadcode` fails a helper the rename orphans; the `manual` golden hoists a block repeated on two pages into SHARED and renumbers | M × L × M | The first two are loud and in `check`; an archaeology hit is fixed in the sentence, never allow-listed. Step 7's predicted-against-regenerated diff isolates a renumbering |
| R11 | `propose` leaving the prose channel loses a protection the channel gave (the quoting rule, the heredoc trim) | L × M × L | `flags.Text` attaches the same rule (`flags/register.go:23–32`); `TestEveryFreeTextVerbShowsTheQuotingRule` holds it; the trim (`Prose.Read`) is applied to `--question` and pinned by a test that passes a value ending in a newline |
| R12 | Another plugin or an outside reader depends on a FEOV record word | L × M × L | None does: §III.2's not-carriers command → 0, re-run in §V.1 |
| R13 | The smoke is run in a universe that is not the branch, bills API credits, or is repeated on every attempt | M × M × L | §V.6: the installed copy is grepped for the change, the auth method is read before the run, the run's own binary is named by absolute path; §III.9 re-arms the smoke only on a hand change to a seat-facing carrier |

## V. Verification Plan

Before any FEOV command: `unset $(env | grep -o '^FEOV_[A-Z_]*')`; `GOTOOLCHAIN=local`. Each check names what
re-arms it. Logs go to `~/.claude/scratch/inquiry-rename/logs/`.

### V.1 Generators and censuses · re-armed by any edit to a source in §II's table, and by every merge of main

1. Each generator of §II re-run, then its check: `protogen` and `vocabdoc` leave `git diff --exit-code` clean;
   `schemagen -check` and `go -C T run ./cmd/feov-record --schema` → 26; `agentgen -check` and
   `go -C scripts run ./check -only enginegen` → clean.
2. S1 holds — the noun grep prints nothing, the fate grep equals the list in the PR — and S2 prints nothing; S1's
   excluded set → only §III.2's prefixes. R12's command → 0.

### V.2 Suites · steps 1 to 3 on the first complete tree and on their re-arm surface (§III.9); step 4 and CI on every attempt

1. `go -C T test -count=1 -timeout 90m ./... > …/tools.log 2>&1`, then `grep -c '^FAIL' …/tools.log` → 0 and
   `grep -c '^ok' …/tools.log` equal to what the same command prints in a detached worktree at
   `git merge-base HEAD origin/main` — the merged base, never a number written here. The branch adds no test
   package.
2. `go -C scripts test -count=1 ./...` → 0 FAIL. `node --test plugins/frank-exchange-of-views/tests/simulator/*.test.mjs` → `fail 0`.
3. **S6, by hand** · re-armed by any change under `T/internal/record/migrate`. Delete, one at a time, each of
   M1–M15's registry rows, the `question` rename, the two `InquiryStatus` rows and the `inquiryArm` field rename;
   run `go -C T test -count=1 ./internal/record/migrate/`. Each deletion fails a named test; the PR lists the
   pairs. A deletion nothing fails is a missing assertion, fixed before merge. The same for N10: restore
   `body.Reason` in `propose`, and a named test fails.
4. `go -C scripts run ./check > …/check.log 2>&1` → the summary line's failed count is 0. `qlty` and `actionlint`
   run only in CI: read the PR's checks by fail count.

### V.3 Real data: every archived run, before and after (S3, S4) · re-armed by any change under `T/internal/{record,report,view}`

1. Before, at the base: `go -C T run ./devcmd/archdump -archives run-archive -out ~/.claude/scratch/inquiry-rename/before`
   (rebuilt at a moved base as §III.9 says).
2. After, on the branch: the same with `-out …/after` → 20 directories; `cat …/after/*/REFUSALS | wc -l` → 0;
   `ls …/after/*/ASSEMBLE_ERROR` → none.
3. Apply the word map to a COPY of `before` (`wordmap.py`, kept in scratch and pasted in the PR) — in `report.md`
   the `**Fate:**` word; in `inquiries.md` its file name, nav label, title, the two fate headings and the `path:`
   words; in `README.md` and the link bars the file name and label; in `run.md` and the dashboard the two metric
   names and the status keys; in `debate.md` and `judgments.md` the verb and subject words where an act is named —
   then `diff -r …/before-mapped …/after` → **no output**. No archive rests at `proposed`
   (`grep -c '^## proposed' …/before/*/avenues.md` → 0 each), so N10 changes no document. A line that differs is a
   defect.
4. Read by hand, in `after`: `2026-10-11_is-91-prime-m20/report.md`'s three inquiry sections and `inquiries.md`
   whole; `2026-09-11_is-91-prime-b9/inquiries.md` (`## underway (7)`).

### V.4 Real data: a record at epoch 25 is refused, then migrates (S5) · re-armed by any change to the epoch or to `migrate`

1. With `feov-record-25` (step 0), make a real epoch-25 record: extract `run-archive/2026-10-12_is-91-prime-wave-b.tar.gz`
   to `~/.claude/scratch/inquiry-rename/e24/` and run `feov-record-25 migrate --from …/e24/<run> --to …/e25/wave-b`.
2. Branch binary: `feov-record show --run …/e25/wave-b --seat-id red-chair inquiries` → **refused**; the message
   names epoch 25, epoch 26 and `migrate`.
   `go -C T test -count=1 -run TestEveryArchivedRunBehindThisBinaryIsRefusedUnmigrated ./internal/record/migrate/`
   passes over all 20 archives.
3. `feov-record migrate --from …/e25/wave-b --to …/e26/wave-b` → exit 0, 0 refusals in `inputs/migration.json`.
4. `feov-record show --run …/e26/wave-b --seat-id red-chair inquiries --json | jq '[.inquiries[].status] | group_by(.) | map({(.[0]): length}) | add'`
   → `{"confirmed": 4}`; `… | jq '.inquiries[0] | keys'` holds `question` and not `line`.

### V.5 The release fuzz · re-armed by any record, flag or enum change (this is one)

1. `FEOV_RELEASE_GATE=1 go -C T test -count=1 -v -timeout 40m ./releasegate/fuzz/ > …/fuzz.log 2>&1` →
   `grep -o 'fuzzed [0-9]* debate runs · [0-9]* failed' …/fuzz.log` prints `fuzzed 40 debate runs · 0 failed`;
   `grep -c -- '^--- FAIL'` → 0; `grep -c -- '^--- SKIP'` → 0 (a skip means the variable was not set). The coverage
   table shows `inquiry record` driving `underway`, `confirmed`, `refuted`, `deferred`, `declined`, `abandoned`.
2. `FEOV_RELEASE_GATE=1 FUZZ_N=6 FUZZ_C=6 go -C T test -count=1 -race -v -timeout 30m -run TestFuzzDebate ./releasegate/fuzz/`
   → `grep -c 'DATA RACE'` → 0. A red in either is run on `origin/main` before the branch is blamed.

### V.6 The engine runs · BEFORE merge (N15) · re-armed by a hand change to a seat-facing carrier (§III.9)

1. `scripts/universe.sh build` from the finished branch. Prove the universe carries the change:
   `grep -rl 'inquiry record' <the installed plugin copy>` prints the chair's and blue's agent definitions, and
   S1's two commands over that copy print only `tools/internal/record/migrate/`, `tools/research/`, the OCR corpus
   and the English list.
2. Two traps, each checked before the run. `CLAUDE_CONFIG_DIR=<universe>/config claude auth status` shows
   `"authMethod": "claude.ai"`: an isolated config directory that lost the subscription bills API credits. Every
   `feov-record` call from outside the run names the run's own binary by absolute path: a bare `feov-record`
   reaches the host's install cache at the earlier epoch and is refused.
3. `SMOKE=1 scripts/universe.sh run "<topic>"`, detached, its stream to a log: one lane, `haiku` on both tiers,
   `--mint-budget 3 --k-max 2`. `scripts/universe.sh watch` reads the board meanwhile.
4. It reaches an outcome, VERIFIED or CEILING, before the PR merges. Read in it: the frontier seat's first
   `inquiry propose` carries `--question` and is accepted first time; at least one
   `inquiry record --as confirmed|refuted`; every refusal a seat met on the `inquiry` group (`feov-record
   dashboard`, the run's `ops log`), each with the sentence that taught the seat what it typed; the chair's
   `inquiry review`; `inquiries.md` in the report set; the outcome and wall clock against the last smoke. A
   refusal that traces to the rename is fixed, and the fix re-arms the smoke.

### V.7 The auditor gate

`/prosthetic-conscience:plan-audit plans/avenue-becomes-inquiry.md` — approved on the second consecutive PASS, or by
gblock when every remaining gap is `local`. Its NOTES travel into the PR, with: S1's English list and its excluded
set, the deleted passages (R7), the three §III.3 classification tables, the §III.7 size table before and after,
the S6 pairs, `final.py` and its output, W, the attempt and smoke counts (§III.9), the smoke's outcome, and the
censuses of §III.2 re-run on the PR's base.
