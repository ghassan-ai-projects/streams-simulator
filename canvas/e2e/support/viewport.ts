import type { Page } from "@playwright/test";
import { expect } from "@playwright/test";

function viewportTransform(): string {
  return document.querySelector<HTMLElement>(".react-flow__viewport")?.style.transform ?? "";
}

async function settledTier(page: Page): Promise<string> {
  await page.evaluate(
    () =>
      new Promise((resolve) => {
        requestAnimationFrame(() => {
          requestAnimationFrame(() => {
            resolve(null);
          });
        });
      }),
  );
  return page.evaluate(tierOfGraph);
}

function tierOfGraph(): string {
  const graph = document.querySelector("[aria-roledescription='plant graph']");
  return graph?.getAttribute("data-zoom-tier") ?? "";
}

export async function waitForViewport(page: Page): Promise<void> {
  let previous = "";
  await expect
    .poll(
      async () => {
        const now = await page.evaluate(viewportTransform);
        const settled = now !== "" && now === previous;
        previous = now;
        return settled;
      },
      { intervals: [120, 160, 200, 400] },
    )
    .toBe(true);
}

export async function fitAndWait(page: Page): Promise<void> {
  await page.getByRole("application", { name: "Plant graph" }).focus();
  await page.keyboard.press("0");
  await waitForViewport(page);
}

/** One zoom step that has taken effect: the keypress is ignored while a previous zoom is still animating. */
export async function stepZoom(page: Page, key: "+" | "-"): Promise<void> {
  await page.keyboard.press(key);
  await waitForViewport(page);
}

export async function zoomInTo(page: Page, target: string, limit: number): Promise<string> {
  let tier = await settledTier(page);
  let step = 0;
  while (step < limit && tier !== target) {
    await stepZoom(page, "+");
    tier = await settledTier(page);
    step += 1;
  }
  return tier;
}

export async function zoomOutTo(
  page: Page,
  target: string,
  limit: number,
  visit: () => Promise<unknown>,
): Promise<string> {
  let tier = await page.evaluate(tierOfGraph);
  let step = 0;
  while (step < limit && tier !== target) {
    await page.keyboard.press("-");
    await waitForViewport(page);
    await visit();
    tier = await page.evaluate(tierOfGraph);
    step += 1;
  }
  return tier;
}
