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
	WORK="${RUNNER_TEMP:-${XDG_CACHE_HOME:-$HOME/.cache}/sc-native}/$HELPER/$TARGET"
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
			curl -fsSL --retry 5 --retry-connrefused --retry-delay 2 --connect-timeout 30 -o "$dest.part" "$url" && mv "$dest.part" "$dest"
		fi
	fi
	echo "$sha  $dest" | sha256sum -c --quiet - || die "$file does not match its pin"
}

# Windows MSVC native helper build for feov-embed (plans/native-helpers.md §III.1.5)
if [ "$HELPER" = "feov-embed" ] && [ "$OS" = "windows" ]; then
	HELPER_PINS="$NATIVE_DIR/$HELPER/PINS.txt"
	model_file="gte-small.int8.${ARCH}.ort"
	fetch_pinned "$model_file" "$HELPER_PINS" "$WORK/src"

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

	HP=python
	command -v python >/dev/null 2>&1 || HP=python3
	[ -d "$WORK/pyenv" ] || "$HP" -m venv "$WORK/pyenv"
	if [ -f "$WORK/pyenv/Scripts/python.exe" ]; then
		PY="$WORK/pyenv/Scripts/python.exe"
	else
		PY="$WORK/pyenv/bin/python"
	fi
	"$PY" -m pip install -q flatbuffers

	fetch_pinned onnxruntime-1.31.1.tar.gz "$HELPER_PINS" "$WORK/src"
	SRC="$WORK/src/onnxruntime"
	if [ ! -d "$SRC" ]; then
		mkdir -p "$SRC"
		tar -C "$SRC" --strip-components=1 -xf "$WORK/src/onnxruntime-1.31.1.tar.gz"
	fi

	if [ ! -f "$WORK/build/ort/Release/onnxruntime_common.lib" ]; then
		echo "=== Building ONNX Runtime ($TARGET) with MSVC ==="
		XARGS=()
		MACH=$("$PY" -c 'import platform; print(platform.machine())' | tr -d '\r')
		[ "$VS_ARCH" = "ARM64" ] && [ "$MACH" != "ARM64" ] && XARGS+=(--arm64)
		"$PY" "$SRC/tools/ci_build/build.py" \
			--build_dir "$WORK/build/ort" \
			--config Release \
			--update --build --parallel "$J" \
			--skip_tests --skip_submodule_sync --compile_no_warning_as_error \
			--cmake_generator "$GEN" \
			--minimal_build --disable_exceptions --disable_rtti --disable_ml_ops \
			--include_ops_by_config "$NATIVE_DIR/feov-embed/ops.config" \
			--enable_reduced_operator_type_support \
			--enable_msvc_static_runtime ${XARGS[@]+"${XARGS[@]}"} \
			--cmake_extra_defines onnxruntime_BUILD_UNIT_TESTS=OFF onnxruntime_DISABLE_DEVICE_DISCOVERY=ON
		cmake --build "$WORK/build/ort/Release" --config Release --target cpuinfo -j "$J" || true
	fi

	find "$WORK/build/ort/Release" -name '*.lib' \
		! -name 'protoc*.lib' ! -name 'libprotoc*.lib' ! -name 'libprotobuf.lib' ! -name 'libprotobufd.lib' \
		! -name 'flatbuffers.lib' ! -name 'upb*.lib' ! -path '*gtest*' ! -path '*gmock*' ! -path '*benchmark*' ! -name '*test*' \
		! -path '*/feov-embed/*' | sort > "$WORK/libs-unix.txt"
	while IFS= read -r l; do cygpath -m "$l"; done < "$WORK/libs-unix.txt" | tr -d '\r' > "$WORK/libs.txt"

	LIB_EXE="$("$VSW" -latest -products '*' -find '**/Hostx64/x64/lib.exe' | head -n 1 | tr -d '\r')"
	if [ -z "$LIB_EXE" ] || [ ! -f "$LIB_EXE" ]; then
		LIB_EXE="$(find "$VSDIR/VC/Tools/MSVC" -name "lib.exe" 2>/dev/null | grep "Hostx64/x64/lib.exe" | head -n 1 | tr -d '\r')"
	fi
	[ -n "$LIB_EXE" ] && [ -f "$LIB_EXE" ] || die "lib.exe not found in $VSDIR"

	ort_static_lib="$WORK/build/ort/Release/onnxruntime_static.lib"
	rm -f "$ort_static_lib"
	ort_static_lib_win="$(cygpath -w "$ort_static_lib" | tr -d '\r')"
	{
		echo "/NOLOGO"
		echo "/OUT:$ort_static_lib_win"
		while IFS= read -r l; do
			cygpath -w "$l" | tr -d '\r'
		done < "$WORK/libs-unix.txt"
	} > "$WORK/libs-rsp.txt"

	rsp_win="$(cygpath -w "$WORK/libs-rsp.txt" | tr -d '\r')"
	echo "=== Merging ONNX Runtime static libraries with lib.exe ==="
	"$LIB_EXE" @"$rsp_win"
	[ -f "$ort_static_lib" ] || die "lib.exe failed to create $ort_static_lib"

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
	cp "$EMBED_DIR/b/Release/feov-embed.exe" "$OUTDIR/${HELPER}${EXT}"
	cp "$WORK/src/$model_file" "$OUTDIR/gte-small.int8.ort"
	cp "$WORK/src/$model_file" "$OUTDIR/$model_file"
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
CXXFLAGS_ALL="-D_USE_MATH_DEFINES"
[ "$SYSNAME" = Darwin ] && CXXFLAGS_ALL="$CXXFLAGS_ALL -fno-cxx-modules"
CROSS+=(-DCMAKE_CXX_FLAGS="$CXXFLAGS_ALL")

