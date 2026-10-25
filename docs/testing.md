# Verification Strategy

Varnqor uses several layers of automated verification.

## Core

```bash
go test ./internal/...
go vet ./internal/...
```

This covers parsers, incremental indexing, SQLite behavior, Burn Rate, Waste Intelligence, Forecasting, Runway, settings and Doctor summarization.

## Frontend

```bash
cd frontend
npm install
npm run build
```

TypeScript is strict and the production Vite bundle must compile on every pull request and push to `main`.

## Windows desktop smoke build

CI builds the real Wails application on `windows-latest`, then executes CLI commands from the generated `.exe`.

On pushes to `main`, the workflow also uploads that executable as a short-lived Actions artifact. This artifact is a **CI development build**, not a signed public release.

## Release gate

Automated compilation is necessary but not sufficient for the first public pre-alpha. Before tagging `v0.1.0-alpha.1`, perform a clean-machine Windows launch test and approve the final Varnqor application icon.
