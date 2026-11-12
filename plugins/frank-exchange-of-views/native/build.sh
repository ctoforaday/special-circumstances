#!/bin/bash
# build.sh — build one native helper for one target.
#
# Usage:
#   build.sh <helper> <target> [outdir] [workdir]
#
# Helpers: feov-ocr | feov-embed
# Targets: linux-amd64 | linux-arm64 | windows-amd64 | windows-arm64 | darwin-arm64
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
NATIVE_DIR="$SCRIPT_DIR"
REPO_ROOT="$(cd "$NATIVE_DIR/../../.." && pwd)"
TOOLS_DIR="$REPO_ROOT/plugins/frank-exchange-of-views/tools"
TOOLCHAIN_PINS="$NATIVE_DIR/PINS.txt"

die() { echo "native/build.sh: $*" >&2; exit 1; }

[ $# -ge 2 ] || die "usage: build.sh <helper> <target> [outdir] [workdir]"

HELPER="$1"
TARGET="$2"
OUTDIR="${3:-}"
WORK_ARG="${4:-}"

if command -v cygpath >/dev/null 2>&1; then
	[ -n "$OUTDIR" ] && OUTDIR="$(cygpath -u "$OUTDIR")"
	[ -n "$WORK_ARG" ] && WORK_ARG="$(cygpath -u "$WORK_ARG")"
	[ -n "${RUNNER_TEMP:-}" ] && RUNNER_TEMP="$(cygpath -u "$RUNNER_TEMP")"
fi

[ -d "$NATIVE_DIR/$HELPER" ] || die "unknown helper $HELPER (no $NATIVE_DIR/$HELPER directory)"
[ -f "$NATIVE_DIR/$HELPER/ASSETS" ] || die "$HELPER is not a helper (no ASSETS file)"
[ -f "$NATIVE_DIR/$HELPER/PINS.txt" ] || die "$HELPER has no PINS.txt"

case "$TARGET" in
	linux-amd64)   TRIPLE=x86_64-linux-musl;  SYSNAME=Linux;   PROC=x86_64;  OS="linux";   ARCH="amd64"; EXT=""     ;;
	linux-arm64)   TRIPLE=aarch64-linux-musl; SYSNAME=Linux;   PROC=aarch64; OS="linux";   ARCH="arm64"; EXT=""     ;;
	windows-amd64) TRIPLE=x86_64-windows-gnu; SYSNAME=Windows; PROC=AMD64;   OS="windows"; ARCH="amd64"; EXT=".exe" ;;
	windows-arm64) TRIPLE=aarch64-windows-gnu; SYSNAME=Windows; PROC=aarch64; OS="windows"; ARCH="arm64"; EXT=".exe" ;;
	darwin-arm64)  TRIPLE=aarch64-macos;      SYSNAME=Darwin;  PROC=arm64;   OS="darwin";  ARCH="arm64"; EXT=""     ;;
	*) die "unknown target $TARGET" ;;
esac

# 1. nativedigest -check and nativedigest -print
if command -v go >/dev/null 2>&1; then
	( cd "$TOOLS_DIR" && go run ./devcmd/nativedigest -check ) || die "nativedigest -check failed; run devcmd/nativedigest"
	DIGEST="$(cd "$TOOLS_DIR" && go run ./devcmd/nativedigest -print "$HELPER")"
	[ -n "$DIGEST" ] || die "nativedigest -print returned empty digest"
else
	die "go is required to run nativedigest"
fi

if [ -n "$WORK_ARG" ]; then
	WORK="$WORK_ARG"
else
	# Keyed by the source digest: a work directory holds intermediate builds (ONNX Runtime, the
	# engine libraries) that are reused when present, so one left by other sources is never linked.
	# Its first 16 hex digits name it: ONNX Runtime's KleidiAI sources sit ~200 characters below
	# the work directory, and the full 64 put them past Windows' 260-character path limit.
	WORK="${RUNNER_TEMP:-${XDG_CACHE_HOME:-$HOME/.cache}/sc-native}/$HELPER/$TARGET/${DIGEST:0:16}"
fi

[ -n "$OUTDIR" ] || OUTDIR="$WORK/dist"
mkdir -p "$OUTDIR" "$WORK" "$WORK/src" "$WORK/tools" "$WORK/build" "$WORK/prefix"
PREFIX="$WORK/prefix"

# Exclusive lock on workdir
exec 9>"$WORK/.build.lock"
if command -v flock >/dev/null 2>&1; then
	if ! flock -n 9; then
		echo "native/build.sh: waiting for the lock on $WORK (another build is running)"
		flock 9
	fi
fi

J="${CSTACK_JOBS:-$(nproc 2>/dev/null || sysctl -n hw.ncpu 2>/dev/null || echo 4)}"

# ---- pin manifest access ------------------------------------------------------------
pin_field() { # pin_field <file> <manifest> <n: 1=sha 3=url>
	awk -v f="$1" -v n="$3" '$0 !~ /^#/ && $2 == f { print $n }' "$2" | tr -d '\r'
}

fetch_pinned() { # fetch_pinned <file> <manifest> <destdir>
	local file="$1" manifest="$2" dest="$3/$1" sha url
	sha="$(pin_field "$file" "$manifest" 1)"
	url="$(pin_field "$file" "$manifest" 3)"
	[ -n "$sha" ] && [ -n "$url" ] || die "$file is not pinned in $manifest"
	if [ ! -f "$dest" ]; then
		echo "fetching $file"
		local fetched=0
		if command -v gh >/dev/null 2>&1 && [[ "$url" =~ github\.com/([^/]+)/([^/]+)/releases/download/([^/]+)/(.+) ]]; then
			local owner="${BASH_REMATCH[1]}"
			local repo="${BASH_REMATCH[2]}"
			local tag="${BASH_REMATCH[3]}"
			local asset="${BASH_REMATCH[4]}"
			if gh release download "$tag" -R "$owner/$repo" -p "$asset" -D "$3" 2>/dev/null && [ -f "$dest" ]; then
				fetched=1
			fi
		fi
		if [ "$fetched" -eq 0 ]; then
			curl -fsSL --retry 5 --retry-all-errors --connect-timeout 20 --max-time 600 -o "$dest.part" "$url" ||
				die "downloading $file from $url failed"
			mv "$dest.part" "$dest"
		fi
	fi
	echo "$sha  $dest" | sha256sum -c --quiet - || die "$file does not match its pin"
}

