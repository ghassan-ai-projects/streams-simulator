import { describe, expect, it } from "vitest";
import { withOrigin } from "./origins";

const real = { instance: "zone-01", kind: "real", board: "uno-07", device: "zone-01", bootId: "61" } as const;
const simulated = { instance: "fan-01", kind: "simulated", board: "mega-01", device: "fan-01", bootId: "b-1" } as const;

describe("origins of a run", () => {
  it("keeps the latest origin of each instance and ignores a device that is no instance", () => {
    const first = withOrigin(new Map(), real);
    const rebooted = withOrigin(withOrigin(first, simulated), { ...real, bootId: "62" });
    expect(rebooted.get("zone-01")?.bootId).toBe("62");
    expect(rebooted.size).toBe(2);
    expect(withOrigin(first, { ...real, instance: "" })).toBe(first);
  });
});
