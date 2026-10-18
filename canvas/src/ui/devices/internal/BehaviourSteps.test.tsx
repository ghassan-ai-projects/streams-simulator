import { fireEvent, screen, within } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { mountApp } from "../../../test/appHarness";
import { cannedFailure } from "../../../test/fakeBridge";
import { failure, success } from "../../../test/outcome";
import { choicesPath, coolerPath, sensorPath, twinBridge } from "../../../test/twinFixture";
import { behaviourOf, hash, pickStep } from "../../../test/twinHarness";
import { sampleWorkspace } from "../../../test/workspaceFixture";

describe("the devices of a twin", () => {
  it("adds another device with its readings, and leaves it out again", async () => {
    await behaviourOf();
    const side = within(screen.getByRole("navigation", { name: "Devices in the twin" }));
    fireEvent.click(side.getByRole("button", { name: "Add sensor to the twin" }));
    expect(await screen.findByRole("group", { name: "Reading humidity" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { level: 2, name: "sensor" })).toBeInTheDocument();
    expect(side.getByText("No other devices in the workspace.")).toBeInTheDocument();
    expect(side.queryByRole("button", { name: "Leave cooler out of the twin" })).toBeNull();
    fireEvent.click(side.getByRole("button", { name: "Leave sensor out of the twin" }));
    expect(screen.queryByRole("group", { name: "Reading humidity" })).toBeNull();
    expect(side.getByRole("button", { name: "Add sensor to the twin" })).toBeInTheDocument();
  });

  it("names a device whose file cannot be read by the name of the file", async () => {
    const fake = twinBridge();
    const kinds = [
      { path: coolerPath, kind: "cooler", title: "" },
      { path: sensorPath, kind: "", title: "" },
    ];
    fake.workspaceOutcome = success({ ...sampleWorkspace, kinds });
    await behaviourOf(fake);
    expect(screen.getByRole("button", { name: "Add sensor to the twin" })).toBeInTheDocument();
  });
});

describe("opening and leaving the behaviour", () => {
  it("waits while the choices are read and says why they could not be", async () => {
    const fake = twinBridge();
    const gates = fake.holdRead(choicesPath, 1);
    mountApp(fake, hash);
    await screen.findByRole("region", { name: "Editing cooler" });
    pickStep(/^Behaviour/);
    expect(await screen.findByText("Opening the behaviour…")).toBeInTheDocument();
    gates[0]?.resolve(failure(cannedFailure("boom", 500, "the bridge broke")));
    expect(await screen.findByRole("alert")).toHaveTextContent("The behaviour could not be opened: the bridge broke");
  });

  it("keeps Create off until the interface is saved, and shows the interface step as unsaved", async () => {
    await behaviourOf();
    pickStep(/^Interface/);
    fireEvent.change(screen.getByLabelText("Name"), { target: { value: "cooler2" } });
    pickStep(/^Behaviour/);
    expect(await screen.findByText(/Save the interface first/)).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Create twin" })).toHaveAttribute("aria-disabled", "true");
    const steps = within(screen.getByRole("navigation", { name: "Device steps" }));
    expect(steps.getByRole("button", { name: /^Interface/ })).toHaveTextContent("Unsaved");
  });

  it("offers a reload when the choices file changed on disk", async () => {
    const { fake } = await behaviourOf();
    fake.writeOutcomes.set(choicesPath, failure(cannedFailure("conflict", 409, "changed")));
    const name = screen.getByLabelText("Twin name");
    fireEvent.change(name, { target: { value: "cooler-b" } });
    fireEvent.blur(name);
    expect(await screen.findByText(/Choices not saved: the file changed on disk/)).toBeInTheDocument();
    fake.writeOutcomes.clear();
    fireEvent.click(screen.getByRole("button", { name: "Reload from disk" }));
    await vi.waitFor(() => {
      expect(screen.getByLabelText("Twin name")).toHaveValue("cooler-twin");
    });
  });

  it("keeps the choices when the step is left and opened again", async () => {
    const { fake } = await behaviourOf();
    fireEvent.change(screen.getByLabelText("Twin name"), { target: { value: "cooler-b" } });
    pickStep(/^Interface/);
    pickStep(/^Behaviour/);
    await vi.waitFor(() => {
      expect(screen.getByLabelText("Twin name")).toHaveValue("cooler-b");
    });
    expect(fake.writes.some((write) => write.path === choicesPath)).toBe(true);
  });
});
