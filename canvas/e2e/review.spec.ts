import AxeBuilder from "@axe-core/playwright";
import type { Page } from "@playwright/test";
import { expect, test } from "@playwright/test";
import { env } from "./support/env";
import { nodesOutsideCanvas } from "./support/fitAudit";
import { textClashes } from "./support/textClashes";

test.use({ storageState: env("CANVAS_APP_STATE") });
test.setTimeout(90_000);

async function openIntegrity(page: Page) {
  await page.getByRole("complementary", { name: "Run details" }).locator("summary").first().click();
}

test("a recorded run plays back with the binary's report and a Verified badge", async ({ page }) => {
  await page.goto(env("CANVAS_APP_URL"));
  await page.getByRole("button", { name: "Review w-evaluate" }).click();
  const header = page.getByRole("heading", { name: "Run w-evaluate" });
  await expect(header).toBeVisible();
  await expect(page.getByRole("status").filter({ hasText: /^Verified$/ })).toHaveCount(1, { timeout: 60_000 });
  await expect(page.getByText("Evaluate", { exact: true })).toBeVisible();
  await expect(page.getByText("seed 7 · 30 min")).toBeVisible();
  const report = page.getByRole("complementary", { name: "Run details" });
  await expect(report.locator("summary").first()).toHaveText("Verified: the files match and the replay reproduces.");
  await openIntegrity(page);
  await expect(report.getByRole("region", { name: "File check" })).toContainText("files match");
  await expect(report.getByRole("region", { name: "Replay check" })).toContainText("Reproduces");
  await expect(page.getByRole("status").filter({ hasText: "plant file changed after this run" })).toBeVisible();
  await expect(report.getByRole("region", { name: "True results" })).toContainText("CH-1");
  await expect(page.getByText(/\d+ observations/)).toBeVisible({ timeout: 60_000 });
  await page.getByRole("navigation", { name: "Outline" }).getByRole("button", { name: "AC-1" }).click();
  const observed = page.getByRole("region", { name: "Observed · AC-1" });
  await expect(observed.getByRole("row", { name: /alarm state/ })).toContainText(/normal|unacknowledged|acknowledged/);
  const chiller = page.getByRole("group", { name: /^CH-1, chiller,/ });
  await expect(chiller.locator("dl")).toContainText("power");
  const slider = page.getByRole("slider", { name: "Playback time" });
  await expect(slider).not.toHaveAttribute("aria-valuetext", "00:30:00");
  await page.getByRole("button", { name: "600×" }).click();
  await page.getByRole("button", { name: "Play" }).click();
  await expect(slider).toHaveAttribute("aria-valuetext", "00:30:00", { timeout: 15_000 });
  await expect(page.getByRole("button", { name: "Play" })).toBeVisible();
  await slider.fill("60");
  await expect(slider).toHaveAttribute("aria-valuetext", "00:01:00");
  await slider.focus();
  await page.keyboard.press("Space");
  await expect(page.getByRole("button", { name: "Pause" })).toBeVisible();
  await page.keyboard.press("Space");
  await expect(page.getByRole("button", { name: "Play" })).toBeVisible();
  expect((await new AxeBuilder({ page }).analyze()).violations).toEqual([]);
});

test("a run whose files were edited after it ended is Not verified", async ({ page }) => {
  await page.goto(`${env("CANVAS_APP_URL")}/#/run?world=w-edited`);
  await expect(page.getByRole("status").filter({ hasText: /^Not verified$/ })).toHaveCount(1, { timeout: 60_000 });
  await openIntegrity(page);
  await expect(
    page.getByRole("complementary", { name: "Run details" }).getByRole("region", { name: "File check" }),
  ).toContainText("files differ");
  await expect(page.getByRole("alert").filter({ hasText: "differ from what the run wrote" })).toBeVisible();
  await expect(
    page.getByRole("complementary", { name: "Run details" }).getByRole("region", { name: "Replay check" }),
  ).toContainText("Does not verify: trace.jsonl differs");
  const graph = await page.getByRole("application", { name: "Plant graph" }).boundingBox();
  expect(graph?.height ?? 0).toBeGreaterThan(300);
});

test("a run that does not reproduce reads Not verified with verify's own verdict and what differs", async ({
  page,
}) => {
  await page.goto(`${env("CANVAS_APP_URL")}/#/run?world=w-diverged`);
  await expect(page.getByRole("status").filter({ hasText: /^Not verified$/ })).toHaveCount(1, { timeout: 60_000 });
  await openIntegrity(page);
  const replay = page.getByRole("complementary", { name: "Run details" }).getByRole("region", { name: "Replay check" });
  await expect(replay).toContainText(/Does not (reproduce|verify)/);
  await replay.getByText("What differs").click();
  await expect(replay.locator("pre")).not.toBeEmpty();
  await expect(page.getByRole("note")).toContainText("This run did not verify");
});

test("a run whose recorded plant was tampered with reads Not verified, not Not checked", async ({ page }) => {
  await page.goto(`${env("CANVAS_APP_URL")}/#/run?world=w-tampered`);
  await expect(page.getByRole("status").filter({ hasText: /^Not verified$/ })).toHaveCount(1, { timeout: 60_000 });
  await openIntegrity(page);
  const replay = page.getByRole("complementary", { name: "Run details" }).getByRole("region", { name: "Replay check" });
  await replay.getByText("What differs").click();
  await expect(replay).toContainText("nothing was replayed");
  await expect(replay.locator("pre")).toContainText("does not match its digest");
});

