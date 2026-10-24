import { counted } from "../../../model/workspace";
import type { DeviceFailure } from "../../ports";
import type { DeviceState } from "./deviceDraft";
import { isDirty } from "./deviceSelectors";
import type { CheckRow } from "./checkRow";
import type { Built, TwinDraft } from "./twinDraft";
import { calibrationRow, hardwareRow, interfaceRow, plausibilityRow, refusal } from "./claimRows";
import { partialOf } from "./twinScope";
import { builtOf, verdictOf } from "./twinSelectors";

function row(
  id: CheckRow["id"],
  state: CheckRow["state"],
  text: string,
  notes: readonly string[] = [],
  details: readonly string[] = [],
): CheckRow {
  return { id, state, status: "", text, notes, details };
}

function failureRow(id: CheckRow["id"], text: string, failure: DeviceFailure): CheckRow {
  const { notes, details } = refusal(failure);
  return row(id, "bad", text, notes, details);
}

function deviceRow(device: DeviceState): CheckRow {
  if (device.phase !== "ready") return row("device", "pending", "Opening the device definition…");
  if (isDirty(device)) return row("device", "idle", "The device definition has unsaved changes");
  if (device.check.kind === "valid") return row("device", "ok", "The device definition is valid");
  if (device.check.kind === "refused")
    return row("device", "bad", "The device definition has issues", [
      `${counted(device.review.schema.length, "issue")} in the interface step`,
    ]);
  return row("device", "idle", "The device definition has not been checked");
}

function twinRow(draft: TwinDraft): CheckRow {
  const { build } = draft;
  if (build.kind === "building") return row("twin", "pending", "Creating the twin…");
  if (build.kind === "exists") return row("twin", "bad", "A plant with this name already exists");
  if (build.kind === "refused" || build.kind === "failed") {
    return failureRow("twin", "The twin could not be created", build.failure);
  }
  const built = builtOf(draft);
  if (built !== null) return builtRow(built);
  const changed = build.kind === "built";
  return row(
    "twin",
    "idle",
    changed ? "The choices changed since the twin was created" : "No twin has been created yet",
  );
}

function builtRow(built: Built): CheckRow {
  const verdict = verdictOf(built);
  if (verdict === "ready") return row("twin", "ok", "Twin created");
  if (verdict === "pending") return row("twin", "pending", "The twin was created and is being checked");
  return row("twin", "bad", "The twin was created, but it has problems");
}

function scopeRow(draft: TwinDraft): readonly CheckRow[] {
  const partial = partialOf(draft);
  return partial === ""
    ? []
    : [
        {
          id: "scope",
          state: "warn",
          status: "Partial twin",
          text: "Some devices are left out of the twin",
          notes: [partial],
          details: [],
        },
      ];
}

export function checkRows(device: DeviceState, draft: TwinDraft): readonly CheckRow[] {
  const built = builtOf(draft);
  return [
    deviceRow(device),
    twinRow(draft),
    ...scopeRow(draft),
    interfaceRow(built),
    plausibilityRow(built),
    calibrationRow(),
    hardwareRow(),
  ];
}