# ---- extra assets (native/<helper>/ASSETS) ------------------------------------------
# nativedigest -assets resolves each ASSETS line to the export pinned for this arch, so the
# arch -> export mapping is ASSETS', read by one parser, never composed here.
ASSET_PAIRS="$(cd "$TOOLS_DIR" && go run ./devcmd/nativedigest -assets "$HELPER" -arch "$ARCH")" ||
	die "nativedigest -assets $HELPER -arch $ARCH failed"

fetch_assets() { # fetch each export ASSETS names for $ARCH from the helper's pins
	local file exported
	while read -r file exported; do
		[ -n "$file" ] || continue
		fetch_pinned "$exported" "$NATIVE_DIR/$HELPER/PINS.txt" "$WORK/src"
	done <<<"$ASSET_PAIRS"
}

ship_assets() { # copy each export into OUTDIR under its release name; install-helpers and
	# fetch-bin.sh install it under the ASSETS line's file name
	local file exported
	while read -r file exported; do
		[ -n "$file" ] || continue
		cp "$WORK/src/$exported" "$OUTDIR/$exported"
	done <<<"$ASSET_PAIRS"
}

# The flatbuffers Python runtime ONNX Runtime's build scripts import: the pinned wheel, put on
# PYTHONPATH (a pure-Python wheel imports from the zip), so no package index is consulted.
FLATBUFFERS_WHEEL=flatbuffers-25.2.10-py2.py3-none-any.whl

# ONNX Runtime's CMake fetches its dependencies from the URLs in its cmake/deps.txt, and every
# declaration but these passes deps.txt's SHA1 as URL_HASH, which CMake checks before it extracts.
# These are declared with no hash: each is pinned in PINS.txt and put in ORT's mirror directory
# (<source>/mirror/<URL without https://>), which ORT reads in place of the download.
ORT_UNHASHED_DEPS=(mp11)

