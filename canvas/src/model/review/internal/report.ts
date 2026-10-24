import { groupBy } from "../../shared/collect";
import { humanise } from "../../shared/words";

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
  readonly clearedNs: number | null;
  readonly clearedBy: string;
}

interface EffectorCall {
  readonly atNs: number;
  readonly entity: string;
  readonly effector: string;
  readonly outcome: string;
}

export function callWords(call: Pick<EffectorCall, "effector" | "outcome">): string {
  return `${humanise(call.effector)} · ${call.outcome}`;
}

export function faultNames(
  report: Report,
  keep: (disturbance: Disturbance) => boolean = () => true,
): Map<string, string[]> {
  const faults = report.disturbances.filter((d) => d.kind === "fault" && keep(d));
  const named = groupBy(faults, (d) => d.target);
  return new Map([...named].map(([target, same]) => [target, same.map((d) => humanise(d.name))] as const));
}

export interface AlarmActivation {
  readonly atNs: number;
  readonly entity: string;
  readonly state: string;
  readonly priority: string;
}

export interface KpiRow {
  readonly entity: string;
  readonly state: string;
  readonly mean: number;
  readonly min: number;
  readonly max: number;
}

interface Termination {
  readonly reason: "completed" | "stopped_by_operator";
  readonly atNs: number;
}

export interface Report {
  readonly runId: string;
  readonly mode: string;
  readonly ended: boolean;
  readonly complete: boolean;
  readonly incompleteReason: string;
  readonly unblinded: boolean;
  readonly domainId: string;
  readonly domainVersion: string;
  readonly adapterId: string;
  readonly seed: number | null;
  readonly simVersion: string;
  readonly startNs: number;
  readonly endNs: number;
  readonly requestedEndNs: number | null;
  readonly termination: Termination | null;
  readonly verification: string;
  readonly checks: readonly Check[];
  readonly counts: Readonly<Record<string, number>>;
  readonly dropped: Readonly<Record<string, number>>;
  readonly disturbances: readonly Disturbance[];
  readonly effectorCalls: readonly EffectorCall[];
  readonly kpis: readonly KpiRow[];
  readonly alarms: Readonly<Record<string, number>>;
  readonly alarmActivations: readonly AlarmActivation[];
  readonly scoring: { readonly status: string; readonly reason: string } | null;
  readonly truthUnavailable: string;
}

export interface Replay {
  readonly matches: boolean;
  readonly domainChanged: boolean;
  readonly detail: string;
  readonly reason: string;
  readonly file: string;
}

const reasons: Readonly<Record<string, (file: string) => string>> = {
  match: () => "Reproduces: replaying the run gave the same output.",
  recorded_spec: () => "Does not verify: a spec the run recorded does not match its digest, so nothing was replayed.",
  world_digest: () => "Does not verify: the recorded specs replay to a different world than the run recorded.",
  trace_divergence: () => "Does not reproduce: replaying the run gave a different trace.",
  journal_divergence: () => "Does not reproduce: replaying the run gave a different journal.",
  evidence_divergence: (file) => `Does not reproduce: replaying the run gave a different ${file || "evidence file"}.`,
  unblinded_stamp: () => "Does not verify: the run's unblinded stamp does not match its replay.",
  digest_missing: (file) => `Does not verify: run.json records no digest for ${file || "an evidence file"}.`,
  file_differs: (file) => `Does not verify: ${file || "a file"} differs from what the run recorded.`,
  file_missing: (file) => `Does not verify: ${file || "a file"} is missing from the run folder.`,
  incomplete: () => "Does not verify: the run ended incomplete, so it cannot be verified as complete.",
  incomplete_stamp: (file) =>
    `Does not verify: run.json says the run ended incomplete, which ${file || "the run's evidence"} contradicts.`,
};

export function replayLine(verify: VerifyState): string {
  if (verify.phase === "loading") return "Replaying the run to check that it reproduces…";
  if (verify.phase === "failed") return "Not checked: this run could not be replayed for its verdict.";
  const phrase = reasons[verify.value.reason];
  if (phrase !== undefined) return phrase(verify.value.file);
  return verify.value.matches
    ? "Reproduces: replaying the run gave the same output."
    : "Does not verify: replaying the run gave a different output than recorded.";
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
  incomplete: "Did not finish",
};

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
  if (observed !== null) return observed;
  return report === null ? { startNs: 0, endNs: 0 } : { startNs: report.startNs, endNs: report.startNs };
}
