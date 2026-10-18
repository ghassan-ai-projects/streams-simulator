import type { BoardCapture, CapturedKind } from "../../../app/ports";
import { isRecord, text } from "../../../model/shared/json";
import { rows } from "./decoders";
import { issueList } from "./http";

type Comparison = CapturedKind["comparison"];
type LaterBirth = BoardCapture["laterBirths"][number];

function count(value: unknown, least: number): number | null {
  return typeof value === "number" && Number.isInteger(value) && value >= least ? value : null;
}

function comparisonOf(value: unknown): Comparison | null {
  if (!isRecord(value)) return null;
  const [target, file] = [text(value.target), text(value.file)];
  if (target === "") return null;
  if (value.status === "new") return { status: "new", target };
  if (file === "") return null;
  if (value.status === "unreadable") return { status: "unreadable", target, file };
  const fileDigest = text(value.file_digest);
  if (value.status === "same") return { status: "same", target, file, fileDigest };
  if (value.status !== "differs") return null;
  return {
    status: "differs",
    target,
    file,
    fileDigest,
    differsAt: text(value.differs_at),
    collidesWith: text(value.collides_with),
  };
}

function strings(value: unknown): readonly string[] | null {
  if (!Array.isArray(value)) return null;
  const found = value.filter((entry): entry is string => typeof entry === "string");
  return found.length === value.length ? found : null;
}

function capturedKind(entry: Record<string, unknown>): CapturedKind | null {
  const comparison = comparisonOf(entry.comparison);
  const claims = strings(entry.claims);
  const { draft } = entry;
  const kind = text(entry.kind);
  const digest = text(entry.kind_digest);
  if (comparison === null || claims === null || !isRecord(draft) || kind === "" || digest === "") return null;
  return { kind, digest, draft, claims, issues: issueList(entry.issues), comparison };
}

function environmentOf(value: unknown): BoardCapture["environment"] | null {
  return value === "hardware" || value === "simulated" ? value : null;
}

function laterBirth(entry: Record<string, unknown>): LaterBirth | null {
  const [line, bdSeq] = [count(entry.line, 1), count(entry.bd_seq, 0)];
  if (line === null || bdSeq === null || typeof entry.same_content !== "boolean") return null;
  return { line, bdSeq, sameContent: entry.same_content };
}

function takenOf(value: unknown): BoardCapture["taken"] | null {
  if (!isRecord(value)) return null;
  const [line, bdSeq] = [count(value.line, 1), count(value.bd_seq, 0)];
  return line === null || bdSeq === null ? null : { line, bdSeq };
}

export function boardCapture(value: unknown): BoardCapture | null {
  if (!isRecord(value)) return null;
  const environment = environmentOf(value.environment);
  const [bdSeq, birthsSeen] = [count(value.bd_seq, 0), count(value.births_seen, 1)];
  const [taken, kinds] = [takenOf(value.taken), rows(value.kinds, capturedKind)];
  const later = value.later_births === undefined ? [] : rows(value.later_births, laterBirth);
  if (environment === null || bdSeq === null || birthsSeen === null) return null;
  if (taken === null || kinds === null || later === null) return null;
  return { environment, bdSeq, birthsSeen, taken, laterBirths: later, issues: issueList(value.issues), kinds };
}
