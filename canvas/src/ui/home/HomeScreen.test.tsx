import { fireEvent, render, screen, within } from "@testing-library/react";
import axe from "axe-core";
import { describe, expect, it } from "vitest";
import { NavigationStore } from "../../app/navigationStore";
import { WorkspaceStore } from "../../app/workspaceStore";
import { failed, success } from "../../model/bridge";
import { FakeBridge, failure } from "../../test/fakeBridge";
import { HomeScreen } from "./HomeScreen";

function navigation() {
  const location = {
    hash: "#/",
    read() {
      return this.hash;
    },
    write(hash: string) {
      this.hash = hash;
    },
    subscribe: () => () => undefined,
  };
  return { location, store: new NavigationStore(location) };
}

async function show(bridge: FakeBridge, nav = navigation()) {
  const store = new WorkspaceStore(bridge);
  await store.reload();
  return render(<HomeScreen workspace={store} navigation={nav.store} />);
}

describe("home screen", () => {
  it("lists plants, runs and compositions with no accessibility violation", async () => {
    const view = await show(new FakeBridge());
    expect(screen.getByText("1 plant · 1 composition · 0 scenarios · 2 runs")).toBeInTheDocument();
    const runs = screen.getByRole("region", { name: "Runs" });
    expect(within(runs).getByText("w-1")).toBeInTheDocument();
    expect(within(runs).getByText("Evaluate")).toBeInTheDocument();
    expect(within(runs).getByText("not closed")).toBeInTheDocument();
    expect(within(runs).getAllByText("not checked yet")).toHaveLength(2);
    expect(within(runs).getAllByRole("button", { name: /Open review of/ })).toHaveLength(2);
    expect(within(screen.getByRole("region", { name: "Plants" })).getByText("factory-utilities")).toBeInTheDocument();
    expect(
      within(screen.getByRole("region", { name: "Scenarios" })).getByText("No scenarios yet."),
    ).toBeInTheDocument();
    expect((await axe.run(view.container)).violations).toEqual([]);
  });

  it("shows the badge of a run already checked in this session", async () => {
    const bridge = new FakeBridge();
    const store = new WorkspaceStore(bridge);
    await store.reload();
    render(
      <HomeScreen
        workspace={store}
        navigation={navigation().store}
        badgeFor={(id) => (id === "w-1" ? "verified" : null)}
      />,
    );
    expect(screen.getByText("Verified")).toBeInTheDocument();
  });

  it("opens a plant from its row", async () => {
    const nav = navigation();
    await show(new FakeBridge(), nav);
    fireEvent.click(screen.getByRole("button", { name: "Open plant factory-utilities" }));
    expect(nav.location.hash).toBe("#/plant?domain=domains%2Ffactory-utilities.domain.json");
  });

  it("says what to do in an empty workspace", async () => {
    const bridge = new FakeBridge();
    bridge.workspaceOutcome = success({ name: "fresh", domains: [], compositions: [], scenarios: [], runs: [] });
    await show(bridge);
    expect(screen.getByRole("heading", { name: "fresh is empty" })).toBeInTheDocument();
  });

  it("shows the loading state and the bridge's failure", async () => {
    const bridge = new FakeBridge();
    render(<HomeScreen workspace={new WorkspaceStore(bridge)} navigation={navigation().store} />);
    expect(screen.getByRole("status")).toHaveTextContent("Loading the workspace");
    bridge.workspaceOutcome = failed(failure("bridge_failed", 500, "disk full"));
    await show(bridge);
    expect(screen.getByRole("alert")).toHaveTextContent("disk full (bridge_failed)");
  });
});
