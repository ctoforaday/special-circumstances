package lens

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchortext"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/bluedoc"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/blue"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/enumhelp"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/reportproj"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/reportvoice"
)

// mint: put a gap on the board.
//
// The id is TOOL-assigned, one run-global sequence (G<n>). Letting seats choose produced
// four different "R5-1"s in one round of run 3, and the collision class dies here
// rather than being policed downstream.
//
// The duplicate screen runs AT THE WRITE because another lens may mint while this one works, so a
// `near-match` read earlier cannot answer for the board as the mint finds it. A soft tell in the
// printed text ("this run", "epoch 3") is flagged and not refused because a subject can use those
// words too. What to do where a matched gap IS this defect is on the page (help/mint.md), once.
func newMint() *cobra.Command {
	var severity, likelihood, impact, cx flags.GradeValue
	var supersedes, foundBy, distinctFrom flags.CSV

	c := seat.Prose(seat.New("mint", func(s seat.Context, cmd *cobra.Command) (seat.Result, error) {
		run, err := s.Run()
		if err != nil {
			return nil, err
		}
		// The argument is resolved BEFORE the retry check, as `edit` and `prove` resolve theirs: a
		// retry is the same act, judged on the same inputs, and the first call recorded them.
		text, err := seat.Reason(cmd)
		if err != nil {
			return nil, err
		}
		// The id is minted once the retry check has passed: a retry answers with the stored one.
		var gapID string
		// ONE FLAG, ONE FIELD. --problem states what is wrong and is the only thing that does;
		// --reason is red's argument for the gap and is never read as the problem. The problem
		// used to fall back to --reason when --problem was absent, so the same flag meant the
		// problem on one mint and the argument on the next, and a seat could not tell from its own
		// call which field its text had landed in. An absent --problem is refused by the record's
		// required field, in the problem's own words.
		problem := seat.Str(cmd, flags.Problem)
		// RED'S ARGUMENT IS KEPT.
		//
		// MEASURED 2026-08-16 BY ASKING THE BENCH. Dispatched to a petition sitting about what
		// a required_fix may demand, it reported first among its missing things: "Red's closing
		// or reasoning — Why did red mint this gap? Do they believe blue DID run a search and
		// just didn't describe it? The gap's problem statement is clear, but red's intent for
		// the fix isn't visible to me." It was adjudicating red's intent with red's intent
		// unrecorded.
		//
		// `problem` says WHAT IS WRONG. This says why red thinks so, which is the half a seat
		// answers and a bench weighs.
		p := &recordpb.Mint{}
		if text != "" {
			p.MintReason = proto.String(text)
		}
		p.MintKey = proto.String(seat.Str(cmd, flags.Key))
		// THE CLASS IS A WORD OF THE CLOSED GapClass SET; the flag's parser refused any other, so
		// an empty word here is an absent flag, which the required-field walk names.
		if w := seat.Str(cmd, flags.Class); w != "" {
			gc, ok := record.GapClassOf(w)
			if !ok {
				return nil, feov.Errorf(feov.Validation, "%q is not a gap class — it says what KIND of defect this is", w)
			}
			p.Class = &gc
		}
		var guards []record.Guard

		// THE LOCATION IS MATCHED AGAINST THE REPORT, and it was not.
		//
		// `lens finding --quote` is REFUSED on a mis-quote — the anchor cannot be spliced
		// where the text is not. Mint's took prose, on the reasoning that a gap's location is
		// "recorded for a reader". Measured (#359): three gaps were minted naming report
		// sections that do not exist, two of them `--existence verified`. A seat asserted it
		// had checked a defect at the leaf, at a place the tool never confirmed was there.
		//
		// A location a reader cannot find is not a location. It is refused now, by the same
		// rule the lens has always been held to.
		//
		// AN OMISSION HAS A PLACE, AND IT IS NOT A BORROWED SENTENCE: an absence names the
		// section it is missing from, the inquiry whose question it argues against, or the gap
		// it is about, and the record CHECKS the reference. The two verbs of one act spell it
		// one way, so the gap minted from such a finding carries the same handle the finding does.
		about, aboutRefP, aerr := record.ResolveAbout(run, record.MustEnum("mint", "about_kind"), seat.Str(cmd, flags.AboutKind), seat.Str(cmd, flags.About))
		if aerr != nil {
			return nil, aerr
		}
		loc := seat.Location(cmd)
		// ONE SUBJECT. An anchor is OPTIONAL on a gap — that is this verb's pre-existing shape and
		// is not changed here — but a gap that claims both is claiming two.
		if strings.TrimSpace(loc) != "" && about != nil {
			return nil, feov.Errorf(feov.Validation, "takes --quote OR --about, not both: a gap has one subject")
		}
		// THE GAP'S ANCHOR IS PLACED as every kind's is: at the end of its quote, which must occur
		// once within one paragraph of the report the mint lands on (seat.AppendPlaced, below). A gap
		// is one sentence (R-4), and the board reads its location from where its anchor stands.
		p.Location = proto.String(loc)
		p.AboutKind = about
		p.AboutRef = aboutRefP
		p.Problem = proto.String(problem)
		p.RequiredFix = proto.String(seat.Str(cmd, flags.Fix))
		// THE CONCRETE PROPOSAL, AND WHY fix_basis IS DERIVED (#267 stage 3).
		//
		// `required_fix` is prose ("acknowledge the shared definition"), which blue cannot
		// apply verbatim. Red MAY instead state the exact span and the exact replacement —
		// and stating one legally is impossible from memory of what the report probably
		// says, which is the forced re-read this whole axis exists for: all three of the
		// smoke's round-2 gaps were contradictions between blue's new text and text red
		// never re-read before prescribing.
		//
		// So `fix_basis` is COMPUTED from whether a validated pair is present. It is not a
		// flag. A seat asked to self-report "verified | proposed" reports the flattering
		// one, and the field becomes decoration — the same reason the finding id and the
		// gap id are tool-assigned rather than claimed.
		//
		// THE PAIR IS CHECKED AGAINST THE REPORT THE MINT LANDS ON. Blue's edit can commit in
		// parallel, and a pair validated against the report before it is a "verified" fix of text
		// the report no longer holds.
		p.FixBasis = proto.String("proposed")
		if fixNew := seat.Str(cmd, flags.New); fixNew != "" {
			guards = append(guards, record.CheckOnReport(func(l *record.Landing) error {
				report, err := l.Report()
				if err != nil {
					return err
				}
				return verifiedFix(p, report, gapID, seat.Str(cmd, flags.Quote), fixNew)
			}))
		}
		p.AcceptanceCheck = proto.String(seat.Str(cmd, flags.Check))

		if w := seat.Str(cmd, flags.CheckKind); w != "" {
			ck, ok := record.CheckKindOf(w)
			if !ok {
				return nil, feov.Errorf(feov.Validation, "%q is not a check kind — it says WHAT WOULD SETTLE the acceptance check", w)
			}
			p.CheckKind = &ck
		}
		p.Severity = seat.GradeOrNil(&severity)
		p.Likelihood = seat.GradeOrNil(&likelihood)
		p.Impact = seat.GradeOrNil(&impact)
		p.ComplexityCost = seat.GradeOrNil(&cx)
		p.Supersedes = supersedes.Value()
		p.FoundBy = foundBy.Value()
		p.DistinctFrom = distinctFrom.Value()

		// Crash-retry idempotency: --key (the stable local label, e.g. the source lens finding) makes
		// a retried mint return the EXISTING id — when the retry states the gap the key names. It is
		// asked once the gap is built, so the retry is held to the same inputs the first call was.
		if prior, err := record.ExistingMintByKey(run, s.SeatID, seat.Str(cmd, flags.Key)); err != nil {
			return nil, err
		} else if prior != nil {
			// --new is recorded by the fix's verification at the write; the retry states it as given.
			retry := proto.Clone(p).(*recordpb.Mint)
			if fixNew := seat.Str(cmd, flags.New); fixNew != "" {
				retry.FixNew = proto.String(fixNew)
			}
			if err := seat.RepeatsKeyed(seat.Str(cmd, flags.Key), "gap", prior.GetGapId(), prior, retry, mintAsserted...); err != nil {
				return nil, err
			}
			// A mint on the record without its anchor — one migrate brought forward from an archive
			// whose run crashed between the two — is finished by the retry, at the location it stored.
			if err := seat.PlaceOwed(s, prior.GetGapId(), seat.Str(cmd, flags.Key)); err != nil {
				return nil, err
			}
			return mintResult{GapID: prior.GetGapId(), Check: prior.GetAcceptanceCheck(), Idempotent: true}, nil
		}
		gapID = record.NewID("gap")
		p.GapId = proto.String(gapID)

		// ESTOPPEL: RED IS BOUND BY THE FIX IT PRESCRIBED (#267 stage 4).
		//
		// A gap whose location is text red PRESCRIBED and blue applied VERBATIM is an
		// amendment to the gap that prescribed it, not a fresh gap. Measured: 3 of 3 of the
		// 2026-08-04 smoke's round-2 gaps were about text blue added in round 1 at red's own
		// instruction, and blue had no channel to say so.
		//
		// Red is not silenced — it argues the point on the ORIGINAL gap, where its own
		// prescription sits on the record beside the complaint, or mints with --supersedes so
		// the lineage is explicit. What it may not do is open a clean slate against its own
		// words.
		//
		// THE REJECTION IS LOGGED, AS AN ESTOPPEL, not merely refused. A guard that silently
		// blocks is invisible: this makes "how often does red relitigate its own
		// prescriptions" a per-run number on the record, which is the measurement of the
		// pathology the guard prevents. Same discipline as `blue cite`'s unreachable-source
		// rejection — the tool records the block, never the seat's memory of it.
		//
		// ASKED OF THE RECORD THE MINT LANDS ON: blue's `edit --accept` of this lens's prescription
		// commits in parallel, and a mint checked against the record before it opens a fresh gap
		// against the lens's own words. The refusal is decided under the write lock; the log entry
		// that records it is written after, as its own act.
		guards = append(guards, record.Check(func(l *record.Landing) error {
			fam, berr := l.Family()
			if berr != nil {
				return nil
			}
			if prior, prescribed := record.EstoppelConflict(fam, seat.Str(cmd, flags.Quote)); prior != "" &&
				!contains(supersedes.Value(), prior) {
				return estopped{prior: prior, prescribed: prescribed}
			}
			return nil
		}))

		var est estopped
		switch err := seat.AppendPlaced(s, p, gapID, loc, mintPlacementRefusal, guards...); {
		case errors.As(err, &est):
			prior, prescribed := est.prior, est.prescribed
			msg := fmt.Sprintf("estoppel — this gap's location is text YOU prescribed for %s and blue applied verbatim. The prescription is red's; raise it as an amendment to %s (argue it there, or mint with --supersedes %s so the lineage is explicit) rather than as a fresh gap against your own words. Prescribed text: %q",
				prior, prior, prior, prescribed)
			// The KIND is a field, not something a reader infers from the wording. The
			// prose above is aimed at a seat and must stay editable; the count an
			// operator reads must not move when it is edited (#283).
			// THE TYPE IS A VALUE NOW, not a key/word pair. `text` is the field the log
			// projection reads, and the old payload set the wrong key — every estoppel
			// block recorded an entry that rendered with EMPTY TEXT, so the block was
			// logged and its explanation was not. The schema removes the way to make that
			// mistake: one prose field, one typed assertion, one source.
			//
			// SOURCE IS TOOL, and that is the axis the old single enum conflated: this is
			// the tool refusing a mint, not a seat filing a report about its sitting, and
			// an operator filtering for seat reports must not have to know which VALUES
			// were tool values to exclude it.
			lt := recordpb.LogType_LOG_TYPE_ESTOPPEL
			src := recordpb.LogSource_LOG_SOURCE_TOOL
			fr := &recordpb.Log{
				Text:   proto.String(msg),
				Type:   &lt,
				Source: &src,
			}
			if _, ferr := record.Append(s.Identity(), fr); ferr != nil {
				return nil, ferr
			}
			return nil, record.ToolLogged(feov.Errorf(feov.Conflict, "%s", msg))
		case err != nil:
			return nil, err
		}
		// THE AMBIGUOUS TELLS ARE ADVICE, and they ride back on the confirmation. The write path has
		// already refused the unambiguous ones (record.refuseMintReportVoice), so what remains is the
		// set a pattern cannot tell from subject prose — "this run" or "the red team" may be exactly
		// right in a report about red teams. Same three fields the report prints, never the reason.
		var tells []string
		for _, text := range []string{p.GetProblem(), p.GetRequiredFix(), p.GetFixNew()} {
			for _, f := range reportvoice.Advised(text) {
				tells = append(tells, f.String())
			}
		}
		return mintResult{GapID: gapID, Check: seat.Str(cmd, flags.Check), VoiceTells: tells}, nil
	}))

	c.Flags().String(flags.Key, "", flags.DescKey+". A repeat under the handle that states the gap otherwise — its quote, class, problem, fix, check, grades, subject, --reason, --supersedes or --found-by — is refused: a handle names one gap")
	enumhelp.Flag(c, flags.Class, record.MustEnum("mint", "class"),
		"the KIND of defect, from the values below; its material default is recorded with the gap. unclassified only when none fits, and then --reason names the nearest value and what it does not cover")
	flags.Text(c, flags.Quote, "the ONE sentence the defect lives at — "+flags.DescQuoteAlone+". The tool anchors the gap at its end, and the board shows that sentence with its section")
	enumhelp.Flag(c, flags.AboutKind, record.MustEnum("mint", "about_kind"),
		"anchor this gap to something that is NOT report text — a section for what is missing from it, an inquiry, or a gap already on the board; use instead of --quote")
	flags.Text(c, flags.About, "the reference --about-kind names: a section heading, an inquiry id, or a gap id. It is CHECKED against the record")
	flags.Text(c, flags.Problem, "what is wrong with the report, in the subject's voice — printed in the report's risk matrix. Your argument for the gap goes in --reason, which the report never prints")
	flags.Text(c, flags.Fix, "the required fix, as prose — what must become true. This is the substantive channel: research it, enumerate it, qualify it")
	flags.Text(c, flags.New, fmt.Sprintf("concrete proposal, TEXTUAL DEFECTS ONLY: the exact text --quote should become. A replacement more than %d characters longer than the span is refused as AUTHORING — a substantive addition is blue's to write, and you say so in --fix. Passing it records fix_basis: verified, and it is refused where blue's accept would refuse it", bluedoc.MaxProposalGrowth))
	flags.Text(c, flags.Check, "the acceptance check red will RUN at re-audit — the pre-agreed contract, not a description. For a class defect, the check is the enumerating command and its pass condition")
	enumhelp.Flag(c, flags.CheckKind, record.MustEnum("mint", "check_kind"), ("what would SETTLE that check"))
	c.Flags().Var(&severity, flags.Severity, flags.GradeUsage(record.AxisMeans(recordpb.GradeDimension_GRADE_DIMENSION_SEVERITY)))
	c.Flags().Var(&likelihood, flags.Likelihood, record.AxisMeans(recordpb.GradeDimension_GRADE_DIMENSION_LIKELIHOOD))
	c.Flags().Var(&impact, flags.Impact, record.AxisMeans(recordpb.GradeDimension_GRADE_DIMENSION_IMPACT))
	c.Flags().Var(&cx, flags.Complexity, record.AxisMeans(recordpb.GradeDimension_GRADE_DIMENSION_COMPLEXITY)+", on the same scale")
	c.Flags().Var(&supersedes, flags.Supersedes, "comma-separated ancestor ids this gap replaces; lineage is never dropped")
	c.Flags().Var(&foundBy, flags.FoundBy, "comma-separated ids of the lens findings that surfaced it, as the findings view lists them")
	c.Flags().Var(&distinctFrom, flags.DistinctFrom, "comma-separated OPEN gaps the mint's duplicate screen matched that this gap is NOT — your claim, on the record, that you read each and it is a different defect")
	// THE GAP ID IS REQUIRED OF THE RECORD AND SUPPLIED BY THE VERB — declared here, at the code
	// that does the supplying, so the fact and what makes it true cannot drift apart. Without it
	// the contract gate reads "mint declares gap_id required and registers no --id" and is right
	// to: a requirement with no flag behind it is invisible to a seat unless something says the
	// tool meets it.
	seat.Supplies(c, "gap_id", "the tool mints it (record.NewID), at random — an id a seat chose would name a gap the record does not hold")
	seat.Supplies(c, "class_material", "stamped from the class's material default at the write path")
	return c
}

