import { act, fireEvent, screen, within } from "@testing-library/react";
import { beforeEach, describe, expect, it } from "vitest";
import { failure as failed, success } from "../../../test/outcome";
import { mountApp } from "../../../test/appHarness";
import { deferred } from "../../../test/deferred";
import type { Canned } from "../../../test/fakeFailures";
import { stubFlowDom } from "../../../test/dom";
import { cannedFailure } from "../../../test/fakeBridge";
import { reportWith } from "../../../test/runFixture";
import { reviewBridge, reviewHash, streamJournal } from "../../../test/screenFixture";

beforeEach(() => {
  stubFlowDom();
});

async function reviewing(bridge = reviewBridge(), hash = reviewHash()) {
  const app = mountApp(bridge, hash);
  await screen.findByRole("region", { name: "Run w-1" });
  return { ...app, bridge };
}

function played(app: Awaited<ReturnType<typeof reviewing>>) {
  act(() => {
    streamJournal(app.bridge);
    app.timers.advance(250);
  });
}

function outlinePick(id: RegExp) {
  fireEvent.click(within(screen.getByRole("navigation", { name: "Outline" })).getByRole("button", { name: id }));
}

describe("the review screen", () => {
  it("names the run, its mode and its verification once the report is read", async () => {
    const { view } = await reviewing();
    expect(await screen.findByText("Evaluate")).toBeInTheDocument();
    await screen.findByText("Verified", { selector: "[data-part='status-chip']" });
    expect(view.container.querySelector("[data-part='status-chip']")).toHaveTextContent("Verified");
    expect(screen.getByRole("heading", { name: "Run w-1" })).toHaveAttribute("aria-current", "location");
  });

  it("says the run is being read while the report is on its way", async () => {
    const bridge = reviewBridge();
    const gate = deferred<Canned<unknown>>();
    bridge.handlers.set("run.report", () => gate.promise);
    await reviewing(bridge);
    expect(screen.getAllByText(/Reading the run/).length).toBeGreaterThan(0);
    await act(async () => {
      gate.resolve(success(reportWith({})));
      await Promise.resolve();
    });
    expect(await screen.findByText("Evaluate")).toBeInTheDocument();
  });

  it("shows the failure line when the report cannot be read", async () => {
    const bridge = reviewBridge();
    bridge.actions.set("run.report", failed(cannedFailure("bridge_failed", 422, "run.json is damaged")));
    await reviewing(bridge);
    expect(await screen.findByText(/run\.json is damaged/)).toBeInTheDocument();
    expect(screen.getAllByText("No report").length).toBeGreaterThan(0);
  });

  it("offers a way back when the run is not in the workspace", async () => {
    const bridge = reviewBridge();
    bridge.actions.set("run.report", failed(cannedFailure("not_found", 404, "no such run")));
    const app = mountApp(bridge, reviewHash("w-99"));
    expect(await screen.findByText("Run w-99 is not in this workspace.")).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Back to the workspace" }));
    expect(app.stores.navigation.snapshot()).toEqual({ screen: "workspace" });
  });

  it("draws the plant the run recorded and lists its instances in the outline", async () => {
    await reviewing();
    expect(await screen.findByRole("group", { name: /^CH-1, chiller/ })).toBeInTheDocument();
    expect(screen.getByRole("navigation", { name: "Outline" })).toBeInTheDocument();
  });

  it("shows what an instance reported at the playback time when it is picked", async () => {
    const app = await reviewing();
    played(app);
    await screen.findByRole("group", { name: /^CH-1, chiller/ });
    outlinePick(/CH-1/);
    expect(screen.getByRole("button", { name: "Inspect · CH-1" })).toHaveAttribute("aria-pressed", "true");
    const observed = await screen.findByRole("region", { name: "Observed · CH-1" });
    expect(within(observed).getByRole("table")).toBeInTheDocument();
    expect(within(observed).getByText(/Chiller power|power/i)).toBeInTheDocument();
  });

  it("asks for a selection on the Inspect tab when nothing is picked", async () => {
    await reviewing();
    fireEvent.click(screen.getByRole("button", { name: "Inspect" }));
    expect(screen.getByText(/Select an instance on the graph/)).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Report" }));
    expect(screen.getByRole("region", { name: "Report" })).toBeInTheDocument();
  });

  it("plays and pauses with the space bar", async () => {
    const app = await reviewing();
    played(app);
    fireEvent.keyDown(document.body, { key: " " });
    expect(app.stores.review.snapshot().playing).toBe(true);
    fireEvent.keyDown(document.body, { key: " " });
    expect(app.stores.review.snapshot().playing).toBe(false);
  });

  it("leaves the space bar to a focused button", async () => {
    const app = await reviewing();
    played(app);
    fireEvent.keyDown(screen.getByRole("button", { name: "Inspect" }), { key: " " });
    expect(app.stores.review.snapshot().playing).toBe(false);
  });

  it("opens the plant from the run and proposes another run with the same settings", async () => {
    const app = await reviewing();
    fireEvent.click(await screen.findByRole("button", { name: "Set up another run" }));
    expect(app.stores.navigation.snapshot()).toMatchObject({ screen: "plant", hall: null });
    expect(await screen.findByRole("button", { name: "Run plant" })).toHaveAttribute("aria-pressed", "true");
  });

  it("opens the plant the run used", async () => {
    const app = await reviewing();
    fireEvent.click(await screen.findByRole("button", { name: "Open plant" }));
    expect(app.stores.navigation.snapshot()).toMatchObject({ screen: "plant", hall: null });
  });

  it("marks an instance on the graph as faulted from the time its fault begins", async () => {
    const app = await reviewing();
    played(app);
    const before = await screen.findByRole("group", { name: /^CT-1, tower/ });
    expect(before.getAttribute("aria-label")).not.toMatch(/fault/i);
    act(() => {
      app.stores.review.seek(700e9);
    });
    const after = await screen.findByRole("group", { name: /^CT-1, Fault: fan failure, tower/ });
    expect(within(after).getByText("Fault: fan failure")).toBeInTheDocument();
  });

  it("lists what is active on the picked instance now, with its latest readings", async () => {
    const app = await reviewing();
    played(app);
    await screen.findByRole("group", { name: /^CT-1, tower/ });
    act(() => {
      app.stores.review.seek(700e9);
      app.stores.review.select("CT-1");
    });
    const observed = await screen.findByRole("region", { name: "Observed · CT-1" });
    expect(within(observed).getByRole("list", { name: "Active now" })).toHaveTextContent(/fan failure/i);
    expect(within(observed).getByRole("table")).toHaveTextContent("basin temp27.0");
  });

  it("shows an effector call on the instance it was sent to", async () => {
    const app = await reviewing();
    played(app);
    await screen.findByRole("group", { name: /^CH-1, chiller/ });
    act(() => {
      app.stores.review.seek(1_250e9);
      app.stores.review.select("CH-1");
    });
    const observed = await screen.findByRole("region", { name: "Observed · CH-1" });
    expect(within(observed).getByRole("list", { name: "Active now" })).toHaveTextContent("Effector call: start");
  });

  it("says an instance without readings has none at this time", async () => {
    const app = await reviewing();
    played(app);
    await screen.findByRole("group", { name: /^MB-1, meter/ });
    act(() => {
      app.stores.review.select("MB-1");
    });
    expect(await screen.findByText("No observation of MB-1 at this time.")).toBeInTheDocument();
  });
});
