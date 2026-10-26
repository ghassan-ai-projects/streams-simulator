import type { Locator, Page } from "@playwright/test";
import { expect, test } from "./support/test";

test.setTimeout(90_000);

function rowOf(page: Page, world: string): Locator {
  return page
    .getByRole("region", { name: "Runs" })
    .getByRole("listitem")
    .filter({ hasText: new RegExp(`${world} ·`) });
}

async function startWatchedRun(page: Page): Promise<string> {
  await page.goto("/#/plant?domain=domains%2Ffactory-utilities.domain.json&launch=1");
  const panel = page.getByRole("complementary", { name: "Run Factory utilities" });
  await panel.getByLabel("Duration", { exact: true }).fill("30");
  await panel.getByRole("button", { name: "Start the run" }).click();
  const started = panel.getByText(/^Running Factory utilities as run w-/);
  await expect(started).toBeVisible();
  return /run (w-[^\s·.]+)/.exec((await started.textContent()) ?? "")?.[1] ?? "";
}

async function expectLegibleIcon(pill: Locator): Promise<void> {
  const inside = await pill.evaluate((element) => {
    const box = element.getBoundingClientRect();
    const mark = element.querySelector("svg")?.getBoundingClientRect();
    const gap = Number.parseFloat(getComputedStyle(element).columnGap);
    if (mark === undefined) return null;
    return { size: mark.width, left: mark.left - box.left, spill: mark.bottom - box.bottom, gap };
  });
  expect(inside).not.toBeNull();
  expect(inside?.size).toBeGreaterThanOrEqual(14);
  expect(inside?.left).toBeGreaterThan(4);
  expect(inside?.spill).toBeLessThanOrEqual(0);
  expect(inside?.gap).toBeGreaterThanOrEqual(6);
}

test("a run still running lists as Running, never Did not finish, and stops reading Running once it ends", async ({
  page,
}) => {
  const world = await startWatchedRun(page);
  expect(world).not.toBe("");
  await page.getByRole("button", { name: "Streams Simulator, workspace home" }).click();
  const row = rowOf(page, world);
  const pill = row.getByText("Running", { exact: true });
  await expect(pill).toBeVisible();
  await expect(row).not.toContainText("Did not finish");
  await expectLegibleIcon(pill);
  await page.getByRole("button", { name: /^Running Factory utilities · / }).click();
  const panel = page.getByRole("complementary", { name: "Run Factory utilities" });
  await panel.getByRole("button", { name: "Stop now" }).click();
  await panel.getByRole("button", { name: "Open the review" }).click({ timeout: 60_000 });
  await expect(page.getByRole("heading", { level: 1, name: new RegExp(`^Run ${world}`) })).toBeVisible();
  await page.getByRole("button", { name: "Streams Simulator, workspace home" }).click();
  await expect(rowOf(page, world)).toBeVisible();
  await expect(rowOf(page, world).getByText("Running", { exact: true })).toHaveCount(0);
  await expect(rowOf(page, world)).not.toContainText("Did not finish");
});

test("opening the review of a run that is still running reads Running in its header, not Did not finish", async ({
  page,
}) => {
  const world = await startWatchedRun(page);
  await page.getByRole("button", { name: "Streams Simulator, workspace home" }).click();
  await rowOf(page, world)
    .getByRole("button", { name: `Review ${world}` })
    .click();
  await expect(page.getByRole("status").filter({ hasText: /^Running$/ })).toHaveCount(1, { timeout: 60_000 });
  await expect(page.getByRole("status").filter({ hasText: /Did not finish/ })).toHaveCount(0);
  await expect(page.getByText("Did not finish", { exact: true })).toHaveCount(0);
});

test("a running Explore job lists with its own name, its plant and its mode, not Unknown plant or Evaluate", async ({
  page,
}) => {
  await page.goto("/#/plant?domain=domains%2Ffactory-utilities.domain.json&launch=1");
  const panel = page.getByRole("complementary", { name: "Run Factory utilities" });
  await panel.getByLabel("Name").fill("Loss and recovery");
  await panel.getByLabel("Duration", { exact: true }).fill("30");
  await panel.getByRole("button", { name: "Start the run" }).click();
  await expect(panel.getByText(/^Running Loss and recovery on Factory utilities as run w-/)).toBeVisible();
  await page.getByRole("button", { name: "Streams Simulator, workspace home" }).click();
  const row = page.getByRole("region", { name: "Runs" }).getByRole("listitem").filter({ hasText: "Running" });
  await expect(row).toContainText("Loss and recovery");
  await expect(row).toContainText("Explore");
  await expect(row).toContainText("Factory utilities");
  await expect(row).not.toContainText("Unknown plant");
  await expect(row).not.toContainText("Evaluate");
});
