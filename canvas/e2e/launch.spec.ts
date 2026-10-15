import AxeBuilder from "@axe-core/playwright";
import { expect, test } from "@playwright/test";
import { env } from "./support/env";

test.use({ storageState: env("CANVAS_APP_STATE") });
test.setTimeout(90_000);

test("a plant runs with a planned fault and the finished run opens in the review", async ({ page }) => {
  await page.goto(`${env("CANVAS_APP_URL")}/#/plant?domain=domains%2Ffactory-utilities.domain.json`);
  await page.getByRole("button", { name: "Run plant" }).click();
  const panel = page.getByRole("complementary", { name: "Run Factory utilities" });
  await panel.getByLabel("Seed", { exact: true }).fill("11");
  await panel.getByLabel("Duration", { exact: true }).fill("5");
  await panel.getByRole("button", { name: "Add fault" }).click();
  await panel.getByLabel("Fault 1 instance").selectOption("CT-1");
  await panel.getByLabel("Fault 1 fault").selectOption({ label: "fan failure" });
  await panel.getByLabel("Fault 1 starts at minute").fill("2");
  await panel.getByRole("radio", { name: /As fast as possible/ }).check();
  expect((await new AxeBuilder({ page }).include("aside").analyze()).violations).toEqual([]);
  await panel.getByRole("button", { name: "Start the run" }).click();
  const header = page.getByRole("heading", { level: 1, name: /^Run w-/ });
  await expect(header).toBeVisible({ timeout: 60_000 });
  await expect(page.getByText("Explore", { exact: true })).toBeVisible();
  await expect(page.getByText(/^seed 11 · /)).toBeVisible();
  const slider = page.getByRole("slider", { name: "Playback time" });
  await expect(slider).toHaveAttribute("max", "300", { timeout: 60_000 });
  const disturbances = page
    .getByRole("complementary", { name: "Run details" })
    .getByRole("region", { name: "Disturbances" });
  await expect(disturbances).toContainText("00:02:00");
  await expect(disturbances).toContainText("CT-1");
  await expect(disturbances).toContainText("fan failure");
});

test("Enter in the fault select chooses nothing and does not start the run", async ({ page }) => {
  await page.goto(`${env("CANVAS_APP_URL")}/#/plant?domain=domains%2Ffactory-utilities.domain.json`);
  await page.getByRole("button", { name: "Run plant" }).click();
  const panel = page.getByRole("complementary", { name: "Run Factory utilities" });
  await panel.getByLabel("Seed", { exact: true }).fill("12");
  await panel.getByLabel("Duration", { exact: true }).fill("5");
  await panel.getByRole("button", { name: "Add fault" }).click();
  await panel.getByLabel("Fault 1 instance").selectOption("CT-1");
  await panel.getByLabel("Fault 1 fault").click();
  await page.keyboard.type("fan");
  await page.keyboard.press("Enter");
  await panel.getByLabel("Fault 1 fault").selectOption({ label: "fan failure" });
  await expect(panel.getByRole("button", { name: "Start the run" })).toBeVisible();
  await expect(page.getByRole("heading", { level: 1, name: /^Run w-/ })).toHaveCount(0);
});

test("choosing a fault keeps the run panel where it is, and Start stays apart from the add buttons", async ({
  page,
}) => {
  await page.setViewportSize({ width: 1440, height: 640 });
  await page.goto(`${env("CANVAS_APP_URL")}/#/plant?domain=domains%2Ffactory-utilities.domain.json`);
  await page.getByRole("button", { name: "Run plant" }).click();
  const panel = page.getByRole("complementary", { name: "Run Factory utilities" });
  await panel.getByRole("button", { name: "Add fault" }).click();
  await panel.getByLabel("Fault 1 instance").selectOption("CT-1");
  const scrolled = () => panel.evaluate((element) => element.scrollTop);
  const addCall = panel.getByRole("button", { name: "Add effector call" });
  const before = await scrolled();
  const lineBefore = (await addCall.boundingBox())?.y;
  await panel.getByLabel("Fault 1 fault").selectOption({ label: "fan failure" });
  await page.keyboard.press("Tab");
  expect(await scrolled()).toBe(before);
  expect((await addCall.boundingBox())?.y).toBe(lineBefore);
  await expect(panel.getByRole("button", { name: "Start the run" })).toBeInViewport();
});

test("an effector call is made at its minute and its outcome is listed before the review", async ({ page }) => {
  await page.goto(`${env("CANVAS_APP_URL")}/#/plant?domain=domains%2Ffactory-utilities.domain.json`);
  await page.getByRole("button", { name: "Run plant" }).click();
  const panel = page.getByRole("complementary", { name: "Run Factory utilities" });
  await panel.getByLabel("Seed", { exact: true }).fill("11");
  await panel.getByLabel("Duration", { exact: true }).fill("5");
  await panel.getByRole("radio", { name: /As fast as possible/ }).check();
  await panel.getByRole("button", { name: "Add effector call" }).click();
  await panel.getByLabel("Effector call 1 instance").selectOption("CH-1");
  await panel.getByLabel("Effector call 1 effector").selectOption({ label: "set setpoint" });
  await panel.getByLabel("Effector call 1 setpoint c").fill("20");
  await panel.getByRole("button", { name: "Start the run" }).click();
  await expect(panel.getByLabel("Effector call 1 setpoint c")).toHaveAccessibleDescription(
    "Effector call 1: setpoint c is a number from 4 to 12.",
  );
  await panel.getByLabel("Effector call 1 setpoint c").fill("7");
  await panel.getByLabel("Effector call 1 starts at minute").fill("2");
  expect((await new AxeBuilder({ page }).include("aside").analyze()).violations).toEqual([]);
  await panel.getByRole("button", { name: "Start the run" }).click();
  const results = panel.getByRole("list", { name: "Effector call results" });
  await expect(results).toContainText("00:02:00 · CH-1 · set setpoint: ", { timeout: 60_000 });
  await expect(page.getByText(/finished/)).toHaveCount(1);
  await panel.getByRole("button", { name: "Open the review" }).click();
  await expect(page.getByRole("heading", { level: 1, name: /^Run w-/ })).toBeVisible();
  const slider = page.getByRole("slider", { name: "Playback time" });
  await expect(slider).toHaveAttribute("max", "300", { timeout: 60_000 });
  await slider.fill("130");
  await expect(page.getByRole("group", { name: /^CH-1, .*effector call: set setpoint · applied/ })).toBeVisible();
  await slider.fill("30");
  await expect(page.getByRole("group", { name: /^CH-1, chiller/ })).not.toContainText("effector call");
  await slider.fill("200");
  await expect(page.getByRole("group", { name: /^CH-1, chiller/ })).not.toContainText("effector call");
});

