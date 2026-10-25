# Local Usage Store

Milestone 2 moves Varnqor from full transcript rescans to an incremental local index backed by SQLite.

## Why SQLite

Varnqor is a desktop application and the usage database belongs to the user. SQLite keeps the store embedded, transactional and portable without requiring a server process.

The Go driver is `modernc.org/sqlite`, which is CGo-free. This supports Varnqor's goal of straightforward cross-platform packaging.

## What is stored

The database stores resource metadata needed for analytics:

- provider and deduplication key;
- timestamp, model, project and session identity;
- input/output/cache token counters;
- whether the call came from a subagent transcript;
- source file path and byte offset;
- per-file incremental scan state.

Prompt text, response content, tool arguments and source code are not stored by the usage index.

## Incremental indexing

For every transcript Varnqor records:

```text
path
observed file size
modification time
last complete JSONL byte offset
checkpoint hash
parse warning count
```

On the next scan, Varnqor starts at the last complete-line offset instead of re-reading the file from byte zero.

A trailing partial JSONL line is deliberately not committed. The offset remains before that line so it can be retried after Claude finishes writing it.

## Safe rebuilds

Append-only growth is the fast path. Varnqor rebuilds the Claude index when it sees evidence that the source history changed underneath it, including:

- an indexed transcript disappears;
- a transcript shrinks;
- a completed file changes without growing;
- the checkpoint hash before the saved offset no longer matches.

The rebuild favors correctness over preserving a stale cache.

## Deduplication

The SQLite primary key is `(provider, dedupe_key)`. This keeps Claude call identifiers isolated from future Codex/Gemini adapters.

When a streaming snapshot conflicts with an existing row, the store keeps the better observation using the same ordering as the Usage Engine: final record first, then larger output, then larger total activity, then later timestamp.

## Hourly history

Today's hourly series is derived from indexed calls in the user's local timezone. Empty hours are returned as zero-valued buckets so the desktop UI can render a stable daily timeline.
