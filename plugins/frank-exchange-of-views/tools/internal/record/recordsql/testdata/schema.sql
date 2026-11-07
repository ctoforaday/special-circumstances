
-- THE EPOCH IS A FIELD ON THE RECORD, not a shape a reader infers from it.
--
-- A reader has to know whether the binary in its hands and the record in front of it agree about
-- what an event IS. That was answered twice and neither answer was the question: setup compares
-- the binary's epoch against the plugin manifest ONCE, at the start of a run, which cannot see a
-- seat invoking a different binary later; and the open path inferred it from SCHEMA SHAPE, asking
-- whether the record carried a table or view the binary did not declare. Shape cannot tell a stale
-- binary from a view the current binary deliberately DROPPED, so it made removing a view a two-step
-- change and reported the wrong party as stale.
--
-- One row, written in the transaction that creates the schema, so a database cannot exist without
-- its epoch. Compared at every open: lower than this binary's is an older run and migrates; higher
-- means THIS binary is the stale one. Equal proceeds.
CREATE TABLE "schema_epoch" (
  -- The CHECK is the whole point of the column: one record has ONE epoch, and a second row would
  -- make "the record's epoch" a question with two answers.
  "id"    INTEGER PRIMARY KEY CHECK ("id" = 1),
  "epoch" INTEGER NOT NULL
) STRICT;

