package peer

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/gofrs/flock"

	"github.com/ctoforaday/special-circumstances/plugins/gray-area/tools/internal/catalogue"
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

// RecordShape is the one shape of server record this binary reads and writes. A file in the
// registry with any other shape is not a record: the reader refuses it and Prune deletes it.
const RecordShape = 2

// Server is one Antigravity language server launch, stored keyed by server in the runtime
// directory so any conversation that links on a restarted daemon makes every idle conversation
// on that daemon wakeable.
type Server struct {
	Shape     int    `json:"shape"`
	Key       string `json:"key"`
	PID       int    `json:"pid,omitempty"`
	StartTime int64  `json:"start_time,omitempty"`
	// PidDomain is catalogue.LocalPidDomain() of the process that wrote the record: a pid means
	// something only inside the namespace it was read in, so a record from another one reads
	// Unknown (RecordLiveness), never Ended.
	PidDomain     string               `json:"pid_domain,omitempty"`
	Address       string               `json:"address"`
	Token         string               `json:"token"`
	Daemon        bool                 `json:"daemon,omitempty"`
	RegisteredAt  int64                `json:"registered_at"` // when this server launch was first recorded
	Conversations []LinkedConversation `json:"conversations"`
}

// LinkedConversation is one conversation a server record lists, and when it last linked there.
type LinkedConversation struct {
	ID       string `json:"id"`
	LinkedAt int64  `json:"linked_at"`
}

// Lists reports whether s lists conversation.
func (s Server) Lists(conversation string) bool {
	_, ok := s.linkedAt(conversation)
	return ok
}

