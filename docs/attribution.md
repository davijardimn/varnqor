# Attribution

Attribution is Varnqor's explanation layer for resource usage.

The current implementation answers:

> Which projects, sessions, models and agent classes produced the activity in the current burn window?

## Window alignment

Attribution uses the same rolling 60-minute window as Current Burn Rate:

```text
window = [now - 60m, now)
```

This is intentional. A burn signal and its explanation must refer to the same data window.

## Dimensions

Varnqor currently groups indexed usage by:

- **project** — the project identifier recorded by the Claude Code local transcript layout;
- **session** — project + Claude session identifier;
- **model** — model name present in the assistant usage record;
- **agent class** — main transcript vs transcript discovered under subagent paths.

No prompt or response body is required for these groupings.

## Share

Each item's share is calculated from total token activity in the same window:

```text
share = item total tokens / window total tokens * 100
```

The displayed percentage is rounded to one decimal place. Ranking uses the unrounded token total.

`total tokens` follows Varnqor's existing Usage Engine accounting:

```text
input + output + cache creation + cache read
```

The share is an attribution of observed token activity, not provider cost or quota share.

## Ranking

Projects, sessions and models are ordered by total token activity, then by call count. The desktop currently asks for the top five entries and renders up to four to keep the overview compact.

Main-agent and subagent entries are both returned when present.

## Session identity

The database keeps the complete session identifier. The desktop may shorten it visually, but tooltips retain the complete project/session pair.

## Project identity

Varnqor does not attempt to reverse Claude Code's project-directory encoding into a filesystem path using guesses. The stored project identifier is shown as observed until a reliable source of a human-friendly project name is available.

## Privacy

Attribution is computed entirely from the local SQLite metadata index. It does not inspect prompt text, generated response text or source-code content.

## What attribution does not mean

A large share does not imply waste. A project can legitimately dominate a window because it is the only active workload.

Waste Intelligence will consume attribution as evidence, but must add separate rules and evidence before describing any usage as potentially avoidable.
