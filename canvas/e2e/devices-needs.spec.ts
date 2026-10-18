import AxeBuilder from "@axe-core/playwright";
import { existsSync, readFileSync, writeFileSync } from "node:fs";
import { createDevice, kindPath, outline, step, verdict, writeKind } from "./support/devices";
import { expect, test } from "./support/test";

const needy = {
  kind: "e2e_needs",
  metrics: [
    { name: "flow_a", data_type: "number", unit: "m", writable: false },
    { name: "flow_b", data_type: "number", unit: "m", writable: false },
    { name: "flow_c", data_type: "number", resolution: 1, writable: false },
  ],
  commands: [],
};

test("everything a twin still needs is found at once in the interface step, before the behaviour is opened", async ({
  page,
  lab,
}) => {
  writeKind(lab, "e2e_needs", needy);
  await page.goto("/");
  await page.getByRole("region", { name: "Devices" }).getByRole("button", { name: "Edit e2e_needs" }).click();
  await expect(page.getByRole("heading", { level: 1, name: "e2e_needs" })).toBeFocused();
  await expect(page.getByText("Needed for a twin: 3")).toBeVisible();
  await expect(step(page, /^Interface/)).toContainText("Valid");
  const needs = page.getByRole("region", { name: "Needed for a twin" });
  await expect(needs.getByRole("button")).toHaveCount(3);
  await expect(needs.getByRole("button", { name: /Reading flow_c · Unit/ })).toBeVisible();
  await outline(page)
    .getByRole("button", { name: /^flow_a/ })
    .click();
  await expect(page.getByLabel("Resolution", { exact: true })).toHaveAccessibleDescription(/Needs a resolution/);
  await expect(page.getByLabel("Resolution", { exact: true })).not.toHaveAttribute("aria-invalid", "true");
  expect((await new AxeBuilder({ page }).analyze()).violations).toEqual([]);

  await step(page, /^Behaviour/).click();
  const first = page.getByRole("region", { name: "What the twin needs first" });
  await expect(first.getByRole("button")).toHaveCount(3);
  await expect(page.getByRole("heading", { level: 2, name: "Before the behaviour" })).toBeFocused();
  await first.getByRole("button", { name: /flow_c needs a unit/ }).click();
  await expect(page.getByRole("region", { name: "Editing e2e_needs" })).toBeVisible();
  await expect(page.getByLabel("Unit", { exact: true })).toBeFocused();
  await expect(page.getByLabel("Name", { exact: true })).toHaveValue("flow_c");
});

test("leaving with a definition the binary refuses explains why it was not saved, keeps the page and leaves the file as it was", async ({
  page,
  lab,
}) => {
  await createDevice(page, "e2e-leave");
  await page.getByRole("button", { name: "Add a reading" }).click();
  await page.getByLabel("Name", { exact: true }).fill("ok_reading");
  await page.getByRole("button", { name: "Save", exact: true }).click();
  await expect(verdict(page, "Saved and valid.")).toBeVisible();
  await page.getByLabel("Name", { exact: true }).fill("bad name");
  const home = page.getByRole("button", { name: "Streams Simulator, workspace home" });
  await home.click();
  const dialog = page.getByRole("dialog", { name: "Leave with unsaved changes?" });
  await expect(dialog).toHaveAttribute("aria-modal", "true");
  await dialog.getByRole("button", { name: "Save and leave" }).click();
  await expect(dialog.getByRole("alert")).toContainText("Not saved: the binary refuses it");
  await expect(page.getByRole("region", { name: "Editing e2e-leave" })).toBeVisible();
  const saved = JSON.parse(readFileSync(kindPath(lab, "e2e-leave"), "utf8")) as { metrics: { name: string }[] };
  expect(saved.metrics.map((metric) => metric.name)).toEqual(["ok_reading"]);
  await dialog.getByRole("button", { name: "Save and leave" }).focus();
  await page.keyboard.press("Tab");
  await expect(dialog.getByRole("button", { name: "Stay" })).toBeFocused();
  await dialog.getByRole("button", { name: "Stay" }).click();
  await expect(dialog).toHaveCount(0);
  await expect(home).toBeFocused();
});

test("reloading from disk after a conflict asks before it throws the edits away", async ({ page, lab }) => {
  await createDevice(page, "e2e-conflict");
  await page.getByRole("button", { name: "Add a reading" }).click();
  await page.getByLabel("Name", { exact: true }).fill("pump_speed");
  await page.getByRole("button", { name: "Save", exact: true }).click();
  await expect(verdict(page, "Saved and valid.")).toBeVisible();
  expect(existsSync(kindPath(lab, "e2e-conflict"))).toBe(true);
  writeFileSync(
    kindPath(lab, "e2e-conflict"),
    JSON.stringify({
      kind: "e2e-conflict",
      metrics: [{ name: "external_speed", data_type: "number", writable: false }],
      commands: [],
    }),
  );
  await page.getByLabel("Unit", { exact: true }).fill("rpm");
  await page.getByRole("button", { name: "Save", exact: true }).click();
  await expect(verdict(page, /the file changed on disk/)).toBeVisible();
  await page.getByRole("button", { name: "Reload from disk" }).click();
  const question = page.getByRole("group", { name: "Reload and lose your changes?" });
  await expect(question.getByRole("button", { name: "Keep editing" })).toBeFocused();
  await question.getByRole("button", { name: "Keep editing" }).click();
  await expect(page.getByLabel("Unit", { exact: true })).toHaveValue("rpm");
  await page.getByRole("button", { name: "Reload from disk" }).click();
  await page.getByRole("button", { name: "Reload and lose them" }).click();
  await expect(outline(page).getByRole("button", { name: /^external_speed/ })).toBeVisible();
  await expect(outline(page).getByRole("button", { name: /^pump_speed/ })).toHaveCount(0);
});
