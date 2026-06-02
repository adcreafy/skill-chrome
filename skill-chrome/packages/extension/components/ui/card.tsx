import type { HTMLAttributes, ReactNode } from "react"

interface CardProps extends HTMLAttributes<HTMLDivElement> {
  children: ReactNode
  padding?: "none" | "sm" | "md" | "lg"
}

const paddings = {
  none: "",
  sm: "p-3",
  md: "p-4",
  lg: "p-5",
}

export function Card({
  children,
  padding = "md",
  className = "",
  ...rest
}: CardProps) {
  return (
    <div
      className={`bg-surface-card rounded-lg border border-hairline-soft shadow-card ${paddings[padding]} ${className}`}
      {...rest}>
      {children}
    </div>
  )
}
