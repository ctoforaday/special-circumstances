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
# Source patches for non-MSVC Windows (zig/mingw), each harmless elsewhere; listed in the log.
patch_src() { # patch_src <file> <sed expr>
  sed -i.orig -e "$2" "$SRC/$1" && echo "patched $1: $2"
}
if [ "$TC" = zig ]; then
  # The MSVC STL /DELAYLOAD workaround is applied to the onnxruntime target, an INTERFACE target in a static build.
  # ORT_WARNING_FLAGS is MSVC's /wdNNNN list under `if (WIN32)`; clang reads "/wd4251" as an input file.
  patch_src cmake/CMakeLists.txt '/^set(ORT_WARNING_FLAGS)$/{n;s/^if (WIN32)$/if (MSVC)/;}'
  patch_src cmake/onnxruntime.cmake 's/^if (WIN32 AND NOT CMAKE_CXX_STANDARD_LIBRARIES MATCHES kernel32.lib)/if (MSVC AND NOT CMAKE_CXX_STANDARD_LIBRARIES MATCHES kernel32.lib)/'
fi

case $TARGET in
  linux-amd64)   TRIPLE=x86_64-linux-musl;   SYS=Linux;   PROC=x86_64 ;;
  linux-arm64)   TRIPLE=aarch64-linux-musl;  SYS=Linux;   PROC=aarch64 ;;
  windows-amd64) TRIPLE=x86_64-windows-gnu;  SYS=Windows; PROC=AMD64 ;;
  windows-arm64) TRIPLE=aarch64-windows-gnu; SYS=Windows; PROC=ARM64 ;;
  darwin-amd64)  TRIPLE=x86_64-macos;        SYS=Darwin;  PROC=x86_64 ;;
  darwin-arm64)  TRIPLE=aarch64-macos;       SYS=Darwin;  PROC=arm64 ;;
  *) echo "unknown target $TARGET" >&2; exit 2 ;;
esac

ARGS=(--build_dir "$BD/build" ${ORT_PHASES:---update --build} --parallel "$J" --skip_tests --skip_submodule_sync
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
if [ "$TC" = native ]; then case "$(uname -s)" in MINGW*|MSYS*|CYGWIN*)
  VSW="/c/Program Files (x86)/Microsoft Visual Studio/Installer/vswhere.exe"
  V=$("$VSW" -latest -products '*' -property installationVersion | tr -d '\r'); echo "Visual Studio $V"
  case ${V%%.*} in 17) ARGS+=(--cmake_generator "Visual Studio 17 2022") ;; 18) ARGS+=(--cmake_generator "Visual Studio 18 2026") ;; esac ;;
esac; fi
CFG=$(printf '%s\n' "${ARGS[@]}" | grep -A1 -- '--config' | tail -1)

AR=ar
if [ "$TC" = zig ]; then
  ZIG=${ZIG:-zig}; ZB=$WORK/zigbin-$TARGET; mkdir -p "$ZB"
  # zig cc maps -march to a CPU name and rejects ORT's per-file "-march=armv8.2-a+bf16" (and the cmake probe of it);
  # the same request spelled as zig CPU features is -mcpu=generic+v8_2a+bf16 (fp16 is "fullfp16" there).
  cat > "$ZB/marchfix" <<'EOS'
for a; do shift; case $a in
  -march=armv8.2-a*) r=$(printf '%s' "${a#-march=armv8.2-a}" | sed 's/+fp16/+fullfp16/g'); set -- "$@" "-mcpu=generic+v8_2a$r" ;;
  *) set -- "$@" "$a" ;; esac; done
EOS
  # Windows: ORT's minimal build adds -fno-unwind-tables -fno-asynchronous-unwind-tables, and cmake passes compile
  # flags to the link, where zig builds mingw's crtexe.c with them and its .seh_handler asm then fails
  # (".seh_ directive must appear within an active frame"); x64/arm64 Windows needs unwind info anyway.
  case $TRIPLE in *windows*) cat >> "$ZB/marchfix" <<'EOS'
for a; do shift; case $a in -fno-unwind-tables|-fno-asynchronous-unwind-tables) ;; *) set -- "$@" "$a" ;; esac; done
EOS
  ;; esac
  { printf '#!/bin/sh\n'; cat "$ZB/marchfix"; printf 'exec %q cc -target %s "$@"\n' "$ZIG" "$TRIPLE"; } > "$ZB/zigcc"
  CXXEXTRA=
  if [ "$SYS" = Darwin ]; then
    # abseil's cctz asks CoreFoundation for the local zone name on __APPLE__; zig carries no macOS SDK, so the six
    # declarations it uses are supplied here (Apple's signatures). The link resolves them against CoreFoundation,
    # which FEOV's darwin allowlist already holds.
    mkdir -p "$ZB/cfstub/CoreFoundation"
    cat > "$ZB/cfstub/CoreFoundation/CFTimeZone.h" <<'EOS'
#pragma once
#include <stdint.h>
#ifdef __cplusplus
extern "C" {
#endif
typedef const struct __CFTimeZone *CFTimeZoneRef;
typedef const struct __CFString *CFStringRef;
typedef const void *CFTypeRef;
typedef long CFIndex;
typedef uint32_t CFStringEncoding;
typedef unsigned char Boolean;
enum { kCFStringEncodingUTF8 = 0x08000100 };
CFTimeZoneRef CFTimeZoneCopyDefault(void);
CFStringRef CFTimeZoneGetName(CFTimeZoneRef tz);
CFIndex CFStringGetLength(CFStringRef theString);
CFIndex CFStringGetMaximumSizeForEncoding(CFIndex length, CFStringEncoding encoding);
Boolean CFStringGetCString(CFStringRef theString, char *buffer, CFIndex bufferSize, CFStringEncoding encoding);
void CFRelease(CFTypeRef cf);
#ifdef __cplusplus
}
#endif
EOS
    CXXEXTRA="-isystem $ZB/cfstub"
  fi
  { printf '#!/bin/sh\n'; cat "$ZB/marchfix"; printf 'exec %q c++ -target %s %s "$@"\n' "$ZIG" "$TRIPLE" "$CXXEXTRA"; } > "$ZB/zigcxx"
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
  # ORT links Iconv on every non-Windows target but calls none of it; macOS keeps iconv outside libSystem, and zig
  # carries no SDK to find it in, so it is declared built in rather than adding a libiconv dependency.
  # mlas and the platform probe choose kernels by CMAKE_OSX_ARCHITECTURES on Apple; unset, they pick the host's (x86).
  [ "$SYS" = Darwin ] && DEFS+=(Iconv_IS_BUILT_IN=TRUE CMAKE_OSX_ARCHITECTURES=$PROC)
fi

ARGS+=(--cmake_extra_defines "${DEFS[@]}")
start=$(date +%s)
( cd "$SRC" && "$PY" "$SRC/tools/ci_build/build.py" "${ARGS[@]}" ) > "$BD/build.log" 2>&1 || {
  rc=$?; echo "build.py failed rc=$rc after $(( $(date +%s) - start ))s; last lines:"; tail -60 "$BD/build.log"; exit $rc; }
echo "build.py ok in $(( $(date +%s) - start ))s ($NAME)"
"$HERE/combine.sh" "$BD/build/$CFG" "$SRC" "$OUT" "$AR"
find "$BD/build/$CFG" \( -name '*.a' -o -name '*.lib' \) -exec ls -l {} + > "$OUT/archive-sizes.txt"
echo "$NAME" > "$OUT/name.txt"
