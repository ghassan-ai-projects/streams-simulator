import { fireEvent, render, screen, within } from "@testing-library/react";
import axe from "axe-core";
import { describe, expect, it, vi } from "vitest";
import { parseReport } from "../../model/report";
import { report } from "../../test/runFixture";
import { ReportPanel } from "./ReportPanel";
import { Scrubber } from "./Scrubber";

describe("report panel", () => {
  it("lists every section of the binary's report with no accessibility violation", async () => {
    const view = render(<ReportPanel report={parseReport(report)} unit={() => ""} entity={null} />);
    expect(within(screen.getByRole("region", { name: "Verification" })).getByText("trace")).toBeInTheDocument();
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
    expect(within(screen.getByRole("region", { name: "KPI truth" })).getByText("4.65 – 8.85")).toBeInTheDocument();
    expect(within(screen.getByRole("region", { name: "KPI truth" })).getByText("5.95")).toBeInTheDocument();
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
      within(screen.getByRole("region", { name: "KPI truth" })).getByText("No KPI truth for CT-1."),
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

describe("scrubber", () => {
  it("seeks with the slider and jumps to a disturbance", () => {
    const seek = vi.fn();
    render(
      <Scrubber
        startNs={0}
        endNs={3_600e9}
        positionNs={60e9}
        report={parseReport(report)}
        density={[1, 3, 2]}
        onSeek={seek}
      />,
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

  it("works before the report arrives", () => {
    render(<Scrubber startNs={0} endNs={0} positionNs={0} report={null} density={[]} onSeek={vi.fn()} />);
    expect(screen.getByRole("list", { name: "Events" })).toBeEmptyDOMElement();
    expect(screen.getByRole("button", { name: "Next event" })).toBeDisabled();
  });
});
