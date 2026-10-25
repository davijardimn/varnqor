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
    date: '',
    calls: 0,
    sessions: 0,
    projects: 0,
    filesScanned: 0,
    parseWarnings: 0,
    cacheReadShare: 0,
    tokens: emptyTokens,
    models: [],
  },
  hourly: [],
  sync: {
    filesSeen: 0,
    filesChanged: 0,
    bytesIndexed: 0,
    recordsObserved: 0,
    parseWarnings: 0,
    rebuilt: false,
  },
  storageReady: false,
}

const compactNumber = new Intl.NumberFormat(undefined, {
  notation: 'compact',
  maximumFractionDigits: 2,
})

function formatTokens(value: number) {
  return compactNumber.format(value)
}

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

function metric(label: string, value: number, detail: string) {
  return (
    <article className="metric" key={label}>
      <span className="label">{label}</span>
      <strong>{formatTokens(value)}</strong>
      <small>{detail}</small>
    </article>
  )
}

export default function App() {
  const [status, setStatus] = useState<VarnqorSystemStatus>(emptyStatus)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    const load = async () => {
      const getSystemStatus = window.go?.main?.App?.GetSystemStatus
      if (getSystemStatus) {
        try {
          setStatus(await getSystemStatus())
        } catch (error) {
          console.error('Unable to read Varnqor system status', error)
        }
      }
      setLoading(false)
    }
    void load()
  }, [])

  const claude = status.claude
  const usage = status.usage
  const hourly = status.hourly ?? []
  const models = usage.models ?? []
  const hasUsage = usage.calls > 0
  const maxHourly = Math.max(1, ...hourly.map((bucket) => bucket.tokens.total))

  const headline = loading
    ? 'Indexing local usage…'
    : status.usageError
      ? 'Local index needs attention.'
      : hasUsage
        ? 'Usage history is live.'
        : claude.detected
          ? 'Claude Code detected.'
          : 'Claude Code not detected.'

  return (
    <main className="shell">
      <header className="topbar">
        <div>
          <div className="wordmark">VARNQOR</div>
          <div className="tagline">Resource intelligence for AI coding agents.</div>
        </div>
        <div className="milestone">M2 · LOCAL STORE</div>
      </header>

      <section className="hero">
        <div className="eyebrow">CLAUDE CODE · TODAY</div>
        <h1>{headline}</h1>
        <p>
          Incremental local telemetry backed by SQLite. Varnqor only re-reads transcript bytes that
          have not already been indexed, while preserving deduplication across sessions and files.
        </p>
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

      <section className="chartPanel" aria-label="Hourly token activity">
        <div className="panelHeader">
          <span className="label">TOKEN ACTIVITY · HOURLY</span>
          <span>{formatTokens(usage.tokens.total)} today</span>
        </div>
        <div className="bars">
          {hourly.map((bucket, index) => {
            const height = bucket.tokens.total === 0 ? 2 : Math.max(5, (bucket.tokens.total / maxHourly) * 100)
            return (
              <div className="barColumn" key={bucket.start} title={`${hourLabel(bucket.start)} · ${formatTokens(bucket.tokens.total)} tokens`}>
                <div className="barTrack"><div className="barFill" style={{ height: `${height}%` }} /></div>
                <span>{index % 3 === 0 ? hourLabel(bucket.start) : ''}</span>
              </div>
            )
          })}
        </div>
      </section>

      <section className="lowerGrid">
        <div className="panel">
          <div className="panelHeader"><span className="label">MODELS OBSERVED</span><span>{models.length}</span></div>
          {models.length === 0 ? <p className="empty">No token-bearing Claude calls found for today.</p> : (
            <div className="models">
              {models.slice(0, 4).map((model) => (
                <div className="modelRow" key={model.model}>
                  <div><strong>{model.model}</strong><small>{model.calls} calls</small></div>
                  <span>{formatTokens(model.tokens.total)}</span>
                </div>
              ))}
            </div>
          )}
        </div>

        <div className="panel signalPanel">
          <span className="label">LOCAL INDEX</span>
          <strong>{status.storageReady ? 'SQLITE READY' : 'UNAVAILABLE'}</strong>
          <p>
            {usage.filesScanned} sources · {status.sync.filesChanged} changed · {formatTokens(status.sync.bytesIndexed)} bytes indexed
          </p>
          <p>{formatDate(usage.lastActivity)} · {usage.parseWarnings} parse warnings</p>
          <div className="privacy">{status.sync.rebuilt ? 'INDEX REBUILT · VERIFIED' : 'LOCAL-FIRST · READ-ONLY'}</div>
        </div>
      </section>

      <footer>Control the burn. <span>v0.1 pre-alpha</span></footer>
    </main>
  )
}
