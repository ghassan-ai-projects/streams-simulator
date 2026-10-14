import { MarkerType } from "@xyflow/react";
import { describe, expect, it } from "vitest";
import { parsePlant } from "../../model/plant";
import type { ComponentSpec } from "../../model/ports";
import { plantView } from "../../model/plantView";
import { composition, domain } from "../../test/plantFixture";
import { flowEdges } from "./flow";

describe("plant wire direction", () => {
  it("points each wire at its target and keeps selected edges emphasized", () => {
    const specs = new Map<string, ComponentSpec>([
      ["lib/chiller", { ports: [{ name: "heat_out", type: "heat_rejection", direction: "out" }], category: "cool" }],
    ]);
    const view = plantView(parsePlant(domain, composition), specs, null);
    const selected = flowEdges(view, false, "CH-1").find((edge) => edge.source === "CH-1");
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
    const unrelated = flowEdges(extended, false, "CH-1").find((edge) => edge.id === "unrelated");
    expect(unrelated).toMatchObject({
      className: "wire family-heat wire-off",
      markerEnd: {
        type: MarkerType.ArrowClosed,
        color: "color-mix(in srgb, var(--text-tertiary) 30%, transparent)",
      },
    });
  });
});
