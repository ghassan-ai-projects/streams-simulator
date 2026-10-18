import {
  catalogueKey,
  chosenInstrumentWords,
  kindFacts,
  metricChoiceOf,
  partsWords,
  takesInstrument,
} from "../../../model/devices";
import type { Built, TwinDraft } from "./twinDraft";
import { builtOf, pairsOf } from "./twinSelectors";

export interface CheckFact {
  readonly label: string;
  readonly values: readonly string[];
}

function fact(label: string, ...values: readonly string[]): CheckFact {
  return { label, values };
}

function instrumentLines(twin: TwinDraft): readonly string[] {
  return pairsOf(twin).flatMap(({ kind, choice }) =>
    kind.metrics
      .filter((metric) => metric.name !== "" && takesInstrument(metric))
      .map((metric) => {
        const chosen = metricChoiceOf(choice, metric.key).instrument;
        return `${metric.name}: ${chosenInstrumentWords(chosen, twin.instruments[catalogueKey(metric)] ?? [])}`;
      }),
  );
}

function countsOf(twin: TwinDraft): string {
  const kinds = twin.members.map((member) => member.kind);
  const readings = kinds.reduce((total, kind) => total + kind.metrics.length, 0);
  const commands = kinds.reduce((total, kind) => total + kind.commands.length, 0);
  return kindFacts(readings, commands);
}

function digestOf(built: Built): string {
  const { made, implemented } = built;
  if (made !== null) return made.digest;
  return implemented.kind === "passed" ? implemented.checked.digest : "";
}

function builtFacts(built: Built): readonly CheckFact[] {
  const { made } = built;
  const digest = digestOf(built);
  return [
    fact("Plant file", built.path),
    fact("Digest", digest === "" ? "Not reported" : digest),
    ...(made === null
      ? []
      : [
          fact("Parts", partsWords(made.channels, made.effectors, made.faults)),
          fact("Devices", made.implements.join(", ")),
        ]),
  ];
}

export function checkFacts(twin: TwinDraft): readonly CheckFact[] {
  const lines = instrumentLines(twin);
  const built = builtOf(twin);
  return [
    fact("Choices file", twin.choicesPath),
    fact("Readings and commands", countsOf(twin)),
    fact("Instruments", ...(lines.length === 0 ? ["None: every reading is a plain reading"] : lines)),
    ...(built === null ? [] : builtFacts(built)),
  ];
}
