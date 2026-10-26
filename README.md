<div align="center">

# Varnqor

### Resource intelligence for AI coding agents.

**Observe. Investigate. Catch pressure before it becomes waste.**

[![Version](https://img.shields.io/badge/version-2.0.0-111318)](#varnqor-20)
[![License: MIT](https://img.shields.io/badge/license-MIT-111318)](LICENSE)
[![Local first](https://img.shields.io/badge/local--first-yes-6366F1)](#privacy)

</div>

---

Varnqor is a **local-first desktop application and CLI** for understanding resource usage across AI coding agents. It reads supported local session telemetry, normalizes token activity into a common model, attributes usage to providers/projects/sessions/models, and turns that history into explainable investigation and advisory signals.

```text
OBSERVE -> NORMALIZE -> ATTRIBUTE -> EXPLORE -> ANALYZE -> ADVISE
```

Varnqor does **not** bypass provider limits, upload prompts, inspect source-code bodies for analytics, or present an estimate as an official quota.

## Varnqor 2.0

Version 2 changes the center of gravity from a current-state dashboard to an **investigation workspace**.

### Explorer

Explorer queries the local SQLite history across arbitrary supported windows from 15 minutes to 7 days. The desktop ships fast presets for **1h, 6h, 24h, and 7d**.

You can filter by:
- provider
- project
- model

Every selected window is compared with the immediately preceding window of the same duration. Varnqor reports token/call deltas, cache-read movement, timeline buckets, provider/project/model attribution, high-usage sessions, and recent token-bearing calls.

### Pulse

Pulse is an explainable local guardrail engine. It has no opaque score and takes no automatic action.

Default signals cover:
- token spikes versus the previous comparable window
- project concentration
- session concentration
- low cache reuse when cache telemetry is actually present
- fragmented session activity

Thresholds and minimum evaluation volume are stored in local settings and are user-configurable.

### Faster desktop workflow

The Varnqor 2 desktop adds:
- global 1h / 6h / 24h / 7d range switching
- Explorer filters
- JSON and CSV snapshots copied directly from the selected view
- `Ctrl+K` / `Cmd+K` quick navigation and range switching
- Overview driven by the same selected Explorer window
- dedicated Pulse surface

The app intentionally keeps product branding out of the sidebar workspace; navigation starts directly with the tool.

## Supported local providers

| Provider | Usage telemetry | Tool metadata | Waste analysis | Notes |
| --- | --- | --- | --- | --- |
| Claude Code | Yes | Yes | Yes | Deepest integration |
| Codex | Yes | Not yet | Not yet | JSONL sessions; compressed `.jsonl.zst` files are reported but not indexed |
| Gemini CLI | Yes | Not yet | Not yet | Current JSONL and legacy JSON chat sessions |

Provider support is capability-based. The **Agents** view reports what Varnqor can actually observe for each provider rather than treating every detected executable as equivalent.

## Intelligence layers

### Burn
Burn Rate compares the latest rolling 60-minute pace with the median of completed active hours. Until enough active baseline history exists, Varnqor explicitly stays in **Learning**.

### Resource pressure and forecast
The forecast engine uses a transparent persistence model over observed activity. Its pace envelope is context for current behavior; it is **not a statistical confidence interval and not provider quota data**.

### Trusted runway
Runway is enabled only when a trusted limit exists. The universally supported source is a **user-configured local budget**, always labeled as configured rather than official provider quota.

### Waste intelligence
Claude Code tool metadata enables explainable heuristics for repeated reads/searches, oversized responses, and repeated failed operations. Raw tool-result bodies are not persisted by the analytics index.

### Advisor
Advisor combines burn, attribution, forecast, and Waste evidence into prioritized read-only recommendations. It does not rewrite prompts, switch models, kill agents, or execute tools.

## Desktop surfaces

```text
Observe
  Overview
  Explorer
  Pulse
  Activity

Analyze
  Burn
  Advisor
  Projects
  Models

System
  Agents
  Settings
```

## CLI

```bash
varnqor
varnqor status
varnqor providers
varnqor explore --range 6h
varnqor explore --range 24h --provider codex --json
varnqor pulse --range 1h
varnqor export --range 24h --format csv
varnqor analyze
varnqor forecast
varnqor doctor
varnqor config show
varnqor config budget 10m
varnqor version
```

Explorer accepts `--provider`, `--project`, `--model`, `--range`, and `--limit`. JSON remains available for automation, while `export` supports JSON and CSV output.

## Privacy

Varnqor is designed around a local privacy boundary.

By default:
- no Varnqor cloud account
- no prompt upload
- no source-code body upload
- no raw tool-result body persistence
- no API key required for local usage analysis
- no product telemetry required for analysis

The usage index stores resource metadata such as token counts, timestamps, models, provider/project/session labels, and source offsets. Claude tool analysis stores privacy-minimized evidence such as tool name, safe range fields, fingerprints, result byte size, and error state.

See [docs/privacy.md](docs/privacy.md) for the detailed boundary.

## Data correctness

Varnqor separates three concepts that are easy to conflate:
- **observed activity** — a local session/file exists
- **token-bearing activity** — a supported usage event contains resource counts
- **provider health** — supported local data can be read and indexed

Provider identity is preserved in attribution so identical project/session/model labels from different agents do not collide.

## Local storage

Historical telemetry is stored in local SQLite. Append-only session files resume from stored offsets. Rewritten formats such as legacy Gemini JSON are safely re-indexed when changed.

Explorer reads this local index; it does not call provider APIs to reconstruct history.

## Build

Varnqor uses Go, Wails, React + TypeScript, and SQLite.

```bash
go test ./internal/...
go vet ./internal/...
cd frontend && npm install && npm run build
wails build
```

GitHub Actions performs core tests/vet, frontend production compilation, a real Windows Wails build, CLI smoke tests (including Explorer/Pulse/export), deterministic application-icon generation, and SHA-256 checksum generation.

## Intentional boundaries

- Claude remains the only provider with tool-level Waste analysis until other adapters meet the same evidence/privacy bar.
- Compressed Codex session archives are reported but not silently decoded.
- Provider limits are not assumed to be interchangeable.
- Pulse is advisory and local; it does not automate agent behavior.
- Explorer currently retains a maximum selected window of 7 days even if the local database contains older history.

Varnqor prefers **unavailable over fabricated precision**.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md). Provider adapters should preserve the local privacy boundary and include parser fixtures/tests before being advertised as supported.

## License

MIT. See [LICENSE](LICENSE).

---

<div align="center">

**Control the burn.**

</div>
