import {
  askedCommands,
  choicesOfDocument,
  commandChoiceOf,
  keyChoices,
  metricChoiceOf,
  suggestionNote,
  takesInstrument,
} from "../../../model/devices";
import type { KindChoice, KindFile, TwinChoices } from "../../../model/devices";
import type { TwinDraft } from "./twinDraft";
import { pairsOf } from "./twinSelectors";

export interface ChoiceRef {
  readonly path: string;
  readonly key: string;
  readonly field: "instrument" | "response";
}

export function choiceId({ path, field, key }: ChoiceRef): string {
  return `${path}|${field}|${key}`;
}

function suggestedChoices(draft: TwinDraft): TwinChoices | null {
  return draft.suggested === null ? null : keyChoices(choicesOfDocument(draft.suggested), draft.members);
}

function instrumentRefs(kind: KindFile, choice: KindChoice, offered: KindChoice | undefined): readonly ChoiceRef[] {
  return kind.metrics
    .filter((metric) => metric.name !== "" && takesInstrument(metric))
    .filter((metric) => offered?.metrics[metric.key]?.instrument === metricChoiceOf(choice, metric.key).instrument)
    .map((metric) => ({ path: choice.kind, key: metric.key, field: "instrument" }));
}

function responseRefs(kind: KindFile, choice: KindChoice, offered: KindChoice | undefined): readonly ChoiceRef[] {
  return askedCommands(kind)
    .filter((command) => offered?.commands[command.key]?.response === commandChoiceOf(choice, command.key).response)
    .map((command) => ({ path: choice.kind, key: command.key, field: "response" }));
}

export function suggestedRefs(draft: TwinDraft): readonly ChoiceRef[] {
  const offered = suggestedChoices(draft);
  if (offered === null) return [];
  return pairsOf(draft)
    .flatMap(({ kind, choice }) => {
      const same = offered.kinds.find((entry) => entry.kind === choice.kind);
      return [...instrumentRefs(kind, choice, same), ...responseRefs(kind, choice, same)];
    })
    .filter((ref) => !draft.reviewed.includes(choiceId(ref)));
}

export function isSuggested(draft: TwinDraft, ref: ChoiceRef): boolean {
  const wanted = choiceId(ref);
  return suggestedRefs(draft).some((entry) => choiceId(entry) === wanted);
}

export function suggestionNoteOf(draft: TwinDraft): string | null {
  return suggestionNote(suggestedRefs(draft).length);
}
