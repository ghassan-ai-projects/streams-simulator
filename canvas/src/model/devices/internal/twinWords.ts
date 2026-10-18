import { counted, spaced } from "../../shared/words";
import { writableReadings } from "./twinMembers";
import type { Drive } from "./twinTypes";
import type { KindFile, Member } from "./types";
import type { Choices } from "./words";

export const responseChoices: Choices = [
  ["none", "Has no effect in the twin"],
  ["direct", "Sets a reading straight away"],
  ["first_order", "Sets a reading, and another reading follows gradually"],
];

export const gainHint = "How much the following reading changes for each unit set. Use a negative number to lower it.";

export const timeConstantHint =
  "Seconds the reading takes to cover about two thirds of its way to the new value. Above zero.";

export const noSetpointHint = "Mark a reading as a setpoint and give it a range in the Interface step";

export function argumentLabel(argument: string): string {
  return `Reading set by the ${spaced(argument)} value`;
}

export const noFollowerHint = "No reading can follow yet. Add a second reading in the Interface step.";

const separator = "@";

function followValue(path: string, key: string): string {
  return `${key}${separator}${path}`;
}

export function driveFollows(drive: Drive): string {
  return drive.kind === "" && drive.metric === "" ? "" : followValue(drive.kind, drive.metric);
}

export function followTarget(value: string): { readonly kind: string; readonly metric: string } {
  const at = value.indexOf(separator);
  return at < 0 ? { kind: "", metric: "" } : { kind: value.slice(at + 1), metric: value.slice(0, at) };
}

function followable(kind: KindFile): KindFile["metrics"] {
  return kind.metrics
    .filter((metric) => metric.name !== "" && !metric.writable)
    .filter((metric) => metric.dataType === "number" || metric.dataType === "integer");
}

export function followChoices(members: readonly Member[]): Choices {
  const all = members.flatMap(({ path, kind }) => followable(kind).map((metric) => ({ path, kind, metric })));
  const names = all.map((entry) => entry.metric.name);
  return all.map(({ path, kind, metric }) => {
    const shared = names.filter((name) => name === metric.name).length > 1;
    return [followValue(path, metric.key), shared ? `${metric.name} (${kind.kind})` : metric.name] as const;
  });
}

export function settableChoices(kind: KindFile): Choices {
  return [["", "Not used"], ...writableReadings(kind).map((metric) => [metric.key, metric.name] as const)];
}

export function partsWords(channels: number, effectors: number, faults: number): string {
  return [counted(channels, "channel"), counted(effectors, "effector"), counted(faults, "fault")].join(" · ");
}

function readingName(members: readonly Member[], path: string, key: string): string {
  const kind = members.find((member) => member.path === path)?.kind;
  return kind?.metrics.find((metric) => metric.key === key)?.name ?? "";
}

export const slowTimeWords = "Choose a time constant above zero.";

export function effectWords(drive: Drive, path: string, members: readonly Member[]): string {
  if (drive.timeConstantS !== null && drive.timeConstantS <= 0) return slowTimeWords;
  if (drive.gain === null || drive.timeConstantS === null || drive.from === "" || drive.metric === "") return "";
  const from = readingName(members, path, drive.from);
  const metric = readingName(members, drive.kind, drive.metric);
  const move = `${metric} moves toward its starting value plus ${String(drive.gain)} × ${from}`;
  return `${move}, closing about two thirds of the gap every ${String(drive.timeConstantS)} seconds.`;
}
