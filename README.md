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

Most usage tools answer *how much was used?* Varnqor is being built to answer the next questions:

- Where did the usage go?
- Why did this session consume so much?
- How quickly am I burning resources?
- Which sessions and projects are driving consumption?
- Which usage patterns may have been avoidable?
- When is the current burn rate likely to become a problem?

> **Control the burn.**

## First Signal

The current milestone is intentionally small: detect Claude Code and prove that Varnqor can turn local agent activity into a reliable signal.

The desktop shell currently targets:

```text
Claude Code
    |
    v
Local detection
    |
    +-- project count
    +-- session count
    +-- last local activity
    |
    v
Varnqor Desktop
```

No prompt text is needed for this milestone.

## Product direction

```text
OBSERVE -> NORMALIZE -> ATTRIBUTE -> ANALYZE -> PREDICT -> OPTIMIZE
```

Varnqor will grow in that order. Optimization comes after the measurements are trustworthy.

## Desktop architecture

- **Wails v2** for the native desktop shell
- **Go** for collectors, analytics and the core engine
- **React + TypeScript + Vite** for the interface
- **SQLite** for local history in a later v0.1 milestone

Wails v2 is used deliberately while Wails v3 remains beta.

## Planned v0.1

- [x] Project foundation
- [x] Claude Code local detector
- [x] Claude project/session discovery
- [x] Desktop First Signal screen
- [ ] Claude usage event parser
- [ ] Token accounting
- [ ] SQLite history
- [ ] Burn rate
- [ ] Session/project attribution
- [ ] Waste Analyzer
- [ ] Usage forecasting

## Local-first

The default product direction is:

```text
No cloud account
No telemetry by default
No prompt uploads
No source-code uploads
Read-only local analysis
```

See [docs/privacy.md](docs/privacy.md).

## Honest quota intelligence

Providers expose limits differently. Varnqor will distinguish values that are:

- `official`
- `observed`
- `estimated`
- `configured`

An estimate must never be presented as an official quota.

## Development

Requirements:

- Go 1.26+
- Node.js 24+
- Wails v2.17+

Install Wails:

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@v2.17.0
```

Run the desktop app:

```bash
cd frontend
npm install
cd ..
wails dev
```

Run the core tests:

```bash
go test ./internal/...
```

## Repository map

```text
.
├── internal/
│   └── providers/
│       └── claude/
├── frontend/
│   └── src/
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
