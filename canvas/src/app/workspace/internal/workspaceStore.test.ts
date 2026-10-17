import { describe, expect, it } from "vitest";
import { failure as failed, success } from "../../ports";
import { flushPromises } from "../../../test/deferred";
import { FakeBridge, cannedFailure } from "../../../test/fakeBridge";
import { sampleWorkspace } from "../../../test/workspaceFixture";
import { WorkspaceStore } from "./workspaceStore";

describe("workspace", () => {
  it("lists the workspace and reloads when a file changes on disk", async () => {
    const bridge = new FakeBridge();
    const store = new WorkspaceStore(bridge.workspace);
    await store.open();
    expect(store.snapshot()).toEqual({ phase: "ready", workspace: sampleWorkspace });
    bridge.workspaceOutcome = success({ ...sampleWorkspace, runs: [] });
    bridge.changeOnDisk();
    await flushPromises();
    expect(store.snapshot()).toMatchObject({ phase: "ready", workspace: { runs: [] } });
  });

  it("watches once however often it opens, and stops on close", async () => {
    const bridge = new FakeBridge();
    const store = new WorkspaceStore(bridge.workspace);
    await store.open();
    await store.open();
    expect(bridge.watchers.size).toBe(1);
    store.close();
    expect(bridge.watchers.size).toBe(0);
  });

  it("shows the bridge's failure", async () => {
    const bridge = new FakeBridge();
    bridge.workspaceOutcome = failed(cannedFailure("bridge_failed", 500, "disk"));
    const store = new WorkspaceStore(bridge.workspace);
    await store.open();
    expect(store.snapshot()).toMatchObject({
      phase: "failed",
      failure: { kind: "unavailable", diagnostic: { message: "disk", displayCode: "bridge_failed" } },
    });
  });

  it("publishes two overlapping reloads in request order, not in completion order", async () => {
    const bridge = new FakeBridge();
    const store = new WorkspaceStore(bridge.workspace);
    await store.open();
    const reads = bridge.holdWorkspace();
    bridge.changeOnDisk();
    bridge.changeOnDisk();
    await flushPromises();
    reads[1]?.resolve(success({ ...sampleWorkspace, name: "newer" }));
    await flushPromises();
    reads[0]?.resolve(success({ ...sampleWorkspace, name: "older" }));
    await flushPromises();
    expect(store.snapshot()).toMatchObject({ phase: "ready", workspace: { name: "newer" } });
  });

  it("reloads on a change that arrives while the first load is still running", async () => {
    const bridge = new FakeBridge();
    const reads = bridge.holdWorkspace(1);
    bridge.workspaceOutcome = success({ ...sampleWorkspace, name: "changed" });
    const store = new WorkspaceStore(bridge.workspace);
    const opening = store.open();
    bridge.changeOnDisk();
    await flushPromises();
    reads[0]?.resolve(success({ ...sampleWorkspace, name: "initial" }));
    await opening;
    expect(store.snapshot()).toMatchObject({ phase: "ready", workspace: { name: "changed" } });
    store.close();
  });

  it("keeps reloading on a change on disk after opening again", async () => {
    const bridge = new FakeBridge();
    const store = new WorkspaceStore(bridge.workspace);
    await store.open();
    await store.open();
    bridge.workspaceOutcome = success({ ...sampleWorkspace, name: "changed" });
    bridge.changeOnDisk();
    await flushPromises();
    expect(store.snapshot()).toMatchObject({ phase: "ready", workspace: { name: "changed" } });
  });

  it("does not publish a load that completes after close, and leaves no watcher", async () => {
    const bridge = new FakeBridge();
    const held = bridge.holdWorkspace();
    const store = new WorkspaceStore(bridge.workspace);
    const opening = store.open();
    await flushPromises();
    store.close();
    held[0]?.resolve(success({ ...sampleWorkspace, name: "late" }));
    await opening;
    expect(bridge.watchers.size).toBe(0);
    expect(store.snapshot()).toEqual({ phase: "loading" });
  });

  it("does not let a closed visit publish or watch after the workspace reopens", async () => {
    const bridge = new FakeBridge();
    const reads = bridge.holdWorkspace();
    const store = new WorkspaceStore(bridge.workspace);
    const closedVisit = store.open();
    await flushPromises();
    store.close();
    const reopenedVisit = store.open();
    await flushPromises();
    reads[0]?.resolve(success({ ...sampleWorkspace, name: "late" }));
    await closedVisit;
    expect(bridge.watchers.size).toBe(1);
    expect(store.snapshot()).toEqual({ phase: "loading" });
    reads[1]?.resolve(success({ ...sampleWorkspace, name: "current" }));
    await reopenedVisit;
    expect(bridge.watchers.size).toBe(1);
    expect(store.snapshot()).toMatchObject({ phase: "ready", workspace: { name: "current" } });
    store.close();
  });
});
