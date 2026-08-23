// Package ctxusage reads a session's own context figures out of its transcript.
//
// Every value here is a TRI-STATE: a number and whether it could be measured at all.
// That is not defensiveness, it is the whole contract. The absent case and the
// healthy case are otherwise the same bytes — a session that has never compacted has
// no ceiling to report, and printing a guessed one renders a confident percentage of
// a denominator nobody measured; a scan that could not reach back to the note has no
// turn count, and a partial one reports the STALEST note as the freshest, which is
// the failure the freshness design exists to catch.
//
// Reads are bounded because a hook must not pull 13 MB on a tick, and are best-effort
// because a hook must never fail over provenance: a missing or rotated transcript is
// an unmeasured read, not an error that takes the session's restore with it.
package ctxusage

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"time"
)

// The scan window, fixed rather than described. "The last N KB" is not a
// specification: an implementer picking N decides where this package starts lying,
// and the failure is silent.
const (
	windowBytes = 256 * 1024
	widenBytes  = 1024 * 1024
)

// Measure is what one bounded read could establish. Read the *Known/*Measured flag
// before the number, always: a zero with its flag false means "could not tell", and
// the two are different claims.
type Measure struct {
	// Tokens is the live context size: input + cache-read + cache-creation on the
	// most recent assistant entry. The sum is the figure; input_tokens alone reads 2
	// on a session holding a megabyte.
	Tokens      int
	TokensKnown bool

	// Turns counts assistant entries at or after the note's write time. Measured only
	// when the scan actually reached back past that time.
	Turns         int
	TurnsMeasured bool

	// Dropped is cumulativeDroppedTokens at that same boundary. Growth is monotone
	// only when computed through it: the raw counter RESETS at a boundary
	// (1,001,875 -> 12,823 measured), so a naive subtraction across one goes negative
	// and the stalest note reads as the freshest in the file.
	Dropped      int
	DroppedKnown bool
}

// worthParsing is a CHEAP PREFILTER, and it is a performance device only — every
// line it admits is then parsed properly, and the fields are read from the parsed
// struct, never from the bytes.
//
// It exists because criterion 3's 5 ms budget is not met by unmarshalling every line:
// measured on a live 3.1 MB transcript, the parse-everything read cost p50 19 ms and
// p95 203 ms, 4x to 40x over. Nearly all of that work was spent on user entries and
// tool results, which carry no figure this package reads.
//
// The risk of a prefilter is a FALSE NEGATIVE — a line silently skipped is a
// measurement silently lost — so each marker is a superset of the field it stands in
// for, and a false positive costs only a parse that finds nothing.
//
// A LOOSER first attempt matched bare "assistant", which admitted 47 of 113 lines in a
// real window because tool results carry sourceToolAssistantUUID. Measured admit rates
// on that window: "assistant" 47, `"type":"assistant"` 43, "usage" 53, `"usage"` 43,
// "compact_boundary" 7. A prefilter that admits everything is not a prefilter, and the
// bare form quietly was one.
func worthParsing(line []byte) bool {
	return bytes.Contains(line, []byte(`"type":"assistant"`)) ||
		bytes.Contains(line, []byte(`"usage"`)) ||
		bytes.Contains(line, []byte("compact_boundary"))
}

type entry struct {
	Type      string    `json:"type"`
	Subtype   string    `json:"subtype"`
	Timestamp time.Time `json:"timestamp"`
	Message   struct {
		Usage struct {
			Input         int `json:"input_tokens"`
			CacheRead     int `json:"cache_read_input_tokens"`
			CacheCreation int `json:"cache_creation_input_tokens"`
		} `json:"usage"`
	} `json:"message"`
	CompactMetadata *struct {
		CumulativeDroppedTokens int `json:"cumulativeDroppedTokens"`
	} `json:"compactMetadata"`
}

