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
  const hasUsage = usage.calls > 0
  const models = usage.models ?? []

  const headline = loading
    ? 'Scanning local usage…'
    : status.usageError
      ? 'Usage scan needs attention.'
      : hasUsage
        ? 'Usage signal acquired.'
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
        <div className="milestone">M1 · USAGE ENGINE</div>
      </header>

      <section className="hero">
        <div className="eyebrow">CLAUDE CODE · TODAY</div>
        <h1>{headline}</h1>
        <p>
          Read-only local telemetry, deduplicated by API call. Prompts, responses and source code
          are not retained by the usage engine.
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
        <div className="summaryItem">
          <span className="label">TOTAL ACTIVITY</span>
          <strong>{formatTokens(usage.tokens.total)}</strong>
        </div>
        <div className="summaryItem">
          <span className="label">API CALLS</span>
          <strong>{usage.calls}</strong>
        </div>
        <div className="summaryItem">
          <span className="label">ACTIVE SESSIONS</span>
          <strong>{usage.sessions}</strong>
        </div>
        <div className="summaryItem">
          <span className="label">CACHE READ SHARE</span>
          <strong>{usage.cacheReadShare.toFixed(1)}%</strong>
        </div>
      </section>

      <section className="lowerGrid">
        <div className="panel">
          <div className="panelHeader">
            <span className="label">MODELS OBSERVED</span>
            <span>{models.length}</span>
          </div>
          {models.length === 0 ? (
            <p className="empty">No token-bearing Claude calls found for today.</p>
          ) : (
            <div className="models">
              {models.slice(0, 4).map((model) => (
                <div className="modelRow" key={model.model}>
                  <div>
                    <strong>{model.model}</strong>
                    <small>{model.calls} calls</small>
                  </div>
                  <span>{formatTokens(model.tokens.total)}</span>
                </div>
              ))}
            </div>
          )}
        </div>

        <div className="panel signalPanel">
          <span className="label">LAST TOKEN ACTIVITY</span>
          <strong>{formatDate(usage.lastActivity)}</strong>
          <p>
            {usage.filesScanned} JSONL files scanned · {usage.parseWarnings} parse warnings
          </p>
          <div className="privacy">LOCAL-FIRST · READ-ONLY</div>
        </div>
      </section>

      <footer>
        Control the burn. <span>v0.1 pre-alpha</span>
      </footer>
    </main>
  )
}
