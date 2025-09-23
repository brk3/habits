package config

import (
	"os"
	"path/filepath"
	"testing"

	"go.yaml.in/yaml/v4"
)

func TestLoad_MissingConfig(t *testing.T) {
	t.Setenv("HABITS_CONFIG", "nonexistent.yaml")
	_, err := Load()
	if err == nil {
		t.Fatal("expected error for missing config file, got nil")
	}
}

func TestLoad_CustomConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")
	t.Setenv("HABITS_CONFIG", configFile)

	c := Config{}
	c.Server.Port = 1234
	d, err := yaml.Marshal(&c)
	if err != nil {
		t.Fatalf("failed to marshal config: %v", err)
	}
	if err := os.WriteFile(configFile, d, 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	// check custom setting
	cfg, err := Load()
	if err != nil {
		t.Fatal("error opening config:", err)
	}
	if cfg.Server.Port != 1234 {
		t.Errorf("expected server port 1234, got %d", cfg.Server.Port)
	}

	// check default setting
	if cfg.Server.Host != "0.0.0.0" {
		t.Errorf("expected default server host 0.0.0.0, got %s", cfg.Server.Host)
	}
}
