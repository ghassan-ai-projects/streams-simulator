import { counted } from "../../shared/words";
import { writableReadings } from "./twinDefaults";
import type { Drive } from "./twinTypes";
import type { KindFile } from "./types";
import type { Choices } from "./words";

export const responseChoices: Choices = [
  ["none", "Has no effect in the twin"],
  ["direct", "Sets a reading straight away"],
  ["first_order", "Sets a reading, and another reading follows gradually"],
];

export const gainHint = "How much the following reading changes for each unit set. Use a negative number to lower it.";

export const timeConstantHint =
  "Seconds the reading takes to cover about two thirds of its way to the new value. Above zero.";

const separator = "/";

function followValue(kind: string, metric: string): string {
  return `${kind}${separator}${metric}`;
}

export function driveFollows(drive: Drive): string {
  return drive.kind === "" && drive.metric === "" ? "" : followValue(drive.kind, drive.metric);
}

export function followTarget(value: string): { readonly kind: string; readonly metric: string } {
  const at = value.indexOf(separator);
  return at < 0 ? { kind: "", metric: "" } : { kind: value.slice(0, at), metric: value.slice(at + 1) };
}

function followable(kind: KindFile): Choices {
  return kind.metrics
    .filter((metric) => metric.name !== "" && !metric.writable)
    .filter((metric) => metric.dataType === "number" || metric.dataType === "integer")
    .map((metric) => [followValue(kind.kind, metric.name), `${kind.kind} · ${metric.name}`] as const);
}

export function followChoices(kinds: readonly KindFile[]): Choices {
  return kinds.flatMap(followable);
}

export function settableChoices(kind: KindFile): Choices {
  return [["", "Not used"], ...writableReadings(kind).map((name) => [name, name] as const)];
}

export function partsWords(channels: number, effectors: number, faults: number): string {
  return [counted(channels, "channel"), counted(effectors, "effector"), counted(faults, "fault")].join(" · ");
}

export function effectWords(drive: Drive): string {
  if (drive.gain === null || drive.timeConstantS === null || drive.from === "" || drive.metric === "") return "";
  const move = `${drive.metric} moves toward its starting value plus ${String(drive.gain)} × ${drive.from}`;
  return `${move}, closing about two thirds of the gap every ${String(drive.timeConstantS)} seconds.`;
}
