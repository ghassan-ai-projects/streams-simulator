import { act, fireEvent, render, screen } from "@testing-library/react";
import axe from "axe-core";
import { describe, expect, it, vi } from "vitest";
import { LaunchStore } from "../../app/launchStore";
import type { Outcome } from "../../model/bridge";
import { failed, success } from "../../model/bridge";
import { parsePlant } from "../../model/plant";
import { FakeBridge, failure } from "../../test/fakeBridge";
import { domain } from "../../test/plantFixture";
import { LaunchPanel } from "./LaunchPanel";

const path = "domains/mini.domain.json";

function bridge() {
  const fake = new FakeBridge();
  fake.actions.set("world.create", success({ world_id: "w-7" }));
  fake.actions.set("clock.advance", success({}));
  fake.actions.set("fault.inject", success({ fault_id: "f-0" }));
  fake.actions.set("run.end", success({}));
  fake.actions.set("world.destroy", success({}));
  return fake;
}

function renderPanel(fake: FakeBridge, onFinished = vi.fn<(worldId: string) => void>()) {
  const launch = new LaunchStore(fake);
  const view = render(
    <LaunchPanel launch={launch} plant={parsePlant(domain)} path={path} onClose={vi.fn()} onFinished={onFinished} />,
  );
  return { launch, view, onFinished };
}

function change(label: string, value: string) {
  fireEvent.change(screen.getByLabelText(label), { target: { value } });
}

describe("launch panel", () => {
  it("plans a run with a fault and opens the finished run", async () => {
    const fake = bridge();
    const { view, onFinished } = renderPanel(fake);
    expect((await axe.run(view.container)).violations).toEqual([]);
    expect(screen.getByLabelText("Seed")).toHaveFocus();
    change("Seed", "9");
    fireEvent.click(screen.getByRole("radio", { name: /^Evaluate/ }));
    fireEvent.click(screen.getByRole("button", { name: "Add fault" }));
    expect(screen.getByLabelText("Fault 1 instance")).toHaveFocus();
    expect(screen.getByRole("option", { name: "CH-1 · chiller" })).toBeInTheDocument();
    change("Fault 1 instance", "CH-1");
    expect(screen.getByRole("option", { name: "fouling" })).toBeInTheDocument();
    change("Fault 1 fault", "chiller__fouling");
    change("Fault 1 starts at minute", "10");
    await act(async () => {
      fireEvent.click(screen.getByRole("button", { name: "Start the run" }));
      await Promise.resolve();
    });
    await vi.waitFor(() => {
      expect(onFinished).toHaveBeenCalledWith("w-7");
    });
    expect(fake.calls[0]).toEqual({ name: "world.create", args: { domain: "mini", seed: 9, explore: false } });
    expect(fake.calls.find((call) => call.name === "fault.inject")?.args).toEqual({
      world_id: "w-7",
      entity_id: "CH-1",
      fault: "chiller__fouling",
    });
  });

  it("lists the plan's problems only when starting, and does not start", () => {
    const fake = bridge();
    renderPanel(fake);
    change("Duration", "0");
    fireEvent.click(screen.getByRole("button", { name: "Add fault" }));
    expect(screen.queryByText("The duration is at least one minute.")).not.toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Start the run" }));
    expect(screen.getByLabelText("Duration")).toHaveFocus();
    expect(screen.getByLabelText("Duration")).toHaveAccessibleDescription("The duration is at least one minute.");
    expect(screen.getByLabelText("Fault 1 instance")).toHaveAccessibleDescription(
      "Fault 1 needs an instance and a fault.",
    );
    expect(screen.getByRole("status")).toHaveTextContent(/^2 to fix before starting/);
    expect(fake.calls).toEqual([]);
    fireEvent.click(screen.getByRole("button", { name: "Remove fault 1" }));
    expect(screen.getByRole("button", { name: "Add fault" })).toHaveFocus();
    expect(screen.getByText("None: the plant runs healthy.")).toBeInTheDocument();
  });

  it("names the missing fault after an instance is chosen and blocks the run", () => {
    const fake = bridge();
    renderPanel(fake);
    fireEvent.click(screen.getByRole("button", { name: "Add fault" }));
    change("Fault 1 instance", "CH-1");
    expect(screen.queryByText(/needs attention before starting/)).not.toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Start the run" }));
    expect(screen.getByText("1 entry needs attention before starting.")).toBeInTheDocument();
    const fault = screen.getByLabelText("Fault 1 fault");
    expect(fault).toHaveFocus();
    expect(fault).toHaveAccessibleDescription("Fault 1 needs a fault.");
    expect(fake.calls).toEqual([]);
  });

  it("shows the binary's refusal and returns to the plan", async () => {
    const fake = bridge();
    fake.actions.set("world.create", failed(failure("domain_invalid", 422, "bad domain")));
    renderPanel(fake);
    change("Seed", "4");
    fireEvent.click(screen.getByRole("button", { name: "Start the run" }));
    expect(await screen.findByRole("alert")).toHaveTextContent("The run did not finish. bad domain");
    expect(screen.getByRole("button", { name: "Change the plan" })).toHaveFocus();
    fireEvent.click(screen.getByRole("button", { name: "Change the plan" }));
    expect(screen.getByLabelText("Seed")).toHaveValue("4");
  });

  it("shows progress and stops early, keeping the shorter run and opening it", async () => {
    const fake = new HeldBridge();
    const { onFinished } = renderPanel(fake);
    fireEvent.click(screen.getByRole("button", { name: "Start the run" }));
    expect(await screen.findByText("Running Mini plant as run w-7")).toHaveFocus();
    expect(screen.getByText("30 min of simulated time · Explore · no faults")).toBeInTheDocument();
    expect(screen.getByText("00:00:00 of 00:30:00 simulated")).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Stop now" }));
    expect(screen.getByRole("status")).toHaveTextContent("Stopping at the next step");
    await act(async () => {
      fake.release();
      await Promise.resolve();
    });
    await vi.waitFor(() => {
      expect(onFinished).toHaveBeenCalledWith("w-7");
    });
    expect(fake.calls.at(-1)?.name).toBe("run.end");
  });

  it("says which partial run was kept when the binary refuses a step", async () => {
    const fake = bridge();
    fake.actions.set("clock.advance", failed(failure("world_gone", 404, "the world is gone")));
    renderPanel(fake);
    fireEvent.click(screen.getByRole("button", { name: "Start the run" }));
    expect(await screen.findByRole("alert")).toHaveTextContent(
      "The run did not finish. the world is gone What ran before the failure is kept as run w-7.",
    );
  });
});

