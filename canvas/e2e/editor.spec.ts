import AxeBuilder from "@axe-core/playwright";
import type { Page } from "@playwright/test";
import { expect, test } from "@playwright/test";
import { execFileSync } from "node:child_process";
import { mkdtempSync, readFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { env } from "./support/env";

test.use({ storageState: env("CANVAS_APP_STATE") });
test.setTimeout(60_000);

async function add(page: Page, component: string, name: string) {
  const build = page.getByRole("complementary", { name: "Build" });
  await build.getByLabel("Find equipment").fill(component.split(" ")[0] ?? "");
  await build.getByLabel("Equipment", { exact: true }).selectOption({ label: component });
  await expect(build.getByLabel("Size").getByRole("option")).not.toHaveCount(1);
  await build.getByLabel("Name").fill(name);
  await build.getByLabel("Area").selectOption("Chiller plant");
  await build.getByRole("button", { name: "Add to the plant" }).click();
  await expect(page.getByRole("complementary", { name })).toBeVisible();
  await page.getByRole("button", { name: "Close the instance" }).click();
}

function pin(page: Page, node: string, port: string) {
  return page.locator(`.react-flow__node[data-id="${node}"] .react-flow__handle[data-handleid="${port}"]`);
}

test("a new plant is built in Chrome, refused with the issue on its element, fixed, composed and opened", async ({
  page,
}) => {
  await page.goto(env("CANVAS_APP_URL"));
  await page.getByRole("button", { name: "New plant" }).click();
  const form = page.getByRole("form", { name: "New plant" });
  await form.getByLabel("Name").fill("Lab cooling");
  await expect(form.getByLabel("Id")).toHaveValue("lab-cooling");
  await form.getByLabel("Start from").selectOption({ label: "Factory utilities" });
  await form.getByRole("radio", { name: /^An empty plant/ }).check();
  await form.getByRole("button", { name: "Create and open the editor" }).click();
  await expect(page.getByRole("heading", { level: 1, name: "Lab cooling" })).toBeVisible();
  await expect(page.getByText("No equipment yet.")).toBeVisible();
  await add(page, "Water-cooled chiller", "CH-1");
  await add(page, "Cooling tower", "CT 1");
  await pin(page, "CH-1", "heat_out").dragTo(pin(page, "CT 1", "heat_in"));
  await page.getByRole("navigation", { name: "Outline" }).getByRole("button", { name: "CT 1" }).click();
  await expect(page.getByRole("heading", { name: "Links (1)" })).toBeVisible();
  expect((await new AxeBuilder({ page }).analyze()).violations).toEqual([]);
  await page.getByRole("button", { name: "Save" }).click();
  await expect(page.getByRole("status").filter({ hasText: /^Saved, not composed/ })).toBeVisible();
  await expect(page.getByRole("group", { name: /^CT 1, .*has issues$/ })).toBeVisible();
  await expect(page.getByRole("region", { name: "Issues on CT 1" })).toContainText("does not match pattern");
  await page.getByLabel("Name").fill("CT-1");
  await page.getByRole("button", { name: "Rename" }).click();
  await page.getByRole("button", { name: "Save" }).click();
  await expect(page.getByRole("status").filter({ hasText: /^Saved\. Composed \d+ channels/ })).toBeVisible();
  const folder = env("CANVAS_WORKSPACE");
  const cli = join(mkdtempSync(join(tmpdir(), "compose-")), "lab.domain.json");
  execFileSync(env("CANVAS_BINARY"), [
    "domain",
    "compose",
    join(folder, "compositions", "lab-cooling.composition.json"),
    "--out",
    cli,
  ]);
  expect(readFileSync(join(folder, "domains", "lab-cooling.domain.json"), "utf8")).toBe(readFileSync(cli, "utf8"));
  await page.getByRole("button", { name: "Open plant" }).click();
  await expect(page.getByText("2 instances · 1 links")).toBeVisible();
  await expect(page.getByRole("button", { name: "Edit plant" })).toBeVisible();
});

test("an existing plant opens in the editor from its plant screen, unchanged until edited", async ({ page }) => {
  await page.goto(`${env("CANVAS_APP_URL")}/#/plant?domain=domains%2Ffactory-utilities.domain.json`);
  await page.getByRole("button", { name: "Edit plant" }).click();
  await expect(page.getByRole("heading", { level: 1, name: "Factory utilities" })).toBeVisible();
  await expect(page.getByRole("navigation", { name: "Outline" }).getByRole("button")).toHaveCount(23);
  await expect(page.getByRole("status").filter({ hasText: "No changes" })).toBeVisible();
  await expect(page.getByRole("button", { name: "Revert" })).toHaveAttribute("aria-disabled", "true");
  await page.getByRole("navigation", { name: "Outline" }).getByRole("button", { name: "MB-PROC" }).click();
  await page.getByRole("button", { name: "Delete MB-PROC" }).click();
  await expect(page.getByRole("navigation", { name: "Outline" }).getByRole("button")).toHaveCount(22);
  await expect(page.getByRole("status").filter({ hasText: "Unsaved changes" })).toBeVisible();
  await page.getByRole("button", { name: "Revert" }).click();
  await page.getByRole("button", { name: "Discard changes" }).click();
  await expect(page.getByRole("navigation", { name: "Outline" }).getByRole("button")).toHaveCount(23);
});

test("a plant copied from a shipped example opens in the editor with its equipment", async ({ page }) => {
  await page.goto(env("CANVAS_APP_URL"));
  await page.getByRole("button", { name: "New plant" }).click();
  const form = page.getByRole("form", { name: "New plant" });
  await form.getByLabel("Start from").selectOption({ label: "Chiller pilot" });
  await form.getByLabel("Name").fill("Pilot copy");
  await form.getByRole("button", { name: "Create and open the editor" }).click();
  await expect(page.getByRole("heading", { level: 1, name: "Pilot copy" })).toBeVisible();
  await expect(page.getByRole("navigation", { name: "Outline" }).getByRole("button")).not.toHaveCount(0);
});
