import { describe, expect, it } from "vitest";
import type { PlantView } from "./plantView";
import { corridorGutter, routeView } from "./wireView";

type NodeRow = readonly [string, number, number, number, number, readonly string[], readonly string[]];
type FrameRow = readonly [string, number, number, number, number];

const nodeRows: readonly NodeRow[] = [
  [
    "CH-1",
    680,
    456.5,
    232,
    176,
    ["load_in", "flow_in", "cond_water_in", "tariff_in"],
    ["heat_out", "supply_out", "return_out", "power_out"],
  ],
  [
    "CH-2",
    1336,
    456.5,
    232,
    176,
    ["load_in", "flow_in", "cond_water_in", "tariff_in"],
    ["heat_out", "supply_out", "return_out", "power_out"],
  ],
  ["AHU-1", 24, 1168.9, 232, 152, ["chw_in", "outdoor_in", "shift_in"], ["load_out"]],
  ["CT-1", 1008, 456.5, 232, 152, ["heat_in", "wet_bulb_in", "cw_flow_in"], ["basin_out", "power_out"]],
  ["P-1", 24, 456.5, 232, 176, ["level_in"], ["flow_out", "cw_flow_out", "liquid_out", "power_out"]],
  ["P-2", 24, 656.5, 232, 176, ["level_in"], ["flow_out", "cw_flow_out", "liquid_out", "power_out"]],
  ["P-3", 24, 856.5, 232, 176, ["level_in"], ["flow_out", "cw_flow_out", "liquid_out", "power_out"]],
  ["VFM-1", 352, 456.5, 232, 128, ["load_in"], ["power_out"]],
  ["VFM-2", 352, 608.5, 232, 128, ["load_in"], ["power_out"]],
  ["VFM-3", 352, 760.5, 232, 128, ["load_in"], ["power_out"]],
  ["AHU-2", 24, 1457.4, 232, 152, ["chw_in", "outdoor_in", "shift_in"], ["load_out"]],
  ["W-1", 24, 192, 232, 128, [], ["dry_bulb_out", "wet_bulb_out"]],
  ["SC-1", 24, 40, 232, 128, [], ["shift_out", "tariff_out"]],
  ["FM-1", 1008, 632.5, 232, 128, ["loads_in"], ["power_out"]],
  [
    "CH-3",
    680,
    656.5,
    232,
    176,
    ["load_in", "flow_in", "cond_water_in", "tariff_in"],
    ["heat_out", "supply_out", "return_out", "power_out"],
  ],
];

const frameRows: readonly FrameRow[] = [
  ["Site", 0, 0, 280, 344],
  ["Plant room", 0, 416.5, 1592, 640],
  ["Floors 1-3", 0, 1128.9, 280, 216],
  ["Floors 4 to 6", 0, 1417.4, 280, 216],
];

const edgeIds: readonly string[] = [
  "W-1.dry_bulb_out>AHU-1.outdoor_in",
  "W-1.dry_bulb_out>AHU-2.outdoor_in",
  "W-1.wet_bulb_out>CT-1.wet_bulb_in",
  "SC-1.shift_out>AHU-1.shift_in",
  "SC-1.shift_out>AHU-2.shift_in",
  "SC-1.tariff_out>CH-1.tariff_in",
  "CH-1.supply_out>AHU-1.chw_in",
  "CH-1.heat_out>CT-1.heat_in",
  "CH-1.power_out>FM-1.loads_in",
  "CH-2.supply_out>AHU-2.chw_in",
  "CT-1.basin_out>CH-1.cond_water_in",
  "CT-1.basin_out>CH-2.cond_water_in",
  "P-1.flow_out>CH-1.flow_in",
  "P-1.power_out>FM-1.loads_in",
  "VFM-1.power_out>FM-1.loads_in",
  "P-2.flow_out>CH-2.flow_in",
  "P-2.power_out>FM-1.loads_in",
  "P-3.power_out>FM-1.loads_in",
  "AHU-1.load_out>CH-1.load_in",
  "AHU-2.load_out>CH-2.load_in",
  "VFM-2.power_out>FM-1.loads_in",
  "VFM-3.power_out>FM-1.loads_in",
  "P-3.flow_out>CH-3.flow_in",
];

function officeTower(): PlantView {
  const pin = (name: string) => ({ name });
  const nodes = nodeRows.map(([id, , , , , ins, outs]) => ({
    id,
    label: id,
    noteLines: 1,
    pins: { ins: ins.map(pin), outs: outs.map(pin) },
  }));
  const edges = edgeIds.map((id) => {
    const [from = "", rest = ""] = id.split(">");
    const [fromId = "", fromHandle = ""] = from.split(".");
    const [toId = "", toHandle = ""] = rest.split(".");
    return { id, from: fromId, to: toId, fromHandle, toHandle, family: null, count: 1 };
  });
  const box = (x: number, y: number, width: number, height: number) => ({ x, y, width, height });
  return {
    nodes,
    edges,
    layout: {
      nodes: new Map(nodeRows.map(([id, x, y, w, h]) => [id, box(x, y, w, h)] as const)),
      groups: new Map(frameRows.map(([title, x, y, w, h]) => [title, box(x, y, w, h)] as const)),
    },
    halls: [],
  } as unknown as PlantView;
}

function within(
  points: readonly { x: number; y: number }[],
  rect: { x: number; y: number; width: number; height: number },
) {
  return points.every(
    (point) =>
      point.x >= rect.x - 1e-6 &&
      point.x <= rect.x + rect.width + 1e-6 &&
      point.y >= rect.y - 1e-6 &&
      point.y <= rect.y + rect.height + 1e-6,
  );
}

describe("office tower wires", () => {
  const view = officeTower();
  const boxOf = (id: string) => {
    const box = view.layout.nodes.get(id);
    if (box === undefined) throw new Error(`no node ${id}`);
    return box;
  };

  it("keeps every wire inside the rectangle of its source and target inflated by one gutter, at every tier", () => {
    for (const tier of ["full", "compact", "overview"] as const) {
      const routes = routeView(view, tier, false, 2);
      expect(routes.size).toBe(view.edges.length);
      for (const edge of view.edges) {
        const from = boxOf(edge.from);
        const to = boxOf(edge.to);
        const x = Math.min(from.x, to.x) - corridorGutter;
        const y = Math.min(from.y, to.y) - corridorGutter;
        const corridor = {
          x,
          y,
          width: Math.max(from.x + from.width, to.x + to.width) + corridorGutter - x,
          height: Math.max(from.y + from.height, to.y + to.height) + corridorGutter - y,
        };
        const route = routes.get(edge.id) ?? [];
        expect(route.length, `${edge.id} at ${tier}`).toBeGreaterThan(1);
        expect(within(route, corridor), `${edge.id} at ${tier}`).toBe(true);
      }
    }
  });
});
