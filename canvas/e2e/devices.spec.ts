import AxeBuilder from "@axe-core/playwright";
import type { Page } from "@playwright/test";
import { createDevice, outline, step, verdict } from "./support/devices";
import { expect, test } from "./support/test";
import { existsSync, readFileSync } from "node:fs";
import { join } from "node:path";

interface KindFile {
  readonly kind: string;
  readonly metrics: readonly { readonly name: string; readonly data_type: string }[];
  readonly commands: readonly { readonly name: string }[];
}

async function addReadingAndCommand(page: Page): Promise<void> {
  await page.getByRole("button", { name: "Add a reading" }).click();
  await page.getByLabel("Name", { exact: true }).fill("duty_permille");
  await page.getByLabel("Unit", { exact: true }).fill("1");
  await page.getByRole("button", { name: "Add a command" }).click();
  await page.getByLabel("Name", { exact: true }).fill("set_duty");
  await page.getByLabel("Risk class").selectOption({ label: "Medium" });
}

test("a device with one reading and one command is created, saved, listed after a reload, and a bad value shows the binary's issue on its field", async ({
  page,
  lab,
}) => {
  await createDevice(page, "e2e-fan");
  await addReadingAndCommand(page);
  await expect(verdict(page, "Unsaved changes")).toBeVisible();
  await page.getByRole("button", { name: "Save", exact: true }).click();
  await expect(verdict(page, "Saved and valid.")).toBeVisible();
  await expect(verdict(page, "Saved and valid.")).toHaveAttribute("title", "1 reading · 1 command");
  const file = join(lab.workspace, "kinds", "e2e-fan.kind.json");
  expect(existsSync(file)).toBe(true);
  const written = JSON.parse(readFileSync(file, "utf8")) as KindFile;
  expect(written).toMatchObject({
    kind: "e2e-fan",
    metrics: [{ name: "duty_permille", data_type: "number" }],
    commands: [{ name: "set_duty" }],
  });
  expect((await new AxeBuilder({ page }).analyze()).violations).toEqual([]);

  await page.reload();
  await expect(outline(page).getByRole("button", { name: /^duty_permille/ })).toBeVisible();
  await page.getByRole("button", { name: "Streams Simulator, workspace home" }).click();
  const devices = page.getByRole("region", { name: "Devices" });
  await expect(devices.getByText("1 reading · 1 command")).toBeVisible();
  await devices.getByRole("button", { name: "Edit e2e-fan" }).click();

  await outline(page)
    .getByRole("button", { name: /^duty_permille/ })
    .click();
  const name = page.getByLabel("Name", { exact: true });
  await name.fill("bad name");
  await expect(name).toHaveAttribute("aria-invalid", "true");
  await page.getByRole("button", { name: "Save", exact: true }).click();
  await expect(verdict(page, /Not saved: the binary refuses it/)).toBeVisible();
  await expect(page.getByRole("region", { name: "Issues" }).getByText("Reading bad name · Name")).toBeVisible();
  const onDisk = JSON.parse(readFileSync(file, "utf8")) as KindFile;
  expect(onDisk.metrics.map((metric) => metric.name)).toEqual(["duty_permille"]);

  await name.fill("duty_permille");
  await expect(name).not.toHaveAttribute("aria-invalid", "true");
  await expect(verdict(page, "Saved and valid.")).toBeVisible();
  await expect(page.getByRole("button", { name: "Save", exact: true })).toHaveAttribute("aria-disabled", "true");
});

async function addNumberReading(page: Page, name: string, unit: string): Promise<void> {
  await page.getByRole("button", { name: "Add a reading" }).click();
  await page.getByLabel("Name", { exact: true }).fill(name);
  await page.getByLabel("Unit", { exact: true }).fill(unit);
  await page.getByLabel("Minimum").fill("0");
  await page.getByLabel("Maximum").fill("100");
}

async function defineCooler(page: Page): Promise<void> {
  await createDevice(page, "e2e_cooler");
  await addNumberReading(page, "temperature", "degC");
  await addNumberReading(page, "duty", "%");
  await page.getByLabel("Resolution", { exact: true }).fill("0.1");
  await page.getByLabel("Setpoint").check();
  await page.getByRole("button", { name: "Add a command" }).click();
  await page.getByLabel("Name", { exact: true }).fill("set_duty");
  await page.getByLabel("Risk class").selectOption({ label: "Medium" });
  await page.getByRole("button", { name: "Add argument" }).click();
  await page.getByLabel("Argument 1 name").fill("duty");
  await page.getByLabel("Argument 1 minimum").fill("0");
  await page.getByLabel("Argument 1 maximum").fill("100");
  await page.getByRole("button", { name: "Save", exact: true }).click();
  await expect(verdict(page, "Saved and valid.")).toBeVisible();
}

async function chooseThermistor(page: Page): Promise<void> {
  const temperature = page.getByRole("group", { name: "Reading temperature" });
  await temperature.getByRole("button", { name: "Change the instrument of temperature" }).click();
  await temperature.getByLabel("Find an instrument for temperature").fill("thermistor");
  await temperature
    .getByRole("radio", { name: /thermistor/i })
    .first()
    .check();
  await temperature.getByRole("button", { name: "Use this instrument" }).click();
  await expect(temperature.getByLabel("Instrument", { exact: true })).toHaveValue(/thermistor/i);
  await temperature.getByLabel("Starts at").fill("30");
}

