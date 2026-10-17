import { describe, expect, it } from "vitest";
import type { Workspace } from "./types";
import { sampleWorkspace } from "../../../test/workspaceFixture";
import {
  outcomeWords,
  plantFacts,
  plantsOf,
  runFacts,
  runRows,
  suggestedId,
  whenOf,
  workspaceEmpty,
  workspaceSummary,
} from "./workspace";

const run = (worldId: string, plant: string, recordedAt: string) => ({
  worldId,
  plant,
  recordedAt,
  ended: true,
  complete: true,
  incompleteReason: "",
  explore: false,
  seed: null,
  durationS: 60,
  faults: 0,
});

const lab: Workspace = {
  name: "lab",
  domains: [
    { path: "domains/factory.domain.json", id: "factory", title: "Factory", instances: 23 },
    { path: "domains/legacy.domain.json", id: "", title: "", instances: 0 },
  ],
  compositions: [
    { path: "compositions/factory.composition.json", id: "factory", title: "Factory", instances: 0 },
    { path: "compositions/draft.composition.json", id: "draft", title: "Draft plant", instances: 0 },
  ],
  scenarios: [],
  runs: [
    run("w-2", "factory", "2026-10-14T10:00:00Z"),
    run("w-10", "factory", "2026-10-14T12:00:00Z"),
    run("w-9", "gone", "2026-10-14T12:00:00Z"),
  ],
};

describe("workspace catalogue", () => {
  it("makes one plant of a domain and its composition, with its runs", () => {
    expect(plantsOf(lab)).toEqual([
      {
        id: "draft",
        title: "Draft plant",
        domain: null,
        composition: "compositions/draft.composition.json",
        instances: 0,
        runs: 0,
        lastRun: "",
      },
      {
        id: "factory",
        title: "Factory",
        domain: "domains/factory.domain.json",
        composition: "compositions/factory.composition.json",
        instances: 23,
        runs: 2,
        lastRun: "2026-10-14T12:00:00Z",
      },
      {
        id: "legacy",
        title: "legacy",
        domain: "domains/legacy.domain.json",
        composition: null,
        instances: 0,
        runs: 0,
        lastRun: "",
      },
    ]);
    const [draft, factory, legacy] = plantsOf(lab);
    expect(draft && plantFacts(draft)).toBe("Not runnable yet. Edit it, then save.");
    expect(factory && plantFacts(factory)).toMatch(/^23 instances · 2 runs · last 1[45] Oct 2026/);
    expect(legacy && plantFacts(legacy)).toBe("No runs yet");
    expect(factory && plantFacts({ ...factory, instances: 1, runs: 0, lastRun: "" })).toBe("1 instance · no runs yet");
  });

  it("lists runs newest first with their plant's title, and filters by plant", () => {
    expect(runRows(lab, null).map((row) => [row.run.worldId, row.plantTitle])).toEqual([
      ["w-10", "Factory"],
      ["w-9", "gone"],
      ["w-2", "Factory"],
    ]);
    expect(runRows(lab, "factory").map((row) => row.run.worldId)).toEqual(["w-10", "w-2"]);
  });

  it("summarises and recognises an empty workspace", () => {
    expect(workspaceSummary(sampleWorkspace)).toBe("1 plant · 2 runs");
    expect(workspaceEmpty(sampleWorkspace)).toBe(false);
    expect(workspaceEmpty({ name: "x", domains: [], compositions: [], scenarios: [], runs: [] })).toBe(true);
  });

  it("says when in words", () => {
    expect(whenOf("2026-10-14T12:05:00Z")).toMatch(/^1[45] Oct 2026, \d\d:05$/);
    expect(whenOf("")).toBe("");
    expect(whenOf("nonsense")).toBe("");
  });

  it("suggests a free id from a plant's name", () => {
    expect(suggestedId("Cooling Hall 2!", [])).toBe("cooling-hall-2");
    expect(suggestedId("Cooling hall", ["cooling-hall", "cooling-hall-2"])).toBe("cooling-hall-3");
    expect(suggestedId("42 AB", [])).toBe("plant-ab");
    expect(suggestedId("Kühlhalle", [])).toBe("kuhlhalle");
  });

  it("lists the seed and the outcome of a run in its facts", () => {
    const finished = {
      worldId: "w",
      ended: true,
      complete: true,
      incompleteReason: "",
      explore: true,
      seed: 7,
      plant: "p",
      durationS: 1800,
      recordedAt: "",
      faults: 0,
    };
    expect(runFacts(finished)).toEqual(["30 min", "seed 7", "no faults"]);
    expect(runFacts({ ...finished, seed: null })).toEqual(["30 min", "no faults"]);
    const abandoned = { ...finished, complete: false, incompleteReason: "run was stopped", faults: 2 };
    expect(runFacts(abandoned)).toEqual(["30 min", "seed 7", "Did not finish: run was stopped"]);
    expect(outcomeWords({ ...abandoned, incompleteReason: "" })).toBe("Did not finish: the run did not close.");
    expect(outcomeWords(finished)).toBe("");
    expect(outcomeWords({ ...abandoned, ended: false })).toBe("");
  });

  it("says how many faults a finished run injected", () => {
    const finished = {
      worldId: "w",
      ended: true,
      complete: true,
      incompleteReason: "",
      explore: true,
      seed: null,
      plant: "p",
      durationS: 60,
      recordedAt: "",
    };
    const faultsOf = (run: typeof finished & { faults: number }) =>
      runFacts(run).filter((part) => part.includes("fault"));
    expect(faultsOf({ ...finished, faults: 1 })).toEqual(["1 fault"]);
    expect(faultsOf({ ...finished, faults: 0 })).toEqual(["no faults"]);
    expect(faultsOf({ ...finished, ended: false, faults: 3 })).toEqual([]);
  });
});
