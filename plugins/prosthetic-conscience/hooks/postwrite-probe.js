#!/usr/bin/env node
// Phase 1 harness spike: prove PostToolUse hooks fire — including inside subagents.
// Appends one line per Write/Edit to a log in the project dir. Never blocks the tool call.
//
// This is spike instrumentation. In later phases the real quality hook (qlty fmt/check,
// capability-gated) replaces it; see plans/claude-port-plan.md §3a and §3a' (env preflight).

const fs = require("fs");
const path = require("path");

let input = "";
process.stdin.on("data", (chunk) => (input += chunk));
process.stdin.on("end", () => {
  let tool = "unknown";
  let file = "unknown";
  try {
    const data = JSON.parse(input || "{}");
    tool = data.tool_name || tool;
    file =
      (data.tool_input && (data.tool_input.file_path || data.tool_input.path)) ||
      file;
  } catch (_e) {
    // malformed/empty stdin — still log the firing
  }

  const projectDir = process.env.CLAUDE_PROJECT_DIR || process.cwd();
  const logDir = path.join(projectDir, ".claude");
  const logFile = path.join(logDir, "prosthetic-conscience-hook.log");
  const line = `${new Date().toISOString()} PostToolUse ${tool} -> ${file}\n`;

  try {
    fs.mkdirSync(logDir, { recursive: true });
    fs.appendFileSync(logFile, line);
  } catch (_e) {
    // never fail the originating tool call because instrumentation broke
  }
  process.exit(0);
});
