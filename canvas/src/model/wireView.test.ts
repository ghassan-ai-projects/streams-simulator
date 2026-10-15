import { describe, expect, it } from "vitest";
import { chillerPorts, composition, domain, towerPorts } from "../test/plantFixture";
import { parsePlant } from "./plant";
import type { PlantView } from "./plantView";
import { headerHeight, nodeSize, plantView } from "./plantView";
import { parsePorts } from "./ports";
import type { Rect } from "./viewport";
import { routeView, titleBox, wireModeOf } from "./wireView";

const specs = new Map([
  ["lib/chiller", { ports: parsePorts(chillerPorts), category: "cool" as const }],
  ["lib/tower", { ports: parsePorts(towerPorts), category: "cool" as const }],
]);

function boxOf(view: PlantView, id: string): Rect {
  const box = view.layout.nodes.get(id);
  if (box === undefined) throw new Error(`no ${id}`);
  return box;
}

function viewOf(review: boolean) {
  return plantView(parsePlant(domain, composition), specs, null, review);
}

describe("wire view", () => {
  it("leaves each wire from its output pin row on the source's right edge and enters on the target's left edge", () => {
    const view = viewOf(false);
    const route = routeView(view, "full", false).get("CH-1.heat_out>CT-1.heat_in") ?? [];
    const box = boxOf(view, "CH-1");
    const pinY = box.y + headerHeight("CH-1") + 5 + 0.5 * nodeSize.pin;
    expect(route[0]).toEqual({ x: box.x + box.width, y: pinY });
    expect(route.at(-1)?.x).toBe(boxOf(view, "CT-1").x);
    expect(route.at(-2)?.y).toBe(route.at(-1)?.y);
  });

  it("puts a review wire on the row of its port, outputs after inputs, below the note row", () => {
    const view = viewOf(true);
    const route = routeView(view, "full", true).get("CH-1.heat_out>CT-1.heat_in") ?? [];
    const inputs = 2;
    const top = boxOf(view, "CH-1").y + headerHeight("CH-1") + nodeSize.note + 1;
    expect(route[0]?.y).toBe(top + (inputs + 0.5) * nodeSize.pin);
  });

  it("spaces wire ends that share one pin row by at least 6 px", () => {
    const view = viewOf(false);
    const first = view.edges[0];
    if (first === undefined) throw new Error("fixture has no links");
    const twin = { ...first, id: `${first.id}-twin` };
    const routes = routeView({ ...view, edges: [first, twin] }, "full", false);
    const starts = [routes.get(first.id)?.[0]?.y ?? 0, routes.get(twin.id)?.[0]?.y ?? 0];
    expect(Math.abs((starts[0] ?? 0) - (starts[1] ?? 0))).toBeGreaterThanOrEqual(6);
  });

  it("anchors on the middle of a node at the overview tiers, where pins are not drawn", () => {
    expect(wireModeOf("overview", false)).toBe("centre");
    expect(wireModeOf("compact", true)).toBe("rows");
    expect(wireModeOf("full", false)).toBe("pins");
    const view = viewOf(false);
    const route = routeView(view, "overview", false).get("CH-1.heat_out>CT-1.heat_in") ?? [];
    const box = boxOf(view, "CH-1");
    expect(Math.abs((route[0]?.y ?? 0) - (box.y + box.height / 2))).toBeLessThanOrEqual(6);
  });

  it("boxes a title's text: inside the frame at full size, above it, wrapped to the frame's width, when zoomed out", () => {
    const frame = { x: 10, y: 100, width: 300, height: 200 };
    expect(titleBox(frame, "full", 1)).toEqual({ x: 10, y: 106, width: 300, height: 16 });
    const one = titleBox(frame, "overview", 2, "Hall");
    expect(one).toEqual({ x: 10, y: 100 - (8 + 34), width: (4 * 6.3 + 4) * 2, height: 34 });
    const wrapped = titleBox(frame, "overview", 4, "Production hall A");
    expect(wrapped.height).toBe(17 * 2 * 4);
    expect(wrapped.width).toBe(300);
    expect(wrapped.y + wrapped.height).toBe(100 - 16);
  });

  it("joins two nodes of one column bottom to top, in a straight run through the row gap", () => {
    const view = viewOf(false);
    const upper = view.layout.nodes.get("CH-1");
    const first = view.edges[0];
    if (upper === undefined || first === undefined) throw new Error("fixture incomplete");
    const lower = { x: upper.x, y: upper.y + upper.height + 24, width: upper.width, height: 120 };
    const nodes = new Map([...view.layout.nodes, ["CT-1", lower] as const]);
    const stacked = {
      ...view,
      layout: { ...view.layout, nodes },
      edges: [{ ...first, id: "down", from: "CH-1", to: "CT-1" }],
    };
    const route = routeView(stacked, "full", false).get("down") ?? [];
    const xs = new Set(route.map((point) => point.x));
    expect(xs.size).toBe(1);
    expect(route[0]?.y).toBe(upper.y + upper.height);
    expect(route.at(-1)?.y).toBe(lower.y);
    const up = { ...stacked, edges: [{ ...first, id: "up", from: "CT-1", to: "CH-1" }] };
    const back = routeView(up, "full", false).get("up") ?? [];
    expect(back[0]?.y).toBe(lower.y);
    expect(back.at(-1)?.y).toBe(upper.y + upper.height);
  });

  it("spaces the two wires of a stacked pair that meet on one edge", () => {
    const view = viewOf(false);
    const upper = view.layout.nodes.get("CH-1");
    const first = view.edges[0];
    if (upper === undefined || first === undefined) throw new Error("fixture incomplete");
    const lower = { x: upper.x, y: upper.y + upper.height + 24, width: upper.width, height: 120 };
    const nodes = new Map([...view.layout.nodes, ["CT-1", lower] as const]);
    const both = [
      { ...first, id: "down", from: "CH-1", to: "CT-1" },
      { ...first, id: "up", from: "CT-1", to: "CH-1" },
    ];
    const routes = routeView({ ...view, layout: { ...view.layout, nodes }, edges: both }, "full", false);
    const downEnd = routes.get("down")?.at(-1);
    const upStart = routes.get("up")?.[0];
    expect(Math.abs((downEnd?.x ?? 0) - (upStart?.x ?? 0))).toBeGreaterThanOrEqual(6);
  });
});
