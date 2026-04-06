package main

import "testing"

func TestBumpPatch(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"v0.3.0", "v0.3.1"},
		{"v0.3.1", "v0.3.2"},
		{"v1.0.0", "v1.0.1"},
		{"v0.0.0", "v0.0.1"},
	}
	for _, tt := range tests {
		got, err := bumpPatch(tt.in)
		if err != nil {
			t.Errorf("bumpPatch(%q) error: %v", tt.in, err)
			continue
		}
		if got != tt.want {
			t.Errorf("bumpPatch(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestBumpMinor(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"v0.3.0", "v0.4.0"},
		{"v0.3.5", "v0.4.0"},
		{"v1.2.3", "v1.3.0"},
		{"v0.0.0", "v0.1.0"},
	}
	for _, tt := range tests {
		got, err := bumpMinor(tt.in)
		if err != nil {
			t.Errorf("bumpMinor(%q) error: %v", tt.in, err)
			continue
		}
		if got != tt.want {
			t.Errorf("bumpMinor(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestInferBumpKind(t *testing.T) {
	tests := []struct {
		name string
		log  string
		want string
	}{
		{
			name: "feat triggers minor",
			log:  "abc1234 feat: add new endpoint\ndef5678 fix: typo in error message",
			want: "minor",
		},
		{
			name: "refactor triggers minor",
			log:  "abc1234 refactor: rename ChatClient to LLMClient",
			want: "minor",
		},
		{
			name: "only fixes triggers patch",
			log:  "abc1234 fix: nil pointer in handler\ndef5678 fix: typo",
			want: "patch",
		},
		{
			name: "docs and config trigger patch",
			log:  "abc1234 docs: update README\ndef5678 config: add lint rule",
			want: "patch",
		},
		{
			name: "mixed with feat triggers minor",
			log:  "abc1234 fix: bug\ndef5678 docs: readme\nghi9012 feat: new API",
			want: "minor",
		},
		{
			name: "empty log triggers patch",
			log:  "",
			want: "patch",
		},
		{
			name: "non-conventional commits trigger patch",
			log:  "abc1234 some random commit message",
			want: "patch",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := inferBumpKind(tt.log)
			if got != tt.want {
				t.Errorf("inferBumpKind(%q) = %q, want %q", tt.log, got, tt.want)
			}
		})
	}
}
