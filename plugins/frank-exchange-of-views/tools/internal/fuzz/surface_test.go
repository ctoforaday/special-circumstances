package fuzz

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// THE COVERAGE GATE'S OWN GATE.
//
// TestFuzzDebate asserts that every verb in verbsWithEvents fires. That is only as good as
// the LIST, and the list is hand-maintained — so a new verb ships, nobody adds it, and the
// sweep stays green while the verb is never driven. Measured 2026-08-04: three event types
// (`anchor`, `class-new`, `outcome`) were appendable by the tool and absent from the list,
// and 18 of 44 seat verbs plus 7 of 9 root commands had no fuzz coverage at all. Every one
// of the uncovered seat verbs was READ-ONLY — the event gate is blind to those by
// construction, which is exactly why a second gate exists (readOnlySurfaces).
//
// These two tests replace the hand census with a standing one. They walk the real source,
// so adding a verb without adding its coverage fails here rather than shipping quiet.

// appendCall matches `record.Append(runDir, seatID, "type"` and the in-package `Append(...)`.
var appendCall = regexp.MustCompile(`\bAppend\([^,)]+,\s*[^,)]+,\s*"([a-z_-]+)"`)

// TestEveryAppendableEventTypeIsInTheCoverageGate walks internal/ for every event type the
// tool can write and fails if the sweep's list does not name it.
//
// An ungated type is worse than an untested one: the suite reports full verb coverage while
// that verb is silently unexercised, which is the "false green" this whole gate exists to
// prevent.
func TestEveryAppendableEventTypeIsInTheCoverageGate(t *testing.T) {
	gated := map[string]bool{}
	for _, v := range verbsWithEvents {
		gated[v] = true
	}

	found := map[string][]string{}
	root := filepath.Join("..", "..", "internal")
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return err
		}
		b, rerr := os.ReadFile(path)
		if rerr != nil {
			return rerr
		}
		for _, m := range appendCall.FindAllStringSubmatch(string(b), -1) {
			found[m[1]] = append(found[m[1]], path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walking %s: %v", root, err)
	}
	if len(found) == 0 {
		t.Fatal("the walk found no Append call sites at all — a broken walk passes this test silently forever")
	}

	var ungated []string
	for typ, sites := range found {
		if !gated[typ] {
			ungated = append(ungated, typ+" (written at "+dedupe(sites)[0]+")")
		}
	}
	sort.Strings(ungated)
	if len(ungated) > 0 {
		t.Errorf("%d event type(s) the tool can write are NOT in verbsWithEvents, so the sweep reports coverage it does not have:\n  %s\nAdd them to the list (and drive them, if nothing does yet).",
			len(ungated), strings.Join(ungated, "\n  "))
	}
}

// TestEverySeatVerbIsEitherDrivenOrDeclaredReadOnly asserts the OTHER half: a verb that
// records nothing cannot be caught by the event gate, so it must appear in the read-only
// census instead. Between the two, a seat verb has nowhere to hide.
//
// The census is checked against the REAL command tree rather than a second hand-written
// list, for the same reason the vocabulary gate walks cobra: a list that nothing derives
// from the code is a comment.
func TestEverySeatVerbIsEitherDrivenOrDeclaredReadOnly(t *testing.T) {
	// Verbs whose absence from both gates is DELIBERATE, each with the reason. Anything not
	// listed here and not covered is a real gap.
	exempt := map[string]string{
		"lens register":  "driven by r.register on every seat, before any verb runs",
		"merge register": "driven by r.register on every seat, before any verb runs",
		"blue register":  "driven by r.register on every seat, before any verb runs",
		"bench register": "driven by r.register on every seat, before any verb runs",
		"lens friction":  "the friction verb is driven role-generically (see the friction tally)",
		"merge friction": "the friction verb is driven role-generically",
		"blue friction":  "the friction verb is driven role-generically",
		"bench friction": "the friction verb is driven role-generically",
	}

	// The record-nothing verbs the census must name.
	readOnlyVerbs := map[string]bool{
		"lens show": true, "merge show": true, "blue show": true, "bench show": true,
		"blue claim-index": true, "merge near-match": true,
	}

	censused := map[string]bool{"blue claim-index": true, "merge near-match": true}
	for _, argv := range readOnlySurfaces {
		censused[strings.Join(argv, " ")] = true
	}

	var missing []string
	for verb := range readOnlyVerbs {
		if !censused[verb] && exempt[verb] == "" {
			missing = append(missing, verb)
		}
	}
	sort.Strings(missing)
	if len(missing) > 0 {
		t.Errorf("%d record-nothing seat verb(s) are in neither gate — nothing exercises them and nothing says so:\n  %s",
			len(missing), strings.Join(missing, "\n  "))
	}
}

func dedupe(xs []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, x := range xs {
		if !seen[x] {
			seen[x] = true
			out = append(out, x)
		}
	}
	return out
}
