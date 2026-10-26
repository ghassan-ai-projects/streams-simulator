import { expect, test } from "./support/test";
import { partsOverlapping, selectInOutline, rowsOutsideNode, startWatchRun } from "./support/watchRun";
import type { RunMode } from "./support/watchRun";
import { zoomInTo } from "./support/viewport";

test.setTimeout(120_000);

const busiest = [
  ["CH-1", 8],
  ["M-1", 7],
  ["R-1", 4],
  ["CT-1", 4],
] as const;

async function atCompactZoom(page: Parameters<typeof startWatchRun>[0], mode: RunMode) {
  await startWatchRun(page, mode);
  await page.getByRole("application", { name: "Plant graph" }).focus();
  await page.keyboard.press("0");
  expect(await zoomInTo(page, "compact", 4)).toBe("compact");
  await expect(page.getByText(/^Running · 00:(0[2-9]|[1-5]\d)/)).toBeVisible({ timeout: 60_000 });
}

for (const mode of ["Explore", "Evaluate"] as RunMode[]) {
  test(`${mode}: every readout row of the busiest live nodes lies inside its node box`, async ({ page }) => {
    await atCompactZoom(page, mode);
    for (const [id, minimum] of busiest) {
      await selectInOutline(page, id);
      await expect
        .poll(async () => (await page.evaluate(rowsOutsideNode, id)).values, { timeout: 30_000 })
        .toBeGreaterThanOrEqual(1);
      const fit = await page.evaluate(rowsOutsideNode, id);
      expect(fit.rows, `${id} rows`).toBeGreaterThanOrEqual(minimum);
      expect(fit.outside, `${id} rows outside the node`).toEqual([]);
    }
  });
}

test("a fault note that appears on a live node grows the node to hold it", async ({ page }) => {
  await atCompactZoom(page, "Explore");
  await selectInOutline(page, "CT-1");
  await expect(page.getByRole("group", { name: /^CT-1, / })).toContainText("Fault · fan failure", {
    timeout: 30_000,
  });
  await expect.poll(async () => (await page.evaluate(rowsOutsideNode, "CT-1")).outside).toEqual([]);
});

test("a node with a fault and a command status line holds both without overlapping its id, subtitle or readings", async ({
  page,
}) => {
  await startWatchRun(page, "Explore");
  const panel = page.getByRole("complementary", { name: "Run Factory utilities" });
  await page.getByRole("button", { name: /^Running Factory utilities · / }).click();
  await expect(panel.getByText(/^Running · 00:0[2-9]/)).toBeVisible({ timeout: 60_000 });
  await panel.getByRole("button", { name: "Pause" }).click();
  await panel.getByLabel("Call now instance").selectOption("CT-1");
  await panel.getByLabel("Call now effector").selectOption({ label: "set fan speed" });
  await panel.getByLabel("Call now speed").fill("0.6");
  await panel.getByRole("button", { name: "Call now" }).click();
  await panel.getByRole("button", { name: "Close the run panel" }).click();
  await page.getByRole("application", { name: "Plant graph" }).focus();
  await page.keyboard.press("0");
  expect(await zoomInTo(page, "compact", 4)).toBe("compact");
  const node = page.getByRole("group", { name: /^CT-1, / });
  await expect(node).toContainText("Fault · fan failure", { timeout: 30_000 });
  await expect(node).toContainText("set fan speed · accepted");
  await expect.poll(async () => page.evaluate(partsOverlapping, "CT-1")).toEqual([]);
  await expect.poll(async () => (await page.evaluate(rowsOutsideNode, "CT-1")).outside).toEqual([]);
});
