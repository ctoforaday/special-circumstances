
CREATE TABLE "events" (
  "id"      INTEGER PRIMARY KEY,
  "seat_id" TEXT    NOT NULL,
  "round"   INTEGER NOT NULL,
  "ts"      TEXT    NOT NULL,
  "type"    TEXT    NOT NULL REFERENCES "enum_event_type"("value"),
  -- The key is the fact that has to be unique, and the partial index below enforces it globally.
  -- This was UNIQUE (seat_id, nonce, seq): a counter nothing read, scoped by a sitting that no
  -- longer exists. Both are gone.
  "key"     TEXT
) STRICT;

CREATE UNIQUE INDEX "events_key" ON "events" ("key") WHERE "key" IS NOT NULL;
CREATE INDEX "events_type" ON "events" ("type");
CREATE INDEX "events_round" ON "events" ("round");

CREATE TRIGGER "events_are_append_only_update" BEFORE UPDATE ON "events" BEGIN
  SELECT RAISE(ABORT, 'the record is append-only: an event cannot be edited after it is written');
END;
CREATE TRIGGER "events_are_append_only_delete" BEFORE DELETE ON "events" BEGIN
  SELECT RAISE(ABORT, 'the record is append-only: an event cannot be removed after it is written');
END;

CREATE TABLE "enum_event_type" (
  "value" TEXT PRIMARY KEY,
  "means" TEXT NOT NULL
) STRICT;
INSERT INTO "enum_event_type" ("value", "means") VALUES ('anchor', 'evidence tied to a finding: where in the artifact the claim actually lives');
INSERT INTO "enum_event_type" ("value", "means") VALUES ('avenue', 'a line of inquiry, from proposed through pursued, declined, deferred or abandoned');
INSERT INTO "enum_event_type" ("value", "means") VALUES ('blue_edit', 'a change to the living report, recorded as old and new so the edit itself is auditable');
INSERT INTO "enum_event_type" ("value", "means") VALUES ('certify', 'a seat''s signed statement about its own work — what it asserts on the record');
INSERT INTO "enum_event_type" ("value", "means") VALUES ('cite', 'a source brought into the debate, with the hash and access date that make it re-checkable');
INSERT INTO "enum_event_type" ("value", "means") VALUES ('class_new', 'a defect class coined in this run, with its definition and the neighbour it is distinguished from');
INSERT INTO "enum_event_type" ("value", "means") VALUES ('close', 'a merge closing a gap on a verified repair — red''s half of the closing vocabulary');
INSERT INTO "enum_event_type" ("value", "means") VALUES ('closing', 'a seat''s closing statement on a gap: the argument, not the disposition');
INSERT INTO "enum_event_type" ("value", "means") VALUES ('declare', 'the bench stating a holding that later sittings are expected to apply');
INSERT INTO "enum_event_type" ("value", "means") VALUES ('finding', 'something red found, graded but not yet minted as a gap');
INSERT INTO "enum_event_type" ("value", "means") VALUES ('friction', 'a capability the tool did not have, recorded so the tooling gets fixed rather than worked around');
INSERT INTO "enum_event_type" ("value", "means") VALUES ('friction_none', 'a seat stating it hit no friction — the negative answer, recorded so silence and `none` are different facts');
INSERT INTO "enum_event_type" ("value", "means") VALUES ('halt', 'the bench ending the run on a safety, ethics, consent or integrity boundary');
INSERT INTO "enum_event_type" ("value", "means") VALUES ('inquiry_review', 'a review of the lines of inquiry themselves, rather than of a finding');
INSERT INTO "enum_event_type" ("value", "means") VALUES ('manifest_row', 'one row of the run''s manifest, tying a gap to what shipped for it');
INSERT INTO "enum_event_type" ("value", "means") VALUES ('mint', 'a gap put on the board — the act that creates the entity every other act refers to');
INSERT INTO "enum_event_type" ("value", "means") VALUES ('motion', 'a motion filed: a grade contested, a petition to the bench, or a direction proposed');
INSERT INTO "enum_event_type" ("value", "means") VALUES ('motion_appeal', 'an appeal of a ruling already made on a motion');
INSERT INTO "enum_event_type" ("value", "means") VALUES ('motion_rule', 'the bench''s ruling on a filed motion, and whom it binds');
INSERT INTO "enum_event_type" ("value", "means") VALUES ('observe', 'an observation recorded without a claim attached to it');
INSERT INTO "enum_event_type" ("value", "means") VALUES ('opinion', 'the bench ruling on a gap, with the principle applied and the tension acknowledged');
INSERT INTO "enum_event_type" ("value", "means") VALUES ('outcome', 'the run''s terminal act: how it ended and whether the question was answered');
INSERT INTO "enum_event_type" ("value", "means") VALUES ('position', 'a seat''s stated position going into a round');
INSERT INTO "enum_event_type" ("value", "means") VALUES ('proof', 'a script that was RUN, with its hash and exit status — the answer a computation check demands');
INSERT INTO "enum_event_type" ("value", "means") VALUES ('register', 'a seat took its seat — the first act of any seat, stamping the tool version it ran under');
INSERT INTO "enum_event_type" ("value", "means") VALUES ('regrade', 'a gap''s grade changed, with the basis for the change');
INSERT INTO "enum_event_type" ("value", "means") VALUES ('reproduce', 'an attempt to re-run a recorded proof, and whether what it computes is sound');
INSERT INTO "enum_event_type" ("value", "means") VALUES ('retire', 'a claim withdrawn from the report, with the reason and what supersedes it');
INSERT INTO "enum_event_type" ("value", "means") VALUES ('revision', 'a revision to a seat''s own earlier text');
INSERT INTO "enum_event_type" ("value", "means") VALUES ('spot_check', 'red re-checking a sample of prior work, or stating that it checked none and why');
INSERT INTO "enum_event_type" ("value", "means") VALUES ('verdict', 'red''s round gate: PASS or FAIL against the open board');
INSERT INTO "enum_event_type" ("value", "means") VALUES ('verify', 'a citation checked at the leaf: what the source did for the claim, and how sure the reader is');

