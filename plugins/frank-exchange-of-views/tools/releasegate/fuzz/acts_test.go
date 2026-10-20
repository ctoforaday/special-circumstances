package fuzz

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// EVERY ACT THE DRIVE ISSUES SAYS WHAT IT EXPECTS THE TOOL TO ANSWER, and the answer is read.
//
// A drive that discards the tool's answer turns a behaviour change into a flake: when the record
// began refusing a self-anchored close, the drive's admitted close was refused, the refusal was
// thrown away, the gap stayed open, and a terminal-fate oracle three verbs later fired on one run
// in forty. Each act therefore goes through one of three doors, and a raw r.exec is reserved for
// a caller that reads both results itself (acts_vet_test.go refuses a discarded one):
//
//   - admit: the drive expects the act ADMITTED. A refusal fails the run, naming the act and the
//     refusal — at the act, not at whichever oracle its absence later trips.
//   - refuse: the drive drives the act TO BE REFUSED. An admission fails the run, and so does a
//     refusal of another code than the one the drive drives: read from the envelope's field.
//   - either: the outcome is legitimately the draw's. The caller checks what it can on the record,
//     says why through noteEither, and the sweep's log tallies each such act by what it got, so a
//     drive that has stopped reaching one side is a number someone can read.

// noteActErr keeps the run's first act whose answer was not the one the drive expects; driveDebate
// fails the run on it ahead of any oracle downstream of the act.
func (r *runner) noteActErr(msg string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.actErr == "" {
		r.actErr = msg
	}
}

// admit runs an act the drive expects the tool to admit. ok is false when it was refused, and the
// run then fails on it.
func (r *runner) admit(args ...string) (out string, ok bool) {
	out, err := r.exec(args...)
	if err != nil {
		r.noteActErr("refused, and the drive expects it admitted: " + err.Error())
		return out, false
	}
	return out, true
}

// refuse runs an act the drive drives to be refused with code. --json is added so the code is the
// envelope's field rather than a reading of the sentence.
func (r *runner) refuse(code feov.Code, args ...string) {
	args = append(args, "--json")
	out, err := r.exec(args...)
	argv := strings.Join(args, " ")
	if err == nil {
		r.noteActErr(argv + ": admitted, and the drive drives it to be refused (" + string(code) + "): " + firstLine(out))
		return
	}
	env, ok := refusalEnvelopeIn(out)
	switch {
	case !ok:
		r.noteActErr(argv + ": refused under --json with no refusal envelope: " + firstLine(out))
	case env.Code != string(code):
		r.noteActErr(fmt.Sprintf("%s: refused with code %q, and the drive drives a %q refusal: %s", argv, env.Code, code, firstLine(env.Error)))
	}
}

// refusalEnvelopeIn is the refusal envelope among out's lines — combined output, so a line on
// stderr may sit beside it.
func refusalEnvelopeIn(out string) (feov.RefusalEnvelope, bool) {
	for _, line := range strings.Split(out, "\n") {
		if env, ok := feov.DecodeRefusal(strings.TrimSpace(line)); ok {
			return env, true
		}
	}
	return feov.RefusalEnvelope{}, false
}

// The either tally: by command and reason, how many were admitted and how many refused.
var (
	eitherMu  sync.Mutex
	eitherRow = map[string]*[2]int{}
)

func noteEither(args []string, why string, admitted bool) {
	key := commandPathOf(args) + " — " + why
	eitherMu.Lock()
	defer eitherMu.Unlock()
	row := eitherRow[key]
	if row == nil {
		row = &[2]int{}
		eitherRow[key] = row
	}
	if admitted {
		row[0]++
	} else {
		row[1]++
	}
}

