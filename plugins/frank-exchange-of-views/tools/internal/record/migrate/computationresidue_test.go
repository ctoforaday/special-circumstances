package migrate

import (
	"database/sql"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"
	_ "modernc.org/sqlite"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/compute"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// The assertions and the accepted residue of plans/compute-and-show.md §III.1.6, each over the
// archives as they are. A count pinned here is the archive's, measured: an archive added to
// run-archive/ that holds none of the shape reads as zero and changes nothing, and one that does
// changes a row here, which is the census the plan's table states.

// sourceEvents is the named archive's events as the source records them, before any translation.
func sourceEvents(t *testing.T, name string) []OldEvent {
	t.Helper()
	src := archivedSource(t, recordtest.RequireArchive(t, name))
	evs, err := src.Events()
	if err != nil {
		t.Fatal(err)
	}
	return evs
}

// sourceProofStage reports whether an archive's source was written before the computation record:
// it holds a `proof` event or a proofs/ store. A native archive writes `computation` and `reproduce`
// as the live record does, and migrates through as written; the proof→computation residue below is
// about the other stage only.
func sourceProofStage(evs []OldEvent, dir string) bool {
	for _, e := range evs {
		if e.Word == "proof" {
			return true
		}
	}
	_, err := os.Stat(filepath.Join(dir, "proofs"))
	return err == nil
}

// proofStage is sourceProofStage over the named archive's source.
func proofStage(t *testing.T, name string) bool {
	t.Helper()
	return sourceProofStage(sourceEvents(t, name), recordtest.RequireArchive(t, name))
}

// sourceRan is a source reproduce's `ran` column read as a bool.
func sourceRan(e OldEvent) bool {
	v := e.Fields["ran"]
	return v == int64(1) || v == true
}

// migratedBodies is every body the named archive's migrated record holds, in record order, with the
// migration's manifest and the migrated run directory.
func migratedBodies(t *testing.T, name string) ([]any, []string, string) {
	t.Helper()
	res, dir := MigratedArchive(t, name)
	if len(res.Refusals) != 0 {
		t.Fatalf("%s: %d refusal(s); the first: %+v", name, len(res.Refusals), res.Refusals[0])
	}
	merged, err := record.MergedEvents(runtest.Open(t, dir))
	if err != nil {
		t.Fatal(err)
	}
	var out []any
	for _, e := range merged.Events {
		if b, ok := recordpb.Body(e); ok {
			out = append(out, b)
		}
	}
	return out, res.MissingArtifacts, dir
}

// archivedProofHex is the eight hex of an archived proof id, lower-cased: `P-` and the era's `p-`
// carry the same hex in two spellings.
func archivedProofHex(id string) string {
	if len(id) == 10 && (strings.HasPrefix(id, "P-") || strings.HasPrefix(id, "p-")) {
		return id[2:]
	}
	return ""
}

// EVERY `K-` ID'S HEX IS ITS SOURCE ID'S, AND EVERY ACT HAS ITS OUTPUT: over every archive, the
// migrated computations are exactly the archived proofs, each under its own hex, and the manifest
// names no artifact the source lacked — every act has acts/<K-id>/output. A native archive's
// computations are its source's, each keeping its `K-` id.
func TestEveryComputationKeepsItsProofsHexAndHasItsOutput(t *testing.T) {
	total, native := 0, 0
	for _, name := range archivedRuns(t) {
		proofs := proofStage(t, name)
		var want []string
		for _, e := range sourceEvents(t, name) {
			switch {
			case proofs && e.Word == "proof":
				id, _ := e.Fields["proof_id"].(string)
				if h := archivedProofHex(id); h != "" {
					want = append(want, h)
				} else {
					t.Errorf("%s: archived proof id %q is not a proof's id", name, id)
				}
			case !proofs && e.Word == "computation":
				id, _ := e.Fields["computation_id"].(string)
				if !strings.HasPrefix(id, "K-") {
					t.Errorf("%s: native computation id %q is not a computation's id", name, id)
				}
				want = append(want, strings.TrimPrefix(id, "K-"))
				native++
			}
		}
		bodies, missing, dir := migratedBodies(t, name)
		if len(missing) != 0 {
			t.Errorf("%s: the manifest names %d artifact(s) the source lacks; the first: %s", name, len(missing), missing[0])
		}
		var got []string
		for _, b := range bodies {
			c, ok := b.(*recordpb.Computation)
			if !ok {
				continue
			}
			got = append(got, strings.TrimPrefix(c.GetComputationId(), "K-"))
			if _, err := compute.ActOutput(dir, c.GetBundleSha(), c.GetComputationId()); err != nil {
				t.Errorf("%s: %s has no acts/<K-id>/output: %v", name, c.GetComputationId(), err)
			}
		}
		slices.Sort(want)
		slices.Sort(got)
		if !slices.Equal(want, got) {
			t.Errorf("%s: migrated computation hex %v, archived hex %v", name, got, want)
		}
		total += len(got)
	}
	if total == 0 {
		t.Fatal("no archive held a proof — a pass that checked nothing")
	}
	if native == 0 {
		t.Fatal("no native archive held a computation — the native branch checked nothing")
	}
}

// AN ERA PROOF'S OUTPUT IS ITS EVENT'S, OR THE FILE ITS EVENT EXCERPTS: the shard era kept a
// proof's output on the event, and that is the act's stored output, byte for byte (is-7-prime-run2's
// `p-568517ec` among them) — except where the event holds the first 2000 bytes of
// proofs/<sha>/output and a line saying so, and then the act's output is the file's. Two acts of
// record-store-authority hold such an excerpt, of 2643 and 4755 bytes.
func TestAnEraProofsOutputIsItsEvents(t *testing.T) {
	checked := 0
	fromFile := map[string]int{}
	for _, name := range archivedRuns(t) {
		inline := map[string]string{}
		src := recordtest.RequireArchive(t, name)
		for _, e := range sourceEvents(t, name) {
			if out, ok := e.Fields["output"].(string); ok && e.Word == "proof" {
				id, _ := e.Fields["proof_id"].(string)
				sha, _ := e.Fields["sha256"].(string)
				if full, err := os.ReadFile(filepath.Join(src, "proofs", sha, "output")); err == nil && len(full) > 2000 &&
					out == string(full[:2000])+"\n… truncated; the full output is in <run>/proofs/<sha256>/output" {
					out = string(full)
					fromFile["K-"+archivedProofHex(id)] = len(full)
				}
				inline["K-"+archivedProofHex(id)] = out
			}
		}
		if len(inline) == 0 {
			continue
		}
		bodies, _, dir := migratedBodies(t, name)
		for _, b := range bodies {
			c, ok := b.(*recordpb.Computation)
			if !ok {
				continue
			}
			want, held := inline[c.GetComputationId()]
			if !held {
				t.Errorf("%s: %s has no era output in its source", name, c.GetComputationId())
				continue
			}
			got, err := compute.ActOutput(dir, c.GetBundleSha(), c.GetComputationId())
			if err != nil || string(got) != want {
				t.Errorf("%s: %s stores %d byte(s) (%v); its event's output is %d", name, c.GetComputationId(), len(got), err, len(want))
			}
			checked++
		}
	}
	if checked != 36 {
		t.Errorf("checked %d era proofs' outputs, want the 36 the six shard-era archives hold", checked)
	}
	if want := map[string]int{"K-5fddd364": 2643, "K-7e3a9ab5": 4755}; !maps.Equal(fromFile, want) {
		t.Errorf("the acts whose event excerpts proofs/<sha>/output: %v, want %v", fromFile, want)
	}
}

// AN ARCHIVED NOTE LABEL IS A COMPUTATION'S: wave-b's finding names `[^P5]`, and the migrated
// finding names `[^K5]`.
func TestAnArchivedNoteLabelReadsAsAComputationsNote(t *testing.T) {
	bodies, _, _ := migratedBodies(t, "2026-10-12_is-91-prime-wave-b.tar.gz")
	k5 := 0
	for _, b := range bodies {
		eachString(b.(proto.Message).ProtoReflect(), func(field, v string) {
			if argumentFields[field] {
				k5 += strings.Count(v, "[^K5]")
				if strings.Contains(v, "[^P5]") {
					t.Errorf("%s keeps [^P5]", field)
				}
			}
		})
	}
	if k5 == 0 {
		t.Error("wave-b's [^P5] reads as no [^K5] in any argument field")
	}
}

var (
	archivedProofIDWord = regexp.MustCompile(`\bP-[0-9a-f]{8}\b`)
	computationIDWord   = regexp.MustCompile(`\bK-([0-9a-f]{8})\b`)
	proofStorePath      = regexp.MustCompile(`proofs/[0-9a-f]`)
)

// EVERY ARCHIVED ID IN THE DEBATE IS A `K-` ID WITH ITS HEX. m19's events name its proofs 114 times
// (the plan's 125 counts every text column of the database, its non-event tables included); the
// migrated record names none by `P-` outside a printed plan, and its computations' hex by `K-` at
// least as often.
func TestM19sProofIdsReadAsComputationIds(t *testing.T) {
	const name = "2026-10-10_is-91-prime-m19.tar.gz"
	hex, archived := map[string]bool{}, 0
	for _, e := range sourceEvents(t, name) {
		if e.Word == "proof" {
			id, _ := e.Fields["proof_id"].(string)
			hex[archivedProofHex(id)] = true
		}
		eachField(e, func(col, v string) {
			if col != "plan.printed" {
				archived += len(archivedProofIDWord.FindAllString(v, -1))
			}
		})
	}
	if archived != 114 {
		t.Fatalf("m19's source names a proof by id %d times, want 114 — the fixture this test reads moved", archived)
	}
	bodies, _, _ := migratedBodies(t, name)
	respelled := 0
	for _, b := range bodies {
		eachString(b.(proto.Message).ProtoReflect(), func(field, v string) {
			if field == "Plan.printed" {
				return
			}
			if archivedProofIDWord.MatchString(v) {
				t.Errorf("%s keeps %q", field, archivedProofIDWord.FindString(v))
			}
			for _, m := range computationIDWord.FindAllStringSubmatch(v, -1) {
				if hex[m[1]] {
					respelled++
				}
			}
		})
	}
	if respelled < archived {
		t.Errorf("the migrated record names m19's computations %d times, the source its proofs %d", respelled, archived)
	}
}

// eachField visits every string column of an archived event, as `<word>.<column>`, and every list
// value.
func eachField(e OldEvent, visit func(col, v string)) {
	for col, v := range e.Fields {
		if s, ok := v.(string); ok {
			visit(e.Word+"."+col, s)
		}
	}
	for col, vs := range e.Lists {
		for _, s := range vs {
			visit(e.Word+"."+col, s)
		}
	}
	for arm, cols := range e.Arms {
		for col, v := range cols {
			if s, ok := v.(string); ok {
				visit(e.Word+"_"+arm+"."+col, s)
			}
		}
	}
}

// reproducesOf is the migrated record's reproduces, in record order.
func reproducesOf(bodies []any) []*recordpb.Reproduce {
	var out []*recordpb.Reproduce
	for _, b := range bodies {
		if r, ok := b.(*recordpb.Reproduce); ok {
			out = append(out, r)
		}
	}
	return out
}

// AN ARCHIVED `not_run` IS AN ACT THAT RAN NOTHING, ON EVERY ARCHIVE (T11): each migrated `ran` =
// false stands on a `ran` reproduce of its bundle before it and states that reproduce's comparison.
// Two archived shapes become one: a `not_run` — 56 over the seven runs the plan's census counted,
// and 43 more in 1565, m35 and m36, archived after it, every one on a re-run that reproduced — and a
// reproduce before epoch 28 carried to a second anchor on its script (reproduceanchor_test.go). A
// native archive writes the act as `ran` = false on its bundle's re-run, and it comes through so.
//
// NO MIGRATED REPRODUCE HOLDS A SENTENCE (Reproduce.claim, ruling L5): the report an archived judgement
// saw is not re-derivable at its event, so migrate reads none back into it, and each archive's count of
// judgements with no recorded sentence is logged once for the plan's residue.
func TestEveryArchivedNotRunStandsOnItsBundlesReRun(t *testing.T) {
	notRun := map[string]int{
		"2026-10-17_is-91-prime-final.tar.gz": 3,
		"2026-10-17_is-91-prime-m27.tar.gz":   5,
		"2026-10-18_is-91-prime-m28.tar.gz":   15,
		"2026-10-18_is-91-prime-m29.tar.gz":   10,
		"2026-10-18_is-91-prime-n1.tar.gz":    7,
		"2026-10-19_is-91-prime-m30.tar.gz":   6,
		"2026-10-24_is-91-prime-m32.tar.gz":   10,
		"2026-10-24_is-91-prime-1565.tar.gz":  18,
		"2026-10-27_is-91-prime-m35.tar.gz":   24,
		"2026-10-29_is-91-prime-m36.tar.gz":   1,
	}
	carried := map[string]int{
		"2026-08-23_research-loop-counterparts.tar.gz": 4,
		"2026-09-02_quadratic-formula.tar.gz":          1,
		"2026-09-09_is-91-prime-b.tar.gz":              20,
		"2026-09-11_is-91-prime-b6.tar.gz":             4,
		"2026-09-11_is-91-prime-b7.tar.gz":             5,
		"2026-09-11_is-91-prime-b8.tar.gz":             3,
		"2026-09-11_is-91-prime-b9.tar.gz":             3,
		"2026-10-09_is-91-prime-wave-a.tar.gz":         32,
		"2026-10-10_is-91-prime-m19.tar.gz":            12,
		"2026-10-11_is-91-prime-m20.tar.gz":            28,
		"2026-10-12_is-91-prime-wave-b.tar.gz":         52,
		"2026-10-13_is-91-prime-m23.tar.gz":            6,
		"2026-10-17_is-91-prime-07d.tar.gz":            27,
		"2026-10-17_is-91-prime-m26.tar.gz":            64,
	}
	// A native archive's acts written `ran` = false, each on its bundle's re-run before it.
	ranNothing := map[string]int{
		"2026-10-30_mersenne-31-prime-sum.tar.gz": 10,
		"2026-11-04_eiffel-tower-height.tar.gz":   8,
	}
	ledger, native, unsentenced := 0, 0, 0
	for _, name := range archivedRuns(t) {
		proofs := proofStage(t, name)
		archived, wrote := 0, 0
		for _, e := range sourceEvents(t, name) {
			if e.Word != "reproduce" {
				continue
			}
			if e.Fields["rerun"] == "not_run" {
				archived++
			}
			if !proofs && !sourceRan(e) {
				wrote++
			}
		}
		if archived != notRun[name] {
			t.Errorf("%s: the archive holds %d not_run, want %d", name, archived, notRun[name])
		}
		if wrote != ranNothing[name] {
			t.Errorf("%s: the native archive holds %d act(s) written ran=false, want %d", name, wrote, ranNothing[name])
		}
		native += wrote
		bodies, _, _ := migratedBodies(t, name)
		standing := map[string]*recordpb.Reproduce{}
		stood := 0
		rs := reproducesOf(bodies)
		for _, r := range rs {
			if r.Claim != nil {
				t.Errorf("%s: migrated reproduce of %s holds the sentence %q; an archived judgement recorded none", name, r.GetComputationId(), r.GetClaim())
			}
		}
		if len(rs) > 0 {
			t.Logf("%s: %d reproduce(s) with no recorded sentence", name, len(rs))
		}
		unsentenced += len(rs)
		for _, r := range rs {
			if r.GetRan() {
				standing[r.GetBundleSha()] = r
				continue
			}
			stood++
			st := standing[r.GetBundleSha()]
			switch {
			case st == nil:
				t.Errorf("%s: %s ran nothing and no reproduce before it ran %s", name, r.GetComputationId(), r.GetBundleSha()[:12])
			case r.GetRerun() != st.GetRerun() || r.GetDrift() != st.GetDrift():
				t.Errorf("%s: %s states %s %q; its bundle's re-run is %s %q", name, r.GetComputationId(), r.GetRerun(), r.GetDrift(), st.GetRerun(), st.GetDrift())
			case r.GetRerunSha() != st.GetRerunSha():
				t.Errorf("%s: %s names execution %q; its bundle's re-run is %q", name, r.GetComputationId(), r.GetRerunSha(), st.GetRerunSha())
			case notRun[name] > 0 && r.GetRerun() != recordpb.Rerun_RERUN_REPRODUCED:
				t.Errorf("%s: %s stands on a re-run that %s; every archived not_run stood on one that reproduced", name, r.GetComputationId(), r.GetRerun())
			}
		}
		if stood != notRun[name]+carried[name]+ranNothing[name] {
			t.Errorf("%s: %d act(s) ran nothing, want %d not_run, %d carried and %d written ran=false", name, stood, notRun[name], carried[name], ranNothing[name])
		}
		if name != "2026-10-24_is-91-prime-1565.tar.gz" && name != "2026-10-27_is-91-prime-m35.tar.gz" && name != "2026-10-29_is-91-prime-m36.tar.gz" {
			ledger += archived
		}
	}
	if ledger != 56 {
		t.Errorf("the plan's census archives hold %d not_run, want the ledger's 56", ledger)
	}
	if native == 0 {
		t.Error("no native archive wrote an act ran=false — the native branch checked nothing")
	}
	if unsentenced == 0 {
		t.Error("no archive migrated a reproduce — the no-sentence assertion checked nothing")
	}
}

// EVERY ARCHIVED `differs` CARRIES WHAT THE PROGRAM WROTE (§III.1.6): the 13 of the database era —
// quadratic-formula 6, is-91-prime-b 2, m20 5 — and the shard era's 4 — is-7-prime-run2 1,
// record-store-authority 3, which the plan's scan did not count — are DIFFERS with drift "on re-run
// the program wrote: " and the archived observed output, byte for byte: no id in a program's bytes
// is respelled, and run2's holds `<!--cite:c-…-->` tokens the program printed.
func TestEveryArchivedDiffersCarriesWhatTheProgramWrote(t *testing.T) {
	want := map[string]int{
		"2026-09-02_quadratic-formula.tar.gz":      6,
		"2026-09-09_is-91-prime-b.tar.gz":          2,
		"2026-10-11_is-91-prime-m20.tar.gz":        5,
		"2026-08-22_is-7-prime-run2.tar.gz":        1,
		"2026-08-22_record-store-authority.tar.gz": 3,
	}
	total := 0
	for _, name := range archivedRuns(t) {
		var observed []string
		for _, e := range sourceEvents(t, name) {
			if v, ok := e.Fields["observed_output"].(string); ok && v != "" && e.Word == "reproduce" {
				observed = append(observed, reproduceDrift(v))
			}
		}
		bodies, _, _ := migratedBodies(t, name)
		var drifts []string
		for _, r := range reproducesOf(bodies) {
			if r.GetRerun() != recordpb.Rerun_RERUN_DIFFERS {
				continue
			}
			if !r.GetRan() {
				t.Errorf("%s: %s differs and ran nothing — no archive holds that case", name, r.GetComputationId())
			}
			drifts = append(drifts, r.GetDrift())
			// A MIGRATED DRIFT IS BOUNDED: the archive truncated what a re-run wrote at 400 bytes
			// and marked the cut, and no more was recorded.
			if n := len(strings.TrimPrefix(r.GetDrift(), reproduceDrift(""))); n > 400+len("…") {
				t.Errorf("%s: %s's drift carries %d bytes of program output", name, r.GetComputationId(), n)
			}
		}
		slices.Sort(observed)
		slices.Sort(drifts)
		if !slices.Equal(observed, drifts) {
			t.Errorf("%s: %d DIFFERS drift(s) do not equal the %d archived observed output(s) under the prefix", name, len(drifts), len(observed))
		}
		if len(drifts) != want[name] {
			t.Errorf("%s: %d DIFFERS, want %d", name, len(drifts), want[name])
		}
		total += len(drifts)
	}
	if total != 17 {
		t.Errorf("%d migrated DIFFERS, want 17", total)
	}
}

// RESIDUE: A PRINTED PLAN KEEPS ITS WORDING. `Plan.printed` is what the plan was when it was
// printed, and its `P-` ids stay: 50 in m28, m29, m30 and m32 (26, 6, 2, 16), and 82 more in the three
// archived after the plan's census.
func TestAPrintedPlanKeepsItsProofIds(t *testing.T) {
	want := map[string]int{
		"2026-10-18_is-91-prime-m28.tar.gz":  26,
		"2026-10-18_is-91-prime-m29.tar.gz":  6,
		"2026-10-19_is-91-prime-m30.tar.gz":  2,
		"2026-10-24_is-91-prime-m32.tar.gz":  16,
		"2026-10-24_is-91-prime-1565.tar.gz": 31,
		"2026-10-27_is-91-prime-m35.tar.gz":  50,
		"2026-10-29_is-91-prime-m36.tar.gz":  1,
	}
	for _, name := range archivedRuns(t) {
		bodies, _, _ := migratedBodies(t, name)
		kept := 0
		for _, b := range bodies {
			if p, ok := b.(*recordpb.Plan); ok {
				kept += len(archivedProofIDWord.FindAllString(p.GetPrinted(), -1))
			}
		}
		if kept != want[name] {
			t.Errorf("%s: the printed plans name %d proof id(s), want %d", name, kept, want[name])
		}
	}
}

// RESIDUE: A SEAT'S ACCOUNT OF WHERE IT READ IS ITS OWN WORDS. A `proofs/<sha>` path in a seat's
// argument stays as it was written; the act it describes keeps its id. Counted over every argument
// field, per archive, as the migrated record holds them — a note carried to a second anchor counts
// once per anchor.
func TestASeatsProofStorePathStaysAsWritten(t *testing.T) {
	want := map[string]int{
		"2026-09-02_quadratic-formula.tar.gz":  15,
		"2026-09-09_is-91-prime-b.tar.gz":      4, // and two in its inquiry reviews, which translate to none
		"2026-09-10_is-91-prime-b3.tar.gz":     2,
		"2026-09-11_is-91-prime-b4.tar.gz":     1,
		"2026-09-11_is-91-prime-b5.tar.gz":     6,
		"2026-09-11_is-91-prime-b7.tar.gz":     11,
		"2026-09-11_is-91-prime-b8.tar.gz":     8,
		"2026-10-09_is-91-prime-wave-a.tar.gz": 3,
		"2026-10-17_is-91-prime-m26.tar.gz":    5,
	}
	for _, name := range archivedRuns(t) {
		bodies, _, _ := migratedBodies(t, name)
		kept := 0
		for _, b := range bodies {
			eachString(b.(proto.Message).ProtoReflect(), func(field, v string) {
				if argumentFields[field] && field != "Computation.drift" && field != "Computation.invariant" {
					kept += len(proofStorePath.FindAllString(v, -1))
				}
			})
		}
		if kept != want[name] {
			t.Errorf("%s: seats' arguments name the proof store %d time(s), want %d", name, kept, want[name])
		}
	}
}

// RESIDUE: A PROGRAM'S BYTES ARE WHAT IT WROTE. A store path inside what a program printed stays:
// 9 migrated drifts (quadratic-formula 2, is-91-prime-b 2, m20 5) and two bundles' stored outputs
// (the plan's census, by its own pattern, counted one).
func TestAStorePathAProgramPrintedStays(t *testing.T) {
	drifts, outputs := 0, 0
	for _, name := range archivedRuns(t) {
		bodies, _, dir := migratedBodies(t, name)
		named := map[string]bool{}
		for _, b := range bodies {
			switch x := b.(type) {
			case *recordpb.Reproduce:
				if proofStorePath.MatchString(x.GetDrift()) {
					drifts++
				}
			case *recordpb.Computation:
				out, err := compute.ActOutput(dir, x.GetBundleSha(), x.GetComputationId())
				if err == nil && proofStorePath.Match(out) && !named[x.GetBundleSha()] {
					named[x.GetBundleSha()] = true
					outputs++
				}
			}
		}
	}
	if drifts != 9 || outputs != 2 {
		t.Errorf("%d drift(s) and %d stored output(s) name the proof store, want 9 and 2", drifts, outputs)
	}
}

// RESIDUE: ACTS SHARING A SCRIPT SHARE ITS ONE STORED OUTPUT. The archive stored one output per
// script, so every act on a migrated bundle holds the same bytes, and every one was recorded
// reproducible — by its basis, the one output is each act's.
func TestActsSharingABundleHoldItsOneOutput(t *testing.T) {
	acts, bundles, runs := 0, 0, 0
	for _, name := range archivedRuns(t) {
		bodies, _, dir := migratedBodies(t, name)
		on := map[string][]*recordpb.Computation{}
		for _, b := range bodies {
			if c, ok := b.(*recordpb.Computation); ok {
				on[c.GetBundleSha()] = append(on[c.GetBundleSha()], c)
			}
		}
		shared := false
		for bundle, cs := range on {
			if len(cs) < 2 {
				continue
			}
			shared = true
			bundles++
			acts += len(cs)
			first, _ := compute.ActOutput(dir, bundle, cs[0].GetComputationId())
			for _, c := range cs {
				if c.GetBasis() != recordpb.Basis_BASIS_REPRODUCIBLE {
					t.Errorf("%s: %s shares %s and is %s", name, c.GetComputationId(), bundle[:12], c.GetBasis())
				}
				if out, _ := compute.ActOutput(dir, bundle, c.GetComputationId()); string(out) != string(first) {
					t.Errorf("%s: %s on %s holds another output than %s", name, c.GetComputationId(), bundle[:12], cs[0].GetComputationId())
				}
			}
		}
		if shared {
			runs++
		}
	}
	// Measured over the 35 archives; the plan's census counted 110 acts on 27 bundles in 20 runs
	// over 30, before 1565, m35, m36, mersenne-31-prime-sum (11 acts on 3 bundles) and
	// eiffel-tower-height (4 acts on 2 bundles) were archived.
	if acts != 173 || bundles != 44 || runs != 28 {
		t.Errorf("%d acts share %d bundles in %d runs, want 173, 44, 28", acts, bundles, runs)
	}
}

// A MIGRATED RE-RUN THAT REPRODUCED NAMES THE EXECUTION LIVE COMPUTE WOULD WRITE (gblock,
// 2026-11-10). An archived reproduce that ran and reproduced names the execution of its act's stored
// output, the act's recorded exit and no out/ — rerun_sha is that execution's compute sha, rerun_exit
// the act's exit, and rerun/<sha>/ holds it, so the per-act comparison against it reproduces. One that
// ran nothing names its bundle's standing re-run. Over the 33 proof-stage archives, 368 re-runs
// reproduced, and 360 acts ran nothing, each on a re-run that reproduced.
//
// RESIDUE: the re-run's exit is the act's, which no archive recorded. A re-run that differed is
// stored only as its first 400 characters, so its reproduce names no execution, nor does one that
// ran nothing on it: 17 differed.
//
// RESIDUE: proofs/ IS NOT CARRIED. The store is rebuilt from the record; drafts under proofs/synth/
// and every other file of the old store stay in the archive, and the argv names the stored file,
// never the author's script path.
//
// A NATIVE ARCHIVE'S RAW RESULTS COME THROUGH: each reproduce's execution and exit, each
// computation's argv, and the store's rerun/ directories are the source's.
func TestAMigratedRunNamesItsArchivedExecutionsAndHoldsNoProofStore(t *testing.T) {
	scriptFile := regexp.MustCompile(`^script\.(py|sh|js|mjs|go)$`)
	native, reproduced, differed, stood := 0, 0, 0, 0
	for _, name := range archivedRuns(t) {
		bodies, _, dir := migratedBodies(t, name)
		if _, err := os.Stat(filepath.Join(dir, "proofs")); !os.IsNotExist(err) {
			t.Errorf("%s: the migrated run holds proofs/ (%v)", name, err)
		}
		if !proofStage(t, name) {
			native++
			nativeRawResultsComeThrough(t, name, bodies, dir)
			continue
		}
		exitOf := map[string]int32{}
		standing := map[string]*recordpb.Reproduce{}
		derived := map[string]bool{}
		for _, b := range bodies {
			if c, ok := b.(*recordpb.Computation); ok {
				exitOf[c.GetComputationId()] = c.GetExit()
			}
			r, ok := b.(*recordpb.Reproduce)
			if !ok {
				continue
			}
			id, bundle := r.GetComputationId(), r.GetBundleSha()
			switch {
			case !r.GetRan():
				stood++
				if want := standing[bundle].GetRerunSha(); r.GetRerunSha() != want || r.RerunExit != nil {
					t.Errorf("%s: %s ran nothing and names %q (exit %v), want its standing re-run's %q", name, id, r.GetRerunSha(), r.RerunExit, want)
				}
			case r.GetRerun() == recordpb.Rerun_RERUN_REPRODUCED:
				reproduced++
				output, err := compute.ActOutput(dir, bundle, id)
				if err != nil {
					t.Fatalf("%s: %s: %v", name, id, err)
				}
				want := compute.Execution{Output: string(output), Exit: int(exitOf[id])}
				if r.GetRerunSha() != want.Sha() || r.RerunExit == nil || r.GetRerunExit() != exitOf[id] {
					t.Errorf("%s: %s names %q (exit %v), want %q, the execution of its output with exit %d", name, id, r.GetRerunSha(), r.RerunExit, want.Sha(), exitOf[id])
					continue
				}
				e, err := compute.ReadRerun(dir, bundle, r.GetRerunSha(), int(r.GetRerunExit()))
				if err != nil {
					t.Errorf("%s: %s: %v", name, id, err)
					continue
				}
				if c, err := compute.Compare(dir, bundle, id, int(exitOf[id]), e); err != nil || !c.Matches {
					t.Errorf("%s: %s's comparison against the execution it names does not reproduce: %+v %v", name, id, c, err)
				}
				derived[filepath.Join(compute.StoreDir, bundle, "rerun", r.GetRerunSha())] = true
			default:
				differed++
				if r.GetRerunSha() != "" || r.RerunExit != nil {
					t.Errorf("%s: %s's re-run %s and names an execution (%q, exit %v)", name, id, recordpb.Word(r.GetRerun()), r.GetRerunSha(), r.RerunExit)
				}
			}
			if r.GetRan() {
				standing[bundle] = r
			}
		}
		for _, b := range bodies {
			if c, ok := b.(*recordpb.Computation); ok {
				if argv := c.GetArgv(); len(argv) == 0 || !scriptFile.MatchString(argv[len(argv)-1]) {
					t.Errorf("%s: %s's argv %q does not name the stored file", name, c.GetComputationId(), argv)
				}
			}
		}
		reruns, _ := filepath.Glob(filepath.Join(dir, compute.StoreDir, "*", "rerun", "*"))
		held := map[string]bool{}
		for _, p := range reruns {
			rel, _ := filepath.Rel(dir, p)
			held[rel] = true
		}
		// The verdict reads as a live one does: a bundle whose re-run reproduced has it standing.
		comps, err := record.RecordedComputations(runtest.Open(t, dir))
		if err != nil {
			t.Fatal(err)
		}
		for _, p := range comps {
			if r := p.Rerun; r != nil && r.Rerun == recordpb.Rerun_RERUN_REPRODUCED && (p.Standing == nil || p.Standing.RerunSha == "") {
				t.Errorf("%s: %s reproduced and its bundle has no standing re-run", name, p.Label)
			}
		}
		if !maps.Equal(held, derived) {
			t.Errorf("%s: the migrated store's rerun/ entries %v, the executions its reproduces name %v", name, slices.Sorted(maps.Keys(held)), slices.Sorted(maps.Keys(derived)))
		}
	}
	if reproduced != 368 || differed != 17 || stood != 360 {
		t.Errorf("%d archived re-runs reproduced, %d differed and %d acts ran nothing, want 368, 17 and 360", reproduced, differed, stood)
	}
	if native == 0 {
		t.Error("no native archive — the check that its raw results come through checks nothing")
	}
	// The drafts this residue row names are in the archive, so the row is about something.
	if _, err := os.Stat(filepath.Join(recordtest.RequireArchive(t, synthArchive), "proofs", "synth")); err != nil {
		t.Errorf("%s holds no proofs/synth/: %v", synthArchive, err)
	}
}

// nativeRawResultsComeThrough asserts a native archive's reproduces name the source's executions and
// exits in order, its computations keep the source's argv, and the store holds the source's rerun/
// directories and no other.
func nativeRawResultsComeThrough(t *testing.T, name string, bodies []any, dir string) {
	t.Helper()
	type raw struct {
		id, sha string
		exit    any
	}
	var want, got []raw
	argv := map[string][]string{}
	for _, e := range sourceEvents(t, name) {
		switch e.Word {
		case "reproduce":
			id, _ := e.Fields["computation_id"].(string)
			sha, _ := e.Fields["rerun_sha"].(string)
			var exit any
			if v, ok := e.Fields["rerun_exit"].(int64); ok {
				exit = int32(v)
			}
			want = append(want, raw{id, sha, exit})
		case "computation":
			id, _ := e.Fields["computation_id"].(string)
			argv[id] = e.Lists["argv"]
		}
	}
	for _, r := range reproducesOf(bodies) {
		var exit any
		if r.RerunExit != nil {
			exit = r.GetRerunExit()
		}
		got = append(got, raw{r.GetComputationId(), r.GetRerunSha(), exit})
	}
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Errorf("%s: the migrated reproduces' executions are not the source's:\n got  %v\n want %v", name, got, want)
	}
	if len(want) == 0 || want[0].sha == "" {
		t.Errorf("%s: the native source names no execution — the check that it comes through checks nothing", name)
	}
	for _, b := range bodies {
		if c, ok := b.(*recordpb.Computation); ok {
			if src, held := argv[c.GetComputationId()]; !held || !slices.Equal(c.GetArgv(), src) {
				t.Errorf("%s: %s's argv %q, the source's %q", name, c.GetComputationId(), c.GetArgv(), src)
			}
		}
	}
	reruns := func(root string) []string {
		m, _ := filepath.Glob(filepath.Join(root, compute.StoreDir, "*", "rerun", "*"))
		for i := range m {
			m[i], _ = filepath.Rel(root, m[i])
		}
		return m
	}
	if g, w := reruns(dir), reruns(recordtest.RequireArchive(t, name)); !slices.Equal(g, w) || len(w) == 0 {
		t.Errorf("%s: the migrated store's rerun/ entries %v, the source's %v", name, g, w)
	}
}

