package migrate_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/migrate"
)

func writeShard(t *testing.T, dir, seat, nonce string, lines ...string) {
	t.Helper()
	name := fmt.Sprintf("events-%s-%s.jsonl", seat, nonce)
	if err := os.WriteFile(filepath.Join(dir, name), []byte(strings.Join(lines, "\n")+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func line(seat, nonce string, seq int, ts, typ, key, payload string) string {
	return fmt.Sprintf(`{"seq":%d,"ts":%q,"seatId":%q,"nonce":%q,"round":1,"type":%q,"key":%q,"payload":%s}`,
		seq, ts, seat, nonce, typ, key, payload)
}

// THE WINNER IS A FACT OF THE RECORD, NOT OF THE COPY. The era resolved two sittings by file
// mtime — the Linux/Windows split of 2026-08-16 — and a tarball's mtimes are whatever the
// copy preserved. The reconstruction selects by the sitting that STARTED last, and counts
// what the era silently dropped: the loser's events, and the keys the winner never rewrote.
func TestASupersededSittingIsDiscardedAndCounted(t *testing.T) {
	dir := t.TempDir()
	// First sitting: two acts, one of which (the friction) the retry never rewrites.
	writeShard(t, dir, "blue-lane-1", "aaaaaaaa",
		line("blue-lane-1", "aaaaaaaa", 0, ts1, "register", "blue-lane-1:register:aaaaaaaa", `{"tool_version":"old"}`),
		line("blue-lane-1", "aaaaaaaa", 1, ts2, "friction", "blue-lane-1:friction:#1", `{"reason":"lost work"}`),
	)
	// The re-dispatch starts later and rewrites only the register key's shape.
	writeShard(t, dir, "blue-lane-1", "bbbbbbbb",
		line("blue-lane-1", "bbbbbbbb", 0, ts3, "register", "blue-lane-1:register:bbbbbbbb", `{"tool_version":"new"}`),
		line("blue-lane-1", "bbbbbbbb", 1, ts4, "friction_none", "blue-lane-1:friction-none:#1", `{"reason":"clean"}`),
	)
	src, err := migrate.OpenJSONL(dir)
	if err != nil {
		t.Fatal(err)
	}
	evs, err := src.Events()
	if err != nil {
		t.Fatal(err)
	}
	if len(evs) != 2 {
		t.Fatalf("the winner holds 2 acts; got %d", len(evs))
	}
	for _, ev := range evs {
		if ev.TS != ts3 && ev.TS != ts4 {
			t.Errorf("an act of the superseded sitting replayed: %+v", ev)
		}
	}
	d := src.Discarded()
	if len(d) != 1 || d[0].Nonce != "aaaaaaaa" || d[0].Events != 2 || d[0].UnrewrittenKeys != 2 {
		t.Fatalf("the discarded sitting must be counted, unrewritten keys and all: %+v", d)
	}
}

// A torn line is not a row to skip — the era's reader counted anomalies and read on, which
// is how a record gets quietly shorter than the one it claims to be.
func TestATornShardLineRefuses(t *testing.T) {
	dir := t.TempDir()
	writeShard(t, dir, "blue-lane-1", "aaaaaaaa",
		line("blue-lane-1", "aaaaaaaa", 0, ts1, "register", "k1", `{"tool_version":"x"}`),
		`{"seq":1,"ts":"2026-08-22T`, // torn mid-write
	)
	if _, err := migrate.OpenJSONL(dir); err == nil || !strings.Contains(err.Error(), "line 2") {
		t.Fatalf("a torn line must refuse naming itself, got %v", err)
	}
}

// The era spelled words with hyphens; the registry speaks the schema's separators. One
// convention, normalized at the adapter — a REAL rename still refuses downstream.
func TestHyphenatedEraWordsNormalize(t *testing.T) {
	dir := t.TempDir()
	writeShard(t, dir, "blue-lane-1", "aaaaaaaa",
		line("blue-lane-1", "aaaaaaaa", 0, ts1, "friction-none", "k1", `{"reason":"clean"}`),
	)
	src, err := migrate.OpenJSONL(dir)
	if err != nil {
		t.Fatal(err)
	}
	evs, err := src.Events()
	if err != nil {
		t.Fatal(err)
	}
	if evs[0].Word != "friction_none" {
		t.Fatalf("separator normalization: got %q", evs[0].Word)
	}
}