// mintResult reports the tool-assigned gap id; a retry returns the existing one.
type mintResult struct {
	GapID      string `json:"gap_id"`
	Idempotent bool   `json:"idempotent,omitempty"`
	// Check is the acceptance check AS RECORDED, echoed back because the seat cannot otherwise
	// see what landed.
	//
	// MEASURED: in research/2026-09-02_quadratic-formula (red-merge-r3) backticks in two --check
	// strings were substituted by the shell BEFORE mint ever saw them, so one gap's recorded check
	// held the command OUTPUTS instead of the commands and another lost its projection name. The
	// mint succeeded, printed only the new id, and echoed nothing — the corruption surfaced only
	// when someone read the board afterwards. The tool cannot prevent a shell from rewriting its
	// own argv; what it can do is show what it stored, which is the one thing that makes the
	// difference visible while the seat is still there to fix it. No verb amends a check after
	// mint, so the moment of the write is the only moment.
	Check string `json:"acceptance_check,omitempty"`
	// VoiceTells is ADVICE on the problem, the fix and a --new replacement, and the gap is already on the board by the
	// time it renders. The fix for a tell that turns out to be real is the next mint's wording, or
	// a supersession; nothing here asks the seat to undo the write.
	VoiceTells []string `json:"voice_tells,omitempty"`
}

