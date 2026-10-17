import type { Page } from "@playwright/test";
import { expect, test } from "@playwright/test";
import { env } from "./support/env";
import { fitAndWait, stepZoom } from "./support/viewport";
import { guard } from "./support/audit";
import { wireEndsTooClose, wiresNearTitles, wiresUnderNodes } from "./support/wireAudit";

test.use({ storageState: env("CANVAS_APP_STATE") });
test.setTimeout(90_000);

async function checkEveryTier(page: Page) {
  const graph = page.getByRole("application", { name: "Plant graph" });
  const budget = { intervals: [250, 500, 1000], timeout: 20_000 };
  await fitAndWait(page);
  for (let step = 0; step < 6; step += 1) {
    await expect.poll(async () => guard(await page.evaluate(wiresUnderNodes, 10)), budget).toEqual([]);
    await expect.poll(async () => guard(await page.evaluate(wireEndsTooClose, 20)), budget).toEqual([]);
    const tier = await graph.getAttribute("data-zoom-tier");
    if (tier === "compact" || tier === "full") {
      await expect.poll(async () => guard(await page.evaluate(wiresNearTitles, 20)), budget).toEqual([]);
    }
    await stepZoom(page, "+");
  }
}

test("no wire runs under a node and no two wire ends meet, at every zoom tier of a plant", async ({ page }) => {
  await page.goto(`${env("CANVAS_APP_URL")}/#/plant?domain=domains%2Ffactory-utilities.domain.json`);
  await expect(page.getByRole("group", { name: /^CT-1, cooling tower,/ })).toBeVisible();
  await checkEveryTier(page);
});

test("no wire runs under a node and no two wire ends meet, at every zoom tier of a review", async ({ page }) => {
  await page.goto(`${env("CANVAS_APP_URL")}/#/review?world=w-evaluate`);
  await expect(page.getByText(/\d+ observations/)).toBeVisible({ timeout: 60_000 });
  await checkEveryTier(page);
});

test("no wire runs under a node in the editor of the composed fixture plant, at every zoom tier", async ({ page }) => {
  await page.goto(`${env("CANVAS_APP_URL")}/#/edit?composition=compositions%2Ffactory-utilities.composition.json`);
  await expect(page.getByRole("navigation", { name: "Outline" }).getByRole("button")).not.toHaveCount(0);
  await checkEveryTier(page);
});
