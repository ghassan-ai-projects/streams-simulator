import type { BoardCapture, CapturedKind } from "../../../app/ports";
import { isRecord, text } from "../../../model/shared/json";
import { rows } from "./decoders";
import { issueList } from "./http";

function comparisonOf(value: unknown): CapturedKind["comparison"] | null {
  if (value === undefined) return { status: "new" };
  if (!isRecord(value)) return null;
  const file = text(value.file);
  const fileDigest = text(value.file_digest);
  if (value.status === "new") return { status: "new" };
  if (value.status === "same" && file !== "") return { status: "same", file, fileDigest };
  if (value.status === "differs" && file !== "") {
    return { status: "differs", file, fileDigest, differsAt: text(value.differs_at) };
  }
  return null;
}

function capturedKind(entry: Record<string, unknown>): CapturedKind | null {
  const comparison = comparisonOf(entry.comparison);
  const { draft } = entry;
  const kind = text(entry.kind);
  const digest = text(entry.kind_digest);
  if (comparison === null || !isRecord(draft) || kind === "" || digest === "") return null;
  return { kind, digest, draft, issues: issueList(entry.issues), comparison };
}

function environmentOf(value: unknown): BoardCapture["environment"] | null {
  return value === "hardware" || value === "simulated" ? value : null;
}

export function boardCapture(value: unknown): BoardCapture | null {
  if (!isRecord(value)) return null;
  const environment = environmentOf(value.environment);
  const { bd_seq: bdSeq } = value;
  const kinds = rows(value.kinds, capturedKind);
  if (environment === null || kinds === null) return null;
  if (typeof bdSeq !== "number" || !Number.isInteger(bdSeq) || bdSeq < 0) return null;
  return { environment, bdSeq, kinds };
}
