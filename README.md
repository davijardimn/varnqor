<div align="center">

# VARNQOR

### Resource intelligence for AI coding agents.

**Observe usage. Predict pressure. Detect waste. Advise with evidence.**

[![Status: Pre-alpha](https://img.shields.io/badge/status-pre--alpha-111111)](#status)
[![License: MIT](https://img.shields.io/badge/license-MIT-111111)](LICENSE)
[![Core: Go](https://img.shields.io/badge/core-Go-111111)](go.mod)

</div>

---

Varnqor is a **local-first desktop application** for understanding how AI coding agents consume resources.

It is being built around one loop:

```text
OBSERVE -> NORMALIZE -> ATTRIBUTE -> ANALYZE -> PREDICT -> ADVISE -> OPTIMIZE
```

> **Control the burn.**

## Current block — Forecasting + Advisor

Varnqor can now move beyond retrospective usage and answer two new questions:

> If this pace continues, what does the next two hours look like?

> Given the evidence already collected, what should I inspect first?

### M6 — Forecasting & Resource Pressure

The forecast uses the latest rolling 60-minute activity rate and projects it forward with an intentionally simple **persistence model**. It also exposes an observed pace envelope from current pace, previous pace and the active-hour baseline.

The envelope is **not** a statistical confidence interval, and the forecast is **not** a provider quota forecast.

Runway stays disabled until a trusted limit source exists. Varnqor never turns a local baseline into fake quota remaining.

See [docs/forecasting.md](docs/forecasting.md).

### M7 — Advisor Engine

Advisor consumes Burn Rate, Resource Pressure, Attribution and Waste Intelligence to produce prioritized, explainable recommendations.

Current guidance can cover:

```text
resource pressure
repeated file reads
repeated search
oversized tool responses
repeated failed operations
subagent fan-out under pressure
dominant-session investigation
```

Advisor is **read-only**. It does not change prompts, models, files, agents or provider settings.

See [docs/advisor.md](docs/advisor.md).

## Waste Intelligence

Varnqor currently runs four independent evidence rules:

```text
Repeated File Reads
Repeated Grep / Glob searches
Oversized Tool Responses
Repeated Failed Operations
```

Tool inputs are fingerprinted locally with SHA-256 and discarded; raw commands, search queries, file contents and tool-result text are not stored.

Varnqor deliberately does not publish a single Waste Score yet because heuristic signals can overlap.

See [docs/waste-analysis.md](docs/waste-analysis.md).

## Planned v0.1

- [x] Project foundation
- [x] Claude Code detection and session discovery
- [x] Deduplicated token accounting
- [x] SQLite incremental history
- [x] Hourly usage series
- [x] Burn Rate
- [x] Deep session/project/model/subagent attribution
- [x] Waste Intelligence v0.1
- [x] Usage forecasting and Resource Pressure
- [x] Read-only Advisor Engine
- [ ] Trusted limit adapters + quota-aware runway
- [ ] Live refresh / tray monitoring
- [ ] First packaged pre-alpha release

## Local-first

```text
No cloud account
No telemetry by default
No prompt uploads
No source-code uploads
Read-only source analysis
```

## Honest resource intelligence

Varnqor distinguishes measurements and estimates. Provider limits will be labelled by source such as `official`, `observed`, `estimated`, or `configured`. An estimate must never be presented as an official quota.

## Desktop architecture

- **Wails v2** — native desktop shell
- **Go** — collectors, indexing and analytics
- **React + TypeScript + Vite** — interface
- **SQLite** — local incremental history

## Development

Requirements: Go 1.26+, Node.js 24+, Wails v2.17+.

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@v2.17.0
cd frontend && npm install && cd ..
wails dev
```

Run core tests:

```bash
go test ./internal/...
```

## Roadmap

```text
v0.1  Claude Code resource intelligence
v0.2  Provider abstraction + Codex
v0.3  Gemini
v0.4  Deeper Waste Intelligence
v0.5  Trusted quota/runway adapters
v0.6  Advisor refinement
v0.7  Explicitly approved resource policies
v0.8  Adapter SDK
v0.9  Reports & share cards
v1.0  AI resource control plane
```

## Status

**Pre-alpha.** APIs, models and UI are expected to change.

## License

[MIT](LICENSE)
