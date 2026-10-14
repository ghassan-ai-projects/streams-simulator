import { fireEvent, render, screen, within } from "@testing-library/react";
import axe from "axe-core";
import { describe, expect, it, vi } from "vitest";
import { PlantMaker } from "../../app/plantMaker";
import { NavigationStore } from "../../app/navigationStore";
import { WorkspaceStore } from "../../app/workspaceStore";
import { failed, success } from "../../model/bridge";
import { FakeBridge, failure, sampleWorkspace } from "../../test/fakeBridge";
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
  return render(<HomeScreen workspace={store} navigation={nav.store} maker={new PlantMaker(bridge)} />);
}

describe("home screen", () => {
  it("lists each plant once with its actions, and its runs newest first, with no accessibility violation", async () => {
    const view = await show(new FakeBridge());
    expect(screen.getByText("1 plant · 2 runs")).toBeInTheDocument();
    const plants = screen.getByRole("region", { name: "Plants" });
    expect(within(plants).getAllByRole("listitem")).toHaveLength(1);
    expect(within(plants).getByText("Factory utilities")).toBeInTheDocument();
    expect(within(plants).getByText(/^2 runs · last /)).toBeInTheDocument();
    for (const action of ["Open", "Edit", "Run"]) {
      expect(within(plants).getByRole("button", { name: `${action} Factory utilities` })).toBeInTheDocument();
    }
    const runs = within(screen.getByRole("region", { name: "Runs" })).getAllByRole("listitem");
    expect(runs.map((row) => row.textContent)).toEqual([
      expect.stringMatching(/^Factory utilitiesExplore.*w-2 · .*not finishedReview$/),
      expect.stringMatching(/^Factory utilitiesEvaluate.*w-1 · .* · 30 minReview$/),
    ]);
    expect(screen.queryByRole("region", { name: "Scenarios" })).not.toBeInTheDocument();
    expect(screen.queryByText(/domains\//)).not.toBeInTheDocument();
    expect((await axe.run(view.container)).violations).toEqual([]);
  });

  it("filters runs by plant and pages a long list", async () => {
    const bridge = new FakeBridge();
    const runs = Array.from({ length: 25 }, (_, i) => ({
      worldId: `w-${String(i)}`,
      ended: true,
      explore: false,
      plant: i % 5 === 0 ? "other" : "factory-utilities",
      durationS: 60,
      recordedAt: `2026-10-14T10:${String(i).padStart(2, "0")}:00Z`,
    }));
    bridge.workspaceOutcome = success({
      ...sampleWorkspace,
      domains: [...sampleWorkspace.domains, { path: "domains/other.domain.json", id: "other", title: "Other" }],
      runs,
    });
    await show(bridge);
    const region = screen.getByRole("region", { name: "Runs" });
    expect(within(region).getAllByRole("listitem")).toHaveLength(20);
    fireEvent.click(within(region).getByRole("button", { name: "Show 5 more of 5" }));
    expect(within(region).getAllByRole("listitem")).toHaveLength(25);
    fireEvent.change(within(region).getByLabelText("Show runs of"), { target: { value: "other" } });
    expect(within(region).getAllByRole("listitem")).toHaveLength(5);
  });

  it("shows the badge of a run already checked in this session", async () => {
    const bridge = new FakeBridge();
    const store = new WorkspaceStore(bridge);
    await store.reload();
    render(
      <HomeScreen
        workspace={store}
        navigation={navigation().store}
        maker={new PlantMaker(bridge)}
        badgeFor={(id) => (id === "w-1" ? "verified" : null)}
      />,
    );
    expect(screen.getByText("Verified")).toBeInTheDocument();
  });

  it("opens a plant from its row", async () => {
    const nav = navigation();
    await show(new FakeBridge(), nav);
    fireEvent.click(screen.getByRole("button", { name: "Open Factory utilities" }));
    expect(nav.location.hash).toBe("#/plant?domain=domains%2Ffactory-utilities.domain.json");
    fireEvent.click(screen.getByRole("button", { name: "Run Factory utilities" }));
    expect(nav.location.hash).toBe("#/plant?domain=domains%2Ffactory-utilities.domain.json&run=1");
    fireEvent.click(screen.getByRole("button", { name: "Review w-1" }));
    expect(nav.location.hash).toBe("#/run?world=w-1");
  });

  it("edits a composition from its row and creates a new plant from a base", async () => {
    const nav = navigation();
    const bridge = new FakeBridge();
    const base = "compositions/factory-utilities.composition.json";
    const factory = {
      id: "factory-utilities",
      equipment: [{ use: "lib/chiller@1", names: ["CH-1"] }],
      connections: [],
    };
    bridge.files.set(base, success({ path: base, content: JSON.stringify(factory), etag: "b" }));
    await show(bridge, nav);
    fireEvent.click(screen.getByRole("button", { name: "Edit Factory utilities" }));
    expect(nav.location.hash).toBe("#/edit?composition=compositions%2Ffactory-utilities.composition.json");
    fireEvent.click(screen.getByRole("button", { name: "New plant" }));
    const form = screen.getByRole("form", { name: "New plant" });
    fireEvent.change(within(form).getByLabelText("Name"), { target: { value: "Factory utilities!" } });
    expect(within(form).getByLabelText("Id")).toHaveValue("factory-utilities-2");
    fireEvent.change(within(form).getByLabelText("Name"), { target: { value: "Lab two" } });
    expect(within(form).getByLabelText("Id")).toHaveValue("lab-two");
    fireEvent.click(within(form).getByRole("button", { name: "Create and open the editor" }));
    await vi.waitFor(() => {
      expect(nav.location.hash).toBe("#/edit?composition=compositions%2Flab-two.composition.json");
    });
    expect(bridge.writes[0]).toMatchObject({ path: "compositions/lab-two.composition.json", etag: "" });
    expect(JSON.parse(bridge.writes[0]?.content ?? "")).toMatchObject({
      title: "Lab two",
      equipment: factory.equipment,
    });
  });

  it("creates an empty plant from a base and keeps an id typed by hand", async () => {
    const nav = navigation();
    const bridge = new FakeBridge();
    const base = "compositions/factory-utilities.composition.json";
    const factory = { id: "factory-utilities", equipment: [{ use: "lib/chiller@1", names: ["CH-1"] }] };
    bridge.files.set(base, success({ path: base, content: JSON.stringify(factory), etag: "b" }));
    await show(bridge, nav);
    fireEvent.click(screen.getByRole("button", { name: "New plant" }));
    const form = screen.getByRole("form", { name: "New plant" });
    fireEvent.change(within(form).getByLabelText("Name"), { target: { value: "Bare" } });
    fireEvent.change(within(form).getByLabelText("Id"), { target: { value: "bare-hall" } });
    fireEvent.change(within(form).getByLabelText("Name"), { target: { value: "Bare hall" } });
    fireEvent.click(within(form).getByRole("radio", { name: /^An empty plant/ }));
    fireEvent.click(within(form).getByRole("button", { name: "Create and open the editor" }));
    await vi.waitFor(() => {
      expect(bridge.writes[0]?.path).toBe("compositions/bare-hall.composition.json");
    });
    expect(JSON.parse(bridge.writes[0]?.content ?? "")).toMatchObject({ title: "Bare hall", equipment: [] });
  });

  it("reports a plant that could not be created and cancels", async () => {
    const bridge = new FakeBridge();
    await show(bridge);
    fireEvent.click(screen.getByRole("button", { name: "New plant" }));
    const form = screen.getByRole("form", { name: "New plant" });
    fireEvent.change(within(form).getByLabelText("Name"), { target: { value: "Lab two" } });
    fireEvent.click(within(form).getByRole("button", { name: "Create and open the editor" }));
    expect(await within(form).findByRole("alert")).toHaveTextContent("no compositions/factory-utilities");
    fireEvent.keyDown(within(form).getByLabelText("Name"), { key: "Escape" });
    expect(screen.getByRole("button", { name: "New plant" })).toHaveFocus();
    fireEvent.click(screen.getByRole("button", { name: "New plant" }));
    expect(screen.getByLabelText("Name")).toHaveFocus();
    fireEvent.click(screen.getByRole("button", { name: "Cancel" }));
    expect(screen.getByRole("button", { name: "New plant" })).toHaveFocus();
  });

  it("says what to do in an empty workspace", async () => {
    const bridge = new FakeBridge();
    bridge.workspaceOutcome = success({ name: "fresh", domains: [], compositions: [], scenarios: [], runs: [] });
    await show(bridge);
    expect(screen.getByRole("heading", { name: "fresh has no plants yet" })).toBeInTheDocument();
    const crumbs = screen.getByRole("navigation", { name: "Breadcrumb" });
    expect(within(crumbs).getByRole("heading", { level: 1, name: "Workspace" })).toBeInTheDocument();
    expect(screen.getAllByRole("listitem")).toHaveLength(4);
  });

  it("offers the shipped examples in an empty workspace and opens a tried one ready to run", async () => {
    const nav = navigation();
    const bridge = new FakeBridge();
    bridge.workspaceOutcome = success({ name: "fresh", domains: [], compositions: [], scenarios: [], runs: [] });
    bridge.actions.set(
      "component.examples",
      success({ examples: [{ id: "pilot", title: "Pilot", description: "A small plant." }] }),
    );
    bridge.actions.set("component.example", success({ example: { id: "pilot", title: "Pilot", equipment: [] } }));
    bridge.actions.set("domain.compose", success({}));
    const view = await show(bridge, nav);
    const examples = await screen.findByRole("region", { name: "Try an example" });
    expect(within(examples).getByText("A small plant.")).toBeInTheDocument();
    expect((await axe.run(view.container)).violations).toEqual([]);
    fireEvent.click(within(examples).getByRole("button", { name: "Try Pilot" }));
    await vi.waitFor(() => {
      expect(nav.location.hash).toBe("#/plant?domain=domains%2Fpilot.domain.json&run=1");
    });
    expect(bridge.writes[0]?.path).toBe("compositions/pilot.composition.json");
  });

  it("says why an example could not be set up", async () => {
    const bridge = new FakeBridge();
    bridge.workspaceOutcome = success({ name: "fresh", domains: [], compositions: [], scenarios: [], runs: [] });
    bridge.actions.set("component.examples", success({ examples: [{ id: "pilot", title: "Pilot" }] }));
    bridge.actions.set("component.example", failed(failure("component_failed", 422, "gone")));
    await show(bridge);
    fireEvent.click(await screen.findByRole("button", { name: "Try Pilot" }));
    expect(await screen.findByRole("alert")).toHaveTextContent("Pilot could not be set up: gone");
  });

  it("creates a new plant from a shipped example", async () => {
    const nav = navigation();
    const bridge = new FakeBridge();
    bridge.actions.set("component.examples", success({ examples: [{ id: "pilot", title: "Pilot" }] }));
    bridge.actions.set(
      "component.example",
      success({ example: { id: "pilot", title: "Pilot", equipment: [{ use: "x@1" }] } }),
    );
    await show(bridge, nav);
    await vi.waitFor(() => {
      expect(bridge.calls.some((call) => call.name === "component.examples")).toBe(true);
    });
    fireEvent.click(screen.getByRole("button", { name: "New plant" }));
    const form = screen.getByRole("form", { name: "New plant" });
    fireEvent.change(within(form).getByLabelText("Start from"), { target: { value: "example:pilot" } });
    fireEvent.change(within(form).getByLabelText("Name"), { target: { value: "My pilot" } });
    fireEvent.click(within(form).getByRole("button", { name: "Create and open the editor" }));
    await vi.waitFor(() => {
      expect(nav.location.hash).toBe("#/edit?composition=compositions%2Fmy-pilot.composition.json");
    });
    expect(JSON.parse(bridge.writes[0]?.content ?? "")).toMatchObject({ id: "my-pilot", equipment: [{ use: "x@1" }] });
  });

  it("shows the loading state and the bridge's failure", async () => {
    const bridge = new FakeBridge();
    render(
      <HomeScreen
        workspace={new WorkspaceStore(bridge)}
        navigation={navigation().store}
        maker={new PlantMaker(bridge)}
      />,
    );
    expect(screen.getByRole("status")).toHaveTextContent("Loading the workspace");
    bridge.workspaceOutcome = failed(failure("bridge_failed", 500, "disk full"));
    await show(bridge);
    expect(screen.getByRole("alert")).toHaveTextContent("disk full (bridge_failed)");
  });
});
