package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/gray-area/tools/internal/catalogue"
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
// repeating it every step would cost a step's attention for nothing. It asks only when neither
// this server nor this conversation has a server record that liveness reads Live — a restarted
// server has a new PID/start time, and telepathy deletes the records of servers that are gone. It
// never asks where linking would fail: no runtime directory, or no telepathy binary beside this
// one.
func linkRequest(req hookcore.Request, runtimeDir, telepathy string, liveness func(peer.Server) catalogue.Liveness) string {
	return linkRequestWithHost(req, runtimeDir, telepathy, peer.HostServer, liveness)
}

func linkRequestWithHost(req hookcore.Request, runtimeDir, telepathy string, hostFn func() peer.ServerIdentity, liveness func(peer.Server) catalogue.Liveness) string {
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
	if hostFn != nil {
		host := hostFn()
		if host.PID > 0 && host.StartTime > 0 {
			if hs, err := peer.LookupServerByKey(runtimeDir, peer.ServerKey(host.PID, host.StartTime, "")); err == nil && hs != nil && liveness(*hs) == catalogue.Live {
				return ""
			}
		}
	}
	if s, err := peer.LookupConversationServer(runtimeDir, req.SessionID, liveness); err == nil && s != nil {
		return ""
	}
	// A STALE LINK IS SAID TO BE STALE. Measured: a conversation whose history shows `linked` from an
	// earlier turn ignored the plain request on the next one, after its server had restarted. The
	// record that would say whether there was an earlier link is deleted once its server is gone, so
	// the request names the stale link conditionally, which is true with or without one.
	return "Any earlier `telepathy link` in this conversation points at a language server that is no longer running, so other agents cannot wake it." +
		" Before your task, run this shell command once: " + shellQuote(telepathy) + " link — it prints `linked` and needs nothing further."
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

func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}
