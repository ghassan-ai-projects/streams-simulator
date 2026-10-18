import { resolveIssues } from "../../../model/devices";
import type { KindFile } from "../../../model/devices";
import type { DeviceFailure } from "../../ports";

export const unreadable: DeviceFailure = {
  kind: "unreadable",
  diagnostic: { displayCode: "kind_unreadable", message: "The file is not a device definition." },
};

export function refusalWords(kind: KindFile, failure: DeviceFailure): string {
  if (failure.kind !== "refused" || failure.issues.length === 0) return failure.diagnostic.message;
  return resolveIssues(kind, failure.issues)
    .map((issue) => issue.text)
    .join(" ");
}
