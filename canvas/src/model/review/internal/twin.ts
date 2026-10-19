export const twinVerdicts = ["agrees", "diverges", "cannot_judge"] as const;
export const twinJudgeReasons = [
  "no_commands_applied",
  "no_readings_compared",
  "node_stop_not_applied",
  "node_stop_unconfirmed",
] as const;
export const commandOrigins = ["host", "node"] as const;
export const commandStatuses = ["applied", "not_applied", "twin_refused"] as const;
export const notAppliedReasons = [
  "interrupt_refused",
  "interrupt_unresolved",
  "never_forwarded",
  "refused_by_board",
  "not_accepted",
  "board_reported_failure",
] as const;
export const notComparedReasons = [
  "historical",
  "quality_not_good",
  "not_numeric",
  "no_twin_entity",
  "no_twin_channel",
  "no_tolerance_declared",
  "before_first_command",
  "twin_not_finite",
] as const;
export const toleranceSources = ["declared", "default_span", "none"] as const;

export type TwinVerdict = (typeof twinVerdicts)[number];
export type TwinJudgeReason = (typeof twinJudgeReasons)[number];
export type CommandOrigin = (typeof commandOrigins)[number];
export type CommandStatus = (typeof commandStatuses)[number];
export type NotAppliedReason = (typeof notAppliedReasons)[number];
export type NotComparedReason = (typeof notComparedReasons)[number];
export type ToleranceSource = (typeof toleranceSources)[number];

export interface TwinRequest {
  readonly run: string;
  readonly domain: string;
  readonly tolerance: string;
}

export interface TwinSummary {
  readonly devices: number;
  readonly metrics: number;
  readonly compared: number;
  readonly within: number;
  readonly diverged: number;
  readonly notCompared: number;
  readonly unattributed: number;
  readonly metricsCannotJudge: number;
  readonly divergences: number;
  readonly commandsApplied: number;
  readonly commandsNotApplied: number;
}

export interface TwinCommand {
  readonly commandId: string;
  readonly device: string;
  readonly command: string;
  readonly origin: CommandOrigin;
  readonly atNs: number | null;
  readonly status: CommandStatus;
  readonly reason: NotAppliedReason | null;
  readonly detail: string;
}

export interface TwinTolerance {
  readonly value: number | null;
  readonly source: ToleranceSource;
}

interface TwinPoint {
  readonly atNs: number;
  readonly board: number;
  readonly twin: number;
}

export interface TwinSeries {
  readonly points: readonly TwinPoint[];
  readonly total: number;
  readonly truncated: boolean;
}

interface MetricCounts {
  readonly compared: number;
  readonly within: number;
  readonly diverged: number;
  readonly notCompared: number;
}

interface ReasonCount {
  readonly reason: NotComparedReason;
  readonly count: number;
}

export interface TwinMetric {
  readonly device: string;
  readonly kind: string;
  readonly metric: string;
  readonly unit: string;
  readonly tolerance: TwinTolerance;
  readonly verdict: TwinVerdict;
  readonly counts: MetricCounts;
  readonly notCompared: readonly ReasonCount[];
  readonly series: TwinSeries;
}

export interface TwinDivergence {
  readonly name: string;
  readonly device: string;
  readonly metric: string;
  readonly firstNs: number;
  readonly lastNs: number;
  readonly readings: number;
  readonly tolerance: number;
  readonly worstError: number;
  readonly worstAtNs: number;
  readonly boardValue: number;
  readonly twinValue: number;
  readonly commands: readonly string[];
}

export interface TwinReport {
  readonly verdict: TwinVerdict;
  readonly reason: TwinJudgeReason | null;
  readonly summary: TwinSummary;
  readonly twinDomain: string;
  readonly evidenceBound: boolean;
  readonly ruleDescription: string;
  readonly commands: readonly TwinCommand[];
  readonly metrics: readonly TwinMetric[];
  readonly divergences: readonly TwinDivergence[];
}
