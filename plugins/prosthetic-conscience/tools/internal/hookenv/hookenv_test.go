package hookenv

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestProjectDirPrefersTheEnvironmentAndFallsBackToThePayload(t *testing.T) {
	for _, c := range []struct{ env, cwd, want string }{
		{"/from/env", "/from/payload", "/from/env"},
		{"", "/from/payload", "/from/payload"},
		{"/from/env", "", "/from/env"},
		{"", "", ""},
	} {
		if got := ProjectDir(c.env, c.cwd); got != c.want {
			t.Errorf("ProjectDir(%q, %q) = %q, want %q", c.env, c.cwd, got, c.want)
		}
	}
}

// The one fallback that must NOT exist. os.Getwd() would turn "I do not know where the
// project is" into "I will write somewhere", and the somewhere is unrelated to the
// session — the exact defect sc-recall-index was already patched for, in one binary.
func TestUnknownStaysUnknownRatherThanBecomingTheWorkingDirectory(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if got := ProjectDir("", ""); got != "" {
		t.Fatalf("ProjectDir with nothing to go on returned %q", got)
	}
	if got := ProjectDir("", ""); got == wd {
		t.Errorf("resolved to the working directory %q — that is the fallback this package exists to refuse", wd)
	}
}

// A hook that no-ops silently is indistinguishable from a hook with nothing to do. That
// ambiguity is how the class survived: sc-checkpoint-restore returned 0 and printed
// nothing, and the session read it as "this project has no checkpoint".
func TestAnUnresolvedRootSaysSoAndStopsTheCaller(t *testing.T) {
	var errOut bytes.Buffer
	if Explain("", &errOut, "sc-probe") {
		t.Error("Explain told the caller to continue with no project root")
	}
	msg := errOut.String()
	if !strings.Contains(msg, "sc-probe") {
		t.Errorf("the line does not name the hook: %q", msg)
	}
	for _, want := range []string{"CLAUDE_PROJECT_DIR", "cwd"} {
		if !strings.Contains(msg, want) {
			t.Errorf("the line does not name %s, so a reader cannot tell which input was missing: %q", want, msg)
		}
	}

	errOut.Reset()
	if !Explain("/somewhere", &errOut, "sc-probe") {
		t.Error("Explain stopped a caller that HAS a project root")
	}
	if errOut.Len() != 0 {
		t.Errorf("Explain spoke on the healthy path: %q", errOut.String())
	}
}

// THE CLASS, HELD CLOSED.
//
// Eight binaries read CLAUDE_PROJECT_DIR and, finding it empty, each did something quiet
// and different — one of them writing state into whatever directory it started in. That
// one was found and fixed IN PLACE; the other seven kept the hole, which is the same
// instance-not-class failure this suite keeps recording about itself.
//
// This walks the real source. A new hook that reads the environment variable directly,
// instead of resolving through this package, fails here.
func TestNoHookReadsTheEnvironmentVariableWithoutResolvingThroughThisPackage(t *testing.T) {
	cmds, err := filepath.Glob(filepath.Join("..", "..", "cmd", "*", "main.go"))
	if err != nil || len(cmds) == 0 {
		t.Fatalf("found no hook sources to walk (%v) — a broken walk would pass this test silently forever", err)
	}

	// The env var may be read ONCE per binary, in main(), as the first input to
	// ProjectDir. Anywhere else is a second resolution path, which is what drifted.
	readsEnv := regexp.MustCompile(`os\.Getenv\("CLAUDE_PROJECT_DIR"\)`)
	var offenders []string
	for _, path := range cmds {
		src, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		body := string(src)
		reads := len(readsEnv.FindAllString(body, -1))
		if reads == 0 {
			continue // doctor and the pure gates need no project root
		}
		uses := strings.Contains(body, "hookenv.ProjectDir(")
		switch {
		case !uses:
			offenders = append(offenders, path+": reads CLAUDE_PROJECT_DIR but never resolves through hookenv.ProjectDir")
		case reads > 1:
			offenders = append(offenders, path+": reads CLAUDE_PROJECT_DIR more than once — a second resolution path is a second version of the rule")
		}
	}
	if len(offenders) > 0 {
		t.Errorf("hooks resolving the project root on their own:\n  %s\n\nCLAUDE_PROJECT_DIR is not the only statement of where the project is: the hook payload carries `cwd` as well, and a hook that reads only the environment goes silent when it is unset — silence that is indistinguishable from having nothing to do. Resolve through hookenv.ProjectDir(env, in.CWD) and report an unresolved root with hookenv.Explain.",
			strings.Join(offenders, "\n  "))
	}
}
