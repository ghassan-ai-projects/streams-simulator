import { describe, expect, it } from "vitest";
import {
  claimSummary,
  instrumentAssumptions,
  notReproduced,
  partialWords,
  reproduced,
  standInWords,
  suggestionNote,
} from "./twinClaims";

const thermistor = {
  ref: "streamsim/thermistor@1",
  title: "Thermistor",
  unit: "Cel",
  min: -20,
  max: 120,
  match: "exact" as const,
  fit: null,
  words: "",
};

describe("what a twin reproduces", () => {
  it("lists what the binary carries and what it leaves out, and calls the response a stand-in", () => {
    expect(reproduced.join(" ")).toMatch(/instrument/);
    expect(reproduced.join(" ")).toMatch(/none, direct or delayed/);
    expect(notReproduced.join(" ")).toMatch(/firmware/);
    expect(notReproduced.join(" ")).toMatch(/Saturation/);
    expect(standInWords).toBe("The response is a simple stand-in until it is compared with recorded field data.");
  });
});

describe("which claims passed", () => {
  it("states each check on its own and never claims calibration or a board comparison", () => {
    const base = { interface: "passed", plausibility: "passed", warnings: 0, errors: 0 } as const;
    expect(claimSummary(base)).toBe(
      "Interface check passed; plausibility check passed; not calibrated; not compared with a board.",
    );
    expect(claimSummary({ ...base, plausibility: "warnings", warnings: 2 })).toBe(
      "Interface check passed; plausibility check passed with 2 warnings; not calibrated; not compared with a board.",
    );
    expect(claimSummary({ ...base, interface: "failed", plausibility: "unavailable" })).toBe(
      "Interface check failed; plausibility check not run; not calibrated; not compared with a board.",
    );
    expect(claimSummary({ ...base, interface: "running", plausibility: "failed", errors: 1 })).toBe(
      "Interface check running; plausibility check found 1 error; not calibrated; not compared with a board.",
    );
    expect(claimSummary({ ...base, plausibility: "running" })).toMatch(/plausibility check running/);
  });
});

describe("a partial twin", () => {
  it("names every excluded device with its reason", () => {
    expect(partialWords([{ name: "indicator", reason: "the setting scan_class is not reproduced by a twin" }])).toBe(
      "This twin excludes indicator (the setting scan_class is not reproduced by a twin).",
    );
    expect(
      partialWords([
        { name: "a", reason: "x" },
        { name: "b", reason: "y" },
      ]),
    ).toBe("This twin excludes a (x), b (y).");
    expect(partialWords([])).toBe("");
  });
});

describe("suggested choices", () => {
  it("asks for a review without blocking, in the singular and the plural", () => {
    expect(suggestionNote(0)).toBeNull();
    expect(suggestionNote(1)).toBe(
      "1 choice is still the suggested default: review the instrument and response of each reading.",
    );
    expect(suggestionNote(3)).toBe(
      "3 choices are still the suggested defaults: review the instrument and response of each reading.",
    );
  });

  it("summarises only what the library states about the chosen instrument", () => {
    expect(instrumentAssumptions(thermistor)).toBe("On record: Cel, -20 to 120.");
    expect(instrumentAssumptions({ ...thermistor, match: "convertible", unit: "K" })).toBe(
      "On record: K, -20 to 120, measured in K, converted.",
    );
    expect(instrumentAssumptions(undefined)).toBe("Nothing is added to the true value.");
  });
});
