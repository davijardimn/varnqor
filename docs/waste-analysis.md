# Waste Analyzer

Waste Analyzer is Varnqor's evidence layer for **potentially avoidable** resource usage.

It does not declare tool activity to be waste. Each rule emits a separate, inspectable signal with its own evidence and confidence.

The current analysis window is the same rolling 60 minutes used by Burn Rate and Attribution:

```text
window = [now - 60m, now)
```

## Why there is no Waste Score yet

The current rules can overlap. For example, a repeated file read may also return an oversized response.

Varnqor therefore exposes `estimated signal context` as a rough diagnostic volume and explicitly warns that overlap is possible. It does **not** sum the rules and call the result `wasted tokens`.

A future Waste Score must define overlap handling and be benchmarked against task success before it becomes a product metric.

## Rule 1 — Repeated File Reads

A `Read` call is considered a repeat only when all of the following match:

- same Claude transcript scope;
- same file path;
- same `offset`;
- same `limit`;
- repeated within 15 minutes;
- no observed `Edit`, `Write`, `MultiEdit`, or `NotebookEdit` to that file in the same transcript between the reads;
- the repeated `Read` did not return an error.

Different file ranges are intentionally not merged.

## Rule 2 — Repeated Search

The current search rule covers `Grep` and `Glob`.

A search is considered repeated when:

- it runs in the same transcript;
- the tool name matches;
- the canonical tool-input fingerprint matches;
- it runs again within 10 minutes;
- neither matching call is an error.

Varnqor stores the search scope/path when available, but it does not persist the raw Grep pattern, glob expression, or other tool input.

## Rule 3 — Oversized Tool Responses

A successful tool result is flagged when its decoded/serialized result payload exceeds:

```text
32 KiB
```

The rule does **not** assume all of that result was unnecessary. Its estimate includes only volume above the threshold:

```text
estimated excess context = (result bytes - 32 KiB) / 4
```

Responses at or above 128 KiB receive `medium` signal confidence; smaller matches stay `low` confidence.

A large response may be completely legitimate. This rule is a context-pressure signal, not a correctness judgment.

## Rule 4 — Repeated Failed Operations

A failed operation becomes a repeat signal when:

- the same tool is used in the same transcript;
- the canonical input fingerprint matches;
- the operation fails again within 10 minutes.

A successful identical operation resets the failure sequence.

The token estimate uses only repeated error-result bytes. It does not attempt to guess the model reasoning or API usage surrounding the retry.

## Tool-input fingerprints

To compare operations without retaining their inputs, Varnqor canonicalizes the tool input JSON and stores:

```text
SHA-256(canonical tool input JSON)
```

The original input is discarded after parsing.

This means raw Bash commands, Grep patterns, Write/Edit payloads, and other tool inputs are not stored in the SQLite tool index. The fingerprint is a local comparison mechanism, not encryption, and it is never uploaded by default.

Updating from the v0.1 tool index triggers a one-time local transcript backfill so historical tool calls receive fingerprints. Afterwards scanning remains incremental.

## Result-size metadata

Varnqor does **not** store tool-result content.

For a `tool_result`, it stores only the byte length of the returned payload. For string results, the byte count is taken from the decoded string. For structured results, compact JSON payload length is used as an approximation.

## Token estimates

Where a rule needs a context-volume estimate, v0.1 uses:

```text
estimated tokens = relevant result bytes / 4
```

The calculation rounds upward. This is **not provider-reported token usage**; tokenization varies by model and content.

## Confidence

The current rules intentionally cap confidence at `medium`.

- `low` — a weak/single repeat signal, incomplete result-size evidence, or a moderately large response;
- `medium` — repeated evidence with result-size metadata, or a very large response;
- `insufficient` — no current finding matched.

Confidence describes evidence strength, not the probability that usage was actually unnecessary.

## Severity

Severity is based on one finding's approximate context volume:

| Estimated tokens | Severity |
| --- | --- |
| `< 5K` | low |
| `5K - <25K` | medium |
| `>= 25K` | high |

Severity does not imply provider quota pressure or financial cost.

## Privacy boundary

The local tool index stores only metadata required for analysis:

```text
tool use id
timestamp
tool name
project/session identifiers
main agent vs subagent
file/search scope path when present
Read offset/limit when present
SHA-256 input fingerprint
tool-result byte length
error flag
```

It does not store:

```text
prompt text
assistant text
file contents
tool-result text
raw Bash commands
raw Grep patterns
Write/Edit payloads
raw tool inputs
```

## What the analyzer cannot know

A matching signal can be legitimate. Examples include context refresh after compaction, externally modified files, broad exploration that truly requires a large result, or a retry that fixes transient state outside the tool input.

Varnqor therefore uses wording such as **potential repeated context**, **large response**, and **repeated failure signal** rather than making factual claims of waste.
