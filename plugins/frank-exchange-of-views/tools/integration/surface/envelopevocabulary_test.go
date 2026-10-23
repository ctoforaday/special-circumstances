package surface

import (
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/debatejs"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/repotree"
)

// WHAT AN ENVELOPE MAY SAY IS GENERATED, AND THESE TESTS HOLD WHAT GENERATION CANNOT.
//
// Every closed set a seat's envelope offers, and the schema of the plan the chair relays, are
// written into debate.js from the record's own declarations (devcmd/enginegen), and
// `enginegen -check` fails a region that is stale or edited by hand: an envelope word the record
// refuses, and a plan field the schema does not carry, cannot be committed. Three things are
// still authored, and each has its test here: the DUTY the engine hands blue for each fate the
// bench can rule; the envelope shapes other than the plan, which nothing in Go declares; and the
// Go readers that index a returned envelope by key.

// planJSON is a plan as `dispatch next` prints it: record.Plan through its own JSON tags, with
// every list the verb always prints an array and never null, so a schedule relays the shape the
// verb does and holds no key list of its own.
func planJSON(t *testing.T, p record.Plan) map[string]any {
	t.Helper()
	var arrays func(v reflect.Value)
	arrays = func(v reflect.Value) {
		for i := range v.NumField() {
			f := v.Field(i)
			if f.Kind() != reflect.Slice {
				continue
			}
			if f.IsNil() && !strings.Contains(v.Type().Field(i).Tag.Get("json"), "omitempty") {
				f.Set(reflect.MakeSlice(f.Type(), 0, 0))
			}
			for j := range f.Len() {
				if f.Index(j).Kind() == reflect.Struct {
					arrays(f.Index(j))
				}
			}
		}
	}
	arrays(reflect.ValueOf(&p).Elem())
	b, err := json.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	var out map[string]any
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatal(err)
	}
	return out
}

const rulingsMarker = "GAPS THE BENCH HAS RULED, AND WHAT EACH RULING OBLIGES OF YOU: "

