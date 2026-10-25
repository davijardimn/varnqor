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

Collectors should prefer metadata required for resource analysis over conversation content.

The usage index stores token counts, timestamps, provider/model/session identifiers and other accounting metadata. It does not store prompt or assistant-response text.

Waste Analyzer v0.1 additionally indexes privacy-minimised tool metadata. For file tools, Varnqor may store the tool name, file path, Read offset/limit, timestamps, session identifiers, error state and the byte length of a tool result. It does **not** store the tool-result text, file contents, Bash command text, or Write/Edit payloads.

All current analysis runs locally against SQLite.

## Local paths

File paths are metadata and can still reveal local project structure. They stay in the local Varnqor database and are not uploaded by the current application.

## Future online features

If Varnqor later adds optional online features, they must be opt-in and documented separately. Local analysis should remain useful without them.