ort_mirror_unhashed() { # ort_mirror_unhashed <ORT source dir>
	local src="$1" dep url file
	for dep in "${ORT_UNHASHED_DEPS[@]}"; do
		url="$(awk -F';' -v d="$dep" '$1 == d { print $2 }' "$src/cmake/deps.txt" | tr -d '\r')"
		[ -n "$url" ] || die "ONNX Runtime's cmake/deps.txt names no $dep"
		file="${url##*/}"
		fetch_pinned "$file" "$NATIVE_DIR/$HELPER/PINS.txt" "$WORK/src"
		mkdir -p "$src/mirror/$(dirname "${url#https://}")"
		cp "$WORK/src/$file" "$src/mirror/${url#https://}"
	done
}

ort_check_fetches() { # ort_check_fetches <ORT CMake build dir>: refuses a dependency fetched unverified
	# Run between configure and build. Configure has already downloaded and extracted every
	# dependency and add_subdirectory'd it, so an unhashed one's CMake code has run by now; none is
	# compiled or linked until the build, which this refusal stops. It cannot refuse earlier:
	# FetchContent.cmake has no include guard, so ORT's include(FetchContent) replaces any override
	# of FetchContent_Declare, and a dependency provider never sees ORT's FetchContent_Populate
	# calls. Each URL download CMake makes writes a download-<name>.cmake script, which holds a
	# file(<ALGO> ...) call only when the declaration carried a hash.
	local dir="$1" n=0 s
	while IFS= read -r s; do
		n=$((n + 1))
		grep -qE '^ *file\("(SHA1|SHA224|SHA256|SHA384|SHA512|SHA3_224|SHA3_256|SHA3_384|SHA3_512)" ' "$s" ||
			die "ONNX Runtime's build downloads a dependency it does not hash-check: $s (pin it in ORT_UNHASHED_DEPS)"
	done < <(find "$dir" -name 'download-*.cmake')
	[ -z "$(find "$dir" -name 'gitclone-*.cmake' -print -quit)" ] || die "ONNX Runtime's build clones a git repository, which no hash pins"
	[ "$n" -gt 0 ] || die "no dependency download under $dir: the hash check would have checked nothing"
	echo "native/build.sh: each of ONNX Runtime's $n dependency downloads checks its hash"
}

# Windows MSVC native helper build for feov-embed (plans/native-helpers.md §III.1.5)
if [ "$HELPER" = "feov-embed" ] && [ "$OS" = "windows" ]; then
	HELPER_PINS="$NATIVE_DIR/$HELPER/PINS.txt"
	fetch_assets

	VSW="/c/Program Files (x86)/Microsoft Visual Studio/Installer/vswhere.exe"
	[ -f "$VSW" ] || die "vswhere.exe not found at $VSW"
	V=$("$VSW" -latest -products '*' -property installationVersion | tr -d '\r')
	VSDIR=$("$VSW" -latest -products '*' -property installationPath | tr -d '\r')
	case ${V%%.*} in
	17) GEN="Visual Studio 17 2022" ;;
	18) GEN="Visual Studio 18 2026" ;;
	*) die "Visual Studio version $V unknown" ;;
	esac
	echo "native/build.sh: Visual Studio $V at $VSDIR (generator: $GEN)"

	case "$ARCH" in
	amd64) VS_ARCH="x64" ;;
	arm64) VS_ARCH="ARM64" ;;
	*) die "unsupported Windows architecture $ARCH" ;;
	esac

	PY=python
	command -v python >/dev/null 2>&1 || PY=python3
	fetch_pinned "$FLATBUFFERS_WHEEL" "$HELPER_PINS" "$WORK/src"
	PYTHONPATH="$(cygpath -w "$WORK/src/$FLATBUFFERS_WHEEL")"
	export PYTHONPATH

	fetch_pinned onnxruntime-1.31.1.tar.gz "$HELPER_PINS" "$WORK/src"
	SRC="$WORK/src/onnxruntime"
	if [ ! -d "$SRC" ]; then
		mkdir -p "$SRC"
		tar -C "$SRC" --strip-components=1 -xf "$WORK/src/onnxruntime-1.31.1.tar.gz"
	fi
	ort_mirror_unhashed "$SRC"

	if [ ! -f "$WORK/build/ort/Release/onnxruntime_common.lib" ]; then
		echo "=== Building ONNX Runtime ($TARGET) with MSVC ==="
		XARGS=()
		MACH=$("$PY" -c 'import platform; print(platform.machine())' | tr -d '\r')
		[ "$VS_ARCH" = "ARM64" ] && [ "$MACH" != "ARM64" ] && XARGS+=(--arm64)
		ORT_ARGS=(--build_dir "$WORK/build/ort" \
			--config Release \
			--parallel "$J" \
			--skip_tests --skip_submodule_sync --compile_no_warning_as_error \
			--cmake_generator "$GEN" \
			--minimal_build --disable_exceptions --disable_rtti --disable_ml_ops \
			--include_ops_by_config "$NATIVE_DIR/feov-embed/ops.config" \
			--enable_reduced_operator_type_support \
			--enable_msvc_static_runtime ${XARGS[@]+"${XARGS[@]}"} \
			--cmake_extra_defines onnxruntime_BUILD_UNIT_TESTS=OFF onnxruntime_DISABLE_DEVICE_DISCOVERY=ON)
		"$PY" "$SRC/tools/ci_build/build.py" --update "${ORT_ARGS[@]}"
		ort_check_fetches "$WORK/build/ort/Release"
		"$PY" "$SRC/tools/ci_build/build.py" --build "${ORT_ARGS[@]}"
		cmake --build "$WORK/build/ort/Release" --config Release --target cpuinfo -j "$J"
	fi

	ort_find_dir="$(cygpath -u "$WORK/build/ort/Release")"
	find "$ort_find_dir" -name '*.lib' \
		! -name 'onnxruntime_static.lib' \
		! -name 'protoc*.lib' ! -name 'libprotoc*.lib' ! -name 'libprotobuf.lib' ! -name 'libprotobufd.lib' \
		! -name 'flatbuffers.lib' ! -name 'upb*.lib' ! -path '*gtest*' ! -path '*gmock*' ! -path '*benchmark*' ! -name '*test*' \
		! -path '*/embed/*' | sort > "$WORK/libs-unix.txt"
	[ -s "$WORK/libs-unix.txt" ] || die "no static libraries found in $ort_find_dir"
	echo "Found $(wc -l < "$WORK/libs-unix.txt") static libraries to merge"
	while IFS= read -r l; do cygpath -m "$l"; done < "$WORK/libs-unix.txt" | tr -d '\r' > "$WORK/libs.txt"

	LIB_EXE="$("$VSW" -latest -products '*' -find '**/Hostx64/x64/lib.exe' | head -n 1 | tr -d '\r')"
	if [ -z "$LIB_EXE" ] || [ ! -f "$LIB_EXE" ]; then
		LIB_EXE="$(find "$VSDIR/VC/Tools/MSVC" -name "lib.exe" 2>/dev/null | grep "Hostx64/x64/lib.exe" | head -n 1 | tr -d '\r')"
	fi
	[ -n "$LIB_EXE" ] && [ -f "$LIB_EXE" ] || die "lib.exe not found in $VSDIR"

	ort_static_lib="$WORK/build/ort/Release/onnxruntime_static.lib"
	rm -f "$ort_static_lib"
	ort_static_lib_win="$(cygpath -m "$ort_static_lib" | tr -d '\r')"
	{
		printf '/OUT:%s\n' "$ort_static_lib_win"
		while IFS= read -r l; do
			printf '%s\n' "$l"
		done < "$WORK/libs.txt"
	} > "$WORK/libs-rsp.txt"

	rsp_win="$(cygpath -m "$WORK/libs-rsp.txt" | tr -d '\r')"
	echo "=== Merging ONNX Runtime static libraries with lib.exe ==="
	echo "libs-rsp.txt entries: $(wc -l < "$WORK/libs-rsp.txt")"
	head -n 5 "$WORK/libs-rsp.txt"
	lib_out="$("$LIB_EXE" @"$rsp_win" 2>&1)" || {
		echo "lib.exe failed with exit code $?: $lib_out"
		exit 1
	}
	[ -n "$lib_out" ] && echo "$lib_out"
	[ -f "$ort_static_lib" ] || die "lib.exe did not create $ort_static_lib"

	echo "=== Compiling feov-embed ($TARGET) with MSVC ==="
	EMBED_DIR="$WORK/embed"
	mkdir -p "$EMBED_DIR"
	cat > "$EMBED_DIR/CMakeLists.txt" <<EOS
