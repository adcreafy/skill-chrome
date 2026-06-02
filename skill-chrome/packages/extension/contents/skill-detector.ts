import type { PlasmoCSConfig } from "plasmo"
import { scanPageForSkills } from "~lib/scan-page-skills"

export const config: PlasmoCSConfig = {
  matches: ["https://*/*", "http://*/*"],
  run_at: "document_idle",
}

function publish(skills: ReturnType<typeof scanPageForSkills>) {
  for (const payload of skills) {
    chrome.runtime.sendMessage({ type: "SKILL_DETECTED", payload })
  }
}

function runDetect() {
  const skills = scanPageForSkills()
  publish(skills)
  return skills
}

// Initial + SPA updates
runDetect()

let debounceTimer: ReturnType<typeof setTimeout> | null = null
const observer = new MutationObserver(() => {
  if (debounceTimer) clearTimeout(debounceTimer)
  debounceTimer = setTimeout(runDetect, 300)
})

if (document.body) {
  observer.observe(document.body, { childList: true, subtree: true })
}

chrome.runtime.onMessage.addListener((msg, _sender, sendResponse) => {
  if (msg.type === "RUN_DETECT") {
    sendResponse({ skills: runDetect() })
    return true
  }
})
