import type { Page } from "@playwright/test";
import { expect } from "@playwright/test";
import { existsSync, readdirSync, readFileSync } from "node:fs";
import { join } from "node:path";
import type { Lab } from "./lab";
import { repository } from "./workspace";

const birthFile = join(repository, "contracts", "plant-wire", "fixtures", "valid", "NBIRTH.json");

type Environment = "simulated" | "hardware";

function birthRecord(): Record<string, unknown> {
  return JSON.parse(readFileSync(birthFile, "utf8")) as Record<string, unknown>;
}

export function announcement(environment: Environment = "simulated"): string {
  return JSON.stringify({ ...birthRecord(), environment });
}

export function tamperedAnnouncement(): string {
  return announcement().replace('"m3/h"', '"m3/s"');
}

export function announcedKind(): unknown {
  const [kind] = birthRecord().kinds as Record<string, unknown>[];
  const draft = { ...kind };
  delete draft.kind_digest;
  return draft;
}

export function savedKinds(lab: Lab): string[] {
  const folder = join(lab.workspace, "kinds");
  return existsSync(folder) ? readdirSync(folder) : [];
}

export function captureButton(page: Page) {
  return page.getByRole("region", { name: "Devices" }).getByRole("button", { name: "Capture from a board" });
}

export function reviewHeading(page: Page) {
  return page.getByRole("heading", { level: 2, name: "Review what the board announced" });
}

export function refusedHeading(page: Page) {
  return page.getByRole("heading", { level: 2, name: "The board's announcement was not accepted" });
}

export function cardOf(page: Page, name: string) {
  return page.getByRole("article", { name });
}

export async function openCapture(page: Page): Promise<void> {
  await page.goto("/");
  await captureButton(page).click();
  await expect(page.getByRole("heading", { level: 1, name: "Capture from a board" })).toBeFocused();
}

export async function readAnnouncement(page: Page, text: string): Promise<void> {
  await page.getByLabel("Announcement text").fill(text);
  await page.getByRole("button", { name: "Read the announcement" }).click();
}
