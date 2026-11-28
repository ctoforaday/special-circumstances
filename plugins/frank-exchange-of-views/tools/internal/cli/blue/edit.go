package blue

import (
	"errors"
	"fmt"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"slices"
	"strings"

	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchortext"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/bluedoc"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/claimcount"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/reportproj"
)

// edit: the write path for the report's text once the lanes' parts are recorded.
//
// It replaces the exact current span --old with --new. A rewrite preserves every anchor, of every kind
// in the anchor kinds table: an anchor the replacement leaves out is put back where its sentence
// survives word for word, once; otherwise it is REJECTED, naming that sentence. A pure cut takes the
// anchors of the sentences it cuts and records them (removed_anchors). Each applied edit appends a
// `blue_edit` event, a report op the render folds after the parts.
//
// PROVENANCE: --answers names the gap this edit responds to. It is validated against the
// board like every other reference (refs.go), and it is what makes `required_fix` and the
// change blue actually made JOINABLE — the pair every #267 measurement reads. Naming a real
// gap id in --reason while leaving --answers empty is REFUSED: the convention it replaces
// held 19 times in 26 and that is exactly the reliability a key may not have.
//
// Crash-safety is EVENT-FIRST: the event is appended only AFTER validation succeeds (so a
// mis-quote never lands a phantom op), then the write is applied; a crash between leaves the
// event durable and the write reconciled idempotently on retry — no wedge, no phantom op.
//
// --accept still requires --reason because accepting red's text is a decision, not a transcription:
// the argument for the change is blue's own.
func newEdit() *cobra.Command {
	c := seat.Prose(seat.New("edit", editResult{}, func(s seat.Context, cmd *cobra.Command) (seat.Result, error) {
		run, err := s.Run()
		if err != nil {
			return nil, err
		}
		reason, err := seat.Reason(cmd)
		if err != nil {
			return nil, err
		}
		if strings.TrimSpace(reason) == "" {
			return nil, feov.Errorf(feov.MissingField, "requires --reason: why this change (the argument red re-audits against)")
		}
		oldStr := seat.Str(cmd, flags.Quote)
		newStr := seat.Str(cmd, flags.New)
		gapID := seat.Str(cmd, flags.Answers)
		// --accept G NAMES THE GAP whose statement blue takes, and the edit answers it: one id, so
		// --answers beside it would be a second source for one fact.
		accepting := seat.Str(cmd, flags.Accept) != ""
		if accepting {
			if gapID != "" {
				return nil, feov.Errorf(feov.Validation, "--accept %s names the gap this edit answers — drop --answers, which would be a second source for one fact", seat.Str(cmd, flags.Accept))
			}
			gapID = seat.Str(cmd, flags.Accept)
		}

		// ACCEPTANCE: THE TOOL SUPPLIES THE TEXT, SO BLUE CANNOT MIS-TRANSCRIBE IT.
		//
		// The ordinary path asks blue to retype red's span and replacement and then decides, by
		// comparing bytes, whether it got them right. A stray space, a smart quote or a re-wrapped
		// line means blue did exactly what red asked while the record says it did not — and
		// estoppel, which exists because blue was once penalised for doing as it was told, is what
		// blue loses. Here the pair comes off the mint, so the agreement is structural.
		//
		// WHAT BLUE STILL SUPPLIES IS ITS REASON. Accepting is an act it should argue for, and
		// --reason is the argument red re-audits against; nothing about agreeing removes that.
		if accepting {
			if oldStr != "" || newStr != "" {
				return nil, feov.Errorf(feov.Validation, "--accept supplies --quote and --new for you, from the fix red recorded on %s — passing either would be a second source for one fact. Drop them, or drop --accept and write the edit yourself", gapID)
			}
		} else if err := typedPair(oldStr, newStr); err != nil {
			return nil, err
		}
		key := seat.Str(cmd, flags.Key)

		// Crash-retry: a committed blue_edit for this key means the op is already on the
		// stack — reconcile the write idempotently, do NOT append a second op.
		prior, err := record.ExistingBlueEditByKey(run, s.SeatID, key)
		if err != nil {
			return nil, err
		}
		if prior != nil {
			if err := repeatsEdit(key, prior, gapID, accepting, oldStr, newStr, reason); err != nil {
				return nil, err
			}
			// The op is already on the diff-stack, so the render already reflects it — there is
			// nothing to re-apply. Under report-as-record the edit IS the event; no file to reconcile.
			//
			// ADVISORY, so an idempotent retry carries it too — a seat re-running a crashed edit
			// should hear the same note, not lose it to the fast path. An accepted edit's text is the
			// fix red recorded, read here for the note alone.
			tells := spanVoiceTells(newStr)
			if accepting {
				if peek, err := record.RenderRun(run); err == nil {
					if _, fix, found, err := record.Proposal(run, gapID, peek); err == nil && found {
						tells = spanVoiceTells(fix)
					}
				}
			}
			return editResult{VoiceTells: tells, Idempotent: true}, nil
		}

		// FRESH: validated against the report the edit lands on, BEFORE the event commits, so a
		// mis-quote, or a rewrite that would drop, duplicate or invent an anchor, never lands a phantom
		// stack op. The report is the render
		// of the record so far — there is no file — and it is rendered inside the write's transaction:
		// another seat's anchor or blue's own parallel edit can commit in between, and an edit planned
		// against the report before it is an edit replay cannot apply, which fails every render of the
		// report from then on (#1252).
		body := &recordpb.BlueEdit{
			EditKey: proto.String(seat.Str(cmd, flags.Key)),
			Answers: proto.String(gapID),
			Text:    proto.String(reason),
		}
		var tells []string
		var bunched struct {
			report, old, applied string
			exact                bool
		}
		_, err = record.Append(s.Identity(), body, append([]record.Guard{record.Check(func(l *record.Landing) error {
			peek, err := l.Report()
			if err != nil {
				return err
			}
			var a Accepted
			if accepting {
				st, found, err := l.Statement(gapID)
				if err != nil {
					return err
				}
				if !found {
					return feov.Errorf(feov.NotFound, "--accept: no gap %s on the board", gapID)
				}
				loc, fix := st.Location, st.FixNew
				switch {
				case st.Withdrawn:
					return feov.Errorf(feov.Conflict, "--accept: gap %s's raiser stated no text in its closing, withdrawing every pair it offered, so there is nothing to accept. Write the edit yourself with --quote and --new", gapID)
				case st.Retire:
					return feov.Errorf(feov.Conflict, "--accept: gap %s's raiser states a retire of its anchor in its closing, not a pair: take it with `retire --accept %s`", gapID, gapID)
				}
				if strings.TrimSpace(fix) == "" {
					return feov.Errorf(feov.Conflict, "--accept: gap %s prescribes no concrete text, so there is nothing to accept — red raised it without proposing a fix. Write the edit yourself with --quote and --new", gapID)
				}
				a, err = PlanAccept(peek, gapID, loc, fix)
				// A PRESCRIPTION THAT CHANGES NOTHING IS NOT STALE. `lens mint` refuses to verify one, but
				// the report can move under a verified fix until it changes nothing, and calling that stale
				// would send blue to hunt for a change in the report that never happened.
				if errors.Is(err, reportproj.ErrNoChange) {
					return fmt.Errorf("--accept: the fix recorded on %s changes nothing — applied to the report as it stands, "+
						"it leaves every byte where it is (%w). There is nothing to accept. If the defect red named is still there, write the "+
						"edit yourself with --quote and --new; if it is not, argue that on the gap", gapID, err)
				}
				// THE ORDINARY PATH'S ADVICE IS UNACTIONABLE HERE. Those refusals tell blue to adjust
				// --quote or to copy an anchor into --new, and an accepting caller passes neither. The mint
				// was judged by this same planner over the report as it stood, so what refuses it now is a
				// change since: commonest, a citation blue placed on the sentence red located, which red's
				// text has no word-for-word sentence to put back on.
				if err != nil {
					return fmt.Errorf("--accept: the fix recorded on %s no longer applies to the report as it now stands — %w. "+
						"The prescription is stale rather than wrong: write the edit yourself with --quote and --new, carrying red's text and whatever has changed around it", gapID, err)
				}
			} else if cut, removed, isCut, err := planCut(peek, oldStr, newStr); isCut {
				// A PURE CUT (ruling 2): the tool decides it and stamps the anchors it takes, so judgeAct's
				// L6 and replay read the list the write applies.
				if err != nil {
					return err
				}
				a, body.RemovedAnchors = cut, removed
			} else if a, err = planTyped(peek, oldStr, newStr); err != nil {
				return err
			}
			oldStr, newStr, planned, applied, exact := a.Old, a.New, a.Planned, a.Applied, a.Exact
			// Nothing here can refuse: the tells are collected and ride back on the result.
			tells = spanVoiceTells(newStr)

			body.Old = proto.String(oldStr)
			body.New = proto.String(applied) // with any anchor the tool put back, so replay splices the same bytes
			// WHAT THIS EDIT REOPENED. An anchor is never lost unrecorded: a rewrite that drops one is
			// refused, droppedMarker backstops the whole edit, and a cut records each anchor it takes in
			// removed_anchors. But one that SURVIVES onto rewritten prose is a citation backing a
			// sentence nobody read, and nothing said so. Text moves through this verb alone,
			// so the no-loss computation and the requires-review mark belong on this one channel.
			//
			// Computed from the paragraphs the edit touched, before and after (TouchedParagraphs),
			// cut from the render the validation used, which is the report the edit lands on — and
			// stamped on the edit, its own record of what it changed.
			start, end, located := bluedoc.SpliceSpan(peek, oldStr, exact)
			if !located {
				return feov.Errorf(feov.Conflict, "the edit planned over the report and its span no longer locates in it by replay's rule: %q", oldStr)
			}
			oldParas, newParas := bluedoc.TouchedParagraphs(peek, planned, start, end)
			body.OldParagraphs, body.NewParagraphs = proto.String(oldParas), proto.String(newParas)
			body.Reopened = bluedoc.ReopenedAnchors(oldParas, newParas)
			if err := refuseBaredAnchors(peek, planned); err != nil {
				return err
			}
			// THE LITERAL SPAN, RECORDED. Replay cannot re-derive this choice — an edit recorded before
			// the field existed, whose trimmed span was a no-op, must still replay as one — so the event
			// says which span it replaced. Absent, not false, on every ordinary edit.
			if exact {
				body.ExactSpan = proto.Bool(true)
			}
			// ESTOPPEL, RECORDED BY THE TOOL COMPARING BYTES (#267 stage 4).
			//
			// If blue applied red's own proposed text EXACTLY, there is nothing left for red to
			// complain about at this site BY CONSTRUCTION, not by good behaviour. The fact is
			// computed here — never claimed by either seat — and `lens mint` reads it to refuse
			// a fresh gap relitigating text red itself prescribed.
			//
			// Blue is not obliged to reach this state: a counter-edit simply does not set the
			// flag, and record.DeclineStats counts that as blue exercising its right to disagree.
			if accepting {
				// BY CONSTRUCTION, NOT BY COMPARISON. The bytes came from the mint a few lines above,
				// so there is nothing to compare and nothing a transcription slip could cost. Both
				// facts are recorded: `accepted` says blue took red's text through the accept path,
				// `applied_verbatim` says the text is red's — the second is what estoppel reads, and
				// it must be true here for the same reason it is true after a perfect transcription.
				body.Accepted = proto.Bool(true)
				body.AppliedVerbatim = proto.Bool(true)
			} else if gapID != "" {
				verbatim, err := l.ProposalAppliedVerbatim(gapID, peek, oldStr, newStr)
				if err != nil {
					return err
				}
				if verbatim {
					body.AppliedVerbatim = proto.Bool(true)
				}
			}
			// Red's own text, taken whole, places the gap's anchor where the tool puts it (record.Proposal).
			// A bunched run is refused with what the record says each anchor marks, worded once the
			// write has rolled back (bunchedRefusal).
			if !body.GetAppliedVerbatim() {
				if ids, _, _ := bunchedRun(peek, oldStr, applied, exact); ids != nil {
					bunched.report, bunched.old, bunched.applied, bunched.exact = peek, oldStr, applied, exact
					return errBunched
				}
			}
			return nil
		})}, roadsHold(editRemedy)...)...)
		if errors.Is(err, errBunched) {
			return nil, bunchedRefusal(run, bunched.report, bunched.old, bunched.applied, bunched.exact)
		}
		if err != nil {
			return nil, err
		}
		return editResult{VoiceTells: tells, Removed: body.GetRemovedAnchors()}, nil
	}))

	c.Flags().String(flags.Key, "", flags.DescKey+". A repeat under the handle with another --quote, --new, --answers, --accept or --reason is refused: a handle names one edit")
	// --quote is refused when absent UNLESS --accept supplies it, so its marker states the
	// condition rather than the bare word — the bare marker would tell an accepting seat to type
	// what the tool fills in.
	flags.Text(c, flags.Quote, flags.DescQuote+". In a rewrite, an anchor standing at its end stays where it is unless the quote carries it; a cut of whole sentences takes their anchors, those at the end included")
	seat.RequiredUnlessGiven(c, flags.Quote, flags.Accept)
	flags.Text(c, flags.New, "the text that span should become. Where an anchor the quote leaves out stands at the span's end, leave the closing punctuation off: a replacement that ends its own sentence there is refused")
	c.Flags().Var(flags.GapID().WithCheck(record.GapExists), flags.Answers, "the gap id this edit responds to — the provenance join key; omit only for an edit that answers no gap")
	c.Flags().Var(flags.GapID().WithCheck(record.GapExists), flags.Accept, flags.DescAccept)
	// The handler refuses an edit with no argument; the marker says so where the seat reads.
	return seat.SaysRequired(c, flags.Reason)
}

