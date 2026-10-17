import { expect, test } from "./support/test";
import { expectVisibleRing } from "./support/keys";

test("the keyboard dialog opens on a heading whose focus is visible, in both themes", async ({ page }) => {
  await page.goto("/#/plant?domain=domains%2Ffactory-utilities.domain.json");
  await expect(page.getByRole("button", { name: "Keyboard" })).toBeVisible();
  for (const theme of ["light", "dark"]) {
    await page.evaluate((value) => {
      document.documentElement.dataset.theme = value;
    }, theme);
    await page.keyboard.press("?");
    await expect(page.getByRole("dialog", { name: "Keyboard" })).toBeVisible();
    await expect(page.getByRole("heading", { level: 2, name: "Keyboard" })).toBeFocused();
    await expectVisibleRing(page);
    await page.keyboard.press("Escape");
    await expect(page.getByRole("dialog", { name: "Keyboard" })).toBeHidden();
  }
});
