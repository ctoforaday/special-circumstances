package trajectory

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiscoverClaude(t *testing.T) {
	tmp := t.TempDir()
	projDir := filepath.Join(tmp, "my-proj")
	sessDir := filepath.Join(projDir, "sess-1")
	subagentsDir := filepath.Join(sessDir, "subagents")
	wfSubagentsDir := filepath.Join(sessDir, "subagents", "workflows", "wf1")
	if err := os.MkdirAll(wfSubagentsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// 1. Session root transcript
	_ = os.WriteFile(filepath.Join(tmp, "ignored.jsonl"), []byte("{}"), 0644) // not a project tier
	_ = os.WriteFile(filepath.Join(projDir, "sess-1.jsonl"), []byte("{}"), 0644)
	// 2. Subagent transcript
	_ = os.WriteFile(filepath.Join(subagentsDir, "agent-sub1.jsonl"), []byte("{}"), 0644)
	// 3. Workflow subagent transcript
	_ = os.WriteFile(filepath.Join(wfSubagentsDir, "agent-sub2.jsonl"), []byte("{}"), 0644)

	files, err := DiscoverClaude(tmp)
	if err != nil {
		t.Fatalf("DiscoverClaude: %v", err)
	}
	if len(files) != 3 {
		t.Fatalf("len(files) = %d, want 3 (got %v)", len(files), files)
	}

	foundRoot, foundSub1, foundSub2 := false, false, false
	for _, f := range files {
		if f.SessionID == "sess-1" && f.AgentID == "" {
			foundRoot = true
		}
		if f.SessionID == "sess-1" && f.AgentID == "sub1" {
			foundSub1 = true
		}
		if f.SessionID == "sess-1" && f.AgentID == "sub2" {
			foundSub2 = true
		}
	}
	if !foundRoot || !foundSub1 || !foundSub2 {
		t.Errorf("discovery missing files: root=%v, sub1=%v, sub2=%v", foundRoot, foundSub1, foundSub2)
	}
}

func TestDiscoverAntigravity(t *testing.T) {
	tmp := t.TempDir()

	// Parent session
	parentID := "conv-parent"
	parentLogDir := filepath.Join(tmp, parentID, ".system_generated", "logs")
	parentSubDir := filepath.Join(tmp, parentID, ".system_generated", "subagents")
	if err := os.MkdirAll(parentLogDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(parentSubDir, 0755); err != nil {
		t.Fatal(err)
	}
	_ = os.WriteFile(filepath.Join(parentLogDir, "transcript.jsonl"), []byte("{}"), 0644)

	// Subagent session
	subID := "conv-child"
	childLogDir := filepath.Join(tmp, subID, ".system_generated", "logs")
	if err := os.MkdirAll(childLogDir, 0755); err != nil {
		t.Fatal(err)
	}
	_ = os.WriteFile(filepath.Join(childLogDir, "transcript.jsonl"), []byte("{}"), 0644)

	// Subagent descriptor in parent
	desc := `{"conversationId":"conv-child","subagentDescriptor":{"typeName":"researcher"}}`
	_ = os.WriteFile(filepath.Join(parentSubDir, subID+".json"), []byte(desc), 0644)

	files, err := DiscoverAntigravity(tmp)
	if err != nil {
		t.Fatalf("DiscoverAntigravity: %v", err)
	}
	if len(files) != 2 {
		t.Fatalf("len(files) = %d, want 2 (got %v)", len(files), files)
	}

	foundParent, foundChild := false, false
	for _, f := range files {
		if f.SessionID == parentID && f.AgentID == "" {
			foundParent = true
		}
		if f.SessionID == parentID && f.AgentID == subID {
			foundChild = true
		}
	}
	if !foundParent || !foundChild {
		t.Errorf("discovery missing files: parent=%v, child=%v", foundParent, foundChild)
	}
}

func TestDiscoverAllAndSession(t *testing.T) {
	_ = DefaultClaudeProjectsDir()
	_ = DefaultAntigravityBrainDirs()

	claudeDir := t.TempDir()
	projDir := filepath.Join(claudeDir, "my-proj")
	if err := os.MkdirAll(projDir, 0755); err != nil {
		t.Fatal(err)
	}
	_ = os.WriteFile(filepath.Join(projDir, "sess-1.jsonl"), []byte("{}"), 0644)

	agyDir := t.TempDir()
	parentID := "conv-parent"
	parentLogDir := filepath.Join(agyDir, parentID, ".system_generated", "logs")
	if err := os.MkdirAll(parentLogDir, 0755); err != nil {
		t.Fatal(err)
	}
	_ = os.WriteFile(filepath.Join(parentLogDir, "transcript.jsonl"), []byte("{}"), 0644)

	all, err := DiscoverAll(claudeDir, []string{agyDir})
	if err != nil {
		t.Fatalf("DiscoverAll: %v", err)
	}
	if len(all) != 2 {
		t.Fatalf("DiscoverAll got %d files, want 2", len(all))
	}

	sess1, err := DiscoverSession(claudeDir, []string{agyDir}, "sess-1")
	if err != nil || len(sess1) != 1 {
		t.Fatalf("DiscoverSession sess-1 got %v (err %v), want 1", sess1, err)
	}

	sessParent, err := DiscoverSession(claudeDir, []string{agyDir}, parentID)
	if err != nil || len(sessParent) != 1 {
		t.Fatalf("DiscoverSession parent got %v (err %v), want 1", sessParent, err)
	}
}
