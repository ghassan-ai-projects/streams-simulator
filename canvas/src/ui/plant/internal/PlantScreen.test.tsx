import { act, fireEvent, screen, within } from "@testing-library/react";
import { beforeEach, describe, expect, it } from "vitest";
import { success } from "../../../test/outcome";
import { mountApp } from "../../../test/appHarness";
import { deferred } from "../../../test/deferred";
import type { Canned } from "../../../test/fakeFailures";
import { stubFlowDom } from "../../../test/dom";
import { campusBridge, campusHash, miniBridge, miniWorkspace, planHash } from "../../../test/screenFixture";

beforeEach(() => {
  stubFlowDom();
});

async function opened(hash = planHash(), bridge = miniBridge()) {
  const app = mountApp(bridge, hash);
  await screen.findByRole("region", { name: /^Plant / });
  return { ...app, bridge };
}

function graph() {
  return screen.getByRole("application", { name: "Plant graph" });
}

function pressOn(target: HTMLElement, key: string) {
  fireEvent.keyDown(target, { key });
}

describe("the plant screen", () => {
  it("opens the plant the route names with its summary and an unpressed Run button", async () => {
    await opened();
    expect(screen.getByText("3 instances · 3 links")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Run plant" })).toHaveAttribute("aria-pressed", "false");
    expect(screen.getByRole("group", { name: "CH-1, chiller, 2 inputs, 2 outputs" })).toBeInTheDocument();
  });

  it("says the plant is not in the workspace when its file is gone, and goes back", async () => {
    const bridge = miniBridge();
    bridge.files.clear();
    const app = mountApp(bridge, planHash());
    expect(await screen.findByRole("alert")).toHaveTextContent("This plant is not in the workspace any more.");
    fireEvent.click(screen.getByRole("button", { name: "Back to the workspace" }));
    expect(app.stores.navigation.snapshot()).toEqual({ screen: "workspace" });
  });

  it("shows a loading status until the workspace listing arrives", async () => {
    const bridge = miniBridge();
    const gate = bridge.holdWorkspace();
    mountApp(bridge, planHash());
    expect(await screen.findByText("Opening the plant…")).toBeInTheDocument();
    gate[0]?.resolve(success(miniWorkspace));
    expect(await screen.findByRole("region", { name: "Plant mini" })).toBeInTheDocument();
  });

  it("opens the inspector on an instance picked in the outline and clears it with the close button", async () => {
    await opened();
    const outline = screen.getByRole("navigation", { name: "Outline" });
    fireEvent.click(within(outline).getByRole("button", { name: /CH-1/ }));
    const inspector = screen.getByRole("complementary", { name: "Inspector" });
    expect(within(inspector).getByRole("heading", { name: "CH-1" })).toBeInTheDocument();
    expect(within(inspector).getByText("lib/chiller")).toBeInTheDocument();
    fireEvent.click(within(inspector).getByRole("button", { name: "Close the inspector" }));
    expect(screen.getByRole("region", { name: "Areas" })).toBeInTheDocument();
  });

  it("follows a link in the inspector to the instance on the other end", async () => {
    await opened();
    fireEvent.click(within(screen.getByRole("navigation", { name: "Outline" })).getByRole("button", { name: /CH-1/ }));
    fireEvent.click(screen.getByRole("button", { name: /^Select receiver CT-1;/ }));
    expect(
      within(screen.getByRole("complementary", { name: "Inspector" })).getByRole("heading", { level: 2 }),
    ).toHaveTextContent("CT-1");
  });

  it("ignores an inspector link to an instance the plant does not have", async () => {
    const bridge = miniBridge();
    bridge.addFile("domains/mini.domain.json", {
      id: "mini",
      site: { name: "Mini", areas: [{ name: "Mini" }] },
      instances: [{ kind: "chiller", names: ["CH-1"], area: "Mini" }],
      links: [{ from: "CH-1", to: "GHOST", type: "tower__heat_in__chiller" }],
    });
    await opened(planHash(), bridge);
    fireEvent.click(within(screen.getByRole("navigation", { name: "Outline" })).getByRole("button", { name: /CH-1/ }));
    fireEvent.click(screen.getByRole("button", { name: /^Select receiver GHOST;/ }));
    expect(screen.getByRole("heading", { name: "CH-1" })).toBeInTheDocument();
  });

  it("moves between nodes with the arrow keys, says where it is, and selects with Enter then clears with Escape", async () => {
    await opened();
    const flow = graph();
    flow.focus();
    fireEvent.focus(flow);
    pressOn(flow, "ArrowRight");
    const said = document.getElementById("graph-said");
    expect(said?.textContent).not.toBe("");
    pressOn(flow, "Enter");
    expect(screen.getByRole("complementary", { name: "Inspector" })).toBeInTheDocument();
    pressOn(flow, "Escape");
    expect(screen.getByRole("region", { name: "Areas" })).toBeInTheDocument();
  });

  it("selects the focused node with Space", async () => {
    await opened();
    const flow = graph();
    fireEvent.focus(flow);
    pressOn(flow, " ");
    expect(screen.getByRole("complementary", { name: "Inspector" })).toBeInTheDocument();
  });

  it("zooms with + - and 0 from the keyboard", async () => {
    await opened();
    const flow = graph();
    const viewport = () => flow.querySelector<HTMLElement>(".react-flow__viewport")?.style.transform ?? "";
    fireEvent.focus(flow);
    const fitted = viewport();
    pressOn(flow, "+");
    const zoomedIn = viewport();
    expect(zoomedIn).not.toBe(fitted);
    pressOn(flow, "-");
    pressOn(flow, "0");
    expect(viewport()).toBe(fitted);
  });

  it("opens the launch panel from the Run button and closes it back to the inspector", async () => {
    await opened();
    const run = screen.getByRole("button", { name: "Run plant" });
    fireEvent.click(run);
    expect(run).toHaveAttribute("aria-pressed", "true");
    fireEvent.click(run);
    expect(run).toHaveAttribute("aria-pressed", "false");
  });

  it("starts with the launch panel open when the route asks and consumes the request", async () => {
    const app = await opened(`${planHash()}&launch=1`);
    expect(screen.getByRole("button", { name: "Run plant" })).toHaveAttribute("aria-pressed", "true");
    expect(app.location.hash).not.toContain("launch");
  });

  it("goes to the editor from Edit plant", async () => {
    const app = await opened();
    fireEvent.click(screen.getByRole("button", { name: "Edit plant" }));
    expect(app.stores.navigation.snapshot()).toEqual({
      screen: "edit",
      composition: "compositions/mini.composition.json",
    });
  });

  it("offers no Edit button for a plant without a composition", async () => {
    const bridge = miniBridge();
    bridge.workspaceOutcome = success({ ...miniWorkspace, compositions: [] });
    await opened(planHash(), bridge);
    expect(screen.queryByRole("button", { name: "Edit plant" })).toBeNull();
  });

  it("lists a plant's recent runs in the overview and opens one for review", async () => {
    const app = await opened();
    fireEvent.click(screen.getByRole("button", { name: "Review w-1" }));
    expect(app.stores.navigation.snapshot()).toEqual({ screen: "review", worldId: "w-1" });
  });
});

describe("a campus plant", () => {
  it("opens a hall with Enter on its node and returns with Escape", async () => {
    const app = await opened(campusHash(), campusBridge());
    const flow = graph();
    fireEvent.focus(flow);
    pressOn(flow, "Enter");
    expect(app.stores.navigation.snapshot()).toMatchObject({ screen: "plant", hall: "h01" });
    pressOn(graph(), "Escape");
    expect(app.stores.navigation.snapshot()).toMatchObject({ screen: "plant", hall: null });
  });

  it("shows a hall's own crumb trail and goes back to the whole plant from it", async () => {
    const app = await opened(campusHash("h01"), campusBridge());
    expect(screen.getByRole("heading", { name: "h01" })).toHaveAttribute("aria-current", "location");
    expect(document.title).toBe("h01 · Campus · Streams Simulator");
    fireEvent.click(
      within(screen.getByRole("navigation", { name: "Breadcrumb" })).getByRole("button", { name: "Campus" }),
    );
    expect(app.stores.navigation.snapshot()).toMatchObject({ screen: "plant", hall: null });
  });
});

describe("a campus plant's halls", () => {
  it("summarises a hall picked on the graph without opening it", async () => {
    await opened(campusHash(), campusBridge());
    fireEvent.click(screen.getByRole("group", { name: /^h01, / }));
    const inspector = screen.getByRole("complementary", { name: "Inspector" });
    expect(within(inspector).getByRole("heading", { name: "h01" })).toBeInTheDocument();
    expect(within(inspector).getByText(/Hall · 2 instances · Enter or double-click opens it/)).toBeInTheDocument();
  });

  it("opens a hall on double-click", async () => {
    const app = await opened(campusHash(), campusBridge());
    fireEvent.doubleClick(screen.getByRole("group", { name: /^h02, / }));
    expect(app.stores.navigation.snapshot()).toMatchObject({ screen: "plant", hall: "h02" });
  });

  it("selects the instance at the other end of a link even when it sits outside the open hall", async () => {
    const app = await opened(campusHash("h01"), campusBridge());
    fireEvent.click(
      within(screen.getByRole("navigation", { name: "Outline" })).getByRole("button", { name: /h01\/M-1/ }),
    );
    fireEvent.click(screen.getByRole("button", { name: /^Select source CAL-1;/ }));
    expect(app.stores.plants.snapshot()).toMatchObject({ selected: "CAL-1" });
  });
});

describe("a plant with a run going", () => {
  const origin = { domain: "mini", title: "Mini plant", path: "domains/mini.domain.json" };
  const plan = {
    seed: 1,
    durationS: 3600,
    explore: true,
    faults: [{ entity: "CH-1", fault: "chiller__fouling", atS: 0 }],
    calls: [],
  };

  async function running() {
    const bridge = miniBridge();
    const gate = deferred<Canned<unknown>>();
    bridge.handlers.set("clock.advance", () => gate.promise);
    const app = await opened(planHash(), bridge);
    act(() => {
      void app.stores.launch.start(origin, plan);
    });
    await screen.findByRole("button", { name: /^Running · / });
    return { ...app, gate };
  }

  it("shows the run's clock on the Run button and marks the instances its faults hit", async () => {
    await running();
    expect(screen.getByRole("button", { name: /^Running · 00:00:00/ })).toBeInTheDocument();
    expect(screen.getByRole("group", { name: /^CH-1, Fault: fouling/ })).toBeInTheDocument();
    expect(screen.getByRole("group", { name: /^CT-1, tower/ })).toBeInTheDocument();
  });

  it("holds the editor until the run is over", async () => {
    const app = await running();
    const edit = screen.getByRole("button", { name: "Edit plant" });
    expect(edit).toHaveAttribute("aria-disabled", "true");
    fireEvent.click(edit);
    expect(app.stores.navigation.snapshot()).toMatchObject({ screen: "plant" });
  });
});
