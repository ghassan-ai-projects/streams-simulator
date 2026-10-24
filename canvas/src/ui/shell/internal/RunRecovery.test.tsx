import { act, fireEvent, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { mountApp } from "../../../test/appHarness";
import { flushPromises } from "../../../test/deferred";
import { stubFlowDom } from "../../../test/dom";
import { elapse } from "../../../test/launchFixture";
import { attributeValues, canary } from "../../../test/runningPlant";
import { miniBridge, planHash } from "../../../test/screenFixture";

beforeEach(() => {
  stubFlowDom();
});

const request = { domain: "mini", seed: 4, explore: false, duration_s: 1800, rate: 60, faults: [], calls: [] };

const planned = {
  ...request,
  name: "Cooling loss",
  purpose: "Observe the chiller with the tower out",
  faults: [{ entity_id: "CH-1", fault: "chiller__CANARY_TRUTH_9f3a", at_s: 60, until_s: 1200 }],
  perturbations: [{ kind: "CANARY_TRUTH_drop", params: { rate: 0.5 }, at_s: 120 }],
};

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

  it("brings an exploration's faults back with it: marks on the plant, the count, and its name and purpose", async () => {
    const bridge = miniBridge();
    mountApp(bridge, planHash());
    bridge.jobs.adopt(300, { ...planned, explore: true });
    const pill = await screen.findByRole("button", {
      name: "Running Cooling loss · 00:05:00 of 00:30:00 · 60 s per s",
    });
    fireEvent.click(pill);
    expect(await screen.findByText("Running Cooling loss on Mini plant as run w-7")).toBeInTheDocument();
    expect(screen.getByText("Observe the chiller with the tower out")).toBeInTheDocument();
    expect(screen.getByText("30 min of simulated time · Explore · 1 fault · 1 delivery condition")).toBeInTheDocument();
    expect(document.body.textContent).toMatch(canary);
    expect(await screen.findByRole("group", { name: canary })).toHaveAccessibleName(/^CH-1, Fault: /);
    expect(screen.getAllByTitle(/CH-1 CANARY.TRUTH.9f3a at 00:01:00/i)).toHaveLength(1);
  });

  it("brings an evaluation back with its name and purpose but nothing of its sealed plan", async () => {
    const bridge = miniBridge();
    mountApp(bridge, planHash());
    bridge.jobs.adopt(300, planned);
    fireEvent.click(await screen.findByRole("button", { name: /^Running Cooling loss · / }));
    expect(await screen.findByText("Observe the chiller with the tower out")).toBeInTheDocument();
    expect(screen.getByText("30 min of simulated time · Evaluate")).toBeInTheDocument();
    expect(document.body.textContent).not.toMatch(canary);
    expect(attributeValues().filter((value) => canary.test(value))).toEqual([]);
    expect(screen.getByRole("group", { name: /^CH-1, chiller/ })).toBeInTheDocument();
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
