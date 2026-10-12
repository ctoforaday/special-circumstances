// Package goldenmods holds the one list of Go modules whose suites carry goldens.
//
// Two tools read it, and that is why it is a package rather than a var in either: `golden`
// runs each module's suite to verify goldens, and `check` runs the SAME command as that
// module's `:test` gate (plain `go test -count=1 ./...`, or scripts/testsplit for a module with
// a row in internal/testsplit's table). When one `check` invocation runs both, the
// golden leg is the test gate executed a second time — ~620s of a ~1550s local run
// re-deriving a result the run already has (#626). check can only subsume a leg it can
// prove is duplicate, and that proof is this shared record: if the lists lived in the two
// tools separately, the subsumption would be a hope about two files agreeing.
package goldenmods

// Modules are the Go modules whose suites carry goldens, relative to the repository root.
var Modules = []string{
	"plugins/frank-exchange-of-views/tools",
}
