package main

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestAppendToReposYAML(t *testing.T) {
	dir := t.TempDir()

	// Write initial repos.yaml
	initial := reposFile{
		Repos: []repo{
			{Name: "axon", URL: "https://github.com/benaskins/axon.git"},
			{Name: "axon-chat", URL: "https://github.com/benaskins/axon-chat.git"},
		},
	}
	data, _ := yaml.Marshal(&initial)
	os.WriteFile(filepath.Join(dir, "repos.yaml"), data, 0644)

	// Append a new repo
	err := appendToReposYAML(dir, repo{
		Name: "axon-face",
		URL:  "https://github.com/benaskins/axon-face.git",
	})
	if err != nil {
		t.Fatalf("appendToReposYAML: %v", err)
	}

	// Read back and verify
	result, _ := os.ReadFile(filepath.Join(dir, "repos.yaml"))
	var f reposFile
	if err := yaml.Unmarshal(result, &f); err != nil {
		t.Fatalf("parse result: %v", err)
	}

	if len(f.Repos) != 3 {
		t.Fatalf("expected 3 repos, got %d", len(f.Repos))
	}

	// Should be sorted
	expected := []string{"axon", "axon-chat", "axon-face"}
	for i, name := range expected {
		if f.Repos[i].Name != name {
			t.Errorf("repos[%d] = %q, want %q", i, f.Repos[i].Name, name)
		}
	}
}

func TestAppendToReposYAML_CreatesFile(t *testing.T) {
	dir := t.TempDir()

	err := appendToReposYAML(dir, repo{
		Name: "axon-face",
		URL:  "https://github.com/benaskins/axon-face.git",
	})
	if err != nil {
		t.Fatalf("appendToReposYAML: %v", err)
	}

	result, _ := os.ReadFile(filepath.Join(dir, "repos.yaml"))
	var f reposFile
	yaml.Unmarshal(result, &f)

	if len(f.Repos) != 1 || f.Repos[0].Name != "axon-face" {
		t.Errorf("unexpected repos: %+v", f.Repos)
	}
}

func TestSortReposByName(t *testing.T) {
	repos := []repo{
		{Name: "axon-chat"},
		{Name: "aurelia"},
		{Name: "axon"},
		{Name: "axon-face"},
	}
	sortReposByName(repos)

	expected := []string{"aurelia", "axon", "axon-chat", "axon-face"}
	for i, name := range expected {
		if repos[i].Name != name {
			t.Errorf("repos[%d] = %q, want %q", i, repos[i].Name, name)
		}
	}
}
