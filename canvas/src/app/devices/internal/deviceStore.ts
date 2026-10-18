import type { KindFile, Selection } from "../../../model/devices";
import {
  addCommand,
  addMetric,
  draftProblems,
  hasItem,
  parseKind,
  removeItem,
  resolveIssues,
  schemaIssues,
  serialiseKind,
  twinNeeds,
} from "../../../model/devices";
import type { DeviceFailure, DevicesGateway, KindChecked, Outcome, WorkspaceGateway } from "../../ports";
import { DraftStore, createOperationIdentity } from "../../shared";
import type { Operation, OperationIdentity } from "../../shared";
import type { DeviceSession, TwinSession } from "./api";
import type { DeviceState, Draft, Review } from "./deviceDraft";
import { unreadable } from "./deviceFailures";
import { canSave, isBusy, isSaving } from "./deviceSelectors";
import { createTwinSession } from "./twinStore";

type Watch = WorkspaceGateway["watch"];

const kindSelected: Selection = { part: "kind" };

const noWatch: Watch = () => () => undefined;

const noReview: Review = { state: "idle", schema: [], needs: [] };

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
    review: noReview,
  };
}

function withResult(draft: Draft, result: Outcome<KindChecked, DeviceFailure>): Draft {
  if (result.ok) return { ...draft, check: { kind: "valid", summary: result.value } };
  const { failure } = result;
  return { ...draft, check: { kind: failure.kind === "refused" ? "refused" : "failed", failure } };
}

function withReview(draft: Draft, checked: KindFile, result: Outcome<KindChecked, DeviceFailure>): Draft {
  if (result.ok) return { ...draft, review: { state: "done", schema: [], needs: [] } };
  const { failure } = result;
  if (failure.kind !== "refused" || failure.issues.length === 0) {
    return { ...draft, review: { state: "failed", schema: [], needs: [] } };
  }
  const issues = resolveIssues(checked, failure.issues);
  return { ...draft, review: { state: "done", schema: schemaIssues(issues), needs: twinNeeds(issues) } };
}

function withRefusal(draft: Draft, checked: KindFile, failure: DeviceFailure): Draft {
  if (failure.kind !== "refused") return { ...draft, save: { kind: "write-failed", failure } };
  const issues = schemaIssues(resolveIssues(checked, failure.issues));
  const review: Review = { state: "done", schema: issues, needs: draft.review.needs };
  return { ...draft, save: { kind: "refused", failure }, review };
}

function afterEdit(draft: Draft, doc: KindFile): Draft {
  return { ...draft, doc, save: draft.save.kind === "refused" ? { kind: "idle" } : draft.save };
}

class DeviceStore extends DraftStore<DeviceState, Draft> {
  private readonly checks = createOperationIdentity();
  private etag = "";
  private running: Promise<void> | null = null;
  private again = false;
  private stopWatching: (() => void) | null = null;
  readonly twin: TwinSession;

  constructor(
    private readonly gateway: DevicesGateway,
    private readonly operations: OperationIdentity,
    private readonly watch: Watch,
  ) {
    super({ phase: "idle" });
    this.twin = createTwinSession(gateway, () => this.draft()?.saved ?? null);
  }

  protected override loading(path: string): DeviceState {
    return { phase: "loading", path };
  }

  protected override toDraft(state: DeviceState): Draft | null {
    return state.phase === "ready" ? state : null;
  }

  close(): void {
    this.stopWatching?.();
    this.stopWatching = null;
    this.twin.close();
    this.operations.invalidate();
    this.checks.invalidate();
    this.etag = "";
    this.again = false;
    this.set({ phase: "idle" });
  }

  protected override async restart(path: string): Promise<void> {
    this.twin.close();
    this.stopWatching ??= this.watch(() => {
      void this.twin.retry();
    });
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
    await Promise.all([this.runCheck(request, path), this.requestReview(), this.twin.open(path)]);
  }

  private holds(operation: Operation, path: string): boolean {
    const state = this.snapshot();
    return operation.live() && state.phase !== "idle" && state.path === path;
  }

  select(selection: Selection): void {
    this.update((draft) => ({ ...draft, selected: selection }));
  }

