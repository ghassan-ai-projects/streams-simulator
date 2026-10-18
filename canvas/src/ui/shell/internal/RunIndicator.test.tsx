import { act, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { createLaunchStore } from "../../../app/launch";
import { createNavigationStore } from "../../../app/shell";
import { failure as failed, success } from "../../../test/outcome";
import { cannedFailure } from "../../../test/fakeBridge";
import { fakeClock } from "../../../test/fakeClock";
import { launchBridge } from "../../../test/launchFixture";
import { memoryLocation } from "../../../test/location";
import { RunIndicator } from "./RunIndicator";

const origin = { domain: "mini", title: "Mini plant", path: "domains/mini.domain.json" };
const plan = { seed: 1, durationS: 60, explore: true, faults: [], calls: [] };

function setup(status?: object) {
  const bridge = launchBridge("w-5");
  if (status !== undefined) bridge.actions.set("session.status", success(status));
  const location = memoryLocation();
  const timers = fakeClock();
  const launch = createLaunchStore(bridge.launch, timers, timers);
  render(<RunIndicator launch={launch} navigation={createNavigationStore(location)} />);
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

  it("leads back to the live view of a running connected session and hides itself there", async () => {
    const { location } = setup({ run: "c-9", state: "running", domain: "mini" });
    const open = await screen.findByRole("button", { name: "Connected session · running · Open live view" });
    fireEvent.click(open);
    expect(location.hash).toBe("#/review?world=c-9");
    await waitFor(() => {
      expect(screen.queryByRole("button", { name: "Connected session · running · Open live view" })).toBeNull();
    });
  });
});
