// Package ai provides an OpenAI-compatible streaming chat client
// for connecting Fluui TUI apps to real LLM backends.
//
// Configuration is loaded from a .env file or environment variables.
// API keys are NEVER hardcoded — they must come from the environment.
package ai

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Config holds the LLM connection settings.
type Config struct {
	// APIKey is the authentication key (required).
	APIKey string

	// BaseURL is the OpenAI-compatible API endpoint.
	// Defaults to "https://open.bigmodel.cn/api/coding/paas/v4" (ZAI/智谱).
	BaseURL string

	// Model is the model name to use.
	// Defaults to "glm-5.2".
	Model string

	// SystemPrompt is the optional system message prepended to conversations.
	SystemPrompt string
}

// LoadConfig reads configuration from (in priority order):
//  1. Environment variables (FLUUI_LLM_*)
//  2. .env file in the given path (or current directory)
//
// At minimum, APIKey must be present. Other fields have defaults.
func LoadConfig(envPath ...string) (*Config, error) {
	// Start with defaults
	cfg := &Config{
		BaseURL:      "https://open.bigmodel.cn/api/coding/paas/v4",
		Model:        "glm-5.2",
		SystemPrompt: "You are a helpful assistant.",
	}

	// Try to load .env file
	path := ".env"
	if len(envPath) > 0 && envPath[0] != "" {
		path = envPath[0]
	}
	loadDotEnv(path)

	// Read from environment
	if v := os.Getenv("FLUUI_LLM_API_KEY"); v != "" {
		cfg.APIKey = v
	}
	if v := os.Getenv("FLUUI_LLM_BASE_URL"); v != "" {
		cfg.BaseURL = v
	}
	if v := os.Getenv("FLUUI_LLM_MODEL"); v != "" {
		cfg.Model = v
	}
	if v := os.Getenv("FLUUI_LLM_SYSTEM_PROMPT"); v != "" {
		cfg.SystemPrompt = v
	}

	if cfg.APIKey == "" {
		return nil, fmt.Errorf("FLUUI_LLM_API_KEY not set, configure .env or set environment variable (see .env.example)")
	}

	return cfg, nil
}

// loadDotEnv reads a .env file and sets environment variables.
// Lines starting with # are comments. Format: KEY=VALUE
// Existing environment variables take precedence (not overwritten).
func loadDotEnv(path string) {
	f, err := os.Open(path)
	if err != nil {
		return // file doesn't exist, that's OK
	}
	defer func() { _ = f.Close() }()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse KEY=VALUE
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Remove surrounding quotes
		value = strings.Trim(value, `"'`)

		// Don't overwrite existing env vars
		if os.Getenv(key) == "" {
			_ = os.Setenv(key, value)
		}
	}
}

// MaskedKey returns the API key with only the first 4 and last 4 chars visible.
// Returns "****" if the key is too short.
func (c *Config) MaskedKey() string {
	k := c.APIKey
	if len(k) <= 8 {
		return "****"
	}
	return k[:4] + "..." + k[len(k)-4:]
}

// String returns a human-readable config summary (without the full API key).
func (c *Config) String() string {
	return fmt.Sprintf("Model=%s BaseURL=%s Key=%s", c.Model, c.BaseURL, c.MaskedKey())
}
