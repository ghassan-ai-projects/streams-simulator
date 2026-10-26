import type { Locator } from "@playwright/test";
import { expect, test } from "./support/test";
import type { Bench } from "./support/session";
import {
  addThermalBench,
  openBench,
  openSessionPanel,
  planOneRealBoard,
  runIdOf,
  startGateway,
} from "./support/session";

test.setTimeout(90_000);

let bench: Bench | null = null;

test.afterEach(() => {
  bench?.close();
  bench = null;
});

async function expectKindMetric(box: Locator): Promise<void> {
  await expect(box.getByText("level percent", { exact: true })).toBeVisible();
  await expect(box.getByText("duty permille", { exact: true })).toBeHidden();
}

test("a real fan's box shows the kind's metric, live and in the saved review, not the twin's channel @slow", async ({
  page,
  lab,
}) => {
  addThermalBench(lab);
  const open = openBench(lab.binary);
  bench = open;
  const panel = await openSessionPanel(page);
  await planOneRealBoard(panel, open, "fan-01", "fan-01");
  await panel.getByRole("button", { name: "Start the session" }).click();
  await expect(panel.getByText(/^Connected session · run c-[0-9a-f]{8}$/)).toBeVisible();
  const fan = page.getByRole("group", { name: /^fan-01,/ });
  await expect(fan.getByText("duty permille", { exact: true })).toBeVisible();
  startGateway(open, "fan-01", "fan");
  await expect(panel.getByText(/announced every real board/)).toBeVisible({ timeout: 30_000 });
  await expectKindMetric(fan);
  await panel.getByRole("button", { name: "Stop and save" }).click();
  const saved = panel.getByText(/^The session was stopped and saved as run c-[0-9a-f]{8}\.$/);
  await expect(saved).toBeVisible({ timeout: 30_000 });
  const run = runIdOf(await saved.textContent());
  await panel.getByRole("button", { name: "Open the review" }).click();
  await expect(page.getByRole("heading", { level: 1, name: `Run ${run}` })).toBeVisible();
  await expectKindMetric(page.getByRole("group", { name: /^fan-01,/ }));
});
