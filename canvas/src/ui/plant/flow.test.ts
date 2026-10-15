import { MarkerType } from "@xyflow/react";
import { describe, expect, it } from "vitest";
import { parsePlant } from "../../model/plant";
import type { ComponentSpec } from "../../model/ports";
import { plantView } from "../../model/plantView";
import { composition, domain } from "../../test/plantFixture";
import { flowEdges, flowNodes } from "./flow";

describe("plant wires", () => {
  it("carries each routed polyline to the wire edge, on the generic anchors, always under the nodes", () => {
    const view = plantView(parsePlant(domain, composition), new Map(), null);
    const routes = new Map([[view.edges[0]?.id ?? "", [{ x: 1, y: 2 }]]]);
    const edges = flowEdges(view, routes, view.edges[0]?.from ?? null);
    expect(edges[0]).toMatchObject({
      type: "wire",
      sourceHandle: "out",
      targetHandle: "in",
      data: { points: [{ x: 1, y: 2 }] },
      zIndex: 0,
    });
    expect(edges[1]?.data).toEqual({ points: [] });
    expect(edges.every((edge) => edge.zIndex === 0)).toBe(true);
  });
});

describe("plant wire direction", () => {
  it("points each wire at its target and keeps selected edges emphasized", () => {
    const specs = new Map<string, ComponentSpec>([
      ["lib/chiller", { ports: [{ name: "heat_out", type: "heat_rejection", direction: "out" }], category: "cool" }],
    ]);
    const view = plantView(parsePlant(domain, composition), specs, null);
    const selected = flowEdges(view, new Map(), "CH-1").find((edge) => edge.source === "CH-1");
    expect(selected).toMatchObject({
      source: "CH-1",
      target: "CT-1",
      markerEnd: { type: MarkerType.ArrowClosed, width: 10, height: 10, color: "var(--port-heat)" },
      className: "wire family-heat wire-on",
      ariaLabel: "CH-1 to CT-1",
    });
  });

  it("dims unrelated target markers with their selected edges", () => {
    const view = plantView(parsePlant(domain, composition), new Map(), null);
    const [first] = view.edges;
    if (first === undefined) throw new Error("fixture has no links");
    const extended = {
      ...view,
      edges: [...view.edges, { ...first, id: "unrelated", from: "CT-1", to: "MB-1", family: "heat" as const }],
    };
    const unrelated = flowEdges(extended, new Map(), "CH-1").find((edge) => edge.id === "unrelated");
    expect(unrelated).toMatchObject({
      className: "wire family-heat wire-off",
      markerEnd: {
        type: MarkerType.ArrowClosed,
        color: "color-mix(in srgb, var(--text-tertiary) 30%, transparent)",
      },
    });
  });
});

describe("plant frame titles", () => {
  it("draws each area title above the wires, in its own node", () => {
    const view = plantView(parsePlant(domain, composition), new Map(), null);
    const titles = flowNodes(view, null).filter((node) => node.type === "frameTitle");
    expect(titles.map((node) => [node.id, node.zIndex])).toEqual(
      [...view.layout.groups.keys()].map((group) => [`title:${group}`, 2]),
    );
    expect(
      flowNodes(view, null)
        .filter((node) => node.type === "frame")
        .every((node) => node.zIndex === -1),
    ).toBe(true);
  });
});
