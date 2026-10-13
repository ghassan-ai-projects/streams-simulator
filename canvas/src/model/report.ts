export type Badge = "verifying" | "not-checked" | "verified" | "not-verified" | "incomplete";

interface Check {
  readonly name: string;
  readonly status: string;
}

interface Disturbance {
  readonly kind: "fault" | "perturbation" | "environment";
  readonly action: string;
  readonly injectedNs: number;
  readonly onsetNs: number;
  readonly target: string;
  readonly name: string;
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
  readonly ended: boolean;
  readonly complete: boolean;
  readonly incompleteReason: string;
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
  readonly scoring: { readonly status: string; readonly reason: string } | null;
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
      kind: "fault",
      action: str(f.action),
      injectedNs: num(f.at_ns),
      onsetNs: num(f.onset_ns) || num(f.at_ns),
      target: str(f.entity_id),
      name: str(f.fault) || str(f.fault_id),
    };
  });
}

function perturbations(truth: Json): Disturbance[] {
  return arr(truth.perturbations).map((raw) => {
    const p = obj(raw);
    return {
      kind: "perturbation",
      action: str(p.action),
      injectedNs: num(p.at_ns),
      onsetNs: num(p.from_ns) || num(p.at_ns),
      target: "delivery",
      name: str(p.perturbation),
    };
  });
}

function environment(truth: Json): Disturbance[] {
  return arr(truth.environment).map((raw) => {
    const e = obj(raw);
    return {
      kind: "environment",
      action: "inject",
      injectedNs: num(e.at_ns),
      onsetNs: num(e.at_ns),
      target: str(e.target),
      name: str(e.fault),
    };
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

function scoring(value: unknown): Report["scoring"] {
  const s = obj(value);
  return str(s.status) ? { status: str(s.status), reason: str(s.reason) } : null;
}

function runFacts(run: Json) {
  return {
    runId: str(run.run_id),
    mode: str(run.mode),
    ended: run.ended === true,
    complete: run.complete === true,
    incompleteReason: str(run.incomplete_reason),
    unblinded: run.unblinded === true,
    domainId: str(obj(run.domain).id),
    adapterId: str(obj(run.adapter).id),
    seed: num(run.seed),
    startNs: num(run.start_time_ns),
    endNs: num(run.end_time_ns),
  };
}

export function parseReport(value: unknown): Report {
  const body = obj(value);
  const truth = obj(body.truth);
  const counts = obj(body.counts);
  const verification = obj(body.verification);
  return {
    ...runFacts(obj(body.run)),
    verification: str(verification.status),
    checks: arr(verification.checks).map((c) => ({ name: str(obj(c).name), status: str(obj(c).status) })),
    counts: numbers(counts),
    dropped: numbers(counts.dropped_by_reason),
    disturbances: [...faults(truth), ...perturbations(truth), ...environment(truth)].sort(
      (a, b) => a.onsetNs - b.onsetNs,
    ),
    effectorCalls: effectorCalls(truth),
    kpis: kpis(truth),
    alarms: numbers(obj(truth.kpis).alarm_system),
    scoring: scoring(body.scoring),
    truthUnavailable: str(body.truth_unavailable),
  };
}

export interface Replay {
  readonly matches: boolean;
  readonly domainChanged: boolean;
  readonly detail: string;
}

export function parseReplay(value: unknown): Replay | null {
  const body = obj(value);
  if (typeof body.matches !== "boolean") return null;
  return {
    matches: body.matches,
    domainChanged: body.domain_changed === true,
    detail: typeof body.detail === "string" ? body.detail : "",
  };
}

export function replayLine(verify: VerifyState): string {
  if (verify.phase === "loading") return "Replaying the run to check that it reproduces…";
  if (verify.phase === "failed") return "Not checked: streamsim verify could not replay this run.";
  return verify.value.matches
    ? "Reproduces: replaying the run gave the same output."
    : "Does not reproduce: replaying the run gave different output.";
}

export type VerifyState =
  { readonly phase: "loading" } | { readonly phase: "failed" } | { readonly phase: "ready"; readonly value: Replay };

export function badgeOf(report: Report, verify: VerifyState): Badge {
  if (!report.ended || !report.complete) return "incomplete";
  if (report.verification === "files_differ" || (verify.phase === "ready" && !verify.value.matches))
    return "not-verified";
  if (verify.phase === "loading") return "verifying";
  if (verify.phase === "failed") return "not-checked";
  return report.verification === "files_match" ? "verified" : "not-checked";
}

export const badgeLabels: Readonly<Record<Badge, string>> = {
  verifying: "Verifying…",
  "not-checked": "Not checked",
  verified: "Verified",
  "not-verified": "Not verified",
  incomplete: "Incomplete",
};

export function humanise(id: string): string {
  const parts = id.split("__");
  const meaningful = parts.length > 1 ? parts.slice(1) : parts;
  return meaningful.map((part) => part.replaceAll("_", " ")).join(" · ");
}

const day = 86_400e9;

export function wallClock(ns: number, startNs = ns): string {
  const offset = Math.floor(ns / day) - Math.floor(startNs / day);
  const time = new Date(ns / 1e6).toISOString().slice(11, 19);
  return offset > 0 ? `+${String(offset)}d ${time}` : time;
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
