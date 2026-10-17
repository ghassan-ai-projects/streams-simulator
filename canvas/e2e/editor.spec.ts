import AxeBuilder from "@axe-core/playwright";
import type { Page } from "@playwright/test";
import { expect, test } from "./support/test";
import { execFileSync } from "node:child_process";
import { mkdtempSync, readFileSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";

test.setTimeout(60_000);

async function add(page: Page, component: string, name: string) {
  const build = page.getByRole("complementary", { name: "Build" });
  await build.getByLabel("Find equipment").fill(component.split(" ")[0] ?? "");
  await build.getByLabel("Equipment", { exact: true }).selectOption({ label: component });
  await expect(build.getByLabel("Size").getByRole("option")).not.toHaveCount(1);
  await build.getByLabel("Name", { exact: true }).fill(name);
  await build.getByLabel("Area", { exact: true }).selectOption({ index: 0 });
  await build.getByRole("button", { name: "Add to the plant" }).click();
  await expect(page.getByRole("complementary", { name })).toBeVisible();
  await page.getByRole("button", { name: "Close the instance" }).click();
}

function pin(page: Page, node: string, port: string) {
  return page.locator(`.react-flow__node[data-id="${node}"] .react-flow__handle[data-handleid="${port}"]`);
}

test("a new plant is built in Chrome, a name the binary would refuse is stopped in the field, then it composes and opens", async ({
  page,
  lab,
}) => {
  await page.goto("/");
  await page.getByRole("button", { name: "New plant" }).click();
  const form = page.getByRole("form", { name: "New plant" });
  await form.getByLabel("Name").fill("Lab cooling");
  await expect(form.getByLabel("Id")).toHaveValue("lab-cooling");
  await form.getByLabel("Start from").selectOption({ label: "Factory utilities" });
  await form.getByRole("radio", { name: /^An empty plant/ }).check();
  await form.getByRole("button", { name: "Create and open the editor" }).click();
  await expect(page.getByRole("heading", { level: 1, name: "Lab cooling" })).toBeVisible();
  await expect(page.getByRole("heading", { name: "No equipment yet" })).toBeVisible();
  await add(page, "Water-cooled chiller", "CH-1");
  const build = page.getByRole("complementary", { name: "Build" });
  await build.getByLabel("Find equipment").fill("Cooling");
  await build.getByLabel("Equipment", { exact: true }).selectOption({ label: "Cooling tower" });
  await build.getByLabel("Name", { exact: true }).fill("CT 1");
  await expect(build.getByText(/"CT 1" is not a valid name\. An instance name: a letter or digit/)).toBeVisible();
  await expect(build.getByRole("button", { name: "Add to the plant" })).toHaveAttribute("aria-disabled", "true");
  await build.getByLabel("Name", { exact: true }).fill("CT-1");
  await build.getByRole("button", { name: "Add to the plant" }).click();
  await page.getByRole("button", { name: "Close the instance" }).click();
  await expect(pin(page, "CH-1", "heat_out")).toHaveCSS("cursor", "crosshair");
  await pin(page, "CH-1", "heat_out").dragTo(pin(page, "CT-1", "heat_in"));
  await page.getByRole("navigation", { name: "Outline" }).getByRole("button", { name: "CT-1" }).click();
  await expect(page.getByRole("heading", { name: "Links 1" })).toBeVisible();
  expect((await new AxeBuilder({ page }).analyze()).violations).toEqual([]);
  await page.getByRole("button", { name: "Save" }).click();
  await expect(page.getByRole("status").filter({ hasText: /^Ready to run\. Saved\.$/ })).toBeVisible();
  const folder = lab.workspace;
  const cli = join(mkdtempSync(join(tmpdir(), "compose-")), "lab.domain.json");
  execFileSync(lab.binary, [
    "domain",
    "compose",
    join(folder, "compositions", "lab-cooling.composition.json"),
    "--out",
    cli,
  ]);
  expect(readFileSync(join(folder, "domains", "lab-cooling.domain.json"), "utf8")).toBe(readFileSync(cli, "utf8"));
  await page.getByRole("button", { name: "Open plant" }).click();
  await expect(page.getByText("2 instances · 1 link")).toBeVisible();
  await expect(page.getByRole("button", { name: "Edit plant" })).toBeVisible();
});

test("an existing plant opens in the editor from its plant screen, unchanged until edited", async ({ page }) => {
  await page.goto("/#/plant?domain=domains%2Ffactory-utilities.domain.json");
  await page.getByRole("button", { name: "Edit plant" }).click();
  await expect(page.getByRole("heading", { level: 1, name: "Factory utilities" })).toBeVisible();
  await expect(page.getByRole("navigation", { name: "Outline" }).getByRole("button")).toHaveCount(23);
  await expect(page.getByRole("status").filter({ hasText: "No changes" })).toBeVisible();
  await expect(page.getByRole("button", { name: "Revert" })).toHaveAttribute("aria-disabled", "true");
  await page.getByRole("navigation", { name: "Outline" }).getByRole("button", { name: "MB-PROC" }).click();
  await page.getByRole("button", { name: "Delete MB-PROC" }).click();
  await page.getByRole("button", { name: /^Delete MB-PROC and its \d+ links?\?$/ }).click();
  await expect(page.getByRole("navigation", { name: "Outline" }).getByRole("button")).toHaveCount(22);
  await expect(page.getByRole("status").filter({ hasText: "Unsaved changes" })).toBeVisible();
  await page.getByRole("button", { name: "Revert" }).click();
  await page.getByRole("button", { name: "Discard changes" }).click();
  await expect(page.getByRole("navigation", { name: "Outline" }).getByRole("button")).toHaveCount(23);
});

test("leaving the editor with an unsaved change asks first, Stay keeps the draft and Discard and leave drops it", async ({
  page,
}) => {
  await page.goto("/#/plant?domain=domains%2Ffactory-utilities.domain.json");
  await page.getByRole("button", { name: "Edit plant" }).click();
  await expect(page.getByRole("heading", { level: 1, name: "Factory utilities" })).toBeVisible();
  await page.getByRole("navigation", { name: "Outline" }).getByRole("button", { name: "MB-PROC" }).click();
  await page.getByRole("button", { name: "Delete MB-PROC" }).click();
  await page.getByRole("button", { name: /^Delete MB-PROC and its \d+ links?\?$/ }).click();
  await expect(page.getByRole("status").filter({ hasText: "Unsaved changes" })).toBeVisible();
  await page.getByRole("button", { name: "Workspace", exact: true }).click();
  const dialog = page.getByRole("dialog", { name: "Leave with unsaved changes?" });
  await expect(dialog).toBeVisible();
  await dialog.getByRole("button", { name: "Stay" }).click();
  await expect(page.getByRole("navigation", { name: "Outline" }).getByRole("button")).toHaveCount(22);
  await page.getByRole("button", { name: "Workspace", exact: true }).click();
  await page
    .getByRole("dialog", { name: "Leave with unsaved changes?" })
    .getByRole("button", { name: "Discard and leave" })
    .click();
  await expect(page.getByRole("heading", { level: 1, name: "Workspace" })).toBeVisible();
});

test("a plant copied from a shipped example opens in the editor with its equipment", async ({ page }) => {
  await page.goto("/");
  await page.getByRole("button", { name: "New plant" }).click();
  const form = page.getByRole("form", { name: "New plant" });
  await form.getByLabel("Start from").selectOption({ label: "Chiller pilot" });
  await form.getByLabel("Name").fill("Pilot copy");
  await form.getByRole("button", { name: "Create and open the editor" }).click();
  await expect(page.getByRole("heading", { level: 1, name: "Pilot copy" })).toBeVisible();
  await expect(page.getByRole("navigation", { name: "Outline" }).getByRole("button")).not.toHaveCount(0);
});

test("Enter in the instance name adds the equipment, as the button does", async ({ page }) => {
  await page.goto("/");
  await page.getByRole("button", { name: "New plant" }).click();
  const form = page.getByRole("form", { name: "New plant" });
  await form.getByLabel("Name").fill("Enter probe");
  await form.getByLabel("Start from").selectOption({ label: "Factory utilities" });
  await form.getByRole("radio", { name: /^An empty plant/ }).check();
  await form.getByRole("button", { name: "Create and open the editor" }).click();
  await expect(page.getByRole("heading", { name: "No equipment yet" })).toBeVisible();
  const build = page.getByRole("complementary", { name: "Build" });
  await build.getByLabel("Find equipment").fill("pump");
  await build.getByLabel("Equipment", { exact: true }).selectOption({ label: "Centrifugal pump" });
  await build.getByLabel("Area", { exact: true }).selectOption({ index: 0 });
  await build.getByLabel("Name", { exact: true }).fill("pump-2");
  await build.getByLabel("Name", { exact: true }).press("Enter");
  await expect(page.getByRole("navigation", { name: "Outline" }).getByRole("button", { name: "pump-2" })).toBeVisible();
});

test("Enter in the equipment search only filters; a name another instance uses is refused in the add form", async ({
  page,
}) => {
  await page.goto("/");
  await page.getByRole("button", { name: "New plant" }).click();
  const form = page.getByRole("form", { name: "New plant" });
  await form.getByLabel("Name").fill("Search probe");
  await form.getByLabel("Start from").selectOption({ label: "Factory utilities" });
  await form.getByRole("radio", { name: /^An empty plant/ }).check();
  await form.getByRole("button", { name: "Create and open the editor" }).click();
  await expect(page.getByRole("heading", { name: "No equipment yet" })).toBeVisible();
  const build = page.getByRole("complementary", { name: "Build" });
  await build.getByLabel("Find equipment").fill("pump");
  await build.getByLabel("Equipment", { exact: true }).selectOption({ label: "Centrifugal pump" });
  await build.getByLabel("Area", { exact: true }).selectOption({ index: 0 });
  await build.getByLabel("Name", { exact: true }).fill("pump-1");
  await build.getByLabel("Find equipment").press("Enter");
  await expect(page.getByRole("heading", { name: "No equipment yet" })).toBeVisible();
  await expect(page.getByRole("navigation", { name: "Outline" }).getByRole("button")).toHaveCount(0);
  await build.getByRole("button", { name: "Add to the plant" }).click();
  await expect(page.getByRole("navigation", { name: "Outline" }).getByRole("button", { name: "pump-1" })).toBeVisible();
  await page.getByRole("button", { name: "Close the instance" }).click();
  await build.getByLabel("Find equipment").fill("pump");
  await build.getByLabel("Equipment", { exact: true }).selectOption({ label: "Centrifugal pump" });
  await build.getByLabel("Name", { exact: true }).fill("pump-1");
  await expect(build.getByText("pump-1 is already used by another instance.")).toBeVisible();
  await expect(build.getByRole("button", { name: "Add to the plant" })).toHaveAttribute("aria-disabled", "true");
});

test("a saved plant that has no domain yet is checked by Save, so its issues show without an edit", async ({
  page,
  lab,
}) => {
  const folder = lab.workspace;
  const file = join(folder, "compositions", "unchecked-plant.composition.json");
  const source = readFileSync(join(folder, "compositions", "factory-utilities.composition.json"), "utf8");
  writeFileSync(
    file,
    source
      .replace('"id": "factory-utilities"', '"id": "unchecked-plant"')
      .replace('"title": "Factory utilities"', '"title": "Unchecked plant"'),
  );
  await page.goto(`/#/edit?composition=${encodeURIComponent("compositions/unchecked-plant.composition.json")}`);
  await expect(page.getByRole("heading", { level: 1, name: "Unchecked plant" })).toBeVisible();
  await expect(page.getByRole("status").filter({ hasText: "Not checked yet" })).toBeVisible();
  await page.getByRole("button", { name: "Save" }).click();
  await expect(page.getByRole("status").filter({ hasText: /^Ready to run\. Saved\.$/ })).toBeVisible();
});

test("a full copy of a plant is runnable as soon as it is created", async ({ page }) => {
  await page.goto("/");
  await page.getByRole("button", { name: "New plant" }).click();
  const form = page.getByRole("form", { name: "New plant" });
  await form.getByLabel("Name").fill("Factory copy");
  await form.getByLabel("Start from").selectOption({ label: "Factory utilities" });
  await form.getByRole("radio", { name: /^A full copy/ }).check();
  await form.getByRole("button", { name: "Create and open the editor" }).click();
  await expect(page.getByRole("heading", { level: 1, name: "Factory copy" })).toBeVisible();
  await page.getByRole("button", { name: "Open plant" }).click();
  await expect(page.getByRole("button", { name: "Run plant" })).toBeVisible();
  await page.goto("/");
  const plants = page.getByRole("region", { name: "Plants" });
  await expect(plants.getByRole("listitem").filter({ hasText: "Factory copy" })).not.toContainText("Not runnable yet");
});

test("a long plant name is cut short in the top bar and never widens the page", async ({ page }) => {
  const name =
    'Hunt & "Ålborg" pumps / 冷却 test: 01 ünïcode ✓ with a very long name that keeps going to test the header wrapping';
  await page.goto("/");
  await page.getByRole("button", { name: "New plant" }).click();
  const form = page.getByRole("form", { name: "New plant" });
  await form.getByLabel("Name").fill(name);
  await form.getByLabel("Id").fill("long-plant-name");
  await form.getByLabel("Start from").selectOption({ label: "Factory utilities" });
  await form.getByRole("radio", { name: /^A full copy/ }).check();
  await form.getByRole("button", { name: "Create and open the editor" }).click();
  await expect(page.getByRole("heading", { level: 1, name })).toBeVisible();
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth)).toBe(true);
  await expect(
    page.getByRole("button", { name: "Theme: System" }).or(page.getByRole("button", { name: /Theme:/ })),
  ).toBeVisible();
  await expect(page.getByRole("button", { name: "Keyboard" })).toBeInViewport();
  const crumb = page.getByRole("navigation", { name: "Breadcrumb" }).getByRole("button", { name: "Workspace" });
  expect(await crumb.evaluate((el) => el.scrollWidth <= el.clientWidth)).toBe(true);
});
