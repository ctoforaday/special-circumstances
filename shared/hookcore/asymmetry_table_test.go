package hookcore_test

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

// asymmetryRow declares one harness asymmetry the repository relies on, paired directly with its
// Claude Code and Antigravity live verification functions. Storing the two func(*testing.T) values
// directly in this table makes Claude/Antigravity pairing compiler-checked and eliminates any
// string-to-test lookup table or AST scanner.
type asymmetryRow struct {
	ID              string
	Claim           string
	ClaudeBehaviour string
	AgyBehaviour    string
	RelyingOn       []string
	Upstream        []string
	Claude          func(*testing.T)
	Agy             func(*testing.T)
}

// harnessAsymmetries is the single authoritative catalogue of cross-harness behavioural
// asymmetries relied on by this repository (#1926). Adding a new workaround for either harness
// requires adding a paired row here.
var harnessAsymmetries = []asymmetryRow{
	{
		ID:              "session-start-on-resume",
		Claim:           "Resuming an existing session fires SessionStart(source=resume) on Claude Code, whereas Antigravity does not fire SessionStart on --conversation (even across a fresh language server); only PreInvocation with invocationNum==0 fires.",
		ClaudeBehaviour: "Fires SessionStart with source=\"resume\" and the resumed session_id on `claude --resume <id>`.",
		AgyBehaviour:    "Fires SessionStart on turn 1, then fires no SessionStart on `agy --conversation <id>` (fresh language server); only PreInvocation with invocationNum==0 fires.",
		RelyingOn: []string{
			"plugins/prosthetic-conscience/tools/internal/sessionstart/main.go",
			"plugins/gray-area/tools/cmd/gray-area-capture/link.go",
			"shared/hookcore/hookcore.go",
			"docs/local-development.md",
		},
		Claude: asymSessionStartOnResumeClaude,
		Agy:    asymSessionStartOnResumeAgy,
	},
	{
		ID:              "hook-plugin-root-env",
		Claim:           "Claude Code passes CLAUDE_PLUGIN_ROOT (and no PLUGIN_ROOT) to hook processes; Antigravity passes PLUGIN_ROOT, PLUGIN_DATA, and ANTIGRAVITY_CONVERSATION_ID (and no CLAUDE_PLUGIN_ROOT) on both SessionStart and PreInvocation.",
		ClaudeBehaviour: "Hook environment includes CLAUDE_PLUGIN_ROOT and omits PLUGIN_ROOT.",
		AgyBehaviour:    "Both SessionStart and PreInvocation hook environments include PLUGIN_ROOT, PLUGIN_DATA, and ANTIGRAVITY_CONVERSATION_ID, and omit CLAUDE_PLUGIN_ROOT.",
		RelyingOn: []string{
			"shared/hookcore/hookcore.go",
			"shared/hookfailures/failures.go",
		},
		Claude: asymHookPluginRootEnvClaude,
		Agy:    asymHookPluginRootEnvAgy,
	},
	{
		ID:              "hook-wake-credentials",
		Claim:           "Antigravity hooks receive no language-server address or CSRF token (only ANTIGRAVITY_CONVERSATION_ID), while the agent shell receives all three; Claude Code advertises every running session in ~/.claude/sessions/<pid>.json visible directly to hooks with no shell link step.",
		ClaudeBehaviour: "Creates ~/.claude/sessions/<pid>.json containing pid, sessionId, and cwd readable during SessionStart with no shell link command.",
		AgyBehaviour:    "Hook processes never receive ANTIGRAVITY_LS_ADDRESS or ANTIGRAVITY_CSRF_TOKEN (only ANTIGRAVITY_CONVERSATION_ID); the agent's run_command shell receives all three.",
		RelyingOn: []string{
			"plugins/gray-area/tools/cmd/gray-area-capture/link.go",
			"plugins/gray-area/tools/internal/peer/peer.go",
			"plugins/gray-area/tools/internal/telecli/link.go",
			"docs/local-development.md",
		},
		Upstream: []string{"google-antigravity/antigravity-cli#1741"},
		Claude:   asymHookWakeCredentialsClaude,
		Agy:      asymHookWakeCredentialsAgy,
	},
	{
		ID:              "standalone-server-daemon-wake-route",
		Claim:           "A standalone Claude Code session advertises its live session file (~/.claude/sessions/<pid>.json) and removes it on exit; a standalone non-daemon Antigravity session (measured via `agy -p` as proxy) writes no wake-credential discovery file under ~/.gemini even while alive with ANTIGRAVITY_LS_ADDRESS and ANTIGRAVITY_CSRF_TOKEN in its shell.",
		ClaudeBehaviour: "Writes ~/.claude/sessions/<pid>.json matching pid, sessionId, and cwd while alive, and removes that file when the standalone process exits.",
		AgyBehaviour:    "While alive with ANTIGRAVITY_LS_ADDRESS and ANTIGRAVITY_CSRF_TOKEN in its shell, standalone `agy -p` leaves zero files under ~/.gemini containing either wake credential.",
		RelyingOn: []string{
			"plugins/gray-area/tools/internal/telecli/link.go",
			"plugins/gray-area/tools/internal/telecli/send.go",
			"plugins/gray-area/tools/internal/peer/agy_proc_linux.go",
		},
		Claude: asymStandaloneWakeRouteClaude,
		Agy:    asymStandaloneWakeRouteAgy,
	},
	{
		ID:              "hooks-manifest-location",
		Claim:           "Antigravity loads hooks only from the plugin root hooks.json (.antigravity-plugin/hooks.json is never loaded and a hooks field in .antigravity-plugin/plugin.json is ignored); Claude Code loads hooks from .claude-plugin/hooks.json referenced by .claude-plugin/plugin.json.",
		ClaudeBehaviour: "Reads \"hooks\": \"./.claude-plugin/hooks.json\" from .claude-plugin/plugin.json, validates it cleanly, and invokes those hooks at runtime.",
		AgyBehaviour:    "Rejects .antigravity-plugin/hooks.json as never loaded and ignores \"hooks\" in .antigravity-plugin/plugin.json; neither fires hooks at runtime unless root hooks.json is present.",
		RelyingOn: []string{
			"scripts/validatejson/hooks.go",
			"scripts/pluginparity/main.go",
			"plans/dual-target-claude-antigravity.md",
		},
		Claude: asymHooksManifestLocationClaude,
		Agy:    asymHooksManifestLocationAgy,
	},
	{
		ID:              "plugin-manifest-accepted-fields",
		Claim:           "Antigravity's .antigravity-plugin/plugin.json accepts only identity and display fields and fails strict validation on Claude-only fields (dependencies, hooks); Claude Code's .claude-plugin/plugin.json accepts hooks and agents.",
		ClaudeBehaviour: "`claude plugin validate` accepts \"hooks\" and \"agents\" in .claude-plugin/plugin.json.",
		AgyBehaviour:    "`agy plugin validate --strict` emits unknown-field warnings and exits non-zero when .antigravity-plugin/plugin.json carries \"dependencies\" or \"hooks\".",
		RelyingOn: []string{
			"scripts/marketplacegen/main.go",
			"scripts/pluginparity/main.go",
			"plans/dual-target-claude-antigravity.md",
		},
		Claude: asymPluginManifestAcceptedFieldsClaude,
		Agy:    asymPluginManifestAcceptedFieldsAgy,
	},
	{
		ID:              "failed-shell-command-hook",
		Claim:           "A failing shell command fires PostToolUseFailure with a non-empty error on Claude Code, whereas Antigravity fires no PostToolUseFailure and its PostToolUse carries an empty error and no result field even though the transcript records the non-zero exit.",
		ClaudeBehaviour: "Fires PostToolUseFailure with a non-empty error string when a Bash command exits non-zero.",
		AgyBehaviour:    "Fires no PostToolUseFailure on a non-zero run_command exit; PostToolUse carries error:\"\" and no result/tool_response field while transcript.jsonl records exit code 2.",
		RelyingOn: []string{
			"shared/hookcore/failedsteps.go",
			"plugins/frank-exchange-of-views/tools/internal/hookcmd/agysitting.go",
			"plugins/prosthetic-conscience/tools/internal/sessionstart/main.go",
		},
		Claude: asymFailedShellCommandHookClaude,
		Agy:    asymFailedShellCommandHookAgy,
	},
	{
		ID:              "workspace-plugin-shadowing",
		Claim:           "On Antigravity, a workspace .agents/plugins.json completely shadows the installed plugin copy under ~/.gemini/config/plugins; on Claude Code, a workspace .agents/plugins.json is ignored and does not shadow a loaded plugin.",
		ClaudeBehaviour: "Ignores .agents/plugins.json in the workspace; hooks from the explicitly loaded plugin still fire.",
		AgyBehaviour:    "Loads the workspace plugin named in .agents/plugins.json instead of the installed copy in ~/.gemini/config/plugins, so the installed copy's hooks never fire.",
		RelyingOn: []string{
			"scripts/pluginparity/main.go",
			"AGENTS.md",
			"shared/hookcore/live_system_install_test.go",
		},
		Claude: asymWorkspacePluginShadowingClaude,
		Agy:    asymWorkspacePluginShadowingAgy,
	},
	{
		ID:              "remote-control-serve-help",
		Claim:           "`claude remote-control --help` prints usage text and exits 0 immediately; `agy remote-control serve --help` ignores --help and enters the serve sign-in/startup path (while `agy remote-control --help` prints usage).",
		ClaudeBehaviour: "`claude remote-control --help` prints help text and exits 0.",
		AgyBehaviour:    "`agy remote-control --help` prints usage and exits 0, whereas `agy remote-control serve --help` ignores --help and fails sign-in (\"not signed in\") instead of printing usage.",
		RelyingOn: []string{
			"plugins/gray-area/tools/internal/peer/agy_proc_linux.go",
		},
		Upstream: []string{"google-antigravity/antigravity-cli#1742"},
		Claude:   asymRemoteControlServeHelpClaude,
		Agy:      asymRemoteControlServeHelpAgy,
	},
	{
		ID:              "system-message-metadata-panic",
		Claim:           "Returning a systemMessage carrying metadata.user_markdown from SessionStart panics Antigravity's hookcaller and drops the message (while the same systemMessage without metadata is delivered); Claude Code accepts systemMessage with metadata on SessionStart cleanly.",
		ClaudeBehaviour: "Accepts {\"systemMessage\":\"...\",\"metadata\":{\"user_markdown\":\"...\"}} from SessionStart with exit_code 0 and completes cleanly.",
		AgyBehaviour:    "Delivers systemMessage without metadata into transcript.jsonl, but panics hookcaller with \"type mismatch: cannot convert string to message\" and drops the message when metadata.user_markdown is present.",
		RelyingOn: []string{
			"shared/hookcore/hookcore.go",
			"scripts/fetchbingen/fetch-bin.sh",
			"plans/dual-target-claude-antigravity.md",
		},
		Upstream: []string{"google-antigravity/antigravity-cli#1774"},
		Claude:   asymSystemMessageMetadataPanicClaude,
		Agy:      asymSystemMessageMetadataPanicAgy,
	},
	{
		ID:              "workflow-child-attribution",
		Claim:           "Claude Code's Agent subagents fire SubagentStart/SubagentStop and carry agent_id on child tool hooks; Antigravity's run_workflow child agents fire neither and carry neither parentConversationId nor agentName.",
		ClaudeBehaviour: "Fires SubagentStart and SubagentStop with non-empty agent_id and stamps agent_id on the child's PostToolUse hooks.",
		AgyBehaviour:    "Fires no SubagentStart or SubagentStop for run_workflow child agents, and child PostToolUse hooks carry empty agent_id/agent_type and no parentConversationId/agentName.",
		RelyingOn: []string{
			"plugins/frank-exchange-of-views/tools/internal/hookcmd/agysitting.go",
			"plugins/frank-exchange-of-views/skills/research-protocol/scripts/debate.py",
		},
		Claude: asymWorkflowChildAttributionClaude,
		Agy:    asymWorkflowChildAttributionAgy,
	},
	{
		ID:              "finish-tool-overwrite",
		Claim:           "Claude Code honors PreToolUse updatedInput on StructuredOutput; Antigravity ignores PreToolUse overwrite on the run_workflow finish tool.",
		ClaudeBehaviour: "Replaces the StructuredOutput payload with PreToolUse hookSpecificOutput.updatedInput.",
		AgyBehaviour:    "Fires PreToolUse for RawToolName=\"finish\" but discards overwrite and returns the child agent's original answer.",
		RelyingOn: []string{
			"plugins/frank-exchange-of-views/skills/research-protocol/scripts/debate.py",
		},
		Claude: asymFinishToolOverwriteClaude,
		Agy:    asymFinishToolOverwriteAgy,
	},
	{
		ID:              "post-tool-use-feedback",
		Claim:           "Claude Code delivers PostToolUse hookSpecificOutput.additionalContext to the agent; Antigravity PostToolUse cannot return feedback to the agent (silently discards overwriteResult).",
		ClaudeBehaviour: "Delivers hookSpecificOutput.additionalContext returned by PostToolUse directly into the model's context.",
		AgyBehaviour:    "Discards overwriteResult returned by PostToolUse for run_command; the model only sees the tool's real stdout and never sees the hook's marker.",
		RelyingOn: []string{
			"shared/hookcore/hookcore.go",
			"plugins/frank-exchange-of-views/tools/internal/hookcmd/post.go",
			"plugins/frank-exchange-of-views/hooks/README.md",
		},
		Claude: asymPostToolUseFeedbackClaude,
		Agy:    asymPostToolUseFeedbackAgy,
	},
	{
		ID:              "unknown-pretool-hook-stdout",
		Claim:           "Non-schema JSON printed to stdout by a PreToolUse hook is ignored by Claude Code (which runs the tool call), whereas Antigravity's strict protojson unmarshaler fails the tool step with \"failed to unmarshal result from hook\".",
		ClaudeBehaviour: "Ignores non-schema JSON {\"not_claude_schema\":\"...\"} from PreToolUse and executes the Bash command normally.",
		AgyBehaviour:    "Fails the run_command step with \"failed to unmarshal result from hook\" in transcript_full.jsonl and blocks command execution.",
		RelyingOn: []string{
			"plugins/prosthetic-conscience/tools/internal/pretooluse/main.go",
			"scripts/fetchbingen/fetch-bin.sh",
			"shared/hookcore/failedsteps.go",
		},
		Claude: asymUnknownPreToolHookStdoutClaude,
		Agy:    asymUnknownPreToolHookStdoutAgy,
	},
	{
		ID:              "strict-protojson-posttool-hook",
		Claim:           "Antigravity's strict protojson unmarshaler rejects unknown fields such as {\"systemMessage\":\"...\"} on PostToolUse hooks too (recording \"failed to unmarshal result from hook\"), whereas Claude Code tolerates non-schema JSON on PostToolUse.",
		ClaudeBehaviour: "Tolerates non-schema JSON on PostToolUse stdout and finishes cleanly without hook unmarshal errors.",
		AgyBehaviour:    "Records \"failed to unmarshal result from hook\" when a PostToolUse hook emits {\"systemMessage\":\"...\"}.",
		RelyingOn: []string{
			"shared/hookcore/hookcore.go",
			"scripts/fetchbingen/fetch-bin.sh",
			"plans/dual-target-claude-antigravity.md",
		},
		Claude: asymStrictProtojsonPostToolHookClaude,
		Agy:    asymStrictProtojsonPostToolHookAgy,
	},
	{
		ID:              "session-end-hook-event",
		Claim:           "Claude Code fires SessionEnd when a headless -p session finishes; Antigravity does not support SessionEnd (never invokes it on exit).",
		ClaudeBehaviour: "Fires SessionEnd (with reason=\"other\" in headless -p mode) when the session exits.",
		AgyBehaviour:    "Does not invoke a SessionEnd entry declared in hooks.json when the session exits.",
		RelyingOn: []string{
			"scripts/validatejson/hooks.go",
			"plugins/prosthetic-conscience/hooks.json",
			"plugins/prosthetic-conscience/.claude-plugin/hooks.json",
		},
		Claude: asymSessionEndHookEventClaude,
		Agy:    asymSessionEndHookEventAgy,
	},
	{
		ID:              "root-plugin-manifest-discovery",
		Claim:           "Antigravity accepts a plugin with only a root plugin.json; Claude Code rejects a plugin directory lacking .claude-plugin/plugin.json.",
		ClaudeBehaviour: "`claude plugin validate` fails with \"No manifest found in directory. Expected .claude-plugin/marketplace.json or .claude-plugin/plugin.json\".",
		AgyBehaviour:    "`agy plugin validate --strict` passes on a plugin containing only root plugin.json.",
		RelyingOn: []string{
			"scripts/pluginparity/main.go",
			"plans/dual-target-claude-antigravity.md",
		},
		Claude: asymRootPluginManifestDiscoveryClaude,
		Agy:    asymRootPluginManifestDiscoveryAgy,
	},
	{
		ID:              "agent-frontmatter-scalar-tools",
		Claim:           "Claude Code parses a scalar comma-separated string for `tools:` in agent frontmatter at runtime; Antigravity fails YAML unmarshaling into `[]string` and skips the agent.",
		ClaudeBehaviour: "Spawns a live subagent with scalar CSV `tools: Read, Bash` whose `Bash` tool executes cleanly, while a control subagent with `tools: Read` has no `Bash` tool.",
		AgyBehaviour:    "`agy plugin validate --strict` fails with \"cannot unmarshal !!str ... into []string\" and registers 0 agents.",
		RelyingOn: []string{
			"scripts/frontmatter/main.go",
			"scripts/pluginparity/harness_load_test.go",
		},
		Claude: asymAgentFrontmatterScalarToolsClaude,
		Agy:    asymAgentFrontmatterScalarToolsAgy,
	},
	{
		ID:              "agent-frontmatter-foreign-tool-names",
		Claim:           "Claude Code accepts foreign tool names in an agent's tools: frontmatter without failing execution; Antigravity passes both `agy plugin validate --strict` and `agy agents` on a Claude tool name (Read) but fails `agy -p --agent` executor construction with \"not found in registry\".",
		ClaudeBehaviour: "Executes an agent whose tools: lists Antigravity tool names (`view_file`, `run_command`) alongside `Bash` without failing agent construction.",
		AgyBehaviour:    "Both `agy plugin validate --strict` and `agy agents` accept an agent declaring `tools: [Read]`, but `agy -p --agent` fails with `tool \"Read\" not found in registry`.",
		RelyingOn: []string{
			"scripts/frontmatter/main.go",
			"scripts/agentgen/main.go",
			"scripts/pluginparity/harness_load_test.go",
		},
		Claude: asymAgentFrontmatterForeignToolNamesClaude,
		Agy:    asymAgentFrontmatterForeignToolNamesAgy,
	},
	{
		ID:              "skill-size-ceiling",
		Claim:           "Antigravity enforces a 24,000-byte body ceiling on SKILL.md (failing strict validation); Claude Code imposes no 24 KB limit on SKILL.md.",
		ClaudeBehaviour: "`claude plugin validate` accepts a 26,000-byte SKILL.md body with no size warning or error.",
		AgyBehaviour:    "`agy plugin validate --strict` warns \"only the first 24000 are loaded\" and exits non-zero.",
		RelyingOn: []string{
			"plans/dual-target-claude-antigravity.md",
			"plans/hide-non-user-facing-skills-slash-commands.md",
		},
		Claude: asymSkillSizeCeilingClaude,
		Agy:    asymSkillSizeCeilingAgy,
	},
	{
		ID:              "logged-out-session-start",
		Claim:           "Logged-out `claude -p` (empty isolated HOME with no credentials) still fires SessionStart hooks before exiting non-zero; logged-out `agy -p` blocks on interactive authentication and never fires SessionStart.",
		ClaudeBehaviour: "Fires SessionStart in an unauthenticated HOME before exiting non-zero.",
		AgyBehaviour:    "Prints an interactive OAuth authentication prompt, waits for login, and fires zero SessionStart hooks.",
		RelyingOn: []string{
			"scripts/installshape/contract_test.go",
			"scripts/installshape/env_test.go",
		},
		Claude: asymLoggedOutSessionStartClaude,
		Agy:    asymLoggedOutSessionStartAgy,
	},
	{
		ID:              "transcript-model-field",
		Claim:           "Both Claude Code (`message.model`) and Antigravity (`PLANNER_RESPONSE.model`) record served model IDs in their session transcripts (read by servedmodel.Observe), while `agy -p --output-format json` still omits any model field (unlike `claude -p --output-format json` modelUsage).",
		ClaudeBehaviour: "Records non-empty `message.model` on assistant turns in `~/.claude/projects/*/*.jsonl` and non-empty `modelUsage` in `claude -p --output-format json`.",
		AgyBehaviour:    "Records non-empty top-level `model` on `PLANNER_RESPONSE` lines in `transcript.jsonl`, while `agy -p --output-format json` still omits `model` and `modelUsage`.",
		RelyingOn: []string{
			"plugins/frank-exchange-of-views/tools/internal/servedmodel/servedmodel.go",
			"plugins/frank-exchange-of-views/tools/internal/cost/cost.go",
			"plans/cross-model-seats.md",
		},
		Claude: asymTranscriptModelFieldClaude,
		Agy:    asymTranscriptModelFieldAgy,
	},
	{
		ID:              "bash-oauth-token-stripping",
		Claim:           "Claude Code strips CLAUDE_CODE_OAUTH_TOKEN from a Bash tool's environment so lent subscription access tokens cannot leak into archived transcripts; Antigravity passes an inherited CLAUDE_CODE_OAUTH_TOKEN through to run_command.",
		ClaudeBehaviour: "A Bash tool invoked inside `claude -p` sees an unset/empty CLAUDE_CODE_OAUTH_TOKEN even when the parent process authenticated with it.",
		AgyBehaviour:    "A run_command tool invoked inside `agy -p` inherits CLAUDE_CODE_OAUTH_TOKEN from the parent process environment.",
		RelyingOn: []string{
			"scripts/universe.sh",
		},
		Claude: asymBashOAuthTokenStrippingClaude,
		Agy:    asymBashOAuthTokenStrippingAgy,
	},
	{
		ID:              "thinking-summaries-persistence",
		Claim:           "Antigravity persists model reasoning (`thinking`) into transcript.jsonl natively with default settings; Claude Code defaults `showThinkingSummaries` to off and omits readable thinking summaries from transcripts unless explicitly enabled.",
		ClaudeBehaviour: "With default settings (no showThinkingSummaries), Claude Code persists no non-empty thinking text in the session transcript.",
		AgyBehaviour:    "With default settings, Antigravity persists a non-empty `thinking` field on PLANNER_RESPONSE records in transcript.jsonl.",
		RelyingOn: []string{
			"plugins/prosthetic-conscience/tools/internal/doctor/capture.go",
			"plugins/frank-exchange-of-views/tools/internal/seatturn/seatturn.go",
		},
		Claude: asymThinkingSummariesPersistenceClaude,
		Agy:    asymThinkingSummariesPersistenceAgy,
	},
	{
		ID:              "post-tool-delivery-cap",
		Claim:           "Claude Code delivers a 9,000-byte (DeliveryBudget) PostToolUse additionalContext inline but spills a 12,000-byte payload (>10,000 cap) to tool-results/*-additionalContext.txt with a <persisted-output> stub; Antigravity PreInvocation injectSteps systemMessage delivers a 12,000-byte payload intact into transcript_full.jsonl (PostToolUse cannot deliver text at all, covered by post-tool-use-feedback).",
		ClaudeBehaviour: "Delivers 9,000-byte PostToolUse additionalContext inline with zero spill file, and spills 12,000-byte additionalContext to tool-results/*-additionalContext.txt replacing inline text with <persisted-output>.",
		AgyBehaviour:    "Delivers a 12,000-byte PreInvocation injectSteps systemMessage intact with head, middle, and tail markers present on a single >=12,000-byte step in transcript_full.jsonl.",
		RelyingOn: []string{
			"plugins/frank-exchange-of-views/tools/internal/sittingwrite/delivery.go",
			"shared/hookcore/hookcore.go",
		},
		Claude: asymPostToolDeliveryCapClaude,
		Agy:    asymPostToolDeliveryCapAgy,
	},
	{
		ID:              "strict-protojson-postinvocation-hook",
		Claim:           "Antigravity's strict protojson unmarshaler rejects unknown fields such as {\"systemMessage\":\"...\"} on PostInvocation hooks (recording \"failed to unmarshal result from hook\"), whereas Claude Code tolerates {\"systemMessage\":\"...\"} on SubagentStop without error.",
		ClaudeBehaviour: "Executes SubagentStop hook emitting {\"systemMessage\":\"...\"} exactly once (stopCount==1) with clean JSON parse/validation in --debug-file.",
		AgyBehaviour:    "Logs post-invocation hook \"failed to unmarshal result from hook\" when PostInvocation emits {\"systemMessage\":\"...\"}.",
		RelyingOn: []string{
			"scripts/fetchbingen/fetch-bin.sh",
			"shared/hookcore/hookcore.go",
		},
		Claude: asymStrictProtojsonPostInvocationHookClaude,
		Agy:    asymStrictProtojsonPostInvocationHookAgy,
	},
	{
		ID:              "strict-protojson-stop-hook",
		Claim:           "Antigravity's strict protojson unmarshaler rejects unknown fields such as {\"systemMessage\":\"...\"} on Stop hooks (recording \"failed to unmarshal result from hook\"), whereas Claude Code accepts {\"systemMessage\":\"...\"} on Stop cleanly.",
		ClaudeBehaviour: "Executes Stop hook emitting {\"systemMessage\":\"...\"} cleanly without unmarshal failure.",
		AgyBehaviour:    "Logs stop hook \"failed to unmarshal result from hook\" when Stop emits {\"systemMessage\":\"...\"}.",
		RelyingOn: []string{
			"scripts/fetchbingen/fetch-bin.sh",
			"shared/hookcore/hookcore.go",
		},
		Claude: asymStrictProtojsonStopHookClaude,
		Agy:    asymStrictProtojsonStopHookAgy,
	},
	{
		ID:              "headless-workspace-paths",
		Claim:           "Both Claude Code (`cwd`) and headless `agy -p` (`workspacePaths`) populate the workspace path in hook stdin payloads across SessionStart and PreInvocation.",
		ClaudeBehaviour: "SessionStart stdin JSON carries `cwd` equal to the session working directory.",
		AgyBehaviour:    "Headless `agy -p` SessionStart and PreInvocation stdin JSON both carry a non-empty `workspacePaths` slice containing the session working directory.",
		RelyingOn: []string{
			"shared/hookcore/hookcore.go",
			"plans/dual-target-claude-antigravity.md",
			"plans/antigravity-hooks-empirical-proof.md",
		},
		Claude: asymHeadlessWorkspacePathsClaude,
		Agy:    asymHeadlessWorkspacePathsAgy,
	},
	{
		ID:              "pretool-deny-transcript-step",
		Claim:           "A PreToolUse deny is recorded in Claude Code's transcript as a `tool_result` block with `is_error: true` carrying the deny reason, whereas Antigravity records it as a `GENERIC` step with `status: \"ERROR\"` and the deny reason in `error`.",
		ClaudeBehaviour: "Blocks Bash execution and records a `tool_result` with `is_error: true` containing the PreToolUse deny reason.",
		AgyBehaviour:    "Blocks run_command execution and records a `GENERIC` step with `status: \"ERROR\"` whose `error` field carries the PreToolUse deny reason.",
		RelyingOn: []string{
			"shared/hookcore/live_system_install_test.go",
		},
		Claude: asymPreToolDenyTranscriptStepClaude,
		Agy:    asymPreToolDenyTranscriptStepAgy,
	},
	{
		ID:              "plugin-root-inheritance-precedence",
		Claim:           "Antigravity's `run_command` shell does not populate `PLUGIN_ROOT` on its own (requiring skill fallbacks) yet inherits an existing `PLUGIN_ROOT` from its parent process; a Claude Code session spawned under that shell sees both its own `CLAUDE_PLUGIN_ROOT` and the inherited `PLUGIN_ROOT`, so `pluginroot.Resolve` must prefer `CLAUDE_PLUGIN_ROOT` over `PLUGIN_ROOT`.",
		ClaudeBehaviour: "When parent env sets `PLUGIN_ROOT`, Claude SessionStart hooks receive both `CLAUDE_PLUGIN_ROOT` (the Claude plugin) and the inherited `PLUGIN_ROOT`.",
		AgyBehaviour:    "Antigravity `run_command` sees empty `PLUGIN_ROOT` when unset in parent, and inherits `PLUGIN_ROOT` verbatim when set in parent.",
		RelyingOn: []string{
			"shared/pluginroot/pluginroot.go",
			"plugins/frank-exchange-of-views/skills/research/SKILL.md",
		},
		Claude: asymPluginRootInheritancePrecedenceClaude,
		Agy:    asymPluginRootInheritancePrecedenceAgy,
	},
	{
		ID:              "workflow-resume-from-run-id",
		Claim:           "Claude Code's `Workflow` tool accepts `resumeFromRunId` and replays cached `agent()` steps without re-executing them; Antigravity's `run_workflow` tool rejects `ResumeFromRunId` at tool-schema validation and carries no `resumeFromRunId`/`resume_from_run_id` parameter.",
		ClaudeBehaviour: "Executes a native `Workflow` script with a cached `agent()` step that appends a tick, replays it via `resumeFromRunId` returning the cached result with 0 new ticks (1 total), and verifies an uncached control call without `resumeFromRunId` appends a second tick (2 total).",
		AgyBehaviour:    "Executes `run_workflow` cleanly, rejects `ResumeFromRunId` at runtime tool validation (`additional properties 'ResumeFromRunId' not allowed`), and its isolated builtin `run-workflow/SKILL.md` documents `ScriptPath`/`Args` with no case-insensitive `resumeFromRunId`/`resume_from_run_id` parameter.",
		RelyingOn: []string{
			"scripts/universe.sh",
		},
		Claude: asymWorkflowResumeFromRunIDClaude,
		Agy:    asymWorkflowResumeFromRunIDAgy,
	},
	{
		ID:              "post-compact-cannot-inject",
		Claim:           "On a real `/compact` followed by a post-compact turn, Claude Code's `PostCompact` hook fires but cannot inject `hookSpecificOutput.additionalContext` into model-visible context (while `SessionStart(source=compact)` does); Antigravity does not support `PostCompact` at all.",
		ClaudeBehaviour: "Real `/compact` followed by turn 3 fires both `SessionStart(source=compact)` and `PostCompact`; the session's transcript `hook_additional_context` carries `SessionStart`'s marker and omits `PostCompact`'s marker.",
		AgyBehaviour:    "`agy plugin validate --strict` rejects `PostCompact` with `unknown event \"PostCompact\"`, and runtime still loads the plugin and fires `SessionStart` cleanly without crashing.",
		RelyingOn: []string{
			"plugins/prosthetic-conscience/tools/internal/postcompactobserve/main.go",
			"plugins/prosthetic-conscience/requirements.json",
			"plugins/prosthetic-conscience/README.md",
		},
		Claude: asymPostCompactCannotInjectClaude,
		Agy:    asymPostCompactCannotInjectAgy,
	},
	{
		ID:              "tool-search-lazy-loading",
		Claim:           "Claude Code registers `ToolSearch` in its session tool set for deferred tool loading; Antigravity loads tools directly without `ToolSearch` and fails `agy -p --agent` with `not found in registry` when an agent declares `ToolSearch`.",
		ClaudeBehaviour: "Lists `ToolSearch` in the `tools` array of `claude -p --output-format stream-json --verbose` `system/init`.",
		AgyBehaviour:    "Runs an agent with native `tools: [read_file]` cleanly, but fails `agy -p --agent` with `not found in registry` when an agent declares `tools: [ToolSearch]`.",
		RelyingOn: []string{
			"scripts/agentgen/main.go",
		},
		Claude: asymToolSearchLazyLoadingClaude,
		Agy:    asymToolSearchLazyLoadingAgy,
	},
	{
		ID:              "hooks-json-unknown-keys",
		Claim:           "In hooks manifests carrying top-level `_comment` and `description`, Claude Code passes `claude plugin validate`, runs the hooks, accepts `description`, and emits a runtime debug warning ignoring `_comment`; Antigravity rejects both `_comment` and `description` in root `hooks.json` with `failed to parse hooks` and loads zero hooks from that file.",
		ClaudeBehaviour: "Passes `claude plugin validate`, runs `SessionStart`, logs `unknown key \"_comment\" ignored` in `--debug-file`, and accepts `description` without warning.",
		AgyBehaviour:    "Logs `failed to parse hooks` in `cli.log` and fires zero hooks for both `_comment` and `description` top-level string keys in root `hooks.json` (while a clean sibling plugin fires `SessionStart`).",
		RelyingOn: []string{
			"scripts/validatejson/main.go",
			"scripts/validatejson/hooks.go",
			"shared/hookcore/live_shipped_plugins_test.go",
		},
		Claude: asymHooksJSONUnknownKeysClaude,
		Agy:    asymHooksJSONUnknownKeysAgy,
	},
	{
		ID:              "agy-oauth-refresh-threshold",
		Claim:           "Antigravity enforces a 5-minute OAuth refresh threshold (`shared/testhome.AgyRefreshThreshold`): a refresh-stripped staged token succeeds at 315s (>5m) remaining but hangs prompting for OAuth login at 240s (<5m); Claude Code accepts a lent `CLAUDE_CODE_OAUTH_TOKEN` via environment with no credentials file on disk.",
		ClaudeBehaviour: "Runs `claude -p` cleanly with `CLAUDE_CODE_OAUTH_TOKEN` in env and no `~/.claude/.credentials.json` file on disk.",
		AgyBehaviour:    "Succeeds with a refresh-stripped staged token at 315s remaining, and prompts `Authentication required` at 240s remaining.",
		RelyingOn: []string{
			"shared/testhome/liveauth.go",
		},
		Claude: asymAgyOAuthRefreshThresholdClaude,
		Agy:    asymAgyOAuthRefreshThresholdAgy,
	},
}

