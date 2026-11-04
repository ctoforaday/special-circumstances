#!/usr/bin/env bash
# run.sh <probe binary> <models dir: gte-small.ort, gte-small.int8.ort, vocab.txt> <data dir> <out dir> <repo root>
# Linkage (staticproof + the platform's own lister), fidelity on speed.jsonl (4 threads) for float32 and int8 against
# ref_gte.npy, and one-thread speed on len200.jsonl, 3 interleaved rounds per model. Never stops on a failed step.
set -uo pipefail
BIN=$1; M=$(cd "$2" && pwd); D=$(cd "$3" && pwd); O=$(mkdir -p "$4" && cd "$4" && pwd); REPO=$(cd "$5" && pwd)
export VOCAB=$M/vocab.txt
{
  echo "== host"; uname -a; (lscpu 2>/dev/null || sysctl -n machdep.cpu.brand_string hw.ncpu 2>/dev/null || wmic cpu get name,numberofcores 2>/dev/null || true) | head -30
  echo "== file"; file "$BIN" 2>&1 || true
  echo "== size"; wc -c < "$BIN"
  echo "== staticproof"; go -C "$REPO/plugins/frank-exchange-of-views/tools" run ./third_party/pins/staticproof "$BIN"; echo "staticproof exit=$?"
  echo "== platform dependency listing"
  case "$(uname -s)" in
    Linux) ldd "$BIN"; echo "ldd exit=$?"; readelf -d "$BIN" 2>&1 | head -20 ;;
    Darwin) otool -L "$BIN"; echo "otool exit=$?" ;;
    *) if command -v objdump >/dev/null; then objdump -p "$BIN" | grep -i "DLL Name"; echo "objdump exit=$?"; fi
       VSW="/c/Program Files (x86)/Microsoft Visual Studio/Installer/vswhere.exe"
       if [ -x "$VSW" ]; then VS=$("$VSW" -latest -products '*' -property installationPath | tr -d '\r')
         DB=$(find "$(cygpath -u "$VS")/VC/Tools/MSVC" -iname dumpbin.exe -path '*Hostx64/x64*' 2>/dev/null | head -1)
         [ -z "$DB" ] && DB=$(find "$(cygpath -u "$VS")/VC/Tools/MSVC" -iname dumpbin.exe 2>/dev/null | head -1)
         [ -n "$DB" ] && { "$DB" /nologo /dependents "$(cygpath -w "$BIN")"; echo "dumpbin exit=$?"; }; fi ;;
  esac
} > "$O/linkage.txt" 2>&1
cat "$O/linkage.txt"
for m in gte-small gte-small.int8; do
  "$BIN" "$M/$m.ort" "$D/speed.jsonl" "$O/speed-$m.bin" 4 > "$O/fid-$m.txt" 2>&1; echo "run exit=$?" >> "$O/fid-$m.txt"
  "$BIN" cos "$O/speed-$m.bin" "$D/ref_gte.npy" >> "$O/fid-$m.txt" 2>&1; echo "cos exit=$?" >> "$O/fid-$m.txt"
  echo "== fidelity $m"; cat "$O/fid-$m.txt"
done
for r in 1 2 3; do for m in gte-small gte-small.int8; do
  echo "round=$r model=$m $("$BIN" "$M/$m.ort" "$D/len200.jsonl" "$O/len200-$m.bin" 1 2>&1 | tail -1)" | tee -a "$O/speed.txt"
done; done
rm -f "$O"/*.bin
