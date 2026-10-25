import type { Locator, Page } from "@playwright/test";
import { expect, test } from "./support/test";
import { panelOf } from "./support/jobControls";

test.setTimeout(120_000);

const plant = "/#/plant?domain=domains%2Ffactory-utilities.domain.json&launch=1";

async function planExplore(panel: Locator): Promise<void> {
  await panel.getByLabel("Seed", { exact: true }).fill("5");
  await panel.getByLabel("Duration", { exact: true }).fill("60");
  await panel.getByRole("radio", { name: /^Explore/ }).check();
  await panel.getByRole("button", { name: "Add fault" }).click();
  await panel.getByLabel("Fault 1 instance").selectOption("CT-1");
  await panel.getByLabel("Fault 1 fault").selectOption({ label: "fan failure" });
  await panel.getByLabel("Fault 1 starts at minute").fill("2");
  await panel.getByLabel("Fault 1 ends at minute").fill("9");
  await panel.getByRole("button", { name: "Add effector call" }).click();
  await panel.getByLabel("Effector call 1 instance").selectOption("CH-1");
  await panel.getByLabel("Effector call 1 effector").selectOption({ label: "set setpoint" });
  await panel.getByLabel("Effector call 1 setpoint c").fill("7");
  await panel.getByLabel("Effector call 1 starts at minute").fill("1");
}

async function actNow(panel: Locator): Promise<void> {
  await panel.getByLabel("Call now instance").selectOption("CH-1");
  await panel.getByLabel("Call now effector").selectOption({ label: "set setpoint" });
  await panel.getByLabel("Call now setpoint c").fill("8");
  await panel.getByRole("button", { name: "Call now" }).click();
  await panel.getByLabel("Fault now instance").selectOption("CH-1");
  await panel.getByLabel("Fault now fault").selectOption({ label: "condenser fouling" });
  await panel.getByRole("button", { name: "Inject now" }).click();
  await expect(panel.getByRole("list", { name: "Commands given now" })).toContainText("fault injected");
}

async function pausedAfterTheRecovery(page: Page): Promise<Locator> {
  await page.goto(plant);
  const panel = panelOf(page);
  await planExplore(panel);
  await panel.getByRole("button", { name: "Start the run" }).click();
  await expect(panel.getByText(/^Running · 00:(1\d|[2-5]\d)/)).toBeVisible({ timeout: 60_000 });
  await panel.getByRole("button", { name: "Pause" }).click();
  await actNow(panel);
  return panel;
}

async function expectTheRunAsPlanned(page: Page, panel: Locator): Promise<void> {
  await expect(panel.getByText("1 h of simulated time · Explore · 1 fault · 1 effector call")).toBeVisible();
  await expect(panel.getByTitle("CH-1 set setpoint at 00:01:00")).toHaveCount(1);
  await expect(panel.getByTitle("CT-1 fan failure at 00:02:00")).toHaveCount(1);
  const commands = panel.getByRole("list", { name: "Commands given now" });
  await expect(commands).toContainText(/CH-1 · set setpoint: accepted · now/);
  await expect(commands).toContainText(/CH-1 · fault injected: condenser fouling · now/);
  await expect(page.getByRole("group", { name: /^CH-1, Fault: condenser fouling/ })).toBeVisible();
  await expect(page.getByRole("group", { name: /^CT-1, Fault/ })).toHaveCount(0);
}

test("an Explore run found again after a reload shows its plan, its commands given now and what has recovered", async ({
  page,
}) => {
  const panel = await pausedAfterTheRecovery(page);
  await expectTheRunAsPlanned(page, panel);
  await page.reload();
  const banner = page.getByRole("button", { name: /^Paused Factory utilities · / });
  await expect(banner).toBeVisible();
  await banner.click();
  await expectTheRunAsPlanned(page, panelOf(page));
});