CREATE TABLE "enum_schema_version" (
  "value" TEXT PRIMARY KEY,
  "means" TEXT NOT NULL
) STRICT;
INSERT INTO "enum_schema_version" ("value", "means") VALUES ('1', 'the protobuf record: one event stream, one row per act, schema derived from these descriptors');

CREATE TABLE "enum_verdict" (
  "value" TEXT PRIMARY KEY,
  "means" TEXT NOT NULL
) STRICT;
INSERT INTO "enum_verdict" ("value", "means") VALUES ('fail', 'at least one gap is still open, or you are not satisfied it was answered');
INSERT INTO "enum_verdict" ("value", "means") VALUES ('pass', 'every gap on the board is resolved — this is CHECKED against the open board, not taken on your word');

CREATE TABLE "enum_run_outcome" (
  "value" TEXT PRIMARY KEY,
  "means" TEXT NOT NULL
) STRICT;
INSERT INTO "enum_run_outcome" ("value", "means") VALUES ('ceiling', 'the round ceiling was reached with work still open — NOT a judged failure to verify, and the stamp says so');
INSERT INTO "enum_run_outcome" ("value", "means") VALUES ('halted', 'the bench ended the run on a safety, ethics, consent or integrity boundary');
INSERT INTO "enum_run_outcome" ("value", "means") VALUES ('unverified', 'the run ended without the question being answered, and no ceiling or halt explains it');
INSERT INTO "enum_run_outcome" ("value", "means") VALUES ('verified', 'red passed the board and the bench agrees the question was answered');

CREATE TABLE "enum_motion_subject" (
  "value" TEXT PRIMARY KEY,
  "means" TEXT NOT NULL
) STRICT;
INSERT INTO "enum_motion_subject" ("value", "means") VALUES ('direction', 'a ruling on a line of inquiry blue proposed; the id is the AVENUE''s own, because the proposal IS the filing');
INSERT INTO "enum_motion_subject" ("value", "means") VALUES ('grade', 'you contest a gap''s grade on one dimension');
INSERT INTO "enum_motion_subject" ("value", "means") VALUES ('petition', 'you ask the bench to intervene — the constitutional short-circuit available to any party seat');

