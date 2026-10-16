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
  const doomed: string[] = [];
  page.on("response", (response) => {
    if (response.status() >= 400) doomed.push(`${String(response.status())} ${response.url()}`);
  });
  await expect(page.getByRole("status").filter({ hasText: /^Did not finish$/ })).toHaveCount(1, { timeout: 60_000 });
  await expect(page.getByText(reason).first()).toBeVisible();
  await expect(page.getByRole("status").filter({ hasText: /^Verified$/ })).toHaveCount(0);
  await expect(page.getByRole("status").filter({ hasText: /^Not checked$/ })).toHaveCount(0);
  expect(doomed).toEqual([]);
});

test("a world the listing still shows as open reviews from the listing, with no failed request", async ({ page }) => {
  const created = await action(page, "world.create", {
    domain: "factory-utilities",
    seed: 9,
    adapter: "native-jsonl",
    time_mode: "stepped",
  });
  const world = String(created.world_id);

  await page.goto(env("CANVAS_APP_URL"));
  const row = page
    .getByRole("region", { name: "Runs" })
    .getByRole("listitem")
    .filter({ hasText: new RegExp(`${world} ·`) });
  await expect(row).toHaveCount(1);

  const doomed: string[] = [];
  page.on("response", (response) => {
    if (response.status() >= 400) doomed.push(`${String(response.status())} ${response.url()}`);
  });
  await page.goto(`${env("CANVAS_APP_URL")}/#/run?world=${world}`);
  await expect(page.getByRole("status").filter({ hasText: /^Did not finish$/ })).toHaveCount(1, { timeout: 60_000 });
  await expect(page.getByText("run.json is absent: the run is still open or did not end").first()).toBeVisible();
  await expect(page.getByRole("status").filter({ hasText: /^Verified$/ })).toHaveCount(0);
  expect(doomed).toEqual([]);
});

test("a review opened live loads its recorded plant when the run completes", async ({ page }) => {
  const created = await action(page, "world.create", {
    domain: "factory-utilities",
    seed: 11,
    adapter: "native-jsonl",
    time_mode: "stepped",
  });
  const world = String(created.world_id);
  await page.goto(`${env("CANVAS_APP_URL")}/#/run?world=${world}`);
  await expect(page.getByRole("status").filter({ hasText: /^Did not finish$/ })).toHaveCount(1);
  const recordedPlant = page.waitForResponse((response) => {
    const url = new URL(response.url());
    return url.pathname === "/api/file" && url.searchParams.get("path") === `runs/${world}/run.json`;
  });
  await action(page, "clock.advance", { world_id: world, by_ns: 60e9 });
  await action(page, "run.end", { world_id: world });
  expect((await recordedPlant).status()).toBe(200);
  await expect(page.getByRole("status").filter({ hasText: /^Verified$/ })).toHaveCount(1);
  await expect(page.getByRole("application", { name: "Plant graph" })).toBeVisible();
});
