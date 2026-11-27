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
// one; nor where no record could ever read Live: a host with no pid domain (no /etc/machine-id,
// or not Linux), where asking would repeat on every turn for a record send never uses.
//
// Stated residue: it does not ask while the conversation's own record reads Live, even when the
// server now hosting it is a different, unrecorded one (a standalone agy while the daemon hosts
// the conversation); sends then queue with the unlinked daemon named.
func linkRequest(req hookcore.Request, runtimeDir, telepathy string, liveness func(peer.Server) catalogue.Liveness) string {
	return linkRequestWithHost(req, runtimeDir, telepathy, peer.HostServer, liveness)
}

func linkRequestWithHost(req hookcore.Request, runtimeDir, telepathy string, hostFn func() peer.ServerIdentity, liveness func(peer.Server) catalogue.Liveness) string {
	if req.Engine != hookcore.EngineAntigravity || req.SessionID == "" || runtimeDir == "" || telepathy == "" || localPidDomain() == "" {
		return ""
	}
	var p struct {
		InvocationNum int `json:"invocationNum"`
	}
	_ = json.Unmarshal(req.Raw, &p)
	if p.InvocationNum != 0 {
		return ""
	}
	var host peer.ServerIdentity
	if hostFn != nil {
		host = hostFn()
		if host.PID > 0 && host.StartTime > 0 {
			if hs, err := peer.LookupServerByKey(runtimeDir, peer.ServerKey(host.PID, host.StartTime, "")); err == nil && hs != nil && liveness(*hs) == catalogue.Live {
				return ""
			}
		}
	}
	if s, err := peer.LookupConversationServer(runtimeDir, req.SessionID, liveness); err == nil && s != nil {
		return ""
	}
	// NO SERVER PROCESS TO NAME, NONE TO ASK FOR. When this hook finds no Antigravity process above
	// it and the conversation's record names none either, linking again writes the same record,
	// which send cannot use (link says so): asking every turn would repeat that, in a wording that
	// would call a running server gone.
	if host.PID <= 0 && hasPidlessRecord(runtimeDir, req.SessionID) {
		return ""
	}
	// A STALE LINK IS SAID TO BE STALE, AND ONLY A STALE ONE. Measured: a conversation whose history
	// shows `linked` from an earlier turn ignored the plain request on the next one, after its server
	// had restarted; and a fresh conversation ignored the stale wording on each of three turns while
	// it acted on the plain one. Prune deletes a gone server's record, so the evidence of an earlier
	// link is the conversation's marker in peer.LinkedDir, which prune leaves. The marker lives on
	// tmpfs: after a reboot a conversation that had linked is asked in the plain wording.
	why := firstLinkWhy
	if peer.HasLinked(runtimeDir, req.SessionID) {
		why = staleLinkWhy
	}
	return why + " Before your task, run this shell command once: " + shellQuote(telepathy) + " link — it prints `linked` and needs nothing further."
}

// hasPidlessRecord reports whether a record listing conversation names no server process.
func hasPidlessRecord(runtimeDir, conversation string) bool {
	servers, err := peer.ListServers(runtimeDir)
	if err != nil {
		return false
	}
	for _, s := range servers {
		if s.PID <= 0 && s.Lists(conversation) {
			return true
		}
	}
	return false
}

// The two openings of the link request: for a conversation that has never linked, and for one
// whose earlier link points at a server that is gone.
const (
	firstLinkWhy = "Other agents cannot yet wake this conversation with `telepathy send`."
	staleLinkWhy = "This conversation's earlier `telepathy link` points at a language server that is no longer running, so other agents cannot wake it."
)

// localPidDomain is the pid domain a record written on this host carries; a variable so a test can
// stand in a host.
var localPidDomain = catalogue.LocalPidDomain

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
