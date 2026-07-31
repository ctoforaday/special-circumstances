package transcript

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// line builds one transcript record with the given tool_use blocks.
func line(t *testing.T, uses ...[2]string) string {
	t.Helper()
	var bs []map[string]any
	for _, u := range uses {
		bs = append(bs, map[string]any{
			"type":  "tool_use",
			"name":  u[0],
			"input": map[string]any{"file_path": u[1]},
		})
	}
	rec := map[string]any{"type": "assistant", "message": map[string]any{"content": bs}}
	b, err := json.Marshal(rec)
	if err != nil {
		t.Fatal(err)
	}
	return string(b) + "\n"
}

func write(t *testing.T, dir, name, body string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestWrittenCollectsEditToolsOnly(t *testing.T) {
	proj := t.TempDir()
	tp := write(t, t.TempDir(), "t.jsonl",
		line(t, [2]string{"Write", filepath.Join(proj, "a.go")})+
			line(t, [2]string{"Edit", filepath.Join(proj, "pkg/b.go")})+
			line(t, [2]string{"NotebookEdit", filepath.Join(proj, "nb.ipynb")})+
			// THE TRAP: Read carries file_path too. Filtering on the FIELD rather than
			// the tool NAME would collect files the session merely looked at, and a
			// session that only read would look like a session that worked.
			line(t, [2]string{"Read", filepath.Join(proj, "read-only.go")})+
			line(t, [2]string{"Grep", filepath.Join(proj, "grepped.go")}))

	got := Written(tp, proj)
	want := []string{"a.go", "nb.ipynb", "pkg/b.go"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("Written = %v, want %v", got, want)
	}
}

func TestWrittenIsProjectRelativeAndBounded(t *testing.T) {
	proj := t.TempDir()
	outside := t.TempDir()
	tp := write(t, t.TempDir(), "t.jsonl",
		line(t, [2]string{"Write", filepath.Join(proj, "in.go")})+
			line(t, [2]string{"Write", filepath.Join(outside, "out.go")})+
			// The hooks' own state directory. Counting it would make every session
			// look like it touched something — the seal writes snapshots there.
			line(t, [2]string{"Write", filepath.Join(proj, ".claude", "checkpoints", "CHECKPOINT.md")})+
			line(t, [2]string{"Write", filepath.Join(proj, ".claude")}))

	got := Written(tp, proj)
	if len(got) != 1 || got[0] != "in.go" {
		t.Fatalf("Written = %v; want only the in-project, non-.claude path", got)
	}
	for _, p := range got {
		if filepath.IsAbs(p) || strings.HasPrefix(p, "..") {
			t.Errorf("path %q is not project-relative", p)
		}
	}
}

func TestWrittenDeduplicatesAndSorts(t *testing.T) {
	proj := t.TempDir()
	tp := write(t, t.TempDir(), "t.jsonl",
		line(t, [2]string{"Edit", filepath.Join(proj, "z.go")})+
			line(t, [2]string{"Edit", filepath.Join(proj, "a.go")})+
			line(t, [2]string{"Edit", filepath.Join(proj, "z.go")}))

	got := Written(tp, proj)
	if strings.Join(got, ",") != "a.go,z.go" {
		t.Fatalf("Written = %v; want deduplicated and sorted", got)
	}
}

// A transcript line carries whole tool results and routinely exceeds bufio.Scanner's 64KB
// token ceiling, where Scanner stops with ErrTooLong and the reader silently sees a
// TRUNCATED session — reporting "wrote nothing" for a session that wrote plenty. That is
// the flattering failure, so it gets a test with a line well past the limit.
func TestWrittenSurvivesHugeLines(t *testing.T) {
	proj := t.TempDir()
	huge := map[string]any{
		"type": "user",
		"message": map[string]any{"content": []map[string]any{{
			"type": "tool_result", "content": strings.Repeat("x", 300_000),
		}}},
	}
	b, err := json.Marshal(huge)
	if err != nil {
		t.Fatal(err)
	}
	tp := write(t, t.TempDir(), "t.jsonl",
		string(b)+"\n"+line(t, [2]string{"Write", filepath.Join(proj, "after-the-big-one.go")}))

	got := Written(tp, proj)
	if len(got) != 1 || got[0] != "after-the-big-one.go" {
		t.Fatalf("Written = %v; a >64KB line must not truncate the session", got)
	}
}

// Every degradation path yields less output, never a failure: this feeds a diagnostic at a
// seam, and a diagnostic that can fail the seal costs more than it explains.
func TestWrittenDegradesQuietly(t *testing.T) {
	proj := t.TempDir()
	cases := map[string]string{
		"missing file":        filepath.Join(t.TempDir(), "nope.jsonl"),
		"empty transcript":    write(t, t.TempDir(), "empty.jsonl", ""),
		"not json at all":     write(t, t.TempDir(), "junk.jsonl", "this is not json\n"),
		"content is a string": write(t, t.TempDir(), "str.jsonl", `{"message":{"content":"plain text"}}`+"\n"),
		"no message field":    write(t, t.TempDir(), "nomsg.jsonl", `{"type":"summary"}`+"\n"),
		"tool_use without path": write(t, t.TempDir(), "nopath.jsonl",
			`{"message":{"content":[{"type":"tool_use","name":"Write","input":{}}]}}`+"\n"),
	}
	for name, tp := range cases {
		t.Run(name, func(t *testing.T) {
			if got := Written(tp, proj); len(got) != 0 {
				t.Errorf("Written = %v, want nothing", got)
			}
		})
	}
	if got := Written("", proj); got != nil {
		t.Errorf("no transcript path must yield nothing, got %v", got)
	}
	if got := Written("x.jsonl", ""); got != nil {
		t.Errorf("no project dir must yield nothing, got %v", got)
	}
}

// A malformed record in the middle must not discard the records after it.
func TestWrittenRecoversAfterABadRecord(t *testing.T) {
	proj := t.TempDir()
	tp := write(t, t.TempDir(), "t.jsonl",
		line(t, [2]string{"Write", filepath.Join(proj, "before.go")})+
			"{\"message\": }\n"+
			line(t, [2]string{"Write", filepath.Join(proj, "after.go")}))

	got := Written(tp, proj)
	joined := strings.Join(got, ",")
	if joined != "after.go,before.go" {
		t.Fatalf("Written = %v; a bad line must cost exactly that line — both neighbours must survive", got)
	}
}

// The two silences must be distinguishable. An empty result means "wrote nothing" OR "I
// could not read it", and a caller that conflates them reports a healthy session when its
// own reader is broken — silence in the flattering direction.
func TestReadSeparatesEmptyFromUnreadable(t *testing.T) {
	proj := t.TempDir()

	// Read successfully, found nothing to report: NOT a failure.
	quiet := write(t, t.TempDir(), "q.jsonl", line(t, [2]string{"Read", "/x/only-read.go"}))
	if paths, why := Read(quiet, proj); len(paths) != 0 || why != "" {
		t.Errorf("a readable transcript with no writes must report no reason: paths=%v why=%q", paths, why)
	}

	// The format moved: every record fails to parse. This is the case that must never
	// read as a healthy session.
	moved := write(t, t.TempDir(), "m.jsonl", "not json\nalso not json\n")
	paths, why := Read(moved, proj)
	if len(paths) != 0 {
		t.Errorf("paths = %v", paths)
	}
	if !strings.Contains(why, "format has moved") {
		t.Errorf("a wholly unparseable transcript must say so, got %q", why)
	}

	// A missing file is a different failure again, and still not silence.
	if _, why := Read(filepath.Join(t.TempDir(), "nope.jsonl"), proj); !strings.Contains(why, "could not be opened") {
		t.Errorf("a missing transcript must be named, got %q", why)
	}

	// An empty file held no records.
	if _, why := Read(write(t, t.TempDir(), "e.jsonl", ""), proj); !strings.Contains(why, "no records") {
		t.Errorf("an empty transcript must be named, got %q", why)
	}

	// ONE bad line among good ones is noise, not a format change.
	mixed := write(t, t.TempDir(), "mix.jsonl",
		line(t, [2]string{"Write", filepath.Join(proj, "a.go")})+"{oops\n")
	if paths, why := Read(mixed, proj); why != "" || len(paths) != 1 {
		t.Errorf("one bad line must not be reported as a format change: paths=%v why=%q", paths, why)
	}

	// Nothing offered is not a failure to report.
	if _, why := Read("", proj); why != "" {
		t.Errorf("no transcript path must produce no reason, got %q", why)
	}
}
