package installshape

// THE INSTALL-SHAPE CONTRACT.
//
// "Where does this session's plugin live" has been derived in a dozen places, each tested against
// a layout its own author built, and it flipped for one install shape or another a dozen times
// (feov-slim/hook-observation/ARCHAEOLOGY.md). This test puts the client's layout in the loop:
// it installs the suite through the real clients, lets a no-login session hand each hook its
// CLAUDE_PLUGIN_ROOT, and checks every reader against THAT root, never against a path it computed:
//
//	(i)   every hooks.json command of every plugin runs its binary from that root's bin/;
//	(ii)  after SessionStart that bin/ holds this PR's release bytes, and no other plugin tree
//	      changed;
//	(iii) sc-doctor, run as the doctor skill runs it, reads that root and its siblings;
//	(iv)  feov-record setup's binDir is that root's bin/;
//	(v)   the hook stamp a registration would carry comes from that root's PreToolUse hook.
//
// Shapes: a git-subdir source from a file:// repository (with an upgrade leg that leaves two
// versions cached), a directory marketplace, --plugin-dir, and `agy plugin install`. Every HOME
// also holds a decoy: the whole suite at a newer version in the HOST cache, $HOME/.claude, while
// CLAUDE_CONFIG_DIR points elsewhere.
//
// Antigravity cannot run a session without OAuth, so for agy the root is the installed `path`
// `agy plugin list --json` reports, and the root hooks.json commands are driven with that path as
// their working directory: the `${PWD}` derivation they rely on is ASSUMED there, not observed.

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/ctoforaday/special-circumstances/scripts/internal/gitx"
)

const (
	marketplaceName = "special-circumstances"
	oldVersion      = "0.0.1"  // the upgrade leg's first install
	decoyVersion    = "99.0.0" // newer than any real version, so a "newest cached" reader picks it
)

// suite is what every shape shares: the clients, the checkout, its plugins and the release.
type suite struct {
	repo    string
	plugins []plugin
	rel     *release
	trace   string // tripwires append here: one line per stand-in binary that ran
	claude  client
	agy     client
	pinned  pins
}

func TestInstallShapeContract(t *testing.T) {
	if testing.Short() {
		t.Skip("-short: the install-shape contract drives the real clients")
	}
	if why := supported(); why != "" {
		t.Skip(why)
	}
	refuseCredentialEnv(t)
	s := &suite{pinned: readPins(t)}
	s.claude = resolveClient(t, "claude")
	s.agy = resolveClient(t, "agy")
	s.checkPins(t)

	repo, err := gitx.Root()
	if err != nil {
		t.Fatal(err)
	}
	s.repo = repo
	work := t.TempDir()
	s.trace = filepath.Join(work, "trace.log")
	probeStage := filepath.Join(work, "plugins-as-checked-out")
	stageTree(t, repo, probeStage)
	s.plugins = stagedPlugins(t, probeStage)
	for _, p := range s.plugins {
		if p.hasBinaries() && len(p.cmds) == 0 {
			t.Fatalf("%s ships helpers but no tools/cmd binary", p.name)
		}
	}
	start := time.Now()
	s.rel = buildRelease(t, repo, filepath.Join(work, "release"), s.plugins, oldVersion, s.trace)
	t.Logf("built the local release (%s, build stamp %s) in %s", s.rel.dir, s.rel.revision, time.Since(start).Round(time.Second))

	t.Run("git-subdir", s.gitShape)
	t.Run("directory", s.directoryShape)
	t.Run("plugin-dir", s.pluginDirShape)
	t.Run("agy", s.agyShape)
}

