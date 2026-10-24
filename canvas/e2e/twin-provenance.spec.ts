import AxeBuilder from "@axe-core/playwright";
import { addThermalTwin, recordTwinRun, strengthenFan } from "./support/thermalTwin";
import { expect, test } from "./support/test";

test.setTimeout(60_000);

test("a device included in a twin says so, and the twin's plant says why it is read-only and opens the Behaviour step", async ({
  page,
  lab,
}) => {
  addThermalTwin(lab);
  await page.goto("/");
  const devices = page.getByRole("region", { name: "Devices", exact: true });
  await expect(devices.getByText("Included in twin arduino-thermal-twin")).toHaveCount(2);
  await expect(devices.getByText("Needs a twin")).toHaveCount(0);

  const plants = page.getByRole("region", { name: "Plants", exact: true });
  await expect(plants.getByText("Read-only: this plant is generated from a twin.")).toBeVisible();
  await plants.getByRole("button", { name: "Open Arduino thermal bench twin, scaffolded" }).click();
  const notice = page.getByRole("region", { name: "Why this plant cannot be edited here" });
  await expect(notice.getByText("Read-only: this plant is generated from a twin.")).toBeVisible();
  await page.getByRole("button", { name: "Lineage" }).click();
  const why = page.getByRole("region", { name: "Why this plant is read-only" });
  await expect(why.getByText(/Generated from the twin choices twins\/fan\.twin\.json/)).toBeVisible();
  await expect(why.getByText(/its links are named driver_drives_driven and carry no ports/)).toBeVisible();
  expect((await new AxeBuilder({ page }).analyze()).violations).toEqual([]);

  await notice.getByRole("button", { name: "Open the twin's Behaviour step" }).click();
  await expect(page.getByRole("region", { name: "Behaviour of fan" })).toBeVisible();
});

test("changing the gain of a twin after a run is shown in the run's review as a change of the model that kept its version", async ({
  page,
  lab,
}) => {
  addThermalTwin(lab);
  recordTwinRun(lab, "w-before");
  strengthenFan(lab, -0.05);
  await page.goto("/#/review?world=w-before");
  await page.getByRole("button", { name: "Compare model with the current plant" }).click();
  const dialog = page.getByRole("dialog", { name: "What changed in the model" });
  await expect(dialog.getByRole("alert")).toHaveText("The model changed but its version did not.");
  const group = dialog.getByRole("region", { name: "Changes to thermal_sensor" });
  await expect(group.getByText("gain changed")).toBeVisible();
  await expect(group.getByText("Before: -0.02")).toBeVisible();
  await expect(group.getByText("After: -0.05")).toBeVisible();
  await expect(group.getByText("Affects zone-01")).toBeVisible();
  expect((await new AxeBuilder({ page }).analyze()).violations).toEqual([]);
});
