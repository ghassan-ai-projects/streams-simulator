import type { Composition, NewEquipment } from "../model/composition";
import { componentOf, domainOf, equipmentOf, parseComposition } from "../model/composition";
import type { IssueTargets } from "../model/compositionEdits";
import { addEquipment } from "../model/compositionEdits";
import type { LibraryEntry, Presets } from "../model/library";
import type { ComponentSpec } from "../model/ports";
import { componentSpecs } from "./componentSpecs";
import { Observable, createOperationIdentity } from "./shared";
import type { OperationIdentity, OperationId } from "./shared";
import type { EditorFailure, EditorGateway } from "./ports";
import { applySaveResult, captureSaveAttempt, saveCaptured, savedRevision } from "./editorSave";
import type { SaveAttempt } from "./editorSave";

type Outcome =
  | { readonly kind: "composed"; readonly summary: string; readonly revision: number }
  | { readonly kind: "refused"; readonly message: string | null; readonly revision: number }
  | null;

export interface Draft {
  readonly phase: "ready";
  readonly path: string;
  readonly etag: string;
  readonly domainPath: string;
  readonly domainEtag: string;
  readonly saved: Composition;
  readonly doc: Composition;
  readonly revision: number;
  readonly savedRevision: number;
  readonly saveStatus: "idle" | "saved-compose-failed";
  readonly composeFailure: EditorFailure | null;
  readonly past: readonly Composition[];
  readonly future: readonly Composition[];
  readonly specs: ReadonlyMap<string, ComponentSpec>;
  readonly library: readonly LibraryEntry[];
  readonly selected: string | null;
  readonly saving: boolean;
  readonly issues: IssueTargets | null;
  readonly outcome: Outcome;
  readonly failure: EditorFailure | null;
}

export type EditorState =
  | { readonly phase: "idle" }
  | { readonly phase: "loading"; readonly path: string }
  | { readonly phase: "failed"; readonly path: string; readonly failure: EditorFailure }
  | Draft;

const unreadable: EditorFailure = {
  kind: "unreadable",
  diagnostic: { displayCode: "composition_unreadable", message: "The file is not a JSON composition." },
};

const historyLimit = 100;

export function checked(draft: Draft): boolean {
  return (
    !isDirty(draft) &&
    draft.saveStatus === "idle" &&
    ((draft.domainPath === domainOf(draft.doc) && draft.domainEtag !== "") ||
      (draft.outcome?.kind === "composed" && draft.outcome.revision === draft.revision))
  );
}

export function emptyPlant(draft: Draft): boolean {
  return equipmentOf(draft.doc).length === 0;
}

export function canSave(draft: Draft): boolean {
  return isDirty(draft) || draft.saveStatus === "saved-compose-failed" || (!checked(draft) && !emptyPlant(draft));
}

export function isDirty(draft: Draft): boolean {
  return draft.doc !== draft.saved;
}

function kept(doc: Composition, draft: Draft): string | null {
  const selected = draft.selected;
  return selected !== null && equipmentOf(doc).some((entry) => entry.names.includes(selected)) ? selected : null;
}

export class EditorStore extends Observable<EditorState> {
  private active: OperationId | null = null;
  private attemptSequence = 0;
  private activeAttempt: number | null = null;

  constructor(
    private readonly gateway: EditorGateway,
    private readonly operations: OperationIdentity = createOperationIdentity(),
  ) {
    super({ phase: "idle" });
  }

  async open(path: string): Promise<void> {
    const state = this.snapshot();
    if (state.phase === "loading" && state.path === path) return;
    if (state.phase === "ready" && state.path === path && isDirty(state)) return;
    if (state.phase !== "ready" || state.path !== path) this.set({ phase: "loading", path });
    const request = this.operations.begin();
    this.active = request;
    await this.load(request, path, state.phase === "ready" && state.path === path ? state.etag : null);
  }

  async reload(): Promise<void> {
    const draft = this.draft();
    if (draft !== null) {
      const request = this.operations.begin();
      this.active = request;
      await this.load(request, draft.path, null);
    }
  }

  private async load(request: OperationId, path: string, held: string | null): Promise<void> {
    const file = await this.gateway.readFile(path);
    if (file.ok && file.value.etag === held) return;
    const doc = file.ok ? parseComposition(file.value.content) : null;
    if (!this.current(request, path)) return;
    if (!file.ok || doc === null) {
      this.set({ phase: "failed", path, failure: file.ok ? unreadable : file.failure });
      return;
    }
    const [listed, specs, domainEtag] = await Promise.all([
      this.gateway.library(),
      componentSpecs(
        this.gateway,
        equipmentOf(doc).map((entry) => componentOf(entry.use)),
      ),
      this.versionOf(domainOf(doc)),
    ]);
    const library = listed.ok ? listed.value : [];
    if (this.current(request, path))
      this.set({ ...this.fresh(path, file.value.etag, doc), domainEtag, library, specs });
  }

