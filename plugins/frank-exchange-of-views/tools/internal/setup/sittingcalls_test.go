package setup

import (
	"bytes"
	"strconv"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/sittingcap"
)

// THE LIMIT SETUP WRITES IS THE LIMIT THE HOOK READS. Setup writes the term through runConfig and
// the hook reads it through sittingcap.Limit, which cannot import this package; this drives both
// ends, so a key renamed on one side fails here instead of every run silently getting the default.
func TestSetupWritesTheSittingCallLimitTheHookReads(t *testing.T) {
	for _, tc := range []struct {
		flag, want int
	}{
		{0, sittingcap.DefaultMaxCalls},
		{40, 40},
	} {
		cfg, runDir := runCfg(t, reports(strconv.Itoa(record.EventSchema)))
		cfg.MaxSittingCalls = tc.flag
		var out, errb bytes.Buffer
		if code := Run(cfg, &out, &errb); code != 0 {
			t.Fatalf("setup exit %d:\n%s", code, errb.String())
		}
		got, err := sittingcap.Limit(runDir)
		if err != nil {
			t.Fatal(err)
		}
		if got != tc.want {
			t.Errorf("--max-sitting-calls %d: the hook reads a limit of %d, want %d", tc.flag, got, tc.want)
		}
	}
}
