package testhome

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// ClaudeOAuthEnv returns "CLAUDE_CODE_OAUTH_TOKEN=<accessToken>" read from the host user's
// ~/.claude/.credentials.json, lending ONLY the short-lived access token and never copying
// .credentials.json or its refreshToken into an isolated test HOME (#1890).
//
// Copying .credentials.json into a throwaway HOME lets a live test rotate the OAuth refresh grant
// into the throwaway copy, stranding host Remote Control sessions on a spent refresh token.
func ClaudeOAuthEnv() (string, error) {
	hostHome, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("testhome: resolve user home for Claude credentials: %w", err)
	}
	return ClaudeOAuthEnvFrom(hostHome)
}

// ClaudeOAuthEnvFrom is ClaudeOAuthEnv scoped to an explicit hostHome directory (for tests).
func ClaudeOAuthEnvFrom(hostHome string) (string, error) {
	credPath := filepath.Join(hostHome, ".claude", ".credentials.json")
	raw, err := os.ReadFile(credPath)
	if err != nil {
		return "", fmt.Errorf("testhome: read Claude credentials %s: %w", credPath, err)
	}
	var creds struct {
		OAuth struct {
			AccessToken string `json:"accessToken"`
		} `json:"claudeAiOauth"`
	}
	if err := json.Unmarshal(raw, &creds); err != nil {
		return "", fmt.Errorf("testhome: parse Claude credentials %s: %w", credPath, err)
	}
	if creds.OAuth.AccessToken == "" {
		return "", fmt.Errorf("testhome: %s has no claudeAiOauth.accessToken", credPath)
	}
	return "CLAUDE_CODE_OAUTH_TOKEN=" + creds.OAuth.AccessToken, nil
}

// StageAgyAccessToken stages the host user's Antigravity OAuth access token into fakeHome with
// every refresh_token field stripped fail-closed, plus non-secret companion state files
// (antigravity_state.pbtxt, installation_id) when present (#1890).
func StageAgyAccessToken(fakeHome string) error {
	hostHome, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("testhome: resolve user home for Antigravity credentials: %w", err)
	}
	return StageAgyAccessTokenFrom(hostHome, fakeHome)
}

// StageAgyAccessTokenFrom is StageAgyAccessToken scoped to an explicit hostHome directory (for tests).
func StageAgyAccessTokenFrom(hostHome, fakeHome string) error {
	srcDir := filepath.Join(hostHome, ".gemini", "antigravity-cli")
	srcTok := filepath.Join(srcDir, "antigravity-oauth-token")
	raw, err := os.ReadFile(srcTok)
	if err != nil {
		return fmt.Errorf("testhome: read Antigravity credentials %s: %w", srcTok, err)
	}
	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		return fmt.Errorf("testhome: parse Antigravity credentials %s: %w", srcTok, err)
	}
	tokObj, ok := doc["token"].(map[string]any)
	if !ok {
		return fmt.Errorf("testhome: %s missing token object", srcTok)
	}
	accessTok, _ := tokObj["access_token"].(string)
	if accessTok == "" {
		return fmt.Errorf("testhome: %s missing token.access_token", srcTok)
	}
	if expStr, _ := tokObj["expiry"].(string); expStr != "" {
		if exp, perr := time.Parse(time.RFC3339Nano, expStr); perr == nil && !exp.After(time.Now()) {
			return fmt.Errorf("testhome: Antigravity access token in %s expired at %s — run a host `agy` turn to renew it", srcTok, expStr)
		}
	}

	stripJSONKey(doc, "refresh_token")
	if hasJSONKey(doc, "refresh_token") {
		return errors.New("testhome: refresh_token survived strip in Antigravity token object")
	}
	stripped, err := json.Marshal(doc)
	if err != nil {
		return fmt.Errorf("testhome: marshal refresh-stripped Antigravity token: %w", err)
	}
	if bytes.Contains(stripped, []byte("refresh_token")) {
		return errors.New("testhome: refresh_token substring survived strip in Antigravity token JSON")
	}

	dstDir := filepath.Join(fakeHome, ".gemini", "antigravity-cli")
	if err := os.MkdirAll(dstDir, 0o700); err != nil {
		return fmt.Errorf("testhome: mkdir %s: %w", dstDir, err)
	}
	dstTok := filepath.Join(dstDir, "antigravity-oauth-token")
	if err := os.WriteFile(dstTok, stripped, 0o600); err != nil {
		return fmt.Errorf("testhome: write %s: %w", dstTok, err)
	}
	for _, extra := range []string{"antigravity_state.pbtxt", "installation_id"} {
		data, err := os.ReadFile(filepath.Join(srcDir, extra))
		if err != nil {
			continue
		}
		if err := os.WriteFile(filepath.Join(dstDir, extra), data, 0o600); err != nil {
			return fmt.Errorf("testhome: write %s: %w", extra, err)
		}
	}
	return nil
}

func stripJSONKey(v any, key string) {
	switch x := v.(type) {
	case map[string]any:
		delete(x, key)
		for _, child := range x {
			stripJSONKey(child, key)
		}
	case []any:
		for _, child := range x {
			stripJSONKey(child, key)
		}
	}
}

func hasJSONKey(v any, key string) bool {
	switch x := v.(type) {
	case map[string]any:
		if _, ok := x[key]; ok {
			return true
		}
		for _, child := range x {
			if hasJSONKey(child, key) {
				return true
			}
		}
	case []any:
		for _, child := range x {
			if hasJSONKey(child, key) {
				return true
			}
		}
	}
	return false
}