// AN ACT THAT RAN NOTHING ON A RE-RUN THAT DIFFERED CARRIES ITS DRIFT (T9). No archive holds the
// case, so m32 is rewritten to hold it: its one re-run of 0b153a35, event 127, differed and wrote "X". Every
// later not_run on that bundle migrates DIFFERS with the drift, and the write's two-way check
// accepts each.
func TestANotRunOnADifferingReRunCarriesItsDrift(t *testing.T) {
	src := recordtest.RequireArchive(t, "2026-10-24_is-91-prime-m32.tar.gz")
	db, err := sql.Open("sqlite", filepath.Join(src, "records", "record.db"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`UPDATE "reproduce" SET "rerun" = 'differs', "observed_output" = 'X' WHERE "event_id" = 127`); err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	to := filepath.Join(recordtest.TmpRun(t), "migrated")
	res, err := Migrate(src, to, Entries(), Options{})
	if cerr := recordsql.CloseUnder(to); cerr != nil && err == nil {
		err = cerr
	}
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Refusals) != 0 {
		t.Fatalf("%d refusal(s); the first: %+v", len(res.Refusals), res.Refusals[0])
	}
	merged, err := record.MergedEvents(runtest.Open(t, to))
	if err != nil {
		t.Fatal(err)
	}
	carried := 0
	for _, e := range merged.Events {
		r, ok := recordpb.BodyAs[*recordpb.Reproduce](e)
		if !ok || r.GetRan() || r.GetRerun() != recordpb.Rerun_RERUN_DIFFERS {
			continue
		}
		if r.GetDrift() != reproduceDrift("X") {
			t.Errorf("%s ran nothing on a re-run that differed and carries %q", r.GetComputationId(), r.GetDrift())
		}
		carried++
	}
	if carried != 10 {
		t.Errorf("%d act(s) that ran nothing carry the differing re-run, want the 10 on 0b153a35", carried)
	}
}

