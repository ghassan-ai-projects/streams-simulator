import { expect, test } from "@playwright/test";
import type { Page } from "@playwright/test";
import { readFileSync } from "node:fs";
import { join } from "node:path";
import { env } from "./support/env";

test.use({ storageState: env("CANVAS_APP_STATE") });
test.setTimeout(90_000);

async function add(page: Page, component: string, name: string, size?: string) {
  const build = page.getByRole("complementary", { name: "Build" });
  await build.getByLabel("Find equipment").fill(component.split(" ")[0] ?? "");
  await build.getByLabel("Equipment", { exact: true }).selectOption({ label: component });
  await expect(build.getByLabel("Size").getByRole("option")).not.toHaveCount(1);
  if (size !== undefined) await build.getByLabel("Size").selectOption({ label: size });
  await build.getByLabel("Name", { exact: true }).fill(name);
  await build.getByLabel("Area", { exact: true }).selectOption("Chiller plant");
  await build.getByRole("button", { name: "Add to the plant" }).click();
  await expect(page.getByRole("complementary", { name })).toBeVisible();
  await page.getByRole("button", { name: "Close the instance" }).click();
}

test("two pumps, two chillers of different sizes and two air-handling units compose in one plant", async ({ page }) => {
  await page.goto(env("CANVAS_APP_URL"));
  await page.getByRole("button", { name: "New plant" }).click();
  const form = page.getByRole("form", { name: "New plant" });
  await form.getByLabel("Name").fill("Pairs");
  await form.getByLabel("Start from").selectOption({ label: "Factory utilities" });
  await form.getByRole("radio", { name: /^An empty plant/ }).check();
  await form.getByRole("button", { name: "Create and open the editor" }).click();
  await expect(page.getByRole("heading", { name: "No equipment yet" })).toBeVisible();
  await add(page, "Centrifugal pump", "pump-a");
  await add(page, "Centrifugal pump", "pump-b");
  await add(page, "Water-cooled chiller", "chiller-a", "Medium (about 300 TR)");
  await add(page, "Water-cooled chiller", "chiller-b", "Large (about 800 TR)");
  await add(page, "Air-handling unit", "ahu-a");
  await add(page, "Air-handling unit", "ahu-b");
  await page.getByRole("button", { name: "Save" }).click();
  await expect(page.getByRole("status").filter({ hasText: /^Ready to run\. Composed \d+ channels/ })).toBeVisible();
  const file = join(env("CANVAS_WORKSPACE"), "compositions", "pairs.composition.json");
  const written = JSON.parse(readFileSync(file, "utf8")) as { equipment: { names: string[] }[] };
  expect(written.equipment.map((entry) => entry.names)).toEqual([
    ["pump-a", "pump-b"],
    ["chiller-a"],
    ["chiller-b"],
    ["ahu-a", "ahu-b"],
  ]);
});
