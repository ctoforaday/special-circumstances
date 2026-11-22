"""Frank Exchange of Views — Antigravity run_workflow orchestrator.

Mirrors debate.js's roundless, record-driven debate engine using Antigravity's
native run_workflow primitives (agent, parallel, phase, log).
"""

import json
import os
import re
import subprocess
import uuid

a = json.loads(args) if isinstance(args, str) else (args or {})
if "maxRounds" in a or "max_rounds" in a:
    raise RuntimeError(
        "debate: refusing dispatch — maxRounds is not a term of this engine any more "
        "(plans/roundless.md). A run ends on the record, not on a count: the chair's "
        "`dispatch next` says who sits, and the run ends when nobody is ready — PASS "
        "permitted — or at the run's epoch limit (CEILING). Drop maxRounds from args."
    )

topic = a.get("topic")
run_dir = a.get("runDir") or a.get("run_dir")
lanes = int(a.get("lanes", 3))
seats_cfg = a.get("seats")
lane_floor_override = a.get("laneFloorOverride") or a.get("lane_floor_override")
bin_dir = a.get("binDir") or a.get("bin_dir")

if not topic or not run_dir or "undefined" in str(run_dir) or str(topic) == "undefined":
    raise RuntimeError(
        f"debate: refusing dispatch — topic/runDir unbound (topic={topic!r}, runDir={run_dir!r})"
    )
if not seats_cfg or not isinstance(seats_cfg, dict):
    raise RuntimeError(
        "debate: refusing dispatch — seats unset. The engine does not guess a tier or inherit "
        "the session model: pass seats keyed by seat id ({ harness: 'antigravity', model })."
    )
if not bin_dir:
    raise RuntimeError(
        "debate: refusing dispatch — binDir unset. The engine does not run without the record."
    )
if lanes < 3 and not lane_floor_override:
    raise RuntimeError(
        f"debate: lanes={lanes} is below the floor of 3 — pass laneFloorOverride: '<reason>' "
        "to run under-provisioned deliberately"
    )

# BEGIN GENERATED SEATS — go -C plugins/frank-exchange-of-views/tools run ./devcmd/enginegen writes this from the record's roster. DO NOT EDIT BY HAND.
SEATS = json.loads("{\"blue-respond\":{\"agentType\":\"blue-researcher\",\"role\":\"blue\",\"tier\":\"bulk\"},\"blue-synthesize\":{\"agentType\":\"blue-synthesizer\",\"role\":\"blue\",\"tier\":\"judgment\"},\"frontier\":{\"agentType\":\"blue-researcher\",\"role\":\"blue\",\"tier\":\"bulk\"},\"judge\":{\"agentType\":\"lead-judge\",\"role\":\"bench\",\"tier\":\"judgment\"},\"red-chair\":{\"agentType\":\"red-chair\",\"role\":\"chair\",\"tier\":\"judgment\"},\"red-lens-adversary\":{\"agentType\":\"red-lens-adversary\",\"area\":\"adversary\",\"role\":\"lens\",\"tier\":\"bulk\"},\"red-lens-architecture\":{\"agentType\":\"red-lens-architecture\",\"area\":\"architecture\",\"role\":\"lens\",\"tier\":\"bulk\"},\"red-lens-computation\":{\"agentType\":\"red-lens-computation\",\"area\":\"computation\",\"role\":\"lens\",\"tier\":\"bulk\"},\"red-lens-dark-side\":{\"agentType\":\"red-lens-dark-side\",\"area\":\"dark-side\",\"role\":\"lens\",\"tier\":\"bulk\"},\"red-lens-evidence\":{\"agentType\":\"red-lens-evidence\",\"area\":\"evidence\",\"role\":\"lens\",\"tier\":\"bulk\"},\"red-lens-logic\":{\"agentType\":\"red-lens-logic\",\"area\":\"logic\",\"role\":\"lens\",\"tier\":\"bulk\"},\"red-lens-voice\":{\"agentType\":\"red-lens-voice\",\"area\":\"voice\",\"role\":\"lens\",\"tier\":\"bulk\"}}")
LANE = json.loads("{\"agentType\":\"blue-lane\",\"prefix\":\"blue-lane-\",\"role\":\"lane\",\"tier\":\"bulk\"}")
OCCASION = json.loads("{\"assemble\":\"assemble\",\"docket\":\"docket\",\"petition\":\"petition\",\"position\":\"position\",\"terminal\":\"terminal\"}")
MASS_MAPPING_VERSION = json.loads("\"v2\"")
# END GENERATED SEATS