// refuseBaredAnchors refuses a rewrite that would leave an anchor BARE — no prose before it in its
// sentence — where it stood on prose before (claimcount.BareAnchorIDs).
//
// NO ANCHOR STANDS BARE (gblock 2026-11-20: an anchor "can't just be floating on nothing"). A rewrite
// that would leave a token of any kind on a sentence with no words is refused, carried or not: an
// anchor marks a sentence, and one on nothing marks none. A cut whose replacement holds nothing of the
// sentence takes its tokens itself (removed_anchors), so the refusal names the two routes: keep words
// on the sentence, or cut it whole. after is the planned report with a cut's removes applied, so a
// cut's own tokens are never read bare here.
func refuseBaredAnchors(before, after string) error {
	bared := BaredAnchors(before, after)
	if len(bared) == 0 {
		return nil
	}
	what, it, its := "anchor "+bared[0], "it", "its"
	if len(bared) > 1 {
		what, it, its = "anchors "+strings.Join(bared, ", "), "them", "their"
	}
	return feov.Errorf(feov.Validation,
		"this edit would leave %s standing bare — no prose before %s in %s sentence — and an anchor marks a sentence, so one on nothing marks none. Nothing is recorded: "+
			"keep words on that sentence in --new, or cut it whole — a replacement holding nothing of it — which takes %s with it",
		what, it, its, it)
}