cmake_minimum_required(VERSION 3.21)
project(feovembed CXX)
set(CMAKE_MSVC_RUNTIME_LIBRARY MultiThreaded)
add_executable(feov-embed
	"$(cygpath -m "$NATIVE_DIR/feov-embed/main.cpp")"
	"$(cygpath -m "$NATIVE_DIR/frame/frame.cpp")"
)
target_compile_features(feov-embed PRIVATE cxx_std_17)
target_compile_definitions(feov-embed PRIVATE FEOV_HELPER_DIGEST=\"${DIGEST}\")
target_include_directories(feov-embed PRIVATE
	"$(cygpath -m "$SRC/include/onnxruntime/core/session")"
	"$(cygpath -m "$NATIVE_DIR/frame")"
	"$(cygpath -m "$NATIVE_DIR/feov-embed")"
)
target_link_libraries(feov-embed PRIVATE
	"$(cygpath -m "$ort_static_lib")"
	advapi32 dbghelp setupapi pathcch shlwapi dxgi
)
target_compile_options(feov-embed PRIVATE /Brepro)
target_link_options(feov-embed PRIVATE /Brepro)
EOS
	cmake -S "$EMBED_DIR" -B "$EMBED_DIR/b" -G "$GEN" -A "$VS_ARCH"
	cmake --build "$EMBED_DIR/b" --config Release -j "$J"

	OUT_NAME="${HELPER}_${OS}_${ARCH}${EXT}"
	cp "$EMBED_DIR/b/Release/feov-embed.exe" "$OUTDIR/$OUT_NAME"
	ship_assets
	echo "native/build.sh: $OUTDIR/$OUT_NAME ready"
	exit 0
fi

# ---- Toolchain ----------------------------------------------------------------------
tool_pin() { # tool_pin <name>
	awk -v n="$1" '$0 !~ /^#/ && $2 ~ "^"n"-" { print $2 }' "$TOOLCHAIN_PINS"
}

unpack_tool() { # unpack_tool <tarball>
	local file="$1" dir="$WORK/tools/${1%.tar.*}"
	if [ ! -d "$dir" ]; then
		fetch_pinned "$file" "$TOOLCHAIN_PINS" "$WORK/src" >&2
		mkdir -p "$dir.part"
		tar -C "$dir.part" --strip-components=1 -xf "$WORK/src/$file"
		mv "$dir.part" "$dir"
	fi
	echo "$dir"
}

if [ -z "${ZIG:-}" ]; then
	[ "$(uname -sm)" = "Linux x86_64" ] || die "pinned zig is linux-x86_64; set ZIG on other hosts"
	f="$(tool_pin zig)"; [ -n "$f" ] || die "PINS.txt pins no zig"
	ZIG="$(unpack_tool "$f")/zig"
fi

if [ -z "${CMAKE:-}" ]; then
	[ "$(uname -sm)" = "Linux x86_64" ] || die "pinned cmake is linux-x86_64; set CMAKE on other hosts"
	f="$(tool_pin cmake)"; [ -n "$f" ] || die "PINS.txt pins no cmake"
	CMAKE="$(unpack_tool "$f")/bin/cmake"
fi

# Compiler wrappers
mkdir -p "$WORK/bin"
for tool in cc c++ ar ranlib; do
	name="zig${tool/c++/cxx}"
	printf '#!/bin/sh\nexec %q %s -target %s "$@"\n' "$ZIG" "$tool" "$TRIPLE" > "$WORK/bin/$name"
	case "$tool" in ar|ranlib) printf '#!/bin/sh\nexec %q %s "$@"\n' "$ZIG" "$tool" > "$WORK/bin/$name" ;; esac
	chmod +x "$WORK/bin/$name"
done
export CC="$WORK/bin/zigcc" CXX="$WORK/bin/zigcxx"
AR="$WORK/bin/zigar" RANLIB="$WORK/bin/zigranlib"

CROSS=(
	-DCMAKE_SYSTEM_NAME="$SYSNAME" -DCMAKE_SYSTEM_PROCESSOR="$PROC"
	-DCMAKE_C_COMPILER="$CC" -DCMAKE_CXX_COMPILER="$CXX"
	-DCMAKE_AR="$AR" -DCMAKE_RANLIB="$RANLIB"
	-DCMAKE_C_COMPILER_AR="$AR" -DCMAKE_CXX_COMPILER_AR="$AR"
	-DCMAKE_C_COMPILER_RANLIB="$RANLIB" -DCMAKE_CXX_COMPILER_RANLIB="$RANLIB"
	-DCMAKE_BUILD_TYPE=Release -DCMAKE_INSTALL_PREFIX="$PREFIX"
	-DCMAKE_FIND_ROOT_PATH="$PREFIX"
	-DCMAKE_FIND_ROOT_PATH_MODE_PROGRAM=NEVER
	-DCMAKE_FIND_ROOT_PATH_MODE_LIBRARY=ONLY
	-DCMAKE_FIND_ROOT_PATH_MODE_INCLUDE=ONLY
	-DCMAKE_NO_SYSTEM_FROM_IMPORTED=ON
	-DCMAKE_C_FLAGS=-D_USE_MATH_DEFINES
)
# Every helper links stripped: a published helper carries no debug information or local symbol
# table. zig's -s strips ELF and PE outright; on Mach-O it keeps the local symbols, which the linker's -x drops.
# staticproof bounds each helper's size, so a link that ships debug information fails the build.
STRIP_LDFLAGS=(-s)
[ "$SYSNAME" = Darwin ] && STRIP_LDFLAGS+=(-Xlinker -x)

CXXFLAGS_ALL="-D_USE_MATH_DEFINES"
[ "$SYSNAME" = Darwin ] && CXXFLAGS_ALL="$CXXFLAGS_ALL -fno-cxx-modules"
CROSS+=(-DCMAKE_CXX_FLAGS="$CXXFLAGS_ALL")

