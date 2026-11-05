import { clockOf } from "../../shared/duration";
import { counted, humanise } from "../../shared/words";
import type {
  ComparedCall,
  ComparedEnvironment,
  ComparedFault,
  ComparedPerturbation,
  Comparison,
  Diff,
} from "./comparison";

export interface PlanRow {
  readonly key: string;
  readonly change: "only_in_reference" | "only_in_this" | "changed";
  readonly label: string;
  readonly text: string;
  readonly reference?: string;
  readonly thisRun?: string;
  readonly differences: readonly string[];
}

export interface PlanLane {
  readonly key: string;
  readonly title: string;
  readonly identical: string;
  readonly rows: readonly PlanRow[];
}

interface Lane<T> {
  readonly key: string;
  readonly title: string;
  readonly noun: [string, string];
  readonly diff: Diff<T>;
  readonly words: (item: T, startNs: number) => string;
  readonly heading: (item: T) => string;
}

const clock = (ns: number, startNs: number) => clockOf((ns - startNs) / 1e9);

const argsWords = (args: Readonly<Record<string, string>>) => {
  const shown = Object.entries(args).map(([name, value]) => `${humanise(name)} ${value}`);
  return shown.length === 0 ? "" : ` (${shown.join(", ")})`;
};

function faultWords(fault: ComparedFault, start: number): string {
  const cleared = fault.clearedNs === null ? "" : ` until ${clock(fault.clearedNs, start)}`;
  return `${fault.entity} ${humanise(fault.fault)} from ${clock(fault.onsetNs, start)}${cleared}`;
}

function perturbationWords(item: ComparedPerturbation, start: number): string {
  const kind = humanise(item.kind);
  if (item.action === "clear") return `${kind} cleared at ${clock(item.atNs, start)}`;
  const from = clock(item.fromNs > 0 ? item.fromNs : item.atNs, start);
  const until = item.untilNs > 0 ? `, ends ${clock(item.untilNs, start)}` : "";
  return `${kind} starts ${from}${until}${argsWords(item.params)}`;
}

const environmentWords = (item: ComparedEnvironment, start: number) =>
  `${item.target} ${humanise(item.fault)} at ${clock(item.atNs, start)}`;

function callWords(call: ComparedCall, start: number): string {
  const outcome = call.outcome === "" ? "" : ` · ${call.outcome}`;
  return `${call.entity} ${humanise(call.effector)}${argsWords(call.args)} at ${clock(call.atNs, start)}${outcome}`;
}

const fieldWords: Readonly<Record<string, string>> = {
  onset_ns: "start time",
  cleared_ns: "clear time",
  cleared_by: "how it clears",
  at_ns: "time",
  from_ns: "start time",
  until_ns: "end time",
  args: "arguments",
  params: "settings",
  outcome: "outcome",
};

const differenceWords = (name: string) => fieldWords[name] ?? name.replace(/_ns$/, "").replaceAll("_", " ");

function rowsOf<T>(lane: Lane<T>, starts: readonly number[]): readonly PlanRow[] {
  const [a = 0, b = 0] = starts;
  const { key, diff, words } = lane;
  return [
    ...diff.onlyInA.map((item, index): PlanRow => ({
      key: `${key}-only_in_a-${String(index)}`,
      change: "only_in_reference",
      label: "Only in reference",
      text: words(item, a),
      differences: [],
    })),
    ...diff.onlyInB.map((item, index): PlanRow => ({
      key: `${key}-only_in_b-${String(index)}`,
      change: "only_in_this",
      label: "Only in this run",
      text: words(item, b),
      differences: [],
    })),
    ...diff.different.map((pair, index): PlanRow => ({
      key: `${key}-changed-${String(index)}`,
      change: "changed",
      label: "Changed",
      text: lane.heading(pair.a),
      reference: words(pair.a, a),
      thisRun: words(pair.b, b),
      differences: pair.differences.map(differenceWords).sort(),
    })),
  ];
}

function laneOf<T>(lane: Lane<T>, starts: readonly number[]): PlanLane {
  const same = lane.diff.identical.length;
  const [one, many] = lane.noun;
  const identical = same === 0 ? "" : `${counted(same, one, many)} ${same === 1 ? "is" : "are"} the same in both runs`;
  return { key: lane.key, title: lane.title, identical, rows: rowsOf(lane, starts) };
}

const faultLane = (diff: Diff<ComparedFault>): Lane<ComparedFault> => ({
  key: "faults",
  title: "Faults",
  noun: ["fault", "faults"],
  diff,
  words: faultWords,
  heading: (fault) => `${fault.entity} ${humanise(fault.fault)}`,
});

const perturbationLane = (diff: Diff<ComparedPerturbation>): Lane<ComparedPerturbation> => ({
  key: "perturbations",
  title: "Delivery conditions",
  noun: ["delivery condition", "delivery conditions"],
  diff,
  words: perturbationWords,
  heading: (item) => humanise(item.kind),
});

const environmentLane = (diff: Diff<ComparedEnvironment>): Lane<ComparedEnvironment> => ({
  key: "environment",
  title: "Environment",
  noun: ["environment fault", "environment faults"],
  diff,
  words: environmentWords,
  heading: (item) => `${item.target} ${humanise(item.fault)}`,
});

const callLane = (diff: Diff<ComparedCall>): Lane<ComparedCall> => ({
  key: "effectorCalls",
  title: "Effector calls",
  noun: ["effector call", "effector calls"],
  diff,
  words: callWords,
  heading: (call) => `${call.entity} ${humanise(call.effector)}`,
});

export function planLanes({ a, b, plan }: Comparison): readonly PlanLane[] {
  const starts = [a.startNs, b.startNs];
  return [
    laneOf(faultLane(plan.faults), starts),
    laneOf(perturbationLane(plan.perturbations), starts),
    laneOf(environmentLane(plan.environment), starts),
    laneOf(callLane(plan.effectorCalls), starts),
  ];
}
