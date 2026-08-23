package flags

import (
	"errors"
	"os"
	"path/filepath"
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
// The heredoc case is the one that cost a seat its sitting and is worth naming: `--reason-file -`
// with `<< 'EOF'` is how a seat passes a paragraph, the shell appends a newline, and the seat
// should not have to know that.
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
	got, err := p.Read(strings.NewReader(""))
	if err != nil || got != "a paragraph" {
		t.Fatalf("got %q, %v", got, err)
	}
}

func TestProseFromAFile(t *testing.T) {
	f := filepath.Join(t.TempDir(), "r.md")
	if err := os.WriteFile(f, []byte("from a file\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, p := proseCmd(t, "--reason-file", f)
	got, err := p.Read(strings.NewReader(""))
	if err != nil || got != "from a file" {
		t.Fatalf("got %q, %v — the trailing newline a file keeps must not reach the record", got, err)
	}
}

// THE HEREDOC. `cmd --reason-file - << 'EOF' … EOF` is the form a seat reaches for with a
// paragraph, and the shell always leaves a trailing newline behind it.
func TestProseFromAHeredocOnStdin(t *testing.T) {
	const body = "line one\n\nline two, after a blank\n"
	_, p := proseCmd(t, "--reason-file", "-")
	got, err := p.Read(strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	if want := "line one\n\nline two, after a blank"; got != want {
		t.Fatalf("got %q, want %q — interior blank lines are the seat's, the trailing one is the shell's", got, want)
	}
}

// STDIN IS A STREAM AND IS READ ONCE. A verb that resolved the channel twice — once for one
// payload key and once for another — got the prose the first time and nothing the second.
func TestProseFromStdinIsStableAcrossReads(t *testing.T) {
	_, p := proseCmd(t, "--reason-file", "-")
	first, err := p.Read(strings.NewReader("read me once"))
	if err != nil {
		t.Fatal(err)
	}
	second, err := p.Read(strings.NewReader(""))
	if err != nil {
		t.Fatal(err)
	}
	if first != second {
		t.Errorf("the channel yielded %q then %q — a verb filling two payload keys from it "+
			"would leave the second empty and be refused for a field the seat supplied", first, second)
	}
	if p.String() != first {
		t.Errorf("String() is %q, want the resolved value %q", p.String(), first)
	}
}

func TestProseRefusesBothSpellings(t *testing.T) {
	f := filepath.Join(t.TempDir(), "r.md")
	if err := os.WriteFile(f, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, p := proseCmd(t, "--reason", "inline", "--reason-file", f)
	if _, err := p.Read(strings.NewReader("")); err == nil {
		t.Fatal("both spellings accepted — one of them was silently dropped, and the seat cannot know which")
	}
}

func TestProseRefusesAnEmptyStdin(t *testing.T) {
	_, p := proseCmd(t, "--reason-file", "-")
	if _, err := p.Read(strings.NewReader("")); err == nil {
		t.Fatal("`--reason-file -` with nothing on stdin recorded an empty reason rather than refusing")
	}
}

func TestProseRefusesAMissingFile(t *testing.T) {
	_, p := proseCmd(t, "--reason-file", filepath.Join(t.TempDir(), "absent.md"))
	if _, err := p.Read(strings.NewReader("")); err == nil {
		t.Fatal("a staged file that is not there read as empty prose — measured twice in one run before the channel existed")
	}
}

func TestProseNeitherSpellingIsEmptyNotAnError(t *testing.T) {
	_, p := proseCmd(t)
	got, err := p.Read(strings.NewReader(""))
	if err != nil || got != "" {
		t.Fatalf("got %q, %v — a verb whose prose is optional must be able to omit it", got, err)
	}
}

// AND READING A CHANNEL NOBODY REGISTERED IS LOUD.
//
// ReadPayload used GetString and discarded the error, so this case returned ("", nil): a verb
// that skipped registration got empty prose and no complaint, and the write was refused for a
// field the seat believed it had supplied.
func TestReadingAnUnregisteredChannelIsAnError(t *testing.T) {
	c := &cobra.Command{Use: "bare"}
	_, err := ReadPayload(c, strings.NewReader("something"))
	if err == nil {
		t.Fatal("a command with no prose channel resolved to empty prose and no error — the resolver's own plausible zero")
	}
	if !strings.Contains(err.Error(), "never registered") {
		t.Errorf("the error must say the channel was never registered, not that the prose was empty: %v", err)
	}
	_ = errors.Unwrap(err)
}

// RegisterRequired makes the channel mandatory THROUGH EITHER SPELLING. MarkFlagRequired names
// one flag and would refuse the file form, which is what a hand-registered `spot-check` did.
func TestRegisterRequiredAcceptsEitherSpelling(t *testing.T) {
	for _, args := range [][]string{{"--reason", "inline"}, {"--reason-file", "-"}} {
		var p Prose
		c := &cobra.Command{Use: "verb", RunE: func(*cobra.Command, []string) error { return nil }}
		p.RegisterRequired(c)
		c.SetArgs(args)
		if err := c.Execute(); err != nil {
			t.Errorf("%v was refused by a required channel that should accept either spelling: %v", args, err)
		}
	}
	var p Prose
	c := &cobra.Command{Use: "verb", RunE: func(*cobra.Command, []string) error { return nil }}
	p.RegisterRequired(c)
	c.SetArgs(nil)
	c.SilenceErrors, c.SilenceUsage = true, true
	if err := c.Execute(); err == nil {
		t.Error("a required channel accepted a bare call, so the duty it gates can be discharged with nothing")
	}
}
