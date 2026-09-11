// Package sittingcap counts a seat's tool calls within ONE SITTING, so the PreToolUse hook can stop
// a seat that has made more than the run's limit of them.
//
// # The sitting is the register
//
// The record already defines a sitting: a seat's Nth sitting is the window opened by its Nth
// register event (record.Clock, events_w). This counts in that same window rather than inventing
// a second one. `register` writes a header naming the agent's seat and sitting number, and every
// call after it is counted against that sitting until the next register replaces the header.
//
// SubagentStart and the sitting_open span were the other candidates, and both mark a DISPATCH
// rather than a sitting. A warm session resumed for a second sitting fires neither, and a seat
// that is the main session of a headless `claude -p` process never fires SubagentStart at all.
// Every sitting under either engine begins with a register.
//
// # Why a file and not the record
//
// The hook runs as a fresh process once per tool call and may not link the record (its import
// graph is gated: integration/surface/hookgraph_test.go). A record write per tool call would also
// put thousands of bookkeeping rows on a record whose readers want acts. So the count lives in the
// run directory, keyed by agent id and sitting, and the record carries the one fact worth keeping
// — that a sitting reached the limit — through feov-sitting-write, spawned once per such sitting.
//
// # Concurrency
//
// Each call appends ONE byte to its sitting's count file with O_APPEND, which the kernel serialises,
// so no two calls lose an increment and the file size is the count. Concurrent seats write
// different files. Parallel calls from one seat can read a size that includes a sibling's append
// made after their own, so at the boundary a call may be refused up to the seat's parallelism early;
// never late. The first call past the limit is chosen by an O_EXCL create, so exactly one call
// hands the limit to the record however many cross it together.
//
// It links nothing expensive: the standard library only.
package sittingcap

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// DefaultMaxCalls is the limit setup writes when the operator names none. The longest single
// sitting measured is about 130 tool calls (red-lens-evidence, B3), so 150 stops a runaway and
// leaves every observed sitting untouched.
const DefaultMaxCalls = 150

// ConfigKey is the run term's key in inputs/run-config.json.
const ConfigKey = "maxSittingCalls"

// Dir is the counters' directory under the run directory.
const Dir = ".sitting-calls"

// Header is what register writes: which seat the agent sits as and which of that seat's sittings
// is open. The sitting number is the record's own (the seat's register count).
type Header struct {
	SeatID  string `json:"seatId"`
	Sitting int    `json:"sitting"`
}

// Decision is what one counted call found.
type Decision struct {
	Header
	Count int
	Limit int
	// Over is true when this call is past the limit and must be refused.
	Over bool
	// First is true for exactly one refused call per sitting: the one that records the limit.
	First bool
}

// Open starts counting a sitting. register calls it after its event is written.
//
// The header is written to a temporary file and renamed into place, so a concurrent reader sees
// the old header or the new one, never a torn one.
func Open(runDir, agentID string, h Header) error {
	if err := validAgent(agentID); err != nil {
		return err
	}
	if h.SeatID == "" || h.Sitting < 1 {
		return fmt.Errorf("sittingcap: a sitting needs a seat and a number from 1, got %q/%d", h.SeatID, h.Sitting)
	}
	dir := filepath.Join(runDir, Dir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	b, err := json.Marshal(h)
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, agentID+".*.tmp")
	if err != nil {
		return err
	}
	if _, err := tmp.Write(b); err != nil {
		tmp.Close()
		os.Remove(tmp.Name())
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmp.Name())
		return err
	}
	if err := os.Rename(tmp.Name(), headerPath(runDir, agentID)); err != nil {
		os.Remove(tmp.Name())
		return err
	}
	return nil
}

// Count counts one tool call for agentID and says whether it is past the limit.
//
// ok is false when the agent has no open sitting — no header, because it never registered as a
// seat. That covers the main session, an operator, and any agent before its first register, and
// none of them is limited.
func Count(runDir, agentID string) (d Decision, ok bool, err error) {
	if runDir == "" || validAgent(agentID) != nil {
		return d, false, nil
	}
	b, err := os.ReadFile(headerPath(runDir, agentID))
	if errors.Is(err, fs.ErrNotExist) {
		return d, false, nil
	}
	if err != nil {
		return d, false, err
	}
	if err := json.Unmarshal(b, &d.Header); err != nil || d.Sitting < 1 {
		return d, false, fmt.Errorf("sittingcap: header for %s does not parse (%v)", agentID, err)
	}
	limit, err := Limit(runDir)
	if err != nil {
		return d, false, err
	}
	d.Limit = limit

	base := filepath.Join(runDir, Dir, fmt.Sprintf("%s.%d", agentID, d.Sitting))
	f, err := os.OpenFile(base+".calls", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return d, false, err
	}
	_, werr := f.Write([]byte{'.'})
	st, serr := f.Stat()
	f.Close()
	if werr != nil {
		return d, false, werr
	}
	if serr != nil {
		return d, false, serr
	}
	d.Count = int(st.Size())
	if d.Count <= d.Limit {
		return d, true, nil
	}
	d.Over = true
	m, err := os.OpenFile(base+".limited", os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if err == nil {
		m.Close()
		d.First = true
	}
	return d, true, nil
}

// Limit reads the run's per-sitting call limit from inputs/run-config.json.
//
// Absent file or absent key is DefaultMaxCalls, the value setup writes when told nothing. A value
// that is present and below 1 is an error: setup refuses it, so one on disk is a config somebody
// edited, and repairing it silently would hide that.
func Limit(runDir string) (int, error) {
	b, err := os.ReadFile(filepath.Join(runDir, "inputs", "run-config.json"))
	if err != nil {
		return DefaultMaxCalls, nil
	}
	var rc map[string]json.RawMessage
	if err := json.Unmarshal(b, &rc); err != nil {
		return DefaultMaxCalls, fmt.Errorf("sittingcap: inputs/run-config.json does not parse: %w", err)
	}
	raw, present := rc[ConfigKey]
	if !present {
		return DefaultMaxCalls, nil
	}
	var n int
	if err := json.Unmarshal(raw, &n); err != nil || n < 1 {
		return DefaultMaxCalls, fmt.Errorf("sittingcap: run-config.json %s = %s — the limit is a whole number of tool calls from 1", ConfigKey, raw)
	}
	return n, nil
}

func headerPath(runDir, agentID string) string {
	return filepath.Join(runDir, Dir, agentID+".json")
}

// validAgent refuses an agent id that would escape the counters' directory when used as a file
// name. The id comes off the hook payload, which is not a trusted input just because we read it.
func validAgent(agentID string) error {
	if agentID == "" || agentID == "." || agentID == ".." || strings.ContainsAny(agentID, `/\:`) {
		return fmt.Errorf("sittingcap: %q cannot name a counter file", agentID)
	}
	for _, r := range agentID {
		if r < 0x20 || r == 0x7f {
			return fmt.Errorf("sittingcap: %q cannot name a counter file", agentID)
		}
	}
	return nil
}
