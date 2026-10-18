import { step, verdict, writeKind } from "./support/devices";
import { expect, test } from "./support/test";

const ready = {
  kind: "e2e_ok",
  metrics: [
    { name: "level", data_type: "number", unit: "m", eu_range: { min: 0, max: 10 }, resolution: 0.1, writable: false },
  ],
  commands: [],
};

const unfinished = {
  kind: "e2e_unfinished",
  metrics: [{ name: "flow", data_type: "number", writable: false }],
  commands: [],
};

const fixed = {
  kind: "e2e_unfinished",
  metrics: [
    { name: "flow", data_type: "number", unit: "m", eu_range: { min: 0, max: 10 }, resolution: 0.1, writable: false },
  ],
  commands: [],
};

const rangeless = {
  kind: "e2e_rangeless",
  metrics: [{ name: "heat", data_type: "number", unit: "degC", resolution: 0.1, writable: false }],
  commands: [],
};

test("opening a device from the list puts the focus on its form, so the next Tab reaches a field", async ({
  page,
  lab,
}) => {
  writeKind(lab, "e2e_ok", ready);
  await page.goto("/");
  await page.getByRole("region", { name: "Devices" }).getByRole("button", { name: "Edit e2e_ok" }).click();
  await expect(page.getByRole("region", { name: "Editing e2e_ok" })).toBeVisible();
  await expect(page.getByRole("heading", { level: 2, name: "Device", exact: true })).toBeFocused();
  await page.keyboard.press("Tab");
  await expect(page.getByLabel("Name", { exact: true })).toBeFocused();
});

test("an unfinished device does not block another one, is named when it is added to a twin, and joins it once it is fixed, with no reload", async ({
  page,
  lab,
}) => {
  writeKind(lab, "e2e_ok", ready);
  writeKind(lab, "e2e_unfinished", unfinished);
  await page.goto("/");
  await page.getByRole("region", { name: "Devices" }).getByRole("button", { name: "Edit e2e_ok" }).click();
  await step(page, /^Behaviour/).click();
  await expect(page.getByRole("region", { name: "Behaviour of e2e_ok" })).toBeVisible();
  await expect(page.getByLabel("Twin name")).toHaveValue("e2e-ok-twin");
  await expect(page.getByRole("button", { name: "Create twin" })).toHaveAttribute("aria-disabled", "false");

  await page.getByRole("button", { name: "Add e2e_unfinished to the twin" }).click();
  const notice = page.getByRole("region", { name: "Devices that are not ready for a twin" });
  await expect(notice).toContainText("e2e_unfinished still needs");
  await expect(notice.getByRole("button", { name: "Open e2e_unfinished" })).toBeVisible();
  await expect(notice.getByRole("button", { name: "Retry" })).toBeVisible();
  await expect(page.getByRole("group", { name: "Reading flow" })).toHaveCount(0);
  await expect(page.getByRole("button", { name: "Create twin" })).toHaveAttribute("aria-disabled", "false");

  writeKind(lab, "e2e_unfinished", fixed);
  await expect(page.getByRole("group", { name: "Reading flow" })).toBeVisible({ timeout: 15_000 });
  await expect(notice).toHaveCount(0);
  await expect(page.getByRole("region", { name: "Behaviour of e2e_ok" })).toBeVisible();
});

test("an instrument that cannot read a reading is greyed out with one plain reason, and can be used once the reading states its range", async ({
  page,
  lab,
}) => {
  writeKind(lab, "e2e_rangeless", rangeless);
  await page.goto("/");
  await page.getByRole("region", { name: "Devices" }).getByRole("button", { name: "Edit e2e_rangeless" }).click();
  await step(page, /^Behaviour/).click();
  const heat = page.getByRole("group", { name: "Reading heat" });
  await heat.getByRole("button", { name: "Change the instrument of heat" }).click();
  await expect(heat.locator("input[type=radio]:disabled").first()).toBeVisible();
  await expect(heat.getByText("Set a range for this reading in the Interface step").first()).toBeVisible();
  await expect(heat.getByText(/does not fit the reading/)).toHaveCount(0);

  await heat.getByRole("button", { name: "Set it in the Interface step" }).click();
  await expect(page.getByRole("region", { name: "Editing e2e_rangeless" })).toBeVisible();
  await expect(page.getByLabel("Minimum")).toBeFocused();
  await page.getByLabel("Minimum").fill("0");
  await page.getByLabel("Maximum").fill("100");
  await page.getByRole("button", { name: "Save", exact: true }).click();
  await expect(verdict(page, "Saved and valid.")).toBeVisible();

  await step(page, /^Behaviour/).click();
  await heat.getByRole("button", { name: "Change the instrument of heat" }).click();
  await expect(heat.getByRole("radio", { name: /thermistor/i }).first()).toBeEnabled();
  await expect(heat.getByText("Set a range for this reading in the Interface step")).toHaveCount(0);
});
