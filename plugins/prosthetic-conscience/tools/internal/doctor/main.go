// sc-doctor is the prosthetic-conscience environment preflight (Go binary).
//
// Contract (Design by Contract):
//
//	It MUST produce a deterministic table + verdict (READY / DEGRADED / BLOCKED).
//	Plain run is read-only. `-fix` rebuilds missing hook binaries (go build) when Go
//	is present, else prints the release-asset fetch instructions — it MUST NOT
//	install external tools (that stays consent-gated at the agent layer).
package doctor

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/buildid"
	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/toolchain"
)

type requirements struct {
	Tools []toolchain.Tool `json:"tools"`
}

type binStatus struct {
	Name  string
	Built bool
	// Plugin/Root/Version identify WHICH plugin owns this binary. The doctor used
	// to know only its own: prosthetic-conscience shipped the only Go binaries, so
	// "the binaries" and "our binaries" were the same set. frank-exchange-of-views
	// now ships feov-record, and a seat that reaches for a record tool nobody
	// installed fails MID-ROUND — the expensive failure this preflight exists to
	// move to setup time.
	Plugin  string
	Root    string
	Version string
}

// AssetName is THE CONTRACT between the release job and this program, in one function.
//
// The release workflow cross-compiles into `dist/${name}_${os}_${arch}${ext}` and checksums
// the directory; `fetchRelease` then asks GitHub for an asset by name and looks that name up
// in SHA256SUMS. Two programs, one string, and until this existed nothing joined them: the
// naming lived as a shell expansion in a workflow and as a Sprintf here, and a rename on
// either side would have failed SILENTLY at install time, on a user's machine — the release
// job would publish `sc-doctor_linux_amd64`, this would ask for something else, `gh` would
// match no asset, and the fetch would fall back to building from source as though no release
// existed.
//
// Taking goos/goarch as ARGUMENTS rather than reading runtime is what makes it testable: the
// contract has to hold for all six published pairs, not just the one the test happens to run
// on. release_contract_test.go reads the workflow and checks every pair against this.
func AssetName(name, goos, goarch string) string {
	// Not buildid.ExeName: this is a CROSS-compile name, so the extension follows the
	// argument goos, never the host. sc-doctor -fix on Linux must be able to name the
	// Windows asset, and a host-derived suffix would quietly produce the Linux one.
	ext := ""
	if goos == "windows" {
		ext = ".exe"
	}
	return fmt.Sprintf("%s_%s_%s%s", name, goos, goarch, ext)
}

// binariesOf lists every command under a plugin root's tools/cmd and whether
// bin/<name> exists.
func binariesOf(root, plugin, version string) []binStatus {
	entries, err := os.ReadDir(filepath.Join(root, "tools", "cmd"))
	if err != nil {
		return nil
	}
	var out []binStatus
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		bin := filepath.Join(root, "bin", e.Name()+buildid.ExeName(""))
		_, err := os.Stat(bin)
		out = append(out, binStatus{
			Name: e.Name(), Built: err == nil,
			Plugin: plugin, Root: root, Version: version,
		})
	}
	return out
}

// hookBinaries lists this plugin's own commands.
func hookBinaries(root string) []binStatus {
	return binariesOf(root, filepath.Base(filepath.Dir(root)), filepath.Base(root))
}

// siblingBinaries walks the marketplace cache for every OTHER installed SC plugin
// that ships a Go module, so `--fix` provisions the whole suite rather than only
// the plugin the doctor happens to live in. Each plugin owns its own bin/ and its
// own release tag; nothing here assumes a shared one.
func siblingBinaries(root string) []binStatus {
	marketDir := filepath.Dir(filepath.Dir(root))
	self := filepath.Base(filepath.Dir(root))
	entries, err := os.ReadDir(marketDir)
	if err != nil {
		return nil
	}
	var out []binStatus
	for _, e := range entries {
		if !e.IsDir() || e.Name() == self {
			continue
		}
		v := newestVersionDir(filepath.Join(marketDir, e.Name()))
		if v == "" {
			continue
		}
		out = append(out, binariesOf(filepath.Join(marketDir, e.Name(), v), e.Name(), v)...)
	}
	return out
}

