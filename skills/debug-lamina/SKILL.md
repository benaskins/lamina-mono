---
name: debug-lamina
description: Use when a build fails, tests break across repos, or dependencies are out of sync. Extends /debug with lamina-specific tools.
---

# Debug Lamina

Follow `/debug` — here's how each step works in the lamina workspace.

## Reproducing

```bash
lamina test                     # Run tests across all axon-* libraries
lamina test <repo>              # Test a specific module
lamina doctor                   # Check workspace health — stale deps, version mismatches
```

## Common hypotheses

| Symptom | Start here |
|---|---|
| Build failure in a service | Stale dependency — `lamina deps` to trace the chain |
| Tests pass locally but fail in service | Service go.mod has stale `replace` directives |
| Module not found | Check `lamina repo` — is it cloned? Does it have a go.mod? |
| Version mismatch | `lamina doctor` flags inconsistent versions across services |

## Fixing dependency issues

1. `lamina deps --json` to see the full chain
2. Check the service's `go.mod` for stale `replace` directives
3. `go mod tidy` in the service directory
4. `lamina test <library>` to verify the library itself passes
5. `lamina doctor` to confirm clean

Then `/verify`.

$ARGUMENTS
