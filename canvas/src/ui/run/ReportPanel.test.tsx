import { fireEvent, render, screen, within } from "@testing-library/react";
import axe from "axe-core";
import { describe, expect, it, vi } from "vitest";
import { parseReport } from "../../model/report";
import { report } from "../../test/runFixture";
import { ReportPanel } from "./ReportPanel";
import { Scrubber } from "./Scrubber";

describe("report panel", () => {
  it("lists every section of the binary's report with no accessibility violation", async () => {
    const view = render(<ReportPanel report={parseReport(report)} />);
    expect(within(screen.getByRole("region", { name: "Verification" })).getByText("trace")).toBeInTheDocument();
    expect(
      within(screen.getByRole("region", { name: "Deliveries" })).getByText("dropped by perturbation"),
    ).toBeInTheDocument();
    expect(
      within(screen.getByRole("region", { name: "Disturbances" })).getByText("tower__fan_failure"),
    ).toBeInTheDocument();
    expect(
      within(screen.getByRole("region", { name: "Effector calls" })).getByText("refused: interlock"),
    ).toBeInTheDocument();
    expect(within(screen.getByRole("region", { name: "KPI truth" })).getByText("4.65 – 8.85")).toBeInTheDocument();
    expect(
      within(screen.getByRole("region", { name: "Scoring" })).getByText(/no consumer verdict/),
    ).toBeInTheDocument();
    expect((await axe.run(view.container)).violations).toEqual([]);
  });

  it("says None for an empty section and shows a scoring result", () => {
    render(<ReportPanel report={{ ...parseReport(report), effectorCalls: [], scoring: "scored" }} />);
    expect(within(screen.getByRole("region", { name: "Effector calls" })).getByText("None")).toBeInTheDocument();
    expect(within(screen.getByRole("region", { name: "Scoring" })).getByText("scored")).toBeInTheDocument();
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
    fireEvent.click(screen.getByRole("button", { name: "Jump to 00:10:00, CT-1 tower__fan_failure" }));
    expect(seek).toHaveBeenLastCalledWith(600e9);
  });

  it("works before the report arrives", () => {
    render(<Scrubber startNs={0} endNs={0} positionNs={0} report={null} density={[]} onSeek={vi.fn()} />);
    expect(screen.queryAllByRole("button")).toEqual([]);
  });
});