// EVERY FATE THE BENCH CAN RULE REACHES BLUE WITH A DUTY OF ITS OWN.
//
// Agreeing on the WORDS is not agreeing on the CONSEQUENCE. The disposition vocabulary is
// generated, so the envelope offers exactly the record's fates; what each obliges of blue is
// BLUE_DUTY_BY_DISPOSITION, authored text keyed by fate, and a fate with no entry falls through to
// `UNMAPPED DISPOSITION` — a seat told to go and find its own instruction, on a run nobody is
// watching. Adding `moot` to the record forced it onto the envelope and required nothing of the
// duty map (#847).
//
// THE SCRIPT IS RUN: the bench rules one gap under each fate the record's docket ruling accepts —
// and `remanded` twice, once with the exchange the plan still owes and once without — and blue's
// next prompt is read for each row's duty.
func TestEveryRulingFateHandsBlueADuty(t *testing.T) {
	fates := record.Names(record.SeatFilable(record.MotionVerdicts["docket"]))
	if len(fates) == 0 {
		t.Fatal("the record's docket ruling accepts no fate — nothing would be ruled, which reads as every fate carrying a duty")
	}
	const remanded, direction = "remanded", "READ THE PRIMARY SOURCE AT THE LEAF"
	gapOf := map[string]string{}
	var gaps []string
	var rulings []any
	rule := func(gap, fate string) {
		gaps = append(gaps, gap)
		gapOf[gap] = fate
		rulings = append(rulings, map[string]any{"gap_id": gap, "disposition": fate, "rationale": "r", "settled": "the proposition " + gap + " settles"})
	}
	for i, f := range fates {
		rule("G-"+strconv.Itoa(i+1), f)
	}
	const owed, notOwed = "G-1000", "G-1001"
	rule(owed, remanded)
	rule(notOwed, remanded)
	if gapOf["G-"+strconv.Itoa(len(fates))] == "" || !containsWord(fates, remanded) {
		t.Fatalf("the record's docket fates %v hold no %q — the remand arms of this schedule are measuring nothing", fates, remanded)
	}

	chair := 0
	backend := func(seatID, label, prompt string) debatejs.Envelope {
		e := debatejs.Envelope{"rulings": []any{}, "dispositions": []any{},
			"manifest": []any{}, "found_closed": []any{}, "saturation_reached": false, "open_gaps": 0}
		switch {
		case seatID == "red-chair":
			chair++
			switch chair {
			case 1:
				e["plan"] = planJSON(t, record.Plan{Head: 2, Docket: gaps,
					Parties:  []record.Party{{SeatID: "judge", GapIDs: gaps, Occasions: []string{"docket"}}},
					Blockers: []record.PlanBlocker{{Kind: record.BlockerMaterialGap, Subject: owed, Owner: "red-lens-logic"}}})
			case 2:
				e["plan"] = planJSON(t, record.Plan{Head: 2,
					Parties:    []record.Party{{SeatID: "blue-respond", GapIDs: gaps}},
					RemandOwed: []record.RemandOwed{{GapID: owed, Direction: direction}},
					Blockers:   []record.PlanBlocker{{Kind: record.BlockerMaterialGap, Subject: owed, Owner: "red-lens-logic"}}})
			default:
				e["plan"] = planJSON(t, record.Plan{Head: 2, PassPermitted: true})
				e["verdict"] = "PASS"
			}
		case seatID == "judge" && debatejs.OccasionOf(prompt) == "docket":
			e["dispositions"] = rulings
		}
		return e
	}
	script, err := repotree.DebateJS()
	if err != nil {
		t.Fatal(err)
	}
	ds, _, err := debatejs.CaptureRun(script, debatejs.Config{Topic: "disposition schedule", Lanes: 3, Model: "m-bulk", JudgmentModel: "m-judgment",
		RunDir: t.TempDir(), BinDir: t.TempDir(), Backend: backend})
	if err != nil {
		t.Fatal(err)
	}
	blue, err := debatejs.Last(ds, "blue-respond")
	if err != nil {
		t.Fatal(err)
	}
	at := strings.Index(blue.Prompt, rulingsMarker)
	if at < 0 {
		t.Fatalf("blue's prompt after the bench ruled carries no rulings (%q) — the schedule reached no ruled gap, which reads as every fate carrying a duty", rulingsMarker)
	}
	var rows []map[string]any
	if err := json.NewDecoder(strings.NewReader(blue.Prompt[at+len(rulingsMarker):])).Decode(&rows); err != nil {
		t.Fatalf("the rulings blue is handed are not a list of rows: %v", err)
	}
	if len(rows) != len(gaps) {
		t.Fatalf("the bench ruled %d gap(s) and blue is handed %d row(s)", len(gaps), len(rows))
	}

	dutyOf := map[string]string{} // fate -> the duty its row carries
	byGap := map[string]map[string]any{}
	for _, r := range rows {
		gap, _ := r["gap_id"].(string)
		fate, _ := r["disposition"].(string)
		duty, _ := r["your_duty"].(string)
		byGap[gap] = r
		if fate != gapOf[gap] {
			t.Errorf("%s was ruled %q and blue's row says %q", gap, gapOf[gap], fate)
		}
		if strings.TrimSpace(duty) == "" || strings.Contains(duty, "UNMAPPED DISPOSITION") {
			t.Errorf("the bench can rule %q, and the engine hands blue no duty for it (%q) — BLUE_DUTY_BY_DISPOSITION has no entry for the fate", fate, duty)
		}
		if gap != notOwed {
			dutyOf[fate] = duty
		}
	}
	// A FATE'S DUTY IS ITS OWN: a new fate keyed onto another's sentence tells blue the wrong thing
	// in the right shape.
	seen := map[string]string{}
	for _, f := range fates {
		if other, dup := seen[dutyOf[f]]; dup {
			t.Errorf("%q and %q hand blue the same duty: %q", other, f, dutyOf[f])
		}
		seen[dutyOf[f]] = f
	}
	// A REMAND'S DUTY LASTS AS LONG AS ITS EXCHANGE IS OWED: the plan's direction on the gap it
	// readies, and neither the direction nor the remand's duty on the one it does not.
	if got, _ := byGap[owed]["direction"].(string); got != direction {
		t.Errorf("the plan owes %s's remand exchange with direction %q, and blue's row carries %q", owed, direction, got)
	}
	if _, has := byGap[notOwed]["direction"]; has {
		t.Errorf("the plan owes no exchange on %s, and blue's row carries a direction: %v", notOwed, byGap[notOwed]["direction"])
	}
	if byGap[owed]["your_duty"] == byGap[notOwed]["your_duty"] {
		t.Errorf("a remand whose exchange is owed and one whose exchange is not hand blue the same duty: %v", byGap[owed]["your_duty"])
	}
}

