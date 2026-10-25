import { fireEvent, screen, within } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { mountApp } from "../../../test/appHarness";
import { kindCheck } from "../../../test/kindCheckFake";
import { choicesPath, coolerDocument, coolerPath, refusal, suggestion, twinBridge } from "../../../test/twinFixture";
import { behaviourOf, createTwin, hash, inCommand, inReading, pickStep } from "../../../test/twinHarness";

const lacking = {
  ...coolerDocument,
  metrics: [
    { name: "temperature", data_type: "number", unit: "Cel", eu_range: { min: 0, max: 100 } },
    { name: "duty", data_type: "number", unit: "%", eu_range: { min: 0, max: 100 }, writable: true },
    { name: "pressure", data_type: "number", resolution: 1 },
  ],
};

function bridge(stored = false) {
  const fake = twinBridge();
  fake.addFile(coolerPath, lacking, "k1");
  if (stored) fake.addFile(choicesPath, suggestion, "c1");
  return fake;
}

function outline() {
  return within(screen.getByRole("navigation", { name: "Device outline" }));
}

function step(name: RegExp) {
  return within(screen.getByRole("navigation", { name: "Device steps" })).getByRole("button", { name });
}

describe("what a twin needs, shown in the interface step before the behaviour", () => {
  it("lists every need at once, beside the readings that lack it, and keeps it apart from the interface being valid", async () => {
    mountApp(bridge(), hash);
    expect(await screen.findByText("Needed for a twin: 3")).toBeInTheDocument();
    expect(step(/^Interface/)).toHaveTextContent("Valid");
    expect(outline().getAllByRole("img", { name: "Needed for a twin" })).toHaveLength(3);
    const needs = within(screen.getByRole("region", { name: "Needed for a twin" }));
    expect(needs.getAllByRole("button").map((button) => button.textContent)).toEqual([
      "Reading temperature · ResolutionNeeds a resolution",
      "Reading duty · ResolutionNeeds a resolution",
      "Reading pressure · Unitpressure needs a unit",
    ]);
    fireEvent.click(outline().getByRole("button", { name: /^pressure/ }));
    const unit = screen.getByLabelText("Unit");
    expect(unit).toHaveAccessibleDescription(/pressure needs a unit/);
    expect(unit).not.toHaveAttribute("aria-invalid", "true");
    expect(screen.queryByRole("region", { name: "Issues" })).toBeNull();
    fireEvent.click(needs.getByRole("button", { name: /Reading temperature/ }));
    expect(screen.getByLabelText("Resolution")).toHaveAccessibleDescription(/Needs a resolution/);
  });

  it("clears a need as soon as the draft meets it, without saving", async () => {
    mountApp(bridge(), hash);
    await screen.findByText("Needed for a twin: 3");
    fireEvent.click(outline().getByRole("button", { name: /^temperature/ }));
    fireEvent.change(screen.getByLabelText("Resolution"), { target: { value: "0.1" } });
    expect(await screen.findByText("Needed for a twin: 2")).toBeInTheDocument();
    expect(await screen.findByText("Unsaved changes")).toBeInTheDocument();
  });

  it("shows the same needs on the readings of the behaviour step, with a way back to each field, and blocks Create", async () => {
    await behaviourOf(bridge(true));
    const reading = inReading("temperature");
    expect(reading.getByText("Needs a resolution")).toBeInTheDocument();
    expect(inReading("pressure").getByText("pressure needs a unit")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Create twin" })).toHaveAttribute("aria-disabled", "true");
    expect(screen.getByText("Set what the twin needs in the Interface step first.")).toBeInTheDocument();
    fireEvent.click(inReading("pressure").getByRole("button", { name: "Set it in the Interface step" }));
    expect(await screen.findByRole("region", { name: "Editing cooler" })).toBeInTheDocument();
    expect(screen.getByLabelText("Unit")).toHaveFocus();
    expect(screen.getByLabelText("Name")).toHaveValue("pressure");
  });

  it("says what to set first when the binary can suggest no behaviour for a device that lacks it, and leads to each field", async () => {
    mountApp(bridge(), hash);
    await screen.findByText("Needed for a twin: 3");
    pickStep(/^Behaviour/);
    const lackingFirst = within(await screen.findByRole("region", { name: "What the twin needs first" }));
    expect(lackingFirst.getAllByRole("button")).toHaveLength(3);
    expect(screen.queryByRole("group", { name: "Reading temperature" })).toBeNull();
    fireEvent.click(lackingFirst.getByRole("button", { name: /pressure needs a unit/ }));
    expect(await screen.findByRole("region", { name: "Editing cooler" })).toBeInTheDocument();
    expect(screen.getByLabelText("Unit")).toHaveFocus();
    fireEvent.change(screen.getByLabelText("Unit"), { target: { value: "Pa" } });
    fireEvent.click(outline().getByRole("button", { name: /^temperature/ }));
    fireEvent.change(screen.getByLabelText("Resolution"), { target: { value: "0.1" } });
    fireEvent.click(outline().getByRole("button", { name: /^duty/ }));
    fireEvent.change(screen.getByLabelText("Resolution"), { target: { value: "1" } });
    fireEvent.click(screen.getByRole("button", { name: "Save" }));
    expect(await screen.findByText("Saved and valid.")).toBeInTheDocument();
    pickStep(/^Behaviour/);
    expect(await screen.findByRole("group", { name: "Reading temperature" })).toBeInTheDocument();
  });

  it("leads from a command that lacks a safe-stop statement to the toggle in the interface step", async () => {
    const fake = twinBridge();
    const need = {
      pointer: "/commands/0/safe_stop",
      message: 'needs "safe_stop"',
      keyword: "twin_member_required",
      member: "safe_stop",
    };
    fake.handlers.set("kind.check", (args) => (args.twin === true ? refusal(need) : kindCheck(args)));
    await behaviourOf(fake);
    const command = inCommand("set_duty");
    expect(command.getByText("Needs a safe-stop choice")).toBeInTheDocument();
    fireEvent.click(command.getByRole("button", { name: "Set it in the Interface step" }));
    expect(await screen.findByRole("region", { name: "Editing cooler" })).toBeInTheDocument();
    expect(screen.getByLabelText("Safe stop")).toHaveFocus();
  });

  it("shows a need of the device itself above the readings, with a way to its field", async () => {
    const fake = twinBridge();
    const name = { pointer: "/kind", message: "kind name", keyword: "twin_name_pattern", member: "kind" };
    fake.handlers.set("kind.check", (args) => (args.twin === true ? refusal(name) : kindCheck(args)));
    await behaviourOf(fake);
    expect(screen.getByText("A twin needs a simpler name: lower-case letters, digits and _")).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Set it in the Interface step" }));
    expect(await screen.findByRole("region", { name: "Editing cooler" })).toBeInTheDocument();
    expect(screen.getByLabelText("Name")).toHaveFocus();
  });

  it("lets Create through once the interface lacks nothing for a twin", async () => {
    await behaviourOf();
    expect(screen.getByRole("button", { name: "Create twin" })).toHaveAttribute("aria-disabled", "false");
    createTwin();
    expect(await screen.findByText(/^Twin created\./)).toBeInTheDocument();
    pickStep(/^Interface/);
    expect(await screen.findByText("Ready for a twin")).toBeInTheDocument();
  });
});
