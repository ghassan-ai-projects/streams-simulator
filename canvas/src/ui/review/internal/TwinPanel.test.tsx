import { fireEvent, render, screen, waitFor, within } from "@testing-library/react";
import axe from "axe-core";
import { describe, expect, it } from "vitest";
import { createReviewSession } from "../../../app/review";
import { verdictWords } from "../../../model/review";
import { deferred } from "../../../test/deferred";
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
  twinReportWire,
} from "../../../test/twinReportFixture";
import { TwinPanel } from "./TwinPanel";

interface Options {
  readonly report?: object;
  readonly entries?: readonly object[];
}

async function openSaved({
  report = fieldTraceReportWire,
  entries = [sessionEntryWire, originEntryWire],
}: Options = {}) {
  const bridge = new FakeBridge();
  bridge.actions.set("run.report", success(report));
  bridge.actions.set("run.verify", success({ matches: true }));
  bridge.handlers.set("run.twin_report", () => success(twinReportWire()));
  const timers = fakeClock();
  const review = createReviewSession(bridge.review, timers, timers);
  review.open("w-1");
  bridge.stream("w-1", entries, "closed");
  timers.advance(250);
  await waitFor(() => {
    expect(review.snapshot().report.phase).toBe("ready");
  });
  const view = render(<TwinPanel review={review} state={review.snapshot()} startNs={0} />);
  return { bridge, review, view };
}

const twinCalls = (bridge: FakeBridge) => bridge.calls.filter((call) => call.name === "run.twin_report");
const compare = () => screen.getByRole("button", { name: "Compare with twin" });
const resultGroup = () => screen.findByRole("group", { name: "Twin comparison result" });

