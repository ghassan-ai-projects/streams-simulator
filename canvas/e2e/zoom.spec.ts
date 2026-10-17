import { expect, test } from "@playwright/test";
import { env } from "./support/env";
import { guard } from "./support/audit";
import { idProblems } from "./support/idFit";
import { zoomOutTo } from "./support/viewport";

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
  const kinds = page.locator('.react-flow__node-instance [data-part="kind"]');
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
  await expect.poll(async () => guard(await page.evaluate(idProblems, 10))).toEqual([]);
  const tier = await zoomOutTo(page, "overview", 6, () =>
    expect.poll(async () => guard(await page.evaluate(idProblems, 10))).toEqual([]),
  );
  expect(tier).toBe("overview");
  await expect(graph).toHaveAttribute("data-zoom-tier", "overview");
});

function titleProblems(atLeast = 1): { considered: number; problems: string[]; met: boolean } {
  const textBoxes = (part: Element | null): DOMRect[] => {
    const walker = document.createTreeWalker(part ?? document.body, NodeFilter.SHOW_TEXT);
    const found: DOMRect[] = [];
    while (walker.nextNode()) {
      const range = document.createRange();
      range.selectNodeContents(walker.currentNode);
      found.push(...[...range.getClientRects()].filter((box) => box.width > 0));
    }
    return found;
  };
  const titles = [...document.querySelectorAll<HTMLElement>('[data-part="frame-title"]')];
  const problems = titles.flatMap((title) => {
    const box = title.getBoundingClientRect();
    const parts = ["frame-name", "frame-count"].map((part) => textBoxes(title.querySelector(`[data-part="${part}"]`)));
    const all = parts.flat();
    const name = title.getAttribute("title") ?? "";
    const spread = Math.max(...all.map((part) => part.bottom)) - Math.min(...all.map((part) => part.bottom));
    const outside = all.some((part) => part.right > box.right + 0.5);
    return [...(outside ? [`${name}: text outside the title`] : []), ...(spread > 3 ? [`${name}: wraps`] : [])];
  });
  return { considered: titles.length, problems, met: titles.length >= atLeast };
}

test("an area header is one line at every zoom and keeps the last word of its name beside the count", async ({
  page,
}) => {
  await page.goto(`${env("CANVAS_APP_URL")}/#/plant?domain=domains%2Fboiler-steam-loop.domain.json`);
  const graph = page.getByRole("application", { name: "Plant graph" });
  await expect(page.locator(".react-flow__node-instance").first()).toBeAttached();
  await graph.focus();
  for (let step = 0; step < 5; step += 1) {
    await page.keyboard.press("-");
    await expect.poll(async () => guard(await page.evaluate(titleProblems, 3))).toEqual([]);
  }
  await expect(page.locator('[data-part="name-tail"]').filter({ hasText: "house" })).toBeVisible();
});
