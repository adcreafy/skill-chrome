// --- Skill Detection (from content script) ---

export interface DetectedSkill {
  name: string
  source: SkillSource
  command: string
  description?: string
  pageUrl: string
  detectedAt: number
  confidence: "high" | "medium" | "low"
}

export type SkillSource =
  | { type: "registry"; registry: string }
  | { type: "github"; owner: string; repo: string }
  | { type: "command"; raw: string }

// --- Native Host Response Types ---

export interface DetectedEngine {
  id: string
  name: string
  detected: boolean
  skillsPath: string
  apps: DetectedApp[]
}

export interface DetectedApp {
  name: string
  bundlePath: string
}

export interface InstallSucceeded {
  skillName: string
  engines: string[]
  fileCount: number
}

export interface InstallFailed {
  source: string
  error: string
}

export interface InstallResult {
  succeeded: InstallSucceeded[]
  failed: InstallFailed[]
}
