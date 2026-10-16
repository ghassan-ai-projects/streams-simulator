import { expect, test } from "@playwright/test";
import { env } from "./support/env";

test.use({ storageState: env("CANVAS_APP_STATE") });

test("a native inspector disclosure updates its bottom cue without scrolling or resizing", async ({ page }) => {
  await page.setViewportSize({ width: 1440, height: 1080 });
  await page.goto(`${env("CANVAS_APP_URL")}/#/plant?domain=domains%2Ffactory-utilities.domain.json`);
  await page
    .getByRole("navigation", { name: "Outline" })
    .getByRole("button", { name: /^CH-1 chiller$/ })
    .click();
  const inspector = page.getByRole("complementary", { name: "Inspector", exact: true });
  await expect(inspector).not.toHaveAttribute("data-cut", "true");
  const faults = inspector.locator("summary").filter({ hasText: /^Faults/ });
  await faults.click();
  await expect.poll(() => inspector.evaluate((node) => node.scrollHeight - node.clientHeight)).toBeGreaterThan(1);
  await expect(inspector).toHaveAttribute("data-cut", "true");
  await expect(inspector).toHaveCSS("mask-image", /linear-gradient/);
  await faults.click();
  await expect.poll(() => inspector.evaluate((node) => node.scrollHeight - node.clientHeight)).toBeLessThanOrEqual(1);
  await expect(inspector).not.toHaveAttribute("data-cut", "true");
});
