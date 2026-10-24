import type { OpenQuestion, ReportReplay, ReplayStatus } from "../../../model/review";
import { isRecord, text } from "../../../model/shared/json";

const statuses: readonly ReplayStatus[] = ["not_checked", "reproduces", "does_not_reproduce", "incomplete"];

const notReplayed: ReportReplay = { status: "not_checked", detail: "" };

export function replayOf(value: unknown): ReportReplay | null {
  if (value === undefined) return notReplayed;
  if (!isRecord(value)) return null;
  const status = statuses.find((known) => known === value.status);
  return status === undefined ? null : { status, detail: text(value.detail) };
}

function questionOf(value: unknown): OpenQuestion | null {
  if (!isRecord(value) || typeof value.code !== "string" || typeof value.message !== "string") return null;
  return { code: value.code, message: value.message };
}

export function questionsOf(value: unknown): readonly OpenQuestion[] | null {
  if (value === undefined) return [];
  if (!Array.isArray(value)) return null;
  const read = value.map(questionOf);
  return read.every((question) => question !== null) ? read : null;
}
