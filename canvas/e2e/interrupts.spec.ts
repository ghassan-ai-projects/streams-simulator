import AxeBuilder from "@axe-core/playwright";
import { execFileSync } from "node:child_process";
import { existsSync } from "node:fs";
import { join } from "node:path";
import type { Locator, Page } from "@playwright/test";
import { expect, test } from "./support/test";
import type { Bench } from "./support/session";
import {
  addThermalBench,
  openScriptedBench,
  openSessionPanel,
  planOneRealBoard,
  runIdOf,
  startGateway,
} from "./support/session";
import { repository } from "./support/workspace";

test.setTimeout(180_000);

let bench: Bench | null = null;

test.afterEach(() => {
  bench?.close();
  bench = null;
});

const lease = { duty_permille: 300, lease_ms: 5000 };

function controlsOf(page: Page): Locator {
  return page.getByRole("region", { name: "Interrupts", exact: true });
}

async function startedSession(page: Page, lab: { readonly binary: string }): Promise<Locator> {
  const open = openScriptedBench(lab.binary, [
    { id: "ref-a", afterS: 200, device: "fan-01", command: "set_pwm_lease", args: lease },
    { id: "ref-b", afterS: 240, device: "fan-01", command: "set_pwm_lease", args: lease },
  ]);
  bench = open;
  const panel = await openSessionPanel(page);
  await planOneRealBoard(panel, open, "zone-01");
  await panel.getByRole("button", { name: "Start the session" }).click();
  await expect(panel.getByText(/^Connected session · run c-[0-9a-f]{8}$/)).toBeVisible();
  startGateway(open, "zone-01");
  await expect(panel.getByText(/announced every real board/)).toBeVisible({ timeout: 30_000 });
  return panel;
}

async function holdAndReleaseOnThePlant(page: Page): Promise<void> {
  const controls = controlsOf(page);
  await expect(controls.getByText("No holds or approval modes are set.")).toBeVisible({ timeout: 30_000 });
  await controls.getByLabel("Scope", { exact: true }).selectOption({ label: "Device fan-01" });
  const hold = controls.getByRole("button", { name: "Hold", exact: true });
  await hold.click();
  await expect(controls.getByText("Hold on device fan-01 was applied.")).toBeVisible();
  await expect(hold).toHaveAttribute("aria-pressed", "true");
  const active = controls.getByRole("list", { name: "Active interrupts" });
  await expect(active.getByText("Device fan-01: Held", { exact: true })).toBeVisible();
  await expect(page.getByRole("group", { name: /^fan-01,.*Held/ })).toBeVisible();
  await controls.getByRole("button", { name: "Release", exact: true }).click();
  await expect(controls.getByText("Release of device fan-01 was applied.")).toBeVisible();
  await expect(controls.getByText("No holds or approval modes are set.")).toBeVisible();
}

async function approveAndDeny(page: Page, controls: Locator): Promise<void> {
  await controls.getByLabel("Scope", { exact: true }).selectOption({ label: "All devices" });
  const each = controls.getByRole("button", { name: "Approve each command", exact: true });
  await each.click();
  await expect(controls.getByText("Approve each command on all devices was applied.")).toBeVisible();
  await expect(each).toHaveAttribute("aria-pressed", "true");
  const approve = controls.getByRole("button", { name: "Approve ref-a", exact: true });
  await expect(approve).toBeVisible({ timeout: 90_000 });
  await expect(controls.getByText(/^ref-a · expires in .*\(run time\)$/)).toBeVisible();
  await approve.click();
  const waiting = controls.getByRole("group", { name: "Commands waiting for approval" });
  await expect(waiting.getByRole("heading")).toBeFocused();
  const table = page.getByRole("table", { name: "Command evidence" });
  await expect(table.getByRole("row", { name: /ref-a/ })).toContainText("Refused");
  await expect(table.getByRole("row", { name: /ref-a/ })).toContainText("Policy not allowed");
  const deny = controls.getByRole("button", { name: "Deny ref-b", exact: true });
  await expect(deny).toBeVisible({ timeout: 90_000 });
  await deny.click();
  await expect(table.getByRole("row", { name: /ref-b/ })).toContainText("Refused at the node: local control");
  await each.click();
  await expect(controls.getByText("Stop approving each command on all devices was applied.")).toBeVisible();
}

