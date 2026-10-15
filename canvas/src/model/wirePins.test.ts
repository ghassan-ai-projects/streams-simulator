import { describe, expect, it } from "vitest";
import { behindNodes, notEnteringAtPin, viewOf } from "../test/wireAudit";
import { huntSpecialNames, officeTower } from "../test/wireLayouts";
import { routeView } from "./wireView";

const plants = [
  ["a hunt plant with aligned stacked nodes", huntSpecialNames],
  ["the office tower", officeTower],
] as const;

describe("wires end on their pins", () => {
  it.each(plants)("%s: every wire enters its target on the target's pin side, at every tier", (_, dump) => {
    const view = viewOf(dump);
    for (const [tier, review, pixel] of [
      ["full", false, 1],
      ["compact", false, 2],
      ["compact", true, 2],
      ["overview", false, 3],
    ] as const) {
      const routes = routeView(view, tier, review, pixel);
      expect(notEnteringAtPin(view, routes), `${tier} ${String(pixel)}`).toEqual([]);
    }
  });

  it.each(plants)("%s: no wire runs behind a node", (_, dump) => {
    const view = viewOf(dump);
    for (const [tier, pixel] of [
      ["full", 1],
      ["compact", 2],
      ["overview", 3],
    ] as const) {
      expect(behindNodes(view, routeView(view, tier, false, pixel)), `${tier} ${String(pixel)}`).toEqual([]);
    }
  });
});