// synthArchive is the one archive whose proof store held drafts beside the stored scripts.
const synthArchive = "2026-08-22_record-store-authority.tar.gz"

// EACH ARCHIVED EXTENSION NAMES THE COMMAND THE ARCHIVE RAN IT WITH, the one stored file its last
// argument. The archives hold `.py`, `.sh` and `.go` scripts; the table's `.js` and `.mjs` rows
// are held here, where an archive cannot hold them for it. An extension no archived interpreter ran
// is refused by name.
func TestEachArchivedExtensionNamesItsCommand(t *testing.T) {
	for script, want := range map[string][]string{
		"a/b/reconcile.py": {"python3", "script.py"},
		"run.sh":           {"bash", "script.sh"},
		"check.js":         {"node", "script.js"},
		"check.mjs":        {"node", "script.mjs"},
		"main.go":          {"go", "run", "script.go"},
		"UPPER.PY":         {"python3", "script.PY"},
	} {
		argv, file, bundle, err := archivedBundle(script, strings.Repeat("0", 64))
		if err != nil || !slices.Equal(argv, want) || file != want[len(want)-1] || bundle == "" {
			t.Errorf("%s: argv %q, file %q, bundle %q, err %v; want argv %q", script, argv, file, bundle, err, want)
		}
	}
	if _, _, _, err := archivedBundle("prog.rb", strings.Repeat("0", 64)); err == nil || !strings.Contains(err.Error(), `".rb"`) {
		t.Errorf("an extension no archived interpreter ran: err %v, want a refusal naming it", err)
	}
}
