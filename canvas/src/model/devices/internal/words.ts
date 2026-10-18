import { counted } from "../../shared/words";
import type { Command, Metric } from "./types";

export type Choices = readonly (readonly [string, string])[];

export const typeChoices: Choices = [
  ["number", "Number"],
  ["integer", "Whole number"],
  ["boolean", "On / off"],
  ["enum", "One of a list"],
  ["string", "Text"],
];

export const argumentTypeChoices: Choices = [
  ["number", "Number"],
  ["integer", "Whole number"],
  ["boolean", "On / off"],
  ["string", "Text"],
];

export const riskChoices: Choices = [
  ["low", "Low"],
  ["medium", "Medium"],
  ["high", "High"],
  ["critical", "Critical"],
];

export const unitChoices: Choices = [
  ["Cel", "°C"],
  ["%", "percent"],
  ["Pa", "pascal"],
  ["kPa", "kilopascal"],
  ["bar", "bar"],
  ["V", "volt"],
  ["A", "ampere"],
  ["W", "watt"],
  ["Hz", "hertz"],
  ["rpm", "revolutions per minute"],
  ["s", "second"],
  ["ms", "millisecond"],
  ["m/s", "metre per second"],
  ["1", "no unit"],
];

export function choicesFor(choices: Choices, value: string): Choices {
  if (value === "") return choices.some(([key]) => key === "") ? choices : [["", "Choose…"], ...choices];
  return choices.some(([key]) => key === value) ? choices : [...choices, [value, `${value} (not accepted)`]];
}

export function kindFacts(metrics: number | null, commands: number | null): string {
  if (metrics === null || commands === null) return "Could not be read";
  return `${counted(metrics, "reading")} · ${counted(commands, "command")}`;
}

export function stemOf(path: string): string {
  return (path.split("/").at(-1) ?? path).replace(/\.kind\.json$/, "");
}

export function commandFacts(command: Command): string {
  const risk = riskChoices.find(([key]) => key === command.riskClass)?.[1] ?? command.riskClass;
  const parts = [risk === "" ? "" : `${risk} risk`, command.safeStop ? "safe stop" : ""];
  return parts.filter((part) => part !== "").join(" · ");
}

export function readingFacts(metric: Metric): string {
  const type = typeChoices.find(([key]) => key === metric.dataType)?.[1] ?? metric.dataType;
  return [type, metric.unit].filter((part) => part !== "").join(" · ");
}
