package stopnudge

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/checkpoint"
	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/ctxusage"
	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/statefile"
)

func ctxAt(tokens int) ctxusage.Measure { return ctxusage.Measure{Tokens: tokens, TokensKnown: true} }

func ctxBands() Thresholds {
	return Thresholds{ContextNotice: 150_000, ContextWarn: 300_000, ContextUrgent: 600_000}
}

func readState(t *testing.T, dir string) (State, statefile.Status) {
	t.Helper()
	return statefile.Read[State](StatePath(dir))
}

// A session with no note is exactly the one that ran to compaction unwarned, so each edge it
// crosses is said once, at its own band, and the line names the act owed to the human.
func TestAContextBandIsSaidOnceAtItsOwnSeverity(t *testing.T) {
	dir := t.TempDir()
	steps := []struct {
		tokens int
		want   Band
	}{
		{149_999, ""},
		{150_000, BandNotice},
		{200_000, ""}, // NOTICE already spent
		{300_000, BandWarn},
		{650_000, BandUrgent},
		{700_000, ""},
	}
	for _, s := range steps {
		d := DecideContext(dir, "s1", false, ctxAt(s.tokens), ctxBands())
		if d.Band != s.want {
			t.Fatalf("at %d tokens: band %q, want %q", s.tokens, d.Band, s.want)
		}
		// Both acts, note first: the human may not be there to hear the second. The note is named
		// by the path the search reads, so a session that complies stops being warned.
		if s.want != "" {
			w, h := strings.Index(d.Emit, "write "+checkpoint.FallbackPath("")), strings.Index(d.Emit, "tell the human")
			if w < 0 || h < 0 || w > h {
				t.Errorf("at %d tokens the line does not name writing the note and then telling the human: %q", s.tokens, d.Emit)
			}
		}
	}
	st, _ := readState(t, dir)
	if !slices.Equal(st.ContextBandsSpent, []Band{BandNotice, BandWarn, BandUrgent}) || st.Emissions != 3 {
		t.Errorf("state after three bands: %+v", st)
	}
	if len(st.BandsSpent) != 0 {
		t.Errorf("context bands leaked into the note bands nudge_answered reads: %v", st.BandsSpent)
	}
}

// Below every edge nothing is said AND nothing is written: the seal record reads the nudge's
// liveness from whether nudge.json exists.
func TestALightContextWithNoNoteWritesNothing(t *testing.T) {
	dir := t.TempDir()
	if d := DecideContext(dir, "s1", false, ctxAt(40_000), ctxBands()); d.Emit != "" {
		t.Fatalf("emitted below every edge: %q", d.Emit)
	}
	if _, err := os.Stat(StatePath(dir)); err == nil {
		t.Error("wrote nudge.json for a session nothing was said to")
	}
}

// A compaction or a fresh start drops the context below an edge, and the next climb past it is
// said again. Without the re-arm, one warning per session is all a long session ever gets.
func TestAContextBandReArmsWhenTheContextFallsBelowItsEdge(t *testing.T) {
	dir := t.TempDir()
	for _, tok := range []int{160_000, 320_000} {
		DecideContext(dir, "s1", false, ctxAt(tok), ctxBands())
	}
	if d := DecideContext(dir, "s1", false, ctxAt(60_000), ctxBands()); d.Emit != "" {
		t.Fatalf("the drop itself emitted: %q", d.Emit)
	}
	st, _ := readState(t, dir)
	if len(st.ContextBandsSpent) != 0 {
		t.Fatalf("a drop below every edge left bands spent: %v", st.ContextBandsSpent)
	}
	if d := DecideContext(dir, "s1", false, ctxAt(155_000), ctxBands()); d.Band != BandNotice {
		t.Errorf("the climb after a compaction was not said again: %+v", d)
	}
}

// A partial drop re-arms only the bands it fell below.
func TestAPartialDropReArmsOnlyTheBandsItFellBelow(t *testing.T) {
	dir := t.TempDir()
	for _, tok := range []int{160_000, 320_000} {
		DecideContext(dir, "s1", false, ctxAt(tok), ctxBands())
	}
	DecideContext(dir, "s1", false, ctxAt(200_000), ctxBands())
	st, _ := readState(t, dir)
	if !slices.Equal(st.ContextBandsSpent, []Band{BandNotice}) {
		t.Errorf("after falling to 200k: spent %v, want [NOTICE]", st.ContextBandsSpent)
	}
}

func TestTheContextPathKeepsTheLoopGuards(t *testing.T) {
	t.Run("stop_hook_active", func(t *testing.T) {
		dir := t.TempDir()
		if d := DecideContext(dir, "s1", true, ctxAt(900_000), ctxBands()); d.Emit != "" {
			t.Errorf("emitted on a re-entry: %q", d.Emit)
		}
	})
	t.Run("unmeasured context abstains", func(t *testing.T) {
		dir := t.TempDir()
		if d := DecideContext(dir, "s1", false, ctxusage.Measure{Tokens: 900_000}, ctxBands()); d.Emit != "" {
			t.Errorf("emitted on an unmeasured figure: %q", d.Emit)
		}
	})
	t.Run("no context edges, no file", func(t *testing.T) {
		dir := t.TempDir()
		if d := DecideContext(dir, "s1", false, ctxAt(900_000), Thresholds{}); d.Emit != "" {
			t.Errorf("emitted with no edges: %q", d.Emit)
		}
		if _, err := os.Stat(StatePath(dir)); err == nil {
			t.Error("wrote state with no edges")
		}
	})
	t.Run("the hard cap is shared", func(t *testing.T) {
		dir := t.TempDir()
		if err := statefile.Write(StatePath(dir), State{SessionID: "s1", Emissions: maxEmissions}); err != nil {
			t.Fatal(err)
		}
		if d := DecideContext(dir, "s1", false, ctxAt(900_000), ctxBands()); d.Emit != "" {
			t.Errorf("emitted past the session cap: %q", d.Emit)
		}
	})
	t.Run("a corrupt record says nothing", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.MkdirAll(filepath.Dir(StatePath(dir)), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(StatePath(dir), []byte("{not json"), 0o644); err != nil {
			t.Fatal(err)
		}
		if d := DecideContext(dir, "s1", false, ctxAt(900_000), ctxBands()); d.Emit != "" {
			t.Errorf("emitted over an unreadable record: %q", d.Emit)
		}
	})
	t.Run("another session's record is not this one's", func(t *testing.T) {
		dir := t.TempDir()
		if err := statefile.Write(StatePath(dir), State{SessionID: "old", Emissions: maxEmissions,
			ContextBandsSpent: []Band{BandNotice}}); err != nil {
			t.Fatal(err)
		}
		if d := DecideContext(dir, "s2", false, ctxAt(160_000), ctxBands()); d.Band != BandNotice {
			t.Errorf("a new session inherited the old one's spent bands or cap: %+v", d)
		}
	})
}

// Criterion 4's byte budget holds at any figure a session can reach.
func TestTheContextLineFitsTheEmissionBudget(t *testing.T) {
	if line := renderContext(999_999_999, BandUrgent); len(line) > 200 {
		t.Errorf("line is %d bytes, budget 200: %q", len(line), line)
	}
}
