// Package migration is the record a migration leaves in the run it published: the manifest at
// inputs/migration.json, its shape, and the one reader every other party uses to ask "is this run a
// migration, and of what".
//
// It is DATA ONLY and imports nothing of the record. Two parties read it — feov-migrate, which
// writes it, and capture, which states it — and only the first may link the migration itself:
// internal/record/migrate links into feov-migrate and no other binary (scripts/pluginparity), so a
// reader of the manifest reads it from here.
package migration

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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
	SourceHash  string       `json:"source_hash"`
	SourceFiles []SourceFile `json:"source_files"`
	ToolVersion string       `json:"tool_version"`
	// EventSchema is the epoch the migrated record was WRITTEN at — the migrating binary's. SourceEpoch
	// is the one the source stated, absent when the source predates the field, which is the record a
	// migration most exists for. The pair is what makes "the migration moved the epoch" a fact a
	// reader can check instead of an intention: the destination's stored epoch is this field, and the
	// distance travelled is only legible beside where it started.
	EventSchema int  `json:"event_schema"`
	SourceEpoch *int `json:"source_epoch,omitempty"`
	// SeatTurns is how many per-turn measurements were carried across. Zero is a real answer — a run
	// whose capture never went has none — and it is a stated zero rather than an absent field, because
	// a dropped carry and an uncaptured run produce the same empty timing view downstream.
	SeatTurns  int            `json:"seat_turns"`
	MigratedAt string         `json:"migrated_at"`
	In         map[string]int `json:"events_in"`
	Out        map[string]int `json:"events_out"`
	Refusals   []Refusal      `json:"refusals,omitempty"`
	// StatedFills are the values the migration SUPPLIED where the source predates a field and no
	// value can say "never recorded" in its own terms — an enum has no such word. Each names where
	// it was written, the slug, the value and why, so the fill is on the record rather than posing
	// as what the archive said.
	StatedFills []StatedFill `json:"stated_fills,omitempty"`
	// ClassTranslations is each archived class slug some mint carried that resolved through an
	// authored row of migrate's gapclass.go — a merged or dropped registry class, or a coined slug —
	// with the GapClass word it became and how many mints carried it. A slug that is a current word
	// respelled (`report-voice`) has no row: the fold states it already. Empty is stated as [].
	ClassTranslations []ClassTranslation `json:"class_translations"`
	// SupersededOutcomes is each source outcome a later outcome replaced, which the replay did not
	// write (migrate/outcomes.go): the record holds one outcome, the latest. Empty is stated as [].
	SupersededOutcomes []SupersededOutcome `json:"superseded_outcomes"`
	Accepted           map[string]string   `json:"accepted_losses,omitempty"`
	IDs                map[string]string   `json:"ids,omitempty"` // archived id or finding label -> migrated id
	// GapAnchors is what bringing the run's gaps onto anchors did: rewrites per shape, and by name the
	// gaps whose quote never placed (they read gone) and those placed by the fallback.
	GapAnchors GapPlacement `json:"gap_anchors"`
	// ReportOrder is what became of the order of the report's text (migrate's order.go): no_report,
	// in_order, reordered, or as_written for a text out of order that reordering the parts does not
	// replay.
	ReportOrder string `json:"report_order"`
	// ProseRewrites counts, per field, the archived ids rewritten inside a seat's own wording. Such
	// a rewrite cannot tell a reference from a subject word that spells the same — "Q1 revenue" —
	// so the count is stated, and zero is stated too: an absent count and a migration that rewrote
	// no prose would otherwise be the same bytes.
	ProseRewrites map[string]int `json:"prose_rewrites"`
	// Rewritten is each of those rewrites with the text around it, for a reader who judges them —
	// the archive's golden list is read from it.
	Rewritten  []ProseRewrite `json:"-"`
	Serialized map[string]int `json:"serialized_instances,omitempty"` // instance seat -> events moved after instance 1
	// Unclassified names empty old tables no decomposition rule could place — noise, but
	// noise on the record rather than in a log nobody keeps.
	Unclassified []string `json:"unclassified_tables,omitempty"`
	// Discarded names the shard-era sittings the era's own reader would have dropped
	// (winner-per-seat), with the count of keys the winner never rewrote — the genuine loss.
	Discarded []DiscardedSitting `json:"discarded_sittings,omitempty"`

	// MissingArtifacts names each computation (or re-run) the record holds whose stored program or
	// output the source run did not hold, with where it was looked for. Nothing is written in its
	// place: the migrated run's evidence document says the artifact is missing.
	MissingArtifacts []string `json:"missing_artifacts,omitempty"`
	// Warnings are what the publish could not guarantee — a directory whose sync failed. They arise
	// after the manifest is written, so they are never part of it; feov-migrate prints each.
	Warnings []string `json:"-"`
}

