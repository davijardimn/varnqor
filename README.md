<div align="center">

# Varnqor

### Resource intelligence for AI coding agents.

**Observe. Understand. Predict. Optimize with evidence.**

[![Version](https://img.shields.io/badge/version-3.0.0--dev-111318)](#varnqor-3)
[![License: MIT](https://img.shields.io/badge/license-MIT-111318)](LICENSE)
[![Local first](https://img.shields.io/badge/local--first-yes-397A63)](#privacy)

</div>

---

Varnqor is a **local-first desktop application and CLI** for understanding and improving resource usage across AI coding agents. It reads supported local session telemetry, normalizes token activity into a common model, attributes usage to providers/projects/sessions/models, detects pressure and waste, forecasts near-term behavior, and turns evidence into bounded optimization playbooks.

```text
OBSERVE -> NORMALIZE -> ATTRIBUTE -> EXPLORE -> ANALYZE -> PREDICT -> ADVISE -> OPTIMIZE
```

Varnqor does **not** bypass provider limits, upload prompts, inspect source-code bodies for analytics, silently rewrite prompts, switch models, or execute external agent actions without the user.

## Varnqor 3

Version 3 closes the original Varnqor loop as far as the current evidence model can support safely: from observation all the way to **reviewable optimization plans**.

### Optimize

Optimize converts the selected local usage window into a target scenario and a small set of explainable playbooks. The engine can surface actions around:

- repeated or oversized tool context
- dominant long-running sessions that may benefit from a checkpoint
- sustained burn above an established active baseline
- weak cache reuse when cache telemetry is actually available
- session fragmentation
- verbose routine output when the user explicitly selects Economy mode

Every playbook contains evidence, confidence, steps, scope, and a bounded impact range when Varnqor has enough information to estimate one.

**Impact ranges are heuristic and non-additive.** Multiple actions may affect the same future tokens, so Varnqor never sums them into a fake guaranteed saving.

The optimization target is a separate scenario, configurable by the user. Goals are:

- **Balanced** — reduce resource pressure without aggressively trading away workflow flexibility
- **Economy** — prioritize lower resource usage
- **Throughput** — preserve pace while removing avoidable overhead

Varnqor 3 supports Preview and Manual modes. Both remain user-controlled: external automatic execution is intentionally unavailable.

### Session Lab

Sessions are now first-class investigation units. The desktop can open a hot session and inspect its:

- scoped token total
- call count
- average tokens per call
- cache-read share
- adaptive activity timeline
- recent token-bearing calls
- provider identity
- first/last activity
- selected-window share

A session can be sent directly into Optimize. Scoped analysis does **not** borrow the global burn baseline and pretend it belongs to that session.

### Adaptive desktop

Varnqor 3 ships a new desktop generation rather than layering features onto the old dashboard.

The UI supports:

- **Light**, **Dark**, or **System** appearance
- persistent local appearance settings
- system-theme following when requested
- a warm, restrained light enterprise design
- a graphite/copper dark design without neon dashboard styling
- `Ctrl+K` / `Cmd+K` command palette
- global range switching
- provider/project/model/session scoping
- decision-oriented Overview with the highest-value Optimize actions

The app keeps the Varnqor logo/wordmark out of the sidebar workspace so it behaves like a desktop product instead of a website inside a window.

### Languages

The desktop language can be selected independently of the operating-system theme.

Included locales:

- System language
- English
- Português (Brasil)
- Español
- Deutsch
- Français

English is the canonical fallback for any untranslated low-level diagnostic or provider-generated text. Language and theme preferences stay in the local Varnqor settings file.

### Agent identity

Supported providers now use recognizable locally bundled marks in the Agents view, badges, session rows, and attribution surfaces:

- Claude Code
- Codex
- Gemini CLI

No CDN request is made to render those marks. See [THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md) for icon licensing and trademark attribution.

## Explorer and Pulse

Explorer retains the Varnqor 2 investigation engine with windows from 15 minutes to 7 days and desktop presets for **1h, 6h, 24h, and 7d**.

Filters include:

- provider
- project
- session
- model

Every selected window is compared with the immediately preceding equal-duration window. When the previous window has no comparable data, Varnqor reports the activity as **NEW / unavailable** instead of fabricating a percentage.

Pulse remains an explainable local guardrail engine. Default signals cover token spikes, project concentration, session concentration, cache-read floors when cache telemetry exists, and fragmented session activity.

## Supported local providers

| Provider | Usage telemetry | Tool metadata | Waste analysis | Notes |
| --- | --- | --- | --- | --- |
| Claude Code | Yes | Yes | Yes | Deepest integration |
| Codex | Yes | Not yet | Not yet | JSONL sessions; compressed `.jsonl.zst` files are reported but not indexed |
| Gemini CLI | Yes | Not yet | Not yet | Current JSONL and legacy JSON chat sessions |

Provider support is capability-based. Varnqor never treats a detected executable as equivalent to a fully supported telemetry source.

## Intelligence layers

### Burn
Burn Rate compares the latest rolling 60-minute pace with the median of completed active hours. Until enough active baseline history exists, Varnqor explicitly stays in **Learning**.

### Resource pressure and forecast
The forecast engine uses a transparent persistence model over observed activity. Its pace envelope is context for current behavior; it is **not a statistical confidence interval and not provider quota data**.

### Trusted runway
Runway is enabled only when a trusted limit exists. The universally supported source remains a **user-configured local budget**, always labeled as configured rather than official provider quota.

### Waste intelligence
Claude Code tool metadata enables explainable heuristics for repeated reads/searches, oversized responses, and repeated failed operations. Raw tool-result bodies are not persisted by the analytics index.

### Advisor
Advisor combines burn, attribution, forecast, and Waste evidence into prioritized recommendations.

### Optimize
Optimize goes one step further: it converts current evidence into a target scenario and reviewable playbooks. It remains advisory toward external agents in Varnqor 3.

## Desktop surfaces

```text
Observe
  Overview
  Explorer
  Sessions
  Pulse

Intelligence
  Optimize
  Activity
  Burn
  Advisor
  Projects
  Models

System
  Agents
  Settings
```

## CLI

```bash
varnqor
varnqor status
varnqor providers
varnqor explore --range 6h
varnqor explore --range 24h --provider codex --json
varnqor explore --range 6h --provider claude --session <id>
varnqor pulse --range 1h
varnqor optimize --range 6h
varnqor optimize --range 1h --provider claude --json
varnqor export --range 24h --format csv
varnqor analyze
varnqor forecast
varnqor doctor
varnqor config show
varnqor config language pt-BR
varnqor config theme dark
varnqor config optimize economy --target 25 --horizon 360 --mode preview
varnqor config budget 10m
varnqor version
```

Investigation commands accept provider/project/session/model/range filters. Machine-readable JSON remains available for automation; export supports JSON and CSV.

## Privacy

Varnqor is designed around a local privacy boundary.

By default:

- no Varnqor cloud account
- no prompt upload
- no source-code body upload
- no raw tool-result body persistence
- no API key required for local usage analysis
- no product telemetry required for analysis
- no remote service required for themes, translations, or provider marks

The usage index stores resource metadata such as token counts, timestamps, models, provider/project/session labels, and source offsets. Claude tool analysis stores privacy-minimized evidence such as tool name, safe range fields, fingerprints, result byte size, and error state.

See [docs/privacy.md](docs/privacy.md) for the detailed boundary.

## Data correctness

Varnqor separates concepts that are easy to conflate:

- **observed activity** — a local session/file exists
- **token-bearing activity** — a supported usage event contains resource counts
- **provider health** — supported local data can be read and indexed
- **global baseline** — learned aggregate activity; never silently reused as a scoped session/project baseline
- **optimization target** — a user-selected scenario, not a provider limit
- **estimated impact** — a heuristic non-additive range, not guaranteed savings

Provider identity is preserved in attribution so identical project/session/model labels from different agents do not collide.

## Local storage

Historical telemetry and preferences are local. Append-only session files resume from stored offsets. Rewritten formats such as legacy Gemini JSON are safely re-indexed when changed.

Explorer, Session Lab, Pulse, Advisor, and Optimize read the local index; they do not call provider APIs to reconstruct history.

## Build

Varnqor uses Go, Wails, React + TypeScript, and SQLite.

```bash
go test ./internal/...
go vet ./internal/...
cd frontend && npm install && npm run build
wails build
```

GitHub Actions performs core tests/vet, frontend production compilation, a real Windows Wails build, CLI smoke tests including Optimize and adaptive settings, deterministic application-icon generation, and SHA-256 checksum generation.

## Intentional boundaries

- Claude remains the only provider with tool-level Waste analysis until other adapters meet the same evidence/privacy bar.
- Compressed Codex session archives are reported but not silently decoded.
- Provider limits are not assumed to be interchangeable.
- Optimize does not automatically rewrite prompts, switch models, kill agents, run tools, or mutate provider settings.
- Explorer retains a maximum selected window of 7 days even if the local database contains older history.
- Dynamic provider/diagnostic evidence may fall back to English even when the desktop shell is localized.

Varnqor prefers **unavailable over fabricated precision**.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md). Provider adapters should preserve the local privacy boundary and include parser fixtures/tests before being advertised as supported.

## License

Varnqor is MIT licensed. See [LICENSE](LICENSE). Third-party icon notices are in [THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md).

---

<div align="center">

**Control the burn.**

</div>
