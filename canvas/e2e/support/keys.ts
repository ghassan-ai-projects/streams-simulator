import type { Locator, Page } from "@playwright/test";
import { expect } from "@playwright/test";

interface Ring {
  readonly outline: string;
  readonly width: string;
  readonly color: string;
}

function focusRing(): Ring {
  const active = document.activeElement;
  const style = active === null ? null : getComputedStyle(active);
  return {
    outline: style?.outlineStyle ?? "",
    width: style?.outlineWidth ?? "",
    color: style?.outlineColor ?? "",
  };
}

function ringOfFocus(page: Page): Promise<Ring> {
  return page.evaluate(focusRing);
}

export async function expectVisibleRing(page: Page): Promise<void> {
  const ring = await ringOfFocus(page);
  expect(ring.outline).toBe("solid");
  expect(ring.width).not.toBe("0px");
  expect(ring.color).not.toBe("rgba(0, 0, 0, 0)");
}

function isFocused(target: Locator): Promise<boolean> {
  return target.evaluate((element) => element === document.activeElement);
}

export async function tabUntil(page: Page, target: Locator, key: "Tab" | "Shift+Tab" = "Tab"): Promise<void> {
  for (let presses = 0; presses < 80; presses += 1) {
    if (await isFocused(target)) return;
    await page.keyboard.press(key);
  }
  await expect(target).toBeFocused();
}

export async function countPointerEvents(page: Page): Promise<void> {
  await page.addInitScript(() => {
    const seen: string[] = [];
    Object.assign(window, { pointerLog: seen });
    for (const type of ["pointerdown", "mousedown", "mouseup", "wheel"]) {
      document.addEventListener(type, () => seen.push(type), true);
    }
  });
}

export function pointerEvents(page: Page): Promise<string[]> {
  return page.evaluate(() => (window as unknown as { pointerLog: string[] }).pointerLog);
}
