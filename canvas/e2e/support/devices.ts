import type { Page } from "@playwright/test";
import { expect } from "@playwright/test";
import { mkdirSync, writeFileSync } from "node:fs";
import { join } from "node:path";
import type { Lab } from "./lab";

export function outline(page: Page) {
  return page.getByRole("navigation", { name: "Device outline" });
}

export function verdict(page: Page, text: string | RegExp) {
  return page.getByRole("status").filter({ hasText: text });
}

export function step(page: Page, name: RegExp) {
  return page.getByRole("navigation", { name: "Device steps" }).getByRole("button", { name });
}

export async function createDevice(page: Page, name: string): Promise<void> {
  await page.goto("/");
  const devices = page.getByRole("region", { name: "Devices", exact: true });
  await expect(devices.getByText(/No devices yet/)).toBeVisible();
  await devices.getByRole("button", { name: "Create device" }).click();
  const form = page.getByRole("form", { name: "New device" });
  await form.getByLabel("Name", { exact: true }).fill(name);
  await form.getByRole("button", { name: "Create and open the device" }).click();
  await expect(page.getByRole("region", { name: `Editing ${name}` })).toBeVisible();
  await expect(page.getByRole("heading", { level: 1 })).toHaveAccessibleName(name);
  await expect(verdict(page, "Saved and valid.")).toBeVisible();
}

export function kindPath(lab: Lab, name: string): string {
  return join(lab.workspace, "kinds", `${name}.kind.json`);
}

export function writeKind(lab: Lab, name: string, document: unknown): void {
  mkdirSync(join(lab.workspace, "kinds"), { recursive: true });
  writeFileSync(kindPath(lab, name), JSON.stringify(document));
}
