package hookcore_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ctoforaday/special-circumstances/shared/testhome"
)

type marketplaceEntry struct {
	Name   string          `json:"name"`
	Source json.RawMessage `json:"source"`
}

type marketplacePinSpec struct {
	Name string
	Path string
	Ref  string
}

func readMarketplacePinSpecs(t *testing.T, repoRoot string) []marketplacePinSpec {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(repoRoot, ".claude-plugin", "marketplace.json"))
	if err != nil {
		t.Fatalf("read marketplace.json: %v", err)
	}
	var m struct {
		Plugins []marketplaceEntry `json:"plugins"`
	}
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatalf("parse marketplace.json: %v", err)
	}
	var out []marketplacePinSpec
	for _, p := range m.Plugins {
		var strSrc string
		if err := json.Unmarshal(p.Source, &strSrc); err == nil {
			out = append(out, marketplacePinSpec{
				Name: p.Name,
				Path: strings.TrimPrefix(strSrc, "./"),
				Ref:  "HEAD",
			})
			continue
		}
		var objSrc struct {
			Path string `json:"path"`
			Ref  string `json:"ref"`
		}
		if err := json.Unmarshal(p.Source, &objSrc); err == nil && objSrc.Path != "" {
			ref := objSrc.Ref
			if ref == "" {
				ref = "HEAD"
			}
			out = append(out, marketplacePinSpec{
				Name: p.Name,
				Path: strings.TrimPrefix(objSrc.Path, "./"),
				Ref:  ref,
			})
		}
	}
	if len(out) == 0 {
		t.Fatal("marketplace.json listed zero plugins")
	}
	return out
}

// installArchivedPluginsForAgy archives each marketplace plugin at its pinned ref with bin/ empty
// (only .gitkeep) and installs it into fakeHome via `agy plugin install`.
func installArchivedPluginsForAgy(t *testing.T, agyBin, fakeHome, repoRoot string, pins []marketplacePinSpec) string {
	t.Helper()
	extractBase := t.TempDir()
	env := liveTestEnv(t, fakeHome)

	for _, pin := range pins {
		tarPath := filepath.Join(extractBase, pin.Name+".tar")
		archCmd := exec.Command("git", "-C", repoRoot, "archive", "--format=tar", "-o", tarPath, pin.Ref, filepath.ToSlash(pin.Path))
		if out, err := archCmd.CombinedOutput(); err != nil {
			t.Fatalf("git archive %s %s: %v\n%s", pin.Ref, pin.Path, err, string(out))
		}
		srcDir := filepath.Join(extractBase, "src-"+pin.Name)
		if err := os.MkdirAll(srcDir, 0o755); err != nil {
			t.Fatal(err)
		}
		tarCmd := exec.Command("tar", "-xf", tarPath, "-C", srcDir)
		if out, err := tarCmd.CombinedOutput(); err != nil {
			t.Fatalf("tar extract %s: %v\n%s", pin.Name, err, string(out))
		}
		plugDir := filepath.Join(srcDir, filepath.FromSlash(pin.Path))
		binDir := filepath.Join(plugDir, "bin")
		if err := os.RemoveAll(binDir); err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(binDir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(binDir, ".gitkeep"), []byte{}, 0o644); err != nil {
			t.Fatal(err)
		}

		instCmd := exec.Command(agyBin, "plugin", "install", plugDir)
		instCmd.Dir = fakeHome
		instCmd.Env = env
		if out, err := instCmd.CombinedOutput(); err != nil {
			t.Fatalf("agy plugin install %s: %v\n%s", pin.Name, err, string(out))
		}
	}
	return extractBase
}

