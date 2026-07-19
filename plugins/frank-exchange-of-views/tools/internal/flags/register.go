package flags

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// RegisterComment attaches the universal free-text field in all three forms.
//
// Three spellings, one destination. --comment for a phrase, --comment-file for anything
// long, --comment-stdin for anything that would have to survive shell quoting. The run's
// transcripts showed why the third matters: 68 commands carried escaped quotes and 37
// staged a temp file first, twice failing because the staged file was not there. A seat
// should never be choosing between saying the thing and escaping the thing.
//
// Attached through this function rather than by hand so no verb can register two of the
// three, or word them differently, or forget one.
func RegisterComment(c *cobra.Command) {
	c.Flags().String(Comment, "", DescComment)
	c.Flags().String(CommentFile, "", DescCommentFile)
	c.Flags().Bool(CommentStdin, false, DescCommentStdin)
}

// RegisterPayload attaches the prose payload in its file and inline forms, with the
// canonical wording. `--file -` reads stdin, which is the conventional spelling and
// costs no extra flag.
func RegisterPayload(c *cobra.Command) {
	c.Flags().String(File, "", DescFile)
	c.Flags().String(Text, "", DescText)
}

// Comment resolves the free-text field from whichever form was used.
//
// Exactly one may be given: silently preferring one over another would mean a seat that
// passed both --comment and --comment-file loses half of what it said, and losing what a
// seat tried to record is the failure this whole field exists to prevent.
func ReadComment(c *cobra.Command, stdin io.Reader) (string, error) {
	inline, _ := c.Flags().GetString(Comment)
	file, _ := c.Flags().GetString(CommentFile)
	useStdin, _ := c.Flags().GetBool(CommentStdin)

	given := 0
	for _, on := range []bool{inline != "", file != "", useStdin} {
		if on {
			given++
		}
	}
	if given > 1 {
		return "", fmt.Errorf("--%s, --%s and --%s are three spellings of one field: pass exactly one", Comment, CommentFile, CommentStdin)
	}
	switch {
	case file != "":
		b, err := os.ReadFile(file)
		if err != nil {
			return "", fmt.Errorf("--%s: %w", CommentFile, err)
		}
		return strings.TrimRight(string(b), "\n"), nil
	case useStdin:
		b, err := io.ReadAll(stdin)
		if err != nil {
			return "", fmt.Errorf("--%s: %w", CommentStdin, err)
		}
		s := strings.TrimRight(string(b), "\n")
		if s == "" {
			return "", fmt.Errorf("--%s was given but stdin was empty", CommentStdin)
		}
		return s, nil
	default:
		return inline, nil
	}
}

// ReadPayload resolves the prose payload: --text inline, --file from disk, or `--file -`
// from stdin. Same one-of rule and same reason as ReadComment.
func ReadPayload(c *cobra.Command, stdin io.Reader) (string, error) {
	text, _ := c.Flags().GetString(Text)
	file, _ := c.Flags().GetString(File)
	if text != "" && file != "" {
		return "", fmt.Errorf("--%s and --%s are two spellings of one payload: pass exactly one", Text, File)
	}
	if file == "-" {
		b, err := io.ReadAll(stdin)
		if err != nil {
			return "", fmt.Errorf("--%s -: %w", File, err)
		}
		s := strings.TrimRight(string(b), "\n")
		if s == "" {
			return "", fmt.Errorf("--%s - was given but stdin was empty", File)
		}
		return s, nil
	}
	if file != "" {
		b, err := os.ReadFile(file)
		if err != nil {
			return "", fmt.Errorf("--%s: %w", File, err)
		}
		return strings.TrimRight(string(b), "\n"), nil
	}
	return text, nil
}

// AliasString registers a legacy spelling of a canonical flag: same destination, hidden
// from help so the vocabulary taught is the shared one.
//
// The audit found three concepts with two names each — the gap being acted on was --id on
// four verbs and --gap-id on the bench's; the fate was --as on five verbs and
// --disposition on the bench's; the justification was --basis on four and --evidence on
// blue's dispute. Each is individually sensible and collectively a dialect, and a seat
// generalises from whichever verb it learned first. Renaming outright would break every
// prompt at once, so the canonical name is taught and the old one keeps working.
func AliasString(c *cobra.Command, canonical, legacy, usage string) {
	c.Flags().String(canonical, "", usage)
	c.Flags().String(legacy, "", "deprecated alias for --"+canonical)
	_ = c.Flags().MarkHidden(legacy)
}

// Either reads a canonical flag, falling back to its legacy alias.
func Either(c *cobra.Command, canonical, legacy string) string {
	if v, _ := c.Flags().GetString(canonical); v != "" {
		return v
	}
	v, _ := c.Flags().GetString(legacy)
	return v
}

// SetEither writes the canonical-or-legacy value under a payload key, ONLY when it is
// non-empty.
//
// Setting it unconditionally is a trap worth naming: required-field validation asks
// whether the key is PRESENT, so writing an empty string makes a missing flag look
// supplied and the check passes on nothing. That regression was introduced while
// canonicalising --gap-id to --id and caught by the bench's own required-fields test.
// Callers get the guard for free rather than having to remember it.
func SetEither[P interface{ Set(string, any) P }](p P, key string, c *cobra.Command, canonical, legacy string) {
	if v := Either(c, canonical, legacy); v != "" {
		p.Set(key, v)
	}
}
