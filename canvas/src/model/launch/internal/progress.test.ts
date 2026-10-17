import { describe, expect, it } from "vitest";
import { runProgressWords } from "./progress";

describe("runProgressWords", () => {
  it("says how far the run is and how fast it goes", () => {
    const progress = { atS: 60, durationS: 3600, rate: 59.6 };
    expect(runProgressWords(progress)).toBe("00:01:00 of 01:00:00 · 60 s per s");
    expect(runProgressWords(progress, " simulated")).toBe("00:01:00 of 01:00:00 simulated · 60 s per s");
  });

  it("leaves the speed out when the rate is not known", () => {
    expect(runProgressWords({ atS: 0, durationS: 120, rate: null })).toBe("00:00:00 of 00:02:00");
  });
});
