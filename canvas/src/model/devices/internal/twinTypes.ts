import type { Fields } from "./types";

export type Response = "direct" | "first_order" | "none";

export interface Drive {
  readonly from: string;
  readonly kind: string;
  readonly metric: string;
  readonly gain: number | null;
  readonly timeConstantS: number | null;
}

export interface MetricChoice {
  readonly instrument: string;
  readonly initial: number | null;
}

export interface CommandChoice {
  readonly response: Response;
  readonly sets: Readonly<Record<string, string>>;
  readonly drives: readonly Drive[];
}

export interface KindChoice {
  readonly kind: string;
  readonly instance: string;
  readonly metrics: Readonly<Record<string, MetricChoice>>;
  readonly commands: Readonly<Record<string, CommandChoice>>;
}

export interface TwinChoices {
  readonly id: string;
  readonly title: string;
  readonly kinds: readonly KindChoice[];
  readonly extra: Fields;
}

export type FitCode = "unit_incompatible" | "range_unstated" | "range_narrower" | "reading_not_number" | "";

export interface InstrumentFit {
  readonly ok: boolean;
  readonly code: FitCode;
  readonly member: string;
}

export interface InstrumentQuery {
  readonly unit: string;
  readonly min: number | null;
  readonly max: number | null;
}

export interface Instrument {
  readonly ref: string;
  readonly title: string;
  readonly unit: string;
  readonly min: number | null;
  readonly max: number | null;
  readonly match: "exact" | "convertible" | "";
  readonly fit: InstrumentFit | null;
  readonly words: string;
}

export interface TwinIssue {
  readonly scope: "twin" | "kind" | "metric" | "command" | "elsewhere";
  readonly kind: string;
  readonly key: string;
  readonly field: string;
  readonly spot: string;
  readonly pointer: string;
  readonly keyword: string;
  readonly message: string;
  readonly text: string;
}
