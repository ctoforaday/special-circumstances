package telecli

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/gray-area/tools/internal/peer"
)

var mailboxID = regexp.MustCompile(`[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}\.json`)

// writeTranscript gives conv a transcript of the given step lines.
func (w *sendWorld) writeTranscript(t *testing.T, conv string, lines ...string) string {
	t.Helper()
	path := filepath.Join(w.brain, conv, ".system_generated", "logs", "transcript.jsonl")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	body := strings.Join(lines, "\n")
	if len(lines) > 0 {
		body += "\n"
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

// EVERY QUEUED MESSAGE SAYS WHETHER IT WAS WOKEN; one that was not states why and the two times
// that bound the wait. One golden per cause, rendered against an injected platform, boot, daemon,
// port and holder so it is the same on every host.
func TestSendQueuedCauseGoldens(t *testing.T) {
	for _, tc := range []struct {
		cause string
		setup func(t *testing.T, w *sendWorld, names map[string]string)
	}{
		{"linux-only", func(t *testing.T, w *sendWorld, _ map[string]string) {
			w.provisioned(t, http.StatusOK)
			w.env.GOOS = "darwin"
		}},
		{"not-daemon-root", func(t *testing.T, w *sendWorld, names map[string]string) {
			w.provisioned(t, http.StatusOK)
			w.env.DaemonRoot = t.TempDir()
			names[w.env.DaemonRoot] = "<DAEMON-ROOT>"
		}},
		{"no-token-file", func(t *testing.T, w *sendWorld, _ map[string]string) {}},
		{"token-file-refused", func(t *testing.T, w *sendWorld, _ map[string]string) {
			w.provisioned(t, http.StatusOK)
			if err := os.Chmod(w.tokenFile, 0o640); err != nil {
				t.Fatal(err)
			}
		}},
		{"no-daemon", func(t *testing.T, w *sendWorld, _ map[string]string) {
			w.provisioned(t, http.StatusOK)
			w.env.CurrentDaemon = func() (peer.ServerIdentity, bool) { return peer.ServerIdentity{}, false }
		}},
		{"port-not-found", func(t *testing.T, w *sendWorld, _ map[string]string) {
			w.provisioned(t, http.StatusOK)
			w.env.DaemonAddress = func(*http.Client, int) (string, error) {
				return "", errors.New("0 of pid 284285's listening ports answer as its HTTP port (127.0.0.1:43795: HTTP 400)")
			}
		}},
		{"second-writer", func(t *testing.T, w *sendWorld, _ map[string]string) {
			w.provisioned(t, http.StatusOK)
			w.env.ConversationBusy = func(string, string, int) peer.Holder { return dbHolder(9999) }
		}},
		{"uninspectable", func(t *testing.T, w *sendWorld, _ map[string]string) {
			w.provisioned(t, http.StatusOK)
			w.env.ConversationBusy = func(string, string, int) peer.Holder {
				return peer.Holder{Kind: peer.HolderUninspectable, Detail: "cannot read /proc to verify no second writer: permission denied"}
			}
		}},
		{"wake-failed", func(t *testing.T, w *sendWorld, names map[string]string) {
			addr, _ := w.provisioned(t, http.StatusUnauthorized)
			names[addr] = "<DAEMON-ADDR>"
		}},
		{"woken", func(t *testing.T, w *sendWorld, names map[string]string) {
			addr, _ := w.provisioned(t, http.StatusOK)
			names[addr] = "<DAEMON-ADDR>"
		}},
	} {
		t.Run(tc.cause, func(t *testing.T) {
			if runtime.GOOS == "windows" {
				t.Skip("the daemon's token file is read on Unix only")
			}
			w := newSendWorld(t)
			w.writeTranscript(t, agyPeer,
				`{"step_index":0,"type":"USER_INPUT","created_at":"2026-09-09T11:41:00Z"}`,
				`{"step_index":1,"type":"PLANNER_RESPONSE","created_at":"2026-09-09T11:52:30Z"}`)
			names := map[string]string{w.brain: "<BRAIN>", filepath.Dir(w.brain): "<AGY-ROOT>", filepath.Dir(filepath.Dir(w.tokenFile)): "<CONFIG>"}
			tc.setup(t, w, names)
			out := w.sendFromClaude(t)
			// Longest first, so a path is replaced before a directory that contains it.
			froms := make([]string, 0, len(names))
			for from := range names {
				froms = append(froms, from)
			}
			sort.Slice(froms, func(i, j int) bool { return len(froms[i]) > len(froms[j]) })
			for _, from := range froms {
				out = strings.ReplaceAll(out, filepath.ToSlash(from), names[from])
			}
			out = mailboxID.ReplaceAllString(out, "<MESSAGE-ID>.json")
			assertGolden(t, "send-queued-"+tc.cause, out)
		})
	}
}

func TestQueuedLastStepTime(t *testing.T) {
	w := newSendWorld(t)
	for _, tc := range []struct {
		name  string
		lines []string
		want  string
	}{
		{"three lines", []string{
			`{"step_index":0,"created_at":"2026-09-09T11:00:00Z"}`,
			`{"step_index":1,"created_at":"2026-09-09T11:10:00Z"}`,
			`{"step_index":2,"created_at":"2026-09-09T11:20:00.5Z","content":"` + strings.Repeat("x", 200<<10) + `"}`,
		}, "last step recorded at 2026-09-09 11:20:00 UTC"},
		{"one line", []string{`{"step_index":0,"created_at":"2026-09-09T09:30:00Z"}`}, "last step recorded at 2026-09-09 09:30:00 UTC"},
		{"no lines", nil, "no step recorded"},
		{"a last line that is not a step", []string{`{"step_index":0,"created_at":"2026-09-09T09:30:00Z"}`, `{torn`}, "last step not measured: "},
	} {
		w.writeTranscript(t, agyPeer, tc.lines...)
		if got := lastStepLine(filepath.Join(w.brain, agyPeer)); !strings.HasPrefix(got, tc.want) {
			t.Errorf("%s: lastStepLine = %q, want %q", tc.name, got, tc.want)
		}
	}
	if got := lastStepLine(filepath.Join(w.brain, agySelf)); got != "no step recorded" {
		t.Errorf("missing transcript: lastStepLine = %q, want no step recorded", got)
	}
}

func TestQueuedDaemonLine(t *testing.T) {
	boot := func() (int64, bool) { return bootAt.Unix(), true }
	for _, tc := range []struct {
		name          string
		looked, found bool
		start         int64
		boot          func() (int64, bool)
		want          string
	}{
		{"measured", true, true, 30000, boot, "daemon started 2026-09-09 11:45:00 UTC"},
		{"no daemon", true, false, 0, boot, "no daemon is running"},
		{"never looked for", false, false, 0, boot, "daemon start not measured"},
		{"no boot", true, true, 30000, func() (int64, bool) { return 0, false }, "daemon start not measured"},
		{"nil boot", true, true, 30000, nil, "daemon start not measured"},
		{"no start time", true, true, 0, boot, "daemon start not measured"},
	} {
		w := wakeDaemon{env: &Env{Boot: tc.boot}, looked: tc.looked, found: tc.found, id: peer.ServerIdentity{PID: 1, StartTime: tc.start}}
		if got := w.daemonLine(); got != tc.want {
			t.Errorf("%s: daemonLine = %q, want %q", tc.name, got, tc.want)
		}
	}
}