func halfSkipError(label string, skipped bool) error {
	if !skipped {
		return nil
	}
	return fmt.Errorf("%s called t.Skip; paired asymmetry halves must FAIL rather than skip", label)
}

func runAsymmetryHalf(t *testing.T, label string, fn func(*testing.T)) {
	t.Helper()
	var skipped bool
	t.Run(label, func(sub *testing.T) {
		sub.Cleanup(func() { skipped = sub.Skipped() })
		fn(sub)
	})
	if err := halfSkipError(label, skipped); err != nil {
		t.Error(err)
	}
}

func logLiveAsymmetryProvenance(t *testing.T) {
	t.Helper()
	headStr := "unknown"
	if headOut, err := exec.Command("git", "rev-parse", "HEAD").Output(); err == nil {
		headStr = strings.TrimSpace(string(headOut))
	}
	treeCleanStr := "unknown"
	if dirtyOut, err := exec.Command("git", "status", "--porcelain").Output(); err == nil {
		if strings.TrimSpace(string(dirtyOut)) == "" {
			treeCleanStr = "true"
		} else {
			treeCleanStr = "false"
		}
	}
	claudeOut, _ := exec.Command(requireClaudeBin(t), "--version").CombinedOutput()
	agyOut, _ := exec.Command(requireAgyBin(t), "--version").CombinedOutput()
	t.Logf("provenance: git_head=%s tree_clean=%s claude=%q agy=%q rows=%d",
		headStr,
		treeCleanStr,
		strings.TrimSpace(string(claudeOut)),
		strings.TrimSpace(string(agyOut)),
		len(harnessAsymmetries))
}

