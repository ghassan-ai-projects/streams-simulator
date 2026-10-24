import type { Locator, Page } from "@playwright/test";
import { expect } from "@playwright/test";

const plant = "/#/plant?domain=domains%2Ffactory-utilities.domain.json&launch=1";

export type Mode = "Explore" | "Evaluate";

export function panelOf(page: Page): Locator {
  return page.getByRole("complementary", { name: "Run Factory utilities" });
}

export async function startOpen(page: Page, mode: Mode, minutes = 60): Promise<Locator> {
  await page.goto(plant);
  const panel = panelOf(page);
  await panel.getByLabel("Seed", { exact: true }).fill("5");
  await panel.getByLabel("Duration", { exact: true }).fill(String(minutes));
  await panel.getByRole("radio", { name: new RegExp(`^${mode}`) }).check();
  await panel.getByRole("button", { name: "Start the run" }).click();
  await expect(panel.getByText(/^Running Factory utilities as run w-/)).toBeVisible();
  return panel;
}

function clockLine(panel: Locator): Locator {
  return panel.getByText(/^(Running|Paused) · \d\d:\d\d:\d\d of \d\d:\d\d:\d\d simulated/);
}

export async function simulatedSeconds(panel: Locator): Promise<number> {
  const text = (await clockLine(panel).textContent()) ?? "";
  const [h = 0, m = 0, s = 0] = (/(\d\d):(\d\d):(\d\d) of/.exec(text)?.slice(1) ?? []).map(Number);
  return h * 3600 + m * 60 + s;
}

export function clockWords(panel: Locator): Locator {
  return clockLine(panel);
}

export async function sameAcrossPolls(page: Page, read: () => Promise<number>, polls = 3): Promise<number[]> {
  const seen: number[] = [];
  for (let poll = 0; poll < polls; poll += 1) {
    await page.waitForResponse((response) => response.url().includes("/api/actions/job.status"));
    seen.push(await read());
  }
  return seen;
}

export async function pausedPanel(page: Page, mode: Mode, minutes = 60): Promise<Locator> {
  const panel = await startOpen(page, mode, minutes);
  await expect.poll(() => simulatedSeconds(panel), { timeout: 30_000 }).toBeGreaterThan(0);
  await panel.getByRole("button", { name: "Pause" }).click();
  await expect(clockLine(panel)).toHaveText(/^Paused · /);
  return panel;
}