# BEGIN GENERATED ENVELOPE VOCABULARY — go -C plugins/frank-exchange-of-views/tools run ./devcmd/enginegen writes this from the record's enums, record.Plan and record.LensEnvelope. DO NOT EDIT BY HAND.
GRADE = json.loads("{\"enum\":[\"trivial\",\"low\",\"low_medium\",\"medium\",\"medium_high\",\"high\",\"certain\",\"realized\"],\"type\":\"string\"}")
DISPUTE_DIMENSION = json.loads("{\"enum\":[\"severity\",\"likelihood\",\"impact\",\"complexity\"],\"type\":\"string\"}")
RUN_OUTCOME = json.loads("{\"enum\":[\"VERIFIED\",\"CEILING\",\"HALTED\",\"UNVERIFIED\",\"QUALIFIED\"],\"type\":\"string\"}")
BENCH_HEARINGS = json.loads("[\"petition\",\"position\"]")
GAP_ID = json.loads("{\"pattern\":\"^(?:G-)[0-9a-f]{8}$\",\"type\":\"string\"}")
STANDING_CLAUSES = json.loads("[{\"clause\":\"unchecked\",\"kind\":\"unchecked\"},{\"clause\":\"on a failed check\",\"kind\":\"failed_check\"},{\"clause\":\"on a reading debt\",\"kind\":\"reading_debt\"},{\"clause\":\"inquiries open\",\"kind\":\"open_inquiry\"},{\"clause\":\"inquiries unchecked\",\"kind\":\"unchecked_inquiry\"},{\"clause\":\"the question with no part split out\",\"kind\":\"question_unsplit\"},{\"clause\":\"question parts with no coverage\",\"kind\":\"open_part\"},{\"clause\":\"question parts unchecked\",\"kind\":\"unchecked_part\"}]")
BENCH_PARTY = json.loads("{\"additionalProperties\":false,\"properties\":{\"gap_ids\":{\"items\":{\"type\":\"string\"},\"minItems\":1,\"type\":\"array\"},\"seat_id\":{\"enum\":[\"judge\"],\"type\":\"string\"}},\"required\":[\"seat_id\",\"gap_ids\"],\"type\":\"object\"}")
LENS_OR_BLUE_PARTY = json.loads("{\"additionalProperties\":false,\"properties\":{\"gap_ids\":{\"items\":{\"type\":\"string\"},\"type\":\"array\"},\"seat_id\":{\"enum\":[\"blue-respond\",\"red-lens-adversary\",\"red-lens-architecture\",\"red-lens-computation\",\"red-lens-dark-side\",\"red-lens-evidence\",\"red-lens-logic\",\"red-lens-voice\"],\"type\":\"string\"}},\"required\":[\"seat_id\",\"gap_ids\"],\"type\":\"object\"}")
PLAN = json.loads("{\"additionalProperties\":false,\"properties\":{\"blockers\":{\"items\":{\"additionalProperties\":false,\"properties\":{\"kind\":{\"enum\":[\"material_gap\",\"remand_owed\",\"unchecked\",\"failed_check\",\"reading_debt\",\"no_report\",\"lens_ready\",\"stale_area\",\"grade_movement\",\"closing_owed\",\"unruled_motion\",\"upheld_position\",\"open_inquiry\",\"unchecked_inquiry\",\"question_unsplit\",\"open_part\",\"unchecked_part\"],\"type\":\"string\"},\"owner\":{\"type\":\"string\"},\"subject\":{\"description\":\"the reference that names it: a gap id, a motion id, a lens seat, an anchor; empty for the run-wide kinds\",\"type\":\"string\"}},\"required\":[\"kind\",\"subject\",\"owner\"],\"type\":\"object\"},\"type\":\"array\"},\"ceiling\":{\"type\":\"boolean\"},\"docket\":{\"items\":{\"type\":\"string\"},\"type\":\"array\"},\"epoch\":{\"description\":\"the epoch the record has reached: the chair sittings opened so far, this one included\",\"type\":\"integer\"},\"epoch_limit_reached\":{\"description\":\"the run's epoch limit leaves undispatched a party the board readies — any, at the sitting that opens the run's last epoch; blue, the only one, at the sitting before it — so nobody is dispatched, and ceiling is set for that reason\",\"type\":\"boolean\"},\"exchanges\":{\"items\":{\"additionalProperties\":false,\"properties\":{\"count\":{\"type\":\"integer\"},\"gap_id\":{\"type\":\"string\"},\"unresolved\":{\"type\":\"integer\"}},\"required\":[\"gap_id\",\"count\",\"unresolved\"],\"type\":\"object\"},\"type\":\"array\"},\"final_sitting\":{\"description\":\"this chair sitting opens the run's last epoch: no seat is dispatched after it, and of the motions filed now only a petition is heard, by the bench at the exit\",\"type\":\"boolean\"},\"head\":{\"description\":\"events.id of the report head the parties audit\",\"type\":\"integer\"},\"key\":{\"description\":\"the plan's identity: the same in two chair sittings whose plans decide the same — everything here but epoch, final_sitting and passed\",\"type\":\"string\"},\"max_epochs\":{\"description\":\"the run's epoch limit, a term setup records; 0 when the run is held to none\",\"type\":\"integer\"},\"parties\":{\"items\":{\"oneOf\":[{\"additionalProperties\":false,\"properties\":{\"gap_ids\":{\"items\":{\"type\":\"string\"},\"minItems\":1,\"type\":\"array\"},\"seat_id\":{\"enum\":[\"judge\"],\"type\":\"string\"}},\"required\":[\"seat_id\",\"gap_ids\"],\"type\":\"object\"},{\"additionalProperties\":false,\"properties\":{\"gap_ids\":{\"items\":{\"type\":\"string\"},\"type\":\"array\"},\"seat_id\":{\"enum\":[\"blue-respond\",\"red-lens-adversary\",\"red-lens-architecture\",\"red-lens-computation\",\"red-lens-dark-side\",\"red-lens-evidence\",\"red-lens-logic\",\"red-lens-voice\"],\"type\":\"string\"}},\"required\":[\"seat_id\",\"gap_ids\"],\"type\":\"object\"}]},\"type\":\"array\"},\"pass_permitted\":{\"type\":\"boolean\"},\"passed\":{\"description\":\"true when the chair recorded the PASS in this sitting; absent while it has not\",\"type\":\"boolean\"},\"remand_owed\":{\"items\":{\"additionalProperties\":false,\"properties\":{\"direction\":{\"description\":\"never blank: the record refuses a remand without one\",\"type\":\"string\"},\"gap_id\":{\"type\":\"string\"}},\"required\":[\"gap_id\",\"direction\"],\"type\":\"object\"},\"type\":\"array\"},\"stale_areas\":{\"items\":{\"additionalProperties\":false,\"properties\":{\"pin\":{\"type\":\"integer\"},\"seat_id\":{\"type\":\"string\"}},\"required\":[\"seat_id\",\"pin\"],\"type\":\"object\"},\"type\":\"array\"},\"why\":{\"items\":{\"type\":\"string\"},\"type\":\"array\"}},\"required\":[\"key\",\"head\",\"parties\",\"docket\",\"remand_owed\",\"exchanges\",\"pass_permitted\",\"stale_areas\",\"ceiling\",\"max_epochs\",\"epoch\",\"final_sitting\",\"epoch_limit_reached\",\"blockers\",\"why\"],\"type\":\"object\"}")
PLAN_FIELD_TYPES = json.loads("{\"blockers\":\"array\",\"ceiling\":\"boolean\",\"docket\":\"array\",\"epoch\":\"integer\",\"epoch_limit_reached\":\"boolean\",\"exchanges\":\"array\",\"final_sitting\":\"boolean\",\"head\":\"integer\",\"key\":\"string\",\"max_epochs\":\"integer\",\"parties\":\"array\",\"pass_permitted\":\"boolean\",\"remand_owed\":\"array\",\"stale_areas\":\"array\",\"why\":\"array\"}")
PLAN_EMPTY_ONLY_WITH = json.loads("{\"blockers\":\"pass_permitted\"}")
PLAN_ITEM_TYPES = json.loads("{\"blockers\":{\"kind\":\"string\",\"owner\":\"string\",\"subject\":\"string\"},\"exchanges\":{\"count\":\"integer\",\"gap_id\":\"string\",\"unresolved\":\"integer\"},\"parties\":{\"gap_ids\":\"array\",\"seat_id\":\"string\"},\"remand_owed\":{\"direction\":\"string\",\"gap_id\":\"string\"},\"stale_areas\":{\"pin\":\"integer\",\"seat_id\":\"string\"}}")
RECORD_READ = json.loads("{\"additionalProperties\":false,\"properties\":{\"bench_owed\":{\"items\":{\"type\":\"string\"},\"type\":\"array\"},\"event\":{\"description\":\"the events.id the record was read at\",\"type\":\"integer\"},\"halted\":{\"description\":\"the bench has halted the run\",\"type\":\"boolean\"},\"outcome\":{\"description\":\"the outcome stamped on the record; absent while none is\",\"enum\":[\"VERIFIED\",\"CEILING\",\"HALTED\",\"UNVERIFIED\",\"QUALIFIED\"],\"type\":\"string\"}},\"required\":[\"event\",\"bench_owed\",\"halted\"],\"type\":\"object\"}")
LENS_ENVELOPE = json.loads("{\"additionalProperties\":false,\"properties\":{\"record_read\":{\"additionalProperties\":false,\"properties\":{\"bench_owed\":{\"items\":{\"type\":\"string\"},\"type\":\"array\"},\"event\":{\"description\":\"the events.id the record was read at\",\"type\":\"integer\"},\"halted\":{\"description\":\"the bench has halted the run\",\"type\":\"boolean\"},\"outcome\":{\"description\":\"the outcome stamped on the record; absent while none is\",\"enum\":[\"VERIFIED\",\"CEILING\",\"HALTED\",\"UNVERIFIED\",\"QUALIFIED\"],\"type\":\"string\"}},\"required\":[\"event\",\"bench_owed\",\"halted\"],\"type\":\"object\"}},\"required\":[],\"type\":\"object\"}")
# END GENERATED ENVELOPE VOCABULARY

