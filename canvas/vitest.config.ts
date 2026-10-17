import react from "@vitejs/plugin-react";
import { defineConfig } from "vitest/config";

export default defineConfig({
  plugins: [react()],
  test: {
    environment: "jsdom",
    include: ["src/**/*.test.ts", "src/**/*.test.tsx"],
    setupFiles: ["src/test/setup.ts"],
    restoreMocks: true,
    coverage: {
      provider: "v8",
      include: ["src/**/*.{ts,tsx}"],
      exclude: ["src/**/*.test.{ts,tsx}", "src/test/**", "src/main.tsx"],
      thresholds: {
        "src/model/**": { lines: 90, functions: 90, branches: 85, statements: 90 },
        "src/app/**": { lines: 80, functions: 80, branches: 75, statements: 80 },
        "src/adapters/**": { lines: 99, functions: 99, branches: 97, statements: 99 },
      },
    },
  },
});
