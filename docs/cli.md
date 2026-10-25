# Varnqor CLI

The desktop application and CLI share the same local Go core and SQLite index.

## Commands

```text
varnqor                     open the desktop application
varnqor status              concise local state
varnqor analyze             Burn + Attribution + Waste + Advisor
varnqor forecast            projection + trusted Runway
varnqor doctor              local environment diagnostics
varnqor version             application version
varnqor help                command reference
```

`status`, `analyze`, `forecast`, and `doctor` support `--json` for scripts.

## Examples

```bash
varnqor status
varnqor analyze --json
varnqor forecast
varnqor doctor --json
```

## Exit behavior

- `0` — command completed without a Varnqor storage/query failure;
- `1` — diagnostics found a failure or status collection failed;
- `2` — unknown command.

Claude Code not being installed is a Doctor warning rather than a fatal application error. This lets the executable be diagnosed before the first Claude session exists.

## Privacy

JSON output can contain local filesystem paths in diagnostics and identifiers from local metadata. Review output before sharing it publicly.

The CLI does not upload the result anywhere.
