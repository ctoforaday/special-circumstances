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
    sitting_writes = []
    chair_turn = {"n": 0}
    refusal_once = {"done": False}
    agent_types = {}

    class WorkflowError(Exception):
        pass

    class _FakeCompleted:
        def __init__(self, stdout="", stderr="", returncode=0):
            self.stdout = stdout
            self.stderr = stderr
            self.returncode = returncode

    def agent(prompt=None, schema=None, role=None, type_name="self", workspace="inherit", model=None):
        class _SeatAgent:
            async def __call__(self, turn_prompt, turn_schema=None):
                calls.append({"type_name": type_name, "model": model, "role": role, "prompt": turn_prompt})
                if scenario == "lens_error" and type_name.startswith("red-lens-"):
                    raise WorkflowError("simulated seat timeout")
                return {}
        return _SeatAgent()

    def fake_sitting_write(args_list, stdin_text=None):
        phase_name = args_list[args_list.index("-phase") + 1]
        agent_id = args_list[args_list.index("-agent-id") + 1]
        sitting_writes.append((phase_name, agent_id))
        if phase_name == "open":
            at = args_list[args_list.index("-agent-type") + 1]
            agent_types[agent_id] = at
            return _FakeCompleted(stdout="WORK LIST")
        if phase_name == "envelope":
            at = agent_types.get(agent_id, "")
            rr = {"event": 10, "bench_owed": [], "halted": False}
            if scenario == "missing_read" and at.endswith("blue-synthesizer"):
                return _FakeCompleted(stdout=json.dumps({}))
            if scenario == "refusal_retry" and at.endswith("red-lens-logic") and not refusal_once["done"]:
                refusal_once["done"] = True
                return _FakeCompleted(stdout=json.dumps({"refusal": "open blocker stands"}))
            if at.endswith("red-chair"):
                chair_turn["n"] += 1
                n = chair_turn["n"]
                if scenario == "no_progress":
                    plan = {
                        "key": "stuck-key", "head": 2,
                        "parties": [{"seat_id": "red-lens-logic", "gap_ids": []}],
                        "docket": [], "remand_owed": [], "exchanges": [],
                        "pass_permitted": False, "stale_areas": [],
                        "ceiling": False, "max_epochs": 12, "epoch": n,
                        "final_sitting": False, "epoch_limit_reached": False,
                        "blockers": [{"kind": "unchecked", "owner": "red-lens-logic", "subject": "report"}],
                        "why": ["open"],
                    }
                elif n == 1:
                    plan = {
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
                        "why": ["open"],
                    }
                else:
                    plan = {
                        "key": "k2", "head": 2,
                        "parties": [], "docket": [], "remand_owed": [], "exchanges": [],
                        "pass_permitted": True, "stale_areas": [],
                        "ceiling": False, "max_epochs": 12, "epoch": 2,
                        "final_sitting": False, "epoch_limit_reached": False,
                        "blockers": [], "why": [], "passed": True,
                    }
                return _FakeCompleted(stdout=json.dumps({"record_read": rr, "envelope": plan}))
            return _FakeCompleted(stdout=json.dumps({"record_read": rr}))
        return _FakeCompleted()

    async def parallel(items, fn):
        return await asyncio.gather(*(fn(x) for x in items))

    def phase(name):
        phases.append(name)

    def log(*a):
        logs.append(" ".join(str(x) for x in a))

    ns = {"args": args_dict, "agent": agent, "parallel": parallel, "phase": phase, "log": log, "WorkflowError": WorkflowError, "CONVERSATION_ID": "lead-conv-sim"}
    exec(compile(body, py_path, "exec"), ns)
    ns["_run_sitting_write"] = fake_sitting_write
    return ns, phases, calls, logs

import tempfile
tmp_run = tempfile.mkdtemp(prefix="debate-py-sim-")
base_seats = {k: {"harness": "antigravity", "model": "flash"} for k in [
    "frontier", "blue-lane", "blue-synthesize", "blue-respond", "red-chair", "judge", "assemble",
    "red-lens-adversary", "red-lens-architecture", "red-lens-computation",
    "red-lens-dark-side", "red-lens-evidence", "red-lens-logic", "red-lens-voice",
]}
base_args = {"topic": "test", "runDir": tmp_run, "lanes": 1, "laneFloorOverride": "smoke", "seats": base_seats, "binDir": "/tmp/bin"}

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

# 4. Envelope refusal retries on the same seat agent
ns, _, calls, logs = make_env("refusal_retry", base_args)
res = asyncio.run(ns["run_debate"]())
assert res["verdict"] == "PASS", res
assert any("envelope refused (retry 1)" in l for l in logs), logs
assert any("open blocker stands" in c["prompt"] for c in calls), calls

# 5. No-progress valve stops after NO_PROGRESS_EPOCHS identical plans
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
