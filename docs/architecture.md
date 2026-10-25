# Architecture

Varnqor is a local-first resource intelligence layer for AI coding agents.

## Current pipeline

```text
Claude Code
    |
    v
JSONL transcripts (read-only)
    |
    v
Incremental collector
    |
    +--> usage calls
    +--> privacy-minimised tool metadata
    |
    v
SQLite local index
    |
    +--> Burn Rate
    +--> Attribution
    +--> Waste Intelligence
    |       |
    |       v
    +--> Forecast / Resource Pressure
    |       |
    +-------+-------> Advisor
                      |
                      v
                 Desktop UI
```

## Product loop

```text
OBSERVE -> NORMALIZE -> ATTRIBUTE -> ANALYZE -> PREDICT -> ADVISE -> OPTIMIZE
```

`OPTIMIZE` is intentionally not automatic yet. The current Advisor is read-only and only surfaces recommendations.

## Boundaries

- `internal/providers/claude` owns Claude-specific discovery and parsing.
- provider-specific parsing should not leak into analytics contracts.
- `internal/storage` owns local persistence and indexing semantics.
- `internal/analytics` owns Burn Rate, Waste Intelligence, Forecasting and Advisor logic.
- raw prompts, responses, tool-result text and source-code contents are not required for the current analytics.
- transcript access is read-only.
- estimates must be labelled as estimates.
- local baseline must never be presented as provider quota.

## Forecasting boundary

The current forecast is a persistence model over observed local activity. Resource Pressure compares current pace with the active-hour baseline. Runway is unavailable until a trusted provider limit source exists.

## Advisor boundary

Advisor consumes existing signals and returns prioritized guidance. It does not execute recommendations or mutate agents, prompts, models, provider settings or files.

See [forecasting.md](forecasting.md), [advisor.md](advisor.md), [waste-analysis.md](waste-analysis.md), [attribution.md](attribution.md) and [burn-rate.md](burn-rate.md).
