import { useEffect, useMemo, useState } from 'react'

type ViewName = 'overview' | 'burn' | 'advisor' | 'activity' | 'projects' | 'models' | 'settings'
type IconName = ViewName | 'pulse' | 'clock' | 'spark' | 'check' | 'warning' | 'arrow' | 'shield' | 'database' | 'terminal'
type Tone = 'neutral' | 'good' | 'warn' | 'danger' | 'info'

const compact = new Intl.NumberFormat(undefined, { notation: 'compact', maximumFractionDigits: 2 })
const number = new Intl.NumberFormat()

const emptyTokens: TokenTotals = { input: 0, output: 0, cacheCreation: 0, cacheRead: 0, cacheCreation5m: 0, cacheCreation1h: 0, total: 0 }
const emptyQuota: QuotaStatus = { available: false, source: '', metric: 'tokens', label: '', limit: 0, used: 0, remaining: 0, utilizationPercent: 0, currentPerHour: 0, runwayAvailable: false, runwayHours: 0, projectedAtReset: 0, willExceedBeforeReset: false, confidence: '' }
const defaultSettings: VarnqorSettings = { version: 1, refreshSeconds: 5, budget: { enabled: false, label: 'Configured daily budget', limitTokens: 0, resetHour: 0 } }
const emptyStatus: VarnqorSystemStatus = {
  version: '0.1.0-alpha.1',
  claude: { detected: false, binaryDetected: false, dataDetected: false, claudeHome: '', projectsPath: '', projects: 0, sessions: 0 },
  usage: { date: '', calls: 0, sessions: 0, projects: 0, filesScanned: 0, parseWarnings: 0, cacheReadShare: 0, tokens: emptyTokens, models: [] },
  hourly: [],
  burn: { ready: false, currentPerHour: 0, previousPerHour: 0, baselinePerHour: 0, ratio: 0, changePercent: 0, trend: 'idle', level: 'learning', confidence: 'insufficient', baselineSamples: 0 },
  forecast: { ready: false, method: 'persistence', pressure: 'learning', confidence: 'insufficient', currentPerHour: 0, previousPerHour: 0, baselinePerHour: 0, pressureRatio: 0, trend: 'idle', horizonMinutes: 120, projectedTokens: 0, baselineProjection: 0, envelopeLow: 0, envelopeHigh: 0, runwayAvailable: false, limitSource: 'unavailable', points: [] },
  quota: emptyQuota,
  attribution: { calls: 0, tokens: emptyTokens, projects: [], sessions: [], models: [], agents: [] },
  waste: { toolCalls: 0, repeatedReads: 0, repeatedSearches: 0, oversizedResponses: 0, repeatedFailures: 0, estimatedRepeatedTokens: 0, estimatedSignalTokens: 0, confidence: 'insufficient', rules: [], findings: [] },
  advisor: { mode: 'advisory', high: 0, medium: 0, low: 0, recommendations: [] },
  sync: { filesSeen: 0, filesChanged: 0, bytesIndexed: 0, recordsObserved: 0, parseWarnings: 0, rebuilt: false },
  storageReady: false,
}

const nav: Array<{ id: ViewName; label: string }> = [
  { id: 'overview', label: 'Overview' },
  { id: 'burn', label: 'Burn' },
  { id: 'advisor', label: 'Advisor' },
  { id: 'activity', label: 'Activity' },
  { id: 'projects', label: 'Projects' },
  { id: 'models', label: 'Models' },
  { id: 'settings', label: 'Settings' },
]

const pageCopy: Record<ViewName, { title: string; subtitle: string }> = {
  overview: { title: 'Overview', subtitle: 'Live resource intelligence from local agent activity.' },
  burn: { title: 'Burn', subtitle: 'Pace, pressure, projection and attribution in one place.' },
  advisor: { title: 'Advisor', subtitle: 'Prioritized actions grounded in observable evidence.' },
  activity: { title: 'Activity', subtitle: 'Token movement, cache behavior and local indexing health.' },
  projects: { title: 'Projects', subtitle: 'Understand which workspaces are driving current resource use.' },
  models: { title: 'Models', subtitle: 'Compare model activity and token composition without cloud telemetry.' },
  settings: { title: 'Settings', subtitle: 'Tune live refresh, trusted budgets and local diagnostics.' },
}

