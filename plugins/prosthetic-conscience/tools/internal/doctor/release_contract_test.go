package doctor

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"testing"
)

// THE RELEASE PATH IS THE LARGEST UNTESTED SURFACE IN THIS REPOSITORY, and this file is the
// part of it that can be tested without publishing anything.
//
// The workflow cross-compiles into `dist/${name}_${os}_${arch}${ext}` and checksums that
// directory. `fetchRelease` asks GitHub for an asset BY NAME and looks the same name up in
// SHA256SUMS. Two programs, one string contract, and nothing joined them — the plan document
// that governs this path claimed the bash was "tested by use", which means tested in
// production, once per tag, with no way to fail safely.
//
// The failure mode is quiet, which is why it is worth a test: a rename on either side does not
// error. `gh` matches no asset, the fetch fails, and `fix` falls back to building from source
// exactly as it would if no release existed at all. The user gets a working binary and never
// learns the release assets have been unreachable for months.
//
// These tests read the WORKFLOW as the authority rather than restating its naming scheme here.
// A third declaration of the same fact would be the defect, not the fix.

// repoRoot walks up to the directory holding .github, so these tests do not depend on where
// `go test` was invoked from.
func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 12; i++ {
		if _, err := os.Stat(filepath.Join(dir, ".github")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	t.Skip("no .github/ above the test directory — not a full checkout")
	return ""
}

func workflowSource(t *testing.T) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(repoRoot(t), ".github", "workflows", "hooks.yml"))
	if err != nil {
		t.Fatalf("reading the workflow, which is the AUTHORITY for the asset contract: %v", err)
	}
	return string(b)
}

var (
	// The `for pair in windows/amd64 …` line: the published platform set.
	rePairs = regexp.MustCompile(`for pair in ((?:[a-z0-9]+/[a-z0-9]+\s*)+);`)
	// The -o template the build writes to.
	reOutput = regexp.MustCompile(`-o "dist/([^"]+)"`)
	// The windows extension rule.
	reExtRule = regexp.MustCompile(`\[ "\$os" = "(windows)" \] && ext="(\.[a-z]+)"`)
)

// publishedPairs is the platform set the release job actually builds, read from the workflow.
func publishedPairs(t *testing.T, wf string) [][2]string {
	t.Helper()
	m := rePairs.FindStringSubmatch(wf)
	if m == nil {
		t.Fatal("could not find the platform-pair loop in the workflow. If the release job was " +
			"restructured, this test must be rewritten against the new shape rather than deleted — " +
			"an asset-name contract with no test is how it broke silently before.")
	}
	var pairs [][2]string
	for _, f := range strings.Fields(m[1]) {
		os, arch, ok := strings.Cut(f, "/")
		if !ok {
			t.Fatalf("unparseable platform pair %q", f)
		}
		pairs = append(pairs, [2]string{os, arch})
	}
	if len(pairs) == 0 {
		t.Fatal("the platform set parsed as EMPTY — a vacuous pass is what this test exists to prevent")
	}
	return pairs
}

// THE JOIN. For every platform the release job publishes, the name it writes must be the name
// this program asks for.
func TestAssetNameMatchesWhatTheReleaseJobWrites(t *testing.T) {
	wf := workflowSource(t)
	pairs := publishedPairs(t, wf)

	out := reOutput.FindStringSubmatch(wf)
	if out == nil {
		t.Fatal("could not find the release job's -o template")
	}
	template := out[1] // ${name}_${os}_${arch}${ext}

	ext := reExtRule.FindStringSubmatch(wf)
	if ext == nil {
		t.Fatal("could not find the windows extension rule; without it this test cannot tell " +
			"whether .exe is still expected")
	}
	extOS, extVal := ext[1], ext[2]

	// Render the workflow's shell template the way the shell would, for one binary name.
	const name = "sc-doctor"
	shellRender := func(goos, goarch string) string {
		e := ""
		if goos == extOS {
			e = extVal
		}
		r := strings.NewReplacer("${name}", name, "${os}", goos, "${arch}", goarch, "${ext}", e)
		return r.Replace(template)
	}

	for _, p := range pairs {
		goos, goarch := p[0], p[1]
		want := shellRender(goos, goarch)
		got := AssetName(name, goos, goarch)
		if got != want {
			t.Errorf("%s/%s: the release job writes %q and this program asks for %q.\n"+
				"gh would match no asset, the fetch would fail, and `fix` would fall back to "+
				"building from source exactly as if no release existed — silently.",
				goos, goarch, want, got)
		}
	}
}

