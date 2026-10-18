import type { DeviceState } from "./deviceDraft";
import { isDirty } from "./deviceSelectors";
import type { TwinState } from "./twinDraft";
import { hasUnsavedChoices } from "./twinSelectors";

export function hasUnsavedWork(device: DeviceState, twin: TwinState): boolean {
  return isDirty(device) || hasUnsavedChoices(twin);
}