# ---- Build dependencies -------------------------------------------------------------
HELPER_PINS="$NATIVE_DIR/$HELPER/PINS.txt"

if [ "$HELPER" = "feov-ocr" ]; then
	for f in zlib-1.3.2.tar.gz libpng-1.6.58.tar.gz leptonica-1.87.0.tar.gz tesseract-5.5.3.tar.gz; do
		fetch_pinned "$f" "$HELPER_PINS" "$WORK/src"
	done

	cd "$WORK/build"

	# zlib
	if [ ! -f "$PREFIX/lib/libz.a" ] && [ ! -f "$PREFIX/lib/libzlibstatic.a" ] && [ ! -f "$PREFIX/lib/libzs.a" ]; then
		echo "=== Building zlib ($TARGET) ==="
		rm -rf zlib && mkdir zlib
		tar -C zlib --strip-components=1 -xf "$WORK/src/zlib-1.3.2.tar.gz"
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
		tar -C libpng --strip-components=1 -xf "$WORK/src/libpng-1.6.58.tar.gz"
		"$CMAKE" -S libpng -B libpng/b "${CROSS[@]}" \
			-DPNG_SHARED=OFF -DPNG_STATIC=ON -DPNG_TESTS=OFF -DPNG_TOOLS=OFF \
			-DPNG_FRAMEWORK=OFF -DPNG_DEBUG=OFF \
			-DZLIB_ROOT="$PREFIX" -DZLIB_USE_STATIC_LIBS=ON
		"$CMAKE" --build libpng/b -j"$J"
		"$CMAKE" --install libpng/b
	fi

	# leptonica
	if [ ! -f "$PREFIX/lib/libleptonica.a" ] && [ ! -f "$PREFIX/lib/libleptonica-1.87.0.dll" ] && [ ! -f "$PREFIX/lib/liblept.a" ]; then
		echo "=== Building leptonica ($TARGET) ==="
		rm -rf leptonica && mkdir leptonica
		tar -C leptonica --strip-components=1 -xf "$WORK/src/leptonica-1.87.0.tar.gz"
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
	if [ ! -f "$PREFIX/lib/libtesseract.a" ] && [ ! -f "$PREFIX/lib/libtesseract55.a" ]; then
		echo "=== Building tesseract ($TARGET) ==="
		rm -rf tesseract && mkdir tesseract
		tar -C tesseract --strip-components=1 -xf "$WORK/src/tesseract-5.5.3.tar.gz"
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
		alias_lib libleptonica-1.87.0.dll libleptonica.a
		alias_lib libtesseract55.a libtesseract.a
		;;
	Darwin)
		GOSRC="$(go env GOROOT)/src"
		tbd_target="$PROC-macos"
		dyn_syms() {
			grep -rhoE "//go:cgo_import_dynamic [^ ]+ [^ ]+ \"[^\"]*$1\"" "$GOSRC" |
				awk '{ print "_" $3 }' | sort -u | paste -sd, - | sed 's/,/, /g'
		}
		write_tbd() {
			mkdir -p "$(dirname "$1")"
			{
				printf -- "--- !tapi-tbd\ntbd-version: 4\ntargets: [ %s ]\ninstall-name: '%s'\ncurrent-version: 1.0\n" "$tbd_target" "$2"
				[ -n "${3:-}" ] && printf -- "exports:\n  - targets: [ %s ]\n    symbols: [ %s ]\n" "$tbd_target" "$3"
				printf -- '...\n'
			} > "$1"
		}
		cf_syms="$(dyn_syms CoreFoundation)"
		sec_syms="$(dyn_syms Security)"
		write_tbd "$PREFIX/lib/libresolv.tbd" '/usr/lib/libresolv.9.dylib'
		write_tbd "$PREFIX/Frameworks/CoreFoundation.framework/CoreFoundation.tbd" \
			'/System/Library/Frameworks/CoreFoundation.framework/Versions/A/CoreFoundation' "$cf_syms"
		write_tbd "$PREFIX/Frameworks/Security.framework/Security.tbd" \
			'/System/Library/Frameworks/Security.framework/Versions/A/Security' "$sec_syms"
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

	LDFLAGS=(-L"$PREFIX/lib" -ltesseract -lleptonica -lpng16 -lz -lc++)
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
	"$CXX" -o "$OUT_NAME" main.o shim.o frame.o traineddata.o "${LDFLAGS[@]}"

	# Run staticproof check on built helper
	if [ -f "$REPO_ROOT/scripts/staticproof/main.go" ]; then
		go -C "$REPO_ROOT/scripts" run ./staticproof -artifact "$HELPER" "$PWD/$OUT_NAME"
	fi

	cp "$OUT_NAME" "$OUTDIR/$OUT_NAME"
	cp "$OUT_NAME" "$OUTDIR/${HELPER}${EXT}"
	echo "native/build.sh: $OUTDIR/$OUT_NAME ready"
