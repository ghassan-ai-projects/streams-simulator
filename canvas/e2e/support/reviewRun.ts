import type { Locator, Page } from "@playwright/test";
import { expect } from "@playwright/test";
import { panelOf } from "./jobControls";

const plant = "/#/plant?domain=domains%2Ffactory-utilities.domain.json&launch=1";

export interface FastRun {
  readonly minutes: number;
  readonly fault?: { readonly at: number; readonly until: number };
  readonly call?: { readonly at: number; readonly setpoint: number };
}

async function addFastPlan(panel: Locator, run: FastRun): Promise<void> {
  await panel.getByLabel("Seed", { exact: true }).fill("5");
  await panel.getByLabel("Duration", { exact: true }).fill(String(run.minutes));
  await panel.getByRole("radio", { name: /As fast as possible/ }).check();
  if (run.fault !== undefined) {
    await panel.getByRole("button", { name: "Add fault" }).click();
    await panel.getByLabel("Fault 1 instance").selectOption("CT-1");
    await panel.getByLabel("Fault 1 fault").selectOption({ label: "fan failure" });
    await panel.getByLabel("Fault 1 starts at minute").fill(String(run.fault.at));
    await panel.getByLabel("Fault 1 ends at minute").fill(String(run.fault.until));
  }
  if (run.call === undefined) return;
  await panel.getByRole("button", { name: "Add effector call" }).click();
  await panel.getByLabel("Effector call 1 instance").selectOption("CH-1");
  await panel.getByLabel("Effector call 1 effector").selectOption({ label: "set setpoint" });
  await panel.getByLabel("Effector call 1 setpoint c").fill(String(run.call.setpoint));
  await panel.getByLabel("Effector call 1 starts at minute").fill(String(run.call.at));
}

export async function reviewOfFastRun(page: Page, run: FastRun): Promise<Locator> {
  await page.goto(plant);
  const panel = panelOf(page);
  await addFastPlan(panel, run);
  await panel.getByRole("button", { name: "Start the run" }).click();
  await panel.getByRole("button", { name: "Open the review" }).click({ timeout: 60_000 });
  await expect(page.getByRole("heading", { level: 1, name: /^Run w-/ })).toBeVisible();
  const details = page.getByRole("complementary", { name: "Run details" });
  await expect(details.getByRole("region", { name: "This run" })).toBeVisible({ timeout: 60_000 });
  return details;
}
