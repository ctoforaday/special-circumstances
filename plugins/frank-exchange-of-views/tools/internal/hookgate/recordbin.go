package hookgate

import "strings"

// RecordBin is the record tool's name as a seat invokes it — the name its prompt's tool path and its
// own refusal lines carry ("feov-record: …").
const RecordBin = "feov-record"

// CallsRecord reports whether a Bash command names the record's tool — the only way a seat's call
// records an act. It is a cheap filter on WHICH calls a hook spends a process on, never a gate on
// what a call may do: a call it misses is the tool reached some other way, and what that costs is
// the hook's to say where it relies on this.
func CallsRecord(command string) bool { return strings.Contains(command, RecordBin) }
