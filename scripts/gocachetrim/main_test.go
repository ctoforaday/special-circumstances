package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// cache lays out a fake Go build cache: name -> (size, age before now).
func cache(t *testing.T, readme bool, files map[string]struct {
	size int
	age  time.Duration
}) string {
	t.Helper()
	dir := t.TempDir()
	if readme {
		if err := os.WriteFile(filepath.Join(dir, "README"), []byte(readmeLine+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	now := time.Now()
	for name, f := range files {
		p := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, make([]byte, f.size), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.Chtimes(p, now.Add(-f.age), now.Add(-f.age)); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func exists(dir, name string) bool {
	_, err := os.Stat(filepath.Join(dir, filepath.FromSlash(name)))
	return err == nil
}

// The two passes, in order: what the run did not use goes whatever the cap, then the oldest of
// what it used until the rest fits. Files outside the shards are never touched.
func TestTrimDropsTheUnusedThenTheOldestOverTheCap(t *testing.T) {
	const mb = 1 << 20
	dir := cache(t, true, map[string]struct {
		size int
		age  time.Duration
	}{
		"00/stale-a":     {mb, 48 * time.Hour},
		"01/stale-d":     {mb, 48 * time.Hour},
		"02/used-old-d":  {mb, 30 * time.Minute},
		"03/used-mid-d":  {mb, 20 * time.Minute},
		"04/used-new-a":  {mb, time.Minute},
		"trim.txt":       {10, 72 * time.Hour},
		"fuzz/x/corpus":  {mb, 72 * time.Hour},
		"0g/not-a-shard": {mb, 72 * time.Hour},
	})
	if err := run(dir, time.Now().Add(-time.Hour), 2); err != nil {
		t.Fatal(err)
	}
	for name, want := range map[string]bool{
		"00/stale-a": false, "01/stale-d": false, "02/used-old-d": false,
		"03/used-mid-d": true, "04/used-new-a": true,
		"trim.txt": true, "fuzz/x/corpus": true, "0g/not-a-shard": true, "README": true,
	} {
		if got := exists(dir, name); got != want {
			t.Errorf("%s kept=%v, want %v", name, got, want)
		}
	}
}

// A directory that is not a Go build cache loses nothing.
func TestTrimRefusesADirectoryThatIsNotAGoCache(t *testing.T) {
	dir := cache(t, false, map[string]struct {
		size int
		age  time.Duration
	}{"00/old-a": {1, 48 * time.Hour}})
	if err := run(dir, time.Now(), 1); err == nil {
		t.Fatal("trimmed a directory with no Go cache README")
	}
	if !exists(dir, "00/old-a") {
		t.Error("deleted a file in a directory it refused")
	}
}

// Under the cap after the unused pass, nothing used is removed.
func TestDecideKeepsEverythingUsedThatFits(t *testing.T) {
	now := time.Now()
	p := decide([]entry{{"a", 5, now}, {"b", 5, now.Add(-time.Second)}}, now.Add(-time.Minute), 10)
	if len(p.unused) != 0 || len(p.over) != 0 || p.kept != 2 || p.keptBytes != 10 {
		t.Errorf("decide = %+v, want both kept", p)
	}
}
