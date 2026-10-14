import AxeBuilder from "@axe-core/playwright";
import { expect, test } from "@playwright/test";
import { existsSync, readFileSync } from "node:fs";
import { join } from "node:path";
import { env } from "./support/env";

test.use({ storageState: env("CANVAS_EMPTY_STATE") });

test("a blank plant is created from an empty workspace with a name and an id, and opens in the editor", async ({
  page,
}) => {
  await page.goto(env("CANVAS_EMPTY_URL"));
  await expect(page.getByRole("heading", { name: /has no plants yet/ })).toBeVisible();
  const start = page.getByRole("button", { name: "New plant" });
  await expect(start).toHaveAttribute("aria-disabled", "false");
  await start.click();
  const form = page.getByRole("form", { name: "New plant" });
  await expect(form.getByLabel("Start from")).toHaveCount(0);
  await form.getByLabel("Name").fill("Office tower HVAC");
  await expect(form.getByLabel("Id")).toHaveValue("office-tower-hvac");
  expect((await new AxeBuilder({ page }).analyze()).violations).toEqual([]);
  await form.getByRole("button", { name: "Create and open the editor" }).click();
  await expect(page.getByRole("heading", { level: 1, name: "Office tower HVAC" })).toBeVisible();
  await expect(page.getByRole("heading", { name: "No equipment yet" })).toBeVisible();
  const file = join(env("CANVAS_EMPTY_WORKSPACE"), "compositions", "office-tower-hvac.composition.json");
  expect(existsSync(file)).toBe(true);
  const written = JSON.parse(readFileSync(file, "utf8")) as Record<string, unknown>;
  expect(written).toMatchObject({ id: "office-tower-hvac", title: "Office tower HVAC" });
  expect(written.equipment ?? []).toEqual([]);
});
