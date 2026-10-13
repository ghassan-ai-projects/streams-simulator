export type Badge = "verified" | "not-verified" | "incomplete";

interface Check {
  readonly name: string;
  readonly status: string;
}

interface Disturbance {
  readonly atNs: number;
  readonly what: string;
  readonly target: string;
  readonly detail: string;
}

interface EffectorCall {
  readonly atNs: number;
  readonly entity: string;
  readonly effector: string;
  readonly outcome: string;
}

interface KpiRow {
  readonly entity: string;
  readonly state: string;
  readonly mean: number;
  readonly min: number;
  readonly max: number;
}

export interface Report {
  readonly runId: string;
  readonly mode: string;
  readonly complete: boolean;
  readonly unblinded: boolean;
  readonly domainId: string;
  readonly adapterId: string;
  readonly seed: number;
  readonly startNs: number;
  readonly endNs: number;
  readonly verification: string;
  readonly checks: readonly Check[];
  readonly counts: Readonly<Record<string, number>>;
  readonly dropped: Readonly<Record<string, number>>;
  readonly disturbances: readonly Disturbance[];
  readonly effectorCalls: readonly EffectorCall[];
  readonly kpis: readonly KpiRow[];
  readonly alarms: Readonly<Record<string, number>>;
  readonly scoring: string;
  readonly truthUnavailable: string;
}

type Json = Record<string, unknown>;

const obj = (value: unknown): Json =>
  typeof value === "object" && value !== null && !Array.isArray(value) ? (value as Json) : {};
const arr = (value: unknown): readonly unknown[] => (Array.isArray(value) ? value : []);
const str = (value: unknown): string => (typeof value === "string" ? value : "");
const num = (value: unknown): number => (typeof value === "number" ? value : 0);

function numbers(value: unknown): Record<string, number> {
  return Object.fromEntries(
    Object.entries(obj(value)).filter((entry): entry is [string, number] => typeof entry[1] === "number"),
  );
}

function faults(truth: Json): Disturbance[] {
  return arr(truth.faults).map((raw) => {
    const f = obj(raw);
    return {
      atNs: num(f.onset_ns) || num(f.at_ns),
      what: `fault ${str(f.action)}`,
      target: str(f.entity_id),
      detail: str(f.fault),
    };
  });
}

function perturbations(truth: Json): Disturbance[] {
  return arr(truth.perturbations).map((raw) => {
    const p = obj(raw);
    return {
      atNs: num(p.from_ns) || num(p.at_ns),
      what: `perturbation ${str(p.action)}`,
      target: "delivery",
      detail: str(p.perturbation),
    };
  });
}

function environment(truth: Json): Disturbance[] {
  return arr(truth.environment).map((raw) => {
    const e = obj(raw);
    return { atNs: num(e.at_ns), what: "environment", target: str(e.target), detail: str(e.fault) };
  });
}

function outcome(value: unknown): string {
  const o = obj(value);
  if (o.accepted !== true) return `refused${str(o.refusal) ? `: ${str(o.refusal)}` : ""}`;
  return o.effect_applied === true
    ? `applied · ${String(Math.round(num(o.ack_latency_ms)))} ms`
    : `accepted, no effect · ${str(o.mode)}`;
}

function effectorCalls(truth: Json): EffectorCall[] {
  return arr(truth.effector_calls).map((raw) => {
    const call = obj(raw);
    return {
      atNs: num(call.at_ns),
      entity: str(call.entity_id),
      effector: str(call.effector),
      outcome: outcome(call.outcome),
    };
  });
}

function kpis(truth: Json): KpiRow[] {
  return arr(obj(truth.kpis).entities).flatMap((raw) => {
    const entity = obj(raw);
    return arr(entity.kpis).map((k) => {
      const row = obj(k);
      return {
        entity: str(entity.entity_id),
        state: str(row.state),
        mean: num(row.mean),
        min: num(row.min),
        max: num(row.max),
      };
    });
  });
}

function scoring(value: unknown): string {
  const s = obj(value);
  return str(s.status) ? `${str(s.status)}${str(s.reason) ? ` · ${str(s.reason)}` : ""}` : "not scored";
}

export function parseReport(value: unknown): Report {
  const body = obj(value);
  const run = obj(body.run);
  const truth = obj(body.truth);
  const counts = obj(body.counts);
  const verification = obj(body.verification);
  return {
    runId: str(run.run_id),
    mode: str(run.mode),
    complete: run.complete === true,
    unblinded: run.unblinded === true,
    domainId: str(obj(run.domain).id),
    adapterId: str(obj(run.adapter).id),
    seed: num(run.seed),
    startNs: num(run.start_time_ns),
    endNs: num(run.end_time_ns),
    verification: str(verification.status),
    checks: arr(verification.checks).map((c) => ({ name: str(obj(c).name), status: str(obj(c).status) })),
    counts: numbers(counts),
    dropped: numbers(counts.dropped_by_reason),
    disturbances: [...faults(truth), ...perturbations(truth), ...environment(truth)].sort((a, b) => a.atNs - b.atNs),
    effectorCalls: effectorCalls(truth),
    kpis: kpis(truth),
    alarms: numbers(obj(truth.kpis).alarm_system),
    scoring: scoring(body.scoring),
    truthUnavailable: str(body.truth_unavailable),
  };
}

export function verifyMatches(value: unknown): boolean {
  return obj(value).matches === true;
}

export function badgeOf(report: Report, matches: boolean | null): Badge {
  if (!report.complete) return "incomplete";
  if (matches === false || report.verification === "files_differ") return "not-verified";
  return "verified";
}

export const badgeLabels: Readonly<Record<Badge, string>> = {
  verified: "Verified",
  "not-verified": "Not verified",
  incomplete: "Incomplete",
};

export function wallClock(ns: number): string {
  return new Date(ns / 1e6).toISOString().slice(11, 19);
}

export function runSpan(
  report: Pick<Report, "startNs" | "endNs"> | null,
  observed: { readonly startNs: number; readonly endNs: number } | null,
): { readonly startNs: number; readonly endNs: number } {
  if (report !== null && report.endNs > report.startNs) {
    return { startNs: report.startNs, endNs: report.endNs };
  }
  return observed ?? { startNs: 0, endNs: 0 };
}
