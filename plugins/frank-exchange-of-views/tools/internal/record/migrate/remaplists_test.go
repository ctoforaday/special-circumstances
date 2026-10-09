package migrate

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// stringFields visits every string field of every message reachable from an event body — nested
// messages and oneof arms included — once per message type, by its "Message.field" name.
func stringFields(visit func(name string, fd protoreflect.FieldDescriptor)) {
	seen := map[protoreflect.FullName]bool{}
	var walk func(md protoreflect.MessageDescriptor)
	walk = func(md protoreflect.MessageDescriptor) {
		if seen[md.FullName()] {
			return
		}
		seen[md.FullName()] = true
		for i := 0; i < md.Fields().Len(); i++ {
			fd := md.Fields().Get(i)
			switch fd.Kind() {
			case protoreflect.MessageKind:
				walk(fd.Message())
			case protoreflect.StringKind:
				visit(string(md.Name())+"."+string(fd.Name()), fd)
			}
		}
	}
	bodies := (&recordpb.Event{}).ProtoReflect().Descriptor().Oneofs().ByName("body").Fields()
	for i := 0; i < bodies.Len(); i++ {
		walk(bodies.Get(i).Message())
	}
}

// EVERY STRING FIELD AN EVENT BODY REACHES IS IN EXACTLY ONE OF THE THREE LISTS (R-19), so a field
// added to record.proto is classed before it migrates, and a list entry naming no field is a
// stale one. The schema's own `(prose)` mark agrees with the seat-argument list everywhere but
// Cite.title, which is a source's title: the subject's words, not the seat's (R-8, A-4).
func TestEveryStringFieldIsInOneList(t *testing.T) {
	lists := map[string]map[string]bool{"id": idFields, "seat-argument": argumentFields, "source-data": dataFields}
	reached := map[string]bool{}
	motionArms := 0
	stringFields(func(name string, fd protoreflect.FieldDescriptor) {
		reached[name] = true
		var in []string
		for class, set := range lists {
			if set[name] {
				in = append(in, class)
			}
		}
		sort.Strings(in)
		if len(in) != 1 {
			t.Errorf("%s is in %d of the three field lists %v — class it in exactly one of migrate/remap.go's lists before it migrates", name, len(in), in)
		}
		if prose, _ := recordpb.IsProse(fd); prose && !argumentFields[name] && name != "Cite.title" {
			t.Errorf("%s is marked (prose) in record.proto and is not in the seat-argument list", name)
		}
		switch strings.SplitN(name, ".", 2)[0] {
		case "GradeMotion", "DocketMotion", "AvenueMotion", "DocketRuling":
			motionArms++
		}
	})
	for class, set := range lists {
		for name := range set {
			if !reached[name] {
				t.Errorf("the %s list names %s, which no event body reaches", class, name)
			}
		}
	}
	for field := range mints {
		if !idFields[field] {
			t.Errorf("mints names %s, which is not in the id list", field)
		}
	}
	if motionArms != 8 {
		t.Errorf("walked %d string fields of Motion's nested filings and ruling, want 8 — the walk stopped at the top-level bodies", motionArms)
	}
	if got := fmt.Sprint(len(idFields), len(argumentFields), len(dataFields)); got != "34 50 68" {
		t.Errorf("the lists hold %s fields (id, seat-argument, source-data), want 34 50 68", got)
	}
}

// archivedRuns are the run tarballs in run-archive/, by file name.
func archivedRuns(t *testing.T) []string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for ; filepath.Dir(dir) != dir; dir = filepath.Dir(dir) {
		if tarballs, _ := filepath.Glob(filepath.Join(dir, "run-archive", "*.tar.gz")); len(tarballs) > 0 {
			for i := range tarballs {
				tarballs[i] = filepath.Base(tarballs[i])
			}
			return tarballs
		}
	}
	t.Skip("run-archive/ not found above this checkout")
	return nil
}

// archivedSource opens an extracted archived run the way Migrate does.
func archivedSource(t *testing.T, runDir string) Source {
	t.Helper()
	records := filepath.Join(runDir, "records")
	if _, err := os.Stat(filepath.Join(records, dbSiblings[0])); err != nil {
		js, err := OpenJSONL(records)
		if err != nil {
			t.Fatal(err)
		}
		return js
	}
	sq, err := OpenSQLite(records, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { sq.Close() })
	return sq
}

// anyToken is an anchor token in either spelling, the test's own reader; tokenless spells each as
// its tag alone, so two texts that differ only in the ids their tokens carry compare equal.
var anyToken = regexp.MustCompile(`<!--(fx|cite|proof|gap):[A-Za-z0-9-]+-->`)

