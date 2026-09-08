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
	return fmt.Sprintf("migrated record: replayed from %s by %s — %d event(s) in -> %d out, %d refusal(s), %d accepted loss(es); inputs/%s",
		m.SourcePath, m.ToolVersion, in, out, len(m.Refusals), len(m.Accepted), migrate.ManifestName)
}
