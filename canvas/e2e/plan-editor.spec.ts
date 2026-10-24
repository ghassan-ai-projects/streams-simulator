import type { Page } from "@playwright/test";
import { expect, test } from "./support/test";

test.setTimeout(120_000);

const plant = "/#/plant?domain=domains%2Ffactory-utilities.domain.json&launch=1";
const name = "Cooling loss and recovery";
const purpose = "How does cooling respond when the tower fan fails and recovers?";

function panelOf(page: Page) {
  return page.getByRole("complementary", { name: "Run Factory utilities" });
}

async function planned(page: Page, options: { readonly evaluate?: boolean; readonly fast?: boolean } = {}) {
  await page.goto(plant);
  const panel = panelOf(page);
  await panel.getByLabel("Name", { exact: true }).fill(name);
  await panel.getByLabel("Purpose", { exact: true }).fill(purpose);
  await panel.getByLabel("Seed", { exact: true }).fill("11");
  await panel.getByLabel("Duration", { exact: true }).fill(options.fast === true ? "5" : "30");
  if (options.evaluate === true) await panel.getByRole("radio", { name: /^Evaluate/ }).check();
  if (options.fast === true) await panel.getByRole("radio", { name: /As fast as possible/ }).check();
  await panel.getByRole("button", { name: "Add fault" }).click();
  await panel.getByLabel("Fault 1 instance").selectOption("CT-1");
  await panel.getByLabel("Fault 1 fault").selectOption({ label: "fan failure" });
  await panel.getByLabel("Fault 1 starts at minute").fill("1");
  await panel.getByLabel("Fault 1 ends at minute").fill("3");
  await panel.getByRole("button", { name: "Add delivery condition" }).click();
  await panel.getByLabel("Delivery condition 1 kind").selectOption("drop");
  await panel.getByLabel("Delivery condition 1 starts at minute").fill("1");
  await panel.getByLabel("Delivery condition 1 ends at minute").fill("4");
  await panel.getByLabel("Delivery condition 1 rate").fill("0.5");
  await panel.getByRole("button", { name: "Add effector call" }).click();
  await panel.getByLabel("Effector call 1 instance").selectOption("CH-1");
  await panel.getByLabel("Effector call 1 effector").selectOption({ label: "set setpoint" });
  await panel.getByLabel("Effector call 1 setpoint c").fill("7");
  await panel.getByLabel("Effector call 1 starts at minute").fill("2");
  return panel;
}

test("a named plan with a fault that recovers, a delivery condition and a call reads in time order, runs, and is set up again exactly", async ({
  page,
}) => {
  const panel = await planned(page, { fast: true });
  const steps = panel.getByRole("list", { name: "Plan in time order" }).getByRole("listitem");
  await expect(steps).toHaveCount(5);
  await expect(steps.nth(0)).toContainText(/01:00.*fault.*CT-1 fan failure/i);
  await expect(steps.nth(1)).toContainText(/01:00.*delivery.*drop starts \(rate 0\.5\)/i);
  await expect(steps.nth(2)).toContainText(/02:00.*command.*CH-1 set setpoint \(setpoint c 7\)/i);
  await expect(steps.nth(3)).toContainText(/03:00.*clear.*CT-1 fan failure ends/i);
  await expect(steps.nth(4)).toContainText(/04:00.*delivery.*drop ends/i);
  await steps.nth(3).getByRole("button").click();
  await expect(panel.getByLabel("Fault 1 ends at minute")).toBeFocused();
  const review = panel.getByRole("region", { name: "Before you start" });
  await expect(review).toContainText("Unscored investigation: no outside system is scoring this run.");
  await expect(review).toContainText(name);

  await panel.getByRole("button", { name: "Start the run" }).click();
  await panel.getByRole("button", { name: "Open the review" }).click({ timeout: 60_000 });
  const heading = page.getByRole("heading", { level: 1, name: /^Run w-/ });
  await expect(heading).toBeVisible();
  await expect(heading).toHaveAccessibleDescription(name);
  await expect(page.getByText(name, { exact: true })).toBeVisible();
  await expect(page.getByRole("slider", { name: "Playback time" })).toHaveAttribute("max", "300", { timeout: 60_000 });

  await page.getByRole("button", { name: "Set up another run" }).click();
  const again = panelOf(page);
  await expect(again.getByLabel("Name", { exact: true })).toHaveValue(name);
  await expect(again.getByLabel("Purpose", { exact: true })).toHaveValue(purpose);
  await expect(again.getByLabel("Fault 1 instance")).toHaveValue("CT-1");
  await expect(again.getByLabel("Fault 1 starts at minute")).toHaveValue("1");
  await expect(again.getByLabel("Fault 1 ends at minute")).toHaveValue("3");
  await expect(again.getByLabel("Delivery condition 1 kind")).toHaveValue("drop");
  await expect(again.getByLabel("Delivery condition 1 starts at minute")).toHaveValue("1");
  await expect(again.getByLabel("Delivery condition 1 ends at minute")).toHaveValue("4");
  await expect(again.getByLabel("Delivery condition 1 rate")).toHaveValue("0.5");
  await expect(again.getByLabel("Effector call 1 setpoint c")).toHaveValue("7");
  const note = again.getByRole("note");
  await expect(note).toContainText("Prefilled from w-");
  await expect(note).not.toContainText("Not copied");
  await expect(again.getByRole("list", { name: "Plan in time order" }).getByRole("listitem")).toHaveCount(5);

  await page.getByRole("button", { name: "Streams Simulator, workspace home" }).click();
  await expect(page.getByRole("region", { name: "Runs" }).getByRole("listitem").filter({ hasText: name })).toHaveCount(
    1,
  );
});

