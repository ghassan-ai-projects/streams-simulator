import AxeBuilder from "@axe-core/playwright";
import { expect, test } from "@playwright/test";
import { env } from "./support/env";

test.use({ storageState: env("CANVAS_APP_STATE") });
test.setTimeout(90_000);

test("a plant runs with a planned fault and the finished run opens in the review", async ({ page }) => {
  await page.goto(`${env("CANVAS_APP_URL")}/#/plant?domain=domains%2Ffactory-utilities.domain.json`);
  await page.getByRole("button", { name: "Run plant" }).click();
  const panel = page.getByRole("complementary", { name: "Run Factory utilities" });
  await panel.getByLabel("Seed").fill("11");
  await panel.getByLabel("Duration", { exact: true }).fill("5");
  await panel.getByRole("button", { name: "Add fault" }).click();
  await panel.getByLabel("Fault 1 instance").selectOption("CT-1");
  await panel.getByLabel("Fault 1 fault").selectOption({ label: "fan failure" });
  await panel.getByLabel("Fault 1 starts at minute").fill("2");
  expect((await new AxeBuilder({ page }).include("aside").analyze()).violations).toEqual([]);
  await panel.getByRole("button", { name: "Start the run" }).click();
  const header = page.getByRole("heading", { level: 1, name: /^Run w-/ });
  await expect(header).toBeVisible({ timeout: 60_000 });
  await expect(page.getByText("Explore", { exact: true })).toBeVisible();
  await expect(page.getByText(/^seed 11 · /)).toBeVisible();
  const slider = page.getByRole("slider", { name: "Playback time" });
  await expect(slider).toHaveAttribute("max", "300", { timeout: 60_000 });
  const disturbances = page
    .getByRole("complementary", { name: "Run details" })
    .getByRole("region", { name: "Disturbances" });
  await expect(disturbances).toContainText("00:02:00");
  await expect(disturbances).toContainText("CT-1");
  await expect(disturbances).toContainText("fan failure");
});

test("a plan with problems cannot start", async ({ page }) => {
  await page.goto(`${env("CANVAS_APP_URL")}/#/plant?domain=domains%2Ffactory-utilities.domain.json`);
  await page.getByRole("button", { name: "Run plant" }).click();
  const panel = page.getByRole("complementary", { name: "Run Factory utilities" });
  await panel.getByRole("button", { name: "Add fault" }).click();
  await panel.getByLabel("Fault 1 instance").selectOption("CT-1");
  await panel.getByRole("button", { name: "Start the run" }).click();
  await expect(panel.getByRole("list", { name: "Fix before starting" })).toBeFocused();
  await expect(panel.getByText("Fault 1 needs an instance and a fault.")).toBeVisible();
  await panel.getByRole("button", { name: "Close the run panel" }).click();
  await expect(page.getByRole("complementary", { name: "Inspector" })).toBeVisible();
  await expect(page.getByRole("button", { name: "Run plant" })).toBeFocused();
});

test("opening the run panel shows the details panel even when it was hidden", async ({ page }) => {
  await page.goto(`${env("CANVAS_APP_URL")}/#/plant?domain=domains%2Ffactory-utilities.domain.json`);
  await page.getByRole("button", { name: "Details panel" }).click();
  await expect(page.getByRole("complementary", { name: "Inspector" })).toBeHidden();
  await page.getByRole("button", { name: "Run plant" }).click();
  await expect(page.getByRole("complementary", { name: "Run Factory utilities" })).toBeVisible();
  await expect(page.getByRole("button", { name: "Details panel" })).toHaveAttribute("aria-pressed", "true");
});
