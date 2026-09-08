package doctor

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"
)

// fakeFS answers only the paths it holds; everything else is a missing file, which is the
// normal case for a settings chain rather than an error worth reporting.
func fakeFS(files map[string]string) func(string) ([]byte, error) {
	return func(p string) ([]byte, error) {
		if body, ok := files[filepath.Clean(p)]; ok {
			return []byte(body), nil
		}
		return nil, errors.New("no such file")
	}
}

func userSettings(home, body string) (string, string) {
	return filepath.Clean(filepath.Join(home, ".claude", "settings.json")), body
}

const grayOn = `{"enabledPlugins":{"gray-area@special-circumstances":true}}`

// THE WHOLE POINT OF THE CHECK: gray-area capturing with reasoning switched off produces a
// trajectory whose deeds are complete and whose thought is empty — and it reads as a complete
// capture, which is why a human has to be told rather than left to notice.
func TestWarnsWhenGrayAreaIsOnAndSummariesAreOff(t *testing.T) {
	p, b := userSettings("/home/u", grayOn)
	got := captureWarnings("/home/u", "/proj", fakeFS(map[string]string{p: b}))
	if len(got) != 1 {
		t.Fatalf("want 1 warning, got %v", got)
	}
	for _, want := range []string{"THINKING CAPTURE OFF", "showThinkingSummaries", "is not set"} {
		if !strings.Contains(got[0], want) {
			t.Errorf("warning does not mention %q: %s", want, got[0])
		}
	}
	// The cost is the part that makes it worth acting on rather than dismissing.
	if !strings.Contains(got[0], "billed") {
		t.Errorf("warning does not say the thinking is billed anyway: %s", got[0])
	}
}

// A DELIBERATE false IS NOT AN OVERSIGHT, and must not be described as one. Same warning —
// capture is still degraded — different sentence, because a doctor that reports a decision in
// the words of a mistake is one a reader learns to skip.
func TestAnExplicitFalseIsReportedAsADecisionNotAnOversight(t *testing.T) {
	p, b := userSettings("/home/u", `{"enabledPlugins":{"gray-area@special-circumstances":true},"showThinkingSummaries":false}`)
	got := captureWarnings("/home/u", "/proj", fakeFS(map[string]string{p: b}))
	if len(got) != 1 {
		t.Fatalf("want 1 warning, got %v", got)
	}
	if !strings.Contains(got[0], "is set to false") {
		t.Errorf("an explicit false was described as absent: %s", got[0])
	}
	if strings.Contains(got[0], "defaults to off") {
		t.Errorf("an explicit false was described as a default: %s", got[0])
	}
}

func TestSilentWhenSummariesAreOn(t *testing.T) {
	p, b := userSettings("/home/u", `{"enabledPlugins":{"gray-area@special-circumstances":true},"showThinkingSummaries":true}`)
	if got := captureWarnings("/home/u", "/proj", fakeFS(map[string]string{p: b})); len(got) != 0 {
		t.Errorf("warned with the setting on: %v", got)
	}
}

// The check is FOR gray-area. A box without it installed is not missing anything, and a
// warning that fires there is noise in a report whose value is that its lines mean something.
func TestSilentWhenGrayAreaIsNotInstalled(t *testing.T) {
	p, b := userSettings("/home/u", `{"enabledPlugins":{"frank-exchange-of-views@special-circumstances":true}}`)
	if got := captureWarnings("/home/u", "/proj", fakeFS(map[string]string{p: b})); len(got) != 0 {
		t.Errorf("warned on a box with no gray-area: %v", got)
	}
}

// PRECEDENCE, NOT UNION — the case a simpler implementation gets wrong.
//
// A project-local `false` overrides a user-level `true`. Asking "does any file say true" would
// report capture as live on a box where it is switched off, which is the exact failure this
// whole check exists to prevent, reproduced inside the check itself.
func TestALocalFalseOverridesAUserTrue(t *testing.T) {
	up, ub := userSettings("/home/u", `{"enabledPlugins":{"gray-area@special-circumstances":true},"showThinkingSummaries":true}`)
	local := filepath.Clean(filepath.Join("/proj", ".claude", "settings.local.json"))
	got := captureWarnings("/home/u", "/proj", fakeFS(map[string]string{
		up:    ub,
		local: `{"showThinkingSummaries":false}`,
	}))
	if len(got) != 1 {
		t.Fatalf("the local override was ignored; want a warning, got %v", got)
	}
	if !strings.Contains(got[0], "is set to false") {
		t.Errorf("resolved to the wrong file: %s", got[0])
	}
}

