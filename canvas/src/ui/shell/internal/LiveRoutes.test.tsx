import { act, fireEvent, screen, within } from "@testing-library/react";
import { beforeEach, describe, expect, it } from "vitest";
import { mountApp } from "../../../test/appHarness";
import { stubFlowDom } from "../../../test/dom";
import { liveWire } from "../../../test/liveFixture";
import { success } from "../../../test/outcome";
import { miniWorkspace, planHash, reviewBridge } from "../../../test/screenFixture";

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
});
