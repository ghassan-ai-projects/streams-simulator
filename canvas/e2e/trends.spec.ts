import AxeBuilder from "@axe-core/playwright";
import type { Page } from "@playwright/test";
import { expect, test } from "./support/test";
import { selectInOutline, startWatchRun } from "./support/watchRun";

test.setTimeout(120_000);

const fanFailure = /fan[ _]failure|Fault ·|Fault starts/i;

async function pinLive(page: Page, id: string, channel: string) {
  await selectInOutline(page, id);
  const inspector = page.getByRole("complementary", { name: "Inspector" });
  await inspector.getByRole("button", { name: `Watch ${channel}` }).click();
}

async function pinTwoLive(page: Page) {
  await pinLive(page, "CT-1", "basin temp");
  await pinLive(page, "CH-1", "power");
  await page.getByRole("button", { name: /^Trends/ }).click();
  return page.getByRole("region", { name: "Trends" });
}

test("a Watch run: two channels from two assets share one time axis, each with its own unit", async ({ page }) => {
  await startWatchRun(page, "Explore");
  await expect(page.getByText(/^Running · 00:(0[3-9]|[1-5]\d)/)).toBeVisible({ timeout: 60_000 });
  const panel = await pinTwoLive(page);
  const basin = panel.getByRole("slider", { name: /^CT-1 · basin temp trend, °C$/ });
  const power = panel.getByRole("slider", { name: /^CH-1 · power trend, kW$/ });
  await expect(basin).toBeVisible();
  await expect(power).toBeVisible();
  await expect(panel.locator("[data-part='cursor-words']")).toHaveText(/^Run clock \d\d:\d\d:\d\d$/);
  await basin.focus();
  await page.keyboard.press("ArrowLeft");
  await expect(panel.locator("[data-part='cursor-words']")).toHaveText(/^Probe \d\d:\d\d:\d\d$/);
  const probed = await panel.locator("[data-part='cursor-words']").innerText();
  const at = /\d\d:\d\d:\d\d/.exec(probed)?.[0] ?? "";
  await expect(power).toHaveAttribute(
    "aria-valuetext",
    new RegExp(`^CH-1 · power: [\\d.]+ kW, observed at (\\d\\d:\\d\\d:\\d\\d)`),
  );
  await expect(basin).toHaveAttribute("aria-valuetext", new RegExp(`observed at ${at}`));
  const clipped = await page.evaluate(() => document.documentElement.scrollWidth > window.innerWidth);
  expect(clipped).toBe(false);
  await expect(page.getByRole("group", { name: /^CT-1, / })).toBeVisible();
});

test("an Explore run shows the injected fault and the call marks on the axis; picking a plot selects its asset", async ({
  page,
}) => {
  await startWatchRun(page, "Explore");
  const panel = await pinTwoLive(page);
  await expect(page.getByRole("group", { name: /^CT-1, / })).toContainText("Fault · fan failure", {
    timeout: 45_000,
  });
  await panel.getByRole("button", { name: "Whole run" }).click();
  await panel.getByText(/^Events in this window/).click();
  await expect(panel).toContainText("Fault starts · CT-1 fan failure");
  await panel.getByRole("slider", { name: /^CH-1 · power trend/ }).focus();
  await page.keyboard.press("Enter");
  await expect(page.getByRole("complementary", { name: "Inspector" })).toContainText("CH-1");
});

test("an Evaluate run: the panel shows no fault anywhere while it runs", async ({ page }) => {
  await startWatchRun(page, "Evaluate");
  await expect(page.getByText(/^Running · 00:(0[6-9]|[1-5]\d)/)).toBeVisible({ timeout: 60_000 });
  const panel = await pinTwoLive(page);
  await expect(panel.getByRole("slider")).toHaveCount(2);
  await panel.getByText("How to read these plots").click();
  const text = await page.locator("body").innerText();
  const attributes = await page.evaluate(() =>
    [...document.body.querySelectorAll("*")].flatMap((element) => [...element.attributes].map((a) => a.value)),
  );
  expect(text).not.toMatch(fanFailure);
  expect(attributes.filter((value) => fanFailure.test(value))).toEqual([]);
  expect(await page.locator("body").ariaSnapshot()).not.toMatch(fanFailure);
  await expect(panel.locator("[data-kind='fault']")).toHaveCount(0);
  expect((await new AxeBuilder({ page }).include("[aria-label='Trends']").analyze()).violations).toEqual([]);
});

test("Review: the plots follow the cursor and a plot moves it", async ({ page }) => {
  await page.goto("/#/review?world=w-evaluate");
  await expect(page.getByText(/\d+ observations/)).toBeVisible({ timeout: 60_000 });
  for (const [id, channel] of [
    ["CH-1", "power"],
    ["CT-1", "basin temp"],
  ] as const) {
    await page
      .getByRole("navigation", { name: "Outline" })
      .getByRole("button", { name: new RegExp(`^${id}\\s`) })
      .click();
    await page
      .getByRole("region", { name: `Observed · ${id}` })
      .getByRole("button", { name: `Watch ${channel}` })
      .click();
  }
  await page.getByRole("button", { name: /^Trends/ }).click();
  const panel = page.getByRole("region", { name: "Trends" });
  const plot = panel.getByRole("slider", { name: /^CH-1 · power trend/ });
  await expect(panel.getByRole("slider")).toHaveCount(2);
  const slider = page.getByRole("slider", { name: "Playback time" });
  await slider.fill("600");
  await expect(plot).toHaveAttribute("aria-valuetext", /observed at 00:(09:[0-5]\d|10:0\d)/);
  await expect(panel.locator("[data-part='cursor-words']")).toHaveText("Cursor 00:10:00");
  await plot.focus();
  await page.keyboard.press("Home");
  await page.keyboard.press("Enter");
  await expect(slider).toHaveAttribute("aria-valuetext", /^00:00:\d\d$/);
  expect((await new AxeBuilder({ page }).include("[aria-label='Trends']").analyze()).violations).toEqual([]);
});
