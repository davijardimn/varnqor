# Advisor Engine

Varnqor Advisor turns existing measurements into prioritized recommendations.

It consumes:

```text
Burn Rate
Forecast / Resource Pressure
Attribution
Waste Intelligence
```

## Advisory-only boundary

The current Advisor is read-only. It does not:

- modify prompts;
- compact context;
- switch models;
- stop or spawn agents;
- change provider settings;
- edit files;
- execute tool calls.

It only explains evidence and suggests a next action the user may choose.

## Current recommendation families

### Resource pressure

When pressure is elevated or higher, Advisor can recommend narrowing the next step or deferring nonessential parallel exploration.

### Waste-rule guidance

Each Waste Intelligence rule maps to a specific recommendation:

- repeated file reads → reuse recent file context when still valid;
- repeated Grep/Glob → reuse or narrow search;
- oversized responses → narrow Read/search scopes;
- repeated failures → inspect the cause or change input before retrying.

The per-rule heuristic context estimate is carried into the recommendation as evidence. It is not labelled as saved or wasted tokens.

### Subagent pressure

When resource pressure is elevated and subagents account for at least 35% of the current burn window, Advisor can suggest reviewing fan-out. This is not a claim that subagents are wasteful.

### Session concentration

When pressure is elevated and one session accounts for at least 60% of the window, Advisor can point to that session as the best place to investigate the burn.

## Priority

Recommendations are ranked `high`, `medium`, or `low` from evidence strength, signal volume and resource pressure. The desktop shows at most five recommendations to avoid turning diagnostics into noise.

Priority is not an instruction to automatically act.
