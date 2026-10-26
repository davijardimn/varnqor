# Changelog

All notable Varnqor changes are tracked here.

## 1.0.0 — 2026-10-26

### Product
- Promoted Varnqor from pre-alpha to the first stable product model.
- Added a provider-first desktop experience with dedicated Overview, Agents, Burn, Advisor, Activity, Projects, Models, and Settings views.
- Replaced generic empty dashboards with state-aware UX for observed sessions, token-bearing activity, baseline learning, and missing trusted limits.
- Removed the entire pre-1.0 frontend generation so only one design system ships.

### Providers
- Added a generic provider status and capability model.
- Added Codex local session detection and JSONL token-usage indexing.
- Added Gemini CLI local session detection and JSON/JSONL token-usage indexing.
- Kept Claude Code as the deepest adapter with usage plus privacy-minimized tool metadata.
- Added an Agents / Provider Hub surface that reports per-provider health, sessions, indexed calls, tokens, capabilities, unsupported files, and last activity.
- Codex compressed `.jsonl.zst` sessions are explicitly reported as unsupported rather than silently ignored.

### Data correctness
- Split observed local sessions from token-bearing usage.
- Added aggregate multi-provider usage, hourly history, burn, attribution, forecast, and configured-budget runway.
- Added automatic cleanup for pre-1.0 zero-token usage rows.
- Removed internal/synthetic model labels from user-facing product surfaces.
- Made attribution provider-aware to prevent project/session/model collisions across providers.
- Added safe re-indexing for rewritten legacy Gemini JSON sessions.
- Isolated provider index failures so one adapter cannot take down the whole dashboard.

### Intelligence
- Preserved explainable Burn Rate, Resource Pressure, forecasting, configured trusted runway, Waste Intelligence, and Advisor.
- Global burn/forecast now includes supported token usage from Claude, Codex, and Gemini.
- Waste Intelligence remains explicitly Claude-tool scoped in 1.0 until Codex/Gemini tool adapters meet the same evidence/privacy bar.
- Advisor remains read-only and advisory.

### CLI & diagnostics
- Added `varnqor providers` / `varnqor agents`.
- Updated `status`, `analyze`, and `forecast` for aggregate multi-provider telemetry.
- Updated Doctor to detect supported providers without treating optional missing providers as failures.
- Added a 1.0 data-model hygiene check to Doctor.
- JSON output remains available for automation.

### Privacy
- No prompt text, source-code bodies, or tool-result bodies are required for usage analysis.
- Claude tool metadata continues to store only privacy-minimized evidence needed by Waste heuristics.
- Configured budgets are explicitly labeled as user-configured, never official provider quotas.

### Build & release
- Version set to `1.0.0`.
- Windows Wails smoke builds, CLI smoke tests, deterministic app icon generation, and SHA-256 checksums remain release gates.

## 0.1.0-alpha.1 — development line

The pre-1.0 line established the original Claude Code collector, SQLite history, Burn Rate, Attribution, Waste Intelligence, Forecasting, Advisor, configured budgets/runway, live refresh, CLI, and Windows development builds.
