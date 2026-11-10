#!/bin/bash
# build.sh — build one native helper for one target.
#
# Usage:
#   build.sh <helper> <target> [outdir] [workdir]
#
# Helpers: feov-ocr
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
if ! flock -n 9; then
	echo "native/build.sh: waiting for the lock on $WORK (another build is running)"
	flock 9
fi

J="${CSTACK_JOBS:-$(nproc 2>/dev/null || sysctl -n hw.ncpu 2>/dev/null || echo 4)}"

# ---- pin manifest access ------------------------------------------------------------
pin_field() { # pin_field <file> <manifest> <n: 1=sha 3=url>
	awk -v f="$1" -v n="$3" '$0 !~ /^#/ && $2 == f { print $n }' "$2"
}

fetch_pinned() { # fetch_pinned <file> <manifest> <destdir>
	local file="$1" manifest="$2" dest="$3/$1" sha url
	sha="$(pin_field "$file" "$manifest" 1)"
	url="$(pin_field "$file" "$manifest" 3)"
	[ -n "$sha" ] && [ -n "$url" ] || die "$file is not pinned in $manifest"
	if [ ! -f "$dest" ]; then
		echo "fetching $file"
		curl -fsSL --retry 5 --retry-all-errors --connect-timeout 20 --max-time 600 -o "$dest.part" "$url" ||
			die "downloading $file from $url failed"
		mv "$dest.part" "$dest"
	fi
	echo "$sha  $dest" | sha256sum -c --quiet - || die "$file does not match its pin"
}

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
		alias_lib libleptonica-${LEPTONICA_VERSION}.dll libleptonica.a
		alias_lib libtesseract${TESSERACT_ABI}.a libtesseract.a
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