# ---- Build dependencies -------------------------------------------------------------
HELPER_PINS="$NATIVE_DIR/$HELPER/PINS.txt"

stack_pin() { # stack_pin <library>: the file of the one <library>-<version>.tar.gz record in HELPER_PINS
	local files n
	files="$(awk -v n="$1" '$0 !~ /^#/ && index($2, n "-") == 1 && $2 ~ /\.tar\.gz$/ { v = substr($2, length(n) + 2); if (v ~ /^[0-9]/) print $2 }' "$HELPER_PINS")"
	n="$(printf '%s' "$files" | grep -c . || true)"
	[ "$n" = 1 ] || die "$HELPER_PINS holds $n $1-<version>.tar.gz records, want exactly one"
	echo "$files"
}
pin_version() { # pin_version <library> <file>
	local v="${2#"$1"-}"
	echo "${v%.tar.gz}"
}

if [ "$HELPER" = "feov-ocr" ]; then
	ZLIB_TAR="$(stack_pin zlib)"
	LIBPNG_TAR="$(stack_pin libpng)"
	LEPTONICA_TAR="$(stack_pin leptonica)"
	TESSERACT_TAR="$(stack_pin tesseract)"
	LEPTONICA_VERSION="$(pin_version leptonica "$LEPTONICA_TAR")"
	TESSERACT_VERSION="$(pin_version tesseract "$TESSERACT_TAR")"
	# tesseract names its Windows static library after its major and minor version: 5.5.3 builds
	# libtesseract55.a.
	TESSERACT_ABI="$(echo "$TESSERACT_VERSION" | cut -d. -f1,2 | tr -d .)"
	for f in "$ZLIB_TAR" "$LIBPNG_TAR" "$LEPTONICA_TAR" "$TESSERACT_TAR"; do
		fetch_pinned "$f" "$HELPER_PINS" "$WORK/src"
	done

	cd "$WORK/build"

	# zlib
	if [ ! -f "$PREFIX/lib/libz.a" ] && [ ! -f "$PREFIX/lib/libzlibstatic.a" ] && [ ! -f "$PREFIX/lib/libzs.a" ]; then
		echo "=== Building zlib ($TARGET) ==="
		rm -rf zlib && mkdir zlib
		tar -C zlib --strip-components=1 -xf "$WORK/src/$ZLIB_TAR"
		"$CMAKE" -S zlib -B zlib/b "${CROSS[@]}" \
			-DCMAKE_POLICY_VERSION_MINIMUM=3.5 \
			-DZLIB_BUILD_TESTING=OFF -DZLIB_BUILD_SHARED=OFF -DZLIB_BUILD_STATIC=ON \
			-DZLIB_BUILD_MINIZIP=OFF -DZLIB_INSTALL=ON
		"$CMAKE" --build zlib/b -j"$J"
		"$CMAKE" --install zlib/b
		if [ "$SYSNAME" = "Windows" ]; then
			( cd "$PREFIX/lib" && if [ -e libzs.a ] && [ ! -e libz.a ]; then ln -sf libzs.a libz.a; fi )
			( cd "$PREFIX/lib" && if [ -e libzlibstatic.a ] && [ ! -e libz.a ]; then ln -sf libzlibstatic.a libz.a; fi )
		fi
	fi

	# libpng
	if [ ! -f "$PREFIX/lib/libpng16.a" ] && [ ! -f "$PREFIX/lib/libpng.a" ]; then
		echo "=== Building libpng ($TARGET) ==="
		rm -rf libpng && mkdir libpng
		tar -C libpng --strip-components=1 -xf "$WORK/src/$LIBPNG_TAR"
		"$CMAKE" -S libpng -B libpng/b "${CROSS[@]}" \
			-DPNG_SHARED=OFF -DPNG_STATIC=ON -DPNG_TESTS=OFF -DPNG_TOOLS=OFF \
			-DPNG_FRAMEWORK=OFF -DPNG_DEBUG=OFF \
			-DZLIB_ROOT="$PREFIX" -DZLIB_USE_STATIC_LIBS=ON
		"$CMAKE" --build libpng/b -j"$J"
		"$CMAKE" --install libpng/b
	fi

	# leptonica
	if [ ! -f "$PREFIX/lib/libleptonica.a" ] && [ ! -f "$PREFIX/lib/libleptonica-${LEPTONICA_VERSION}.dll" ] && [ ! -f "$PREFIX/lib/liblept.a" ]; then
		echo "=== Building leptonica ($TARGET) ==="
		rm -rf leptonica && mkdir leptonica
		tar -C leptonica --strip-components=1 -xf "$WORK/src/$LEPTONICA_TAR"
		"$CMAKE" -S leptonica -B leptonica/b "${CROSS[@]}" \
			-DCMAKE_PREFIX_PATH="$PREFIX" \
			-DBUILD_SHARED_LIBS=OFF -DBUILD_PROG=OFF \
			-DENABLE_PNG=ON -DENABLE_ZLIB=ON \
			-DENABLE_GIF=OFF -DENABLE_JPEG=OFF -DENABLE_TIFF=OFF \
			-DENABLE_WEBP=OFF -DENABLE_OPENJPEG=OFF -DSW_BUILD=OFF \
			-DZLIB_ROOT="$PREFIX" -DPNG_ROOT="$PREFIX" -DZLIB_USE_STATIC_LIBS=ON
		"$CMAKE" --build leptonica/b -j"$J"
		"$CMAKE" --install leptonica/b
	fi

	# tesseract
	if [ ! -f "$PREFIX/lib/libtesseract.a" ] && [ ! -f "$PREFIX/lib/libtesseract${TESSERACT_ABI}.a" ]; then
		echo "=== Building tesseract ($TARGET) ==="
		rm -rf tesseract && mkdir tesseract
		tar -C tesseract --strip-components=1 -xf "$WORK/src/$TESSERACT_TAR"
		sed -i 's/set(LIB_Ws2_32 Ws2_32)/set(LIB_Ws2_32 ws2_32)/' tesseract/CMakeLists.txt
		"$CMAKE" -S tesseract -B tesseract/b "${CROSS[@]}" \
			-DCMAKE_PREFIX_PATH="$PREFIX" \
			-DBUILD_SHARED_LIBS=OFF \
			-DBUILD_TRAINING_TOOLS=OFF -DBUILD_TESTS=OFF \
			-DDISABLE_CURL=ON -DDISABLE_ARCHIVE=ON -DDISABLE_TIFF=ON \
			-DGRAPHICS_DISABLED=ON -DDISABLED_LEGACY_ENGINE=ON \
			-DOPENMP_BUILD=OFF -DSW_BUILD=OFF -DENABLE_LTO=OFF \
			-DUSE_SYSTEM_ICU=OFF -DLEPT_TIFF_RESULT=1
		"$CMAKE" --build tesseract/b -j"$J"
		"$CMAKE" --install tesseract/b
	fi

	# Fixups
	case "$SYSNAME" in
	Windows)
		cd "$PREFIX/lib"
		alias_lib() { if [ -e "$1" ] && [ ! -e "$2" ]; then ln -sf "$1" "$2"; fi; }
		alias_lib libzs.a libz.a
		alias_lib libzlibstatic.a libz.a
		alias_lib "libleptonica-${LEPTONICA_VERSION}.dll" libleptonica.a
		alias_lib "libtesseract${TESSERACT_ABI}.a" libtesseract.a
		;;
	esac

	# ---- Compile feov-ocr ---------------------------------------------------------------
	echo "=== Compiling feov-ocr ($TARGET) ==="
	cd "$WORK/build"

	# Embed eng.traineddata
	TRAINEDDATA_SRC="$NATIVE_DIR/feov-ocr/eng.traineddata"
	[ -f "$TRAINEDDATA_SRC" ] || die "missing $TRAINEDDATA_SRC"

	python3 -c "