CREATE TABLE "enum_grade_dimension" (
  "value" TEXT PRIMARY KEY,
  "means" TEXT NOT NULL
) STRICT;
INSERT INTO "enum_grade_dimension" ("value", "means") VALUES ('complexity', 'what the fix costs; it is what makes risk_accepted arguable');
INSERT INTO "enum_grade_dimension" ("value", "means") VALUES ('impact', 'how far the damage reaches');
INSERT INTO "enum_grade_dimension" ("value", "means") VALUES ('likelihood', 'how likely the CONSEQUENCE is — not how sure you are the defect exists, which is a separate axis');
INSERT INTO "enum_grade_dimension" ("value", "means") VALUES ('severity', 'how bad it is if it bites');

CREATE TABLE "enum_grade" (
  "value" TEXT PRIMARY KEY,
  "means" TEXT NOT NULL
) STRICT;
INSERT INTO "enum_grade" ("value", "means") VALUES ('certain', 'the top of the scale — for LIKELIHOOD, reserve it for a consequence that is itself certain, never for a defect you merely verified exists');
INSERT INTO "enum_grade" ("value", "means") VALUES ('high', 'serious');
INSERT INTO "enum_grade" ("value", "means") VALUES ('low', 'minor');
INSERT INTO "enum_grade" ("value", "means") VALUES ('low_medium', 'between minor and material');
INSERT INTO "enum_grade" ("value", "means") VALUES ('medium', 'material');
INSERT INTO "enum_grade" ("value", "means") VALUES ('medium_high', 'between material and serious');
INSERT INTO "enum_grade" ("value", "means") VALUES ('realized', 'it has already happened. Contributes ZERO mass by design: mass forecasts what is still to come, and a realized defect is measured by its damage instead');
INSERT INTO "enum_grade" ("value", "means") VALUES ('trivial', 'cosmetic; nothing downstream changes if it is wrong');

CREATE TABLE "enum_petition_class" (
  "value" TEXT PRIMARY KEY,
  "means" TEXT NOT NULL
) STRICT;
INSERT INTO "enum_petition_class" ("value", "means") VALUES ('integrity', 'the record or the process has been compromised');
INSERT INTO "enum_petition_class" ("value", "means") VALUES ('process', 'the mechanics are obstructing the work');
INSERT INTO "enum_petition_class" ("value", "means") VALUES ('safety', 'a safety, ethics or consent boundary is in question');
INSERT INTO "enum_petition_class" ("value", "means") VALUES ('scope', 'the question being answered has drifted from the one asked');

CREATE TABLE "enum_grade_ruling" (
  "value" TEXT PRIMARY KEY,
  "means" TEXT NOT NULL
) STRICT;
INSERT INTO "enum_grade_ruling" ("value", "means") VALUES ('accepted', 'the proposed grade stands');
INSERT INTO "enum_grade_ruling" ("value", "means") VALUES ('rejected', 'the grade on the board stands');

CREATE TABLE "enum_petition_ruling" (
  "value" TEXT PRIMARY KEY,
  "means" TEXT NOT NULL
) STRICT;
INSERT INTO "enum_petition_ruling" ("value", "means") VALUES ('denied', 'the petition fails; the run continues as it was');
INSERT INTO "enum_petition_ruling" ("value", "means") VALUES ('granted', 'the relief asked for is ordered');

CREATE TABLE "enum_direction_ruling" (
  "value" TEXT PRIMARY KEY,
  "means" TEXT NOT NULL
) STRICT;
INSERT INTO "enum_direction_ruling" ("value", "means") VALUES ('endorsed', 'worth this run''s time — pursue it');
INSERT INTO "enum_direction_ruling" ("value", "means") VALUES ('out_of_scope', 'a real question, but not THIS question');
INSERT INTO "enum_direction_ruling" ("value", "means") VALUES ('too_thin', 'in scope, but the hypothesis does not carry its budget');

CREATE TABLE "enum_ruling_binds" (
  "value" TEXT PRIMARY KEY,
  "means" TEXT NOT NULL
) STRICT;
INSERT INTO "enum_ruling_binds" ("value", "means") VALUES ('all', 'every seat from here is bound by this ruling');
INSERT INTO "enum_ruling_binds" ("value", "means") VALUES ('filer', 'only the filing seat is bound');
INSERT INTO "enum_ruling_binds" ("value", "means") VALUES ('none', 'advisory — the ruling is on the record and obliges nobody');

