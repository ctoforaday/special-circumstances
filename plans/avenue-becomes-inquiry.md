# An avenue becomes an inquiry: the noun, its question, its fates and the act that records them

> STATUS: **PROPOSED 2026-10-13; every fork ruled (§III.1, gblock, 2026-10-13/14).** Awaits the plan-auditor gate.

`F` is `plugins/frank-exchange-of-views`, `T` is `F/tools`; `L/F` is lines / files. Each census is a command run from
the repository root after S1's assignments. Line numbers and counts are at `54e0ff06` (`origin/main`, 2026-10-14,
event-schema epoch 25); the implementer re-runs each on the PR's base and the PR carries the figures. **`E0` is
`origin/main`'s event-schema epoch at the attempt and `E` is `E0 + 1`** (§III.9 step 3). Archive figures come from
`python3 -I ~/.claude/scratch/inquiry-rename/plan/final.py <dir>` over the 20 extracted `run-archive/` tarballs; the
PR pastes the script and its output.

## I. Summary & Goals

**The problem.** Blue records each approach the research could take, and what became of it. Five names on that
concept say something other than what the thing is.

| Today | What it fails to say | Evidence |
|---|---|---|
| `avenue` (the noun) | It holds a question, a hypothesis and a finding, and names none. The report labels its text `**Question:**` and its id letter is Q | `T/internal/flags/shapes.go:174`; `plans/rulebook-audit.md:109–111` |
| `line`, filled by `--reason` | That it is a question. `--reason` names two different texts on the two acts of one group, and `propose` writes its question into the field `reason` as well | 275 of 275 archived proposals hold `reason == line` (`final.py`) |
| `concluded` | Whether the answer was yes | on universe-m22 two seats read it two ways and the dispute held the gate (`T/internal/record/recordpb/record.proto:431`) |
| `pursued` | That it is still open: the one running state reads as a finished past tense | a prompt has to explain it (`F/skills/research-protocol/scripts/debate.js:225`) |
| `avenue move` | What re-recording the same state is: `move --as pursued` on a pursued avenue moves nothing | `T/internal/cli/seat/help/move.md:7` |

**The objective.** One noun, **inquiry**; its text is its **question**; it is **underway** while open, and
**confirmed** or **refuted** when it ran and answered; the act that says what became of it is **`inquiry record`**.
After one pull request the live tree accepts and prints only these words, at epoch `E`, one `migrate` translation each.

### Success criteria

| # | Criterion | Command | Today | Target |
|---|---|---|---|---|
| S1 | No old word and no earlier name in the live tree | under this table | noun 2,216 lines / 228 files; fate words 366 / 116; earlier names 593 / 267 | noun: no output; fate words and earlier names: the English list, byte for byte |
| S2 | Inside `migrate`, an old word or earlier name is a string literal and nothing else | under this table | 41 lines; 2 file names | no output |
| S3 | Every archive migrates | `go -C T test -count=1 -run 'TestEveryArchivedRunRenders\|TestEveryArchivedRunMigrates' ./internal/record/migrate/` | 20 runs, 0 refusals at 25 | 20 runs, 0 refusals at `E` |
| S4 | Assembled archive documents differ only in the renamed words | §V.3 | — | the mapped diff prints nothing |
| S5 | A record at `E0` is refused, then migrates | §V.4 | — | the refusal names `E0`, `E` and `migrate`; 0 refusals after |
| S6 | No translation is unexercised | §V.2 step 3 | — | every deleted registry row fails a named test |
| S7 | Gates | `go -C scripts run ./check`; the release fuzz (§V.5) | — | fail count 0; `fuzzed 40 debate runs · 0 failed` |
| S8 | The engine runs | §V.6, before merge | — | a VERIFIED or CEILING outcome whose report set holds `inquiries.md` |

S1 — every tracked file outside §III.2's not-the-live-tree list; both filters test the PATH field only:

```
F=plugins/frank-exchange-of-views; T=$F/tools
export NOT="^(plans|law|research|run-archive)/|^feov-memory/[a-z]+-scorecard[.]md\$|^$T/(research|internal/record/migrate|internal/tessocr/testdata/corpus)/"
export GEN="^$F/agents/|[.]golden\$|[.]pb[.]go\$|/testdata/|^$F/docs/vocabulary[.]md\$|^$F/skills/(research-protocol|adversarial-audit)/SKILL[.]md\$"
live() { awk -F: '$1 !~ ENVIRON["NOT"]'; }
git grep -lIi 'avenue' | live                                                   # the noun: no output
git grep -IiE 'conclud|pursu' | live | sort                                     # the fate words, every inflection
git grep -IiE '\bdirections?\b|\blines?[-_ ]of[-_ ]inquir' | live | sort        # the earlier names
```

The noun's grep prints nothing: plain English holding it is reworded too (§III.6). The fate pattern is a substring:
366 lines, 134 generated, 131 in tests, 101 authored (split on the path by `GEN`, then `_test.go`/`.mjs`). The
earlier names are those the registry entry for this concept lists (`T/internal/terms/terms.json:924–957`): 593 lines,
112 generated; `direction` is the concept at `T/internal/cli/root.go:76` (rendered into two constitutions),
`record.proto:257`, `cli/motion/verbs.go:300` and `report/motions.go:32`, and English elsewhere. **Every line of the
second and third commands and of §III.2's split-phrase run is read (N14, §III.3)**: the concept is renamed, English
stays as written. After the rename they print the English list, pasted in the PR; S1 holds when `diff` against it
prints nothing (no line numbers, so a moved line stays equal). The excluded set (`$1 ~`) is those prefixes only.

