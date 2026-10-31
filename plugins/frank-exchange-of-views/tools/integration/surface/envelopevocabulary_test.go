package surface

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/debatejs"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/repotree"
)

// WHAT AN ENVELOPE MAY SAY IS GENERATED, AND THESE TESTS HOLD WHAT GENERATION CANNOT.
//
// Every closed set a seat's envelope offers, and the schema of the plan the chair relays, are
// written into debate.js from the record's own declarations (devcmd/enginegen), and
// `enginegen -check` fails a region that is stale or edited by hand: an envelope word the record
// refuses, and a plan field the schema does not carry, cannot be committed. Two things are still
// authored, and each has its test here: the envelope shapes other than the plan, which nothing in
// Go declares; and the Go readers that index a returned envelope by key.

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
	// The plan is the chair's envelope and the read every seat's: a walk that reaches neither read no
	// schema, and every key would then be "undeclared" or "absent".
	if schemas == 0 || !keys["parties"] || !keys["verdict"] || !keys["bench_owed"] {
		t.Fatalf("the drive read %d schema(s) declaring %d key(s), without the chair's plan or the record's read — this is measuring nothing", schemas, len(keys))
	}
	return keys
}

// recordCarried is each fact the record carries that an envelope once restated, with where the
// record holds it.
var recordCarried = map[string]string{
	"petitions":               "a petition is a motion on the record, and the read on a close reports its hearing owed",
	"halt":                    "a halt is on the record, and the read on a close reports it",
	"unruled_motions":         "what stands unruled is the plan's blockers",
	"log":                     "a seat's log entry is a log event",
	"claim_count":             "the report's claim count is counted from the report",
	"sitting_record_appended": "what a sitting put on the record is read off the record",
	"rulings":                 "a petition ruling is a motion ruling on the record, with a grant's relief and whom it binds, and every seat's work list carries the relief",
	"relief":                  "a grant's relief is on its ruling, and every seat's work list carries it",
	"dispositions":            "a docket ruling is a motion ruling on the record, and every seat's work list carries what a closing ruling settled and what would reopen it",
	"settled":                 "a docket ruling's settled proposition is on the ruling, and every seat's work list carries it",
	"infra_debts":             "a defect_owed_elsewhere ruling is on the record, and the judgments render it",
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
	if files == 0 || len(read) == 0 || len(read["parties"]) == 0 {
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
