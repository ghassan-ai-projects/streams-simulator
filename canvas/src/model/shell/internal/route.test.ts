import { describe, expect, it } from "vitest";
import { formatRoute, parseRoute } from "./route";

describe("routes", () => {
  it("round-trips every screen", () => {
    const routes = [
      { screen: "workspace" },
      { screen: "plant", domain: "domains/a b.domain.json", hall: null },
      { screen: "plant", domain: "domains/c.domain.json", hall: "h07" },
      { screen: "plant", domain: "domains/c.domain.json", hall: null, launch: true },
      { screen: "review", worldId: "w-1" },
      { screen: "edit", composition: "compositions/lab.composition.json" },
      { screen: "device", path: "kinds/fan.kind.json" },
      { screen: "capture" },
    ] as const;
    for (const route of routes) {
      expect(parseRoute(formatRoute(route))).toEqual(route);
    }
  });

  it("falls back to the workspace for anything it does not know", () => {
    expect(parseRoute("")).toEqual({ screen: "workspace" });
    expect(parseRoute("#/plant")).toEqual({ screen: "workspace" });
    expect(parseRoute("#/edit")).toEqual({ screen: "workspace" });
    expect(parseRoute("#/device")).toEqual({ screen: "workspace" });
    expect(parseRoute("#/elsewhere?x=1")).toEqual({ screen: "workspace" });
  });
});
