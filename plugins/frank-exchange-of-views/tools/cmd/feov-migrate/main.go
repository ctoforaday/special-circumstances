// Command feov-migrate is the operator's replay-as-migration: an old run's record re-driven
// through the current write path into a fresh run. Everything lives in internal/migratecli; this
// file hands over and exits with its code.
package main

import (
	"os"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/migratecli"
)

func main() {
	os.Exit(migratecli.Main(os.Args[1:], os.Stdout, os.Stderr))
}
