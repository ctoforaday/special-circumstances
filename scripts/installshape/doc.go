// Package installshape holds the install-shape contract test: it installs the suite through the
// real Claude Code and Antigravity clients in every install shape, fires the hooks a session with
// no login fires, and asserts that every reader of "where this session's plugin lives" agrees
// with the root the client actually handed the hooks.
//
// clients.json pins the client versions; the scripts job in .github/workflows/hooks.yml installs
// exactly those on a pull request, and the latest on the nightly (INSTALLSHAPE_CLIENTS=latest).
package installshape
