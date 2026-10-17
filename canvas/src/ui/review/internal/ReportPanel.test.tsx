import { render, screen, within } from "@testing-library/react";
import axe from "axe-core";
import { describe, expect, it, vi } from "vitest";
import { report } from "../../../test/runFixture";
import { ReportPanel } from "./ReportPanel";

describe("report panel", () => {
  it("opens with the integrity block collapsed, on every review", () => {
    const view = render(
      <ReportPanel
        report={report}
        unit={() => ""}
        entity={null}
        verify={{ phase: "ready", value: { matches: true, domainChanged: false, detail: "", reason: "", file: "" } }}
      />,
    );
    const integrity = view.container.querySelector("details");
    expect(integrity).not.toBeNull();
    expect(integrity).not.toHaveAttribute("open");
    expect(integrity).toHaveTextContent("The files match and the replay reproduces.");
    expect(integrity).not.toHaveTextContent(/Verified:/);
  });

  it.each([undefined, "kpi"] as const)("cautions rejected incomplete truth in the %s view", (only) => {
    render(
      <ReportPanel
        report={{ ...report, complete: false }}
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
    expect(screen.getByRole("region", { name: "True results" })).toBeInTheDocument();
  });

  it("lists every section of the binary's report with no accessibility violation", async () => {
    const view = render(
      <ReportPanel
        report={report}
        unit={() => ""}
        entity={null}
        verify={{ phase: "ready", value: { matches: true, domainChanged: false, detail: "", reason: "", file: "" } }}
      />,
    );
    expect(within(screen.getByRole("region", { name: "Replay check" })).getByText(/^Reproduces/)).toBeInTheDocument();
    const fileCheck = screen.getByRole("region", { name: "File check" });
    expect(within(fileCheck).getByText("Files match the record.")).toBeInTheDocument();
    const files = within(fileCheck).getByText("For developers").closest("details");
    if (files === null) throw new Error("file check has no For developers disclosure");
    expect(within(files).getByText("trace")).toBeInTheDocument();
    expect(within(files).getByText(/compares the files in the run folder/)).toBeInTheDocument();
    const replay = screen.getByRole("region", { name: "Replay check" });
    const command = within(replay).getByText("For developers").closest("details");
    if (command === null) throw new Error("replay check has no For developers disclosure");
    expect(within(command).getByText(/streamsim verify command/)).toBeInTheDocument();
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
    expect(within(screen.getByRole("region", { name: "True results" })).getByText("4.65 – 8.85")).toBeInTheDocument();
    expect(within(screen.getByRole("region", { name: "True results" })).getByText("5.95")).toBeInTheDocument();
    expect(
      within(screen.getByRole("region", { name: "Scoring" })).getByText(
        "Not scored: no outside system is connected to this run.",
      ),
    ).toBeInTheDocument();
    expect((await axe.run(view.container)).violations).toEqual([]);
  });

  it("says None for an empty section and shows a scoring result", () => {
    render(
      <ReportPanel
        report={{ ...report, effectorCalls: [], scoring: { status: "scored", reason: "" } }}
        unit={() => ""}
        entity={null}
      />,
    );
    expect(within(screen.getByRole("region", { name: "Effector calls" })).getByText("None")).toBeInTheDocument();
    expect(within(screen.getByRole("region", { name: "Scoring" })).getByText("scored")).toBeInTheDocument();
  });

  it("marks the disturbance row the picked mark stands for, and no row otherwise", () => {
    Element.prototype.scrollIntoView = vi.fn();
    const parsed = report;
    const first = parsed.disturbances[0];
    if (first === undefined) throw new Error("fixture has no disturbance");
    const picked = { atNs: first.onsetNs, entity: first.target };
    const { rerender } = render(<ReportPanel report={parsed} unit={() => ""} entity={null} picked={picked} />);
    const rows = within(screen.getByRole("region", { name: "Disturbances" })).getAllByRole("row");
    expect(rows.filter((row) => row.getAttribute("aria-current") === "true")).toHaveLength(1);
    rerender(<ReportPanel report={parsed} unit={() => ""} entity={null} picked={{ ...picked, entity: "other" }} />);
    expect(document.querySelectorAll("[aria-current='true']")).toHaveLength(0);
  });

  it("filters KPI truth to the selected instance and states what is not available", () => {
    const { rerender } = render(<ReportPanel report={report} unit={() => ""} entity="CT-1" />);
    expect(
      within(screen.getByRole("region", { name: "True results" })).getByText("No true results for CT-1."),
    ).toBeInTheDocument();
    rerender(
      <ReportPanel
        report={{ ...report, ended: false, truthUnavailable: "the run is still sealed" }}
        unit={() => ""}
        entity={null}
      />,
    );
    expect(within(screen.getByRole("region", { name: "Deliveries" })).getAllByText("n/a")).toHaveLength(3);
    expect(
      within(screen.getByRole("region", { name: "True results" })).getByText("Not available: the run is still sealed"),
    ).toBeInTheDocument();
    expect(
      within(screen.getByRole("region", { name: "Scoring" })).getByText(
        "Not scored: no outside system is connected to this run.",
      ),
    ).toBeInTheDocument();
  });

  it("states verify's own verdict and detail, and cautions that unverified truth may not describe the files", () => {
    render(
      <ReportPanel
        report={report}
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
