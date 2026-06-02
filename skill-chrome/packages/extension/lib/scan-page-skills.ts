/** Shared page scan logic (content script + injected fallback). */

export interface ScannedSkill {
  name: string
  source: Record<string, unknown>
  command: string
  pageUrl: string
  detectedAt: number
  confidence: string
}

const NPX_PATTERN = /npx\s+skills\s+add\s+(\S+)/gi

function deriveName(raw: string): string {
  const cleaned = raw.replace(/\/+$/, "")
  return cleaned.split("/").pop() ?? raw
}

function inferSource(raw: string): Record<string, unknown> {
  if (raw.includes("github.com")) {
    const ghMatch = /github\.com\/([^/]+)\/([^/]+)/.exec(raw)
    if (ghMatch) {
      return { type: "github", owner: ghMatch[1], repo: ghMatch[2] }
    }
  }
  if (raw.includes("skills.sh")) {
    return { type: "registry", registry: "skills.sh" }
  }
  return { type: "command", raw }
}

export function scanPageForSkills(): ScannedSkill[] {
  const pageUrl = window.location.href
  const visited = new Set<string>()
  const results: ScannedSkill[] = []

  function collect(text: string, confidence: string) {
    NPX_PATTERN.lastIndex = 0
    let match: RegExpExecArray | null
    while ((match = NPX_PATTERN.exec(text)) !== null) {
      const raw = match[1]
      if (visited.has(raw)) continue
      visited.add(raw)
      results.push({
        name: deriveName(raw),
        source: inferSource(raw),
        command: raw,
        pageUrl,
        detectedAt: Date.now(),
        confidence,
      })
    }
  }

  const codeElements = document.querySelectorAll(
    "pre, code, .highlight, [class*='code'], [class*='Code'], [class*='terminal'], [class*='Terminal'], [class*='shell'], [class*='Shell']",
  )
  codeElements.forEach((el) => {
    collect(el.textContent?.trim() ?? "", "high")
  })

  collect(document.body?.innerText ?? "", "medium")

  return results
}
