# Trusted Limits and Runway

Varnqor does not convert a usage baseline into a provider quota.

## Limit sources

The data model is designed to distinguish limit provenance. The first implemented source is:

```text
configured
```

A configured limit is a budget entered by the user. It is **not** an official Claude limit and the UI says so explicitly.

Future adapters may add sources such as:

```text
official
observed
```

An estimated provider limit, if ever supported, must remain visibly labelled `estimated` and must never be presented as official.

## Current configured budget

The first policy is a local daily token-activity budget with a user-selected local reset hour.

For a limit `L`, usage `U`, and rolling current pace `P` tokens/hour:

```text
remaining = max(L - U, 0)
runway_hours = remaining / P
```

Runway is unavailable while `P == 0` because there is no active burn pace to extrapolate.

The reset projection is:

```text
projected_at_reset = used + current_pace * hours_until_reset
```

`willExceedBeforeReset` is only a persistence projection. It does not predict future user behavior.

## Storage

Settings are stored locally in the operating system user config directory under `varnqor/settings.json`. The file contains monitoring preferences and configured budget values only. No provider credentials are required for configured budgets.
