package seatprobe

import (
	"net/http"
	"strings"
	"testing"
)

// THE GUARD IS INSTALLED IN THIS BINARY, which is a different question from whether it works.
//
// internal/nonet proves the function refuses; this proves it was CALLED here. Remove the
// nonet.OnlyLoopback() line from testbuild.Main and nonet's own tests still pass while every
// package it was protecting quietly reaches the internet again — the wiring is the part that
// goes missing, not the mechanism.
//
// THIS PACKAGE SPECIFICALLY, because this is where it happened: the 2026-09-06 nightly failed
// TestEveryProbeBoardStillBuilds/sources on an example.com timeout and reported it as a fixture
// speaking a retired model. #807 made the board serve its own source; this makes the next one
// unable to stop doing so silently.
func TestThisBinaryCannotReachThePublicInternet(t *testing.T) {
	_, err := http.Get("https://example.com/")
	if err == nil {
		t.Fatal("this test binary reached example.com — the loopback guard is not installed here, " +
			"so a probe board can go back to depending on a host nobody in this repo controls")
	}
	if !strings.Contains(err.Error(), "nonet:") {
		t.Errorf("the fetch failed for some other reason, so this proves nothing about the guard: %v", err)
	}
}
