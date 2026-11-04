import { describe, expect, it } from "vitest";
import type { ConnectState } from "../../launch";
import type { WorkspaceState } from "../../workspace";
import { offeredSession } from "../../../test/interruptFixture";
import { sampleWorkspace } from "../../../test/workspaceFixture";
import { runIsLive } from "./runLive";

const loading: WorkspaceState = { phase: "loading" };
const idle: ConnectState = { phase: "idle" };
const session: ConnectState = { ...offeredSession, run: "w-1" };

function listing(connected: "running" | "saved"): WorkspaceState {
  const [run] = sampleWorkspace.runs;
  if (run === undefined) throw new Error("the sample workspace has no run");
  return { phase: "ready", workspace: { ...sampleWorkspace, runs: [{ ...run, worldId: "w-1", connected }] } };
}

describe("a run counts as live", () => {
  it("when this browser's session follows it", () => {
    expect(runIsLive(session, loading, "w-1")).toBe(true);
    expect(runIsLive(session, loading, "w-2")).toBe(false);
  });

  it("when the workspace lists it as running", () => {
    expect(runIsLive(idle, listing("running"), "w-1")).toBe(true);
    expect(runIsLive(idle, listing("running"), "w-2")).toBe(false);
    expect(runIsLive(idle, listing("saved"), "w-1")).toBe(false);
  });

  it("not before the workspace has loaded", () => {
    expect(runIsLive(idle, loading, "w-1")).toBe(false);
  });
});
