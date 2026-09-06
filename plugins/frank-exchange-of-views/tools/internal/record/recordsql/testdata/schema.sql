
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

CREATE TABLE "enum_event_type" (
  "value" TEXT PRIMARY KEY,
  "means" TEXT NOT NULL
) STRICT;
INSERT INTO "enum_event_type" ("value", "means") VALUES ('anchor', 'evidence tied to a finding: where in the artifact the claim actually lives');
INSERT INTO "enum_event_type" ("value", "means") VALUES ('avenue', 'a line of inquiry, from proposed through pursued, declined, deferred or abandoned');
INSERT INTO "enum_event_type" ("value", "means") VALUES ('base_ingest', 'the frozen round-0 report, stored verbatim as the origin the diff-stack replays over');
INSERT INTO "enum_event_type" ("value", "means") VALUES ('blue_edit', 'a change to the living report, recorded as old and new so the edit itself is auditable');
INSERT INTO "enum_event_type" ("value", "means") VALUES ('certify', 'a seat''s signed statement about its own work — what it asserts on the record');
INSERT INTO "enum_event_type" ("value", "means") VALUES ('cite', 'a source brought into the debate, with the hash and access date that make it re-checkable');
INSERT INTO "enum_event_type" ("value", "means") VALUES ('class_new', 'a defect class coined in this run, with its definition and the neighbour it is distinguished from');
INSERT INTO "enum_event_type" ("value", "means") VALUES ('close', 'a merge closing a gap on a verified repair — red''s half of the closing vocabulary');
INSERT INTO "enum_event_type" ("value", "means") VALUES ('closing', 'a seat''s closing statement on a gap: the argument, not the disposition');
INSERT INTO "enum_event_type" ("value", "means") VALUES ('declare', 'the bench stating a holding that later sittings are expected to apply');
INSERT INTO "enum_event_type" ("value", "means") VALUES ('finding', 'something red found, graded but not yet minted as a gap');
INSERT INTO "enum_event_type" ("value", "means") VALUES ('halt', 'the bench ending the run on a safety, ethics, consent or integrity boundary');
INSERT INTO "enum_event_type" ("value", "means") VALUES ('inquiry_review', 'a review of the lines of inquiry themselves, rather than of a finding');
INSERT INTO "enum_event_type" ("value", "means") VALUES ('log', 'an entry addressed to the operator who can retool the seat: a defect, a request, an impediment, or a nominal sitting');
INSERT INTO "enum_event_type" ("value", "means") VALUES ('manifest_row', 'one row of the run''s manifest, tying a gap to what shipped for it');
INSERT INTO "enum_event_type" ("value", "means") VALUES ('mint', 'a gap put on the board — the act that creates the entity every other act refers to');
INSERT INTO "enum_event_type" ("value", "means") VALUES ('motion', 'a motion filed: a grade contested, a petition to the bench, or a direction proposed');
INSERT INTO "enum_event_type" ("value", "means") VALUES ('motion_appeal', 'an appeal of a ruling already made on a motion');
INSERT INTO "enum_event_type" ("value", "means") VALUES ('motion_rule', 'the bench''s ruling on a filed motion, and whom it binds');
INSERT INTO "enum_event_type" ("value", "means") VALUES ('observe', 'an observation recorded without a claim attached to it');
INSERT INTO "enum_event_type" ("value", "means") VALUES ('outcome', 'the run''s terminal act: how it ended and whether the question was answered');
INSERT INTO "enum_event_type" ("value", "means") VALUES ('position', 'a seat''s stated position going into a round');
INSERT INTO "enum_event_type" ("value", "means") VALUES ('proof', 'a script that was RUN, with its hash and exit status — the answer a computation check demands');
INSERT INTO "enum_event_type" ("value", "means") VALUES ('register', 'a seat took its seat — the first act of any seat, stamping the tool version it ran under');
INSERT INTO "enum_event_type" ("value", "means") VALUES ('regrade', 'a gap''s grade changed, with the basis for the change');
INSERT INTO "enum_event_type" ("value", "means") VALUES ('reproduce', 'an attempt to re-run a recorded proof, and whether what it computes is sound');
INSERT INTO "enum_event_type" ("value", "means") VALUES ('retire', 'a claim withdrawn from the report, with the reason and what supersedes it');
INSERT INTO "enum_event_type" ("value", "means") VALUES ('revision', 'a revision to a seat''s own earlier text');
INSERT INTO "enum_event_type" ("value", "means") VALUES ('sitting_close', 'the harness''s agent returning — the other end of that span');
INSERT INTO "enum_event_type" ("value", "means") VALUES ('sitting_open', 'the harness dispatching an agent — one end of a sitting''s span, observed by a hook rather than claimed by a seat');
INSERT INTO "enum_event_type" ("value", "means") VALUES ('spot_check', 'red re-checking a sample of prior work, or stating that it checked none and why');
INSERT INTO "enum_event_type" ("value", "means") VALUES ('verdict', 'red''s round gate: PASS or FAIL against the open board');
INSERT INTO "enum_event_type" ("value", "means") VALUES ('verify', 'a citation checked at the leaf: what the source did for the claim, and how sure the reader is');

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
INSERT INTO "enum_motion_subject" ("value", "means") VALUES ('docket', 'a gap put before the BENCH for disposition: the filer states the case, the bench rules and its word decides the gap''s fate');
INSERT INTO "enum_motion_subject" ("value", "means") VALUES ('grade', 'you contest a gap''s grade on one dimension');
INSERT INTO "enum_motion_subject" ("value", "means") VALUES ('petition', 'you ask the bench to intervene — the constitutional short-circuit available to any party seat');

