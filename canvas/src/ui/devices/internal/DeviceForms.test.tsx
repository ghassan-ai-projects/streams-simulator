import { fireEvent, screen, within } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import type { Draft } from "../../../app/devices";
import { deviceBridge, editingDevice, fan, valid } from "../../../test/deviceHarness";

function choose(name: RegExp) {
  fireEvent.click(within(screen.getByRole("navigation", { name: "Device outline" })).getByRole("button", { name }));
}

function draftOf(app: Awaited<ReturnType<typeof editingDevice>>): Draft {
  const state = app.stores.devices.snapshot();
  if (state.phase !== "ready") throw new Error(state.phase);
  return state;
}

function change(label: string, value: string) {
  fireEvent.change(screen.getByLabelText(label), { target: { value } });
}

describe("the device name", () => {
  it("renames the device, leaves the pattern to the binary, and names a blank one", async () => {
    const app = await editingDevice();
    change("Name", "cooling fan");
    expect(draftOf(app).doc.kind).toBe("cooling fan");
    expect(screen.queryByText(/likely to refuse/)).toBeNull();
    await vi.waitFor(() => {
      expect(screen.getByLabelText("Name")).toHaveAccessibleDescription(/Use letters, digits and \. _ : - only/);
    });
    expect(screen.getByLabelText("Name")).toHaveAttribute("aria-invalid", "true");
    change("Name", "");
    expect(screen.getByLabelText("Name")).toHaveAttribute("aria-invalid", "true");
    expect(screen.getAllByText("Give the device a name.").length).toBeGreaterThan(0);
    expect(screen.getByLabelText("Name")).toHaveFocus();
  });
});

describe("a reading", () => {
  it("changes its type, unit, description, scan class and setpoint flag", async () => {
    const app = await editingDevice();
    choose(/^duty/);
    change("Type", "integer");
    change("Unit", "rpm");
    change("Scan class", "fast");
    change("Description", "Fan duty");
    fireEvent.click(screen.getByLabelText("Setpoint"));
    expect(draftOf(app).doc.metrics[0]).toMatchObject({
      dataType: "integer",
      unit: "rpm",
      scanClass: "fast",
      description: "Fan duty",
      writable: false,
    });
    expect(document.querySelectorAll("datalist option").length).toBeGreaterThan(5);
  });

  it("offers a range and a resolution for numbers only", async () => {
    const fake = deviceBridge();
    fake.handlers.set("kind.check", () => valid);
    await editingDevice(fake);
    choose(/^mode/);
    expect(screen.queryByLabelText("Resolution")).toBeNull();
    expect(screen.queryByLabelText("Minimum")).toBeNull();
    change("Type", "number");
    expect(screen.getByLabelText("Resolution")).toBeInTheDocument();
    change("Type", "boolean");
    expect(screen.queryByLabelText("Resolution")).toBeNull();
  });

  it("shows a resolution a twin needs even on a reading that is not a number", async () => {
    await editingDevice();
    choose(/^mode/);
    expect(screen.getByLabelText("Resolution")).toBeInTheDocument();
    expect(screen.queryByLabelText("Minimum")).toBeNull();
  });

  it("takes numbers as typed, refuses text in a number field, and warns when a range runs backwards", async () => {
    const app = await editingDevice();
    choose(/^duty/);
    change("Minimum", "abc");
    expect(screen.getByText("Enter a number.")).toBeInTheDocument();
    expect(screen.getByLabelText("Minimum")).toHaveAttribute("aria-invalid", "true");
    expect(draftOf(app).doc.metrics[0]?.range.min).toBe(0);
    fireEvent.blur(screen.getByLabelText("Minimum"));
    expect(screen.getByLabelText("Minimum")).toHaveValue("abc");
    change("Minimum", "150");
    expect(screen.queryByText("Enter a number.")).toBeNull();
    expect(screen.getByText("The minimum is above the maximum.")).toBeInTheDocument();
    fireEvent.blur(screen.getByLabelText("Minimum"));
    change("Maximum", "");
    change("Resolution", "0.5");
    expect(draftOf(app).doc.metrics[0]).toMatchObject({ range: { min: 150, max: null }, resolution: 0.5 });
    change("Resolution", "");
    expect(draftOf(app).doc.metrics[0]?.resolution).toBeNull();
  });

  it("lists the options of a list reading as typed, and keeps them as the file's list", async () => {
    const app = await editingDevice();
    choose(/^mode/);
    expect(screen.getByLabelText("Options")).toHaveValue("auto, manual");
    change("Options", "auto, manual, off,");
    expect(screen.getByLabelText("Options")).toHaveValue("auto, manual, off,");
    expect(draftOf(app).doc.metrics[1]?.extra).toEqual({ enum_values: ["auto", "manual", "off"] });
  });

  it("shows a type the binary does not accept, and asks for one when there is none", async () => {
    await editingDevice(
      deviceBridge({ kind: "k", metrics: [{ name: "a", data_type: "weird" }, { name: "b" }], commands: [] }),
    );
    choose(/^a/);
    expect(
      within(screen.getByLabelText("Type")).getByRole("option", { name: "weird (not accepted)" }),
    ).toBeInTheDocument();
    choose(/^b/);
    expect(within(screen.getByLabelText("Type")).getByRole("option", { name: "Choose…" })).toBeInTheDocument();
  });
});

