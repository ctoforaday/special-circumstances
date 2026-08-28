package setup

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// A RUN THAT CANNOT RESOLVE IS REFUSED BEFORE ANYTHING IS LAID OUT.
//
// The state is the documented one: a run directory carrying .records-elsewhere, with no pointer
// in the user cache and no FEOV_RECORD_ROOT. That is what a copied run directory or a cleaned
// cache looks like, and the record is real but unreachable — so reporting an empty board would
// be a lie, and building a skeleton on top of it is that same lie with a directory tree attached.
//
// It used to get one, and the failure did not surface late — IT DID NOT SURFACE AT ALL.
// BuildSkeleton re-resolved the record directory itself and swallowed the error
// (`if recDir, err := record.RecordsDir(runDir); err == nil`), on the reasonable-sounding
// grounds that setup's job is to lay out a directory. Measured against the pre-migration code
// with this exact fixture: run-setup EXITED 0, wrote nothing to stderr, built the whole run —
// skeleton, mirrors, run-live marker, class join — and printed a summary that reads like a
// healthy setup. StageClassRegistry's "cannot resolve the record directory" arrived as a reason
// string inside a success report, on a line an operator scans past.
//
// Resolving ONCE into a record.Run moves the refusal to the front, where it is the exit code
// rather than a phrase, and where the operator has nothing to clean up.
func TestAnUnresolvableRunIsRefusedBeforeTheSkeletonExists(t *testing.T) {
	cfg, runDir := runCfg(t, reports(fmt.Sprint(record.EventSchema)))

	// The marker WITHOUT its pointer. Created here rather than by writing a record, because the
	// point is the half-state: the run says its record lives elsewhere and nothing can say where.
	if err := os.MkdirAll(runDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(runDir, ".records-elsewhere"), []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv(record.RecordRootEnv, "")

	var out, errb bytes.Buffer
	if code := Run(cfg, &out, &errb); code != 2 {
		t.Fatalf("exit %d, want 2 — an unreachable record must refuse the run:\nstdout:\n%s\nstderr:\n%s", code, out.String(), errb.String())
	}
	// NAMED, not merely non-zero: the operator's next move is to re-declare the root, and a
	// refusal that does not say so sends them looking at the topic or the model instead.
	if !strings.Contains(errb.String(), record.RecordRootEnv) {
		t.Errorf("the refusal must name %s, which is what fixes it:\n%s", record.RecordRootEnv, errb.String())
	}
	// THE ORDERING CLAIM, and the reason this test exists rather than a record-level one: the
	// skeleton's own directories must not be there. Refusing after building is the same failure
	// with extra steps.
	for _, d := range []string{"inputs", "blue", "red"} {
		if _, err := os.Stat(filepath.Join(runDir, d)); !os.IsNotExist(err) {
			t.Errorf("%s/ was created before the run was refused (%v)", d, err)
		}
	}
}
