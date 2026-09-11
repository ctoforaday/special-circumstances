// Command marketplacegen pins every plugin that ships binaries to its own release tag in
// .claude-plugin/marketplace.json.
//
// Dev tooling for this repository only.
//
// # Why
//
// A relative-path source is copied from the marketplace's HEAD, and only the version string
// decides whether a consumer updates, so consumers ran main's hooks.json and skills against
// binaries built at the tag. A git-subdir source pinned to <name>--v<version> installs the
// release's own tree, whose hooks call only binaries that release carries.
//
// The ref is a function of plugin.json's version, so it is generated rather than hand-kept;
// -check fails when a version bump did not regenerate it. A plugin without tools/cmd has no
// binaries and no tag (the release job refuses one), so it keeps its relative path.
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ctoforaday/special-circumstances/scripts/internal/gitx"
)

const (
	manifest = ".claude-plugin/marketplace.json"
	repoURL  = "https://github.com/ctoforaday/special-circumstances.git"
)

type gitSubdir struct {
	Source string `json:"source"`
	URL    string `json:"url"`
	Path   string `json:"path"`
	Ref    string `json:"ref"`
}

type entry struct {
	Name        string          `json:"name"`
	Source      json.RawMessage `json:"source"`
	Description string          `json:"description"`
}

// market is the whole manifest. Decoding refuses a field it does not name, so a field added by
// hand is an error here rather than something the rewrite silently drops.
type market struct {
	Name  string `json:"name"`
	Owner struct {
		Name string `json:"name"`
	} `json:"owner"`
	Description string  `json:"description"`
	Plugins     []entry `json:"plugins"`
}

func main() {
	check := flag.Bool("check", false, "verify the manifest instead of rewriting it")
	flag.Parse()

	root, err := gitx.Root()
	if err != nil {
		fmt.Fprintln(os.Stderr, "marketplacegen:", err)
		os.Exit(1)
	}
	path := filepath.Join(root, manifest)
	have, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, "marketplacegen:", err)
		os.Exit(1)
	}
	want, err := render(root, have)
	if err != nil {
		fmt.Fprintln(os.Stderr, "marketplacegen:", err)
		os.Exit(1)
	}
	if *check {
		if !bytes.Equal(have, want) {
			fmt.Fprintf(os.Stderr, "marketplacegen: %s does not pin each plugin to the release its plugin.json names — run `go -C scripts run ./marketplacegen` (a version bump regenerates the ref)\n", manifest)
			os.Exit(1)
		}
		fmt.Printf("marketplacegen: %s pins every plugin with binaries to its release tag\n", manifest)
		return
	}
	if err := os.WriteFile(path, want, 0o644); err != nil {
		fmt.Fprintln(os.Stderr, "marketplacegen:", err)
		os.Exit(1)
	}
	fmt.Printf("marketplacegen: wrote %s\n", manifest)
}

// render returns the manifest with every plugin's source set from its tree.
func render(root string, data []byte) ([]byte, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	var m market
	if err := dec.Decode(&m); err != nil {
		return nil, fmt.Errorf("%s: %w", manifest, err)
	}
	if len(m.Plugins) == 0 {
		return nil, fmt.Errorf("%s lists no plugins — refusing to generate nothing", manifest)
	}
	for i, p := range m.Plugins {
		var src any = "./plugins/" + p.Name
		if st, err := os.Stat(filepath.Join(root, "plugins", p.Name, "tools", "cmd")); err == nil && st.IsDir() {
			version, err := pluginVersion(root, p.Name)
			if err != nil {
				return nil, err
			}
			src = gitSubdir{Source: "git-subdir", URL: repoURL, Path: "plugins/" + p.Name, Ref: p.Name + "--v" + version}
		}
		raw, err := json.Marshal(src)
		if err != nil {
			return nil, err
		}
		m.Plugins[i].Source = raw
	}
	var b bytes.Buffer
	enc := json.NewEncoder(&b)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(m); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func pluginVersion(root, name string) (string, error) {
	p := filepath.Join(root, "plugins", name, ".claude-plugin", "plugin.json")
	b, err := os.ReadFile(p)
	if err != nil {
		return "", err
	}
	var pj struct {
		Version string `json:"version"`
	}
	if err := json.Unmarshal(b, &pj); err != nil {
		return "", fmt.Errorf("%s: %w", p, err)
	}
	if pj.Version == "" {
		return "", fmt.Errorf("%s has no version, so plugin %s has no release tag to pin to", p, name)
	}
	return pj.Version, nil
}