describe("launch panel picking", () => {
  it("takes a fault's instance from the graph and rolls a new seed", () => {
    const launch = new LaunchStore(bridge());
    const plant = parsePlant(domain);
    const panel = (selected: string | null) => (
      <LaunchPanel
        launch={launch}
        plant={plant}
        path={path}
        selected={selected}
        onClose={vi.fn()}
        onFinished={vi.fn()}
      />
    );
    const view = render(panel(null));
    fireEvent.click(screen.getByRole("button", { name: "Add fault" }));
    view.rerender(panel("CH-1"));
    expect(screen.getByLabelText("Fault 1 instance")).toHaveValue("CH-1");
    fireEvent.click(screen.getByRole("button", { name: "Add fault on CH-1" }));
    expect(screen.getByLabelText("Fault 2 instance")).toHaveValue("CH-1");
    vi.spyOn(Math, "random").mockReturnValue(0.5);
    fireEvent.click(screen.getByRole("button", { name: "Pick a new seed" }));
    expect(screen.getByLabelText("Seed")).toHaveValue("500000");
    vi.restoreAllMocks();
  });
});

describe("launch panel lifetime", () => {
  it("does not jump to a run that finished while nobody watched it", async () => {
    const launch = new LaunchStore(bridge());
    await launch.start(
      { domain: "mini", title: "Mini plant", path },
      { seed: 1, durationS: 60, explore: true, faults: [] },
    );
    expect(launch.snapshot()).toMatchObject({ phase: "done" });
    const onFinished = vi.fn<(worldId: string) => void>();
    render(
      <LaunchPanel launch={launch} plant={parsePlant(domain)} path={path} onClose={vi.fn()} onFinished={onFinished} />,
    );
    expect(screen.getByRole("button", { name: "Start the run" })).toBeInTheDocument();
    expect(onFinished).not.toHaveBeenCalled();
  });

  it("starts from a plan proposed for this plant, such as Run again", () => {
    const launch = new LaunchStore(bridge());
    launch.propose("mini", {
      seed: 5,
      durationS: 7200,
      explore: false,
      faults: [{ entity: "CH-1", fault: "chiller__fouling", atS: 600 }],
    });
    const view = render(
      <LaunchPanel launch={launch} plant={parsePlant(domain)} path={path} onClose={vi.fn()} onFinished={vi.fn()} />,
    );
    expect(screen.getByLabelText("Seed")).toHaveValue("5");
    expect(screen.getByLabelText("Duration")).toHaveValue("2");
    expect(screen.getByLabelText("Duration unit")).toHaveValue("hours");
    expect(screen.getByRole("radio", { name: /^Evaluate/ })).toBeChecked();
    expect(screen.getByLabelText("Fault 1 instance")).toHaveValue("CH-1");
    expect(screen.getByLabelText("Fault 1 starts at minute")).toHaveValue("10");
    expect(launch.proposal("other")).toBeNull();
    expect(screen.queryByRole("note")).toBeNull();
    view.unmount();
    expect(launch.proposal("mini")).toBeNull();
  });

  it("keeps a run's seconds and says what Run again could not copy", () => {
    const launch = new LaunchStore(bridge());
    const faults = [{ entity: "CH-1", fault: "chiller__fouling", atS: 620 }];
    launch.propose("mini", { seed: 5, durationS: 90, explore: true, faults }, ["1 delivery perturbation"]);
    render(
      <LaunchPanel launch={launch} plant={parsePlant(domain)} path={path} onClose={vi.fn()} onFinished={vi.fn()} />,
    );
    expect(screen.getByLabelText("Duration")).toHaveValue("1.5");
    expect(screen.getByLabelText("Fault 1 starts at minute")).toHaveValue("10.33");
    expect(screen.getByRole("note")).toHaveTextContent("Not copied: 1 delivery perturbation.");
  });

  it("keeps a run going when the panel closes and opens its review when it is back on that plant", async () => {
    const fake = new HeldBridge();
    const launch = new LaunchStore(fake);
    const onFinished = vi.fn<(worldId: string) => void>();
    const plant = parsePlant(domain);
    const panel = () => (
      <LaunchPanel launch={launch} plant={plant} path={path} onClose={vi.fn()} onFinished={onFinished} />
    );
    const first = render(panel());
    fireEvent.click(screen.getByRole("button", { name: "Start the run" }));
    await screen.findByText("Running Mini plant as run w-7");
    first.unmount();
    expect(launch.snapshot()).toMatchObject({ phase: "running", stopping: false, origin: { domain: "mini", path } });
    render(panel());
    expect(screen.getByText("Running Mini plant as run w-7")).toBeInTheDocument();
    await act(async () => {
      fake.free();
      await Promise.resolve();
    });
    await vi.waitFor(() => {
      expect(onFinished).toHaveBeenCalledWith("w-7");
    });
  });

  it("tells another plant that a run is in progress and leaves a finished run of another plant alone", async () => {
    const fake = new HeldBridge();
    const launch = new LaunchStore(fake);
    const elsewhere = { domain: "other", title: "Other plant", path: "domains/other.domain.json" };
    void launch.start(elsewhere, { seed: 1, durationS: 60, explore: true, faults: [] });
    const onFinished = vi.fn<(worldId: string) => void>();
    render(
      <LaunchPanel launch={launch} plant={parsePlant(domain)} path={path} onClose={vi.fn()} onFinished={onFinished} />,
    );
    expect(await screen.findByText("A run of Other plant is in progress.")).toBeInTheDocument();
    await act(async () => {
      fake.free();
      await Promise.resolve();
    });
    await vi.waitFor(() => {
      expect(launch.snapshot()).toMatchObject({ phase: "done", origin: { domain: "other" } });
    });
    expect(onFinished).not.toHaveBeenCalled();
    expect(screen.getByRole("button", { name: "Start the run" })).toBeInTheDocument();
  });
});

class HeldBridge extends FakeBridge {
  private held: (() => void) | null = null;
  private open = false;

  constructor() {
    super();
    bridge().actions.forEach((outcome, name) => this.actions.set(name, outcome));
  }

  override action(name: string, args: Readonly<Record<string, unknown>>): Promise<Outcome<unknown>> {
    if (name !== "clock.advance" || this.open) return super.action(name, args);
    this.calls.push({ name, args });
    return new Promise((resolve) => {
      this.held = () => {
        resolve(success({}));
      };
    });
  }

  release(): void {
    this.held?.();
  }

  free(): void {
    this.open = true;
    this.held?.();
  }
}
