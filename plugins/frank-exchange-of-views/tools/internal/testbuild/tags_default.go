//go:build !tessocr

package testbuild

// EngineCompiledIn says this test binary was built with the OCR engine; see tags_tessocr.go.
const EngineCompiledIn = false

var engineBuildFlags []string
