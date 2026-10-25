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

## Current milestone — Waste Analyzer v0.1

Varnqor can now move from **where usage came from** to a first explainable signal for **why some of that activity may have been avoidable**.

The first Waste Intelligence rule detects exact file reads repeated within a short window:

```text
same transcript
same file
same offset / limit
within 15 minutes
no observed write/edit in between
```

For matching reads, Varnqor stores only privacy-minimised tool metadata and the byte length of returned tool results. It never stores the file content or tool-result text.

The desktop can show:

```text
repeated-read findings
number of repeated reads
rough repeated-context token estimate
per-finding confidence
file / project / session evidence
```

The token estimate is explicitly heuristic (`returned bytes / 4`) and is never presented as provider-reported usage. A repeated read can be legitimate, so findings are described as **potential repeated context**, not factual waste.

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
- [ ] Additional waste rules
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
