import type { Page } from "@playwright/test";
import { expect, test } from "./support/test";

test.setTimeout(120_000);

const plant = "/#/plant?domain=domains%2Ffactory-utilities.domain.json&launch=1";
const fanFailure = /fan[ _]failure|Fault ·/i;

function holdJournal(page: Page): Promise<() => void> {
  return new Promise((ready) => {
    let release = (): void => undefined;
    const gate = new Promise<void>((resolve) => {
      release = resolve;
    });
    void page.route(/\/api\/runs\/[^/]+\/events/, async (route) => {
      await gate;
      await route.continue();
    });
    ready(release);
  });
}

async function startRun(page: Page, mode: "Explore" | "Evaluate") {
  await page.goto(plant);
  const panel = page.getByRole("complementary", { name: "Run Factory utilities" });
  await panel.getByLabel("Seed", { exact: true }).fill("11");
  await panel.getByLabel("Duration", { exact: true }).fill("60");
  await panel.getByRole("radio", { name: new RegExp(`^${mode}`) }).check();
  await panel.getByRole("button", { name: "Add fault" }).click();
  await panel.getByLabel("Fault 1 instance").selectOption("CT-1");
  await panel.getByLabel("Fault 1 fault").selectOption({ label: "fan failure" });
  await panel.getByLabel("Fault 1 starts at minute").fill("5");
  await panel.getByRole("button", { name: "Start the run" }).click();
  await expect(panel.getByText(/^Running Factory utilities as run w-/)).toBeVisible();
  await panel.getByRole("button", { name: "Close the run panel" }).click();
}

async function inspect(page: Page, instance: RegExp) {
  await page.getByRole("navigation", { name: "Outline" }).getByRole("button", { name: instance }).click();
  return page.getByRole("region", { name: "Observed readings" });
}

test("an Explore run shows what an instance reports now: value, unit, when, how old and its trend", async ({
  page,
}) => {
  const release = await holdJournal(page);
  await startRun(page, "Explore");
  const observed = await inspect(page, /CT-1/);
  await expect(observed.getByText("Awaiting first reading")).toHaveCount(2);
  release();
  const row = observed.getByRole("row", { name: /basin temp/ });
  await expect(row).toContainText(/\d+(\.\d+)?\s*°C/);
  await expect(row.locator("[data-part='observed-at']")).toHaveText(/^\d\d:\d\d:\d\d$/);
  await expect(row.locator("[data-part='age']")).toHaveText(/^(just now|\d+ (s|min|h)( \d+ min)? old)$/);
  await expect(row.locator("[data-part='trend']")).toBeVisible();
  await expect(observed.getByText("Awaiting first reading")).toHaveCount(0);
  await expect(page.getByRole("region", { name: "Channels" }).locator("details")).not.toHaveAttribute("open");
  await expect(page.getByRole("group", { name: /^CT-1, / })).toContainText("Fault · fan failure", { timeout: 30_000 });
});

test("an Evaluate run shows the same readings and nothing of the planned fault anywhere @slow", async ({ page }) => {
  await startRun(page, "Evaluate");
  const observed = await inspect(page, /CH-1/);
  const row = observed.getByRole("row", { name: /power/ }).first();
  await expect(row).toContainText(/\d+(\.\d+)?\s*kW/, { timeout: 30_000 });
  await expect(row.locator("[data-part='trend']")).toBeVisible();
  await expect(page.getByText(/^Running · 00:(0[6-9]|1\d)/)).toBeVisible({ timeout: 30_000 });
  const text = await page.locator("body").innerText();
  const attributes = await page.evaluate(() =>
    [...document.body.querySelectorAll("*")].flatMap((element) => [...element.attributes].map((a) => a.value)),
  );
  const names = await page.locator("body").ariaSnapshot();
  expect(text).not.toMatch(fanFailure);
  expect(attributes.filter((value) => fanFailure.test(value))).toEqual([]);
  expect(names).not.toMatch(fanFailure);
  await expect(observed).not.toContainText(/fault|broken|failed/i);
});

function cellsOverflowing(): string[] {
  const table = document.querySelector("[data-part='observed']");
  if (table === null) return ["no table"];
  const spill: string[] = [];
  for (const cell of table.querySelectorAll("tbody th, tbody td")) {
    const box = cell.getBoundingClientRect();
    if (cell.scrollWidth > cell.clientWidth + 1) spill.push(`${cell.textContent} clipped`);
    for (const part of cell.querySelectorAll("[data-part], svg")) {
      const rect = part.getBoundingClientRect();
      if (rect.left < box.left - 0.5 || rect.right > box.right + 0.5)
        spill.push(`${part.getAttribute("data-part") ?? "svg"} outside`);
    }
  }
  return spill;
}

function widestLine(): number {
  const trend = document.querySelector("[data-part='trend']");
  const lines = [...(trend?.querySelectorAll("polyline") ?? [])];
  const used = lines.reduce((width, line) => Math.max(width, line.getBoundingClientRect().width), 0);
  return used / (trend?.getBoundingClientRect().width ?? 1);
}

test("the observed table keeps every value inside its cell, labels whole, column labels visible and the trend across its cell", async ({
  page,
}) => {
  await startRun(page, "Explore");
  const observed = await inspect(page, /CT-1/);
  await expect(page.getByText(/^Running · 00:(0[6-9]|[1-5]\d)/)).toBeVisible({ timeout: 60_000 });
  await expect(observed.locator("[data-part='trend']").first()).toBeVisible();
  expect(await page.evaluate(cellsOverflowing)).toEqual([]);
  await expect(observed.getByRole("columnheader", { name: "Observed at" })).toBeVisible();
  await expect(observed.getByRole("columnheader", { name: "Trend" })).toBeVisible();
  expect(await page.evaluate(widestLine)).toBeGreaterThan(0.5);
});