function Icon({ name, size = 18 }: { name: IconName; size?: number }) {
  const common = { width: size, height: size, viewBox: '0 0 24 24', fill: 'none', stroke: 'currentColor', strokeWidth: 1.7, strokeLinecap: 'round' as const, strokeLinejoin: 'round' as const, 'aria-hidden': true }
  switch (name) {
    case 'overview': return <svg {...common}><rect x="4" y="4" width="6" height="6" rx="1.5"/><rect x="14" y="4" width="6" height="6" rx="1.5"/><rect x="4" y="14" width="6" height="6" rx="1.5"/><rect x="14" y="14" width="6" height="6" rx="1.5"/></svg>
    case 'burn': return <svg {...common}><path d="M4 19V12M9 19V8M14 19V10M19 19V4"/></svg>
    case 'advisor': case 'spark': return <svg {...common}><path d="m12 3 1.25 3.75L17 8l-3.75 1.25L12 13l-1.25-3.75L7 8l3.75-1.25L12 3Z"/><path d="m18 14 .7 2 2 .7-2 .7-.7 2-.7-2-2-.7 2-.7.7-2Z"/></svg>
    case 'activity': case 'clock': return <svg {...common}><circle cx="12" cy="12" r="8.5"/><path d="M12 7v5l3 2"/></svg>
    case 'projects': return <svg {...common}><path d="M3.5 7.5h6l2-2h9v13h-17z"/></svg>
    case 'models': return <svg {...common}><path d="m12 3 8 4.5-8 4.5-8-4.5L12 3Z"/><path d="m4 12 8 4.5 8-4.5M4 16.5l8 4.5 8-4.5"/></svg>
    case 'settings': return <svg {...common}><circle cx="12" cy="12" r="3"/><path d="M19 13.5v-3l-2-.6-.7-1.7 1-1.8-2.1-2.1-1.8 1-.9-.3-.8-.4-.6-2h-3l-.6 2-1.7.7-1.8-1-2.1 2.1 1 1.8-.7 1.7-2 .6v3l2 .6.7 1.7-1 1.8 2.1 2.1 1.8-1 1.7.7.6 2h3l.6-2 1.7-.7 1.8 1 2.1-2.1-1-1.8.7-1.7 2-.6Z"/></svg>
    case 'pulse': return <svg {...common}><path d="M3 12h4l2-6 4 12 2-6h6"/></svg>
    case 'check': return <svg {...common}><path d="m5 12 4 4L19 6"/></svg>
    case 'warning': return <svg {...common}><path d="M12 4 3.5 19h17L12 4Z"/><path d="M12 9v4M12 16h.01"/></svg>
    case 'arrow': return <svg {...common}><path d="m9 5 7 7-7 7"/></svg>
    case 'shield': return <svg {...common}><path d="M12 3 5 6v5c0 4.7 3 8.2 7 10 4-1.8 7-5.3 7-10V6l-7-3Z"/><path d="m9 12 2 2 4-4"/></svg>
    case 'database': return <svg {...common}><ellipse cx="12" cy="5.5" rx="7.5" ry="2.5"/><path d="M4.5 5.5v6c0 1.4 3.4 2.5 7.5 2.5s7.5-1.1 7.5-2.5v-6M4.5 11.5v6c0 1.4 3.4 2.5 7.5 2.5s7.5-1.1 7.5-2.5v-6"/></svg>
    case 'terminal': return <svg {...common}><rect x="3.5" y="4.5" width="17" height="15" rx="2"/><path d="m7 9 3 3-3 3M12.5 15H17"/></svg>
  }
}

function formatTokens(value: number) { return compact.format(Math.max(0, value)) }
function formatBytes(value: number) { return `${compact.format(Math.max(0, value))}B` }
function formatTime(value?: string) {
  if (!value) return 'No activity'
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? 'No activity' : date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
}
function formatRunway(quota: QuotaStatus) {
  if (!quota.available) return 'Not set'
  if (!quota.runwayAvailable) return 'Idle'
  return quota.runwayHours >= 24 ? `${(quota.runwayHours / 24).toFixed(1)}d` : `${quota.runwayHours.toFixed(1)}h`
}
function pressureTone(value: string): Tone {
  if (value === 'critical' || value === 'high') return 'danger'
  if (value === 'elevated') return 'warn'
  if (value === 'normal' || value === 'low') return 'good'
  return 'neutral'
}
function burnMeta(burn: BurnRate) {
  if (!burn.ready) return 'Learning baseline'
  if (!burn.changePercent) return 'Stable vs previous hour'
  return `${burn.changePercent > 0 ? '+' : ''}${burn.changePercent.toFixed(0)}% vs previous hour`
}
function shortSession(value?: string) {
  if (!value) return 'Unknown'
  return value.length > 12 ? `${value.slice(0, 8)}…` : value
}
function shortPath(value?: string) {
  if (!value) return ''
  const parts = value.replace(/\\/g, '/').split('/').filter(Boolean)
  return parts.length <= 2 ? value : parts.slice(-2).join('/')
}

function Empty({ label }: { label: string }) {
  return <div className="emptyState"><span className="emptyGlyph"><Icon name="pulse" size={17}/></span><span>{label}</span></div>
}