CREATE TABLE "enum_check_kind" (
  "value" TEXT PRIMARY KEY,
  "means" TEXT NOT NULL
) STRICT;
INSERT INTO "enum_check_kind" ("value", "means") VALUES ('computation', 'RUNNING something settles it. This check CANNOT be closed by prose: it closes only when a proof answers the gap. Reach for it wherever the answer would be PRODUCED rather than asserted — arithmetic, a simulation, a forecast, a parse, a count, a re-derivation are common cases and not the whole of it; if you can imagine a script that would end the argument, this is the kind');
INSERT INTO "enum_check_kind" ("value", "means") VALUES ('document', 'reading a shipped artifact settles it — the check is answered by prose that quotes what is there');
INSERT INTO "enum_check_kind" ("value", "means") VALUES ('source', 'verifying an external source settles it — the claim stands or falls on what the cited material actually says');

CREATE TABLE "enum_disposition" (
  "value" TEXT PRIMARY KEY,
  "means" TEXT NOT NULL,
  "closes" INTEGER NOT NULL CHECK ("closes" IN (0, 1))
) STRICT;
INSERT INTO "enum_disposition" ("value", "means", "closes") VALUES ('amends_prior', 'a defect found BETWEEN two repairs that each closed clean earlier — REQUIRES supersedes so the lineage is explicit', 1);
INSERT INTO "enum_disposition" ("value", "means", "closes") VALUES ('carried', 'NOT a closure: the gap survives to the next round with a stated research direction the coming seat owes', 0);
INSERT INTO "enum_disposition" ("value", "means", "closes") VALUES ('closed', 'the repair was verified at the leaf and nothing regressed', 1);
INSERT INTO "enum_disposition" ("value", "means", "closes") VALUES ('closed_with_regression', 'repaired, but something else broke — REQUIRES a successor naming the gap that carries the regression forward', 1);
INSERT INTO "enum_disposition" ("value", "means", "closes") VALUES ('rebuttal_sustained', 'blue argued the finding was wrong and the argument held; nothing was repaired because nothing needed to be', 1);
INSERT INTO "enum_disposition" ("value", "means", "closes") VALUES ('risk_accepted', 'the fix costs more than the defect (complexity above likelihood x impact) and the risk is taken KNOWINGLY, with the argument on the record', 1);
INSERT INTO "enum_disposition" ("value", "means", "closes") VALUES ('routed_to_infrastructure', 'a real defect whose fix is owned outside this debate; it leaves here and is not silently dropped', 1);

CREATE TABLE "enum_source_outcome" (
  "value" TEXT PRIMARY KEY,
  "means" TEXT NOT NULL
) STRICT;
INSERT INTO "enum_source_outcome" ("value", "means") VALUES ('absent', 'you read the source and the claim is simply not in it. Distinct from `refutes`: silence is not contradiction, and a reader deciding what to do about it needs to know which it was');
INSERT INTO "enum_source_outcome" ("value", "means") VALUES ('refutes', 'you read the source and it CONTRADICTS the claim — the strongest finding this verb can carry, and until 0.60.0 it had no field at all');
INSERT INTO "enum_source_outcome" ("value", "means") VALUES ('supports', 'you read the source at the leaf and it says what the claim says');
INSERT INTO "enum_source_outcome" ("value", "means") VALUES ('supports_with_bridge', 'it supports the claim but you had to bridge something — a summary, a secondary citation, a near-restatement');
INSERT INTO "enum_source_outcome" ("value", "means") VALUES ('unreachable', 'you could not read it — paywall, dead link, a format you could not extract. Say what you tried in --reason; an untried "unable to corroborate" is an incomplete audit');
INSERT INTO "enum_source_outcome" ("value", "means") VALUES ('weak', 'it gestures at the claim, or is itself uncorroborated: thin support, not none');

CREATE TABLE "enum_confidence" (
  "value" TEXT PRIMARY KEY,
  "means" TEXT NOT NULL
) STRICT;
INSERT INTO "enum_confidence" ("value", "means") VALUES ('high', 'you read the source at the leaf and would defend this determination as it stands');
INSERT INTO "enum_confidence" ("value", "means") VALUES ('low', 'your reading may be wrong: an ambiguous passage, thin evidence, or a source you could only partly read. This is a call for more evidence, NOT an automatic fail — blue digs further');
INSERT INTO "enum_confidence" ("value", "means") VALUES ('medium', 'you are reasonably sure, but the reading bridges something — a summary, a secondary source, a near-restatement rather than the exact statement');