// verdict computes READY / DEGRADED / BLOCKED from tool + binary status.
func verdict(tools []toolchain.Status, bins []binStatus) string {
	degraded := false
	for _, t := range tools {
		if !t.Found {
			// Absent BY DESIGN is not absent. A cloud session has no gh and is not
			// going to get one; counting it DEGRADED made the verdict permanently
			// wrong there, which is worse than useless — it teaches the operator to
			// discount the verdict on the one axis it exists to be trusted on.
			if t.NotApplicable {
				continue
			}
			if t.Tier == "required" {
				return "BLOCKED"
			}
			if t.Tier == "recommended" {
				degraded = true
			}
			continue
		}
		// PRESENT AND UNUSABLE COUNTS THE SAME AS ABSENT, because the work it gates cannot run
		// either way. A required tool below its minimum is BLOCKED — that is the whole point of
		// declaring the minimum, and a verdict that said READY over it would be the presence
		// probe again, one field along.
		if t.TooOld {
			if t.Tier == "required" {
				return "BLOCKED"
			}
			degraded = true
		}
		// An unverifiable minimum is NOT READY. It may well be fine; the run cannot say so, and
		// "we could not tell" belongs on the degraded side of a preflight whose whole job is to
		// make readiness legible before work starts.
		if t.VersionUnmeasured != "" {
			degraded = true
		}
	}
	for _, b := range bins {
		if !b.Built {
			degraded = true
			continue
		}
		// A STALE BINARY DEGRADES THE VERDICT, and until now it did not.
		//
		// The detection was already here and already correct: Compare returns Stale when the
		// build stamp's revision is not HEAD, and Describe already says "this binary predates
		// the source it ships with; run `doctor --fix` to rebuild" — under a comment observing
		// that "the whole failure mode is that nothing about a stale binary suggests a rebuild is
		// needed". All of it reached the TABLE and none of it reached the verdict, so the doctor
		// printed the row and returned READY in the same breath.
		//
		// MEASURED 2026-08-22. A fix to the PreToolUse hook's env injection was committed at
		// 07:50; the binaries in the plugin's bin/ were from 05:31. A debate ran for sixteen
		// minutes against the old hook, reproducing the exact defect the fix removed, and the
		// only reason it was caught is that the fix's author happened to check the binary's
		// mtime before believing a negative result. A hook binary is the case that punishes this
		// hardest: it keeps working perfectly, just with old behaviour.
		//
		// Dirty does NOT degrade: built from HEAD with uncommitted edits is the ordinary state of
		// a working tree, and degrading on it would make the verdict permanently wrong during
		// development — the failure the NotApplicable arm above already names.
		// ReferenceCommit, not HeadCommit: in an INSTALL the reference is the commit the
		// client recorded at install time, and using HEAD meant this arm never fired outside a
		// checkout — inert in the only place it runs in production.
		if ref := ReferenceCommit(b); ref != "" {
			if Compare(ReadBuildStamp(binPath(b)), ref) == Stale {
				degraded = true
			}
		}
	}
	if degraded {
		return "DEGRADED"
	}
	return "READY"
}

// UnstampedCount reports how many BUILT binaries carry no build stamp, so their staleness could
// not be judged either way.
//
// NOT MEASURED IS NOT MEASURED. An unstamped binary is what a plain `go build` produces, and
// Compare answers Unstamped for it — which is neither Current nor Stale. Folding that into READY
// would be the substitution this whole gate exists to refuse: the verdict would read exactly the
// same whether every binary was verified current or none of them could be checked at all.
func UnstampedCount(bins []binStatus) int {
	n := 0
	for _, b := range bins {
		if !b.Built {
			continue
		}
		if Compare(ReadBuildStamp(binPath(b)), ReferenceCommit(b)) == Unstamped {
			n++
		}
	}
	return n
}

