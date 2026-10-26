# Changelog

All notable Varnqor changes are tracked here.

## 4.0.0 — development

### Control Plane
- Reframed Varnqor from a set of observability dashboards into a local decision-support control plane.
- Added a unified Control Room fed by provider health, Explorer, Burn, Forecast, configured runway and Optimize.
- Added explicit `nominal`, `watch`, `intervene`, `critical`, `idle`, and `disabled` control states without introducing an opaque score.
- Added explainable incidents for provider health, burn ceiling, project concentration, mission/session concentration, cache-read floor and configured-budget pressure.
- Added a prioritized Decision Queue combining control-specific interventions with existing bounded Optimize playbooks.
- Provider ingestion failures and critical configured-budget conditions now outrank a quiet token window rather than being mislabeled as Idle.

### Missions
- Added operational Mission objects derived from recent provider/session activity.
- Added Active, Cooling and Idle lifecycle states with configurable local windows.
- Added mission pressure, selected-window share, normalized velocity, call count, provider identity and activity age.
- Added direct mission routing into scoped Investigate and Optimize workflows.
- Mission activity is explicitly an inference from local telemetry, not external process control.

### Policies
- Added settings schema v4 and automatic migration from V3.
- Added Balanced, Economy and Throughput Control Plane policy profiles.
- Added visible thresholds for burn ceiling, session/project dominance, cache floor, configured-budget warning, mission lifecycle, decision horizon and minimum evaluation volume.
- Preserved explicit user Light theme choices during migration while moving V3 System/default installs to the V2-dark identity.
- Added `varnqor config control` with preset selection and threshold overrides.

### Desktop generation 4
- Replaced the V3 desktop entrypoint completely; `AppV3.tsx` and its design system no longer ship.
- Restored the actual V2 dark color system: `#0A0C10`, `#0D1015`, `#10141A`, `#171D27`, `#7267FF`, `#9C95FF`, `#5B91FF`, and `#38D6A2`.
- Built a new compact icon rail and operational workspace rather than restoring the old V2 layout.
- Reorganized the product around Control Room, Missions, Investigate, Optimize, Policies, Agents and Settings.
- Kept the workspace free of Varnqor logo/wordmark chrome.
- Changed the native Wails startup background to V2 dark to avoid the old warm-light flash before React loads.
- Dark is the V4 default; Light and System remain supported using the same cool V2 accent family.

### Control Brief
- Added a privacy-safe markdown Control Brief containing resource posture, incidents, decisions and bounded action steps.
- Added `varnqor brief` with the same provider/project/session/model/range scope as investigation commands.
- Control Briefs do not reconstruct prompts, source-code bodies or raw tool-result bodies.

### CLI and release gates
- Added `varnqor control [--json]`.
- Added `varnqor brief`.
- Extended Windows CI and stable release smokes to verify settings schema v4, V2-dark theme persistence, control-policy persistence, Control Plane output and Control Brief output against the actual Wails executable.
- Development version advanced to `4.0.0-dev`.

### Safety and correctness
- External automatic execution remains unavailable: V4 does not silently rewrite prompts, change models, terminate sessions or execute provider tools.
- Scoped control views do not borrow the global learned burn baseline and present it as scope-specific evidence.
- Configured budgets remain local user budgets and are never labeled as provider quotas.
- Impact ranges remain heuristic and non-additive.

## 3.0.0 — 2026-10-26

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
- Added Session Lab with scoped totals, calls, cache share, timeline, provider identity and recent calls.
- Added one-step routing from a hot session into session-scoped Optimize.
- Prevented global learned baselines from being presented as session/project/model-specific baselines.

### Desktop generation 3
- Replaced the V2 frontend with one V3 desktop architecture.
- Added Light, Dark and System appearance choices, localization and provider marks.
- Added Sessions and Optimize surfaces while retaining Explorer/Pulse/Burn/Advisor intelligence.

### Localization and agent identity
- Added System, English, Portuguese (Brazil), Spanish, German and French desktop language choices.
- Added locally bundled Claude Code, Codex and Gemini CLI marks and third-party notices.

## 2.0.0 — 2026-10-26

### Investigation workspace
- Added Explorer with selectable windows from 15 minutes to 7 days and presets for 1h, 6h, 24h and 7d.
- Added provider/project/model filtering and equal-window comparisons.
- Added timeline buckets, attribution, hot-session ranking and recent token-bearing calls.
- Added JSON and CSV snapshots.

### Pulse
- Added deterministic local guardrails for token spikes, project/session concentration, cache reuse and fragmentation.
- Added local guardrail settings and minimum evaluation volume.

### Desktop and CLI
- Rebuilt the active frontend around investigation workflows.
- Added `varnqor explore`, `varnqor pulse`, and `varnqor export`.
- Preserved local-first privacy and provider-aware attribution.

## 1.0.0 — 2026-10-26

### Product and providers
- Promoted Varnqor to the first stable product model.
- Added Claude Code, Codex and Gemini CLI local provider adapters.
- Added provider capability/health reporting, multi-provider usage, attribution, Burn, Forecast, configured runway, Waste and Advisor.
- Kept Claude as the deepest tool-level adapter.

### Privacy and correctness
- Split observed sessions from token-bearing usage.
- Preserved provider identity in normalized history.
- Kept prompt text, source-code bodies and raw tool-result bodies outside the analytics index.
- Labeled configured budgets as user-configured rather than official quota.

## 0.1.0-alpha.1 — development line

The pre-1.0 line established the original Claude Code collector, SQLite history, Burn Rate, Attribution, Waste Intelligence, Forecasting, Advisor, configured budgets/runway, live refresh, CLI, and Windows development builds.
