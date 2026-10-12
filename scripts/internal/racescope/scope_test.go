package racescope

import (
	"strings"
	"testing"
)

// The scope is derived in the gate's own module, keeps only that module's packages, and a
// target that resolves to nothing is an ERROR — a race gate over no packages would pass while
// measuring nothing.
func TestScopeDerivesTheModuleLocalGraph(t *testing.T) {
	// A test binary runs in its package directory, so the module root is two levels up.
	module, pkgs, err := Scope("../..", "./check")
	if err != nil {
		t.Fatal(err)
	}
	if module != "github.com/ctoforaday/special-circumstances/scripts" {
		t.Errorf("module = %q", module)
	}
	if len(pkgs) < 2 {
		t.Fatalf("scope = %v, want check and the internal packages it imports", pkgs)
	}
	for _, p := range pkgs {
		if !strings.HasPrefix(p, module+"/") {
			t.Errorf("the scope leaked a package outside the module: %s", p)
		}
	}
	if _, _, err := Scope("../..", "./no-such-target"); err == nil {
		t.Error("a target go list cannot resolve gave a scope without error — that silence would race nothing and pass")
	}
}

func TestLookupMatchesWholeNames(t *testing.T) {
	if _, ok := Lookup("time.AfterFunc"); !ok {
		t.Error("time.AfterFunc is a row and was not found")
	}
	for _, near := range []string{"time.After", "mytime.AfterFunc", "(*net/http.Server).Serv", ""} {
		if _, ok := Lookup(near); ok {
			t.Errorf("%q matched a row; a callee is matched by its whole name", near)
		}
	}
}
