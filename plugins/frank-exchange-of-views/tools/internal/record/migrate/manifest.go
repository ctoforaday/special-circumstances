package migrate

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"slices"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/buildid"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/migration"
)

// classTranslations is the counted slugs as rows, ordered by slug, never nil.
func classTranslations(mints map[string]int) []migration.ClassTranslation {
	out := []migration.ClassTranslation{}
	for slug, n := range mints {
		out = append(out, migration.ClassTranslation{From: slug, To: archivedGapClasses[slug], Mints: n})
	}
	slices.SortFunc(out, func(a, b migration.ClassTranslation) int { return strings.Compare(a.From, b.From) })
	return out
}

// NewManifest assembles the manifest for one replay.
func NewManifest(sourcePath string, files []migration.SourceFile, unclassified []string, res *Result) *migration.Manifest {
	return &migration.Manifest{
		SourcePath:         sourcePath,
		SourceHash:         res.SourceHash,
		SourceFiles:        files,
		ToolVersion:        buildid.Revision(),
		EventSchema:        record.EventSchema,
		MigratedAt:         record.Now().Format("2006-01-02T15:04:05Z"),
		In:                 res.In,
		Out:                res.Out,
		Refusals:           res.Refusals,
		StatedFills:        res.StatedFills,
		ClassTranslations:  res.ClassTranslations,
		SupersededOutcomes: res.SupersededOutcomes,
		Accepted:           res.AcceptedLosses,
		IDs:                res.IDs,
		Serialized:         res.Serialized,
		GapAnchors:         res.GapAnchors,
		ReportOrder:        res.ReportOrder,
		ProseRewrites:      proseCounts(res.ProseRewrites),
		Rewritten:          res.ProseRewrites,
		Unclassified:       unclassified,
	}
}

// proseCounts is the rewrites per field, never nil.
func proseCounts(rw []migration.ProseRewrite) map[string]int {
	out := map[string]int{}
	for _, r := range rw {
		out[r.Field]++
	}
	return out
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
