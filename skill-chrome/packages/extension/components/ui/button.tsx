import type { ButtonHTMLAttributes, ReactNode } from "react"

type Variant = "primary" | "secondary" | "ghost"
type Size = "sm" | "md" | "lg"

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: Variant
  size?: Size
  loading?: boolean
  icon?: ReactNode
}

const variants: Record<Variant, string> = {
  primary: "bg-primary text-on-primary hover:bg-primary-active",
  secondary:
    "bg-surface-card text-body-strong border border-hairline hover:bg-hairline-soft",
  ghost: "bg-transparent text-body hover:bg-hairline-soft",
}

const sizes: Record<Size, string> = {
  sm: "h-8 px-3 text-caption gap-1.5",
  md: "h-10 px-4 text-button gap-2",
  lg: "h-11 px-5 text-button gap-2",
}

export function Button({
  variant = "primary",
  size = "md",
  loading = false,
  icon,
  children,
  className = "",
  disabled,
  ...rest
}: ButtonProps) {
  return (
    <button
      className={`
        inline-flex items-center justify-center rounded-md font-medium
        transition-colors duration-150
        disabled:opacity-40 disabled:cursor-not-allowed
        ${variants[variant]} ${sizes[size]} ${className}
      `}
      disabled={disabled || loading}
      {...rest}>
      {loading ? (
        <span className="h-4 w-4 animate-spin rounded-full border-2 border-current border-t-transparent" />
      ) : icon ? (
        <span className="shrink-0">{icon}</span>
      ) : null}
      {children}
    </button>
  )
}
