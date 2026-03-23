package docreview

import (
	"testing"
)

func TestUserMessage(t *testing.T) {
	msg := userMessage("axon-tool", "v0.1.6", "v0.1.7")
	if msg == "" {
		t.Error("expected non-empty message")
	}
	if !contains(msg, "axon-tool") || !contains(msg, "v0.1.7") || !contains(msg, "v0.1.6") {
		t.Errorf("message missing expected content: %s", msg)
	}
}

func TestUserMessage_FirstRelease(t *testing.T) {
	msg := userMessage("axon-book", "", "v0.1.0")
	if !contains(msg, "first release") {
		t.Error("expected 'first release' for empty oldTag")
	}
}

func TestNewEngine(t *testing.T) {
	dir := t.TempDir()
	writer := &DryRunWriter{Dir: dir}
	// Pass nil client — just testing construction, not LLM calls
	engine := NewEngine(nil, "test-model", dir, writer)
	if engine.model != "test-model" {
		t.Errorf("expected model test-model, got %q", engine.model)
	}
	if len(engine.tools) != 6 {
		t.Errorf("expected 6 tools, got %d", len(engine.tools))
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsStr(s, sub))
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
