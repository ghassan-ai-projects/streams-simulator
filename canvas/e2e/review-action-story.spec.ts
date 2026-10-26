import type { Page } from "@playwright/test";
import { expect, test } from "./support/test";
import { reviewOfFastRun } from "./support/reviewRun";

test.setTimeout(120_000);

const run = { minutes: 20, fault: { at: 5, until: 9 }, call: { at: 3, setpoint: 7 } };

function playbackAt(page: Page) {
  return page.getByRole("slider", { name: "Playback time" });
}

test("a click anywhere on an effector call row opens its action story, and Show 1 min before moves the cursor back 60 s @smoke", async ({
  page,
}) => {
  const details = await reviewOfFastRun(page, run);
  const calls = details.getByRole("region", { name: "Effector calls" });
  await calls.getByRole("cell", { name: "set setpoint" }).click();
  const story = details.getByRole("region", { name: "Action story" });
  await expect(story).toBeVisible();
  await expect(story).toContainText("setpoint_c = 7");
  await expect(playbackAt(page)).toHaveValue("180");
  const at = 180;
  await expect(story.getByRole("button", { name: "Show 1 min before" })).toBeInViewport();
  await story.getByRole("button", { name: "Show 1 min before" }).click();
  await expect(playbackAt(page)).toHaveValue(String(at - 60));
  await story.getByRole("button", { name: "Show 1 min after" }).click();
  await expect(playbackAt(page)).toHaveValue(String(at + 60));
});

test("a call picked on the timeline opens the story in the report even from the Inspect tab", async ({ page }) => {
  const details = await reviewOfFastRun(page, run);
  await details.getByRole("button", { name: /^Inspect/ }).click();
  await expect(details.getByRole("region", { name: "Inspect" })).toBeVisible();
  await page.getByRole("button", { name: /^00:03:00 CH-1/ }).click();
  await expect(details.getByRole("button", { name: "Report" })).toHaveAttribute("aria-pressed", "true");
  await expect(details.getByRole("region", { name: "Action story" })).toContainText("set setpoint on CH-1");
});

test("a click on a disturbance row moves the cursor to its start and marks the row", async ({ page }) => {
  const details = await reviewOfFastRun(page, run);
  const disturbances = details.getByRole("region", { name: "Disturbances" });
  await disturbances.getByRole("cell", { name: "fan failure" }).click();
  await expect(playbackAt(page)).toHaveValue("300");
  await expect(disturbances.getByRole("row", { selected: false }).filter({ hasText: "fan failure" })).toHaveAttribute(
    "aria-current",
    "true",
  );
  await expect(disturbances.getByRole("cell", { name: "00:09:00" })).toBeVisible();
});
