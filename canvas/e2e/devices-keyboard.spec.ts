import { step, verdict } from "./support/devices";
import { countPointerEvents, expectVisibleRing, pointerEvents, tabUntil } from "./support/keys";
import { expect, test } from "./support/test";

test("a device is created, defined, given a behaviour and made into a twin with the keyboard alone", async ({
  page,
}) => {
  await countPointerEvents(page);
  await page.goto("/");
  const create = page.getByRole("region", { name: "Devices" }).getByRole("button", { name: "Create device" });
  await tabUntil(page, create);
  await expectVisibleRing(page);
  await page.keyboard.press("Enter");
  await expect(page.getByRole("form", { name: "New device" }).getByLabel("Name")).toBeFocused();
  await page.keyboard.type("e2e_kbd");
  await page.keyboard.press("Enter");
  await expect(page.getByRole("region", { name: "Editing e2e_kbd" })).toBeVisible();
  await expect(page.getByRole("heading", { level: 1, name: "e2e_kbd" })).toBeFocused();

  await tabUntil(page, page.getByRole("button", { name: "Add a reading" }));
  await page.keyboard.press("Enter");
  await expect(page.getByLabel("Name", { exact: true })).toBeFocused();
  await page.keyboard.type("level");
  await tabUntil(page, page.getByLabel("Unit", { exact: true }));
  await page.keyboard.type("m");
  await tabUntil(page, page.getByLabel("Minimum"));
  await page.keyboard.type("0");
  await page.keyboard.press("Tab");
  await expect(page.getByLabel("Maximum")).toBeFocused();
  await page.keyboard.type("10");
  await page.keyboard.press("Tab");
  await expect(page.getByLabel("Resolution", { exact: true })).toBeFocused();
  await page.keyboard.type("1");
  await page.keyboard.press("ControlOrMeta+s");
  await expect(verdict(page, "Saved and valid.")).toBeVisible();
  await expect(page.getByText("Ready for a twin")).toBeVisible();

  await tabUntil(page, step(page, /^Behaviour/), "Shift+Tab");
  await page.keyboard.press("Enter");
  await expect(page.getByRole("heading", { level: 2, name: "Twin", exact: true })).toBeFocused();
  await tabUntil(page, page.getByRole("button", { name: "Create twin" }), "Shift+Tab");
  await expectVisibleRing(page);
  await page.keyboard.press("Enter");
  await expect(step(page, /^Behaviour/)).toContainText("Created");
  await expect(page.getByRole("button", { name: "Replace twin" })).toBeVisible();
  expect(await pointerEvents(page)).toEqual([]);
});
