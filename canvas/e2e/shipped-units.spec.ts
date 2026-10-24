import { expect, test } from "@playwright/test";
import { readdirSync, readFileSync } from "node:fs";
import { join, resolve } from "node:path";

const root = resolve(import.meta.dirname, "../..");
const shipped = ["domains", "examples", "internal/component/internal/files"];

function jsonFiles(folder: string): string[] {
  return readdirSync(folder, { withFileTypes: true }).flatMap((entry) => {
    const path = join(folder, entry.name);
    if (entry.isDirectory()) return jsonFiles(path);
    return entry.name.endsWith(".json") ? [path] : [];
  });
}

function unitsIn(value: unknown, found: Set<string>): void {
  if (Array.isArray(value)) {
    value.forEach((item) => {
      unitsIn(item, found);
    });
  }
  if (typeof value !== "object" || value === null || Array.isArray(value)) return;
  for (const [key, item] of Object.entries(value)) {
    if (key === "unit" && typeof item === "string") found.add(item);
    unitsIn(item, found);
  }
}

function shippedUnits(): string[] {
  const found = new Set<string>();
  for (const folder of shipped) {
    for (const path of jsonFiles(join(root, folder))) unitsIn(JSON.parse(readFileSync(path, "utf8")), found);
  }
  found.delete("");
  return [...found].sort();
}

test("every unit code in the shipped domains, examples and components has a presentation mapping", async () => {
  const path = resolve(import.meta.dirname, "../src/model/shared/units.ts");
  const table = (await import(path)) as { presentedUnits(): readonly string[] };
  const presented = new Set(table.presentedUnits());
  const found = shippedUnits();
  expect(found.length).toBeGreaterThan(30);
  expect(found.filter((code) => !presented.has(code))).toEqual([]);
});
