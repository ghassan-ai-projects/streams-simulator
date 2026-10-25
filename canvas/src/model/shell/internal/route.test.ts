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
      { screen: "review", worldId: "w-1", model: true },
      { screen: "edit", composition: "compositions/lab.composition.json" },
      { screen: "device", path: "kinds/fan.kind.json" },
      { screen: "device", path: "kinds/fan.kind.json", step: "behaviour" },
      { screen: "capture" },
    ] as const;
    for (const route of routes) {
      expect(parseRoute(formatRoute(route))).toEqual(route);
    }
  });

  it("opens a device on its Behaviour step only when asked, and ignores a step it does not know", () => {
    expect(formatRoute({ screen: "device", path: "kinds/fan.kind.json", step: "behaviour" })).toBe(
      "#/device?device=kinds%2Ffan.kind.json&step=behaviour",
    );
    expect(parseRoute("#/device?device=kinds%2Ffan.kind.json&step=check")).toEqual({
      screen: "device",
      path: "kinds/fan.kind.json",
    });
  });

  it("falls back to the workspace for anything it does not know", () => {
    expect(parseRoute("")).toEqual({ screen: "workspace" });
    expect(parseRoute("#/plant")).toEqual({ screen: "workspace" });
    expect(parseRoute("#/edit")).toEqual({ screen: "workspace" });
    expect(parseRoute("#/device")).toEqual({ screen: "workspace" });
    expect(parseRoute("#/elsewhere?x=1")).toEqual({ screen: "workspace" });
  });
});
