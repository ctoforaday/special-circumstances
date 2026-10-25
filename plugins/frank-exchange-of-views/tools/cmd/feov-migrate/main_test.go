package main

import (
	"errors"
	"os/exec"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/goldentest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/migratecli"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/testbuild"
)

func TestMain(m *testing.M) { testbuild.Main(m) }

// THE HELP IS THE BUILT BINARY'S, pinned whole. It is the operator's only description of the tool,
// and a test of migratecli alone would pass with this file's call to it deleted.
func TestTheHelpPage(t *testing.T) {
	out, err := exec.Command(testbuild.Binary(t, "feov-migrate"), "--help").CombinedOutput()
	if err != nil {
		t.Fatalf("feov-migrate --help: %v\n%s", err, out)
	}
	goldentest.Assert(t, "help", string(out))
}

// A REFUSAL IS THE PROCESS'S EXIT CODE, not only migratecli's return value: main exits with it.
func TestARefusalExitsTwo(t *testing.T) {
	cmd := exec.Command(testbuild.Binary(t, "feov-migrate"))
	var stderr strings.Builder
	cmd.Stderr = &stderr
	err := cmd.Run()
	var exit *exec.ExitError
	if !errors.As(err, &exit) || exit.ExitCode() != 2 {
		t.Fatalf("feov-migrate with no flags: %v, want exit 2", err)
	}
	if want := migratecli.Name + ": --from and --to are both required"; !strings.HasPrefix(stderr.String(), want) {
		t.Errorf("stderr = %q, want it to open %q", stderr.String(), want)
	}
}
