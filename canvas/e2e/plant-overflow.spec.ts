import { expect, test } from "@playwright/test";
import { env } from "./support/env";
import { nodesOutsideCanvas } from "./support/fitAudit";
import { waitForViewport } from "./support/viewport";

test.use({ storageState: env("CANVAS_APP_STATE") });

function graphZoom(): number {
  const graph = document.querySelector<HTMLElement>("[aria-roledescription='plant graph']");
  return Number(graph?.style.getPropertyValue("--zoom") ?? "0");
}

test("the plant view opens a plant too large for 40 % at the 40 % floor, with a fade where to pan", async ({
  page,
}) => {
  await page.setViewportSize({ width: 1280, height: 720 });
  await page.goto(`${env("CANVAS_APP_URL")}/#/plant?domain=domains%2Ffactory-utilities.domain.json`);
  await expect(page.getByRole("group", { name: /^CT-1, cooling tower,/ })).toBeVisible();
  await waitForViewport(page);
  expect(await page.evaluate(graphZoom)).toBeCloseTo(0.4, 5);
  expect(await page.evaluate(nodesOutsideCanvas)).not.toEqual([]);
  await expect(page.locator("[data-pan-cue='right']")).toBeVisible();
  await page.getByRole("button", { name: "Fit the plant in view" }).click();
  await expect.poll(() => page.evaluate(nodesOutsideCanvas)).toEqual([]);
  await expect(page.locator("[data-pan-cue]")).toHaveCount(0);
});
