import { describe, expect, it } from "vitest";
import type { Workspace } from "./bridge";
import { sampleWorkspace } from "../test/fakeBridge";
import { durationOf, plantsOf, runRows, whenOf, workspaceEmpty, workspaceSummary } from "./workspace";

const run = (worldId: string, plant: string, recordedAt: string) => ({
  worldId,
  plant,
  recordedAt,
  ended: true,
  explore: false,
  durationS: 60,
});

const lab: Workspace = {
  name: "lab",
  domains: [
    { path: "domains/factory.domain.json", id: "factory", title: "Factory" },
    { path: "domains/legacy.domain.json", id: "", title: "" },
  ],
  compositions: [
    { path: "compositions/factory.composition.json", id: "factory", title: "Factory" },
    { path: "compositions/draft.composition.json", id: "draft", title: "Draft plant" },
  ],
  scenarios: [],
  runs: [
    run("w-2", "factory", "2026-10-14T10:00:00Z"),
    run("w-10", "factory", "2026-10-14T12:00:00Z"),
    run("w-9", "gone", "2026-10-14T12:00:00Z"),
  ],
};

describe("workspace", () => {
  it("makes one plant of a domain and its composition, with its runs", () => {
    expect(plantsOf(lab)).toEqual([
      {
        id: "draft",
        title: "Draft plant",
        domain: null,
        composition: "compositions/draft.composition.json",
        runs: 0,
        lastRun: "",
      },
      {
        id: "factory",
        title: "Factory",
        domain: "domains/factory.domain.json",
        composition: "compositions/factory.composition.json",
        runs: 2,
        lastRun: "2026-10-14T12:00:00Z",
      },
      { id: "legacy", title: "legacy", domain: "domains/legacy.domain.json", composition: null, runs: 0, lastRun: "" },
    ]);
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

  it("says when and how long in words", () => {
    expect(whenOf("2026-10-14T12:05:00Z")).toMatch(/^1[45] Oct 2026, \d\d:05$/);
    expect(whenOf("")).toBe("");
    expect(whenOf("nonsense")).toBe("");
    expect([durationOf(0), durationOf(45), durationOf(1800), durationOf(5400), durationOf(86400)]).toEqual([
      "",
      "45 s",
      "30 min",
      "1 h 30 min",
      "24 h",
    ]);
  });
});
