package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/ctoforaday/special-circumstances/scripts/internal/staticproof"
)

// PluginFromRef is the plugin a release tag names: refs/tags/<plugin>--v<version>.
func PluginFromRef(ref string) (string, error) {
	tag, ok := strings.CutPrefix(ref, "refs/tags/")
	if !ok {
		return "", fmt.Errorf("%q is not a tag ref; a release is a refs/tags/<plugin>--v<version> ref", ref)
	}
	plugin, version, ok := strings.Cut(tag, "--v")
	if !ok || plugin == "" || version == "" {
		return "", fmt.Errorf("tag %q does not name <plugin>--v<version>", tag)
	}
	return plugin, nil
}

// AssetExport is one ASSETS line: the file installed beside the helper, and the release asset
// that carries it for each arch.
type AssetExport struct {
	File   string
	ByArch map[string]string
}

// ReadAssets parses a helper's ASSETS file: each non-empty, non-comment line is the installed
// file name followed by <arch>=<release asset> fields.
func ReadAssets(path string) ([]AssetExport, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var out []AssetExport
	for n, line := range strings.Split(string(b), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 0 || strings.HasPrefix(fields[0], "#") {
			continue
		}
		a := AssetExport{File: fields[0], ByArch: map[string]string{}}
		for _, f := range fields[1:] {
			arch, asset, ok := strings.Cut(f, "=")
			if !ok || arch == "" || asset == "" {
				return nil, fmt.Errorf("%s:%d: field %q is not <arch>=<release asset>", path, n+1, f)
			}
			a.ByArch[arch] = asset
		}
		out = append(out, a)
	}
	return out, nil
}

// ExpectedDist is every name a plugin's release dist holds, each mapped to the artifact whose
// staticness contract it meets, or to the empty string for a file that is not an executable (an
// ASSETS export). The names are built from their parts, never parsed back out of a file name:
// every command under tools/cmd and every helper under native/ (a directory holding ASSETS), each
// for every published pair, and every ASSETS export for the arches the pairs name.
func ExpectedDist(pluginDir string, targets []Target) (map[string]string, error) {
	cmdDir := filepath.Join(pluginDir, "tools", "cmd")
	entries, err := os.ReadDir(cmdDir)
	if err != nil {
		return nil, fmt.Errorf("listing the plugin's commands: %w", err)
	}
	want := map[string]string{}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		for _, t := range targets {
			want[fmt.Sprintf("%s_%s_%s%s", e.Name(), t.OS, t.Arch, t.Ext)] = e.Name()
		}
	}
	if len(want) == 0 {
		return nil, fmt.Errorf("%s holds no commands for the published pairs", cmdDir)
	}
	nativeDir := filepath.Join(pluginDir, "native")
	helpers, err := FindHelpers(nativeDir)
	if err != nil {
		return nil, err
	}
	for _, h := range helpers {
		for _, t := range targets {
			want[fmt.Sprintf("%s_%s_%s%s", h, t.OS, t.Arch, t.Ext)] = h
		}
		exports, err := ReadAssets(filepath.Join(nativeDir, h, "ASSETS"))
		if err != nil {
			return nil, err
		}
		for _, a := range exports {
			for _, t := range targets {
				asset, ok := a.ByArch[t.Arch]
				if !ok {
					return nil, fmt.Errorf("%s ASSETS: %s names no release asset for %s", h, a.File, t.Arch)
				}
				want[asset] = ""
			}
		}
	}
	return want, nil
}

// CheckDist holds distDir to exactly the names ExpectedDist builds for the plugin at pluginDir,
// plus SHA256SUMS when it is already there: an unexpected name or a missing one is refused. Every
// executable in it passes staticproof against its own artifact's allowlist.
func CheckDist(distDir, pluginDir string, targets []Target, stdout io.Writer) error {
	want, err := ExpectedDist(pluginDir, targets)
	if err != nil {
		return err
	}
	entries, err := os.ReadDir(distDir)
	if err != nil {
		return fmt.Errorf("reading dist dir %s: %w", distDir, err)
	}
	present := map[string]bool{}
	var unexpected []string
	for _, e := range entries {
		name := e.Name()
		present[name] = true
		if _, ok := want[name]; (ok || name == "SHA256SUMS") && !e.IsDir() {
			continue
		}
		unexpected = append(unexpected, name)
	}
	var missing []string
	for name := range want {
		if !present[name] {
			missing = append(missing, name)
		}
	}
	slices.Sort(unexpected)
	slices.Sort(missing)
	if len(unexpected) > 0 || len(missing) > 0 {
		return fmt.Errorf("dist %s does not hold exactly the release's assets: missing %v, unexpected %v", distDir, missing, unexpected)
	}
	names := make([]string, 0, len(want))
	for name := range want {
		names = append(names, name)
	}
	slices.Sort(names)
	for _, name := range names {
		artifact := want[name]
		if artifact == "" {
			continue
		}
		if err := staticproof.Check(artifact, filepath.Join(distDir, name)); err != nil {
			return err
		}
		fmt.Fprintf(stdout, "proven static: %s (%s)\n", name, artifact)
	}
	return nil
}
