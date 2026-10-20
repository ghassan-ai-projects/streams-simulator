import AxeBuilder from "@axe-core/playwright";
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

test.setTimeout(180_000);

let bench: Bench | null = null;

test.afterEach(() => {
  bench?.close();
  bench = null;
});

const cooling = { level_percent: 30 };

async function recordSessionWithCommands(page: Page, lab: { readonly binary: string }): Promise<string> {
  const open = openScriptedBench(lab.binary, [
    { id: "twin-a", afterS: 20, device: "fan-01", command: "set_cooling", args: cooling },
    { id: "twin-b", afterS: 30, device: "fan-01", command: "set_cooling", args: cooling },
  ]);
  bench = open;
  const panel = await openSessionPanel(page);
  await planOneRealBoard(panel, open, "fan-01", "fan-01");
  await panel.getByRole("button", { name: "Start the session" }).click();
  await expect(panel.getByText(/^Connected session · run c-[0-9a-f]{8}$/)).toBeVisible();
  startGateway(open, "fan-01", "fan");
  await expect(panel.getByText(/announced every real board/)).toBeVisible({ timeout: 30_000 });
  await page.getByRole("button", { name: "Open live view", exact: true }).click();
  const last = page.getByRole("table", { name: "Command evidence" }).getByRole("row", { name: /twin-b/ });
  await expect(last).toBeVisible({ timeout: 120_000 });
  await expect(last).not.toContainText("Unconfirmed");
  await page.getByRole("button", { name: "Open plant", exact: true }).click();
  await page.getByRole("button", { name: "Run plant", exact: true }).click();
  await panel.getByRole("button", { name: "Stop and save" }).click();
  const saved = panel.getByText(/^The session was stopped and saved as run c-[0-9a-f]{8}\.$/);
  await expect(saved).toBeVisible({ timeout: 30_000 });
  const run = runIdOf(await saved.textContent());
  await panel.getByRole("button", { name: "Open the review" }).click();
  await expect(page.getByRole("heading", { level: 1, name: `Run ${run}` })).toBeVisible();
  return run;
}

async function compareWithTwin(twin: Locator): Promise<Locator> {
  await expect(twin.getByRole("group", { name: "Twin comparison result" })).toBeHidden();
  await twin.getByRole("button", { name: "Compare with twin" }).click();
  const result = twin.getByRole("group", { name: "Twin comparison result" });
  await expect(result).toBeVisible();
  return result;
}

async function nothingClippedAt1280(page: Page): Promise<void> {
  await page.setViewportSize({ width: 1280, height: 800 });
  const details = page.getByRole("complementary", { name: "Run details" });
  await expect.poll(() => details.evaluate((aside) => aside.scrollWidth <= aside.clientWidth)).toBe(true);
}

test("the twin versus board report compares a saved session with the twin on user action, and says honestly what it could not judge", async ({
  page,
  lab,
}) => {
  addThermalBench(lab);
  await recordSessionWithCommands(page, lab);
  const twin = page.getByRole("region", { name: "Twin vs board", exact: true });
  await expect(twin.getByLabel("Tolerances (optional)", { exact: true })).toBeVisible();

  await twin.getByLabel("Tolerances (optional)", { exact: true }).fill("nobody.metric=1");
  await twin.getByRole("button", { name: "Compare with twin" }).click();
  await expect(twin.getByRole("alert")).toContainText("The comparison could not be made:");
  await expect(twin.getByRole("alert")).toContainText("tolerance");
  await twin.getByLabel("Tolerances (optional)", { exact: true }).fill("");

  const result = await compareWithTwin(twin);
  await expect(result.getByText("Cannot judge", { exact: true })).toBeVisible();
  await expect(result.getByText("No tolerance was typed.", { exact: false })).toBeVisible();
  const commands = page.getByRole("list", { name: "Commands and the twin" });
  for (const id of ["twin-a", "twin-b"]) {
    const item = commands.getByRole("listitem").filter({ hasText: `${id} · set_cooling on fan-01` });
    await expect(item).toContainText("Not applied");
  }
  await expect(page.getByRole("region", { name: "Where the twin and the board part ways", exact: true })).toContainText(
    "Nothing to show: no reading could be compared with the twin.",
  );
  expect((await new AxeBuilder({ page }).include("aside").analyze()).violations).toEqual([]);
  await nothingClippedAt1280(page);

  await page
    .getByRole("navigation", { name: "Outline" })
    .getByRole("button", { name: /^fan-01/ })
    .click();
  const overlay = page.getByRole("region", { name: "Twin prediction · fan-01", exact: true });
  await expect(overlay).toBeVisible();
  await expect(
    overlay.getByText("The board reported no reading of level_percent to compare.", { exact: true }),
  ).toBeVisible();
  expect((await new AxeBuilder({ page }).include("aside").analyze()).violations).toEqual([]);
});