func (r mintResult) Human() string {
	head := "minted " + r.GapID
	if r.Idempotent {
		head += " (idempotent retry — existing id returned)"
	}
	note := reportvoice.Note("the risk matrix, the first sentence of the problem and of the fix, while the gap is open; the body, a --new replacement blue accepts", r.VoiceTells)
	if r.Check == "" {
		return head + note
	}
	return head + "\n  acceptance check RECORDED as: " + r.Check +
		"\n  (read it back — a shell may have rewritten it before this tool saw it, and no verb amends a check after mint)" + note
}

// mintPlacementRefusal is finding's placement refusal, worded for a gap: one whose quote repeats
// may name its section instead, and one about something not in the report names that.
func mintPlacementRefusal(err error) error {
	r := placementRefusal("gap", err)
	switch {
	case errors.Is(err, anchortext.ErrAmbiguous):
		return fmt.Errorf("%w; name its section with --about-kind section", r)
	case errors.Is(err, anchortext.ErrOnHeading):
		return r
	}
	return fmt.Errorf("%w\n\nFor a gap about something that is NOT in the report, do not borrow a nearby sentence: name it with --about-kind/--about, the same pair `finding` takes", r)
}

// estopped is the estoppel guard's refusal: the gap prior prescribed text the mint's quote names, and
// blue applied it verbatim. The verb words it and logs it once the write has rolled back.
type estopped struct{ prior, prescribed string }

