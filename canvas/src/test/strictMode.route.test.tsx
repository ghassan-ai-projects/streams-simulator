import { useEffect, useState } from "react";
import { describe, expect, it } from "vitest";
import { HomeScreen } from "../ui/workspace";
import { FakeBridge } from "./fakeBridge";
import { navigation } from "./homeHarness";
import { renderStrict } from "./strictMode";
import type { NavigationStore } from "../app/shell";
import { PlantMaker } from "../app/plant";
import { WorkspaceStore } from "../app/workspace";

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
      <HomeScreen workspace={workspace} navigation={nav} maker={new PlantMaker(bridge.plant)} />
    </>
  );
}

describe("a route under strict mode", () => {
  it("survives the setup-cleanup-setup replay with one live watcher and the route on screen", async () => {
    const bridge = new FakeBridge();
    const workspace = new WorkspaceStore(bridge.workspace);
    const nav = navigation();
    const view = renderStrict(<Route workspace={workspace} nav={nav.store} bridge={bridge} />);
    await view.findByText("opened");
    expect(bridge.watchers.size).toBe(1);
    expect(view.getByRole("heading", { name: "lab-workspace" })).toBeVisible();
    view.unmount();
    expect(bridge.watchers.size).toBe(0);
  });
});
