export type PerturbParamValue = number | string | readonly string[];

export type PerturbParams = Readonly<Record<string, PerturbParamValue>>;

export interface PlannedPerturbation {
  readonly kind: string;
  readonly params: PerturbParams;
  readonly atS: number;
  readonly untilS?: number;
}

export interface PerturbationParamSpec {
  readonly name: string;
  readonly type: "number" | "integer" | "enum" | "string_list";
  readonly minimum?: number;
  readonly maximum?: number;
  readonly values?: readonly string[];
}

export interface PerturbationKind {
  readonly kind: string;
  readonly category: string;
  readonly description: string;
  readonly requiresEnd: boolean;
  readonly params: readonly PerturbationParamSpec[];
}

export type FaultCategory = "world" | "measurement" | "world_and_measurement";

export type FaultCategories = Readonly<Record<string, Readonly<Record<string, FaultCategory>>>>;

export const faultCategoryWords: Readonly<Record<FaultCategory, string>> = {
  world: "Process (world)",
  measurement: "Measurement (sensor)",
  world_and_measurement: "World and measurement",
};

export const faultCategoryOrder: readonly FaultCategory[] = ["world", "measurement", "world_and_measurement"];

export function kindNamed(kinds: readonly PerturbationKind[], kind: string): PerturbationKind | undefined {
  return kinds.find((known) => known.kind === kind);
}

export function paramSpecOf(kind: PerturbationKind | undefined, name: string): PerturbationParamSpec | undefined {
  return kind?.params.find((param) => param.name === name);
}
