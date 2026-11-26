package peer

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"testing"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/gray-area/tools/internal/catalogue"
)

// allLive reads every record Live, for tests about which record is chosen rather than whether
// its server runs.
func allLive(Server) catalogue.Liveness { return catalogue.Live }

// livenessByKey reads each record by its key, Unknown when unlisted.
func livenessByKey(m map[string]catalogue.Liveness) func(Server) catalogue.Liveness {
	return func(s Server) catalogue.Liveness {
		if l, ok := m[s.Key]; ok {
			return l
		}
		return catalogue.Unknown
	}
}

func writeRecord(t *testing.T, rt, name string, v any) {
	t.Helper()
	dir := RegistryDir(rt)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	var b []byte
	switch x := v.(type) {
	case string:
		b = []byte(x)
	default:
		var err error
		if b, err = json.Marshal(v); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(dir, name), b, 0o600); err != nil {
		t.Fatal(err)
	}
}

func record(key string, pid int, start int64, conv ...LinkedConversation) Server {
	return Server{Shape: RecordShape, Key: key, PID: pid, StartTime: start, PidDomain: "linux:test:pid",
		Address: "localhost:" + key, Token: "tok", RegisteredAt: at.Unix(), Conversations: conv}
}

func registryNames(t *testing.T, rt string) []string {
	t.Helper()
	entries, err := os.ReadDir(RegistryDir(rt))
	if err != nil {
		t.Fatal(err)
	}
	var out []string
	for _, e := range entries {
		out = append(out, e.Name())
	}
	return out
}

// THE CONVERSATION'S LATEST LINK WINS, NOT THE SERVER'S FIRST RECORD. S1 was recorded first and the
// conversation linked to it again later; S2 was recorded in between. Ordering by registered_at
// would pick S2.
func TestLookupPicksGreatestLinkedAt(t *testing.T) {
	rt := t.TempDir()
	s1 := record("s1", 11, 111, LinkedConversation{ID: "conv", LinkedAt: at.Add(3 * time.Hour).Unix()})
	s2 := record("s2", 22, 222, LinkedConversation{ID: "conv", LinkedAt: at.Add(time.Hour).Unix()})
	s2.RegisteredAt = at.Add(time.Hour).Unix()
	writeRecord(t, rt, "s1.json", s1)
	writeRecord(t, rt, "s2.json", s2)
	got, err := LookupConversationServer(rt, "conv", allLive)
	if err != nil || got == nil || got.Key != "s1" {
		t.Errorf("LookupConversationServer = %+v, %v; want s1, the greatest linked_at", got, err)
	}
}

// A DEAD SERVER NEVER SHADOWS A LIVE ONE: the newer link is to a server that has gone.
func TestLookupSkipsEndedNewerRecord(t *testing.T) {
	rt := t.TempDir()
	writeRecord(t, rt, "old.json", record("old", 11, 111, LinkedConversation{ID: "conv", LinkedAt: at.Unix()}))
	writeRecord(t, rt, "new.json", record("new", 22, 222, LinkedConversation{ID: "conv", LinkedAt: at.Add(time.Hour).Unix()}))
	live := livenessByKey(map[string]catalogue.Liveness{"old": catalogue.Live, "new": catalogue.Ended})
	got, err := LookupConversationServer(rt, "conv", live)
	if err != nil || got == nil || got.Key != "old" {
		t.Errorf("LookupConversationServer = %+v, %v; want old, the Live record", got, err)
	}
	unknown := livenessByKey(map[string]catalogue.Liveness{"old": catalogue.Unknown, "new": catalogue.Ended})
	if got, err := LookupConversationServer(rt, "conv", unknown); got != nil || err != nil {
		t.Errorf("with no Live record: LookupConversationServer = %+v, %v; want nil — Unknown is never a target", got, err)
	}
}

func TestLinkedAtRefreshesOnRelink(t *testing.T) {
	rt := t.TempDir()
	h := hostID(100, 5000, false)
	if _, got, err := RegisterWith(rt, agyShell("conv-1", "localhost:1", "tok"), h, at); got != Recorded || err != nil {
		t.Fatalf("first link = %v, %v", got, err)
	}
	if _, got, err := RegisterWith(rt, agyShell("conv-2", "localhost:1", "tok"), h, at.Add(time.Minute)); got != Recorded || err != nil {
		t.Fatalf("second conversation = %v, %v", got, err)
	}
	s, got, err := RegisterWith(rt, agyShell("conv-1", "localhost:1", "tok"), h, at.Add(time.Hour))
	if got != AlreadyLinked || err != nil {
		t.Fatalf("relink = %v, %v; want AlreadyLinked", got, err)
	}
	stored, err := LookupServerByKey(rt, s.Key)
	if err != nil || stored == nil {
		t.Fatalf("LookupServerByKey = %+v, %v", stored, err)
	}
	if at1, _ := stored.linkedAt("conv-1"); at1 != at.Add(time.Hour).Unix() {
		t.Errorf("conv-1 linked_at = %d, want the relink's %d", at1, at.Add(time.Hour).Unix())
	}
	if at2, _ := stored.linkedAt("conv-2"); at2 != at.Add(time.Minute).Unix() {
		t.Errorf("conv-2 linked_at = %d, want its own %d untouched", at2, at.Add(time.Minute).Unix())
	}
	if stored.RegisteredAt != at.Unix() || len(stored.Conversations) != 2 {
		t.Errorf("record = %+v; want registered_at kept and two conversations", stored)
	}
}

