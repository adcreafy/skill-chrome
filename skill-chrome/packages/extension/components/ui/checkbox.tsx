import { Check } from "lucide-react"

interface CheckboxProps {
  checked: boolean
  onChange: (checked: boolean) => void
  label?: string
  disabled?: boolean
}

export function Checkbox({
  checked,
  onChange,
  label,
  disabled = false,
}: CheckboxProps) {
  return (
    <label
      className={`inline-flex items-center gap-2.5 cursor-pointer select-none ${
        disabled ? "opacity-40 cursor-not-allowed" : ""
      }`}>
      <button
        type="button"
        role="checkbox"
        aria-checked={checked}
        disabled={disabled}
        onClick={() => onChange(!checked)}
        className={`flex items-center justify-center w-[18px] h-[18px] rounded-xs border transition-colors duration-150 ${
          checked
            ? "bg-primary border-primary"
            : "bg-surface-card border-hairline-strong hover:border-muted"
        }`}>
        {checked && <Check className="h-3.5 w-3.5 text-on-primary" />}
      </button>
      {label && <span className="text-body-sm text-body">{label}</span>}
    </label>
  )
}