  edit(change: (kind: KindFile) => KindFile): void {
    this.changing((draft) => {
      const doc = change(draft.doc);
      return doc === draft.doc ? draft : afterEdit(draft, doc);
    });
  }

  addMetric(): void {
    this.add(addMetric);
  }

  addCommand(): void {
    this.add(addCommand);
  }

  remove(selection: Selection): void {
    this.changing((draft) => ({ ...afterEdit(draft, removeItem(draft.doc, selection)), selected: kindSelected }));
  }

  revert(): void {
    const draft = this.draft();
    if (draft === null || isSaving(draft) || draft.doc === draft.saved) return;
    const selected = hasItem(draft.saved, draft.selected) ? draft.selected : kindSelected;
    this.changing((current) => ({ ...afterEdit(current, current.saved), selected }));
  }

  async save(): Promise<void> {
    const draft = this.draft();
    const operation = this.operations.current();
    if (draft === null || operation === null || !canSave(draft)) return;
    const { path, doc } = draft;
    const content = serialiseKind(doc);
    this.set({ ...draft, save: { kind: "saving" } });
    const checked = await this.gateway.checkContent(content, false);
    if (!this.holds(operation, path)) return;
    if (!checked.ok) {
      this.update((current) => withRefusal(current, doc, checked.failure));
      return;
    }
    await this.write(operation, path, doc, checked.value);
  }

  async check(): Promise<void> {
    const draft = this.draft();
    const operation = this.operations.current();
    if (draft === null || operation === null || isBusy(draft)) return;
    await this.runCheck(operation, draft.path);
  }

  private async write(operation: Operation, path: string, doc: KindFile, summary: KindChecked): Promise<void> {
    const written = await this.gateway.writeFile(path, serialiseKind(doc), this.etag);
    if (!this.holds(operation, path)) return;
    if (!written.ok) {
      this.update((current) => ({ ...current, save: { kind: "write-failed", failure: written.failure } }));
      return;
    }
    this.etag = written.value.etag;
    this.update((current) => ({ ...current, saved: doc, save: { kind: "idle" }, check: { kind: "valid", summary } }));
    void this.twin.open(path);
  }

  private async runCheck(operation: Operation, path: string): Promise<void> {
    const check = this.checks.begin();
    this.update((draft) => ({ ...draft, check: { kind: "checking" } }));
    const result = await this.gateway.check(path);
    if (!check.live() || !this.holds(operation, path)) return;
    this.update((draft) => withResult(draft, result));
  }

  private requestReview(): Promise<void> {
    if (this.running !== null) {
      this.again = true;
      return this.running;
    }
    this.running = this.reviewLoop();
    return this.running;
  }

  private takeAgain(): boolean {
    const again = this.again;
    this.again = false;
    return again;
  }

  private async reviewLoop(): Promise<void> {
    this.again = false;
    try {
      do {
        await this.reviewOnce();
      } while (this.takeAgain());
    } finally {
      this.running = null;
    }
  }

  private async reviewOnce(): Promise<void> {
    const draft = this.draft();
    const operation = this.operations.current();
    if (draft === null || operation === null) return;
    const { path, doc } = draft;
    if (draftProblems(doc).length > 0) {
      this.update((current) => ({ ...current, review: noReview }));
      return;
    }
    this.update((current) => ({ ...current, review: { ...current.review, state: "checking" } }));
    const result = await this.gateway.checkContent(serialiseKind(doc), true);
    if (this.holds(operation, path)) this.update((current) => withReview(current, doc, result));
  }

  private changing(change: (draft: Draft) => Draft): void {
    const before = this.draft()?.doc;
    this.update(change);
    if (this.draft()?.doc !== before) void this.requestReview();
  }

  private add(addition: Addition): void {
    this.changing((draft) => {
      const added = addition(draft.doc);
      return { ...afterEdit(draft, added.kind), selected: added.selected };
    });
  }
}

export function createDeviceSession(gateway: DevicesGateway, watch: Watch = noWatch): DeviceSession {
  return new DeviceStore(gateway, createOperationIdentity(), watch);
}
