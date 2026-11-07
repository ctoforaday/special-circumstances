package sittingwrite

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

func entry(key, digest string) record.QueueEntry {
	return record.QueueEntry{Key: key, Kind: record.ItemReaudit, Digest: digest, Items: []record.Item{}}
}

// WHAT A LENS IS HANDED AFTER A CALL TURNS ON WHAT IT WAS HANDED BEFORE, and on nothing else. The
// current item stays current until it leaves the queue, whatever arrives ahead of it; a moved item is
// handed again and says so; another key leaving is one line; nothing changed is nothing.
func TestTheNextItemTurnsOnWhatWasHanded(t *testing.T) {
	w := record.WorkJSON{}
	q := []record.QueueEntry{entry("a", "1"), entry("b", "1"), entry("c", "1")}
	cases := []struct {
		name     string
		prev     delivered
		had      bool
		q        []record.QueueEntry
		want     string // a substring of what is handed; "" for nothing
		wantKey  string
		notQueue bool
	}{
		{name: "nothing changed", prev: delivered{Key: "b", Digest: "1", Queue: []string{"a", "b", "c"}}, had: true, q: q, want: "", wantKey: "b"},
		{name: "an item arrived ahead: no preemption", prev: delivered{Key: "b", Digest: "1", Queue: []string{"b", "c"}}, had: true, q: q, want: "", wantKey: "b"},
		{name: "another left", prev: delivered{Key: "b", Digest: "1", Queue: []string{"a", "b", "c", "d"}}, had: true, q: q, want: `["d"] left your list; your current item is still b`, wantKey: "b"},
		{name: "current moved", prev: delivered{Key: "b", Digest: "0", Queue: []string{"a", "b", "c"}}, had: true, q: q, want: `"changed_since_delivered": true`, wantKey: "b"},
		{name: "current discharged", prev: delivered{Key: "z", Digest: "1", Queue: []string{"z", "a"}}, had: true, q: q, want: nextArrived, wantKey: "a"},
		{name: "file lost", q: q, want: nextArrived, wantKey: "a"},
		{name: "last one discharged", prev: delivered{Key: "z", Digest: "1", Queue: []string{"z"}}, had: true, want: "z left your list, and nothing more is queued for you", wantKey: ""},
		{name: "nothing was queued, nothing is", prev: delivered{Queue: []string{}}, had: true, want: "", wantKey: ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			text, now := nextOf(w, c.q, c.prev, c.had)
			if c.want == "" && text != "" {
				t.Errorf("handed %q, want nothing", text)
			}
			if c.want != "" && !strings.Contains(text, c.want) {
				t.Errorf("handed %q, want it to carry %q", text, c.want)
			}
			if now.Key != c.wantKey {
				t.Errorf("current on file = %q, want %q", now.Key, c.wantKey)
			}
		})
	}
}

// THE CURRENT GAP CARRIES ITS PROBLEM ONCE, AND A PASSAGE OVER BUDGET IS CUT TO ITS ANCHOR'S SENTENCE
// AND SAYS SO.
func TestTheDeliveryFitsAndSaysWhatItCut(t *testing.T) {
	long := strings.Repeat("Filler sentence about nothing in particular. ", 300)
	loc := "the sentence the gap is about"
	g := record.WorkGapJSON{ID: "G-1", Problem: "the problem", ProblemSynopsis: "the problem", YourTurn: true,
		Location: loc, Passage: long + loc + ". " + long, MintedBy: "red-lens-logic", Backing: []record.GapBackingJSON{}, RegradesOwed: []record.OwedRegradeJSON{}}
	w := record.WorkJSON{Open: []record.WorkGapJSON{g}, Sitting: record.SittingJSON{Seat: "red-lens-logic", Role: "lens", Open: []record.Item{}}}
	text := RenderLensDelivery(w)
	if len(text) > DeliveryBudget {
		t.Fatalf("delivery is %d bytes, over the budget of %d", len(text), DeliveryBudget)
	}
	body := text[strings.Index(text, "{"):]
	var d struct {
		Current struct {
			Gap map[string]any `json:"gap"`
		} `json:"current"`
	}
	if err := json.Unmarshal([]byte(body), &d); err != nil {
		t.Fatal(err)
	}
	if _, ok := d.Current.Gap["problem_synopsis"]; ok {
		t.Error("the current gap carries its problem twice")
	}
	if d.Current.Gap["passage_clipped"] != true || !strings.Contains(d.Current.Gap["passage"].(string), loc) {
		t.Errorf("the clipped passage is %v, clipped %v — want the anchor's sentence kept and the cut stated", d.Current.Gap["passage"], d.Current.Gap["passage_clipped"])
	}
}

