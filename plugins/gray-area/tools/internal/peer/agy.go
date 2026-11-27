package peer

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// EnvConversation is the variable Antigravity puts in the environment of an agent's shell commands
// and of its hooks, naming the conversation.
const EnvConversation = "ANTIGRAVITY_CONVERSATION_ID"

// ServerIdentity identifies one Antigravity language-server process launch.
// On Linux, PID + StartTime (field 22 of /proc/<pid>/stat, clock ticks since boot)
// uniquely identifies a server launch even if a PID is reused after a restart.
type ServerIdentity struct {
	PID       int
	StartTime int64
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

// agyDialTimeout bounds opening a connection to a language server, below any whole-request bound.
const agyDialTimeout = 2 * time.Second

// NewAgyClient is the client SendWake and DaemonAddress are meant to be given: one request bounded
// by timeout, a dial bounded by agyDialTimeout, no proxy (a language server listens on loopback),
// and no redirect followed.
func NewAgyClient(timeout time.Duration) *http.Client {
	return &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			Proxy:       nil,
			DialContext: (&net.Dialer{Timeout: agyDialTimeout}).DialContext,
		},
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}
}

// WakeSender is the sender of the wake message, and WakeTitle and WakeContent its title and text:
// the exact message measured (agy 1.5.18, V1 7/7) to start a turn that first reads every queued
// mailbox message. It asks for nothing, so a wake repeated or lost costs a turn at most. A wake
// leaves its own entry in the mailbox under this sender until a turn reads it.
const (
	WakeSender  = "telepathy-wake"
	WakeTitle   = "telepathy"
	WakeContent = "You have queued messages."
)

// SendWake sends the wake message for conversation to the language server at address, with the
// server's CSRF token. Any answer but 200 is an error; the error never carries the token.
func SendWake(client *http.Client, address, token, conversation string) error {
	body, err := json.Marshal(map[string]string{
		"recipient": conversation, "content": WakeContent, "senderId": WakeSender, "displayTitle": WakeTitle,
	})
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost,
		"http://"+address+"/exa.language_server_pb.LanguageServerService/SendAgentMessage", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-codeium-csrf-token", token)
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("the language server at %s: %w", address, err)
	}
	defer resp.Body.Close()
	msg, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("the language server at %s answered HTTP %d: %s", address, resp.StatusCode, bytes.TrimSpace(msg))
	}
	return nil
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
// conversation reads it at the start of its next turn, whatever starts that turn (a wake among them).
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
	// agent session; Unreadable those whose file is missing or does not parse; Notices agy's own;
	// Wakes telepathy's wake messages (WakeSender). Neither Notices nor Wakes is mail: neither is in
	// Messages nor in Oldest.
	Yours, FromShell, Unreadable, Notices, Wakes int
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
				switch msg.Sender {
				case SystemSender:
					w.Notices++
					continue
				case WakeSender:
					w.Wakes++
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
			if w.Messages > 0 || w.Notices > 0 || w.Wakes > 0 {
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

// safeName refuses an id that could name a path outside the directory it is joined to.
func safeName(id string) bool {
	return id != "" && id != "." && id != ".." && filepath.Base(id) == id && !bytes.ContainsAny([]byte(id), `/\`)
}
