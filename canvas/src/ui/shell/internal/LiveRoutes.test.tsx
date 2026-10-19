import { act, fireEvent, screen, within } from "@testing-library/react";
import { beforeEach, describe, expect, it } from "vitest";
import { mountApp } from "../../../test/appHarness";
import { stubFlowDom } from "../../../test/dom";
import { interruptsWire, liveWire } from "../../../test/liveFixture";
import { success } from "../../../test/outcome";
import { reportWith } from "../../../test/runFixture";
import { miniWorkspace, planHash, reviewBridge, reviewHash } from "../../../test/screenFixture";

beforeEach(() => {
  stubFlowDom();
});

function liveBridge() {
  const bridge = reviewBridge();
  const run = {
    worldId: "c-9",
    ended: false,
    complete: false,
    incompleteReason: "",
    explore: false,
    seed: 1,
    plant: "mini",
    durationS: 0,
    recordedAt: "2026-10-18T10:00:00Z",
    faults: 0,
    connected: "running" as const,
  };
  bridge.workspaceOutcome = success({ ...miniWorkspace, runs: [...miniWorkspace.runs, run] });
  bridge.actions.set("session.status", success({ run: "c-9", state: "running", domain: "mini" }));
  return bridge;
}

function boardSnapshot(atNs: number) {
  const wire = liveWire(atNs);
  return { ...wire, devices: wire.devices.map((device) => ({ ...device, instance: "CH-1" })) };
}

