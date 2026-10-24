import { fireEvent, render, screen, within } from "@testing-library/react";
import axe from "axe-core";
import { describe, expect, it, vi } from "vitest";
import { blankReport, report } from "../../../test/runFixture";
import { noReplayControl } from "./OpenQuestions";
import { ReportPanel } from "./ReportPanel";
import { ThisRun } from "./ThisRun";

const verified = {
  phase: "ready" as const,
  value: { matches: true, domainChanged: false, detail: "", reason: "", file: "" },
};
const idle = { phase: "idle" as const };
const pending = { phase: "loading" as const };

function shown(overrides: Parameters<typeof blankReport>[0] = {}, extra: Partial<Parameters<typeof ThisRun>[0]> = {}) {
  const run = blankReport({ ...report, startNs: 0, endNs: 1_800e9, counts: { observed: 412 }, ...overrides });
  return render(
    <ThisRun report={run} verify={verified} connection="none" replay={noReplayControl} handover={idle} {...extra} />,
  );
}

describe("the This run block", () => {
  it("states the four assessments as separate labelled lines, in order", () => {
    shown({ name: "Cooling loss", purpose: "Does the tower recover?" });
    const region = screen.getByRole("region", { name: "This run" });
    expect(within(region).getByText("Cooling loss")).toBeInTheDocument();
    expect(within(region).getByText("Does the tower recover?")).toBeInTheDocument();
    const terms = within(region)
      .getAllByRole("term")
      .map((term) => term.textContent);
    expect(terms).toEqual(["Execution", "Recording integrity", "Evidence", "Scoring"]);
    expect(region.querySelector("[data-part='execution']")).toHaveTextContent("Completed · 30 min of plant time");
    expect(region.querySelector("[data-part='recording-integrity']")).toHaveTextContent(
      "The files match and the replay reproduces.",
    );
    expect(region.querySelector("[data-part='evidence']")).toHaveTextContent("Observations: 412");
    expect(region.querySelector("[data-part='evidence']")).toHaveTextContent("Objective: Not assessed");
    expect(region.querySelector("[data-part='scoring']")).toHaveTextContent(
      "Not scored: no outside system is connected",
    );
  });

  it("says a run with no purpose is an investigation without a stated question", () => {
    shown({ purpose: "" });
    expect(screen.getByText("Investigation without a stated question")).toBeInTheDocument();
  });

  it("shows the purpose as text, never as markup", () => {
    shown({ purpose: "<b>bold</b> why" });
    expect(screen.getByText("<b>bold</b> why")).toBeInTheDocument();
  });

  it("names where an operator stop happened and the consumer's verdict when there is one", () => {
    shown({
      termination: { reason: "stopped_by_operator", atNs: 540e9 },
      requestedEndNs: 1_800e9,
      scoring: { status: "scored", reason: "verdict r-1" },
    });
    expect(screen.getByText("Stopped by operator at 00:09:00 of 00:30:00")).toBeInTheDocument();
    expect(screen.getByText("Scored · verdict r-1")).toBeInTheDocument();
  });

  it("states the replay held by the report next to the integrity", () => {
    shown({ replay: { status: "reproduces", detail: "" } });
    expect(screen.getByText("Replay in this report: Reproduces")).toBeInTheDocument();
  });

  it("says the replay this page ran on opening reproduced a run whose report was not replayed", () => {
    shown({ replay: { status: "not_checked", detail: "" } });
    expect(screen.getByText("Replay checked when this page opened: reproduces")).toBeInTheDocument();
    expect(screen.getByText("Replay in this report: Not checked in this report")).toBeInTheDocument();
  });

  it("says one thing about the replay: the page's own check, and not that nobody has shown it", () => {
    const asked = { code: "replay_not_checked", message: "The recording was not replayed." };
    shown({ replay: { status: "not_checked", detail: "" }, openQuestions: [asked] });
    const region = screen.getByRole("region", { name: "This run" });
    expect(region).toHaveTextContent("Replay checked when this page opened: reproduces");
    expect(region).not.toHaveTextContent("nobody has shown");
    expect(region.querySelector("[data-code='replay_not_checked']")).toBeNull();
    expect(screen.getByText("Nothing open.")).toBeInTheDocument();
    expect(region.querySelector("[data-part='opened-replay']")).toContainElement(
      screen.getByRole("button", { name: "Check replay now" }),
    );
  });

  it("keeps the binary's question when the page's own replay did not match", () => {
    const asked = { code: "replay_not_checked", message: "The recording was not replayed." };
    const differs = { ...verified, value: { ...verified.value, matches: false } };
    shown({ replay: { status: "not_checked", detail: "" }, openQuestions: [asked] }, { verify: differs });
    expect(screen.getByText(asked.message)).toBeInTheDocument();
    expect(screen.queryByText(/Replay checked when this page opened/)).toBeNull();
  });

  it("states who could see what once, from the mode and whether the run has ended", () => {
    shown({ mode: "explore" });
    expect(screen.getAllByText(/Truth (was|stayed|stays)/)).toHaveLength(1);
    expect(screen.getByText("Truth was visible while the run was going.")).toBeInTheDocument();
  });
});

