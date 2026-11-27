package telecli

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/gray-area/tools/internal/peer"
)

// provisionEnv is an Env whose configuration and agy install live in temporary directories.
func provisionEnv(t *testing.T) (*harness, string) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("the daemon's token is provisioned on Unix only")
	}
	h := newColdHarness(t)
	cfg := t.TempDir()
	h.env.GOOS = "linux"
	h.env.TokenFile = peer.DaemonTokenFile(cfg)
	h.env.DropInDir = filepath.Join(cfg, "systemd", "user", peer.DaemonUnit+".d")
	h.env.AgyTelepathy = filepath.Join(t.TempDir(), "bin", "telepathy")
	if err := os.MkdirAll(filepath.Dir(h.env.AgyTelepathy), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(h.env.AgyTelepathy, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	h.env.RunProgram = func(string, ...string) error { return nil }
	return h, cfg
}

// AGY-PROVISION WRITES THE TOKEN AND THE DROP-IN THAT NAMES IT, prints the restart for the human,
// and never prints the token.
func TestProvisionWritesBothFiles(t *testing.T) {
	h, _ := provisionEnv(t)
	out, errOut, code := h.run(t, "agy-provision")
	if code != 0 {
		t.Fatalf("exit %d: %s", code, errOut)
	}
	tok, err := peer.ReadDaemonToken(h.env.TokenFile)
	if err != nil {
		t.Fatalf("the token file does not read back: %v", err)
	}
	if strings.Contains(out+errOut, tok) {
		t.Errorf("agy-provision printed the token")
	}
	dropIn := filepath.Join(h.env.DropInDir, "telepathy.conf")
	b, err := os.ReadFile(dropIn)
	if err != nil {
		t.Fatal(err)
	}
	want := "[Service]\nEnvironmentFile=-" + h.env.TokenFile + "\nExecStartPost=-" + h.env.AgyTelepathy + " agy-wake\n"
	if !strings.HasSuffix(string(b), want) {
		t.Errorf("drop-in =\n%s\nwant it to end\n%s", b, want)
	}
	if st, _ := os.Stat(dropIn); st.Mode().Perm() != 0o644 {
		t.Errorf("drop-in mode %04o, want 0644", st.Mode().Perm())
	}
	for _, s := range []string{"the human runs:\n  systemctl --user daemon-reload && systemctl --user restart antigravity-cli-daemon.service", "ends every turn running on the daemon"} {
		if !strings.Contains(out, s) {
			t.Errorf("stdout lacks %q:\n%s", s, out)
		}
	}
	if strings.Contains(out, "WARNING") {
		t.Errorf("a warning with agy's install in place:\n%s", out)
	}
	if _, _, code := h.run(t, "agy-provision"); code != 0 {
		t.Fatal("a second run failed")
	}
	if again, _ := peer.ReadDaemonToken(h.env.TokenFile); again == tok {
		t.Errorf("a second run did not rotate the token")
	}
}

func TestProvisionWarnsWhenAgyInstallCannotWake(t *testing.T) {
	h, _ := provisionEnv(t)
	h.env.RunProgram = func(string, ...string) error { return errors.New("exit status 2") }
	out, _, code := h.run(t, "agy-provision")
	if code != 0 || !strings.Contains(out, "WARNING: "+filepath.ToSlash(h.env.AgyTelepathy)+" does not know agy-wake") {
		t.Errorf("exit %d, stdout:\n%s", code, out)
	}
	if err := os.Remove(h.env.AgyTelepathy); err != nil {
		t.Fatal(err)
	}
	out, _, _ = h.run(t, "agy-provision")
	if !strings.Contains(out, "WARNING: agy's install of telepathy is not at") {
		t.Errorf("stdout:\n%s", out)
	}
}

func TestProvisionRefusesOffLinuxAndUnsafePaths(t *testing.T) {
	h, _ := provisionEnv(t)
	h.env.GOOS = "darwin"
	if _, errOut, code := h.run(t, "agy-provision"); code != 1 || !strings.Contains(errOut, "Linux only") {
		t.Errorf("off Linux: exit %d, %s", code, errOut)
	}
	h.env.GOOS = "linux"
	h.env.AgyTelepathy = "/home/a b/telepathy"
	if _, errOut, code := h.run(t, "agy-provision"); code != 1 || !strings.Contains(errOut, "cannot be written into a systemd unit") {
		t.Errorf("a path with a space: exit %d, %s", code, errOut)
	}
	if _, err := os.Stat(h.env.TokenFile); err == nil {
		t.Errorf("a refused run wrote the token file")
	}
}

// wakeWorld is a daemon root with conversations whose mailboxes the test fills, a provisioned token,
// and a stand-in daemon that records the conversations it is sent wakes for.
type wakeWorld struct {
	*sendWorld
	woken <-chan map[string]string
}

func newWakeWorld(t *testing.T) wakeWorld {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("the daemon's token file is read on Unix only")
	}
	w := newSendWorld(t)
	_, got := w.provisioned(t, http.StatusOK)
	// A clock that only sleeping moves, so a wait is measured in the run's own time.
	clock := frozen
	w.env.Now = func() time.Time { return clock }
	w.env.Sleep = func(d time.Duration) { clock = clock.Add(d) }
	return wakeWorld{sendWorld: w, woken: got}
}