describe("a command", () => {
  it("changes its risk class, safe stop flag, name and description", async () => {
    const app = await editingDevice();
    choose(/^stop/);
    change("Risk class", "critical");
    fireEvent.click(screen.getByLabelText("Safe stop"));
    change("Name", "halt");
    change("Description", "Stop it");
    expect(draftOf(app).doc.commands[1]).toMatchObject({
      name: "halt",
      riskClass: "critical",
      safeStop: false,
      description: "Stop it",
    });
  });

  it("adds, edits and removes arguments", async () => {
    const app = await editingDevice();
    choose(/^stop/);
    fireEvent.click(screen.getByRole("button", { name: "Add argument" }));
    fireEvent.change(screen.getByLabelText("Argument 1 name"), { target: { value: "lease_ms" } });
    fireEvent.change(screen.getByLabelText("Argument 1 type"), { target: { value: "integer" } });
    fireEvent.change(screen.getByLabelText("Argument 1 minimum"), { target: { value: "5000" } });
    fireEvent.change(screen.getByLabelText("Argument 1 maximum"), { target: { value: "10000" } });
    fireEvent.click(screen.getByRole("checkbox", { name: "Required" }));
    expect(draftOf(app).doc.commands[1]?.args).toEqual({
      kind: "rows",
      rows: [{ name: "lease_ms", type: "integer", min: 5000, max: 10000, required: true }],
    });
    fireEvent.click(screen.getByRole("button", { name: "Remove argument 1" }));
    expect(screen.queryByLabelText("Argument 1 name")).toBeNull();
  });

  it("says what is wrong with an unnamed argument", async () => {
    await editingDevice();
    choose(/^stop/);
    fireEvent.click(screen.getByRole("button", { name: "Add argument" }));
    expect(screen.getAllByText("Give every argument a name.").length).toBeGreaterThan(0);
    expect(screen.getByRole("button", { name: "Save" })).toHaveAttribute("aria-disabled", "true");
  });

  it("keeps arguments it cannot show as they are, and says so", async () => {
    const custom = {
      kind: "k",
      metrics: [],
      commands: [{ name: "go", args_schema: { oneOf: [{ type: "object" }] } }],
    };
    const app = await editingDevice(deviceBridge(custom));
    choose(/^go/);
    expect(screen.getByText(/arguments use a definition this editor cannot show/)).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Add argument" })).toBeNull();
    expect(draftOf(app).doc.commands[0]?.args).toMatchObject({ kind: "custom" });
  });
});

describe("the outline", () => {
  it("says what to do when the device has no readings or commands yet", async () => {
    await editingDevice(deviceBridge({ kind: "empty", metrics: [], commands: [] }));
    expect(screen.getByText(/No readings yet/)).toBeInTheDocument();
    expect(screen.getByText(/No commands yet/)).toBeInTheDocument();
    expect(screen.getByText(/Choose one in the outline/)).toBeInTheDocument();
  });

  it("keeps the file's other fields when it saves", async () => {
    const app = await editingDevice(deviceBridge({ ...fan, title: "Cooling fan" }));
    change("Name", "fan2");
    fireEvent.click(screen.getByRole("button", { name: "Save" }));
    await screen.findByText("Saved and valid.");
    expect(JSON.parse(app.fake.writes[0]?.content ?? "{}")).toMatchObject({ kind: "fan2", title: "Cooling fan" });
  });
});
