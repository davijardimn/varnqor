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

type VarnqorSystemStatus = {
  claude: ClaudeDetection
  usage: UsageOverview
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
