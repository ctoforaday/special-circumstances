package sessionstart

import (
	"bytes"
	"encoding/json"
	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/hookfailures"
	"os"
	"strings"
	"testing"
	"time"
)

func TestNoRootIsRecordedAndNotMarkedSaid(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	var o, e bytes.Buffer
	run(nil, strings.NewReader(`{}`), &o, &e, "", time.Now(), nil)
	if o.Len() != 0 {
		t.Fatalf("no project root, and stdout was written: %q", o.String())
	}
	if recorded, unsaid := unsaidRoot(t); !recorded || !unsaid {
		t.Fatalf("recorded=%v unsaid=%v — the entry must be kept and NOT marked as told", recorded, unsaid)
	}
}

// unsaidRoot reports whether a project-root entry is on the record AND has never been marked said.
// A binary with no project root writes nothing to stdout, so it must not mark anything as told —
// otherwise the human never hears it and the next call that could tell them stays quiet for ten
// minutes.
func unsaidRoot(t *testing.T) (recorded, unsaid bool) {
	t.Helper()
	path, err := hookfailures.Path("prosthetic-conscience")
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return false, false
	}
	var rec hookfailures.Record
	if err := json.Unmarshal(b, &rec); err != nil {
		t.Fatal(err)
	}
	for _, f := range rec.Failures {
		if f.Stage == "project-root" {
			return true, f.Notified.IsZero()
		}
	}
	return false, false
}
