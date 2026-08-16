package merge

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/bluedoc"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/enumhelp"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// mint: put a gap on the board.
//
// The id is TOOL-assigned, sequential per round. Letting seats choose produced
// four different "R5-1"s in one round of run 3, and the collision class dies here
// rather than being policed downstream.
func newMint() *cobra.Command {
	var severity, likelihood, impact, cx flags.GradeValue
	var supersedes, foundBy flags.CSV

	c := seat.Prose(seat.New("mint",
		`mint a board gap (id is TOOL-assigned; --key <stable-label> makes retries idempotent): --class <slug>|--class-new <slug> --definition --neighbor --distinguisher, --location "..." --problem "..."|--reason --fix "..." --check "<acceptance check red runs at re-audit>" --severity/--likelihood/--impact/--cx <grade> [--supersedes R1-2,R1-7] [--found-by L5-F3,L6-F2]`,
		func(s seat.Context, cmd *cobra.Command) (seat.Result, error) {
			// Crash-retry idempotency: --key (the stable local label, e.g. the source
			// lens finding) makes a retried mint return the EXISTING id.
			prior, err := record.ExistingMintByKey(s.RunDir, s.SeatID, seat.Str(cmd, flags.Key))
			if err != nil {
				return nil, err
			}
			if prior != "" {
				return mintResult{GapID: prior, Idempotent: true}, nil
			}
			// The round comes from the seat's CONTEXT, not from re-reading its id.
			// Both answer the same today (Of passes record.RoundOf as the inference),
			// but a gap id is the run's primary public identifier — it is printed in
			// the report and referenced by --supersedes, --id and found_by — so it
			// must be minted from the FACT the dispatcher supplies, not from a guess
			// about a string's shape. The moment FEOV_ROUND is injected these diverge,
			// and this call site would have kept the guess (#348).
			gapID, err := record.MintGapID(s.RunDir, s.Round)
			if err != nil {
				return nil, err
			}
			text, err := seat.Reason(cmd)
			if err != nil {
				return nil, err
			}
			problem := seat.Str(cmd, flags.Problem)
			if problem == "" {
				problem = text
			}
			// RED'S ARGUMENT IS KEPT WHEN IT IS NOT THE PROBLEM STATEMENT ITSELF.
			//
			// It was discarded. `--reason` fell back to `--problem` only when problem was empty,
			// so the ordinary mint — which supplies both — recorded the problem and threw the
			// reasoning away, while `--reason`'s own help promises "the substance the report
			// renders and the other side answers". The write returned "minted R1-4" either way.
			//
			// MEASURED 2026-08-16 BY ASKING THE BENCH. Dispatched to a petition sitting about what
			// a required_fix may demand, it reported first among its missing things: "Red's closing
			// or reasoning — Why did red mint this gap? Do they believe blue DID run a search and
			// just didn't describe it? The gap's problem statement is clear, but red's intent for
			// the fix isn't visible to me." It was adjudicating red's intent with red's intent
			// unrecorded.
			//
			// `problem` says WHAT IS WRONG. This says why red thinks so, which is the half a seat
			// answers and a bench weighs. Stored only when the two differ, so a mint that passed
			// its problem through --reason does not carry it twice.
			p := record.NewPayload().Set("gap_id", gapID)
			if text != "" && text != problem {
				p.Set("mint_reason", text)
			}
			seat.Set(cmd, p, "mint_key", flags.Key)
			// --class-new both names the class and mints it, so it wins over --class.
			if seat.Given(cmd, flags.ClassNew) {
				p.Set("class", seat.Str(cmd, flags.ClassNew))
			} else {
				seat.Set(cmd, p, "class", flags.Class)
			}
			p.Set("class_new", seat.Given(cmd, flags.ClassNew))

			// THE LOCATION IS MATCHED AGAINST THE REPORT, and it was not.
			//
			// `lens finding --location` is REFUSED on a mis-quote — the anchor cannot be spliced
			// where the text is not. Mint's took prose, on the reasoning that a gap's location is
			// "recorded for a reader". Measured (#359): three gaps were minted naming report
			// sections that do not exist, two of them `--existence verified`. A seat asserted it
			// had checked a defect at the leaf, at a place the tool never confirmed was there.
			//
			// A location a reader cannot find is not a location. It is refused now, by the same
			// rule the lens has always been held to.
			//
			// AN OMISSION STILL HAS A PLACE. A gap about something MISSING has no span of its own
			// — so quote the sentence where it should be, which is exactly what a lens finding
			// about an omission already does. The refusal says so, because a seat that reads
			// "not found in report.md" and has an omission on its hands would otherwise conclude
			// the verb cannot express what it is holding.
			if loc := seat.Str(cmd, flags.Location); strings.TrimSpace(loc) != "" {
				report, err := record.ReadBlueReport(s.RunDir)
				if err != nil {
					return nil, err
				}
				if _, _, lerr := bluedoc.LocateUnique("merge mint --location", string(report), loc); lerr != nil {
					return nil, fmt.Errorf("%w\n\nQuote the exact sentence the defect lives at, from blue/report.md and nothing else — a section heading plus a sentence will not match. For a gap about something MISSING, quote the sentence where it SHOULD be; that is how a lens finding anchors an omission", lerr)
				}
			}
			seat.SetSame(cmd, p, flags.Definition, flags.Neighbor, flags.Distinguisher, flags.Location)
			p.Set("problem", problem)
			seat.Set(cmd, p, "required_fix", flags.Fix)
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
			// one, and the field becomes decoration — the same reason the finding label and the
			// gap id are tool-assigned rather than claimed.
			basis := "proposed"
			if seat.Given(cmd, flags.FixOld) || seat.Given(cmd, flags.FixNew) {
				fixOld, fixNew := seat.Str(cmd, flags.FixOld), seat.Str(cmd, flags.FixNew)
				if fixOld == "" || fixNew == "" {
					return nil, fmt.Errorf("merge mint: a concrete proposal needs BOTH --fix-old (the exact span you say is wrong) and --fix-new (the exact text it should become) — one alone is half a proposal blue cannot apply")
				}
				report, err := record.ReadBlueReport(s.RunDir)
				if err != nil {
					return nil, err
				}
				if err := bluedoc.ValidateProposal("merge mint", string(report), fixOld, fixNew); err != nil {
					return nil, err
				}
				p.Set("fix_old", fixOld)
				p.Set("fix_new", fixNew)
				basis = "verified"
			}
			p.Set("fix_basis", basis)
			seat.Set(cmd, p, "acceptance_check", flags.Check)
			seat.Set(cmd, p, "check_kind", flags.CheckKind)
			seat.SetGrade(p, "severity", &severity)
			seat.SetGrade(p, "likelihood", &likelihood)
			seat.SetGrade(p, "impact", &impact)
			seat.SetGrade(p, "complexity_cost", &cx)
			seat.SetList(p, "supersedes", &supersedes)
			seat.SetList(p, "found_by", &foundBy)

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
			// THE REJECTION IS LOGGED AS FRICTION, not merely refused. A guard that silently
			// blocks is invisible: this makes "how often does red relitigate its own
			// prescriptions" a per-run number on the record, which is the measurement of the
			// pathology the guard prevents. Same discipline as `blue cite`'s unreachable-source
			// rejection — the tool records the block, never the seat's memory of it.
			if board, berr := record.BoardState(s.RunDir); berr == nil {
				if prior, prescribed := record.EstoppelConflict(board, seat.Str(cmd, flags.Location)); prior != "" &&
					!contains(supersedes.Value(), prior) {
					msg := fmt.Sprintf("merge mint: estoppel — this gap's location is text YOU prescribed for %s and blue applied verbatim. The prescription is red's; raise it as an amendment to %s (argue it there, or mint with --supersedes %s so the lineage is explicit) rather than as a fresh gap against your own words. Prescribed text: %q",
						prior, prior, prior, prescribed)
					// The KIND is a field, not something a reader infers from the wording. The
					// prose above is aimed at a seat and must stay editable; the count an
					// operator reads must not move when it is edited (#283).
					fr := record.NewPayload().
						Set("text", msg).
						Set(record.FrictionKindKey, record.FrictionKindEstoppel).
						Set("estopped_by", prior)
					if _, ferr := record.Append(s.Identity(), "friction", fr); ferr != nil {
						return nil, ferr
					}
					return nil, errors.New(msg)
				}
			}

			if _, err := record.Append(s.Identity(), "mint", p); err != nil {
				return nil, err
			}
			if seat.Given(cmd, flags.ClassNew) {
				cn := record.NewPayload().Set("slug", seat.Str(cmd, flags.ClassNew))
				seat.SetSame(cmd, cn, flags.Definition, flags.Neighbor, flags.Distinguisher)
				if _, err := record.Append(s.Identity(), "class-new", cn); err != nil {
					return nil, err
				}
			}
			return mintResult{GapID: gapID}, nil
		}))

	c.Flags().String(flags.Key, "", "a stable local label (e.g. the source lens finding) making a retried mint idempotent")
	c.Flags().String(flags.Class, "", "the gap's class slug from the registry — what KIND of defect this is")
	c.Flags().String(flags.ClassNew, "", "mint a new class slug; requires --definition, --neighbor and --distinguisher")
	c.Flags().String(flags.Definition, "", "what the new class is, in one line")
	c.Flags().String(flags.Neighbor, "", "the existing class this one sits closest to")
	c.Flags().String(flags.Distinguisher, "", "the tie-break question that tells the two apart")
	c.Flags().String(flags.Location, "", "REQUIRED-IF-GIVEN — the EXACT sentence the defect lives at, quoted from blue/report.md and nothing else (matched, and refused if not found — for an omission, quote the sentence where it SHOULD be). Was prose until 0.63.0, when three gaps turned out to name sections that did not exist. against the report — it is recorded for a reader, so prose is fine here")
	c.Flags().String(flags.Problem, "", "what is wrong (or pass it via --reason)")
	c.Flags().String(flags.Fix, "", "the required fix, as prose — what must become true. This is the substantive channel: research it, enumerate it, qualify it")
	c.Flags().String(flags.FixOld, "", "OPTIONAL concrete proposal, TEXTUAL DEFECTS ONLY: the exact current span you say is wrong (must be present and unique in blue/report.md — a proposal you cannot state legally is one blue cannot apply). Requires --fix-new")
	c.Flags().String(flags.FixNew, "", "the exact text that span should become. Bounded: a replacement more than 120 characters longer than the span is refused as AUTHORING — a substantive addition is blue's to write, and you say so in --fix. Together these two derive fix_basis: verified")
	c.Flags().String(flags.Check, "", "the acceptance check red will RUN at re-audit — the pre-agreed contract, not a description")
	enumhelp.Flag(c, flags.CheckKind, record.MustEnum("mint", "check_kind"), ("what would SETTLE that check — read a document, RUN a computation, or verify a source. A `computation` check cannot be closed by prose: it closes only when a proof answers this gap"))
	c.Flags().Var(&severity, flags.Severity, flags.GradeUsage("how bad this is"))
	c.Flags().Var(&likelihood, flags.Likelihood, "how likely the CONSEQUENCE is — never how likely the defect is to BE there, which is what the grade meant before v2 split them")
	c.Flags().Var(&impact, flags.Impact, "how bad the consequence is if it lands")
	c.Flags().Var(&cx, flags.Complexity, "complexity_cost — what fixing it costs, on the same scale")
	c.Flags().Var(&supersedes, flags.Supersedes, "comma-separated ancestor ids this gap replaces; lineage is never dropped")
	c.Flags().Var(&foundBy, flags.FoundBy, "comma-separated lens findings that surfaced it (L5-F3,L6-F2)")
	return c
}

// mintResult reports the tool-assigned gap id; a retry returns the existing one.
type mintResult struct {
	GapID      string `json:"gap_id"`
	Idempotent bool   `json:"idempotent,omitempty"`
}

func (r mintResult) Human() string {
	if r.Idempotent {
		return "minted " + r.GapID + " (idempotent retry — existing id returned)"
	}
	return "minted " + r.GapID
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
