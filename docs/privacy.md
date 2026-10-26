# Privacy

Varnqor is local-first by design.

## Default boundary

Varnqor 1.0 does not require a Varnqor cloud account or provider API key for local analysis. Its default product boundary is:

- no prompt upload;
- no source-code body upload;
- no assistant-response upload;
- no tool-result body persistence;
- no hidden background data export;
- no product telemetry required for local analysis;
- read-only access to supported agent session data.

## What the usage index stores

For supported providers, Varnqor stores only resource/accounting metadata needed for observability:

- provider;
- timestamp;
- model label when available;
- project/workspace label when available;
- session identifier;
- subagent flag when the provider exposes it;
- input/output/cache token counts;
- source-file path and source offset used for incremental indexing.

Observed session files and token-bearing usage are treated as different concepts. A local session can exist without producing a token usage row.

## Claude tool metadata

Claude Code currently has the deepest tool-level adapter. For explainable Waste rules, Varnqor may store:

- tool name;
- timestamp;
- project/session identifiers;
- file or safe scope path when exposed by the tool input;
- Read offset/limit;
- tool-result byte length;
- error state;
- SHA-256 fingerprint of canonical tool input.

Raw tool input is discarded after fingerprinting. Varnqor does not persist prompt text, assistant text, file contents, tool-result text, Bash commands, Grep patterns, or Write/Edit payloads in the tool index.

The fingerprint exists only to compare whether two local operations had the same canonical input. It is not encryption and is not uploaded by default.

## Codex and Gemini

Codex and Gemini 1.0 adapters index usage/resource metadata from supported local session formats. They do not contribute tool-level Waste findings in 1.0.

Codex compressed `.jsonl.zst` archives are detected and reported as unsupported rather than silently decoded or ignored.

## Local storage

Historical metadata is stored in Varnqor's local SQLite database under the user's application configuration directory. Settings are also local.

The 1.0 upgrade performs a one-time data-hygiene migration that removes legacy zero-token usage rows and normalizes internal placeholder model labels. It does not delete valid token history or Claude tool metadata.

## Future online features

Any future feature that sends Varnqor data off-device must be explicit, opt-in, and documented separately. Local analysis should remain useful without an online service.
