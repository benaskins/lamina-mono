version := `git describe --tags --always --dirty 2>/dev/null || echo dev`

# Clone sub-repos from repos.yaml (no lamina binary required)
bootstrap:
    #!/usr/bin/env bash
    set -euo pipefail
    name="" url=""
    pids=()
    while IFS= read -r line; do
        if [[ "$line" =~ name:\ (.+) ]]; then name="${BASH_REMATCH[1]}"; fi
        if [[ "$line" =~ url:\ (.+) ]]; then url="${BASH_REMATCH[1]}"; fi
        if [ -n "$name" ] && [ -n "$url" ]; then
            if [ -d "$name/.git" ]; then
                echo "  [ok] $name"
            else
                echo "  [clone] $name"
                git clone "$url" "$name" &
                pids+=($!)
            fi
            name="" url=""
        fi
    done < repos.yaml
    failed=0
    for pid in "${pids[@]+"${pids[@]}"}"; do
        wait "$pid" || ((failed++))
    done
    if [ "$failed" -gt 0 ]; then
        echo "Bootstrap finished with $failed failed clone(s)"
        exit 1
    fi
    echo "Bootstrap complete — run 'just install' next"

build:
    go build -ldflags "-X main.version={{version}}" -o bin/lamina ./cmd/lamina/

install: build
    cp bin/lamina ~/.local/bin/lamina
    @echo "Installed lamina {{version}}"

test:
    go vet ./cmd/lamina/
    bin/lamina test

# Symlink skills from skills/ into .claude/skills/ for Claude Code discovery
install-skills:
    mkdir -p .claude/skills
    for dir in skills/*/; do \
        name=$(basename "$dir"); \
        ln -sfn "$(pwd)/$dir" ".claude/skills/$name"; \
    done
    @echo "Installed $(ls -1 skills/ | wc -l | tr -d ' ') skill(s) to .claude/skills/"

# Clone, build, and install flux.swift (FLUX.1 image generation on Apple Silicon)
install-flux:
    #!/usr/bin/env bash
    set -euo pipefail
    FLUX_DIR="$HOME/.local/src/flux.swift.cli"
    if [ ! -d "$FLUX_DIR" ]; then
        echo "Cloning flux.swift.cli..."
        mkdir -p "$HOME/.local/src"
        git clone https://github.com/mzbac/flux.swift.cli.git "$FLUX_DIR"
    else
        echo "Updating flux.swift.cli..."
        git -C "$FLUX_DIR" pull --ff-only
    fi
    echo "Building flux.swift.cli (release)..."
    cd "$FLUX_DIR"
    swift build -c release
    cp "$(swift build -c release --show-bin-path)/flux.swift.cli" "$HOME/.local/bin/flux"
    echo "Installed flux to ~/.local/bin/flux"
