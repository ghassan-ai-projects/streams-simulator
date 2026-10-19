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

const receiptWords: Readonly<Record<SessionCommand["receipt"], string>> = {
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

const reasonWords: ReadonlyMap<string, string> = new Map(
  Object.entries({
    not_ready: "Not ready",
    local_control: "Refused at the node: local control",
    stale_epoch: "Stale authority epoch",
    expired: "Expired before it was approved or accepted",
    policy_not_allowed: "Policy not allowed",
    session_ending: "Session ending",
    unknown_device: "The node has no such device",
    unknown_board: "The node has no such board",
    unknown_command:
      "That command no longer waits, or its content is not what was shown: it may have been decided or expired",
    too_many_scopes: "Too many holds and approval scopes are set; release one first",
    node_busy: "The node was busy and may still apply this later",
    malformed: "The node could not read the request",
    duplicate: "A command with different content already uses this command id or key",
  }),
);

export function refusalWords(reason: string): string {
  return reasonWords.get(reason) ?? reason.replaceAll("_", " ");
}

export function acceptanceWords(command: SessionCommand): string {
  if (command.receipt === "refused" && command.reason === "local_control") return refusalWords(command.reason);
  return receiptWords[command.receipt];
}

const policyRefusalSides: Readonly<Record<string, string>> = {
  real: "Policy not allowed by the gateway",
  simulated: "Policy not allowed by this session",
};

export function acceptanceDetail(command: SessionCommand): string {
  if (command.reason === "" || (command.receipt === "refused" && command.reason === "local_control")) return "";
  if (command.reason === "policy_not_allowed") {
    return policyRefusalSides[originKindOf(command.environment)] ?? refusalWords(command.reason);
  }
  return refusalWords(command.reason);
}

export const copyFailedWords = "Could not copy; select the digest instead.";

export function copiedWords(commandId: string): string {
  return `Copied the policy digest of ${commandId}.`;
}

export function digestListedWords(domain: string): string {
  return `The digest is in the draft session plan of ${domain}. It applies to the next session; a running session keeps its list.`;
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
  const boot = device.health === "simulated" ? "" : `boot ${device.bootId}`;
  return [healthWords[device.health], `board ${device.board}`, boot, reboots].filter((part) => part !== "").join(" · ");
}

const staleUpdateMs = 30_000;

export function updateAgeWords(ageMs: number): string {
  const seconds = Math.max(0, Math.floor(ageMs / 1000));
  return ageMs >= staleUpdateMs
    ? `No live snapshot for ${String(seconds)} s`
    : `Last live snapshot ${String(seconds)} s ago`;
}