func (w wakeWorld) wokenSet() map[string]bool {
	m := map[string]bool{}
	for {
		select {
		case r := <-w.woken:
			m[r["recipient"]] = true
		default:
			return m
		}
	}
}

// AGY-WAKE WAKES ONLY FOR MAIL: a readable message from an agent or a shell. agy's notices, earlier
// wake messages and unreadable markers are no reason, because no turn is measured to clear them.
func TestAgyWakeOnlyRealMail(t *testing.T) {
	w := newWakeWorld(t)
	mail, wakeOnly, notice, torn, held := agyPeer, agySelf, "a9e00000-0000-4000-8000-0000000000aa", "a9e00000-0000-4000-8000-0000000000bb", "a9e00000-0000-4000-8000-0000000000cc"
	queueMessage(t, w.brain, mail, "m1", mailboxJSON("telepathy", "2026-09-09T09:00:00Z"))
	queueMessage(t, w.brain, mail, "m2", mailboxJSON(peer.WakeSender, "2026-09-09T09:01:00Z"))
	queueMessage(t, w.brain, wakeOnly, "w1", mailboxJSON(peer.WakeSender, "2026-09-09T09:00:00Z"))
	queueMessage(t, w.brain, notice, "n1", mailboxJSON(peer.SystemSender, "2026-09-09T09:00:00Z"))
	queueMessage(t, w.brain, torn, "t1", "{torn")
	queueMessage(t, w.brain, held, "h1", mailboxJSON("agent-x", "2026-09-09T09:00:00Z"))
	w.env.ConversationBusy = func(_, conv string, exclude int) peer.Holder {
		if exclude != newDaemon.PID {
			t.Errorf("the holder check excluded pid %d, not the daemon", exclude)
		}
		if conv == held {
			return dbHolder(4321)
		}
		return peer.Holder{}
	}
	_, errOut, code := w.run(t, "agy-wake")
	if code != 0 {
		t.Fatalf("exit %d: %s", code, errOut)
	}
	if got := w.wokenSet(); len(got) != 1 || !got[mail] {
		t.Errorf("woken %v, want only %s\n%s", got, mail, errOut)
	}
	for _, line := range []string{"woke " + mail + " (1 messages waiting)", "not woken " + held + ": held by pid 4321"} {
		if !strings.Contains(errOut, line) {
			t.Errorf("stderr lacks %q:\n%s", line, errOut)
		}
	}
	if strings.Contains(errOut, testToken) {
		t.Errorf("agy-wake printed the token")
	}
}

