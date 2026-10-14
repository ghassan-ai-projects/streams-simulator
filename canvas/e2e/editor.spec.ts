import AxeBuilder from "@axe-core/playwright";
import type { Page } from "@playwright/test";
import { expect, test } from "@playwright/test";
import { env } from "./support/env";

test.use({ storageState: env("CANVAS_APP_STATE") });
test.setTimeout(60_000);

async function add(page: Page, component: string, name: string) {
  const build = page.getByRole("complementary", { name: "Build" });
  await build.getByLabel("Component").selectOption({ label: component });
  await expect(build.getByLabel("Preset").getByRole("option")).not.toHaveCount(1);
  await build.getByLabel("Name").fill(name);
  await build.getByLabel("Area").selectOption("Chiller plant");
  await build.getByRole("button", { name: "Add to the plant" }).click();
  await expect(page.getByRole("complementary", { name })).toBeVisible();
  await page.getByRole("button", { name: "Close the instance" }).click();
}

function pin(page: Page, node: string, port: string) {
  return page.locator(`.react-flow__node[data-id="${node}"] .react-flow__handle[data-handleid="${port}"]`);
}

test("a new plant is built in Chrome, refused with the issue on its element, fixed, composed and opened", async ({
  page,
}) => {
  await page.goto(env("CANVAS_APP_URL"));
  await page.getByRole("button", { name: "New plant" }).click();
  const form = page.getByRole("form", { name: "New plant" });
  await form.getByLabel("Id").fill("lab-cooling");
  await form.getByLabel("Title").fill("Lab cooling");
  await form.getByLabel("Based on").selectOption({ label: "Factory utilities" });
  await form.getByRole("button", { name: "Create and edit" }).click();
  await expect(page.getByRole("heading", { level: 1, name: "Lab cooling" })).toBeVisible();
  await expect(page.getByText("No equipment yet.")).toBeVisible();
  await add(page, "Water-cooled chiller", "CH-1");
  await add(page, "Cooling tower", "CT 1");
  await pin(page, "CH-1", "heat_out").dragTo(pin(page, "CT 1", "heat_in"));
  await page.getByRole("navigation", { name: "Outline" }).getByRole("button", { name: "CT 1" }).click();
  await expect(page.getByRole("heading", { name: "Links (1)" })).toBeVisible();
  expect((await new AxeBuilder({ page }).analyze()).violations).toEqual([]);
  await page.getByRole("button", { name: "Save" }).click();
  await expect(page.getByRole("status").filter({ hasText: /^Saved, not composed/ })).toBeVisible();
  await expect(page.getByRole("group", { name: /^CT 1, .*has issues$/ })).toBeVisible();
  await expect(page.getByRole("list", { name: "Issues" })).toContainText("does not match pattern");
  await page.getByLabel("Name").fill("CT-1");
  await page.getByRole("button", { name: "Rename" }).click();
  await page.getByRole("button", { name: "Save" }).click();
  await expect(page.getByRole("status").filter({ hasText: /^Saved\. Composed \d+ channels/ })).toBeVisible();
  await page.getByRole("button", { name: "Open plant" }).click();
  await expect(page.getByText("2 instances · 1 links")).toBeVisible();
  await expect(page.getByRole("button", { name: "Edit plant" })).toBeVisible();
});

test("an existing plant opens in the editor from its plant screen, unchanged until edited", async ({ page }) => {
  await page.goto(`${env("CANVAS_APP_URL")}/#/plant?domain=domains%2Ffactory-utilities.domain.json`);
  await page.getByRole("button", { name: "Edit plant" }).click();
  await expect(page.getByRole("heading", { level: 1, name: "Factory utilities" })).toBeVisible();
  await expect(page.getByRole("navigation", { name: "Outline" }).getByRole("button")).toHaveCount(23);
  await expect(page.getByRole("status").filter({ hasText: "No changes" })).toBeVisible();
  await expect(page.getByRole("button", { name: "Revert" })).toHaveAttribute("aria-disabled", "true");
  await page.getByRole("navigation", { name: "Outline" }).getByRole("button", { name: "MB-PROC" }).click();
  await page.getByRole("button", { name: "Delete MB-PROC" }).click();
  await expect(page.getByRole("navigation", { name: "Outline" }).getByRole("button")).toHaveCount(22);
  await expect(page.getByRole("status").filter({ hasText: "Unsaved changes" })).toBeVisible();
  await page.getByRole("button", { name: "Revert" }).click();
  await expect(page.getByRole("navigation", { name: "Outline" }).getByRole("button")).toHaveCount(23);
});
