#!/usr/bin/env bash
# Runs every arm. Usage: run.sh <bench/embed dir> <work dir> <results dir>
# Phase 1: 3 interleaved rounds, every arm at ONE thread, on len200.jsonl.
# Phase 2: once at 4 threads/workers on len200.jsonl.
# Phase 3: fidelity of every arm on speed.jsonl against ref_gte.npy (embedded at 4 threads; threads do not change the math).
# A failed arm is recorded with its exit status and output; the script never stops on one.
set -uo pipefail
S=$(cd "$1" && pwd); W=$(cd "$2" && pwd); R=$3; mkdir -p "$R/out" "$R/runs"; R=$(cd "$R" && pwd)
LEN=$S/data/len200.jsonl; SPD=$S/data/speed.jsonl; HF=$W/models/gte-small; PY=$W/venv/bin/python
GM=$W/models/gte-small.gtemodel; F16=$W/models/gte-small-f16.gguf; Q8=$W/models/gte-small-q8_0.gguf
export OMP_NUM_THREADS=1 MKL_NUM_THREADS=1

{
  echo "== lscpu"; lscpu
  echo "== nproc"; nproc
  echo "== cpuinfo flags (avx512*, sve*, asimd*)"
  grep -m1 -E '^(flags|Features)' /proc/cpuinfo | tr ' ' '\n' | grep -E '^(avx512|sve|asimd)' | sort -u | tr '\n' ' '; echo
  echo "== model name"; grep -m1 -E '^model name' /proc/cpuinfo || echo "(none; see lscpu Model name)"
  echo "== uptime"; uptime
  echo "== openblas"; "$W/bin/blasinfo"
  if [ -x "$W/bin/simdwidth" ]; then echo "== go simd"; "$W/bin/simdwidth"; fi
} > "$R/host.txt" 2>&1
cat "$R/host.txt"

ARMS="gtego_asm gtego_blas llama_noblas_f16 llama_blas_f16 ort_f32 llama_noblas_q8 llama_blas_q8 ort_int8"
[ -x "$W/bin/gtego_simd" ] && ARMS="gtego_asm gtego_simd ${ARMS#gtego_asm }"

cmd() { # arm threads input output
  local a=$1 t=$2 i=$3 o=$4
  case $a in
    gtego_*) echo env GOMAXPROCS=$t OPENBLAS_NUM_THREADS=1 "$W/bin/$a" "$GM" "$i" "$o" "$t" ;;
    llama_noblas_f16) echo env OPENBLAS_NUM_THREADS=$t "$W/bin/harness_noblas" "$F16" "$i" "$o" "$t" 1 512 1 ;;
    llama_noblas_q8) echo env OPENBLAS_NUM_THREADS=$t "$W/bin/harness_noblas" "$Q8" "$i" "$o" "$t" 1 512 1 ;;
    llama_blas_f16) echo env OPENBLAS_NUM_THREADS=$t "$W/bin/harness_blas" "$F16" "$i" "$o" "$t" 1 512 1 ;;
    llama_blas_q8) echo env OPENBLAS_NUM_THREADS=$t "$W/bin/harness_blas" "$Q8" "$i" "$o" "$t" 1 512 1 ;;
    ort_f32|ort_int8) echo "$PY" -I "$S/ort/ort.py" "$HF" "$W/models" run $a "$t" "$i" "$o" ;;
  esac
}

run() { # phase round arm threads input output
  local ph=$1 r=$2 a=$3 t=$4 i=$5 o=$6 log="$R/runs/$1-r$2-$3-t$4.log" u0 st line
  u0=$(cut -d' ' -f1-3 /proc/loadavg)
  $(cmd "$a" "$t" "$i" "$o") > "$log" 2>&1; st=$?
  line=$(grep -E 'per_s=' "$log" | tail -1)
  [ -n "$line" ] || line="NO-RESULT: $(tail -3 "$log" | tr '\n' ' ')"
  printf '%s\t%s\t%s\t%s\t%s\t%s\t%s\n' "$ph" "$r" "$a" "$t" "$st" "$u0" "$line" | tee -a "$R/runs.tsv"
}

printf 'phase\tround\tarm\tthreads\texit\tload_before\tresult\n' > "$R/runs.tsv"
for r in 1 2 3; do for a in $ARMS; do run t1 $r $a 1 "$LEN" "$R/out/$a.len200.t1.bin"; done; done
for a in $ARMS; do run t4 1 $a 4 "$LEN" "$R/out/$a.len200.t4.bin"; done
for a in $ARMS; do run fid 1 $a 4 "$SPD" "$R/out/$a.speed.bin"; done

FA=(); for a in $ARMS; do FA+=("$a=$R/out/$a.speed.bin"); done
"$PY" -I "$S/fid.py" "$S/data/ref_gte.npy" "$R" "${FA[@]}" 2>&1 | tee "$R/fidelity.txt"
"$PY" -I "$S/summarize.py" "$R" 2>&1 | tee "$R/summary.md"
rm -f "$R"/out/*.len200.t*.bin
echo "== uptime end"; uptime | tee -a "$R/host.txt"
