import { counted } from "../../../model/workspace";
import type { DeviceFailure } from "../../ports";
import type { DeviceState } from "./deviceDraft";
import { isDirty } from "./deviceSelectors";
import type { SaveStatus } from "./saveStatus";
import type { Built, TwinDraft } from "./twinDraft";
import { isStale, problemsOf } from "./twinSelectors";

function status(text: string, tone: SaveStatus["tone"], detail: string | null = null): SaveStatus {
  return { text, detail, tone };
}

function notSaved(failure: DeviceFailure): SaveStatus {
  if (failure.kind === "conflict") {
    return status(
      "Choices not saved: the file changed on disk. Reload it to continue.",
      "warn",
      failure.diagnostic.message,
    );
  }
  return status(`Choices not saved: ${failure.diagnostic.message}`, "warn");
}

function refused(draft: TwinDraft, failure: DeviceFailure): SaveStatus {
  const issues = draft.issues.length;
  const words = issues > 0 ? counted(issues, "issue") : failure.diagnostic.message;
  return status(`The binary refuses the choices: ${words}.`, "warn");
}

function created(draft: TwinDraft, built: Built): SaveStatus {
  if (isStale(draft))
    return status("Choices changed since the twin was created. Create it again to update it.", "warn");
  const { implemented } = built;
  if (implemented.kind === "checking") return status("Checking the twin…", "busy");
  if (implemented.kind === "passed") return status("Twin ready.", "ok", built.path);
  return status("Twin created, but it does not match its device.", "warn", implemented.failure.diagnostic.message);
}

function idle(draft: TwinDraft): SaveStatus {
  if (problemsOf(draft).length > 0) return status("Finish the choices marked in the form.", "warn");
  return status("Choices are ready. Create the twin when you are.", "");
}

function buildStatus(draft: TwinDraft): SaveStatus {
  const { build } = draft;
  switch (build.kind) {
    case "building":
      return status("Creating the twin…", "busy");
    case "exists":
      return status("A plant with this name already exists. Replace it, or choose another name.", "warn");
    case "refused":
      return refused(draft, build.failure);
    case "failed":
      return status(`The twin could not be created: ${build.failure.diagnostic.message}`, "warn");
    case "built":
      return created(draft, build);
    case "idle":
      return idle(draft);
  }
}

export function twinStatusOf(draft: TwinDraft, device: DeviceState): SaveStatus {
  if (draft.save.kind === "write-failed") return notSaved(draft.save.failure);
  if (isDirty(device)) return status("Save the interface first: the twin is made from the saved device.", "warn");
  return buildStatus(draft);
}
