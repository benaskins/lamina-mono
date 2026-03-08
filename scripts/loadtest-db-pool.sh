#!/bin/bash
# loadtest-db-pool.sh — compare DB performance across pool sizes
#
# Usage: ./scripts/loadtest-db-pool.sh
#
# Requires: hey, tuned macOS limits

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
REPO_ROOT="$(dirname "$SCRIPT_DIR")"
BASE_PORT=9900
INSTANCES=4

if ! command -v hey &>/dev/null; then
    echo "Install hey: go install github.com/rakyll/hey@latest"
    exit 1
fi

echo "=== Building loadtest service ==="
(cd "$REPO_ROOT/examples/loadtest" && go build -o /tmp/axon-loadtest .)

run_pool_test() {
    local pool_size=$1
    local PIDS=()

    echo ""
    echo "############################################"
    echo "  Pool size: $pool_size per instance"
    echo "  Total connections: $((pool_size * INSTANCES))"
    echo "############################################"
    echo ""

    # Start instances
    for i in $(seq 0 $((INSTANCES - 1))); do
        port=$((BASE_PORT + i))
        PORT=$port DB_MAX_CONNS=$pool_size /tmp/axon-loadtest &
        PIDS+=($!)
    done
    sleep 2

    # Verify
    for i in $(seq 0 $((INSTANCES - 1))); do
        port=$((BASE_PORT + i))
        if ! curl -sf "http://localhost:$port/health" >/dev/null; then
            echo "ERROR: instance on :$port failed to start"
            for pid in "${PIDS[@]}"; do kill "$pid" 2>/dev/null || true; done
            wait 2>/dev/null
            return 1
        fi
    done

    # DB Write test (c=500, n=50000)
    echo "--- Writes (c=500, n=50000) ---"
    for i in $(seq 0 $((INSTANCES - 1))); do
        port=$((BASE_PORT + i))
        printf "  :$port  "
        hey -n 12500 -c 500 -q 0 -m POST \
            -H "Content-Type: application/json" -d '{"name":"pooltest"}' \
            "http://localhost:$port/api/items" 2>&1 | grep -E "Requests/sec|Average|Slowest" | tr '\n' ' '
        echo ""
    done

    # DB Read test (c=500, n=50000)
    echo "--- Reads (c=500, n=50000) ---"
    for i in $(seq 0 $((INSTANCES - 1))); do
        port=$((BASE_PORT + i))
        printf "  :$port  "
        hey -n 12500 -c 500 -q 0 -m GET \
            "http://localhost:$port/api/items" 2>&1 | grep -E "Requests/sec|Average|Slowest" | tr '\n' ' '
        echo ""
    done

    # Mixed test (concurrent)
    echo "--- Mixed concurrent (c=200, n=10000 writes + reads per instance) ---"
    local MIX_PIDS=()
    for i in $(seq 0 $((INSTANCES - 1))); do
        port=$((BASE_PORT + i))
        hey -n 2500 -c 200 -q 0 -m POST \
            -H "Content-Type: application/json" -d '{"name":"mixed"}' \
            "http://localhost:$port/api/items" > "/tmp/loadtest-mix-w-$port" 2>&1 &
        MIX_PIDS+=($!)
        hey -n 2500 -c 200 -q 0 -m GET \
            "http://localhost:$port/api/items" > "/tmp/loadtest-mix-r-$port" 2>&1 &
        MIX_PIDS+=($!)
    done
    wait "${MIX_PIDS[@]}" 2>/dev/null

    for i in $(seq 0 $((INSTANCES - 1))); do
        port=$((BASE_PORT + i))
        printf "  :$port W "
        grep -E "Requests/sec|Average|Slowest" "/tmp/loadtest-mix-w-$port" | tr '\n' ' '
        echo ""
        printf "  :$port R "
        grep -E "Requests/sec|Average|Slowest" "/tmp/loadtest-mix-r-$port" | tr '\n' ' '
        echo ""
    done

    # Stats
    echo "--- Pool stats ---"
    for i in $(seq 0 $((INSTANCES - 1))); do
        port=$((BASE_PORT + i))
        printf "  :$port  "
        curl -sf "http://localhost:$port/api/stats" 2>/dev/null
        echo ""
    done

    # Shutdown
    for pid in "${PIDS[@]}"; do kill "$pid" 2>/dev/null || true; done
    wait 2>/dev/null
    sleep 1
}

echo "=============================="
echo "  DB Pool Size Comparison"
echo "  Postgres max_connections: $(docker exec $(docker ps -q --filter ancestor=pgvector/pgvector:pg16 2>/dev/null | head -1) psql -U aurelia -t -c "SHOW max_connections;" 2>/dev/null | tr -d ' ')"
echo "=============================="

run_pool_test 25
run_pool_test 50
run_pool_test 100

echo ""
echo "=== System state ==="
echo "  TIME_WAIT sockets: $(netstat -an 2>/dev/null | grep -c TIME_WAIT || echo 'n/a')"
echo ""
