import { render, screen, within } from "@testing-library/react";
import axe from "axe-core";
import { describe, expect, it } from "vitest";
import { WorkspaceStore } from "../../app/workspaceStore";
import { failed, success } from "../../model/bridge";
import { FakeBridge, failure } from "../../test/fakeBridge";
import { HomeScreen } from "./HomeScreen";

async function show(bridge: FakeBridge) {
  const store = new WorkspaceStore(bridge);
  await store.reload();
  return render(<HomeScreen workspace={store} />);
}

describe("home screen", () => {
  it("lists plants, runs and compositions with no accessibility violation", async () => {
    const view = await show(new FakeBridge());
    expect(screen.getByText("1 plant · 1 composition · 0 scenarios · 2 runs")).toBeInTheDocument();
    const runs = screen.getByRole("region", { name: "Runs" });
    expect(within(runs).getByText("w-1")).toBeInTheDocument();
    expect(within(runs).getByText("Evaluate")).toBeInTheDocument();
    expect(within(runs).getByText("running")).toBeInTheDocument();
    expect(within(screen.getByRole("region", { name: "Plants" })).getByText("factory-utilities")).toBeInTheDocument();
    expect(
      within(screen.getByRole("region", { name: "Scenarios" })).getByText("No scenarios yet."),
    ).toBeInTheDocument();
    expect((await axe.run(view.container)).violations).toEqual([]);
  });

  it("says what to do in an empty workspace", async () => {
    const bridge = new FakeBridge();
    bridge.workspaceOutcome = success({ name: "fresh", domains: [], compositions: [], scenarios: [], runs: [] });
    await show(bridge);
    expect(screen.getByRole("heading", { name: "fresh is empty" })).toBeInTheDocument();
  });

  it("shows the loading state and the bridge's failure", async () => {
    const bridge = new FakeBridge();
    render(<HomeScreen workspace={new WorkspaceStore(bridge)} />);
    expect(screen.getByRole("status")).toHaveTextContent("Loading the workspace");
    bridge.workspaceOutcome = failed(failure("bridge_failed", 500, "disk full"));
    await show(bridge);
    expect(screen.getByRole("alert")).toHaveTextContent("disk full (bridge_failed)");
  });
});
