// Package transcript reads the harness session transcript for the one fact the hook
// payload does not carry: WHICH FILES THIS SESSION WROTE.
//
// WHY THIS EXISTS. sc-checkpoint-seal fires at every seam and can compare the note's
// validation loop against the work the session actually did — but only if it knows what
// that work was. The hook payload carries session_id, transcript_path, cwd and the event
// fields, and no file list. `transcript_path` was declared by two hooks in this plugin and
// read by neither.
//
// MEASURED, not assumed (2026-07-31, against a real 1486-line transcript): each line is one
// JSON object; `message.content` is an array of blocks; a block with `"type":"tool_use"`
// carries `name` and `input`. Write and Edit put the path in `input.file_path`.
//
//	tool_use by name: Bash 230, Edit 53, Read 20, Write 19, …
//	entries carrying file_path: 92
//
// THE TRAP IN THAT LAST NUMBER: 92 > 53+19. `Read` carries `file_path` too, so filtering on
// the FIELD collects files the session merely looked at. The filter is on the tool NAME.
//
// # What this deliberately does not see
//
// Only Write, Edit and NotebookEdit count. A file written by a Bash heredoc, a build, a
// formatter or a generator is invisible here. That is a real gap and it is stated rather
// than glossed: a caller must treat "wrote nothing" as "wrote nothing THROUGH THE EDIT
// TOOLS", never as proof the session was idle. The same partial-coverage honesty the strike
// counter in #127 insists on.
package transcript

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// writeTools are the tools whose tool_use means "this session changed this file".
var writeTools = map[string]bool{
	"Write":        true,
	"Edit":         true,
	"NotebookEdit": true,
}

// entry is one transcript line. Only the fields this package needs are declared; the
// harness adds others freely and an unknown field must never be an error.
type entry struct {
	Message struct {
		// Content is an array of blocks for assistant turns and a plain string for
		// some user turns, so it is decoded lazily rather than typed.
		Content json.RawMessage `json:"content"`
	} `json:"message"`
}

type block struct {
	Type  string `json:"type"`
	Name  string `json:"name"`
	Input struct {
		FilePath string `json:"file_path"`
	} `json:"input"`
}

// Written returns the project-relative, slash-separated paths this session wrote, sorted
// and deduplicated.
//
// Best-effort by every measure: an unreadable transcript, a malformed line, or a path
// outside the project yields less output, never an error. This feeds a diagnostic line at a
// seam, and a diagnostic that can fail the seal would cost more than it explains.
func Written(transcriptPath, projectDir string) []string {
	paths, _ := Read(transcriptPath, projectDir)
	return paths
}

// Read is Written plus the reason it found nothing.
//
// TWO SILENCES THAT LOOK IDENTICAL AND ARE NOT. "the session wrote nothing" and "I could
// not read the transcript" both yield an empty slice, and a caller that cannot tell them
// apart reports a healthy session when its own reader is broken — silence in the
// FLATTERING direction. The transcript format is vendor-internal and version-unstable
// (gray-area's capture hook records what produced every manifest row for exactly that
// reason), so "I could not read it" is a live possibility, not a theoretical one.
//
// unreadable is empty when the transcript was read successfully, whatever it contained.
func Read(transcriptPath, projectDir string) (paths []string, unreadable string) {
	if transcriptPath == "" {
		return nil, "" // nothing was offered; that is not a failure
	}
	if projectDir == "" {
		return nil, ""
	}
	f, err := os.Open(transcriptPath)
	if err != nil {
		return nil, "transcript could not be opened: " + err.Error()
	}
	defer func() { _ = f.Close() }()

	absProject, err := filepath.Abs(projectDir)
	if err != nil {
		return nil, "project dir could not be resolved: " + err.Error()
	}

	seen := map[string]bool{}
	var out []string
	lines, unparsed := 0, 0

	// Read LINE BY LINE and decode each line on its own. Two failure modes rule out the
	// obvious alternatives, and both were measured rather than reasoned about:
	//
	//   bufio.Scanner — transcript lines carry whole tool results and routinely exceed
	//   its 64KB token ceiling, where it stops with ErrTooLong and the reader silently
	//   sees a TRUNCATED session: "wrote nothing" for a session that wrote plenty.
	//
	//   json.Decoder over the whole stream — a syntax error leaves the decoder unable to
	//   resync, so one malformed record discards EVERY record after it. Measured: a
	//   three-line fixture with a broken middle line returned only the first path.
	//
	// bufio.Reader.ReadString has no size limit and JSONL is line-framed, so a bad line
	// costs exactly that line.
	r := bufio.NewReader(f)
	for {
		raw, err := r.ReadString('\n')
		if strings.TrimSpace(raw) != "" {
			lines++
			uses, ok := toolUses(raw)
			if !ok {
				unparsed++
			}
			for _, b := range uses {
				if !writeTools[b.Name] || b.Input.FilePath == "" {
					continue
				}
				rel, ok := relativeTo(absProject, b.Input.FilePath)
				if !ok || seen[rel] {
					continue
				}
				seen[rel] = true
				out = append(out, rel)
			}
		}
		if err != nil {
			break // io.EOF, or a read error: whatever was collected still stands
		}
	}
	sort.Strings(out)

	switch {
	case lines == 0:
		return out, "transcript held no records"
	// Every line failing to parse is a FORMAT change, not a corrupt session. One bad
	// line is noise; all of them means this reader no longer understands the file.
	case unparsed == lines:
		return out, fmt.Sprintf("none of the transcript's %d record(s) parsed — the harness format has moved", lines)
	}
	return out, ""
}

// toolUses returns the tool_use blocks of one transcript line, or nothing at all when the
// line does not parse. Isolation is the point: this is where a malformed record stops.
func toolUses(raw string) ([]block, bool) {
	if strings.TrimSpace(raw) == "" {
		return nil, true
	}
	var e entry
	if json.Unmarshal([]byte(raw), &e) != nil {
		return nil, false
	}
	var out []block
	for _, b := range blocks(e.Message.Content) {
		if b.Type == "tool_use" {
			out = append(out, b)
		}
	}
	return out, true
}

// blocks decodes a content field that may be an array of blocks or a bare string.
func blocks(raw json.RawMessage) []block {
	if len(raw) == 0 {
		return nil
	}
	var bs []block
	if err := json.Unmarshal(raw, &bs); err != nil {
		return nil // a string content field, or anything else: no tool calls in it
	}
	return bs
}

// relativeTo makes an absolute transcript path project-relative, and reports whether it
// belongs to the project at all.
//
// Changes under .claude/ are excluded: that directory is where these hooks keep their own
// state, and counting the seal's own snapshots as "the session's work" would make every
// session look like it touched something. sc-filechanged-rearm excludes it for the same
// reason, one event earlier.
func relativeTo(absProject, file string) (string, bool) {
	if file == "" {
		return "", false
	}
	abs := file
	if !filepath.IsAbs(abs) {
		abs = filepath.Join(absProject, abs)
	}
	rel, err := filepath.Rel(absProject, abs)
	if err != nil {
		return "", false
	}
	rel = filepath.ToSlash(rel)
	if rel == ".." || strings.HasPrefix(rel, "../") || filepath.IsAbs(rel) {
		return "", false
	}
	if rel == ".claude" || strings.HasPrefix(rel, ".claude/") {
		return "", false
	}
	return rel, true
}
