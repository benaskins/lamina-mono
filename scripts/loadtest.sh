#!/bin/bash
# loadtest.sh — progressive load test for axon services
#
# Usage:
#   ./scripts/loadtest.sh              # single instance, echo only
#   ./scripts/loadtest.sh --scale 4    # 4 instances
#   ./scripts/loadtest.sh --db         # include database read/write tests
#   ./scripts/loadtest.sh --scale 4 --db
#
# Requires: hey (go install github.com/rakyll/hey@latest)

set -euo pipefail

BASE_PORT=9900
INSTANCES=1
WITH_DB=false
PIDS=()

while [[ $# -gt 0 ]]; do
    case "$1" in
        --scale) INSTANCES="$2"; shift 2 ;;
        --db) WITH_DB=true; shift ;;
        *) echo "Usage: $0 [--scale N] [--db]"; exit 1 ;;
    esac
done

if ! command -v hey &>/dev/null; then
    echo "Install hey: go install github.com/rakyll/hey@latest"
    exit 1
fi

cleanup() {
    echo ""
    echo "=== Shutting down ==="
    for pid in "${PIDS[@]}"; do
        kill "$pid" 2>/dev/null || true
    done
    wait 2>/dev/null
    echo "Done."
}
trap cleanup EXIT

# Build the loadtest service
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
REPO_ROOT="$(dirname "$SCRIPT_DIR")"

echo "=== Building loadtest service ==="
(cd "$REPO_ROOT/examples/loadtest" && go build -o /tmp/axon-loadtest .)

# Start instances
echo "=== Starting $INSTANCES instance(s) ==="
for i in $(seq 0 $((INSTANCES - 1))); do
    port=$((BASE_PORT + i))
    PORT=$port /tmp/axon-loadtest &
    PIDS+=($!)
    echo "  Instance $((i + 1)) on :$port (pid $!)"
done

sleep 2

# Verify all instances are up
for i in $(seq 0 $((INSTANCES - 1))); do
    port=$((BASE_PORT + i))
    if ! curl -sf "http://localhost:$port/health" >/dev/null; then
        echo "ERROR: instance on :$port failed to start"
        exit 1
    fi
done
echo "All instances healthy."
echo ""

# Snapshot CPU and memory for all instances
snapshot_resources() {
    for pid in "${PIDS[@]}"; do
        if kill -0 "$pid" 2>/dev/null; then
            ps -o pid=,pcpu=,rss= -p "$pid" 2>/dev/null
        fi
    done
}

print_resources() {
    local label=$1
    echo "  Resources $label:"
    printf "    %8s %6s %10s\n" "PID" "%CPU" "RSS(MB)"
    snapshot_resources | while read pid cpu rss; do
        printf "    %8s %6s %10s\n" "$pid" "$cpu" "$(echo "scale=1; $rss/1024" | bc)"
    done
}

# Run hey against all instances (or single)
run_hey() {
    local concurrency=$1
    local requests=$2
    local method=$3
    local path=$4
    local body="${5:-}"

    if [[ $INSTANCES -gt 1 ]]; then
        local per_instance=$((requests / INSTANCES))
        for i in $(seq 0 $((INSTANCES - 1))); do
            port=$((BASE_PORT + i))
            echo "  Instance :$port"
            if [[ -n "$body" ]]; then
                hey -n "$per_instance" -c "$concurrency" -q 0 -m "$method" \
                    -H "Content-Type: application/json" -d "$body" \
                    "http://localhost:$port$path" 2>&1 | grep -E "Requests/sec|Average|Fastest|Slowest|Status code"
            else
                hey -n "$per_instance" -c "$concurrency" -q 0 -m "$method" \
                    "http://localhost:$port$path" 2>&1 | grep -E "Requests/sec|Average|Fastest|Slowest|Status code"
            fi
        done
    else
        if [[ -n "$body" ]]; then
            hey -n "$requests" -c "$concurrency" -q 0 -m "$method" \
                -H "Content-Type: application/json" -d "$body" \
                "http://localhost:$BASE_PORT$path" 2>&1 | grep -E "Requests/sec|Average|Fastest|Slowest|Status code"
        else
            hey -n "$requests" -c "$concurrency" -q 0 -m "$method" \
                "http://localhost:$BASE_PORT$path" 2>&1 | grep -E "Requests/sec|Average|Fastest|Slowest|Status code"
        fi
    fi
}