// The windows rule is the one asymmetry in the contract, and the one a refactor drops first.
func TestOnlyWindowsGetsAnExtension(t *testing.T) {
	for _, p := range publishedPairs(t, workflowSource(t)) {
		got := AssetName("x", p[0], p[1])
		hasExt := strings.HasSuffix(got, ".exe")
		if p[0] == "windows" && !hasExt {
			t.Errorf("%s/%s produced %q with no .exe", p[0], p[1], got)
		}
		if p[0] != "windows" && hasExt {
			t.Errorf("%s/%s produced %q with an .exe it must not have", p[0], p[1], got)
		}
	}
}

// verifySHA256 is the OTHER end of the same string. The release job writes
// `sha256sum * > SHA256SUMS`, whose format is "<digest>  <name>" — and GNU sha256sum marks
// binary mode with a leading '*' on the name, which is why verifySHA256 trims it. A test that
// only fed it clean names would pass while the real file never matched.
func TestVerifySHA256AcceptsTheFormatTheReleaseJobWrites(t *testing.T) {
	asset := AssetName("sc-doctor", "windows", "amd64")
	const digest = "abc123"
	for _, line := range []string{
		digest + "  " + asset, // sha256sum text mode
		digest + " *" + asset, // sha256sum binary mode
	} {
		if !verifySHA256(line+"\n", asset, digest) {
			t.Errorf("verifySHA256 rejected a line sha256sum can emit: %q", line)
		}
	}
	if verifySHA256(digest+"  someone-elses-asset\n", asset, digest) {
		t.Error("verifySHA256 matched a DIFFERENT asset's line — the digest is not the key, the name is")
	}
}

// EVERY published platform must actually build. A cmd that compiles on the developer's box and
// not on windows/arm64 is discovered at TAG TIME today, when the release job is the first thing
// that has ever tried it.
//
// Compile-only (`go build -o os.DevNull`-style via a temp dir) rather than a full release
// build: the question is whether the code is buildable for that platform, not whether the
// linker flags produce a shippable artifact.
func TestEveryPublishedPlatformCompiles(t *testing.T) {
	if testing.Short() {
		t.Skip("cross-compiling every platform is slow; -short skips it")
	}
	requireCrossCompileHost(t)
	root := repoRoot(t)
	pairs := publishedPairs(t, workflowSource(t))

	cmds, err := filepath.Glob(filepath.Join(root, "plugins", "prosthetic-conscience", "tools", "cmd", "*"))
	if err != nil || len(cmds) == 0 {
		t.Fatalf("no cmd/ binaries found (%v) — a vacuous pass", err)
	}
	sort.Strings(cmds)

	tmp := t.TempDir()
	for _, p := range pairs {
		goos, goarch := p[0], p[1]
		for _, c := range cmds {
			name := filepath.Base(c)
			outPath := filepath.Join(tmp, AssetName(name, goos, goarch))
			cmd := exec.Command("go", "build", "-trimpath", "-o", outPath, "./cmd/"+name)
			cmd.Dir = filepath.Join(root, "plugins", "prosthetic-conscience", "tools")
			cmd.Env = append(os.Environ(), "GOOS="+goos, "GOARCH="+goarch, "CGO_ENABLED=0")
			if out, err := cmd.CombinedOutput(); err != nil {
				t.Errorf("%s does not build for %s/%s — this is discovered at TAG TIME today:\n%s",
					name, goos, goarch, strings.TrimSpace(string(out)))
				continue
			}
			if _, err := os.Stat(outPath); err != nil {
				t.Errorf("%s/%s built but produced no file at the contracted name %q",
					goos, goarch, filepath.Base(outPath))
			}
		}
	}
}