CREATE TABLE "events" (
  "id"      INTEGER PRIMARY KEY,
  "seat_id" TEXT    NOT NULL,
  "ts"      TEXT    NOT NULL,
  "type"    TEXT    NOT NULL REFERENCES "enum_event_type"("value"),
  -- The key is the fact that has to be unique, and the partial index below enforces it globally.
  -- This was UNIQUE (seat_id, nonce, seq): a counter nothing read, scoped by a sitting that no
  -- longer exists. Both are gone.
  "key"     TEXT,
  -- THE SITTING IS A FIELD, NOT A NUMBER EACH READER RECOUNTS (#1151). It is the id of the event
  -- that OPENED the sitting this act belongs to; an opening event's own id, for itself. NULL is
  -- the honest answer for a row with no sitting open for its seat — the harness's own bookkeeping,
  -- and anything a seat writes before it has opened a sitting.
  --
  -- It is stamped by recordpb.SeatOpeningSitting at the write, inside the inserting transaction,
  -- and no reader re-derives it. Five readers used to recount openings for themselves and two of
  -- the spellings had already drifted: a window counting openings counted EVERY register while the
  -- idempotency key's count excluded a repair, so one seat's repair sitting was 2 to the view and
  -- 1 to the key. A stored id cannot disagree with itself.
  --
  -- It is not a value a SEAT can stamp wrong, which is the objection the two windows were built to
  -- answer: nothing on the wire carries it, the foreign key requires the event it names to exist,
  -- and the only thing a seat can do to move it is open a sitting.
  "sitting_id" INTEGER REFERENCES "events" ("id"),
  -- THE TWO WINDOWS (plans/roundless.md §III.A.0), STAMPED BESIDE sitting_id IN THE SAME WRITE.
  --
  -- "sitting" is the rank of the row's sitting among its owner's sittings — which sitting of that
  -- seat this act belongs to, its opening included; 0 for a row in no sitting. "epoch" is the count
  -- of red-chair's sittings opened at or before this row, whoever wrote it; 0 before the chair's
  -- first, which is the base phase (frontier, lanes, synthesis).
  --
  -- STORED BECAUSE BOTH ARE FIXED THE MOMENT THE ROW IS WRITTEN. Each depends only on rows with a
  -- lower id, and the record is append-only, so no later write can move either; a view that
  -- re-ranked every sitting on every read was paying, on every reader, for a number the write
  -- already knew. The write path (recordsql.sittingOf) is the one statement of both rules.
  --
  -- A ROW IN A SITTING HAS A RANK AND A ROW IN NONE HAS 0, and the CHECK refuses the two
  -- disagreeing.
  "sitting"    INTEGER NOT NULL CHECK ("sitting" >= 0),
  "epoch"      INTEGER NOT NULL CHECK ("epoch" >= 0),
  CHECK (("sitting_id" IS NULL) = ("sitting" = 0)),
  -- A REGISTER WITHOUT A SITTING IS UNREPRESENTABLE, and that is what retires the degradation rule
  -- this column replaces. A register opens a sitting (its own id), joins the one its agent's
  -- bracket opened (the bracket's id), or repairs one (the sitting it repairs), so NULL here means
  -- the row went in without the write path deciding — a forged event, which the readers used to
  -- have to agree about in six places and disagreed about in three. It is refused instead: the
  -- forger holds the file, but it does not hold the trigger.
  CHECK ("type" <> 'register' OR "sitting_id" IS NOT NULL)
) STRICT;

CREATE UNIQUE INDEX "events_key" ON "events" ("key") WHERE "key" IS NOT NULL;
CREATE INDEX "events_type" ON "events" ("type");
CREATE INDEX "events_sitting" ON "events" ("sitting_id");
-- The rows that opened a sitting, by the seat on their envelope: one half of a seat's latest
-- sitting (recordsql.LatestSitting). Partial, so it holds the openings and none of the acts.
CREATE INDEX "events_opening" ON "events" ("seat_id") WHERE "sitting_id" = "id";
-- A seat's own acts of one type, where the type's body table holds nothing to narrow by: the
-- report parts a seat filed (ReportAtPart, partStandingAt) and the verifies a seat carried
-- (ownAnswersAt), and a seat's mints about one kind (gapsMintedAbout).
CREATE INDEX "events_seat" ON "events" ("seat_id");

CREATE TRIGGER "events_are_append_only_update" BEFORE UPDATE ON "events" BEGIN
  SELECT RAISE(ABORT, 'the record is append-only: an event cannot be edited after it is written');
END;
CREATE TRIGGER "events_are_append_only_delete" BEFORE DELETE ON "events" BEGIN
  SELECT RAISE(ABORT, 'the record is append-only: an event cannot be removed after it is written');
END;

CREATE TABLE "seat_turn" (
  "agent_id"       TEXT    NOT NULL,
  "turn_idx"       INTEGER NOT NULL,
  "ts_ms"          INTEGER NOT NULL,
  "model"          TEXT    NOT NULL,
  "input_tokens"   INTEGER NOT NULL,
  "output_tokens"  INTEGER NOT NULL,
  "cache_read"     INTEGER NOT NULL,
  "cache_creation" INTEGER NOT NULL,
  -- A turn that carried a thinking block, and a turn that carried a tool_use block. Both are
  -- read off the content blocks rather than inferred from token counts: the run-4 forensics
  -- inferred thinking from a low output_tokens with a long span, which is a correlation, and
  -- the block is the fact.
  "is_thinking"    INTEGER NOT NULL,
  "is_tool"        INTEGER NOT NULL,
  PRIMARY KEY ("agent_id", "turn_idx")
) STRICT;

CREATE INDEX "seat_turn_agent" ON "seat_turn" ("agent_id");

CREATE TABLE "report_checkpoint" (
  "event_id" INTEGER PRIMARY KEY REFERENCES "events" ("id"),
  "text"     TEXT,
  "failure"  TEXT,
  CHECK (("text" IS NULL) <> ("failure" IS NULL))
) STRICT;

CREATE TRIGGER "report_checkpoint_is_append_only_update" BEFORE UPDATE ON "report_checkpoint" BEGIN
  SELECT RAISE(ABORT, 'a report checkpoint is the text an event left: it cannot be edited after it is written');
END;
CREATE TRIGGER "report_checkpoint_is_append_only_delete" BEFORE DELETE ON "report_checkpoint" BEGIN
  SELECT RAISE(ABORT, 'a report checkpoint is the text an event left: it cannot be removed after it is written');
END;

CREATE TABLE "enum_event_type" (
  "value" TEXT PRIMARY KEY,
  "means" TEXT NOT NULL
) STRICT;
INSERT INTO "enum_event_type" ("value", "means") VALUES ('anchor', 'a marker placed in the report');
INSERT INTO "enum_event_type" ("value", "means") VALUES ('blue_edit', 'a change to the report, recorded as old and new so the edit itself is auditable');
INSERT INTO "enum_event_type" ("value", "means") VALUES ('cast', 'the run''s admissible seats, written once by setup before any seat registers — what register and the dispatch verb check a seat id against');
INSERT INTO "enum_event_type" ("value", "means") VALUES ('certify', 'a seat''s signed statement about its own work — what it asserts on the record');
INSERT INTO "enum_event_type" ("value", "means") VALUES ('cite', 'a source brought into the debate, with the hash and access date that make it re-checkable');
INSERT INTO "enum_event_type" ("value", "means") VALUES ('close', 'red closing a gap on a verified repair — red''s half of the closing vocabulary');
INSERT INTO "enum_event_type" ("value", "means") VALUES ('closing', 'a party''s argument to the bench on one docketed gap — its minting lens''s or blue-respond''s');
INSERT INTO "enum_event_type" ("value", "means") VALUES ('computation', 'a program blue ran twice on a bundle, with the bundle''s hash, its exit and its basis — the answer a computation check demands');
INSERT INTO "enum_event_type" ("value", "means") VALUES ('declare', 'the bench stating a holding that later sittings are expected to apply');
INSERT INTO "enum_event_type" ("value", "means") VALUES ('dispatch', 'the chair engaging one party — a seat and the gaps it is engaged on — pinned to the report head it audits; the parties of one chair sitting are one dispatch');
INSERT INTO "enum_event_type" ("value", "means") VALUES ('envelope_admitted', 'the envelope an agent returned, admitted: the record refused nothing on the seat''s work list and the harness accepted the envelope against its schema, so the call ending its sitting succeeded — a closed sitting of a seat whose work arrives as items, with none of these, failed');
INSERT INTO "enum_event_type" ("value", "means") VALUES ('finding', 'something red found, graded but not yet minted as a gap');
INSERT INTO "enum_event_type" ("value", "means") VALUES ('halt', 'the bench ending the run on a safety, ethics, consent or integrity boundary');
INSERT INTO "enum_event_type" ("value", "means") VALUES ('inquiry', 'an inquiry: its question, its hypothesis and each fate recorded on it');
INSERT INTO "enum_event_type" ("value", "means") VALUES ('inquiry_check', 'the owning lens''s judgement that an inquiry''s latest fate is sound as the report''s account of it, with the reason it holds — a settled inquiry holds the PASS until one stands');
INSERT INTO "enum_event_type" ("value", "means") VALUES ('log', 'an entry addressed to the operator who can retool the seat — a defect, a request or friction — never a clean sitting, which is derived from having sat and filed no log entry');
INSERT INTO "enum_event_type" ("value", "means") VALUES ('manifest_row', 'one row of the run''s manifest, tying a gap to what shipped for it');
INSERT INTO "enum_event_type" ("value", "means") VALUES ('mint', 'a gap put on the board — the act that creates the entity every other act refers to');
INSERT INTO "enum_event_type" ("value", "means") VALUES ('motion', 'a motion filed: a grade contested, a petition to the bench, or a gap docketed for the bench');
INSERT INTO "enum_event_type" ("value", "means") VALUES ('motion_rule', 'the bench''s ruling on a filed motion, and whom it binds');
INSERT INTO "enum_event_type" ("value", "means") VALUES ('outcome', 'the run''s terminal act: how it ended and whether the question was answered');
INSERT INTO "enum_event_type" ("value", "means") VALUES ('plan', 'the plan the chair''s dispatch printed, whole — what the chair relays and the workflow dispatches; one per decision, with parties or none');
INSERT INTO "enum_event_type" ("value", "means") VALUES ('position', 'the chair''s argument to the bench on what is about no one gap');
INSERT INTO "enum_event_type" ("value", "means") VALUES ('register', 'a seat took its seat — the first act of any seat, stamping the tool version it ran under');
INSERT INTO "enum_event_type" ("value", "means") VALUES ('regrade', 'a gap''s grade changed, with the basis for the change');
INSERT INTO "enum_event_type" ("value", "means") VALUES ('report_part', 'a lane''s part of the report, stored verbatim as the lane ingested it; the report is the parts, in event order, with every recorded edit, anchor and retire folded over them');
INSERT INTO "enum_event_type" ("value", "means") VALUES ('report_read', 'the tool serving a lens the whole report at a head — evidence the text was served to it, not that it was read; the latest in a sitting the lens sat through is its pin');
INSERT INTO "enum_event_type" ("value", "means") VALUES ('reproduce', 'red checking a recorded computation against its bundle''s re-run, and whether what it computes is sound');
INSERT INTO "enum_event_type" ("value", "means") VALUES ('retire', 'a claim retired from the report, with the reason and what supersedes it');
INSERT INTO "enum_event_type" ("value", "means") VALUES ('sitting_close', 'the harness''s agent returning — the other end of that span');
INSERT INTO "enum_event_type" ("value", "means") VALUES ('sitting_limit', 'a seat''s sitting stopped at the run''s per-sitting tool-call limit — the hook refuses every further call in it, and this records which seat, which sitting and the limit');
INSERT INTO "enum_event_type" ("value", "means") VALUES ('sitting_open', 'the harness dispatching an agent — one end of a sitting''s span, observed by a hook rather than claimed by a seat');
INSERT INTO "enum_event_type" ("value", "means") VALUES ('spot_check', 'red re-checking a sample of prior work, or stating that it checked none and why');
INSERT INTO "enum_event_type" ("value", "means") VALUES ('stale', 'the tool refusing a seat''s act on a gap that changed since the seat''s sitting opened, having shown the seat what changed — the seat''s next act on that gap in this sitting reads from here');
INSERT INTO "enum_event_type" ("value", "means") VALUES ('verdict', 'the chair''s verdict for its epoch: PASS or FAIL against the open board');
INSERT INTO "enum_event_type" ("value", "means") VALUES ('verify', 'a citation checked at the leaf: what the source did for the claim, and how sure the reader is');

CREATE TABLE "enum_occasion" (
  "value" TEXT PRIMARY KEY,
  "means" TEXT NOT NULL
) STRICT;
INSERT INTO "enum_occasion" ("value", "means") VALUES ('assemble', 'assembling the final report by union-copy. The last step of the run');
INSERT INTO "enum_occasion" ("value", "means") VALUES ('docket', 'ruling the docket: the gaps that reached impasse and were docketed for adjudication. The chair dispatches this sitting');
INSERT INTO "enum_occasion" ("value", "means") VALUES ('petition', 'hearing the petitions filed since the bench last sat to hear one. The engine convenes it as soon as nothing is running, when a close''s record_read reports a petition unheard: in the turn that filed it, at no epoch');
INSERT INTO "enum_occasion" ("value", "means") VALUES ('position', 'hearing the chair''s position at a sitting that withheld a PASS its plan permitted: upheld names the work the debate continues on, overruled sends the PASS back to the chair. The engine convenes it as soon as nothing is running, when a close''s record_read reports a position unheard: in the turn that filed it, at no epoch');
INSERT INTO "enum_occasion" ("value", "means") VALUES ('terminal', 'the terminal disposition at the exit boundary: what still stands, and every motion left unruled. Nothing can be remanded from here');

CREATE TABLE "enum_verdict" (
  "value" TEXT PRIMARY KEY,
  "means" TEXT NOT NULL
) STRICT;
INSERT INTO "enum_verdict" ("value", "means") VALUES ('fail', 'the board, as it stands, still stops you — refused over a converged board');
INSERT INTO "enum_verdict" ("value", "means") VALUES ('pass', 'nothing on the board holds the gate — `show work` marks what does, and a refused PASS names every one — and this is CHECKED against the board, not taken on your word');

CREATE TABLE "enum_run_outcome" (
  "value" TEXT PRIMARY KEY,
  "means" TEXT NOT NULL
) STRICT;
INSERT INTO "enum_run_outcome" ("value", "means") VALUES ('ceiling', 'the run stopped at a limit of its terms, with nobody it could still dispatch and PASS not permitted: its epoch limit, as the derivation''s reason states; NOT a judged failure to verify: what stays open on the board, and its re-audit, travels OUT of the run');
INSERT INTO "enum_run_outcome" ("value", "means") VALUES ('halted', 'the bench ended the run on a safety, ethics, consent or integrity boundary');
INSERT INTO "enum_run_outcome" ("value", "means") VALUES ('unverified', 'the run ended without the question being answered, and no ceiling or halt explains it');
INSERT INTO "enum_run_outcome" ("value", "means") VALUES ('verified', 'red passed the board and the bench agrees the question was answered');

CREATE TABLE "enum_class_material" (
  "value" TEXT PRIMARY KEY,
  "means" TEXT NOT NULL
) STRICT;
INSERT INTO "enum_class_material" ("value", "means") VALUES ('always', 'every gap of this class is material, whatever its grade or the reader decision it is judged to change — the class decides it, and no ruling or holding makes it immaterial');
INSERT INTO "enum_class_material" ("value", "means") VALUES ('by_grade', 'a gap of this class is material when its current severity is medium or above');
INSERT INTO "enum_class_material" ("value", "means") VALUES ('never', 'no gap of this class is material, whatever its grade — it stays on the board and never holds the gate');

CREATE TABLE "enum_gap_class" (
  "value" TEXT PRIMARY KEY,
  "means" TEXT NOT NULL,
  "material_default" TEXT NOT NULL REFERENCES "enum_class_material"("value")
) STRICT;
INSERT INTO "enum_gap_class" ("value", "means", "material_default") VALUES ('artifact_preservation', 'an artifact a reader needs to check the report — a computation, program, cached source or data file — is not kept, or does not run, where the report says it is', 'by_grade');
INSERT INTO "enum_gap_class" ("value", "means", "material_default") VALUES ('citation_figure_misattribution', 'a cited source does not say what the report attributes to it — the figure, statement or support is absent from its text, the cached text is a different document from the one named, or the anchor backs a sentence its source is not about', 'always');
INSERT INTO "enum_gap_class" ("value", "means", "material_default") VALUES ('claim_contradicts_own_record', 'the report and the run''s own record disagree: the report states what the record contradicts — an inquiry''s fate, method or reason, an edit, a computation''s result — or a record entry contradicts what the report shows was done', 'always');
INSERT INTO "enum_gap_class" ("value", "means", "material_default") VALUES ('claim_contradicts_source', 'a source read against the sentence — cited by blue or found by red — states otherwise than the report', 'always');
INSERT INTO "enum_gap_class" ("value", "means", "material_default") VALUES ('cross_section_contradiction', 'two passages of the report state things that cannot both be true — figures, verdicts, grades, or the reasons given for them', 'always');
INSERT INTO "enum_gap_class" ("value", "means", "material_default") VALUES ('derivation_status_overclaim', 'a claim carries more certainty or status than how it was established supports — inferred stated as verified, hedged in one place and flat in another, assumed stated as shown', 'always');
INSERT INTO "enum_gap_class" ("value", "means", "material_default") VALUES ('doctrine_vs_implementation', 'what a system, method or report says it does differs from what its implementation does — a stated rule its code, script or procedure does not follow', 'by_grade');
INSERT INTO "enum_gap_class" ("value", "means", "material_default") VALUES ('exhaustive_sweep_omits_case', 'a list, survey, search or sweep the report presents as complete omits a member it should hold', 'always');
INSERT INTO "enum_gap_class" ("value", "means", "material_default") VALUES ('false_universal', 'a claim stated universally — every, none, always, the only — that a case the evidence reaches, or one it never tested, contradicts', 'always');
INSERT INTO "enum_gap_class" ("value", "means", "material_default") VALUES ('figure_recount_fails', 'a figure, count or computed result the report states differs from what an independent recount, recomputation or re-run produces', 'always');
INSERT INTO "enum_gap_class" ("value", "means", "material_default") VALUES ('incomplete_repair_propagation', 'a repair fixed a defect at some of its sites and left it standing at others — a sibling sentence, row or section still carries what the repair removed', 'always');
INSERT INTO "enum_gap_class" ("value", "means", "material_default") VALUES ('invalid_inference', 'a reason the report offers for a conclusion does not entail it — the step fails even where every stated fact is true, including a check that gives the same result on the case the conclusion must exclude', 'by_grade');
INSERT INTO "enum_gap_class" ("value", "means", "material_default") VALUES ('live_source_drift', 'a claim or computation rests on a source''s or tree''s current state rather than the state the report cites or pins, so it reads differently now from when it was written', 'always');
INSERT INTO "enum_gap_class" ("value", "means", "material_default") VALUES ('metric_conflation', 'two quantities or categories that measure or mean different things are treated as one — compared, added or given one label — across different bases, methods or definitions', 'always');
INSERT INTO "enum_gap_class" ("value", "means", "material_default") VALUES ('missing_root_invariant', 'one fact is restated in several places with nothing tying the copies to a single source, so they have drifted or can drift unnoticed', 'by_grade');
INSERT INTO "enum_gap_class" ("value", "means", "material_default") VALUES ('out_of_scope', 'an inquiry asks a real question, but not the question this run was asked — its fate should set it aside, declined or deferred with why, rather than carry it as this run''s research', 'always');
INSERT INTO "enum_gap_class" ("value", "means", "material_default") VALUES ('partial_control_coverage', 'a control credited with covering a surface works on part of it and leaves the rest — an axis, channel, verb or case — uncovered', 'by_grade');
INSERT INTO "enum_gap_class" ("value", "means", "material_default") VALUES ('policy_without_mechanism', 'a rule, safeguard or guarantee is asserted with nothing that enforces it', 'by_grade');
INSERT INTO "enum_gap_class" ("value", "means", "material_default") VALUES ('reader_channel_mismatch', 'a fact the reader needs is held — in the record or elsewhere in the report — but not where the reader who needs it reads, so a reader acting on that section acts without it', 'by_grade');
INSERT INTO "enum_gap_class" ("value", "means", "material_default") VALUES ('report_voice', 'report prose speaks about the run that made it rather than the subject — process voice, a claim wearing the seat that produced it, the report''s own draft history, or the machinery that checked it', 'always');
INSERT INTO "enum_gap_class" ("value", "means", "material_default") VALUES ('risk_coverage_omission', 'the report''s account of risk leaves out a risk, or a combination of risks, that its own analysis or its subject makes live', 'by_grade');
INSERT INTO "enum_gap_class" ("value", "means", "material_default") VALUES ('self_attestation', 'the report asserts its own checking — that it verified, re-ran, searched or matched something — and the record does not bear the assertion out', 'always');
INSERT INTO "enum_gap_class" ("value", "means", "material_default") VALUES ('self_defeating_mitigation', 'a control credited with closing a gap cannot deliver that closure by its own operation — it fires after the decision it gates, acts on a variable that does not trigger the harm, erases the signal it was meant to distinguish, rests on an assumption the same document discredits, or is starved of its trigger by another control''s short-circuit', 'by_grade');
INSERT INTO "enum_gap_class" ("value", "means", "material_default") VALUES ('spec_underspecification', 'a term, rule, procedure or recommendation the report relies on is not specified enough for a reader to apply or check it — undefined, defined only by what it is not, or left as an undecided either/or', 'by_grade');
INSERT INTO "enum_gap_class" ("value", "means", "material_default") VALUES ('structure_noncompliance', 'the report''s structure does not carry its argument: material a reader has to take in order is present only as unordered prose, or a part of the answer has no place in it. The report''s sections are its author''s to name and order, and a heading the report does not have is not this', 'never');
INSERT INTO "enum_gap_class" ("value", "means", "material_default") VALUES ('unclassified', 'no other value fits this gap. The mint''s reason names the nearest value and what it does not cover', 'by_grade');
INSERT INTO "enum_gap_class" ("value", "means", "material_default") VALUES ('unhandled_degenerate_case', 'a procedure or design the report describes breaks or is undefined at an edge input — empty, zero, one, a boundary, a duplicate — that the report does not address', 'by_grade');
INSERT INTO "enum_gap_class" ("value", "means", "material_default") VALUES ('unsourced_claim', 'a claim of fact or a quotation the REPORT states with no citation or computation of its own behind it — no anchor, an anchor that resolves to nothing in the evidence, or an appeal to a result it never names. A source a lens read that backs the sentence is not the report''s citation: the gap stands until blue cites one', 'by_grade');
INSERT INTO "enum_gap_class" ("value", "means", "material_default") VALUES ('unverified_composition', 'a property of a whole is claimed because its parts were each checked, and the combination itself was never checked', 'always');
INSERT INTO "enum_gap_class" ("value", "means", "material_default") VALUES ('verification_scope_blindspot', 'a check, experiment or computation covers less than the claim it is credited with — a read test cited for a write path, one artifact''s check credited to its siblings, a proof that establishes half its sentence', 'always');
INSERT INTO "enum_gap_class" ("value", "means", "material_default") VALUES ('vote_laundering', 'agreement among sources, programs or reviewers is counted as independent corroboration when they share an origin, so the count overstates the evidence', 'by_grade');
INSERT INTO "enum_gap_class" ("value", "means", "material_default") VALUES ('within_source_condition_misattribution', 'the report carries a source''s or model''s result outside the conditions under which the source establishes it — a condition, bound, regime or date the source states is dropped', 'always');

CREATE TABLE "enum_grade_dimension" (
  "value" TEXT PRIMARY KEY,
  "means" TEXT NOT NULL
) STRICT;
INSERT INTO "enum_grade_dimension" ("value", "means") VALUES ('complexity', 'what fixing it costs');
INSERT INTO "enum_grade_dimension" ("value", "means") VALUES ('impact', 'how bad the consequence is if it lands');
INSERT INTO "enum_grade_dimension" ("value", "means") VALUES ('likelihood', 'how likely the CONSEQUENCE is — never how likely the defect is to BE there');
INSERT INTO "enum_grade_dimension" ("value", "means") VALUES ('severity', 'how bad the defect is in itself');

CREATE TABLE "enum_grade" (
  "value" TEXT PRIMARY KEY,
  "means" TEXT NOT NULL,
  "mass" REAL NOT NULL
) STRICT;
INSERT INTO "enum_grade" ("value", "means", "mass") VALUES ('certain', 'the top of the scale — for LIKELIHOOD, reserve it for a consequence that is itself certain, never for a defect you merely verified exists', 3.5);
INSERT INTO "enum_grade" ("value", "means", "mass") VALUES ('high', 'serious', 3);
INSERT INTO "enum_grade" ("value", "means", "mass") VALUES ('low', 'minor', 1);
INSERT INTO "enum_grade" ("value", "means", "mass") VALUES ('low_medium', 'below the by-grade material floor', 1.5);
INSERT INTO "enum_grade" ("value", "means", "mass") VALUES ('medium', 'the by-grade material floor', 2);
INSERT INTO "enum_grade" ("value", "means", "mass") VALUES ('medium_high', 'above the by-grade material floor, below serious', 2.5);
INSERT INTO "enum_grade" ("value", "means", "mass") VALUES ('realized', 'it has already happened. Contributes ZERO mass by design: mass forecasts what is still to come, and a realized defect is measured by its damage instead', 0);
INSERT INTO "enum_grade" ("value", "means", "mass") VALUES ('trivial', 'cosmetic; nothing downstream changes if it is wrong', 0.5);

CREATE TABLE "enum_petition_class" (
  "value" TEXT PRIMARY KEY,
  "means" TEXT NOT NULL
) STRICT;
INSERT INTO "enum_petition_class" ("value", "means") VALUES ('constitutional', 'the instruction itself conflicts with the rules the run is bound by');
INSERT INTO "enum_petition_class" ("value", "means") VALUES ('ethical', 'proceeding would require acting against the interests of someone the run affects');
INSERT INTO "enum_petition_class" ("value", "means") VALUES ('integrity', 'proceeding would require asserting what you believe false, or burying a real finding');
INSERT INTO "enum_petition_class" ("value", "means") VALUES ('safety', 'proceeding would create or conceal a hazard');

CREATE TABLE "enum_grade_ruling" (
  "value" TEXT PRIMARY KEY,
  "means" TEXT NOT NULL
) STRICT;
INSERT INTO "enum_grade_ruling" ("value", "means") VALUES ('accepted', 'the filer is right and the grade should move. The ruling moves nothing itself: the gap''s originating lens moves it with `regrade`, so say in --reason which grade and to what');
INSERT INTO "enum_grade_ruling" ("value", "means") VALUES ('rejected', 'the grade stands. Your --reason is what a seat that disagrees argues against before the bench, so it carries the argument, not the conclusion');

CREATE TABLE "enum_petition_ruling" (
  "value" TEXT PRIMARY KEY,
  "means" TEXT NOT NULL
) STRICT;
INSERT INTO "enum_petition_ruling" ("value", "means") VALUES ('denied', 'the objection does not hold, and your reason must say why at the leaf — a refusal without one is a decoration the petitioner cannot contest');
INSERT INTO "enum_petition_ruling" ("value", "means") VALUES ('granted', 'the objection holds. The relief BINDS the seats that come after, so state it as an instruction they can follow');

CREATE TABLE "enum_disposition" (
  "value" TEXT PRIMARY KEY,
  "means" TEXT NOT NULL,
  "closes" INTEGER NOT NULL CHECK ("closes" IN (0, 1))
) STRICT;
INSERT INTO "enum_disposition" ("value", "means", "closes") VALUES ('amends_prior', 'a defect found BETWEEN two repairs that each closed clean earlier — its lineage is the supersedes the gap was minted with; the close itself carries none and nothing checks it', 1);
INSERT INTO "enum_disposition" ("value", "means", "closes") VALUES ('defect_accepted', 'the fix costs more than the defect (complexity above likelihood x impact) and the risk is taken KNOWINGLY, with the argument on the record', 1);
INSERT INTO "enum_disposition" ("value", "means", "closes") VALUES ('defect_owed_elsewhere', 'a real defect whose fix is owned outside this debate; it leaves here and is not silently dropped', 1);
INSERT INTO "enum_disposition" ("value", "means", "closes") VALUES ('moot', 'the gap''s predicate expired: the claim or artifact it attached to is no longer in the report, so there is nothing left to repair or to argue about — neither not_a_defect nor repaired', 1);
INSERT INTO "enum_disposition" ("value", "means", "closes") VALUES ('not_a_defect', 'blue argued the finding was wrong and the argument held; nothing was repaired because nothing needed to be', 1);
INSERT INTO "enum_disposition" ("value", "means", "closes") VALUES ('remanded', 'NOT a closure: the gap goes back to the debate for ONE more exchange — the dispatch readies blue on it, owing the research direction the ruling states, and its minting lens once blue has acted — a remand closes nothing, so it states a direction and never --final. The exchange begins only when blue records an act on the gap. If it leaves the gap at impasse it is docketed again, to be closed — a second remand at impasse is refused — and if it moves the gap the gap''s limits count afresh from it', 0);
INSERT INTO "enum_disposition" ("value", "means", "closes") VALUES ('repaired', 'the repair was verified at the leaf and nothing regressed', 1);
INSERT INTO "enum_disposition" ("value", "means", "closes") VALUES ('repaired_with_regression', 'repaired, but something else broke — REQUIRES a successor naming the gap that carries the regression forward', 1);

CREATE TABLE "enum_position_ruling" (
  "value" TEXT PRIMARY KEY,
  "means" TEXT NOT NULL
) STRICT;
INSERT INTO "enum_position_ruling" ("value", "means") VALUES ('overruled', 'the objection does not hold: the PASS goes back to the chair, whose next sitting decides on the PASS with a verdict and may not argue the same withholding again. It names no work, and it is refused on a position whose class is always material: the class decides that materiality, so such a position is upheld');
INSERT INTO "enum_position_ruling" ("value", "means") VALUES ('upheld', 'the objection holds and the debate continues on the work you name: a lens of the cast, a retired one included, readied to audit its area for the defect the position argues, or open gaps, each sent back for one more exchange, begun by blue''s act on it and audited by its minting lens after — a remand, and the gap''s one remand at impasse. Name one or the other, with --direction stating the work');

CREATE TABLE "enum_ruling_binds" (
  "value" TEXT PRIMARY KEY,
  "means" TEXT NOT NULL
) STRICT;
INSERT INTO "enum_ruling_binds" ("value", "means") VALUES ('blue', 'the relief binds the response seat — what blue must do, or must not, in its coming sitting');
INSERT INTO "enum_ruling_binds" ("value", "means") VALUES ('both', 'it binds the whole exchange, and every dispatched seat carries it');
INSERT INTO "enum_ruling_binds" ("value", "means") VALUES ('red', 'it binds the audit seats: the lenses and the chair');

CREATE TABLE "enum_about_kind" (
  "value" TEXT PRIMARY KEY,
  "means" TEXT NOT NULL
) STRICT;
INSERT INTO "enum_about_kind" ("value", "means") VALUES ('anchor', 'a citation or computation, by anchor, whose check failed: a computation judged unsound, a citation whose work went unread');
INSERT INTO "enum_about_kind" ("value", "means") VALUES ('gap', 'a gap already on the board, by its id: a dispute with its grade, its fix, or whether it stands');
INSERT INTO "enum_about_kind" ("value", "means") VALUES ('inquiry', 'an inquiry, by its id: an objection to the inquiry as the record holds it and the report prints it — the REASON it was declined, deferred or abandoned, the METHOD its fate states, or a fate the report''s text or a computation contradicts. The steelman duty''s own anchor');
INSERT INTO "enum_about_kind" ("value", "means") VALUES ('section', 'a named report section, for something MISSING from it — the anchor a quote cannot provide, because the text you are objecting to is not there');

CREATE TABLE "enum_check_kind" (
  "value" TEXT PRIMARY KEY,
  "means" TEXT NOT NULL
) STRICT;
INSERT INTO "enum_check_kind" ("value", "means") VALUES ('computation', 'RUNNING something settles it. The demand on its close: a gap whose check kind is `source` or `computation` closes, under every disposition, only on a live answer (a live citation or computation of that kind placed as its answer), or once its lens mints a gap of the same check kind superseding it, which takes the demand on, or once a retire takes out the anchor it is about, or whose reading it carries, or once the seat whose check it carries checks that anchor again after the gap was minted and the check holds (under `source` a reading that backs the sentence, under `computation` a `sound` judgement); a demand its lens no longer holds goes before the bench, whose closing ruling is not held to it. Reach for it wherever the answer would be PRODUCED rather than asserted — arithmetic, a simulation, a forecast, a parse, a count, a re-derivation, among others: if a script could end the argument, this is the kind');
INSERT INTO "enum_check_kind" ("value", "means") VALUES ('document', 'reading a shipped artifact settles it — the check is answered by prose that quotes what is there');
INSERT INTO "enum_check_kind" ("value", "means") VALUES ('source', 'verifying an external source settles it: the claim stands or falls on what the cited material says. The demand on its close: a gap whose check kind is `source` or `computation` closes, under every disposition, only on a live answer (a live citation or computation of that kind placed as its answer), or once its lens mints a gap of the same check kind superseding it, which takes the demand on, or once a retire takes out the anchor it is about, or whose reading it carries, or once the seat whose check it carries checks that anchor again after the gap was minted and the check holds (under `source` a reading that backs the sentence, under `computation` a `sound` judgement); a demand its lens no longer holds goes before the bench, whose closing ruling is not held to it');

CREATE TABLE "enum_source_text_origin" (
  "value" TEXT PRIMARY KEY,
  "means" TEXT NOT NULL
) STRICT;
INSERT INTO "enum_source_text_origin" ("value", "means") VALUES ('embedded', 'the retrieved bytes carry the text: markup, plain text, or a PDF''s own text layer');
INSERT INTO "enum_source_text_origin" ("value", "means") VALUES ('none', 'the tool holds no text for the source: an image, or a scan nobody has read');
INSERT INTO "enum_source_text_origin" ("value", "means") VALUES ('not_recorded', 'this citation predates the field; written only by migrate');
INSERT INTO "enum_source_text_origin" ("value", "means") VALUES ('ocr', 'machine-read off page images by the local engine — deterministic, and it can misread, so a quote from it is checked against the page''s pixels');

CREATE TABLE "enum_work_status" (
  "value" TEXT PRIMARY KEY,
  "means" TEXT NOT NULL
) STRICT;
INSERT INTO "enum_work_status" ("value", "means") VALUES ('not_checked', 'no index was consulted about this work — it carries no doi, or the index did not answer. It says NOTHING about whether the work stands, and must not be read as reassurance');
INSERT INTO "enum_work_status" ("value", "means") VALUES ('not_recorded', 'this citation predates the field; written only by migrate');
INSERT INTO "enum_work_status" ("value", "means") VALUES ('retracted', 'an index reports the work RETRACTED. It may still be cited — as retracted, which the footnote and the Bibliography say for you — and its findings support no claim: a quotation from it is evidence of what the withdrawn paper said, and nothing more');
INSERT INTO "enum_work_status" ("value", "means") VALUES ('standing', 'an index was asked and reports no retraction: the work stands in the literature as published');

CREATE TABLE "enum_source_completeness" (
  "value" TEXT PRIMARY KEY,
  "means" TEXT NOT NULL
) STRICT;
INSERT INTO "enum_source_completeness" ("value", "means") VALUES ('abstract', 'the copy is the work''s ABSTRACT page: its platform printed a paywall or page images in place of the body. What it confirms is what the abstract says, not what the study shows, and the footnote and the Bibliography mark it ABSTRACT ONLY for the reader');
INSERT INTO "enum_source_completeness" ("value", "means") VALUES ('full', 'the copy carries the work''s body: its platform marks the body sections, or it is a pdf covering the page span the work declares');
INSERT INTO "enum_source_completeness" ("value", "means") VALUES ('not_asked', 'the question did not apply or could not be put: a page that is not about a scholarly work, a pdf with no declared span, or a source outside the run''s cache. It says NOTHING about whether the copy is complete');
INSERT INTO "enum_source_completeness" ("value", "means") VALUES ('not_recorded', 'this citation predates the field; written only by migrate');
INSERT INTO "enum_source_completeness" ("value", "means") VALUES ('not_the_work', 'the copy is not the work at all: a sign-in page, a book''s sales page. A reading of it is no reading of the cited work, whatever it found there: it leaves the citation unread');
INSERT INTO "enum_source_completeness" ("value", "means") VALUES ('unverified', 'nothing the tool knows could tell whether the copy carries the body — a platform with no known marker, a pdf much shorter than the work, a book''s pdf');

CREATE TABLE "enum_copy_substitution" (
  "value" TEXT PRIMARY KEY,
  "means" TEXT NOT NULL
) STRICT;
INSERT INTO "enum_copy_substitution" ("value", "means") VALUES ('followed_link', 'the typed url''s page named its own full text, and the tool followed that pointer: the cited bytes are the page at copy_url, the same work at the place its publisher named');
INSERT INTO "enum_copy_substitution" ("value", "means") VALUES ('fuller_copy', 'the typed url''s page was not known to carry the work''s body, and the tool took a copy carrying more of it from another source, at copy_url — another copy of the work, possibly another version');
INSERT INTO "enum_copy_substitution" ("value", "means") VALUES ('named_backend', 'a fetch that named a backend cached its answer under the typed url: the cited bytes came from that backend, at copy_url. With no copy_url the bytes are the tool''s own statement of what the backend found, not a document');
INSERT INTO "enum_copy_substitution" ("value", "means") VALUES ('none', 'the cited bytes are the typed url''s own: the url answered, and its answer is what was read');
INSERT INTO "enum_copy_substitution" ("value", "means") VALUES ('not_cached', 'a reading only: this run''s fetch cache held no copy of the url when red recorded it — red read the source outside the tool, or the cache that held it is gone — so the tool cannot say whose bytes were read. It says NOTHING about the copy, and a citation never carries it: cite fetches what it records');
INSERT INTO "enum_copy_substitution" ("value", "means") VALUES ('not_recorded', 'this citation or reading predates the field; written only by migrate');
INSERT INTO "enum_copy_substitution" ("value", "means") VALUES ('recovered', 'the typed url refused this container, and the tool''s recovery took the bytes from another source — an archive snapshot, an open-access copy, arXiv, or a bibliographic record — at copy_url. With no copy_url the bytes are the tool''s own statement of what the recovery found, not a document');

CREATE TABLE "enum_source_outcome" (
  "value" TEXT PRIMARY KEY,
  "means" TEXT NOT NULL
) STRICT;
INSERT INTO "enum_source_outcome" ("value", "means") VALUES ('absent', 'you read the source and the claim is simply not in it. Distinct from `refutes`: the source says nothing about the claim, where a refuting source says otherwise. A reading of the work with this outcome on a citation mints a gap on the sentence in the same act; on a source you found it records the reading and mints nothing');
INSERT INTO "enum_source_outcome" ("value", "means") VALUES ('refutes', 'you read the source and it CONTRADICTS the claim. A reading of the work with this outcome mints a gap on the sentence in the same act');
INSERT INTO "enum_source_outcome" ("value", "means") VALUES ('supports', 'you read the source and it says what the claim says. Where the run''s copy of it is only the work''s abstract, the record stamps that beside your verdict: it confirms what the abstract says, not what the study shows. On a source you found (no citation), a reading of the work with this outcome on a sentence the report cites nothing for mints an `unsourced_claim` gap in the same act');
INSERT INTO "enum_source_outcome" ("value", "means") VALUES ('supports_with_bridge', 'it supports the claim but you had to bridge something — a summary, a secondary citation, a near-restatement. On a source you found (no citation), a reading of the work with this outcome on a sentence the report cites nothing for mints an `unsourced_claim` gap in the same act');
INSERT INTO "enum_source_outcome" ("value", "means") VALUES ('unreachable', 'you could not read it — paywall, dead link, a format you could not extract. Say what you tried in --reason; an untried "unable to corroborate" is an incomplete audit. On a citation it reads nothing, so the citation stays unread: PASS is refused while the report cites it, and the first lens to read nothing of it owes a gap about it, outside its mint budget, asking blue to retire it or cite a source a lens can read');
INSERT INTO "enum_source_outcome" ("value", "means") VALUES ('weak', 'it gestures at the claim, or is itself uncorroborated: thin support, not none. On a source you found (no citation), a reading of the work with this outcome on a sentence the report cites nothing for mints an `unsourced_claim` gap in the same act');

CREATE TABLE "enum_confidence" (
  "value" TEXT PRIMARY KEY,
  "means" TEXT NOT NULL
) STRICT;
INSERT INTO "enum_confidence" ("value", "means") VALUES ('high', 'you read the source at the leaf and would defend this determination as it stands');
INSERT INTO "enum_confidence" ("value", "means") VALUES ('low', 'your reading may be wrong: an ambiguous passage, thin evidence, or a source you could only partly read. This is a call for more evidence, NOT an automatic fail — blue digs further');
INSERT INTO "enum_confidence" ("value", "means") VALUES ('medium', 'you are reasonably sure, but the reading bridges something — a summary, a secondary source, a near-restatement rather than the exact statement');

CREATE TABLE "enum_bridge_kind" (
  "value" TEXT PRIMARY KEY,
  "means" TEXT NOT NULL
) STRICT;
INSERT INTO "enum_bridge_kind" ("value", "means") VALUES ('near_restatement', 'the source says the same thing in other words: every term names the same object it names in the claim, and nothing the claim asserts is added, dropped or reversed');
INSERT INTO "enum_bridge_kind" ("value", "means") VALUES ('secondary_citation', 'the source reports the claim on another work''s authority — it cites the work that makes it — so the support is second-hand');
INSERT INTO "enum_bridge_kind" ("value", "means") VALUES ('summary', 'the source summarises what the claim states in detail, or the claim summarises the source: the same assertion, at a different length');

CREATE TABLE "enum_basis" (
  "value" TEXT PRIMARY KEY,
  "means" TEXT NOT NULL
) STRICT;
INSERT INTO "enum_basis" ("value", "means") VALUES ('reproducible', 'both runs wrote the same output, exit and out/ files, byte for byte');
INSERT INTO "enum_basis" ("value", "means") VALUES ('variant', 'the two runs differed, and blue stated what holds across them');

CREATE TABLE "enum_soundness" (
  "value" TEXT PRIMARY KEY,
  "means" TEXT NOT NULL
) STRICT;
INSERT INTO "enum_soundness" ("value", "means") VALUES ('sound', 'you READ the program and it computes what it claims to compute');
INSERT INTO "enum_soundness" ("value", "means") VALUES ('unsound', 'it re-runs cleanly and establishes nothing, or something other than the claim it is anchored to');

CREATE TABLE "enum_rerun" (
  "value" TEXT PRIMARY KEY,
  "means" TEXT NOT NULL
) STRICT;
INSERT INTO "enum_rerun" ("value", "means") VALUES ('calls_the_tool', 'the bundle called feov-record, which refuses every call made inside a re-run: the check cannot run in a checker''s hands, so nothing is reproduced');
INSERT INTO "enum_rerun" ("value", "means") VALUES ('differs', 'the bundle''s re-run wrote something other than this act''s recorded output, exit or out/; drift names what');
INSERT INTO "enum_rerun" ("value", "means") VALUES ('reproduced', 'the bundle''s re-run wrote this act''s recorded output, exit and out/, byte for byte');

CREATE TABLE "enum_inquiry_status" (
  "value" TEXT PRIMARY KEY,
  "means" TEXT NOT NULL
) STRICT;
INSERT INTO "enum_inquiry_status" ("value", "means") VALUES ('abandoned', 'you STARTED it and NO ANSWER CAME — the method broke, the data was unavailable, or the question was given up. REQUIRES a reason saying what stopped it. An approach that ran and refuted its hypothesis is `refuted`, not this');
INSERT INTO "enum_inquiry_status" ("value", "means") VALUES ('confirmed', 'you RAN the approach and the hypothesis HELD — REQUIRES a reason saying what it found, which is in the report; it owes no further fate. An approach that ran and refuted its hypothesis is `refuted`, and one that gave no answer is `abandoned`');
INSERT INTO "enum_inquiry_status" ("value", "means") VALUES ('declined', 'you considered it and did NOT run it, and set it aside: one worth a later run is `deferred`. REQUIRES a reason. An approach you ran is not this, whatever it found: `confirmed` if the hypothesis held, `refuted` if what it found contradicted it, `abandoned` if it was started and no answer came');
INSERT INTO "enum_inquiry_status" ("value", "means") VALUES ('deferred', 'left for a later research run, not this one — kept, not set aside. REQUIRES a reason saying what a later run should pick it up FOR');
INSERT INTO "enum_inquiry_status" ("value", "means") VALUES ('proposed', 'put forward and undecided — the state `inquiry propose` gives an inquiry; it owes a fate, and holds the PASS until it has one that settles it');
INSERT INTO "enum_inquiry_status" ("value", "means") VALUES ('refuted', 'you RAN the approach and what it found CONTRADICTED the hypothesis — REQUIRES a reason saying what it found, which is in the report; it owes no further fate. A refutation is an answer, and the row prints beside the confirmed ones. An approach that gave no answer is `abandoned`, not this');
INSERT INTO "enum_inquiry_status" ("value", "means") VALUES ('underway', 'you are working the question and it is still open — say what you learned in --reason. Each sitting re-records it with what changed, or records the fate it reached');

CREATE TABLE "enum_log_type" (
  "value" TEXT PRIMARY KEY,
  "means" TEXT NOT NULL,
  "seat_may_file" INTEGER NOT NULL CHECK ("seat_may_file" IN (0, 1))
) STRICT;
INSERT INTO "enum_log_type" ("value", "means", "seat_may_file") VALUES ('defect', 'something is broken: it did the wrong thing, or failed where it should have worked. A tool that fails INTERNALLY records this too, as (TOOL, DEFECT)', 1);
INSERT INTO "enum_log_type" ("value", "means", "seat_may_file") VALUES ('estoppel', 'the TOOL refused a mint because the defect lives in text blue applied verbatim from red''s own --fix-new. Recorded by the tool, not filed by the seat: argue it on the original gap, or mint with --supersedes so the lineage is explicit', 0);
INSERT INTO "enum_log_type" ("value", "means", "seat_may_file") VALUES ('failure', 'a call FAILED where no verb could see it. A seat''s tool call, recorded under that seat as the harness saw it: the tool, what it was aimed at, and the error''s first line. Or the ENGINE''s own stop, recorded under `harness` with the error it threw: a run the engine stopped, as against one that was cut off. Recorded by the tool, not filed by a seat; a refusal the record tool gave is `refusal`, not this', 0);
INSERT INTO "enum_log_type" ("value", "means", "seat_may_file") VALUES ('friction', 'Friction is one type of log entry: something cost the seat a call, a guess or an act — a refusal, a guessed name, a misread shape, a workaround — and it notes what it expected and why, whether or not anything should change; an act no surface offers at all is a request, not friction.', 1);
INSERT INTO "enum_log_type" ("value", "means", "seat_may_file") VALUES ('refusal', 'the TOOL refused a seat''s call, and recorded it as it refused: the command, the flags given (never their values) and the refusal''s first line. Recorded by the tool, not filed by the seat; the seat''s own entry says what it expected and where the expectation came from', 0);
INSERT INTO "enum_log_type" ("value", "means", "seat_may_file") VALUES ('request', 'a capability that does not exist — the act you wanted was on no surface, so there was nothing to get wrong. Distinct from a defect because the fix is to build, not to repair', 1);

CREATE TABLE "enum_log_source" (
  "value" TEXT PRIMARY KEY,
  "means" TEXT NOT NULL
) STRICT;
INSERT INTO "enum_log_source" ("value", "means") VALUES ('seat', 'a seat filed this about its own sitting');
INSERT INTO "enum_log_source" ("value", "means") VALUES ('tool', 'the tool emitted this itself, rather than a seat filing it');

CREATE TABLE "register" (
  "event_id" INTEGER PRIMARY KEY REFERENCES "events"("id"),
  "tool_version" TEXT,
  "hook_version" TEXT,
  "agent_id" TEXT,
  "run_via" TEXT,
  "agent_type" TEXT,
  "occasion" TEXT,
  FOREIGN KEY ("occasion") REFERENCES "enum_occasion"("value")
) STRICT;

CREATE TABLE "gate" (
  "event_id" INTEGER PRIMARY KEY REFERENCES "events"("id"),
  "verdict" TEXT,
  FOREIGN KEY ("verdict") REFERENCES "enum_verdict"("value")
) STRICT;

CREATE TABLE "outcome" (
  "event_id" INTEGER PRIMARY KEY REFERENCES "events"("id"),
  "verdict" TEXT NOT NULL,
  "prose" TEXT NOT NULL,
  "verdict_why" TEXT,
  "verdict_basis" TEXT,
  FOREIGN KEY ("verdict") REFERENCES "enum_run_outcome"("value")
) STRICT;

CREATE TABLE "position" (
  "event_id" INTEGER PRIMARY KEY REFERENCES "events"("id"),
  "text" TEXT,
  "position_id" TEXT NOT NULL UNIQUE,
  "class" TEXT,
  FOREIGN KEY ("class") REFERENCES "enum_gap_class"("value")
) STRICT;

CREATE TABLE "halt" (
  "event_id" INTEGER PRIMARY KEY REFERENCES "events"("id"),
  "opinion" TEXT NOT NULL
) STRICT;

CREATE TABLE "certify" (
  "event_id" INTEGER PRIMARY KEY REFERENCES "events"("id"),
  "statement" TEXT NOT NULL
) STRICT;

CREATE TABLE "declare" (
  "event_id" INTEGER PRIMARY KEY REFERENCES "events"("id"),
  "holding" TEXT
) STRICT;

CREATE TABLE "declare_gap_ids" (
  "event_id" INTEGER NOT NULL REFERENCES "declare"("event_id"),
  "ord"      INTEGER NOT NULL,
  "value"    TEXT    NOT NULL,
  PRIMARY KEY ("event_id", "ord")
) STRICT;

CREATE TABLE "motion" (
  "event_id" INTEGER PRIMARY KEY REFERENCES "events"("id"),
  "motion_id" TEXT NOT NULL UNIQUE,
  "basis" TEXT,
  "relief" TEXT,
  "filing_case" TEXT
) STRICT;

CREATE TABLE "motion_grade" (
  "event_id" INTEGER PRIMARY KEY REFERENCES "motion"("event_id"),
  "gap_id" TEXT,
  "dimension" TEXT,
  "proposed" TEXT,
  FOREIGN KEY ("gap_id") REFERENCES "mint"("gap_id"),
  FOREIGN KEY ("dimension") REFERENCES "enum_grade_dimension"("value"),
  FOREIGN KEY ("proposed") REFERENCES "enum_grade"("value")
) STRICT;

CREATE TABLE "motion_petition" (
  "event_id" INTEGER PRIMARY KEY REFERENCES "motion"("event_id"),
  "class" TEXT,
  FOREIGN KEY ("class") REFERENCES "enum_petition_class"("value")
) STRICT;

CREATE TABLE "motion_docket" (
  "event_id" INTEGER PRIMARY KEY REFERENCES "motion"("event_id"),
  "gap_id" TEXT NOT NULL,
  "grounds" TEXT,
  "grounds_quote" TEXT,
  FOREIGN KEY ("gap_id") REFERENCES "mint"("gap_id")
) STRICT;

CREATE TABLE "motion_position" (
  "event_id" INTEGER PRIMARY KEY REFERENCES "motion"("event_id"),
  "position_id" TEXT
) STRICT;

CREATE TABLE "motion_rule" (
  "event_id" INTEGER PRIMARY KEY REFERENCES "events"("id"),
  "motion_id" TEXT NOT NULL,
  "opinion" TEXT NOT NULL,
  "binds" TEXT,
  "relief" TEXT,
  "grade" TEXT,
  "petition" TEXT,
  "ruling_case" TEXT,
  CHECK (("grade" IS NOT NULL) + ("petition" IS NOT NULL) + ("ruling_case" IS NOT NULL) <= 1),
  FOREIGN KEY ("binds") REFERENCES "enum_ruling_binds"("value"),
  FOREIGN KEY ("grade") REFERENCES "enum_grade_ruling"("value"),
  FOREIGN KEY ("petition") REFERENCES "enum_petition_ruling"("value")
) STRICT;

CREATE TABLE "motion_rule_docket" (
  "event_id" INTEGER PRIMARY KEY REFERENCES "motion_rule"("event_id"),
  "disposition" TEXT NOT NULL,
  "principle" TEXT NOT NULL,
  "tension" TEXT NOT NULL,
  "review_flag" TEXT NOT NULL,
  "settled" TEXT NOT NULL,
  "reopens_on" TEXT,
  "final" INTEGER,
  "direction" TEXT,
  CHECK ("final" IS NULL OR "final" IN (0, 1)),
  CHECK ("disposition" = 'remanded' OR "reopens_on" IS NOT NULL OR "final" IS NOT NULL),
  CHECK ("reopens_on" IS NULL OR "final" IS NULL),
  CHECK ("disposition" <> 'remanded' OR (coalesce(trim("direction", ' ' || char(9, 10, 13)), '') <> '' AND "reopens_on" IS NULL AND "final" IS NULL)),
  CHECK ("disposition" = 'remanded' OR "direction" IS NULL),
  FOREIGN KEY ("disposition") REFERENCES "enum_disposition"("value")
) STRICT;

CREATE TABLE "motion_rule_position" (
  "event_id" INTEGER PRIMARY KEY REFERENCES "motion_rule"("event_id"),
  "ruling" TEXT NOT NULL,
  "lens" TEXT,
  "direction" TEXT,
  CHECK ("ruling" = 'upheld' OR ("lens" IS NULL AND "direction" IS NULL)),
  CHECK ("ruling" <> 'upheld' OR coalesce(trim("direction", ' ' || char(9, 10, 13)), '') <> ''),
  FOREIGN KEY ("ruling") REFERENCES "enum_position_ruling"("value")
) STRICT;

CREATE TABLE "motion_rule_position_gap_ids" (
  "event_id" INTEGER NOT NULL REFERENCES "motion_rule_position"("event_id"),
  "ord"      INTEGER NOT NULL,
  "value"    TEXT    NOT NULL,
  PRIMARY KEY ("event_id", "ord")
) STRICT;

CREATE TABLE "mint" (
  "event_id" INTEGER PRIMARY KEY REFERENCES "events"("id"),
  "gap_id" TEXT NOT NULL UNIQUE,
  "mint_key" TEXT,
  "class" TEXT NOT NULL,
  "location" TEXT,
  "about_kind" TEXT,
  "about_ref" TEXT,
  "problem" TEXT NOT NULL,
  "required_fix" TEXT,
  "fix_new" TEXT,
  "fix_basis" TEXT,
  "acceptance_check" TEXT NOT NULL,
  "check_kind" TEXT NOT NULL,
  "severity" TEXT NOT NULL,
  "likelihood" TEXT NOT NULL,
  "impact" TEXT NOT NULL,
  "complexity_cost" TEXT,
  "mint_reason" TEXT,
  "class_material" TEXT NOT NULL,
  CHECK ("class" <> 'unclassified' OR coalesce(trim("mint_reason", ' ' || char(9, 10, 13)), '') <> ''),
  FOREIGN KEY ("class") REFERENCES "enum_gap_class"("value"),
  FOREIGN KEY ("about_kind") REFERENCES "enum_about_kind"("value"),
  FOREIGN KEY ("check_kind") REFERENCES "enum_check_kind"("value"),
  FOREIGN KEY ("severity") REFERENCES "enum_grade"("value"),
  FOREIGN KEY ("likelihood") REFERENCES "enum_grade"("value"),
  FOREIGN KEY ("impact") REFERENCES "enum_grade"("value"),
  FOREIGN KEY ("complexity_cost") REFERENCES "enum_grade"("value"),
  FOREIGN KEY ("class_material") REFERENCES "enum_class_material"("value")
) STRICT;

CREATE TABLE "mint_supersedes" (
  "event_id" INTEGER NOT NULL REFERENCES "mint"("event_id"),
  "ord"      INTEGER NOT NULL,
  "value"    TEXT    NOT NULL,
  PRIMARY KEY ("event_id", "ord")
) STRICT;

CREATE TABLE "mint_found_by" (
  "event_id" INTEGER NOT NULL REFERENCES "mint"("event_id"),
  "ord"      INTEGER NOT NULL,
  "value"    TEXT    NOT NULL,
  PRIMARY KEY ("event_id", "ord")
) STRICT;

CREATE TABLE "close" (
  "event_id" INTEGER PRIMARY KEY REFERENCES "events"("id"),
  "gap_id" TEXT NOT NULL,
  "closure_class" TEXT,
  "anchor_seat" TEXT,
  "anchor_tool" TEXT,
  "anchor_target" TEXT,
  "carried_from" TEXT,
  "successor" TEXT,
  "prose" TEXT,
  CHECK ("closure_class" IS NULL OR "closure_class" IN ('amends_prior', 'defect_accepted', 'defect_owed_elsewhere', 'moot', 'not_a_defect', 'repaired', 'repaired_with_regression')),
  CHECK ("closure_class" <> 'repaired_with_regression' OR "successor" IS NOT NULL),
  CHECK ("carried_from" IS NOT NULL OR "prose" IS NOT NULL),
  FOREIGN KEY ("gap_id") REFERENCES "mint"("gap_id"),
  FOREIGN KEY ("closure_class") REFERENCES "enum_disposition"("value"),
  FOREIGN KEY ("successor") REFERENCES "mint"("gap_id")
) STRICT;

CREATE TABLE "close_anchors" (
  "event_id" INTEGER NOT NULL REFERENCES "close"("event_id"),
  "ord"      INTEGER NOT NULL,
  "value"    TEXT    NOT NULL,
  PRIMARY KEY ("event_id", "ord")
) STRICT;

CREATE TABLE "closing" (
  "event_id" INTEGER PRIMARY KEY REFERENCES "events"("id"),
  "gap_id" TEXT,
  "text" TEXT NOT NULL,
  FOREIGN KEY ("gap_id") REFERENCES "mint"("gap_id")
) STRICT;

CREATE TABLE "regrade" (
  "event_id" INTEGER PRIMARY KEY REFERENCES "events"("id"),
  "gap_id" TEXT,
  "severity" TEXT,
  "likelihood" TEXT,
  "impact" TEXT,
  "complexity_cost" TEXT,
  "basis" TEXT NOT NULL,
  "result_severity" TEXT,
  "result_likelihood" TEXT,
  "result_impact" TEXT,
  "result_complexity_cost" TEXT,
  FOREIGN KEY ("gap_id") REFERENCES "mint"("gap_id"),
  FOREIGN KEY ("severity") REFERENCES "enum_grade"("value"),
  FOREIGN KEY ("likelihood") REFERENCES "enum_grade"("value"),
  FOREIGN KEY ("impact") REFERENCES "enum_grade"("value"),
  FOREIGN KEY ("complexity_cost") REFERENCES "enum_grade"("value"),
  FOREIGN KEY ("result_severity") REFERENCES "enum_grade"("value"),
  FOREIGN KEY ("result_likelihood") REFERENCES "enum_grade"("value"),
  FOREIGN KEY ("result_impact") REFERENCES "enum_grade"("value"),
  FOREIGN KEY ("result_complexity_cost") REFERENCES "enum_grade"("value")
) STRICT;

CREATE TABLE "spot_check" (
  "event_id" INTEGER PRIMARY KEY REFERENCES "events"("id"),
  "reason" TEXT
) STRICT;

CREATE TABLE "spot_check_ids" (
  "event_id" INTEGER NOT NULL REFERENCES "spot_check"("event_id"),
  "ord"      INTEGER NOT NULL,
  "value"    TEXT    NOT NULL,
  PRIMARY KEY ("event_id", "ord")
) STRICT;

CREATE TABLE "spot_check_areas" (
  "event_id" INTEGER NOT NULL REFERENCES "spot_check"("event_id"),
  "ord"      INTEGER NOT NULL,
  "value"    TEXT    NOT NULL,
  PRIMARY KEY ("event_id", "ord")
) STRICT;

CREATE TABLE "finding" (
  "event_id" INTEGER PRIMARY KEY REFERENCES "events"("id"),
  "id" TEXT NOT NULL UNIQUE,
  "finding_key" TEXT,
  "location" TEXT,
  "text" TEXT,
  "severity" TEXT,
  "likelihood" TEXT,
  "impact" TEXT,
  "about_kind" TEXT,
  "about_ref" TEXT,
  FOREIGN KEY ("severity") REFERENCES "enum_grade"("value"),
  FOREIGN KEY ("likelihood") REFERENCES "enum_grade"("value"),
  FOREIGN KEY ("impact") REFERENCES "enum_grade"("value"),
  FOREIGN KEY ("about_kind") REFERENCES "enum_about_kind"("value")
) STRICT;

CREATE TABLE "anchor" (
  "event_id" INTEGER PRIMARY KEY REFERENCES "events"("id"),
  "id" TEXT UNIQUE,
  "location" TEXT,
  "sentence_of" TEXT
) STRICT;

CREATE TABLE "cite" (
  "event_id" INTEGER PRIMARY KEY REFERENCES "events"("id"),
  "label" TEXT,
  "url" TEXT,
  "sha256" TEXT,
  "title" TEXT,
  "location" TEXT,
  "access_date" TEXT,
  "cite_key" TEXT,
  "text" TEXT,
  "source_text_origin" TEXT,
  "source_span" TEXT,
  "ocr_engine" TEXT,
  "ocr_text_sha" TEXT,
  "span_start" INTEGER,
  "span_end" INTEGER,
  "span_text_sha" TEXT,
  "source_before" TEXT,
  "source_matched" TEXT,
  "source_after" TEXT,
  "work_status" TEXT,
  "source_completeness" TEXT,
  "copy_substitution" TEXT,
  "copy_url" TEXT,
  "answers" TEXT,
  FOREIGN KEY ("source_text_origin") REFERENCES "enum_source_text_origin"("value"),
  FOREIGN KEY ("work_status") REFERENCES "enum_work_status"("value"),
  FOREIGN KEY ("source_completeness") REFERENCES "enum_source_completeness"("value"),
  FOREIGN KEY ("copy_substitution") REFERENCES "enum_copy_substitution"("value")
) STRICT;

CREATE TABLE "cite_pages" (
  "event_id" INTEGER NOT NULL REFERENCES "cite"("event_id"),
  "ord"      INTEGER NOT NULL,
  "value"    INTEGER NOT NULL,
  PRIMARY KEY ("event_id", "ord")
) STRICT;

CREATE TABLE "verify" (
  "event_id" INTEGER PRIMARY KEY REFERENCES "events"("id"),
  "claim" TEXT NOT NULL,
  "url" TEXT,
  "title" TEXT,
  "anchor" TEXT,
  "access_date" TEXT,
  "outcome" TEXT NOT NULL,
  "confidence" TEXT NOT NULL,
  "text" TEXT NOT NULL,
  "page" INTEGER,
  "page_render_sha" TEXT,
  "reading_render_sha" TEXT,
  "source_completeness" TEXT,
  "bridge" TEXT,
  "copy_substitution" TEXT,
  "copy_url" TEXT,
  "carried_by" TEXT,
  "carrier_minted" INTEGER,
  CHECK ("carrier_minted" IS NULL OR "carrier_minted" IN (0, 1)),
  FOREIGN KEY ("outcome") REFERENCES "enum_source_outcome"("value"),
  FOREIGN KEY ("confidence") REFERENCES "enum_confidence"("value"),
  FOREIGN KEY ("source_completeness") REFERENCES "enum_source_completeness"("value"),
  FOREIGN KEY ("bridge") REFERENCES "enum_bridge_kind"("value"),
  FOREIGN KEY ("copy_substitution") REFERENCES "enum_copy_substitution"("value"),
  FOREIGN KEY ("carried_by") REFERENCES "mint"("gap_id")
) STRICT;

CREATE TABLE "computation" (
  "event_id" INTEGER PRIMARY KEY REFERENCES "events"("id"),
  "computation_id" TEXT,
  "key" TEXT,
  "bundle_sha" TEXT,
  "answers" TEXT,
  "cites" TEXT,
  "drift" TEXT,
  "text" TEXT,
  "exit" INTEGER,
  "location" TEXT,
  "toolchain_path" TEXT,
  "toolchain_sha256" TEXT,
  "basis" TEXT NOT NULL,
  "invariant" TEXT,
  CHECK (("basis" = 'variant' AND coalesce(trim("drift", ' ' || char(9, 10, 13)), '') <> '' AND coalesce(trim("invariant", ' ' || char(9, 10, 13)), '') <> '') OR ("basis" = 'reproducible' AND "drift" IS NULL AND "invariant" IS NULL)),
  FOREIGN KEY ("basis") REFERENCES "enum_basis"("value")
) STRICT;

CREATE TABLE "computation_argv" (
  "event_id" INTEGER NOT NULL REFERENCES "computation"("event_id"),
  "ord"      INTEGER NOT NULL,
  "value"    TEXT    NOT NULL,
  PRIMARY KEY ("event_id", "ord")
) STRICT;

CREATE TABLE "reproduce" (
  "event_id" INTEGER PRIMARY KEY REFERENCES "events"("id"),
  "bundle_sha" TEXT,
  "soundness" TEXT,
  "note" TEXT,
  "computation_id" TEXT NOT NULL,
  "rerun" TEXT,
  "ran" INTEGER NOT NULL,
  "rerun_exit" INTEGER,
  "drift" TEXT,
  "rerun_sha" TEXT,
  CHECK ("ran" IS NULL OR "ran" IN (0, 1)),
  CHECK (("rerun" = 'differs') = (coalesce(trim("drift", ' ' || char(9, 10, 13)), '') <> '')),
  CHECK ("rerun_exit" IS NULL OR ("ran" = 1 AND "rerun_sha" IS NOT NULL)),
  FOREIGN KEY ("soundness") REFERENCES "enum_soundness"("value"),
  FOREIGN KEY ("rerun") REFERENCES "enum_rerun"("value")
) STRICT;

CREATE TABLE "inquiry" (
  "event_id" INTEGER PRIMARY KEY REFERENCES "events"("id"),
  "inquiry_id" TEXT,
  "question" TEXT,
  "hypothesis" TEXT,
  "method" TEXT,
  "status" TEXT NOT NULL,
  "supersedes_status" TEXT,
  "reason" TEXT,
  FOREIGN KEY ("status") REFERENCES "enum_inquiry_status"("value")
) STRICT;

CREATE TABLE "blue_edit" (
  "event_id" INTEGER PRIMARY KEY REFERENCES "events"("id"),
  "edit_key" TEXT,
  "answers" TEXT,
  "old" TEXT,
  "new" TEXT,
  "text" TEXT,
  "applied_verbatim" INTEGER,
  "accepted" INTEGER,
  "exact_span" INTEGER,
  CHECK ("applied_verbatim" IS NULL OR "applied_verbatim" IN (0, 1)),
  CHECK ("accepted" IS NULL OR "accepted" IN (0, 1)),
  CHECK ("exact_span" IS NULL OR "exact_span" IN (0, 1))
) STRICT;

CREATE TABLE "blue_edit_reopened" (
  "event_id" INTEGER NOT NULL REFERENCES "blue_edit"("event_id"),
  "ord"      INTEGER NOT NULL,
  "value"    TEXT    NOT NULL,
  PRIMARY KEY ("event_id", "ord")
) STRICT;

CREATE TABLE "retire" (
  "event_id" INTEGER PRIMARY KEY REFERENCES "events"("id"),
  "claim" TEXT NOT NULL,
  "reason" TEXT NOT NULL,
  "superseded_by" TEXT,
  "removal_basis" TEXT,
  "claim_kept" INTEGER,
  "answers" TEXT,
  CHECK ("claim_kept" IS NULL OR "claim_kept" IN (0, 1))
) STRICT;

CREATE TABLE "retire_anchors" (
  "event_id" INTEGER NOT NULL REFERENCES "retire"("event_id"),
  "ord"      INTEGER NOT NULL,
  "value"    TEXT    NOT NULL,
  PRIMARY KEY ("event_id", "ord")
) STRICT;

CREATE TABLE "manifest_row" (
  "event_id" INTEGER PRIMARY KEY REFERENCES "events"("id"),
  "gap_id" TEXT,
  "row" TEXT,
  FOREIGN KEY ("gap_id") REFERENCES "mint"("gap_id")
) STRICT;

CREATE TABLE "log" (
  "event_id" INTEGER PRIMARY KEY REFERENCES "events"("id"),
  "text" TEXT,
  "type" TEXT NOT NULL,
  "source" TEXT,
  FOREIGN KEY ("type") REFERENCES "enum_log_type"("value"),
  FOREIGN KEY ("source") REFERENCES "enum_log_source"("value")
) STRICT;

CREATE TABLE "report_part" (
  "event_id" INTEGER PRIMARY KEY REFERENCES "events"("id"),
  "text" TEXT
) STRICT;

CREATE TABLE "sitting_open" (
  "event_id" INTEGER PRIMARY KEY REFERENCES "events"("id"),
  "agent_id" TEXT,
  "agent_type" TEXT,
  "seat_id" TEXT
) STRICT;

CREATE TABLE "sitting_close" (
  "event_id" INTEGER PRIMARY KEY REFERENCES "events"("id"),
  "agent_id" TEXT,
  "agent_transcript_path" TEXT
) STRICT;

CREATE TABLE "cast" (
  "event_id" INTEGER PRIMARY KEY REFERENCES "events"("id")
) STRICT;

CREATE TABLE "cast_seat_ids" (
  "event_id" INTEGER NOT NULL REFERENCES "cast"("event_id"),
  "ord"      INTEGER NOT NULL,
  "value"    TEXT    NOT NULL,
  PRIMARY KEY ("event_id", "ord")
) STRICT;

CREATE TABLE "cast_lane_seat_ids" (
  "event_id" INTEGER NOT NULL REFERENCES "cast"("event_id"),
  "ord"      INTEGER NOT NULL,
  "value"    TEXT    NOT NULL,
  PRIMARY KEY ("event_id", "ord")
) STRICT;

CREATE TABLE "dispatch" (
  "event_id" INTEGER PRIMARY KEY REFERENCES "events"("id"),
  "pin" INTEGER NOT NULL,
  "seat_id" TEXT NOT NULL
) STRICT;

CREATE TABLE "dispatch_gap_ids" (
  "event_id" INTEGER NOT NULL REFERENCES "dispatch"("event_id"),
  "ord"      INTEGER NOT NULL,
  "value"    TEXT    NOT NULL,
  PRIMARY KEY ("event_id", "ord")
) STRICT;

CREATE TABLE "sitting_limit" (
  "event_id" INTEGER PRIMARY KEY REFERENCES "events"("id"),
  "seat_id" TEXT,
  "sitting" INTEGER,
  "limit" INTEGER
) STRICT;

CREATE TABLE "plan" (
  "event_id" INTEGER PRIMARY KEY REFERENCES "events"("id"),
  "printed" TEXT NOT NULL
) STRICT;

CREATE TABLE "stale" (
  "event_id" INTEGER PRIMARY KEY REFERENCES "events"("id"),
  "gap_id" TEXT,
  "through" INTEGER,
  FOREIGN KEY ("gap_id") REFERENCES "mint"("gap_id"),
  FOREIGN KEY ("through") REFERENCES "events"("id")
) STRICT;

CREATE TABLE "report_read" (
  "event_id" INTEGER PRIMARY KEY REFERENCES "events"("id"),
  "head" INTEGER NOT NULL,
  FOREIGN KEY ("head") REFERENCES "events"("id")
) STRICT;

CREATE TABLE "envelope_admitted" (
  "event_id" INTEGER PRIMARY KEY REFERENCES "events"("id"),
  "agent_id" TEXT NOT NULL
) STRICT;

CREATE TABLE "inquiry_check" (
  "event_id" INTEGER PRIMARY KEY REFERENCES "events"("id"),
  "inquiry_id" TEXT NOT NULL,
  "reason" TEXT NOT NULL
) STRICT;

CREATE INDEX "gate_verdict" ON "gate" ("verdict");
-- A gap's latest regrade, whose stored result grades the write path overlays the next regrade on
-- (recordsql.stampRegradeResult).
CREATE INDEX "regrade_gap" ON "regrade" ("gap_id", "event_id");
-- A motion's rulings: motion_answers, inquiry_state, the gap view's unruled docket filing, and the
-- bench's docket readers.
CREATE INDEX "motion_rule_motion" ON "motion_rule" ("motion_id");
-- The docket motions on a gap: the bench's latest ruling on it and its unruled filing.
CREATE INDEX "motion_docket_gap" ON "motion_docket" ("gap_id");
-- The evidence anchors placed as a gap's answer, which the gap view's live-answer demand and the
-- closing and estoppel checks ask per gap: a computation's and a citation's.
CREATE INDEX "computation_answers" ON "computation" ("answers");
CREATE INDEX "cite_answers" ON "cite" ("answers");
-- The retires naming an evidence anchor: its first is the retire the evidence_anchor view reads,
-- and the retirement that retired an anchor (AnchorRetiredAt).
CREATE INDEX "retire_anchors_value" ON "retire_anchors" ("value", "event_id");
-- The computations applying a citation as their method (Computation.cites): what a retire of that
-- citation is refused over (record.requireCitationUncited).
CREATE INDEX "computation_cites" ON "computation" ("cites");
-- The gaps minted about an evidence anchor, read on the gap's side of a retire's reach
-- (record.reachingRetiresOf) and by the demand_reach view the gap view's evidence_met reads.
CREATE INDEX "mint_about" ON "mint" ("about_kind", "about_ref");
-- The readings a gap carries (carrying_reading), asked per gap by the gap's demand
-- (record.EvidenceDemandOf). Partial: a reading that carries no gap is never asked this way.
CREATE INDEX "verify_carried_by" ON "verify" ("carried_by") WHERE "carried_by" IS NOT NULL;
-- The harness brackets naming a seat: the other half of a seat's latest sitting
-- (recordsql.LatestSitting).
CREATE INDEX "sitting_open_seat" ON "sitting_open" ("seat_id");
-- An agent's latest bracket, which its admission and its close are each held to (OneAdmissionASitting,
-- OneCloseASitting), and which sittingOf reads on every write.
CREATE INDEX "sitting_open_agent" ON "sitting_open" ("agent_id");
-- An agent's registrations: the seat it registered as (RegisteredSeatOfAgent) and the sitting it
-- holds (registeredSitting).
CREATE INDEX "register_agent" ON "register" ("agent_id");
-- A seat's keyed retry of an act, found by the key it sent (ExistingBlueEditByKey, FindingByKey,
-- ExistingMintByKey, ExistingCiteByKey, ExistingComputationByKey, ExistingProposalByText).
CREATE INDEX "blue_edit_key" ON "blue_edit" ("edit_key");
CREATE INDEX "finding_key" ON "finding" ("finding_key");
CREATE INDEX "mint_key" ON "mint" ("mint_key");
CREATE INDEX "cite_key" ON "cite" ("cite_key");
CREATE INDEX "computation_key" ON "computation" ("key");
CREATE INDEX "inquiry_question" ON "inquiry" ("question");
-- The anchors an act names, each checked to exist: a citation by its label and a computation by
-- its id (requireAboutAnchor, requireComputationAnchor, requireNewGrounds), an inquiry by its id
-- (requireInquiry, requireInquiryRef, requireClosureAnchors).
CREATE INDEX "cite_label" ON "cite" ("label");
CREATE INDEX "computation_id" ON "computation" ("computation_id");
CREATE INDEX "inquiry_id" ON "inquiry" ("inquiry_id");
-- A computation bundle's latest execution (requireNamedExecution).
CREATE INDEX "reproduce_bundle" ON "reproduce" ("bundle_sha");
-- A gap's closures (priorClosureEpochs, requireNoStandingDocket).
CREATE INDEX "close_gap" ON "close" ("gap_id");

