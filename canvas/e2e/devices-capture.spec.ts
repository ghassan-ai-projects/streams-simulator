import AxeBuilder from "@axe-core/playwright";
import type { Page } from "@playwright/test";
import { existsSync, readFileSync } from "node:fs";
import {
  announcement,
  cardOf,
  openCapture,
  pumpKind,
  readAnnouncement,
  refusedHeading,
  reviewHeading,
  savedKinds,
  tamperedAnnouncement,
} from "./support/capture";
import { kindPath, outline, verdict } from "./support/devices";
import { expect, test } from "./support/test";

const claim = "A board's announcement is a claim. Nothing is saved until you review and accept it.";

async function violations(page: Page): Promise<unknown[]> {
  return (await new AxeBuilder({ page }).analyze()).violations;
}

async function changeUnitOfFlow(page: Page): Promise<void> {
  await outline(page).getByRole("button", { name: /^flow/ }).click();
  await page.getByLabel("Unit", { exact: true }).fill("m3/s");
  await page.getByRole("button", { name: "Save", exact: true }).click();
  await expect(verdict(page, "Saved and valid.")).toBeVisible();
}

test("an announced device is reviewed and saved, found again as the same, and replaced only after a confirmation", async ({
  page,
  lab,
}) => {
  const file = kindPath(lab, "pump");
  expect(savedKinds(lab)).toEqual([]);
  await openCapture(page);
  await readAnnouncement(page, announcement());
  await expect(reviewHeading(page)).toBeFocused();
  await expect(page.getByLabel("Announcement text")).toHaveCount(0);
  await expect(page.getByText("Simulated", { exact: true })).toBeVisible();
  await expect(page.getByText(claim)).toBeVisible();
  const card = cardOf(page, "pump");
  await expect(card.getByText("New device")).toBeVisible();
  await expect(card.getByText("3 readings · 3 commands")).toBeVisible();
  await expect(
    card.getByText("sha256:b5e93cadac03ad9537622cacf3e529d44de2271811d35455712b7eefb3f24fb1"),
  ).toBeAttached();
  await expect(card.getByText(/Taken from the board's announcement, not verified/)).toBeVisible();
  expect(await violations(page)).toEqual([]);
  expect(existsSync(file)).toBe(false);

  await card.getByRole("button", { name: /^Save as device/ }).click();
  await expect(card.getByText(/Saved as kinds\/pump\.kind\.json/)).toBeVisible();
  await expect(card.getByText("Saved", { exact: true })).toBeVisible();
  expect(JSON.parse(readFileSync(file, "utf8"))).toEqual(pumpKind());
  await card.getByRole("button", { name: /^Open pump in the device editor/ }).click();
  await expect(page.getByRole("region", { name: "Editing pump" })).toBeVisible();

  await page.goBack();
  await expect(reviewHeading(page)).toBeVisible();
  await expect(card.getByText(/Saved as kinds\/pump\.kind\.json/)).toBeVisible();

  await openCapture(page);
  await readAnnouncement(page, announcement());
  await expect(cardOf(page, "pump").getByText(/Same as .*pump\.kind\.json/)).toBeVisible();
  await expect(cardOf(page, "pump").getByRole("button", { name: /^Save as device/ })).toHaveCount(0);
  await expect(cardOf(page, "pump").getByRole("button", { name: /^Skip/ })).toHaveCount(0);
  await cardOf(page, "pump")
    .getByRole("button", { name: /^Open the saved device/ })
    .click();
  await expect(page.getByRole("region", { name: "Editing pump" })).toBeVisible();
  await changeUnitOfFlow(page);

  await openCapture(page);
  await readAnnouncement(page, announcement());
  await expect(card.getByText(/Differs from .*pump\.kind\.json/)).toBeVisible();
  await expect(card.getByText(/The first difference is in /)).toBeVisible();
  const edited = readFileSync(file, "utf8");
  await card.getByRole("button", { name: /^Replace the saved device/ }).click();
  const question = page.getByRole("group", { name: "Replace the saved pump?" });
  await expect(question.getByRole("button", { name: "Keep the saved device" })).toBeFocused();
  expect(await violations(page)).toEqual([]);
  await question.getByRole("button", { name: "Keep the saved device" }).click();
  expect(readFileSync(file, "utf8")).toBe(edited);

  await card.getByRole("button", { name: /^Replace the saved device/ }).click();
  await question.getByRole("button", { name: "Yes, replace it" }).click();
  await expect(card.getByText(/Saved as kinds\/pump\.kind\.json/)).toBeVisible();
  expect(JSON.parse(readFileSync(file, "utf8"))).toEqual(pumpKind());
});

test("an announcement that does not hold is refused in plain words, with the binary's words behind them, and nothing is written", async ({
  page,
  lab,
}) => {
  await openCapture(page);
  await readAnnouncement(page, tamperedAnnouncement());
  await expect(refusedHeading(page)).toBeVisible();
  const sentence = page.getByRole("alert");
  await expect(sentence).toContainText("Copy the announcement again from the board.");
  await expect(sentence).toBeFocused();
  await expect(page.getByText("Nothing was saved.")).toBeVisible();
  await expect(reviewHeading(page)).toHaveCount(0);
  await page.getByText("Details", { exact: true }).click();
  await expect(page.getByText(/kind_failed/)).toBeVisible();
  expect(await violations(page)).toEqual([]);

  await page.getByRole("button", { name: "Retry" }).click();
  await expect(sentence).toBeFocused();
  expect(savedKinds(lab)).toEqual([]);

  await readAnnouncement(page, announcement());
  await expect(reviewHeading(page)).toBeFocused();
  await expect(cardOf(page, "pump").getByText("New device")).toBeVisible();
  expect(savedKinds(lab)).toEqual([]);
});

test("text with no announcement in it is refused as such, and a recorded run without one too", async ({
  page,
  lab,
}) => {
  await openCapture(page);
  await readAnnouncement(page, "this is not an announcement");
  await expect(page.getByRole("heading", { level: 2, name: "No announcement found" })).toBeVisible();
  await page.getByLabel("Recorded run", { exact: true }).selectOption("w-evaluate");
  await page.getByRole("button", { name: "Read this run" }).click();
  await expect(page.getByRole("heading", { level: 2, name: "No announcement found" })).toBeVisible();
  await expect(page.getByText("Nothing was saved.")).toBeVisible();
  expect(savedKinds(lab)).toEqual([]);
});

test("a text over 1 MiB is refused in plain words before it is sent", async ({ page }) => {
  await openCapture(page);
  await page.getByLabel("Announcement text").fill("x".repeat(1024 * 1024 + 1));
  await expect(page.getByRole("alert")).toContainText("larger than 1 MiB");
  await expect(page.getByRole("button", { name: "Read the announcement" })).toHaveAttribute("aria-disabled", "true");
});