fi

if [ "$HELPER" = "feov-embed" ]; then
	model_file="gte-small.int8.${ARCH}.ort"
	fetch_pinned "$model_file" "$HELPER_PINS" "$WORK/src"

	if [ ! -f "$PREFIX/lib/libonnxruntime_static.a" ]; then
		echo "=== Providing ONNX Runtime ($TARGET) ==="
		mkdir -p "$PREFIX/lib" "$PREFIX/include"
		if [ -n "${ORT_STATIC_DIR:-}" ] && [ -f "$ORT_STATIC_DIR/lib/libonnxruntime_static.a" ]; then
			cp "$ORT_STATIC_DIR"/include/* "$PREFIX/include/"
			cp "$ORT_STATIC_DIR"/lib/* "$PREFIX/lib/"
		elif [ -f "$HOME/.claude/scratch/vecplan/ortstatic/build/${TARGET}-zig-min/out/lib/libonnxruntime_static.a" ]; then
			cp "$HOME/.claude/scratch/vecplan/ortstatic/build/${TARGET}-zig-min/out/include/"* "$PREFIX/include/"
			cp "$HOME/.claude/scratch/vecplan/ortstatic/build/${TARGET}-zig-min/out/lib/"* "$PREFIX/lib/"
		elif [ -f "$HOME/.claude/scratch/vecplan/ortstatic/out/min/lib/libonnxruntime_static.a" ]; then
			cp "$HOME/.claude/scratch/vecplan/ortstatic/out/min/include/"* "$PREFIX/include/"
			cp "$HOME/.claude/scratch/vecplan/ortstatic/out/min/lib/"* "$PREFIX/lib/"
		else
			fetch_pinned onnxruntime-1.31.1.tar.gz "$HELPER_PINS" "$WORK/src"
			echo "=== Building ONNX Runtime ($TARGET) ==="
			rm -rf "$WORK/build/ort" && mkdir -p "$WORK/build/ort"
			tar -C "$WORK/build/ort" --strip-components=1 -xf "$WORK/src/onnxruntime-1.31.1.tar.gz"
			python3 -m pip install flatbuffers 2>/dev/null || pip3 install flatbuffers 2>/dev/null || pip install flatbuffers 2>/dev/null || python3 -m pip install --break-system-packages flatbuffers 2>/dev/null || true
			python3 "$WORK/build/ort/tools/ci_build/build.py" \
				--build_dir "$WORK/build/ort/b" \
				--config Release \
				--minimal_build --disable_exceptions --disable_rtti --disable_ml_ops \
				--include_ops_by_config "$NATIVE_DIR/feov-embed/ops.config" \
				--enable_reduced_operator_type_support \
				--skip_tests --skip_submodule_sync --compile_no_warning_as_error \
				--parallel "$J" \
				--cmake_extra_defines CMAKE_SYSTEM_NAME="$SYSNAME" CMAKE_SYSTEM_PROCESSOR="$PROC" \
					CMAKE_C_COMPILER="$CC" CMAKE_CXX_COMPILER="$CXX" \
					CMAKE_AR="$AR" CMAKE_RANLIB="$RANLIB"
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
	fi

	# Fixups for Darwin
	case "$SYSNAME" in
	Darwin)
		GOSRC="$(go env GOROOT)/src"
		tbd_target="$PROC-macos"
		dyn_syms() {
			grep -rhoE "//go:cgo_import_dynamic [^ ]+ [^ ]+ \"[^\"]*$1\"" "$GOSRC" |
				awk '{ print "_" $3 }' | sort -u | paste -sd, - | sed 's/,/, /g'
		}
		write_tbd() {
			mkdir -p "$(dirname "$1")"
			{
				printf -- "--- !tapi-tbd\ntbd-version: 4\ntargets: [ %s ]\ninstall-name: '%s'\ncurrent-version: 1.0\n" "$tbd_target" "$2"
				[ -n "${3:-}" ] && printf -- "exports:\n  - targets: [ %s ]\n    symbols: [ %s ]\n" "$tbd_target" "$3"
				printf -- '...\n'
			} > "$1"
		}
		cf_syms="$(dyn_syms CoreFoundation)"
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

	LDFLAGS=(-L"$PREFIX/lib" -lonnxruntime_static -lc++)
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
	cp "$OUT_NAME" "$OUTDIR/${HELPER}${EXT}"
	cp "$WORK/src/$model_file" "$OUTDIR/gte-small.int8.ort"
	cp "$WORK/src/$model_file" "$OUTDIR/$model_file"
	echo "native/build.sh: $OUTDIR/$OUT_NAME ready"
fi
