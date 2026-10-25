# Privacy

Varnqor is local-first.

The default product direction is:

- no cloud account required;
- no telemetry by default;
- no prompt upload;
- no source-code upload;
- no hidden background data export;
- read-only access to supported agent data unless a future feature clearly asks for permission.

## Data minimisation

Collectors prefer metadata required for resource analysis over conversation content.

The local usage index stores token/accounting metadata. The tool index stores only fields needed for explainable rules: tool name, timestamps, project/session identifiers, file or search-scope path when available, Read range, result byte length, error state, and a SHA-256 fingerprint of canonical tool input.

Raw tool input is discarded after fingerprinting. Varnqor does not persist prompt text, assistant text, file contents, tool-result text, Bash commands, Grep patterns, or Write/Edit payloads.

The fingerprint is used only to compare whether two local operations had the same canonical input. It is not encryption and is not uploaded by default.

## Future online features

If Varnqor later adds optional online features, they must be opt-in and documented separately. Local analysis should remain useful without them.
