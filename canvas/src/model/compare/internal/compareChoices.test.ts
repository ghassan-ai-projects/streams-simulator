import { describe, expect, it } from "vitest";
import { referenceChoices, referenceRole, thisRunRole } from "./compareChoices";

const run = (worldId: string, extra: Record<string, unknown> = {}) => ({
  worldId,
  ended: true,
  complete: true,
  running: false,
  explore: true,
  seed: 11,
  durationS: 1800,
  plant: "factory",
  recordedAt: "2026-10-24T10:00:00Z",
  ...extra,
});

describe("the runs that can be the reference", () => {
  it("names the two roles", () => {
    expect([referenceRole, thisRunRole]).toEqual(["Reference (A)", "This run (B)"]);
  });

  it("lists the other ended runs of the same plant, in the order given", () => {
    const runs = [
      run("w-3"),
      run("w-2", { plant: "chiller" }),
      run("w-1"),
      run("w-4", { ended: false, running: true }),
    ];
    expect(referenceChoices(runs, { current: "w-3", plant: "factory" }).map((each) => each.worldId)).toEqual(["w-1"]);
  });

  it("describes a run by its name or id, mode, seed, length and how it ended", () => {
    const choices = referenceChoices(
      [
        run("w-1", { name: "Baseline", explore: false }),
        run("w-2", { requestedS: 1800, durationS: 540, terminationReason: "stopped_by_operator" }),
        run("w-5", { complete: false, incompleteReason: "killed" }),
      ],
      { current: "w-9", plant: "factory" },
    );
    expect(choices).toEqual([
      { worldId: "w-1", title: "Baseline", facts: "w-1 · Evaluate · seed 11 · 30 min · complete" },
      { worldId: "w-2", title: "w-2", facts: "Explore · seed 11 · 9 min of 30 min · stopped by operator" },
      { worldId: "w-5", title: "w-5", facts: "Explore · seed 11 · 30 min · did not finish" },
    ]);
  });

  it("offers nothing when no other run of the plant has ended", () => {
    expect(referenceChoices([run("w-1")], { current: "w-1", plant: "factory" })).toEqual([]);
  });
});