describe("the twin comparison in the review", () => {
  it("offers the comparison for a saved session with a field trace, and asks for nothing until it is pressed", async () => {
    const { bridge } = await openSaved();
    const region = screen.getByRole("region", { name: "Twin vs board" });
    expect(within(region).getByLabelText("Tolerances (optional)")).toHaveValue("");
    expect(within(region).getByRole("button", { name: "Compare with twin" })).toHaveAttribute("aria-disabled", "false");
    expect(twinCalls(bridge)).toEqual([]);
  });

  it("shows the verdict by word and shape, the counts, the rule and where the twin and the board part ways", async () => {
    const { bridge, view } = await openSaved();
    fireEvent.click(compare());
    const result = await resultGroup();
    expect(within(result).getByText("Diverges", { exact: true })).toBeInTheDocument();
    expect(twinCalls(bridge)).toEqual([{ name: "run.twin_report", args: { run: "w-1", domain: "mini" } }]);
    expect(within(result).getByRole("heading", { name: verdictWords.diverges.line })).toHaveFocus();
    expect(result.querySelector("svg")).not.toBeNull();
    const counts = within(screen.getByLabelText("Twin comparison counts"));
    expect(counts.getByText("Readings compared").nextSibling).toHaveTextContent("5");
    expect(counts.getByText("Beyond tolerance").nextSibling).toHaveTextContent("2");
    expect(counts.getByText("Commands not applied").nextSibling).toHaveTextContent("2");
    expect(within(result).getByText(/Default tolerance: 5 % of the metric's range\./)).toBeInTheDocument();
    expect(within(result).getByText("Twin: domain mini. No tolerance was typed.")).toBeInTheDocument();
    expect((await axe.run(view.container)).violations).toEqual([]);
  });

  it("lists which commands were applied to the twin and which were not, and why", async () => {
    await openSaved();
    fireEvent.click(compare());
    const list = await screen.findByRole("list", { name: "Commands and the twin" });
    const [applied, refused, twinRefused] = within(list).getAllByRole("listitem");
    expect(applied).toHaveTextContent("cmd-1 · set_mode on sensor · 00:00:50");
    expect(applied).toHaveTextContent("Applied to the twin");
    expect(refused).toHaveTextContent("Not applied: Refused by the board");
    expect(refused).toHaveTextContent("not_ready");
    expect(twinRefused).toHaveTextContent("Refused by the twin");
    expect(twinRefused).toHaveTextContent("unknown effector");
  });

  it("names each divergence with its interval, worst error and commands, and moves the cursor to its start", async () => {
    const { review } = await openSaved();
    fireEvent.click(compare());
    const list = await screen.findByRole("list", { name: "Divergences" });
    expect(list).toHaveTextContent(
      "00:06:40 to 00:08:20 · 2 readings · worst error 8 Cel at 00:08:20 (board 30 Cel, twin 22 Cel)",
    );
    expect(list).toHaveTextContent("Commands in force: cmd-1");
    fireEvent.click(within(list).getByRole("button", { name: "Go to the start of sensor.temperature#1" }));
    expect(review.snapshot().positionNs).toBe(400e9);
  });

  it("gives each metric its tolerance with the source, and the readings it could not compare", async () => {
    await openSaved();
    fireEvent.click(compare());
    const list = await screen.findByRole("list", { name: "Metrics and tolerances" });
    expect(list).toHaveTextContent("sensor.temperature");
    expect(list).toHaveTextContent("5 readings compared · 2 beyond tolerance · ±2.5 Cel (default share of the range)");
    expect(list).toHaveTextContent("Not compared: 2 × readings before the first command was applied to the twin.");
  });

  it("passes the tolerances exactly as typed and says what was typed", async () => {
    const { bridge } = await openSaved();
    const typed = "sensor.temperature=2 , thermal_sensor.humidity=5";
    fireEvent.change(screen.getByLabelText("Tolerances (optional)"), { target: { value: typed } });
    fireEvent.click(compare());
    expect(await screen.findByText(`Twin: domain mini. Tolerances as typed: ${typed}`)).toBeInTheDocument();
    expect(twinCalls(bridge)[0]?.args).toEqual({ run: "w-1", domain: "mini", tolerance: typed });
  });

  it("shows the binary's refusal as it reports it, and moves focus to it", async () => {
    const { bridge } = await openSaved();
    bridge.handlers.set("run.twin_report", () =>
      failure(cannedFailure("plantwire_failed", 422, "invalid tolerance: sensor.speed names no metric")),
    );
    fireEvent.click(compare());
    const alert = await screen.findByRole("alert");
    expect(alert).toHaveTextContent(
      "The comparison could not be made: invalid tolerance: sensor.speed names no metric (plantwire_failed)",
    );
    expect(alert).toHaveFocus();
    expect(screen.queryByRole("group", { name: "Twin comparison result" })).toBeNull();
  });

  it("says plainly that it is comparing, and sends nothing for a second press", async () => {
    const { bridge } = await openSaved();
    const gate = deferred<Canned<unknown>>();
    bridge.handlers.set("run.twin_report", () => gate.promise);
    fireEvent.click(compare());
    expect(await screen.findByText("Comparing the board with the twin…")).toBeInTheDocument();
    expect(compare()).toHaveAttribute("aria-disabled", "true");
    fireEvent.click(compare());
    expect(twinCalls(bridge)).toHaveLength(1);
    gate.resolve(success(twinReportWire()));
    expect(within(await resultGroup()).getByText("Diverges", { exact: true })).toBeInTheDocument();
    expect(screen.queryByText("Comparing the board with the twin…")).toBeNull();
  });

  it("tells a report that cannot judge from one that agrees, with the reason", async () => {
    const { bridge } = await openSaved();
    bridge.handlers.set("run.twin_report", () => success(cannotJudgeWire()));
    fireEvent.click(compare());
    expect(within(await resultGroup()).getByText("Cannot judge", { exact: true })).toBeInTheDocument();
    expect(screen.getByText(/No command the board accepted was applied to the twin/)).toBeInTheDocument();
    expect(screen.getByText("The host sent no command to a real board in this run.")).toBeInTheDocument();
    expect(screen.getByText("Nothing to show: no reading could be compared with the twin.")).toBeInTheDocument();
    expect(screen.queryByRole("list", { name: "Divergences" })).toBeNull();
    bridge.handlers.set("run.twin_report", () =>
      success(twinReportWire({ verdict: "agrees", divergences: [], commands: [], metrics: [] })),
    );
    fireEvent.click(compare());
    await waitFor(() => {
      expect(
        within(screen.getByRole("group", { name: "Twin comparison result" })).getByText("Agrees"),
      ).toBeInTheDocument();
    });
    expect(screen.getByText("No divergence: every compared reading stayed within its tolerance.")).toBeInTheDocument();
    expect(screen.getByText("The board reported no metric the twin could be compared on.")).toBeInTheDocument();
  });

  it("withholds the comparison, with the reason, when the run has no field trace or the domain is unknown", async () => {
    await openSaved({ report: reportWire });
    expect(
      screen.getByText("This run recorded no field trace, so it has no real board to compare."),
    ).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Compare with twin" })).toBeNull();
  });

  it("withholds it when the session record does not name the domain", async () => {
    const nameless = { ...sessionEntryWire, plan: { instances: [] } };
    await openSaved({ entries: [nameless, originEntryWire] });
    expect(screen.getByText("The session record does not name the domain the twin would run.")).toBeInTheDocument();
  });

  it("shows nothing for a run that is not a saved session", async () => {
    const { view } = await openSaved({ entries: [] });
    expect(view.container).toBeEmptyDOMElement();
  });
});
