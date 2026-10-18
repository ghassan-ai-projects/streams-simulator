import { fireEvent, render, screen, within } from "@testing-library/react";
import { ReactFlowProvider } from "@xyflow/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { plantView } from "../../../model/graph";
import type { PlanMark, ReadoutRow } from "../../../model/graph";
import { parsePlant } from "../../../model/plant";
import { stubFlowDom } from "../../../test/dom";
import { composition, domain } from "../../../test/plantFixture";
import { PlantGraph } from "./PlantGraph";
import type { NodeLive, PlantGraphProps } from "./types";

beforeEach(() => {
  stubFlowDom();
});

const view = plantView(parsePlant(domain, composition), new Map(), { open: null });

function row(overrides: Partial<ReadoutRow>): ReadoutRow {
  return {
    key: "k",
    label: "power",
    present: true,
    value: "92.7",
    unit: "kW",
    word: false,
    attention: null,
    outside: "",
    observed: true,
    ...overrides,
  };
}

function live(overrides: Partial<NodeLive>): NodeLive {
  const readout = { rows: [row({})], hidden: 0, hiddenAttention: null };
  return {
    shown: [],
    observed: new Set(),
    readout,
    status: null,
    count: 0,
    fault: "",
    effector: "",
    origin: null,
    ...overrides,
  };
}

function draw(props: Partial<PlantGraphProps>) {
  return render(
    <ReactFlowProvider>
      <PlantGraph view={view} viewKey="v" selected={null} onSelect={vi.fn()} {...props} />
    </ReactFlowProvider>,
  );
}

function node(id: string) {
  return screen.getByRole("group", { name: new RegExp(`^${id},`) });
}

describe("a graph drawn from a connected session", () => {
  it("marks a real board and a simulated instance on their nodes, in words for every reader", () => {
    const real = { instance: "CH-1", kind: "real", board: "uno-07", device: "ch-1", bootId: "61" } as const;
    const simulated = { ...real, instance: "CT-1", kind: "simulated", board: "mega-01" } as const;
    const origins = new Map<string, NodeLive["origin"]>([
      ["CH-1", real],
      ["CT-1", simulated],
    ]);
    draw({ live: (id) => live({ origin: origins.get(id) ?? null }) });
    expect(node("CH-1").getAttribute("aria-label")).toContain("Real · board uno-07 · boot 61");
    expect(node("CH-1").querySelector("[data-origin='real']")).not.toBeNull();
    expect(node("CT-1").querySelector("[data-origin='simulated']")).not.toBeNull();
    expect(node("MB-1").querySelector("[data-origin]")).toBeNull();
  });
});

describe("a graph drawn from a run", () => {
  it("shows an alarm count, the fault behind it and an effector call on the instance", () => {
    draw({
      live: (id) =>
        id === "CH-1" ? live({ status: "alarm", count: 2, fault: "fouling", effector: "start · applied" }) : live({}),
    });
    const chiller = node("CH-1");
    expect(chiller.getAttribute("aria-label")).toContain("effector call: start · applied");
    expect(within(chiller).getByText("Fault · fouling")).toBeInTheDocument();
    expect(within(chiller).getByText("start · applied")).toBeInTheDocument();
    expect(within(chiller).getByText("Effector call: start · applied")).toBeInTheDocument();
    expect(within(chiller).getByText("Fault: fouling")).toBeInTheDocument();
  });

  it("tells a channel the run never reported from one that has not reported yet", () => {
    draw({
      live: () =>
        live({
          readout: {
            rows: [
              row({ key: "a", label: "cop", present: false, observed: false }),
              row({ key: "b", label: "flow", present: false, observed: true }),
              row({ key: "c", label: "temp", value: "41", unit: "", attention: "warn", outside: "above 40" }),
            ],
            hidden: 0,
            hiddenAttention: null,
          },
        }),
    });
    const chiller = node("CH-1");
    expect(within(chiller).getByText("not observed")).toBeInTheDocument();
    expect(within(chiller).getByText("no reading yet")).toBeInTheDocument();
    expect(within(chiller).getByText("41")).toBeInTheDocument();
  });

  it("folds channels that do not fit and says how many are in Inspect", () => {
    draw({
      live: () => live({ readout: { rows: [row({})], hidden: 3, hiddenAttention: "alarm" } }),
    });
    expect(within(node("CH-1")).getByText("+3 more")).toBeInTheDocument();
    expect(within(node("CH-1")).getByText("in Inspect")).toBeInTheDocument();
  });

  it("says an instance has no channels when its readout is empty", () => {
    draw({ live: () => live({ readout: { rows: [], hidden: 0, hiddenAttention: null } }) });
    expect(within(node("MB-1")).getByText("no channels")).toBeInTheDocument();
  });

  it("flags an instance with issues for screen readers and in words", () => {
    draw({ flagged: new Set(["CT-1"]) });
    expect(node("CT-1").getAttribute("aria-label")).toContain("has issues");
    expect(within(node("CT-1")).getByText("Issue")).toBeInTheDocument();
  });
});

describe("a graph drawn before any run", () => {
  it("marks the instances a planned fault will hit and the calls it will make", () => {
    const marks = new Map<string, PlanMark>([
      ["CH-1", { status: "fault", count: 1, fault: "fouling", effector: "start" }],
    ]);
    draw({ marks });
    expect(node("CH-1").getAttribute("aria-label")).toContain("effector call: start");
    expect(within(node("CH-1")).getByText(/Fault · fouling/)).toBeInTheDocument();
  });

  it("selects the clicked node and clears on the empty pane", () => {
    const onSelect = vi.fn();
    const { container } = draw({ onSelect });
    fireEvent.click(node("CT-1"));
    expect(onSelect).toHaveBeenLastCalledWith("CT-1");
    const pane = container.querySelector(".react-flow__pane");
    if (pane === null) throw new Error("no pane");
    fireEvent.click(pane);
    expect(onSelect).toHaveBeenLastCalledWith(null);
  });
});
