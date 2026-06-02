import { defineConfig } from "vitest/config"
import { resolve } from "path"

export default defineConfig({
  resolve: {
    alias: {
      "~types": resolve(__dirname, "types/index.ts"),
      "~config": resolve(__dirname, "config"),
      "~services": resolve(__dirname, "services"),
      "~components": resolve(__dirname, "components"),
      "~lib": resolve(__dirname, "lib"),
    },
  },
  test: {
    globals: true,
    environment: "node",
  },
})