// And the other direction: a project file may turn it ON for a box whose user settings are
// silent, and that must not warn.
func TestAProjectTrueSatisfiesASilentUserSetting(t *testing.T) {
	up, ub := userSettings("/home/u", grayOn)
	proj := filepath.Clean(filepath.Join("/proj", ".claude", "settings.json"))
	got := captureWarnings("/home/u", "/proj", fakeFS(map[string]string{
		up:   ub,
		proj: `{"showThinkingSummaries":true}`,
	}))
	if len(got) != 0 {
		t.Errorf("warned though the project turns it on: %v", got)
	}
}

// AN UNREADABLE CHAIN IS NOT A CLEAN BOARD. Silence here would be indistinguishable from
// "summaries are on", and the two states have opposite consequences for what a captured
// trajectory contains.
func TestAnUnreadableChainSaysSoRatherThanPassing(t *testing.T) {
	// THE FIXTURE PATH IS BUILT THE WAY THE CODE BUILDS IT — fakeFS keys on filepath.Join,
	// exactly as settingsChain composes. The first draft hand-rolled
	// `HasPrefix(p, "/home/u") && HasSuffix(p, ".claude/settings.json")` instead, which on
	// Windows matched nothing: the chain composes `\home\u\.claude\settings.json` and the
	// prefix test wants forward slashes. Green on Linux, red on the runner — the trap
	// gray-area's manifest_test.go documents at the top of the file, and the reason to
	// build fixtures through the same function rather than to loosen the matcher.
	p, b := userSettings("/home/u", grayOn)
	read := fakeFS(map[string]string{p: b})
	// Sanity: with that chain the setting is absent but readable, so the ordinary warning fires.
	if got := captureWarnings("/home/u", "/proj", read); len(got) != 1 || !strings.Contains(got[0], "THINKING CAPTURE OFF") {
		t.Fatalf("precondition wrong: %v", got)
	}
	// Now the readable=false arm, asserted on the resolver itself: nothing in the chain answers.
	on, explicit, readable := thinkingSummariesOn(settingsChain("/home/u", "/proj"), func(string) ([]byte, error) {
		return nil, errors.New("permission denied")
	})
	if on || explicit || readable {
		t.Errorf("thinkingSummariesOn on an unreadable chain = (%v,%v,%v); want all false — "+
			"a chain nothing answered must be distinguishable from one that answered false", on, explicit, readable)
	}
}

// Malformed JSON must not decide the question, and must not stop a later file from deciding it.
func TestAMalformedFileDoesNotDecideOrBlock(t *testing.T) {
	local := filepath.Clean(filepath.Join("/proj", ".claude", "settings.local.json"))
	up, _ := userSettings("/home/u", "")
	got := captureWarnings("/home/u", "/proj", fakeFS(map[string]string{
		local: `{"showThinkingSummaries": tru`, // truncated
		up:    `{"enabledPlugins":{"gray-area@special-circumstances":true},"showThinkingSummaries":true}`,
	}))
	if len(got) != 0 {
		t.Errorf("a malformed higher-precedence file masked a valid answer below it: %v", got)
	}
}

// THE WIRING, PINNED APART FROM THE MECHANISM.
//
// Every other test in this file exercises captureWarnings directly. All of them stay green if
// the call is deleted from run(), and the doctor then stops asking the question in silence —
// the same shape as the defect it checks for. This is the only test that fails when the wiring
// goes, which is why the seam exists at all.
func TestBoxWarningsReachTheVerdict(t *testing.T) {
	orig := boxWarnings
	t.Cleanup(func() { boxWarnings = orig })
	boxWarnings = func() []string { return []string{"THINKING CAPTURE OFF: synthetic"} }

	root := doctorRoot(t, manifestJSON("optional"))
	out, _ := doctor(t, root)

	if !strings.Contains(out, "THINKING CAPTURE OFF: synthetic") {
		t.Fatalf("the box warning never reached stdout — run() is not wired to boxWarnings:\n%s", out)
	}
	// It must count toward the verdict like any other warning, not merely print. A line a
	// reader can skip past while the headline still says READY is the failure the dance
	// warnings already learned.
	if !strings.Contains(out, "VERDICT: DEGRADED") {
		t.Fatalf("a box warning did not downgrade READY:\n%s", out)
	}
}

// And the quiet case must stay quiet all the way through: a configured box prints no capture
// line and keeps its READY.
func TestAConfiguredBoxAddsNoLineAndKeepsReady(t *testing.T) {
	orig := boxWarnings
	t.Cleanup(func() { boxWarnings = orig })
	boxWarnings = func() []string { return nil }

	root := doctorRoot(t, manifestJSON("optional"))
	out, _ := doctor(t, root)
	if strings.Contains(out, "THINKING CAPTURE") {
		t.Errorf("a configured box was warned at anyway:\n%s", out)
	}
}