// checkPins holds the clients to clients.json. INSTALLSHAPE_CLIENTS=latest is the scheduled run on
// whatever the installers serve today; it asserts nothing about versions and names them instead.
// Off CI a mismatch is logged rather than failed, so a developer's own clients still run it.
func (s *suite) checkPins(t *testing.T) {
	t.Helper()
	t.Logf("clients: claude %s (%s), agy %s (%s); pinned claude %s, agy %s", s.claude.version, s.claude.path, s.agy.version, s.agy.path, s.pinned.Claude.Version, s.pinned.Agy.Version)
	if os.Getenv("INSTALLSHAPE_CLIENTS") == "latest" {
		return
	}
	var off []string
	if s.claude.version != s.pinned.Claude.Version {
		off = append(off, fmt.Sprintf("claude %s (pinned %s)", s.claude.version, s.pinned.Claude.Version))
	}
	if s.agy.version != s.pinned.Agy.Version {
		off = append(off, fmt.Sprintf("agy %s (pinned %s)", s.agy.version, s.pinned.Agy.Version))
	}
	if len(off) == 0 {
		return
	}
	if os.Getenv("CI") == "true" {
		t.Fatalf("the clients are not the pinned ones: %s. The job installs scripts/installshape/clients.json's versions.", strings.Join(off, ", "))
	}
	t.Logf("NOT THE PINNED CLIENTS: %s — this run measures them, not the pins", strings.Join(off, ", "))
}

// failPrefix names the clients on every failure of the scheduled latest-client run, so a client
// change reads as one.
func (s *suite) label(shape string) string {
	if os.Getenv("INSTALLSHAPE_CLIENTS") == "latest" {
		return fmt.Sprintf("%s [claude %s, agy %s]", shape, s.claude.version, s.agy.version)
	}
	return shape
}

// shapeRun is one shape's install, observed.
type shapeRun struct {
	s        *suite
	t        *testing.T
	b        *box
	label    string
	base     string // every tree this shape made lives under it
	stage    string // the staged tree the client installs from
	rootsLog string
	probes   map[string]bool
	agy      bool
	session  func() // a no-login Claude session; nil for agy
	roots    map[string]string
}

func (s *suite) newShape(t *testing.T, name string, agy bool) *shapeRun {
	base := t.TempDir()
	r := &shapeRun{s: s, t: t, label: s.label(name), base: base, agy: agy,
		stage: filepath.Join(base, "stage"), rootsLog: filepath.Join(base, "roots.log")}
	r.b = newBox(t, r.label, base, s.rel.url, s.claude, s.agy)
	stageTree(t, s.repo, r.stage)
	if !agy {
		r.probes = injectProbes(t, r.stage, s.plugins, r.rootsLog)
	}
	plantDecoy(t, r.stage, r.b.home, s.plugins, decoyVersion, s.trace)
	if !agy {
		r.b.assertLoggedOut()
	}
	return r
}

// errorf fails the shape with its label.
func (r *shapeRun) errorf(format string, a ...any) {
	r.t.Helper()
	r.t.Errorf("%s: %s", r.label, fmt.Sprintf(format, a...))
}

func (r *shapeRun) fatalf(format string, a ...any) {
	r.t.Helper()
	r.t.Fatalf("%s: %s", r.label, fmt.Sprintf(format, a...))
}

// --- the shapes ---

