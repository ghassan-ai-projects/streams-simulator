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

export interface Instrument {
  readonly ref: string;
  readonly title: string;
  readonly family: string;
  readonly unit: string;
  readonly min: number | null;
  readonly max: number | null;
  readonly words: string;
}

export interface TwinIssue {
  readonly scope: "twin" | "kind" | "metric" | "command" | "elsewhere";
  readonly kind: string;
  readonly key: string;
  readonly field: string;
  readonly spot: string;
  readonly message: string;
}
