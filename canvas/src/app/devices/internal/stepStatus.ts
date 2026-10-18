import type { DeviceState } from "./deviceDraft";
import { isDirty } from "./deviceSelectors";
import type { TwinState } from "./twinDraft";
import { builtOf, isBuilding, problemsOf } from "./twinSelectors";

export type Step = "interface" | "behaviour" | "check";

export interface StepStatus {
  readonly text: string;
  readonly tone: "ok" | "warn" | "busy" | "";
}

function state(text: string, tone: StepStatus["tone"]): StepStatus {
  return { text, tone };
}

function interfaceStatus(device: DeviceState): StepStatus {
  if (device.phase !== "ready") return state("Opening", "busy");
  if (isDirty(device)) return state("Unsaved", "warn");
  const { check } = device;
  if (check.kind === "valid") return state("Valid", "ok");
  if (check.kind === "refused") return state("Has issues", "warn");
  return check.kind === "checking" ? state("Checking", "busy") : state("Not checked", "");
}

function behaviourStatus(twin: TwinState): StepStatus {
  if (twin.phase === "loading") return state("Opening", "busy");
  if (twin.phase !== "ready") return state("Not started", "");
  if (isBuilding(twin)) return state("Creating", "busy");
  if (builtOf(twin) !== null) return state("Created", "ok");
  if (["refused", "failed", "exists"].includes(twin.build.kind)) return state("Needs attention", "warn");
  if (problemsOf(twin).length > 0) return state("Needs choices", "warn");
  return twin.touched ? state("Ready", "") : state("Not started", "");
}

function checkStatus(twin: TwinState): StepStatus {
  const built = twin.phase === "ready" ? builtOf(twin) : null;
  if (built === null) return state("Not run", "");
  const { implemented } = built;
  if (implemented.kind === "checking") return state("Checking", "busy");
  return implemented.kind === "passed" ? state("Passed", "ok") : state("Failed", "warn");
}

export function stepStatuses(device: DeviceState, twin: TwinState): Readonly<Record<Step, StepStatus>> {
  return { interface: interfaceStatus(device), behaviour: behaviourStatus(twin), check: checkStatus(twin) };
}
