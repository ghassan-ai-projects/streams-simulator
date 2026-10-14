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
    <LaunchPanel launch={launch} plant={parsePlant(domain)} onClose={vi.fn()} onFinished={onFinished} />,
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
    change("Seed", "9");
    fireEvent.click(screen.getByRole("button", { name: "Add fault" }));
    change("Fault 1 instance", "CH-1");
    expect(screen.getByRole("option", { name: "fouling" })).toBeInTheDocument();
    change("Fault 1 fault", "chiller__fouling");
    change("Fault 1 minute", "10");
    await act(async () => {
      fireEvent.click(screen.getByRole("button", { name: "Start the run" }));
      await Promise.resolve();
    });
    await vi.waitFor(() => {
      expect(onFinished).toHaveBeenCalledWith("w-7");
    });
    expect(fake.calls[0]).toEqual({ name: "world.create", args: { domain: "mini", seed: 9, explore: true } });
    expect(fake.calls.find((call) => call.name === "fault.inject")?.args).toEqual({
      world_id: "w-7",
      entity_id: "CH-1",
      fault: "chiller__fouling",
    });
  });

  it("lists the plan's problems and does not start", () => {
    const fake = bridge();
    renderPanel(fake);
    change("Duration", "0");
    fireEvent.click(screen.getByRole("button", { name: "Add fault" }));
    const start = screen.getByRole("button", { name: "Start the run" });
    expect(start).toHaveAttribute("aria-disabled", "true");
    expect(screen.getByText("The duration is at least one second.")).toBeInTheDocument();
    fireEvent.click(start);
    expect(fake.calls).toEqual([]);
    fireEvent.click(screen.getByRole("button", { name: "Remove fault 1" }));
    expect(screen.getByText("None: the plant runs healthy.")).toBeInTheDocument();
  });

  it("shows the binary's refusal and returns to the plan", async () => {
    const fake = bridge();
    fake.actions.set("world.create", failed(failure("domain_invalid", 422, "bad domain")));
    renderPanel(fake);
    fireEvent.click(screen.getByRole("button", { name: "Start the run" }));
    expect(await screen.findByRole("alert")).toHaveTextContent("The run did not finish: bad domain");
    fireEvent.click(screen.getByRole("button", { name: "Change the plan" }));
    expect(screen.getByRole("button", { name: "Start the run" })).toBeInTheDocument();
  });

  it("shows progress and cancels while running", async () => {
    const fake = new HeldBridge();
    renderPanel(fake);
    fireEvent.click(screen.getByRole("button", { name: "Start the run" }));
    expect(await screen.findByText("Running w-7")).toHaveAttribute("role", "status");
    expect(screen.getByText("00:00:00 of 00:30:00 simulated")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Close the run panel" })).toBeDisabled();
    fireEvent.click(screen.getByRole("button", { name: "Cancel" }));
    await act(async () => {
      fake.release();
      await Promise.resolve();
    });
    expect(await screen.findByRole("alert")).toHaveTextContent("The run was canceled.");
    expect(fake.calls.at(-1)?.name).toBe("world.destroy");
  });
});

class HeldBridge extends FakeBridge {
  private held: (() => void) | null = null;

  constructor() {
    super();
    bridge().actions.forEach((outcome, name) => this.actions.set(name, outcome));
  }

  override action(name: string, args: Readonly<Record<string, unknown>>): Promise<Outcome<unknown>> {
    if (name !== "clock.advance") return super.action(name, args);
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
}
