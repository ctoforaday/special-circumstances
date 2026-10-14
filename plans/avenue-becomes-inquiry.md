# An avenue becomes an inquiry: the noun, its question, its fates and the act that records them

> STATUS: **PROPOSED 2026-10-13; every fork ruled (§III.1, gblock, 2026-10-13/14).** Awaits the plan-auditor gate.

`F` is `plugins/frank-exchange-of-views`, `T` is `F/tools`, `K` is `~/.claude/scratch/inquiry-rename/kit`; `L/F` is
lines / files. Each census runs from the repository root after `. $K/kit.sh`, whose helpers are named where used
(`oldwords/main.go` is a Go tokenizer printing `path:line:kind:text`, kind `name`, `ident`, `comment`, `literal` or
`text`); the PR pastes the kit. Line numbers and counts are at `815082ad` (`origin/main`, 2026-10-14, event-schema
epoch 25); the implementer re-runs each on the PR's base and the PR carries the figures. **`E0` is `origin/main`'s
event-schema epoch at the attempt and `E` is `E0 + 1`** (§III.9 step 3). Archive figures come from `python3 -I
~/.claude/scratch/inquiry-rename/plan/final.py <dir>` over the extracted `run-archive/*.tar.gz` (20 today).

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
| S1 | No old word or earlier name in the live tree: in any case, inside any name, across a line break, in a path name | the kit, under this table | noun 2,216 lines / 228 files; fates and earlier names 966 / 314; path names 11; split pairs 2; 27 files `git grep` cannot read | noun, path names: no output; commands 2 and 4: the English list; command 5: the same 27 names |
| S2 | Inside `migrate`, an old word is a literal and every such literal a key | under this table | 45 tokens; 41 literals | no output; the literal census prints keys and English only |
| S3 | Every archive migrates | §V.3 step 1 | `PASS: 2`; 20 subtests; 0 refusals at 25 | the same at `E`, one subtest per tarball |
| S4 | Assembled archive documents differ only in the renamed words | §V.3 | — | the mapped diff prints nothing |
| S5 | A record at `E0` is refused, then migrates | §V.4 | — | the refusal names `E0`, `E` and `migrate`; 0 refusals after |
| S6 | No translation is unexercised | §V.2 step 3 | — | every deleted registry row fails a named test |
| S7 | Gates | `go -C scripts run ./check`; the release fuzz (§V.5) | — | fail count 0; `fuzzed 40 debate runs · 0 failed` |
| S8 | The engine runs | §V.6, before merge | — | a VERIFIED or CEILING outcome whose report set holds `inquiries.md` |

S1 — every tracked file outside §III.2's not-the-live-tree list; both filters test the PATH field only. `OLD` is a
substring (`motion_direction`, `MOTION_SUBJECT_DIRECTION`, `xDirection`); `any` matches at a name boundary in every
case style (`GetLine`, `avenue_line`, `MAX_LINES`). The kit (whose `s1` runs the five) uses the system `grep` — the
agent shell's is a function that can refuse a pattern and print nothing — sets `pipefail`, and warns on an
uncommitted tree, since `git grep` reads the working tree:

```
export NOT="^(plans|law|research|run-archive)/|^feov-memory/[a-z]+-scorecard[.]md\$|^$T/(research|internal/record/migrate|internal/tessocr/testdata/corpus)/"
export GEN="^$F/agents/|[.]golden\$|[.]pb[.]go\$|/testdata/(schema[.]sql|payload-keys[.]txt)\$|^$F/docs/vocabulary[.]md\$|^$F/skills/(research-protocol|adversarial-audit)/SKILL[.]md\$"
OLD='avenue|conclud|pursu|direction|lines?[-_ ]?of[-_ ]?inquir'
any() { echo "(^|[^A-Za-z0-9])$1([^a-z0-9]|\$)|(^|[^A-Z0-9])$2([^a-z0-9]|\$)|(^|[^A-Z0-9])$3([^A-Z0-9]|\$)"; }
LINE=$(any 'lines?' 'Lines?' 'LINES?'); MOVE=$(any 'mov(e|ed|es|ing)' 'Mov(e|ed|es|ing)' 'MOV(E|ED|ES|ING)')
live() { awk -F: '$1 !~ ENVIRON["NOT"]'; }; authored() { awk -F: '$1 !~ ENVIRON["GEN"]'; }
git grep -lIi avenue | live                                                       # 1 the noun: no output
git grep -IiE 'conclud|pursu|direction|lines?[-_ ]?of[-_ ]?inquir' | live | sort  # 2 fates and earlier names
git ls-files | live | command grep -iE "$OLD"                                     # 3 path names: no output
git grep -lI '' | live | xargs awk -v old="$OLD" -f $K/pairs.awk | sort           # 4 split pairs, narrations
comm -23 <(git ls-files | live | sort) <(git grep -lI '' | sort)                  # 5 what git grep cannot read
```

Command 2: 966 lines (363 fate words, 616 earlier names, 35 inside a longer name, `planguard/growth.go:61` among
them); 230 generated, 404 in tests and test data, 332 authored. The earlier names are those the registry entry for
this concept lists (`T/internal/terms/terms.json:924–957`); `direction` is the concept at `T/internal/cli/root.go:76`
(rendered into two constitutions), `record.proto:257`, `cli/motion/verbs.go:300`, `report/motions.go:32`, and
English or a remand's direction elsewhere. Command 3: the 11 files §III.4's rule renames. Command 4: 2 split pairs
(`record.proto:960`, "line / of inquiry", and its `record.pb.go` copy) and N7's narrations. Command 5: two empty
files, the OCR model, a PDF, ten PNGs drawn by `tessocr/testdata/{main,pages}.go`, and 13 gzipped third-party pages
the kit's `unread_gz` searches (`direction` in CSS and titles). **Every line of commands 2 and 4 is read (N14,
§III.3)**: the concept is renamed, English stays. After the rename they print the English list, pasted in the PR;
S1 holds when `diff` against it, line numbers cut (`cut -d: -f1,3-` on command 4), prints nothing. A line main adds
that §III.9 step 5 reads as English joins the list.

