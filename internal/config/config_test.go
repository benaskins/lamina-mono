package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_MissingFile(t *testing.T) {
	cfg, err := Load("/nonexistent/path/config.yaml")
	if err != nil {
		t.Fatalf("expected no error for missing file, got: %v", err)
	}
	if cfg.LLM != nil {
		t.Error("expected nil LLM config for missing file")
	}
}

func TestLoad_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	os.WriteFile(path, []byte(""), 0644)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.LLM != nil {
		t.Error("expected nil LLM config for empty file")
	}
}

func TestLoad_WithLLM(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	os.WriteFile(path, []byte(`
llm:
  provider: anthropic
  model: claude-sonnet-4-20250514
  api_key_env: ANTHROPIC_API_KEY
`), 0644)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.LLM == nil {
		t.Fatal("expected LLM config to be set")
	}
	if cfg.LLM.Provider != "anthropic" {
		t.Errorf("expected provider anthropic, got %q", cfg.LLM.Provider)
	}
	if cfg.LLM.Model != "claude-sonnet-4-20250514" {
		t.Errorf("expected model claude-sonnet-4-20250514, got %q", cfg.LLM.Model)
	}
	if cfg.LLM.APIKeyEnv != "ANTHROPIC_API_KEY" {
		t.Errorf("expected api_key_env ANTHROPIC_API_KEY, got %q", cfg.LLM.APIKeyEnv)
	}
}

func TestLLM_Configured(t *testing.T) {
	cfg := &Config{}
	if cfg.LLMConfigured() {
		t.Error("expected LLMConfigured false with nil LLM")
	}

	cfg.LLM = &LLM{Provider: "anthropic", Model: "test"}
	if cfg.LLMConfigured() {
		t.Error("expected LLMConfigured false without api_key_env")
	}

	cfg.LLM.APIKeyEnv = "MY_KEY"
	if !cfg.LLMConfigured() {
		t.Error("expected LLMConfigured true with all fields set")
	}
}

func TestLLM_APIKey(t *testing.T) {
	llm := &LLM{APIKeyEnv: "TEST_LAMINA_KEY_12345"}
	t.Setenv("TEST_LAMINA_KEY_12345", "sk-test-123")

	key, err := llm.APIKey()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if key != "sk-test-123" {
		t.Errorf("expected sk-test-123, got %q", key)
	}
}

func TestLLM_APIKey_Missing(t *testing.T) {
	llm := &LLM{APIKeyEnv: "NONEXISTENT_KEY_LAMINA_TEST"}
	_, err := llm.APIKey()
	if err == nil {
		t.Error("expected error for missing env var")
	}
}
