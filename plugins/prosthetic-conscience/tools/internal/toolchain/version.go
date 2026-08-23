package toolchain

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// CHECK_CMD WAS DECLARED ON EVERY TOOL AND WAS NEVER A CHECK.
//
// Probe resolved presence with LookPath on the FIRST TOKEN and nothing else; the rest of the
// string ran only to PRINT a version beside a row already decided ✓. So a tool present but too
// old to build the project reported ✓ with the disqualifying number rendered helpfully next to
// it. Measured (#521): a machine carrying go1.22.2 against a `go 1.25` directive would have
// shown `go ✓ go version go1.22.2 linux/amd64` — had `go` been declared at all, which it was
// not.
//
// That is facts-are-fields clause 3. A presence-only probe cannot tell USABLE from PRESENT BUT
// UNUSABLE, and both render as the healthy string, so the miss is the same bytes as the pass.
//
// MinVersion is what makes check_cmd load-bearing: declare it and the command runs, its output
// is parsed, and a tool below the line is its OWN state — not ✓, and not ✗ either, because the
// remedies differ. Missing means install; too old means upgrade. Telling an operator to install
// what they already have is the same wrong-place failure as telling a seat to pass a flag it
// already passed.

// CompareVersions returns -1, 0 or +1 for a < b, a == b, a > b, comparing dot-separated numeric
// components and treating an absent component as zero — so "1.25" and "1.25.0" are equal, which
// matters because a Go directive is written `1.25` and no release is ever called that.
//
// Non-numeric trailing text on a component (a "-rc1", a "+meta") is ignored rather than ranked:
// ordering pre-releases correctly is a whole specification, and guessing at it here would put a
// confident wrong answer where an honest coarse one belongs.
func CompareVersions(a, b string) int {
	as, bs := strings.Split(a, "."), strings.Split(b, ".")
	n := len(as)
	if len(bs) > n {
		n = len(bs)
	}
	for i := 0; i < n; i++ {
		if d := numAt(as, i) - numAt(bs, i); d != 0 {
			if d < 0 {
				return -1
			}
			return 1
		}
	}
	return 0
}

func numAt(parts []string, i int) int {
	if i >= len(parts) {
		return 0
	}
	s := parts[i]
	end := 0
	for end < len(s) && s[end] >= '0' && s[end] <= '9' {
		end++
	}
	n, _ := strconv.Atoi(s[:end])
	return n
}

// ExtractVersion pulls the first dotted numeric run out of a tool's version output.
//
// The shapes are not uniform and never will be: `go version go1.22.2 linux/amd64`,
// `git version 2.43.0`, `gh version 2.97.0 (2026-07-31)`, `jq-1.7`, `Python 3.11.15`. A parser
// per tool is a table that goes stale silently; taking the first dotted run reads all of them
// and is honest about the one thing it cannot do — see the empty return, which callers MUST
// treat as "not measured" rather than as a failure to meet the minimum.
func ExtractVersion(out string) string {
	for i := 0; i < len(out); i++ {
		if out[i] < '0' || out[i] > '9' {
			continue
		}
		j, dots := i, 0
		for j < len(out) && (out[j] == '.' || (out[j] >= '0' && out[j] <= '9')) {
			if out[j] == '.' {
				// A trailing dot is sentence punctuation, not a component.
				if j+1 >= len(out) || out[j+1] < '0' || out[j+1] > '9' {
					break
				}
				dots++
			}
			j++
		}
		if dots >= 1 {
			return out[i:j]
		}
		i = j // skip the bare integer whole rather than restarting inside it
	}
	return ""
}

// versionOutput runs a check command and returns its combined output. It is a variable so a
// test can drive the parsing and the comparison without depending on which tools happen to be
// installed where the test runs.
var versionOutput = func(checkCmd, dir string) (string, error) {
	fields := strings.Fields(checkCmd)
	if len(fields) == 0 {
		return "", exec.ErrNotFound
	}
	cmd := exec.Command(fields[0], fields[1:]...)
	cmd.Dir = dir // empty means the caller's own cwd, which is the old behaviour
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// measure fills in a found tool's version state. It runs the check command ONLY when a minimum
// was declared, so declaring one is the whole cost and every other tool stays a single LookPath.
func measure(s *Status, baseDir string) {
	if s.MinVersion == "" || !s.Found {
		return
	}
	dir := ""
	if s.CheckDir != "" && baseDir != "" {
		dir = filepath.Join(baseDir, s.CheckDir)
		// A CheckDir that is not there makes the answer unknowable, not old. Running the
		// command from somewhere else would produce a confident number about the wrong place,
		// which is the defect this field exists to fix.
		if st, err := os.Stat(dir); err != nil || !st.IsDir() {
			s.VersionUnmeasured = "the declared check directory " + strconv.Quote(s.CheckDir) + " is not present, so the version could not be read where it matters"
			return
		}
	}
	out, err := versionOutput(s.CheckCmd, dir)
	s.VersionLine = strings.TrimSpace(firstLine(out))
	if err != nil && strings.TrimSpace(out) == "" {
		s.VersionUnmeasured = "the check command did not run: " + err.Error()
		return
	}
	v := ExtractVersion(out)
	if v == "" {
		// LOUD, not ✓. An unreadable version means the minimum is UNVERIFIED, and reporting
		// that as a pass is the exact collapse this field exists to end.
		s.VersionUnmeasured = "no version number found in " + strconv.Quote(s.VersionLine)
		return
	}
	s.Version = v
	s.TooOld = CompareVersions(v, s.MinVersion) < 0
}

func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i]
	}
	return s
}