// gitShape: plugin sources are git-subdir entries over a file:// repository built from this tree,
// tagged per plugin.json. The client refuses `marketplace add file://…`, so the marketplace itself
// is a directory — a clone of that repository, as the client's own GitHub marketplaces are — and
// what makes this the GitHub shape is the plugin source, which installs into the versioned cache.
// The upgrade leg installs every plugin at 0.0.1 (a tripwire release), fills it, then updates.
func (s *suite) gitShape(t *testing.T) {
	r := s.newShape(t, "git-subdir", false)
	repo := filepath.Join(r.base, "repo")
	copyTree(t, r.stage, repo)
	git(t, repo, "init", "-q")
	for _, p := range s.plugins {
		setVersion(t, filepath.Join(repo, "plugins", p.name), oldVersion)
	}
	git(t, repo, "add", "-A")
	git(t, repo, "commit", "-qm", "old")
	for _, p := range s.plugins {
		git(t, repo, "tag", p.tag(oldVersion))
	}
	for _, p := range s.plugins {
		setVersion(t, filepath.Join(repo, "plugins", p.name), p.version)
	}
	git(t, repo, "add", "-A")
	git(t, repo, "commit", "-qm", "new")
	for _, p := range s.plugins {
		git(t, repo, "tag", p.tag(p.version))
	}
	mkt := filepath.Join(r.base, "marketplace")
	git(t, r.base, "clone", "-q", "file://"+repo, mkt)
	pin := func(version func(plugin) string) {
		byName := map[string]plugin{}
		for _, p := range s.plugins {
			byName[p.name] = p
		}
		setSources(t, mkt, func(name string) any {
			return map[string]any{"source": "git-subdir", "url": "file://" + repo, "path": "plugins/" + name, "ref": byName[name].tag(version(byName[name]))}
		})
	}
	pin(func(plugin) string { return oldVersion })
	r.b.must(s.claude, "plugin", "marketplace", "add", mkt)
	for _, p := range s.plugins {
		r.b.must(s.claude, "plugin", "install", p.name+"@"+marketplaceName)
	}
	r.session = func() { r.b.session() }

	// The old install: observed and filled from the 0.0.1 tripwire release.
	r.fire()
	oldRoots := r.roots
	for _, p := range s.plugins {
		if p.hasBinaries() {
			if !r.waitFilled(p, oldRoots[p.name], p.tag(oldVersion), time.Now().Add(2*time.Minute)) {
				t.FailNow()
			}
			removeOne(t, oldRoots[p.name], p)
		}
	}

	pin(func(p plugin) string { return p.version })
	r.b.must(s.claude, "plugin", "marketplace", "update", marketplaceName)
	for _, p := range s.plugins {
		r.b.must(s.claude, "plugin", "update", p.name+"@"+marketplaceName)
	}
	upgraded := time.Now()
	r.check()
	r.t.Logf("%s: upgrade leg checked in %s", r.label, time.Since(upgraded).Round(time.Second))
	for name, old := range oldRoots {
		if _, err := os.Stat(old); err != nil {
			r.errorf("the upgrade leg did not leave two versions cached: %s's %s root %s is gone (%v), so this run cannot tell the loaded version from a neighbour", name, oldVersion, old, err)
		} else if old == r.roots[name] {
			r.errorf("%s still loads from its %s root %s after the update", name, oldVersion, old)
		}
	}
}

// directoryShape: a directory marketplace, which the client runs in place.
func (s *suite) directoryShape(t *testing.T) {
	r := s.newShape(t, "directory", false)
	setSources(t, r.stage, func(name string) any { return "./plugins/" + name })
	r.b.must(s.claude, "plugin", "marketplace", "add", r.stage)
	for _, p := range s.plugins {
		r.b.must(s.claude, "plugin", "install", p.name+"@"+marketplaceName)
	}
	r.session = func() { r.b.session() }
	r.check()
}

// pluginDirShape: no install at all; each plugin directory is passed to the session.
func (s *suite) pluginDirShape(t *testing.T) {
	r := s.newShape(t, "plugin-dir", false)
	var args []string
	for _, p := range s.plugins {
		args = append(args, "--plugin-dir", filepath.Join(r.stage, "plugins", p.name))
	}
	r.session = func() { r.b.session(args...) }
	r.check()
}

// agyShape: `agy plugin install` of the staged marketplace.
func (s *suite) agyShape(t *testing.T) {
	r := s.newShape(t, "agy", true)
	setSources(t, r.stage, func(name string) any { return "./plugins/" + name })
	r.b.must(s.agy, "plugin", "install", r.stage)
	var listed []struct {
		Name string `json:"name"`
		Path string `json:"path"`
	}
	out := r.b.must(s.agy, "plugin", "list", "--json")
	if err := json.Unmarshal([]byte(out), &listed); err != nil {
		r.fatalf("agy plugin list --json: %v\n%s", err, out)
	}
	r.roots = map[string]string{}
	for _, l := range listed {
		r.roots[l.Name] = l.Path
	}
	r.check()
}

// --- the checks every shape runs ---

// check fires the session that fills bin/ and runs (ii), (iii), (iv), (v) and then (i).
func (r *shapeRun) check() {
	r.t.Helper()
	before := snapshot(r.t, r.base)
	r.fire()
	filled, deadline := true, time.Now().Add(2*time.Minute)
	for _, p := range r.s.plugins {
		if p.hasBinaries() {
			filled = r.waitFilled(p, r.roots[p.name], p.tag(p.version), deadline) && filled
		}
	}
	// Before giving up on an unfilled root, say which tree the session filled instead.
	r.unchangedElsewhere(before, snapshot(r.t, r.base))
	if !filled {
		r.t.FailNow()
	}
	for _, p := range r.s.plugins {
		if p.hasBinaries() {
			r.releaseBytes(p)
		}
	}
	r.doctor()
	project, session := r.setup()
	r.hookStamp(project, session)
	r.hooksResolve(project)
}

