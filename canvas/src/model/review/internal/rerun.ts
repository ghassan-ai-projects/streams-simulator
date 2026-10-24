import type { LaunchPlan } from "../../launch";
import { freshPlanOf } from "../../launch";
import { clockOf } from "../../shared/duration";
import { counted } from "../../shared/words";
import type { Report } from "./report";

export interface Rerun {
  readonly plan: LaunchPlan;
  readonly left: readonly string[];
  readonly fresh: boolean;
}

const minimumS = 60;

export function rerunOf(report: Report): Rerun {
  if (!recoverable(report)) return { plan: freshPlanOf(report.mode === "explore"), left: [], fresh: true };
  return { plan: planOfRun(report), left: leftOut(report), fresh: false };
}

function recoverable(report: Report): boolean {
  return report.ended && report.complete && report.seed !== null;
}

function secondsFrom(report: Report, ns: number): number {
  return Math.round((ns - report.startNs) / 1e9);
}

function lengthS(report: Report): number {
  return secondsFrom(report, report.requestedEndNs ?? report.endNs);
}

function authored(report: Report) {
  return {
    ...(report.name === "" ? {} : { name: report.name }),
    ...(report.purpose === "" ? {} : { purpose: report.purpose }),
  };
}

function planOfRun(report: Report): LaunchPlan {
  return {
    ...authored(report),
    seed: report.seed ?? 0,
    durationS: Math.max(minimumS, lengthS(report)),
    explore: report.mode === "explore",
    pace: report.rate !== null && report.rate > 0 ? "watch" : "fast",
    faults: report.disturbances.filter((d) => d.kind === "fault").map((fault) => plannedFault(report, fault)),
    perturbations: report.disturbances.filter(appliedDelivery).map((applied) => plannedDelivery(report, applied)),
    calls: report.effectorCalls.map((call) => ({
      entity: call.entity,
      effector: call.effector,
      atS: Math.max(0, secondsFrom(report, call.atNs)),
      args: call.args,
    })),
  };
}

type Disturbance = Report["disturbances"][number];

function plannedFault(report: Report, fault: Disturbance): LaunchPlan["faults"][number] {
  const until = fault.clearedBy === "command" ? fault.clearedNs : null;
  return {
    entity: fault.target,
    fault: fault.name,
    atS: Math.max(0, secondsFrom(report, fault.onsetNs)),
    ...(until === null ? {} : { untilS: secondsFrom(report, until) }),
  };
}

function appliedDelivery(disturbance: Disturbance): boolean {
  return disturbance.kind === "perturbation" && disturbance.action === "apply";
}

function plannedDelivery(report: Report, applied: Disturbance): LaunchPlan["perturbations"][number] {
  return {
    kind: applied.name,
    params: applied.params ?? {},
    atS: Math.max(0, secondsFrom(report, applied.onsetNs)),
    ...(applied.untilNs === undefined ? {} : { untilS: secondsFrom(report, applied.untilNs) }),
  };
}

function count(n: number, one: string, many: string): string {
  return n === 0 ? "" : counted(n, one, many);
}

function leftOut(report: Report): readonly string[] {
  const environment = report.disturbances.filter((d) => d.kind === "environment").length;
  return [
    lengthS(report) < minimumS ? "a length under one minute (one minute is used)" : "",
    count(environment, "environment change", "environment changes"),
    unrecordedAfterStop(report),
  ].filter((part) => part !== "");
}

function unrecordedAfterStop(report: Report): string {
  const { termination } = report;
  if (termination?.reason !== "stopped_by_operator") return "";
  const at = clockOf((termination.atNs - report.startNs) / 1e9);
  return `anything planned after the stop at ${at}, which the run never recorded`;
}
