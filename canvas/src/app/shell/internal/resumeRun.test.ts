import { describe, expect, it } from "vitest";
import { createWorkspaceStore } from "../../workspace";
import { flushPromises } from "../../../test/deferred";
import { FakeBridge } from "../../../test/fakeBridge";
import { launchHarness } from "../../../test/launchFixture";
import { success } from "../../../test/outcome";
import { miniWorkspace } from "../../../test/screenFixture";
import { resumeRun } from "./resumeRun";

const request = { domain: "mini", seed: 4, explore: true, duration_s: 1800, rate: 60, faults: [], calls: [] };

function setup() {
  const bridge = new FakeBridge();
  bridge.workspaceOutcome = success(miniWorkspace);
  const { launch, timers } = launchHarness(bridge);
  bridge.jobs.adopt(300, request);
  const workspace = createWorkspaceStore(bridge.workspace);
  return { bridge, launch, timers, workspace };
}

describe("finding a run after a reload", () => {
  it("waits for the workspace, then takes over the running job under the plant's own title and path", async () => {
    const { bridge, launch, workspace } = setup();
    const stop = resumeRun(launch, workspace);
    await flushPromises();
    expect(launch.snapshot()).toEqual({ phase: "idle" });
    expect(bridge.calls.map((call) => call.name)).not.toContain("job.list");
    await workspace.open();
    await flushPromises();
    expect(launch.snapshot()).toMatchObject({
      phase: "running",
      worldId: "w-7",
      atS: 300,
      origin: { domain: "mini", path: "domains/mini.domain.json" },
    });
    stop();
    workspace.close();
  });

  it("asks once, however often the workspace changes, and not after it is stopped", async () => {
    const { bridge, launch, workspace } = setup();
    const stop = resumeRun(launch, workspace);
    await workspace.open();
    await flushPromises();
    bridge.changeOnDisk();
    await flushPromises();
    expect(bridge.calls.filter((call) => call.name === "job.list")).toHaveLength(1);
    stop();
    workspace.close();
  });

  it("does nothing once stopped before the workspace is ready", async () => {
    const { bridge, launch, workspace } = setup();
    resumeRun(launch, workspace)();
    await workspace.open();
    await flushPromises();
    expect(bridge.calls.map((call) => call.name)).not.toContain("job.list");
    workspace.close();
  });
});