// BaredAnchors is THE judgement of which anchors a rewrite leaves on a sentence with no words
// (refuseBaredAnchors): every token it leaves so, of every kind, carried or not. before and after are
// the report the edit plans over and the report it plans. `lens mint` asks it of a prescription over
// the report the mint lands on, so a fix the accept would refuse for this is refused where red writes
// it.
func BaredAnchors(before, after string) []string {
	was := map[string]bool{}
	for _, id := range claimcount.BareAnchorIDs(before) {
		was[id] = true
	}
	var bared []string
	for _, id := range claimcount.BareAnchorIDs(after) {
		if !was[id] {
			bared = append(bared, anchor.Token(id))
		}
	}
	return bared
}

// holdsNothing is whether a replacement holds nothing of a sentence: the empty string, or only
// whitespace and sentence terminators.
func holdsNothing(new string) bool { return strings.Trim(new, " \t\r\n.!?;:…") == "" }

// planCut is THE decision of a PURE CUT (ruling 2): the located span covers every word of one or more
// whole sentences, and the replacement holds nothing of them (holdsNothing). Every other edit is a
// rewrite, and isCut is false for it — a span holding no prose, a replacement with words, or a span
// whose sentence keeps prose outside it.
//
// The cut widens to the cut sentences' abutting tokens, whether or not the quote carries them: it
// takes every token the splice leaves on no words, in reading order, and lists them (removed_anchors).
// The report it plans is replay's: the splice, which keeps every token the replacement does not carry
// (reportproj.ApplySplice), then the list's removes (reportproj.RemoveListed). A token of the span that
// still stands on words after the splice means the span did not cover its sentence: that edit is a
// rewrite, refused by the dropped-token rule.
func planCut(report, old, new string) (a Accepted, removed []string, isCut bool, err error) {
	if !holdsNothing(new) {
		return Accepted{}, nil, false, nil
	}
	start, end, lerr := bluedoc.LocateUniqueReplacing(report, old)
	if lerr != nil || !claimcount.HasProse(report[start:end]) {
		return Accepted{}, nil, false, nil
	}
	spliced := reportproj.ApplySplice(report, start, end, new)
	was := map[string]bool{}
	for _, id := range claimcount.BareAnchorIDs(report) {
		was[id] = true
	}
	for _, id := range claimcount.BareAnchorIDs(spliced) {
		if !was[id] {
			removed = append(removed, id)
		}
	}
	for _, id := range anchor.IDs(report[start:end]) {
		if !slices.Contains(removed, id) {
			return Accepted{}, nil, false, nil
		}
	}
	if len(removed) == 0 {
		return Accepted{}, nil, false, nil
	}
	planned, err := reportproj.RemoveListed(spliced, removed)
	if err != nil {
		return Accepted{}, nil, true, feov.Wrap(feov.Internal, err, "internal error — the cut's plan could not take the anchors it lists (report unchanged): %v", err)
	}
	if dropped := droppedMarker(report, planned, removed); dropped != "" {
		return Accepted{}, nil, true, feov.Errorf(feov.Internal, "internal error — this edit would drop %s (report unchanged)", anchor.Label(dropped))
	}
	if run := reportproj.DoubledTerminator(report, planned); run != "" {
		return Accepted{}, nil, true, feov.Errorf(feov.Validation, "this cut would leave %q in the report — a terminator beside another, where the cut meets the text it keeps. Quote the sentences with their own terminators, exactly as `show report` prints them", run)
	}
	a = Accepted{Old: old, New: new, Planned: planned, Applied: new}
	return a, removed, true, composedRefusal(report, planned)
}

