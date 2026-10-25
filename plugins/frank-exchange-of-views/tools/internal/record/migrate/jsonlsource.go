package migrate

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/migration"
)

// A shard is named events-<seat>-<nonce>.jsonl, and every line in it carries the seat and the
// nonce as fields. The fields are the sitting; the name is only how the era's files were found.
const (
	shardPrefix = "events-"
	shardSuffix = ".jsonl"
)

// isShard is whether a records/ file is one of the era's shards, by the shape every one was named.
func isShard(name string) bool {
	return strings.HasPrefix(name, shardPrefix) && strings.HasSuffix(name, shardSuffix)
}

// jsonlLine is one act as the shard era wrote it: the envelope inline, the body as a loose
// payload keyed by the era's own field names.
type jsonlLine struct {
	Seq     int64          `json:"seq"`
	TS      string         `json:"ts"`
	SeatID  string         `json:"seatId"`
	Nonce   string         `json:"nonce"`
	Round   int32          `json:"round"`
	Type    string         `json:"type"`
	Key     string         `json:"key"`
	Payload map[string]any `json:"payload"`
}

// JSONLSource reads a shard-era record: one JSONL file per (seat, sitting).
//
// # The merge rule, reconstructed without its bug
//
// The era's own reader picked ONE winning sitting per seat and dropped the rest — by file
// mtime, strictly-after, which made the surviving sitting depend on the filesystem (the
// Linux/Windows split measured 2026-08-16; see the history in record/replay.go). The
// migration keeps the era's SEMANTICS — the fold saw only the winner, so the migrated record
// must tell the same story — and replaces the selector: the winning sitting is the one that
// STARTED last, by its first event's nanosecond stamp, which is a fact of the record rather
// than of the copy. What the era silently dropped, the manifest now carries as fields.
//
// Global order is the era's own repaired sort, (TS, SeatID, Seq), minus the winner step's
// filename accident.
type JSONLSource struct {
	events    []OldEvent
	files     []migration.SourceFile
	discarded []migration.DiscardedSitting
}

// OpenJSONL reads every shard under recordsDir. The originals are opened read-only and never
// written; unlike SQLite there is no WAL to replay, so no copy is needed.
func OpenJSONL(recordsDir string) (*JSONLSource, error) {
	entries, err := os.ReadDir(recordsDir)
	if err != nil {
		return nil, err
	}
	var sittings []*sitting
	s := &JSONLSource{}
	for _, e := range entries {
		if e.IsDir() || !isShard(e.Name()) {
			continue
		}
		lines, sum, size, err := readShard(filepath.Join(recordsDir, e.Name()))
		if err != nil {
			return nil, fmt.Errorf("shard %s: %w", e.Name(), err)
		}
		s.files = append(s.files, migration.SourceFile{Name: e.Name(), SHA256: sum, Bytes: size})
		st, err := sittingOf(e.Name(), lines)
		if err != nil {
			return nil, fmt.Errorf("shard %s: %w", e.Name(), err)
		}
		if st != nil {
			sittings = append(sittings, st)
		}
	}
	if len(sittings) == 0 {
		return nil, fmt.Errorf("no event shards in %s", recordsDir)
	}

	// Winner per seat: the sitting that started last. An empty shard cannot win (it has no
	// start), and cannot lose anything either.
	bySeat := map[string][]*sitting{}
	for _, st := range sittings {
		bySeat[st.seat] = append(bySeat[st.seat], st)
	}
	var winners []*sitting
	for _, group := range bySeat {
		sort.SliceStable(group, func(i, j int) bool { return startOf(group[i]) < startOf(group[j]) })
		win := group[len(group)-1]
		winners = append(winners, win)
		wonKeys := map[string]bool{}
		for _, l := range win.lines {
			wonKeys[l.Key] = true
		}
		for _, lost := range group[:len(group)-1] {
			d := migration.DiscardedSitting{Seat: lost.seat, Nonce: lost.nonce, Events: len(lost.lines)}
			for _, l := range lost.lines {
				if !wonKeys[l.Key] {
					d.UnrewrittenKeys++
				}
			}
			s.discarded = append(s.discarded, d)
		}
	}
	sort.Slice(s.discarded, func(i, j int) bool {
		return s.discarded[i].Seat+s.discarded[i].Nonce < s.discarded[j].Seat+s.discarded[j].Nonce
	})

	var all []jsonlLine
	for _, w := range winners {
		all = append(all, w.lines...)
	}
	// The stamps are fixed-width nanosecond UTC, so lexicographic IS chronological.
	sort.SliceStable(all, func(i, j int) bool {
		a, b := all[i], all[j]
		if a.TS != b.TS {
			return a.TS < b.TS
		}
		if a.SeatID != b.SeatID {
			return a.SeatID < b.SeatID
		}
		return a.Seq < b.Seq
	})
	for i, l := range all {
		ev, err := oldEventOf(l, int64(i+1))
		if err != nil {
			return nil, err
		}
		s.events = append(s.events, ev)
	}
	return s, nil
}

