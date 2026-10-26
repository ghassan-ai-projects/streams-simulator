import type { Locator, Page } from "@playwright/test";
import { readFileSync } from "node:fs";
import { expect, test } from "./support/test";
import { startOpen } from "./support/jobControls";

test.setTimeout(180_000);

const plant = "/#/plant?domain=domains%2Ffactory-utilities.domain.json&launch=1";

function panelOf(page: Page): Locator {
  return page.getByRole("complementary", { name: "Run Factory utilities" });
}

async function runFast(page: Page, options: { readonly fault: boolean }): Promise<string> {
  await page.goto(plant);
  const panel = panelOf(page);
  await panel.getByLabel("Seed", { exact: true }).fill("11");
  await panel.getByLabel("Duration", { exact: true }).fill("5");
  await panel.getByRole("radio", { name: /As fast as possible/ }).check();
  if (options.fault) {
    await panel.getByRole("button", { name: "Add fault" }).click();
    await panel.getByLabel("Fault 1 instance").selectOption("CT-1");
    await panel.getByLabel("Fault 1 fault").selectOption({ label: "fan failure" });
    await panel.getByLabel("Fault 1 starts at minute").fill("2");
  }
  await panel.getByRole("button", { name: "Start the run" }).click();
  await panel.getByRole("button", { name: "Open the review" }).click({ timeout: 60_000 });
  const heading = page.getByRole("heading", { level: 1, name: /^Run w-/ });
  await expect(heading).toBeVisible();
  return /^Run (w-\S+)/.exec((await heading.textContent()) ?? "")?.[1] ?? "";
}

async function download(page: Page, button: Locator): Promise<{ readonly name: string; readonly text: string }> {
  const [saved] = await Promise.all([page.waitForEvent("download"), button.click()]);
  const path = await saved.path();
  return { name: saved.suggestedFilename(), text: readFileSync(path, "utf8") };
}

test("a reviewer reads what ran, checks the replay, exports a handover and compares with a baseline @slow", async ({
  page,
}) => {
  const baseline = await runFast(page, { fault: false });
  const run = await runFast(page, { fault: true });
  expect(run).not.toBe(baseline);

  const block = page.getByRole("region", { name: "This run" });
  await expect(block).toBeVisible({ timeout: 60_000 });
  await expect(block.getByRole("term")).toHaveText(["Execution", "Recording integrity", "Evidence", "Scoring"]);
  await expect(block.locator("[data-part='execution']")).toContainText("Completed");
  await expect(block.locator("[data-part='evidence']")).toContainText("Objective: Not assessed");
  await expect(block.locator("[data-part='scoring']")).toContainText("Not scored: no outside system is connected");
  await expect(block.locator("[data-part='purpose']")).toHaveText("Investigation without a stated question");
  await expect(block.locator("[data-part='visibility']")).toHaveText("Truth was visible while the run was going.");
  const questions = block.getByRole("region", { name: "Open questions" });
  const opened = block.locator("[data-part='opened-replay']");
  await expect(opened).toContainText("Replay checked when this page opened: reproduces", { timeout: 90_000 });
  await expect(questions.locator("[data-code='replay_not_checked']")).toHaveCount(0);
  await expect(block).not.toContainText("nobody has shown");
  await expect(questions.locator("[data-code='no_outside_scoring']")).toBeVisible();

  await opened.getByRole("button", { name: "Check replay now" }).click();
  await expect(block.locator("[data-part='replay-status']")).toHaveText("Replay in this report: Reproduces", {
    timeout: 90_000,
  });
  await expect(questions.locator("[data-code='replay_not_checked']")).toHaveCount(0);

  const handover = await download(page, page.getByRole("button", { name: "Export handover" }));
  const recorded = /Run ID \| `([^`]+)`/.exec(handover.text)?.[1] ?? "";
  expect(recorded).not.toBe("");
  expect(handover.name).toBe(`${recorded}-handover.md`);
  expect(handover.text).toContain("## Open questions");
  expect(handover.text).toContain("| Replay | reproduces |");
  await expect(block.getByText(`Handover saved as ${recorded}-handover.md.`)).toBeVisible();
  const confirmation = page.getByText(`Saved ${recorded}-handover.md`, { exact: true });
  await expect(confirmation).toBeVisible();
  await expect(page.getByRole("status").filter({ has: confirmation })).toHaveCount(1);
  await expect(page.getByRole("button", { name: "Export handover" })).toBeInViewport();

  await page.getByRole("button", { name: "Compare with…" }).click();
  const dialog = page.getByRole("dialog", { name: "Compare with another run" });
  await expect(dialog).toContainText("Reference (A)");
  await expect(dialog).toContainText("This run (B)");
  await dialog.getByRole("button", { name: new RegExp(baseline) }).click();
  const differs = dialog.getByRole("region", { name: "What differs" });
  await expect(differs).toBeVisible({ timeout: 60_000 });
  await expect(differs.getByRole("region", { name: "Faults" })).toContainText("Only in this run");
  await expect(differs.getByRole("region", { name: "Faults" })).toContainText("CT-1 fan failure from 00:02:00");
  const limits = dialog.getByRole("region", { name: "Comparability" });
  await expect(limits.locator("[data-code='single_run_per_side']")).toBeVisible();
  const kpis = dialog.getByRole("region", { name: "KPIs" });
  await expect(kpis).toContainText("Whole run: reference 05:00, this run 05:00 simulated");
  await expect(kpis.locator("tbody td", { hasText: /^[−+]\d/ }).first()).toBeVisible();
  await expect(dialog.getByText(/^Event-window comparisons are not available:/)).toBeVisible();
  expect(await dialog.innerText()).not.toMatch(/better|worse|winner/i);

  await page.keyboard.press("Escape");
  await expect(dialog).toHaveCount(0);
  await expect(page.getByRole("button", { name: "Compare with…" })).toBeFocused();
});

test("a run that is still going cannot be compared or exported, and the binary says why", async ({ page }) => {
  const panel = await startOpen(page, "Explore", 60);
  const world = /run (w-\S+)/.exec(
    (await panel.getByText(/^Running Factory utilities as run w-/).textContent()) ?? "",
  )?.[1];
  expect(world).toBeDefined();
  await page.goto(`/#/review?world=${world ?? ""}`);
  const compare = page.getByRole("button", { name: "Compare with…" });
  await expect(compare).toHaveAttribute("aria-disabled", "true", { timeout: 60_000 });
  const reported = await page.request.post("/api/actions/run.report", { data: { args: { run: world } } });
  const body = (await reported.json()) as { result: { open_questions: { code: string; message: string }[] } };
  const asked = body.result.open_questions.find((question) =>
    ["evaluate_unsealed", "run_not_ended"].includes(question.code),
  );
  expect(asked?.message).toMatch(/^The run has not ended/);
  await expect(compare).toHaveAttribute("title", /^The run has not ended/);
  await expect(page.getByRole("button", { name: "Export handover" })).toHaveAttribute("aria-disabled", "true");
  const refused = await page.request.post("/api/actions/run.compare", { data: { args: { a: world, b: world } } });
  expect(refused.ok()).toBe(false);
  expect(await refused.text()).toContain("run_not_ended");
  expect(await page.locator("body").innerText()).not.toMatch(/fan failure|heat wave/i);
});
