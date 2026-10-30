package seat

import (
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchortext"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/enumhelp"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/compute"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/jsonshape"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/report"
	// Places renders the report through the renderer reportproj registers with the record.
	_ "github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/reportproj"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatenv"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/sittingcap"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/view"
)

// The verbs every role shares, defined once.
//
// register, log, position and closing are the SAME contract
// wherever they appear — a log entry from a lens and one from the bench are
// the same event with the same payload. Restating them per role would be four
// copies to drift apart, and the drift would be silent because each copy would
// still pass its own tests.
//
// The help lives in help/<key>.md. `register` means the same thing in every chair and is ONE
// document; blue's page includes it and adds the repair, which only blue's surface carries.
// `closing` genuinely differs by party — a RED CLOSING is the minting lens's, a BLUE CLOSING
// blue-respond's — and is keyed by role. `position` is the chair's alone.

// Register is the register every red seat and the bench mounts.
func Register() *cobra.Command { return NewKeyed("register", "register", registerResult{}, register) }

// BenchRegister is the bench's register: the same act, and the one seat whose id does not say what
// the sitting is for. The bench is asked four different questions under one id — rule the docket,
// hear a petition, dispose at the exit, assemble the report — so it alone names the occasion, and
// the record refuses a bench register without one.
//
// The flag is on THIS surface only, never on a lens's or blue's and always refused there: a flag
// a seat can see and never legitimately use is an invitation to try it.
func BenchRegister() *cobra.Command {
	c := NewKeyed("register", "register-bench", registerResult{}, register)
	enumhelp.Flag(c, flags.Occasion, record.MustEnum("register", "occasion"),
		"what this sitting was convened to do. Your prompt says which; this is where it reaches the record, and once the run is archived it is the only thing telling your four sittings apart")
	return Require(c, flags.Occasion)
}

func register(s Context, cmd *cobra.Command) (Result, error) {
	// The occasion is absent on every surface but the bench's, where the record REQUIRES it: Str
	// reads "" and checkOccasion refuses that for a bench seat and accepts it for everyone else.
	dispatch, _, err := record.RegisterSeat(s.Identity(), string(s.RunVia), Str(cmd, flags.Occasion))
	if err != nil {
		return nil, err
	}
	r := registerResult{SeatID: s.SeatID, Dispatch: dispatch, RunVia: string(s.RunVia)}
	// THE SITTING'S TOOL-CALL COUNT STARTS HERE, keyed by the sitting this register opened or
	// joined: the seat's stored sitting count, the number every `#N` on the record reads. A register
	// that opens a sitting starts a fresh count; one that joins its bracket's sitting, or a second
	// register in one sitting, keeps counting in that sitting. An operator is never limited; an
	// agent with no id cannot be counted.
	if agent := seatenv.AgentID(); agent != "" && s.SeatID != record.OperatorRole {
		sitting, err := record.SittingsOf(s.Identity().Run, s.SeatID)
		if err == nil {
			err = sittingcap.Open(s.Identity().Run.Dir(), agent, sittingcap.Header{SeatID: s.SeatID, Sitting: sitting})
		}
		if err != nil {
			r.TurnLimitUnarmed = err.Error()
		}
	}
	// THE IDENTITY DID NOT ARRIVE, AND THE SEAT IS THE ONLY PARTY THAT CAN STILL ACT ON IT.
	//
	// register is where the seat writes its own agent->seat binding, so it is also the first and
	// only place the seat can see the binding was never possible. The record already handles this
	// correctly at the write — absent is recorded as absent — and capture already reports it
	// hours later. What nothing did was tell the SEAT, while the sitting it affects is still
	// running.
	//
	// MEASURED, and it is the whole of #512. In research/2026-08-22_is-7-prime all FOURTEEN
	// registers carry no agent_id: the hook never fired for that run at all, so every later
	// call was refused "this agent has not registered" whatever shape it took. Across the
	// session that message was returned 92 times and was FALSE every single time — zero of
	// them were a seat that had genuinely not registered.
	//
	// A WARNING, NOT A REFUSAL. Refusing here would wedge a run whose only fault is a hook
	// that did not fire, and the seat can work perfectly well by passing --seat-id.
	if seatenv.AgentID() == "" && s.SeatID != record.OperatorRole {
		r.IdentityAbsent = true
	}
	// THE HOOK IS NOT REACHING THIS RUN — a bigger fact than a missing identity, and one the
	// hook itself records nowhere. It injects FEOV_RUN on every Bash call in a live run, so
	// a run directory that arrived any OTHER way means the hook declined, was never invoked,
	// or found no marker from the payload's cwd.
	//
	// THE WRAPPER MAKES THIS A CLEAN READING ON ITS OWN. `setup` bakes the run into
	// <runDir>/.bin/feov-record, and the hook — when it fires — sets FEOV_RUN, which
	// outranks it. So resolving by WRAPPER is a direct statement that the hook did not
	// inject, needing no pairing with a missing identity to mean it. Inference still needs
	// the pair, because a seat that simply typed --run looks identical from the run alone.
	if s.RunVia == seatenv.RunFromWrapper || (r.IdentityAbsent && s.RunVia == seatenv.RunFromInference) {
		r.HookAbsent = true
	}
	return r, nil
}

// Log records an entry addressed to the OPERATOR WHO CAN RETOOL THE SEAT — a defect, a request, or
// friction. It is not material for the debate.
//
// WHY THE CLEAN CASE IS NOT AN ENTRY. "Nothing on the record" was once equally consistent with a
// clean sitting and with a seat that hit walls and never used the channel — the same bytes, which
// is this project's recurring defect. The harness now brackets every dispatch at both ends, so the
// record holds that a seat SAT; clean is derived from a sitting that filed nothing, and the seat
// spends no call saying so.
//
// It is not that seats were unwilling. Most refusals they meet are their OWN errors, correctly
// read as such — that is not a defect and should not be filed. But one seat, blocked by a motion
// it could not read, reasoned in its own words that no read verb existed, searched ten-plus calls,
// and then GUESSED rather than reporting it: filing costs a turn and does not unblock, while
// guessing might. A seat under that pressure needs the report to be a duty it discharges, not an
// invitation it declines.
//
// THE TYPE IS WHY THE OPERATOR CAN TRIAGE. Measured on 2026-09-02_quadratic-formula, the channel
// ran 142,891 characters and a reader had to read all of it to learn which entries were
// actionable. `--type` moves that from the reader's inference to the writer's assertion — and
// `friction` is a first-class value precisely so an entry that is NOT a request for change has an
// honest home instead of having to pose as one.
//
// THE PAGE (help/log.md) INSTRUCTS AND THIS IS ITS WHY. The type is asserted because the operator
// triages by FILTERING on it instead of reading every entry. A guessed name is friction even when
// the guess was the seat's own, because something taught it to expect that name, and that is what
// the operator retools. A missing capability is reported and not engineered around, because a
// workaround leaves no trace and the capability then looks exactly like one nobody wanted. An entry
// does not retell the sitting, because the record holds it and a diary buries the one sentence meant
// for the operator. A sitting that recorded nothing owes no entry: silence is ambiguous only where
// the sitting DID things, and one with no acts is bracketed by the harness with the seat's identity,
// so that it ran is on the record — deriving the clean case from the emptiness beats asserting it.
func Log() *cobra.Command {
	c := Prose(New("log", Msg{}, func(s Context, cmd *cobra.Command) (Result, error) {
		word := Str(cmd, flags.Type)
		text, err := Reason(cmd)
		if err != nil {
			return nil, err
		}
		// THE TYPE IS THE TRIAGE, so it is required and refused rather than defaulted — at flag
		// parse, not here. The schema marks the field required, so cobra refuses a call without
		// --type before this closure runs, and the flag's own enum set is SeatLogTypeEnum, so a
		// word outside the three a seat may file — a tool-only word included — is refused at parse
		// with the seat's set listed. A branch here that refused an empty word could never fire,
		// and the one that did told a seat "a clean sitting is `nominal`" after the value was
		// retired (#1123): a refusal no test drives is where a retired word survives. The miss
		// below is the same kind of branch, so it lists no set — the seat's set has ONE
		// derivation, the ToolOnly facet SeatLogTypeEnum reads, and a word list here would be
		// a second.
		lt, ok := record.LogTypeOf(word)
		if !ok {
			return nil, feov.Errorf(feov.Validation, "%q is not a log type this record can carry", word)
		}
		// NO SEAT-SIDE FACET CHECK HERE EITHER. The invariant that holds for every writer — a
		// tool-only word carries source=TOOL — lives at the record's write path, where the
		// tool's own estoppel write passes it too.
		src := recordpb.LogSource_LOG_SOURCE_SEAT
		if _, err := record.Append(s.Identity(), &recordpb.Log{
			Text:   proto.String(text),
			Type:   &lt,
			Source: &src,
		}); err != nil {
			return nil, err
		}
		return Msg{Message: "log entry recorded: " + recordpb.Word(lt)}, nil
	}))
	// THE HELP OFFERS WHAT THE WRITE PATH ACCEPTS, and it used to offer one word more. A surface
	// that lists a value its own verb refuses is lying about its contract, and this is the page a
	// seat reads to choose.
	enumhelp.Flag(c, flags.Type, record.SeatLogTypeEnum(),
		"what this entry asserts")
	// validate refuses an entry with no sentence; the marker says so where the seat reads.
	return SaysRequired(c, flags.Reason)
}

// Position is the chair's argument to the bench on what is about no one gap (record.PositionSittings
// says when one is owed). Only the chair's surface mounts it, and the write path refuses it from any
// other seat (record.PositionRefusal): a party argues a gap in its closing.
func Position() *cobra.Command {
	// validate refuses an empty position — "a duty discharged by nothing".
	return SaysRequired(Prose(NewKeyed("position", "position", Msg{}, func(s Context, cmd *cobra.Command) (Result, error) {
		text, err := Reason(cmd)
		if err != nil {
			return nil, err
		}
		if _, err := record.Append(s.Identity(), &recordpb.Position{Text: proto.String(text)}); err != nil {
			return nil, err
		}
		return Msg{Message: "position recorded"}, nil
	})), flags.Reason)
}

// Closing is a party's argument to the bench on one docketed gap: the gap's minting lens's, and
// blue-respond's. The write path refuses it from any other seat, and on a gap no unruled docket
// motion stands on (record.ClosingRefusal).
func Closing(key string) *cobra.Command {
	c := Prose(NewKeyed("closing", key, closingResult{}, func(s Context, cmd *cobra.Command) (Result, error) {
		text, err := Reason(cmd)
		if err != nil {
			return nil, err
		}
		if _, err := record.Append(s.Identity(), &recordpb.Closing{
			GapId: proto.String(Str(cmd, flags.ID)),
			Text:  proto.String(text),
		}); err != nil {
			return nil, err
		}
		return closingResult{GapID: Str(cmd, flags.ID)}, nil
	}))
	c.Flags().Var(flags.GapID().WithCheck(record.GapExists), flags.ID, "the gap id this closing argues")
	// validate refuses a closing that names no gap — "a receipt naming no gap cannot be audited".
	return SaysRequired(c, flags.ID)
}

