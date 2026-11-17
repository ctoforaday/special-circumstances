package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/repotree"
)

func TestDebatePyCompilesAndSimulatesWorkflowPrimitives(t *testing.T) {
	if _, err := exec.LookPath("python3"); err != nil {
		t.Skip("python3 not installed")
	}
	pyPath, err := repotree.DebatePy()
	if err != nil {
		t.Fatal(err)
	}
	compileCheck := `import ast, sys
src = open(sys.argv[1]).read()
compile(src, sys.argv[1], "exec", flags=ast.PyCF_ALLOW_TOP_LEVEL_AWAIT)
`
	if out, err := exec.Command("python3", "-c", compileCheck, pyPath).CombinedOutput(); err != nil {
		t.Fatalf("top-level await compile failed: %v\n%s", err, out)
	}

	harness := `
import asyncio, json, sys

py_path = sys.argv[1]
with open(py_path, "r") as f:
    src = f.read()

# Strip top-level await run_debate() invocation so we can inject mock primitives and test scenarios.
body = src.split("result = await run_debate()")[0]

def make_env(scenario, args_dict):
    logs = []
    phases = []
    calls = []
    chair_turn = {"n": 0}

    class WorkflowError(Exception):
        pass

    async def agent(prompt=None, schema=None, role=None, type_name="self", workspace="inherit", model=None):
        calls.append({"type_name": type_name, "model": model, "role": role})
        rr = {"event": 10, "bench_owed": [], "halted": False}
        if scenario == "lens_error" and type_name.startswith("red-lens-"):
            raise WorkflowError("simulated seat timeout")
        if scenario == "missing_read" and type_name == "blue-synthesizer":
            return {}
        if type_name == "red-chair":
            chair_turn["n"] += 1
            n = chair_turn["n"]
            if scenario == "no_progress":
                return {
                    "key": "stuck-key", "head": 2,
                    "parties": [{"seat_id": "red-lens-logic", "gap_ids": []}],
                    "docket": [], "remand_owed": [], "exchanges": [],
                    "pass_permitted": False, "stale_areas": [],
                    "ceiling": False, "max_epochs": 12, "epoch": n,
                    "final_sitting": False, "epoch_limit_reached": False,
                    "blockers": [{"kind": "unchecked", "owner": "red-lens-logic", "subject": "report"}],
                    "why": ["open"], "record_read": rr,
                }
            if n == 1:
                return {
                    "key": "k1", "head": 2,
                    "parties": [
                        {"seat_id": "red-lens-logic", "gap_ids": []},
                        {"seat_id": "blue-respond", "gap_ids": []},
                    ],
                    "docket": [], "remand_owed": [], "exchanges": [],
                    "pass_permitted": False, "stale_areas": [],
                    "ceiling": False, "max_epochs": 12, "epoch": 1,
                    "final_sitting": False, "epoch_limit_reached": False,
                    "blockers": [{"kind": "unchecked", "owner": "red-lens-logic", "subject": "report"}],
                    "why": ["open"], "record_read": rr,
                }
            return {
                "key": "k2", "head": 2,
                "parties": [], "docket": [], "remand_owed": [], "exchanges": [],
                "pass_permitted": True, "stale_areas": [],
                "ceiling": False, "max_epochs": 12, "epoch": 2,
                "final_sitting": False, "epoch_limit_reached": False,
                "blockers": [], "why": [], "verdict": "PASS", "record_read": rr,
            }
        return {"record_read": rr}

    async def parallel(items, fn):
        return await asyncio.gather(*(fn(x) for x in items))

    def phase(name):
        phases.append(name)

    def log(*a):
        logs.append(" ".join(str(x) for x in a))

    ns = {"args": args_dict, "agent": agent, "parallel": parallel, "phase": phase, "log": log, "WorkflowError": WorkflowError}
    exec(compile(body, py_path, "exec"), ns)
    return ns, phases, calls, logs

base_seats = {k: {"harness": "antigravity", "model": "flash"} for k in [
    "frontier", "blue-lane", "blue-synthesize", "blue-respond", "red-chair", "judge", "assemble",
    "red-lens-adversary", "red-lens-architecture", "red-lens-computation",
    "red-lens-dark-side", "red-lens-evidence", "red-lens-logic", "red-lens-voice",
]}
base_args = {"topic": "test", "runDir": "/tmp/run-1", "lanes": 1, "laneFloorOverride": "smoke", "seats": base_seats, "binDir": "/tmp/bin"}

# 1. Refuses maxRounds
try:
    make_env("pass", {**base_args, "maxRounds": 3})
    raise AssertionError("expected maxRounds error")
except RuntimeError as e:
    assert "maxRounds" in str(e)

# 2. Full PASS flow
ns, phases, calls, _ = make_env("pass", base_args)
res = asyncio.run(ns["run_debate"]())
assert res["verdict"] == "PASS" and res["epochs"] == 2, res
assert phases == ["Frontier", "Blue", "Debate", "Assemble"], phases
assert all(":" not in c["type_name"] for c in calls), calls

# 3. WorkflowError on lens returns None without crashing engine
ns, _, _, logs = make_env("lens_error", base_args)
res = asyncio.run(ns["run_debate"]())
assert res["verdict"] == "PASS", res
assert any("simulated seat timeout" in l for l in logs), logs

# 4. No-progress valve stops after NO_PROGRESS_EPOCHS identical plans
ns, _, _, logs = make_env("no_progress", base_args)
res = asyncio.run(ns["run_debate"]())
assert res["verdict"] == "UNVERIFIED" and res["epochs"] == 3, res
assert any("no-progress valve" in l for l in logs), logs
print("ALL_PY_SIM_OK")
`
	harnessPath := filepath.Join(t.TempDir(), "sim.py")
	if err := os.WriteFile(harnessPath, []byte(harness), 0o644); err != nil {
		t.Fatal(err)
	}
	out, err := exec.Command("python3", harnessPath, pyPath).CombinedOutput()
	if err != nil {
		t.Fatalf("python simulator failed: %v\n%s", err, out)
	}
	if !strings.Contains(string(out), "ALL_PY_SIM_OK") {
		t.Fatalf("unexpected simulator output: %s", out)
	}
}
