import { list, record } from "../../shared/json";
import { spaced } from "../../shared/words";

export type ArgValue = string | number | boolean;

export interface EffectorArg {
  readonly name: string;
  readonly type: "number" | "integer" | "string" | "boolean" | "other";
  readonly required: boolean;
  readonly options: readonly string[];
  readonly min: number | null;
  readonly max: number | null;
}

export interface EffectorSpec {
  readonly name: string;
  readonly entityArg: string | null;
  readonly args: readonly EffectorArg[];
}

const bound = (value: unknown): number | null => (typeof value === "number" && Number.isFinite(value) ? value : null);

function typeOf(value: unknown): EffectorArg["type"] {
  return value === "number" || value === "integer" || value === "string" || value === "boolean" ? value : "other";
}

function argOf(name: string, schema: unknown, required: readonly unknown[]): EffectorArg {
  const entry = record(schema);
  return {
    name,
    type: typeOf(entry.type),
    required: required.includes(name),
    options: list(entry.enum).map(String),
    min: bound(entry.minimum),
    max: bound(entry.maximum),
  };
}

function argsOf(schema: unknown): readonly EffectorArg[] {
  const body = record(schema);
  const required = list(body.required);
  return Object.entries(record(body.properties)).map(([name, entry]) => argOf(name, entry, required));
}

function entityArgOf(args: readonly EffectorArg[]): string | null {
  const ids = args.filter((arg) => arg.type === "string" && arg.name.endsWith("_id"));
  return ids.length === 1 ? (ids[0]?.name ?? null) : null;
}

export function parseEffectors(value: unknown): readonly EffectorSpec[] {
  return list(value).flatMap((raw): EffectorSpec[] => {
    const entry = record(raw);
    const name = typeof entry.name === "string" ? entry.name : "";
    const args = argsOf(entry.args_schema);
    const entityArg = entityArgOf(args);
    return name === "" ? [] : [{ name, entityArg, args: args.filter((arg) => arg.name !== entityArg) }];
  });
}

function typed(arg: EffectorArg, text: string): ArgValue {
  if (arg.type === "number" || arg.type === "integer") return Number(text);
  return arg.type === "boolean" ? text === "true" : text;
}

export function callArgs(
  spec: EffectorSpec | undefined,
  entity: string,
  values: Readonly<Record<string, string>>,
): Readonly<Record<string, ArgValue>> {
  if (spec === undefined) return {};
  const out: Record<string, ArgValue> = spec.entityArg === null ? {} : { [spec.entityArg]: entity };
  for (const arg of spec.args) {
    const text = (values[arg.name] ?? "").trim();
    if (text !== "") out[arg.name] = typed(arg, text);
  }
  return out;
}

function rangeWords(arg: EffectorArg): string {
  if (arg.min !== null && arg.max !== null) return ` from ${String(arg.min)} to ${String(arg.max)}`;
  if (arg.min !== null) return ` of ${String(arg.min)} or more`;
  return arg.max === null ? "" : ` of ${String(arg.max)} or less`;
}

function numberProblem(arg: EffectorArg, value: ArgValue): string {
  const number = typeof value === "number" ? value : Number.NaN;
  const whole = arg.type === "integer" ? "a whole number" : "a number";
  const wrong = !Number.isFinite(number) || (arg.type === "integer" && !Number.isInteger(number));
  if (wrong || (arg.min !== null && number < arg.min) || (arg.max !== null && number > arg.max))
    return `${spaced(arg.name)} is ${whole}${rangeWords(arg)}.`;
  return "";
}

function valueProblem(arg: EffectorArg, value: ArgValue | undefined): string {
  if (value === undefined) return arg.required ? `${spaced(arg.name)} needs a value.` : "";
  if (arg.type === "number" || arg.type === "integer") return numberProblem(arg, value);
  if (arg.options.length > 0 && !arg.options.includes(String(value)))
    return `${spaced(arg.name)} is one of ${arg.options.join(", ")}.`;
  return "";
}

export function argProblem(spec: EffectorSpec, args: Readonly<Record<string, ArgValue>>): string {
  for (const arg of spec.args) {
    const problem =
      arg.type === "other" && arg.required
        ? `${spaced(arg.name)} cannot be entered here.`
        : valueProblem(arg, args[arg.name]);
    if (problem !== "") return problem;
  }
  return "";
}
