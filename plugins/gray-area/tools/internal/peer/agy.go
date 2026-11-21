package peer

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"time"
)

// The three variables Antigravity puts in the environment of an agent's shell commands.
const (
	EnvConversation = "ANTIGRAVITY_CONVERSATION_ID"
	EnvAddress      = "ANTIGRAVITY_LS_ADDRESS"
	EnvToken        = "ANTIGRAVITY_CSRF_TOKEN"
)

// ServerIdentity identifies one Antigravity language-server process launch.
// On Linux, PID + StartTime (field 22 of /proc/<pid>/stat, clock ticks since boot)
// uniquely identifies a server launch even if a PID is reused after a restart.
type ServerIdentity struct {
	PID       int   `json:"pid,omitempty"`
	StartTime int64 `json:"start_time,omitempty"`
	Daemon    bool  `json:"daemon,omitempty"`
}

// Server is one Antigravity language server launch, stored keyed by server in the runtime
// directory so any conversation that links on a restarted daemon makes every idle conversation
// on that daemon wakeable.
type Server struct {
	Key           string   `json:"key"`
	PID           int      `json:"pid,omitempty"`
	StartTime     int64    `json:"start_time,omitempty"`
	Address       string   `json:"address"`
	Token         string   `json:"token"`
	Daemon        bool     `json:"daemon,omitempty"`
	Conversations []string `json:"conversations,omitempty"`
	RegisteredAt  int64    `json:"registered_at"` // when this server launch was first recorded
}

// ServerKey returns the filename-safe registry key for a language server launch.
// When PID and StartTime are known, it keys on the process launch; otherwise it falls back to
// a sanitised form of the listen address.
func ServerKey(pid int, startTime int64, address string) string {
	if pid > 0 && startTime > 0 {
		return fmt.Sprintf("pid-%d-%d", pid, startTime)
	}
	var b strings.Builder
	b.WriteString("addr-")
	for _, r := range address {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '.' || r == '-' {
			b.WriteRune(r)
		} else {
			b.WriteByte('_')
		}
	}
	return b.String()
}

// RegistryDir is where server records live under a runtime directory ($XDG_RUNTIME_DIR): a
// per-user tmpfs, so the token never reaches a disk and is gone at logout. The directory is 0700
// and each record 0600 — the token controls the whole language server, so it is held exactly as
// tightly as the user's own session.
func RegistryDir(runtimeDir string) string {
	if runtimeDir == "" {
		return ""
	}
	return filepath.Join(runtimeDir, "gray-area", "antigravity-servers")
}

// Linked is what Register found.
type Linked int

const (
	// NotAnAntigravityShell: the process's environment carries no server to record.
	NotAnAntigravityShell Linked = iota
	// NoRuntimeDir: there is nowhere private to keep the record ($XDG_RUNTIME_DIR unset).
	NoRuntimeDir
	// AlreadyLinked: the record already names this server launch and conversation; nothing was written.
	AlreadyLinked
	// Recorded: the record was written.
	Recorded
)

