package acceptance_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/cucumber/godog"
	"github.com/sommerfeld-io/github-metrics-exporter/internal/config"
)

type configScenarioState struct {
	configPath string
	tmpDir     string
	loadErr    error
	cfg        *config.Config
}

func (s *configScenarioState) aConfigFileWherePortIs(yamlValue string) error {
	content := fmt.Sprintf("port: %s\n", yamlValue)
	if err := os.WriteFile(s.configPath, []byte(content), 0600); err != nil {
		return fmt.Errorf("setup: write config: %w", err)
	}
	return nil
}

func (s *configScenarioState) aConfigFileWithNoPortKey() error {
	if err := os.WriteFile(s.configPath, []byte("other_key: value\n"), 0600); err != nil {
		return fmt.Errorf("setup: write config: %w", err)
	}
	return nil
}

func (s *configScenarioState) theConfigIsLoaded() error {
	s.cfg, s.loadErr = config.Load(s.configPath)
	return nil
}

func (s *configScenarioState) theConfigLoaderReturnsAnError() error {
	if s.loadErr == nil {
		return fmt.Errorf("expected config.Load to return an error, got nil (cfg=%+v)", s.cfg)
	}
	return nil
}

func (s *configScenarioState) theConfigLoaderSucceedsWithPort(port int) error {
	if s.loadErr != nil {
		return fmt.Errorf("expected config.Load to succeed, got error: %v", s.loadErr)
	}
	if s.cfg.Port != port {
		return fmt.Errorf("expected port %d, got %d", port, s.cfg.Port)
	}
	return nil
}

// InitializeConfigScenario registers config scenario step definitions with GoDog.
func InitializeConfigScenario(ctx *godog.ScenarioContext) {
	s := &configScenarioState{}

	ctx.Before(func(goCtx context.Context, sc *godog.Scenario) (context.Context, error) {
		var err error
		s.tmpDir, err = os.MkdirTemp("", "ghme-config-test-*")
		if err != nil {
			return goCtx, fmt.Errorf("setup: create temp dir: %w", err)
		}
		s.configPath = filepath.Join(s.tmpDir, "config.yml")
		s.loadErr = nil
		s.cfg = nil
		return goCtx, nil
	})

	ctx.After(func(goCtx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		_ = os.RemoveAll(s.tmpDir)
		return goCtx, nil
	})

	ctx.Step(`^a config file where port is (.+)$`, s.aConfigFileWherePortIs)
	ctx.Step(`^a config file with no port key$`, s.aConfigFileWithNoPortKey)
	ctx.Step(`^the config is loaded$`, s.theConfigIsLoaded)
	ctx.Step(`^the config loader returns an error$`, s.theConfigLoaderReturnsAnError)
	ctx.Step(`^the config loader succeeds with port (\d+)$`, s.theConfigLoaderSucceedsWithPort)
}
