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
  const site = ((await area.getByLabel("Inside").locator("option").nth(1).textContent()) ?? "").trim();
  await area.getByLabel("Inside").selectOption({ index: 1 });
  await area.getByRole("button", { name: "Add area" }).click();
  const outline = page.getByRole("navigation", { name: "Outline" });
  await expect(outline.getByText("Pump house")).toBeVisible();
  await build.getByLabel("Find equipment").fill("pump");
  await build.getByLabel("Equipment", { exact: true }).selectOption({ label: "Centrifugal pump" });
  await build.getByLabel("Area", { exact: true }).selectOption({ label: `${site} › Pump house` });
  await build.getByRole("button", { name: "Add to the plant" }).click();
  await expect(outline.getByRole("group").filter({ hasText: "Pump house" }).getByRole("button")).toHaveCount(1);
  await page.getByRole("button", { name: "Save" }).click();
  await expect(page.getByRole("status").filter({ hasText: /^Ready to run\. Saved\.$/ })).toBeVisible();
  const file = join(env("CANVAS_WORKSPACE"), "compositions", "tiered.composition.json");
  const written = JSON.parse(readFileSync(file, "utf8")) as { areas: { name: string; parent?: string }[] };
  expect(written.areas).toContainEqual({ name: "Pump house", parent: site });
});

test("an area is renamed and an empty one deleted from the Build panel, and the outline follows", async ({ page }) => {
  await page.goto(env("CANVAS_APP_URL"));
  await page.getByRole("button", { name: "New plant" }).click();
  const form = page.getByRole("form", { name: "New plant" });
  await form.getByLabel("Name").fill("Area renames");
  await form.getByLabel("Start from").selectOption({ label: "Factory utilities" });
  await form.getByRole("radio", { name: /^An empty plant/ }).check();
  await form.getByRole("button", { name: "Create and open the editor" }).click();
  await expect(page.getByRole("heading", { name: "No equipment yet" })).toBeVisible();
  const build = page.getByRole("complementary", { name: "Build" });
  const area = build.getByRole("form", { name: "Add area" });
  for (const name of ["Scratch", "Pump house"]) {
    await area.getByLabel("Area name").fill(name);
    await area.getByRole("button", { name: "Add area" }).click();
  }
  const areas = build.getByRole("region", { name: "Areas" });
  const row = areas.locator("li").filter({ has: page.getByLabel("Name of Pump house", { exact: true }) });
  await row.getByLabel("Name of Pump house", { exact: true }).fill("Pumps");
  await row.getByRole("button", { name: "Rename" }).click();
  const outline = page.getByRole("navigation", { name: "Outline" });
  await expect(outline.getByRole("group").filter({ hasText: "Pumps" })).toHaveCount(1);
  await expect(outline.getByRole("group").filter({ hasText: "Pump house" })).toHaveCount(0);
  await areas.getByRole("button", { name: "Delete Scratch" }).click();
  await expect(outline.getByRole("group").filter({ hasText: "Scratch" })).toHaveCount(0);
  await expect(areas.getByLabel("Name of Pumps", { exact: true })).toBeVisible();
});

