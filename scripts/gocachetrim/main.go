// Command gocachetrim caps a Go build cache before CI saves it.
//
// Dev tooling for this repository only; nothing here ships.
//
// THE CACHE IS RESTORED FROM THE LAST SAVE AND SAVED UNDER A NEW KEY (.github/actions/go-cache),
// so without a trim it only grows: every entry a past run wrote rides along into every later save,
// whether or not anything still asks for it. Two passes, in this order:
//
//  1. UNUSED THIS RUN. The Go command stamps an entry's mtime when it uses one whose mtime is more
//     than an hour old, and a restored entry carries the mtime its save archived. So an entry
//     whose mtime is still before -since — the moment the restore finished — was neither written
//     nor read by this run, and it goes. A run that starts within an hour of the save it restored
//     can read an entry without stamping it; that entry goes too, and costs one rebuild.
//  2. OVER THE CAP. If what is left is still larger than -max-mib, the oldest entries go first
//     until it fits.
//
// Deleting a cache file is always safe: the Go command checks that an action's output exists, at
// the size the action recorded, before using it, and treats anything else as a miss.
//
// It refuses a directory that does not carry the Go cache's own README, so a wrong -dir deletes
// nothing.
package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// readmeLine is the first line the Go command writes into every build cache's README.
const readmeLine = "This directory holds cached build artifacts from the Go build system."

// entry is one cache file: an action record (-a) or an output (-d).
type entry struct {
	path  string
	size  int64
	mtime time.Time
}

// plan is what a trim removes.
type plan struct {
	unused, over           []entry
	unusedBytes, overBytes int64
	keptBytes              int64
	kept                   int
}

// decide removes every entry older than since, then the oldest of the rest until they fit in max.
func decide(entries []entry, since time.Time, max int64) plan {
	var p plan
	var rest []entry
	for _, e := range entries {
		if e.mtime.Before(since) {
			p.unused = append(p.unused, e)
			p.unusedBytes += e.size
			continue
		}
		rest = append(rest, e)
		p.keptBytes += e.size
	}
	sort.Slice(rest, func(i, j int) bool {
		if !rest[i].mtime.Equal(rest[j].mtime) {
			return rest[i].mtime.Before(rest[j].mtime)
		}
		return rest[i].path < rest[j].path
	})
	i := 0
	for ; i < len(rest) && p.keptBytes > max; i++ {
		p.over = append(p.over, rest[i])
		p.overBytes += rest[i].size
		p.keptBytes -= rest[i].size
	}
	p.kept = len(rest) - i
	return p
}

// isShard reports whether name is one of the cache's two-hex-digit subdirectories.
func isShard(name string) bool {
	if len(name) != 2 {
		return false
	}
	for _, c := range name {
		if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'f') {
			return false
		}
	}
	return true
}

// scan lists the cache's entries, refusing a directory that is not a Go build cache.
func scan(dir string) ([]entry, error) {
	readme, err := os.ReadFile(filepath.Join(dir, "README"))
	if err != nil || !bytes.HasPrefix(readme, []byte(readmeLine)) {
		return nil, fmt.Errorf("%s holds no Go build cache README (%v); refusing to delete anything in it", dir, err)
	}
	shards, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var out []entry
	for _, s := range shards {
		if !s.IsDir() || !isShard(s.Name()) {
			continue
		}
		files, err := os.ReadDir(filepath.Join(dir, s.Name()))
		if err != nil {
			return nil, err
		}
		for _, f := range files {
			n := f.Name()
			if f.IsDir() || len(n) < 3 || (n[len(n)-2:] != "-a" && n[len(n)-2:] != "-d") {
				continue
			}
			info, err := f.Info()
			if errors.Is(err, fs.ErrNotExist) {
				continue
			}
			if err != nil {
				return nil, err
			}
			out = append(out, entry{path: filepath.Join(dir, s.Name(), n), size: info.Size(), mtime: info.ModTime()})
		}
	}
	return out, nil
}

func mib(n int64) float64 { return float64(n) / (1 << 20) }

func run(dir string, since time.Time, maxMiB int64) error {
	entries, err := scan(dir)
	if err != nil {
		return err
	}
	p := decide(entries, since, maxMiB<<20)
	for _, e := range append(append([]entry{}, p.unused...), p.over...) {
		if err := os.Remove(e.path); err != nil && !errors.Is(err, fs.ErrNotExist) {
			return err
		}
	}
	fmt.Printf("gocachetrim: %s: %d entries; removed %d unused since %s (%.0f MiB) and %d over the cap (%.0f MiB); kept %d, %.0f MiB of the %d MiB cap\n",
		dir, len(entries), len(p.unused), since.UTC().Format(time.RFC3339), mib(p.unusedBytes),
		len(p.over), mib(p.overBytes), p.kept, mib(p.keptBytes), maxMiB)
	return nil
}

func main() {
	dir := flag.String("dir", "", "the Go build cache (go env GOCACHE)")
	since := flag.Int64("since", 0, "Unix seconds at which the restore finished; an entry older than this was unused by the run")
	maxMiB := flag.Int64("max-mib", 0, "the cap, in MiB, on what is kept")
	flag.Parse()
	if *dir == "" || *since <= 0 || *maxMiB <= 0 {
		fmt.Fprintln(os.Stderr, "gocachetrim: -dir, -since and -max-mib are all required")
		os.Exit(2)
	}
	if err := run(*dir, time.Unix(*since, 0), *maxMiB); err != nil {
		fmt.Fprintln(os.Stderr, "gocachetrim:", err)
		os.Exit(1)
	}
}
