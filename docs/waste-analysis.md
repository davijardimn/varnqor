# Waste Analyzer

Waste Analyzer is Varnqor's evidence layer for potentially avoidable resource usage.

The first rule is intentionally narrow:

> Detect an exact file read repeated soon after an earlier read when Varnqor did not observe a mutation to that file in between.

This is a heuristic. It does not prove that the repeated read was unnecessary.

## v0.1 rule — Repeated File Reads

A `Read` call is considered a repeat only when all of the following match:

- same Claude transcript scope;
- same file path;
- same `offset`;
- same `limit`;
- repeated within 15 minutes;
- no observed `Edit`, `Write`, `MultiEdit`, or `NotebookEdit` to that file in the same transcript between the reads;
- the repeated `Read` did not return an error.

This deliberately ignores looser patterns such as reading different ranges of the same file.

## Why transcript scope matters

Subagents can inspect the same file independently. Treating reads from different transcripts as one sequence would create false positives, so the current rule scopes repetition to the transcript that issued the tool call.

## Result-size metadata

Varnqor does **not** store tool-result content.

For a `tool_result`, it stores only the byte length of the returned payload. This allows the analyzer to estimate how much context was returned again without retaining the source text itself.

For string results, the byte count is taken from the decoded string. For structured results, the compact JSON payload length is used as an approximation.

## Token estimate

The v0.1 token estimate is deliberately simple:

```text
estimated repeated tokens = repeated result bytes / 4
```

The calculation rounds upward to the nearest token.

This is **not provider-reported token usage**. Tokenization varies by model and content, so the number is only a rough context-volume estimate.

For that reason, the repeated-read rule never reports confidence above `medium`.

## Confidence

- `low` — one repeated read, or result-size evidence is incomplete;
- `medium` — at least two repeated reads and result-size metadata is available;
- `insufficient` — no finding matched the rule.

Confidence describes the strength of the repeated-read signal, not the probability that the usage was actually wasteful.

## Severity

Severity is based on the approximate repeated-context token volume for one finding:

| Estimated tokens | Severity |
| --- | --- |
| `< 5K` | low |
| `5K - <25K` | medium |
| `>= 25K` | high |

Severity does not imply provider quota pressure or financial cost.

## Analysis window

Waste Analyzer currently uses the same rolling 60-minute window as Burn Rate and Attribution:

```text
window = [now - 60m, now)
```

This keeps the product narrative aligned: Varnqor can show the current burn, explain its drivers, and then show any waste signals found in that exact period.

## Privacy boundary

The local tool index stores only metadata required for analysis:

```text
tool use id
timestamp
tool name
project/session identifiers
main agent vs subagent
file path when present
Read offset/limit when present
tool-result byte length
error flag
```

It does not store:

```text
prompt text
assistant text
file contents
tool-result text
Bash commands
Write/Edit payloads
```

The analyzer runs locally against SQLite.

## What this rule cannot know

A repeated read can be legitimate. Examples include:

- the agent intentionally refreshing context;
- external processes changing a file without a Claude `Edit`/`Write` event;
- a tool result being needed again after context compaction;
- a model choosing to re-read for correctness.

Varnqor therefore uses the wording **potential repeated context** and **potentially avoidable**, never `wasted tokens` as a factual claim.

## Future rules

Planned Waste Intelligence can add separate evidence-based rules for:

- repeated search/grep operations;
- oversized tool responses;
- context duplication;
- low cache efficiency;
- runaway subagent fan-out;
- repeated failed operations.

Each rule should remain independently explainable and testable.
