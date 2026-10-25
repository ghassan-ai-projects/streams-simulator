import { act, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { createLaunchStore } from "../../../app/launch";
import { createNavigationStore } from "../../../app/shell";
import { failure as failed, success } from "../../../test/outcome";
import { cannedFailure } from "../../../test/fakeBridge";
import { fakeClock } from "../../../test/fakeClock";
import { elapse, launchBridge } from "../../../test/launchFixture";
import { memoryLocation } from "../../../test/location";
import { RunIndicator } from "./RunIndicator";

const origin = { domain: "mini", title: "Mini plant", path: "domains/mini.domain.json" };
const plan = { seed: 1, durationS: 60, explore: true, faults: [], perturbations: [], calls: [] };

function setup(status?: object) {
  const bridge = launchBridge("w-5");
  if (status !== undefined) bridge.actions.set("session.status", success(status));
  const location = memoryLocation();
  const timers = fakeClock();
  bridge.jobs.drive(timers);
  const launch = createLaunchStore(bridge.launch, timers, timers);
  render(<RunIndicator launch={launch} navigation={createNavigationStore(location)} />);
  return { bridge, launch, location, timers };
}

describe("run indicator", () => {
  it("is absent without a run and opens the review of a finished run", async () => {
    const { launch, location, timers } = setup();
    expect(screen.getByRole("status")).toBeEmptyDOMElement();
    await act(async () => {
      await launch.start(origin, plan);
      await elapse(timers, 500);
    });
    expect(screen.getByRole("status")).toHaveTextContent("The review of run w-5 is ready");
    fireEvent.click(screen.getByRole("button", { name: "Run w-5 finished · Open review" }));
    expect(location.hash).toBe("#/review?world=w-5");
    expect(screen.getByRole("status")).toBeEmptyDOMElement();
  });

  it("names a named run in the pill and in the announcement, and its review once finished", async () => {
    const { launch, timers } = setup();
    const named = { ...plan, name: "Cooling loss", durationS: 1800, pace: "watch" as const };
    await act(async () => {
      await launch.start(origin, named);
      await elapse(timers, 500);
    });
    expect(
      screen.getByRole("button", { name: "Running Cooling loss · 00:00:30 of 00:30:00 · 60 s per s" }),
    ).toBeVisible();
    expect(screen.getByRole("status")).toHaveTextContent("Run of Cooling loss on Mini plant started");
    await act(async () => {
      await elapse(timers, 30_000);
    });
    expect(screen.getByRole("button", { name: "Run w-5 (Cooling loss) finished · Open review" })).toBeVisible();
  });

  it("shows progress, links back to the plant's run panel, and lets a failure be dismissed", async () => {
    const { bridge, launch, location } = setup();
    bridge.actions.set("job.start", failed(cannedFailure("domain_invalid", 422, "bad")));
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

  it("shows the live rate, then Paused with the frozen clock, and announces the pause", async () => {
    const { launch, timers } = setup();
    const long = { ...plan, durationS: 1800, pace: "watch" as const };
    await act(async () => {
      await launch.start(origin, long);
      await elapse(timers, 500);
    });
    expect(
      screen.getByRole("button", { name: "Running Mini plant · 00:00:30 of 00:30:00 · 60 s per s" }),
    ).toBeVisible();
    await act(async () => {
      await launch.setRate(10);
      await launch.pause();
    });
    expect(screen.getByRole("button", { name: "Paused Mini plant · 00:00:30 of 00:30:00" })).toBeVisible();
    expect(screen.getByRole("status")).toHaveTextContent("Run of Mini plant paused");
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
