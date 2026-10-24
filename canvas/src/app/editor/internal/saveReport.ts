import { digestWords, plausibilityWords } from "../../../model/editor";
import type { EditorFailure } from "../../ports";
import type { Draft, Plausibility } from "./editorDraft";

export function failureWords(failure: EditorFailure): string {
  return `${failure.diagnostic.message} (${failure.diagnostic.displayCode})`;
}

export interface ReportRow {
  readonly id: "composition" | "digest" | "plausibility" | "interface";
  readonly label: string;
  readonly state: "ok" | "warn" | "bad" | "pending" | "neutral";
  readonly text: string;
  readonly title: string;
  readonly notes: readonly string[];
}

function row(id: ReportRow["id"], label: string, state: ReportRow["state"], text: string): ReportRow {
  return { id, label, state, text, title: "", notes: [] };
}

function findingWords(finding: { severity: string; entity: string; detail: string }): string {
  return `${finding.severity}: ${finding.entity}: ${finding.detail}`;
}

export function plausibilityRow(plausibility: Plausibility): ReportRow {
  if (plausibility.kind === "running") return row("plausibility", "Plausibility", "pending", "Checking…");
  if (plausibility.kind === "failed") {
    return row("plausibility", "Plausibility", "warn", `Could not be checked: ${plausibility.message}`);
  }
  const { errors, warnings, findings } = plausibility.report;
  const words = plausibilityWords(errors, warnings);
  return { ...row("plausibility", "Plausibility", words.state, words.text), notes: findings.map(findingWords) };
}

export function saveReportOf(draft: Draft): readonly ReportRow[] | null {
  const { save } = draft;
  if (save.kind !== "composed") return null;
  return [
    row("composition", "Composition", "ok", save.summary),
    { ...row("digest", "Digest", "neutral", digestWords(save.digest)), title: save.digest },
    plausibilityRow(save.plausibility),
    row("interface", "Interface", "neutral", "Not applicable: a composed plant has no device interface to check"),
  ];
}
