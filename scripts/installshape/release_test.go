package installshape

// THE RELEASE THIS PR SHIPS, BUILT HERE, SERVED THROUGH SC_RELEASE_BASE_URL.
//
// Every Go binary is built from this tree exactly as the release job builds it (CGO off, -trimpath,
// -s -w), for this host. The native helpers and their exports are STUBS: building them needs the
// pinned zig/cmake toolchain and ONNX Runtime, minutes per helper. fetch-bin installs all or
// none, so the stubs are real release assets in every way fetch-bin checks — named per ASSETS,
// listed in SHA256SUMS — and only their bytes are not a helper.
//
// Two more releases sit beside it, and neither holds a working binary:
//   - <plugin>--v0.0.1, the version the upgrade leg installs first;
//   - the host-cache decoy at 99.0.0 is filled from the same tripwires directly.
// Their files are TRIPWIRES: scripts that write their own path to the trace and exit 0. A reader
// that resolves the wrong tree runs a tripwire, and the trace names the tree it reached.

import (
	"bufio"
	"crypto/sha256"
	"debug/buildinfo"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"
)

// plugin is what the checkout says one plugin carries, and what its release must therefore hold.
type plugin struct {
	name, version string
	cmds          []string          // tools/cmd/* — one binary each
	installed     map[string]string // bin/<installed name> → release asset name, for this host
}

func (p plugin) tag(version string) string { return p.name + "--v" + version }

// hasBinaries is whether fetch-bin installs anything for this plugin.
func (p plugin) hasBinaries() bool { return len(p.installed) > 0 }

// readPlugin derives a plugin's binary set the way fetch-bin.sh does: one per tools/cmd
// directory, one per native/*/ASSETS helper, and each ASSETS line's export for this arch.
func readPlugin(t *testing.T, dir string) plugin {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(dir, ".claude-plugin", "plugin.json"))
	if err != nil {
		t.Fatal(err)
	}
	var pj struct{ Name, Version string }
	if err := jsonUnmarshal(b, &pj); err != nil {
		t.Fatalf("%s: %v", dir, err)
	}
	p := plugin{name: pj.Name, version: pj.Version, installed: map[string]string{}}
	cmds, _ := filepath.Glob(filepath.Join(dir, "tools", "cmd", "*"))
	for _, c := range cmds {
		if fi, err := os.Stat(c); err == nil && fi.IsDir() {
			n := filepath.Base(c)
			p.cmds = append(p.cmds, n)
			p.installed[n] = n + "_" + runtime.GOOS + "_" + runtime.GOARCH
		}
	}
	sort.Strings(p.cmds)
	assets, _ := filepath.Glob(filepath.Join(dir, "native", "*", "ASSETS"))
	for _, a := range assets {
		h := filepath.Base(filepath.Dir(a))
		p.installed[h] = h + "_" + runtime.GOOS + "_" + runtime.GOARCH
		f, err := os.Open(a)
		if err != nil {
			t.Fatal(err)
		}
		sc := bufio.NewScanner(f)
		for sc.Scan() {
			fields := strings.Fields(sc.Text())
			if len(fields) == 0 || strings.HasPrefix(fields[0], "#") {
				continue
			}
			export := ""
			for _, pair := range fields[1:] {
				if arch, asset, ok := strings.Cut(pair, "="); ok && arch == runtime.GOARCH {
					export = asset
				}
			}
			if export == "" {
				t.Fatalf("%s: %s names no export for %s", a, fields[0], runtime.GOARCH)
			}
			p.installed[fields[0]] = export
		}
		f.Close()
	}
	return p
}

// release is a local release directory and what it holds.
type release struct {
	dir, url string
	sums     map[string]map[string]string // tag → asset → sha256
	revision string                       // the build stamp the Go binaries carry, as buildid.Revision reads it
}

// assetSum is the sha256 of the asset installed as bin/<name> from tag.
func (r *release) assetSum(p plugin, tag, name string) string {
	return r.sums[tag][p.installed[name]]
}