func containsWord(ws []string, w string) bool {
	for _, x := range ws {
		if x == w {
			return true
		}
	}
	return false
}

// everyEnvelopeKey is every property name any envelope schema of the script declares, read off
// the dispatches of a drive that reaches every agent() call.
func everyEnvelopeKey(t *testing.T) map[string]bool {
	t.Helper()
	agentPaths, err := repotree.Constitutions()
	if err != nil {
		t.Fatal(err)
	}
	var agentTypes []string
	for _, p := range agentPaths {
		agentTypes = append(agentTypes, strings.TrimSuffix(filepath.Base(p), ".md"))
	}
	ds, _ := driveEverySite(t, agentTypes, debatejs.Config{Topic: "envelope keys", Lanes: 3, Model: "m-bulk", JudgmentModel: "m-judgment"})
	keys := map[string]bool{}
	schemas := 0
	for _, d := range ds {
		if d.Schema == nil {
			continue
		}
		schemas++
		for _, k := range debatejs.SchemaKeys(d.Schema) {
			keys[k] = true
		}
	}
	// The plan is relayed under the chair's schema and a disposition under the bench's: a walk
	// that reaches neither read no schema, and every key would then be "undeclared" or "absent".
	if schemas == 0 || !keys["plan"] || !keys["parties"] || !keys["disposition"] {
		t.Fatalf("the drive read %d schema(s) declaring %d key(s), without the chair's plan or the bench's disposition — this is measuring nothing", schemas, len(keys))
	}
	return keys
}

// recordCarried is each fact the record carries that an envelope once restated, with where the
// record holds it.
var recordCarried = map[string]string{
	"petitions":               "a petition is a motion on the record, and the chair's plan convenes the bench for it",
	"unruled_motions":         "what stands unruled is the plan's blockers",
	"log":                     "a seat's log entry is a log event",
	"claim_count":             "the report's claim count is counted from the report",
	"sitting_record_appended": "what a sitting put on the record is read off the record",
}

// NO ENVELOPE RESTATES A FACT THE RECORD CARRIES. The petition channel was an envelope field
// that lenses — whose returns are prose — could never fill, so their petitions reached no bench;
// the terminal sitting fired on the chair's own count of unruled motions, which nothing audited;
// the envelope's log was a copy of entries every seat writes to the record; and blue typed a
// claim count and an attestation beside the tool's own (#1203, #1267, #1298).
//
// GENERATION DOES NOT REACH THIS: the plan's schema is written from record.Plan, and the blue,
// bench, petition and assembly envelopes are authored in the script with nothing in Go declaring
// them. So the schemas are read as dispatched, and none may declare one of these keys. What the
// engine does with a key a seat returns anyway is the simulator's to hold.
func TestNoEnvelopeRestatesARecordFact(t *testing.T) {
	keys := everyEnvelopeKey(t)
	for k, where := range recordCarried {
		if keys[k] {
			t.Errorf("an envelope schema declares %q — %s; a seat's envelope restating it is a second channel for one fact", k, where)
		}
	}
}

