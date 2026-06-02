import type { Config } from "tailwindcss"

const config: Config = {
  content: [
    "./sidepanel.tsx",
    "./background.ts",
    "./components/**/*.{tsx,ts}",
    "./services/**/*.{tsx,ts}",
    "./config/**/*.{tsx,ts}",
    "./contents/**/*.{tsx,ts}",
  ],
  theme: {
    extend: {
      colors: {
        canvas: {
          DEFAULT: "#f5f5f5",
          soft: "#fafafa",
          deep: "#0c0a09",
        },
        ink: "#0c0a09",
        body: {
          DEFAULT: "#4e4e4e",
          strong: "#292524",
        },
        muted: {
          DEFAULT: "#777169",
          soft: "#a8a29e",
        },
        hairline: {
          DEFAULT: "#e7e5e4",
          soft: "#f0efed",
          strong: "#d6d3d1",
        },
        surface: {
          card: "#ffffff",
          strong: "#f0efed",
          dark: "#0c0a09",
          "dark-elevated": "#1c1917",
        },
        primary: {
          DEFAULT: "#292524",
          active: "#0c0a09",
        },
        "on-primary": "#ffffff",
        success: "#16a34a",
        error: "#dc2626",
      },
      borderRadius: {
        xs: "4px",
        sm: "6px",
        md: "8px",
        lg: "12px",
        xl: "16px",
        "2xl": "24px",
        pill: "9999px",
      },
      fontFamily: {
        sans: ["Inter", "system-ui", "-apple-system", "sans-serif"],
      },
      fontSize: {
        "title-md": ["20px", { lineHeight: "1.35", fontWeight: "500" }],
        "title-sm": ["18px", { lineHeight: "1.44", fontWeight: "500", letterSpacing: "0.18px" }],
        "body-md": ["16px", { lineHeight: "1.5", fontWeight: "400", letterSpacing: "0.16px" }],
        "body-strong": ["16px", { lineHeight: "1.5", fontWeight: "500", letterSpacing: "0.16px" }],
        "body-sm": ["15px", { lineHeight: "1.47", fontWeight: "400", letterSpacing: "0.15px" }],
        caption: ["14px", { lineHeight: "1.5", fontWeight: "400" }],
        "caption-upper": ["12px", { lineHeight: "1.4", fontWeight: "600", letterSpacing: "0.96px" }],
        button: ["15px", { lineHeight: "1", fontWeight: "500" }],
      },
      spacing: {
        xxs: "4px",
        xs: "8px",
        sm: "12px",
        base: "16px",
        md: "20px",
        lg: "24px",
        xl: "32px",
        "2xl": "48px",
      },
      boxShadow: {
        card: "0 4px 16px rgba(0, 0, 0, 0.04)",
      },
    },
  },
  plugins: [],
}

export default config
