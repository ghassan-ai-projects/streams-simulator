import { describe, expect, it } from "vitest";
import { failure as failed, success } from "../../../test/outcome";
import { FakeBridge, cannedFailure } from "../../../test/fakeBridge";
import { flushPromises } from "../../../test/deferred";
import { createSession } from "../../session";
import { WorkspaceStore } from "../../workspace";
import { sessionGate } from "./sessionGate";

function setup(bridge: FakeBridge) {
  const session = createSession(bridge.session);
  const workspace = new WorkspaceStore(bridge.workspace);
  return { session, workspace, stop: sessionGate(session, workspace) };
}

describe("the session gate", () => {
  it("checks the session and opens the workspace once the browser is paired", async () => {
    const bridge = new FakeBridge();
    bridge.sessionOutcome = success(true);
    const { session, workspace, stop } = setup(bridge);
    expect(workspace.snapshot().phase).toBe("loading");
    await flushPromises();
    expect(session.snapshot().phase).toBe("paired");
    expect(workspace.snapshot()).toMatchObject({ phase: "ready", workspace: { name: "lab-workspace" } });
    expect(bridge.watchers.size).toBe(1);
    stop();
    expect(bridge.watchers.size).toBe(0);
  });

  it("leaves the workspace closed for an unpaired browser, and opens it once it pairs", async () => {
    const bridge = new FakeBridge();
    const { session, workspace, stop } = setup(bridge);
    await flushPromises();
    expect(session.snapshot().phase).toBe("unpaired");
    expect(bridge.watchers.size).toBe(0);
    bridge.sessionOutcome = success(true);
    await session.check();
    await flushPromises();
    expect(workspace.snapshot().phase).toBe("ready");
    expect(bridge.watchers.size).toBe(1);
    stop();
  });

  it("checks the session again when the bridge refuses the workspace as unpaired", async () => {
    const bridge = new FakeBridge();
    bridge.sessionOutcome = success(true);
    const { session, stop } = setup(bridge);
    await flushPromises();
    bridge.sessionOutcome = success(false);
    bridge.workspaceOutcome = failed(cannedFailure("unpaired"));
    bridge.changeOnDisk();
    await flushPromises();
    expect(session.snapshot().phase).toBe("unpaired");
    expect(bridge.watchers.size).toBe(0);
    stop();
  });

  it("stops watching the workspace when the gate stops", async () => {
    const bridge = new FakeBridge();
    bridge.sessionOutcome = success(true);
    const { stop } = setup(bridge);
    await flushPromises();
    stop();
    stop();
    expect(bridge.watchers.size).toBe(0);
  });
});
