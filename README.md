<div align="center">

# Varnqor

### Local resource control plane for AI coding agents.

**Observe. Understand. Decide. Control the burn.**

[![Version](https://img.shields.io/badge/version-4.0.0--dev-111318)](#varnqor-4)
[![License: MIT](https://img.shields.io/badge/license-MIT-111318)](LICENSE)
[![Local first](https://img.shields.io/badge/local--first-yes-7267FF)](#privacy)

</div>

---

Varnqor is a **local-first Windows desktop application and CLI** that turns local AI coding-agent resource telemetry into an operational control picture.

```text
OBSERVE -> NORMALIZE -> ATTRIBUTE -> EXPLORE -> ANALYZE -> PREDICT -> ADVISE -> OPTIMIZE
                                                     \-> CONTROL PLANE
```

It supports Claude Code, Codex and Gemini CLI telemetry without requiring a Varnqor cloud account. Varnqor does not bypass provider limits, reconstruct prompts, upload source-code bodies, or silently take control of an external agent.

# Varnqor 4

V4 changes the product model. Burn, Forecast, Pulse, Advisor and Optimize are no longer a collection of disconnected dashboard destinations. They become evidence feeding a **Control Plane**.

The active desktop is organized around:

```text
Control Room
Missions
Investigate
Optimize
Policies
Agents
Settings
```

There is no Varnqor logo or wordmark occupying the workspace rail. The app behaves like an operational desktop tool rather than a website inside an executable.

## Control Room

Control Room is the default operating surface. It combines:

- current resource pace and learned baseline context
- near-term projected token activity
- active and cooling missions
- provider health
- explainable policy incidents
- a prioritized decision queue
- configured-budget runway when a trusted local budget exists
- the highest-value bounded Optimize interventions

There is deliberately **no opaque health score**. Every Watch / Intervene / Critical state is backed by visible evidence or a visible policy threshold.

## Missions

A mission is Varnqor's operational view of a recent agent session. Sessions are classified as:

- **Active** — recent token-bearing activity inside the configured active window
- **Cooling** — recently active and still operationally relevant
- **Idle** — outside the cooling window

Each mission preserves provider identity and shows project, session, token share, velocity, calls, recency and pressure. A mission can be focused directly into Investigate or Optimize.

Mission state is local inference from activity metadata, not a claim that Varnqor controls the provider process.

## Incidents

The Control Plane creates deterministic incidents from evidence such as:

- provider ingestion failures or partial observability
- learned burn pace crossing the configured ceiling
- project concentration
- session / mission concentration
- weak cache reuse when cache telemetry is actually available
- configured local budget pressure

Provider-health failures outrank a quiet token window: Varnqor will not report **Idle** when its telemetry source is broken.

## Decision Queue

Decisions combine control-specific interventions with existing Optimize playbooks. Examples include:

- repair a broken provider adapter before trusting fleet totals
- contain the next work block when a configured budget is at risk
- checkpoint a dominant session
- reduce repeated tool context
- improve reusable context boundaries
- reduce fragmentation
- bring burn back toward the learned active envelope

Every decision remains evidence-backed and reviewable. Estimated impact ranges are heuristic and **non-additive** because multiple interventions can affect the same future tokens.

## Policies

V4 introduces an explicit local policy layer. Three presets provide coherent starting points:

| Profile | Intent |
| --- | --- |
| Balanced | Catch material waste without overreacting to normal productive bursts |
| Economy | Intervene earlier and prioritize resource reduction |
| Throughput | Allow higher concentration and burn before surfacing intervention |

Every preset expands into visible thresholds. Users can tune:

- burn ceiling
- session dominance
- project dominance
- cache-read floor
- configured-budget warning threshold
- active mission window
- cooling mission window
- decision horizon
- minimum evaluation volume

Policies decide **what Varnqor surfaces**, not what an external agent is allowed to do.

## Control Brief

Control Room can generate a portable markdown brief containing only resource metadata, incidents and decisions. It is designed to be pasted into a coding agent or issue without reconstructing prompt text or source-code bodies.

CLI equivalent:

```bash
varnqor brief --range 6h
varnqor brief --provider claude --project my-project --range 1h
```

## Visual system

V4 returns to the visual identity that worked best in V2, while using a completely new interface architecture.

Primary dark palette:

```text
Background     #0A0C10
Rail           #0D1015
Surface        #10141A
Surface raised #171D27
Violet         #7267FF
Violet light   #9C95FF
Blue           #5B91FF
Green          #38D6A2
```

Dark is the V4 default. Light and System remain available, but the light theme uses cool neutral surfaces with the same V2 accent system rather than the V3 cream/copper direction.

## Agents

Provider marks are bundled locally for:

- Claude Code
- Codex
- Gemini CLI

The Agents surface reports local provider health, observed sessions, indexed calls, token activity and actual supported capabilities. Detection is not presented as full support when an adapter lacks evidence.

| Provider | Usage | Tool metadata | Waste intelligence |
| --- | --- | --- | --- |
| Claude Code | Yes | Yes | Yes |
| Codex | Yes | Not yet | Not yet |
| Gemini CLI | Yes | Not yet | Not yet |

See [THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md) for icon attribution and trademark notices.

## Investigation and optimization

V4 keeps the proven V2/V3 intelligence beneath the Control Plane:

- Explorer windows from 15 minutes to 7 days
- provider / project / session / model scope
- equal-window comparisons
- adaptive timelines
- Burn Rate against completed active-hour baseline
- persistence forecast and pace envelope
- configured-budget runway
- Claude Waste heuristics
- Advisor recommendations
- bounded Optimize target scenarios and playbooks

A scoped project/session/model view never borrows the global learned baseline and presents it as a scoped baseline.

## Appearance and languages

Appearance:

- Dark — V2 palette, default
- Light
- System

Languages:

- System language
- English
- Português (Brasil)
- Español
- Deutsch
- Français

English remains the fallback for low-level diagnostic/provider text that is generated dynamically.

## CLI

```bash
varnqor
varnqor control --range 1h
varnqor control --range 6h --json
varnqor brief --range 6h
varnqor providers
varnqor explore --range 24h --provider codex --json
varnqor explore --range 6h --provider claude --session <id>
varnqor pulse --range 1h
varnqor optimize --range 6h
varnqor forecast
varnqor doctor

varnqor config theme dark
varnqor config language pt-BR
varnqor config optimize economy --target 25 --horizon 360 --mode preview
varnqor config control economy
varnqor config control balanced --burn 1.70 --session 62 --project 72
varnqor config budget 10m
```

Control / Explore / Optimize accept provider, project, session, model and range scope where applicable.

## Privacy

Varnqor is local-first by design. By default:

- no Varnqor cloud account
- no prompt upload
- no source-code body upload
- no assistant-response upload
- no raw tool-result body persistence
- no hidden external execution
- no provider API key required for local telemetry analysis
- no CDN required for provider marks or UI localization

The usage index stores resource/accounting metadata such as timestamps, provider/project/session/model labels, token counts and source offsets. Claude tool analysis stores only privacy-minimized evidence needed by deterministic Waste rules.

The V4 Control Plane and Control Brief operate on this existing metadata boundary. They do not add prompt or source persistence.

See [docs/privacy.md](docs/privacy.md).

## Correctness boundaries

Varnqor distinguishes:

- **observed session** from token-bearing activity
- **provider health** from provider detection
- **learned aggregate baseline** from scoped activity
- **configured budget** from official provider quota
- **optimization target** from a provider limit
- **estimated impact** from guaranteed savings
- **mission activity** from provider process control
- **decision support** from automatic external execution

Varnqor prefers **unavailable over fabricated precision**.

## Build and release gates

Varnqor uses Go, Wails, React + TypeScript and SQLite.

```bash
go test ./internal/...
go vet ./internal/...
cd frontend && npm install && npm run build
wails build
```

Stable Windows releases are blocked unless the actual executable passes:

- Go tests + vet
- TypeScript / Vite production build
- Wails Windows build
- provider / Doctor smokes
- V4 settings migration and persistence
- V2-dark theme preference
- Optimize configuration
- Control policy configuration
- `explore`, `pulse`, `optimize`, `control`, `brief` and `export`
- SHA-256 generation

## License

Varnqor is MIT licensed. See [LICENSE](LICENSE).

---

<div align="center">

**Control the burn.**

</div>
