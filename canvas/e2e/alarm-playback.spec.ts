import { expect, test } from "@playwright/test";
import { env } from "./support/env";

test.use({ storageState: env("CANVAS_APP_STATE") });

test("reported alarm activation chips and Jump to alarm seek to the simulator's scan instant", async ({ page }) => {
  await page.goto(`${env("CANVAS_APP_URL")}/#/run?world=w-alarm`);
  await expect(page.getByRole("status").filter({ hasText: /^Verified$/ })).toHaveCount(1);
  const slider = page.getByRole("slider", { name: "Playback time" });
  await slider.fill("0");
  const chip = page
    .getByRole("list", { name: "Events" })
    .getByRole("button", { name: "00:02:20 HEAT-1 temp alarm activated" });
  await expect(chip).toContainText("00:02:20 HEAT-1");
  await expect(page.getByRole("list", { name: "Event types" })).toContainText("Alarm");
  await chip.click();
  await expect(slider).toHaveAttribute("aria-valuetext", "00:02:20");
  await expect(page.getByRole("region", { name: "Observed · HEAT-1" })).toContainText("85");
  await slider.fill("300");
  await page.getByRole("button", { name: "Jump to alarm: HEAT-1 temp alarm activated" }).click();
  await expect(slider).toHaveAttribute("aria-valuetext", "00:02:20");
  await page.getByRole("button", { name: "Report", exact: true }).click();
  await expect(page.getByRole("region", { name: "Alarms" })).toContainText("Alarms raised");
});
