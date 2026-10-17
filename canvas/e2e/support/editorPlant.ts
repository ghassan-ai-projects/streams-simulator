import type { Locator, Page } from "@playwright/test";
import { expect } from "@playwright/test";
import { readFileSync, writeFileSync } from "node:fs";
import { join } from "node:path";
import type { Lab } from "./lab";

export async function openEmptyPlant(page: Page, name: string): Promise<void> {
  await page.goto("/");
  await page.getByRole("button", { name: "New plant" }).click();
  const form = page.getByRole("form", { name: "New plant" });
  await form.getByLabel("Name").fill(name);
  await form.getByLabel("Start from").selectOption({ label: "Factory utilities" });
  await form.getByRole("radio", { name: /^An empty plant/ }).check();
  await form.getByRole("button", { name: "Create and open the editor" }).click();
  await expect(page.getByRole("heading", { name: "No equipment yet" })).toBeVisible();
}

export async function addEquipment(page: Page, search: string, component: string, name: string): Promise<void> {
  const build = page.getByRole("complementary", { name: "Build" });
  await build.getByLabel("Find equipment").fill(search);
  await build.getByLabel("Equipment", { exact: true }).selectOption({ label: component });
  await build.getByLabel("Name", { exact: true }).fill(name);
  await build.getByLabel("Area", { exact: true }).selectOption({ index: 0 });
  await build.getByRole("button", { name: "Add to the plant" }).click();
  await expect(page.getByRole("complementary", { name })).toBeVisible();
  await page.getByRole("button", { name: "Close the instance" }).click();
}

export function pin(page: Page, node: string, port: string): Locator {
  return page.locator(`.react-flow__node[data-id="${node}"] .react-flow__handle[data-handleid="${port}"]`);
}

export async function dragPin(page: Page, from: Locator, to: Locator): Promise<void> {
  await from.hover();
  await page.mouse.down();
  const box = await to.boundingBox();
  if (box === null) throw new Error("the target pin has no box");
  await page.mouse.move(box.x + box.width / 2, box.y + box.height / 2, { steps: 12 });
  await page.mouse.up();
}

export function writeProbePlant(lab: Lab, id: string, title: string): string {
  const compositions = join(lab.workspace, "compositions");
  const source = JSON.parse(readFileSync(join(compositions, "factory-utilities.composition.json"), "utf8")) as {
    equipment: { names: string[] }[];
  };
  const equipment = source.equipment.filter((entry) => entry.names[0] === "CH-1");
  const file = join(compositions, `${id}.composition.json`);
  writeFileSync(file, JSON.stringify({ ...source, id, title, equipment, connections: [] }, null, 2));
  return `/#/edit?composition=${encodeURIComponent(`compositions/${id}.composition.json`)}`;
}

export function savedPlant(lab: Lab, id: string): { equipment: { names: string[] }[]; connections: unknown[] } {
  const file = join(lab.workspace, "compositions", `${id}.composition.json`);
  return JSON.parse(readFileSync(file, "utf8")) as { equipment: { names: string[] }[]; connections: unknown[] };
}
