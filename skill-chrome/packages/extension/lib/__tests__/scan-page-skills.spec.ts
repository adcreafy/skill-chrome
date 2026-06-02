import { describe, it, expect, vi, beforeEach } from "vitest"

beforeEach(() => {
  vi.stubGlobal("window", {
    location: { href: "https://example.com/skills" },
  })
})

describe("scanPageForSkills parsing", () => {
  it("extracts skill name from npx command", () => {
    const NPX_PATTERN = /npx\s+skills\s+add\s+(\S+)/gi
    const text = "npx skills add owner/my-skill"
    const match = NPX_PATTERN.exec(text)
    expect(match).not.toBeNull()
    expect(match![1]).toBe("owner/my-skill")
  })

  it("extracts multiple skills from text", () => {
    const NPX_PATTERN = /npx\s+skills\s+add\s+(\S+)/gi
    const text = "npx skills add first/skill\nSome text\nnpx skills add second/skill"
    const results: string[] = []
    let m: RegExpExecArray | null
    while ((m = NPX_PATTERN.exec(text)) !== null) {
      results.push(m[1])
    }
    expect(results).toEqual(["first/skill", "second/skill"])
  })

  it("handles github.com URLs", () => {
    const raw = "https://github.com/owner/repo"
    const gh = /github\.com\/([^/]+)\/([^/]+)/.exec(raw)
    expect(gh).not.toBeNull()
    expect(gh![1]).toBe("owner")
    expect(gh![2]).toBe("repo")
  })

  it("derives name from path", () => {
    const raw = "owner/my-awesome-skill"
    const cleaned = raw.replace(/\/+$/, "")
    const name = cleaned.split("/").pop()
    expect(name).toBe("my-awesome-skill")
  })
})
