// Package gitleakstoml translates the secrets gate's gitleaks.toml the way the gitleaks command
// line does: viper reads it, and gitleaks' Translate extends it with gitleaks' default
// configuration, compiles every pattern and validates every rule. It is the source the secrets
// package's generated table is written from and checked against; no hook imports it.
package gitleakstoml

import (
	_ "embed"
	"fmt"
	"strings"
	"sync"

	"github.com/spf13/viper"
	"github.com/zricethezav/gitleaks/v8/config"
)

// File is the configuration's path relative to the tools module.
const File = "internal/secrets/gitleakstoml/gitleaks.toml"

//go:embed gitleaks.toml
var text string

var translated struct {
	once sync.Once
	cfg  config.Config
	err  error
}

// Config returns gitleaks.toml translated.
//
// Once per process, because gitleaks loads a configuration through viper's global instance and
// counts its `[extend]` depth in a package global that is never reset: a second translation in
// one process extends nothing and validates nothing.
func Config() (config.Config, error) {
	translated.once.Do(func() {
		viper.SetConfigType("toml")
		if err := viper.ReadConfig(strings.NewReader(text)); err != nil {
			translated.err = fmt.Errorf("reading %s: %w", File, err)
			return
		}
		var vc config.ViperConfig
		if err := viper.Unmarshal(&vc); err != nil {
			translated.err = fmt.Errorf("decoding %s: %w", File, err)
			return
		}
		translated.cfg, translated.err = vc.Translate()
	})
	return translated.cfg, translated.err
}