// THE OPENING END STORES WHAT IT HANDED, AND THE FIRST CALL AFTER IT COMPARES AGAINST THAT: nothing
// moved, so nothing is said. A lost file hands the head of the queue again.
func TestAfterTheOpeningEndNothingMovedIsNothingSaid(t *testing.T) {
	dir := newRun(t)
	recordtest.Seed(t, dir,
		recordtest.At(t, record.HarnessSeat, "harness:cast", &recordpb.Cast{
			SeatIds: []string{"red-chair", "red-lens-evidence", "blue-respond", "judge"}}),
		recordtest.At(t, record.HarnessSeat, "harness:ingest",
			&recordpb.ReportPart{Text: proto.String("# report\n\nOne paragraph.\n")}),
		recordtest.At(t, "red-chair", "chair:dispatch", &recordpb.Dispatch{
			SeatId: proto.String("red-lens-evidence"), Pin: proto.Int64(1)}))
	const agentType = "frank-exchange-of-views:red-lens-evidence"
	if err := Write(Sitting{RunDir: dir, Phase: Open, AgentID: "agent_01", AgentType: agentType}, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := WriteNextItem(dir, "agent_01", agentType, &out); err != nil {
		t.Fatal(err)
	}
	if out.Len() != 0 {
		t.Errorf("nothing moved and the lens was handed:\n%s", out.String())
	}
	files, _ := filepath.Glob(filepath.Join(dir, deliveryDir, "agent_01-*.json"))
	if len(files) != 1 {
		t.Fatalf("delivery files = %v, want the one the opening end stored", files)
	}
	if err := os.Remove(files[0]); err != nil {
		t.Fatal(err)
	}
	if err := WriteNextItem(dir, "agent_01", agentType, &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), `"key": "audit:report:`) {
		t.Errorf("a lost file did not hand the head of the queue again:\n%s", out.String())
	}
	// ANOTHER SEAT'S CALL IS HANDED NOTHING: its work arrives whole.
	out.Reset()
	if err := WriteNextItem(dir, "agent_02", "frank-exchange-of-views:red-chair", &out); err != nil || out.Len() != 0 {
		t.Errorf("a chair's call was handed %q (%v)", out.String(), err)
	}
}

