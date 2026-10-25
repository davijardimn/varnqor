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

type VarnqorSystemStatus = {
  claude: ClaudeDetection
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
