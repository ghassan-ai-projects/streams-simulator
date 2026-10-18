import type { TwinDraft, TwinSession } from "../../../app/devices";
import type { NavigationStore } from "../../../app/shell";
import type { FieldIssue, Selection, TwinChoices, TwinIssue } from "../../../model/devices";

export type Reveal = (selection: Selection, field: string) => void;

export interface TwinProps {
  readonly twin: TwinSession;
  readonly draft: TwinDraft;
  readonly issues: readonly TwinIssue[];
  readonly needs: readonly FieldIssue[];
  readonly reveal: Reveal;
  readonly navigation: NavigationStore;
}

export function persist(twin: TwinSession): void {
  void twin.persist();
}

export function choose(twin: TwinSession, change: (choices: TwinChoices) => TwinChoices): void {
  twin.edit(change);
  persist(twin);
}

export function commit(twin: TwinSession): () => void {
  return () => {
    persist(twin);
  };
}