// fire runs the session that hands each hook its root, and records the roots. For agy it drives
// each plugin's root hooks.json SessionStart commands with the installed path as working directory.
func (r *shapeRun) fire() {
	r.t.Helper()
	if r.agy {
		for _, p := range r.s.plugins {
			root := r.roots[p.name]
			r.underBase(p.name, root)
			if _, err := os.Stat(filepath.Join(root, "hooks.json")); err != nil {
				continue // a plugin with no hooks
			}
			for _, h := range hookCommands(r.t, filepath.Join(root, "hooks.json"), true) {
				if h.event != "SessionStart" {
					continue
				}
				res := r.b.exec(root, r.b.env(), agyPayload("SessionStart", r.b.work), time.Minute, "/bin/sh", "-c", h.command)
				if res.code != 0 {
					r.errorf("%s's SessionStart command exited %d at %s: %s%s", p.name, res.code, root, res.stdout, res.stderr)
				}
			}
		}
		r.t.Logf("%s: roots are the installed paths agy reports (hook cwd ASSUMED, not observed): %s", r.label, r.describeRoots())
		return
	}
	if err := os.WriteFile(r.rootsLog, nil, 0o644); err != nil {
		r.t.Fatal(err)
	}
	r.session()
	r.roots = r.readRoots()
	r.t.Logf("%s: the client handed the hooks %s", r.label, r.describeRoots())
}

// readRoots reads the root probe's lines: one root per plugin, the same for every event that fired.
func (r *shapeRun) readRoots() map[string]string {
	r.t.Helper()
	b, err := os.ReadFile(r.rootsLog)
	if err != nil {
		r.fatalf("the root probe wrote nothing (%v): no SessionStart hook of any plugin fired", err)
	}
	roots := map[string]string{}
	events := map[string][]string{}
	for _, line := range strings.Split(strings.TrimSpace(string(b)), "\n") {
		f := strings.Split(line, "\t")
		if len(f) != 3 {
			r.fatalf("root probe line %q is not plugin<TAB>event<TAB>root", line)
		}
		name, ev, root := f[0], f[1], f[2]
		if root == "" {
			r.errorf("%s's %s hook was handed no CLAUDE_PLUGIN_ROOT", name, ev)
			continue
		}
		if prev, ok := roots[name]; ok && prev != root {
			r.errorf("%s's hooks were handed two roots in one session: %s and %s (%s)", name, prev, root, ev)
		}
		roots[name] = root
		events[name] = append(events[name], ev)
	}
	for _, p := range r.s.plugins {
		if _, err := os.Stat(filepath.Join(r.stage, "plugins", p.name, ".claude-plugin", "hooks.json")); err != nil {
			continue
		}
		if !contains(events[p.name], "SessionStart") {
			r.fatalf("%s's SessionStart hook did not fire (events seen: %v)", p.name, events[p.name])
		}
		r.underBase(p.name, roots[p.name])
	}
	return roots
}

func (r *shapeRun) describeRoots() string {
	var parts []string
	for _, p := range r.s.plugins {
		if root, ok := r.roots[p.name]; ok {
			rel, _ := filepath.Rel(r.base, root)
			parts = append(parts, p.name+"="+rel)
		}
	}
	return strings.Join(parts, " ")
}

// underBase refuses a root outside this shape's temp tree: a client that installed into the real
// ~/.claude or ~/.gemini would have escaped the isolation.
func (r *shapeRun) underBase(name, root string) {
	r.t.Helper()
	rel, err := filepath.Rel(r.base, root)
	if root == "" || err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		r.fatalf("%s's root %q is outside the isolated tree %s", name, root, r.base)
	}
}

