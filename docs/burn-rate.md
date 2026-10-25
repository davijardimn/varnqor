# Burn Rate

Burn Rate is Varnqor's first resource-intelligence signal.

It answers a narrow question:

> Is the latest hour of activity unusually expensive compared with this user's active hours today?

It does **not** claim to know provider quota remaining, reset windows, or a sustainable provider-defined allowance.

## Current window

The current burn value is the total token activity observed during the rolling 60 minutes ending at the time Varnqor refreshes.

```text
current = usage(now - 60m, now)
```

The previous window is the 60 minutes immediately before that:

```text
previous = usage(now - 120m, now - 60m)
```

These are exact rolling windows queried from the local SQLite usage index, not approximations from chart buckets.

## Active baseline

The baseline uses completed hourly buckets from the current local day.

Hours with zero usage are excluded. This prevents sleep, lunch, or other idle periods from making normal work appear artificially expensive.

Varnqor then takes the median of those active hourly totals:

```text
baseline = median(completed non-zero hourly totals)
```

The median is deliberately used instead of the mean so a single unusually expensive hour has less influence on the baseline.

At least two completed active hours are required before Varnqor labels the burn ratio. Until then the signal remains `learning`.

## Ratio

```text
burn ratio = current rolling 60m / active-hour baseline
```

Current v0.1 bands:

| Ratio | Level |
| --- | --- |
| `< 0.75x` | low |
| `0.75x - <1.20x` | normal |
| `1.20x - <1.50x` | elevated |
| `1.50x - <2.00x` | high |
| `>= 2.00x` | critical |

These bands are Varnqor heuristics, not Claude limits.

## Trend

The trend compares the two rolling windows:

```text
change = (current - previous) / previous
```

- more than `+10%` → `accelerating`
- below `-10%` → `cooling`
- otherwise → `steady`
- no previous activity with current activity → `new-spike`
- no activity in either window → `idle`

## Confidence

Confidence refers only to how much local baseline evidence exists:

| Active baseline hours | Confidence |
| --- | --- |
| `< 2` | insufficient |
| `2` | low |
| `3-5` | medium |
| `6+` | high |

It is not a probability and does not measure provider certainty.

## What Burn Rate is not

Burn Rate is not:

- quota remaining;
- cost forecasting;
- a provider-defined throttle threshold;
- a claim that high activity is waste;
- a recommendation to bypass provider limits.

Future milestones can combine Burn Rate with official or explicitly estimated quota information, but the source and confidence of that quota must remain visible.
