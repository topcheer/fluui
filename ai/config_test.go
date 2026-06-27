package ai

import (
	"strings"
	"testing"
)

func TestConfigDefaults(t *testing.T) {
	// LoadConfig with no .env and no env vars should fail (no API key)
	t.Setenv("FLUUI_LLM_API_KEY", "")
	cfg, err := LoadConfig("/nonexistent/.env")
	if err == nil {
		t.Error("LoadConfig should fail without API key")
	}
	if cfg != nil {
		t.Error("cfg should be nil on error")
	}
}

func TestConfigFromEnv(t *testing.T) {
	t.Setenv("FLUUI_LLM_API_KEY", "test-key-1234567890")
	t.Setenv("FLUUI_LLM_BASE_URL", "https://api.example.com/v1")
	t.Setenv("FLUUI_LLM_MODEL", "test-model")
	t.Setenv("FLUUI_LLM_SYSTEM_PROMPT", "You are a test bot.")

	cfg, err := LoadConfig("/nonexistent/.env")
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}
	if cfg.APIKey != "test-key-1234567890" {
		t.Errorf("APIKey = %q", cfg.APIKey)
	}
	if cfg.BaseURL != "https://api.example.com/v1" {
		t.Errorf("BaseURL = %q", cfg.BaseURL)
	}
	if cfg.Model != "test-model" {
		t.Errorf("Model = %q", cfg.Model)
	}
	if cfg.SystemPrompt != "You are a test bot." {
		t.Errorf("SystemPrompt = %q", cfg.SystemPrompt)
	}
}

func TestConfigDefaultBaseURL(t *testing.T) {
	t.Setenv("FLUUI_LLM_API_KEY", "sk-1234567890abcdef")
	// Don't set BASE_URL → should default
	t.Setenv("FLUUI_LLM_BASE_URL", "")

	cfg, err := LoadConfig("/nonexistent/.env")
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}
	// When env var is empty, loadDotEnv won't overwrite, and default stays
	if cfg.BaseURL == "" {
		t.Error("BaseURL should have a default value")
	}
}

func TestConfigMaskedKey(t *testing.T) {
	cfg := &Config{APIKey: "sk-abcdefghijklmn"}
	masked := cfg.MaskedKey()
	if !strings.HasPrefix(masked, "sk-a") {
		t.Errorf("MaskedKey should start with first 4 chars, got %q", masked)
	}
	if !strings.HasSuffix(masked, "klmn") {
		t.Errorf("MaskedKey should end with last 4 chars, got %q", masked)
	}
	if strings.Contains(masked, "sk-abcdefghijklmn") {
		t.Error("MaskedKey should not contain the full key")
	}
}

func TestConfigMaskedKeyShort(t *testing.T) {
	cfg := &Config{APIKey: "short"}
	masked := cfg.MaskedKey()
	if masked != "****" {
		t.Errorf("MaskedKey for short key should be ****, got %q", masked)
	}
}

func TestConfigString(t *testing.T) {
	cfg := &Config{
		APIKey:  "sk-abcdefghijklmn",
		BaseURL: "https://api.test.com/v1",
		Model:   "gpt-test",
	}
	s := cfg.String()
	if !strings.Contains(s, "gpt-test") {
		t.Errorf("String should contain model name, got %q", s)
	}
	if !strings.Contains(s, "api.test.com") {
		t.Errorf("String should contain base URL, got %q", s)
	}
	if strings.Contains(s, "sk-abcdefghijklmn") {
		t.Error("String should not contain full API key")
	}
}

func TestNewClient(t *testing.T) {
	cfg := &Config{
		APIKey:  "test-key",
		BaseURL: "https://api.example.com/v1/",
		Model:   "test-model",
	}
	client := NewClient(cfg)
	if client.APIKey != "test-key" {
		t.Errorf("APIKey = %q", client.APIKey)
	}
	if client.Model != "test-model" {
		t.Errorf("Model = %q", client.Model)
	}
	// BaseURL should have trailing slash removed
	if client.BaseURL != "https://api.example.com/v1" {
		t.Errorf("BaseURL = %q, want no trailing slash", client.BaseURL)
	}
	if client.HTTP == nil {
		t.Error("HTTP client should not be nil")
	}
}
