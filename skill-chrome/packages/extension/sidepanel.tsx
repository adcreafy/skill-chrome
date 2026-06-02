import { useEffect, useState, useCallback, useMemo } from "react"
import "~styles/globals.css"
import {
  Loader2,
  CheckCircle2,
  AlertCircle,
  Terminal,
  Copy,
  Check,
} from "lucide-react"
import { SkillCard } from "~components/skill-card"
import { BottomBar } from "~components/bottom-bar"
import { EmptyState } from "~components/empty-state"
import { Button } from "~components/ui/button"
import { ping, detectAgents, installSkills } from "~services/native-client"
import type { DetectedSkill, DetectedEngine } from "~types"

function skillKey(skill: DetectedSkill) {
  return skill.command
}

export default function SidePanel() {
  const [hostAvailable, setHostAvailable] = useState<boolean | null>(null)
  const [detectedSkills, setDetectedSkills] = useState<DetectedSkill[]>([])
  const [selectedSkills, setSelectedSkills] = useState<Set<string>>(new Set())
  const [engines, setEngines] = useState<DetectedEngine[]>([])
  const [selectedEngines, setSelectedEngines] = useState<Set<string>>(new Set())
  const [scanning, setScanning] = useState(true)
  const [installing, setInstalling] = useState(false)
  const [status, setStatus] = useState<{
    type: "success" | "error"
    message: string
  } | null>(null)

  const checkHost = useCallback(async () => {
    const available = await ping()
    setHostAvailable(available)
    if (available) {
      try {
        const detected = await detectAgents()
        setEngines(detected)
        const detectedIds = detected
          .filter((e) => e.detected)
          .map((e) => e.id)
        setSelectedEngines(new Set(detectedIds))
      } catch {
        // Host available but detect failed — show engines empty
      }
    }
  }, [])

  const refreshPageSkills = useCallback(async () => {
    setScanning(true)
    try {
      const [tab] = await chrome.tabs.query({
        active: true,
        currentWindow: true,
      })
      if (!tab?.id) {
        setDetectedSkills([])
        return
      }

      await new Promise<void>((resolve) => {
        chrome.runtime.sendMessage(
          { type: "DETECT_ON_TAB", tabId: tab.id },
          (response) => {
            const skills = (response?.skills ?? []) as DetectedSkill[]
            setDetectedSkills(skills)
            setSelectedSkills(new Set())
            resolve()
          },
        )
      })
    } finally {
      setScanning(false)
    }
  }, [])

  useEffect(() => {
    checkHost()
    refreshPageSkills()

    const listener = (msg: {
      type: string
      skills?: DetectedSkill[]
      payload?: DetectedSkill
    }) => {
      if (msg.type === "TAB_SKILLS_UPDATED" && msg.skills) {
        setDetectedSkills(msg.skills)
        setSelectedSkills(new Set())
      }
      if (msg.type === "SKILL_DETECTED" && msg.payload) {
        setDetectedSkills((prev) => {
          const exists = prev.some((s) => s.command === msg.payload!.command)
          if (exists) return prev
          return [msg.payload!, ...prev]
        })
      }
    }

    chrome.runtime.onMessage.addListener(listener)
    return () => chrome.runtime.onMessage.removeListener(listener)
  }, [checkHost, refreshPageSkills])

  function toggleSkill(skill: DetectedSkill) {
    const key = skillKey(skill)
    setSelectedSkills((prev) => {
      const next = new Set(prev)
      if (next.has(key)) next.delete(key)
      else next.add(key)
      return next
    })
  }

  function toggleEngine(id: string) {
    setSelectedEngines((prev) => {
      const next = new Set(prev)
      if (next.has(id)) next.delete(id)
      else next.add(id)
      return next
    })
  }

  async function handleInstall() {
    const sources = detectedSkills
      .filter((s) => selectedSkills.has(skillKey(s)))
      .map((s) => s.command)

    if (sources.length === 0 || selectedEngines.size === 0) return

    setInstalling(true)
    setStatus(null)
    try {
      const result = await installSkills(sources, [...selectedEngines])

      const ok = result.succeeded.length
      const fail = result.failed.length
      if (fail === 0) {
        setStatus({
          type: "success",
          message: `Installed ${ok} skill${ok !== 1 ? "s" : ""} successfully.`,
        })
        setSelectedSkills(new Set())
      } else if (ok > 0) {
        setStatus({
          type: "error",
          message: `${ok} succeeded, ${fail} failed.`,
        })
      } else {
        setStatus({
          type: "error",
          message: result.failed[0]?.error ?? "Install failed",
        })
      }
    } catch (err) {
      setStatus({
        type: "error",
        message: err instanceof Error ? err.message : "Install failed",
      })
    } finally {
      setInstalling(false)
    }
  }

  // Host not yet checked
  if (hostAvailable === null) {
    return (
      <div className="flex items-center justify-center h-screen bg-canvas">
        <Loader2 className="w-5 h-5 animate-spin text-muted" />
      </div>
    )
  }

  // Host not installed — show setup guide
  if (!hostAvailable) {
    return <SetupGuide onRecheck={checkHost} />
  }

  return (
    <div className="flex flex-col h-screen bg-canvas">
      {status && (
        <div
          className={`shrink-0 px-4 py-2 text-caption flex items-center gap-2 ${
            status.type === "success"
              ? "bg-success/10 text-success"
              : "bg-error/10 text-error"
          }`}>
          {status.type === "success" ? (
            <CheckCircle2 className="w-4 h-4 shrink-0" />
          ) : (
            <AlertCircle className="w-4 h-4 shrink-0" />
          )}
          <span>{status.message}</span>
        </div>
      )}

      <div className="flex-1 overflow-y-auto px-4 py-3">
        {scanning ? (
          <div className="flex items-center justify-center gap-2 py-16 text-caption text-muted">
            <Loader2 className="w-4 h-4 animate-spin" />
            Scanning page...
          </div>
        ) : detectedSkills.length === 0 ? (
          <EmptyState onRescan={refreshPageSkills} />
        ) : (
          <>
            <div className="flex items-center justify-between mb-3">
              <p className="text-caption text-muted">
                {detectedSkills.length} on this page
              </p>
              <button
                type="button"
                onClick={() => {
                  const all = detectedSkills.map(skillKey)
                  const allSelected = all.every((k) => selectedSkills.has(k))
                  setSelectedSkills(allSelected ? new Set() : new Set(all))
                }}
                className="text-caption text-primary hover:text-primary-active">
                {detectedSkills.every((s) =>
                  selectedSkills.has(skillKey(s)),
                )
                  ? "Clear all"
                  : "Select all"}
              </button>
            </div>
            <div className="flex flex-col gap-2 pb-2">
              {detectedSkills.map((skill) => (
                <SkillCard
                  key={skillKey(skill)}
                  skill={skill}
                  selected={selectedSkills.has(skillKey(skill))}
                  onToggle={toggleSkill}
                />
              ))}
            </div>
          </>
        )}
      </div>

      <BottomBar
        engines={engines}
        selectedEngineIds={selectedEngines}
        selectedSkillCount={selectedSkills.size}
        installing={installing}
        onToggleEngine={toggleEngine}
        onInstall={handleInstall}
      />
    </div>
  )
}