describe("a running connected session across screens", () => {
  it("shows the board's health on the plant and keeps the session's history in the live view", async () => {
    const bridge = liveBridge();
    const app = mountApp(bridge, planHash());
    const open = await screen.findByRole("button", { name: "Open live view" });
    act(() => {
      bridge.stream("c-9", [boardSnapshot(10e9)], "caught_up");
      app.timers.advance(250);
      app.stores.plants.select("CH-1");
    });
    const instance = await screen.findByRole("region", { name: "Live instance" });
    expect(within(instance).getByText("Online · board uno-07 · boot 61", { exact: true })).toBeInTheDocument();
    expect(within(instance).getByText("28 degC · good")).toBeInTheDocument();
    const history = app.stores.liveReview.snapshot().liveHistory;
    fireEvent.click(open);
    expect(await screen.findByText("Live · following newest record", { exact: true })).toBeInTheDocument();
    expect(app.stores.liveReview.snapshot().liveHistory).toBe(history);
    expect(screen.getByRole("region", { name: "Connection health" })).toBeInTheDocument();
  });

  it("opens the live view from the plant without verify or sealed files, drawing the session plant", async () => {
    const bridge = reviewBridge();
    bridge.actions.set("session.status", success({ run: "c-9", state: "running", domain: "mini" }));
    const app = mountApp(bridge, planHash());
    const open = await screen.findByRole("button", { name: "Open live view" });
    const session = { type: "session", seq: 1, stream: "observer", at_ns: 0, mode: "connected" };
    act(() => {
      bridge.stream("c-9", [{ ...session, plan: { domain: "mini", instances: [] } }, boardSnapshot(10e9)], "caught_up");
      app.timers.advance(250);
    });
    fireEvent.click(open);
    const outline = await screen.findByRole("navigation", { name: "Outline" });
    expect(await within(outline).findByRole("button", { name: /^CH-1/ })).toBeInTheDocument();
    expect(screen.queryByText("Opening the plant this run used…")).toBeNull();
    expect(screen.getByText("Live · following newest record", { exact: true })).toBeInTheDocument();
    expect(bridge.calls.filter((call) => call.name === "run.verify")).toEqual([]);
    expect(bridge.reads.filter((path) => path.startsWith("runs/c-9/"))).toEqual([]);
  });

  it("catches up, then draws the plant the session names even after a binding and with an unnamed report", async () => {
    const bridge = reviewBridge();
    bridge.actions.set("run.report", success(reportWith({ domain: {} })));
    bridge.actions.set("session.status", success({ run: "c-9", state: "running", domain: "mini" }));
    const app = mountApp(bridge, planHash());
    fireEvent.click(await screen.findByRole("button", { name: "Open live view" }));
    expect(await screen.findByText("Catching up to live…", { exact: true })).toBeInTheDocument();
    const at = { stream: "observer", at_ns: 0 };
    const opened = { ...at, seq: 0, type: "run.opened", mode: "evaluate", adapter: { id: "native-jsonl" } };
    const binding = { ...at, seq: 1, type: "binding", outcome: "matched", issues: [] };
    const session = { ...at, seq: 2, type: "session", mode: "connected", plan: { domain: "mini", instances: [] } };
    const board = { instance: "CH-1", device: "sensor", board: "uno-07", boot_id: "61", environment: "hardware" };
    const origin = { ...at, seq: 3, type: "origin", ...board };
    act(() => {
      bridge.stream("c-9", [opened, binding, session, origin, boardSnapshot(10e9)], null);
      app.timers.advance(250);
    });
    expect(screen.getByText("Catching up to live…", { exact: true })).toBeInTheDocument();
    act(() => {
      bridge.stream("c-9", [], "caught_up");
    });
    const outline = await screen.findByRole("navigation", { name: "Outline" });
    expect(await within(outline).findByRole("button", { name: /^CH-1/ })).toBeInTheDocument();
    expect(screen.getByText("Live · following newest record", { exact: true })).toBeInTheDocument();
  });

  it("follows a running session once, however many screens show it, so a second tab keeps connections free", async () => {
    const bridge = liveBridge();
    bridge.actions.set("session.status", success({ run: "c-9", state: "running", domain: "mini", controls: true }));
    const app = mountApp(bridge, reviewHash("c-9"));
    await screen.findByRole("region", { name: "Run c-9" });
    const session = { type: "session", seq: 1, stream: "observer", at_ns: 0, mode: "connected" };
    const binding = { type: "binding", seq: 0, stream: "observer", at_ns: 0, outcome: "matched", issues: [] };
    act(() => {
      bridge.streamFollow(0, [binding], null);
      bridge.stream("c-9", [{ ...session, plan: { domain: "mini", instances: [] } }, boardSnapshot(10e9)], "caught_up");
      app.timers.advance(250);
    });
    await within(await screen.findByRole("navigation", { name: "Outline" })).findByRole("button", { name: /^CH-1/ });
    expect(bridge.openFollows("c-9")).toBe(1);
    expect(app.stores.review.snapshot().worldId).toBeNull();
  });

  it("keeps the interrupt controls and the hold of a board on the plant screen and on the live view", async () => {
    const bridge = liveBridge();
    bridge.actions.set("session.status", success({ run: "c-9", state: "running", domain: "mini", controls: true }));
    const app = mountApp(bridge, planHash());
    const open = await screen.findByRole("button", { name: "Open live view" });
    const held = [{ device: "sensor", board: "uno-07", mode: "hold", latched: false }];
    const snapshot = { ...boardSnapshot(10e9), interrupts: interruptsWire({ devices: held }) };
    act(() => {
      bridge.stream("c-9", [snapshot], "caught_up");
      app.timers.advance(250);
      app.stores.plants.select("CH-1");
    });
    expect(await screen.findByRole("region", { name: "Interrupts" })).toBeInTheDocument();
    expect(screen.getByRole("group", { name: /^CH-1,.*Held/ })).toBeInTheDocument();
    const instance = await screen.findByRole("region", { name: "Live instance" });
    expect(within(instance).getByText("Held", { exact: true })).toHaveAttribute("data-interrupt", "hold");
    fireEvent.click(open);
    expect(await screen.findByText("Live · following newest record", { exact: true })).toBeInTheDocument();
    expect(screen.getByRole("region", { name: "Interrupts" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Emergency stop" })).toBeEnabled();
    act(() => {
      app.stores.liveReview.select("CH-1");
    });
    expect(await screen.findByText("Held", { exact: true, selector: "p" })).toHaveAttribute("data-interrupt", "hold");
  });
});
