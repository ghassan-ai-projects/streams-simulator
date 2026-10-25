import { screen, waitFor, within } from "@testing-library/react";
import axe from "axe-core";
import { describe, expect, it } from "vitest";
import { cannedFailure, FakeBridge } from "../../../test/fakeBridge";
import { failure, success } from "../../../test/outcome";
import { kindCheck } from "../../../test/kindCheckFake";
import { lineageWire } from "../../../test/provenanceFixture";
import { sampleWorkspace } from "../../../test/workspaceFixture";
import { show } from "../../../test/workspaceHarness";

const names = ["fan", "thermal_sensor", "heater", "lonely"];
const kinds = names.map((kind) => ({ path: `kinds/${kind}.kind.json`, kind, title: "" }));

function bridge() {
  const fake = new FakeBridge();
  fake.workspaceOutcome = success({ ...sampleWorkspace, kinds });
  names.forEach((kind) => {
    fake.addFile(`kinds/${kind}.kind.json`, { kind, metrics: [], commands: [] });
  });
  fake.addFile("twins/fan.twin.json", { id: "arduino-thermal-twin", kinds: [] });
  fake.handlers.set("kind.check", kindCheck);
  fake.actions.set("domain.lineage", success(lineageWire));
  return fake;
}

function row(name: string) {
  const devices = within(screen.getByRole("region", { name: "Devices" }));
  return within(devices.getByText(name).closest("li") as HTMLElement);
}

describe("the label of a device in the workspace list, from the lineage of the workspace", () => {
  it("says a device is included in another device's twin, that one has a twin of its own, and that one needs one", async () => {
    await show(bridge());
    expect(await row("thermal_sensor").findByText("Included in twin arduino-thermal-twin")).toBeInTheDocument();
    expect(row("heater").getByText("Twin created")).toBeInTheDocument();
    expect(row("lonely").getByText("Needs a twin")).toBeInTheDocument();
    expect(row("fan").queryByText("Needs a twin")).toBeNull();
  });

  it("keeps the earlier wording when the bridge gives no lineage", async () => {
    const fake = bridge();
    fake.actions.set("domain.lineage", failure(cannedFailure("unknown_action", 404)));
    await show(fake);
    expect(await row("thermal_sensor").findByText("Needs a twin")).toBeInTheDocument();
  });

  it("asks for the lineage once when the workspace is read, and again when it changes", async () => {
    const fake = bridge();
    await show(fake);
    await waitFor(() => {
      expect(fake.calls.filter((call) => call.name === "domain.lineage")).toHaveLength(1);
    });
    fake.changeOnDisk();
    await waitFor(() => {
      expect(fake.calls.filter((call) => call.name === "domain.lineage")).toHaveLength(2);
    });
  });

  it("has no accessibility violations", async () => {
    const view = await show(bridge());
    await row("thermal_sensor").findByText("Included in twin arduino-thermal-twin");
    expect((await axe.run(view.container)).violations).toEqual([]);
  });
});
