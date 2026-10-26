# Provider support

Varnqor 1.0 uses capability-based provider adapters. A provider may support usage telemetry without supporting tool metadata or quota/limit sources.

| Provider | Detection | Usage | Tool metadata | Waste | Limits |
| --- | --- | --- | --- | --- | --- |
| Claude Code | binary + `~/.claude` | yes | yes | yes | configured budget only |
| Codex | binary + `~/.codex` | yes for JSONL sessions | no | no | configured budget only |
| Gemini CLI | binary + `~/.gemini` | yes for JSON/JSONL chat sessions | no | no | configured budget only |

## Provider states

The Agents view uses explicit states:

- `connected` — supported local data is present and can be indexed;
- `partial` — supported data is indexed but additional files/features were detected that Varnqor cannot read safely yet;
- `ready` — the provider executable is installed but local session telemetry is not present;
- `unavailable` — neither supported local data nor executable was detected;
- `error` — the provider was detected but its adapter hit an indexing error.

These states describe Varnqor's local observability, not the provider account's authentication or billing status.

## Claude Code

Claude is the deepest 1.0 integration. Varnqor reads local JSONL transcripts incrementally and extracts token accounting plus privacy-minimized tool metadata. Tool inputs/results are not persisted as raw content.

## Codex

Varnqor discovers session JSONL files under the local Codex session stores. Current rollouts are accounted from per-response `token_usage_record` entries and deduplicated by response ID. Legacy rollouts that contain only cumulative `token_count` snapshots use the latest-turn contribution as a fallback; when per-response records are present, overlapping snapshots are ignored. Current model context and session metadata are used when available.

Compressed `.jsonl.zst` archives are counted and shown as unsupported/partial in 1.0. Varnqor does not claim those files were analyzed.

## Gemini CLI

Varnqor discovers chat sessions under the Gemini local temporary project store. Current append-only JSONL sessions and legacy JSON conversations are supported for usage telemetry. JSONL session metadata supplies the session ID, project directory, and main/subagent kind when available.

Legacy JSON files may be rewritten rather than appended, so Varnqor re-indexes those sources safely when they change. Repeated message IDs in JSONL are upserted so later token-bearing revisions replace earlier versions rather than being double-counted.

## Adding a provider

A new adapter should not be advertised as supported until it has:

1. deterministic local detection;
2. documented source formats;
3. token normalization rules;
4. parser fixtures/tests;
5. incremental or safely repeatable indexing;
6. privacy review;
7. explicit capability flags.

Do not infer official quotas from local usage. Provider limits need an independently trusted source before Varnqor can label them official/observed.
