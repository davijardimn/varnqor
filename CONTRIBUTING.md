# Contributing to Varnqor

Varnqor is pre-alpha. Contributions should strengthen **explainable, local-first resource intelligence** rather than inflate the number of metrics.

## Development principles

1. Prefer explainable measurements over opaque scores.
2. Never present an estimate as an official provider quota.
3. Keep provider-specific parsing behind provider adapters.
4. Keep the default experience local-first and privacy-conscious.
5. Add tests for collector and analytics behavior before expanding heuristics.
6. Observation and advice must remain distinct from actions that mutate an agent or provider.

## Local development

Requirements: Go 1.26+, Node.js 24+, Wails v2.17+.

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@v2.17.0
cd frontend && npm install && cd ..
wails dev
```

Before opening a pull request:

```bash
go test ./internal/...
go vet ./internal/...
cd frontend && npm run build
```

See [docs/testing.md](docs/testing.md) for the desktop CI layer.

## Data-access review

If a change reads or persists a new field from agent logs, explain why that field is required. Do not persist prompt text, response text, source-code content, raw tool-result text, raw Bash commands or search queries merely because they are available.

## Provider limits

New limit adapters must state provenance. Supported source labels are conceptually `official`, `observed`, `configured`, and—only if clearly justified—`estimated`. A local baseline is not a provider quota.

## Pull requests

Keep changes focused. Describe:

- the problem being solved;
- how the change was tested;
- whether local data access or persistence changes;
- whether displayed values are measured, official, configured, observed, or estimated;
- whether the feature only observes/advises or can change agent behavior.
