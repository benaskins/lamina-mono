package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

var healCmd = &cobra.Command{
	Use:   "heal",
	Short: "Fix issues found by doctor",
	Long: `Automatically fix healable issues in the workspace.

Can run standalone or consume doctor's JSON output via pipe:
  lamina doctor --json | lamina heal
  lamina heal                         # runs doctor internally
  lamina heal --dry-run               # show what would be done`,
	RunE: runHeal,
}

func init() {
	healCmd.Flags().Bool("dry-run", false, "Show what would be done without doing it")
	rootCmd.AddCommand(healCmd)
}

func runHeal(cmd *cobra.Command, args []string) error {
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	root, err := workspaceRoot()
	if err != nil {
		return err
	}

	diags, err := loadDiagnostics()
	if err != nil {
		return err
	}

	// Separate release actions (need dependency ordering) from other actions
	var releaseActions []Diagnostic
	var otherActions []Diagnostic
	for _, d := range diags {
		switch d.Kind {
		case "untagged", "ahead-of-tag":
			releaseActions = append(releaseActions, d)
		case "agent-docs-missing":
			otherActions = append(otherActions, d)
		}
	}

	if len(releaseActions) == 0 && len(otherActions) == 0 {
		fmt.Println("Nothing to heal")
		return nil
	}

	// Non-release actions first (e.g. docs) — no ordering needed
	for _, d := range otherActions {
		if err := healAgentDocs(d, dryRun); err != nil {
			fmt.Printf("  FAIL %s: %v\n", d.Name, err)
		}
	}

	// Release actions in dependency order so libraries are tagged before dependents
	if len(releaseActions) > 0 {
		ordered := orderReleaseActions(root, releaseActions)
		for _, d := range ordered {
			switch d.Kind {
			case "untagged":
				if err := healUntagged(cmd.Context(), root, d, dryRun); err != nil {
					fmt.Printf("  FAIL %s: %v\n", d.Name, err)
				}
			case "ahead-of-tag":
				if err := healAheadOfTag(cmd.Context(), root, d, dryRun); err != nil {
					fmt.Printf("  FAIL %s: %v\n", d.Name, err)
				}
			}
		}
	}

	return nil
}

func loadDiagnostics() ([]Diagnostic, error) {
	// Check if stdin has data (piped)
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		// Reading from pipe
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return nil, fmt.Errorf("reading stdin: %w", err)
		}
		var diags []Diagnostic
		if err := json.Unmarshal(data, &diags); err != nil {
			return nil, fmt.Errorf("parsing diagnostics JSON: %w", err)
		}
		return diags, nil
	}

	// No pipe — run doctor internally
	root, err := workspaceRoot()
	if err != nil {
		return nil, err
	}
	return runDiagnostics(root), nil
}

// repoNameFromDiagnostic extracts the repo name from a diagnostic Name field
// (e.g., "imago tags" → "imago").
func repoNameFromDiagnostic(d Diagnostic) string {
	return strings.TrimSuffix(d.Name, " tags")
}

func healUntagged(ctx context.Context, root string, d Diagnostic, dryRun bool) error {
	if d.Dir == "" {
		return fmt.Errorf("no directory in diagnostic")
	}

	tag := "v0.1.0"
	name := repoNameFromDiagnostic(d)
	return releaseOne(ctx, root, name, tag, dryRun)
}

func healAheadOfTag(ctx context.Context, root string, d Diagnostic, dryRun bool) error {
	if d.Dir == "" {
		return fmt.Errorf("no directory in diagnostic")
	}
	if d.LatestTag == "" {
		return fmt.Errorf("no latest tag in diagnostic")
	}

	name := repoNameFromDiagnostic(d)

	// Determine bump kind from commit history since last tag
	log := gitOutput(d.Dir, "log", d.LatestTag+"..HEAD", "--oneline")
	bump := inferBumpKind(log)

	var nextTag string
	var err error
	if bump == "minor" {
		nextTag, err = bumpMinor(d.LatestTag)
	} else {
		nextTag, err = bumpPatch(d.LatestTag)
	}
	if err != nil {
		return fmt.Errorf("computing next version: %w", err)
	}

	// Keep bumping if the computed tag already exists
	for {
		existing := gitOutput(d.Dir, "tag", "-l", nextTag)
		if existing == "" {
			break
		}
		nextTag, err = bumpPatch(nextTag)
		if err != nil {
			return fmt.Errorf("computing next version: %w", err)
		}
	}

	return releaseOne(ctx, root, name, nextTag, dryRun)
}