CREATE TABLE "enum_grade_dimension" (
  "value" TEXT PRIMARY KEY,
  "means" TEXT NOT NULL
) STRICT;
INSERT INTO "enum_grade_dimension" ("value", "means") VALUES ('complexity', 'what the fix costs; it is what makes defect_accepted arguable');
INSERT INTO "enum_grade_dimension" ("value", "means") VALUES ('impact', 'how far the damage reaches');
INSERT INTO "enum_grade_dimension" ("value", "means") VALUES ('likelihood', 'how likely the CONSEQUENCE is — not how sure you are the defect exists, which is a separate axis');
INSERT INTO "enum_grade_dimension" ("value", "means") VALUES ('severity', 'how bad it is if it bites');

CREATE TABLE "enum_grade" (
  "value" TEXT PRIMARY KEY,
  "means" TEXT NOT NULL,
  "mass" REAL NOT NULL
) STRICT;
INSERT INTO "enum_grade" ("value", "means", "mass") VALUES ('certain', 'the top of the scale — for LIKELIHOOD, reserve it for a consequence that is itself certain, never for a defect you merely verified exists', 3.5);
INSERT INTO "enum_grade" ("value", "means", "mass") VALUES ('high', 'serious', 3);
INSERT INTO "enum_grade" ("value", "means", "mass") VALUES ('low', 'minor', 1);
INSERT INTO "enum_grade" ("value", "means", "mass") VALUES ('low_medium', 'between minor and material', 1.5);
INSERT INTO "enum_grade" ("value", "means", "mass") VALUES ('medium', 'material', 2);
INSERT INTO "enum_grade" ("value", "means", "mass") VALUES ('medium_high', 'between material and serious', 2.5);
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

CREATE TABLE "enum_disposition" (
  "value" TEXT PRIMARY KEY,
  "means" TEXT NOT NULL,
  "closes" INTEGER NOT NULL CHECK ("closes" IN (0, 1))
) STRICT;
INSERT INTO "enum_disposition" ("value", "means", "closes") VALUES ('amends_prior', 'a defect found BETWEEN two repairs that each closed clean earlier — REQUIRES supersedes so the lineage is explicit', 1);
INSERT INTO "enum_disposition" ("value", "means", "closes") VALUES ('carried', 'NOT a closure: the gap survives to the next round with a stated research direction the coming seat owes', 0);
INSERT INTO "enum_disposition" ("value", "means", "closes") VALUES ('defect_accepted', 'the fix costs more than the defect (complexity above likelihood x impact) and the risk is taken KNOWINGLY, with the argument on the record', 1);
INSERT INTO "enum_disposition" ("value", "means", "closes") VALUES ('defect_owed_elsewhere', 'a real defect whose fix is owned outside this debate; it leaves here and is not silently dropped', 1);
INSERT INTO "enum_disposition" ("value", "means", "closes") VALUES ('not_a_defect', 'blue argued the finding was wrong and the argument held; nothing was repaired because nothing needed to be', 1);
INSERT INTO "enum_disposition" ("value", "means", "closes") VALUES ('repaired', 'the repair was verified at the leaf and nothing regressed', 1);
INSERT INTO "enum_disposition" ("value", "means", "closes") VALUES ('repaired_with_regression', 'repaired, but something else broke — REQUIRES a successor naming the gap that carries the regression forward', 1);

CREATE TABLE "enum_ruling_binds" (
  "value" TEXT PRIMARY KEY,
  "means" TEXT NOT NULL
) STRICT;
INSERT INTO "enum_ruling_binds" ("value", "means") VALUES ('blue', 'the relief binds the response seat — what blue must do, or must not, in the coming round');
INSERT INTO "enum_ruling_binds" ("value", "means") VALUES ('both', 'it binds the whole exchange, and every dispatched seat carries it');
INSERT INTO "enum_ruling_binds" ("value", "means") VALUES ('red', 'it binds the audit seats: the lenses and the merge');

