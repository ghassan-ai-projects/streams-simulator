import { useEffect, useState } from "react";
import { describe, expect, it } from "vitest";
import { WorkspaceScreen } from "../ui/workspace";
import { FakeBridge } from "./fakeBridge";
import { navigation } from "./workspaceHarness";
import { renderStrict } from "./strictMode";
import type { NavigationStore } from "../app/shell";
import { createDeviceDesk } from "../app/devices";
import { createPlantMaker } from "../app/plant";
import type { WorkspaceStore } from "../app/workspace";
import { createWorkspaceStore } from "../app/workspace";

function Route({
  workspace,
  nav,
  bridge,
}: {
  readonly workspace: WorkspaceStore;
  readonly nav: NavigationStore;
  readonly bridge: FakeBridge;
}) {
  const [opened, setOpened] = useState(false);
  useEffect(() => {
    void workspace.open().then(() => {
      setOpened(true);
    });
    return () => {
      workspace.close();
    };
  }, [workspace]);
  return (
    <>
      <p>{opened ? "opened" : "opening"}</p>
      <WorkspaceScreen
        workspace={workspace}
        navigation={nav}
        maker={createPlantMaker(bridge.plant)}
        desk={createDeviceDesk(bridge.devices)}
        badgeFor={() => null}
      />
    </>
  );
}

describe("a route under strict mode", () => {
  it("survives the setup-cleanup-setup replay with one live watcher and the route on screen", async () => {
    const bridge = new FakeBridge();
    const workspace = createWorkspaceStore(bridge.workspace);
    const nav = navigation();
    const view = renderStrict(<Route workspace={workspace} nav={nav.store} bridge={bridge} />);
    await view.findByText("opened");
    expect(bridge.watchers.size).toBe(1);
    expect(view.getByRole("heading", { name: "lab-workspace" })).toBeVisible();
    view.unmount();
    expect(bridge.watchers.size).toBe(0);
  });
});
