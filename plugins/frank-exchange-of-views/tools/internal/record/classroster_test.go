package record

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// THE ROSTER AND THE CHECK MUST NAME THE SAME VOCABULARY, and this is the assertion that keeps
// them from drifting. A seat reads the roster to choose a `--neighbor`; validateClass refuses on
// its own view of what is known. If those two ever disagree, the surface teaches a value the tool
// then rejects — which is worse than having no roster at all, because the seat has no reason to
// doubt it.
func TestTheRosterListsExactlyWhatTheNeighborCheckAccepts(t *testing.T) {
	run, lens := liveRegistryRun(t, map[string]recordpb.ClassMaterial{
		"staged-alpha": recordpb.ClassMaterial_CLASS_MATERIAL_ALWAYS,
		"staged-beta":  recordpb.ClassMaterial_CLASS_MATERIAL_BY_GRADE,
	})
	if _, err := Append(lens, &recordpb.ClassNew{
		Slug:            proto.String("coined-gamma"),
		Definition:      proto.String("what it is"),
		Neighbor:        proto.String("staged-alpha"),
		Distinguisher:   proto.String("how to tell them apart"),
		MaterialDefault: recordpb.ClassMaterial_CLASS_MATERIAL_NEVER.Enum(),
	}); err != nil {
		t.Fatal(err)
	}

	rows, err := ClassRoster(run)
	if err != nil {
		t.Fatal(err)
	}
	inRoster := map[string]bool{}
	for _, r := range rows {
		inRoster[r.Slug] = true
	}

	known, _, err := knownClasses(viewOf(t, run))
	if err != nil {
		t.Fatal(err)
	}
	for slug := range known {
		if !inRoster[slug] {
			t.Errorf("the check accepts %q and the roster does not list it — a seat cannot choose a value it is never shown", slug)
		}
	}
	for slug := range inRoster {
		if !known[slug] {
			t.Errorf("the roster lists %q and the check refuses it — the surface teaches a value the tool rejects", slug)
		}
	}
}

// A COINED ROW CARRIES WHAT A NEIGHBOUR CHOICE NEEDS, and a staged row whose registry entry states
// only the slug and the default carries no blank standing in for the rest.
func TestTheRosterSeparatesCoinedRowsFromStagedOnes(t *testing.T) {
	run, lens := liveRegistryRun(t, map[string]recordpb.ClassMaterial{
		"staged-alpha": recordpb.ClassMaterial_CLASS_MATERIAL_ALWAYS,
	})
	if _, err := Append(lens, &recordpb.ClassNew{
		Slug:            proto.String("coined-gamma"),
		Definition:      proto.String("what it is"),
		Neighbor:        proto.String("staged-alpha"),
		Distinguisher:   proto.String("how to tell them apart"),
		MaterialDefault: recordpb.ClassMaterial_CLASS_MATERIAL_NEVER.Enum(),
	}); err != nil {
		t.Fatal(err)
	}
	rows, err := ClassRoster(run)
	if err != nil {
		t.Fatal(err)
	}

	byslug := map[string]ClassRow{}
	for _, r := range rows {
		byslug[r.Slug] = r
	}

	staged, ok := byslug["staged-alpha"]
	if !ok {
		t.Fatal("staged class missing from the roster")
	}
	if staged.Coined {
		t.Error("a staged class is marked as coined on this run")
	}
	// THE WORD THE FLAG TAKES, not the protobuf enum name. A seat reads this column to decide, and
	// then types the value; two spellings of one value would read as two values.
	if staged.MaterialDefault != "always" {
		t.Errorf("staged material default = %q, want %q — the column must hold the word `--material-default` accepts", staged.MaterialDefault, "always")
	}
	if staged.Definition != "" || staged.Neighbor != "" {
		t.Errorf("a staged row carries a definition or neighbour the registry never staged: %+v", staged)
	}

	coined, ok := byslug["coined-gamma"]
	if !ok {
		t.Fatal("coined class missing from the roster")
	}
	if !coined.Coined {
		t.Error("a class coined on this run is not marked as such")
	}
	if coined.Definition == "" || coined.Neighbor == "" || coined.Distinguisher == "" {
		t.Errorf("a coined row dropped fields the record holds: %+v", coined)
	}
	if coined.MaterialDefault != "never" {
		t.Errorf("coined material default = %q, want %q", coined.MaterialDefault, "never")
	}

	// Ordering is part of the contract: staged first, because that is the vocabulary the seat is
	// choosing WITHIN; this run's additions after it.
	if rows[len(rows)-1].Slug != "coined-gamma" {
		t.Errorf("coined classes do not come last: %v", rows)
	}
}

// AN ABSENT REGISTRY IS A REFUSAL, NOT AN EMPTY LIST. This is the shape the whole package keeps
// re-learning: with no registry staged every `--class` is accepted, so a roster printed as empty
// would say "no vocabulary" in the same bytes it uses for "the gate is off".
func TestARunWithNoStagedRegistryIsRefusedRatherThanListedEmpty(t *testing.T) {
	dir := recordtest.TmpRun(t)
	run := mustRun(t, dir)
	_, err := ClassRoster(run)
	if err == nil {
		t.Fatal("a run with no staged registry returned a roster instead of refusing")
	}
	if !strings.Contains(err.Error(), "no gap-class registry") {
		t.Errorf("the refusal does not name what is missing: %v", err)
	}
}

// A STAGED ROW CARRIES WHAT THE REGISTRY STATES OF IT. The shipped registry gives two of its classes
// a definition, a neighbour and a distinguisher, and `class list` printed none of them: m28's lenses
// spent 22 queries on definitions the staged file held. A row the registry states nothing more of
// carries nothing more — the roster invents no text.
func TestAStagedRowCarriesTheDefinitionTheRegistryStates(t *testing.T) {
	dir := recordtest.TmpRun(t)
	run := mustRun(t, dir)
	reg := `{"classes": [
	  {"slug": "staged-defined", "material_default": "by_grade", "definition": "what it is",
	   "neighbor": "staged-bare", "distinguisher": "how to tell them apart"},
	  {"slug": "staged-bare", "material_default": "always"}
	]}`
	if err := os.MkdirAll(run.Records(), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(run.Records(), "class-registry.json"), []byte(reg), 0o644); err != nil {
		t.Fatal(err)
	}
	rows, err := ClassRoster(run)
	if err != nil {
		t.Fatal(err)
	}
	want := []ClassRow{
		{Slug: "staged-defined", MaterialDefault: "by_grade", Definition: "what it is", Neighbor: "staged-bare", Distinguisher: "how to tell them apart"},
		{Slug: "staged-bare", MaterialDefault: "always"},
	}
	if len(rows) != len(want) {
		t.Fatalf("roster = %+v, want %+v", rows, want)
	}
	for i := range want {
		if rows[i] != want[i] {
			t.Errorf("row %d = %+v, want %+v", i, rows[i], want[i])
		}
	}
}