func (s Server) linkedAt(conversation string) (int64, bool) {
	for _, c := range s.Conversations {
		if c.ID == conversation {
			return c.LinkedAt, true
		}
	}
	return 0, false
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

// LinkedDir is where RegisterWith leaves an empty marker per conversation that has ever linked
// (<runtimeDir>/gray-area/antigravity-linked/<conversation>). Prune never touches it: it is the
// evidence, surviving its server's record, that a conversation linked before. It lives on the same
// tmpfs as the records, so a reboot clears it.
func LinkedDir(runtimeDir string) string {
	if runtimeDir == "" {
		return ""
	}
	return filepath.Join(runtimeDir, "gray-area", "antigravity-linked")
}

// HasLinked reports whether conversation has linked since the runtime directory was created.
func HasLinked(runtimeDir, conversation string) bool {
	dir := LinkedDir(runtimeDir)
	if dir == "" || !safeName(conversation) {
		return false
	}
	_, err := os.Lstat(filepath.Join(dir, conversation))
	return err == nil
}

// Linked is what RegisterWith found.
type Linked int

const (
	// NotAnAntigravityShell: the process's environment carries no server to record.
	NotAnAntigravityShell Linked = iota
	// NoRuntimeDir: there is nowhere private to keep the record ($XDG_RUNTIME_DIR unset).
	NoRuntimeDir
	// AlreadyLinked: the record already named this server launch and conversation; its linked_at
	// was refreshed.
	AlreadyLinked
	// Recorded: the record was written.
	Recorded
)

// RegisterWith records the server of the Antigravity shell this process runs in, keyed by server,
// using host to identify the server process launch. It sets the conversation's linked_at to now
// whether or not the record already listed it, and marks the conversation as having linked
// (LinkedDir) in the same critical section, before the record is written.
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

	if err := os.MkdirAll(dir, 0o700); err != nil {
		return Server{}, NotAnAntigravityShell, err
	}
	if runtime.GOOS != "windows" {
		if err := os.Chmod(dir, 0o700); err != nil {
			return Server{}, NotAnAntigravityShell, err
		}
	}

	unlock, err := lockRegistry(dir)
	if err != nil {
		return Server{}, NotAnAntigravityShell, err
	}
	defer unlock()

	linked := LinkedConversation{ID: conv, LinkedAt: now.Unix()}
	s := Server{
		Shape:         RecordShape,
		Key:           key,
		PID:           id.PID,
		StartTime:     id.StartTime,
		PidDomain:     catalogue.LocalPidDomain(),
		Address:       addr,
		Token:         tok,
		Daemon:        id.Daemon,
		RegisteredAt:  now.Unix(),
		Conversations: []LinkedConversation{linked},
	}
	if err := markLinked(runtimeDir, conv); err != nil {
		return Server{}, NotAnAntigravityShell, err
	}
	got := Recorded
	if prev, err := LookupServerByKey(runtimeDir, key); err == nil && prev != nil &&
		prev.Address == addr && prev.Token == tok {
		s.RegisteredAt = prev.RegisteredAt
		s.Daemon = prev.Daemon || id.Daemon
		s.Conversations = nil
		for _, c := range prev.Conversations {
			if c.ID == conv {
				got = AlreadyLinked
				continue
			}
			s.Conversations = append(s.Conversations, c)
		}
		s.Conversations = append(s.Conversations, linked)
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
	if err := renameRecord(tmp.Name(), filepath.Join(dir, key+".json")); err != nil {
		return Server{}, NotAnAntigravityShell, err
	}
	return s, got, nil
}

// THE READERS TAKE NO LOCK. Every write to the registry — RegisterWith's and Prune's — runs under
// lockRegistry and replaces a record only by renaming a complete temporary file over it in the same
// directory, or removes it whole, so a reader sees a whole record or none. A record that vanishes
// between listing the directory and reading the file was pruned meanwhile, and reads as absent.

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
	if s.Shape != RecordShape {
		return nil, fmt.Errorf("server record %s has shape %d, not %d", key, s.Shape, RecordShape)
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

// LookupConversationServer returns the server record listing conversation that liveness reads
// Live, choosing the one the conversation linked to most recently. None is (nil, nil).
func LookupConversationServer(runtimeDir, conversation string, liveness func(Server) catalogue.Liveness) (*Server, error) {
	if runtimeDir == "" || !safeName(conversation) {
		return nil, nil
	}
	servers, err := ListServers(runtimeDir)
	if err != nil {
		return nil, err
	}
	return NewestLiveListing(servers, conversation, liveness), nil
}

// NewestLiveListing is the record in servers that lists conversation, reads Live, and has the
// greatest linked_at for it; nil when none does.
func NewestLiveListing(servers []Server, conversation string, liveness func(Server) catalogue.Liveness) *Server {
	var best *Server
	var bestAt int64
	for i := range servers {
		at, ok := servers[i].linkedAt(conversation)
		if !ok || liveness(servers[i]) != catalogue.Live {
			continue
		}
		if best == nil || at > bestAt {
			s := servers[i]
			best, bestAt = &s, at
		}
	}
	return best
}

// RecordLiveness is whether the language server a record names is still running: the catalogue's
// three answers, by the catalogue's pid-domain rule. Live needs the record's pid to be running
// under the record's start time in the record's own pid domain. Ended needs positive evidence: no
// such process, or that pid running under a different start time. Everything else, including a
// record keyed by address (no pid), a record with no domain, and a /proc that cannot be read, is
// Unknown, which nothing deletes and nothing delivers to.
func RecordLiveness(s Server) catalogue.Liveness {
	return recordLiveness(s, catalogue.LocalPidDomain(), procStartTime)
}

func recordLiveness(s Server, localDomain string, start func(int) (int64, error)) catalogue.Liveness {
	if s.PidDomain == "" || catalogue.ForeignDomain(s.PidDomain, localDomain) {
		return catalogue.Unknown
	}
	if s.PID <= 0 || s.StartTime <= 0 {
		return catalogue.Unknown
	}
	got, err := start(s.PID)
	switch {
	case errors.Is(err, fs.ErrNotExist):
		return catalogue.Ended
	case err != nil:
		return catalogue.Unknown
	case got != s.StartTime:
		return catalogue.Ended // the pid now belongs to another process
	}
	return catalogue.Live
}

// Prune deletes, in one critical section under the registry lock — listing, reading, judging and
// deleting all while holding it — every file in the registry that is not a record of a server that
// may still be running: a JSON object whose shape is not RecordShape, and a record
// liveness reads Ended. It keeps everything else — a file it cannot read or parse, a record that
// reads Unknown, and every file not named *.json, the lock among them. With no registry directory
// it does nothing and creates nothing. It returns the names it deleted.
func Prune(runtimeDir string, liveness func(Server) catalogue.Liveness) ([]string, error) {
	dir := RegistryDir(runtimeDir)
	if dir == "" {
		return nil, nil
	}
	if _, err := os.Stat(dir); errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	unlock, err := lockRegistry(dir)
	if err != nil {
		return nil, err
	}
	defer unlock()
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var removed []string
	for _, e := range entries {
		if !strings.HasSuffix(e.Name(), ".json") || !e.Type().IsRegular() {
			continue
		}
		path := filepath.Join(dir, e.Name())
		b, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var fields map[string]json.RawMessage
		if json.Unmarshal(b, &fields) != nil || fields == nil {
			continue
		}
		var shape int
		if raw, ok := fields["shape"]; !ok || json.Unmarshal(raw, &shape) != nil || shape != RecordShape {
			if os.Remove(path) == nil {
				removed = append(removed, e.Name())
			}
			continue
		}
		var s Server
		if json.Unmarshal(b, &s) != nil {
			continue
		}
		if liveness(s) == catalogue.Ended {
			if os.Remove(path) == nil {
				removed = append(removed, e.Name())
			}
		}
	}
	return removed, nil
}

// userHZ is the Linux userspace clock-tick rate, in which /proc/<pid>/stat states a process's
// start time. It is fixed by the kernel ABI; reading it through sysconf would need cgo.
const userHZ = 100

// StartTimeAt converts a start time in clock ticks since boot to the instant, given the boot in
// Unix seconds.
func StartTimeAt(bootUnix, ticks int64) time.Time {
	return time.Unix(bootUnix, 0).Add(time.Duration(ticks) * time.Second / userHZ)
}

// AgyMessage is one message to an Antigravity conversation.
type AgyMessage struct {
	Recipient string
	Sender    string
	Title     string
	Content   string
}

// SendAgyServer delivers m through the language server s, reporting route as the route it took. It
// wakes an idle conversation and loads an unloaded one from disk and starts a turn.
func SendAgyServer(client *http.Client, s Server, m AgyMessage, route Route) (Delivery, error) {
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
		var op *net.OpError
		if errors.As(err, &op) && op.Op == "dial" {
			return Delivery{}, &UnreachedError{Address: s.Address, Err: err}
		}
		return Delivery{}, fmt.Errorf("the language server recorded at %s was sent the message and did not confirm it: %w", s.Address, err)
	}
	defer resp.Body.Close()
	msg, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
	if resp.StatusCode != http.StatusOK {
		return Delivery{}, fmt.Errorf("the language server recorded at %s answered HTTP %d: %s", s.Address, resp.StatusCode, bytes.TrimSpace(msg))
	}
	return Delivery{Route: route, Via: s.Address}, nil
}

// UnreachedError is a SendAgyServer failure in which the message never reached the server: no
// connection to it could be opened. Every other failure — no answer, a dropped connection, an error
// status — can follow the server's receipt of the message, so whether it arrived is unknown.
type UnreachedError struct {
	Address string
	Err     error
}

func (e *UnreachedError) Error() string {
	return fmt.Sprintf("the language server recorded at %s cannot be reached (it has restarted or gone since the record was made): %v", e.Address, e.Err)
}

func (e *UnreachedError) Unwrap() error { return e.Err }

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

// HolderKind is how a process was found holding a conversation.
type HolderKind int

const (
	// HolderNone: no process but the excluded one holds the conversation.
	HolderNone HolderKind = iota
	// HolderPresenceLock: a process holds presence/<id>.lock.
	HolderPresenceLock
	// HolderDBOpen: a process has conversations/<id>.db (or its -wal/-shm) open.
	HolderDBOpen
	// HolderUninspectable: whether a process holds it could not be established.
	HolderUninspectable
)

// Holder is what ConversationBusy found. PID is set for every kind but HolderNone and
// HolderUninspectable; callers branch on PID and Kind, and Detail is only ever printed.
type Holder struct {
	PID    int
	Kind   HolderKind
	Detail string
}

// SystemSender is the sender agy stamps on its own notices in a mailbox. They are not mail between
// agents.
const SystemSender = "system"

// ShellSender is the sender telepathy stamps on a message sent from outside any agent session.
const ShellSender = "telepathy"

// Waiting is one Antigravity conversation whose mailbox holds messages its client has not read.
type Waiting struct {
	Conversation string
	// Messages counts every marker that is not one of agy's own notices: the readable ones and the
	// unreadable ones.
	Messages int
	// Oldest is the earliest timestamp among the readable messages; zero when none is readable.
	Oldest time.Time
	// Yours counts the messages whose sender is the caller; FromShell those sent from outside any
	// agent session; Unreadable those whose file is missing or does not parse; Notices agy's own.
	Yours, FromShell, Unreadable, Notices int
}

// ScanMailboxes lists, across brainDirs, every conversation with at least one marker under
// .system_generated/messages/undelivered — the client's own record of what it has not read — and
// counts its messages; caller is the scanning session's sender id ("" or ShellSender for none). It
// takes no lock the client honours. A marker that disappears while it scans is a message the
// client just read, and is skipped; a marker whose message file is missing, or does not parse,
// is unreadable and counted.
// A brain directory that does not exist holds no conversation. A directory that cannot be listed
// is an error, returned beside whatever the other directories gave, never folded into "none".
func ScanMailboxes(brainDirs []string, caller string) ([]Waiting, []error) {
	var out []Waiting
	var errs []error
	for _, brain := range brainDirs {
		convs, err := os.ReadDir(brain)
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}
		if err != nil {
			errs = append(errs, err)
			continue
		}
		for _, c := range convs {
			if !c.IsDir() {
				continue
			}
			box := filepath.Join(brain, c.Name(), ".system_generated", "messages")
			markers, err := os.ReadDir(filepath.Join(box, "undelivered"))
			if errors.Is(err, fs.ErrNotExist) {
				continue
			}
			if err != nil {
				errs = append(errs, err)
				continue
			}
			w := Waiting{Conversation: c.Name()}
			for _, m := range markers {
				if m.IsDir() {
					continue
				}
				var msg agyMailboxMessage
				b, err := os.ReadFile(filepath.Join(box, m.Name()+".json"))
				if err != nil {
					// The client reads a message while this scans: a marker gone by now is a
					// message just read, not an unreadable one.
					if _, merr := os.Lstat(filepath.Join(box, "undelivered", m.Name())); errors.Is(merr, fs.ErrNotExist) {
						continue
					}
					w.Messages++
					w.Unreadable++
					continue
				}
				if json.Unmarshal(b, &msg) != nil {
					w.Messages++
					w.Unreadable++
					continue
				}
				if msg.Sender == SystemSender {
					w.Notices++
					continue
				}
				at, err := time.Parse(time.RFC3339Nano, msg.Timestamp)
				if err != nil {
					w.Messages++
					w.Unreadable++
					continue
				}
				w.Messages++
				if w.Oldest.IsZero() || at.Before(w.Oldest) {
					w.Oldest = at
				}
				switch {
				case msg.Sender == ShellSender:
					w.FromShell++
				case caller != "" && caller != ShellSender && msg.Sender == caller:
					w.Yours++
				}
			}
			if w.Messages > 0 || w.Notices > 0 {
				out = append(out, w)
			}
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		a, b := out[i], out[j]
		if a.Oldest.IsZero() != b.Oldest.IsZero() {
			return !a.Oldest.IsZero()
		}
		if !a.Oldest.Equal(b.Oldest) {
			return a.Oldest.Before(b.Oldest)
		}
		return a.Conversation < b.Conversation
	})
	return out, errs
}

