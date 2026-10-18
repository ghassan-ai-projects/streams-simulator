import { resolveIssues } from "../../../model/devices";
import type { KindFile } from "../../../model/devices";
import type { DeviceFailure } from "../../ports";

export const unreadable: DeviceFailure = {
  kind: "unreadable",
  diagnostic: { displayCode: "kind_unreadable", message: "The file is not a device definition." },
};

export function conflictWords(failure: DeviceFailure, way: "create" | "replace", name: string): DeviceFailure {
  if (failure.kind !== "conflict") return failure;
  const message =
    way === "create"
      ? `A device named ${name} already exists. Capture again to compare it with the saved one.`
      : `The saved device ${name} changed since it was compared. Capture again to compare it.`;
  return { ...failure, diagnostic: { ...failure.diagnostic, message } };
}

export function refusalWords(kind: KindFile, failure: DeviceFailure): string {
  if (failure.kind !== "refused" || failure.issues.length === 0) return failure.diagnostic.message;
  return resolveIssues(kind, failure.issues)
    .map((issue) => issue.text)
    .join(" ");
}