// waitFilled waits for fetch-bin's detached fetch to finish in root: the stamp names tag and the
// lock is gone, or the fetch recorded a failure. It reports false, with the error already recorded, when neither came.
func (r *shapeRun) waitFilled(p plugin, root, tag string, deadline time.Time) bool {
	r.t.Helper()
	for {
		b, err := os.ReadFile(filepath.Join(root, ".fetch", "installed"))
		if err == nil && strings.TrimSpace(string(b)) == tag {
			// The fetch has finished only when it has released its lock: it stamps first and
			// exits after, and whatever it does in between must land before the snapshot.
			if _, err := os.Stat(filepath.Join(root, ".fetch", "lock")); os.IsNotExist(err) {
				return true
			}
		}
		if f, err := os.ReadFile(filepath.Join(root, ".fetch", "failed")); err == nil {
			log, _ := os.ReadFile(filepath.Join(root, ".fetch", "log"))
			r.errorf("(ii) fetch-bin failed to fill %s's root %s with %s: %s\n%s", p.name, root, tag, f, log)
			return false
		}
		if time.Now().After(deadline) {
			log, _ := os.ReadFile(filepath.Join(root, ".fetch", "log"))
			r.errorf("(ii) after SessionStart, %s's loaded root %s was never filled with %s (stamp %q): the session's self-heal filled another tree, or none\nfetch log:\n%s", p.name, root, tag, strings.TrimSpace(string(b)), log)
			return false
		}
		time.Sleep(200 * time.Millisecond)
	}
}

// unchangedElsewhere is (ii)'s second half: the session changed no plugin tree but the roots it
// loaded.
func (r *shapeRun) unchangedElsewhere(before, after map[string]string) {
	r.t.Helper()
	loaded := map[string]bool{}
	for _, root := range r.roots {
		loaded[root] = true
	}
	var trees []string
	for tree := range after {
		trees = append(trees, tree)
	}
	sort.Strings(trees)
	for _, tree := range trees {
		if loaded[tree] {
			continue
		}
		was, existed := before[tree]
		switch {
		case !existed:
			r.errorf("(ii) the session created plugin tree %s, which no hook was handed (loaded: %s)", tree, r.describeRoots())
		case was != after[tree]:
			r.errorf("(ii) the session changed plugin tree %s, which no hook was handed (loaded: %s)", tree, r.describeRoots())
		}
	}
}

// releaseBytes is (ii): the loaded root's bin/ is exactly this PR's release for this host.
func (r *shapeRun) releaseBytes(p plugin) {
	r.t.Helper()
	root := r.roots[p.name]
	tag := p.tag(p.version)
	entries, err := os.ReadDir(filepath.Join(root, "bin"))
	if err != nil {
		r.errorf("(ii) %s's root %s has no bin/: %v", p.name, root, err)
		return
	}
	seen := map[string]bool{}
	for _, e := range entries {
		if e.Name() == ".gitkeep" {
			continue
		}
		seen[e.Name()] = true
		if _, ok := p.installed[e.Name()]; !ok {
			r.errorf("(ii) %s's root %s holds bin/%s, which the release does not ship", p.name, root, e.Name())
			continue
		}
		if got, want := fileSum(r.t, filepath.Join(root, "bin", e.Name())), r.s.rel.assetSum(p, tag, e.Name()); got != want {
			r.errorf("(ii) %s's root %s: bin/%s is not this PR's release asset %s/%s", p.name, root, e.Name(), tag, p.installed[e.Name()])
		}
	}
	for name := range p.installed {
		if !seen[name] {
			r.errorf("(ii) %s's root %s lacks bin/%s after SessionStart", p.name, root, name)
		}
	}
}

