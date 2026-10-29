package fuzz

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
)

// These hold the split sweep's plumbing without a sweep: a part's tally survives the file, the
// parts' merge folds to what the whole sweep folds to, and a set of tallies that is not one whole
// sweep is refused before any gate reads it.

// populate sets every field reachable from v to a value other than its zero, drawing distinct
// values from *n. It fails on a field the tally file cannot carry: unexported, or with no json
// name, so a field added without one fails here rather than vanishing from every part's file.
func populate(t *testing.T, v reflect.Value, n *int, path string) {
	t.Helper()
	*n++
	switch v.Kind() {
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			f := v.Type().Field(i)
			name, _, _ := strings.Cut(f.Tag.Get("json"), ",")
			if !f.IsExported() || name == "" || name == "-" {
				t.Fatalf("%s.%s has no json name the tally file carries", path, f.Name)
			}
			populate(t, v.Field(i), n, path+"."+f.Name)
		}
	case reflect.Int, reflect.Int64:
		v.SetInt(int64(*n))
	case reflect.String:
		v.SetString(fmt.Sprintf("s%d", *n))
	case reflect.Bool:
		v.SetBool(true)
	case reflect.Map:
		m := reflect.MakeMap(v.Type())
		k := reflect.New(v.Type().Key()).Elem()
		populate(t, k, n, path+"[key]")
		e := reflect.New(v.Type().Elem()).Elem()
		populate(t, e, n, path+"[]")
		m.SetMapIndex(k, e)
		v.Set(m)
	case reflect.Slice:
		e := reflect.New(v.Type().Elem()).Elem()
		populate(t, e, n, path+"[]")
		v.Set(reflect.Append(reflect.MakeSlice(v.Type(), 0, 1), e))
	case reflect.Array:
		for i := 0; i < v.Len(); i++ {
			populate(t, v.Index(i), n, fmt.Sprintf("%s[%d]", path, i))
		}
	default:
		t.Fatalf("%s is a %s, which populate does not fill; teach it, so the round trip covers the field", path, v.Kind())
	}
}

func TestSweepTallyRoundTripsEveryField(t *testing.T) {
	var st sweepTally
	n := 0
	populate(t, reflect.ValueOf(&st).Elem(), &n, "sweepTally")
	path := filepath.Join(t.TempDir(), "tally.json")
	if err := writeTally(path, st); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	got, err := decodeTally(b)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, st) {
		t.Errorf("the tally changed through its file:\n got %+v\nwant %+v", got, st)
	}
}