func (e estopped) Error() string { return "estoppel against " + e.prior }

// verifiedFix marks p's fix verified where --quote and --new are a pair blue's accept applies to report
// with the gap's anchor placed — judged by the accept's own planner (blue.PlanAccept), so a fix
// verified here is one the accept applies, and one it would refuse is refused here.
func verifiedFix(p *recordpb.Mint, report, gapID, quote, fixNew string) error {
	if err := bluedoc.RefuseAuthoring(quote, fixNew); err != nil {
		return err
	}
	placed, err := anchortext.Attach(report, gapID, quote)
	if err != nil {
		return err
	}
	if _, err := blue.PlanAccept(placed, gapID, quote, fixNew); errors.Is(err, reportproj.ErrNoChange) {
		return fmt.Errorf("%w A prescription blue would apply to no effect cannot be verified: fix the quote so its replacement changes the report, or state the fix as prose in --fix", err)
	} else if err != nil {
		// The accept's refusal of a dropped anchor is worded for blue, whose acts it names; red's is the
		// same fact, the anchor and its sentence.
		why := err.Error()
		var dropped *bluedoc.ErrAnchorDropped
		if errors.As(err, &dropped) {
			why = fmt.Sprintf("the span carries %s %s %s, and --new neither carries it nor keeps that sentence word for word", anchor.Kind(dropped.ID), anchor.Token(dropped.ID), dropped.On())
		}
		return feov.Wrap(feov.Validation, err, "--new is not a fix blue's accept would apply to the report as it stands, with this gap's anchor at the end of --quote: %s. "+
			"NOTHING IS RECORDED. A fix is verified only where the accept applies it, so what the accept would refuse is refused here. "+
			"The accept moves the gap's and the findings' anchors at the sentence's end to the end of --new; any other anchor in the span — a citation or a proof, which backs its sentence, or one inside the sentence — stays only where --new keeps its sentence word for word or holds the anchor itself. "+
			"Where such an anchor is what does not survive: keep the sentence it marks word for word in --new, and the tool puts the anchor back on it; "+
			"or write the anchor into --new after the text it marks, as `show report` prints it; "+
			"or leave --new off and state the fix as prose in --fix", why)
	}
	p.FixNew = proto.String(fixNew)
	p.FixBasis = proto.String("verified")
	return nil
}

