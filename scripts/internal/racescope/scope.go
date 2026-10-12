package racescope

import (
	"fmt"
	"os/exec"
	"strings"
)

// Scope returns the module path of dir and the module-local packages in target's import graph,
// in `go list -deps` order — the same derivation race-nightly.yml performs with `go list -deps |
// grep`, so neither carrier holds a copy the other could drift from. The module path comes from
// the module itself (`go list -m`), not a constant here.
//
// AN EMPTY SCOPE IS AN ERROR, NEVER AN EMPTY GATE. `go test -race` with no packages would test
// the current directory and report a pass that measured almost nothing. A target that resolves
// to no module-local packages means the target or the module is broken, and the gate says so
// instead of going green.
func Scope(dir, target string) (module string, pkgs []string, err error) {
	mod := exec.Command("go", "list", "-m")
	mod.Dir = dir
	modOut, err := mod.Output()
	if err != nil {
		return "", nil, fmt.Errorf("scope of %s: go list -m in %s: %w", target, dir, err)
	}
	module = strings.TrimSpace(string(modOut))
	deps := exec.Command("go", "list", "-deps", target)
	deps.Dir = dir
	depsOut, err := deps.Output()
	if err != nil {
		return "", nil, fmt.Errorf("scope of %s: go list -deps in %s: %w", target, dir, err)
	}
	for _, p := range strings.Split(strings.TrimSpace(string(depsOut)), "\n") {
		if p == module || strings.HasPrefix(p, module+"/") {
			pkgs = append(pkgs, p)
		}
	}
	if len(pkgs) == 0 {
		return "", nil, fmt.Errorf("scope of %s has no packages under module %s — not measured, not clean", target, module)
	}
	return module, pkgs, nil
}