function StatCard({ icon, label, value, meta, tone = 'neutral', onClick }: { icon: IconName; label: string; value: string; meta: string; tone?: Tone; onClick?: () => void }) {
  return <button className={`statCard tone-${tone}`} onClick={onClick}>
    <div className="statTop"><span className="statIcon"><Icon name={icon}/></span><span>{label}</span></div>
    <strong>{value}</strong>
    <small>{meta}</small>
  </button>
}

function ActivityChart({ hourly, total, large = false }: { hourly: HourUsage[]; total: number; large?: boolean }) {
  const max = Math.max(1, ...hourly.map(item => item.tokens.total))
  return <section className={`card activityCard ${large ? 'activityLarge' : ''}`}>
    <div className="cardHeader">
      <div><h2>Activity</h2><p>Hourly token movement today</p></div>
      <span className="quietPill">{formatTokens(total)} today</span>
    </div>
    <div className="chartArea">
      <div className="chartGrid"><i/><i/><i/><i/></div>
      <div className="barSeries">
        {hourly.map((item, index) => {
          const height = item.tokens.total === 0 ? 2 : Math.max(4, (item.tokens.total / max) * 100)
          return <div className="barItem" key={item.start} title={`${formatTime(item.start)} · ${number.format(item.tokens.total)} tokens`}>
            <div className="barValue" style={{ height: `${height}%` }}/>
            <span>{index % 4 === 0 ? new Date(item.start).toLocaleTimeString([], { hour: '2-digit' }) : ''}</span>
          </div>
        })}
      </div>
    </div>
  </section>
}

function SystemsCard({ status, refreshing, lastUpdated }: { status: VarnqorSystemStatus; refreshing: boolean; lastUpdated: Date | null }) {
  const rows = [
    ['Claude Code', status.claude.detected ? 'Connected' : 'Not detected', status.claude.detected],
    ['Local index', status.storageReady ? 'Ready' : 'Unavailable', status.storageReady],
    ['Sessions', number.format(status.usage.sessions), true],
    ['Projects', number.format(status.usage.projects), true],
    ['Parse health', status.usage.parseWarnings ? `${status.usage.parseWarnings} warnings` : 'Healthy', status.usage.parseWarnings === 0],
  ] as const
  return <section className="card systemsCard">
    <div className="cardHeader">
      <div className="statusTitle"><span className={`statusDot ${refreshing ? 'sync' : 'ok'}`}/><h2>{refreshing ? 'Syncing' : 'Systems'}</h2></div>
      <span className="mutedSmall">{lastUpdated ? `Updated ${lastUpdated.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}` : 'Waiting'}</span>
    </div>
    <div className="systemRows">
      {rows.map(([label, value, healthy]) => <div className="systemRow" key={label}><span>{label}</span><strong className={healthy ? '' : 'textWarn'}>{value}</strong></div>)}
    </div>
  </section>
}

function ProjectsCard({ items, onOpen }: { items: AttributionItem[]; onOpen: () => void }) {
  return <section className="card compactCard">
    <div className="cardHeader"><div className="headerWithIcon"><Icon name="projects"/><h2>Projects</h2></div><button className="textButton" onClick={onOpen}>View all</button></div>
    <div className="tableList">
      {items.length === 0 ? <Empty label="No project activity in this window"/> : items.slice(0, 5).map((item, index) => <div className="tableRow" key={item.key}>
        <span className={`projectSwatch swatch-${index % 4}`}/>
        <div className="rowMain"><strong>{item.label}</strong><small>{item.calls} calls</small></div>
        <div className="rowMetric"><strong>{formatTokens(item.tokens.total)}</strong><small>{item.sharePercent.toFixed(1)}%</small></div>
      </div>)}
    </div>
  </section>
}

function ModelsCard({ models, onOpen }: { models: ModelUsage[]; onOpen: () => void }) {
  const total = models.reduce((sum, model) => sum + model.tokens.total, 0)
  return <section className="card compactCard">
    <div className="cardHeader"><div className="headerWithIcon"><Icon name="models"/><h2>Models</h2></div><button className="textButton" onClick={onOpen}>View all</button></div>
    <div className="tableList">
      {models.length === 0 ? <Empty label="No token-bearing model calls today"/> : models.slice(0, 5).map((model, index) => <div className="tableRow" key={model.model}>
        <span className={`modelBadge model-${index % 4}`}>{model.model.slice(0, 1).toUpperCase()}</span>
        <div className="rowMain"><strong>{model.model}</strong><small>{model.calls} calls</small></div>
        <div className="rowMetric"><strong>{formatTokens(model.tokens.total)}</strong><small>{total ? `${(model.tokens.total / total * 100).toFixed(0)}% share` : '0%'}</small></div>
      </div>)}
    </div>
  </section>
}

