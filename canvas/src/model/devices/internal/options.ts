import { list } from "../../shared/json";
import type { Fields, Metric } from "./types";

export function enumOptions(metric: Metric): string {
  return list(metric.extra.enum_values)
    .filter((value): value is string => typeof value === "string")
    .join(", ");
}

export function withEnumOptions(extra: Fields, typed: string): Fields {
  const values = typed
    .split(",")
    .map((value) => value.trim())
    .filter((value) => value !== "");
  const others = Object.fromEntries(Object.entries(extra).filter(([name]) => name !== "enum_values"));
  return values.length === 0 ? others : { ...others, enum_values: values };
}