// repeatsEdit refuses a keyed retry that states another edit than the one stored under its key: the
// gap it answers, whether it accepts that gap's fix, its reason, and — for an edit blue typed — the span and its
// replacement. AN EDIT PLACES NO ANCHOR, so its --quote is not a location a retry finishes at: it is
// half of the change, and a different span is a different edit. The stored replacement carries any
// anchor the tool put back, so both replacements are compared without their anchors. An accepting
// retry's pair is the tool's, read off the gap, and is not the seat's to repeat.
func repeatsEdit(key string, stored *recordpb.BlueEdit, gapID string, accepting bool, oldStr, newStr, reason string) error {
	retry := &recordpb.BlueEdit{Answers: proto.String(gapID), Old: proto.String(oldStr), New: proto.String(withoutAnchors(newStr)), Text: proto.String(reason)}
	if accepting {
		retry.Accepted = proto.Bool(true)
	}
	view := proto.Clone(stored).(*recordpb.BlueEdit)
	view.New = proto.String(withoutAnchors(stored.GetNew()))
	asserted := []seat.Asserted{{Flag: flags.Answers, Field: "answers"}, {Flag: flags.Accept, Field: "accepted"}, {Flag: flags.Reason, Field: "text"}}
	if !accepting {
		asserted = append(asserted, seat.Asserted{Flag: flags.Quote, Field: "old"}, seat.Asserted{Flag: flags.New, Field: "new"})
	}
	return seat.RepeatsKeyed(key, "edit", "", view, retry, asserted...)
}

