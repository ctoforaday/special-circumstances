//go:build !unix

package main

import "testing"

func holdPresenceLock(t *testing.T, _, _ string) func() {
	t.Skip("presence lock negative control requires Unix flock")
	return func() {}
}