EMPTY_ENVELOPE = {
    "type": "object",
    "additionalProperties": False,
    "properties": {},
}
CHAIR_ENVELOPE = {
    "type": "object",
    "additionalProperties": False,
    "description": "The orchestrator reads the authoritative plan from the record after your sitting. Return it empty, or holding only notes.",
    "properties": {"notes": {"type": "string"}},
}
BLUE_ENVELOPE = {
    "type": "object",
    "additionalProperties": False,
    "properties": {"notes": {"type": "string"}},
}
BENCH_ENVELOPE = {
    "type": "object",
    "additionalProperties": False,
    "properties": {
        "halt_opinion": {"type": "string"},
        "notes": {"type": "string"},
    },
}
LENS_FINISH_ENVELOPE = {
    "type": "object",
    "additionalProperties": False,
    "properties": {},
}

LANE_METHODS = [
    "adversarial-disconfirming-first (hunt evidence AGAINST the frontier hypotheses before evidence for them)",
    "primary-literature (papers, specs, standards — leaf sources over commentary)",
    "local-repo critical-stance (audit the subject artifacts/codebase directly; trust nothing secondhand)",
    "practitioner-production (issue trackers, postmortems, field reports — how it fails in practice)",
    "adversarial-disconfirming-first, second seat (redundancy floor — this method must not have single-point coverage)",
]

