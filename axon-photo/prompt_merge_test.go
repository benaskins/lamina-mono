package photo_test

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"

	photo "github.com/benaskins/axon-photo"
	tool "github.com/benaskins/axon-tool"
)

func fakeGenerator(response string) tool.TextGenerator {
	return func(ctx context.Context, prompt string) (string, error) {
		return response, nil
	}
}

func captureGenerator(captured *string, response string) tool.TextGenerator {
	return func(ctx context.Context, prompt string) (string, error) {
		*captured = prompt
		return response, nil
	}
}

func TestPromptMerger_MergePrompt(t *testing.T) {
	config := &photo.ImageGenConfig{
		BaselinePrompt:   "baseline rules here",
		MergeInstruction: "Merge {baseline} with {system_prompt} and {conversation} for {scene}",
	}

	var captured string
	gen := captureGenerator(&captured, "merged prompt result")
	merger := photo.NewPromptMerger(gen, config)

	messages := []photo.Message{
		{Role: "user", Content: "hello"},
		{Role: "assistant", Content: "hi there"},
	}

	result, err := merger.MergePrompt("you are helpful", messages, "a sunset")
	if err != nil {
		t.Fatal(err)
	}

	if result != "merged prompt result" {
		t.Errorf("result = %q, want %q", result, "merged prompt result")
	}

	// Verify placeholders were replaced in the instruction sent to the generator
	if !strings.Contains(captured, "baseline rules here") {
		t.Error("expected {baseline} to be replaced with baseline prompt")
	}
	if !strings.Contains(captured, "you are helpful") {
		t.Error("expected {system_prompt} to be replaced")
	}
	if !strings.Contains(captured, "user: hello") {
		t.Error("expected conversation context to include user message")
	}
	if !strings.Contains(captured, "a sunset") {
		t.Error("expected {scene} to be replaced")
	}
}

func TestPromptMerger_MergePromptPrivate_UsesPrivateConfig(t *testing.T) {
	config := &photo.ImageGenConfig{
		BaselinePrompt:        "standard baseline",
		PrivateBaselinePrompt: "private baseline",
		MergeInstruction:      "standard merge {baseline}",
		PrivateMergeInstr:     "private merge {baseline}",
	}

	var captured string
	gen := captureGenerator(&captured, "private result")
	merger := photo.NewPromptMerger(gen, config)

	result, err := merger.MergePromptPrivate("sys", nil, "scene")
	if err != nil {
		t.Fatal(err)
	}

	if result != "private result" {
		t.Errorf("result = %q, want %q", result, "private result")
	}
	if !strings.Contains(captured, "private baseline") {
		t.Error("expected private baseline to be used")
	}
	if !strings.Contains(captured, "private merge") {
		t.Error("expected private merge instruction to be used")
	}
}

func TestPromptMerger_MergePromptPrivate_FallsBackToStandard(t *testing.T) {
	config := &photo.ImageGenConfig{
		BaselinePrompt:   "standard baseline",
		MergeInstruction: "standard merge {baseline}",
		// No private variants set
	}

	var captured string
	gen := captureGenerator(&captured, "fallback result")
	merger := photo.NewPromptMerger(gen, config)

	_, err := merger.MergePromptPrivate("sys", nil, "scene")
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(captured, "standard baseline") {
		t.Error("expected fallback to standard baseline")
	}
	if !strings.Contains(captured, "standard merge") {
		t.Error("expected fallback to standard merge instruction")
	}
}

func TestPromptMerger_TrimsWhitespace(t *testing.T) {
	gen := fakeGenerator("  result with spaces  \n")
	merger := photo.NewPromptMerger(gen, &photo.ImageGenConfig{
		MergeInstruction: "{scene}",
	})

	result, err := merger.MergePrompt("", nil, "test")
	if err != nil {
		t.Fatal(err)
	}
	if result != "result with spaces" {
		t.Errorf("result = %q, want trimmed", result)
	}
}

func TestPromptMerger_EmptyMessages(t *testing.T) {
	var captured string
	gen := captureGenerator(&captured, "ok")
	merger := photo.NewPromptMerger(gen, &photo.ImageGenConfig{
		MergeInstruction: "conv={conversation}",
	})

	_, err := merger.MergePrompt("", nil, "test")
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(captured, "conv=") {
		t.Error("expected conversation placeholder to be replaced even when empty")
	}
}

func TestLoadImageGenConfig(t *testing.T) {
	config := photo.ImageGenConfig{
		BaselinePrompt:   "base",
		MergeInstruction: "merge {baseline} {scene}",
		MergeModel:       "llama3",
	}
	data, _ := json.Marshal(config)

	f, err := os.CreateTemp("", "imgconfig-*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.Write(data)
	f.Close()

	loaded, err := photo.LoadImageGenConfig(f.Name())
	if err != nil {
		t.Fatal(err)
	}
	if loaded.BaselinePrompt != "base" {
		t.Errorf("BaselinePrompt = %q, want %q", loaded.BaselinePrompt, "base")
	}
	if loaded.MergeModel != "llama3" {
		t.Errorf("MergeModel = %q, want %q", loaded.MergeModel, "llama3")
	}
}

func TestLoadImageGenConfig_FileNotFound(t *testing.T) {
	_, err := photo.LoadImageGenConfig("/nonexistent/file.json")
	if err == nil {
		t.Error("expected error for missing file")
	}
}
