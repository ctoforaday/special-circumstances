package capture

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// EnvelopeReadAudit recomputes each read the journal holds against the record at the event it names.
//
// THE ENGINE ACTS ON `record_read` AND CANNOT CHECK IT. The hook writes it and overwrites what a seat
// typed; the one case the engine cannot tell is a dead hook and a seat that typed the field itself,
// which is a value of the right shape nothing wrote. So every read is recomputed here (record.
// RecordReadAt, the fold the hook printed), and a read that differs was not the record's. It also
// names the envelope that carries none: the engine stops on that, so a journal holding one is a run
// the engine stopped, and a run that went on past one is the engine's defect.
//
// The journal holds the rewritten input, which is what the engine received. The seat's own transcript
// keeps what it typed, so this reads the journal and never a transcript.
func EnvelopeReadAudit(run record.Run, results []map[string]any, journalPresent bool) Audit {
	const check = "envelope-read"
	if !journalPresent {
		return Audit{Check: check, Verdict: "SKIP", Detail: "NOT MEASURED — no workflow journal, so no read the engine acted on could be compared with the record"}
	}
	fam, err := record.FamilyOf(run)
	if err != nil {
		return Audit{Check: check, Verdict: "FAIL", Detail: "the record could not be read: " + err.Error()}
	}
	var findings []string
	reads := 0
	for k, r := range results {
		raw, carried := r["record_read"]
		if !carried {
			findings = append(findings, fmt.Sprintf("envelope %d of the journal carries no record_read", k+1))
			continue
		}
		reads++
		b, err := json.Marshal(raw)
		if err != nil {
			findings = append(findings, fmt.Sprintf("envelope %d: its record_read does not encode: %v", k+1, err))
			continue
		}
		var got record.RecordRead
		if err := json.Unmarshal(b, &got); err != nil {
			findings = append(findings, fmt.Sprintf("envelope %d: its record_read does not read as one: %v", k+1, err))
			continue
		}
		want, err := record.RecordReadAt(fam.Events, fam.At, got.Event)
		if err != nil {
			findings = append(findings, fmt.Sprintf("envelope %d: its record_read names event %d: %v", k+1, got.Event, err))
			continue
		}
		if !sameRead(got, want) {
			findings = append(findings, fmt.Sprintf("envelope %d: its record_read at event %d says %s, and the record there says %s", k+1, got.Event, readWords(got), readWords(want)))
		}
	}
	if len(findings) > 0 {
		return Audit{Check: check, Verdict: "FAIL", Detail: strings.Join(findings, "; ")}
	}
	if reads == 0 {
		return Audit{Check: check, Verdict: "SKIP", Detail: "NOT MEASURED — the journal holds no envelope, so no read was compared with the record"}
	}
	return Audit{Check: check, Verdict: "PASS", Detail: fmt.Sprintf("%d read(s) in the journal, each what the record says at the event it names", reads)}
}

func sameRead(a, b record.RecordRead) bool {
	return a.Event == b.Event && a.Halted == b.Halted && a.Outcome == b.Outcome && slices.Equal(a.BenchOwed, b.BenchOwed)
}

func readWords(r record.RecordRead) string {
	outcome := "no outcome"
	if r.Outcome != "" {
		outcome = "outcome " + string(r.Outcome)
	}
	return fmt.Sprintf("bench owed %v, halted %v, %s", r.BenchOwed, r.Halted, outcome)
}
