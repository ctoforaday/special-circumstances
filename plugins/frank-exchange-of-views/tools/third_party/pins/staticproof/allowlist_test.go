package main

import "testing"

// THE GATE OVER EVERY PUBLISHED BINARY, AND IT HAD NO TEST. This tool is the last thing
// between a release artifact and a consumer's machine, and both of its allowlists fail
// green: invert either and a binary needing libraries the target does not have sails
// through with the same "fully static" line the honest case prints. The predicates are
// pure, so the contract is pinned here rather than inferred from whichever artifacts a
// release happened to build.
func TestPEAllowlistAdmitsOnlyKernel32NtdllAndTheUCRTApiSets(t *testing.T) {
	for _, c := range []struct {
		lib   string
		allow bool
	}{
		{"kernel32.dll", true},
		{"ntdll.dll", true}, // the loader maps it into every process; no target can lack it
		{"api-ms-win-crt-runtime-l1-1-0.dll", true},
		{"api-ms-win-crt-math-l1-1-0.dll", true},
		{"user32.dll", false},
		{"msvcrt.dll", false},
		{"ws2_32.dll", false},
		{"advapi32.dll", false},
		{"api-ms-win-core-file-l1-1-0.dll", false}, // core api-sets are NOT the CRT ones
		{"", false},
	} {
		if got := peLibAllowed(c.lib); got != c.allow {
			t.Errorf("peLibAllowed(%q) = %v, want %v", c.lib, got, c.allow)
		}
	}
}

func TestMachOAllowlistAdmitsOnlyOSProvidedLibraries(t *testing.T) {
	for _, c := range []struct {
		lib   string
		allow bool
	}{
		{"/usr/lib/libSystem.B.dylib", true},
		{"/usr/lib/libresolv.9.dylib", true},
		{"/System/Library/Frameworks/CoreFoundation.framework/Versions/A/CoreFoundation", true},
		{"/System/Library/Frameworks/Security.framework/Versions/A/Security", true}, // crypto/x509 on darwin
		{"/System/Library/Frameworks/CoreServices.framework/Versions/A/CoreServices", false},
		{"/usr/local/lib/libtesseract.5.dylib", false},
		{"/opt/homebrew/lib/libleptonica.dylib", false},
		{"/usr/lib/libc++.1.dylib", false},
		{"", false},
	} {
		if got := machOLibAllowed(c.lib); got != c.allow {
			t.Errorf("machOLibAllowed(%q) = %v, want %v", c.lib, got, c.allow)
		}
	}
}
