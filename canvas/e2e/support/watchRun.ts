import type { Page } from "@playwright/test";
import { expect } from "@playwright/test";

const plant = "/#/plant?domain=domains%2Ffactory-utilities.domain.json&launch=1";

export type RunMode = "Explore" | "Evaluate";

export async function startWatchRun(page: Page, mode: RunMode, minutes = 60): Promise<void> {
  await page.goto(plant);
  const panel = page.getByRole("complementary", { name: "Run Factory utilities" });
  await panel.getByLabel("Seed", { exact: true }).fill("7");
  await panel.getByLabel("Duration", { exact: true }).fill(String(minutes));
  await panel.getByRole("radio", { name: new RegExp(`^${mode}`) }).check();
  await panel.getByRole("button", { name: "Add fault" }).click();
  await panel.getByLabel("Fault 1 instance").selectOption("CT-1");
  await panel.getByLabel("Fault 1 fault").selectOption({ label: "fan failure" });
  await panel.getByLabel("Fault 1 starts at minute").fill("4");
  await panel.getByRole("button", { name: "Start the run" }).click();
  await expect(panel.getByText(/^Running Factory utilities as run w-/)).toBeVisible();
  await panel.getByRole("button", { name: "Close the run panel" }).click();
}

export async function selectInOutline(page: Page, id: string): Promise<void> {
  await page
    .getByRole("navigation", { name: "Outline" })
    .getByRole("button", { name: new RegExp(`^${id}\\s`) })
    .click();
}

export interface Fit {
  readonly rows: number;
  readonly outside: readonly string[];
  readonly values: number;
}

export function rowsOutsideNode(id: string): Fit {
  const node = [...document.querySelectorAll<HTMLElement>(".react-flow__node-instance")].find(
    (candidate) => candidate.dataset.id === id,
  );
  if (node === undefined) return { rows: 0, outside: ["node not rendered"], values: 0 };
  const box = node.getBoundingClientRect();
  const rows = [...node.querySelectorAll("dl > div")];
  const outside = rows
    .filter((row) => {
      const rect = row.getBoundingClientRect();
      return (
        rect.top < box.top - 0.5 ||
        rect.bottom > box.bottom + 0.5 ||
        rect.left < box.left - 0.5 ||
        rect.right > box.right + 0.5
      );
    })
    .map((row) => row.textContent.slice(0, 24));
  const values = rows.filter((row) => /\d/.test(row.querySelector("dd")?.textContent ?? "")).length;
  return { rows: rows.length, outside, values };
}