// markLinked leaves conversation's marker in LinkedDir (directory 0700, marker 0600 and empty).
func markLinked(runtimeDir, conversation string) error {
	dir := LinkedDir(runtimeDir)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	if runtime.GOOS != "windows" {
		if err := os.Chmod(dir, 0o700); err != nil {
			return err
		}
	}
	f, err := os.OpenFile(filepath.Join(dir, conversation), os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	return f.Close()
}

// lockRegistry serialises RegisterWith's read-modify-write updates and Prune's deletions in dir,
// across every telepathy process on the box, whichever install it came from. The lock is a flock
// on .register.lock, which nothing ever unlinks, renames or recreates: a flock belongs to an open
// file description, so a lock file replaced while held would let two processes both hold "it".
//
// It waits at most registryLockWait: every telepathy verb prunes, so a holder that never lets go
// would otherwise hang every verb on the box; past the bound the caller is refused and says so.
func lockRegistry(dir string) (func(), error) {
	path := filepath.Join(dir, ".register.lock")
	fl := flock.New(path, flock.SetFlag(os.O_CREATE|os.O_RDWR), flock.SetPermissions(0o600))
	ctx, cancel := context.WithTimeout(context.Background(), registryLockWait)
	defer cancel()
	ok, err := fl.TryLockContext(ctx, 10*time.Millisecond)
	switch {
	case ok:
		return func() { _ = fl.Unlock() }, nil
	case err != nil && !errors.Is(err, context.DeadlineExceeded):
		return nil, err
	}
	return nil, fmt.Errorf("another process has held %s for more than %s", path, registryLockWait)
}

// registryLockWait bounds the wait for the registry lock; a variable so a test can change it.
var registryLockWait = 2 * time.Second

// safeName refuses an id that could name a path outside the directory it is joined to.
func safeName(id string) bool {
	return id != "" && id != "." && id != ".." && filepath.Base(id) == id && !bytes.ContainsAny([]byte(id), `/\`)
}
