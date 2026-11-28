package peer

import "path/filepath"

// DaemonTokenVar is the variable whose value Antigravity's language server adopts as its CSRF
// token (agy 1.5.18; TestLiveAntigravityMailboxWake/ChosenTokenIsHonoured pins it). It reaches no hook, MCP
// server or agent shell, so the daemon gets it from a file the human provisions.
const DaemonTokenVar = "ANTIGRAVITY_SERVER_CSRF_TOKEN"

// DaemonUnit is the systemd user unit that runs Antigravity's remote-control daemon.
const DaemonUnit = "antigravity-cli-daemon.service"

// DaemonTokenFile is where the daemon's token lives under a configuration directory
// ($XDG_CONFIG_HOME, else ~/.config).
func DaemonTokenFile(configDir string) string {
	if configDir == "" {
		return ""
	}
	return filepath.Join(configDir, "gray-area", "antigravity-daemon.env")
}