// SourceFile names one file the source was read from, with the hash of its bytes AS FOUND.
// These are file hashes, labelled as such; the manifest's source hash is the LOGICAL one
// (hash over the event stream as read), because file bytes differ across WAL checkpoint
// states of the same record.
type SourceFile struct {
	Name   string `json:"name"`
	SHA256 string `json:"sha256"`
	Bytes  int64  `json:"bytes"`
}

// Refusal is one event the replay could not land, with the words of whichever layer refused
// it — the registry, or the current write path's own validation. Refusals are the tool's
// OUTPUT: an old record may hold acts that were legal under the rules of their day, and
// each refusal class gets decided in review. There is no flag that skips validation.
type Refusal struct {
	OldID int64  `json:"old_id"`
	Word  string `json:"word"`
	Err   string `json:"error"`
}

// StatedFill is one value the migration supplied for a field its source predates.
type StatedFill struct {
	Where string `json:"where"` // "mint <gap_id>": a mint whose source predates class_material
	Slug  string `json:"slug"`
	Value string `json:"value"`
	Why   string `json:"why"`
}

// SupersededOutcome is one source outcome the replay did not write, and the one that replaced it.
type SupersededOutcome struct {
	OldID        int64  `json:"old_id"`
	Seat         string `json:"seat"`    // the source's seat id, as written
	Verdict      string `json:"verdict"` // the dropped outcome's verdict, as the record spells it
	SupersededBy int64  `json:"superseded_by"`
}

// ClassTranslation is one archived class slug an authored row carried onto a GapClass word.
type ClassTranslation struct {
	From  string `json:"from"`
	To    string `json:"to"`
	Mints int    `json:"mints"`
}

// GapPlacement is what bringing an archived run's placements onto anchors did (F-c): each Anchor
// added and each rewrite counted by its shape, and by name the gaps with no place in the report — a
// quote that never placed, or an anchor an edit's replacement holds no prose sentence to carry —
// and those placed by the fallback. The shapes are named where each is counted.
type GapPlacement struct {
	Shapes      map[string]int `json:"shapes,omitempty"`
	NeverPlaced []string       `json:"never_placed,omitempty"`
	Fallback    []string       `json:"fallback,omitempty"`
}

// ProseRewrite is one archived id rewritten inside a seat's own wording: the field, the id as
// archived, the id it became, and the text around it as the source wrote it.
type ProseRewrite struct {
	Field   string `json:"field"`
	Old     string `json:"old"`
	New     string `json:"new"`
	Context string `json:"context"`
}

// DiscardedSitting is one superseded dispatch the era's own reader would have dropped, kept
// VISIBLE: the seat, the sitting, how many events it held, and how many of its keys the
// winning sitting never rewrote — the part that is genuine loss rather than a retry's echo.
type DiscardedSitting struct {
	Seat            string `json:"seat"`
	Nonce           string `json:"nonce"`
	Events          int    `json:"events"`
	UnrewrittenKeys int    `json:"unrewritten_keys"`
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
		return nil, fmt.Errorf("%s is present and unreadable — a run that claims to be a migration and cannot say of what must not read as native: %w", ManifestName, err)
	}
	return &m, nil
}
