import AxeBuilder from "@axe-core/playwright";
import { execFileSync } from "node:child_process";
import { join } from "node:path";
import type { Locator } from "@playwright/test";
import { expect, test } from "./support/test";
import type { Bench } from "./support/session";
import { addThermalBench, openBench, openSessionPanel, runIdOf, startGateway } from "./support/session";
import { repository } from "./support/workspace";

test.setTimeout(90_000);

let bench: Bench | null = null;

test.afterEach(() => {
  bench?.close();
  bench = null;
});

async function planOneRealBoard(panel: Locator, open: Bench, device: string): Promise<void> {
  await panel.getByLabel("Host socket", { exact: true }).fill(open.host);
  await panel.getByLabel("Gateway socket", { exact: true }).fill(open.gateway);
  await panel.getByRole("button", { name: "Add a real board" }).click();
  await panel.getByLabel("Real board 1 instance", { exact: true }).selectOption("zone-01");
  await panel.getByLabel("Real board 1 board", { exact: true }).fill("uno-07");
  await panel.getByLabel("Real board 1 device", { exact: true }).fill(device);
}

test("a connected session links the plant to a host and a real board, and stops into a run that verifies", async ({
  page,
  lab,
}) => {
  addThermalBench(lab);
  const open = openBench(lab.binary);
  bench = open;
  const panel = await openSessionPanel(page);
  await planOneRealBoard(panel, open, "zone-01");
  await panel.getByRole("button", { name: "Check the plan" }).click();
  const instances = panel.getByRole("region", { name: "Instances in this session" });
  await expect(
    instances.getByText("zone-01 · Real board · device zone-01 on board uno-07", { exact: true }),
  ).toBeVisible();
  await expect(instances.getByText("fan-01 · Simulated", { exact: true })).toBeVisible();
  expect((await new AxeBuilder({ page }).include("aside").analyze()).violations).toEqual([]);
  await panel.getByRole("button", { name: "Start the session" }).click();
  await expect(panel.getByText(/^Connected session · run c-[0-9a-f]{8}$/)).toBeVisible();
  startGateway(open, "zone-01");
  await expect(panel.getByText(/announced every real board of the plan/)).toBeVisible({ timeout: 30_000 });
  await panel.getByRole("button", { name: "Stop and save" }).click();
  const saved = panel.getByText(/^The session was stopped and saved as run c-[0-9a-f]{8}\.$/);
  await expect(saved).toBeVisible({ timeout: 30_000 });
  const run = runIdOf(await saved.textContent());
  const artifact = join(lab.workspace, "runs", run, "run.json");
  const verified = execFileSync(lab.binary, ["verify", artifact, "--domains-dir", join(lab.workspace, "domains")], {
    cwd: repository,
  });
  expect(JSON.parse(verified.toString()) as unknown).toMatchObject({ matches: true });
  await panel.getByRole("button", { name: "Open the review" }).click();
  await expect(page.getByRole("heading", { level: 1, name: `Run ${run}` })).toBeVisible();
  const outline = page.getByRole("navigation", { name: "Outline" });
  await expect(outline.getByRole("button", { name: /^zone-01.*Real · board uno-07 · boot 61$/ })).toBeVisible();
  await expect(outline.getByRole("button", { name: /^fan-01.*Simulated · board mega-01$/ })).toBeVisible();
  await expect(page.getByRole("group", { name: /^zone-01,.*Real · board uno-07 · boot 61/ })).toBeVisible();
  await outline.getByRole("button", { name: /^zone-01/ }).click();
  const observed = page.getByRole("region", { name: "Observed · zone-01" });
  await expect(observed.getByText(/This instance is a real board/)).toBeVisible();
});

test("a gateway that does not announce the bound device is refused, and the panel says which rule", async ({
  page,
  lab,
}) => {
  addThermalBench(lab);
  const open = openBench(lab.binary);
  bench = open;
  const panel = await openSessionPanel(page);
  await planOneRealBoard(panel, open, "zone-09");
  await panel.getByRole("button", { name: "Start the session" }).click();
  await expect(panel.getByText(/^Connected session · run c-/)).toBeVisible();
  startGateway(open, "zone-01");
  const refused = panel.getByRole("alert");
  await expect(refused).toContainText("The gateway was refused (inventory_mismatch)", { timeout: 30_000 });
  await expect(refused).toContainText('holds no device "zone-09"');
  await panel.getByRole("button", { name: "Stop and save" }).click();
  await expect(panel.getByText(/^The session was stopped and saved as run c-/)).toBeVisible({ timeout: 30_000 });
});

test("a real board without a gateway is refused by the binary at the gateway socket field", async ({ page, lab }) => {
  addThermalBench(lab);
  const panel = await openSessionPanel(page);
  await panel.getByLabel("Host socket", { exact: true }).fill("/tmp/ssc-none/h.sock");
  await panel.getByRole("button", { name: "Add a real board" }).click();
  await panel.getByLabel("Real board 1 board", { exact: true }).fill("uno-07");
  await panel.getByLabel("Real board 1 device", { exact: true }).fill("zone-01");
  await panel.getByRole("button", { name: "Check the plan" }).click();
  const gateway = panel.getByLabel("Gateway socket", { exact: true });
  await expect(gateway).toHaveAttribute("aria-invalid", "true");
  await expect(gateway).toHaveAccessibleDescription(/the plan needs gateway_socket/);
});