test("a plan with problems cannot start", async ({ page }) => {
  await page.goto(`${env("CANVAS_APP_URL")}/#/plant?domain=domains%2Ffactory-utilities.domain.json`);
  await page.getByRole("button", { name: "Run plant" }).click();
  const panel = page.getByRole("complementary", { name: "Run Factory utilities" });
  await panel.getByRole("button", { name: "Add fault" }).click();
  await panel.getByLabel("Fault 1 instance").selectOption("CT-1");
  await panel.getByRole("button", { name: "Start the run" }).click();
  await expect(panel.getByLabel("Fault 1 fault")).toBeFocused();
  await expect(panel.getByLabel("Fault 1 fault")).toHaveAccessibleDescription("Fault 1 needs a fault.");
  await panel.getByRole("button", { name: "Close the run panel" }).click();
  await expect(page.getByRole("complementary", { name: "Inspector" })).toBeVisible();
  await expect(page.getByRole("button", { name: "Run plant" })).toBeFocused();
});

test("opening the run panel shows the details panel even when it was hidden", async ({ page }) => {
  await page.goto(`${env("CANVAS_APP_URL")}/#/plant?domain=domains%2Ffactory-utilities.domain.json`);
  await page.getByRole("button", { name: "Inspector panel" }).click();
  await expect(page.getByRole("complementary", { name: "Inspector" })).toBeHidden();
  await page.getByRole("button", { name: "Run plant" }).click();
  await expect(page.getByRole("complementary", { name: "Run Factory utilities" })).toBeVisible();
  await expect(page.getByRole("button", { name: "Inspector panel" })).toHaveAttribute("aria-pressed", "true");
});

test("the top-bar run indicator brings back a closed run panel, and Stop now keeps the shorter run", async ({
  page,
}) => {
  await page.goto(`${env("CANVAS_APP_URL")}/#/plant?domain=domains%2Ffactory-utilities.domain.json&run=1`);
  const panel = page.getByRole("complementary", { name: "Run Factory utilities" });
  await panel.getByLabel("Duration", { exact: true }).fill("24");
  await panel.getByLabel("Duration unit").selectOption("hours");
  await panel.getByRole("button", { name: "Start the run" }).click();
  await expect(panel.getByText(/^Running Factory utilities as run w-/)).toBeVisible();
  const progress = panel.getByText(/^Running · \d\d:\d\d:\d\d of 24:00:00 simulated/);
  const first = (await progress.textContent()) ?? "";
  await expect(progress).toHaveText(/ · \d+ s per s$/, { timeout: 15_000 });
  await expect(progress).not.toHaveText(first.split(" · ")[1] ?? "");
  await panel.getByRole("button", { name: "Close the run panel" }).click();
  await expect(panel).toBeHidden();
  await page.getByRole("button", { name: /^Running Factory utilities · / }).click();
  await expect(panel.getByRole("button", { name: "Stop now" })).toBeVisible();
  await panel.getByRole("button", { name: "Stop now" }).click();
  await expect(page.getByRole("heading", { level: 1, name: /^Run w-/ })).toBeVisible({ timeout: 60_000 });
});

test("a planned fault appears on its node when the run clock reaches its minute and stays to the end", async ({
  page,
}) => {
  await page.goto(`${env("CANVAS_APP_URL")}/#/plant?domain=domains%2Ffactory-utilities.domain.json&run=1`);
  const panel = page.getByRole("complementary", { name: "Run Factory utilities" });
  await panel.getByLabel("Seed", { exact: true }).fill("11");
  await panel.getByLabel("Duration", { exact: true }).fill("20");
  await panel.getByRole("button", { name: "Add fault" }).click();
  await panel.getByLabel("Fault 1 instance").selectOption("CT-1");
  await panel.getByLabel("Fault 1 fault").selectOption({ label: "fan failure" });
  await panel.getByLabel("Fault 1 starts at minute").fill("5");
  await panel.getByRole("button", { name: "Start the run" }).click();
  const node = page.locator(".react-flow__node-instance").filter({ hasText: "CT-1" });
  await expect(node).not.toContainText("Fault", { timeout: 10_000 });
  await expect(node).toContainText("Fault · fan failure", { timeout: 30_000 });
  await expect(panel.getByText(/^Running · 00:(0[5-9]|1\d)/)).toBeVisible();
  await expect(page.getByRole("heading", { level: 1, name: /^Run w-/ })).toBeVisible({ timeout: 60_000 });
});
