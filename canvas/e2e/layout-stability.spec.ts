import type { Page } from "@playwright/test";
import { expect, test } from "./support/test";
import { canvasProbe, coveredBy, graphBoxes, heightAt, layoutShift, watchLayout } from "./support/layoutShift";

test.setTimeout(90_000);

const budget = 0.02;
const factory = "/#/plant?domain=domains%2Ffactory-utilities.domain.json";
const finalChip = /^(Verified|Not verified|Running|Did not finish)$/;

async function settled(page: Page): Promise<number> {
  await page.evaluate(() => new Promise((resolve) => requestAnimationFrame(() => requestAnimationFrame(resolve))));
  return layoutShift(page);
}

async function createWorld(page: Page, seed: number): Promise<string> {
  const response = await page.request.post("/api/actions/world.create", {
    data: { args: { domain: "factory-utilities", seed, adapter: "native-jsonl", time_mode: "stepped" } },
  });
  const body = (await response.json()) as { result: { world_id: string } };
  return body.result.world_id;
}

async function openReview(page: Page, world: string): Promise<void> {
  await page.goto(`/#/review?world=${world}`);
  await expect(page.getByRole("status").filter({ hasText: finalChip })).toHaveCount(1, { timeout: 60_000 });
}

async function expectStableCanvas(page: Page): Promise<void> {
  await expect.poll(async () => (await canvasProbe(page)).elapsed, { timeout: 15_000 }).toBeGreaterThanOrEqual(1000);
  const { samples } = await canvasProbe(page);
  expect(heightAt(samples, 1000)).toBe(heightAt(samples, 100));
  expect(samples).toHaveLength(1);
}

test.beforeEach(async ({ page }) => {
  await watchLayout(page);
});

test("the workspace does not shift while it loads", async ({ page }) => {
  await page.goto("/");
  await expect(page.getByRole("region", { name: "Plants" }).getByRole("listitem").first()).toBeVisible();
  expect(await settled(page)).toBeLessThanOrEqual(budget);
});

test("the plant screen does not shift while it loads", async ({ page }) => {
  await page.goto(factory);
  await expect(page.getByRole("application", { name: "Plant graph" })).toBeVisible();
  expect(await settled(page)).toBeLessThanOrEqual(budget);
});

test("opening the launch panel does not shift the plant screen", async ({ page }) => {
  await page.goto(factory);
  await page.getByRole("button", { name: "Run plant" }).click();
  await expect(page.getByRole("complementary", { name: "Run Factory utilities" })).toBeVisible();
  expect(await settled(page)).toBeLessThanOrEqual(budget);
});

test("the editor does not shift while it loads", async ({ page }) => {
  await page.goto(`/#/edit?composition=${encodeURIComponent("compositions/factory-utilities.composition.json")}`);
  await expect(page.getByRole("heading", { level: 1, name: "Factory utilities" })).toBeVisible();
  expect(await settled(page)).toBeLessThanOrEqual(budget);
});

for (const world of ["w-evaluate", "w-tampered", "w-incomplete-stamp", "w-edited"]) {
  test(`the review of ${world} keeps its layout and canvas height from the first paint`, async ({ page }) => {
    await openReview(page, world);
    expect(await settled(page)).toBeLessThanOrEqual(budget);
    await expectStableCanvas(page);
  });
}

test("the review of a world that is still open does not shift", async ({ page }) => {
  await openReview(page, await createWorld(page, 5));
  expect(await settled(page)).toBeLessThanOrEqual(budget);
});

test("the review of an unfinished, destroyed world keeps its canvas height", async ({ page }) => {
  const world = await createWorld(page, 6);
  await page.request.post("/api/actions/world.destroy", { data: { args: { world_id: world } } });
  await openReview(page, world);
  expect(await settled(page)).toBeLessThanOrEqual(budget);
  await expectStableCanvas(page);
});

const differ = "The run's files differ from what the run wrote: playback shows the files as they are now.";
const noticed = [
  { world: "w-edited", alerts: [differ] },
  { world: "w-tampered", alerts: [] },
];

for (const { world, alerts } of noticed) {
  test(`the notices of ${world} never cover the graph, keep their text for assistive tech and keep the layout steady`, async ({
    page,
  }) => {
    await openReview(page, world);
    const note = page.getByRole("status").filter({ hasText: "The plant file changed after this run" });
    await expect(note).toHaveCount(1);
    await expect(note).toHaveAttribute("title", /plant file changed after this run/);
    await expect(page.getByRole("alert")).toHaveCount(alerts.length);
    for (const text of alerts) {
      await expect(page.getByRole("alert", { name: text })).toHaveAttribute("title", text);
    }
    for (const zoom of ["Zoom in", "Zoom in", "Fit the plant in view", "Zoom out", "Zoom out"]) {
      await page.getByRole("button", { name: zoom }).click();
      const boxes = await graphBoxes(page);
      expect(boxes.notices).toHaveLength(alerts.length + 1);
      expect(boxes.targets.length).toBeGreaterThan(0);
      expect(coveredBy(boxes)).toEqual([]);
    }
    expect(await settled(page)).toBeLessThanOrEqual(budget);
    await expectStableCanvas(page);
  });
}