// table renders the preflight report.
func table(tools []toolchain.Status, bins []binStatus) string {
	var sb strings.Builder
	for _, t := range tools {
		switch {
		// TOO OLD IS ITS OWN ROW, before the plain ✓ that used to swallow it. The disqualifying
		// number was already being printed beside the tick; what was missing was anything that
		// read it (#521). The remedy word is UPGRADE, because telling an operator to install
		// what they are demonstrably holding sends them to the wrong place.
		case t.Found && t.TooOld:
			fmt.Fprintf(&sb, "%-18s ✗ (%s) %s is below the minimum %s — upgrade: %s\n",
				t.Name, t.Tier, t.Version, t.MinVersion, t.Install[runtime.GOOS])
		// AND UNMEASURED IS NOT A PASS. A declared minimum that could not be read leaves the
		// question open, and rendering that as ✓ is the collapse the field exists to end.
		case t.Found && t.VersionUnmeasured != "":
			fmt.Fprintf(&sb, "%-18s ? present, but the minimum %s is NOT VERIFIED — %s\n",
				t.Name, t.MinVersion, t.VersionUnmeasured)
		case t.Found:
			fmt.Fprintf(&sb, "%-18s ✓ %s\n", t.Name, displayVersionOf(t))
		case t.NotApplicable:
			// No install string: printing one here is the false alarm this branch
			// exists to remove.
			fmt.Fprintf(&sb, "%-18s – n/a in %s (%s)\n", t.Name, toolchain.Environment(), t.Purpose)
		default:
			fmt.Fprintf(&sb, "%-18s ✗ (%s) install: %s\n", t.Name, t.Tier, t.Install[runtime.GOOS])
		}
	}
	// BUILT IS NOT THE SAME QUESTION AS CURRENT (#450). A binary that exists but predates its
	// source runs, exits 0, and does exactly what it did when it was built — so "✓ built" was the
	// same report for a working hook and for one whose fix merged weeks ago and never shipped.
	// Measured: sc-filechanged-rearm ran 23 commits behind its source while reporting ✓.
	//
	// The revision is read from the binary's own VCS stamp, which the Go toolchain writes on every
	// build with no flags — see builtfrom.go for why that is read rather than a stamp of our own.
	// ONLY ASKED WHERE IT MEANS SOMETHING. "Is this binary older than its source" is a question
	// about a DEVELOPMENT CHECKOUT. A consumer installs the plugin from a release asset into a
	// directory that is not a git tree at all: HEAD is unknowable there, every binary would be
	// annotated, and the one line that matters would drown in fourteen that do not. So an
	// unknowable HEAD means the check DOES NOT APPLY and is silent — not that everything is
	// suspect. (The consumer's version of this question is answered by the plugin version.)
	for _, b := range bins {
		// PER BINARY. This took its reference from bins[0] for the whole table, which was
		// already wrong the moment a second plugin shipped Go binaries: feov's and this
		// plugin's roots are different trees, and one plugin's HEAD says nothing about
		// another's binaries.
		ref := ReferenceCommit(b)
		mark := "✓ built"
		switch {
		case !b.Built:
			mark = "✗ not built (run -fix)"
		case ref == "":
			// Nothing to compare against: staleness is not a question here, so say nothing.
		default:
			// Only the non-current states earn extra words.
			if s := ReadBuildStamp(binPath(b)); Compare(s, ref) != Current {
				v := Compare(s, ref)
				mark = fmt.Sprintf("! %s — %s", v, Describe(v, s, ref))
			}
		}
		fmt.Fprintf(&sb, "%-18s %s\n", b.Name, mark)
	}
	return sb.String()
}

// binPath is where a binStatus's executable lives, matching binariesOf.
func binPath(b binStatus) string {
	return filepath.Join(b.Root, "bin", buildid.ExeName(b.Name))
}

// versionOf best-effort runs a check command and returns its first output line.
func versionOf(checkCmd string) string {
	fields := strings.Fields(checkCmd)
	if len(fields) == 0 {
		return ""
	}
	out, err := exec.Command(fields[0], fields[1:]...).CombinedOutput()
	if err != nil {
		return ""
	}
	if i := strings.IndexByte(string(out), '\n'); i >= 0 {
		return strings.TrimSpace(string(out[:i]))
	}
	return strings.TrimSpace(string(out))
}

// ---- Cross-plugin aggregation + dance-state (efficiency phase PR-C) ----

// pluginReq is one sibling plugin's requirements, tagged with its origin.
type pluginReq struct {
	Plugin string
	Tools  []toolchain.Tool
}