func TestReaderRejectsOldShape(t *testing.T) {
	rt := t.TempDir()
	writeRecord(t, rt, "pid-1-2.json", `{"key":"pid-1-2","pid":1,"start_time":2,"address":"localhost:1","token":"t","conversations":["conv"],"registered_at":1}`)
	if s, err := LookupServerByKey(rt, "pid-1-2"); s != nil || err == nil {
		t.Errorf("LookupServerByKey on a record with no shape = %+v, %v; want a refusal", s, err)
	}
	writeRecord(t, rt, "pid-3-4.json", `{"shape":3,"key":"pid-3-4","address":"localhost:1","token":"t","conversations":[]}`)
	if s, err := LookupServerByKey(rt, "pid-3-4"); s != nil || err == nil {
		t.Errorf("LookupServerByKey on shape 3 = %+v, %v; want a refusal", s, err)
	}
	if servers, err := ListServers(rt); err != nil || len(servers) != 0 {
		t.Errorf("ListServers = %+v, %v; want no records", servers, err)
	}
}

func TestRecordLiveness(t *testing.T) {
	const local = "linux:machine:pid:[4026531836]"
	running := map[int]int64{100: 5000, 200: 7000}
	start := func(pid int) (int64, error) {
		if st, ok := running[pid]; ok {
			return st, nil
		}
		return 0, &fs.PathError{Op: "open", Path: fmt.Sprintf("/proc/%d/stat", pid), Err: fs.ErrNotExist}
	}
	unreadable := func(int) (int64, error) { return 0, errors.New("permission denied") }
	rec := func(pid int, st int64, domain string) Server {
		return Server{Shape: RecordShape, Key: ServerKey(pid, st, "localhost:1"), PID: pid, StartTime: st, PidDomain: domain}
	}
	for _, tc := range []struct {
		name  string
		s     Server
		start func(int) (int64, error)
		want  catalogue.Liveness
	}{
		{"live", rec(100, 5000, local), start, catalogue.Live},
		{"reused pid", rec(200, 6000, local), start, catalogue.Ended},
		{"missing pid", rec(300, 9000, local), start, catalogue.Ended},
		{"unreadable /proc", rec(100, 5000, local), unreadable, catalogue.Unknown},
		{"foreign pid_domain", rec(300, 9000, "linux:other:pid:[1]"), start, catalogue.Unknown},
		{"no pid_domain", rec(300, 9000, ""), start, catalogue.Unknown},
		{"addr- key", rec(0, 0, local), start, catalogue.Unknown},
	} {
		if got := recordLiveness(tc.s, local, tc.start); got != tc.want {
			t.Errorf("%s: recordLiveness = %s, want %s", tc.name, got, tc.want)
		}
	}
	if got := recordLiveness(rec(100, 5000, local), "", start); got != catalogue.Unknown {
		t.Errorf("no local domain: recordLiveness = %s, want unknown", got)
	}
}

// THE RECORD REGISTERWITH WRITES CARRIES WHAT LIVENESS NEEDS. Measured against this process: a
// record of the running test binary reads Live, and the same record for a pid that does not exist
// reads Ended — the real /proc, the real pid domain, no stand-ins.
func TestRegisteredRecordOfARunningServerReadsLive(t *testing.T) {
	if runtime.GOOS != "linux" || catalogue.LocalPidDomain() == "" {
		t.Skip("liveness is measured on Linux with a machine id only")
	}
	me := os.Getpid()
	start, ok := ProcessStartTime(me)
	if !ok {
		t.Fatal("ProcessStartTime of this process is unmeasurable")
	}
	rt := t.TempDir()
	s, _, err := RegisterWith(rt, agyShell("conv-1", "localhost:1", "tok"), hostID(me, start, false), at)
	if err != nil {
		t.Fatal(err)
	}
	if got := RecordLiveness(s); got != catalogue.Live {
		t.Errorf("RecordLiveness of this running process's record = %s, want live", got)
	}
	gone := 4194303
	for ; gone > 1; gone-- {
		if _, err := os.Stat(fmt.Sprintf("/proc/%d", gone)); errors.Is(err, fs.ErrNotExist) {
			break
		}
	}
	s.PID = gone
	if got := RecordLiveness(s); got != catalogue.Ended {
		t.Errorf("RecordLiveness of pid %d, which does not exist = %s, want ended", gone, got)
	}
}

