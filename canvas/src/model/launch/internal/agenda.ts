import { humanise } from "../../shared/words";
import type { UnitOf } from "./agendaWords";
import { callWords, noUnits } from "./agendaWords";
import type { FaultCategory } from "./perturbation";
import type { LaunchPlan } from "./launch";

export type AgendaLane = "fault" | "clear" | "delivery" | "delivery-end" | "call";

export interface AgendaItem {
  readonly key: string;
  readonly lane: AgendaLane;
  readonly label: string;
  readonly time: string;
  readonly text: string;
  readonly editor: string;
}

type CategoryOf = (entity: string, fault: string) => FaultCategory | null;

interface Entry extends AgendaItem {
  readonly atS: number;
  readonly order: number;
}

const rank: Readonly<Record<AgendaLane, number>> = { fault: 0, clear: 1, delivery: 2, "delivery-end": 3, call: 4 };

const faultLabels: Readonly<Record<FaultCategory, string>> = {
  world: "Process fault",
  measurement: "Sensor fault",
  world_and_measurement: "Process and sensor fault",
};

function timeOf(atS: number): string {
  if (!Number.isFinite(atS)) return "–";
  const whole = Math.max(0, Math.round(atS));
  const pad = (n: number) => String(n).padStart(2, "0");
  const clock = `${pad(Math.floor((whole % 3600) / 60))}:${pad(whole % 60)}`;
  return whole >= 3600 ? `${String(Math.floor(whole / 3600))}:${clock}` : clock;
}

function entry(lane: AgendaLane, index: number, atS: number, label: string, text: string, editor: string): Entry {
  const key = `${lane}-${String(index + 1)}`;
  return {
    key,
    lane,
    label,
    time: timeOf(atS),
    text,
    editor,
    atS: Number.isFinite(atS) ? atS : Infinity,
    order: index,
  };
}

function faultEntries(plan: LaunchPlan, categoryOf: CategoryOf): Entry[] {
  return plan.faults.flatMap((fault, index) => {
    const name =
      fault.entity === "" || fault.fault === "" ? "a fault to choose" : `${fault.entity} ${humanise(fault.fault)}`;
    const category = categoryOf(fault.entity, fault.fault);
    const editor = `launch-fault-${String(index + 1)}`;
    const label = category === null ? "Fault" : faultLabels[category];
    const start = entry("fault", index, fault.atS, label, name, editor);
    if (fault.untilS === undefined) return [start];
    return [start, entry("clear", index, fault.untilS, "Clear", `${name} ends`, `${editor}-end`)];
  });
}

function paramWords(params: Readonly<Record<string, unknown>>): string {
  const parts = Object.entries(params).map(([name, value]) => `${humanise(name)} ${String(value)}`);
  return parts.length === 0 ? "" : ` (${parts.join(", ")})`;
}

function deliveryEntries(plan: LaunchPlan): Entry[] {
  return plan.perturbations.flatMap((planned, index) => {
    const editor = `launch-delivery-${String(index + 1)}`;
    if (planned.kind === "")
      return [entry("delivery", index, planned.atS, "Delivery", "a delivery condition to choose", editor)];
    const name = humanise(planned.kind);
    const start = entry(
      "delivery",
      index,
      planned.atS,
      "Delivery",
      `${name} starts${paramWords(planned.params)}`,
      editor,
    );
    if (planned.untilS === undefined) return [start];
    return [start, entry("delivery-end", index, planned.untilS, "Delivery", `${name} ends`, `${editor}-end`)];
  });
}

function callEntries(plan: LaunchPlan, unitOf: UnitOf): Entry[] {
  return plan.calls.map((call, index) =>
    entry("call", index, call.atS, "Command", callWords(call, unitOf), `launch-call-${String(index + 1)}`),
  );
}

export function agendaOf(plan: LaunchPlan, categoryOf: CategoryOf, unitOf: UnitOf = noUnits): readonly AgendaItem[] {
  const entries = [...faultEntries(plan, categoryOf), ...deliveryEntries(plan), ...callEntries(plan, unitOf)];
  const sorted = entries.sort((a, b) => a.atS - b.atS || rank[a.lane] - rank[b.lane] || a.order - b.order);
  return sorted.map(({ key, lane, label, time, text, editor }) => ({ key, lane, label, time, text, editor }));
}
