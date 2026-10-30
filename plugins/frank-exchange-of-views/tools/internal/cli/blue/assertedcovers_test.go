package blue

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
)

// computeWrite and citeWrite are what each verb's keyed retry states, built by the verb's own writer,
// with --reason read the way the verb reads it.
func computeWrite(c *cobra.Command) proto.Message {
	why, _ := seat.Reason(c)
	return computeRetry(c, why)
}

func citeWrite(c *cobra.Command) proto.Message {
	why, _ := seat.Reason(c)
	return citeRetry(c, why)
}

// A KEYED RETRY RESTATES EVERY FIELD ITS VERB RECORDS FROM A FLAG. Each flag of compute and cite is
// driven through the verb's own writer, and a field it moves that the asserted set does not list is
// a retry that could state another act and be answered as the stored one.
func TestTheAssertedSetsCoverEveryRecordedFlag(t *testing.T) {
	for name, c := range map[string]struct {
		newCmd   func() *cobra.Command
		write    func(*cobra.Command) proto.Message
		asserted []seat.Asserted
	}{
		"compute": {newCompute, computeWrite, computeAsserted},
		"cite":    {newCite, citeWrite, citeAsserted},
	} {
		t.Run(name, func(t *testing.T) {
			if err := seat.AssertedCovers(c.newCmd, c.write, c.asserted); err != nil {
				t.Error(err)
			}
		})
	}
}

// AND IT NAMES THE FLAG: dropping any one entry from either set fails, naming the flag it covered.
func TestDroppingAnAssertedEntryIsNamed(t *testing.T) {
	for name, c := range map[string]struct {
		newCmd   func() *cobra.Command
		write    func(*cobra.Command) proto.Message
		asserted []seat.Asserted
	}{
		"compute": {newCompute, computeWrite, computeAsserted},
		"cite":    {newCite, citeWrite, citeAsserted},
	} {
		for i, a := range c.asserted {
			dropped := append(append([]seat.Asserted{}, c.asserted[:i]...), c.asserted[i+1:]...)
			err := seat.AssertedCovers(c.newCmd, c.write, dropped)
			if err == nil {
				t.Errorf("%s without {%s %s}: the check passed", name, a.Flag, a.Field)
				continue
			}
			if !strings.Contains(err.Error(), "--"+a.Flag+" moves "+string(a.Field)) {
				t.Errorf("%s without {%s %s}: %v", name, a.Flag, a.Field, err)
			}
		}
	}
}
