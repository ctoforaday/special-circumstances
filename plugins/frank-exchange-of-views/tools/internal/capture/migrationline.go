package capture

import (
	"fmt"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/migrate"
)

// migrationLine is the capture report's statement that this run's record is a TRANSLATION,
// and of what. Empty for a native run — absence is the healthy common case and must not
// share bytes with any other state, which is why an unreadable manifest returns its error
// as the line rather than nothing.
func migrationLine(runDir string) string {
	m, err := migrate.ReadManifest(runDir)
	if err != nil {
		return "migrated record: MANIFEST UNREADABLE — " + err.Error()
	}
	if m == nil {
		return ""
	}
	in, out := 0, 0
	for _, n := range m.In {
		in += n
	}
	for _, n := range m.Out {
		out += n
	}
	// The epoch pair and the turn count are here because this line is where an auditor
	// meets a translated record: the pair says what move was made, and the count is the one part of
	// a migration whose loss would be invisible downstream — an empty timing view reads the same
	// whether the measurements were dropped or never taken.
	epoch := fmt.Sprintf("epoch %d", m.EventSchema)
	if m.SourceEpoch != nil {
		epoch = fmt.Sprintf("epoch %d -> %d", *m.SourceEpoch, m.EventSchema)
	}
	return fmt.Sprintf("migrated record: replayed from %s by %s — %d event(s) in -> %d out, %d refusal(s), %d accepted loss(es), "+
		"%s, %d per-turn measurement(s) brought over; inputs/%s",
		m.SourcePath, m.ToolVersion, in, out, len(m.Refusals), len(m.Accepted), epoch, m.SeatTurns, migrate.ManifestName)
}
