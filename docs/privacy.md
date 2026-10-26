# Privacy

Varnqor is local-first by design.

## Default boundary

Varnqor does not require a Varnqor cloud account or provider API key for local analysis. Its default product boundary is:

- no prompt upload;
- no source-code body upload;
- no assistant-response upload;
- no raw tool-result body persistence;
- no hidden background data export;
- no product telemetry required for local analysis;
- read-only access to supported agent session data;
- no external agent action executed by Optimize automatically.

Explorer, Session Lab, Pulse, Advisor, and Optimize operate on local metadata. Themes, translations, and provider identification marks are also bundled locally and do not require a runtime CDN.

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

## Optimize

Varnqor 3's Optimize engine reads the same local usage and privacy-minimized Waste metadata used by the rest of the product. It computes target scenarios and playbooks locally.

Optimize does not require prompt bodies or source-code contents. It does not silently:

- rewrite a prompt;
- switch a provider or model;
- stop or launch an agent;
- execute a tool;
- mutate a repository;
- change provider credentials or limits.

`preview` and `manual` describe how the plan is presented to the user. They are not remote execution permissions.

Estimated impact ranges are heuristic and may overlap. They are not transmitted to a remote optimization service and are not represented as guaranteed savings.

## Codex and Gemini

Codex and Gemini adapters index usage/resource metadata from supported local session formats. They do not currently contribute tool-level Waste findings.

Codex compressed `.jsonl.zst` archives are detected and reported as unsupported rather than silently decoded or ignored.

## Local storage and preferences

Historical metadata is stored in Varnqor's local SQLite database under the user's application configuration directory. Settings are also local and include:

- refresh interval;
- configured budget;
- Pulse guardrails;
- language;
- appearance preference;
- Optimize strategy and target scenario.

Settings migrations preserve supported previous-version preferences while filling new fields with documented defaults.

## Future online features

Any future feature that sends Varnqor data off-device must be explicit, opt-in, and documented separately. Local analysis should remain useful without an online service.
