import { act, fireEvent, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { mountApp } from "../../../test/appHarness";
import { flushPromises } from "../../../test/deferred";
import { stubFlowDom } from "../../../test/dom";
import { elapse } from "../../../test/launchFixture";
import { miniBridge, planHash } from "../../../test/screenFixture";

beforeEach(() => {
  stubFlowDom();
});

const request = { domain: "mini", seed: 4, explore: false, duration_s: 1800, rate: 60, faults: [], calls: [] };

describe("a run that was going when the page was reloaded", () => {
  it("returns to the top bar and the launch panel with its progress and Stop now, and says nothing about faults", async () => {
    const bridge = miniBridge();
    const app = mountApp(bridge, planHash());
    bridge.jobs.adopt(300, request);
    const pill = await screen.findByRole("button", { name: "Running Mini plant · 00:05:00 of 00:30:00 · 60 s per s" });
    expect(bridge.calls.map((call) => call.name)).not.toContain("job.start");
    fireEvent.click(pill);
    expect(await screen.findByText("Running Mini plant as run w-7")).toBeInTheDocument();
    expect(screen.getByText("30 min of simulated time · Evaluate")).toBeInTheDocument();
    expect(screen.queryByText(/no faults/)).toBeNull();
    fireEvent.click(screen.getByRole("button", { name: "Stop now" }));
    await act(async () => {
      await elapse(app.timers, 500);
    });
    expect(bridge.calls.map((call) => call.name)).toContain("job.stop");
    expect(screen.getByRole("button", { name: "Run w-7 stopped early · Open review" })).toBeInTheDocument();
  });

  it("does not bring back a run that ended while the page was away", async () => {
    const bridge = miniBridge();
    mountApp(bridge, planHash());
    bridge.jobs.adopt(1800, request);
    await vi.waitFor(() => {
      expect(bridge.calls.map((call) => call.name)).toContain("job.list");
    });
    await act(async () => {
      await flushPromises();
    });
    expect(screen.queryByRole("button", { name: /^Running / })).toBeNull();
    expect(screen.queryByRole("button", { name: /Open review/ })).toBeNull();
  });
});