CREATE TABLE "enum_check_kind" (
  "value" TEXT PRIMARY KEY,
  "means" TEXT NOT NULL
) STRICT;
INSERT INTO "enum_check_kind" ("value", "means") VALUES ('computation', 'RUNNING something settles it. This check CANNOT be closed by prose: it closes only when a proof answers the gap. Reach for it wherever the answer would be PRODUCED rather than asserted — arithmetic, a simulation, a forecast, a parse, a count, a re-derivation are common cases and not the whole of it; if you can imagine a script that would end the argument, this is the kind');
INSERT INTO "enum_check_kind" ("value", "means") VALUES ('document', 'reading a shipped artifact settles it — the check is answered by prose that quotes what is there');
INSERT INTO "enum_check_kind" ("value", "means") VALUES ('source', 'verifying an external source settles it — the claim stands or falls on what the cited material actually says');

CREATE TABLE "enum_finding_about" (
  "value" TEXT PRIMARY KEY,
  "means" TEXT NOT NULL
) STRICT;
INSERT INTO "enum_finding_about" ("value", "means") VALUES ('gap', 'a gap already on the docket, by its id — a defect in the record rather than in the report');
INSERT INTO "enum_finding_about" ("value", "means") VALUES ('inquiry', 'a line of inquiry, by its avenue id: an argument against the REASON it was declined, deferred or abandoned. The steelman duty''s own anchor');
INSERT INTO "enum_finding_about" ("value", "means") VALUES ('section', 'a named report section, for something MISSING from it — the anchor a quote cannot provide, because the text you are objecting to is not there');

CREATE TABLE "enum_source_text_read" (
  "value" TEXT PRIMARY KEY,
  "means" TEXT NOT NULL
) STRICT;
INSERT INTO "enum_source_text_read" ("value", "means") VALUES ('leaf', 'the source''s own text was read at the leaf, in the bytes the run cached. The only value that licenses a claim about what the source SAYS');
INSERT INTO "enum_source_text_read" ("value", "means") VALUES ('summary_only', 'read only through someone else''s account of it — an abstract, a secondary description, or the summary of an INTERESTED party. Everything the report says about its contents is that account, not the source');
INSERT INTO "enum_source_text_read" ("value", "means") VALUES ('unread', 'the text was never read — the citation rests on a record that the source EXISTS (a bibliographic index, a search result), not on anything it says');

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

CREATE TABLE "enum_log_type" (
  "value" TEXT PRIMARY KEY,
  "means" TEXT NOT NULL
) STRICT;
INSERT INTO "enum_log_type" ("value", "means") VALUES ('defect', 'something is broken: it did the wrong thing, or failed where it should have worked. A tool that fails INTERNALLY records this too, as (TOOL, DEFECT) — an error nobody learns about is one nothing improves on');
INSERT INTO "enum_log_type" ("value", "means") VALUES ('estoppel', 'the TOOL refused a mint because the defect lives in text blue applied verbatim from red''s own --fix-new. Recorded by the tool, not filed by the seat: argue it on the original gap, or mint with --supersedes so the lineage is explicit');
INSERT INTO "enum_log_type" ("value", "means") VALUES ('friction', 'the work was impeded and you are noting it; NOT necessarily actionable and not necessarily advisable to change. The honest home for an entry that would otherwise have to pose as a defect');
INSERT INTO "enum_log_type" ("value", "means") VALUES ('nominal', 'the surface met the work — the sitting is clean, said in the positive. An entry exists, so an attested-clean sitting stays distinguishable from a channel nobody used');
INSERT INTO "enum_log_type" ("value", "means") VALUES ('request', 'a capability that does not exist — the act you wanted was on no surface, so there was nothing to get wrong. Distinct from a defect because the fix is to build, not to repair');

CREATE TABLE "enum_log_source" (
  "value" TEXT PRIMARY KEY,
  "means" TEXT NOT NULL
) STRICT;
INSERT INTO "enum_log_source" ("value", "means") VALUES ('seat', 'a seat filed this about its own sitting');
INSERT INTO "enum_log_source" ("value", "means") VALUES ('tool', 'the tool emitted this itself, rather than a seat filing it');