async function chooseGradualResponse(page: Page, timeConstant: string): Promise<void> {
  const command = page.getByRole("group", { name: "Command set_duty" });
  await command.getByLabel("Response").selectOption({ label: "Sets a reading, and another reading follows gradually" });
  const effect = command.getByRole("group", { name: "Effect 1 of set_duty" });
  await effect.getByLabel("Reading that follows").selectOption({ label: "temperature" });
  await effect.getByLabel("Gain").fill("-0.1");
  await effect.getByLabel("Time constant").fill(timeConstant);
}

test("a device gets a twin: what it still needs is found before the behaviour, the instrument and a gradual response are chosen, the twin is created, and it is found created after a reload", async ({
  page,
  lab,
}) => {
  await defineCooler(page);
  await expect(page.getByText("Needed for a twin: 1")).toBeVisible();
  await expect(step(page, /^Interface/)).toContainText("Valid");
  await step(page, /^Behaviour/).click();
  const first = page.getByRole("region", { name: "What the twin needs first" });
  await expect(first).toBeVisible();
  await expect(page.getByRole("button", { name: "Create twin" })).toHaveCount(0);
  await first.getByRole("button", { name: /temperature/ }).click();
  await expect(page.getByRole("region", { name: "Editing e2e_cooler" })).toBeVisible();
  await expect(page.getByLabel("Resolution", { exact: true })).toBeFocused();
  await page.getByLabel("Resolution", { exact: true }).fill("0.1");
  await expect(verdict(page, "Unsaved changes")).toBeVisible();
  await page.getByRole("button", { name: "Save", exact: true }).click();
  await expect(verdict(page, "Saved and valid.")).toBeVisible();
  await expect(page.getByText("Ready for a twin")).toBeVisible();

  await step(page, /^Behaviour/).click();
  await expect(page.getByRole("region", { name: "Behaviour of e2e_cooler" })).toBeVisible();
  await expect(page.getByRole("heading", { level: 2, name: "Twin", exact: true })).toBeFocused();
  await expect(page.getByLabel("Twin name")).toHaveValue("e2e-cooler-twin");
  await chooseThermistor(page);
  await chooseGradualResponse(page, "-5");
  const effect = page.getByRole("group", { name: "Effect 1 of set_duty" });
  await expect(effect.getByText("Must be above zero")).toBeVisible();
  await expect(effect.getByLabel("Time constant")).toHaveAttribute("aria-invalid", "true");
  await expect(page.getByRole("button", { name: "Create twin" })).toHaveAttribute("aria-disabled", "true");
  await expect(page.getByRole("button", { name: "Create twin" })).toHaveAccessibleDescription(/Finish the choices/);
  await expect(effect.getByText("Choose a time constant above zero.")).toBeVisible();
  await expect(step(page, /^Behaviour/)).toContainText("Needs choices");
  await effect.getByLabel("Time constant").fill("180");
  await expect(effect.getByText("Must be above zero")).toHaveCount(0);
  await expect(step(page, /^Behaviour/)).toContainText("Ready");
  expect((await new AxeBuilder({ page }).analyze()).violations).toEqual([]);

  await page.getByRole("button", { name: "Create twin" }).click();
  await expect(verdict(page, "Twin ready.")).toBeVisible();
  await expect(step(page, /^Check/)).toContainText("Passed");
  expect(existsSync(join(lab.workspace, "domains", "e2e-cooler-twin.domain.json"))).toBe(true);
  const choices = JSON.parse(readFileSync(join(lab.workspace, "twins", "e2e_cooler.twin.json"), "utf8")) as unknown;
  expect(choices).toMatchObject({ id: "e2e-cooler-twin", kinds: [{ kind: "e2e_cooler" }] });

  await step(page, /^Check/).click();
  await expect(page.getByRole("heading", { level: 2, name: "Check", exact: true })).toBeFocused();
  const check = page.getByRole("region", { name: "Check of e2e_cooler" });
  await expect(check.getByText("The twin matches its device")).toBeVisible();
  await check.getByText("Technical details").click();
  await expect(check.getByText("domains/e2e-cooler-twin.domain.json")).toBeVisible();
  expect((await new AxeBuilder({ page }).analyze()).violations).toEqual([]);

  await page.reload();
  await expect(step(page, /^Behaviour/)).toContainText("Created");
  await step(page, /^Behaviour/).click();
  await expect(page.getByLabel("Twin name")).toHaveValue("e2e-cooler-twin");
  const temperature = page.getByRole("group", { name: "Reading temperature" });
  await expect(temperature.getByLabel("Instrument", { exact: true })).toHaveValue(/thermistor/i);
  await expect(temperature.getByLabel("Starts at")).toHaveValue("30");
  const again = page.getByRole("group", { name: "Effect 1 of set_duty" });
  await expect(again.getByLabel("Time constant")).toHaveValue("180");
  await expect(again.getByLabel("Gain")).toHaveValue("-0.1");
  await expect(verdict(page, "Twin ready.")).toBeVisible();
  await expect(page.getByRole("button", { name: "Create twin" })).toHaveCount(0);

  await page.getByRole("button", { name: "Replace twin" }).click();
  await expect(verdict(page, "Twin ready.")).toBeVisible();
  await page.getByRole("button", { name: "Open as a plant" }).click();
  await expect(page).toHaveURL(/#\/plant\?domain=domains%2Fe2e-cooler-twin\.domain\.json/);
});
