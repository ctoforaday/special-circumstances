package seat

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
)

// ONE RENDERER FOR EVERY REFUSAL. A handler's error, the record's, cobra's own (an unknown flag, a
// missing required flag, a bad value, wrong arguments, a flag group), a wrong identity and a command
// that is not on the surface all leave the tool through RenderRefusal, in one of two forms picked
// by the argv: the JSON envelope on stdout, or one line under the tool's name on stderr.
//
// WHO IS REFUSED AND AT WHAT COMES FROM THE COMMAND, NEVER FROM THE SENTENCE. The role is the
// tree's annotation, the verb and the path are the resolved command's own, and the code is the
// error's. A diagnosis therefore starts at what is wrong: a message that names its own role or
// path says it a second time, beside the renderer's.

// Refusal is a diagnosis together with the help page that answers it, held as FIELDS so each form
// can carry what suits it: the human line prints the page under the diagnosis, and the JSON
// envelope — one string a consumer reads as one sentence — names the command whose `--help` prints
// that page.
type Refusal struct {
	// Diagnosis is complete on its own: what was wrong and what to do, referring to no text after it.
	Diagnosis string
	// Help is the page of the command at Path, or "" when that command renders none.
	Help string
	// Path is the command path whose page Help is — the leaf for a flag refusal, a group or the
	// root for a command that was not found under it.
	Path string
	// Lists says what that page holds, for the pointer the JSON form carries in its place.
	Lists string
	// Cause is the error this refusal answers, when one exists beneath it; its code is the refusal's.
	Cause error
}

func (r *Refusal) Unwrap() error { return r.Cause }

func (r *Refusal) Error() string {
	if strings.TrimSpace(r.Help) == "" {
		return r.Diagnosis + "\n\n" + r.pointer()
	}
	return r.Diagnosis + "\n\n" + r.Help
}

// pointer names the page instead of printing it.
func (r *Refusal) pointer() string {
	return fmt.Sprintf("Run `%s --help` for %s.", r.Path, r.Lists)
}

// PathOf is a command's path below the root — `finding`, `motion grade file` — and "" for the root
// itself, which is what resolves when no command did.
func PathOf(c *cobra.Command) string {
	if c == nil || c == c.Root() {
		return ""
	}
	return strings.TrimPrefix(c.CommandPath(), c.Root().Name()+" ")
}

// Diagnosis is the refusal's own sentence: the whole message, or for a Refusal the part in front
// of its help page. The log entry the tool writes for a refusal and both rendered forms read it
// here, so what is logged is what the seat was told.
func Diagnosis(err error) string {
	var r *Refusal
	if errors.As(err, &r) {
		return r.Diagnosis
	}
	return err.Error()
}

// JSONRequested reads --json off the RAW ARGV.
//
// It cannot come from the parsed flag set: a refusal may be raised where parsing FAILED, so the
// flag may never have been bound. Scanning argv is the only source that survives that, and the
// flag is a plain boolean with no value form to confuse it.
func JSONRequested(argv []string) bool {
	for _, a := range argv {
		if a == "--"+flags.JSON || a == "--"+flags.JSON+"=true" {
			return true
		}
	}
	return false
}

// EnvelopeOf is the envelope for a refusal of the command that resolved.
func EnvelopeOf(leaf *cobra.Command, err error) feov.RefusalEnvelope {
	env := feov.RefusalEnvelope{Role: roleOf(leaf), Code: feov.CodeOf(err), Error: err.Error()}
	if leaf != nil && leaf != leaf.Root() {
		env.Verb = leaf.Name()
	}
	var r *Refusal
	if errors.As(err, &r) {
		env.Error = r.Diagnosis + "\n\n" + r.pointer()
	}
	return env
}

// RenderRefusal writes a refusal in the form the argv asked for: the envelope on stdout under
// --json, otherwise `<tool>: <path>: <diagnosis>` on stderr, the path printed once and left out
// when no command resolved.
func RenderRefusal(stdout, stderr io.Writer, leaf *cobra.Command, argv []string, err error) {
	if JSONRequested(argv) {
		_ = json.NewEncoder(stdout).Encode(EnvelopeOf(leaf, err))
		return
	}
	head := "feov-record"
	if leaf != nil {
		head = leaf.Root().Name()
	}
	if p := PathOf(leaf); p != "" {
		head += ": " + p
	}
	fmt.Fprintf(stderr, "%s: %v\n", head, err)
}