// withoutAnchors is s with every anchor token taken out.
func withoutAnchors(s string) string {
	return anchor.Replace(s, func(string, string) string { return "" })
}

// typedPair refuses an edit pair that names no span, or that changes nothing.
func typedPair(old, new string) error {
	if old == "" {
		return feov.Errorf(feov.MissingField, "requires --quote: the EXACT current span to replace (matched across the invisible anchors, like the Edit tool)")
	}
	if old == new {
		return feov.Errorf(feov.Validation, "--old and --new are identical — no change to make")
	}
	return nil
}

// Accepted is the edit `--accept` makes of a gap's prescription: the pair it sends, the report it
// plans, the replacement with any anchor the tool put back, and whether it took the literal span.
type Accepted struct {
	Old, New, Planned, Applied string
	Exact                      bool
}

// PlanAccept is THE judgement of whether a gap's prescription applies: loc and fix as its mint records
// them, over report with the gap's anchor standing in it. `edit --accept` applies the edit it plans,
// and `lens mint` records a fix verified exactly where it plans one over the report the mint lands
// on, its anchor placed — so a verified fix is one the accept applies, and one the accept would refuse
// is refused at the mint. Two judges of one pair are how a mint came to verify fixes every accept
// refused (m25: 0 of 3 applied).
//
// THE PAIR CARRIES THE GAP'S TOKEN, which the report holds once by construction (record.ProposalOver
// locates the span by it). Where the gap's sentence is words the report holds more than once — a
// reading's gap is placed by its citation's identity, not by text — the ordinary locate, which skips
// the anchor layer, finds every copy and refuses the pair as ambiguous; the accept then takes the
// span byte for byte (planExact), the cited occurrence alone, and records it as an exact span that
// replay reads and never re-decides.
func PlanAccept(report, gapID, loc, fix string) (Accepted, error) {
	old, new := record.ProposalOver(report, gapID, loc, fix)
	if err := typedPair(old, new); err != nil {
		return Accepted{}, err
	}
	if _, _, err := anchortext.LocateOnce(report, old, anchortext.CrossParagraphs); errors.Is(err, anchortext.ErrAmbiguous) && strings.Contains(old, anchor.Token(gapID)) {
		if a, ok, err := planExact(report, old, new); ok {
			return a, err
		}
	}
	return planTyped(report, old, new)
}

// AcceptPair is the edit `--accept` makes of a gap's pair loc/fix, planned with PlanAccept over the
// report l lands on, written into body: the span, the replacement with the anchors the tool put back,
// the anchors it reopened, the literal span where it took one, and that the text is red's, applied
// verbatim. The bench's answer for red (or for the assignee) applies red's closing's road through it,
// in the ruling's transaction.
func AcceptPair(l *record.Landing, body *recordpb.BlueEdit, gapID, loc, fix string) error {
	report, err := l.Report()
	if err != nil {
		return err
	}
	a, err := PlanAccept(report, gapID, loc, fix)
	if err != nil {
		return err
	}
	if err := refuseBaredAnchors(report, a.Planned); err != nil {
		return err
	}
	body.Old, body.New = proto.String(a.Old), proto.String(a.Applied)
	// Stamped as every edit's write stamps it: the paragraphs it touched, before and after, and what
	// it reopened read from them.
	start, end, located := bluedoc.SpliceSpan(report, a.Old, a.Exact)
	if !located {
		return feov.Errorf(feov.Conflict, "the edit planned over the report and its span no longer locates in it by replay's rule: %q", a.Old)
	}
	oldParas, newParas := bluedoc.TouchedParagraphs(report, a.Planned, start, end)
	body.OldParagraphs, body.NewParagraphs = proto.String(oldParas), proto.String(newParas)
	body.Reopened = bluedoc.ReopenedAnchors(oldParas, newParas)
	if a.Exact {
		body.ExactSpan = proto.Bool(true)
	}
	body.Accepted, body.AppliedVerbatim = proto.Bool(true), proto.Bool(true)
	return nil
}

