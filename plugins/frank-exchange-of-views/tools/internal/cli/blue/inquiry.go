package blue

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/enumhelp"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// line of inquiry: a line of inquiry, and what became of it.
//
// think-around-problem mandates exploring genuinely distinct alternatives before
// a significant decision. terse-communication forbids narrating options. So the
// exploration was required, invisible and unverifiable — a dead letter by the
// standard scorecards.md sets out, and the same shape as "confidence
// self-graded", which was mandated and practised five times in 1,892 lines.
//
// This is what instruments it. Exploration becomes a RECORD and the record
// becomes a report section, so the rule stops being self-attested and starts
// leaving artifacts.
//
// THREE STATUSES, each answering a different question:
//
//	declined   considered, not taken     — was the rejection REASONED, or decoration?
//	abandoned  pursued, then died        — what killed it? (the negative result)
//	pursued    became the report's spine — does the report reflect it honestly?
//
// `abandoned` is the most valuable and the most routinely lost: dead ends are
// exactly what a future run needs so it does not re-walk them, and they are
// direct feedstock for the sleeper service's subject mining. Nothing in the
// engine preserved them before this verb — they died in a seat's context.
//
// TWO FORMS, BECAUSE THE UNIT IS THE CHOICE AND NOT THE ENTRY (#246).
//
//	PROPOSE  --line --hypothesis [--status] [--method]   -> the tool assigns A1, A2 …
//	MOVE     --id A1 --status <new> --reason "<what changed>"
//
// The move form is the whole point. Measured over 86 line of inquiry events in six runs: ZERO were
// ever recorded twice and ZERO statuses ever changed, because there was no id and no update
// path. 83 of 86 landed in round 0 — so "pursued" meant "I intend to", nothing could ever
// falsify it, and a direction that died mid-run had no way to say so. The rejection rate
// that came out of that measured only what blue could rule out BEFORE STARTING.
//
// --hypothesis is what makes a later move checkable: a line of inquiry that says what would be true
// if it paid off can be abandoned against its own claim rather than on a shrug.
func newInquiry() *cobra.Command {
	c := seat.Prose(seat.New("line-of-inquiry",
		`PROPOSE a line of inquiry (--line "<approach>" --hypothesis "<what would be true if it pays off>" [--status] [--method]; the tool assigns its id), or MOVE one you already proposed (--id A1 --status proposed|pursued|declined|abandoned --reason "<what changed>")`,
		func(s seat.Context, cmd *cobra.Command) (seat.Result, error) {
			id, line := seat.Str(cmd, flags.ID), seat.Str(cmd, flags.Line)
			if id != "" && line != "" {
				return nil, fmt.Errorf("blue line of inquiry: --id MOVES a line of inquiry you already proposed and --line PROPOSES a new one — passing both leaves it ambiguous which you meant. Drop --line to move A%s, or drop --id to propose", "")
			}
			p := record.NewPayload()
			moving := id != ""
			if moving {
				// A move must name a real line of inquiry, and must say what changed: a status that
				// slides with no reason is exactly the unfalsifiable "pursued" this replaces.
				if err := record.RequireInquiryRef(s.RunDir, id); err != nil {
					return nil, err
				}
				if strings.TrimSpace(seat.Str(cmd, flags.Reason)) == "" {
					return nil, fmt.Errorf("blue line of inquiry: moving %s requires --reason — what you learned that changed its fate. The choosing is the evidence; a status that moves silently records none of it", id)
				}
				p.Set("inquiry_id", id).Set("supersedes_status", "1")

				// THE CONTEST MOVED OUT OF HERE, TO `motion direction appeal` (#344), AND A REAL
				// RUN IS WHY.
				//
				// `contests_ruling` was set here as a side effect of moving a line to `pursued`
				// against an adverse ruling. That coupling is the defect: it can only record
				// disagreement that WINS. In research/2026-08-10_pre-motion-real-record the merge
				// ruled a line too-thin, blue argued against that reasoning at the leaf, and then
				// declined the line anyway — the ordinary outcome of an argument — and the field
				// recorded nothing. `contests_ruling` appears zero times in the whole record.
				//
				// So the number it existed to produce ("how often does blue pursue a line red
				// ruled out") measured DEFIANCE, not disagreement, and every seat that argued and
				// yielded fell out of the count silently.
				//
				// An appeal is its own event with its own reason, filed against the ruling,
				// independent of what the line's status does next. The READ side of this field
				// stays forever — record/inquiry.go and compat.go still recover it from
				// pre-collapse records, where it is the only legacy spelling of an appeal.
			} else {
				if line == "" {
					return nil, fmt.Errorf("blue line-of-inquiry requires --line (the approach you are proposing) or --id (a line you are moving) — see `blue line-of-inquiry --help`")
				}
				fresh, err := record.MintInquiryID(s.RunDir)
				if err != nil {
					return nil, err
				}
				id = fresh
				p.Set("inquiry_id", id)
				seat.SetSame(cmd, p, flags.Line, flags.Method)
				seat.Set(cmd, p, "hypothesis", flags.Hypothesis)
			}
			// A fresh proposal with no stated fate is `proposed` — the state the old shape
			// could not express, which forced blue to declare a fate before it had one.
			status := seat.Str(cmd, flags.Status)
			if status == "" && !moving {
				status = "proposed"
			}
			p.Set("status", status)
			if err := seat.SetReason(cmd, p, "reason"); err != nil {
				return nil, err
			}
			if _, err := record.Append(s.Identity(), "line-of-inquiry", p); err != nil {
				return nil, err
			}
			return inquiryResult{ID: id, Status: status, Line: line, Moved: moving}, nil
		}))

	c.Flags().Var(flags.InquiryID().WithCheck(record.InquiryExists), flags.ID, "a line of inquiry you already proposed (A1, A2 …) whose fate you are MOVING; omit to propose a new one")
	c.Flags().String(flags.Line, "", "the question or approach you are proposing — what you are going to try")
	c.Flags().String(flags.Hypothesis, "", "what would be TRUE if this line pays off — the claim a later abandonment is judged against, so the fate is checkable rather than a shrug")
	// THE VALUES ARE NOT RE-LISTED HERE, and this was the one enumhelp.Flag caller of twelve
	// that re-listed them. The hand-written line carried FOUR of the five statuses — `deferred`
	// had been added to InquiryStatuses and never to this string — and glossed the four it did
	// carry differently from the enum ("put forward, undecided" against "put forward and not yet
	// resolved — the default, and the state that owes a move"). Two copies of one enum in a
	// single `--help`, and a seat reading the usage line rather than the Enumerated values block
	// below it could not learn that `deferred` exists at all.
	//
	// enumhelp renders every value with its own meaning from the record, so the usage line's job
	// is to say what the FIELD is for. That is what the other eleven callers do.
	enumhelp.Flag(c, flags.Status, record.MustEnum("line-of-inquiry", "status"), "the fate of this line of inquiry — omit on a fresh proposal (it defaults to `proposed`); REQUIRED when moving one with --id. Re-recording `pursued` with what you learned is a reaffirmation and settles the line for this round")
	c.Flags().String(flags.Method, "", "the source class or technique it belonged to, when that is what distinguishes it")
	return c
}

type inquiryResult struct {
	ID     string `json:"inquiry_id"`
	Status string `json:"status"`
	Line   string `json:"line,omitempty"`
	Moved  bool   `json:"moved,omitempty"`
}

func (r inquiryResult) Human() string {
	if r.Moved {
		return "line of inquiry " + r.ID + " moved to " + r.Status
	}
	return "line of inquiry " + r.ID + " recorded (" + r.Status + "): " + r.Line
}