// doctor is (iii): sc-doctor, run from the loaded prosthetic-conscience root as the doctor skill
// runs it — no CLAUDE_PLUGIN_ROOT in its environment — lists every binary of every plugin as
// present. Every tree it could read instead lacks one: the host-cache decoy and the upgrade
// leg's old version had one removed, and a directory source's cache copy has an empty bin/.
func (r *shapeRun) doctor() {
	r.t.Helper()
	root, ok := r.roots["prosthetic-conscience"]
	if !ok {
		r.errorf("(iii) no prosthetic-conscience root was observed, so sc-doctor cannot be run from it")
		return
	}
	res := r.b.exec(r.b.work, r.b.env(), "", 2*time.Minute, filepath.Join(root, "bin", "sc-doctor"))
	if res.code != 0 {
		r.errorf("(iii) sc-doctor exited %d\n%s%s", res.code, res.stdout, res.stderr)
		return
	}
	rows := map[string]string{}
	for _, line := range strings.Split(res.stdout, "\n") {
		f := strings.Fields(line)
		if len(f) >= 2 {
			rows[f[0]] = strings.Join(f[1:], " ")
		}
	}
	for _, p := range r.s.plugins {
		var names []string
		for n := range p.installed {
			names = append(names, n)
		}
		sort.Strings(names)
		for _, n := range names {
			mark, ok := rows[n]
			switch {
			case !ok:
				r.errorf("(iii) sc-doctor from %s lists no row for %s's %s: it did not read %s's root %s\n%s", root, p.name, n, p.name, r.roots[p.name], res.stdout)
			case !strings.HasPrefix(mark, "✓") && !strings.HasPrefix(mark, "!"):
				r.errorf("(iii) sc-doctor from %s reports %s's %s as %q, but the loaded root %s holds it: the doctor read another tree\n%s", root, p.name, n, mark, r.roots[p.name], res.stdout)
			}
		}
	}
}

// setup is (iv): feov-record setup, run from the loaded FEOV root's bin/, prints that bin/ as the
// workflow's binDir. It returns the project and session the run was set up in, for (v).
func (r *shapeRun) setup() (project, session string) {
	r.t.Helper()
	root, ok := r.roots["frank-exchange-of-views"]
	if !ok {
		r.fatalf("(iv) no frank-exchange-of-views root was observed")
	}
	project = filepath.Join(r.base, "project")
	mkdir(r.t, project)
	git(r.t, project, "init", "-q")
	git(r.t, project, "commit", "-q", "--allow-empty", "-m", "project")
	session = "installshape-" + strings.ReplaceAll(r.t.Name(), "/", "-")
	res := r.b.exec(project, r.b.env("CLAUDE_CODE_SESSION_ID="+session), "", 2*time.Minute,
		filepath.Join(root, "bin", "feov-record"), "setup", filepath.Join(project, "research", "installshape"), "--seat-id", "operator",
		"--topic", "installshape", "--bulk", "haiku", "--judgment", "haiku", "--skip-model-check")
	if res.code != 0 {
		r.fatalf("(iv) feov-record setup from %s refused (exit %d)\n%s%s", root, res.code, res.stdout, res.stderr)
	}
	lines := strings.Split(res.stdout, "\n")
	binDir := ""
	for i, l := range lines {
		if strings.Contains(l, "binDir arg:") && i+1 < len(lines) {
			binDir = strings.TrimSpace(lines[i+1])
		}
	}
	if binDir == "" {
		r.fatalf("(iv) feov-record setup printed no binDir line\n%s", res.stdout)
	}
	if want := filepath.Join(root, "bin"); !samePath(binDir, want) {
		r.errorf("(iv) feov-record setup's binDir is %s, not the loaded root's %s", binDir, want)
	}
	return project, session
}