// contains reports membership, so an --supersedes that already names the estopping gap
// lets the mint through: declaring the lineage IS the amendment the guard asks for.
func contains(xs []string, want string) bool {
	for _, x := range xs {
		if x == want {
			return true
		}
	}
	return false
}

// mintAsserted is what a gap states, and so what a retry under its key must state again: everything
// the seat says of the defect and where, and its argument for it. Not compared: --distinct-from,
// which answers the duplicate screen against the board as the seat read it — a retry rebuilt after
// a crash reads a board that has moved, and refusing it would send the seat to mint the same gap
// again under another key.
var mintAsserted = []seat.Asserted{
	{Flag: flags.Quote, Field: "location"},
	{Flag: flags.Reason, Field: "mint_reason"},
	{Flag: flags.Class, Field: "class"},
	{Flag: flags.Problem, Field: "problem"},
	{Flag: flags.Fix, Field: "required_fix"},
	{Flag: flags.New, Field: "fix_new"},
	{Flag: flags.Check, Field: "acceptance_check"},
	{Flag: flags.CheckKind, Field: "check_kind"},
	{Flag: flags.Severity, Field: "severity"},
	{Flag: flags.Likelihood, Field: "likelihood"},
	{Flag: flags.Impact, Field: "impact"},
	{Flag: flags.Complexity, Field: "complexity_cost"},
	{Flag: flags.AboutKind, Field: "about_kind"},
	{Flag: flags.About, Field: "about_ref"},
	{Flag: flags.Supersedes, Field: "supersedes"},
	{Flag: flags.FoundBy, Field: "found_by"},
}
