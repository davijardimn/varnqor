import { useEffect, useState } from 'react'

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
}

function formatLastActivity(value?: string) {
  if (!value) return 'No local activity found'

  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return 'Unknown'

  return date.toLocaleString()
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

  return (
    <main className="shell">
      <header className="topbar">
        <div>
          <div className="wordmark">VARNQOR</div>
          <div className="tagline">Resource intelligence for AI coding agents.</div>
        </div>
        <div className="milestone">M0 · FIRST SIGNAL</div>
      </header>

      <section className="hero">
        <div className="eyebrow">LOCAL SIGNAL</div>
        <h1>{loading ? 'Scanning this machine…' : claude.detected ? 'Claude Code detected.' : 'Claude Code not detected.'}</h1>
        <p>
          Varnqor is starting with one promise: turn local agent activity into reliable,
          explainable resource intelligence.
        </p>
      </section>

      <section className="grid" aria-label="Claude Code signal">
        <article className="metric">
          <span className="label">STATUS</span>
          <strong className={claude.detected ? 'ok' : 'muted'}>
            {claude.detected ? 'CONNECTED' : 'NO SIGNAL'}
          </strong>
          <small>Binary or local Claude data</small>
        </article>

        <article className="metric">
          <span className="label">PROJECTS</span>
          <strong>{claude.projects}</strong>
          <small>Local project stores discovered</small>
        </article>

        <article className="metric">
          <span className="label">SESSIONS</span>
          <strong>{claude.sessions}</strong>
          <small>Main JSONL transcripts found</small>
        </article>
      </section>

      <section className="signal">
        <div>
          <span className="label">LAST LOCAL ACTIVITY</span>
          <p>{formatLastActivity(claude.lastActivity)}</p>
        </div>
        <div className="privacy">LOCAL-FIRST · READ-ONLY</div>
      </section>

      <footer>
        Control the burn. <span>v0.1 pre-alpha</span>
      </footer>
    </main>
  )
}
