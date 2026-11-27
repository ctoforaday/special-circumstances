package fuzz

import (
	"os"
	"strconv"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
)

func TestZZScratchRuns(t *testing.T) {
	bin := buildBinary(t)
	wrapped := debateWrapped(t)
	seeds := []int64{28, 1, 2, 3, 4, 5}
	if s := os.Getenv("ZZ_SEEDS"); s != "" {
		n, _ := strconv.Atoi(s)
		seeds = []int64{int64(n)}
	}
	for _, seed := range seeds {
		o := runOne(t, wrapped, bin, seed, seed == unverifiedSeed, seed == estoppelSeed, seed == verifiedSeed || positionFormOf(seed) != "")
		stale := 0
		if run, err := record.NewRun(o.runDir); err == nil {
			evs, _, _ := record.EventsOf(run, recordpb.EventType_EVENT_TYPE_STALE)
			stale = len(evs)
		}
		t.Logf("seed %d: verdict=%s err=%q stale=%d gates=%v", seed, o.verdict, o.err, stale, o.proposalGates)
		_ = recordsql.CloseUnder(o.runDir)
		_ = os.RemoveAll(o.runDir)
	}
}
