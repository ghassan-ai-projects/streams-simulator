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

type Routes = ReturnType<typeof routeView>;

function lastY(routes: Routes, id: string): number {
  return routes.get(id)?.at(-1)?.y ?? 0;
}

function firstY(routes: Routes, id: string): number {
  return routes.get(id)?.[0]?.y ?? 0;
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

  it("moves a review wire down with the extra note lines of its node", () => {
    const plant = parsePlant(domain, composition);
    const view = plantView(plant, specs, null, true, undefined, new Map([["CH-1", 3]]));
    const route = routeView(view, "full", true).get("CH-1.heat_out>CT-1.heat_in") ?? [];
    const top = boxOf(view, "CH-1").y + headerHeight("CH-1") + 3 * nodeSize.note + 1;
    expect(route[0]?.y).toBe(top + (2 + 0.5) * nodeSize.pin);
  });

  it("spaces wire ends that share one input row by 6 screen pixels and keeps one source port's wires together", () => {
    const view = viewOf(false);
    const first = view.edges[0];
    if (first === undefined) throw new Error("fixture has no links");
    const twin = { ...first, id: `${first.id}-twin` };
    const pixel = 2.5;
    const routes = routeView({ ...view, edges: [first, twin] }, "full", false, pixel);
    expect(Math.abs(lastY(routes, first.id) - lastY(routes, twin.id))).toBeGreaterThanOrEqual(6 * pixel);
    expect(routes.get(first.id)?.[0]).toEqual(routes.get(twin.id)?.[0]);
  });

  it("spaces the starts of wires from different ports that meet in the middle of one node", () => {
    const view = viewOf(false);
    const first = view.edges[0];
    if (first === undefined) throw new Error("fixture has no links");
    const other = { ...first, id: "other", fromHandle: "other_out" };
    const pixel = 4;
    const routes = routeView({ ...view, edges: [first, other] }, "overview", false, pixel);
    expect(Math.abs(firstY(routes, first.id) - firstY(routes, "other"))).toBeGreaterThanOrEqual(6 * pixel);
  });

  it("keeps the ends on neighbouring input rows apart too, so a crowded row never runs into the next", () => {
    const view = viewOf(false);
    const first = view.edges[0];
    const target = view.nodes.find((node) => node.id === "CH-1");
    const [one, two] = target?.pins.ins ?? [];
    if (first === undefined || one === undefined || two === undefined) throw new Error("fixture incomplete");
    const edges = [
      ...[0, 1, 2, 3].map((index) => ({
        ...first,
        id: `a${String(index)}`,
        from: "CT-1",
        to: "CH-1",
        toHandle: one.name,
      })),
      ...[0, 1].map((index) => ({ ...first, id: `b${String(index)}`, from: "CT-1", to: "CH-1", toHandle: two.name })),
    ];
    const pixel = 2;
    const routes = routeView({ ...view, edges }, "full", false, pixel);
    const ys = edges.map((edge) => lastY(routes, edge.id)).sort((a, b) => a - b);
    ys.slice(1).forEach((y, index) => {
      expect(y - (ys[index] ?? 0)).toBeGreaterThanOrEqual(6 * pixel - 1e-6);
    });
  });

  it("never spreads ends outside the node, shrinking the gap when many wires meet one row", () => {
    const view = viewOf(false);
    const first = view.edges[0];
    if (first === undefined) throw new Error("fixture has no links");
    const many = Array.from({ length: 12 }, (_, index) => ({ ...first, id: `${first.id}-${String(index)}` }));
    const routes = routeView({ ...view, edges: many }, "full", false, 4);
    const box = boxOf(view, first.to);
    const ys = many.map((edge) => lastY(routes, edge.id));
    expect(Math.min(...ys)).toBeGreaterThanOrEqual(box.y);
    expect(Math.max(...ys)).toBeLessThanOrEqual(box.y + box.height);
    expect(new Set(ys).size).toBe(12);
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

  it("boxes a title's one header line: inside the frame at full size, above it and full width when zoomed out", () => {
    const frame = { x: 10, y: 100, width: 300, height: 200 };
    expect(titleBox(frame, "full", 1)).toEqual({ x: 10, y: 106, width: 300, height: 16 });
    const one = titleBox(frame, "overview", 2);
    expect(one).toEqual({ x: 10, y: 100 - (8 + 34), width: 300, height: 34 });
    const capped = titleBox(frame, "map", 20);
    expect(capped.width).toBe(300);
    expect(capped.height).toBeCloseTo(17 * (30 / 13), 6);
  });

  it("joins two nodes of one column from the right edge to the target's pin on its left edge, outside both", () => {
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
    expect(route[0]?.x).toBe(upper.x + upper.width);
    expect(route.at(-1)?.x).toBe(lower.x);
    expect(route.at(-2)?.y).toBe(route.at(-1)?.y);
    expect(route.at(-1)?.y).toBeGreaterThan(lower.y);
    expect(route.at(-1)?.y).toBeLessThan(lower.y + lower.height);
    const inside = (point: { x: number; y: number }, box: Rect) =>
      point.x > box.x && point.x < box.x + box.width && point.y > box.y && point.y < box.y + box.height;
    expect(route.some((point) => inside(point, upper) || inside(point, lower))).toBe(false);
  });
});