// semverLess compares dotted numeric versions (missing parts = 0).
func semverLess(a, b string) bool {
	pa, pb := strings.Split(a, "."), strings.Split(b, ".")
	for i := 0; i < 3; i++ {
		x, y := 0, 0
		if i < len(pa) {
			fmt.Sscanf(pa[i], "%d", &x)
		}
		if i < len(pb) {
			fmt.Sscanf(pb[i], "%d", &y)
		}
		if x != y {
			return x < y
		}
	}
	return false
}

// newestVersionDir returns the highest-semver subdirectory name, "" if none.
func newestVersionDir(pluginDir string) string {
	entries, err := os.ReadDir(pluginDir)
	if err != nil {
		return ""
	}
	best := ""
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if best == "" || semverLess(best, e.Name()) {
			best = e.Name()
		}
	}
	return best
}

// siblingRequirements walks the marketplace cache (root/../..) and aggregates every OTHER
// installed SC plugin's newest requirements.json — each plugin owns its manifest; the
// doctor reports the whole suite (backlog: doctor cross-plugin aggregation).
func siblingRequirements(root string) []pluginReq {
	marketDir := filepath.Dir(filepath.Dir(root)) // <plugin>/<version> -> cache root
	selfPlugin := filepath.Base(filepath.Dir(root))
	entries, err := os.ReadDir(marketDir)
	if err != nil {
		return nil
	}
	var out []pluginReq
	for _, e := range entries {
		if !e.IsDir() || e.Name() == selfPlugin {
			continue
		}
		v := newestVersionDir(filepath.Join(marketDir, e.Name()))
		if v == "" {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(marketDir, e.Name(), v, "requirements.json"))
		if err != nil {
			continue
		}
		var req requirements
		if err := json.Unmarshal(raw, &req); err != nil {
			continue
		}
		out = append(out, pluginReq{Plugin: e.Name(), Tools: req.Tools})
	}
	return out
}

// danceWarnings computes update-dance state from observable cache facts: a newer version
// dir than the one running means the dance is mid-flight; a newest dir with an empty bin/
// is the empty-bin window (the 0.7.0 crash-storm class). Deterministic verdicts, not
// remembered rituals.
func danceWarnings(root string) []string {
	pluginDir := filepath.Dir(root)
	current := filepath.Base(root)
	newest := newestVersionDir(pluginDir)
	var out []string
	if newest != "" && semverLess(current, newest) {
		out = append(out, fmt.Sprintf("DANCE INCOMPLETE: this doctor runs from %s but the cache holds %s — finish the dance (/plugin update -> /reload-plugins -> /reload-skills -> doctor --fix).", current, newest))
	}
	if newest != "" {
		binDir := filepath.Join(pluginDir, newest, "bin")
		entries, err := os.ReadDir(binDir)
		if err != nil || len(entries) == 0 {
			out = append(out, fmt.Sprintf("EMPTY-BIN WINDOW: cache %s has no hook binaries — hooks degrade to guard warnings until doctor --fix runs there.", newest))
		}
	}
	return out
}

// releaseRepo is where CI publishes the cross-compiled hook binaries.
const releaseRepo = "ctoforaday/special-circumstances"

// verifySHA256 checks that sums (the SHA256SUMS file content) records digest for asset.
func verifySHA256(sums, asset, digest string) bool {
	for _, line := range strings.Split(sums, "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 2 && strings.TrimPrefix(fields[1], "*") == asset {
			return strings.EqualFold(fields[0], digest)
		}
	}
	return false
}

// downloadArgs builds the `gh release download` argv for one binary.
//
// The release is PINNED to the owning plugin's tag. Falling back to "latest"
// would resolve to whichever plugin tagged most recently, whose assets do not
// contain this binary at all — a download failure that reads like a network
// problem rather than the version confusion it is.
//
// FIX (defect): that reasoning was in the comment and NOT in the code. The tag was
// built only when plugin and version were both known, and an unknown one silently
// omitted the positional tag argument — which is how `gh release download` is spelled
// for "latest". The comment named the hazard and the fallback branch implemented it.
//
// Reachability, measured rather than assumed: neither constructor can currently
// produce an empty Plugin or Version — hookBinaries derives both from path basenames,
// and siblingBinaries skips a plugin whose newestVersionDir is "". So this was a
// LATENT defect, not a live one. It is closed anyway, because the next constructor is
// what makes a latent guard a live bug, and an unpinned install is not a failure this
// suite should discover in the field.
func downloadArgs(b binStatus, asset, dir string) ([]string, error) {
	if b.Plugin == "" || b.Version == "" {
		return nil, fmt.Errorf("cannot pin a release tag for %s (plugin=%q version=%q) — refusing an untagged download, which resolves to whichever plugin tagged most recently",
			b.Name, b.Plugin, b.Version)
	}
	return []string{
		"release", "download", releaseTag(b),
		"--repo", releaseRepo, "--pattern", asset, "--pattern", "SHA256SUMS", "--dir", dir,
	}, nil
}

