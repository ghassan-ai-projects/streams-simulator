import { expect, test } from "@playwright/test";
import { env } from "./support/env";
import { idProblems } from "./support/idFit";

test.use({ storageState: env("CANVAS_APP_STATE") });

function leftTitles(): { readonly left: number; readonly top: number }[] {
  const graph = document.querySelector("[aria-roledescription='plant graph']")?.getBoundingClientRect();
  const titles = [...document.querySelectorAll(".react-flow__node-frameTitle")];
  if (graph === undefined) return [];
  return titles
    .map((title) => title.getBoundingClientRect())
    .map((box) => ({ left: box.left - graph.left, top: box.top - graph.top }));
}

function transform(): string {
  return document.querySelector<HTMLElement>(".react-flow__viewport")?.style.transform ?? "";
}

test("zooming in keeps the plant's first area titles in view, and Fit returns the opening view exactly", async ({
  page,
}) => {
  await page.goto(`${env("CANVAS_APP_URL")}/#/plant?domain=domains%2Ffactory-utilities.domain.json`);
  await expect(page.getByRole("group", { name: /^CT-1, cooling tower,/ })).toBeVisible();
  const graph = page.getByRole("application", { name: "Plant graph" });
  await graph.focus();
  await page.keyboard.press("0");
  await expect.poll(() => page.evaluate(transform)).not.toBe("");
  let previous = "";
  await expect
    .poll(
      async () => {
        const now = await page.evaluate(transform);
        const settled = now === previous;
        previous = now;
        return settled;
      },
      { intervals: [300] },
    )
    .toBe(true);
  const opening = await page.evaluate(transform);
  await page.getByRole("button", { name: "Zoom in" }).click();
  await page.getByRole("button", { name: "Zoom in" }).click();
  await expect.poll(() => page.evaluate(transform)).not.toBe(opening);
  await expect
    .poll(async () => (await page.evaluate(leftTitles)).filter((box) => box.left < 0 || box.top < 0).length)
    .toBe(0);
  await page.getByRole("button", { name: "Fit the plant in view" }).click();
  await expect.poll(() => page.evaluate(transform)).toBe(opening);
});

test("a large plant opens no smaller than 40 % and shows every kind in full", async ({ page }) => {
  await page.setViewportSize({ width: 1280, height: 420 });
  await page.goto(`${env("CANVAS_APP_URL")}/#/plant?domain=domains%2Ffactory-utilities.domain.json`);
  const graph = page.getByRole("application", { name: "Plant graph" });
  await expect(page.locator(".react-flow__node-instance").first()).toBeAttached();
  await expect
    .poll(async () => Number(await graph.evaluate((element) => getComputedStyle(element).getPropertyValue("--zoom"))))
    .toBeGreaterThanOrEqual(0.4);
  const kinds = page.locator(".react-flow__node-instance [class*='kind']");
  await expect(kinds.first()).toBeVisible();
  const cut = await kinds.evaluateAll((elements) =>
    elements.filter(
      (element) => element.scrollWidth > element.clientWidth + 1 || element.scrollHeight > element.clientHeight + 1,
    ),
  );
  expect(cut).toHaveLength(0);
});

test("below 40 % an id stays on one line and never breaks mid-word, down to the overview tier", async ({ page }) => {
  await page.goto(`${env("CANVAS_APP_URL")}/#/plant?domain=domains%2Fboiler-steam-loop.domain.json`);
  const graph = page.getByRole("application", { name: "Plant graph" });
  await expect(page.locator(".react-flow__node-instance").first()).toBeAttached();
  await graph.focus();
  const zoomOf = () => graph.evaluate((element) => Number(getComputedStyle(element).getPropertyValue("--zoom")));
  for (let step = 0; step < 6; step += 1) {
    const before = await zoomOf();
    await page.keyboard.press("-");
    await expect.poll(zoomOf).toBeLessThan(before);
    await expect.poll(() => page.evaluate(idProblems)).toEqual([]);
  }
  await expect(graph).toHaveClass(/tier-overview/);
});
