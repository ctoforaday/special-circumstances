#!/usr/bin/env bash
# run.sh <label=binary[,label=binary...]> <models dir: gte-small.ort, gte-small.int8.ort, vocab.txt> <data dir> <out dir> <repo root>
# Per binary: linkage (staticproof + the platform's own lister) and fidelity on speed.jsonl (4 threads) for float32
# and int8 against ref_gte.npy. Then one-thread speed on len200.jsonl, 3 rounds, interleaved across binaries and
# models on this one host. Never stops on a failed step.
set -uo pipefail
LIST=$1; M=$(cd "$2" && pwd); D=$(cd "$3" && pwd); O=$(mkdir -p "$4" && cd "$4" && pwd); REPO=$(cd "$5" && pwd)
export VOCAB=$M/vocab.txt
IFS=, read -r -a ARMS <<< "$LIST"
{
  echo "== host"; uname -a
  (lscpu 2>/dev/null || { sysctl -n machdep.cpu.brand_string hw.ncpu hw.memsize; } 2>/dev/null || true) | head -30
  [ "$(uname -s)" = Darwin ] || true
  case "$(uname -s)" in MINGW*|MSYS*) (powershell -NoProfile -Command "Get-CimInstance Win32_Processor | Select-Object Name,NumberOfCores,NumberOfLogicalProcessors | Format-List" 2>/dev/null || true) ;; esac
} > "$O/host.txt" 2>&1
cat "$O/host.txt"
for arm in "${ARMS[@]}"; do
  L=${arm%%=*}; BIN=${arm#*=}
  [ -f "$BIN" ] || { echo "$L: no binary at $BIN" | tee -a "$O/missing.txt"; continue; }
  chmod +x "$BIN" 2>/dev/null || true
  {
    echo "== $L file"; file "$BIN" 2>&1 || true
    echo "== $L size"; wc -c < "$BIN"
    echo "== $L staticproof"; go -C "$REPO/plugins/frank-exchange-of-views/tools" run ./third_party/pins/staticproof "$BIN"; echo "staticproof exit=$?"
    echo "== $L platform dependency listing"
    case "$(uname -s)" in
      Linux) ldd "$BIN"; echo "ldd exit=$?" ;;
      Darwin) otool -L "$BIN"; echo "otool exit=$?"; codesign -dv "$BIN" 2>&1 | head -3 ;;
      *) if command -v objdump >/dev/null; then objdump -p "$BIN" | grep -i "DLL Name"; echo "objdump exit=$?"; fi ;;
    esac
  } > "$O/linkage-$L.txt" 2>&1
  cat "$O/linkage-$L.txt"
  for m in gte-small gte-small.int8; do
    { "$BIN" "$M/$m.ort" "$D/speed.jsonl" "$O/speed-$L-$m.bin" 4; echo "run exit=$?"
      "$BIN" cos "$O/speed-$L-$m.bin" "$D/ref_gte.npy"; echo "cos exit=$?"; } > "$O/fid-$L-$m.txt" 2>&1
    echo "== fidelity $L $m"; cat "$O/fid-$L-$m.txt"
  done
done
for r in 1 2 3; do for arm in "${ARMS[@]}"; do L=${arm%%=*}; BIN=${arm#*=}; [ -f "$BIN" ] || continue
  for m in gte-small gte-small.int8; do
    echo "round=$r arm=$L model=$m $("$BIN" "$M/$m.ort" "$D/len200.jsonl" "$O/len200.bin" 1 2>&1 | tail -1)" | tee -a "$O/speed.txt"
  done; done; done
rm -f "$O"/*.bin
