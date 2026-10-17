import type { Page } from "@playwright/test";
import { expect, test } from "./support/test";
import { savedPlant, writeProbePlant } from "./support/editorPlant";
import { countPointerEvents, expectVisibleRing, pointerEvents, tabUntil } from "./support/keys";

function regions(page: Page) {
  return {
    build: page.getByRole("complementary", { name: "Build" }),
    outline: page.getByRole("navigation", { name: "Outline" }),
    crumb: page.getByRole("navigation", { name: "Breadcrumb" }).getByRole("button", { name: "Workspace" }),
  };
}

async function addWithKeys(page: Page, search: string, option: string, name: string): Promise<void> {
  const { build } = regions(page);
  const library = build.getByLabel("Equipment", { exact: true });
  await tabUntil(page, build.getByLabel("Find equipment"));
  await expectVisibleRing(page);
  await page.keyboard.type(search);
  await page.keyboard.press("Tab");
  await expect(library).toBeFocused();
  await page.keyboard.press("ArrowDown");
  await expect(library).toHaveValue(option);
  await page.keyboard.press("Enter");
  await expect(library).toBeFocused();
  await page.keyboard.press("Tab");
  await page.keyboard.press("Tab");
  await expect(build.getByLabel("Name", { exact: true })).toBeFocused();
  await page.keyboard.type(name);
  await page.keyboard.press("Enter");
  await expect(page.getByRole("heading", { level: 2, name })).toBeFocused();
}

test("a plant is built and edited with the keyboard alone, from the library search to Save", async ({ page, lab }) => {
  await countPointerEvents(page);
  await page.goto(writeProbePlant(lab, "keys-built", "Keys built"));
  await expect(page.getByRole("heading", { level: 1, name: "Keys built" })).toBeVisible();
  const { outline } = regions(page);
  await expect(outline.getByRole("button")).toHaveCount(1);
  await addWithKeys(page, "cooling", "streamsim/cooling-tower@1", "CT-1");
  await expect(outline.getByRole("button")).toHaveCount(2);
  await tabUntil(page, outline.getByRole("button", { name: "CH-1" }), "Shift+Tab");
  await expectVisibleRing(page);
  await page.keyboard.press("Enter");
  const panel = page.getByRole("complementary", { name: "CH-1" });
  await expect(panel).toBeVisible();
  await tabUntil(page, panel.getByRole("button", { name: "Rename" }));
  await page.keyboard.press("Shift+Tab");
  await expect(panel.getByLabel("Name", { exact: true })).toBeFocused();
  await page.keyboard.press("ControlOrMeta+a");
  await page.keyboard.type("CH-9");
  await page.keyboard.press("Enter");
  const renamed = page.getByRole("complementary", { name: "CH-9" });
  await expect(renamed.getByLabel("Name", { exact: true })).toBeFocused();
  const link = page.getByRole("form", { name: "Link an output" });
  const selects = link.locator("select");
  const [output, target, input] = [selects.nth(0), selects.nth(1), selects.nth(2)];
  await tabUntil(page, output);
  await page.keyboard.type("heat");
  await expect(output).toHaveValue("heat_out");
  await page.keyboard.press("Tab");
  await expect(target).toBeFocused();
  await page.keyboard.type("CT");
  await expect(target).toHaveValue("CT-1");
  await page.keyboard.press("Tab");
  await expect(input).toBeFocused();
  await page.keyboard.type("heat");
  await expect(input).toHaveValue("heat_in");
  await page.keyboard.press("Tab");
  await expect(link.getByRole("button", { name: "Link" })).toBeFocused();
  await page.keyboard.press("Enter");
  await expect(page.getByRole("heading", { name: "Links 1" })).toBeVisible();
  await page.keyboard.press("ControlOrMeta+z");
  await expect(page.getByRole("heading", { name: "Links 0" })).toBeVisible();
  await page.keyboard.press("ControlOrMeta+Shift+z");
  await expect(page.getByRole("heading", { name: "Links 1" })).toBeVisible();
  await page.keyboard.press("ControlOrMeta+s");
  await expect(page.getByRole("status").filter({ hasText: /^Ready to run\. Saved\.$/ })).toBeVisible();
  const saved = savedPlant(lab, "keys-built");
  expect(saved.equipment.flatMap((entry) => entry.names)).toEqual(["CH-9", "CT-1"]);
  expect(saved.connections).toEqual([{ from: "CH-9.heat_out", to: "CT-1.heat_in" }]);
  expect(await pointerEvents(page)).toEqual([]);
});

test("leaving with unsaved changes works with the keyboard alone: focus enters the dialog and returns when it closes", async ({
  page,
  lab,
}) => {
  await countPointerEvents(page);
  await page.goto(writeProbePlant(lab, "keys-leave", "Keys leave"));
  await expect(page.getByRole("heading", { level: 1, name: "Keys leave" })).toBeVisible();
  await addWithKeys(page, "pump", "streamsim/pump@1", "P-1");
  const { crumb } = regions(page);
  const dialog = page.getByRole("dialog", { name: "Leave with unsaved changes?" });
  await tabUntil(page, crumb, "Shift+Tab");
  await page.keyboard.press("Enter");
  await expect(dialog.getByRole("button", { name: "Stay" })).toBeFocused();
  await expectVisibleRing(page);
  await page.keyboard.press("Escape");
  await expect(dialog).toBeHidden();
  await expect(crumb).toBeFocused();
  await page.keyboard.press("Enter");
  await expect(dialog.getByRole("button", { name: "Stay" })).toBeFocused();
  await page.keyboard.press("Enter");
  await expect(dialog).toBeHidden();
  await expect(crumb).toBeFocused();
  await page.keyboard.press("Enter");
  await page.keyboard.press("Tab");
  await expect(dialog.getByRole("button", { name: "Discard and leave" })).toBeFocused();
  await expectVisibleRing(page);
  await page.keyboard.press("Tab");
  await expect(dialog.getByRole("button", { name: "Save and leave" })).toBeFocused();
  await page.keyboard.press("Shift+Tab");
  await page.keyboard.press("Enter");
  await expect(page.getByRole("heading", { level: 1, name: "Workspace" })).toBeVisible();
  expect(savedPlant(lab, "keys-leave").equipment.flatMap((entry) => entry.names)).toEqual(["CH-1"]);
  expect(await pointerEvents(page)).toEqual([]);
});