// journalFraming is each key the journal's and a transcript's readers index that is no envelope
// key, with what it is. An entry no reader indexes, or one an envelope schema declares, fails:
// the list cannot outlive a read or shadow a declared key.
var journalFraming = map[string]string{
	"result":           "the workflow journal row's field holding what a seat returned — the envelope itself",
	"agentId":          "the workflow journal row's agent",
	"message":          "a transcript row's message",
	"content":          "a transcript message's content blocks",
	"type":             "a transcript content block's kind",
	"input":            "a tool-use block's input",
	"id":               "a tool-use block's id, which its result names",
	"tool_use_id":      "a tool-result block's call: the tool-use block's id",
	"timestamp":        "a transcript row's stamp, which places a seat's close among its calls",
	"assemble":         "the dashboard's seat-class counter, keyed by the class word",
	"Integrity audits": "capture's fold of the run document, keyed by section heading",
}

// envelopeReaders are the packages that read what seats returned out of the workflow journal.
var envelopeReaders = []string{"capture", "dashboard", "scorecard"}

// EVERY ENVELOPE KEY A GO READER INDEXES IS ONE A SCHEMA DECLARES.
//
// Capture, the dashboard and the scorecard read seats' returns out of the workflow journal as
// decoded JSON, by key. A key no schema declares is a key no seat is asked for: the read finds
// nothing, and nothing reads exactly like a seat that returned none — an envelope log read for
// entries no seat could return, a scorecard row scored off an attestation no schema carried. The
// class recurred three times, each as a reader left behind when a schema key was removed.
//
// So every string a reader indexes a map by, in those packages, is a key some dispatched schema
// declares or a declared piece of journal or transcript framing. A removed schema key still read
// fails here, naming the file.
func TestEveryEnvelopeKeyAGoReaderIndexesIsDeclared(t *testing.T) {
	keys := everyEnvelopeKey(t)
	srcs, err := repotree.ToolSources()
	if err != nil {
		t.Fatal(err)
	}
	read := map[string][]string{} // key -> the files that index by it
	files := 0
	fset := token.NewFileSet()
	for _, p := range srcs {
		if strings.HasSuffix(p, "_test.go") || !containsWord(envelopeReaders, filepath.Base(filepath.Dir(p))) ||
			filepath.Base(filepath.Dir(filepath.Dir(p))) != "internal" {
			continue
		}
		f, err := parser.ParseFile(fset, p, nil, parser.SkipObjectResolution)
		if err != nil {
			t.Fatal(err)
		}
		files++
		ast.Inspect(f, func(n ast.Node) bool {
			ix, ok := n.(*ast.IndexExpr)
			if !ok {
				return true
			}
			lit, ok := ix.Index.(*ast.BasicLit)
			if !ok || lit.Kind != token.STRING {
				return true
			}
			k, err := strconv.Unquote(lit.Value)
			if err != nil {
				t.Fatal(err)
			}
			where := filepath.Join(filepath.Base(filepath.Dir(p)), filepath.Base(p)) + ":" + strconv.Itoa(fset.Position(ix.Pos()).Line)
			read[k] = append(read[k], where)
			return true
		})
	}
	if files == 0 || len(read) == 0 || len(read["plan"]) == 0 {
		t.Fatalf("parsed %d file(s) of %v and found %d indexed key(s), without capture's read of the relayed plan — this is measuring nothing", files, envelopeReaders, len(read))
	}
	var undeclared []string
	for k, at := range read {
		if !keys[k] && journalFraming[k] == "" {
			sort.Strings(at)
			undeclared = append(undeclared, k+" ("+strings.Join(at, ", ")+")")
		}
	}
	sort.Strings(undeclared)
	for _, u := range undeclared {
		t.Errorf("a Go reader indexes %s, which no envelope schema declares and journalFraming does not name — "+
			"no seat is asked for it, so the read finds nothing and reads as a seat that returned none", u)
	}
	for k := range journalFraming {
		if len(read[k]) == 0 {
			t.Errorf("journalFraming names %q, which no reader in %v indexes — remove it", k, envelopeReaders)
		}
		if keys[k] {
			t.Errorf("journalFraming names %q, which an envelope schema declares — it is an envelope key, and the entry shadows it", k)
		}
	}
}
