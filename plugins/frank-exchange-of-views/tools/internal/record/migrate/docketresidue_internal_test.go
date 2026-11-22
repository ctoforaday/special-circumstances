package migrate

import "github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/migration"

// docketResidueIDs is every source id a manifest's DocketResidue lists.
func docketResidueIDs(d migration.DocketResidue) map[int64]bool {
	out := map[int64]bool{}
	for _, l := range [][]int64{d.Opinions, d.PartyDockets, d.Superseded, d.EndedRulings, d.BenchRegrades, d.SpotChecks} {
		for _, id := range l {
			out[id] = true
		}
	}
	return out
}
