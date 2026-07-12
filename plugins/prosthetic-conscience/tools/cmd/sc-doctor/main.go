// sc-doctor is the prosthetic-conscience environment preflight (Go binary).
//
// Contract (Design by Contract):
//   It MUST produce a deterministic table + verdict (READY / DEGRADED / BLOCKED).
//   Plain run is read-only. `-fix` rebuilds missing hook binaries (go build) when Go
//   is present, else prints the release-asset fetch instructions — it MUST NOT
//   install external tools (that stays consent-gated at the agent layer).
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/toolchain"
)

const version = "0.1.0"

type requirements struct {
	Tools []toolchain.Tool `json:"tools"`
}

type binStatus struct {
	Name  string
	Built bool
}

func exeSuffix() string {
	if runtime.GOOS == "windows" {
		return ".exe"
	}
	return ""
}

// hookBinaries lists every command under tools/cmd and whether bin/<name> exists.
func hookBinaries(root string) []binStatus {
	entries, err := os.ReadDir(filepath.Join(root, "tools", "cmd"))
	if err != nil {
		return nil
	}
	var out []binStatus
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		bin := filepath.Join(root, "bin", e.Name()+exeSuffix())
		_, err := os.Stat(bin)
		out = append(out, binStatus{Name: e.Name(), Built: err == nil})
	}
	return out
}

// verdict computes READY / DEGRADED / BLOCKED from tool + binary status.
func verdict(tools []toolchain.Status, bins []binStatus) string {
	degraded := false
	for _, t := range tools {
		if !t.Found {
			if t.Tier == "required" {
				return "BLOCKED"
			}
			if t.Tier == "recommended" {
				degraded = true
			}
		}
	}
	for _, b := range bins {
		if !b.Built {
			degraded = true
		}
	}
	if degraded {
		return "DEGRADED"
	}
	return "READY"
}

// table renders the preflight report.
func table(tools []toolchain.Status, bins []binStatus) string {
	var sb strings.Builder
	for _, t := range tools {
		if t.Found {
			fmt.Fprintf(&sb, "%-18s ✓ %s\n", t.Name, versionOf(t.CheckCmd))
		} else {
			fmt.Fprintf(&sb, "%-18s ✗ (%s) install: %s\n", t.Name, t.Tier, t.Install[runtime.GOOS])
		}
	}
	for _, b := range bins {
		mark := "✓ built"
		if !b.Built {
			mark = "✗ not built (run -fix)"
		}
		fmt.Fprintf(&sb, "%-18s %s\n", b.Name, mark)
	}
	return sb.String()
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

// fix builds every missing binary via go build; returns actions taken / instructions.
func fix(root string, bins []binStatus) []string {
	var report []string
	goPresent := toolchain.Present("go")
	for _, b := range bins {
		if b.Built {
			continue
		}
		if goPresent {
			out := filepath.Join(root, "bin", b.Name+exeSuffix())
			cmd := exec.Command("go", "build", "-C", filepath.Join(root, "tools"), "-o", out, "./cmd/"+b.Name)
			if msg, err := cmd.CombinedOutput(); err != nil {
				report = append(report, fmt.Sprintf("%s: BUILD FAILED: %s", b.Name, strings.TrimSpace(string(msg))))
			} else {
				report = append(report, fmt.Sprintf("%s: built", b.Name))
			}
		} else {
			report = append(report, fmt.Sprintf(
				"%s: Go not found — fetch release asset %s_%s_%s%s from the latest ctoforaday/special-circumstances release (verify SHA256SUMS) into %s",
				b.Name, b.Name, runtime.GOOS, runtime.GOARCH, exeSuffix(), filepath.Join(root, "bin")))
		}
	}
	return report
}

func main() {
	doFix := flag.Bool("fix", false, "build/fetch missing hook binaries")
	showVersion := flag.Bool("version", false, "print version and exit")
	rootFlag := flag.String("root", "", "plugin root (default: CLAUDE_PLUGIN_ROOT or the binary's parent dir)")
	flag.Parse()
	if *showVersion {
		fmt.Println("sc-doctor", version)
		return
	}

	root := *rootFlag
	if root == "" {
		root = os.Getenv("CLAUDE_PLUGIN_ROOT")
	}
	if root == "" {
		if exe, err := os.Executable(); err == nil {
			root = filepath.Dir(filepath.Dir(exe)) // bin/sc-doctor -> plugin root
		}
	}

	raw, err := os.ReadFile(filepath.Join(root, "requirements.json"))
	if err != nil {
		fmt.Printf("sc-doctor: cannot read requirements.json under %q: %v\n", root, err)
		os.Exit(0)
	}
	var req requirements
	if err := json.Unmarshal(raw, &req); err != nil {
		fmt.Printf("sc-doctor: malformed requirements.json: %v\n", err)
		os.Exit(0)
	}

	tools := toolchain.Probe(req.Tools)
	bins := hookBinaries(root)

	if *doFix {
		for _, line := range fix(root, bins) {
			fmt.Println(line)
		}
		bins = hookBinaries(root) // re-probe
	}

	fmt.Print(table(tools, bins))
	fmt.Println("VERDICT:", verdict(tools, bins))
}
