import { readFileSync } from "node:fs";
import { join } from "node:path";
import { expect, test } from "./support/test";

test.setTimeout(60_000);

interface Entry {
  readonly use: string;
  readonly names: readonly string[];
  readonly parameters?: Record<string, number>;
}

test("a chiller's setpoint is refused by the binary at its field when out of bounds, then edited and saved with a digest and a separate plausibility", async ({
  page,
  lab,
}) => {
  await page.goto("/#/edit?composition=compositions%2Ffactory-utilities.composition.json");
  await page.getByRole("navigation", { name: "Outline" }).getByRole("button", { name: "CH-1" }).click();
  const panel = page.getByRole("complementary", { name: "CH-1" });
  const parameters = panel.getByRole("region", { name: "Size and parameters" });
  const setpoint = parameters.getByRole("spinbutton", { name: "Chilled-water setpoint" });
  await expect(parameters.getByText("4 to 12 °C · default 7 °C")).toBeVisible();
  await setpoint.fill("40");
  await setpoint.blur();
  await page.getByRole("button", { name: "Save", exact: true }).click();
  await expect(parameters.getByRole("alert")).toContainText("40 is outside the bounds of parameter");
  await setpoint.fill("6");
  await setpoint.blur();
  await expect(parameters.getByRole("alert")).toHaveCount(0);
  await expect(page.getByRole("status").filter({ hasText: "Unsaved changes" })).toBeVisible();
  await page.getByRole("button", { name: "Save", exact: true }).click();
  const report = panel.getByRole("region", { name: "Last save" });
  await expect(report.getByText(/^Composed \d+ channels/)).toBeVisible();
  await expect(report.getByTitle(/^sha256:[0-9a-f]{64}$/)).toBeVisible();
  await expect(report.getByText("Not applicable: a composed plant has no device interface to check")).toBeVisible();
  await expect(report.getByText(/Readings stay plausible|Plausible, with|issue/)).toBeVisible({ timeout: 20_000 });
  const file = join(lab.workspace, "compositions", "factory-utilities.composition.json");
  const saved = JSON.parse(readFileSync(file, "utf8")) as { equipment: Entry[] };
  expect(saved.equipment.find((entry) => entry.names.includes("CH-1"))).toMatchObject({
    parameters: { setpoint_c: 6 },
  });
});
