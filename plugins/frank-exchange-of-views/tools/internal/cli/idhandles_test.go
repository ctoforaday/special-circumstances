package cli

import (
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// POSITIONAL HANDLES FOR TABLE ROWS.
//
// An id is minted at random, so a table row written before its run exists cannot spell one. A row
// names its subject by POSITION instead — G1 is the first gap the run minted, M2 the second motion
// filed, Q1 the first avenue proposed — and the driver resolves the handle against the record at
// the moment the step runs. The handle never reaches the tool: what the verb is handed is the id
// the record minted.

// handleArg is an argument that is nothing but handles: one, or a comma list of them.
var handleArg = regexp.MustCompile(`^[GMQ][0-9]+(,[GMQ][0-9]+)*$`)

// handleWord is a handle standing as a word in prose.
var handleWord = regexp.MustCompile(`\b[GMQ][0-9]+\b`)

// mintedIDs are the ids the record minted, by letter, in the order they were minted.
func mintedIDs(t *testing.T, runDir string) map[string][]string {
	t.Helper()
	ids := map[string][]string{}
	add := func(letter, id string) {
		if id != "" && !slices.Contains(ids[letter], id) {
			ids[letter] = append(ids[letter], id)
		}
	}
	for _, e := range events(t, runDir) {
		if m, ok := recordpb.BodyAs[*recordpb.Mint](e); ok {
			add("G", m.GetGapId())
		}
		if m, ok := recordpb.BodyAs[*recordpb.Motion](e); ok {
			add("M", m.GetMotionId())
		}
		if a, ok := recordpb.BodyAs[*recordpb.Avenue](e); ok {
			add("Q", a.GetAvenueId())
		}
	}
	return ids
}

// handleID is the id a handle names. A handle past what the run minted names an id in the kind's
// shape that the record does not hold — the row's way of saying "one nobody minted", which has to
// pass the flag's shape check to reach the refusal the row is about.
func handleID(ids map[string][]string, handle string) string {
	letter := handle[:1]
	n, _ := strconv.Atoi(handle[1:])
	if n >= 1 && n <= len(ids[letter]) {
		return ids[letter][n-1]
	}
	return fmt.Sprintf("%s-%08x", letter, n)
}

// byHandle is argv with every handle argument replaced by the id it names in this run. A --key's
// value is the seat's own retry key and is left as written, whatever it looks like.
func byHandle(t *testing.T, runDir string, args []string) []string {
	t.Helper()
	var ids map[string][]string
	out := slices.Clone(args)
	for i, a := range out {
		if !handleArg.MatchString(a) || i > 0 && out[i-1] == "--key" {
			continue
		}
		if ids == nil {
			ids = mintedIDs(t, runDir)
		}
		parts := strings.Split(a, ",")
		for j, p := range parts {
			parts[j] = handleID(ids, p)
		}
		out[i] = strings.Join(parts, ",")
	}
	return out
}

// runAt is run for a call that names its subject by handle: the handles are resolved against the
// run the call's own --run names, and the verb is handed the ids. A call that names no --run has no
// record to resolve against and goes to the tool as written — the rows that omit it are the ones
// asserting the tool refuses for want of it.
func runAt(t *testing.T, args ...string) (string, error) {
	t.Helper()
	dir := flagValue(args, "--run")
	if dir == "" {
		return run(t, args...)
	}
	return run(t, byHandle(t, dir, args)...)
}

// runMintAt is runMint for a mint that names another gap by handle.
func runMintAt(t *testing.T, runDir string, args ...string) (string, error) {
	t.Helper()
	return runMint(t, runDir, byHandle(t, runDir, args)...)
}

// handleText is prose with every handle word replaced by the id it names in this run: what a row
// expects a verb to SAY about its subject.
func handleText(t *testing.T, runDir, s string) string {
	t.Helper()
	if !handleWord.MatchString(s) {
		return s
	}
	ids := mintedIDs(t, runDir)
	return handleWord.ReplaceAllStringFunc(s, func(h string) string { return handleID(ids, h) })
}

// onlyID is the one id of a letter the run has minted, for a fixture that mints exactly one.
func onlyID(t *testing.T, runDir, letter string) string {
	t.Helper()
	ids := mintedIDs(t, runDir)[letter]
	if len(ids) != 1 {
		t.Fatalf("the run holds %d %s- id(s), want exactly one: %v", len(ids), letter, ids)
	}
	return ids[0]
}