S2 — inside `migrate` an old word is a translation key, or a test's archived input or planted row, and nothing else:
never an identifier, a file name or a comment. The tokenizer reads Go by token, so a raw string spanning lines is one
literal and a comment is never taken for one; any other file is read by line:

```
oldwords "$OLD" $(git ls-files $T/internal/record/migrate) | command grep -v ':literal:'           # no output
oldwords "$OLD|\blines?\b" $(git ls-files $T/internal/record/migrate) | command grep ':literal:'   # the literal census
```

Today the first prints 45 (25 comments, `eraentries.go:207`'s `motion_direction` among them; 18 identifiers; 2 file
names). A literal is not thereby a key: the literal census (41 today) is classified in §III.5. S1 and S2 run by hand,
at landing and on every attempt; neither is a gate.

### The value case

- **What it buys.** The five rows above stop being taught. Two have a measured cost: a gate held on universe-m22 by
  two readings of one fate, and one flag meaning two texts in one group. The report already prints "Question /
  Hypothesis / Fate"; the record, the verbs, the flag, the help pages and the prompts then say the same words.
- **What it costs.** The noun is on 2,216 lines in 228 files: 641 regenerate and 1,575 are edited, 949 of them in
  tests; one review of a diff near 4,000 lines each way. Reading no tool shortens: 966 fate-word and earlier-name
  lines, 1,123 `line` and 630 `move` lines (§III.3), five help pages read to closure (N19). The narrations N7 deletes
  keep their reasoning in the commit history only. One epoch: every live run and scratch record is refused until
  migrated. The landing: one or more merge attempts and one 45-minute smoke (§III.9). No test and no generator is
  added for the fate words (N9).

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
  (`integration/surface/promptverbs_test.go:926`) fails a `--<flag> <value>` no enum bound to that flag holds.
  `TestVocabularyProse` (`integration/surface/vocabulary_prose_test.go:47`) runs the terms registry over
  `vocabularySources(t)` (the GENERATED agents, not their sources), failing a banned variant and an idle mask or allow.
  `TestNoSeatPromptGrowsPastItsCeiling` counts bytes for four seats. `archaeology` reads ADDED lines of agent-facing
  files, so a renamed line re-arms it on wording it already held; `rulesweep` needs its two commit trailers;
  `decisions` reads `F/docs/seat-command-triggers.md`.
- **CI and the merge.** The `hooks` workflow on a pull request takes about 5.5 minutes; its longest job runs
  `go test ./...` whole on `T`; Windows runs `-short`. The release fuzz is not in it. `main` has no branch protection
  and no ruleset (`gh api repos/ctoforaday/special-circumstances/branches/main/protection` → 404; `…/rulesets` → `[]`):
  a PR whose CI ran against an older `main` still merges.
- **Constraints.** CLAUDE.md, Change discipline: the live code accepts and prints only the new form; `migrate` is
  the only upgrader. Versions do not move. No network, secret or privilege; no step uses `gopls` (not installed).

## III. Proposed Changes

