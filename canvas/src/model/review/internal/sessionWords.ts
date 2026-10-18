import { originLabels } from "../../shared/origin";
import type { DeviceHealth, LinkState, SessionCommand, SessionDevice } from "./sessionLive";
import { originKindOf } from "./sessionLive";

export const linkWords: Readonly<Record<LinkState, string>> = {
  online: "Online",
  offline: "Offline",
  unknown: "Unknown",
  not_applicable: "No gateway",
};

const healthWords: Readonly<Record<DeviceHealth, string>> = {
  online: "Online",
  offline: "Offline",
  simulated: "Simulated · no link",
};

export const receiptWords: Readonly<Record<SessionCommand["receipt"], string>> = {
  pending: "Unconfirmed",
  accepted: "Accepted",
  refused: "Refused",
};

export const reportWords: Readonly<Record<SessionCommand["report"], string>> = {
  pending: "No board report",
  succeeded: "Board reports success",
  failed: "Board reports failure",
  unknown: "Board report unknown",
};

const reasonWords: Readonly<Record<string, string>> = {
  not_ready: "Not ready",
  local_control: "Under local control",
  stale_epoch: "Stale authority epoch",
  expired: "Expired",
  policy_not_allowed: "Policy not allowed",
  session_ending: "Session ending",
  unknown_device: "Unknown device",
};

export function refusalWords(reason: string): string {
  return reasonWords[reason] ?? reason.replaceAll("_", " ");
}

export function effectWords(command: SessionCommand): string {
  return command.effect === "observed" ? "Host observed" : "Effect unknown";
}

export function commandSide(command: SessionCommand): string {
  const kind = originKindOf(command.environment);
  return kind === "real" ? "Real board" : originLabels[kind];
}

function rebootWords(count: number): string {
  if (count <= 0) return "";
  return count === 1 ? "1 reboot" : `${String(count)} reboots`;
}

export function deviceLine(device: SessionDevice): string {
  const reboots = rebootWords(device.reboots);
  return [healthWords[device.health], `board ${device.board}`, `boot ${device.bootId}`, reboots]
    .filter((part) => part !== "")
    .join(" · ");
}

const staleUpdateMs = 30_000;

export function updateAgeWords(ageMs: number): string {
  const seconds = Math.max(0, Math.floor(ageMs / 1000));
  return ageMs >= staleUpdateMs ? `No update for ${String(seconds)} s` : `Last update ${String(seconds)} s ago`;
}
