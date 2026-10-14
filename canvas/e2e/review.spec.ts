import AxeBuilder from "@axe-core/playwright";
import { expect, test } from "@playwright/test";
import { env } from "./support/env";

test.use({ storageState: env("CANVAS_APP_STATE") });
test.setTimeout(90_000);

test("a recorded run plays back with the binary's report and a Verified badge", async ({ page }) => {
  await page.goto(env("CANVAS_APP_URL"));
  await page.getByRole("button", { name: "Open review of w-evaluate" }).click();
  const header = page.getByRole("heading", { name: "w-evaluate" });
  await expect(header).toBeVisible();
  await expect(page.getByRole("status").filter({ hasText: /^Verified$/ })).toHaveCount(1, { timeout: 60_000 });
  await expect(page.getByText("Evaluate · ended")).toBeVisible();
  const report = page.getByRole("complementary", { name: "Run details" });
  await expect(report.getByRole("region", { name: "File check" })).toContainText("files match");
  await expect(report.getByRole("region", { name: "Replay check" })).toContainText("Reproduces");
  await expect(page.getByRole("status").filter({ hasText: "plant file changed after this run" })).toBeVisible();
  await expect(report.getByRole("region", { name: "KPI truth" })).toContainText("CH-1");
  await expect(page.getByText(/^[\d,]+ observations$/)).toBeVisible({ timeout: 60_000 });
  const chiller = page.getByRole("group", { name: /^CH-1, chiller,/ });
  await expect(chiller.locator("dl")).toContainText("power");
  const slider = page.getByRole("slider", { name: "Playback time" });
  await expect(slider).toHaveAttribute("aria-valuetext", "00:30:00");
  await slider.fill("60");
  await expect(slider).toHaveAttribute("aria-valuetext", "00:01:00");
  expect((await new AxeBuilder({ page }).analyze()).violations).toEqual([]);
});

test("a run whose files were edited after it ended is Not verified", async ({ page }) => {
  await page.goto(`${env("CANVAS_APP_URL")}/#/run?world=w-edited`);
  await expect(page.getByRole("status").filter({ hasText: /^Not verified$/ })).toHaveCount(1, { timeout: 60_000 });
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
  const replay = page.getByRole("complementary", { name: "Run details" }).getByRole("region", { name: "Replay check" });
  await expect(replay).toContainText(/Does not (reproduce|verify)/);
  await replay.getByText("What differs").click();
  await expect(replay.locator("pre")).not.toBeEmpty();
  await expect(page.getByRole("note")).toContainText("This run did not verify");
});

test("a run whose recorded plant was tampered with reads Not verified, not Not checked", async ({ page }) => {
  await page.goto(`${env("CANVAS_APP_URL")}/#/run?world=w-tampered`);
  await expect(page.getByRole("status").filter({ hasText: /^Not verified$/ })).toHaveCount(1, { timeout: 60_000 });
  const replay = page.getByRole("complementary", { name: "Run details" }).getByRole("region", { name: "Replay check" });
  await replay.getByText("What differs").click();
  await expect(replay).toContainText("nothing was replayed");
  await expect(replay.locator("pre")).toContainText("does not match its digest");
});

test("a hall carries its faulted members, and inside it each names its fault", async ({ page }) => {
  await page.goto(`${env("CANVAS_APP_URL")}/#/run?world=w-hall`);
  await expect(page.getByRole("status").filter({ hasText: /^Verified$/ })).toHaveCount(1, { timeout: 60_000 });
  const hall = page.getByRole("group", { name: /^h02, Fault in 2, / });
  await expect(hall).toContainText("Fault 2");
  await expect(page.getByRole("group", { name: /^h01, / })).not.toContainText("Fault");
  await page.getByRole("navigation", { name: "Outline" }).getByRole("button", { name: "h02", exact: true }).click();
  await expect(page.getByRole("heading", { level: 1, name: "h02" })).toBeVisible();
  await expect(page.getByRole("group", { name: /^M-1, Fault: product jam, / })).toContainText("product jam");
  await page.getByRole("button", { name: "w-hall" }).click();
  await expect(page.getByRole("heading", { level: 1, name: "w-hall" })).toBeVisible();
});
