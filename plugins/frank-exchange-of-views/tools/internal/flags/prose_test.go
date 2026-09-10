package flags

import (
	"errors"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// THE PROSE CHANNEL'S CONTRACT, TESTED ONCE, HERE.
//
// Before this type the contract lived in a resolver function and was exercised only through
// whichever verbs a CLI test happened to drive — eight of fifty-one, as it turned out, with all
// three verbs that got it wrong outside the eight. A contract tested per-caller is tested
// wherever someone remembered.
//
// The heredoc case is worth naming: a seat following the quoting rule captures a paragraph with
// `X=$(cat <<'EOF' … EOF)`, and a variable filled some other way can carry the shell's trailing
// newline. The seat should not have to know that.
func proseCmd(t *testing.T, args ...string) (*cobra.Command, *Prose) {
	t.Helper()
	var p Prose
	c := &cobra.Command{Use: "verb", RunE: func(*cobra.Command, []string) error { return nil }}
	p.Register(c)
	c.SetArgs(args)
	c.SetOut(nil)
	if err := c.Execute(); err != nil {
		t.Fatalf("parse %v: %v", args, err)
	}
	return c, &p
}

func TestProseInline(t *testing.T) {
	_, p := proseCmd(t, "--reason", "a paragraph")
	if got := p.Read(); got != "a paragraph" {
		t.Fatalf("got %q", got)
	}
	if p.String() != p.Read() {
		t.Errorf("String() is %q, want the resolved value %q", p.String(), p.Read())
	}
}

// THE HEREDOC. A paragraph captured from a heredoc can arrive with the shell's trailing newline
// still on it.
func TestProseTrimsAHeredocsTrailingNewline(t *testing.T) {
	const body = "line one\n\nline two, after a blank\n"
	_, p := proseCmd(t, "--reason", body)
	if got, want := p.Read(), "line one\n\nline two, after a blank"; got != want {
		t.Fatalf("got %q, want %q — interior blank lines are the seat's, the trailing one is the shell's", got, want)
	}
}

func TestProseOmittedIsEmptyNotAnError(t *testing.T) {
	c, p := proseCmd(t)
	if got := p.Read(); got != "" {
		t.Fatalf("got %q — a verb whose prose is optional must be able to omit it", got)
	}
	got, err := ReadPayload(c)
	if err != nil || got != "" {
		t.Fatalf("ReadPayload got %q, %v — a registered, omitted channel is empty prose, not an error", got, err)
	}
}

// AND READING A CHANNEL NOBODY REGISTERED IS LOUD.
//
// ReadPayload used GetString and discarded the error, so this case returned ("", nil): a verb
// that skipped registration got empty prose and no complaint, and the write was refused for a
// field the seat believed it had supplied.
func TestReadingAnUnregisteredChannelIsAnError(t *testing.T) {
	c := &cobra.Command{Use: "bare"}
	_, err := ReadPayload(c)
	if err == nil {
		t.Fatal("a command with no prose channel resolved to empty prose and no error — the resolver's own plausible zero")
	}
	if !strings.Contains(err.Error(), "never registered") {
		t.Errorf("the error must say the channel was never registered, not that the prose was empty: %v", err)
	}
	_ = errors.Unwrap(err)
}

// RegisterRequired makes the channel mandatory: --reason is accepted, a bare call is refused.
func TestRegisterRequiredRefusesABareCall(t *testing.T) {
	var p Prose
	c := &cobra.Command{Use: "verb", RunE: func(*cobra.Command, []string) error { return nil }}
	p.RegisterRequired(c)
	c.SetArgs([]string{"--reason", "inline"})
	if err := c.Execute(); err != nil {
		t.Errorf("--reason was refused by a required channel: %v", err)
	}
	var q Prose
	c = &cobra.Command{Use: "verb", RunE: func(*cobra.Command, []string) error { return nil }}
	q.RegisterRequired(c)
	c.SetArgs(nil)
	c.SilenceErrors, c.SilenceUsage = true, true
	if err := c.Execute(); err == nil {
		t.Error("a required channel accepted a bare call, so the duty it gates can be discharged with nothing")
	}
}
