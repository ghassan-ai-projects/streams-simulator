import { describe, expect, it } from "vitest";
import { beltComposition, beltDomain, beltSpecs, composition, domain } from "../../../test/plantFixture";
import { parsePlant } from "../../plant";
import { headerHeight, nodeSize, plantView } from "./plantView";
import { routeView } from "./wireView";
import { canGroupWires, channelRows, openingWires, rowPins } from "./wires";

const belt = parsePlant(beltDomain, beltComposition);

describe("grouped wires", () => {
  it("opens a plant grouped once two instances share three links one way, and offers it from two", () => {
    expect(openingWires(belt)).toBe("grouped");
    expect(canGroupWires(belt)).toBe(true);
    const mini = parsePlant(domain, composition);
    expect(openingWires(mini)).toBe("each");
    expect(canGroupWires(mini)).toBe(false);
    const two = { ...belt, links: belt.links.filter((link) => link.fromPort !== "enabled_out") };
    expect(openingWires(two)).toBe("each");
    expect(canGroupWires(two)).toBe(true);
  });

  it("draws one wire per pair of instances, led by the way that carries more material", () => {
    const view = plantView(belt, beltSpecs, { open: null, wires: "grouped" });
    expect(view.wires).toBe("grouped");
    expect(view.edges.map((edge) => [edge.id, edge.count, edge.reverse, edge.family])).toEqual([
      ["F-1.>B-1.", 1, 1, null],
      ["D-1.>B-1.", 3, 2, null],
    ]);
  });

  it("keeps one family's colour on a wire whose links all share it", () => {
    const power = { ...belt, links: belt.links.filter((link) => /drag|run_cap/.test(link.fromPort ?? "")) };
    const view = plantView(power, beltSpecs, { open: null, wires: "grouped" });
    expect(view.edges.map((edge) => [edge.id, edge.count, edge.reverse, edge.family])).toEqual([
      ["D-1.>B-1.", 2, 0, "power"],
    ]);
  });

  it("puts the feeder before the belt it feeds, whatever their ids", () => {
    const view = plantView(belt, beltSpecs, { open: null, wires: "grouped" });
    const x = (id: string) => view.layout.nodes.get(id)?.x ?? 0;
    expect(x("F-1")).toBeLessThan(x("B-1"));
  });

  it("sizes a grouped node by its measured channels, not its ports", () => {
    const view = plantView(belt, beltSpecs, { open: null, wires: "grouped" });
    const drive = view.nodes.find((node) => node.id === "D-1");
    expect(drive?.ported).toBe(false);
    expect(drive?.units[1]).toBe("°C");
    expect(drive === undefined ? null : rowPins(drive)).toEqual({ ins: [], outs: [] });
    expect(view.layout.nodes.get("D-1")?.height).toBe(headerHeight("D-1") + nodeSize.padding + 6 * nodeSize.pin);
    expect(view.layout.nodes.get("B-1")?.height).toBe(headerHeight("B-1") + nodeSize.padding + nodeSize.pin);
  });

  it("lists six channels, or five and a line for the rest", () => {
    expect(channelRows(0)).toEqual({ listed: 0, more: 0 });
    expect(channelRows(6)).toEqual({ listed: 6, more: 0 });
    expect(channelRows(7)).toEqual({ listed: 5, more: 2 });
  });

  it("keeps pins and one wire per link when wires are not grouped", () => {
    const view = plantView(belt, beltSpecs, { open: null });
    expect(view.wires).toBe("each");
    expect(view.edges).toHaveLength(7);
    expect(view.nodes.every((node) => node.ported)).toBe(true);
  });

  it("ends a grouped wire on the sides of the two nodes that face each other", () => {
    const view = plantView(belt, beltSpecs, { open: null, wires: "grouped" });
    const routes = routeView(view, "full", "pins", 1);
    const feeder = view.layout.nodes.get("F-1");
    const conveyor = view.layout.nodes.get("B-1");
    const route = routes.get("F-1.>B-1.") ?? [];
    expect(route[0]?.x).toBe((feeder?.x ?? 0) + (feeder?.width ?? 0));
    expect(route.at(-1)?.x).toBe(conveyor?.x);
  });
});
