export {}

chrome.sidePanel.setPanelBehavior({ openPanelOnActionClick: true })

const tabSkills = new Map<number, unknown[]>()

function setTabSkills(tabId: number, skills: unknown[]) {
  tabSkills.set(tabId, skills)
  chrome.runtime
    .sendMessage({
      type: "TAB_SKILLS_UPDATED",
      tabId,
      skills,
    })
    .catch(() => {})
}

async function detectOnTab(tabId: number): Promise<unknown[]> {
  try {
    const response = await chrome.tabs.sendMessage(tabId, { type: "RUN_DETECT" })
    if (response?.skills?.length) {
      setTabSkills(tabId, response.skills)
      return response.skills
    }
  } catch {
    // Content script not ready — inject scan function
  }

  try {
    const [result] = await chrome.scripting.executeScript({
      target: { tabId },
      func: () => {
        const NPX = /npx\s+skills\s+add\s+(\S+)/gi
        const visited = new Set<string>()
        const skills: Array<{
          name: string
          source: Record<string, unknown>
          command: string
          pageUrl: string
          detectedAt: number
          confidence: string
        }> = []
        const pageUrl = window.location.href

        const collect = (text: string, confidence: string) => {
          NPX.lastIndex = 0
          let m: RegExpExecArray | null
          while ((m = NPX.exec(text)) !== null) {
            const raw = m[1]
            if (visited.has(raw)) continue
            visited.add(raw)
            const cleaned = raw.replace(/\/+$/, "")
            const name = cleaned.split("/").pop() ?? raw
            let source: Record<string, unknown> = { type: "command", raw }
            const gh = /github\.com\/([^/]+)\/([^/]+)/.exec(raw)
            if (gh) source = { type: "github", owner: gh[1], repo: gh[2] }
            else if (raw.includes("skills.sh"))
              source = { type: "registry", registry: "skills.sh" }
            skills.push({
              name,
              source,
              command: raw,
              pageUrl,
              detectedAt: Date.now(),
              confidence,
            })
          }
        }

        document
          .querySelectorAll(
            "pre, code, [class*='code'], [class*='Code'], [class*='terminal'], [class*='shell']",
          )
          .forEach((el) => collect(el.textContent?.trim() ?? "", "high"))
        collect(document.body?.innerText ?? "", "medium")
        return skills
      },
    })

    const skills = result?.result ?? []
    setTabSkills(tabId, skills)
    return skills
  } catch {
    return tabSkills.get(tabId) ?? []
  }
}

chrome.runtime.onMessage.addListener((msg, sender, sendResponse) => {
  if (msg.type === "SKILL_DETECTED" && sender.tab?.id) {
    const tabId = sender.tab.id
    const existing = [...(tabSkills.get(tabId) ?? [])]
    const skill = msg.payload
    const isDuplicate = existing.some(
      (s: { command?: string }) => s.command === skill.command,
    )
    if (!isDuplicate) {
      existing.push(skill)
      setTabSkills(tabId, existing)
    }
    sendResponse({ ok: true })
  }

  if (msg.type === "GET_TAB_SKILLS") {
    sendResponse({ skills: tabSkills.get(msg.tabId) ?? [] })
    return true
  }

  if (msg.type === "DETECT_ON_TAB") {
    detectOnTab(msg.tabId).then((skills) => sendResponse({ skills }))
    return true
  }

  if (msg.type === "NATIVE_REQUEST") {
    const { action, payload } = msg
    chrome.runtime.sendNativeMessage(
      "com.nicepkg.skill_chrome",
      { action, payload },
      (response) => {
        if (chrome.runtime.lastError) {
          sendResponse({ ok: false, error: chrome.runtime.lastError.message })
        } else {
          sendResponse(response)
        }
      },
    )
    return true
  }

  return true
})

chrome.tabs.onRemoved.addListener((tabId) => {
  tabSkills.delete(tabId)
})

chrome.tabs.onActivated.addListener(async (activeInfo) => {
  const cached = tabSkills.get(activeInfo.tabId) ?? []
  if (cached.length === 0) {
    await detectOnTab(activeInfo.tabId)
  } else {
    setTabSkills(activeInfo.tabId, cached)
  }
})

chrome.tabs.onUpdated.addListener((tabId, changeInfo) => {
  if (changeInfo.url) {
    tabSkills.delete(tabId)
    chrome.tabs.query({ active: true, currentWindow: true }, (tabs) => {
      if (tabs[0]?.id === tabId) {
        detectOnTab(tabId)
      }
    })
  }
})

if (chrome.sidePanel.onOpened) {
  chrome.sidePanel.onOpened.addListener(() => {
    chrome.tabs.query({ active: true, currentWindow: true }, (tabs) => {
      if (tabs[0]?.id) detectOnTab(tabs[0].id)
    })
  })
}