// TestLiveAsymmetries is the single entry point executing every paired row in harnessAsymmetries
// against the real Claude Code and Antigravity CLIs under HOOKCORE_LIVE=1.
func TestLiveAsymmetries(t *testing.T) {
	requireLiveAsymmetry(t)
	logLiveAsymmetryProvenance(t)
	for _, row := range harnessAsymmetries {
		t.Run(row.ID, func(t *testing.T) {
			runAsymmetryHalf(t, "Claude", row.Claude)
			runAsymmetryHalf(t, "Antigravity", row.Agy)
		})
	}
}

var reAsymmetryRowID = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

func funcShortName(fn func(*testing.T)) string {
	if fn == nil {
		return ""
	}
	full := runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name()
	if i := strings.LastIndex(full, "."); i >= 0 {
		return full[i+1:]
	}
	return full
}

func validateAsymmetryTable(rows []asymmetryRow, repoRoot string) []string {
	var problems []string
	if len(rows) == 0 {
		return []string{"harnessAsymmetries table is empty"}
	}
	seenIDs := map[string]bool{}
	seenFuncs := map[uintptr]string{}
	for i, r := range rows {
		label := r.ID
		if label == "" {
			label = fmt.Sprintf("row[%d]", i)
		}
		if !reAsymmetryRowID.MatchString(r.ID) {
			problems = append(problems, fmt.Sprintf("%s: id %q must be kebab-case", label, r.ID))
		}
		if seenIDs[r.ID] {
			problems = append(problems, fmt.Sprintf("%s: duplicate id", label))
		}
		seenIDs[r.ID] = true
		if strings.TrimSpace(r.Claim) == "" {
			problems = append(problems, fmt.Sprintf("%s: empty Claim", label))
		}
		if strings.TrimSpace(r.ClaudeBehaviour) == "" {
			problems = append(problems, fmt.Sprintf("%s: empty ClaudeBehaviour", label))
		}
		if strings.TrimSpace(r.AgyBehaviour) == "" {
			problems = append(problems, fmt.Sprintf("%s: empty AgyBehaviour", label))
		}
		if len(r.RelyingOn) == 0 {
			problems = append(problems, fmt.Sprintf("%s: RelyingOn must list at least one repo path", label))
		}
		for _, rel := range r.RelyingOn {
			if strings.TrimSpace(rel) == "" || filepath.IsAbs(rel) {
				problems = append(problems, fmt.Sprintf("%s: invalid RelyingOn path %q", label, rel))
				continue
			}
			if _, err := os.Stat(filepath.Join(repoRoot, filepath.FromSlash(rel))); err != nil {
				problems = append(problems, fmt.Sprintf("%s: RelyingOn path %q does not exist: %v", label, rel, err))
			}
		}
		if r.Claude == nil {
			problems = append(problems, fmt.Sprintf("%s: Claude func is nil", label))
		}
		if r.Agy == nil {
			problems = append(problems, fmt.Sprintf("%s: Agy func is nil", label))
		}
		if r.Claude != nil && r.Agy != nil {
			cPtr := reflect.ValueOf(r.Claude).Pointer()
			aPtr := reflect.ValueOf(r.Agy).Pointer()
			if cPtr == aPtr {
				problems = append(problems, fmt.Sprintf("%s: Claude and Agy point to the same function", label))
			}
			if prev, dup := seenFuncs[cPtr]; dup {
				problems = append(problems, fmt.Sprintf("%s: Claude func shared with %s", label, prev))
			} else {
				seenFuncs[cPtr] = label + "/Claude"
			}
			if prev, dup := seenFuncs[aPtr]; dup {
				problems = append(problems, fmt.Sprintf("%s: Agy func shared with %s", label, prev))
			} else {
				seenFuncs[aPtr] = label + "/Agy"
			}
			if name := funcShortName(r.Claude); !strings.HasSuffix(name, "Claude") {
				problems = append(problems, fmt.Sprintf("%s: Claude func %q must end with Claude (guards against Claude/Agy swap)", label, name))
			}
			if name := funcShortName(r.Agy); !strings.HasSuffix(name, "Agy") {
				problems = append(problems, fmt.Sprintf("%s: Agy func %q must end with Agy (guards against Claude/Agy swap)", label, name))
			}
		}
	}
	return problems
}

