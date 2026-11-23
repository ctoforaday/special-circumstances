package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The headroom is the smallest of MemAvailable and what each cgroup ancestor leaves under its
// memory.high and memory.max; "max" limits nothing, and a level with no memory files is passed.
func TestReadHeadroomTakesTheTightestLimit(t *testing.T) {
	const gib = 1 << 30
	write := func(t *testing.T, path, body string) {
		t.Helper()
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	type level struct{ high, max, current string }
	for _, c := range []struct {
		name        string
		avail       string // MemAvailable in kB, "" for no meminfo
		levels      map[string]level
		wantBytes   uint64
		wantBinding string
		wantOK      bool
	}{
		{"no limits: MemAvailable", "8388608", map[string]level{"user.slice/a.service": {"max", "max", "1000"}},
			8 * gib, "MemAvailable", true},
		{"memory.high binds", "8388608", map[string]level{"user.slice/a.service": {"6442450944", "8589934592", "6000000000"}},
			6442450944 - 6000000000, "cgroup user.slice/a.service memory.high", true},
		{"memory.max of an ancestor binds", "8388608", map[string]level{
			"user.slice/a.service": {"max", "max", "1000"},
			"user.slice":           {"max", "3221225472", "2147483648"},
		}, gib, "cgroup user.slice memory.max", true},
		{"over the limit reads zero", "8388608", map[string]level{"a.service": {"1000", "max", "5000"}},
			0, "cgroup a.service memory.high", true},
		{"nothing readable", "", nil, 0, "", false},
	} {
		t.Run(c.name, func(t *testing.T) {
			root := t.TempDir()
			proc, cg := filepath.Join(root, "proc"), filepath.Join(root, "cgroup")
			if c.avail != "" {
				write(t, filepath.Join(proc, "meminfo"), "MemTotal: 16000000 kB\nMemAvailable: "+c.avail+" kB\n")
			}
			deepest := ""
			for rel, l := range c.levels {
				if len(rel) > len(deepest) {
					deepest = rel
				}
				write(t, filepath.Join(cg, rel, "memory.high"), l.high+"\n")
				write(t, filepath.Join(cg, rel, "memory.max"), l.max+"\n")
				write(t, filepath.Join(cg, rel, "memory.current"), l.current+"\n")
			}
			if deepest != "" {
				write(t, filepath.Join(proc, "self", "cgroup"), "0::/"+deepest+"\n")
			}
			got := readHeadroom(proc, cg)
			if got.ok != c.wantOK || got.bytes != c.wantBytes || got.binding != c.wantBinding {
				t.Errorf("readHeadroom = %+v, want {bytes:%d binding:%q ok:%v}", got, c.wantBytes, c.wantBinding, c.wantOK)
			}
		})
	}
}

// A gate that declares its memory does not start short of it, and names the limit; one that
// declares none, or a headroom nothing could read, starts as before.
func TestMemoryShort(t *testing.T) {
	g := gate{id: "deadcode", memory: 4 << 30}
	short := headroom{bytes: 1 << 30, binding: "cgroup a.service memory.high", ok: true}
	if why := memoryShort(g, short); !strings.Contains(why, "not started") || !strings.Contains(why, "memory.high leaves 1024 MB") {
		t.Errorf("memoryShort = %q, want a refusal naming the limit", why)
	}
	for name, c := range map[string]struct {
		g gate
		h headroom
	}{
		"enough":           {g, headroom{bytes: 5 << 30, binding: "MemAvailable", ok: true}},
		"no declared need": {gate{id: "fmt"}, short},
		"unreadable":       {g, headroom{}},
	} {
		if why := memoryShort(c.g, c.h); why != "" {
			t.Errorf("%s: memoryShort = %q, want it to start", name, why)
		}
	}
}

// deadcode is the gate measured at 3.26 GiB; deleting its declaration must fail here.
func TestDeadcodeDeclaresItsMemory(t *testing.T) {
	for _, g := range gateSet() {
		if g.id == "deadcode" {
			if g.memory < 3<<30 {
				t.Errorf("deadcode declares %d bytes; measured peak is 3.26 GiB", g.memory)
			}
			return
		}
	}
	t.Fatal("no deadcode gate")
}

// runGate asks before it starts: a gate declaring more than any machine has is refused without
// running. Deleting the check from runGate fails this.
func TestRunGateRefusesAGateShortOfMemory(t *testing.T) {
	if !readHeadroom("/proc", "/sys/fs/cgroup").ok {
		t.Skip("no memory reading on this system: a declared gate runs as before")
	}
	g := gate{id: "huge", kind: kindFmt, dir: ".", memory: 1 << 62}
	r := runGate(t.TempDir(), g, "")
	if r.status != "FAIL" || !strings.Contains(r.details, "not started") {
		t.Errorf("runGate = %s %q, want FAIL not started", r.status, r.details)
	}
}
