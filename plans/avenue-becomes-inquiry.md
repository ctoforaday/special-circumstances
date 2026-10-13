# An avenue becomes an inquiry: the noun, its question, its fates and the act that records them

> STATUS: **PROPOSED 2026-10-13.** The names are ruled (§III.1); the forks in §III.10 are not. It awaits the
> plan-auditor gate and gblock's approval. No code changes with this file.

`F` is `plugins/frank-exchange-of-views`, `T` is `F/tools`. Every command runs from the repository root unless it
begins `go -C`. Line numbers and counts are at `e8f6ec9c` (`origin/main`, 2026-10-13, event-schema epoch 25, after
PR #1362). "Noun census" is `git grep -niI avenue`; a count written `L/F` is lines / files. Archive figures come from
`~/.claude/scratch/inquiry-rename/plan/{scan.py,final.py}` over the 20 tarballs of `run-archive/`, extracted to
`~/.claude/scratch/inquiry-rename/plan/x/`; the census scripts beside them (`census1.sh` … `census8.sh`) print every
number in this plan.

## I. Summary & Goals

**The problem.** Blue records each approach the research could take, and what became of it. Five names on that
concept say something other than what the thing is.

| Today | What it fails to say | Evidence |
|---|---|---|
| `avenue` (the noun) | It holds a question, a hypothesis and a finding; "avenue" names none of them. The report already labels its text `**Question:**` and its id letter is Q, "for the QUESTION the line holds" (`T/internal/flags/shapes.go:174`) | `plans/rulebook-audit.md:109–111` (2026-07-18): "'Avenue' alone fails nominative determinism — an outsider cannot tell what it holds." The 2026-09-28 rename onto it (`5dfddb60`) cites no collision and no ruling: it took the schema's word |
| `line`, filled by `--reason` (the proposal's text) | That it is a question. One flag, `--reason`, names two different texts on the two acts of one group, and `propose` writes its question into the field `reason` as well | 275 of 275 archived proposal acts hold `reason == line` (`final.py`). Four shipped reports (universe-m7, m12, m19, m22) carry a row whose line states a predicted result the body contradicts (`T/internal/report/assemble.go:647–652`) |
| `concluded` | Whether the answer was yes. Until epoch 24 it meant "answered, either way", and in English it still does | 3 of 25 archived inquiries resting at `concluded` state a refutation in their reason (`final.py`: wave-a `Q2`, m19 `Q-2d14f733`, m20 `Q-5b68435a`). On universe-m22 a lane filed a refuted hypothesis as `abandoned`, a lens read it as `concluded`, and the dispute held the gate (`T/internal/record/recordpb/record.proto:432–435`) |
| `pursued` | That it is still open. It is the one running state and reads as a finished past tense | 143 of 267 archived inquiries rest at `pursued` on finished runs (`final.py`); the steelman prompt has to say "there 'pursued' means only 'I intend to'" (`F/skills/research-protocol/scripts/debate.js:218`) |
| `avenue move` | What re-recording the same state is. `move --as pursued` on a pursued avenue moves nothing; the help calls it a REAFFIRMATION (`T/internal/cli/seat/help/move.md:7`) | 508 of 945 archived acts are `pursued`, most of them re-recordings (`final.py`) |

**The objective.** One noun, **inquiry**; its text is its **question**; it is **underway** while open, and
**confirmed** or **refuted** when it ran and answered; the act that says what became of it is **`inquiry record`**.
The old words leave every live surface in one pull request, with epoch 26 and a translation in `migrate` per word.

### Success criteria

| # | Criterion | Command | Today | Target |
|---|---|---|---|---|
| S1 | The noun is gone outside history | `git grep -lIi avenue -- . ':!plans' ':!law' ':!feov-memory' ':!research' ':!F/tools/research' ':!F/docs' ':!F/tools/internal/tessocr/testdata'` | 230 files | exactly the files of §III.3's "holds an old word on purpose" rows; the PR pastes the list and the lines |
| S2 | The old fates are gone as record words | the command under this table | 102 lines in 40 files (23 of them enum identifiers, in 13 files) | only `T/internal/record/migrate/` (the two translations and their tests) and lines S1 already lists |
| S3 | Every archive migrates | `go -C T test -count=1 -run 'TestEveryArchivedRunRenders\|TestEveryArchivedRunMigrates' ./internal/record/migrate/` | 20 runs, 0 refusals at epoch 25 | 20 runs, 0 refusals at epoch 26 |
| S4 | Shipped reports differ only in the renamed words | §V.3 (`archdump` before and after, word-mapped diff) | — | 0 differing lines after the map of §III.1 is applied to the "before" dump, except the lines §III.7 names |
| S5 | A record at epoch 25 is refused, then migrates | §V.4 | — | the refusal names 25, 26 and `migrate`; the migration reports 0 refusals |
| S6 | No translation is unexercised | §V.2 step 3 (delete each registry row, run the suite) | — | every deletion fails a named test |
| S7 | Gates | `go -C scripts run ./check`; the release fuzz and its race variant (§V.5) | 51 gates listed | fail count 0 in each log; `fuzzed 40 debate runs · 0 failed` |
| S8 | The engine runs | §V.6 universe smoke after merge | — | a VERIFIED or CEILING outcome whose report set holds `inquiries.md` and no old word |

S2's command — a fate in quotes or backticks, after `--as`, or as its enum identifier:

```
git grep -nIE "STATUS_(CONCLUDED|PURSUED)|[\"'\`](concluded|pursued)[\"'\`]|--as (concluded|pursued)" \
  -- . ':!plans' ':!F/docs' ':!research' ':!F/tools/research' ':!feov-memory' ':!*.pb.go'
```

### The value case, against leaving `avenue`

This is the **fourth** rename of the noun in three months: verb `avenue` under a section "Lines of Inquiry"
(`713aa58c`, 2026-07-18); everything to `line-of-inquiry` (`860806da`, 2026-08-16); everything to `avenue` (`5dfddb60`,
2026-09-28, PR #1196, 195 files, +3,852 −3,562, epoch 15→16); and this one.

- **What it buys.** The five rows above. Three are defects with a measured cost: a gate held on universe-m22 by two
  readings of `concluded`; four shipped rows contradicting their reports because nothing told the seat the line is a
  question; `--reason` meaning two things in one group. PR #1362 fixed the report row and added `refuted`; it left the
  record, the verbs, the flag and the prompts saying `line`, `concluded`, `pursued` and `move`. Leaving `avenue` keeps
  a report that prints "Question / Hypothesis / Fate" over a surface that names none of those.
- **What it costs.** The noun is on 2,393 lines in 259 files (3,405 occurrences); the two fate words on 422 lines in
  109 files, overlapping; the act word `move` and the shorthand `line` are not greppable (§III.8). About 670 of the
  noun's lines regenerate (`record.pb.go` 205, agents 235, difftest goldens 133, `schema.sql` 45, simulator goldens
  40, `docs/vocabulary.md` 8); about 110 are history and stay; about 1,600 are edited by hand or by a type-aware
  rename, 939 of them in 95 test files.
  One epoch (25→26): every live run and scratch record is refused until migrated. One review of a diff near the size of
  `5dfddb60`'s. Main takes about 18 first-parent commits a day (55 in three days), and of those 21 touched `debate.js`, 27
  the generated agents and 29 the difftest goldens (`hist.sh`), so the branch conflicts within hours (R2).
- **Why it should be the last.** The 09-28 rename chose between two words already in the system by taking the
  schema's. This one is chosen against a collision census (§III.2) and a ruling per word (§III.1), and three things
  hold it: the terms registry entry with GATED bans on the outgoing noun and the outgoing fates, run by
  `TestVocabularyProse` over every seat-facing text; the name and its reasons recorded here, where the 09-28 choice has
  no record; and `inquiry` being the word the report, the id letter and the research vocabulary already used — the
  rename removes a split, it does not open one. What does **not** hold it: nothing gates a fifth rename, and the plan
  claims no such gate.

## II. Technical Context

Go 1.27 (`GOTOOLCHAIN=local`), one module at `T`; Node for the engine script and simulator. The record is SQLite,
append-only.

- **Schema and epoch.** `T/internal/record/recordpb/record.proto` is the source: message and field names become
  table and column names, enum values become the words stored and typed, `(means)` becomes help text and the
  `enum_*` tables. `go -C scripts run ./protogen` writes `record.pb.go` and `record.proto.sha256`. The stored views are
  hand-written SQL in `T/internal/record/recordsql/views.go`. `record.EventSchema` is generated from
  `F/requirements.json` (`eventSchema: 25`) by `go -C scripts run ./schemagen`. The epoch is compared at `setup` and at
  every open of a database the binary did not create (`requireDeclaredSchema`, with tables, columns and view
  definitions); a mismatch either way is refused with a pointer to `migrate`. A view-only change does not bump the
  epoch, and a stale stored view is refused at open all the same.
- **Migrate.** `feov-record migrate --from <old> --to <new>` replays an old record through the current write path.
  The registry is **keyed by the old word, one hop to the current word** (`Entries()`,
  `T/internal/record/migrate/entries.go:15`; `eraEntries()`, `eraentries.go`), not chained by epoch. Three more
  tables carry renames: `renamedBodyTables` (`sqlitesource.go:228`, an old body TABLE read under the word it
  recorded), `valueRenames` (`valuerenames.go:19`, an enum WORD, keyed by the **current** enum's full name), and
  `avenueArm` (`eraentries.go:211`, a motion's arm table and a ruling's enum column). `remap.go:74–109` lists every
  id, reference, prose and data field by `Message.field`; `TestEveryStringFieldIsInOneList`
  (`remaplists_test.go:51`) fails a field no list names, and
  `TestEveryArchivedRunBehindThisBinaryIsRefusedUnmigrated` (`unmigratedarchive_test.go:39`) holds the refusal of
  every archive behind the binary. A migrated id is `sha256(source hash ‖ archived id)` under the
  kind's letter (`remap.go:173–182`), so it does not depend on the kind's NAME. Migrate reads tables, never views
  (`sqlitesource.go:201`).
- **What is generated from what.**

  | Generator | Source | Output | Check |
  |---|---|---|---|
  | `scripts/protogen` | `record.proto` | `record.pb.go`, `record.proto.sha256` | conflicts resolve by re-running it |
  | `scripts/schemagen` | `F/requirements.json` | `recordpb.EventSchema` | `schemagen -check` |
  | `T/devcmd/enginegen` | the record's roster, enums and `record.Plan` | `debate.js:106–125` and `:357–411` | gate `enginegen` |
  | `scripts/agentgen` | `scripts/agentgen/src/**`, the tool's `manual` | `F/agents/*.md` (11), `F/skills/{research-protocol,adversarial-audit}/SKILL.md` | `agentgen -check` |
  | `scripts/vocabdoc` | `T/internal/terms/terms.json` | `F/docs/vocabulary.md` | gate `vocabdoc` |
  | `UPDATE_GOLDENS=1 go test` / `scripts/golden -review` | the binary | `T/internal/difftest/testdata/*.golden` (22 files, 5,505 lines), `recordsql/testdata/schema.sql`, `recordpb/testdata/payload-keys.txt`, `migrate/testdata/archived_*.golden`, `dashboard/testdata/*.golden`, `F/tests/simulator/testdata/prompt-*.golden` | gate `golden`; node suites |

- **Gates that bear on a rename.** `TestVocabularyProse` (`T/integration/surface/vocabulary_prose_test.go:47`) runs
  every GATED ban of `terms.json` over constitutions, prompt text, help pages, seat-facing Go string literals, the
  proto's annotation text, the protocol skill and both READMEs; `F/docs/*.md` is not scanned. Today it BANS
  `\binquir(y|ies)\b` and `\blines? of inquiry\b` (`terms.json:935–942`). `TestNoSeatPromptGrowsPastItsCeiling`
  (`T/internal/seatprobe/promptsize_test.go:92`) counts bytes for four seats (§III.9).
  `TestDetailsDoNotRestateTheFlagsCobraPrints` (`T/internal/cli/help_test.go:120`) refuses a help detail naming a
  flag cobra prints. `TestTheSeatPromptsNameNoVerb` and `TestNoPromptGrowsItsCommandCatalogue` refuse a command path
  in a prompt. `TestReasonIsNamesOnlyProseVerbs` refuses a `ReasonIs` key no prose verb carries.
  `TestEveryPassBlockerKindAgreesOnEverySurface` pins the blocker kinds. `archaeology` reads ADDED lines of
  agent-facing files (`scripts/archaeology/sweep.go:39–53`; `plans/` is out of scope, `sweep_test.go:52`), so a
  renamed line re-arms it on wording the line already held. `rulesweep` needs `Rule-Class:` and `Sibling-Sweep:`
  trailers on the commit. `decisions` reads `F/docs/seat-command-triggers.md`. `pluginparity` fails a non-test package
  outside `T/devcmd/` importing `repotree`. `deadcode` fails any function nothing reaches.
- **Constraints.** No backwards compatibility: no alias, dual reader or read-time fallback; `migrate` is the only
  upgrader and holds every translation. Versions do not move in this PR. No network, secret or privilege is involved.

## III. Proposed Changes

### III.1 Rulings (gblock, 2026-10-13)

| # | Question | RULED |
|---|---|---|
| N1 | The noun | `avenue` → **inquiry**: `show inquiries`, `inquiries.md`, tables, events, keys and the chair's acts follow |
| N2 | The proposal's text | `line` → **question**: flag `--question`, field `question` |
| N3 | The act that says what became of it | `avenue move` → **`inquiry record`**: one verb, every state, `--as <fate> --reason …` |
| N4 | The fates | `concluded` → **confirmed**; `pursued` → **underway**; `refuted`, `abandoned`, `declined`, `deferred` unchanged |
| N5 | Kept | the verb `propose` and the state `proposed`; the fate's prose flag `--reason`; the id letter `Q-`; `--hypothesis`, `--method`; the three report headings; `## Open questions`; the title's `**Question:**` line; the grade-motion ruling `rejected`; the `inquest` verb |
| N6 | Delivery | ONE pull request, epoch 25 → 26, one translation per changed word in `migrate`, each tested on an archived run |
| N7 | Accepted residues | (a) a migrated `concluded` reads `confirmed` whatever its reason says — 3 of 25 resting rows state a refutation, and migrate cannot read prose; (b) rows that rest at `pursued` because nothing ended them read `underway` on finished reports — 143 of 267 across the 20 archives, 88 of them in the ten SQLite records of 2026-09-02…09-11, where `concluded` did not exist; (c) the report prints `**Question:**` with two referents: the run's brief under the title, and each inquiry's row |

### III.2 The naming census the rulings rest on

`census6.sh`, over `T`, `debate.js`, `F/agents`, `F/skills` and `scripts/agentgen/src`.

| New word | Existing uses | Verdict |
|---|---|---|
| `inquiry`, `Inquiry` | 37 lines in 12 files, every one a migrate old-word key, a quoted 2026-08-17 incident (`T/internal/cli/root.go:288–297`, `refusalteaches_test.go:201–205`) or the ban itself | free; the old-word keys become identity (§III.5) |
| `question` | no flag, field, JSON key or enum value. `questionLine = "**Question:**"` and the layout row `Heading: "Question"` (`T/internal/report/layout.go:110,130`, the run's brief); `## Open questions`; the Catechism | residue N7(c) |
| `underway` | 0 hits in the repository | free |
| `confirmed` | plain English only; no enum value, field, flag or key | free |
| `record` (as a verb) | no verb, group or flag of that name (`Use: "record…"` → 0). The noun is everywhere: the tool is `feov-record`, the store is "the record", `seat.Records` marks a verb's event | ruled N3; R6 |
| `inquest` | a verb group on the bench and the chair, 62 lines in 18 Go files; it reads the record, and sits beside `inquiry` on the chair's root help | kept N5; R7 |
| kept because taken: `open` | the gap state; "an open avenue — proposed or pursued"; `## Open questions` | why `proposed` stays |
| kept because taken: `finding`, `outcome`, `settled` | event types, verbs, views and flags of their own (`terms.json:637,650`; `DocketRuling.settled`) | why the fate's prose stays `--reason` |

### III.3 Carrier census (consumer census for every contract below)

Command per row: `git grep -niI avenue -- <path> | wc -l` and `-l` for files (`census7.sh`). **Complete when
re-running the noun census surfaces no path this table omits.**

| Carrier | Path | L/F | Changes | How |
|---|---|---|---|---|
| Schema | `T/internal/record/recordpb/record.proto` | 50/1 | yes | by hand, from §III.4 |
| Generated code | `record.pb.go` | 205/1 | yes | `protogen` |
| Stored views | `T/internal/record/recordsql/views.go` | 20/1 | yes | by hand (2 views) |
| Schema goldens | `recordsql/testdata/schema.sql`, `recordpb/testdata/payload-keys.txt` | 45/1, 1/1 | yes | regenerate, read |
| Migrate | `T/internal/record/migrate/*.go` non-test | 28/6 | yes; **holds old words on purpose** | by hand, §III.5 |
| Record folds | `T/internal/record/*.go` non-test | 210/22 | yes | type-aware rename, then literals |
| CLI | `T/internal/cli/**` non-test `.go` | 105/17 | yes | by hand |
| Help pages | `T/internal/cli/seat/help/{propose,move,review,dispatch,mint}.md` | 15/5 | yes; `move.md` → `record.md` | by hand |
| Report, view, scorecard, consistency, seatprobe, flags | `T/internal/{report,view,scorecard,consistency,seatprobe,flags}` non-test | 45/6, 32/1, 15/1, 11/1, 32/3, 10/3 | yes | rename, then literals |
| Other Go | `anchor`, `graph`, `diagnostics`, `refusalcheck`, `reportvoice`, `enumhelp` | 1–2 each | yes (comments and one id-kind row) | by hand |
| Go tests | `*_test.go` | 939/95 | yes | compiler, then literals |
| — of them the release fuzz | `T/releasegate/fuzz/` | 167/3 | yes | **by hand** (§III.6 step 8) |
| Engine prompts | `F/skills/research-protocol/scripts/debate.js` | 20/1 | yes; 1 line (`:399`, blocker kind) is inside the generated region | by hand; `enginegen` |
| Skill sources, constitutions | `scripts/agentgen/src/**` | 15/6 | yes | by hand |
| Generated agents and skills | `F/agents/*.md`; `F/skills/**/SKILL.md` | 235/11; in 19/4 | yes | `agentgen` |
| Skill references | `F/skills/research-protocol/references/{report_template,run-directory}.md` | in 19/4 | yes | by hand |
| Goldens | `T/internal/difftest/testdata/` | 133/22 | yes | regenerate, §III.6 step 7 |
| Simulator | `F/tests/simulator/` (`debate.test.mjs` 29 lines, 9 prompt goldens, 1 fixture) | 69/11 | yes | by hand; regenerate |
| Terms | `T/internal/terms/terms.json`; `F/docs/vocabulary.md` | 7/1; 8/1 | yes; **holds the old word on purpose** (the ban) | by hand; `vocabdoc` |
| Live docs | `F/README.md:9`; `F/docs/seat-command-triggers.md` rows 77, 88, 89, 120, 122 (the live verb and trigger cells) | 1; 5 | yes | by hand |
| Files named for it | 13 (`git ls-files \| grep -ci avenue`) | — | `git mv` | list in §III.4 |
| Test functions named for it | 30 (`git grep -nIE 'func Test[A-Za-z0-9_]*Avenue'`) | — | yes | by hand; one, `TestRedCannotProposeAAvenue`, already carries the 09-28 article defect |

**Not carriers** (verified, `census3.sh`, `census7.sh`): `plugins/gray-area`, `plugins/prosthetic-conscience`,
`plugins/sleeper-service`, `F/hooks`, `T/internal/capture`, `T/cmd`, `T/devcmd` and `run-archive/` hold 0 lines of
`avenue` and 0 of `inquir`. `scripts/` holds it only under `agentgen/src`.

**Do not rename** — history, data, or another sense:

| Path | L/F | Why |
|---|---|---|
| `plans/**` (this file aside) | 74/14, of them `plans/historical/` 32/4 | design records, dated |
| `law/precedents.md`, `law/proposed/class-assertion-without-inquiry--…md` | 4/2 | holdings; the filename is a record |
| `feov-memory/{blue-scorecard,protocol-class-registry}.md` | 5/2 | appended history; the scorecard holds `lines_of_inquiry`, `thin_inquiry_reasons` and `thin_avenue_reasons` rows from three eras |
| `research/**`, `T/research/**` | 5/4, 8/1 | the corpus: reports and records as written |
| `F/docs/{duty-docket-preregistration,seat-surface-naming,why-a-seat-stops}.md`; the retired and CUT cells of `seat-command-triggers.md` (`:120` "`merge avenue-rule` (retired)", `:121`, `:213`) | 10/3, 3 | measurement notes quoting verbs as typed then |
| `T/internal/tessocr/testdata/…/reading.golden:26`; `T/internal/fetchcache/backends.go:866` | 1, 1 | an OCR reading; a paper's title |
| `record.proto:238` (`avenue_rule`, a retired event type; introduced by `d51e06e9`) and `T/integration/surface/promptverbs_test.go:22–23,97` (`merge rule-avenue`, `avenue-rule`) | 4 | the authentic word of July 2026 |
| `T/internal/cli/root.go:288–297`, `refusalteaches_test.go:201–205` | — | the verbatim 2026-08-17 incident (`inquiry-support`, `motion inquiry rule`); it gains a clause saying the verbs are that date's (R8) |
| Plain English: "we pursued X" (`record.proto:1279`), "concluded" in 16 non-concept sentences (`census5.sh` lists all 35 fate-word lines that do not name the noun; 16 are plain, 6 quote a measurement, 13 are the concept) | — | another sense |

**Quoted history is resolved line by line, never by replacement.** The worklist is
`git grep -nIiE 'avenue|\b(concluded|pursued)\b' -- plugins scripts ':!*.golden' ':!*.pb.go' ':!*testdata*' ':!F/agents/*' ':!F/docs/*' ':!F/tools/research/*' | grep -E 'MEASURED|[Mm]easured|20[0-9]{2}-[0-9]{2}-[0-9]{2}|universe-m[0-9]+|#[0-9]{3,}|\b[0-9]+ of [0-9]+\b|verbatim'`
— 24 noun lines in 19 files and 4 fate lines (`hist2.sh`). For each: `git log -S'<the phrase>' -- <file>` finds the
commit that wrote it. A quote written before 2026-09-28 with another word was falsified by `5dfddb60` (which changed
1,288 lines holding `inquir` or `direction`, 414 of them comments) and is restored from `5dfddb60^`; a quote
authentic in `avenue` stays; a sentence in the code's present voice is renamed. Each resolved line goes into the
PR's S1 list with its commit.

### III.4 The name table

Numbers, positions and every unlisted name are unchanged.

**[MODIFY] `record.proto` (then `protogen`).**

| Kind | Old | New |
|---|---|---|
| event type | `EVENT_TYPE_AVENUE = 3` (`avenue`) | `EVENT_TYPE_INQUIRY = 3` (`inquiry`) |
| event type | `EVENT_TYPE_AVENUE_REVIEW = 32` (`avenue_review`) | `EVENT_TYPE_INQUIRY_REVIEW = 32` (`inquiry_review`) |
| enum | `AvenueStatus`, `AVENUE_STATUS_*` | `InquiryStatus`, `INQUIRY_STATUS_*` |
| enum value | `AVENUE_STATUS_PURSUED = 2` (`pursued`) | `INQUIRY_STATUS_UNDERWAY = 2` (`underway`) |
| enum value | `AVENUE_STATUS_CONCLUDED = 6` (`concluded`) | `INQUIRY_STATUS_CONFIRMED = 6` (`confirmed`) |
| enum values | `PROPOSED = 1`, `DEFERRED = 3`, `DECLINED = 4`, `ABANDONED = 5`, `REFUTED = 7` | same words under the new prefix |
| enum | `AvenueRuling`, `AVENUE_RULING_{ENDORSED,OUT_OF_SCOPE,TOO_THIN}` | `InquiryRuling`, `INQUIRY_RULING_*` (words unchanged) |
| enum value | `MOTION_SUBJECT_AVENUE = 3` (`avenue`) | `MOTION_SUBJECT_INQUIRY = 3` (`inquiry`) |
| enum value | `ABOUT_KIND_AVENUE = 2` (`avenue`) | `ABOUT_KIND_INQUIRY = 2` (`inquiry`) |
| message, `Event` arm | `Avenue avenue = 44` | `Inquiry inquiry = 44` |
| field | `Avenue.avenue_id = 1` | `Inquiry.inquiry_id = 1` |
| field | `Avenue.line = 2`, `(sql).flag: "reason"` | `Inquiry.question = 2`, `(sql).flag: "question"` |
| fields | `hypothesis = 3`, `method = 4`, `status = 5`, `supersedes_status = 6`, `reason = 7` | unchanged |
| message, `Event` arm | `AvenueReview avenue_review = 51` | `InquiryReview inquiry_review = 51` |
| message, `Motion` arm | `AvenueMotion avenue = 12`; `AvenueMotion.avenue_id = 1` | `InquiryMotion inquiry = 12`; `inquiry_id = 1` |
| `MotionRule` arm | `AvenueRuling avenue = 12` | `InquiryRuling inquiry = 12` |
| `(means)` text | 11 annotations: the seven `InquiryStatus` meanings, the two event types', the motion subject's and the about-kind's | rewritten under §III.8; each is seat-facing help |

**[MODIFY] SQL** (derived from the proto, except the views).

| Old | New |
|---|---|
| tables `avenue`, `avenue_review`, `motion_avenue` | `inquiry`, `inquiry_review`, `motion_inquiry` |
| tables `enum_avenue_status`, `enum_avenue_ruling` | `enum_inquiry_status`, `enum_inquiry_ruling` |
| columns `avenue.avenue_id`, `avenue.line`, `motion_avenue.avenue_id`, `motion_rule.avenue` | `inquiry.inquiry_id`, `inquiry.question`, `motion_inquiry.inquiry_id`, `motion_rule.inquiry` |
| rows `enum_event_type`: `avenue`, `avenue_review`; `enum_motion_subject`: `avenue`; `enum_about_kind`: `avenue`; `enum_avenue_status`: `pursued`, `concluded` | `inquiry`, `inquiry_review`; `inquiry`; `inquiry`; `underway`, `confirmed` |
| view `avenue_state` (`views.go:793`), columns `avenue_id`, `line`, `avenue_ruling` | view `inquiry_state`, columns `inquiry_id`, `question`, `inquiry_ruling` |
| view `motion_answers` (`views.go:756`), column `avenue` | column `inquiry` |
| comments `views.go:338` ("a line of avenue's proposal") and `:787` ("A AVENUE, whole") — two 09-28 defects still live | written fresh |

**[MODIFY] CLI surface** — each row is a contract; its consumers are the 63 authored and 161 test lines of
`census8.sh` (`git grep -nE 'avenue move|avenue propose|avenue review|motion avenue|show avenues|--view avenues'`).

| Old | New |
|---|---|
| `blue avenue propose --reason <line> [--hypothesis] [--method]` | `blue inquiry propose --question <question> [--hypothesis] [--method]` |
| `blue avenue move --id <Q-…> --as <fate> --reason <account>` | `blue inquiry record --id <Q-…> --as <fate> --reason <account>` |
| `chair avenue review --reason …` | `chair inquiry review --reason …` |
| `motion avenue rule --id <Q-…> --as endorsed\|out_of_scope\|too_thin` | `motion inquiry rule …` |
| `show avenues` (and `--json`: `{"avenues":[{"id","line",…}]}`) | `show inquiries`; `{"inquiries":[{"id","question",…}]}` |
| `--about-kind section\|avenue\|gap` | `--about-kind section\|inquiry\|gap` |
| `--as pursued\|concluded` | `--as underway\|confirmed` |
| result keys `avenue_id`, `line`, `moved` (`blue/avenue.go:193–196`) | `inquiry_id`, `question`; `moved` deleted (no reader: `git grep -nE '"moved"\|\.moved\b'` → the declaration only) |
| flag shape `avenue-id` (`flags/shapes.go:177`), id kind `avenue` (`anchor/anchor.go:70`) | `inquiry-id`, `inquiry`; the prefix stays `Q-` |
| blocker kind `avenue_review` (`record/passblockers.go:39`; `debate.js:399`, generated) | `inquiry_review` |
| scorecard metrics `avenues`, `thin_avenue_reasons` (`scorecard/scorecard.go:451–473`) | `inquiries`, `thin_inquiry_reasons` (F3) |
| consistency check `avenue-md` (`consistency/consistency.go:506`) | `inquiry-md` |
| run file `avenues.md`, nav `Avenues`, title `avenues` (`report/docs.go:54,136`) | `inquiries.md`, `Inquiries`, `inquiries` |
| headings of that file: `## pursued (n)`, `## concluded (n)` | `## underway (n)`, `## confirmed (n)` (from the enum) |
| help template value `{{.AvenueSectionsRule}}`; `debate.js` constants `AVENUE_SECTIONS`, `avenueClause` | `{{.InquirySectionsRule}}`; `INQUIRY_SECTIONS`, `inquiryClause` |

**[NEW] `flags.Question = "question"`**, registered on `propose` with `flags.Text` (the free-text registration
`--hypothesis` uses, which attaches the quoting rule) and `seat.Require`. `propose` stops being a `seat.Prose` verb:
its `ReasonIs` row is [DELETE]d (`seat/seat.go:941`), `satisfiedByAnyOf["line"]` (`seat.go:537`) and the payload-key
row `"line": Reason` (`flags/names.go:430`) become `"question": {flags.Question}` and `"question": Question`, and
`record.go:1307`'s refusal names `--question` through `flags.ForPayloadKey`. The 12 call sites that pass `--reason`
to `propose` (`census7.sh`) change with it.

**[MODIFY] Go identifiers**, by package (`out/idents.txt`: 48 distinct identifiers).

| Package | Old → New |
|---|---|
| `recordpb` (generated) | every `Avenue*`, `AVENUE_*`, `GetAvenueId`, `GetLine`, `MotionRule_Avenue`, `Motion_Avenue` → the §III.4 proto names |
| `record` | `Avenue` (and `.Line`) → `Inquiry` (`.Question`); `AvenuesOf`, `StaleAvenuesOf`, `AvenueStatuses`, `AvenueStatusNames`, `AvenueStatusOf`, `AvenueRuling`, `AvenueReviewDueOf`, `AvenueJSON`, `AvenuesJSON`, `AvenuesJSONOf`, `RequireAvenueRef`, `AvenueExists`, `StruckAvenueTexts`, `BlockerAvenueReview` → the same with `Inquiry`/`Inquiries`; `MotionSubjects` and `enums.go:278` key `"avenue"` → `"inquiry"`; `ExistingProposalByText(…, line, …)` → parameter `question` |
| `report` | `AvenueSections`, `AvenueSectionsRule`, `FileAvenues`, `avenues()` → `InquirySections`, `InquirySectionsRule`, `FileInquiries`, `inquiries()`; `ran()` reads `underway`, `confirmed`, `refuted` |
| `view` | `AvenueBody`, `avenueMD`, view key `"avenues"` → `InquiryBody`, `inquiryMD`, `"inquiries"` |
| `flags` | `AvenueID` → `InquiryID` |
| `cli/blue`, `cli/chair` | `newAvenue`, `newAvenuePropose`, `newAvenueMove`, `newAvenueReview`, `avenueResult`, `avenueReviewResult` → `newInquiry`, `newInquiryPropose`, `newInquiryRecord`, `newInquiryReview`, `inquiryResult`, `inquiryReviewResult` |
| `seatprobe` | `Avenue`, `Boards().Avenues`, the verb strings in `boards.go:235,367,504,505` and `seatprobe.go:161–197` |
| `migrate` | `avenueArm` → `inquiryArm`; the field lists of `remap.go:79–109` |
| `releasegate/fuzz` | 25 identifiers (`avenueMoves`, `avenueFates`, `nextAvenueStatus`, `avenueRulingOracle`, …) → `inquiry…`; `avenueMoves` → `inquiryRecords` |

**Files.** [`git mv`] `T/internal/cli/{avenue_test,avenuevoice_test}.go`, `cli/blue/avenue.go`,
`cli/chair/{avenue.go,avenuereview_test.go}`, `record/{avenue.go,avenue_test.go,avenuerulingerror_test.go}`,
`record/migrate/{avenuefate_test,avenuerename_test}.go`, `scorecard/avenuezero_test.go`, `view/avenue_test.go`,
`releasegate/fuzz/avenueoracle_test.go` → the same names with `inquiry`. [DELETE]
`T/internal/cli/seat/help/move.md`; [NEW] `help/record.md`. [NEW] `T/devcmd/archdump/main.go` (F4). [MODIFY]
`F/requirements.json` (`eventSchema: 26`).

```
F/
├── requirements.json                      [MODIFY] eventSchema 26
├── agents/*.md                            [MODIFY] generated (11)
├── skills/research-protocol/
│   ├── SKILL.md                           [MODIFY] generated
│   ├── references/{report_template,run-directory}.md   [MODIFY]
│   └── scripts/debate.js                  [MODIFY] four prompt constants; the generated regions by enginegen
├── tests/simulator/                       [MODIFY] debate.test.mjs, 9 prompt goldens, 1 fixture
└── tools/
    ├── devcmd/archdump/main.go            [NEW] (F4)
    ├── internal/cli/blue/inquiry.go       [git mv] propose, record
    ├── internal/cli/chair/inquiry.go      [git mv] review
    ├── internal/cli/seat/help/record.md   [NEW]; move.md [DELETE]
    ├── internal/record/inquiry.go         [git mv]
    ├── internal/record/recordpb/record.proto, record.pb.go, record.proto.sha256   [MODIFY]
    ├── internal/record/recordsql/views.go [MODIFY] inquiry_state, motion_answers
    ├── internal/record/migrate/{entries,eraentries,valuerenames,sqlitesource,remap}.go   [MODIFY]
    ├── internal/terms/terms.json          [MODIFY]
    └── releasegate/fuzz/inquiryoracle_test.go   [git mv]
scripts/agentgen/src/**                    [MODIFY] 6 sources
```

### III.5 Migration: one translation per changed word

Old-word keys map straight to the current word, so a key already named `inquiry` becomes identity and is deleted,
and a key named `avenue` becomes a new translation. No word passes through two hops. Counts: `scan.py`, `final.py`.

| # | Old word | Where, with counts | Maps to | Mechanism |
|---|---|---|---|---|
| M1 | body table `avenue` | 14 SQLite archives (2026-09-02 … 10-12), 540 acts | `inquiry` | [NEW] `renamedBodyTables["avenue"]`, entry `"avenue": mapped("inquiry", ren{"avenue_id": "inquiry_id", "line": "question"}, …)` |
| M2 | event `line-of-inquiry` | 6 JSONL runs (2026-08-22/23), 405 events | `inquiry` | [MODIFY] `eraentries.go:58`: `mapped("inquiry", ren{"line": "question"}, …)`; its `inquiry_id` rename is identity, deleted |
| M3 | field `line` | 275 proposal acts in all 20 archives | `question` | the `ren` of M1 and M2 |
| M4 | status `pursued` | 508 acts in 16 archives | `underway` | [NEW] `valueRenames["feov.record.v1.InquiryStatus"]` |
| M5 | status `concluded` | 33 acts in 4 archives (wave-a, m19, m20, wave-b) | `confirmed` | same row |
| M6 | body table `avenue_review` | 4 archives, 28 events | `inquiry_review` | [NEW] `renamedBodyTables` row and entry |
| M7 | body table `inquiry_review` | 10 archives, 40 events | identity | [DELETE] `entries.go:25` and `sqlitesource.go:230` — the table is current again |
| M8 | event `inquiry-support` | 6 JSONL runs, 277 events | `inquiry_review` | [MODIFY] `eraentries.go:63` target; drops unchanged |
| M9 | motion subject `direction` | 5 rulings (quadratic-formula) | `inquiry` | [MODIFY] `valueRenames` `MotionSubject: {"direction": "inquiry", "avenue": "inquiry"}` |
| M10 | motion subject `avenue` | 14 rulings (m20 10, wave-b 4) | `inquiry` | same row |
| M11 | era motion subject `inquiry` | 5 JSONL rulings (4 + 1) | `inquiry` | [MODIFY] `eraentries.go:193–199`: the arm and enum renamed |
| M12 | ruling column `motion_rule.direction`, `motion_rule.avenue`; arm table `motion_direction`, `motion_avenue` (`avenue_id`) | columns: the 19 rulings of M9, M10; arm tables: 0 rows in every archive | arm `inquiry`, field `inquiry_id` | [MODIFY] `avenueArm` → `inquiryArm`, renaming both old arms and the arm's id field |
| M13 | `about_kind = inquiry` | 5 findings, 3 mints (b, b5, b6, b7) | identity | [DELETE] `valuerenames.go:24` |
| M14 | `about_kind = avenue` | 33 findings, 4 mints (wave-a, m19, m20, wave-b) | `inquiry` | [NEW] `AboutKind: {"avenue": "inquiry"}` |
| M15 | a proposal's `reason` equal to its `line` | 275 of 275 proposal acts | dropped (F1) | in the M1 and M2 entries: a row with no `supersedes_status` whose `reason` equals its `line` carries no reason forward |

Not translated, and why: `supersedes_status` holds `"1"` or nothing in all 945 acts, never a status word; the
`enum_*` tables and views are not read by migrate; ids keep their value (§II). `remap.go`'s lists change name only.

**Tests, each on an archived run** ([MODIFY] `inquiryfate_test.go`, `inquiryrename_test.go`,
`shardarchive_test.go`; names final in the PR):

| Translation | Archive | Asserts |
|---|---|---|
| M1, M3 | `2026-09-11_is-91-prime-b9` | 26 acts arrive as `inquiry`; every proposal's `question` equals the archived `line` |
| M2, M8 | `2026-08-22_sqlite-schema-source` | 57 `line-of-inquiry` and 18 `inquiry-support` events arrive as `inquiry` and `inquiry_review` |
| M4 | b9 | the 11 `pursued` acts arrive `underway`; no migrated act holds `pursued` |
| M5 | `2026-10-11_is-91-prime-m20` | the 7 `concluded` acts arrive `confirmed`; `abandoned` and `declined` are untouched; none arrives `refuted` |
| M6, M7 | m20; b9 | 5 and 6 reviews arrive as `inquiry_review` |
| M9, M12 | `2026-09-02_quadratic-formula` | its 5 `direction` rulings arrive under subject `inquiry` with the ruling in the `inquiry` arm |
| M10 | m20 | its 10 rulings likewise |
| M11 | `2026-08-22_is-7-prime-run2` | its 4 era rulings likewise |
| M12 (arm table) | b9 with a planted `motion_direction` row, as `avenuerename_test.go:40–48` plants a ruling today | the arm and its `inquiry_id` arrive; no archive holds such a row, and the test says so |
| M13, M14 | `2026-09-09_is-91-prime-b`; `2026-10-12_is-91-prime-wave-b` | 2 and 7 findings (and wave-b's 3 mints) arrive `about_kind = inquiry` |
| M15 | b9 | no migrated proposal holds a reason; every later act's reason is byte-equal to the archive's |
| all | `TestEveryArchivedRunRenders`, `everyarchive_test.go` | 20 runs, 0 refusals; the three `archived_*.golden` files are expected unchanged (they hold the report's text and gaps, no inquiry word) — a changed line is read, not regenerated past |

### III.6 Order of work, and the check that closes each step

By kind, never a repository-wide replace. No step begins until the one before it passes its check. One commit per
step, so a rebase conflict is confined to a kind.

| # | Step | The check | The 09-28 failure it prevents |
|---|---|---|---|
| 0 | Baseline at `e8f6ec9c`: commit `T/devcmd/archdump` (F4); dump every archive's assembled set to `~/.claude/scratch/inquiry-rename/before/`; build `~/.claude/scratch/inquiry-rename/bin/feov-record-25`; save the censuses | 20 directories, 0 `REFUSALS` lines, 0 `ASSEMBLE_ERROR` | — |
| 1 | Schema: edit `record.proto` from §III.4; `protogen`; `schemagen` to 26; rewrite the two views by hand | `go -C T build ./internal/record/recordpb/`; `UPDATE_GOLDENS=1 go -C T test -count=1 ./internal/record/recordsql/ ./internal/record/recordpb/`, then read every changed line of `schema.sql`: `SELECT name FROM sqlite_master GROUP BY name HAVING count(*) > 1` is empty, and `inquiry_state` reads FROM `"inquiry"` | a view renamed onto the acts table |
| 2 | Go identifiers: `gopls rename` per identifier of §III.4, package by package (type-aware; it cannot touch a comment or a literal) | `go -C T build ./... && go -C T vet ./...` | phrases broken across literals |
| 3 | Migrate: §III.5, with its tests first | `go -C T test -count=1 ./internal/record/migrate/`; then S6 for M1–M15 | a translation nothing exercises |
| 4 | CLI: verbs, flags, JSON keys, refusals, string literals in the 17 CLI files and the record folds | `go -C T test -count=1 ./internal/cli/... ./internal/record/...`; each of `inquiry propose`, `inquiry record`, `inquiry review`, `motion inquiry rule`, `show inquiries` rendered with `--help` and read against the page saved in step 0 | a help page losing an instruction |
| 5 | Terms: the `inquiry` entry and the ban flip (§III.7) | `go -C T test -count=1 -run TestVocabularyProse ./integration/surface/` **fails, and its failure list is the worklist for step 6** — every seat-facing `avenue` left | a carrier nobody listed |
| 6 | Seat-facing prose: help pages, `(means)`, the four prompt constants, constitutions, skill sources, references — sentence by sentence (§III.8), inside the byte budget (§III.9); then `enginegen`, `agentgen`, `vocabdoc` | `TestVocabularyProse`, `TestNoSeatPromptGrowsPastItsCeiling -v`, `TestTheSeatPromptsNameNoVerb`, `TestNoPromptGrowsItsCommandCatalogue`, `TestDetailsDoNotRestateTheFlagsCobraPrints` pass; `agentgen -check` and the `enginegen` gate are clean | "a line of avenue"; the lens prompt past its ceiling |
| 7 | Goldens: for each golden, write the PREDICTED file — the old golden with the §III.4 word map applied by a script, in scratch — regenerate (`UPDATE_GOLDENS=1 go -C T test -count=1 ./...`; `go -C scripts run ./golden -review`), and `diff` predicted against regenerated | every line the two disagree on is read and explained in the PR (a reworded sentence, a re-wrapped column, a block the `manual` golden hoisted into SHARED and renumbered); lines that agree are the rename and nothing else | a golden diff too large to read |
| 8 | The release fuzz, by hand: the three files, the drive (`fuzz_test.go:1938–1988`, `:5121–5163`), the coverage key per recording verb, the oracle's messages | §V.5; and the mutation of #1362 repeated: take `refuted`, then `underway`, out of `nextInquiryStatus`'s rotation — the coverage gate turns red each time | a fuzz that never drives the renamed verb |
| 9 | Simulator and live docs: `debate.test.mjs`, the prompt goldens, `README.md:9`, the four live rows of `seat-command-triggers.md`, the references | `node --test F/tests/simulator/*.test.mjs`; gates `decisions`, `mjsparity` | — |
| 10 | History: resolve the §III.3 worklist line by line; run S1 and S2; paste both lists into the PR | S1's file list equals the "on purpose" rows; no line of it is in the code's present voice | verbatim measurements falsified |
| 11 | §V in full | S3–S7 | — |

### III.7 Terms registry and what the reports show

**[MODIFY] `terms.json`, then `go -C scripts run ./vocabdoc`.** The entry `avenue` (`:924`) becomes **inquiry**:
"An inquiry is one question the research could take up — proposed by blue with the hypothesis that would be true if
it paid off, recorded as underway, confirmed, refuted, deferred, declined or abandoned, ruled on by the chair, and
reviewed against the report once per epoch." Bans: `\bavenues?\b` GATED (variant "avenue (an inquiry)");
`\blines? of inquiry\b` stays GATED; `\binquir(y|ies)\b` is [DELETE]d; `direction` stays REGISTRY-ONLY; the fates
`pursued` and `concluded` per F2. Collisions: `direction` unchanged; `line` becomes "an inquiry is never called a
line, and its text is its question"; [NEW] `question` — "the run's brief prints as **Question:** under the report's
title, and Open questions is the report's own section; neither is an inquiry"; [NEW] `record` — "`inquiry record` is
the act that says what became of an inquiry; the record is the store it writes to"; [NEW] `inquest` — "reads the
record's debate, motions or telemetry; it is not an act on an inquiry". The entry at `:838` ("an avenue's state
keeps the word") follows. One plain-English use trips the new ban and is reworded: "BOTH AVENUES" in
`scripts/agentgen/src/blue-researcher.md:53,58` (what can be read and what can be computed) becomes "BOTH ROUTES".

**What S4 allows to differ.** In `report.md`, the `**Fate:**` word on rows that rested at `pursued` or `concluded`.
In the renamed `inquiries.md`: its file name, nav label and title; the `## pursued`, `## concluded` headings and the
`path:` words; nothing else. In `README.md` and the link bars: the file name and label. In `run.md` and the
dashboard: the two scorecard metric names and the `{"pursued":…}` keys. `debate.md` and `judgments.md`: the verb and
subject words where an act is named. No archive rests at `proposed` (`grep -c '^## proposed'` over the baseline dump
→ 0), so M15 changes no archived document.

### III.8 The words grep cannot find: `line` and `move`

`\blines?\b` is mostly text lines, and `\bmov(e|ed|es|ing)\b` is mostly gaps and grades. The 09-28 commit calls the
shorthand "a hundred places". Three passes, each with a closed list:

1. **The compiler lists the readers.** Renaming `Avenue.Line` and `GetLine()` breaks every site that reads the field
   (78 `GetLine()|Line:` hits repository-wide, of which the inquiry's fail to compile). At each, the comment above
   and the literal beside it are read.
2. **Every sentence of the files that name the concept is read.** The list is closed by command (`census7.sh`): 17
   authored seat-facing files — `debate.js` (29 `line` words), `record.proto` (30), `help/propose.md` (6),
   `terms.json` (4), `help/dispatch.md` (2), the two references (3), the six `agentgen` sources (5), `README.md` (1),
   and `help/{move,review,mint}.md` (0) — and 46 Go files whose literals or comments hold the word beside the noun
   (largest: `report/assemble.go` 48, `seat/verbs.go` 43, `report/layout.go` 36, `seat/selector.go` 30,
   `record/record.go` 22, `view/view.go` 22, `report/layoutview.go` 22, `record/avenue.go` 20). Each `line` is
   classified **inquiry** (the thing: "a declined line", "settles the line"), **question** (its text: "the line you
   propose"), or **unrelated** ("one line on why", "the `**Outcome:**` line"), and each `move`/`moved`/`owes a move`
   (104 hits in the authored seat-facing files, 49 lines of "owes a move"-shaped phrases) as **record** ("say what
   became of it", "its fate is recorded") or unrelated. The classification table is in the PR.
3. **The rendered text is read, not its source.** After step 6 the eleven regenerated `F/agents/*.md` (each inlines
   its seat's `manual`) and the nine simulator prompt goldens are searched for `\blines?\b` and
   `\bmov(e|ed|es|ing)\b`, and every hit is classified again. A sentence assembled from two constants shows its
   seam only here.

[NEW] `TestTheInquiryPagesCallNothingALine` (`T/internal/cli/`): the rendered `inquiry propose`, `inquiry record`
and `inquiry review` pages and the proto's `InquiryStatus` meanings hold no `\blines?\b` and no `\bmov(e|ed|es|ing)\b`
— those pages have no other use for either word, so a zero there is checkable where a repository-wide ban is not.

### III.9 The byte budget for prompts

`TestNoSeatPromptGrowsPastItsCeiling -v` at `e8f6ec9c` (`out/promptsize.txt`). `inquiry` is 1 byte longer than
`avenue`, `inquiries` 2 longer than `avenues`, `underway` 1 longer than `pursued`, `confirmed` equal to `concluded`,
`question` 4 longer than `line`, `inquiry` 3 longer.

| Seat (gated) | Now / ceiling | Concept words in the prompt | Delta | After |
|---|---|---|---|---|
| `red-lens-evidence` | 12,289 / 12,300 | **none**: the evidence lens takes the ledger clause; the steelman clause goes to the logic and dark-side lenses only (`debate.js:1011`), and `prompt-red-lens-evidence*.golden` holds 0 of `avenue`, `pursued`, `concluded` | 0 | 12,289 |
| `red-chair` | 10,131 / 10,200 | `debate.js:1030`: "VOTE EVERY AVENUE", "answer every line", "once per line", "BLUE'S AVENUES", "becomes of the avenue" | +1 +3 +3 +2 +1 = **+10** word for word; "answer each against that pass" in place of "answer every line…" makes it +1 | ≤ 10,141 |
| `blue-respond` | 11,170 / 15,200 | `:1035`: 2 × avenue, 1 × AVENUES, 2 × line, 1 × pursued | ≤ +12 | ≤ 11,182 |
| `judge` | 7,056 / 7,900 | none | 0 | 7,056 |

Prompts no ceiling measures (stated so the growth is seen): the logic and dark-side lens prompts gain about 23 bytes
from the steelman clause (8 × avenue, 3 × avenues, 3 × pursued, 2 × line → inquiry); the frontier and lane prompts
SHRINK, because "ITS LINE IS THE QUESTION the avenue asks" and "THE LINE IS THE QUESTION AND HOLDS NO PREDICTED
RESULT" lose their first half once the field is called the question. No ceiling is raised. The implementer pastes
the `-v` table before and after.

### III.10 Forks for the owner — not ruled, each with a recommendation

| # | Fork | Options | Recommendation |
|---|---|---|---|
| F1 | `propose` writes one text into two fields today (`blue/avenue.go:116–117`), and the report hides the copy with `a.Reason != a.Line` (`report/assemble.go:684`). With `--question`, does a proposal still fill `reason`? | **(a) No**: a proposal holds a question and no reason; migrate drops the copy (M15, 275 acts); the guard is deleted; `show inquiries` stops printing a proposed inquiry's question twice; the thin-reason detector skips `proposed` (it owes no reason). **(b) Yes**: no behaviour moves; the field named `reason` goes on holding a question, and the guard stays | **(a).** The copy is the defect N2 names, kept under a new flag. Cost: one detector condition, one translation, one visible change mid-run (`show inquiries`), none in an archived document |
| F2 | What holds the old FATE words out? `concluded` and `pursued` are also plain English, so `\bpursued\b` cannot be banned | **(a)** GATED bans on the quoted forms only — `` `pursued` ``, `` `concluded` ``, `'pursued'`, `--as pursued`, `--as concluded`. **(b)** REGISTRY-ONLY: a note, no gate | **(a).** A seat-facing fate is always quoted; plain prose is untouched |
| F3 | The scorecard metric `thin_inquiry_reasons` returns a name retired on 09-28, so `feov-memory/blue-scorecard.md` will hold that name before 09-28 and after this PR, with `thin_avenue_reasons` between; `avenues` → `inquiries` likewise splits the series | **(a)** rename, accept a three-era history. **(b)** keep the two metric names as they are | **(a)**: a metric named for a banned noun is the half-state; no reader aggregates the history by name (`git grep -l thin_avenue_reasons` → the writer, two dashboard goldens, the history file, one historical plan) |
| F4 | The before/after measurement of S4 needs a harness that migrates and assembles every archive. #1362 ran one from scratch (`~/.claude/scratch/avenue-row/zzarchive_test.go`, copied into the tree and removed) | **(a)** commit it as `T/devcmd/archdump` (about 60 lines, the directory CLAUDE.md gives such a harness), in its own first commit. **(b)** keep using the scratch file | **(a)**: it is the second change in two days to need it, and S4 is not reproducible by a reviewer otherwise |
| F5 | Main moves about 20 commits a day through the same prompts, goldens and generated agents (R2) | **(a)** a declared freeze on FEOV merges from the final rebase to the merge — a few hours. **(b)** no freeze; rebase and regenerate until one lands | **(a)**, announced to the other sessions when §V.2 is green; (b) is the fallback and costs a full §V per attempt |

## IV. Risk & Mitigation

Graded likelihood × impact × cost to mitigate (L/M/H).

| # | Risk | L × I × cost | Mitigation, and the step or gate that carries it |
|---|---|---|---|
| R1 | The 09-28 failure modes recur: a view on a table's name; a phrase split across literals; a measurement falsified; a check passing for the wrong reason | H × H × M | Step 1's `sqlite_master` check and read of `schema.sql`; step 2 is type-aware and touches no literal; §III.8 pass 3 reads rendered text; §III.3's worklist resolves quotes by commit; S6 deletes each translation. No step uses a repository-wide replace |
| R2 | Another session merges mid-rename: 55 first-parent commits in three days, 21 on `debate.js`, 27 on the agents, 29 on the difftest goldens | H × M × M | One commit per kind; generated files are never resolved by hand — on conflict take main's side of the SOURCE, re-apply that step, and re-run `protogen`, `enginegen`, `agentgen`, `vocabdoc`, `UPDATE_GOLDENS=1`; merge main in rather than rebase (generated agents conflict in every commit otherwise, as #1362 found); after each merge, step 5's gate and S1 are re-run, because main may have ADDED the old word; F5's freeze for the last hours. Before any push: `gh pr view <n> --json state` |
| R3 | A sentence added to main with the old word after the branch's sweep | M × M × L | `TestVocabularyProse` fails it once the ban is flipped (seat-facing); S1 fails it elsewhere; both run in CI on the merge |
| R4 | The epoch bump refuses every live run, every scratch record and every universe at epoch ≤ 25 | H × M × L | That is the rule working. The PR names `feov-record migrate --from <old> --to <new>`; no run is launched from a branch between the merge and the next release; a running session keeps the binary it loaded. §V.4 proves the refusal's wording and the migration on a real epoch-25 record |
| R5 | A translation is wrong and silent: a fate word migrate does not know is refused loudly, but `question` filled from the wrong column, or M15 dropping a real reason, is not | M × H × L | M1/M3's test compares `question` with the archived `line` byte for byte; M15's test asserts every non-proposal reason is byte-equal; S4 diffs every assembled document of 20 archives |
| R6 | `inquiry record` reads as a noun phrase beside "the record" and `feov-record` (ruled N3) | M × L × L | The group's Short and the page's menu line open with the act ("say what became of an inquiry you proposed"); the `record` collision is registered (§III.7); the seat-probe board that expects the verb (`seatprobe/boards.go:367`) shows whether a seat finds it. If the smoke shows a seat typing `inquiry move`, that is a finding for the owner, not a reason for an alias |
| R7 | `inquest` beside `inquiry` on the chair's root help: one is read for the other | M × L × L | Registered as a collision; the two Shorts are read side by side in step 4's `--help` review; the refusal of an unknown subcommand already names the seat that holds an act (`teachUnknownSubcommand`) |
| R8 | Quoted history reads as current: the 2026-08-17 incident says `inquiry-support` and `motion inquiry rule` — after this change the second is a LIVE verb again | H × M × L | `root.go:288–297` and `refusalteaches_test.go:201–205` keep the quote and gain one clause dating the verbs; step 10 reads each resolved quote for the same trap |
| R9 | Accepted residues mislead a reader of an old report (N7 a, b) | H × L × H | Accepted, as ruled. They exist only on migrated archives; the PR body states both with the counts (3 of 25; 143 of 267) |
| R10 | The chair prompt passes its ceiling (59 bytes spare after +10) | L × M × L | §III.9's wording; gate `TestNoSeatPromptGrowsPastItsCeiling`, in `check` |
| R11 | A lens prompt no ceiling measures grows unseen | M × L × L | §III.9 states it; the nine simulator prompt goldens show every byte in the diff |
| R12 | Another plugin or an outside reader depends on a FEOV record word | L × M × L | Verified none: `git grep -ciIE 'avenue\|inquir' -- plugins ':!plugins/frank-exchange-of-views'` → 0 lines; `F/hooks`, `T/internal/capture`, `T/cmd` → 0. Re-run in §V.1 |
| R13 | `archaeology` fails a renamed line for wording it already carried (it reads added lines), or `deadcode` fails a helper the rename orphans | M × L × L | Both are loud and in `check`; an archaeology hit is fixed in the sentence, never allow-listed |
| R14 | The `manual` golden hoists any block of 80 characters repeated on two pages into SHARED and renumbers, so one reworded sentence moves many lines | M × L × M | Step 7's predicted-against-regenerated diff isolates it; each renumbering is named in the PR |
| R15 | A box under load turns a suite red or a timing false | M × M × L | Check the load before believing a red; reproduce quiet; read fail COUNTS, never an exit code through a pipe |
| R16 | `propose` leaving the prose channel loses a protection the channel gave (the quoting rule, the heredoc trim) | L × M × L | `flags.Text` attaches the same rule (`flags/register.go:23–32`); `TestEveryFreeTextVerbShowsTheQuotingRule` holds it; the trim (`Prose.Read`) is applied to `--question` in `propose` and pinned by a test that passes a value ending in a newline |

## V. Verification Plan

Before any FEOV command: `unset $(env | grep -o '^FEOV_[A-Z_]*')`; `GOTOOLCHAIN=local`. Each check names what
re-arms it. The order is the order they run in.

### V.1 Generators and censuses · re-armed by any edit to a source in §II's table

1. `go -C scripts run ./protogen && git diff --exit-code -- T/internal/record/recordpb/` → no diff.
2. `go -C scripts run ./schemagen -check` → 26; `go -C T run ./cmd/feov-record --schema` → `26`.
3. `go -C scripts run ./agentgen -check`; `go -C scripts run ./vocabdoc && git diff --exit-code -- F/docs/vocabulary.md`;
   `go -C scripts run ./check -only enginegen` → clean.
4. S1 and S2's commands → the pasted lists. R12's command → 0.

### V.2 Suites · re-armed by any commit

1. `go -C T test -count=1 -timeout 90m ./... > ~/.claude/scratch/inquiry-rename/logs/tools.log 2>&1`, then
   `grep -c '^FAIL' …` → 0 and `grep -c '^ok' …` equal to the count at `e8f6ec9c`.
2. `go -C scripts test -count=1 ./...` → 0 FAIL. `node --test F/tests/simulator/*.test.mjs` → `fail 0`.
3. **S6, by hand.** Delete, one at a time, each of M1–M15's registry rows, the `question` rename, the two
   `InquiryStatus` rows, the `inquiryArm` field rename; run `go -C T test -count=1 ./internal/record/migrate/`. Each
   deletion fails a named test; the PR lists the pairs. A deletion nothing fails is a missing assertion, fixed
   before merge.
4. `go -C scripts run ./check > …/check.log 2>&1` → the summary line's failed count is 0. `qlty` and `actionlint`
   run only in CI: read the PR's checks by fail count.

### V.3 Real data: every archived run, before and after (S3, S4) · re-armed by any change under `T/internal/{record,report,view}`

1. Before, at `e8f6ec9c`: `go -C T run ./devcmd/archdump -archives run-archive -out ~/.claude/scratch/inquiry-rename/before`.
2. After, on the branch: the same with `-out …/after` → 20 directories; `cat …/after/*/REFUSALS | wc -l` → 0;
   `ls …/after/*/ASSEMBLE_ERROR` → none.
3. Apply the word map to a COPY of `before` (`wordmap.py`, kept in scratch and pasted in the PR): the file name
   `avenues.md` → `inquiries.md`; in every document the nav label, title and link; in `inquiries.md` and on `**Fate:**`
   and `path:` lines the words `pursued` → `underway`, `concluded` → `confirmed`; the metric names and verb words of
   §III.7. Then `diff -r …/before-mapped …/after` → **no output**. A line that differs is a behaviour change: it is
   explained by §III.7, or it is a defect.
4. Read by hand, in `after`: `2026-10-11_is-91-prime-m20/report.md`'s three inquiry sections and `inquiries.md`
   whole; `2026-09-11_is-91-prime-b9/inquiries.md` (a pre-epoch record: `## underway (7)`).

### V.4 Real data: a record at epoch 25 is refused, then migrates (S5) · re-armed by any change to the epoch or to `migrate`

1. With `feov-record-25` (step 0), make a real epoch-25 record: extract `run-archive/2026-10-12_is-91-prime-wave-b.tar.gz`
   to `~/.claude/scratch/inquiry-rename/e24/` and run `feov-record-25 migrate --from …/e24/<run> --to …/e25/wave-b`.
2. Branch binary: `feov-record show --run …/e25/wave-b --seat-id red-chair inquiries` → **refused**; the message
   names epoch 25, epoch 26 and `migrate`, and names no old word as a verb.
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
   → `grep -c 'DATA RACE'` → 0.
3. If either is red: run it on `origin/main` before blaming the branch.

### V.6 The engine runs · after merge

`scripts/universe.sh build`; prove the universe carries the change — `grep -rl 'inquiry record' <the installed
plugin copy>` prints the chair's and blue's agent definitions and `grep -rli avenue` prints only `docs/` and
`terms.json`; then `scripts/universe.sh run` with the engine's `--smoke` (one lane, `haiku` on both tiers,
`--mint-budget 1 --k-max 2`), streamed, never `--output-format json`. Read in it: the frontier seat's first
`inquiry propose` carries `--question` and is accepted first time; at least one `inquiry record --as confirmed|refuted`;
no refusal naming `avenue`, `move`, `--reason` on a propose, `pursued` or `concluded`
(`feov-record dashboard`, and the run's `ops log`); the chair's `inquiry review` and any `motion inquiry rule`;
`inquiries.md` in the report set; the outcome and wall clock against the last smoke. A seat that typed an old word
is reported with the sentence that taught it.

### V.7 The auditor gate

`/prosthetic-conscience:plan-audit plans/avenue-becomes-inquiry.md` — approved on the second consecutive PASS, or by
gblock once F1–F5 are ruled and every remaining gap is `local`. Its NOTES travel into the PR, with: the S1 and S2
lists, the §III.8 classification table, the §III.9 size table before and after, the S6 pairs, and the fresh censuses
of §III.3 re-run on the PR's base.
