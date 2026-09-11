package cli

import (
	"testing"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
)

// EVERY KEY NAMES A PROSE VERB, AND THAT VERB RENDERS IT. seat.ReasonIs is keyed by verb name; a key
// no verb carries is a description nobody reads, and a verb renamed out from under its key would
// quietly fall back to the generic --reason text this table exists to replace.
func TestReasonIsNamesOnlyProseVerbs(t *testing.T) {
	seen := map[string]bool{}
	for role, root := range AllRoots() {
		walk(root, func(c *cobra.Command, path []string) {
			want, ok := seat.ReasonIs[c.Name()]
			if !ok || flags.ProseOf(c) == nil {
				return // not a prose verb
			}
			f := c.Flags().Lookup(flags.Reason)
			if f == nil {
				t.Errorf("%s `%s` is a ReasonIs key but registers no --%s", role, c.CommandPath(), flags.Reason)
				return
			}
			if f.Usage != want && f.Usage != "REQUIRED — "+want {
				t.Errorf("%s `%s --%s` renders %q, not its ReasonIs text", role, c.CommandPath(), flags.Reason, f.Usage)
			}
			seen[c.Name()] = true
		})
	}
	for k := range seat.ReasonIs {
		if !seen[k] {
			t.Errorf("seat.ReasonIs names %q, which no seat surface carries", k)
		}
	}
}
