type ClaudeDetection = {
  detected: boolean
  binaryDetected: boolean
  dataDetected: boolean
  claudeHome: string
  projectsPath: string
  projects: number
  sessions: number
  lastActivity?: string
}

type TokenTotals = {
  input: number
  output: number
  cacheCreation: number
  cacheRead: number
  cacheCreation5m: number
  cacheCreation1h: number
  total: number
}

type ModelUsage = {
  model: string
  calls: number
  tokens: TokenTotals
}

type UsageOverview = {
  date: string
  calls: number
  sessions: number
  projects: number
  filesScanned: number
  parseWarnings: number
  cacheReadShare: number
  firstActivity?: string
  lastActivity?: string
  tokens: TokenTotals
  models: ModelUsage[] | null
}

type HourUsage = {
  start: string
  calls: number
  tokens: TokenTotals
}

type BurnRate = {
  ready: boolean
  currentPerHour: number
  previousPerHour: number
  baselinePerHour: number
  ratio: number
  changePercent: number
  trend: 'accelerating' | 'cooling' | 'steady' | 'new-spike' | 'idle'
  level: 'learning' | 'low' | 'normal' | 'elevated' | 'high' | 'critical'
  confidence: 'insufficient' | 'low' | 'medium' | 'high'
  baselineSamples: number
}

type AttributionItem = {
  key: string
  label: string
  project?: string
  session?: string
  calls: number
  tokens: TokenTotals
  sharePercent: number
}

type AttributionReport = {
  start?: string
  end?: string
  calls: number
  tokens: TokenTotals
  projects: AttributionItem[] | null
  sessions: AttributionItem[] | null
  models: AttributionItem[] | null
  agents: AttributionItem[] | null
}

type WasteFinding = {
  ruleId: string
  title: string
  severity: 'low' | 'medium' | 'high'
  confidence: 'low' | 'medium'
  project: string
  session: string
  filePath: string
  readOffset?: number
  readLimit?: number
  reads: number
  repeats: number
  repeatedBytes: number
  estimatedTokens: number
  explanation: string
}

type WasteReport = {
  start?: string
  end?: string
  toolCalls: number
  repeatedReads: number
  estimatedRepeatedTokens: number
  confidence: 'insufficient' | 'low' | 'medium'
  findings: WasteFinding[] | null
}

type SyncStats = {
  filesSeen: number
  filesChanged: number
  bytesIndexed: number
  recordsObserved: number
  parseWarnings: number
  rebuilt: boolean
}

type VarnqorSystemStatus = {
  claude: ClaudeDetection
  usage: UsageOverview
  hourly: HourUsage[] | null
  burn: BurnRate
  attribution: AttributionReport
  waste: WasteReport
  sync: SyncStats
  storageReady: boolean
  usageError?: string
}

interface Window {
  go?: {
    main?: {
      App?: {
        GetSystemStatus: () => Promise<VarnqorSystemStatus>
      }
    }
  }
}