// buildRelease builds this tree's release for every plugin with binaries, and the tripwire
// release at oldVersion beside it.
func buildRelease(t *testing.T, repo, dir string, plugins []plugin, oldVersion, trace string) *release {
	t.Helper()
	r := &release{dir: dir, url: "file://" + dir, sums: map[string]map[string]string{}}
	for _, p := range plugins {
		if !p.hasBinaries() {
			continue
		}
		tagDir := filepath.Join(dir, p.tag(p.version))
		mkdir(t, tagDir)
		if len(p.cmds) > 0 {
			out := t.TempDir()
			cmd := exec.Command("go", "build", "-C", filepath.Join(repo, "plugins", p.name, "tools"),
				"-trimpath", "-ldflags", "-s -w", "-o", out+string(filepath.Separator), "./cmd/...")
			cmd.Env = append(os.Environ(), "CGO_ENABLED=0")
			if b, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("building %s's release binaries: %v\n%s", p.name, err, b)
			}
			for _, n := range p.cmds {
				copyFile(t, filepath.Join(out, n), filepath.Join(tagDir, p.installed[n]), 0o755)
			}
		}
		for name, asset := range p.installed {
			if contains(p.cmds, name) {
				continue
			}
			// A stub: unique to this build, so a byte comparison cannot pass against another.
			write(t, filepath.Join(tagDir, asset), fmt.Sprintf("#!/bin/sh\n# installshape stub release asset %s (%s) of %s\nexit 3\n", asset, name, p.tag(p.version)), 0o755)
		}
		r.sums[p.tag(p.version)] = writeSums(t, tagDir)

		old := filepath.Join(dir, p.tag(oldVersion))
		mkdir(t, old)
		for _, asset := range p.installed {
			write(t, filepath.Join(old, asset), tripwire(trace), 0o755)
		}
		r.sums[p.tag(oldVersion)] = writeSums(t, old)
	}
	for _, p := range plugins {
		if len(p.cmds) == 0 {
			continue
		}
		rev, err := revisionOf(filepath.Join(dir, p.tag(p.version), p.installed[p.cmds[0]]))
		if err != nil {
			t.Fatal(err)
		}
		if r.revision != "" && rev != r.revision {
			t.Fatalf("the release's binaries carry two build stamps, %s and %s", r.revision, rev)
		}
		r.revision = rev
	}
	return r
}

// revisionOf reads a binary's build stamp the way shared/buildid.Revision reads its own.
func revisionOf(path string) (string, error) {
	info, err := buildinfo.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("%s carries no build info: %w", path, err)
	}
	rev, modified := "", false
	for _, kv := range info.Settings {
		switch kv.Key {
		case "vcs.revision":
			rev = kv.Value
		case "vcs.modified":
			modified = kv.Value == "true"
		}
	}
	if rev == "" {
		return "unknown", nil
	}
	if len(rev) > 7 {
		rev = rev[:7]
	}
	if modified {
		rev += "+dirty"
	}
	return rev, nil
}

// writeSums writes SHA256SUMS over every file in dir and returns asset → sum.
func writeSums(t *testing.T, dir string) map[string]string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	sums := map[string]string{}
	var b strings.Builder
	for _, e := range entries {
		if e.IsDir() || e.Name() == "SHA256SUMS" {
			continue
		}
		s := fileSum(t, filepath.Join(dir, e.Name()))
		sums[e.Name()] = s
		fmt.Fprintf(&b, "%s  %s\n", s, e.Name())
	}
	write(t, filepath.Join(dir, "SHA256SUMS"), b.String(), 0o644)
	return sums
}

// tripwire is a stand-in binary that records the path it was run as, and its arguments.
func tripwire(trace string) string {
	return "#!/bin/sh\n# installshape tripwire\nprintf '%s\\t%s\\n' \"$0\" \"$*\" >> " + shellQuote(trace) + "\ncat >/dev/null 2>&1\nexit 0\n"
}

func fileSum(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	s := sha256.Sum256(b)
	return hex.EncodeToString(s[:])
}

func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

func contains(xs []string, x string) bool {
	for _, v := range xs {
		if v == x {
			return true
		}
	}
	return false
}
