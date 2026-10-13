import AxeBuilder from "@axe-core/playwright";
import { expect, test } from "@playwright/test";

test("the shell renders with no accessibility violation in both themes", async ({ page }) => {
  await page.goto("/");
  await expect(page.getByRole("heading", { name: "No workspace connected" })).toBeVisible();
  for (const label of ["Theme: follow the system", "Theme: light"]) {
    await page.getByRole("button", { name: label }).click();
    const results = await new AxeBuilder({ page }).analyze();
    expect(results.violations).toEqual([]);
  }
  await expect(page.locator("html")).toHaveAttribute("data-theme", "dark");
});
