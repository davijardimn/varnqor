<div align="center">

# Varnqor

### Resource intelligence for AI coding agents.

**Observe usage. Understand burn. Detect waste. Act with evidence.**

[![Version](https://img.shields.io/badge/version-1.0.0-111318)](#varnqor-10)
[![License: MIT](https://img.shields.io/badge/license-MIT-111318)](LICENSE)
[![Local first](https://img.shields.io/badge/local--first-yes-6366F1)](#privacy)

</div>

---

Varnqor is a **local-first desktop application and CLI** for understanding resource usage across AI coding agents. It reads supported local session telemetry, normalizes usage into a common model, attributes burn to projects/sessions/models, forecasts resource pressure, and surfaces evidence-based recommendations.

```text
OBSERVE -> NORMALIZE -> ATTRIBUTE -> ANALYZE -> PREDICT -> ADVISE
```

Varnqor does **not** bypass provider limits, modify agent behavior automatically, upload prompts, or pretend an estimated value is an official quota.

## Varnqor 1.0

Version 1.0 turns Varnqor from a Claude-only prototype into a provider-aware resource intelligence product.

### Supported local providers

| Provider | Usage telemetry | Tool metadata | Waste analysis | Notes |
| --- | --- | --- | --- | --- |
| Claude Code | Yes | Yes | Yes | Full 1.0 analysis path |
| Codex | Yes | Not yet | Not yet | JSONL sessions; compressed `.jsonl.zst` files are reported but not indexed |
| Gemini CLI | Yes | Not yet | Not yet | JSON and JSONL chat sessions |

Provider support is capability-based. The desktop **Agents** view shows exactly what Varnqor can observe for each provider instead of using a generic “connected” label for unsupported features.

## What Varnqor measures

### Usage
- input and output tokens
- cache reads and supported cache writes
- token-bearing calls
- hourly activity
- model usage
- project and session attribution

### Burn
Burn Rate compares the latest 60-minute pace against the median of completed active hours. Varnqor remains in **Learning** until it has enough active baseline history rather than showing a misleading multiplier.

### Resource pressure
The forecast engine uses a transparent persistence model. It projects the current pace over short horizons and reports an observed pace envelope derived from recent activity and the active baseline.

This is a usage projection, **not a statistical confidence interval and not provider quota data**.

### Trusted runway
Varnqor only computes runway when it has a trusted limit source. In 1.0, the universally supported source is a **user-configured local budget**.

A configured budget is always labeled as configured. It is never presented as an official Claude, Codex, or Gemini quota.

### Waste intelligence
Claude Code tool metadata enables explainable heuristics for:
- repeated file reads
- repeated searches
- oversized tool responses
- repeated failed operations

Findings show evidence and confidence. Varnqor does not collapse overlapping heuristics into a fake “waste percentage”.

### Advisor
Advisor combines burn, attribution, forecast, and Waste signals into prioritized recommendations. It is **advisory only** in 1.0: it does not rewrite prompts, switch models, kill agents, compact context, or execute tools.

## Desktop experience

The 1.0 desktop app includes:
- Overview
- Agents / Provider Hub
- Burn
- Advisor
- Activity
- Projects
- Models
- Settings + Doctor

Empty states are data-aware. For example, Varnqor distinguishes a locally observed session from a token-bearing usage event, so a detected session no longer appears as a phantom zero-token model call.

## CLI

```bash
varnqor
varnqor status
varnqor providers
varnqor analyze
varnqor forecast
varnqor doctor
varnqor config show
varnqor config budget 10m
varnqor version
```

Machine-readable output is available with `--json` on status, providers, analyze, forecast, and doctor commands.

## Privacy

Varnqor is designed around a local privacy boundary.

By default:
- no Varnqor cloud account
- no prompt upload
- no source-code body upload
- no tool-result content persistence
- no API keys required for local usage analysis
- no product telemetry required for analysis

The usage index stores resource metadata such as token counts, timestamps, models, project/session labels, and source offsets. Claude tool analysis stores privacy-minimized metadata such as tool name, safe file/range fields, SHA-256 input fingerprints, result byte size, and error state — not the tool result body itself.

See [docs/privacy.md](docs/privacy.md) for the detailed boundary.

## Data correctness

Varnqor 1.0 explicitly separates:

- **observed activity** — a local session/file exists
- **token-bearing activity** — a supported usage event contains actual resource counts
- **provider health** — local data can be read and indexed

Pre-1.0 zero-token usage rows are cleaned automatically on upgrade. Internal placeholder model labels are normalized or hidden from product surfaces.

## Local storage

Historical telemetry is stored in a local SQLite database. Indexing is incremental: append-only session files resume from stored source offsets instead of rescanning the entire history on every refresh.

Legacy Gemini JSON sessions are safely re-indexed when rewritten because those files are not append-only.

## Build

Varnqor uses:
- Go
- Wails
- React + TypeScript
- SQLite

```bash
go test ./internal/...
go vet ./internal/...
cd frontend && npm install && npm run build
wails build
```

GitHub Actions additionally performs a real Windows Wails build, CLI smoke tests, application-icon generation, and SHA-256 checksum generation.

## Limits of 1.0

- Claude has the deepest integration because its tool metadata adapter is already tested.
- Codex compressed session archives are visible as unsupported files but are not silently decoded in 1.0.
- Gemini/Codex contribute to usage, burn, forecast, model and project intelligence, but do not contribute tool-level Waste findings yet.
- Provider limits are not assumed to be interchangeable. Runway remains unavailable unless a trusted limit exists.

These boundaries are intentional. Varnqor prefers **“unavailable” over fabricated precision**.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md). Provider adapters should preserve the local privacy boundary and must include parser fixtures/tests before being advertised as supported.

## License

MIT. See [LICENSE](LICENSE).

---

<div align="center">

**Control the burn.**

</div>
