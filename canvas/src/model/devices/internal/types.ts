export type Fields = Readonly<Record<string, unknown>>;

export interface Range {
  readonly min: number | null;
  readonly max: number | null;
}

export interface Metric {
  readonly key: string;
  readonly name: string;
  readonly dataType: string;
  readonly unit: string;
  readonly range: Range;
  readonly resolution: number | null;
  readonly scanClass: string;
  readonly writable: boolean;
  readonly description: string;
  readonly extra: Fields;
}

export interface ArgRow {
  readonly name: string;
  readonly type: string;
  readonly min: number | null;
  readonly max: number | null;
  readonly required: boolean;
}

export type Arguments =
  | { readonly kind: "rows"; readonly rows: readonly ArgRow[]; readonly closed?: true }
  | { readonly kind: "custom"; readonly schema: Fields };

export interface Command {
  readonly key: string;
  readonly name: string;
  readonly riskClass: string;
  readonly safeStop: boolean;
  readonly description: string;
  readonly args: Arguments;
  readonly extra: Fields;
}

export interface KindFile {
  readonly kind: string;
  readonly metrics: readonly Metric[];
  readonly commands: readonly Command[];
  readonly extra: Fields;
}

export type Selection =
  | { readonly part: "kind" }
  | { readonly part: "metric"; readonly key: string }
  | { readonly part: "command"; readonly key: string };

export interface RawIssue {
  readonly pointer: string;
  readonly message: string;
  readonly keyword?: string;
  readonly member?: string;
}

export interface FieldIssue {
  readonly part: "kind" | "metric" | "command" | "file";
  readonly key: string;
  readonly field: string;
  readonly keyword: string;
  readonly message: string;
  readonly text: string;
  readonly twin: boolean;
}

export interface Member {
  readonly path: string;
  readonly kind: KindFile;
}
