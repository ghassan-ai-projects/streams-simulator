import type { Page } from "@playwright/test";
import { expect, test } from "./support/test";

test.setTimeout(120_000);

const plant = "/#/plant?domain=domains%2Ffactory-utilities.domain.json&launch=1";

function panelOf(page: Page) {
  return page.getByRole("complementary", { name: "Run Factory utilities" });
}

async function action(page: Page, name: string, args: Record<string, unknown>): Promise<Record<string, unknown>> {
  const response = await page.request.post(`/api/actions/${name}`, { data: { args } });
  const body = (await response.json()) as { result?: Record<string, unknown> };
  if (!response.ok() || body.result === undefined) throw new Error(`${name} failed: ${await response.text()}`);
  return body.result;
}

async function setUpAnother(page: Page) {
  await page.getByRole("button", { name: "Set up another run" }).click();
  await expect(panelOf(page).getByLabel("Seed", { exact: true })).toBeVisible();
  return panelOf(page);
}

test("another run set up from a finished run repeats its seed, length, pace, fault and effector call", async ({
  page,
}) => {
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
  await expect(page.getByRole("slider", { name: "Playback time" })).toHaveAttribute("max", "300", {
    timeout: 60_000,
  });

  const again = await setUpAnother(page);
  await expect(again.getByLabel("Seed", { exact: true })).toHaveValue("11");
  await expect(again.getByLabel("Duration", { exact: true })).toHaveValue("5");
  await expect(again.getByLabel("Duration unit")).toHaveValue("minutes");
  await expect(again.getByRole("radio", { name: /^Explore/ })).toBeChecked();
  await expect(again.getByRole("radio", { name: /As fast as possible/ })).toBeChecked();
  await expect(again.getByLabel("Fault 1 instance")).toHaveValue("CT-1");
  await expect(again.getByLabel("Fault 1 fault")).toContainText("fan failure");
  await expect(again.getByLabel("Fault 1 starts at minute")).toHaveValue("2");
  await expect(again.getByLabel("Effector call 1 instance")).toHaveValue("CH-1");
  await expect(again.getByLabel("Effector call 1 effector")).toContainText("set setpoint");
  await expect(again.getByLabel("Effector call 1 setpoint c")).toHaveValue("7");
  await expect(again.getByLabel("Effector call 1 starts at minute")).toHaveValue("1");
  const note = again.getByRole("note");
  await expect(note).toContainText("Prefilled from w-");
  await expect(note).not.toContainText("Not copied");
});

test("another run set up from a stopped run repeats the length it was asked for, not where it stopped", async ({
  page,
}) => {
  await page.goto(plant);
  const panel = panelOf(page);
  await panel.getByLabel("Seed", { exact: true }).fill("11");
  await panel.getByLabel("Duration", { exact: true }).fill("30");
  await panel.getByRole("button", { name: "Start the run" }).click();
  await expect(panel.getByText(/^Running · 00:0[1-9]/)).toBeVisible({ timeout: 30_000 });
  await panel.getByRole("button", { name: "Stop now" }).click();
  await expect(page.getByText(/^Stopped by operator at 00:[01]\d:\d\d of 00:30:00$/).first()).toBeVisible({
    timeout: 60_000,
  });

  const again = await setUpAnother(page);
  await expect(again.getByLabel("Duration", { exact: true })).toHaveValue("30");
  await expect(again.getByLabel("Duration unit")).toHaveValue("minutes");
  await expect(again.getByLabel("Seed", { exact: true })).toHaveValue("11");
  await expect(again.getByRole("radio", { name: /^Watch/ })).toBeChecked();
  await expect(again.getByRole("note")).toContainText("anything planned after the stop at 00:");
});

test("another run set up from a run that never finished says its plan is gone and offers fresh settings", async ({
  page,
}) => {
  const created = await action(page, "world.create", {
    domain: "factory-utilities",
    seed: 5,
    adapter: "native-jsonl",
    time_mode: "stepped",
    explore: true,
  });
  const world = String(created.world_id);
  await action(page, "world.destroy", { world_id: world });
  await page.goto(`/#/review?world=${world}`);
  await expect(page.getByRole("status").filter({ hasText: /^Did not finish$/ })).toHaveCount(1, { timeout: 60_000 });

  const again = await setUpAnother(page);
  await expect(again.getByRole("note")).toHaveText(
    "The plan of this run could not be recovered because it did not finish; these are fresh settings. The mode is the one the run opened with: Explore.",
  );
  await expect(again.getByRole("radio", { name: /^Explore/ })).toBeChecked();
  await expect(again.getByLabel("Seed", { exact: true })).toHaveValue("1");
  await expect(again.getByLabel("Duration", { exact: true })).toHaveValue("30");
  await expect(again.getByLabel("Fault 1 instance")).toHaveCount(0);
  await expect(again.getByLabel("Effector call 1 instance")).toHaveCount(0);
});
