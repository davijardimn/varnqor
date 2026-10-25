const compact = new Intl.NumberFormat(undefined, { notation: 'compact', maximumFractionDigits: 2 })

function tokens(value: number) { return compact.format(value) }
function horizonLabel(minutes: number) { return minutes % 60 === 0 ? `${minutes / 60}h` : `${minutes}m` }

export default function ForecastPanel({ forecast }: { forecast: ForecastReport }) {
  const points = forecast.points ?? []
  return <section className={`forecastPanel pressure-${forecast.pressure}`} aria-label="Resource forecast">
    <div className="forecastHeader">
      <div>
        <span className="label">M6 · RESOURCE PRESSURE</span>
        <h2>What happens if this pace continues?</h2>
        <p>Persistence forecast from the latest rolling 60-minute pace. The envelope is observed pace range, not a statistical confidence interval.</p>
      </div>
      <div className="pressureBadge">
        <span>PRESSURE</span>
        <strong>{forecast.pressure.toUpperCase()}</strong>
        <small>{forecast.confidence.toUpperCase()} confidence</small>
      </div>
    </div>

    <div className="forecastGrid">
      <div className="forecastHero">
        <span className="label">PROJECTED · {horizonLabel(forecast.horizonMinutes)}</span>
        <strong>{forecast.ready ? tokens(forecast.projectedTokens) : '—'}</strong>
        <p>{forecast.baselineProjection > 0 ? `${tokens(forecast.baselineProjection)} at baseline pace` : 'Baseline still learning'}</p>
      </div>
      <div className="forecastFact"><span className="label">CURRENT PACE</span><strong>{tokens(forecast.currentPerHour)}/h</strong><small>{forecast.pressureRatio > 0 ? `${forecast.pressureRatio.toFixed(2)}× baseline` : 'ratio unavailable'}</small></div>
      <div className="forecastFact"><span className="label">PACE ENVELOPE</span><strong>{tokens(forecast.envelopeLow)}–{tokens(forecast.envelopeHigh)}</strong><small>projected {horizonLabel(forecast.horizonMinutes)}</small></div>
      <div className="forecastFact"><span className="label">RUNWAY</span><strong>{forecast.runwayAvailable ? 'AVAILABLE' : 'NO TRUSTED LIMIT'}</strong><small>{forecast.runwayAvailable ? forecast.limitSource : 'quota-aware runway stays off'}</small></div>
    </div>

    <div className="forecastPoints">
      {points.map((point) => <div key={point.minutes}>
        <span>{horizonLabel(point.minutes)}</span>
        <strong>{tokens(point.projectedTokens)}</strong>
        <small>{tokens(point.envelopeLow)}–{tokens(point.envelopeHigh)} envelope</small>
      </div>)}
    </div>
  </section>
}