// AGY-WAKE NEVER FAILS THE DAEMON'S START: every soft failure exits 0 with one line saying why.
func TestAgyWakeExitsZeroOnSoftFailure(t *testing.T) {
	w := newWakeWorld(t)
	queueMessage(t, w.brain, agyPeer, "m1", mailboxJSON("telepathy", "2026-09-09T09:00:00Z"))
	for _, tc := range []struct {
		name string
		set  func()
		want string
	}{
		{"off Linux", func() { w.env.GOOS = "darwin" }, "not run: a wake is sent on Linux only"},
		{"no token", func() { w.env.TokenFile = filepath.Join(t.TempDir(), "none") }, "not run: no daemon token file at"},
		{"no daemon", func() {
			w.env.CurrentDaemon = func() (peer.ServerIdentity, bool) { return peer.ServerIdentity{}, false }
		}, "not run: no daemon is running"},
		{"no port", func() {
			w.env.DaemonAddress = func(*http.Client, int) (string, error) { return "", errors.New("no listener yet") }
		}, "not run: the daemon's HTTP port cannot be told: no listener yet"},
	} {
		saved := w.env
		tc.set()
		_, errOut, code := w.run(t, "agy-wake")
		if code != 0 || !strings.Contains(errOut, tc.want) {
			t.Errorf("%s: exit %d, stderr %q, want %q", tc.name, code, errOut, tc.want)
		}
		w.env = saved
	}
	if got := w.wokenSet(); len(got) != 0 {
		t.Errorf("a failed run sent wakes: %v", got)
	}
}

// AGY-WAKE WAITS FOR THE PORT, within its bound: a daemon whose port appears on the third look is
// woken through.
func TestAgyWakeWaitsForThePort(t *testing.T) {
	w := newWakeWorld(t)
	queueMessage(t, w.brain, agyPeer, "m1", mailboxJSON("telepathy", "2026-09-09T09:00:00Z"))
	real := w.env.DaemonAddress
	looks := 0
	w.env.DaemonAddress = func(c *http.Client, pid int) (string, error) {
		if looks++; looks < 3 {
			return "", errors.New("not listening yet")
		}
		return real(c, pid)
	}
	if _, errOut, _ := w.run(t, "agy-wake"); !strings.Contains(errOut, "woke "+agyPeer) || looks != 3 {
		t.Errorf("looks %d, stderr:\n%s", looks, errOut)
	}
}

// THE WHOLE RUN ENDS WITHIN ITS BOUND, whatever the daemon does: systemd stops the daemon when a
// start command outlives the start timeout. A port that never appears ends the wait at 45 s; a
// conversation reached after the bound is named and not woken.
func TestAgyWakeBound(t *testing.T) {
	w := newWakeWorld(t)
	queueMessage(t, w.brain, agyPeer, "m1", mailboxJSON("telepathy", "2026-09-09T09:00:00Z"))
	queueMessage(t, w.brain, agySelf, "m2", mailboxJSON("telepathy", "2026-09-09T09:00:00Z"))
	clock := time.Date(2026, 9, 9, 12, 0, 0, 0, time.UTC)
	start := clock
	w.env.Now = func() time.Time { return clock }
	w.env.Sleep = func(d time.Duration) { clock = clock.Add(d) }
	saved := w.env.DaemonAddress
	w.env.DaemonAddress = func(*http.Client, int) (string, error) { return "", errors.New("never listening") }
	_, errOut, code := w.run(t, "agy-wake")
	if code != 0 || !strings.Contains(errOut, "not run: the daemon's HTTP port cannot be told: never listening") || clock.Sub(start) > portWait {
		t.Errorf("exit %d after %s, stderr:\n%s", code, clock.Sub(start), errOut)
	}
	// The port answers only once the whole bound is spent: no conversation may be woken.
	clock = start
	w.env.DaemonAddress = func(c *http.Client, pid int) (string, error) {
		clock = clock.Add(61 * time.Second)
		return saved(c, pid)
	}
	_, errOut, _ = w.run(t, "agy-wake")
	if strings.Contains(errOut, "woke ") || strings.Count(errOut, "the 1m0s bound ran out") != 2 {
		t.Errorf("stderr:\n%s", errOut)
	}
	if got := w.wokenSet(); len(got) != 0 {
		t.Errorf("wakes sent past the bound: %v", got)
	}
}