// planExact is the edit of old by new over report at old's literal span (bluedoc.LocateLiteral),
// held to the checks every planned edit meets — anchors transit a rewrite unchanged, no marker dropped
// that the cut's list does not name, no terminator doubled, a change made, a composed document — and false where old does not occur
// byte for byte exactly once.
func planExact(report, old, new string) (Accepted, bool, error) {
	start, end, _, ok := bluedoc.LocateLiteral(report, old)
	if !ok {
		return Accepted{}, false, nil
	}
	applied, err := bluedoc.AnchorsTransitUnchanged(report[start:end], new)
	if err != nil {
		return Accepted{}, true, err
	}
	planned := reportproj.ApplySplice(report, start, end, applied)
	if planned == report {
		return Accepted{}, true, &reportproj.NoChangeError{}
	}
	if dropped := droppedMarker(report, planned, nil); dropped != "" {
		return Accepted{}, true, feov.Errorf(feov.Internal, "internal error — this edit would drop %s (report unchanged)", anchor.Label(dropped))
	}
	if run := reportproj.DoubledTerminator(report, planned); run != "" {
		return Accepted{}, true, feov.Errorf(feov.Validation, "this replacement would leave %q in the report — a terminator beside another, where the edit meets the text it keeps", run)
	}
	a := Accepted{Old: old, New: new, Planned: planned, Applied: applied, Exact: true}
	return a, true, composedRefusal(report, planned)
}

// planTyped is the edit of old by new over report, as every edit plans it.
func planTyped(report, old, new string) (Accepted, error) {
	planned, applied, exact, err := validateEdit(report, old, new)
	if err == nil {
		err = composedRefusal(report, planned)
	}
	return Accepted{Old: old, New: new, Planned: planned, Applied: applied, Exact: exact}, err
}

// errBunched is the edit guard's refusal of a bunched run (bunchedRun); the verb words it with
// bunchedRefusal once the write has rolled back.
var errBunched = feov.Errorf(feov.Validation, "the replacement carries a run of anchors to the end of its last sentence")

// planEdit is the PURE core: it computes the new report from replacing the span of `old`
// with `new`.
//
// The three LEGALITY checks — present-and-unique, no word split, and anchors transit a rewrite
// unchanged, no marker dropped that the cut's list does not name — live in internal/bluedoc, because `lens mint` now has to answer the same
// question about a concrete proposed fix before red may attach one. What stays here is the
// part only blue does: the SPLICE. Under report-as-record no file is written — the BlueEdit event
// IS the mutation and reportproj.Render replays this same splice. The validation peek reuses this
// via validateEdit. Fuzzed directly (edit_fuzz_test.go).
//
// applied is new with the anchors PlanSplice put back, which the caller records; exact reports that
// the edit took the quote AS WRITTEN rather than the trimmed span — the decision
// reportproj.PlanSplice makes, which the caller records as exact_span.
func planEdit(report, old, new string) (string, string, bool, error) {
	// ANCHORS TRANSIT A REWRITE — never created, destroyed or duplicated by one; a cut takes only what
	// its list names.
	//
	// This guard used to REJECT any span containing an anchor ("edit around it"). Combined with
	// the uniqueness guard that produces a DEADLOCK, demonstrated: when a word appears twice and
	// the only disambiguating context carries red's anchor, the minimal quote is refused as
	// ambiguous and the contextual quote is refused as anchor-spanning. The anchored occurrence —
	// the one red actually flagged — becomes uneditable, while the unanchored one edits fine. And
	// 71% of anchored quotes in the smoke had their anchor mid-span, so this is the common shape,
	// not a corner. PlanSplice runs AnchorsTransitUnchanged on whichever span it takes.
	next, applied, exact, err := reportproj.PlanSplice(report, old, new)
	if err != nil {
		return "", "", false, err
	}
	if dropped := droppedMarker(report, next, nil); dropped != "" {
		return "", "", false, feov.Errorf(feov.Internal, "internal error — this edit would drop %s (report unchanged)", anchor.Label(dropped))
	}
	return next, applied, exact, nil
}

