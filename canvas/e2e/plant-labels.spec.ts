import { expect, test } from "./support/test";
import { guard } from "./support/audit";
import { zoomInTo } from "./support/viewport";
import { textClashes } from "./support/textClashes";

test("no two labels inside one node intersect at the compact zoom, on a flat plant and on a hall", async ({ page }) => {
  for (const view of [
    "domain=domains%2Ffactory-utilities.domain.json",
    "domain=domains%2Fcampus-3.domain.json&hall=h01",
  ]) {
    await page.goto(`/#/plant?${view}`);
    const graph = page.getByRole("application", { name: "Plant graph" });
    await graph.focus();
    await page.keyboard.press("0");
    expect(await zoomInTo(page, "compact", 3)).toBe("compact");
    await expect(page.locator('.react-flow__node-instance [data-part="pin-name"]').first()).toBeVisible();
    expect(guard(await page.evaluate(textClashes, 50))).toEqual([]);
  }
});