test("an exploration found again after a reload brings back its name, its fault marks and its plan ticks", async ({
  page,
}) => {
  const panel = await planned(page);
  await panel.getByRole("button", { name: "Start the run" }).click();
  await expect(panel.getByText(new RegExp(`^Running ${name} on Factory utilities as run w-`))).toBeVisible();
  await expect(panel.getByText(/^Running · 00:0[2-9]/)).toBeVisible({ timeout: 30_000 });
  await page.reload();
  const banner = page.getByRole("button", { name: new RegExp(`^Running ${name} · 00:`) });
  await expect(banner).toBeVisible();
  await banner.click();
  await expect(panel.getByText(purpose)).toBeVisible();
  await expect(panel.getByText("30 min of simulated time · Explore · 1 fault · 1 delivery condition")).toBeVisible();
  await expect(page.getByRole("group", { name: /^CT-1, Fault: fan failure/ })).toBeVisible();
  await expect(panel.locator("[title='CT-1 fan failure at 00:01:00']")).toHaveCount(1);
  await panel.getByRole("button", { name: "Stop now" }).click();
  await panel.getByRole("button", { name: "Open the review" }).click({ timeout: 60_000 });
  await expect(page.getByRole("heading", { level: 1, name: /^Run w-/ })).toBeVisible();
});

test("an evaluation found again after a reload keeps its name and purpose and still names no fault or delivery condition", async ({
  page,
}) => {
  const panel = await planned(page, { evaluate: true });
  await panel.getByRole("button", { name: "Start the run" }).click();
  await expect(panel.getByText(/^Running · 00:0[1-9]/)).toBeVisible({ timeout: 30_000 });
  await page.reload();
  await page.getByRole("button", { name: new RegExp(`^Running ${name} · `) }).click();
  await expect(panel.getByText(purpose)).toBeVisible();
  await expect(panel.getByText("30 min of simulated time · Evaluate")).toBeVisible();
  const text = await page.locator("body").innerText();
  expect(text).not.toMatch(/fan[ _]failure|Fault ·|drop starts|rate 0\.5/i);
  await expect(page.getByRole("list", { name: "Plan in time order" })).toHaveCount(0);
  await panel.getByRole("button", { name: "Stop now" }).click();
  await panel.getByRole("button", { name: "Open the review" }).click({ timeout: 60_000 });
  await expect(page.getByRole("heading", { level: 1, name: /^Run w-/ })).toBeVisible();
});

test("a plan the binary refuses shows the binary's own message and keeps the plan", async ({ page }) => {
  await page.route("**/api/actions/job.start", async (route) => {
    const sent = route.request().postDataJSON() as { args: { faults: { at_s: number; until_s?: number }[] } };
    const [fault] = sent.args.faults;
    if (fault !== undefined) fault.until_s = fault.at_s;
    await route.continue({ postData: JSON.stringify(sent) });
  });
  const panel = await planned(page, { fast: true });
  await panel.getByRole("button", { name: "Start the run" }).click();
  const alert = page.getByRole("alert").filter({ hasText: "The run did not start." });
  await expect(alert).toContainText("until_s must be after at_s and within the run, up to duration_s");
  await alert.getByRole("button", { name: "Change the plan" }).click();
  await expect(panel.getByLabel("Name", { exact: true })).toHaveValue(name);
  await expect(panel.getByLabel("Fault 1 ends at minute")).toHaveValue("3");
});
