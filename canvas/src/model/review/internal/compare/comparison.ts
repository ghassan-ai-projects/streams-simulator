export interface ComparedRun {
  readonly runId: string;
  readonly name: string;
  readonly startNs: number;
}

export interface ComparedFault {
  readonly entity: string;
  readonly fault: string;
  readonly onsetNs: number;
  readonly clearedNs: number | null;
  readonly clearedBy: string;
}

export interface ComparedPerturbation {
  readonly action: string;
  readonly kind: string;
  readonly atNs: number;
  readonly fromNs: number;
  readonly untilNs: number;
  readonly params: Readonly<Record<string, string>>;
}

export interface ComparedEnvironment {
  readonly atNs: number;
  readonly target: string;
  readonly fault: string;
}

export interface ComparedCall {
  readonly atNs: number;
  readonly entity: string;
  readonly effector: string;
  readonly args: Readonly<Record<string, string>>;
  readonly outcome: string;
}

export interface Paired<T> {
  readonly a: T;
  readonly b: T;
  readonly differences: readonly string[];
}

export interface Diff<T> {
  readonly identical: readonly T[];
  readonly different: readonly Paired<T>[];
  readonly onlyInA: readonly T[];
  readonly onlyInB: readonly T[];
}

interface ComparedPlan {
  readonly faults: Diff<ComparedFault>;
  readonly perturbations: Diff<ComparedPerturbation>;
  readonly environment: Diff<ComparedEnvironment>;
  readonly effectorCalls: Diff<ComparedCall>;
}

export interface ComparabilityLimit {
  readonly code: string;
  readonly side: "a" | "b" | null;
  readonly message: string;
}

export interface Scope {
  readonly aLengthNs: number | null;
  readonly bLengthNs: number | null;
}

export type Presence = "both" | "only_in_a" | "only_in_b";

export interface KpiStats {
  readonly mean: number;
  readonly min: number;
  readonly max: number;
  readonly samples: number;
}

export interface ComparedKpi {
  readonly entity: string;
  readonly kpi: string;
  readonly presence: Presence;
  readonly a: KpiStats | null;
  readonly b: KpiStats | null;
  readonly delta: number | null;
  readonly relative: number | null;
  readonly scope: Scope;
}

export interface AlarmStats {
  readonly priority: string;
  readonly activations: number;
  readonly activeS: number;
}

export interface ComparedAlarm {
  readonly entity: string;
  readonly alarm: string;
  readonly presence: Presence;
  readonly a: AlarmStats | null;
  readonly b: AlarmStats | null;
  readonly delta: { readonly activations: number; readonly activeS: number } | null;
  readonly scope: Scope;
}

export type SystemCounts = Readonly<Record<string, number>>;

export interface ComparedAlarmSystem {
  readonly a: SystemCounts | null;
  readonly b: SystemCounts | null;
  readonly delta: SystemCounts | null;
  readonly scope: Scope;
}

export interface Comparison {
  readonly a: ComparedRun;
  readonly b: ComparedRun;
  readonly plan: ComparedPlan;
  readonly limits: readonly ComparabilityLimit[];
  readonly kpis: readonly ComparedKpi[];
  readonly alarms: readonly ComparedAlarm[];
  readonly alarmSystem: ComparedAlarmSystem;
  readonly windowedReason: string;
}
