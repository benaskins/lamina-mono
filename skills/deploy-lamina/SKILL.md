---
name: deploy-lamina
description: Use when releasing library changes or deploying services across the lamina workspace. Extends /deploy with lamina-specific tools.
---

# Deploy Lamina

Follow `/deploy` — here's how each step works in the lamina workspace.

## Releasing libraries

```bash
lamina release <module> <version>       # Tag and push a module
lamina release --dry-run <module> v0.2.0  # Preview first
lamina release --all                    # Release in dependency order
```

`lamina release` checks for dirty state, existing tags, and warns if dependencies have unpublished changes.

## Pre-flight

```bash
lamina doctor               # Check workspace health before releasing
lamina test                 # Run full test suite across all modules
lamina repo                 # Verify all repos are clean
```

## Deploying services

After a library release, services that depend on it need redeployment. Use `/deploy-aurelia` for the service deployment itself. The lamina layer handles:

1. `lamina doctor` — confirm dependencies are consistent
2. `lamina test` — full suite passes
3. `lamina release` — tag the library
4. Then `/deploy-aurelia` for each affected service

Then `/verify`.

$ARGUMENTS
