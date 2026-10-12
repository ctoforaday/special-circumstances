package serial

import (
	"fmt"
	"os"
	"testing"
)

// state is package-level and unsynchronised: these tests are serial by construction, which is
// what earns a package a row.
var state int

func TestSplitSeesTheCallersEnvironment(t *testing.T) {
	if os.Getenv("TESTSPLIT_FIXTURE") != "1" {
		t.Fatal("TESTSPLIT_FIXTURE did not reach a process of a split package")
	}
}

func TestSplitRunsInThePackageDirectory(t *testing.T) {
	if _, err := os.Stat("testdata/marker.txt"); err != nil {
		t.Fatalf("the process is not in its package directory: %v", err)
	}
}

func TestA(t *testing.T) { state++ }
func TestB(t *testing.T) { state++ }
func TestC(t *testing.T) { state++ }
func TestD(t *testing.T) { t.Skip("always") }

// Prints without a trailing newline: the `--- PASS` line that follows is glued to it, and the
// verdict still arrives as an event.
func TestPrintsWithoutANewline(t *testing.T) { fmt.Print("no newline") }

func TestHasSubtests(t *testing.T) {
	t.Run("one", func(t *testing.T) {})
	t.Run("two", func(t *testing.T) {})
}

func FuzzListed(f *testing.F) {
	f.Add(1)
	f.Fuzz(func(t *testing.T, n int) {})
}

func Example() {
	fmt.Println("listed")
	// Output: listed
}
