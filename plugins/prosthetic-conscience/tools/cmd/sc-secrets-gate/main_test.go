package main

import "testing"

func TestDenyShape(t *testing.T) {
	d := deny("why")
	if d.HookSpecificOutput.PermissionDecision != "deny" ||
		d.HookSpecificOutput.HookEventName != "PreToolUse" ||
		d.HookSpecificOutput.PermissionDecisionReason != "why" {
		t.Fatalf("deny() malformed: %+v", d)
	}
}