// views are the projections a seat may read. `defaultFor` is the role whose default this view
// is; "*" means every role.
//
// EVERY SEAT DEFAULTS TO ITS PENDING WORK. It did not: blue's bare `show` returned `changelog`
// — a record of what blue had ALREADY done, handed to it before it had done anything — the lens
// got `citation-ledger` and the bench got `debate`. Asked what would tell them a sitting was
// finished, only the chair could name a mechanism; blue and the bench answered with another
// seat's future act ("red agrees it's sound"), which is not observable at the moment they have
// to decide to stop.
//
// THREE VIEWS ALSO CLAIMED "chair" AND THE LAST ONE SILENTLY WON, because the resolution loop
// keeps overwriting. A default decided by slice order is a default nobody chose.
// jsonByName marks a projection whose NATIVE form is already JSON, so `--json` on it is a
// no-op — the same bytes, not a second form (Show says why it stopped being refused).
//
// IT IS A FIELD BECAUSE IT WAS TWO HAND-KEPT LISTS. The fact lived in the `long` prose (each
// such view opens "STRUCTURED JSON:") and again in a switch case naming six views by hand, and
// nothing held them together. TestDebateJSONViewAndOneWayContract's own comment records what
// that costs: `friction` sat in a list like this after it stopped being a view, and the
// assertion went on passing because it demands an error and an unknown view is an error too — a
// stale name checking nothing while reading as coverage.
//
// MEASURED COST OF THE PROSE HALF (#593): six seats across two runs piped `show work --json`
// into python and crashed on KeyError: 'sitting'. They had not guessed the schema wrong — the
// schema is published and correct, and `sitting` is a real key on the bare projection. They read
// a description shouting STRUCTURED JSON, reached for the tool's --json flag, and got a REFUSAL
// whose envelope ({ok,code,error}) parses as cleanly as the projection does. The refusal's whole
// justification is that "a wrong guess fails loudly"; delivered to a pipeline it failed silently,
// as a missing key. So the refusal was dropped (--json on these views is the same bytes), and the
// note generated from this field says where the envelope can still arrive: on an error.
// shape is the ZERO VALUE OF THE TYPE THE PROJECTION MARSHALS, and it is a field for the same
// reason jsonByName is one. The key names a seat needs are a fact about that struct; written
// into `long` as prose they become a second copy with no writer who can refuse it, stale on the
// first field rename with every test still green. Held as the type itself, the help is
// GENERATED (jsonshape.Tree) and TestProjectionShapeMatchesEmittedKeys pins it against the keys
// the verb actually emits — the same behaviour-not-self-consistency rule
// TestJSONByNameMarkMatchesWhatTheBareViewEmits was written under.
//
// nil is a real answer, not an omission: a markdown projection has no field tree, and inventing
// one for it would document a shape no reader receives.
var views = []struct {
	name, short, long, defaultFor string
	jsonByName                    bool
	shape                         any
	// inquest moves this projection off `show` and onto the BENCH'S OWN GROUP.
	//
	// The bench rules between red and blue, so it may not take either side's account of what was
	// argued or contested — it reads the record itself. That is a different act from a seat reading
	// its own work, and it was under the same word: every seat saw `show debate` and `show motions`
	// in its listing, and measured across two runs the lenses and blue spent 24 calls in projections
	// that answer a question only the bench asks.
	inquest bool
}{
	{"report", "THE REPORT, as it stands now. `changes` says how it got that way. Written by the opening synthesis and blue's `edit`, with anchors from blue's `cite`, the lens's `finding` and `mint` and blue's `compute`", reportLong, "", false, record.ReportJSON{}, false},
	{"board", "EVERY GAP THE RUN HAS, yours or not — open and closed, with grades, fates and closure prose. `work` narrows this to what is yours and blocking. Written by the lens's `mint`, `close` and `regrade`, and blue's `retire`", "THE BOARD — open and closed gaps with grades, closures, anchors, observations and their fates, counts, and any replay anomalies. Each gap's `minted_by` is the seat that minted it, the one seat that may close or regrade it; its `found_by` is the findings it was minted from. `material` is whether the gap holds a PASS while it is open — its class's answer, else graded medium or above; `remanded` is whether the bench remanded it and it stands open with nothing pending, and `docket_direction` the direction the latest remand states, empty when the bench never remanded it. JSON by default; --format markdown gives the human-verification rendering. " + record.LocationStates + ". Written by the lens's `mint`, `close` and `regrade`, and blue's `retire`", "", true, record.BoardJSON{}, false},
	{"findings", "THE RAW LENS FINDINGS, BEFORE they are minted into gaps — several findings can become one gap, and this is where you see which. Written by the lens's `finding`", "Every lens finding on the record (id, area, seat, epoch, grades, location, text) — the minting lens coalesces these into gaps", "", true, record.FindingsJSON{}, false},
	{"work", "WHAT IS OPEN TO YOU, AND WHETHER YOU MAY STOP — your pending work, not the whole board. Handed to you at dispatch under a preamble that says so — except where your agent configuration seats several seats (blue-researcher's) or your dispatch was resumed, and then this is your first read. Every act you record says where you then stand, so you rarely need to ask again. Written by the lens's `mint` and `close` and the bench's `motion docket rule` and `declare`", viewLong("work", "", ""), "*", true, record.WorkJSON{}, false},
	{"motions", "WHAT HAS BEEN CONTESTED AND HOW IT WAS RULED — the ask in the filer's words, and the ruling if it has one. `debate` is what each side ARGUED; this is what was formally disputed. Written by `motion` and `rule`", "Every motion and its answer — id, subject, filer, the BASIS (the ask in the filer's words), and the ruling if it has one. An unruled motion blocks a PASS verdict, and this is the only way to read what it asks. Written by `motion <subject> file` and `rule`", "", true, record.MotionsJSON{}, true},
	{"debate", "WHAT EACH SIDE ARGUED, epoch by epoch — the transcript, in order. Written by the chair's `position` and `spot-check`, the lens's and blue's `closing` on a docketed gap (`### RED CLOSING` is the minting lens's, `### BLUE CLOSING` blue-respond's), a lens's `regrade` that keeps a gap's grade, and the bench's `motion docket rule`", "the transcript epoch by epoch (an epoch is one chair sitting), every seat's sections in order; --json gives the structured form below. Written by the chair's `position` and `spot-check`, the lens's and blue's `closing` on a docketed gap (`### RED CLOSING` is the minting lens's, `### BLUE CLOSING` blue-respond's), a lens's `regrade` that keeps a gap's grade (`### RED KEEPS THE GRADE`; `red_keeps` in --json), and the bench's `motion docket rule`", "", false, record.DebateJSON{}, true},
	{"changes", "HOW THE REPORT GOT THAT WAY — every edit and citation retire in record order, and with `--id <gap>` the fix red asked for beside the changes answering it. Written by blue's `edit` and `retire`", "every recorded change to the report, in record order — each edit, and each citation blue retired from a sentence the report keeps; add --id <gap> to put red's required_fix and the changes answering it SIDE BY SIDE — the comparison that replaces inferring whether a gap was fixed. Written by blue's `edit` and `retire`", "", false, record.ChangesJSON{}, false},
	{"evidence", "WHAT BACKS A CLAIM, AND WHAT RED MADE OF IT — the lookup table for an anchor you are holding while reading, and with `--anchor <id>` on a computation, its bundle's files and the output it recorded. Written by blue's `cite` and `compute`, and the lens's `verify` and `reproduce`", "WHAT BACKS THE REPORT, AND WHAT HAS BEEN CHECKED OF IT — every source keyed by the `" + anchor.Token("C-…") + "` anchor in the text (url, title, sha256, the sentence it backs, and `source_text_origin`: where its text came from). `source_text` is WHAT THE SOURCE SAYS where blue quoted it: blue's span, the same words as the source holds them (`matched`), and the source's own text `before` and `after` them, copied from the cached text when blue cited it — read the sentence against this. `blue_reason` is blue's account of why the source backs the sentence: blue's words, not the source's. A source the tool holds no text for (`source_text_origin` none) has no `source_text`. `work_status` is what a maintained index says about the WORK — `retracted` means the paper was withdrawn: the bytes are genuine, the fetch was sound, and no re-reading of the source can discover it, so a claim resting on it is a finding to file however well it reads. `not_checked` is not reassurance; it says nobody asked. `source_completeness` is which part of the work the copy is, on each source and each of red's verdicts — `abstract` means the citation or the verdict rests on the abstract alone, not on the study; `unverified` and `not_asked` say nobody could tell. A source with `pages` quotes OCR text — a machine's reading, which can misread — and `pages` holds the PDF page the tool found its span on: check the span against that page's image, not against the reading. Every computation keyed by its `" + anchor.Token("K-…") + "` anchor — the id the lens's `reproduce --anchor` takes — with its bundle's sha256, its command and red's checks of it, and red's verified claims with their confidence.\n\n" +
		"`copy_substitution` says whose bytes were read — on a source, the citation's; on a verification, the reading's: `none` — the url's own; otherwise a copy stood in for the url, and `copy_url` is where it was read — `followed_link`, the full text the url's page pointed to; `recovered`, a copy taken after the url refused; `fuller_copy`, another copy carrying more of the work; `named_backend`, a backend a fetch named. Read the copy, not the url, when you check what the citation read. `not_cached`, on a verification or a corroboration, is a source the run's fetch cache held no copy of when red recorded it, so which bytes it read is not known; `not_recorded` is an act from before the field.\n\n" +
		"A COMPUTATION'S VERDICT IS FLAT ON ITS ROW, IN TWO HALVES. `rerun` is the BUNDLE's: `reproduced`, `differs`, `calls_the_tool` (the bundle calls this tool, which refuses every call made while a checker runs it, so nothing it printed is a result), or `not_rerun` — one bundle can back several computations, and one re-run of it stands for every one of them. " +
		"`soundness` is THIS ANCHOR's sentence: `sound`, `unsound`, `not_judged`, or `reopened` (an edit changed the sentence after it was judged). `verified` is true when the bundle reproduced AND this sentence is judged sound. " +
		"A standing `unsound` from any lens reads `unsound` until that lens judges the sentence again or an edit reopens it; `checks` lists every judgement of the anchor, each marked `standing` or not. " +
		"`counts.computations_unverified` is how many computations are not verified: read it beside the rows, and a row-by-row reading that disagrees with it has read the wrong level. `counts.sources_unverified` is how many citations no reading of the sentence the report carries has reached: unchecked, or checked and nothing read. `counts.sources_refuted` is how many citations red finds against NOW: a `refutes` or `absent` check of the sentence the report carries that still stands. A refutation of a sentence an edit has since replaced, or that a later check of the citation backs, or of a retired citation, is in the rows and not in the count.\n\n" +
		"`unanswered_contradictions` IS EVERY CONTRADICTION RED RECORDED THAT NOTHING HAS ANSWERED: a source a lens read refutes, or is silent on, a sentence of the report. Each row carries its `id`, the `claim` the reading was of, the `reader` who owes the finding and its gap, and the citation `anchor` the reading checked (empty for a corroboration). " +
		"Only a finding naming it — `--about-kind contradiction --about <id>` — raises it, whether or not the report still holds the sentence; a finding whose quote contains the claim raises nothing. `raised_by` names the raising finding, and a source read against the claim after it opens a new one. The gap minted from the raising finding, naming it in its `found_by` as the finding that surfaced it, answers it; the finding alone does not. A row with an `anchor` is answered as well when blue retires that citation, or once a lens has verified the citation's sentence as it now stands and every lens's latest verify of it backs it: the reading stops counting when its lens reads the citation's source again or an edit changes the sentence after its sitting opened, and another lens's backing verify while it counts is a disagreement that answers nothing. A PASS waits while a row stands.\n\nAn `unchecked` ROW IS A CITATION OR COMPUTATION WITH NO CHECK OF THE SENTENCE THE REPORT CARRIES: nobody has checked it, or an edit changed its sentence and no sitting opened since has checked it. `edit` and `last_check` are event ids: the edit (0 where none changed the sentence), and the anchor's latest check (0 where it has none), which does not clear it. A `verify` of the citation that reads its source, or a `reproduce` naming the computation's anchor — from a sitting opened after `edit`, where there is one — removes the row, whichever lens records it. A verify `--as unreachable`, or one of a copy the tool stamped not_the_work, reads nothing: it removes the row and leaves the citation on a failed check, which holds the PASS as the row did. Its owning lens owes that check and is the seat readied for it: a citation is the evidence lens's; a computation is the computation lens's. A PASS waits while a row stands.\n\nTO READ ONE ENTRY IN FULL, add --anchor <id>. On a computation's `" + anchor.Token("K-…") + "` id that is the computation's row WITH THE COMPUTATION ITSELF: `files` is the stored bundle, each file by its `path` with its `body` — the bytes the lens's `reproduce` executes with the row's `argv` — and `output` is what this computation's run printed. The table carries no program, so THIS IS HOW A PROGRAM IS READ BEFORE IT IS JUDGED. A bundle or output the computation store does not hold is refused, never shown empty. On a source's `" + anchor.Token("C-…") + "` id it is that source's one row. Either carries the anchor's `unchecked` row while one stands.\n\nTHIS IS HOW YOU RESOLVE AN EVIDENCE ANCHOR you are reading in the report; a `" + anchor.Token("F-…") + "` is a finding's, which `show findings` resolves, and a `" + anchor.Token("G-…") + "` is a gap's, which `show board` resolves. Written by blue's `cite` and `compute`, and the lens's `verify` and `reproduce`", "", true, record.EvidenceJSON{}, false},
	{"inquiries", "WHICH INQUIRIES THE RESEARCH TOOK UP AND WHAT BECAME OF EACH — every inquiry with its current fate and the reason for it. Written by blue's `inquiry propose` and `inquiry record` and the chair's `motion inquiry rule`", "the exploration space, read-only. An inquiry is one question the research could take up, proposed by blue with the hypothesis that would be true if it paid off; this lists every one with its question and its hypothesis, its current fate, and the reason and method recorded with that fate. The fate is the `status` key, one of " + strings.Join(record.Names(record.MustEnum("inquiry", "status").Values), ", ") + " — the meaning of each is on `inquiry record`'s page. The bare call groups them by fate, each with the path its fate took (every fate recorded on it, in order, with the epoch it was recorded in) and the chair's ruling, then lists those awaiting a decision: still at an open fate, `proposed` or `underway`, with nothing recorded on it this epoch (the current chair sitting) — each owes a fate blue records with `inquiry record`. --json gives the same inquiries with their types intact, and no awaiting list: `reason` and `method` are the latest fate's, `epoch` the epoch it was recorded in, `history` each fate it took, in order, as `e<epoch> <fate>`, `seat_id` the seat that recorded the latest, and `ruling` the chair's ruling from `motion inquiry rule` (`endorsed`, `out_of_scope` or `too_thin` — an argument that changes no fate), empty while none. Red reads the awaiting list as decisions blue has not made. Written by blue's `inquiry propose` and `inquiry record` and the chair's `motion inquiry rule`", "", false, record.InquiriesJSON{}, false},
	{"telemetry", "HOW THE NUMBERS MOVED ACROSS EPOCHS — a trend, not a snapshot: one line per epoch (chair sitting), and the signal the STOPPING judgment reads. Computed from the record, so no verb fills it", "JSONL, one line per epoch (chair sitting): the trend the STOPPING judgment reads — the bench's signal for whether the findings are still changing character or merely recurring", "", true, view.TelemetryLineShape(), true},
}

