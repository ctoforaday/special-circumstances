package peer

import (
	"errors"
	"fmt"
	"path/filepath"
)

// DaemonTokenVar is the variable whose value Antigravity's language server adopts as its CSRF
// token (agy 1.5.18; TestLiveAntigravityChosenTokenIsHonoured pins it). It reaches no hook, MCP
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

// maxTokenLen bounds the token's value.
const maxTokenLen = 256

// parseDaemonToken reads the file's content: exactly one line DaemonTokenVar=<value> with at most
// one trailing newline, the value 1–maxTokenLen bytes of [A-Za-z0-9-]. No error carries the value.
func parseDaemonToken(b []byte) (string, error) {
	s := string(b)
	if n := len(s); n > 0 && s[n-1] == '\n' {
		s = s[:n-1]
	}
	prefix := DaemonTokenVar + "="
	if len(s) < len(prefix) || s[:len(prefix)] != prefix {
		return "", fmt.Errorf("it is not one line %s=<token>", DaemonTokenVar)
	}
	v := s[len(prefix):]
	if v == "" || len(v) > maxTokenLen {
		return "", fmt.Errorf("its token is empty or longer than %d bytes", maxTokenLen)
	}
	for i := 0; i < len(v); i++ {
		c := v[i]
		if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '-') {
			return "", errors.New("its token holds a byte other than a letter, a digit or '-'")
		}
	}
	return v, nil
}
