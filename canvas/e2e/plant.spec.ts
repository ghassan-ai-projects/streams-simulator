import AxeBuilder from "@axe-core/playwright";
import { expect, test } from "@playwright/test";
import { env } from "./support/env";

test.use({ storageState: env("CANVAS_APP_STATE") });

test("the factory opens from the workspace as a typed plant graph", async ({ page }) => {
  await page.goto(env("CANVAS_APP_URL"));
  await page.getByRole("button", { name: "Open plant factory-utilities" }).click();
  const graph = page.getByRole("region", { name: "Plant factory-utilities" });
  await expect(graph.getByText("23 instances · 59 links")).toBeVisible();
  await expect(page.getByRole("navigation", { name: "Outline" }).getByRole("button")).toHaveCount(23);
  await expect(page.locator(".react-flow__edge").first()).toBeVisible();
  await expect(page.getByRole("group", { name: /^CT-1, cooling tower,/ })).toBeVisible();
  expect((await new AxeBuilder({ page }).analyze()).violations).toEqual([]);
});

test("selecting an instance in the outline shows its kind in the inspector", async ({ page }) => {
  await page.goto(`${env("CANVAS_APP_URL")}/#/plant?domain=domains%2Ffactory-utilities.domain.json`);
  await page.getByRole("navigation", { name: "Outline" }).getByRole("button", { name: "CT-1" }).click();
  const inspector = page.getByRole("complementary", { name: "Inspector" });
  await expect(inspector.getByRole("heading", { name: "CT-1" })).toBeVisible();
  await expect(inspector.getByRole("region", { name: "Channels" })).toContainText("cooling_tower.basin_temp · degC");
  await expect(inspector.getByRole("region", { name: "Links in" })).toContainText("CH-1.heat_out → heat_in");
  await expect(inspector.getByRole("region", { name: "Faults" })).toContainText("cooling_tower__fan_failure");
  await page.getByRole("button", { name: "Workspace", exact: true }).click();
  await expect(page.getByRole("heading", { name: "lab-workspace" })).toBeVisible();
});

test("the graph works from the keyboard: arrows move, Enter selects, Escape clears, ? lists the keys", async ({
  page,
}) => {
  await page.goto(`${env("CANVAS_APP_URL")}/#/plant?domain=domains%2Ffactory-utilities.domain.json`);
  await expect(page).toHaveTitle("Factory utilities · plant · Streams Simulator");
  const graph = page.getByRole("application", { name: "Plant graph" });
  await graph.focus();
  await page.keyboard.press("ArrowRight");
  await page.keyboard.press("Enter");
  const inspector = page.getByRole("complementary", { name: "Inspector" });
  await expect(inspector.getByRole("heading", { level: 2 })).toBeVisible();
  await expect(graph.locator("[aria-live=polite]")).toContainText("selected");
  await page.keyboard.press("Escape");
  await expect(inspector).toContainText("Select an instance");
  await page.keyboard.press("?");
  await expect(page.getByRole("dialog", { name: "Keyboard" })).toBeVisible();
  await page.keyboard.press("Escape");
  await expect(page.getByRole("dialog", { name: "Keyboard" })).toBeHidden();
});
