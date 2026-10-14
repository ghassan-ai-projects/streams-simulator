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
  await panel.getByLabel("Fault 1 minute").fill("2");
  expect((await new AxeBuilder({ page }).include("aside").analyze()).violations).toEqual([]);
  await panel.getByRole("button", { name: "Start the run" }).click();
  const header = page.getByRole("heading", { level: 1, name: /^w-/ });
  await expect(header).toBeVisible({ timeout: 60_000 });
  await expect(page.getByText("Explore · unblinded · ended")).toBeVisible();
  await expect(page.getByText(/· seed 11 ·/)).toBeVisible();
  const slider = page.getByRole("slider", { name: "Playback time" });
  await expect(slider).toHaveAttribute("aria-valuetext", "00:05:00", { timeout: 60_000 });
});

test("a plan with problems cannot start", async ({ page }) => {
  await page.goto(`${env("CANVAS_APP_URL")}/#/plant?domain=domains%2Ffactory-utilities.domain.json`);
  await page.getByRole("button", { name: "Run plant" }).click();
  const panel = page.getByRole("complementary", { name: "Run Factory utilities" });
  await panel.getByRole("button", { name: "Add fault" }).click();
  await panel.getByLabel("Fault 1 instance").selectOption("CT-1");
  await expect(panel.getByRole("button", { name: "Start the run" })).toHaveAttribute("aria-disabled", "true");
  await expect(panel.getByText("Fault 1 needs an instance and a fault.")).toBeVisible();
  await panel.getByRole("button", { name: "Close the run panel" }).click();
  await expect(page.getByRole("complementary", { name: "Inspector" })).toBeVisible();
});
