#!/usr/bin/env bash
# build.sh <x64|ARM64> <work> <out>
# Windows only (Git Bash on a hosted runner). Builds ONNX Runtime v1.31.1 with MSVC as static libraries against the
# STATIC C runtime (/MT, --enable_msvc_static_runtime): minimal build, operators and types reduced to
# ../ops.config, Release, device discovery off (as in the zig builds). Then builds embed-probe.cpp with /MT and links
# every runtime archive into it. Writes <out>/embed-probe.exe, <out>/libs.txt, <out>/crt-directives.txt and logs.
set -euo pipefail
ARCH=$1; WORK=$(mkdir -p "$2" && cd "$2" && pwd); OUT=$(mkdir -p "$3" && cd "$3" && pwd)
HERE=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd); OPS=$HERE/../ops.config
J=${JOBS:-8}; SRC=$WORK/src/onnxruntime; BD=$WORK/build
VSW="/c/Program Files (x86)/Microsoft Visual Studio/Installer/vswhere.exe"
V=$("$VSW" -latest -products '*' -property installationVersion | tr -d '\r')
VSDIR=$("$VSW" -latest -products '*' -property installationPath | tr -d '\r')
case ${V%%.*} in 17) GEN="Visual Studio 17 2022" ;; 18) GEN="Visual Studio 18 2026" ;; *) echo "Visual Studio $V unknown" >&2; exit 2 ;; esac
echo "Visual Studio $V at $VSDIR; generator $GEN; arch $ARCH"

HP=python; command -v python >/dev/null 2>&1 || HP=python3
"$HP" -c 'import platform,sys; print("python", sys.version.split()[0], platform.machine())'
[ -d "$WORK/pyenv" ] || "$HP" -m venv "$WORK/pyenv"
PY=$WORK/pyenv/Scripts/python.exe
"$PY" -m pip install -q flatbuffers
MACH=$("$PY" -c 'import platform; print(platform.machine())' | tr -d '\r'); echo "venv python machine: $MACH"
XARGS=()
# build.py picks the Visual Studio platform from the interpreter's machine; an emulated x64 Python on an arm64
# host would build x64, so arm64 is then asked for explicitly.
[ "$ARCH" = ARM64 ] && [ "$MACH" != ARM64 ] && XARGS+=(--arm64)

[ -d "$SRC/.git" ] || git clone -q --depth 1 --branch v1.31.1 https://github.com/microsoft/onnxruntime.git "$SRC"
git -C "$SRC" log -1 --format='ort source %H %cd'
grep -n 'enable_msvc_static_runtime' "$SRC/tools/ci_build/build_args.py"

start=$(date +%s)
( cd "$SRC" && "$PY" tools/ci_build/build.py --build_dir "$BD" --config Release --update --build --parallel "$J" \
    --skip_tests --skip_submodule_sync --compile_no_warning_as_error --cmake_generator "$GEN" \
    --minimal_build --disable_exceptions --disable_rtti --disable_ml_ops \
    --include_ops_by_config "$OPS" --enable_reduced_operator_type_support \
    --enable_msvc_static_runtime ${XARGS[@]+"${XARGS[@]}"} \
    --cmake_extra_defines onnxruntime_BUILD_UNIT_TESTS=OFF onnxruntime_DISABLE_DEVICE_DISCOVERY=ON ) \
  > "$OUT/ort-build.log" 2>&1 || { rc=$?; echo "build.py failed rc=$rc after $(( $(date +%s) - start ))s"; tail -80 "$OUT/ort-build.log"; exit $rc; }
echo "ORT build.py seconds: $(( $(date +%s) - start ))"
grep -m3 -E 'CMAKE_MSVC_RUNTIME_LIBRARY|MultiThreaded' "$OUT/ort-build.log" || true
# A static build with no executable in ALL leaves cpuinfo unbuilt; it reaches the link only through the INTERFACE target.
cmake --build "$BD/Release" --config Release --target cpuinfo -j "$J" >> "$OUT/ort-build.log" 2>&1 && echo "built cpuinfo" || echo "no target cpuinfo"

