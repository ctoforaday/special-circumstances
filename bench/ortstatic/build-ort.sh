#!/usr/bin/env bash
# build-ort.sh <target> <work> <native|zig> [variant]
# Builds ONNX Runtime v1.31.1 as static libraries, CPU EP only, reduced to the operators in $ORT_OPS_CONFIG,
# then merges the archives into <work>/<target>-<toolchain>-<variant>/out/lib/libonnxruntime_static.a.
#   target:    linux-amd64 | linux-arm64 | windows-amd64 | windows-arm64 | darwin-amd64 | darwin-arm64
#   toolchain: native (the host's own compiler; target must be the host) | zig (zig cc/c++ cross, as build-cstack.sh does)
#   variant:   min (default: --minimal_build, no exceptions/RTTI/ML ops, typed op reduction, MinSizeRel)
#              minext (as min, --minimal_build extended)
#              reduced (full runtime, .onnx loading, ops reduced by config without type reduction; exceptions on)
#              full (no reduction at all; for the size baseline)
# Env: ORT_OPS_CONFIG (required for min/minext/reduced), JOBS (default 8), PYTHON (default: a venv under <work> with flatbuffers), ZIG, CMAKE_DIR,
#      ORT_SRC (default <work>/src/onnxruntime), ORT_TAG (default v1.31.1).
set -euo pipefail
TARGET=$1; WORK=$(mkdir -p "$2" && cd "$2" && pwd); TC=$3; VAR=${4:-min}
J=${JOBS:-8}; PY=${PYTHON:-}; TAG=${ORT_TAG:-v1.31.1}; SRC=${ORT_SRC:-$WORK/src/onnxruntime}
HERE=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
NAME=$TARGET-$TC-$VAR; BD=$WORK/$NAME; OUT=$BD/out; mkdir -p "$BD"
[ -n "${CMAKE_DIR:-}" ] && export PATH=$CMAKE_DIR:$PATH
cmake --version | head -1

# build.py's reduced-ops path imports the flatbuffers Python module; without it parse_config is missing.
if [ -z "$PY" ]; then
  HP=python3; command -v python3 >/dev/null 2>&1 && python3 -c 1 >/dev/null 2>&1 || HP=python
  if [ ! -d "$WORK/pyenv" ]; then "$HP" -m venv "$WORK/pyenv"; fi
  PY=$WORK/pyenv/bin/python; [ -x "$PY" ] || PY=$WORK/pyenv/Scripts/python.exe
  "$PY" -m pip install -q flatbuffers
fi
if [ ! -d "$SRC/.git" ]; then
  git clone -q --depth 1 --branch "$TAG" https://github.com/microsoft/onnxruntime.git "$SRC"
fi
git -C "$SRC" log -1 --format='ort source %H %cd'

case $TARGET in
  linux-amd64)   TRIPLE=x86_64-linux-musl;   SYS=Linux;   PROC=x86_64 ;;
  linux-arm64)   TRIPLE=aarch64-linux-musl;  SYS=Linux;   PROC=aarch64 ;;
  windows-amd64) TRIPLE=x86_64-windows-gnu;  SYS=Windows; PROC=AMD64 ;;
  windows-arm64) TRIPLE=aarch64-windows-gnu; SYS=Windows; PROC=ARM64 ;;
  darwin-amd64)  TRIPLE=x86_64-macos;        SYS=Darwin;  PROC=x86_64 ;;
  darwin-arm64)  TRIPLE=aarch64-macos;       SYS=Darwin;  PROC=arm64 ;;
  *) echo "unknown target $TARGET" >&2; exit 2 ;;
esac

ARGS=(--build_dir "$BD/build" --update --build --parallel "$J" --skip_tests --skip_submodule_sync
      --compile_no_warning_as_error)
DEFS=(onnxruntime_BUILD_UNIT_TESTS=OFF)
case $VAR in
  min)     ARGS+=(--config MinSizeRel --minimal_build --disable_exceptions --disable_rtti --disable_ml_ops
                  --include_ops_by_config "$ORT_OPS_CONFIG" --enable_reduced_operator_type_support) ;;
  minext)  ARGS+=(--config MinSizeRel --minimal_build extended --disable_exceptions --disable_rtti --disable_ml_ops
                  --include_ops_by_config "$ORT_OPS_CONFIG" --enable_reduced_operator_type_support) ;;
  reduced) ARGS+=(--config MinSizeRel --disable_ml_ops --include_ops_by_config "$ORT_OPS_CONFIG") ;;
  full)    ARGS+=(--config Release) ;;
  *) echo "unknown variant $VAR" >&2; exit 2 ;;
esac
CFG=$(printf '%s\n' "${ARGS[@]}" | grep -A1 -- '--config' | tail -1)

AR=ar
if [ "$TC" = zig ]; then
  ZIG=${ZIG:-zig}; ZB=$WORK/zigbin-$TARGET; mkdir -p "$ZB"
  printf '#!/bin/sh\nexec %q cc -target %s "$@"\n' "$ZIG" "$TRIPLE" > "$ZB/zigcc"
  printf '#!/bin/sh\nexec %q c++ -target %s "$@"\n' "$ZIG" "$TRIPLE" > "$ZB/zigcxx"
  printf '#!/bin/sh\nexec %q ar "$@"\n' "$ZIG" > "$ZB/zigar"
  printf '#!/bin/sh\nexec %q ranlib "$@"\n' "$ZIG" > "$ZB/zigranlib"
  chmod +x "$ZB"/*
  export ZIG_GLOBAL_CACHE_DIR=$WORK/zig-cache
  AR=$ZB/zigar
  DEFS+=(CMAKE_SYSTEM_NAME=$SYS CMAKE_SYSTEM_PROCESSOR=$PROC
         CMAKE_C_COMPILER=$ZB/zigcc CMAKE_CXX_COMPILER=$ZB/zigcxx CMAKE_ASM_COMPILER=$ZB/zigcc
         CMAKE_AR=$ZB/zigar CMAKE_RANLIB=$ZB/zigranlib
         CMAKE_C_COMPILER_AR=$ZB/zigar CMAKE_CXX_COMPILER_AR=$ZB/zigar
         CMAKE_C_COMPILER_RANLIB=$ZB/zigranlib CMAKE_CXX_COMPILER_RANLIB=$ZB/zigranlib
         CMAKE_POLICY_VERSION_MINIMUM=3.5 ${ZIG_EXTRA_DEFINES:-})
fi

ARGS+=(--cmake_extra_defines "${DEFS[@]}")
start=$(date +%s)
( cd "$SRC" && "$PY" "$SRC/tools/ci_build/build.py" "${ARGS[@]}" ) > "$BD/build.log" 2>&1 || {
  rc=$?; echo "build.py failed rc=$rc after $(( $(date +%s) - start ))s; last lines:"; tail -60 "$BD/build.log"; exit $rc; }
echo "build.py ok in $(( $(date +%s) - start ))s ($NAME)"
"$HERE/combine.sh" "$BD/build/$CFG" "$SRC" "$OUT" "$AR"
find "$BD/build/$CFG" \( -name '*.a' -o -name '*.lib' \) -exec ls -l {} + > "$OUT/archive-sizes.txt"
echo "$NAME" > "$OUT/name.txt"
