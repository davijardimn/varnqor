import { useEffect, useState } from 'react'

const emptyTokens: TokenTotals = {
  input: 0,
  output: 0,
  cacheCreation: 0,
  cacheRead: 0,
  cacheCreation5m: 0,
  cacheCreation1h: 0,
  total: 0,
}

const emptyStatus: VarnqorSystemStatus = {
  claude: {
    detected: false,
    binaryDetected: false,
    dataDetected: false,
    claudeHome: '',
    projectsPath: '',
    projects: 0,
    sessions: 0,
  },
  usage: {
    date: '', calls: 0, sessions: 0, projects: 0, filesScanned: 0, parseWarnings: 0,
    cacheReadShare: 0, tokens: emptyTokens, models: [],
  },
  hourly: [],
  burn: {
    ready: false, currentPerHour: 0, previousPerHour: 0, baselinePerHour: 0, ratio: 0,
    changePercent: 0, trend: 'idle', level: 'learning', confidence: 'insufficient', baselineSamples: 0,
  },
  attribution: {
    calls: 0, tokens: emptyTokens, projects: [], sessions: [], models: [], agents: [],
  },
  sync: {
    filesSeen: 0, filesChanged: 0, bytesIndexed: 0, recordsObserved: 0, parseWarnings: 0, rebuilt: false,
  },
  storageReady: false,
}

const compactNumber = new Intl.NumberFormat(undefined, { notation: 'compact', maximumFractionDigits: 2 })

function formatTokens(value: number) { return compactNumber.format(value) }
function formatDate(value?: string) {
  if (!value) return 'No activity today'
  const date = new Date(value)
  if (Number.isNaN(date.getTime()) || date.getUTCFullYear() <= 1) return 'No activity today'
  return date.toLocaleString()
}
function hourLabel(value: string) {
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? '--' : date.toLocaleTimeString([], { hour: '2-digit' })
}
function signedPercent(value: number) {
  if (value === 0) return '0.0%'
  return `${value > 0 ? '+' : ''}${value.toFixed(1)}%`
}
function trendLabel(trend: BurnRate['trend']) {
  switch (trend) {
    case 'accelerating': return 'ACCELERATING'
    case 'cooling': return 'COOLING'
    case 'steady': return 'STEADY'
    case 'new-spike': return 'NEW ACTIVITY'
    default: return 'IDLE'
  }
}
function shortSession(value?: string) {
  if (!value) return 'unknown'
  return value.length > 12 ? `${value.slice(0, 8)}…` : value
}
function metric(label: string, value: number, detail: string) {
  return <article className="metric" key={label}><span className="label">{label}</span><strong>{formatTokens(value)}</strong><small>{detail}</small></article>
}

function DriverList({ title, items, kind }: { title: string; items: AttributionItem[]; kind: 'project' | 'session' | 'model' | 'agent' }) {
  return (
    <div className="driverGroup">
      <div className="driverTitle"><span className="label">{title}</span><span>{items.length}</span></div>
      {items.length === 0 ? <p className="empty">No activity in this window.</p> : (
        <div className="driverRows">
          {items.slice(0, kind === 'agent' ? 2 : 4).map((item) => {
            const label = kind === 'session' ? shortSession(item.session || item.label) : item.label
            return (
              <div className="driverRow" key={item.key} title={kind === 'session' ? `${item.project} · ${item.session}` : item.label}>
                <div className="driverCopy">
                  <strong>{label}</strong>
                  <small>{kind === 'session' ? `${item.project} · ` : ''}{item.calls} calls</small>
                </div>
                <div className="driverNumbers"><strong>{formatTokens(item.tokens.total)}</strong><span>{item.sharePercent.toFixed(1)}%</span></div>
                <div className="driverTrack"><div style={{ width: `${Math.min(100, item.sharePercent)}%` }} /></div>
              </div>
            )
          })}
        </div>
      )}
    </div>
  )
}

