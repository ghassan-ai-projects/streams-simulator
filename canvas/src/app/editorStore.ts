import type { Composition, NewEquipment } from "../model/composition";
import { componentOf, domainOf, equipmentOf, parseComposition, serialise } from "../model/composition";
import type { IssueTargets } from "../model/compositionEdits";
import { addEquipment, issueTargets } from "../model/compositionEdits";
import type { LibraryEntry, Presets } from "../model/library";
import type { ComponentSpec } from "../model/ports";
import { componentSpecs } from "./componentSpecs";
import { Observable } from "./observable";
import { createOperationIdentity } from "./shared/operationIdentity";
import type { OperationIdentity, OperationId } from "./shared/operationIdentity";
import type { EditorFailure, EditorGateway } from "./ports/editor";
import type { ComposeSummary } from "./ports/shared";

type Outcome =
  | { readonly kind: "composed"; readonly summary: string }
  | { readonly kind: "refused"; readonly message: string | null }
  | null;

export interface Draft {
  readonly phase: "ready";
  readonly path: string;
  readonly etag: string;
  readonly domainEtag: string;
  readonly saved: Composition;
  readonly doc: Composition;
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

function summaryOf(result: ComposeSummary): string {
  return `Composed ${String(result.channels)} channels, ${String(result.faults)} faults, ${String(result.effectors)} effectors.`;
}

const historyLimit = 100;

export function checked(draft: Draft): boolean {
  return draft.domainEtag !== "" || draft.outcome?.kind === "composed";
}

export function emptyPlant(draft: Draft): boolean {
  return equipmentOf(draft.doc).length === 0;
}

export function canSave(draft: Draft): boolean {
  return draft.doc !== draft.saved || (!checked(draft) && !emptyPlant(draft));
}

function kept(doc: Composition, draft: Draft): string | null {
  const selected = draft.selected;
  return selected !== null && equipmentOf(doc).some((entry) => entry.names.includes(selected)) ? selected : null;
}

export class EditorStore extends Observable<EditorState> {
  private active: OperationId | null = null;

  constructor(
    private readonly gateway: EditorGateway,
    private readonly operations: OperationIdentity = createOperationIdentity(),
  ) {
    super({ phase: "idle" });
  }

  async open(path: string): Promise<void> {
    const state = this.snapshot();
    if (state.phase === "loading" && state.path === path) return;
    if (state.phase === "ready" && state.path === path && state.doc !== state.saved) return;
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
    this.set({ phase: "idle" });
  }

  private async versionOf(path: string): Promise<string> {
    const file = await this.gateway.readFile(path);
    return file.ok ? file.value.etag : "";
  }

  private fresh(path: string, etag: string, doc: Composition): Draft {
    const empty = { specs: new Map(), library: [], selected: null, issues: null, outcome: null, failure: null };
    const history = { past: [], future: [] };
    return { phase: "ready", path, etag, domainEtag: "", saved: doc, doc, saving: false, ...history, ...empty };
  }

  savable(): boolean {
    const state = this.snapshot();
    return state.phase === "ready" && canSave(state);
  }

  unsavedChanges(): boolean {
    const state = this.snapshot();
    return state.phase === "ready" && state.doc !== state.saved;
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
    if (draft !== null) this.edit(() => draft.saved, null);
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
    this.set({ ...draft, saving: true, failure: null });
    const written = await this.gateway.writeFile(draft.path, serialise(draft.doc), draft.etag);
    if (!this.current(operation, draft.path)) return;
    if (!written.ok) {
      this.update((current) => ({ ...current, saving: false, failure: written.failure }));
      return;
    }
    this.update((current) => ({ ...current, etag: written.value.etag, saved: draft.doc }));
    await this.compose(operation, draft.path, draft.doc);
  }

  private async compose(operation: OperationId, path: string, doc: Composition): Promise<void> {
    const draft = this.draft();
    if (draft === null || !this.current(operation, path)) return;
    const out = domainOf(doc);
    const composed = await this.gateway.compose({ composition: draft.path, output: out, etag: draft.domainEtag });
    if (!this.current(operation, path)) return;
    const domainEtag = composed.ok ? await this.versionOf(out) : draft.domainEtag;
    if (!this.current(operation, path)) return;
    this.update((state) =>
      composed.ok
        ? {
            ...state,
            domainEtag,
            saving: false,
            issues: null,
            outcome: { kind: "composed", summary: summaryOf(composed.value) },
          }
        : this.refused(state, doc, composed.failure),
    );
  }

  private refused(state: Draft, doc: Composition, failure: EditorFailure): Draft {
    const failures = "issues" in failure ? (failure.issues ?? []) : [];
    const issues = issueTargets(doc, failures);
    const message = failures.length === 0 ? failure.diagnostic.message : null;
    return { ...state, saving: false, issues, outcome: { kind: "refused", message } };
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