// validateEdit rejects a mis-quote, or a rewrite that would drop, duplicate or invent an anchor,
// against a snapshot, WITHOUT mutating — so no event is recorded for an edit that cannot apply. It RETURNS the planned
// document so the caller can record what this edit reopens, computed from the same snapshot the
// validation used rather than from a re-read that may have moved, and whether the edit took the
// literal span, which the caller records as exact_span.
//
// AN EDIT THAT CHANGES NOTHING IS REFUSED (reportproj.ErrNoChange, raised by PlanSplice), for the
// reason identical --quote and --new already are — but this one gets past that check. Measured in
// #861's arm-B rerun: blue quoted `on their own.).` to drop the stray period, the trim left that
// period OUTSIDE the span, the span was replaced with itself, and the verb said "blue edit
// recorded". That shape now takes the literal span and applies; what still reaches the refusal is
// a no-op the literal quote cannot rescue, and the refusal says which kind it is.
func validateEdit(report, old, new string) (string, string, bool, error) {
	planned, applied, exact, err := planEdit(report, old, new)
	if err != nil {
		return "", "", false, err
	}
	// A RUN THAT STAYS PUT, AFTER A REPLACEMENT THAT ENDS ITS OWN SENTENCE (gblock, 2026-10-10).
	// The quote left the run out, so it stands after the replacement; the seam tidy takes the
	// replacement's terminator and the run lands on the replacement's last sentence, whichever
	// sentence held its claims. Measured on universe-m19: all four stay-put edits whose --new ended
	// its sentence meant something other than a rewording (one left a gap on the wrong
	// sentence), and the nine without one were clean. The sentence rule is anchor.Sentences, read
	// over the splice before the tidy: the run opens a sentence of its own exactly when the
	// replacement closed the one before it.
	//
	// The pair --accept sends, and `lens mint` judges, carries the run holding the gap's anchor
	// (record.ProposalOver), so it is never stay-put.
	// A literal span holds the terminator, and a run past a terminator is on no sentence of the span.
	if start, end, err := bluedoc.LocateUniqueReplacing(report, old); err == nil && !exact {
		if run, punct := bluedoc.AbuttingRun(report, end); run != "" && punct == 0 && claimcount.HasProse(applied) {
			at := start + len(applied)
			for _, sp := range anchor.Sentences(report[:start] + applied + report[end:]) {
				if sp[0] != at {
					continue
				}
				var labels []string
				for _, id := range anchor.IDs(run) {
					labels = append(labels, anchor.Label(id))
				}
				return "", "", false, feov.Errorf(feov.Validation, "your quote stops before %s. An anchor the quote leaves out stays where it stands, on the text that precedes it, "+
					"and your replacement ends its own sentence, which would leave every anchor named here on that sentence, whichever sentence held the claim it marks. "+
					"To reword the sentence, leave its closing punctuation off --new. To move an anchor, quote it as `show report` prints it and write it where it goes; to take its claim out, cut its sentence whole, which takes the anchor with it",
					strings.Join(labels, ", "))
			}
		}
	}
	if run := reportproj.DoubledTerminator(report, planned); run != "" {
		return "", "", false, feov.Errorf(feov.Validation, "this replacement would leave %q in the report — a terminator beside another, where the edit meets the text it keeps. "+
			"A quote's TRAILING punctuation is trimmed before the span is located, so the span your --quote names stops SHORT of the "+
			"terminator; replacing it with text that carries its own terminator leaves the original one standing after it. This is "+
			"how a repair makes the document strictly worse while reading as applied. The quote as written, punctuation included, "+
			"is used instead only when it occurs exactly once in the report: quote it exactly as the report prints it, with enough "+
			"of the text before it to be unique, or leave the terminator out of --new", run)
	}
	return planned, applied, exact, nil
}

// bunchedRun is the anchor run a replacement carries from the end of the text it replaces to the end
// of its own last sentence, when it holds MORE sentences than that text — with those sentences.
//
// A run at a sentence's end says which sentence its claims are on and nothing finer, and a split is
// the first moment a finer answer is needed. Measured (gblock's ruling, 2026-10-10, and the
// elicitation that refined it): blue reads every anchor in place and copies the run after the last
// sentence it writes — universe-m19 ev400 (one sentence to two, four anchors, the citation's
// footnote moved onto a sentence its source does not back), universe-m18 ev263 (one to two, three)
// and ev267 (one to four, the five-anchor end run).
//
// The smallest form that holds all three and none of the corpus's other 115 anchor-carrying edits:
// the sentences are anchor.Sentences' that hold prose; the run is two anchors or more with no
// prose after them in the replaced text's last sentence; every one of them stands the same way in
// the replacement's last sentence; and that sentence is not the old one word for word. An anchor
// the replacement writes anywhere else, a sentence count that does not grow, and a last sentence
// kept as it was all apply — the last is how a seat says the run belongs where it stands: reword
// the sentence in one edit, add the others in a second that keeps it.
func bunchedRun(report, old, applied string, exact bool) (run []string, was int, now []string) {
	start, end, err := bluedoc.LocateUniqueReplacing(report, old)
	if exact {
		start, end, _, _ = bluedoc.LocateLiteral(report, old)
	} else if err != nil {
		return nil, 0, nil
	}
	prose := func(s string) (out []string) {
		for _, sp := range anchor.Sentences(s) {
			if claimcount.HasProse(s[sp[0]:sp[1]]) {
				out = append(out, s[sp[0]:sp[1]])
			}
		}
		return out
	}
	atEnd := func(s string) (ids []string) {
		anchor.Each(s, func(_, stop int, id string) {
			if !claimcount.HasProse(s[stop:]) {
				ids = append(ids, id)
			}
		})
		return ids
	}
	before, now := prose(report[start:end]), prose(applied)
	if len(before) == 0 || len(now) <= len(before) || quoteCore(before[len(before)-1]) == quoteCore(now[len(now)-1]) {
		return nil, 0, nil
	}
	run = atEnd(before[len(before)-1])
	stands := atEnd(now[len(now)-1])
	for _, id := range run {
		if !slices.Contains(stands, id) {
			return nil, 0, nil
		}
	}
	if len(run) < 2 {
		return nil, 0, nil
	}
	return run, len(before), now
}