with open('$TRAINEDDATA_SRC', 'rb') as f:
    data = f.read()
with open('traineddata.c', 'w') as f:
    f.write('#include <stddef.h>\nconst unsigned char g_eng_traineddata[] = {\n')
    for i in range(0, len(data), 32):
        f.write(','.join(str(b) for b in data[i:i+32]) + ',\n')
    f.write('};\nconst size_t g_eng_traineddata_len = sizeof(g_eng_traineddata);\n')
"
	"$CC" -O2 -c -o traineddata.o traineddata.c

	# Compile C++ objects
	CXXFLAGS=(-std=c++17 -O2 -DFEOV_HELPER_DIGEST="\"$DIGEST\"" -D_USE_MATH_DEFINES)
	CXXFLAGS+=(-I"$PREFIX/include" -I"$NATIVE_DIR/frame" -I"$NATIVE_DIR/feov-ocr")
	[ "$SYSNAME" = Darwin ] && CXXFLAGS+=(-fno-cxx-modules)

	"$CXX" "${CXXFLAGS[@]}" -c -o frame.o "$NATIVE_DIR/frame/frame.cpp"
	"$CXX" "${CXXFLAGS[@]}" -c -o shim.o "$NATIVE_DIR/feov-ocr/shim.cpp"
	"$CXX" "${CXXFLAGS[@]}" -c -o main.o "$NATIVE_DIR/feov-ocr/main.cpp"

	LDFLAGS=("${STRIP_LDFLAGS[@]}" -L"$PREFIX/lib" -ltesseract -lleptonica -lpng16 -lz -lc++)
	case "$SYSNAME" in
	Linux)
		LDFLAGS+=(-static)
		;;
	Windows)
		LDFLAGS+=(-static -lws2_32)
		;;
	esac

	OUT_NAME="${HELPER}_${OS}_${ARCH}${EXT}"
	"$CXX" -o "$OUT_NAME" main.o shim.o frame.o traineddata.o "${LDFLAGS[@]}"

	# Run staticproof check on built helper
	if [ -f "$REPO_ROOT/scripts/staticproof/main.go" ]; then
		go -C "$REPO_ROOT/scripts" run ./staticproof -artifact "$HELPER" "$PWD/$OUT_NAME"
	fi

	cp "$OUT_NAME" "$OUTDIR/$OUT_NAME"
	echo "native/build.sh: $OUTDIR/$OUT_NAME ready"
fi

