package record_test

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/migrate"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/migration"
)

// The tests in package record that read an archived run read it MIGRATED: an archive's event
// tables are its creating binary's, and migrate is the one thing that brings a record forward.
// Package record cannot import migrate, which imports it; this package can, and hands the
// migration to them before any test runs, with each act the manifest lists in lost_acts, and fails
// a migration the manifest lists a refusal for.
func init() {
	*record.ArchiveMigrator = func(src, dst string) ([]record.LostAct, error) {
		m, err := migrate.Migrate(src, dst, migrate.Entries(), migrate.Options{})
		if err != nil {
			return nil, err
		}
		// A REFUSED ACT IS A MIGRATION THAT FAILED, not a record to test against: every act after it
		// that cites it cascades, and the test would read a record migrate does not write.
		if len(m.Refusals) > 0 {
			return nil, fmt.Errorf("%d act(s) refused migrating %s; the first: %+v", len(m.Refusals), src, m.Refusals[0])
		}
		return lostActs(src, m)
	}
}

// lostActs is each act of m's lost_acts with the source event it was: its word, its time, and the
// migrated id of the gap or motion it names, mapped as the migration maps an archived id.
func lostActs(src string, m *migration.Manifest) ([]record.LostAct, error) {
	if len(m.LostActs) == 0 {
		return nil, nil
	}
	scratch, err := os.MkdirTemp("", "lost-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(scratch)
	records := filepath.Join(src, "records")
	var s migrate.Source
	if _, err := os.Stat(filepath.Join(records, "record.db")); err != nil {
		if s, err = migrate.OpenJSONL(records); err != nil {
			return nil, err
		}
	} else if s, err = migrate.OpenSQLite(records, scratch); err != nil {
		return nil, err
	}
	defer s.Close()
	evs, err := s.Events()
	if err != nil {
		return nil, err
	}
	byID := map[int64]migrate.OldEvent{}
	for _, ev := range evs {
		byID[ev.ID] = ev
	}
	var out []record.LostAct
	for _, a := range m.LostActs {
		ev, ok := byID[a.OldID]
		if !ok {
			return nil, fmt.Errorf("lost act %d is no source event", a.OldID)
		}
		g, _ := ev.Fields["gap_id"].(string)
		mo, _ := ev.Fields["motion_id"].(string)
		out = append(out, record.LostAct{OldID: a.OldID, Word: a.Word, TS: ev.TS, Seat: ev.SeatID, Gap: liveID(m, g), Motion: liveID(m, mo)})
	}
	return out, nil
}

// liveGapID is a gap id in the live shape; any other archived gap id is mapped by hash, as migrate
// maps it (migrate/lostids_test.go).
var liveGapID = regexp.MustCompile(`^G-[0-9a-f]{8}$`)

func liveID(m *migration.Manifest, id string) string {
	if id == "" {
		return ""
	}
	if to, ok := m.IDs[id]; ok {
		return to
	}
	if id[0] == 'G' && !liveGapID.MatchString(id) {
		sum := sha256.Sum256([]byte(m.SourceHash + id))
		return fmt.Sprintf("G-%x", sum[:4])
	}
	return id
}