NO_PROGRESS_EPOCHS = 3
ENVELOPE_REFUSAL_RETRIES = 5
slug = os.path.basename(os.path.normpath(run_dir)) or "run"
feov_record_bin = os.path.join(bin_dir, "feov-record.exe" if os.name == "nt" else "feov-record")
feov_sitting_write_bin = os.path.join(bin_dir, "feov-sitting-write.exe" if os.name == "nt" else "feov-sitting-write")
agy_sittings_dir = os.path.join(run_dir, ".agy-sittings")
lead_session = globals().get("CONVERSATION_ID") or os.environ.get("ANTIGRAVITY_CONVERSATION_ID") or ""


def _run_sitting_write(args_list, stdin_text=None):
    return subprocess.run(
        [feov_sitting_write_bin, *args_list],
        input=stdin_text,
        text=True,
        capture_output=True,
        check=False,
    )


def role_of_seat(seat_id):
    if seat_id in SEATS:
        return SEATS[seat_id]["role"]
    if seat_id.startswith(LANE["prefix"]):
        return LANE["role"]
    raise RuntimeError(f"debate: seat_id {seat_id!r} is not on the roster")


def dispatch_spec(seat_id):
    if seat_id in SEATS:
        info = SEATS[seat_id]
        cfg_key = seat_id
    elif seat_id.startswith(LANE["prefix"]):
        info = LANE
        cfg_key = "blue-lane"
    else:
        raise RuntimeError(f"debate: seat_id {seat_id!r} is not on the roster")
    cfg = seats_cfg.get(cfg_key)
    if not isinstance(cfg, dict) or not cfg.get("model"):
        raise RuntimeError(
            f"debate: no model configured for seat {seat_id!r} (key {cfg_key!r}) in args.seats"
        )
    return info["agentType"], cfg["model"]


