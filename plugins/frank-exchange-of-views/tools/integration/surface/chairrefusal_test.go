package surface

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// THE CONVERGENCE REFUSAL NAMES ONLY ACTS ON THE CHAIR'S SURFACE.
//
// The chair is the one seat with `verdict`, so the FAIL refusal over a converged board has exactly
// one reader. It told that reader to "raise something material — a material finding … minted as a
// gap": a lens's `finding` and `mint`, on no surface the chair holds, and with no chair route to ready
// a lens. The chair could only log the request or work around it, and a refusal is the moment a seat
// is most likely to do as it is told.
//
// IT READS THE RENDERED REFUSAL over a real converged board — once bare, and once with the chair's
// own item (the inquiry review) folded into it, since that item's work-list line is spliced in — and
// asks the chair's real command tree. Two matchers:
//   - every backticked act resolves on the chair's tree (a `--flag` resolves on `verdict`, the verb
//     refused), and at least two do, so a reworded refusal that names nothing fails here instead of
//     passing empty;
//   - no sentence that tells the chair what to do — one naming an act in backticks, as a refusal's
//     instructions do — names in prose an act only another surface holds: each other root's command
//     path, its last word in any inflection (`ocr read` is two words, `mint` is one), unless the
//     word is a chair verb itself (`closing` is the chair's; `close` is not). A sentence stating the
//     board ("this epoch minted no fresh material gap") instructs nobody and is not read.
//
// The tree is the authority — cli.AllRoots — never a list of who holds what.
func TestTheConvergenceRefusalNamesOnlyTheChairsActs(t *testing.T) {
	roots := cli.AllRoots()
	chair := roots["chair"]
	if chair == nil {
		t.Fatal("cli.AllRoots has no chair tree")
	}
	chairWords := map[string]bool{}
	walk(chair, nil, func(path []string, _ *cobra.Command) {
		for _, w := range path {
			chairWords[w] = true
		}
	})
	// THE OTHER SURFACES' ACTS: every leaf path another root holds whose last word the chair's tree
	// does not, keyed by the path a sentence would have to say.
	foreign := map[string][]string{}
	for role, root := range roots {
		if role == "chair" {
			continue
		}
		walk(root, nil, func(path []string, c *cobra.Command) {
			if c.HasSubCommands() || chairWords[path[len(path)-1]] {
				return
			}
			key := strings.Join(path, " ")
			foreign[key] = append(foreign[key], role)
		})
	}
	if len(foreign) < 5 {
		t.Fatalf("only %d act(s) found on the other surfaces — the walk is reading nothing", len(foreign))
	}

	for _, tc := range []struct {
		name string
		add  func(b *board)
	}{
		{"converged, nothing else", func(*board) {}},
		{"converged, the chair's own inquiry review owed", func(b *board) {
			b.add("blue-respond", &recordpb.Inquiry{InquiryId: proto.String("Q1"),
				Status: recordpb.InquiryStatus_INQUIRY_STATUS_PROPOSED.Enum(), Question: proto.String("an inquiry")})
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			b := convergedBoard(t)
			tc.add(b)
			run := b.seed()
			_, err := record.Append(record.Identity{Run: run, SeatID: "red-chair"},
				&recordpb.Gate{Verdict: recordpb.Verdict_VERDICT_FAIL.Enum()})
			if err == nil || !strings.Contains(err.Error(), "the board has converged") {
				t.Fatalf("fixture: FAIL over the converged board = %v, want the convergence refusal", err)
			}
			text := err.Error()
			if v := refusalActsOffSurface(chair, foreign, chairWords, text); len(v) > 0 {
				t.Errorf("the convergence refusal, read by the chair alone, names %d act(s) off the chair's surface:\n  %s\n\nrefusal: %s\n\n"+
					"Name what THIS seat can do with a defect that still stops it. Whether the chair should hold the act is a ruling, not a refusal edit.",
					len(v), strings.Join(v, "\n  "), text)
			}
		})
	}
}

var (
	backticked = regexp.MustCompile("`([^`]+)`")
	proseWord  = regexp.MustCompile(`[a-z][a-z-]*`)
	// A sentence ends at a full stop or semicolon followed by space, or at the end of the text.
	sentenceEnd = regexp.MustCompile(`[.;]\s+`)
)

// refusalActsOffSurface is every act the text names that the chair cannot run.
func refusalActsOffSurface(chair *cobra.Command, foreign map[string][]string, chairWords map[string]bool, text string) []string {
	var out []string
	resolved := 0
	for _, m := range backticked.FindAllStringSubmatch(text, -1) {
		if why := resolveOnChair(chair, strings.Fields(m[1])); why != "" {
			out = append(out, "`"+m[1]+"`: "+why)
			continue
		}
		resolved++
	}
	if resolved < 2 {
		out = append(out, fmt.Sprintf("only %d backticked act(s) resolve on the chair's tree — the refusal names fewer than its PASS and the act a defect goes in, or the matcher stopped reading it", resolved))
	}
	instructions := 0
	for _, sentence := range sentenceEnd.Split(text, -1) {
		if !backticked.MatchString(sentence) {
			continue
		}
		instructions++
		words := proseWord.FindAllString(strings.ToLower(backticked.ReplaceAllString(sentence, " ")), -1)
		for path, roles := range foreign {
			act := strings.Fields(path)
			for i := 0; i+len(act) <= len(words); i++ {
				w := words[i+len(act)-1]
				if chairWords[w] || !inflects(w, act[len(act)-1]) || strings.Join(words[i:i+len(act)-1], " ") != strings.Join(act[:len(act)-1], " ") {
					continue
				}
				sort.Strings(roles)
				out = append(out, fmt.Sprintf("%q is the act `%s` (%s), which the chair does not hold — in: %s", strings.Join(words[i:i+len(act)], " "), path, strings.Join(roles, ", "), strings.TrimSpace(sentence)))
				break
			}
		}
	}
	if instructions == 0 {
		out = append(out, "no sentence names an act in backticks — the refusal tells the chair nothing to do, or the matcher stopped reading it")
	}
	sort.Strings(out)
	return out
}