// workPage is `show work`'s page, one paragraph a row, each with the roles whose list it describes.
//
// THE LIST IS BUILT PER ROLE, SO ITS PAGE IS TOO. record.WorkOfSeat states a mint budget and the
// regrades owed on a lens's list alone, `engaged` on blue's, and an item another seat clears on the
// chair's; a paragraph about a field the reader's list never holds is text it cannot act on, read at
// every sitting's opening act. A row naming no role is every seat's. The page a seat's `--help`
// prints and the copy scripts/agentgen inlines into its constitution are one rendering (viewLong),
// and a tree built for no role — the operator's reading — holds every row.
//
// WHY THREE OF ITS FIELDS EXIST, WHICH THE PAGE DOES NOT ARGUE. The blue-answered item says "cannot
// yet say" because blue still sitting and blue having said nothing are the same silence on the
// record and want different acts from the lens — one is waiting, the other is a fact to state
// (record/blueanswered.go). `yours_to_close` answers who may close a gap so no seat decodes it from
// `found_by`. `estopped` is listed because re-raising a ruled gap is relitigation, not diligence.
var workPage = []scopedPara{
	{text: "**WHERE YOUR DISPATCH HANDED YOU THIS LIST — it arrives under a preamble that says so — YOU ARE PROBABLY NOT MEANT TO RUN THIS: every act you record answers `may I stop` for you afterwards, so reach for it only when neither reached you, or when you want the items that do NOT block you. A seat of blue-researcher's configuration, which seats several seats, and a resumed dispatch are handed no list, and read it first.** EVERYTHING OPEN TO YOU, in one list. `sitting.open` is every work item, each with `blocks` (whether it stops you closing); `sitting.complete` is true exactly when nothing blocking is left."},
	{roles: []string{"lens", "blue", "bench"}, text: "A BLOCKING ITEM HOLDS YOUR SITTING: the call that returns your envelope is refused while any blocking item your own act clears stands, and the refusal names each — clear them, then return it again. An envelope the record cannot be read to check is refused as well, naming the read error. An item your sitting cannot clear is listed with `blocks: false` and says why — another seat clears it, or a later sitting does."},
	{roles: []string{"chair"}, text: "ON THE CHAIR'S LIST a blocking item holds a PASS, not your sitting — your envelope is refused on none of them — and one with `another_seat_clears` is the board's state: another seat's act clears it, and it is not the chair's to clear. No other seat's list holds one. An envelope the record cannot be read to check is refused, naming the read error."},
	{text: "An item with `blocks: false` is work nobody will refuse you for skipping — a citation nobody verified or a computation nobody re-ran (each on its owning lens's list alone: a citation on the evidence lens's; a computation on the computation lens's), a source blue never cited, an inquiry never revisited, a grade you could move, a motion you could file. IT IS STILL YOUR WORK: `complete: true` with items open means the gates are satisfied, NOT that nothing is left."},
	{roles: []string{"lens"}, text: "FOR EACH GAP YOU MINTED AND LEFT OPEN, an item says whether blue ANSWERED it — and where the record cannot yet say, it says that instead of reporting no answer."},
	{roles: []string{"blue"}, text: "A GAP WHOSE LENS KEPT ITS GRADE after your last answer has an item carrying the lens's reason — what its own check still finds failing — until you answer the gap again."},
	{roles: []string{"lens"}, text: "A LENS'S LIST STATES ITS MINT BUDGET, and it is the count a mint is refused on: `sitting.mint_budget` is the gaps you minted that are `open` now, your `budget` — the most you may hold open at once: the larger of the run's `floor` and `mints` for every `per` of the `units` your area is sized by, named in `unit` — and the mints `left` to you. A gap of yours that closes frees one. A gap minted from a finding of yours that raised a contradiction takes none: it is admitted at the cap, and `answering` counts those you hold open outside `open`. `measured: false` says the units cannot be read, and `budget` and `left` are then the floor's. An item states the same numbers; decide gap or finding from it. Absent on every other role's list."},
	{text: "`open` holds OPEN gaps only, and each carries WHAT IT TAKES TO ACT ON IT: the grades, class and location, the WHOLE problem (`problem_synopsis` is the first 140 characters, for scanning a long list), the `required_fix` and the `acceptance_check` you will be re-audited against, `minted_by` and `yours_to_close` — only the seat that minted a gap may close or regrade it. YOU SHOULD NOT NEED THE BOARD TO ACT ON YOUR OWN WORK. " + record.LocationStates + "; `edited_since` lists the edits that changed a gap's sentence since you last sat. One with `remanded` was REMANDED by the bench, and `remand_stage` says where a remand that sends the gap back stands: one ruled at impasse, or any remand of a gap that is not material. `owed`: the gap is back in the debate for ONE more exchange — the dispatch readies its minting lens and blue on it, owing the research direction the ruling states (`docket_direction`), and an item on each of their lists says so, holding that seat's sitting until it records an act on the gap. `spent`: that exchange has been had — it begins when blue acts on the gap, and the minting lens's act alone begins nothing; if it leaves the gap at impasse, the dispatch dockets it for the bench again, to be closed, and if it moved the gap, its limits count afresh from it. No `remand_stage`: no remand was ruled while the gap was at impasse, and no remand of a gap that is not material is owed."},
	{roles: []string{"lens"}, text: "AN ACCEPTED GRADE MOTION WAITS FOR THE MINTING LENS'S REGRADE, and the gap says which: `regrades_owed` is each grade motion on the gap that was ACCEPTED and that no regrade has followed — the `motion_id`, the `dimension` it contested and the `proposed` grade the ruling accepted — on the minting lens's list, and empty on every other. The item in `sitting.open` that names the gap states the same motions."},
	{roles: []string{"bench"}, text: "ON THE BENCH'S LIST an unruled motion whose gavel is yours blocks where this sitting is convened to rule it: at a docket sitting, every docket motion on a gap it is convened on — a gap can hold more than one; at a petition sitting, every petition filed before it opened; at the terminal and assembly sittings, every one. Any other waits for the sitting convened for it and is listed without blocking. Every one holds the run's outcome: the run cannot end in a verdict while it stands."},
	{roles: []string{"bench", "chair"}, text: "THE LIST CLOSES WITH THE RUN'S LAW: whether setup staged a law corpus for this run, and where it is — the law you read before you rule. It never blocks."},
	{roles: []string{"lens", "blue"}, text: "WHERE THE LENS PRESCRIBED EXACT TEXT, the gap has it as a quoted span: `fix_old` and `fix_new` are the proposal over the report as it stands — the pair blue's accepting `edit` sends — and `fix_basis` says what the fix rests on: `verified`, the tool checked the span against the report, or `proposed`, prose only, and then the gap has no pair."},
	{roles: []string{"blue"}, text: "`engaged` IS BLUE'S RESPONSE SITTING'S, and null on every other seat's list: `gap_ids` are the gaps its dispatch named — the latest dispatch naming blue, before blue registers for it as well as after — and `found_closed` is each of them its lens closed after that dispatch and before blue sat, with the seat that closed it, the `fate` that close gave, and its `successors` — each open gap the record names as carrying it forward (its close's `superseded_by`, or a mint naming it in `supersedes`), empty when none. A gap there owes no answer and no manifest row. `successors_owed` is the open successors no dispatch row of this epoch engages blue on yet — exactly the set `engage` names, empty once it is recorded — and blue sits on them a second time this epoch. An empty `found_closed` says every engaged gap was open when blue sat."},
	{text: "`estopped` IS WHAT YOU MAY NOT RE-RAISE: the gaps the BENCH ruled, each with id, location, class and the `fate` that ended it. New evidence against one is a lineage successor, minted under a new id naming the ruled gap in `supersedes` and saying what the ruling did not account for. YOUR OWN closures are not here and are not a bar: red may reopen what red closed, and `show board` shows you those with `closed_by` at the moment you are deciding reopen-or-new."},
	{text: "`declarations` ARE THIS RUN'S BENCH HOLDINGS, every one on the record, in the order the bench declared them: each `holding` as the bench recorded it, and the `epoch` it was declared in. A holding binds how you read the record. `gaps` is each gap the bench declared the holding about, with where the BOARD says that gap stands as this list is read — `status` is `open`, or `closed` with the `fate` that closed it, `closed_by` — `bench` or `red` — and `closed_epoch`. A gap the holding's text names and the bench did not declare it about is not among them. The tool places the status beside the prose, and the prose is the bench's and does not move when a gap does: WHERE A HOLDING AND ITS `gaps` DISAGREE ABOUT A GAP, `status` IS WHERE THE GAP STANDS. Every declaration is listed, because the record cannot say which one a later declaration replaces — a later one that does says so in its own text."},
	{text: "Fate defect_owed_elsewhere means still broken and NOT yours to fix; repaired_with_regression means a live successor exists. Written by the lens's `mint` and `close` and the bench's `motion docket rule` and `declare`"},
}