func TestProcessTallyAddMergesEachKindByItsRule(t *testing.T) {
	a := processTally{
		ExecRuns: map[string]int{"lens mint": 3}, ExecFail: map[string]int{"lens mint": 3},
		ExecFailWhy: map[string]string{"lens mint": "first"}, ExecInProc: map[string]int{"show": 1},
		ExecEdges:    map[string]map[string]int{"red-lens-logic": {"lens mint": 2}},
		ExecFlags:    map[string]map[string]bool{"lens mint": {"class": true}},
		ExecValues:   map[string]map[string]bool{"lens mint --class": {"false_universal": true}},
		NamingWalked: 2, NamingWrong: map[string]string{"line": "a"},
		DrivenView: map[string]bool{"board": true}, MootSpent: true, LookedFirst: 1, RegisteredTwice: 2, SittingsClosed: 7, SittingsLeftOpen: 1, AdmittedNeverClosed: 1, FailedSittings: 2,
		ClassLanded: map[string]bool{"false_universal": true},
		Either:      map[string][2]int{"close — why": {1, 2}}, Envelope: map[string]int{"blue: ok": 1}, Fail: map[string]int{"x": 1},
	}
	b := processTally{
		ExecRuns: map[string]int{"lens mint": 2, "close": 1}, ExecFail: map[string]int{},
		ExecFailWhy: map[string]string{"lens mint": "second", "close": "c"}, ExecInProc: map[string]int{"show": 2},
		ExecEdges:    map[string]map[string]int{"red-lens-logic": {"lens mint": 1, "close": 1}},
		ExecFlags:    map[string]map[string]bool{"lens mint": {"severity": true}},
		ExecValues:   map[string]map[string]bool{"lens mint --class": {"self_attestation": true}},
		NamingWalked: 3, NamingWrong: map[string]string{"line": "b", "other": "c"},
		DrivenView: map[string]bool{"gaps": true}, LookedFirst: 4, RegisteredTwice: 3, SittingsClosed: 5, SittingsLeftOpen: 2, AdmittedNeverClosed: 3, FailedSittings: 4,
		ClassLanded: map[string]bool{"self_attestation": true},
		Either:      map[string][2]int{"close — why": {3, 4}, "edit — w": {0, 1}}, Envelope: map[string]int{"blue: ok": 2}, Fail: map[string]int{"y": 1},
	}
	want := processTally{
		ExecRuns: map[string]int{"lens mint": 5, "close": 1}, ExecFail: map[string]int{"lens mint": 3},
		ExecFailWhy: map[string]string{"lens mint": "first", "close": "c"}, ExecInProc: map[string]int{"show": 3},
		ExecEdges:    map[string]map[string]int{"red-lens-logic": {"lens mint": 3, "close": 1}},
		ExecFlags:    map[string]map[string]bool{"lens mint": {"class": true, "severity": true}},
		ExecValues:   map[string]map[string]bool{"lens mint --class": {"false_universal": true, "self_attestation": true}},
		NamingWalked: 5, NamingWrong: map[string]string{"line": "a", "other": "c"},
		DrivenView: map[string]bool{"board": true, "gaps": true}, MootSpent: true, LookedFirst: 5, RegisteredTwice: 5, SittingsClosed: 12, SittingsLeftOpen: 3, AdmittedNeverClosed: 4, FailedSittings: 6,
		ClassLanded: map[string]bool{"false_universal": true, "self_attestation": true},
		Either:      map[string][2]int{"close — why": {4, 6}, "edit — w": {0, 1}}, Envelope: map[string]int{"blue: ok": 3}, Fail: map[string]int{"x": 1, "y": 1},
	}
	var got processTally
	got.add(a)
	got.add(b)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("merged process tally:\n got %+v\nwant %+v", got, want)
	}
}

// sampleRuns is n runs, each populated with its own values, seeded 1..n, every third one failed.
func sampleRuns(t *testing.T, n int) []runTally {
	t.Helper()
	draw := 0
	runs := make([]runTally, n)
	for i := range runs {
		populate(t, reflect.ValueOf(&runs[i]).Elem(), &draw, "runTally")
		runs[i].Seed = int64(i + 1)
		runs[i].Verdict = []string{"VERIFIED", "UNVERIFIED", "CEILING"}[i%3]
		runs[i].Epochs = 1 + i%4
		runs[i].EpochLimit = i%2 == 0
		runs[i].Err = ""
		if i%3 == 2 {
			runs[i].Err = "seed failed"
		}
	}
	return runs
}

