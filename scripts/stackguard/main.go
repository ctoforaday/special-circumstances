// Command stackguard polices ONE thing: a stacked pull request must not merge into a base
// branch that is already contained in main.
//
// Dev tooling for this repository only. Nothing here ships to an installing project.
//
// # WHY IT EXISTS
//
// On 2026-08-28, #633 — the fourth and final step of the record.Run migration, 131 files —
// was stacked on #631's branch `record-run-read-leaves`. The timeline is the whole defect:
//
//	17:00  #631 merges `record-run-read-leaves` into main. The branch is NOT deleted.
//	19:47  #633 merges into `record-run-read-leaves`, which nobody reads any more.
//
// GitHub marked #633 MERGED and main never received a line of it. A stacked pull request
// whose base branch is merged but not DELETED is not retargeted, so the second merge lands
// in a branch that has already had its moment. `internal/record` sat on main still taking 53
// `runDir string` for a day, with the pull request that fixed them showing green and merged.
//
// NOTHING FAILED. That is the part this guard answers. Every check on #633 passed, because
// every check asked whether the CONTENT was good and none asked where it was going.
//
// # THE TWO MODES, AND WHY ONE IS NOT ENOUGH
//
// -base is the pull-request-time check: cheap, pure git, no token. It refuses a stacked pull
// request opened against a base that is ALREADY in main.
//
// It would NOT have caught #633, and saying so is the point. At the moment #633's checks ran,
// #631 had not merged yet — the base was live and the answer was honestly "fine". The base
// merged three hours later, and a green check from before is not a claim about after. A gate
// that is only ever evaluated at open time cannot see a hazard that arrives at merge time.
//
// -sweep closes that window from the other side: it runs on every push to main and asks the
// inverse question — now that main has moved, is any OPEN stacked pull request pointing at a
// base this push just absorbed? On the #633 timeline that fires at 17:00, when #631 lands,
// which is ~3 hours before the mis-merge.
//
// # WHAT IT CANNOT SEE, STATED RATHER THAN IMPLIED
//
// Containment is measured with merge-base --is-ancestor, so it detects a base merged by MERGE
// COMMIT or by REBASE. A base branch that was SQUASH-merged leaves no ancestry, and this
// reports it as live. That limit is real and is not worth a heuristic: squash-merging a branch
// that still has pull requests stacked on it is a different act with different tells, and a
// content-similarity guess here would trade a stated blind spot for an unstated wrong answer.
//
// The durable fix is not this tool. It is `delete_branch_on_merge` on the repository, because
// GitHub retargets a stacked pull request when its base branch is DELETED. This guard is what
// you build for the window where that setting is off or the deletion is skipped, and it says
// so out loud rather than standing in for it.
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/ctoforaday/special-circumstances/scripts/internal/gitx"
)

// Exit codes. 0 clean, 1 a finding, 2 the tool could not run.
//
// THERE IS NO "COULD NOT MEASURE" PASS. Every path that fails to answer exits non-zero, and
// that is deliberate: the defect class this tool exists for is a green check that never looked,
// so a stackguard which cannot reach git or the API must not resemble a stackguard that looked
// and found nothing.
const (
	exitClean = 0
	exitFound = 1
	exitError = 2
)

// mainRef is the branch a stacked base must not already be inside of.
const mainRef = "origin/main"

func main() {
	base := flag.String("base", "", "the pull request's base ref (bare branch name, e.g. record-run-read-leaves)")
	sweep := flag.Bool("sweep", false, "check every OPEN pull request whose base is not main (needs GH_TOKEN)")
	repo := flag.String("repo", "", "owner/name, for -sweep; defaults to $GITHUB_REPOSITORY")
	flag.Parse()

	if *base == "" && !*sweep {
		fmt.Fprintln(os.Stderr, "stackguard: pass -base <ref> or -sweep")
		os.Exit(exitError)
	}

	root, err := gitx.Root()
	if err != nil {
		fmt.Fprintf(os.Stderr, "stackguard: %v\n", err)
		os.Exit(exitError)
	}
	contains := gitContains(root)

	if *sweep {
		name := *repo
		if name == "" {
			name = os.Getenv("GITHUB_REPOSITORY")
		}
		if name == "" {
			fmt.Fprintln(os.Stderr, "stackguard: -sweep needs -repo or $GITHUB_REPOSITORY")
			os.Exit(exitError)
		}
		prs, err := openStackedPRs(name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "stackguard: %v\n", err)
			os.Exit(exitError)
		}
		os.Exit(reportSweep(os.Stdout, prs, contains))
	}

	os.Exit(reportOne(os.Stdout, *base, contains))
}

// pullRequest is the little of a pull request this tool reads.
type pullRequest struct {
	Number int
	Base   string
	Title  string
}

// containsFn answers whether ref is already inside main. The error is NOT a bool — a ref that
// cannot be resolved is a question this tool failed to ask, not a "no".
type containsFn func(ref string) (bool, error)

