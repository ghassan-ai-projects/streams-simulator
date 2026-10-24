import type { Page } from "@playwright/test";
import { expect, test } from "./support/test";

test.setTimeout(120_000);

const plant = "/#/plant?domain=domains%2Ffactory-utilities.domain.json&launch=1";
const fanFailure = /fan[ _]failure|Fault ·/i;

function recordActions(page: Page): string[] {
  const names: string[] = [];
  page.on("request", (request) => {
    const named = /\/api\/actions\/([^/?]+)/.exec(request.url());
    if (request.method() === "POST" && named?.[1] !== undefined) names.push(decodeURIComponent(named[1]));
  });
  return names;
}

function panelOf(page: Page) {
  return page.getByRole("complementary", { name: "Run Factory utilities" });
}

async function startWatched(page: Page, evaluate = false) {
  await page.goto(plant);
  const panel = panelOf(page);
  await panel.getByLabel("Seed", { exact: true }).fill("11");
  await panel.getByLabel("Duration", { exact: true }).fill("30");
  if (evaluate) await panel.getByRole("radio", { name: /^Evaluate/ }).check();
  await panel.getByRole("button", { name: "Start the run" }).click();
  await expect(panel.getByText(/^Running Factory utilities as run w-/)).toBeVisible();
  return panel;
}

test("the browser sends the plan once and then only reads status: no advance, fault or effector call", async ({
  page,
}) => {
  const names = recordActions(page);
  await page.goto(plant);
  const panel = panelOf(page);
  await panel.getByLabel("Seed", { exact: true }).fill("11");
  await panel.getByLabel("Duration", { exact: true }).fill("5");
  await panel.getByRole("radio", { name: /As fast as possible/ }).check();
  await panel.getByRole("button", { name: "Add fault" }).click();
  await panel.getByLabel("Fault 1 instance").selectOption("CT-1");
  await panel.getByLabel("Fault 1 fault").selectOption({ label: "fan failure" });
  await panel.getByLabel("Fault 1 starts at minute").fill("2");
  await panel.getByRole("button", { name: "Add effector call" }).click();
  await panel.getByLabel("Effector call 1 instance").selectOption("CH-1");
  await panel.getByLabel("Effector call 1 effector").selectOption({ label: "set setpoint" });
  await panel.getByLabel("Effector call 1 setpoint c").fill("7");
  await panel.getByLabel("Effector call 1 starts at minute").fill("1");
  await panel.getByRole("button", { name: "Start the run" }).click();
  await panel.getByRole("button", { name: "Open the review" }).click({ timeout: 60_000 });
  await expect(page.getByRole("heading", { level: 1, name: /^Run w-/ })).toBeVisible();
  expect(names.filter((name) => name === "job.start")).toHaveLength(1);
  expect(names).toContain("job.status");
  for (const forbidden of [
    "clock.advance",
    "fault.inject",
    "effector.invoke",
    "world.create",
    "run.end",
    "world.destroy",
  ]) {
    expect(names).not.toContain(forbidden);
  }
});

test("a reload in the middle of a run brings back the banner, the progress and Stop now, and Stop now ends it", async ({
  page,
}) => {
  await startWatched(page);
  await expect(panelOf(page).getByText(/^Running · 00:0[1-9]/)).toBeVisible({ timeout: 30_000 });
  await page.reload();
  const banner = page.getByRole("button", { name: /^Running Factory utilities · 00:\d\d:\d\d of 00:30:00/ });
  await expect(banner).toBeVisible();
  await banner.click();
  const panel = panelOf(page);
  await expect(panel.getByText(/^Running · \d\d:\d\d:\d\d of 00:30:00 simulated/)).toBeVisible();
  await expect(panel.getByRole("button", { name: "Stop now" })).toBeVisible();
  await panel.getByRole("button", { name: "Stop now" }).click();
  await expect(page.getByRole("heading", { level: 1, name: /^Run w-/ })).toBeVisible({ timeout: 60_000 });
  await expect(page.getByText(/^Stopped by operator at \d\d:\d\d:\d\d of 00:30:00$/)).toBeVisible();
});

test("a run left running across a reload still ends by itself and opens as a complete run", async ({ page }) => {
  await startWatched(page);
  await expect(panelOf(page).getByText(/^Running · 00:0[1-9]/)).toBeVisible({ timeout: 30_000 });
  await page.reload();
  await page.getByRole("button", { name: /^Running Factory utilities · / }).click();
  await expect(page.getByRole("heading", { level: 1, name: /^Run w-/ })).toBeVisible({ timeout: 90_000 });
  await expect(page.getByText(/Stopped by operator/)).toHaveCount(0);
  await expect(page.getByText("Did not finish", { exact: true })).toHaveCount(0);
});

test("an Evaluate run found again after a reload names no fault and no plan it cannot know", async ({ page }) => {
  await page.goto(plant);
  const panel = panelOf(page);
  await panel.getByLabel("Duration", { exact: true }).fill("30");
  await panel.getByRole("radio", { name: /^Evaluate/ }).check();
  await panel.getByRole("button", { name: "Add fault" }).click();
  await panel.getByLabel("Fault 1 instance").selectOption("CT-1");
  await panel.getByLabel("Fault 1 fault").selectOption({ label: "fan failure" });
  await panel.getByLabel("Fault 1 starts at minute").fill("20");
  await panel.getByRole("button", { name: "Start the run" }).click();
  await expect(panel.getByText(/^Running Factory utilities as run w-/)).toBeVisible();
  await page.reload();
  await page.getByRole("button", { name: /^Running Factory utilities · / }).click();
  await expect(panel.getByText("30 min of simulated time · Evaluate")).toBeVisible();
  await expect(panel.getByText(/\bfaults?\b/i)).toHaveCount(0);
  expect(await page.locator("body").innerText()).not.toMatch(fanFailure);
  await panel.getByRole("button", { name: "Stop now" }).click();
  await expect(page.getByRole("heading", { level: 1, name: /^Run w-/ })).toBeVisible({ timeout: 60_000 });
});

test("Stop now before the end keeps a shorter run that says it was stopped, and where", async ({ page }) => {
  const panel = await startWatched(page);
  await expect(panel.getByText(/^Running · 00:0[1-9]/)).toBeVisible({ timeout: 30_000 });
  await panel.getByRole("button", { name: "Stop now" }).click();
  await expect(page.getByRole("heading", { level: 1, name: /^Run w-/ })).toBeVisible({ timeout: 60_000 });
  await expect(page.getByText(/^Stopped by operator at 00:[01]\d:\d\d of 00:30:00$/)).toBeVisible();
  await expect(page.getByText("Did not finish", { exact: true })).toHaveCount(0);
  const slider = page.getByRole("slider", { name: "Playback time" });
  const longest = Number(await slider.getAttribute("max"));
  expect(longest).toBeGreaterThan(0);
  expect(longest).toBeLessThan(1800);
});