async function emergencyStopAndRelease(page: Page, controls: Locator): Promise<void> {
  await controls.getByRole("button", { name: "Emergency stop", exact: true }).click();
  await expect(controls.getByText(/^Emergency stop on all devices was applied\./)).toBeVisible();
  await expect(controls.getByText(/The scope stays held until it is released\.$/)).toBeVisible();
  const active = controls.getByRole("list", { name: "Active interrupts" });
  await expect(active.getByText("All devices: Held · latched by emergency stop", { exact: true })).toBeVisible();
  await expect(page.getByRole("group", { name: /^fan-01,.*latched by emergency stop/ })).toBeVisible();
  await expect(page.getByRole("list", { name: "Event types" }).getByText("Interrupt control")).toBeVisible();
  await controls.getByRole("button", { name: "Release", exact: true }).click();
  await expect(controls.getByText("No holds or approval modes are set.")).toBeVisible();
}

interface Box {
  readonly x: number;
  readonly y: number;
  readonly width: number;
  readonly height: number;
}

function boxOf(box: Box | null): Box {
  return box ?? { x: 0, y: 0, width: 0, height: 0 };
}

async function fitsAt1280(page: Page, controls: Locator): Promise<void> {
  await page.setViewportSize({ width: 1280, height: 800 });
  await expect.poll(() => controls.evaluate((region) => region.scrollWidth <= region.clientWidth)).toBe(true);
  const region = boxOf(await controls.boundingBox());
  expect(region.y + region.height).toBeLessThanOrEqual(800);
  for (const name of ["Hold", "Release", "Approve each command", "Emergency stop"]) {
    const box = boxOf(await controls.getByRole("button", { name, exact: true }).boundingBox());
    expect(box.x).toBeGreaterThanOrEqual(region.x);
    expect(box.x + box.width).toBeLessThanOrEqual(1280);
  }
}

test("the interrupt controls hold, approve and stop the node of a live session, and the node says so", async ({
  page,
  lab,
}) => {
  addThermalBench(lab);
  const panel = await startedSession(page, lab);
  await holdAndReleaseOnThePlant(page);
  await page.getByRole("button", { name: "Open live view", exact: true }).click();
  await expect(page.getByText("Live · following newest record", { exact: true })).toBeVisible();
  const controls = controlsOf(page);
  await approveAndDeny(page, controls);
  await emergencyStopAndRelease(page, controls);
  expect((await new AxeBuilder({ page }).include("[aria-label='Interrupts']").analyze()).violations).toEqual([]);
  await fitsAt1280(page, controls);
  await page.getByRole("button", { name: "Open plant", exact: true }).click();
  await page.getByRole("button", { name: "Run plant", exact: true }).click();
  await panel.getByRole("button", { name: "Stop and save" }).click();
  const saved = panel.getByText(/^The session was stopped and saved as run c-[0-9a-f]{8}\.$/);
  await expect(saved).toBeVisible({ timeout: 30_000 });
  const run = runIdOf(await saved.textContent());
  expect(existsSync(join(lab.workspace, "runs", run, "interrupts.jsonl"))).toBe(true);
  const verified = execFileSync(
    lab.binary,
    ["verify", join(lab.workspace, "runs", run, "run.json"), "--domains-dir", join(lab.workspace, "domains")],
    {
      cwd: repository,
    },
  );
  expect(JSON.parse(verified.toString()) as unknown).toMatchObject({ matches: true });
});
