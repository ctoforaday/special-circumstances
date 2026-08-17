package toolchainnudge

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/buildid"
	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/runlive"
	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/toolchain"
)

// allMissing / allPresent stand in for a PATH probe so a test's verdict never
// depends on what happens to be installed on the machine running it.
func allMissing(tools []toolchain.Tool) []toolchain.Status {
	out := make([]toolchain.Status, 0, len(tools))
	for _, t := range tools {
		out = append(out, toolchain.Status{Tool: t, Found: false})
	}
	return out
}

func allPresent(tools []toolchain.Tool) []toolchain.Status {
	out := make([]toolchain.Status, 0, len(tools))
	for _, t := range tools {
		out = append(out, toolchain.Status{Tool: t, Found: true})
	}
	return out
}

const manifest = `{"tools":[
  {"name":"qlty","tier":"recommended","check_cmd":"qlty --version"},
  {"name":"jq","tier":"optional","check_cmd":"jq --version"}
]}`

// pluginRoot writes a requirements.json (body "" means: write no file at all).
func pluginRoot(t *testing.T, body string) string {
	t.Helper()
	dir := t.TempDir()
	if body != "" {
		if err := os.WriteFile(filepath.Join(dir, "requirements.json"), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

// projectWithMarker plants a live-run marker (body "" means: no marker).
func projectWithMarker(t *testing.T, body string) string {
	t.Helper()
	dir := t.TempDir()
	if body != "" {
		claude := filepath.Join(dir, ".claude")
		if err := os.MkdirAll(claude, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(claude, "run-live.json"), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func call(t *testing.T, root, project string, probe func([]toolchain.Tool) []toolchain.Status, args ...string) (string, int) {
	t.Helper()
	var out bytes.Buffer
	code := run(args, nil, &out, root, project, probe)
	return out.String(), code
}

const liveMarker = `{"runDir":"research/2026-07-18_x","pinnedPaths":["ideas/backlog.md"],"started":"2026-07-18T09:00:00Z"}`

func TestSessionStartNudges(t *testing.T) {
	cases := []struct {
		name     string
		manifest string
		marker   string
		probe    func([]toolchain.Tool) []toolchain.Status
		want     []string // substrings; nil = expect total silence
		wantLine int      // expected number of output lines
	}{
		{"healthy toolchain, no run: total silence", manifest, "", allPresent, nil, 0},
		{"missing recommended tool nudges once", manifest, "", allMissing, []string{"qlty", "doctor"}, 1},
		{"optional-only misses stay silent", `{"tools":[{"name":"jq","tier":"optional"}]}`, "", allMissing, nil, 0},
		{"live run is announced", manifest, liveMarker, allPresent, []string{"LIVE", "research/2026-07-18_x", "do NOT update plugins"}, 1},
		{"both nudges are separate lines", manifest, liveMarker, allMissing, []string{"qlty", "LIVE"}, 2},
		{"no plugin root still announces a live run", "", liveMarker, allMissing, []string{"LIVE"}, 1},
		{"malformed manifest degrades to silence, not error", `{"tools":`, "", allMissing, nil, 0},
		{"malformed manifest still announces a live run", `{"tools":`, liveMarker, allMissing, []string{"LIVE"}, 1},
		{"malformed marker is not a live run", manifest, `{"runDir":`, allPresent, nil, 0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			out, code := call(t, pluginRoot(t, c.manifest), projectWithMarker(t, c.marker), c.probe)
			if code != 0 {
				t.Fatalf("exit %d; a SessionStart hook must never fail the session", code)
			}
			if len(c.want) == 0 {
				if out != "" {
					t.Fatalf("want silence, got %q", out)
				}
				return
			}
			for _, w := range c.want {
				if !strings.Contains(out, w) {
					t.Errorf("output = %q; missing %q", out, w)
				}
			}
			if got := len(strings.Split(strings.TrimRight(out, "\n"), "\n")); got != c.wantLine {
				t.Errorf("want %d line(s), got %d: %q", c.wantLine, got, out)
			}
		})
	}
}

// An absent plugin root and an absent manifest are the same silence, and neither may
// be read relative to the working directory.
func TestNoPluginRootReadsNothingRelative(t *testing.T) {
	cwd := t.TempDir()
	if err := os.WriteFile(filepath.Join(cwd, "requirements.json"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(cwd)
	out, code := call(t, "", t.TempDir(), allMissing)
	if code != 0 || out != "" {
		t.Fatalf("empty plugin root must be silent, got %q (exit %d)", out, code)
	}
}

func TestLiveNudge(t *testing.T) {
	if got := liveNudge(nil); got != "" {
		t.Fatalf("no marker must be silent, got %q", got)
	}
	m := &runlive.Marker{RunDir: "research/x", Started: "2026-07-18T09:00:00Z"}
	got := liveNudge(m)
	for _, w := range []string{"research/x", "2026-07-18T09:00:00Z", "LIVE", "pinned paths"} {
		if !strings.Contains(got, w) {
			t.Errorf("liveNudge = %q; missing %q", got, w)
		}
	}
	if strings.Contains(got, "\n") {
		t.Errorf("the live nudge must be one line: %q", got)
	}
}

func TestVersionFlag(t *testing.T) {
	out, code := call(t, pluginRoot(t, manifest), projectWithMarker(t, liveMarker), allMissing, "-version")
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	if !strings.Contains(out, "sc-toolchain-nudge") || !strings.Contains(out, buildid.Revision()) {
		t.Fatalf("version output = %q", out)
	}
	if strings.Contains(out, "LIVE") || strings.Contains(out, "qlty") {
		t.Errorf("-version must print only the version: %q", out)
	}
}
