package telecli

import (
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/gray-area/tools/internal/catalogue"
)

// WITHOUT A BOOT, `agents` STILL ANSWERS AND `--lost` REFUSES. Plain `agents` on a platform with no
// boot reader lists the advertised sessions exactly as before plus ONE line saying what it could
// not do; `--lost` has nothing to measure from and exits 1 rather than guess. A nil Boot — every
// Env literal that predates the field — behaves the same way and never panics.
func TestWithoutABootAgentsStillAnswersAndLostRefuses(t *testing.T) {
	for _, tc := range []struct {
		name string
		boot func() (int64, bool)
	}{
		{"unmeasurable", func() (int64, bool) { return 0, false }},
		{"nil", nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h := newHarness(t)
			h.env.Boot = tc.boot
			out, errOut, code := h.run(t, "agents")
			if code != 0 {
				t.Fatalf("agents exited %d without a boot, stderr:\n%s", code, errOut)
			}
			for _, id := range []string{alphaID, betaID} {
				if !strings.Contains(out, catalogue.Short(id)) {
					t.Errorf("the advertised session %s is missing:\n%s", catalogue.Short(id), out)
				}
			}
			if n := strings.Count(out, errBootUnmeasured.Error()); n != 1 {
				t.Errorf("the unmeasurable boot is said %d times, want once:\n%s", n, out)
			}
			if strings.Contains(out, "lost") {
				t.Errorf("sessions were listed as lost with no boot to measure from:\n%s", out)
			}
			_, errOut, code = h.run(t, "agents", "--lost")
			if code != 1 || !strings.Contains(errOut, "boot time not measurable here — pass --boot") {
				t.Errorf("--lost without a boot: exit %d, stderr %q — want exit 1 and the refusal", code, errOut)
			}
			out, _, code = h.run(t, "agents", "--lost", "--boot", bootArg())
			if code != 0 || !strings.Contains(out, "(from --boot)") {
				t.Errorf("--boot did not stand in for the reader: exit %d\n%s", code, out)
			}
		})
	}
}

// THE SECOND LINE CARRIES THE BRING-BACK ARITHMETIC, computed here so the skill's first step reads
// a verdict rather than doing subtraction. The newest pre-boot activity is beta's last act.
func TestTheSecondLineStatesTheBringBackVerdict(t *testing.T) {
	h := newHarness(t)
	newest := frozen.Add(-29 * time.Minute)
	for _, tc := range []struct {
		after time.Duration
		want  string
	}{
		{time.Hour, "newest pre-boot activity 2026-09-09 11:31:00 UTC (session bbbbbbbb) — 1h00m ago, " +
			"inside the ~4 h window: open server-spawned sessions in the app first"},
		{7 * time.Hour, "newest pre-boot activity 2026-09-09 11:31:00 UTC (session bbbbbbbb) — 7h00m ago, " +
			"past the ~4 h a restarted Remote Control server brings its sessions back in"},
	} {
		h.env.Now = func() time.Time { return newest.Add(tc.after) }
		out, errOut, code := h.run(t, "agents", "--lost", "--boot", bootArg())
		if code != 0 {
			t.Fatalf("exit %d: %s", code, errOut)
		}
		if got := strings.Split(out, "\n")[1]; got != tc.want {
			t.Errorf("%s after: second line\n  %q\nwant\n  %q", tc.after, got, tc.want)
		}
	}
	h.env.Now = func() time.Time { return frozen }
	out, _, _ := h.run(t, "agents", "--lost", "--boot", fmt.Sprint(frozen.Add(-10*time.Hour).Unix()))
	if got := strings.Split(out, "\n")[1]; got != "newest pre-boot activity: none in the store — run telepathy backfill, then ask again" {
		t.Errorf("with nothing before the boot the second line reads %q", got)
	}
}

// `unknown` IS NEVER `lost`: alpha and beta precede the boot and are not ours to judge.
func TestUnknownIsNeverLost(t *testing.T) {
	h := newHarness(t)
	out, _, code := h.run(t, "agents", "--lost", "--boot", bootArg())
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	for _, id := range []string{alphaID, betaID} {
		if strings.Contains(out, "lost "+id) {
			t.Errorf("the unknown session %s was listed as lost:\n%s", catalogue.Short(id), out)
		}
	}
	for _, id := range []string{gammaID, deltaID} {
		if !strings.Contains(out, "lost "+id) {
			t.Errorf("%s is in the store, not advertised, and precedes the boot, but is not listed:\n%s", catalogue.Short(id), out)
		}
	}
}

// AN EMPTY STORE IS A SENTENCE, not a clean "nothing was cut off".
func TestAnEmptyStoreIsSaidUnderLost(t *testing.T) {
	h := newColdHarness(t)
	db, err := catalogue.Open(h.store, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	db.Close()
	out, _, code := h.run(t, "agents", "--lost", "--boot", bootArg())
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	if !strings.Contains(out, "the store holds no sessions — run `telepathy backfill`, or capture has not run here") ||
		strings.Contains(out, "nothing was cut off") {
		t.Errorf("an empty store under --lost:\n%s", out)
	}
}

// A BAD --window OR --boot IS THE CALLER'S TYPO: exit 2.
func TestAgentsRefusesABadWindowOrBoot(t *testing.T) {
	h := newHarness(t)
	for _, args := range [][]string{{"agents", "--window", "0s"}, {"agents", "--boot", "-5"}} {
		if _, _, code := h.run(t, args...); code != 2 {
			t.Errorf("%v exited %d, want 2", args, code)
		}
	}
}