CREATE TABLE "enum_soundness" (
  "value" TEXT PRIMARY KEY,
  "means" TEXT NOT NULL
) STRICT;
INSERT INTO "enum_soundness" ("value", "means") VALUES ('sound', 'you READ the script and it computes what it claims to compute');
INSERT INTO "enum_soundness" ("value", "means") VALUES ('unsound', 'it re-runs cleanly and establishes nothing, or something other than the claim it is anchored to — the dangerous cell, because it looks maximally credible');

CREATE TABLE "enum_avenue_status" (
  "value" TEXT PRIMARY KEY,
  "means" TEXT NOT NULL
) STRICT;
INSERT INTO "enum_avenue_status" ("value", "means") VALUES ('abandoned', 'you started and stopped. REQUIRES a reason — what killed it is the part a future run actually needs');
INSERT INTO "enum_avenue_status" ("value", "means") VALUES ('declined', 'you considered it and chose not to. REQUIRES a reason — the road not taken is worthless without why');
INSERT INTO "enum_avenue_status" ("value", "means") VALUES ('deferred', 'not this run. REQUIRES a reason saying what a later run should pick it up FOR: a deferral with no stated reason is indistinguishable from forgetting, and this status exists precisely to be read by a run that has not happened yet');
INSERT INTO "enum_avenue_status" ("value", "means") VALUES ('proposed', 'you intend to follow this line; the tool assigns it an id and red may rule on it');
INSERT INTO "enum_avenue_status" ("value", "means") VALUES ('pursued', 'you took the line — what it produced belongs in the report');

CREATE TABLE "enum_friction_kind" (
  "value" TEXT PRIMARY KEY,
  "means" TEXT NOT NULL
) STRICT;
INSERT INTO "enum_friction_kind" ("value", "means") VALUES ('estoppel', 'the TOOL refused a mint because the defect lives in text blue applied verbatim from red''s own --fix-new. Recorded by the tool, not filed by the seat: argue it on the original gap, or mint with --supersedes so the lineage is explicit');
INSERT INTO "enum_friction_kind" ("value", "means") VALUES ('tool_error', 'the TOOL failed internally — unparseable input, an undecodable row, a check that could not run. Recorded rather than printed or swallowed, because an error nobody learns about is one nothing improves on. Distinct from a seat''s own friction so the counts an operator reads stay about capability gaps');

CREATE TABLE "register" (
  "event_id" INTEGER PRIMARY KEY REFERENCES "events"("id"),
  "tool_version" TEXT
) STRICT;

CREATE TABLE "round_verdict" (
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
  "ended" TEXT,
  FOREIGN KEY ("verdict") REFERENCES "enum_run_outcome"("value")
) STRICT;

