import type { DetectedEngine, InstallResult } from "~types"

const HOST_NAME = "com.nicepkg.skill_chrome"

interface NativeResponse {
  ok: boolean
  error?: string
  [key: string]: unknown
}

async function sendToHost<T extends NativeResponse>(
  action: string,
  payload?: unknown,
): Promise<T> {
  return new Promise((resolve, reject) => {
    const message: Record<string, unknown> = { action }
    if (payload !== undefined) {
      message.payload = payload
    }
    chrome.runtime.sendNativeMessage(HOST_NAME, message, (response) => {
      if (chrome.runtime.lastError) {
        reject(new Error(chrome.runtime.lastError.message))
        return
      }
      resolve(response as T)
    })
  })
}

export async function ping(): Promise<boolean> {
  try {
    const resp = await sendToHost<NativeResponse>("ping")
    return resp.ok === true
  } catch {
    return false
  }
}

export async function detectAgents(): Promise<DetectedEngine[]> {
  const resp = await sendToHost<NativeResponse & { engines?: DetectedEngine[] }>(
    "detect_agents",
  )
  if (!resp.ok) {
    throw new Error(resp.error ?? "detect_agents failed")
  }
  return resp.engines ?? []
}

export async function installSkills(
  sources: string[],
  agentIds: string[],
): Promise<InstallResult> {
  const resp = await sendToHost<
    NativeResponse & { succeeded?: InstallResult["succeeded"]; failed?: InstallResult["failed"] }
  >("install_skills", { sources, agentIds })
  if (!resp.ok) {
    throw new Error(resp.error ?? "install_skills failed")
  }
  return {
    succeeded: resp.succeeded ?? [],
    failed: resp.failed ?? [],
  }
}
