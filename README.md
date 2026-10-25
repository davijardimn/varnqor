<div align="center">

# VARNQOR

### Resource intelligence for AI coding agents.

**Observe usage. Predict pressure. Detect waste.**

[![Status: Pre-alpha](https://img.shields.io/badge/status-pre--alpha-111111)](#status)
[![License: MIT](https://img.shields.io/badge/license-MIT-111111)](LICENSE)
[![Core: Go](https://img.shields.io/badge/core-Go-111111)](go.mod)

</div>

---

Varnqor is a **local-first desktop application** for understanding how AI coding agents consume resources.

Most usage tools answer *how much was used?* Varnqor is being built to answer the next questions: where usage went, why a session consumed so much, how quickly resources are burning, and which patterns may have been avoidable.

> **Control the burn.**

## Current milestone — Waste Intelligence expansion

Varnqor now runs four independent, explainable rules against privacy-minimised Claude Code tool metadata:

```text
Repeated File Reads
Repeated Grep / Glob searches
Oversized Tool Responses
Repeated Failed Operations
```

The rules share the same rolling 60-minute window as Burn Rate and Attribution. Each finding includes its evidence, confidence, severity and a rough context-volume estimate when result-size metadata is available.

Varnqor deliberately does **not** publish a single Waste Score yet. Rules can overlap, and turning overlapping heuristic estimates into one number before defining deduplication and validation would be misleading.

Tool inputs are not stored. Varnqor persists only a SHA-256 fingerprint of canonical tool input so exact repeated operations can be recognized locally without retaining raw Bash commands, Grep patterns or Write/Edit payloads.

See [docs/waste-analysis.md](docs/waste-analysis.md), [docs/attribution.md](docs/attribution.md), [docs/burn-rate.md](docs/burn-rate.md), [docs/usage-engine.md](docs/usage-engine.md), and [docs/storage.md](docs/storage.md).

## Product direction

```text
OBSERVE -> NORMALIZE -> ATTRIBUTE -> ANALYZE -> PREDICT -> OPTIMIZE
```

## Desktop architecture

- **Wails v2** — native desktop shell
- **Go** — collectors, indexing and analytics
- **React + TypeScript + Vite** — interface
- **SQLite** — local incremental history

## Planned v0.1

- [x] Project foundation
- [x] Claude Code local detector
- [x] Claude project/session discovery
- [x] Desktop First Signal screen
- [x] Claude usage event parser
- [x] Deduplicated token accounting
- [x] Model/session/project daily aggregation
- [x] SQLite incremental history
- [x] Hourly usage series
- [x] Burn rate
- [x] Deep session/project attribution
- [x] Waste Analyzer v0.1 — repeated file reads
- [x] Waste Intelligence — repeated search, large responses, repeated failures
- [ ] Usage forecasting

## Local-first

```text
No cloud account
No telemetry by default
No prompt uploads
No source-code uploads
Read-only source analysis
```

The local index stores resource and tool metadata required for analysis, not prompt, response, file or tool-result content. See [docs/privacy.md](docs/privacy.md).

## Honest quota intelligence

Varnqor distinguishes values that are `official`, `observed`, `estimated`, or `configured`. An estimate must never be presented as an official provider quota.

## Development

Requirements:

- Go 1.26+
- Node.js 24+
- Wails v2.17+

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@v2.17.0
cd frontend
npm install
cd ..
wails dev
```

Run core tests with:

```bash
go test ./internal/...
```

## Repository map

```text
.
├── internal/
│   ├── analytics/
│   ├── domain/
│   ├── indexer/
│   ├── providers/claude/
│   └── storage/
├── frontend/src/
├── docs/
├── app.go
├── main.go
└── wails.json
```

## Roadmap

```text
v0.1  Claude Code observability
v0.2  Provider abstraction + Codex
v0.3  Gemini
v0.4  Waste Intelligence
v0.5  Forecasting & quota pressure
v0.6  Advisor
v0.7  Resource policies
v0.8  Adapter SDK
v0.9  Reports & share cards
v1.0  AI resource control plane
```

## Status

**Pre-alpha.** APIs, data models and UI are expected to change.

## License

[MIT](LICENSE)
