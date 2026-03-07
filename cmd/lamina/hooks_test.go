package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestHooksCmd_InstallsPreCommitHook(t *testing.T) {
	dir := t.TempDir()

	// Initialize a git repo
	if err := exec.Command("git", "init", dir).Run(); err != nil {
		t.Fatalf("git init: %v", err)
	}

	err := installHooks(dir)
	if err != nil {
		t.Fatalf("installHooks: %v", err)
	}

	hookPath := filepath.Join(dir, ".git", "hooks", "pre-commit")
	data, err := os.ReadFile(hookPath)
	if err != nil {
		t.Fatalf("read hook: %v", err)
	}

	content := string(data)

	// Hook should be a bash script
	if !strings.HasPrefix(content, "#!/usr/bin/env bash") {
		t.Error("hook should start with bash shebang")
	}

	// Hook should run go vet
	if !strings.Contains(content, "go vet") {
		t.Error("hook should run go vet")
	}

	// Hook should check coverage
	if !strings.Contains(content, "coverprofile") {
		t.Error("hook should run coverage")
	}

	// Hook should run slop-guard (optional)
	if !strings.Contains(content, "slop-guard") {
		t.Error("hook should include slop-guard")
	}

	// Hook should be executable
	info, _ := os.Stat(hookPath)
	if info.Mode()&0111 == 0 {
		t.Error("hook should be executable")
	}
}

func TestHooksCmd_ReadsThreshold(t *testing.T) {
	dir := t.TempDir()
	exec.Command("git", "init", dir).Run()

	// Write a threshold file
	os.WriteFile(filepath.Join(dir, ".coverage-threshold"), []byte("65\n"), 0644)

	err := installHooks(dir)
	if err != nil {
		t.Fatalf("installHooks: %v", err)
	}

	hookPath := filepath.Join(dir, ".git", "hooks", "pre-commit")
	data, _ := os.ReadFile(hookPath)
	content := string(data)

	// Hook should reference the threshold
	if !strings.Contains(content, ".coverage-threshold") {
		t.Error("hook should read .coverage-threshold")
	}
}

func TestHooksCmd_NoGitDir(t *testing.T) {
	dir := t.TempDir()

	err := installHooks(dir)
	if err == nil {
		t.Error("expected error when no .git directory")
	}
}
