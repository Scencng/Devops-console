package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const DefaultConfigPath = "config/app.yaml"

type AppConfig struct {
	Server   ServerConfig   `yaml:"server"`
	Frontend FrontendConfig `yaml:"frontend"`
}

type ServerConfig struct {
	Address string `yaml:"address"`
}

type FrontendConfig struct {
	DistDir string `yaml:"dist_dir"`
}

type LoadOptions struct {
	BaseDir                 string
	ConfigPath              string
	DefaultConfigPath       string
	OverrideAddress         string
	OverrideFrontendDistDir string
}

func Load(options LoadOptions) (AppConfig, error) {
	baseDir := options.BaseDir
	if baseDir == "" {
		baseDir = "."
	}

	defaultConfigPath := options.DefaultConfigPath
	if defaultConfigPath == "" {
		defaultConfigPath = DefaultConfigPath
	}

	cfg := AppConfig{
		Server: ServerConfig{
			Address: "0.0.0.0:8080",
		},
		Frontend: FrontendConfig{
			DistDir: filepath.Join(baseDir, "..", "frontend", "dist"),
		},
	}

	configPath := options.ConfigPath
	if configPath == "" {
		configPath = defaultConfigPath
	}

	resolvedConfigPath := resolvePath(baseDir, configPath)
	configPathProvided := options.ConfigPath != ""
	if stat, err := os.Stat(resolvedConfigPath); err == nil && !stat.IsDir() {
		fileData, readErr := os.ReadFile(resolvedConfigPath)
		if readErr != nil {
			return AppConfig{}, fmt.Errorf("read config file failed: %w", readErr)
		}

		if unmarshalErr := yaml.Unmarshal(fileData, &cfg); unmarshalErr != nil {
			return AppConfig{}, fmt.Errorf("parse config file failed: %w", unmarshalErr)
		}
	} else if configPathProvided {
		if err == nil {
			return AppConfig{}, errors.New("config path points to a directory")
		}
		return AppConfig{}, fmt.Errorf("config file not found: %s", resolvedConfigPath)
	}

	if options.OverrideAddress != "" {
		cfg.Server.Address = options.OverrideAddress
	}

	if options.OverrideFrontendDistDir != "" {
		cfg.Frontend.DistDir = options.OverrideFrontendDistDir
	}

	if cfg.Server.Address == "" {
		cfg.Server.Address = "0.0.0.0:8080"
	}

	if cfg.Frontend.DistDir == "" {
		cfg.Frontend.DistDir = filepath.Join(baseDir, "..", "frontend", "dist")
	}

	cfg.Frontend.DistDir = resolvePath(baseDir, cfg.Frontend.DistDir)
	return cfg, nil
}

func resolvePath(baseDir, target string) string {
	if target == "" {
		return target
	}

	if filepath.IsAbs(target) {
		return filepath.Clean(target)
	}

	return filepath.Clean(filepath.Join(baseDir, target))
}
