import { act, fireEvent, screen, within } from "@testing-library/react";
import { beforeEach, describe, expect, it } from "vitest";
import { mountApp } from "../../../test/appHarness";
import { stubFlowDom } from "../../../test/dom";
import { success } from "../../../test/outcome";
import { reviewBridge, reviewHash } from "../../../test/screenFixture";
import {
  fieldTraceReportWire,
  originEntryWire,
  sessionEntryWire,
  twinReportWire,
} from "../../../test/twinReportFixture";

beforeEach(() => {
  stubFlowDom();
});

describe("the twin versus board report in a reviewed session", () => {
  it("compares the run with the twin from the report tab and draws the prediction in Inspect", async () => {
    const bridge = reviewBridge();
    bridge.actions.set("run.report", success(fieldTraceReportWire));
    bridge.handlers.set("run.twin_report", () => success(twinReportWire()));
    const app = mountApp(bridge, reviewHash());
    await screen.findByRole("region", { name: "Run w-1" });
    act(() => {
      bridge.stream("w-1", [sessionEntryWire, originEntryWire]);
      app.timers.advance(250);
    });
    fireEvent.click(await screen.findByRole("button", { name: "Compare with twin" }));
    const result = await screen.findByRole("group", { name: "Twin comparison result" });
    expect(within(result).getByText("Diverges")).toBeInTheDocument();
    fireEvent.click(
      within(await screen.findByRole("list", { name: "Divergences" })).getByRole("button", {
        name: "Go to the start of sensor.temperature#1",
      }),
    );
    expect(app.stores.review.snapshot().positionNs).toBe(400e9);
    fireEvent.click(within(screen.getByRole("navigation", { name: "Outline" })).getByRole("button", { name: /^CH-1/ }));
    const overlay = await screen.findByRole("region", { name: "Twin prediction · CH-1" });
    expect(within(overlay).getByText("Twin prediction")).toBeInTheDocument();
    expect(within(overlay).getByRole("img", { name: /^temperature: the board's readings/ })).toBeInTheDocument();
  });

  it("shows no twin comparison for a run that is not a session", async () => {
    const bridge = reviewBridge();
    bridge.actions.set("run.report", success(fieldTraceReportWire));
    mountApp(bridge, reviewHash());
    await screen.findByRole("region", { name: "This run" });
    expect(screen.queryByRole("button", { name: "Compare with twin" })).toBeNull();
    expect(screen.queryByRole("region", { name: "Twin vs board" })).toBeNull();
  });
});
