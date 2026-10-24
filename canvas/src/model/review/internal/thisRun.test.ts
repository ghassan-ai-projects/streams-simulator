import { describe, expect, it } from "vitest";
import { grouped } from "../../shared/units";
import { blankReport, report } from "../../../test/runFixture";
import {
  endedGate,
  evidenceWords,
  executionWords,
  noPurposeWords,
  openedReplayWords,
  purposeWords,
  replayStatusWords,
  scoringWords,
  visibilityWords,
} from "./thisRun";

const ended = blankReport({ ...report, startNs: 0, endNs: 1_800e9, requestedEndNs: 1_800e9 });

describe("what the run was for", () => {
  it("shows the purpose as written, and says so when the run states none", () => {
    expect(purposeWords(blankReport({ purpose: "Does the tower recover?" }))).toBe("Does the tower recover?");
    expect(purposeWords(blankReport({ purpose: "  " }))).toBe(noPurposeWords);
    expect(noPurposeWords).toBe("Investigation without a stated question");
  });
});

describe("how the run ended", () => {
  it("names a completed run with the plant time it covered", () => {
    expect(executionWords({ ...ended, termination: { reason: "completed", atNs: 1_800e9 } })).toBe(
      "Completed · 30 min of plant time",
    );
  });

  it("names where an operator stop happened and what was asked", () => {
    const stopped = { ...ended, endNs: 540e9, termination: { reason: "stopped_by_operator" as const, atNs: 540e9 } };
    expect(executionWords(stopped)).toBe("Stopped by operator at 00:09:00 of 00:30:00");
  });

  it("gives the reason a run did not finish, and says a run that has not ended is still open", () => {
    expect(executionWords({ ...ended, complete: false, incompleteReason: "the process was killed" })).toBe(
      "Did not finish: the process was killed",
    );
    expect(executionWords({ ...ended, complete: false, incompleteReason: "" })).toBe(
      "Did not finish: the run did not close.",
    );
    expect(executionWords({ ...ended, ended: false, complete: false })).toBe("Not ended: the run is still open.");
  });
});

describe("what evidence the run left", () => {
  it("counts the observations and says no objective was assessed", () => {
    const lines = evidenceWords({ ...ended, counts: { observed: 1234 } });
    expect(lines).toEqual([
      `Observations: ${grouped(1234)}`,
      "Objective: Not assessed (the run declares no objective)",
    ]);
  });

  it("notes an incomplete record and a journal cut short from the questions the binary asked", () => {
    const asked = [
      { code: "run_incomplete", message: "The run ended incomplete: x. Its evidence is partial." },
      { code: "torn_tail", message: "The last line of journal.jsonl was cut short." },
      { code: "stopped_early", message: "The operator stopped the run." },
    ];
    expect(evidenceWords({ ...ended, counts: { observed: 1 }, openQuestions: asked })).toEqual([
      "Observations: 1 · incomplete record · a journal's last line was cut short",
      "Objective: Not assessed (the run declares no objective)",
    ]);
  });

  it("says nothing was observed rather than leave the count out", () => {
    expect(evidenceWords({ ...ended, counts: {} })[0]).toBe("Observations: none recorded");
    expect(evidenceWords({ ...ended, counts: { observed: 0 } })[0]).toBe("Observations: none recorded");
  });
});

describe("what scored the run", () => {
  it("says no outside system was connected when the report holds no verdict", () => {
    expect(scoringWords(blankReport())).toBe("Not scored: no outside system is connected");
  });

  it("reports the verdict status and reason the binary recorded", () => {
    expect(scoringWords(blankReport({ scoring: { status: "scored", reason: "verdict r-1" } }))).toBe(
      "Scored · verdict r-1",
    );
    expect(scoringWords(blankReport({ scoring: { status: "not_scored", reason: "" } }))).toBe("Not scored");
  });
});

describe("who could see what, and when", () => {
  it("says an Explore run showed its truth while it ran", () => {
    expect(visibilityWords(blankReport({ mode: "explore", ended: true }), "none")).toBe(
      "Truth was visible while the run was going.",
    );
  });

  it("says an Evaluate run hid its truth until it ended", () => {
    expect(visibilityWords(blankReport({ mode: "evaluate", ended: true }), "none")).toBe(
      "Truth stayed hidden from the observer while the run was going and is shown now that it has ended.",
    );
    expect(visibilityWords(blankReport({ mode: "evaluate", ended: false }), "none")).toBe(
      "Truth stays hidden from the observer while the run is going and is shown once it has ended.",
    );
  });

  it("describes a connected session as the simulator's own record", () => {
    expect(visibilityWords(blankReport({ mode: "explore", ended: true }), "saved")).toMatch(
      /^Connected session: .*plant wire host/,
    );
  });

  it("never states two different things about the same run", () => {
    const words = [
      visibilityWords(blankReport({ mode: "explore", ended: true }), "none"),
      visibilityWords(blankReport({ mode: "evaluate", ended: true }), "none"),
    ].join(" ");
    expect(words).not.toMatch(/could not see|true results below/);
  });
});

describe("the replay the report holds", () => {
  it.each([
    ["not_checked", "", "Not checked in this report"],
    ["not_checked", "run not ended", "Not checked in this report: run not ended"],
    ["reproduces", "", "Reproduces"],
    ["does_not_reproduce", "trace differs", "Does not reproduce: trace differs"],
    ["incomplete", "", "Cannot be verified: the run ended incomplete"],
  ] as const)("reads %s %s", (status, detail, words) => {
    expect(replayStatusWords({ status, detail })).toBe(words);
  });
});

describe("whether a run can be exported or compared", () => {
  it("opens for a run that has ended", () => {
    expect(endedGate(ended)).toEqual({ open: true });
  });

  it("closes for an Evaluate run that is not sealed, with the binary's own message", () => {
    const message = "The run has not ended, so its truth is still sealed: this report withholds the faults.";
    const unsealed = blankReport({ mode: "evaluate", openQuestions: [{ code: "evaluate_unsealed", message }] });
    expect(endedGate(unsealed)).toEqual({ open: false, why: message });
  });

  it("closes for an Explore run still open, with the binary's own message", () => {
    const message = "The run has not ended: its record is still open and its results are not final.";
    const open = blankReport({ mode: "explore", openQuestions: [{ code: "run_not_ended", message }] });
    expect(endedGate(open)).toEqual({ open: false, why: message });
  });

  it("still says why when the report carries no such question", () => {
    expect(endedGate(blankReport())).toEqual({ open: false, why: "The run has not ended yet." });
  });
});

describe("the replay this page ran when it opened", () => {
  const ready = (matches: boolean) => ({
    phase: "ready" as const,
    value: { matches, domainChanged: false, detail: "", reason: "", file: "" },
  });

  it("says so when the page's own replay reproduced a run whose report was not replayed", () => {
    expect(openedReplayWords(blankReport(), ready(true))).toBe(
      "The replay this page ran when it opened reproduced the run.",
    );
  });

  it.each([
    ["the page's replay did not reproduce", blankReport(), ready(false)],
    ["the page's replay is still running", blankReport(), { phase: "loading" as const }],
    ["the page's replay failed", blankReport(), { phase: "failed" as const }],
    ["the report holds its own replay", blankReport({ replay: { status: "reproduces", detail: "" } }), ready(true)],
  ])("says nothing when %s", (_name, run, verify) => {
    expect(openedReplayWords(run, verify)).toBe("");
  });
});
