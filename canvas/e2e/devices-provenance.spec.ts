import AxeBuilder from "@axe-core/playwright";
import type { Page } from "@playwright/test";
import { step, verdict, writeKind } from "./support/devices";
import { expect, test } from "./support/test";

const ready = {
  kind: "e2e_ok",
  metrics: [
    { name: "level", data_type: "number", unit: "m", eu_range: { min: 0, max: 10 }, resolution: 0.1, writable: false },
  ],
  commands: [],
};

const indicator = {
  kind: "e2e_indicator",
  metrics: [
    {
      name: "level_percent",
      data_type: "number",
      unit: "%",
      eu_range: { min: 0, max: 100 },
      resolution: 1,
      scan_class: "fast",
      writable: false,
    },
  ],
  commands: [],
};

const setting = "The setting scan_class of reading 'level_percent' is not reproduced by a twin.";

async function openBehaviour(page: Page): Promise<void> {
  await page.goto("/");
  await page.getByRole("region", { name: "Devices", exact: true }).getByRole("button", { name: "Edit e2e_ok" }).click();
  await step(page, /^Behaviour/).click();
  await expect(page.getByRole("region", { name: "Behaviour of e2e_ok" })).toBeVisible();
}

test("a device with a setting a twin does not reproduce is named, left out, and the partial twin says so", async ({
  page,
  lab,
}) => {
  writeKind(lab, "e2e_ok", ready);
  writeKind(lab, "e2e_indicator", indicator);
  await openBehaviour(page);
  await expect(page.getByRole("region", { name: "Suggested choices" })).toContainText(
    "1 choice is still the suggested default",
  );

  await page.getByRole("button", { name: "Add e2e_indicator to the twin" }).click();
  const notice = page.getByRole("region", { name: "Devices that are not ready for a twin" });
  await expect(notice.getByText(setting)).toBeVisible();
  await expect(notice).toContainText("e2e_indicator is left out of the twin as a whole");
  await notice.getByText("Details").click();
  await expect(notice.getByText(/is not part of the interface a twin reproduces/)).toBeVisible();

  const scope = page.getByRole("region", { name: "What goes into this twin" });
  await expect(scope).toContainText("e2e_ok");
  await expect(scope).toContainText("e2e_indicator: the setting scan_class of reading 'level_percent'");
  expect((await new AxeBuilder({ page }).analyze()).violations).toEqual([]);

  await page.getByRole("button", { name: "Create twin" }).click();
  await expect(verdict(page, /^Partial twin created\. This twin excludes e2e_indicator/)).toBeVisible();

  await step(page, /^Check/).click();
  const check = page.getByRole("region", { name: "Check of e2e_ok" });
  await expect(check.getByText("Partial twin", { exact: true })).toBeVisible();
  await expect(check.getByText("Interface: the twin matches its device")).toBeVisible();
  await expect(check.getByText("Plausibility: readings stay plausible in a nominal day")).toBeVisible({
    timeout: 15_000,
  });
  await expect(check.getByText("Calibration: not performed")).toBeVisible();
  await expect(check.getByText("Hardware validation: not run")).toBeVisible();
  expect((await new AxeBuilder({ page }).analyze()).violations).toEqual([]);
});

test("the suggested defaults are labelled until the person keeps or changes them", async ({ page, lab }) => {
  writeKind(lab, "e2e_ok", ready);
  await openBehaviour(page);
  const level = page.getByRole("group", { name: "Reading level" });
  await expect(level.getByText("Suggested", { exact: true })).toBeVisible();
  await page.getByRole("button", { name: "Keep the suggestions" }).click();
  await expect(page.getByRole("region", { name: "Suggested choices" })).toHaveCount(0);
  await expect(level.getByText("Your choice", { exact: true })).toBeVisible();
  await expect(page.getByRole("button", { name: "Create twin" })).toHaveAttribute("aria-disabled", "false");
  const details = page.getByRole("complementary", { name: "Twin details" });
  await expect(details.getByText("What this twin reproduces")).toBeVisible();
  await expect(details.getByText(/simple stand-in until it is compared with recorded field data/)).toBeVisible();
});
