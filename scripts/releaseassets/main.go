package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ctoforaday/special-circumstances/scripts/internal/gitx"
)

func main() {
	var (
		flagScope  = flag.Bool("scope", false, "compute native jobs scope (outputs matrix, cache, full)")
		flagMatrix = flag.Bool("matrix", false, "print native-run targets matrix JSON to stdout")
		flagCheck  = flag.Bool("check", false, "verify assets in dist directory")
		flagDist   = flag.String("dist", "", "dist directory (defaults to $DIST or dist)")
		flagRef    = flag.String("ref", "", "git ref override (defaults to $GITHUB_REF)")
		flagEvent  = flag.String("event", "", "github event override (defaults to $GITHUB_EVENT_NAME)")
		flagWF     = flag.String("workflow", ".github/workflows/hooks.yml", "path to workflow file")
		flagNative = flag.String("native-dir", "plugins/frank-exchange-of-views/native", "path to native dir")
	)
	flag.Parse()

	if root, err := gitx.Root(); err == nil {
		if *flagWF == ".github/workflows/hooks.yml" && !filepath.IsAbs(*flagWF) {
			if _, statErr := os.Stat(*flagWF); os.IsNotExist(statErr) {
				*flagWF = filepath.Join(root, *flagWF)
			}
		}
		if *flagNative == "plugins/frank-exchange-of-views/native" && !filepath.IsAbs(*flagNative) {
			if _, statErr := os.Stat(*flagNative); os.IsNotExist(statErr) {
				*flagNative = filepath.Join(root, *flagNative)
			}
		}
	}

	pairs, err := ReadReleasePairs(*flagWF)
	if err != nil {
		fmt.Fprintf(os.Stderr, "releaseassets: %v\n", err)
		os.Exit(1)
	}
	targets, err := ParseTargets(pairs)
	if err != nil {
		fmt.Fprintf(os.Stderr, "releaseassets: %v\n", err)
		os.Exit(1)
	}
	helpers, err := FindHelpers(*flagNative)
	if err != nil {
		fmt.Fprintf(os.Stderr, "releaseassets: %v\n", err)
		os.Exit(1)
	}

	if *flagMatrix {
		var runLegs []RunLeg
		for _, t := range targets {
			runLegs = append(runLegs, RunLeg{
				Target: t.Name,
				Runner: t.Runner,
				OS:     t.OS,
				Arch:   t.Arch,
				Ext:    t.Ext,
			})
		}
		data, err := json.Marshal(Matrix[RunLeg]{Include: runLegs})
		if err != nil {
			fmt.Fprintf(os.Stderr, "releaseassets: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(string(data))
		return
	}

	if *flagScope {
		event := *flagEvent
		if event == "" {
			event = os.Getenv("GITHUB_EVENT_NAME")
		}
		ref := *flagRef
		if ref == "" {
			ref = os.Getenv("GITHUB_REF")
		}

		var changed []string
		if event == "pull_request" {
			// Read changed files via git diff
			baseRef := os.Getenv("GITHUB_BASE_REF")
			var cmd *exec.Cmd
			if baseRef != "" {
				cmd = exec.Command("git", "diff", "--name-only", "origin/"+baseRef+"...HEAD")
			} else {
				cmd = exec.Command("git", "diff", "--name-only", "HEAD~1", "HEAD")
			}
			out, err := cmd.Output()
			if err == nil {
				for _, line := range strings.Split(string(out), "\n") {
					line = strings.TrimSpace(line)
					if line != "" {
						changed = append(changed, line)
					}
				}
			}
		}

		consumers, err := DeriveConsumerTargets(*flagWF)
		if err != nil {
			fmt.Fprintf(os.Stderr, "releaseassets: %v\n", err)
			os.Exit(1)
		}

		scope, err := DecideScope(event, ref, changed, targets, consumers, helpers)
		if err != nil {
			fmt.Fprintf(os.Stderr, "releaseassets: %v\n", err)
			os.Exit(1)
		}

		matrixJSON, _ := json.Marshal(scope.Matrix)
		runMatrixJSON, _ := json.Marshal(scope.RunMatrix)

		// Print decision
		fmt.Printf("cache: %s\n", scope.Cache)
		fmt.Printf("full: %v\n", scope.Full)
		fmt.Printf("matrix: %s\n", string(matrixJSON))
		fmt.Printf("run-matrix: %s\n", string(runMatrixJSON))

		// Write to $GITHUB_OUTPUT if present
		if ghOut := os.Getenv("GITHUB_OUTPUT"); ghOut != "" {
			f, err := os.OpenFile(ghOut, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0o644)
			if err != nil {
				fmt.Fprintf(os.Stderr, "releaseassets: writing GITHUB_OUTPUT: %v\n", err)
				os.Exit(1)
			}
			defer f.Close()
			fmt.Fprintf(f, "cache=%s\n", scope.Cache)
			fmt.Fprintf(f, "full=%v\n", scope.Full)
			fmt.Fprintf(f, "matrix=%s\n", string(matrixJSON))
			fmt.Fprintf(f, "run-matrix=%s\n", string(runMatrixJSON))
		}
		return
	}

	if *flagCheck {
		distDir := *flagDist
		if distDir == "" {
			distDir = os.Getenv("DIST")
		}
		if distDir == "" {
			distDir = "dist"
		}
		distDir, _ = filepath.Abs(distDir)
		ref := *flagRef
		if ref == "" {
			ref = os.Getenv("GITHUB_REF")
		}

		if err := CheckDist(distDir, ref, targets, helpers); err != nil {
			fmt.Fprintf(os.Stderr, "releaseassets check: %v\n", err)
			os.Exit(1)
		}
		return
	}

	fmt.Fprintln(os.Stderr, "releaseassets: specify -scope, -matrix, or -check")
	os.Exit(1)
}