if [ "$HELPER" = "feov-embed" ]; then
	fetch_assets

	if [ ! -f "$PREFIX/lib/libonnxruntime_static.a" ]; then
		echo "=== Building ONNX Runtime ($TARGET) ==="
		mkdir -p "$PREFIX/lib" "$PREFIX/include"
		fetch_pinned onnxruntime-1.31.1.tar.gz "$HELPER_PINS" "$WORK/src"
		fetch_pinned "$FLATBUFFERS_WHEEL" "$HELPER_PINS" "$WORK/src"
		ORT="$WORK/build/ort"
		rm -rf "$ORT" && mkdir -p "$ORT"
		tar -C "$ORT" --strip-components=1 -xf "$WORK/src/onnxruntime-1.31.1.tar.gz"
		ort_mirror_unhashed "$ORT"
		ort_patch() { # ort_patch <file> <sed expr>: a pattern that no longer matches the pinned source fails the build
			local before
			before="$(cksum < "$ORT/$1")"
			sed -i -e "$2" "$ORT/$1"
			[ "$(cksum < "$ORT/$1")" != "$before" ] || die "patching $1 changed nothing: $2"
		}
		# mlas's q4dq command-line tool is an executable in ALL; on darwin its link asks for
		# Foundation, which zig cannot supply. Nothing here needs the tool, and without it nothing in
		# ALL builds cpuinfo, which is built by name below.
		ort_patch cmake/onnxruntime_mlas.cmake 's/^  set_target_properties(onnxruntime_mlas_q4dq PROPERTIES FOLDER "ONNXRuntimeTest")$/&\n  set_target_properties(onnxruntime_mlas_q4dq PROPERTIES EXCLUDE_FROM_ALL TRUE)/'
		ORT_CC="$WORK/bin/ortcc" ORT_CXX="$WORK/bin/ortcxx"
		ORT_DEFS=(onnxruntime_BUILD_UNIT_TESTS=OFF CMAKE_POLICY_VERSION_MINIMUM=3.5
			CMAKE_SYSTEM_NAME="$SYSNAME" CMAKE_SYSTEM_PROCESSOR="$PROC"
			CMAKE_C_COMPILER="$ORT_CC" CMAKE_CXX_COMPILER="$ORT_CXX" CMAKE_ASM_COMPILER="$ORT_CC"
			CMAKE_AR="$AR" CMAKE_RANLIB="$RANLIB"
			CMAKE_C_COMPILER_AR="$AR" CMAKE_CXX_COMPILER_AR="$AR"
			CMAKE_C_COMPILER_RANLIB="$RANLIB" CMAKE_CXX_COMPILER_RANLIB="$RANLIB")
		# zig cc maps -march to a CPU name and rejects ORT's per-file -march=armv8.2-a+bf16 (and
		# the CMakeLists.txt probe of it) and KleidiAI's plain -march=armv8-a; the same request
		# spelled as zig CPU features is -mcpu=generic+v8_2a+bf16, where fp16 is "fullfp16".
		cat > "$WORK/bin/marchfix" <<'EOS'
for a; do shift; case $a in
  -march=armv8-a*) r=$(printf '%s' "${a#-march=armv8-a}" | sed 's/+fp16/+fullfp16/g'); set -- "$@" "-mcpu=generic$r" ;;
  -march=armv8.[0-9]-a*) v=${a#-march=armv8.}; n=${v%%-a*}; r=$(printf '%s' "${v#*-a}" | sed 's/+fp16/+fullfp16/g'); set -- "$@" "-mcpu=generic+v8_${n}a$r" ;;
  *) set -- "$@" "$a" ;; esac; done
