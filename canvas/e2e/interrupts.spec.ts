import AxeBuilder from "@axe-core/playwright";
import { execFileSync } from "node:child_process";
import { existsSync, readFileSync } from "node:fs";
import { join } from "node:path";
import type { Locator, Page } from "@playwright/test";
import { expect, test } from "./support/test";
import type { Bench } from "./support/session";
import { addThermalBench, runIdOf, scriptPolicy, shortScriptPolicy } from "./support/session";
import { boxOf, controlsOf, startedSession } from "./support/interruptBench";
import { repository } from "./support/workspace";

test.setTimeout(180_000);

let bench: Bench | null = null;

test.afterEach(() => {
  bench?.close();
  bench = null;
});

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
  await expect(controls.getByText("Released device fan-01.")).toBeVisible();
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
  await expect(controls.getByText(/^ref-a · expires in about .*\(run time\)$/)).toBeVisible();
  await approve.click();
  const waiting = controls.getByRole("group", { name: "Commands waiting for approval" });
  await expect(waiting.getByRole("heading")).toBeFocused();
  const table = page.getByRole("table", { name: "Command evidence" });
  const executed = table.getByRole("row", { name: /ref-a/ });
  await expect(executed).toContainText("Accepted", { timeout: 30_000 });
  await expect(executed).toContainText("Board reports success", { timeout: 30_000 });
  await expect(executed).not.toContainText("Policy not allowed");
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
  const started = await startedSession(page, lab, [scriptPolicy]);
  bench = started.bench;
  const { panel } = started;
  await expect(panel.getByRole("region", { name: "Allowed policies in effect" })).toContainText(shortScriptPolicy);
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
  const artifact = JSON.parse(readFileSync(join(lab.workspace, "runs", run, "run.json"), "utf8")) as {
    command_log: { op: string }[];
  };
  expect(artifact.command_log.some((command) => command.op === "effector.invoke")).toBe(true);
});
