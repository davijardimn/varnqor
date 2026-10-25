# Architecture

Varnqor is a local-first resource intelligence layer for AI coding agents.

The first milestone intentionally keeps the system small:

```text
Claude Code
    |
    v
Claude detector
    |
    v
Normalized local signal
    |
    +--> Wails bridge --> Desktop UI
    |
    +--> future analytics/storage pipeline
```

## Boundaries

- `internal/providers/claude` owns Claude-specific discovery and parsing.
- the desktop shell must consume provider-neutral data as the project grows.
- raw prompts and source code are not required for basic usage metrics.
- provider estimates must be labeled as estimates; Varnqor must not present inferred quota as official quota.

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

## Milestone 0: First Signal

Milestone 0 proves that Varnqor can:

1. detect a Claude Code installation or local data directory;
2. discover project stores under `~/.claude/projects`;
3. count top-level session transcripts without reading prompt content;
4. expose that signal to the desktop UI.

No quota claims or token estimates are made in this milestone.