// sitting is one (seat, nonce) shard's worth of acts.
type sitting struct {
	seat, nonce string
	lines       []jsonlLine
}

// sittingOf is the sitting a shard's lines are: the seat and the nonce every line carries. A shard
// whose lines name two sittings, or whose name says another than its lines do, is refused — its
// events would otherwise be filed under a seat they were not written by, or dropped. An empty
// shard holds no act and is no sitting: its file is still listed and hashed.
func sittingOf(name string, lines []jsonlLine) (*sitting, error) {
	if len(lines) == 0 {
		return nil, nil
	}
	st := &sitting{seat: lines[0].SeatID, nonce: lines[0].Nonce, lines: lines}
	if st.seat == "" || st.nonce == "" {
		return nil, fmt.Errorf("line 1 names seat %q and nonce %q — a shard line carries both", st.seat, st.nonce)
	}
	for i, l := range lines[1:] {
		if l.SeatID != st.seat || l.Nonce != st.nonce {
			return nil, fmt.Errorf("line %d is seat %q nonce %q and line 1 is seat %q nonce %q — one shard is one sitting",
				i+2, l.SeatID, l.Nonce, st.seat, st.nonce)
		}
	}
	if want := shardPrefix + st.seat + "-" + st.nonce + shardSuffix; name != want {
		return nil, fmt.Errorf("its lines are seat %q nonce %q, which the era wrote to %s", st.seat, st.nonce, want)
	}
	return st, nil
}

func startOf(st *sitting) string {
	if len(st.lines) == 0 {
		return ""
	}
	return st.lines[0].TS
}

func readShard(path string) (lines []jsonlLine, sha string, size int64, err error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, "", 0, err
	}
	defer f.Close()
	h := sha256.New()
	sc := bufio.NewScanner(io.TeeReader(f, h))
	sc.Buffer(make([]byte, 0, 1024*1024), 16*1024*1024)
	n := 0
	for sc.Scan() {
		n++
		raw := strings.TrimSpace(sc.Text())
		if raw == "" {
			continue
		}
		var l jsonlLine
		if err := json.Unmarshal([]byte(raw), &l); err != nil {
			// A TORN LINE IS NOT A ROW TO SKIP. The era's reader counted these as anomalies
			// and read on; a migration that does the same manufactures a record shorter than
			// the one it claims to translate.
			return nil, "", 0, fmt.Errorf("line %d does not parse: %w", n, err)
		}
		lines = append(lines, l)
	}
	if err := sc.Err(); err != nil {
		return nil, "", 0, err
	}
	st, err := f.Stat()
	if err != nil {
		return nil, "", 0, err
	}
	return lines, hex.EncodeToString(h.Sum(nil)), st.Size(), nil
}

// oldEventOf shapes one shard line as an OldEvent. The era spelled multi-word types with
// hyphens where the schema now uses underscores (`friction-none`, `motion-rule`); that is a
// separator convention, not a rename, and it is normalized here so the registry speaks one
// spelling. Payload values normalize to the shapes the registry's converters expect: JSON
// booleans become 0/1, integral numbers become integers, and a list must be a list of
// strings — anything else is a shape this era never wrote, refused by name.
func oldEventOf(l jsonlLine, id int64) (OldEvent, error) {
	ev := OldEvent{
		ID:     id,
		SeatID: l.SeatID,
		Round:  l.Round,
		TS:     l.TS,
		Key:    l.Key,
		Word:   strings.ReplaceAll(l.Type, "-", "_"),
	}
	for k, v := range l.Payload {
		switch t := v.(type) {
		case string:
			setField(&ev, k, t)
		case bool:
			n := int64(0)
			if t {
				n = 1
			}
			setField(&ev, k, n)
		case float64:
			if t != math.Trunc(t) {
				return ev, fmt.Errorf("%s.%s holds %v — the era wrote no fractional numbers, so this is not a shape to guess at", ev.Word, k, t)
			}
			setField(&ev, k, int64(t))
		case []any:
			var vals []string
			for _, item := range t {
				str, ok := item.(string)
				if !ok {
					return ev, fmt.Errorf("%s.%s holds a list of %T — the era's lists held strings", ev.Word, k, item)
				}
				vals = append(vals, str)
			}
			if ev.Lists == nil {
				ev.Lists = map[string][]string{}
			}
			ev.Lists[k] = vals
		case nil:
			// absent is absent
		default:
			return ev, fmt.Errorf("%s.%s holds %T, which this era never wrote", ev.Word, k, v)
		}
	}
	return ev, nil
}

func setField(ev *OldEvent, k string, v any) {
	if ev.Fields == nil {
		ev.Fields = map[string]any{}
	}
	ev.Fields[k] = v
}

func (s *JSONLSource) Events() ([]OldEvent, error)             { return s.events, nil }
func (s *JSONLSource) Files() []migration.SourceFile           { return s.files }
func (s *JSONLSource) Discarded() []migration.DiscardedSitting { return s.discarded }
func (s *JSONLSource) Close() error                            { return nil }
