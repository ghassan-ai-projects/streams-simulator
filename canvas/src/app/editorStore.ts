import type { BridgeFailure } from "../model/bridge";
import type { Composition, NewEquipment } from "../model/composition";
import { componentOf, domainOf, equipmentOf, parseComposition, serialise } from "../model/composition";
import type { IssueTargets } from "../model/compositionEdits";
import { addEquipment, issueTargets, newComposition } from "../model/compositionEdits";
import type { LibraryEntry, Presets } from "../model/library";
import { parseLibrary, parsePresets } from "../model/library";
import type { ComponentSpec } from "../model/ports";
import { componentSpecs } from "./componentSpecs";
import { Observable } from "./observable";
import type { Bridge } from "./ports";

type Outcome = { readonly kind: "composed"; readonly summary: string } | { readonly kind: "refused" } | null;

export interface Draft {
  readonly phase: "ready";
  readonly path: string;
  readonly etag: string;
  readonly saved: Composition;
  readonly doc: Composition;
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

export class EditorStore extends Observable<EditorState> {
  constructor(private readonly bridge: Bridge) {
    super({ phase: "idle" });
  }

  async open(path: string): Promise<void> {
    const state = this.snapshot();
    if (state.phase !== "idle" && state.phase !== "failed" && state.path === path) return;
    this.set({ phase: "loading", path });
    const file = await this.bridge.readFile(path);
    const doc = file.ok ? parseComposition(file.value.content) : null;
    if (!file.ok || doc === null) {
      this.set({ phase: "failed", path, failure: file.ok ? unreadable : file.failure });
      return;
    }
    const listed = await this.bridge.action("component.list", {});
    const library = parseLibrary(listed.ok ? listed.value : null);
    const specs = await componentSpecs(
      this.bridge,
      equipmentOf(doc).map((entry) => componentOf(entry.use)),
    );
    this.set({ ...this.fresh(path, file.value.etag, doc), library, specs });
  }

  private fresh(path: string, etag: string, doc: Composition): Draft {
    const empty = { specs: new Map(), library: [], selected: null, issues: null, outcome: null, failure: null };
    return { phase: "ready", path, etag, saved: doc, doc, saving: false, ...empty };
  }

  select(id: string | null): void {
    this.update((draft) => ({ ...draft, selected: id }));
  }

  edit(change: (doc: Composition) => Composition, selected?: string | null): void {
    this.update((draft) => ({ ...draft, doc: change(draft.doc), selected: selected ?? draft.selected }));
  }

  revert(): void {
    this.update((draft) => ({ ...draft, doc: draft.saved, selected: null }));
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
    const out = domainOf(doc);
    const current = await this.bridge.readFile(out);
    const etag = current.ok ? current.value.etag : "";
    const draft = this.draft();
    const composed = await this.bridge.action("domain.compose", { composition: draft?.path ?? "", out, etag });
    this.update((state) =>
      composed.ok
        ? { ...state, saving: false, issues: null, outcome: { kind: "composed", summary: summaryOf(composed.value) } }
        : this.refused(state, doc, composed.failure),
    );
  }

  private refused(state: Draft, doc: Composition, failure: BridgeFailure): Draft {
    const issues = issueTargets(doc, failure.issues);
    const general = failure.issues.length === 0 ? failure : null;
    return { ...state, saving: false, issues, outcome: { kind: "refused" }, failure: general };
  }

  async create(base: string, id: string, title: string): Promise<BridgeFailure | string> {
    const file = await this.bridge.readFile(base);
    const doc = file.ok ? parseComposition(file.value.content) : null;
    if (!file.ok || doc === null) return file.ok ? unreadable : file.failure;
    const path = `compositions/${id}.composition.json`;
    const written = await this.bridge.writeFile(path, serialise(newComposition(doc, id, title)), "");
    return written.ok ? path : written.failure;
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