function InsightCard({ status, onOpen }: { status: VarnqorSystemStatus; onOpen: () => void }) {
  const recommendations = status.advisor.recommendations ?? []
  const rules = status.waste.rules ?? []
  const rows: Array<{ title: string; meta: string; priority: string }> = recommendations.slice(0, 2).map(item => ({ title: item.title, meta: item.priority.toUpperCase(), priority: item.priority }))
  for (const rule of rules.filter(rule => rule.signals > 0).slice(0, Math.max(0, 3 - rows.length))) rows.push({ title: rule.label, meta: `~${formatTokens(rule.estimatedTokens)}`, priority: 'medium' })
  return <section className="card compactCard">
    <div className="cardHeader"><div className="headerWithIcon accent"><Icon name="advisor"/><h2>Advisor & Waste</h2></div><button className="textButton" onClick={onOpen}>View all</button></div>
    <div className="insightList">
      {rows.length === 0 ? <Empty label="No actionable signal right now"/> : rows.map((item, index) => <button className="insightRow" key={`${item.title}-${index}`} onClick={onOpen}>
        <span className={`rank rank-${item.priority}`}>{index + 1}</span>
        <div><strong>{item.title}</strong><small>{item.meta}</small></div>
        <Icon name="arrow" size={15}/>
      </button>)}
    </div>
  </section>
}

function Overview({ status, refreshing, lastUpdated, setView }: { status: VarnqorSystemStatus; refreshing: boolean; lastUpdated: Date | null; setView: (value: ViewName) => void }) {
  const projects = status.attribution.projects ?? []
  const models = status.usage.models ?? []
  const recommendations = status.advisor.recommendations ?? []
  return <>
    <div className="statsGrid">
      <StatCard icon="burn" label="Burn rate" value={status.burn.ready ? `${status.burn.ratio.toFixed(2)}×` : 'Learning'} meta={burnMeta(status.burn)} tone={status.burn.ready && status.burn.ratio >= 1.5 ? 'warn' : 'neutral'} onClick={() => setView('burn')}/>
      <StatCard icon="pulse" label="Current activity" value={`${formatTokens(status.burn.currentPerHour)}/h`} meta={`${formatTokens(status.usage.tokens.total)} today`} tone="info" onClick={() => setView('activity')}/>
      <StatCard icon="warning" label="Resource pressure" value={status.forecast.pressure.toUpperCase()} meta={status.forecast.ready ? `${status.forecast.pressureRatio.toFixed(2)}× active baseline` : 'Learning from local history'} tone={pressureTone(status.forecast.pressure)} onClick={() => setView('burn')}/>
      <StatCard icon="clock" label="Runway" value={formatRunway(status.quota)} meta={status.quota.available ? `${status.quota.utilizationPercent.toFixed(0)}% of configured budget` : 'Set a trusted budget to enable'} tone={status.quota.willExceedBeforeReset ? 'danger' : 'neutral'} onClick={() => setView('settings')}/>
    </div>
    <div className="overviewMain"><ActivityChart hourly={status.hourly ?? []} total={status.usage.tokens.total}/><SystemsCard status={status} refreshing={refreshing} lastUpdated={lastUpdated}/></div>
    <div className="overviewBottom"><ProjectsCard items={projects} onOpen={() => setView('projects')}/><ModelsCard models={models} onOpen={() => setView('models')}/><InsightCard status={status} onOpen={() => setView('advisor')}/></div>
    {recommendations.length === 0 && status.waste.estimatedSignalTokens === 0 && <div className="softStatus"><span className="statusDot ok"/>No high-value intervention is needed right now.</div>}
  </>
}

