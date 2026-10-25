# Usage Engine

The Claude usage engine reads token-bearing assistant records from local Claude Code JSONL transcripts.

Its purpose is measurement, not conversation inspection.

## Fields

For each API call Varnqor extracts only structural/resource metadata needed for usage analytics:

- timestamp;
- message ID and request ID for deduplication;
- model;
- project/session location;
- input tokens;
- output tokens;
- cache creation tokens;
- cache read tokens;
- 5-minute and 1-hour cache creation split when present;
- stop reason, used to prefer final streaming records.

Prompt text, response content, tool arguments and source code are not retained by the usage engine.

## Token semantics

`input_tokens` is treated as uncached input. Cache creation and cache read tokens are kept as separate categories.

Varnqor's `total activity` is:

```text
input + output + cache creation + cache read
```

It is a resource activity measure, not a claim about provider quota consumption or billing.

`cache read share` is:

```text
cache read / (input + cache creation + cache read)
```

It describes the share of prompt-side tokens served from cache. It is not labelled as an official provider cache-hit metric.

## Deduplication

Claude Code may write multiple streaming snapshots for the same API call. The same call can also appear in more than one transcript file.

Varnqor deduplicates by:

1. `message.id + requestId` when both exist;
2. `message.id` when request ID is missing;
3. `requestId` when message ID is missing;
4. file-local identity only as a final fallback.

When duplicates exist, Varnqor prefers:

1. a record with a non-empty `stop_reason`;
2. the record with the larger `output_tokens` value;
3. the record with the larger total token snapshot;
4. the later timestamp.

This avoids first-stream-snapshot undercounting and naive per-line overcounting.

## Scope

Milestone 1 scans local JSONL files recursively so subagent usage can be observed. Project and parent-session identity are derived from the local transcript layout.

SQLite incremental indexing is intentionally deferred. The current scanner is the correctness-first baseline against which the indexed implementation can later be tested.