EOS
		ORT_INC=""
		if [ "$SYSNAME" = Darwin ]; then
			# The Apple log sink is Objective-C++ over Foundation (NSLog); zig carries no macOS SDK
			# and Foundation is outside the darwin allowlist, so the portable clog sink is used and
			# device discovery (IOKit) is off.
			ort_patch cmake/onnxruntime_common.cmake '/core\/platform\/apple\/logging\/\*\.h"/,/core\/platform\/apple\/logging\/\*\.mm"/d'
			ort_patch onnxruntime/core/platform/logging/make_platform_default_log_sink.cc 's/^#elif defined(__APPLE__)$/#elif defined(__APPLE__) \&\& 0/'
			# abseil's cctz asks CoreFoundation for the local zone name on __APPLE__: the six
			# declarations it uses, with Apple's signatures; the link resolves them against
			# CoreFoundation. ORT's posix env.cc includes <sys/syscall.h>, which zig's macOS headers
			# lack, and uses it only on Linux.
			ORT_INC="$WORK/ortinc"
			mkdir -p "$ORT_INC/CoreFoundation" "$ORT_INC/sys"
			cat > "$ORT_INC/CoreFoundation/CFTimeZone.h" <<'EOS'
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
			printf '#pragma once\n' > "$ORT_INC/sys/syscall.h"
			# ORT links Iconv on every non-Windows target and calls none of it; macOS keeps iconv
			# outside libSystem and zig has no SDK to find it in. mlas and the platform probe pick
			# kernels by CMAKE_OSX_ARCHITECTURES on Apple; unset, they pick the host's (x86).
			ORT_DEFS+=(onnxruntime_DISABLE_DEVICE_DISCOVERY=ON Iconv_IS_BUILT_IN=TRUE CMAKE_OSX_ARCHITECTURES="$PROC")
		fi
		for tool in cc c++; do
			{
				printf '#!/bin/sh\n'
				cat "$WORK/bin/marchfix"
				if [ -n "$ORT_INC" ]; then
					printf 'exec %q %s -target %s -isystem %q "$@"\n' "$ZIG" "$tool" "$TRIPLE" "$ORT_INC"
				else
					printf 'exec %q %s -target %s "$@"\n' "$ZIG" "$tool" "$TRIPLE"
				fi
			} > "$WORK/bin/ort${tool/c++/cxx}"
			chmod +x "$WORK/bin/ort${tool/c++/cxx}"
		done
		ORT_ARGS=(--cmake_path "$CMAKE" --ctest_path "$(dirname "$CMAKE")/ctest" \
			--build_dir "$ORT/b" \
			--config Release \
			--minimal_build --disable_exceptions --disable_rtti --disable_ml_ops \
			--include_ops_by_config "$NATIVE_DIR/feov-embed/ops.config" \
			--enable_reduced_operator_type_support \
			--skip_tests --skip_submodule_sync --compile_no_warning_as_error \
			--parallel "$J" \
			--cmake_extra_defines "${ORT_DEFS[@]}")
		PYTHONPATH="$WORK/src/$FLATBUFFERS_WHEEL" python3 "$ORT/tools/ci_build/build.py" --update "${ORT_ARGS[@]}"
		ort_check_fetches "$ORT/b/Release"
		PYTHONPATH="$WORK/src/$FLATBUFFERS_WHEEL" python3 "$ORT/tools/ci_build/build.py" --build "${ORT_ARGS[@]}"
		"$CMAKE" --build "$ORT/b/Release" --config Release --target cpuinfo -j "$J"
		ort_lib_out="$PREFIX/lib/libonnxruntime_static.a"
		rm -f "$ort_lib_out"
		libs=()
		while IFS= read -r l; do libs+=("$l"); done < <(find "$WORK/build/ort/b/Release" \( -name '*.a' -o -name '*.lib' \) \
			! -name 'libprotoc.a' ! -name 'libprotobuf.a' ! -name 'libupb*.a' ! -name 'libflatbuffers.a' \
			! -name 'protoc*.lib' ! -name 'libprotobuf.lib' ! -name 'libprotobufd.lib' ! -name 'flatbuffers.lib' ! -name 'upb*.lib' \
			! -name 'libonnxruntime_static.a' ! -path '*gtest*' ! -path '*gmock*' ! -path '*benchmark*' ! -name '*test*' | sort)
		case "$(uname -s)" in
		Darwin) if [ "$AR" = ar ]; then libtool -static -o "$ort_lib_out" "${libs[@]}"; else { echo "CREATE $ort_lib_out"; for l in "${libs[@]}"; do echo "ADDLIB $l"; done; echo SAVE; echo END; } | "$AR" -M; fi ;;
		*) { echo "CREATE $ort_lib_out"; for l in "${libs[@]}"; do echo "ADDLIB $l"; done; echo SAVE; echo END; } | "$AR" -M ;;
		esac
		cp "$WORK/build/ort"/include/onnxruntime/core/session/onnxruntime_*.h "$PREFIX/include/"
	fi

	# Fixups for Darwin
	case "$SYSNAME" in
	Darwin)
		GOSRC="$(go env GOROOT)/src"
		tbd_target="$PROC-macos"
		dyn_syms() { # dyn_syms <library> [symbol...]: what Go imports from it, and the symbols named
			local lib="$1"
			shift
			{
				grep -rhoE "//go:cgo_import_dynamic [^ ]+ [^ ]+ \"[^\"]*$lib\"" "$GOSRC" | awk '{ print "_" $3 }'
				[ $# -eq 0 ] || printf '%s\n' "$@"
			} | sort -u | paste -sd, - | sed 's/,/, /g'
		}
		write_tbd() {
			mkdir -p "$(dirname "$1")"
			{
				printf -- "--- !tapi-tbd\ntbd-version: 4\ntargets: [ %s ]\ninstall-name: '%s'\ncurrent-version: 1.0\n" "$tbd_target" "$2"
				[ -n "${3:-}" ] && printf -- "exports:\n  - targets: [ %s ]\n    symbols: [ %s ]\n" "$tbd_target" "$3"
				printf -- '...\n'
			} > "$1"
		}
		# abseil cctz, inside ONNX Runtime, calls these six CoreFoundation functions.
		cf_syms="$(dyn_syms CoreFoundation _CFTimeZoneCopyDefault _CFTimeZoneGetName _CFStringGetLength \
			_CFStringGetMaximumSizeForEncoding _CFStringGetCString _CFRelease)"
		sec_syms="$(dyn_syms Security)"
		write_tbd "$PREFIX/lib/libresolv.tbd" '/usr/lib/libresolv.9.dylib'
		write_tbd "$PREFIX/Frameworks/CoreFoundation.framework/CoreFoundation.tbd" \
			'/System/Library/Frameworks/CoreFoundation.framework/Versions/A/CoreFoundation' "$cf_syms"
		write_tbd "$PREFIX/Frameworks/Security.framework/Security.tbd" \
			'/System/Library/Frameworks/Security.framework/Versions/A/Security' "$sec_syms"
		;;
	esac

	# ---- Compile feov-embed -------------------------------------------------------------
	echo "=== Compiling feov-embed ($TARGET) ==="
	cd "$WORK/build"

	# Compile C++ objects
	CXXFLAGS=(-std=c++17 -O2 -DFEOV_HELPER_DIGEST="\"$DIGEST\"" -D_USE_MATH_DEFINES)
	CXXFLAGS+=(-I"$PREFIX/include" -I"$NATIVE_DIR/frame" -I"$NATIVE_DIR/feov-embed")
	[ "$SYSNAME" = Darwin ] && CXXFLAGS+=(-fno-cxx-modules)

	"$CXX" "${CXXFLAGS[@]}" -c -o frame.o "$NATIVE_DIR/frame/frame.cpp"
	"$CXX" "${CXXFLAGS[@]}" -c -o main.o "$NATIVE_DIR/feov-embed/main.cpp"

	LDFLAGS=("${STRIP_LDFLAGS[@]}" -L"$PREFIX/lib" -lonnxruntime_static -lc++)
	case "$SYSNAME" in
	Linux)
		LDFLAGS+=(-static)
		;;
	Windows)
		LDFLAGS+=(-static -lws2_32)
		;;
	Darwin)
		LDFLAGS+=(-F"$PREFIX/Frameworks" -framework CoreFoundation -framework Security -lresolv)
		;;
	esac

	OUT_NAME="${HELPER}_${OS}_${ARCH}${EXT}"
	"$CXX" -o "$OUT_NAME" main.o frame.o "${LDFLAGS[@]}"

	# Run staticproof check on built helper
	if [ -f "$REPO_ROOT/scripts/staticproof/main.go" ]; then
		go -C "$REPO_ROOT/scripts" run ./staticproof -artifact "$HELPER" "$PWD/$OUT_NAME"
	fi

	cp "$OUT_NAME" "$OUTDIR/$OUT_NAME"
	ship_assets
	echo "native/build.sh: $OUTDIR/$OUT_NAME ready"
fi