func TestThePartsFoldToWhatTheWholeSweepFolds(t *testing.T) {
	const n, of = 12, 4
	whole := sampleRuns(t, n)
	want := foldRuns(whole)
	if want.completed != n || want.forcedVerdict == "" || want.verifiedVerdict == "" || !want.limitForced {
		t.Fatalf("the sample does not reach the forced seeds' facts: %+v", want)
	}

	backwards := slices.Clone(whole)
	slices.Reverse(backwards)
	if got := foldRuns(backwards); !reflect.DeepEqual(got, want) {
		t.Errorf("the fold depends on run order:\n got %+v\nwant %+v", got, want)
	}

	dir := t.TempDir()
	var wantProcess processTally
	wantProcess.add(processTally{})
	for k := 0; k < of; k++ {
		part := sweepPart{K: k, Of: of}
		st := sweepTally{Part: k, Parts: of, N: n, Process: processTally{
			ExecRuns: map[string]int{"lens mint": k + 1}, MootSpent: k == 0, ClassLanded: map[string]bool{fmt.Sprint(k): true},
		}}
		for _, r := range whole {
			if part.owns(r.Seed) {
				st.Runs = append(st.Runs, r)
			}
		}
		wantProcess.add(st.Process)
		if err := writeTally(filepath.Join(dir, fmt.Sprintf("tally-%d.json", k)), st); err != nil {
			t.Fatal(err)
		}
	}
	merged, err := loadTallies(dir)
	if err != nil {
		t.Fatal(err)
	}
	if got := foldRuns(merged.Runs); !reflect.DeepEqual(got, want) {
		t.Errorf("the parts fold to another sweep than the whole:\n got %+v\nwant %+v", got, want)
	}
	if !reflect.DeepEqual(merged.Process, wantProcess) {
		t.Errorf("merged process:\n got %+v\nwant %+v", merged.Process, wantProcess)
	}
	if merged.Parts != of || merged.N != n {
		t.Errorf("merged states %d part(s) of %d seeds, want %d of %d", merged.Parts, merged.N, of, n)
	}
}

func TestLoadTalliesRefusesWhatIsNotOneWholeSweep(t *testing.T) {
	run := func(seeds ...int64) []runTally {
		var out []runTally
		for _, s := range seeds {
			out = append(out, runTally{Seed: s})
		}
		return out
	}
	tally := func(part, parts, n int, seeds ...int64) string {
		b, err := json.Marshal(sweepTally{Part: part, Parts: parts, N: n, Runs: run(seeds...)})
		if err != nil {
			t.Fatal(err)
		}
		return string(b)
	}
	cases := []struct {
		name  string
		files map[string]string
		want  string
	}{
		{"no file", map[string]string{}, "holds no tally"},
		{"a file that is not json", map[string]string{"a.json": tally(0, 1, 1, 1), "notes.txt": "x"}, "is not a part's tally"},
		{"an unknown field", map[string]string{"a.json": `{"part":0,"parts":1,"n":1,"runs":[{"seed":1}],"process":{},"extra":1}`}, "unknown field"},
		{"a second value", map[string]string{"a.json": tally(0, 1, 1, 1) + tally(0, 1, 1, 1)}, "more than one value"},
		{"parts disagree", map[string]string{"a.json": tally(0, 2, 2, 1), "b.json": tally(1, 3, 2, 2)}, "part(s) and"},
		{"seed counts disagree", map[string]string{"a.json": tally(0, 2, 2, 1), "b.json": tally(1, 2, 4, 2)}, "-seed sweep and"},
		{"a part out of range", map[string]string{"a.json": tally(2, 2, 2, 1)}, "states part 2 of 2"},
		{"a part twice", map[string]string{"a.json": tally(0, 2, 2, 1), "b.json": tally(0, 2, 2, 1)}, "part 0 appears twice"},
		{"a part missing", map[string]string{"a.json": tally(0, 2, 2, 1)}, "part(s) 1 of 2 hold no tally"},
		{"a seed in a part that does not run it", map[string]string{"a.json": tally(0, 2, 2, 2), "b.json": tally(1, 2, 2, 1)}, "which part 0 of 2 does not run"},
		{"a seed twice", map[string]string{"a.json": tally(0, 2, 4, 1, 3, 3), "b.json": tally(1, 2, 4, 2, 4)}, "seed 3 appears twice"},
		{"a seed outside the sweep", map[string]string{"a.json": tally(0, 2, 2, 1, 3), "b.json": tally(1, 2, 2, 2)}, "outside 1..2"},
		{"a seed with no run", map[string]string{"a.json": tally(0, 2, 4, 1, 3), "b.json": tally(1, 2, 4, 2)}, "1 of 4 seeds hold no run: 4"},
		{"no parts stated", map[string]string{"a.json": tally(0, 0, 2, 1)}, "states 0 part(s)"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			dir := t.TempDir()
			for name, body := range c.files {
				if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			_, err := loadTallies(dir)
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Errorf("loadTallies = %v, want a refusal containing %q", err, c.want)
			}
		})
	}
	t.Run("one whole sweep", func(t *testing.T) {
		dir := t.TempDir()
		for name, body := range map[string]string{"a.json": tally(0, 2, 3, 1, 3), "b.json": tally(1, 2, 3, 2)} {
			if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
				t.Fatal(err)
			}
		}
		st, err := loadTallies(dir)
		if err != nil {
			t.Fatal(err)
		}
		var seeds []int64
		for _, r := range st.Runs {
			seeds = append(seeds, r.Seed)
		}
		if !slices.Equal(seeds, []int64{1, 2, 3}) {
			t.Errorf("merged seeds %v, want 1 2 3 in order", seeds)
		}
	})
}

