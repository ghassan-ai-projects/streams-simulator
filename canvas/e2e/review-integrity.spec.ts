import type { Page } from "@playwright/test";
import { expect, test } from "./support/test";
import { tokenColor } from "./support/tokenColor";

test.setTimeout(90_000);

async function chipLook(page: Page, label: RegExp) {
  const chip = page.locator("[data-part='status-chip']").filter({ hasText: label });
  await expect(chip).toHaveCount(1, { timeout: 60_000 });
  return chip.evaluate((node) => {
    const style = getComputedStyle(node);
    return { color: style.color, background: style.backgroundColor };
  });
}

async function openIntegrity(page: Page) {
  await page.getByRole("complementary", { name: "Run details" }).locator("summary").first().click();
}

test("an incomplete stamp never reads Reproduces when the completed journals contradict it", async ({ page }) => {
  await page.goto("/#/review?world=w-incomplete-stamp");
  await expect(page.getByRole("status").filter({ hasText: /^Did not finish$/ })).toHaveCount(1, { timeout: 60_000 });
  const details = page.getByRole("complementary", { name: "Run details" });
  await openIntegrity(page);
  const replay = details.getByRole("region", { name: "Replay check" });
  await expect(replay).toContainText("run.json says the run ended incomplete");
  await expect(replay).toContainText("contradicts");
  await expect(replay).not.toContainText("Reproduces");
  await expect(details.getByRole("note")).toContainText("This run did not verify");
  await expect(page.getByRole("status").filter({ hasText: /^Verified$/ })).toHaveCount(0);
  await replay.getByText("For developers").click();
  await expect(replay.locator("pre")).toContainText("incomplete");
  await page
    .getByRole("navigation", { name: "Outline" })
    .getByRole("button", { name: /^CH-1 / })
    .click();
  await expect(details.getByRole("region", { name: "True results" })).toContainText("CH-1");
  await expect(details.getByRole("note")).toContainText("This run did not verify");
});

test("a complete run missing a recorded digest fails both file and replay checks", async ({ page }) => {
  await page.goto("/#/review?world=w-missing-digest");
  await expect(page.getByRole("status").filter({ hasText: /^Not verified$/ })).toHaveCount(1, { timeout: 60_000 });
  const details = page.getByRole("complementary", { name: "Run details" });
  await openIntegrity(page);
  await expect(details.getByRole("region", { name: "File check" })).toContainText("The files changed after the run.");
  await expect(details.getByRole("region", { name: "Replay check" })).toContainText(
    "records no digest for ledger.jsonl",
  );
  await expect(details.getByRole("region", { name: "Replay check" })).not.toContainText("Reproduces");
});

test("the Not verified chip reads as an alarm and Did not finish as a warning, neither as the neutral chip @smoke", async ({
  page,
}) => {
  await page.goto("/#/review?world=w-tampered");
  const alarm = await chipLook(page, /^Not verified$/);
  await page.goto("/#/review?world=w-incomplete-stamp");
  const warning = await chipLook(page, /^Did not finish$/);
  const neutral = {
    color: await tokenColor(page, "--text-secondary", "color"),
    background: await tokenColor(page, "--surface-hover", "backgroundColor"),
  };
  expect(alarm.color).toBe(await tokenColor(page, "--status-alarm", "color"));
  expect(warning.color).toBe(await tokenColor(page, "--status-warn", "color"));
  expect(alarm).not.toEqual(neutral);
  expect(warning).not.toEqual(neutral);
  expect(alarm).not.toEqual(warning);
  expect(alarm.background).not.toBe(warning.background);
});
