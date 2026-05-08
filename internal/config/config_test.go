package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultEnablesSafeTools(t *testing.T) {
	cfg := Default()
	if !cfg.Tools.Calc || !cfg.Tools.RAG {
		t.Fatal("safe tools should default on")
	}
	if cfg.Tools.Shell.Enabled || cfg.Tools.FS.Enabled {
		t.Fatal("privileged tools should default off")
	}
}

func TestLoadFileOverrides(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	os.WriteFile(path, []byte(`{"server":{"name":"custom"},"log":{"level":"debug"}}`), 0o644)
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Server.Name != "custom" {
		t.Fatalf("name=%q", cfg.Server.Name)
	}
	if cfg.Log.Level != "debug" {
		t.Fatalf("level=%q", cfg.Log.Level)
	}
}

func TestEnvOverride(t *testing.T) {
	t.Setenv("MCPKIT_LOG_LEVEL", "error")
	t.Setenv("MCPKIT_FS_ROOT", "/data")
	cfg, err := Load("")
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Log.Level != "error" {
		t.Fatalf("env level not applied: %q", cfg.Log.Level)
	}
	if !cfg.Tools.FS.Enabled || cfg.Tools.FS.Root != "/data" {
		t.Fatalf("env fs root not applied: %+v", cfg.Tools.FS)
	}
}