export default function App() {
  const [status, setStatus] = useState<VarnqorSystemStatus>(emptyStatus)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    const load = async () => {
      const getSystemStatus = window.go?.main?.App?.GetSystemStatus
      if (getSystemStatus) {
        try { setStatus(await getSystemStatus()) } catch (error) { console.error('Unable to read Varnqor system status', error) }
      }
      setLoading(false)
    }
    void load()
  }, [])

  const claude = status.claude
  const usage = status.usage
  const hourly = status.hourly ?? []
  const burn = status.burn
  const attribution = status.attribution
  const projects = attribution.projects ?? []
  const sessions = attribution.sessions ?? []
  const burnModels = attribution.models ?? []
  const agents = attribution.agents ?? []
  const models = usage.models ?? []
  const hasUsage = usage.calls > 0
  const maxHourly = Math.max(1, ...hourly.map((bucket) => bucket.tokens.total))
  const meterWidth = burn.ready ? Math.min(100, Math.max(2, (burn.ratio / 2.5) * 100)) : 0

  const headline = loading ? 'Indexing local usage…' : status.usageError ? 'Local index needs attention.' : burn.ready ? 'Burn drivers are visible.' : hasUsage ? 'Learning your usage baseline.' : claude.detected ? 'Claude Code detected.' : 'Claude Code not detected.'

  return (
    <main className="shell">
      <header className="topbar">
        <div><div className="wordmark">VARNQOR</div><div className="tagline">Resource intelligence for AI coding agents.</div></div>
        <div className="milestone">M4 · ATTRIBUTION</div>
      </header>

      <section className="hero">
        <div className="eyebrow">CLAUDE CODE · TODAY</div>
        <h1>{headline}</h1>
        <p>Varnqor now connects the burn signal to the projects, sessions, models and agent activity that produced it. Attribution uses the exact same rolling 60-minute window as Current Burn Rate.</p>
        {status.usageError && <div className="warning">{status.usageError}</div>}
      </section>

      <section className="grid" aria-label="Token usage today">
        {metric('INPUT', usage.tokens.input, 'Uncached input tokens')}
        {metric('OUTPUT', usage.tokens.output, 'Generated output tokens')}
        {metric('CACHE READ', usage.tokens.cacheRead, 'Input served from cache')}
        {metric('CACHE WRITE', usage.tokens.cacheCreation, 'Input written to cache')}
      </section>

      <section className="summary" aria-label="Usage summary">
        <div className="summaryItem"><span className="label">TOTAL ACTIVITY</span><strong>{formatTokens(usage.tokens.total)}</strong></div>
        <div className="summaryItem"><span className="label">API CALLS</span><strong>{usage.calls}</strong></div>
        <div className="summaryItem"><span className="label">ACTIVE SESSIONS</span><strong>{usage.sessions}</strong></div>
        <div className="summaryItem"><span className="label">CACHE READ SHARE</span><strong>{usage.cacheReadShare.toFixed(1)}%</strong></div>
      </section>

      <section className={`burnPanel burn-${burn.level}`} aria-label="Burn rate">
        <div className="burnPrimary">
          <span className="label">CURRENT BURN RATE</span>
          <div className="burnHeadline"><strong>{burn.ready ? `${burn.ratio.toFixed(2)}×` : '—'}</strong><span>{burn.ready ? burn.level.toUpperCase() : 'LEARNING'}</span></div>
          <div className="burnMeter" aria-hidden="true"><div className="burnMeterFill" style={{ width: `${meterWidth}%` }} /><i className="burnBaselineMarker" /></div>
          <p>{burn.ready ? `Current activity is ${Math.abs((burn.ratio - 1) * 100).toFixed(0)}% ${burn.ratio >= 1 ? 'above' : 'below'} the active-hour baseline.` : `Need at least 2 completed active hours. ${burn.baselineSamples} available so far.`}</p>
        </div>
        <div className="burnFacts">
          <div><span className="label">LAST 60 MIN</span><strong>{formatTokens(burn.currentPerHour)}</strong></div>
          <div><span className="label">PREVIOUS 60 MIN</span><strong>{formatTokens(burn.previousPerHour)}</strong></div>
          <div><span className="label">ACTIVE BASELINE</span><strong>{burn.baselinePerHour ? formatTokens(burn.baselinePerHour) : '—'}</strong></div>
          <div><span className="label">TREND</span><strong>{trendLabel(burn.trend)}</strong><small>{burn.previousPerHour > 0 ? signedPercent(burn.changePercent) : 'No prior comparison'}</small></div>
          <div><span className="label">CONFIDENCE</span><strong>{burn.confidence.toUpperCase()}</strong><small>{burn.baselineSamples} baseline samples</small></div>
        </div>
      </section>

      <section className="attributionPanel" aria-label="Burn attribution">
        <div className="attributionHeader">
          <div><span className="label">BURN DRIVERS · LAST 60 MIN</span><h2>Where the current activity came from.</h2></div>
          <div className="attributionTotal"><span>{attribution.calls} calls</span><strong>{formatTokens(attribution.tokens.total)}</strong></div>
        </div>
        <div className="driverGrid">
          <DriverList title="PROJECTS" items={projects} kind="project" />
          <DriverList title="SESSIONS" items={sessions} kind="session" />
          <DriverList title="MODELS" items={burnModels} kind="model" />
          <DriverList title="AGENT SPLIT" items={agents} kind="agent" />
        </div>
      </section>

      <section className="chartPanel" aria-label="Hourly token activity">
        <div className="panelHeader"><span className="label">TOKEN ACTIVITY · HOURLY</span><span>{formatTokens(usage.tokens.total)} today</span></div>
        <div className="bars">
          {hourly.map((bucket, index) => {
            const height = bucket.tokens.total === 0 ? 2 : Math.max(5, (bucket.tokens.total / maxHourly) * 100)
            return <div className="barColumn" key={bucket.start} title={`${hourLabel(bucket.start)} · ${formatTokens(bucket.tokens.total)} tokens`}><div className="barTrack"><div className="barFill" style={{ height: `${height}%` }} /></div><span>{index % 3 === 0 ? hourLabel(bucket.start) : ''}</span></div>
          })}
        </div>
      </section>

      <section className="lowerGrid">
        <div className="panel">
          <div className="panelHeader"><span className="label">MODELS OBSERVED · TODAY</span><span>{models.length}</span></div>
          {models.length === 0 ? <p className="empty">No token-bearing Claude calls found for today.</p> : <div className="models">{models.slice(0, 4).map((model) => <div className="modelRow" key={model.model}><div><strong>{model.model}</strong><small>{model.calls} calls</small></div><span>{formatTokens(model.tokens.total)}</span></div>)}</div>}
        </div>
        <div className="panel signalPanel">
          <span className="label">LOCAL INDEX</span><strong>{status.storageReady ? 'SQLITE READY' : 'UNAVAILABLE'}</strong>
          <p>{usage.filesScanned} sources · {status.sync.filesChanged} changed · {formatTokens(status.sync.bytesIndexed)} bytes indexed</p>
          <p>{formatDate(usage.lastActivity)} · {usage.parseWarnings} parse warnings</p>
          <div className="privacy">{status.sync.rebuilt ? 'INDEX REBUILT · VERIFIED' : 'LOCAL-FIRST · READ-ONLY'}</div>
        </div>
      </section>

      <footer>Control the burn. <span>v0.1 pre-alpha</span></footer>
    </main>
  )
}
