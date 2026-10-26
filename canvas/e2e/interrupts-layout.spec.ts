import type { Locator, Page } from "@playwright/test";
import { expect, test } from "./support/test";
import type { Bench } from "./support/session";
import { addThermalBench, openScriptedBench, planOneRealBoard, startGateway } from "./support/session";
import { guard } from "./support/audit";
import { nodesOutsideCanvas } from "./support/fitAudit";
import {
  boxOf,
  controlsOf,
  cooling,
  refsAfterApproveMode,
  runOfSession,
  startedSession,
} from "./support/interruptBench";

test.setTimeout(240_000);

const benches: Bench[] = [];

test.afterEach(() => {
  benches.splice(0).forEach((bench) => {
    bench.close();
  });
});

const height = 800;

async function insideTheWindow(controls: Locator): Promise<void> {
  await expect
    .poll(async () => {
      const region = boxOf(await controls.boundingBox());
      return region.y >= 0 && region.y + region.height <= height;
    })
    .toBe(true);
}

async function oneRow(controls: Locator): Promise<void> {
  const hold = boxOf(await controls.getByRole("button", { name: "Hold", exact: true }).boundingBox());
  const stop = boxOf(await controls.getByRole("button", { name: "Emergency stop", exact: true }).boundingBox());
  expect(Math.abs(stop.y - hold.y)).toBeLessThanOrEqual(4);
}

async function plantInView(page: Page): Promise<void> {
  const plant = boxOf(await page.locator(".react-flow").boundingBox());
  expect(plant.height).toBeGreaterThan(0);
  expect(plant.y + plant.height).toBeLessThanOrEqual(height);
}

test("the interrupt bar stays inside a 1280 by 800 window in one row, and the plant's bottom stays in view, while a command waits and a stop latches", async ({
  page,
  lab,
}) => {
  await page.setViewportSize({ width: 1280, height });
  addThermalBench(lab);
  const started = await startedSession(page, lab, refsAfterApproveMode);
  benches.push(started.bench);
  await page.getByRole("button", { name: "Open live view", exact: true }).click();
  const controls = controlsOf(page);
  await expect(controls.getByText("No holds or approval modes are set.")).toBeVisible({ timeout: 30_000 });
  await insideTheWindow(controls);
  await oneRow(controls);
  await plantInView(page);

  await controls.getByRole("button", { name: "Approve each command", exact: true }).click();
  await expect(controls.getByRole("button", { name: "Approve ref-a", exact: true })).toBeVisible({ timeout: 90_000 });
  await page.mouse.wheel(0, 3000);
  await insideTheWindow(controls);
  await oneRow(controls);
  await plantInView(page);
  const waiting = controls.getByRole("listitem").filter({ hasText: "ref-a" }).getByText("set_cooling on fan-01");
  await expect.poll(() => waiting.evaluate((line) => line.scrollWidth <= line.clientWidth)).toBe(true);

  const stop = controls.getByRole("button", { name: "Emergency stop", exact: true });
  const hold = controls.getByRole("button", { name: "Hold", exact: true });
  const stopBefore = boxOf(await stop.boundingBox());
  await hold.click();
  await expect(hold).toHaveAttribute("aria-pressed", "true");
  await expect(controls.getByRole("button", { name: "Approve each command", exact: true })).toHaveAttribute(
    "aria-pressed",
    "true",
  );
  await oneRow(controls);
  const stopHeld = boxOf(await stop.boundingBox());
  expect(Math.abs(stopHeld.y - stopBefore.y)).toBeLessThanOrEqual(1);
  expect(Math.abs(stopHeld.x - stopBefore.x)).toBeLessThanOrEqual(1);
  await plantInView(page);
  await expect.poll(async () => guard(await page.evaluate(nodesOutsideCanvas, 1))).toEqual([]);
  await controls.getByRole("button", { name: "Release", exact: true }).click();
  await expect(hold).toHaveAttribute("aria-pressed", "false");

  await stop.click();
  await expect(controls.getByText(/The scope stays held until it is released\./)).toBeVisible();
  await insideTheWindow(controls);
  await oneRow(controls);
  await plantInView(page);
  await controls.getByRole("button", { name: "Release", exact: true }).click();
  await expect(hold).toHaveAttribute("aria-pressed", "false");
  await controls.getByRole("button", { name: "Approve each command", exact: true }).click();
  await expect(controls.getByText("No holds or approval modes are set.")).toBeVisible();
  await plantInView(page);
});

