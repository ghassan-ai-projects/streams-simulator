import { runClock } from "./timeline";
import { humanise } from "../../shared/words";
import type { OperatorCall } from "./callLog";
import type { Report } from "./report";
import type { Span } from "./trend";

export interface OverlayEvent {
  readonly key: string;
  readonly kind: "effector" | "fault";
  readonly atNs: number;
  readonly entity: string;
  readonly refused: boolean;
  readonly label: string;
}

const sameCallNs = 1e9;

export interface PlannedFault {
  readonly entity: string;
  readonly fault: string;
  readonly atS: number;
}

export interface Planned {
  readonly startNs: number;
  readonly faults: readonly PlannedFault[];
}

const noPlan: Planned = { startNs: 0, faults: [] };

export function truthAllowed(report: Report | null): boolean {
  if (report?.truthUnavailable !== "") return false;
  return report.mode !== "evaluate" || report.ended;
}

function callEvent(call: OperatorCall): OverlayEvent {
  const outcome = call.accepted ? "accepted" : "refused";
  return {
    key: `call:${String(call.atNs)}:${call.entity}:${call.effector}`,
    kind: "effector",
    atNs: call.atNs,
    entity: call.entity,
    refused: !call.accepted,
    label: `${call.entity} ${humanise(call.effector)} · ${outcome}`,
  };
}

function reportedCalls(report: Report | null): readonly OperatorCall[] {
  if (!truthAllowed(report)) return [];
  return (report?.effectorCalls ?? []).map((call) => ({
    atNs: call.atNs,
    entity: call.entity,
    effector: call.effector,
    accepted: !call.outcome.startsWith("refused") && call.outcome !== "",
    reason: "",
  }));
}

function merged(calls: readonly OperatorCall[], reported: readonly OperatorCall[]): readonly OperatorCall[] {
  const extra = reported.filter(
    (entry) =>
      !calls.some(
        (call) =>
          call.entity === entry.entity &&
          call.effector === entry.effector &&
          Math.abs(call.atNs - entry.atNs) < sameCallNs,
      ),
  );
  return [...calls, ...extra];
}

function faultEvents(report: Report | null): readonly OverlayEvent[] {
  if (!truthAllowed(report)) return [];
  return (report?.disturbances ?? [])
    .filter((entry) => entry.kind === "fault")
    .map((fault) => ({
      key: `fault:${String(fault.onsetNs)}:${fault.target}:${fault.name}`,
      kind: "fault" as const,
      atNs: fault.onsetNs,
      entity: fault.target,
      refused: false,
      label: `${fault.target} ${humanise(fault.name)}`,
    }));
}

function plannedEvents(planned: Planned, reported: readonly OverlayEvent[]): readonly OverlayEvent[] {
  return planned.faults
    .map((fault) => ({
      key: `fault:${String(planned.startNs + fault.atS * 1e9)}:${fault.entity}:${fault.fault}`,
      kind: "fault" as const,
      atNs: planned.startNs + fault.atS * 1e9,
      entity: fault.entity,
      refused: false,
      label: `${fault.entity} ${humanise(fault.fault)}`,
    }))
    .filter(
      (event) =>
        !reported.some((known) => known.entity === event.entity && Math.abs(known.atNs - event.atNs) < sameCallNs),
    );
}

export function overlayOf(
  calls: readonly OperatorCall[],
  report: Report | null,
  span: Span,
  planned: Planned = noPlan,
): readonly OverlayEvent[] {
  const faults = faultEvents(report);
  const all = [...merged(calls, reportedCalls(report)).map(callEvent), ...faults, ...plannedEvents(planned, faults)];
  return all.filter((event) => event.atNs >= span.startNs && event.atNs <= span.endNs).sort((a, b) => a.atNs - b.atNs);
}

export function eventLine(event: OverlayEvent, startNs: number): string {
  const what = event.kind === "fault" ? "Fault starts" : "Effector call";
  return `${runClock(event.atNs, startNs)} · ${what} · ${event.label}`;
}
