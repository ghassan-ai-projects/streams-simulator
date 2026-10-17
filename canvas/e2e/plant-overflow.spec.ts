import { expect, test } from "./support/test";
import { guard } from "./support/audit";
import { nodesOutsideCanvas } from "./support/fitAudit";
import { waitForViewport } from "./support/viewport";

function graphZoom(): number {
  const graph = document.querySelector<HTMLElement>("[aria-roledescription='plant graph']");
  return Number(graph?.style.getPropertyValue("--zoom") ?? "0");
}

test("the editor opens a plant on the plant view's opening floor, one opening rule for one plant", async ({ page }) => {
  await page.setViewportSize({ width: 1280, height: 720 });
  const opening = async (route: string) => {
    await page.goto(route);
    await expect.poll(() => page.evaluate(graphZoom)).toBeCloseTo(0.4, 5);
    return page.evaluate(graphZoom);
  };
  const plant = await opening("/#/plant?domain=domains%2Ffactory-utilities.domain.json");
  const editor = await opening("/#/edit?composition=compositions%2Ffactory-utilities.composition.json");
  expect(editor).toBeCloseTo(plant, 5);
});

test("the plant view opens a plant too large for 40 % at the 40 % floor, with a fade where to pan", async ({
  page,
}) => {
  await page.setViewportSize({ width: 1280, height: 720 });
  await page.goto("/#/plant?domain=domains%2Ffactory-utilities.domain.json");
  await expect(page.getByRole("group", { name: /^CT-1, cooling tower,/ })).toBeVisible();
  await waitForViewport(page);
  expect(await page.evaluate(graphZoom)).toBeCloseTo(0.4, 5);
  const overflowing = await page.evaluate(nodesOutsideCanvas, 5);
  expect(overflowing.met && overflowing.problems.length > 0).toBe(true);
  await expect(page.locator("[data-pan-cue='right']")).toBeVisible();
  await page.getByRole("button", { name: "Fit the plant in view" }).click();
  await expect.poll(async () => guard(await page.evaluate(nodesOutsideCanvas, 5))).toEqual([]);
  await expect(page.locator("[data-pan-cue]")).toHaveCount(0);
});
