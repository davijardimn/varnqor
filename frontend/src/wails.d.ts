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
