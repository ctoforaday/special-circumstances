package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSemverLess(t *testing.T) {
	cases := []struct {
		a, b string
		want bool
	}{
		{"0.7.0", "0.8.0", true},
		{"0.8.0", "0.7.0", false},
		{"0.8.0", "0.8.0", false},
		{"0.9.0", "0.10.0", true}, // numeric, not lexicographic
		{"0.8", "0.8.1", true},
	}
	for _, c := range cases {
		if got := semverLess(c.a, c.b); got != c.want {
			t.Fatalf("semverLess(%q,%q) = %v; want %v", c.a, c.b, got, c.want)
		}
	}
}

// cacheFixture builds <tmp>/<plugin>/<version>[/bin/tool.exe][/requirements.json].
func cacheFixture(t *testing.T) string {
	t.Helper()
	market := t.TempDir()
	mk := func(parts ...string) string {
		p := filepath.Join(append([]string{market}, parts...)...)
		if err := os.MkdirAll(p, 0o755); err != nil {
			t.Fatal(err)
		}
		return p
	}
	// Own plugin: running from 0.7.0 while 0.8.0 exists with an EMPTY bin.
	mk("prosthetic-conscience", "0.7.0", "bin")
	os.WriteFile(filepath.Join(mk("prosthetic-conscience", "0.7.0", "bin"), "sc-doctor.exe"), []byte("x"), 0o755)
	mk("prosthetic-conscience", "0.8.0", "bin")
	// Sibling plugin with its own requirements.json at its newest version.
	sib := mk("frank-exchange-of-views", "0.6.0")
	os.WriteFile(filepath.Join(sib, "requirements.json"),
		[]byte(`{"tools":[{"name":"node","purpose":"scripts","tier":"recommended","check_cmd":"node --version","install":{"windows":"winget","darwin":"brew","linux":"apt"}}]}`), 0o644)
	mk("frank-exchange-of-views", "0.5.0") // older version must be ignored
	return market
}

func TestDanceWarnings(t *testing.T) {
	market := cacheFixture(t)
	root := filepath.Join(market, "prosthetic-conscience", "0.7.0")
	ws := danceWarnings(root)
	if len(ws) != 2 {
		t.Fatalf("want 2 warnings (stale version + empty bin), got %d: %v", len(ws), ws)
	}
	if !strings.Contains(ws[0], "DANCE INCOMPLETE") || !strings.Contains(ws[0], "0.8.0") {
		t.Fatalf("stale-version warning wrong: %q", ws[0])
	}
	if !strings.Contains(ws[1], "EMPTY-BIN WINDOW") {
		t.Fatalf("empty-bin warning wrong: %q", ws[1])
	}
	// Running from the newest version with a populated bin: no warnings.
	os.WriteFile(filepath.Join(market, "prosthetic-conscience", "0.8.0", "bin", "sc-doctor.exe"), []byte("x"), 0o755)
	newest := filepath.Join(market, "prosthetic-conscience", "0.8.0")
	if ws := danceWarnings(newest); len(ws) != 0 {
		t.Fatalf("healthy cache should warn nothing, got %v", ws)
	}
}

func TestSiblingRequirements(t *testing.T) {
	market := cacheFixture(t)
	root := filepath.Join(market, "prosthetic-conscience", "0.7.0")
	prs := siblingRequirements(root)
	if len(prs) != 1 {
		t.Fatalf("want 1 sibling with requirements, got %d", len(prs))
	}
	if prs[0].Plugin != "frank-exchange-of-views" || len(prs[0].Tools) != 1 || prs[0].Tools[0].Name != "node" {
		t.Fatalf("sibling aggregation wrong: %+v", prs[0])
	}
}
