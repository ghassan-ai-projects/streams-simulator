import AxeBuilder from "@axe-core/playwright";
import { expect, test } from "@playwright/test";
import { readFileSync, rmSync, writeFileSync } from "node:fs";
import { join } from "node:path";
import { env } from "./support/env";

test.use({ storageState: env("CANVAS_APP_STATE") });

test("the workspace lists the real plant, composition and runs in both themes", async ({ page }) => {
  await page.goto(env("CANVAS_APP_URL"));
  const plants = page.getByRole("region", { name: "Plants" });
  await expect(plants.getByText("factory-utilities", { exact: true })).toBeVisible();
  await expect(
    page.getByRole("region", { name: "Compositions" }).getByText("factory-utilities", { exact: true }),
  ).toBeVisible();
  const runs = page.getByRole("region", { name: "Runs" });
  await expect(runs.getByText("w-explore")).toBeVisible();
  await expect(runs.getByText("w-evaluate")).toBeVisible();
  for (const label of ["Theme: follow the system", "Theme: light"]) {
    await page.getByRole("button", { name: label }).click();
    expect((await new AxeBuilder({ page }).analyze()).violations).toEqual([]);
  }
});

test("a file added on disk appears without a reload", async ({ page }) => {
  await page.goto(env("CANVAS_APP_URL"));
  const compositions = page.getByRole("region", { name: "Compositions" });
  await expect(compositions.getByText("factory-utilities.composition.json")).toBeVisible();
  const folder = join(env("CANVAS_WORKSPACE"), "compositions");
  const copy = join(folder, "copy.composition.json");
  const source = readFileSync(join(folder, "factory-utilities.composition.json"), "utf8");
  writeFileSync(
    copy,
    source
      .replace('"id": "factory-utilities"', '"id": "copied-plant"')
      .replace('"title": "Factory utilities"', '"title": "Copied plant"'),
  );
  try {
    await expect(compositions.getByText("compositions/copy.composition.json")).toBeVisible({ timeout: 5000 });
  } finally {
    rmSync(copy);
  }
});
