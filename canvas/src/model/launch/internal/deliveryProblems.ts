import { endText } from "./endProblems";
import type { LaunchPlan } from "./launch";
import type { PerturbationKind, PerturbationParamSpec, PlannedPerturbation } from "./perturbation";
import { kindNamed } from "./perturbation";

type DeliveryPart = "kind" | "start" | "end" | "param";

export interface DeliveryProblem {
  readonly field: "delivery";
  readonly index: number;
  readonly part: DeliveryPart;
  readonly text: string;
}

type Finding = Pick<DeliveryProblem, "part" | "text">;

const wholeStart = (plan: LaunchPlan, atS: number) => Number.isInteger(atS) && atS >= 0 && atS < plan.durationS;

function kindFinding(planned: PlannedPerturbation, name: string, kinds: readonly PerturbationKind[]): Finding | null {
  if (planned.kind === "") return { part: "kind", text: `${name} needs a kind.` };
  if (kinds.length === 0 || kindNamed(kinds, planned.kind) !== undefined) return null;
  return { part: "kind", text: `${name}: the simulator offers no delivery condition named ${planned.kind}.` };
}

function timeFinding(
  planned: PlannedPerturbation,
  name: string,
  plan: LaunchPlan,
  known?: PerturbationKind,
): Finding | null {
  if (!wholeStart(plan, planned.atS)) return { part: "start", text: `${name} starts inside the run, before its end.` };
  const end = endText(name, planned.atS, planned.untilS, plan.durationS);
  if (end !== null) return { part: "end", text: end };
  if (known?.requiresEnd === true && planned.untilS === undefined)
    return { part: "end", text: `${name} needs an end: ${known.description}` };
  return null;
}

const rangeWords = (spec: PerturbationParamSpec) =>
  spec.minimum !== undefined && spec.maximum !== undefined
    ? ` from ${String(spec.minimum)} to ${String(spec.maximum)}`
    : "";

function enumText(spec: PerturbationParamSpec, value: unknown): string {
  const values = spec.values ?? [];
  return typeof value === "string" && values.includes(value) ? "" : `${spec.name} is one of ${values.join(", ")}.`;
}

const outside = (spec: PerturbationParamSpec, number: number) =>
  number < (spec.minimum ?? -Infinity) || number > (spec.maximum ?? Infinity);

function numberText(spec: PerturbationParamSpec, value: unknown): string {
  const number = typeof value === "number" ? value : Number.NaN;
  const whole = spec.type === "integer";
  const wrong = !Number.isFinite(number) || (whole && !Number.isInteger(number)) || outside(spec, number);
  return wrong ? `${spec.name} is ${whole ? "a whole number" : "a number"}${rangeWords(spec)}.` : "";
}

function paramText(spec: PerturbationParamSpec, value: unknown): string {
  if (spec.type === "string_list") return "";
  return spec.type === "enum" ? enumText(spec, value) : numberText(spec, value);
}

function paramFinding(planned: PlannedPerturbation, name: string, known?: PerturbationKind): Finding | null {
  for (const [param, value] of Object.entries(planned.params)) {
    const spec = known?.params.find((candidate) => candidate.name === param);
    const text = spec === undefined ? "" : paramText(spec, value);
    if (text !== "") return { part: "param", text: `${name}: ${text}` };
  }
  return null;
}

function findingOf(planned: PlannedPerturbation, index: number, plan: LaunchPlan, kinds: readonly PerturbationKind[]) {
  const name = `Delivery condition ${String(index + 1)}`;
  const known = kindNamed(kinds, planned.kind);
  const kind = kindFinding(planned, name, kinds);
  if (kind !== null) return kind;
  return timeFinding(planned, name, plan, known) ?? paramFinding(planned, name, known);
}

export function deliveryProblems(plan: LaunchPlan, kinds: readonly PerturbationKind[]): readonly DeliveryProblem[] {
  return plan.perturbations.flatMap((planned, index): DeliveryProblem[] => {
    const found = findingOf(planned, index, plan, kinds);
    return found === null ? [] : [{ field: "delivery", index, ...found }];
  });
}
