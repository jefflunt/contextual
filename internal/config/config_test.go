package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadMissingFile(t *testing.T) {
	// Point home to a temp dir with no config file.
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg == nil {
		t.Fatal("expected non-nil config")
	}
	if cfg.Atlassian.Host != "" {
		t.Errorf("expected empty host, got %q", cfg.Atlassian.Host)
	}
}

func TestLoadValidFile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	cfgDir := filepath.Join(dir, ".contextual")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatal(err)
	}

	content := "atlassian:\n  host: example.atlassian.net\n  api_user: user@example.com\n  api_token: secret\n"
	if err := os.WriteFile(filepath.Join(cfgDir, "config.yml"), []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Atlassian.Host != "example.atlassian.net" {
		t.Errorf("expected example.atlassian.net, got %q", cfg.Atlassian.Host)
	}
	if cfg.Atlassian.APIUser != "user@example.com" {
		t.Errorf("expected user@example.com, got %q", cfg.Atlassian.APIUser)
	}
	if cfg.Atlassian.APIToken != "secret" {
		t.Errorf("expected secret, got %q", cfg.Atlassian.APIToken)
	}
}

func TestLoadInvalidYAML(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	cfgDir := filepath.Join(dir, ".contextual")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(cfgDir, "config.yml"), []byte(":\tinvalid:\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}
