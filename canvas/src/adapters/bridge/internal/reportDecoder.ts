import type { Replay, Report } from "../../../model/review";
import type { AlarmActivation } from "../../../model/review";
import { isRecord, list, number, record, text } from "../../../model/shared/json";

export function parseReportPayload(value: unknown): Report | null {
  if (!isRecord(value) || !isRecord(value.run) || !isRecord(value.counts)) {
    return null;
  }
  if (!isRecord(value.verification) || typeof value.verification.status !== "string") return null;
  return decodeReport(value);
}

function decodeReport(value: Record<string, unknown>): Report {
  const truth = record(value.truth);
  const counts = record(value.counts);
  const verification = record(value.verification);
  const kpis = record(truth.kpis);
  return {
    ...runFacts(record(value.run)),
    verification: text(verification.status),
    checks: list(verification.checks).map((check) => ({
      name: text(record(check).name),
      status: text(record(check).status),
    })),
    counts: numbers(counts),
    dropped: numbers(counts.dropped_by_reason),
    disturbances: [...faults(truth), ...perturbations(truth), ...environment(truth)].sort(
      (a, b) => a.onsetNs - b.onsetNs,
    ),
    effectorCalls: effectorCalls(truth),
    kpis: kpiRows(truth),
    alarms: numbers(kpis.alarm_system),
    alarmActivations: alarmActivations(kpis.entities),
    scoring: scoring(value.scoring),
    truthUnavailable: text(value.truth_unavailable),
  };
}

function runFacts(run: Record<string, unknown>) {
  return {
    runId: text(run.run_id),
    mode: text(run.mode),
    ended: run.ended === true,
    complete: run.complete === true,
    incompleteReason: text(run.incomplete_reason),
    unblinded: run.unblinded === true,
    domainId: text(record(run.domain).id),
    domainVersion: text(record(run.domain).version),
    adapterId: text(record(run.adapter).id),
    seed: typeof run.seed === "number" ? run.seed : null,
    simVersion: text(run.sim_version),
    startNs: number(run.start_time_ns),
    endNs: number(run.end_time_ns),
  };
}

function faults(truth: Record<string, unknown>) {
  return list(truth.faults).map((raw) => {
    const fault = record(raw);
    return {
      kind: "fault" as const,
      action: "inject",
      injectedNs: number(fault.injected_at_ns),
      onsetNs: number(fault.onset_ns),
      target: text(fault.entity_id),
      name: text(fault.fault) || text(fault.fault_id),
      clearedNs: typeof fault.cleared_ns === "number" ? fault.cleared_ns : null,
      clearedBy: text(fault.cleared_by),
    };
  });
}

function perturbations(truth: Record<string, unknown>) {
  return list(truth.perturbations).map((raw) => {
    const perturbation = record(raw);
    return {
      kind: "perturbation" as const,
      action: text(perturbation.action),
      injectedNs: number(perturbation.at_ns),
      onsetNs: number(perturbation.from_ns) || number(perturbation.at_ns),
      target: "delivery",
      name: text(perturbation.perturbation),
      clearedNs: null,
      clearedBy: "",
    };
  });
}

function environment(truth: Record<string, unknown>) {
  return list(truth.environment).map((raw) => {
    const event = record(raw);
    return {
      kind: "environment" as const,
      action: "inject",
      injectedNs: number(event.at_ns),
      onsetNs: number(event.at_ns),
      target: text(event.target),
      name: text(event.fault),
      clearedNs: null,
      clearedBy: "",
    };
  });
}

function effectorCalls(truth: Record<string, unknown>) {
  return list(truth.effector_calls).map((raw) => {
    const call = record(raw);
    const outcome = record(call.outcome);
    return {
      atNs: number(call.at_ns),
      entity: text(call.entity_id),
      effector: text(call.effector),
      outcome: outcomeText(outcome),
    };
  });
}

function outcomeText(outcome: Record<string, unknown>): string {
  if (outcome.accepted !== true) return `refused${text(outcome.refusal) ? `: ${text(outcome.refusal)}` : ""}`;
  return outcome.effect_applied === true
    ? `applied · ${String(Math.round(number(outcome.ack_latency_ms)))} ms`
    : `accepted, no effect · ${text(outcome.mode)}`;
}

function kpiRows(truth: Record<string, unknown>) {
  return list(record(truth.kpis).entities).flatMap((raw) => {
    const entity = record(raw);
    return list(entity.kpis).map((item) => {
      const row = record(item);
      return {
        entity: text(entity.entity_id),
        state: text(row.state),
        mean: number(row.mean),
        min: number(row.min),
        max: number(row.max),
      };
    });
  });
}

function alarmActivations(entities: unknown): readonly AlarmActivation[] {
  return list(entities).flatMap((raw) => {
    const entity = record(raw);
    return list(entity.alarms).flatMap((rawAlarm) => {
      const alarm = record(rawAlarm);
      return list(alarm.activation_times_ns)
        .filter((at): at is number => typeof at === "number" && Number.isFinite(at))
        .map((atNs) => ({
          atNs,
          entity: text(entity.entity_id),
          state: text(alarm.state),
          priority: text(alarm.priority),
        }));
    });
  });
}

function scoring(value: unknown): Report["scoring"] {
  const entry = record(value);
  return text(entry.status) ? { status: text(entry.status), reason: text(entry.reason) } : null;
}

function numbers(value: unknown): Record<string, number> {
  return Object.fromEntries(
    Object.entries(record(value)).filter((entry): entry is [string, number] => typeof entry[1] === "number"),
  );
}

export function parseReplayPayload(value: unknown): Replay | null {
  if (!isRecord(value) || typeof value.matches !== "boolean") return null;
  return {
    matches: value.matches,
    domainChanged: value.domain_changed === true,
    detail: text(value.detail),
    reason: text(value.reason),
    file: text(value.file),
  };
}