S2 — inside `migrate` an old word or earlier name is a translation key, or a test's archived input or planted row:
a string literal, never an identifier, file name or comment. Each line's literals (double-quoted, or backtick closed
on that line) are removed FIRST and the rest, comment whole, is grepped; a raw string spanning lines holds current
names (`remap.go`'s field lists) and is grepped whole:

```
OLD='avenue|conclud|pursu|\bdirections?\b|\blines?[-_ ]of[-_ ]inquir'
git grep -nIiE "$OLD" -- $T/internal/record/migrate | perl -ne 'my ($p,$c)=/^([^:]+:\d+:)(.*)$/;
  $c=~s{("(?:[^"\\]|\\.)*"|`[^`]*`|//.*$)}{substr($1,0,2) eq "//" ? $1 : ""}ge; print "$p$c\n" if $c=~/'"$OLD"'/i'
git ls-files $T/internal/record/migrate | grep -iE "$OLD"
```

Today: 41 lines (6 an earlier name in a comment; `sqlitesource.go:219`'s English is reworded, since in `migrate` a
hit means a key) and two file names. S1 and S2 run by hand, at landing and on every attempt; neither is a gate.

### The value case

- **What it buys.** The five rows above stop being taught. Two have a measured cost: a gate held on universe-m22 by
  two readings of one fate, and one flag meaning two texts in one group. The report already prints "Question /
  Hypothesis / Fate"; the record, the verbs, the flag, the help pages and the prompts then say the same words.
- **What it costs.** The noun is on 2,216 lines in 228 files: 643 regenerate and 1,573 are edited, 947 of them in
  tests; one review of a diff near 4,000 lines each way. Reading no tool shortens: 366 fate-word lines, 593
  earlier-name lines, 1,032 `line` and 590 `move` lines (§III.3). The narrations N7 deletes keep their reasoning in
  the commit history only. One epoch: every live run and scratch record is refused until migrated. The landing: two
  or three merge attempts and one 45-minute smoke, two to four hours from the finished tree (§III.9). No test and no
  generator is added for the fate words (N9).

## II. Technical Context

Go 1.27 (`GOTOOLCHAIN=local`), one module at `T`; Node for the engine script and simulator; SQLite, append-only.

- **Schema and epoch.** `T/internal/record/recordpb/record.proto` is the source: message and field names become
  tables and columns, enum values the words stored and typed, `(means)` help text and the `enum_*` tables. Stored
  views are hand-written SQL (`T/internal/record/recordsql/views.go`). The epoch (`F/requirements.json`,
  `eventSchema`) is compared at `setup` and at every open of a database the binary did not create; a mismatch either
  way is refused with a pointer to `migrate`.
- **Migrate** replays an old record through the current write path. Its registry maps each old word one hop to the
  current word (`Entries()`, `T/internal/record/migrate/entries.go:15`; `eraEntries()`); `renamedBodyTables`
  (`sqlitesource.go:228`), `valueRenames` (`valuerenames.go:19`) and `avenueArm` (`eraentries.go:211`) carry the
  other renames. A migrated id (`remap.go:173–182`) does not depend on the kind's name. It reads tables, never views.
- **Generators.** Each is re-run, never hand-resolved.

  | Command | Source → output | Check |
  |---|---|---|
  | `go -C scripts run ./protogen` | `record.proto` → `record.pb.go`, `record.proto.sha256` | `git diff --exit-code` after a re-run |
  | `go -C scripts run ./schemagen` | `F/requirements.json` → `recordpb.EventSchema` | `schemagen -check` |
  | `go -C T run ./devcmd/enginegen` | roster, enums, `record.Plan` → `debate.js:106–125`, `:364–423` | gate `enginegen` |
  | `go -C scripts run ./agentgen` | `scripts/agentgen/src/**`, the tool's `manual` → `F/agents/*.md` (11), two `SKILL.md` | `agentgen -check` |
  | `go -C scripts run ./vocabdoc` | `T/internal/terms/terms.json` → `F/docs/vocabulary.md` | gate `vocabdoc` |
  | `UPDATE_GOLDENS=1 go -C T test -count=1 <pkg>` | the binary → `difftest/testdata/*.golden` (22), `recordsql/testdata/schema.sql`, `recordpb/testdata/payload-keys.txt`, `dashboard/testdata/*.golden` | gate `golden`; `go -C scripts run ./golden -review` |
  | `UPDATE_GOLDENS=1 node --test plugins/frank-exchange-of-views/tests/simulator/*.test.mjs` | `debate.js` → `F/tests/simulator/testdata/prompt-*.golden` (14) | the same without `UPDATE_GOLDENS` → `fail 0` |

- **Gates that bear on the rename** (under `T`). `TestEveryEnumValueNamedInAPromptIsAccepted`
  (`integration/surface/promptverbs_test.go:926`) fails a `--<flag> <value>` no enum bound to that flag holds, so an
  old `--as` value in a prompt fails once the enum loses it. `TestVocabularyProse`
  (`integration/surface/vocabulary_prose_test.go:47`) runs the terms registry over `vocabularySources(t)`, failing a
  banned variant and a mask or allow that did no work. `TestNoSeatPromptGrowsPastItsCeiling` counts bytes for four
  seats. `archaeology` reads ADDED lines of agent-facing files, so a renamed line re-arms it on wording it already
  held; `rulesweep` needs its two commit trailers; `decisions` reads `F/docs/seat-command-triggers.md`.
- **CI.** The `hooks` workflow on a pull request takes about 5.5 minutes from push to every check done; its longest
  job runs `go test ./...` whole on `T`; Windows runs `-short`. The release fuzz is not in it.
- **Constraints.** CLAUDE.md, Change discipline: the live code accepts and prints only the new form; `migrate` is
  the only upgrader. Versions do not move. No network, secret or privilege; no step uses `gopls` (not installed).

## III. Proposed Changes

### III.1 Rulings (gblock; N1–N16 2026-10-13, N9, N15, N17, N18 2026-10-14)

| # | RULED |
|---|---|
| N1 | The noun `avenue` → **inquiry**: `show inquiries`, `inquiries.md`, tables, events, keys and the chair's acts follow |
| N2 | The proposal's text `line` → **question**: flag `--question`, field `question` |
| N3 | The act `avenue move` → **`inquiry record`**: one verb, every state, `--as <fate> --reason …` |
| N4 | The fates `concluded` → **confirmed**, `pursued` → **underway**; `refuted`, `abandoned`, `declined`, `deferred` unchanged |
| N5 | Kept: the verb `propose` and the state `proposed`; the fate's prose flag `--reason`; the id letter `Q-`; `--hypothesis`, `--method`; the three report headings; `## Open questions`; the title's `**Question:**` line; the grade-motion ruling `rejected`; the `inquest` verb |
| N6 | ONE pull request, ONE epoch bump (to `E`), one translation per changed word in `migrate`, each tested on an archived run |
| N7 | Every occurrence of an old word in the live tree is renamed. A passage whose only job is to say what something was called, or to narrate a run or incident in the old words, is DELETED; its why goes in the commit. What is not the live tree is §III.2's list and nothing else |
| N8 | The terms registry defines `inquiry` and holds no ban, no note about an earlier name and no new collision note. No word is added to any roster of retired names |
| N9 | FATE WORDING TAKES THE LEAST MACHINERY. No new fate test: once the old words leave the enum, `TestEveryEnumValueNamedInAPromptIsAccepted` fails an old `--as` value in a prompt. The `debate.js` sections constant stays a hand copy, pinned byte for byte to the Go sentence by `report/assemble_test.go:440`; the static fate lists are reworded to need no list; the passages that teach particular fates are renamed by the sweep and held by review. Nothing new is generated (§III.6) |
| N10 | A proposal holds a question and NO reason: `propose` stops filling `reason`, migrate drops the copied text (M15), the report's guard on the copy is deleted, the thin-reason detector skips `proposed` |
| N11 | The scorecard metrics are `inquiries` and `thin_inquiry_reasons` |
| N12 | The archive before/after harness is committed as `T/devcmd/archdump`, in the PR's first commit |
| N13 | No freeze. The rename lands by merge-main-in, regenerate, re-verify, each time main moves (§III.9). Merging on green CI is the owner's standing instruction for this workstream |
| N14 | READ EVERY ONE. The sweep takes every occurrence of each old fate word in the live tree; each is read and classified. A fate is renamed; plain English stays as written and is not reworded. The classification is pasted in the PR and never kept in the tree (S1) |
| N15 | THE SMOKE RUNS BEFORE MERGE, on the FIRST COMPLETE TREE. A later attempt re-arms it only by a hand change to a seat-facing file; a clean merge of main plus regeneration does not (§III.9, §V.6) |
| N16 | `F/tests/simulator/testdata/m16-chair6-relay.json`, a captured chair relay, is a TEST INPUT in the live tree and not a record: its text is rewritten in the new words like everything else |
| N17 | The plan carries what the implementer needs once; measured figures only the PR can settle (the attempt window, attempt and smoke counts, prompt sizes after, the deleted passages) go to the PR |
| N18 | NO ARCHAEOLOGY; THE HELP PAGE TEACHES THE VERB. Each renamed verb's `--help` teaches it completely on its own: what each flag holds, every fate's meaning, what is refused and what to do next. Prompts, constitutions and skills keep the DUTY — what a seat must do and why — and carry no command mechanics: mechanics the sweep finds there move to the help page. No renamed text says what anything was called or narrates an incident (§III.6, §III.8 steps 4 and 6) |

**The naming census the rulings rest on.** `git grep -nIi inquir -- $T $F/skills $F/agents scripts/agentgen/src` → 35
lines in 11 files: `migrate` keys, one narrated incident (N7), two registry bans (N8).
`git grep -nIE '"question"|--question|flags\.Question' -- "$T/*.go" ':!*_test.go'` → nothing;
`git grep -ci underway -- $F scripts` → 0; `confirmed`: English only; `git grep -n 'Use: *"record' -- $T` → 0;
`git grep -nIi inquest -- "$T/*.go"` → a bench and chair verb group, 67 lines in 18 files (R5).

### III.2 Carrier census (the consumer census for every contract below)

Per row: `git grep -niI avenue -- <path>`, lines and `-l` files. **Complete when S1's noun grep prints nothing.**

| Carrier | Path | L/F | How it changes |
|---|---|---|---|
| Schema, generated code | `T/internal/record/recordpb/record.proto`; `record.pb.go` | 50/1; 205/1 | by hand from §III.4; `protogen` |
| Stored views, schema goldens | `recordsql/views.go`; `recordsql/testdata/schema.sql`, `recordpb/testdata/payload-keys.txt` | 20/1; 46/2 | by hand (2 views); regenerate, read |
| Migrate | `T/internal/record/migrate/*.go` | 28/6 non-test | by hand, §III.5; old words stay only as literals (S2) |
| Record folds | `T/internal/record/*.go` non-test | 210/22 | identifier rename, then literals and comments |
| CLI and help pages | `T/internal/cli/**` non-test `.go`; `cli/seat/help/{propose,move,review,dispatch,mint}.md` | 107/17; 13/5 | by hand; `move.md` → `record.md`; each page teaches its verb (N18) |
| Report, view, scorecard, consistency, seatprobe, flags | `T/internal/{report,view,scorecard,consistency,seatprobe,flags}` non-test | 145/15 | rename, then literals and comments |
| Other Go | `anchor`, `graph`, `diagnostics`, `refusalcheck`, `reportvoice`, `fetchcache` | 1–2 each | by hand |
| Go tests; of them the release fuzz | `*_test.go`; `T/releasegate/fuzz/` | 918/92; 167/3 | compiler, then literals and comments; the fuzz by hand (§III.8 step 8) |
| Engine prompts | `F/skills/research-protocol/scripts/debate.js` | 20/1 | by hand; the blocker kind at `:409` by `enginegen` |
| Constitutions and skills | `scripts/agentgen/src/**` → `F/agents/*.md`, `F/skills/**` | 15/7 → 212/11; 17/3 | sources and references by hand; `agentgen` |
| Goldens, simulator | `T/internal/difftest/testdata/`; `F/tests/simulator/` | 125/22; 70/12 | regenerate (§III.8 step 7); `debate.test.mjs` and the relay fixture (N16) by hand |
| Terms | `T/internal/terms/terms.json`; `F/docs/vocabulary.md` | 7/1; 8/1 | by hand (§III.6); `vocabdoc` |
| Docs, class registry | `F/README.md`; `F/docs/{seat-command-triggers,seat-surface-naming,duty-docket-preregistration,why-a-seat-stops}.md`; `feov-memory/protocol-class-registry.md:110` | 1; 17/4; 1 | by hand |
| Names | 13 files (`git ls-files \| grep -i avenue \| grep -vc '^plans/'`); 30 test functions (`git grep -nIE 'func Test[A-Za-z0-9_]*Avenue'`) | — | `git mv`; by hand |

**Not carriers**: `git grep -ciIE 'avenue|inquir' -- plugins ':!plugins/frank-exchange-of-views'` → 0; the same over
`$F/hooks $T/internal/capture $T/cmd $T/devcmd` → 0. **Not the live tree** — the whole list, S1's exclusions:

| Path | Why |
|---|---|
| `T/internal/record/migrate/` — the string keys a translation maps from, and the tests that read archived runs | that is what `migrate` is; S2 holds it to literals |
| `run-archive/` | the gzipped record of each captured run |
| `plans/` (this file included) | dated design records: `plans/README.md` keeps a plan as the record of what changed |
| `law/` | holdings and proposed classes, as ruled |
| `research/`, `T/research/` | the research corpus: reports and records as written |
| `feov-memory/*-scorecard.md` | scorecard history, appended run over run at capture |
| `T/internal/tessocr/testdata/corpus/` | a third party's page as OCR read it (`FIFTH AVENUE BANK`, one line): not this concept, not this repository's text |

**Passages deleted under N7.** A narration's old word and its marker are often on adjacent lines, and an old phrase
can itself break across lines (`record.proto:959–960`: "line / of inquiry"), so the finder reads each line JOINED to
the one before it, comment leaders stripped:

```
git ls-files -- plugins scripts feov-memory ':!*.golden' ':!*.pb.go' ':!*testdata*' ":!$F/agents" ":!$T/research" \
    ":!$T/internal/record/migrate" ':!feov-memory/*-scorecard.md' | xargs awk 'FNR==1 {p=""}
  { c=$0; sub(/^[ \t]*(\/\/|#|\*|--)?[ \t]*/,"",c); j=p" "c; l=tolower(j)
    if (l ~ /avenue|inquir|conclud|pursu|\<directions?\>|\<lines? of inquir/ && j ~ /MEASURED|[Mm]easured|20[0-9][0-9]-[0-9][0-9]-[0-9][0-9]|universe-m[0-9]+|#[0-9][0-9][0-9]|\<[0-9]+ of [0-9]+\>|verbatim|retired|CUT|used to|became/) print FILENAME":"FNR
    p=c }'
```

— 248 line pairs in 91 files, `record.proto:960` among them (the position is the pair's second line). Each hit is
a sentence in the present voice, renamed, or a narration, deleted whole and never renamed in place; an instruction a
deleted passage carried is restated in the present tense where it stood (R7). **S1's split-phrase run** is the same
reading over `git ls-files | live` with the marker test replaced by `tolower(p" "c) ~ r && tolower(p) !~ r &&
tolower(c) !~ r` (`r`: S1's patterns as one regex): 2 today, `record.proto:960` and its `record.pb.go` copy.

### III.3 The words grep cannot settle: the earlier names, `line` and `move`

A fate word or earlier name is the concept or English; `\blines?\b` is mostly text lines; `\bmov(e|ed|es|ing)\b`
mostly gaps and grades. Each sweep's file set is closed by a command; every classification table goes in the PR.

1. **The earlier names, BEFORE their ban is deleted.** S1's third command and the split-phrase run are classified at
   step 0, **concept** (→ `inquiry`) or **English**; step 5 deletes the registry entry only after this table exists.
2. **`line` and `move`, in every authored file that names the concept:**
   `for w in '\blines?\b' '\bmov(e|ed|es|ing)\b'; do git grep -lIi avenue | live | awk -F: '$1 !~ ENVIRON["GEN"]' | xargs git grep -nIiE "$w" --; done`
   → `line` 1,032 lines in 133 files (524 in 60 outside tests), `move` 590 in 114 (292 in 60). A `line` is
   **inquiry** ("a declined line"), **question** ("the line you propose") or **unrelated**; a `move` is **record**
   or unrelated. A file that says `move` of this concept and never names it is reached by passes 3 and 4.
3. **The compiler lists the readers.** Renaming `Avenue.Line` and `GetLine()` breaks every reading site; at each,
   the comment above and the literal beside it are read.
4. **The rendered text is read, not its source.** After §III.8 step 6 the eleven `F/agents/*.md`, the two generated
   `SKILL.md`, the fourteen prompt goldens and the relay fixture are searched for every pattern above (`line` 209,
   `move` 208 today) and each hit classified again: a sentence assembled from two constants shows its seam only here.

### III.4 The name table

Numbers, positions and every unlisted name are unchanged. **The rule:** a name holding `avenue`, `Avenue` or
`AVENUE` takes `inquiry` in the same case and position (plural `inquiries`): the 127 Go names (54 outside tests;
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
| `(means)` text | 11 annotations: the seven status meanings, the two event types', the motion subject's, the about-kind's | rewritten under §III.3 and N18: each is the help page's own teaching of that value |

**[MODIFY] CLI surface** — each row is a contract; its consumers are the lines of
`git grep -nE 'avenue move|avenue propose|avenue review|motion avenue|show avenues' | live` — 59 authored, 87 in
tests (`_test.go`, `.mjs`) and 60 generated (the path split of S1) — and every one changes.

| Old | New |
|---|---|
| `blue avenue propose --reason <line> [--hypothesis] [--method]` | `blue inquiry propose --question <question> [--hypothesis] [--method]` |
| `blue avenue move --id <Q-…> --as <fate> --reason <account>` | `blue inquiry record --id <Q-…> --as <fate> --reason <account>` |
| `chair avenue review --reason …`; `motion avenue rule --id <Q-…> --as endorsed\|out_of_scope\|too_thin` | `chair inquiry review …`; `motion inquiry rule …` |
| `show avenues` (and `--json`: `{"avenues":[{"id","line",…}]}`) | `show inquiries`; `{"inquiries":[{"id","question",…}]}` |
| `--about-kind section\|avenue\|gap`; `--as pursued\|concluded` | `--about-kind section\|inquiry\|gap`; `--as underway\|confirmed` |
| result keys `avenue_id`, `line`, `moved` (`blue/avenue.go:199–202`; `moved` is read at `:218`) | `inquiry_id`, `question`, `recorded` |
| flag shape `avenue-id` (`flags/shapes.go:177`), id kind `avenue` (`anchor/anchor.go:70`) | `inquiry-id`, `inquiry`; the prefix stays `Q-` |
| blocker kind `avenue_review` (`record/passblockers.go:39`; `debate.js:409`, generated) | `inquiry_review` |
| scorecard metrics `avenues`, `thin_avenue_reasons` (`scorecard/scorecard.go:451–473`) | `inquiries`, `thin_inquiry_reasons` |
| consistency check `avenue-md` (`consistency/consistency.go:506`) | `inquiry-md` |
| run file `avenues.md`, nav `Avenues`, title `avenues` (`report/docs.go:54,136`); its headings `## pursued (n)`, `## concluded (n)` | `inquiries.md`, `Inquiries`, `inquiries`; `## underway (n)`, `## confirmed (n)` (from the enum) |
| help template value `{{.AvenueSectionsRule}}`; `debate.js` constants `AVENUE_SECTIONS`, `avenueClause` | `{{.InquirySectionsRule}}`; `INQUIRY_SECTIONS` (a hand copy, §III.6), `inquiryClause` |

**Identifiers that are not the rule:** `.Line`, `GetLine` → `.Question`, `GetQuestion`; `newAvenueMove` →
`newInquiryRecord`; `avenueResult.Moved` → `inquiryResult.Recorded`; the fuzz's `avenueMoves` → `inquiryRecords`;
`ExistingProposalByText`'s parameter `line` → `question`; `TestRedCannotProposeAAvenue` →
`TestRedCannotProposeAnInquiry`. `report.ran()` reads `underway`, `confirmed`, `refuted`.

**[NEW] `flags.Question = "question"`**, registered on `propose` with `flags.Text` (which attaches the quoting rule)
and `seat.Require`. `propose` leaves `seat.Prose`: its `ReasonIs` row (`seat/seat.go:941`) is [DELETE]d;
`satisfiedByAnyOf["line"]` (`seat.go:537`) and `"line": Reason` (`flags/names.go:430`) become `"question"` rows;
`record.go:1307`'s refusal names `--question` through `flags.ForPayloadKey`. The 12 call sites passing `--reason` to
`propose` (`git grep -nIE 'avenue.{0,3}propose.{0,80}--reason|"propose".{0,60}"--reason"' -- plugins scripts`) follow.

**[MODIFY] A proposal holds no reason (N10).** Consumer census —
`git grep -nE '\b(a|q)\.Reason\b|\ba\.GetReason\(\)' -- $T/internal/{report,view,scorecard} $T/internal/record/avenue.go ':!*_test.go'`
→ 8 lines, with the writer:

| Site | Changes |
|---|---|
| `cli/blue/avenue.go:119–120` (the writer) | yes: `propose` writes `question` and leaves `reason` unset |
| `report/assemble.go:684` (`a.Reason != "" && a.Reason != a.Line`) | yes: the second operand is [DELETE]d |
| `scorecard/scorecard.go:434` (`st != "pursued" && len(reason) < 20`) | yes: the detector skips `proposed` and `underway` |
| `report/assemble.go:689`; `view/view.go:775–776`; `record/avenue.go:94,158,329` | no: they print or carry a reason when one exists (`StruckAvenueTexts`, `:94`, takes a proposal's question and a recorded fate's reason) |

```
F/requirements.json                               [MODIFY] eventSchema E (= origin/main's + 1, §III.9)
F/{README.md,docs,agents,skills,tests/simulator}  [MODIFY] by hand or generated, per §III.2
T/devcmd/archdump/main.go [NEW] (N12); T/internal/cli/seat/help/record.md [NEW], move.md [DELETE]
T/internal/{cli/blue,cli/chair,record}/inquiry.go [git mv]; and 10 `inquiry*_test.go` beside their packages
T/internal/record/recordpb/, recordsql/views.go, migrate/*.go, T/internal/terms/terms.json   [MODIFY]
scripts/agentgen/src/** (7), feov-memory/protocol-class-registry.md (1 line)                 [MODIFY]
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

Not translated: `supersedes_status` (`"1"` or nothing in all 945 acts); the `enum_*` tables and views, which migrate
does not read; ids (§II). `remap.go`'s lists change name only; `migrate`'s comments name current words (S2).
**Residue.** `migrate` re-reads no prose: of the 25 archived inquiries that end `concluded`, 3 state a refutation in
their reason and arrive `confirmed`. Accepted.

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

### III.6 The terms registry, the fate wording, and duty against mechanics

**[MODIFY] `terms.json`, then `vocabdoc`.** The entry at `:924` becomes **inquiry**: "An inquiry is one question the
research could take up — proposed by blue with the hypothesis that would be true if it paid off, recorded with what
became of it, ruled on by the chair, and reviewed against the report once per epoch." Its bans and collisions
(`:933–957`) are [DELETE]d after §III.3 pass 1, and nothing is added (`TestVocabularyProse` requires no entry). `:838`
reads "an inquiry's state keeps the word"; `retiredSurfaces` gains no row; "BOTH AVENUES"
(`scripts/agentgen/src/blue-researcher.md:53,58`) becomes "BOTH ROUTES".

**Fate wording, by the least machinery (N9).** The census of authored non-test lines holding three or more of the
seven status words — `W` the seven joined by `|`; `git grep -lIiE "$W" | live | awk -F: '$1 !~ ENVIRON["GEN"] && $1 !~
/(_test[.]go|[.]mjs)$/' | xargs git grep -nIiE "$W" --`, kept where `awk` counts three or more of the seven in the
lowered line — prints 23 lines in 13 files, in three kinds. Nothing new is generated and no test is added.

1. **The sections constant stays a hand copy**: `debate.js:196` (`INQUIRY_SECTIONS`) restates
   `report.InquirySections`, and `TestBluesPromptsNameTheSectionsTheToolComposes` (`report/assemble_test.go:440`)
   already holds the rendered blue-lane prompt to it byte for byte; `debate.test.mjs:694` is renamed with it.
2. **Static lists are reworded to need no list**: `cli/seat/verbs.go:292` ("every inquiry with its current fate
   and the reason for it"), `record.proto:244` ("an inquiry: its question, its hypothesis and each fate recorded on
   it"), `terms.json:925` (above), `F/docs/seat-command-triggers.md:89`, `references/report_template.md:74`,
   `report/docs.go:137` ("every inquiry on the record: the path each one took…").
3. **Passages that teach particular fates stay hand-written**, renamed by the sweep and held by review:
   `cli/seat/help/move.md:7` (→ `record.md`), the `(means)` at `record.proto:431,436,444,981`, `view/view.go:807`,
   and the prompt text `debate.js:215` and `:225` — the prompt two only as far as they state duty (below). The rest
   are code (`report/assemble.go:636`, `ran()`, pinned per fate) and comments, renamed or deleted under N7.

**Residue.** A value legal for another verb's `--as`, and an old word after `--as` in a help page, are held by the
landing sweep only; a fate word in running prose is held by review alone after merge.

**Duty against mechanics (N18).** Each sentence the sweep renames in a prompt, constitution source or skill is read
for what it is: **mechanics** (what a flag holds, what a value means, what the tool refuses or prints, which command
to type) or **duty** (what the seat must do, when, and why). Mechanics MOVE to the verb's help page — its Long, a
flag's usage, or the enum's `(means)` — unless the page already says it, and leave the prompt. The known case is
`debate.js:215`: what each fate means moves to `inquiry record`'s page through `(means)`; the prompt keeps the duty
(record each approach with the fate it had, owe no fate an entry, keep the question free of a predicted result). The
PR lists each moved sentence. No renamed text says what anything was called or narrates an incident.

### III.7 The byte budget for prompts

At `54e0ff06` (`go -C T test -count=1 -run TestNoSeatPromptGrowsPastItsCeiling -v ./internal/seatprobe/`): red-chair
8,489 / 10,200, red-lens-evidence 10,570 / 12,300, blue-respond 9,456 / 15,200, judge 5,889 / 7,900. The word
changes add a few bytes each and N18's moved mechanics remove more; no ceiling is raised; the PR pastes the table after.

### III.8 Order of work, and the check that closes each step

By kind, never a repository-wide text replace (R1); each step begins once the one before passes its check; one commit
per step. The PR merges by merge commit, so step 1 may close on a tree that does not compile, and says so.

| # | Step | The check that closes it |
|---|---|---|
| 0 | Baseline at the base: commit `T/devcmd/archdump` (N12), touching no other path; dump every archive's assembled set to `~/.claude/scratch/inquiry-rename/before/`; build `…/bin/feov-record-E0`; save each `--help` page of the group; classify the earlier names (§III.3 pass 1) | 20 directories, 0 `REFUSALS` lines, 0 `ASSEMBLE_ERROR`; the earlier-name table covers every line S1's third command and the joined-pair run print |
| 1 | Schema: `record.proto` from §III.4; `protogen`; `requirements.json` to `E` and `schemagen`; the two views by hand | `go -C T build ./internal/record/recordpb/` passes and nothing that imports it does, which is step 2's worklist |
| 2 | Go identifiers, led by the compiler. First the names under the rule, generated and hand-written alike: ONE `gofmt -l -w -r '<Old> -> <New>'` pass PER DISTINCT NAME — §III.4's census prints the 127, one per line, and a loop runs a pass for each over `git ls-files "$T/*.go" ':!*.pb.go'`; it rewrites whole identifiers, declarations and uses, and no comment or literal. A lower-case identifier with no capital after the noun (`avenue`, `avenues`) is outside the census pattern and gets its own pass; any the passes miss is a hit of S1's noun grep at step 9. Then the names outside the rule (`.Line`, `GetLine`, `Moved`, `avenueMoves`, a parameter): each declaration by hand, and `go -C T vet ./... 2>&1 \| grep -E '\.go:[0-9]+'` — which type-checks tests as well — names every reader; repeat until it prints nothing | `go -C T build ./... && go -C T vet ./...`; `UPDATE_GOLDENS=1 go -C T test -count=1 ./internal/record/recordsql/ ./internal/record/recordpb/`, then every changed line of `schema.sql` read: `SELECT name FROM sqlite_master GROUP BY name HAVING count(*) > 1` is empty, and `inquiry_state` reads FROM `"inquiry"` |
| 3 | Migrate: §III.5, tests first | the tests of §III.5's table but its last row, by `-run` on their names; S6; S2 prints nothing. The package whole — whose render tests print words step 4 moves — closes at step 4 |
| 4 | CLI: verbs, flags, JSON keys, refusals, N10, literals in the 17 CLI files and the record folds; the help pages of `inquiry propose`, `inquiry record`, `inquiry review`, `motion inquiry rule` and `show inquiries` written so each teaches its verb on its own (N18): what `--question`, `--hypothesis`, `--method`, `--id`, `--as` (every fate, with its meaning) and `--reason` hold, what is refused, and what to do next | `go -C T test -count=1 ./internal/cli/... ./internal/record/...`, migrate whole among them; each of the five pages rendered with `--help` and read twice: against step 0's page for a lost instruction, and AS A SEAT WITH NO OTHER CONTEXT — the page alone says what to type, what each value means, what will be refused and what to do next; a question the page leaves open is answered in the page |
| 5 | Terms and the fate wording: §III.6's registry entry and items 1–3; then `agentgen`, `vocabdoc` | `go -C T test -count=1 -run 'TestVocabularyProse\|TestEveryEnumValueNamedInAPromptIsAccepted' ./integration/surface/` passes, the first with no stale mask; `TestBluesPromptsNameTheSectionsTheToolComposes` holds the sections copy; `agentgen -check` clean |
| 6 | Seat-facing prose: help pages, `(means)`, the prompt constants, constitutions, skill sources, references — sentence by sentence (§III.3), each in a prompt, constitution or skill classified duty or mechanics and mechanics moved (§III.6, N18); then `enginegen`, `agentgen`, `vocabdoc` | `TestEveryEnumValueNamedInAPromptIsAccepted`, `TestNoSeatPromptGrowsPastItsCeiling -v`, `TestTheSeatPromptsNameNoVerb`, `TestNoPromptGrowsItsCommandCatalogue`, `TestDetailsDoNotRestateTheFlagsCobraPrints` pass; `agentgen -check` and gate `enginegen` are clean; each moved sentence is found on its destination page by rendering it |
| 7 | Goldens: for each, write the PREDICTED file in scratch — the old golden with §III.4's word map applied by a script — regenerate (both commands of §II's table), and `diff` predicted against regenerated | every line the two disagree on is read and explained in the PR; lines that agree are the rename and nothing else |
| 8 | The release fuzz, by hand: the three files, the drive (`fuzz_test.go:1938–1988`, `:5121–5163`), the coverage key per recording verb, the oracle's messages | §V.5; and by mutation: take `refuted`, then `underway`, out of `nextInquiryStatus`'s rotation — the coverage gate turns red each time |
| 9 | Simulator, docs, comments, tests: `debate.test.mjs`, the relay fixture (N16), the prompt goldens, `F/README.md`, `F/docs/*.md`, the class registry line, §III.2's deletions | `node --test plugins/frank-exchange-of-views/tests/simulator/*.test.mjs`; gates `decisions`, `mjsparity`, `archaeology`; **S1 holds** |
| 10 | §V.1 to §V.5 in full, then the smoke (§V.6) | S2–S8 |

### III.9 Landing without a freeze (N13), with the smoke inside it (N15)

Main touched a carrier this rename regenerates 38 times in the 72 hours before `1bd4332f`, a median 80 minutes
apart (`git log --first-parent --since='72 hours ago' --format=%ct origin/main -- <§II's generated carriers>`): two
or three 30-minute attempts, the smoke beside the first, in two to four hours. The PR records the measured figures.

**One attempt**, begun at the first push and again each time main moves a carrier:

1. `gh pr view <n> --json state` → `OPEN`; note `git merge-base HEAD origin/main`; `git fetch origin main && git
   merge origin/main` — never a rebase.
2. Conflicts: in an AUTHORED source take main's hunk and re-apply §III.4's names; in a GENERATED file take either
   side, and step 4 overwrites it.
3. **The epoch is main's plus one, on every attempt.** Main may bump the epoch for another shape change; both sides
   then make the same edit to `requirements.json` and `EventSchema`, git merges it cleanly, and two shape changes
   share one epoch. So:
   ```
   E0=$(git show origin/main:$F/requirements.json | jq .eventSchema); E=$(jq .eventSchema $F/requirements.json)
   [ "$E" -eq $((E0 + 1)) ] || echo "EPOCH MISMATCH: branch $E, origin/main $E0"
   ```
   A mismatch fails the attempt: `eventSchema` is set to `E0 + 1`, `schemagen` re-run, `feov-record-E0` rebuilt at
   the merged base, §V.4 and §V.5 re-armed. §III.5's rows still map each old word straight to the current one.
4. Regenerate, in this order: `protogen`; `schemagen`; `go -C T build ./...` (the compiler names every identifier
   main added under an old name); `enginegen`; `agentgen`; `vocabdoc`; both golden commands of §II's table.
5. Read what main ADDED, each line printed as `path:line:` in `origin/main`'s copy:
   ```
   git diff --unified=0 <the noted merge base>..origin/main | awk '/^diff --git /{h=1; next}
     h && /^\+\+\+ /{f=substr($0,7); next} /^@@ /{h=0; split($3,a,/[+,]/); n=a[2]; next}
     !h && /^\+/{print f":"n": "substr($0,2); n++}' \
     | grep -iE 'avenue|conclud|pursu|\bdirections?\b|lines?[-_ ]of[-_ ]inquir|\blines?\b|\bmov(e|ed|es|ing)\b'
   ```
   The header state keeps an added line that itself begins `+`. The reader opens a hit with `git show
   origin/main:<path> | sed -n '<n>p'`, finds it in the merged tree by its text, and classifies it as §III.3 does.
   Then S1 holds, S2 prints nothing, and `git diff HEAD -- '*.golden' '*.sql'` shows only lines main changed.
6. The fast subset, fail count 0 in each: `go -C T vet ./...`; `go -C T test -count=1 ./internal/record/...
   ./internal/cli/... ./internal/report/ ./internal/view/ ./internal/seatprobe/ ./internal/difftest/ ./integration/surface/`;
   the simulator suite; `go -C scripts run ./check -only archaeology,frontmatter,decisions,rulesweep,validatejson,vocabdoc,enginegen,agentgen,golden`.
7. Commit, check `gh pr view <n> --json state`, push, read CI by fail count. **Merge when CI is green and the smoke
   stands** (N13, N15). A main commit that touches no carrier needs only steps 1, 3, 5 and 6.

**What gates a later attempt, and what runs once.** Every attempt: steps 3 and 6 and CI. Once on the first complete
tree, then only on their re-arm surface: §V.2 steps 1–3; §V.3 and §V.4 (`T/internal/{record,report,view}`, or an
epoch re-bump); §V.5 (a record, flag or enum change). **`before` at a moved base** (§V.3): step 0's commit touches
nothing but `archdump`, so `git worktree add --detach ~/.claude/scratch/inquiry-rename/base origin/main`, `git -C
…/base cherry-pick <step 0's commit>`, and dump there; it fails to compile, loudly, if main changed what it calls.

**The smoke, once (N15)**, on the first complete tree beside that attempt's CI. A later attempt re-arms it only by
a hand change to a seat-facing file — a step 2 conflict or a step 5 rename in a file `vocabularySources(t)` reads, a
cobra command or flag declaration, or `record.proto`; a clean merge and a regeneration re-arm nothing. Residue: the
merged tree differs from the smoked one by main's commits and their regeneration.

## IV. Risk & Mitigation

Graded likelihood × impact × cost to mitigate (L/M/H).

| # | Risk | L × I × cost | Mitigation, and the step or gate that carries it |
|---|---|---|---|
| R1 | A mechanical rename damages what it cannot read: a view lands on a table's name, a phrase is split across literals or lines, an article is left wrong, a check passes for the wrong reason | H × H × M | Step 2's `sqlite_master` check; step 2 renames whole identifiers and touches no literal; §III.2's joined-pair finder; §III.3 pass 4 reads rendered text; S6 deletes each translation; `git grep -nIE '\ba inquir'` → 0 at step 9. No step uses a repository-wide text replace |
| R2 | Main moves under the branch: adds a sentence in the old words, or bumps the epoch | H × M × M | §III.9: generated files are never resolved by hand; step 3 fails an attempt whose epoch is not main's plus one; step 5 reads what main added on every attempt and S1 is re-run; `gh pr view --json state` before every push |
| R3 | After the merge, new prose names a fate or a surface that does not exist | M × L × L | An identifier fails to compile; an `--as` value no enum holds fails `TestEveryEnumValueNamedInAPromptIsAccepted`; a verb or view a prompt names fails `TestEveryVerbNamedInAPromptExists` or `TestEveryViewNamedInAPromptExists`. A fate word in running prose, a value legal for another verb's `--as`, and an old word after `--as` in a help page are held by review alone (§III.6, residue) |
| R4 | The epoch bump refuses every live run, scratch record and universe below `E` | H × M × L | That is the rule working. The PR names `feov-record migrate --from <old> --to <new>`; a running session keeps the binary it loaded. §V.4 proves the refusal's wording and the migration on a real record at `E0` |
| R5 | `inquiry record` reads as a noun phrase beside "the record" and `feov-record`; `inquest` is read for `inquiry` on the chair's root help | M × L × L | The group's Short and the page's menu line open with the act ("say what became of an inquiry you proposed"); the Shorts are read side by side in step 4; the seat-probe board (`seatprobe/boards.go:367`) and the smoke (§V.6), which runs before merge, show whether a seat finds the verb |
| R6 | A translation is wrong and silent: `question` filled from the wrong column, or M15 dropping a real reason | M × H × L | M1/M3's test compares `question` with the archived `line` byte for byte; M15's asserts every non-proposal reason is byte-equal; S4 diffs every assembled document of 20 archives |
| R7 | A deleted passage carried the only statement of a live instruction, or a deleted docs row is one `decisions` reads; a moved mechanic lands on no page | M × M × L | Each deletion is read for an instruction first, and the instruction restated in the present tense (§III.2); gate `decisions` at step 9; the PR lists every deleted passage and every moved sentence; step 6 renders each destination page |
| R8 | A seat prompt passes its ceiling; a lens prompt no ceiling measures grows unseen | L × M × L | §III.7; gate `TestNoSeatPromptGrowsPastItsCeiling`; the fourteen simulator prompt goldens show every byte in the diff |
| R9 | `archaeology` fails a renamed line for wording it already carried; `deadcode` fails a helper the rename orphans; the `manual` golden hoists a block repeated on two pages into SHARED and renumbers | M × L × M | The first two are loud and in `check`; an archaeology hit is fixed in the sentence, never allow-listed. Step 7's predicted-against-regenerated diff isolates a renumbering |
| R10 | `propose` leaving the prose channel loses a protection the channel gave (the quoting rule, the heredoc trim) | L × M × L | `flags.Text` attaches the same rule (`flags/register.go:23–32`); `TestEveryFreeTextVerbShowsTheQuotingRule` holds it; the trim (`Prose.Read`) is applied to `--question` and pinned by a test that passes a value ending in a newline |
| R11 | Another plugin or an outside reader depends on a FEOV record word | L × M × L | None does: §III.2's not-carriers command → 0, re-run in §V.1 |

## V. Verification Plan

Before any FEOV command: `unset $(env | grep -o '^FEOV_[A-Z_]*')`; `GOTOOLCHAIN=local`. Logs: `~/.claude/scratch/inquiry-rename/logs/`.

### V.1 Generators and censuses · re-armed by any edit to a source in §II's table, and by every merge of main

1. Each generator of §II re-run, then its check: `protogen` and `vocabdoc` leave `git diff --exit-code` clean;
   `schemagen -check`; `go -C T run ./cmd/feov-record --schema` → `E` (§III.9 step 3); `agentgen -check`, gate `enginegen`.
2. S1, S2 and R11's command, as §I and §IV state them.

### V.2 Suites · steps 1 to 3 on the first complete tree and on their re-arm surface (§III.9); step 4 and CI on every attempt

1. `go -C T test -count=1 -timeout 90m ./... > …/tools.log 2>&1` → `grep -c '^FAIL'` 0, and `grep -c '^ok'` equal to
   the same command's count in a detached worktree at `git merge-base HEAD origin/main`; the branch adds no package.
2. `go -C scripts test -count=1 ./...` → 0 FAIL. `node --test plugins/frank-exchange-of-views/tests/simulator/*.test.mjs` → `fail 0`.
3. **S6, by hand** · re-armed by any change under `T/internal/record/migrate`. Delete, one at a time, each of
   M1–M15's registry rows, the `question` rename, the two `InquiryStatus` rows and the `inquiryArm` field rename;
   run `go -C T test -count=1 ./internal/record/migrate/`. Each fails a named test (the PR lists the pairs); one
   nothing fails is a missing assertion, fixed before merge. Likewise restore `body.Reason` in `propose` (N10).
4. `go -C scripts run ./check > …/check.log 2>&1` → the summary line's failed count is 0. `qlty` and `actionlint`
   run only in CI: read the PR's checks by fail count.

### V.3 Real data: every archived run, before and after (S3, S4) · re-armed by any change under `T/internal/{record,report,view}`

1. `go -C T run ./devcmd/archdump -archives run-archive -out ~/.claude/scratch/inquiry-rename/before` at the base
   (§III.9 at a moved base); then on the branch with `-out …/after` → 20 directories, `cat …/after/*/REFUSALS | wc -l`
   → 0, no `…/after/*/ASSEMBLE_ERROR`.
2. Apply the word map3. Apply the word map to a COPY of `before` (`wordmap.py`, kept in scratch and pasted in the PR) — in `report.md`
   the `**Fate:**` word; in `inquiries.md` its file name, nav label, title, the two fate headings and the `path:`
   words; in `README.md` and the link bars the file name and label; in `run.md` and the dashboard the two metric
   names and the status keys; in `debate.md` and `judgments.md` the verb and subject words where an act is named —
   then `diff -r …/before-mapped …/after` → **no output**; a line that differs is a defect. No archive rests at
   `proposed` (`grep -c '^## proposed' …/before/*/avenues.md` → 0 each), so N10 changes no document.
3. Read by hand, in `after`: `2026-10-11_is-91-prime-m20/report.md`'s three inquiry sections and `inquiries.md`
   whole; `2026-09-11_is-91-prime-b9/inquiries.md` (`## underway (7)`).

### V.4 Real data: a record at `E0` is refused, then migrates (S5) · re-armed by any change to the epoch or to `migrate`

1. With `feov-record-E0` (step 0, rebuilt on a re-bump), make a real record at `E0`: extract
   `run-archive/2026-10-12_is-91-prime-wave-b.tar.gz` to `~/.claude/scratch/inquiry-rename/archived/`, then
   `feov-record-E0 migrate --from …/archived/<run> --to …/at-e0/wave-b`.
2. Branch binary: `feov-record show --run …/at-e0/wave-b --seat-id red-chair inquiries` → **refused**, naming `E0`,
   `E` and `migrate`; `TestEveryArchivedRunBehindThisBinaryIsRefusedUnmigrated` (migrate) passes over all 20.
3. `feov-record migrate --from …/at-e0/wave-b --to …/at-e/wave-b` → exit 0, 0 refusals in `inputs/migration.json`.
4. `feov-record show --run …/at-e/wave-b --seat-id red-chair inquiries --json | jq '[.inquiries[].status] | group_by(.) | map({(.[0]): length}) | add'`
   → `{"confirmed": 4}`; `… | jq '.inquiries[0] | keys'` holds `question` and not `line`.

### V.5 The release fuzz · re-armed by any record, flag or enum change (this is one), and by an epoch re-bump

1. `FEOV_RELEASE_GATE=1 go -C T test -count=1 -v -timeout 40m ./releasegate/fuzz/ > …/fuzz.log 2>&1` →
   `grep -o 'fuzzed [0-9]* debate runs · [0-9]* failed' …/fuzz.log` prints `fuzzed 40 debate runs · 0 failed`;
   `^--- FAIL` and `^--- SKIP` (the variable unset) count 0; the coverage table shows `inquiry record` driving
   `underway`, `confirmed`, `refuted`, `deferred`, `declined`, `abandoned`.
2. `FEOV_RELEASE_GATE=1 FUZZ_N=6 FUZZ_C=6 go -C T test -count=1 -race -v -timeout 30m -run TestFuzzDebate ./releasegate/fuzz/`
   → `grep -c 'DATA RACE'` → 0. A red in either is run on `origin/main` before the branch is blamed.

### V.6 The engine runs · BEFORE merge, on the first complete tree (N15) · re-armed only by a hand change to a seat-facing file (§III.9)

1. `scripts/universe.sh build` from the finished branch; prove it carries the change: `grep -rl 'inquiry record'
   <the installed plugin copy>` prints the chair's and blue's agent definitions, and S1's commands over that copy
   print only `tools/internal/record/migrate/`, `tools/research/`, the OCR corpus and the English list.
2. Before the run: `CLAUDE_CONFIG_DIR=<universe>/config claude auth status` shows `"authMethod": "claude.ai"` (a
   config that lost the subscription bills API credits), and every `feov-record` call from outside the run names
   the run's own binary by absolute path (a bare one reaches the host cache at the earlier epoch and is refused).
3. `SMOKE=1 scripts/universe.sh run "<topic>"`, detached, its stream to a log: one lane, `haiku` on both tiers,
   `--mint-budget 3 --k-max 2`; `scripts/universe.sh watch` reads the board.
4. It reaches VERIFIED or CEILING before the PR merges. Read: the frontier seat's first `inquiry propose` carries
   `--question` and is accepted first time; at least one `inquiry record --as confirmed|refuted`; every refusal on
   the `inquiry` group (`feov-record dashboard`, `ops log`) with the help page or prompt sentence that taught what
   the seat typed; the chair's `inquiry review`; `inquiries.md` in the report set; outcome and wall clock against
   the last smoke. A refusal that traces to the rename is fixed; the fix, a hand change, re-arms the smoke.

### V.7 The auditor gate

`/prosthetic-conscience:plan-audit plans/avenue-becomes-inquiry.md` — approved on the second consecutive PASS, or by
gblock when every remaining gap is `local`. Its NOTES travel into the PR with S1's English list and excluded set, the
earlier-name, `line` and `move` tables, the deleted passages and moved sentences (R7), the §III.7 table after, the S6
pairs, `final.py` and its output, the attempt and smoke figures (N17), the smoke's outcome, and §III.2 re-run.
