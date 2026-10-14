import { describe, expect, it } from "vitest";
import { failed, success } from "../model/bridge";
import { FakeBridge, failure, sampleWorkspace } from "../test/fakeBridge";
import { WorkspaceStore } from "./workspaceStore";

describe("workspace", () => {
  it("lists the workspace and reloads when a file changes on disk", async () => {
    const bridge = new FakeBridge();
    const store = new WorkspaceStore(bridge);
    await store.open();
    expect(store.snapshot()).toEqual({ phase: "ready", workspace: sampleWorkspace });
    bridge.workspaceOutcome = success({ ...sampleWorkspace, runs: [] });
    bridge.changeOnDisk();
    await Promise.resolve();
    expect(store.snapshot()).toMatchObject({ phase: "ready", workspace: { runs: [] } });
  });

  it("watches once however often it opens, and stops on close", async () => {
    const bridge = new FakeBridge();
    const store = new WorkspaceStore(bridge);
    await store.open();
    await store.open();
    expect(bridge.watchers.size).toBe(1);
    store.close();
    expect(bridge.watchers.size).toBe(0);
  });

  it("shows the bridge's failure", async () => {
    const bridge = new FakeBridge();
    bridge.workspaceOutcome = failed(failure("bridge_failed", 500, "disk"));
    const store = new WorkspaceStore(bridge);
    await store.reload();
    expect(store.snapshot()).toMatchObject({ phase: "failed", failure: { message: "disk" } });
  });
});
