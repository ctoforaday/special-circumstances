package testhome_test

import (
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/shared/testhome"
)

func TestClaudeOAuthEnvFrom(t *testing.T) {
	host := t.TempDir()
	claudeDir := filepath.Join(host, ".claude")
	if err := os.MkdirAll(claudeDir, 0o700); err != nil {
		t.Fatal(err)
	}
	raw := `{"claudeAiOauth":{"accessToken":"sk-ant-oat01-test-access","refreshToken":"sk-ant-ort01-secret-refresh","expiresAt":9999999999999}}`
	if err := os.WriteFile(filepath.Join(claudeDir, ".credentials.json"), []byte(raw), 0o600); err != nil {
		t.Fatal(err)
	}

	kv, err := testhome.ClaudeOAuthEnvFrom(host)
	if err != nil {
		t.Fatalf("ClaudeOAuthEnvFrom: %v", err)
	}
	if kv != "CLAUDE_CODE_OAUTH_TOKEN=sk-ant-oat01-test-access" {
		t.Fatalf("got %q, want access token env var only", kv)
	}
	if strings.Contains(kv, "refresh") {
		t.Fatalf("ClaudeOAuthEnvFrom leaked refresh token: %q", kv)
	}

	// Missing file fails loudly.
	if _, err := testhome.ClaudeOAuthEnvFrom(t.TempDir()); err == nil {
		t.Fatal("expected error for missing .credentials.json")
	}

	// Missing accessToken fails loudly.
	emptyHost := t.TempDir()
	_ = os.MkdirAll(filepath.Join(emptyHost, ".claude"), 0o700)
	_ = os.WriteFile(filepath.Join(emptyHost, ".claude", ".credentials.json"), []byte(`{"claudeAiOauth":{"refreshToken":"only-refresh"}}`), 0o600)
	if _, err := testhome.ClaudeOAuthEnvFrom(emptyHost); err == nil {
		t.Fatal("expected error when accessToken is absent")
	}
}

func TestStageAgyAccessTokenFrom(t *testing.T) {
	host := t.TempDir()
	srcDir := filepath.Join(host, ".gemini", "antigravity-cli")
	if err := os.MkdirAll(srcDir, 0o700); err != nil {
		t.Fatal(err)
	}
	raw := `{
  "refresh_token": "top-level-refresh",
  "token": {
    "access_token": "ya29.test-access",
    "refresh_token": "1//secret-refresh-token",
    "expiry": "2030-01-01T00:00:00Z"
  },
  "nested": [{"refresh_token": "nested-refresh"}]
}`
	if err := os.WriteFile(filepath.Join(srcDir, "antigravity-oauth-token"), []byte(raw), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "antigravity_state.pbtxt"), []byte("state: 1\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "installation_id"), []byte("inst-123"), 0o600); err != nil {
		t.Fatal(err)
	}

	fake := t.TempDir()
	if err := testhome.StageAgyAccessTokenFrom(host, fake); err != nil {
		t.Fatalf("StageAgyAccessTokenFrom: %v", err)
	}

	dstTok := filepath.Join(fake, ".gemini", "antigravity-cli", "antigravity-oauth-token")
	info, err := os.Stat(dstTok)
	if err != nil {
		t.Fatalf("stat staged token: %v", err)
	}
	if runtime.GOOS != "windows" && info.Mode().Perm() != 0o600 {
		t.Errorf("staged token mode = %o, want 0600", info.Mode().Perm())
	}
	data, err := os.ReadFile(dstTok)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "refresh_token") || strings.Contains(string(data), "secret-refresh-token") {
		t.Fatalf("staged token retained refresh_token: %s", data)
	}
	var parsed struct {
		Token struct {
			AccessToken string `json:"access_token"`
			Expiry      string `json:"expiry"`
		} `json:"token"`
	}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("parse staged token: %v", err)
	}
	if parsed.Token.AccessToken != "ya29.test-access" {
		t.Fatalf("access_token = %q, want ya29.test-access", parsed.Token.AccessToken)
	}
	for _, extra := range []string{"antigravity_state.pbtxt", "installation_id"} {
		if _, err := os.Stat(filepath.Join(fake, ".gemini", "antigravity-cli", extra)); err != nil {
			t.Errorf("expected companion file %s to be staged: %v", extra, err)
		}
	}

	// Missing token fails loudly.
	if err := testhome.StageAgyAccessTokenFrom(t.TempDir(), t.TempDir()); err == nil {
		t.Fatal("expected error for missing antigravity-oauth-token")
	}

	// Missing access_token fails loudly.
	badHost := t.TempDir()
	_ = os.MkdirAll(filepath.Join(badHost, ".gemini", "antigravity-cli"), 0o700)
	_ = os.WriteFile(filepath.Join(badHost, ".gemini", "antigravity-cli", "antigravity-oauth-token"), []byte(`{"token":{"refresh_token":"only"}}`), 0o600)
	if err := testhome.StageAgyAccessTokenFrom(badHost, t.TempDir()); err == nil {
		t.Fatal("expected error when token.access_token is absent")
	}

	// Expired access_token fails fast instead of letting agy hang on interactive OAuth.
	expHost := t.TempDir()
	_ = os.MkdirAll(filepath.Join(expHost, ".gemini", "antigravity-cli"), 0o700)
	_ = os.WriteFile(filepath.Join(expHost, ".gemini", "antigravity-cli", "antigravity-oauth-token"), []byte(`{"token":{"access_token":"ya29.expired","expiry":"2020-01-01T00:00:00Z"}}`), 0o600)
	if err := testhome.StageAgyAccessTokenFrom(expHost, t.TempDir()); err == nil || !strings.Contains(err.Error(), "expired") {
		t.Fatalf("expected expired-token error, got %v", err)
	}
}