// bunchedRefusal refuses bunchedRun's edit, printing beside each anchor of the run what the record
// says it is about — the one thing that tells a seat which sentence holds its claim. Never the
// anchor's stored location: in every measured case that is the whole old sentence, the same words
// for every anchor of the run.
func bunchedRefusal(run record.Run, report, old, applied string, exact bool) error {
	ids, was, now := bunchedRun(report, old, applied, exact)
	if ids == nil {
		return nil
	}
	fam, err := record.FamilyOf(run)
	if err != nil {
		return err
	}
	computations, err := record.RecordedComputations(run)
	if err != nil {
		return err
	}
	line := func(s string) string {
		r := []rune(strings.Join(strings.Fields(claimcount.StripAnchors(s)), " "))
		if len(r) > 200 {
			return string(r[:200]) + "…"
		}
		return string(r)
	}
	var b strings.Builder
	fmt.Fprintf(&b, "your replacement holds %d sentences where the text it replaces holds %d, and it carries the %d anchors that stood at the end of that text to the end of its last sentence. "+
		"An anchor belongs on the sentence that holds its claim, and a run at a sentence's end does not say which sentence that is now. What each one marks, from the record:\n", len(now), was, len(ids))
	for _, id := range ids {
		about := ""
		switch anchor.Kind(id) {
		case "gap":
			if g := fam.Gap(id); g != nil && g.Mint != nil {
				about = g.Current().GetProblem()
			}
		case "computation":
			for _, p := range computations {
				if p.Label == id {
					about = p.Reason
				}
			}
		default:
			if at, err := record.EvidenceAt(run, id); err == nil && at.Source != nil {
				about = citationAbout(at.Source)
			}
		}
		if about = line(about); about == "" {
			about = "the record holds no text for it"
		}
		fmt.Fprintf(&b, "  %s %s: %s\n", anchor.Kind(id), anchor.Token(id), about)
	}
	b.WriteString("Your replacement's sentences:\n")
	for i, s := range now {
		fmt.Fprintf(&b, "  %d. %s\n", i+1, line(s))
	}
	fmt.Fprintf(&b, "Write each anchor after the sentence that now holds its claim. Where every one of them belongs on sentence %d, make the change in two edits: "+
		"reword that sentence, its anchors kept, in one, and add the other sentences in a second that keeps it word for word", len(now))
	return feov.Errorf(feov.Validation, "%s", b.String())
}

// droppedMarker returns an anchor id present in before, absent from after, and not in the edit's
// removed_anchors, or "" — so no edit drops an anchor it does not record. after is the planned report
// after the list's removes, the same after refuseBaredAnchors reads.
func droppedMarker(before, after string, removed []string) string {
	have := map[string]bool{}
	for _, id := range claimcount.ProtectedAnchorIDs(after) {
		have[id] = true
	}
	for _, id := range claimcount.ProtectedAnchorIDs(before) {
		if !have[id] && !slices.Contains(removed, id) {
			return id
		}
	}
	return ""
}

type editResult struct {
	Idempotent bool `json:"idempotent,omitempty"`
	// Removed is what a pure cut took out of the report, as the edit records it (removed_anchors):
	// the act names what it removed, so the seat reads it here rather than by diffing the report.
	// Empty on a rewrite, which carries every anchor of its span.
	Removed []string `json:"removed_anchors,omitempty"`
	// VoiceTells is ADVICE, and the edit is already recorded by the time it is rendered. The
	// report is written for a reader of its SUBJECT, and this names where the new text sounds
	// like the run talking about itself. It does not refuse: a pattern cannot tell a report
	// narrating its own construction from a report quoting a source that narrates something,
	// and a gate that cannot tell the difference would silently cost the second one. Red's
	// voice lens has the judgement; this only says where to look, while the seat is still here.
	VoiceTells []string `json:"voice_tells,omitempty"`
}

func (r editResult) Human() string {
	head := "blue edit recorded — a rewrite, appended to the record carrying every anchor its span held; the report is re-derived on read"
	if len(r.Removed) > 0 {
		head = "blue edit recorded — a cut, appended to the record: it took " + strings.Join(r.Removed, ", ") + " out of the report with the sentences it cut, under your reason; the report is re-derived on read"
	}
	if r.Idempotent {
		head = seat.Repeated("blue edit")
	}
	if len(r.VoiceTells) == 0 {
		return head
	}
	return head + "\n\nNOTE — this text sounds like the run rather than the subject. The edit is\nrecorded; this is not a refusal, and it may be wrong:\n  - " +
		strings.Join(r.VoiceTells, "\n  - ") +
		"\nSeparation, never deletion: an operational limit belongs in the log, and\nthe part that limits the CONCLUSION stays here, re-voiced."
}

// citationAbout is what a citation backs, for the refusal that asks where its anchor goes: the
// source's own words at its span where the tool found one, and otherwise the citer's reason.
func citationAbout(src *record.EvidenceSourceJSON) string {
	switch {
	case src.SourceText != nil:
		return src.Title + " — the source: \"" + src.SourceText.Matched + "\""
	case src.BlueReason != nil:
		return src.Title + " — " + *src.BlueReason
	case len(src.Verified) > 0:
		return src.Title + " — " + src.Verified[0].Text
	}
	return src.Title
}
