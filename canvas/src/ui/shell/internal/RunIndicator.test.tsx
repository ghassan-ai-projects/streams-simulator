import { act, fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { LaunchStore } from "../../../app/launch";
import { NavigationStore } from "../../../app/shell";
import { failure as failed } from "../../../test/outcome";
import { cannedFailure } from "../../../test/fakeBridge";
import { fakeClock } from "../../../test/fakeClock";
import { launchBridge } from "../../../test/launchFixture";
import { memoryLocation } from "../../../test/location";
import { RunIndicator } from "./RunIndicator";

const origin = { domain: "mini", title: "Mini plant", path: "domains/mini.domain.json" };
const plan = { seed: 1, durationS: 60, explore: true, faults: [], calls: [] };

function setup() {
  const bridge = launchBridge("w-5");
  const location = memoryLocation();
  const timers = fakeClock();
  const launch = new LaunchStore(bridge.launch, timers, timers);
  render(<RunIndicator launch={launch} navigation={new NavigationStore(location)} />);
  return { bridge, launch, location };
}

describe("run indicator", () => {
  it("is absent without a run and opens the review of a finished run", async () => {
    const { launch, location } = setup();
    expect(screen.getByRole("status")).toBeEmptyDOMElement();
    await act(async () => {
      await launch.start(origin, plan);
    });
    expect(screen.getByRole("status")).toHaveTextContent("The review of run w-5 is ready");
    fireEvent.click(screen.getByRole("button", { name: "Run w-5 finished · Open review" }));
    expect(location.hash).toBe("#/review?world=w-5");
    expect(screen.getByRole("status")).toBeEmptyDOMElement();
  });

  it("shows progress, links back to the plant's run panel, and lets a failure be dismissed", async () => {
    const { bridge, launch, location } = setup();
    bridge.actions.set("world.create", failed(cannedFailure("domain_invalid", 422, "bad")));
    let running = Promise.resolve();
    act(() => {
      running = launch.start(origin, plan);
    });
    expect(screen.getByRole("button", { name: "Running Mini plant · 00:00:00 of 00:01:00" })).toBeInTheDocument();
    expect(screen.getByRole("status")).toHaveTextContent(/^Run of Mini plant started$/);
    fireEvent.click(screen.getByRole("button", { name: /^Running Mini plant/ }));
    expect(location.hash).toBe("#/plant?domain=domains%2Fmini.domain.json&launch=1");
    await act(async () => {
      await running;
    });
    expect(screen.getByRole("status")).toHaveTextContent("The run of Mini plant did not finish");
    const heading = document.createElement("h1");
    heading.tabIndex = -1;
    document.body.prepend(heading);
    fireEvent.click(screen.getByRole("button", { name: "Dismiss" }));
    expect(screen.getByRole("status")).toBeEmptyDOMElement();
    expect(heading).toHaveFocus();
    heading.remove();
  });
});
