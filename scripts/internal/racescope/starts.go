// Package racescope holds the two records the race gate is derived from: the SCOPE (a binary's
// module-local import graph) and the table of library calls that put caller code on another
// goroutine. `scripts/racescope` reads both; `scripts/check` reads the scope marker.
//
// Dev tooling for this repository only. Nothing here ships to an installing project.
package racescope

// Start is one library call that runs caller code on another goroutine, or makes code the
// caller wrote overlap with code that is already running. A `go` statement needs no row: the
// tool finds it by syntax.
//
// Callee is the callee's types.Func.FullName(), so a row is matched against type information
// and never against the text of a call.
type Start struct {
	Callee string
	Why    string
}

// Table is every such call the tool knows. A library that runs our callback on its own
// goroutine and has no row here is NOT raced on a pull request; the nightly whole-graph run is
// where it surfaces, and the fix for that red is a row.
//
// Each row has a fixture test in scripts/racescope/testdata that only this row selects, so a
// deleted row is a red test rather than a smaller race set.
var Table = []Start{
	{"(*testing.T).Parallel", "the test resumes beside every other parallel test in its package"},
	{"net/http/httptest.NewServer", "the handler runs on the server's connection goroutines"},
	{"net/http/httptest.NewTLSServer", "the handler runs on the server's connection goroutines"},
	{"net/http/httptest.NewUnstartedServer", "the handler runs on connection goroutines once the caller starts it"},
	{"(*net/http.Server).Serve", "the handler runs on a goroutine per connection"},
	{"(*net/http.Server).ServeTLS", "the handler runs on a goroutine per connection"},
	{"(*net/http.Server).ListenAndServe", "the handler runs on a goroutine per connection"},
	{"(*net/http.Server).ListenAndServeTLS", "the handler runs on a goroutine per connection"},
	{"net/http.Serve", "the handler runs on a goroutine per connection"},
	{"net/http.ServeTLS", "the handler runs on a goroutine per connection"},
	{"net/http.ListenAndServe", "the handler runs on a goroutine per connection"},
	{"net/http.ListenAndServeTLS", "the handler runs on a goroutine per connection"},
	{"time.AfterFunc", "the callback runs on its own goroutine when the timer fires"},
	{"os/signal.Notify", "the runtime delivers on the channel from its own goroutine, beside whatever the receiver shares"},
	{"context.AfterFunc", "the callback runs on its own goroutine when the context ends"},
	{"(*golang.org/x/sync/errgroup.Group).Go", "the function runs on a new goroutine"},
	{"(*sync.WaitGroup).Go", "the function runs on a new goroutine"},
}

// Lookup returns the row for a callee's full name.
func Lookup(fullName string) (Start, bool) {
	for _, s := range Table {
		if s.Callee == fullName {
			return s, true
		}
	}
	return Start{}, false
}
