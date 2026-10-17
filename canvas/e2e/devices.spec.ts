import AxeBuilder from "@axe-core/playwright";
import type { Page } from "@playwright/test";
import { expect, test } from "./support/test";
import { existsSync, readFileSync } from "node:fs";
import { join } from "node:path";

interface KindFile {
  readonly kind: string;
  readonly metrics: readonly { readonly name: string; readonly data_type: string }[];
  readonly commands: readonly { readonly name: string }[];
}

function outline(page: Page) {
  return page.getByRole("navigation", { name: "Device outline" });
}

function verdict(page: Page, text: string | RegExp) {
  return page.getByRole("status").filter({ hasText: text });
}

async function createDevice(page: Page, name: string): Promise<void> {
  await page.goto("/");
  const devices = page.getByRole("region", { name: "Devices" });
  await expect(devices.getByText(/No devices yet/)).toBeVisible();
  await devices.getByRole("button", { name: "Create device" }).click();
  const form = page.getByRole("form", { name: "New device" });
  await form.getByLabel("Name").fill(name);
  await form.getByRole("button", { name: "Create and open the device" }).click();
  await expect(page.getByRole("region", { name: `Editing ${name}` })).toBeVisible();
  await expect(page.getByRole("heading", { level: 1 })).toHaveAccessibleName(name);
  await expect(verdict(page, "Saved and valid.")).toBeVisible();
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
  await expect(page.getByText("The binary is likely to refuse this name.")).toBeVisible();
  await page.getByRole("button", { name: "Save", exact: true }).click();
  await expect(verdict(page, /Saved, but the binary refuses it/)).toBeVisible();
  await expect(name).toHaveAttribute("aria-invalid", "true");
  await expect(page.getByText(/does not match pattern/).first()).toBeVisible();
  await expect(page.getByRole("region", { name: "Issues" }).getByText("Reading bad name · Name")).toBeVisible();

  await name.fill("duty_permille");
  await page.getByRole("button", { name: "Save", exact: true }).click();
  await expect(verdict(page, "Saved and valid.")).toBeVisible();
  await expect(name).not.toHaveAttribute("aria-invalid", "true");
});
