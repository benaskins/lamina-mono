# Beads Integration Plan

Integrate Beads (`bd`) into Lamina so that agents working across the workspace have persistent work memory — task tracking that survives session boundaries and spans sub-repos.

## Design Principles

- **CLI-first**: Beads has no public Go library. Shell out to `bd` with `--json`, same pattern Lamina uses for git.
- **Per-repo beads**: Each sub-repo gets its own `.beads/` (issues are repo-scoped). Lamina provides the aggregate view.
- **Opt-in**: Not every repo needs beads. Commands gracefully skip repos without `.beads/`.

## Steps

### 1. `lamina init` — add beads bootstrapping

Extend `lamina init` to optionally run `bd init` in each cloned repo.

- Add `--beads` flag to `lamina init`
- After cloning, run `bd init --contributor` in each repo (contributor mode keeps beads data off main branch)
- Skip repos that already have `.beads/`
- Requires `bd` on PATH — warn and continue if missing

**One file**: `cmd/lamina/init.go`

### 2. `lamina beads` — aggregate task view

New command that iterates sub-repos and aggregates beads output.

```
lamina beads                  # Show ready tasks across all repos
lamina beads list             # List all open tasks across repos
lamina beads ready            # Ready tasks (default)
lamina beads stats            # Per-repo task statistics
lamina beads sync             # Sync all repos
```

Implementation:
- Walk repos with `findRepos()`, filter to those with `.beads/`
- Run `bd <subcommand> --json` in each, parse JSON output
- Aggregate and display with repo name prefix on each task ID
- Support `--json` for machine-readable output (consistent with other lamina commands)

**New file**: `cmd/lamina/beads.go`

### 3. `lamina doctor` — add beads health check

Add a diagnostic check to the existing doctor command:

- `checkBeadsHealth()` — for each repo with `.beads/`, run `bd doctor --json`
- Report issues: unsynced repos, stale databases, missing migrations
- Surface repos that have beads tasks but haven't been synced recently

**Edit**: `cmd/lamina/doctor.go`

### 4. Beads skill for Claude Code

New skill that teaches agents how to use beads within the lamina workspace.

Contents:
- When starting a session, run `lamina beads ready` to see available work
- How to claim tasks (`bd update <id> --claim`)
- How to file new tasks when discovering work
- Land-the-plane protocol: update beads before ending a session
- Cross-repo awareness: if a task in axon-chat is blocked on axon-tool, use `bd dep add`

**New files**: `skills/beads-workflow/SKILL.md`

### 5. CLAUDE.md update

Add a "Beads" section to the root CLAUDE.md:
- One-liner: "Use `lamina beads ready` at session start to find work"
- Reference the beads-workflow skill for full protocol