-- A SITTING IS A ROW THAT NAMES ITSELF, and this view is the one place that says so.
--
-- events.sitting_id is the opening event of the sitting a row belongs to, stamped at the write
-- (schema.go). An OPENING is therefore exactly a row whose sitting_id is its own id — there is no
-- predicate here to keep level with the Go one, because the predicate ran once, at the insert.
--
-- The OWNER is the other half, and it is the reason this is a view rather than a WHERE clause
-- repeated five times: a register's sitting belongs to the seat on its envelope, and a harness
-- bracket's belongs to the seat its BODY names — the configuration resolved once by the writer
-- (#1149). Every reader that asks "which sittings has this seat had" joins here and gets both.
--
-- The one question it cannot answer by index is a seat's LATEST sitting: its seat is a COALESCE
-- across the join, so a max over it by seat walks the events from the end. recordsql.LatestSitting
-- asks that question, as this rule split at the COALESCE into two indexed lookups, and is held to
-- this view on every archived run.
CREATE VIEW "sittings" AS
SELECT e."id"                                                AS "id",
       COALESCE(NULLIF(o."seat_id", ''), e."seat_id")         AS "seat_id",
       e."ts"                                                AS "ts",
       e."type"                                              AS "opened_by"
FROM "events" e
LEFT JOIN "sitting_open" o ON o."event_id" = e."id"
WHERE e."sitting_id" = e."id";

-- THE AGENT -> SEAT BINDING, the one statement of the rule; every reader of "which seat is this
-- agent" asks it here (record.SeatOfAgent, and seat_of_agent below for the windows):
--
--   AN AGENT'S SEAT IS NAMED BY ITS LATEST REGISTER. A re-dispatch writes a fresh register event,
--   and a resumed seat legitimately arrives under a new agent id claiming a seat that is already
--   bound; treating that as a conflict would refuse every resume.
--
--   AN AGENT NO REGISTER NAMES IS BOUND BY ITS LATEST HOOK BRACKET, IF THAT BRACKET NAMES A SEAT.
--   A seat woken with nothing to do need not register — the SubagentStart hook brackets its sitting
--   and the writer resolves the configuration to a seat — so a register-only binding reports those
--   agents as anonymous, and every total joined to it drops their rows or reports a null seat. The
--   bracket that counts is the agent's LATEST, seat-naming or not: a later bracket naming no seat
--   says this agent's configuration seats several, and naming an earlier seat would be a guess the
--   record no longer makes. That agent is unbound.
--
-- The register WINS where both exist: it is the seat's own word. agent_type comes along from
-- whichever act binds: it is what the HARNESS called the seat, beside what the seat called itself,
-- and the two disagreeing is a thing worth being able to see rather than a thing to collapse here.
-- occasion is what the register says its sitting was convened to do — the bench's, whose four
-- sittings share one id; NULL on a bracket, which states none.
--
-- NO WINDOWS HERE, deliberately: the hook and every seat call ask this binding alone, and
-- seat_of_agent adds the windows for the readers that want them.
CREATE VIEW "agent_binding" AS
SELECT
  r."agent_id"   AS "agent_id",
  e."seat_id"    AS "seat_id",
  r."agent_type" AS "agent_type",
  r."occasion"   AS "occasion",
  r."event_id"   AS "event_id"
FROM "register" r
JOIN "events" e ON e."id" = r."event_id"
WHERE r."agent_id" IS NOT NULL AND r."agent_id" != ''
  AND r."event_id" = (SELECT MAX(r2."event_id") FROM "register" r2 WHERE r2."agent_id" = r."agent_id")
UNION ALL
SELECT
  o."agent_id"   AS "agent_id",
  o."seat_id"    AS "seat_id",
  o."agent_type" AS "agent_type",
  NULL           AS "occasion",
  o."event_id"   AS "event_id"
FROM "sitting_open" o
WHERE o."agent_id" IS NOT NULL AND o."agent_id" != ''
  AND o."seat_id" IS NOT NULL AND o."seat_id" != ''
  AND NOT EXISTS (SELECT 1 FROM "register" r2 WHERE r2."agent_id" = o."agent_id")
  AND o."event_id" = (SELECT MAX(o2."event_id") FROM "sitting_open" o2 WHERE o2."agent_id" = o."agent_id");

-- THE BINDING WITH THE TWO WINDOWS OF THE ACT THAT BOUND IT: the EPOCH (the chair sittings at or
-- before it) and the SITTING (that seat's own sitting count, the opening that began it included),
-- read off the binding act's stored columns as every window is.
CREATE VIEW "seat_of_agent" AS
SELECT
  b."agent_id"   AS "agent_id",
  b."seat_id"    AS "seat_id",
  b."agent_type" AS "agent_type",
  e."epoch"      AS "registered_epoch",
  e."sitting"    AS "sitting",
  b."occasion"   AS "occasion"
FROM "agent_binding" b
JOIN "events" e ON e."id" = b."event_id";

-- WHAT A SEAT COST, from the turns ingested at capture (#684 F16).
--
-- A LEFT JOIN, deliberately. A seat whose turns were measured but which nothing binds — no register
-- and no bracket naming its seat — still appears, with a null seat_id — the alternative drops its
-- rows from every total and reports a
-- cheaper run than happened, which is the failure this whole issue keeps finding.
--
-- NULLIF(ts_ms, 0) is the parser's contract honoured in SQL: 0 means the line carried no
-- timestamp, and letting it into MIN() would put the seat's start at the epoch and make wall_ms
-- the age of the universe. A seat with no timestamped turn gets a null span, which is "not
-- measured" and not zero.
CREATE VIEW "seat_metrics" AS
SELECT
  t."agent_id"                                             AS "agent_id",
  s."seat_id"                                              AS "seat_id",
  s."agent_type"                                           AS "agent_type",
  COUNT(*)                                                 AS "turns",
  SUM(t."is_thinking")                                     AS "thinking_turns",
  SUM(t."is_tool")                                         AS "tool_turns",
  SUM(t."input_tokens")                                    AS "input_tokens",
  SUM(t."output_tokens")                                   AS "output_tokens",
  SUM(t."cache_read")                                      AS "cache_read",
  SUM(t."cache_creation")                                  AS "cache_creation",
  MIN(NULLIF(t."ts_ms", 0))                                AS "first_ts_ms",
  MAX(NULLIF(t."ts_ms", 0))                                AS "last_ts_ms",
  MAX(NULLIF(t."ts_ms", 0)) - MIN(NULLIF(t."ts_ms", 0))    AS "wall_ms",
  -- WHAT THE SITTING BOUGHT, so its cost can be read against it.
  --
  -- turns, tool_turns and wall_ms say what a sitting SPENT. None of them says whether it
  -- produced anything, and without that a reader cannot tell a hard sitting from a wasteful one:
  -- chair time confounds task complexity with efficiency, and comparing two runs on duration
  -- alone credits or blames the engine for a question that simply got easier.
  --
  -- With this column the decomposition is readable off one row:
  --   tool_turns / acts   calls per recorded act — the part the engine's design controls
  --   wall_ms / tool_turns   time per call — mostly the environment (model, box, cache)
  --   acts = 0            an EMPTY sitting, which has no task complexity in it at all, so its
  --                       cost is pure overhead and needs no normalising
  --
  -- BOOKKEEPING IS NOT AN ACT: the types excluded are NotActs, the one list every count of acts
  -- reads, so no two of them can count one sitting differently.
  --
  -- SCOPED TO THE SITTING, not the run: events carries the sitting ordinal and seat_of_agent
  -- carries this agent's, so the count is what THIS sitting recorded rather than what the seat
  -- has ever recorded. A seat's second sitting borrowing its first one's acts is the same
  -- defect the work list had before duties were made per-sitting.
  (SELECT COUNT(*) FROM "events" ew
     WHERE ew."seat_id" = s."seat_id" AND ew."sitting" = s."sitting"
       AND ew."type" NOT IN ('register','log','sitting_open','sitting_close','stale','report_read','envelope_admitted','plan'))  AS "acts"
FROM "seat_turn" t
LEFT JOIN "seat_of_agent" s ON s."agent_id" = t."agent_id"
GROUP BY t."agent_id";

-- HOW LONG EACH TURN TOOK, which the transcript never states: a turn's span is the gap to the one
-- before it, so it is a window function over the seat's own ordering and cannot be a column.
--
-- The FIRST turn of a seat has no predecessor and gets a null span rather than a zero. Zero would
-- say "instant", and a bucket summing it would quietly under-report every seat by its opening
-- turn.
CREATE VIEW "seat_turn_span" AS
SELECT
  "agent_id"      AS "agent_id",
  "turn_idx"      AS "turn_idx",
  "ts_ms"         AS "ts_ms",
  "is_thinking"   AS "is_thinking",
  "is_tool"       AS "is_tool",
  "output_tokens" AS "output_tokens",
  "ts_ms" - LAG("ts_ms") OVER (PARTITION BY "agent_id" ORDER BY "turn_idx") AS "span_ms"
FROM "seat_turn"
WHERE "ts_ms" > 0;

-- WHERE A SEAT'S WALL CLOCK WENT, bucketed by WHAT THE TURN CONTAINED.
--
-- #684 F11 decomposed a run by hand into thinking / round-trip tail / big generation / stalls, and
-- those last three are threshold judgements — 70-83 tok/s is "healthy generation", a 16-minute
-- turn with 66 output tokens is "a stall". THOSE THRESHOLDS ARE NOT ENCODED HERE, on purpose. A
-- number chosen once from one run, frozen into the schema, would be applied to every future run by
-- readers who never saw it chosen, and a view is the worst place to hide a judgement because it
-- looks like a measurement.
--
-- So the split is by fact: did the turn carry a thinking block, a tool_use block, both, or
-- neither. An analysis that wants F11's buckets has span_ms and output_tokens per turn in
-- seat_turn_span and can apply its own cutoffs, in the open, where the next reader can disagree.
CREATE VIEW "seat_time_decomposition" AS
SELECT
  "agent_id"                     AS "agent_id",
  CASE
    WHEN "is_thinking" = 1 AND "is_tool" = 1 THEN 'thinking+tool'
    WHEN "is_thinking" = 1                   THEN 'thinking'
    WHEN "is_tool" = 1                       THEN 'tool'
    ELSE 'text'
  END                            AS "bucket",
  COUNT(*)                       AS "turns",
  SUM("span_ms")                 AS "span_ms",
  SUM("output_tokens")           AS "output_tokens"
FROM "seat_turn_span"
GROUP BY "agent_id", "bucket";

-- THE CHANGE LOG, AS A VIEW: every recorded edit to the report with the text on both sides.
--
-- Folding the WHOLE event stream in Go to project ONE event family is the most expensive way to ask
-- this question: MergedEvents loads every row and every body, so its cost scales with the size of
-- the RECORD, not with the number of edits. This view reads the edit rows and their clock.
--
-- Nothing here is new knowledge. The edits are one table, and the clock is columns of events.
--
-- THE ARMS ARE UNIONED FIRST AND JOINED TO events ONCE, so the join is planned over the change rows
-- rather than once per arm.
--
-- "delta" is in CHARACTERS because SQLite's length() counts code points on TEXT, which is what Go's
-- len([]rune(s)) counted. length() on a BLOB would count bytes, and these columns are TEXT.
--
-- AN EVIDENCE ANCHOR'S RETIRE IS A CHANGE TOO. It takes a citation's or a computation's anchor off a
-- sentence the report keeps word for word, and it names the gap it answers with the key an edit
-- carries — so the gap's change log lists it beside the edits, or an anchor blue took off a sentence
-- because a gap asked for it reads as no answer at all. Its row holds the sentence on both sides (the
-- words did not move) and "retired" names the anchor that left; "retired" is '' on an edit.
CREATE VIEW "change" AS
SELECT
  e."seat_id"                           AS "seat_id",
  e."sitting"                           AS "sitting",
  e."epoch"                             AS "epoch",
  x."answers"                           AS "answers",
  x."old"                               AS "old",
  x."new"                               AS "new",
  x."delta"                             AS "delta",
  x."reason"                            AS "reason",
  x."applied_verbatim"                  AS "applied_verbatim",
  x."accepted"                          AS "accepted",
  x."retired"                           AS "retired",
  x."event_id"                          AS "event_id"
FROM (
  SELECT COALESCE(b."answers", '')          AS "answers",
         COALESCE(b."old", '')              AS "old",
         COALESCE(b."new", '')              AS "new",
         length(COALESCE(b."new", '')) - length(COALESCE(b."old", '')) AS "delta",
         COALESCE(b."text", '')             AS "reason",
         COALESCE(b."applied_verbatim", 0)  AS "applied_verbatim",
         COALESCE(b."accepted", 0)          AS "accepted",
         ''                                 AS "retired",
         b."event_id"                       AS "event_id"
  FROM "blue_edit" b
  UNION ALL
  SELECT COALESCE(r."answers", ''), r."claim", r."claim", 0, r."reason", 0, 0, ra."value", r."event_id"
  FROM "retire" r
  JOIN "retire_anchors" ra ON ra."event_id" = r."event_id"
  WHERE r."claim_kept" = 1
) x
JOIN "events" e ON e."id" = x."event_id";

-- THE BENCH'S REMANDS: every docket ruling with a disposition that does not close the gap
-- ("remanded"), one row per ruling — the gap it is about, its events.id, the sitting it was ruled in,
-- the research direction it states, which the DocketRuling CHECKs require of a remand, and the
-- proposition it settles, empty where it settles none — and every gap an upheld ruling on the chair's
-- withheld PASS sends back for one more exchange, one row per gap (gblock, 2026-10-30): that is a
-- remand too, on the ruling's direction, settling nothing, and the gap's one remand at impasse like
-- any other.
--
-- ONE DEFINITION FOR EVERY READER OF A REMAND. The gap view's "remanded" and "docket_direction"
-- read it, and so does the dispatch's remand fold (record.remandRulingsOf), which groups the rows
-- into bench sittings by "sitting_id" and counts a sitting only where the gap was at impasse when the
-- bench ruled — a question about the exchange count, which the record's Go fold answers and this
-- view cannot.
CREATE VIEW "remand" AS
SELECT md."gap_id"                          AS "gap_id",
       mr."event_id"                        AS "event_id",
       COALESCE(e."sitting_id", e."id")     AS "sitting_id",
       rd."direction"                       AS "direction",
       rd."settled"                         AS "settled"
FROM "motion_docket" md
JOIN "motion" mo ON mo."event_id" = md."event_id"
JOIN "motion_rule" mr ON mr."motion_id" = mo."motion_id"
JOIN "motion_rule_docket" rd ON rd."event_id" = mr."event_id"
JOIN "enum_disposition" d ON d."value" = rd."disposition"
JOIN "events" e ON e."id" = mr."event_id"
WHERE NOT d."closes"
UNION ALL
SELECT pg."value"                           AS "gap_id",
       mr."event_id"                        AS "event_id",
       COALESCE(e."sitting_id", e."id")     AS "sitting_id",
       rp."direction"                       AS "direction",
       NULL                                 AS "settled"
FROM "motion_rule_position_gap_ids" pg
JOIN "motion_rule_position" rp ON rp."event_id" = pg."event_id"
JOIN "motion_rule" mr ON mr."event_id" = rp."event_id"
JOIN "events" e ON e."id" = mr."event_id"
WHERE rp."ruling" = 'upheld';

-- A GAP'S GRADES OVER THE RECORD, AND WHETHER IT IS MATERIAL: one row per gap (per mint) per stretch
-- of the record over which its grades held — from "from_seq" up to, not including, "until_seq" (NULL:
-- they hold now). The row that holds at a point P is the one with from_seq <= P and until_seq NULL or
-- above P. The gap view reads the row that holds now; convergence_at reads the one that held at each
-- verdict. This is the ONE statement of the materiality rule.
--
-- THE GRADES ARE STORED, NOT DERIVED HERE. A stretch begins at the mint and at each regrade, the
-- only points where a gap's grades can change: the mint's row carries the mint's grades, and a
-- regrade's row the result grades the write path stamped on it (recordsql.stampRegradeResult, the one
-- statement of the regrade overlay). LEAD over the stretches of one mint ends each at the next.
--
-- THE MATERIALITY RULE: the class's default says 'always' or 'never', and a 'by_grade' class is
-- material at a severity of medium (mass 2.0) and above. A gap that is not material stays on the
-- board, and blue may answer it when engaged for something else, but it alone cannot spin the cycle
-- (gblock, 2026-09-08).
--
-- "mass" is MASS[likelihood] * MASS[impact] and "severity_mass" MASS[severity] — a grade's weight is
-- a facet on the vocabulary (enum_grade.mass).
CREATE VIEW "gap_grade" AS
WITH "stretch" AS (
  SELECT m."event_id" AS "minted_event", m."event_id" AS "from_seq", m."class_material" AS "class_material",
         m."severity" AS "severity", m."likelihood" AS "likelihood", m."impact" AS "impact",
         m."complexity_cost" AS "complexity_cost"
  FROM "mint" m
  UNION ALL
  SELECT m."event_id", a."event_id", m."class_material",
         a."result_severity", a."result_likelihood", a."result_impact", a."result_complexity_cost"
  FROM "regrade" a
  JOIN "mint" m ON m."gap_id" = a."gap_id"
),
"graded" AS (
  SELECT s.*, LEAD(s."from_seq") OVER (PARTITION BY s."minted_event" ORDER BY s."from_seq") AS "until_seq"
  FROM "stretch" s
)
SELECT g."minted_event"                                      AS "minted_event",
       g."from_seq"                                          AS "from_seq",
       g."until_seq"                                         AS "until_seq",
       g."severity"                                          AS "severity",
       g."likelihood"                                        AS "likelihood",
       g."impact"                                            AS "impact",
       g."complexity_cost"                                   AS "complexity_cost",
       COALESCE(gl."mass", 0.0) * COALESCE(gi."mass", 0.0)   AS "mass",
       COALESCE(gs."mass", 0.0)                              AS "severity_mass",
       (g."class_material" = 'always'
          OR (g."class_material" = 'by_grade' AND COALESCE(gs."mass", 0.0) >= 2.0)) AS "material"
FROM "graded" g
LEFT JOIN "enum_grade" gl ON gl."value" = g."likelihood"
LEFT JOIN "enum_grade" gi ON gi."value" = g."impact"
LEFT JOIN "enum_grade" gs ON gs."value" = g."severity";

-- EVERY ACT THAT CLOSES A GAP, by its arm: red's close ('close'), a docket ruling whose
-- disposition closes ('bench') — whether it closes is read off the vocabulary rather than decided
-- here — and the mint that superseded it ('supersede'). The gap view reads the LAST of each arm, and
-- which arm acted last; convergence_at asks whether any was written at or before a verdict.
--
-- A MINT SUPERSEDING AN OPEN GAP CLOSES IT (ruled by gblock 2026-11-04): the ancestor's purpose
-- passes to the successor in that act, and it has no state of its own left to hold the board. The
-- arm is DERIVED from the mint's own "supersedes" list rather than written as a second event: the record
-- already holds the link, and a close the tool wrote beside the mint would be a second copy of it
-- and an act the seat never stated. It is the FIRST such mint, and only where no closing act of
-- another arm precedes it: a mint naming a gap already closed carries its lineage forward and
-- closes nothing.
--
-- TWO HOPS ON THE BENCH ARM, BECAUSE THE GAP RIDES THE FILING. The bench's disposition was its own
-- event carrying a gap_id; it is a docket motion's RULING, and the gap is on the motion that asked.
-- So: the ruling arm gives the disposition, its motion_rule gives the motion id, and the docket
-- FILING gives the gap. Written here once rather than at each reader — the same join was
-- hand-written at eight readers before it.
--
-- SQLite does not validate a view body at CREATE, so a view left reading a table that has gone
-- applies cleanly and returns no bench closures at all: every disposed gap reading as undisposed.
CREATE VIEW "closure_act" AS
WITH "red" AS (
  SELECT c."gap_id" AS "gap_id", c."event_id" AS "event_id"
  FROM "close" c
),
"bench" AS (
  SELECT md."gap_id" AS "gap_id", mr."event_id" AS "event_id"
  FROM "motion_rule_docket" rd
  JOIN "motion_rule" mr ON mr."event_id" = rd."event_id"
  JOIN "motion" mo ON mo."motion_id" = mr."motion_id"
  JOIN "motion_docket" md ON md."event_id" = mo."event_id"
  JOIN "enum_disposition" d ON d."value" = rd."disposition"
  WHERE d."closes"
)
SELECT "gap_id", "event_id", 'close' AS "arm" FROM "red"
UNION ALL
SELECT "gap_id", "event_id", 'bench' FROM "bench"
UNION ALL
SELECT s."value", MIN(s."event_id"), 'supersede'
FROM "mint_supersedes" s
WHERE NOT EXISTS(SELECT 1 FROM "red" r WHERE r."gap_id" = s."value" AND r."event_id" < s."event_id")
  AND NOT EXISTS(SELECT 1 FROM "bench" b WHERE b."gap_id" = s."value" AND b."event_id" < s."event_id")
GROUP BY s."value";

-- THE READINGS A GAP CARRIES: each verify whose carried_by names the gap, with the anchor it read
-- and the seat that read it. EVERY reader of the reading-to-gap link reads it here and nowhere
-- else (demand_reach, and through it the gap view's evidence_answered, among them), so the day the link is stated
-- from the gap's side instead of the reading's this view is the one statement that changes.
CREATE VIEW "carrying_reading" AS
SELECT v."carried_by" AS "gap_id", v."anchor" AS "anchor", e."seat_id" AS "seat_id", v."event_id" AS "event_id"
FROM "verify" v
JOIN "events" e ON e."id" = v."event_id"
WHERE v."carried_by" IS NOT NULL;

-- EVERY EVIDENCE ANCHOR ON THE RECORD, ONE ROW EACH: a citation or a computation, the check kind
-- a live one answers, the gap it was placed as the answer to, the act that placed it and the
-- retire that took it out (NULL while it stands). One arm per recordsql.EvidenceTables row,
-- generated from the row, so a backing kind added to that table lands here with no SQL written.
-- A retired anchor never comes back (an edit invents no anchor), so its retire is the first
-- retire naming it: one lookup per anchor on retire_anchors_value, never a grouped walk of every
-- retire joined back.
CREATE VIEW "evidence_anchor" AS
WITH "placed" AS (
  SELECT x."label" AS "anchor_id", 'citation' AS "kind", 'source' AS "check_kind",
         x."answers" AS "answers", x."event_id" AS "placed_event"
  FROM "cite" x
  WHERE x."label" IS NOT NULL
  UNION ALL
  SELECT x."computation_id" AS "anchor_id", 'computation' AS "kind", 'computation' AS "check_kind",
         x."answers" AS "answers", x."event_id" AS "placed_event"
  FROM "computation" x
  WHERE x."computation_id" IS NOT NULL
)
SELECT p."anchor_id" AS "anchor_id", p."kind" AS "kind", p."check_kind" AS "check_kind",
       p."answers" AS "answers", p."placed_event" AS "placed_event",
       (SELECT MIN(ra."event_id") FROM "retire_anchors" ra WHERE ra."value" = p."anchor_id") AS "retired_event"
FROM "placed" p;

-- THE ANCHORS A GAP'S LIVE-ANSWER DEMAND REACHES, and whose check of each it is: the anchor the gap
-- is about, checked by its minter, and the anchor of each reading it carries, checked by that
-- reading's seat. A retire of one, or that seat's check of it again, meets the demand (the gap view's
-- evidence_met arms 2 and 3); this is the one statement of the reach, which the close's refusal
-- (record.EvidenceDemandOf) and a retire's reach (record.reachingRetiresOf) read too.
--
-- THE DEMAND MOVES TO THE SUCCESSOR, AND ITS REACH WITH IT (ruled by gblock 2026-11-05). A mint
-- superseding an open gap of the same source or computation check kind closes it and takes its
-- demand on, so the successor's demand is met through the ancestor's anchors too: retiring the
-- citation a refuted reading read is the remedy its failed check names, and it meets the successor.
-- Transitively, through every same-kind ancestor of the chain of superseding mints. An inherited
-- row is the successor's from its mint: a retire before it had met the ancestor, which then passed no
-- demand on, and a check counts for the gap only at or after its mint. An ancestor whose demand was
-- met, or disposed of by the bench, before the successor named it passes its reach on all the same:
-- what would have met it after the successor's mint meets the successor too (plans/compute-and-show.md,
-- "Supersede demand").
--
--   from_gap  the gap whose about or reading the row is: gap_id itself, or an ancestor it carries
--             the demand from.
--   minted    gap_id's mint event: a check of the anchor counts for gap_id at or after it.
--   reached   the event from which the anchor is in gap_id's reach: its own mint for the anchor it
--             is about, the reading's act for a reading it carries, its mint for an inherited row.
CREATE VIEW "demand_reach" AS
WITH RECURSIVE "carries" AS MATERIALIZED (
  -- EACH SUPERSESSION THAT PASSED A DEMAND ON: the first mint naming the ancestor in its supersedes,
  -- at the ancestor's check kind, one a live answer is demanded under, with the ancestor's subject
  -- and mint. The first naming mint is the one that closed an open ancestor (closure_act's
  -- 'supersede' arm). It is not asked whether the ancestor was open, or its demand unmet, when the
  -- mint named it: the gap view's closure is not this view's to read (closure_act a second time in
  -- every statement over the gap view), and that ancestor's demand was met, or disposed of by the
  -- bench, before the successor's mint, which the inherited rows' acts all follow.
  SELECT h."gap_id" AS "heir", a."gap_id" AS "ancestor", a."about_kind" AS "about_kind", a."about_ref" AS "about_ref",
         a."event_id" AS "event_id"
  FROM (SELECT "value", MIN("event_id") AS "by" FROM "mint_supersedes" GROUP BY "value") fs
  JOIN "mint" h ON h."event_id" = fs."by"
  JOIN "mint" a ON a."gap_id" = fs."value" AND a."check_kind" = h."check_kind"
  WHERE a."check_kind" IN ('source', 'computation')
),
"lineage" AS (
  -- EACH GAP, AND EVERY ANCESTOR IT CARRIES THE DEMAND FROM, with that gap's subject and mint.
  SELECT m."gap_id" AS "gap_id", m."gap_id" AS "from_gap", m."check_kind" AS "check_kind", m."event_id" AS "minted",
         m."about_kind" AS "about_kind", m."about_ref" AS "about_ref", m."event_id" AS "from_event"
  FROM "mint" m
  UNION
  SELECT l."gap_id", cy."ancestor", l."check_kind", l."minted", cy."about_kind", cy."about_ref", cy."event_id"
  FROM "lineage" l
  JOIN "carries" cy ON cy."heir" = l."from_gap"
)
SELECT l."gap_id" AS "gap_id", l."from_gap" AS "from_gap", l."check_kind" AS "check_kind", l."minted" AS "minted",
       'about' AS "via", l."about_ref" AS "anchor", e."seat_id" AS "seat", l."minted" AS "reached"
FROM "lineage" l
JOIN "events" e ON e."id" = l."from_event"
WHERE l."about_kind" = 'anchor'
UNION ALL
SELECT l."gap_id", l."from_gap", l."check_kind", l."minted",
       'reading', cr."anchor", cr."seat_id", CASE WHEN l."from_gap" = l."gap_id" THEN cr."event_id" ELSE l."minted" END
FROM "lineage" l
JOIN "carrying_reading" cr ON cr."gap_id" = l."from_gap";

-- THE GAP, AND WHETHER IT IS MATERIAL. Its current grades and its "material" column are gap_grade's
-- row that stands now (until_seq IS NULL) — the one definition of the materiality rule, which
-- convergence_at reads at every other point of the record.
--
-- THIS COLUMN LIST IS HAND-AUTHORED, AND IT IS WHERE A NEW GAP FIELD GOES MISSING. The table DDL
-- is DERIVED from the protobuf descriptors, so a new field on Mint gets its column and its write
-- for free. The READ is hand-written at every step: this list, and then each reader naming the
-- columns again. A field that stops at either stores correctly and renders NOWHERE, with no error
-- anywhere — the JSON simply omits it, which reads exactly like a gap with nothing to say there.
-- Mint's about_kind/about_ref is the measured case: the row right in SQLite, the board printing
-- an empty location.
--
-- A grep for this view's name in a FROM clause finds the readers. Widen the ones that ENUMERATE
-- a gap's content — record/viewjson.go's BOARD query (its rows.Scan and its GapJSON literal) and
-- WORK query (its Scan, WorkGapState, and the WorkGapJSON / ClosedIndexJSON assembly), and
-- record/gapstates.go, which builds the Go Gap struct most other readers take. The rest ask one
-- narrow question of the view (dispatch's open-gap fold, estoppel's credit join) and are unaffected
-- by a new descriptive field. Grep rather than trust the list above: readers of this view are added
-- without touching it, and a count written here goes stale silently — which is this comment's own
-- failure mode, not just the schema's.
--
-- A field added to Finding does not have this problem: the findings view reads the proto.
--
-- EACH SUB-QUESTION IS ASKED ONCE FOR THE WHOLE BOARD AND JOINED. closure_act, remand, the unruled
-- docket filing, the successor whose mint closed the gap, the evidence-answered gaps and the retired
-- answers are each one small set keyed by gap id. Asked as correlated subqueries per gap instead,
-- closure_act and remand are each planned three times and the evidence and supersedes lookups scan
-- their tables once per gap.
CREATE VIEW "gap" AS
WITH "closure" AS MATERIALIZED (
  -- HOW THE GAP CLOSED, stated here and nowhere else: the last closing act of each arm by its
  -- events.id (NULL where the arm has none), and the first of each, which says which arm's act
  -- ended the gap.
  --
  -- A SUPERSESSION IS THE LENS'S OWN CLOSING ACT, so beside the bench it is read as red's: whose act
  -- ended the gap and whose word stands compare the bench with red's close and the superseding mint
  -- alike ("red_first", "red_last").
  --
  -- judged is the last closing act that is not a carry — a close with no carried_from, a closing
  -- ruling, or the superseding mint: a carry restates a closure the record holds, and each of the
  -- others judged the gap anew.
  SELECT ca."gap_id"                                                                AS "gap_id",
         MAX(CASE WHEN ca."arm" = 'close' THEN ca."event_id" END)                   AS "close_event",
         MAX(CASE WHEN ca."arm" = 'bench' THEN ca."event_id" END)                   AS "bench_event",
         MIN(CASE WHEN ca."arm" = 'supersede' THEN ca."event_id" END)               AS "supersede_event",
         MIN(CASE WHEN ca."arm" IN ('close', 'supersede') THEN ca."event_id" END)   AS "red_first",
         MAX(CASE WHEN ca."arm" IN ('close', 'supersede') THEN ca."event_id" END)   AS "red_last",
         MIN(CASE WHEN ca."arm" = 'bench' THEN ca."event_id" END)                   AS "first_bench",
         MAX(CASE WHEN ca."arm" IN ('bench', 'supersede') OR cl."carried_from" IS NULL
                  THEN ca."event_id" END)                                           AS "judged"
  FROM "closure_act" ca
  LEFT JOIN "close" cl ON cl."event_id" = ca."event_id" AND ca."arm" = 'close'
  GROUP BY ca."gap_id"
),
"latest_remand" AS MATERIALIZED (
  -- THE LATEST REMAND ON EACH GAP, off the "remand" view every reader of a remand reads.
  --
  -- HERE AN ORDER IS BOTH AVAILABLE AND MEANINGFUL. The events primary key is monotonic, unique,
  -- and nothing to do with the motion-id spelling that has no usable order. The LATEST remand is
  -- the live one — an earlier remand's direction has already had its exchange.
  SELECT "gap_id", "direction", "settled" FROM (
    SELECT "gap_id", "direction", "settled",
           row_number() OVER (PARTITION BY "gap_id" ORDER BY "event_id" DESC) AS "n"
    FROM "remand")
  WHERE "n" = 1
),
"unruled_docket" AS MATERIALIZED (
  -- THE DOCKET MOTION STANDING UNRULED, by its filing's events.id: the newest docket motion on the
  -- gap that no ruling names. ASKED PER MOTION, NEVER AS A COUNT. A count of ruling rows against a
  -- count of dockets reads two rulings on one motion as covering an unruled sibling (#1201).
  SELECT md."gap_id" AS "gap_id", max(md."event_id") AS "filed"
  FROM "motion_docket" md
  JOIN "motion" mo ON mo."event_id" = md."event_id"
  WHERE NOT EXISTS(SELECT 1 FROM "motion_rule" mr WHERE mr."motion_id" = mo."motion_id")
  GROUP BY md."gap_id"
),
"superseded" AS MATERIALIZED (
  -- THE SUCCESSOR WHOSE MINT CLOSED THE GAP: the gap the closure's 'supersede' arm names.
  SELECT c."gap_id" AS "gap_id", m2."gap_id" AS "by"
  FROM "closure" c
  JOIN "mint" m2 ON m2."event_id" = c."supersede_event"
),
"evidence" AS MATERIALIZED (
  -- THE EVIDENCE ANCHORS, read once for the board: every arm below joins this set.
  SELECT "anchor_id", "kind", "check_kind", "answers", "retired_event" FROM "evidence_anchor"
),
"reach" AS MATERIALIZED (
  -- THE ANCHORS EACH GAP'S DEMAND REACHES, and whose check it is (the demand_reach view), read once
  -- for the board.
  SELECT "gap_id", "from_gap", "check_kind", "minted", "anchor", "seat", "reached" FROM "demand_reach"
),
"latest_check" AS MATERIALIZED (
  -- EACH SEAT'S LATEST CHECK OF EACH EVIDENCE ANCHOR THAT READ IT, and whether it holds: the anchor
  -- ledger's replacement rule (a seat's check is replaced only by its own later check that reads)
  -- and its holds, one arm per recordsql.EvidenceTables row, generated from the row's check table,
  -- reads and holds.
  SELECT "anchor", "seat", "check_kind", "event_id", "holds" FROM (
    SELECT x."anchor" AS "anchor", ce."seat_id" AS "seat", 'source' AS "check_kind",
           x."event_id" AS "event_id", ((x."outcome" IN ('supports', 'supports_with_bridge', 'weak') AND COALESCE(x."source_completeness", '') <> 'not_the_work')) AS "holds",
           row_number() OVER (PARTITION BY x."anchor", ce."seat_id" ORDER BY x."event_id" DESC) AS "n"
    FROM "verify" x
    JOIN "events" ce ON ce."id" = x."event_id"
    WHERE (x."outcome" IS NOT NULL AND x."outcome" <> 'unreachable' AND COALESCE(x."source_completeness", '') <> 'not_the_work')
    UNION ALL
    SELECT x."computation_id" AS "anchor", ce."seat_id" AS "seat", 'computation' AS "check_kind",
           x."event_id" AS "event_id", (x."soundness" = 'sound') AS "holds",
           row_number() OVER (PARTITION BY x."computation_id", ce."seat_id" ORDER BY x."event_id" DESC) AS "n"
    FROM "reproduce" x
    JOIN "events" ce ON ce."id" = x."event_id"
    WHERE 1
  ) WHERE "n" = 1
),
"evidence_met" AS MATERIALIZED (
  -- THE LIVE-ANSWER DEMAND, MET (recordpb.DemandClause). Three arms, each off rows:
  --   1. a live evidence anchor of the gap's check kind was placed as its answer;
  --   2. an anchor the demand reaches (demand_reach) is retired: the anchor the gap is about, or
  --      whose reading it carries, or one an ancestor it carries the demand from was about or
  --      carried a reading of, retired since the successor's mint;
  --   3. the check again: for an anchor the demand reaches that is live and of the gap's check
  --      kind, that row's seat's latest check of it that read it is at or after the gap's mint and
  --      holds. Another seat's check meets nothing, and a check of one kind never meets the other
  --      kind's demand.
  SELECT ea."answers" AS "gap_id"
  FROM "evidence" ea
  JOIN "mint" m ON m."gap_id" = ea."answers" AND m."check_kind" = ea."check_kind"
  WHERE ea."retired_event" IS NULL
  UNION
  SELECT dr."gap_id"
  FROM "reach" dr
  JOIN "evidence" ea ON ea."anchor_id" = dr."anchor"
  WHERE ea."retired_event" IS NOT NULL AND (dr."from_gap" = dr."gap_id" OR ea."retired_event" > dr."reached")
  UNION
  SELECT dr."gap_id"
  FROM "reach" dr
  JOIN "evidence" ea ON ea."anchor_id" = dr."anchor"
  JOIN "latest_check" lc ON lc."anchor" = dr."anchor" AND lc."seat" = dr."seat" AND lc."check_kind" = dr."check_kind"
  WHERE ea."retired_event" IS NULL AND lc."event_id" >= dr."minted" AND lc."holds"
),
"answers_retired" AS MATERIALIZED (
  -- THE GAPS WHOSE ANSWERS ARE ALL RETIRED: at least one evidence anchor was placed as the gap's
  -- answer, and every one has a retire; retired is the last of them.
  SELECT "answers" AS "gap_id", MAX("retired_event") AS "retired"
  FROM "evidence"
  WHERE "answers" IS NOT NULL
  GROUP BY "answers"
  HAVING count(*) = count("retired_event")
)
SELECT
  m."gap_id"                                   AS "gap_id",
  m."class"                                    AS "class",
  m."class_material"                           AS "class_material",
  m."location"                                 AS "location",
  m."about_kind"                               AS "about_kind",
  m."about_ref"                                AS "about_ref",
  m."problem"                                  AS "problem",
  m."mint_reason"                              AS "mint_reason",
  m."required_fix"                             AS "required_fix",
  m."acceptance_check"                         AS "acceptance_check",
  m."check_kind"                               AS "check_kind",
  m."fix_basis"                                AS "fix_basis",
  m."fix_new"                                  AS "fix_new",
  m."severity"                                 AS "severity",
  m."likelihood"                               AS "likelihood",
  m."impact"                                   AS "impact",
  m."complexity_cost"                          AS "complexity_cost",
  e."id"                                       AS "minted_seq",
  e."epoch"                                    AS "minted_epoch",
  e."seat_id"                                  AS "minted_by",
  -- A reader takes a closing act's word, epoch and seat from the event the id names.
  --
  -- A gap can be closed more than once — red re-adjudicates across epochs (defect_accepted in one,
  -- repaired in a later one, or the chair carries an earlier closure forward), and the bench can
  -- rule many times, remanding in one epoch and disposing in the next. The LAST close and the LAST
  -- closing ruling are the words that stand. ONE ROW PER GAP, because closure is grouped by gap: a
  -- join to "close" fans out one row per close event, and a gap with two closes then counts twice
  -- in board_counts while the raw event walk counts it once (a projection disagreement the
  -- consistency oracle catches).
  c."close_event"                              AS "close_event",
  c."bench_event"                              AS "bench_close_event",
  -- WHO CLOSED IT is the arm of the FIRST closing act. A docket motion filed while its gap is open
  -- stands unruled after the minting lens closes the gap, and the bench rules it to clear the PASS:
  -- in universe-m25 the lenses closed G-e205e589 and G-a810bbab (ev440, ev460) before the bench
  -- ruled their k-max dockets (ev488, ev489), and read by its later act the run holds four lens
  -- closures counted as the bench's. The ruling is the bench's word on the gap; it is not what
  -- ended it.
  (c."first_bench" IS NOT NULL
     AND (c."red_first" IS NULL OR c."first_bench" < c."red_first"))                    AS "closed_by_bench",
  -- WHOSE WORD STANDS is the later arm's: a bench ruling after red's close overrules red's word, and
  -- the closure's epoch is that act's.
  (c."bench_event" IS NOT NULL
     AND (c."red_last" IS NULL OR c."bench_event" > c."red_last"))                      AS "bench_closed_last",
  -- The mint that closed the gap by superseding it, NULL where none did.
  c."supersede_event"                                                                   AS "supersede_event",
  (c."gap_id" IS NULL)                                                                  AS "open",
  -- THE REPEATED FIELDS, ANSWERED HERE RATHER THAN STORED. A gap's lineage and its credited
  -- findings live in child tables, so "does this gap supersede anything" is a join every reader
  -- would otherwise write for itself — and several did, differently.
  --
  -- Derived rather than denormalised, deliberately: a has_lineage column on mint, maintained by
  -- an insert trigger on the child, is a second copy of a fact the child table already holds, and
  -- it would have to be UPDATEd into a record that is append-only. The count is free here and
  -- cannot disagree with the rows it counts.
  (SELECT count(*) FROM "mint_supersedes" s WHERE s."event_id" = m."event_id") AS "supersedes_count",
  (SELECT count(*) FROM "mint_found_by"   f WHERE f."event_id" = m."event_id") AS "found_by_count",
  -- THE CURRENT GRADES, one per axis, off gap_grade's standing row. The plain "severity" columns
  -- above are MINT-TIME grades — a reader that wants what the gap is graded NOW and reaches for them
  -- silently reads a number a regrade may have moved, which is why the current grades are answered
  -- here rather than left to each reader.
  gg."severity"                                                                       AS "current_severity",
  gg."likelihood"                                                                     AS "current_likelihood",
  gg."impact"                                                                         AS "current_impact",
  gg."complexity_cost"                                                                AS "current_complexity_cost",
  -- THE LIVE-ANSWER DEMAND, answered where every asker can share it: a gap whose check kind an
  -- evidence anchor answers closes only once evidence_met holds it, and "awaiting evidence" is the
  -- debt list blue is handed. The demanded kinds are recordsql.EvidenceTables' check words; the
  -- Go home for which kind is owed (Gap.AnswerKind) reads the same rows, and THIS is the one SQL
  -- statement of the demand.
  (em."gap_id" IS NOT NULL)                                                           AS "evidence_answered",
  (c."gap_id" IS NULL
     AND m."check_kind" IN ('source', 'computation') AND em."gap_id" IS NULL)              AS "awaiting_evidence",
  -- THE ANSWERS TO A CLOSED GAP, ALL RETIRED: the last retire that left none of them live, while
  -- no closing act other than a carry follows it. The closure stands; the column is the gap's
  -- lens's news, and NULL on every open gap and on one judged again after the retire.
  (CASE WHEN c."gap_id" IS NOT NULL
             AND (c."judged" IS NULL OR c."judged" < ar."retired")
        THEN ar."retired" END)                                                        AS "answers_retired_event",
  -- The dispatch plan reads the unruled filing to ready the bench and to key the bench's sitting on
  -- the filing; remanded and unruled_docket_motion (outer select) read it, as the "nothing pending"
  -- arm and as the motion's id. The PASS gate asks the same per-motion question of the events
  -- (motionsAt), and
  -- TestTheDispatchPlanReadiesTheBenchForTheMotionsItListsWhileMotionsAreFiled (record's
  -- narrowedview_test.go) holds the two level.
  ud."filed"                                                                          AS "unruled_docket_filed",
  -- THE DIRECTION THE LATEST REMAND STATES, and the proposition it settles: a remand closes nothing,
  -- and may still bar a party from re-asserting what the bench decided on the way to sending the gap
  -- back.
  lr."direction"                                                                      AS "docket_direction",
  lr."settled"                                                                        AS "docket_settled",
  -- The successor whose mint closed this gap, NULL where no supersession closed it.
  sp."by"                                                                             AS "superseded_by",
  -- The mint's own event id, so board order is a sort key rather than a join every reader
  -- writes for itself.
  m."event_id"                                                                        AS "minted_event",
  gg."material"                                                                       AS "material",
  -- THE BENCH REMANDED IT AND NOTHING IS PENDING, and this is the column that lets a seat be told so.
  --
  -- A remand ruled at impasse sends the gap back to the debate for ONE more exchange between its
  -- minting lens and blue, on the direction the ruling states (docket_direction). Where a remanded
  -- gap stands against that exchange — owed or had — turns on the exchanges and the acts in them,
  -- so the record's Go fold decides it (exchangesOf, over the "remand" view's
  -- rows); this column says only that a remand stands on the gap and no docket motion on it stands
  -- unruled. Without it the chair was told only "gap G1 is open and material — PASS is refused while
  -- it is", which is true of a gap nobody has ever put before the bench and of one the bench has
  -- heard and sent back.
  --
  -- AND NOTHING PENDING, which is the arm that keeps this from double-counting. A gap docketed
  -- again and awaiting an answer is before the bench, not back in the debate. "Nothing pending" is
  -- unruled_docket_filed's own source, so that the two columns cannot answer "does a docket motion
  -- stand unruled" differently.
  (c."gap_id" IS NULL
     AND ud."filed" IS NULL AND lr."gap_id" IS NOT NULL)                              AS "remanded",
  -- THE MOTION ID OF THAT FILING, which the dispatch plan's reason names. Read off the same filing
  -- remanded reads: the filing and its id cannot then name two motions.
  um."motion_id"                                                                      AS "unruled_docket_motion"
FROM "mint" m
JOIN "events" e ON e."id" = m."event_id"
LEFT JOIN "gap_grade" gg ON gg."minted_event" = m."event_id" AND gg."until_seq" IS NULL
LEFT JOIN "closure" c ON c."gap_id" = m."gap_id"
LEFT JOIN "latest_remand" lr ON lr."gap_id" = m."gap_id"
LEFT JOIN "unruled_docket" ud ON ud."gap_id" = m."gap_id"
LEFT JOIN "superseded" sp ON sp."gap_id" = m."gap_id"
LEFT JOIN "evidence_met" em ON em."gap_id" = m."gap_id"
LEFT JOIN "answers_retired" ar ON ar."gap_id" = m."gap_id"
LEFT JOIN "motion" um ON um."event_id" = ud."filed";

-- THE CONVERGENCE RULE'S INPUTS, AS THE BOARD STOOD AT A POINT: one row per verdict, and one for the
-- record's newest event.
--
-- A verdict is judged on the board AS IT STOOD WHEN IT WAS ISSUED: the acts written at or before it,
-- and no other — one axis, events.id, for the regrades that grade a gap and the acts that close it
-- alike. The FAIL refusal at the write path reads the row at the newest event — the point a verdict
-- being written is judged at, where "as it stood" is "now".
--
-- The columns, each at the point:
--   mass                  gap_grade's mass summed over the gaps open at the point, on the grades that
--                         held there.
--   peak                  the largest mass at any point up to and including this one: the run's peak
--                         board as it stood then.
--   max_severity_mass     the top severity's mass among the open gaps; reported, decides nothing.
--   material_open         open gaps that are material on the grades that held at the point
--                         (gap_grade's rule — the same column the gap view reads now).
--   fresh_material_mints  gaps minted in the point's epoch at or before it, superseding nothing, and
--                         material at the point: new discovery, not a repair of known work.
--
-- "peak" is a running maximum over the points in seq order, framed by ROWS: seq is unique, so the
-- default RANGE frame selects the same rows and only costs the peer comparison.
--
-- THE FRACTION IS NOT HERE. It is a term of the run (inputs/run-config.json, record.Params), not a
-- fact in the record, so the predicate over these columns is one Go function every reader calls
-- (record.Convergence) rather than a constant written into this view.
CREATE VIEW "convergence_at" AS
WITH "point" AS MATERIALIZED (
  SELECT e."id" AS "seq", e."epoch" AS "epoch", v."verdict" AS "verdict"
  FROM "events" e
  LEFT JOIN "gate" v ON v."event_id" = e."id"
  WHERE e."type" = 'verdict' OR e."id" = (SELECT MAX("id") FROM "events")
),
"minted" AS MATERIALIZED (
  SELECT m."event_id" AS "minted_event", m."gap_id" AS "gap_id", e."epoch" AS "minted_epoch",
         NOT EXISTS(SELECT 1 FROM "mint_supersedes" s WHERE s."event_id" = m."event_id") AS "fresh"
  FROM "mint" m
  JOIN "events" e ON e."id" = m."event_id"
),
"closed" AS MATERIALIZED (
  SELECT x."gap_id" AS "gap_id", x."event_id" AS "seq"
  FROM "closure_act" x
),
"held" AS MATERIALIZED (
  SELECT "minted_event", "from_seq", "until_seq", "mass", "severity_mass", "material" FROM "gap_grade"
),
"board" AS (
  SELECT p."seq"                                    AS "seq",
         b."minted_epoch"                           AS "minted_epoch",
         b."fresh"                                  AS "fresh",
         NOT EXISTS(SELECT 1 FROM "closed" c
                     WHERE c."gap_id" = b."gap_id" AND c."seq" <= p."seq") AS "open",
         h."mass"                                   AS "mass",
         h."severity_mass"                          AS "severity_mass",
         h."material"                               AS "material"
  FROM "point" p
  JOIN "minted" b ON b."minted_event" <= p."seq"
  JOIN "held" h ON h."minted_event" = b."minted_event" AND h."from_seq" <= p."seq"
               AND (h."until_seq" IS NULL OR h."until_seq" > p."seq")
),
"at" AS (
  SELECT p."seq"                                                              AS "seq",
         p."epoch"                                                            AS "epoch",
         p."verdict"                                                          AS "verdict",
         COALESCE(SUM(CASE WHEN w."open" THEN w."mass" END), 0.0)             AS "mass",
         COALESCE(MAX(CASE WHEN w."open" THEN w."severity_mass" END), 0.0)    AS "max_severity_mass",
         COALESCE(SUM(w."open" AND w."material"), 0)                          AS "material_open",
         COALESCE(SUM(w."minted_epoch" = p."epoch" AND w."fresh"
                      AND w."material"), 0)                                   AS "fresh_material_mints"
  FROM "point" p
  LEFT JOIN "board" w ON w."seq" = p."seq"
  GROUP BY p."seq"
)
SELECT a."seq"                                    AS "seq",
       a."epoch"                                  AS "epoch",
       a."verdict"                                AS "verdict",
       a."mass"                                   AS "mass",
       MAX(a."mass") OVER (ORDER BY a."seq" ROWS UNBOUNDED PRECEDING) AS "peak",
       a."max_severity_mass"                      AS "max_severity_mass",
       a."material_open"                          AS "material_open",
       a."fresh_material_mints"                   AS "fresh_material_mints"
FROM "at" a;

-- The board's own count, asked once. Every consumer that wants "how many gaps are open" reads this
-- rather than folding the stream again with its own idea of what closed means.
--
-- ONE PASS OVER THE GAP VIEW: both counts are sums over the same rows, and the view is the costly
-- part of the question.
CREATE VIEW "board_counts" AS
SELECT
  COALESCE(SUM("open"), 0)          AS "open_gaps",
  COALESCE(SUM(NOT "open"), 0)      AS "closed_gaps",
  (SELECT count(*) FROM "events")   AS "events"
FROM "gap";

-- THE ANSWERS TO A MOTION, one row per answered id, whichever filing shape asked. The FIRST
-- ruling is the one that counts, and a second is refused at the write. The view states the first-wins rule ONCE, so a row inserted past the write cannot
-- multiply anybody's join. Keyed on the id alone rather than joined to 'motion',
-- because a position motion has no filing row (the chair's position IS the
-- filing) and its answers must be askable all the same.
CREATE VIEW "motion_answers" AS
SELECT
  ids."motion_id"                                        AS "motion_id",
  fr."grade"                                             AS "grade",
  fr."petition"                                          AS "petition",
  -- THE DOCKET ARM IS A TABLE, NOT A COLUMN, and that is why it is joined rather than read.
  -- Its two siblings are enums and land as columns on "motion_rule"; the bench's is a MESSAGE
  -- (it carries the principle, the tension and what would reopen it), so its disposition lives
  -- one table down. Left out of the COALESCE below, "ruling" is NULL for every bench ruling ever
  -- made and RequireUnruledMotion reads the whole docket as unanswered.
  rd."disposition"                                       AS "docket",
  COALESCE(fr."grade", fr."petition", rd."disposition") AS "ruling",
  fre."seat_id"                                          AS "ruled_by",
  fre."id"                                            AS "ruled_seq"
FROM (SELECT DISTINCT "motion_id" FROM "motion_rule") ids
LEFT JOIN "motion_rule" fr ON fr."event_id" =
  (SELECT MIN(x."event_id") FROM "motion_rule" x WHERE x."motion_id" = ids."motion_id")
LEFT JOIN "motion_rule_docket" rd ON rd."event_id" = fr."event_id"
LEFT JOIN "events" fre ON fre."id" = fr."event_id";

-- AN INQUIRY, whole: proposed by whom, saying what, where its status stands now, and its latest
-- check. The proposal row carries the substance; the LATEST inquiry event carries the status; the
-- LATEST inquiry_check carries the check. Whether that check STANDS is not a column: it stands while
-- it is later than the status and than the latest gap about the inquiry that was material at its
-- mint, which the Go fold decides (record/inquirycheck.go) and InquiriesOf reads off the same rows.
CREATE VIEW "inquiry_state" AS
SELECT
  p."inquiry_id"                       AS "inquiry_id",
  pe."seat_id"                         AS "proposed_by",
  pe."id"                           AS "proposed_seq",
  fp."question"                        AS "question",
  ls."status"                          AS "status",
  lse."id"                          AS "status_seq",
  ce."id"                           AS "check_seq",
  ce."seat_id"                         AS "checked_by"
FROM (SELECT a0."inquiry_id" AS "inquiry_id", MIN(a0."event_id") AS "pid"
        FROM "inquiry" a0
        WHERE COALESCE(a0."inquiry_id", '') != '' AND a0."supersedes_status" IS NULL
        GROUP BY a0."inquiry_id") p
JOIN "inquiry" fp ON fp."event_id" = p."pid"
JOIN "events" pe ON pe."id" = p."pid"
LEFT JOIN "inquiry" ls ON ls."event_id" =
  (SELECT MAX(x."event_id") FROM "inquiry" x WHERE x."inquiry_id" = p."inquiry_id")
LEFT JOIN "events" lse ON lse."id" = ls."event_id"
LEFT JOIN "events" ce ON ce."id" =
  (SELECT MAX(c."event_id") FROM "inquiry_check" c WHERE c."inquiry_id" = p."inquiry_id");

-- report_op is the ORDERED STREAM OF TEXT MUTATIONS that reconstruct blue's report. The report is
-- this stream folded from the empty text, so the SELECTION of which events mutate the text — and
-- how each names its change — is a projection, authored HERE where every reader sees the same one,
-- not re-derived by walking every event in Go. Each row is (event_id, kind, a, b, exact):
--   append → a=a lane's part, as it ingested it; exact=0. On the empty text the part is the text;
--            otherwise the text, its trailing newlines trimmed, a blank line, then the part.
--   edit   → a=old span, b=new span (blue edit's splice, located and replaced at replay);
--            exact=1 when the edit recorded exact_span, so replay locates a AS WRITTEN.
--   insert → a=the anchoring quote, b=the marker id (Token(b) is spliced at that quote); exact=0.
--            One row per anchor event placed by its quote: every act that places a marker — a
--            finding, a gap, a citation, a computation — appends one after its own
--            event, so where an id is placed is read from one table.
--   sentence_of → a=the id of an anchor already in the text, b=the marker id (Token(b) is spliced
--            at the end of the sentence Token(a) sits in, found by that token, never by text);
--            exact=0. One row per anchor event placed by sentence_of — a reading's gap.
--   remove → a=an anchor id a retire took out with its claim (Token(a) and the husk it leaves
--            are removed). One row per named anchor; exact=0.
-- The FOLD itself stays in Go: replaying a splice needs the running text a prior op left, which
-- SQL cannot carry — but nothing here builds state by scanning the whole event log.
--
-- Each arm reads its body table alone: every body row's event_id references events(id) and the
-- record opens with foreign keys enforced, so a join back to events filters nothing.
CREATE VIEW "report_op" AS
  SELECT p."event_id" AS "event_id", 'append' AS "kind", p."text" AS "a", NULL AS "b", 0 AS "exact"
    FROM "report_part" p
  UNION ALL
  SELECT b."event_id", 'edit', b."old", b."new", COALESCE(b."exact_span", 0)
    FROM "blue_edit" b
  UNION ALL
  SELECT a."event_id", 'insert', a."location", a."id", 0
    FROM "anchor" a
    WHERE COALESCE(a."location", '') != '' AND COALESCE(a."id", '') != ''
  UNION ALL
  SELECT a."event_id", 'sentence_of', a."sentence_of", a."id", 0
    FROM "anchor" a
    WHERE COALESCE(a."sentence_of", '') != '' AND COALESCE(a."id", '') != ''
  UNION ALL
  SELECT ra."event_id", 'remove', ra."value", NULL, 0
    FROM "retire_anchors" ra;
