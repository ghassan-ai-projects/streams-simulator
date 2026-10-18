import { act, fireEvent, screen, within } from "@testing-library/react";
import axe from "axe-core";
import { beforeEach, describe, expect, it } from "vitest";
import { mountApp } from "../../../test/appHarness";
import { stubFlowDom } from "../../../test/dom";
import { reviewBridge, reviewHash } from "../../../test/screenFixture";
import { liveWire } from "../../../test/liveFixture";

beforeEach(() => {
  stubFlowDom();
});

describe("connected session evidence", () => {
  it("keeps acceptance, refusal, board report and host observation independent", async () => {
    const bridge = reviewBridge();
    const app = mountApp(bridge, reviewHash());
    await screen.findByRole("region", { name: "Run w-1" });
    const wire = liveWire();
    const accepted = wire.commands[0];
    act(() => {
      bridge.stream("w-1", [
        { type: "session", seq: 1, stream: "observer", at_ns: 0, mode: "connected", plan: { instances: [] } },
        {
          ...wire,
          commands: [
            accepted,
            { ...accepted, command_id: "cmd-2", receipt: "refused", reason: "local_control", report: "unknown" },
            { ...accepted, command_id: "cmd-3", receipt: "pending", report: "pending" },
            {
              ...accepted,
              command_id: "cmd-4",
              environment: "hardware",
              effect: "observed",
              effect_evidence: "Host checked flow: observed",
            },
          ],
        },
      ]);
      app.timers.advance(250);
    });
    app.stores.review.seek(10e9);
    const table = await screen.findByRole("table", { name: "Command evidence" });
    expect(within(table).getAllByText("Accepted", { exact: true })).toHaveLength(2);
    expect(within(table).getByText("Refused", { exact: true })).toBeInTheDocument();
    expect(within(table).getByText("Under local control", { exact: true })).toBeInTheDocument();
    expect(within(table).getByText("Unconfirmed", { exact: true })).toBeInTheDocument();
    expect(within(table).getByText("Host observed", { exact: true })).toBeInTheDocument();
    expect(within(table).getAllByText("Effect unknown", { exact: true })).toHaveLength(3);
    expect(within(table).getByRole("columnheader", { name: "Sent at" })).toBeInTheDocument();
    expect(within(table).getAllByText(/· Simulated ·/)).toHaveLength(3);
    expect(within(table).getAllByText(/· Real board ·/)).toHaveLength(1);
    const health = screen.getByRole("region", { name: "Connection health" });
    expect(within(health).getByText("Host: Online", { exact: true })).toBeInTheDocument();
    expect(within(health).getByText("Online · board uno-07 · boot 61", { exact: true })).toBeInTheDocument();
    expect((await axe.run(health)).violations).toEqual([]);
  });

  it("shows stale board readings as stale and never as a normal numeric value", async () => {
    const bridge = reviewBridge();
    const app = mountApp(bridge, reviewHash());
    await screen.findByRole("region", { name: "Run w-1" });
    const wire = liveWire();
    act(() => {
      bridge.stream("w-1", [
        {
          type: "origin",
          seq: 1,
          stream: "observer",
          at_ns: 0,
          instance: "CH-1",
          device: "sensor",
          board: "uno-07",
          environment: "hardware",
          boot_id: "61",
        },
        {
          ...wire,
          devices: [
            {
              ...wire.devices[0],
              instance: "CH-1",
              readings: [{ name: "temperature_c", value: 28, unit: "degC", quality: "stale", state: "stale" }],
            },
          ],
        },
      ]);
      app.timers.advance(250);
      app.stores.review.seek(10e9);
    });
    fireEvent.click(within(screen.getByRole("navigation", { name: "Outline" })).getByRole("button", { name: /^CH-1/ }));
    const readings = screen.getByRole("table", { name: "Board readings" });
    expect(within(readings).getByText("stale", { exact: true })).toBeInTheDocument();
    expect((await axe.run(readings)).violations).toEqual([]);
    expect(within(readings).queryByText("28 degC")).not.toBeInTheDocument();
    expect(screen.getByText(/^This instance is a real board\. These are its own readings/)).toBeInTheDocument();
  });
});
