import { reconcileChoices, resolveTwinIssues, serialiseChoices, twinDomainPath } from "../../../model/devices";
import type { TwinChoices } from "../../../model/devices";
import type { DeviceFailure, DevicesGateway, TwinMade } from "../../ports";
import { DraftStore, createOperationIdentity } from "../../shared";
import type { Operation } from "../../shared";
import type { TwinSession } from "./api";
import type { Built, Member, TwinDraft, TwinState } from "./twinDraft";
import type { Loaded } from "./twinLoad";
import { loadTwin, readMember } from "./twinLoad";
import { isBuilding, isChoicesDirty, problemsOf } from "./twinSelectors";

function fresh(path: string, loaded: Loaded): TwinDraft {
  return {
    phase: "ready",
    path,
    choicesPath: loaded.choicesPath,
    members: loaded.members,
    known: loaded.known,
    instruments: loaded.instruments,
    touched: loaded.touched,
    saved: loaded.saved,
    doc: loaded.doc,
    save: { kind: "idle" },
    build: { kind: "idle" },
    issues: [],
  };
}

function withMembers(draft: TwinDraft, members: readonly Member[]): TwinDraft {
  const kinds = members.map((member) => member.kind);
  const doc = reconcileChoices(draft.doc, kinds, draft.instruments ?? []);
  return { ...draft, members, doc, touched: true };
}

function afterRefusal(draft: TwinDraft, failure: DeviceFailure, sent: TwinChoices): TwinDraft {
  if (failure.kind === "conflict") return { ...draft, build: { kind: "exists" } };
  if (failure.kind !== "refused") return { ...draft, build: { kind: "failed", failure } };
  const kinds = draft.members.map((member) => member.kind);
  return { ...draft, build: { kind: "refused", failure }, issues: resolveTwinIssues(sent, failure.issues, kinds) };
}

class TwinStore extends DraftStore<TwinState, TwinDraft> implements TwinSession {
  private readonly operations = createOperationIdentity();
  private etag = "";
  private domainEtag = "";
  private queue: Promise<void> = Promise.resolve();

  constructor(private readonly gateway: DevicesGateway) {
    super({ phase: "idle" });
  }

  protected override loading(path: string): TwinState {
    return { phase: "loading", path };
  }

  protected override toDraft(state: TwinState): TwinDraft | null {
    return state.phase === "ready" ? state : null;
  }

  protected override async restart(path: string): Promise<void> {
    const request = this.operations.begin();
    await this.queue;
    const loaded = await loadTwin(this.gateway, path);
    if (!this.holds(request, path)) return;
    if (!loaded.ok) {
      this.set({ phase: "failed", path, failure: loaded.failure });
      return;
    }
    this.etag = loaded.value.etag;
    this.domainEtag = "";
    this.set(fresh(path, loaded.value));
  }

  close(): void {
    const draft = this.draft();
    if (draft !== null && isChoicesDirty(draft)) void this.enqueue(() => this.flush(draft, null));
    this.operations.invalidate();
    this.set({ phase: "idle" });
  }

  edit(change: (choices: TwinChoices) => TwinChoices): void {
    this.update((draft) => {
      const doc = change(draft.doc);
      return doc === draft.doc ? draft : { ...draft, doc, touched: true };
    });
  }

  async include(path: string): Promise<void> {
    const draft = this.draft();
    const operation = this.operations.current();
    if (draft === null || operation === null || draft.members.some((member) => member.path === path)) return;
    const member = await readMember(this.gateway, path);
    if (!member.ok || !this.holds(operation, draft.path)) return;
    this.update((current) => withMembers(current, [...current.members, member.value]));
  }

  leave(kind: string): void {
    this.update((draft) =>
      withMembers(
        draft,
        draft.members.filter((member, index) => index === 0 || member.kind.kind !== kind),
      ),
    );
  }

  persist(): Promise<void> {
    const draft = this.draft();
    if (draft === null || !isChoicesDirty(draft)) return this.queue;
    const operation = this.operations.current();
    return this.enqueue(async () => {
      await this.flush(draft, operation);
    });
  }