function BurnView({ status }: { status: VarnqorSystemStatus }) {
  const projects = status.attribution.projects ?? []
  const sessions = status.attribution.sessions ?? []
  const points = status.forecast.points ?? []
  return <div className="pageStack">
    <section className={`card burnHero tone-${pressureTone(status.forecast.pressure)}`}>
      <div className="burnIdentity">
        <span className="sectionEyebrow">Current pace</span>
        <div className="burnValue">{status.burn.ready ? `${status.burn.ratio.toFixed(2)}×` : '—'}</div>
        <div className="burnStatus"><span className={`statusDot ${status.burn.ready ? 'ok' : 'sync'}`}/>{status.burn.ready ? status.burn.level.toUpperCase() : 'LEARNING'}</div>
      </div>
      <div className="burnFacts">
        <div><span>Last 60 min</span><strong>{formatTokens(status.burn.currentPerHour)}</strong></div>
        <div><span>Previous 60 min</span><strong>{formatTokens(status.burn.previousPerHour)}</strong></div>
        <div><span>Active baseline</span><strong>{status.burn.baselinePerHour ? formatTokens(status.burn.baselinePerHour) : '—'}</strong></div>
        <div><span>Confidence</span><strong>{status.burn.confidence.toUpperCase()}</strong></div>
      </div>
    </section>

    <div className="burnGrid">
      <section className="card forecastCard">
        <div className="cardHeader"><div><h2>Projection</h2><p>Persistence model from current pace</p></div><span className={`pressureChip pressure-${status.forecast.pressure}`}>{status.forecast.pressure.toUpperCase()}</span></div>
        <div className="projectionHero"><strong>{status.forecast.ready ? formatTokens(status.forecast.projectedTokens) : '—'}</strong><span>projected over {status.forecast.horizonMinutes / 60}h</span></div>
        <div className="projectionRows">
          {points.slice(0, 3).map(point => <div key={point.minutes}><span>{point.minutes < 60 ? `${point.minutes}m` : `${point.minutes / 60}h`}</span><strong>{formatTokens(point.projectedTokens)}</strong><small>{formatTokens(point.envelopeLow)}–{formatTokens(point.envelopeHigh)}</small></div>)}
        </div>
      </section>
      <section className="card runwayCard">
        <div className="cardHeader"><div><h2>Trusted runway</h2><p>{status.quota.available ? status.quota.source.toUpperCase() : 'No configured limit'}</p></div><Icon name="shield"/></div>
        <div className="runwayValue">{formatRunway(status.quota)}</div>
        <div className="quotaTrack"><span style={{ width: `${Math.min(100, status.quota.utilizationPercent)}%` }}/></div>
        <div className="quotaMeta"><span>{status.quota.available ? `${formatTokens(status.quota.used)} used` : 'Configure a daily budget'}</span><strong>{status.quota.available ? `${formatTokens(status.quota.remaining)} remaining` : ''}</strong></div>
      </section>
    </div>

    <section className="card tableCard">
      <div className="cardHeader"><div><h2>Burn drivers</h2><p>Last 60 minutes</p></div><span className="quietPill">{status.attribution.calls} calls</span></div>
      <div className="dataTable dataTableProjects">
        <div className="dataHead"><span>Project</span><span>Calls</span><span>Tokens</span><span>Share</span></div>
        {projects.length === 0 ? <Empty label="No project attribution in this window"/> : projects.slice(0, 8).map(item => <div className="dataRow" key={item.key}><strong>{item.label}</strong><span>{item.calls}</span><span>{formatTokens(item.tokens.total)}</span><span>{item.sharePercent.toFixed(1)}%</span></div>)}
      </div>
    </section>

    <section className="card tableCard">
      <div className="cardHeader"><div><h2>Expensive sessions</h2><p>Highest contributors in the current burn window</p></div></div>
      <div className="dataTable dataTableSessions">
        <div className="dataHead"><span>Session</span><span>Project</span><span>Calls</span><span>Tokens</span></div>
        {sessions.length === 0 ? <Empty label="No sessions attributed in this window"/> : sessions.slice(0, 8).map(item => <div className="dataRow" key={item.key}><strong>{shortSession(item.session ?? item.label)}</strong><span>{item.project || '—'}</span><span>{item.calls}</span><span>{formatTokens(item.tokens.total)}</span></div>)}
      </div>
    </section>
  </div>
}

function AdvisorView({ status }: { status: VarnqorSystemStatus }) {
  const recommendations = status.advisor.recommendations ?? []
  const findings = status.waste.findings ?? []
  const rules = status.waste.rules ?? []
  return <div className="pageStack">
    <div className="advisorSummaryGrid">
      <StatCard icon="spark" label="High priority" value={String(status.advisor.high)} meta="Immediate review" tone={status.advisor.high ? 'danger' : 'neutral'}/>
      <StatCard icon="spark" label="Medium priority" value={String(status.advisor.medium)} meta="Worth reviewing" tone={status.advisor.medium ? 'warn' : 'neutral'}/>
      <StatCard icon="pulse" label="Signal context" value={status.waste.estimatedSignalTokens ? `~${formatTokens(status.waste.estimatedSignalTokens)}` : '—'} meta="Heuristic, overlap possible" tone="info"/>
    </div>

    <section className="card recommendationsCard">
      <div className="cardHeader"><div><h2>Recommendations</h2><p>Read-only guidance. Nothing is changed automatically.</p></div><span className="quietPill">ADVISORY ONLY</span></div>
      <div className="recommendationList">
        {recommendations.length === 0 ? <Empty label="No recommendation is strong enough to surface"/> : recommendations.map((item, index) => <article className={`recommendation priority-${item.priority}`} key={item.id}>
          <span className="recommendationIndex">{String(index + 1).padStart(2, '0')}</span>
          <div className="recommendationBody"><div><span className={`priorityLabel priority-${item.priority}`}>{item.priority}</span><strong>{item.title}</strong></div><p>{item.evidence}</p><small>{item.recommendation}</small></div>
          <div className="recommendationMetric">{item.estimatedTokens ? <><strong>~{formatTokens(item.estimatedTokens)}</strong><span>signal context</span></> : <span>{item.category}</span>}</div>
        </article>)}
      </div>
    </section>

    <section className="card wasteCard">
      <div className="cardHeader"><div><h2>Waste intelligence</h2><p>Explainable signals, not a synthetic score</p></div><span className="quietPill">{status.waste.toolCalls} tool calls analyzed</span></div>
      <div className="ruleGrid">
        {rules.map(rule => <div className={`ruleTile ${rule.findings ? 'active' : ''}`} key={rule.ruleId}><span>{rule.label}</span><strong>{rule.signals}</strong><small>{rule.findings} findings{rule.estimatedTokens > 0 ? ` · ~${formatTokens(rule.estimatedTokens)}` : ''}</small></div>)}
      </div>
      <div className="findingsList">
        {findings.length === 0 ? <Empty label="No current waste heuristic matched"/> : findings.slice(0, 8).map((finding, index) => <div className="findingRow" key={`${finding.ruleId}-${index}`} title={finding.explanation}>
          <div><strong>{finding.filePath ? shortPath(finding.filePath) : finding.target || finding.toolName || finding.title}</strong><small>{finding.title} · {finding.project || 'unknown'}</small></div>
          <span>{finding.toolName || finding.ruleId}</span>
          <strong>{finding.estimatedTokens > 0 ? `~${formatTokens(finding.estimatedTokens)}` : 'n/a'}</strong>
          <span>{finding.confidence.toUpperCase()}</span>
        </div>)}
      </div>
    </section>
  </div>
}

