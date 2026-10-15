import { describe, expect, it } from "vitest";
import { parsePlant } from "../../model/plant";
import type { ComponentSpec } from "../../model/ports";
import { plantView } from "../../model/plantView";
import { composition, domain } from "../../test/plantFixture";
import { flowEdges, rowPins } from "./flow";

describe("review wires on readout rows", () => {
  const specs = new Map<string, ComponentSpec>([
    [
      "lib/chiller",
      { ports: [{ name: "heat_out", type: "heat_rejection", direction: "out", channel: "heat" }], category: "cool" },
    ],
  ]);
  const shown = (channel: string) => ({
    channel,
    value: "1",
    unit: "",
    attention: null,
    word: false,
    outside: "",
  });
  const live = (channels: readonly string[]) => () => ({
    shown: channels.map(shown),
    status: null,
    count: 0,
    fault: "",
    effector: "",
  });

  it("ends a wire on the row of its port's channel when that row is shown", () => {
    const view = plantView(parsePlant(domain, composition), specs, null);
    const rows = rowPins(view, live(["heat"]), Number.POSITIVE_INFINITY);
    expect(flowEdges(view, false, null, rows).find((edge) => edge.source === "CH-1")).toMatchObject({
      sourceHandle: "heat_out",
    });
  });

  it("falls back to the side anchor when the port's row is not shown or is hidden at this tier", () => {
    const view = plantView(parsePlant(domain, composition), specs, null);
    const missing = rowPins(view, live(["power"]), Number.POSITIVE_INFINITY);
    const hidden = rowPins(view, live(["heat"]), 0);
    expect(flowEdges(view, false, null, missing).find((edge) => edge.source === "CH-1")).toMatchObject({
      sourceHandle: "out",
    });
    expect(flowEdges(view, false, null, hidden).find((edge) => edge.source === "CH-1")).toMatchObject({
      sourceHandle: "out",
    });
  });
});