def record_clause(seat_id):
    return (
        f" Run `{feov_record_bin} <verb> --run {run_dir} --seat-id {seat_id}` via `run_command` "
        f"(check `--help` first if unsure of flags). Pass the exact `--seat-id {seat_id}` on every call."
    )


halted = False
halt_opinion = None
read_stop = None


def read_of(seat_id, env):
    global halted, halt_opinion, read_stop
    if not isinstance(env, dict):
        return None
    rr = env.get("record_read")
    if not isinstance(rr, dict) or not isinstance(rr.get("event"), int) or rr["event"] < 1:
        read_stop = {"seat": seat_id}
        log(f"stopping: {seat_id} returned an envelope without a hook-written record_read")
        return None
    if rr.get("halted") is True:
        halted = True
        if env.get("halt_opinion"):
            halt_opinion = env["halt_opinion"]
    return rr


def require_plan(plan, epoch):
    if not isinstance(plan, dict):
        raise RuntimeError(f"red-chair sitting {epoch} returned non-object plan")
    for field, t in PLAN_FIELD_TYPES.items():
        val = plan.get(field)
        ok = (
            isinstance(val, int) and not isinstance(val, bool)
            if t == "integer"
            else isinstance(val, bool)
            if t == "boolean"
            else isinstance(val, str)
            if t == "string"
            else isinstance(val, list)
        )
        if not ok:
            raise RuntimeError(
                f"red-chair sitting {epoch} relayed a plan whose `{field}` is {val!r}, not {t}"
            )
    for lst, types in PLAN_ITEM_TYPES.items():
        for i, item in enumerate(plan[lst]):
            if not isinstance(item, dict):
                raise RuntimeError(f"red-chair sitting {epoch} relayed invalid {lst}[{i}]")
            for f, t in types.items():
                v = item.get(f)
                ok = (
                    isinstance(v, list)
                    if t == "array"
                    else isinstance(v, str)
                    if t == "string"
                    else isinstance(v, bool)
                    if t == "boolean"
                    else isinstance(v, int) and not isinstance(v, bool)
                )
                if not ok:
                    raise RuntimeError(
                        f"red-chair sitting {epoch} relayed `{lst}[{i}].{f}` = {v!r}, not {t}"
                    )
    for p in plan["parties"]:
        if role_of_seat(p["seat_id"]) == "bench" and not p["gap_ids"]:
            raise RuntimeError(
                f"red-chair sitting {epoch} relayed a bench party with no gap"
            )
    for i, o in enumerate(plan["remand_owed"]):
        if not o["direction"].strip():
            raise RuntimeError(
                f"red-chair sitting {epoch} relayed blank remand_owed[{i}].direction"
            )
    for lst, flag in PLAN_EMPTY_ONLY_WITH.items():
        if not plan[lst] and plan[flag] is not True:
            raise RuntimeError(
                f"red-chair sitting {epoch} relayed empty `{lst}` with `{flag}`={plan[flag]!r}"
            )
    if "passed" in plan and not isinstance(plan["passed"], bool):
        raise RuntimeError(
            f"red-chair sitting {epoch} relayed a plan whose `passed` is {plan['passed']!r}, not a boolean or absent"
        )


