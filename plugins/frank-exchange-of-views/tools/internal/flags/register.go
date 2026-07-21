package flags

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// RegisterPayload attaches the prose payload in its inline and file forms, with the
// canonical wording. `--reason-file -` reads stdin, which is the conventional spelling
// and costs no extra flag.
//
// Attached through this function rather than by hand so no verb can register one form and
// forget the other, or word them differently — the leak that let `close` ship with a
// hand-rolled --file and no --text at all.
func RegisterPayload(c *cobra.Command) {
	c.Flags().String(Reason, "", DescReason)
	c.Flags().String(ReasonFile, "", DescReasonFile)
}

// ReadPayload resolves the prose payload: --reason inline, --reason-file from disk, or
// `--reason-file -` from stdin. The two are one field, so passing both is refused rather
// than silently ranked — a seat that passes both should be told which one this verb would
// have dropped, not discover it in a projection three rounds later.
func ReadPayload(c *cobra.Command, stdin io.Reader) (string, error) {
	inline, _ := c.Flags().GetString(Reason)
	file, _ := c.Flags().GetString(ReasonFile)
	if inline != "" && file != "" {
		return "", fmt.Errorf("--%s and --%s are two spellings of one payload: pass exactly one", Reason, ReasonFile)
	}
	if file == "-" {
		b, err := io.ReadAll(stdin)
		if err != nil {
			return "", fmt.Errorf("--%s -: %w", ReasonFile, err)
		}
		s := strings.TrimRight(string(b), "\n")
		if s == "" {
			return "", fmt.Errorf("--%s - was given but stdin was empty", ReasonFile)
		}
		return s, nil
	}
	if file != "" {
		b, err := os.ReadFile(file)
		if err != nil {
			return "", fmt.Errorf("--%s: %w", ReasonFile, err)
		}
		return strings.TrimRight(string(b), "\n"), nil
	}
	return inline, nil
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