### III.1 Rulings (gblock; N1–N16 2026-10-13; N9, N15, N16 widened, N17–N20 2026-10-14)

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
| N16 | The captured chair relays, `git ls-files "$F/tests/simulator/testdata/*.json"` (m16, m3, o3), are TEST INPUTS in the live tree and not records: their text is rewritten in the new words like everything else (o3 holds only a remand's `direction`) |
| N17 | The plan carries what the implementer needs once; measured figures only the PR can settle (main's rate of change, the attempt window, attempt and smoke counts, prompt sizes after, the deleted passages) go to the PR |
| N18 | NO ARCHAEOLOGY; THE HELP PAGE TEACHES THE VERB. Each renamed verb's `--help` teaches it completely on its own: what each flag holds, every fate's meaning, what is refused and what to do next. Prompts, constitutions and skills keep the DUTY — what a seat must do and why — and carry no command mechanics: mechanics the sweep finds there move to the help page. No renamed text says what anything was called or narrates an incident (§III.6, §III.8 steps 4 and 6) |
| N19 | A FRESH READER CHECKS EACH HELP PAGE. For `inquiry propose`, `inquiry record`, `inquiry review`, `motion inquiry rule` and `show inquiries`, a fresh subagent given ONLY the page rendered from the built binary answers what the verb does, what each flag takes, what is refused and why, and what to do next. Each open question it records is closed by a change to the page; the reading repeats until it records none. The PR carries each page's final questions and answers (§III.8 step 4) |
| N20 | EVERY CHECK SEES ALL IT CERTIFIES, AND NOTHING RELIED ON IS STALE. Each census, sweep, test or reading sees every input it certifies — any case, inside any name, across a line break, path names, generated, binary and test-data files, every directory of the live tree, and what main adds; a value main can move (the epoch, the base, a count, CI, a generated file, a before-image, the smoked tree) is re-read where it is relied on (§III.9) |

**The naming census the rulings rest on.** `git grep -nIi inquir -- $T $F/skills $F/agents scripts/agentgen/src` → 35
lines in 11 files: `migrate` keys, one narrated incident (N7), two registry bans (N8).
`git grep -nIE '"question"|--question|flags\.Question|\bQuestion\b' -- $T $F/skills $F/agents scripts/agentgen/src ':!*.pb.go' ":!$T/research"`
→ 24 lines in 7 files, no flag, key or identifier: the title's `**Question:**` (`report/layout.go:110`,
`references/report_template.md:52`), the RUN's question (N5), as in the `(means)` at `record.proto:266,320,323,485`;
and each approach's `- **Question:**` row (`report/assemble.go:670`). Both senses already share the word; §III.6
keeps them apart. `git grep -ci underway -- $F scripts` → 0; `confirmed`: English only; `git grep -n 'Use: *"record' -- $T` → 0;
`git grep -nIi inquest -- "$T/*.go"` → a bench and chair verb group, 67 lines in 18 files (R5).

### III.2 Carrier census (the consumer census for every contract below)

Per row: `git grep -niI avenue -- <path>`, lines and `-l` files. **Complete when S1 commands 1 and 3 print nothing.**

| Carrier | Path | L/F | How it changes |
|---|---|---|---|
| Schema, generated code | `T/internal/record/recordpb/record.proto`; `record.pb.go` | 50/1; 205/1 | by hand from §III.4; `protogen` |
| Stored views, schema goldens | `recordsql/views.go`; `recordsql/testdata/schema.sql`, `recordpb/testdata/payload-keys.txt` | 20/1; 46/2 | by hand (2 views); regenerate, read |
| Migrate | `T/internal/record/migrate/*.go` | 28/6 non-test | by hand, §III.5; old words stay only as keys (S2) |
| Record folds | `T/internal/record/*.go` non-test | 210/22 | identifier rename, then literals and comments |
| CLI and help pages | `T/internal/cli/**` non-test `.go`; `cli/seat/help/{propose,move,review,dispatch,mint}.md` | 107/17; 13/5 | by hand; `move.md` → `record.md`; each page teaches its verb (N18, N19) |
| Report, view, scorecard, consistency, seatprobe, flags | `T/internal/{report,view,scorecard,consistency,seatprobe,flags}` non-test | 145/15 | rename, then literals and comments |
| Other Go | `anchor`, `graph`, `diagnostics`, `refusalcheck`, `reportvoice`, `fetchcache` | 1–2 each | by hand |
| Go tests; of them the release fuzz | `*_test.go`; `T/releasegate/fuzz/` | 918/92; 167/3 | compiler, then literals and comments; the fuzz by hand (§III.8 step 8) |
| Engine prompts | `F/skills/research-protocol/scripts/debate.js` | 20/1 | by hand; the blocker kind at `:409` by `enginegen` |
| Constitutions and skills | `scripts/agentgen/src/**` → `F/agents/*.md`, `F/skills/**` | 15/7 → 212/11; 17/3 | sources and references by hand; `agentgen` |
| Goldens, simulator | `T/internal/difftest/testdata/`; `T/internal/dashboard/testdata/render-{live,terminal}.golden`; `F/tests/simulator/` | 125/22; 2/2; 70/12 | regenerate (§III.8 step 7); `debate.test.mjs` and the relay fixtures (N16) by hand |
| Terms | `T/internal/terms/terms.json`; `F/docs/vocabulary.md` | 7/1; 8/1 | by hand (§III.6); `vocabdoc` |
| Docs, class registry | `F/README.md`; `F/docs/{seat-command-triggers,seat-surface-naming,duty-docket-preregistration,why-a-seat-stops}.md`; `feov-memory/protocol-class-registry.md:110` | 1; 17/4; 1 | by hand |
| Names | 13 files (S1 command 3, and S2's 2 in `migrate`); 30 test functions (`git grep -nIE 'func Test[A-Za-z0-9_]*Avenue'`) | — | `git mv`; by hand |

**Not carriers**: `git grep -ciIE 'avenue|inquir' -- plugins ':!plugins/frank-exchange-of-views'` → 0; the same over
`$F/hooks $T/internal/capture $T/cmd $T/devcmd` → 0. **Not the live tree** — the whole list, S1's exclusions:

| Path | Why |
|---|---|
| `T/internal/record/migrate/` — the string keys a translation maps from, and the tests that read archived runs | that is what `migrate` is; S2 holds it to keys |
| `run-archive/` | the gzipped record of each captured run |
| `plans/` (this file included) | dated design records: `plans/README.md` keeps a plan as the record of what changed |
| `law/` | holdings and proposed classes, as ruled. Residue: `precedents.md:15,21,69` hold the noun, and `setup.MirrorLaw` copies `law/` into each run's `inputs/law`, which the judge reads (`debate.js:517`) |
| `research/`, `T/research/` | the research corpus: reports and records as written |
| `feov-memory/*-scorecard.md` | scorecard history, appended run over run at capture |
| `T/internal/tessocr/testdata/corpus/` | a third party's page as OCR read it (`FIFTH AVENUE BANK`, one line): not this concept, not this repository's text |

**Passages deleted under N7.** A narration's old word and its marker are often on adjacent lines, and an old phrase
can break across lines or literals, so S1's command 4 (`$K/pairs.awk`) reads each line JOINED to the one before it —
comment leaders, a string-literal join and a line-end hyphen stripped — and prints a **split** pair (an old word only
the joined text holds) and a **narration** (an old word or `inquir` beside a dated or measured marker: `MEASURED`,
a date, `universe-mN`, `#NNN`, "N of M", `verbatim`, `retired`, `CUT`, `used to`, `became`; generated files skipped).
Over the whole live tree: 256 narration pairs in 96 files — `.github/workflows/hooks.yml`, `F/docs`,
`scripts/agentgen/src`, two other plugins' Go and `record.proto:960` among them. Each is a sentence in the present
voice, renamed, or a narration, deleted whole and never renamed in place; an instruction a deleted passage carried is
restated in the present tense where it stood (R7).

### III.3 The words grep cannot settle: the earlier names, `line` and `move`

A fate word or earlier name is the concept or English; a `line` is mostly a text line; a `move` mostly gaps and
grades. Each sweep's file set is closed by a command; every classification table goes in the PR.

1. **Fates and earlier names, BEFORE their ban is deleted.** S1 commands 2 and 4 are classified at step 0, compounds
   like the rest: **concept** (→ `inquiry`, or the new fate) or **English** (stays); step 5 deletes the registry
   entry only after this table exists.
2. **`line` and `move`, in every authored file that names the concept, the relay fixtures included:**
   `for w in "$LINE" "$MOVE"; do git grep -lIi avenue | live | authored | xargs git grep -nIE "$w" --; done`
   → `line` 1,123 lines in 135 files (562 in 61 outside tests and test data), `move` 630 in 115 (294 in 60). A `line`
   is **inquiry** ("a declined line"), **question** ("the line you propose") or **unrelated**; a `move` is **record**
   or unrelated. A file that says `move` of this concept and never names it is reached by passes 3 and 4.
3. **The compiler lists the readers.** Renaming `Avenue.Line` and `GetLine()` breaks every reading site; at each,
   the comment above and the literal beside it are read.
4. **The rendered text is read, not its source.** After §III.8 step 7, every pattern above is searched in what a seat
   reads as printed — `$F/agents`, the two generated `SKILL.md`, `$F/tests/simulator/testdata` (the prompt goldens
   and the relays), `$T/internal/difftest/testdata` (every help page and refusal as printed) and
   `$T/internal/dashboard/testdata` (`line` 260, `move` 286 today) — and each hit classified again: a sentence
   assembled from two constants shows its seam only here.

### III.4 The name table

Numbers, positions and every unlisted name are unchanged. **The rule:** a name holding `avenue` in any case takes
`inquiry` in the same case and position (plural `inquiries`), and a name holding `pursued` or `concluded` takes
`underway` or `confirmed` the same way. The Go names come from the tokenizer, so a lower-case, snake or compound name
is in the list — `oldwords 'avenue|conclud|pursu' $(git ls-files "$T/*.go" ':!*.pb.go') | command grep ':ident:' | cut -d: -f4- | sort -u`
→ 126 with the noun (64 outside tests) and 4 with a fate only (the two enum constants, `view_test.go`'s `pursuedAt`,
`TestAPursuedAvenueReaffirmedThisRoundIsNotStale`); the 13 files; the 30 test functions; and every row below that
states no exception.

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
`git grep -nE 'avenue move|avenue propose|avenue review|motion avenue|show avenues' | live` — 59 authored, 88 in
tests and test data, 59 generated (S1's path split) — and every one changes.

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
`record.go:1307`'s refusal names `--question` through `flags.ForPayloadKey`. The sites passing `--reason` to
`propose` — the kit's `reasonsites avenue`, which reads each file whole so a call broken across two lines is seen
(`avenue_test.go:32`, `avenuevoice_test.go:23`, `seatprobe/build.go:231`) — are 36, and follow. A site it misses
fails loudly: `propose` no longer registers `--reason`, so cobra refuses it, and a prompt spells no flag
(`TestNoRenderedPromptSpellsAFlag`).

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
does not read; ids (§II). **Residue.** `migrate` re-reads no prose: of the 25 archived inquiries that end
`concluded`, 3 state a refutation in their reason and arrive `confirmed`. Accepted.

**Migrate's literals hold old AND current words.** Each of S2's literal census (41) is a **key** — a word an archive
holds, translated from (`"line_of_inquiry"`, `valueRenames`' left side, a test's archived SQL): it stays; a
**current** word — a translation target, an id or anchor kind, a table, arm or field of the current schema
(`mapped("avenue", …)`, `valueRenames`' right side, `remap.go:51` and its field lists at `:76–109`): it is renamed;
or English (`line` as a text line). After the rename the census prints keys and English only (in the PR). A missed
current word fails an archive test (identity refuses what the current schema does not declare), `TestValueRenames`
(a target the enum lacks) or `anchor.ID` (a panic) — except in `anchor.IDPattern`, which drops a name it does not
know (`anchor.go:113–123`): there only this census and §V.3's diff see it. Making `IDPattern` refuse an unknown name
changes `anchor` for every kind, beyond this concept: the implementer's call.

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
| all | `TestEveryArchivedRunRenders`, `TestEveryArchivedRunMigrates` | S3; the three `migrate/testdata/archived_*.golden` files are expected unchanged — a changed line is read, not regenerated past |

### III.6 The terms registry, the fate wording, and duty against mechanics

**[MODIFY] `terms.json`, then `vocabdoc`.** The entry at `:924` becomes **inquiry**: "An inquiry is one question the
research could take up — proposed by blue with the hypothesis that would be true if it paid off, recorded with what
became of it, ruled on by the chair, and reviewed against the report once per epoch." Its bans and collisions
(`:933–957`) are [DELETE]d after §III.3 pass 1, and nothing is added (`TestVocabularyProse` requires no entry). `:838`
reads "an inquiry's state keeps the word"; `retiredSurfaces` gains no row; "BOTH AVENUES"
(`scripts/agentgen/src/blue-researcher.md:52,57`) becomes "BOTH ROUTES".

**Fate wording, by the least machinery (N9).** The census of authored non-test lines holding three or more of the
seven status words (`W` the seven joined by `|`; `awk` counts the seven in each lowered line) —
`git grep -lIiE "$W" | live | authored | awk '$0 !~ /(_test[.]go|[.]mjs)$|\/testdata\//' | xargs git grep -nIiE "$W" --`
— prints 23 lines in 13 files, in three kinds. Nothing new is generated and no test is added.

1. **The sections constant stays a hand copy**: `debate.js:196` (`INQUIRY_SECTIONS`) restates
   `report.InquirySections`, and `TestBluesPromptsNameTheSectionsTheToolComposes` (`report/assemble_test.go:408`)
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

**Two questions.** The run has a question and so does each inquiry. At step 6, every line on the seat surface
(§III.9) or in the rendered agents and skills that says `question` within three lines of one naming an inquiry —
the kit's `questions inquir` (`git grep -z -C3` over those files; 103 today as `questions avenue`) — is read, and
where both could be meant the sentence names whose it is. N19's reader is asked whose question `--question` holds.

### III.7 The byte budget for prompts

At `54e0ff06` (`go -C T test -count=1 -run TestNoSeatPromptGrowsPastItsCeiling -v ./internal/seatprobe/`): red-chair
8,489 / 10,200, red-lens-evidence 10,570 / 12,300, blue-respond 9,456 / 15,200, judge 5,889 / 7,900. The word
changes add a few bytes each and N18's moved mechanics remove more; no ceiling is raised; the PR pastes the table after.

### III.8 Order of work, and the check that closes each step

By kind, never a repository-wide text replace (R1); each step begins once the one before passes its check; one commit
per step. The PR merges by merge commit, so step 1 may close on a tree that does not compile, and says so. A check
closes on a COUNT read from its output (the kit's `tally`), never on `ok` or an exit code; commands holding a pipe
are in the block after the table, by step number.

| # | Step | The check that closes it |
|---|---|---|
| 0 | Baseline at the base: commit `T/devcmd/archdump` (N12), touching no other path; dump every archive's assembled set to `~/.claude/scratch/inquiry-rename/before/`; build `…/bin/feov-record-E0`; render the help pages under their base names (block, 4); classify fates and earlier names (§III.3 pass 1) | one directory per tarball, 0 `REFUSALS` lines, 0 `ASSEMBLE_ERROR`; the classification covers every line S1 commands 2 and 4 print |
| 1 | Schema: `record.proto` from §III.4; `protogen`; `requirements.json` to `E` and `schemagen`; the two views by hand | `go -C T build ./internal/record/recordpb/` passes and nothing that imports it does, which is step 2's worklist |
| 2 | Go identifiers, led by the compiler. First the names under the rule, generated and hand-written alike: ONE `gofmt -l -w -r '<Old> -> <New>'` pass PER DISTINCT NAME over `git ls-files "$T/*.go" ':!*.pb.go'`, the names from §III.4's tokenizer census; a pass rewrites whole identifiers, declarations and uses, and no comment or literal. Then the names outside the rule (`.Line`, `GetLine`, `Moved`, `avenueMoves`, a parameter), each declaration by hand, the readers from `go vet` (block, 2) until it prints nothing | `go -C T build ./... && go -C T vet ./...`; `UPDATE_GOLDENS=1 go -C T test -count=1 ./internal/record/recordsql/ ./internal/record/recordpb/`, then every changed line of `schema.sql` read: `SELECT name FROM sqlite_master GROUP BY name HAVING count(*) > 1` is empty, and `inquiry_state` reads FROM `"inquiry"` |
| 3 | Migrate: §III.5, tests first; the literal census classified | §III.5's tests by name (block, 3); S6; S2 prints nothing and its literal census only keys and English. The package whole — whose render tests print words step 4 moves — closes at step 4 |
| 4 | CLI: verbs, flags, JSON keys, refusals, N10, literals in the 17 CLI files and the record folds; the five help pages written so each teaches its verb on its own (N18): what `--question`, `--hypothesis`, `--method`, `--id`, `--as` (every fate, with its meaning) and `--reason` hold, what is refused, and what to do next | `cli` and `record` whole (block, 4), migrate among them; each page rendered from the built binary and read twice: against step 0's page — or the latest base's (§III.9) — for a lost instruction; and by N19's fresh reader, to closure, its final questions and answers saved for the PR |
| 5 | Terms and the fate wording: §III.6's registry entry and items 1–3; then `agentgen`, `vocabdoc` | the two named tests (block, 5), `TestVocabularyProse` with no stale mask; `TestBluesPromptsNameTheSectionsTheToolComposes` holds the sections copy; `agentgen -check` clean |
| 6 | Seat-facing prose: help pages, `(means)`, the prompt constants, constitutions, skill sources, references — sentence by sentence (§III.3), each in a prompt, constitution or skill classified duty or mechanics and mechanics moved (§III.6, N18); the two questions (§III.6); then `enginegen`, `agentgen`, `vocabdoc` | the six named tests (block, 6); `agentgen -check` and gate `enginegen` clean; each moved sentence found on its destination page by rendering it |
| 7 | Goldens: for each, write the PREDICTED file in scratch — the old golden with §III.4's word map applied by a script — regenerate (both commands of §II's table), and `diff` predicted against regenerated; then §III.3 pass 4 | every line the two disagree on is read and explained in the PR; lines that agree are the rename and nothing else |
| 8 | The release fuzz, by hand: the three files, the drive (`fuzz_test.go:1938–1988`, `:5121–5163`), the coverage key per recording verb, the oracle's messages | §V.5; and by mutation: take `refuted`, then `underway`, out of `nextInquiryStatus`'s rotation — the coverage gate turns red each time |
| 9 | Simulator, docs, comments, tests: `debate.test.mjs`, the relay fixtures (N16), the prompt goldens, `F/README.md`, `F/docs/*.md`, the class registry line, §III.2's deletions | the simulator suite (`fail 0`); gates `decisions`, `mjsparity`, `archaeology`; **S1 holds** |
| 10 | §V.1 to §V.5 in full, then the smoke (§V.6) | S2–S8 |

```
go -C $T vet ./... 2>&1 | command grep -E '[.]go:[0-9]+'      # 2: the readers; repeat until it prints nothing
go -C $T test -count=1 -v -run '^(<the names, joined by |>)$' ./internal/record/migrate/ 2>&1 | tally   # 3: PASS = names
go -C $T test -count=1 ./internal/cli/... ./internal/record/... 2>&1 | tally   # 4: ok + ? = `go -C $T list` of the two
pages $K/help/after inquiry inquiries record                                   # 4: six renders (step 0: avenue avenues move)
go -C $T test -count=1 -v -run '^(TestVocabularyProse|TestEveryEnumValueNamedInAPromptIsAccepted)$' ./integration/surface/ 2>&1 | tally  # 5: PASS 2
go -C $T test -count=1 -v -run '^(TestEveryEnumValueNamedInAPromptIsAccepted|TestNoSeatPromptGrowsPastItsCeiling|TestTheSeatPromptsNameNoVerb|TestNoPromptGrowsItsCommandCatalogue|TestDetailsDoNotRestateTheFlagsCobraPrints|TestBluesPromptsNameTheSectionsTheToolComposes)$' ./integration/surface/ ./internal/seatprobe/ ./internal/cli/ ./internal/report/ 2>&1 | tally   # 6: PASS 6
```

Every count is read with no `FAIL`.

### III.9 Landing without a freeze (N13), with the smoke inside it (N15)

The rename lands at epoch 26 first and the slimming workstream's wave C schema PR builds on it at 27 (gblock,
2026-10-14; context, not a ruling of this plan): `E0 + 1` derives that. On merge the implementer messages that
workstream with the merge commit and epoch, naming each open PR whose `gh pr diff` holds a match of `OLD`.

**One attempt**, begun at the first push and again each time main moves:

1. `gh pr view <n> --json state` → `OPEN`; note `git merge-base HEAD origin/main`; `git fetch origin main && git
   merge origin/main` — never a rebase.
2. Conflicts: in an AUTHORED source take main's hunk and re-apply §III.4's names; in a GENERATED file take either
   side, and step 4 overwrites it.
3. **The epoch is main's plus one, on every attempt.** Main may bump the epoch for another shape change; both sides
   then make the same edit, git merges it cleanly, and two shape changes would share one epoch. The kit's `epoch`
   (step 7's two epoch lines, on the working tree) exits non-zero on a mismatch, naming both; then `eventSchema` is
   set to `E0 + 1`, `schemagen` re-run, `feov-record-E0` rebuilt in a base image of the new base, §V.4 and §V.5
   re-armed. §III.5's rows still map each old word straight to the current one.
4. Regenerate, in this order: `protogen`; `schemagen`; `go -C T build ./...` (the compiler names every identifier
   main added under an old name); `enginegen`; `agentgen`; `vocabdoc`; both golden commands of §II's table.
5. Read what main ADDED: `added <the noted merge base> > $K/added` (the kit: each added line as `path:line:` in
   `origin/main`'s copy, from `git diff --unified=0`, a header state keeping an added line that itself begins `+`);
   then `command grep -iE "$OLD" $K/added; command grep -E "$LINE|$MOVE" $K/added`. A hit is opened with `git show
   origin/main:<path> | sed -n '<n>p'`, found in the merged tree by its text and classified as §III.3 does. Then S1
   — all five commands, so a path, an unreadable file or a split pair main added is seen — and S2 hold, and
   `git diff HEAD -- '*.golden' '*.sql'` shows only lines main changed.
6. The fast subset, no `FAIL` in each: `go -C T vet ./...`; `go -C T test -count=1 ./internal/record/...
   ./internal/cli/... ./internal/report/ ./internal/view/ ./internal/seatprobe/ ./internal/difftest/ ./integration/surface/ 2>&1 | tally`;
   the simulator suite; `go -C scripts run ./check -only archaeology,frontmatter,decisions,rulesweep,validatejson,vocabdoc,enginegen,agentgen,golden`.
7. Commit, check `gh pr view <n> --json state`, push, wait for CI. **Merge only while `origin/main` is the commit
   this attempt merged and epoch-checked, CI is green on the pushed head, and the smoke stands** (N13, N15):
   `bash $K/premerge.sh <n>`, run IMMEDIATELY before the merge it ends in; a non-zero exit begins a new attempt (a
   pending check: wait and run it again).
   ```
   . ~/.claude/scratch/inquiry-rename/kit/kit.sh; set -eu; N=$1; H=$(git rev-parse HEAD)
   git fetch origin main; B=$(git rev-parse origin/main)
   git merge-base --is-ancestor "$B" HEAD                  # main has not moved since this attempt merged it
   E0=$(git show "$B:$F/requirements.json" | jq -e .eventSchema); E=$(git show "HEAD:$F/requirements.json" | jq -e .eventSchema)
   [ "$E" -eq $((E0 + 1)) ] || { echo "EPOCH MISMATCH: branch $E, origin/main $E0" >&2; exit 1; }
   [ "$(gh pr view "$N" --json state,headRefOid -q '.state + " " + .headRefOid')" = "OPEN $H" ]
   gh pr checks "$N" --json name,bucket | jq -e 'length > 0 and all(.[]; .bucket == "pass" or .bucket == "skipping")
     and any(.[]; .name == "test (ubuntu-latest)" and .bucket == "pass")' > /dev/null
   read -r SB S < $K/smoked; surface > "$K/surface.$H"; cmp -s <(own "$SB" "$S") <(own "$B" HEAD)  # else: §V.6 again
   gh pr merge "$N" --merge --match-head-commit "$H"
   M=$(gh pr view "$N" --json mergeCommit -q .mergeCommit.oid); git fetch origin main; [ "$(git rev-parse "$M^1")" = "$B" ]
   ```
   `--match-head-commit` refuses a head that moved after the check. A commit landing on `main` in the seconds between
   the fetch and the merge is the one window nothing here closes (§II: no branch protection); the last line makes it
   loud, and S1 and the epoch check then run on `origin/main` at once and a defect is fixed forward.

**Once-only checks.** Every attempt runs steps 3, 5, 6, CI and step 7. A once-only check records the commit it ran on
(`R`, in the PR) and runs again when `git diff --name-only R HEAD -- <surface>` prints a path — from either side,
main's included: §V.2 step 3 (S6) on `$T/internal/record/migrate`; §V.3 on `$(gosrc ./devcmd/archdump)` and
`run-archive/`; §V.4 on `$F/requirements.json` and `$T/internal/record/migrate`; §V.5 on `record.proto`,
`$T/internal/{flags,cli}`, `$T/releasegate/fuzz` and `$F/hooks/hooks.json`; N19's readings when a page rendered
from HEAD differs from the page last read. Later attempts take §V.2 step 1 from CI's `test (ubuntu-latest)`.

**A base image** — `before` (§V.3), the before-pages (step 4), `feov-record-E0`, §V.2's `ok` count — is made at the
commit compared against, in a worktree of its own under the main checkout, removed after: `X=$(git rev-parse --short
origin/main); D=$(dirname "$(git rev-parse --path-format=absolute --git-common-dir)")/.claude/worktrees/base-$X; git
worktree add --detach "$D" origin/main`, step 0's file written in from `git show <step 0's commit>:$T/devcmd/archdump/main.go`
(it fails to compile, loudly, if main changed what it calls), and `git worktree remove --force "$D"` after.

**The smoke, once (N15)**, on the first complete tree beside that attempt's CI; `$K/smoked` holds its base and
commit (`SB S`). The **seat surface** is every hand-edited source whose text reaches a seat — the kit's `surface`:
`gosrc ./cmd/...` (the Go, cgo, build-constrained and embedded files of every shipped binary, by `go list -deps`:
help pages, report layout, view text, refusals, hook messages), `record.proto`, `$F/skills` and `$F/hooks` as
tracked, and the generators' sources and code (`scripts/agentgen` with `src/**`, `$T/devcmd/enginegen`), never their
output: 365 paths at `815082ad`, written as `$K/surface.<sha>` at the smoked commit and at each head. A later attempt
re-arms the smoke only when the branch's OWN change there differs from the one smoked — `own SB S` against
`own B HEAD` (step 7), the `+`/`-` lines over the union of those lists, headers and positions dropped — as a conflict
resolution or a rename of a line main added does, and a clean merge and a regeneration do not. Residue: the merged
tree differs from the smoked one by main's commits and their regeneration.

## IV. Risk & Mitigation

Graded likelihood × impact × cost to mitigate (L/M/H).

| # | Risk | L × I × cost | Mitigation, and the step or gate that carries it |
|---|---|---|---|
| R1 | A mechanical rename damages what it cannot read, or a check passes for what it cannot see: a view lands on a table's name, a phrase is split across literals or lines, an old word sits inside a longer name or a path, an article is left wrong | H × H × M | Step 2's `sqlite_master` check; step 2 renames whole identifiers from a token census and touches no literal; S1's substring and name-boundary patterns, path names, unread files and joined pairs; §III.3 pass 4 reads rendered text; S6 deletes each translation; `git grep -nIE '\ba inquir'` → 0 at step 9. No step uses a repository-wide text replace (N20) |
| R2 | Main moves under the branch: adds a sentence in the old words, bumps the epoch, or lands between a green CI and the merge | H × M × M | §III.9: generated files are never resolved by hand; step 3 fails an attempt whose epoch is not main's plus one; step 5 reads what main added and S1 is re-run whole; step 7's block merges only while main is the commit the attempt merged, with `--match-head-commit`, and makes the remaining seconds-wide window loud |
| R3 | After the merge, new prose names a fate or a surface that does not exist | M × L × L | An identifier fails to compile; an `--as` value no enum holds fails `TestEveryEnumValueNamedInAPromptIsAccepted`; a verb or view a prompt names fails `TestEveryVerbNamedInAPromptExists` or `TestEveryViewNamedInAPromptExists`; the open PRs holding an old word are named on merge (§III.9). A fate word in running prose, a value legal for another verb's `--as`, and an old word after `--as` in a help page are held by review alone (§III.6, residue) |
| R4 | The epoch bump refuses every live run, scratch record and universe below `E` | H × M × L | That is the rule working. The PR names `feov-record migrate --from <old> --to <new>`; a running session keeps the binary it loaded. §V.4 proves the refusal's wording and the migration on a real record at `E0` |
| R5 | `inquiry record` reads as a noun phrase beside "the record" and `feov-record`; `inquest` is read for `inquiry` on the chair's root help; `question` is read as the run's | M × L × L | The group's Short and the page's menu line open with the act ("say what became of an inquiry you proposed"); the Shorts are read side by side in step 4; N19's reader and §III.6's two-questions reading; the seat-probe board (`seatprobe/boards.go:367`) and the smoke (§V.6) show whether a seat finds the verb |
| R6 | A translation is wrong and silent: `question` filled from the wrong column, M15 dropping a real reason, a current word left in a migrate literal | M × H × L | M1/M3's test compares `question` with the archived `line` byte for byte; M15's asserts every non-proposal reason is byte-equal; §III.5's literal census; S4 diffs every assembled document of every archive |
| R7 | A deleted passage carried the only statement of a live instruction, or a deleted docs row is one `decisions` reads; a moved mechanic lands on no page | M × M × L | Each deletion is read for an instruction first, and the instruction restated in the present tense (§III.2); gate `decisions` at step 9; the PR lists every deleted passage and every moved sentence; step 6 renders each destination page |
| R8 | A seat prompt passes its ceiling; a lens prompt no ceiling measures grows unseen | L × M × L | §III.7; gate `TestNoSeatPromptGrowsPastItsCeiling`; the fourteen simulator prompt goldens show every byte in the diff |
| R9 | `archaeology` fails a renamed line for wording it already carried; `deadcode` fails a helper the rename orphans; the `manual` golden hoists a block repeated on two pages into SHARED and renumbers | M × L × M | The first two are loud and in `check`; an archaeology hit is fixed in the sentence, never allow-listed. Step 7's predicted-against-regenerated diff isolates a renumbering |
| R10 | `propose` leaving the prose channel loses a protection the channel gave (the quoting rule, the heredoc trim) | L × M × L | `flags.Text` attaches the same rule (`flags/register.go:23–32`); `TestEveryFreeTextVerbShowsTheQuotingRule` holds it; the trim (`Prose.Read`) is applied to `--question` and pinned by a test that passes a value ending in a newline |
| R11 | Another plugin or an outside reader depends on a FEOV record word | L × M × L | None does: §III.2's not-carriers command → 0, re-run in §V.1 |

## V. Verification Plan

Before any FEOV command: `unset $(env | grep -o '^FEOV_[A-Z_]*')`; `GOTOOLCHAIN=local`; `. $K/kit.sh`. Logs:
`~/.claude/scratch/inquiry-rename/logs/`; the kit's `tally` counts a run's top-level `--- PASS/FAIL/SKIP` and its
`ok`, `?` and `FAIL` package lines, keeping the log as `last.log`.

### V.1 Generators and censuses · re-armed by any edit to a source in §II's table, and by every merge of main

1. Each generator of §II re-run, then its check: `protogen` and `vocabdoc` leave `git diff --exit-code` clean;
   `schemagen -check`; `go -C T run ./cmd/feov-record --schema` → `E` (§III.9 step 3); `agentgen -check`, gate `enginegen`.
2. S1, S2 and R11's command, as §I and §III.2 state them.

### V.2 Suites · steps 1 to 3 on the first complete tree and on their re-arm surface (§III.9); step 4 and CI on every attempt

1. `go -C T test -count=1 -timeout 90m ./... 2>&1 | tally` → no `FAIL`, and `ok` equal to the same command's count
   in the base image (§III.9); the branch adds no package.
2. `go -C scripts test -count=1 ./... 2>&1 | tally` → no `FAIL`. `node --test plugins/frank-exchange-of-views/tests/simulator/*.test.mjs` → `fail 0`.
3. **S6, by hand.** Delete, one at a time, each of M1–M15's registry rows, the `question` rename, the two
   `InquiryStatus` rows and the `inquiryArm` field rename; run `go -C T test -count=1 ./internal/record/migrate/ 2>&1 |
   tally`. Each prints a `FAIL` naming a test (the PR lists the pairs); one that prints none is a missing assertion,
   fixed before merge. Likewise restore `body.Reason` in `propose` (N10).
4. `go -C scripts run ./check > …/check.log 2>&1` → the summary line's failed count is 0. `qlty` and `actionlint`
   run only in CI: read the PR's checks by fail count.

### V.3 Real data: every archived run, before and after (S3, S4) · re-armed per §III.9

1. **S3.** `go -C T test -count=1 -v -run '^(TestEveryArchivedRunRenders|TestEveryArchivedRunMigrates)$'
   ./internal/record/migrate/ 2>&1 | tally` → `PASS: 2`, no `FAIL`; `command grep -c '^    --- PASS:
   TestEveryArchivedRunMigrates/' …/last.log` equals `git ls-files 'run-archive/*.tar.gz' | wc -l` (20 today).
2. `go -C T run ./devcmd/archdump -archives run-archive -out ~/.claude/scratch/inquiry-rename/before` in the base
   image (§III.9); then on the branch with `-out …/after` → one directory per tarball, `cat …/after/*/REFUSALS | wc -l`
   → 0, no `…/after/*/ASSEMBLE_ERROR`.
3. Apply the word map to a COPY of `before` (`wordmap.py`, kept in scratch and pasted in the PR) — in `report.md`
   the `**Fate:**` word; in `inquiries.md` its file name, nav label, title, the two fate headings and the `path:`
   words; in `README.md` and the link bars the file name and label; in `run.md` and the dashboard the two metric
   names and the status keys; in `debate.md` and `judgments.md` the verb and subject words where an act is named —
   then `diff -r …/before-mapped …/after` → **no output**; a line that differs is a defect. No archive rests at
   `proposed` (`command grep -c '^## proposed' …/before/*/avenues.md` → 0 each), so N10 changes no document.
4. Read by hand, in `after`: `2026-10-11_is-91-prime-m20/report.md`'s three inquiry sections and `inquiries.md`
   whole; `2026-09-11_is-91-prime-b9/inquiries.md` (`## underway (7)`).

### V.4 Real data: a record at `E0` is refused, then migrates (S5) · re-armed per §III.9

1. With `feov-record-E0` (built in the base image, rebuilt on a re-bump), make a real record at `E0`: extract
   `run-archive/2026-10-12_is-91-prime-wave-b.tar.gz` to `~/.claude/scratch/inquiry-rename/archived/`, then
   `feov-record-E0 migrate --from …/archived/<run> --to …/at-e0/wave-b`.
2. Branch binary: `feov-record show --run …/at-e0/wave-b --seat-id red-chair inquiries` → **refused**, naming `E0`,
   `E` and `migrate`; `TestEveryArchivedRunBehindThisBinaryIsRefusedUnmigrated` (migrate) passes over every archive.
3. `feov-record migrate --from …/at-e0/wave-b --to …/at-e/wave-b` → exit 0, 0 refusals in `inputs/migration.json`.
4. `feov-record show --run …/at-e/wave-b --seat-id red-chair inquiries --json | jq '[.inquiries[].status] | group_by(.) | map({(.[0]): length}) | add'`
   → `{"confirmed": 4}`; `… | jq '.inquiries[0] | keys'` holds `question` and not `line`.

### V.5 The release fuzz · re-armed per §III.9, and by an epoch re-bump

1. `FEOV_RELEASE_GATE=1 go -C T test -count=1 -v -timeout 40m ./releasegate/fuzz/ > …/fuzz.log 2>&1` →
   `command grep -o 'fuzzed [0-9]* debate runs · [0-9]* failed' …/fuzz.log` prints `fuzzed 40 debate runs · 0 failed`;
   `^--- FAIL` and `^--- SKIP` (the variable unset) count 0; the coverage table shows `inquiry record` driving
   `underway`, `confirmed`, `refuted`, `deferred`, `declined`, `abandoned`.
2. `FEOV_RELEASE_GATE=1 FUZZ_N=6 FUZZ_C=6 go -C T test -count=1 -race -v -timeout 30m -run TestFuzzDebate ./releasegate/fuzz/`
   → `command grep -c 'DATA RACE'` → 0. A red in either is run on `origin/main` before the branch is blamed.

### V.6 The engine runs · BEFORE merge, on the first complete tree (N15) · re-armed only by the branch's own seat-surface change (§III.9)

1. `scripts/universe.sh build` from the finished branch; prove it carries the change, with `command grep -r` (the
   installed copy is no git repository, where `git grep` errors and prints nothing): `-rlIi 'inquiry record'` over
   `<the installed plugin copy>/agents` prints the chair's and blue's definitions; `-ac 'inquiry record'` over its
   `bin/feov-record` is non-zero; `-rlIiE "$OLD"` over the copy prints only `tools/internal/record/migrate/`,
   `tools/research/`, the OCR corpus and files holding English-list lines.
2. Before the run: `CLAUDE_CONFIG_DIR=<universe>/config claude auth status` shows `"authMethod": "claude.ai"` (a
   config that lost the subscription bills API credits), and every `feov-record` call from outside the run names
   the run's own binary by absolute path (a bare one reaches the host cache at the earlier epoch and is refused).
3. `SMOKE=1 scripts/universe.sh run "<topic>"`, detached, its stream to a log: one lane, `haiku` on both tiers,
   `--mint-budget 3 --k-max 2`; `scripts/universe.sh watch` reads the board. `$K/smoked` and `$K/surface.<S>` are
   written for the commit it ran.
4. It reaches VERIFIED or CEILING before the PR merges. Read: the frontier seat's first `inquiry propose` carries
   `--question` and is accepted first time; at least one `inquiry record --as confirmed|refuted`; every refusal on
   the `inquiry` group (`feov-record dashboard`, `ops log`) with the help page or prompt sentence that taught what
   the seat typed; the chair's `inquiry review`; `inquiries.md` in the report set; outcome and wall clock against
   the last smoke. A refusal that traces to the rename is fixed; the fix, a hand change, re-arms the smoke.

### V.7 The auditor gate

`/prosthetic-conscience:plan-audit plans/avenue-becomes-inquiry.md` — approved on the second consecutive PASS, or by
gblock when every remaining gap is `local`. Its NOTES travel into the PR with S1's English list and excluded set, the
fate and earlier-name, `line` and `move` tables, the migrate literal classes, the deleted passages and moved
sentences (R7), N19's final questions and answers, the §III.7 table after, the S6 pairs, the kit, `final.py` and
their output, main's rate of change and the attempt and smoke figures (N17), the smoke's outcome, and §III.2 re-run.