function ActivityView({ status }: { status: VarnqorSystemStatus }) {
  const tokens = status.usage.tokens
  return <div className="pageStack">
    <div className="tokenGrid">
      <StatCard icon="pulse" label="Input" value={formatTokens(tokens.input)} meta="Uncached input" tone="neutral"/>
      <StatCard icon="pulse" label="Output" value={formatTokens(tokens.output)} meta="Generated output" tone="info"/>
      <StatCard icon="database" label="Cache read" value={formatTokens(tokens.cacheRead)} meta={`${status.usage.cacheReadShare.toFixed(1)}% read share`} tone="good"/>
      <StatCard icon="database" label="Cache write" value={formatTokens(tokens.cacheCreation)} meta="Created cache input" tone="neutral"/>
    </div>
    <ActivityChart hourly={status.hourly ?? []} total={tokens.total} large/>
    <div className="activityFacts">
      <section className="card factCard"><Icon name="terminal"/><div><span>API calls</span><strong>{number.format(status.usage.calls)}</strong><small>Token-bearing calls today</small></div></section>
      <section className="card factCard"><Icon name="projects"/><div><span>Active sessions</span><strong>{number.format(status.usage.sessions)}</strong><small>Observed today</small></div></section>
      <section className="card factCard"><Icon name="database"/><div><span>Indexed sources</span><strong>{number.format(status.usage.filesScanned)}</strong><small>{formatBytes(status.sync.bytesIndexed)} added in last sync</small></div></section>
      <section className="card factCard"><Icon name={status.usage.parseWarnings ? 'warning' : 'check'}/><div><span>Parse health</span><strong>{status.usage.parseWarnings ? `${status.usage.parseWarnings} warnings` : 'Healthy'}</strong><small>Local-only parser state</small></div></section>
    </div>
  </div>
}

function ProjectsView({ status }: { status: VarnqorSystemStatus }) {
  const projects = status.attribution.projects ?? []
  const sessions = status.attribution.sessions ?? []
  return <div className="pageStack">
    <section className="card tableCard">
      <div className="cardHeader"><div><h2>Projects</h2><p>Share of activity in the current 60-minute window</p></div><span className="quietPill">{projects.length} active</span></div>
      <div className="dataTable dataTableProjects">
        <div className="dataHead"><span>Name</span><span>Calls</span><span>Tokens</span><span>Share</span></div>
        {projects.length === 0 ? <Empty label="No project activity in this window"/> : projects.map(item => <div className="dataRow" key={item.key}><strong>{item.label}</strong><span>{item.calls}</span><span>{formatTokens(item.tokens.total)}</span><span>{item.sharePercent.toFixed(1)}%</span></div>)}
      </div>
    </section>
    <section className="card tableCard">
      <div className="cardHeader"><div><h2>Sessions</h2><p>Most active sessions by token volume</p></div></div>
      <div className="dataTable dataTableSessions">
        <div className="dataHead"><span>Session</span><span>Project</span><span>Calls</span><span>Tokens</span></div>
        {sessions.length === 0 ? <Empty label="No session attribution available"/> : sessions.slice(0, 12).map(item => <div className="dataRow" key={item.key}><strong>{shortSession(item.session ?? item.label)}</strong><span>{item.project || '—'}</span><span>{item.calls}</span><span>{formatTokens(item.tokens.total)}</span></div>)}
      </div>
    </section>
  </div>
}

