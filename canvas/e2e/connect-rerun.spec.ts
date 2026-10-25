import { expect, test } from "./support/test";
import { addThermalBench, openSessionPanel, runIdOf } from "./support/session";

test.setTimeout(90_000);

test("another run set up from a connected session that never connected opens its plan form with fresh settings and says why", async ({
  page,
  lab,
}) => {
  addThermalBench(lab);
  const panel = await openSessionPanel(page);
  await panel.getByLabel("Host socket", { exact: true }).fill("/tmp/ssc-absent/h.sock");
  await panel.getByRole("button", { name: "Start the session" }).click();
  const failure = panel.getByRole("alert");
  await expect(failure).toContainText("plantwire connect", { timeout: 30_000 });
  const run = runIdOf(await panel.getByText(/^What ran is kept in run c-[0-9a-f]{8}\.$/).textContent());

  await page.goto(`/#/review?world=${run}`);
  await expect(page.getByRole("heading", { level: 1, name: `Run ${run}` })).toBeVisible({ timeout: 30_000 });
  await page.getByRole("button", { name: "Set up another run" }).click();

  await expect(panel.getByRole("alert")).toHaveCount(0);
  await expect(panel.getByRole("radio", { name: /^Connected session/ })).toBeChecked();
  await expect(panel.getByLabel("Host socket", { exact: true })).toBeVisible();
  await expect(panel.getByLabel("Host socket", { exact: true })).toHaveValue("");
  await expect(panel.getByRole("note")).toHaveText(
    "The plan of this run could not be recovered because it did not finish; these are fresh settings.",
  );
});
