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

test("a run still running lists as Running, never Did not finish, and stops reading Running once it ends", async ({
  page,
}) => {
  const world = await startWatchedRun(page);
  expect(world).not.toBe("");
  await page.getByRole("button", { name: "Streams Simulator, workspace home" }).click();
  const row = rowOf(page, world);
  await expect(row.getByText("Running", { exact: true })).toBeVisible();
  await expect(row).not.toContainText("Did not finish");
  await page.getByRole("button", { name: /^Running Factory utilities · / }).click();
  await page
    .getByRole("complementary", { name: "Run Factory utilities" })
    .getByRole("button", { name: "Stop now" })
    .click();
  await expect(page.getByRole("heading", { level: 1, name: new RegExp(`^Run ${world}`) })).toBeVisible({
    timeout: 60_000,
  });
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
