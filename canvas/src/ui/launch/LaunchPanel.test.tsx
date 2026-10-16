import { act, fireEvent, render, screen } from "@testing-library/react";
import axe from "axe-core";
import { describe, expect, it, vi } from "vitest";
import { LaunchStore } from "../../app/launchStore";
import { failed } from "../../model/bridge";
import { parsePlant } from "../../model/plant";
import type { FakeBridge } from "../../test/fakeBridge";
import { failure } from "../../test/fakeBridge";
import { HeldBridge } from "../../test/heldBridge";
import { launchBridge } from "../../test/launchFixture";
import { domain } from "../../test/plantFixture";
import { LaunchPanel } from "./LaunchPanel";

const path = "domains/mini.domain.json";

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

const finishedWait = { timeout: 3000 };

describe("launch panel", () => {
  it("cues the cut on the run plan's own scroll region while its fields continue below", () => {
    renderPanel(launchBridge());
    const region = document.querySelector<HTMLElement>("[data-scroll='run-plan']");
    expect(region).not.toBeNull();
    expect(region).not.toHaveAttribute("data-cut");
    const metrics = (client: number, total: number, top: number) => {
      if (region === null) return;
      Object.defineProperty(region, "clientHeight", { value: client, configurable: true });
      Object.defineProperty(region, "scrollHeight", { value: total, configurable: true });
      Object.defineProperty(region, "scrollTop", { value: top, configurable: true });
      fireEvent.scroll(region);
    };
    metrics(200, 500, 0);
    expect(region).toHaveAttribute("data-cut", "true");
    metrics(200, 500, 300);
    expect(region).not.toHaveAttribute("data-cut");
  });

  it("plans a run with a fault and opens the finished run", async () => {
    const fake = launchBridge();
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
      fireEvent.click(screen.getByRole("radio", { name: /As fast as possible/ }));
      fireEvent.click(screen.getByRole("button", { name: "Start the run" }));
      await Promise.resolve();
    });
    await vi.waitFor(() => {
      expect(onFinished).toHaveBeenCalledWith("w-7");
    }, finishedWait);
    expect(fake.calls[0]).toEqual({ name: "world.create", args: { domain: "mini", seed: 9, explore: false } });
    expect(fake.calls.find((call) => call.name === "fault.inject")?.args).toEqual({
      world_id: "w-7",
      entity_id: "CH-1",
      fault: "chiller__fouling",
    });
  });

  it("offers to watch the run by default, says how long that takes, and can run as fast as possible", () => {
    renderPanel(launchBridge());
    expect(screen.getByRole("radio", { name: /Watch: about 30 s/ })).toBeChecked();
    expect(screen.getByText(/paced to take about 30 s/)).toBeInTheDocument();
    change("Duration", "10");
    expect(screen.getByRole("radio", { name: /Watch: about 10 s/ })).toBeChecked();
    fireEvent.click(screen.getByRole("radio", { name: /As fast as possible/ }));
    expect(screen.getByText(/as fast as it can/)).toBeInTheDocument();
  });

  it("lists the plan's problems only when starting, and does not start", () => {
    const fake = launchBridge();
    renderPanel(fake);
    change("Duration", "0");
    fireEvent.click(screen.getByRole("button", { name: "Add fault" }));
    expect(screen.queryByText("The duration is at least one minute.")).not.toBeInTheDocument();
    fireEvent.click(screen.getByRole("radio", { name: /As fast as possible/ }));
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
    expect(screen.getByText("None: no fault is scheduled for this run.")).toBeInTheDocument();
    expect(screen.queryByText(/runs healthy/)).not.toBeInTheDocument();
  });

  it("names the missing fault after an instance is chosen and blocks the run", () => {
    const fake = launchBridge();
    renderPanel(fake);
    fireEvent.click(screen.getByRole("button", { name: "Add fault" }));
    change("Fault 1 instance", "CH-1");
    expect(screen.queryByText(/needs attention before starting/)).not.toBeInTheDocument();
    fireEvent.click(screen.getByRole("radio", { name: /As fast as possible/ }));
    fireEvent.click(screen.getByRole("button", { name: "Start the run" }));
    expect(screen.getByText("Fault 1 needs attention before starting.")).toBeInTheDocument();
    const fault = screen.getByLabelText("Fault 1 fault");
    expect(fault).toHaveFocus();
    expect(fault).toHaveAccessibleDescription("Fault 1 needs a fault.");
    expect(fake.calls).toEqual([]);
  });

  it("shows the binary's refusal and returns to the plan", async () => {
    const fake = launchBridge();
    fake.actions.set("world.create", failed(failure("domain_invalid", 422, "bad domain")));
    renderPanel(fake);
    change("Seed", "4");
    fireEvent.click(screen.getByRole("radio", { name: /As fast as possible/ }));
    fireEvent.click(screen.getByRole("button", { name: "Start the run" }));
    expect(await screen.findByRole("alert")).toHaveTextContent(
      "The run did not finish. The run stopped before it finished.",
    );
    expect(screen.getByText("bad domain")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Change the plan" })).toHaveFocus();
    fireEvent.click(screen.getByRole("button", { name: "Change the plan" }));
    expect(screen.getByLabelText("Seed")).toHaveValue("4");
  });

  it("shows progress and stops early, keeping the shorter run and opening it", async () => {
    const fake = new HeldBridge();
    const { onFinished } = renderPanel(fake);
    fireEvent.click(screen.getByRole("radio", { name: /As fast as possible/ }));
    fireEvent.click(screen.getByRole("button", { name: "Start the run" }));
    expect(await screen.findByText("Running Mini plant as run w-7")).toHaveFocus();
    expect(screen.getByText("30 min of simulated time · Explore · no faults")).toBeInTheDocument();
    expect(screen.getByText("Running · 00:00:00 of 00:30:00 simulated")).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Stop now" }));
    expect(screen.getByRole("status")).toHaveTextContent("Stopping at the next step");
    await act(async () => {
      fake.release();
      await Promise.resolve();
    });
    expect(screen.getByRole("status")).toHaveTextContent(/^Saved at 00:\d\d:\d\d of 00:30:00/);
    await vi.waitFor(() => {
      expect(onFinished).toHaveBeenCalledWith("w-7");
    }, finishedWait);
    expect(fake.calls.at(-1)?.name).toBe("run.end");
  });

  it("says which partial run was kept when the binary refuses a step", async () => {
    const fake = launchBridge();
    fake.actions.set("clock.advance", failed(failure("world_gone", 404, "the world is gone")));
    renderPanel(fake);
    fireEvent.click(screen.getByRole("radio", { name: /As fast as possible/ }));
    fireEvent.click(screen.getByRole("button", { name: "Start the run" }));
    expect(await screen.findByRole("alert")).toHaveTextContent(
      "The run did not finish. The run stopped before it finished. Check the plan and start again. What ran before the failure is kept as run w-7.",
    );
  });
});

