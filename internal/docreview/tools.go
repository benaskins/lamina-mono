package docreview

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	tool "github.com/benaskins/axon-tool"
)

// Writer abstracts file writing so we can swap in dry-run mode.
type Writer interface {
	Write(path, content string) error
}

// FileWriter writes files to disk.
type FileWriter struct {
	Dir string
}

func (w *FileWriter) Write(path, content string) error {
	full := filepath.Join(w.Dir, path)
	return os.WriteFile(full, []byte(content), 0644)
}

// Change records a proposed file modification.
type Change struct {
	Path    string
	Content string
}

// DryRunWriter captures proposed changes without writing to disk.
type DryRunWriter struct {
	Dir     string
	Changes []Change
}

func (w *DryRunWriter) Write(path, content string) error {
	w.Changes = append(w.Changes, Change{Path: path, Content: content})
	return nil
}

// Tools returns the doc review tool set for a given repo directory.
func Tools(dir string, writer Writer) map[string]tool.ToolDef {
	return map[string]tool.ToolDef{
		"read_file":    readFileTool(dir),
		"list_files":   listFilesTool(dir),
		"list_exports": listExportsTool(dir),
		"read_go_mod":  readGoModTool(dir),
		"git_log":      gitLogTool(dir),
		"write_file":   writeFileTool(writer),
	}
}

func readFileTool(dir string) tool.ToolDef {
	return tool.ToolDef{
		Name:        "read_file",
		Description: "Read the contents of a file in the repository.",
		Parameters: tool.ParameterSchema{
			Type:     "object",
			Required: []string{"path"},
			Properties: map[string]tool.PropertySchema{
				"path": {Type: "string", Description: "Relative path to the file (e.g. README.md, AGENTS.md)"},
			},
		},
		Execute: func(ctx *tool.ToolContext, args map[string]any) tool.ToolResult {
			path, _ := args["path"].(string)
			full := filepath.Join(dir, path)
			// Prevent path traversal
			rel, err := filepath.Rel(dir, full)
			if err != nil || strings.HasPrefix(rel, "..") {
				return tool.ToolResult{Content: "error: path traversal not allowed"}
			}
			data, err := os.ReadFile(full)
			if err != nil {
				return tool.ToolResult{Content: fmt.Sprintf("error: %v", err)}
			}
			return tool.ToolResult{Content: string(data)}
		},
	}
}

func listFilesTool(dir string) tool.ToolDef {
	return tool.ToolDef{
		Name:        "list_files",
		Description: "List all files in the repository, showing the directory structure.",
		Parameters: tool.ParameterSchema{
			Type:       "object",
			Properties: map[string]tool.PropertySchema{},
		},
		Execute: func(ctx *tool.ToolContext, args map[string]any) tool.ToolResult {
			cmd := exec.Command("find", ".", "-type", "f", "-not", "-path", "./.git/*", "-not", "-path", "./vendor/*")
			cmd.Dir = dir
			out, err := cmd.Output()
			if err != nil {
				return tool.ToolResult{Content: fmt.Sprintf("error: %v", err)}
			}
			return tool.ToolResult{Content: string(out)}
		},
	}
}

func listExportsTool(dir string) tool.ToolDef {
	return tool.ToolDef{
		Name:        "list_exports",
		Description: "Run 'go doc' on a package to list exported types and functions.",
		Parameters: tool.ParameterSchema{
			Type:     "object",
			Required: []string{"package"},
			Properties: map[string]tool.PropertySchema{
				"package": {Type: "string", Description: "Package path relative to the module root (e.g. '.' for root package, './gl' for a sub-package)"},
			},
		},
		Execute: func(ctx *tool.ToolContext, args map[string]any) tool.ToolResult {
			pkg, _ := args["package"].(string)
			cmd := exec.Command("go", "doc", "-all", pkg)
			cmd.Dir = dir
			out, err := cmd.CombinedOutput()
			if err != nil {
				return tool.ToolResult{Content: fmt.Sprintf("error: %s\n%s", err, out)}
			}
			return tool.ToolResult{Content: string(out)}
		},
	}
}

func readGoModTool(dir string) tool.ToolDef {
	return tool.ToolDef{
		Name:        "read_go_mod",
		Description: "Read the go.mod file to check the module path, Go version, and dependencies.",
		Parameters: tool.ParameterSchema{
			Type:       "object",
			Properties: map[string]tool.PropertySchema{},
		},
		Execute: func(ctx *tool.ToolContext, args map[string]any) tool.ToolResult {
			data, err := os.ReadFile(filepath.Join(dir, "go.mod"))
			if err != nil {
				return tool.ToolResult{Content: fmt.Sprintf("error: %v", err)}
			}
			return tool.ToolResult{Content: string(data)}
		},
	}
}

func gitLogTool(dir string) tool.ToolDef {
	return tool.ToolDef{
		Name:        "git_log",
		Description: "Show git log for a given range (e.g. 'v0.2.0..v0.3.0' or 'HEAD~10..HEAD').",
		Parameters: tool.ParameterSchema{
			Type:     "object",
			Required: []string{"range"},
			Properties: map[string]tool.PropertySchema{
				"range": {Type: "string", Description: "Git revision range (e.g. 'v0.1.0..v0.2.0')"},
			},
		},
		Execute: func(ctx *tool.ToolContext, args map[string]any) tool.ToolResult {
			revRange, _ := args["range"].(string)
			cmd := exec.Command("git", "log", "--oneline", revRange)
			cmd.Dir = dir
			out, err := cmd.CombinedOutput()
			if err != nil {
				return tool.ToolResult{Content: fmt.Sprintf("error: %s\n%s", err, out)}
			}
			return tool.ToolResult{Content: string(out)}
		},
	}
}

func writeFileTool(writer Writer) tool.ToolDef {
	return tool.ToolDef{
		Name:        "write_file",
		Description: "Write content to a file. Use this to update README.md or AGENTS.md after reviewing the documentation.",
		Parameters: tool.ParameterSchema{
			Type:     "object",
			Required: []string{"path", "content"},
			Properties: map[string]tool.PropertySchema{
				"path":    {Type: "string", Description: "Relative path to write (e.g. README.md)"},
				"content": {Type: "string", Description: "Full file content to write"},
			},
		},
		Execute: func(ctx *tool.ToolContext, args map[string]any) tool.ToolResult {
			path, _ := args["path"].(string)
			content, _ := args["content"].(string)
			if err := writer.Write(path, content); err != nil {
				return tool.ToolResult{Content: fmt.Sprintf("error: %v", err)}
			}
			return tool.ToolResult{Content: fmt.Sprintf("wrote %s (%d bytes)", path, len(content))}
		},
	}
}
