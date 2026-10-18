import { describe, expect, it } from "vitest";
import { originWords, plannedSimulated, unknownOrigin } from "./origin";
import type { InstanceOrigin } from "./origin";

const real: InstanceOrigin = { instance: "zone-01", kind: "real", board: "uno-07", device: "zone-01", bootId: "61" };

describe("instance origin", () => {
  it("names a real board with its boot and a simulated one by its board", () => {
    expect(originWords(real)).toBe("Real · board uno-07 · boot 61");
    expect(originWords({ ...real, bootId: "" })).toBe("Real · board uno-07");
    expect(originWords({ ...real, kind: "simulated", board: "mega-01" })).toBe("Simulated · board mega-01");
  });

  it("says an origin the run did not record is unknown, never simulated", () => {
    expect(unknownOrigin("fan-01")).toEqual({ instance: "fan-01", kind: "unknown", board: "", device: "", bootId: "" });
    expect(originWords(unknownOrigin("fan-01"))).toBe("Origin unknown: the run recorded no origin for it");
    expect(originWords(plannedSimulated("fan-01"))).toBe("Simulated · by the session plan");
  });
});
