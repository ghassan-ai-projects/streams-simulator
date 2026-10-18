import { counted } from "../../../model/workspace";
import type { DeviceFailure, LintReport } from "../../ports";
import type { DeviceState } from "./deviceDraft";
import { isDirty } from "./deviceSelectors";
import type { Built, TwinDraft } from "./twinDraft";
import { builtOf } from "./twinSelectors";

export interface CheckRow {
  readonly id: "device" | "twin" | "implements" | "lint";
  readonly state: "ok" | "bad" | "pending" | "idle";
  readonly text: string;
  readonly notes: readonly string[];
}

const shown = 6;

function row(id: CheckRow["id"], state: CheckRow["state"], text: string, notes: readonly string[] = []): CheckRow {
  return { id, state, text, notes };
}

function failureNotes(failure: DeviceFailure): readonly string[] {
  if (failure.kind !== "refused" || failure.issues.length === 0) return [failure.diagnostic.message];
  return failure.issues.map((issue) => issue.message);
}

function deviceRow(device: DeviceState): CheckRow {
  if (device.phase !== "ready") return row("device", "pending", "Opening the device definition…");
  if (isDirty(device)) return row("device", "idle", "The device definition has unsaved changes");
  if (device.check.kind === "valid") return row("device", "ok", "The device definition is valid");
  if (device.check.kind === "refused")
    return row("device", "bad", "The device definition has issues", [
      `${counted(device.issues.length, "issue")} in the interface step`,
    ]);
  return row("device", "idle", "The device definition has not been checked");
}

function twinRow(draft: TwinDraft): CheckRow {
  const { build } = draft;
  if (build.kind === "building") return row("twin", "pending", "Creating the twin…");
  if (build.kind === "exists") return row("twin", "bad", "A plant with this name already exists");
  if (build.kind === "refused" || build.kind === "failed") {
    return row("twin", "bad", "The twin could not be created", failureNotes(build.failure));
  }
  if (builtOf(draft) !== null) return row("twin", "ok", "Twin ready");
  const changed = build.kind === "built";
  return row(
    "twin",
    "idle",
    changed ? "The choices changed since the twin was created" : "No twin has been created yet",
  );
}

function implementsRow(built: Built | null): CheckRow {
  if (built === null) return row("implements", "idle", "Not checked yet");
  const { implemented } = built;
  if (implemented.kind === "checking")
    return row("implements", "pending", "Checking that the twin matches its device…");
  if (implemented.kind === "passed") return row("implements", "ok", "The twin matches its device");
  return row("implements", "bad", "The twin does not match its device", failureNotes(implemented.failure));
}

function findings(report: LintReport): readonly string[] {
  const lines = report.findings.slice(0, shown).map((finding) => `${finding.entity}: ${finding.detail}`);
  const more = report.findings.length - shown;
  return more > 0 ? [...lines, `…and ${String(more)} more`] : lines;
}

function lintRow(built: Built | null): readonly CheckRow[] {
  if (built === null || built.lint.kind === "unavailable") return [];
  if (built.lint.kind === "running") return [row("lint", "pending", "Checking that the readings stay plausible…")];
  const { report } = built.lint;
  if (report.errors > 0) {
    return [row("lint", "bad", `The plausibility check found ${counted(report.errors, "error")}`, findings(report))];
  }
  const text =
    report.warnings > 0
      ? `The plausibility check has ${counted(report.warnings, "warning")}`
      : "The readings stay plausible";
  return [row("lint", "ok", text, findings(report))];
}

export function checkRows(device: DeviceState, draft: TwinDraft): readonly CheckRow[] {
  const built = builtOf(draft);
  return [deviceRow(device), twinRow(draft), implementsRow(built), ...lintRow(built)];
}
