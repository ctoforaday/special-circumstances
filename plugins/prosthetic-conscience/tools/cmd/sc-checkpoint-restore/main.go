// Command sc-checkpoint-restore is a thin shim: the logic lives in internal/checkpointrestore so it can be
// tested, shared, and later composed with the other hooks on its event (#201).
package main

import (
	"os"

	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/checkpointrestore"
)

func main() { os.Exit(checkpointrestore.Main()) }