// EVERY DELIVERY FITS THE BUDGET, IN BYTES, WHATEVER THE BOARD HOLDS — and the current item arrives
// whole but for its passage. An adversarial board: sixty other open gaps, forty readied gaps of the
// lens's own, thirty estopped rulings, twenty long declarations and petitions, twenty notes, all in
// multi-byte text, and a current gap with a long passage. What is cut is counted, never silent.
func TestEveryDeliveryFitsTheBudgetByConstruction(t *testing.T) {
	long := strings.Repeat("Ünïcödé prose — every rune here is more than one byte. ", 40)
	row := func(id, by string, turn bool) record.WorkGapJSON {
		return record.WorkGapJSON{ID: id, MintedBy: by, YourTurn: turn, Class: "false_universal", Material: turn,
			Location: "the sentence " + id, Passage: long + "the sentence " + id + ". " + long, Problem: "problem of " + id,
			RequiredFix: "fix " + id, AcceptanceCheck: "check " + id, Backing: backingOf(60), EditedSince: editsOf(10), RegradesOwed: []record.OwedRegradeJSON{}}
	}
	w := record.WorkJSON{Sitting: record.SittingJSON{Seat: "red-lens-voice", Role: "lens", Open: []record.Item{}}}
	for i := 0; i < 40; i++ {
		id := fmt.Sprintf("G-own-%02d", i)
		w.Open = append(w.Open, row(id, "red-lens-voice", true))
		w.Sitting.Open = append(w.Sitting.Open, record.Item{Kind: record.ItemReaudit, Subject: &record.ItemSubject{Kind: record.SubjectGap, ID: id}, What: "gap " + id + " was answered — " + long[:300]})
	}
	for i := 0; i < 60; i++ {
		w.Open = append(w.Open, row(fmt.Sprintf("G-other-%02d", i), "red-lens-logic", false))
	}
	for i := 0; i < 20; i++ {
		w.Sitting.Open = append(w.Sitting.Open, record.Item{Kind: record.ItemNews, Subject: &record.ItemSubject{Kind: record.SubjectGap, ID: fmt.Sprintf("G-n%d", i)}, What: long})
		w.Declarations = append(w.Declarations, record.DeclarationJSON{Holding: long, Epoch: i, Gaps: []record.DeclaredGapJSON{{ID: "G-1", Status: "open"}}})
		w.Relief = append(w.Relief, record.ReliefJSON{Motion: fmt.Sprintf("M-%d", i), Petitioner: "blue-respond", Binds: "red", Relief: long, Epoch: i})
	}
	for i := 0; i < 30; i++ {
		w.Estopped = append(w.Estopped, record.EstoppedJSON{ID: fmt.Sprintf("G-e%d", i), Class: "false_universal", Fate: "not_a_defect", Settled: long})
	}
	q := record.QueueOf(w)
	check := func(name, text string, at int) {
		t.Helper()
		if len(text) > DeliveryBudget {
			t.Fatalf("%s: %d bytes, over the budget of %d", name, len(text), DeliveryBudget)
		}
		var d lensDelivery
		if err := json.Unmarshal([]byte(text[strings.Index(text, "{"):]), &d); err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		e := q[at]
		if d.Current == nil || d.Current.Key != e.Key || len(d.Current.Items) != len(e.Items) || d.Current.Items[0].What != e.Items[0].What ||
			d.Current.Gap.Problem != e.Gap.Problem || d.Current.Gap.RequiredFix != e.Gap.RequiredFix || !d.Current.Gap.PassageClipped {
			t.Errorf("%s: the current item did not arrive whole but for its passage: %+v", name, d.Current)
		}
		// The backing the CI run carried past the budget (#1772): its anchors are all there, as ids.
		if ids, _ := d.Current.Gap.Backing.([]any); !d.Current.Gap.BackingClipped || len(ids) != len(e.Gap.Backing) {
			t.Errorf("%s: backing arrived as %v (clipped %v), want all %d anchors as ids", name, d.Current.Gap.Backing, d.Current.Gap.BackingClipped, len(e.Gap.Backing))
		}
		// Every other queued entry and open gap is a line or a count: nothing is dropped silently.
		if got, want := len(d.Index)+d.IndexOmitted, len(q)-1+60; got != want {
			t.Errorf("%s: index holds %d lines and %d omitted, want %d in all", name, len(d.Index), d.IndexOmitted, want)
		}
		// The estopped lines ride the delivery at dispatch; a next item carries the item and the index.
		if got := len(d.Estopped) + d.EstoppedOmitted; d.Sitting != nil && got != 30 {
			t.Errorf("%s: estopped holds %d lines and %d omitted, want 30 in all", name, len(d.Estopped), d.EstoppedOmitted)
		}
	}
	check("at dispatch", RenderLensDelivery(w), 0)
	check("next", renderNext(w, q, 3, false), 3)
	check("changed", renderNext(w, q, 0, true), 0)
}

func backingOf(n int) []record.GapBackingJSON {
	out := []record.GapBackingJSON{}
	for i := 0; i < n; i++ {
		out = append(out, record.GapBackingJSON{Anchor: fmt.Sprintf("C-%04d", i), Kind: "citation", Outcome: "supports_with_qualification",
			Confidence: "moderate", VerifiedBy: "red-lens-evidence", Event: int64(1000 + i)})
	}
	return out
}

func editsOf(n int) []record.GapEdit {
	out := []record.GapEdit{}
	for i := 0; i < n; i++ {
		out = append(out, record.GapEdit{Epoch: i, EditedBy: "blue-respond", Old: strings.Repeat("old sentence — ", 20), New: strings.Repeat("new sentence — ", 20)})
	}
	return out
}

