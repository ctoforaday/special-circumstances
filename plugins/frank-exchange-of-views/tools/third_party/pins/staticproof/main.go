// Staticproof asserts the per-format staticness contract of plan §V.3 on one binary:
//
//	linux (ELF):    fully static — no interpreter, no dynamic section at all.
//	windows (PE):   imports only KERNEL32.dll and the api-ms-win-crt-* UCRT api-sets,
//	                which Windows 10+ provides; no third-party DLLs.
//	darwin (Mach-O): load commands reference only OS-provided libraries — libSystem.B,
//	                CoreFoundation, libresolv.9. Fully static does not exist on macOS
//	                (no static libSystem); the latter two are Go-runtime link flags
//	                satisfied by .tbd stubs at build time and the OS at run time.
//
// It parses the real import records with debug/elf, debug/pe and debug/macho rather than
// shelling to platform tools, because the build host cannot audit two of the three
// formats natively: ldd reads only ELF, and GNU objdump has no pei-aarch64 backend — the
// Wave 0 spike had to fall back to a string-scan for windows/arm64, which can only hope
// about string shape where a parse can refuse. Exit 0 is the proof; any violation names
// what it found and exits 1.
//
// Usage: go run ./third_party/pins/staticproof <binary>
// (or via `build-cstack.sh check <binary>`)
package main

import (
	"debug/elf"
	"debug/macho"
	"debug/pe"
	"fmt"
	"os"
	"strings"
)

func main() {
	if len(os.Args) != 2 {
		fail("usage: staticproof <binary>")
	}
	path := os.Args[1]
	// Try the three formats by parse rather than by extension: the format is a fact the
	// file carries, and .exe on a Mach-O would otherwise route it to the wrong proof.
	if f, err := elf.Open(path); err == nil {
		defer f.Close()
		checkELF(path, f)
		return
	}
	if f, err := pe.Open(path); err == nil {
		defer f.Close()
		checkPE(path, f)
		return
	}
	if f, err := macho.Open(path); err == nil {
		defer f.Close()
		checkMachO(path, f)
		return
	}
	fail("%s: not ELF, PE or Mach-O — nothing to prove", path)
}

func checkELF(path string, f *elf.File) {
	// "Fully static" has two independent witnesses and both must be absent: PT_INTERP
	// (asks for a loader) and PT_DYNAMIC (asks for libraries). Checking only one lets a
	// static-PIE-with-DT_NEEDED hybrid through.
	for _, p := range f.Progs {
		switch p.Type {
		case elf.PT_INTERP:
			fail("%s: ELF requests an interpreter — not statically linked", path)
		case elf.PT_DYNAMIC:
			libs, _ := f.ImportedLibraries()
			fail("%s: ELF has a dynamic section (needs %v) — not statically linked", path, libs)
		}
	}
	fmt.Printf("%s: ELF, fully static (no interpreter, no dynamic section)\n", path)
}

func checkPE(path string, f *pe.File) {
	// ImportedSymbols yields "symbol:LIBRARY.dll" rows; the library set is what the
	// contract bounds. Case-folded: PE import names are case-insensitive on disk.
	syms, err := f.ImportedSymbols()
	if err != nil {
		fail("%s: reading PE imports: %v", path, err)
	}
	libs := map[string]bool{}
	for _, s := range syms {
		if i := strings.LastIndex(s, ":"); i >= 0 {
			libs[strings.ToLower(s[i+1:])] = true
		}
	}
	var bad []string
	for lib := range libs {
		if lib != "kernel32.dll" && !strings.HasPrefix(lib, "api-ms-win-crt-") {
			bad = append(bad, lib)
		}
	}
	if len(bad) > 0 {
		fail("%s: PE imports beyond KERNEL32 + UCRT api-sets: %v", path, bad)
	}
	fmt.Printf("%s: PE, imports only KERNEL32 + UCRT api-sets (%d libraries)\n", path, len(libs))
}

func checkMachO(path string, f *macho.File) {
	allowed := map[string]bool{
		"/usr/lib/libSystem.B.dylib": true,
		"/usr/lib/libresolv.9.dylib": true,
		"/System/Library/Frameworks/CoreFoundation.framework/Versions/A/CoreFoundation": true,
	}
	libs, err := f.ImportedLibraries()
	if err != nil {
		fail("%s: reading Mach-O load commands: %v", path, err)
	}
	var bad []string
	for _, lib := range libs {
		if !allowed[lib] {
			bad = append(bad, lib)
		}
	}
	if len(bad) > 0 {
		fail("%s: Mach-O references non-OS libraries: %v", path, bad)
	}
	fmt.Printf("%s: Mach-O, references only OS-provided libraries (%v)\n", path, libs)
}

func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "staticproof: "+format+"\n", args...)
	os.Exit(1)
}
