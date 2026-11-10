// Command releaseassets holds the facts the native jobs and the release job share: the published
// target set (the release job's `for pair in` line), each target's runner, each helper's builder,
// which native-build legs an event needs (-scope), the native-run matrix (-matrix), and the check
// a release's dist passes before SHA256SUMS is written (-check).
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
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
		flagReq    = flag.Bool("testpackages", false, "print the FEOV tools packages that call nativehelper.Require, the packages native-run runs")
	)
	flag.Parse()

	if err := run(*flagScope, *flagMatrix, *flagCheck, *flagReq, *flagDist, *flagRef, *flagEvent, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "releaseassets: %v\n", err)
		os.Exit(1)
	}
}

func run(doScope, doMatrix, doCheck, doRequirePackages bool, dist, ref, event string, stdout io.Writer) error {
	root, err := gitx.Root()
	if err != nil {
		return err
	}
	wf := filepath.Join(root, ".github", "workflows", "hooks.yml")
	pairs, err := ReadReleasePairs(wf)
	if err != nil {
		return err
	}
	targets, err := ParseTargets(pairs)
	if err != nil {
		return err
	}
	if ref == "" {
		ref = os.Getenv("GITHUB_REF")
	}

	switch {
	case doMatrix:
		var runLegs []RunLeg
		for _, t := range targets {
			runLegs = append(runLegs, RunLeg{Target: t.Name, Runner: t.Runner, OS: t.OS, Arch: t.Arch, Ext: t.Ext})
		}
		data, err := json.Marshal(Matrix[RunLeg]{Include: runLegs})
		if err != nil {
			return err
		}
		fmt.Fprintln(stdout, string(data))
		return nil
	case doScope:
		if event == "" {
			event = os.Getenv("GITHUB_EVENT_NAME")
		}
		nativeSet, err := DeriveNativeSet(root)
		if err != nil {
			return err
		}
		return Scope(root, event, ref, nativeSet, targets, stdout, os.Getenv("GITHUB_OUTPUT"))
	case doRequirePackages:
		pkgs, err := RequirePackages(root)
		if err != nil {
			return err
		}
		fmt.Fprintln(stdout, strings.Join(pkgs, " "))
		return nil
	case doCheck:
		if dist == "" {
			dist = os.Getenv("DIST")
		}
		if dist == "" {
			return errors.New("-check needs -dist or $DIST")
		}
		dist, err = filepath.Abs(dist)
		if err != nil {
			return err
		}
		plugin, err := PluginFromRef(ref)
		if err != nil {
			return err
		}
		return CheckDist(dist, filepath.Join(root, "plugins", plugin), targets, stdout)
	}
	return errors.New("specify -scope, -matrix, -check or -testpackages")
}

// Scope decides one event's native jobs for the repository at root and prints the decision; with
// ghOutput set it also appends the decision there as step outputs. A pull request's changed files
// are its merge commit's diff against the base tip, HEAD^1; a pull request whose HEAD is not that
// merge commit, or any git failure, is an error, never an empty change list.
func Scope(root, event, ref string, nativeSet NativeSet, targets []Target, stdout io.Writer, ghOutput string) error {
	wf := filepath.Join(root, ".github", "workflows", "hooks.yml")
	helpers, err := FindHelpers(filepath.Join(root, "plugins", "frank-exchange-of-views", "native"))
	if err != nil {
		return err
	}
	var changed []string
	if event == "pull_request" {
		if changed, err = PullRequestChanges(root); err != nil {
			return err
		}
	}
	consumers, err := DeriveConsumerTargets(wf)
	if err != nil {
		return err
	}
	scope, err := DecideScope(event, ref, changed, nativeSet, targets, consumers, helpers)
	if err != nil {
		return err
	}
	matrixJSON, err := json.Marshal(scope.Matrix)
	if err != nil {
		return err
	}
	runMatrixJSON, err := json.Marshal(scope.RunMatrix)
	if err != nil {
		return err
	}
	if event == "pull_request" {
		fmt.Fprintf(stdout, "changed files: %d, native set touched: %v\n", len(changed), nativeSet.Touched(changed))
	}
	fmt.Fprintf(stdout, "cache: %s\nfull: %v\nmatrix: %s\nrun-matrix: %s\n", scope.Cache, scope.Full, matrixJSON, runMatrixJSON)
	if ghOutput == "" {
		return nil
	}
	f, err := os.OpenFile(ghOutput, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0o644)
	if err != nil {
		return fmt.Errorf("writing GITHUB_OUTPUT: %w", err)
	}
	if _, err := fmt.Fprintf(f, "cache=%s\nfull=%v\nmatrix=%s\nrun-matrix=%s\n", scope.Cache, scope.Full, matrixJSON, runMatrixJSON); err != nil {
		f.Close()
		return fmt.Errorf("writing GITHUB_OUTPUT: %w", err)
	}
	return f.Close()
}

// PullRequestChanges is the files a pull request's merge commit changes against its first parent,
// the base tip. The checkout is the merge commit with fetch-depth 2, so both parents' commits are
// present; HEAD without a second parent is not the merge commit, and its diff is not the pull
// request's.
func PullRequestChanges(root string) ([]string, error) {
	if _, err := gitx.Run(root, "rev-parse", "-q", "--verify", "HEAD^2"); err != nil {
		return nil, fmt.Errorf("HEAD is not the pull request's merge commit, so its changes cannot be read: %w", err)
	}
	out, err := gitx.Run(root, "diff", "--name-only", "HEAD^1", "HEAD")
	if err != nil {
		return nil, fmt.Errorf("reading the pull request's changes: %w", err)
	}
	return gitx.Lines(out), nil
}