// THE CURRENT ITEM ITSELF PAST THE BUDGET: a gap whose sentence carries hundreds of verified anchors,
// edited dozens of times, and whose problem, fix, check and fix pair each run past the budget alone.
// It still fits, and every part cut short is marked by a field: nothing the seat reads as whole is not.
func TestAGapPastTheBudgetIsCutAndSaysSo(t *testing.T) {
	long := strings.Repeat("Ünïcödé prose — every rune here is more than one byte. ", 200)
	for _, tc := range []struct {
		name    string
		backing int
		edits   int
		prose   string
	}{
		{"huge backing, short prose", 400, 0, "short"},
		{"huge edits, short prose", 0, 300, "short"},
		{"long prose, huge lists", 1000, 500, long},
	} {
		g := record.WorkGapJSON{ID: "G-1", MintedBy: "red-lens-architecture", YourTurn: true, Class: "false_universal", Material: true,
			Location: "the sentence", Passage: long + "the sentence. " + long, Problem: "problem " + tc.prose, RequiredFix: "fix " + tc.prose,
			AcceptanceCheck: "check " + tc.prose, FixBasis: "quoted", FixOld: "old " + tc.prose, FixNew: "new " + tc.prose,
			Backing: backingOf(tc.backing), EditedSince: editsOf(tc.edits), RegradesOwed: []record.OwedRegradeJSON{}, FoundBy: []string{"F-1"}}
		w := record.WorkJSON{Sitting: record.SittingJSON{Seat: "red-lens-architecture", Role: "lens",
			Open: []record.Item{{Kind: record.ItemReaudit, Subject: &record.ItemSubject{Kind: record.SubjectGap, ID: "G-1"}, What: "gap G-1 was answered"}}},
			Open: []record.WorkGapJSON{g}}
		q := record.QueueOf(w)
		for _, r := range []struct {
			arm  string
			text string
		}{{"at dispatch", RenderLensDelivery(w)}, {"next", renderNext(w, q, 0, false)}} {
			name := tc.name + ", " + r.arm
			if len(r.text) > DeliveryBudget {
				t.Fatalf("%s: %d bytes, over the budget of %d", name, len(r.text), DeliveryBudget)
			}
			var d lensDelivery
			if err := json.Unmarshal([]byte(r.text[strings.Index(r.text, "{"):]), &d); err != nil {
				t.Fatalf("%s: %v", name, err)
			}
			dg := d.Current.Gap
			if dg == nil || dg.ID != g.ID || dg.Location != g.Location {
				t.Fatalf("%s: the current gap lost its id or location: %+v", name, dg)
			}
			for _, f := range []struct {
				field, got, whole string
				clipped           bool
			}{
				{"passage", dg.Passage, g.Passage, dg.PassageClipped},
				{"problem", dg.Problem, g.Problem, dg.ProblemClipped},
				{"required_fix", dg.RequiredFix, g.RequiredFix, dg.RequiredFixClipped},
				{"acceptance_check", dg.AcceptanceCheck, g.AcceptanceCheck, dg.AcceptanceCheckClipped},
				{"fix_old", dg.FixOld, g.FixOld, dg.FixOldClipped},
				{"fix_new", dg.FixNew, g.FixNew, dg.FixNewClipped},
			} {
				if (f.got != f.whole) != f.clipped {
					t.Errorf("%s: %s is cut %v and marked clipped %v", name, f.field, f.got != f.whole, f.clipped)
				}
			}
			if tc.prose == "short" && (dg.ProblemClipped || dg.RequiredFixClipped || dg.AcceptanceCheckClipped) {
				t.Errorf("%s: short prose was clipped though the lists could give way first", name)
			}
			ids, _ := dg.Backing.([]any)
			if want := len(g.Backing); want > 0 && (!dg.BackingClipped || len(ids)+dg.BackingOmitted != want) {
				t.Errorf("%s: backing holds %d ids and %d omitted (clipped %v), want %d in all", name, len(ids), dg.BackingOmitted, dg.BackingClipped, want)
			}
			marks, _ := dg.EditedSince.([]any)
			if want := len(g.EditedSince); want > 0 && (!dg.EditedSinceClipped || len(marks)+dg.EditedSinceOmitted != want) {
				t.Errorf("%s: edited_since holds %d and %d omitted (clipped %v), want %d in all", name, len(marks), dg.EditedSinceOmitted, dg.EditedSinceClipped, want)
			}
			t.Logf("%s: %d bytes; backing %d ids + %d omitted; edits %d + %d omitted", name, len(r.text), len(ids), dg.BackingOmitted, len(marks), dg.EditedSinceOmitted)
		}
	}
}
