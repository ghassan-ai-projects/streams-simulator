import { describe, expect, it } from "vitest";
import { bindingLine, connectionWords, shownOrigins, withOrigin } from "./origins";

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

describe("origins shown for a connected run", () => {
  const recorded = new Map([[real.instance, real]]);

  it("marks every instance without an origin record unknown in a connected run, and adds nothing otherwise", () => {
    const shown = shownOrigins(recorded, ["zone-01", "fan-01"], "saved");
    expect(shown.get("zone-01")).toBe(real);
    expect(shown.get("fan-01")).toEqual({ instance: "fan-01", kind: "unknown", board: "", device: "", bootId: "" });
    expect(shownOrigins(recorded, ["zone-01", "fan-01"], "none")).toBe(recorded);
    expect(shownOrigins(recorded, ["zone-01", "fan-01"], "saved", ["fan-01"]).get("fan-01")?.kind).toBe("simulated");
  });

  it("names a connected run and its gateway binding", () => {
    expect(connectionWords("running")).toBe("Connected session · running");
    expect(connectionWords("saved")).toBe("Connected session");
    expect(bindingLine({ matched: false, refusal: "inventory_mismatch", reason: "", issues: [] })).toContain(
      "The gateway was refused (inventory_mismatch)",
    );
  });
});
