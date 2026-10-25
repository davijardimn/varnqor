# Live Monitor

Varnqor Desktop now refreshes automatically while it is open.

## How it works

The frontend requests a fresh system status at the configured interval (default: 5 seconds). Each refresh runs the existing incremental Claude indexer.

This does **not** re-read every transcript. Unchanged sources keep their saved offsets and only new complete JSONL data is processed.

The UI shows:

```text
● LIVE · 5S
```

and temporarily switches to `SYNC` while a refresh is in flight.

## Refresh controls

The refresh interval is configurable between 2 and 60 seconds in Settings. Varnqor allows only one in-flight refresh from the desktop polling loop, preventing overlapping local sync requests.

## Background behavior

This milestone is live **while the desktop application is open**. It is not yet an always-running operating-system service and it does not silently install a background daemon.

System-tray/background monitoring can be added later as an explicit product capability.
