import type { Page } from "@playwright/test";
import { expect } from "@playwright/test";
import { createHash } from "node:crypto";
import { existsSync, readdirSync, readFileSync } from "node:fs";
import { join } from "node:path";
import type { Lab } from "./lab";
import { repository } from "./workspace";

const birthFile = join(repository, "contracts", "plant-wire", "fixtures", "valid", "NBIRTH.json");

type Environment = "simulated" | "hardware";

export type Kind = Record<string, unknown>;

function birthRecord(): Record<string, unknown> {
  return JSON.parse(readFileSync(birthFile, "utf8")) as Record<string, unknown>;
}

function canonical(value: unknown): string {
  if (Array.isArray(value)) return `[${value.map(canonical).join(",")}]`;
  if (typeof value !== "object" || value === null) return JSON.stringify(value);
  const members = Object.entries(value)
    .sort(([a], [b]) => (a < b ? -1 : 1))
    .map(([key, member]) => `${JSON.stringify(key)}:${canonical(member)}`);
  return `{${members.join(",")}}`;
}

function digestOf(value: unknown): string {
  return `sha256:${createHash("sha256").update(canonical(value)).digest("hex")}`;
}

function capabilityOf(kinds: readonly Kind[]): string {
  const byName = (a: Kind, b: Kind) => (String(a.kind) < String(b.kind) ? -1 : 1);
  const entries = [...kinds].sort(byName).map((kind) => ({
    kind: kind.kind,
    commands: [...(kind.commands as Kind[])].sort((a, b) => (String(a.name) < String(b.name) ? -1 : 1)),
  }));
  return digestOf(entries);
}

export function pumpKind(): Kind {
  const [kind] = birthRecord().kinds as Kind[];
  const draft = { ...kind };
  delete draft.kind_digest;
  return draft;
}

export function fanKind(): Kind {
  return {
    kind: "fan",
    metrics: [
      { name: "duty", data_type: "number", unit: "%", eu_range: { min: 0, max: 100 }, resolution: 1, writable: true },
    ],
    commands: [{ name: "stop", args_schema: { type: "object" }, risk_class: "low", safe_stop: true }],
  };
}

export function invertedKind(): Kind {
  return {
    kind: "inverted",
    metrics: [
      { name: "level", data_type: "number", unit: "m", eu_range: { min: 10, max: 5 }, resolution: 1, writable: false },
    ],
    commands: [],
  };
}

export function birthOf(kinds: readonly Kind[], over: Record<string, unknown> = {}): string {
  const digested = kinds.map((kind) => ({ ...kind, kind_digest: digestOf(kind) }));
  const record = { ...birthRecord(), kinds: digested, capability_digest: capabilityOf(kinds), ...over };
  return JSON.stringify(record);
}

export function announcement(environment: Environment = "simulated"): string {
  const record = birthRecord();
  const boards = (record.boards as Record<string, unknown>[]).map((board) => ({ ...board, environment }));
  return JSON.stringify({ ...record, environment, boards });
}

export function digestRuleHolds(): boolean {
  const [fixture] = birthRecord().kinds as Kind[];
  return digestOf(pumpKind()) === fixture?.kind_digest;
}

export function tamperedAnnouncement(): string {
  return announcement().replace('"m3/h"', '"m3/s"');
}

export function savedKinds(lab: Lab): string[] {
  const folder = join(lab.workspace, "kinds");
  return existsSync(folder) ? readdirSync(folder) : [];
}

export function captureButton(page: Page) {
  return page
    .getByRole("region", { name: "Devices", exact: true })
    .getByRole("button", { name: "Capture from a board" });
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