// releaseTag is the one place the tag is spelled, so the argv and the error message that
// explains a failure cannot disagree about what was asked for.
func releaseTag(b binStatus) string { return fmt.Sprintf("%s--v%s", b.Plugin, b.Version) }

// describeFetchFailure turns gh's output into a sentence that names the condition.
//
// THE MESSAGE WAS WRONG FOR EVERY REAL INVOCATION. "release download failed: <gh output>"
// reads as a network or auth problem — something a retry or a `gh auth login` might fix.
// Measured 2026-08-15 (#405): NO plugin has a tag matching its own version. frank-exchange
// -of-views ships 1.33.0 against a newest tag of 0.50.0, prosthetic-conscience 0.38.0
// against 0.16.0, gray-area 0.8.0 against 0.2.0, sleeper-service has no tag at all. So the
// download asks for a tag that does not exist EVERY time, and the operator was told the
// wrong thing about why 100% of the time.
//
// A no-match returns the raw output rather than a guess, which is the honest degradation:
// gh's wording is not a schema and this must not claim "no release exists" because a phrase
// moved. The miss lands on today's behaviour, not on a confident falsehood.
func describeFetchFailure(tag, ghOutput string) string {
	out := strings.TrimSpace(ghOutput)
	low := strings.ToLower(out)
	if strings.Contains(low, "release not found") || strings.Contains(low, "404") {
		return fmt.Sprintf("no release published for %s — this plugin's version has never been tagged, "+
			"so there is no CI-built asset to fetch (not a network problem; see #405)", tag)
	}
	return fmt.Sprintf("release download failed for %s: %s", tag, out)
}

// fetchRelease downloads the CI-built asset for this platform, verifies its
// checksum against the release SHA256SUMS, and installs it into bin/.
func fetchRelease(b binStatus) error {
	if !toolchain.Present("gh") {
		return errors.New("gh not on PATH")
	}
	root, name := b.Root, b.Name
	asset := AssetName(name, runtime.GOOS, runtime.GOARCH)
	tmp, err := os.MkdirTemp("", "sc-doctor-fetch-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)
	args, err := downloadArgs(b, asset, tmp)
	if err != nil {
		return err
	}
	dl := exec.Command("gh", args...)
	if msg, err := dl.CombinedOutput(); err != nil {
		// The cause is WRAPPED, not flattened. describeFetchFailure renders gh's output
		// for a human; err carries the exit status, and dropping it left a caller with no
		// way to tell "gh refused" from "gh was killed" — the two need different advice.
		return fmt.Errorf("%s: %w", describeFetchFailure(releaseTag(b), string(msg)), err)
	}
	data, err := os.ReadFile(filepath.Join(tmp, asset))
	if err != nil {
		return err
	}
	sums, err := os.ReadFile(filepath.Join(tmp, "SHA256SUMS"))
	if err != nil {
		return err
	}
	digest := fmt.Sprintf("%x", sha256.Sum256(data))
	if !verifySHA256(string(sums), asset, digest) {
		return fmt.Errorf("checksum mismatch for %s — refusing to install", asset)
	}
	dst := filepath.Join(root, "bin", buildid.ExeName(name))
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0o755)
}

// buildFromSource is the local-toolchain boundary: the dev-convenience fallback when
// no release asset can be fetched. It returns the build output for the report.
func buildFromSource(b binStatus) (string, error) {
	out := filepath.Join(b.Root, "bin", b.Name+buildid.ExeName(""))
	cmd := exec.Command("go", "build", "-C", filepath.Join(b.Root, "tools"), "-o", out, "./cmd/"+b.Name)
	msg, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(msg)), err
}

