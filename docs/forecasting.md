# Forecasting and Resource Pressure

Varnqor forecasting is deliberately simple and inspectable.

## Persistence forecast

The current implementation assumes the latest rolling 60-minute activity rate continues unchanged over the forecast horizon.

```text
projected activity = current 60m rate × horizon
```

The desktop currently uses a two-hour horizon and exposes 30-minute, 60-minute and two-hour checkpoints.

This is a **persistence forecast**, not a claim that future model usage is known.

## Pace envelope

Varnqor also shows an observed pace envelope built from:

- current rolling 60-minute activity;
- previous rolling 60-minute activity, when available;
- active-hour median baseline, when available.

The lowest and highest of those observed rates are projected over the same horizon.

This envelope is **not a statistical confidence interval**. It only answers: what would the future volume look like if one of the recently observed paces persisted?

## Resource pressure

When Burn Rate has enough baseline history, Resource Pressure uses the same current-to-baseline ratio and classifications:

```text
low       < 0.75x
normal    < 1.20x
elevated  < 1.50x
high      < 2.00x
critical  >= 2.00x
```

If the baseline is still learning, the projection can still be shown from observed current pace, but pressure remains `learning`.

## Confidence

Forecast confidence inherits the Burn Rate baseline confidence. If a current pace exists before a reliable baseline does, the forecast is marked `low` confidence.

## Runway and quota honesty

Forecasting resource activity does not imply Varnqor knows the provider's quota.

Runway remains unavailable until Varnqor has a trusted limit source such as:

```text
official
observed
configured
```

The current forecast never converts a local usage baseline into a fake quota. Until a trusted limit exists, the UI says `NO TRUSTED LIMIT`.
