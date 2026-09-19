package setup

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// NARROWING THE CAST DROPS AN AUDIT DIMENSION, so the reason is a term of the run and not a
// remark. Measured across the nine archived is-91-prime runs: every one seated evidence, logic
// and voice only, all seven lens agents having existed since before the first of them, and not
// one states why. Five shipped VERIFIED, and of 34 gaps minted across them exactly ONE was a
// defect in the mathematics — found by the LOGIC lens, because `computation` was never seated.
//
// The skill has always said to narrow "only with a reason". These tests are the difference
// between saying that and holding anyone to it.

func TestANarrowedCastWithoutAReasonIsRefused(t *testing.T) {
	cfg, _ := runCfg(t, reports(fmt.Sprint(record.EventSchema)))
	cfg.LensAreas = []string{"evidence", "logic", "voice"}

	var out, errb bytes.Buffer
	if code := Run(cfg, &out, &errb); code != 2 {
		t.Fatalf("a silent narrowing was accepted: exit %d\n%s", code, errb.String())
	}
	msg := errb.String()
	// THE REFUSAL NAMES WHAT IS BEING GIVEN UP, not merely that a flag is missing. An operator
	// who cannot see which dimensions go dark cannot judge whether the topic needs them.
	for _, want := range []string{
		"3 of 7 lens areas",
		"evidence, logic, voice",
		"computation",
		"architecture",
		"--lens-area-reason",
	} {
		if !strings.Contains(msg, want) {
			t.Errorf("the refusal does not say %q:\n%s", want, msg)
		}
	}
}

func TestANarrowedCastWithAReasonRunsAndRecordsIt(t *testing.T) {
	cfg, runDir := runCfg(t, reports(fmt.Sprint(record.EventSchema)))
	cfg.LensAreas = []string{"evidence", "logic", "voice"}
	const why = "elementary arithmetic: no code, no design and no adversary surface to audit"
	cfg.LensAreaReason = why

	var out, errb bytes.Buffer
	if code := Run(cfg, &out, &errb); code != 0 {
		t.Fatalf("a justified narrowing was refused: exit %d\n%s", code, errb.String())
	}

	// THE REASON OUTLIVES THE LAUNCHER. run-config.json is archived with the run, so a reader a
	// year later can see what the narrowing was for without the script that made it — which is
	// the whole point, since every launcher that narrowed these nine runs lives outside the repo.
	b, err := os.ReadFile(filepath.Join(runDir, "inputs", "run-config.json"))
	if err != nil {
		t.Fatal(err)
	}
	var rc struct {
		LensAreas      []string `json:"lensAreas"`
		LensAreaReason string   `json:"lensAreaReason"`
	}
	if err := json.Unmarshal(b, &rc); err != nil {
		t.Fatal(err)
	}
	if rc.LensAreaReason != why {
		t.Errorf("run-config.json carries reason %q, want %q", rc.LensAreaReason, why)
	}
	if len(rc.LensAreas) != 3 {
		t.Errorf("run-config.json carries areas %v, want the three that were seated", rc.LensAreas)
	}
}

// A FULL CAST IS NOT A NARROWING, and must not be made to justify itself. Requiring a reason for
// the default would teach an operator to type a word to get past a prompt, which is how a gate
// becomes a formality — and the absent fields are what let a reader tell a full cast from a
// narrowed one whose reason nobody wrote.
func TestAFullCastNeedsNoReasonAndRecordsNoNarrowing(t *testing.T) {
	for _, tc := range []struct {
		name  string
		areas []string
	}{
		{"default, no --lens-area at all", nil},
		{"every area named explicitly", record.DefaultCastAreas},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg, runDir := runCfg(t, reports(fmt.Sprint(record.EventSchema)))
			cfg.LensAreas = tc.areas

			var out, errb bytes.Buffer
			if code := Run(cfg, &out, &errb); code != 0 {
				t.Fatalf("a full cast was refused: exit %d\n%s", code, errb.String())
			}
			b, err := os.ReadFile(filepath.Join(runDir, "inputs", "run-config.json"))
			if err != nil {
				t.Fatal(err)
			}
			if bytes.Contains(b, []byte("lensAreaReason")) {
				t.Errorf("a full cast recorded a narrowing reason:\n%s", b)
			}
		})
	}
}