// fix provisions every missing binary: CI-built release asset first (checksum-
// verified); a from-source build is only the dev-convenience fallback.
func fix(bins []binStatus) []string {
	return fixWith(bins, fetchRelease, buildFromSource, toolchain.Present("go"))
}

// fixWith is fix with both process boundaries passed in — the network (gh) and the
// local toolchain (go build).
//
// They are parameters because the REPORT is the product here: this is what `--fix`
// prints to an operator whose session is already degraded, and every branch of it was
// untested. Nine of the fifteen operator-mutation survivors in this file lived in the
// fetch path, for the ordinary reason that a test cannot reach a function that shells
// out to the network. Now it can.
func fixWith(bins []binStatus, fetch func(binStatus) error, build func(binStatus) (string, error), goPresent bool) []string {
	var report []string
	for _, b := range bins {
		if b.Built {
			continue
		}
		fetchErr := fetch(b)
		if fetchErr == nil {
			report = append(report, fmt.Sprintf("%s (%s): fetched CI-built release asset (checksum verified)", b.Name, b.Plugin))
			continue
		}
		if goPresent {
			if msg, err := build(b); err != nil {
				report = append(report, fmt.Sprintf("%s: fetch failed (%v); BUILD FAILED: %s", b.Name, fetchErr, msg))
			} else {
				report = append(report, fmt.Sprintf("%s: fetch failed (%v); built from source (dev fallback)", b.Name, fetchErr))
			}
		} else {
			// AssetName, not a second assembly of it. This message names the file the
			// human must place, and the release job names the file it publishes; those are
			// the SAME fact, so they get one statement of it. Hand-composing the pattern
			// here meant a change to the contract would leave this advice pointing at a
			// filename that never existed.
			report = append(report, fmt.Sprintf("%s: fetch failed (%v) and Go not found — install gh or Go, or place %s into %s manually",
				b.Name, fetchErr, AssetName(b.Name, runtime.GOOS, runtime.GOARCH), filepath.Join(b.Root, "bin")))
		}
	}
	return report
}

// resolveRoot picks the plugin root: the -root flag wins, then CLAUDE_PLUGIN_ROOT,
// then the binary's own grandparent (bin/sc-doctor -> plugin root). executable is a
// parameter so the fallback is reachable in a test.
func resolveRoot(rootFlag, envRoot string, executable func() (string, error)) string {
	if rootFlag != "" {
		return rootFlag
	}
	if envRoot != "" {
		return envRoot
	}
	if exe, err := executable(); err == nil {
		return filepath.Dir(filepath.Dir(exe))
	}
	return ""
}

