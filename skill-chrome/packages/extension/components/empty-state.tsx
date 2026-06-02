import { Compass, RefreshCw } from "lucide-react"

interface EmptyStateProps {
  onRescan?: () => void
}

export function EmptyState({ onRescan }: EmptyStateProps) {
  return (
    <div className="flex flex-col items-center justify-center py-16 px-6 text-center">
      <div className="w-11 h-11 rounded-xl bg-surface-strong flex items-center justify-center mb-4">
        <Compass className="w-5 h-5 text-muted" />
      </div>
      <p className="text-body-sm text-body-strong mb-1">No skills detected</p>
      <p className="text-caption text-muted max-w-[220px] mb-4">
        Open a page with an install command, or rescan this tab.
      </p>
      {onRescan && (
        <button
          type="button"
          onClick={onRescan}
          className="inline-flex items-center gap-2 text-caption text-primary hover:text-primary-active">
          <RefreshCw className="w-3.5 h-3.5" />
          Rescan page
        </button>
      )}
    </div>
  )
}
