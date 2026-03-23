package docreview

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tool "github.com/benaskins/axon-tool"
)

func TestReadFileTool(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "README.md"), []byte("# Hello\n\nThis is a test."), 0644)

	td := readFileTool(dir)
	result := td.Execute(&tool.ToolContext{}, map[string]any{"path": "README.md"})
	if strings.HasPrefix(result.Content, "error:") {
		t.Fatalf("unexpected error: %s", result.Content)
	}
	if result.Content != "# Hello\n\nThis is a test." {
		t.Errorf("unexpected content: %q", result.Content)
	}
}

func TestReadFileTool_PathTraversal(t *testing.T) {
	dir := t.TempDir()
	td := readFileTool(dir)
	result := td.Execute(&tool.ToolContext{}, map[string]any{"path": "../../../etc/passwd"})
	if !strings.HasPrefix(result.Content, "error:") {
		t.Error("expected error for path traversal")
	}
}

func TestListExportsTool(t *testing.T) {
	td := listExportsTool("/Users/benaskins/dev/lamina/axon-tool")
	result := td.Execute(&tool.ToolContext{}, map[string]any{"package": "."})
	if strings.HasPrefix(result.Content, "error:") {
		t.Fatalf("unexpected error: %s", result.Content)
	}
	if len(result.Content) == 0 {
		t.Error("expected non-empty go doc output")
	}
}

func TestReadGoModTool(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/test\n\ngo 1.26.1\n"), 0644)

	td := readGoModTool(dir)
	result := td.Execute(&tool.ToolContext{}, map[string]any{})
	if strings.HasPrefix(result.Content, "error:") {
		t.Fatalf("unexpected error: %s", result.Content)
	}
	if result.Content != "module example.com/test\n\ngo 1.26.1\n" {
		t.Errorf("unexpected content: %q", result.Content)
	}
}

func TestGitLogTool(t *testing.T) {
	td := gitLogTool("/Users/benaskins/dev/lamina")
	result := td.Execute(&tool.ToolContext{}, map[string]any{"range": "HEAD~3..HEAD"})
	if strings.HasPrefix(result.Content, "error:") {
		t.Fatalf("unexpected error: %s", result.Content)
	}
	if len(result.Content) == 0 {
		t.Error("expected non-empty git log output")
	}
}

func TestWriteFileTool(t *testing.T) {
	dir := t.TempDir()
	writer := &FileWriter{Dir: dir}
	td := writeFileTool(writer)

	result := td.Execute(&tool.ToolContext{}, map[string]any{
		"path":    "README.md",
		"content": "# Updated\n\nNew content.",
	})
	if strings.HasPrefix(result.Content, "error:") {
		t.Fatalf("unexpected error: %s", result.Content)
	}

	data, err := os.ReadFile(filepath.Join(dir, "README.md"))
	if err != nil {
		t.Fatalf("file not written: %v", err)
	}
	if string(data) != "# Updated\n\nNew content." {
		t.Errorf("unexpected file content: %q", string(data))
	}
}

func TestWriteFileTool_DryRun(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "README.md"), []byte("old"), 0644)

	writer := &DryRunWriter{Dir: dir}
	td := writeFileTool(writer)

	result := td.Execute(&tool.ToolContext{}, map[string]any{
		"path":    "README.md",
		"content": "new",
	})
	if strings.HasPrefix(result.Content, "error:") {
		t.Fatalf("unexpected error: %s", result.Content)
	}

	// File should NOT be modified
	data, _ := os.ReadFile(filepath.Join(dir, "README.md"))
	if string(data) != "old" {
		t.Error("dry run should not modify files")
	}

	// Should have captured the change
	if len(writer.Changes) != 1 {
		t.Fatalf("expected 1 change, got %d", len(writer.Changes))
	}
	if writer.Changes[0].Path != "README.md" {
		t.Errorf("expected path README.md, got %q", writer.Changes[0].Path)
	}
}

func TestListFilesTool(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main"), 0644)
	os.WriteFile(filepath.Join(dir, "main_test.go"), []byte("package main"), 0644)
	os.Mkdir(filepath.Join(dir, "sub"), 0755)
	os.WriteFile(filepath.Join(dir, "sub", "helper.go"), []byte("package sub"), 0644)

	td := listFilesTool(dir)
	result := td.Execute(&tool.ToolContext{}, map[string]any{})
	if strings.HasPrefix(result.Content, "error:") {
		t.Fatalf("unexpected error: %s", result.Content)
	}
	if len(result.Content) == 0 {
		t.Error("expected non-empty file listing")
	}
}