// run is the doctor with its process boundary passed in. fixFn is a parameter because
// the real one reaches the network and the local Go toolchain; a preflight's REPORTING
// must be testable without either.
func run(args []string, stdout io.Writer, envRoot string, executable func() (string, error), fixFn func([]binStatus) []string) int {
	fs := flag.NewFlagSet("sc-doctor", flag.ContinueOnError)
	fs.SetOutput(stdout)
	doFix := fs.Bool("fix", false, "build/fetch missing hook binaries")
	showVersion := fs.Bool("version", false, "print version and exit")
	rootFlag := fs.String("root", "", "plugin root (default: CLAUDE_PLUGIN_ROOT or the binary's parent dir)")
	if err := fs.Parse(args); err != nil {
		return 0
	}
	if *showVersion {
		fmt.Fprintln(stdout, buildid.Line("sc-doctor"))
		return 0
	}

	root := resolveRoot(*rootFlag, envRoot, executable)

	raw, err := os.ReadFile(filepath.Join(root, "requirements.json"))
	if err != nil {
		fmt.Fprintf(stdout, "sc-doctor: cannot read requirements.json under %q: %v\n", root, err)
		return 0
	}
	var req requirements
	if err := json.Unmarshal(raw, &req); err != nil {
		fmt.Fprintf(stdout, "sc-doctor: malformed requirements.json: %v\n", err)
		return 0
	}

	// THE MANIFEST'S OWN DIRECTORY, because a tool declaring CheckDir is asking a question
	// whose answer depends on where it is asked — see toolchain.Tool.CheckDir (#521).
	tools := toolchain.ProbeInDir(req.Tools, toolchain.Environment(), root)
	// Own binaries AND every sibling plugin's: the suite is provisioned as a
	// whole, because a seat blocked by another plugin's missing binary fails just
	// as hard as one blocked by ours.
	allBins := func() []binStatus { return append(hookBinaries(root), siblingBinaries(root)...) }
	bins := allBins()

	if *doFix {
		for _, line := range fixFn(bins) {
			fmt.Fprintln(stdout, line)
		}
		bins = allBins() // re-probe
	}

	fmt.Fprint(stdout, table(tools, bins))

	// Cross-plugin aggregation: every installed SC plugin's own requirements.json,
	// probed and reported here — one preflight for the whole suite.
	//
	// AND THE SIBLINGS REACH THE VERDICT, which for a long time they did not. Their tables
	// printed and `verdict` then read the doctor's OWN manifest alone, so `required` in any
	// plugin but this one could not produce BLOCKED: frank-exchange-of-views declares `node`
	// required and its engine refuses dispatch without it, gray-area declares `git` required,
	// and both showed ✗ over a verdict that said fine. Two readers of one board disagreeing,
	// with the plugin author's hard requirement the half that lost.
	suite := []([]toolchain.Status){tools}
	for _, pr := range siblingRequirements(root) {
		probed := toolchain.Probe(pr.Tools)
		fmt.Fprintf(stdout, "-- %s --\n", pr.Plugin)
		fmt.Fprint(stdout, table(probed, nil))
		suite = append(suite, probed)
	}
	// One obligation per tool, strictest declaration winning — see toolchain.MergeStrictest for
	// why the strongest need is the box's answer and why order cannot decide it.
	obligations := toolchain.MergeStrictest(suite...)

	warnings := danceWarnings(root)
	for _, w := range warnings {
		fmt.Fprintln(stdout, w)
	}
	v := verdict(obligations, bins)
	if v == "READY" && len(warnings) > 0 {
		v = "DEGRADED"
	}
	// AND SAY WHAT COULD NOT BE JUDGED, in the same line as the verdict.
	//
	// An unstamped binary is what a plain `go build` produces, and it is neither current nor
	// stale — Compare says Unstamped. Printing a bare READY over a set nobody could check would
	// read identically to READY over a set verified current, which is the substitution the
	// staleness arm above exists to refuse. It qualifies the verdict rather than changing it:
	// unstamped is not evidence of a problem, only the absence of evidence against one.
	// The two reasons a staleness question goes unanswered are reported SEPARATELY, because
	// they name different parties. An unstamped binary is a build problem; a missing reference
	// is the doctor having nothing to compare against, and saying "carries no build stamp"
	// about a stamped binary sent every reader after the wrong thing.
	if n := UnstampedCount(bins); n > 0 {
		fmt.Fprintf(stdout, "VERDICT: %s (%d binar%s carry no build stamp — staleness NOT MEASURED for %s)\n",
			v, n, map[bool]string{true: "y", false: "ies"}[n == 1], map[bool]string{true: "it", false: "them"}[n == 1])
		return 0
	}
	if n := NoReferenceCount(bins); n > 0 {
		fmt.Fprintf(stdout, "VERDICT: %s (%d stamped binar%s have no reference commit to check against — staleness NOT MEASURED for %s)\n",
			v, n, map[bool]string{true: "y", false: "ies"}[n == 1], map[bool]string{true: "it", false: "them"}[n == 1])
		return 0
	}
	fmt.Fprintln(stdout, "VERDICT:", v)
	return 0
}

// Main is the process boundary: it wires the real environment in and returns the
// exit code, so cmd/ stays a three-line shim and this stays testable.
func Main() int {
	return run(os.Args[1:], os.Stdout, os.Getenv("CLAUDE_PLUGIN_ROOT"), os.Executable, fix)
}

// displayVersion prefers the line the check command already produced during Probe over running
// it a second time. Only tools declaring a minimum are measured there, so everything else still
// pays exactly one exec for its version — the same cost as before, asked once instead of twice.
func displayVersionOf(t toolchain.Status) string {
	if t.VersionLine != "" {
		return t.VersionLine
	}
	return versionOf(t.CheckCmd)
}