describe("launch panel picking", () => {
  it("takes a fault's instance from the graph and rolls a new seed", () => {
    const launch = new LaunchStore(launchBridge());
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
    const launch = new LaunchStore(launchBridge());
    await launch.start(
      { domain: "mini", title: "Mini plant", path },
      { seed: 1, durationS: 60, explore: true, faults: [], calls: [] },
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
    const launch = new LaunchStore(launchBridge());
    launch.propose("mini", {
      seed: 5,
      durationS: 7200,
      explore: false,
      faults: [{ entity: "CH-1", fault: "chiller__fouling", atS: 600 }],
      calls: [],
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

  it("keeps a run's seconds, names the run it was prefilled from, and says what was not copied", () => {
    const launch = new LaunchStore(launchBridge());
    const faults = [{ entity: "CH-1", fault: "chiller__fouling", atS: 620 }];
    launch.propose(
      "mini",
      { seed: 5, durationS: 90, explore: true, faults, calls: [] },
      ["1 delivery perturbation"],
      "w-4",
    );
    render(
      <LaunchPanel launch={launch} plant={parsePlant(domain)} path={path} onClose={vi.fn()} onFinished={vi.fn()} />,
    );
    expect(screen.getByLabelText("Duration")).toHaveValue("1.5");
    expect(screen.getByLabelText("Fault 1 starts at minute")).toHaveValue("10.33");
    expect(screen.getByRole("note")).toHaveTextContent(
      "Prefilled from w-4: seed, duration, mode and faults. Change them if needed. Not copied: 1 delivery perturbation.",
    );
  });
});
