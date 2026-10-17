import type { Composition, NewEquipment } from "../../../model/editor";
import { componentOf, domainOf, equipmentOf, parseComposition } from "../../../model/editor";
import { addEquipment } from "../../../model/editor";
import type { Presets } from "../../../model/plant";
import { componentSpecs } from "../../plant";
import { Observable } from "../../shared";
import type { OperationIdentity, OperationId } from "../../shared";
import type { EditorFailure, EditorGateway } from "../../ports";
import { applySaveResult, captureSaveAttempt, markSaved, saveCaptured } from "./editorSave";
import type { SaveAttempt, SaveResult } from "./editorSave";
import type { Draft, EditorState } from "./editorDraft";
import { isSaving } from "./editorSelectors";
import { emptyMemo } from "./editorMemo";
import type { StoreMemo } from "./editorMemo";

const unreadable: EditorFailure = {
  kind: "unreadable",
  diagnostic: { displayCode: "composition_unreadable", message: "The file is not a JSON composition." },
};

const historyLimit = 100;

function kept(doc: Composition, draft: Draft): string | null {
  const selected = draft.selected;
  return selected !== null && equipmentOf(doc).some((entry) => entry.names.includes(selected)) ? selected : null;
}

export class EditorStore extends Observable<EditorState> {
  private active: OperationId | null = null;
  private attemptSequence = 0;
  private activeAttempt: number | null = null;
  private memo: StoreMemo = emptyMemo;

  constructor(
    private readonly gateway: EditorGateway,
    private readonly operations: OperationIdentity,
  ) {
    super({ phase: "idle" });
  }

  async open(path: string): Promise<void> {
    this.set({ phase: "loading", path });
    await this.restart(path);
  }

  async reload(): Promise<void> {
    const draft = this.draft();
    if (draft !== null) await this.restart(draft.path);
  }

  private async restart(path: string): Promise<void> {
    const request = this.operations.begin();
    this.active = request;
    await this.load(request, path);
  }

  private async load(request: OperationId, path: string): Promise<void> {
    const file = await this.gateway.readFile(path);
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
    if (!this.current(request, path)) return;
    this.memo = { etag: file.value.etag, domainPath: domainOf(doc), domainEtag };
    this.set({ ...this.fresh(path, doc, domainEtag), library, specs });
  }

  private current(request: OperationId, path: string): boolean {
    const state = this.snapshot();
    return this.operations.isCurrent(request) && state.phase !== "idle" && state.path === path;
  }

  close(): void {
    this.operations.invalidate();
    this.active = null;
    this.activeAttempt = null;
    this.memo = emptyMemo;
    this.set({ phase: "idle" });
  }

  private async versionOf(path: string): Promise<string> {
    const file = await this.gateway.readFile(path);
    return file.ok ? file.value.etag : "";
  }

  private fresh(path: string, doc: Composition, domainEtag: string): Draft {
    return {
      phase: "ready",
      path,
      domain: domainEtag === "" ? null : domainOf(doc),
      saved: doc,
      doc,
      save: { kind: "idle" },
      issues: null,
      past: [],
      future: [],
      specs: new Map(),
      library: [],
      selected: null,
    };
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
    if (draft === null || isSaving(draft) || draft.doc === draft.saved) return;
    const past = [...draft.past, draft.doc].slice(-historyLimit);
    this.set({ ...draft, doc: draft.saved, past, future: [], selected: null });
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
    if (draft === null || isSaving(draft)) return;
    const operation = this.active;
    if (operation === null) return;
    const attempt: SaveAttempt = captureSaveAttempt(++this.attemptSequence, operation, draft, this.memo);
    this.activeAttempt = attempt.id;
    this.set({ ...draft, save: { kind: "saving" } });
    if (!this.currentAttempt(attempt)) return;
    const result = await saveCaptured(
      this.gateway,
      attempt,
      () => this.currentAttempt(attempt),
      (etag) => {
        this.fileWritten(attempt, etag);
      },
      (path) => this.versionOf(path),
    );
    if (result.kind !== "stale") this.finishAttempt(attempt, result);
  }

  private fileWritten(attempt: SaveAttempt, etag: string): void {
    this.memo = { ...this.memo, etag };
    this.update((current) => markSaved(current, attempt));
  }

  private currentAttempt(attempt: SaveAttempt): boolean {
    return this.activeAttempt === attempt.id && this.current(attempt.operation, attempt.path);
  }

  private finishAttempt(attempt: SaveAttempt, result: SaveResult): void {
    if (!this.currentAttempt(attempt)) return;
    this.activeAttempt = null;
    if (result.kind === "composed")
      this.memo = { ...this.memo, domainPath: attempt.output, domainEtag: result.outputEtag };
    this.update((current) => applySaveResult(current, attempt, result));
  }

  private draft(): Draft | null {
    const state = this.snapshot();
    return state.phase === "ready" ? state : null;
  }

  private update(change: (draft: Draft) => Draft): void {
    const draft = this.draft();
    if (draft !== null) this.set(change(draft));
  }
}