// eitherReport renders the either tally for the sweep's log.
func eitherReport() string {
	eitherMu.Lock()
	defer eitherMu.Unlock()
	keys := make([]string, 0, len(eitherRow))
	for k := range eitherRow {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	b.WriteString("acts whose outcome is the draw's (admitted/refused):")
	for _, k := range keys {
		fmt.Fprintf(&b, "\n  %s: %d/%d", k, eitherRow[k][0], eitherRow[k][1])
	}
	return b.String()
}

// The fluent builder's doors, and result for a caller that reads both of exec's results.
func (c *cmd) admit() (string, bool)   { return c.r.admit(c.args...) }
func (c *cmd) refuse(code feov.Code)   { c.r.refuse(code, c.args...) }
func (c *cmd) result() (string, error) { return c.r.exec(c.args...) }

// writeSetup writes one file a run is built from, with its directory. Its error is the caller's to
// fail on: a run built without its report or its terms is not the run the sweep measures.
func writeSetup(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("setup: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("setup: %w", err)
	}
	return nil
}

// writeScript writes a proof script into the run directory. A write that fails fails the run: the
// proof it was for is never driven, and the drive would otherwise skip it in silence.
func (r *runner) writeScript(name, body string) error {
	err := os.WriteFile(filepath.Join(r.runDir, name), []byte(body), 0o644)
	if err != nil {
		r.noteActErr("write the proof script " + name + ": " + err.Error())
	}
	return err
}

// sourceTitle is the title the drive gives a source it cites or corroborates. It is a source's own
// title, as a seat is told to give one: a title naming the seat that found the source is refused
// by the report's voice check, and the refusal would stand in for the act the drive meant.
const sourceTitle = "A Survey of the Sampled Trend, Fuzz Press"

// firstThisSitting reports whether seatID has not yet verified or reproduced anchor in its sitting
// in progress, and marks it: the record keeps one such act per anchor per sitting, so a second is a
// repeat the seat would know it made.
func (r *runner) firstThisSitting(seatID, anchor string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.judged == nil {
		r.judged = map[string]map[string]bool{}
	}
	if r.judged[seatID] == nil {
		r.judged[seatID] = map[string]bool{}
	}
	if r.judged[seatID][anchor] {
		return false
	}
	r.judged[seatID][anchor] = true
	return true
}

// corroborations numbers the sources the corroborate drive reads, one fresh source per act.
var corroborations atomic.Int64

// backsTheClaim is the outcomes under which a corroboration places a citation, and so is repeated
// as a retry rather than recorded again.
func backsTheClaim(outcome string) bool {
	return outcome == "supports" || outcome == "supports_with_bridge" || outcome == "weak"
}

// retryCorroboration repeats an admitted corroboration that placed a citation. An identical repeat
// is a retry: admitted, idempotent, and answered with the stored row. A repeat that reads the
// source with another confidence is a second reading, refused as a conflict with the stored one.
func (r *runner) retryCorroboration(args []string, first string) {
	switch {
	case r.coin(20):
		out, ok := r.admit(args...)
		if !ok {
			return
		}
		var a, b struct {
			Result struct {
				Label      string `json:"label"`
				Outcome    string `json:"outcome"`
				Idempotent bool   `json:"idempotent"`
			} `json:"result"`
		}
		if json.Unmarshal([]byte(first), &a) != nil || json.Unmarshal([]byte(out), &b) != nil ||
			!b.Result.Idempotent || b.Result.Label != a.Result.Label || b.Result.Outcome != a.Result.Outcome {
			r.noteActErr("an identical corroboration retry did not answer with the stored row: " + firstLine(out))
		}
	case r.coin(15):
		other := append([]string(nil), args...)
		for i := range other {
			if other[i] == "--confidence" && i+1 < len(other) {
				other[i+1] = map[string]string{"high": "low", "medium": "high", "low": "medium"}[other[i+1]]
			}
		}
		r.refuse(feov.Conflict, other...)
	}
}

// admitFresh runs an act about gapID that the drive expects admitted, as a seat
// issues it: the record refuses an act on a gap another seat changed since the acting seat's read
// point, records a Stale event moving that point, and admits the same act recorded again. A
// refusal is that one only when it is a conflict AND the record gained a Stale event of this seat
// on this gap; then the act is recorded again, up to twice. Any other refusal fails the run.
func (r *runner) admitFresh(gapID string, args ...string) (string, bool) {
	seatID := seatOfArgs(args)
	for range 3 {
		before := r.staleReads(seatID, gapID)
		out, err := r.exec(append(append([]string(nil), args...), "--json")...)
		if err == nil {
			return out, true
		}
		env, ok := refusalEnvelopeIn(out)
		if !ok || env.Code != string(feov.Conflict) || r.staleReads(seatID, gapID) <= before {
			r.noteActErr("refused, and the drive expects it admitted: " + err.Error())
			return out, false
		}
		noteEither(args[:1], "an act on a gap another seat changed since this seat's read point, recorded again", false)
	}
	r.noteActErr(strings.Join(args, " ") + ": refused as a stale read three times running")
	return "", false
}

// staleReads counts the Stale events seatID has on gapID.
func (r *runner) staleReads(seatID, gapID string) int {
	b, ok := r.board()
	if !ok {
		return 0
	}
	n := 0
	for _, e := range b.Events {
		if st, ok := recordpb.BodyAs[*recordpb.Stale](e); ok && e.GetSeatId() == seatID && st.GetGapId() == gapID {
			n++
		}
	}
	return n
}
