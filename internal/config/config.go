// Package config defines the mcpkit server configuration and loads it from an
// optional JSON file layered with environment-variable overrides.
package config

// Config is the top-level server configuration.
type Config struct {
	Server  ServerConfig  `json:"server"`
	Log     LogConfig     `json:"log"`
	Gateway GatewayConfig `json:"gateway"`
	Tools   ToolsConfig   `json:"tools"`
}

// ServerConfig controls server identity and dispatch.
type ServerConfig struct {
	Name         string `json:"name"`
	Version      string `json:"version"`
	Instructions string `json:"instructions"`
	Concurrency  int    `json:"concurrency"`
}

// LogConfig controls logging.
type LogConfig struct {
	Level  string `json:"level"`  // debug, info, warn, error
	Format string `json:"format"` // text or json
}

// GatewayConfig controls the HTTP gateway binary.
type GatewayConfig struct {
	Address string `json:"address"`
}

// ToolsConfig enables and configures individual tools.
type ToolsConfig struct {
	Calc      bool           `json:"calc"`
	Fetch     FetchConfig    `json:"fetch"`
	FS        FSConfig       `json:"fs"`
	Shell     ShellConfig    `json:"shell"`
	KV        KVConfig       `json:"kv"`
	WebSearch WebConfig      `json:"webSearch"`
	RAG       bool           `json:"rag"`
	Time      bool           `json:"time"`
	Text      bool           `json:"text"`
	JSONQuery bool           `json:"jsonQuery"`
}

// FetchConfig configures the http_fetch tool.
type FetchConfig struct {
	Enabled      bool  `json:"enabled"`
	AllowPrivate bool  `json:"allowPrivate"`
	MaxBytes     int64 `json:"maxBytes"`
}

// FSConfig configures the filesystem tools.
type FSConfig struct {
	Enabled  bool   `json:"enabled"`
	Root     string `json:"root"`
	ReadOnly bool   `json:"readOnly"`
}

// ShellConfig configures the shell_exec tool.
type ShellConfig struct {
	Enabled        bool     `json:"enabled"`
	Allowlist      []string `json:"allowlist"`
	TimeoutSeconds int      `json:"timeoutSeconds"`
}

// KVConfig configures the key/value tools.
type KVConfig struct {
	Enabled bool   `json:"enabled"`
	Path    string `json:"path"`
}

// WebConfig configures the web_search tool.
type WebConfig struct {
	Enabled  bool   `json:"enabled"`
	Endpoint string `json:"endpoint"`
}

// Default returns a configuration with the safe tools enabled and the network,
// filesystem and shell tools disabled until explicitly turned on.
func Default() Config {
	return Config{
		Server: ServerConfig{
			Name:        "mcpkit",
			Version:     "0.1.0",
			Concurrency: 16,
		},
		Log:     LogConfig{Level: "info", Format: "text"},
		Gateway: GatewayConfig{Address: ":8080"},
		Tools: ToolsConfig{
			Calc:      true,
			RAG:       true,
			Time:      true,
			Text:      true,
			JSONQuery: true,
			Fetch:     FetchConfig{Enabled: false, MaxBytes: 5 << 20},
			FS:        FSConfig{Enabled: false, ReadOnly: true},
			Shell:     ShellConfig{Enabled: false, TimeoutSeconds: 15},
			KV:        KVConfig{Enabled: true},
			WebSearch: WebConfig{Enabled: false},
		},
	}
}