function ModelsView({ status }: { status: VarnqorSystemStatus }) {
  const models = status.usage.models ?? []
  const total = models.reduce((sum, model) => sum + model.tokens.total, 0)
  return <div className="pageStack">
    <section className="card tableCard">
      <div className="cardHeader"><div><h2>Models observed</h2><p>Today, from local Claude Code usage events</p></div><span className="quietPill">{models.length} models</span></div>
      <div className="dataTable dataTableModels">
        <div className="dataHead"><span>Model</span><span>Calls</span><span>Tokens</span><span>Share</span><span>Cache read</span></div>
        {models.length === 0 ? <Empty label="No token-bearing model calls today"/> : models.map(model => <div className="dataRow" key={model.model}><strong>{model.model}</strong><span>{model.calls}</span><span>{formatTokens(model.tokens.total)}</span><span>{total ? `${(model.tokens.total / total * 100).toFixed(1)}%` : '0%'}</span><span>{formatTokens(model.tokens.cacheRead)}</span></div>)}
      </div>
    </section>
  </div>
}

function SettingsView({ settings, setSettings, onSave, saving, saveError, doctor, onDoctor, doctorLoading, version }: {
  settings: VarnqorSettings
  setSettings: (value: VarnqorSettings) => void
  onSave: () => void
  saving: boolean
  saveError: string
  doctor: DoctorReport | null
  onDoctor: () => void
  doctorLoading: boolean
  version: string
}) {
  const budget = settings.budget
  const updateBudget = (patch: Partial<ConfiguredBudgetSettings>) => setSettings({ ...settings, budget: { ...budget, ...patch } })
  return <div className="settingsLayout">
    <section className="card settingsCard">
      <div className="settingsHeader"><div><h2>Live monitor</h2><p>Polling is local and only runs while the desktop app is open.</p></div></div>
      <label className="settingRow"><div><strong>Refresh interval</strong><small>2 to 60 seconds</small></div><div className="inputWithSuffix"><input type="number" min={2} max={60} value={settings.refreshSeconds} onChange={event => setSettings({ ...settings, refreshSeconds: Number(event.target.value) })}/><span>s</span></div></label>

      <div className="settingsDivider"/>
      <div className="settingsHeader"><div><h2>Trusted budget</h2><p>A configured budget is your limit, not an official provider quota.</p></div><button className={`switch ${budget.enabled ? 'on' : ''}`} onClick={() => updateBudget({ enabled: !budget.enabled })}><i/></button></div>
      <label className={`settingRow ${!budget.enabled ? 'disabled' : ''}`}><div><strong>Daily token budget</strong><small>Used for remaining and runway calculations</small></div><input type="number" min={0} step={100000} disabled={!budget.enabled} value={budget.limitTokens} onChange={event => updateBudget({ limitTokens: Number(event.target.value) })}/></label>
      <label className={`settingRow ${!budget.enabled ? 'disabled' : ''}`}><div><strong>Reset hour</strong><small>Local time, 0–23</small></div><input type="number" min={0} max={23} disabled={!budget.enabled} value={budget.resetHour} onChange={event => updateBudget({ resetHour: Number(event.target.value) })}/></label>
      {saveError && <div className="inlineError">{saveError}</div>}
      <div className="settingsActions"><button className="primaryButton" disabled={saving} onClick={onSave}>{saving ? 'Saving…' : 'Save changes'}</button></div>
    </section>

    <aside className="settingsSide">
      <section className="card doctorCard">
        <div className="cardHeader"><div><h2>Diagnostics</h2><p>Local environment checks</p></div><button className="secondaryButton" onClick={onDoctor} disabled={doctorLoading}>{doctorLoading ? 'Running…' : 'Run doctor'}</button></div>
        {!doctor ? <Empty label="Run Doctor to verify the local environment"/> : <div className="doctorList">{(doctor.checks ?? []).map(check => <div className="doctorItem" key={check.id}><span className={`doctorDot state-${check.status}`}/><div><strong>{check.label}</strong><small>{check.detail}</small></div></div>)}</div>}
      </section>
      <section className="card aboutCard"><span>Version</span><strong>{version}</strong><small>Local-first · no telemetry by default</small></section>
    </aside>
  </div>
}

