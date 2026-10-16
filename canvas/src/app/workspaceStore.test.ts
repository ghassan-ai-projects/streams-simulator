import { describe, expect, it } from "vitest";
import { failed, success } from "../model/bridge";
import { deferred, settle } from "../test/deferred";
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

  it("publishes two overlapping reloads in request order, not in completion order", async () => {
    const bridge = new FakeBridge();
    const first = deferred<Awaited<ReturnType<FakeBridge["workspace"]>>>();
    const second = deferred<Awaited<ReturnType<FakeBridge["workspace"]>>>();
    const reads = [first, second];
    bridge.workspace = () => {
      const next = reads.shift();
      if (next === undefined) throw new Error("no held read left");
      return next.promise;
    };
    const store = new WorkspaceStore(bridge);
    const older = store.reload();
    const newer = store.reload();
    second.resolve(success({ ...sampleWorkspace, name: "newer" }));
    await newer;
    first.resolve(success({ ...sampleWorkspace, name: "older" }));
    await older;
    expect(store.snapshot()).toMatchObject({ phase: "ready", workspace: { name: "newer" } });
  });

  it("keeps reloading on a change on disk after opening again", async () => {
    const bridge = new FakeBridge();
    const store = new WorkspaceStore(bridge);
    await store.open();
    await store.open();
    bridge.workspaceOutcome = success({ ...sampleWorkspace, name: "changed" });
    bridge.changeOnDisk();
    await settle();
    expect(store.snapshot()).toMatchObject({ phase: "ready", workspace: { name: "changed" } });
  });

  it("does not publish a reload that completes after close, and installs no watcher", async () => {
    const bridge = new FakeBridge();
    const held = deferred<Awaited<ReturnType<FakeBridge["workspace"]>>>();
    bridge.workspace = () => held.promise;
    const store = new WorkspaceStore(bridge);
    const opening = store.open();
    store.close();
    held.resolve(success({ ...sampleWorkspace, name: "late" }));
    await opening;
    expect(bridge.watchers.size).toBe(0);
  });

  it("does not let a closed visit install a watcher after the workspace reopens", async () => {
    const bridge = new FakeBridge();
    const first = deferred<Awaited<ReturnType<FakeBridge["workspace"]>>>();
    const second = deferred<Awaited<ReturnType<FakeBridge["workspace"]>>>();
    const reads = [first, second];
    bridge.workspace = () => {
      const next = reads.shift();
      if (next === undefined) throw new Error("no held read left");
      return next.promise;
    };
    const store = new WorkspaceStore(bridge);
    const closedVisit = store.open();
    store.close();
    const reopenedVisit = store.open();
    first.resolve(success({ ...sampleWorkspace, name: "late" }));
    await closedVisit;
    expect(bridge.watchers.size).toBe(0);
    second.resolve(success({ ...sampleWorkspace, name: "current" }));
    await reopenedVisit;
    expect(bridge.watchers.size).toBe(1);
    expect(store.snapshot()).toMatchObject({ phase: "ready", workspace: { name: "current" } });
    store.close();
  });
});
