import { act, fireEvent, screen, within } from "@testing-library/react";
import axe from "axe-core";
import { describe, expect, it, vi } from "vitest";
import { mountApp } from "../../../test/appHarness";
import { deviceBridge, editingDevice, fan, hash, path, refusal, valid } from "../../../test/deviceHarness";
import { cannedFailure } from "../../../test/fakeBridge";
import { failure as failed } from "../../../test/outcome";

function outline() {
  return within(screen.getByRole("navigation", { name: "Device outline" }));
}

function choose(name: RegExp) {
  fireEvent.click(outline().getByRole("button", { name }));
}

function shortcut(key: string, target: Element = document.body) {
  return fireEvent.keyDown(target, { key, ctrlKey: true });
}

describe("the device editor screen", () => {
  it("lists the readings and commands in the outline and opens on the device, with no accessibility violation", async () => {
    const { view } = await editingDevice();
    expect(screen.getByRole("heading", { level: 1, name: "fan" })).toHaveAttribute("aria-current", "location");
    for (const name of [/^fan/, /^duty/, /^mode/, /^set_duty/, /^stop/]) {
      expect(outline().getByRole("button", { name })).toBeInTheDocument();
    }
    expect(outline().getByRole("button", { name: /^stop/ })).toHaveTextContent("Low risk · safe stop");
    expect(screen.getByRole("heading", { name: "Device" })).toBeInTheDocument();
    expect(screen.getByLabelText("Name")).toHaveValue("fan");
    expect(await screen.findByText("Saved and valid.")).toHaveAttribute("title", "2 readings · 2 commands");
    expect((await axe.run(view.container)).violations).toEqual([]);
  });

  it("keeps the file path and digest out of the headline, in a details disclosure", async () => {
    await editingDevice();
    const details = screen.getByText("Technical details").closest("details");
    expect(details).not.toHaveAttribute("open");
    expect(within(details as HTMLElement).getByText(path)).toBeInTheDocument();
    expect(within(details as HTMLElement).getByText("sha256:abc")).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "Device details" })).toBeInTheDocument();
  });

  it("waits with a status while the file is read", async () => {
    const fake = deviceBridge();
    fake.holdRead(path, 1);
    mountApp(fake, hash);
    expect(await screen.findByText("Opening the device…")).toBeInTheDocument();
  });

  it.each([
    [
      "is missing",
      (fake: ReturnType<typeof deviceBridge>) => {
        fake.files.delete(path);
      },
    ],
    [
      "is not a device definition",
      (fake: ReturnType<typeof deviceBridge>) => {
        fake.addFile(path, "[1]");
      },
    ],
  ])("says why a device cannot be opened when its file %s", async (_name, spoil) => {
    const fake = deviceBridge();
    spoil(fake);
    mountApp(fake, hash);
    expect(await screen.findByRole("alert")).toHaveTextContent("The device could not be opened:");
  });

  it("shows a reading's fields with its range and unit, and a command's with its arguments", async () => {
    await editingDevice();
    choose(/^duty/);
    expect(screen.getByRole("heading", { name: "Reading" })).toBeInTheDocument();
    expect(screen.getByLabelText("Name")).toHaveValue("duty");
    expect(screen.getByLabelText("Type")).toHaveValue("number");
    expect(screen.getByLabelText("Unit")).toHaveValue("%");
    expect(screen.getByLabelText("Minimum")).toHaveValue("0");
    expect(screen.getByLabelText("Maximum")).toHaveValue("100");
    expect(screen.getByLabelText("Setpoint")).toBeChecked();
    choose(/^set_duty/);
    expect(screen.getByRole("heading", { name: "Command" })).toBeInTheDocument();
    expect(screen.getByLabelText("Risk class")).toHaveValue("medium");
    expect(screen.getByLabelText("Argument 1 name")).toHaveValue("duty");
    expect(screen.getByLabelText("Argument 1 type")).toHaveValue("number");
    expect(screen.getByLabelText("Argument 1 minimum")).toHaveValue("0");
    expect(screen.getByLabelText("Argument 1 maximum")).toHaveValue("100");
    expect(screen.getByRole("checkbox", { name: "Required" })).toBeChecked();
    choose(/^stop/);
    expect(screen.getByLabelText("Safe stop")).toBeChecked();
  });

  it("marks the device unsaved on an edit and saves it over the version it read, then shows the verdict", async () => {
    const { fake } = await editingDevice();
    choose(/^duty/);
    fireEvent.change(screen.getByLabelText("Name"), { target: { value: "speed" } });
    expect(await screen.findByText("Unsaved changes")).toBeInTheDocument();
    expect(outline().getByRole("button", { name: /^speed/ })).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Save" }));
    expect(await screen.findByText("Saved and valid.")).toBeInTheDocument();
    expect(fake.writes[0]).toMatchObject({ path, etag: "k1" });
    const written = JSON.parse(fake.writes[0]?.content ?? "{}") as { readonly metrics: readonly unknown[] };
    expect(written.metrics[0]).toMatchObject({ name: "speed", unit: "%" });
  });

  it("saves from the keyboard, and only when there is something to save", async () => {
    const { fake } = await editingDevice();
    shortcut("s");
    expect(fake.writes).toHaveLength(0);
    choose(/^duty/);
    fireEvent.change(screen.getByLabelText("Unit"), { target: { value: "Cel" } });
    shortcut("s", screen.getByLabelText("Unit"));
    await vi.waitFor(() => {
      expect(fake.writes).toHaveLength(1);
    });
    expect(await screen.findByText("Saved and valid.")).toBeInTheDocument();
  });

  it("puts the binary's issue next to the field it names, in the details list, and in the status", async () => {
    const fake = deviceBridge();
    fake.actions.set("kind.validate", refusal("/metrics/0/data_type", "unknown type"));
    const { stores } = await editingDevice(fake);
    expect(await screen.findByText("Saved, but the binary refuses it: 1 issue.")).toBeInTheDocument();
    expect(outline().getByRole("img", { name: "Has issues" })).toBeInTheDocument();
    const list = within(screen.getByRole("region", { name: "Issues" }));
    expect(list.getByText("Reading duty · Type")).toBeInTheDocument();
    fireEvent.click(list.getByRole("button", { name: /unknown type/ }));
    const type = screen.getByLabelText("Type");
    expect(type).toHaveAttribute("aria-invalid", "true");
    expect(type).toHaveAccessibleDescription(/unknown type/);
    expect(stores.devices.snapshot()).toMatchObject({ selected: { part: "metric", key: "m1" } });
  });

  it("asks for the file to be checked again, and shows a check that could not run", async () => {
    const { fake } = await editingDevice();
    fake.actions.set("kind.validate", failed(cannedFailure("bridge_failed", 500, "the binary is gone")));
    fireEvent.click(screen.getByRole("button", { name: "Check again" }));
    expect(await screen.findByText(/could not be checked: the binary is gone/)).toBeInTheDocument();
    fake.actions.set("kind.validate", valid);
    fireEvent.click(screen.getByRole("button", { name: "Check again" }));
    expect(await screen.findByText("Saved and valid.")).toBeInTheDocument();
  });

  it("adds a reading and a command, moves to the new one's name, and keeps Save off until it is named", async () => {
    await editingDevice();
    fireEvent.click(screen.getByRole("button", { name: "Add a reading" }));
    expect(screen.getByLabelText("Name")).toHaveFocus();
    expect(outline().getByRole("button", { name: /^Unnamed reading/ })).toHaveAttribute("aria-current", "true");
    expect(screen.getByRole("button", { name: "Save" })).toHaveAttribute("aria-disabled", "true");
    expect(screen.getAllByText("Give this reading a name.").length).toBeGreaterThan(0);
    expect(await screen.findByText("Unsaved changes. Name everything to save.")).toBeInTheDocument();
    fireEvent.change(screen.getByLabelText("Name"), { target: { value: "rpm" } });
    expect(screen.getByRole("button", { name: "Save" })).toHaveAttribute("aria-disabled", "false");
    fireEvent.click(screen.getByRole("button", { name: "Add a command" }));
    expect(screen.getByRole("heading", { name: "Command" })).toBeInTheDocument();
    expect(screen.getByLabelText("Name")).toHaveFocus();
  });

  it("removes the selected reading or command, and puts them back on revert", async () => {
    const { stores } = await editingDevice();
    choose(/^mode/);
    fireEvent.click(screen.getByRole("button", { name: "Remove this reading" }));
    expect(outline().queryByRole("button", { name: /^mode/ })).toBeNull();
    choose(/^stop/);
    fireEvent.click(screen.getByRole("button", { name: "Remove this command" }));
    expect(outline().queryByRole("button", { name: /^stop/ })).toBeNull();
    expect(screen.getByRole("heading", { name: "Device" })).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Revert" }));
    expect(outline().getByRole("button", { name: /^mode/ })).toBeInTheDocument();
    expect(outline().getByRole("button", { name: /^stop/ })).toBeInTheDocument();
    expect(stores.devices.snapshot()).toMatchObject({ selected: { part: "kind" } });
  });

  it("stops at a conflict, offers a reload, and reads the file again on request", async () => {
    const { fake } = await editingDevice();
    fake.writeOutcomes.set(path, failed(cannedFailure("conflict", 409, "changed on disk")));
    fireEvent.change(screen.getByLabelText("Name"), { target: { value: "fan2" } });
    fireEvent.click(screen.getByRole("button", { name: "Save" }));
    expect(await screen.findByText(/the file changed on disk/)).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Save" })).toHaveAttribute("aria-disabled", "true");
    fake.addFile(path, { ...fan, kind: "fan3" }, "k2");
    fireEvent.click(screen.getByRole("button", { name: "Reload from disk" }));
    await vi.waitFor(() => {
      expect(screen.getByLabelText("Name")).toHaveValue("fan3");
    });
  });

  it("asks before leaving with unsaved changes, and leaves the three ways it offers", async () => {
    const { stores, fake } = await editingDevice();
    fireEvent.change(screen.getByLabelText("Name"), { target: { value: "fan2" } });
    const home = screen.getByRole("button", { name: "Streams Simulator, workspace home" });
    fireEvent.click(home);
    const dialog = screen.getByRole("dialog", { name: "Leave with unsaved changes?" });
    expect(within(dialog).getByRole("button", { name: "Stay" })).toHaveFocus();
    fireEvent.keyDown(dialog, { key: "Escape" });
    expect(screen.queryByRole("dialog")).toBeNull();
    fireEvent.click(home);
    fireEvent.click(screen.getByRole("button", { name: "Save and leave" }));
    expect(await screen.findByRole("heading", { name: "lab-workspace" })).toBeInTheDocument();
    expect(fake.writes).toHaveLength(1);
    expect(stores.devices.snapshot()).toEqual({ phase: "idle" });
  });

  it("discards the changes when asked to, and stays when saving is refused", async () => {
    const { fake } = await editingDevice();
    fireEvent.change(screen.getByLabelText("Name"), { target: { value: "" } });
    fireEvent.click(screen.getByRole("button", { name: "Streams Simulator, workspace home" }));
    fireEvent.click(screen.getByRole("button", { name: "Save and leave" }));
    await vi.waitFor(() => {
      expect(screen.queryByRole("dialog")).toBeNull();
    });
    expect(fake.writes).toHaveLength(0);
    fireEvent.click(screen.getByRole("button", { name: "Streams Simulator, workspace home" }));
    fireEvent.click(screen.getByRole("button", { name: "Discard and leave" }));
    expect(await screen.findByRole("heading", { name: "lab-workspace" })).toBeInTheDocument();
  });

  it("warns before the tab closes only while there are unsaved changes", async () => {
    await editingDevice();
    const clean = new Event("beforeunload", { cancelable: true });
    window.dispatchEvent(clean);
    expect(clean.defaultPrevented).toBe(false);
    fireEvent.change(screen.getByLabelText("Name"), { target: { value: "fan2" } });
    const dirty = new Event("beforeunload", { cancelable: true });
    act(() => {
      window.dispatchEvent(dirty);
    });
    expect(dirty.defaultPrevented).toBe(true);
  });

  it("lists its own keys in the keyboard help", async () => {
    await editingDevice();
    fireEvent.click(screen.getByRole("button", { name: "Keyboard" }));
    const dialog = screen.getByRole("dialog", { name: "Keyboard" });
    expect(within(dialog).getByRole("region", { name: "Device editor" })).toBeInTheDocument();
  });
});