  async create(): Promise<void> {
    const start = this.begin();
    if (start === null) return;
    const { draft, operation } = start;
    const written = await this.enqueue(() => this.settle(draft, operation));
    if (!this.holds(operation, draft.path)) return;
    if (!written) {
      this.update((current) => ({ ...current, build: draft.build }));
      return;
    }
    const out = twinDomainPath(draft.doc.id);
    const made = await this.gateway.twin({ choices: draft.choicesPath, out, etag: this.domainEtag });
    if (!this.holds(operation, draft.path)) return;
    if (made.ok) await this.verify(operation, draft, made.value, out);
    else this.update((current) => afterRefusal(current, made.failure, draft.doc));
  }

  async replace(): Promise<void> {
    const draft = this.draft();
    if (draft?.build.kind !== "exists") return;
    const file = await this.gateway.readFile(twinDomainPath(draft.doc.id));
    this.domainEtag = file.ok ? file.value.etag : "";
    await this.create();
  }

  private begin(): { readonly draft: TwinDraft; readonly operation: Operation } | null {
    const draft = this.draft();
    const operation = this.operations.current();
    if (draft === null || operation === null || isBuilding(draft) || problemsOf(draft).length > 0) return null;
    this.update((current) => ({ ...current, build: { kind: "building" }, issues: [] }));
    return { draft, operation };
  }

  private holds(operation: Operation, path: string): boolean {
    const state = this.snapshot();
    return operation.live() && state.phase !== "idle" && state.path === path;
  }

  private enqueue<T>(job: () => Promise<T>): Promise<T> {
    const next = this.queue.then(job);
    this.queue = next.then(
      () => undefined,
      () => undefined,
    );
    return next;
  }

  private settle(draft: TwinDraft, operation: Operation): Promise<boolean> {
    const current = this.draft() ?? draft;
    if (!isChoicesDirty(current) && this.etag !== "") return Promise.resolve(true);
    return this.flush(current, operation);
  }

  private async flush(draft: TwinDraft, operation: Operation | null): Promise<boolean> {
    const live = () => operation !== null && this.holds(operation, draft.path);
    if (live()) this.update((current) => ({ ...current, save: { kind: "saving" } }));
    const written = await this.gateway.writeFile(draft.choicesPath, serialiseChoices(draft.doc), this.etag);
    if (!written.ok) {
      if (live()) this.update((current) => ({ ...current, save: { kind: "write-failed", failure: written.failure } }));
      return false;
    }
    this.etag = written.value.etag;
    if (live()) this.update((current) => ({ ...current, saved: draft.doc, save: { kind: "idle" } }));
    return true;
  }

  private async verify(operation: Operation, draft: TwinDraft, made: TwinMade, path: string): Promise<void> {
    const file = await this.gateway.readFile(path);
    if (file.ok) this.domainEtag = file.value.etag;
    const fingerprint = serialiseChoices(draft.doc);
    const started: Built = {
      kind: "built",
      made,
      path,
      fingerprint,
      implemented: { kind: "checking" },
      lint: { kind: "running" },
    };
    this.update((current) => ({ ...current, build: started }));
    const checked = await this.gateway.validateDomain(path);
    if (!this.holds(operation, draft.path)) return;
    this.patchBuilt(
      checked.ok
        ? { implemented: { kind: "passed", checked: checked.value } }
        : { implemented: { kind: "failed", failure: checked.failure }, lint: { kind: "unavailable" } },
    );
    if (checked.ok) await this.lint(operation, draft, path);
  }

  private async lint(operation: Operation, draft: TwinDraft, path: string): Promise<void> {
    const report = await this.gateway.lint(path);
    if (!this.holds(operation, draft.path)) return;
    this.patchBuilt({ lint: report.ok ? { kind: "done", report: report.value } : { kind: "unavailable" } });
  }

  private patchBuilt(change: Partial<Built>): void {
    this.update((draft) => (draft.build.kind === "built" ? { ...draft, build: { ...draft.build, ...change } } : draft));
  }
}

export function createTwinSession(gateway: DevicesGateway): TwinSession {
  return new TwinStore(gateway);
}