func tokenless(s string) string { return anyToken.ReplaceAllString(s, "<!--$1-->") }

// eachString visits every set string value of a message, nested messages included, in schema order.
func eachString(m protoreflect.Message, visit func(name, v string)) {
	for fields, i := m.Descriptor().Fields(), 0; i < fields.Len(); i++ {
		fd := fields.Get(i)
		if !m.Has(fd) {
			continue
		}
		name := string(m.Descriptor().Name()) + "." + string(fd.Name())
		switch {
		case fd.Kind() == protoreflect.MessageKind && fd.IsList():
			for l, j := m.Get(fd).List(), 0; j < l.Len(); j++ {
				eachString(l.Get(j).Message(), visit)
			}
		case fd.Kind() == protoreflect.MessageKind && !fd.IsMap():
			eachString(m.Get(fd).Message(), visit)
		case fd.Kind() == protoreflect.StringKind && fd.IsList():
			for l, j := m.Get(fd).List(), 0; j < l.Len(); j++ {
				visit(name, l.Get(j).String())
			}
		case fd.Kind() == protoreflect.StringKind:
			visit(name, m.Get(fd).String())
		}
	}
}

func values(m proto.Message) (names, vals []string) {
	eachString(m.ProtoReflect(), func(name, v string) { names, vals = append(names, name), append(vals, v) })
	return names, vals
}

// oldShape is a whole value in any id shape an archive holds and the migrated record does not.
var oldShape = regexp.MustCompile(`^([GQM]\d+|[fcp]-[0-9a-f]+|[a-z-]+-F\d+|L\d+-F\d+|R\d+-\d+)$`)

