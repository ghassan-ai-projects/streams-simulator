import AxeBuilder from "@axe-core/playwright";
import { readFileSync, statSync } from "node:fs";
import { join } from "node:path";
import type { Page } from "@playwright/test";
import { expect, test } from "./support/test";
import type { Bench } from "./support/session";
import {
  addThermalBench,
  listPolicies,
  openScriptedBench,
  openSessionPanel,
  planOneRealBoard,
  scriptPolicy,
  shortScriptPolicy,
  startGateway,
} from "./support/session";

test.setTimeout(180_000);

let bench: Bench | null = null;

test.afterEach(() => {
  bench?.close();
  bench = null;
});

const cooling = { level_percent: 30 };
const first = `sha256:${"a".repeat(64)}`;
const second = `sha256:${"c".repeat(64)}`;

async function fitsAt1280(page: Page): Promise<void> {
  await page.setViewportSize({ width: 1280, height: 800 });
  const details = page.getByRole("complementary", { name: "Run details" });
  await expect.poll(() => details.evaluate((aside) => aside.scrollWidth <= aside.clientWidth)).toBe(true);
  const scroller = page.getByRole("region", { name: "Command evidence, scrollable" });
  await expect.poll(() => scroller.evaluate((region) => region.scrollWidth <= region.clientWidth)).toBe(true);
  for (const name of ["Copy policy digest", "Use this digest"]) {
    const box = await scroller.getByRole("button", { name }).boundingBox();
    expect((box?.x ?? 0) + (box?.width ?? 2000)).toBeLessThanOrEqual(1280);
  }
}

test("a command refused for its policy shows the host's digest, copies it on request, and adds it to the next plan only when asked", async ({
  page,
  lab,
}) => {
  addThermalBench(lab);
  const open = openScriptedBench(lab.binary, [
    { id: "pol-a", afterS: 20, device: "fan-01", command: "set_cooling", args: cooling },
  ]);
  bench = open;
  const panel = await openSessionPanel(page);
  await planOneRealBoard(panel, open, "zone-01");
  await panel.getByRole("button", { name: "Check the plan" }).click();
  const inEffect = panel.getByRole("region", { name: "Allowed policies in effect" });
  await expect(inEffect).toContainText("Allowed policies: none. Commands to simulated devices are refused");
  await panel.getByRole("button", { name: "Start the session" }).click();
  await expect(panel.getByText(/^Connected session · run c-[0-9a-f]{8}$/)).toBeVisible();
  startGateway(open, "zone-01");
  await expect(panel.getByText(/announced every real board/)).toBeVisible({ timeout: 30_000 });
  await page.getByRole("button", { name: "Open live view", exact: true }).click();
  const row = page.getByRole("table", { name: "Command evidence" }).getByRole("row", { name: /pol-a/ });
  await expect(row).toContainText("Refused", { timeout: 120_000 });
  await expect(row).toContainText("Policy not allowed by this session");
  await expect(row.getByText(shortScriptPolicy, { exact: true })).toBeVisible();
  await expect(page.getByText("Allowed policies: none.")).toBeVisible();

  await page.context().grantPermissions(["clipboard-read", "clipboard-write"], { origin: new URL(lab.url).origin });
  await row.getByRole("button", { name: "Copy policy digest" }).click();
  await expect(row.getByText("Copied the policy digest of pol-a.")).toBeVisible();
  await expect.poll(() => page.evaluate(() => navigator.clipboard.readText())).toBe(scriptPolicy);
  await row.getByRole("button", { name: "Use this digest" }).click();
  await expect(row.getByText(/draft session plan of \S+\. It applies to the next session/)).toBeVisible();
  expect((await new AxeBuilder({ page }).include("aside").analyze()).violations).toEqual([]);
  await fitsAt1280(page);

  await page.getByRole("button", { name: "Open plant", exact: true }).click();
  await page.getByRole("button", { name: "Run plant", exact: true }).click();
  await expect(panel.getByRole("region", { name: "Allowed policies in effect" })).toContainText(
    "Allowed policies: none.",
  );
  await panel.getByRole("button", { name: "Stop and save" }).click();
  await expect(panel.getByText(/^The session was stopped and saved as run c-[0-9a-f]{8}\.$/)).toBeVisible({
    timeout: 30_000,
  });
  await panel.getByRole("button", { name: "Set up another run" }).click();
  await panel.getByRole("radio", { name: /^Connected session/ }).check();
  await expect(panel.getByLabel("Allowed policy digests", { exact: true })).toHaveValue(scriptPolicy);
  await expect(panel.getByText("Kept from the previous run: 1 digest.")).toBeVisible();
  await expect(panel.getByLabel("Host socket", { exact: true })).toHaveValue("");
  await expect(panel.getByRole("radio", { name: /^Connected session/ })).toBeChecked();
});

