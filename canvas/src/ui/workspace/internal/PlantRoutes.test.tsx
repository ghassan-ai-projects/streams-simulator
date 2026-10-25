import { fireEvent, screen, within } from "@testing-library/react";
import axe from "axe-core";
import { describe, expect, it } from "vitest";
import { FakeBridge } from "../../../test/fakeBridge";
import { success } from "../../../test/outcome";
import { lineageWire } from "../../../test/provenanceFixture";
import { sampleWorkspace } from "../../../test/workspaceFixture";
import { navigation, show } from "../../../test/workspaceHarness";

const domain = (id: string) => ({ path: `domains/${id}.domain.json`, id, title: id, instances: 2 });

function bridge() {
  const fake = new FakeBridge();
  fake.workspaceOutcome = success({
    ...sampleWorkspace,
    domains: ["arduino-thermal-twin", "chiller-pilot", "alarm-playback"].map(domain),
    compositions: [],
    kinds: [{ path: "kinds/fan.kind.json", kind: "fan", title: "" }],
    runs: [],
  });
  fake.addFile("kinds/fan.kind.json", { kind: "fan", metrics: [], commands: [] });
  fake.actions.set("domain.lineage", success(lineageWire));
  return fake;
}

function plant(name: string) {
  const row = screen.getByRole("region", { name: "Plants" }).querySelector<HTMLElement>(`li[title*="${name}"]`);
  if (row === null) throw new Error(`no row for ${name}`);
  return within(row);
}

describe("why a read-only plant in the list is read-only", () => {
  it("states the reason under each read-only plant, with the binary's sentence on hover", async () => {
    await show(bridge());
    expect(
      await plant("arduino-thermal-twin").findByText("Read-only: this plant is generated from a twin."),
    ).toHaveAttribute("title", expect.stringContaining("twins/fan.twin.json"));
    expect(plant("chiller-pilot").getByText("Read-only: a shipped example.")).toBeInTheDocument();
    expect(plant("alarm-playback").getByText("Read-only: written by hand as a domain file.")).toBeInTheDocument();
  });

  it("offers the Behaviour step of the twin's device, and nothing to open for the others", async () => {
    const nav = navigation();
    await show(bridge(), nav);
    await plant("arduino-thermal-twin").findByText(/generated from a twin/);
    expect(plant("chiller-pilot").queryByRole("button", { name: /Behaviour/ })).toBeNull();
    fireEvent.click(plant("arduino-thermal-twin").getByRole("button", { name: /^Open the Behaviour step of/ }));
    expect(nav.location.hash).toBe("#/device?device=kinds%2Ffan.kind.json&step=behaviour");
  });

  it("has no accessibility violations", async () => {
    const view = await show(bridge());
    await plant("chiller-pilot").findByText("Read-only: a shipped example.");
    expect((await axe.run(view.container)).violations).toEqual([]);
  });
});
