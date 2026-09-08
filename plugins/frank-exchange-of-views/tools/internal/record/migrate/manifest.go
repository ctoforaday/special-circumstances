package migrate

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash"
	"os"
	"path/filepath"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/buildid"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// ManifestName is where a migrated run says what it is. inputs/ because that is where a
// run's provenance lives (run-config.json is the precedent), and a FIELD-bearing file
// because other machinery acts on it — capture's "migrated from" row reads these fields,
// never prose.
const ManifestName = "migration.json"

// Manifest records what a migration did, in fields a writer could have refused.
type Manifest struct {
	// SourcePath is where the old record was read from, as given.
	SourcePath string `json:"source_path"`
	// SourceHash is the LOGICAL hash: sha256 over the event stream as read. The per-file
	// hashes below are byte hashes of what was on disk, labelled as such — the same record
	// hashes differently across WAL checkpoint states, which is why both exist.
	SourceHash  string            `json:"source_hash"`
	SourceFiles []SourceFile      `json:"source_files"`
	ToolVersion string            `json:"tool_version"`
	EventSchema int               `json:"event_schema"`
	MigratedAt  string            `json:"migrated_at"`
	In          map[string]int    `json:"events_in"`
	Out         map[string]int    `json:"events_out"`
	Refusals    []Refusal         `json:"refusals,omitempty"`
	Accepted    map[string]string `json:"accepted_losses,omitempty"`
	// Unclassified names empty old tables no decomposition rule could place — noise, but
	// noise on the record rather than in a log nobody keeps.
	Unclassified []string `json:"unclassified_tables,omitempty"`
}

// NewManifest assembles the manifest for one replay.
func NewManifest(sourcePath string, files []SourceFile, unclassified []string, res *Result) *Manifest {
	return &Manifest{
		SourcePath:   sourcePath,
		SourceHash:   res.SourceHash,
		SourceFiles:  files,
		ToolVersion:  buildid.Revision(),
		EventSchema:  record.EventSchema,
		MigratedAt:   record.Now().Format("2006-01-02T15:04:05Z"),
		In:           res.In,
		Out:          res.Out,
		Refusals:     res.Refusals,
		Accepted:     res.AcceptedLosses,
		Unclassified: unclassified,
	}
}

// Write lands the manifest in the migrated run's inputs/.
func (m *Manifest) Write(runDir string) error {
	dir := filepath.Join(runDir, "inputs")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, ManifestName), append(b, '\n'), 0o644)
}

// ReadManifest reports whether a run is a migration, and of what. The nil, nil return is
// the honest common case — a native run — and it is DISTINGUISHABLE from an unreadable
// manifest, which is an error: a run that says it is a migration and cannot say of what
// must not read as native.
func ReadManifest(runDir string) (*Manifest, error) {
	b, err := os.ReadFile(filepath.Join(runDir, "inputs", ManifestName))
	if os.IsNotExist(err) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	var m Manifest
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, fmt.Errorf("migrate: %s is present and unreadable — a run that claims to be a migration and cannot say of what must not read as native: %w", ManifestName, err)
	}
	return &m, nil
}

// streamHasher folds the event stream into the logical source hash. The encoding is
// length-prefixed field-by-field rather than fmt.Sprintf of a struct, so a reordering of
// Go struct fields cannot silently change every hash.
type streamHasher struct{ h hash.Hash }

func newStreamHasher() *streamHasher { return &streamHasher{h: sha256.New()} }

func (s *streamHasher) str(parts ...string) {
	for _, p := range parts {
		fmt.Fprintf(s.h, "%d:%s;", len(p), p)
	}
}

func (s *streamHasher) event(ev OldEvent) {
	s.str("event", fmt.Sprint(ev.ID), ev.SeatID, fmt.Sprint(ev.Round), ev.TS, ev.Word, ev.Key)
	for _, k := range sortedKeys(ev.Fields) {
		s.str("field", k, fmt.Sprint(ev.Fields[k]))
	}
	for _, k := range sortedKeys(ev.Lists) {
		s.str("list", k)
		s.str(ev.Lists[k]...)
	}
	for _, k := range sortedKeys(ev.Arms) {
		s.str("arm", k)
		for _, c := range sortedKeys(ev.Arms[k]) {
			s.str("col", c, fmt.Sprint(ev.Arms[k][c]))
		}
	}
}

func (s *streamHasher) sum() string { return hex.EncodeToString(s.h.Sum(nil)) }
