import { defineConfig, devices } from "@playwright/test";

const ci = process.env["CI"] === "true";

export default defineConfig({
  testDir: "e2e",
  fullyParallel: true,
  forbidOnly: true,
  retries: 0,
  reporter: ci ? "github" : "list",
  use: { baseURL: "http://127.0.0.1:4173", trace: "retain-on-failure" },
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
  webServer: { command: "npm run preview", url: "http://127.0.0.1:4173", reuseExistingServer: !ci },
});