const claudeMDPointer = "# CLAUDE.md\n\nRead [AGENTS.md](./AGENTS.md) for project context.\n"

func healAgentDocs(d Diagnostic, dryRun bool) error {
	if d.Dir == "" {
		return fmt.Errorf("no directory in diagnostic")
	}

	name := strings.TrimSuffix(d.Name, " docs")
	claudePath := d.Dir + "/CLAUDE.md"
	agentsPath := d.Dir + "/AGENTS.md"

	// Create AGENTS.md if missing
	if _, err := os.Stat(agentsPath); err != nil {
		if dryRun {
			fmt.Printf("  [dry-run] would create %s/AGENTS.md\n", name)
		} else {
			content := fmt.Sprintf("# %s\n\n## Build & Test\n\n```bash\ngo test ./...\ngo vet ./...\n```\n", name)
			if err := os.WriteFile(agentsPath, []byte(content), 0644); err != nil {
				return fmt.Errorf("writing AGENTS.md: %w", err)
			}
			fmt.Printf("  Created %s/AGENTS.md\n", name)
		}
	}

	// Create or fix CLAUDE.md
	needsClaude := false
	if data, err := os.ReadFile(claudePath); err != nil {
		needsClaude = true
	} else if !strings.Contains(string(data), "AGENTS.md") {
		needsClaude = true
	}

	if needsClaude {
		if dryRun {
			fmt.Printf("  [dry-run] would create %s/CLAUDE.md pointing to AGENTS.md\n", name)
		} else {
			if err := os.WriteFile(claudePath, []byte(claudeMDPointer), 0644); err != nil {
				return fmt.Errorf("writing CLAUDE.md: %w", err)
			}
			fmt.Printf("  Created %s/CLAUDE.md\n", name)
		}
	}

	return nil
}

// orderReleaseActions topologically sorts release diagnostics so dependencies
// are released before the modules that depend on them.
func orderReleaseActions(root string, diags []Diagnostic) []Diagnostic {
	// Index diagnostics by repo name
	byName := make(map[string]Diagnostic)
	var modules []releaseModule
	for _, d := range diags {
		name := repoNameFromDiagnostic(d)
		byName[name] = d

		// Read workspace deps from go.mod
		modPath := filepath.Join(d.Dir, "go.mod")
		deps := workspaceDeps(modPath)
		modules = append(modules, releaseModule{name: name, deps: deps})
	}

	ordered := topoSort(modules)

	var result []Diagnostic
	for _, name := range ordered {
		if d, ok := byName[name]; ok {
			result = append(result, d)
		}
	}
	return result
}

// inferBumpKind reads a git log (oneline format) and returns "minor" if any
// commit is a feat: or refactor:, otherwise "patch".
func inferBumpKind(log string) string {
	for _, line := range strings.Split(log, "\n") {
		// Skip the commit hash prefix to find the conventional commit type
		parts := strings.SplitN(line, " ", 2)
		if len(parts) < 2 {
			continue
		}
		msg := parts[1]
		if strings.HasPrefix(msg, "feat:") || strings.HasPrefix(msg, "feat(") ||
			strings.HasPrefix(msg, "refactor:") || strings.HasPrefix(msg, "refactor(") {
			return "minor"
		}
	}
	return "patch"
}

// bumpMinor increments the minor version and resets patch: v0.3.1 -> v0.4.0
func bumpMinor(tag string) (string, error) {
	v := strings.TrimPrefix(tag, "v")
	parts := strings.Split(v, ".")
	if len(parts) != 3 {
		return "", fmt.Errorf("unexpected version format: %s", tag)
	}
	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return "", fmt.Errorf("parsing minor version: %w", err)
	}
	return fmt.Sprintf("v%s.%d.0", parts[0], minor+1), nil
}

// bumpPatch increments the patch version: v0.3.0 -> v0.3.1
func bumpPatch(tag string) (string, error) {
	v := strings.TrimPrefix(tag, "v")
	parts := strings.Split(v, ".")
	if len(parts) != 3 {
		return "", fmt.Errorf("unexpected version format: %s", tag)
	}
	patch, err := strconv.Atoi(parts[2])
	if err != nil {
		return "", fmt.Errorf("parsing patch version: %w", err)
	}
	return fmt.Sprintf("v%s.%s.%d", parts[0], parts[1], patch+1), nil
}
