import { durationOf } from "../../shared/duration";
import { grouped } from "../../shared/units";
import { spaced } from "../../shared/words";
import type { Connection } from "./origins";
import type { OpenQuestion, Report, ReportReplay, VerifyState } from "./report";
import { terminationWords } from "./reportViews";

export const noPurposeWords = "Investigation without a stated question";

export function purposeWords(report: Report): string {
  return report.purpose.trim() === "" ? noPurposeWords : report.purpose;
}

export function executionWords(report: Report): string {
  if (!report.ended) return "Not ended: the run is still open.";
  if (!report.complete) return `Did not finish: ${report.incompleteReason || "the run did not close."}`;
  const stopped = terminationWords(report);
  if (stopped !== "") return stopped;
  const span = durationOf(Math.round((report.endNs - report.startNs) / 1e9));
  return span === "" ? "Completed" : `Completed · ${span} of plant time`;
}

const evidenceNotes: Readonly<Record<string, string>> = {
  evidence_foreign: "evidence from another run",
  run_incomplete: "incomplete record",
  no_observations: "nothing observed",
  torn_tail: "a journal's last line was cut short",
  no_kpi_results: "no whole-run results",
};

function observationWords(report: Report): string {
  const observed = report.counts.observed ?? 0;
  return observed === 0 ? "Observations: none recorded" : `Observations: ${grouped(observed)}`;
}

function noteWords(questions: readonly OpenQuestion[]): readonly string[] {
  return questions.flatMap((question) => {
    const note = evidenceNotes[question.code];
    return note === undefined || question.code === "no_observations" ? [] : [note];
  });
}

export function evidenceWords(report: Report): readonly string[] {
  const facts = [observationWords(report), ...noteWords(report.openQuestions)];
  return [facts.join(" · "), "Objective: Not assessed (the run declares no objective)"];
}

export function scoringWords(report: Report): string {
  if (report.scoring === null) return "Not scored: no outside system is connected";
  const status = spaced(report.scoring.status);
  const named = status.charAt(0).toUpperCase() + status.slice(1);
  return report.scoring.reason === "" ? named : `${named} · ${report.scoring.reason}`;
}

const connectedWords =
  "Connected session: the true results are the simulator's own record of the plant while it was linked to the plant wire host.";

export function visibilityWords(report: Report, connection: Connection): string {
  if (connection !== "none") return connectedWords;
  if (report.mode === "explore") return "Truth was visible while the run was going.";
  return report.ended
    ? "Truth stayed hidden from the observer while the run was going and is shown now that it has ended."
    : "Truth stays hidden from the observer while the run is going and is shown once it has ended.";
}

export function replayStatusWords(replay: ReportReplay): string {
  switch (replay.status) {
    case "reproduces":
      return "Reproduces";
    case "does_not_reproduce":
      return `Does not reproduce: ${replay.detail}`;
    case "incomplete":
      return "Cannot be verified: the run ended incomplete";
    case "not_checked":
      return replay.detail === "" ? "Not checked in this report" : `Not checked in this report: ${replay.detail}`;
  }
}

export type EndedGate = { readonly open: true } | { readonly open: false; readonly why: string };

const unendedCodes = ["evaluate_unsealed", "run_not_ended", "evidence_foreign"];

export function endedGate(report: Report): EndedGate {
  if (report.ended) return { open: true };
  const asked = report.openQuestions.find((question) => unendedCodes.includes(question.code));
  return { open: false, why: asked?.message ?? "The run has not ended yet." };
}

export function openedReplayWords(report: Report, verify: VerifyState): string {
  const reproduced = verify.phase === "ready" && verify.value.matches;
  if (report.replay.status !== "not_checked" || !reproduced) return "";
  return "The replay this page ran when it opened reproduced the run.";
}