func TestPruneDeletesOnlyEndedAndOldShape(t *testing.T) {
	rt := t.TempDir()
	writeRecord(t, rt, "live.json", record("live", 1, 1))
	writeRecord(t, rt, "ended.json", record("ended", 2, 2))
	writeRecord(t, rt, "unknown.json", record("unknown", 3, 3))
	writeRecord(t, rt, "legacy.json", `{"key":"legacy","pid":4,"start_time":4,"address":"a","token":"t","conversations":["c"],"registered_at":1}`)
	writeRecord(t, rt, "garbage.json", `{not json`)
	writeRecord(t, rt, "array.json", `[1,2]`)
	writeRecord(t, rt, "notes.txt", `{"key":"x"}`)
	writeRecord(t, rt, ".register.lock", "")
	removed, err := Prune(rt, livenessByKey(map[string]catalogue.Liveness{
		"live": catalogue.Live, "ended": catalogue.Ended, "unknown": catalogue.Unknown, "legacy": catalogue.Ended,
	}))
	if err != nil {
		t.Fatal(err)
	}
	slices.Sort(removed)
	if want := []string{"ended.json", "legacy.json"}; !slices.Equal(removed, want) {
		t.Errorf("Prune removed %v, want %v", removed, want)
	}
	want := []string{".register.lock", "array.json", "garbage.json", "live.json", "notes.txt", "unknown.json"}
	if got := registryNames(t, rt); !slices.Equal(got, want) {
		t.Errorf("registry holds %v, want %v", got, want)
	}
}

func TestPruneNoRegistryDirIsSilent(t *testing.T) {
	rt := t.TempDir()
	removed, err := Prune(rt, allLive)
	if err != nil || removed != nil {
		t.Errorf("Prune with no registry = %v, %v; want nothing", removed, err)
	}
	if _, err := os.Stat(filepath.Join(rt, "gray-area")); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("Prune created %s: %v", filepath.Join(rt, "gray-area"), err)
	}
	if removed, err := Prune("", allLive); err != nil || removed != nil {
		t.Errorf("Prune with no runtime directory = %v, %v; want nothing", removed, err)
	}
}

// A LINK THAT ARRIVES WHILE PRUNE HOLDS THE LOCK WAITS, THEN LANDS. The liveness callback runs
// under the lock, starts the link, and sees it still blocked before letting prune finish.
func TestPruneUnderLockWithWaiter(t *testing.T) {
	rt := t.TempDir()
	writeRecord(t, rt, "stale.json", record("stale", 9, 9))
	done := make(chan error, 1)
	started := false
	var key string
	liveness := func(s Server) catalogue.Liveness {
		if !started {
			started = true
			go func() {
				rec, _, err := RegisterWith(rt, agyShell("conv-new", "localhost:7", "tok"), hostID(70, 700, false), at)
				key = rec.Key
				done <- err
			}()
			select {
			case err := <-done:
				t.Errorf("RegisterWith finished while Prune held the lock (err %v)", err)
				done <- err
			case <-time.After(200 * time.Millisecond):
			}
		}
		if s.Key == "stale" {
			return catalogue.Ended
		}
		return catalogue.Live
	}
	if _, err := Prune(rt, liveness); err != nil {
		t.Fatal(err)
	}
	if err := <-done; err != nil {
		t.Fatalf("RegisterWith after prune: %v", err)
	}
	if s, err := LookupServerByKey(rt, key); err != nil || s == nil || !s.Lists("conv-new") {
		t.Errorf("the waiting link's record = %+v, %v; want present", s, err)
	}
	if s, _ := LookupServerByKey(rt, "stale"); s != nil {
		t.Errorf("the Ended record survived prune")
	}
}

func TestTicksToTime(t *testing.T) {
	boot := time.Date(2026, 11, 26, 3, 0, 0, 0, time.UTC).Unix()
	if got, want := StartTimeAt(boot, 12345), time.Date(2026, 11, 26, 3, 2, 3, 450_000_000, time.UTC); !got.Equal(want) {
		t.Errorf("StartTimeAt(boot, 12345) = %s, want %s", got.UTC(), want)
	}
}
