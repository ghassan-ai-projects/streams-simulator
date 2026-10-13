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
  await expect(page.getByText("Verified", { exact: true })).toBeVisible({ timeout: 60_000 });
  await expect(page.getByText("Evaluate · ended")).toBeVisible();
  const report = page.getByRole("complementary", { name: "Run details" });
  await expect(report.getByRole("region", { name: "File check" })).toContainText("files match");
  await expect(report.getByRole("region", { name: "Replay check" })).toContainText("reproduced its trace");
  await expect(page.getByRole("status").filter({ hasText: "plant file changed after this run" })).toBeVisible();
  await expect(report.getByRole("region", { name: "KPI truth" })).toContainText("CH-1");
  await expect(page.getByRole("status").filter({ hasText: /^[\d,]+ observations$/ })).toBeVisible({ timeout: 60_000 });
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
  await expect(page.getByText("Not verified", { exact: true })).toBeVisible({ timeout: 60_000 });
  await expect(
    page.getByRole("complementary", { name: "Run details" }).getByRole("region", { name: "File check" }),
  ).toContainText("files differ");
  await expect(page.getByRole("alert").filter({ hasText: "differ from what the run wrote" })).toBeVisible();
  const graph = await page.getByRole("application", { name: "Plant graph" }).boundingBox();
  expect(graph?.height ?? 0).toBeGreaterThan(300);
});
