# Contributing to Varnqor

Varnqor is pre-alpha. The architecture is intentionally narrow while the first data model is being validated.

## Development principles

1. Prefer explainable measurements over opaque scores.
2. Never present an estimate as an official provider quota.
3. Keep provider-specific logic behind provider adapters.
4. Keep the default experience local-first and privacy-conscious.
5. Add tests for collector behavior before expanding heuristics.

## Local development

Requirements:

- Go 1.26+
- Node.js 24+
- Wails v2.17+

Install Wails:

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@v2.17.0
```

Then:

```bash
cd frontend
npm install
cd ..
wails dev
```

## Pull requests

Keep pull requests focused. Explain:

- the problem being solved;
- how the change was tested;
- whether it changes local data access;
- whether any displayed value is measured, official, configured, observed, or estimated.
