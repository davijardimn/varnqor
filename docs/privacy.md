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

For example, Milestone 0 only checks the Claude Code data directory and session transcript files. It does not parse or transmit prompt text.

## Future online features

If Varnqor later adds optional online features, they must be opt-in and documented separately. Local analysis should remain useful without them.
