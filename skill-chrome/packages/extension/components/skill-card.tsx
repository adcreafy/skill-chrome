import { ExternalLink } from "lucide-react"
import { Checkbox } from "~components/ui/checkbox"
import type { DetectedSkill } from "~types"

interface SkillCardProps {
  skill: DetectedSkill
  selected: boolean
  onToggle: (skill: DetectedSkill) => void
}

export function SkillCard({ skill, selected, onToggle }: SkillCardProps) {
  return (
    <div
      className={`flex items-start gap-3 py-3 px-3 rounded-md border transition-colors ${
        selected
          ? "border-primary/30 bg-surface-card"
          : "border-hairline-soft bg-surface-card"
      }`}>
      <div className="pt-0.5">
        <Checkbox checked={selected} onChange={() => onToggle(skill)} />
      </div>
      <div className="flex-1 min-w-0">
        <p className="text-body-sm text-body-strong truncate">{skill.name}</p>
        {skill.description && (
          <p className="text-caption text-muted line-clamp-2 mt-0.5">
            {skill.description}
          </p>
        )}
        {skill.pageUrl && (
          <a
            href={skill.pageUrl}
            target="_blank"
            rel="noopener noreferrer"
            className="inline-flex items-center gap-1 text-caption text-muted hover:text-body mt-1.5 transition-colors"
            onClick={(e) => e.stopPropagation()}>
            <ExternalLink className="w-3 h-3" />
            View page
          </a>
        )}
      </div>
    </div>
  )
}
