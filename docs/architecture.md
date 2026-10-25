# Architecture

Varnqor is a local-first resource intelligence layer for AI coding agents.

## Current pipeline

```text
Claude Code
    |
    v
Local JSONL transcripts
    |
    +--> detector
    |
    +--> usage scanner
            |
            v
       deduplication
            |
            v
       token events
            |
            v
       daily summary
            |
            v
       Wails bridge
            |
            v
       Desktop UI
```

## Boundaries

- `internal/providers/claude` owns Claude-specific discovery and parsing.
- provider-specific parsing must not leak into analytics/UI contracts as the project grows.
- raw prompts and source code are not required for resource metrics.
- provider estimates must be labelled as estimates; Varnqor must not present inferred quota as official quota.
- transcript access is read-only.

## Milestone 0: First Signal — complete

Milestone 0 proved that Varnqor can detect Claude Code, discover project/session stores and surface a local signal in the desktop UI.

## Milestone 1: Usage Engine — current

Milestone 1 adds:

1. token-bearing assistant-record parsing;
2. streaming/cross-file deduplication;
3. input/output/cache token accounting;
4. cache TTL breakdown support;
5. daily project/session/model aggregation;
6. a real usage overview in the desktop app.

See [usage-engine.md](usage-engine.md) for the accounting rules.

## Planned pipeline

```text
Provider collector
      |
      v
Usage normalizer
      |
      v
UsageEvent
      |
  +---+---------+-----------+
  |             |           |
SQLite       Analytics   Live events
                |
         +------+------+
         |             |
      Burn rate     Waste rules
         |             |
         +------+------+
                |
             Forecast
                |
            Desktop UI
```

The current full-file scanner is intentionally simple. The next storage milestone will add incremental SQLite indexing without changing the accounting semantics.
