package lamina

import _ "embed"

// DefaultReposYAML is the embedded repos.yaml used as a fallback
// when the file doesn't exist on disk.
//
//go:embed repos.yaml
var DefaultReposYAML []byte
