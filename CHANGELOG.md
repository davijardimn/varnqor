# Changelog

All notable Varnqor changes are tracked here. The project uses semantic versioning once releases are tagged.

## Unreleased

### Added
- Trusted configured daily budgets with transparent source labeling.
- Quota-aware remaining, reset projection, and current-pace runway.
- Live incremental desktop refresh with configurable interval.
- Local Settings panel and first-run onboarding.
- Visual Doctor diagnostics plus `varnqor doctor` and `varnqor status` CLI commands.
- `varnqor analyze` and `varnqor forecast`, including JSON output for scripts.
- CLI configuration for refresh interval and configured daily budgets.
- Human-friendly token budget parsing (`500k`, `10m`, `1.5g`).
- Windows desktop smoke builds and short-lived downloadable development artifacts.
- SHA-256 checksum generation for development and release binaries.
- Tag-driven Windows prerelease workflow.
- Dependabot for Go, frontend npm, and GitHub Actions dependencies.
- CODEOWNERS and repository verification documentation.

## 0.1.0-alpha.1 - planned

First installable pre-alpha focused on Claude Code local resource intelligence. It will include usage accounting, SQLite history, Burn Rate, Attribution, Waste Intelligence, Forecasting, Advisor, configured-budget Runway, live refresh, settings and diagnostics.
