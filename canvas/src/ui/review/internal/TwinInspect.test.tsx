import { fireEvent, render, screen, waitFor, within } from "@testing-library/react";
import axe from "axe-core";
import { describe, expect, it } from "vitest";
import type { ReviewSession } from "../../../app/review";
import { createReviewSession } from "../../../app/review";
import { FakeBridge, cannedFailure } from "../../../test/fakeBridge";
import type { Canned } from "../../../test/fakeFailures";
import { fakeClock } from "../../../test/fakeClock";
import { failure, success } from "../../../test/outcome";
import { reportWire } from "../../../test/runFixture";
import {
  cannotJudgeWire,
  fieldTraceReportWire,
  originEntryWire,
  sessionEntryWire,
  twinMetricWire,
  twinReportWire,
} from "../../../test/twinReportFixture";
import { TwinInspect } from "./TwinInspect";

interface Options {
  readonly report?: object;
  readonly entity?: string;
}

async function openSaved({ report = fieldTraceReportWire, entity = "CH-1" }: Options = {}) {
  const bridge = new FakeBridge();
  bridge.actions.set("run.report", success(report));
  bridge.actions.set("run.verify", success({ matches: true }));
  bridge.handlers.set("run.twin_report", () => success(twinReportWire()));
  const timers = fakeClock();
  const review = createReviewSession(bridge.review, timers, timers);
  review.open("w-1");
  bridge.stream("w-1", [sessionEntryWire, originEntryWire], "closed");
  timers.advance(250);
  await waitFor(() => {
    expect(review.snapshot().report.phase).toBe("ready");
  });
  review.seek(300e9);
  const view = render(<TwinInspect review={review} state={review.snapshot()} entity={entity} startNs={0} />);
  return { bridge, review, view };
}

const region = () => screen.getByRole("region", { name: "Twin prediction · CH-1" });
const compared = (review: ReviewSession) => review.twin.compare({ run: "w-1", domain: "mini", tolerance: "" });

describe("the twin overlay in Inspect", () => {
  it("says the comparison is not made yet and leads to it, without drawing a prediction", async () => {
    const { review } = await openSaved();
    expect(within(region()).getByText(/^Compare the run with the twin to draw its prediction/)).toBeInTheDocument();
    expect(within(region()).queryByRole("img")).toBeNull();
    fireEvent.click(within(region()).getByRole("button", { name: "Open Twin vs board" }));
    expect(review.snapshot().tab).toBe("report");
  });

  it("draws the twin's prediction over the board's readings, with a legend that does not rely on colour", async () => {
    const { review, view } = await openSaved();
    await compared(review);
    const chart = await screen.findByRole("img", { name: /^temperature: the board's readings as a solid line/ });
    expect(chart).toHaveAccessibleName(/the twin's prediction as a dashed line, 1 stretch where they part ways\.$/);
    expect(chart.querySelectorAll("polyline")).toHaveLength(2);
    expect(chart.querySelectorAll("[class*='shade']")).toHaveLength(1);
    expect(chart.querySelectorAll("[class*='cursor']")).toHaveLength(1);
    const legend = within(screen.getByRole("list", { name: "Legend" }));
    expect(legend.getByText("Board reading")).toBeInTheDocument();
    expect(legend.getByText("Twin prediction")).toBeInTheDocument();
    expect(legend.getByText("Where they part ways")).toBeInTheDocument();
    expect(within(region()).getByText("temperature (Cel)")).toBeInTheDocument();
    expect(within(region()).getByText(/5 readings compared · 2 beyond tolerance · ±2.5 Cel/)).toBeInTheDocument();
    expect((await axe.run(view.container)).violations).toEqual([]);
  });

  it("says when the series shown is only the first part of what was compared", async () => {
    const { bridge, review } = await openSaved();
    const metric = twinMetricWire();
    const series = { ...metric.series, total: 3000, truncated: true };
    bridge.handlers.set("run.twin_report", () => success(twinReportWire({ metrics: [{ ...metric, series }] })));
    await compared(review);
    expect(await screen.findByText("Showing the first 5 of 3000 compared readings.")).toBeInTheDocument();
  });

  it("explains why a metric cannot be judged instead of drawing nothing", async () => {
    const { bridge, review } = await openSaved();
    bridge.handlers.set("run.twin_report", () => success(cannotJudgeWire()));
    await compared(review);
    expect(
      await screen.findByText(
        "The twin cannot judge temperature: readings before the first command was applied to the twin (3).",
      ),
    ).toBeInTheDocument();
    expect(within(region()).queryByRole("img")).toBeNull();
  });

  it("says the report does not cover a board that none of its metrics belong to", async () => {
    const { bridge, review } = await openSaved();
    bridge.handlers.set("run.twin_report", () =>
      success(twinReportWire({ metrics: [twinMetricWire({ device: "other" })], divergences: [] })),
    );
    await compared(review);
    expect(
      await screen.findByText("The twin report does not cover sensor: the field trace holds none of its metrics."),
    ).toBeInTheDocument();
  });

  it("tells a comparison in progress from a refused one, and leads to the report for the words", async () => {
    const { bridge, review } = await openSaved();
    bridge.handlers.set("run.twin_report", () => new Promise<Canned<unknown>>(() => undefined));
    void compared(review);
    expect(await screen.findByText("Comparing the board with the twin…")).toBeInTheDocument();
  });

  it("says the comparison could not be made when the binary refused it", async () => {
    const { bridge, review } = await openSaved();
    bridge.handlers.set("run.twin_report", () => failure(cannedFailure("plantwire_failed", 422, "words")));
    await compared(review);
    expect(await screen.findByText("The comparison could not be made. The Report tab says why.")).toBeInTheDocument();
    expect(within(region()).getByRole("button", { name: "Open Twin vs board" })).toBeInTheDocument();
  });

  it("shows nothing for a board that is simulated, a run with no field trace, or an instance that is not a board", async () => {
    const simulated = await openSaved({ entity: "CT-1" });
    expect(simulated.view.container).toBeEmptyDOMElement();
    simulated.view.unmount();
    const untraced = await openSaved({ report: reportWire });
    expect(untraced.view.container).toBeEmptyDOMElement();
  });
});
