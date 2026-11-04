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
	"time"
)

// The three variables Antigravity puts in the environment of an agent's shell commands.
const (
	EnvConversation = "ANTIGRAVITY_CONVERSATION_ID"
	EnvAddress      = "ANTIGRAVITY_LS_ADDRESS"
	EnvToken        = "ANTIGRAVITY_CSRF_TOKEN"
)

// Server is the language server hosting one Antigravity conversation, as that conversation's own
// shell reported it.
type Server struct {
	Conversation string `json:"conversation"`
	Address      string `json:"address"`
	Token        string `json:"token"`
	RegisteredAt int64  `json:"registered_at"` // when this server launch was first recorded
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
	// AlreadyLinked: the record already names this server launch; nothing was written.
	AlreadyLinked
	// Recorded: the record was written.
	Recorded
)

// Register records the server of the Antigravity conversation whose shell this process runs in,
// once per server launch.
//
// The token is the server launch's identity: Antigravity mints a new one every time a server
// starts, so a record holding the same address and token is this server already recorded, and a
// record holding anything else is a server that has since gone. That is exact where a pid is not —
// the shell's parent chain says nothing about which server answers the address.
func Register(runtimeDir string, getenv func(string) string, now time.Time) (Server, Linked, error) {
	s := Server{Conversation: getenv(EnvConversation), Address: getenv(EnvAddress), Token: getenv(EnvToken),
		RegisteredAt: now.Unix()}
	if s.Conversation == "" || s.Address == "" || s.Token == "" {
		return s, NotAnAntigravityShell, nil
	}
	dir := RegistryDir(runtimeDir)
	if dir == "" {
		return s, NoRuntimeDir, nil
	}
	if !safeName(s.Conversation) {
		return s, NotAnAntigravityShell, fmt.Errorf("%s %q is not a conversation id", EnvConversation, s.Conversation)
	}
	if prev, err := LookupServer(runtimeDir, s.Conversation); err == nil && prev != nil &&
		prev.Address == s.Address && prev.Token == s.Token {
		return *prev, AlreadyLinked, nil
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return s, NotAnAntigravityShell, err
	}
	b, err := json.Marshal(s)
	if err != nil {
		return s, NotAnAntigravityShell, err
	}
	tmp, err := os.CreateTemp(dir, ".register-*")
	if err != nil {
		return s, NotAnAntigravityShell, err
	}
	defer os.Remove(tmp.Name())
	if err := tmp.Chmod(0o600); err != nil {
		tmp.Close()
		return s, NotAnAntigravityShell, err
	}
	if _, err := tmp.Write(b); err != nil {
		tmp.Close()
		return s, NotAnAntigravityShell, err
	}
	if err := tmp.Close(); err != nil {
		return s, NotAnAntigravityShell, err
	}
	if err := os.Rename(tmp.Name(), filepath.Join(dir, s.Conversation+".json")); err != nil {
		return s, NotAnAntigravityShell, err
	}
	return s, Recorded, nil
}

// LookupServer reads a conversation's server record. A missing record is (nil, nil): the
// conversation has not run telepathy since its server started, which is the ordinary case.
func LookupServer(runtimeDir, conversation string) (*Server, error) {
	dir := RegistryDir(runtimeDir)
	if dir == "" || !safeName(conversation) {
		return nil, nil
	}
	b, err := os.ReadFile(filepath.Join(dir, conversation+".json"))
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var s Server
	if err := json.Unmarshal(b, &s); err != nil {
		return nil, fmt.Errorf("server record for %s is unreadable: %w", conversation, err)
	}
	return &s, nil
}

// AgyMessage is one message to an Antigravity conversation.
type AgyMessage struct {
	Recipient string
	Sender    string
	Title     string
	Content   string
}

// SendAgyServer delivers m through the language server that hosts the recipient. It is the only
// route that wakes an idle conversation.
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
