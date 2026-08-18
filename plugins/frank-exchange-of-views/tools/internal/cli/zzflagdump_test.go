package cli

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func TestZZFlagDump(t *testing.T) {
	f, err := os.Create(os.Getenv("FLAGDUMP"))
	if err != nil {
		t.Skip("no FLAGDUMP")
	}
	defer f.Close()
	var walk func(c *cobra.Command, prefix string)
	walk = func(c *cobra.Command, prefix string) {
		for _, sub := range c.Commands() {
			name := sub.Name()
			if name == "help" || name == "completion" {
				continue
			}
			path := strings.TrimSpace(prefix + " " + name)
			if sub.HasSubCommands() {
				walk(sub, path)
				continue
			}
			sub.LocalFlags().VisitAll(func(fl *pflag.Flag) {
				req := ""
				if fl.Annotations[cobra.BashCompOneRequiredFlag] != nil {
					req = "REQ"
				}
				fmt.Fprintf(f, "%s\t%s\t%s\t%s\t%s\n", path, fl.Name, fl.Value.Type(), req, strings.ReplaceAll(fl.Usage, "\t", " "))
			})
		}
	}
	walk(newRoot(), "")
}
