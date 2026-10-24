import type { Composition } from "../../../model/editor";
import { domainOf, serialise } from "../../../model/editor";
import { issueTargets } from "../../../model/editor";
import type { Draft, Plausibility } from "./editorDraft";
import type { StoreMemo } from "./editorMemo";
import { outputEtagOf } from "./editorMemo";
import type { ComposeSummary, EditorFailure, EditorGateway } from "../../ports";
import type { Operation } from "../../shared";

export interface SaveAttempt {
  readonly operation: Operation;
  readonly saving: Operation;
  readonly path: string;
  readonly doc: Composition;
  readonly source: string;
  readonly sourceEtag: string;
  readonly output: string;
  readonly outputEtag: string;
}

export type SaveResult =
  | { readonly kind: "stale" }
  | { readonly kind: "write-failed"; readonly failure: EditorFailure }
  | { readonly kind: "compose-failed"; readonly failure: EditorFailure }
  | { readonly kind: "composed"; readonly summary: ComposeSummary; readonly outputEtag: string };

export function captureSaveAttempt(
  operation: Operation,
  saving: Operation,
  draft: Draft,
  memo: StoreMemo,
): SaveAttempt {
  const output = domainOf(draft.doc);
  return {
    operation,
    saving,
    path: draft.path,
    doc: draft.doc,
    source: serialise(draft.doc),
    sourceEtag: memo.etag,
    output,
    outputEtag: outputEtagOf(memo, output),
  };
}

export function markSaved(draft: Draft, attempt: SaveAttempt): Draft {
  return { ...draft, saved: attempt.doc };
}

export function applySaveResult(draft: Draft, attempt: SaveAttempt, result: SaveResult): Draft {
  if (result.kind === "write-failed") return { ...draft, save: result };
  if (result.kind === "compose-failed") return composeFailed(draft, attempt, result.failure);
  if (result.kind === "composed") return composed(draft, attempt, result);
  return draft;
}

export function withPlausibility(draft: Draft, attempt: SaveAttempt, plausibility: Plausibility): Draft {
  const { save } = draft;
  return save.kind === "composed" && save.doc === attempt.doc ? { ...draft, save: { ...save, plausibility } } : draft;
}

function composeFailed(draft: Draft, attempt: SaveAttempt, failure: EditorFailure): Draft {
  const failures = "issues" in failure ? (failure.issues ?? []) : [];
  const issues = issueTargets(attempt.doc, failures);
  return { ...draft, save: { kind: "compose-failed", failure }, issues };
}

function composed(draft: Draft, attempt: SaveAttempt, result: SaveResult & { readonly kind: "composed" }): Draft {
  const { channels, faults, effectors } = result.summary;
  const summary = `Composed ${String(channels)} channels, ${String(faults)} faults, ${String(effectors)} effectors.`;
  return {
    ...draft,
    domain: result.outputEtag === "" ? null : attempt.output,
    save: {
      kind: "composed",
      summary,
      doc: attempt.doc,
      digest: result.summary.digest,
      plausibility: { kind: "running" },
    },
    issues: null,
  };
}

export async function saveCaptured(
  gateway: EditorGateway,
  attempt: SaveAttempt,
  current: () => boolean,
  written: (etag: string) => void,
  versionOf: (path: string) => Promise<string>,
): Promise<SaveResult> {
  const result = await gateway.writeFile(attempt.path, attempt.source, attempt.sourceEtag);
  if (!current()) return { kind: "stale" };
  if (!result.ok) return { kind: "write-failed", failure: result.failure };
  written(result.value.etag);
  if (!current()) return { kind: "stale" };
  const composed = await gateway.compose({
    composition: attempt.path,
    output: attempt.output,
    etag: attempt.outputEtag,
  });
  if (!current()) return { kind: "stale" };
  if (!composed.ok) return { kind: "compose-failed", failure: composed.failure };
  const outputEtag = await versionOf(attempt.output);
  if (!current()) return { kind: "stale" };
  return { kind: "composed", summary: composed.value, outputEtag };
}