// findStrayAsymmetryTests is an advisory static scan over the `shared/` module catching
// `TestLiveAsymmetry*` functions that bypass the `harnessAsymmetries` table.
func findStrayAsymmetryTests(rootDir string) ([]string, error) {
	fset := token.NewFileSet()
	var stray []string
	err := filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			if d.Name() == ".git" || d.Name() == "vendor" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(d.Name(), "_test.go") {
			return nil
		}
		f, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(rootDir, path)
		for _, decl := range f.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv != nil {
				continue
			}
			if strings.HasPrefix(fn.Name.Name, "TestLiveAsymmetry") && fn.Name.Name != "TestLiveAsymmetries" {
				stray = append(stray, fmt.Sprintf("%s: stray top-level test %s bypasses harnessAsymmetries table", filepath.ToSlash(rel), fn.Name.Name))
			}
		}
		return nil
	})
	return stray, err
}

// TestAsymmetryTableIntegrity runs in normal CI (without HOOKCORE_LIVE=1) and enforces that every
// row in harnessAsymmetries is complete, distinct, points at real relying_on paths, uses unique
// correctly-sided Claude/Agy funcs, and that no unregistered TestLiveAsymmetry* function bypasses
// the table anywhere under shared/.
func TestAsymmetryTableIntegrity(t *testing.T) {
	repoRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolve repo root: %v", err)
	}
	for _, p := range validateAsymmetryTable(harnessAsymmetries, repoRoot) {
		t.Error(p)
	}
	stray, err := findStrayAsymmetryTests(filepath.Join(repoRoot, "shared"))
	if err != nil {
		t.Fatalf("scan stray asymmetry tests: %v", err)
	}
	for _, s := range stray {
		t.Error(s)
	}
}