// hookStamp is (v): the PreToolUse hook command, driven with the loaded root in a live run, stamps
// a seat's Bash call with the build of the release this PR ships. The tool stamp is the build of
// the feov-record that (iv) found in that root, whose bytes (ii) already compared.
//
// The registration itself is not driven: `register` needs an agent the client attributed to the
// run through SubagentStart, which only a logged-in session produces.
func (r *shapeRun) hookStamp(project, session string) {
	r.t.Helper()
	if r.agy {
		// Antigravity's PreToolUse payload and output are its own protocol; the stamp is
		// injected by the same binary, whose bytes (ii) compared.
		return
	}
	root := r.roots["frank-exchange-of-views"]
	var cmd string
	for _, h := range hookCommands(r.t, filepath.Join(root, ".claude-plugin", "hooks.json"), false) {
		if h.event == "PreToolUse" && !r.probes[h.command] {
			cmd = h.command
			break
		}
	}
	if cmd == "" {
		r.fatalf("(v) frank-exchange-of-views has no PreToolUse command")
	}
	payload, _ := json.Marshal(map[string]any{
		"session_id": session, "cwd": project, "hook_event_name": "PreToolUse", "tool_name": "Bash",
		"tool_input": map[string]any{"command": "printenv FEOV_HOOK_VERSION"},
		"agent_id":   "installshape-agent", "agent_type": "installshape",
	})
	res := r.b.exec(project, r.b.env("CLAUDE_PLUGIN_ROOT="+root), string(payload), time.Minute, "/bin/sh", "-c", cmd)
	var out struct {
		HookSpecificOutput struct {
			UpdatedInput struct {
				Command string `json:"command"`
			} `json:"updatedInput"`
		} `json:"hookSpecificOutput"`
	}
	if err := json.Unmarshal([]byte(res.stdout), &out); err != nil || out.HookSpecificOutput.UpdatedInput.Command == "" {
		r.errorf("(v) the PreToolUse hook at %s rewrote nothing in the live run (exit %d): %s%s", root, res.code, res.stdout, res.stderr)
		return
	}
	stamp := r.b.exec(project, r.b.env(), "", time.Minute, "/bin/sh", "-c", out.HookSpecificOutput.UpdatedInput.Command)
	if got := strings.TrimSpace(stamp.stdout); got != r.s.rel.revision {
		r.errorf("(v) the hook stamp from %s is %q, not this PR's release build %q", root, got, r.s.rel.revision)
	}
}

// hooksResolve is (i). Every root's bin/ and fetch-bin.sh become tripwires; then, for Claude, a
// session fires its hooks once more, and every hooks.json command of every plugin is driven with
// the root the client handed over. Each command must run exactly one file, and that file must lie
// in its own plugin's loaded root. A guard that falls through to `fetch-bin.sh hook` did not find
// its binary, and fails as well.
func (r *shapeRun) hooksResolve(project string) {
	r.t.Helper()
	for _, p := range r.s.plugins {
		if p.hasBinaries() {
			fillTripwires(r.t, r.roots[p.name], p, r.s.trace)
		}
	}
	if !r.agy {
		r.truncateTrace()
		r.session()
		ran := r.readTrace()
		if len(ran) == 0 {
			r.errorf("(i) the client fired no hook command that ran a binary or fetch-bin.sh")
		}
		r.t.Logf("%s: (i) the client fired %d hook commands that ran a file, all checked against the loaded roots", r.label, len(ran))
		for _, line := range ran {
			if !r.inSomeRoot(line.path) {
				r.errorf("(i) a hook the client fired ran %s, which is in no loaded root (%s)", line.path, r.describeRoots())
			}
		}
	}
	driven := 0
	defer func() { r.t.Logf("%s: (i) drove %d hooks.json commands against the loaded roots", r.label, driven) }()
	for _, p := range r.s.plugins {
		root := r.roots[p.name]
		file := filepath.Join(root, ".claude-plugin", "hooks.json")
		if r.agy {
			file = filepath.Join(root, "hooks.json")
		}
		if _, err := os.Stat(file); err != nil {
			continue
		}
		for _, h := range hookCommands(r.t, file, r.agy) {
			if r.probes[h.command] {
				continue
			}
			driven++
			r.truncateTrace()
			env, dir, payload := r.b.env("CLAUDE_PLUGIN_ROOT="+root), project, claudePayload(h.event, project)
			if r.agy {
				env, dir, payload = r.b.env(), root, agyPayload(h.event, project)
			}
			r.b.exec(dir, env, payload, time.Minute, "/bin/sh", "-c", h.command)
			ran := r.readTrace()
			switch {
			case len(ran) != 1:
				r.errorf("(i) %s %s ran %d files from root %s, want exactly one: %v", p.name, h.event, len(ran), root, ran)
			case !within(root, ran[0].path):
				r.errorf("(i) %s %s ran %s, outside its loaded root %s", p.name, h.event, ran[0].path, root)
			case strings.HasPrefix(ran[0].args, "hook "):
				r.errorf("(i) %s %s fell through to %s %s: its binary is not under %s/bin", p.name, h.event, ran[0].path, ran[0].args, root)
			}
		}
	}
}

