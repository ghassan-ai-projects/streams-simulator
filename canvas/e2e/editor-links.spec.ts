import { expect, test } from "./support/test";
import { addEquipment, dragPin, openEmptyPlant, pin } from "./support/editorPlant";
import { waitForViewport } from "./support/viewport";

function transform(): string {
  return document.querySelector<HTMLElement>(".react-flow__viewport")?.style.transform ?? "";
}

test.beforeEach(async ({ page }) => {
  await openEmptyPlant(page, "Drag lab");
  await addEquipment(page, "Water-cooled", "Water-cooled chiller", "CH-1");
  await addEquipment(page, "Cooling", "Cooling tower", "CT-1");
  await addEquipment(page, "pump", "Centrifugal pump", "P-1");
  await waitForViewport(page);
});

test("dragging from an output pin to a matching input pin links them, and Undo removes the link", async ({ page }) => {
  const edges = page.locator(".react-flow__edge");
  await expect(edges).toHaveCount(0);
  await dragPin(page, pin(page, "CH-1", "heat_out"), pin(page, "CT-1", "heat_in"));
  await expect(edges).toHaveCount(1);
  await expect(page.getByRole("alert")).toHaveCount(0);
  await page.getByRole("navigation", { name: "Outline" }).getByRole("button", { name: "CT-1" }).click();
  await expect(page.getByRole("heading", { name: "Links 1" })).toBeVisible();
  await expect(page.getByRole("status").filter({ hasText: "Unsaved changes" })).toBeVisible();
  await page.getByRole("button", { name: "Undo" }).click();
  await expect(edges).toHaveCount(0);
  await expect(page.getByRole("heading", { name: "Links 0" })).toBeVisible();
});

test("dropping an output on an input of another type is refused with the reason, and nothing is linked", async ({
  page,
}) => {
  await dragPin(page, pin(page, "CH-1", "heat_out"), pin(page, "P-1", "level_in"));
  const alert = page.getByRole("alert");
  await expect(alert).toContainText("CH-1's heat out is heat rejection");
  await expect(alert).toContainText("P-1's level in takes level");
  await expect(page.locator(".react-flow__edge")).toHaveCount(0);
  await page.getByRole("navigation", { name: "Outline" }).getByRole("button", { name: "CH-1" }).click();
  await expect(page.getByRole("heading", { name: "Links 0" })).toBeVisible();
  await dragPin(page, pin(page, "CH-1", "heat_out"), pin(page, "CT-1", "heat_in"));
  await expect(page.locator(".react-flow__edge")).toHaveCount(1);
  await expect(alert).toHaveCount(0);
});

test("dragging from empty canvas pans the view and links nothing", async ({ page }) => {
  const before = await page.evaluate(transform);
  const graph = page.getByRole("application", { name: "Plant graph" });
  await graph.dragTo(graph, { sourcePosition: { x: 60, y: 60 }, targetPosition: { x: 160, y: 120 } });
  await expect.poll(() => page.evaluate(transform)).not.toBe(before);
  await expect(page.locator(".react-flow__edge")).toHaveCount(0);
  await expect(page.getByRole("alert")).toHaveCount(0);
});
