import { describe, expect, it } from "vitest";
import { lengthFacts } from "./length";
import type { RunEntry } from "./types";

const finished: RunEntry = {
  worldId: "w",
  ended: true,
  complete: true,
  running: false,
  incompleteReason: "",
  explore: false,
  seed: 7,
  plant: "p",
  durationS: 1800,
  recordedAt: "",
  faults: 0,
};

const stopped = (durationS: number, requestedS: number): RunEntry => ({
  ...finished,
  durationS,
  requestedS,
  terminationReason: "stopped_by_operator",
});

describe("how a run row states its length", () => {
  it.each([
    ["a run with no kept plan", finished, ["30 min"]],
    [
      "a run that completed its requested length",
      { ...finished, requestedS: 1800, terminationReason: "completed" },
      ["30 min"],
    ],
    ["a run stopped in whole minutes of the same unit", stopped(540, 1800), ["9 of 30 min", "stopped"]],
    ["a run stopped within the first minute", stopped(45, 1800), ["45 s of 30 min", "stopped"]],
    ["a run stopped before any simulated time passed", stopped(0, 1800), ["0 s of 30 min", "stopped"]],
    ["a run stopped between hours and minutes", stopped(3900, 7200), ["1 h 5 min of 2 h", "stopped"]],
    ["a run stopped short of a multi-part length", stopped(600, 4500), ["10 min of 1 h 15 min", "stopped"]],
    ["a run stopped in hours of a longer length", stopped(7200, 14400), ["2 of 4 h", "stopped"]],
    [
      "a stop whose requested length was not kept",
      { ...finished, terminationReason: "stopped_by_operator" },
      ["30 min"],
    ],
    ["a still running run", { ...stopped(0, 1800), ended: false, running: true }, []],
  ] satisfies [string, RunEntry, string[]][])("%s", (_name, run, parts) => {
    expect(lengthFacts(run)).toEqual(parts);
  });
});
