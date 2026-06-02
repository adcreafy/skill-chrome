import { useState } from "react"
import { Download, ChevronDown, ChevronUp } from "lucide-react"
import { Button } from "~components/ui/button"
import { Checkbox } from "~components/ui/checkbox"
import { Badge } from "~components/ui/badge"
import type { DetectedEngine } from "~types"

interface BottomBarProps {
  engines: DetectedEngine[]
  selectedEngineIds: Set<string>
  selectedSkillCount: number
  installing: boolean
  onToggleEngine: (id: string) => void
  onInstall: () => void
}

export function BottomBar({
  engines,
  selectedEngineIds,
  selectedSkillCount,
  installing,
  onToggleEngine,
  onInstall,
}: BottomBarProps) {
  const [expanded, setExpanded] = useState(true)

  const canInstall =
    selectedSkillCount > 0 && selectedEngineIds.size > 0 && !installing

  return (
    <div className="shrink-0 border-t border-hairline-soft bg-surface-card px-4 py-3 safe-area-pb">
      <button
        type="button"
        onClick={() => setExpanded((v) => !v)}
        className="flex items-center justify-between w-full mb-2 text-caption text-muted hover:text-body">
        <span>Install to ({selectedEngineIds.size} selected)</span>
        {expanded ? (
          <ChevronDown className="w-4 h-4" />
        ) : (
          <ChevronUp className="w-4 h-4" />
        )}
      </button>

      {expanded && (
        <div className="mb-3 max-h-[180px] overflow-y-auto flex flex-col gap-1">
          {engines.map((engine) => (
            <div key={engine.id}>
              <div className="flex items-center gap-2 py-1.5 px-2 rounded-md">
                <Checkbox
                  checked={selectedEngineIds.has(engine.id)}
                  onChange={() => onToggleEngine(engine.id)}
                  disabled={installing}
                />
                <span className="text-body-sm text-body-strong flex-1 truncate">
                  {engine.name}
                </span>
                {engine.detected && (
                  <Badge variant="success">detected</Badge>
                )}
              </div>
              {engine.apps.length > 0 && (
                <div className="ml-8 mb-1">
                  {engine.apps.map((app) => (
                    <p
                      key={app.bundlePath}
                      className="text-caption text-muted py-0.5 truncate">
                      includes: {app.name}
                    </p>
                  ))}
                </div>
              )}
            </div>
          ))}
        </div>
      )}

      <Button
        onClick={onInstall}
        disabled={!canInstall}
        loading={installing}
        className="w-full"
        icon={<Download className="w-4 h-4" />}>
        {installing
          ? `Installing ${selectedSkillCount} skill${selectedSkillCount !== 1 ? "s" : ""}...`
          : selectedSkillCount > 0
            ? `Install ${selectedSkillCount} skill${selectedSkillCount !== 1 ? "s" : ""}`
            : "Select skills to install"}
      </Button>
    </div>
  )
}