export default function App() {
  const [view, setView] = useState<ViewName>('overview')
  const [status, setStatus] = useState<VarnqorSystemStatus>(emptyStatus)
  const [settings, setSettings] = useState<VarnqorSettings>(defaultSettings)
  const [loading, setLoading] = useState(true)
  const [refreshing, setRefreshing] = useState(false)
  const [lastUpdated, setLastUpdated] = useState<Date | null>(null)
  const [saving, setSaving] = useState(false)
  const [saveError, setSaveError] = useState('')
  const [doctor, setDoctor] = useState<DoctorReport | null>(null)
  const [doctorLoading, setDoctorLoading] = useState(false)

  useEffect(() => {
    let active = true
    const loadSettings = async () => {
      const fn = window.go?.main?.App?.GetSettings
      if (!fn) return
      try {
        const value = await fn()
        if (active) setSettings(value)
      } catch (error) {
        console.error('Unable to load settings', error)
      }
    }
    void loadSettings()
    return () => { active = false }
  }, [])

  useEffect(() => {
    let disposed = false
    let busy = false
    const loadStatus = async () => {
      if (busy) return
      const fn = window.go?.main?.App?.GetSystemStatus
      if (!fn) { setLoading(false); return }
      busy = true
      setRefreshing(true)
      try {
        const value = await fn()
        if (!disposed) {
          setStatus(value)
          setLastUpdated(new Date())
        }
      } catch (error) {
        console.error('Unable to read system status', error)
      } finally {
        busy = false
        if (!disposed) {
          setLoading(false)
          setRefreshing(false)
        }
      }
    }
    void loadStatus()
    const id = window.setInterval(() => void loadStatus(), Math.max(2, settings.refreshSeconds) * 1000)
    return () => { disposed = true; window.clearInterval(id) }
  }, [settings.refreshSeconds])

  const saveSettings = async () => {
    const fn = window.go?.main?.App?.SaveSettings
    if (!fn) return
    setSaving(true)
    setSaveError('')
    try {
      const saved = await fn(settings)
      setSettings(saved)
    } catch (error) {
      setSaveError(error instanceof Error ? error.message : String(error))
    } finally {
      setSaving(false)
    }
  }

  const runDoctor = async () => {
    const fn = window.go?.main?.App?.RunDoctor
    if (!fn) return
    setDoctorLoading(true)
    try {
      setDoctor(await fn())
    } catch (error) {
      console.error('Unable to run Doctor', error)
    } finally {
      setDoctorLoading(false)
    }
  }

  const body = useMemo(() => {
    switch (view) {
      case 'overview': return <Overview status={status} refreshing={refreshing} lastUpdated={lastUpdated} setView={setView}/>
      case 'burn': return <BurnView status={status}/>
      case 'advisor': return <AdvisorView status={status}/>
      case 'activity': return <ActivityView status={status}/>
      case 'projects': return <ProjectsView status={status}/>
      case 'models': return <ModelsView status={status}/>
      case 'settings': return <SettingsView settings={settings} setSettings={setSettings} onSave={saveSettings} saving={saving} saveError={saveError} doctor={doctor} onDoctor={runDoctor} doctorLoading={doctorLoading} version={status.version}/>
    }
  }, [view, status, refreshing, lastUpdated, settings, saving, saveError, doctor, doctorLoading])

  const page = pageCopy[view]
  const liveState = status.usageError ? 'Attention' : refreshing ? 'Syncing' : status.claude.detected ? 'Live' : 'Offline'

  return <div className="appShell">
    <aside className="sidebar" aria-label="Primary navigation">
      <div className="sidebarSpacer"/>
      <nav className="navList">
        {nav.slice(0, 6).map(item => <button className={`navItem ${view === item.id ? 'active' : ''}`} key={item.id} onClick={() => setView(item.id)}><Icon name={item.id}/><span>{item.label}</span></button>)}
      </nav>
      <div className="sidebarBottom">
        <button className={`navItem ${view === 'settings' ? 'active' : ''}`} onClick={() => setView('settings')}><Icon name="settings"/><span>Settings</span></button>
        <div className="liveRail"><span className={`statusDot ${refreshing ? 'sync' : status.claude.detected ? 'ok' : 'off'}`}/><div><strong>{liveState}</strong><small>{lastUpdated ? `Updated ${lastUpdated.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}` : 'Waiting for local signal'}</small></div></div>
      </div>
    </aside>

    <main className="workspace">
      <header className="workspaceHeader">
        <div className="pageIdentity"><h1>{page.title}</h1><p>{page.subtitle}</p></div>
        <div className="headerTools">
          <span className={`livePill ${refreshing ? 'syncing' : status.claude.detected ? 'connected' : 'offline'}`}><span className="statusDot"/>{liveState}</span>
          {view !== 'settings' && <span className="windowPill"><Icon name="clock" size={15}/>Last 60 minutes</span>}
        </div>
      </header>

      <div className="workspaceContent">
        {loading ? <div className="loadingState"><div className="loadingRing"/><span>Reading local activity…</span></div> : body}
        {status.usageError && <div className="globalNotice error"><Icon name="warning" size={16}/><span>{status.usageError}</span></div>}
      </div>
    </main>
  </div>
}
