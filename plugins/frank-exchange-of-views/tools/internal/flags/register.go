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
// long, and `--comment-file -` for anything that would have to survive shell quoting. The run's
// transcripts showed why the third matters: 68 commands carried escaped quotes and 37
// staged a temp file first, twice failing because the staged file was not there. A seat
// should never be choosing between saying the thing and escaping the thing.
//
// Attached through this function rather than by hand so no verb can register two of the
// three, or word them differently, or forget one.
func RegisterComment(c *cobra.Command) {
	c.Flags().String(Comment, "", DescComment)
	c.Flags().String(CommentFile, "", DescCommentFile)
}

// guardSingleStdinReader refuses an invocation in which two fields both claim stdin.
//
// There is ONE stdin. `--file -` with `--comment-file -` used to mean the first reader
// consumed everything and the second reported "stdin was empty" — an error that blames the
// wrong flag and sends a seat looking for a problem with its payload. Two claims on one
// stream is a mistake in the command, and the command should say so.
func guardSingleStdinReader(c *cobra.Command) error {
	payload, _ := c.Flags().GetString(File)
	comment, _ := c.Flags().GetString(CommentFile)
	if payload == "-" && comment == "-" {
		return fmt.Errorf("--%s - and --%s - both read stdin, and there is only one: pass one of them a path", File, CommentFile)
	}
	return nil
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
	useStdin := file == "-"
	if useStdin {
		file = ""
	}
	if err := guardSingleStdinReader(c); err != nil {
		return "", err
	}

	given := 0
	for _, on := range []bool{inline != "", file != "", useStdin} {
		if on {
			given++
		}
	}
	if given > 1 {
		return "", fmt.Errorf("--%s and --%s are two spellings of one field: pass exactly one", Comment, CommentFile)
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
			return "", fmt.Errorf("--%s -: %w", CommentFile, err)
		}
		s := strings.TrimRight(string(b), "\n")
		if s == "" {
			return "", fmt.Errorf("--%s - was given but stdin was empty", CommentFile)
		}
		return s, nil
	default:
		return inline, nil
	}
}

// ReadPayload resolves the prose payload: --text inline, --file from disk, or `--file -`
// from stdin. Same one-of rule and same reason as ReadComment.
func ReadPayload(c *cobra.Command, stdin io.Reader) (string, error) {
	if err := guardSingleStdinReader(c); err != nil {
		return "", err
	}
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

// Set writes a flag's value under a payload key, ONLY when it is non-empty.
//
// Setting it unconditionally is a trap worth naming: required-field validation asks
// whether the key is PRESENT, so writing an empty string makes a missing flag look
// supplied and the check passes on nothing. That regression was introduced while
// renaming --gap-id to --id and caught by the bench's own required-fields test.
func Set[P interface{ Set(string, any) P }](p P, key string, c *cobra.Command, flag string) {
	if v, _ := c.Flags().GetString(flag); v != "" {
		p.Set(key, v)
	}
}
