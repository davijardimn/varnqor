# Varnqor CLI

The desktop application and CLI share the same local Go core, settings and SQLite index.

## Commands

```text
varnqor                         open the desktop application
varnqor status [--json]         concise local state
varnqor analyze [--json]        Burn + Attribution + Waste + Advisor
varnqor forecast [--json]       projection + trusted Runway
varnqor doctor [--json]         local environment diagnostics
varnqor config show [--json]    local settings
varnqor config refresh 5        set live refresh interval
varnqor config budget 10m       set a configured daily token budget
varnqor config budget off       disable the configured budget
varnqor version                 application version
varnqor help                    command reference
```

Budget amounts accept plain values or decimal suffixes: `500k`, `10m`, `1.5g`.

An optional reset hour can be supplied in local time:

```bash
varnqor config budget 10m --reset-hour 6
```

The configured budget is explicitly a user budget, not an official provider quota.

## Examples

```bash
varnqor status
varnqor analyze --json
varnqor forecast
varnqor doctor --json
varnqor config show --json
```

## Exit behavior

- `0` — command completed without a Varnqor storage/query failure;
- `1` — diagnostics found a failure, settings could not be saved, or status collection failed;
- `2` — invalid/unknown command or arguments.

Claude Code not being installed is a Doctor warning rather than a fatal application error. This lets the executable be diagnosed before the first Claude session exists.

## Privacy

JSON output can contain local filesystem paths in diagnostics and identifiers from local metadata. Review output before sharing it publicly.

The CLI does not upload the result anywhere.
