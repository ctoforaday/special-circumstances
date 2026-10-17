package cli

import (
	"fmt"
	"os"
	"strings"
)

// refuseInsideProofRerun answers a call made by a proof's script while a checker runs it: the call
// is appended to the marker file the re-run named (seatenv.VarProofRerun) and refused, before any
// verb runs — a read included, because a proof that reads the debate's record measures the run as
// it stood for its author, which is not a computation anyone else can repeat.
//
// THE MARKER IS THE SIGNAL, THE REFUSAL IS FOR WHOEVER READS THE SCRIPT'S OUTPUT. A script may
// swallow the refusal; it cannot swallow the line this writes, and the re-run reads that line.
// A marker that cannot be written still refuses: the call is not run either way.
func refuseInsideProofRerun(marker string, argv []string) {
	line := strings.Join(append([]string{InvokedAs()}, argv...), " ")
	if f, err := os.OpenFile(marker, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0o600); err == nil {
		fmt.Fprintln(f, line)
		f.Close()
	}
	fmt.Fprintf(os.Stderr, "%s: refused — this call was made by a proof's script while a checker runs it. "+
		"A proof is a computation anyone can run again; a script that calls the debate's tool reads or writes the run as it stood for its author, which no checker can repeat, "+
		"so every such call is refused and the re-run records that the script cannot be checked\n", InvokedAs())
}
