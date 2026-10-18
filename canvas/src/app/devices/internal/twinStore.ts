import { keepCurrentIssues, serialiseChoices, stemOf, twinDomainPath } from "../../../model/devices";
import type { KindFile, TwinChoices } from "../../../model/devices";
import type { DevicesGateway, TwinMade } from "../../ports";
import { DraftStore, createOperationIdentity } from "../../shared";
import type { Operation } from "../../shared";
import type { TwinSession } from "./api";
import { loadCatalogue } from "./instrumentCatalogue";
import type { Built, TwinDraft, TwinState } from "./twinDraft";
import { loadTwin, readMember } from "./twinLoad";
import {
  afterFlush,
  afterRefusal,
  fresh,
  holdBack,
  joinMembers,
  missingLead,
  pendingOf,
  sameKind,
  withMembers,
} from "./twinReduce";
import { isBuilding, isChoicesDirty, problemsOf } from "./twinSelectors";
import { suggestFor, unreadableHeld } from "./twinSuggest";

class TwinStore extends DraftStore<TwinState, TwinDraft> implements TwinSession {
  private readonly operations = createOperationIdentity();
  private etag = "";
  private domainEtag = "";
  private queue: Promise<void> = Promise.resolve();
  private creating = false;

  constructor(
    private readonly gateway: DevicesGateway,
    private readonly lead: () => KindFile | null,
  ) {
    super({ phase: "idle" });
  }

  protected override loading(path: string): TwinState {
    return { phase: "loading", path };
  }

  protected override toDraft(state: TwinState): TwinDraft | null {
    return state.phase === "ready" ? state : null;
  }

  override async open(path: string): Promise<void> {
    const current = this.draft();
    const state = this.snapshot();
    if (current?.path === path) await this.refresh(current);
    else if (state.phase === "failed" && state.path === path) await this.restart(path);
    else await super.open(path);
  }

  protected override async restart(path: string): Promise<void> {
    const request = this.operations.begin();
    await this.queue;
    const lead = this.lead();
    const loaded = lead === null ? null : await loadTwin(this.gateway, path, lead);
    if (!this.holds(request, path)) return;
    if (!loaded?.ok) {
      this.set({ phase: "failed", path, ...(loaded === null ? { failure: missingLead, held: [] } : loaded.failure) });
      return;
    }
    this.etag = loaded.value.etag;
    this.domainEtag = loaded.value.domainEtag;
    this.set(fresh(path, loaded.value));
    if (loaded.value.onDisk) await this.verifyExisting(request);
  }