// scopedPara is one paragraph of a projection's page. roles names the roles whose page carries it;
// none means every role's.
type scopedPara struct {
	roles []string
	text  string
}

// reportLong is the report's page for every seat.
var reportLong = "THE REPORT, as red audits it and blue amends it; add --anchor <id> to read just the passage AT one anchor (with its section and line numbers) rather than the whole document. Anchors are shown AS THEY ARE: blue's `edit` refuses an edit that drops one, so a token inside the span you are replacing is yours to carry into --new. TO LOOK ONE UP rather than carry it: `show findings` resolves `" + anchor.Token("F-…") + "`, `show evidence` resolves `" + anchor.Token("C-…") + "` and `" + anchor.Token("K-…") + "`, and `show board` resolves `" + anchor.Token("G-…") + "`. --match and --quote select lines by their text WITH THE ANCHORS OUT, so a pattern crosses an anchor and never names one: to read at an anchor, --anchor is the selector. Written by the opening synthesis and blue's `edit`.\n\n--json gives the same lines structured: `lines` is every line you asked for — the whole report, the window under --anchor, the matching lines under --match or --quote — each with its `line` number and the section `heading` it sits under, and `head` is the report head they were read at. Under --match or --quote, `selection` says how many lines matched and of how many. The bare call is the form to READ; this is the form to select from.\n\n" + report.AssembledNote

// reportPage is the report's page: every seat's, and the lens's paragraph on what its whole read
// records (record.ServeReport).
var reportPage = []scopedPara{
	{text: reportLong},
	{roles: []string{"lens"}, text: "A LENS'S WHOLE READ IS ON THE RECORD. The bare form and --json with no selector record that the whole report was served to you at its head — once per head in a sitting; evidence the text was served, not that it was read — and print it only once that is recorded: a read the record cannot take is refused, never served unrecorded. That head is your pin, the head you are readied past, once you have sat the sitting through; --anchor, --match, --quote and --assembled serve part of the report or none of its text and record nothing, so a sitting that ends without the whole report served to you at the head leaves you ready at it. Being a record, a whole read registers you for the sitting where you have not registered, so a `register` after it is refused as a repeat."},
}

// scopedPages are the projections whose page is scoped by role, by name.
var scopedPages = map[string][]scopedPara{"work": workPage, "report": reportPage}

// viewLong is a projection's page for one role: the paragraphs of its scoped page that are that
// role's, in order, or `long` where the projection has none. role "" is no seat's tree and gets
// every paragraph — which is what the views table holds as `long`.
func viewLong(name, long, role string) string {
	scoped, ok := scopedPages[name]
	if !ok {
		return long
	}
	var out []string
	for _, p := range scoped {
		if role == "" || len(p.roles) == 0 || slices.Contains(p.roles, role) {
			out = append(out, p.text)
		}
	}
	return strings.Join(out, "\n\n")
}

// ViewNames is the projection vocabulary — the single source behind the help text, the
// unknown-view error, and (exported for this reason) the gate that asserts every `--view`
// an agent-facing surface NAMES actually exists. See cli.ViewNames.
// showGroup resolves the `show` GROUP from wherever a refusal is raised, so the listing a seat
// gets is the one cobra renders for that group — not a second copy with its own column widths.
//
// A refusal may be raised from the group itself (bare `show`) or from one of its view
// subcommands, and only the group's help carries the projection list.
func showGroup(c *cobra.Command) *cobra.Command {
	if c == nil {
		return c
	}
	if c.Name() != "show" && c.Parent() != nil && c.Parent().Name() == "show" {
		return c.Parent()
	}
	return c
}

// GroupOf is the command group a projection lives under — `show` for a seat's own reads, `inquest`
// for the bench's raw-record ones. Derived from the table so a caller cannot hold a stale opinion
// about where a view moved to.
func GroupOf(view string) string {
	for _, v := range views {
		if v.name == view && v.inquest {
			return "inquest"
		}
	}
	return "show"
}

// JSONByNameViews are the projections whose native form is already JSON. Derived from the
// table, so the refusal below and the warning in each description cannot name different sets.
func JSONByNameViews() []string {
	out := []string{}
	for _, v := range views {
		if v.jsonByName {
			out = append(out, v.name)
		}
	}
	return out
}

// shapeHelpSuffix appends the projection's FIELD TREE, generated from the type it marshals.
//
// IT IS HERE SO THE SEAT NEVER HAS TO DISCOVER IT AT RUNTIME. Measured across one debate run:
// 25 `list(d.keys())` key-dumps and 251 `json.load` re-parses against three uses of `jq`
// (#684 F7). A seat that does not know the key names spools the projection to a scratch file
// and reaches for python — three tool calls, three hook fires and three ~20s round-trips to
// answer one question. The names are a fact about a Go struct, and this is the page the record
// clause already REQUIRES the seat to open before it runs the verb, so the answer arrives
// before the question costs anything.
//
// A view with no shape says nothing. Rendering `{}` for a markdown projection would claim it
// emits an empty object, which is a different and false statement.
func shapeHelpSuffix(name string, jsonByName bool, shape any) string {
	tree := jsonshape.Tree(shape)
	if tree == "" {
		return ""
	}
	// WHICH FORM THE TREE DESCRIBES IS PART OF THE TREE. `debate` is markdown bare and JSON only
	// behind --json; printing its field tree unqualified tells a seat the bare call returns an
	// object, and the bare call returns prose. The fact is already a field, so the sentence is
	// generated from it rather than remembered per view.
	unit := "OUTPUT (JSON)"
	switch {
	case name == "telemetry":
		// JSONL: the tree describes ONE LINE, and a reader parsing all of stdout as a single
		// document gets nothing — the same plausible zero the JSON-by-name warning names.
		unit = "OUTPUT (JSONL — one such line PER EPOCH, not one document)"
	case !jsonByName:
		unit = "OUTPUT (JSON, with --json — the bare call is the markdown form)"
	}
	return "\n\n" + unit + ": " + tree
}

func jsonByNameHelpSuffix(jsonByName bool) string {
	if jsonByName {
		return jsonByNameWarning
	}
	return ""
}

func viewIsJSONByName(name string) bool {
	for _, v := range views {
		if v.name == name {
			return v.jsonByName
		}
	}
	return false
}

// jsonByNameWarning is appended to every JSON-by-name view's help, at the point a seat forms the
// guess about which flag gets it JSON.
//
// It used to warn that --json was REFUSED and that the refusal parsed like data, because six seats
// had piped that refusal into python and crashed on a missing key. The refusal is gone (Show says
// why), but the trap it described is not: under --json ANY error on this view is a JSON envelope on
// stdout, which a pipeline reads as data unless it checks `ok`. So the note says both halves — the
// flag is harmless on success, and the error it can still get parses — rather than "changes
// nothing", which an independent review measured false on the error path.
var jsonByNameWarning = "\n\nTHIS PROJECTION IS ALREADY THE JSON: --json is accepted and, on success, byte-for-byte the same. " +
	"On an ERROR it prints a JSON envelope on stdout — one object whose keys are " + strings.Join(feov.RefusalKeys(), ", ") +
	", with `ok` false — so a pipeline must check `ok` before reading keys."

// ShapeOf is the value a projection's documented OUTPUT tree is generated from, or nil for a view
// with no JSON form. Exported so a test can hold what the view EMITS to that same type — the
// help and the check read one declaration.
func ShapeOf(view string) any {
	for _, v := range views {
		if v.name == view {
			return v.shape
		}
	}
	return nil
}

func ViewNames() []string {
	out := make([]string, 0, len(views))
	for _, v := range views {
		out = append(out, v.name)
	}
	return out
}

// Show prints a projection to STDOUT.
//
// READS BELONG IN THE TOOL. Every seat could already WRITE through the tool while still
// having to READ the run by opening markdown files at paths it learned from a prompt —
// which is how a seat comes to trust a hand-written artifact over the event log, and how
// the two came to disagree. A seat that asks the tool gets the answer the events support.
//
// It does NOT re-derive the markdown. It renders through the exact path `render` uses and
// then prints the resulting file, so the bytes on stdout and the bytes on disk cannot
// differ. A second renderer would be a second reader of one artifact, which is the defect
// class this whole tool exists to remove — writing one to serve reads would reintroduce it
// at the read surface.
// Show is a GROUP, not a verb with a flag.
//
// It was `show --view <name>`, and cobra models commands and flags — never a flag's VALUE
// space. That is the same undiscoverability the motion collapse fixed one layer up: a value has
// no --help of its own, no completion, and no place to say what it is for, so every projection's
// description had to be crammed into one flag's usage string. Making `--view` optional (a bare
// `show` now answers with the seat's pending work) made it worse rather than better: the flag
// became something a seat had no reason to discover at all.
//
// As subcommands each projection is a first-class thing: `show board`, `show motions`, its own
// --help, its own completion, and an unknown one gets the refusal that lists the whole surface
// rather than a flag-parse error naming the value.
//
// role scopes each page to its reader (viewLong); "" builds every role's paragraphs.
func Show(role string) *cobra.Command { return viewGroup(false, role) }

// Inquest is the BENCH'S OWN READ SURFACE, and it is a group rather than three more entries on
// `show` because it answers a different question.
//
// A seat reads `show` to learn what is open to IT. The bench reads the record to adjudicate
// between two parties, neither of whose account it may take: what each side actually argued, what
// was formally contested and how it was ruled, and how the numbers moved. Those are the raw
// record, not a digest of it, and needing them is what distinguishes ruling from working.
//
// IT IS MOUNTED FOR THE SEATS THAT ADJUDICATE — the bench and the chair — and for no others.
// Under `show` every seat saw these in its listing, and a name in a listing is a thing a seat
// reaches for. A lens and blue are the parties being judged; the transcript of what each side
// argued, the register of what was contested, and the convergence trend are not their work.
func Inquest() *cobra.Command { return viewGroup(true, "") }

