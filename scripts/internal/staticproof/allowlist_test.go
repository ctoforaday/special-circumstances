package staticproof

import "testing"

func TestPEAllowlistAdmitsOnlyKernel32NtdllAndTheUCRTApiSets(t *testing.T) {
	for _, c := range []struct {
		lib   string
		allow bool
	}{
		{"kernel32.dll", true},
		{"ntdll.dll", true}, // feov-ocr's C runtime imports it directly
		{"api-ms-win-crt-runtime-l1-1-0.dll", true},
		{"api-ms-win-crt-math-l1-1-0.dll", true},
		{"user32.dll", false},
		{"msvcrt.dll", false},
		{"ws2_32.dll", false},
		{"advapi32.dll", false},
		{"api-ms-win-core-file-l1-1-0.dll", false}, // core api-sets are NOT the CRT ones
		{"", false},
	} {
		for _, art := range []string{"", "feov-record", "feov-ocr"} {
			if got := PELibAllowed(art, c.lib); got != c.allow {
				t.Errorf("PELibAllowed(%q, %q) = %v, want %v", art, c.lib, got, c.allow)
			}
		}
	}
}

func TestTheEmbedAllowlistIsItsOwn(t *testing.T) {
	for _, lib := range []string{
		"advapi32.dll",
		"dbghelp.dll",
		"api-ms-win-core-path-l1-1-0.dll",
	} {
		if !PELibAllowed("feov-embed", lib) {
			t.Errorf("PELibAllowed(%q, %q) = false, want true", "feov-embed", lib)
		}
		for _, other := range []string{"", "feov-record", "feov-ocr"} {
			if PELibAllowed(other, lib) {
				t.Errorf("PELibAllowed(%q, %q) = true, want false", other, lib)
			}
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
		{"/System/Library/Frameworks/Security.framework/Versions/A/Security", true}, // crypto/x509 links it on darwin
		{"/System/Library/Frameworks/CoreServices.framework/Versions/A/CoreServices", false},
		{"/usr/local/lib/libtesseract.5.dylib", false},
		{"/opt/homebrew/lib/libleptonica.dylib", false},
		{"/usr/lib/libc++.1.dylib", false},
		{"", false},
	} {
		if got := MachOLibAllowed("feov-ocr", c.lib); got != c.allow {
			t.Errorf("MachOLibAllowed(%q) = %v, want %v", c.lib, got, c.allow)
		}
	}
}
