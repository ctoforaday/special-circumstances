//go:build tessocr

package testbuild

// EngineCompiledIn says this test binary was built with the OCR engine, so a binary it builds is
// built with it too. The ldflags are hooks.yml's tagged engine step's, a second hand-written copy:
// the workflow is YAML and this is Go, so nothing generates one from the other, and a drift fails
// loudly as a link error in the tagged test rather than silently.
const EngineCompiledIn = true

var engineBuildFlags = []string{"-tags", "tessocr", "-ldflags", `-linkmode external -extldflags "-static"`}
