import { expect, test } from "@playwright/test";
import { env } from "./support/env";

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
