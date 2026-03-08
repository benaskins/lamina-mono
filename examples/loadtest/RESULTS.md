# Load Test Results

Machine: Mac Studio, Apple M3 Ultra, 32 cores, 512GB RAM
Postgres: pgvector/pgvector:pg16 (containerised via OrbStack)
Date: 2026-03-08

## Stateless (echo endpoint, no database)

Single instance plateaus at ~54k req/s. Adding instances scales linearly
for stateless requests — 4 instances sustained ~200k req/s aggregate with
no degradation.

| Stage | Concurrency | Req/s (per instance) | Avg latency | Slowest |
|-------|-------------|---------------------|-------------|---------|
| Warm up | 10 | 25,000 | 0.4ms | 2ms |
| Moderate | 50 | 32,000 | 1.5ms | 7ms |
| Heavy | 200 | 50,000 | 3.8ms | 25ms |
| Stress | 500 | 54,000 | 8.7ms | 34ms |
| Extreme | 1000 | 54,000 | 18ms | 60ms |

Memory: 13MB at rest → 65MB peak per instance. CPU not a factor.

## Database writes (INSERT with RETURNING)

Each instance configured with 25-connection pool against containerised Postgres.

| Stage | Concurrency | Req/s (per instance) | Avg latency | Slowest |
|-------|-------------|---------------------|-------------|---------|
| Warm up | 10 | 5,400 | 1.8ms | 13ms |
| Moderate | 50 | 6,200 | 7.5ms | 25ms |
| Heavy | 200 | 15,300 | 11ms | 92ms |
| Stress | 500 | 17,500 | 25ms | 300ms |

## Database reads (SELECT LIMIT 20 ORDER BY id DESC)

| Stage | Concurrency | Req/s (per instance) | Avg latency | Slowest |
|-------|-------------|---------------------|-------------|---------|
| Warm up | 10 | 16,400 | 0.6ms | 3ms |
| Moderate | 50 | 17,000 | 2.7ms | 18ms |
| Heavy | 200 | 21,700 | 8.3ms | 64ms |
| Stress | 500 | 23,000 | 19ms | 200ms |

## Mixed concurrent (reads + writes in parallel)

~2,300 req/s per stream. Connection pool contention is the bottleneck —
800 goroutines competing for 100 total DB connections (25 per instance × 4).

## Async commit (synchronous_commit = off)

Writes improved from ~17,500 to ~24,100 req/s (+38%). Confirms that ~38%
of write latency is WAL fsync overhead.

## macOS tuning impact

Without tuning, 4-instance extreme test degraded from 28k → 8k req/s on
later instances due to ephemeral port exhaustion (16k range, 30s TIME_WAIT).

After tuning (wider port range, 5s TIME_WAIT, raised fd limits), all
instances held steady at ~50k req/s.

See `scripts/tune-macos.sh` for the specific settings.

## Key findings

1. **Go HTTP layer is never the bottleneck.** 54k req/s per instance, ~60MB RAM.
2. **Database is the ceiling.** Writes plateau at ~17k req/s (sync commit).
3. **macOS defaults are too conservative for load testing.** File descriptors (256),
   ephemeral ports (16k), and TIME_WAIT (30s) all need tuning.
4. **Connection pool size doesn't help** when Postgres itself is the limit.
   The pool never fully saturated — raising from 25 to 100 made no difference.
5. **Container overhead is measurable** (~2-4ms per DB round-trip through
   OrbStack virtualisation). Native Postgres would improve DB throughput by
   an estimated 30-50%.