// resolveOnChair says why a backticked span is not an act on the chair's tree, or "" when it is. A
// span opening with a flag is a flag of `verdict`, the verb the refusal answers.
func resolveOnChair(chair *cobra.Command, fields []string) string {
	if len(fields) == 0 {
		return "empty"
	}
	cmd := chair
	if strings.HasPrefix(fields[0], "--") {
		v, _, err := chair.Find([]string{"verdict"})
		if err != nil || v.Name() != "verdict" {
			return "the chair's tree has no verdict"
		}
		cmd = v
	} else {
		for len(fields) > 0 && !strings.HasPrefix(fields[0], "-") {
			next := sub(cmd, fields[0])
			if next == nil {
				break
			}
			cmd, fields = next, fields[1:]
		}
		if cmd == chair {
			return "no command on the chair's tree"
		}
	}
	for _, f := range fields {
		if name, ok := strings.CutPrefix(f, "--"); ok {
			if cmd.Flags().Lookup(name) == nil && cmd.InheritedFlags().Lookup(name) == nil {
				return "`" + cmd.CommandPath() + "` has no --" + name
			}
		}
	}
	return ""
}

func sub(c *cobra.Command, name string) *cobra.Command {
	for _, s := range c.Commands() {
		if s.Name() == name {
			return s
		}
	}
	return nil
}

func walk(c *cobra.Command, prefix []string, fn func(path []string, c *cobra.Command)) {
	for _, s := range c.Commands() {
		if s.Name() == "help" || s.Name() == "completion" {
			continue
		}
		path := append(append([]string{}, prefix...), s.Name())
		fn(path, s)
		walk(s, path, fn)
	}
}

// inflects reports whether word is verb or one of its regular inflections — mint, mints, minted,
// minting; close, closes, closed, closing; verify, verifies, verified.
func inflects(word, verb string) bool {
	if word == verb {
		return true
	}
	stems := []string{verb}
	if s, ok := strings.CutSuffix(verb, "e"); ok {
		stems = append(stems, s)
	}
	if s, ok := strings.CutSuffix(verb, "y"); ok {
		stems = append(stems, s+"i")
	}
	for _, s := range stems {
		for _, suf := range []string{"s", "es", "d", "ed", "ing"} {
			if word == s+suf {
				return true
			}
		}
	}
	return false
}

// board is a seeded record built one act at a time.
type board struct {
	t   *testing.T
	dir string
	evs []*recordpb.Event
}

func (b *board) add(seatID string, body proto.Message) *board {
	b.evs = append(b.evs, recordtest.At(b.t, seatID, fmt.Sprintf("%s:%d", seatID, len(b.evs)+1), body))
	return b
}

func (b *board) register(seatID string) *board {
	if record.SeatOwesOccasion(seatID) {
		return b.add(seatID, &recordpb.Register{Occasion: recordpb.Occasion_OCCASION_DOCKET.Enum()})
	}
	return b.add(seatID, &recordpb.Register{})
}

func (b *board) dispatch(pin int64, seatID string) *board {
	return b.add("red-chair", &recordpb.Dispatch{Pin: proto.Int64(pin), SeatId: proto.String(seatID)})
}

func (b *board) seed() record.Run {
	recordtest.Seed(b.t, b.dir, b.evs...)
	run, err := record.NewRun(b.dir)
	if err != nil {
		b.t.Fatal(err)
	}
	return run
}

// convergedBoard is a board the convergence rule holds on: the lens minted a high gap, red failed at
// that peak, the gap closed, and only a low gap is open — nothing material, mass below the fraction
// of the peak, nothing fresh and material this epoch — and two barren sittings retired the lens, so
// nothing another seat must clear holds the PASS.
func convergedBoard(t *testing.T) *board {
	const lens = "red-lens-evidence"
	dir := recordtest.TmpRun(t)
	mint := func(id string, g recordpb.Grade) *recordpb.Mint {
		return &recordpb.Mint{GapId: proto.String(id), Class: recordpb.GapClass_GAP_CLASS_SPEC_UNDERSPECIFICATION.Enum(), Problem: proto.String("p"),
			AcceptanceCheck: proto.String("c"), CheckKind: recordpb.CheckKind_CHECK_KIND_DOCUMENT.Enum(),
			Severity: g.Enum(), Likelihood: g.Enum(), Impact: g.Enum()}
	}
	b := &board{t: t, dir: dir}
	b.add("harness", &recordpb.Cast{SeatIds: []string{lens, "red-chair", "blue-respond", "judge"}}).
		add("harness", &recordpb.BaseIngest{Text: proto.String("# report")}).
		register("red-chair").register(lens).
		add(lens, mint("G1", recordpb.Grade_GRADE_HIGH)).
		add("red-chair", &recordpb.Gate{Verdict: recordpb.Verdict_VERDICT_FAIL.Enum()}).
		register("red-chair").register(lens).
		add(lens, &recordpb.Close{GapId: proto.String("G1"), ClosureClass: recordpb.Disposition_DISPOSITION_REPAIRED.Enum(),
			AnchorSeat: proto.String("L1"), AnchorTool: proto.String("t"), AnchorTarget: proto.String("x"), Prose: proto.String("fixed")}).
		add(lens, mint("G2", recordpb.Grade_GRADE_LOW)).
		register("red-chair").dispatch(2, lens).register(lens).dispatch(2, lens).register(lens).register("red-chair")
	return b
}
