import type { Page } from "@playwright/test";
import { expect, test } from "@playwright/test";
import { env } from "./support/env";

test.use({ storageState: env("CANVAS_APP_STATE") });
test.setTimeout(90_000);

async function openIntegrity(page: Page) {
  await page.getByRole("complementary", { name: "Run details" }).locator("summary").first().click();
}

test("an incomplete stamp never reads Reproduces when the completed journals contradict it", async ({ page }) => {
  await page.goto(`${env("CANVAS_APP_URL")}/#/run?world=w-incomplete-stamp`);
  await expect(page.getByRole("status").filter({ hasText: /^Incomplete$/ })).toHaveCount(1, { timeout: 60_000 });
  const details = page.getByRole("complementary", { name: "Run details" });
  await openIntegrity(page);
  const replay = details.getByRole("region", { name: "Replay check" });
  await expect(replay).toContainText("run.json says the run ended incomplete");
  await expect(replay).toContainText("contradicts");
  await expect(replay).not.toContainText("Reproduces");
  await expect(details.getByRole("note")).toContainText("This run did not verify");
  await expect(page.getByRole("status").filter({ hasText: /^Verified$/ })).toHaveCount(0);
  await replay.getByText("What differs").click();
  await expect(replay.locator("pre")).toContainText("incomplete");
  await page.getByRole("navigation", { name: "Outline" }).getByRole("button", { name: "CH-1", exact: true }).click();
  await expect(details.getByRole("region", { name: "KPI truth" })).toContainText("CH-1");
  await expect(details.getByRole("note")).toContainText("This run did not verify");
});

test("a complete run missing a recorded digest fails both file and replay checks", async ({ page }) => {
  await page.goto(`${env("CANVAS_APP_URL")}/#/run?world=w-missing-digest`);
  await expect(page.getByRole("status").filter({ hasText: /^Not verified$/ })).toHaveCount(1, { timeout: 60_000 });
  const details = page.getByRole("complementary", { name: "Run details" });
  await openIntegrity(page);
  await expect(details.getByRole("region", { name: "File check" })).toContainText("files differ");
  await expect(details.getByRole("region", { name: "Replay check" })).toContainText(
    "records no digest for ledger.jsonl",
  );
  await expect(details.getByRole("region", { name: "Replay check" })).not.toContainText("Reproduces");
});
