import type { BridgeFailure } from "../model/bridge";
import type { Composition, NewEquipment } from "../model/composition";
import { componentOf, domainOf, equipmentOf, parseComposition, serialise } from "../model/composition";
import type { IssueTargets } from "../model/compositionEdits";
import { addEquipment, issueTargets } from "../model/compositionEdits";
import type { LibraryEntry, Presets } from "../model/library";
import { parseLibrary, parsePresets } from "../model/library";
import type { ComponentSpec } from "../model/ports";
import { componentSpecs } from "./componentSpecs";
import { Observable } from "./observable";
import type { Bridge } from "./ports";

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
  readonly failure: BridgeFailure | null;
}

export type EditorState =
  | { readonly phase: "idle" }
  | { readonly phase: "loading"; readonly path: string }
  | { readonly phase: "failed"; readonly path: string; readonly failure: BridgeFailure }
  | Draft;

const unreadable: BridgeFailure = {
  status: 0,
  code: "composition_unreadable",
  message: "The file is not a JSON composition.",
  issues: [],
};

function summaryOf(result: unknown): string {
  const body = typeof result === "object" && result !== null ? (result as Record<string, unknown>) : {};
  const count = (key: string) => (typeof body[key] === "number" ? String(body[key]) : "0");
  return `Composed ${count("channels")} channels, ${count("faults")} faults, ${count("effectors")} effectors.`;
}

const historyLimit = 100;

export class EditorStore extends Observable<EditorState> {
  constructor(private readonly bridge: Bridge) {
    super({ phase: "idle" });
  }

  async open(path: string): Promise<void> {
    const state = this.snapshot();
    if (state.phase === "loading" && state.path === path) return;
    if (state.phase === "ready" && state.path === path && state.doc !== state.saved) return;
    if (state.phase !== "ready" || state.path !== path) this.set({ phase: "loading", path });
    await this.load(path, state.phase === "ready" && state.path === path ? state.etag : null);
  }

  async reload(): Promise<void> {
    const draft = this.draft();
    if (draft !== null) await this.load(draft.path, null);
  }

  private async load(path: string, held: string | null): Promise<void> {
    const file = await this.bridge.readFile(path);
    if (file.ok && file.value.etag === held) return;
    const doc = file.ok ? parseComposition(file.value.content) : null;
    if (!this.current(path)) return;
    if (!file.ok || doc === null) {
      this.set({ phase: "failed", path, failure: file.ok ? unreadable : file.failure });
      return;
    }
    const [listed, specs, domainEtag] = await Promise.all([
      this.bridge.action("component.list", {}),
      componentSpecs(
        this.bridge,
        equipmentOf(doc).map((entry) => componentOf(entry.use)),
      ),
      this.versionOf(domainOf(doc)),
    ]);
    const library = parseLibrary(listed.ok ? listed.value : null);
    if (this.current(path)) this.set({ ...this.fresh(path, file.value.etag, doc), domainEtag, library, specs });
  }

  private current(path: string): boolean {
    const state = this.snapshot();
    return state.phase !== "idle" && state.path === path;
  }

  private async versionOf(path: string): Promise<string> {
    const file = await this.bridge.readFile(path);
    return file.ok ? file.value.etag : "";
  }

  private fresh(path: string, etag: string, doc: Composition): Draft {
    const empty = { specs: new Map(), library: [], selected: null, issues: null, outcome: null, failure: null };
    const history = { past: [], future: [] };
    return { phase: "ready", path, etag, domainEtag: "", saved: doc, doc, saving: false, ...history, ...empty };
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
      return { ...draft, doc: previous, past: draft.past.slice(0, -1), future: [draft.doc, ...draft.future] };
    });
  }

  redo(): void {
    this.update((draft) => {
      const next = draft.future[0];
      if (next === undefined) return draft;
      return { ...draft, doc: next, past: [...draft.past, draft.doc], future: draft.future.slice(1) };
    });
  }

  async add(entry: NewEquipment): Promise<void> {
    this.edit((doc) => addEquipment(doc, entry), entry.name);
    const draft = this.draft();
    const component = componentOf(entry.use);
    if (draft === null || draft.specs.has(component)) return;
    const specs = await componentSpecs(this.bridge, [component]);
    this.update((current) => ({ ...current, specs: new Map([...current.specs, ...specs]) }));
  }

  async presets(use: string): Promise<Presets> {
    const shown = await this.bridge.action("component.show", { ref: componentOf(use) });
    return parsePresets(shown.ok ? shown.value : null);
  }

  async save(): Promise<void> {
    const draft = this.draft();
    if (draft === null || draft.saving) return;
    this.set({ ...draft, saving: true, failure: null });
    const written = await this.bridge.writeFile(draft.path, serialise(draft.doc), draft.etag);
    if (!written.ok) {
      this.update((current) => ({ ...current, saving: false, failure: written.failure }));
      return;
    }
    this.update((current) => ({ ...current, etag: written.value.etag, saved: draft.doc }));
    await this.compose(draft.doc);
  }

  private async compose(doc: Composition): Promise<void> {
    const draft = this.draft();
    if (draft === null) return;
    const out = domainOf(doc);
    const composed = await this.bridge.action("domain.compose", {
      composition: draft.path,
      out,
      etag: draft.domainEtag,
    });
    const domainEtag = composed.ok ? await this.versionOf(out) : draft.domainEtag;
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

  private refused(state: Draft, doc: Composition, failure: BridgeFailure): Draft {
    const issues = issueTargets(doc, failure.issues);
    const message = failure.issues.length === 0 ? failure.message : null;
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