async def sit(seat_id, prompt, schema, role_title):
    type_name, model = dispatch_spec(seat_id)
    agent_id = f"agy-{uuid.uuid4().hex[:16]}"
    agent_type = f"frank-exchange-of-views:{type_name}"
    os.makedirs(agy_sittings_dir, exist_ok=True)
    seat_reg_path = os.path.join(agy_sittings_dir, f"seat-{seat_id}.json")
    seat_tr_path = os.path.join(agy_sittings_dir, f"seat-{seat_id}.transcript")
    with open(seat_reg_path, "w") as f:
        json.dump(
            {
                "agent_id": agent_id,
                "agent_type": agent_type,
                "seat_id": seat_id,
                "run_dir": run_dir,
                "lead_session": lead_session,
            },
            f,
        )

    open_res = _run_sitting_write(
        ["-run", run_dir, "-phase", "open", "-agent-id", agent_id, "-agent-type", agent_type]
    )
    if open_res.returncode != 0:
        log(f"{seat_id} sitting open failed: {open_res.stderr.strip()}")
    work_list = (open_res.stdout or "").strip()
    full_prompt = f"{prompt}\n\n{work_list}" if work_list else prompt

    seat_agent = agent(role=role_title, type_name=type_name, model=model)
    try:
        env = await seat_agent(full_prompt, schema)
        for attempt in range(ENVELOPE_REFUSAL_RETRIES + 1):
            env_res = _run_sitting_write(
                ["-run", run_dir, "-phase", "envelope", "-agent-id", agent_id],
                stdin_text=json.dumps(env if isinstance(env, dict) else {}),
            )
            answer = {}
            if env_res.stdout and env_res.stdout.strip():
                try:
                    answer = json.loads(env_res.stdout)
                except Exception as exc:
                    log(f"{seat_id} envelope answer parse failed: {exc}")
                    return None
            if env_res.returncode != 0 and not answer.get("refusal"):
                log(f"{seat_id} envelope check failed: {env_res.stderr.strip()}")
                return None
            refusal = answer.get("refusal") or ""
            if not refusal:
                admit_res = _run_sitting_write(
                    ["-run", run_dir, "-phase", "admission", "-agent-id", agent_id]
                )
                if admit_res.returncode != 0:
                    log(f"{seat_id} admission failed: {admit_res.stderr.strip()}")
                out = dict(answer["envelope"]) if isinstance(answer.get("envelope"), dict) else (dict(env) if isinstance(env, dict) else {})
                if isinstance(answer.get("record_read"), dict):
                    out["record_read"] = answer["record_read"]
                return out
            if attempt >= ENVELOPE_REFUSAL_RETRIES:
                log(f"{seat_id} envelope refused {attempt + 1} times: {refusal}")
                return None
            log(f"{seat_id} envelope refused (retry {attempt + 1}): {refusal[:160]}")
            env = await seat_agent(
                f"Your envelope was refused: {refusal}\n\nResolve this on the record and call `finish` again.",
                schema,
            )
        return None
    except Exception as exc:
        log(f"{seat_id} sitting failed: {exc}")
        return None
    finally:
        transcript = ""
        if os.path.exists(seat_tr_path):
            try:
                transcript = open(seat_tr_path).read().strip()
            except OSError:
                transcript = ""
        close_args = ["-run", run_dir, "-phase", "close", "-agent-id", agent_id, "-agent-type", agent_type]
        if transcript:
            close_args.extend(["-transcript", transcript])
        close_res = _run_sitting_write(close_args)
        if close_res.returncode != 0:
            log(f"{seat_id} sitting close failed: {close_res.stderr.strip()}")
        for p in (seat_reg_path, seat_tr_path):
            try:
                os.remove(p)
            except OSError:
                pass
        del seat_agent


