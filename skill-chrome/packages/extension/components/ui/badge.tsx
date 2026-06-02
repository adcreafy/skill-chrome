import type { ReactNode } from "react"

type BadgeVariant = "default" | "success" | "error" | "muted"

interface BadgeProps {
  variant?: BadgeVariant
  children: ReactNode
}

const variants: Record<BadgeVariant, string> = {
  default: "bg-hairline-soft text-body-strong",
  success: "bg-green-50 text-success",
  error: "bg-red-50 text-error",
  muted: "bg-surface-strong text-muted",
}

export function Badge({ variant = "default", children }: BadgeProps) {
  return (
    <span
      className={`inline-flex items-center px-2 py-0.5 rounded-pill text-caption font-medium ${variants[variant]}`}>
      {children}
    </span>
  )
}
