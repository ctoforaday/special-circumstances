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
- `F/tools/internal/record/passblockers.go` (6) — unchanged — an accepted or rejected ruling is a ruling (`rulingWord`); `requireBenchMotionsRuled`'s refusal (`:933`) names D's pair for a docketing (§III.B4, M17 amendment 4)
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
| `record.go:1750` `rulingWord` (→ `Motion.Ruled()`, `motionsAt`, Compute) | `Word(disposition)` | returns `accepted` / `rejected` — **the one switch**; without it an accepted or rejected M reads unruled everywhere |
| `grademovement.go:90` (p3) | restarts the movement window from `m.ruled` | **re-keyed (amendment 3, L41, audit p7 gap 4):** restarts only at a `GRADE_MOVEMENT` docketing's ruling on G, for either side, (and, per axis, after the latest bench `Regrade`); no other ruling and no assignee's accept restarts it |
| `recordsql/views.go:601` `motion_answers` | `COALESCE(grade, petition, rd.disposition)` | `rd.disposition` → `CASE rd.accepted WHEN 1 THEN 'accepted' WHEN 0 THEN 'rejected' ELSE rd.disposition END` (view; epoch) |
| `motion.go:544-574` `requireUnanswered`/`RequireUnruledMotion` | first ruling wins, in the write transaction | unchanged: the minter's accept is a ruling, so a second answer of either kind is refused here (gap 5) |
| `recordsql/gapstate.go:189-193` "bench" CTE; `:198-219` closure, unruled filing | joins `enum_disposition` where `closes` | **gains one arm**: an `accepted` ruling on a motion with a `motion_supersede` row, written by a seat other than G's stored `assignee` (the bench) — the reader fix (ruled 2026-11-13); NULL disposition otherwise joins nothing; any ruling clears `unruled_docket_*` |
| `bench-closed-readers.txt`: `closed_by_bench`, `bench_close_event`, `bench_closed_last` in views.go (8), viewjson.go (26), replay.go (6), gapstate.go (6), consistency.go (6), scorecard.go, remand.go, record.go, gapstates.go, declarations.go, answersretired.go; 12 agent OUTPUT schemas | read the closing bench ruling | unchanged: a bench close accept IS a closing ruling; a minter close accept is a `Close` (red's close, as it is) |
| `supersession.go:36` | `Closes(disposition)` | unchanged: `Closes(UNSPECIFIED)` is false (`facets.go:229`) |
| `replay.go:139-148,244-259` `BenchClosure`, `reason()` | closing rulings | also the bench's accepted supersede ruling (role `bench`), fate `superseded` through `benchClosureWord` |
| `viewjson.go:568,642-645` `lastBenchClosure`; the estopped row's `fate` | closing rulings | read the gap view's `bench_close_event` (which now names the accepted supersede ruling); word through `benchClosureWord` |
| `remand.go:206-235` `requireRemandWithinLimit`; `sitting.go:526`; `cli/motion/verbs.go:293`; `passblockers.go:933` `requireBenchMotionsRuled` (p3) | name the closing dispositions, or `--as <verdict>` | on a docketing holding a proposal name accept \| reject; on one holding none, the closing words, the bench's regrade and remand (plan §III.B13) |
| `remand.go:259` `ClosingDispositions()`; `enumvalue.go:225-245` `evsOf` → the generated `--as` lists (p3) | every `closes` word | unchanged: no disposition is added |
| `benchsitting.go:26-108` `benchHeld`/`docketHeld` (p3) | the bench sitting's blocking motions | also the write's admission (`requireBenchHeld`); `docketHeld` holds the standing docketings, and a `PROPOSAL` docketing's own motion — no proposal is filed on a docketed gap (plan §III.B13, L38; amends B2's "before the dispatch that convened the sitting" and amendment 1's "before the bench sitting's opening") |
| `capture.go:1453-1463` precedent harvest | every docket ruling | the bench's rulings only; word from `rulingWord` |
| `consistency.go:222-250` | non-closing word → `carriedCount++` | an accepted ruling is skipped, except the bench's accepted supersede ruling, which its own fold reads as a bench closure (role `bench`) and compares with the view |
| `report/assemble.go:1402-1414` LEAD; `view/view.go:731-740`; `viewjson.go:2108-2116` `Lead` | every docket ruling as the bench's opinion | the bench's rulings only (event seat's role `bench`); word from `rulingWord` |
| `record.go:1366-1420` DocketRuling write checks; `record.proto:1981-2010` CHECKs | disposition required | (amendment 4) `("sided_with" IS NULL) = (COALESCE("disposition",'') = 'remanded')`; `reopens_on`, `final` only with a closing word, `direction` only with `remanded`; `sided_with` in D's pair, and the assignee's accept the filer's side; `requireBenchOpinion`; the side against a proposal from the bench only; every bench ruling names its docketing (`DocketRuling.docketing_id`), and `--accept`/`--reject` on a docketing holding no proposal is refused by `requireSubjectMatches`' docket arm, naming the closing words, the bench's regrade and remand (plan §III.B13; replaces B2's no-act refusal) |
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
- `F/tools/internal/record/grademovement.go` (3) DOCKETING-STANDS — `restoring` (`:145-160`) kept, the bench's for-blue answer at grade movement (amendments 2–4, M17, L35); the window restart (`:80-90`) at the latest `GRADE_MOVEMENT` docketing's ruling, for either side, only, per axis after the latest bench `Regrade` (L41, audit p7 gap 4); `pending` reads the standing docketing
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
- **every `case route` switch** (`git grep -n 'case route' a488cad41 -- plugins/frank-exchange-of-views/tools`: `available.go:291-301`, `dispatch.go:458-517`, `remand.go:329,385-389`) DOCKETING-STANDS — each gains a `routeProposal` arm (its assignee's window; the window's end writes `Docketing{PROPOSAL}`); no other route is added
- `F/tools/internal/seatprobe/boards.go`, `build.go` (`docketing`) S — the bench board convened on a docketing; the dispatch board writes no motion
- `F/tools/internal/record/migrate/*` S — Translation 3 (plan §III.B13; O2, O3 ruled 2026-11-20) and Translation 4 (§III.B14, M17); `entries.go`, `replay.go`'s `docketing` mentions (TEXT); `remap.go` `idFields` += `Docketing.docketing_id`, `.gap_id`, `.motion_id`, `DocketRuling.docketing_id`, `Tombstone.docketing_id`
- test-side per-event-type lists (M15 command (5)) S — `releasegate/fuzz/fuzz_test.go` `verbsWithEvents` and `reportExemptions` (`docketing`: written by the dispatch, carrying the tool's `basis`), `cli/vocabulary_test.go` `noVerbOfItsOwn`, `recordpb/correspondence_test.go`
- tests S — `record/remand_test.go` (24), `m37replay_test.go` (22), `testdata/plans.golden` (7, every changed line read), `proposalanswers_test.go` (6: `TestTheGradeMovementFilingRestoresTheNetMovement` DELETE; `TestRestoringIsTheFirstFromAgainstTheLastTo` kept, re-pointed at the reject, M17), `grademovement_test.go` (6), `passblockers_test.go` (5), `cli/stategraph_test.go` (5), `lensturn_test.go` (4), `dispatchnext_test.go` (4), `proposalhelpers_test.go` (2), `m38replay_test.go` (2), `docketruling_test.go` (2), `blockerowner_test.go` (2), `cli/epochlimitexit_test.go` (2), `withheldpass_test.go`, `planfields_test.go`, `narrowedview_test.go`, `motionview_test.go`, `gavel_test.go`, `cli/petitionreadiness_test.go`, `cli/dispatch_test.go` (1 each); `releasegate/fuzz/remandroute_test.go` (6), `fuzz_test.go` (6), `oneproposal_test.go` (3), `benchorders_test.go` (3), `converseoracle_test.go` (2), `benchlift_test.go` (2), `proposal_test.go`, `gaproads_test.go` (1 each); `integration/surface/promptverbs_test.go`, `constitutiondirective_test.go`, `devcmd/enginegen/vocabulary_test.go` (1 each, `restoring`/`routeBench` in prompt text: re-read); [NEW] T-B22–T-B26
- O — `fetchcache/fetchcache.go`, `extractor.go`, `testdata/walls/pmlr.html` (3), `cli/ocr.go`, `cli/lens/verify.go`, `vecembed/vocab.txt`, the five `tessocr/testdata/corpus/gpo-record-threecolumn/reading.*.golden`: `restoring` in another sense

## M17 the bench's own closing word and own regrade, and every carrier of the docketing's answers — Part B (amendments 2 and 4, 2026-11-20, plan §III.B14)

**Amendment 4 (each side states its position; side-named answers; a stored side).** Every CHOOSE line
below that names `--accept | --reject` names D's pair instead — `--for-red | --for-blue`, or
`--for-filer | --for-assignee` on a red seat's proposal on another red seat's gap — and teaches each
side's statement (red's text in its closing, blue's word in its closing, a proposal, the moved or the
recorded grades) and that a side stating nothing cannot win. Every reader of `DocketRuling.accepted`
reads `sided_with` (tag **SIDED**, replacing REJECT-MOVES): "the held act applied" is `sided_with IN
(BLUE, FILER)`. Sweep re-run at the build: `git grep -n -E 'GetAccepted\(\)|"accepted"|"rejected"|--accept\b|--reject\b|flags\.(Accept|Reject)' a488cad41 -- plugins/frank-exchange-of-views/tools plugins/frank-exchange-of-views/skills scripts/agentgen ':!*_test.go'`,
each hit tagged SIDED, CHOOSE (the docket rule's flags), or O (`blue edit --accept`, `motion docket
accept`, another verb's flag).

Command, on `a488cad41` (`~/.claude/scratch/1695/amend2/m17-files.txt`, 49 files, 143 hits):

```
git grep -c -i -E 'regrade-bench|bench.?s (own )?regrade|BenchRegrade|requireRegradeWithinBenchLock|benchSetAxesOf|docket rule[^"]*--as|--as remanded|flags\.As|own closing word|closing word of its own|accept_without_proposal_refused|no proposal stands|ruleDocketAtExit|ruleStandingDockets' a488cad41 -- plugins/frank-exchange-of-views scripts/agentgen ':!*/vecembed/vocab.txt'
```

plus, read by hand, the agentgen sources the pattern does not reach (`lead-judge.md:42-43,82,112`,
`red-chair.md`), debate.js `:491,927,1123-1128` and the assembly prompt, `terms.json`'s `disposition`
(`:815-857`). Tags: **CHOOSE** (a surface where the bench chooses a word or grade: deleted, or reworded to
D's pair (amendment 4: `--for-red | --for-blue`, `--for-filer | --for-assignee`) | remand and the positions), **TOOL-WRITES** (record code that now writes the word or
grade from the record), **DELETE**, **TEXT**, **O** (another sense — `flags.As` of another verb).

- `F/tools/internal/record/recordpb/record.proto` TOOL-WRITES — `DocketRuling.disposition` loses `flag: "as"`, its `why` and the side-or-remand CHECK's `why` say the tool sets it (red's `repaired`, blue's stated word or `not_a_defect`, a proposal's stored word, `remanded`); `accepted = 9` reserved and `sided_with = 11` (`SidedWith`: `RED`, `BLUE`, `FILER`, `ASSIGNEE`; means "the position whose statement this ruling applied, as the answer named it") with its CHECKs (amendment 4); `Closing.closure_class = 6` (blue's word, `subset: "closes"`, `flag: "as"`; M19); `Disposition`'s "the bench's word" comment; `remanded`'s means unchanged; `Closing.location`, `fix_new`, `retire` (M19)
- `F/tools/internal/record/motion.go:540` CHOOSE — the docket subject's `--as` values (`MotionVerdicts`) → none; `requireSubjectMatches`' docket arm loses "no proposal stands"; `requireUnanswered` unchanged
- `F/tools/internal/record/record.go` (3) TOOL-WRITES — validate's `DocketRuling` arm: `sided_with` in D's pair; each answer's word is its row's (`repaired` after the bench's edit of red's text, blue's closing's word or `not_a_defect`, the stored word); for-red at impasse refused with no red text; a for-red answer's bench `BlueEdit`/`Retire` is the re-plan of red's closing; the bench `Regrade` in an accepted regrade proposal's or a reject at grade movement's transaction carries the stored or restored grades; no remand on a grade-movement docketing; no other bench `Regrade`; validate's `BlueEdit`/`Retire` seat arm admits the bench only in a for-red ruling's transaction (M20)
- `F/tools/internal/record/benchgrade.go` (8) TOOL-WRITES + TEXT — `requireRegradeWithinBenchLock` and `benchSetAxesOf` stay, for the two tool-written bench regrades; the file comment's "the bench files an ordinary `regrade`" (2026-10-30) → the tool writes it in the ruling's transaction
- `F/tools/internal/record/grademovement.go` (1) TOOL-WRITES — `restoring` kept (§III.B13 deleted it), now the for-blue answer's at grade movement (picked 2026-11-20; amendment 4's side name); `GradeMovementRule`'s "that ruling rules on the gap's movement" → keeps or restores it; the window (L41, M16)
- `F/tools/internal/record/impasse.go` (1) TOOL-WRITES + TEXT — `:308`'s comment ("The bench's regrade sets the grade its ruling rules on") → the tool's; `:314` reads a `BlueEdit` as blue's act in the exchange — the bench's for-red edit is not (M20). Amendment 2's upheld round and `GapExchanges.upheld` are withdrawn (amendment 3)
- `F/tools/internal/record/remand.go` S — `requireRemandWithinLimit`'s refusal text ("takes a disposition that closes it", `verbs.go:237` too) → a side of D's pair; a remand on a `GRADE_MOVEMENT` docketing refused (L40); amendment 2's `upheld` gate on `routeOf`'s impasse arms withdrawn
- `F/tools/internal/record/derivedword.go` — not built: amendment 3 withdraws `derivedClosingWord` (L34, the side's word)
- `F/tools/internal/record/supersede.go:303`, `enums.go` (6), `cli/chair/verdict.go` (5), `cli/bench/outcome.go` (3), `cli/blue/inquiry.go` (3), `cli/lens/reproduce.go` (3), `cli/lens/verify.go` (3), `cli/seat/closure.go` (3), `cli/motion/propose.go:68-76` O — `flags.As` of another verb (the proposal's own closure class, `bench outcome --as`, the verdict, an inquiry's status, a reproduction's or verification's outcome)
- `F/tools/internal/record/viewjson.go` (1), `sitting.go` S — the bench's work-list row on a docketing names the change at issue and G as it stands, and what each answer writes
- `F/tools/internal/cli/bench/command.go:23` DELETE — `seat.Regrade("regrade-bench")`; `F/tools/internal/cli/seat/help/regrade-bench.md` DELETE; `cli/seat/regrade.go:20-21` TEXT (one act, the lens's; the bench's regrade is the tool's)
- `F/tools/internal/cli/motion/verbs.go` (5) CHOOSE — `motion docket rule`: `--for-red | --for-blue | --for-filer | --for-assignee | --remand` exactly one, from D's pair or the remand (amendment 4), `--direction` with `--remand`, `RequiredUnless(c, flags.As, …)` (`:408`) gone; help `:226-240,293,405` teaches the positions table; `rulingSaid` (`:492`) unchanged
- `F/tools/internal/cli/motion/propose.go:218` TEXT — "The bench's regrade is the grade its ruling rules on" → the stored regrade, written as the bench's
- `F/tools/internal/cli/seat/help/regrade.md:11` TEXT — "AN AXIS THE BENCH SET" stays (the lock), set now by an accepted proposal or a restore; `close.md`, `mint.md`, `reproduce.md`, `verify.md` (each 1, "the bench's word") TEXT — a bench closure's word is a proposal's or the tool's; `help_gen.go` D
- `scripts/agentgen/src/lead-judge.md:42-43,82,112` CHOOSE — "REASONS, NOT JUST FATES" keeps its four fields; "WHERE YOU HOLD A GAP'S GRADE WRONG, YOU MUST SET IT BEFORE YOU RULE" deleted; the impasse paragraph and `:112`'s "you have the disposition for it" → a side of D's pair (`--for-red | --for-blue`; `--for-filer | --for-assignee`) or remand, each side's statement (red's text, blue's word, a proposal, the grades), that the tool applies the chosen side's and that a side stating nothing cannot win (amendment 4); `src/red-chair.md`, `src/fragments/lens-duties.md`, `lens-audit-and-mint.md` TEXT (grade movement: the bench keeps or restores; amendment 4: blue's closing may state its word, red's carries its text); `F/agents/*.md` (9) D
- `F/tools/internal/record/passblockers.go:933` TEXT — `requireBenchMotionsRuled`'s "`--as <verdict>` … --accept or --reject in place of --as" → D's pair (`--for-red | --for-blue`, or `--for-filer | --for-assignee`) on every docketing
- `F/docs/seat-command-triggers.md:116` (the `bench regrade` row) DELETE — audit p7 NOTE; promptverbs_test's trigger-doc check fails loudly once the verb is gone
- `F/tools/internal/cli/seat/verbs.go:291` (`show board`'s "the bench regrades a gap it is hearing") TEXT → a ruled grade (an accepted regrade proposal's, or a restore); inlined into the 12 agent definitions (D)
- **SIDED** (amendment 4, replacing amendment 3's REJECT-MOVES: each reader of `accepted` reads `sided_with`; a for-blue answer at grade movement writes the restore; either side at impasse, and for red on blue's close or supersede with red's text, closes G) — `F/tools/internal/record/inquirycheck.go:139-141` (`moved` counts accepts and not rejects → by the row's writes), T-B8's reject row (G byte-identical → per row), `proposalanswers_test.go:455` (`TestRestoringIsTheFirstFromAgainstTheLastTo`, re-pointed at the reject), `tombstonesOwed` (places by shape: no change), `consistency.go:222-250` (`carriedCount` reads closing words: no change). Sweep: `git grep -n -E '"rejected"|Ruling *!?= *"rejected"|GetAccepted\(\)' a488cad41 -- plugins/frank-exchange-of-views/tools ':!*_test.go'`, each hit read for "a reject moved nothing" (12 hits on `a488cad41`): `inquirycheck.go:141` and `retirement.go:249` (gap news words a grade-movement reject "rejected" → names the restore by `sided_with: BLUE`) SIDED; `record.go:1699-1702` and `cli/motion/verbs.go:495-498` (`rulingWord`, `rulingSaid`: the word itself → the side's words) SIDED; `proposal.go:243,257,392` (the assignee's accept) and `consistency.go:246` (M18) another reader; `record.pb.go:5202,5838` D
- `F/skills/research-protocol/scripts/debate.js` `:491` (a docket ruling's "fate-changing --as"), `:907` ("only you and the bench hearing it regrade it" → a ruled grade), `:927` ("its fate"), `:1123-1128` (terminal), the assembly prompt CHOOSE — a side on every docketing (red or blue; filer or assignee), the tool applying its statement, nothing remanded at the exit; `tests/simulator/testdata/prompt-red-chair.golden`, `prompt-judge-terminal.golden` D
- `F/tools/internal/terms/terms.json` (`disposition` `:815-857`: the closing word, on a bench closure the statement of the side the bench sided with; `grade movement` `:874`; `side` `:252` unchanged, red or blue) TEXT; `definitions_gen.go`, `F/docs/vocabulary.md` D
- `F/tools/internal/seatprobe/boards.go:643` TEXT — "only the bench's regrade binds the lens" → a ruled grade (an accepted proposal's or a restore's)
- `F/tools/internal/difftest/testdata/docket_ruling_requires_each_unconditional_field.golden` (16), `manual_runs_every_help_page_live.golden` (2), `projections_debate_changelog_citations.golden` (1) D — regenerated, every changed line read
- tests S — `cli/cli_test.go` (7), `crossseat_test.go` (2), `answers_test.go` (2), `benchorders_test.go`, `lens/axis_usage_test.go`, `record/benchorders_test.go` (4), `docketruling_test.go`, `proposalanswers_test.go` (`TestRestoringIsTheFirstFromAgainstTheLastTo` kept), `regraderesult_test.go`, `integration/surface/promptverbs_test.go` (3): bench regrades through the verb → through a for-blue answer to blue's regrade or a for-blue answer at grade movement; `--as <word>` and `--accept | --reject` on a docket rule → D's pair or `--remand`; `GetAccepted()` assertions → `GetSidedWith()`
- fuzz S — `fuzz_test.go` (13: `ruleDocketAtExit`, `ruleStandingDockets` draw from D's pair, blue's closing drive states a word on half its impasse closings; the bench regrade drive deleted), `benchlift_test.go` (6, the lock's lift by a later bench regrade → by a later accepted proposal or restore), `benchorders_test.go` (7), `sweeptally_test.go` (6: `accept_without_proposal_refused` DELETE; §III.B14's gates and oracles, amendment 4's `blue_word_applied` and `side_stated_nothing` included)
- `F/tools/internal/record/migrate/*` S — Translation 4 (plan §III.B14): each of the 19 archived dispatch closing words → a for-blue answer at impasse, `{sided_with: BLUE, disposition: not_a_defect}` (F-A (a), picked 2026-11-20); archived remands → `remanded`, no `sided_with`; a live-rule cause other than the basis's fails T-B31; the one archived bench `Regrade` dropped; archived `Closing` rows unchanged

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