// THE REMAP CHANGES IDS AND ARGUMENTS, AND NOTHING ELSE (R-8, audit G1). A unit case holds the
// spellings that are an id in one field and the subject's words in another; then, over every
// archived run, every source-data field of every translated body equals its source value with the
// ids in its anchor tokens respelled and nothing else.
func TestIdRemapChangesOnlyIdAndArgumentFields(t *testing.T) {
	t.Run("unit", func(t *testing.T) {
		r := newRemap("source")
		for _, b := range []proto.Message{
			&recordpb.Avenue{AvenueId: proto.String("Q1")}, &recordpb.Avenue{AvenueId: proto.String("Q2")},
			&recordpb.Mint{GapId: proto.String("G2")},
		} {
			r.apply(b, "")
		}
		q1, g2, q2 := r.ids["Q1"], r.ids["G2"], r.ids["Q2"]
		sum := sha256.Sum256([]byte("sourceG2"))
		if want := fmt.Sprintf("G-%x", sum[:4]); g2 != want || !strings.HasPrefix(q1, "Q-") || q1 == q2 {
			t.Fatalf("G2 -> %q, want %q (the first eight hex of sha256(source hash ‖ archived id)); Q1 -> %q, Q2 -> %q", g2, want, q1, q2)
		}
		kept := []proto.Message{
			&recordpb.Proof{Script: proto.String("quartile Q1")},
			&recordpb.Cite{Title: proto.String("G2 summit")},
			&recordpb.Retire{SupersededBy: proto.String("Q2 earnings")},
			&recordpb.Mint{MintKey: proto.String("G2")},
		}
		for _, b := range kept {
			was := proto.Clone(b)
			if r.apply(b, ""); !proto.Equal(was, b) {
				t.Errorf("a source-data field was rewritten: %v -> %v", was, b)
			}
		}
		// A comment that only looks like a token, and a hyphenated word, are a seat's prose.
		// So is a hyphen, though an archived citation was labelled with one.
		r.apply(&recordpb.Cite{Label: proto.String("-")}, "")
		shown := &recordpb.Log{Text: proto.String("the form is <!--cite:---> on an unverified-composition gap")}
		asWritten := proto.Clone(shown)
		if r.apply(shown, ""); !proto.Equal(asWritten, shown) {
			t.Errorf("prose that names no id was rewritten: %v (table %v)", shown, r.ids)
		}
		l := &recordpb.Log{Text: proto.String("closes G2, not G20 or G2x")}
		d := &recordpb.Dispatch{GapIds: []string{"G2", "G9"}}
		m := &recordpb.Mint{GapId: proto.String("G3"), Problem: proto.String("as G2 found"), Supersedes: []string{"G2"}}
		mo := &recordpb.MotionRule{Ruling: &recordpb.MotionRule_Docket{Docket: &recordpb.DocketRuling{Principle: proto.String("G2 stands")}}}
		p := &recordpb.Proof{ProofId: proto.String("p-0a1b2c3d"), Drift: proto.String("G2 drifted"), Location: proto.String("x<!--gap:G2--><!--cite:c-00112233-->")}
		for _, b := range []proto.Message{l, d, m, mo, p} {
			r.apply(b, "")
		}
		for what, got := range map[string]string{
			"Log.text":               l.GetText(),
			"Dispatch.gap_ids":       strings.Join(d.GetGapIds(), ","),
			"Mint.problem":           m.GetProblem(),
			"Mint.supersedes":        m.GetSupersedes()[0],
			"DocketRuling.principle": mo.GetDocket().GetPrinciple(),
			"Proof.drift":            p.GetDrift(),
			"Proof.proof_id":         p.GetProofId(),
			"Proof.location":         p.GetLocation(),
		} {
			want := map[string]string{
				"Log.text": "closes " + g2 + ", not G20 or G2x", "Dispatch.gap_ids": g2 + ",G9", "Mint.problem": "as " + g2 + " found",
				"Mint.supersedes": g2, "DocketRuling.principle": g2 + " stands", "Proof.drift": g2 + " drifted",
				"Proof.proof_id": "P-0a1b2c3d", "Proof.location": "x<!--gap:" + g2 + "--><!--cite:C-00112233-->",
			}[what]
			if got != want {
				t.Errorf("%s = %q, want %q", what, got, want)
			}
		}
		// A finding's archived label names its id from then on, and a finding the archive gave no id
		// takes one from its label.
		f := &recordpb.Finding{Id: proto.String("f-7cdcd115")}
		r.apply(f, "adversary-F2")
		legacy := &recordpb.Finding{}
		r.apply(legacy, "L5-F1")
		credit := &recordpb.Mint{GapId: proto.String("G4"), FoundBy: []string{"adversary-F2", "L5-F1"}}
		r.apply(credit, "")
		if f.GetId() != "F-7cdcd115" || credit.GetFoundBy()[0] != "F-7cdcd115" || credit.GetFoundBy()[1] != legacy.GetId() || !currentID.MatchString(legacy.GetId()) {
			t.Errorf("finding %q, legacy %q, credited as %v", f.GetId(), legacy.GetId(), credit.GetFoundBy())
		}
		// An id already in the one shape stands, so a record written at this epoch migrates unchanged.
		now := &recordpb.Mint{GapId: proto.String("G-5e10a3c2"), Location: proto.String("x<!--gap:G-5e10a3c2-->")}
		was := proto.Clone(now)
		if r.apply(now, ""); !proto.Equal(was, now) {
			t.Errorf("a current id was respelled: %v", now)
		}
	})

	changed, minted, strays := map[string]int{}, map[string]int{}, 0
	for _, name := range archivedRuns(t) {
		runDir := recordtest.ExtractArchive(t, name)
		evs, err := archivedSource(t, runDir).Events()
		if err != nil {
			t.Fatal(err)
		}
		hash, _ := hashEvents(evs)
		evs, _ = serializeInstances(evs)
		rm, reg := newRemap(hash), Entries()
		for _, old := range evs { // the table is filled first, as Replay fills it
			bodies, _ := reg.Translate(old, record.Run{})
			for _, body := range bodies {
				rm.apply(body, old.Fields["label"])
			}
		}
		for _, old := range evs {
			bodies, err := reg.Translate(old, record.Run{})
			if err != nil {
				continue // a refused or lost word translates to nothing the remap could change
			}
			for _, body := range bodies {
				names, was := values(body)
				rm.apply(body, old.Fields["label"])
				_, now := values(body)
				for i, field := range names {
					class := "source-data"
					switch {
					case idFields[field]:
						class = "id"
					case argumentFields[field]:
						class = "seat-argument"
						strays += len(strayID.FindAllString(now[i], -1))
					case tokenless(was[i]) != tokenless(now[i]):
						t.Errorf("%s event %d: source-data field %s changed beyond its anchor tokens:\n  was %q\n  now %q", name, old.ID, field, was[i], now[i])
					}
					if was[i] != now[i] {
						changed[class]++
					}
				}
			}
		}
		for _, id := range rm.ids {
			minted[id[:1]]++
		}
		// THE PROOF STORE STILL ANSWERS TO ITS SHA: a migrated proof names its script by the sha256 of
		// the bytes the store holds under it.
		scripts, _ := filepath.Glob(filepath.Join(runDir, "proofs", "*", "script*"))
		for _, script := range scripts {
			body, err := os.ReadFile(script)
			if err != nil {
				t.Fatal(err)
			}
			if got, want := fmt.Sprintf("%x", sha256.Sum256(body)), filepath.Base(filepath.Dir(script)); got != want {
				t.Errorf("%s: the stored script under proof sha %s hashes to %s", name, want, got)
			}
		}
	}
	if changed["id"] == 0 || changed["seat-argument"] == 0 || changed["source-data"] == 0 {
		t.Errorf("fields rewritten per class = %v — the archive holds ids, ids in prose and anchor tokens, so a zero is a pass that did not run", changed)
	}
	t.Logf("census over the archive: archived ids and labels remapped per kind letter %v; fields rewritten per class %v; id-shaped words left in seat-argument fields (no id the run minted) %d", minted, changed, strays)
}