  private async refresh(draft: TwinDraft): Promise<void> {
    const lead = this.lead();
    const operation = this.operations.current();
    const first = draft.members[0];
    if (lead === null || operation === null || first === undefined || sameKind(first.kind, lead)) return;
    const members = draft.members.map((member, index) => (index === 0 ? { ...member, kind: lead } : member));
    const [instruments, suggestion] = await Promise.all([
      loadCatalogue(this.gateway, members, draft.instruments),
      suggestFor(this.gateway, members),
    ]);
    if (!this.holds(operation, draft.path)) return;
    this.update((current) => {
      const suggested = suggestion.outcome.ok ? suggestion.outcome.value : current.suggested;
      const grown = withMembers({ ...current, suggested }, members);
      return { ...grown, edited: current.edited, instruments, held: [...suggestion.held, ...pendingOf(current)] };
    });
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
      if (doc === draft.doc) return draft;
      const issues = keepCurrentIssues(draft.issues, draft.doc, doc, draft.members);
      return { ...draft, doc, issues, touched: true, edited: true };
    });
  }

  async include(path: string): Promise<void> {
    const draft = this.draft();
    const operation = this.operations.current();
    if (draft === null || operation === null || draft.members.some((member) => member.path === path)) return;
    const member = await readMember(this.gateway, path);
    if (!this.holds(operation, draft.path)) return;
    if (!member.ok) {
      this.update((current) => holdBack(current, unreadableHeld(path)));
      return;
    }
    const members = [...draft.members, member.value];
    const [suggestion, instruments] = await Promise.all([
      suggestFor(this.gateway, members),
      loadCatalogue(this.gateway, members, draft.instruments),
    ]);
    if (!this.holds(operation, draft.path)) return;
    const waiting = suggestion.held.find((held) => held.path === path);
    const suggested = suggestion.outcome.ok ? suggestion.outcome.value : null;
    this.update((current) =>
      waiting === undefined
        ? joinMembers(current, members, instruments, { suggested, held: suggestion.held })
        : holdBack(current, waiting, suggestion.held),
    );
  }

  leave(path: string): void {
    this.update((draft) => ({
      ...withMembers(
        draft,
        draft.members.filter((member, index) => index === 0 || member.path !== path),
      ),
      held: draft.held.filter((held) => held.path !== path),
    }));
  }

  async retry(): Promise<void> {
    const state = this.snapshot();
    if (state.phase === "failed") return this.restart(state.path);
    const draft = this.draft();
    if (draft === null || draft.held.length === 0) return;
    await this.persist();
    await this.restart(draft.path);
    for (const held of draft.held.filter((entry) => entry.adding)) await this.include(held.path);
  }

  discard(): void {
    this.update((draft) => ({ ...draft, doc: draft.saved, edited: false, issues: [] }));
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
    if (this.creating) return;
    this.creating = true;
    try {
      await this.build();
    } finally {
      this.creating = false;
    }
  }

  async replace(): Promise<void> {
    const draft = this.draft();
    if (draft?.build.kind !== "exists" || this.creating) return;
    this.creating = true;
    try {
      const file = await this.gateway.readFile(twinDomainPath(draft.doc.id));
      this.domainEtag = file.ok ? file.value.etag : "";
      await this.build();
    } finally {
      this.creating = false;
    }
  }

  private async build(): Promise<void> {
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
    const kinds = draft.members.map((member) => stemOf(member.path));
    const made = await this.gateway.twin({ kinds, choices: draft.choicesPath, out, etag: this.domainEtag });
    if (!this.holds(operation, draft.path)) return;
    if (made.ok) await this.verify(operation, draft, made.value, out);
    else this.update((current) => afterRefusal(current, made.failure, draft.doc));
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
    const text = serialiseChoices(draft.doc, draft.members);
    const written = await this.gateway.writeFile(draft.choicesPath, text, this.etag);
    if (!written.ok) {
      if (live()) this.update((current) => ({ ...current, save: { kind: "write-failed", failure: written.failure } }));
      return false;
    }
    this.etag = written.value.etag;
    if (live()) this.update((current) => afterFlush(current, draft.doc, text));
    return true;
  }

  private verifyExisting(operation: Operation): Promise<void> {
    const draft = this.draft();
    if (draft === null) return Promise.resolve();
    return this.verify(operation, draft, null, twinDomainPath(draft.doc.id));
  }

  private async verify(operation: Operation, draft: TwinDraft, made: TwinMade | null, path: string): Promise<void> {
    const file = await this.gateway.readFile(path);
    if (file.ok) this.domainEtag = file.value.etag;
    const started: Built = {
      kind: "built",
      made,
      path,
      fingerprint: serialiseChoices(draft.doc, draft.members),
      implemented: { kind: "checking" },
      lint: { kind: "running" },
    };
    this.update((current) => ({ ...current, build: started, onDisk: true }));
    const checked = await this.gateway.validateDomain(path);
    if (!this.holds(operation, draft.path)) return;
    this.patchBuilt(
      checked.ok
        ? { implemented: { kind: "passed", checked: checked.value } }
        : { implemented: { kind: "failed", failure: checked.failure }, lint: { kind: "unavailable" } },
    );
    if (checked.ok) await this.lint(operation, draft.path, path);
  }

  private async lint(operation: Operation, owner: string, path: string): Promise<void> {
    const report = await this.gateway.lint(path);
    if (!this.holds(operation, owner)) return;
    this.patchBuilt({ lint: report.ok ? { kind: "done", report: report.value } : { kind: "unavailable" } });
  }

  private patchBuilt(change: Partial<Built>): void {
    this.update((draft) => (draft.build.kind === "built" ? { ...draft, build: { ...draft.build, ...change } } : draft));
  }
}

export function createTwinSession(gateway: DevicesGateway, lead: () => KindFile | null): TwinSession {
  return new TwinStore(gateway, lead);
}