// TestAsymmetryGateDefaultOff pins both M4a (TestLiveAsymmetries calls requireLiveAsymmetry) and
// M4b (requireLiveAsymmetry skips when HOOKCORE_LIVE is unset) without invoking claude or agy.
func TestAsymmetryGateDefaultOff(t *testing.T) {
	t.Setenv("HOOKCORE_LIVE", "")
	if liveAsymmetryEnabled() {
		t.Fatal("liveAsymmetryEnabled() returned true with HOOKCORE_LIVE unset")
	}
	var directSkipped bool
	t.Run("requireLiveAsymmetry skips when unset", func(sub *testing.T) {
		sub.Cleanup(func() { directSkipped = sub.Skipped() })
		requireLiveAsymmetry(sub)
		sub.Fatal("requireLiveAsymmetry did not skip with HOOKCORE_LIVE unset")
	})
	if !directSkipped {
		t.Fatal("expected requireLiveAsymmetry to skip when HOOKCORE_LIVE is unset")
	}
	var runnerSkipped bool
	t.Run("TestLiveAsymmetries skips when unset (M4a)", func(sub *testing.T) {
		sub.Setenv("PATH", "")
		sub.Cleanup(func() { runnerSkipped = sub.Skipped() })
		TestLiveAsymmetries(sub)
	})
	if !runnerSkipped {
		t.Fatal("expected TestLiveAsymmetries to skip when HOOKCORE_LIVE is unset (gate call removed?)")
	}
	t.Setenv("HOOKCORE_LIVE", "1")
	if !liveAsymmetryEnabled() {
		t.Fatal("liveAsymmetryEnabled() returned false with HOOKCORE_LIVE=1")
	}
}

