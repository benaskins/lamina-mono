package main

import "testing"

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
