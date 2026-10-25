import type { DeviceLineage, Lineage } from "./lineageTypes";

export type DeviceStatus = "invalid" | "needs-twin" | "twin-ready";

export interface DeviceBadge {
  readonly words: string;
  readonly tone: "unfinished" | "badge" | "verified";
}

function deviceAt(lineage: Lineage | null, path: string): DeviceLineage | null {
  return lineage?.devices.find((device) => device.file === path) ?? null;
}

function includedIn(device: DeviceLineage | null): string {
  const names = device?.twinsIncluding.map((entry) => entry.twin) ?? [];
  const [first = "", ...rest] = names;
  if (first === "") return "";
  return rest.length === 0 ? first : `${first} and ${String(rest.length)} more`;
}

export function deviceBadge(status: DeviceStatus, path: string, lineage: Lineage | null): DeviceBadge {
  if (status === "invalid") return { words: "Not valid", tone: "unfinished" };
  const device = deviceAt(lineage, path);
  if (status === "twin-ready" || device?.hasOwnTwin === true) return { words: "Twin created", tone: "verified" };
  const included = includedIn(device);
  if (included !== "") return { words: `Included in twin ${included}`, tone: "verified" };
  return { words: "Needs a twin", tone: "badge" };
}