// TestRunAsymmetryHalfFailsOnSkip pins M11 end to end: exercises the real runAsymmetryHalf helper
// in a child test binary subprocess and asserts a skipping half causes a non-zero exit naming
// "called t.Skip". Disabling the skip check in runAsymmetryHalf turns this test red.
func TestRunAsymmetryHalfFailsOnSkip(t *testing.T) {
	if os.Getenv("ASYMMETRY_SKIP_SUBPROC_PROBE") == "1" {
		runAsymmetryHalf(t, "probe", func(sub *testing.T) {
			sub.Skip("simulated skip")
		})
		return
	}
	if err := halfSkipError("probe", false); err != nil {
		t.Fatalf("halfSkipError(false) = %v, want nil", err)
	}
	if err := halfSkipError("probe", true); err == nil || !strings.Contains(err.Error(), "probe called t.Skip") {
		t.Fatalf("halfSkipError(true) = %v, want error naming 'probe called t.Skip'", err)
	}
	cmd := exec.Command(os.Args[0], "-test.run=^TestRunAsymmetryHalfFailsOnSkip$")
	cmd.Env = append(os.Environ(), "ASYMMETRY_SKIP_SUBPROC_PROBE=1")
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("expected runAsymmetryHalf subprocess to fail when half calls t.Skip, got exit 0:\n%s", out)
	}
	if !strings.Contains(string(out), "probe called t.Skip") {
		t.Fatalf("expected failure output to name 'probe called t.Skip', got:\n%s", out)
	}
}