describe("the open questions", () => {
  it("says nothing is open when the binary asked nothing", () => {
    shown({ openQuestions: [] });
    const questions = within(screen.getByRole("region", { name: "Open questions" }));
    expect(questions.getByText("Nothing open.")).toBeInTheDocument();
    expect(questions.queryByRole("button", { name: "Check replay now" })).toBeNull();
  });

  it("lists the binary's questions word for word, each with its stable code", () => {
    shown({
      openQuestions: [
        { code: "stopped_early", message: "The operator stopped the run at 00:09:00 of the requested 00:30:00." },
        { code: "no_outside_scoring", message: "This is an unscored investigation." },
      ],
    });
    const items = within(screen.getByRole("region", { name: "Open questions" })).getAllByRole("listitem");
    expect(items.map((item) => item.getAttribute("data-code"))).toEqual(["stopped_early", "no_outside_scoring"]);
    expect(items[0]).toHaveTextContent("The operator stopped the run at 00:09:00 of the requested 00:30:00.");
  });

  it("offers Check replay now only beside the question that the replay was not checked", () => {
    const onCheck = vi.fn();
    shown(
      {
        openQuestions: [
          { code: "stopped_early", message: "stopped" },
          { code: "replay_not_checked", message: "The recording was not replayed." },
        ],
      },
      { replay: { check: idle, onCheck }, verify: pending },
    );
    const button = screen.getByRole("button", { name: "Check replay now" });
    expect(button.closest("li")).toHaveAttribute("data-code", "replay_not_checked");
    fireEvent.click(button);
    expect(onCheck).toHaveBeenCalledOnce();
  });

  it("shows the replay running, and the binary's failure word for word", () => {
    const open = [{ code: "replay_not_checked", message: "The recording was not replayed." }];
    const { rerender } = shown(
      { openQuestions: open },
      { replay: { check: { phase: "checking" }, onCheck: vi.fn() }, verify: pending },
    );
    expect(screen.getByRole("status")).toHaveTextContent("Replaying the run…");
    expect(screen.getByRole("button", { name: "Check replay now" })).toBeDisabled();
    rerender(
      <ThisRun
        report={blankReport({ ...report, openQuestions: open })}
        verify={pending}
        connection="none"
        replay={{ check: { phase: "failed", message: "the recording is unreadable" }, onCheck: vi.fn() }}
        handover={idle}
      />,
    );
    expect(screen.getByRole("alert")).toHaveTextContent("the recording is unreadable");
  });
});

describe("the handover notice", () => {
  it("says in the block where the file went, or why it did not, and leaves announcing to the button's own note", () => {
    const { rerender } = shown({}, { handover: { phase: "saved", run: "r-1", name: "r-1-handover.md" } });
    expect(screen.getByText("Handover saved as r-1-handover.md.")).toBeInTheDocument();
    rerender(
      <ThisRun
        report={blankReport(report)}
        verify={verified}
        connection="none"
        replay={noReplayControl}
        handover={{ phase: "failed", run: "r-1", message: "run r-1 has not ended" }}
      />,
    );
    expect(screen.getByText("run r-1 has not ended")).toBeInTheDocument();
  });
});

describe("the page says who could see what once", () => {
  it.each(["explore", "evaluate"])("has no contradictory truth sentence for an %s run", (mode) => {
    render(<ReportPanel report={{ ...report, mode }} unit={() => ""} entity={null} verify={verified} />);
    const text = document.body.textContent;
    expect(text).not.toContain("the true results below were visible while it ran");
    expect(text).not.toContain("An outside system could not see it while the run was live");
    expect(text).not.toContain("hidden from the outside system while it ran");
    expect(screen.getAllByText(/^Truth (was|stayed) /)).toHaveLength(1);
  });
});

describe("an Evaluate run that is not sealed", () => {
  const unsealed = blankReport({
    ...report,
    mode: "evaluate",
    ended: false,
    complete: false,
    termination: null,
    openQuestions: [{ code: "evaluate_unsealed", message: "The run has not ended, so its truth is still sealed." }],
  });

  it("exposes nothing sealed even if the report object carried it", () => {
    render(<ReportPanel report={unsealed} unit={() => ""} entity={null} verify={verified} />);
    const text = document.body.textContent;
    expect(text).not.toMatch(/fan failure|heat wave|tower__fan_failure|interlock|4\.65|5\.95/);
    expect(screen.queryByRole("region", { name: "Effector calls" })).toBeNull();
    expect(screen.queryByRole("region", { name: "Disturbances" })).toBeNull();
    expect(screen.getByRole("region", { name: "True results" })).toHaveTextContent("Not available yet");
    expect(
      screen.getByText("Truth stays hidden from the observer while the run is going and is shown once it has ended."),
    ).toBeInTheDocument();
  });

  it("repeats the binary's own message among the open questions", () => {
    render(<ReportPanel report={unsealed} unit={() => ""} entity={null} verify={verified} />);
    const item = document.querySelector("[data-code='evaluate_unsealed']");
    expect(item).toHaveTextContent("The run has not ended, so its truth is still sealed.");
  });
});

describe("accessibility", () => {
  it("has no violation with open questions and a replay button", async () => {
    const view = shown(
      { openQuestions: [{ code: "replay_not_checked", message: "The recording was not replayed." }] },
      { verify: pending },
    );
    expect((await axe.run(view.container)).violations).toEqual([]);
  });
});