  private current(request: OperationId, path: string): boolean {
    const state = this.snapshot();
    return this.operations.isCurrent(request) && state.phase !== "idle" && state.path === path;
  }

  close(): void {
    this.operations.invalidate();
    this.active = null;
    this.activeAttempt = null;
    this.set({ phase: "idle" });
  }

  private async versionOf(path: string): Promise<string> {
    const file = await this.gateway.readFile(path);
    return file.ok ? file.value.etag : "";
  }

  private fresh(path: string, etag: string, doc: Composition): Draft {
    const empty = { specs: new Map(), library: [], selected: null, issues: null, outcome: null, failure: null };
    const history = { past: [], future: [] };
    return {
      phase: "ready",
      path,
      etag,
      domainPath: domainOf(doc),
      domainEtag: "",
      saved: doc,
      doc,
      revision: 0,
      savedRevision: 0,
      saveStatus: "idle",
      composeFailure: null,
      saving: false,
      ...history,
      ...empty,
    };
  }

  savable(): boolean {
    const state = this.snapshot();
    return state.phase === "ready" && canSave(state);
  }

  unsavedChanges(): boolean {
    const state = this.snapshot();
    return state.phase === "ready" && isDirty(state);
  }

  select(id: string | null): void {
    this.update((draft) => ({ ...draft, selected: id }));
  }

  edit(change: (doc: Composition) => Composition, selected?: string | null): void {
    this.update((draft) => {
      const doc = change(draft.doc);
      if (doc === draft.doc) return draft;
      const past = [...draft.past, draft.doc].slice(-historyLimit);
      return { ...draft, doc, past, future: [], selected: selected === undefined ? draft.selected : selected };
    });
  }

  revert(): void {
    const draft = this.draft();
    if (draft !== null && draft.doc !== draft.saved) {
      const past = [...draft.past, draft.doc].slice(-historyLimit);
      this.set({
        ...draft,
        doc: draft.saved,
        revision: draft.savedRevision,
        past,
        future: [],
        selected: null,
      });
    }
  }

  undo(): void {
    this.update((draft) => {
      const previous = draft.past.at(-1);
      if (previous === undefined) return draft;
      const past = draft.past.slice(0, -1);
      return { ...draft, doc: previous, past, future: [draft.doc, ...draft.future], selected: kept(previous, draft) };
    });
  }

  redo(): void {
    this.update((draft) => {
      const next = draft.future[0];
      if (next === undefined) return draft;
      const future = draft.future.slice(1);
      return { ...draft, doc: next, past: [...draft.past, draft.doc], future, selected: kept(next, draft) };
    });
  }

  async add(entry: NewEquipment): Promise<void> {
    const operation = this.active;
    this.edit((doc) => addEquipment(doc, entry), entry.name);
    const draft = this.draft();
    const component = componentOf(entry.use);
    if (draft === null || draft.specs.has(component)) return;
    const specs = await componentSpecs(this.gateway, [component]);
    if (operation === null || !this.current(operation, draft.path)) return;
    this.update((current) => ({ ...current, specs: new Map([...current.specs, ...specs]) }));
  }

  async presets(use: string): Promise<Presets> {
    const shown = await this.gateway.presets(componentOf(use));
    return shown.ok ? shown.value : { options: [], preferred: "", description: "" };
  }

  async save(): Promise<void> {
    const draft = this.draft();
    if (draft === null || draft.saving) return;
    const operation = this.active;
    if (operation === null) return;
    const attempt: SaveAttempt = captureSaveAttempt(++this.attemptSequence, operation, draft);
    this.activeAttempt = attempt.id;
    this.set({ ...draft, saving: true, failure: null });
    if (!this.currentAttempt(attempt)) return;
    const result = await saveCaptured(
      this.gateway,
      attempt,
      () => this.currentAttempt(attempt),
      (etag) => {
        this.update((current) => savedRevision(current, attempt, etag));
      },
      (path) => this.versionOf(path),
    );
    if (result.kind === "stale") return;
    this.finishAttempt(attempt, (current) => applySaveResult(current, attempt, result));
  }

  private currentAttempt(attempt: SaveAttempt): boolean {
    return this.activeAttempt === attempt.id && this.current(attempt.operation, attempt.path);
  }

  private finishAttempt(attempt: SaveAttempt, change: (draft: Draft) => Draft): void {
    if (!this.currentAttempt(attempt)) return;
    this.activeAttempt = null;
    this.update(change);
  }

  private draft(): Draft | null {
    const state = this.snapshot();
    return state.phase === "ready" ? state : null;
  }

  private update(change: (draft: Draft) => Draft): void {
    const draft = this.draft();
    if (draft !== null) {
      const next = change(draft);
      this.set(next.doc === draft.doc ? next : { ...next, revision: draft.revision + 1 });
    }
  }
}
