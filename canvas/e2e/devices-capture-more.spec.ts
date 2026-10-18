import { readFileSync, writeFileSync } from "node:fs";
import {
  announcement,
  birthOf,
  cardOf,
  digestRuleHolds,
  fanKind,
  invertedKind,
  openCapture,
  pumpKind,
  readAnnouncement,
  reviewHeading,
  savedKinds,
} from "./support/capture";
import type { Kind } from "./support/capture";
import { kindPath, writeKind } from "./support/devices";
import { expect, test } from "./support/test";

function editedPump(): Kind {
  const pump = pumpKind();
  const [flow, ...rest] = pump.metrics as Kind[];
  return { ...pump, metrics: [{ ...flow, unit: "m3/s" }, ...rest] };
}

test("a saved device that changed after the comparison is not replaced, and the file is left as it is", async ({
  page,
  lab,
}) => {
  const file = kindPath(lab, "pump");
  writeKind(lab, "pump", editedPump());
  await openCapture(page);
  await readAnnouncement(page, announcement());
  const card = cardOf(page, "pump");
  await expect(card.getByText(/Differs from .*pump\.kind\.json/)).toBeVisible();
  await expect(card.getByText(/Saved: "m3\/s"\. The board says: "m3\/h"\./)).toBeVisible();
  const moved = JSON.stringify({ ...editedPump(), description: "changed on disk after the comparison" });
  writeFileSync(file, moved);
  await card.getByRole("button", { name: /^Replace the saved device/ }).click();
  await page.getByRole("button", { name: "Yes, replace it" }).click();
  await expect(card.getByRole("alert")).toContainText("changed since it was compared");
  expect(readFileSync(file, "utf8")).toBe(moved);
  await card.getByRole("button", { name: /^Replace the saved device/ }).click();
  await page.getByRole("button", { name: "Yes, replace it" }).click();
  await expect(card.getByRole("alert")).toContainText("changed since it was compared");
  expect(readFileSync(file, "utf8")).toBe(moved);
});

test("a board with two devices: each is saved on its own, a skip moves the focus on and can be undone", async ({
  page,
  lab,
}) => {
  expect(digestRuleHolds()).toBe(true);
  await openCapture(page);
  await readAnnouncement(page, birthOf([pumpKind(), fanKind()]));
  await expect(reviewHeading(page)).toBeFocused();
  await expect(page.getByRole("article")).toHaveCount(2);

  await cardOf(page, "pump").getByRole("button", { name: /^Skip/ }).click();
  await expect(cardOf(page, "fan").getByRole("button", { name: /^Save as device/ })).toBeFocused();
  await cardOf(page, "pump")
    .getByRole("button", { name: /^Undo skipping/ })
    .click();
  await expect(cardOf(page, "pump").getByRole("button", { name: /^Save as device/ })).toBeFocused();

  await cardOf(page, "fan")
    .getByRole("button", { name: /^Save as device/ })
    .click();
  await expect(cardOf(page, "fan").getByText(/Saved as kinds\/fan\.kind\.json/)).toBeVisible();
  expect(savedKinds(lab)).toEqual(["fan.kind.json"]);
  await expect(cardOf(page, "pump").getByRole("button", { name: /^Save as device/ })).toBeVisible();

  await cardOf(page, "pump").getByRole("button", { name: /^Skip/ }).click();
  await expect(reviewHeading(page)).toBeFocused();
  expect(JSON.parse(readFileSync(kindPath(lab, "fan"), "utf8"))).toEqual(fanKind());
  expect(savedKinds(lab)).toEqual(["fan.kind.json"]);
});

test("a log with several announcements takes the first and says what follows it", async ({ page }) => {
  const log = [birthOf([pumpKind()]), birthOf([pumpKind()], { bd_seq: 1 }), birthOf([fanKind()], { bd_seq: 2 })];
  await openCapture(page);
  await readAnnouncement(page, log.join("\n"));
  await expect(reviewHeading(page)).toBeFocused();
  const notice = page.getByRole("region", { name: "About this announcement" });
  await expect(notice).toContainText("This log holds 3 announcements; showing the first (line 1). Later ones differ.");
  await expect(page.getByRole("article")).toHaveCount(1);
  await expect(cardOf(page, "pump")).toBeVisible();
});

test("a draft that still breaks a rule is saved as announced, and the editor shows the issues", async ({
  page,
  lab,
}) => {
  await openCapture(page);
  await readAnnouncement(page, birthOf([invertedKind()]));
  const card = cardOf(page, "inverted");
  await expect(card.getByText("Not yet valid")).toBeVisible();
  await card.getByRole("button", { name: /^Save as device/ }).click();
  await expect(card.getByText(/Saved as kinds\/inverted\.kind\.json/)).toBeVisible();
  expect(JSON.parse(readFileSync(kindPath(lab, "inverted"), "utf8"))).toEqual(invertedKind());
  await card.getByRole("button", { name: /^Open inverted in the device editor/ }).click();
  await expect(page.getByRole("region", { name: "Editing inverted" })).toBeVisible();
  await expect(page.getByRole("region", { name: "Issues" })).toBeVisible();
});
