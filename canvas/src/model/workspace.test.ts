import { describe, expect, it } from "vitest";
import { sampleWorkspace } from "../test/fakeBridge";
import { workspaceEmpty, workspaceSummary } from "./workspace";

describe("workspace summary", () => {
  it("counts each kind of file in words", () => {
    expect(workspaceSummary(sampleWorkspace)).toBe("1 plant · 1 composition · 0 scenarios · 2 runs");
  });

  it("knows an empty workspace", () => {
    expect(workspaceEmpty(sampleWorkspace)).toBe(false);
    expect(workspaceEmpty({ name: "x", domains: [], compositions: [], scenarios: [], runs: [] })).toBe(true);
  });
});
