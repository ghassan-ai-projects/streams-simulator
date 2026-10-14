import { defineConfig, devices } from "@playwright/test";

const ci = process.env["CI"] === "true";

export default defineConfig({
  testDir: "e2e",
  globalSetup: "./e2e/global-setup.ts",
  fullyParallel: true,
  forbidOnly: true,
  retries: 0,
  timeout: 30_000,
  reporter: ci ? "github" : "list",
  use: { trace: "retain-on-failure" },
  projects: [
    {
      name: "desktop-chrome",
      use: {
        ...devices["Desktop Chrome"],
        viewport: { width: 1440, height: 900 },
        ...(ci ? {} : { channel: "chrome" }),
      },
    },
  ],
});