// gitContains asks git whether a branch is an ancestor of main.
//
// It resolves the ref FIRST and treats an unresolvable one as an error. `--is-ancestor` exits
// 1 both for "not an ancestor" and for a ref git cannot parse, so without the explicit
// resolution a typo'd or unfetched base reads exactly like a healthy independent branch — the
// plausible zero this repository names in [[facts-are-fields]].
func gitContains(root string) containsFn {
	return func(ref string) (bool, error) {
		full := "origin/" + bareRef(ref)
		if _, err := gitx.Run(root, "rev-parse", "--verify", "--quiet", full+"^{commit}"); err != nil {
			return false, fmt.Errorf("cannot resolve %s — is the branch fetched? (fetch-depth: 0 and a full refspec): %w", full, err)
		}
		err := exec.Command("git", "-C", root, "merge-base", "--is-ancestor", full, mainRef).Run()
		if err == nil {
			return true, nil
		}
		var ee *exec.ExitError
		if errors.As(err, &ee) && ee.ExitCode() == 1 {
			return false, nil // resolved, and genuinely not an ancestor
		}
		return false, fmt.Errorf("merge-base --is-ancestor %s %s: %w", full, mainRef, err)
	}
}

// bareRef strips the remote prefix, so `main`, `origin/main` and `refs/heads/main` are one
// answer rather than three.
//
// NOT COSMETIC. Callers disagree about the form on purpose: `github.base_ref` is a bare branch
// name, while every sibling gate in scripts/ is invoked as `-base "origin/$BASE_REF"` and
// `scripts/check` defaults the flag to `origin/main`. Comparing the raw string to "main" makes
// the local gate set refuse itself — origin/main is not equal to main, so the tool would ask
// whether main is contained in main, get yes, and report the repository's own default branch as
// a stranded stack on every developer run.
func bareRef(ref string) string {
	ref = strings.TrimPrefix(ref, "refs/heads/")
	ref = strings.TrimPrefix(ref, "origin/")
	return ref
}

// reportOne is the pull-request-time check on a single base ref.
func reportOne(out *os.File, base string, contains containsFn) int {
	if bareRef(base) == "main" || base == "" {
		fmt.Fprintln(out, "stackguard: base is main — nothing to check")
		return exitClean
	}
	merged, err := contains(base)
	if err != nil {
		fmt.Fprintf(os.Stderr, "stackguard: %v\n", err)
		return exitError
	}
	if !merged {
		fmt.Fprintf(out, "stackguard: base %q is not yet in main — this stack is live\n", base)
		return exitClean
	}
	fmt.Fprintf(out, "%s\n", stranded(base))
	return exitFound
}

// reportSweep is the push-to-main check across every open stacked pull request.
func reportSweep(out *os.File, prs []pullRequest, contains containsFn) int {
	if len(prs) == 0 {
		fmt.Fprintln(out, "stackguard: no open pull request targets a base other than main")
		return exitClean
	}
	code := exitClean
	for _, pr := range prs {
		merged, err := contains(pr.Base)
		if err != nil {
			// LOUD, and it does not become a pass for the others: one unresolvable base means
			// this sweep did not answer for that pull request, which is the state the tool
			// refuses to render as clean.
			fmt.Fprintf(os.Stderr, "stackguard: #%d (base %s): %v\n", pr.Number, pr.Base, err)
			code = exitError
			continue
		}
		if !merged {
			continue
		}
		fmt.Fprintf(out, "#%d %s\n%s\n", pr.Number, pr.Title, stranded(pr.Base))
		if code == exitClean {
			code = exitFound
		}
	}
	if code == exitClean {
		fmt.Fprintf(out, "stackguard: %d open stacked pull request(s), every base still live\n", len(prs))
	}
	return code
}

// stranded is the refusal text. It names the act to take, because "your base is merged" is a
// fact and "retarget before merging" is the thing a reader has to do with it.
func stranded(base string) string {
	return fmt.Sprintf(
		"::error::the base branch %q is ALREADY IN MAIN. Merging into it now lands the work in a "+
			"branch nobody reads: GitHub will mark the pull request MERGED and main will not "+
			"receive a line of it (#633, 2026-08-28). RETARGET this pull request to main — "+
			"`gh pr edit <n> --base main` — and resolve the merge there. Verify with: "+
			"git merge-base --is-ancestor <merge-commit> origin/main", base)
}

// openStackedPRs lists open pull requests whose base is not main, via gh.
//
// gh rather than a hand-rolled HTTP client: it is already the authenticated boundary every
// other operator step in this repository uses, and it resolves the token the same way, so
// there is one story about credentials instead of two.
func openStackedPRs(repo string) ([]pullRequest, error) {
	cmd := exec.Command("gh", "api", "--paginate",
		fmt.Sprintf("repos/%s/pulls?state=open&per_page=100", repo),
		"--jq", `.[] | select(.base.ref != "main") | {Number: .number, Base: .base.ref, Title: .title}`)
	out, err := cmd.Output()
	if err != nil {
		var stderr string
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			stderr = strings.TrimSpace(string(ee.Stderr))
		}
		return nil, fmt.Errorf("listing open pull requests for %s: %w: %s", repo, err, stderr)
	}
	return decodePRs(string(out))
}

// decodePRs reads gh's --jq output, which is one JSON object per line rather than an array.
func decodePRs(s string) ([]pullRequest, error) {
	var prs []pullRequest
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var pr pullRequest
		if err := json.Unmarshal([]byte(line), &pr); err != nil {
			return nil, fmt.Errorf("decoding %q: %w", line, err)
		}
		prs = append(prs, pr)
	}
	return prs, nil
}