test("an empty area is one line of at most 28 px in the outline, however long its name", async ({ page }) => {
  await page.goto(env("CANVAS_APP_URL"));
  await page.getByRole("button", { name: "New plant" }).click();
  const form = page.getByRole("form", { name: "New plant" });
  await form.getByLabel("Name").fill("Empty rows");
  await form.getByLabel("Start from").selectOption({ label: "Factory utilities" });
  await form.getByRole("radio", { name: /^An empty plant/ }).check();
  await form.getByRole("button", { name: "Create and open the editor" }).click();
  const area = page.getByRole("complementary", { name: "Build" }).getByRole("form", { name: "Add area" });
  for (const name of ["Production hall", "Compressor room with a very long name that keeps going"]) {
    await area.getByLabel("Area name").fill(name);
    await area.getByRole("button", { name: "Add area" }).click();
  }
  const rows = page.getByRole("navigation", { name: "Outline" }).locator('summary:has([title="No equipment yet"])');
  await expect(rows).toHaveCount(3);
  const shapes = await rows.evaluateAll((elements) =>
    elements.map((element) => {
      const height = element.getBoundingClientRect().height;
      const lines = [...element.querySelectorAll("span")].map((span) => span.getClientRects().length);
      return { height, lines };
    }),
  );
  for (const shape of shapes) {
    expect(shape.height).toBeLessThanOrEqual(28);
    expect(shape.lines.every((count) => count === 1)).toBe(true);
  }
  for (const style of await rows.evaluateAll((elements) =>
    elements.map((element) => {
      const computed = getComputedStyle(element);
      return { transform: computed.textTransform, weight: computed.fontWeight };
    }),
  )) {
    expect(style).toEqual({ transform: "none", weight: "600" });
  }
});

function headerLines(): { name: string; lines: number; sameBaseline: boolean; full: string }[] {
  return [...document.querySelectorAll(".react-flow__node-frameTitle")].map((title) => {
    const name = title.querySelector("[class*='frameName']");
    const count = title.querySelector("[class*='frameCount']");
    const boxes = [name, count].map((part) => {
      const walker = document.createTreeWalker(part ?? title, NodeFilter.SHOW_TEXT);
      const found: DOMRect[] = [];
      while (walker.nextNode()) {
        const range = document.createRange();
        range.selectNodeContents(walker.currentNode);
        found.push(...[...range.getClientRects()].filter((box) => box.width > 0));
      }
      return found;
    });
    const tops = boxes.flat().map((box) => box.top);
    const lines = Math.max(...tops) - Math.min(...tops) > 3 ? 2 : 1;
    const bottoms = boxes.map((list) => list[0]?.bottom ?? 0);
    const same = Math.abs((bottoms[0] ?? 0) - (bottoms[1] ?? 0)) <= 1.5;
    const full = title.querySelector<HTMLElement>("[class*='frameTitle']:not([class*='Node'])")?.title ?? "";
    return { name: name?.textContent ?? "", lines, sameBaseline: same, full };
  });
}

test("an area header is one line at every zoom, its name cut with an ellipsis and its count kept beside it", async ({
  page,
}) => {
  await page.goto(env("CANVAS_APP_URL"));
  await page.getByRole("button", { name: "New plant" }).click();
  const form = page.getByRole("form", { name: "New plant" });
  await form.getByLabel("Name").fill("Long header");
  await form.getByLabel("Start from").selectOption({ label: "Factory utilities" });
  await form.getByRole("radio", { name: /^An empty plant/ }).check();
  await form.getByRole("button", { name: "Create and open the editor" }).click();
  const build = page.getByRole("complementary", { name: "Build" });
  const longName = "Compressor room with a very long name that keeps going";
  await build.getByRole("form", { name: "Add area" }).getByLabel("Area name").fill(longName);
  await build.getByRole("form", { name: "Add area" }).getByRole("button", { name: "Add area" }).click();
  await build.getByLabel("Find equipment").fill("pump");
  await build.getByLabel("Equipment", { exact: true }).selectOption({ label: "Centrifugal pump" });
  const areaOption = build.getByLabel("Area", { exact: true }).locator("option").filter({ hasText: longName });
  await build.getByLabel("Area", { exact: true }).selectOption({ label: (await areaOption.textContent()) ?? "" });
  await build.getByRole("button", { name: "Add to the plant" }).click();
  const graph = page.getByRole("application", { name: "Plant graph" });
  await expect(page.locator(".react-flow__node-frameTitle")).toHaveCount(1);
  await graph.focus();
  for (const key of ["0", "-", "-", "+", "+", "+", "+"]) {
    await page.keyboard.press(key);
    await expect
      .poll(() => page.evaluate(headerLines))
      .toEqual([{ name: longName, lines: 1, sameBaseline: true, full: `${longName} · 1 node` }]);
  }
});
