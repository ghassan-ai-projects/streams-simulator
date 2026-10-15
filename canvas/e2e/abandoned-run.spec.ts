import type { Page } from "@playwright/test";
import { expect, test } from "@playwright/test";
import { env } from "./support/env";

test.use({ storageState: env("CANVAS_APP_STATE") });
test.setTimeout(90_000);

const reason = "the world was destroyed before its run ended";

async function action(page: Page, name: string, args: Record<string, unknown>): Promise<Record<string, unknown>> {
  const response = await page.request.post(`${env("CANVAS_APP_URL")}/api/actions/${name}`, { data: { args } });
  const body = (await response.json()) as { result?: Record<string, unknown> };
  if (!response.ok() || body.result === undefined) throw new Error(`${name} failed: ${await response.text()}`);
  return body.result;
}

test("a world created and destroyed through the bridge lists and reviews as Did not finish", async ({ page }) => {
  const created = await action(page, "world.create", {
    domain: "factory-utilities",
    seed: 5,
    adapter: "native-jsonl",
    time_mode: "stepped",
  });
  const world = String(created.world_id);
  await action(page, "world.destroy", { world_id: world });

  await page.goto(env("CANVAS_APP_URL"));
  const row = page
    .getByRole("region", { name: "Runs" })
    .getByRole("listitem")
    .filter({ hasText: new RegExp(`${world} ·`) });
  await expect(row).toContainText("seed 5");
  await expect(row).toContainText(`Did not finish: ${reason}`);
  await expect(row).not.toContainText(/Verified|Not checked/);
  await expect(row.getByText("Did not finish", { exact: true })).toHaveAttribute("title", new RegExp(reason));

  await page.goto(`${env("CANVAS_APP_URL")}/#/run?world=${world}`);
  await expect(page.getByRole("status").filter({ hasText: /^Did not finish$/ })).toHaveCount(1, { timeout: 60_000 });
  await expect(page.getByText(reason).first()).toBeVisible();
  await expect(page.getByRole("status").filter({ hasText: /^Verified$/ })).toHaveCount(0);
});
