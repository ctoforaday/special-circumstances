package fuzz

import (
	"slices"
	"testing"
)

// EVERY KIND OF BRIDGE IS NAMED BY ANY THREE READINGS OF ONE PROCESS, and by the first reading of each
// of three parts: the drive the sweep's enum gate reads `verify --bridge` from (#1943).
func TestTheBridgeKindCyclesOverTheProcess(t *testing.T) {
	defer bridgeDrive.Store(0)
	for start := range int64(2 * len(bridgeKinds)) {
		bridgeDrive.Store(start)
		var named []string
		for range bridgeKinds {
			named = append(named, nextBridgeKind())
		}
		for _, k := range bridgeKinds {
			if !slices.Contains(named, k) {
				t.Errorf("three readings from %d named %v, never %s", start, named, k)
			}
		}
	}
	var first []string
	for k := range len(bridgeKinds) {
		resetBridgeDrive(sweepPart{K: k, Of: 4})
		first = append(first, nextBridgeKind())
	}
	for _, k := range bridgeKinds {
		if !slices.Contains(first, k) {
			t.Errorf("the first readings of parts 0 to %d named %v, never %s", len(bridgeKinds)-1, first, k)
		}
	}
}
