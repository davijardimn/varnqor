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

```text
OBSERVE -> NORMALIZE -> ATTRIBUTE -> ANALYZE -> PREDICT -> ADVISE -> OPTIMIZE
```

> **Control the burn.**

## What works now

The current Claude Code pre-alpha stack includes:

- deduplicated input/output/cache accounting;
- incremental SQLite history and hourly activity;
- Burn Rate against a robust active-hour baseline;
- project/session/model/subagent attribution;
- explainable Waste Intelligence rules;
- persistence forecasting and Resource Pressure;
- read-only prioritized Advisor recommendations;
- **configured trusted budgets + quota-aware Runway**;
- **live incremental refresh while the desktop is open**;
- local Settings and first-run onboarding;
- visual diagnostics plus `varnqor doctor` / `varnqor status`;
- automated Windows desktop smoke builds and prerelease packaging.

## Trusted limits, not fake quotas

Varnqor never turns your historical baseline into a provider quota.

The first trusted-limit source is `configured`: you may set your own daily token-activity budget and reset hour. Varnqor then calculates remaining budget, current-pace runway and projected usage at reset while clearly identifying the source as user-configured.

```text
CONFIGURED DAILY BUDGET
Used        6.2M / 10M
Remaining   3.8M
Runway      2.7h at current pace
Reset       00:00 local
Source      CONFIGURED
```

This is a budget, **not an official Claude plan limit**. See [docs/limits-and-runway.md](docs/limits-and-runway.md).

## Live monitor

The desktop refreshes the incremental local index every 5 seconds by default. The interval can be changed from 2–60 seconds.

```text
● LIVE · 5S
```

Live mode runs only while Varnqor Desktop is open; no hidden service is installed. See [docs/live-monitor.md](docs/live-monitor.md).

## Diagnostics

The same diagnostic layer is available in the UI and CLI:

```bash
varnqor doctor
varnqor doctor --json
varnqor status
varnqor status --json
varnqor version
```

Doctor checks Claude discovery, local data, SQLite, settings and config-directory write access.

## Privacy boundary

```text
No cloud account
No telemetry by default
No prompt uploads
No source-code uploads
Read-only source analysis
```

Raw prompts, assistant text, file contents, raw Bash commands, Grep patterns and tool-result text are not persisted by the current analytics pipeline. Tool inputs used for repeat detection are fingerprinted locally and discarded.

## Architecture

- **Wails v2.17** — native desktop shell
- **Go** — collectors, indexing, analytics, settings and diagnostics
- **React + TypeScript + Vite** — interface
- **SQLite** — local incremental history

See [docs/architecture.md](docs/architecture.md), [docs/forecasting.md](docs/forecasting.md), [docs/advisor.md](docs/advisor.md) and [docs/waste-analysis.md](docs/waste-analysis.md).

## v0.1 pre-alpha checklist

- [x] Claude Code detection and session discovery
- [x] Deduplicated token accounting
- [x] SQLite incremental history
- [x] Burn Rate and attribution
- [x] Waste Intelligence
- [x] Forecasting and Resource Pressure
- [x] Read-only Advisor
- [x] Configured trusted budget and Runway
- [x] Live desktop refresh
- [x] Settings, onboarding and Doctor
- [x] Windows CI smoke build
- [x] Windows prerelease workflow
- [ ] Approved Varnqor application icon
- [ ] Clean-machine installation test
- [ ] Tag `v0.1.0-alpha.1`

## Development

Requirements: Go 1.26+, Node.js 24+, Wails v2.17+.

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@v2.17.0
cd frontend && npm install && cd ..
wails dev
```

Core verification:

```bash
go test ./internal/...
go vet ./internal/...
cd frontend && npm run build
```

## Releasing

The repository contains a Windows packaging workflow. A manual workflow run produces an artifact; a `v*` tag produces a GitHub prerelease. See [docs/releasing.md](docs/releasing.md).

## Roadmap

```text
v0.1  Claude Code resource intelligence
v0.2  Provider abstraction + Codex
v0.3  Gemini
v0.4  Deeper Waste Intelligence
v0.5  Official/observed quota adapters
v0.6  Advisor refinement
v0.7  Explicitly approved resource policies
v0.8  Adapter SDK
v0.9  Reports & share cards
v1.0  AI resource control plane
```

## Status

**Pre-alpha.** APIs, models and UI are expected to change. The first public tag has not been cut yet.

## License

[MIT](LICENSE)
