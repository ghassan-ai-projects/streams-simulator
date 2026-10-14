import { fireEvent, render, screen, within } from "@testing-library/react";
import axe from "axe-core";
import { describe, expect, it, vi } from "vitest";
import { parseReport } from "../../model/report";
import { report } from "../../test/runFixture";
import { ReportPanel } from "./ReportPanel";
import { Scrubber } from "./Scrubber";

describe("report panel", () => {
  it.each([undefined, "kpi"] as const)("cautions rejected incomplete truth in the %s view", (only) => {
    render(
      <ReportPanel
        report={{ ...parseReport(report), complete: false }}
        unit={() => ""}
        entity={null}
        {...(only === undefined ? {} : { only })}
        verify={{
          phase: "ready",
          value: {
            matches: false,
            domainChanged: false,
            detail: "",
            reason: "incomplete_stamp",
            file: "journal.jsonl",
          },
        }}
      />,
    );
    expect(screen.getByRole("note")).toHaveTextContent("This run did not verify");
    expect(screen.getByRole("region", { name: "True KPIs" })).toBeInTheDocument();
  });

  it("lists every section of the binary's report with no accessibility violation", async () => {
    const view = render(
      <ReportPanel
        report={parseReport(report)}
        unit={() => ""}
        entity={null}
        verify={{ phase: "ready", value: { matches: true, domainChanged: false, detail: "", reason: "", file: "" } }}
      />,
    );
    expect(within(screen.getByRole("region", { name: "File check" })).getByText("trace")).toBeInTheDocument();
    expect(within(screen.getByRole("region", { name: "Replay check" })).getByText(/^Reproduces/)).toBeInTheDocument();
    expect(
      within(screen.getByRole("region", { name: "Deliveries" })).getByText("dropped by perturbation"),
    ).toBeInTheDocument();
    expect(within(screen.getByRole("region", { name: "Disturbances" })).getByText("fan failure")).toHaveAttribute(
      "title",
      "tower__fan_failure",
    );
    expect(
      within(screen.getByRole("region", { name: "Effector calls" })).getByText("refused: interlock"),
    ).toBeInTheDocument();
    expect(within(screen.getByRole("region", { name: "True KPIs" })).getByText("4.65 – 8.85")).toBeInTheDocument();
    expect(within(screen.getByRole("region", { name: "True KPIs" })).getByText("5.95")).toBeInTheDocument();
    expect(
      within(screen.getByRole("region", { name: "Scoring" })).getByText("The report has no scoring section."),
    ).toBeInTheDocument();
    expect((await axe.run(view.container)).violations).toEqual([]);
  });

  it("says None for an empty section and shows a scoring result", () => {
    render(
      <ReportPanel
        report={{ ...parseReport(report), effectorCalls: [], scoring: { status: "scored", reason: "" } }}
        unit={() => ""}
        entity={null}
      />,
    );
    expect(within(screen.getByRole("region", { name: "Effector calls" })).getByText("None")).toBeInTheDocument();
    expect(within(screen.getByRole("region", { name: "Scoring" })).getByText("scored")).toBeInTheDocument();
  });

  it("filters KPI truth to the selected instance and states what is not available", () => {
    const { rerender } = render(<ReportPanel report={parseReport(report)} unit={() => ""} entity="CT-1" />);
    expect(
      within(screen.getByRole("region", { name: "True KPIs" })).getByText("No true KPIs for CT-1."),
    ).toBeInTheDocument();
    rerender(
      <ReportPanel
        report={{ ...parseReport(report), ended: false, truthUnavailable: "the run is still sealed" }}
        unit={() => ""}
        entity={null}
      />,
    );
    expect(within(screen.getByRole("region", { name: "Deliveries" })).getAllByText("n/a")).toHaveLength(3);
    expect(
      within(screen.getByRole("region", { name: "Truth" })).getByText("Not available: the run is still sealed"),
    ).toBeInTheDocument();
    expect(
      within(screen.getByRole("region", { name: "Scoring" })).getByText("The report has no scoring section."),
    ).toBeInTheDocument();
  });
});

const idle = { playing: false, speed: 60, onPlay: vi.fn(), onPause: vi.fn(), onSpeed: vi.fn() };

describe("scrubber", () => {
  it("seeks with the slider and jumps to a disturbance", () => {
    const seek = vi.fn();
    render(
      <Scrubber {...idle} startNs={0} endNs={3_600e9} positionNs={60e9} report={parseReport(report)} onSeek={seek} />,
    );
    const slider = screen.getByRole("slider", { name: "Playback time" });
    expect(slider).toHaveAttribute("aria-valuetext", "00:01:00");
    fireEvent.change(slider, { target: { value: "120" } });
    expect(seek).toHaveBeenLastCalledWith(120e9);
    fireEvent.click(screen.getByRole("button", { name: "00:10:00 CT-1 fan failure" }));
    expect(seek).toHaveBeenLastCalledWith(600e9);
    fireEvent.click(screen.getByRole("button", { name: "Next event" }));
    expect(seek).toHaveBeenLastCalledWith(300e9);
    fireEvent.click(screen.getByRole("button", { name: "Forward 10 minutes" }));
    expect(seek).toHaveBeenLastCalledWith(660e9);
    fireEvent.keyDown(slider, { key: "ArrowRight", shiftKey: true });
    expect(seek).toHaveBeenLastCalledWith(120e9);
  });

  it("plays, pauses and picks a speed", () => {
    const controls = { ...idle, onPlay: vi.fn(), onPause: vi.fn(), onSpeed: vi.fn() };
    const view = render(
      <Scrubber {...controls} startNs={0} endNs={60e9} positionNs={0} report={null} onSeek={vi.fn()} />,
    );
    fireEvent.click(screen.getByRole("button", { name: "Play" }));
    expect(controls.onPlay).toHaveBeenCalled();
    expect(screen.getByRole("button", { name: "60×" })).toHaveAttribute("aria-pressed", "true");
    fireEvent.click(screen.getByRole("button", { name: "600×" }));
    expect(controls.onSpeed).toHaveBeenCalledWith(600);
    view.rerender(
      <Scrubber {...controls} playing startNs={0} endNs={60e9} positionNs={0} report={null} onSeek={vi.fn()} />,
    );
    fireEvent.click(screen.getByRole("button", { name: "Pause" }));
    expect(controls.onPause).toHaveBeenCalled();
  });

  it("works before the report arrives", () => {
    render(<Scrubber {...idle} startNs={0} endNs={0} positionNs={0} report={null} onSeek={vi.fn()} />);
    expect(screen.getByRole("list", { name: "Events" })).toBeEmptyDOMElement();
    expect(screen.getByRole("button", { name: "Next event: none" })).toHaveAttribute("aria-disabled", "true");
  });

  it("states verify's own verdict and detail, and cautions that unverified truth may not describe the files", () => {
    render(
      <ReportPanel
        report={parseReport(report)}
        unit={() => ""}
        entity={null}
        verify={{
          phase: "ready",
          value: { matches: false, domainChanged: false, detail: "trace.jsonl differs", reason: "", file: "" },
        }}
      />,
    );
    const replay = screen.getByRole("region", { name: "Replay check" });
    expect(within(replay).getByText(/^Does not verify/)).toBeInTheDocument();
    expect(within(replay).getByText("trace.jsonl differs")).toBeInTheDocument();
    expect(screen.getByRole("note")).toHaveTextContent("This run did not verify");
  });
});