// TestNoRawHostCredentialCopies enforces #1890 across every *_test.go file in the repository:
// live tests MUST lend access tokens through testhome.ClaudeOAuthEnv / testhome.StageAgyAccessToken
// rather than reading ~/.claude/.credentials.json or ~/.gemini/antigravity-cli/antigravity-oauth-token
// directly (which copies refresh tokens into temp HOMEs and signs out host Remote Control sessions).
func TestNoRawHostCredentialCopies(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("repo root: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "shared", "testhome", "liveauth.go")); err != nil {
		t.Fatalf("expected repo root at %s: %v", root, err)
	}

	var violations []string
	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			switch d.Name() {
			case ".git", ".claude", "vendor", "node_modules":
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(d.Name(), "_test.go") {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if rel == "shared/testhome/liveauth_test.go" {
			return nil
		}

		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			return err
		}
		var hasAgyTok bool
		ast.Inspect(f, func(n ast.Node) bool {
			if lit, ok := n.(*ast.BasicLit); ok && lit.Kind == token.STRING {
				if s, err := strconv.Unquote(lit.Value); err == nil && strings.Contains(s, "antigravity-oauth-token") {
					hasAgyTok = true
				}
			}
			return true
		})
		if hasAgyTok {
			violations = append(violations, rel+": references antigravity-oauth-token directly (use testhome.StageAgyAccessToken)")
		}
		for _, decl := range f.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil {
				continue
			}
			var fnClaudeCred, fnUserHome, fnReadOrWrite bool
			ast.Inspect(fn.Body, func(n ast.Node) bool {
				switch x := n.(type) {
				case *ast.BasicLit:
					if x.Kind == token.STRING {
						if s, err := strconv.Unquote(x.Value); err == nil && strings.Contains(s, ".credentials.json") {
							fnClaudeCred = true
						}
					}
				case *ast.SelectorExpr:
					switch x.Sel.Name {
					case "UserHomeDir":
						fnUserHome = true
					case "ReadFile", "Open", "WriteFile":
						fnReadOrWrite = true
					}
				case *ast.Ident:
					if strings.HasPrefix(strings.ToLower(x.Name), "copyfile") {
						fnReadOrWrite = true
					}
				}
				return true
			})
			if fnClaudeCred && fnUserHome && fnReadOrWrite {
				violations = append(violations, rel+":"+fn.Name.Name+": reads/copies host .credentials.json directly (use testhome.ClaudeOAuthEnv)")
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk repo: %v", err)
	}
	for _, v := range violations {
		t.Error(v)
	}
}
