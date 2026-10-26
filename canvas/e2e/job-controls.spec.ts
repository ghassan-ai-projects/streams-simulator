import type { Locator, Page } from "@playwright/test";
import { expect, test } from "./support/test";
import { clockWords, panelOf, pausedPanel, sameAcrossPolls, simulatedSeconds, startOpen } from "./support/jobControls";

test.setTimeout(120_000);

const fanFailure = /fan[ _]failure|Fault ·|Fault:/i;

function recordActions(page: Page): string[] {
  const names: string[] = [];
  page.on("request", (request) => {
    const named = /\/api\/actions\/([^/?]+)/.exec(request.url());
    if (request.method() === "POST" && named?.[1] !== undefined) names.push(decodeURIComponent(named[1]));
  });
  return names;
}

async function callSetpointNow(panel: Locator, value: string): Promise<void> {
  await panel.getByLabel("Call now instance").selectOption("CH-1");
  await panel.getByLabel("Call now effector").selectOption({ label: "set setpoint" });
  await panel.getByLabel("Call now setpoint c").fill(value);
  await panel.getByRole("button", { name: "Call now" }).click();
}

test("pause freezes the simulated clock for consecutive polls, step moves it one chunk, resume goes on @slow", async ({
  page,
}) => {
  const panel = await pausedPanel(page, "Explore");
  const topBar = page.getByRole("button", { name: /^Paused Factory utilities · \d\d:\d\d:\d\d of 01:00:00$/ });
  await expect(topBar).toBeVisible();
  const frozen = await sameAcrossPolls(page, () => simulatedSeconds(panel));
  expect(new Set(frozen).size).toBe(1);
  const before = frozen[0] ?? 0;
  await expect(panel.getByRole("button", { name: "Step one chunk" })).toHaveAttribute("aria-disabled", "false");
  await panel.getByRole("button", { name: "Step one chunk" }).click();
  await expect.poll(() => simulatedSeconds(panel)).toBe(before + 120);
  await expect(clockWords(panel)).toHaveText(/^Paused · /);
  expect(new Set(await sameAcrossPolls(page, () => simulatedSeconds(panel))).size).toBe(1);
  await panel.getByRole("button", { name: "Resume" }).click();
  await expect.poll(() => simulatedSeconds(panel)).toBeGreaterThan(before + 120);
  await expect(clockWords(panel)).toHaveText(/^Running · .* · 120 s per s$/);
  await expect(page.getByRole("button", { name: /^Running Factory utilities · .* · 120 s per s$/ })).toBeVisible();
  await panel.getByRole("button", { name: "Stop now" }).click();
  await expect(page.getByRole("heading", { level: 1, name: /^Run w-/ })).toBeVisible({ timeout: 60_000 });
});

test("the speed select changes the rate the run goes at, and the words follow it @smoke", async ({ page }) => {
  const panel = await startOpen(page, "Explore");
  await panel.getByRole("combobox", { name: "Speed" }).selectOption({ label: "10 s per s" });
  await expect(clockWords(panel)).toHaveText(/ · 10 s per s$/);
  await expect(page.getByRole("button", { name: /^Running Factory utilities · .* · 10 s per s$/ })).toBeVisible();
  await expect(panel.getByRole("combobox", { name: "Speed" })).toHaveValue("10");
  await panel.getByRole("button", { name: "Stop now" }).click();
  await expect(page.getByRole("heading", { level: 1, name: /^Run w-/ })).toBeVisible({ timeout: 60_000 });
});

