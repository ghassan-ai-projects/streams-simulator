import { counted, spaced } from "../../shared/words";
import { withUnit } from "../../shared/units";
import type {
  AssumedCommand,
  AssumedDynamics,
  AssumedFault,
  AssumedKind,
  AssumedReading,
  Assumptions,
  Range,
} from "./assumptionTypes";

interface FactRow {
  readonly label: string;
  readonly value: string;
}

export interface AssumptionRow {
  readonly key: string;
  readonly title: string;
  readonly statement: string;
  readonly facts: readonly FactRow[];
  readonly suggestion: string;
}

export interface KindAssumptions {
  readonly kind: string;
  readonly summary: string;
  readonly readings: readonly AssumptionRow[];
  readonly dynamics: readonly AssumptionRow[];
  readonly commands: readonly AssumptionRow[];
  readonly faults: readonly AssumptionRow[];
}

const recorded: Readonly<Record<string, string>> = {
  composition: "the composition",
  twin_choices: "the twin choices",
  domain_file: "the domain file",
};

const responses: Readonly<Record<string, string>> = {
  direct: "moves what it sets at once",
  first_order: "moves the readings it drives through a first-order lag",
  none: "changes nothing",
  fixed_safe_stop: "a fixed safe stop",
  declared: "as declared in the domain file",
};

function rangeWords(range: Range | null, unit: string): string {
  return range === null ? "not stated" : `${String(range.min)} to ${withUnit(String(range.max), unit)}`;
}

function suggestionWords(suggestion: boolean | null): string {
  if (suggestion === null) return "";
  return suggestion ? "Suggested by the binary" : "Differs from the binary's suggestion";
}

function instrumentWords(reading: AssumedReading): string {
  if (reading.instrument !== "") return reading.instrument;
  return reading.basis === "domain_declared" ? "declared in the domain file" : "plain channel, no library instrument";
}

function cadenceWords({ mode, periodS }: AssumedReading["cadence"]): string {
  return periodS === null ? spaced(mode) : `${spaced(mode)}, every ${String(periodS)} s`;
}

function readingRow(reading: AssumedReading): AssumptionRow {
  const noise = `${spaced(reading.noise.model)}, sigma ${withUnit(String(reading.noise.sigma), reading.unit)}`;
  return {
    key: reading.channel,
    title: reading.channel,
    statement: "",
    facts: [
      { label: "Instrument", value: instrumentWords(reading) },
      { label: "Range", value: rangeWords(reading.range, reading.unit) },
      { label: "Resolution", value: withUnit(String(reading.resolution), reading.unit) },
      { label: "Noise", value: noise },
      { label: "Cadence", value: cadenceWords(reading.cadence) },
      { label: "Recorded in", value: recorded[reading.recordedIn] ?? reading.recordedIn },
    ],
    suggestion: suggestionWords(reading.suggestion),
  };
}

function driveFacts(command: AssumedCommand): readonly FactRow[] {
  return command.drives.map((drive) => ({
    label: `Drives ${drive.from}`,
    value: `${drive.kind}.${drive.metric}: gain ${String(drive.gain)}, time constant ${String(drive.timeConstantS)} s`,
  }));
}

function commandRow(command: AssumedCommand): AssumptionRow {
  return {
    key: command.name,
    title: command.name,
    statement: command.statement,
    facts: [
      { label: "Response", value: responses[command.response] ?? command.response },
      ...driveFacts(command),
      {
        label: "Dead time / time constant",
        value: `${String(command.deadTimeS)} s / ${String(command.timeConstantS)} s`,
      },
      { label: "Risk", value: command.riskClass === "" ? "not stated" : command.riskClass },
      { label: "Safe stop", value: command.safeStop ? "yes" : "no" },
      { label: "Recorded in", value: recorded[command.recordedIn] ?? command.recordedIn },
    ],
    suggestion: suggestionWords(command.suggestion),
  };
}

function dynamicsRow(dynamics: AssumedDynamics): AssumptionRow {
  const facts = [
    { label: "Form", value: spaced(dynamics.form === "" ? "not stated" : dynamics.form) },
    ...(dynamics.gain === null ? [] : [{ label: "Gain", value: String(dynamics.gain) }]),
    ...(dynamics.timeConstantS === null
      ? []
      : [{ label: "Time constant", value: `${String(dynamics.timeConstantS)} s` }]),
    ...(dynamics.clamp === null ? [] : [{ label: "Clamp", value: rangeWords(dynamics.clamp, "") }]),
  ];
  return { key: dynamics.target, title: dynamics.target, statement: "", facts, suggestion: "" };
}

function faultRow(fault: AssumedFault): AssumptionRow {
  const affects = fault.affects.length === 0 ? [] : [{ label: "Affects", value: fault.affects.join(", ") }];
  return {
    key: fault.id,
    title: fault.id,
    statement: fault.description,
    facts: [{ label: "Onset", value: spaced(fault.onsetShape) }, ...affects],
    suggestion: "",
  };
}

export function kindAssumptions(kind: AssumedKind): KindAssumptions {
  const parts = [counted(kind.readings.length, "reading"), counted(kind.commands.length, "command")];
  return {
    kind: kind.kind,
    summary: [kind.instances.join("; "), ...parts, counted(kind.faults.length, "fault")].filter(Boolean).join(" · "),
    readings: kind.readings.map(readingRow),
    dynamics: kind.dynamics.map(dynamicsRow),
    commands: kind.commands.map(commandRow),
    faults: kind.faults.map(faultRow),
  };
}

export function originWords(origin: Assumptions["origin"]): string {
  if (origin.kind === "composition") return "Read from the plant composed from a composition.";
  if (origin.kind === "domain") return "Read from the domain file alone.";
  if (origin.twin?.status === "in_sync")
    return "Read from a plant generated from twin choices that are in sync with it.";
  return "The twin choices no longer produce this plant: read from the domain file alone.";
}
