const compact = new Intl.NumberFormat(undefined, { notation: 'compact', maximumFractionDigits: 2 })

export default function AdvisorPanel({ advisor }: { advisor: AdvisorReport }) {
  const recommendations = advisor.recommendations ?? []
  return <section className="advisorPanel" aria-label="Varnqor Advisor">
    <div className="advisorHeader">
      <div>
        <span className="label">M7 · VARNQOR ADVISOR</span>
        <h2>Evidence into next actions.</h2>
        <p>Read-only recommendations from pressure, attribution and waste signals. Nothing is changed automatically.</p>
      </div>
      <div className="advisorCounts"><span>{advisor.high} HIGH</span><span>{advisor.medium} MED</span><span>{advisor.low} LOW</span></div>
    </div>

    {recommendations.length === 0 ? <div className="advisorEmpty">No recommendation is currently strong enough to surface. Advisor remains passive.</div> : <div className="advisorRows">
      {recommendations.map((item, index) => {
        const estimatedTokens = item.estimatedTokens ?? 0
        return <article className={`advisorRow priority-${item.priority}`} key={item.id}>
          <div className="advisorNumber">{String(index + 1).padStart(2, '0')}</div>
          <div className="advisorCopy">
            <div className="advisorTitle"><span>{item.priority.toUpperCase()}</span><strong>{item.title}</strong></div>
            <p>{item.evidence}</p>
            <small>{item.recommendation}</small>
          </div>
          <div className="advisorEstimate">{estimatedTokens > 0 ? <><strong>~{compact.format(estimatedTokens)}</strong><span>signal context</span></> : <span>{item.category.toUpperCase()}</span>}</div>
        </article>
      })}
    </div>}
    <div className="advisorBoundary">ADVISORY ONLY · NO PROMPT / MODEL / AGENT / FILE CHANGES</div>
  </section>
}
