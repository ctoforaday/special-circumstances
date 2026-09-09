package catalogue

import (
	"os"
	"path/filepath"
	"testing"
)

const domain = "linux:aaaabbbbccccdddd:pid:[4026531836]"

// A RECYCLED PID MUST NOT READ AS ALIVE. pids are reused, so pid alone would report a dead agent
// as live — a wrong answer indistinguishable from a right one, which is the whole class this
// plugin refuses. procStart makes the pair a durable identity.
func TestStaleProcStartIsEnded(t *testing.T) {
	sf := SessionFile{PID: 4242, SessionID: "S", ProcStart: "5677", PidDomain: domain}
	// The pid exists, but it started at a different time: a different process wearing the number.
	got := Of(sf, domain, func(int) (string, bool) { return "999999", true })
	if got != Ended {
		t.Errorf("liveness = %q, want %q — a recycled pid must not read as live", got, Ended)
	}
	// Same pid, same start time: genuinely the session's process.
	if got := Of(sf, domain, func(int) (string, bool) { return "5677", true }); got != Live {
		t.Errorf("liveness = %q, want %q", got, Live)
	}
	// Gone entirely.
	if got := Of(sf, domain, func(int) (string, bool) { return "", false }); got != Ended {
		t.Errorf("liveness = %q, want %q", got, Ended)
	}
}

// A FOREIGN pidDomain IS UNKNOWN, NOT ENDED. The vendor writes linux:<machine-id>:<pid ns>
// precisely because a pid means nothing outside its namespace; a session file from a container or
// another host sharing $HOME would otherwise resolve to Ended while its agent is running.
func TestForeignPidDomainIsUnknown(t *testing.T) {
	sf := SessionFile{PID: 1, SessionID: "S", ProcStart: "5677", PidDomain: "linux:SOMEONE-ELSE:pid:[4026531999]"}
	// Even with a locally matching pid AND start time, the answer is unknown.
	got := Of(sf, domain, func(int) (string, bool) { return "5677", true })
	if got != Unknown {
		t.Errorf("liveness = %q, want %q — a pid from another domain is not ours to judge", got, Unknown)
	}
}

// Off Linux LocalPidDomain is empty, and every session reads Unknown rather than being guessed at.
func TestNoLocalDomainMeansUnknown(t *testing.T) {
	sf := SessionFile{PID: 1, SessionID: "S", ProcStart: "5677", PidDomain: domain}
	if got := Of(sf, "", func(int) (string, bool) { return "5677", true }); got != Unknown {
		t.Errorf("liveness = %q, want %q when the querier cannot establish its own domain", got, Unknown)
	}
}

func TestReadSessionFilesSkipsUnusableEntries(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "1.json"), []byte(`{"pid":1,"sessionId":"S1","procStart":"11","pidDomain":"`+domain+`"}`), 0o600)
	os.WriteFile(filepath.Join(dir, "2.json"), []byte(`not json`), 0o600)
	os.WriteFile(filepath.Join(dir, "3.json"), []byte(`{"pid":3}`), 0o600) // no sessionId: names nothing
	got, err := ReadSessionFiles(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].SessionID != "S1" {
		t.Fatalf("got %+v, want exactly the one usable entry", got)
	}
	if got[0].PidDomain != domain {
		t.Errorf("pidDomain was dropped: %q", got[0].PidDomain)
	}
}
