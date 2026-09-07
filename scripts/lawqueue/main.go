// Command lawqueue gives the law pipeline a reader, and refuses a holding promoted without one.
//
// Dev tooling for this repository only. Nothing here ships to an installing project.
//
// # The pipe was write-only
//
// `capture` harvests every ruling a run makes into `law/proposed/` as PERSUASIVE, and
// `bench declare` writes there too. NOTHING READS IT. Measured 2026-09-07: 76 holdings across 25
// run files, and `law/precedents.md` empty since it was founded on 2026-07-18 — seven weeks of
// harvest, nothing promoted, and no surface anywhere that would tell a human a queue existed.
//
// That is the friction complaint one level up: the bench's own reasoning, filed as prose into a
// channel nothing forces anyone to read. Promotion is deliberately a human act and stays one —
// this does not promote anything. It makes the queue VISIBLE, which is the part that was missing.
//
// # The one thing it fails on
//
// A holding in `law/precedents.md` carrying a `<reviewer: …>` stub. The harvest writes those
// placeholders on purpose and never invents their contents — `facts`, `holding` and `scope-limits`
// are what a reviewer supplies, and supplying them IS the review. So a stub that reaches the
// affirmed corpus is a holding nobody read, presented as law. That is an error; a long queue is
// not.
//
// Usage:
//
//	go run ./lawqueue            report the queue, fail on an unreviewed promotion
//	go run ./lawqueue -selftest  prove the stub matcher can still fire
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/ctoforaday/special-circumstances/scripts/internal/gitx"
)

// reviewerStub is the placeholder the harvest writes and a reviewer replaces. It is matched
// loosely on purpose: the wording after the colon has already changed once, and a matcher keyed on
// the full sentence would stop firing without anyone noticing — which is the failure this whole
// tool is about.
var reviewerStub = regexp.MustCompile(`<reviewer:`)

// holding counts a `## <id> [PERSUASIVE]` heading. AFFIRMED holdings in precedents.md are counted
// by their own heading, so the two corpora are never mixed by one pattern.
var (
	persuasive = regexp.MustCompile(`(?m)^##\s+\S+\s+\[PERSUASIVE\]`)
	affirmed   = regexp.MustCompile(`(?m)^##\s+`)
)

func main() {
	selftest := flag.Bool("selftest", false, "prove the stub matcher can still fire, then exit")
	flag.Parse()
	if *selftest {
		if !reviewerStub.MatchString("holding: <reviewer: state the rule this ruling applied>") {
			fmt.Fprintln(os.Stderr, "lawqueue: the stub matcher no longer fires on the placeholder the harvest writes")
			os.Exit(1)
		}
		if reviewerStub.MatchString("holding: a hand-authored apparatus count may be read, not docked") {
			fmt.Fprintln(os.Stderr, "lawqueue: the stub matcher fires on a REVIEWED holding — every promotion would be refused")
			os.Exit(1)
		}
		fmt.Println("lawqueue: the stub matcher still fires on a placeholder and not on a reviewed holding")
		return
	}

	root, err := gitx.Root()
	if err != nil {
		fmt.Fprintln(os.Stderr, "lawqueue:", err)
		os.Exit(1)
	}
	lawDir := filepath.Join(root, "law")

	// THE AFFIRMED CORPUS FIRST, because it is the only thing here that can FAIL.
	precedents := filepath.Join(lawDir, "precedents.md")
	body, err := os.ReadFile(precedents)
	if err != nil {
		fmt.Fprintf(os.Stderr, "lawqueue: %s is unreadable (%v) — the affirmed corpus is the one file this must check\n", precedents, err)
		os.Exit(1)
	}
	if n := len(reviewerStub.FindAllIndex(body, -1)); n > 0 {
		fmt.Fprintf(os.Stderr, "lawqueue: law/precedents.md carries %d unfilled `<reviewer: …>` placeholder(s).\n\n"+
			"The harvest writes those and never invents their contents: `facts`, `holding` and `scope-limits`\n"+
			"are what a reviewer supplies, and supplying them IS the review. A stub here is a holding nobody\n"+
			"read, sitting in the corpus that other rulings are decided against.\n\n"+
			"Fill them from the cited record, or move the holding back to law/proposed/.\n", n)
		os.Exit(1)
	}

	// THE QUEUE, which is reported and never failed on. A long queue is a decision nobody has
	// made yet; a gate here would only teach people to empty it carelessly.
	entries, err := os.ReadDir(filepath.Join(lawDir, "proposed"))
	if err != nil {
		fmt.Println("lawqueue: no law/proposed/ — nothing has been harvested yet")
		return
	}
	type file struct {
		name     string
		holdings int
	}
	var files []file
	total := 0
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		b, rerr := os.ReadFile(filepath.Join(lawDir, "proposed", e.Name()))
		if rerr != nil {
			continue
		}
		n := len(persuasive.FindAllIndex(b, -1))
		if n == 0 {
			continue
		}
		files = append(files, file{e.Name(), n})
		total += n
	}
	sort.Slice(files, func(i, j int) bool { return files[i].holdings > files[j].holdings })

	promoted := len(affirmed.FindAllIndex(body, -1))
	fmt.Printf("law: %d holding(s) awaiting review across %d run file(s); %d affirmed in precedents.md\n",
		total, len(files), promoted)
	if total == 0 {
		return
	}
	for _, f := range files {
		fmt.Printf("  %3d  law/proposed/%s\n", f.holdings, f.name)
	}
	if promoted == 0 {
		fmt.Printf("\nNOTHING HAS EVER BEEN PROMOTED. The harvest has run %d times and the affirmed corpus is\n"+
			"empty, so every ruling this suite has made is PERSUASIVE and nothing is settled law. That is a\n"+
			"review nobody has done, not a property of the design — see law/README.md for what promotion is.\n",
			len(files))
	}
}
