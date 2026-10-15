import AxeBuilder from "@axe-core/playwright";
import { expect, test } from "@playwright/test";
import { readFileSync, rmSync, writeFileSync } from "node:fs";
import { join } from "node:path";
import { env } from "./support/env";

test.use({ storageState: env("CANVAS_APP_STATE") });

test("the workspace lists each real plant once with its actions, and its runs, in both themes", async ({ page }) => {
  await page.goto(env("CANVAS_APP_URL"));
  const plants = page.getByRole("region", { name: "Plants" });
  await expect(plants.getByRole("button", { name: "Edit Factory utilities" })).toBeVisible();
  await expect(plants.getByRole("button", { name: "Run Factory utilities" })).toBeVisible();
  await expect(plants.getByText("factory-utilities", { exact: true })).toHaveCount(1);
  await expect(plants.getByRole("listitem").filter({ hasText: "Factory utilities" })).toContainText("23 instances");
  await expect(plants).not.toContainText("can expand into several instances");
  const runs = page.getByRole("region", { name: "Runs" });
  await expect(runs.getByRole("listitem").filter({ hasText: "w-explore" })).toContainText("Factory utilities");
  await expect(runs.getByRole("listitem").filter({ hasText: "w-explore" })).toContainText("seed 7");
  await expect(runs.getByRole("listitem").filter({ hasText: "w-evaluate" })).toContainText("30 min");
  for (const label of ["Theme: system", "Theme: light"]) {
    await page.getByRole("button", { name: label }).click();
    expect((await new AxeBuilder({ page }).analyze()).violations).toEqual([]);
  }
});

test("a file added on disk appears without a reload", async ({ page }) => {
  await page.goto(env("CANVAS_APP_URL"));
  const plants = page.getByRole("region", { name: "Plants" });
  await expect(plants.getByText("Factory utilities", { exact: true })).toBeVisible();
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
    await expect(plants.getByRole("listitem").filter({ hasText: "Copied plant" })).toContainText("Not runnable yet", {
      timeout: 5000,
    });
  } finally {
    rmSync(copy);
  }
});