// RegisterWith records the server of the Antigravity shell this process runs in, keyed by server,
// using host to identify the server process launch.
func RegisterWith(runtimeDir string, getenv func(string) string, host func() ServerIdentity, now time.Time) (Server, Linked, error) {
	conv := getenv(EnvConversation)
	addr := getenv(EnvAddress)
	tok := getenv(EnvToken)
	if conv == "" || addr == "" || tok == "" {
		return Server{}, NotAnAntigravityShell, nil
	}
	dir := RegistryDir(runtimeDir)
	if dir == "" {
		return Server{}, NoRuntimeDir, nil
	}
	if !safeName(conv) {
		return Server{}, NotAnAntigravityShell, fmt.Errorf("%s %q is not a conversation id", EnvConversation, conv)
	}
	var id ServerIdentity
	if host != nil {
		id = host()
	}
	key := ServerKey(id.PID, id.StartTime, addr)
	if !safeName(key) {
		return Server{}, NotAnAntigravityShell, fmt.Errorf("server key %q is not a valid filename", key)
	}

	s := Server{
		Key:           key,
		PID:           id.PID,
		StartTime:     id.StartTime,
		Address:       addr,
		Token:         tok,
		Daemon:        id.Daemon,
		Conversations: []string{conv},
		RegisteredAt:  now.Unix(),
	}
	if prev, err := LookupServerByKey(runtimeDir, key); err == nil && prev != nil &&
		prev.Address == addr && prev.Token == tok {
		if slices.Contains(prev.Conversations, conv) {
			return *prev, AlreadyLinked, nil
		}
		s = *prev
		s.Conversations = append(slices.Clone(prev.Conversations), conv)
		if id.Daemon {
			s.Daemon = true
		}
	}

	if err := os.MkdirAll(dir, 0o700); err != nil {
		return Server{}, NotAnAntigravityShell, err
	}
	b, err := json.Marshal(s)
	if err != nil {
		return Server{}, NotAnAntigravityShell, err
	}
	tmp, err := os.CreateTemp(dir, ".register-*")
	if err != nil {
		return Server{}, NotAnAntigravityShell, err
	}
	defer os.Remove(tmp.Name())
	if err := tmp.Chmod(0o600); err != nil {
		tmp.Close()
		return Server{}, NotAnAntigravityShell, err
	}
	if _, err := tmp.Write(b); err != nil {
		tmp.Close()
		return Server{}, NotAnAntigravityShell, err
	}
	if err := tmp.Close(); err != nil {
		return Server{}, NotAnAntigravityShell, err
	}
	if err := os.Rename(tmp.Name(), filepath.Join(dir, key+".json")); err != nil {
		return Server{}, NotAnAntigravityShell, err
	}
	return s, Recorded, nil
}

// LookupServerByKey reads one server record by its registry key. A missing record is (nil, nil).
func LookupServerByKey(runtimeDir, key string) (*Server, error) {
	dir := RegistryDir(runtimeDir)
	if dir == "" || !safeName(key) {
		return nil, nil
	}
	b, err := os.ReadFile(filepath.Join(dir, key+".json"))
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var s Server
	if err := json.Unmarshal(b, &s); err != nil {
		return nil, fmt.Errorf("server record %s is unreadable: %w", key, err)
	}
	if s.Key == "" || s.Address == "" || s.Token == "" {
		return nil, fmt.Errorf("server record %s is incomplete", key)
	}
	return &s, nil
}

// LookupDaemonServer reads the server record for the daemon identified by id.
// Returns (nil, nil) when no conversation has linked on that daemon launch.
func LookupDaemonServer(runtimeDir string, id ServerIdentity) (*Server, error) {
	if runtimeDir == "" || id.PID <= 0 || id.StartTime <= 0 {
		return nil, nil
	}
	return LookupServerByKey(runtimeDir, ServerKey(id.PID, id.StartTime, ""))
}

// ListServers returns every valid server record in runtimeDir, newest RegisteredAt first.
func ListServers(runtimeDir string) ([]Server, error) {
	dir := RegistryDir(runtimeDir)
	if dir == "" {
		return nil, nil
	}
	entries, err := os.ReadDir(dir)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var out []Server
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") || strings.HasPrefix(e.Name(), ".") {
			continue
		}
		key := strings.TrimSuffix(e.Name(), ".json")
		s, err := LookupServerByKey(runtimeDir, key)
		if err != nil || s == nil {
			continue
		}
		out = append(out, *s)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].RegisteredAt != out[j].RegisteredAt {
			return out[i].RegisteredAt > out[j].RegisteredAt
		}
		return out[i].Key > out[j].Key
	})
	return out, nil
}