func (r *shapeRun) inSomeRoot(path string) bool {
	for _, root := range r.roots {
		if within(root, path) {
			return true
		}
	}
	return false
}

func (r *shapeRun) truncateTrace() {
	r.t.Helper()
	if err := os.WriteFile(r.s.trace, nil, 0o644); err != nil {
		r.t.Fatal(err)
	}
}

type ranFile struct{ path, args string }

func (r *shapeRun) readTrace() []ranFile {
	r.t.Helper()
	b, err := os.ReadFile(r.s.trace)
	if err != nil {
		r.t.Fatal(err)
	}
	var out []ranFile
	for _, line := range strings.Split(strings.TrimSpace(string(b)), "\n") {
		if line == "" {
			continue
		}
		path, args, _ := strings.Cut(line, "\t")
		out = append(out, ranFile{path: path, args: args})
	}
	return out
}

// hookCmd is one hooks.json command and the event it is registered on.
type hookCmd struct{ event, command string }

// hookCommands reads every command of a hooks.json: Claude's {"hooks": {Event: [{hooks: [...]}]}}
// or Antigravity's {"<plugin>-hooks": {Event: [{command} | {matcher, hooks: [...]}]}}.
func hookCommands(t *testing.T, path string, agy bool) []hookCmd {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]map[string][]struct {
		Command string `json:"command"`
		Hooks   []struct {
			Command string `json:"command"`
		} `json:"hooks"`
	}
	if !agy {
		var claude struct {
			Hooks map[string][]struct {
				Hooks []struct {
					Command string `json:"command"`
				} `json:"hooks"`
			} `json:"hooks"`
		}
		if err := json.Unmarshal(b, &claude); err != nil {
			t.Fatalf("%s: %v", path, err)
		}
		var out []hookCmd
		for ev, groups := range claude.Hooks {
			for _, g := range groups {
				for _, h := range g.Hooks {
					out = append(out, hookCmd{ev, h.Command})
				}
			}
		}
		sortCmds(out)
		return out
	}
	if err := json.Unmarshal(b, &doc); err != nil {
		t.Fatalf("%s: %v", path, err)
	}
	var out []hookCmd
	for _, events := range doc {
		for ev, items := range events {
			for _, it := range items {
				if it.Command != "" {
					out = append(out, hookCmd{ev, it.Command})
				}
				for _, h := range it.Hooks {
					out = append(out, hookCmd{ev, h.Command})
				}
			}
		}
	}
	sortCmds(out)
	if len(out) == 0 {
		t.Fatalf("%s holds no hook command", path)
	}
	return out
}

func sortCmds(c []hookCmd) {
	sort.Slice(c, func(i, j int) bool {
		if c[i].event != c[j].event {
			return c[i].event < c[j].event
		}
		return c[i].command < c[j].command
	})
}

func claudePayload(event, cwd string) string {
	b, _ := json.Marshal(map[string]any{
		"session_id": "installshape-drive", "cwd": cwd, "hook_event_name": event,
		"tool_name": "Bash", "tool_input": map[string]any{"command": "true"},
	})
	return string(b)
}

func agyPayload(event, cwd string) string {
	b, _ := json.Marshal(map[string]any{
		"conversationId": "installshape-drive", "workspacePaths": []string{cwd}, "hookEventName": event,
	})
	return string(b)
}

// within reports whether path lies under root.
func within(root, path string) bool {
	rel, err := filepath.Rel(root, path)
	return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

func samePath(a, b string) bool {
	ra, errA := filepath.EvalSymlinks(a)
	rb, errB := filepath.EvalSymlinks(b)
	return errA == nil && errB == nil && ra == rb
}

// git runs git in dir with a fixed identity and no signing, so the host's config cannot change it.
func git(t *testing.T, dir string, args ...string) {
	t.Helper()
	full := append([]string{"-c", "user.name=installshape", "-c", "user.email=installshape@invalid", "-c", "commit.gpgsign=false", "-c", "tag.gpgsign=false", "-c", "init.defaultBranch=main"}, args...)
	cmd := exec.Command("git", full...)
	cmd.Dir = dir
	if b, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %s in %s: %v\n%s", strings.Join(args, " "), dir, err, b)
	}
}
