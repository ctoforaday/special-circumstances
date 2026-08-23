// Command sc-stop is a thin shim for the Stop nudge.
//
// One binary per EVENT (#201 step 3). It is the ONLY writer of the nudge's state, which
// is why the band record lives in its own file rather than beside the gauge's: the gauge
// file has four writing binaries, and a json.Marshal of one struct erases the others'
// keys — an erased band being a re-emission on Stop, which is the loop.
package main

import (
	"os"

	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/stopnudge"
)

func main() { os.Exit(stopnudge.Main()) }
