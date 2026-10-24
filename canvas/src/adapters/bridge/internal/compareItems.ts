import type {
  ComparedCall,
  ComparedEnvironment,
  ComparedFault,
  ComparedPerturbation,
  Diff,
  Paired,
} from "../../../model/review";
import { isRecord, record, text } from "../../../model/shared/json";
import { outcomeText } from "./reportDecoder";

type Item<T> = (value: unknown) => T | null;

function instant(value: unknown): number | null {
  return typeof value === "number" && Number.isFinite(value) ? value : null;
}

function strings(value: unknown): Readonly<Record<string, string>> {
  const shown = Object.entries(record(value)).flatMap(([name, each]): [string, string][] => {
    if (typeof each === "string" || typeof each === "number" || typeof each === "boolean")
      return [[name, String(each)]];
    return Array.isArray(each) && each.every((entry) => typeof entry === "string") ? [[name, each.join(", ")]] : [];
  });
  return Object.fromEntries(shown);
}

export const faultItem: Item<ComparedFault> = (value) => {
  const item = record(value);
  const onsetNs = instant(item.onset_ns);
  if (typeof item.entity_id !== "string" || typeof item.fault !== "string" || onsetNs === null) return null;
  return {
    entity: item.entity_id,
    fault: item.fault,
    onsetNs,
    clearedNs: instant(item.cleared_ns),
    clearedBy: text(item.cleared_by),
  };
};

export const perturbationItem: Item<ComparedPerturbation> = (value) => {
  const item = record(value);
  const atNs = instant(item.at_ns);
  if (typeof item.action !== "string" || atNs === null) return null;
  return {
    action: item.action,
    kind: text(item.perturbation) || text(item.perturb_id),
    atNs,
    fromNs: instant(item.from_ns) ?? 0,
    untilNs: instant(item.until_ns) ?? 0,
    params: strings(item.params),
  };
};

export const environmentItem: Item<ComparedEnvironment> = (value) => {
  const item = record(value);
  const atNs = instant(item.at_ns);
  if (typeof item.target !== "string" || typeof item.fault !== "string" || atNs === null) return null;
  return { atNs, target: item.target, fault: item.fault };
};

export const callItem: Item<ComparedCall> = (value) => {
  const item = record(value);
  const atNs = instant(item.at_ns);
  if (typeof item.effector !== "string" || atNs === null) return null;
  const outcome = isRecord(item.outcome) ? outcomeText(item.outcome) : "";
  return { atNs, entity: text(item.entity_id), effector: item.effector, args: strings(item.args), outcome };
};

function itemsOf<T>(value: unknown, item: Item<T>): readonly T[] | null {
  if (!Array.isArray(value)) return null;
  const read = value.map(item);
  return read.every((each) => each !== null) ? read : null;
}

function pairOf<T>(value: unknown, item: Item<T>): Paired<T> | null {
  const pair = record(value);
  const a = item(pair.a);
  const b = item(pair.b);
  const differences = Array.isArray(pair.differences) ? pair.differences.filter((n) => typeof n === "string") : null;
  return a === null || b === null || differences === null ? null : { a, b, differences };
}

export function diffOf<T>(value: unknown, item: Item<T>): Diff<T> | null {
  const lane = record(value);
  const identical = itemsOf(lane.identical, item);
  const onlyInA = itemsOf(lane.only_in_a, item);
  const onlyInB = itemsOf(lane.only_in_b, item);
  const different = itemsOf(lane.different, (pair) => pairOf(pair, item));
  if (identical === null || onlyInA === null || onlyInB === null || different === null) return null;
  return { identical, different, onlyInA, onlyInB };
}