// Read measures the transcript at path, counting turns since the note was written.
//
// It never returns an error for an absent or unreadable transcript — the caller gets
// an all-unmeasured Measure, which is the honest answer and the one a hook can act on.
func Read(path string, noteWrittenAt time.Time) (Measure, error) {
	m, err := readWindow(path, noteWrittenAt, windowBytes)
	if err != nil {
		return Measure{}, nil
	}
	// Widen ONCE, and ONLY for a missing token figure.
	//
	// The turn count deliberately does NOT trigger a widen, and the reason is measured
	// rather than assumed. For a genuinely old note — the case this design exists for,
	// so the common one — a wider window does not reach back either, so the second read
	// cannot change the answer. Widening on it measured **p95 160 ms against a 5 ms
	// budget** with a 2 ms raw floor. Probing first (one seek, one line, no parse of the
	// rest) still put p50 at 5.9 ms, over budget, because it is a second file open on
	// every call.
	//
	// So Turns gives up after one window and says so. The answer past the window is
	// "I could not see far enough", and paying 30x the budget to say it more slowly does
	// not make it any more true.
	if !m.TokensKnown {
		if wider, err := readWindow(path, noteWrittenAt, widenBytes); err == nil &&
			(wider.TokensKnown || wider.TurnsMeasured) {
			return wider, nil
		}
	}
	return m, nil
}

func readWindow(path string, noteWrittenAt time.Time, window int64) (Measure, error) {
	f, err := os.Open(path)
	if err != nil {
		return Measure{}, err
	}
	defer f.Close()

	size, err := f.Seek(0, io.SeekEnd)
	if err != nil {
		return Measure{}, err
	}
	offset := size - window
	if offset < 0 {
		offset = 0
	}
	if _, err := f.Seek(offset, io.SeekStart); err != nil {
		return Measure{}, err
	}
	buf, err := io.ReadAll(f)
	if err != nil {
		return Measure{}, err
	}

	lines := bytes.Split(buf, []byte("\n"))
	// A mid-file offset lands inside a line; that fragment is not an entry.
	if offset > 0 && len(lines) > 0 {
		lines = lines[1:]
	}

	var m Measure
	var sawOldest bool
	var oldest time.Time

	// The oldest entry is the FIRST complete line: a transcript is append-only, so its
	// lines are already in time order and finding the earliest costs one parse rather
	// than one per line.
	for _, line := range lines {
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}
		var head struct {
			Timestamp time.Time `json:"timestamp"`
		}
		if err := json.Unmarshal(line, &head); err == nil && !head.Timestamp.IsZero() {
			oldest, sawOldest = head.Timestamp, true
		}
		break
	}

	// DECIDE REACHABILITY FIRST, because it decides how much work is worth doing.
	//
	// If the window does not reach back past the note, the turn count is going to be
	// discarded — so counting it is work whose result is thrown away, and on a real
	// transcript that is most of the parsing. In that case only the newest usage figure
	// is needed, and it is found by scanning BACKWARD and stopping at the first hit.
	reaches := offset == 0 || (sawOldest && oldest.Before(noteWrittenAt))

	if !reaches {
		for i := len(lines) - 1; i >= 0; i-- {
			line := lines[i]
			if len(bytes.TrimSpace(line)) == 0 || !worthParsing(line) {
				continue
			}
			var e entry
			if err := json.Unmarshal(line, &e); err != nil {
				continue
			}
			if e.CompactMetadata != nil && !m.DroppedKnown {
				m.Dropped, m.DroppedKnown = e.CompactMetadata.CumulativeDroppedTokens, true
			}
			if e.Type == "assistant" {
				u := e.Message.Usage
				if total := u.Input + u.CacheRead + u.CacheCreation; total > 0 {
					m.Tokens, m.TokensKnown = total, true
					break
				}
			}
		}
		m.TurnsMeasured = false
		return m, nil
	}

	for _, line := range lines {
		if len(bytes.TrimSpace(line)) == 0 || !worthParsing(line) {
			continue
		}
		var e entry
		// A transcript is appended to while it is read, so the final line is routinely
		// half-written. Skipping an unparseable line is right; failing the read is not.
		if err := json.Unmarshal(line, &e); err != nil {
			continue
		}
		switch {
		case e.Type == "assistant":
			u := e.Message.Usage
			if total := u.Input + u.CacheRead + u.CacheCreation; total > 0 {
				m.Tokens, m.TokensKnown = total, true
			}
			if !e.Timestamp.Before(noteWrittenAt) {
				m.Turns++
			}
		case e.CompactMetadata != nil:
			m.Dropped, m.DroppedKnown = e.CompactMetadata.CumulativeDroppedTokens, true
		}
	}

	// Turns is a count because the window saw the note's own era — decided above, before
	// any of this parsing was paid for.
	m.TurnsMeasured = true
	return m, nil
}
