import AxeBuilder from "@axe-core/playwright";
import { expect, test } from "./support/test";

test("selecting a port type in the legend hides its wires and selecting it again brings them back", async ({
  page,
}) => {
  await page.goto("/");
  await page.getByRole("button", { name: "Open Factory utilities" }).click();
  const wires = page.locator(".react-flow__edge");
  const power = page.locator(".react-flow__edge.family-power");
  await expect(wires).toHaveCount(59);
  await expect(power).toHaveCount(18);
  const legend = page.getByRole("button", { name: "Legend" });
  await legend.click();
  const key = page.getByRole("group", { name: "Wire colour and pin shape by port type" });
  await key.getByRole("button", { name: "Power" }).click();
  await expect(power).toHaveCount(0);
  await expect(wires).toHaveCount(41);
  await expect(key.getByRole("button", { name: "Power hidden" })).toHaveAttribute("aria-pressed", "false");
  expect((await new AxeBuilder({ page }).analyze()).violations).toEqual([]);
  await page.keyboard.press("Escape");
  await expect(page.getByRole("button", { name: "Legend · 1 hidden" })).toBeFocused();
  await page.getByRole("button", { name: "Legend · 1 hidden" }).click();
  await key.getByRole("button", { name: "Power hidden" }).click();
  await expect(power).toHaveCount(18);
  await expect(legend).toBeVisible();
});