func dummySampleClaude(*testing.T) {}
func dummySampleAgy(*testing.T)    {}
func dummyOtherClaude(*testing.T)  {}
func dummyOtherAgy(*testing.T)     {}

func TestValidateAsymmetryTableCatchesDefects(t *testing.T) {
	repoRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	valid := asymmetryRow{
		ID:              "sample-row",
		Claim:           "claim",
		ClaudeBehaviour: "claude",
		AgyBehaviour:    "agy",
		RelyingOn:       []string{"shared/hookcore/hookcore.go"},
		Claude:          dummySampleClaude,
		Agy:             dummySampleAgy,
	}
	second := asymmetryRow{
		ID:              "second-row",
		Claim:           "claim2",
		ClaudeBehaviour: "claude2",
		AgyBehaviour:    "agy2",
		RelyingOn:       []string{"shared/hookcore/hookcore.go"},
		Claude:          dummyOtherClaude,
		Agy:             dummyOtherAgy,
	}

	cases := []struct {
		name string
		rows []asymmetryRow
		want string
	}{
		{"empty table", nil, "table is empty"},
		{"bad id", []asymmetryRow{func() asymmetryRow { r := valid; r.ID = "Bad_ID"; return r }()}, "kebab-case"},
		{"duplicate id", []asymmetryRow{valid, func() asymmetryRow { r := second; r.ID = valid.ID; return r }()}, "duplicate id"},
		{"empty claim", []asymmetryRow{func() asymmetryRow { r := valid; r.Claim = ""; return r }()}, "empty Claim"},
		{"missing relying_on", []asymmetryRow{func() asymmetryRow { r := valid; r.RelyingOn = nil; return r }()}, "at least one repo path"},
		{"nonexistent relying_on", []asymmetryRow{func() asymmetryRow { r := valid; r.RelyingOn = []string{"no/such/file.go"}; return r }()}, "does not exist"},
		{"nil Claude func", []asymmetryRow{func() asymmetryRow { r := valid; r.Claude = nil; return r }()}, "Claude func is nil"},
		{"nil Agy func", []asymmetryRow{func() asymmetryRow { r := valid; r.Agy = nil; return r }()}, "Agy func is nil"},
		{"same func pointer", []asymmetryRow{func() asymmetryRow { r := valid; r.Agy = dummySampleClaude; return r }()}, "same function"},
		{"shared func across rows (M6b)", []asymmetryRow{valid, func() asymmetryRow { r := second; r.Claude = dummySampleClaude; return r }()}, "shared with"},
		{"swapped Claude/Agy funcs (M7)", []asymmetryRow{func() asymmetryRow { r := valid; r.Claude = dummySampleAgy; r.Agy = dummySampleClaude; return r }()}, "must end with Claude"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			probs := validateAsymmetryTable(tc.rows, repoRoot)
			joined := strings.Join(probs, "\n")
			if !strings.Contains(joined, tc.want) {
				t.Fatalf("expected error containing %q, got:\n%s", tc.want, joined)
			}
		})
	}

	t.Run("detects stray top-level TestLiveAsymmetry function in nested directory (M3c)", func(t *testing.T) {
		tmp := t.TempDir()
		nested := filepath.Join(tmp, "subpkg", "nested")
		if err := os.MkdirAll(nested, 0o755); err != nil {
			t.Fatal(err)
		}
		src := "package hookcore_test\nimport \"testing\"\nfunc TestLiveAsymmetryStray(t *testing.T) {}\n"
		if err := os.WriteFile(filepath.Join(nested, "stray_test.go"), []byte(src), 0o644); err != nil {
			t.Fatal(err)
		}
		stray, err := findStrayAsymmetryTests(tmp)
		if err != nil {
			t.Fatal(err)
		}
		if len(stray) != 1 || !strings.Contains(stray[0], "TestLiveAsymmetryStray") {
			t.Fatalf("expected stray detection, got %v", stray)
		}
	})
}
