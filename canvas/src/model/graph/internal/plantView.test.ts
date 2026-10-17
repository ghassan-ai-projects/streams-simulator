import { describe, expect, it } from "vitest";
import { campus, chillerPorts, composition, domain, towerPorts } from "../../../test/plantFixture";
import { parsePlant } from "../../plant";
import { describeNode } from "./nodeWords";
import { reviewView } from "../../../test/modelFixtures";
import {
  areaCounts,
  hallCount,
  hallMembers,
  hallOf,
  hallsOf,
  headerHeight,
  idExtra,
  nodeSize,
  plainView,
  plantView,
} from "./plantView";

const specs = new Map([
  ["lib/chiller", { ports: chillerPorts.ports, category: "cool" as const }],
  ["lib/tower", { ports: towerPorts.ports, category: "cool" as const }],
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
    expect(view.layout.nodes.get("CH-1")?.height).toBe(headerHeight("CH-1") + nodeSize.padding + 2 * nodeSize.pin);
    expect([...view.layout.groups.keys()]).toEqual(["Cooling", "Power"]);
  });

  it("ends a link without a recorded port on the first unconnected pin of its side", () => {
    const plant = parsePlant(domain, composition);
    const unrecorded = { ...plant, links: plant.links.map((link) => ({ ...link, fromPort: null, toPort: null })) };
    const view = plantView(unrecorded, specs, null);
    expect(view.edges.map((edge) => [edge.from, edge.fromHandle, edge.to, edge.toHandle])).toEqual([
      ["CH-1", "heat_out", "CT-1", "heat_in"],
      ["CT-1", "basin_out", "CH-1", "cond_in"],
      ["CH-1", "heat_out", "MB-1", null],
    ]);
  });

  it("starts a campus at hall level with bundled links and shared instances", () => {
    const plant = parsePlant(campus(3));
    expect(hallOf("h02/M-1")).toBe("h02");
    expect(hallOf("CAL-1")).toBeNull();
    expect(hallsOf(plant)).toEqual(["h01", "h02", "h03"]);
    const view = plantView(plant, new Map(), null);
    expect(view.nodes.map((node) => [node.id, node.detail])).toEqual([
      ["CAL-1", "calendar · feeds all 3 halls"],
      ["h01", "2 instances"],
      ["h02", "2 instances"],
      ["h03", "2 instances"],
    ]);
    expect(view.edges).toEqual([]);
    const inside = plantView(plant, new Map(), "h01");
    expect(hallCount(inside, "h01")).toBe("2 instances plus 1 shared");
    expect(view.layout.nodes.get("h01")?.width).toBe(nodeSize.groupWidth);
  });

  it("keeps a shared instance's links when it feeds only some halls", () => {
    const partial = campus(3);
    const plant = parsePlant({ ...partial, links: partial.links.filter((link) => !link.to.startsWith("h03/")) });
    const view = plantView(plant, new Map(), null);
    expect(view.nodes[0]?.detail).toBe("calendar");
    expect(view.edges.map((edge) => [edge.from, edge.to, edge.count])).toEqual([
      ["CAL-1", "h01", 2],
      ["CAL-1", "h02", 2],
    ]);
  });

  it("opens one hall with the shared instances wired into it", () => {
    const view = plantView(parsePlant(campus(3)), new Map(), "h02");
    expect(view.nodes.map((node) => node.id)).toEqual(["CAL-1", "h02/M-1", "h02/M-2"]);
    expect(view.nodes.map((node) => node.label)).toEqual(["CAL-1", "M-1", "M-2"]);
    expect(view.edges).toHaveLength(3);
    expect(plantView(parsePlant(campus(3)), new Map(), "h99").nodes).toHaveLength(4);
  });

  it("treats a single prefix as an ordinary plant", () => {
    const one = parsePlant({ instances: [{ kind: "k", names: ["h01/a", "h01/b"], area: "A" }] });
    expect(hallsOf(one)).toEqual([]);
    expect(plantView(one, new Map(), null).nodes).toHaveLength(2);
  });

  it("describes an instance with its status at the playback time", () => {
    const node = plantView(parsePlant(domain), new Map(), null).nodes.find((entry) => entry.id === "CH-1");
    expect(node && describeNode(node, false, "alarm")).toMatch(/^CH-1, Alarm, /);
    expect(node && describeNode(node, true)).toMatch(/^CH-1, [^A].*, selected$/);
  });

  it("counts a hall's members by kind", () => {
    const plant = parsePlant(campus(3));
    const hall = hallsOf(plant)[0] ?? "";
    expect(hallMembers(plant, hall).reduce((sum, [, n]) => sum + n, 0)).toBe(
      plant.instances.filter((instance) => hallOf(instance.id) === hall).length,
    );
    expect(hallMembers(plant, "nowhere")).toEqual([]);
  });

  it("counts areas in plant order, matching the outline, not by count", () => {
    const plant = parsePlant({
      id: "plant",
      site: {
        name: "Site",
        areas: [{ name: "Site" }, { name: "Hall 1", parent: "Site" }, { name: "Utilities", parent: "Site" }],
      },
      instances: [
        { kind: "meter", names: ["U-1", "U-2", "U-3"], area: "Utilities" },
        { kind: "pump", names: ["P-1"], area: "Hall 1" },
        { kind: "pump", names: ["P-2"], area: "Yard" },
      ],
    });
    expect(areaCounts(plant)).toEqual([
      ["Hall 1", 1],
      ["Utilities", 3],
      ["Yard", 1],
    ]);
    expect(areaCounts(parsePlant(domain))).toEqual([
      ["Cooling", 2],
      ["Power", 1],
    ]);
  });
});

describe("node id lines", () => {
  it("adds a header line only once an id is longer than one line holds", () => {
    expect(idExtra("CH-1")).toBe(0);
    expect(idExtra("a".repeat(nodeSize.idChars))).toBe(0);
    expect(idExtra("a".repeat(nodeSize.idChars + 1))).toBe(nodeSize.idLine);
  });

  it("makes a node taller by one header line when its id wraps", () => {
    const longer = JSON.parse(JSON.stringify({ domain, composition }).replaceAll("CH-1", "CHILLER-ONE-LONG")) as {
      domain: typeof domain;
      composition: typeof composition;
    };
    const view = plantView(parsePlant(longer.domain, longer.composition), specs, null);
    const box = view.layout.nodes.get("CHILLER-ONE-LONG");
    expect(box?.height).toBe(headerHeight("CHILLER-ONE-LONG") + nodeSize.padding + 2 * nodeSize.pin);
  });
});

describe("review note lines", () => {
  it("makes a review node taller by the lines its fault and effector text wraps to, and no other node", () => {
    const plant = parsePlant(domain, composition);
    const plain = plantView(plant, specs, null, reviewView);
    const noted = plantView(plant, specs, null, reviewView, undefined, new Map([["CH-1", 3]]));
    const base = plain.layout.nodes.get("CH-1")?.height ?? 0;
    expect(noted.layout.nodes.get("CH-1")?.height).toBe(base + 2 * nodeSize.note);
    expect(noted.layout.nodes.get("CT-1")?.height).toBe(plain.layout.nodes.get("CT-1")?.height);
    expect(noted.nodes.find((node) => node.id === "CH-1")?.noteLines).toBe(3);
    expect(
      plantView(plant, specs, null, plainView, undefined, new Map([["CH-1", 3]])).layout.nodes.get("CH-1")?.height,
    ).toBe(headerHeight("CH-1") + nodeSize.padding + 2 * nodeSize.pin);
  });
});