// LookupConversationServer returns the most recently registered server record that includes
// conversation. A missing record is (nil, nil).
func LookupConversationServer(runtimeDir, conversation string) (*Server, error) {
	if runtimeDir == "" || !safeName(conversation) {
		return nil, nil
	}
	servers, err := ListServers(runtimeDir)
	if err != nil {
		return nil, err
	}
	for i := range servers {
		if slices.Contains(servers[i].Conversations, conversation) {
			s := servers[i]
			return &s, nil
		}
	}
	return nil, nil
}

// AgyMessage is one message to an Antigravity conversation.
type AgyMessage struct {
	Recipient string
	Sender    string
	Title     string
	Content   string
}

// SendAgyServer delivers m through the language server s. It wakes an idle conversation and, on a
// restarted daemon, loads an unloaded conversation from disk and starts a turn.
func SendAgyServer(client *http.Client, s Server, m AgyMessage) (Delivery, error) {
	body, err := json.Marshal(map[string]string{
		"recipient": m.Recipient, "content": m.Content, "senderId": m.Sender, "displayTitle": m.Title,
	})
	if err != nil {
		return Delivery{}, err
	}
	req, err := http.NewRequest(http.MethodPost,
		"http://"+s.Address+"/exa.language_server_pb.LanguageServerService/SendAgentMessage", bytes.NewReader(body))
	if err != nil {
		return Delivery{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-codeium-csrf-token", s.Token)
	resp, err := client.Do(req)
	if err != nil {
		return Delivery{}, fmt.Errorf("its language server at %s did not answer (it has restarted or gone since the record was made): %w", s.Address, err)
	}
	defer resp.Body.Close()
	msg, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
	if resp.StatusCode != http.StatusOK {
		return Delivery{}, fmt.Errorf("its language server at %s answered HTTP %d: %s", s.Address, resp.StatusCode, bytes.TrimSpace(msg))
	}
	return Delivery{Route: RouteAgyServer, Via: s.Address}, nil
}

// agyMailboxMessage is the file Antigravity's own send_message writes into a recipient's mailbox.
type agyMailboxMessage struct {
	ID            string `json:"id"`
	Recipient     string `json:"recipient"`
	Sender        string `json:"sender"`
	Priority      string `json:"priority"`
	Timestamp     string `json:"timestamp"`
	RenderDetails struct {
		MessageTitle string `json:"messageTitle"`
	} `json:"renderDetails"`
	Content string `json:"content"`
}

// EnqueueAgy writes m into the recipient's mailbox under its brain directory (brain/<id>). The
// conversation reads it at its next turn; nothing wakes it.
func EnqueueAgy(conversationDir string, m AgyMessage, now time.Time) (Delivery, error) {
	id, err := newID()
	if err != nil {
		return Delivery{}, err
	}
	box := filepath.Join(conversationDir, ".system_generated", "messages")
	if err := os.MkdirAll(filepath.Join(box, "undelivered"), 0o755); err != nil {
		return Delivery{}, err
	}
	msg := agyMailboxMessage{ID: id, Recipient: m.Recipient, Sender: m.Sender, Priority: "MESSAGE_PRIORITY_HIGH",
		Timestamp: now.UTC().Format(time.RFC3339Nano), Content: m.Content}
	msg.RenderDetails.MessageTitle = m.Title
	b, err := json.Marshal(msg)
	if err != nil {
		return Delivery{}, err
	}
	path := filepath.Join(box, id+".json")
	// The body lands before the marker, and by rename, so the client never sees a marker whose
	// message is missing or half written.
	tmp := filepath.Join(box, "."+id+".tmp")
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return Delivery{}, err
	}
	if err := os.Rename(tmp, path); err != nil {
		os.Remove(tmp)
		return Delivery{}, err
	}
	if err := os.WriteFile(filepath.Join(box, "undelivered", id), nil, 0o644); err != nil {
		return Delivery{}, err
	}
	return Delivery{Route: RouteAgyMailbox, Via: path}, nil
}

// safeName refuses an id that could name a path outside the directory it is joined to.
func safeName(id string) bool {
	return id != "" && id != "." && id != ".." && filepath.Base(id) == id && !bytes.ContainsAny([]byte(id), `/\`)
}
