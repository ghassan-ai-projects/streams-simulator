import AxeBuilder from "@axe-core/playwright";
import { expect, test } from "./support/test";
test.use({ target: "unpaired" });
test.describe.configure({ mode: "serial" });

test("a wrong code is refused, the printed code pairs and opens the workspace", async ({ page, lab }) => {
  await page.goto("/");
  await expect(page.getByRole("heading", { name: "Pair this browser" })).toBeVisible();
  expect((await new AxeBuilder({ page }).analyze()).violations).toEqual([]);
  const field = page.getByLabel("Pairing code");
  await field.fill("AAAA-BBBB-CCCC");
  await page.getByRole("button", { name: "Pair" }).click();
  await expect(page.getByRole("alert")).toContainText("That code did not match");
  await field.fill(lab.code);
  await page.getByRole("button", { name: "Pair" }).click();
  await expect(page.getByRole("heading", { name: "lab-workspace" })).toBeVisible();
});
