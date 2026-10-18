import type { DeviceFailure } from "../../ports";

export const unreadable: DeviceFailure = {
  kind: "unreadable",
  diagnostic: { displayCode: "kind_unreadable", message: "The file is not a device definition." },
};