// The dry run end to end: build one binary for every platform, checksum the directory the way
// the release job does, and confirm every contracted name is present in SHA256SUMS.
func TestSHA256SUMSCoversEveryContractedName(t *testing.T) {
	if testing.Short() {
		t.Skip("builds; -short skips it")
	}
	requireCrossCompileHost(t)
	if _, err := exec.LookPath("sha256sum"); err != nil {
		// LOUD, not folded into a pass: on a machine without sha256sum this check did not run.
		t.Skip("sha256sum not on PATH — the SHA256SUMS half of the contract was NOT checked here")
	}
	root := repoRoot(t)
	pairs := publishedPairs(t, workflowSource(t))
	dist := t.TempDir()

	const name = "sc-doctor"
	var want []string
	for _, p := range pairs {
		asset := AssetName(name, p[0], p[1])
		cmd := exec.Command("go", "build", "-trimpath", "-o", filepath.Join(dist, asset), "./cmd/"+name)
		cmd.Dir = filepath.Join(root, "plugins", "prosthetic-conscience", "tools")
		cmd.Env = append(os.Environ(), "GOOS="+p[0], "GOARCH="+p[1], "CGO_ENABLED=0")
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("building %s: %v\n%s", asset, err, out)
		}
		want = append(want, asset)
	}

	// NO SHELL. The first version of this ran `sh -c "cd <dist> && sha256sum * > SHA256SUMS"`,
	// which failed on the Windows runner — not because sha256sum is missing (Git Bash ships
	// it, so the LookPath guard above did not fire) but because an interpolated Windows path
	// went through a shell that ate its backslashes: `cd C:UsersRUNNER~1AppData...`.
	//
	// Running the tool directly with cmd.Dir and capturing stdout removes the shell from the
	// test entirely. It is also closer to what is being verified: the release job's `sha256sum
	// * > SHA256SUMS` differs from this only in who expands the glob, and the bytes sha256sum
	// emits are the same either way.
	sums := exec.Command("sha256sum", want...)
	sums.Dir = dist
	stdout, err := sums.Output()
	if err != nil {
		t.Fatalf("checksumming: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dist, "SHA256SUMS"), stdout, 0o644); err != nil {
		t.Fatal(err)
	}
	body := stdout

	for _, asset := range want {
		digest := ""
		for _, line := range strings.Split(string(body), "\n") {
			f := strings.Fields(line)
			if len(f) >= 2 && strings.TrimPrefix(f[1], "*") == asset {
				digest = f[0]
			}
		}
		if digest == "" {
			t.Errorf("%s is missing from SHA256SUMS — fetchRelease would refuse to install it", asset)
			continue
		}
		// And the consumer must accept the line the producer just wrote.
		if !verifySHA256(string(body), asset, digest) {
			t.Errorf("verifySHA256 rejected the SHA256SUMS line for %s that sha256sum itself wrote", asset)
		}
	}
	if t.Failed() {
		t.Log(fmt.Sprintf("dist/ contained: %v", want))
	}
}

// requireCrossCompileHost keeps the build-heavy checks on ONE host.
//
// `go build` for windows/arm64 produces the same answer whether it runs on Linux or on
// Windows — cross-compilation does not depend on the host — so running these on both CI legs
// costs minutes and buys nothing. The Windows leg took 150s on the doctor package for exactly
// that duplicated work.
//
// Skipping is stated rather than silent, and the pure contract tests (the asset name, the
// .exe rule, the SHA256SUMS format) still run on EVERY platform, which is where a
// platform-specific naming bug would actually show up.
func requireCrossCompileHost(t *testing.T) {
	t.Helper()
	if runtime.GOOS != "linux" {
		t.Skipf("cross-compile checks run on the linux leg only (host is %s); the output does "+
			"not depend on the host, so a second leg would duplicate ~66 builds", runtime.GOOS)
	}
}

// The platform SET can shrink silently, and nothing above would notice.
//
// Found by taking a surviving mutant seriously rather than dismissing it: adding a platform to
// the workflow killed no test, correctly — `AssetName` is generic over goos, so a new platform
// cannot break the name contract. But the inverse is not harmless. Drop `linux/arm64` from the
// loop and the release simply stops carrying it; `fix` on that machine finds no asset, falls
// back to building from source, and reports success. The platform disappears from releases with
// no error anywhere.
//
// The floor asserted here is the one that is defensible without guessing at users: the
// platforms THIS REPOSITORY'S OWN CI runs on. If CI tests a platform, releases must carry it —
// otherwise the project is testing a target it does not ship.
func TestPublishedSetCoversThePlatformsCITestsOn(t *testing.T) {
	wf := workflowSource(t)
	published := map[string]bool{}
	for _, p := range publishedPairs(t, wf) {
		published[p[0]+"/"+p[1]] = true
	}
	// runs-on: ubuntu-latest and windows-latest are amd64 hosts.
	required := map[string]string{
		"linux/amd64":   "ubuntu-latest runners",
		"windows/amd64": "windows-latest runners",
	}
	for pair, why := range required {
		if !published[pair] {
			t.Errorf("CI tests on %s (%s) and the release job does not publish it — "+
				"`fix` on that platform finds no asset, silently builds from source, and reports success",
				pair, why)
		}
	}
	if len(published) < 2 {
		t.Errorf("the published platform set is %d entries; a set this small is more likely a "+
			"parse failure than a decision", len(published))
	}
}