func viewGroup(inquest bool, role string) *cobra.Command {
	if inquest {
		return inquestGroup()
	}
	c := &cobra.Command{
		Use: "show",
		// THE NAME LIST CAME OFF THIS LINE when the root page started marking which entries hold
		// commands. Ten projection names in a parent's listing is a slice of the surface handed
		// out where the seat cannot yet see what any of them is FOR — and a seat that reads a
		// name here and never opens the page has learned a word, which is how one of them typed a
		// projection name as a verb. The group heading now says this page hides commands; each
		// projection says what it is for on its own page, which is where the choosing happens.
		Short: "read a projection of the record — the tool is the read path, and the .md files are for human verification. Bare, it answers with YOUR PENDING WORK",
		// LONG CARRIES THE CONSUMPTION RULE, and it is on the GROUP because that is the one page
		// where a projection is CHOSEN — the leaf pages answer "what is in this one", this page
		// answers "how do I read any of them". Set explicitly rather than left to fall back to
		// Short, which is written for the root listing.
		//
		// IT IS HERE AND NOT IN THE SEAT PROMPT. The prompt-size ceiling asks of new text: does
		// it teach a VERB, a FLAG, a REFUSAL, or a RENDERING? This teaches how to consume a
		// rendering, so it belongs where the seat reads while choosing. The first draft of this
		// repair put it in debate.js's recordClause and was refused twice over — once for naming
		// a command invocation in a prompt, once for pushing judge past its ceiling. Both
		// refusals were right, and this is where they pointed.
		Long: "read a projection of the record — the tool is the read path, and the .md files are for human verification. Bare, it answers with YOUR PENDING WORK.\n\n" +
			"PIPE A PROJECTION, DO NOT SPOOL IT. Most are JSON on stdout (each page says which), and `jq` is a REQUIRED tool of " +
			"this suite, so it is installed here: piped into it, a projection answers your question in ONE call; spooled to a " +
			"scratch file and re-read in python it takes THREE — the dump, a key-dump to learn the field names, the extract — " +
			"each a process spawn and a round-trip. A spool you truly need goes in the run directory's scratch/, under a name " +
			"starting with your seat id.\n\n" +
			"YOU DO NOT NEED TO DISCOVER THE FIELD NAMES. Every JSON projection's page ends with its OUTPUT field tree — the " +
			"key names in their nesting, generated from the type the tool marshals. Read it before you run the view.",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		// Its bare form is the seat's pending work — a capability, not a refusal.
		Annotations: map[string]string{"bare-is-a-capability": "yes"},
	}
	// An unknown projection gets the SURFACE, not a parse error. Cobra's default answers
	// `unknown command "x" for "feov-record blue show"` and stops — at the one moment a seat is
	// definitively looking for what exists, which is the same argument that makes the root teach.
	c.Args = cobra.ArbitraryArgs
	c.RunE = func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			return RefuseAndTeach(cmd, feov.Validation, fmt.Sprintf(
				"no projection named %q — this reads one projection at a time, and each is listed with the verb that fills it.%s", args[0], NamedElsewhere(cmd, args[0])))
		}
		// The seat's own default — its pending work and whether the sitting is finished.
		return renderView(cmd, "")
	}
	for _, v := range views {
		if v.inquest {
			continue // the bench's own group; see Inquest
		}
		v := v
		sub := &cobra.Command{
			Use: v.name,
			// SHORT IS THE MENU, LONG IS THE INSTRUCTION, and they are separate fields because a
			// seat reads them at two different moments.
			//
			// They were ONE string. Cobra prints Short in the parent's listing and, with no Long
			// set, prints it AGAIN as the whole body of the leaf page — so `show --help` carried
			// all eleven essays (6,615 bytes, the largest page in the system) and every
			// `show <view> --help` beneath it repeated one of them verbatim. A seat that read the
			// group page had already read every leaf page, which is why `pagesNeverUsed` came back
			// at 11 of 13 for a seat that opened its whole tree: the pages beneath held nothing new.
			//
			// SHORT IS WRITTEN TO BE READ AS A SET, not alone. Eleven of them arrive on one page
			// and the seat's question there is which ONE to open, so each says what question it
			// answers and how it differs from its neighbours — `work` narrows `board`, `changes`
			// explains `report`, `motions` is what was disputed where `debate` is what was argued.
			// Each still names the verb that fills it: viewnaming_test asks that of `show --help`,
			// which renders Short, and a projection whose writer is unnamed teaches a name that
			// does not work.
			Short:        v.short,
			Long:         viewLong(v.name, v.long, role) + jsonByNameHelpSuffix(v.jsonByName) + shapeHelpSuffix(v.name, v.jsonByName, v.shape) + atAnchorShapeSuffix(v.name) + assembledShapeSuffix(v.name),
			Args:         cobra.NoArgs,
			SilenceUsage: true,
			RunE:         func(cmd *cobra.Command, _ []string) error { return renderView(cmd, v.name) },
		}

		// ONE COMMAND, BOTH FORMS. `ledger` and `archive` were markdown-only views of data the
		// board JSON already carries whole — three names for one projection, which is the alias
		// problem this vocabulary bans everywhere else. --format is the flag `graph` already
		// uses for the same question.
		if v.name == "board" {
			sub.Flags().String(flags.Format, "json", "json (the form a seat acts on) | markdown (the human-verification rendering: open gaps, then the closure archive with its prose; --match and --quote are refused beside it)")
		}
		// READING AT AN ANCHOR, rather than pulling the whole document to check one sentence.
		// The window is addressed by anchor because a line number is a fact about a rendering —
		// see internal/anchor/window.go for why that distinction is load-bearing here.
		// EVERY VIEW TAKES THE SAME SELECTOR PAIR (gblock: "everywhere you have one selector you
		// probably want them all"). It began on three — `report`, `board` and `changes` — which left
		// the rest with no working way to ask for part of a projection while all of them advertised
		// `--id`, which only `changes` honours. A view now gains selection by being in this table,
		// not by anyone remembering to add a case.
		//
		// THE NOUN COMES FROM THE VIEW so the help reads in the seat's terms ("only the gaps matching",
		// "only the edits matching"); `entries` is the honest default for a projection whose rows have
		// no better word, rather than a name invented per view in a list that would go stale.
		//
		// TELEMETRY TAKES NO SELECTOR: it is a series of JSONL lines, one per epoch, with no document
		// whose arrays hold entries and nowhere to state what a selection kept — so the pair it carried
		// returned every line for every pattern. Selecting epochs out of a trend is not a read it has.
		if v.name != "telemetry" {
			// A MARKDOWN VIEW SELECTS ITS --json FORM, and says so on the line of the flag a seat is
			// about to type, as the report's page does for its own bounds. Said in the noun, before the
			// explanation every view's pair closes with, so the manual still lifts that explanation once.
			what := selectorNoun(v.name)
			if !v.jsonByName && v.name != "report" {
				what += markdownSelectorBound
			}
			AddSelectorFlags(sub, what)
		}
		// --id IS DECLARED ON THE ONE VIEW IT SCOPES. A flag printed on a page is a flag that works
		// there: declared on the group, it stood in every view's `Global Flags` block beside the
		// run and the seat, and two lenses in one run typed `show board --id <gap>` off that block.
		// On any other view it is a flag the view does not have, and that refusal names the views
		// and verbs on the seat's surface that do take it (RefuseUnknownFlag).
		if v.name == "changes" {
			sub.Flags().Var(flags.GapID().WithCheck(record.GapExists), flags.ID,
				"scope this projection to one gap — red's required_fix beside the edits and citation retires answering it. No other view is scoped by a gap: for part of any projection by its TEXT, --match (a regex) or --quote (a literal, the same span the acting verbs take)")
		}
		if v.name == "report" {
			// WHICH SELECTORS COMBINE IS SAID BESIDE EACH ONE, because a seat reads the line of the
			// flag it is about to type and not the page above it: in one run --window was typed
			// without --anchor twice and --match beside a bare --assembled three times, each a
			// refused call. The refusals below (renderView) say the same things at the call.
			sub.Flags().Var(flags.AnchorID(), flags.Anchor,
				"read the report AT one anchor `id` (F-…, C-…, K-… or G-…) rather than whole — you get the LIVE text there, its section heading, and line numbers to quote back. Takes --window and --json beside it and no other selector: --match, --quote and --assembled are each refused with it")
			sub.Flags().Int(flags.Window, anchor.DefaultWindow,
				"ONLY with --anchor, and refused without it: how many paragraphs of content either side of the anchor (blank lines are kept, not counted)")
			sub.Flags().Bool(flags.Assembled, false,
				"read the ASSEMBLED report set rather than the report's text: every section in shipping order, who composes it and how content reaches it — a listing WITHOUT the sections' text, which the --json form carries. With --json, --match and --quote select its sections; without --json they are refused beside it, and --anchor and --window are refused beside it in either form")
			for _, name := range []string{flags.Match, flags.Quote} {
				sub.Flags().Lookup(name).Usage += reportSelectorBound
			}
		}
		// ONE ENTRY, BY THE TOKEN THE SEAT IS HOLDING. The same flag `report` takes for the same
		// id — one declaration (flags.AnchorID), so both refuse a malformed id in the same words
		// and both take a token pasted whole: there it reads the text around the anchor, here it
		// reads what the anchor points at.
		if v.name == "evidence" {
			sub.Flags().Var(flags.AnchorID(), flags.Anchor,
				"read the ONE entry at this anchor `id` rather than the table — a computation's (K-…) with its bundle's files and recorded output, or a source's (C-…)")
		}
		c.AddCommand(sub)
	}
	return c
}

// atAnchorShapeSuffix is the field tree of the evidence view's read at one anchor, generated from
// the type it marshals for the reason shapeHelpSuffix gives: the key names are a fact about a Go
// struct, and the page a seat opens before the read is where they cost nothing.
func atAnchorShapeSuffix(name string) string {
	if name != "evidence" {
		return ""
	}
	return "\n\nOUTPUT WITH --anchor (JSON — `source` or `computation`, whichever the anchor names): " + jsonshape.Tree(record.EvidenceAtJSON{})
}

// reportSelectorBound closes --match and --quote on the report's page with what each combines
// with there: the pair is every view's, and only this view has an anchor read and an assembled form.
const reportSelectorBound = ". Refused beside --anchor. Beside --assembled it selects SECTIONS, and only with --json"

// markdownSelectorBound qualifies what --match and --quote select on a view whose bare form is markdown.
const markdownSelectorBound = " of the --json form (the bare form is markdown and refuses this flag)"

// assembledShapeSuffix is the field tree of the report view's assembled layout, for the same reason.
//
// THE TWO JSON FORMS SPELL `heading` DIFFERENTLY, and the page says so where the field is listed:
// a line of the report's text carries the heading line it sits under as written, hash marks and
// all, and a section of the assembled set carries its name. A seat that selected on one form's
// spelling in the other got nothing back, twice in one run.
func assembledShapeSuffix(name string) string {
	if name != "report" {
		return ""
	}
	return "\n\nOUTPUT WITH --assembled --json: " + jsonshape.Tree(report.LayoutJSON{}) +
		"\n\n`heading` IS SPELLED TWO WAYS. In --assembled --json it is the section's bare name — `Research areas`, no `#` marks. " +
		"In the --json `lines` it is the heading line as the report writes it — `## Research areas`." +
		"\n\nEVERY SECTION CARRIES `text`, a string: a section that does not ship carries `\"\"` and `no_text`, which says why it has none."
}

