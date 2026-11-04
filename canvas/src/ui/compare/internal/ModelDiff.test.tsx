import { act, fireEvent, screen, within } from "@testing-library/react";
import axe from "axe-core";
import { beforeEach, describe, expect, it } from "vitest";
import { mountApp } from "../../../test/appHarness";
import { comparisonWire } from "../../../test/compareFixture";
import { stubFlowDom } from "../../../test/dom";
import { cannedFailure } from "../../../test/fakeBridge";
import { failure, success } from "../../../test/outcome";
import { diffWire } from "../../../test/provenanceFixture";
import { bridgeWith, reviewing } from "../../../test/reviewActionsHarness";
import { reviewHash } from "../../../test/screenFixture";

beforeEach(() => {
  stubFlowDom();
});

async function opened() {
  const app = await reviewing();
  app.bridge.actions.set("domain.diff", success(diffWire));
  const opener = screen.getByRole("button", { name: "Compare model with the current plant" });
  opener.focus();
  fireEvent.click(opener);
  const dialog = await screen.findByRole("dialog", { name: "What changed in the model" });
  return { ...app, dialog };
}

describe("Compare model with the current plant", () => {
  it("compares the model the run recorded with the plant file on disk and says the version did not move", async () => {
    const { dialog, bridge } = await opened();
    await within(dialog).findByRole("region", { name: "Changes to thermal_sensor" });
    expect(bridge.calls.find((call) => call.name === "domain.diff")?.args).toEqual({
      a: "w-1",
      b: "domains/mini.domain.json",
    });
    expect(within(dialog).getByRole("alert")).toHaveTextContent("The model changed but its version did not.");
    expect(within(dialog).getByText("1 change in 1 device kind")).toBeInTheDocument();
  });

  it("groups the changes by device kind with the binary's sentence, before and after, and the instance it affects", async () => {
    const { dialog } = await opened();
    const group = within(await within(dialog).findByRole("region", { name: "Changes to thermal_sensor" }));
    expect(group.getByText("gain changed")).toBeInTheDocument();
    expect(group.getByText(/coef changed from -0.02 to -0.05/)).toBeInTheDocument();
    expect(group.getByText("Before: -0.02")).toBeInTheDocument();
    expect(group.getByText("After: -0.05")).toBeInTheDocument();
    expect(group.getByText("Affects zone-01")).toBeInTheDocument();
    expect(group.getByText(/fan_drives_thermal_sensor: fan to thermal_sensor \(fan-01>zone-01\)/)).toBeInTheDocument();
    expect(within(dialog).getByText("Unchanged: 2 kinds, 5 readings, 4 commands, 14 faults.")).toBeInTheDocument();
  });

  it("carries no warning when the version moved with the model", async () => {
    const app = await reviewing();
    const moved = {
      ...diffWire,
      identity: { ...diffWire.identity, same_id_and_version_different_digest: false, same_version: false },
    };
    app.bridge.actions.set("domain.diff", success(moved));
    fireEvent.click(screen.getByRole("button", { name: "Compare model with the current plant" }));
    const dialog = await screen.findByRole("dialog", { name: "What changed in the model" });
    await within(dialog).findByRole("region", { name: "Changes to thermal_sensor" });
    expect(within(dialog).queryByRole("alert")).toBeNull();
  });

  it("lets the person choose another run of the plant instead and compares the two recorded models", async () => {
    const { dialog, bridge } = await opened();
    await within(dialog).findByRole("region", { name: "Changes to thermal_sensor" });
    fireEvent.change(within(dialog).getByLabelText("Compare the model of this run with"), { target: { value: "w-3" } });
    await act(async () => {
      await Promise.resolve();
    });
    expect(bridge.calls.filter((call) => call.name === "domain.diff").at(-1)?.args).toEqual({ a: "w-1", b: "w-3" });
  });

  it("shows the binary's refusal", async () => {
    const app = await reviewing();
    app.bridge.actions.set(
      "domain.diff",
      failure(cannedFailure("artifact_invalid", 422, "run.json has no domain_spec")),
    );
    fireEvent.click(screen.getByRole("button", { name: "Compare model with the current plant" }));
    const dialog = await screen.findByRole("dialog", { name: "What changed in the model" });
    expect(await within(dialog).findByRole("alert")).toHaveTextContent(
      "run.json has no domain_spec (artifact_invalid)",
    );
  });

  it("closes on Escape and gives the focus back to the button", async () => {
    const { dialog } = await opened();
    await within(dialog).findByRole("region", { name: "Changes to thermal_sensor" });
    fireEvent.keyDown(document, { key: "Escape" });
    expect(screen.queryByRole("dialog", { name: "What changed in the model" })).toBeNull();
    expect(screen.getByRole("button", { name: "Compare model with the current plant" })).toHaveFocus();
  });

  it("has no accessibility violations", async () => {
    const { dialog } = await opened();
    await within(dialog).findByRole("region", { name: "Changes to thermal_sensor" });
    expect((await axe.run(dialog)).violations).toEqual([]);
  });

  it("opens from the lineage of the plant with the dialog already open", async () => {
    const { bridge } = bridgeWith();
    bridge.actions.set("domain.diff", success(diffWire));
    mountApp(bridge, `${reviewHash()}&model=1`);
    expect(await screen.findByRole("dialog", { name: "What changed in the model" })).toBeInTheDocument();
  });
});

describe("the model link in the run comparison", () => {
  const different = {
    ...comparisonWire,
    comparability_limits: [
      { code: "different_domain_digest", message: "The runs used different versions of the domain." },
      ...comparisonWire.comparability_limits,
    ],
  };

  async function compared(wire: object) {
    const app = await reviewing();
    app.bridge.handlers.set("run.compare", () => success(wire));
    app.bridge.actions.set("domain.diff", success(diffWire));
    fireEvent.click(screen.getByRole("button", { name: "Compare with…" }));
    const dialog = await screen.findByRole("dialog", { name: "Compare with another run" });
    await act(async () => {
      fireEvent.click(within(dialog).getByRole("button", { name: /Baseline/ }));
      await Promise.resolve();
    });
    await within(dialog).findByRole("region", { name: "What differs" });
    return { ...app, dialog };
  }

  it("offers to show what changed in the model when the runs recorded different domain digests", async () => {
    const { dialog, bridge } = await compared(different);
    const link = within(dialog).getByRole("region", { name: "The model of the two runs" });
    expect(bridge.calls.some((call) => call.name === "domain.diff")).toBe(false);
    fireEvent.click(within(link).getByRole("button", { name: "See what changed in the model" }));
    expect(await within(link).findByRole("region", { name: "Changes to thermal_sensor" })).toBeInTheDocument();
    expect(bridge.calls.find((call) => call.name === "domain.diff")?.args).toEqual({ a: "w-3", b: "w-1" });
  });

  it("says nothing about the model when both runs used the same one", async () => {
    const { dialog } = await compared(comparisonWire);
    expect(within(dialog).queryByRole("region", { name: "The model of the two runs" })).toBeNull();
  });
});
