import { describe, it, expect, vi, beforeEach } from "vitest"

const mockSendNativeMessage = vi.fn()
const mockLastError = { message: "" }

vi.stubGlobal("chrome", {
  runtime: {
    sendNativeMessage: mockSendNativeMessage,
    get lastError() {
      return mockLastError.message ? { message: mockLastError.message } : null
    },
  },
})

import { ping, detectAgents, installSkills } from "~services/native-client"

beforeEach(() => {
  mockSendNativeMessage.mockReset()
  mockLastError.message = ""
})

describe("ping", () => {
  it("returns true when host responds ok", async () => {
    mockSendNativeMessage.mockImplementation((_host, _msg, cb) => {
      cb({ ok: true, status: "ready" })
    })
    expect(await ping()).toBe(true)
    expect(mockSendNativeMessage).toHaveBeenCalledWith(
      "com.nicepkg.skill_chrome",
      { action: "ping" },
      expect.any(Function),
    )
  })

  it("returns false on runtime error", async () => {
    mockSendNativeMessage.mockImplementation((_host, _msg, cb) => {
      mockLastError.message = "Native host has exited"
      cb(undefined)
      mockLastError.message = ""
    })
    expect(await ping()).toBe(false)
  })
})

describe("detectAgents", () => {
  it("returns engines array", async () => {
    const engines = [
      { id: "cursor", name: "Cursor", detected: true, skillsPath: "/home/.cursor/skills", apps: [] },
      { id: "claude-code", name: "Claude Code", detected: false, skillsPath: "/home/.claude/skills", apps: [] },
    ]
    mockSendNativeMessage.mockImplementation((_host, _msg, cb) => {
      cb({ ok: true, engines })
    })

    const result = await detectAgents()
    expect(result).toHaveLength(2)
    expect(result[0].id).toBe("cursor")
    expect(result[0].detected).toBe(true)
  })

  it("throws on error response", async () => {
    mockSendNativeMessage.mockImplementation((_host, _msg, cb) => {
      cb({ ok: false, error: "cannot determine home" })
    })
    await expect(detectAgents()).rejects.toThrow("cannot determine home")
  })
})

describe("installSkills", () => {
  it("returns install results on success", async () => {
    const responseData = {
      ok: true,
      succeeded: [{ skillName: "test-skill", engines: ["cursor"], fileCount: 2 }],
      failed: [],
    }
    mockSendNativeMessage.mockImplementation((_host, _msg, cb) => {
      cb(responseData)
    })

    const result = await installSkills(["owner/repo"], ["cursor"])
    expect(result.succeeded).toHaveLength(1)
    expect(result.succeeded[0].skillName).toBe("test-skill")
    expect(result.failed).toHaveLength(0)

    expect(mockSendNativeMessage).toHaveBeenCalledWith(
      "com.nicepkg.skill_chrome",
      { action: "install_skills", payload: { sources: ["owner/repo"], agentIds: ["cursor"] } },
      expect.any(Function),
    )
  })

  it("throws when no sources", async () => {
    mockSendNativeMessage.mockImplementation((_host, _msg, cb) => {
      cb({ ok: false, error: "no sources provided" })
    })
    await expect(installSkills([], ["cursor"])).rejects.toThrow("no sources provided")
  })
})