# Run a full stage with resource snapshots
run_stage() {
    local label=$1
    local concurrency=$2
    local requests=$3
    local method=$4
    local path=$5
    local body="${6:-}"

    echo "=== Stage: $label (c=$concurrency, n=$requests) ==="
    print_resources "before"
    run_hey "$concurrency" "$requests" "$method" "$path" "$body"
    print_resources "after"
    echo ""
}

echo "=============================="
echo "  Progressive Load Test"
echo "  Instances: $INSTANCES"
echo "  Database:  $WITH_DB"
echo "=============================="
echo ""

# ---- Phase 1: Echo (no DB) ----
echo "################################"
echo "  Phase 1: Echo (no database)"
echo "################################"
echo ""

run_stage "Warm up"  10  1000  GET "/api/echo?msg=loadtest"
run_stage "Moderate" 50  5000  GET "/api/echo?msg=loadtest"
run_stage "Heavy"    200 20000 GET "/api/echo?msg=loadtest"
run_stage "Stress"   500 50000 GET "/api/echo?msg=loadtest"

if [[ "$WITH_DB" == "true" ]]; then
    # ---- Phase 2: DB writes ----
    echo "################################"
    echo "  Phase 2: Database writes"
    echo "################################"
    echo ""

    run_stage "DB Write Warm up"  10  1000  POST "/api/items" '{"name":"loadtest"}'
    run_stage "DB Write Moderate" 50  5000  POST "/api/items" '{"name":"loadtest"}'
    run_stage "DB Write Heavy"    200 20000 POST "/api/items" '{"name":"loadtest"}'
    run_stage "DB Write Stress"   500 50000 POST "/api/items" '{"name":"loadtest"}'

    # ---- Phase 3: DB reads (table now has rows) ----
    echo "################################"
    echo "  Phase 3: Database reads"
    echo "################################"
    echo ""

    run_stage "DB Read Warm up"  10  1000  GET "/api/items"
    run_stage "DB Read Moderate" 50  5000  GET "/api/items"
    run_stage "DB Read Heavy"    200 20000 GET "/api/items"
    run_stage "DB Read Stress"   500 50000 GET "/api/items"

    # ---- Phase 4: Mixed read/write (alternating stages) ----
    echo "################################"
    echo "  Phase 4: Mixed read/write"
    echo "################################"
    echo ""

    echo "=== Mixed: concurrent writes + reads (c=200, n=10000 each) ==="
    print_resources "before"

    # Run writes and reads in parallel
    if [[ $INSTANCES -gt 1 ]]; then
        for i in $(seq 0 $((INSTANCES - 1))); do
            port=$((BASE_PORT + i))
            echo "  Instance :$port writes"
            hey -n 2500 -c 100 -q 0 -m POST \
                -H "Content-Type: application/json" -d '{"name":"mixed"}' \
                "http://localhost:$port/api/items" 2>&1 | grep -E "Requests/sec|Average|Fastest|Slowest|Status code" &
        done
        for i in $(seq 0 $((INSTANCES - 1))); do
            port=$((BASE_PORT + i))
            echo "  Instance :$port reads"
            hey -n 2500 -c 100 -q 0 -m GET \
                "http://localhost:$port/api/items" 2>&1 | grep -E "Requests/sec|Average|Fastest|Slowest|Status code" &
        done
    else
        echo "  Writes"
        hey -n 10000 -c 200 -q 0 -m POST \
            -H "Content-Type: application/json" -d '{"name":"mixed"}' \
            "http://localhost:$BASE_PORT/api/items" 2>&1 | grep -E "Requests/sec|Average|Fastest|Slowest|Status code" &
        echo "  Reads"
        hey -n 10000 -c 200 -q 0 -m GET \
            "http://localhost:$BASE_PORT/api/items" 2>&1 | grep -E "Requests/sec|Average|Fastest|Slowest|Status code" &
    fi
    wait

    print_resources "after"
    echo ""
fi

# Collect stats from all instances
echo "=== Final stats ==="
for i in $(seq 0 $((INSTANCES - 1))); do
    port=$((BASE_PORT + i))
    echo -n "  :$port — "
    curl -sf "http://localhost:$port/api/stats" 2>/dev/null || echo "unreachable"
done
echo ""

echo "=== System state ==="
echo "  TIME_WAIT sockets: $(netstat -an 2>/dev/null | grep -c TIME_WAIT || echo 'n/a')"
echo "  ESTABLISHED sockets: $(netstat -an 2>/dev/null | grep -c ESTABLISHED || echo 'n/a')"
echo ""
