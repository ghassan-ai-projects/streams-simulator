import { expect, test } from "@playwright/test";
import { readFileSync } from "node:fs";
import { join } from "node:path";
import { env } from "./support/env";

test.use({ storageState: env("CANVAS_APP_STATE") });
test.setTimeout(60_000);

test("an area is added inside the site, shown nested in the outline, used by equipment and composed", async ({
  page,
}) => {
  await page.goto(env("CANVAS_APP_URL"));
  await page.getByRole("button", { name: "New plant" }).click();
  const form = page.getByRole("form", { name: "New plant" });
  await form.getByLabel("Name").fill("Tiered");
  await form.getByLabel("Start from").selectOption({ label: "Factory utilities" });
  await form.getByRole("radio", { name: /^An empty plant/ }).check();
  await form.getByRole("button", { name: "Create and open the editor" }).click();
  await expect(page.getByRole("heading", { name: "No equipment yet" })).toBeVisible();
  const build = page.getByRole("complementary", { name: "Build" });
  const area = build.getByRole("form", { name: "Add area" });
  await area.getByLabel("Area name").fill("Pump house");
  await area.getByLabel("Inside").selectOption({ label: "Site" });
  await area.getByRole("button", { name: "Add area" }).click();
  const outline = page.getByRole("navigation", { name: "Outline" });
  await expect(outline.getByText("Pump house")).toBeVisible();
  await build.getByLabel("Find equipment").fill("pump");
  await build.getByLabel("Equipment", { exact: true }).selectOption({ label: "Centrifugal pump" });
  await build.getByLabel("Area", { exact: true }).selectOption({ label: "Site › Pump house" });
  await build.getByRole("button", { name: "Add to the plant" }).click();
  await expect(outline.getByRole("group").filter({ hasText: "Pump house" }).getByRole("button")).toHaveCount(1);
  await page.getByRole("button", { name: "Save" }).click();
  await expect(page.getByRole("status").filter({ hasText: /^Saved\. Composed \d+ channels/ })).toBeVisible();
  const file = join(env("CANVAS_WORKSPACE"), "compositions", "tiered.composition.json");
  const written = JSON.parse(readFileSync(file, "utf8")) as { areas: { name: string; parent?: string }[] };
  expect(written.areas).toContainEqual({ name: "Pump house", parent: "Site" });
});
