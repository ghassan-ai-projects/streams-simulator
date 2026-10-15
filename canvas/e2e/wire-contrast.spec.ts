import { expect, test } from "@playwright/test";
import { env } from "./support/env";

test.use({ storageState: env("CANVAS_APP_STATE") });

interface Reading {
  readonly canvas: string;
  readonly strokes: Readonly<Record<string, string>>;
  readonly dimmed: { readonly colour: string; readonly opacity: number };
}

function read(theme: string): Reading {
  document.documentElement.dataset.theme = theme;
  const style = getComputedStyle(document.documentElement);
  const value = (name: string) => style.getPropertyValue(name).trim();
  const families = ["air", "pressure", "water", "cw", "temp", "heat", "power", "signal", "env"];
  const strokes: Record<string, string> = { none: value("--border-strong") };
  for (const family of families) strokes[family] = value(`--port-${family}`);
  return {
    canvas: value("--surface-canvas"),
    strokes,
    dimmed: { colour: value("--text-tertiary"), opacity: Number(value("--wire-off-opacity")) },
  };
}

function channels(hex: string): number[] {
  return [1, 3, 5].map((at) => parseInt(hex.slice(at, at + 2), 16));
}

function luminance(rgb: readonly number[]): number {
  const [r = 0, g = 0, b = 0] = rgb.map((value) => {
    const unit = value / 255;
    return unit <= 0.03928 ? unit / 12.92 : ((unit + 0.055) / 1.055) ** 2.4;
  });
  return 0.2126 * r + 0.7152 * g + 0.0722 * b;
}

function contrast(a: readonly number[], b: readonly number[]): number {
  const [high = 0, low = 0] = [luminance(a), luminance(b)].sort((x, y) => y - x);
  return (high + 0.05) / (low + 0.05);
}

for (const theme of ["light", "dark"]) {
  test(`every wire colour reaches 3:1 against the canvas in ${theme}, the dimmed wire included`, async ({ page }) => {
    await page.goto(env("CANVAS_APP_URL"));
    const reading = await page.evaluate(read, theme);
    const canvas = channels(reading.canvas);
    for (const [name, colour] of Object.entries(reading.strokes)) {
      expect([name, contrast(channels(colour), canvas) >= 3]).toEqual([name, true]);
    }
    const over = channels(reading.dimmed.colour);
    const blended = over.map(
      (value, index) => reading.dimmed.opacity * value + (1 - reading.dimmed.opacity) * (canvas[index] ?? 0),
    );
    expect(contrast(blended, canvas)).toBeGreaterThanOrEqual(3);
  });
}