CREATE TABLE "position" (
  "event_id" INTEGER PRIMARY KEY REFERENCES "events"("id"),
  "text" TEXT
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

CREATE TABLE "motion" (
  "event_id" INTEGER PRIMARY KEY REFERENCES "events"("id"),
  "motion_id" TEXT NOT NULL UNIQUE,
  "subject" TEXT,
  "basis" TEXT,
  "relief" TEXT,
  "filing_case" TEXT,
  FOREIGN KEY ("subject") REFERENCES "enum_motion_subject"("value")
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

CREATE TABLE "motion_direction" (
  "event_id" INTEGER PRIMARY KEY REFERENCES "motion"("event_id"),
  "avenue_id" TEXT
) STRICT;

CREATE TABLE "motion_rule" (
  "event_id" INTEGER PRIMARY KEY REFERENCES "events"("id"),
  "motion_id" TEXT NOT NULL,
  "subject" TEXT,
  "opinion" TEXT,
  "binds" TEXT,
  "grade" TEXT,
  "petition" TEXT,
  "direction" TEXT,
  CHECK (("grade" IS NOT NULL) + ("petition" IS NOT NULL) + ("direction" IS NOT NULL) <= 1),
  FOREIGN KEY ("motion_id") REFERENCES "motion"("motion_id"),
  FOREIGN KEY ("subject") REFERENCES "enum_motion_subject"("value"),
  FOREIGN KEY ("binds") REFERENCES "enum_ruling_binds"("value"),
  FOREIGN KEY ("grade") REFERENCES "enum_grade_ruling"("value"),
  FOREIGN KEY ("petition") REFERENCES "enum_petition_ruling"("value"),
  FOREIGN KEY ("direction") REFERENCES "enum_direction_ruling"("value")
) STRICT;

CREATE TABLE "motion_appeal" (
  "event_id" INTEGER PRIMARY KEY REFERENCES "events"("id"),
  "motion_id" TEXT,
  "subject" TEXT,
  "reason" TEXT,
  FOREIGN KEY ("motion_id") REFERENCES "motion"("motion_id"),
  FOREIGN KEY ("subject") REFERENCES "enum_motion_subject"("value")
) STRICT;

CREATE TABLE "mint" (
  "event_id" INTEGER PRIMARY KEY REFERENCES "events"("id"),
  "gap_id" TEXT NOT NULL UNIQUE,
  "mint_key" TEXT,
  "class" TEXT NOT NULL,
  "class_new" INTEGER,
  "definition" TEXT,
  "neighbor" TEXT,
  "distinguisher" TEXT,
  "location" TEXT,
  "problem" TEXT NOT NULL,
  "required_fix" TEXT,
  "fix_new" TEXT,
  "fix_basis" TEXT,
  "acceptance_check" TEXT NOT NULL,
  "check_kind" TEXT NOT NULL,
  "severity" TEXT,
  "likelihood" TEXT NOT NULL,
  "impact" TEXT NOT NULL,
  "complexity_cost" TEXT,
  "mint_reason" TEXT,
  CHECK ("class_new" IS NULL OR "class_new" IN (0, 1)),
  FOREIGN KEY ("check_kind") REFERENCES "enum_check_kind"("value"),
  FOREIGN KEY ("severity") REFERENCES "enum_grade"("value"),
  FOREIGN KEY ("likelihood") REFERENCES "enum_grade"("value"),
  FOREIGN KEY ("impact") REFERENCES "enum_grade"("value"),
  FOREIGN KEY ("complexity_cost") REFERENCES "enum_grade"("value")
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

CREATE TABLE "class_new" (
  "event_id" INTEGER PRIMARY KEY REFERENCES "events"("id"),
  "slug" TEXT,
  "definition" TEXT,
  "neighbor" TEXT,
  "distinguisher" TEXT
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
  "prose" TEXT NOT NULL,
  CHECK ("closure_class" IS NULL OR "closure_class" IN ('amends_prior', 'closed', 'closed_with_regression', 'rebuttal_sustained', 'risk_accepted', 'routed_to_infrastructure')),
  CHECK ("closure_class" <> 'closed_with_regression' OR "successor" IS NOT NULL),
  FOREIGN KEY ("gap_id") REFERENCES "mint"("gap_id"),
  FOREIGN KEY ("closure_class") REFERENCES "enum_disposition"("value"),
  FOREIGN KEY ("successor") REFERENCES "mint"("gap_id")
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
  FOREIGN KEY ("gap_id") REFERENCES "mint"("gap_id"),
  FOREIGN KEY ("severity") REFERENCES "enum_grade"("value"),
  FOREIGN KEY ("likelihood") REFERENCES "enum_grade"("value"),
  FOREIGN KEY ("impact") REFERENCES "enum_grade"("value"),
  FOREIGN KEY ("complexity_cost") REFERENCES "enum_grade"("value")
) STRICT;

CREATE TABLE "spot_check" (
  "event_id" INTEGER PRIMARY KEY REFERENCES "events"("id"),
  "none" INTEGER,
  "reason" TEXT,
  CHECK ("none" IS NULL OR "none" IN (0, 1))
) STRICT;

CREATE TABLE "spot_check_ids" (
  "event_id" INTEGER NOT NULL REFERENCES "spot_check"("event_id"),
  "ord"      INTEGER NOT NULL,
  "value"    TEXT    NOT NULL,
  PRIMARY KEY ("event_id", "ord")
) STRICT;

CREATE TABLE "opinion" (
  "event_id" INTEGER PRIMARY KEY REFERENCES "events"("id"),
  "gap_id" TEXT NOT NULL,
  "disposition" TEXT NOT NULL,
  "principle" TEXT NOT NULL,
  "tension" TEXT NOT NULL,
  "review_flag" TEXT NOT NULL,
  "rationale" TEXT NOT NULL,
  FOREIGN KEY ("gap_id") REFERENCES "mint"("gap_id"),
  FOREIGN KEY ("disposition") REFERENCES "enum_disposition"("value")
) STRICT;

CREATE TABLE "finding" (
  "event_id" INTEGER PRIMARY KEY REFERENCES "events"("id"),
  "finding_id" TEXT,
  "finding_key" TEXT,
  "label" TEXT,
  "location" TEXT,
  "text" TEXT,
  "severity" TEXT,
  "likelihood" TEXT,
  "impact" TEXT,
  FOREIGN KEY ("severity") REFERENCES "enum_grade"("value"),
  FOREIGN KEY ("likelihood") REFERENCES "enum_grade"("value"),
  FOREIGN KEY ("impact") REFERENCES "enum_grade"("value")
) STRICT;

CREATE TABLE "observe" (
  "event_id" INTEGER PRIMARY KEY REFERENCES "events"("id"),
  "label" TEXT,
  "text" TEXT,
  "observation" TEXT
) STRICT;

CREATE TABLE "anchor" (
  "event_id" INTEGER PRIMARY KEY REFERENCES "events"("id"),
  "id" TEXT,
  "location" TEXT,
  "finding_id" TEXT,
  "finding_key" TEXT,
  "label" TEXT,
  "text" TEXT
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
  "text" TEXT
) STRICT;

CREATE TABLE "verify" (
  "event_id" INTEGER PRIMARY KEY REFERENCES "events"("id"),
  "claim" TEXT NOT NULL,
  "url" TEXT,
  "title" TEXT,
  "anchor" TEXT,
  "independent" INTEGER,
  "access_date" TEXT,
  "outcome" TEXT NOT NULL,
  "confidence" TEXT NOT NULL,
  "text" TEXT NOT NULL,
  CHECK ("independent" IS NULL OR "independent" IN (0, 1)),
  FOREIGN KEY ("outcome") REFERENCES "enum_source_outcome"("value"),
  FOREIGN KEY ("confidence") REFERENCES "enum_confidence"("value")
) STRICT;

CREATE TABLE "proof" (
  "event_id" INTEGER PRIMARY KEY REFERENCES "events"("id"),
  "proof_id" TEXT,
  "proof_key" TEXT,
  "proof_sha" TEXT,
  "proof_basis" TEXT,
  "answers" TEXT,
  "cites" TEXT,
  "drift" TEXT,
  "text" TEXT,
  "script" TEXT,
  "exit" INTEGER
) STRICT;

CREATE TABLE "reproduce" (
  "event_id" INTEGER PRIMARY KEY REFERENCES "events"("id"),
  "proof_sha" TEXT,
  "reproduced" INTEGER,
  "soundness" TEXT,
  "recorded_output" TEXT,
  "observed_output" TEXT,
  "note" TEXT,
  CHECK ("reproduced" IS NULL OR "reproduced" IN (0, 1)),
  FOREIGN KEY ("soundness") REFERENCES "enum_soundness"("value")
) STRICT;

CREATE TABLE "avenue" (
  "event_id" INTEGER PRIMARY KEY REFERENCES "events"("id"),
  "avenue_id" TEXT,
  "line" TEXT NOT NULL,
  "hypothesis" TEXT,
  "method" TEXT,
  "status" TEXT NOT NULL,
  "supersedes_status" TEXT,
  "reason" TEXT,
  FOREIGN KEY ("status") REFERENCES "enum_avenue_status"("value")
) STRICT;

CREATE TABLE "blue_edit" (
  "event_id" INTEGER PRIMARY KEY REFERENCES "events"("id"),
  "edit_key" TEXT,
  "answers" TEXT,
  "old" TEXT,
  "new" TEXT,
  "text" TEXT,
  "applied_verbatim" INTEGER,
  CHECK ("applied_verbatim" IS NULL OR "applied_verbatim" IN (0, 1))
) STRICT;

CREATE TABLE "revision" (
  "event_id" INTEGER PRIMARY KEY REFERENCES "events"("id"),
  "text" TEXT
) STRICT;

CREATE TABLE "retire" (
  "event_id" INTEGER PRIMARY KEY REFERENCES "events"("id"),
  "claim" TEXT NOT NULL,
  "reason" TEXT NOT NULL,
  "superseded_by" TEXT,
  "removal_basis" TEXT
) STRICT;

CREATE TABLE "manifest_row" (
  "event_id" INTEGER PRIMARY KEY REFERENCES "events"("id"),
  "gap_id" TEXT,
  "row" TEXT,
  FOREIGN KEY ("gap_id") REFERENCES "mint"("gap_id")
) STRICT;

CREATE TABLE "friction" (
  "event_id" INTEGER PRIMARY KEY REFERENCES "events"("id"),
  "text" TEXT,
  "kind" TEXT,
  "estopped_by" TEXT,
  FOREIGN KEY ("kind") REFERENCES "enum_friction_kind"("value")
) STRICT;

CREATE TABLE "friction_none" (
  "event_id" INTEGER PRIMARY KEY REFERENCES "events"("id"),
  "text" TEXT
) STRICT;

CREATE TABLE "inquiry_review" (
  "event_id" INTEGER PRIMARY KEY REFERENCES "events"("id"),
  "reason" TEXT
) STRICT;

CREATE VIEW "gap" AS
SELECT
  m."gap_id"                                   AS "gap_id",
  m."class"                                    AS "class",
  m."problem"                                  AS "problem",
  m."required_fix"                             AS "required_fix",
  m."acceptance_check"                         AS "acceptance_check",
  m."check_kind"                               AS "check_kind",
  m."severity"                                 AS "severity",
  m."likelihood"                               AS "likelihood",
  m."impact"                                   AS "impact",
  m."complexity_cost"                          AS "complexity_cost",
  e."round"                                    AS "minted_round",
  e."seat_id"                                  AS "minted_by",
  c."closure_class"                            AS "closure_class",
  c."successor"                                AS "successor",
  ce."round"                                   AS "merge_closed_round",
  bo."disposition"                             AS "bench_disposition",
  be."seat_id"                                 AS "bench_closed_by",
  be."round"                                   AS "bench_closed_round",
  COALESCE(MIN(ce."round", be."round"), ce."round", be."round") AS "closed_round",
  (c."event_id" IS NULL AND bc."event_id" IS NULL)              AS "open"
FROM "mint" m
JOIN "events" e ON e."id" = m."event_id"
LEFT JOIN "close" c ON c."gap_id" = m."gap_id"
LEFT JOIN "events" ce ON ce."id" = c."event_id"
-- The bench's closing ruling, if it made one. A gap can be ruled on many times — carried in one
-- round and disposed of in the next — so this is the EARLIEST ruling whose disposition closes,
-- and whether it closes is read off the vocabulary rather than decided here.
LEFT JOIN (
  SELECT o."gap_id" AS "gap_id", MIN(o."event_id") AS "event_id"
  FROM "opinion" o
  JOIN "enum_disposition" d ON d."value" = o."disposition"
  WHERE d."closes"
  GROUP BY o."gap_id"
) bc ON bc."gap_id" = m."gap_id"
LEFT JOIN "opinion" bo ON bo."event_id" = bc."event_id"
LEFT JOIN "events" be ON be."id" = bc."event_id";

-- The board's own count, asked once. Every consumer that wants "how many gaps are open" reads this
-- rather than folding the stream again with its own idea of what closed means.
CREATE VIEW "board_counts" AS
SELECT
  (SELECT count(*) FROM "gap" WHERE "open")     AS "open_gaps",
  (SELECT count(*) FROM "gap" WHERE NOT "open") AS "closed_gaps",
  (SELECT count(*) FROM "events")               AS "events";

-- A motion with its filing and its ruling on one row. This join is hand-written at eight readers in
-- the file-backed record, each keying a disposition on a gap_id that the ruling does not carry.
CREATE VIEW "motion_state" AS
SELECT
  m."motion_id"                        AS "motion_id",
  m."subject"                          AS "subject",
  me."seat_id"                         AS "filed_by",
  me."round"                           AS "filed_round",
  g."gap_id"                           AS "gap_id",
  r."grade"                            AS "grade_ruling",
  r."petition"                         AS "petition_ruling",
  r."direction"                        AS "direction_ruling",
  re."seat_id"                         AS "ruled_by",
  re."round"                           AS "ruled_round",
  (r."event_id" IS NULL)               AS "unruled"
FROM "motion" m
JOIN "events" me ON me."id" = m."event_id"
LEFT JOIN "motion_grade" g ON g."event_id" = m."event_id"
LEFT JOIN "motion_rule" r ON r."motion_id" = m."motion_id"
LEFT JOIN "events" re ON re."id" = r."event_id";
