# The C stack, and why it is built with zig

`feov-record` links a static C/C++ stack: tesseract and leptonica (with zlib and libpng) for local OCR, behind the `tessocr` build tag. Every release target gets one self-contained binary with no runtime library to install.

| Piece | Where |
|---|---|
| Pinned sources and toolchain (zig, cmake), each by sha256 | `plugins/frank-exchange-of-views/tools/third_party/pins/PINS.txt` |
| Build for one target; print its cgo env; check a binary | `plugins/frank-exchange-of-views/tools/third_party/pins/build-cstack.sh` |
| Staticness check (ELF, PE and Mach-O import records) | `plugins/frank-exchange-of-views/tools/third_party/pins/staticproof/` |
| Release builds, per target | `.github/workflows/hooks.yml` (the release job) |

## How a target is built

One linux/x86_64 host builds all six targets: linux-{amd64,arm64}, windows-{amd64,arm64} and darwin-{amd64,arm64}. `build-cstack.sh <target> <workdir>` compiles the stack with `zig cc` / `zig c++ -target <triple>`:

| Target | Triple |
|---|---|
| linux-amd64 / linux-arm64 | `x86_64-linux-musl` / `aarch64-linux-musl` |
| windows-amd64 / windows-arm64 | `x86_64-windows-gnu` / `aarch64-windows-gnu` |
| darwin-amd64 / darwin-arm64 | `x86_64-macos` / `aarch64-macos` |

`build-cstack.sh env <target> <workdir>` prints the `CGO_*` exports for that prefix. The Go build then links with `-linkmode external`, with zig as the external linker, for every command, so all commands go through the same link that `staticproof` checks. `build-cstack.sh check <binary>` runs that check.

## Why zig, and not plain clang or gcc

`zig cc` and `zig c++` are clang and lld. Zig contributes the rest of what a cross-build needs, in one pinned download:

- **A libc for every target.** musl for Linux, mingw-w64 for Windows, and stub libraries for macOS. `zig c++ -target aarch64-windows-gnu` works from a Linux host with nothing else installed.
- **The C++ runtime and compiler support.** libc++, libunwind and compiler-rt, built for each target on demand. That is what makes the result fully static.
- **macOS without Apple's SDK.** Zig links against `.tbd` stubs, which describe a system library's exported symbols. `build-cstack.sh` writes the three it needs (libresolv, CoreFoundation, Security). Apple's SDK licence permits its use only on Apple hardware, so a Linux host cannot use it.

Plain clang can target all six triples too, but needs a sysroot for each one that we would fetch, pin and keep consistent ourselves:

- **Windows:** mingw-w64, or Microsoft's SDK, whose licence restricts redistribution.
- **macOS:** Apple's SDK, which runs into the licence above. osxcross, the usual workaround, still needs the real SDK.
- **Linux arm64:** a cross glibc or musl.

GCC needs a separate cross-toolchain built per target, and cannot target macOS at all.

Without a cross-compiler, the alternative is a native runner per operating system in the release job: macOS and Windows runners, each with its own toolchain and its own idea of static linking.

**The cost:** zig is pre-1.0, and its toolchain has quirks. Each workaround in `build-cstack.sh` is commented with the failure it prevents. Read those comments before changing a flag. Bump zig's pin in `PINS.txt` deliberately, and rebuild all six targets when you do.

## Adding a C or C++ library

A library joins the stack only if it builds under zig for all six targets and the resulting binaries pass `build-cstack.sh check`. Prove that on all six before relying on it. If it builds only with a platform's native toolchain, the release job needs that platform's runner, and that is a decision in its own right, not a detail of the change.
