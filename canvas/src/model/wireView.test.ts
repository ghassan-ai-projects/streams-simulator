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

  it("keeps the title obstacle inside the frame at full size and above it when zoomed out", () => {
    const frame = { x: 10, y: 100, width: 300, height: 200 };
    expect(titleBox(frame, "full")).toEqual({ x: 10, y: 100, width: 300, height: 28 });
    expect(titleBox(frame, "compact").y + titleBox(frame, "compact").height).toBe(100);
  });
});
