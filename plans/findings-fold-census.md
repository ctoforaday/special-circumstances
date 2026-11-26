# Census appendix to `findings-fold.md` (#1695)

Run on `origin/main` `6a5252aa6`; audit p2 re-ran rows A01–X2 on `55187482b` and they reproduce, with
`failedblue_test.go` added to A12, M5, M6, M7 (below); M8's debate.js lines, M9, M11's new rows and M13
are `3eca8f516`'s; audit p3-A re-ran A01–X2 on `3eca8f516` and on #1828's head `c6f061897` (A12 191 hits /
21 files); M14 and M15's grounds rows are `c6f061897`'s; audit p4 re-ran A01–X2 on `94b43bac0`
(`~/.claude/scratch/1695/audit-p4/cat/`: same files, line drift; M4, M5, M7, M13 counts restated below),
and M11's reader table and M15's tombstone rows are `94b43bac0`'s, M15's `Tombstone` body rows (command (5)) `32f152f65`'s; audit p5 re-ran A01–X2, M13, M14, M15 (1), (4), (5) and (6) on `32f152f65` (`~/.claude/scratch/1695/audit-p5/`: they reproduce), and this revision added on `32f152f65` M1's dashboard rows, M11's three untagged readers, A13's `cli/seat/verbs.go:300` (#1874) and command (5)'s test-side lists. A and B land together (§V), so a row's change is the change's,
not a part's interim state. One line per file: `path (hits) — change`. `F/` = `plugins/frank-exchange-of-views/`.
Re-run: rows A01–X2 through the plan's §V check 5 (a copy of `audit-p2/r2/scripts/cats.sh` reading
`REV` and `OUT` from the environment), and the row commands below for A13, M8–M17 (M16's and M17's on `a488cad41`). A file listed under a row changes as that row says; a hit in the
row's raw output that is another sense is kept and named in §"Other-sense hits read".

## A01 finding writers/schema — Part A — 53 hits / 24 files

- `F/tools/internal/anchor/anchor.go` (1) — deleted
- `F/tools/internal/cli/lens/command.go` (2) — deleted
- `F/tools/internal/cli/lens/finding.go` (9) — deleted
- `F/tools/internal/cli/lens/mint.go` (1) — deleted
- `F/tools/internal/consistency/consistency.go` (1) — deleted
- `F/tools/internal/flags/shapes.go` (4) — deleted
- `F/tools/internal/record/actjudge.go` (1) — deleted
- `F/tools/internal/record/available.go` (1) — deleted
- `F/tools/internal/record/citationid.go` (1) — deleted
- `F/tools/internal/record/enums.go` (1) — deleted
- `F/tools/internal/record/gapacts.go` (1) — deleted
- `F/tools/internal/record/gapstates.go` (1) — deleted
- `F/tools/internal/record/id.go` (4) — deleted
- `F/tools/internal/record/inquirycheck.go` (1) — deleted
- `F/tools/internal/record/migrate/remap.go` (2) — deleted
- `F/tools/internal/record/migrate/replay.go` (2) — deleted
- `F/tools/internal/record/mintbudget.go` (1) — deleted
- `F/tools/internal/record/record.go` (2) — deleted
- `F/tools/internal/record/recordpb/record.proto` (4) — deleted
- `F/tools/internal/record/recordsql/schema.go` (2) — deleted
- `F/tools/internal/record/replay.go` (1) — deleted
- `F/tools/internal/record/viewjson.go` (5) — deleted
- `F/tools/internal/report/assemble.go` (3) — deleted
- `F/tools/internal/verify/verify.go` (2) — deleted

## A02 finding dispute routing — Part A — 35 hits / 16 files

- `F/tools/internal/cli/lens/finding.go` (1) — deleted; the redirect names the refused seat's road per case (plan §III.A2): own gap — supersede or regrade; another seat's — nothing more, or a proposed act
- `F/tools/internal/cli/seat/help/dispatch.md` (1) — deleted; the redirect names the refused seat's road per case (plan §III.A2): own gap — supersede or regrade; another seat's — nothing more, or a proposed act
- `F/tools/internal/cli/seat/help/finding.md` (1) — deleted; the redirect names the refused seat's road per case (plan §III.A2): own gap — supersede or regrade; another seat's — nothing more, or a proposed act
- `F/tools/internal/cli/seat/help/mint.md` (1) — deleted; the redirect names the refused seat's road per case (plan §III.A2): own gap — supersede or regrade; another seat's — nothing more, or a proposed act
- `F/tools/internal/cli/seat/verbs.go` (1) — deleted; the redirect names the refused seat's road per case (plan §III.A2): own gap — supersede or regrade; another seat's — nothing more, or a proposed act
- `F/tools/internal/record/available.go` (13) — deleted; the redirect names the refused seat's road per case (plan §III.A2): own gap — supersede or regrade; another seat's — nothing more, or a proposed act
- `F/tools/internal/record/composedsource.go` (1) — deleted; the redirect names the refused seat's road per case (plan §III.A2): own gap — supersede or regrade; another seat's — nothing more, or a proposed act
- `F/tools/internal/record/grounds.go` (2) — deleted; the redirect names the refused seat's road per case (plan §III.A2): own gap — supersede or regrade; another seat's — nothing more, or a proposed act
- `F/tools/internal/record/inquirycheck.go` (1) — deleted; the redirect names the refused seat's road per case (plan §III.A2): own gap — supersede or regrade; another seat's — nothing more, or a proposed act
- `F/tools/internal/record/itemkind.go` (3) — deleted; the redirect names the refused seat's road per case (plan §III.A2): own gap — supersede or regrade; another seat's — nothing more, or a proposed act
- `F/tools/internal/record/onegapperclass.go` (2) — deleted; the redirect names the refused seat's road per case (plan §III.A2): own gap — supersede or regrade; another seat's — nothing more, or a proposed act
- `F/tools/internal/record/queue.go` (1) — deleted; the redirect names the refused seat's road per case (plan §III.A2): own gap — supersede or regrade; another seat's — nothing more, or a proposed act
- `F/tools/internal/record/recordpb/record.proto` (1) — deleted; the redirect names the refused seat's road per case (plan §III.A2): own gap — supersede or regrade; another seat's — nothing more, or a proposed act
- `F/tools/internal/record/retirement.go` (4) — deleted; the redirect names the refused seat's road per case (plan §III.A2): own gap — supersede or regrade; another seat's — nothing more, or a proposed act
- `F/tools/internal/record/stale.go` (1) — deleted; the redirect names the refused seat's road per case (plan §III.A2): own gap — supersede or regrade; another seat's — nothing more, or a proposed act
- `F/tools/internal/record/supersession.go` (1) — deleted; the redirect names the refused seat's road per case (plan §III.A2): own gap — supersede or regrade; another seat's — nothing more, or a proposed act

## A03 mint-cap overflow — Part — — 18 hits / 8 files

- `F/skills/research-protocol/scripts/debate.js` (1) — #1828's; `lens-duties.md:5` reworded in Part A
- `F/tools/internal/cli/lens/command.go` (1) — #1828's; `lens-duties.md:5` reworded in Part A
- `F/tools/internal/cli/seat/help/mint.md` (1) — #1828's; `lens-duties.md:5` reworded in Part A
- `F/tools/internal/cli/seat/verbs.go` (1) — #1828's; `lens-duties.md:5` reworded in Part A
- `F/tools/internal/record/mintbudget.go` (10) — #1828's; `lens-duties.md:5` reworded in Part A
- `F/tools/internal/record/viewjson.go` (1) — #1828's; `lens-duties.md:5` reworded in Part A
- `scripts/agentgen/src/fragments/lens-duties.md` (2) — #1828's; `lens-duties.md:5` reworded in Part A
- `scripts/agentgen/src/fragments/log-duty-red.md` (1) — #1828's; `lens-duties.md:5` reworded in Part A

## A04 inquiry notes — Part A — 17 hits / 10 files

- `F/skills/research-protocol/scripts/debate.js` (1) — finding arm deleted (`capture/*` other sense, kept)
- `F/tools/internal/capture/capture.go` (1) — finding arm deleted (`capture/*` other sense, kept)
- `F/tools/internal/capture/enveloperead.go` (1) — finding arm deleted (`capture/*` other sense, kept)
- `F/tools/internal/cli/lens/finding.go` (1) — finding arm deleted (`capture/*` other sense, kept)
- `F/tools/internal/cli/seat/help/check.md` (1) — finding arm deleted (`capture/*` other sense, kept)
- `F/tools/internal/consistency/consistency.go` (1) — finding arm deleted (`capture/*` other sense, kept)
- `F/tools/internal/record/inquirycheck.go` (8) — finding arm deleted (`capture/*` other sense, kept)
- `F/tools/internal/record/passblockers.go` (1) — finding arm deleted (`capture/*` other sense, kept)
- `scripts/agentgen/src/fragments/lens-audit-and-mint.md` (1) — finding arm deleted (`capture/*` other sense, kept)
- `scripts/agentgen/src/red-lens-logic.md` (1) — finding arm deleted (`capture/*` other sense, kept)

## A05 found_by — Part A — 94 hits / 25 files

- `F/skills/research-protocol/scripts/debate.js` (1) — deleted
- `F/tools/internal/cli/lens/finding.go` (2) — deleted
- `F/tools/internal/cli/lens/mint.go` (5) — deleted
- `F/tools/internal/cli/motion/verbs.go` (1) — deleted
- `F/tools/internal/cli/seat/help/mint.md` (1) — deleted
- `F/tools/internal/cli/seat/seat.go` (2) — deleted
- `F/tools/internal/cli/seat/selector.go` (1) — deleted
- `F/tools/internal/cli/seat/verbs.go` (2) — deleted
- `F/tools/internal/cli/verify.go` (1) — deleted
- `F/tools/internal/flags/csv.go` (2) — deleted
- `F/tools/internal/flags/lensareas.go` (1) — deleted
- `F/tools/internal/flags/names.go` (2) — deleted
- `F/tools/internal/flags/shapes.go` (2) — deleted
- `F/tools/internal/record/id.go` (1) — deleted
- `F/tools/internal/record/mintbudget.go` (2) — deleted
- `F/tools/internal/record/record.go` (1) — deleted
- `F/tools/internal/record/recordpb/record.proto` (4) — deleted
- `F/tools/internal/record/refs.go` (4) — deleted
- `F/tools/internal/record/replay.go` (1) — deleted
- `F/tools/internal/record/required.go` (1) — deleted
- `F/tools/internal/record/viewjson.go` (23) — deleted
- `F/tools/internal/report/assemble.go` (14) — deleted
- `F/tools/internal/seatenv/identity.go` (1) — deleted
- `F/tools/internal/verify/verify.go` (12) — deleted
- `F/tools/internal/view/view.go` (7) — deleted

## A06 unminted/observation views — Part A — 82 hits / 21 files

- `F/tools/internal/anchor/anchor.go` (1) — deleted
- `F/tools/internal/anchor/window.go` (1) — deleted
- `F/tools/internal/cli/blue/edit.go` (1) — deleted
- `F/tools/internal/cli/lens/finding.go` (1) — deleted
- `F/tools/internal/cli/lens/mint.go` (1) — deleted
- `F/tools/internal/cli/seat/help/mint.md` (1) — deleted
- `F/tools/internal/cli/seat/selector.go` (2) — deleted
- `F/tools/internal/cli/seat/verbs.go` (5) — deleted
- `F/tools/internal/consistency/consistency.go` (8) — deleted
- `F/tools/internal/dashboard/model.go` (1) — deleted
- `F/tools/internal/dashboard/render.go` (2) — deleted
- `F/tools/internal/flags/shapes.go` (3) — deleted
- `F/tools/internal/record/evidenceview.go` (1) — deleted
- `F/tools/internal/record/gapstates.go` (4) — deleted
- `F/tools/internal/record/id.go` (2) — deleted
- `F/tools/internal/record/refs.go` (1) — deleted
- `F/tools/internal/record/replay.go` (2) — deleted
- `F/tools/internal/record/viewjson.go` (27) — deleted
- `F/tools/internal/report/assemble.go` (10) — deleted
- `F/tools/internal/verify/verify.go` (4) — deleted
- `F/tools/internal/view/view.go` (4) — deleted

## A07 F- anchor — Part A — 55 hits / 29 files

- `F/tools/internal/anchor/anchor.go` (4) — the kind leaves
- `F/tools/internal/anchor/window.go` (2) — the kind leaves
- `F/tools/internal/anchortext/anchortext.go` (3) — the kind leaves
- `F/tools/internal/claimcount/claimcount.go` (1) — the kind leaves
- `F/tools/internal/claimcount/paragraphs.go` (1) — the kind leaves
- `F/tools/internal/cli/blue/edit.go` (1) — the kind leaves
- `F/tools/internal/cli/blue/retire.go` (1) — the kind leaves
- `F/tools/internal/cli/lens/close.go` (1) — the kind leaves
- `F/tools/internal/cli/lens/finding.go` (3) — the kind leaves
- `F/tools/internal/cli/seat/selector.go` (1) — the kind leaves
- `F/tools/internal/cli/seat/verbs.go` (3) — the kind leaves
- `F/tools/internal/flags/csv.go` (1) — the kind leaves
- `F/tools/internal/flags/names.go` (1) — the kind leaves
- `F/tools/internal/flags/shapes.go` (4) — the kind leaves
- `F/tools/internal/hookgate/hookgate.go` (1) — the kind leaves
- `F/tools/internal/record/available.go` (1) — the kind leaves
- `F/tools/internal/record/estoppel.go` (2) — the kind leaves
- `F/tools/internal/record/grounds.go` (1) — the kind leaves
- `F/tools/internal/record/itemkind.go` (1) — the kind leaves
- `F/tools/internal/record/recordpb/record.proto` (2) — the kind leaves
- `F/tools/internal/record/refs.go` (3) — the kind leaves
- `F/tools/internal/record/viewjson.go` (1) — the kind leaves
- `F/tools/internal/report/docs.go` (1) — the kind leaves
- `F/tools/internal/report/md.go` (1) — the kind leaves
- `F/tools/internal/report/seatread.go` (2) — the kind leaves
- `F/tools/internal/reportproj/splice.go` (6) — the kind leaves
- `F/tools/internal/reportvoice/tells.go` (1) — the kind leaves
- `F/tools/internal/seatprobe/boards.go` (1) — the kind leaves
- `F/tools/internal/terms/terms.json` (4) — the kind leaves

## A08 migrate — Part A — 32 hits / 7 files

- `F/tools/internal/record/migrate/entries.go` (2) — Part A migration (`gapclass.go:48` other sense, kept)
- `F/tools/internal/record/migrate/eraentries.go` (3) — Part A migration (`gapclass.go:48` other sense, kept)
- `F/tools/internal/record/migrate/gapclass.go` (1) — Part A migration (`gapclass.go:48` other sense, kept)
- `F/tools/internal/record/migrate/remap.go` (12) — Part A migration (`gapclass.go:48` other sense, kept)
- `F/tools/internal/record/migrate/replay.go` (12) — Part A migration (`gapclass.go:48` other sense, kept)
- `F/tools/internal/record/migrate/valuerenames.go` (1) — Part A migration (`gapclass.go:48` other sense, kept)
- `F/tools/internal/record/migration/manifest.go` (1) — Part A migration (`gapclass.go:48` other sense, kept)

## A09a help pages — Part A — 13 hits / 6 files

- `F/tools/internal/cli/seat/help/check.md` (1) — rewritten
- `F/tools/internal/cli/seat/help/close.md` (1) — rewritten
- `F/tools/internal/cli/seat/help/edit.md` (1) — rewritten
- `F/tools/internal/cli/seat/help/finding.md` (2) — rewritten
- `F/tools/internal/cli/seat/help/mint.md` (7) — rewritten
- `F/tools/internal/cli/seat/help/verify.md` (1) — rewritten

## A09b agentgen sources — Part A — 32 hits / 12 files

- `scripts/agentgen/src/fragments/blue-search-and-cite.md` (1) — rewritten; regenerate
- `scripts/agentgen/src/fragments/lens-audit-and-mint.md` (3) — rewritten; regenerate
- `scripts/agentgen/src/fragments/lens-duties.md` (3) — rewritten; regenerate
- `scripts/agentgen/src/red-lens-adversary.md` (2) — rewritten; regenerate
- `scripts/agentgen/src/red-lens-architecture.md` (2) — rewritten; regenerate
- `scripts/agentgen/src/red-lens-computation.md` (2) — rewritten; regenerate
- `scripts/agentgen/src/red-lens-dark-side.md` (2) — rewritten; regenerate
- `scripts/agentgen/src/red-lens-evidence.md` (2) — rewritten; regenerate
- `scripts/agentgen/src/red-lens-logic.md` (3) — rewritten; regenerate
- `scripts/agentgen/src/red-lens-voice.md` (3) — rewritten; regenerate
- `scripts/agentgen/src/skills/adversarial-audit/SKILL.md` (7) — rewritten; regenerate
- `scripts/agentgen/src/skills/research-protocol/SKILL.md` (2) — rewritten; regenerate

## A09c debate.js — Part A — 14 hits / 1 files

- `F/skills/research-protocol/scripts/debate.js` (14) — rewritten

## A09d terms.json — Part A — 9 hits / 1 files

- `F/tools/internal/terms/terms.json` (9) — rewritten; vocabgen + vocabdoc

## A09e Go flag text — Part A — 37 hits / 10 files

- `F/tools/internal/cli/fetch.go` (1) — rewritten
- `F/tools/internal/cli/fetchsummary.go` (1) — rewritten
- `F/tools/internal/cli/lens/finding.go` (13) — rewritten
- `F/tools/internal/cli/lens/mint.go` (2) — rewritten
- `F/tools/internal/cli/root.go` (1) — rewritten
- `F/tools/internal/cli/seat/seat.go` (2) — rewritten
- `F/tools/internal/cli/seat/selector.go` (1) — rewritten
- `F/tools/internal/cli/seat/verbs.go` (7) — rewritten
- `F/tools/internal/cli/verify.go` (1) — rewritten
- `F/tools/internal/flags/shapes.go` (8) — rewritten

## A10 docs — Part A — 100 hits / 12 files

- `README.md` (2) — rewritten
- `F/README.md` (1) — rewritten
- `F/docs/citations-and-bibliography.md` (3) — rewritten
- `F/docs/finding-markers.md` (33) — rewritten
- `F/docs/propagation-and-anchoring.md` (3) — rewritten
- `F/docs/record-flow.md` (12) — rewritten
- `F/docs/seat-command-triggers.md` (5) — rewritten
- `F/docs/seat-surface-naming.md` (22) — rewritten
- `F/docs/why-a-seat-stops.md` (13) — rewritten
- `F/hooks/README.md` (1) — rewritten
- `F/skills/research-protocol/references/report_template.md` (4) — rewritten
- `F/skills/research/SKILL.md` (1) — rewritten

## A11 seatprobe — Part A — 8 hits / 3 files

- `F/tools/internal/seatprobe/attempted.go` (1) — rewritten
- `F/tools/internal/seatprobe/attempted_test.go` (3) — rewritten
- `F/tools/internal/seatprobe/boards.go` (4) — rewritten

## A12 fuzz/sweeptally — Part A — 189 hits / 20 files

- `F/tools/releasegate/fuzz/anchorevidence_test.go` (1) — Part A fuzz
- `F/tools/releasegate/fuzz/anothersource_test.go` (4) — Part A fuzz
- `F/tools/releasegate/fuzz/benchlift_test.go` (1) — Part A fuzz
- `F/tools/releasegate/fuzz/converseoracle_test.go` (1) — Part A fuzz
- `F/tools/releasegate/fuzz/demandanswer_test.go` (1) — Part A fuzz
- `F/tools/releasegate/fuzz/failedblue_test.go` (2) — fixture sentence "to anchor findings" (`:73`, `:269`) reworded
- `F/tools/releasegate/fuzz/forcedbound_test.go` (1) — Part A fuzz
- `F/tools/releasegate/fuzz/fuzz_test.go` (94) — Part A fuzz
- `F/tools/releasegate/fuzz/lensturn_test.go` (1) — Part A fuzz
- `F/tools/releasegate/fuzz/lensturnacts_test.go` (4) — Part A fuzz
- `F/tools/releasegate/fuzz/lifecycleseed_test.go` (3) — Part A fuzz
- `F/tools/releasegate/fuzz/lostreading_test.go` (4) — Part A fuzz
- `F/tools/releasegate/fuzz/onegaptally_test.go` (4) — Part A fuzz
- `F/tools/releasegate/fuzz/owedcomputation_test.go` (1) — Part A fuzz
- `F/tools/releasegate/fuzz/parts_test.go` (2) — Part A fuzz
- `F/tools/releasegate/fuzz/reading_test.go` (12) — Part A fuzz
- `F/tools/releasegate/fuzz/readingdebt_test.go` (11) — Part A fuzz
- `F/tools/releasegate/fuzz/registerbeforeappend_test.go` (1) — Part A fuzz
- `F/tools/releasegate/fuzz/staleretry_test.go` (11) — Part A fuzz
- `F/tools/releasegate/fuzz/supersede_test.go` (1) — Part A fuzz
- `F/tools/releasegate/fuzz/sweeptally_test.go` (31) — Part A fuzz

## M1 grade machinery — Part B — 76 hits / 20 files

- `F/skills/research-protocol/scripts/debate.js` (3) — deleted, or moved to the regrade proposal
- `F/tools/devcmd/enginegen/vocabulary.go` (1) — deleted, or moved to the regrade proposal
- `F/tools/internal/cli/motion/verbs.go` (3) — deleted, or moved to the regrade proposal
- `F/tools/internal/cli/seat/verbs.go` (1) — deleted, or moved to the regrade proposal
- `F/tools/internal/record/available.go` (23) — deleted, or moved to the regrade proposal
- `F/tools/internal/record/benchgrade.go` (2) — deleted, or moved to the regrade proposal
- `F/tools/internal/record/itemkind.go` (3) — deleted, or moved to the regrade proposal
- `F/tools/internal/record/keptgrade.go` (1) — deleted, or moved to the regrade proposal
- `F/tools/internal/record/migrate/eraentries.go` (4) — deleted, or moved to the regrade proposal
- `F/tools/internal/record/migrate/remap.go` (1) — deleted, or moved to the regrade proposal
- `F/tools/internal/record/motion.go` (5) — deleted, or moved to the regrade proposal
- `F/tools/internal/record/planguard/growth.go` (1) — deleted, or moved to the regrade proposal
- `F/tools/internal/record/queue.go` (1) — deleted, or moved to the regrade proposal
- `F/tools/internal/record/record.go` (7) — deleted, or moved to the regrade proposal
- `F/tools/internal/record/recordpb/motionsubject.go` (1) — deleted, or moved to the regrade proposal
- `F/tools/internal/record/recordpb/record.proto` (7) — deleted, or moved to the regrade proposal
- `F/tools/internal/record/recordsql/schema.go` (3) — deleted, or moved to the regrade proposal
- `F/tools/internal/record/retirement.go` (3) — deleted, or moved to the regrade proposal
- `F/tools/internal/record/viewjson.go` (4) — deleted, or moved to the regrade proposal
- `F/tools/internal/verify/verify.go` (2) — deleted, or moved to the regrade proposal
- Added on `32f152f65` (audit p5 NOTE; the subject's string `"grade"`, not the row's pattern): `F/tools/internal/dashboard/model.go:73,486-492` (`Judiciary.Disputes` and its `case "grade"` arm) — deleted; `F/tools/internal/dashboard/render.go:381-386,406` (the heading's "grade motions", the comment, the section gate's `Disputes.Raised != 0` arm, the "grade motions" row) — deleted; `dashboard/judiciary_test.go:125-126`, `render_test.go:47` — deleted; `dashboard/testdata/render-{decided,live,terminal}.golden` D — the row and the heading's words go. Not re-pointed at act motions: the "rulings by type" row (`j.Rulings`, M11 bench-only) already counts the bench's `accepted` and `rejected` (plan §III.B6)

## M2 docket readers — Part B — 129 hits / 34 files

Each file's `Ruled()` / `m.ruled` / docket-ruling reads are tagged in M11's reader table (bench-only | any answer | remand-only); the blanket "an accepted ruling is a ruling" is withdrawn (audit p4 gap 2).

- `F/tools/internal/cli/seat/verbs.go` (1) — read per reader in M11's reader table (audit p4 gap 2)
- `F/tools/internal/dashboard/model.go` (1) — read per reader in M11's reader table (audit p4 gap 2)
- `F/tools/internal/graph/graph.go` (1) — read per reader in M11's reader table (audit p4 gap 2)
- `F/tools/internal/record/actjudge.go` (3) — read per reader in M11's reader table (audit p4 gap 2)
- `F/tools/internal/record/available.go` (2) — read per reader in M11's reader table (audit p4 gap 2)
- `F/tools/internal/record/benchgrade.go` (3) — read per reader in M11's reader table (audit p4 gap 2)
- `F/tools/internal/record/benchsitting.go` (10) — read per reader in M11's reader table (audit p4 gap 2)
- `F/tools/internal/record/closingowed.go` (14) — read per reader in M11's reader table (audit p4 gap 2)
- `F/tools/internal/record/composedsource.go` (1) — read per reader in M11's reader table (audit p4 gap 2)
- `F/tools/internal/record/dispatch.go` (17) — read per reader in M11's reader table (audit p4 gap 2)
- `F/tools/internal/record/grademovement.go` (1) — read per reader in M11's reader table (audit p4 gap 2)
- `F/tools/internal/record/inquirycheck.go` (1) — read per reader in M11's reader table (audit p4 gap 2)
- `F/tools/internal/record/migrate/replay.go` (2) — read per reader in M11's reader table (audit p4 gap 2)
- `F/tools/internal/record/motion.go` (8) — read per reader in M11's reader table (audit p4 gap 2)
- `F/tools/internal/record/motionview.go` (7) — read per reader in M11's reader table (audit p4 gap 2)
- `F/tools/internal/record/passblockers.go` (6) — unchanged — a ruling for either side, or a grade answer, is a ruling (`rulingWord`); `requireBenchMotionsRuled`'s refusal (`:933`) names D's pair for a docketing (§III.B4, M17 amendment 4)
- `F/tools/internal/record/positionheard.go` (1) — read per reader in M11's reader table (audit p4 gap 2)
- `F/tools/internal/record/record.go` (11) — read per reader in M11's reader table (audit p4 gap 2)
- `F/tools/internal/record/recordread.go` (1) — read per reader in M11's reader table (audit p4 gap 2)
- `F/tools/internal/record/recordsql/gapstate.go` (4) — read per reader in M11's reader table (audit p4 gap 2)
- `F/tools/internal/record/recordsql/schema.go` (2) — read per reader in M11's reader table (audit p4 gap 2)
- `F/tools/internal/record/recordsql/views.go` (4) — read per reader in M11's reader table (audit p4 gap 2)
- `F/tools/internal/record/retirement.go` (1) — read per reader in M11's reader table (audit p4 gap 2)
- `F/tools/internal/record/sitting.go` (5) — read per reader in M11's reader table (audit p4 gap 2)
- `F/tools/internal/record/stale.go` (1) — read per reader in M11's reader table (audit p4 gap 2)
- `F/tools/internal/record/supersession.go` (4) — read per reader in M11's reader table (audit p4 gap 2)
- `F/tools/internal/record/viewjson.go` (4) — read per reader in M11's reader table (audit p4 gap 2)
- `F/tools/internal/report/assemble.go` (2) — read per reader in M11's reader table (audit p4 gap 2)
- `F/tools/internal/report/motions.go` (1) — read per reader in M11's reader table (audit p4 gap 2)
- `F/tools/internal/scorecard/scorecard.go` (1) — read per reader in M11's reader table (audit p4 gap 2)
- `F/tools/internal/seatprobe/boards.go` (4) — read per reader in M11's reader table (audit p4 gap 2)
- `F/tools/internal/seatprobe/build.go` (3) — read per reader in M11's reader table (audit p4 gap 2)
- `F/tools/internal/verify/verify.go` (1) — read per reader in M11's reader table (audit p4 gap 2)
- `F/tools/internal/view/view.go` (1) — read per reader in M11's reader table (audit p4 gap 2)

## M3 docket filing and gavel — Part B — 86 hits / 23 files

- `F/tools/internal/cli/motion/command.go` (3) — filing checks and the minter's accept (Part B §III.B2–B3)
- `F/tools/internal/cli/motion/verbs.go` (5) — filing checks and the minter's accept (Part B §III.B2–B3)
- `F/tools/internal/consistency/consistency.go` (1) — filing checks and the minter's accept (Part B §III.B2–B3)
- `F/tools/internal/graph/graph.go` (1) — filing checks and the minter's accept (Part B §III.B2–B3)
- `F/tools/internal/record/actjudge.go` (1) — filing checks and the minter's accept (Part B §III.B2–B3)
- `F/tools/internal/record/available.go` (1) — filing checks and the minter's accept (Part B §III.B2–B3)
- `F/tools/internal/record/dispatch.go` (12) — filing checks and the minter's accept (Part B §III.B2–B3)
- `F/tools/internal/record/dispatchnext.go` (4) — filing checks and the minter's accept (Part B §III.B2–B3)
- `F/tools/internal/record/grounds.go` (2) — deleted whole (M15, plan §III.B12)
- `F/tools/internal/record/itemkind.go` (3) — filing checks and the minter's accept (Part B §III.B2–B3)
- `F/tools/internal/record/migrate/entries.go` (4) — filing checks and the minter's accept (Part B §III.B2–B3)
- `F/tools/internal/record/migrate/remap.go` (5) — filing checks and the minter's accept (Part B §III.B2–B3)
- `F/tools/internal/record/motion.go` (2) — filing checks and the minter's accept (Part B §III.B2–B3)
- `F/tools/internal/record/record.go` (5) — filing checks and the minter's accept (Part B §III.B2–B3)
- `F/tools/internal/record/recordpb/record.proto` (8) — filing checks and the minter's accept (Part B §III.B2–B3)
- `F/tools/internal/record/recordsql/schema.go` (1) — filing checks and the minter's accept (Part B §III.B2–B3)
- `F/tools/internal/record/recordsql/views.go` (1) — filing checks and the minter's accept (Part B §III.B2–B3)
- `F/tools/internal/record/replay.go` (3) — filing checks and the minter's accept (Part B §III.B2–B3)
- `F/tools/internal/record/rulerfiles.go` (17) — filing checks and the minter's accept (Part B §III.B2–B3)
- `F/tools/internal/record/viewjson.go` (3) — filing checks and the minter's accept (Part B §III.B2–B3)
- `F/tools/internal/seatprobe/seatprobe.go` (1) — filing checks and the minter's accept (Part B §III.B2–B3)
- `F/tools/internal/verify/verify.go` (2) — filing checks and the minter's accept (Part B §III.B2–B3)
- `F/tools/internal/view/view.go` (1) — filing checks and the minter's accept (Part B §III.B2–B3)

## M4 agent text on motions — Part B — 64 hits / 19 files on `94b43bac0`

- `F/skills/research-protocol/scripts/debate.js` (23) — rewritten; `:1055` (#1861's comment on the dispatched docket sitting) is a `docketHeld` path, not a new ruling path: unchanged
- `F/tools/internal/cli/seat/help/close.md` (2) — rewritten
- `F/tools/internal/cli/seat/help/closing-blue.md` (2) — rewritten
- `F/tools/internal/cli/seat/help/closing-red.md` (2) — rewritten
- `F/tools/internal/cli/seat/help/declare.md` (2) — rewritten
- `F/tools/internal/cli/seat/help/dispatch.md` (4) — rewritten
- `F/tools/internal/cli/seat/help/mint.md` (2) — rewritten
- `F/tools/internal/cli/seat/help/outcome.md` (1) — rewritten
- `F/tools/internal/cli/seat/help/position.md` (1) — rewritten
- `F/tools/internal/cli/seat/help/regrade-bench.md` (3) — rewritten
- `F/tools/internal/cli/seat/help/regrade.md` (1) — rewritten
- `F/tools/internal/cli/seat/help/reproduce.md` (1) — rewritten
- `F/tools/internal/cli/seat/help/verify.md` (1) — rewritten
- `F/tools/internal/terms/terms.json` (4) — rewritten
- `scripts/agentgen/src/blue-researcher.md` (1) — rewritten
- `scripts/agentgen/src/fragments/lens-audit-and-mint.md` (1) — rewritten
- `scripts/agentgen/src/fragments/lens-duties.md` (2) — rewritten
- `scripts/agentgen/src/lead-judge.md` (7) — rewritten
- `scripts/agentgen/src/red-chair.md` (4) — rewritten

## M5 fuzz on motions — Part B — 441 hits / 14 files on `94b43bac0`

- `F/tools/releasegate/fuzz/benchlift_test.go` (12) — Part B fuzz
- `F/tools/releasegate/fuzz/benchorders_test.go` (27) — Part B fuzz
- `F/tools/releasegate/fuzz/converseoracle_test.go` (9) — Part B fuzz
- `F/tools/releasegate/fuzz/coverage_test.go` (3) — Part B fuzz
- `F/tools/releasegate/fuzz/failedblue_test.go` — `TestTheAssemblyRulesOnlyTheDocketsFiledAfterTheLastPlan` (`:235`): its `motion docket file` (`:238`) has no act and its gap no counted exchange, so Part B refuses it twice over; it files an act on a gap with one exchange, and the assembly accepts or rejects it (T-B9 path 3)
- `F/tools/releasegate/fuzz/fuzz_test.go` (229) — Part B fuzz
- `F/tools/releasegate/fuzz/lensturn_test.go` (21) — Part B fuzz
- `F/tools/releasegate/fuzz/positionheard_test.go` (11) — Part B fuzz
- `F/tools/releasegate/fuzz/remandroute_test.go` (63) — Part B fuzz
- `F/tools/releasegate/fuzz/siblingdocket_test.go` (15) — Part B fuzz
- `F/tools/releasegate/fuzz/supersede_test.go` (1) — Part B fuzz
- `F/tools/releasegate/fuzz/sweeptally_test.go` (22) — Part B fuzz
- `F/tools/releasegate/fuzz/termination_test.go` (11) — Part B fuzz
- `F/tools/releasegate/fuzz/trajectory_test.go` (3) — Part B fuzz

## X1 about-kind gap — Part A — 11 hits / 8 files

- `F/tools/internal/cli/lens/finding.go` (1) — refused; reserved
- `F/tools/internal/flags/shapes.go` (1) — refused; reserved
- `F/tools/internal/record/about.go` (2) — refused; reserved
- `F/tools/internal/record/available.go` (2) — refused; reserved
- `F/tools/internal/record/gapacts.go` (1) — refused; reserved
- `F/tools/internal/record/grounds.go` (1) — refused; reserved
- `F/tools/internal/record/onegapperclass.go` (2) — refused; reserved
- `F/tools/internal/record/recordpb/record.proto` (1) — refused; reserved

## X2 corroboration — Part — — 19 hits / 9 files

- `F/tools/internal/cli/lens/verify.go` (1) — none (every hit another sense)
- `F/tools/internal/fetchcache/pageread.go` (3) — none (every hit another sense)
- `F/tools/internal/record/migrate/eraentries.go` (5) — none (every hit another sense)
- `F/tools/internal/record/migrate/replay.go` (1) — none (every hit another sense)
- `F/tools/internal/record/recordpb/record.proto` (3) — none (every hit another sense)
- `scripts/agentgen/src/fragments/blue-search.md` (1) — none (every hit another sense)
- `scripts/agentgen/src/fragments/leaf-node-read.md` (1) — none (every hit another sense)
- `scripts/agentgen/src/fragments/lens-audit-and-mint.md` (3) — none (every hit another sense)
- `scripts/agentgen/src/red-lens-evidence.md` (1) — none (every hit another sense)

## M6 Go tests naming motion identifiers — Part B — 133 files

Each is run in Part B's test loop; a test of a deleted identifier is deleted with it. `failedblue_test.go` joined on `55187482b` (M5 says what changes).

- `F/tools/releasegate/fuzz/failedblue_test.go`
- `F/tools/integration/surface/promptverbs_test.go`
- `F/tools/internal/capture/capture_test.go`
- `F/tools/internal/capture/lensturn_test.go`
- `F/tools/internal/cli/actrows_test.go`
- `F/tools/internal/cli/adversarial_test.go`
- `F/tools/internal/cli/answers_test.go`
- `F/tools/internal/cli/bench/declare.go`
- `F/tools/internal/cli/benchorders_test.go`
- `F/tools/internal/cli/checkkind_test.go`
- `F/tools/internal/cli/cli_test.go`
- `F/tools/internal/cli/closuredemand_test.go`
- `F/tools/internal/cli/crossseat_test.go`
- `F/tools/internal/cli/deadpath_test.go`
- `F/tools/internal/cli/democarriers_test.go`
- `F/tools/internal/cli/disputereply_test.go`
- `F/tools/internal/cli/dockethelpers_test.go`
- `F/tools/internal/cli/enum_help_test.go`
- `F/tools/internal/cli/integration_test.go`
- `F/tools/internal/cli/missingflag_test.go`
- `F/tools/internal/cli/motion/command.go`
- `F/tools/internal/cli/motion/verbs.go`
- `F/tools/internal/cli/motion_test.go`
- `F/tools/internal/cli/payload_test.go`
- `F/tools/internal/cli/referencechecks_test.go`
- `F/tools/internal/cli/refs_test.go`
- `F/tools/internal/cli/refusalshape_test.go`
- `F/tools/internal/cli/repeatedflag_test.go`
- `F/tools/internal/cli/requiredmarker_test.go`
- `F/tools/internal/cli/rulerfiles_test.go`
- `F/tools/internal/cli/seat/closure.go`
- `F/tools/internal/cli/seat/help/close.md`
- `F/tools/internal/cli/seat/help/declare.md`
- `F/tools/internal/cli/seat/help/mint.md`
- `F/tools/internal/cli/seat/help/reproduce.md`
- `F/tools/internal/cli/seat/help/verify.md`
- `F/tools/internal/cli/seat/help_gen.go`
- `F/tools/internal/cli/seat/refusal.go`
- `F/tools/internal/cli/seat/seat.go`
- `F/tools/internal/cli/seat/verbs.go`
- `F/tools/internal/cli/siblingverbs_test.go`
- `F/tools/internal/cli/stalerefusal_test.go`
- `F/tools/internal/cli/standing_test.go`
- `F/tools/internal/cli/stategraph_test.go`
- `F/tools/internal/cli/surface.go`
- `F/tools/internal/cli/verbs_test.go`
- `F/tools/internal/cli/vocabulary_test.go`
- `F/tools/internal/consistency/consistency.go`
- `F/tools/internal/consistency/consistency_test.go`
- `F/tools/internal/dashboard/judiciary_test.go`
- `F/tools/internal/difftest/fuzz_test.go`
- `F/tools/internal/difftest/harness_test.go`
- `F/tools/internal/difftest/testdata/a_repeated_act_names_its_answer.golden`
- `F/tools/internal/difftest/testdata/docket_ruling_requires_each_unconditional_field.golden`
- `F/tools/internal/difftest/testdata/error_catalogue.golden`
- `F/tools/internal/difftest/testdata/ids_stay_distinct_across_rounds.golden`
- `F/tools/internal/difftest/testdata/manual_runs_every_help_page_live.golden`
- `F/tools/internal/difftest/testdata/projections_debate_changelog_citations.golden`
- `F/tools/internal/flags/register.go`
- `F/tools/internal/graph/graph_test.go`
- `F/tools/internal/record/available.go`
- `F/tools/internal/record/available_test.go`
- `F/tools/internal/record/benchgrade.go`
- `F/tools/internal/record/benchorders_test.go`
- `F/tools/internal/record/closingowed_test.go`
- `F/tools/internal/record/closureview_test.go`
- `F/tools/internal/record/demand.go`
- `F/tools/internal/record/dispatch_test.go`
- `F/tools/internal/record/dispatchnext.go`
- `F/tools/internal/record/enums.go`
- `F/tools/internal/record/envelope_test.go`
- `F/tools/internal/record/epochreach_test.go`
- `F/tools/internal/record/gapacts_test.go`
- `F/tools/internal/record/grounds.go`
- `F/tools/internal/record/itemkind.go`
- `F/tools/internal/record/lensturn.go`
- `F/tools/internal/record/migrate/benchfiling_test.go`
- `F/tools/internal/record/migrate/entries.go`
- `F/tools/internal/record/migrate/eraentries.go`
- `F/tools/internal/record/migrate/remap.go`
- `F/tools/internal/record/migrate/remaplists_test.go`
- `F/tools/internal/record/migrate/testdata/prose-rewrites.golden`
- `F/tools/internal/record/motion.go`
- `F/tools/internal/record/motionguards_test.go`
- `F/tools/internal/record/motionqueries_parity_test.go`
- `F/tools/internal/record/motionview_test.go`
- `F/tools/internal/record/narrowedview_test.go`
- `F/tools/internal/record/passblockers_test.go`
- `F/tools/internal/record/queries_parity_test.go`
- `F/tools/internal/record/queue.go`
- `F/tools/internal/record/queue_test.go`
- `F/tools/internal/record/record.go`
- `F/tools/internal/record/recordpb/motionsubject.go`
- `F/tools/internal/record/recordpb/record.pb.go`
- `F/tools/internal/record/recordpb/record.proto`
- `F/tools/internal/record/recordsql/decode_gen.go`
- `F/tools/internal/record/recordsql/gapstate_test.go`
- `F/tools/internal/record/recordsql/readmix_test.go`
- `F/tools/internal/record/recordsql/store_test.go`
- `F/tools/internal/record/recordsql/viewfamilies_test.go`
- `F/tools/internal/record/remand.go`
- `F/tools/internal/record/remandedview_test.go`
- `F/tools/internal/record/replay.go`
- `F/tools/internal/record/replay_test.go`
- `F/tools/internal/record/rulerfiles.go`
- `F/tools/internal/record/rulerfiles_test.go`
- `F/tools/internal/record/viewjson.go`
- `F/tools/internal/record/viewjson_test.go`
- `F/tools/internal/record/worklist_test.go`
- `F/tools/internal/record/writerace_test.go`
- `F/tools/internal/refusalcheck/refusalcheck_test.go`
- `F/tools/internal/report/assemble_integration_test.go`
- `F/tools/internal/report/assemble_test.go`
- `F/tools/internal/report/layout.go`
- `F/tools/internal/report/seatinput_test.go`
- `F/tools/internal/report/showsthework_test.go`
- `F/tools/internal/report/storedsitting_test.go`
- `F/tools/internal/scorecard/remanded_test.go`
- `F/tools/internal/seatprobe/attempted.go`
- `F/tools/internal/seatprobe/attempted_test.go`
- `F/tools/internal/seatprobe/boards.go`
- `F/tools/internal/seatprobe/seatprobe.go`
- `F/tools/internal/seatprobe/surfacecoverage_test.go`
- `F/tools/internal/sittingwrite/envelope_test.go`
- `F/tools/internal/verify/verify.go`
- `F/tools/internal/verify/verify_test.go`
- `F/tools/internal/view/view_test.go`
- `F/tools/releasegate/fuzz/benchlift_test.go`
- `F/tools/releasegate/fuzz/benchorders_test.go`
- `F/tools/releasegate/fuzz/fuzz_test.go`
- `F/tools/releasegate/fuzz/remandroute_test.go`
- `F/tools/releasegate/fuzz/siblingdocket_test.go`
- `F/tools/releasegate/fuzz/trajectory_test.go`

## M7 goldens naming a motion — Part B — 218 files on `94b43bac0` (217 goldens and `failedblue_test.go`, which M6 carries)

Regenerated with `UPDATE_GOLDENS=1`; every changed line read. Audit p2's re-run lists `failedblue_test.go` here too, matched through the row command's bare `$F` pathspec; it holds no golden, so M6 carries it. `tests/simulator/testdata/prompt-judge-terminal.golden` changes with the terminal prompt (M8).

- `F/agents/blue-lane.md`
- `F/agents/blue-researcher.md`
- `F/agents/blue-synthesizer.md`
- `F/agents/lead-judge.md`
- `F/agents/red-chair.md`
- `F/agents/red-lens-adversary.md`
- `F/agents/red-lens-architecture.md`
- `F/agents/red-lens-computation.md`
- `F/agents/red-lens-dark-side.md`
- `F/agents/red-lens-evidence.md`
- `F/agents/red-lens-logic.md`
- `F/agents/red-lens-voice.md`
- `F/docs/duty-docket-preregistration.md`
- `F/docs/seat-command-triggers.md`
- `F/docs/seat-surface-naming.md`
- `F/docs/vocabulary.md`
- `F/docs/why-a-seat-stops.md`
- `F/skills/research-protocol/scripts/debate.js`
- `F/tests/simulator/debate.test.mjs`
- `F/tests/simulator/testdata/prompt-blue-respond.golden`
- `F/tests/simulator/testdata/prompt-judge-petition.golden`
- `F/tests/simulator/testdata/prompt-red-chair.golden`
- `F/tools/devcmd/enginegen/vocabulary.go`
- `F/tools/integration/surface/promptverbs_test.go`
- `F/tools/internal/capture/capture.go`
- `F/tools/internal/capture/capture_test.go`
- `F/tools/internal/capture/lensturn.go`
- `F/tools/internal/capture/lensturn_test.go`
- `F/tools/internal/cli/actrows_test.go`
- `F/tools/internal/cli/adversarial_test.go`
- `F/tools/internal/cli/answers_test.go`
- `F/tools/internal/cli/benchorders_test.go`
- `F/tools/internal/cli/chair/dispatch.go`
- `F/tools/internal/cli/cli_test.go`
- `F/tools/internal/cli/crossseat_test.go`
- `F/tools/internal/cli/dispatch_test.go`
- `F/tools/internal/cli/disputereply_test.go`
- `F/tools/internal/cli/enum_help_test.go`
- `F/tools/internal/cli/epochlimitexit_test.go`
- `F/tools/internal/cli/integration_test.go`
- `F/tools/internal/cli/missingflag_test.go`
- `F/tools/internal/cli/motion/command.go`
- `F/tools/internal/cli/motion/verbs.go`
- `F/tools/internal/cli/motion_test.go`
- `F/tools/internal/cli/motiondispatchedseat_test.go`
- `F/tools/internal/cli/motionfooter_test.go`
- `F/tools/internal/cli/pagestates_test.go`
- `F/tools/internal/cli/payload_test.go`
- `F/tools/internal/cli/refs_test.go`
- `F/tools/internal/cli/requiredmarker_test.go`
- `F/tools/internal/cli/retiresmoke_test.go`
- `F/tools/internal/cli/root.go`
- `F/tools/internal/cli/rulerfiles_test.go`
- `F/tools/internal/cli/seat/help/close.md`
- `F/tools/internal/cli/seat/help/declare.md`
- `F/tools/internal/cli/seat/help/dispatch.md`
- `F/tools/internal/cli/seat/help/mint.md`
- `F/tools/internal/cli/seat/help/regrade-bench.md`
- `F/tools/internal/cli/seat/help/regrade.md`
- `F/tools/internal/cli/seat/help_gen.go`
- `F/tools/internal/cli/seat/refusal.go`
- `F/tools/internal/cli/seat/verbs.go`
- `F/tools/internal/cli/siblingverbs_test.go`
- `F/tools/internal/cli/stalerefusal_test.go`
- `F/tools/internal/cli/standing_test.go`
- `F/tools/internal/cli/stategraph_test.go`
- `F/tools/internal/cli/surface.go`
- `F/tools/internal/cli/verbs_test.go`
- `F/tools/internal/cli/vocabulary_test.go`
- `F/tools/internal/cli/workpage_test.go`
- `F/tools/internal/consistency/consistency.go`
- `F/tools/internal/consistency/consistency_test.go`
- `F/tools/internal/dashboard/judiciary_test.go`
- `F/tools/internal/dashboard/model.go`
- `F/tools/internal/dashboard/render.go`
- `F/tools/internal/dashboard/testdata/render-decided.golden`
- `F/tools/internal/dashboard/testdata/render-live.golden`
- `F/tools/internal/dashboard/testdata/render-terminal.golden`
- `F/tools/internal/difftest/contract_test.go`
- `F/tools/internal/difftest/fuzz_test.go`
- `F/tools/internal/difftest/harness_test.go`
- `F/tools/internal/difftest/scenarios_test.go`
- `F/tools/internal/difftest/testdata/a_repeated_act_names_its_answer.golden`
- `F/tools/internal/difftest/testdata/error_catalogue.golden`
- `F/tools/internal/difftest/testdata/help_contracts.golden`
- `F/tools/internal/difftest/testdata/ids_stay_distinct_across_rounds.golden`
- `F/tools/internal/difftest/testdata/manual_runs_every_help_page_live.golden`
- `F/tools/internal/difftest/testdata/projections_debate_changelog_citations.golden`
- `F/tools/internal/difftest/testdata/role_boundaries_and_help_contracts.golden`
- `F/tools/internal/flags/register.go`
- `F/tools/internal/graph/graph.go`
- `F/tools/internal/graph/graph_test.go`
- `F/tools/internal/record/actjudge.go`
- `F/tools/internal/record/available.go`
- `F/tools/internal/record/available_test.go`
- `F/tools/internal/record/benchgrade.go`
- `F/tools/internal/record/benchorders_test.go`
- `F/tools/internal/record/benchsitting.go`
- `F/tools/internal/record/blockerowner_test.go`
- `F/tools/internal/record/closingowed.go`
- `F/tools/internal/record/closingowed_test.go`
- `F/tools/internal/record/closureview_test.go`
- `F/tools/internal/record/composedsource.go`
- `F/tools/internal/record/dispatch.go`
- `F/tools/internal/record/dispatch_test.go`
- `F/tools/internal/record/dispatchnext.go`
- `F/tools/internal/record/dispatchnext_test.go`
- `F/tools/internal/record/docketruling_test.go`
- `F/tools/internal/record/enumof.go`
- `F/tools/internal/record/enums_test.go`
- `F/tools/internal/record/envelope_test.go`
- `F/tools/internal/record/epochreach_test.go`
- `F/tools/internal/record/gapacts.go`
- `F/tools/internal/record/gapacts_test.go`
- `F/tools/internal/record/grademovement.go`
- `F/tools/internal/record/grademovement_test.go`
- `F/tools/internal/record/grounds.go`
- `F/tools/internal/record/itemkind.go`
- `F/tools/internal/record/keptgrade.go`
- `F/tools/internal/record/lensowes_test.go`
- `F/tools/internal/record/lensturn.go`
- `F/tools/internal/record/m37replay_test.go`
- `F/tools/internal/record/migrate/archive_test.go`
- `F/tools/internal/record/migrate/argumentresidue_test.go`
- `F/tools/internal/record/migrate/benchfiling_test.go`
- `F/tools/internal/record/migrate/closerarchive_test.go`
- `F/tools/internal/record/migrate/entries.go`
- `F/tools/internal/record/migrate/eraentries.go`
- `F/tools/internal/record/migrate/evidenceanchor_test.go`
- `F/tools/internal/record/migrate/registry.go`
- `F/tools/internal/record/migrate/remanddirection_test.go`
- `F/tools/internal/record/migrate/replay.go`
- `F/tools/internal/record/migrate/testdata/prose-rewrites.golden`
- `F/tools/internal/record/motion.go`
- `F/tools/internal/record/motionguards_test.go`
- `F/tools/internal/record/motionqueries_parity_test.go`
- `F/tools/internal/record/motionview.go`
- `F/tools/internal/record/motionview_test.go`
- `F/tools/internal/record/narrowedview_test.go`
- `F/tools/internal/record/passblockers.go`
- `F/tools/internal/record/passblockers_test.go`
- `F/tools/internal/record/planguard/growth.go`
- `F/tools/internal/record/positionheard.go`
- `F/tools/internal/record/queue.go`
- `F/tools/internal/record/reachingretire_test.go`
- `F/tools/internal/record/record.go`
- `F/tools/internal/record/recordpb/correspondence_test.go`
- `F/tools/internal/record/recordpb/keycensus_test.go`
- `F/tools/internal/record/recordpb/record.pb.go`
- `F/tools/internal/record/recordpb/record.proto`
- `F/tools/internal/record/recordsql/decode_gen.go`
- `F/tools/internal/record/recordsql/gapstate.go`
- `F/tools/internal/record/recordsql/gapstate_test.go`
- `F/tools/internal/record/recordsql/readmix_test.go`
- `F/tools/internal/record/recordsql/schema.go`
- `F/tools/internal/record/recordsql/schema_test.go`
- `F/tools/internal/record/recordsql/store_test.go`
- `F/tools/internal/record/recordsql/testdata/schema.sql`
- `F/tools/internal/record/recordsql/viewfamilies_test.go`
- `F/tools/internal/record/recordsql/views.go`
- `F/tools/internal/record/remand.go`
- `F/tools/internal/record/remand_test.go`
- `F/tools/internal/record/remandclosed_test.go`
- `F/tools/internal/record/remandedview_test.go`
- `F/tools/internal/record/replay.go`
- `F/tools/internal/record/replay_test.go`
- `F/tools/internal/record/retirement.go`
- `F/tools/internal/record/retiresmoke_test.go`
- `F/tools/internal/record/rulerfiles_test.go`
- `F/tools/internal/record/sitting.go`
- `F/tools/internal/record/stale.go`
- `F/tools/internal/record/supersede.go`
- `F/tools/internal/record/supersedem38_test.go`
- `F/tools/internal/record/supersession.go`
- `F/tools/internal/record/testdata/plans.golden`
- `F/tools/internal/record/testmain_test.go`
- `F/tools/internal/record/viewjson.go`
- `F/tools/internal/record/viewjson_test.go`
- `F/tools/internal/record/writerace_test.go`
- `F/tools/internal/refusalcheck/refusalcheck_test.go`
- `F/tools/internal/report/assemble.go`
- `F/tools/internal/report/assemble_integration_test.go`
- `F/tools/internal/report/assemble_test.go`
- `F/tools/internal/report/docs.go`
- `F/tools/internal/report/layout.go`
- `F/tools/internal/report/motions.go`
- `F/tools/internal/report/motions_test.go`
- `F/tools/internal/report/seatinput_test.go`
- `F/tools/internal/report/showsthework_test.go`
- `F/tools/internal/report/storedsitting_test.go`
- `F/tools/internal/scorecard/remanded_test.go`
- `F/tools/internal/scorecard/scorecard.go`
- `F/tools/internal/seatprobe/attempted.go`
- `F/tools/internal/seatprobe/attempted_test.go`
- `F/tools/internal/seatprobe/boards.go`
- `F/tools/internal/seatprobe/build.go`
- `F/tools/internal/seatprobe/promptsize_test.go`
- `F/tools/internal/seatprobe/seatprobe.go`
- `F/tools/internal/seatprobe/surfacecoverage_test.go`
- `F/tools/internal/sittingwrite/envelope_test.go`
- `F/tools/internal/terms/definitions_gen.go`
- `F/tools/internal/terms/terms.json`
- `F/tools/internal/verify/verify.go`
- `F/tools/internal/verify/verify_test.go`
- `F/tools/internal/view/view.go`
- `F/tools/internal/view/view_test.go`
- `F/tools/releasegate/fuzz/benchlift_test.go`
- `F/tools/releasegate/fuzz/benchorders_test.go`
- `F/tools/releasegate/fuzz/coverage_test.go`
- `F/tools/releasegate/fuzz/fuzz_test.go`
- `F/tools/releasegate/fuzz/lensturn_test.go`
- `F/tools/releasegate/fuzz/remandroute_test.go`
- `F/tools/releasegate/fuzz/siblingdocket_test.go`
- `F/tools/releasegate/fuzz/supersede_test.go`
- `F/tools/releasegate/fuzz/sweeptally_test.go`
- `F/tools/releasegate/fuzz/termination_test.go`
- `F/tools/releasegate/fuzz/trajectory_test.go`

## A13 the word "finding" outside rows A01–A12 — Part A

Command (audit p2's universe): `git grep -n -i -w -E "findings?|found.?by" origin/main -- <the 117 files of audit-p2/finding-residual.txt>`. **Residual after the change, committed so it needs no scratch list** (plan §V check 5): `git grep -n -i -w -E 'findings?|found.?by' HEAD -- plugins/frank-exchange-of-views scripts/agentgen ':!*/record/migrate/*'` over the whole tree → every hit is an O line of A01–A13 or a generated carrier whose source holds none. Audit p3-A re-read three O lines as seat-facing and they become S: `F/tools/internal/record/standingreading.go:699,711,731` ("That is a reading's finding" → "That is what a reading finds"), `F/tools/internal/cli/lens/reproduce.go:350` ("Both are findings; neither is recorded for you" → "Either is a defect to mint; neither is recorded for you"), `F/tools/internal/cli/seat/help/halt.md:7,9` ("not a finding about the artifact", "treated a boundary as a finding" → "not a defect of the artifact", "treated a boundary as a defect") — a seat must not look for a finding act. S = same sense (rewritten or deleted), D = generated (regenerated from its source), O = another sense (kept).

- `F/tests/simulator/debate.test.mjs:550` S — pins prompt phrase "ids of your findings"
- `F/tests/simulator/debate.test.mjs:561` S — comment: gap not finding about inquiry
- `F/tests/simulator/debate.test.mjs:569` S — asserts "A finding about an inquiry reaches..."
- `F/tests/simulator/debate.test.mjs:570` S — comment: lens finding vs blue blocker
- `F/tests/simulator/debate.test.mjs:574` S — comment: lens filed findings on m21
- `F/tools/internal/capture/sittinglimit.go:17` S — comment: findings a seat files
- `F/tools/internal/cli/assembled.go:25` S — comment: finding as seat-filed act
- `F/tools/internal/cli/bench/command.go:4` S — bench cannot originate findings; comment
- `F/tools/internal/cli/scorecard.go:27` S — findings projection read by scorecard
- `F/tools/internal/cli/scorecard.go:34` S — help text names findings projection
- `F/tools/internal/cli/scorecard.go:59` S — comment: "no findings on the record"
- `F/tools/internal/cli/seat/help.go:8` S — comment names `finding` verb help
- `F/tools/internal/cli/seat/help/dispatch.md:13` S — "another lens's finding" readies a lens
- `F/tools/internal/cli/seat/refusal.go:60` S — PathOf example `finding` command
- `F/tools/internal/cli/seat/verbs.go:300` S — #1874's `copy_captured_at` text: "a `live_source_drift` finding to look for" → "a `live_source_drift` gap to look for"; the 12 agent definitions and difftest's manual golden D
- `F/tools/internal/diagnostics/tokens.go:79` S — example `finding --key F1` verb path
- `F/tools/internal/feov/error.go:28` S — NotFound doc lists finding id
- `F/tools/internal/flags/prose.go:29` S — history: output pasted into a finding
- `F/tools/internal/hookgate/substitution.go:11` S — history: pasted into a finding
- `F/tools/internal/record/about.go:25` S — `lens finding` caller of about
- `F/tools/internal/record/about.go:26` S — finding becomes gap via mint
- `F/tools/internal/record/about.go:60` S — refusing the finding (finding act)
- `F/tools/internal/record/convergence_refusal.go:83` S — chair told material finding minted
- `F/tools/internal/record/convergence_refusal.go:84` S — lens `finding` and `mint` verbs
- `F/tools/internal/record/enumof.go:107` S — AboutKindOf serves finding and gap
- `F/tools/internal/record/enumof.go:108` S — `lens finding` names same about kinds
- `F/tools/internal/record/envelope.go:157` S — lens sitting acts list findings
- `F/tools/internal/record/positionheard.go:253` S — seat-facing: "a defect is a finding and a gap"
- `F/tools/internal/record/queries.go:125` S — findings projection example
- `F/tools/internal/record/recordsql/views.go:424` S — Finding proto / findings view
- `F/tools/internal/record/recordsql/views.go:658` S — finding places a marker anchor
- `F/tools/internal/record/recordtest/recordtest.go:7` S — `finding` verb vs Mint body
- `F/tools/internal/record/redvoice.go:21` S — refuses finding id in prose
- `F/tools/internal/record/redvoice.go:28` S — mint_reason carries run part of finding
- `F/tools/internal/record/supersede.go:62` S — finding keyed on id for supersede
- `F/tools/internal/record/unchecked.go:29` S — edit answering a finding
- `F/tools/internal/record/unchecked.go:33` S — finding id in edit list
- `F/tools/internal/refusalcheck/refusalcheck.go:12` S — example `lens finding requires`
- `F/tools/internal/report/layout.go:166` S — board From: gaps and findings
- `F/tools/internal/report/layout.go:167` S — board Path: lens `finding` opens rows
- `F/tools/internal/report/notes.go:18` S — seat-typed finding text in board
- `F/tools/internal/report/notes.go:21` S — finding words in board document
- `F/tools/internal/reportproj/render.go:23` S — placing verbs include finding
- `F/tools/internal/reportproj/render.go:75` S — marker insertion cite/compute/finding
- `F/tools/internal/reportproj/render.go:117` S — two findings on one sentence
- `F/tools/internal/scorecard/scorecard.go:2` S — reads findings view
- `F/tools/internal/scorecard/scorecard.go:348` S — findings bucketed by lens
- `F/tools/internal/scorecard/scorecard.go:352` S — findings count note
- `F/tools/internal/scorecard/scorecard.go:353` S — findings view rows per area
- `F/tools/internal/scorecard/scorecard.go:356` S — adjudication outcomes per finding
- `F/tools/internal/scorecard/scorecard.go:480` S — findings bucketed by role
- `F/tools/internal/seatprobe/naming.go:65` S — `finding` verb name as English word
- `scripts/agentgen/src/blue-researcher.md:27` S — blue answers every red finding; reword gap
- `scripts/agentgen/src/fragments/log-duty-red.md:1` S — dispute goes in a finding to minter
- `scripts/agentgen/src/lead-judge.md:96` S — stop when findings change character
- `scripts/agentgen/src/lead-judge.md:98` S — findings continuing is normal
- `scripts/agentgen/src/lead-judge.md:99` S — findings becoming internal
- `scripts/agentgen/src/lead-judge.md:102` S — class this run's findings need
- `# SUMMARY: S=57 D=602 O=167; files with S (32):

Files with only O or D hits: 82 (D regenerated; O kept).

## M8 motion text in docs and every help page — Part B (the grade-motion lines in docs are Part B's; a line naming a finding is A13's)

debate.js lines re-read on `3eca8f516` (they moved +3..+17 from `6a5252aa6`).

Command: `git grep -n -i -E "grade motion|motion grade|docket motion|motion docket|docketed|gavel|disposition that decides|closing argument|rules? (it|the gap|a docket)|before the bench|puts? it before" origin/main -- F/docs F/README.md README.md F/skills/research-protocol/references F/skills/research/SKILL.md F/hooks/README.md F/tools/internal/cli F/tools/internal/terms/terms.json scripts/agentgen/src F/skills/research-protocol/scripts/debate.js F/tests/simulator F/tools/internal/seatprobe F/tools/devcmd/seatprobe` (non-test Go). Every help page is in scope; no exclusions.

- `F/docs/duty-docket-preregistration.md:173` S — motion grade appeal count row
- `F/docs/seat-command-triggers.md:45` S — inquest: ruling a docketed gap; docket flow; check vs minter-accepts
- `F/docs/seat-command-triggers.md:46` S — inquest motions: ruling a motion; docket flow; check vs minter-accepts
- `F/docs/seat-command-triggers.md:69` S — lens closing before bench rules
- `F/docs/seat-command-triggers.md:75` S — collapsed into motion grade rule
- `F/docs/seat-command-triggers.md:88` S — blue closing before bench rules
- `F/docs/seat-command-triggers.md:90` S — collapsed into motion grade file
- `F/docs/seat-command-triggers.md:107` S — bench opinion → docket motion disposition
- `F/docs/seat-command-triggers.md:108` S — motion docket file trigger
- `F/docs/seat-command-triggers.md:109` S — motion docket rule: bench disposes
- `F/docs/seat-command-triggers.md:113` S — bench declare row: docket-ruling alternative; check
- `F/docs/seat-command-triggers.md:114` S — bench regrade while docket motion unruled
- `F/docs/seat-command-triggers.md:119` S — "three propose→rule exchanges" header (grade, docket, petition)
- `F/docs/seat-command-triggers.md:123` S — motion grade file row
- `F/docs/seat-command-triggers.md:124` S — motion grade rule row
- `F/docs/seat-command-triggers.md:125` S — grade appeal → motion docket file
- `F/docs/seat-surface-naming.md:251` S — blue surface lists motion grade file/appeal
- `F/docs/seat-surface-naming.md:258` S — blue filed no grade motion
- `F/docs/seat-surface-naming.md:259` S — docket board stages refused grade motion
- `F/docs/why-a-seat-stops.md:47` S — motion grade file / appeal
- `F/skills/research-protocol/references/run-directory.md:73` S — docketed again / remand flow
- `F/skills/research-protocol/scripts/debate.js:347` S — grade motion dimensions enum
- `F/skills/research-protocol/scripts/debate.js:386` S — gap_ids docketed for bench; docket flow; check vs minter-accepts
- `F/skills/research-protocol/scripts/debate.js:395` S — closing on docketed gap; docket flow; check vs minter-accepts
- `F/skills/research-protocol/scripts/debate.js:406` S — docket field: gaps docketed; docket flow; check vs minter-accepts
- `F/skills/research-protocol/scripts/debate.js:475` S — dispute docketed per gap
- `F/skills/research-protocol/scripts/debate.js:494` S — bench disposes docketed gap
- `F/skills/research-protocol/scripts/debate.js:567` S — plan.docket fallback; docket flow; check vs minter-accepts
- `F/skills/research-protocol/scripts/debate.js:727` S — bench rules a docket
- `F/skills/research-protocol/scripts/debate.js:886` S — closing owed per docketed gap; docket flow; check vs minter-accepts
- `F/skills/research-protocol/scripts/debate.js:901` S — closing clause: gap DOCKETED; docket flow; check vs minter-accepts
- `F/skills/research-protocol/scripts/debate.js:912` S — only you and bench regrade
- `F/skills/research-protocol/scripts/debate.js:915` S — grade motions only chair rules
- `F/skills/research-protocol/scripts/debate.js:920` S — chair prompt: RULE THE GRADE MOTIONS
- `F/skills/research-protocol/scripts/debate.js:928` S — blue prompt: grade motion, then docket motion
- `F/skills/research-protocol/scripts/debate.js:930` S — bench prompt: motions docketed for you
- `F/skills/research-protocol/scripts/debate.js:932` S — every docketed gap gets a ruling
- `F/skills/research-protocol/scripts/debate.js:965` S — log "docketed" plan field; docket flow; check vs minter-accepts
- `F/skills/research-protocol/scripts/debate.js:1115` S — chair-ruled grade motion never convenes bench
- `F/skills/research-protocol/scripts/debate.js:1128` S — terminal prompt: keeps "NOTHING CAN BE REMANDED AT A TERMINAL EXIT", gains "on a docketed gap you side with red or blue (filer or assignee on a red seat's proposal), and the tool applies that side's statement" (amendment 4; bench reject; not matched by the row's regex, added by hand)
- `F/skills/research-protocol/scripts/debate.js:1152` S — assembly prompt: gains the sentence that a docket motion still standing at the outcome is ruled first, for a side (red or blue; filer or assignee), nothing remanded (added by hand)
- `F/skills/research-protocol/scripts/debate.js:753,767,1056-1057` O — petition and position prompts' gavel; a stop still seats the bench (unchanged)
- `F/tests/simulator/debate.test.mjs:80` S — docketed plan fixture; docket flow; check vs minter-accepts
- `F/tests/simulator/debate.test.mjs:88` S — bench on its docket fixture; docket flow; check vs minter-accepts
- `F/tests/simulator/debate.test.mjs:230` S — asserts "docketed G-" log; docket flow; check vs minter-accepts
- `F/tests/simulator/debate.test.mjs:281` S — docket motion stands unruled fixture
- `F/tests/simulator/debate.test.mjs:303` S — docket motion unruled fixture
- `F/tests/simulator/debate.test.mjs:453` S — chair gavel (grade) never convenes bench
- `F/tests/simulator/debate.test.mjs:500` S — asserts RULE THE GRADE MOTIONS
- `F/tests/simulator/debate.test.mjs:504` S — docketed gap argued by parties; docket flow; check vs minter-accepts
- `F/tests/simulator/debate.test.mjs:507` S — asserts docketed gap in chairMd; docket flow; check vs minter-accepts
- `F/tests/simulator/debate.test.mjs:509` S — asserts no `motion grade rule` in prompt
- `F/tests/simulator/debate.test.mjs:495` S — asserts the terminal prompt's remand sentence; gains accept-or-reject (with `testdata/prompt-judge-terminal.golden:5`)
- `F/tests/simulator/debate.test.mjs:624` S — pins "a grade motion on the axis"
- `F/tests/simulator/debate.test.mjs:637` S — closing owed, gap DOCKETED; docket flow; check vs minter-accepts
- `F/tests/simulator/debate.test.mjs:638` S — closing owed assert; docket flow; check vs minter-accepts
- `F/tests/simulator/debate.test.mjs:650` S — bench rules docketed gaps test
- `F/tests/simulator/debate.test.mjs:655` S — pins MOTIONS DOCKETED FOR YOU
- `F/tests/simulator/debate.test.mjs:1333` S — closing owed assert; docket flow; check vs minter-accepts
- `F/tests/simulator/debate.test.mjs:1334` S — closing owed assert; docket flow; check vs minter-accepts
- `F/tests/simulator/prompts.test.mjs:67` S — fixture FAILs with docketed gap; docket flow; check vs minter-accepts
- `F/tools/internal/cli/bench/declare.go:24` S — verb is motion docket rule
- `F/tools/internal/cli/bench/declare.go:26` S — docketed gap wants id and fate
- `F/tools/internal/cli/bench/declare.go:43` S — motion docket rule disposes
- `F/tools/internal/cli/chair/dispatch.go:16` S — docketed gap awaiting bench
- `F/tools/internal/cli/chair/dispatch.go:22` S — dispatch files docket motion at impasse
- `F/tools/internal/cli/chair/dispatch.go:23` S — docketed == at impasse
- `F/tools/internal/cli/chair/dispatch.go:87` S — prints "docketed %s for the bench"
- `F/tools/internal/cli/lens/command.go:7` S — bench disposes docketed gap via ruling
- `F/tools/internal/cli/motion/command.go:3` S — motion covers grade dispute
- `F/tools/internal/cli/motion/command.go:9` S — `motion grade file` example
- `F/tools/internal/cli/motion/command.go:19` S — grade motion contract (--dimension)
- `F/tools/internal/cli/motion/command.go:44` S — blue rules own grade motion example
- `F/tools/internal/cli/motion/command.go:76` S — docket Short: bench rules, disposition decides
- `F/tools/internal/cli/motion/command.go:160` S — blue told grade motions unruleable
- `F/tools/internal/cli/motion/verbs.go:24` S — grade subject long: chair rules
- `F/tools/internal/cli/motion/verbs.go:26` S — docket subject long text
- `F/tools/internal/cli/motion/verbs.go:44` S — TERMS: gap before bench in docket motion
- `F/tools/internal/cli/motion/verbs.go:45` S — bench files grade motions
- `F/tools/internal/cli/motion/verbs.go:48` S — HoldsGavel docket branch text
- `F/tools/internal/cli/motion/verbs.go:49` S — disagree → docket motion
- `F/tools/internal/cli/motion/verbs.go:51` S — disagree → motion docket file
- `F/tools/internal/cli/motion/verbs.go:293` S — remand limit on docket rule page; on a docketing past the limit, a side of D's pair
- `F/tools/internal/cli/motion/verbs.go:294` S — second docket motion on closed gap
- `F/tools/internal/cli/motion/verbs.go:325` S — `motion grade rule` refusal comment
- `F/tools/internal/cli/motion/verbs.go:326` S — grade motions belong refusal comment
- `F/tools/internal/cli/motion/verbs.go:355` S — refusal: not a grade-motion ruling
- `F/tools/internal/cli/root.go:81` S — disposition is docket motion ruling
- `F/tools/internal/cli/seat/closure.go:42` S — refusal: motion docket file puts before bench
- `F/tools/internal/cli/seat/help/close.md:9` S — close refused while docket motion unruled
- `F/tools/internal/cli/seat/help/close.md:15` S — demand not held → motion docket file
- `F/tools/internal/cli/seat/help/closing-blue.md:3` S — blue closing on docketed gap; docket flow; check vs minter-accepts
- `F/tools/internal/cli/seat/help/closing-blue.md:7` S — closing per docketing; docket flow; check vs minter-accepts
- `F/tools/internal/cli/seat/help/closing-red.md:3` S — red closing before bench rules; docket flow; check vs minter-accepts
- `F/tools/internal/cli/seat/help/closing-red.md:7` S — closing per docketing; docket flow; check vs minter-accepts
- `F/tools/internal/cli/seat/help/declare.md:7` S — motion docket rule: disposition decides fate
- `F/tools/internal/cli/seat/help/declare.md:11` S — wrong ruling on docket motion
- `F/tools/internal/cli/seat/help/dispatch.md:13` S — accepted grade motion readies lens
- `F/tools/internal/cli/seat/help/dispatch.md:21` S — verb dockets at impasse; docket motion
- `F/tools/internal/cli/seat/help/dispatch.md:25` S — chair gavel (grade); grade/docket motion final sitting
- `F/tools/internal/cli/seat/help/dispatch.md:27` S — grade movement docketed
- `F/tools/internal/cli/seat/help/mint.md:13` S — docket of superseded gap refused
- `F/tools/internal/cli/seat/help/mint.md:23` S — demand not held → motion docket file
- `F/tools/internal/cli/seat/help/regrade-bench.md:3` S — bench regrade while docket motion stands
- `F/tools/internal/cli/seat/help/regrade-bench.md:7` S — docket ruling decides fate
- `F/tools/internal/cli/seat/help/regrade-bench.md:9` S — grade motion on bench axis refused
- `F/tools/internal/cli/seat/help/regrade.md:11` S — bench axis while docket motion stands
- `F/tools/internal/cli/seat/help/reproduce.md:27` S — demand not held → motion docket file
- `F/tools/internal/cli/seat/help/verify.md:22` S — demand not held → motion docket file
- `F/tools/internal/cli/seat/refusal.go:60` S — PathOf example `motion grade file`
- `F/tools/internal/cli/seat/seat.go:630` S — FillsArm docket on motion docket rule
- `F/tools/internal/cli/seat/seat.go:1084` S — closing flag help; docket flow; check vs minter-accepts
- `F/tools/internal/cli/seat/seat.go:1115` S — example `motion docket rule --reason`
- `F/tools/internal/cli/seat/verbs.go:230` S — Closing on docketed gap; docket flow; check vs minter-accepts
- `F/tools/internal/cli/seat/verbs.go:308` S — work written by motion docket rule
- `F/tools/internal/cli/seat/verbs.go:310` S — debate: closing on docketed gap; docket flow; check vs minter-accepts
- `F/tools/internal/cli/seat/verbs.go:343` S — lens list: accepted grade motions, findings
- `F/tools/internal/cli/seat/verbs.go:353` S — docketed gap off list till bench rules
- `F/tools/internal/cli/seat/verbs.go:354` S — remanded / remand_stage
- `F/tools/internal/cli/seat/verbs.go:355` S — accepted grade motion awaits regrade
- `F/tools/internal/cli/seat/verbs.go:356` S — bench list: docket motions
- `F/tools/internal/cli/seat/verbs.go:360` S — estopped: new-grounds docket motion
- `F/tools/internal/cli/seat/verbs.go:364` S — fate written by motion docket rule
- `F/tools/internal/cli/seat/verbs.go:1210` S — chair rules grade motions
- `F/tools/internal/cli/surface.go:140` S — example `motion grade file`
- `F/tools/internal/cli/surface.go:232` S — `motion grade` bare invocation
- `F/tools/internal/seatprobe/attempted.go:110` S — prefix example `motion grade file`
- `F/tools/internal/seatprobe/boards.go:22` S — board staging grade contest + closing
- `F/tools/internal/seatprobe/boards.go:39` S — grade motion probe board
- `F/tools/internal/seatprobe/boards.go:40` S — docket probe board
- `F/tools/internal/seatprobe/boards.go:81` S — grade motion probe board
- `F/tools/internal/seatprobe/boards.go:83` S — grade motion probe board
- `F/tools/internal/seatprobe/boards.go:87` S — docket probe board
- `F/tools/internal/seatprobe/boards.go:223` S — grade motion probe board
- `F/tools/internal/seatprobe/boards.go:231` S — grade motion probe board
- `F/tools/internal/seatprobe/boards.go:320` S — docket probe board
- `F/tools/internal/seatprobe/boards.go:327` S — docket probe board
- `F/tools/internal/seatprobe/boards.go:337` S — docket probe board
- `F/tools/internal/seatprobe/boards.go:339` S — grade motion probe board
- `F/tools/internal/seatprobe/boards.go:358` S — docket probe board
- `F/tools/internal/seatprobe/boards.go:438` S — grade motion probe board
- `F/tools/internal/seatprobe/boards.go:467` S — grade motion probe board
- `F/tools/internal/seatprobe/boards.go:479` S — docket probe board
- `F/tools/internal/seatprobe/boards.go:480` S — docket probe board
- `F/tools/internal/seatprobe/boards.go:493` S — grade motion probe board
- `F/tools/internal/seatprobe/boards.go:494` S — docket probe board
- `F/tools/internal/seatprobe/boards.go:564` S — grade motion probe board
- `F/tools/internal/seatprobe/boards.go:589` S — grade motion probe board
- `F/tools/internal/seatprobe/boards.go:610` S — grade motion probe board
- `F/tools/internal/seatprobe/boards.go:633` S — docket probe board
- `F/tools/internal/seatprobe/boards.go:639` S — docket probe board
- `F/tools/internal/seatprobe/boards.go:661` S — docket probe board
- `F/tools/internal/seatprobe/boards.go:674` S — docket probe board
- `F/tools/internal/seatprobe/boards.go:677` S — docket probe board
- `F/tools/internal/seatprobe/boards.go:680` S — grade motion probe board
- `F/tools/internal/seatprobe/boards.go:740` S — docket probe board
- `F/tools/internal/seatprobe/boards.go:803` S — docket probe board
- `F/tools/internal/seatprobe/boards.go:830` S — grade motion probe board
- `F/tools/internal/seatprobe/boards.go:841` S — docket probe board
- `F/tools/internal/seatprobe/boards.go:894` S — grade motion probe board
- `F/tools/internal/seatprobe/boards.go:895` S — grade motion probe board
- `F/tools/internal/seatprobe/boards.go:909` S — docket probe board
- `F/tools/internal/seatprobe/boards.go:910` S — grade motion probe board
- `F/tools/internal/seatprobe/boards.go:912` S — STALE: "lens files FINDINGS; chair turns them into gaps"
- `F/tools/internal/seatprobe/boards.go:913` S — STALE: "lens files FINDINGS; chair turns them into gaps"
- `F/tools/internal/seatprobe/boards.go:918` S — grade motion probe board
- `F/tools/internal/seatprobe/build.go:79` S — closings after docket motion
- `F/tools/internal/seatprobe/build.go:143` S — docket motion puts gap before bench
- `F/tools/internal/seatprobe/build.go:223` S — closings while docket motion stands
- `F/tools/internal/seatprobe/build.go:249` S — error: no docket motion for closings
- `F/tools/internal/seatprobe/production.go:46` S — bench (docketed) production path
- `F/tools/internal/seatprobe/seatprobe.go:218` S — `motion grade file` unmeetable comment
- `F/tools/internal/seatprobe/seatprobe.go:392` S — rejected grade motion owes no regrade
- `F/tools/internal/terms/terms.json:866` S — disposition: bench ruling on docketed gap
- `F/tools/internal/terms/terms.json:940` S — term "grade motion"
- `F/tools/internal/terms/terms.json:941` S — grade motion definition
- `F/tools/internal/terms/terms.json:985` S — new grounds via docket motion
- `scripts/agentgen/src/blue-researcher.md:29` S — docketed gap: closing until bench rules
- `scripts/agentgen/src/fragments/lens-audit-and-mint.md:8` S — demand not held goes before bench
- `scripts/agentgen/src/fragments/lens-audit-and-mint.md:19` S — blue may file grade motion
- `scripts/agentgen/src/fragments/lens-duties.md:5` S — accepted grade motion; docketed closing
- `scripts/agentgen/src/fragments/lens-duties.md:9` S — docketed gap is the bench's
- `scripts/agentgen/src/lead-judge.md:42` S — rule docketed gap on motion id
- `scripts/agentgen/src/lead-judge.md:43` S — bench regrade before ruling
- `scripts/agentgen/src/lead-judge.md:44` S — docket motion reconsideration on new grounds
- `scripts/agentgen/src/lead-judge.md:56` S — same mechanism as grade motion
- `scripts/agentgen/src/lead-judge.md:57` S — gavel holder differs (grade vs petition)
- `scripts/agentgen/src/lead-judge.md:82` S — docketed at impasse; lens may not close
- `scripts/agentgen/src/lead-judge.md:86` S — grade movement docket motion
- `scripts/agentgen/src/red-chair.md:3` S — description: rules blue's grade motions
- `scripts/agentgen/src/red-chair.md:13` S — chair rules each grade motion
- `scripts/agentgen/src/red-chair.md:15` S — gap put before bench; docketed closings
- `scripts/agentgen/src/red-chair.md:17` S — RULE BLUE'S GRADE MOTIONS

O hits (kept): 51; D hits (regenerated): 24.

## M9 readers of a gap's assignee and of a mint's acting seat — Part B (F-B1, owner from the chain, F-B3, L5, assignee)

**Amendment 7 (plan §III.B15).** The bench writes no `Mint` and no accept writes one, so `actingSeat`
and L5 are deleted: every **ACTOR** line below reads the Mint event's seat again (tag ACTOR → WRITER in
phase B7), and T-B14 becomes `TestAssigneeAndWriterReadersOfASuccessor` with no bench successor. The
ASSIGNEE lines are unchanged; the one case where writer and assignee differ is a lens's successor of
another lens's closed gap, which is F-CS.

Re-read on `3eca8f516` (line numbers are its; `c6f061897` moves them by a few). The assignee is
`gap_state.assignee`, the chain's origin, written by `gap_refresh` through `assigneeSQL` (plan
§III.B4a); the gap view's `assignee` (renamed from `minted_by`, M14) reads it. Three questions (ruled
2026-11-14): **ASSIGNEE** — whose gap (who owes its closing, hears and accepts proposals, may regrade
or close it, whose audit, work list or news it is) → the column; **ACTOR** — which seat acted (whose
debt a mint discharges, whose reading stands behind it, who authored its text or a prescription,
whether a seat minted since) → `actingSeat(e, assignee)`: the event's seat, the assignee where the
bench wrote the Mint (L5); **WRITER** — keyed on the event's own sitting or retry key → the event's
seat. Two sweeps, every hit tagged:
(1) the gap view's `minted_by`: `git grep -n -E 'minted_by|mintedBy|MintedBy' 3eca8f516 -- F/tools/internal ':!*_test.go' ':!*_gen.go' ':!*.pb.go'`, the 51 lines / 14 files of `~/.claude/scratch/1695/audit-p2/r2/mintedby-readers.txt`;
(2) a Mint event's seat: `BodyAs[*recordpb.Mint]`, `GetMint()` and `case *recordpb.Mint` sites (`~/.claude/scratch/1695/p3/mint-sites.txt` plus the type switches), each read for whether it takes `e.GetSeatId()`.
Tags: **SOURCE**; **ASSIGNEE** (reads the column; change named); **ACTOR** (switches to `actingSeat`); **WRITER** (keeps the event's seat); **#1828** (deleted with the mint cap); **NONE** (reads no seat).

Sweep (1), the gap view's `minted_by` (the assignee from Part B on):
[SOURCE] F/record/recordsql/views.go:477: `e."seat_id" AS "minted_by"` → `s."assignee" AS "assignee"` (M14)
[ASSIGNEE] F/cli/seat/verbs.go:291,337 (and :306,352 text): "the one seat that may close it" — whose gap; text in M12
[ASSIGNEE] F/record/answersretired.go:34: the answers-retired item goes to the gap's assignee ("which you closed")
[ACTOR] F/record/composedsource.go:77-78: "its own mints the auditor does not owe" is authorship → `actingSeat` of the Mint event of the loop at `:82`; was the view's column [CHANGE]
[ASSIGNEE] F/record/dispatch.go:357-374: engage the gap's lens on its audit turn
[ASSIGNEE] F/record/dispatch.go:430: closingPartiesOf(cast, g.mintedBy)
[ASSIGNEE] F/record/dispatch.go:493: plan text naming whose audit
[ASSIGNEE] F/record/dispatch.go:1517-1532: openGaps, source of the three above
[#1828] F/record/mintbudget.go:446,469,504-509
[ASSIGNEE] F/record/onegapperclass.go:180-203: the refusal names the gap's assignee, to whom a docket motion on it goes
[ASSIGNEE] F/record/passblockers.go:701-732: a material gap's and a remand's blocker Owner, who clears it
[ASSIGNEE] F/record/record.go:1797 requireOriginator; :1864 the closing write check
[ASSIGNEE] F/record/retirement.go:168 lensOwesOf, :215 gapNewsOf: news owed to the gap's lens
[ASSIGNEE] F/record/standingreading.go:1046,1054: assembly failure text "open material gap G by X" — whose gap holds the assembly
[ASSIGNEE] F/record/supersession.go:87: requireSupersedable, an open ancestor is its assignee's to supersede
[#1828] F/record/supersession.go:130: supersedesOwnOpen
[ASSIGNEE] F/record/viewjson.go:174-177,440,468,890-895,1037-1039,1206,1241,1404-1410: board, gap and work JSON; YoursToClose, YourTurn
[ASSIGNEE] F/sittingwrite/delivery.go:457,466: sitting brief, whose gap

Sweep (2), a Mint event's seat:
[SOURCE] F/record/available.go:258-268: mintedBy(evs), a fold of the first Mint's seat → one read of `gap_state ("gap_id", "assignee")` (`mintedBy` → `assigneeOf`) or the caller's gap rows
[ASSIGNEE] F/record/available.go:149: availableOf via mintedBy (its regradesAfforded use leaves with B6)
[ASSIGNEE] F/record/available.go:460: remandItemsOf, remand item for the assignee
[ASSIGNEE] F/record/viewjson.go:1598: regradesAfforded — deleted with the grade chain (B6)
[ASSIGNEE] F/record/lensturn.go:50: earlyLensDispatchesOf via mintedBy, whether a lens dispatch was its turn
[ASSIGNEE] F/record/blueanswered.go:56: ANSWERED news to the assignee
[ASSIGNEE] F/record/closingowed.go:117-118: closingDutiesOf, the red closing owed → mintedBy [CHANGE]
[ASSIGNEE] F/record/impasse.go:287-291: exchangeFold, the lens party of a gap's exchanges → mintedBy [CHANGE]
[ASSIGNEE] F/record/stale.go:88-89: staleOf feeds gapNewsOf the gap's lens, whose news counts → mintedBy [CHANGE]
[ASSIGNEE] F/record/recordsql/views.go:734: demandReachSQL `seat`, whose re-check meets the demand → assigneeSQL rendered for the ancestor mint [CHANGE]
[ACTOR] F/record/estoppel.go:132,139: Estoppel.PrescribedBy via minterOf, "the author of the prescription" → `actingSeat` [CHANGE]; on a bench-written successor the assignee
[WRITER] F/record/retirement.go:298-301: lensStates, productive sittings are this seat's own fresh mints in its own sittings — keyed on the event's sitting, which on a bench accept is the bench's (L19)
[ACTOR] F/record/answersretired.go:68: supersededBySeatAfter, whether this seat itself minted a successor after the retire → `actingSeat` [CHANGE]
[ACTOR] F/record/failedcheck.go:80-81: failedChecksOf, a gap the check's own seat minted answers its check → `actingSeat` [CHANGE]: a bench-accepted successor about the anchor answers the assignee's check
[ACTOR] F/record/standingreading.go:745-767: readingBehind (three arms), the reading the mint was checked on → `actingSeat` [CHANGE] (the p2 revision keyed it on the bench on a bench accept; per L5 it is the assignee)
[ACTOR] F/record/standingreading.go:858-862: mintRowsOf, a mint discharges its acting seat's reading debt → `actingSeat` [CHANGE]
[ACTOR] F/record/inquirycheck.go:122-133: the gap label "G-… (seat)" names who acted about the inquiry → `actingSeat` [CHANGE]
[WRITER] F/record/replay.go:413: ExistingMintByKey, a retry is keyed on the writer
[#1828] F/record/mintbudget.go:372,487
[NONE] F/record/gapacts.go:54,133,241 (supersedes, about, place), bluesittings.go:81, benchgrade.go:86, gapstates.go:76, grademovement.go:96, keptgrade.go:41, inquiry.go:152, successor.go:40, supersession.go:39, blueanswered.go:173, standingreading.go:321, viewjson.go:630,1751, actjudge.go:41, record.go:923,958, consistency.go:156 (its own fold of the record, no seat), cli/lens/mint.go:109

Counts: ASSIGNEE changes 4 (closingowed, impasse, stale, demandReachSQL) plus the rename of every ASSIGNEE line (M14); ACTOR changes 9 reads in 6 files (estoppel.go:139, answersretired.go:68, failedcheck.go:81, standingreading.go:754,761,767,862, inquirycheck.go:133, composedsource.go:77 — `actingSeat`, one function); WRITER 2 unchanged; SOURCE 2.

Tests: T-B14 `TestAssigneeAndActingSeatReadersOfASuccessor` has one row per line above, each run on a lineage successor S1 (lens B of A's closed gap) and a bench successor S2 (A's G, proposal accepted by the bench): ASSIGNEE rows return A on both; ACTOR rows return B on S1 and A on S2; WRITER rows return B and the bench. Its failed-check case covers failedcheck.go's three write-side readers and the bench-accepted answer of A's check. T-B16 holds the ASSIGNEE rows outside the assignee's area. The schema and shapes goldens gain the column (no index: `assigneeSQL`'s reads are key reads).
Text: F/cli/seat/verbs.go:306,352 ("Each gap's `minted_by` is the seat that minted it, the one seat that may close it") → "Each gap's `assignee` is the seat that minted its chain's first gap — the gap itself, or the gap its successor supersedes, whoever wrote the successor — and the one seat that may close it" (with M8; the rest of the owner text is M12).
Fuzz: oracle `assignee_drift` (§III.B9) recomputes the assignee from the stream, independent of the column; `bench_accepts_by_act`'s supersede cell, `successor_chains`, `lineage_successors_reowned` read the stored column.

Other-sense hits (unchanged):
F/record/keptgrade.go:55, benchgrade.go:107, grademovement.go:107, impasse.go:294: seat of Regrade, not Mint
F/record/gapacts.go:90: seat of each act on a gap (Mint contributes only its supersedes), not assignee derivation
F/record/remand.go:165: seat of the closing act, not the minter
F/consistency/consistency.go:138: register/sitting seat, no Mint
F/view/view.go:280, record/viewjson.go:1766: Finding's seat (finding area), not Mint (Part A deletes them)
No hits in capture, gapnews, ledger, scorecard, dashboard, report, sitting.go for the Mint-event-seat pattern (they read gap view rows).

## M10 what the gap verbs add beyond the append — Part B (§III.B3)

F/cli/lens/mint.go:44: seat.Reason resolves --reason before retry check (prose arg → mint_reason via fill:273) | no (body field) | no
F/cli/lens/mint.go:50 (fill 259-325): builds body; FixBasis hard-set "proposed" (311); class/check-kind word parse | validate requires check_kind/likelihood/impact, unclassified reason | no
F/cli/lens/mint.go:71: record.ResolveAbout resolves --about-kind/--about against the record (reference exists) | partly: validate requireAboutAnchor (record.go:971), judgeAct requireAboutAnchorLive | no
F/cli/lens/mint.go:79: refuses --quote AND --about together | validate requireMintSubject (record.go:965) asks exactly-one too | no
F/cli/lens/mint.go:90-97: verifiedFix guard (mint.go:414) — --new checked as a pair blue.PlanAccept would apply at the gap's placed anchor; MUTATES body FixNew + FixBasis="verified" (mint.go:450-451) | NO (verb-only guard) | no
F/cli/lens/mint.go:100: p.FoundBy = --found-by finding ids (F- refs) | validate requireFindings (record.go:1019) — dies with the finding event | no
F/cli/lens/mint.go:105-117: crash-retry idempotency — ExistingMintByKey(run, s.SeatID, --key) + seat.RepeatsKeyed; returns stored gap id | no (replay.go:411 is a read; record validate has no key check for mint) | YES (s.SeatID)
F/cli/lens/mint.go:123: record.SupersessionAdmitted(run, s.SeatID, p) — minter-only supersede + successor carries demand, pre-asked before placement | yes: validate requireSupersedable (record.go:1027), judgeAct requireSuccessorCarriesDemand (actjudge.go:64) | YES
F/cli/lens/mint.go:129: gap id minted by verb (record.NewID) | no (validate only requires present) | no
F/cli/lens/mint.go:158-160: estoppel guard requireNotEstopped (mint.go:542) via record.EstoppelConflict on Landing.Family | NO (verb-only guard) | no (guard); message at 170 compares est.PrescribedBy == s.SeatID
F/cli/lens/mint.go:163: seat.AppendPlaced → record.AppendPlaced: adds a companion recordpb.Anchor{Id: gapID, Location: quote} body in the same transaction (record.go:658-663) | Anchor body is validated as its own event (shape: location xor sentence_of; Part B adds `stands_on`, M15); placement itself is NOT | no
F/cli/seat/verbs.go:1385-1388: AppendPlaced prepends Places guard — anchortext.Attach(current report, id, loc) refuses a quote the report does not hold once | NO (verb-only guard) | no
F/cli/lens/mint.go:164-198: on estoppel refusal, writes a separate recordpb.Log{type ESTOPPEL, source TOOL} event after rollback, returns ToolLogged conflict | no (verb effect) | identity only (record.Append(s.Identity()))
F/cli/lens/mint.go:206-212: reportvoice.Advised soft tells on problem/fix/new returned as advice (unambiguous ones refused by judgeAct refuseMintReportVoice) | advice: no; refusal half: judgeAct | no
F/cli/lens/mint.go:229-230: seat.Supplies gap_id, class_material (stamped at write path) | class_material stamped in write path | no
(judgeAct-only Mint rules, not in the verb: requireOwnStandingReading, requireOneGapPerClassAndSubject, requireSupersededOffDocket, requireMintWithinBudget — all take the writer's seatID from identity)
F/cli/seat/closure.go:29-57: ClosurePayload — --as word must be a disposition AND recordpb.Closes (refuses `remanded`, pointing to `motion docket file`) | DB CHECK on closure_class subset; refusal text verb-only | no
F/cli/seat/closure.go:20,24: --id / --superseded-by flags carry WithCheck(record.GapExists) (flag-parse existence check) | validate requireGap/requireOpenGap also | no
F/cli/lens/close.go:43-50: verification triple (AnchorSeat/Tool/Target) + Anchors ids from flags | validate checks anchored triple + anchor ids exist | no (--verified-by is typed, not the caller)
F/cli/lens/close.go:67-70: requireEvidenceAnswered(EvidenceDemandOf(run,id)) — live-answer demand pre-check, unreadable record refuses | yes: judgeAct requireDemandMet (actjudge.go:88) | no
F/cli/lens/close.go:71: plain record.Append — NO AppendFresh/Fresh guard, NO AppendPlaced | — | identity only
(validate-side Close rules not in the verb: requireOriginator (record.go:1101, reads gap.minted_by vs writer seat), requireOpenGap (1176); judgeAct: requireNoStandingDocket (actjudge.go:81), requireFailedCheckAnswered (94))
F/cli/seat/regrade.go:32-36: seat.Reason → body.Basis | basis is a required field (schema), not verb logic | no
F/cli/seat/regrade.go:39-44: AppendFresh(s, gapID, ...) — passes record.Fresh(s.SeatID, gapID) guard (stale.go:56) refusing if another seat acted on the gap after this seat's read point | NO (guard, live only) | YES (s.SeatID)
F/cli/seat/verbs.go:1407-1418: on StaleError, AppendFresh writes a companion recordpb.Stale{GapId, Through} event, then returns the conflict | no (verb effect) | identity
F/cli/seat/regrade.go:48: --id WithCheck(record.GapExists) + Require | validate requireOpenGap (record.go:1291) | no
(validate-side Regrade rules not in the verb: requireOriginator (record.go:1284, gap.minted_by, skipped for bench), requireRegradeWithinBenchLock (1295); grades stamped by recordsql.stampRegradeResult)
- Only two guards are verb-only and absent from validate AND judgeAct: mint's verifiedFix (which also mutates fix_basis/fix_new) and mint's estoppel; plus the Places placement guard shared by every placing verb. Moving the mint into another code path (e.g. a motion's accepted act) must carry these three or lose them.
- Regrade has a Fresh guard and Stale companion; close has neither — a close is not refused when the gap moved under the seat.
- Caller seat id is read by the verb itself only in: mint retry lookup (105), SupersessionAdmitted (123), estoppel message (170), regrade's Fresh guard (39). Every record rule takes seatID from the write identity.

## M11 readers of a docket ruling and of bench closure — Part B (audit gap 2)

Command: `~/.claude/scratch/1695/audit-p1/docket-rule-readers.txt` (90 lines) and
`bench-closed-readers.txt` (70 lines, 23 files), both on `6a5252aa6`; the rows marked (p3) re-read on
`3eca8f516`. A bench close accept is today's granted ruling with the stored disposition, and a bench
supersede accept is an `accepted` ruling the gap state's bench CTE reads as the bench's close of G
(plan §III.B4; no `superseded` disposition, L12 withdrawn), so the bench-closure readers that take the
word from the ruling's `disposition` read the fate `superseded` through one function, `benchClosureWord`. The new rows are a ruling with `accepted: true` and no disposition (the assignee, any act;
the bench, a regrade) and one with `accepted: false` (the bench's reject); from phase B2a (amendment 4) `sided_with` replaces `accepted` — the filer's side for `true`, the other for `false`. Each reader, and what it
does with those rows:

| Reader | Today | Change |
|---|---|---|
| `record.go:1750` `rulingWord` (→ `Motion.Ruled()`, `motionsAt`, Compute) | `Word(disposition)` | returns the closing word where set, else the side (`for red` … `for assignee`) or the grade answer (`grades kept` \| `grades restored`, amendment 5) — **the one switch**; without it a side-only or grade ruling reads unruled everywhere |
| `grademovement.go:90` (p3) | restarts the movement window from `m.ruled` | **re-keyed (amendment 3, L41, audit p7 gap 4):** restarts only at a `GRADE_MOVEMENT` docketing's ruling on G, either answer, (and, per axis, after the latest bench `Regrade` and the latest regrade proposal's ruling naming it, L41′); no other ruling and no assignee's accept restarts it |
| `recordsql/views.go:601` `motion_answers` | `COALESCE(grade, petition, rd.disposition)` | `rd.disposition` → `CASE rd.accepted WHEN 1 THEN 'accepted' WHEN 0 THEN 'rejected' ELSE rd.disposition END` (view; epoch) |
| `motion.go:544-574` `requireUnanswered`/`RequireUnruledMotion` | first ruling wins, in the write transaction | unchanged: the minter's accept is a ruling, so a second answer of either kind is refused here (gap 5) |
| `recordsql/gapstate.go:189-193` "bench" CTE; `:198-219` closure, unruled filing | joins `enum_disposition` where `closes` | **gains one arm**: an `accepted` ruling on a motion with a `motion_supersede` row, written by a seat other than G's stored `assignee` (the bench) — the reader fix (ruled 2026-11-13); NULL disposition otherwise joins nothing; any ruling clears `unruled_docket_*` |
| `bench-closed-readers.txt`: `closed_by_bench`, `bench_close_event`, `bench_closed_last` in views.go (8), viewjson.go (26), replay.go (6), gapstate.go (6), consistency.go (6), scorecard.go, remand.go, record.go, gapstates.go, declarations.go, answersretired.go; 12 agent OUTPUT schemas | read the closing bench ruling | unchanged: a bench close accept IS a closing ruling; a minter close accept is a `Close` (red's close, as it is) |
| `supersession.go:36` | `Closes(disposition)` | unchanged: `Closes(UNSPECIFIED)` is false (`facets.go:229`) |
| `replay.go:139-148,244-259` `BenchClosure`, `reason()` | closing rulings | also the bench's accepted supersede ruling (role `bench`), fate `superseded` through `benchClosureWord` |
| `viewjson.go:568,642-645` `lastBenchClosure`; the estopped row's `fate` | closing rulings | read the gap view's `bench_close_event` (which now names the accepted supersede ruling); word through `benchClosureWord` |
| `remand.go:206-235` `requireRemandWithinLimit`; `sitting.go:526`; `cli/motion/verbs.go:293`; `passblockers.go:933` `requireBenchMotionsRuled` (p3) | name the closing dispositions, or `--as <verdict>` | name D's pair (`--for-red \| --for-blue`; `--for-filer \| --for-assignee` on a red seat's proposal on another red seat's gap; `--keep \| --restore` at grade movement) and the remand where one is available (plan §III.B14, amendment 5) |
| `remand.go:259` `ClosingDispositions()`; `enumvalue.go:225-245` `evsOf` → the generated `--as` lists (p3) | every `closes` word | unchanged: no disposition is added |
| `benchsitting.go:26-108` `benchHeld`/`docketHeld` (p3) | the bench sitting's blocking motions | also the write's admission (`requireBenchHeld`); `docketHeld` holds the standing docketings, and a `PROPOSAL` docketing's own motion — no proposal is filed on a docketed gap (plan §III.B13, L38; amends B2's "before the dispatch that convened the sitting" and amendment 1's "before the bench sitting's opening") |
| `capture.go:1453-1463` precedent harvest | every docket ruling | the bench's rulings only; word from `rulingWord` |
| `consistency.go:222-250` | non-closing word → `carriedCount++` | an accepted ruling is skipped, except the bench's accepted supersede ruling, which its own fold reads as a bench closure (role `bench`) and compares with the view |
| `report/assemble.go:1402-1414` LEAD; `view/view.go:731-740`; `viewjson.go:2108-2116` `Lead` | every docket ruling as the bench's opinion | the bench's rulings only (event seat's role `bench`); word from `rulingWord` |
| `record.go:1366-1420` DocketRuling write checks; `record.proto:1981-2010` CHECKs | disposition required | (amendment 5) exactly one of `sided_with`, `moved_grades`, `disposition = 'remanded'`; `reopens_on`, `final` only with a closing word, `direction` only with `remanded`; `sided_with` in D's pair, and the assignee's accept the filer's side; `requireBenchOpinion`; the side against a proposal from the bench only; every bench ruling names its docketing (`DocketRuling.docketing_id`), and `--accept`/`--reject` on a docketing holding no proposal is refused by `requireSubjectMatches`' docket arm, naming the closing words, the bench's regrade and remand (plan §III.B13; replaces B2's no-act refusal) |
| `remand.go:162-170`, `views.go:281-313` remand view | remanded rulings / closing acts | unchanged |
| `motion.go:352,400-410`, `replay.go:286-300`, `consistency.go:113`, `lensturn.go:121`, `remand.go:196`, `verify.go:213`, `sitting.go:728`, `gapacts.go:86`, `migrate/*` | read the filing's gap | **re-keyed (audit p7 gaps 1–2):** a reader of a docket RULING resolves its gap through `docketing_id` (`DocketRulingGap`, which replaces `replay.go:281-305`'s `DocketGapByMotion`; appendix M18 tags every hit); a reader of a FILING keeps the filing's gap; `lensturn.go:121` reads `Docketing` rows (plan §III.B13); the grade arm leaves (M1) |
| `graph.go:209`, `verify.go:129`, `view.go:339` | comments | unchanged |

**Readers of "a docket motion is ruled", read for meaning (audit p4 gap 2; `94b43bac0`).** Command:
every `Ruled()` / `m.ruled` / `Closes(` / `.Ruling` reader (`~/.claude/scratch/1695/audit-p4/m11.txt`,
108 lines; `closes-readers.txt`, 48). Three answers now leave an open G open — the assignee's accept
of a regrade (a lens's `MotionRule`), the bench's regrade accept, the bench's reject — and the
assignee's accept of any act is a lens's ruling, so "ruled" no longer means "the bench ruled" nor, on
an open gap, "remanded". Each reader is tagged by the one predicate it needs: **bench-only** (the
ruling event's seat has role `bench`, `seatclass.RoleOfSeat(m.RulingBy)`, as the LEAD fix keys it),
**any answer** (`Ruled()`), **remand-only** (`m.Ruling == remanded`, as `remandFloorsOf` already
reads); a reader whose item names the ruling takes the ruling's own word (`m.Ruling`), never
"remand" for another. The column says what the three open-gap answers are to it; T-B10 drives each.

| Reader | Tag | Assignee's regrade accept / bench regrade accept / bench reject (amendment 4: the bench's for-blue / for-red answer to blue's regrade; each `accepted` / `rejected` word below is the side's, `for blue` / `for red`) |
|---|---|---|
| `retirement.go:233-245` `gapNewsOf` (comment "A RULED DOCKET MOTION ON AN OPEN GAP IS A REMAND"; its item is what carries a material gap back into the debate) | bench-only, in its own word | no item (the assignee's own act) / "the bench's regrade of G on M" / "the bench's rejection of the <act> M proposed on G" — each readies G's assignee, as a remand does ("the bench's remand of G in M", unchanged); the comment becomes "a bench ruling that leaves a material gap open is news, in its own word" |
| `scorecard.go:363-430` `docketRulings` (`remanded_share`, the opinion scan) | bench-only | not counted / counted (not remanded) / counted (not remanded); the comment "the write requires a principle on every docket ruling" holds again, for the bench's (F-B2) |
| `dashboard/model.go:483` `j.Rulings` (beside the judge's sittings) | bench-only | not counted / `accepted` / `rejected` |
| `stale.go:118-119` (a fact since the seat's read) | any answer, worded by ruler | "<assignee> accepted motion M's regrade of G" / "the bench ruled G accepted in M" / "the bench ruled G rejected in M" — the bench named only for the bench's |
| `inquirycheck.go:150` (`moved`: a cause that readies the inquiry owner) | any answer but a reject | moved / moved / not moved (G stands as it was) |
| `verify.go:468` `GapsWithDisposition` | bench-only | not counted / counted / counted |
| `graph.go:85` `motionsRuled`, `dispositions` | any answer; `dispositions` bench-only | ruled, no disposition / ruled, disposition / ruled, disposition |
| `blueanswered.go:285-300` `remandFloorsOf` | remand-only (already) | no floor / no floor / no floor |
| `grademovement.go:86-90` (window restarts at the latest ruling) | any answer | restarts (the accepted regrade, after the ruling, is party movement) / restarts / restarts — R3 |
| `closingowed.go:142-165` (closing duties end at the ruling) | any answer | ends / ends / ends |
| `available.go:111`, `composedsource.go:71`, `passblockers.go:793`, `benchgrade.go:234`, `gapstate.go:219` (a docket standing unruled) | any answer | lifts / lifts / lifts |
| `dispatch.go:815` (ruled since the chair's last sitting) | any answer | counted / counted / counted |
| `record/motionview.go:79` (`show motions` JSON `ruled`, `ruling`, `ruling_by`) | any answer, in its own word | `accepted` by <assignee> / `accepted` by <bench> / `rejected` by <bench> — the ruler named, so the assignee's accept reads as the assignee's (audit p5 NOTE) |
| `cli/motion/verbs.go:507` (`ruled.Human`: "motion M ruled <word>") | any answer, in its own word | "ruled accepted" / "ruled accepted" / "ruled rejected" — printed to the ruler that just wrote it |
| `record/relief.go:34` (a granted petition's relief) | other subject (petition) | — |
| `report/motions.go:65` (the docket document's "ruled <word> by <seat>") | any answer | "ruled accepted by <assignee>" / "ruled accepted by <bench>" / "ruled rejected by <bench>" |
| `capture.go:1457`, `consistency.go:211-222`, `viewjson.go:641,2068`, `view/view.go:726`, `report/assemble.go:1398` | bench-only (table above) | as the rows above state |
| `record.go:1730-1745` `rulingWord`, `motion.go:399-413,545-567`, `views.go:601` | definitional (table above) | `accepted` / `accepted` / `rejected` |
| `positionheard.go:237`, `positionowed.go:206`, `petitionorder.go`, `report/assemble.go:1476` | other subjects (position, petition) | — |
| `migrate/*`, `record.proto` options, `seatprobe`, `rulerfiles.go`, `refs.go:472`, `roles.go:31`, `supersede.go:127`, `supersession.go:35` | archived records (no accepted ruling), gavel text, or `Closes(disposition)` (false on NULL, table above) | — |

## M12 carriers that teach a successor's assignee or `--supersedes` — Part B (F-B3)

Census: `git grep -n -i 'supersed' origin/main -- <F>/tools/internal/cli/seat/help <F>/tools/internal/cli/lens/*.go <F>/tools/internal/cli/seat/*.go scripts/agentgen/src <F>/skills <F>/debate.js <F>/tools/internal/terms/terms.json <F>/tools/internal/flags ':!*_test.go' ':!*_gen.go'` (63 hits / 23 files on `1852ce534`, `~/.claude/scratch/1695/supersede-a/carriers.txt`), plus `minted_by`/"yours for its whole life"/"you minted" in the lens prompt and fragments. Each hit read for what it MEANS: an S line teaches who a successor or a lineage mint belongs to, or what the writer owes after it; an O line is another sense (an inquiry's or certification's "superseded", `--superseded-by`'s regression, the successor's closing of an open ancestor, flag plumbing). After F-B3 every S line says: a gap is its chain's first minter's for its whole life, whoever wrote it; a successor of your own gap is yours; a successor you mint of another lens's closed gap is that lens's to audit and close, your act done at the mint; a successor names one gap. Prompts name acts, never the flag.

S `F/skills/research-protocol/scripts/debate.js:912` lens prompt: "A gap you minted is yours for its whole life … a successor you mint names the gaps it supersedes" → "A gap is its chain's first minter's for its whole life: yours when you minted it or it succeeds one of yours, whoever wrote it; a successor you mint of another lens's closed gap is that lens's to audit and close, and your act is done when you mint it" (authored region; the prompt golden changes)
S `F/tools/internal/cli/seat/help/mint.md:13` "Naming a closed gap carries lineage and closes nothing." → "… closes nothing, and your gap goes to that gap's minter — its chain's origin — who audits and closes it; you have done your act. A successor names one gap: to fold several, supersede one and close the rest." (M13)
S `F/tools/internal/cli/lens/mint.go:222` `--supersedes` flag text: "comma-separated ancestor ids …" → "the one gap this gap replaces; the new gap is that gap's minter's — its chain's origin — whoever mints it" (M13)
S `F/tools/internal/cli/lens/mint.go:173` estoppel refusal "argue it there, or mint with --supersedes %s so the lineage is explicit" → adds "a successor of it is that gap's assignee's to close" (generic: `est.PrescribedBy` is the prescription's acting seat, M9 ACTOR); comment `:140,:156` reworded to match
S `F/tools/internal/record/recordpb/record.proto:1814` `LOG_TYPE_ESTOPPEL` means text "or mint with --supersedes so the lineage is explicit" → "… so the lineage is explicit; that successor is the prescribing gap's assignee's"
S `F/tools/internal/record/supersession.go:93` `requireSupersedable` refusal (Part A rewrites its finding road); its bench arm's refusal (§III.B4); the second-ancestor refusal (§III.B11, at the flag)
S `F/tools/internal/cli/lens/mint.go:118-120` comment "ONLY AN OPEN GAP'S MINTER SUPERSEDES IT …" → adds that a closed gap's successor is its chain origin's
S `scripts/agentgen/src/fragments/lens-audit-and-mint.md:8` "the seat whose check it carries — you, for a gap you minted about an anchor" and "A successor you mint superseding it" → "the gap's minter — its chain's first, whoever minted it"; regenerated into every lens agent definition (`agentgen -check`)
S `F/tools/internal/terms/terms.json:1130` `superseded`: "when its own lens mints a successor naming it" → "when its minter, or the tool on a proposal its minter or the bench accepts, mints a successor naming it while it is open"; then `vocabgen`, `vocabdoc`
S `F/tools/internal/terms/terms.json:1148` "superseded (a closed gap)" resolution → adds "the successor is the closed gap's minter's, whoever mints it"
S `F/tools/internal/cli/seat/verbs.go:306,352` `minted_by` text → `assignee` (M9 Text, M14)
O `F/tools/internal/cli/seat/help/engage.md:7` blue's engagement on a successor (names no assignee)
O `F/tools/internal/cli/seat/help/close.md:13,17`, `carry.md:11` a successor closes its open ancestor
O `F/tools/internal/cli/seat/help/regrade.md:11`, `regrade-bench.md:9` bench lock on successors' grades
O `F/tools/internal/cli/seat/closure.go:18,24,54`, `flags/names.go:214-215,333,419-420`, `flags/csv.go`, `flags/shapes.go` `--superseded-by` (a regression) and flag plumbing
O `F/tools/internal/cli/seat/answer.go:18,51` record's act superseders (another sense)
O `F/tools/internal/cli/seat/verbs.go:311,321,359,360` bench statements, retires, engagement, estoppel list
O `F/tools/internal/cli/lens/mint.go:35,99,127,159,215,500,528,538-547` flag plumbing and the estoppel guard's own logic
O `F/skills/*`, `scripts/agentgen/src/skills/*`, `lead-judge.md:129` inquiry and certification "superseded"; "grade stability along supersedes chains"
O `F/skills/research-protocol/scripts/debate.js:205,539-547,1011-1044` blue's successor engagement

Fuzz/seatprobe: `lineage_successors_reowned`, `second_supersedes_refused` (§III.B9); a seatprobe board for a lens minting a successor of another lens's closed gap asserts the lens is told the gap goes to its origin.

## M13 readers of `Mint.supersedes` — Part B (one ancestor per mint, §III.B11)

Command (on `3eca8f516`): `git grep -n -E 'GetSupersedes\(\)|\.Supersedes\b|mint_supersedes|supersedes_count|SupersedesCount|"supersedes"|--supersedes|Supersedes:' origin/main -- F scripts ':!*.pb.go' ':!*.golden' ':!*/testdata/*' ':!plans'` → 222 lines (`~/.claude/scratch/1695/p3/supersedes-readers.txt`); 221 on `94b43bac0` (audit p4: `supersession.go` one fewer; new fixtures `mintroster_test.go`, `runstat_test.go`, each S with its subject). Each read for what it does with the list: L = loops or `Contains` over it (becomes one id, `""` none); T = the table or its index/trigger/marker (becomes the `mint."supersedes"` column); C = `supersedes_count` (becomes `"supersedes" IS NULL`); X = text naming the flag or the list; O = another sense.

- `F/tools/internal/record/recordpb/record.proto:993` T — `repeated string supersedes = 20` → 20 reserved, `optional string supersedes = 27`; its comment ("ALWAYS PRESENT, even empty … an empty list") → "absent: lineage none"
- `F/tools/internal/record/recordsql/schema.go:288` T — index `mint_supersedes_value` → `mint ("supersedes", "event_id")`
- `F/tools/internal/record/recordsql/decode_gen.go:869,874` T — regenerated (protogen)
- `F/tools/internal/record/recordsql/demandlineage.go:16,52,60` T — trigger on `mint` `WHEN NEW."supersedes" IS NOT NULL`; first-naming test on the column; diamond comment goes
- `F/tools/internal/record/recordsql/gapstate.go:47,129,211,290` T, C — marker folds into the mint's; `supersede_event` reads the column; `supersedes_count` → `supersedes`
- `F/tools/internal/record/recordsql/views.go:439,484,541` C — the gap view's column and `fresh` ask `IS NULL`
- `F/tools/internal/record/recordsql/store.go:760` O — comment on a list row's id; reworded with the table gone
- `F/tools/internal/record/grounds.go:120` T — `EXISTS (… mint_supersedes …)` → `m."supersedes" = ?2`
- `F/tools/internal/record/planguard/rule.go:22` O — example plan text naming the table; re-pick an example
- `F/tools/internal/record/retirement.go:345,366` C, L — `IS NULL`; `g.Supersedes == ""`
- `F/tools/internal/record/viewjson.go:179,432,468,631,1180,1268,1435` L, T — JSON `supersedes` one id (`""` none, always written); `listValuesByEvent(q, "mint_supersedes")` → the mint row's column; `Contains` → `==`; `inherits:` + the id
- `F/tools/internal/record/supersession.go:40,78,92-93,104,111,125` L, X — one ancestor; refusal texts say "names %s" of one gap
- `F/tools/internal/record/successor.go:41`, `gapacts.go:55`, `bluesittings.go:82`, `lensturn.go:55`, `impasse.go:294`, `inquirycheck.go:123`, `benchgrade.go:96,211,219,225`, `demand.go:173,200,241`, `record.go:1022,1024`, `onegapperclass.go:47,73`, `answersretired.go:68` L — one id
- `F/tools/internal/record/estoppel.go:32`, `failedcheck.go:239,244`, `record.go:1190`, `refs.go:17`, `required.go:13` X — text naming `--supersedes`; read: each names one gap already (unchanged) or is a comment (reworded where it says "gaps")
- `F/tools/internal/consistency/consistency.go:164,171` L — its own fold keeps one ancestor
- `F/tools/internal/dashboard/model.go:499`, `graph/graph.go:232,280`, `verify/verify.go:230`, `view/view.go:304,457,467` L — one id
- `F/tools/internal/cli/lens/mint.go:99,127,140,156,173,215,222,500,528` L, X — single-value flag refusing a comma and a repeat (§III.B11); flag text; `contains` at `:500` → `==`
- `F/tools/internal/flags/csv.go:8,14`, `shapes.go:19,77`, `names.go:214` X — `--supersedes` leaves the CSV type for a single-value type that refuses a second; comments reworded
- `F/tools/internal/terms/terms.json:1130` (and `definitions_gen.go:59`, regenerated), `F/docs/vocabulary.md:652` (vocabdoc) X — "naming it": one gap (with M12's owner text)
- `F/tools/internal/record/recordpb/record.proto:1814` (and `annotations_gen.go:338`) X — estoppel means text (M12)
- `F/agents/*.md` (7 red lenses × 3, `blue-researcher.md`, `blue-synthesizer.md`, `lead-judge.md`, `red-chair.md`) X — regenerated from agentgen sources and terms (`agentgen -check`)
- Tests (113 lines / 54 files, listed in the command's output) — fixtures that build a `Supersedes` list build one id; `store_test.go:92` and `planguard_test.go:36,83` read the table and move to the column; `TestTheStoredGapStateIsTheRules` (`gapstate_test.go:154,166,208`) reads `supersedes` for `supersedes_count`; `mintbudget_test.go` goes with #1828
- `F/tools/internal/record/migrate/*` — Translation 1 (§III.B7): the archived list's `ord` 0 into field 27

Fuzz: drives that name ancestors name one; `second_supersedes_refused` (forced seed, both forms).

## M14 the rename `minted_by` → `assignee` — Part B (ruled 2026-11-14)

Command: `git grep -n -E 'minted_by|mintedBy|MintedBy' origin/feov/delete-mint-cap -- plugins scripts ':!*/testdata/walls/*'` → 169 lines (`~/.claude/scratch/1695/p4/mintedby-all.txt`): Go 80 lines / 19 files, tests 22 / 11, goldens 19 / 3 (`difftest/testdata/manual_runs_every_help_page_live.golden` 12, `record/testdata/plans.golden` 6, `recordsql/testdata/schema.sql` 1), generated agent definitions 48 / 12. Free name: `git grep -l -w -i assignee` → one HTML fixture (`fetchcache/testdata/walls/acl.html`, a URL parameter, other sense). Every hit is S (the same column, renamed) except where noted; "minter" and the Mint event's seat keep meaning who minted.

- `F/tools/internal/record/recordsql/views.go` (1) — the gap view's column: `s."assignee" AS "assignee"` (M9 SOURCE)
- `F/tools/internal/record/viewjson.go` (23) — JSON field `assignee` on board, gap and work rows; `YoursToClose`/`YourTurn` read it
- `F/tools/internal/record/dispatch.go` (10), `record.go` (8), `closingowed.go` (6), `passblockers.go` (5), `available.go` (4, `mintedBy` → `assigneeOf`), `onegapperclass.go` (4, `heldGap.mintedBy`), `supersession.go` (4), `retirement.go` (2), `standingreading.go` (2), `composedsource.go` (2, ACTOR: switches to `actingSeat`, M9), `answersretired.go`, `blueanswered.go`, `gapacts.go`, `lensturn.go`, `stale.go` (1 each) — renamed
- `F/tools/internal/cli/seat/verbs.go` (2) — help text: "`assignee` and `yours_to_close`"; the `minted_by` sentence (M9 Text)
- `F/tools/internal/sittingwrite/delivery.go` (2) — sitting brief field
- tests (22 lines / 11 files) — renamed with their subject; goldens regenerated (`UPDATE_GOLDENS=1`, every changed line read)
- `F/agents/*.md` (48 lines / 12 files) — generated (`agentgen`; the help manual and output schemas carry the field)

## M15 the reconsideration grounds, estoppel's history fold and the work list's `estopped`, replaced by the written `Tombstone` rows — Part B (plan §III.B12)

Commands (`c6f061897`): (1) `git grep -n -i -E 'grounds|reconsider' <rev> -- F scripts/agentgen ':!*.pb.go'` → 339 lines (`~/.claude/scratch/1695/p4/grounds-all.txt`); (2) `git grep -n -E 'estoppingPrescriptions|verbatimPrescriptions|EstoppelConflict|minEstoppelOverlap|TestARewordEnds|\.estopped\b' <rev> -- F`; (3) `git grep -n -i 'estopped' <rev> -- F/tools/internal F/skills scripts/agentgen ':!*_test.go' ':!*.golden'`; (4) on `94b43bac0`, the readers of the facts the ending acts carry and of the work list (`gapMarkers` among them, read and unchanged by tombstones): `git grep -n -E 'GetReopened\(\)|ReopenedAnchors|BaredAnchors|blue_edit_reopened|gapMarkers|EstoppedJSON|benchClosedLast' 94b43bac0 -- F/tools ':!*_test.go'`. S = changes, D = generated, O = another sense.

The grounds (deleted):
- `F/tools/internal/record/grounds.go` (32) S — deleted whole
- `F/tools/internal/record/actjudge.go:21,96` S — `requireNewGrounds` → `requireNotEstoppedOn`; comment
- `F/tools/internal/record/motion.go:357-363` S — `Fields["grounds"]`, `["grounds_quote"]` go
- `F/tools/internal/report/motions.go:58-63` S — "on new grounds" lines go
- `F/tools/internal/cli/motion/verbs.go:26,145-170,227-231,289,590-592` S — the flags, the Places guard on `--quote`, `groundsQuoteRefusal`, the help sentences; `:289` ("A MOTION THAT NAMES NEW GROUNDS…") → the bench's rulings are final, for the ruled gap and the successor a supersede accept left standing, while the ruled gap's tombstone stands
- `F/tools/internal/flags/names.go:259-261,332,359` S — `Grounds` goes; `:50` O ("a ruling's grounds", a prose argument)
- `F/tools/internal/record/recordpb/record.proto:1955-1962` S — fields 2, 3 reserved; `:1959` comment goes; `:1983` O (the closing ruling's `reopens_on` why-text, L15)
- `F/tools/internal/record/recordsql/schema.go:276` S — comment names `requireNewGrounds`; `testdata/schema.sql` D
- `F/tools/internal/record/viewjson.go:1010` S — comment
- `F/tools/internal/record/migrate/remap.go:119,135` S — the two field entries go
- `F/tools/internal/cli/seat/verbs.go:344` S — the `estopped` paragraph: "a docket motion on the gap on NEW GROUNDS" → the ruled gap's tombstone, the gaps it bars (the ruled one, and the successor a supersede accept left standing), and what ends it (blue's reword of the gap's sentence, not a cut; for a gap about a citation or computation its reword or retire; for an inquiry its next fate; a section gap's, and a cut gap's, stands for the run)
- `F/tools/internal/terms/terms.json:984-985` S — `new grounds` deleted; `estopped` added; `definitions_gen.go` D; `docs/vocabulary.md:55,575-577` D (`vocabdoc`)
- `F/docs/seat-command-triggers.md:108` S — "on NEW GROUNDS" → a proposed act, refused while a ruling's tombstone bars the gap (the ruled gap, or the successor its supersede accept left standing)
- `scripts/agentgen/src/lead-judge.md:44` S — "A DOCKET MOTION ON A GAP YOU RULED CLOSED IS A RECONSIDERATION … NEW GROUNDS" → what you rule, but a remand, is final for the ruled gap and the gap your ruling leaves standing while its tombstone stands; blue rewording the gap's sentence (a cut ends nothing), rewording or retiring the anchor it is about, or recording the next fate of the inquiry it is about ends it
- `F/tools/internal/record/benchgrade.go:204` S — deleted with the grade machinery (M1)
- tests: `record/benchorders_test.go` (22), `cli/benchorders_test.go` (15), `releasegate/fuzz/benchorders_test.go` (17), `fuzz/sweeptally_test.go` (10), `fuzz/fuzz_test.go` (8), `record/closureview_test.go` (4), `cli/answers_test.go`, `recordpb/requiredfields_test.go`, `record/terminalverdict_test.go`, `record/outcomeverdict_test.go`, `fuzz/remandroute_test.go`, `fuzz/releasegate_test.go` S — reconsideration cases become tombstone cases (T-B19) or go
- goldens: `difftest/testdata/manual_runs_every_help_page_live.golden` (14), `difftest/testdata/docket_ruling_requires_each_unconditional_field.golden` (2), `migrate/testdata/prose-rewrites.golden` (5), `record/testdata/plans.golden` (2) D — regenerated, every line read
- `F/agents/*.md` D — regenerated
- O: `debate.js:292,338` and every `prompt-*.golden` line ("your grounds beside the acts", an argument); `docs/citations-and-bibliography.md:220-221` ("grounds an … gap", a verb); `docs/seat-surface-naming.md:444`; `skills/adversarial-audit/SKILL.md:49` and its agentgen source `:29`; `fragments/lens-audit-and-mint.md:3,7,10`; `fetchcache/pagerender.go:98`; `setup/setup.go:94`; `record/retirement.go:20`; `record/verdict.go:53`; `report/assemble.go:1671`; `seatprobe/boards.go:917`; `record/record.go:1397` (the `reopens_on` refusal, L15)

Estoppel's history fold (replaced by the written estoppel `Tombstone` rows):
- `F/tools/internal/record/estoppel.go:35-103` S — the F12 comment, `estoppingPrescriptions`, `verbatimPrescriptions`' `reworded` arm go; `appliedVerbatim` stays (`DeclineStatsOf`); `:117` `EstoppelConflict` reads the standing estoppel rows (`tombstone_standing`, through `tombstone_kind_gap`) and their gaps' `fix_new`; `:139` `minterOf` → `actingSeat` (M9)
- `F/tools/internal/record/standingreading.go:194,225-227,461,957` S — `estopped` (the chain's last edit applied red's text verbatim) → a standing estoppel row on the reading's gap
- `F/tools/internal/cli/lens/mint.go:145,157,159-198,537-548` S — the guard's comments and `requireNotEstopped` read the rows; the TOOL `Log` stays
- `F/tools/internal/cli/blue/edit.go:179-214` S — the estoppel comments; the edit's tombstone rows are written by the write path (`writeTombstones`), not the verb; `:283-305` `BaredAnchors` O (unchanged: a cut ends nothing, plan §III.B12)
- tests: `cli/blueedit_test.go:580-582` `TestARewordEndsEstoppelAndAReapplyRestoresIt` and `record/estoppel_test.go:60-107,199-207` `TestARewordEndsEstoppel`, `record/locatorclass_test.go:27,128-134` S — over the rows
- O (command (4), 54 lines, `~/.claude/scratch/1695/p5/m15-4.txt`): `record/anchorchecks.go:262`, `record/gapacts.go:221`, `viewjson.go:2151` (other readers of `reopened`, unchanged); `migrate/gapplace.go:166-168` (migrate's own `reopened`, unchanged; a cut, archived or live, ends nothing); `runtest/reopened.go` (a test helper over `blue_edit_reopened`); `cli/lens/mint.go:446-447` (`BaredAnchors`' two existing sets, unchanged); `answersretired.go:49`, `gapstates.go:116`, `viewjson.go:572-646,1060,1219` (`benchClosedLast` as closure state, unchanged); `record/testdata/plans.golden` D (the new statements' plans)

The work list's `estopped` (rebuilt from the standing ruling rows):
- `F/tools/internal/record/viewjson.go:714-719,947-1000,1093,1407-1425` S — one row per gap a standing ruling `Tombstone` row protects (G, and S after a supersede accept; was `benchClosedLast`); the row gains the ruling's motion and event, its word (a closing word, an accepted act, `rejected`) and what ends it; the board, gap and work rows gain `ruling_tombstone` and `estoppel_tombstone`, read from `tombstone_standing`
- `F/tools/internal/sittingwrite/delivery.go:36,61-78,155,222,370,472-476` S — index lines from those rows
- `F/tools/internal/cli/seat/verbs.go:326,344` S — the page's paragraph and comment
- `F/tools/internal/record/declarations.go:28`, `sitting.go:628`, `enums.go:166` O — the gap-view rows and fate vocabulary, unchanged
- `F/tools/internal/cli/enginestopped.go`, `cli/root.go:263`, `record/outcomeclosed.go:23` O (`engine-stopped`); `migrate/entries.go:131-184`, `record.proto:1835-1838`, `payload-keys.txt:37` O (`estopped_by`, a reserved log key)

The `Tombstone` body (new; every writer and reader listed). Command (5), on `32f152f65`, the carriers every event body has, read off an existing tool-written body, tests included (audit p5 NOTE): `git grep -n -E 'EnvelopeAdmitted|ENVELOPE_ADMITTED|envelope_admitted' 32f152f65 -- F scripts ':!*.pb.go' ':!*.golden' ':!*testdata*'` → `record.proto` (enum value, oneof arm, message), `body_gen.go`, `decode_gen.go`, `required_gen.go`, `annotations_gen.go`, `views.go:710` `NotActs`, `migrate/remap.go:134` (the field lists), its own write site and guard (`sittingwrite/write.go`, `failedsitting.go`, `migrate/admission.go`) and `hooks/README.md:84` (prose on the hook that writes it; O for `Tombstone`, whose writer is no hook); and 12 test lines: the four per-event-type lists a new body joins (S, rows below) and 8 tests of the body itself (O: `migrate/admission_test.go:32`, `record/plansweep_test.go:44`, `record/sittingstop_test.go:19,23`, `sittingwrite/admission_test.go:29,39`, `fuzz/dispatchreturn_test.go:24,182`); and `git grep -n -i -w tombstone 32f152f65 -- F scripts/agentgen ':!*/vecembed/vocab.txt'` → 0. Command (6), the readers of the act's own event after a write that now writes rows after it: `git grep -n -i -E 'MAX\("id"\)' 32f152f65 -- F/tools/internal ':!*_test.go' ':!*testdata*'` → 11 lines, listed below as S or O. Every row below is a site the change creates or a per-body carrier command (5) found:
- `F/tools/internal/record/recordpb/record.proto` S — `EVENT_TYPE_TOMBSTONE` [NEW] with its means ("the tool's record that a bench ruling or a verbatim edit placed a tombstone on a gap, or that a later act ended it"); the `Event` oneof arm `Tombstone tombstone`; `message Tombstone` (`gap_id` required, `kind` required, `motion_id`, `ends`) and `enum TombstoneKind` with means texts; CHECKs: `motion_id` set iff `kind` is RULING and `ends` unset
- `F/tools/internal/record/recordpb/record.pb.go`, `body_gen.go`, `record.proto.sha256`; `recordsql/decode_gen.go` (the `tombstone` table reader); `record/required_gen.go`; `recordpb/annotations_gen.go`; `testdata/schema.sql` (`tombstone`); `record.EventSchema` D — `protogen`, `schemagen`, `cmdtreegen`, `vocabgen`; `testdata/schema.epoch.json` restamped (`UPDATE_GOLDENS=1`)
- `F/tools/internal/record/recordsql/schema.go` (`QueryIndexDDL`) S — `tombstone_kind_gap` on `tombstone ("kind", "gap_id", "event_id")`; planguard plans each statement below; `testdata/plans.golden` D
- `F/tools/internal/record/recordsql/views.go` S — view `tombstone_standing` (a placing row with no later row on its (`gap_id`, `kind`)); `NotActs` (`:703-712`) gains `EVENT_TYPE_TOMBSTONE` (the tool's bookkeeping, not a seat's act: `seat_metrics.acts`, `sittingRecordedNothing` and `counterpartyOf` count none); `internal/record/testdata/shapes.golden` read (`PLANGUARD_ACCEPT_VIEW_COST=1` for the new view, the reason in the commit)
- `F/tools/internal/record/migrate/remap.go:112-140` S — `Tombstone.gap_id`, `Tombstone.docketing_id` in `idFields`; `TestEveryStringFieldIsInOneList` fails either missing
- `F/tools/internal/record/migrate/entries.go` S — `"tombstone": {Translate: toNoEvent}` (`:71`'s mechanism): a source row is written as no event, the replay's write path re-deriving it from its act (plan §III.B12 Migrate (2), audit p5 gap 1)
- test-side per-event-type lists (command (5)) S — `releasegate/fuzz/fuzz_test.go:8367` (`verbsWithEvents`: `tombstone` beside `envelope_admitted`) and `:8539` (`reportExemptions`, the exhaustive `dialecticProseKey` gate: "written by the tool with the act that places or ends it, carrying no prose"); `cli/vocabulary_test.go:155` (`noVerbOfItsOwn`: "written by the write path with the act that owes it; no verb records one"); `recordpb/correspondence_test.go:90` (the per-type comment list); `TestEveryAppendableEventTypeIsInTheCoverageGate`, `TestEveryRequiredFieldIsMarkedInTheHelp`, `TestEveryEventTypeHasABodyAndViceVersa` and the release fuzz fail any one missing
- `F/tools/internal/record/tombstone.go` [NEW] S — `writeTombstones` (the one write site), `tombstonesOwed` over `tombstoneArms` (the placing and ending arms of §III.B12's tables), the standing reads by gap and by kind, S's read through its ancestor (`motion_supersede."gap_id"`)
- `F/tools/internal/record/record.go` (`writeIn`, `:793` on; `appendAct`, `:712`) S — `writeIn` calls `writeTombstones` right after `InsertTx` (`:824`), for every writer (`appendAct`, `RegisterSeat` `:290`, `dispatchNext`'s closure, the replay), and returns the id `InsertTx` assigned, which `appendAct` returns beside the event; validate's body switch gains the `*recordpb.Tombstone` arm, the body's shape checks only (no owed-set refusal, plan §III.B12)
- `F/tools/internal/record/replaywriter.go:84-101` S — `Replay.Write` returns the act's id beside its event; it writes through `appendAct` → `writeIn`, so the replay writes the rows with no other change here; its test-side callers, `record/replaywriter_test.go:57-234` (12 `b.Write` calls) and `record/supersededemand_test.go:93`, take the second return (every `writeIn`/`appendAct`/`Replay.Write` caller, tests included: `git grep -n -E '\bwriteIn\(|\bappendAct\(|BeginReplay\(|batch\.Write\(' 32f152f65 -- F/tools`, plus `b.Write(` in `replaywriter_test.go`; `migrate/order.go:48,83` and `migrate/proserefs_internal_test.go:26` begin a replay and write through `replay`, O); `dispatchnext.go:69` O (writes through `writeIn`, its filings owing no row)
- `F/tools/internal/record/recordtest/recordtest.go:136` (`Seed`, `recordsql.Insert`) and `F/tools/internal/record/runtest/overtaken.go:42` (`Overtaken`, `InsertTx`) O — test helpers that bypass `writeIn` and write no `Tombstone` row; no tombstone test builds its record through them (plan §III.B12)
- `F/tools/internal/record/writeview.go:25`, `dispatchnext.go:63-71` — B2 built `byDispatch`; phase B2a deletes it (plan §III.B13, appendix M16): no exemption remains
- `F/tools/internal/record/onegapperclass.go:59-185` S — the held set gains the gaps of the class with a standing ruling row, by their own subject; exempt: the ruling's own successor (F-J), the owed failed check; the refusal's estopped case
- `F/tools/internal/record/actjudge.go:96` S — `requireNotEstoppedOn` (G's standing ruling row, and S's through its docketing's ruling's proposal), with `requireExchangedOn`, for every writer of a `Motion`
- `F/tools/internal/record/viewjson.go` (the rows of the work-list section above) S — `ruling_tombstone`, `estoppel_tombstone` on board, gap and work rows, read from `tombstone_standing`
- `F/tools/internal/cli/blue/edit.go`, `cli/blue/inquiry.go:141`, `cli/blue/retire.go` (help only) S — each page says what its act ends (`edit`: a reword, not a cut; `retire`: a citation's or computation's retire); no verb, flag or help page writes or names the body
- `F/tools/internal/terms/terms.json` S — `tombstone` added ("a bench ruling's or a verbatim application's bar on a gap, a row the tool writes with the act, standing until the act that ends it writes its end row"); `definitions_gen.go`, `docs/vocabulary.md` D
- `F/agents/*.md` D — regenerated (the gap row's new fields in the output schemas)
- tests [NEW]: `record/tombstone_test.go` (`TestTheStoredTombstonesAreTheRule`, `TestEveryTombstoneArmIsNeeded`), T-B19 (`internal/cli`), T-B20 (`migrate`), T-B21 (`internal/cli`, through `runMigrate`); `recordpb/requiredfields_test.go` S (the new body's required fields)
- `F/tools/internal/record/migrate/replay.go:573-581` S — `newID[old.ID]` is the id `batch.Write` returns for the act (`recordpb.Event` carries none; `writeIn` hands `InsertTx`'s up through `appendAct` and `Replay.Write`), not `SELECT MAX("id")`, which after a placing or ending act is its last row; `res.Out` counts the rows (command (6)); `:368` O (the dropped act's "last event written": a row is part of the record the next act saw); `replay.go:229` `newID` otherwise unchanged; `:526,537,585,700` (every `batch.Write` call) take the second return, `:537`'s alone using it; a source `tombstone` reaches no `batch.Write` (`entries.go`'s `toNoEvent`), and `:382-403`'s `!landed` arms refuse any reference naming one
- command (6) O: `record.go:294` (a register's sitting; a register places nothing), `views.go:537` and `convergence_refusal.go:57` (the board at the latest event: a row is an event of the record), `queries.go:183` (epochs), `reportjson.go:57` (filtered to report types), `recordsql/eventcache.go:126`, `recordsql/store.go:801,863` (the next id; a comment), `planguard/rule.go:56` (a comment)
- `F/tools/internal/record/recordsql/gapstate.go`, `gapstate_test.go` O — unchanged by tombstones (no column, `gap_refresh` rule or `gapMarkers()` row)

## M16 the dispatch-filed motion, and every duty keyed on a docket motion that a docketing now carries — Part B (amendment 2026-11-20, plan §III.B13)

Command, on `a488cad41` (the branch's B2 head; `~/.claude/scratch/1695/m16-files.txt`, 68 files, 315
hits):

```
git grep -c -E 'unruled_docket|unruledFiled|unruledMotion|unruledDocket|standingClosingDuties|closingDutiesOf|benchSatFor|docketHeld|requireNoStandingDocket|requireSupersededOffDocket|docketedThisSitting|routeDocket|routeRedocket|routeGradeMovement|routeBench|unruledMotionBlockers|BlockerUnruledMotion|requireBenchMotionsRuled|requireRemandOfOpenGap|requireRemandWithinLimit|byDispatch|ToFile|restoring|noActRefusal|requireRegradeWithinBenchLock' a488cad41 -- plugins/frank-exchange-of-views/tools ':!*/record/migrate/*'
```

plus the word `docketing` (`git grep -n -i -w -E 'docketings?' a488cad41 -- plugins/frank-exchange-of-views
scripts/agentgen ':!*_test.go' ':!*.golden'`, 31 files) and the new body's per-body carriers (M15
command (5), re-run for `docketing`). Each reader is tagged by the fact it needs: **PROPOSAL-STANDS**
(a party's motion stands unanswered), **DOCKETING-STANDS** (a docketing stands unruled), **DELETE**
(B2's dispatch filing), **TEXT** (a carrier that says a docketing is a motion's filing), **O** (other
sense). Plan §III.B13's duty table is the change; T-B23 a subtest per DOCKETING-STANDS row.

- `F/tools/internal/record/recordpb/record.proto` S — `Docketing`, `DocketingCause`, `EVENT_TYPE_DOCKETING`, `Event` arm 66; `DocketRuling.docketing_id = 10`; `MotionRule.motion_id` not required (its `why` and comment rewritten); `Motion`'s CHECK (act iff docket); `Tombstone` field 3 → `docketing_id`; `MOTION_SUBJECT_DOCKET` means, the `DocketMotion` and `DocketRuling` comments (TEXT); generated `record.pb.go`, `body_gen.go`, `decode_gen.go`, `required_gen.go`, `annotations_gen.go` D
- `F/tools/internal/record/recordsql/schema.go` S — `docketing_gap` in `QueryIndexDDL`; `recordsql/testdata/schema.sql` (8) D; `schema.epoch.json` D (`main` + 3)
- `F/tools/internal/record/recordsql/gapstate.go` (4) DOCKETING-STANDS — `standing_docketing`, `standing_docketing_at` columns and their `gap_refresh` arm (reads `docketing_gap`); `unruled_docket_*` kept, PROPOSAL-STANDS; `gapMarkers()` docketing arm; `gapstate_test.go` (10) S
- `F/tools/internal/record/recordsql/views.go` (3) S — the gap view's two new columns; `NotActs` gains `EVENT_TYPE_DOCKETING`; `motion_answers` unchanged; the `remand` view keyed on the docketing's gap
- `F/tools/internal/record/dispatch.go` (45) DELETE + DOCKETING-STANDS — `ToFile`, `basis`, `restore` → `ToDocket`; `toBench` writes no motion; `routeOf`'s callers; `routeProposal` [NEW]; `benchSatFor` keyed on the docketing; `docketedThisSitting`; why-lines ("docket motion M stands unruled" → "docketing D"); `:1313`'s `len(plan.ToFile) > 0` → `ToDocket`
- `F/tools/internal/record/dispatchnext.go` (5) DELETE — `writeAs`' `byDispatch`, the `Motion` loop and `plan.restore` (`:63-71,114-135`) → the `Docketing` loop; `:117-121`'s comment (TEXT)
- `F/tools/internal/record/writeview.go` (2) DELETE — `byDispatch`
- `F/tools/internal/record/proposal.go` (7) DELETE — `:20-21` comment, `:195` `!w.byDispatch`, `noActRefusal` (`:245-279`); F5's refusal counts proposals only (unchanged code, PROPOSAL-STANDS)
- `F/tools/internal/record/grademovement.go` (3) DOCKETING-STANDS — `restoring` (`:145-160`) kept, the bench's `--restore` at grade movement (amendments 2–5, M17, L35); the window restart (`:80-90`) at the latest `GRADE_MOVEMENT` docketing's ruling, either answer, only, per axis after the latest bench `Regrade` and the latest regrade proposal's ruling naming it (L41, L41′); `pending` reads the standing docketing
- `F/tools/internal/record/remand.go` (16) DOCKETING-STANDS — `routeOf` (`:349`): the standing docketing → `routeBench`, a standing proposal → `routeProposal`; `requireRemandOfOpenGap`, `requireRemandWithinLimit`, `pastRemandAnswers` keyed on the docketing's gap; texts name the docketing's three answers (TEXT)
- `F/tools/internal/record/benchsitting.go` (11) DOCKETING-STANDS — `docketHeld`: the standing docketings on the convening rows' gaps, and a `PROPOSAL` docketing's own `motion_id` (no proposal is filed on a docketed gap, L38); `benchHeld`'s terminal/assemble arms: every docketing and proposal
- `F/tools/internal/record/closingowed.go` (8) DOCKETING-STANDS + TEXT — `closingDutiesOf` over docketings (span from the row; for `PROPOSAL`, from the proposal's filing, L30); `standingClosingDuties`; the file comment's "the dispatch files the docket motion"
- `F/tools/internal/record/passblockers.go` (9) DOCKETING-STANDS — `BlockerDocketed` [NEW `docketed`] per standing docketing, owner bench; `unruled_motion` for a proposal in its window; `requireBenchMotionsRuled` (`:919-933`) over docketings and proposals, its text (TEXT)
- `F/tools/internal/record/record.go` (8) DOCKETING-STANDS — `requireNoStandingDocket` (`:1751`) refuses a close or supersede on a standing proposal or docketing and names the docketing, and gains two arms (L38, §III.B14): a proposal's filing while a docketing stands, and blue's edit or retire over the span a standing docketing's red closing replaces; validate's `Docketing` arm refuses a condition docketing written while a proposal stands (L28); the DocketRuling write checks (`requireSubjectMatches`' docket arm, `motion.go:465`: names the docketing; `motion_id` iff `D.motion_id`); `RegisterSeat` (`:290`): the exit register's docketings (L29)
- `F/tools/internal/record/motion.go` S — `requireSubjectMatches`' docket arm; `requireUnanswered`'s docketing arm (a second ruling of D)
- `F/tools/internal/record/benchgrade.go` (8) DOCKETING-STANDS — `requireRegradeWithinBenchLock`'s "a docket standing unruled" → a docketing; the bench's own regrade is deleted (amendment 2, M17)
- `F/tools/internal/record/supersession.go` (4) DOCKETING-STANDS — `requireSupersededOffDocket`
- `F/tools/internal/record/available.go` (3) DOCKETING-STANDS + PROPOSAL-STANDS — the freeze off the lists (`:58-63`)
- `F/tools/internal/record/sitting.go` (4), `viewjson.go` (9) DOCKETING-STANDS — the work list's docket rows name the docketing, its cause and basis; the bench's held set; the remand-limit clause (TEXT)
- `F/tools/internal/record/actjudge.go` (3) PROPOSAL-STANDS — `requireExchangedOn`, `requireNotEstoppedOn` for every `Motion` writer (no `byDispatch`); `requireNoStandingDocket` on `Close`
- `F/tools/internal/record/tombstone.go` (B3) S — placement `{G, RULING, docketing_id}`; S's read through D's ruling's `motion_id`
- `F/tools/internal/record/verdict.go`, `recordread.go`, `positionheard.go` (1 each) DOCKETING-STANDS — read "a docket stands" for the outcome; re-read at the build, each keyed on the docketing
- `F/tools/internal/cli/seat/regrade.go` (1) DOCKETING-STANDS — the bench regrade's help ("while a docket motion stands" → "while the gap is docketed")
- `F/tools/internal/cli/motion/verbs.go:234` TEXT — "A DOCKET THE DISPATCH FILES SAYS WHY IN ITS BASIS" → the docketing's cause and basis, on `motion docket rule --help`; `propose.go:207` comment (DELETE); `cli/chair/dispatch.go:22-23` TEXT; `cli/seat/help/dispatch.md:21` TEXT; `closing-red.md`, `closing-blue.md` TEXT; `help_gen.go` D
- `F/tools/internal/terms/terms.json` TEXT — `docketing` added, `docket motion`, `impasse`, `grade movement` reworded; `definitions_gen.go`, `docs/vocabulary.md` D
- `scripts/agentgen/src/lead-judge.md`, `src/fragments/lens-duties.md` TEXT; `F/agents/*.md` (10 files) D; `F/skills/research-protocol/scripts/debate.js` (enginegen region, the terminal and assembly prompts) TEXT; `F/docs/seat-command-triggers.md` TEXT; `F/tests/simulator/debate.test.mjs` S
- `F/tools/internal/capture/capture.go` (5 `docketing`) DOCKETING-STANDS — the closings audit (`ClosingDuties`) and precedent harvest read the docketing
- **audit p7 gap 2** — `F/tools/internal/record/lensturn.go:81-135` (`EarlyDockets`, `e.GetSeatId() != chairSeat` at `:121`) DOCKETING-STANDS: reads `Docketing` rows of cause `IMPASSE`; `F/tools/internal/capture/lensturn.go:32-50` (`DocketOrderAudit`) TEXT ("docket motion(s) filed" → docketing(s) written); `record/dispatch.go:470,1467-1477` (`docketedThisSitting`) and `:810-830` (`epochLimitHeldTheBench`) DOCKETING-STANDS; `releasegate/fuzz/lensturn_test.go:467` (the docket-order oracle) S; `record/m37replay_test.go:103`, `m38replay_test.go:91-98`, `capture/lensturn_test.go:65-119` S (non-zero expected, T-B33). Sweep: `git grep -n -E 'chairSeat|RoleOfSeat\([^)]*\) *== *"chair"' a488cad41 -- plugins/frank-exchange-of-views/tools ':!*_test.go'` (30 hits) intersected with readers of `Motion`/`GetDocket()` → `lensturn.go:121`, `dispatch.go:470,823,1476`; the rest read sittings, positions or verdicts (O)
- **audit p7 gap 6, prose carriers no pattern above reaches** TEXT — `F/skills/research-protocol/scripts/debate.js:925` (the bench prompt's "YOUR WORK LIST NAMES THE MOTIONS DOCKETED FOR YOU: the record filed each" → the docketings, each with its cause and the change at issue) and `F/tests/simulator/debate.test.mjs:698` (asserts it); `F/tools/internal/record/passblockers.go:527,533` (the closing-owed blocker's "docketed for the bench (motion %s)" → the docketing) and `closingowed_test.go:165`; `F/tools/internal/record/available.go:283-303` (the remanded-gap chair row's route switch: a `routeProposal` arm, "its assignee is readied for its window"); `dispatch.go:483-484` (`routeNobody`'s "open and not material" — unchanged: amendment 3 deletes the park, so no material gap routes nobody). Sweep: `git grep -n -i -E 'motions? docketed|docketed for the bench \(motion|bench regrades|readies nobody|not material' a488cad41 -- plugins/frank-exchange-of-views ':!*.golden'` (each hit read; the `not material`/`readies nobody` hits are materiality's own sense, O, but those named here)
- **Amendment 5: the final sitting's closing sitting** ("Red gets its closing", picked 2026-11-21) DOCKETING-STANDS — `record/dispatch.go:379-407` (`awaitClosings`: the `plan.FinalSitting` arm and its why-line "no seat is dispatched after this sitting, so none can be filed" deleted), `:702-706` (the limit's emptying keeps each closing party, engaged on its docketed gaps only), `debate.js:412` (`final_sitting`'s description: only the closings owed are filed after it), `:968` (the limit break follows the closing parties' sitting), `:914` (the chair's "NOBODY SITS AFTER YOU" → only the closings owed), the party prompt at a final closing sitting, `record/actjudge.go` (L44's arm: in a sitting opened after a `FinalSitting` plan, a party's act other than its closing refused, naming the final sitting), `testdata/plans.golden` and `prompt-red-chair.golden` D; T-B37, T-B9 (2)
- **Amendment 6: the closings-only sitting — the final one and held-out blue's** (audit p9 gaps 3 and 4; F-HO (b); F-X's pick not folded, so `HeardInTime` gains no arm). Sweep: `git grep -n -i -E 'no seat is dispatched|nobody sits|NOBODY SITS|not dispatched|none can be filed|FinalSitting|final_sitting|heldOut' a488cad41 -- plugins/frank-exchange-of-views scripts/agentgen ':!*_test.go' ':!*.golden' ':!*_gen.go'` → **112 hits / 27 files** (`~/.claude/scratch/1695/amend6/m16-final-sweep.txt`); every file below, its hit count summing to 112, 0 untagged. Each S hit is reworded to say that the closings-only sitting's parties sit and file closings only, and that a motion filed there is still refused for its own reason (no assignee window and no exchange follows, `HeardInTime`):
  - `record/dispatch.go` (40): S — `:116-120` `FinalSitting`'s comment and `desc` (the source of `debate.js:412`'s schema text, audit p9 NOTE: "no seat is dispatched after it" → only the closings owed are filed after it); `:121-126` `EpochLimitReached` (the held-out arm fires only where blue owes no closing); `:175-181` `closingsAwaitedWhy`'s `held` clause (deleted); `:322-337` `engage`'s held-out return (admits blue for a closing duty only); `:384-405` `awaitClosings`' two arms and why-lines (deleted); `:680-692` the one-sitting-early CEILING arm (blue's closings make it a party); `:702-706` (amendment 5's row); `:717-720` ("blue sits neither there nor at the final sitting" → blue sits in both for its closings only, so the remand's exchange still cannot be had); `:732` `isFinalEpoch`'s comment; `:739-751` `blueHeldOut`'s comment and `blueHeldOutClause` ("readies no blue" → none but for its closings); `:772-790` `HeardInTime` (its refusal's "no seat is dispatched after this sitting" reworded; no new arm). K — `:136` (`Blockers`' comment: the relay emptied, the engine's "nobody sits" reading, unchanged), `:374-375` (`waiting`: the lens's re-audit, unchanged), `:522-529`, `:603-604`, `:654-655` (blue held out for acts, unchanged), `:69`, `:325-326` (field and assignment).
  - `cli/motion/verbs.go` (6) S — `:46-53` `finalSittingSays` (the reason "no seat is dispatched after that sitting" → after it only the closings owed are filed, and no assignee window or exchange follows; inlined into every `motion … file` page and the 11 agent definitions below); `:75` K (caller). `cli/motion/propose.go:119` (1) K (renders it).
  - `cli/seat/help/dispatch.md` (2) — `:25` S (`final_sitting`: closings only after it; the held-out sentence: blue dispatched for its closings); `:23` K (the key's exclusions).
  - `skills/research-protocol/scripts/debate.js` (8) — `:412` D (enginegen, from `dispatch.go:116-120`); `:398`, `:401`, `:421` D (field lists); `:914` S ("NOBODY SITS AFTER YOU" → only the closings owed); `:968` S (the closing parties' sitting before the limit break; held-out blue dispatched as a party); `:1100` S ("the parties the board still readied were not dispatched" → beyond their closings); `:597` K (the emptied-relay check).
  - `record/verdict.go:134` (1) S — CEILING's reason, as `debate.js:1100`.
  - `agents/*.md` (11 files, 25 hits: `red-chair.md` 5, every other 2) D — `agentgen` inlines `finalSittingSays` and `dispatch.md`; `agentgen -check`.
  - `record/available.go` (3), `record/sitting.go:122` (1), `record/viewjson.go:1456` (1) K — `blueOut` (blue held out for acts; the chair's remand row).
  - `tests/simulator/debate.test.mjs` (12) — `:117`, `:457` S (held-out blue owing a closing is dispatched for it: a row each, T-B37); `:317`, `:371`, `:380`, `:432`, `:460`, `:1101`, `:1116-1117`, `:1595`, `:1624` K (limit and stop wordings with no closing owed). `tests/simulator/harness.mjs` (2) K.
  - O — `tests/simulator/testdata/{m3-chair3-relays,o3-chair5-relays,m16-chair6-relay}.json` (7, archived relays), `seatprobe/production.go` (2), `devcmd/seatprobe/main.go:100` (1): other sense or a literal `false`.
- **every `case route` switch** (`git grep -n 'case route' a488cad41 -- plugins/frank-exchange-of-views/tools`: `available.go:291-301`, `dispatch.go:458-517`, `remand.go:329,385-389`) DOCKETING-STANDS — each gains a `routeProposal` arm (its assignee's window; the window's end writes `Docketing{PROPOSAL}`); no other route is added
- `F/tools/internal/seatprobe/boards.go`, `build.go` (`docketing`) S — the bench board convened on a docketing; the dispatch board writes no motion
- `F/tools/internal/record/migrate/*` S — Translation 3 (plan §III.B13; O2, O3 ruled 2026-11-20) and Translation 4 (§III.B14, M17); `entries.go`, `replay.go`'s `docketing` mentions (TEXT); `remap.go` `idFields` += `Docketing.docketing_id`, `.gap_id`, `.motion_id`, `DocketRuling.docketing_id`, `Tombstone.docketing_id`
- test-side per-event-type lists (M15 command (5)) S — `releasegate/fuzz/fuzz_test.go` `verbsWithEvents` and `reportExemptions` (`docketing`: written by the dispatch, carrying the tool's `basis`), `cli/vocabulary_test.go` `noVerbOfItsOwn`, `recordpb/correspondence_test.go`
- tests S — `record/remand_test.go` (24), `m37replay_test.go` (22), `testdata/plans.golden` (7, every changed line read), `proposalanswers_test.go` (6: `TestTheGradeMovementFilingRestoresTheNetMovement` DELETE; `TestRestoringIsTheFirstFromAgainstTheLastTo` kept, re-pointed at `--restore`, M17), `grademovement_test.go` (6), `passblockers_test.go` (5), `cli/stategraph_test.go` (5), `lensturn_test.go` (4), `dispatchnext_test.go` (4), `proposalhelpers_test.go` (2), `m38replay_test.go` (2), `docketruling_test.go` (2), `blockerowner_test.go` (2), `cli/epochlimitexit_test.go` (2), `withheldpass_test.go`, `planfields_test.go`, `narrowedview_test.go`, `motionview_test.go`, `gavel_test.go`, `cli/petitionreadiness_test.go`, `cli/dispatch_test.go` (1 each); `releasegate/fuzz/remandroute_test.go` (6), `fuzz_test.go` (6), `oneproposal_test.go` (3), `benchorders_test.go` (3), `converseoracle_test.go` (2), `benchlift_test.go` (2), `proposal_test.go`, `gaproads_test.go` (1 each); `integration/surface/promptverbs_test.go`, `constitutiondirective_test.go`, `devcmd/enginegen/vocabulary_test.go` (1 each, `restoring`/`routeBench` in prompt text: re-read); [NEW] T-B22–T-B26
- O — `fetchcache/fetchcache.go`, `extractor.go`, `testdata/walls/pmlr.html` (3), `cli/ocr.go`, `cli/lens/verify.go`, `vecembed/vocab.txt`, the five `tessocr/testdata/corpus/gpo-record-threecolumn/reading.*.golden`: `restoring` in another sense

## M17 the bench's own closing word and own regrade, and every carrier of the docketing's answers — Part B (amendments 2 and 4, 2026-11-20, plan §III.B14)

**Amendment 4 (each side states its position; side-named answers; a stored side).** Every CHOOSE line
below that names `--accept | --reject` names D's pair instead — `--for-red | --for-blue` (amendment 7
deletes B2a's `--for-filer | --for-assignee`, M21) — and teaches each
side's statement (red's text in its closing, blue's word in its closing, a proposal, the moved or the
recorded grades) and that a side stating nothing cannot win. Every reader of `DocketRuling.accepted`
reads `sided_with` (tag **SIDED**, replacing REJECT-MOVES): "the held act applied" is `sided_with =
BLUE` (amendment 7). Sweep re-run at the build: `git grep -n -E 'GetAccepted\(\)|"accepted"|"rejected"|--accept\b|--reject\b|flags\.(Accept|Reject)' a488cad41 -- plugins/frank-exchange-of-views/tools plugins/frank-exchange-of-views/skills scripts/agentgen ':!*_test.go'`,
each hit tagged SIDED, CHOOSE (the docket rule's flags), or O (`blue edit --accept`, `motion docket
accept`, another verb's flag).

Command, on `a488cad41` (`~/.claude/scratch/1695/amend2/m17-files.txt`, 49 files, 143 hits):

```
git grep -c -i -E 'regrade-bench|bench.?s (own )?regrade|BenchRegrade|requireRegradeWithinBenchLock|benchSetAxesOf|docket rule[^"]*--as|--as remanded|flags\.As|own closing word|closing word of its own|accept_without_proposal_refused|no proposal stands|ruleDocketAtExit|ruleStandingDockets' a488cad41 -- plugins/frank-exchange-of-views scripts/agentgen ':!*/vecembed/vocab.txt'
```

plus, read by hand, the agentgen sources the pattern does not reach (`lead-judge.md:42-43,82,112`,
`red-chair.md`), debate.js `:491,927,1123-1128` and the assembly prompt, `terms.json`'s `disposition`
(`:815-857`). Tags: **CHOOSE** (a surface where the bench chooses a word or grade: deleted, or reworded to
D's pair (amendment 4: `--for-red | --for-blue`; amendment 7 deletes `--for-filer | --for-assignee`) | remand and the positions), **TOOL-WRITES** (record code that now writes the word or
grade from the record), **DELETE**, **TEXT**, **O** (another sense — `flags.As` of another verb).

- `F/tools/internal/record/recordpb/record.proto` TOOL-WRITES — `DocketRuling.disposition` loses `flag: "as"`, its `why` and the side-or-remand CHECK's `why` say the tool sets it (red's `repaired`, blue's stated word or `not_a_defect`, a proposal's stored word, `remanded`); `accepted = 9` reserved and `sided_with = 11` (`SidedWith`: `RED`, `BLUE`, `FILER`, `ASSIGNEE`; means "the position whose statement this ruling applied, as the answer named it") with its CHECKs (amendment 4); `Closing.closure_class = 6` (blue's word, `subset: "closes"`, `flag: "as"`; M19); `Disposition`'s "the bench's word" comment; `remanded`'s means unchanged; `Closing.location`, `fix_new`, `retire` (M19)
- `F/tools/internal/record/motion.go:540` CHOOSE — the docket subject's `--as` values (`MotionVerdicts`) → none; `requireSubjectMatches`' docket arm loses "no proposal stands"; `requireUnanswered` unchanged
- `F/tools/internal/record/record.go` (3) TOOL-WRITES — validate's `DocketRuling` arm: `sided_with` or `moved_grades` in D's pair; each answer's word is its row's (`repaired` after the bench's edit of red's or the assignee's text, blue's closing's word or `not_a_defect`, the stored word); for-red (for-assignee) refused wherever that side's statement is text and its closing carries none (amendment 5); a for-red (for-assignee) answer's bench `BlueEdit`/`Retire` is the re-plan of that closing, and its road meets `EstoppelConflict` (amendment 5, gap 3); the bench `Regrade` in a for-blue (for-filer) answer to a regrade proposal's or `--restore`'s transaction carries the stored or restored grades; no remand on a grade-movement docketing; no other bench `Regrade`; validate's `BlueEdit`/`Retire` seat arm admits the bench only in a for-red ruling's transaction (M20)
- `F/tools/internal/record/benchgrade.go` (8) TOOL-WRITES + TEXT — `requireRegradeWithinBenchLock` and `benchSetAxesOf` stay, for the two tool-written bench regrades, and `benchSetAxesOf` also folds the axes every ruling on a grade decided (`--keep`; either answer to a regrade proposal), locked at the grade it leaves (amendment 5, L35′), with `--restore`; **amendment 6 (L35′, "A ruling re-locks")**: the toggle is deleted — `:22-24` ("one that moves an axis it holds LIFTS the lock … Lock returns to lens … benchSetAxesOf folds that toggle"), `:101` ("the mover's seat toggles the lock") and `:107-108` ("THE BENCH'S MOVE TOGGLES THE LOCK … lifts its own lock") → the ruling re-locks each axis it decided at the grade its transaction leaves, the `Regrade` and the ruling folded as one step; `record/benchorders_test.go:216` ("THE LOCK RETURNS TO THE LENS") and `releasegate/fuzz/benchorders_test.go:15` ("releases it to the lens") re-pointed at the re-lock (T-B36's R1→R2 row); swept by `git grep -n -i -E 'lock returns|returns to the lens|lifts (its|the) lock|toggles? the lock|releases it to the' a488cad41 -- plugins/frank-exchange-of-views scripts/agentgen` (6 hits, the `diagnostics/manual_test.go:106` hit O); the file comment's "the bench files an ordinary `regrade`" (2026-10-30) → the tool writes it in the ruling's transaction
- `F/tools/internal/record/grademovement.go` (1) TOOL-WRITES — `restoring` kept (§III.B13 deleted it), now `--restore`'s at grade movement (picked 2026-11-20; amendment 5's `--keep \| --restore`); `GradeMovementRule`'s "that ruling rules on the gap's movement" → keeps or restores it; the window (L41, its regrade-proposal arm L41′, M16)
- `F/tools/internal/record/impasse.go` (1) TOOL-WRITES + TEXT — `:308`'s comment ("The bench's regrade sets the grade its ruling rules on") → the tool's; `:314` reads a `BlueEdit` as blue's act in the exchange — the bench's for-red edit is not (M20). Amendment 2's upheld round and `GapExchanges.upheld` are withdrawn (amendment 3)
- `F/tools/internal/record/remand.go` S — `requireRemandWithinLimit`'s refusal text ("takes a disposition that closes it", `verbs.go:237` too) → a side of D's pair; a remand on a `GRADE_MOVEMENT` docketing refused (L40); amendment 2's `upheld` gate on `routeOf`'s impasse arms withdrawn
- `F/tools/internal/record/derivedword.go` — not built: amendment 3 withdraws `derivedClosingWord` (L34, the side's word)
- `F/tools/internal/record/supersede.go:303`, `enums.go` (6), `cli/chair/verdict.go` (5), `cli/bench/outcome.go` (3), `cli/blue/inquiry.go` (3), `cli/lens/reproduce.go` (3), `cli/lens/verify.go` (3), `cli/seat/closure.go` (3), `cli/motion/propose.go:68-76` O — `flags.As` of another verb (the proposal's own closure class, `bench outcome --as`, the verdict, an inquiry's status, a reproduction's or verification's outcome)
- `F/tools/internal/record/viewjson.go` (1), `sitting.go` S — the bench's work-list row on a docketing names the change at issue and G as it stands, and what each answer writes
- `F/tools/internal/cli/bench/command.go:23` DELETE — `seat.Regrade("regrade-bench")`; `F/tools/internal/cli/seat/help/regrade-bench.md` DELETE; `cli/seat/regrade.go:20-21` TEXT (one act, the lens's; the bench's regrade is the tool's)
- `F/tools/internal/cli/motion/verbs.go` (5) CHOOSE — `motion docket rule`: `--for-red | --for-blue | --keep | --restore | --remand` exactly one (amendment 7), from D's pair or the remand (amendment 4), `--direction` with `--remand`, `RequiredUnless(c, flags.As, …)` (`:408`) gone; help `:226-240,293,405` teaches the positions table; `rulingSaid` (`:492`) unchanged
- `F/tools/internal/cli/motion/propose.go:218` TEXT — "The bench's regrade is the grade its ruling rules on" → the stored regrade, written as the bench's
- `F/tools/internal/cli/seat/help/regrade.md:11` TEXT — "AN AXIS THE BENCH SET" stays (the lock), set now by a ruling on a grade (a regrade proposal's either answer, `--keep`, `--restore`; L35′); `close.md`, `mint.md`, `reproduce.md`, `verify.md` (each 1, "the bench's word") TEXT — a bench closure's word is a proposal's or the tool's; `help_gen.go` D
- `scripts/agentgen/src/lead-judge.md:42-43,82,112` CHOOSE — "REASONS, NOT JUST FATES" keeps its four fields; "WHERE YOU HOLD A GAP'S GRADE WRONG, YOU MUST SET IT BEFORE YOU RULE" deleted; the impasse paragraph and `:112`'s "you have the disposition for it" → a side of D's pair (`--for-red | --for-blue`; `--for-filer | --for-assignee`) or remand, each side's statement (red's text, blue's word, a proposal, the grades), that the tool applies the chosen side's and that a side stating nothing cannot win (amendment 4); `src/red-chair.md`, `src/fragments/lens-duties.md`, `lens-audit-and-mint.md` TEXT (grade movement: the bench keeps or restores; amendment 4: blue's closing may state its word, red's carries its text); `F/agents/*.md` (9) D
- `F/tools/internal/record/passblockers.go:933` TEXT — `requireBenchMotionsRuled`'s "`--as <verdict>` … --accept or --reject in place of --as" → D's pair (`--for-red | --for-blue`, or `--for-filer | --for-assignee`) on every docketing
- `F/docs/seat-command-triggers.md:116` (the `bench regrade` row) DELETE — audit p7 NOTE; promptverbs_test's trigger-doc check fails loudly once the verb is gone
- `F/tools/internal/cli/seat/verbs.go:291` (`show board`'s "the bench regrades a gap it is hearing") TEXT → a ruled grade (a regrade proposal's ruling, `--keep` or `--restore`); inlined into the 12 agent definitions (D)
- **SIDED** (amendment 4, replacing amendment 3's REJECT-MOVES: each reader of `accepted` reads `sided_with`; `--restore` at grade movement writes the restore, and grade answers are `moved_grades` (amendment 5, tag MOVED); either side at impasse, and for red on blue's close or supersede with red's text, closes G) — `F/tools/internal/record/inquirycheck.go:139-141` (`moved` counts accepts and not rejects → by the row's writes), T-B8's reject row (G byte-identical → per row), `proposalanswers_test.go:455` (`TestRestoringIsTheFirstFromAgainstTheLastTo`, re-pointed at `--restore`), `tombstonesOwed` (places by shape: no change), `consistency.go:222-250` (`carriedCount` reads closing words: no change). Sweep: `git grep -n -E '"rejected"|Ruling *!?= *"rejected"|GetAccepted\(\)' a488cad41 -- plugins/frank-exchange-of-views/tools ':!*_test.go'`, each hit read for "a reject moved nothing" (12 hits on `a488cad41`): `inquirycheck.go:141` and `retirement.go:249` (gap news words a grade-movement reject "rejected" → names the restore by `moved_grades: RESTORED`) MOVED; `record.go:1699-1702` and `cli/motion/verbs.go:495-498` (`rulingWord`, `rulingSaid`: the word itself → the side's words) SIDED; `proposal.go:243,257,392` (the assignee's accept) and `consistency.go:246` (M18) another reader; `record.pb.go:5202,5838` D
- `F/skills/research-protocol/scripts/debate.js` `:491` (a docket ruling's "fate-changing --as"), `:907` ("only you and the bench hearing it regrade it" → a ruled grade), `:927` ("its fate"), `:1123-1128` (terminal), the assembly prompt CHOOSE — a side on every docketing (red or blue; filer or assignee), the tool applying its statement, nothing remanded at the exit; `tests/simulator/testdata/prompt-red-chair.golden`, `prompt-judge-terminal.golden` D
- `F/tools/internal/terms/terms.json` (`disposition` `:815-857`: the closing word, on a bench closure the statement of the side the bench sided with; `grade movement` `:874`; `side` `:252` unchanged, red or blue) TEXT; `definitions_gen.go`, `F/docs/vocabulary.md` D
- `F/tools/internal/seatprobe/boards.go:643` TEXT — "only the bench's regrade binds the lens" → a ruled grade (a regrade proposal's ruling, `--keep` or `--restore`)
- `F/tools/internal/difftest/testdata/docket_ruling_requires_each_unconditional_field.golden` (16), `manual_runs_every_help_page_live.golden` (2), `projections_debate_changelog_citations.golden` (1) D — regenerated, every changed line read
- tests S — `cli/cli_test.go` (7), `crossseat_test.go` (2), `answers_test.go` (2), `benchorders_test.go`, `lens/axis_usage_test.go`, `record/benchorders_test.go` (4), `docketruling_test.go`, `proposalanswers_test.go` (`TestRestoringIsTheFirstFromAgainstTheLastTo` kept), `regraderesult_test.go`, `integration/surface/promptverbs_test.go` (3): bench regrades through the verb → through a for-blue answer to blue's regrade or `--restore` at grade movement; `--as <word>` and `--accept | --reject` on a docket rule → D's pair or `--remand`; `GetAccepted()` assertions → `GetSidedWith()`
- fuzz S — `fuzz_test.go` (13: `ruleDocketAtExit`, `ruleStandingDockets` draw from D's pair, blue's closing drive states a word on half its impasse closings; the bench regrade drive deleted), `benchlift_test.go` (6, the lock's lift by a later bench regrade → by a later accepted proposal or restore), `benchorders_test.go` (7), `sweeptally_test.go` (6: `accept_without_proposal_refused` DELETE; §III.B14's gates and oracles, amendment 4's `blue_word_applied` and `side_stated_nothing` included)
- `F/tools/internal/record/migrate/*` S — Translation 4 (plan §III.B14, amendment 5): the 16 dispatch rulings naming a docketing a `Close` ended written as no event, beside `actOnSupersededGap`'s left-out ruling ("Drop the 16 rulings", picked 2026-11-21); the 3 kept closing words → a for-blue answer at impasse, `{sided_with: BLUE, disposition: not_a_defect}` (F-A (a)); archived remands → `remanded`, neither field; a live-rule cause other than the basis's fails T-B31; the one archived bench `Regrade` dropped; road-less red closings and closings on an ended docketing written as no event by `refusedArgument` (15 red, 9 blue; lead call 2026-11-21)
- **Amendment 5 (the winner's statement applies; `--keep | --restore`).** Every CHOOSE line that names grade movement's answers names `--keep | --restore` (no side); every line teaching `--for-red` on blue's close or supersede "alone" with no red text teaches its refusal; a ruling on a grade locks its axes (L35′). The SIDED sweep above, re-run on `a488cad41` and pasted (`~/.claude/scratch/1695/amend5/m17-sided.txt`, **98 hits**; tagged by `amend5/m17tag.py`, which fails on an untagged hit — 0 untagged, `amend5/m17-tags.out`): **SIDED, the held act applied → `sided_with IN (BLUE, FILER)`** (6) — `proposal.go:243,257,321` (`:321` `requireBenchSuccessor`'s `rd."accepted" = 1`, SQL), `:392`, `recordsql/gapstate.go:208` (the bench CTE's supersede arm, SQL), `consistency.go:246`; **SIDED, the ruling's word → side or grade answer** (9) — `record.go:1699,1700,1702` (`rulingWord`), `cli/motion/verbs.go:356,495,496,498` (`rulingSaid`), `cli/motion/propose.go:193`, `recordsql/views.go:767` (`docketWordSQL`, `motion_answers`' word: `CASE` over `sided_with`, `moved_grades`, `disposition`); **MOVED, the restore → `moved_grades: RESTORED`** (3) — `inquirycheck.go:141`, `retirement.go:247,249`; **CHOOSE** (14) — `cli/motion/verbs.go:303,304,334,335,405-408`, `passblockers.go:929`, `proposal.go:258,269,326`, `remand.go:242`, `record.proto:1980` (the CHECK's `why`); **O** (34) — `blue edit --accept` and `BlueEdit.accepted` (27: `cli/blue/edit.go` 16, `cli/lens/mint.go:493`, `estoppel.go:56,175`, `verdict.go:17`, `viewjson.go:425`, `reportproj/plan.go:17`, `record.proto:1657`, `changesjson.go:81,127`, `recordsql/views.go:257,267` — the blue-edit view), `--accept-loss` (4: `migratecli.go:34`, `migrate/registry.go:21,50`, `feov-migrate/testdata/help.golden`), the marker grammar's "unless --accept" (3: `cli/seat/seat.go:708,722,753`); **D** (32) — `schema.sql` (12), `record.pb.go` (4), `decode_gen.go` (2), `plans.golden` (7), `manual_runs_every_help_page_live.golden` (6), `migrate/testdata/prose-rewrites.golden` (1).

## M18 every reader of a docket ruling's gap or word through its motion — Part B (amendment 3, audit p7 gap 1, plan §III.B13)

Command, on `a488cad41` (the verdict's sweep; `~/.claude/scratch/1695/p8/gap1-sweep.txt`, 110 hits / 42 files, tags by
`~/.claude/scratch/1695/p8/m18.py`, which fails on an untagged hit):

```
git grep -n -E 'DocketGapByMotion|"motion_rule"[^;]*"motion_id"|mr\."motion_id"|GetMotionId\(\)|motionsAt\(|\.Ruled\(\)' a488cad41 -- plugins/frank-exchange-of-views/tools ':!*_test.go'
```

Tags: **DOCKETING-KEY** (reads a docket RULING's gap or word, or a bench-only `Ruled()`: re-keyed through `docketing_id` — `DocketRulingGap` in Go, the same keys in SQL, the docketings fold `docketingsAt` for `Ruled()`; T-B32 a row each), **DOCKETING-STANDS** (reads an unruled docket motion as "before the bench": the docketing, M16), **PROPOSAL** (reads a filing's own standing or answer: unchanged), **PETITION** (a petition or position motion: another subject, unchanged), **DELETE** (`DocketGapByMotion` itself, `noActRefusal`), **S** (the Motions fold's definitions: kept for proposals; `docketingsAt` [NEW] beside them), **TEXT**, **MIGRATE** (Translation 3 resolves archived rulings to docketings), **GEN** (generated or golden: regenerated, each changed line read).

- **DOCKETING-KEY** (45) — `capture/capture.go:1425,1434,1474,1481`; `consistency/consistency.go:113,231,233,238,246`; `dashboard/model.go:479`; `graph/graph.go:84`; `record/grademovement.go:82,86`; `record/inquirycheck.go:139,141`; `record/motion.go:496`; `record/proposal.go:299`; `record/record.go:914,915,1319,1322`; `record/recordsql/gapstate.go:92,102,190,205`; `record/recordsql/views.go:300`; `record/retirement.go:217,242`; `record/stale.go:118,119`; `record/supersession.go:28,37`; `record/tombstone.go:21,30,35`; `record/viewjson.go:1736,1821`; `report/assemble.go:1186,1291,1353`; `report/motions.go:65`; `scorecard/scorecard.go:365`; `verify/verify.go:408`; `view/view.go:589,707`
- **DOCKETING-STANDS** (12) — `record/available.go:110`; `record/benchgrade.go:201`; `record/benchsitting.go:55`; `record/closingowed.go:138,145,165`; `record/composedsource.go:70,71`; `record/dispatch.go:818`; `record/lensturn.go:132`; `record/passblockers.go:791,792`
- **PROPOSAL** (11) — `record/blueanswered.go:72`; `record/keptgrade.go:89`; `record/proposal.go:236,251,270,291,304,318`; `record/recordsql/gapstate.go:234`; `record/recordsql/schema.go:242`; `record/recordsql/views.go:628`
- **PETITION** (5) — `record/positionheard.go:207,230`; `record/positionowed.go:118`; `record/record.go:1289`; `view/view.go:725`
- **DELETE** (5) — `record/proposal.go:245,255`; `record/replay.go:281,289,300`
- **S** (5) — `record/motion.go:132,134,224,239,282`
- **TEXT** (4) — `capture/capture.go:1430`; `record/record.go:1679,1692`; `report/assemble.go:1184`
- **MIGRATE** (2) — `record/migrate/replay.go:842,854`
- **GEN** (21) — `difftest/testdata/bench_petitions_and_halt.golden:53`; `difftest/testdata/projections_debate_changelog_citations.golden:190`; `record/recordpb/record.pb.go:5480,6038,6766`; `record/recordsql/decode_gen.go:837`; `record/recordsql/testdata/schema.sql:1049,1193,1203,1513,1856,1942,1957,1986`; `record/testdata/plans.golden:1424,1429,1434,1439,1964,2700,2708`

## M19 every carrier of `Closing`, which gains red's text and blue's word — Part B (amendments 3 and 4, plan §III.B14, L37, L34)

Command, on `a488cad41` (`~/.claude/scratch/1695/p8/m19-raw.txt`, 135 hits / 72 files):

```
git grep -n -E 'recordpb\.Closing\b|GetClosing\(\)|EVENT_TYPE_CLOSING\b|"closing"|closing-red|seat\.Closing|RED CLOSING' a488cad41 -- plugins/frank-exchange-of-views scripts/agentgen ':!*.golden' ':!*/record.pb.go'
```

Tags: **SHAPE** (the body, its table and validate), **VERB** (the closing verb's flags and `PlanAccept` call, red's only), **RENDER** (a reader that prints red's closing: it prints the text as well), **DUTY** (closings owed: unchanged — the road is a field of the closing, not a second duty), **DRIVE** (a test or fuzz drive filing a red closing on an `IMPASSE` docketing, or on a `PROPOSAL` docketing of blue's close or supersede: it gains a planned pair, or is a refusal row), **O** (reads a closing's existence or text only: unchanged), **D** (generated).

- SHAPE — `F/tools/internal/record/recordpb/record.proto` (`Closing.location = 3`, `fix_new = 4`, `retire = 5`, each `why` naming the for-red answer; `closure_class = 6`, blue's word, its `why` naming the for-blue answer at impasse), `record/record.go` (2: the closing arm refuses a road on blue's closing, on a docketing no answer applies it to, and a missing road where one is required; a word on red's closing, and on blue's off an `IMPASSE` docketing or where its proposal is its statement; the five words, `repaired_with_regression` by L13's check), `recordsql/testdata/schema.sql` D, `recordsql/decode_gen.go` (7) D, `recordpb/annotations_gen.go` D, `record/required_gen.go` D
- VERB — `F/tools/internal/cli/lens/command.go` (`seat.Closing("closing-red")` gains `--quote`, `--new`, `--retire` and the `blue.PlanAccept` plan, as `cli/lens/mint.go:295-306` verifies a replacement), `cli/seat/seat.go` (the `Closing` builder, one per side), `cli/blue/command.go` (blue's closing: no road; `--as` its word, amendment 4), `cli/seat/help/closing-blue.md` TEXT (the five words, where a word is admitted, the `not_a_defect` default, that the bench may side with blue and the word is what lands), `cli/seat/help/closing-red.md` TEXT (the road, its validation, what the bench does with it, the freeze on blue's span), `cli/seat/help_gen.go` (2) D, `cli/seat/verbs.go` (4) TEXT, `scripts/agentgen/src/lead-judge.md` and the lens agentgen sources TEXT → `F/agents/*.md` (10) D, `F/docs/seat-command-triggers.md` TEXT
- RENDER — (each also prints blue's closing word where stated, amendment 4) `record/viewjson.go` (5), `view/view.go` (5), `report/assemble.go` (the RED CLOSING section), `verify/verify.go` (2), `graph/graph.go`, `seatprobe/boards.go` (4), `seatprobe/build.go`
- DUTY — `record/closingowed.go`, `record/gapacts.go`, `record/supersede.go` (M16's re-key only)
- MIGRATE — `record/migrate/eraentries.go`, `migrate/replay.go`: archived closings carry no road and translate unchanged
- DRIVE — `cli/cli_test.go` (3), `cli/answers_test.go` (3), `cli/positionowed_test.go` (3), `cli/referencechecks_test.go` (2), `cli/epochlimitexit_test.go`, `cli/actrows_test.go`, `cli/refs_test.go`, `cli/verbs_test.go`, `cli/verify_test.go`, `record/closingowed_test.go` (8), `record/docketruling_test.go` (2), `record/passblockers_test.go` (2), `record/terminalepoch_test.go`, `record/emptydischarge_test.go`, `record/queue_test.go`, `record/narrowedview_test.go`, `record/m37replay_test.go`, `record/m38heldout_test.go`, `report/*_test.go` (6), `difftest/{contract,fuzz,scenarios}_test.go` (13), `seatprobe/{fidelity,surfacecoverage}_test.go` (3), `view/view_test.go` (3), `record/viewjson_test.go` (2), `capture/capture_test.go`, `verify/verify_test.go`, `releasegate/fuzz/fuzz_test.go` (7: the red closing drive plans a pair from the report it read), `fuzz/lensturn_test.go` — each read at the build: a red closing on an impasse docketing gains its road, or the row is a refusal row (T-B29); a blue closing on an impasse docketing may state a word, and one elsewhere carrying a word is a refusal row (T-B35)
- O — `record/migrate/{archive,argumentresidue,storedplanids}_test.go` (4): archived closings, unchanged
- **Amendment 5.** DRIVE — `record/gapacts_test.go` (a `Closing` drive, audit p8 NOTE: a refusal row where it files a red closing on a docketing). SHAPE / VERB — the road is required on G's assignee's closing for a non-assignee lens's or the chair's close or supersede proposal too, and refused on a regrade proposal's, a grade-movement docketing's and a closed G's; the closing verb and validate's closing arm call `EstoppelConflict` on the road (audit p8 gap 3); `closing-red.md` teaches that a closing `PlanAccept` refuses every pair of cannot be filed (the anchor-in-span case) and that an estopped road is refused naming its row. MIGRATE — `refusedArgument`'s closing arm drops a road-less red closing on a standing `IMPASSE` docketing (13) and any closing with no docketing standing (2 red, 9 blue), counted by T-B31.
- **Amendment 6 (L49).** `Closing.no_text = 7` joins every carrier above that lists `Closing`'s road fields: `record.proto` and the generated carriers (`record.pb.go`, `decode_gen.go`, `body_gen.go`, `required_gen.go`), the DDL with the road (B2a), validate's `Closing` arm (exactly one of a road and `no_text` where the question takes text), `closing-red`'s flags and help (`--no-text`), the report's RED CLOSING rendering ("states no text"), migrate's `refusedArgument` (unchanged: no archived closing carries it), and T-B29's rows.

## M20 every reader of a `BlueEdit`, to which the bench's for-red edit is new — Part B (amendment 3, plan §III.B14)

Command, on `a488cad41` (`~/.claude/scratch/1695/p8/m20-raw.txt`, 37 hits / 25 files):

```
git grep -n -E 'BodyAs\[\*recordpb\.BlueEdit\]|GetBlueEdit\(\)|recordpb\.BlueEdit\b|EVENT_TYPE_BLUE_EDIT\b|"blue_edit"' a488cad41 -- plugins/frank-exchange-of-views/tools ':!*_test.go' ':!*/record.pb.go' ':!*_gen.go' ':!*/testdata/*'
```

Tags: **WRITE-CHECK** (admits the edit), **SEAT** (keys on the edit being blue's act: must not count the bench's), **ESTOPPEL** (verbatim application: places G's estoppel row for the bench's edit as for blue's), **CONTENT** (reads the edit as a change to the report: the bench's counts as any), **VERB** (blue's own verb: `PlanAccept` reused by the closing and the ruling), **MIGRATE** (archived edits are all blue's: unchanged), **O**.

- WRITE-CHECK — `record/record.go:990` (validate's `BlueEdit` arm: the bench's seat only in a for-red ruling's transaction, its bytes the ruling's re-plan); `record/refs.go:410` (the `answers` rule: G, closed in the same transaction)
- SEAT — `record/impasse.go:314` (the exchange fold: not blue's act), `record/blueanswered.go:233` (blue answered: not blue's answer), `record/gapacts.go:62,216` (a gap's acts: the bench's ruling act, listed under the ruling)
- ESTOPPEL — `record/estoppel.go:80-83,150,336` (the verbatim placement, `ProposalOver`: red's prescription applied verbatim places G's estoppel row)
- CONTENT — `record/anchorchecks.go:261`, `record/standingreading.go:391-398`, `record/retirement.go:382`, `record/spotcheck.go:110`, `record/evidenceview.go:629`, `record/viewjson.go:344`, `record/recordsql/views.go:270,686`, `record/recordsql/read.go:266`, `record/recordsql/schema.go:269`, `report/edits.go:25`, `reportproj/render.go:51`, `seatprobe/seatprobe.go:171`, `cli/surface.go:208`
- VERB — `cli/blue/edit.go:119,351-356` (`PlanAccept` and `Accepted` reused; `repeatsEdit` unchanged), `record/blueedit.go:13-17` (keyed retry: the bench writes no key)
- MIGRATE — `record/migrate/eraentries.go:34`, `migrate/gapplace.go:70,142`, `migrate/replay.go:890`
- O — `record/recordpb/record.proto:261` (the type's means: "a change to the report"; unchanged)

> **M21–M25 use the pre-rewrite numbering**: their X-, L- and §III.B13–B17 references name amendments 7–11's rows, which the rewrite's §III.B12 X1–X20 reuse with other meanings. They are kept as the built phases' evidence; M26 maps the same files to the rewrite's rows.

## M21 the supersede proposal, third-party filing and the lens-against-lens sides — Part B (amendment 7, 2026-11-23, plan §III.B15)

Read on `5dfbb7750` (the branch head, B6c built). Command:
`git grep -n -E 'Motion_Supersede|motion_supersede|Mint supersede|SIDED_WITH_FILER|SIDED_WITH_ASSIGNEE|SidedWith_SIDED_WITH_(FILER|ASSIGNEE)|ForFiler|ForAssignee|for-filer|for-assignee|requireSuccessorMintable|successorCheck|successorRefusal|standingSupersede|unwritableSuccessors|judgedLanded|judgeLanded|SuccessorPlans|actingSeat|mintActingSeat|requireBenchSuccessor|ProposalSide|proposalSide|proposeAsBystander|bystander|successor_unmintable_refused|one_proposal_refused|closed_gap_refused|motion:(lens|chair):|docket file supersede|ProposedActs' HEAD -- plugins/frank-exchange-of-views scripts/agentgen`.
Tags: **DELETE** (the line goes), **NARROW** (the line stays for blue's proposals or red against blue),
**TEXT** (agent-facing wording reworded), **GEN** (a generated carrier, regenerated), **O** (another
sense). Every non-test hit is a row of the plan's deletion list X1–X41, which names each by path and
symbol; the test and fuzz functions are listed in `~/.claude/scratch/1695/successor-tests.txt`
(38 functions, and file-level declarations, in 32 files), each DELETE (whole) or NARROW (its supersede, lens-filer, chair-filer or
`FILER`/`ASSIGNEE` rows).

- DELETE — `record/freeze.go` (L48's arm, X8), `record/supersession.go:38-50` (X11), `cli/blue/freeze.go`'s successor road (X18), `cli/motion/propose.go`'s `supersede` subcommand (X19), `cli/motion/docketrule.go:31-32` (X20), `flags/names.go:116-117,348` (X21), `consistency/consistency.go:118,258-260` (X17), `recordsql/gapstate.go:209-221` (X6), `releasegate/fuzz/bystander_test.go` (X33), `forcedseeds_test.go:47` (X35)
- NARROW — `record/proposal.go` (X10), `record/docketing.go` (X12), `record/benchanswer.go` (X13), `record/actjudge.go:190` (X14), `record/record.go:677-708,940-941` (X9), `recordsql/views.go`'s `sided_with` reads (X7), `cli/motion/command.go` (X22: `docket` filed by blue only), `fuzz_test.go` (X34), `proposal_test.go` (X36), `proposalgates_test.go` (X37)
- TEXT — `record/supersession.go:108`, `cli/lens/mint.go:451-452`, `record/remand.go:231`, `record/passblockers.go:1049`, `record/benchanswer.go:20-21,104-105`, `record/tombstone.go:217`, `cli/motion/command.go:69`, `debate.js:910,918`, `scripts/agentgen/src/red-chair.md:18`, `docs/seat-command-triggers.md:110-114`
- GEN — `record.pb.go`, `record.proto.sha256`, `body_gen.go`, `recordsql/decode_gen.go:796,909-912` (protogen); `recordpb/annotations_gen.go:366-367` (vocabgen); `cli/seat/help_gen.go`, `required_gen.go` (cmdtreegen); `agents/*.md`, `agents-claude/*.md` (agentgen; `lead-judge.md:446-448` and each seat's `motion docket file` page); `difftest` goldens (`UPDATE_GOLDENS=1`)
- O — `Mint.supersedes` and `lens mint --supersedes` (the lens's own act, kept); `requireSupersedable`'s rule (F-CS open); `Retire.superseded_by` (blue's claim retire); `requireSuccessorCarriesDemand`, `requireSuccessorKeepsBenchGrades`, `successorsOf` (lineage readers, kept); the word "assignee" and "filer" in their kept senses (G's assignee; blue as filer)

## M22 lineage past a close, `Close.successor`, the closure words, `demand_lineage`, the repair-regression telemetry and the estoppel bar's seat — Part B (amendment 8, 2026-11-23, plan §III.B16)

Read on `26a69d300` (amendment 7's head). Commands (raw output, numbered, in `~/.claude/scratch/1695/amend8/`):
(1) `git grep -n -i -E 'successor|repaired_with_regression|amends_prior|supersedes|superseded_by|demand_lineage|EstoppelConflict' HEAD -- ':!plans' ':!run-archive'`
→ `census-raw.txt`, 2690 lines in 303 files (`census-byfile.txt`, per file and term: successor 206 files, supersedes 215,
repaired_with_regression 81, superseded_by 54, amends_prior 39, demand_lineage 9, EstoppelConflict 6);
(2) the concepts the seven words miss: `git grep -n -i -E 'RepairRegression|repair_regression|lineage_mints|closed_with_regression|requireNotEstopped|roadEstoppel|mintEstopped|lineage|reopen|regress' HEAD -- plugins/frank-exchange-of-views scripts/agentgen ':!*/record/migrate/testdata/*'`, read for the
sites that teach a link past a close, a regression close, or estoppel binding every lens.
Tags as M21's. Each non-test file is a row of the plan's X42–X67 or a stated keep (O), and every row's
paths are in this list; test and fuzz functions are in `~/.claude/scratch/1695/amend8/tests.txt` (47 files,
74 functions by pattern, plus the closed-ancestor cases added by hand), each DELETE or NARROW; goldens are
regenerated and read. `F/` = `plugins/frank-exchange-of-views/`, `T/` = `F/tools/internal/`.

- DELETE / NARROW — `T/record/recordpb/record.proto` (X42, X43, X45), `T/record/recordsql/demandlineage.go` (X49, whole), `T/record/recordsql/views.go` (X49: `demand_reach`, `demandReachSQL`), `T/record/recordsql/gapstate.go` (X49 `:275`; L57 `assigneeSQL`), `T/record/recordsql/schema.go` (X49 `:293`, `evidenceCheckIndexesSQL` moves in), `T/record/record.go` (X50), `T/cli/seat/closure.go`, `T/flags/names.go`, `T/flags/shapes.go:78`, `T/cli/seat/seat.go:1154-1158`, `T/cli/enumhelp/enumhelp.go:28`, `T/cli/lens/close.go:49` (X51), `T/record/successor.go` (X52: the `Close` arm), `T/record/replay.go` (X53; X50 `:165`), `T/record/enums.go`, `T/record/enumvalue.go` (X54), `T/record/docketing.go:523-526` (X55), `T/record/proposal.go:85-143` (X56), `T/record/estoppel.go` (X57), `T/cli/lens/mint.go:404-520`, `T/record/freeze.go:340-352`, `T/record/benchanswer.go:376-391`, `T/record/standingreading.go`, `T/record/passblockers.go:1126` (X58), `T/record/answersretired.go` (X59), `T/record/supersession.go` (X48: the closed-ancestor refusal), `T/view/view.go` (X45), `T/scorecard/scorecard.go` (X45), `T/record/migrate/{lineage,valuerenames,eraentries,remap,replay}.go`, `T/record/migration/manifest.go:99` (X66)
- TEXT — `T/record/supersede.go:82-92` (X60), `T/record/demand.go`, `T/record/actjudge.go:211-214` (X61), `T/cli/seat/help/{mint,close,carry,engage,verify,reproduce}.md`, `T/cli/seat/verbs.go:352-356` (X62), `T/cli/blue/closing.go:23` (X55), `T/cli/motion/propose.go:69` (X56), `T/view/changes.go:93-103`, `T/record/changesjson.go:78`, `T/cli/blue/edit.go:69` (X58), `T/record/refs.go:350` (X50), `T/record/required.go:29` and `T/record/recordpb/descriptions.go:140` (X42/X46: their `repaired_with_regression` examples), `T/record/viewjson.go:877` (X54's comment), `T/flags/ancestor.go:14` (X48: "supersede one and close the rest — your own with `close`"; another lens's gap is no longer a road), `T/record/sitting.go:702-704` (X52), `F/skills/research-protocol/scripts/debate.js:907,910` (X63), `scripts/agentgen/src/fragments/lens-duties.md:5,9`, `fragments/lens-audit-and-mint.md:8,14`, `red-lens-voice.md:15`, `lead-judge.md:43,110` (X63, X45), `T/terms/terms.json` (X64), `F/docs/seat-command-triggers.md:67,89`, `F/docs/citations-and-bibliography.md:181-184`, `F/docs/record-flow.md:56` (X65), `F/tests/simulator/debate.test.mjs:912` (X45)
- GEN — `record.pb.go`, `record.proto.sha256`, `body_gen.go`, `T/record/recordsql/decode_gen.go` (protogen); `T/record/recordpb/annotations_gen.go:196-197`, `T/terms/definitions_gen.go` (vocabgen); `F/docs/vocabulary.md:648` (vocabdoc); `T/cli/seat/help_gen.go`, `required_gen.go` (cmdtreegen); `F/agents/*.md`, `agents-claude/*.md` (agentgen, from X63's sources); `recordsql/testdata/schema.sql`, `T/record/testdata/{plans,shapes}.golden`, the `difftest` goldens naming `--superseded-by`, either word or the estoppel refusal (`UPDATE_GOLDENS=1`, every line read)
- Amendment 7's rows met by this sweep, kept as they are (X8, X10, X12–X22, X40, X41): `T/record/freeze.go`'s successor arm, `T/record/proposal.go`, `T/record/docketing.go`, `T/record/benchanswer.go`, `T/record/actjudge.go`, `T/record/tombstone.go`, `T/record/onegapperclass.go`, `T/consistency/consistency.go:110-118,255-263`, `T/cli/blue/freeze.go`, `T/cli/motion/{propose,docketrule}.go`, `T/record/itemkind.go` (`ItemDocket`), `T/record/dispatch.go:709` (F8's comment, L53), `T/seatprobe/boards.go:994`, `T/record/viewjson.go:1420` (`rd.Successor`, X18). **Missed by M21 and added here:** `scripts/agentgen/src/fragments/lens-duties.md:5,9` and `red-lens-voice.md:15` (the proposal road on another lens's gap), `F/docs/record-flow.md:56` — X63, X65.
- O — an own amendment of an open gap and its readers: `Mint.supersedes`, `T/record/supersession.go`'s `supersessionsOf`/`SupersededGap`/`requireNotSuperseded`, `T/record/{bluesittings,closingowed,gapacts,impasse,lensturn,windowindex,failedcheck,benchgrade,composedsource,inquirycheck,retirement,questionpart}.go`, the gap view's `superseded_by`, `T/record/successor.go`'s `Mint` arm and `engage` (`T/cli/blue/engage.go`, `T/record/itemkind.go`'s `ItemSuccessorOwed`, `T/record/sitting.go`'s owed successors, `debate.js`'s R2 block), `T/record/onegapperclass.go:283` (own open gap), `T/record/recordtest/recordtest.go`, `T/consistency/consistency.go`'s lineage checks (`:171-182`, `:531-536`, `:683-695`), `T/dashboard/{model,render}.go` (argument chains), `T/graph/graph.go`, `T/verify/verify.go` (`supersedesResolve`), `T/flags/csv.go`, `T/flags/ancestor.go`'s one-ancestor rule; `Retire.superseded_by` (`T/cli/blue/retire.go`, `T/record/evidenceview.go:358`, `T/report/assemble.go:854`, `T/record/migration/manifest.go:172`, the `flags.New` mapping); `supersedes_status`/`supersedes_coverage` (`T/cli/blue/{inquiry,question}.go`, `T/record/{blueanswered,questionpart,gapacts}.go`); `T/report/assemble.go:944` and `T/view/view.go:313,351` (their `Successor()` call stays, X53); `T/record/migrate/{docketing,entries}.go` (Translations 1, 3, 4) and `T/record/migrate/sqlitesource.go:262-265` (an archived table's skip); English or another sense: `T/fetchcache/readscanned.go:136`, `T/record/inquiry.go:21`, `T/record/declarations.go:19`, `F/docs/propagation-and-anchoring.md`, `F/docs/seat-surface-naming.md`, `T/vecembed/vocab.txt`, `F/skills/adversarial-audit/SKILL.md` and its source (own chains); history and corpus: `feov-memory/*`, `ideas/*`, `law/precedents.md`, `research/*`, `run-archive/`
- Re-tags of earlier rows (amendment 8). **M9**: the ASSIGNEE | WRITER split collapses — every row reads the mint's writer, the two words kept for the two questions; `assigneeSQL`'s ancestor arm is X68 (L57). **M12**, reversed where it taught the closed-gap successor: `debate.js:912` (now `:910`) → X63; `help/mint.md`'s "Naming a closed gap carries lineage …" → X62, deleted; `cli/lens/mint.go`'s `--supersedes` flag text ("the new gap is that gap's minter's … whoever mints it") → "the one open gap of yours this gap replaces" (X48); `mint.go`'s estoppel refusal → X58; `record.proto`'s `LOG_TYPE_ESTOPPEL` → X44; `fragments/lens-audit-and-mint.md:8` → X63; `terms.json`'s "superseded (a closed gap)" resolution → X64, deleted; M12's O rows `cli/seat/closure.go`, `flags/names.go`, `flags/shapes.go` (`--superseded-by`) → X51, DELETE; `help/close.md`, `carry.md` stay O for a successor closing its open ancestor, less their `--superseded-by` lines (X62); the fuzz gate `lineage_successors_reowned` and the seatprobe board for a lens minting a successor of another lens's closed gap → DELETE and the closed-ancestor refusal respectively (X67). **M13**: `recordsql/demandlineage.go` → X49, DELETE; `answersretired.go`'s supersede read → X59; `supersession.go` → X48; `viewjson.go`'s `inherits:` → X54; the remaining `Mint.supersedes` readers stay O (an own amendment of an open gap).
- Command (2), run untruncated with `-n` (`~/.claude/scratch/1695/amend8/concept2a.txt`, 105 lines / 42 files, for `RepairRegression|repair_regression|lineage_mints|closed_with_regression|requireNotEstopped|roadEstoppel|mintEstopped`; `estopp-text.txt` and `estopp-fuzz.txt` for `estopp` over help, agentgen sources, skills, terms, docs and fuzz): beyond the rows above it adds `T/cli/seat/verbs.go`'s telemetry OUTPUT schema, `T/record/recordpb/testdata/payload-keys.txt`, the `scorecards` rows of `T/dashboard/testdata/render-*.golden`, `T/cli/board_test.go:76` (X45); `T/seatprobe/boards.go:450` and the `difftest` denominator goldens (X50); `T/cli/seat/help/closing-red.md:11` and `help/edit.md:13` (X62); every estoppel drive and oracle (X67); `T/record/viewjson.go:923` (`requireNotEstoppedOn`, the ruling tombstone's guard) is O.

## M23 one dispute rule and one freeze over every subject; gap on a gap; F5; the closing's voice refusal; assignee → minter — Part B (amendment 9, 2026-11-23, plan §III.B17)

Run on `264dddea3` (`git grep -n -E <pattern> HEAD -- plugins scripts ':!*/vecembed/vocab.txt' ':!plans'`); outputs `~/.claude/scratch/1695/amend9/m23-*.txt`. Each non-test file is an X69–X75 row or an O line; each X-row's paths appear here.

- (1) about-kind gap — `ABOUT_KIND_GAP|AboutKind_ABOUT_KIND_GAP|aboutGapRefusal|about_gap_refused|about-kind gap|"gap" *:|aboutAnchor` → 55 lines / 19 files, 7 non-test: X75 (`cli/lens/mint.go`, `flags/shapes.go`), X69 (`record.proto`, `record.pb.go`). O — `cli/blue/edit.go:658`, `record/standingreading.go:267` (`case "gap":` is the gap token's kind), `migrate/remap.go:94` (the archived tag map).
- (2) answers — `GetAnswers\(\)|\.Answers\b|"answers"|--answers|flags\.Answers` → 393 lines / 93 files, 41 non-test: X69 (list in the row).
- (3) the freeze — `requireRoadsHold|BenchRoadsHold|roadsHold|requireNoStandingDocket|requireSupersededOffDocket|StandingRoads|RedTextApplied|frozenIn` → 45 lines / 15 files, all non-test: X70. O — `recordsql/schema.go`, `recordsql/testdata/schema.sql` (the standing-docketing column, kept), `migrate/docketing.go` (archived docketings).
- (4) F5 — `one_proposal_refused|one proposal at a time|siblingdocket|oneProposal|standingProposal` → 12 lines / 9 files, 4 non-test: X72.
- (5) voice — `refuseMintReportVoice|reportvoice\.Refused|refusedTells` → 9 lines / 4 files: X73 adds the closing's and the bench's callers; `redvoice.go`, `actjudge.go:217`, `freeze.go:66` (L48's arm, X8), `cli/lens/mint.go` kept.
- (6) assignee — `\b([Aa]ssignee[A-Za-z_]*|ASSIGNEE[A-Z_]*|assigneeSQL|assigneeOf)\b` → 1472 lines / 212 files, 130 non-test, against 1 line on `origin/main`: X74 whole (a branch-only word; no migration). The generated carriers (agent definitions, goldens, `help_gen.go`) regenerate from their sources.
- (7) cut — `ReopenedAnchors|BaredAnchors|bared` → 53 lines / 13 files, 7 non-test: O on #1889 (`bluedoc.go`, `cli/blue/{edit,retire}.go`, `cli/lens/mint.go`, `migrate/gapplace.go`, `runtest/reopened.go`, `tombstone.go`); the pure-cut end row is P12b's.
- Reverse direction: every path in X69–X75 is a hit above, except the new callers FZ adds (`cli/blue/{inquiry,question}.go`, `cli/lens/close.go`, `cli/seat/regrade.go`), which no pattern can find before the call exists.

## M24 the side rename; amendment 9's names; F-INQ; L55 reversed; the docket freeze — Part B (amendment 10, 2026-11-23, plan §III.B17)

Run on `cf7415b8c` (`git grep -n -I -E <pattern> HEAD -- plugins scripts ':!*/vecembed/vocab.txt' ':!plans'`); outputs `~/.claude/scratch/1695/amend10/m24-*.txt`. Each non-test file is an X-row or an O line.

- (2) sides — `for-red|for-blue|SIDED_WITH_(RED|BLUE)|SidedWith_SIDED_WITH_(RED|BLUE)|ForRed|ForBlue|"for (red|blue)` → 254 lines / 86 files, 38 non-test, against 0 on `origin/main`: X76 whole (branch-only; agent definitions, goldens and `annotations_gen.go` regenerate). `raiser|for-raiser|for-author|SIDED_WITH_(RAISER|AUTHOR)` → 0: no prior sense to collide with; `author` is not in `terms.json`, and the blue-synthesizer's definition names its seat the report's author, the same sense.
- (3) amendment 9's names — `requireSubjectAnswered|Mint\.answers` → 0 in code (plan-only, deleted); `GetAnswers().*Mint` → `estoppel.go:70-71`, O (a `BlueEdit`'s answers looked up against a mint).
- (4) `demand_lineage|demandLineage|demandReachSQL` → 116 lines / 10 files, 7 non-test: all kept (L55 reversed; X49, X61 withdrawn).
- (5) the docket freeze — `requireNoStandingDocket|requireSupersededOffDocket|requireRoadsHold|BenchRoadsHold|requireBenchDocketAnswer|blueHeldOut` → 56 lines / 17 files, 15 non-test: X70 (`record.go`), X77 (`docketing.go`); the rest kept as built (`freeze.go`, `cli/blue/freeze.go`, `supersession.go`, `actjudge.go`, `available.go`, `closingowed.go`, `dispatch.go`, `sitting.go`, `viewjson.go`, `cli/motion/docketrule.go`, `recordsql/schema.go`, `migrate/docketing.go`).
- (6) F-INQ's `Closing` names — `Closing` holds no `status`, `method`, `reason` or `coverage` (`record.proto`); each is `Inquiry`'s or `QuestionPart`'s field of the same meaning, and `closing-red` mounts none of `--as`, `--method`, `--reason` (`cli/lens/closing.go:27-30`). The fate readers: `fateActsOf` counts an `Inquiry` or `QuestionPart` only from a blue seat (`gapacts.go:148,173`), so the bench-written applied record needs X72's admission; `Closing`'s carriers (M19's set: 28 non-test files hold `Closing`) gain the four fields through the generated carriers and `cli/lens/closing.go`. `author` elsewhere in the tree means a report's writer (`blue-synthesizer.md:3`, `lead-judge.md:67`), the same sense, or a cited source's author (bibliography), another sense the rename does not touch.
- Reverse direction: every path in X69–X77 is a hit above, in M23 or in M25, except X69's new `Closing` fields and X70's new callers (`cli/blue/{inquiry,question}.go`, `cli/lens/close.go`, `cli/seat/regrade.go`), which no pattern finds before they exist. M23 (3)'s freeze row and (4)'s F5 row no longer map to X70 and X72 as amendment 9 wrote them: (3) is (5) here, and (4) is kept as built (L62).

## M25 a change to a gap's subject is its lens's turn; no for-raiser where no lens sits after — Part B (amendment 11, 2026-11-23, plan §III.B17)

Run on `d1c5f18d7` (`git grep -n -I -E <pattern> HEAD -- plugins scripts ':!*/vecembed/vocab.txt' ':!plans'`); outputs `~/.claude/scratch/1695/amend11/m25-*.txt`. Each non-test file is an X-row or an O line.

- (1) the lens's-turn readers — `lensTurnActsOf\(|lensTurnOn\(|lensReadAt\(|\.lensTurn\b|\.lensRead\b|blueActed\(` → 27 lines / 8 files, 6 non-test: X72 (`gapacts.go`: `lensTurnActsOf`, `blueActed`, `lensTurnOn`, `lensReadAt`; `blueanswered.go:62-66`, the turn item); readers that see the widened turn, none edited: `impasse.go:175-194,359-421` (round end, `lensAudit`, `lensFirst`; `rounds.acted` and `answeredRemand` read `gapActsOf`, unchanged), `lensturn.go:51-73` (early docket), `remand.go:381-385` (remand stage: discharge reads `answeredRemand`, so a turn discharges none), `viewjson.go:1240`. `gapActsOf\(` → `impasse.go:283`, `lensturn.go`, `gapacts.go`: O (unchanged).
- (2) the epoch refusal — `func (requireBenchDocketAnswer|isFinalEpoch|blueHeldOut)\b` → `docketing.go:392`, `dispatch.go:875,882`: X77 reads the two predicates L44′ reads; no new name.
- Reverse direction: X72's and X77's paths are hits above.


## M26 one party rule: `authorOf`, the raiser's accept, the turn, `lensSitsAfter`, the freeze, tombstones over every subject, the gap on a gap — Part B (rewrite, 2026-11-24, plan §III.B0–§III.R)

Run on `8334980fe`: `git grep -c -E '<pattern>' HEAD -- plugins/frank-exchange-of-views scripts/agentgen ':!*.pb.go'` (`record/migrate/*` included); outputs `~/.claude/scratch/1695/rewrite/census2/`. Each non-test file maps to the row named, or is an O line. `F/` = `plugins/frank-exchange-of-views/tools/internal/`.

- (1) party-role — `RoleOfSeat\([^)]*\) *[!=]= *"(blue|bench)"|[Rr]ole *[!=]= *"blue"|PartyOf\(|blueRespondSeat|IsClosingParty|closingPartiesOf` → 151 hits / 51 files (29 non-test, 3 migrate) → X17 `authorOf`; seat-class tests (`actjudge.go:85`, `cli/seat/source.go:31`, `sitting.go:311`, `viewjson.go:1573`, `available.go:60`) are O (who may sit or source); a `"bench"` test of an applied act keeps its reason (§III.B0). Non-test: `F/cli/seat/source.go` (1), `F/record/actjudge.go` (1), `F/record/available.go` (2), `F/record/benchanswer.go` (1), `F/record/benchgrade.go` (1), `F/record/benchopinion.go` (1), `F/record/bluesittings.go` (5), `F/record/closingowed.go` (7), `F/record/dispatch.go` (21), `F/record/docketing.go` (3), `F/record/envelope.go` (1), `F/record/freeze.go` (1), `F/record/gapacts.go` (1), `F/record/grademovement.go` (1), `F/record/impasse.go` (1), `F/record/inquirycheck.go` (2), `F/record/migrate/docketing.go` (1), `F/record/proposal.go` (4), `F/record/questionpart.go` (4), `F/record/record.go` (3), `F/record/roles.go` (1), `F/record/sitting.go` (7), `F/record/stale.go` (1), `F/record/successor.go` (8), `F/record/tombstone.go` (1), `F/record/viewjson.go` (5), `F/report/assemble.go` (1), `F/seatprobe/production.go` (1), `F/view/view.go` (2).
- (2) turn — `blueActed|fateActsOf|lensTurnActsOf|lensTurnOn|gapActsOf|win\.Reaches|Reaches\(` → 66 hits / 16 files (8 non-test, 0 migrate) → X7, §III.B3. Non-test: `F/record/available.go` (3), `F/record/blueanswered.go` (8), `F/record/composedsource.go` (1), `F/record/gapacts.go` (16), `F/record/impasse.go` (19), `F/record/lensturn.go` (4), `F/record/remand.go` (2), `F/record/windowindex.go` (1).
- (3) accept-ruling — `newAccept|applyAccepted|ProposalSide|proposalSide|docket accept|requireDocketAnswer|requireUnanswered|RequireUnruledMotion|requireBenchOpinion` → 76 hits / 42 files (29 non-test, 0 migrate) → X5; `requireUnanswered` kept for the raiser's accept; `requireBenchOpinion` keeps its bench arm. Non-test: `plugins/frank-exchange-of-views/agents-claude/red-lens-adversary.md` (1), `plugins/frank-exchange-of-views/agents-claude/red-lens-architecture.md` (1), `plugins/frank-exchange-of-views/agents-claude/red-lens-computation.md` (1), `plugins/frank-exchange-of-views/agents-claude/red-lens-dark-side.md` (1), `plugins/frank-exchange-of-views/agents-claude/red-lens-evidence.md` (1), `plugins/frank-exchange-of-views/agents-claude/red-lens-logic.md` (1), `plugins/frank-exchange-of-views/agents-claude/red-lens-voice.md` (1), `plugins/frank-exchange-of-views/agents/red-lens-adversary.md` (1), `plugins/frank-exchange-of-views/agents/red-lens-architecture.md` (1), `plugins/frank-exchange-of-views/agents/red-lens-computation.md` (1), `plugins/frank-exchange-of-views/agents/red-lens-dark-side.md` (1), `plugins/frank-exchange-of-views/agents/red-lens-evidence.md` (1), `plugins/frank-exchange-of-views/agents/red-lens-logic.md` (1), `plugins/frank-exchange-of-views/agents/red-lens-voice.md` (1), `plugins/frank-exchange-of-views/docs/seat-command-triggers.md` (1), `F/cli/motion/command.go` (1), `F/cli/motion/docketrule.go` (1), `F/cli/motion/propose.go` (7), `F/record/benchanswer.go` (1), `F/record/benchopinion.go` (2), `F/record/docketing.go` (9), `F/record/motion.go` (7), `F/record/proposal.go` (1), `F/record/record.go` (6), `F/record/recordpb/record.proto` (1), `F/record/recordpb/requiredfields.go` (1), `F/record/recordsql/views.go` (1), `F/report/layout.go` (1), `F/seatprobe/boards.go` (2).
- (4) sides — `SidedWith|sided_with|SIDED_WITH|for-red|for-blue|ForRed|ForBlue|ForFiler|ForAssignee|for-filer|for-assignee` → 387 hits / 96 files (41 non-test, 2 migrate) → X2, X16. Non-test: `plugins/frank-exchange-of-views/agents-claude/lead-judge.md` (5), `plugins/frank-exchange-of-views/agents-claude/red-lens-adversary.md` (1), `plugins/frank-exchange-of-views/agents-claude/red-lens-architecture.md` (1), `plugins/frank-exchange-of-views/agents-claude/red-lens-computation.md` (1), `plugins/frank-exchange-of-views/agents-claude/red-lens-dark-side.md` (1), `plugins/frank-exchange-of-views/agents-claude/red-lens-evidence.md` (1), `plugins/frank-exchange-of-views/agents-claude/red-lens-logic.md` (1), `plugins/frank-exchange-of-views/agents-claude/red-lens-voice.md` (1), `plugins/frank-exchange-of-views/agents/lead-judge.md` (5), `plugins/frank-exchange-of-views/agents/red-lens-adversary.md` (1), `plugins/frank-exchange-of-views/agents/red-lens-architecture.md` (1), `plugins/frank-exchange-of-views/agents/red-lens-computation.md` (1), `plugins/frank-exchange-of-views/agents/red-lens-dark-side.md` (1), `plugins/frank-exchange-of-views/agents/red-lens-evidence.md` (1), `plugins/frank-exchange-of-views/agents/red-lens-logic.md` (1), `plugins/frank-exchange-of-views/agents/red-lens-voice.md` (1), `plugins/frank-exchange-of-views/docs/seat-command-triggers.md` (1), `F/cli/blue/closing.go` (1), `F/cli/lens/closing.go` (3), `F/cli/motion/docketrule.go` (8), `F/cli/motion/propose.go` (1), `F/consistency/consistency.go` (2), `F/flags/names.go` (5), `F/record/benchanswer.go` (19), `F/record/docketing.go` (18), `F/record/estoppel.go` (1), `F/record/freeze.go` (2), `F/record/grademovement.go` (1), `F/record/migrate/docketing.go` (3), `F/record/passblockers.go` (1), `F/record/proposal.go` (3), `F/record/record.go` (2), `F/record/recordpb/annotations_gen.go` (8), `F/record/recordpb/record.proto` (14), `F/record/recordsql/decode_gen.go` (3), `F/record/recordsql/gapstate.go` (1), `F/record/recordsql/views.go` (1), `F/record/remand.go` (1), `F/record/retirement.go` (2), `F/record/supersession.go` (2), `F/record/tombstone.go` (2).
- (5) epoch-labels — `isFinalEpoch|blueHeldOut|FinalSitting|closingsOnly|requireClosingsOnlySitting|HeardInTime` → 123 hits / 22 files (9 non-test, 0 migrate) → X19 (`lensSitsAfter`); `isFinalEpoch`/`blueHeldOut` stay as its inputs; `HeardInTime` (petitions) O. Non-test: `plugins/frank-exchange-of-views/skills/research-protocol/scripts/debate.js` (6), `F/cli/motion/propose.go` (1), `F/cli/motion/verbs.go` (1), `F/record/actjudge.go` (14), `F/record/available.go` (3), `F/record/dispatch.go` (29), `F/record/record.go` (1), `F/record/sitting.go` (7), `F/record/viewjson.go` (1).
- (6) freeze — `requireRoadsHold|BenchRoadsHold|standingRoadsAt|StandingRoads|frozenIn|frozenDebtIn|RedTextApplied|redTextAppliedAt|requireNoStandingDocket|requireSupersededOffDocket|requireSuccessorMintable|judgeLanded|unwritableSuccessors` → 70 hits / 17 files (15 non-test, 1 migrate) → §III.B5 (kept: roads, `RedTextApplied`, `frozenIn`, `frozenDebtIn`); X3 (successor arm, window loop); X21 (`BenchRoadsHold`, audit p12). Non-test: `F/cli/blue/freeze.go` (9), `F/cli/motion/docketrule.go` (4), `F/cli/motion/propose.go` (1), `F/record/actjudge.go` (3), `F/record/benchanswer.go` (1), `F/record/closingowed.go` (1), `F/record/estoppel.go` (8), `F/record/freeze.go` (18), `F/record/migrate/docketing.go` (1), `F/record/proposal.go` (1), `F/record/record.go` (5), `F/record/recordsql/schema.go` (1), `F/record/sitting.go` (2), `F/record/supersession.go` (3), `F/record/viewjson.go` (10).
- (7) applied — `benchActGap|applied_verbatim|AppliedVerbatim|verbatimPlacement|appliedVerbatim|ProposalAppliedVerbatim` → 126 hits / 54 files (36 non-test, 1 migrate) → §III.B0 (the bench writes the application), §III.B6 (`verbatimPlacement` on every `--accept G`). Non-test: `plugins/frank-exchange-of-views/agents-claude/blue-lane.md` (1), `plugins/frank-exchange-of-views/agents-claude/blue-researcher.md` (1), `plugins/frank-exchange-of-views/agents-claude/blue-synthesizer.md` (1), `plugins/frank-exchange-of-views/agents-claude/lead-judge.md` (1), `plugins/frank-exchange-of-views/agents-claude/red-chair.md` (1), `plugins/frank-exchange-of-views/agents-claude/red-lens-adversary.md` (1), `plugins/frank-exchange-of-views/agents-claude/red-lens-architecture.md` (1), `plugins/frank-exchange-of-views/agents-claude/red-lens-computation.md` (1), `plugins/frank-exchange-of-views/agents-claude/red-lens-dark-side.md` (1), `plugins/frank-exchange-of-views/agents-claude/red-lens-evidence.md` (1), `plugins/frank-exchange-of-views/agents-claude/red-lens-logic.md` (1), `plugins/frank-exchange-of-views/agents-claude/red-lens-voice.md` (1), `plugins/frank-exchange-of-views/agents/blue-lane.md` (1), `plugins/frank-exchange-of-views/agents/blue-researcher.md` (1), `plugins/frank-exchange-of-views/agents/blue-synthesizer.md` (1), `plugins/frank-exchange-of-views/agents/lead-judge.md` (1), `plugins/frank-exchange-of-views/agents/red-chair.md` (1), `plugins/frank-exchange-of-views/agents/red-lens-adversary.md` (1), `plugins/frank-exchange-of-views/agents/red-lens-architecture.md` (1), `plugins/frank-exchange-of-views/agents/red-lens-computation.md` (1), `plugins/frank-exchange-of-views/agents/red-lens-dark-side.md` (1), `plugins/frank-exchange-of-views/agents/red-lens-evidence.md` (1), `plugins/frank-exchange-of-views/agents/red-lens-logic.md` (1), `plugins/frank-exchange-of-views/agents/red-lens-voice.md` (1), `F/cli/blue/edit.go` (6), `F/record/benchanswer.go` (4), `F/record/changesjson.go` (4), `F/record/estoppel.go` (9), `F/record/landing.go` (3), `F/record/recordpb/record.proto` (4), `F/record/recordsql/decode_gen.go` (2), `F/record/recordsql/views.go` (2), `F/record/standingreading.go` (1), `F/record/tombstone.go` (5), `F/record/verdict.go` (1), `F/report/edits.go` (1).
- (8) minter-rename — `\bassignee|Assignee|assigneeSQL|assigneeOf` → 1437 hits / 208 files (111 non-test, 3 migrate) → X15. Non-test: `plugins/frank-exchange-of-views/agents-claude/blue-lane.md` (11), `plugins/frank-exchange-of-views/agents-claude/blue-researcher.md` (16), `plugins/frank-exchange-of-views/agents-claude/blue-synthesizer.md` (15), `plugins/frank-exchange-of-views/agents-claude/lead-judge.md` (21), `plugins/frank-exchange-of-views/agents-claude/red-chair.md` (22), `plugins/frank-exchange-of-views/agents-claude/red-lens-adversary.md` (19), `plugins/frank-exchange-of-views/agents-claude/red-lens-architecture.md` (19), `plugins/frank-exchange-of-views/agents-claude/red-lens-computation.md` (19), `plugins/frank-exchange-of-views/agents-claude/red-lens-dark-side.md` (19), `plugins/frank-exchange-of-views/agents-claude/red-lens-evidence.md` (19), `plugins/frank-exchange-of-views/agents-claude/red-lens-logic.md` (19), `plugins/frank-exchange-of-views/agents-claude/red-lens-voice.md` (20), `plugins/frank-exchange-of-views/agents/blue-lane.md` (11), `plugins/frank-exchange-of-views/agents/blue-researcher.md` (16), `plugins/frank-exchange-of-views/agents/blue-synthesizer.md` (15), `plugins/frank-exchange-of-views/agents/lead-judge.md` (21), `plugins/frank-exchange-of-views/agents/red-chair.md` (22), `plugins/frank-exchange-of-views/agents/red-lens-adversary.md` (19), `plugins/frank-exchange-of-views/agents/red-lens-architecture.md` (19), `plugins/frank-exchange-of-views/agents/red-lens-computation.md` (19), `plugins/frank-exchange-of-views/agents/red-lens-dark-side.md` (19), `plugins/frank-exchange-of-views/agents/red-lens-evidence.md` (19), `plugins/frank-exchange-of-views/agents/red-lens-logic.md` (19), `plugins/frank-exchange-of-views/agents/red-lens-voice.md` (20), `plugins/frank-exchange-of-views/docs/seat-command-triggers.md` (5), `plugins/frank-exchange-of-views/docs/vocabulary.md` (7), `plugins/frank-exchange-of-views/skills/research-protocol/references/run-directory.md` (1), `plugins/frank-exchange-of-views/skills/research-protocol/scripts/debate.js` (7), `plugins/frank-exchange-of-views/tests/simulator/debate.test.mjs` (2), `F/capture/capture.go` (3), `F/capture/lensturn.go` (3), `F/cli/blue/edit.go` (1), `F/cli/chair/dispatch.go` (1), `F/cli/lens/close.go` (2), `F/cli/lens/closing.go` (1), `F/cli/lens/command.go` (2), `F/cli/lens/mint.go` (4), `F/cli/motion/command.go` (1), `F/cli/motion/docketrule.go` (8), `F/cli/motion/propose.go` (12), `F/cli/motion/verbs.go` (5), `F/cli/seat/closure.go` (1), `F/cli/seat/help/closing-blue.md` (1), `F/cli/seat/help/closing-red.md` (1), `F/cli/seat/help/dispatch.md` (3), `F/cli/seat/help/mint.md` (2), `F/cli/seat/help/record.md` (1), `F/cli/seat/help_gen.go` (5), `F/cli/seat/regrade.go` (2), `F/cli/seat/selector.go` (1), `F/cli/seat/verbs.go` (7), `F/consistency/consistency.go` (2), `F/dashboard/model.go` (1), `F/flags/ancestor.go` (1), `F/flags/names.go` (3), `F/graph/graph.go` (1), `F/record/actjudge.go` (2), `F/record/answersretired.go` (2), `F/record/available.go` (14), `F/record/benchanswer.go` (13), `F/record/benchopinion.go` (2), `F/record/benchsitting.go` (2), `F/record/blueanswered.go` (1), `F/record/closingowed.go` (18), `F/record/composedsource.go` (2), `F/record/dispatch.go` (41), `F/record/docketing.go` (18), `F/record/enums.go` (1), `F/record/estoppel.go` (1), `F/record/freeze.go` (25), `F/record/gapacts.go` (8), `F/record/grademovement.go` (2), `F/record/impasse.go` (6), `F/record/keptgrade.go` (1), `F/record/lensturn.go` (2), `F/record/migrate/lineage.go` (2), `F/record/onegapperclass.go` (6), `F/record/passblockers.go` (20), `F/record/positionheard.go` (2), `F/record/proposal.go` (38), `F/record/questionpart.go` (2), `F/record/record.go` (18), `F/record/recordpb/annotations_gen.go` (7), `F/record/recordpb/record.proto` (27), `F/record/recordsql/gapstate.go` (9), `F/record/recordsql/views.go` (4), `F/record/remand.go` (12), `F/record/retirement.go` (5), `F/record/sitting.go` (2), `F/record/stale.go` (3), `F/record/standingreading.go` (2), `F/record/supersession.go` (7), `F/record/tombstone.go` (6), `F/record/viewjson.go` (26), `F/report/assemble.go` (2), `F/report/layout.go` (1), `F/scorecard/scorecard.go` (1), `F/seatprobe/boards.go` (8), `F/seatprobe/build.go` (4), `F/seatprobe/seatprobe.go` (3), `F/sittingwrite/delivery.go` (2), `F/terms/definitions_gen.go` (4), `F/terms/terms.json` (6), `F/verify/verify.go` (1), `F/view/view.go` (2), `scripts/agentgen/src/blue-researcher.md` (1), `scripts/agentgen/src/fragments/lens-duties.md` (1), `scripts/agentgen/src/fragments/log-duty-red.md` (1), `scripts/agentgen/src/lead-judge.md` (3), `scripts/agentgen/src/red-chair.md` (4), `scripts/agentgen/src/red-lens-voice.md` (1).
- (9) answers — `GetAnswers\(\)|\.Answers\b|--answers|flags\.Answers` → 302 hits / 82 files (32 non-test, 1 migrate) → §III.B1, §III.B2: each reader of `answers` widens to `Mint`, `Regrade` (F-GRADE, 2026-11-24), `Inquiry`, `QuestionPart` where it enumerates bodies; the rest O. Non-test: `plugins/frank-exchange-of-views/agents-claude/blue-researcher.md` (8), `plugins/frank-exchange-of-views/agents-claude/blue-synthesizer.md` (8), `plugins/frank-exchange-of-views/agents/blue-researcher.md` (8), `plugins/frank-exchange-of-views/agents/blue-synthesizer.md` (8), `plugins/frank-exchange-of-views/docs/anchor-lifecycle.md` (3), `plugins/frank-exchange-of-views/docs/seat-command-triggers.md` (1), `F/cli/blue/cite.go` (4), `F/cli/blue/compute.go` (4), `F/cli/blue/edit.go` (9), `F/cli/blue/retire.go` (5), `F/flags/names.go` (3), `F/flags/shapes.go` (2), `F/record/benchanswer.go` (2), `F/record/blueanswered.go` (4), `F/record/changesjson.go` (1), `F/record/demand.go` (2), `F/record/estoppel.go` (3), `F/record/gapacts.go` (5), `F/record/impasse.go` (3), `F/record/record.go` (12), `F/record/recordpb/record.proto` (1), `F/record/recordsql/decode_gen.go` (4), `F/record/recordsql/evidencetables.go` (1), `F/record/refs.go` (1), `F/record/standingreading.go` (2), `F/record/tombstone.go` (3), `F/record/windowindex.go` (1), `F/refusalcheck/refusalcheck.go` (1), `F/report/assemble.go` (1), `F/report/edits.go` (1), `F/seatprobe/build.go` (2), `F/view/changes.go` (2).
- (10) tombstone-ends — `rewordedGaps|rewordedAnchorGaps|retiredAnchorGaps|inquiryFateGaps|rewordedUnplacedGaps|tombstoneArms|rulingPlacement` → 29 hits / 3 files (1 non-test, 0 migrate) → X14. Non-test: `F/record/tombstone.go` (23).
- (11) estoppel-seat — `EstoppelConflict|requireNotEstopped|roadEstoppel` → 35 hits / 13 files (7 non-test, 0 migrate) → X11. Non-test: `F/cli/lens/mint.go` (4), `F/record/benchanswer.go` (1), `F/record/docketing.go` (1), `F/record/estoppel.go` (2), `F/record/freeze.go` (4), `F/record/proposal.go` (3), `F/record/viewjson.go` (1).
- (12) about-gap — `ABOUT_KIND_GAP|ABOUT_KIND_SECTION|about-kind gap|aboutGapRefusal|ResolveAbout` → 144 hits / 73 files (7 non-test, 3 migrate) → X10, §III.B7; `ABOUT_KIND_SECTION` readers → §III.B0's section rows (picked 2026-11-25, "Anchor headings"; F-SECTION retired). Non-test: `F/cli/lens/mint.go` (4), `F/record/about.go` (2), `F/record/benchanswer.go` (1), `F/record/docketing.go` (1), `F/record/gapacts.go` (1), `F/record/recordpb/annotations_gen.go` (1), `F/record/recordpb/record.proto` (3).
- (13) supersede-proposal — `Motion_Supersede|motion_supersede|requireBenchSuccessor|actingSeat|mintActingSeat|mintChain|mintActors|requireSupersedable` → 117 hits / 28 files (18 non-test, 0 migrate) → X1, X4; `requireSupersedable` X8 (own open gap only). Non-test: `F/cli/motion/propose.go` (1), `F/record/actjudge.go` (2), `F/record/answersretired.go` (1), `F/record/available.go` (2), `F/record/composedsource.go` (2), `F/record/demand.go` (2), `F/record/estoppel.go` (3), `F/record/failedcheck.go` (1), `F/record/freeze.go` (3), `F/record/inquiry.go` (1), `F/record/inquirycheck.go` (1), `F/record/proposal.go` (20), `F/record/record.go` (2), `F/record/recordsql/decode_gen.go` (11), `F/record/recordsql/gapstate.go` (1), `F/record/standingreading.go` (2), `F/record/supersession.go` (2), `F/record/tombstone.go` (1).
- (14) remand — `requireBenchDocketAnswer|remandLimit|DISPOSITION_REMANDED|flags\.Remand|requireRemand` → 87 hits / 49 files (16 non-test, 4 migrate) → §III.B4 (`lensSitsAfter` on remand); the rest O. Non-test: `F/cli/motion/docketrule.go` (4), `F/record/benchanswer.go` (1), `F/record/benchgrade.go` (1), `F/record/blueanswered.go` (1), `F/record/dispatch.go` (2), `F/record/docketing.go` (4), `F/record/migrate/docketing.go` (1), `F/record/migrate/entries.go` (1), `F/record/record.go` (5), `F/record/recordpb/annotations_gen.go` (1), `F/record/recordpb/record.proto` (1), `F/record/remand.go` (5), `F/record/sitting.go` (4), `F/record/tombstone.go` (1), `F/record/verdict.go` (1), `F/record/viewjson.go` (3).
- (15) answered — `unruled_docket_(motion|filed)|\.Ruled\(\)|rulingWord|motion_answers|motion_rule_docket_answers_a_filing|docketHeld|standingProposal|requireNoStandingProposal|windowEnd|proposalsUnanswered|unansweredProposals` → 139 hits / 50 files → X5 and §III.R's B2 row: each reads an act carrying `accepts = M` as M's answer, where today it reads the accept's `MotionRule` (T-B48). Non-test: `F/cli/motion/verbs.go` (1), `F/dashboard/model.go` (1), `F/graph/graph.go` (3), `F/record/archivedocket.go` (1), `F/record/available.go` (1), `F/record/benchanswer.go` (1), `F/record/benchsitting.go` (4), `F/record/closingowed.go` (2), `F/record/composedsource.go` (1), `F/record/dispatch.go` (5), `F/record/docketing.go` (6), `F/record/freeze.go` (2), `F/record/grademovement.go` (2), `F/record/inquirycheck.go` (2), `F/record/migrate/docketing.go` (1), `F/record/motion.go` (7), `F/record/passblockers.go` (1), `F/record/proposal.go` (1), `F/record/record.go` (8), `F/record/recordsql/gapstate.go` (5), `F/record/recordsql/schema.go` (1), `F/record/recordsql/views.go` (5), `F/record/retirement.go` (1), `F/record/stale.go` (2), `F/record/viewjson.go` (3), `F/report/assemble.go` (1), `F/report/motions.go` (2), `F/scorecard/scorecard.go` (1), `F/verify/verify.go` (1).
- Reverse direction: the 180 non-test files the B commits touched (`git show --name-only` over B1–B6c, generated and test files excluded; `census2/bfiles.txt`) against the union of (1)–(14) (`census2/hitfiles.txt`): 55 not hit. Read for meaning: the help pages `regrade.md`, `close.md` (`--accept M`, the raiser's accept), `edit.md`, `retire.md` (`--accept G`; "a statement standing on a gap holds until its ruling" — §III.B5's wording) are X20; `record/rulerfiles.go` (`filingRoles`) and `record/itemkind.go` (`ItemDocket`) are X6; `record/supersede.go:54,92` ("a lens's accept made the accepted act its own") is X5's text; `report/motions.go`, `record/motionview.go` render a proposal's row and its answer — the answer becomes the raiser's act carrying `accepts` (X5); `record/dispatchnext.go` writes the docketing (§III.B4, kept); `record/archivedocket.go`, `record/replay{,writer}.go`, `record/writeview.go`, `migrate/{eraentries,manifest,registry,remap,replay}.go` carry the built translations and tombstone replay (kept; Translation 5 joins `registry.go`, `manifest.go`); `cli/seat/answer.go` (`answers` on a citation) widens with §III.B1; the rest (`anchor.go`, `fetch.go`, `fetchcache/htmlextract.go`, `flags/{register,repeated}.go`, `planguard/rule.go`, `recordsql/{stmtcache,store,demandlineage}.go`, `queries.go`, `queue.go`, `reading.go`, `oncepersitting.go`, `enumof.go`, `motionsubject.go`, `cli/surface.go`, `cli/seat/{refusal,seat}.go`, `cli/{bench,blue}/command.go`, `dashboard/render.go`, `report/docs.go`, `seatprobe/attempted.go`, `requirements.json`, `enginegen/vocabulary.go`, `grounds.go` (deleted), the help pages `cite.md`, `declare.md`, `outcome.md`, `regrade-bench.md` (deleted), `reproduce.md`, `verify.md`, `lens-audit-and-mint.md`) are O: plumbing, Part A, or text whose sense this spec does not change, re-read in R5's check 6.
- New carriers no pattern finds before they exist: `record/author.go` (`authorOf`), the `--accept G` flags on `retire`, `inquiry record`, `question cover`, `mint`, and the `Closing` statement fields' flags on the closing verbs.


## M27 role literals, each tagged party (P) or seat-class (S) — Part B (audit p10 gap 5, plan §III.B9)

Run on `fb593c3d4`: `git grep -n -E "\"blue\"|'blue'|blue-respond" HEAD -- plugins/frank-exchange-of-views/tools/internal plugins/frank-exchange-of-views/skills/research-protocol/scripts/debate.js plugins/frank-exchange-of-views/skills/research-protocol/scripts/debate.py scripts/agentgen/src ':!*_test.go' ':!*/testdata/*' ':!*.golden' ':!*/record/migrate/*'` (debate.py added at audit p12: +4) → 270 lines (`~/.claude/scratch/1695/p10fix/m27.txt`); it holds every line of the audit's `audit-p10/role-literal.txt` (182 / 43 files) and adds debate.js, help and agentgen. `F` = `T/`. P lines read `authorOf` or are rewritten in R2/R5; S lines are O lines of check 5.

- **P, gap-path party arms:** `F/cli/seat/verbs.go:344,351,355` (`roles: {"blue"}`: the kept-grade item, `engaged`, "a ruling went your way"; widen to a lens as author); `F/record/available.go:60,65` and `F/record/sitting.go:311,316,351,720` (`case "blue"` gap items, remands, `engagedOf`); `F/record/closingowed.go:101,181`; `F/record/dispatch.go:398,467` (engage and await the author); `F/record/benchanswer.go:213`; `F/record/docketing.go:309` (X2), `:504`; `F/record/gapacts.go:148,173,197`; `F/record/impasse.go:254,321,326`; `F/record/viewjson.go:1346,1573,1892`; `F/report/assemble.go:1231,1239`; `F/view/view.go:615,622`; `F/record/record.go:1907`; `F/record/proposal.go:314`, `F/record/recordsql/gapstate.go:221` (`sided_with` words, X16); `F/record/freeze.go:169` (X3); `F/record/successor.go` (all 8, X8's engage-the-successor road); `F/seatprobe/boards.go:383` (the filer is the author), `:870` (a proposal board: kept, re-keyed); `F/capture/capture.go:618`, `F/cli/blue/closing.go:19` (comments); help `closing-blue.md:3,7`, `closing-red.md:7` (then `help_gen.go`), agentgen `lead-judge.md:41`.
- **S, seat-class (O):** the roster and its readers (`F/seatclass/{roster,seatclass}.go`, `F/record/roles.go:89`, `F/record/cast.go:33`, `F/record/dispatch.go:1381,1382,1388`, debate.js:96-98,392); who may sit, source or write a part (`F/cli/seat/source.go:31`, `F/record/actjudge.go:85,133`, `F/record/record.go:1464`, `F/record/envelope.go:182`, `F/cli/root.go:75`, `F/cli/surface.go:267`, `F/cli/blue/{command,ingest}.go`, `F/capture/capture.go:593`, `F/seatprobe/{seatprobe,production,build}.go`, `F/report/layout.go:49`, `F/scorecard/scorecard.go`, `F/dashboard/*`, `F/diagnostics/tokens.go:117`); `blue-respond`'s own report duties, not on a gap (inquiry, question, failed check: `F/record/dispatch.go:660,665,695,705,768-777,1547`, `F/record/sitting.go:355,391,667`, `F/record/questionpart.go:228-288`, `F/record/inquirycheck.go:164`, `F/record/bluesittings.go`); the held-out epoch, which holds out blue's seat (`F/record/dispatch.go:366,813-816`, debate.js `closingsOnly` :898,1025, §III.B4's stated reach); the dispatch by seat class (debate.js `roleOfParty` :999-1002, `bluePrompt` :920, :1031,1058); the chair/blue counterparty pairing (`F/record/viewjson.go:1506`); the remaining `F/seatprobe/boards.go` boards seated on `blue-respond`; `F/cli/seat/verbs.go:338,339,346,350,373` (audiences already including a lens, or blue's own `edit`); the 45 `roles` arrays of `F/terms/terms.json` (who reads a term); history comments (`F/cli/blue/compute.go:81`, `F/cli/seat/seat.go:810`, `F/hookgate/hookgate.go:91`, `F/reportproj/{plan,splice}.go`, `F/seatenv/seatenv.go:9`, `F/sittingwrite/worklist.go:60`).
- Check 5's role pattern is this command; a new hit is tagged here before the phase is green.
- `skills/research-protocol/scripts/debate.py` (Antigravity's engine) joins the pathspec (audit p12): 4 hits, `:47,60,61` (enginegen regions, `-check`) and `:325` (`role_of_seat(...) == "blue"`, dispatch by seat class), all S (O).

## M28 readers of `BlueEdit.reopened`, which gains a section gap's token — Part B (plan §III.B0 sections)

Run on `fb593c3d4`: `git grep -n -E 'GetReopened\(\)|ReopenedAnchors\(' HEAD -- plugins/frank-exchange-of-views ':!*_test.go' ':!*.pb.go'` → the writer `F/bluedoc/bluedoc.go:321` (lists the token for an edit meeting its section), its callers `F/cli/blue/edit.go:179,432`, and six readers: `F/record/anchorchecks.go:262`, `F/record/gapacts.go:250` (the turn), `F/record/tombstone.go:141,156,171` (the end arms, X14's one arm), `F/record/migrate/gapplace.go:166` (replay; Translation 6). Each reads a section token as it reads a sentence's; `anchorchecks.go:262` checks evidence anchors and skips a gap token (O).

## M29 the refusal of an anchor on a heading, which admits a section gap's token — Part B (plan §III.B0 sections)

Run on `fb593c3d4`: `git grep -n -E 'ErrOnHeading|headingAnchors|anchor\.Heading\b' HEAD -- plugins/frank-exchange-of-views/tools/internal ':!*_test.go'` → enforcers that admit the token: `F/anchortext/anchortext.go:583` (`Attach`), `:635` (`AttachToSentenceOf`; the section placement is its own call, so this stays), `F/reportproj/plan.go:72` (`PlanSplice`: an anchor new on a heading; the token placed at the mint is on it before any edit, so the carry rule already admits it once placed), `F/bluedoc/bluedoc.go:467` (an anchor's relocation across an edit skips a heading: it admits the section token, or a heading's rewrite would move the token off it); readers that skip a heading and stay (O): `F/bluedoc/bluedoc.go:373` (a heading is no claim), `F/claimcount/claimcount.go:79`, `F/anchortext/anchortext.go:496` (the duplicate-quote locator's section label), the refusal's wording for every other anchor (`F/anchortext/anchortext.go:439-446,528-541,568,603`, `F/cli/lens/mint.go:260,435`); `F/reportheading/reportheading.go:30` (`HeadingKey` strips anchors, so a key ignores the token).

## M30 what a party may do on a docketed gap, which "Either party concedes" widens — Part B (audit p11 gap 5, plan §III.B9)

Run on `11067fa08`: `git grep -n -i -E "until (its|the) ruling|until the bench rules|bench's to dispose|off your list|close is refused while|until it rules" HEAD -- plugins/frank-exchange-of-views scripts/agentgen ':!*_test.go' ':!*.golden' ':!*help_gen.go'` (the audit's `audit-p11/close-under-docket.txt`, widened by "until it rules"; generated `agents/`, `agents-claude/` and `help_gen.go` lines follow their sources). Each P line names each party's conceding act — the raiser's close of G or `--accept M` of D's proposal, the author's `--accept G` — or, where it states a refusal, the concession the refused seat has instead.

- **P, help and seat text:** `F/cli/seat/verbs.go:346` (`show work`: "A DOCKETED GAP IS OFF YOUR LIST UNTIL THE BENCH RULES IT … a lens's close is refused"); `F/cli/seat/help/close.md:9`, `edit.md:52`, `retire.md:19` ("A STATEMENT STANDING ON A GAP HOLDS UNTIL ITS RULING"); `scripts/agentgen/src/blue-researcher.md:29` ("nothing else until the bench rules it"); `scripts/agentgen/src/fragments/lens-duties.md:9` ("ONCE A GAP OF YOURS IS DOCKETED IT IS THE BENCH'S: you may not close it").
- Every source line of the run is tagged above. The audit's file also holds "the record cannot close" hits (`debate.js:406`, `capture.go:667`, `impasse.go`, …): another sense, a sitting the record cannot close, and O.
- **P, refusals:** `F/record/record.go:1819` (`requireNoStandingDocket`: names the close as the concession, refused only under grade movement); `F/record/supersession.go:146`; `F/cli/blue/freeze.go:54` and `F/record/freeze.go:231` ("holds until its ruling": the statement holds until its ruling or a concession); `F/record/passblockers.go:533` (PASS refused until the bench rules or a party concedes).
- **P, comments:** `F/cli/blue/freeze.go:13`, `F/record/freeze.go:18,244`, `F/record/docketing.go:375` (L27), `F/record/passblockers.go:544`, `F/record/dispatch.go:531`, `F/record/closingowed.go:43`, `F/record/composedsource.go:66`, `F/record/sitting.go:478,483`, `F/record/viewjson.go:1049,1344,1424`.
- **O:** grade movement, where no party concedes: `F/record/passblockers.go:507`, `F/record/dispatch.go:575`, `F/record/grademovement.go:43`, `lead-judge.md`'s grade-movement paragraph; `F/cli/seat/help/dispatch.md:21` and `F/record/viewjson.go:1448` (the supersede proposal: X1, X20); `outcome.md:11` (the bench's own gavel).
- **Widened** (audit p12 gap 5; check 5 runs both, each re-run on `3ef3a2eae` with the counts below): (a) `git grep -n -i -E "docketed|docketing stands|before the bench" HEAD -- F/cli/seat/help F/cli/seat/verbs.go scripts/agentgen/src plugins/frank-exchange-of-views/skills/research-protocol/scripts/debate.js F/terms/terms.json F/seatprobe ':!*_test.go'` → 58 lines (`audit-p12/m30wide.txt`); (b) `git grep -n -i -E "may not close|cannot close|not close it|asks nothing else|nothing else until|is the bench's|the bench's to|has ruled it|until (the bench|it) (has )?rule" HEAD -- plugins/frank-exchange-of-views scripts/agentgen ':!*_test.go' ':!*.golden' ':!*help_gen.go' ':!*/agents/*' ':!*/agents-claude/*' ':!*/record/migrate/*'` → 105 lines (`audit-p12/m30wider.txt`). `F` = `plugins/frank-exchange-of-views/tools/internal`; every line of both is tagged below or above. P beyond the lines above: `debate.js:893` (`closingClause`, rendered into `lensPrompt` `:906` and `bluePrompt` `:920`: "the gap asks nothing else of you until the bench has ruled it"); `scripts/agentgen/src/lead-judge.md:84` ("A lens may not close a gap once it is docketed"); `debate.js:930` ("Every docketing gets a written ruling": a conceded one gets none); `help/closing-red.md:3,9`, `closing-blue.md:3,9` (the closing, or the concession instead; a restatement); `terms.json:899` (`docketing`); `F/record/record.go:1830`; `F/record/closingowed.go:43`, `F/record/composedsource.go:67`; X5's and X6's texts `F/record/docketing.go:360`, `F/record/benchopinion.go:26`, `F/seatprobe/boards.go:499,664,996,998`. O: the bench's own acts and voice ("is the bench's" as owner of a verb, a holding, an account: `capture.go:1538`, `cli/bench/*`, `verbs.go:46,353,354,549`, `graph.go:88`, `dispatch.go:937,1119,1365,1372,1568`, `petitionorder.go:18`, `queries.go:95`, `record.proto`/`record.pb.go` comments, `replay.go:163,169`, `viewjson.go:886,1736`, `assemble.go:674,1333,1385`, `scorecard.go:333`, `delivery.go:393`, `view.go:465`, `available.go:137`, `benchanswer.go:58`, `docketing.go:33,388`, `proposal.go:270,301`, `actjudge.go:159`, `passblockers.go:498,1012,1020`, `seat-command-triggers.md:157`); "a sitting the record cannot close" (`debate.js:406`, `debate.test.mjs:395`, `capture.go:667`, `bluesittings.go:17`, `dispatch.go:43,90`, `gapacts.go:241`, `impasse.go:44,102,104,124,383,392`, `onegapperclass.go:299`, `positionowed.go:58`, `proposal.go:247`, `hookgate/substitution.go:279`); the live-answer demand (`demand.go:163`, `sitting.go:337`, `boards.go:220`, `help/{close,mint,reproduce,verify}.md`'s demand line, `blue-search-and-cite.md:55`, `lens-audit-and-mint.md:8`); `closure.go:42` (remanded closes nothing); grade movement (`grademovement.go:43`, `passblockers.go:507`, `lead-judge.md:88`); the docketing's mechanics and schema (`debate.js:384,393,404,412,414,487,560,878,899,963,972,1060`, `verbs.go:215,302,352`, `help/dispatch.md:17,23,27`, `help/position.md:7`, `help/mint.md:11`, `red-chair.md:15`, `lens-duties.md:5`, `seatprobe/{build,production}.go`, `boards.go:256,354,373,383,449,491,629,733,739`); `boards.go:796-834` (`lensClosing`: the closing as owed, not as the only act). `debate.js:1134,1143,1145-1148` are M31's.

## M31 the docket order: every `Docketing` writer, and every reader that holds on a standing or undocketed cause — Part B (audit p12 gap 1, plan §III.B4)

Writers, run on `a995a886f`: `git grep -n -E 'recordpb\.Docketing\{|DocketingCause_DOCKETING_CAUSE_(IMPASSE|PROPOSAL|GRADE_MOVEMENT)(\.Enum\(\))?,|toBench\(g|docket\(docketingEntry' HEAD -- 'plugins/frank-exchange-of-views/tools/*.go' ':!*_test.go'` (`audit-p12/writers.txt`): the dispatch collects at `F/record/dispatch.go:462` (`toBench`, called at `:568,570,575`) and `:501` (`onProposal`) and writes at `F/record/dispatchnext.go:121`; the exit register writes at `F/record/docketing.go:189` (`docketStandingProposals`, called at `F/record/record.go:310`, L29); migrate at `F/record/migrate/docketing.go:180`, its cause from `F/record/archivedocket.go:27,29`; `recordsql/decode_gen.go` is generated (O). Each writes a cause when first computed and stays as built: the docket order is their rows' order.

Readers, run on `a995a886f`: `git grep -n -E 'docketingsAt\(|standingDocketings\(|standing_docketing\b|docketHeld\(|benchHeld\(|docketedBlockers\(|requireBenchMotionsRuled\(|benchBlockers|routeBench\b|docketStandingProposals\(|benchSatFor\(|requireBenchHeld\(' HEAD -- plugins/frank-exchange-of-views ':!*_test.go' ':!*.golden' ':!*/testdata/*'` → 23 files.

- **P, the docket's first:** `F/record/benchsitting.go:25,48,71,85,97` (`benchHeld` holds only the docket's first unruled in every arm; its comment `:76-80` gains the order); `F/record/proposal.go:274,293` (`requireBenchHeld`'s refusal names the first); `F/record/dispatch.go:529,535` (the routeBench arm convenes the bench on the ready run from the docket's first); `F/record/actjudge.go:141,160`; `debate.js:1145-1148` (`benchBlockers` and the terminal prompt: rule in the docket's order), `:1134,1143`; `F/record/benchanswer.go:138` (`positionsAt`: the latest closing; the restore reads D's own ruling's bench acts).
- **P, texts that say "rule each" with no order:** `F/record/passblockers.go:1043-1045` (`requireBenchMotionsRuled`'s refusal: rule the docket's first before the rest, its sides `--for-raiser | --for-author`, X16).
- **O, held per gap and unchanged by the order:** `F/record/passblockers.go:873,881,883,1026` (every standing docketing holds the PASS; `requireBenchMotionsRuled`'s check as built); `F/record/closingowed.go:197` (closings owed on each); `F/record/freeze.go:270` (the freeze from each docketing's row, F-FREEZE-ALL); `F/record/docketing.go:20,80,88,133,134,150,177,270,461`; `F/record/grademovement.go:99`; `F/record/remand.go:299,345`; `F/record/composedsource.go:71`; `F/record/viewjson.go:995,1139` and `F/record/available.go:114,296` (a gap's route and pending, per gap); `F/record/sitting.go:559`; `F/record/stale.go:118`; `F/record/inquirycheck.go:148`; `F/record/retirement.go:246`; `F/record/recordsql/{gapstate.go:64,352,views.go:415,466}`; `F/record/archivedocket.go:33`; `F/record/dispatch.go:209,717,974,978,1626,1645,1652`; `F/record/record.go:310`.
- Deleted by the docket order: nothing a writer holds. Deleted with it: `cli/blue/freeze.go`'s `BenchRoadsHold` and `cli/motion/docketrule.go:215,230`'s uses (X21).

## M32 every reader of a gap's wording, classified by the version it reads — Part B (ruled 2026-11-26 "AMEND", plan §III.B7)

Run on `adc7262d4` (`~/.claude/scratch/1695/amendv/wording-readers.md`, raw `raw.txt`): `git grep -n -E 'GetProblem\(\)|GetRequiredFix\(\)|GetAcceptanceCheck\(\)|GetFixNew\(\)|GetFixBasis\(\)|\.(Problem|RequiredFix|AcceptanceCheck|FixNew|FixBasis)\b|"problem"|"required_fix"|"acceptance_check"|"fix_new"|"fix_basis"' HEAD -- F/tools/internal ':!*_test.go' ':!*/testdata/*' ':!*.pb.go'`, each site read. `Gap.Mint` is the first `Mint` body (`F/record/gapstates.go:87-102`), so a `g.Mint.Get…()` read is D, as `g.Mint.GetLikelihood()` is beside the regraded `g.Likelihood`.

- **V, through the gap view — reads `current_*`:** `F/record/viewjson.go` (GapJSON `:393-441`, the work gap `:1137-1175,1335-1353`, `locateGap` `:2076`); `F/record/sittingwrite/delivery.go:291-295`; `F/report/assemble.go:718,723` (risk matrix); `F/record/onegapperclass.go:214`; and the gap view's own projection, `F/record/recordsql/views.go:430-436` (`m."problem"` …, which becomes `current_*`; audit p14's re-run). 10 sites, 5 files.
- **D, switched to `Gap.Current` / `current_*`:** `F/record/estoppel.go:180` (`prescriptionAt` → `edit --accept`, `Proposal`, `proposalAppliedVerbatimAt`), `:140` (`DeclineStatsOf`: a verified fix on any version); `F/record/tombstone.go:226` (`verbatimPlacement`); `F/record/changesjson.go:104-118` and `F/view/changes.go:122-150` (`show changes --id`: every version, `wordingAt`); `F/record/composedsource.go:92-105` (and its `:28` comment); `F/report/assemble.go:914-921,1098-1112`; `F/view/view.go:287-288,294,298,317`; `F/cli/blue/edit.go:660`; `F/record/onegapperclass.go:256`. 15 sites, 9 files.
- **D, kept mint-time:** `F/cli/lens/mint.go:69-73,390-393` (`RepeatsKeyed`: a keyed retry restates the mint). 2 sites, 1 file.
- **P, a past moment:** `F/record/estoppel.go:70-71` (`appliedVerbatim`: `wordingAt` the edit), `:103-106` (`EstoppelConflict`: the placing act's text, §III.B6); `F/report/assemble.go:948,1035-1036` and `F/view/view.go:355` (closed gaps: current = in force at the close, since an amend needs an open gap). 5 sites, 3 files.
- **W:** `F/cli/lens/mint.go:97,175-196,306,334-335`; `F/record/redvoice.go:51-53`; `F/record/required_gen.go:101-117` (regenerated for `Amend`); `F/cli/motion/propose.go:106,249`. 7 sites, 4 files.
- **O, the same names on other bodies** (`Closing`, `Road`, a proposal's stored act): `F/record/freeze.go:323,333`, `F/record/docketing.go:502`, `F/cli/lens/closing.go`, `F/record/benchanswer.go:349-398`, `F/cli/motion/docketrule.go:220-228`, `F/cli/blue/freeze.go`, `F/record/viewjson.go:1420-1423`, `F/cli/motion/propose.go:153`, seatprobe's fixtures (`F/seatprobe/build.go:130-134` included), `F/flags/names.go`. 11 sites.
- **Migrate:** `F/record/migrate/remap.go:124,138` (field sets; no value read). debate.js, dashboard, graph, verify, capture, consistency and scorecard read no wording.

## M33 the supersede chain, every carrier deleted, in both directions and both trees — Part B (ruled 2026-11-26 "AMEND", plan X22)

Trees: branch `adc7262d4`, `origin/main` `af3e6ccbe` (`main` merges into R1, P12b with it, so its chain is deleted in the merged tree). Outputs in `~/.claude/scratch/1695/amendv/`.

1. **Forward, by symbol** (`m33-sym-<sha>.txt`): `git grep -n -E 'supersessionsOf|successorClosing|SupersededGap|supersededRefusal|requireSupersedable|requireSupersededOffDocket|requireNotSuperseded|successorsOf|successorsOwed|OwedSuccessor|requireEngage|EngagementOwesNone|isSuccessorOfAny|blueEngagedThisEpoch|GetSupersedes\(\)|AncestorID|dropCrossSeatLink|mintChain|actingSeat|mintActors|mintActingSeat|demand_lineage|demandReachSQL|supersede_event|superseded_by|mint_ancestor|mint_supersedes|SuccessorsOwed|successors_owed|newEngage' <sha> -- F ':!*_test.go' ':!*/testdata/*' ':!*.pb.go' ':!*_gen.go' ':!*.golden' ':!*/agents/*' ':!*/agents-claude/*'` → branch 225 lines / 54 files, `main` 179 / 45. Largest: `record/successor.go` (41 | 34), `record/supersession.go` (25 | 25), `record/proposal.go` (18 | —), `recordsql/gapstate.go` (16 | 20), `record/sitting.go` (10 | 10), `recordsql/demandlineage.go` (10 | 12), `migrate/replay.go` (9 | 5), `recordsql/views.go` (8 | 9).
2. **Forward, by sense** (`sweep-<sha>.txt`, `cls-<sha>.txt`): `git grep -n -i -E 'supersed|successor|lineage|ancestor|chain tip|chain.s origin|from the chain' <sha> -- F scripts/agentgen`, each hit classed authored, test, testdata, generated; authored 1299 | 947 lines; tests 154 files / 1365 hits | 130 / 993; generated carriers (agents, `record.pb.go`, `decode_gen.go`, `help_gen.go`, `required_gen.go`, `annotations_gen.go`, `definitions_gen.go`, `docs/vocabulary.md`, the two SKILL.md) regenerate.
3. **Backward, by declaration** (`decl-<sha>.tsv` from `declmap.bin`; `callers-*.tsv`): each hit's enclosing declaration and each chain declaration's callers. All five groups tagged (audit p14; rubric `amendv/RUBRIC.md`, senses `amend-senses.txt`), pasted below: 790 rows (531 branch, 259 `main`) plus group W. Lines deleted or rewritten (every tag but O and the engagement kept): **1,902 on the branch, 835 on `main`**. Change per tag: C-core → deleted (X22); C-proposal → deleted (X1, X3, X4); C-regression → deleted (X8); C-text → rewritten (X20); MIXED → its chain arm deleted, the declaration kept (X22); C-engage and every engage part of another row → the engagement kept and restated, only the chain input (`successorsOf`'s `supersedes` and `successor` arms, `isSuccessorOfAny`) and the successor wording going (X22a, ruled "Keep, generalized"); O → kept, its sense in the note. Whole chain files deleted: branch 298 lines (`supersession.go` 161, `demandlineage.go` 72, `flags/ancestor.go` 30, `migrate/lineage.go` 35), `main` 203; `successor.go` (262 | 257) and `cli/blue/engage.go` (68 | 68) stay as X22a's engagement, losing their chain input and wording. (1)'s pattern's `requireEngage|EngagementOwesNone|blueEngagedThisEpoch|newEngage` hits are X22a's: O at check 5. A MIXED or C-text row's X row follows the part its note names (X8 for a regression part, X1/X3/X4 for a proposal part, X22a for an engage part).
- **Declarations that lose only their chain arm:** listed in plan X22.
- **O, other senses kept:** `Retire.superseded_by`; `QuestionPart.supersedes_coverage`; `Inquiry.supersedes_status`; `EVENT_TYPE_RETIRE`'s "what supersedes it"; `record/supersede.go`'s superseder table (one answer superseding a refused repeat; its rows `:82,84,92` teaching `--supersedes` are rewritten); the declarations fold; the dispatch's `engage` rows; capture's "ancestors" (directories); a dispatch's "lineage none".
- **Replaces** M12 and M13 (their carriers are now deleted, not rewritten) and M22's lineage rows.
4. **Widened sweep** (check 5, O lines allowed: `\.Supersedes\b` also hits `SupersedesCoverage`/`SupersedesStatus`, which are O): `git grep -n -E 'repair_regression|lineage_mints|EdgeDeltas|edge_deltas|"supersedes"|\.Supersedes\b' <sha> -- F scripts/agentgen` with (1)'s exclusions → branch 55 lines, `main` 32; the 11 | 15 in no tag row are group W.

**Group A** (`amendv/tag-A.tsv`)

| t | file:lines | declaration | tag | del | change | note |
|---|---|---|---|---|---|---|
| B | `T/record/actjudge.go:43-80` | requireClosingsOnlySitting | C-text | 2 | rewritten · X20; its engage part kept, restated · X22a | C-engage: "blue's engage row are no acts" + the DISPATCH arm that admits blue's engage row; L44' closings-only gate stays |
| B | `T/record/actjudge.go:98-187` | judgeAct | MIXED | 10 | chain arm deleted · X22; its engage part kept, restated · X22a | mintActingSeat (bench-written successor's acting seat, L104-109) + Dispatch arm requireEngageNamesEveryOwed (C-engage); act dispatcher stays |
| B | `T/record/actjudge.go:189-218` | judgeMint | MIXED | 12 | chain arm deleted · X22 | except param + requireSupersededOffDocket + requireSuccessorCarriesDemand arms and doc lines; anchor/reading/one-gap/voice checks stay |
| B | `T/record/actjudge.go:24` | - | C-text | 1 | rewritten · X20 | package refusal list names supersession of another lens's gap and successor-at-another-check-kind refusals |
| B | `T/record/answersretired.go:24-55` | answersRetiredItems | C-text | 4 | rewritten · X20 | supersededBySeatAfter skip + "mint it, naming G as the gap it supersedes" teaching; answers-retired news item stays |
| B | `T/record/answersretired.go:57-66` | supersededBySeatAfter | C-core | 10 | deleted · X22 | whether seat minted a successor naming gap in supersedes (mintActors chain) |
| B | `T/record/answersretired.go:21` | - | C-text | 1 | rewritten · X20 | doc bullet: lens minted a gap naming it in supersedes |
| B | `T/record/available.go:166-171` | assigneeOf | MIXED | 2 | chain arm deleted · X22 | assignee via mintChain (chain origin's minter); map stays, becomes the mint's seat |
| B | `T/record/available.go:260-270` | latestDispatchOf | C-text | 1 | rewritten · X20; its engage part kept, restated · X22a | C-engage: comment "or an engage blue wrote for its own next sitting"; latest-dispatch reader stays |
| B | `T/record/benchanswer.go:26-61` | Positions | C-text | 2 | rewritten · X20 | C-proposal: Act comment (Mint = supersede proposal) and TextSide "or a supersede on open G"; docketing positions stay |
| B | `T/record/benchanswer.go:130-235` | positionsAt | MIXED | 2 | chain arm deleted · X22 | C-proposal: *recordpb.Mint in the textual case + comment "or a supersede on open G"; positions fold stays |
| B | `T/record/benchanswer.go:251-272` | (Positions).AnswerWord | C-text | 1 | rewritten · X20 | C-proposal: doc "a successor" statement writes no word |
| B | `T/record/benchanswer.go:431-437` | (Positions).wordRefusal | C-text | 1 | rewritten · X20 | C-proposal: refusal text "a successor" |
| B | `T/record/benchanswer.go:19` | - | C-text | 1 | rewritten · X20 | C-proposal: file doc "statement against a supersede of a closed G" |
| B | `T/record/benchgrade.go:58-65` | axisGrade | C-text | 2 | rewritten · X20 | comments: lock inherited from an ancestor/successor; type stays |
| B | `T/record/benchgrade.go:67-184` | benchSetAxesOf | MIXED | 10 | chain arm deleted · X22 | lineage lock block L92-99 + tombstone "wherever the lineage carried it" (all-gaps sweep could narrow to the gap); axis fold stays |
| B | `T/record/benchgrade.go:241-262` | requireSuccessorKeepsBenchGrades | C-core | 22 | deleted · X22 | refuses a successor moving a ruled-locked grade of its ancestor |
| B | `T/record/benchgrade.go:25` | - | C-text | 1 | rewritten · X20 | file doc: mint superseding the gap refused, lock follows lineage |
| B | `T/record/benchgrade.go:26` | - | C-text | 1 | rewritten · X20 | file doc: successor carries the ruled grade |
| B | `T/record/benchgrade.go:31` | - | C-text | 1 | rewritten · X20 | file doc: tombstone ends the successor's inherited lock |
| B | `T/record/benchsitting.go:9-54` | docketHeld | O | 0 | O | stand in for: stream positions substitute for events.id |
| B | `T/record/blueanswered.go:12-114` | blueAnswers | O | 0 | O | engaged: chair's dispatch engages blue on gaps (blue's sitting engaged on a gap), not successors |
| B | `T/record/blueanswered.go:116-141` | blueAnswer | O | 0 | O | engaged/pending: a blue sitting the chair's dispatch engaged on the gap |
| B | `T/record/blueanswered.go:166-168` | (blueAnswer).answered | O | 0 | O | closed sitting engaged on it: chair dispatch engagement |
| B | `T/record/blueanswered.go:170-282` | blueAnswersOf | O | 0 | O | s.Engaged = dispatch engagement; SupersedesCoverage: a question part record supersedes the part's earlier coverage |
| B | `T/record/bluesittings.go:7-23` | BlueSitting | C-text | 1 | rewritten · X20 | FoundClosed comment "or the mint that closed it by superseding it"; Engaged/FoundClosed stay (work list found_closed) |
| B | `T/record/bluesittings.go:25-66` | BlueSittings | O | 0 | O | engagementOf call/Engaged: chair dispatch engagement of blue |
| B | `T/record/bluesittings.go:68-95` | engagementOf | MIXED | 5 | chain arm deleted · X22 | supersessionsOf map + Mint arm closing the ancestor by supersession (L76,81-84); dispatch found-closed/open split stays |
| B | `T/record/bluesittings.go:97-102` | FoundClosure | C-text | 1 | rewritten · X20 | comment "or the successor's mint that superseded it"; found-closed gap (work list found_closed) stays |
| B | `T/record/bluesittings.go:104-110` | (FoundClosure).Fate | MIXED | 2 | chain arm deleted · X22 | ClosedSuperseded fallback for a supersession-closed gap; close's class stays |
| B | `T/record/closingowed.go:125-149` | closingReadsOf | MIXED | 7 | chain arm deleted · X22 | supersessionsOf block (mint closes ancestor) L139-145; also assigneeOf (chain origin); closing reads stay |
| B | `T/record/composedsource.go:47-110` | composedSourcesOf | MIXED | 3 | chain arm deleted · X22 | mintActors acting seat (bench-written successor = assignee's text) L82-84; composed-source fold stays |
| B | `T/record/composedsource.go:28` | - | C-text | 1 | rewritten · X20 | C-proposal: "problem and fix change only by a successor" |
| B | `T/record/composedsource.go:29` | - | C-text | 1 | rewritten · X20 | C-proposal: proposed supersede carrying reworded text |
| B | `T/record/composedsource.go:30` | - | C-text | 1 | rewritten · X20 | C-proposal: voice lens's own gap it supersedes itself |
| B | `T/record/declarations.go:19` | - | O | 0 | O | a later declaration supersedes an earlier one (declaration fold) |
| B | `T/record/demand.go:14-29` | EvidenceDemand | C-text | 1 | rewritten · X20 | or an ancestor it carries the demand from; demand type stays |
| B | `T/record/demand.go:31-35` | Reach | MIXED | 2 | chain arm deleted · X22 | From field + comment exist only for ancestor reach (demand_reach.from_gap); Anchor/Seat stay |
| B | `T/record/demand.go:54-102` | evidenceDemandAt | MIXED | 2 | chain arm deleted · X22 | from_gap column + ORDER BY from_gap<>gap_id (demand_reach over demand_lineage); demand read stays |
| B | `T/record/demand.go:116-148` | DemandRefusal | C-text | 4 | rewritten · X20 | r.From!=gapID ancestor branch (3) + "a close to a successor" (C-regression wording); refusal stays |
| B | `T/record/demand.go:150-157` | appendNew | C-core | 8 | deleted · X22 | dedup exists only because gap and ancestor reach one anchor; DemandRefusal's two calls become append |
| B | `T/record/demand.go:166-184` | SupersessionAdmitted | C-core | 19 | deleted · X22 | pre-write supersedes checks (requireSupersedable, requireSuccessorCarriesDemand) |
| B | `T/record/demand.go:186-190` | awaitingAncestor | C-core | 5 | deleted · X22 | ancestor awaiting live answer |
| B | `T/record/demand.go:192-214` | requireSuccessorCarriesDemand | C-core | 23 | deleted · X22 | demand_lineage: successor must carry ancestor's check kind |
| B | `T/record/demand.go:216-244` | successorDemandRefusal | C-core | 29 | deleted · X22 | refusal for the above |
| B | `T/record/demand.go:159` | - | C-core | 1 | deleted · X22 | section doc THE DEMAND MOVES TO THE SUCCESSOR |
| B | `T/record/demand.go:162` | - | C-core | 1 | deleted · X22 | same section doc |
| B | `T/record/demand.go:163` | - | C-core | 1 | deleted · X22 | same section doc |
| B | `T/record/demand.go:164` | - | C-core | 1 | deleted · X22 | same section doc |
| B | `T/record/dispatch.go:16-22` | Party | O | 0 | O | engages/engaged: chair dispatch engages a party on gaps |
| B | `T/record/dispatch.go:41-50` | ExchangeCount | O | 0 | O | engaged sittings: dispatch engagement |
| B | `T/record/dispatch.go:60-141` | Plan | O | 0 | O | engage/engaged: chair dispatch engages parties |
| B | `T/record/dispatch.go:325-869` | foldPlan | C-text | 1 | rewritten · X20 | C-proposal: L709 "(F8: a proposed successor of a closed gap)"; all other hits are the engage/engageClosing closure = chair dispatch (O); FLAG: closed-gap docketing arm may lose its only case |
| B | `T/record/dispatch.go:985-994` | dispatchRow | MIXED | 3 | chain arm deleted · X22; its engage part kept, restated · X22a | C-engage: engage field (blue-written engage row); "engaged" in doc is O |
| B | `T/record/dispatch.go:996-1023` | dispatchLedger | MIXED | 1 | chain arm deleted · X22; its engage part kept, restated · X22a | C-engage: engage: isEngageRow(e) in the row literal; ledger stays |
| B | `T/record/dispatch.go:1034-1043` | sittingFor | O | 0 | O | benchSatFor named in doc: bench sat for a docketing (no chain sense) |
| B | `T/record/dispatch.go:1163-1205` | sittingCloser | O | 0 | O | engaged gaps: chair dispatch engagement |
| B | `T/record/dispatch.go:1306-1311` | PartyRow | O | 0 | O | engaged the party on: chair dispatch |
| B | `T/record/dispatch.go:1313-1355` | DispatchGroups | MIXED | 4 | chain arm deleted · X22; its engage part kept, restated · X22a | C-engage: engage row opens no group (3 comment lines + !d.engage); grouping stays |
| B | `T/record/dispatch.go:1497-1519` | owedSitting | MIXED | 4 | chain arm deleted · X22; its engage part kept, restated · X22a | C-engage: engage row owed by the next sitting (L1511-1515); "stand in for" is O; owed-dispatch reader stays |
| B | `T/record/dispatch.go:1593-1601` | requireBenchDocket | O | 0 | O | bench engaged on gaps: chair dispatch |
| B | `T/record/dispatchnext.go:14-28` | DispatchNext | O | 0 | O | engages a party: chair dispatch |
| B | `T/record/dispatchnext.go:39-169` | dispatchNext | O | 0 | O | engages the parties its rows already engage: chair dispatch |
| B | `T/record/docketing.go:452-491` | closingTakesText | MIXED | 2 | chain arm deleted · X22 | C-proposal: act "supersede" in the return + doc "or a supersede a party proposed"; text-road question stays |
| B | `T/record/docketing.go:493-564` | requireClosingStatement | C-text | 2 | rewritten · X20 | C-regression: repaired_with_regression "names the successor" refusal arm on blue's closing |
| B | `T/record/enums.go:160-179` | artifactByClass | C-text | 1 | rewritten · X20 | C-regression: row comment "a live successor carries the remainder" |
| B | `T/record/enums.go:181-194` | ArtifactStateOf | O | 0 | O | walks the lineage: amends_prior's earlier-ruling lineage, not the gap chain |
| B | `T/record/enums.go:218-319` | EnumFields | C-text | 1 | rewritten · X20 | C-regression: Why "the successor invariant that fires on repaired_with_regression" |
| B | `T/record/estoppel.go:77-79` | Estoppel | C-text | 1 | rewritten · X20 | acting seat of that gap's mint (actingSeat): bench-successor authorship; type stays |
| B | `T/record/estoppel.go:118-122` | minterOf | MIXED | 2 | chain arm deleted · X22 | mintActors + doc "assignee where the bench wrote the mint as a successor"; becomes the mint's seat |
| B | `T/record/estoppel.go:35` | - | C-text | 1 | rewritten · X20 | file doc: red mints --supersedes the prescribing gap, as the refusal says |
| B | `T/record/evidenceview.go:348-362` | EvidenceRetireJSON | O | 0 | O | Retire.superseded_by: a retired CLAIM superseded by another claim |
| B | `T/record/evidenceview.go:396-615` | EvidenceJSONOf | O | 0 | O | Retire.superseded_by projected: claim supersession |
| B | `T/record/failedcheck.go:70-126` | failedChecksOf | MIXED | 1 | chain arm deleted · X22 | mintActors acting seat for a mint's seat; failed-check fold stays |
| B | `T/record/failedcheck.go:217-249` | requireFailedCheckAnswered | C-text | 4 | rewritten · X20 | doc lines 221-222 + two refusal sentences teaching --supersedes and demand carry; refusal stays |
| B | `T/record/freeze.go:37-96` | requireSuccessorMintable | C-proposal | 60 | deleted · X1/X3/X4 | L48 successor-mintable check of a supersede proposal |
| B | `T/record/freeze.go:98-102` | successorCheck | C-proposal | 5 | deleted · X1/X3/X4 | error type for the above |
| B | `T/record/freeze.go:104` | (successorCheck).Error | C-proposal | 1 | deleted · X1/X3/X4 | method of successorCheck |
| B | `T/record/freeze.go:105` | (successorCheck).Unwrap | C-proposal | 1 | deleted · X1/X3/X4 | method of successorCheck |
| B | `T/record/freeze.go:107-113` | successorRefusal | C-proposal | 7 | deleted · X1/X3/X4 | filing refusal of an unwritable successor |
| B | `T/record/freeze.go:115-119` | standingSupersede | C-proposal | 5 | deleted · X1/X3/X4 | standing supersede proposal |
| B | `T/record/freeze.go:121-160` | standingSupersedesAt | C-proposal | 40 | deleted · X1/X3/X4 | every standing supersede proposal |
| B | `T/record/freeze.go:162-174` | judgedLanded | C-proposal | 13 | deleted · X1/X3/X4 | which acts the freeze's successor arm judges |
| B | `T/record/freeze.go:176-207` | unwritableSuccessors | C-proposal | 32 | deleted · X1/X3/X4 | pre-write unwritable successors |
| B | `T/record/freeze.go:209-236` | judgeLanded | C-proposal | 28 | deleted · X1/X3/X4 | freeze successor arm L48 |
| B | `T/record/freeze.go:244-254` | Road | MIXED | 3 | chain arm deleted · X22 | C-proposal: Successor field + "or a supersede proposal's stored quote and fix"; closing roads stay |
| B | `T/record/freeze.go:256-259` | (*Landing).StandingRoads | C-text | 2 | rewritten · X20 | C-proposal: doc "each standing supersede proposal's stored quote" |
| B | `T/record/freeze.go:261-326` | standingRoadsAt | MIXED | 12 | chain arm deleted · X22 | C-proposal: standingSupersedesAt loop L315-324; also assigneeOf (chain origin); closing roads stay |
| B | `T/record/freeze.go:21` | - | C-text | 1 | rewritten · X20 | C-proposal: file doc "a standing supersede proposal's successor" |
| B | `T/record/freeze.go:28` | - | C-proposal | 1 | deleted · X1/X3/X4 | L48 bullet THE SUCCESSOR |
| B | `T/record/freeze.go:29` | - | C-proposal | 1 | deleted · X1/X3/X4 | L48 bullet |
| B | `T/record/freeze.go:32` | - | C-proposal | 1 | deleted · X1/X3/X4 | L48 bullet |
| M | `T/record/actjudge.go:36-142` | judgeAct | MIXED | 13 | chain arm deleted · X22; its engage part kept, restated · X22a | main: requireSupersededOffDocket + requireSuccessorCarriesDemand arms (L56-64) + Dispatch requireEngageNamesEveryOwed (C-engage); dispatcher stays |
| M | `T/record/answersretired.go:25-56` | answersRetiredItems | C-text | 4 | rewritten · X20 | main: supersededBySeatAfter skip + "naming G as the gap it supersedes" |
| M | `T/record/answersretired.go:58-66` | supersededBySeatAfter | C-core | 9 | deleted · X22 | main: repeated supersedes (slices.Contains), event seat |
| M | `T/record/available.go:304-365` | regradesAfforded | O | 0 | O | (supersede.go): superseder table, the act that answers a refused repeated act |
| M | `T/record/benchgrade.go:72-79` | axisGrade | C-text | 2 | rewritten · X20 | main: ancestor/successor lock comments |
| M | `T/record/benchgrade.go:81-132` | benchSetAxesOf | MIXED | 9 | chain arm deleted · X22 | main: lineage lock loop over repeated supersedes L94-102 |
| M | `T/record/benchgrade.go:208-230` | requireSuccessorKeepsBenchGrades | C-core | 23 | deleted · X22 | main: repeated-supersedes version |
| M | `T/record/benchsitting.go:9-56` | docketHeld | O | 0 | O | stand in for: positions substitute for events.id |
| M | `T/record/blueanswered.go:11-113` | blueAnswers | O | 0 | O | engaged: chair dispatch engagement of blue |
| M | `T/record/blueanswered.go:169-280` | blueAnswersOf | O | 0 | O | Engaged = dispatch; SupersedesCoverage = question-part coverage supersession |
| M | `T/record/bluesittings.go:68-97` | engagementOf | MIXED | 7 | chain arm deleted · X22 | main: supersessionsOf + Mint arm over repeated supersedes |
| M | `T/record/closingowed.go:104-170` | closingDutiesOf | MIXED | 7 | chain arm deleted · X22 | main: supersessionsOf closedAt block L127-133; closing duties stay |
| M | `T/record/demand.go:116-148` | DemandRefusal | C-text | 4 | rewritten · X20 | main: r.From ancestor branch + "a close to a successor" (C-regression wording) |
| M | `T/record/demand.go:192-215` | requireSuccessorCarriesDemand | C-core | 24 | deleted · X22 | main: loop over repeated supersedes |
| M | `T/record/demand.go:217-245` | successorDemandRefusal | C-core | 29 | deleted · X22 | main: refusal text |
| M | `T/record/dispatch.go:16-22` | Party | O | 0 | O | engages: chair dispatch |
| M | `T/record/dispatch.go:60-143` | Plan | O | 0 | O | engage/engaged: chair dispatch |
| M | `T/record/dispatch.go:301-765` | foldPlan | O | 0 | O | engage closure + engaged: chair dispatch readies parties; no chain text on main |
| M | `T/record/dispatch.go:1452-1483` | benchReadiness | O | 0 | O | engage callback: chair dispatch readies the bench |
| M | `T/record/dispatch.go:1514-1543` | benchSatFor | O | 0 | O | engaged: dispatch engaging the bench/nobody |
| M | `T/record/dispatchnext.go:14-28` | DispatchNext | O | 0 | O | engages a party: chair dispatch |
| M | `T/record/dispatchnext.go:39-162` | dispatchNext | O | 0 | O | engages the parties: chair dispatch |
| M | `T/record/enums.go:160-179` | artifactByClass | C-text | 1 | rewritten · X20 | C-regression: row comment "a live successor carries the remainder" |
| M | `T/record/enums.go:218-320` | EnumFields | C-text | 1 | rewritten · X20 | C-regression: Why "successor invariant ... repaired_with_regression" |

**Group B** (`amendv/tag-B.tsv`)

| t | file:lines | declaration | tag | del | change | note |
|---|---|---|---|---|---|---|
| B | `T/record/gapacts.go:11-43` | gapAct | C-text | 1 | rewritten · X20 | act list names 'a mint superseding it' as an assignee act (l.22-23); gapAct (acts on a gap) stays |
| B | `T/record/gapacts.go:45-93` | gapActsOf | MIXED | 2 | chain arm deleted · X22 | case Mint: on=GetSupersedes arm goes; gapActsOf (every act on a gap, lens turn/exchange) stays |
| B | `T/record/gapacts.go:95-190` | fateActsOf | MIXED | 15 | chain arm deleted · X22 | superseded-gap exclusion goes (115-119 comment, 122 clause, 125 supersessionsOf, 162-165, 181-184); fate acts stay; 152 SupersedesCoverage is O |
| B | `T/record/gapacts.go:204-217` | lensTurnOn | O | 0 | O | engages: the dispatch plan engages the assignee on the gap (not successors) |
| B | `T/record/gapstates.go:77-122` | gapStatesOf | MIXED | 1 | chain arm deleted · X22 | g.SupersededBy = c.supersededBy assignment goes (part of l.117); gap family assembly stays |
| B | `T/record/gapstates.go:124-132` | GapsByID | C-text | 1 | rewritten · X20 | comment 'readers that follow lineage (an ancestor lookup)'; GapsByID is the family's id index, kept |
| B | `T/record/gapstates.go:193-198` | NewFamily | C-text | 1 | rewritten · X20 | comment 'a lineage lookup is a map hit'; NewFamily kept |
| B | `T/record/gapstates.go:200-213` | gapFamilyActs | C-text | 1 | rewritten · X20 | comment names 'superseding mints'; mints stay in the filter as the family's own acts |
| B | `T/record/impasse.go:14-82` | GapExchanges | C-text | 1 | rewritten · X20 | l.30-31 movement clause 'its lineage moved (a mint superseding G)'; 19,44 'engaged' = dispatch engagement (O) |
| B | `T/record/impasse.go:108-127` | partySitting | O | 0 | O | engaged: a dispatch engages a seat's sitting on a gap |
| B | `T/record/impasse.go:129-158` | Exchanges | O | 0 | O | engaged: dispatch events say who was engaged on what |
| B | `T/record/impasse.go:222-264` | gapSittingsOf | O | 0 | O | engaged: dispatch engagement; seat stays engaged after a failed sitting |
| B | `T/record/impasse.go:266-424` | exchangeFold | MIXED | 3 | chain arm deleted · X22 | Mint.supersedes counted as movement on the ancestor (295-297) goes; exchange fold stays |
| B | `T/record/inquiry.go:17-26` | InquiryStatuses | O | 0 | O | successor: a run that queues its own successor run (deferred inquiry), not a gap successor |
| B | `T/record/inquiry.go:93-177` | InquiriesOf | C-text | 1 | rewritten · X20 | comment cites actingSeat (bench-written successor's acting seat); InquiriesOf fold stays |
| B | `T/record/inquiry.go:226-261` | InquiryJSON | O | 0 | O | superseded fate: an inquiry's earlier fate in its history |
| B | `T/record/inquirycheck.go:87-154` | inquiryJudgementsOf | MIXED | 5 | chain arm deleted · X22 | superseding mint as a cause goes (106 comment, 121-123); 131 actingSeatOf -> mint's seat; judgement fold stays |
| B | `T/record/itemkind.go:17-59` | ItemAudit,ItemReaudit,ItemNews,ItemLog,... | MIXED | 3 | chain arm deleted · X22; its engage part kept, restated · X22a | ItemSuccessorOwed const+comment (41-43) goes; 25 'engaged on' = dispatch (O); other kinds stay |
| B | `T/record/itemkind.go:61-71` | ItemKinds | MIXED | 1 | chain arm deleted · X22; its engage part kept, restated · X22a | ItemSuccessorOwed dropped from the returned list (part of l.69) |
| B | `T/record/lensturn.go:9-17` | EarlyLensDispatch | O | 0 | O | engaged a lens: chair's dispatch row engages a lens on a gap |
| B | `T/record/lensturn.go:19-44` | EarlyLensDispatches | C-text | 2 | rewritten · X20 | sentence 23-24: a successor's mint closed the gap, its row engages no turn; 19-20 'engaged' is dispatch (O) |
| B | `T/record/lensturn.go:46-79` | earlyLensDispatchesOf | MIXED | 13 | chain arm deleted · X22 | superseded map (52-61) and skip (70-72) go; early-lens-dispatch audit stays |
| B | `T/record/motion.go:166-217` | Motion | C-text | 1 | rewritten · X20 | Act comment lists 'supersede' among proposed acts; Motion type stays |
| B | `T/record/motion.go:515-544` | requireUnanswered | O | 0 | O | SupersedingAnswer: superseder table, the act that answers a refused repeated act |
| B | `T/record/motion.go:546-571` | RequireUnruledMotion | O | 0 | O | comment cites supersede.go: superseder table (act answering a wrong ruling) |
| B | `T/record/oncepersitting.go:47-80` | requireOncePerSitting | O | 0 | O | superseder table tail (supersede.go) on a once-per-sitting refusal |
| B | `T/record/onegapperclass.go:15-190` | requireOneGapPerClassAndSubject | MIXED | 12 | chain arm deleted · X22 | --supersedes exemption comment (46-48) and exemption block (79-87: ancestor arm C-core + ruling-successor arm C-proposal) go; one-gap-per-class rule stays |
| B | `T/record/onegapperclass.go:232-266` | withRuledGapsOfClassAt | MIXED | 9 | chain arm deleted · X22 | supersede-accept successor arm (241-249, F-J, C-proposal) goes; ruled-gap tombstone join stays |
| B | `T/record/onegapperclass.go:268-295` | oneGapRefusal | C-text | 4 | rewritten · X20 | road text: 'restated by superseding it', successorClause call, '--supersedes %s' advice, 'motion docket file supersede' (proposal) rewritten |
| B | `T/record/onegapperclass.go:320-326` | successorClause | C-proposal | 7 | deleted · X1/X3/X4 | successorClause: names the successor a supersede accept left standing |
| B | `T/record/proposal.go:69-80` | ActWord | MIXED | 2 | chain arm deleted · X22 | Motion_Supersede case goes; ActWord stays for regrade/close |
| B | `T/record/proposal.go:82-83` | ProposedActs | MIXED | 1 | chain arm deleted · X22 | 'supersede' dropped from ProposedActs (part of l.83) |
| B | `T/record/proposal.go:85-94` | ProposedCloseRefusal | C-regression | 10 | deleted · X8 | refuses repaired_with_regression as a proposed close because it names the regression successor (Close.successor) |
| B | `T/record/proposal.go:96-99` | proposedRegressionRefusal | C-regression | 4 | deleted · X8 | refusal sentence: regression successor is a lineage fact no ruling carries; also says 'propose the supersede' |
| B | `T/record/proposal.go:116-162` | requireProposedAct | MIXED | 9 | chain arm deleted · X22 | Close.Successor refusal (141-143, C-regression) and Motion_Supersede arm (154-159, C-proposal) go; regrade/close shape stays |
| B | `T/record/proposal.go:195-244` | requireDocketFiling | MIXED | 17 | chain arm deleted · X22 | --supersedes advice (211), F8 closed-gap-takes-a-supersede (196-197, 225-231), L48 successor-mintable (235-242) go; F1/F5/estoppel/exchange checks stay |
| B | `T/record/proposal.go:301-321` | requireBenchSuccessor | C-proposal | 21 | deleted · X1/X3/X4 | requireBenchSuccessor: the bench's one Mint, the successor a supersede proposal assigned |
| B | `T/record/proposal.go:323-332` | actingSeat | C-core | 10 | deleted · X22 | actingSeat: acting seat through the chain; bench writes a Mint only as a supersede successor (proposal); becomes the event seat |
| B | `T/record/proposal.go:334-342` | mintActors | C-core | 9 | deleted · X22 | mintActors: acting seat via mintChain; becomes each mint event seat; 6 readers with other jobs still need the minter |
| B | `T/record/proposal.go:344-364` | mintChain | MIXED | 6 | chain arm deleted · X22 | ancestor-assignee inheritance (357-359), acting map, comment go; assignee map stays as the minter map (assigneeOf, 12 readers) |
| B | `T/record/proposal.go:366-367` | actingSeatOf | C-core | 2 | deleted · X22 | actingSeatOf wraps mintActors (chain acting seat); caller inquiryJudgementsOf reads the minter instead |
| B | `T/record/proposal.go:369-380` | mintActingSeat | C-core | 12 | deleted · X22 | mintActingSeat: bench successor checked against the ancestor assignee (chain origin) |
| B | `T/record/proposal.go:382-390` | benchClosureWord | MIXED | 6 | chain arm deleted · X22 | `superseded` word for a ruling siding with a supersede proposal goes (382-384 comment, 386-388); returns the disposition word |
| B | `T/record/proposal.go:392-406` | StoredAct | MIXED | 5 | chain arm deleted · X22 | Motion_Supersede arm (class material cleared, 400-403) and comment clause go; StoredAct stays for regrade/close |
| B | `T/record/proposal.go:408-430` | requireNotEstoppedOn | MIXED | 5 | chain arm deleted · X22 | 'successor a supersede accept left standing' (409 comment, 422-425) goes; estoppel refusal stays |
| B | `T/record/proposal.go:16` | - | C-text | 1 | rewritten · X20 | file comment lists 'or a supersede' among proposed acts |
| B | `T/record/questionpart.go:130-134` | questionGapsSQL | C-text | 1 | rewritten · X20 | comment 'which a bench-written successor's writer is not'; with no bench successor the gap_state JOIN for assignee reduces to the mint's seat |
| B | `T/record/questionpart.go:361-390` | requireQuestionPart | O | 0 | O | SupersedesCoverage: a coverage record supersedes a question part's earlier coverage |
| B | `T/record/questionpart.go:467-477` | partCoveragesAtSQL | O | 0 | O | supersedes_coverage: question-part coverage record vs its proposal |
| B | `T/record/reading.go:50-155` | requireNotARetriedReading | O | 0 | O | SupersedingAnswer: superseder table (answer to a refused repeated reading) |
| B | `T/record/record.go:667-716` | appendBodies | MIXED | 6 | chain arm deleted · X22 | unwritableSuccessors read at BEGIN for the freeze's successor arm (676-679, var, 692) goes; appendBodies stays |
| B | `T/record/record.go:804-886` | writeIn | O | 0 | O | superseder table tail (supersede.go) on a duplicate-key refusal |
| B | `T/record/record.go:888-1706` | validate | MIXED | 66 | chain arm deleted · X22 | chain 38 (939-944 stamp, 991-1002 mint arm, carry+docket requireNotSuperseded, 1462-1468 engage); C-regression 28 (1128-1165); supersedes_status O |
| B | `T/record/record.go:1782-1803` | requireOriginator | C-text | 2 | rewritten · X20 | 'the lens that minted its chain's first gap' in comment (1783) and refusal (1800); originator rule stays |
| B | `T/record/record.go:1975-2031` | mintShape | C-text | 1 | rewritten · X20 | comment 'or stored as a proposed supersede'; mintShape (mint statement shape) stays |
| B | `T/record/refs.go:359-369` | gapState | C-text | 1 | rewritten · X20 | comment 'or the mint that superseded it' among closing arms |
| B | `T/record/refs.go:371-388` | requireOpenGap | MIXED | 5 | chain arm deleted · X22 | successor-naming refusal (371-372 comment, 382-384 requireNotSuperseded) goes; closed-gap refusal stays |
| B | `T/record/refs.go:17` | - | C-text | 1 | rewritten · X20 | history comment: 'mint --supersedes' was checked |
| B | `T/record/refs.go:30` | - | C-text | 1 | rewritten · X20 | measurement comment: '6 successors' |
| B | `T/record/refs.go:354` | - | C-text | 1 | rewritten · X20 | dropped-rule comment 'supersedes must name a closed gap' |
| B | `T/record/refs.go:355` | - | C-text | 1 | rewritten · X20 | same comment: superseding a still-open predecessor is normal |
| B | `T/record/remand.go:161-167` | remandedClosedGapSQL | MIXED | 1 | chain arm deleted · X22 | COALESCE(g.supersede_event) term in max() goes |
| B | `T/record/remand.go:370-388` | partiesOnRoute | O | 0 | O | engages: a route engages the assignee and blue (dispatch) |
| B | `T/record/replay.go:111-176` | Gap | MIXED | 4 | chain arm deleted · X22 | SupersededBy field+comment (153-156) goes; 142 'a successor' in Close is Close.successor (C-regression text) |
| B | `T/record/replay.go:230-263` | (*Gap).ClosureReason | MIXED | 5 | chain arm deleted · X22 | supersession arm (252-256) goes; closure word stays |
| B | `T/record/replay.go:265-267` | ClosedSuperseded | C-core | 3 | deleted · X22 | ClosedSuperseded word; also read by benchClosureWord (proposal) and FoundClosure.Fate, consistency.walk |
| B | `T/record/replay.go:269-279` | (*Gap).Successor | MIXED | 5 | chain arm deleted · X22 | SupersededBy arm + comment go; remaining arm returns Close.successor (C-regression) |
| B | `T/record/replay.go:341-365` | priorClosureEpochs | C-text | 2 | rewritten · X20 | comment 'the same rule mint applies to supersedes' |
| B | `T/record/required.go:9-55` | RequiredFields | C-text | 1 | rewritten · X20 | l.13 '--supersedes (optional)' example; l.30 '--successor' is a regression example |
| B | `T/record/retirement.go:205-267` | gapNewsOf | O | 0 | O | engaged: blue's sitting engaged on the gap by dispatch |
| B | `T/record/retirement.go:359-375` | freshMaterialOf | MIXED | 1 | chain arm deleted · X22 | 'AND supersedes IS NULL' (fresh = superseding nothing) goes |
| B | `T/record/retirement.go:377-387` | freshMaterialOfStates | MIXED | 1 | chain arm deleted · X22 | '&& g.Supersedes == ""' goes |
| B | `T/record/retirement.go:20` | - | O | 0 | O | engaged: a lens engaged on gaps by dispatch |
| B | `T/record/retirement.go:41` | - | O | 0 | O | engaged: sittings a seat is engaged for by dispatch |
| B | `T/record/retirement.go:52` | - | C-text | 1 | rewritten · X20 | '(superseding nothing)' in productive-sitting definition |
| B | `T/record/sitting.go:160-626` | sittingOf | MIXED | 13 | chain arm deleted · X22; its engage part kept, restated · X22a | ItemSuccessorOwed block (367-379) goes; 326-341 engagedOf.GapIDs / 'not engaged on it' is blue's dispatch engagement (O) |
| B | `T/record/sitting.go:661-675` | sittingOfBlueRespond | O | 0 | O | engagedOf / found closed: blue's dispatch engagement split by closes, not successors |
| B | `T/record/sitting.go:677-693` | EngagedJSON | MIXED | 6 | chain arm deleted · X22; its engage part kept, restated · X22a | SuccessorsOwed field+comment (687-692) goes; GapIDs/FoundClosed = blue's dispatch engagement (O) |
| B | `T/record/sitting.go:695-706` | FoundClosedJSON | MIXED | 5 | chain arm deleted · X22 | Successors field+comment (696-697, 702-705) goes; found-closed row (id, closer, fate) stays |
| B | `T/record/sitting.go:708-746` | engagedOf | MIXED | 8 | chain arm deleted · X22 | successorsOf/successorsOwedOnTheWorkList reads and Successors fill (732-735, 738-742) go; engaged block stays |
| B | `T/record/sitting.go:764-780` | benchRegisteredSince | O | 0 | O | hit on owedSitting / registered occasion: the sitting a seat owes for its dispatch; no chain sense |
| B | `T/record/standingreading.go:746-784` | readingBehind | MIXED | 1 | chain arm deleted · X22 | mintActors(evs)[gapID] -> the mint's own seat (l.758 rewrite) |
| B | `T/record/standingreading.go:863-878` | mintRowsOf | MIXED | 1 | chain arm deleted · X22 | acting := mintActors(evs) -> e.GetSeatId() (l.867/873 rewrite) |
| B | `T/record/successor.go:26-47` | successorsOf | C-engage | 22 | chain wording/input deleted · X22a | successorsOf: Close.superseded_by (regression) + Mint.supersedes successors; feeds engage and FoundClosedJSON.Successors |
| B | `T/record/successor.go:49-54` | OwedSuccessor | C-engage | 6 | kept, restated · X22a | OwedSuccessor |
| B | `T/record/successor.go:56-65` | successorsOwed | C-engage | 10 | kept, restated · X22a | successorsOwed |
| B | `T/record/successor.go:67-81` | successorsOwedOnTheWorkList | C-engage | 15 | kept, restated · X22a | successorsOwedOnTheWorkList |
| B | `T/record/successor.go:83-106` | owedSuccessorsOf | C-engage | 24 | kept, restated · X22a | owedSuccessorsOf |
| B | `T/record/successor.go:108-123` | blueSittingNow | C-engage | 16 | kept, restated · X22a | blueSittingNow: sole caller successorsOwed |
| B | `T/record/successor.go:125-138` | blueEngagedThisEpoch | C-engage | 14 | kept, restated · X22a | blueEngagedThisEpoch: callers owedSuccessorsOf, requireEngage only |
| B | `T/record/successor.go:140-156` | successorsOwedAt | C-engage | 17 | kept, restated · X22a | successorsOwedAt |
| B | `T/record/successor.go:158-167` | EngagementOwesNone | C-engage | 10 | kept, restated · X22a | EngagementOwesNone: migrate/replay.go reads it |
| B | `T/record/successor.go:169-175` | isEngageRow | C-engage | 7 | kept, restated · X22a | isEngageRow: dispatchLedger/dispatchRow (another job) branch on it |
| B | `T/record/successor.go:177-226` | requireEngage | C-engage | 50 | kept, restated · X22a | requireEngage |
| B | `T/record/successor.go:228-249` | requireEngageNamesEveryOwed | C-engage | 22 | kept, restated · X22a | requireEngageNamesEveryOwed (judgeAct act judgement) |
| B | `T/record/successor.go:251-259` | isSuccessorOfAny | C-engage | 9 | chain wording/input deleted · X22a | isSuccessorOfAny |
| B | `T/record/successor.go:13` | - | C-engage | 1 | reworded · X22a | successor.go header (R2 ruling) |
| B | `T/record/successor.go:17` | - | C-engage | 1 | reworded · X22a | successor.go header |
| B | `T/record/successor.go:20` | - | C-engage | 1 | reworded · X22a | successor.go header |
| B | `T/record/successor.go:21` | - | C-engage | 1 | reworded · X22a | successor.go header |
| B | `T/record/successor.go:22` | - | C-engage | 1 | reworded · X22a | successor.go header |
| B | `T/record/successor.go:23` | - | C-engage | 1 | reworded · X22a | successor.go header |
| B | `T/record/successor.go:24` | - | C-engage | 1 | reworded · X22a | successor.go header |
| B | `T/record/supersede.go:29-44` | superseder | O | 0 | O | superseder table: the act that answers a refused repeated act |
| B | `T/record/supersede.go:46-57` | supersederKey | O | 0 | O | superseder table key |
| B | `T/record/supersede.go:65-107` | superseders | C-text | 4 | rewritten · X20 | superseders rows CLOSE (82-84) and lens docket-accept (91-92) send seats to --supersedes / a supersede proposal; table itself is O |
| B | `T/record/supersede.go:128-146` | AnswerTo | O | 0 | O | superseder table lookup |
| B | `T/record/supersede.go:148-177` | AnswersTo | O | 0 | O | superseder table rows for help |
| B | `T/record/supersede.go:190-210` | (superseder).render | O | 0 | O | superseder row render |
| B | `T/record/supersede.go:212-220` | SupersedingAnswer | O | 0 | O | superseder tail for refusals of a repeated act |
| B | `T/record/supersede.go:243-255` | RefusesARepeat | O | 0 | O | superseder table coverage (RefusesARepeat) |
| B | `T/record/supersede.go:17` | - | O | 0 | O | superseder table header |
| B | `T/record/supersession.go:24-61` | supersessionsOf | C-core | 38 | deleted · X22 | supersessionsOf (incl. bench-sided supersede proposal arm) |
| B | `T/record/supersession.go:63-69` | successorClosing | C-core | 7 | deleted · X22 | successorClosing reads gap.superseded_by |
| B | `T/record/supersession.go:71-75` | SupersededGap | C-core | 5 | deleted · X22 | SupersededGap: migrate actOnSupersededGap reads it |
| B | `T/record/supersession.go:77-83` | supersededRefusal | C-core | 7 | deleted · X22 | supersededRefusal |
| B | `T/record/supersession.go:85-112` | requireSupersedable | C-core | 28 | deleted · X22 | requireSupersedable (refusal text also names the supersede proposal) |
| B | `T/record/supersession.go:114-149` | requireSupersededOffDocket | C-core | 36 | deleted · X22 | requireSupersededOffDocket (except = the proposal's own successor) |
| B | `T/record/supersession.go:151-161` | requireNotSuperseded | C-core | 11 | deleted · X22 | requireNotSuperseded |
| B | `T/record/supersession.go:9` | - | C-core | 1 | deleted · X22 | supersession.go header |
| B | `T/record/supersession.go:10` | - | C-core | 1 | deleted · X22 | supersession.go header |
| B | `T/record/supersession.go:11` | - | C-core | 1 | deleted · X22 | supersession.go header |
| B | `T/record/supersession.go:12` | - | C-core | 1 | deleted · X22 | supersession.go header |
| B | `T/record/supersession.go:15` | - | C-core | 1 | deleted · X22 | supersession.go header |
| B | `T/record/supersession.go:16` | - | C-core | 1 | deleted · X22 | supersession.go header |
| B | `T/record/supersession.go:21` | - | C-core | 1 | deleted · X22 | supersession.go header |
| M | `T/record/gapacts.go:11-43` | gapAct | C-text | 1 | rewritten · X20 | act list names 'a mint superseding it' (l.23); gapAct stays |
| M | `T/record/gapacts.go:45-97` | gapActsOf | MIXED | 2 | chain arm deleted · X22 | case Mint: on = GetSupersedes() arm goes |
| M | `T/record/gapacts.go:99-194` | fateActsOf | MIXED | 15 | chain arm deleted · X22 | superseded-gap exclusion (comment, supersessionsOf call, two DeleteFunc filters) goes; fate acts stay |
| M | `T/record/gapacts.go:208-221` | lensTurnOn | O | 0 | O | engages: the plan engages the minting lens (dispatch) |
| M | `T/record/grounds.go:72-73` | GroundsKinds | C-text | 1 | rewritten · X20 | --grounds kinds phrase 'or superseding it' |
| M | `T/record/grounds.go:113-123` | groundsActSQL | MIXED | 1 | chain arm deleted · X22 | OR EXISTS mint_supersedes term goes; grounds lookup stays |
| M | `T/record/grounds.go:26` | - | C-text | 1 | rewritten · X20 | grounds comment 'superseding it' |
| M | `T/record/impasse.go:14-82` | GapExchanges | C-text | 1 | rewritten · X20 | movement clause 'its lineage moved (a mint superseding G)'; engaged = dispatch (O) |
| M | `T/record/impasse.go:266-421` | exchangeFold | MIXED | 3 | chain arm deleted · X22 | repeated supersedes counted as movement (294-296) goes |
| M | `T/record/inquirycheck.go:89-155` | inquiryJudgementsOf | MIXED | 4 | chain arm deleted · X22 | superseding mint as a cause (108 comment, 123-125) goes |
| M | `T/record/itemkind.go:17-70` | ItemAudit,ItemReaudit,ItemRegradeOwed,ItemFindingOnGap,... | MIXED | 3 | chain arm deleted · X22; its engage part kept, restated · X22a | ItemSuccessorOwed const+comment (46-48) goes; 26 engaged = dispatch (O) |
| M | `T/record/itemkind.go:72-82` | ItemKinds | MIXED | 1 | chain arm deleted · X22; its engage part kept, restated · X22a | ItemSuccessorOwed dropped from list |
| M | `T/record/lensturn.go:46-79` | earlyLensDispatchesOf | MIXED | 13 | chain arm deleted · X22 | superseded map (52-61) and skip (70-72) go |
| M | `T/record/motion.go:540-549` | requireUnanswered | O | 0 | O | SupersedingAnswer: superseder table |
| M | `T/record/oncepersitting.go:47-80` | requireOncePerSitting | O | 0 | O | superseder table tail |
| M | `T/record/onegapperclass.go:15-170` | requireOneGapPerClassAndSubject | MIXED | 4 | chain arm deleted · X22 | --supersedes exemption comment (47-49) and DeleteFunc (77) go |
| M | `T/record/reading.go:54-134` | requireNotARetriedReading | O | 0 | O | superseder table tail |
| M | `T/record/record.go:774-855` | writeIn | O | 0 | O | superseder table tail |
| M | `T/record/record.go:857-1739` | validate | MIXED | 59 | chain arm deleted · X22 | chain 31 (1012-1022 mint arm, carry+docket requireNotSuperseded, 1500-1506 engage); C-regression 28 (1158-1196); supersedes_status O |
| M | `T/record/record.go:1803-1824` | requireOriginator | O | 0 | O | engages: the chair's dispatch engages the lens (main has no chain text here) |
| M | `T/record/remand.go:162-170` | remandedClosedGapSQL | MIXED | 1 | chain arm deleted · X22 | COALESCE(g.supersede_event) term goes |
| M | `T/record/remand.go:366-384` | partiesOnRoute | O | 0 | O | engages: route engages minting lens and blue (dispatch) |
| M | `T/record/replay.go:236-269` | (*Gap).ClosureReason | MIXED | 5 | chain arm deleted · X22 | supersession arm (258-262) goes |
| M | `T/record/required.go:9-55` | RequiredFields | C-text | 1 | rewritten · X20 | l.13 '--supersedes (optional)'; l.30 '--successor' is a regression example |
| M | `T/record/retirement.go:92-130` | (lensFold).readiness | O | 0 | O | engages: the plan engages the lens on its whole_report_owed blocker |
| M | `T/record/retirement.go:216-262` | gapNewsOf | O | 0 | O | engaged: blue's sitting engaged on the gap by dispatch |
| M | `T/record/retirement.go:390-406` | freshMaterialOf | MIXED | 1 | chain arm deleted · X22 | 'AND supersedes_count = 0' goes |
| M | `T/record/retirement.go:408-418` | freshMaterialOfStates | MIXED | 1 | chain arm deleted · X22 | '&& len(g.Supersedes) == 0' goes |
| M | `T/record/sitting.go:150-584` | sittingOf | MIXED | 13 | chain arm deleted · X22; its engage part kept, restated · X22a | ItemSuccessorOwed block (351-363) goes; 310-325 engaged = blue's dispatch engagement (O) |
| M | `T/record/successor.go:26-47` | successorsOf | C-engage | 22 | chain wording/input deleted · X22a | successorsOf (repeated supersedes) |
| M | `T/record/successor.go:148-209` | requireEngage | C-engage | 62 | kept, restated · X22a | requireEngage |
| M | `T/record/successor.go:211-244` | requireEngageNamesEveryOwed | C-engage | 34 | kept, restated · X22a | requireEngageNamesEveryOwed |
| M | `T/record/supersede.go:43-50` | supersederKey | O | 0 | O | superseder table key |
| M | `T/record/supersede.go:59-102` | superseders | O | 0 | O | superseder table (main CLOSE row answers with a docket motion, no chain text) |
| M | `T/record/supersede.go:121-140` | AnswerTo | O | 0 | O | superseder table lookup |
| M | `T/record/supersede.go:142-162` | AnswersTo | O | 0 | O | superseder table rows for help |
| M | `T/record/supersede.go:197-205` | SupersedingAnswer | O | 0 | O | superseder tail for refusals of a repeated act |
| M | `T/record/supersession.go:24-49` | supersessionsOf | C-core | 26 | deleted · X22 | supersessionsOf (repeated supersedes) |
| M | `T/record/supersession.go:73-98` | requireSupersedable | C-core | 26 | deleted · X22 | requireSupersedable (ancestor by ancestor; minter-only) |
| M | `T/record/supersession.go:100-117` | requireSupersededOffDocket | C-core | 18 | deleted · X22 | requireSupersededOffDocket (ancestor by ancestor) |

**Group C** (`amendv/tag-C.tsv`)

| t | file:lines | declaration | tag | del | change | note |
|---|---|---|---|---|---|---|
| B | `T/record/migrate/docketing.go:39-50` | archivedDockets | MIXED | 5 | chain arm deleted · X22 | superseded map + doc (44-48): dockets on gaps a successor's mint closed; docketing/dropped/gapOf stay |
| B | `T/record/migrate/docketing.go:52-85` | docketsOf | MIXED | 4 | chain arm deleted · X22 | superseded init and supersededStanding branch (76-78) go; dispatch-docket pre-pass stays |
| B | `T/record/migrate/docketing.go:117-146` | supersededStanding | C-core | 30 | deleted · X22 | archive: a mint superseding the gap before ruling/close (requireSupersededOffDocket premise) |
| B | `T/record/migrate/docketing.go:154-158` | keptDocket,partyDocket,supersededDocket | MIXED | 1 | chain arm deleted · X22 | supersededDocket value goes; keptDocket/partyDocket stay |
| B | `T/record/migrate/docketing.go:160-211` | (archivedDockets).translate | MIXED | 8 | chain arm deleted · X22 | two superseded arms (175-178, 192-195) go; Translations 3/4 stay |
| B | `T/record/migrate/docketing.go:33` | loose | C-text | 2 | rewritten · X20 | file comment 32-33: dispatch docket on a superseded gap written as no event |
| B | `T/record/migrate/entries.go:13-129` | Entries | MIXED | 1 | chain arm deleted · X22 | oneAncestor( wrapper on the mint entry (90) goes; entry table stays |
| B | `T/record/migrate/entries.go:362-386` | oneAncestor | C-core | 25 | deleted · X22 | Translation 1: a source's supersedes list -> first ancestor |
| B | `T/record/migrate/eraentries.go:28-68` | eraEntries | O | 0 | O | ancestor: inquiry-support was the inquiry review's predecessor event kind |
| B | `T/record/migrate/eraentries.go:268-287` | inquiryEntry | O | 0 | O | supersedes_status: a fate row supersedes the inquiry's proposal row |
| B | `T/record/migrate/lineage.go:12-35` | dropCrossSeatLink | C-core | 24 | deleted · X22 | Translation 2: drops a cross-seat ancestor link because assignee = chain origin |
| B | `T/record/migrate/manifest.go:26-55` | NewManifest | MIXED | 1 | chain arm deleted · X22; its engage part kept, restated · X22a | UnowedEngagements (51, C-engage) goes; SupersededOutcomes (40) O: an outcome a later outcome replaced |
| B | `T/record/migrate/outcomes.go:22-56` | supersededOutcomes | O | 0 | O | an outcome a later outcome replaced (correction/replacement), not the gap chain |
| B | `T/record/migrate/plankinds.go:22-31` | planKindsRetired | C-text | 2 | rewritten · X20 | comment 24-25 justifies stranded_gap by 'a mint superseding an open gap closes it'; map stays |
| B | `T/record/migrate/publish.go:41-49` | build | O | 0 | O | ancestor: nearest existing parent directory of --to |
| B | `T/record/migrate/publish.go:119-159` | newBuild | O | 0 | O | ancestor: nearest existing parent directory |
| B | `T/record/migrate/publish.go:222-259` | (*build).publish | O | 0 | O | ancestor: nearest existing parent directory (fsync order) |
| B | `T/record/migrate/remap.go:100-151` | idFields,argumentFields,dataFields,mints | MIXED | 1 | chain arm deleted · X22 | tokens Mint.supersedes (C-core), Close.successor (C-regression) go; Inquiry.supersedes_status, Retire.superseded_by, QuestionPart.supersedes_coverage O |
| B | `T/record/migrate/replay.go:34-60` | Result | MIXED | 1 | chain arm deleted · X22; its engage part kept, restated · X22a | UnowedEngagements (52, C-engage) goes; SupersededOutcomes (45) O |
| B | `T/record/migrate/replay.go:102-767` | replay | MIXED | 34 | chain arm deleted · X22; its engage part kept, restated · X22a | drops: actOnSupersededGap (458-465), dropCrossSeatLink call, UnowedEngagements block (C-engage), supersededMotions, Dockets.Superseded; SupersededOutcomes O |
| B | `T/record/migrate/replay.go:769-793` | carryInquiryMethod | O | 0 | O | SupersedesStatus: fate row vs proposal row |
| B | `T/record/migrate/replay.go:912-965` | actOnSupersededGap | C-core | 54 | deleted · X22 | drops archive acts on a gap a successor's mint closed (record.SupersededGap) |
| B | `T/record/migrate/sqlitesource.go:236-314` | (*SQLiteSource).classify | C-text | 2 | rewritten · X20 | demand_lineage case STAYS (archived sources hold the table); comment 262-264 'destination folds its own' rewritten |
| B | `T/record/migration/manifest.go:24-120` | Manifest | MIXED | 4 | chain arm deleted · X22; its engage part kept, restated · X22a | UnowedEngagements field+doc (98-101, C-engage) goes; SupersededOutcomes O |
| B | `T/record/migration/manifest.go:167-173` | SupersededOutcome | O | 0 | O | a source outcome a later outcome replaced |
| B | `T/record/migration/manifest.go:201-209` | DiscardedSitting | O | 0 | O | a superseded dispatch: a later dispatch replaced it |
| B | `T/record/migration/manifest.go:242-265` | DocketResidue | MIXED | 4 | chain arm deleted · X22 | Superseded field + doc (251-254) goes; other residues stay |
| B | `T/record/planguard/plan.go:25-30` | Plan | O | 0 | O | ancestors: plan-tree parent nodes |
| B | `T/record/planguard/plan.go:32-65` | BuildPlan | O | 0 | O | ancestor: plan-tree parent node |
| B | `T/record/planguard/shape.go:312-325` | perRow | O | 0 | O | ancestor: plan-tree parent node |
| B | `T/record/recordpb/descriptions.go:40-45` | DemandClause | C-text | 1 | rewritten · X20 | clause 'or once its lens mints a gap ... superseding it, which takes the demand on' cut from the one-line const |
| B | `T/record/recordsql/demandlineage.go:5-25` | DemandLineageDDL | C-core | 21 | deleted · X22 | doc of a chain-only projection; NB line 60 also appends evidenceCheckIndexesSQL (not chain): must move to another DDL var |
| B | `T/record/recordsql/demandlineage.go:26-30` | DemandLineageDDL/table:demand_lineage | C-core | 5 | deleted · X22 | chain-only: with no supersedes every row is (G,G); demand_reach reads mint h = a directly |
| B | `T/record/recordsql/demandlineage.go:31-33` | DemandLineageDDL/index:demand_lineage_from | C-core | 3 | deleted · X22 | heirs-of-a-gap index |
| B | `T/record/recordsql/demandlineage.go:34-36` | DemandLineageDDL/trigger:demand_lineage_is_append_only_update | C-core | 3 | deleted · X22 | append-only guard of demand_lineage |
| B | `T/record/recordsql/demandlineage.go:37-40` | DemandLineageDDL/trigger:demand_lineage_is_append_only_delete | C-core | 4 | deleted · X22 | append-only guard of demand_lineage |
| B | `T/record/recordsql/demandlineage.go:41-45` | DemandLineageDDL/trigger:demand_lineage_mint | C-core | 5 | deleted · X22 | (G,G) self row: an identity join once the chain goes |
| B | `T/record/recordsql/demandlineage.go:46-60` | DemandLineageDDL/trigger:demand_lineage_supersedes | C-core | 15 | deleted · X22 | passes the ancestor's demand to the first superseding mint |
| B | `T/record/recordsql/gapstate.go:9-33` | GapStateDDL | O | 0 | O | demand_reach/demandReachSQL: a gap's own anchor reach, which stays |
| B | `T/record/recordsql/gapstate.go:34-66` | GapStateDDL/table:gap_state | MIXED | 4 | chain arm deleted · X22 | columns supersede_event, supersedes, superseded_by + 'the mint that superseded it' (36); assignee stays (= writer) |
| B | `T/record/recordsql/gapstate.go:123-127` | reachingSQL | O | 0 | O | reads demand_reach (gap's own reach), stays |
| B | `T/record/recordsql/gapstate.go:129-165` | gapMarkers | MIXED | 2 | chain arm deleted · X22 | mint marker's UNION SELECT NEW.supersedes + comment; marker table stays |
| B | `T/record/recordsql/gapstate.go:181-369` | GapRefreshViewDDL | MIXED | 42 | chain arm deleted · X22 | motion_supersede bench arm 208-221 (C-proposal); supersede_event arm+k CTE; lineage join; from_gap; 3 cols; assigneeSQL |
| B | `T/record/recordsql/gapstate.go:371-380` | assigneeSQL | C-core | 10 | deleted · X22 | assignee = chain origin; its 3 sites (gap_refresh, demandReachSQL, proposal.go comment) take the writer e.seat_id |
| B | `T/record/recordsql/schema.go:257-261` | QueryIndexDDL/index:mint_about | O | 0 | O | comment names the demand_reach view, which stays |
| B | `T/record/recordsql/schema.go:292-294` | QueryIndexDDL/index:mint_ancestor | C-core | 3 | deleted · X22 | index for supersede_event and demand_lineage_supersedes |
| B | `T/record/recordsql/schema.go:300-329` | Schema | MIXED | 1 | chain arm deleted · X22 | WriteString(DemandLineageDDL) goes, but evidenceCheckIndexesSQL it carries must still be written |
| B | `T/record/recordsql/schema.go:748-757` | listTable | C-text | 2 | rewritten · X20 | doc cites supersedes as the list that made a gap's lineage unqueryable |
| B | `T/record/recordsql/views.go:342-363` | ViewsDDL/view:demand_reach | MIXED | 8 | chain arm deleted · X22 | inheritance doc 351-357 + FROM demand_lineage; view stays (EvidenceDemandOf, reachingRetiresOf, gap markers) |
| B | `T/record/recordsql/views.go:364-471` | ViewsDDL/view:gap | MIXED | 6 | chain arm deleted · X22 | doc 396-398 + cols supersede_event, supersedes, superseded_by; assignee doc 'chain's origin' |
| B | `T/record/recordsql/views.go:509-535` | ViewsDDL/view:inquiry_state | O | 0 | O | supersedes_status: fate row vs proposal row |
| B | `T/record/recordsql/views.go:536-584` | ViewsDDL/view:question_part_state | O | 0 | O | supersedes_coverage: coverage row vs proposal row |
| B | `T/record/recordsql/views.go:667-679` | demandReachSQL | MIXED | 5 | chain arm deleted · X22 | lineage param, from_gap, lineage JOIN, assigneeSQL->e.seat_id, doc; lineage arg also drives join order (planguard baseline moves) |
| B | `T/record/tombstone.go:294-308` | RulingTombstone | MIXED | 4 | chain arm deleted · X22 | On + successor fields and doc 295-296 (C-proposal); ruling row stays |
| B | `T/record/tombstone.go:310-320` | rulingTombstoneSQL | MIXED | 3 | chain arm deleted · X22 | motion_supersede LEFT JOIN, s.gap_id column, doc (C-proposal) |
| B | `T/record/tombstone.go:322-350` | rulingTombstonesAt | MIXED | 1 | chain arm deleted · X22 | &t.successor scan target (C-proposal) |
| B | `T/record/tombstone.go:352-370` | rulingProtecting | MIXED | 9 | chain arm deleted · X22 | successor arm 363-369 + doc (C-proposal); own-row read stays |
| B | `T/record/tombstone.go:372-397` | rulingsProtectingAll | MIXED | 14 | chain arm deleted · X22 | successor protection 382-394 + doc (C-proposal); ruled-gap map stays |
| B | `T/record/tombstone.go:457-463` | (RulingTombstone).ruled | C-proposal | 7 | deleted · X1/X3/X4 | On-or-Gap exists only for a successor's protection; callers read .Gap |
| B | `T/record/viewjson.go:72-196` | GapJSON | MIXED | 10 | chain arm deleted · X22 | Supersedes/SupersededBy fields+doc (173-179), Assignee 'chain's origin' (170), successor clause in RulingTombstone doc (159) |
| B | `T/record/viewjson.go:198-211` | strs | C-text | 4 | rewritten · X20 | doc 202-205 cites the 'supersedes: []' lineage rule; func stays (2 callers) |
| B | `T/record/viewjson.go:327-335` | BoardJSONOfRun | C-text | 1 | rewritten · X20 | 'the ancestor' in the scalar list |
| B | `T/record/viewjson.go:337-350` | boardView | MIXED | 1 | chain arm deleted · X22 | EVENT_TYPE_MINT loaded only for a successor's mint as closing act |
| B | `T/record/viewjson.go:352-498` | boardJSONOfRecord | MIXED | 4 | chain arm deleted · X22 | supersedes column, ancestor var/scan, Supersedes: ancestor, SupersededBy assign |
| B | `T/record/viewjson.go:500-512` | closeState | MIXED | 1 | chain arm deleted · X22 | supersededBy field |
| B | `T/record/viewjson.go:514-517` | (*closeState).reason | MIXED | 1 | chain arm deleted · X22 | SupersededBy argument |
| B | `T/record/viewjson.go:519-524` | gapClosure | MIXED | 2 | chain arm deleted · X22 | supersedeEvent, supersededBy |
| B | `T/record/viewjson.go:526-527` | gapClosureColumns | MIXED | 1 | chain arm deleted · X22 | supersede_event, superseded_by columns |
| B | `T/record/viewjson.go:529-532` | (*gapClosure).dest | MIXED | 1 | chain arm deleted · X22 | two scan targets |
| B | `T/record/viewjson.go:543-591` | closeStateOf | MIXED | 12 | chain arm deleted · X22 | supersede arm 567-577 + condition 550; close/bench arms stay |
| B | `T/record/viewjson.go:616-661` | WorkJSON | O | 0 | O; its engage part kept, restated · X22a | Engaged: blue's dispatch engagement; SuccessorsOwed lives in sitting.go EngagedJSON (other group) |
| B | `T/record/viewjson.go:722-858` | WorkGapJSON | C-text | 1 | rewritten · X20 | Assignee doc 'its chain's origin'; YourTurn 'engages' is O (dispatch) |
| B | `T/record/viewjson.go:860-931` | EstoppedJSON | C-text | 2 | rewritten · X20 | 878 Close.successor regression 'successor R2-1' in a measurement note; 928 ruling 'on a successor' (C-proposal) |
| B | `T/record/viewjson.go:947-1024` | WorkGapState | MIXED | 1 | chain arm deleted · X22 | Supersedes field; 1000 'route engages' O |
| B | `T/record/viewjson.go:1026-1055` | workReads | MIXED | 3 | chain arm deleted · X22 | supersedes []standingSupersede + doc (C-proposal) |
| B | `T/record/viewjson.go:1088-1183` | workReadsAt | MIXED | 5 | chain arm deleted · X22 | standingSupersedesAt read (C-proposal), supersedes column + scan (C-core) |
| B | `T/record/viewjson.go:1314-1380` | workJSONOfGaps | MIXED | 3 | chain arm deleted · X22 | 'inherits:'+Supersedes else arm (C-core); successor in comment (C-proposal) |
| B | `T/record/viewjson.go:1397-1429` | frozenIn | MIXED | 1 | chain arm deleted · X22 | rd.Successor condition (Road.Successor, C-proposal) |
| B | `T/record/viewjson.go:1431-1453` | frozenDebtIn | C-proposal | 23 | deleted · X1/X3/X4 | freeze on a standing supersede proposal's successor (L48) |
| B | `T/record/viewjson.go:1550-1598` | WorkOfSeat | O | 0 | O | engagedOf: blue's dispatch engagement, stays |
| B | `T/record/viewjson.go:1736-1746` | DebateBenchJSON | O | 0 | O | an earlier certification superseded by the last |
| B | `T/record/windowindex.go:80-84` | reaching | MIXED | 2 | chain arm deleted · X22 | from field + doc; read by blueanswered.go:151-152 (inherited-demand wording) |
| B | `T/record/windowindex.go:86-102` | retiredAnchorsSQL,reachedGapsSQL,firstClosureSQL | MIXED | 4 | chain arm deleted · X22 | comment 94-95, from_gap column + ORDER BY in reachedGapsSQL |
| B | `T/record/windowindex.go:104-153` | reachingRetiresOf | MIXED | 4 | chain arm deleted · X22 | from plumbing (122,129,146) + doc 106 |
| M | `T/record/migrate/eraentries.go:28-69` | eraEntries | O | 0 | O | ancestor: inquiry-support was the inquiry review's predecessor event kind |
| M | `T/record/migrate/lostacts.go:15` | loose | C-text | 1 | rewritten · X20 | 'a mint superseding it' as an act depending on a lost close |
| M | `T/record/migrate/manifest.go:26-52` | NewManifest | O | 0 | O | SupersededOutcomes only: an outcome a later outcome replaced |
| M | `T/record/migrate/remap.go:100-150` | idFields,argumentFields,dataFields,mints | MIXED | 1 | chain arm deleted · X22 | tokens Mint.supersedes, Close.successor (C-regression) go; supersedes_status, Retire.superseded_by, supersedes_coverage O |
| M | `T/record/migrate/replay.go:32-55` | Result | O | 0 | O | SupersededOutcomes: an outcome replaced |
| M | `T/record/migrate/replay.go:103-664` | replay | MIXED | 15 | chain arm deleted · X22 | supersededMotions (248) + actOnSupersededGap drop block (373-386); SupersededOutcomes hits O |
| M | `T/record/migrate/replay.go:809-846` | actOnSupersededGap | C-core | 38 | deleted · X22 | drops archive acts on a gap a successor's mint closed |
| M | `T/record/migrate/sqlitesource.go:236-314` | (*SQLiteSource).classify | C-text | 2 | rewritten · X20 | demand_lineage case STAYS for archived sources; comment rewritten |
| M | `T/record/migration/manifest.go:24-108` | Manifest | O | 0 | O | SupersededOutcomes: an outcome replaced |
| M | `T/record/planguard/rule.go:22` | loose | C-text | 1 | rewritten · X20 | example plan line 'SCAN mint_supersedes USING INDEX ...' |
| M | `T/record/recordpb/descriptions.go:40-45` | DemandClause | C-text | 2 | rewritten · X20 | 'superseding it, which takes the demand on' + '(or, for a demand carried from an ancestor, ...)' |
| M | `T/record/recordsql/demandlineage.go:5-25` | DemandLineageDDL | C-core | 21 | deleted · X22 | doc of chain-only projection (mint_supersedes.value); line 62 appends evidenceCheckIndexesSQL: must move |
| M | `T/record/recordsql/demandlineage.go:46-62` | DemandLineageDDL/trigger:demand_lineage_supersedes | C-core | 17 | deleted · X22 | trigger on mint_supersedes, OR IGNORE for diamonds |
| M | `T/record/recordsql/gapstate.go:34-63` | GapStateDDL/table:gap_state | MIXED | 4 | chain arm deleted · X22 | supersede_event, supersedes_count, superseded_by + comment 36 |
| M | `T/record/recordsql/gapstate.go:120-153` | gapMarkers | MIXED | 4 | chain arm deleted · X22 | mint_supersedes marker (128-129), demand_lineage heir arm (143) + comment; own-marker arm stays as the gap itself |
| M | `T/record/recordsql/gapstate.go:169-337` | GapRefreshViewDDL | MIXED | 28 | chain arm deleted · X22 | supersede_event arm+k CTE; lineage join; met arm 2 from_gap, arm 4 reduced to own marker; cols supersede_event/supersedes_count/superseded_by |
| M | `T/record/recordsql/schema.go:245-247` | QueryIndexDDL/index:mint_about | O | 0 | O | comment names demand_reach view, which stays |
| M | `T/record/recordsql/schema.go:282-284` | QueryIndexDDL/index:mint_supersedes_value | C-core | 3 | deleted · X22 | index on mint_supersedes(value,event_id) |
| M | `T/record/recordsql/store.go:743-767` | Insert | C-text | 3 | rewritten · X20 | comment's example is a mint's two supersedes rows in mint_supersedes; needs another list example |
| M | `T/record/recordsql/views.go:357-457` | ViewsDDL/view:gap | MIXED | 5 | chain arm deleted · X22 | doc 391-392 + cols supersede_event, supersedes_count, superseded_by |
| M | `T/record/recordsql/views.go:742-754` | demandReachSQL | MIXED | 5 | chain arm deleted · X22 | lineage param, from_gap, JOIN via lineage, doc; seat already e.seat_id on main; lineage arg also drives join order (planguard baseline moves) |
| M | `T/record/recordsql/views.go:779-791` | wholeDeliverySQL | O | 0 | O | comment analogy naming demandReachSQL, which stays |
| M | `T/record/viewjson.go:81-201` | GapJSON | MIXED | 5 | chain arm deleted · X22 | Supersedes []string + SupersededBy field/doc (180-184) |
| M | `T/record/viewjson.go:367-375` | BoardJSONOfRun | C-text | 1 | rewritten · X20 | 'list tables answer found_by/supersedes' |
| M | `T/record/viewjson.go:377-391` | boardView | MIXED | 1 | chain arm deleted · X22 | EVENT_TYPE_MINT loaded only for a successor's mint as closing act |
| M | `T/record/viewjson.go:393-560` | boardJSONOfRecord | MIXED | 3 | chain arm deleted · X22 | listValuesByEvent(mint_supersedes), Supersedes: strs(...), SupersededBy assign |
| M | `T/record/viewjson.go:605-653` | closeStateOf | MIXED | 12 | chain arm deleted · X22 | supersede arm 629-639 + condition 612 |
| M | `T/record/viewjson.go:803-946` | WorkGapJSON | O | 0 | O | YourTurn: the dispatch engages this reader (not successors) |
| M | `T/record/viewjson.go:948-1015` | EstoppedJSON | C-text | 1 | rewritten · X20 | 'successor R2-1' (Close.successor regression) in a measurement note |
| M | `T/record/viewjson.go:1031-1105` | WorkGapState | MIXED | 1 | chain arm deleted · X22 | Supersedes []string (shares line with FoundBy); 1082 'engages' O |
| M | `T/record/viewjson.go:1107-1127` | workReads | MIXED | 1 | chain arm deleted · X22 | supersedes map (shares line with foundBy) |
| M | `T/record/viewjson.go:1160-1249` | workReadsAt | MIXED | 3 | chain arm deleted · X22 | listValuesByEvent(mint_supersedes) read + err block |
| M | `T/record/viewjson.go:1251-1308` | workGapStatesOf | MIXED | 1 | chain arm deleted · X22 | g.Supersedes assignment (shares line with FoundBy) |
| M | `T/record/viewjson.go:1377-1441` | workJSONOfGaps | MIXED | 2 | chain arm deleted · X22 | 'inherits:'+Join(Supersedes) arm |
| M | `T/record/viewjson.go:1553-1614` | WorkOfSeat | O | 0 | O | engagedOf: blue's dispatch engagement, stays |
| B | `T/record/recordpb/record.proto:172-174` | (file comment: message checks) | C-regression | 2 | deleted · X8 | example 'repaired_with_regression requires a successor'; goes only with Close.successor |
| B | `T/record/recordpb/record.proto:301` | EventType.EVENT_TYPE_RETIRE | O | 0 | O | 'what supersedes it': a retired claim's replacement (Retire.superseded_by) |
| B | `T/record/recordpb/record.proto:513` | Disposition.DISPOSITION_REPAIRED_WITH_REGRESSION | C-regression | 1 | deleted · X8 | (means) REQUIRES a successor naming the regression's gap |
| B | `T/record/recordpb/record.proto:514` | Disposition.DISPOSITION_AMENDS_PRIOR | C-text | 1 | rewritten · X20 | (means) defines its lineage as 'the supersedes the gap was minted with'; value stays, means rewritten |
| B | `T/record/recordpb/record.proto:891-894` | RepairRegression.lineage_mints | C-core | 4 | deleted · X22 | count of mints with supersedes (view.go telemetry); ratio = lineage/closures goes with it |
| B | `T/record/recordpb/record.proto:999-1002` | Mint.supersedes | C-core | 4 | deleted · X22 | field 27 + doc; 27 joins the reserved list (1003-1005 stay) |
| B | `T/record/recordpb/record.proto:1035-1042` | Close.(check) | C-regression | 8 | deleted · X8 | 'lineage never drops' check: regression closure must name successor |
| B | `T/record/recordpb/record.proto:1078-1083` | Close.successor | C-regression | 6 | deleted · X8 | field 7 + doc |
| B | `T/record/recordpb/record.proto:1226-1227` | AboutKind.(reserved 3) | C-text | 1 | rewritten · X20 | 'a restatement of its defect is a mint superseding it' |
| B | `T/record/recordpb/record.proto:1645-1646` | QuestionPart.supersedes_coverage | O | 0 | O | marks a coverage record rather than a proposal |
| B | `T/record/recordpb/record.proto:1690-1691` | Inquiry.supersedes_status | O | 0 | O | marks a fate record rather than a proposal |
| B | `T/record/recordpb/record.proto:1773-1776` | Retire.reason | O | 0 | O | 'superseded' as a reason a claim leaves the report |
| B | `T/record/recordpb/record.proto:1777` | Retire.superseded_by | O | 0 | O | a retired claim superseded by another claim |
| B | `T/record/recordpb/record.proto:1862` | LogType.LOG_TYPE_ESTOPPEL | C-text | 1 | rewritten · X20 | (means) 'or mint with --supersedes so the lineage is explicit' |
| B | `T/record/recordpb/record.proto:1940` | Motion.supersede | C-proposal | 1 | deleted · X1/X3/X4 | oneof act arm Mint supersede = 17 (reserve 17) |
| B | `T/record/recordpb/record.proto:2564` | Dispatch.(doc) | O | 0 | O | 'lineage none': empty gap_ids, not unknown |
| M | `T/record/recordpb/record.proto:172-174` | (file comment: message checks) | C-regression | 2 | deleted · X8 | example 'repaired_with_regression requires a successor' |
| M | `T/record/recordpb/record.proto:298` | EventType.EVENT_TYPE_RETIRE | O | 0 | O | 'what supersedes it': a retired claim's replacement |
| M | `T/record/recordpb/record.proto:508` | Disposition.DISPOSITION_REPAIRED_WITH_REGRESSION | C-regression | 1 | deleted · X8 | (means) REQUIRES a successor |
| M | `T/record/recordpb/record.proto:509` | Disposition.DISPOSITION_AMENDS_PRIOR | C-text | 1 | rewritten · X20 | (means) lineage = 'the supersedes the gap was minted with' |
| M | `T/record/recordpb/record.proto:893-896` | RepairRegression.lineage_mints | C-core | 4 | deleted · X22 | count of mints with supersedes; ratio goes with it |
| M | `T/record/recordpb/record.proto:1007-1010` | Mint.supersedes | C-core | 4 | deleted · X22 | repeated field 20 (table mint_supersedes) + doc |
| M | `T/record/recordpb/record.proto:1042-1049` | Close.(check) | C-regression | 8 | deleted · X8 | 'lineage never drops' check |
| M | `T/record/recordpb/record.proto:1085-1090` | Close.successor | C-regression | 6 | deleted · X8 | field 7 + doc |
| M | `T/record/recordpb/record.proto:1662-1663` | QuestionPart.supersedes_coverage | O | 0 | O | coverage record vs proposal |
| M | `T/record/recordpb/record.proto:1707-1708` | Inquiry.supersedes_status | O | 0 | O | fate record vs proposal |
| M | `T/record/recordpb/record.proto:1811-1814` | Retire.reason | O | 0 | O | 'superseded' as a retire reason |
| M | `T/record/recordpb/record.proto:1815` | Retire.superseded_by | O | 0 | O | a retired claim superseded by another claim |
| M | `T/record/recordpb/record.proto:1900` | LogType.LOG_TYPE_ESTOPPEL | C-text | 1 | rewritten · X20 | (means) 'mint with --supersedes so the lineage is explicit' |
| M | `T/record/recordpb/record.proto:2539` | Dispatch.(doc) | O | 0 | O | 'lineage none': empty gap_ids |

**Group D** (`amendv/tag-D.tsv`)

| t | file:lines | declaration | tag | del | change | note |
|---|---|---|---|---|---|---|
| B | `F/tools/devcmd/enginegen/vocabulary.go:299-343` | (vocabulary).renderParties | O | 0 | O | engaged: chair dispatch engages a party on gaps (Plan party shape), not successors |
| B | `T/capture/capture.go:1830-2062` | Run | O | 0 | O | ancestors: parent directories of a path; engaged: dispatches engaging a lens (lens-turn audit) |
| B | `T/capture/dispatchparity.go:15-218` | DispatchParityAudit | O | 0 | O | engaged: last dispatching chair sitting engaged the bench (dispatch parity) |
| B | `T/capture/lensturn.go:10-30` | LensTurnAudit | O | 0 | O | engaged/engagement: dispatch engaging a lens before blue acted (lens-turn audit) |
| B | `T/cli/blue/command.go:19-33` | Verbs | MIXED | 1 | chain arm deleted · X22; its engage part kept, restated · X22a | other job: registers every blue verb; newEngage() line goes (C-engage) |
| B | `T/cli/blue/engage.go:16-59` | newEngage | C-engage | 44 | kept, restated · X22a | blue engage verb: successors of gaps found closed, record.requireEngage |
| B | `T/cli/blue/engage.go:61-64` | engageResult | C-engage | 4 | kept, restated · X22a | engage verb result type |
| B | `T/cli/blue/engage.go:66-68` | (engageResult).Human | C-engage | 3 | kept, restated · X22a | engage verb confirmation text |
| B | `T/cli/blue/freeze.go:32-59` | requireRoadsHold | MIXED | 3 | chain arm deleted · X22 | other job: freeze guard on standing roads; successor-quote wording branch (r.Successor) goes (C-proposal) |
| B | `T/cli/blue/freeze.go:61-81` | roadPlans | MIXED | 3 | chain arm deleted · X22 | other job: plans red-text/retire roads; r.Successor arm + doc clause go (C-proposal L48) |
| B | `T/cli/blue/freeze.go:83-96` | SuccessorPlans | C-proposal | 14 | deleted · X1/X3/X4 | plans a proposed supersede's successor quote/fix (L48) |
| B | `T/cli/blue/freeze.go:13` | - | C-text | 1 | rewritten · X20 | freeze package doc: 'standing supersede proposal's stored quote and fix' clause (C-proposal) |
| B | `T/cli/blue/inquiry.go:122-172` | newInquiryRecord | O | 0 | O | SupersedesStatus: an inquiry fate record supersedes its prior status row |
| B | `T/cli/blue/question.go:51-74` | newQuestionCover | O | 0 | O | SupersedesCoverage: a question-part coverage row supersedes the prior coverage |
| B | `T/cli/blue/retire.go:24-167` | newRetire | O | 0 | O | Retire.superseded_by: a retired CLAIM superseded by replacement text (--new) |
| B | `T/cli/capture.go:15-61` | newCapture | O | 0 | O | ancestors: parent directories in capture's path resolution |
| B | `T/cli/enumhelp/enumhelp.go:28` | - | C-regression | 2 | deleted · X8 | enumhelp pkg doc: repaired_with_regression requires --successor (regression flag); stays unless regression goes |
| B | `T/cli/lens/close.go:16-77` | newClose | C-regression | 0 | deleted · X8 | comment 'a close to a successor included' = repaired_with_regression close; newClose stays |
| B | `T/cli/lens/closing.go:16` | - | C-text | 1 | rewritten · X20 | closing.go doc: 'or a supersede proposed on it' (C-proposal); red-brings-text doc stays |
| B | `T/cli/lens/mint.go:25-116` | newMint | MIXED | 17 | chain arm deleted · X22 | other job: mint verb; AncestorID var, p.Supersedes set, SupersessionAdmitted block, --supersedes flag + key-help mention go |
| B | `T/cli/lens/mint.go:384-402` | mintAsserted | C-text | 1 | rewritten · X20 | other job: retry-assertion table; supersedes row goes |
| B | `T/cli/lens/mint.go:409-425` | requireNotEstopped | MIXED | 4 | chain arm deleted · X22 | other job: estoppel guard; supersedes param + 'unless the mint supersedes' exemption go (exemption must be re-expressed for amend) |
| B | `T/cli/lens/mint.go:441-454` | aboutGapRefusal | C-text | 3 | rewritten · X20 | refusal names 'motion docket file supersede' and '--supersedes <gap>'; --about-kind gap refusal stays, needs amend wording |
| B | `T/cli/lens/mint.go:456-546` | WriteMint | MIXED | 7 | chain arm deleted · X22 | other job: mint write + estoppel log; accept-of-supersede doc, --supersedes comments, GetSupersedes guard arg, msg 'mint with --supersedes' go |
| B | `T/cli/lens/mint.go:548-552` | Statement | C-text | 2 | rewritten · X20 | doc: Statement shared with 'motion docket file supersede' (C-proposal); type stays for mint |
| B | `T/cli/lens/mint.go:564-602` | (*Statement).Fill | C-text | 1 | rewritten · X20 | doc 'ancestor' is the writer's; Fill stays |
| B | `T/cli/motion/command.go:38-77` | NewCommandFor | C-text | 1 | rewritten · X20 | docket subject Short: 'or a supersede' (C-proposal) |
| B | `T/cli/motion/docketrule.go:24-36` | docketAnswers | C-text | 3 | rewritten · X20 | docket answer help rows: 'close or supersede', 'a supersede's successor' (C-proposal); table stays |
| B | `T/cli/motion/docketrule.go:38-45` | docketRuleSays | C-text | 2 | rewritten · X20 | docket rule page: supersede-of-closed-gap side, 'a supersede's successor' (C-proposal) |
| B | `T/cli/motion/docketrule.go:172-196` | applyAnswer | MIXED | 2 | chain arm deleted · X22 | other job: applies bench answer; doc clause 'held supersede's successor' + *recordpb.Mint case arm go (C-proposal) |
| B | `T/cli/motion/propose.go:24-28` | proposeSays | C-text | 3 | rewritten · X20 | proposal page: supersede act, 'closed gap only as a supersede', 'gap a successor superseded' sentence (C-proposal+C-core) |
| B | `T/cli/motion/propose.go:30-112` | newDocketFile | MIXED | 27 | chain arm deleted · X22 | other job: docket file group (regrade, close); whole supersede subcommand goes (C-proposal) |
| B | `T/cli/motion/propose.go:114-176` | propose | MIXED | 13 | chain arm deleted · X22 | other job: builds every proposal verb; L48 successor-quote guard block goes (C-proposal) |
| B | `T/cli/motion/propose.go:178-224` | newAccept | C-text | 1 | rewritten · X20 | accept page: 'or the successor the proposal states' (C-proposal) |
| B | `T/cli/motion/propose.go:226-265` | applyAccepted | MIXED | 17 | chain arm deleted · X22 | other job: applies accepted regrade/close; *recordpb.Mint (successor) case + doc goes (C-proposal) |
| B | `T/cli/motion/propose.go:19` | - | C-text | 1 | rewritten · X20 | propose.go doc: 'regrade, close or supersede' (C-proposal) |
| B | `T/cli/motion/verbs.go:18-24` | fileSubjectSays | C-text | 1 | rewritten · X20 | file subject page: 'or its supersede' as proposed act (C-proposal) |
| B | `T/cli/seat/answer.go:18-27` | KeysOn | O | 0 | O | superseders: record table of acts answering a refused repeated act |
| B | `T/cli/seat/answer.go:35-72` | teachAnswer | O | 0 | O | superseders: record table of acts answering a refused repeated act |
| B | `T/cli/seat/closure.go:18-25` | ClosureFlags | C-regression | 1 | deleted · X8 | --superseded-by flag registration (repaired_with_regression successor); --id/--as stay |
| B | `T/cli/seat/closure.go:27-57` | ClosurePayload | C-regression | 1 | deleted · X8 | Close.successor from --superseded-by; closure payload stays |
| B | `T/cli/seat/seat.go:1147-1168` | OptStr | C-regression | 5 | deleted · X8 | OptStr doc: successor='' FK case (regression successor); OptStr has other callers |
| B | `T/cli/seat/selector.go:280-350` | selectJSONArrays | O | 0 | O | superseded: an earlier inquiry fate's words; engaged: work JSON null example |
| B | `T/cli/seat/verbs.go:245-320` | views | O | 0 | O | superseded: earlier certify statements; Retire.superseded_by of a claim |
| B | `T/cli/seat/verbs.go:322-357` | workPage | C-text | 2 | rewritten · X20; its engage part kept, restated · X22a | workPage: engaged para found_closed successors/successors_owed/engage (C-engage); estopped 'successor a supersede ... left' (C-proposal); regression row stays |
| B | `T/cli/surface.go:161-205` | CommandReferences | C-regression | 1 | deleted · X8 | comment example 'superseded-by': gap-id (regression flag); CommandReferences stays |
| B | `T/consistency/consistency.go:42-66` | gtGap | MIXED | 1 | chain arm deleted · X22 | other job: ground-truth gap; supersedes field goes |
| B | `T/consistency/consistency.go:94-300` | walk | MIXED | 23 | chain arm deleted · X22 | other job: consistency raw walk; docketSupersede map, supersedes field, mint-closes-ancestor block (C-core), bench supersede-ruling closure arm (C-proposal) go |
| B | `T/consistency/consistency.go:349-565` | Check | MIXED | 7 | chain arm deleted · X22 | other job: consistency oracle; lineage ancestor-exists + cycle block goes (graph node check stays) |
| B | `T/consistency/consistency.go:683-718` | findCycle | C-core | 36 | deleted · X22 | cycle walk over supersedes edges |
| B | `T/consistency/consistency.go:7` | - | C-text | 1 | rewritten · X20 | pkg doc: 'the lineage graph' (graph package's gap flow; renames) |
| B | `T/dashboard/model.go:454-565` | buildJudiciary | MIXED | 30 | chain arm deleted · X22 | other job: judiciary sittings/rulings/gap lives; union-find over supersedes chains goes, chain spans collapse to per-gap |
| B | `T/dashboard/render.go:214-549` | RenderHTML | C-text | 1 | rewritten · X20 | 'argument chains (supersedes-aware)' row (C-core); line 476 'superseded' = a seat attempt replaced by a later one (O) |
| B | `T/debatejs/debatejs.go:154-174` | HeldBy | O | 0 | O | engaged: dispatch plan parties engaged on gaps |
| B | `T/feov/error.go:27-45` | NotFound,MissingField,RoleViolation,Validation,... | C-text | 1 | rewritten · X20 | Conflict doc example 'a lineage that does not hold' |
| B | `T/fetchcache/readscanned.go:123-236` | renderAndReadPages | O | 0 | O | superseded: a stale page render replaced by a reading (fetchcache) |
| B | `T/fetchcache/shell.go:212-233` | TDMReservation | O | 0 | O | engage: a use engaging a machine-readable rights signal (TDM) |
| B | `T/flags/ancestor.go:8-9` | OneAncestorRule | C-core | 2 | deleted · X22 | --supersedes one-ancestor rule text |
| B | `T/flags/ancestor.go:11-14` | OneAncestorRefusal | C-core | 4 | deleted · X22 | --supersedes one-ancestor refusal text |
| B | `T/flags/ancestor.go:16-20` | AncestorValue | C-core | 5 | deleted · X22 | --supersedes flag value type |
| B | `T/flags/ancestor.go:22-23` | AncestorID | C-core | 2 | deleted · X22 | --supersedes flag constructor |
| B | `T/flags/ancestor.go:25-30` | (*AncestorValue).Set | C-core | 6 | deleted · X22 | --supersedes Set refusing a comma |
| B | `T/flags/csv.go:8-24` | CSV | C-text | 5 | rewritten · X20 | CSV doc: '--supersedes G2,G7' and supersedes-chain array rationale; CSV stays (--ids) |
| B | `T/flags/names.go:26-325` | Run,SeatID,For,Occasion,... | MIXED | 3 | chain arm deleted · X22 | other job: flag-name constants; Supersedes const + lineage comment go; SupersededBy stays (regression) |
| B | `T/flags/names.go:327-354` | All | MIXED | 1 | chain arm deleted · X22 | other job: flag census; Supersedes name goes |
| B | `T/flags/names.go:415-451` | payloadFlag | C-regression | 1 | deleted · X8 | 'successor': SupersededBy row = Close.successor (regression); 'superseded_by': New is Retire (O) |
| B | `T/flags/repeated.go:18-31` | (*Repeat).Said | MIXED | 3 | chain arm deleted · X22 | other job: repeated-flag refusal; AncestorValue branch goes |
| B | `T/flags/shapes.go:69-80` | Checker | C-text | 1 | rewritten · X20 | Checker doc lists --supersedes among GapExists users |
| B | `T/flags/shapes.go:19` | - | C-text | 1 | rewritten · X20 | shapes doc example '--supersedes against the board' |
| B | `T/graph/graph.go:191-248` | gapFlowMermaid | MIXED | 6 | chain arm deleted · X22 | other job: gap-flow mermaid; supersedes edges loop + doc go |
| B | `T/graph/graph.go:257-297` | Dot | MIXED | 5 | chain arm deleted · X22 | other job: graphviz gap flow; supersedes dashed edges go |
| B | `T/report/assemble.go:817-874` | withdrawnClaims | O | 0 | O | Retire.superseded_by: a retired claim superseded by new text |
| B | `T/report/assemble.go:891-970` | boardSection | MIXED | 0 | chain arm deleted · X22 | other job: board section; g.Successor() also reads Close.successor (regression), reader stays, Successor() narrows in record |
| B | `T/report/assemble.go:1575-1633` | recordVerification | O | 0 | O | lineage gap: a gap about lineage in #415 history (comment), not the chain |
| B | `T/report/docs.go:115-279` | (readers).assemble | O | 0 | O | superseded: inquiry path records and later certifications |
| B | `T/report/docs.go:522-550` | supersededAsks | O | 0 | O | supersededAsks: earlier bench certify statements a later one replaced |
| B | `T/report/edits.go:11-66` | editHistory | O | 0 | O | superseded: an earlier wording of a report sentence |
| B | `T/report/layout.go:132-232` | layout | C-text | 1 | rewritten · X20 | docket path text 'regrade\|close\|supersede' (C-proposal); Superseded bench statements heading is O |
| B | `T/runlive/runlive.go:420-439` | canonical | O | 0 | O | ancestor: deepest existing parent directory of a path |
| B | `T/seatprobe/boards.go:270-328` | sources | O | 0 | O | superseded: a probe board's claim the report kept stating |
| B | `T/seatprobe/boards.go:330-387` | docket | O | 0 | O | successor defect: red narrative prose in a probe board |
| B | `T/seatprobe/boards.go:389-455` | audit | O | 0 | O | engages: chair dispatch engages the lens |
| B | `T/seatprobe/boards.go:971-1010` | NoSituation | C-text | 7 | rewritten · X20 | NoSituation rows for 'motion docket file supersede' x4 seats (C-proposal); table stays |
| B | `T/seatprobe/boards.go:1025-1036` | AlwaysTaken | C-text | 1 | rewritten · X20; its engage part kept, restated · X22a | AlwaysTaken 'engage' row (C-engage); table stays |
| B | `T/seatprobe/build.go:66-340` | Build | O | 0 | O | engages: chair dispatch engages blue on a gap (exchange) |
| B | `T/seatprobe/build.go:442-454` | exchangeOn | O | 0 | O | engages: chair dispatch engages blue on a gap (exchange) |
| B | `T/seatprobe/production.go:33-88` | backendFor | O | 0 | O | engaged: plan's dispatch parties engaged on board gaps |
| B | `T/verify/verify.go:79-88` | Run | MIXED | 1 | chain arm deleted · X22 | other job: verify check list; supersedesResolve entry goes |
| B | `T/verify/verify.go:188-204` | supersedesResolve | C-core | 17 | deleted · X22 | verify check: every supersedes ancestor exists |
| B | `T/verify/verify.go:298-365` | Compute | C-text | 1 | rewritten · X20 | dispute stat comment 'or a supersede' (C-proposal) |
| B | `T/view/view.go:251-321` | ledgerMD | MIXED | 5 | chain arm deleted · X22 | other job: board ledger; supersedes column goes; Successor() closure-index column stays (regression) |
| B | `T/view/view.go:323-358` | archiveMD | MIXED | 0 | chain arm deleted · X22 | other job: archive; g.Successor() also reads Close.successor (regression), reader stays |
| B | `T/view/view.go:360-567` | telemetryRows | MIXED | 22 | chain arm deleted · X22 | other job: telemetry series; lineage mints, edge deltas over ancestors, repair-regression ratio/LineageMints go |
| B | `T/view/view.go:818-902` | InquiryBody | O | 0 | O | superseded: an inquiry's earlier fate records |
| B | `T/view/view.go:904-930` | superseded | O | 0 | O | superseded: an inquiry's records before its current fate |
| B | `T/view/view.go:983-987` | repairRegJSON | MIXED | 1 | chain arm deleted · X22 | other job: repair-regression JSON; lineage_mints field goes (closures stay) |
| B | `T/view/view.go:994-1031` | telemetryWire | MIXED | 1 | chain arm deleted · X22 | other job: telemetry wire; LineageMints copy goes |
| M | `T/capture/capture.go:1827-2062` | Run | O | 0 | O | ancestors: parent directories; engaged: dispatches engaging a lens |
| M | `T/cli/blue/command.go:19-33` | Verbs | MIXED | 1 | chain arm deleted · X22; its engage part kept, restated · X22a | other job: registers every blue verb; newEngage() line goes (C-engage) |
| M | `T/cli/blue/retire.go:24-167` | newRetire | O | 0 | O | Retire.superseded_by: a retired claim superseded by --new text |
| M | `T/cli/lens/close.go:16-96` | newClose | C-regression | 0 | deleted · X8 | comment 'a close to a successor included' = repaired_with_regression close |
| M | `T/cli/lens/mint.go:25-232` | newMint | MIXED | 16 | chain arm deleted · X22 | other job: mint verb; repeated --supersedes IDList, Supersedes set, SupersessionAdmitted block, flag + comments go; --found-by stays |
| M | `T/cli/lens/mint.go:500-509` | contains | C-core | 10 | deleted · X22 | contains: membership for the --supersedes estoppel exemption only |
| M | `T/cli/lens/mint.go:511-530` | mintAsserted | C-text | 1 | rewritten · X20 | other job: retry-assertion table; supersedes row goes |
| M | `T/cli/lens/mint.go:537-551` | requireNotEstopped | MIXED | 3 | chain arm deleted · X22 | other job: estoppel guard; []supersedes param + exemption go |
| M | `T/cli/motion/verbs.go:20-27` | fileSubjectSays | C-text | 1 | rewritten · X20 | docket page: 'A GAP A SUCCESSOR SUPERSEDED IS NOT DOCKETED... docket the successor' (C-core) |
| M | `T/cli/seat/answer.go:34-71` | teachAnswer | O | 0 | O | superseders: record table of acts answering a refused repeated act |
| M | `T/cli/seat/closure.go:18-25` | ClosureFlags | C-regression | 1 | deleted · X8 | --superseded-by flag registration (regression successor) |
| M | `T/cli/seat/closure.go:27-57` | ClosurePayload | C-regression | 1 | deleted · X8 | Close.successor from --superseded-by |
| M | `T/cli/seat/selector.go:280-350` | selectJSONArrays | O | 0 | O | superseded: earlier inquiry fate words; engaged: work JSON null example |
| M | `T/cli/seat/verbs.go:237-313` | views | O | 0 | O | superseded: earlier certify statements; Retire.superseded_by of a claim |
| M | `T/cli/seat/verbs.go:315-352` | workPage | C-text | 2 | rewritten · X20; its engage part kept, restated · X22a | workPage: engaged para successors/successors_owed/engage (C-engage); estopped 'lineage successor naming ruled gap in supersedes' (C-core) |
| M | `T/consistency/consistency.go:42-66` | gtGap | MIXED | 1 | chain arm deleted · X22 | other job: ground-truth gap; supersedes []string field goes |
| M | `T/consistency/consistency.go:96-294` | walk | MIXED | 12 | chain arm deleted · X22 | other job: consistency raw walk; supersedes field + mint-closes-ancestors loop go |
| M | `T/consistency/consistency.go:343-573` | Check | MIXED | 8 | chain arm deleted · X22 | other job: consistency oracle; lineage ancestor-exists + cycle block goes |
| M | `T/consistency/consistency.go:700-735` | findCycle | C-core | 36 | deleted · X22 | cycle walk over supersedes edges |
| M | `T/dashboard/model.go:457-578` | buildJudiciary | MIXED | 30 | chain arm deleted · X22 | other job: judiciary; union-find over supersedes chains goes |
| M | `T/dashboard/render.go:214-553` | RenderHTML | C-text | 1 | rewritten · X20 | 'argument chains (supersedes-aware)' row; 'superseded' at 480 is a seat attempt (O) |
| M | `T/feov/error.go:27-45` | NotFound,MissingField,RoleViolation,Validation,... | C-text | 1 | rewritten · X20 | Conflict doc example 'a lineage that does not hold' |
| M | `T/flags/csv.go:8-25` | CSV | C-text | 5 | rewritten · X20 | CSV doc: '--supersedes G2,G7' and supersedes-array rationale; CSV stays (--found-by, --ids) |
| M | `T/flags/names.go:26-314` | Run,SeatID,For,Occasion,... | MIXED | 3 | chain arm deleted · X22 | other job: flag-name constants; Supersedes const + lineage comment go |
| M | `T/flags/names.go:316-342` | All | MIXED | 1 | chain arm deleted · X22 | other job: flag census; Supersedes name goes |
| M | `T/graph/graph.go:180-237` | gapFlowMermaid | MIXED | 6 | chain arm deleted · X22 | other job: gap-flow mermaid; supersedes edges go |
| M | `T/graph/graph.go:246-286` | Dot | MIXED | 5 | chain arm deleted · X22 | other job: graphviz; supersedes edges go |
| M | `T/report/assemble.go:922-1037` | boardSection | MIXED | 0 | chain arm deleted · X22 | other job: board section; g.Successor() also regression, reader stays |
| M | `T/report/docs.go:115-279` | (readers).assemble | O | 0 | O | superseded: inquiry path records and later certifications |
| M | `T/report/layout.go:132-232` | layout | O | 0 | O | Superseded bench statements: earlier certify statements |
| M | `T/seatprobe/boards.go:324-390` | docket | O | 0 | O | successor defect: red narrative prose in a probe board |
| M | `T/seatprobe/boards.go:392-458` | audit | O | 0 | O | engages: chair dispatch engages the lens |
| M | `T/verify/verify.go:79-89` | Run | MIXED | 1 | chain arm deleted · X22 | other job: verify check list; supersedesResolve entry goes |
| M | `T/verify/verify.go:221-239` | supersedesResolve | C-core | 19 | deleted · X22 | verify check: every supersedes ancestor exists |
| M | `T/view/view.go:251-356` | ledgerMD | MIXED | 5 | chain arm deleted · X22 | other job: board ledger; supersedes column goes; Successor() column stays |
| M | `T/view/view.go:395-602` | telemetryRows | MIXED | 22 | chain arm deleted · X22 | other job: telemetry series; lineage mints, edge deltas, LineageMints go |

**Group E** (`amendv/tag-E.tsv`)

| t | file:lines | declaration | tag | del | change | note |
|---|---|---|---|---|---|---|
| B | `F/docs/citations-and-bibliography.md:181` | ## 2. Cite — putting it on the record | C-text | 1 | rewritten · X20 | clause "a close naming a successor included" (close --superseded-by, C-regression); para teaches live-answer demand |
| B | `F/docs/citations-and-bibliography.md:183` | ## 2. Cite — putting it on the record | C-text | 1 | rewritten · X20 | sentence: lens mint superseding a source/computation gap carries its check kind (demand_lineage) |
| B | `F/docs/citations-and-bibliography.md:184` | ## 2. Cite — putting it on the record | C-text | 1 | rewritten · X20 | same sentence: "the demand moves to the successor" (demand_lineage) |
| B | `F/docs/propagation-and-anchoring.md:163` | ## Status | O | 0 | O | successor = replacement metric (unminted_findings replaced undisposed metric #327) |
| B | `F/docs/record-flow.md:12` | ## A wrong act | O | 0 | O | superseders table (internal/record/supersede.go): the act that answers a refused repeated act |
| B | `F/docs/record-flow.md:56` | ## Invariants the diagram encodes | C-text | 1 | rewritten · X20 | a supersede listed among acts a disputing lens proposes (C-proposal teaching); bullet stays for one-gap-per-class |
| B | `F/docs/record-flow.md:87` | ## Out of scope (separate concepts) | C-text | 1 | rewritten · X20 | #62 pt2 deferral names `supersedes` as chair de-editorialising/dedup mechanism; pointer rewritten |
| B | `F/docs/seat-command-triggers.md:67` | ## Merge \| lens close row | C-text | 1 | rewritten · X20 | row stays (lens close); clause "unless it names its successor with --superseded-by" is C-regression |
| B | `F/docs/seat-command-triggers.md:89` | ## Blue \| blue engage row | C-engage | 1 | kept, restated · X22a | whole row: blue engage on successors, envelope successors[] (R2) |
| B | `F/docs/seat-command-triggers.md:112` | ## Bench \| motion docket file supersede row | C-proposal | 1 | deleted · X1/X3/X4 | whole row: supersede docket proposal; also teaches --supersedes for assignee |
| B | `F/docs/seat-command-triggers.md:217` | ## Status of the collapses | O | 0 | O | no successor vocabulary: no replacement vocabulary for checked-held after observe/dispose retired |
| B | `F/docs/seat-surface-naming.md:313` | ## Every claim, checked at the leaf | O | 0 | O | historical word counts of an old prompt/constitution (supersedes, lineage counted); measurement record |
| B | `F/docs/seat-surface-naming.md:316` | ## Every claim, checked at the leaf | O | 0 | O | historical word counts of an old prompt/constitution (supersedes, lineage counted); measurement record |
| B | `F/docs/seat-surface-naming.md:386` | # Interview 1 (superseded), 2026-08-20 | O | 0 | O | document section superseded by a later interview (doc history) |
| B | `F/docs/seat-surface-naming.md:489` | # Superseded, 2026-08-20 — the scoped tree | O | 0 | O | design superseded by a later model (doc history) |
| B | `F/docs/seat-surface-naming.md:496` | # Superseded, 2026-08-20 — the scoped tree | O | 0 | O | superseded configuration (doc history) |
| B | `F/hooks/README.md:133` | ## SubagentStart | O | 0 | O | engages: dispatch engages a seat on gaps (work-list injection) |
| B | `F/skills/research-protocol/references/report_template.md:182` | ## Superseded bench statements | O | 0 | O | superseded bench certification statements (bench statement replaced by later one) |
| B | `F/skills/research-protocol/references/run-directory.md:30` | # The run directory (tree) | O | 0 | O | superseded asks: bench certifications replaced by a later one |
| B | `F/skills/research-protocol/scripts/debate.js:17` | file header comment (TESTING & MODEL-SELECTION STRATEGY) | C-text | 1 | rewritten · X20 | founding regression "lineage-blind docket (run 3)" names docket chain-following |
| B | `F/skills/research-protocol/scripts/debate.js:204` | TOOL_COMPOSED | O | 0 | O | 'Superseded bench statements' report section (bench certifications) |
| B | `F/skills/research-protocol/scripts/debate.js:393` | PLAN schema (envelope vocabulary) | O | 0 | O | gap_ids: the gaps a party is engaged on (dispatch) [generated-region: enginegen] |
| B | `F/skills/research-protocol/scripts/debate.js:406` | PLAN schema (envelope vocabulary) | O | 0 | O | exchanges: per gap a party is engaged on (dispatch) [generated-region: enginegen] |
| B | `F/skills/research-protocol/scripts/debate.js:530` | BLUE_ENVELOPE found_closed comment | O | 0 | O | found_closed (#868): engaged gaps blue found closed; feeds log line; stays |
| B | `F/skills/research-protocol/scripts/debate.js:532` | BLUE_ENVELOPE found_closed comment | C-text | 1 | rewritten · X20 | clause "and whether blue sits again on a successor" |
| B | `F/skills/research-protocol/scripts/debate.js:533` | BLUE_ENVELOPE.found_closed | O | 0 | O | found_closed field: blue found engaged gap closed (dispatch engagement); stays for log |
| B | `F/skills/research-protocol/scripts/debate.js:534` | BLUE_ENVELOPE.successors comment | C-engage | 1 | reworded · X22a | R2 successors comment |
| B | `F/skills/research-protocol/scripts/debate.js:535` | BLUE_ENVELOPE.successors comment | C-engage | 1 | reworded · X22a | R2 successors comment |
| B | `F/skills/research-protocol/scripts/debate.js:536` | BLUE_ENVELOPE.successors comment | C-engage | 1 | reworded · X22a | R2 successors comment |
| B | `F/skills/research-protocol/scripts/debate.js:537` | BLUE_ENVELOPE.successors comment | C-engage | 1 | reworded · X22a | R2 successors comment |
| B | `F/skills/research-protocol/scripts/debate.js:538` | BLUE_ENVELOPE.successors comment | C-engage | 2 | reworded · X22a | R2 successors comment (incl. next line) |
| B | `F/skills/research-protocol/scripts/debate.js:540` | BLUE_ENVELOPE.successors | C-engage | 1 | kept, restated · X22a | successors envelope field |
| B | `F/skills/research-protocol/scripts/debate.js:868` | lens comment above lastReadClause | C-text | 1 | rewritten · X20 | keeping lineage among minter duties |
| B | `F/skills/research-protocol/scripts/debate.js:877` | lastReadClause | O | 0 | O | engages: nothing engages you (dispatch) |
| B | `F/skills/research-protocol/scripts/debate.js:903` | lensPrompt | O | 0 | O | engaged: gaps the lens is engaged on (dispatch) |
| B | `F/skills/research-protocol/scripts/debate.js:904` | lensPrompt | O | 0 | O | engaged: gaps the lens is engaged on (dispatch) |
| B | `F/skills/research-protocol/scripts/debate.js:906` | lensPrompt | O | 0 | O | engaged: gaps the lens is engaged on (dispatch) |
| B | `F/skills/research-protocol/scripts/debate.js:907` | lensPrompt | C-text | 1 | rewritten · X20 | REOPEN ... carrying that history, "forks the lineage" in mint clause |
| B | `F/skills/research-protocol/scripts/debate.js:910` | lensPrompt | C-text | 1 | rewritten · X20 | THE ASSIGNEE CLOSES, AND LINEAGE IS NEVER DROPPED clause: assignee=chain origin, successor names gap it supersedes, docket follows chains |
| B | `F/skills/research-protocol/scripts/debate.js:920` | bluePrompt | O | 0 | O | You are engaged on: (dispatch) |
| B | `F/skills/research-protocol/scripts/debate.js:926` | bluePrompt | O | 0 | O | gap you are engaged on (dispatch) |
| B | `F/skills/research-protocol/scripts/debate.js:927` | bluePrompt | C-engage | 1 | kept, restated · X22a | its open successors; "Engage your next sitting on each successor ... name it in successors" (found_closed text stays) |
| B | `F/skills/research-protocol/scripts/debate.js:929` | benchPrompt | C-text | 1 | rewritten · X20 | ruling on a gap with a lineage chain, READ THE NAMED ANCESTORS' RECORDS |
| B | `F/skills/research-protocol/scripts/debate.js:1015` | epoch loop (while !stopped()) — blue sittings | C-engage | 1 | reworded · X22a | BLUE SITS AGAIN ON A SUCCESSOR (R2) comment |
| B | `F/skills/research-protocol/scripts/debate.js:1016` | epoch loop (while !stopped()) — blue sittings | C-engage | 1 | reworded · X22a | BLUE SITS AGAIN ON A SUCCESSOR (R2) comment |
| B | `F/skills/research-protocol/scripts/debate.js:1017` | epoch loop (while !stopped()) — blue sittings | C-engage | 1 | reworded · X22a | BLUE SITS AGAIN ON A SUCCESSOR (R2) comment |
| B | `F/skills/research-protocol/scripts/debate.js:1018` | epoch loop (while !stopped()) — blue sittings | C-engage | 1 | reworded · X22a | BLUE SITS AGAIN ON A SUCCESSOR (R2) comment |
| B | `F/skills/research-protocol/scripts/debate.js:1019` | epoch loop (while !stopped()) — blue sittings | C-engage | 1 | reworded · X22a | BLUE SITS AGAIN ON A SUCCESSOR (R2) comment |
| B | `F/skills/research-protocol/scripts/debate.js:1021` | epoch loop (while !stopped()) — blue sittings | C-engage | 1 | kept, restated · X22a | blueEngaged set (only bounds successor re-sitting); also blueSittings.again + p.again args go |
| B | `F/skills/research-protocol/scripts/debate.js:1039` | epoch loop (while !stopped()) — blue sittings | O | 0 | O | foundClosed: engaged gaps blue found closed, feeds logs; its use in `again` goes |
| B | `F/skills/research-protocol/scripts/debate.js:1041` | epoch loop (while !stopped()) — blue sittings | C-engage | 1 | kept, restated · X22a | successors from envelope |
| B | `F/skills/research-protocol/scripts/debate.js:1042` | epoch loop (while !stopped()) — blue sittings | C-engage | 1 | kept, restated · X22a | again = successors blue sits on |
| B | `F/skills/research-protocol/scripts/debate.js:1044` | epoch loop (while !stopped()) — blue sittings | C-engage | 5 | kept, restated · X22a | if (again.length) block (5 lines) |
| B | `F/skills/research-protocol/scripts/debate.js:1046` | epoch loop (while !stopped()) — blue sittings | C-engage | 0 | kept, restated · X22a | counted in again.length block: log of re-sitting on successors |
| B | `F/skills/research-protocol/scripts/debate.js:1048` | epoch loop (while !stopped()) — blue sittings | C-engage | 1 | kept, restated · X22a | unsat successors |
| B | `F/skills/research-protocol/scripts/debate.js:1049` | epoch loop (while !stopped()) — blue sittings | C-engage | 1 | kept, restated · X22a | log of unsat successors |
| B | `F/skills/research-protocol/scripts/debate.js:1054` | epoch loop (while !stopped()) — blue sittings | O | 0 | O | blue found every engaged gap closed (dispatch, #868) |
| B | `F/skills/research-protocol/scripts/debate.py:334` | run_lens (async def) | O | 0 | O | Engaged gaps: chair dispatch engages lens on gaps; debate.py has NO successor/found_closed handling (debate.js does) |
| B | `T/cli/seat/help/carry.md:11` | ## detail | C-core | 1 | deleted · X22 | whole para: carry of a superseded gap refused, naming the successor |
| B | `T/cli/seat/help/close.md:13` | ## detail | C-text | 1 | rewritten · X20 | failed-check para stays; "and every successor of it" + successor-superseding sentences (demand_lineage) rewritten |
| B | `T/cli/seat/help/close.md:17` | ## detail | C-core | 1 | deleted · X22 | whole para: superseded gap already closed (C-core) + 2nd sentence --superseded-by names regression (C-regression) |
| B | `T/cli/seat/help/closing-red.md:9` | ## detail | C-text | 1 | rewritten · X20 | or for a close or a supersede proposed on it (C-proposal); para stays (red provides text) |
| B | `T/cli/seat/help/dispatch.md:19` | ## detail | O | 0 | O | engaged: chair dispatch engages lens on a gap (exchange) |
| B | `T/cli/seat/help/dispatch.md:21` | ## detail | C-text | 1 | rewritten · X20 | proposed regrade, close or supersede — supersede proposal (C-proposal) in proposal-window para |
| B | `T/cli/seat/help/dispatch.md:23` | ## detail | O | 0 | O | engaged: party engaged on gap in plan.exchanges |
| B | `T/cli/seat/help/edit.md:52` | ## detail | C-text | 1 | rewritten · X20 | the quote and fix of a supersede proposed on it (C-proposal); standing-statement para stays |
| B | `T/cli/seat/help/engage.md:3` | ## menu | C-engage | 1 | kept, restated · X22a | whole verb help page; menu line: engage verb on successors |
| B | `T/cli/seat/help/engage.md:7` | ## detail | C-engage | 1 | kept, restated · X22a | whole verb help page; engage detail: successors_owed, superseded_by, supersedes, envelope successors |
| B | `T/cli/seat/help/engage.md:9` | ## detail | C-engage | 1 | kept, restated · X22a | whole verb help page; engage refusals: exactly the successor set owed |
| B | `T/cli/seat/help/mint.md:9` | ## detail | C-text | 1 | rewritten · X20 | one-gap-per-class para stays; sentence "A mint naming the open gap as its ancestor ... is not refused" goes |
| B | `T/cli/seat/help/mint.md:11` | ## detail | C-text | 1 | rewritten · X20 | tombstone para stays; "on the successor a supersede it applied" (C-proposal) + "one naming it as ancestor" go |
| B | `T/cli/seat/help/mint.md:13` | ## detail | C-text | 1 | rewritten · X20 | read-board para stays; "mint it as that gap's successor", REOPEN "carries its lineage" rewritten |
| B | `T/cli/seat/help/mint.md:15` | ## detail | C-core | 1 | deleted · X22 | whole para: mint naming open ancestor closes it; assignee-only supersede; demand_lineage; closed-gap lineage |
| B | `T/cli/seat/help/mint.md:21` | ## detail | C-text | 1 | rewritten · X20 | closure-stands para stays; "naming that gap as the one it supersedes" rewritten |
| B | `T/cli/seat/help/mint.md:25` | ## detail | C-text | 1 | rewritten · X20 | estoppel para stays; "or mint declaring that gap as its ancestor" (estoppel --supersedes escape, mint.go:420) goes |
| B | `T/cli/seat/help/mint.md:31` | ## detail | C-text | 1 | rewritten · X20 | fix/check para stays; "corrected by minting its successor" rewritten; NB "no act amends a check after the mint" = existing amend sense |
| B | `T/cli/seat/help/regrade.md:9` | ## detail | O | 0 | O | engaged: dispatch engaged the lens on the gap (exchange) |
| B | `T/cli/seat/help/regrade.md:11` | ## detail | C-text | 1 | rewritten · X20 | ruled-axis para stays; "a gap you mint superseding this one carries the ruled grade" goes |
| B | `T/cli/seat/help/retire.md:19` | ## detail | C-text | 1 | rewritten · X20 | the quote and fix of a supersede proposed on it (C-proposal); standing-statement para stays |
| B | `T/terms/terms.json:673` | term "assignee" (671-683) | C-text | 1 | rewritten · X20 | definition: assignee = minter of the origin of the gap's lineage, carried by every successor (chain origin); term stays |
| B | `T/terms/terms.json:818` | term "report read" (806-821) | O | 0 | O | engages: the chair engages a party at a dispatch pin |
| B | `T/terms/terms.json:893` | term "grade movement" (880-896) | C-text | 1 | rewritten · X20 | collision resolution: exchange moves gap when "its lineage changes" |
| B | `T/terms/terms.json:1056` | term "reopened" (1042-1063) | C-text | 1 | rewritten · X20 | collision: "minting a successor over a closed gap reopens the gap" |
| B | `T/terms/terms.json:1106` | term "superseded" (1105-1140) | C-core | 36 | deleted · X22 | whole term entry (def, seats, ban, collisions) goes; its O collision rows (retire, inquiry/bench) need a new home |
| B | `T/terms/terms.json:1107` | term "superseded" definition | C-core | 0 | deleted · X22 | counted in term entry: definition of supersede-by-mint, demand passing |
| B | `T/terms/terms.json:1119` | term "superseded" ban stranded gap | C-core | 0 | deleted · X22 | counted in term entry: GATED ban pattern "stranded (superseded )?(gaps?\|ancestors?)" |
| B | `T/terms/terms.json:1124` | term "superseded" collision | C-core | 0 | deleted · X22 | counted in term entry: superseded (a closed gap) = lineage/reopen |
| B | `T/terms/terms.json:1125` | term "superseded" collision | C-core | 0 | deleted · X22 | counted in term entry: mint naming closed gap is lineage; successor is assignee's |
| B | `T/terms/terms.json:1128` | term "superseded" collision | C-regression | 0 | deleted · X8 | counted in term entry: `--superseded-by` (close) collision row |
| B | `T/terms/terms.json:1129` | term "superseded" collision | C-regression | 0 | deleted · X8 | counted in term entry: --superseded-by names regression gap, only with repaired_with_regression |
| B | `T/terms/terms.json:1132` | term "superseded" collision | O | 0 | O | superseded_by (retire): a retired claim's replacement; row must move if entry goes |
| B | `T/terms/terms.json:1136` | term "superseded" collision | O | 0 | O | superseded (inquiry record/bench statement): later record standing in place; row must move if entry goes |
| B | `T/vecembed/vocab.txt:5118` | vocab.txt (BERT WordPiece vocabulary, 30522 tokens, go:embed in tokenize.go) | O | 0 | O | embedding tokenizer word list; not FEOV vocabulary, never edited |
| B | `T/vecembed/vocab.txt:6333` | vocab.txt (BERT WordPiece vocabulary, 30522 tokens, go:embed in tokenize.go) | O | 0 | O | embedding tokenizer word list; not FEOV vocabulary, never edited |
| B | `T/vecembed/vocab.txt:8148` | vocab.txt (BERT WordPiece vocabulary, 30522 tokens, go:embed in tokenize.go) | O | 0 | O | embedding tokenizer word list; not FEOV vocabulary, never edited |
| B | `T/vecembed/vocab.txt:8527` | vocab.txt (BERT WordPiece vocabulary, 30522 tokens, go:embed in tokenize.go) | O | 0 | O | embedding tokenizer word list; not FEOV vocabulary, never edited |
| B | `T/vecembed/vocab.txt:10749` | vocab.txt (BERT WordPiece vocabulary, 30522 tokens, go:embed in tokenize.go) | O | 0 | O | embedding tokenizer word list; not FEOV vocabulary, never edited |
| B | `T/vecembed/vocab.txt:13033` | vocab.txt (BERT WordPiece vocabulary, 30522 tokens, go:embed in tokenize.go) | O | 0 | O | embedding tokenizer word list; not FEOV vocabulary, never edited |
| B | `T/vecembed/vocab.txt:13322` | vocab.txt (BERT WordPiece vocabulary, 30522 tokens, go:embed in tokenize.go) | O | 0 | O | embedding tokenizer word list; not FEOV vocabulary, never edited |
| B | `T/vecembed/vocab.txt:18531` | vocab.txt (BERT WordPiece vocabulary, 30522 tokens, go:embed in tokenize.go) | O | 0 | O | embedding tokenizer word list; not FEOV vocabulary, never edited |
| B | `T/vecembed/vocab.txt:19887` | vocab.txt (BERT WordPiece vocabulary, 30522 tokens, go:embed in tokenize.go) | O | 0 | O | embedding tokenizer word list; not FEOV vocabulary, never edited |
| B | `T/vecembed/vocab.txt:20584` | vocab.txt (BERT WordPiece vocabulary, 30522 tokens, go:embed in tokenize.go) | O | 0 | O | embedding tokenizer word list; not FEOV vocabulary, never edited |
| B | `T/vecembed/vocab.txt:24256` | vocab.txt (BERT WordPiece vocabulary, 30522 tokens, go:embed in tokenize.go) | O | 0 | O | embedding tokenizer word list; not FEOV vocabulary, never edited |
| B | `scripts/agentgen/src/blue-researcher.md:57` | (bullet list, no heading) response-sitting duties | O | 0 | O | engaged: blue dispatched/engaged on gaps |
| B | `scripts/agentgen/src/fragments/lens-audit-and-mint.md:8` | ## How a lens audits, mints and closes | C-text | 1 | rewritten · X20 | live-answer bullet stays; sentence "A successor you mint superseding it ... demand passes to it" (demand_lineage) goes |
| B | `scripts/agentgen/src/fragments/lens-audit-and-mint.md:14` | ## How a lens audits, mints and closes | C-text | 1 | rewritten · X20 | whole "Lineage is mandatory" bullet (chains, docket follows chains; regressed closure = C-regression) |
| B | `scripts/agentgen/src/fragments/lens-duties.md:5` | ## What a lens may not do | C-text | 1 | rewritten · X20 | its supersede proposal (C-proposal) + assignee = lens whose mint began its chain, every successor (C-core) |
| B | `scripts/agentgen/src/fragments/lens-duties.md:9` | ## What a lens may not do | C-text | 1 | rewritten · X20 | regrade, close or supersede of it and "a close or a supersede proposed on it" (C-proposal) x2 |
| B | `scripts/agentgen/src/lead-judge.md:43` | (bench duties bullets) Ancestor demanded reads | C-text | 1 | rewritten · X20 | whole bullet: ruling on a gap with a lineage chain reads named ancestors' closures |
| B | `scripts/agentgen/src/lead-judge.md:131` | (bench duties bullets) certification | O | 0 | O | superseded certification text kept in CHANGELOG.md |
| B | `scripts/agentgen/src/red-chair.md:14` | ## What only the chair does | O | 0 | O | engaged: a lens with no gap engaged (dispatch) |
| B | `scripts/agentgen/src/red-lens-voice.md:15` | ## What you audit | C-text | 1 | rewritten · X20 | problem and fix change only by its minter's successor; "propose its successor"/"mint its successor" (C-core/C-proposal) |
| B | `scripts/agentgen/src/skills/adversarial-audit/SKILL.md:22` | # adversarial-audit | C-text | 1 | rewritten · X20 | grade stability along supersedes chains metric |
| B | `scripts/agentgen/src/skills/adversarial-audit/SKILL.md:26` | # adversarial-audit | O | 0 | O | engages: no gap engages you (dispatch) |
| B | `scripts/agentgen/src/skills/adversarial-audit/SKILL.md:31` | # adversarial-audit | C-text | 1 | rewritten · X20 | the successor you propose on that gap (C-proposal) in voice bullet |
| B | `scripts/agentgen/src/skills/research-protocol/SKILL.md:84` | ## Report structure | O | 0 | O | asks a later certification superseded (bench statements) |
| M | `F/docs/citations-and-bibliography.md:181` | ## 2. Cite — putting it on the record | C-text | 1 | rewritten · X20 | clause "a close naming a successor included" (close --superseded-by, C-regression); para teaches live-answer demand |
| M | `F/docs/citations-and-bibliography.md:183` | ## 2. Cite — putting it on the record | C-text | 1 | rewritten · X20 | sentence: lens mint superseding a source/computation gap carries its check kind (demand_lineage) |
| M | `F/docs/citations-and-bibliography.md:184` | ## 2. Cite — putting it on the record | C-text | 1 | rewritten · X20 | same sentence: "the demand moves to the successor" (demand_lineage) |
| M | `F/docs/finding-markers.md:179` | ## Status | O | 0 | O | superseded drafts of the finding-markers design (doc history) |
| M | `F/docs/propagation-and-anchoring.md:163` | ## Status | O | 0 | O | successor = replacement metric (unminted_findings replaced undisposed metric #327) |
| M | `F/docs/record-flow.md:12` | ## A wrong act | O | 0 | O | superseders table (internal/record/supersede.go): the act that answers a refused repeated act |
| M | `F/docs/record-flow.md:89` | ## Out of scope (separate concepts) | C-text | 1 | rewritten · X20 | #62 pt2 deferral names `supersedes` as chair de-editorialising/dedup mechanism; pointer rewritten |
| M | `F/docs/seat-command-triggers.md:68` | ## Merge \| lens close row | C-text | 1 | rewritten · X20 | row stays (lens close); clause "unless it names its successor with --superseded-by" is C-regression |
| M | `F/docs/seat-command-triggers.md:90` | ## Blue \| blue engage row | C-engage | 1 | kept, restated · X22a | whole row: blue engage on successors, envelope successors[] (R2) |
| M | `F/docs/seat-command-triggers.md:117` | ## Bench \| bench grade ruling row | C-text | 1 | rewritten · X20 | row stays; clause "a successor carries it" (bench-moved axis inherited by successor) |
| M | `F/docs/seat-command-triggers.md:218` | ## Status of the collapses | O | 0 | O | no successor vocabulary: no replacement vocabulary for checked-held after observe/dispose retired |
| M | `F/docs/seat-surface-naming.md:313` | ## Every claim, checked at the leaf | O | 0 | O | historical word counts of an old prompt/constitution (supersedes, lineage counted); measurement record |
| M | `F/docs/seat-surface-naming.md:316` | ## Every claim, checked at the leaf | O | 0 | O | historical word counts of an old prompt/constitution (supersedes, lineage counted); measurement record |
| M | `F/docs/seat-surface-naming.md:386` | # Interview 1 (superseded), 2026-08-20 | O | 0 | O | document section superseded by a later interview (doc history) |
| M | `F/docs/seat-surface-naming.md:489` | # Superseded, 2026-08-20 — the scoped tree | O | 0 | O | design superseded by a later model (doc history) |
| M | `F/docs/seat-surface-naming.md:496` | # Superseded, 2026-08-20 — the scoped tree | O | 0 | O | superseded configuration (doc history) |
| M | `F/hooks/README.md:159` | ## SubagentStart | O | 0 | O | engages: dispatch engages a seat on gaps (work-list injection) |
| M | `F/skills/research-protocol/references/report_template.md:186` | ## Superseded bench statements | O | 0 | O | superseded bench certification statements (bench statement replaced by later one) |
| M | `F/skills/research-protocol/references/run-directory.md:30` | # The run directory (tree) | O | 0 | O | superseded asks: bench certifications replaced by a later one |
| M | `F/skills/research-protocol/scripts/debate.js:17` | file header comment (TESTING & MODEL-SELECTION STRATEGY) | C-text | 1 | rewritten · X20 | founding regression "lineage-blind docket (run 3)" names docket chain-following |
| M | `F/skills/research-protocol/scripts/debate.js:204` | TOOL_COMPOSED | O | 0 | O | 'Superseded bench statements' report section (bench certifications) |
| M | `F/skills/research-protocol/scripts/debate.js:394` | PLAN schema (envelope vocabulary) | O | 0 | O | gap_ids: the gaps a party is engaged on (dispatch) [generated-region: enginegen] |
| M | `F/skills/research-protocol/scripts/debate.js:407` | PLAN schema (envelope vocabulary) | O | 0 | O | exchanges: per gap a party is engaged on (dispatch) [generated-region: enginegen] |
| M | `F/skills/research-protocol/scripts/debate.js:536` | BLUE_ENVELOPE found_closed comment | O | 0 | O | found_closed (#868): engaged gaps blue found closed; feeds log line; stays |
| M | `F/skills/research-protocol/scripts/debate.js:538` | BLUE_ENVELOPE found_closed comment | C-text | 1 | rewritten · X20 | clause "and whether blue sits again on a successor" |
| M | `F/skills/research-protocol/scripts/debate.js:539` | BLUE_ENVELOPE.found_closed | O | 0 | O | found_closed field: blue found engaged gap closed (dispatch engagement); stays for log |
| M | `F/skills/research-protocol/scripts/debate.js:540` | BLUE_ENVELOPE.successors comment | C-engage | 1 | reworded · X22a | R2 successors comment |
| M | `F/skills/research-protocol/scripts/debate.js:541` | BLUE_ENVELOPE.successors comment | C-engage | 1 | reworded · X22a | R2 successors comment |
| M | `F/skills/research-protocol/scripts/debate.js:542` | BLUE_ENVELOPE.successors comment | C-engage | 1 | reworded · X22a | R2 successors comment |
| M | `F/skills/research-protocol/scripts/debate.js:543` | BLUE_ENVELOPE.successors comment | C-engage | 1 | reworded · X22a | R2 successors comment |
| M | `F/skills/research-protocol/scripts/debate.js:544` | BLUE_ENVELOPE.successors comment | C-engage | 2 | reworded · X22a | R2 successors comment (incl. next line) |
| M | `F/skills/research-protocol/scripts/debate.js:546` | BLUE_ENVELOPE.successors | C-engage | 1 | kept, restated · X22a | successors envelope field |
| M | `F/skills/research-protocol/scripts/debate.js:877` | lens comment above lastReadClause | C-text | 1 | rewritten · X20 | keeping lineage among minter duties |
| M | `F/skills/research-protocol/scripts/debate.js:886` | lastReadClause | O | 0 | O | engages: nothing engages you (dispatch) |
| M | `F/skills/research-protocol/scripts/debate.js:908` | lensPrompt | O | 0 | O | engaged: gaps the lens is engaged on (dispatch) |
| M | `F/skills/research-protocol/scripts/debate.js:909` | lensPrompt | O | 0 | O | YOU ARE ENGAGED ON (dispatch) |
| M | `F/skills/research-protocol/scripts/debate.js:911` | lensPrompt | O | 0 | O | ${engaged} interpolation (dispatch) |
| M | `F/skills/research-protocol/scripts/debate.js:912` | lensPrompt | C-text | 1 | rewritten · X20 | REOPEN ... carrying that history, "forks the lineage" |
| M | `F/skills/research-protocol/scripts/debate.js:915` | lensPrompt | C-text | 1 | rewritten · X20 | THE ORIGINATOR CLOSES, AND LINEAGE IS NEVER DROPPED: successor names gaps it supersedes, docket follows chains |
| M | `F/skills/research-protocol/scripts/debate.js:925` | bluePrompt | O | 0 | O | You are engaged on: (dispatch) |
| M | `F/skills/research-protocol/scripts/debate.js:931` | bluePrompt | O | 0 | O | gap you are engaged on (dispatch) |
| M | `F/skills/research-protocol/scripts/debate.js:932` | bluePrompt | C-engage | 1 | kept, restated · X22a | its open successors; "Engage your next sitting on each successor ... name it in successors" (found_closed text stays) |
| M | `F/skills/research-protocol/scripts/debate.js:934` | benchPrompt | C-text | 1 | rewritten · X20 | ruling on a gap with a lineage chain, READ THE NAMED ANCESTORS' RECORDS |
| M | `F/skills/research-protocol/scripts/debate.js:1014` | epoch loop (while !stopped()) — blue sittings | C-engage | 1 | reworded · X22a | BLUE SITS AGAIN ON A SUCCESSOR (R2) comment |
| M | `F/skills/research-protocol/scripts/debate.js:1015` | epoch loop (while !stopped()) — blue sittings | C-engage | 1 | reworded · X22a | BLUE SITS AGAIN ON A SUCCESSOR (R2) comment |
| M | `F/skills/research-protocol/scripts/debate.js:1016` | epoch loop (while !stopped()) — blue sittings | C-engage | 1 | reworded · X22a | BLUE SITS AGAIN ON A SUCCESSOR (R2) comment |
| M | `F/skills/research-protocol/scripts/debate.js:1017` | epoch loop (while !stopped()) — blue sittings | C-engage | 1 | reworded · X22a | BLUE SITS AGAIN ON A SUCCESSOR (R2) comment |
| M | `F/skills/research-protocol/scripts/debate.js:1018` | epoch loop (while !stopped()) — blue sittings | C-engage | 1 | reworded · X22a | BLUE SITS AGAIN ON A SUCCESSOR (R2) comment |
| M | `F/skills/research-protocol/scripts/debate.js:1020` | epoch loop (while !stopped()) — blue sittings | C-engage | 1 | kept, restated · X22a | blueEngaged set (only bounds successor re-sitting); also blueSittings.again + p.again args go |
| M | `F/skills/research-protocol/scripts/debate.js:1038` | epoch loop (while !stopped()) — blue sittings | O | 0 | O | foundClosed: engaged gaps blue found closed, feeds logs; its use in `again` goes |
| M | `F/skills/research-protocol/scripts/debate.js:1040` | epoch loop (while !stopped()) — blue sittings | C-engage | 1 | kept, restated · X22a | successors from envelope |
| M | `F/skills/research-protocol/scripts/debate.js:1041` | epoch loop (while !stopped()) — blue sittings | C-engage | 1 | kept, restated · X22a | again = successors blue sits on |
| M | `F/skills/research-protocol/scripts/debate.js:1043` | epoch loop (while !stopped()) — blue sittings | C-engage | 5 | kept, restated · X22a | if (again.length) block (5 lines) |
| M | `F/skills/research-protocol/scripts/debate.js:1045` | epoch loop (while !stopped()) — blue sittings | C-engage | 0 | kept, restated · X22a | counted in again.length block: log of re-sitting on successors |
| M | `F/skills/research-protocol/scripts/debate.js:1047` | epoch loop (while !stopped()) — blue sittings | C-engage | 1 | kept, restated · X22a | unsat successors |
| M | `F/skills/research-protocol/scripts/debate.js:1048` | epoch loop (while !stopped()) — blue sittings | C-engage | 1 | kept, restated · X22a | log of unsat successors |
| M | `F/skills/research-protocol/scripts/debate.js:1053` | epoch loop (while !stopped()) — blue sittings | O | 0 | O | blue found every engaged gap closed (dispatch, #868) |
| M | `F/skills/research-protocol/scripts/debate.py:487` | run_lens (async def) | O | 0 | O | Engaged gaps: chair dispatch engages lens on gaps; debate.py has NO successor/found_closed handling (debate.js does) |
| M | `T/cli/seat/help/carry.md:11` | ## detail | C-core | 1 | deleted · X22 | whole para: carry of a superseded gap refused, naming the successor |
| M | `T/cli/seat/help/close.md:13` | ## detail | C-text | 1 | rewritten · X20 | failed-check para stays; "every successor of it" + successor-superseding sentences (demand_lineage) rewritten |
| M | `T/cli/seat/help/close.md:17` | ## detail | C-core | 1 | deleted · X22 | whole para: superseded gap already closed (C-core) + 2nd sentence --superseded-by names regression (C-regression) |
| M | `T/cli/seat/help/dispatch.md:19` | ## detail | O | 0 | O | engaged: chair dispatch engages lens on a gap (exchange) |
| M | `T/cli/seat/help/dispatch.md:23` | ## detail | O | 0 | O | engaged: party engaged on gap in plan.exchanges |
| M | `T/cli/seat/help/engage.md:3` | ## menu | C-engage | 1 | kept, restated · X22a | whole verb help page; menu line: engage verb on successors |
| M | `T/cli/seat/help/engage.md:7` | ## detail | C-engage | 1 | kept, restated · X22a | whole verb help page; engage detail: successors_owed, superseded_by, supersedes, envelope successors |
| M | `T/cli/seat/help/engage.md:9` | ## detail | C-engage | 1 | kept, restated · X22a | whole verb help page; engage refusals: exactly the successor set owed |
| M | `T/cli/seat/help/mint.md:9` | ## detail | C-text | 1 | rewritten · X20 | one-gap-per-class para stays; "A mint naming the open gap as its ancestor ... is not refused" goes |
| M | `T/cli/seat/help/mint.md:11` | ## detail | C-text | 1 | rewritten · X20 | read-board para stays; "mint it as that gap's successor", REOPEN "carries its lineage" rewritten |
| M | `T/cli/seat/help/mint.md:13` | ## detail | C-core | 1 | deleted · X22 | whole para: mint naming open ancestor(s) closes them; minter-only supersede; demand_lineage; separate mints per check kind (repeated) |
| M | `T/cli/seat/help/mint.md:19` | ## detail | C-text | 1 | rewritten · X20 | closure-stands para stays; "naming that gap as the one it supersedes" rewritten |
| M | `T/cli/seat/help/mint.md:25` | ## detail | C-text | 1 | rewritten · X20 | estoppel para stays; "or mint declaring that gap as its ancestor" (estoppel --supersedes escape, mint.go:420) goes |
| M | `T/cli/seat/help/mint.md:31` | ## detail | C-text | 1 | rewritten · X20 | fix/check para stays; "corrected by minting its successor" rewritten; NB "no act amends a check after the mint" = existing amend sense |
| M | `T/cli/seat/help/regrade-bench.md:9` | ## detail | C-text | 1 | rewritten · X20 | bench-axis para stays; "a gap minted superseding this one carries your grade" goes |
| M | `T/cli/seat/help/regrade.md:9` | ## detail | O | 0 | O | engaged: dispatch engaged the lens on the gap (exchange) |
| M | `T/cli/seat/help/regrade.md:11` | ## detail | C-text | 1 | rewritten · X20 | ruled-axis para stays; "a gap you mint superseding this one carries the ruled grade" goes |
| M | `T/terms/terms.json:859` | term "report read" (806-821) | O | 0 | O | engages: the chair engages a party at a dispatch pin |
| M | `T/terms/terms.json:934` | term "grade movement" (880-896) | C-text | 1 | rewritten · X20 | collision resolution: exchange moves gap when "its lineage changes" |
| M | `T/terms/terms.json:1092` | term "reopened" (1042-1063) | C-text | 1 | rewritten · X20 | collision: "minting a successor over a closed gap reopens the gap" |
| M | `T/terms/terms.json:1142` | term "superseded" (1141-1176) | C-core | 36 | deleted · X22 | whole term entry (def, seats, ban, collisions) goes; its O collision rows (retire, inquiry/bench) need a new home |
| M | `T/terms/terms.json:1143` | term "superseded" definition | C-core | 0 | deleted · X22 | counted in term entry: definition (own lens mints successor with --supersedes) |
| M | `T/terms/terms.json:1155` | term "superseded" ban stranded gap | C-core | 0 | deleted · X22 | counted in term entry: GATED ban pattern "stranded (superseded )?(gaps?\|ancestors?)" |
| M | `T/terms/terms.json:1160` | term "superseded" collision | C-core | 0 | deleted · X22 | counted in term entry: superseded (a closed gap) = lineage/reopen |
| M | `T/terms/terms.json:1161` | term "superseded" collision | C-core | 0 | deleted · X22 | counted in term entry: mint naming closed gap is lineage |
| M | `T/terms/terms.json:1164` | term "superseded" collision | C-regression | 0 | deleted · X8 | counted in term entry: `--superseded-by` (close) collision row |
| M | `T/terms/terms.json:1165` | term "superseded" collision | C-regression | 0 | deleted · X8 | counted in term entry: --superseded-by names regression gap, only with repaired_with_regression |
| M | `T/terms/terms.json:1168` | term "superseded" collision | O | 0 | O | superseded_by (retire): a retired claim's replacement; row must move if entry goes |
| M | `T/terms/terms.json:1172` | term "superseded" collision | O | 0 | O | superseded (inquiry record/bench statement): later record standing in place; row must move if entry goes |
| M | `T/vecembed/vocab.txt:5118` | vocab.txt (BERT WordPiece vocabulary, 30522 tokens, go:embed in tokenize.go) | O | 0 | O | embedding tokenizer word list; not FEOV vocabulary, never edited |
| M | `T/vecembed/vocab.txt:6333` | vocab.txt (BERT WordPiece vocabulary, 30522 tokens, go:embed in tokenize.go) | O | 0 | O | embedding tokenizer word list; not FEOV vocabulary, never edited |
| M | `T/vecembed/vocab.txt:8148` | vocab.txt (BERT WordPiece vocabulary, 30522 tokens, go:embed in tokenize.go) | O | 0 | O | embedding tokenizer word list; not FEOV vocabulary, never edited |
| M | `T/vecembed/vocab.txt:8527` | vocab.txt (BERT WordPiece vocabulary, 30522 tokens, go:embed in tokenize.go) | O | 0 | O | embedding tokenizer word list; not FEOV vocabulary, never edited |
| M | `T/vecembed/vocab.txt:10749` | vocab.txt (BERT WordPiece vocabulary, 30522 tokens, go:embed in tokenize.go) | O | 0 | O | embedding tokenizer word list; not FEOV vocabulary, never edited |
| M | `T/vecembed/vocab.txt:13033` | vocab.txt (BERT WordPiece vocabulary, 30522 tokens, go:embed in tokenize.go) | O | 0 | O | embedding tokenizer word list; not FEOV vocabulary, never edited |
| M | `T/vecembed/vocab.txt:13322` | vocab.txt (BERT WordPiece vocabulary, 30522 tokens, go:embed in tokenize.go) | O | 0 | O | embedding tokenizer word list; not FEOV vocabulary, never edited |
| M | `T/vecembed/vocab.txt:18531` | vocab.txt (BERT WordPiece vocabulary, 30522 tokens, go:embed in tokenize.go) | O | 0 | O | embedding tokenizer word list; not FEOV vocabulary, never edited |
| M | `T/vecembed/vocab.txt:19887` | vocab.txt (BERT WordPiece vocabulary, 30522 tokens, go:embed in tokenize.go) | O | 0 | O | embedding tokenizer word list; not FEOV vocabulary, never edited |
| M | `T/vecembed/vocab.txt:20584` | vocab.txt (BERT WordPiece vocabulary, 30522 tokens, go:embed in tokenize.go) | O | 0 | O | embedding tokenizer word list; not FEOV vocabulary, never edited |
| M | `T/vecembed/vocab.txt:24256` | vocab.txt (BERT WordPiece vocabulary, 30522 tokens, go:embed in tokenize.go) | O | 0 | O | embedding tokenizer word list; not FEOV vocabulary, never edited |
| M | `scripts/agentgen/src/blue-researcher.md:58` | (bullet list, no heading) response-sitting duties | O | 0 | O | engaged: blue dispatched/engaged on gaps |
| M | `scripts/agentgen/src/fragments/lens-audit-and-mint.md:8` | ## How a lens audits, mints and closes | C-text | 1 | rewritten · X20 | live-answer bullet stays; successor-superseding sentence (demand_lineage) goes |
| M | `scripts/agentgen/src/fragments/lens-audit-and-mint.md:14` | ## How a lens audits, mints and closes | C-text | 1 | rewritten · X20 | whole "Lineage is mandatory" bullet (chains, docket follows chains; regressed closure = C-regression) |
| M | `scripts/agentgen/src/lead-judge.md:43` | (bench duties bullets) Ancestor demanded reads | C-text | 1 | rewritten · X20 | whole bullet: ruling on a gap with a lineage chain reads named ancestors' closures |
| M | `scripts/agentgen/src/lead-judge.md:44` | (bench duties bullets) remand/dispositions | C-text | 1 | rewritten · X20 | Two of them require a successor or a lineage (repaired_with_regression=C-regression, amends_prior lineage) |
| M | `scripts/agentgen/src/lead-judge.md:45` | (bench duties bullets) set grade before ruling | C-text | 1 | rewritten · X20 | a successor gap carries it (bench-moved axis) |
| M | `scripts/agentgen/src/lead-judge.md:131` | (bench duties bullets) certification | O | 0 | O | superseded certification text kept in CHANGELOG.md |
| M | `scripts/agentgen/src/red-chair.md:14` | ## What only the chair does | O | 0 | O | engaged: lens with no gap engaged (dispatch) |
| M | `scripts/agentgen/src/red-lens-voice.md:15` | ## What you audit | C-text | 1 | rewritten · X20 | change only by its minter's successor; "mint its successor in the reader's voice" |
| M | `scripts/agentgen/src/skills/adversarial-audit/SKILL.md:22` | # adversarial-audit | C-text | 1 | rewritten · X20 | grade stability along supersedes chains metric |
| M | `scripts/agentgen/src/skills/adversarial-audit/SKILL.md:26` | # adversarial-audit | O | 0 | O | engages: no gap engages you (dispatch) |
| M | `scripts/agentgen/src/skills/research-protocol/SKILL.md:91` | ## Report structure | O | 0 | O | asks a later certification superseded (bench statements) |

**Group W** — the widened regex's hits in no tag row (`p15/wide-<sha>.txt`; `repair_regression|lineage_mints|EdgeDeltas|edge_deltas|"supersedes"|\.Supersedes\b`, tests and generated excluded)

| t | file:lines | declaration | tag | del | change | note |
|---|---|---|---|---|---|---|
| B, M | `scripts/agentgen/src/lead-judge.md:110` | bench's weighing duties | C-text | 3 | rewritten · X8 | teaches `repair_regression.ratio` "already measured": the sentence keeps the estoppel record's measure and drops the ratio |
| B, M | `T/record/recordpb/record.proto:843 \| :845` | `TelemetryLine.repair_regression = 8` | C-regression | 1 | reserved · X8 | lineage-derived telemetry |
| B, M | `T/record/recordpb/record.proto:844,897 \| :846,899` | `edge_deltas = 9`, `message EdgeDeltas` | C-core | 5 | reserved, deleted · X8 | successor's mass less its ancestor's (`view.go:431-444` reads `Mint.supersedes`) |
| B, M | `T/record/refs.go:350 \| :375`; M `T/record/replay.go:170` | comments | C-text | 1 | rewritten · X8 | name the repair-regression denominator |
| B, M | `T/scorecard/scorecard.go:184,200,204` | `repair_regression_ratio` rows | C-regression | 3 | deleted · X8 | the scorecard's "Durable repairs" rows |
| B, M | `T/view/view.go:961 \| :994`; M `:1018` | `RepairRegression`, `LineageMints` | C-regression | 2 | deleted · X8 | telemetry JSON |
| B, M | `T/view/view.go:962 \| :995,1060-1061` | `EdgeDeltas` | C-core | 3 | deleted · X8 | as the proto row |
| B, M | `F/tests/simulator/debate.test.mjs:912 \| :871` | scorecard fixture | test | 1 | rewritten · R6 | `repair_regression_ratio` key |