# The runtime's archives, as combine.sh selects them: host-only tools (protoc, full protobuf, upb, flatc's library)
# and tests left out.
find "$BD/Release" -name '*.lib' ! -name 'protoc*.lib' ! -name 'libprotoc*.lib' ! -name 'libprotobuf.lib' ! -name 'libprotobufd.lib' \
  ! -name 'flatbuffers.lib' ! -name 'upb*.lib' ! -path '*gtest*' ! -path '*gmock*' ! -path '*benchmark*' ! -name '*test*' \
  ! -path '*/embed-probe/*' | sort > "$OUT/libs-unix.txt"
while IFS= read -r l; do cygpath -m "$l"; done < "$OUT/libs-unix.txt" > "$OUT/libs.txt"
wc -l < "$OUT/libs.txt"; while IFS= read -r l; do wc -c < "$l" | tr -d '\n'; echo " $l"; done < "$OUT/libs-unix.txt" > "$OUT/lib-sizes.txt"

# The image carries several toolsets, some partial (an old one's dumpbin printed nothing); the newest host-native one.
find "$VSDIR/VC/Tools/MSVC" -iname dumpbin.exe > "$OUT/dumpbin-all.txt" || true
DUMPBIN=$(grep -i "/Host${ARCH,,}/${ARCH,,}/" "$OUT/dumpbin-all.txt" | sort -V | tail -1 || true)
[ -n "$DUMPBIN" ] || DUMPBIN=$(sort -V "$OUT/dumpbin-all.txt" | tail -1)
echo "$DUMPBIN" > "$OUT/dumpbin-path.txt"; echo "dumpbin: $DUMPBIN"
# Which C runtime each archive asks for: /DEFAULTLIB:LIBCMT (static, /MT) vs MSVCRT (DLL, /MD).
{ while IFS= read -r l; do
    d=$("$DUMPBIN" -nologo -directives "$l" 2>&1 | tr -d '\r' || true)
    printf '%s libcmt=%s msvcrt=%s libcpmt=%s msvcprt=%s\n' "$(basename "$l")" \
      "$(grep -ciE 'DEFAULTLIB:"?LIBCMT' <<<"$d")" "$(grep -ciE 'DEFAULTLIB:"?MSVCRT' <<<"$d")" \
      "$(grep -ciE 'DEFAULTLIB:"?LIBCPMT' <<<"$d")" "$(grep -ciE 'DEFAULTLIB:"?MSVCPRT' <<<"$d")"
  done < "$OUT/libs.txt"; } > "$OUT/crt-directives.txt"
cat "$OUT/crt-directives.txt"

# The probe: /MT, every runtime archive, the OS import libraries ORT may reference (an unreferenced one adds no import).
PB=$WORK/embed-probe; mkdir -p "$PB"
cat > "$PB/CMakeLists.txt" <<EOS
cmake_minimum_required(VERSION 3.21)
project(embedprobe CXX)
set(CMAKE_MSVC_RUNTIME_LIBRARY MultiThreaded)
add_executable(embed-probe "$(cygpath -m "$HERE/embed-probe.cpp")")
target_compile_features(embed-probe PRIVATE cxx_std_17)
target_include_directories(embed-probe PRIVATE "$(cygpath -m "$SRC/include/onnxruntime/core/session")")
file(STRINGS "$(cygpath -m "$OUT/libs.txt")" ORT_LIBS)
target_link_libraries(embed-probe PRIVATE \${ORT_LIBS} advapi32 dbghelp setupapi pathcch shlwapi dxgi)
target_link_options(embed-probe PRIVATE /VERBOSE:LIB)
EOS
start=$(date +%s)
{ cmake -S "$PB" -B "$PB/b" -G "$GEN" -A "$ARCH" && cmake --build "$PB/b" --config Release -j "$J" -- -v:n; } > "$OUT/probe-build.log" 2>&1 \
  || { rc=$?; echo "probe build failed rc=$rc"; grep -E 'error|LNK|warning LNK' "$OUT/probe-build.log" | head -80; exit $rc; }
echo "probe build seconds: $(( $(date +%s) - start ))"
grep -E 'warning LNK|LNK4098|LNK2038' "$OUT/probe-build.log" | sort | uniq -c | head -20 || true
grep -iE 'Searching .*(msvcrt|libcmt|libcpmt|msvcprt|vcruntime|ucrt)' "$OUT/probe-build.log" | sed 's/^ *//' | sort -u | head -20 || true
cp "$PB/b/Release/embed-probe.exe" "$OUT/"
ls -l "$OUT/embed-probe.exe"; ls "$PB/b/Release"
