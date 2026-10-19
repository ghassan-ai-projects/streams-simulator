import type {
  CommandStatus,
  NotAppliedReason,
  NotComparedReason,
  ToleranceSource,
  TwinJudgeReason,
  TwinVerdict,
} from "./twin";

interface VerdictWords {
  readonly label: string;
  readonly line: string;
}

export const verdictWords: Readonly<Record<TwinVerdict, VerdictWords>> = {
  agrees: {
    label: "Agrees",
    line: "The twin agrees with the board: every compared reading was within its tolerance.",
  },
  diverges: {
    label: "Diverges",
    line: "The twin and the board part ways: at least one compared reading was beyond its tolerance.",
  },
  cannot_judge: { label: "Cannot judge", line: "The twin cannot judge this run." },
};

export const noDivergenceWords: Readonly<Record<TwinVerdict, string>> = {
  agrees: "No divergence: every compared reading stayed within its tolerance.",
  diverges: "The report names no stretch where they part ways.",
  cannot_judge: "Nothing to show: no reading could be compared with the twin.",
};

export const judgeReasonWords: Readonly<Record<TwinJudgeReason, string>> = {
  no_commands_applied:
    "No command the board accepted was applied to the twin, so it has no commanded input to predict from.",
  no_readings_compared: "No reading of the board could be compared with the twin.",
};

export const commandStatusWords: Readonly<Record<CommandStatus, string>> = {
  applied: "Applied to the twin",
  not_applied: "Not applied",
  twin_refused: "Refused by the twin",
};

export const notAppliedWords: Readonly<Record<NotAppliedReason, string>> = {
  interrupt_refused: "Refused by the node's interrupt controls",
  interrupt_unresolved: "Left waiting for approval and never resolved",
  never_forwarded: "Never reached the board: the gateway sent no receipt",
  refused_by_board: "Refused by the board",
  not_accepted: "Not accepted: the board only queued it",
  board_reported_failure: "Accepted, but the board reported failure",
};

export const notComparedWords: Readonly<Record<NotComparedReason, string>> = {
  historical: "historical readings",
  quality_not_good: "readings whose quality was not good",
  not_numeric: "readings that are not numbers",
  no_twin_entity: "no twin entity matches the device",
  no_twin_channel: "the twin has no channel of this name",
  no_tolerance_declared: "no tolerance declared and no range in the kind file",
  before_first_command: "readings before the first command was applied to the twin",
  twin_not_finite: "readings where the twin's value was not a finite number",
};

export const toleranceSourceWords: Readonly<Record<ToleranceSource, string>> = {
  declared: "declared",
  default_span: "default share of the range",
  none: "none",
};

export function numberWords(value: number): string {
  return String(Number(value.toPrecision(4)));
}
