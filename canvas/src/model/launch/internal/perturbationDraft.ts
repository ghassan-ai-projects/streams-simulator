import type { PerturbationKind, PerturbParams, PerturbParamValue, PlannedPerturbation } from "./perturbation";
import { minutesText } from "./minutes";
import { kindNamed, paramSpecOf } from "./perturbation";

export interface DraftPerturbation {
  readonly key: number;
  readonly kind: string;
  readonly atMinutes: string;
  readonly untilMinutes: string;
  readonly values: Readonly<Record<string, string>>;
}

export function minutesWhole(text: string): number {
  return Math.round(Number(text.trim() === "" ? Number.NaN : text) * 60);
}

export function endOf(untilMinutes: string): { readonly untilS?: number } {
  return untilMinutes.trim() === "" ? {} : { untilS: minutesWhole(untilMinutes) };
}

function guessed(text: string): PerturbParamValue {
  return Number.isFinite(Number(text)) ? Number(text) : text;
}

const splitList = (text: string) =>
  text
    .split(",")
    .map((part) => part.trim())
    .filter((part) => part !== "");

const typed: Readonly<Record<string, (text: string) => PerturbParamValue>> = {
  number: Number,
  integer: Number,
  enum: (text) => text,
  string_list: splitList,
};

function valueOf(text: string, type: string | undefined): PerturbParamValue {
  return (typed[type ?? ""] ?? guessed)(text);
}

function paramsOf(values: Readonly<Record<string, string>>, kind: PerturbationKind | undefined): PerturbParams {
  const typed = Object.entries(values)
    .map(([name, text]) => [name, text.trim()] as const)
    .filter(([, text]) => text !== "")
    .map(([name, text]) => [name, valueOf(text, paramSpecOf(kind, name)?.type)] as const);
  return Object.fromEntries(typed);
}

export function plannedPerturbation(draft: DraftPerturbation, kinds: readonly PerturbationKind[]): PlannedPerturbation {
  return {
    kind: draft.kind,
    params: paramsOf(draft.values, kindNamed(kinds, draft.kind)),
    atS: minutesWhole(draft.atMinutes),
    ...endOf(draft.untilMinutes),
  };
}

function textOf(value: PerturbParamValue): string {
  return typeof value === "object" ? value.join(", ") : String(value);
}

export function perturbationDraftOf(planned: PlannedPerturbation, key: number): DraftPerturbation {
  return {
    key,
    kind: planned.kind,
    atMinutes: minutesText(planned.atS),
    untilMinutes: planned.untilS === undefined ? "" : minutesText(planned.untilS),
    values: Object.fromEntries(Object.entries(planned.params).map(([name, value]) => [name, textOf(value)])),
  };
}
