package migrate

import (
	"crypto/sha256"
	"fmt"
	"regexp"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/migration"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// LostIDs is the set of source event ids a migration listed in lost_acts. A test that compares
// source acts against migrated ones subtracts these: an archived act the write refuses for want of a
// delivery, or for a gap such a loss took with it, is an accepted loss, and every other act is held
// to its translation as before.
func LostIDs(m *migration.Manifest) map[int64]bool {
	out := map[int64]bool{}
	for _, a := range m.LostActs {
		out[a.OldID] = true
	}
	return out
}

// LostByWord counts a migration's lost acts per source word.
func LostByWord(m *migration.Manifest) map[string]int {
	out := map[string]int{}
	for _, a := range m.LostActs {
		out[a.Word]++
	}
	return out
}

// LostCloseGaps is the migrated id of every gap whose archived close the named run's migration lost:
// the gaps the source closed and the migrated record leaves open for want of a delivery. An archived
// id outside the live shape is mapped as the migration maps it.
func LostCloseGaps(t *testing.T, name string, m *migration.Manifest) map[string]bool {
	t.Helper()
	return lostGapsOf(t, name, m, "close")
}

// LostMintGaps is the migrated id of every gap whose archived mint the named run's migration lost:
// a mint refused because it superseded a gap a lost close left open.
func LostMintGaps(t *testing.T, name string, m *migration.Manifest) map[string]bool {
	t.Helper()
	return lostGapsOf(t, name, m, "mint")
}

// lostGapsOf is the migrated gap id each lost source act of word names in its gap_id.
func lostGapsOf(t *testing.T, name string, m *migration.Manifest, word string) map[string]bool {
	t.Helper()
	lost := LostIDs(m)
	evs, err := archivedSource(t, recordtest.ExtractArchive(t, name)).Events()
	if err != nil {
		t.Fatal(err)
	}
	out := map[string]bool{}
	for _, ev := range evs {
		if ev.Word != word || !lost[ev.ID] {
			continue
		}
		g, _ := ev.Fields["gap_id"].(string)
		if g == "" {
			t.Fatalf("lost %s %d names no gap_id: %v", word, ev.ID, ev.Fields)
		}
		if !liveGapID.MatchString(g) {
			sum := sha256.Sum256([]byte(m.SourceHash + g))
			g = fmt.Sprintf("G-%x", sum[:4])
		}
		out[g] = true
	}
	return out
}

var liveGapID = regexp.MustCompile(`^G-[0-9a-f]{8}$`)