test("Save as default writes the binary's sorted list after a passing check, and a new form after a reload is pre-filled and says so", async ({
  page,
  lab,
}) => {
  addThermalBench(lab);
  await page.setViewportSize({ width: 1280, height: 800 });
  const panel = await openSessionPanel(page);
  await panel.getByLabel("Host socket", { exact: true }).fill("/tmp/ssc-default/h.sock");
  await listPolicies(panel, [second, first]);
  await expect(panel.getByRole("button", { name: "Save as default" })).toHaveCount(0);
  await expect(panel.getByText(/^Trusting 2 policy digests you entered: /)).toBeVisible();
  await panel.getByRole("button", { name: "Check the plan" }).click();
  const result = panel.getByRole("region", { name: "Plan check result" });
  await expect(result.getByRole("region", { name: "Allowed policies in effect" })).toContainText("Allowed policies: 2");
  await expect(result).toBeFocused();
  await expect(result).toBeInViewport();
  const save = panel.getByRole("button", { name: "Save as default" });
  await expect(save).toBeInViewport({ ratio: 1 });
  await save.click();
  const saved = panel.getByText("Saved 2 policy digests as the workspace default.");
  await expect(saved).toBeVisible();
  await expect(saved).toBeInViewport({ ratio: 1 });
  const file = join(lab.workspace, "settings", "default.policies.json");
  expect(JSON.parse(readFileSync(file, "utf8")) as unknown).toEqual({ policies: [first, second] });
  expect(statSync(file).mode % 0o1000).toBe(0o600);
  expect((await new AxeBuilder({ page }).include("aside").analyze()).violations).toEqual([]);

  await page.setViewportSize({ width: 1280, height: 800 });
  await expect.poll(() => panel.evaluate((aside) => aside.scrollWidth <= aside.clientWidth)).toBe(true);

  await page.reload();
  const fresh = await openSessionPanel(page);
  const field = fresh.getByLabel("Allowed policy digests", { exact: true });
  await expect(field).toHaveValue(`${first}\n${second}`);
  await expect(fresh.getByText("Pre-filled from the workspace default.")).toBeVisible();
  const trusting = fresh.getByText(/^Trusting 2 policy digests from the workspace default: /);
  await expect(trusting).toContainText("sha256:aaaaaaaa…aaaaaaaa, sha256:cccccccc…cccccccc");
  await expect(fresh.getByRole("button", { name: "Start the session" })).toHaveAccessibleDescription(
    /Trusting 2 policy digests/,
  );
  await fresh.getByRole("button", { name: "Show the allowed policy digests" }).click();
  await expect(field).toBeFocused();
  await field.fill(first);
  await expect(fresh.getByText(/^Trusting 1 policy digest you entered: /)).toBeVisible();
  await expect(fresh.getByText("Pre-filled from the workspace default.")).toBeHidden();
  await fresh.getByRole("button", { name: "Clear the list" }).click();
  await expect(field).toHaveValue("");
  await expect(field).toBeFocused();
});

test("the binary's refusals of the list sit on the field, by digest, and the plan is not started", async ({
  page,
  lab,
}) => {
  addThermalBench(lab);
  const panel = await openSessionPanel(page);
  await panel.getByLabel("Host socket", { exact: true }).fill("/tmp/ssc-default/h.sock");
  const field = panel.getByLabel("Allowed policy digests", { exact: true });
  await listPolicies(panel, [first, "nope"]);
  await panel.getByRole("button", { name: "Start the session" }).click();
  await expect(field).toHaveAttribute("aria-invalid", "true");
  await expect(field).toBeFocused();
  await expect(field).toHaveAccessibleDescription(
    /Digest 2: A policy digest is sha256: and 64 lower-case hex digits\./,
  );
  await listPolicies(panel, [first, first]);
  await panel.getByRole("button", { name: "Check the plan" }).click();
  await expect(field).toHaveAccessibleDescription(/Digest 2: This digest is already listed\./);
  await expect(panel.getByText(/^Connected session · run/)).toHaveCount(0);
  await expect(panel.getByRole("button", { name: "Save as default" })).toHaveCount(0);
});
