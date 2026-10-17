import type { KindFile, Selection } from "../../../model/devices";
import {
  addCommand,
  addMetric,
  hasItem,
  parseKind,
  removeItem,
  resolveIssues,
  serialiseKind,
} from "../../../model/devices";
import type { DeviceFailure, DevicesGateway, KindChecked, Outcome } from "../../ports";
import { DraftStore, createOperationIdentity } from "../../shared";
import type { Operation, OperationIdentity } from "../../shared";
import type { DeviceSession } from "./api";
import type { DeviceState, Draft } from "./deviceDraft";
import { canSave, isBusy, isSaving } from "./deviceSelectors";

const unreadable: DeviceFailure = {
  kind: "unreadable",
  diagnostic: { displayCode: "kind_unreadable", message: "The file is not a device definition." },
};

const kindSelected: Selection = { part: "kind" };

type Addition = (kind: KindFile) => { readonly kind: KindFile; readonly selected: Selection };

function fresh(path: string, kind: KindFile): Draft {
  return {
    phase: "ready",
    path,
    saved: kind,
    doc: kind,
    selected: kindSelected,
    save: { kind: "idle" },
    check: { kind: "unchecked" },
    issues: [],
  };
}

function withResult(draft: Draft, checked: KindFile, result: Outcome<KindChecked, DeviceFailure>): Draft {
  if (result.ok) return { ...draft, check: { kind: "valid", summary: result.value }, issues: [] };
  const { failure } = result;
  if (failure.kind !== "refused") return { ...draft, check: { kind: "failed", failure }, issues: [] };
  return { ...draft, check: { kind: "refused", failure }, issues: resolveIssues(checked, failure.issues) };
}

class DeviceStore extends DraftStore<DeviceState, Draft> {
  private readonly checks = createOperationIdentity();
  private etag = "";

  constructor(
    private readonly gateway: DevicesGateway,
    private readonly operations: OperationIdentity,
  ) {
    super({ phase: "idle" });
  }

  protected override loading(path: string): DeviceState {
    return { phase: "loading", path };
  }

  protected override toDraft(state: DeviceState): Draft | null {
    return state.phase === "ready" ? state : null;
  }

  close(): void {
    this.operations.invalidate();
    this.checks.invalidate();
    this.etag = "";
    this.set({ phase: "idle" });
  }

  protected override async restart(path: string): Promise<void> {
    const request = this.operations.begin();
    const file = await this.gateway.readFile(path);
    const kind = file.ok ? parseKind(file.value.content) : null;
    if (!this.holds(request, path)) return;
    if (!file.ok || kind === null) {
      this.set({ phase: "failed", path, failure: file.ok ? unreadable : file.failure });
      return;
    }
    this.etag = file.value.etag;
    this.set(fresh(path, kind));
    await this.runCheck(request, path, kind);
  }

  private holds(operation: Operation, path: string): boolean {
    const state = this.snapshot();
    return operation.live() && state.phase !== "idle" && state.path === path;
  }

  select(selection: Selection): void {
    this.update((draft) => ({ ...draft, selected: selection }));
  }

  edit(change: (kind: KindFile) => KindFile): void {
    this.update((draft) => {
      const doc = change(draft.doc);
      return doc === draft.doc ? draft : { ...draft, doc };
    });
  }

  addMetric(): void {
    this.add(addMetric);
  }

  addCommand(): void {
    this.add(addCommand);
  }

  remove(selection: Selection): void {
    this.update((draft) => ({ ...draft, doc: removeItem(draft.doc, selection), selected: kindSelected }));
  }

  revert(): void {
    const draft = this.draft();
    if (draft === null || isSaving(draft) || draft.doc === draft.saved) return;
    const selected = hasItem(draft.saved, draft.selected) ? draft.selected : kindSelected;
    this.set({ ...draft, doc: draft.saved, selected });
  }

  async save(): Promise<void> {
    const draft = this.draft();
    const operation = this.operations.current();
    if (draft === null || operation === null || !canSave(draft)) return;
    const { path, doc } = draft;
    this.set({ ...draft, save: { kind: "saving" } });
    const written = await this.gateway.writeFile(path, serialiseKind(doc), this.etag);
    if (!this.holds(operation, path)) return;
    if (!written.ok) {
      this.update((current) => ({ ...current, save: { kind: "write-failed", failure: written.failure } }));
      return;
    }
    this.etag = written.value.etag;
    this.update((current) => ({ ...current, saved: doc, save: { kind: "idle" } }));
    await this.runCheck(operation, path, doc);
  }

  async check(): Promise<void> {
    const draft = this.draft();
    const operation = this.operations.current();
    if (draft === null || operation === null || isBusy(draft)) return;
    await this.runCheck(operation, draft.path, draft.saved);
  }

  private async runCheck(operation: Operation, path: string, checked: KindFile): Promise<void> {
    const check = this.checks.begin();
    this.update((draft) => ({ ...draft, check: { kind: "checking" } }));
    const result = await this.gateway.check(path);
    if (!check.live() || !this.holds(operation, path)) return;
    this.update((draft) => withResult(draft, checked, result));
  }

  private add(addition: Addition): void {
    this.update((draft) => {
      const added = addition(draft.doc);
      return { ...draft, doc: added.kind, selected: added.selected };
    });
  }
}

export function createDeviceSession(gateway: DevicesGateway): DeviceSession {
  return new DeviceStore(gateway, createOperationIdentity());
}
