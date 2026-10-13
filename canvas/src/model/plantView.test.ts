import { describe, expect, it } from "vitest";
import { campus, chillerPorts, composition, domain, towerPorts } from "../test/plantFixture";
import { parsePlant } from "./plant";
import { hallOf, hallsOf, nodeSize, plantView } from "./plantView";
import { parsePorts } from "./ports";

const specs = new Map([
  ["lib/chiller", { ports: parsePorts(chillerPorts), category: "cool" as const }],
  ["lib/tower", { ports: parsePorts(towerPorts), category: "cool" as const }],
]);

describe("plant view", () => {
  it("shows every instance with typed pins and port-to-port edges", () => {
    const view = plantView(parsePlant(domain, composition), specs, null);
    expect(view.halls).toEqual([]);
    expect(view.nodes.map((node) => [node.id, node.group])).toEqual([
      ["CH-1", "Cooling"],
      ["CT-1", "Cooling"],
      ["MB-1", "Power"],
    ]);
    expect(view.edges.map((edge) => [edge.id, edge.family])).toEqual([
      ["CH-1.heat_out>CT-1.heat_in", "heat"],
      ["CT-1.basin_out>CH-1.cond_in", "temp"],
      ["CH-1.power_out>MB-1.loads_in", "power"],
    ]);
    expect(view.layout.nodes.get("CH-1")?.height).toBe(nodeSize.header + nodeSize.padding + 2 * nodeSize.pin);
    expect([...view.layout.groups.keys()]).toEqual(["Cooling", "Power"]);
  });

  it("starts a campus at hall level with bundled links and shared instances", () => {
    const plant = parsePlant(campus(3));
    expect(hallOf("h02/M-1")).toBe("h02");
    expect(hallOf("CAL-1")).toBeNull();
    expect(hallsOf(plant)).toEqual(["h01", "h02", "h03"]);
    const view = plantView(plant, new Map(), null);
    expect(view.nodes.map((node) => [node.id, node.detail])).toEqual([
      ["CAL-1", "calendar"],
      ["h01", "2 instances"],
      ["h02", "2 instances"],
      ["h03", "2 instances"],
    ]);
    expect(view.edges.map((edge) => [edge.from, edge.to, edge.count])).toEqual([
      ["CAL-1", "h01", 2],
      ["CAL-1", "h02", 2],
      ["CAL-1", "h03", 2],
    ]);
    expect(view.layout.nodes.get("h01")?.width).toBe(nodeSize.groupWidth);
  });

  it("opens one hall with the shared instances wired into it", () => {
    const view = plantView(parsePlant(campus(3)), new Map(), "h02");
    expect(view.nodes.map((node) => node.id)).toEqual(["CAL-1", "h02/M-1", "h02/M-2"]);
    expect(view.edges).toHaveLength(3);
    expect(plantView(parsePlant(campus(3)), new Map(), "h99").nodes).toHaveLength(4);
  });

  it("treats a single prefix as an ordinary plant", () => {
    const one = parsePlant({ instances: [{ kind: "k", names: ["h01/a", "h01/b"], area: "A" }] });
    expect(hallsOf(one)).toEqual([]);
    expect(plantView(one, new Map(), null).nodes).toHaveLength(2);
  });
});
