package telecli

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/gray-area/tools/internal/catalogue"
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

// EVERY QUEUED LINE STATES ITS CAUSE AND THE TWO TIMES THAT BOUND THE WAIT, one golden per cause,
// rendered against an injected platform, boot, daemon and holder so it is the same on every host.
func TestSendQueuedCauseGoldens(t *testing.T) {
	for _, tc := range []struct {
		cause string
		setup func(t *testing.T, w *sendWorld, names map[string]string)
	}{
		{"no-runtime-dir", func(t *testing.T, w *sendWorld, _ map[string]string) { w.env.RuntimeDir = "" }},
		{"linux-only", func(t *testing.T, w *sendWorld, _ map[string]string) { w.env.GOOS = "darwin" }},
		{"records-unreadable", func(t *testing.T, w *sendWorld, _ map[string]string) {
			if err := os.MkdirAll(filepath.Join(w.runtime, "gray-area"), 0o700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(peer.RegistryDir(w.runtime), nil, 0o600); err != nil {
				t.Fatal(err)
			}
		}},
		{"no-daemon", func(t *testing.T, w *sendWorld, _ map[string]string) {
			w.env.CurrentDaemon = func() (peer.ServerIdentity, bool) { return peer.ServerIdentity{}, false }
		}},
		{"not-daemon-root", func(t *testing.T, w *sendWorld, names map[string]string) {
			w.env.DaemonRoot = t.TempDir()
			names[w.env.DaemonRoot] = "<DAEMON-ROOT>"
		}},
		{"daemon-record-unreadable", func(t *testing.T, w *sendWorld, _ map[string]string) {
			dir := peer.RegistryDir(w.runtime)
			if err := os.MkdirAll(dir, 0o700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(dir, peer.ServerKey(newDaemon.PID, newDaemon.StartTime, "")+".json"), []byte("{torn"), 0o600); err != nil {
				t.Fatal(err)
			}
		}},
		{"daemon-not-linked", func(t *testing.T, w *sendWorld, _ map[string]string) {}},
		{"daemon-record-unknown", func(t *testing.T, w *sendWorld, _ map[string]string) {
			w.linkAs(t, agySelf, "localhost:7000", newDaemon, frozen)
			w.env.RecordLiveness = func(peer.Server) catalogue.Liveness { return catalogue.Unknown }
		}},
		{"daemon-unreached", func(t *testing.T, w *sendWorld, names map[string]string) {
			addr := unreachable(t)
			names[addr] = "<DAEMON-ADDR>"
			w.linkAs(t, agySelf, addr, newDaemon, frozen)
		}},
		{"holder-unreached", func(t *testing.T, w *sendWorld, names map[string]string) {
			addr := unreachable(t)
			names[addr] = "<S1-ADDR>"
			w.linkAs(t, agySelf, addr, serverS1, frozen)
			w.holders(map[int]peer.Holder{0: dbHolder(serverS1.PID)})
		}},
		{"holder-took-it", func(t *testing.T, w *sendWorld, names map[string]string) {
			addr := unreachable(t)
			names[addr] = "<S1-ADDR>"
			w.linkAs(t, agyPeer, addr, serverS1, frozen)
			w.linkAs(t, agySelf, "localhost:7000", newDaemon, frozen)
			w.holders(map[int]peer.Holder{newDaemon.PID: dbHolder(serverS1.PID)})
		}},
		{"rate-limited", func(t *testing.T, w *sendWorld, names map[string]string) {
			addr, _ := rateLimited(t, 99, "")
			names[addr] = "<S1-ADDR>"
			w.linkAs(t, agyPeer, addr, serverS1, frozen)
			w.env.Sleep = func(time.Duration) {}
		}},
		{"two-holders", func(t *testing.T, w *sendWorld, _ map[string]string) {
			w.linkAs(t, agyPeer, "localhost:7001", serverS1, frozen)
			w.linkAs(t, agySelf, "localhost:7002", serverS2, frozen)
			w.holders(map[int]peer.Holder{0: dbHolder(serverS1.PID), serverS1.PID: dbHolder(serverS2.PID)})
		}},
		{"second-writer", func(t *testing.T, w *sendWorld, _ map[string]string) {
			w.holders(map[int]peer.Holder{0: dbHolder(9999)})
		}},
		{"uninspectable", func(t *testing.T, w *sendWorld, _ map[string]string) {
			w.holders(map[int]peer.Holder{0: {Kind: peer.HolderUninspectable, Detail: "cannot read /proc to verify no second writer: permission denied"}})
		}},
	} {
		t.Run(tc.cause, func(t *testing.T) {
			if (tc.cause == "records-unreadable" || strings.HasSuffix(tc.cause, "-unreached")) && runtime.GOOS == "windows" {
				t.Skip("the golden carries the operating system's own error text, which is Unix's here")
			}
			w := newSendWorld(t)
			w.env.CurrentDaemon = func() (peer.ServerIdentity, bool) { return newDaemon, true }
			w.writeTranscript(t, agyPeer,
				`{"step_index":0,"type":"USER_INPUT","created_at":"2026-09-09T11:41:00Z"}`,
				`{"step_index":1,"type":"PLANNER_RESPONSE","created_at":"2026-09-09T11:52:30Z"}`)
			names := map[string]string{w.brain: "<BRAIN>", filepath.Dir(w.brain): "<AGY-ROOT>", w.runtime: "<RUNTIME>"}
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
		name  string
		goos  string
		has   bool
		start int64
		boot  func() (int64, bool)
		want  string
	}{
		{"measured", "linux", true, 30000, boot, "daemon started 2026-09-09 11:45:00 UTC"},
		{"no daemon", "linux", false, 0, boot, "no daemon is running"},
		{"not looked for off Linux", "darwin", false, 0, boot, "daemon start not measured"},
		{"no boot", "linux", true, 30000, func() (int64, bool) { return 0, false }, "daemon start not measured"},
		{"nil boot", "linux", true, 30000, nil, "daemon start not measured"},
		{"no start time", "linux", true, 0, boot, "daemon start not measured"},
	} {
		r := &agyRouter{env: &Env{Boot: tc.boot, GOOS: tc.goos}, hasDaemon: tc.has, daemonID: peer.ServerIdentity{PID: 1, StartTime: tc.start}}
		if got := r.daemonLine(); got != tc.want {
			t.Errorf("%s: daemonLine = %q, want %q", tc.name, got, tc.want)
		}
	}
}
