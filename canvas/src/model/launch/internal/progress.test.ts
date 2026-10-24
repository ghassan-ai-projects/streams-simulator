import { describe, expect, it } from "vitest";
import { runProgressWords, runStageWords, runTitle } from "./progress";

describe("run progress words", () => {
  it("says how far the run is and how fast it goes", () => {
    const progress = { atS: 60, durationS: 3600, rate: 59.6, paused: false };
    expect(runProgressWords(progress)).toBe("00:01:00 of 01:00:00 · 60 s per s");
    expect(runProgressWords(progress, " simulated")).toBe("00:01:00 of 01:00:00 simulated · 60 s per s");
  });

  it("leaves the speed out when the rate is not known", () => {
    expect(runProgressWords({ atS: 0, durationS: 120, rate: null, paused: false })).toBe("00:00:00 of 00:02:00");
  });

  it("says the run is paused where it stands and names no speed, since the clock is not moving", () => {
    const paused = { atS: 720, durationS: 1800, rate: 60, paused: true };
    expect(runStageWords(paused)).toBe("Paused");
    expect(runStageWords({ paused: false })).toBe("Running");
    expect(runProgressWords(paused, " simulated")).toBe("00:12:00 of 00:30:00 simulated");
  });
});

describe("the title of a run", () => {
  it("is the plant until the run is named, then the name on the plant", () => {
    expect(runTitle({}, "Mini plant")).toBe("Mini plant");
    expect(runTitle({ name: "Cooling loss" }, "Mini plant")).toBe("Cooling loss on Mini plant");
  });
});
