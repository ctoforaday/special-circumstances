package main

import (
	"encoding/json"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/gray-area/tools/internal/peer"
	"github.com/ctoforaday/special-circumstances/shared/hookcore"
)

// linkRequest is the one line an Antigravity turn is given when the conversation cannot be woken by
// `telepathy send`, and "" when it can or when asking would be useless.
//
// WHY A REQUEST AND NOT AN ACT. The language server's token reaches an agent's shell commands and
// nothing else — this hook gets no ANTIGRAVITY_ variable but the conversation id — and a hook can
// inject messages but never a tool call. So the hook can only ask, and `telepathy link` is the
// command it asks for (peer.Register says why that is the recording point).
//
// WHEN IT ASKS. PreInvocation fires before every model step of a turn (invocationNum 0, 1, 2 …),
// so it asks on the turn's first step only: an ephemeral message is dropped after the turn, and
// repeating it every step would cost a step's attention for nothing. It asks only when there is no
// record for this conversation, or the recorded server no longer accepts a connection — a restarted
// server listens on a new port. It never asks where linking would fail: no runtime directory, or no
// telepathy binary beside this one.
func linkRequest(req hookcore.Request, runtimeDir, telepathy string, dial func(addr string) bool) string {
	if req.Engine != hookcore.EngineAntigravity || req.SessionID == "" || runtimeDir == "" || telepathy == "" {
		return ""
	}
	var p struct {
		InvocationNum int `json:"invocationNum"`
	}
	_ = json.Unmarshal(req.Raw, &p)
	if p.InvocationNum != 0 {
		return ""
	}
	s, err := peer.LookupServer(runtimeDir, req.SessionID)
	if err == nil && s != nil && dial(s.Address) {
		return ""
	}
	// A STALE LINK IS SAID TO BE STALE. Measured: a conversation whose history shows `linked` from an
	// earlier turn ignored the plain request on the next one, after its server had restarted. The
	// request has to say why the earlier link no longer counts.
	why := "Other agents cannot yet wake this conversation with `telepathy send`."
	if err == nil && s != nil {
		why = "This conversation's earlier `telepathy link` points at a language server that is no longer running, so other agents cannot wake it."
	}
	return why + " Before your task, run this shell command once: " + shellQuote(telepathy) + " link — it prints `linked` and needs nothing further."
}

// findTelepathy locates the binary linkRequest names; a variable so a test can stand one in.
var findTelepathy = telepathyBeside

// telepathyBeside is the telepathy binary installed next to this hook's own, or "".
func telepathyBeside() string {
	exe, err := os.Executable()
	if err != nil {
		return ""
	}
	name := "telepathy"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	p := filepath.Join(filepath.Dir(exe), name)
	if st, err := os.Stat(p); err != nil || st.IsDir() {
		return ""
	}
	return p
}

// answers reports whether a server accepts a connection at addr.
func answers(addr string) bool {
	c, err := net.DialTimeout("tcp", addr, 300*time.Millisecond)
	if err != nil {
		return false
	}
	c.Close()
	return true
}

func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}
