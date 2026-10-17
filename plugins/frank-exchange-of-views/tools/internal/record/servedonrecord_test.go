package record_test

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatturn"
)

// THE RECORD'S TURNS NAME THE SERVING MODEL WHERE NO TRAJECTORY CAN BE FOUND (smoke 07d: 13 seats
// NOT MEASURED, 647 turn rows naming the model). The configuration directory searched holds no
// trajectory, so the answer is the agent's first recorded turn; a seat whose agent has no turn on
// the record stays NOT MEASURED and says the record was asked too.
func TestTheRecordsTurnsNameTheServedModel(t *testing.T) {
	t.Setenv("CLAUDE_CONFIG_DIR", t.TempDir())
	dir, run := bracketedRun(t)
	bracket(t, dir, lensConfig, "agent_lens_1")
	bracket(t, dir, chairConfig, "agent_chair_1")
	if _, err := record.AppendSeatTurns(run, "agent_lens_1", []seatturn.Turn{
		{Index: 0, TSMillis: 1, Model: ""},
		{Index: 1, TSMillis: 2, Model: "claude-haiku-5-5"},
		{Index: 2, TSMillis: 3, Model: "claude-other"},
	}); err != nil {
		t.Fatal(err)
	}
	fam, err := record.FamilyOf(run)
	if err != nil {
		t.Fatal(err)
	}
	rows := map[string]record.SeatModel{}
	for _, sm := range record.SeatModels(run, fam) {
		rows[sm.SeatID] = sm
	}
	if lens := rows[lensSeat]; lens.Served != "claude-haiku-5-5" || !lens.Measured() || lens.NotMeasured != "" {
		t.Errorf("lens row = %+v, want served claude-haiku-5-5 off its first recorded turn naming a model", lens)
	}
	if chair := rows["red-chair"]; chair.Measured() || !strings.Contains(chair.NotMeasured, "the record holds no turn of this agent") {
		t.Errorf("chair row = %+v, want NOT MEASURED, saying the record was asked", chair)
	}
}