const INSTALL_COMMANDS = {
  mac: `bash <(curl -sL https://github.com/adcreafy/skill-chrome/releases/latest/download/install.sh)`,
  windows: `irm https://github.com/adcreafy/skill-chrome/releases/latest/download/install.ps1 | iex`,
  linux: `bash <(curl -sL https://github.com/adcreafy/skill-chrome/releases/latest/download/install.sh)`,
} as const

type Platform = keyof typeof INSTALL_COMMANDS

const PLATFORM_LABELS: Record<Platform, string> = {
  mac: "macOS",
  windows: "Windows",
  linux: "Linux",
}

const PLATFORM_HINTS: Record<Platform, string> = {
  mac: "Open Terminal.app and paste:",
  windows: "Open PowerShell and paste:",
  linux: "Open a terminal and paste:",
}

function detectPlatform(): Platform {
  const ua = navigator.userAgent.toLowerCase()
  if (ua.includes("win")) return "windows"
  if (ua.includes("linux")) return "linux"
  return "mac"
}

function SetupGuide({ onRecheck }: { onRecheck: () => void }) {
  const detectedPlatform = useMemo(detectPlatform, [])
  const [platform, setPlatform] = useState<Platform>(detectedPlatform)
  const [copied, setCopied] = useState(false)

  const command = INSTALL_COMMANDS[platform]

  async function handleCopy() {
    try {
      await navigator.clipboard.writeText(command)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    } catch {
      // Fallback: select text
      const el = document.querySelector("[data-install-cmd]") as HTMLElement
      if (el) {
        const range = document.createRange()
        range.selectNodeContents(el)
        const sel = window.getSelection()
        sel?.removeAllRanges()
        sel?.addRange(range)
      }
    }
  }

  return (
    <div className="flex flex-col items-center justify-center h-screen bg-canvas px-5 text-center">
      <div className="w-12 h-12 rounded-xl bg-surface-strong flex items-center justify-center mb-4">
        <Terminal className="w-6 h-6 text-muted" />
      </div>
      <p className="text-body-sm text-body-strong mb-1">
        One-time setup required
      </p>
      <p className="text-caption text-muted mb-4 max-w-[260px]">
        Install the local helper to enable skill installation.
      </p>

      <div className="flex items-center gap-1 mb-3">
        {(Object.keys(INSTALL_COMMANDS) as Platform[]).map((p) => (
          <button
            key={p}
            type="button"
            onClick={() => setPlatform(p)}
            className={`px-2.5 py-1 rounded-md text-caption transition-colors ${
              platform === p
                ? "bg-primary text-on-primary"
                : "bg-surface-strong text-muted hover:text-body"
            }`}>
            {PLATFORM_LABELS[p]}
          </button>
        ))}
      </div>

      <p className="text-caption text-muted mb-2">
        {PLATFORM_HINTS[platform]}
      </p>

      <div className="w-full bg-surface-strong rounded-md p-3 mb-4 relative group">
        <code
          data-install-cmd
          className="text-caption text-body break-all block pr-8 text-left">
          {command}
        </code>
        <button
          type="button"
          onClick={handleCopy}
          className="absolute top-2 right-2 p-1.5 rounded-md bg-canvas/80 hover:bg-canvas text-muted hover:text-body transition-colors"
          title="Copy">
          {copied ? (
            <Check className="w-3.5 h-3.5 text-success" />
          ) : (
            <Copy className="w-3.5 h-3.5" />
          )}
        </button>
      </div>

      <Button variant="secondary" size="sm" onClick={onRecheck}>
        I've completed setup — recheck
      </Button>
    </div>
  )
}