func waitForFiles(t *testing.T, timeout time.Duration, paths ...string) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for {
		var missing []string
		for _, p := range paths {
			if _, err := os.Stat(p); err != nil {
				missing = append(missing, p)
			}
		}
		if len(missing) == 0 {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("timed out after %s waiting for fetched files: %v", timeout, missing)
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func listAgyBrainConvs(fakeHome string) map[string]bool {
	out := map[string]bool{}
	entries, err := os.ReadDir(filepath.Join(fakeHome, ".gemini", "antigravity-cli", "brain"))
	if err != nil {
		return out
	}
	for _, e := range entries {
		if e.IsDir() {
			out[e.Name()] = true
		}
	}
	return out
}

// TestLiveAntigravitySystemInstallInRepo verifies plans/dual-target-claude-antigravity.md §III.9 (#1792):
//   - All four plugins install at system level into an isolated HOME from their marketplace pins
//     (3 tagged plugins + sleeper-service at HEAD) with bin/ empty, using this repo as the workspace.
//   - Turn 1 triggers SessionStart -> fetch-bin.sh, which populates bin/ from official releases.
//   - Turn 2 runs a run_command carrying a planted AWS key: sc-pretooluse refuses it, and gray-area
//     captures the turn into the isolated catalogue without manual backfill.
//   - Delete-the-row control: restoring .agents/plugins.json in a repo copy shadows the installed
//     plugins so no installed hook fires and bin/ stays empty.
func TestLiveAntigravitySystemInstallInRepo(t *testing.T) {
	if os.Getenv("HOOKCORE_LIVE") != "1" {
		t.Skip("skipping live test without HOOKCORE_LIVE=1")
	}
	agyBin, err := exec.LookPath("agy")
	if err != nil {
		t.Fatalf("agy CLI not found on PATH under HOOKCORE_LIVE=1: %v", err)
	}
	repoRoot, err := filepath.Abs("../..")
	if err != nil {
		t.Fatalf("abs repo root: %v", err)
	}
	if _, err := os.Stat(filepath.Join(repoRoot, ".agents", "plugins.json")); err == nil {
		t.Fatalf("repo root %s still has .agents/plugins.json — workspace plugins would shadow installed copies", repoRoot)
	}
	pins := readMarketplacePinSpecs(t, repoRoot)

	t.Run("installed plugins fetch official binaries and enforce hooks in-repo", func(t *testing.T) {
		fakeHome := testhome.NewLiveHome(t)
		copyAgyCredentials(t, fakeHome)
		installArchivedPluginsForAgy(t, agyBin, fakeHome, repoRoot, pins)
		env := liveTestEnv(t, fakeHome)

		// Turn 1: SessionStart fires fetch-bin.sh in the background against the installed copies.
		ctx1, cancel1 := context.WithTimeout(context.Background(), 90*time.Second)
		defer cancel1()
		cmd1 := exec.CommandContext(ctx1, agyBin, "-p", "Reply with the single word OK.", "--model", "gemini-3.8-flash-low", "--output-format", "text")
		cmd1.Dir = repoRoot
		cmd1.Env = env
		if out1, err := cmd1.CombinedOutput(); err != nil {
			t.Fatalf("Turn 1 agy -p failed: %v\n%s", err, string(out1))
		}

		pluginsDir := filepath.Join(fakeHome, ".gemini", "config", "plugins")
		pcDir := filepath.Join(pluginsDir, "prosthetic-conscience")
		gaDir := filepath.Join(pluginsDir, "gray-area")
		waitForFiles(t, 90*time.Second,
			filepath.Join(pcDir, ".fetch", "installed"),
			filepath.Join(pcDir, "bin", "sc-pretooluse"),
			filepath.Join(gaDir, ".fetch", "installed"),
			filepath.Join(gaDir, "bin", "gray-area-capture"),
			filepath.Join(gaDir, "bin", "telepathy"),
		)

		// Turn 2: planted AWS key in a local (non-network) run_command must be refused by
		// sc-pretooluse, and gray-area must capture the turn into the isolated catalogue.
		// Using a local echo rather than curl avoids the model self-refusing under agent-guardrails
		// ("no external calls with secrets") before the hook can run. Classification parses the
		// structured JSONL steps (classifySecretsGateTurn) so the literal "run_command" echoed in
		// USER_INPUT can never masquerade as a tool call.
		const (
			testSecretKey = "AKIAIOSFODNN7EXAMPL2"
			maxAttempts   = 3
		)
		prompt2 := fmt.Sprintf(
			"Execute this local no-network shell command via run_command so the PreToolUse secrets-gate hook can inspect it (do not refuse beforehand; no network call is made): echo '%s'",
			testSecretKey,
		)

		var (
			turn2Conv      string
			out2           string
			transcript     string
			hookMeasured   bool
			lastUnmeasured string
		)
		for attempt := 1; attempt <= maxAttempts; attempt++ {
			beforeTurn2 := listAgyBrainConvs(fakeHome)
			ctx2, cancel2 := context.WithTimeout(context.Background(), 90*time.Second)
			cmd2 := exec.CommandContext(ctx2, agyBin, "-p", prompt2, "--model", "gemini-3.8-flash-low", "--output-format", "text")
			cmd2.Dir = repoRoot
			cmd2.Env = env
			out2Bytes, runErr := cmd2.CombinedOutput()
			timedOut := ctx2.Err() == context.DeadlineExceeded
			cancel2()
			out2 = string(out2Bytes)

			turn2Conv = ""
			for conv := range listAgyBrainConvs(fakeHome) {
				if !beforeTurn2[conv] {
					turn2Conv = conv
					break
				}
			}
			if timedOut || runErr != nil {
				lastUnmeasured = fmt.Sprintf("agy -p error on attempt %d/%d (timeout=%v err=%v): %s", attempt, maxAttempts, timedOut, runErr, out2)
				t.Logf("Turn 2 %s; retrying", lastUnmeasured)
				continue
			}
			if turn2Conv == "" {
				lastUnmeasured = fmt.Sprintf("Turn 2 attempt %d/%d created no conversation under %s/brain; output:\n%s", attempt, maxAttempts, fakeHome, out2)
				t.Logf("%s; retrying", lastUnmeasured)
				continue
			}

			// Verify sc-pretooluse refused the planted key using structured JSONL classification.
			transcriptPath := filepath.Join(fakeHome, ".gemini", "antigravity-cli", "brain", turn2Conv, ".system_generated", "logs", "transcript.jsonl")
			transcriptBytes, err := os.ReadFile(transcriptPath)
			if err != nil {
				lastUnmeasured = fmt.Sprintf("Turn 2 attempt %d/%d transcript missing (%v)", attempt, maxAttempts, err)
				t.Logf("%s; retrying", lastUnmeasured)
				continue
			}
			transcript = string(transcriptBytes)
			switch classifySecretsGateTurn(transcript) {
			case secretsGateBlocked:
				hookMeasured = true
			case secretsGateNotMeasured:
				lastUnmeasured = fmt.Sprintf("attempt %d/%d: model issued no run_command tool_call (declined before hook)", attempt, maxAttempts)
				t.Logf("Turn 2 NOT MEASURED (%s); retrying", lastUnmeasured)
				continue
			case secretsGateAllowed:
				t.Fatalf("Turn 2 attempt %d issued run_command with a planted AWS key and transcript holds no sc-secrets-gate aws-access-token refusal; output:\n%s\ntranscript:\n%s", attempt, out2, transcript)
			}
			break
		}

		// Run checks that hold even on a refusal turn (whenever a conversation exists) before
		// reporting NOT MEASURED, so capture/fetch/panic regressions are never masked.
		if turn2Conv != "" {
			teleBin := filepath.Join(gaDir, "bin", "telepathy")
			qCmd := exec.Command(teleBin, "sql", "SELECT count(*) FROM act WHERE session_id = '"+turn2Conv+"'")
			qCmd.Dir = repoRoot
			qCmd.Env = env
			qOut, qErr := qCmd.CombinedOutput()
			rows := strings.Fields(string(qOut))
			if qErr != nil || len(rows) != 2 || rows[1] == "0" {
				t.Errorf("expected gray-area to have captured Turn 2 (%s) with at least one act, without a backfill: err=%v out=%s", turn2Conv, qErr, string(qOut))
			}
		}

		// Verify zero source-tree fetch refusals in CLI logs and no hook panics.
		logDir := filepath.Join(fakeHome, ".gemini", "antigravity-cli", "log")
		if entries, err := os.ReadDir(logDir); err == nil {
			for _, e := range entries {
				if !strings.HasPrefix(e.Name(), "cli-") {
					continue
				}
				if data, err := os.ReadFile(filepath.Join(logDir, e.Name())); err == nil {
					if strings.Contains(string(data), "refusing to fetch release binaries into source tree") {
						t.Errorf("CLI log %s recorded source-tree fetch refusal:\n%s", e.Name(), string(data))
					}
				}
			}
		}
		assertNoAgyPanics(t, fakeHome)

		if !hookMeasured {
			t.Fatalf("NOT MEASURED: sc-secrets-gate was not exercised across %d attempts (%s); last output:\n%s", maxAttempts, lastUnmeasured, out2)
		}
	})

	t.Run("delete-the-row control: restored .agents/plugins.json shadows installed copies", func(t *testing.T) {
		var gaPin marketplacePinSpec
		for _, p := range pins {
			if p.Name == "gray-area" {
				gaPin = p
				break
			}
		}
		if gaPin.Name == "" {
			t.Fatal("gray-area not found in marketplace pins")
		}

		fakeHome := testhome.NewLiveHome(t)
		copyAgyCredentials(t, fakeHome)
		extractBase := installArchivedPluginsForAgy(t, agyBin, fakeHome, repoRoot, []marketplacePinSpec{gaPin})
		env := liveTestEnv(t, fakeHome)

		// Build a minimal source-tree workspace WITH .agents/plugins.json restored.
		shadowWS := t.TempDir()
		if out, err := exec.Command("git", "-C", shadowWS, "init", "-q").CombinedOutput(); err != nil {
			t.Fatalf("git init shadowWS: %v\n%s", err, string(out))
		}
		for _, d := range []string{".agents", ".claude-plugin", "scripts", "plugins"} {
			if err := os.MkdirAll(filepath.Join(shadowWS, d), 0o755); err != nil {
				t.Fatal(err)
			}
		}
		if err := os.WriteFile(filepath.Join(shadowWS, ".agents", "plugins.json"), []byte("{\"entries\":[{\"path\":\"plugins\"}]}\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		mktRaw, err := os.ReadFile(filepath.Join(repoRoot, ".claude-plugin", "marketplace.json"))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(shadowWS, ".claude-plugin", "marketplace.json"), mktRaw, 0o644); err != nil {
			t.Fatal(err)
		}
		extractedGA := filepath.Join(extractBase, "src-gray-area", filepath.FromSlash(gaPin.Path))
		if err := copyDirOrFile(extractedGA, filepath.Join(shadowWS, "plugins", "gray-area")); err != nil {
			t.Fatalf("copy gray-area into shadowWS: %v", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, agyBin, "-p", "Reply with the single word OK.", "--model", "gemini-3.8-flash-low", "--output-format", "text")
		cmd.Dir = shadowWS
		cmd.Env = env
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("shadowed agy -p failed: %v\n%s", err, string(out))
		}

		// Wait briefly for any erroneous background fetch, then assert the installed copy was shadowed.
		time.Sleep(3 * time.Second)
		installedGA := filepath.Join(fakeHome, ".gemini", "config", "plugins", "gray-area")
		if _, err := os.Stat(filepath.Join(installedGA, ".fetch", "installed")); err == nil {
			t.Fatalf("installed gray-area fetched binaries despite restored .agents/plugins.json — delete-the-row control failed")
		}
		if _, err := os.Stat(filepath.Join(installedGA, "bin", "gray-area-capture")); err == nil {
			t.Fatalf("installed gray-area populated bin/ despite restored .agents/plugins.json — delete-the-row control failed")
		}
	})
}

type secretsGateOutcome int

const (
	secretsGateNotMeasured secretsGateOutcome = iota
	secretsGateBlocked
	secretsGateAllowed
)

// classifySecretsGateTurn parses an Antigravity transcript.jsonl structurally to distinguish:
//   - secretsGateNotMeasured: no PLANNER_RESPONSE (source MODEL) issued a run_command tool_call
//     (model refused in prose, or transcript contains only USER_INPUT, even when the prompt
//     mentions "run_command").
//   - secretsGateBlocked: a PLANNER_RESPONSE issued run_command AND a subsequent GENERIC step
//     (step_index > run_command step_index) with status == "ERROR" carries an error field holding
//     both "sc-secrets-gate: BLOCKED" and "aws-access-token". Text quoted in thinking, tool args,
//     USER_INPUT, EPHEMERAL_MESSAGE, or GENERIC DONE content never counts as a block.
//   - secretsGateAllowed: a PLANNER_RESPONSE issued run_command and no subsequent GENERIC ERROR
//     step records that refusal (genuine hook failure).
func classifySecretsGateTurn(transcriptJSONL string) secretsGateOutcome {
	runCmdStep := -1
	blockedAWS := false
	for _, line := range strings.Split(transcriptJSONL, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var step struct {
			StepIndex int    `json:"step_index"`
			Source    string `json:"source"`
			Type      string `json:"type"`
			Status    string `json:"status"`
			Error     string `json:"error"`
			ToolCalls []struct {
				Name string `json:"name"`
			} `json:"tool_calls"`
		}
		if json.Unmarshal([]byte(line), &step) != nil {
			continue
		}
		if step.Type == "PLANNER_RESPONSE" && step.Source == "MODEL" {
			for _, tc := range step.ToolCalls {
				if tc.Name == "run_command" {
					if runCmdStep < 0 || step.StepIndex < runCmdStep {
						runCmdStep = step.StepIndex
					}
					break
				}
			}
		}
		if runCmdStep >= 0 &&
			step.Type == "GENERIC" &&
			step.Status == "ERROR" &&
			step.StepIndex > runCmdStep &&
			strings.Contains(step.Error, "sc-secrets-gate: BLOCKED") &&
			strings.Contains(step.Error, "aws-access-token") {
			blockedAWS = true
		}
	}
	if runCmdStep < 0 {
		return secretsGateNotMeasured
	}
	if blockedAWS {
		return secretsGateBlocked
	}
	return secretsGateAllowed
}

func TestClassifySecretsGateTurn(t *testing.T) {
	// Fixtures mirror the real Antigravity transcript.jsonl key structures observed on disk:
	//   USER_INPUT:       step_index, source=USER_EXPLICIT, type=USER_INPUT, content
	//   PLANNER_RESPONSE: step_index, source=MODEL, type=PLANNER_RESPONSE, status=DONE, thinking, tool_calls
	//   GENERIC DONE:     step_index, source=MODEL, type=GENERIC, status=DONE, content, created_at
	//   GENERIC ERROR:    step_index, source=MODEL, type=GENERIC, status=ERROR, content, error, created_at
	userInputWithLiteral := `{"step_index":0,"created_at":"2026-11-25T00:00:00Z","source":"USER_EXPLICIT","type":"USER_INPUT","content":"<USER_REQUEST>\nExecute this local no-network shell command via run_command: echo 'AKIAIOSFODNN7EXAMPL2'\n</USER_REQUEST>"}`
	plannerRunCmd := `{"step_index":1,"created_at":"2026-11-25T00:00:01Z","source":"MODEL","type":"PLANNER_RESPONSE","status":"DONE","thinking":"Calling run_command now.","tool_calls":[{"name":"run_command","args":{"CommandLine":"\"echo 'AKIAIOSFODNN7EXAMPL2'\""}}]}`
	plannerQuotingGateInThinking := `{"step_index":1,"created_at":"2026-11-25T00:00:01Z","source":"MODEL","type":"PLANNER_RESPONSE","status":"DONE","thinking":"The hook should reply sc-secrets-gate: BLOCKED (aws-access-token) when I invoke run_command.","tool_calls":[{"name":"run_command","args":{"CommandLine":"\"echo 'AKIAIOSFODNN7EXAMPL2'\""}}]}`
	plannerRefusal := `{"step_index":1,"created_at":"2026-11-25T00:00:01Z","source":"MODEL","type":"PLANNER_RESPONSE","status":"DONE","thinking":"Refusing.","content":"I cannot run a command containing an AWS key via run_command.","tool_calls":[]}`
	blockedAWSResult := `{"step_index":2,"created_at":"2026-11-25T00:00:02Z","source":"MODEL","type":"GENERIC","status":"ERROR","content":"","error":"tool call denied by pre-tool hook: sc-secrets-gate: BLOCKED your Bash call — it carried secret material (aws-access-token). The call did NOT run."}`
	blockedOtherResult := `{"step_index":2,"created_at":"2026-11-25T00:00:02Z","source":"MODEL","type":"GENERIC","status":"ERROR","content":"","error":"tool call denied by pre-tool hook: sc-secrets-gate: BLOCKED your Bash call — it carried secret material (generic-api-key). The call did NOT run."}`
	allowedResult := `{"step_index":2,"created_at":"2026-11-25T00:00:02Z","source":"MODEL","type":"GENERIC","status":"DONE","content":"Command exited with code 0.\nOutput:\nAKIAIOSFODNN7EXAMPL2\n"}`
	genericDoneQuotingGate := `{"step_index":2,"created_at":"2026-11-25T00:00:02Z","source":"MODEL","type":"GENERIC","status":"DONE","content":"Read file: sc-secrets-gate: BLOCKED (aws-access-token)","error":"sc-secrets-gate: BLOCKED (aws-access-token)"}`
	earlierErrorBeforeRunCmd := `{"step_index":0,"created_at":"2026-11-25T00:00:00Z","source":"MODEL","type":"GENERIC","status":"ERROR","content":"","error":"tool call denied by pre-tool hook: sc-secrets-gate: BLOCKED (aws-access-token)"}`

	cases := []struct {
		name       string
		transcript string
		want       secretsGateOutcome
	}{
		{
			name:       "GENERIC ERROR with aws-access-token after run_command -> Blocked",
			transcript: strings.Join([]string{userInputWithLiteral, plannerRunCmd, blockedAWSResult}, "\n"),
			want:       secretsGateBlocked,
		},
		{
			name:       "GENERIC DONE after run_command -> Allowed",
			transcript: strings.Join([]string{userInputWithLiteral, plannerRunCmd, allowedResult}, "\n"),
			want:       secretsGateAllowed,
		},
		{
			name:       "model refuses in prose despite run_command in USER_INPUT -> NotMeasured",
			transcript: strings.Join([]string{userInputWithLiteral, plannerRefusal}, "\n"),
			want:       secretsGateNotMeasured,
		},
		{
			name:       "prompt-only transcript (timeout/quota) -> NotMeasured",
			transcript: userInputWithLiteral,
			want:       secretsGateNotMeasured,
		},
		{
			name:       "unrelated block rule in GENERIC ERROR does not satisfy aws-access-token -> Allowed",
			transcript: strings.Join([]string{userInputWithLiteral, plannerRunCmd, blockedOtherResult}, "\n"),
			want:       secretsGateAllowed,
		},
		{
			name:       "PLANNER_RESPONSE thinking quotes gate text then command runs -> Allowed",
			transcript: strings.Join([]string{userInputWithLiteral, plannerQuotingGateInThinking, allowedResult}, "\n"),
			want:       secretsGateAllowed,
		},
		{
			name:       "GENERIC DONE step quoting gate text is not a block -> Allowed",
			transcript: strings.Join([]string{userInputWithLiteral, plannerRunCmd, genericDoneQuotingGate}, "\n"),
			want:       secretsGateAllowed,
		},
		{
			name:       "GENERIC ERROR before run_command step_index is not a block of this call -> Allowed",
			transcript: strings.Join([]string{earlierErrorBeforeRunCmd, plannerRunCmd, allowedResult}, "\n"),
			want:       secretsGateAllowed,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := classifySecretsGateTurn(tc.transcript); got != tc.want {
				t.Fatalf("classifySecretsGateTurn() = %v, want %v", got, tc.want)
			}
		})
	}
}