test("a hall carries its faulted members, and inside it each names its fault", async ({ page }) => {
  await page.goto(`${env("CANVAS_APP_URL")}/#/run?world=w-hall`);
  await expect(page.getByRole("status").filter({ hasText: /^Verified$/ })).toHaveCount(1, { timeout: 60_000 });
  await page.getByRole("slider", { name: "Playback time" }).fill("1800");
  const hall = page.getByRole("group", { name: /^h02, Fault in 2, / });
  await expect(hall).toContainText("Fault 2");
  await expect(page.getByRole("group", { name: /^h01, / })).not.toContainText("Fault");
  await page.getByRole("navigation", { name: "Outline" }).getByRole("button", { name: "h02", exact: true }).click();
  await expect(page.getByRole("heading", { level: 1, name: "h02" })).toBeVisible();
  await expect(page.getByRole("group", { name: /^M-1, Fault: product jam, / })).toContainText("product jam");
  await page.getByRole("button", { name: "Run w-hall" }).click();
  await expect(page.getByRole("heading", { level: 1, name: "Run w-hall" })).toBeVisible();
});

test("a review node lists every port of its kind as a row, observed or not, at each tier without a text clash", async ({
  page,
}) => {
  await page.goto(`${env("CANVAS_APP_URL")}/#/run?world=w-evaluate`);
  await expect(page.getByText(/\d+ observations/)).toBeVisible({ timeout: 60_000 });
  const chiller = page.getByRole("group", { name: /^CH-1, chiller,/ });
  const label = (await chiller.getAttribute("aria-label")) ?? "";
  const [, inputs = "0", outputs = "0"] = /(\d+) inputs?, (\d+) outputs?/.exec(label) ?? [];
  const ports = Number(inputs) + Number(outputs);
  const rows = await chiller.locator("dl > div").count();
  expect([ports, ports + 1]).toContain(rows);
  await expect(chiller.locator("dl")).toContainText("not observed");
  const graph = page.getByRole("application", { name: "Plant graph" });
  await graph.focus();
  for (let step = 0; step < 3; step += 1) {
    expect(await page.evaluate(textClashes)).toEqual([]);
    await page.keyboard.press("+");
  }
});

function transformOf(): string | undefined {
  return document.querySelector<HTMLElement>(".react-flow__viewport")?.style.transform;
}

function readoutPixels(): { readonly zoom: number; readonly shown: number; readonly smallest: number } {
  const graph = document.querySelector("[aria-roledescription='plant graph']");
  const zoom = Number(graph === null ? 0 : getComputedStyle(graph).getPropertyValue("--zoom"));
  const values = [...document.querySelectorAll(".react-flow__node-instance dd")].filter(
    (value) => value.getClientRects().length > 0 && value.textContent.trim() !== "",
  );
  const sizes = values.map((value) => parseFloat(getComputedStyle(value).fontSize) * zoom);
  return { zoom, shown: values.length, smallest: Math.min(...sizes) };
}

test("a review whose plant fits at 40 % or more opens fitted, and Fit returns the same view", async ({ page }) => {
  await page.setViewportSize({ width: 1920, height: 1080 });
  await page.goto(`${env("CANVAS_APP_URL")}/#/run?world=w-evaluate`);
  await expect(page.getByText(/\d+ observations/)).toBeVisible({ timeout: 60_000 });
  await expect(page.getByRole("group", { name: /^CH-1, chiller,/ })).toBeAttached();
  await expect.poll(() => page.evaluate(nodesOutsideCanvas)).toEqual([]);
  const opening = await page.evaluate(transformOf);
  await page.getByRole("button", { name: "Fit the plant in view" }).click();
  await expect.poll(() => page.evaluate(transformOf)).toBe(opening);
});

test("a review of a plant too large for 40 % opens at 40 % on its first fault, with readouts of 8 px or more", async ({
  page,
}) => {
  await page.setViewportSize({ width: 1280, height: 720 });
  await page.goto(`${env("CANVAS_APP_URL")}/#/run?world=w-late`);
  await expect(page.getByText(/\d+ observations/)).toBeVisible({ timeout: 60_000 });
  const dryer = page.getByRole("group", { name: /^D-1, / });
  await expect(dryer).toBeAttached();
  const canvas = page.locator(".react-flow");
  await expect
    .poll(async () => {
      const box = await dryer.boundingBox();
      const frame = await canvas.boundingBox();
      if (box === null || frame === null) return false;
      return (
        box.x >= frame.x &&
        box.y >= frame.y &&
        box.x + box.width <= frame.x + frame.width &&
        box.y + box.height <= frame.y + frame.height
      );
    })
    .toBe(true);
  const opened = await page.evaluate(readoutPixels);
  expect(opened.zoom).toBeGreaterThanOrEqual(0.4);
  expect(opened.shown).toBeGreaterThan(0);
  expect(opened.smallest).toBeGreaterThanOrEqual(7.99);
  expect(await page.evaluate(nodesOutsideCanvas)).not.toEqual([]);
  await page.getByRole("button", { name: "Fit the plant in view" }).click();
  await expect.poll(() => page.evaluate(nodesOutsideCanvas)).toEqual([]);
});
