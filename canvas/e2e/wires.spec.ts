import type { Page } from "@playwright/test";
import { expect, test } from "@playwright/test";
import { env } from "./support/env";
import { fitAndWait } from "./support/viewport";
import { wireEndsTooClose, wiresNearTitles, wiresUnderNodes } from "./support/wireAudit";

test.use({ storageState: env("CANVAS_APP_STATE") });
test.setTimeout(90_000);

async function checkEveryTier(page: Page) {
  const graph = page.getByRole("application", { name: "Plant graph" });
  const budget = { intervals: [250, 500, 1000], timeout: 20_000 };
  await fitAndWait(page);
  for (let step = 0; step < 6; step += 1) {
    await expect.poll(() => page.evaluate(wiresUnderNodes), budget).toEqual([]);
    await expect.poll(() => page.evaluate(wireEndsTooClose), budget).toEqual([]);
    const tier = (await graph.getAttribute("class"))?.match(/tier-(\w+)/)?.[1];
    if (tier === "compact" || tier === "full") {
      await expect.poll(() => page.evaluate(wiresNearTitles), budget).toEqual([]);
    }
    await page.keyboard.press("+");
  }
}

test("no wire runs under a node and no two wire ends meet, at every zoom tier of a plant", async ({ page }) => {
  await page.goto(`${env("CANVAS_APP_URL")}/#/plant?domain=domains%2Ffactory-utilities.domain.json`);
  await expect(page.getByRole("group", { name: /^CT-1, cooling tower,/ })).toBeVisible();
  await checkEveryTier(page);
});

test("no wire runs under a node and no two wire ends meet, at every zoom tier of a review", async ({ page }) => {
  await page.goto(`${env("CANVAS_APP_URL")}/#/run?world=w-evaluate`);
  await expect(page.getByText(/\d+ observations/)).toBeVisible({ timeout: 60_000 });
  await checkEveryTier(page);
});