CREATE TABLE "register" (
  "event_id" INTEGER PRIMARY KEY REFERENCES "events"("id"),
  "tool_version" TEXT,
  "agent_id" TEXT,
  "served_model" TEXT,
  "requested_model" TEXT,
  "run_via" TEXT,
  "agent_type" TEXT
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

CREATE TABLE "motion_docket" (
  "event_id" INTEGER PRIMARY KEY REFERENCES "motion"("event_id"),
  "gap_id" TEXT NOT NULL,
  FOREIGN KEY ("gap_id") REFERENCES "mint"("gap_id")
) STRICT;

CREATE TABLE "motion_rule" (
  "event_id" INTEGER PRIMARY KEY REFERENCES "events"("id"),
  "motion_id" TEXT NOT NULL,
  "subject" TEXT,
  "opinion" TEXT NOT NULL,
  "binds" TEXT,
  "grade" TEXT,
  "petition" TEXT,
  "direction" TEXT,
  "ruling_case" TEXT,
  CHECK (("grade" IS NOT NULL) + ("petition" IS NOT NULL) + ("direction" IS NOT NULL) + ("ruling_case" IS NOT NULL) <= 1),
  FOREIGN KEY ("subject") REFERENCES "enum_motion_subject"("value"),
  FOREIGN KEY ("binds") REFERENCES "enum_ruling_binds"("value"),
  FOREIGN KEY ("grade") REFERENCES "enum_grade_ruling"("value"),
  FOREIGN KEY ("petition") REFERENCES "enum_petition_ruling"("value"),
  FOREIGN KEY ("direction") REFERENCES "enum_direction_ruling"("value")
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
  CHECK ("final" IS NULL OR "final" IN (0, 1)),
  CHECK ("reopens_on" IS NOT NULL OR "final" IS NOT NULL),
  CHECK ("reopens_on" IS NULL OR "final" IS NULL),
  FOREIGN KEY ("disposition") REFERENCES "enum_disposition"("value")
) STRICT;

CREATE TABLE "motion_appeal" (
  "event_id" INTEGER PRIMARY KEY REFERENCES "events"("id"),
  "motion_id" TEXT,
  "subject" TEXT,
  "reason" TEXT,
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
  "prose" TEXT,
  CHECK ("closure_class" IS NULL OR "closure_class" IN ('amends_prior', 'defect_accepted', 'defect_owed_elsewhere', 'not_a_defect', 'repaired', 'repaired_with_regression')),
  CHECK ("closure_class" <> 'repaired_with_regression' OR "successor" IS NOT NULL),
  CHECK ("carried_from" IS NOT NULL OR "prose" IS NOT NULL),
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
  "about_kind" TEXT,
  "about_ref" TEXT,
  FOREIGN KEY ("severity") REFERENCES "enum_grade"("value"),
  FOREIGN KEY ("likelihood") REFERENCES "enum_grade"("value"),
  FOREIGN KEY ("impact") REFERENCES "enum_grade"("value"),
  FOREIGN KEY ("about_kind") REFERENCES "enum_finding_about"("value")
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
  "text" TEXT,
  "source_text_read" TEXT,
  FOREIGN KEY ("source_text_read") REFERENCES "enum_source_text_read"("value")
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
  "label" TEXT,
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
  "exit" INTEGER,
  "location" TEXT
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
  "line" TEXT,
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
  "accepted" INTEGER,
  CHECK ("applied_verbatim" IS NULL OR "applied_verbatim" IN (0, 1)),
  CHECK ("accepted" IS NULL OR "accepted" IN (0, 1))
) STRICT;

CREATE TABLE "blue_edit_reopened" (
  "event_id" INTEGER NOT NULL REFERENCES "blue_edit"("event_id"),
  "ord"      INTEGER NOT NULL,
  "value"    TEXT    NOT NULL,
  PRIMARY KEY ("event_id", "ord")
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

CREATE TABLE "log" (
  "event_id" INTEGER PRIMARY KEY REFERENCES "events"("id"),
  "text" TEXT,
  "type" TEXT NOT NULL,
  "source" TEXT,
  "estopped_by" TEXT,
  FOREIGN KEY ("type") REFERENCES "enum_log_type"("value"),
  FOREIGN KEY ("source") REFERENCES "enum_log_source"("value")
) STRICT;

CREATE TABLE "inquiry_review" (
  "event_id" INTEGER PRIMARY KEY REFERENCES "events"("id"),
  "reason" TEXT
) STRICT;

CREATE TABLE "base_ingest" (
  "event_id" INTEGER PRIMARY KEY REFERENCES "events"("id"),
  "text" TEXT
) STRICT;

CREATE TABLE "sitting_open" (
  "event_id" INTEGER PRIMARY KEY REFERENCES "events"("id"),
  "agent_id" TEXT,
  "agent_type" TEXT
) STRICT;

CREATE TABLE "sitting_close" (
  "event_id" INTEGER PRIMARY KEY REFERENCES "events"("id"),
  "agent_id" TEXT,
  "agent_type" TEXT
) STRICT;

CREATE INDEX "round_verdict_verdict" ON "round_verdict" ("verdict");

-- THE AGENT -> SEAT BINDING, AS SQL, so a telemetry view can name a seat without any reader
-- re-deriving the rule. It is the same rule record.SeatOfAgent applies in Go and states in prose:
-- THE LAST REGISTER WINS, because a re-dispatch writes a fresh register event and a resumed seat
-- legitimately arrives under a new agent id claiming a seat that is already bound. Treating that
-- as a conflict would refuse every resume.
--
-- agent_type comes along because register records it too: it is what the HARNESS called the
-- seat, beside what the seat called itself, and the two disagreeing is a thing worth being able
-- to see rather than a thing to collapse here.
CREATE VIEW "seat_of_agent" AS
SELECT
  r."agent_id"   AS "agent_id",
  e."seat_id"    AS "seat_id",
  r."agent_type" AS "agent_type",
  e."round"      AS "registered_round"
FROM "register" r
JOIN "events" e ON e."id" = r."event_id"
WHERE r."agent_id" IS NOT NULL AND r."agent_id" != ''
  AND r."event_id" = (SELECT MAX(r2."event_id") FROM "register" r2 WHERE r2."agent_id" = r."agent_id");

-- WHAT A SEAT COST, from the turns ingested at capture (#684 F16).
--
-- A LEFT JOIN, deliberately. A seat whose turns were measured but which never registered still
-- appears, with a null seat_id — the alternative drops its rows from every total and reports a
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
  MAX(NULLIF(t."ts_ms", 0)) - MIN(NULLIF(t."ts_ms", 0))    AS "wall_ms"
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

CREATE VIEW "gap" AS
SELECT
  m."gap_id"                                   AS "gap_id",
  m."class"                                    AS "class",
  m."location"                                 AS "location",
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
  e."round"                                    AS "minted_round",
  e."seat_id"                                  AS "minted_by",
  c."closure_class"                            AS "closure_class",
  c."successor"                                AS "successor",
  ce."round"                                   AS "merge_closed_round",
  bo."disposition"                             AS "bench_disposition",
  be."seat_id"                                 AS "bench_closed_by",
  be."round"                                   AS "bench_closed_round",
  COALESCE(MIN(ce."round", be."round"), ce."round", be."round") AS "closed_round",
  (c."event_id" IS NULL AND bc."event_id" IS NULL)              AS "open",
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
  -- THE CURRENT GRADES, one per axis: the latest regrade that touched the axis, else the
  -- mint's. The plain "severity" columns above are MINT-TIME grades — a reader that wants
  -- what the gap is graded NOW and reaches for them silently reads a number a regrade may
  -- have moved, which is why the overlay is answered here rather than left to each reader.
  COALESCE((SELECT r."severity" FROM "regrade" r WHERE r."gap_id" = m."gap_id"
    AND r."severity" IS NOT NULL ORDER BY r."event_id" DESC LIMIT 1), m."severity")   AS "current_severity",
  COALESCE((SELECT r."likelihood" FROM "regrade" r WHERE r."gap_id" = m."gap_id"
    AND r."likelihood" IS NOT NULL ORDER BY r."event_id" DESC LIMIT 1), m."likelihood") AS "current_likelihood",
  COALESCE((SELECT r."impact" FROM "regrade" r WHERE r."gap_id" = m."gap_id"
    AND r."impact" IS NOT NULL ORDER BY r."event_id" DESC LIMIT 1), m."impact")       AS "current_impact",
  COALESCE((SELECT r."complexity_cost" FROM "regrade" r WHERE r."gap_id" = m."gap_id"
    AND r."complexity_cost" IS NOT NULL ORDER BY r."event_id" DESC LIMIT 1), m."complexity_cost") AS "current_complexity_cost",
  -- THE PROOF JOIN, answered where every asker can share it: a 'computation' gap closes on
  -- a recorded proof naming it in --answers, and "awaiting proof" is the debt list blue is
  -- handed. 'computation' is the vocabulary's own word; the Go home for the question
  -- (Gap.NeedsComputation) reads the enum, and THIS is the one SQL statement of it.
  EXISTS(SELECT 1 FROM "proof" p WHERE p."answers" = m."gap_id")                      AS "proof_answered",
  (c."event_id" IS NULL AND bc."event_id" IS NULL
     AND m."check_kind" = 'computation'
     AND NOT EXISTS(SELECT 1 FROM "proof" p WHERE p."answers" = m."gap_id"))          AS "awaiting_proof",
  -- LINEAGE FROM THE OTHER END: the LAST gap that claimed to replace this one, and whether
  -- that promise is broken — a superseded ancestor still open is the same defect counted
  -- twice, which is what the verdict gate refuses.
  (SELECT m2."gap_id" FROM "mint_supersedes" s2 JOIN "mint" m2 ON m2."event_id" = s2."event_id"
    WHERE s2."value" = m."gap_id" ORDER BY s2."event_id" DESC LIMIT 1)                AS "superseded_by",
  (c."event_id" IS NULL AND bc."event_id" IS NULL
     AND EXISTS(SELECT 1 FROM "mint_supersedes" s3 WHERE s3."value" = m."gap_id"))    AS "stranded",
  -- The mint's own event id, so board order is a sort key rather than a join every reader
  -- writes for itself.
  m."event_id"                                                                        AS "minted_event"
FROM "mint" m
JOIN "events" e ON e."id" = m."event_id"
-- A gap can be closed more than once — red re-adjudicates across rounds (defect_accepted in
-- one, repaired in a later one). Take the EARLIEST close, exactly as the bench-close arm below
-- takes its earliest closing ruling: a plain LEFT JOIN "close" fans out one row per close event,
-- and a gap with two closes then counts twice in board_counts while the raw event walk counts it
-- once (a projection disagreement the consistency oracle catches). closed_round's MIN already
-- assumed the earliest; this makes the row do so too.
LEFT JOIN (
  SELECT c0."gap_id" AS "gap_id", MIN(c0."event_id") AS "event_id"
  FROM "close" c0
  GROUP BY c0."gap_id"
) cx ON cx."gap_id" = m."gap_id"
LEFT JOIN "close" c ON c."event_id" = cx."event_id"
LEFT JOIN "events" ce ON ce."id" = cx."event_id"
-- The bench's closing ruling, if it made one. A gap can be ruled on many times — carried in one
-- round and disposed of in the next — so this is the EARLIEST ruling whose disposition closes,
-- and whether it closes is read off the vocabulary rather than decided here.
--
-- TWO HOPS NOW, BECAUSE THE GAP RIDES THE FILING. The bench's disposition was its own event
-- carrying a gap_id; it is a docket motion's RULING, and the gap is on the motion that asked.
-- So: the ruling arm gives the disposition, its motion_rule gives the motion id, and the docket
-- FILING gives the gap. Written here once rather than at each reader, which is what this view is
-- for — the same join was hand-written at eight readers before motion_state existed.
--
-- AND IT HAD TO CHANGE IN THE SAME COMMIT AS THE DELETE. SQLite does not validate a view body at
-- CREATE, so a view left reading "opinion" after that table went would have applied cleanly and
-- returned no bench closures at all: every disposed gap reading as undisposed, which is the exact
-- defect this whole change exists to remove.
LEFT JOIN (
  SELECT md."gap_id" AS "gap_id", MIN(mr."event_id") AS "event_id"
  FROM "motion_rule_docket" rd
  JOIN "motion_rule" mr ON mr."event_id" = rd."event_id"
  JOIN "motion" mo ON mo."motion_id" = mr."motion_id"
  JOIN "motion_docket" md ON md."event_id" = mo."event_id"
  JOIN "enum_disposition" d ON d."value" = rd."disposition"
  WHERE d."closes"
  GROUP BY md."gap_id"
) bc ON bc."gap_id" = m."gap_id"
LEFT JOIN "motion_rule_docket" bo ON bo."event_id" = bc."event_id"
LEFT JOIN "events" be ON be."id" = bc."event_id";

-- The board's own count, asked once. Every consumer that wants "how many gaps are open" reads this
-- rather than folding the stream again with its own idea of what closed means.
-- THE NEVER-HARD-FAIL DETECTOR, ASKED OF THE RECORD RATHER THAN ASSERTED BY THE ENGINE.
--
-- debate.js computes this every round and writes it to a LOG LINE. The scorecard tried to read
-- it from a telemetry key nothing ever wrote, so the detector reported 0 on every run for seven
-- runs — "no soft fails" in the words it would use for "never measured". The fact existed; its
-- only carrier was prose.
--
-- It is a QUESTION about the record, which is what this file is for, and every input is already
-- here: the round's verdict, the mass of what is still open, the top severity, and whether any
-- gap minted this round is fresh rather than lineage. So it is authored once, where a reader can
-- see the fold, instead of recomputed in whichever consumer wants it.
--
-- MASS IS A JOIN NOW, and that is the change that made this expressible at all. A grade's weight
-- is a facet on the vocabulary (enum_grade.mass), so board mass is
-- MASS[likelihood] * MASS[impact] summed — the same formula the Go and JS copies apply, read
-- off the same table the schema built from the enum. It used to be a hand-written map in two
-- languages with a regex test holding them level, and SQL could not ask the question at all.
--
-- The thresholds are the engine's, restated once here: mass < 35, nothing above medium (mass 2),
-- zero fresh mints, verdict FAIL.
CREATE VIEW "convergence_vs_verdict" AS
SELECT
  v."round"                                        AS "round",
  rv."verdict"                                     AS "verdict",
  COALESCE(b."mass", 0.0)                          AS "mass",
  COALESCE(b."max_severity_mass", 0.0)             AS "max_severity_mass",
  COALESCE(f."fresh_mints", 0)                     AS "fresh_mints",
  (rv."verdict" = 'fail'
     AND COALESCE(b."mass", 0.0) < 35.0
     AND COALESCE(b."max_severity_mass", 0.0) <= 2.0
     AND COALESCE(f."fresh_mints", 0) = 0)         AS "divergent"
FROM (SELECT DISTINCT "round" FROM "events" WHERE "round" > 0) v
JOIN "events" ve ON ve."round" = v."round" AND ve."type" = 'verdict'
JOIN "round_verdict" rv ON rv."event_id" = ve."id"
LEFT JOIN (
  -- Open AT that round: minted on or before it, and not closed before it ends.
  SELECT
    r."round"                                                  AS "round",
    SUM(COALESCE(gl."mass", 0.0) * COALESCE(gi."mass", 0.0))   AS "mass",
    MAX(COALESCE(gs."mass", 0.0))                              AS "max_severity_mass"
  FROM (SELECT DISTINCT "round" FROM "events" WHERE "round" > 0) r
  JOIN "gap" g
    ON g."minted_round" <= r."round"
   AND (g."open" OR g."closed_round" > r."round")
  LEFT JOIN "enum_grade" gl ON gl."value" = g."likelihood"
  LEFT JOIN "enum_grade" gi ON gi."value" = g."impact"
  LEFT JOIN "enum_grade" gs ON gs."value" = g."severity"
  GROUP BY r."round"
) b ON b."round" = v."round"
LEFT JOIN (
  -- FRESH means minted this round and superseding nothing: a lineage mint is a repair of known
  -- work, not new discovery, which is the distinction the detector turns on.
  SELECT "minted_round" AS "round", count(*) AS "fresh_mints"
  FROM "gap"
  WHERE "supersedes_count" = 0
  GROUP BY "minted_round"
) f ON f."round" = v."round";

CREATE VIEW "board_counts" AS
SELECT
  (SELECT count(*) FROM "gap" WHERE "open")     AS "open_gaps",
  (SELECT count(*) FROM "gap" WHERE NOT "open") AS "closed_gaps",
  (SELECT count(*) FROM "events")               AS "events";

-- THE ANSWERS TO A MOTION, one row per answered id, whichever filing shape asked. The FIRST
-- ruling and the FIRST appeal are the ones that count: a second of either is refused at the
-- write precisely because it would replace the first in every later reader — so the view
-- states the first-wins rule ONCE, where a legacy record carrying an illegal second row
-- cannot multiply anybody's join. Keyed on the id alone rather than joined to 'motion',
-- because a direction motion has no filing row (the line of inquiry's proposal IS the
-- filing) and its answers must be askable all the same.
CREATE VIEW "motion_answers" AS
SELECT
  ids."motion_id"                                        AS "motion_id",
  fr."grade"                                             AS "grade",
  fr."petition"                                          AS "petition",
  fr."direction"                                         AS "direction",
  -- THE DOCKET ARM IS A TABLE, NOT A COLUMN, and that is why it is joined rather than read.
  -- Its three siblings are enums and land as columns on "motion_rule"; the bench's is a MESSAGE
  -- (it carries the principle, the tension and what would reopen it), so its disposition lives
  -- one table down. Left out of the COALESCE below, "ruling" is NULL for every bench ruling ever
  -- made and RequireUnruledMotion reads the whole docket as unanswered.
  rd."disposition"                                       AS "docket",
  COALESCE(fr."grade", fr."petition", fr."direction", rd."disposition") AS "ruling",
  fre."seat_id"                                          AS "ruled_by",
  fre."round"                                            AS "ruled_round",
  fa."reason"                                            AS "appeal_reason",
  fae."seat_id"                                          AS "appealed_by",
  fae."round"                                            AS "appealed_round"
FROM (SELECT "motion_id" FROM "motion_rule" UNION SELECT "motion_id" FROM "motion_appeal") ids
LEFT JOIN "motion_rule" fr ON fr."event_id" =
  (SELECT MIN(x."event_id") FROM "motion_rule" x WHERE x."motion_id" = ids."motion_id")
LEFT JOIN "motion_rule_docket" rd ON rd."event_id" = fr."event_id"
LEFT JOIN "events" fre ON fre."id" = fr."event_id"
LEFT JOIN "motion_appeal" fa ON fa."event_id" =
  (SELECT MIN(y."event_id") FROM "motion_appeal" y WHERE y."motion_id" = ids."motion_id")
LEFT JOIN "events" fae ON fae."id" = fa."event_id";

-- A motion with its filing and its ruling on one row. This join is hand-written at eight readers in
-- the file-backed record, each keying a disposition on a gap_id that the ruling does not carry.
-- The answer half comes from motion_answers, so the first-wins rule has ONE statement: the
-- old inline LEFT JOIN on motion_rule multiplied this view's rows for a motion carrying two
-- rulings — exactly the legacy shape the write guard now refuses, read as two motions.
CREATE VIEW "motion_state" AS
SELECT
  m."motion_id"                        AS "motion_id",
  m."subject"                          AS "subject",
  me."seat_id"                         AS "filed_by",
  me."round"                           AS "filed_round",
  -- THE GAP COMES FROM WHICHEVER FILING ARM CARRIES ONE. A bare read off motion_grade was correct
  -- while grade was the only subject about a gap; docket is the second.
  COALESCE(g."gap_id", gd."gap_id")    AS "gap_id",
  a."grade"                            AS "grade_ruling",
  a."petition"                         AS "petition_ruling",
  a."direction"                        AS "direction_ruling",
  a."docket"                           AS "docket_ruling",
  a."ruled_by"                         AS "ruled_by",
  a."ruled_round"                      AS "ruled_round",
  (a."ruled_by" IS NULL)               AS "unruled",
  a."appealed_by"                      AS "appealed_by",
  a."appeal_reason"                    AS "appeal_reason"
FROM "motion" m
JOIN "events" me ON me."id" = m."event_id"
LEFT JOIN "motion_grade" g ON g."event_id" = m."event_id"
LEFT JOIN "motion_docket" gd ON gd."event_id" = m."event_id"
LEFT JOIN "motion_answers" a ON a."motion_id" = m."motion_id";

-- A LINE OF INQUIRY, whole: proposed by whom, saying what, where its status stands now, and
-- how red last ruled the direction — the join that used to live in three separate readers
-- (the Inquiries fold, InquiryRuling, and the report's rows). The proposal row carries the
-- substance; the LATEST avenue event carries the status; the LATEST direction-subject
-- ruling carries red's answer, including a later ruling that carried no word — an unset arm
-- on the newest ruling is red ruling NOTHING, not an invitation to read an older one.
CREATE VIEW "line_of_inquiry" AS
SELECT
  p."avenue_id"                        AS "avenue_id",
  pe."seat_id"                         AS "proposed_by",
  pe."round"                           AS "proposed_round",
  fp."line"                            AS "line",
  ls."status"                          AS "status",
  lse."round"                          AS "status_round",
  lr."direction"                       AS "direction_ruling",
  lre."seat_id"                        AS "ruled_by",
  lre."round"                          AS "ruled_round"
FROM (SELECT "avenue_id", MIN("event_id") AS "pid" FROM "avenue"
        WHERE COALESCE("avenue_id", '') != '' AND "supersedes_status" IS NULL
        GROUP BY "avenue_id") p
JOIN "avenue" fp ON fp."event_id" = p."pid"
JOIN "events" pe ON pe."id" = p."pid"
LEFT JOIN "avenue" ls ON ls."event_id" =
  (SELECT MAX(x."event_id") FROM "avenue" x WHERE x."avenue_id" = p."avenue_id")
LEFT JOIN "events" lse ON lse."id" = ls."event_id"
LEFT JOIN "motion_rule" lr ON lr."event_id" =
  (SELECT MAX(y."event_id") FROM "motion_rule" y
     WHERE y."motion_id" = p."avenue_id" AND y."subject" = 'direction')
LEFT JOIN "events" lre ON lre."id" = lr."event_id";

-- report_op is the ORDERED STREAM OF TEXT MUTATIONS that reconstruct blue's report (#709). The
-- report is base + this stream replayed, so the SELECTION of which events mutate the text — and
-- how each names its change — is a projection, authored HERE where every reader sees the same one,
-- not re-derived by walking every event in Go. Each row is (event_id, kind, a, b):
--   edit   → a=old span, b=new span (blue edit's splice, located and replaced at replay).
--   insert → a=the anchoring quote, b=the marker id (Token(b) is spliced at that quote).
-- The marker inserters are blue cite, blue prove, the finding's anchor event, and a red
-- corroboration (a labelled verify). A marker event with no anchor location placed no marker in
-- THIS report (a board/docket-only citation), so it is excluded rather than replayed as an empty
-- insert. The FOLD itself stays in Go: replaying a splice needs the running text a prior op left,
-- which SQL cannot carry — but nothing here builds state by scanning the whole event log.
CREATE VIEW "report_op" AS
  SELECT e."id" AS "event_id", 'edit' AS "kind", b."old" AS "a", b."new" AS "b"
    FROM "blue_edit" b JOIN "events" e ON e."id" = b."event_id"
  UNION ALL
  SELECT e."id", 'insert', c."location", c."label"
    FROM "cite" c JOIN "events" e ON e."id" = c."event_id"
    WHERE COALESCE(c."location", '') != '' AND COALESCE(c."label", '') != ''
  UNION ALL
  SELECT e."id", 'insert', p."location", p."proof_id"
    FROM "proof" p JOIN "events" e ON e."id" = p."event_id"
    WHERE COALESCE(p."location", '') != '' AND COALESCE(p."proof_id", '') != ''
  UNION ALL
  SELECT e."id", 'insert', a."location", a."id"
    FROM "anchor" a JOIN "events" e ON e."id" = a."event_id"
    WHERE COALESCE(a."location", '') != '' AND COALESCE(a."id", '') != ''
  UNION ALL
  SELECT e."id", 'insert', v."claim", v."label"
    FROM "verify" v JOIN "events" e ON e."id" = v."event_id"
    WHERE COALESCE(v."label", '') != '' AND COALESCE(v."claim", '') != '';
