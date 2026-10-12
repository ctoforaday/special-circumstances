package plain

import (
	"os"
	"testing"
)

func TestPlainSeesTheCallersEnvironment(t *testing.T) {
	if os.Getenv("TESTSPLIT_FIXTURE") != "1" {
		t.Fatal("TESTSPLIT_FIXTURE did not reach a package on the plain path")
	}
}

func TestPlainSkips(t *testing.T) { t.Skip("always") }
