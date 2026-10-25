import {
  catalogueKey,
  chosenInstrumentWords,
  kindFacts,
  metricChoiceOf,
  partsWords,
  suggestedWords,
  chosenWords,
  takesInstrument,
} from "../../../model/devices";
import type { Built, TwinDraft } from "./twinDraft";
import { scopeOf } from "./twinScope";
import { builtOf, pairsOf } from "./twinSelectors";
import { isSuggested } from "./twinSuggestions";

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
        const words = chosenInstrumentWords(chosen, twin.instruments[catalogueKey(metric)] ?? []);
        const label = isSuggested(twin, { path: choice.kind, key: metric.key, field: "instrument" })
          ? suggestedWords
          : chosenWords;
        return `${metric.name}: ${words} (${label.toLowerCase()})`;
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

function scopeFacts(twin: TwinDraft): readonly CheckFact[] {
  const { excluded, blocked } = scopeOf(twin);
  const reasons = (list: typeof excluded) => list.map(({ name, reason }) => `${name}: ${reason}`);
  return [
    ...(excluded.length === 0 ? [] : [fact("Left out of the twin", ...reasons(excluded))]),
    ...(blocked.length === 0 ? [] : [fact("Blocking the twin", ...reasons(blocked))]),
  ];
}

export function checkFacts(twin: TwinDraft): readonly CheckFact[] {
  const lines = instrumentLines(twin);
  const built = builtOf(twin);
  return [
    fact("Choices file", twin.choicesPath),
    ...scopeFacts(twin),
    fact("Readings and commands", countsOf(twin)),
    fact("Instruments", ...(lines.length === 0 ? ["None: every reading is a plain reading"] : lines)),
    ...(built === null ? [] : builtFacts(built)),
  ];
}
