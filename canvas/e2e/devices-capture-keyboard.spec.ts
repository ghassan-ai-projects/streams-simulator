import { readFileSync } from "node:fs";
import { announcement, captureButton, cardOf, pumpKind, reviewHeading } from "./support/capture";
import type { Kind } from "./support/capture";
import { kindPath, writeKind } from "./support/devices";
import { countPointerEvents, expectVisibleRing, pointerEvents, tabUntil } from "./support/keys";
import { expect, test } from "./support/test";

test("an announcement from a real board is read, reviewed and saved as a device with the keyboard alone", async ({
  page,
}) => {
  await countPointerEvents(page);
  await page.goto("/");
  await tabUntil(page, captureButton(page));
  await expectVisibleRing(page);
  await page.keyboard.press("Enter");
  await expect(page.getByRole("heading", { level: 1, name: "Capture from a board" })).toBeFocused();

  await tabUntil(page, page.getByLabel("Announcement text"));
  await page.keyboard.insertText(announcement("hardware"));
  await tabUntil(page, page.getByRole("button", { name: "Read the announcement" }));
  await expectVisibleRing(page);
  await page.keyboard.press("Enter");
  await expect(reviewHeading(page)).toBeFocused();
  await expect(page.getByText("Real board", { exact: true })).toBeVisible();
  await expect(page.getByText("The board announced 1 device.")).toBeAttached();

  const card = cardOf(page, "pump");
  await expect(card.getByText("New device")).toBeVisible();
  await tabUntil(page, card.getByRole("button", { name: /^Save as device/ }));
  await expectVisibleRing(page);
  await page.keyboard.press("Enter");
  const open = card.getByRole("button", { name: /^Open pump in the device editor/ });
  await expect(open).toBeFocused();
  await page.keyboard.press("Enter");
  await expect(page.getByRole("region", { name: "Editing pump" })).toBeVisible();
  expect(await pointerEvents(page)).toEqual([]);
});

test("Escape leaves the capture screen and the focus returns to the button that opened it", async ({ page }) => {
  await page.goto("/");
  await captureButton(page).focus();
  await page.keyboard.press("Enter");
  await expect(page.getByRole("heading", { level: 1, name: "Capture from a board" })).toBeFocused();
  await page.keyboard.press("Escape");
  await expect(captureButton(page)).toBeFocused();
});

test("a device that differs is kept, then replaced, with the keyboard alone", async ({ page, lab }) => {
  const [flow, ...rest] = pumpKind().metrics as Kind[];
  writeKind(lab, "pump", { ...pumpKind(), metrics: [{ ...flow, unit: "m3/s" }, ...rest] });
  const file = kindPath(lab, "pump");
  const edited = readFileSync(file, "utf8");
  await countPointerEvents(page);
  await page.goto("/");
  await tabUntil(page, captureButton(page));
  await page.keyboard.press("Enter");
  await tabUntil(page, page.getByLabel("Announcement text"));
  await page.keyboard.insertText(announcement());
  await tabUntil(page, page.getByRole("button", { name: "Read the announcement" }));
  await page.keyboard.press("Enter");
  await expect(reviewHeading(page)).toBeFocused();

  const card = cardOf(page, "pump");
  const replace = card.getByRole("button", { name: /^Replace the saved device/ });
  const question = page.getByRole("group", { name: "Replace the saved pump?" });
  await tabUntil(page, replace);
  await page.keyboard.press("Enter");
  await expect(question.getByRole("button", { name: "Keep the saved device" })).toBeFocused();
  await page.keyboard.press("Enter");
  await expect(replace).toBeFocused();
  expect(readFileSync(file, "utf8")).toBe(edited);

  await page.keyboard.press("Enter");
  await expect(question.getByRole("button", { name: "Keep the saved device" })).toBeFocused();
  await page.keyboard.press("Shift+Tab");
  await expect(question.getByRole("button", { name: "Yes, replace it" })).toBeFocused();
  await page.keyboard.press("Enter");
  await expect(card.getByText(/Saved as kinds\/pump\.kind\.json/)).toBeVisible();
  expect(JSON.parse(readFileSync(file, "utf8"))).toEqual(pumpKind());
  expect(await pointerEvents(page)).toEqual([]);
});