// strayID is a word in an archived id shape, for the census of what the seat-argument pass left.
var strayID = regexp.MustCompile(`\b([GQM]\d+|[fcp]-[0-9a-f]{8}|[a-z]+(?:-[a-z]+)*-F\d+|L\d+-F\d+|R\d+-\d+)\b`)

// NO ARCHIVED ID SURVIVES A MIGRATION (R9). Over every archived run, no string field of the
// migrated record holds a whole value in an archived id shape — a missed id field fails here,
// listed or not — and no seat-argument field holds, at a word boundary, an id the run minted.
// `*_key` fields are a seat's own handles and are left as written: they hold such values (`G1`,
// `logic-F1`) and name nothing on the record.
func TestNoOldIdSurvivesMigration(t *testing.T) {
	for _, name := range archivedRuns(t) {
		toDir := recordtest.TmpRun(t)
		res, err := Migrate(recordtest.ExtractArchive(t, name), toDir, Entries(), Options{})
		if err != nil {
			t.Fatalf("migrate %s: %v", name, err)
		}
		run, err := record.OpenRun(toDir)
		if err != nil {
			t.Fatal(err)
		}
		merged, err := record.MergedEvents(run)
		if err != nil || len(merged.Events) == 0 {
			t.Fatalf("%s: reading the migrated record: %d events, %v", name, len(merged.Events), err)
		}
		evs := merged.Events
		var olds []string
		for old := range res.IDs {
			olds = append(olds, regexp.QuoteMeta(old))
		}
		mapped := regexp.MustCompile(`\b(?:` + strings.Join(olds, "|") + `)\b`)
		for _, e := range evs {
			body, ok := recordpb.Body(e)
			if !ok {
				continue
			}
			eachString(body.ProtoReflect(), func(field, v string) {
				switch {
				case strings.HasSuffix(field, "_key"):
				case oldShape.MatchString(v):
					t.Errorf("%s event %s: %s holds the archived id %q", name, e.GetKey(), field, v)
				case argumentFields[field] && len(olds) > 0 && mapped.MatchString(v):
					t.Errorf("%s event %s: %s still names the archived id %q", name, e.GetKey(), field, mapped.FindString(v))
				}
			})
		}
	}
}

// TWO MIGRATIONS OF ONE SOURCE AGREE ON EVERY ID, wherever they are written: an id is a function of
// the source stream alone, never of the destination or of a counter.
func TestTwoMigrationsOfOneSourceMintTheSameIds(t *testing.T) {
	const name = "2026-09-11_is-91-prime-b9.tar.gz"
	var runs [2][]string
	var tables [2]map[string]string
	for i := range runs {
		toDir := recordtest.TmpRun(t)
		res, err := Migrate(recordtest.ExtractArchive(t, name), toDir, Entries(), Options{})
		if err != nil {
			t.Fatalf("migrate %s: %v", name, err)
		}
		run, err := record.OpenRun(toDir)
		if err != nil {
			t.Fatal(err)
		}
		merged, err := record.MergedEvents(run)
		if err != nil || len(merged.Events) == 0 {
			t.Fatalf("%s: reading the migrated record: %d events, %v", name, len(merged.Events), err)
		}
		evs := merged.Events
		for _, e := range evs {
			if body, ok := recordpb.Body(e); ok {
				_, vals := values(body)
				runs[i] = append(runs[i], e.GetKey()+" "+strings.Join(vals, "\x00"))
			}
		}
		tables[i] = res.IDs
	}
	if len(tables[0]) == 0 || fmt.Sprint(tables[0]) != fmt.Sprint(tables[1]) {
		t.Errorf("the two migrations' id tables differ, or are empty:\n%v\n%v", tables[0], tables[1])
	}
	if strings.Join(runs[0], "\n") != strings.Join(runs[1], "\n") {
		t.Error("the two migrated records differ in a key or a string field")
	}
}
