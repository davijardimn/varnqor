# Changelog

All notable Varnqor changes are tracked here.

## 3.0.0 — development

### Optimize
- Added the first Varnqor Optimize engine, completing the original intelligence loop through reviewable optimization plans.
- Added configurable Balanced, Economy, and Throughput goals, target reduction, projection horizon, and Preview/Manual modes.
- Added bounded playbooks for repeated tool context, dominant-session checkpointing, sustained burn, cache reuse, session fragmentation, and Economy output budgets.
- Added per-action evidence, priority, confidence, scope, steps, and heuristic impact ranges.
- Marked optimization impact ranges explicitly non-additive to prevent overlapping savings from being presented as guaranteed totals.
- Kept external auto-execution disabled: Varnqor 3 does not silently rewrite prompts, switch models, kill agents, or run tools.
- Added `varnqor optimize` with the same temporal/provider/project/session/model scope as Explorer.

### Session intelligence
- Added exact session filtering to the local Explorer query engine and CLI.
- Added Session Lab to the desktop with scoped totals, calls, average tokens/call, cache-read share, timeline, activity bounds, provider identity, and recent calls.
- Added one-step routing from a hot session into a session-scoped Optimize plan.
- Prevented global learned baselines from being presented as session/project/model-specific baselines.

### Desktop generation 3
- Replaced the V2 frontend with a single V3 desktop architecture.
- Reoriented Overview around current posture, agent fleet, concentration, and the highest-value Optimize actions.
- Added dedicated Sessions and Optimize surfaces.
- Added a new warm light enterprise design system and restrained graphite dark design system.
- Added Light, Dark, and System appearance choices with immediate switching and local persistence.
- Kept Varnqor branding out of the sidebar workspace.
- Preserved Ctrl/Cmd+K navigation and added theme switching to the command palette.

### Localization
- Added settings schema v3 with automatic migration from V1/V2 preferences.
- Added System, English, Portuguese (Brazil), Spanish, German, and French desktop language choices.
- Added centralized message fallback so diagnostics/provider-generated dynamic text can safely fall back to English.
- Added CLI configuration for language and theme preferences.

### Agent identity
- Added locally bundled Claude Code, Codex, and Gemini CLI vector marks to provider badges, fleet cards, session rows, and provider detail surfaces.
- Added third-party notices for the MIT-licensed Lobe Icons source vectors and trademark identification.
- No CDN/network request is required to render provider marks.

### CLI and release gates
- Added `varnqor config language` and `varnqor config theme`.
- Added `varnqor config optimize <balanced|economy|throughput|off>` with target/horizon/mode options.
- Extended Windows CI and release smoke tests to execute theme/language/optimization settings and `varnqor optimize` against the real Wails executable.
- Development version advanced to `3.0.0-dev`.

### Privacy and correctness
- Optimize runs on existing local usage metadata and privacy-minimized Waste evidence.
- No new prompt, source-code body, or raw tool-result persistence is required.
- Session-scoped investigation preserves provider identity and does not claim unavailable scope-specific baselines.
- Optimization target scenarios are not provider quotas.

## 2.0.0 — 2026-10-26

### Investigation workspace
- Added Explorer with selectable windows from 15 minutes to 7 days and desktop presets for 1h, 6h, 24h, and 7d.
- Added provider, project, and model filtering over normalized local usage.
- Added automatic comparison against the immediately preceding equal-length window.
- Added adaptive timeline buckets, provider/project/model dimensions, hot-session ranking, and recent token-bearing call history.
- Added JSON and CSV snapshots for the selected Explorer window.

### Pulse
- Added explainable local guardrails instead of an opaque health/waste score.
- Added token-spike, project-concentration, session-concentration, low-cache-reuse, and session-fragmentation signals.
- Added local guardrail configuration and automatic settings migration from v1.
- Added a minimum-volume threshold to avoid firing signals on trivial usage.
- Pulse is advisory only and never changes agent configuration or behavior.

### Desktop
- Rebuilt the active frontend around Overview, Explorer, Pulse, Activity, Burn, Advisor, Projects, Models, Agents, and Settings.
- Added global range switching and context-aware Explorer filters.
- Added Ctrl/Cmd+K quick navigation and range commands.
- Overview, Activity, Projects, and Models use the same selected temporal window rather than mixing unrelated periods.
- Removed the inactive v1 frontend and design system so only one desktop architecture ships.
- Kept the workspace brand-free: no logo or wordmark occupies the sidebar.

### CLI
- Added `varnqor explore` with range/provider/project/model/limit filters.
- Added `varnqor pulse` for deterministic guardrail output.
- Added `varnqor export` with JSON and CSV output.
- Extended Windows smoke tests to execute the new commands against the built executable.

### Data and correctness
- Added temporal window summaries and equal-window comparison directly on the local SQLite index.
- Added provider-aware dimension/session queries and recent-call retrieval.
- Kept zero-token/internal placeholder records excluded from investigation surfaces.
- Preserved the provider capability boundary: Claude tool metadata drives Waste; Codex/Gemini usage contributes to resource intelligence without pretending tool parity.

### Privacy
- Explorer and Pulse operate entirely on the existing local metadata index.
- No new prompt, source-code body, or raw tool-result persistence is required for Varnqor 2 investigation features.

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
- Global burn/forecast includes supported token usage from Claude, Codex, and Gemini.
- Waste Intelligence remains explicitly Claude-tool scoped until Codex/Gemini tool adapters meet the same evidence/privacy bar.
- Advisor remains read-only and advisory.

### CLI & diagnostics
- Added `varnqor providers` / `varnqor agents`.
- Updated `status`, `analyze`, and `forecast` for aggregate multi-provider telemetry.
- Updated Doctor to detect supported providers without treating optional missing providers as failures.
- Added a data-model hygiene check to Doctor.
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
