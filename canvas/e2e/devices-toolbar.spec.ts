import type { Locator, Page } from "@playwright/test";
import { createDevice } from "./support/devices";
import { expect, test } from "./support/test";

function toolbar(page: Page, name: string): Locator {
  return page
    .getByRole("region", { name: `Editing ${name}` })
    .locator(":scope > *")
    .first();
}

async function expectReachable(bar: Locator): Promise<void> {
  const buttons = bar.getByRole("button");
  const count = await buttons.count();
  expect(count).toBeGreaterThan(0);
  for (let index = 0; index < count; index += 1) {
    const placed = await buttons.nth(index).evaluate((element) => {
      const container = element.closest("section")?.firstElementChild;
      if (container === null || container === undefined) return { inside: false, top: false };
      const box = container.getBoundingClientRect();
      const rect = element.getBoundingClientRect();
      const centre = document.elementFromPoint(rect.x + rect.width / 2, rect.y + rect.height / 2);
      const inside = rect.left >= box.left - 0.5 && rect.right <= box.right + 0.5 && rect.width > 0;
      return { inside, top: centre !== null && element.contains(centre) };
    });
    expect(placed, `button ${String(index)} is inside the toolbar and not covered`).toEqual({
      inside: true,
      top: true,
    });
  }
}

for (const width of [1280, 1440]) {
  test(`the toolbar of a device keeps every button inside its column and reachable at ${String(width)} px, also while it asks before discarding`, async ({
    page,
  }) => {
    await page.setViewportSize({ width, height: 800 });
    await createDevice(page, "e2e-bar");
    const bar = toolbar(page, "e2e-bar");
    await expectReachable(bar);
    await page.getByRole("button", { name: "Add a reading" }).click();
    await page.getByLabel("Name", { exact: true }).fill("a_long_reading_name");
    await expectReachable(bar);
    await page.getByRole("button", { name: "Revert" }).click();
    await expect(page.getByRole("group", { name: "Discard your changes?" })).toBeVisible();
    await expectReachable(bar);
    await page.getByRole("button", { name: "Keep editing" }).click();
    await expectReachable(bar);
  });
}
