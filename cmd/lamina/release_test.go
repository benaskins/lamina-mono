package main

import (
	"strings"
	"testing"
)

func TestValidateVersion(t *testing.T) {
	valid := []string{
		"v0.1.0",
		"v1.0.0",
		"v1.2.3",
		"v10.20.30",
		"v1.2.3-beta.1",
		"v0.0.1-rc1",
		"v1.0.0-alpha.2.3",
	}
	for _, v := range valid {
		if err := validateVersion(v); err != nil {
			t.Errorf("expected %q to be valid, got error: %v", v, err)
		}
	}

	invalid := []string{
		"1.0.0",              // missing v prefix
		"v1.0",               // missing patch
		"v1",                 // missing minor and patch
		"latest",             // not semver
		"v1.0.0; rm -rf /",  // injection attempt
		"v1.0.0 && echo hi", // injection attempt
		"",                   // empty
		"v1.0.0-",            // trailing hyphen
		"v1.0.0-béta",        // non-ASCII
	}
	for _, v := range invalid {
		if err := validateVersion(v); err == nil {
			t.Errorf("expected %q to be invalid, got nil error", v)
		}
	}
}

func TestGenerateReleaseNotes(t *testing.T) {
	log := `583b75e feat: add GET /v1/system endpoint
dce46b7 feat: add internal/sysinfo package
761a203 fix: short-lived transport connections for peer liveness
a73998d fix: close idle connections after failed peer health check
c0c3b2d docs: update security.md with TLS docs
2e246ba test: add fault tolerance integration tests
aa92ecd infra: add Dockerfile for multi-node tests
ea97a9d refactor: extract monitor phase logic`

	notes := generateReleaseNotes("v0.3.0", log)

	// Should have a heading
	if !strings.Contains(notes, "v0.3.0") {
		t.Error("expected notes to contain version")
	}

	// Features grouped
	if !strings.Contains(notes, "### Features") {
		t.Error("expected Features section")
	}
	if !strings.Contains(notes, "add GET /v1/system endpoint") {
		t.Error("expected feature commit in Features section")
	}

	// Fixes grouped
	if !strings.Contains(notes, "### Fixes") {
		t.Error("expected Fixes section")
	}
	if !strings.Contains(notes, "close idle connections") {
		t.Error("expected fix commit in Fixes section")
	}

	// Other section for docs/test/infra/refactor
	if !strings.Contains(notes, "### Other") {
		t.Error("expected Other section")
	}

	// Sections should appear in order: Features, Fixes, Other
	featIdx := strings.Index(notes, "### Features")
	fixIdx := strings.Index(notes, "### Fixes")
	otherIdx := strings.Index(notes, "### Other")
	if featIdx > fixIdx {
		t.Error("Features should come before Fixes")
	}
	if fixIdx > otherIdx {
		t.Error("Fixes should come before Other")
	}
}

func TestGenerateReleaseNotes_OnlyFeatures(t *testing.T) {
	log := `abc1234 feat: add new endpoint
def5678 feat: add new type`

	notes := generateReleaseNotes("v0.1.0", log)

	if !strings.Contains(notes, "### Features") {
		t.Error("expected Features section")
	}
	if strings.Contains(notes, "### Fixes") {
		t.Error("should not have Fixes section when there are no fixes")
	}
	if strings.Contains(notes, "### Other") {
		t.Error("should not have Other section when there are no other commits")
	}
}

func TestGenerateReleaseNotes_EmptyLog(t *testing.T) {
	notes := generateReleaseNotes("v0.1.0", "")

	if !strings.Contains(notes, "v0.1.0") {
		t.Error("expected notes to contain version even with empty log")
	}
}
