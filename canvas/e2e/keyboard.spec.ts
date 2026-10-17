import { expect, test } from "./support/test";

function focused(): { readonly tag: string; readonly outline: string; readonly width: string; readonly color: string } {
  const active = document.activeElement;
  const style = active === null ? null : getComputedStyle(active);
  return {
    tag: active?.tagName ?? "",
    outline: style?.outlineStyle ?? "",
    width: style?.outlineWidth ?? "",
    color: style?.outlineColor ?? "",
  };
}

test("the keyboard dialog opens on a heading whose focus is visible, in both themes", async ({ page }) => {
  await page.goto("/#/plant?domain=domains%2Ffactory-utilities.domain.json");
  await expect(page.getByRole("button", { name: "Keyboard" })).toBeVisible();
  for (const theme of ["light", "dark"]) {
    await page.evaluate((value) => {
      document.documentElement.dataset.theme = value;
    }, theme);
    await page.keyboard.press("?");
    await expect(page.getByRole("dialog", { name: "Keyboard" })).toBeVisible();
    const ring = await page.evaluate(focused);
    expect(ring.tag).toBe("H2");
    expect(ring.outline).toBe("solid");
    expect(ring.width).not.toBe("0px");
    expect(ring.color).not.toBe("rgba(0, 0, 0, 0)");
    await page.keyboard.press("Escape");
    await expect(page.getByRole("dialog", { name: "Keyboard" })).toBeHidden();
  }
});
