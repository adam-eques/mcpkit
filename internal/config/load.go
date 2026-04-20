package config

import (
	"encoding/json"
	"os"
	"strconv"
	"strings"
)

// Load returns the default configuration merged with an optional JSON file and
// then environment-variable overrides. A path of "" skips the file.
func Load(path string) (Config, error) {
	cfg := Default()
	if path != "" {
		raw, err := os.ReadFile(path)
		if err != nil {
			return cfg, err
		}
		if err := json.Unmarshal(raw, &cfg); err != nil {
			return cfg, err
		}
	}
	applyEnv(&cfg)
	return cfg, nil
}

// applyEnv overrides selected fields from MCPKIT_* environment variables so the
// server can be configured in a container without a config file.
func applyEnv(cfg *Config) {
	if v := os.Getenv("MCPKIT_LOG_LEVEL"); v != "" {
		cfg.Log.Level = v
	}
	if v := os.Getenv("MCPKIT_LOG_FORMAT"); v != "" {
		cfg.Log.Format = v
	}
	if v := os.Getenv("MCPKIT_GATEWAY_ADDR"); v != "" {
		cfg.Gateway.Address = v
	}
	if v := os.Getenv("MCPKIT_CONCURRENCY"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.Server.Concurrency = n
		}
	}
	if v := os.Getenv("MCPKIT_FS_ROOT"); v != "" {
		cfg.Tools.FS.Enabled = true
		cfg.Tools.FS.Root = v
	}
	if v := os.Getenv("MCPKIT_FETCH"); v != "" {
		cfg.Tools.Fetch.Enabled = truthy(v)
	}
	if v := os.Getenv("MCPKIT_WEBSEARCH"); v != "" {
		cfg.Tools.WebSearch.Enabled = truthy(v)
	}
}

func truthy(s string) bool {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}