func TestParseSweepPart(t *testing.T) {
	for _, c := range []struct {
		in   string
		want sweepPart
	}{
		{"", wholeSweep},
		{"0/1", sweepPart{0, 1}},
		{"0/4", sweepPart{0, 4}},
		{"3/4", sweepPart{3, 4}},
		{"10/12", sweepPart{10, 12}},
	} {
		got, err := parseSweepPart(c.in)
		if err != nil || got != c.want {
			t.Errorf("parseSweepPart(%q) = %v, %v; want %v", c.in, got, err, c.want)
		}
	}
	for _, in := range []string{"4/4", "0/0", "1", "1/", "/4", "-1/4", "+1/4", "1/+4", "01/4", "1/04", " 1/4", "1/4 ", "1/4/2", "a/4", "1.0/4", "99999999999999999999/1"} {
		if got, err := parseSweepPart(in); err == nil {
			t.Errorf("parseSweepPart(%q) = %v, want a refusal", in, got)
		}
	}
	part := sweepPart{K: 1, Of: 4}
	var owned []int64
	for s := int64(1); s <= 12; s++ {
		if part.owns(s) {
			owned = append(owned, s)
		}
	}
	if !slices.Equal(owned, []int64{2, 6, 10}) {
		t.Errorf("part 1 of 4 owns %v, want 2 6 10", owned)
	}
	if !(sweepPart{K: 0, Of: 4}).armsSweepDrives() || (sweepPart{K: 1, Of: 4}).armsSweepDrives() || !wholeSweep.armsSweepDrives() {
		t.Error("the once-per-sweep drives arm in part 0 and in the whole sweep, and in no other part")
	}
}

// THE FOLD READS EVERY FIELD A RUN CARRIES AND WRITES EVERY FIELD THE GATES READ. A run field the
// fold drops is a count that travels to the merge and is judged by nothing; a fold field nothing
// writes is a gate reading a zero.
func TestTheFoldReadsEveryRunFieldAndWritesEveryFoldField(t *testing.T) {
	runs := sampleRuns(t, 12)
	want := foldRuns(runs)
	fv := reflect.ValueOf(want)
	for i := 0; i < fv.NumField(); i++ {
		if fv.Field(i).IsZero() {
			t.Errorf("sweepFold.%s is zero over runs where every field is set — the fold never writes it", fv.Type().Field(i).Name)
		}
	}
	rt := reflect.TypeOf(runTally{})
	for i := 0; i < rt.NumField(); i++ {
		zeroed := slices.Clone(runs)
		for j := range zeroed {
			reflect.ValueOf(&zeroed[j]).Elem().Field(i).SetZero()
		}
		if reflect.DeepEqual(foldRuns(zeroed), want) {
			t.Errorf("runTally.%s is not read by the fold: zeroing it in every run changes nothing", rt.Field(i).Name)
		}
	}
}