test("a running session opened by its address, or after a reload, shows its plant and its controls", async ({
  page,
  lab,
}) => {
  addThermalBench(lab);
  const started = await startedSession(page, lab);
  benches.push(started.bench);
  const run = runOfSession(await started.panel.getByText(/^Connected session · run c-[0-9a-f]{8}$/).textContent());
  await page.goto(`/#/review?world=${run}`);
  await page.reload();
  await expect(page.getByRole("heading", { level: 1, name: `Run ${run}` })).toBeVisible();
  await expect(page.getByRole("navigation", { name: "Outline" }).getByRole("button", { name: /^zone-01/ })).toBeVisible(
    { timeout: 30_000 },
  );
  await expect(page.getByText("Opening the plant this run used…")).toBeHidden();
  await expect(controlsOf(page)).toBeVisible();
});

test("a second tab on the same running session shows its plant while the first stays open", async ({ page, lab }) => {
  addThermalBench(lab);
  const started = await startedSession(page, lab);
  benches.push(started.bench);
  const run = runOfSession(await started.panel.getByText(/^Connected session · run c-[0-9a-f]{8}$/).textContent());
  await page.goto(`/#/review?world=${run}`);
  await expect(page.getByRole("navigation", { name: "Outline" }).getByRole("button", { name: /^zone-01/ })).toBeVisible(
    {
      timeout: 30_000,
    },
  );
  const second = await page.context().newPage();
  await second.goto(`/#/review?world=${run}`);
  await expect(
    second.getByRole("navigation", { name: "Outline" }).getByRole("button", { name: /^zone-01/ }),
  ).toBeVisible({
    timeout: 30_000,
  });
  await expect(second.getByText("Opening the plant this run used…")).toBeHidden();
  await expect(controlsOf(second)).toBeVisible();
});

async function holdFanAndStop(panel: Locator, controls: Locator): Promise<string> {
  await controls.getByLabel("Scope", { exact: true }).selectOption({ label: "Device fan-01" });
  await controls.getByRole("button", { name: "Hold", exact: true }).click();
  await expect(controls.getByText("Device fan-01: Held", { exact: true })).toBeVisible();
  const status = panel.getByText(/^Connected session · run c-[0-9a-f]{8}$/);
  const run = runOfSession(await status.textContent());
  await panel.getByRole("button", { name: "Stop and save" }).click();
  await expect(panel.getByText(/^The session was stopped and saved as run c-[0-9a-f]{8}\.$/)).toBeVisible({
    timeout: 30_000,
  });
  return run;
}

test("a second session starts with none of the first session's holds, scope or outcome, without a reload", async ({
  page,
  lab,
}) => {
  addThermalBench(lab);
  const started = await startedSession(page, lab);
  benches.push(started.bench);
  const { panel } = started;
  const controls = controlsOf(page);
  await expect(controls.getByText("No holds or approval modes are set.")).toBeVisible({ timeout: 30_000 });
  const first = await holdFanAndStop(panel, controls);

  await panel.getByRole("button", { name: "Set up another run" }).click();
  const next = openScriptedBench(lab.binary, [
    { id: "ref-c", afterS: 300, device: "fan-01", command: "set_cooling", args: cooling },
  ]);
  benches.push(next);
  await expect(panel.getByLabel("Host socket", { exact: true })).toHaveValue("");
  await planOneRealBoard(panel, next, "zone-01");
  await panel.getByRole("button", { name: "Start the session" }).click();
  const status = panel.getByText(/^Connected session · run c-[0-9a-f]{8}$/);
  await expect(status).toBeVisible();
  expect(runOfSession(await status.textContent())).not.toBe(first);
  await expect(controls.getByText("Device fan-01: Held", { exact: true })).toHaveCount(0);
  await expect(controls.getByText(/Hold on device fan-01 was applied\./)).toHaveCount(0);
  await expect(controls.getByLabel("Scope", { exact: true })).toHaveValue("all:");
  startGateway(next, "zone-01");
  await expect(panel.getByText(/announced every real board/)).toBeVisible({ timeout: 30_000 });
  await expect(controls.getByText("No holds or approval modes are set.")).toBeVisible({ timeout: 30_000 });
});