test("an Explore run takes a call and a fault now, shows both, and Review lists them as the operator's @slow", async ({
  page,
}) => {
  const names = recordActions(page);
  const panel = await pausedPanel(page, "Explore");
  await panel.getByLabel("Call now instance").selectOption("CH-1");
  await panel.getByLabel("Call now effector").selectOption({ label: "set setpoint" });
  await panel.getByLabel("Call now setpoint c").fill("20");
  await panel.getByRole("button", { name: "Call now" }).click();
  await expect(panel.getByLabel("Call now setpoint c")).toBeFocused();
  await expect(panel.getByLabel("Call now setpoint c")).toHaveAccessibleDescription(/is a number from/);
  expect(names).not.toContain("job.invoke_effector");
  await panel.getByLabel("Call now setpoint c").fill("7");
  await panel.getByRole("button", { name: "Call now" }).click();
  const results = panel.getByRole("list", { name: "Commands given now" });
  await expect(results).toContainText(/CH-1 · set setpoint: (accepted|command not accepted: .+) · now/);
  await expect(page.getByRole("group", { name: /^CH-1, .*effector call: set setpoint · / })).toBeVisible();
  await panel.getByLabel("Fault now instance").selectOption("CT-1");
  await panel.getByLabel("Fault now fault").selectOption({ label: "fan failure" });
  await panel.getByRole("button", { name: "Inject now" }).click();
  await expect(results).toContainText(/CT-1 · fault injected: fan failure · now/);
  await expect(page.getByRole("group", { name: /^CT-1, Fault: fan failure/ })).toBeVisible();
  await panel.getByRole("button", { name: "Stop now" }).click();
  await expect(panel.getByRole("list", { name: "Effector call results" })).toContainText("· now", { timeout: 60_000 });
  await panel.getByRole("button", { name: "Open the review" }).click();
  await expect(page.getByRole("heading", { level: 1, name: /^Run w-/ })).toBeVisible({ timeout: 60_000 });
  const details = page.getByRole("complementary", { name: "Run details" });
  await expect(details.getByRole("region", { name: "Effector calls" })).toContainText("by the operator during the run");
  await expect(details.getByRole("region", { name: "Disturbances" })).toContainText("fan failure");
});

test("an Evaluate run offers no fault injection, takes a call now, and shows no fault anywhere while running @slow", async ({
  page,
}) => {
  const names = recordActions(page);
  await page.goto("/#/plant?domain=domains%2Ffactory-utilities.domain.json&launch=1");
  const planned = panelOf(page);
  await planned.getByLabel("Duration", { exact: true }).fill("60");
  await planned.getByRole("radio", { name: /^Evaluate/ }).check();
  await planned.getByRole("button", { name: "Add fault" }).click();
  await planned.getByLabel("Fault 1 instance").selectOption("CT-1");
  await planned.getByLabel("Fault 1 fault").selectOption({ label: "fan failure" });
  await planned.getByLabel("Fault 1 starts at minute").fill("3");
  await planned.getByRole("button", { name: "Start the run" }).click();
  await expect(planned.getByText(/^Running Factory utilities as run w-/)).toBeVisible();
  await expect(planned.getByRole("button", { name: "Inject now" })).toHaveCount(0);
  await expect(planned.getByLabel(/^Fault now/)).toHaveCount(0);
  await expect(planned.getByText("Faults can only be injected live in Explore runs.")).toBeVisible();
  await callSetpointNow(planned, "7");
  await expect(planned.getByRole("list", { name: "Commands given now" })).toContainText(/CH-1 · set setpoint/);
  await expect.poll(() => simulatedSeconds(planned), { timeout: 30_000 }).toBeGreaterThan(240);
  const text = await page.locator("body").innerText();
  const attributes = await page.evaluate(() =>
    [...document.body.querySelectorAll("*")].flatMap((element) => [...element.attributes].map((a) => a.value)),
  );
  expect(text).not.toMatch(fanFailure);
  expect(attributes.filter((value) => fanFailure.test(value))).toEqual([]);
  expect(await page.locator("body").ariaSnapshot()).not.toMatch(fanFailure);
  expect(names).not.toContain("job.inject_fault");
  await planned.getByRole("button", { name: "Stop now" }).click();
  await planned.getByRole("button", { name: "Open the review" }).click({ timeout: 60_000 });
  await expect(page.getByRole("heading", { level: 1, name: /^Run w-/ })).toBeVisible({ timeout: 60_000 });
});

test("a reload while the run is paused finds it paused, at the same clock, and Resume goes on @slow", async ({
  page,
}) => {
  const panel = await pausedPanel(page, "Explore");
  const frozen = await simulatedSeconds(panel);
  await page.reload();
  const banner = page.getByRole("button", { name: /^Paused Factory utilities · / });
  await expect(banner).toBeVisible();
  await banner.click();
  const again = panelOf(page);
  await expect(clockWords(again)).toHaveText(/^Paused · /);
  expect(new Set(await sameAcrossPolls(page, () => simulatedSeconds(again))).size).toBe(1);
  expect(await simulatedSeconds(again)).toBe(frozen);
  await again.getByRole("button", { name: "Resume" }).click();
  await expect.poll(() => simulatedSeconds(again)).toBeGreaterThan(frozen);
  await again.getByRole("button", { name: "Stop now" }).click();
  await expect(page.getByRole("heading", { level: 1, name: /^Run w-/ })).toBeVisible({ timeout: 60_000 });
});