async def hear_owed(reads):
    global halted, read_stop
    owed = sorted({occ for rr in reads if rr for occ in (rr.get("bench_owed") or [])})
    for occ in owed:
        if halted or read_stop:
            break
        prompt = (
            f"Bench hearing ({occ}) for topic: \"{topic}\". Read the open {occ} filings on your "
            f"work list and rule on each on the record.{record_clause('judge')} Return the bench envelope."
        )
        env = await sit("judge", prompt, BENCH_ENVELOPE, f"judge · {occ}")
        read_of("judge", env)


async def run_debate():
    global halted, read_stop
    phase("Frontier")
    frontier_prompt = (
        f"Frontier reconnaissance for topic: \"{topic}\". Produce falsifiable hypotheses, "
        f"contested claims, and key sources to test.{record_clause('frontier')} Return the envelope."
    )
    f_env = await sit("frontier", frontier_prompt, EMPTY_ENVELOPE, f"frontier · {slug}")
    if f_env is not None:
        read_of("frontier", f_env)

    phase("Blue")

    async def run_lane(idx):
        lane_id = f"{LANE['prefix']}{idx}"
        method = LANE_METHODS[(idx - 1) % len(LANE_METHODS)]
        prompt = (
            f"Blue lane {idx} of {lanes} for topic: \"{topic}\". Method: {method}. "
            f"Research to saturation and record your part of the report.{record_clause(lane_id)} "
            "Return the envelope."
        )
        return lane_id, await sit(lane_id, prompt, EMPTY_ENVELOPE, f"{lane_id} · {slug}")

    lane_results = await parallel(list(range(1, lanes + 1)), run_lane)
    for lane_id, env in lane_results:
        if env is not None:
            read_of(lane_id, env)

    synth_prompt = (
        f"Blue synthesis for topic: \"{topic}\". Synthesize all lane parts into the report by union "
        f"through your acts on the record (include Title, TL;DR, The Catechism, sections, Open questions)."
        f"{record_clause('blue-synthesize')} Return the envelope."
    )
    synth_env = None
    for failures in range(1, NO_PROGRESS_EPOCHS + 1):
        synth_env = await sit("blue-synthesize", synth_prompt, EMPTY_ENVELOPE, f"blue-synthesize · {slug}")
        if synth_env is not None:
            break
        if failures >= NO_PROGRESS_EPOCHS:
            log(f"blue-synthesize failed {failures} sittings in a row — debate does not open")
            return {"verdict": "UNVERIFIED", "epochs": 0, "termination": "synthesis_failed", "lanes": lanes, "outstanding_gaps": 0}
    read_of("blue-synthesize", synth_env)

    phase("Debate")
    epoch = 0
    same_plan_key = None
    same_plan_epochs = 0
    last_plan = None

    while not halted and not read_stop:
        epoch += 1
        chair_prompt = (
            f"Red chair sitting (epoch {epoch}) for topic: \"{topic}\". Open the next epoch per your "
            f"constitution, record the PASS if permitted or withhold it with your position, and return the relayed plan in your envelope.{record_clause('red-chair')}"
        )
        plan = await sit("red-chair", chair_prompt, CHAIR_ENVELOPE, f"red-chair · ep{epoch}")
        if plan is None:
            log(f"red-chair sitting {epoch} returned no envelope — stopping UNVERIFIED")
            break
        require_plan(plan, epoch)
        last_plan = plan
        read_of("red-chair", plan)
        await hear_owed([plan.get("record_read")])
        if halted or read_stop:
            break
        if not plan["parties"]:
            break

        if plan["key"] == same_plan_key:
            same_plan_epochs += 1
            if same_plan_epochs >= NO_PROGRESS_EPOCHS:
                log(f"no-progress valve triggered after {same_plan_epochs} identical plans")
                break
        else:
            same_plan_key = plan["key"]
            same_plan_epochs = 1

        red_parties = [p for p in plan["parties"] if role_of_seat(p["seat_id"]) == "lens"]
        blue_parties = [p for p in plan["parties"] if role_of_seat(p["seat_id"]) == "blue"]
        bench_parties = [p for p in plan["parties"] if role_of_seat(p["seat_id"]) == "bench"]

        if red_parties:
            async def run_lens(party):
                sid = party["seat_id"]
                gaps = party.get("gap_ids") or []
                prompt = (
                    f"Red lens sitting ({SEATS[sid]['area']}, epoch {epoch}) for topic: \"{topic}\". "
                    f"Engaged gaps: {', '.join(gaps) if gaps else 'none'}.{record_clause(sid)} "
                    "Return the lens envelope."
                )
                return sid, await sit(sid, prompt, LENS_FINISH_ENVELOPE, f"{sid} · ep{epoch}")

            lens_envs = await parallel(red_parties, run_lens)
            reads = [read_of(sid, env) for sid, env in lens_envs if env is not None]
            await hear_owed(reads)
            if halted or read_stop:
                break

        for bp in blue_parties:
            sid = bp["seat_id"]
            gaps = bp.get("gap_ids") or []
            prompt = (
                f"Blue response sitting (epoch {epoch}) for topic: \"{topic}\". "
                f"Address gaps: {', '.join(gaps) if gaps else 'per work list'}.{record_clause(sid)} "
                "Return the blue envelope."
            )
            env = await sit(sid, prompt, BLUE_ENVELOPE, f"{sid} · ep{epoch}")
            rr = read_of(sid, env) if env is not None else None
            await hear_owed([rr])
            if halted or read_stop:
                break

        for jp in bench_parties:
            sid = jp["seat_id"]
            gaps = jp.get("gap_ids") or []
            prompt = (
                f"Bench impasse/ruling sitting (epoch {epoch}) for topic: \"{topic}\". "
                f"Docketed gaps: {', '.join(gaps)}.{record_clause(sid)} Return the bench envelope."
            )
            env = await sit(sid, prompt, BENCH_ENVELOPE, f"{sid} · ep{epoch}")
            if env is not None:
                read_of(sid, env)
            if halted or read_stop:
                break

    phase("Assemble")
    assemble_prompt = (
        f"Assemble final report for topic: \"{topic}\". Stamp the outcome and assemble all documents "
        f"on the record by union-copy.{record_clause('judge')} Return the envelope."
    )
    await sit("judge", assemble_prompt, EMPTY_ENVELOPE, f"assemble · {slug}")

    verdict = "HALTED" if halted else ("PASS" if last_plan and last_plan.get("passed") is True else None) or ("CEILING" if last_plan and last_plan.get("ceiling") else "UNVERIFIED")
    out = {
        "verdict": verdict,
        "epochs": epoch,
        "termination": "halted" if halted else ("pass_permitted" if last_plan and last_plan.get("pass_permitted") else ("ceiling" if last_plan and last_plan.get("ceiling") else "unverified")),
        "lanes": lanes,
        "outstanding_gaps": len(last_plan.get("exchanges", [])) if last_plan else 0,
    }
    if halt_opinion:
        out["halt_opinion"] = halt_opinion
    return out


result = await run_debate()
print(json.dumps(result))
