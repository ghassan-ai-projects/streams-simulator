import { describe, expect, it } from "vitest";
import { formatRoute, parseRoute } from "./route";

describe("routes", () => {
  it("round-trips every screen", () => {
    const routes = [
      { screen: "home" },
      { screen: "plant", domain: "domains/a b.domain.json", hall: null },
      { screen: "plant", domain: "domains/c.domain.json", hall: "h07" },
      { screen: "run", worldId: "w-1" },
    ] as const;
    for (const route of routes) {
      expect(parseRoute(formatRoute(route))).toEqual(route);
    }
  });

  it("falls back to home for anything it does not know", () => {
    expect(parseRoute("")).toEqual({ screen: "home" });
    expect(parseRoute("#/plant")).toEqual({ screen: "home" });
    expect(parseRoute("#/elsewhere?x=1")).toEqual({ screen: "home" });
  });
});
