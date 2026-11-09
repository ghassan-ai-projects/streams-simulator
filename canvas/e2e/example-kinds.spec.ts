import { readdirSync } from "node:fs";
import { join } from "node:path";
import { expect, test } from "./support/test";
import { selectInOutline } from "./support/watchRun";

test.use({ target: "empty" });
test.setTimeout(120_000);

const title = "Belt conveyor, three drives with margin";

test("trying the belt conveyor brings its kind files into the workspace, so it composes and its run shows values", async ({
  page,
  lab,
}) => {
  await page.goto("/");
  await page.getByRole("button", { name: new RegExp(`^Try it\\s*: ${title}$`) }).click();
  const panel = page.getByRole("complementary", { name: `Run ${title}` });
  await expect(panel.getByRole("button", { name: "Start the run" })).toBeVisible();
  expect(readdirSync(join(lab.workspace, "kinds")).sort()).toEqual([
    "belt-controller.board.json",
    "belt_conveyor.kind.json",
    "bulk_feeder.kind.json",
    "conveyor_drive.kind.json",
    "drive-controller.board.json",
    "feeder-controller.board.json",
  ]);
  await panel.getByLabel("Seed", { exact: true }).fill("3");
  await panel.getByLabel("Duration", { exact: true }).fill("30");
  await panel.getByRole("button", { name: "Start the run" }).click();
  await expect(panel.getByText(new RegExp(`^Running ${title} as run w-`))).toBeVisible();
  await panel.getByRole("button", { name: "Close the run panel" }).click();
  await selectInOutline(page, "F-1");
  const observed = page.getByRole("region", { name: "Observed readings" });
  await expect(observed.getByRole("row", { name: /feed/ }).first()).toContainText(/\d/, { timeout: 30_000 });
  await expect(observed.getByText("Awaiting first reading")).toHaveCount(0);
});