// evidenceAt reads the one evidence entry at an anchor, and for a computation joins the artifact the
// record addresses: the bundle's files and this act's output, from the store.
//
// THE STORE IS READ HERE AND NOT IN THE RECORD PACKAGE. `EvidenceAt` answers from events; the
// bundle is content the events point at, and a view of the record that opened files would make
// every narrowed read of it depend on a directory.
func evidenceAt(run record.Run, id string) ([]byte, error) {
	at, err := record.EvidenceAt(run, id)
	if err != nil {
		return nil, err
	}
	if at.Computation != nil {
		files, output, err := compute.Stored(run.Dir(), at.Computation.Sha256, at.Computation.Anchor)
		if err != nil {
			return nil, err
		}
		for _, f := range files {
			at.Computation.Files = append(at.Computation.Files, record.EvidenceFileJSON{Path: f.Path, Body: string(f.Body)})
		}
		at.Computation.Output = string(output)
	}
	b, err := json.MarshalIndent(at, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(b, '\n'), nil
}

// renderView writes one projection. want == "" means this role's default.
func renderView(cmd *cobra.Command, want string) error {
	role := roleOf(cmd)
	// RESOLVED, NOT READ OFF THE FLAG. The engine injects FEOV_RUN and every WRITE verb
	// honours it through Begin/Of — reads did not, so a seat that correctly omitted --run
	// could record all sitting and then be told its board did not exist. Measured with the
	// identity injected: register and two write verbs all succeeded; `show`
	// demanded the flag.
	run, rerr := Of(cmd).RequireRun()
	if rerr != nil {
		return rerr
	}
	if want == "" {
		// FIRST MATCH WINS, not last. The loop used to keep overwriting, so three views
		// claiming "chair" resolved by slice order — a default nobody chose.
		for _, v := range views {
			if v.defaultFor == role || v.defaultFor == "*" {
				want = v.name
				break
			}
		}
	}

	// --json on a read opts into the STRUCTURED form of a view whose native form is
	// markdown. It is the same inherited flag the mutating verbs use for their JSON
	// envelopes; on reads it had been ignored (reads were view-selected only). It now
	// selects the structured debate — the sole view with both a markdown transcript (for
	// the human) and a JSON form (for the audits that used to regex the sections).
	//
	// --json ON A VIEW THAT IS ALREADY JSON IS ACCEPTED AND CHANGES NOTHING. It used to be refused
	// under [[one-way-no-aliases]], and the refusal was the harm: six seats piped `show work
	// --json` into python and crashed on KeyError: 'sitting', because the refusal's own envelope
	// parsed as cleanly as the data (#593). The rule forbids a second way to reach a DIFFERENT
	// form; a flag that yields the identical bytes is not a second form, it is the same one, and
	// refusing it cost a turn every time a seat reached for the obvious flag — six in #861's B3.
	// What stays refused is --json where no JSON form exists (a markdown view), because there the
	// seat asked for something the tool cannot give, and --json beside `--format markdown`, where
	// it asked for two forms at once.
	if asJSON, _ := cmd.Flags().GetBool(flags.JSON); asJSON && viewIsJSONByName(want) {
		if f, _ := cmd.Flags().GetString(flags.Format); f == "markdown" || f == "md" {
			return feov.Errorf(feov.Validation, "--json asks for the JSON the %s view already is, and --format %s asks for its markdown — pass one", want, f)
		}
	} else if asJSON {
		switch want {
		// changes IS A CHANGE LOG AND NOW CARRIES THE CHANGE. The markdown stays — its diff blocks are
		// the human-verification rendering — and this gives the same edits structured, so a seat can
		// select over them instead of rendering the report and grepping. 8 of m10's 23 unscoped
		// `changes` calls asked for a --json that did not exist, then inspected the refusal envelope
		// (`jq 'keys'`, `jq '.error'`) to work out why.
		//
		// --id SCOPES BOTH FORMS through one selection (record.ChangesOf), so the JSON holds the gap's
		// mint and the edits answering it exactly where the markdown does.
		case "changes":
			b, err := record.ChangesJSONBytes(run, Str(cmd, flags.ID))
			if err != nil {
				return err
			}
			return WriteSelected(cmd, b)
		case "debate":
			b, err := record.DebateJSONBytes(run)
			if err != nil {
				return err
			}
			return WriteSelected(cmd, b)
		// THE SCORECARD HAS A STRUCTURED FORM BECAUSE ITS READER IS A MACHINE. Asking for one and
		// being refused was the largest single source of failed commands in the 2026-09-20 run —
		// 10 of 55, plus 7 more as seats fell back to grepping the rendered card. The card's rows
		// were always structured; only the rendering was on offer.
		// THE EXPLORATION SPACE IS STRUCTURED UNDERNEATH — InquiriesOf already folds it into typed
		// lines; only the grouping-by-fate rendering was on offer.
		case "inquiries":
			fam, err := record.FamilyOf(run)
			if err != nil {
				return err
			}
			b, err := json.MarshalIndent(record.InquiriesJSONOf(fam.Events, fam.At), "", "  ")
			if err != nil {
				return err
			}
			return WriteSelected(cmd, append(b, '\n'))
		// THE LOG IS NOT A PROJECTION A SEAT OPENS. Its reader is the OPERATOR's `ops log`
		// (cli/log.go), never a seat view, so no case here names it and the default's message
		// below does not advertise it.
		// THE SET COMES FROM THE TABLE. It was six names written here by hand, and the
		// contract test's own comment records a stale name surviving in a list like this
		// while the assertion read as coverage.
		case "":
			return RefuseAndTeach(showGroup(cmd), feov.MissingField, "name a projection — this reads one at a time, and each is listed with the verb that fills it.")
		// THE REPORT'S JSON IS ITS OWN LINES, selected by the flags its markdown takes, so it is
		// served where those flags are read — the report arm below. It was the one view the shared
		// --json flag was advertised on and refused: 13 refusals on universe m18 and 7 on m17,
		// across the lenses, the chair and blue, and one seat piped the refusal's envelope through
		// jq and read its 150 bytes as the report.
		case "report":
		default:
			// EVERY PROJECTION HAS A JSON FORM, so nothing in the table reaches this; a view added
			// without one is refused by name rather than answered in a form the seat did not ask for.
			return feov.Errorf(feov.Validation, "the %s view has no --json form", want)
		}
	}

	// The BOARD is served as structured JSON, because a seat ACTS on it.
	//
	// Every other view is prose a seat reads; the board is state a seat has to make
	// decisions against — which gap is open, what its grades are, whether a closure
	// carried its anchor. Parsing that back out of markdown is precisely where the
	// scorecard defects came from (anchored_closures_pct read 0 against an 89
	// baseline because it parsed sentences while the anchors sat in structured
	// fields). Markdown for the same state stays available behind --view ledger.
	//
	// Resolved AFTER the role default, not before: `chair show` with no flags must
	// reach this branch, and checking the raw flag first sent the chair's own
	// default view looking for a board.md that no renderer writes.
	if want == "board" {
		// The markdown arm is what `ledger` and `archive` rendered: the open board for a human
		// verification pass, then the closure archive with its prose. Both were separate views
		// of data this JSON already carries whole.
		// AN UNKNOWN --format IS REFUSED, NOT ROUNDED TO JSON.
		//
		// This read `if f == "markdown" || f == "md"` with no else, so every other value —
		// including a typo, and including `dot`, which `graph --format` accepts — fell through to
		// the JSON arm and exited 0. A seat that asked for a rendering it did not get had no way
		// to find out: `--format banana` and `--format json` produced identical bytes.
		//
		// Found 2026-08-16 by widening setInHelp in the enum-help gate. The usage line spells
		// `json (…) | markdown (…)`, which is a closed-set promise, and nothing was keeping it.
		switch f, _ := cmd.Flags().GetString(flags.Format); f {
		case "", "json":
			// The default arm, below.
		case "markdown", "md":
			// THE MARKDOWN HAS NO ENTRIES TO SELECT, as on every markdown form below.
			if sel, serr := SelectorOf(cmd); serr != nil {
				return serr
			} else if sel.Active() {
				return feov.Errorf(feov.Validation, "--match and --quote select gaps from the JSON form of board; --format %s is markdown and has no entries to select. Drop --format", f)
			}
			// One record read for both renders: ledger and archive are two views of the same
			// family, and each view.Markdown call would fetch it again.
			in, err := view.InputOf(run)
			if err != nil {
				return err
			}
			led, err := view.MarkdownFrom(in, "ledger", "")
			if err != nil {
				return err
			}
			arc, err := view.MarkdownFrom(in, "archive", "")
			if err != nil {
				return err
			}
			cmd.OutOrStdout().Write(led)
			cmd.OutOrStdout().Write([]byte("\n"))
			cmd.OutOrStdout().Write(arc)
			return nil
		default:
			return feov.Errorf(feov.Validation, "unknown --format %q (json | markdown) — "+
				"name one of the two; an unrecognised format is refused rather than rendered as the default, where a typo would be invisible", f)
		}
		b, err := record.BoardJSONBytes(run)
		if err != nil {
			return err
		}
		// SELECTED GAPS. Every board read in universe-m10 began `jq '.open[]'` — the seat narrowing a
		// kitchen sink by hand. The selector filters the gap arrays on any of their text, and reports
		// what it kept of what there was, so a pattern that matched nothing is distinguishable from a
		// board that holds nothing.
		return WriteSelected(cmd, b)
	}
	// findings is served as JSON too, and for the same reason: the chair ACTS on it
	// (coalesces findings into gaps), so it reads structured fields, not prose it must
	// parse. This is the channel that replaced red/candidates/*.md.
	if want == "findings" {
		b, err := record.FindingsJSONBytes(run)
		if err != nil {
			return err
		}
		return WriteSelected(cmd, b)
	}
	// motions is JSON by name: a seat reads it to ANSWER a motion, so it needs the filer's
	// basis as a field rather than prose it must find in a transcript.
	// The artifact under audit, through the tool rather than off disk. Anchors intact: they are
	// what `edit` holds a seat responsible for carrying across an edit.
	if want == "report" {
		// BOTH FORMS ARE ONE READ AND ONE SELECTION. The JSON takes the text and its head off one
		// snapshot; what --anchor, --match and --quote choose is decided once, below, and each form
		// prints it.
		asJSON, _ := cmd.Flags().GetBool(flags.JSON)
		// THE ASSEMBLED SET, from the assembler itself: report.Layout assembles the record as it
		// stands and slices what it made, so no second description of the report's shape exists
		// to drift from the one that ships. It reads before any report is ingested, which is when
		// the seat about to write one asks.
		if assembled, _ := cmd.Flags().GetBool(flags.Assembled); assembled {
			if Str(cmd, flags.Anchor) != "" || cmd.Flags().Changed(flags.Window) {
				return feov.Errorf(feov.Validation, "--assembled reads the assembled set section by section, and --anchor reads the report's text at one anchor — pass one or the other")
			}
			layout, err := report.Layout(run)
			if err != nil {
				return err
			}
			if asJSON {
				out, err := json.MarshalIndent(layout, "", "  ")
				if err != nil {
					return err
				}
				return WriteSelected(cmd, append(out, '\n'))
			}
			if sel, serr := SelectorOf(cmd); serr != nil {
				return serr
			} else if sel.Active() {
				return feov.Errorf(feov.Validation, "--match and --quote select sections from the --json form of --assembled; the bare form prints every section")
			}
			if _, werr := fmt.Fprint(cmd.OutOrStdout(), layout.Markdown()); werr != nil {
				return werr
			}
			_, werr := fmt.Fprintln(cmd.ErrOrStderr(), report.LayoutHasNoText)
			return werr
		}
		var b []byte
		var head int64
		// THE WHOLE REPORT SERVED TO A LENS IS RECORDED, and it is the lens's pin (record.ServeReport):
		// both whole forms, bare and --json, and no read at an anchor or a selection, which serve part
		// of it. The text is printed after the read is committed, so a read the record could not take
		// is refused rather than served unrecorded.
		serve := role == "lens" && Str(cmd, flags.Anchor) == "" && !cmd.Flags().Changed(flags.Window)
		if serve {
			sel, serr := SelectorOf(cmd)
			if serr != nil {
				return serr
			}
			serve = !sel.Active()
		}
		if serve {
			s, err := Begin(cmd)
			if err != nil {
				return err
			}
			text, h, err := record.ServeReport(s.Identity())
			if err != nil {
				return err
			}
			b, head = []byte(text), h
		} else if asJSON {
			text, h, err := record.ReportAtHead(run)
			if err != nil {
				return err
			}
			b, head = []byte(text), h
		} else {
			var err error
			if b, err = report.BlueReportForReading(run); err != nil {
				return err
			}
		}
		emit := func(lines []record.ReportLineJSON, selection map[string]any) error {
			out, err := reportJSON(head, lines, selection)
			if err != nil {
				return err
			}
			_, werr := cmd.OutOrStdout().Write(out)
			return werr
		}
		if a := Str(cmd, flags.Anchor); a != "" {
			// ONE WINDOW IS NOT A SET TO SELECT FROM, as at `evidence` below: the pair would be
			// accepted and do nothing, and the seat would read the window as its selection.
			if sel, serr := SelectorOf(cmd); serr != nil {
				return serr
			} else if sel.Active() {
				return feov.Errorf(feov.Validation, "--anchor reads the report at one anchor, and --match/--quote select lines from the whole report — pass one or the other")
			}
			// AN ANCHOR THAT LEFT WITH ITS CLAIM IS NOT STALE. `retire` takes a bare anchor out
			// of the report on the record; saying "stale reference or another run" of it would send
			// the reader hunting for an error that is really a recorded exit.
			if !strings.Contains(string(b), anchor.Token(a)) {
				ev, claim, kept, found, err := record.AnchorRetiredAt(run, a)
				if err != nil {
					return err
				}
				if found && kept {
					return feov.Errorf(feov.Conflict, "anchor %s is not in the report because blue RETIRED the citation at event %d, keeping the sentence it stood on: %q. It is not a stale reference; read that sentence with --quote",
						a, ev, claim)
				}
				if found {
					return feov.Errorf(feov.Conflict, "anchor %s is not in the report because it was RETIRED — it left with the claim %q at event %d (blue's `retire` takes a bare anchor out with its claim). It is not a stale reference; there is no live text around it to read",
						a, claim, ev)
				}
			}
			n, _ := cmd.Flags().GetInt(flags.Window)
			w, err := anchor.ReadAround(string(b), a, n)
			if err != nil {
				return err
			}
			if asJSON {
				return emit(record.ReportLines(string(b))[w.First-1:w.Last], nil)
			}
			cmd.OutOrStdout().Write([]byte(w.Render()))
			return nil
		}
		// --window WITHOUT --anchor IS REFUSED, not quietly ignored. A seat that asked to
		// narrow its read and got the whole document back cannot tell that from a report which
		// is simply that long — the same shape as `--format banana` rendering JSON and exiting
		// 0, one branch up.
		if cmd.Flags().Changed(flags.Window) {
			return feov.Errorf(feov.Validation, "--window sizes a window and there is no window without --anchor <id>. "+
				"Either name the anchor you are reading at, or drop --window and take the whole report")
		}
		// SELECTED LINES, WITH WHAT IT TAKES TO CITE THEM. A seat hunting an undiscovered defect has
		// no anchor to read at — that is what makes it undiscovered — so it rendered the whole report
		// and grepped, then sed'd a line range to find the heading above the hit. This returns the
		// same triple --anchor does (heading, text, line number) keyed on a pattern instead of an id.
		sel, serr := SelectorOf(cmd)
		if serr != nil {
			return serr
		}
		if asJSON {
			lines := record.ReportLines(string(b))
			if !sel.Active() {
				return emit(lines, nil)
			}
			// A PATTERN THAT MATCHED NOTHING IS `lines: []` BESIDE `of`, the report's line count —
			// a read that found nothing, told apart from an empty report by the block it carries.
			hits := hitLines(lines, sel)
			block := selectionBlock(sel, len(hits), len(lines))
			if len(hits) == 0 {
				// THE MISS THAT IS AN ANCHOR SAYS SO IN THE DOCUMENT, where the counts are: see
				// Selector.InAnchorsOnly. Present only when it applies.
				if note := anchorsOnlyNote(anchorsOnly(lines, sel)); note != "" {
					block["note"] = note
				}
			}
			return emit(hits, block)
		}
		if sel.Active() {
			out, n := selectReportLines(string(b), sel)
			fmt.Fprint(cmd.OutOrStdout(), out)
			if n == 0 {
				lines := record.ReportLines(string(b))
				fmt.Fprintf(cmd.OutOrStdout(), "no line of the report matches %s — the report has %d line(s). This is a read that found nothing, not an empty report.\n",
					sel.Describe(), len(lines))
				if note := anchorsOnlyNote(anchorsOnly(lines, sel)); note != "" {
					fmt.Fprintln(cmd.OutOrStdout(), note+".")
				}
			}
			return nil
		}
		_, werr := cmd.OutOrStdout().Write(b)
		return werr
	}
	// evidence is JSON by name: it is a LOOKUP TABLE keyed by the anchor token a seat is holding,
	// and a markdown rendering of it would be a table to parse rather than a field to read.
	if want == "evidence" {
		if a := Str(cmd, flags.Anchor); a != "" {
			// ONE ENTRY IS NOT A SET TO SELECT FROM. The selector filters a projection's arrays and
			// this read has none, so the pair would be accepted and do nothing.
			if sel, serr := SelectorOf(cmd); serr != nil {
				return serr
			} else if sel.Active() {
				return feov.Errorf(feov.Validation, "--anchor reads the one entry at an anchor, and --match/--quote select entries from the table — pass one or the other")
			}
			b, err := evidenceAt(run, a)
			if err != nil {
				return err
			}
			_, werr := cmd.OutOrStdout().Write(b)
			return werr
		}
		b, err := record.EvidenceJSONBytes(run)
		if err != nil {
			return err
		}
		return WriteSelected(cmd, b)
	}
	if want == "motions" {
		b, err := record.MotionsJSONBytes(run)
		if err != nil {
			return err
		}
		return WriteSelected(cmd, b)
	}
	// work is JSON by name too — the seat ACTS on it (scans the open set, screens candidates),
	// so it reads structured fields, not prose. It is the shrinking once-per-turn read that the
	// full board JSON is not, and it is the ONE command a seat is told to run: everything open
	// to it, whether it may close, and which items are what stop it closing.
	if want == "work" {
		b, err := record.WorkJSONBytes(run, role, Of(cmd).SeatID)
		if err != nil {
			return err
		}
		return WriteSelected(cmd, b)
	}
	// telemetry is JSONL by name — one line per epoch, the wire shape the stopping
	// judgment reads. It is a SERIES, not a snapshot, and the series is the whole
	// point: a single epoch's numbers cannot show a trend changing character.
	if want == "telemetry" {
		b, err := view.TelemetryJSONL(run)
		if err != nil {
			return err
		}
		return WriteSelected(cmd, b)
	}
	if want == "" {
		return RefuseAndTeach(showGroup(cmd), feov.MissingField, "name a projection — this reads one at a time, and each is listed with the verb that fills it.")
	}
	known := false
	for _, v := range views {
		if v.name == want {
			known = true
		}
	}
	if !known {
		// THE MENU, NOT THE NAMES. A seat that mistypes a view is a seat that does not know
		// the view space, and a list of twelve nouns tells it which words are legal while
		// leaving it to guess which one holds what it wants. The menu is COBRA'S — every view
		// is a real subcommand with its own Short, so the framework already renders this list
		// and renders it the same way everywhere.
		return RefuseAndTeach(showGroup(cmd), feov.Validation, fmt.Sprintf("unknown view %q — each view is listed with the verb that fills it.", want))
	}

	// THE MARKDOWN FORM HAS NO ENTRIES TO SELECT. Its bytes went through WriteSelected, which hands
	// back what it cannot parse as a document, so `--match` on a bare markdown view printed every
	// entry and exited 0 — the answer a pattern matching all of them gets (universe m26: `show
	// inquiries --match zzqqxx` returned all six). The pair selects the --json form of these views,
	// and is refused here by name.
	if sel, serr := SelectorOf(cmd); serr != nil {
		return serr
	} else if sel.Active() {
		return feov.Errorf(feov.Validation, "--match and --quote select %s from the --json form of %s; the bare form is markdown and has no entries to select. Add --json", selectorNoun(want), want)
	}

	scope := Str(cmd, flags.ID)

	b, err := view.Markdown(run, want, scope)
	if err != nil {
		return err
	}
	_, werr := cmd.OutOrStdout().Write(b)
	return werr
}

// RoleVerbs is the seat's verb set, ready to mount at the ROOT of that seat's tree.
//
// There is no role level: the seat is bound on the record (by its register, or by the hook's
// bracket where the configuration seats one), the tool reads the role from that binding, and the
// root IS the seat's surface. A lens seat runs `mint`; a chair seat running `mint`
// is told the verb is the lens seat's — the boundary drawn by the only fact that decides it.
//
// The seat types no word the tool can already look up: its role is read from its registered seat
// id, and there is one copy of it, on the record.
func RoleVerbs(role string, verbs ...*cobra.Command) []*cobra.Command {
	out := make([]*cobra.Command, 0, len(verbs)+1)
	for _, v := range verbs {
		// Applied HERE rather than in each verb: a verb that had to remember to mark its own
		// required flags is a verb that can forget, and the forgetting is silent — the help
		// simply looks like everything is optional.
		markTree(v)
		out = append(out, v)
	}
	out = append(out, Show(role))
	// THE SEATS THAT ADJUDICATE, WHICH IS NOT ONLY THE BENCH. The chair issues the verdict and
	// rules grade motions; the bench rules the docket, petitions and inquiries. Both decide
	// BETWEEN two parties and may take neither one's account, which is what this group is for —
	// and a refusal that points the chair at `inquest motions` has to reach a chair that has it.
	// A lens and blue do the work being judged, so they do not carry it.
	if role == "bench" || role == "chair" {
		out = append(out, Inquest())
	}
	return out
}

// registerResult and closingResult are the shared-verb results: these verbs are built by
// seat for every role, so their result types live beside them.
// registerResult reports WHICH DISPATCH this is, not an opaque sitting id.
//
// It carried a `nonce` — eight hex characters naming the shard file the seat would write to. There
// is no shard file, and the nonce is gone from the record entirely (it scoped the idempotency
// ordinal per sitting while the key index is global, so a re-dispatched seat collided with its own
// earlier acts and could not record at all). What a seat can actually use is the count: being told
// this is dispatch 2 says a previous attempt exists.
type registerResult struct {
	SeatID   string `json:"seat_id"`
	Dispatch int    `json:"dispatch"`
	// IdentityAbsent is true when a DISPATCHED seat registered with no agent id — the hook did
	// not reach this call, so nothing on the record can bind this agent to this seat.
	IdentityAbsent bool `json:"identity_absent,omitempty"`
	// RunVia says which path supplied the run directory: injected | flag | wrapper | inferred.
	RunVia string `json:"run_via,omitempty"`
	// HookAbsent says the hook is not reaching the RUN, rather than only the identity being
	// missing — every seat in it is affected, not just this one.
	HookAbsent bool `json:"hook_absent,omitempty"`
	// TurnLimitUnarmed is why this sitting's tool calls are not being counted, when the count
	// could not be opened. Empty when it was.
	TurnLimitUnarmed string `json:"turn_limit_unarmed,omitempty"`
}

// hookAbsentSource names the carrier that stood in for the hook, so the operator reading a
// log entry knows which of the two shapes this run was in — a wrapper doing its job, or
// nothing but the marker and a lucky cwd.
func (r registerResult) hookAbsentSource() string {
	if r.RunVia == string(seatenv.RunFromWrapper) {
		return "this run's own wrapper, which setup baked the run into"
	}
	return "the tool's own search for the run marker"
}

func (r registerResult) Human() string {
	out := "registered " + r.SeatID
	if r.Dispatch > 1 {
		out = fmt.Sprintf("registered %s (dispatch %d — a previous dispatch of this seat is on the record)", r.SeatID, r.Dispatch)
	}
	if r.IdentityAbsent {
		// THE SECOND CONSEQUENCE IS NOT THE ONE THE SHARD RECORD CARRIED. That warning ended
		// "re-registering rotates your shard nonce and replay then discards everything the first
		// sitting recorded" — true of shards, and false here: both sittings' events are rows and
		// nothing selects a winner, so a second register costs you a dispatch count and no work.
		// Copying the sentence across would have been a seat-facing threat about a mechanism that
		// no longer exists, which is worse than no warning: it teaches a false model of the record.
		out += "\n\nYOUR IDENTITY DID NOT REACH THE TOOL, so nothing on the record binds this agent " +
			"to this seat. You are registered and your events are recorded; what is missing is the " +
			"binding that lets a LATER call resolve who you are without being told.\n\n" +
			"TWO CONSEQUENCES:\n" +
			"  PASS --seat-id ON EVERY CALL. Without it the tool cannot tell which seat is asking " +
			"and will refuse you.\n" +
			"  REGISTERING AGAIN WILL NOT FIX IT. If a later call says \"this agent has not " +
			"registered\", it is THIS, not a lost registration — and a second register cannot supply " +
			"an agent id the hook never sent. Your work is safe either way; the record is " +
			"append-only and keeps both sittings."
	}
	if r.HookAbsent {
		// SEPARATE FROM THE IDENTITY BLOCK, and readable without it: under a wrapper this fires
		// on its own, and the seat reading it may have no other symptom at all.
		out += "\n\nTHE ENGINE'S HOOK IS NOT REACHING THIS RUN, so EVERY seat in it is affected, not " +
			"just you. The run directory reached this tool from " + r.hookAbsentSource() + ", and not " +
			"from the hook — which injects it on every call when it is working.\n\n" +
			"YOUR WORK IS NOT AT RISK FROM THIS. The run directory is correct and your events are " +
			"recorded against it. What is lost is the identity binding, and the fix for that is " +
			"above. Record the hook's absence ONCE in the log — you are the first party " +
			"that can see it, and the run leaves no other trace of it."
	}
	if r.TurnLimitUnarmed != "" {
		out += "\n\nThis sitting's tool calls are not being counted against the run's limit (" +
			r.TurnLimitUnarmed + "). Record it once in the log."
	}
	return out
}

type closingResult struct {
	GapID string `json:"gap_id"`
}

func (r closingResult) Human() string { return "closing filed for " + r.GapID }

// inquestGroup builds the bench's raw-record surface from the same table `show` reads, so a
// projection cannot be in both or in neither.
func inquestGroup() *cobra.Command {
	c := &cobra.Command{
		Use:          "inquest",
		Short:        "READ THE RECORD YOURSELF — what each side argued, what was contested and how it was ruled, and how the numbers moved. You rule between two parties; you take neither one's account of them",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
	}
	for _, v := range views {
		if !v.inquest {
			continue
		}
		v := v
		sub := &cobra.Command{
			Use:          v.name,
			Short:        v.short,
			Long:         v.long + jsonByNameHelpSuffix(v.jsonByName) + shapeHelpSuffix(v.name, v.jsonByName, v.shape),
			Args:         cobra.NoArgs,
			SilenceUsage: true,
			RunE:         func(cmd *cobra.Command, _ []string) error { return renderView(cmd, v.name) },
		}
		// THE INQUEST VIEWS TAKE THE SELECTOR PAIR TOO. This group is built here rather than in the
		// `show` loop, so a flag added there reaches only half the views — which is how `--match` and
		// The literal selector reached `show board` and not `inquest motions`, in the very change
		// that was meant to stop behaviour being built twice. Both builders call the same helper now.
		//
		// TELEMETRY TAKES NO SELECTOR: it is a series of JSONL lines, one per epoch, with no document
		// whose arrays hold entries and nowhere to state what a selection kept — so the pair it carried
		// returned every line for every pattern. Selecting epochs out of a trend is not a read it has.
		if v.name != "telemetry" {
			// A MARKDOWN VIEW SELECTS ITS --json FORM, and says so on the line of the flag a seat is
			// about to type, as the report's page does for its own bounds. Said in the noun, before the
			// explanation every view's pair closes with, so the manual still lifts that explanation once.
			what := selectorNoun(v.name)
			if !v.jsonByName && v.name != "report" {
				what += markdownSelectorBound
			}
			AddSelectorFlags(sub, what)
		}
		c.AddCommand(sub)
	}
	return c
}

// Location is the location a placing verb stores for its --quote: the quote's visible text, every
// anchor token out. A seat quotes a sentence as `show report` prints it, anchors included; a quote is
// located with anchors skipped, so the stored text places the anchor where the typed quote does, and
// no reader of a location is handed a token as part of the sentence.
func Location(cmd *cobra.Command) string {
	return anchor.Replace(Str(cmd, flags.Quote), func(string, string) string { return "" })
}

// Places is the guard a placing verb writes with: it refuses, in refuse's words, where Attach will not
// place id's anchor at loc in the report the write lands on (record.Landing).
//
// ASKED OF THE REPORT THE ANCHOR LANDS ON, never of a render made before the write: another seat's
// edit that commits in between can take the quote out, and an Anchor at a location replay cannot place
// fails every render of the report from then on (#1252).
func Places(id, loc string, refuse func(error) error) record.Guard {
	return record.CheckOnReport(places(id, loc, refuse))
}

// places is Places' check, for a guard that asks it after reading something else first.
func places(id, loc string, refuse func(error) error) func(*record.Landing) error {
	return func(l *record.Landing) error {
		current, err := l.Report()
		if err != nil {
			return err
		}
		if _, err := anchortext.Attach(current, id, loc); err != nil {
			return refuse(err)
		}
		return nil
	}
}

// AppendPlaced records act and the Anchor that places its marker — id's, at loc — as one write
// (record.AppendPlaced): the placement is checked against the report the pair lands on and refused in
// refuse's words, before any guards the verb passes. A quote is held to the report whether or not the
// act places an anchor at it; an act that names no id places nothing, and one with no quote is checked
// by its guards alone. It returns the act's event as written, which carries the act's key.
func AppendPlaced(s Context, act proto.Message, id, loc string, refuse func(error) error, guards ...record.Guard) (*record.Event, error) {
	if strings.TrimSpace(loc) != "" {
		guards = append([]record.Guard{Places(id, loc, refuse)}, guards...)
	}
	return record.AppendPlaced(s.Identity(), act, id, loc, guards...)
}

// StaleRule is how a page whose act names a gap teaches AppendFresh's refusal: one sentence set for
// every such page, so the pages and the refusal they describe cannot drift apart.
const StaleRule = "A GAP THAT CHANGED AFTER YOUR SITTING OPENED REFUSES THIS ACT ON IT. Other seats sit while you do, so a gap can be closed, regraded, remanded, answered by blue, or argued in another seat's finding after you read it. The refusal records none of the act and lists each change on its own line, with its event number and the seat that made it. Read the gap as it now stands; if the act still holds, record it again — the refusal moved your reading of that gap past what it showed, so the same act is not refused for those changes twice. Only other seats' acts count: your own never refuse you."

func init() {
	HelpValues["StaleRule"] = StaleRule
	HelpValues["GradeMovementRule"] = record.GradeMovementRule
}

// AppendFresh records an act about gapID through write, handing it the guard that refuses the act
// while another seat's act on the gap is newer than this seat's read point (record.Fresh). On that
// refusal it records the Stale event that moves the read point past what the refusal shows, and
// returns the refusal, so the same act recorded again once the seat has read the changes lands.
func AppendFresh(s Context, gapID string, write func(guards ...record.Guard) error) error {
	err := write(record.Fresh(s.SeatID, gapID))
	var stale *record.StaleError
	if !errors.As(err, &stale) {
		return err
	}
	if _, aerr := record.Append(s.Identity(), &recordpb.Stale{GapId: proto.String(gapID), Through: proto.Int64(stale.Through())}); aerr != nil {
		return aerr
	}
	return feov.Wrap(feov.Conflict, stale, "%s", stale.Error())
}

// PlaceOwed appends the Anchor a retried act owes — an act on the record without its anchor, which a
// record migrate brought forward from an archive can hold (record.UnplacedLocation) — at the location
// the act stored, once that location places; otherwise the act stays unplaced. key is the --key the call repeats the act under,
// or "" for an act its source and claim key. Whether the anchor is owed, and whether its location
// places, are asked of the record the Anchor lands on: an anchor placed in between is not placed twice.
//
// Its refusal is its own, never the verb's first-call wording: that wording teaches a seat to fix
// --quote, and a retry is held to the --quote the act stored (RepeatsKeyed), so another one is a new act.
func PlaceOwed(s Context, id, key string) error {
	body := &recordpb.Anchor{Id: proto.String(id)}
	_, err := record.Append(s.Identity(), body, record.CheckOnReport(func(l *record.Landing) error {
		loc, err := l.UnplacedLocation(id)
		if err != nil {
			return err
		}
		if loc == "" {
			return record.ErrWritesNothing
		}
		body.Location = proto.String(loc)
		return places(id, loc, func(err error) error {
			state, repeats, anew := "no longer holds it where an anchor can stand", "naming the same source and claim", ""
			if errors.Is(err, anchortext.ErrAmbiguous) {
				state = "now holds it more than once"
			}
			if key != "" {
				repeats, anew = fmt.Sprintf("under --key %q", key), fmt.Sprintf(", under a --key other than %q", key)
			}
			return feov.Errorf(feov.Conflict, "the %s this call repeats is on the record without its anchor, and the location it stored, %q, does not place: the report %s. "+
				"A call %s repeats that act, its --quote included, so it cannot move the anchor. "+
				"The act stands as recorded, with no anchor in the report. To anchor the same point at text the report holds, record a new act: that text as --quote%s",
				anchor.Kind(id), loc, state, repeats, anew)
		})(l)
	}))
	if errors.Is(err, record.ErrWritesNothing) {
		return nil
	}
	return err
}
